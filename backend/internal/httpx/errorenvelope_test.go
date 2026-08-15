// เทสของ AC-21: รหัสสถานะและรูปแบบ error สอดคล้องกันทั้งระบบ
//
// checklist เทียบกับ AC-21 ใน docs/SPEC.md ทีละบรรทัด
//
//	บรรทัดที่ 1  400 bad_request: พาร์ส body ไม่ได้            -> TestAC21_BadRequest_400/body_พาร์สไม่ได้
//	             400 bad_request: ชนิดข้อมูลผิด                -> TestAC21_BadRequest_400/ชนิดข้อมูลผิด
//	             400 bad_request: query param ผิดรูป (UUID)    -> TestAC21_BadRequest_400/branch_id_ไม่ใช่_UUID
//	             400 bad_request: query param ผิดรูป (page)    -> TestAC21_BadRequest_400/page_ไม่ใช่ตัวเลข
//	             400 bad_request: คำขอที่รูปแบบไม่ถูกต้อง       -> TestAC21_BadRequest_400/content_type_ไม่ใช่_JSON
//	บรรทัดที่ 2  422 validation_failed: ผิดกติกาธุรกิจ          -> TestAC21_ValidationFailed_422/รหัสผ่านสั้นเกินไป
//	             422 validation_failed: ส่งฟิลด์ต้องห้าม        -> TestAC21_ValidationFailed_422/ส่งฟิลด์_role_แถมมา
//	บรรทัดที่ 3  fields มีเฉพาะตอน validation_failed           -> TestAC21_FieldsOnlyOnValidationFailed
//	บรรทัดที่ 4  ไม่มี token / token ใช้ไม่ได้ -> 401           -> TestAC21_UnauthorizedVsForbidden/ไม่แนบ_token, /token_ปลอม
//	             มี token แต่บทบาทไม่พอ -> 403                 -> TestAC21_UnauthorizedVsForbidden/member_เรียก_admin
//	บรรทัดที่ 5  404/405/500 ใช้ format เดียวกัน                -> TestAC21_RouterErrorsUseSameEnvelope, TestAC21_PanicReturns500JSON
//	             ไม่มี response ที่หลุดออกนอกรูปแบบ             -> TestAC21_EveryErrorUsesEnvelope
//
// หมายเหตุ: ตัวอย่าง "check_out <= check_in" และ "ไฟล์อัปโหลดไม่ถูกต้อง" ใน AC-21 เป็นตัวอย่าง
// ของกฎเดียวกันที่ทดสอบไว้แล้วด้วยเคสอื่นในไฟล์นี้ ส่วนพฤติกรรมรายฟีเจอร์อยู่ใน AC-6 และ AC-9
package httpx_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"backend/internal/config"
	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/routes"
	"backend/internal/server"
	"backend/internal/storage"
	"backend/internal/testsupport"
)

