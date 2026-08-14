package account

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/httpx"
	"backend/internal/middleware"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ---------------------------------------------------------------- auth

// POST /api/v1/auth/register — สมัครสมาชิก
func (h *Handler) Register(c *gin.Context) {
	var in RegisterInput
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	tokens, err := h.svc.Register(c.Request.Context(), in, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, tokens)
}

// POST /api/v1/auth/login — ใช้ร่วมกันทั้ง member / admin / superadmin
func (h *Handler) Login(c *gin.Context) {
	var in LoginInput
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	tokens, err := h.svc.Login(c.Request.Context(), in, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, tokens)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// POST /api/v1/auth/refresh
func (h *Handler) Refresh(c *gin.Context) {
	var in refreshRequest
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	tokens, err := h.svc.Refresh(c.Request.Context(), in.RefreshToken)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, tokens)
}

// POST /api/v1/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	var in refreshRequest
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	if err := h.svc.Logout(c.Request.Context(), in.RefreshToken); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}

// ---------------------------------------------------------------- โปรไฟล์

// GET /api/v1/me
func (h *Handler) Me(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	user, err := h.svc.Me(c.Request.Context(), identity.UserID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, user)
}

// PUT /api/v1/me
func (h *Handler) UpdateMe(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	var in UpdateProfileInput
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	user, err := h.svc.UpdateProfile(c.Request.Context(), identity, in, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, user)
}

// POST /api/v1/me/password
func (h *Handler) ChangePassword(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	var in ChangePasswordInput
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), identity, in, middleware.ClientIP(c)); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, map[string]string{"message": "เปลี่ยนรหัสผ่านเรียบร้อยแล้ว กรุณาเข้าสู่ระบบใหม่"})
}

// ---------------------------------------------------------------- แจ้งเตือน

// GET /api/v1/me/notifications
func (h *Handler) ListNotifications(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	items, unread, err := h.svc.ListNotifications(c.Request.Context(), identity.UserID, 30)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, map[string]any{"items": items, "unread_count": unread})
}

// POST /api/v1/me/notifications/read — อ่านทั้งหมด หรือระบุ id เดียว
func (h *Handler) MarkNotificationsRead(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	var in struct {
		ID *string `json:"id"`
	}
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}

	var id *uuid.UUID
	if in.ID != nil && *in.ID != "" {
		parsed, err := httpx.ParseUUID(*in.ID)
		if err != nil {
			httpx.Error(c, err)
			return
		}
		id = &parsed
	}

	if err := h.svc.MarkNotificationsRead(c.Request.Context(), identity.UserID, id); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}

// ---------------------------------------------------------------- สมาชิก (แอดมิน)

// GET /api/v1/admin/members?search=
func (h *Handler) ListMembers(c *gin.Context) {
	page, pageSize, offset := httpx.Pagination(c)

	members, total, err := h.svc.ListMembers(c.Request.Context(), httpx.QueryString(c, "search"), pageSize, offset)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Page(c, members, httpx.Meta{Page: page, PageSize: pageSize, TotalItems: total})
}

// ---------------------------------------------------------------- ผู้ดูแลระบบ

// GET /api/v1/superadmin/staff
func (h *Handler) ListStaff(c *gin.Context) {
	staff, err := h.svc.ListStaff(c.Request.Context())
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, staff)
}

// POST /api/v1/superadmin/staff
func (h *Handler) CreateAdmin(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	var in CreateAdminInput
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	user, err := h.svc.CreateAdmin(c.Request.Context(), identity, in, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, user)
}

// PUT /api/v1/superadmin/staff/:userID
func (h *Handler) UpdateAdmin(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	id, err := httpx.ParseUUID(c.Param("userID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}

	var in UpdateAdminInput
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	user, err := h.svc.UpdateAdmin(c.Request.Context(), identity, id, in, middleware.ClientIP(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, user)
}

// DELETE /api/v1/superadmin/staff/:userID
func (h *Handler) DeleteAdmin(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	id, err := httpx.ParseUUID(c.Param("userID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	if err := h.svc.DeleteAdmin(c.Request.Context(), identity, id, middleware.ClientIP(c)); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}
