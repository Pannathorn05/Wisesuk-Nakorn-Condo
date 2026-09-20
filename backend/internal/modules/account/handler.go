package account

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/oauth"
	"backend/internal/shared/types"
)

// OAuthProvider คือสิ่งเดียวที่ module นี้ต้องรู้เกี่ยวกับ package oauth
// ประกาศไว้ฝั่งผู้ใช้งานตามธรรมเนียมเดียวกับ BranchChecker
type OAuthProvider interface {
	Enabled(provider string) bool
	Names() []string
	AuthCodeURL(provider, state, challenge string) (string, error)
	Exchange(ctx context.Context, provider, code, verifier string) (oauth.Profile, error)
}

// WebConfig คือค่าที่ handler ต้องรู้เกี่ยวกับเบราว์เซอร์ที่เรียกเข้ามา
type WebConfig struct {
	// FrontendCallbackURL คือหน้าเว็บฝั่ง frontend ที่ callback จะ redirect ผู้ใช้กลับไป
	// พร้อม query string code=... หรือ error=...
	FrontendCallbackURL string
	// SecureCookies ปิดเฉพาะตอน development ที่ยังเป็น http ธรรมดา
	SecureCookies bool
}

type Handler struct {
	svc   *Service
	oauth OAuthProvider
	web   WebConfig
}

func NewHandler(svc *Service, provider OAuthProvider, web WebConfig) *Handler {
	return &Handler{svc: svc, oauth: provider, web: web}
}

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

// ---------------------------------------------------------------- oauth

// ชื่อ cookie ที่ถือ state และ PKCE verifier ระหว่างรอผู้ใช้กดยินยอมที่ฝั่ง provider
//
// ใช้ cookie ไม่ใช่ตารางในฐานข้อมูล เพราะค่าสองตัวนี้ต้องผูกกับ "เบราว์เซอร์ที่เริ่มเรื่อง"
// เท่านั้น ถ้าเก็บฝั่ง server แล้วอ้างด้วย state ที่มาจาก URL ก็เท่ากับไม่ได้กัน CSRF อะไรเลย
const (
	cookieOAuthState    = "wisetsuk_oauth_state"
	cookieOAuthVerifier = "wisetsuk_oauth_verifier"
	cookieOAuthPath     = "/api/v1/auth/oauth"
	cookieOAuthMaxAge   = 600 // 10 นาที — เผื่อเวลาผู้ใช้เลือกบัญชีและใส่รหัสผ่านที่ฝั่ง provider
)

// GET /api/v1/auth/oauth — บอก frontend ว่าควรขึ้นปุ่มของเจ้าไหนบ้าง
func (h *Handler) OAuthProviders(c *gin.Context) {
	httpx.OK(c, gin.H{"providers": h.oauth.Names()})
}

// GET /api/v1/auth/oauth/:provider — พาผู้ใช้ไปหน้ายินยอมของ provider
//
// ตอบเป็น 302 ไม่ใช่ JSON เพราะปลายทางคือเบราว์เซอร์ที่กดปุ่มมาตรง ๆ ไม่ใช่ fetch
func (h *Handler) OAuthStart(c *gin.Context) {
	provider := strings.ToLower(c.Param("provider"))
	if !h.oauth.Enabled(provider) {
		httpx.Error(c, httpx.ErrNotFound)
		return
	}

	hs, err := oauth.NewHandshake()
	if err != nil {
		httpx.Error(c, httpx.ErrInternal.Wrap(err))
		return
	}

	target, err := h.oauth.AuthCodeURL(provider, hs.State, hs.Challenge)
	if err != nil {
		httpx.Error(c, httpx.ErrInternal.Wrap(err))
		return
	}

	h.setHandshakeCookies(c, hs)
	c.Redirect(http.StatusFound, target)
}

// GET /api/v1/auth/oauth/:provider/callback — ปลายทางที่ provider ส่งผู้ใช้กลับมา
//
// ทุกทางออกของ handler นี้จบด้วยการ redirect กลับหา frontend เสมอ ไม่ว่าจะสำเร็จหรือพัง
// เพราะสิ่งที่อยู่ปลายทางคือหน้าต่างเบราว์เซอร์ของผู้ใช้ ไม่ใช่โค้ดที่อ่าน JSON ได้
func (h *Handler) OAuthCallback(c *gin.Context) {
	provider := strings.ToLower(c.Param("provider"))
	if !h.oauth.Enabled(provider) {
		httpx.Error(c, httpx.ErrNotFound)
		return
	}

	state, verifier := h.takeHandshakeCookies(c)

	// ผู้ใช้กดยกเลิกที่หน้ายินยอม provider จะส่ง error กลับมาแทน code
	if reason := c.Query("error"); reason != "" {
		h.redirectWithError(c, "คุณยกเลิกการเข้าสู่ระบบด้วยบัญชีภายนอก")
		return
	}

	// state ต้องมีอยู่จริงและตรงกับที่เราเพิ่งออกให้เบราว์เซอร์เครื่องนี้
	// ใช้ != ธรรมดาได้เพราะค่าที่เทียบไม่ใช่ความลับ ผู้โจมตีที่เดา state ถูกก็ต้องรู้ code อยู่ดี
	if state == "" || verifier == "" || c.Query("state") != state {
		h.redirectWithError(c, "คำขอเข้าสู่ระบบไม่ถูกต้องหรือหมดอายุ กรุณาลองใหม่")
		return
	}

	code := c.Query("code")
	if code == "" {
		h.redirectWithError(c, "ไม่ได้รับข้อมูลยืนยันจากผู้ให้บริการ กรุณาลองใหม่")
		return
	}

	profile, err := h.oauth.Exchange(c.Request.Context(), provider, code, verifier)
	if err != nil {
		h.redirectWithError(c, oauthFailureMessage(err))
		return
	}

	exchangeCode, err := h.svc.LoginWithProvider(c.Request.Context(), profile, middleware.ClientIP(c))
	if err != nil {
		var apiErr *httpx.APIError
		if errors.As(err, &apiErr) {
			h.redirectWithError(c, apiErr.Message)
			return
		}
		h.redirectWithError(c, "เข้าสู่ระบบไม่สำเร็จ กรุณาลองใหม่")
		return
	}

	c.Redirect(http.StatusFound, h.frontendURL(url.Values{"code": {exchangeCode}}))
}

