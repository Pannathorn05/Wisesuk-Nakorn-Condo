package booking

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/shared/types"
	"backend/internal/storage"
)

type Handler struct {
	svc   *Service
	files *storage.LocalStore
}

func NewHandler(svc *Service, files *storage.LocalStore) *Handler {
	return &Handler{svc: svc, files: files}
}

// ---------------------------------------------------------------- member

// POST /api/v1/bookings — แบบฟอร์มจองห้องพัก (รายวัน/รายเดือน)
func (h *Handler) Create(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	var in CreateInput
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	b, err := h.svc.Create(c.Request.Context(), identity, in, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, b)
}

// GET /api/v1/bookings — ประวัติการจองของตนเอง
func (h *Handler) ListMine(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	status, err := queryStatus(c)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	stayType, err := queryStayType(c)
	if err != nil {
		httpx.Error(c, err)
		return
	}

	page, pageSize, offset, err := httpx.Pagination(c)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	bookings, total, err := h.svc.ListMine(c.Request.Context(), identity, status, stayType, pageSize, offset)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Page(c, bookings, httpx.Meta{Page: page, PageSize: pageSize, TotalItems: total})
}

// GET /api/v1/bookings/:bookingID — เจ้าของการจอง หรือแอดมินของสาขานั้น
func (h *Handler) Get(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	id, err := httpx.ParseUUID(c.Param("bookingID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	b, err := h.svc.Get(c.Request.Context(), identity, id)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, b)
}

// POST /api/v1/bookings/:bookingID/payment — แจ้งชำระเงิน + อัปโหลดสลิป
// multipart/form-data: slip (ไฟล์), amount, transferred_at, note
func (h *Handler) SubmitPayment(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	bookingID, err := httpx.ParseUUID(c.Param("bookingID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	if err := h.files.ParseMultipart(c); err != nil {
		httpx.Error(c, err)
		return
	}

	amount, err := strconv.ParseFloat(strings.TrimSpace(c.PostForm("amount")), 64)
	if err != nil {
		httpx.Error(c, httpx.ValidationFailed(map[string]string{"amount": "จำนวนเงินต้องเป็นตัวเลข"}))
		return
	}

	transferredAt, err := httpx.ParseFlexibleTime(c.PostForm("transferred_at"))
	if err != nil {
		httpx.Error(c, httpx.ValidationFailed(map[string]string{
			"transferred_at": "รูปแบบวันเวลาที่โอนไม่ถูกต้อง (ต้องเป็น YYYY-MM-DD HH:MM)",
		}))
		return
	}

	note, err := httpx.FormValue(c, "note")
	if err != nil {
		httpx.Error(c, err)
		return
	}

	slipURL, err := h.files.SaveFromRequest(c, "slip", "slips")
	if err != nil {
		httpx.Error(c, err)
		return
	}

	b, err := h.svc.SubmitPayment(c.Request.Context(), identity, bookingID, SubmitPaymentInput{
		Amount:        amount,
		TransferredAt: transferredAt,
		SlipURL:       slipURL,
		Note:          note,
	}, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, b)
}

// POST /api/v1/bookings/:bookingID/cancel
func (h *Handler) Cancel(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	id, err := httpx.ParseUUID(c.Param("bookingID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	b, err := h.svc.Cancel(c.Request.Context(), identity, id, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, b)
}

// ---------------------------------------------------------------- admin

// GET /api/v1/admin/bookings?status=&stay_type=&search=&branch_id=
func (h *Handler) ListForAdmin(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	branchID, err := httpx.QueryUUID(c, "branch_id")
	if err != nil {
		httpx.Error(c, err)
		return
	}
	status, err := queryStatus(c)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	stayType, err := queryStayType(c)
	if err != nil {
		httpx.Error(c, err)
		return
	}

	page, pageSize, offset, err := httpx.Pagination(c)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	bookings, total, err := h.svc.ListForAdmin(c.Request.Context(), identity, branchID, status, stayType,
		httpx.QueryString(c, "search"), pageSize, offset)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Page(c, bookings, httpx.Meta{Page: page, PageSize: pageSize, TotalItems: total})
}

// POST /api/v1/admin/bookings/:bookingID/approve
func (h *Handler) Approve(c *gin.Context) { h.review(c, true) }

// POST /api/v1/admin/bookings/:bookingID/reject
func (h *Handler) Reject(c *gin.Context) { h.review(c, false) }

func (h *Handler) review(c *gin.Context, approve bool) {
	identity := middleware.MustIdentity(c)

	id, err := httpx.ParseUUID(c.Param("bookingID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}

	var in struct {
		Reason string `json:"reason"`
	}
	if c.Request.ContentLength > 0 {
		if err := httpx.DecodeJSON(c, &in); err != nil {
			httpx.Error(c, err)
			return
		}
	}

	b, err := h.svc.Review(c.Request.Context(), identity, id, approve, strings.TrimSpace(in.Reason), middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, b)
}

// PUT /api/v1/admin/bookings/:bookingID/appointment — วันนัดหมายทำสัญญา (รายเดือน)
func (h *Handler) SetAppointment(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	id, err := httpx.ParseUUID(c.Param("bookingID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}

	var in struct {
		AppointmentAt string `json:"appointment_at"` // RFC3339 หรือ "YYYY-MM-DD HH:MM"
		Note          string `json:"note"`
	}
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}

	at, err := httpx.ParseFlexibleTime(in.AppointmentAt)
	if err != nil {
		httpx.Error(c, httpx.ValidationFailed(map[string]string{
			"appointment_at": "รูปแบบวันเวลานัดหมายไม่ถูกต้อง (ต้องเป็น YYYY-MM-DD HH:MM)",
		}))
		return
	}

	b, err := h.svc.SetAppointment(c.Request.Context(), identity, id,
		AppointmentInput{AppointmentAt: at, Note: in.Note}, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, b)
}

// GET /api/v1/admin/members/:memberID/bookings — ประวัติการจองของสมาชิกรายคน
func (h *Handler) ListByMember(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	memberID, err := httpx.ParseUUID(c.Param("memberID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}

	page, pageSize, offset, err := httpx.Pagination(c)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	bookings, total, err := h.svc.ListByMember(c.Request.Context(), identity, memberID, pageSize, offset)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Page(c, bookings, httpx.Meta{Page: page, PageSize: pageSize, TotalItems: total})
}

// ---------------------------------------------------------------- query helpers

func queryStatus(c *gin.Context) (*BookingStatus, error) {
	raw := httpx.QueryString(c, "status")
	if raw == "" || raw == "all" {
		return nil, nil
	}
	st := BookingStatus(raw)
	if !st.Valid() {
		return nil, httpx.BadRequest("status ไม่ถูกต้อง")
	}
	return &st, nil
}

func queryStayType(c *gin.Context) (*types.StayType, error) {
	raw := httpx.QueryString(c, "stay_type")
	if raw == "" {
		return nil, nil
	}
	st := types.StayType(raw)
	if !st.Valid() {
		return nil, httpx.BadRequest("stay_type ต้องเป็น daily หรือ monthly")
	}
	return &st, nil
}
