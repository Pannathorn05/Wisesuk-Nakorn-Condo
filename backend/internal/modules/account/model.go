package account

import (
	"time"

	"backend/internal/shared/types"
)

// User คือบัญชีผู้ใช้ ครอบทั้งสมาชิก ผู้ดูแลระบบ และหัวหน้าผู้ดูแลระบบ
type User struct {
	ID        types.UserID `json:"id"`
	Email     string       `json:"email"`
	FirstName string       `json:"first_name"`
	LastName  string       `json:"last_name"`
	Phone     string       `json:"phone"`
	Role      types.Role   `json:"role"`
	// BranchID คือสาขาที่รับผิดชอบ มีค่าเฉพาะ role = admin ซึ่งดูแลได้สาขาเดียว
	BranchID *types.BranchID `json:"branch_id,omitempty"`
	// BranchName ส่งมาด้วยเพื่อให้ตารางหน้า "จัดการผู้ดูแลระบบ" แสดงชื่อสาขาได้เลย
	// ไม่ต้องยิงถามรายการสาขาเพิ่มเพื่อแปลง id เป็นชื่อ
	BranchName string `json:"branch_name,omitempty"`
	// MustChangePassword = true เมื่อบัญชียังใช้รหัสผ่านตั้งต้นที่ระบบตั้งให้ตอนสร้าง
	MustChangePassword bool       `json:"must_change_password"`
	AvatarURL          string     `json:"avatar_url"`
	IsActive           bool       `json:"is_active"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	// PasswordHash ไม่ถูกส่งออกทาง JSON เด็ดขาด (tag "-")
	PasswordHash string `json:"-"`
}

func (u *User) FullName() string { return u.FirstName + " " + u.LastName }

// Notification คือแจ้งเตือนที่แสดงบนกระดิ่งมุมขวาบน
type Notification struct {
	ID        types.UserID `json:"id"`
	UserID    types.UserID `json:"user_id"`
	Title     string       `json:"title"`
	Body      string       `json:"body"`
	Link      string       `json:"link"`
	ReadAt    *time.Time   `json:"read_at,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

// TokenPair คือผลลัพธ์ของการสมัคร/เข้าสู่ระบบ/ต่ออายุ token
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	User         *User     `json:"user"`
}
