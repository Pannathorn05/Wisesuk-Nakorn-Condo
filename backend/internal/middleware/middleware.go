package middleware

import (
	"log/slog"
	"net"
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
	UserID   uuid.UUID
	Role     types.Role
	Name     string
	BranchID *uuid.UUID // มีค่าเฉพาะ role = admin
}

// IsSuperAdmin ใช้ตัดสินใจว่าจะข้ามการจำกัดสาขาได้หรือไม่
func (i Identity) IsSuperAdmin() bool { return i.Role == types.RoleSuperAdmin }

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

// Authenticate ตรวจ Bearer token และผูก Identity เข้ากับ request context
//
// httpx.Error เรียก c.AbortWithStatusJSON ให้อยู่แล้ว จึงไม่ต้อง c.Abort() ซ้ำ
func Authenticate(mgr *auth.Manager) gin.HandlerFunc {
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

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			httpx.Error(c, httpx.ErrUnauthorized)
			return
		}

		identity := Identity{UserID: userID, Role: claims.Role, Name: claims.Name}
		if claims.BranchID != "" {
			branchID, err := uuid.Parse(claims.BranchID)
			if err != nil {
				httpx.Error(c, httpx.ErrUnauthorized)
				return
			}
			identity.BranchID = &branchID
		}
		// admin ที่ไม่มีสาขาผูกอยู่ ถือว่า token ใช้ไม่ได้
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

// ClientIP อ่าน IP จริงจาก X-Forwarded-For (กรณีอยู่หลัง reverse proxy)
//
// ไม่ใช้ c.ClientIP() ของ gin เพราะผลลัพธ์ขึ้นกับการตั้ง TrustedProxies
// ตัวนี้อ่านตรง ๆ เพื่อให้ค่าที่บันทึกลง activity log เหมือนเดิมทุกประการ
func ClientIP(c *gin.Context) string {
	if fwd := c.GetHeader("X-Forwarded-For"); fwd != "" {
		if first, _, found := strings.Cut(fwd, ","); found {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(fwd)
	}
	if host, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
		return host
	}
	return c.Request.RemoteAddr
}
