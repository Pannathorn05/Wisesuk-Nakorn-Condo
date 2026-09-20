package account

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/oauth"
	"backend/internal/shared/access"
	"backend/internal/shared/audit"
	"backend/internal/shared/types"
	"backend/internal/validate"
)

// BranchChecker คือสิ่งเดียวที่ module นี้ต้องรู้เกี่ยวกับ module branch
// ประกาศไว้ฝั่งผู้ใช้งานตามธรรมเนียม Go เพื่อไม่ให้ผูกกันตรง ๆ
type BranchChecker interface {
	Exists(ctx context.Context, id types.BranchID) (bool, error)
}

type Service struct {
	repo     *Repository
	tx       *database.TxManager
	security *auth.Manager
	audit    *audit.Recorder
	branches BranchChecker
	// defaultPassword คือรหัสผ่านตั้งต้นของบัญชีผู้ดูแลที่หัวหน้าผู้ดูแลสร้างให้
	// (STAFF_DEFAULT_PASSWORD) ใช้คู่กับธง must_change_password เสมอ
	defaultPassword string
}

func NewService(repo *Repository, tx *database.TxManager, sec *auth.Manager, rec *audit.Recorder,
	branches BranchChecker, defaultPassword string) *Service {
	return &Service{repo: repo, tx: tx, security: sec, audit: rec,
		branches: branches, defaultPassword: defaultPassword}
}

// errInvalidCredentials ใช้ข้อความเดียวกันทั้งกรณีอีเมลผิดและรหัสผ่านผิด
// เพื่อไม่ให้ผู้โจมตีเดาได้ว่าอีเมลใดมีอยู่ในระบบ
var errInvalidCredentials = httpx.ErrInvalidCredentials

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

	// บัญชีที่สมัครผ่าน Google/Facebook มี password_hash เป็นค่าว่าง เข้าด้วยรหัสผ่านไม่ได้
	// จนกว่าจะตั้งรหัสผ่านของตัวเอง เช็คตรงนี้ให้ชัดแทนที่จะพึ่งว่า bcrypt จะตกเอง
	if user.PasswordHash == "" {
		s.security.DummyVerify(in.Password)
		return nil, errInvalidCredentials
	}
	if !s.security.VerifyPassword(user.PasswordHash, in.Password) {
		return nil, errInvalidCredentials
	}
	if !user.IsActive {
		return nil, httpx.ErrAccountDisabled
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

// ---------------------------------------------------------------- oauth

// oauthCodeTTL สั้นมากโดยตั้งใจ โค้ดนี้มีชีวิตแค่ช่วงที่เบราว์เซอร์ redirect จาก
// callback ของ backend ไปหน้า frontend แล้วยิงกลับมาแลก ซึ่งกินเวลาไม่กี่วินาที
const oauthCodeTTL = 2 * time.Minute

// ErrOAuthCodeInvalid คือโค้ดแลก token ที่หมดอายุ ถูกใช้ไปแล้ว หรือไม่เคยมีอยู่
var ErrOAuthCodeInvalid = httpx.NewError(http.StatusUnauthorized, httpx.CodeUnauthorized,
	"ลิงก์เข้าสู่ระบบหมดอายุแล้ว กรุณาลองใหม่อีกครั้ง")

// LoginWithProvider คือปลายทางของ OAuth callback ทั้งหมด คืนโค้ดใช้ครั้งเดียวให้ handler
// เอาไปแนบกับ URL ที่ redirect กลับไปหา frontend
//
// ลำดับการตัดสินใจ (สำคัญกว่าตัวโค้ด):
//  1. เจอ identity ที่ผูกไว้แล้ว → เข้าสู่ระบบเลย
//  2. ไม่เจอ แต่มีอีเมลนี้ในระบบ → ผูกบัญชีภายนอกเข้ากับบัญชีเดิม (อีเมลถูกยืนยันจาก
//     provider มาแล้วตั้งแต่ชั้น oauth ไม่งั้นจะสวมบัญชีคนอื่นได้)
//  3. ไม่เจอทั้งคู่ → สมัครสมาชิกใหม่ ได้สิทธิ์ member เสมอ
//
// ทั้งสามทางต้องจบในทรานแซกชันเดียว ถ้าออกโค้ดไม่สำเร็จก็ต้องไม่มีผู้ใช้ใหม่ค้างไว้
func (s *Service) LoginWithProvider(ctx context.Context, p oauth.Profile, ip string) (string, error) {
	plain, hash, err := auth.NewRefreshToken()
	if err != nil {
		return "", httpx.ErrInternal.Wrap(err)
	}

	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		user, action, err := s.resolveOAuthUser(ctx, p)
		if err != nil {
			return err
		}
		if !user.IsActive {
			return httpx.ErrAccountDisabled
		}

		_ = s.repo.TouchLastLogin(ctx, user.ID)
		s.record(ctx, identityOf(user), action, "user", user.ID.String(),
			map[string]any{"provider": p.Provider}, ip)

		return s.repo.StoreOAuthCode(ctx, user.ID, hash, time.Now().Add(oauthCodeTTL))
	})
	if err != nil {
		return "", access.MapErr(err)
	}
	return plain, nil
}

