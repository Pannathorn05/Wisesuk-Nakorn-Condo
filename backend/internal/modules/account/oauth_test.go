// เทสการเข้าสู่ระบบด้วยบัญชีภายนอก (Google / Facebook) ที่ชั้น service
//
// จุดที่คุ้มค่าทดสอบที่สุดของเรื่องนี้คือ "ตัดสินใจว่าโปรไฟล์ที่ได้มาเป็นใคร" ไม่ใช่การคุยกับ
// provider (ซึ่งมีเทสของตัวเองอยู่ที่ internal/oauth) เทสชุดนี้จึงป้อน oauth.Profile เข้าไปตรง ๆ
// บนฐานข้อมูลจริง แล้วตรวจว่าได้บัญชีที่ถูกคนและไม่เกิดบัญชีซ้ำ
package account_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"backend/internal/config"
	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/modules/account"
	"backend/internal/oauth"
	"backend/internal/server"
	"backend/internal/storage"
	"backend/internal/testsupport"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// testStaffDefaultPassword คือค่า STAFF_DEFAULT_PASSWORD ของ environment ทดสอบ
// บัญชีผู้ดูแลที่หัวหน้าผู้ดูแลสร้างจะได้รหัสนี้ แล้วถูกบังคับให้เปลี่ยนก่อนใช้งานจริง
const testStaffDefaultPassword = "StaffDefault!2026"

type env struct {
	svc     *account.Service
	repo    *account.Repository
	pool    *pgxpool.Pool
	fixture testsupport.Fixture
}

// newEnv ประกอบ module account ตัวจริงผ่าน server.New เดียวกับที่ cmd/api ใช้
// เทสจึงได้ service ที่ต่อสายพึ่งพาเหมือนของจริงทุกประการ
func newEnv(t *testing.T) env {
	t.Helper()

	pool := testsupport.NewDatabase(t)
	fixture := testsupport.Seed(t, pool)
	dir := testsupport.UploadDir(t)

	files, err := storage.NewLocalStore(dir, "http://localhost:8080", 5<<20)
	if err != nil {
		t.Fatalf("สร้าง storage ไม่สำเร็จ: %v", err)
	}

	cfg := &config.Config{
		Env:                  "test",
		Port:                 "8080",
		JWTSecret:            testsupport.JWTSecret,
		AccessTokenTTL:       testsupport.AccessTTL,
		RefreshTokenTTL:      testsupport.RefreshTTL,
		UploadDir:            dir,
		PublicBaseURL:        "http://localhost:8080",
		MaxUploadBytes:       5 << 20,
		AllowedOrigins:       []string{"http://localhost:3000"},
		BcryptCost:           bcrypt.MinCost,
		StaffDefaultPassword: testStaffDefaultPassword,

		GoogleClientID:           "test-client",
		GoogleClientSecret:       "test-secret",
		FrontendOAuthCallbackURL: "http://localhost:3000/auth/callback",
	}

	srv := server.New(cfg, pool, files)
	return env{svc: srv.Account.Service, repo: srv.Account.Repo, pool: pool, fixture: fixture}
}

func googleProfile(subject, email, first, last string) oauth.Profile {
	return oauth.Profile{
		Provider:      oauth.ProviderGoogle,
		Subject:       subject,
		Email:         email,
		EmailVerified: true,
		FirstName:     first,
		LastName:      last,
	}
}

// login เดินครบสองขั้นเหมือนของจริง: callback ออกโค้ด แล้ว frontend เอาโค้ดมาแลก token
func login(t *testing.T, e env, p oauth.Profile) *account.TokenPair {
	t.Helper()
	ctx := context.Background()

	code, err := e.svc.LoginWithProvider(ctx, p, "127.0.0.1")
	if err != nil {
		t.Fatalf("LoginWithProvider คืน error: %v", err)
	}
	tokens, err := e.svc.ExchangeOAuthCode(ctx, code)
	if err != nil {
		t.Fatalf("ExchangeOAuthCode คืน error: %v", err)
	}
	return tokens
}

// เข้าครั้งแรกด้วยอีเมลที่ยังไม่มีในระบบ = สมัครสมาชิกใหม่ ต้องได้สิทธิ์ member เท่านั้น
func TestOAuthFirstLoginCreatesMember(t *testing.T) {
	e := newEnv(t)

	tokens := login(t, e, googleProfile("google-1", "newcomer@example.com", "สมชาย", "ใจดี"))

	if tokens.User.Email != "newcomer@example.com" {
		t.Errorf("อีเมล = %q", tokens.User.Email)
	}
	if tokens.User.Role != "member" {
		t.Errorf("role = %q ต้องเป็น member เสมอ", tokens.User.Role)
	}
	if tokens.User.BranchID != nil {
		t.Error("สมาชิกต้องไม่ผูกกับสาขา")
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Error("ต้องได้ทั้ง access token และ refresh token")
	}
}

