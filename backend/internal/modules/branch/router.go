package branch

import (
	"github.com/gin-gonic/gin"

	"backend/internal/database"
	"backend/internal/shared/audit"
)

// Module รวมทุกชั้นของเรื่อง "สาขา": รายละเอียด, รูปภาพ,
// สิ่งอำนวยความสะดวก และสถานที่ใกล้เคียง
type Module struct {
	Repo    *Repository
	Service *Service
	Handler *Handler
}

func New(db *database.TxManager, rec *audit.Recorder) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, rec)
	return &Module{Repo: repo, Service: svc, Handler: NewHandler(svc)}
}

// PublicRoutes — ไม่ต้องเข้าสู่ระบบ
func (m *Module) PublicRoutes(r gin.IRoutes) {
	r.GET("/branches", m.Handler.List)
	r.GET("/branches/:branchID", m.Handler.Get)
	r.GET("/amenities", m.Handler.ListAmenities)
}

// AdminRoutes — ใต้ /api/v1/admin
func (m *Module) AdminRoutes(r gin.IRoutes) {
	r.PUT("/branch", m.Handler.Update)
	r.PUT("/branch/amenities", m.Handler.SetAmenities)
	r.PUT("/branch/nearby", m.Handler.ReplaceNearby)
	r.POST("/branch/images", m.Handler.AddImage)
	r.DELETE("/branch/images/:imageID", m.Handler.DeleteImage)
}

// SuperAdminRoutes — ใต้ /api/v1/superadmin
func (m *Module) SuperAdminRoutes(r gin.IRoutes) {
	r.GET("/branches", m.Handler.ListAll)
}
