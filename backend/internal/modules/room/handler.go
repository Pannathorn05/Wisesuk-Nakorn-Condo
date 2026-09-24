package room

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/shared/types"
	"backend/internal/storage"
)

type Handler struct {
	svc   *Service
	files *storage.DBStore
}

func NewHandler(svc *Service, files *storage.DBStore) *Handler {
	return &Handler{svc: svc, files: files}
}

// POST /api/v1/admin/rooms/:roomID/image — multipart: image
//
// อัปโหลดรูปห้อง ตัวไฟล์ถูกเก็บลงตาราง assets แล้วบันทึกเป็น URL /files/:id
func (h *Handler) UploadImage(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	roomID, err := httpx.ParseID[types.Room](c.Param("roomID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}

	url, err := h.files.SaveFromRequest(c, "image")
	if err != nil {
		httpx.Error(c, err)
		return
	}

	rm, err := h.svc.SetImage(c.Request.Context(), identity, roomID, url, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, rm)
}

// POST /api/v1/admin/rooms/:roomID/images/upload — multipart: image, sort_order
//
// อัปโหลดรูปเข้าแกลเลอรีของห้อง ตัวไฟล์ถูกเก็บลงตาราง assets แล้วบันทึกเป็น URL /files/:id
// คนละช่องกับรูปปก ซึ่งใช้ POST /admin/rooms/:roomID/image (เอกพจน์)
func (h *Handler) UploadGalleryImage(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	roomID, err := httpx.ParseID[types.Room](c.Param("roomID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}

	// ต้องบันทึกไฟล์ก่อน เพราะเป็นจุดที่พาร์ส multipart ให้อ่านฟิลด์ข้อความต่อได้
	url, err := h.files.SaveFromRequest(c, "image")
	if err != nil {
		httpx.Error(c, err)
		return
	}

	sortOrder, err := formInt(c, "sort_order")
	if err != nil {
		httpx.Error(c, err)
		return
	}

	img, err := h.svc.AddImage(c.Request.Context(), identity, roomID, url, sortOrder, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, img)
}

// POST /api/v1/admin/rooms/:roomID/images — JSON: image_url, sort_order
//
// ใช้ผูกรูปที่โฮสต์ไว้ที่อื่นอยู่แล้ว ถ้าจะอัปโหลดไฟล์ให้ใช้ /images/upload
func (h *Handler) AddGalleryImage(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	roomID, err := httpx.ParseID[types.Room](c.Param("roomID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}

	var in struct {
		ImageURL  string `json:"image_url"`
		SortOrder int    `json:"sort_order"`
	}
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}

	img, err := h.svc.AddImage(c.Request.Context(), identity, roomID, in.ImageURL, in.SortOrder, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, img)
}

// DELETE /api/v1/admin/rooms/:roomID/images/:imageID
func (h *Handler) DeleteGalleryImage(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	roomID, err := httpx.ParseID[types.Room](c.Param("roomID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	imageID, err := httpx.ParseID[types.RoomImage](c.Param("imageID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}

	if err := h.svc.DeleteImage(c.Request.Context(), identity, roomID, imageID, middleware.ClientIP(c)); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}

// formInt อ่านฟิลด์ตัวเลขจาก multipart form — ไม่ส่งมาถือว่าเป็น 0
func formInt(c *gin.Context, field string) (int, error) {
	raw := c.PostForm(field)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, httpx.BadRequest(field + " ต้องเป็นจำนวนเต็ม")
	}
	return n, nil
}

// searchInputFrom อ่านตัวกรองทั้งหมดของหน้าค้นหาห้องพักจาก query string
func searchInputFrom(c *gin.Context) (SearchInput, error) {
	var in SearchInput
	var err error

	if in.BranchID, err = httpx.QueryID[types.Branch](c, "branch_id"); err != nil {
		return in, err
	}
	if in.RoomTypeID, err = httpx.QueryID[types.RoomType](c, "room_type_id"); err != nil {
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
	branchID, err := httpx.QueryID[types.Branch](c, "branch_id")
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
	page, pageSize, offset, err := httpx.Pagination(c)
	if err != nil {
		httpx.Error(c, err)
		return
	}
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
	id, err := httpx.ParseID[types.Room](c.Param("roomID"))
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
	page, pageSize, offset, err := httpx.Pagination(c)
	if err != nil {
		httpx.Error(c, err)
		return
	}
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

	id, err := httpx.ParseID[types.Room](c.Param("roomID"))
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

	id, err := httpx.ParseID[types.Room](c.Param("roomID"))
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

	id, err := httpx.ParseID[types.Room](c.Param("roomID"))
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