// resolveOAuthUser หาบัญชีที่ตรงกับโปรไฟล์ หรือสร้างใหม่ถ้ายังไม่มี
// คืนชื่อ action ของ audit log มาด้วยเพราะผู้เรียกแยกไม่ออกว่าเพิ่งสมัครหรือเข้าสู่ระบบเฉย ๆ
func (s *Service) resolveOAuthUser(ctx context.Context, p oauth.Profile) (*User, string, error) {
	user, err := s.repo.GetByIdentity(ctx, p.Provider, p.Subject)
	if err == nil {
		return user, "auth.oauth_login", nil
	}
	if !errors.Is(err, database.ErrNotFound) {
		return nil, "", err
	}

	// อีเมลเดิมมีอยู่แล้ว = เคยสมัครด้วยรหัสผ่าน หรือเคยเข้าด้วย provider อีกเจ้า
	user, err = s.repo.GetByEmail(ctx, p.Email)
	if err == nil {
		if err := s.repo.LinkIdentity(ctx, user.ID, p.Provider, p.Subject, p.Email); err != nil {
			return nil, "", err
		}
		return user, "auth.oauth_link", nil
	}
	if !errors.Is(err, database.ErrNotFound) {
		return nil, "", err
	}

	// PasswordHash เป็นค่าว่าง = ยังไม่มีรหัสผ่าน ตั้งเองภายหลังได้ที่ POST /me/password
	user, err = s.repo.Create(ctx, CreateUserParams{
		Email:        p.Email,
		PasswordHash: "",
		FirstName:    oauthName(p.FirstName, p.Email),
		LastName:     p.LastName,
		Role:         types.RoleMember,
	})
	if err != nil {
		return nil, "", err
	}
	if err := s.repo.LinkIdentity(ctx, user.ID, p.Provider, p.Subject, p.Email); err != nil {
		return nil, "", err
	}
	return user, "auth.oauth_register", nil
}

// ExchangeOAuthCode แลกโค้ดใช้ครั้งเดียวเป็น token pair จริง
func (s *Service) ExchangeOAuthCode(ctx context.Context, code string) (*TokenPair, error) {
	if code == "" {
		return nil, ErrOAuthCodeInvalid
	}

	userID, err := s.repo.ConsumeOAuthCode(ctx, auth.HashRefreshToken(code))
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrOAuthCodeInvalid
		}
		return nil, access.MapErr(err)
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, access.MapErr(err)
	}
	if !user.IsActive {
		return nil, httpx.ErrAccountDisabled
	}
	return s.issueTokens(ctx, user)
}

// oauthName กัน first_name ว่างเปล่าเมื่อ provider ไม่ส่งชื่อมาให้ ใช้ส่วนหน้าของอีเมลแทน
// ผู้ใช้แก้ชื่อจริงเองได้ที่หน้าโปรไฟล์
func oauthName(name, email string) string {
	if name = strings.TrimSpace(name); name != "" {
		return name
	}
	if i := strings.Index(email, "@"); i > 0 {
		return email[:i]
	}
	return "ผู้ใช้"
}

