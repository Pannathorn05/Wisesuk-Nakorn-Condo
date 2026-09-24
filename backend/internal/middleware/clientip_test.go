// เทส IP ของผู้เรียกที่บันทึกลง activity log และ log ของ request
//
// สัญญาที่ต้องยืนยัน: header X-Forwarded-For / X-Real-IP ถูกเชื่อก็ต่อเมื่อคำขอวิ่งมาจาก proxy
// ที่อยู่ใน TRUSTED_PROXIES เท่านั้น ไม่งั้นใครก็ส่ง header มาปลอม IP ของตัวเองได้
package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/testsupport"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// clientIPFor ยิงคำขอเข้า engine ที่ตั้ง trusted proxies ตามที่ระบุ แล้วคืน IP ที่ ClientIP อ่านได้
func clientIPFor(t *testing.T, trusted []string, remoteAddr string, headers map[string]string) string {
	t.Helper()

	r := gin.New()
	if err := r.SetTrustedProxies(trusted); err != nil {
		t.Fatalf("SetTrustedProxies: %v", err)
	}
	var got string
	r.GET("/", func(c *gin.Context) { got = middleware.ClientIP(c) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	r.ServeHTTP(httptest.NewRecorder(), req)
	return got
}

func TestClientIP(t *testing.T) {
	cases := []struct {
		name    string
		trusted []string
		remote  string
		headers map[string]string
		want    string
	}{
		{
			name:    "ไม่มี proxy: X-Forwarded-For ที่ client ปลอมมาถูกเมิน",
			remote:  "198.51.100.7:5000",
			headers: map[string]string{"X-Forwarded-For": "1.2.3.4"},
			want:    "198.51.100.7",
		},
		{
			name:    "ไม่มี proxy: X-Real-IP ที่ client ปลอมมาถูกเมิน",
			remote:  "198.51.100.7:5000",
			headers: map[string]string{"X-Real-IP": "1.2.3.4"},
			want:    "198.51.100.7",
		},
		{
			name:    "มาจาก trusted proxy: ใช้ IP ที่ proxy ส่งต่อมา",
			trusted: []string{"10.0.0.1"},
			remote:  "10.0.0.1:5000",
			headers: map[string]string{"X-Forwarded-For": "203.0.113.9"},
			want:    "203.0.113.9",
		},
		{
			// client ใส่ X-Forwarded-For ปลอมมาเอง proxy จึงต่อ IP จริงไว้ท้ายสุด
			// ต้องได้ตัวขวาสุดที่ไม่ใช่ proxy ไม่ใช่ตัวซ้ายสุดที่ client เขียนเอง
			name:    "มาจาก trusted proxy แต่ client แอบใส่ค่าปลอมไว้ข้างหน้า",
			trusted: []string{"10.0.0.1"},
			remote:  "10.0.0.1:5000",
			headers: map[string]string{"X-Forwarded-For": "1.2.3.4, 203.0.113.9"},
			want:    "203.0.113.9",
		},
		{
			name:    "proxy หลายชั้นในวง CIDR เดียวกัน",
			trusted: []string{"10.0.0.0/8"},
			remote:  "10.0.0.1:5000",
			headers: map[string]string{"X-Forwarded-For": "203.0.113.9, 10.0.0.2"},
			want:    "203.0.113.9",
		},
		{
			name:    "ตั้ง trusted ไว้ แต่คำขอไม่ได้มาจาก proxy นั้น",
			trusted: []string{"10.0.0.1"},
			remote:  "198.51.100.7:5000",
			headers: map[string]string{"X-Forwarded-For": "1.2.3.4"},
			want:    "198.51.100.7",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := clientIPFor(t, tc.trusted, tc.remote, tc.headers); got != tc.want {
				t.Errorf("ClientIP = %q ต้องการ %q", got, tc.want)
			}
		})
	}
}

// ต่อสายจริงผ่าน routes.New: IP ที่ลงใน activity_logs ต้องเป็นค่าที่ปลอมไม่ได้
func TestActivityLogIgnoresSpoofedForwardedFor(t *testing.T) {
	cases := []struct {
		name    string
		trusted []string
		want    string
	}{
		{"ไม่ได้ตั้ง TRUSTED_PROXIES", nil, "198.51.100.7"},
		{"ตั้ง TRUSTED_PROXIES เป็น IP ของผู้เรียก", []string{"198.51.100.7"}, "203.0.113.9"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := testsupport.NewApp(t, func(cfg *config.Config) { cfg.TrustedProxies = tc.trusted })

			body := `{"email":"ip@wisetsuk.test","password":"Passw0rd1","first_name":"ไอพี",` +
				`"last_name":"ทดสอบ","phone":"0800000000"}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Forwarded-For", "203.0.113.9")
			req.RemoteAddr = "198.51.100.7:5000"
			rec := httptest.NewRecorder()
			app.Handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusCreated {
				t.Fatalf("สมัครสมาชิกได้ %d (body=%s)", rec.Code, rec.Body.String())
			}

			var ip string
			if err := app.Pool.QueryRow(context.Background(),
				`SELECT ip_address FROM activity_logs WHERE action = 'auth.register'`).Scan(&ip); err != nil {
				t.Fatalf("อ่าน activity log ไม่ได้: %v", err)
			}
			if ip != tc.want {
				t.Errorf("ip_address = %q ต้องการ %q", ip, tc.want)
			}
		})
	}
}
