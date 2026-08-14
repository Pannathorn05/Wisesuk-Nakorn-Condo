package reporting

import (
	"github.com/gin-gonic/gin"

	"backend/internal/database"
	"backend/internal/modules/booking"
	"backend/internal/modules/room"
)

// Module รวมทุกชั้นของเรื่อง "รายงาน" ไว้ด้วยกัน
type Module struct {
	Repo    *Repository
	Service *Service
	Handler *Handler
}

func New(db *database.TxManager, rooms *room.Service, bookings *booking.Service) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, rooms, bookings)
	return &Module{Repo: repo, Service: svc, Handler: NewHandler(svc)}
}

// AdminRoutes — ใต้ /api/v1/admin (admin + superadmin)
func (m *Module) AdminRoutes(r gin.IRoutes) {
	r.GET("/dashboard", m.Handler.Dashboard)
	r.GET("/activity-logs", m.Handler.ListActivityLogs)
}