// specErrorCodes คือรหัสทั้งหมดที่ SPEC อนุญาต ห้ามมีรหัสอื่นหลุดออกไปหา client
var specErrorCodes = map[string]bool{
	"unauthorized": true, "forbidden": true, "not_found": true, "conflict": true,
	"bad_request": true, "validation_failed": true, "method_not_allowed": true,
	"internal_error": true, "invalid_credentials": true, "account_disabled": true,
	"room_unavailable": true, "invalid_state": true,
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// ---------------------------------------------------------------- ตัวช่วย

type app struct {
	handler http.Handler
	fixture testsupport.Fixture
	pool    *pgxpool.Pool
}

// newApp ประกอบ router ตัวจริงของระบบ ไม่ใช่ engine จำลอง
// เทสจึงตรวจสิ่งที่ผู้ใช้เจอจริง รวมถึง middleware ทุกชั้นที่ routes.New ต่อไว้
func newApp(t *testing.T) app {
	t.Helper()

	pool := testsupport.NewDatabase(t)
	fixture := testsupport.Seed(t, pool)
	dir := testsupport.UploadDir(t)

	files, err := storage.NewLocalStore(dir, "http://localhost:8080", 5<<20)
	if err != nil {
		t.Fatalf("สร้าง storage ไม่สำเร็จ: %v", err)
	}

	cfg := &config.Config{
		Env:             "test",
		Port:            "8080",
		JWTSecret:       testsupport.JWTSecret,
		AccessTokenTTL:  testsupport.AccessTTL,
		RefreshTokenTTL: testsupport.RefreshTTL,
		UploadDir:       dir,
		PublicBaseURL:   "http://localhost:8080",
		MaxUploadBytes:  5 << 20,
		AllowedOrigins:  []string{"http://localhost:3000"},
		BcryptCost:      bcrypt.MinCost,
	}

	return app{handler: routes.New(server.New(cfg, pool, files)), fixture: fixture, pool: pool}
}

type request struct {
	method string
	path   string
	body   string
	token  string
}

func (a app) do(t *testing.T, r request) *httptest.ResponseRecorder {
	t.Helper()

	var body *bytes.Reader
	if r.body == "" {
		body = bytes.NewReader(nil)
	} else {
		body = bytes.NewReader([]byte(r.body))
	}

	req := httptest.NewRequest(r.method, r.path, body)
	if r.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if r.token != "" {
		req.Header.Set("Authorization", "Bearer "+r.token)
	}

	rec := httptest.NewRecorder()
	a.handler.ServeHTTP(rec, req)
	return rec
}

// errorBody แกะ response แล้วยืนยันว่าอยู่ในรูปแบบของ SPEC เป๊ะ ๆ
// คือมี key ชั้นบนสุดชื่อ error เพียงตัวเดียว และข้างในมีได้แค่ code, message, fields
func errorBody(t *testing.T, rec *httptest.ResponseRecorder) (code, message string, fields map[string]any, hasFields bool) {
	t.Helper()

	var top map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &top); err != nil {
		t.Fatalf("response ไม่ใช่ JSON: %v (body=%s)", err, rec.Body.String())
	}
	if len(top) != 1 {
		t.Fatalf("response ต้องมี key ชั้นบนสุดตัวเดียวคือ error แต่มี %d ตัว: %s", len(top), rec.Body.String())
	}
	raw, ok := top["error"]
	if !ok {
		t.Fatalf("response ไม่มี key ชื่อ error: %s", rec.Body.String())
	}

	var inner map[string]any
	if err := json.Unmarshal(raw, &inner); err != nil {
		t.Fatalf("error ไม่ใช่ object: %v", err)
	}
	for k := range inner {
		if k != "code" && k != "message" && k != "fields" {
			t.Fatalf("error มี key ที่ SPEC ไม่ได้กำหนด: %q (body=%s)", k, rec.Body.String())
		}
	}

	code, _ = inner["code"].(string)
	message, _ = inner["message"].(string)
	if !specErrorCodes[code] {
		t.Fatalf("code %q ไม่อยู่ในรายการ 12 code ของ SPEC", code)
	}
	if strings.TrimSpace(message) == "" {
		t.Fatalf("message ต้องไม่ว่าง (body=%s)", rec.Body.String())
	}

	f, hasFields := inner["fields"]
	if hasFields {
		fields, _ = f.(map[string]any)
	}
	return code, message, fields, hasFields
}

func assertStatusCode(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode string) (map[string]any, bool) {
	t.Helper()
	code, _, fields, hasFields := errorBody(t, rec)
	if rec.Code != wantStatus {
		t.Errorf("status = %d ต้องการ %d (body=%s)", rec.Code, wantStatus, rec.Body.String())
	}
	if code != wantCode {
		t.Errorf("code = %q ต้องการ %q", code, wantCode)
	}
	return fields, hasFields
}

const validRegisterBody = `{"email":"newuser@wisetsuk.test","password":"Passw0rd123","first_name":"ทดสอบ","last_name":"ระบบ","phone":"0800000001"}`

// ---------------------------------------------------------------- AC-21 บรรทัดที่ 1

func TestAC21_BadRequest_400(t *testing.T) {
	t.Parallel()
	a := newApp(t)

	cases := []struct {
		name string
		req  request
	}{
		{"body พาร์สไม่ได้", request{http.MethodPost, "/api/v1/auth/register", `{"email":`, ""}},
		{"ชนิดข้อมูลผิด", request{http.MethodPost, "/api/v1/auth/register",
			`{"email":123,"password":"Passw0rd123","first_name":"ก","last_name":"ข","phone":"0800000001"}`, ""}},
		{"branch_id ไม่ใช่ UUID", request{http.MethodGet, "/api/v1/room-types?branch_id=ไม่ใช่-uuid", "", ""}},
		{"page ไม่ใช่ตัวเลข", request{http.MethodGet, "/api/v1/rooms/search?page=abc", "", ""}},
		{"content type ไม่ใช่ JSON", request{http.MethodPost, "/api/v1/auth/register", "ข้อความธรรมดา", ""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := a.do(t, tc.req)
			assertStatusCode(t, rec, http.StatusBadRequest, "bad_request")
		})
	}
}

