package account

import (
	"context"
	"strings"
	"time"

	"backend/internal/database"
	"backend/internal/shared/types"
)

type Repository struct{ db *database.TxManager }

func NewRepository(db *database.TxManager) *Repository { return &Repository{db: db} }

const userColumns = `
	u.id, u.email, u.password_hash, u.first_name, u.last_name, u.phone,
	u.role, u.branch_id, u.avatar_url, u.is_active, u.must_change_password,
	u.last_login_at, u.created_at, u.updated_at, COALESCE(b.name, '')`

const userJoins = ` FROM users u LEFT JOIN branches b ON b.id = u.branch_id `

// notDeleted ต่อท้ายทุก WHERE ที่อ่านผู้ใช้ เพื่อไม่ให้บัญชีที่ถูก soft delete
// (ดู DeleteAdmin) โผล่กลับมาที่ไหนอีก ไม่ว่าจะเป็นล็อกอิน โปรไฟล์ หรือหน้ารายชื่อ
const notDeleted = ` AND u.deleted_at IS NULL`

func scanUser(row interface{ Scan(...any) error }) (*User, error) {
	var u User
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone,
		&u.Role, &u.BranchID, &u.AvatarURL, &u.IsActive, &u.MustChangePassword,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt, &u.BranchName,
	)
	if err != nil {
		return nil, database.NormalizeErr(err)
	}
	return &u, nil
}

type CreateUserParams struct {
	Email              string
	PasswordHash       string
	FirstName          string
	LastName           string
	Phone              string
	Role               types.Role
	BranchID           *types.BranchID
	MustChangePassword bool
}

