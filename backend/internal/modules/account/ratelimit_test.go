// เทส rate limit ของ POST /auth/login (AC-2) ยิงผ่าน router ตัวจริง
//
// checklist เทียบกับ AC-2 ใน docs/SPEC.md
//
//	ผิด 5 ครั้งต่อ (อีเมล, IP) -> ครั้งถัดไป 429 + Retry-After   -> TestLoginRateLimitPerEmailAndIP
//	ถูกจำกัดแล้ว รหัสถูกก็ยัง 429                              -> TestLoginRateLimitPerEmailAndIP
//	อีเมลที่ไม่มีในระบบถูกจำกัดแบบเดียวกัน                        -> TestLoginRateLimitUnknownEmail
//	อีเมลเดียวกันจาก IP อื่นไม่ถูกจำกัด                          -> TestLoginRateLimitOtherIPUnaffected
//	ผิด 20 ครั้งต่อ IP ไม่ว่าอีเมลไหน -> 429                       -> TestLoginRateLimitPerIP
//	สำเร็จแล้วล้างตัวนับของ (อีเมล, IP)                           -> TestLoginSuccessResetsCounter
//	ความล้มเหลวเก่ากว่า 15 นาทีไม่นับ                             -> TestLoginOldFailuresExpire
//	account_disabled ไม่นับเป็นความล้มเหลว                       -> TestLoginDisabledAccountNotCounted
package account_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"backend/internal/testsupport"
)

const (
	memberEmail = "member@wisetsuk.test"
	ipA         = "198.51.100.1:5000"
	ipB         = "198.51.100.2:5000"
)

func postLogin(t *testing.T, app testsupport.App, remoteAddr, email, password string) *httptest.ResponseRecorder {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, req)
	return rec
}

func loginErrorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return body.Error.Code
}

// failLogins ยิงรหัสผิด n ครั้ง และยืนยันว่าทุกครั้งยังเป็น 401 (ยังไม่ถึงเพดาน)
func failLogins(t *testing.T, app testsupport.App, remoteAddr, email string, n int) {
	t.Helper()
	for i := range n {
		rec := postLogin(t, app, remoteAddr, email, "wrong-password-1")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("ครั้งที่ %d: status = %d ต้องการ 401 (body=%s)", i+1, rec.Code, rec.Body.String())
		}
	}
}

func assertRateLimited(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d ต้องการ 429 (body=%s)", rec.Code, rec.Body.String())
	}
	if code := loginErrorCode(t, rec); code != "too_many_requests" {
		t.Errorf("code = %q ต้องการ too_many_requests", code)
	}
	secs, err := strconv.Atoi(rec.Header().Get("Retry-After"))
	if err != nil || secs < 1 || secs > 15*60 {
		t.Errorf("Retry-After = %q ต้องเป็นวินาทีระหว่าง 1–900", rec.Header().Get("Retry-After"))
	}
}

func TestLoginRateLimitPerEmailAndIP(t *testing.T) {
	t.Parallel()
	app := testsupport.NewApp(t)

	failLogins(t, app, ipA, memberEmail, 5)

	// ถูกจำกัดแล้ว แม้รหัสผ่านถูกก็ต้องไม่ให้เข้า ไม่งั้นผู้โจมตีรู้ทันทีว่าเดาถูก
	assertRateLimited(t, postLogin(t, app, ipA, memberEmail, testsupport.TestPassword))
}

func TestLoginRateLimitUnknownEmail(t *testing.T) {
	t.Parallel()
	app := testsupport.NewApp(t)

	failLogins(t, app, ipA, "nobody@wisetsuk.test", 5)
	assertRateLimited(t, postLogin(t, app, ipA, "nobody@wisetsuk.test", "wrong-password-1"))
}

// คนอื่นต้องล็อกบัญชีเหยื่อจากเครื่องตัวเองไม่ได้
func TestLoginRateLimitOtherIPUnaffected(t *testing.T) {
	t.Parallel()
	app := testsupport.NewApp(t)

	failLogins(t, app, ipA, memberEmail, 5)

	rec := postLogin(t, app, ipB, memberEmail, testsupport.TestPassword)
	if rec.Code != http.StatusOK {
		t.Fatalf("IP อื่น: status = %d ต้องการ 200 (body=%s)", rec.Code, rec.Body.String())
	}
}

// ไล่เดาหลายบัญชีจาก IP เดียว (password spraying) ต้องถูกจำกัดด้วยตัวนับราย IP
func TestLoginRateLimitPerIP(t *testing.T) {
	t.Parallel()
	app := testsupport.NewApp(t)

	// 20 อีเมลต่างกัน อีเมลละครั้ง — ไม่มีคู่ (อีเมล, IP) ใดถึง 5
	for i := range 20 {
		failLogins(t, app, ipA, fmt.Sprintf("spray%d@wisetsuk.test", i), 1)
	}

	assertRateLimited(t, postLogin(t, app, ipA, memberEmail, testsupport.TestPassword))

	rec := postLogin(t, app, ipB, memberEmail, testsupport.TestPassword)
	if rec.Code != http.StatusOK {
		t.Errorf("IP อื่น: status = %d ต้องการ 200", rec.Code)
	}
}

func TestLoginSuccessResetsCounter(t *testing.T) {
	t.Parallel()
	app := testsupport.NewApp(t)

	failLogins(t, app, ipA, memberEmail, 4)
	if rec := postLogin(t, app, ipA, memberEmail, testsupport.TestPassword); rec.Code != http.StatusOK {
		t.Fatalf("login ถูก: status = %d ต้องการ 200", rec.Code)
	}

	// ตัวนับเริ่มใหม่ ผิดอีก 4 ครั้งต้องยังเป็น 401 ไม่ใช่ 429
	failLogins(t, app, ipA, memberEmail, 4)
}

func TestLoginOldFailuresExpire(t *testing.T) {
	t.Parallel()
	app := testsupport.NewApp(t)

	failLogins(t, app, ipA, memberEmail, 5)

	// เลื่อนความล้มเหลวทั้งหมดไปเป็นอดีตเกินหน้าต่าง 15 นาที
	if _, err := app.Pool.Exec(context.Background(),
		`UPDATE login_attempts SET created_at = now() - interval '16 minutes'`); err != nil {
		t.Fatalf("เลื่อนเวลาไม่สำเร็จ: %v", err)
	}

	if rec := postLogin(t, app, ipA, memberEmail, testsupport.TestPassword); rec.Code != http.StatusOK {
		t.Fatalf("status = %d ต้องการ 200 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestLoginDisabledAccountNotCounted(t *testing.T) {
	t.Parallel()
	app := testsupport.NewApp(t)

	if _, err := app.Pool.Exec(context.Background(),
		`UPDATE users SET is_active = false WHERE id = $1`, app.Fixture.MemberID); err != nil {
		t.Fatal(err)
	}
	for range 6 {
		rec := postLogin(t, app, ipA, memberEmail, testsupport.TestPassword)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d ต้องการ 403 account_disabled", rec.Code)
		}
	}
}
