// เทสของ package oauth — ไม่แตะฐานข้อมูลและไม่ยิงออกอินเทอร์เน็ตจริง
//
// ส่วนที่ต้องคุยกับ provider ถูกทดสอบด้วย httptest.Server ที่ปลอมเป็น Google/Facebook
// จึงทดสอบได้ทั้งกรณีที่ provider ตอบครบและกรณีที่ตอบไม่ครบ (ซึ่งเกิดขึ้นจริงกับ Facebook)
package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func testConfig() Config {
	return Config{
		GoogleClientID:       "google-client",
		GoogleClientSecret:   "google-secret",
		FacebookClientID:     "facebook-client",
		FacebookClientSecret: "facebook-secret",
		CallbackBaseURL:      "https://api.wisetsuk.test/",
	}
}

func TestEnabledOnlyFullyConfiguredProviders(t *testing.T) {
	m := New(Config{GoogleClientID: "id", GoogleClientSecret: "secret"})

	if !m.Enabled(ProviderGoogle) {
		t.Error("google ตั้งค่าครบแล้วต้องเปิดใช้งาน")
	}
	if m.Enabled(ProviderFacebook) {
		t.Error("facebook ไม่ได้ตั้งค่า ต้องไม่เปิดใช้งาน")
	}
	if got := m.Names(); len(got) != 1 || got[0] != ProviderGoogle {
		t.Errorf("Names() = %v ต้องได้เฉพาะ google", got)
	}
}

// ตั้งค่า client id มาแต่ไม่ตั้ง secret ถือว่าไม่ครบ ปล่อยผ่านไปจะพังตอนแลก token
// ซึ่งสายเกินไปแล้วเพราะผู้ใช้ถูกพาออกไปหน้า provider เรียบร้อย
func TestEnabledRejectsHalfConfiguredProvider(t *testing.T) {
	m := New(Config{GoogleClientID: "id"})
	if m.Enabled(ProviderGoogle) {
		t.Error("ขาด client secret ต้องไม่ถือว่าเปิดใช้งาน")
	}
}

func TestRedirectURITrimsTrailingSlash(t *testing.T) {
	m := New(testConfig())

	want := "https://api.wisetsuk.test/api/v1/auth/oauth/google/callback"
	if got := m.RedirectURI(ProviderGoogle); got != want {
		t.Errorf("RedirectURI() = %q ต้องเป็น %q", got, want)
	}
}

func TestAuthCodeURLCarriesRequiredParams(t *testing.T) {
	m := New(testConfig())

	raw, err := m.AuthCodeURL(ProviderGoogle, "state-123", "challenge-abc")
	if err != nil {
		t.Fatalf("AuthCodeURL คืน error: %v", err)
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("URL ที่ได้ parse ไม่ผ่าน: %v", err)
	}
	q := parsed.Query()

	for field, want := range map[string]string{
		"client_id":             "google-client",
		"response_type":         "code",
		"state":                 "state-123",
		"code_challenge":        "challenge-abc",
		"code_challenge_method": "S256",
		"redirect_uri":          m.RedirectURI(ProviderGoogle),
	} {
		if got := q.Get(field); got != want {
			t.Errorf("%s = %q ต้องเป็น %q", field, got, want)
		}
	}
}

func TestAuthCodeURLRejectsUnknownProvider(t *testing.T) {
	m := New(testConfig())

	if _, err := m.AuthCodeURL("line", "state", "challenge"); !errors.Is(err, ErrUnknownProvider) {
		t.Errorf("err = %v ต้องเป็น ErrUnknownProvider", err)
	}
}

// PKCE จะกัน code ที่ถูกดักได้ก็ต่อเมื่อ challenge เป็น SHA-256 ของ verifier จริง ๆ
func TestNewHandshakeDerivesChallengeFromVerifier(t *testing.T) {
	hs, err := NewHandshake()
	if err != nil {
		t.Fatalf("NewHandshake คืน error: %v", err)
	}

	sum := sha256.Sum256([]byte(hs.Verifier))
	want := base64.RawURLEncoding.EncodeToString(sum[:])
	if hs.Challenge != want {
		t.Errorf("Challenge = %q ต้องเป็น SHA-256 ของ Verifier (%q)", hs.Challenge, want)
	}
	if hs.State == "" || hs.State == hs.Verifier {
		t.Error("State ต้องมีค่าและต้องไม่ซ้ำกับ Verifier")
	}
}