func (r *Repository) Create(ctx context.Context, p CreateUserParams) (*User, error) {
	const q = `
		INSERT INTO users (email, password_hash, first_name, last_name, phone, role, branch_id, must_change_password)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`
	var id types.UserID
	err := r.db.Executor(ctx).QueryRow(ctx, q,
		normalizeEmail(p.Email), p.PasswordHash,
		p.FirstName, p.LastName, p.Phone, p.Role, p.BranchID, p.MustChangePassword,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) GetByID(ctx context.Context, id types.UserID) (*User, error) {
	q := `SELECT ` + userColumns + userJoins + ` WHERE u.id = $1` + notDeleted
	return scanUser(r.db.Executor(ctx).QueryRow(ctx, q, id))
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	q := `SELECT ` + userColumns + userJoins + ` WHERE u.email = $1` + notDeleted
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

func (r *Repository) UpdateProfile(ctx context.Context, id types.UserID, p UpdateProfileParams) (*User, error) {
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

// UpdatePassword ตั้งรหัสผ่านใหม่
//
// mustChange บอกว่ารหัสที่ตั้งนี้เป็นรหัสตั้งต้นที่คนอื่นตั้งให้หรือไม่ —
// ผู้ใช้เปลี่ยนรหัสของตัวเองส่ง false (ปลดธง) ส่วนหัวหน้าผู้ดูแลตั้งรหัสให้คนอื่นส่ง true
func (r *Repository) UpdatePassword(ctx context.Context, id types.UserID, hash string, mustChange bool) error {
	tag, err := r.db.Executor(ctx).Exec(ctx,
		`UPDATE users SET password_hash = $2, must_change_password = $3, updated_at = now() WHERE id = $1`,
		id, hash, mustChange)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (r *Repository) TouchLastLogin(ctx context.Context, id types.UserID) error {
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
		`SELECT COUNT(*) FROM users u WHERE u.role = 'member' AND`+memberSearchClause+notDeleted,
		search).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `SELECT ` + userColumns + userJoins +
		` WHERE u.role = 'member' AND` + memberSearchClause + notDeleted +
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
		` WHERE u.role IN ('admin', 'superadmin')` + notDeleted +
		` ORDER BY u.role DESC, u.created_at`
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
	BranchID  *types.BranchID
	IsActive  bool
}

func (r *Repository) UpdateStaff(ctx context.Context, id types.UserID, p UpdateStaffParams) (*User, error) {
	// หัวหน้าผู้ดูแลต้องไม่ผูกกับสาขา จึงบังคับเป็น NULL ตรงนี้แทนที่จะไว้ใจค่าที่ส่งมา
	const q = `
		UPDATE users SET
			first_name = $2, last_name = $3, phone = $4,
			branch_id  = CASE WHEN role = 'admin' THEN $5::bigint ELSE NULL END,
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

// DeleteAdmin คือ soft delete: ปักธง deleted_at และปิดบัญชี แถวยังอยู่ใน DB
// เพื่อไม่ให้ activity_logs.actor_id, bookings.reviewed_by, payments.reviewed_by
// ที่อ้างอิงไว้หลุดหาย (ดู migration 0007_staff_soft_delete.sql)
func (r *Repository) DeleteAdmin(ctx context.Context, id types.UserID) error {
	tag, err := r.db.Executor(ctx).Exec(ctx,
		`UPDATE users SET deleted_at = now(), is_active = false, updated_at = now()
		 WHERE id = $1 AND role = 'admin' AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

// ---------------------------------------------------------------- refresh tokens

func (r *Repository) StoreRefreshToken(ctx context.Context, userID types.UserID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Executor(ctx).Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt)
	return err
}

// ConsumeRefreshToken ทำ rotation: ตรวจว่า token ยังใช้ได้แล้วเพิกถอนทันทีในคำสั่งเดียว
// การรวมเป็นคำสั่งเดียวทำให้ token เดิมถูกใช้ซ้ำไม่ได้แม้มีคำขอเข้ามาพร้อมกัน
func (r *Repository) ConsumeRefreshToken(ctx context.Context, tokenHash string) (types.UserID, error) {
	const q = `
		UPDATE refresh_tokens SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
		RETURNING user_id`
	var userID types.UserID
	if err := r.db.Executor(ctx).QueryRow(ctx, q, tokenHash).Scan(&userID); err != nil {
		return 0, database.NormalizeErr(err)
	}
	return userID, nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.Executor(ctx).Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`,
		tokenHash)
	return err
}

func (r *Repository) RevokeAllRefreshTokens(ctx context.Context, userID types.UserID) error {
	_, err := r.db.Executor(ctx).Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`,
		userID)
	return err
}

// ---------------------------------------------------------------- notifications

func (r *Repository) CreateNotification(ctx context.Context, userID types.UserID, title, body, link string) error {
	_, err := r.db.Executor(ctx).Exec(ctx,
		`INSERT INTO notifications (user_id, title, body, link) VALUES ($1, $2, $3, $4)`,
		userID, title, body, link)
	return err
}

func (r *Repository) ListNotifications(ctx context.Context, userID types.UserID, limit int) ([]Notification, int, error) {
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
func (r *Repository) MarkNotificationsRead(ctx context.Context, userID types.UserID, id *types.NotificationID) error {
	_, err := r.db.Executor(ctx).Exec(ctx,
		`UPDATE notifications SET read_at = now()
		 WHERE user_id = $1 AND read_at IS NULL AND ($2::bigint IS NULL OR id = $2)`,
		userID, id)
	return err
}

// ---------------------------------------------------------------- oauth identity

// GetByIdentity หาผู้ใช้จากบัญชีภายนอกที่ผูกไว้
//
// จับคู่ด้วย provider_user_id ไม่ใช่อีเมล เพราะผู้ใช้เปลี่ยนอีเมลที่ Google/Facebook ได้
// แต่รหัสผู้ใช้ฝั่ง provider ไม่เปลี่ยน
func (r *Repository) GetByIdentity(ctx context.Context, provider, providerUserID string) (*User, error) {
	q := `SELECT ` + userColumns + userJoins +
		` JOIN user_identities i ON i.user_id = u.id
		  WHERE i.provider = $1 AND i.provider_user_id = $2` + notDeleted
	return scanUser(r.db.Executor(ctx).QueryRow(ctx, q, provider, providerUserID))
}

// LinkIdentity ผูกบัญชีภายนอกเข้ากับผู้ใช้ที่มีอยู่
//
// ON CONFLICT DO NOTHING รองรับกรณีสองคำขอของคนเดียวกันวิ่งเข้ามาพร้อมกัน
// (ผู้ใช้กดปุ่มซ้ำ) ให้ผลลัพธ์เหมือนกันทั้งสองครั้งแทนที่จะพังไปหนึ่งครั้ง
func (r *Repository) LinkIdentity(ctx context.Context, userID types.UserID, provider, providerUserID, email string) error {
	const q = `
		INSERT INTO user_identities (user_id, provider, provider_user_id, email)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, provider_user_id) DO NOTHING`
	_, err := r.db.Executor(ctx).Exec(ctx, q, userID, provider, providerUserID, normalizeEmail(email))
	return err
}

// ListIdentities บอกว่าผู้ใช้คนนี้ผูกบัญชีภายนอกเจ้าใดไว้บ้าง
func (r *Repository) ListIdentities(ctx context.Context, userID types.UserID) ([]string, error) {
	rows, err := r.db.Executor(ctx).Query(ctx,
		`SELECT provider FROM user_identities WHERE user_id = $1 ORDER BY provider`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	providers := []string{}
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		providers = append(providers, p)
	}
	return providers, rows.Err()
}

// ---------------------------------------------------------------- oauth exchange code

func (r *Repository) StoreOAuthCode(ctx context.Context, userID types.UserID, codeHash string, expiresAt time.Time) error {
	_, err := r.db.Executor(ctx).Exec(ctx,
		`INSERT INTO oauth_exchange_codes (user_id, code_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, codeHash, expiresAt)
	return err
}

// ConsumeOAuthCode ตรวจและปิดโค้ดในคำสั่งเดียวด้วยเหตุผลเดียวกับ ConsumeRefreshToken
// คือทำให้โค้ดใบเดิมถูกใช้ซ้ำไม่ได้แม้มีคำขอเข้ามาพร้อมกัน
func (r *Repository) ConsumeOAuthCode(ctx context.Context, codeHash string) (types.UserID, error) {
	const q = `
		UPDATE oauth_exchange_codes SET consumed_at = now()
		WHERE code_hash = $1 AND consumed_at IS NULL AND expires_at > now()
		RETURNING user_id`
	var userID types.UserID
	if err := r.db.Executor(ctx).QueryRow(ctx, q, codeHash).Scan(&userID); err != nil {
		return 0, database.NormalizeErr(err)
	}
	return userID, nil
}

// ---------------------------------------------------------------- login rate limit

// LoginFailures คือจำนวนครั้งที่ login ล้มเหลวจาก IP หนึ่งภายในหน้าต่างเวลา
// พร้อมจำนวนวินาทีที่ความล้มเหลวครั้งเก่าสุดจะหลุดจากหน้าต่าง (ใช้ตอบ Retry-After)
type LoginFailures struct {
	ForEmail      int
	ForEmailRetry int
	ForIP         int
	ForIPRetry    int
}

// CountLoginFailures นับความล้มเหลวจาก ip ภายใน window วินาที — ทั้งหมด และเฉพาะอีเมลนี้
// ใช้ query เดียวเพราะตัวนับรายคู่ (อีเมล, IP) เป็นส่วนย่อยของตัวนับราย IP อยู่แล้ว
// เวลาทั้งหมดคำนวณที่ DB จึงไม่ขึ้นกับนาฬิกาของเครื่องที่รัน API
func (r *Repository) CountLoginFailures(ctx context.Context, email, ip string, window time.Duration) (LoginFailures, error) {
	const q = `
		SELECT
			count(*) FILTER (WHERE email = $1),
			COALESCE(ceil(extract(epoch FROM
				min(created_at) FILTER (WHERE email = $1) + make_interval(secs => $3) - now())), 0)::int,
			count(*),
			COALESCE(ceil(extract(epoch FROM min(created_at) + make_interval(secs => $3) - now())), 0)::int
		FROM login_attempts
		WHERE ip = $2 AND created_at > now() - make_interval(secs => $3)`
	var f LoginFailures
	err := r.db.Executor(ctx).QueryRow(ctx, q, normalizeEmail(email), ip, window.Seconds()).
		Scan(&f.ForEmail, &f.ForEmailRetry, &f.ForIP, &f.ForIPRetry)
	return f, err
}

// RecordLoginFailure บันทึกความล้มเหลวหนึ่งครั้ง แล้วลบแถวที่หลุดหน้าต่างไปแล้วทิ้ง
// ตารางจึงเก็บแค่ข้อมูลของ window ล่าสุด ไม่โตไปเรื่อย ๆ
func (r *Repository) RecordLoginFailure(ctx context.Context, email, ip string, window time.Duration) error {
	exec := r.db.Executor(ctx)
	if _, err := exec.Exec(ctx,
		`INSERT INTO login_attempts (email, ip) VALUES ($1, $2)`, normalizeEmail(email), ip); err != nil {
		return err
	}
	_, err := exec.Exec(ctx,
		`DELETE FROM login_attempts WHERE created_at < now() - make_interval(secs => $1)`, window.Seconds())
	return err
}

// ClearLoginFailures ล้างความล้มเหลวของคู่ (อีเมล, IP) หลัง login สำเร็จ
// ตัวนับราย IP ของอีเมลอื่นยังอยู่ — สำเร็จกับบัญชีตัวเองไม่ได้ลบร่องรอยการไล่เดาบัญชีอื่น
func (r *Repository) ClearLoginFailures(ctx context.Context, email, ip string) error {
	_, err := r.db.Executor(ctx).Exec(ctx,
		`DELETE FROM login_attempts WHERE email = $1 AND ip = $2`, normalizeEmail(email), ip)
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