// ---------------------------------------------------------------- AC-21 บรรทัดที่ 2

func TestAC21_ValidationFailed_422(t *testing.T) {
	t.Parallel()
	a := newApp(t)

	cases := []struct {
		name      string
		body      string
		wantField string
	}{
		{
			"รหัสผ่านสั้นเกินไป",
			`{"email":"short@wisetsuk.test","password":"a1","first_name":"ก","last_name":"ข","phone":"0800000001"}`,
			"password",
		},
		{
			"ส่งฟิลด์ role แถมมา",
			`{"email":"role@wisetsuk.test","password":"Passw0rd123","first_name":"ก","last_name":"ข","phone":"0800000001","role":"admin"}`,
			"role",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := a.do(t, request{http.MethodPost, "/api/v1/auth/register", tc.body, ""})
			fields, hasFields := assertStatusCode(t, rec, http.StatusUnprocessableEntity, "validation_failed")
			if !hasFields || len(fields) == 0 {
				t.Fatalf("422 ต้องมี fields ที่ไม่ว่าง (body=%s)", rec.Body.String())
			}
			if _, ok := fields[tc.wantField]; !ok {
				t.Errorf("fields ต้องมี %q แต่ได้ %v", tc.wantField, fields)
			}
		})
	}
}

// ---------------------------------------------------------------- AC-21 บรรทัดที่ 3

func TestAC21_FieldsOnlyOnValidationFailed(t *testing.T) {
	t.Parallel()
	a := newApp(t)
	memberToken := testsupport.AccessToken(t, a.pool, a.fixture.MemberID)

	cases := []struct {
		name string
		req  request
	}{
		{"400", request{http.MethodPost, "/api/v1/auth/register", `{"email":`, ""}},
		{"401", request{http.MethodGet, "/api/v1/me", "", ""}},
		{"403", request{http.MethodGet, "/api/v1/admin/bookings", "", memberToken}},
		{"404", request{http.MethodGet, "/api/v1/ไม่มีเส้นทางนี้", "", ""}},
		{"405", request{http.MethodDelete, "/api/v1/auth/login", "", ""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := a.do(t, tc.req)
			if _, _, _, hasFields := errorBody(t, rec); hasFields {
				t.Errorf("error ที่ไม่ใช่ validation_failed ต้องไม่มี key fields (body=%s)", rec.Body.String())
			}
		})
	}
}

// ---------------------------------------------------------------- AC-21 บรรทัดที่ 4

func TestAC21_UnauthorizedVsForbidden(t *testing.T) {
	t.Parallel()
	a := newApp(t)

	t.Run("ไม่แนบ token", func(t *testing.T) {
		rec := a.do(t, request{http.MethodGet, "/api/v1/me", "", ""})
		assertStatusCode(t, rec, http.StatusUnauthorized, "unauthorized")
	})

	t.Run("token ปลอม", func(t *testing.T) {
		rec := a.do(t, request{http.MethodGet, "/api/v1/me", "", "ไม่ใช่token"})
		assertStatusCode(t, rec, http.StatusUnauthorized, "unauthorized")
	})

	t.Run("member เรียก admin", func(t *testing.T) {
		token := testsupport.AccessToken(t, a.pool, a.fixture.MemberID)
		rec := a.do(t, request{http.MethodGet, "/api/v1/admin/bookings", "", token})
		assertStatusCode(t, rec, http.StatusForbidden, "forbidden")
	})
}

// ---------------------------------------------------------------- AC-21 บรรทัดที่ 5

func TestAC21_RouterErrorsUseSameEnvelope(t *testing.T) {
	t.Parallel()
	a := newApp(t)

	t.Run("404 route ไม่มีจริง", func(t *testing.T) {
		rec := a.do(t, request{http.MethodGet, "/api/v1/ไม่มีเส้นทางนี้", "", ""})
		assertStatusCode(t, rec, http.StatusNotFound, "not_found")
	})

	t.Run("405 method ไม่รองรับ", func(t *testing.T) {
		rec := a.do(t, request{http.MethodDelete, "/api/v1/auth/login", "", ""})
		assertStatusCode(t, rec, http.StatusMethodNotAllowed, "method_not_allowed")
	})
}