// เข้าซ้ำด้วยบัญชีเดิมต้องเป็นคนเดิม ไม่ใช่สร้างใหม่ทุกครั้ง
func TestOAuthSecondLoginReusesSameUser(t *testing.T) {
	e := newEnv(t)
	p := googleProfile("google-1", "newcomer@example.com", "สมชาย", "ใจดี")

	first := login(t, e, p)
	second := login(t, e, p)

	if first.User.ID != second.User.ID {
		t.Errorf("ได้ผู้ใช้คนละคน: %s กับ %s", first.User.ID, second.User.ID)
	}
	if n := countUsers(t, e.pool, "newcomer@example.com"); n != 1 {
		t.Errorf("มีผู้ใช้ %d ใบด้วยอีเมลเดียวกัน ต้องมีใบเดียว", n)
	}
}

// ผู้ใช้ที่เปลี่ยนอีเมลฝั่ง Google ต้องยังเป็นคนเดิม เพราะจับคู่ด้วยรหัสผู้ใช้ของ provider
func TestOAuthMatchesBySubjectNotEmail(t *testing.T) {
	e := newEnv(t)

	first := login(t, e, googleProfile("google-1", "old@example.com", "สมชาย", "ใจดี"))
	second := login(t, e, googleProfile("google-1", "new@example.com", "สมชาย", "ใจดี"))

	if first.User.ID != second.User.ID {
		t.Error("เปลี่ยนอีเมลที่ provider แล้วต้องยังเป็นผู้ใช้คนเดิม")
	}
}

// อีเมลตรงกับบัญชีที่สมัครด้วยรหัสผ่านไว้แล้ว ต้องผูกเข้ากับบัญชีเดิม ไม่ใช่สร้างใบใหม่
// (ถ้าสร้างใหม่จะชน UNIQUE ของคอลัมน์ email และผู้ใช้จะเข้าระบบไม่ได้เลย)
func TestOAuthLinksToExistingPasswordAccount(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	tokens := login(t, e, googleProfile("google-9", "member@wisetsuk.test", "สมาชิก", "ทดสอบ"))

	if tokens.User.ID != e.fixture.MemberID {
		t.Fatalf("ได้ผู้ใช้ %s ต้องเป็นบัญชีเดิม %s", tokens.User.ID, e.fixture.MemberID)
	}

	providers, err := e.repo.ListIdentities(ctx, e.fixture.MemberID)
	if err != nil {
		t.Fatalf("ListIdentities คืน error: %v", err)
	}
	if len(providers) != 1 || providers[0] != oauth.ProviderGoogle {
		t.Errorf("identities = %v ต้องมี google หนึ่งรายการ", providers)
	}
}

// บัญชีที่ผูก Google ไว้แล้วต้องยังเข้าด้วยรหัสผ่านเดิมได้ตามปกติ
func TestPasswordLoginStillWorksAfterLinking(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	login(t, e, googleProfile("google-9", "member@wisetsuk.test", "สมาชิก", "ทดสอบ"))

	tokens, err := e.svc.Login(ctx, account.LoginInput{
		Email:    "member@wisetsuk.test",
		Password: testsupport.TestPassword,
	}, "127.0.0.1")
	if err != nil {
		t.Fatalf("เข้าด้วยรหัสผ่านเดิมไม่สำเร็จ: %v", err)
	}
	if tokens.User.ID != e.fixture.MemberID {
		t.Error("ได้ผู้ใช้ผิดคน")
	}
}

// บัญชีที่เกิดจาก OAuth ยังไม่มีรหัสผ่าน ใครเดารหัสอย่างไรก็ต้องเข้าไม่ได้
func TestPasswordLoginRejectedForOAuthOnlyAccount(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	login(t, e, googleProfile("google-1", "newcomer@example.com", "สมชาย", "ใจดี"))

	for _, password := range []string{"", " ", testsupport.TestPassword} {
		_, err := e.svc.Login(ctx, account.LoginInput{
			Email: "newcomer@example.com", Password: password,
		}, "127.0.0.1")
		if !isCode(err, httpx.CodeInvalidCredentials) {
			t.Errorf("รหัสผ่าน %q: err = %v ต้องเป็น invalid_credentials", password, err)
		}
	}
}

// ผู้ใช้ที่เข้ามาทาง Google ต้องตั้งรหัสผ่านของตัวเองได้ โดยไม่ต้องกรอก "รหัสผ่านเดิม"
// ที่เขาไม่เคยมี
func TestOAuthUserCanSetFirstPassword(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	tokens := login(t, e, googleProfile("google-1", "newcomer@example.com", "สมชาย", "ใจดี"))
	identity := middleware.Identity{
		UserID: tokens.User.ID, Role: tokens.User.Role, Name: tokens.User.FullName(),
	}

	const newPassword = "Wisetsuk!2027"
	err := e.svc.ChangePassword(ctx, identity, account.ChangePasswordInput{
		NewPassword: newPassword, ConfirmPassword: newPassword,
	}, "127.0.0.1")
	if err != nil {
		t.Fatalf("ตั้งรหัสผ่านครั้งแรกไม่สำเร็จ: %v", err)
	}

	if _, err := e.svc.Login(ctx, account.LoginInput{
		Email: "newcomer@example.com", Password: newPassword,
	}, "127.0.0.1"); err != nil {
		t.Fatalf("ตั้งรหัสผ่านแล้วเข้าด้วยรหัสผ่านไม่ได้: %v", err)
	}
}

