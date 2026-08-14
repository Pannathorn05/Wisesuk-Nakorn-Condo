// Package routes ประกอบ route ทั้งหมดของ API เข้าเป็น http.Handler เดียว
//
// นี่คือที่เดียวที่รู้โครงสร้าง URL ทั้งระบบ (/api/v1, /admin, /superadmin)
// และรู้ว่า module ไหนขึ้นที่ระดับสิทธิ์ใด
// ส่วน endpoint รายตัวยังอยู่ใน router.go ของแต่ละ module — เพิ่ม endpoint ไม่ต้องแก้ไฟล์นี้
package routes

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"backend/internal/httpx"
	mw "backend/internal/middleware"
	"backend/internal/server"
	"backend/internal/shared/types"
)

// New ผูก route ตามสิทธิ์ 4 ระดับใน prototype:
// Guest (สาธารณะ) / Member / Admin / Super Admin
func New(s *server.Server) http.Handler {
	if s.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// ใช้ gin.New() ไม่ใช่ gin.Default() เพราะเรามี Logger/Recoverer ของตัวเอง
	// ที่เขียน log ผ่าน slog และตอบ error เป็น JSON รูปแบบเดียวกับทั้งระบบ
	r := gin.New()

	// ไม่ให้ gin ตีความ proxy header เอง — middleware.ClientIP อ่าน X-Forwarded-For เองอยู่แล้ว
	_ = r.SetTrustedProxies(nil)
	r.HandleMethodNotAllowed = true

	r.Use(mw.RequestID())
	r.Use(mw.Recoverer())
	r.Use(mw.Logger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     s.Config.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300 * time.Second,
	}))

	r.NoRoute(func(c *gin.Context) {
		httpx.Error(c, httpx.ErrNotFound)
	})
	r.NoMethod(func(c *gin.Context) {
		httpx.Error(c, httpx.NewError(http.StatusMethodNotAllowed,
			"method_not_allowed", "ไม่รองรับ HTTP method นี้"))
	})

	r.GET("/health", func(c *gin.Context) {
		httpx.JSON(c, http.StatusOK, gin.H{"status": "healthy", "time": time.Now().UTC()})
	})

	// ไฟล์ที่อัปโหลด (สลิป) เสิร์ฟแบบ read-only
	r.GET("/uploads/*filepath", serveUpload(s.Config.UploadDir))

	v1 := r.Group("/api/v1")

	// ============================================ สาธารณะ (Guest)
	public := v1.Group("")
	s.Account.PublicRoutes(public)
	s.Branch.PublicRoutes(public)
	s.Room.PublicRoutes(public)

	// ============================================ ต้องเข้าสู่ระบบ
	authed := v1.Group("")
	authed.Use(mw.Authenticate(s.Auth))

	s.Account.AuthedRoutes(authed)
	s.Booking.SharedRoutes(authed)

	// -------------------------------------- สมาชิก (Member)
	member := authed.Group("")
	member.Use(mw.RequireRole(types.RoleMember))
	s.Booking.MemberRoutes(member)

	// -------------------------------------- ผู้ดูแลระบบ (Admin + Super Admin)
	admin := authed.Group("/admin")
	admin.Use(mw.RequireRole(types.RoleAdmin, types.RoleSuperAdmin))

	s.Reporting.AdminRoutes(admin)
	s.Booking.AdminRoutes(admin)
	s.Room.AdminRoutes(admin)
	s.Branch.AdminRoutes(admin)
	s.Account.AdminRoutes(admin)

	// -------------------------------------- หัวหน้าผู้ดูแลระบบ (Super Admin)
	superAdmin := authed.Group("/superadmin")
	superAdmin.Use(mw.RequireRole(types.RoleSuperAdmin))

	s.Account.SuperAdminRoutes(superAdmin)
	s.Branch.SuperAdminRoutes(superAdmin)

	return r
}

// serveUpload เสิร์ฟไฟล์ใน uploadDir แบบ read-only และกันไม่ให้ไล่ดูรายชื่อไฟล์ทั้งโฟลเดอร์
func serveUpload(uploadDir string) gin.HandlerFunc {
	fileServer := http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir)))
	return func(c *gin.Context) {
		p := c.Param("filepath")
		if p == "" || p == "/" || p[len(p)-1] == '/' {
			httpx.Error(c, httpx.ErrNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
