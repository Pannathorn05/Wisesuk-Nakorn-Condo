package account

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/shared/access"
	"backend/internal/shared/audit"
	"backend/internal/shared/types"
	"backend/internal/validate"
)

// BranchChecker คือสิ่งเดียวที่ module นี้ต้องรู้เกี่ยวกับ module branch
// ประกาศไว้ฝั่งผู้ใช้งานตามธรรมเนียม Go เพื่อไม่ให้ผูกกันตรง ๆ
type BranchChecker interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type Service struct {
	repo     *Repository
	security *auth.Manager
	audit    *audit.Recorder
	branches BranchChecker
}

func NewService(repo *Repository, sec *auth.Manager, rec *audit.Recorder, branches BranchChecker) *Service {
	return &Service{repo: repo, security: sec, audit: rec, branches: branches}
}

// errInvalidCredentials ใช้ข้อความเดียวกันทั้งกรณีอีเมลผิดและรหัสผ่านผิด
// เพื่อไม่ให้ผู้โจมตีเดาได้ว่าอีเมลใดมีอยู่ในระบบ
var errInvalidCredentials = httpx.NewError(401, "invalid_credentials", "อีเมลหรือรหัสผ่านไม่ถูกต้อง")

// ---------------------------------------------------------------- auth

type RegisterInput struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

func (s *Service) Register(ctx context.Context, in RegisterInput, ip string) (*TokenPair, error) {
	v := validate.New()
	email := v.Email("email", in.Email)
	v.Password("password", in.Password)
	firstName := v.Required("first_name", in.FirstName)
	lastName := v.Required("last_name", in.LastName)
	phone := v.Phone("phone", in.Phone, true)
	v.MaxLen("first_name", firstName, 100)
	v.MaxLen("last_name", lastName, 100)
	if err := v.Err(); err != nil {
		return nil, err
	}

	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, access.MapErr(err)
	}
	if exists {
		return nil, httpx.ValidationFailed(map[string]string{"email": "อีเมลนี้ถูกใช้งานแล้ว"})
	}

	hash, err := s.security.HashPassword(in.Password)
	if err != nil {
		return nil, httpx.ErrInternal.Wrap(err)
	}

	user, err := s.repo.Create(ctx, CreateUserParams{
		Email: email, PasswordHash: hash,
		FirstName: firstName, LastName: lastName, Phone: phone,
		Role: types.RoleMember,
	})
	if err != nil {
		return nil, access.MapErr(err)
	}

	s.record(ctx, identityOf(user), "auth.register", "user", user.ID.String(), nil, ip)
	return s.issueTokens(ctx, user)
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Service) Login(ctx context.Context, in LoginInput, ip string) (*TokenPair, error) {
	user, err := s.repo.GetByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			s.security.DummyVerify(in.Password)
			return nil, errInvalidCredentials
		}
		return nil, access.MapErr(err)
	}

	if !s.security.VerifyPassword(user.PasswordHash, in.Password) {
		return nil, errInvalidCredentials
	}
	if !user.IsActive {
		return nil, httpx.NewError(403, "account_disabled", "บัญชีนี้ถูกระงับการใช้งาน กรุณาติดต่อผู้ดูแลระบบ")
	}

	_ = s.repo.TouchLastLogin(ctx, user.ID)
	s.record(ctx, identityOf(user), "auth.login", "user", user.ID.String(), nil, ip)
	return s.issueTokens(ctx, user)
}

// Refresh หมุน refresh token: ใบเดิมถูกเพิกถอนทันทีและออกใบใหม่แทน
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	if refreshToken == "" {
		return nil, httpx.ErrUnauthorized
	}

	userID, err := s.repo.ConsumeRefreshToken(ctx, auth.HashRefreshToken(refreshToken))
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, httpx.ErrUnauthorized
		}
		return nil, access.MapErr(err)
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, access.MapErr(err)
	}
	if !user.IsActive {
		return nil, httpx.ErrUnauthorized
	}
	return s.issueTokens(ctx, user)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return access.MapErr(s.repo.RevokeRefreshToken(ctx, auth.HashRefreshToken(refreshToken)))
}

func (s *Service) issueTokens(ctx context.Context, user *User) (*TokenPair, error) {
	accessToken, expiresAt, err := s.security.IssueAccessToken(auth.Subject{
		UserID: user.ID, Role: user.Role, Name: user.FullName(), BranchID: user.BranchID,
	})
	if err != nil {
		return nil, httpx.ErrInternal.Wrap(err)
	}

	plain, hash, err := auth.NewRefreshToken()
	if err != nil {
		return nil, httpx.ErrInternal.Wrap(err)
	}
	if err := s.repo.StoreRefreshToken(ctx, user.ID, hash, time.Now().Add(s.security.RefreshTTL())); err != nil {
		return nil, access.MapErr(err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: plain,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
		User:         user,
	}, nil
}

// ---------------------------------------------------------------- profile

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	return user, access.MapErr(err)
}

