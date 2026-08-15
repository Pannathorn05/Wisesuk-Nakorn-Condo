package room

import (
	"github.com/gin-gonic/gin"

	"backend/internal/database"
	"backend/internal/shared/audit"
	"backend/internal/storage"
)

// Module รวมทุกชั้นของเรื่อง "ห้องพัก": ค้นหา, ประเภทห้อง,
// การจัดการห้องของแอดมิน และสถานะห้องแบบ real-time
type Module struct {
	Repo    *Repository
	Service *Service
	Handler *Handler
}

func New(db *database.TxManager, rec *audit.Recorder, files *storage.DBStore) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, rec)
	return &Module{Repo: repo, Service: svc, Handler: NewHandler(svc, files)}
}

// PublicRoutes — ไม่ต้องเข้าสู่ระบบ
func (m *Module) PublicRoutes(r gin.IRoutes) {
	r.GET("/room-types", m.Handler.ListTypes)
	r.GET("/rooms/search", m.Handler.Search)
	r.GET("/rooms/:roomID", m.Handler.Get)
}

// AdminRoutes — ใต้ /api/v1/admin
func (m *Module) AdminRoutes(r gin.IRoutes) {
	r.GET("/rooms", m.Handler.ListForAdmin)
	r.POST("/rooms", m.Handler.Create)
	r.PUT("/rooms/:roomID", m.Handler.Update)
	r.POST("/rooms/:roomID/image", m.Handler.UploadImage)
	r.PATCH("/rooms/:roomID/status", m.Handler.UpdateStatus)
	r.DELETE("/rooms/:roomID", m.Handler.Delete)
}
