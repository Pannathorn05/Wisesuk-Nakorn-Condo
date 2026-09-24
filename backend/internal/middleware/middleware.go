package middleware

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/auth"
	"backend/internal/httpx"
	"backend/internal/shared/types"
)

// key ที่ใช้เก็บของใน gin.Context — ตั้งเป็น const กันพิมพ์ผิด
const (
	ctxKeyIdentity  = "identity"
	ctxKeyRequestID = "request_id"
)

// Identity คือผู้เรียก API ที่ผ่านการยืนยันตัวตนแล้ว
type Identity struct {
	UserID types.UserID
	Role   types.Role
	Name   string
	// BranchID คือสาขาที่รับผิดชอบ มีค่าเฉพาะ role = admin ซึ่งผูกกับสาขาเดียวเสมอ
	// (super admin เห็นทุกสาขาอยู่แล้วจึงไม่ผูกกับสาขาใด ดู shared/access)
	BranchID *types.BranchID
}

// IsSuperAdmin ใช้ตัดสินใจว่าจะข้ามการจำกัดสาขาได้หรือไม่
func (i Identity) IsSuperAdmin() bool { return i.Role == types.RoleSuperAdmin }

// HasBranch บอกว่าผู้เรียกมีสิทธิ์แตะข้อมูลของสาขานี้หรือไม่
//
// super admin ตอบจริงเสมอ ส่วน admin ต้องเป็นสาขาของตัวเองเท่านั้น
// สมาชิกไม่มีสาขาจึงตอบเท็จเสมอ
func (i Identity) HasBranch(id types.BranchID) bool {
	if i.IsSuperAdmin() {
		return true
	}
	return i.BranchID != nil && *i.BranchID == id
}

// IdentityFrom ดึงผู้ใช้จาก context ต้องเรียกหลัง Authenticate เท่านั้น
func IdentityFrom(c *gin.Context) (Identity, bool) {
	v, exists := c.Get(ctxKeyIdentity)
	if !exists {
		return Identity{}, false
	}
	id, ok := v.(Identity)
	return id, ok
}

// MustIdentity ใช้ใน handler ที่อยู่หลัง Authenticate แล้วแน่นอน
func MustIdentity(c *gin.Context) Identity {
	id, ok := IdentityFrom(c)
	if !ok {
		panic("middleware: เรียก MustIdentity นอก route ที่ผ่าน Authenticate")
	}
	return id
}

func RequestIDFrom(c *gin.Context) string { return c.GetString(ctxKeyRequestID) }

// ---------------------------------------------------------------- auth

// Accounts คือสิ่งเดียวที่ Authenticate ต้องรู้เกี่ยวกับบัญชีผู้ใช้
// ประกาศไว้ฝั่งผู้ใช้งาน (account import middleware อยู่แล้ว จะ import กลับไม่ได้)
type Accounts interface {
	// CurrentIdentity คืนตัวตนปัจจุบันของผู้ใช้จากฐานข้อมูล
	// ok = false เมื่อบัญชีไม่มีอยู่ ถูกลบ หรือถูกระงับ · err คือปัญหาของระบบ ไม่ใช่ของผู้ใช้
	CurrentIdentity(ctx context.Context, id types.UserID) (identity Identity, ok bool, err error)
}