type UpdateProfileInput struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     string  `json:"phone"`
	AvatarURL *string `json:"avatar_url"`
}

func (s *Service) UpdateProfile(ctx context.Context, identity middleware.Identity, in UpdateProfileInput, ip string) (*User, error) {
	v := validate.New()
	firstName := v.Required("first_name", in.FirstName)
	lastName := v.Required("last_name", in.LastName)
	phone := v.Phone("phone", in.Phone, true)
	v.MaxLen("first_name", firstName, 100)
	v.MaxLen("last_name", lastName, 100)
	if err := v.Err(); err != nil {
		return nil, err
	}

	user, err := s.repo.UpdateProfile(ctx, identity.UserID, UpdateProfileParams{
		FirstName: firstName, LastName: lastName, Phone: phone, AvatarURL: in.AvatarURL,
	})
	if err != nil {
		return nil, access.MapErr(err)
	}

	s.record(ctx, identity, "user.update_profile", "user", user.ID.String(), nil, ip)
	return user, nil
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

func (s *Service) ChangePassword(ctx context.Context, identity middleware.Identity, in ChangePasswordInput, ip string) error {
	v := validate.New()
	v.Required("current_password", in.CurrentPassword)
	v.Password("new_password", in.NewPassword)
	v.Check(in.NewPassword == in.ConfirmPassword, "confirm_password", "รหัสผ่านยืนยันไม่ตรงกัน")
	v.Check(in.CurrentPassword != in.NewPassword, "new_password", "รหัสผ่านใหม่ต้องไม่ซ้ำกับรหัสผ่านเดิม")
	if err := v.Err(); err != nil {
		return err
	}

	user, err := s.repo.GetByID(ctx, identity.UserID)
	if err != nil {
		return access.MapErr(err)
	}
	if !s.security.VerifyPassword(user.PasswordHash, in.CurrentPassword) {
		return httpx.ValidationFailed(map[string]string{"current_password": "รหัสผ่านเดิมไม่ถูกต้อง"})
	}

	hash, err := s.security.HashPassword(in.NewPassword)
	if err != nil {
		return httpx.ErrInternal.Wrap(err)
	}
	if err := s.repo.UpdatePassword(ctx, user.ID, hash); err != nil {
		return access.MapErr(err)
	}
	// เปลี่ยนรหัสผ่านแล้วต้องเตะ session อื่นออกทั้งหมด
	if err := s.repo.RevokeAllRefreshTokens(ctx, user.ID); err != nil {
		return access.MapErr(err)
	}

	s.record(ctx, identity, "user.change_password", "user", user.ID.String(), nil, ip)
	return nil
}

// ---------------------------------------------------------------- notifications

func (s *Service) ListNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]Notification, int, error) {
	items, unread, err := s.repo.ListNotifications(ctx, userID, limit)
	return items, unread, access.MapErr(err)
}

func (s *Service) MarkNotificationsRead(ctx context.Context, userID uuid.UUID, id *uuid.UUID) error {
	return access.MapErr(s.repo.MarkNotificationsRead(ctx, userID, id))
}

// Notify ให้ module อื่นส่งแจ้งเตือนถึงผู้ใช้ได้ (best-effort)
func (s *Service) Notify(ctx context.Context, userID uuid.UUID, title, body, link string) {
	_ = s.repo.CreateNotification(ctx, userID, title, body, link)
}

// ---------------------------------------------------------------- members

func (s *Service) ListMembers(ctx context.Context, search string, limit, offset int) ([]User, int, error) {
	users, total, err := s.repo.ListMembers(ctx, search, limit, offset)
	return users, total, access.MapErr(err)
}

// ---------------------------------------------------------------- staff (super admin)

func (s *Service) ListStaff(ctx context.Context) ([]User, error) {
	users, err := s.repo.ListStaff(ctx)
	return users, access.MapErr(err)
}

type CreateAdminInput struct {
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	BranchID  uuid.UUID `json:"branch_id"`
}

