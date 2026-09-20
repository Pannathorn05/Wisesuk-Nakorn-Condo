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

func New(db *database.TxManager, authMgr *auth.Manager, rec *audit.Recorder, branches BranchChecker,
	provider OAuthProvider, web WebConfig, staffDefaultPassword string) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, db, authMgr, rec, branches, staffDefaultPassword)
	return &Module{Repo: repo, Service: svc, Handler: NewHandler(svc, provider, web)}
}

// PublicRoutes — ไม่ต้องเข้าสู่ระบบ
func (m *Module) PublicRoutes(r gin.IRoutes) {
	r.POST("/auth/register", m.Handler.Register)
	r.POST("/auth/login", m.Handler.Login)
	r.POST("/auth/refresh", m.Handler.Refresh)
	r.POST("/auth/logout", m.Handler.Logout)

	// เข้าสู่ระบบด้วยบัญชีภายนอก — สองเส้นทางแรกเป็น GET เพราะเบราว์เซอร์เดินทางเอง
	// ส่วน exchange เป็น POST เพราะ frontend เรียกด้วย fetch
	r.GET("/auth/oauth", m.Handler.OAuthProviders)
	r.GET("/auth/oauth/:provider", m.Handler.OAuthStart)
	r.GET("/auth/oauth/:provider/callback", m.Handler.OAuthCallback)
	r.POST("/auth/oauth/exchange", m.Handler.OAuthExchange)
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