func TestAC21_PanicReturns500JSON(t *testing.T) {
	t.Parallel()

	// ใช้ middleware.Recoverer ตัวจริงที่ routes.New ต่อไว้ ไม่ใช่ของจำลอง
	eng := gin.New()
	eng.Use(middleware.RequestID(), middleware.Recoverer())
	eng.GET("/panic", func(*gin.Context) {
		panic("SELECT password_hash FROM users WHERE email = 'super@wisetsuk.test'")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	eng.ServeHTTP(rec, req)

	assertStatusCode(t, rec, http.StatusInternalServerError, "internal_error")

	body := rec.Body.String()
	for _, leak := range []string{"SELECT", "password_hash", "users", "goroutine", ".go:"} {
		if strings.Contains(body, leak) {
			t.Errorf("body 500 ต้องไม่มี %q หลุดออกไป: %s", leak, body)
		}
	}
}

// TestAC21_EveryErrorUsesEnvelope ยืนยันว่าไม่มี response error ตัวใดหลุดออกนอกรูปแบบ
// errorBody จะ fail ทันทีถ้าเจอ key แปลกปลอม รหัสนอกรายการ หรือ message ว่าง
func TestAC21_EveryErrorUsesEnvelope(t *testing.T) {
	t.Parallel()
	a := newApp(t)
	memberToken := testsupport.AccessToken(t, a.pool, a.fixture.MemberID)

	reqs := []request{
		{http.MethodPost, "/api/v1/auth/register", `{"email":`, ""},
		{http.MethodPost, "/api/v1/auth/register", `{"email":"a@b.co","password":"a1","first_name":"ก","last_name":"ข","phone":"08"}`, ""},
		{http.MethodPost, "/api/v1/auth/login", `{"email":"ไม่มี@wisetsuk.test","password":"Passw0rd123"}`, ""},
		{http.MethodGet, "/api/v1/me", "", ""},
		{http.MethodGet, "/api/v1/admin/bookings", "", memberToken},
		{http.MethodGet, "/api/v1/ไม่มีเส้นทางนี้", "", ""},
		{http.MethodDelete, "/api/v1/auth/login", "", ""},
		{http.MethodGet, "/api/v1/room-types?branch_id=ไม่ใช่-uuid", "", ""},
	}

	for _, r := range reqs {
		rec := a.do(t, r)
		if rec.Code < 400 {
			t.Fatalf("%s %s ควรเป็น error แต่ได้ %d", r.method, r.path, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("%s %s: Content-Type = %q ต้องเป็น application/json", r.method, r.path, ct)
		}
		errorBody(t, rec)
	}
}

// TestAC21_HelpersMatchSpec ตรวจที่ระดับ package ว่า helper ที่ทุก module ใช้
// ผูกรหัสกับสถานะตรงตาม SPEC ไม่ใช่แค่บังเอิญถูกที่ปลายทาง
func TestAC21_HelpersMatchSpec(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		err        *httpx.APIError
		wantStatus int
		wantCode   string
	}{
		{"BadRequest", httpx.BadRequest("ข้อความ"), http.StatusBadRequest, "bad_request"},
		{"ValidationFailed", httpx.ValidationFailed(map[string]string{"email": "ซ้ำ"}), http.StatusUnprocessableEntity, "validation_failed"},
		{"Unauthorized", httpx.ErrUnauthorized, http.StatusUnauthorized, "unauthorized"},
		{"Forbidden", httpx.ErrForbidden, http.StatusForbidden, "forbidden"},
		{"NotFound", httpx.ErrNotFound, http.StatusNotFound, "not_found"},
		{"Conflict", httpx.ErrConflict, http.StatusConflict, "conflict"},
		{"Internal", httpx.ErrInternal, http.StatusInternalServerError, "internal_error"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Status != tc.wantStatus {
				t.Errorf("status = %d ต้องการ %d", tc.err.Status, tc.wantStatus)
			}
			if tc.err.Code != tc.wantCode {
				t.Errorf("code = %q ต้องการ %q", tc.err.Code, tc.wantCode)
			}
			if !specErrorCodes[tc.err.Code] {
				t.Errorf("code %q ไม่อยู่ในรายการ 12 code ของ SPEC", tc.err.Code)
			}
		})
	}
}
