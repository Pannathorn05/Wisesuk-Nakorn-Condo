package room

import (
	"github.com/gin-gonic/gin"

	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/shared/types"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// searchInputFrom อ่านตัวกรองทั้งหมดของหน้าค้นหาห้องพักจาก query string
func searchInputFrom(c *gin.Context) (SearchInput, error) {
	var in SearchInput
	var err error

	if in.BranchID, err = httpx.QueryUUID(c, "branch_id"); err != nil {
		return in, err
	}
	if in.RoomTypeID, err = httpx.QueryUUID(c, "room_type_id"); err != nil {
		return in, err
	}
	if in.StayType, err = queryStayType(c); err != nil {
		return in, err
	}
	if in.CheckIn, err = httpx.QueryDate(c, "check_in"); err != nil {
		return in, err
	}
	if in.CheckOut, err = httpx.QueryDate(c, "check_out"); err != nil {
		return in, err
	}
	if in.MoveInDate, err = httpx.QueryDate(c, "move_in_date"); err != nil {
		return in, err
	}
	if in.MinPrice, err = httpx.QueryFloat(c, "min_price"); err != nil {
		return in, err
	}
	if in.MaxPrice, err = httpx.QueryFloat(c, "max_price"); err != nil {
		return in, err
	}
	return in, nil
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

// GET /api/v1/room-types?branch_id=
func (h *Handler) ListTypes(c *gin.Context) {
	branchID, err := httpx.QueryUUID(c, "branch_id")
	if err != nil {
		httpx.Error(c, err)
		return
	}
	list, err := h.svc.ListTypes(c.Request.Context(), branchID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, list)
}

// GET /api/v1/rooms/search — หน้า "ค้นหาห้องพัก"
func (h *Handler) Search(c *gin.Context) {
	in, err := searchInputFrom(c)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	page, pageSize, offset := httpx.Pagination(c)
	in.Limit, in.Offset = pageSize, offset

	rooms, total, err := h.svc.Search(c.Request.Context(), in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Page(c, rooms, httpx.Meta{Page: page, PageSize: pageSize, TotalItems: total})
}

// GET /api/v1/rooms/:roomID
func (h *Handler) Get(c *gin.Context) {
	id, err := httpx.ParseUUID(c.Param("roomID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	rm, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, rm)
}

// ---------------------------------------------------------------- admin

// GET /api/v1/admin/rooms — ห้องทุกสถานะในสาขาที่ดูแล
func (h *Handler) ListForAdmin(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	in, err := searchInputFrom(c)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	page, pageSize, offset := httpx.Pagination(c)
	in.Limit, in.Offset = pageSize, offset

	rooms, total, err := h.svc.ListForAdmin(c.Request.Context(), identity, in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Page(c, rooms, httpx.Meta{Page: page, PageSize: pageSize, TotalItems: total})
}

// POST /api/v1/admin/rooms
func (h *Handler) Create(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	var in SaveInput
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	rm, err := h.svc.Create(c.Request.Context(), identity, in, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, rm)
}

// PUT /api/v1/admin/rooms/:roomID
func (h *Handler) Update(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	id, err := httpx.ParseUUID(c.Param("roomID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}

	var in SaveInput
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	rm, err := h.svc.Update(c.Request.Context(), identity, id, in, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, rm)
}

// PATCH /api/v1/admin/rooms/:roomID/status — อัปเดตสถานะห้องแบบ real-time
func (h *Handler) UpdateStatus(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	id, err := httpx.ParseUUID(c.Param("roomID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}

	var in struct {
		Status string `json:"status"`
	}
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	rm, err := h.svc.UpdateStatus(c.Request.Context(), identity, id, in.Status, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, rm)
}

// DELETE /api/v1/admin/rooms/:roomID
func (h *Handler) Delete(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	id, err := httpx.ParseUUID(c.Param("roomID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), identity, id, middleware.ClientIP(c)); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}