func TestNewHandshakeIsUniquePerCall(t *testing.T) {
	first, err := NewHandshake()
	if err != nil {
		t.Fatalf("NewHandshake คืน error: %v", err)
	}
	second, err := NewHandshake()
	if err != nil {
		t.Fatalf("NewHandshake คืน error: %v", err)
	}

	if first.State == second.State || first.Verifier == second.Verifier {
		t.Error("ค่าสุ่มสองครั้งติดกันต้องไม่เท่ากัน")
	}
}

// ---------------------------------------------------------------- แปลงโปรไฟล์

func TestParseGoogleReadsAllFields(t *testing.T) {
	body := []byte(`{
		"sub": "1122334455",
		"email": "  Somchai@GMAIL.com ",
		"email_verified": true,
		"given_name": "สมชาย",
		"family_name": "ใจดี",
		"picture": "https://lh3.googleusercontent.com/a/abc"
	}`)

	got, err := parseGoogle(body)
	if err != nil {
		t.Fatalf("parseGoogle คืน error: %v", err)
	}

	if got.Subject != "1122334455" {
		t.Errorf("Subject = %q", got.Subject)
	}
	// อีเมลต้องถูก normalize ตั้งแต่ตรงนี้ เพราะคอลัมน์ users.email เก็บตัวพิมพ์เล็กเสมอ
	if got.Email != "somchai@gmail.com" {
		t.Errorf("Email = %q ต้องเป็นตัวพิมพ์เล็กและไม่มีช่องว่าง", got.Email)
	}
	if !got.EmailVerified {
		t.Error("EmailVerified ต้องเป็น true")
	}
	if got.FirstName != "สมชาย" || got.LastName != "ใจดี" {
		t.Errorf("ชื่อ = %q %q", got.FirstName, got.LastName)
	}
}

// Google ส่ง given_name/family_name มาเกือบทุกครั้ง แต่ไม่การันตี ถ้าไม่มีต้องถอยไปใช้ name
func TestParseGoogleFallsBackToFullName(t *testing.T) {
	got, err := parseGoogle([]byte(`{"sub":"1","email":"a@b.c","email_verified":true,"name":"สมหญิง รักเรียน"}`))
	if err != nil {
		t.Fatalf("parseGoogle คืน error: %v", err)
	}
	if got.FirstName != "สมหญิง" || got.LastName != "รักเรียน" {
		t.Errorf("ชื่อ = %q %q ต้องแยกจาก name", got.FirstName, got.LastName)
	}
}

func TestParseFacebookTreatsReturnedEmailAsVerified(t *testing.T) {
	got, err := parseFacebook([]byte(`{
		"id": "998877",
		"first_name": "มานี",
		"last_name": "มีนา",
		"email": "Manee@example.com",
		"picture": {"data": {"url": "https://scontent.example/pic.jpg"}}
	}`))
	if err != nil {
		t.Fatalf("parseFacebook คืน error: %v", err)
	}

	if got.Email != "manee@example.com" || !got.EmailVerified {
		t.Errorf("Email = %q verified = %v", got.Email, got.EmailVerified)
	}
	if got.AvatarURL != "https://scontent.example/pic.jpg" {
		t.Errorf("AvatarURL = %q", got.AvatarURL)
	}
}

func TestSplitName(t *testing.T) {
	cases := []struct{ in, first, last string }{
		{"สมชาย ใจดี", "สมชาย", "ใจดี"},
		{"สมชาย ใจดี มากมาย", "สมชาย", "ใจดี มากมาย"},
		{"Cher", "Cher", ""},
		{"   ", "", ""},
	}
	for _, c := range cases {
		first, last := splitName(c.in)
		if first != c.first || last != c.last {
			t.Errorf("splitName(%q) = %q, %q ต้องเป็น %q, %q", c.in, first, last, c.first, c.last)
		}
	}
}