func (s *Service) CreateAdmin(ctx context.Context, identity middleware.Identity, in CreateAdminInput, ip string) (*User, error) {
	v := validate.New()
	email := v.Email("email", in.Email)
	v.Password("password", in.Password)
	firstName := v.Required("first_name", in.FirstName)
	lastName := v.Required("last_name", in.LastName)
	phone := v.Phone("phone", in.Phone, false)
	v.Check(in.BranchID != uuid.Nil, "branch_id", "กรุณาเลือกสาขาที่รับผิดชอบ")
	if err := v.Err(); err != nil {
		return nil, err
	}

	// ยืนยันว่าสาขามีจริง จะได้ตอบข้อความที่เข้าใจง่ายแทน foreign key error
	ok, err := s.branches.Exists(ctx, in.BranchID)
	if err != nil {
		return nil, access.MapErr(err)
	}
	if !ok {
		return nil, httpx.ValidationFailed(map[string]string{"branch_id": "ไม่พบสาขาที่เลือก"})
	}

	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, access.MapErr(err)
	}
	if exists {
		return nil, httpx.ValidationFailed(map[string]string{"email": "อีเมลนี้ถูกใช้งานแล้ว"})
	}

	hash, err := s.security.HashPassword(in.Password)
	if err != nil {
		return nil, httpx.ErrInternal.Wrap(err)
	}

	branchID := in.BranchID
	user, err := s.repo.Create(ctx, CreateUserParams{
		Email: email, PasswordHash: hash,
		FirstName: firstName, LastName: lastName, Phone: phone,
		Role: types.RoleAdmin, BranchID: &branchID,
	})
	if err != nil {
		return nil, access.MapErr(err)
	}

	s.record(ctx, identity, "admin.create", "user", user.ID.String(),
		map[string]any{"email": email, "branch_id": branchID.String()}, ip)
	return user, nil
}

type UpdateAdminInput struct {
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Phone     string     `json:"phone"`
	BranchID  *uuid.UUID `json:"branch_id"`
	IsActive  *bool      `json:"is_active"`
	Password  string     `json:"password"` // เว้นว่างไว้ถ้าไม่ต้องการเปลี่ยน
}

func (s *Service) UpdateAdmin(ctx context.Context, identity middleware.Identity, adminID uuid.UUID, in UpdateAdminInput, ip string) (*User, error) {
	target, err := s.repo.GetByID(ctx, adminID)
	if err != nil {
		return nil, access.MapErr(err)
	}
	if target.Role == types.RoleMember {
		return nil, httpx.ErrNotFound
	}

	v := validate.New()
	firstName := v.Required("first_name", in.FirstName)
	lastName := v.Required("last_name", in.LastName)
	phone := v.Phone("phone", in.Phone, false)
	if in.Password != "" {
		v.Password("password", in.Password)
	}
	if target.Role == types.RoleAdmin {
		v.Check(in.BranchID != nil && *in.BranchID != uuid.Nil, "branch_id", "ผู้ดูแลระบบต้องมีสาขาที่รับผิดชอบ")
	}
	if err := v.Err(); err != nil {
		return nil, err
	}

	isActive := target.IsActive
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	// กันไม่ให้ superadmin ปิดบัญชีตัวเองจนล็อกตัวเองออกจากระบบ
	if !isActive && target.ID == identity.UserID {
		return nil, httpx.BadRequest("ไม่สามารถระงับบัญชีของตนเองได้")
	}

	updated, err := s.repo.UpdateStaff(ctx, adminID, UpdateStaffParams{
		FirstName: firstName, LastName: lastName, Phone: phone,
		BranchID: in.BranchID, IsActive: isActive,
	})
	if err != nil {
		return nil, access.MapErr(err)
	}

	if in.Password != "" {
		hash, err := s.security.HashPassword(in.Password)
		if err != nil {
			return nil, httpx.ErrInternal.Wrap(err)
		}
		if err := s.repo.UpdatePassword(ctx, adminID, hash); err != nil {
			return nil, access.MapErr(err)
		}
		// รหัสผ่านเปลี่ยนแล้ว session เดิมของบัญชีนั้นต้องใช้ไม่ได้
		_ = s.repo.RevokeAllRefreshTokens(ctx, adminID)
	}

	s.record(ctx, identity, "admin.update", "user", adminID.String(),
		map[string]any{"is_active": isActive}, ip)
	return updated, nil
}

func (s *Service) DeleteAdmin(ctx context.Context, identity middleware.Identity, adminID uuid.UUID, ip string) error {
	if adminID == identity.UserID {
		return httpx.BadRequest("ไม่สามารถลบบัญชีของตนเองได้")
	}
	if err := s.repo.DeleteAdmin(ctx, adminID); err != nil {
		return access.MapErr(err)
	}
	_ = s.repo.RevokeAllRefreshTokens(ctx, adminID)

	s.record(ctx, identity, "admin.delete", "user", adminID.String(), nil, ip)
	return nil
}

// ---------------------------------------------------------------- helpers

func (s *Service) record(ctx context.Context, identity middleware.Identity, action, entityType, entityID string, detail map[string]any, ip string) {
	actorID := identity.UserID
	s.audit.Record(ctx, audit.Entry{
		ActorID: &actorID, ActorRole: identity.Role, ActorName: identity.Name,
		BranchID: identity.BranchID, Action: action,
		EntityType: entityType, EntityID: entityID, Detail: detail, IPAddress: ip,
	})
}

func identityOf(u *User) middleware.Identity {
	return middleware.Identity{
		UserID: u.ID, Role: u.Role, Name: u.FullName(), BranchID: u.BranchID,
	}
}
