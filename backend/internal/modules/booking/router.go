package booking

import (
	"github.com/gin-gonic/gin"

	"backend/internal/database"
	"backend/internal/shared/audit"
	"backend/internal/storage"
)

// Module รวมทุกชั้นของเรื่อง "การจอง": ทำรายการจอง, แจ้งชำระเงิน,
// อนุมัติ/ปฏิเสธ และวันนัดหมายทำสัญญา
type Module struct {
	Repo    *Repository
	Service *Service
	Handler *Handler
}

func New(db *database.TxManager, rooms Rooms, branches Branches, notifier Notifier, rec *audit.Recorder, files *storage.LocalStore) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, db, rooms, branches, notifier, rec)
	return &Module{Repo: repo, Service: svc, Handler: NewHandler(svc, files)}
}

// MemberRoutes — เฉพาะสมาชิกเท่านั้น
func (m *Module) MemberRoutes(r gin.IRoutes) {
	r.POST("/bookings", m.Handler.Create)
	r.GET("/bookings", m.Handler.ListMine)
	r.POST("/bookings/:bookingID/payment", m.Handler.SubmitPayment)
}

// SharedRoutes — เจ้าของการจอง หรือแอดมินของสาขานั้น (ตรวจสิทธิ์ในชั้น service)
func (m *Module) SharedRoutes(r gin.IRoutes) {
	r.GET("/bookings/:bookingID", m.Handler.Get)
	r.POST("/bookings/:bookingID/cancel", m.Handler.Cancel)
}

// AdminRoutes — ใต้ /api/v1/admin
func (m *Module) AdminRoutes(r gin.IRoutes) {
	r.GET("/bookings", m.Handler.ListForAdmin)
	r.POST("/bookings/:bookingID/approve", m.Handler.Approve)
	r.POST("/bookings/:bookingID/reject", m.Handler.Reject)
	r.PUT("/bookings/:bookingID/appointment", m.Handler.SetAppointment)
	r.GET("/members/:memberID/bookings", m.Handler.ListByMember)
}