// ---------------------------------------------------------------- คุยกับ provider

// newFakeProvider แทนที่ปลายทางของ provider ด้วย httptest.Server ที่เราคุมคำตอบได้
func newFakeProvider(t *testing.T, tokenBody, profileBody string, profileStatus int) *Manager {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("อ่าน form ของคำขอ token ไม่ได้: %v", err)
		}
		// PKCE verifier ต้องถูกส่งไปด้วยเสมอ ไม่งั้น provider จริงจะปฏิเสธ
		if r.PostForm.Get("code_verifier") == "" {
			t.Error("คำขอ token ต้องแนบ code_verifier")
		}
		if r.PostForm.Get("grant_type") != "authorization_code" {
			t.Errorf("grant_type = %q", r.PostForm.Get("grant_type"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(tokenBody))
	})
	mux.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer provider-token" {
			t.Errorf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(profileStatus)
		_, _ = w.Write([]byte(profileBody))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	m := New(testConfig())
	m.providers[ProviderGoogle].tokenURL = srv.URL + "/token"
	m.providers[ProviderGoogle].profileURL = srv.URL + "/profile"
	return m
}

func TestExchangeReturnsProfileOnHappyPath(t *testing.T) {
	m := newFakeProvider(t,
		`{"access_token":"provider-token","token_type":"Bearer","expires_in":3600}`,
		`{"sub":"42","email":"member@example.com","email_verified":true,"given_name":"ทดสอบ","family_name":"ระบบ"}`,
		http.StatusOK)

	got, err := m.Exchange(context.Background(), ProviderGoogle, "auth-code", "verifier")
	if err != nil {
		t.Fatalf("Exchange คืน error: %v", err)
	}
	if got.Provider != ProviderGoogle || got.Subject != "42" || got.Email != "member@example.com" {
		t.Errorf("โปรไฟล์ที่ได้ = %+v", got)
	}
}

// เคสจริงของ Facebook: ผู้ใช้กดไม่อนุญาตให้เข้าถึงอีเมล
func TestExchangeRejectsMissingEmail(t *testing.T) {
	m := newFakeProvider(t,
		`{"access_token":"provider-token"}`,
		`{"sub":"42","email_verified":true,"given_name":"ไม่มี","family_name":"อีเมล"}`,
		http.StatusOK)

	if _, err := m.Exchange(context.Background(), ProviderGoogle, "code", "verifier"); !errors.Is(err, ErrNoEmail) {
		t.Errorf("err = %v ต้องเป็น ErrNoEmail", err)
	}
}

// อีเมลที่ยังไม่ยืนยันต้องไม่ถูกใช้ผูกบัญชี ไม่งั้นใครก็อ้างอีเมลของคนอื่นเข้ามาสวมบัญชีได้
func TestExchangeRejectsUnverifiedEmail(t *testing.T) {
	m := newFakeProvider(t,
		`{"access_token":"provider-token"}`,
		`{"sub":"42","email":"member@example.com","email_verified":false}`,
		http.StatusOK)

	if _, err := m.Exchange(context.Background(), ProviderGoogle, "code", "verifier"); !errors.Is(err, ErrEmailNotVerified) {
		t.Errorf("err = %v ต้องเป็น ErrEmailNotVerified", err)
	}
}

func TestExchangeRejectsMissingAccessToken(t *testing.T) {
	m := newFakeProvider(t, `{"error":"invalid_grant"}`, `{}`, http.StatusOK)

	if _, err := m.Exchange(context.Background(), ProviderGoogle, "code", "verifier"); err == nil {
		t.Error("ไม่มี access token ต้องคืน error")
	}
}

func TestExchangeRejectsProviderFailure(t *testing.T) {
	m := newFakeProvider(t, `{"access_token":"provider-token"}`, `{"error":"เกิดข้อผิดพลาด"}`, http.StatusInternalServerError)

	if _, err := m.Exchange(context.Background(), ProviderGoogle, "code", "verifier"); err == nil {
		t.Error("provider ตอบ 500 ต้องคืน error")
	}
}