type oauthExchangeRequest struct {
	Code string `json:"code"`
}

// POST /api/v1/auth/oauth/exchange — frontend เอาโค้ดจาก URL มาแลกเป็น token จริง
func (h *Handler) OAuthExchange(c *gin.Context) {
	var in oauthExchangeRequest
	if err := httpx.DecodeJSON(c, &in); err != nil {
		httpx.Error(c, err)
		return
	}
	tokens, err := h.svc.ExchangeOAuthCode(c.Request.Context(), in.Code)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, tokens)
}

// setHandshakeCookies เก็บ state และ verifier ไว้กับเบราว์เซอร์
//
// SameSite=Lax คือค่าที่ใช้ได้จริงค่าเดียว: Strict จะไม่ส่ง cookie กลับมาตอน provider
// redirect ผู้ใช้ข้ามเว็บมาที่ callback ส่วน None เปิดกว้างเกินจำเป็น
func (h *Handler) setHandshakeCookies(c *gin.Context, hs oauth.Handshake) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(cookieOAuthState, hs.State, cookieOAuthMaxAge, cookieOAuthPath, "", h.web.SecureCookies, true)
	c.SetCookie(cookieOAuthVerifier, hs.Verifier, cookieOAuthMaxAge, cookieOAuthPath, "", h.web.SecureCookies, true)
}

// takeHandshakeCookies อ่านค่าแล้วลบทิ้งทันที คำขอเดิมจึงเล่นซ้ำไม่ได้
func (h *Handler) takeHandshakeCookies(c *gin.Context) (state, verifier string) {
	state, _ = c.Cookie(cookieOAuthState)
	verifier, _ = c.Cookie(cookieOAuthVerifier)

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(cookieOAuthState, "", -1, cookieOAuthPath, "", h.web.SecureCookies, true)
	c.SetCookie(cookieOAuthVerifier, "", -1, cookieOAuthPath, "", h.web.SecureCookies, true)
	return state, verifier
}

func (h *Handler) redirectWithError(c *gin.Context, message string) {
	c.Redirect(http.StatusFound, h.frontendURL(url.Values{"error": {message}}))
}

// frontendURL ต่อ query เข้ากับ URL ปลายทางฝั่ง frontend โดยรักษา query เดิมที่ตั้งค่าไว้
func (h *Handler) frontendURL(extra url.Values) string {
	target, err := url.Parse(h.web.FrontendCallbackURL)
	if err != nil {
		return h.web.FrontendCallbackURL
	}
	q := target.Query()
	for key, values := range extra {
		for _, v := range values {
			q.Set(key, v)
		}
	}
	target.RawQuery = q.Encode()
	return target.String()
}

// oauthFailureMessage แปล error จาก package oauth เป็นข้อความที่ผู้ใช้อ่านแล้วรู้ว่าต้องทำอะไรต่อ
// error อื่นรวบเป็นข้อความกลาง ไม่ปล่อยรายละเอียดของ provider ออกไปให้ผู้ใช้เห็น
func oauthFailureMessage(err error) string {
	switch {
	case errors.Is(err, oauth.ErrNoEmail):
		return "ผู้ให้บริการไม่ได้ส่งอีเมลกลับมา กรุณาอนุญาตให้เข้าถึงอีเมล หรือสมัครสมาชิกด้วยอีเมลแทน"
	case errors.Is(err, oauth.ErrEmailNotVerified):
		return "อีเมลของบัญชีนี้ยังไม่ได้รับการยืนยัน กรุณายืนยันอีเมลกับผู้ให้บริการก่อน"
	default:
		return "เชื่อมต่อกับผู้ให้บริการไม่สำเร็จ กรุณาลองใหม่"
	}
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

	var id *types.NotificationID
	if in.ID != nil && *in.ID != "" {
		parsed, err := httpx.ParseID[types.Notification](*in.ID)
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
	page, pageSize, offset, err := httpx.Pagination(c)
	if err != nil {
		httpx.Error(c, err)
		return
	}

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

	id, err := httpx.ParseID[types.User](c.Param("userID"))
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

	id, err := httpx.ParseID[types.User](c.Param("userID"))
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