// ---------------------------------------------------------------- profile

func (s *Service) Me(ctx context.Context, userID types.UserID) (*User, error) {
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
	user, err := s.repo.GetByID(ctx, identity.UserID)
	if err != nil {
		return access.MapErr(err)
	}

	// บัญชีที่เข้าระบบผ่าน Google/Facebook ยังไม่เคยมีรหัสผ่าน จะเรียกหา "รหัสผ่านเดิม"
	// จากเขาไม่ได้ endpoint เดียวกันนี้จึงทำหน้าที่ "ตั้งรหัสผ่านครั้งแรก" ไปด้วย
	hasPassword := user.PasswordHash != ""

	v := validate.New()
	if hasPassword {
		v.Required("current_password", in.CurrentPassword)
		v.Check(in.CurrentPassword != in.NewPassword, "new_password", "รหัสผ่านใหม่ต้องไม่ซ้ำกับรหัสผ่านเดิม")
	}
	v.Password("new_password", in.NewPassword)
	v.Check(in.NewPassword == in.ConfirmPassword, "confirm_password", "รหัสผ่านยืนยันไม่ตรงกัน")
	if err := v.Err(); err != nil {
		return err
	}

	if hasPassword && !s.security.VerifyPassword(user.PasswordHash, in.CurrentPassword) {
		return httpx.ValidationFailed(map[string]string{"current_password": "รหัสผ่านเดิมไม่ถูกต้อง"})
	}

	hash, err := s.security.HashPassword(in.NewPassword)
	if err != nil {
		return httpx.ErrInternal.Wrap(err)
	}
	if err := s.repo.UpdatePassword(ctx, user.ID, hash, false); err != nil {
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

func (s *Service) ListNotifications(ctx context.Context, userID types.UserID, limit int) ([]Notification, int, error) {
	items, unread, err := s.repo.ListNotifications(ctx, userID, limit)
	return items, unread, access.MapErr(err)
}

func (s *Service) MarkNotificationsRead(ctx context.Context, userID types.UserID, id *types.NotificationID) error {
	return access.MapErr(s.repo.MarkNotificationsRead(ctx, userID, id))
}

// Notify ให้ module อื่นส่งแจ้งเตือนถึงผู้ใช้ได้ (best-effort)
func (s *Service) Notify(ctx context.Context, userID types.UserID, title, body, link string) {
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
	Email     string         `json:"email"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Phone     string         `json:"phone"`
	BranchID  types.BranchID `json:"branch_id"`
}

// CreateAdmin สร้างบัญชีผู้ดูแลจากหน้า "เพิ่มผู้ดูแลใหม่"
//
// ฟอร์มไม่มีช่องรหัสผ่าน ระบบจึงตั้งรหัสตั้งต้นร่วมให้แล้วปักธง must_change_password
// เจ้าตัวต้องเปลี่ยนรหัสเองก่อนใช้งานจริง หัวหน้าผู้ดูแลจึงไม่เคยรู้รหัสผ่านของใคร
func (s *Service) CreateAdmin(ctx context.Context, identity middleware.Identity, in CreateAdminInput, ip string) (*User, error) {
	v := validate.New()
	email := v.Email("email", in.Email)
	firstName := v.Required("first_name", in.FirstName)
	lastName := v.Required("last_name", in.LastName)
	phone := v.Phone("phone", in.Phone, false)
	v.Check(in.BranchID != 0, "branch_id", "กรุณาเลือกสาขาที่รับผิดชอบ")
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

	if s.defaultPassword == "" {
		return nil, httpx.ErrInternal.Wrap(errNoDefaultPassword)
	}
	hash, err := s.security.HashPassword(s.defaultPassword)
	if err != nil {
		return nil, httpx.ErrInternal.Wrap(err)
	}

	branchID := in.BranchID
	user, err := s.repo.Create(ctx, CreateUserParams{
		Email: email, PasswordHash: hash,
		FirstName: firstName, LastName: lastName, Phone: phone,
		Role: types.RoleAdmin, BranchID: &branchID, MustChangePassword: true,
	})
	if err != nil {
		if database.IsUniqueViolation(err) {
			return nil, errBranchTaken
		}
		return nil, access.MapErr(err)
	}

	s.record(ctx, identity, "admin.create", "user", user.ID.String(),
		map[string]any{"email": email, "branch_id": branchID.String()}, ip)
	return user, nil
}

var (
	errNoDefaultPassword = errors.New("account: ยังไม่ได้ตั้ง STAFF_DEFAULT_PASSWORD")

	// สาขาหนึ่งมีผู้ดูแลได้คนเดียว — ดัก unique index uq_admin_per_branch ที่ DB
	// แล้วแปลงเป็นข้อความรายฟิลด์ แทนที่จะปล่อยเป็น 500
	errBranchTaken = httpx.ValidationFailed(map[string]string{
		"branch_id": "สาขานี้มีผู้ดูแลอยู่แล้ว หนึ่งสาขามีผู้ดูแลได้คนเดียว",
	})
)

type UpdateAdminInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	// BranchID คือสาขาที่ย้ายไปรับผิดชอบ บังคับเมื่อเป้าหมายเป็น admin
	BranchID *types.BranchID `json:"branch_id"`
	IsActive *bool           `json:"is_active"`
	Password string          `json:"password"` // เว้นว่างไว้ถ้าไม่ต้องการเปลี่ยน
}

func (s *Service) UpdateAdmin(ctx context.Context, identity middleware.Identity, adminID types.UserID, in UpdateAdminInput, ip string) (*User, error) {
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
		v.Check(in.BranchID != nil && *in.BranchID != 0, "branch_id", "ผู้ดูแลระบบต้องมีสาขาที่รับผิดชอบ")
	} else {
		// หัวหน้าผู้ดูแลเห็นทุกสาขาอยู่แล้ว จึงผูกรายสาขาไม่ได้ (DB ก็ปฏิเสธอยู่ดี)
		v.Check(in.BranchID == nil, "branch_id", "หัวหน้าผู้ดูแลระบบผูกกับสาขาไม่ได้")
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
		if database.IsUniqueViolation(err) {
			return nil, errBranchTaken
		}
		return nil, access.MapErr(err)
	}

	if in.Password != "" {
		hash, err := s.security.HashPassword(in.Password)
		if err != nil {
			return nil, httpx.ErrInternal.Wrap(err)
		}
		// คนอื่นตั้งรหัสให้ จึงต้องบังคับให้เจ้าตัวเปลี่ยนเองเมื่อเข้าใช้งานครั้งถัดไป
		if err := s.repo.UpdatePassword(ctx, adminID, hash, true); err != nil {
			return nil, access.MapErr(err)
		}
		// รหัสผ่านเปลี่ยนแล้ว session เดิมของบัญชีนั้นต้องใช้ไม่ได้
		_ = s.repo.RevokeAllRefreshTokens(ctx, adminID)
	}

	s.record(ctx, identity, "admin.update", "user", adminID.String(),
		map[string]any{"is_active": isActive}, ip)
	return updated, nil
}

func (s *Service) DeleteAdmin(ctx context.Context, identity middleware.Identity, adminID types.UserID, ip string) error {
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
		// การกระทำในโมดูลนี้เป็นเรื่องของบัญชีผู้ใช้ ไม่ได้เจาะจงสาขา
		// จึงผูกกับสาขาของผู้กระทำเท่าที่ระบุได้ (ดู Identity.AuditBranch)
		BranchID: identity.BranchID, Action: action,
		EntityType: entityType, EntityID: entityID, Detail: detail, IPAddress: ip,
	})
}

func identityOf(u *User) middleware.Identity {
	return middleware.Identity{
		UserID: u.ID, Role: u.Role, Name: u.FullName(), BranchID: u.BranchID,
	}
}
