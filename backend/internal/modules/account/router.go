package account

import (
	"github.com/gin-gonic/gin"

	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/shared/audit"
)

// Module รวมทุกชั้นของเรื่อง "บัญชีผู้ใช้" ไว้ด้วยกัน:
// สมัคร/เข้าสู่ระบบ, โปรไฟล์, แจ้งเตือน, รายชื่อสมาชิก และการจัดการผู้ดูแลระบบ
type Module struct {
	Repo    *Repository
	Service *Service
	Handler *Handler
}

func New(db *database.TxManager, authMgr *auth.Manager, rec *audit.Recorder, branches BranchChecker) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, authMgr, rec, branches)
	return &Module{Repo: repo, Service: svc, Handler: NewHandler(svc)}
}

// PublicRoutes — ไม่ต้องเข้าสู่ระบบ
func (m *Module) PublicRoutes(r gin.IRoutes) {
	r.POST("/auth/register", m.Handler.Register)
	r.POST("/auth/login", m.Handler.Login)
	r.POST("/auth/refresh", m.Handler.Refresh)
	r.POST("/auth/logout", m.Handler.Logout)
}

// AuthedRoutes — เข้าสู่ระบบแล้ว ทุกสิทธิ์
func (m *Module) AuthedRoutes(r gin.IRoutes) {
	r.GET("/me", m.Handler.Me)
	r.PUT("/me", m.Handler.UpdateMe)
	r.POST("/me/password", m.Handler.ChangePassword)
	r.GET("/me/notifications", m.Handler.ListNotifications)
	r.POST("/me/notifications/read", m.Handler.MarkNotificationsRead)
}

// AdminRoutes — ใต้ /api/v1/admin
func (m *Module) AdminRoutes(r gin.IRoutes) {
	r.GET("/members", m.Handler.ListMembers)
}

// SuperAdminRoutes — ใต้ /api/v1/superadmin
func (m *Module) SuperAdminRoutes(r gin.IRoutes) {
	r.GET("/staff", m.Handler.ListStaff)
	r.POST("/staff", m.Handler.CreateAdmin)
	r.PUT("/staff/:userID", m.Handler.UpdateAdmin)
	r.DELETE("/staff/:userID", m.Handler.DeleteAdmin)
}
