package account

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/database"
	"backend/internal/shared/types"
)

type Repository struct{ db *database.TxManager }

func NewRepository(db *database.TxManager) *Repository { return &Repository{db: db} }

const userColumns = `
	u.id, u.email, u.password_hash, u.first_name, u.last_name, u.phone,
	u.role, u.branch_id, u.avatar_url, u.is_active, u.last_login_at,
	u.created_at, u.updated_at, COALESCE(b.name, '')`

const userJoins = ` FROM users u LEFT JOIN branches b ON b.id = u.branch_id `

func scanUser(row interface{ Scan(...any) error }) (*User, error) {
	var u User
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone,
		&u.Role, &u.BranchID, &u.AvatarURL, &u.IsActive, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt, &u.BranchName,
	)
	if err != nil {
		return nil, database.NormalizeErr(err)
	}
	return &u, nil
}

type CreateUserParams struct {
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Phone        string
	Role         types.Role
	BranchID     *uuid.UUID
}

func (r *Repository) Create(ctx context.Context, p CreateUserParams) (*User, error) {
	const q = `
		INSERT INTO users (email, password_hash, first_name, last_name, phone, role, branch_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	var id uuid.UUID
	err := r.db.Executor(ctx).QueryRow(ctx, q,
		normalizeEmail(p.Email), p.PasswordHash,
		p.FirstName, p.LastName, p.Phone, p.Role, p.BranchID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	q := `SELECT ` + userColumns + userJoins + ` WHERE u.id = $1`
	return scanUser(r.db.Executor(ctx).QueryRow(ctx, q, id))
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	q := `SELECT ` + userColumns + userJoins + ` WHERE u.email = $1`
	return scanUser(r.db.Executor(ctx).QueryRow(ctx, q, normalizeEmail(email)))
}

func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.Executor(ctx).QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`, normalizeEmail(email)).Scan(&exists)
	return exists, err
}

type UpdateProfileParams struct {
	FirstName string
	LastName  string
	Phone     string
	AvatarURL *string
}

func (r *Repository) UpdateProfile(ctx context.Context, id uuid.UUID, p UpdateProfileParams) (*User, error) {
	const q = `
		UPDATE users SET
			first_name = $2, last_name = $3, phone = $4,
			avatar_url = COALESCE($5, avatar_url),
			updated_at = now()
		WHERE id = $1`
	tag, err := r.db.Executor(ctx).Exec(ctx, q, id, p.FirstName, p.LastName, p.Phone, p.AvatarURL)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, database.ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	tag, err := r.db.Executor(ctx).Exec(ctx,
		`UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`, id, hash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (r *Repository) TouchLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Executor(ctx).Exec(ctx, `UPDATE users SET last_login_at = now() WHERE id = $1`, id)
	return err
}

const memberSearchClause = `
	($1 = '' OR u.first_name ILIKE '%' || $1 || '%'
	         OR u.last_name  ILIKE '%' || $1 || '%'
	         OR u.email      ILIKE '%' || $1 || '%'
	         OR u.phone      ILIKE '%' || $1 || '%')`

// ListMembers ใช้ในหน้า "สมาชิก" ของแอดมิน ค้นได้ด้วยชื่อ/อีเมล/เบอร์โทร
func (r *Repository) ListMembers(ctx context.Context, search string, limit, offset int) ([]User, int, error) {
	search = strings.TrimSpace(search)
	exec := r.db.Executor(ctx)

	var total int
	if err := exec.QueryRow(ctx,
		`SELECT COUNT(*) FROM users u WHERE u.role = 'member' AND`+memberSearchClause,
		search).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `SELECT ` + userColumns + userJoins +
		` WHERE u.role = 'member' AND` + memberSearchClause +
		` ORDER BY u.created_at DESC LIMIT $2 OFFSET $3`
	rows, err := exec.Query(ctx, q, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users, err := collectUsers(rows)
	return users, total, err
}

// ListStaff คืน admin + superadmin สำหรับหน้า "จัดการผู้ดูแลระบบ"
func (r *Repository) ListStaff(ctx context.Context) ([]User, error) {
	q := `SELECT ` + userColumns + userJoins +
		` WHERE u.role IN ('admin', 'superadmin') ORDER BY u.role DESC, u.created_at`
	rows, err := r.db.Executor(ctx).Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectUsers(rows)
}

type UpdateStaffParams struct {
	FirstName string
	LastName  string
	Phone     string
	BranchID  *uuid.UUID
	IsActive  bool
}

func (r *Repository) UpdateStaff(ctx context.Context, id uuid.UUID, p UpdateStaffParams) (*User, error) {
	const q = `
		UPDATE users SET
			first_name = $2, last_name = $3, phone = $4,
			branch_id  = CASE WHEN role = 'admin' THEN $5::uuid ELSE NULL END,
			is_active  = $6,
			updated_at = now()
		WHERE id = $1 AND role IN ('admin', 'superadmin')`
	tag, err := r.db.Executor(ctx).Exec(ctx, q, id, p.FirstName, p.LastName, p.Phone, p.BranchID, p.IsActive)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, database.ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) DeleteAdmin(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Executor(ctx).Exec(ctx, `DELETE FROM users WHERE id = $1 AND role = 'admin'`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

// ---------------------------------------------------------------- refresh tokens

func (r *Repository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Executor(ctx).Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt)
	return err
}

// ConsumeRefreshToken ทำ rotation: ตรวจว่า token ยังใช้ได้แล้วเพิกถอนทันทีในคำสั่งเดียว
// การรวมเป็นคำสั่งเดียวทำให้ token เดิมถูกใช้ซ้ำไม่ได้แม้มีคำขอเข้ามาพร้อมกัน
func (r *Repository) ConsumeRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	const q = `
		UPDATE refresh_tokens SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
		RETURNING user_id`
	var userID uuid.UUID
	if err := r.db.Executor(ctx).QueryRow(ctx, q, tokenHash).Scan(&userID); err != nil {
		return uuid.Nil, database.NormalizeErr(err)
	}
	return userID, nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.Executor(ctx).Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`,
		tokenHash)
	return err
}

func (r *Repository) RevokeAllRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Executor(ctx).Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`,
		userID)
	return err
}

// ---------------------------------------------------------------- notifications

func (r *Repository) CreateNotification(ctx context.Context, userID uuid.UUID, title, body, link string) error {
	_, err := r.db.Executor(ctx).Exec(ctx,
		`INSERT INTO notifications (user_id, title, body, link) VALUES ($1, $2, $3, $4)`,
		userID, title, body, link)
	return err
}

func (r *Repository) ListNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]Notification, int, error) {
	exec := r.db.Executor(ctx)

	var unread int
	if err := exec.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`,
		userID).Scan(&unread); err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	rows, err := exec.Query(ctx,
		`SELECT id, user_id, title, body, link, read_at, created_at
		 FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []Notification{}
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Link, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, n)
	}
	return out, unread, rows.Err()
}

// MarkNotificationsRead อ่านทีละรายการ หรือทั้งหมดเมื่อ id = nil
func (r *Repository) MarkNotificationsRead(ctx context.Context, userID uuid.UUID, id *uuid.UUID) error {
	_, err := r.db.Executor(ctx).Exec(ctx,
		`UPDATE notifications SET read_at = now()
		 WHERE user_id = $1 AND read_at IS NULL AND ($2::uuid IS NULL OR id = $2)`,
		userID, id)
	return err
}

// ---------------------------------------------------------------- helpers

func normalizeEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func collectUsers(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]User, error) {
	users := []User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}