// ตั้งรหัสผ่านแล้วครั้งต่อไปต้องกลับไปบังคับกรอกรหัสผ่านเดิมเหมือนบัญชีทั่วไป
func TestChangePasswordStillRequiresCurrentAfterFirstSet(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	identity := middleware.Identity{UserID: e.fixture.MemberID, Role: "member", Name: "สมาชิก ทดสอบ"}
	const newPassword = "Wisetsuk!2027"

	err := e.svc.ChangePassword(ctx, identity, account.ChangePasswordInput{
		NewPassword: newPassword, ConfirmPassword: newPassword,
	}, "127.0.0.1")
	if !isCode(err, httpx.CodeValidationFailed) {
		t.Errorf("err = %v ต้องเป็น validation_failed เพราะไม่ได้ส่ง current_password", err)
	}
}

// โค้ดแลก token ใช้ได้ครั้งเดียว ถ้าลิงก์ callback รั่วไปแล้วถูกเปิดซ้ำต้องใช้ไม่ได้
func TestExchangeCodeIsSingleUse(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	code, err := e.svc.LoginWithProvider(ctx,
		googleProfile("google-1", "newcomer@example.com", "สมชาย", "ใจดี"), "127.0.0.1")
	if err != nil {
		t.Fatalf("LoginWithProvider คืน error: %v", err)
	}

	if _, err := e.svc.ExchangeOAuthCode(ctx, code); err != nil {
		t.Fatalf("แลกครั้งแรกต้องสำเร็จ: %v", err)
	}
	if _, err := e.svc.ExchangeOAuthCode(ctx, code); !isCode(err, httpx.CodeUnauthorized) {
		t.Errorf("err = %v ต้องเป็น unauthorized เพราะโค้ดถูกใช้ไปแล้ว", err)
	}
}

func TestExchangeRejectsUnknownCode(t *testing.T) {
	e := newEnv(t)

	for _, code := range []string{"", "ไม่เคยมีโค้ดนี้"} {
		if _, err := e.svc.ExchangeOAuthCode(context.Background(), code); !isCode(err, httpx.CodeUnauthorized) {
			t.Errorf("code %q: err = %v ต้องเป็น unauthorized", code, err)
		}
	}
}

// บัญชีที่ถูกระงับต้องเข้าไม่ได้ทุกทาง รวมถึงทาง Google ที่ข้ามหน้ากรอกรหัสผ่านไป
func TestOAuthRejectsDisabledAccount(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	if _, err := e.pool.Exec(ctx,
		`UPDATE users SET is_active = FALSE WHERE id = $1`, e.fixture.MemberID); err != nil {
		t.Fatalf("ปิดบัญชีไม่สำเร็จ: %v", err)
	}

	_, err := e.svc.LoginWithProvider(ctx,
		googleProfile("google-9", "member@wisetsuk.test", "สมาชิก", "ทดสอบ"), "127.0.0.1")
	if !isCode(err, httpx.CodeAccountDisabled) {
		t.Errorf("err = %v ต้องเป็น account_disabled", err)
	}
}

// การเข้าสู่ระบบด้วยบัญชีภายนอกต้องถูกบันทึกลง activity log เหมือนช่องทางอื่น
func TestOAuthLoginIsAudited(t *testing.T) {
	e := newEnv(t)

	login(t, e, googleProfile("google-1", "newcomer@example.com", "สมชาย", "ใจดี"))

	var count int
	err := e.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM activity_logs WHERE action = 'auth.oauth_register'`).Scan(&count)
	if err != nil {
		t.Fatalf("อ่าน activity_logs ไม่สำเร็จ: %v", err)
	}
	if count != 1 {
		t.Errorf("มี log %d รายการ ต้องมี 1", count)
	}
}

// ---------------------------------------------------------------- helpers

func countUsers(t *testing.T, pool *pgxpool.Pool, email string) int {
	t.Helper()

	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM users WHERE email = $1`, email).Scan(&n); err != nil {
		t.Fatalf("นับผู้ใช้ไม่สำเร็จ: %v", err)
	}
	return n
}

// isCode ตรวจว่า error ที่ได้เป็น APIError รหัสที่ต้องการ ไม่ใช่แค่ "มี error"
func isCode(err error, code string) bool {
	var apiErr *httpx.APIError
	return errors.As(err, &apiErr) && apiErr.Code == code
}
