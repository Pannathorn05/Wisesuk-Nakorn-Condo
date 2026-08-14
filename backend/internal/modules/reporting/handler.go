package reporting

import (
	"github.com/gin-gonic/gin"

	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/shared/types"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// GET /api/v1/admin/dashboard
func (h *Handler) Dashboard(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	data, err := h.svc.Dashboard(c.Request.Context(), identity)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, data)
}

// GET /api/v1/admin/activity-logs?actor_role=&actor_id=&action=&branch_id=
func (h *Handler) ListActivityLogs(c *gin.Context) {
	identity := middleware.MustIdentity(c)

	f := ActivityFilter{Action: httpx.QueryString(c, "action")}

	var err error
	if f.ActorID, err = httpx.QueryUUID(c, "actor_id"); err != nil {
		httpx.Error(c, err)
		return
	}
	if f.BranchID, err = httpx.QueryUUID(c, "branch_id"); err != nil {
		httpx.Error(c, err)
		return
	}
	if raw := httpx.QueryString(c, "actor_role"); raw != "" {
		role := types.Role(raw)
		if !role.Valid() {
			httpx.Error(c, httpx.BadRequest("actor_role ต้องเป็น member, admin หรือ superadmin"))
			return
		}
		f.ActorRole = &role
	}

	page, pageSize, offset := httpx.Pagination(c)
	f.Limit, f.Offset = pageSize, offset

	logs, total, err := h.svc.ActivityLogs(c.Request.Context(), identity, f)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Page(c, logs, httpx.Meta{Page: page, PageSize: pageSize, TotalItems: total})
}
