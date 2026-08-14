package branch

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/httpx"
	"backend/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/v1/branches — สาขาในเครือทั้งหมด
func (h *Handler) List(c *gin.Context) {
	branches, err := h.svc.List(c.Request.Context(), false)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, branches)
}

// GET /api/v1/branches/:branchID — พร้อมรูป/สิ่งอำนวยความสะดวก/สถานที่ใกล้เคียง
func (h *Handler) Get(c *gin.Context) {
	id, err := httpx.ParseUUID(c.Param("branchID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	b, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, b)
}

// GET /api/v1/amenities
func (h *Handler) ListAmenities(c *gin.Context) {
	items, err := h.svc.ListAmenities(c.Request.Context())
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, items)
}

// GET /api/v1/superadmin/branches — รวมสาขาที่ปิดใช้งาน
func (h *Handler) ListAll(c *gin.Context) {
	branches, err := h.svc.List(c.Request.Context(), true)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, branches)
}

// ---------------------------------------------------------------- admin

// PUT /api/v1/admin/branch?branch_id=
func (h *Handler) Update(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	branchID, err := httpx.QueryUUID(c, "branch_id")
	if err != nil {
		httpx.Error(c, err)
		return
	}

	var in UpdateInput
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	b, err := h.svc.Update(c.Request.Context(), identity, branchID, in, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, b)
}

// PUT /api/v1/admin/branch/amenities?branch_id=
func (h *Handler) SetAmenities(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	branchID, err := httpx.QueryUUID(c, "branch_id")
	if err != nil {
		httpx.Error(c, err)
		return
	}

	var in struct {
		AmenityIDs []uuid.UUID `json:"amenity_ids"`
	}
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	items, err := h.svc.SetAmenities(c.Request.Context(), identity, branchID, in.AmenityIDs, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, items)
}

// PUT /api/v1/admin/branch/nearby?branch_id= — แทนที่ทั้งชุด
func (h *Handler) ReplaceNearby(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	branchID, err := httpx.QueryUUID(c, "branch_id")
	if err != nil {
		httpx.Error(c, err)
		return
	}

	var in struct {
		Items []NearbyInput `json:"items"`
	}
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	places, err := h.svc.ReplaceNearby(c.Request.Context(), identity, branchID, in.Items, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, places)
}

// POST /api/v1/admin/branch/images?branch_id= — JSON: image_url, caption, sort_order
//
// รูปถูกโฮสต์ไว้ที่บริการภายนอก ฝั่งเว็บอัปโหลดเองแล้วส่งมาแค่ URL
func (h *Handler) AddImage(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	branchID, err := httpx.QueryUUID(c, "branch_id")
	if err != nil {
		httpx.Error(c, err)
		return
	}

	var in struct {
		ImageURL  string `json:"image_url"`
		Caption   string `json:"caption"`
		SortOrder int    `json:"sort_order"`
	}
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}

	img, err := h.svc.AddImage(c.Request.Context(), identity, branchID, in.ImageURL,
		in.Caption, in.SortOrder, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, img)
}

// DELETE /api/v1/admin/branch/images/:imageID?branch_id=
func (h *Handler) DeleteImage(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	branchID, err := httpx.QueryUUID(c, "branch_id")
	if err != nil {
		httpx.Error(c, err)
		return
	}
	imageID, err := httpx.ParseUUID(c.Param("imageID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	if err := h.svc.DeleteImage(c.Request.Context(), identity, branchID, imageID, middleware.ClientIP(c)); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}