// Authenticate ตรวจ Bearer token และผูก Identity เข้ากับ request context
//
// token ใช้ยืนยันแค่ว่า "เป็นใคร" (claim sub) ส่วนสิทธิ์ สาขา และสถานะบัญชีอ่านจากฐานข้อมูล
// ทุกคำขอ การลบ ระงับ หรือย้ายสาขาผู้ดูแลจึงมีผลทันที ไม่ต้องรอ access token หมดอายุ
// (เดิมเชื่อ role/branch_id ใน claim ทำให้บัญชีที่ถูกลบยังใช้งานต่อได้อีกสูงสุด 30 นาที)
//
// httpx.Error เรียก c.AbortWithStatusJSON ให้อยู่แล้ว จึงไม่ต้อง c.Abort() ซ้ำ
func Authenticate(mgr *auth.Manager, accounts Accounts) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			httpx.Error(c, httpx.ErrUnauthorized)
			return
		}

		claims, err := mgr.ParseAccessToken(strings.TrimSpace(token))
		if err != nil {
			httpx.Error(c, httpx.ErrUnauthorized)
			return
		}

		userID, err := types.ParseID[types.User](claims.Subject)
		if err != nil {
			httpx.Error(c, httpx.ErrUnauthorized)
			return
		}

		identity, ok, err := accounts.CurrentIdentity(c.Request.Context(), userID)
		if err != nil {
			httpx.Error(c, httpx.ErrInternal.Wrap(err))
			return
		}
		if !ok {
			httpx.Error(c, httpx.ErrUnauthorized)
			return
		}
		// admin ที่ไม่มีสาขาผูกอยู่ ถือว่าใช้งานไม่ได้ (DB มี CHECK กันไว้อยู่แล้ว เช็คซ้ำกันพลาด)
		if identity.Role == types.RoleAdmin && identity.BranchID == nil {
			httpx.Error(c, httpx.ErrUnauthorized)
			return
		}

		c.Set(ctxKeyIdentity, identity)
		c.Next()
	}
}

// RequireRole อนุญาตเฉพาะ role ที่ระบุ ต้องวางต่อจาก Authenticate
func RequireRole(roles ...types.Role) gin.HandlerFunc {
	allowed := make(map[types.Role]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		identity, ok := IdentityFrom(c)
		if !ok {
			httpx.Error(c, httpx.ErrUnauthorized)
			return
		}
		if !allowed[identity.Role] {
			httpx.Error(c, httpx.ErrForbidden)
			return
		}
		c.Next()
	}
}

// ---------------------------------------------------------------- observability

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(ctxKeyRequestID, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// Logger แทน gin.Logger() เพื่อให้ log เป็น slog รูปแบบเดียวกับส่วนอื่นของระบบ
//
// gin เก็บ status กับจำนวนไบต์ไว้ใน c.Writer ให้แล้ว จึงไม่ต้องห่อ ResponseWriter เองเหมือนตอนใช้ net/http
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"bytes", c.Writer.Size(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", RequestIDFrom(c),
			"ip", ClientIP(c),
		)
	}
}

// Recoverer กัน panic ไม่ให้ทำให้ทั้ง server ล้ม และไม่ส่ง stack trace ออกไปหา client
//
// ไม่ใช้ gin.Recovery() เพราะมันตอบ 500 เป็น text เปล่า ๆ ไม่ใช่ JSON รูปแบบเดียวกับ error อื่น
func Recoverer() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered",
					"error", rec, "path", c.Request.URL.Path,
					"request_id", RequestIDFrom(c))
				if c.Writer.Written() {
					c.Abort()
					return
				}
				httpx.Error(c, httpx.ErrInternal)
			}
		}()
		c.Next()
	}
}

// ClientIP คือ IP ของผู้เรียกที่บันทึกลง activity log และ log ของ request
//
// ใช้ c.ClientIP() ของ gin ซึ่งเชื่อ X-Forwarded-For / X-Real-IP ก็ต่อเมื่อ connection
// มาจาก proxy ที่ตั้งไว้ใน TRUSTED_PROXIES (ดู routes.New) และเลือก IP ขวาสุดที่ไม่ใช่ proxy
// ไม่ใช่ตัวซ้ายสุด ซึ่งเป็นค่าที่ client เขียนมาเองได้
//
// ของเดิมอ่านค่าแรกของ X-Forwarded-For ตรง ๆ ทุกคำขอ ใครก็ส่ง header มาปลอม IP ใน log ได้
func ClientIP(c *gin.Context) string {
	return c.ClientIP()
}
