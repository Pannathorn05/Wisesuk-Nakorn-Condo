// Package oauth คุยกับผู้ให้บริการล็อกอินภายนอก (Google, Facebook) ด้วย Authorization Code + PKCE
//
// ที่ไม่ใช้ไลบรารีสำเร็จรูปเพราะสิ่งที่ต้องการมีแค่สามอย่าง: ประกอบ URL หน้ายินยอม,
// แลก code เป็น access token ของ provider, และขอโปรไฟล์กลับมาหนึ่งชุด ทั้งหมดเป็น
// HTTP ธรรมดาที่ net/http ทำได้ครบ การเพิ่ม dependency จึงไม่คุ้ม
//
// package นี้ไม่รู้จักฐานข้อมูลหรือ user ในระบบเราเลย หน้าที่เดียวคือคืน Profile
// ส่วนการตัดสินใจว่าจะผูกกับบัญชีไหนเป็นเรื่องของ module account
package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ชื่อ provider ต้องตรงกับ CHECK constraint ของคอลัมน์ user_identities.provider
const (
	ProviderGoogle   = "google"
	ProviderFacebook = "facebook"
)

var (
	// ErrUnknownProvider คือ provider ที่ไม่รู้จัก หรือรู้จักแต่ยังไม่ได้ตั้ง client id/secret
	ErrUnknownProvider = errors.New("oauth: ไม่รองรับผู้ให้บริการนี้")

	// ErrNoEmail เกิดเมื่อ provider ไม่ส่งอีเมลกลับมา — Facebook ปฏิเสธได้ถ้าผู้ใช้ไม่ยินยอม
	ErrNoEmail = errors.New("oauth: ผู้ให้บริการไม่ได้ส่งอีเมลกลับมา")

	// ErrEmailNotVerified กันการสวมบัญชี: ถ้า provider ยังไม่ยืนยันอีเมล ใครก็อ้างอีเมล
	// ของคนอื่นแล้วเข้ามาผูกกับบัญชีเดิมในระบบเราได้
	ErrEmailNotVerified = errors.New("oauth: อีเมลยังไม่ได้รับการยืนยันจากผู้ให้บริการ")
)

// Profile คือข้อมูลผู้ใช้รูปแบบกลาง แปลงจากคำตอบของแต่ละ provider แล้ว
type Profile struct {
	Provider      string
	Subject       string // รหัสผู้ใช้ฝั่ง provider — ใช้ค่านี้จับคู่บัญชี ไม่ใช่อีเมล
	Email         string
	EmailVerified bool
	FirstName     string
	LastName      string
	AvatarURL     string
}

// Config คือค่าที่มาจาก environment ทั้งหมด
type Config struct {
	GoogleClientID     string
	GoogleClientSecret string

	FacebookClientID     string
	FacebookClientSecret string

	// CallbackBaseURL คือ origin ของ backend ที่ provider จะ redirect กลับมา
	// ต้องตรงเป๊ะกับที่ลงทะเบียนไว้ใน Google Cloud Console / Meta for Developers
	CallbackBaseURL string
}

type provider struct {
	name         string
	clientID     string
	clientSecret string
	authURL      string
	tokenURL     string
	profileURL   string
	scopes       []string
	parse        func([]byte) (Profile, error)
}

type Manager struct {
	providers    map[string]*provider
	callbackBase string
	client       *http.Client
}

// New สร้าง Manager ที่มีเฉพาะ provider ที่ตั้ง client id และ secret ไว้ครบ
// provider ที่ไม่ได้ตั้งค่าถือว่าไม่รองรับ ระบบจึงรันได้ปกติแม้เปิดใช้เจ้าเดียวหรือไม่เปิดเลย
func New(cfg Config) *Manager {
	m := &Manager{
		providers:    make(map[string]*provider, 2),
		callbackBase: strings.TrimRight(cfg.CallbackBaseURL, "/"),
		// timeout สั้นเพราะเป็นการเรียกที่คั่นกลาง request ของผู้ใช้จริง
		client: &http.Client{Timeout: 10 * time.Second},
	}

	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		m.providers[ProviderGoogle] = &provider{
			name:         ProviderGoogle,
			clientID:     cfg.GoogleClientID,
			clientSecret: cfg.GoogleClientSecret,
			authURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			tokenURL:     "https://oauth2.googleapis.com/token",
			profileURL:   "https://openidconnect.googleapis.com/v1/userinfo",
			scopes:       []string{"openid", "email", "profile"},
			parse:        parseGoogle,
		}
	}

	if cfg.FacebookClientID != "" && cfg.FacebookClientSecret != "" {
		m.providers[ProviderFacebook] = &provider{
			name:         ProviderFacebook,
			clientID:     cfg.FacebookClientID,
			clientSecret: cfg.FacebookClientSecret,
			authURL:      "https://www.facebook.com/v21.0/dialog/oauth",
			tokenURL:     "https://graph.facebook.com/v21.0/oauth/access_token",
			profileURL:   "https://graph.facebook.com/v21.0/me?fields=id,first_name,last_name,name,email,picture.width(256)",
			scopes:       []string{"email", "public_profile"},
			parse:        parseFacebook,
		}
	}

	return m
}

// Enabled บอกว่า provider นี้เปิดใช้งานอยู่หรือไม่
func (m *Manager) Enabled(name string) bool {
	_, ok := m.providers[name]
	return ok
}

// Names คืนรายชื่อ provider ที่เปิดใช้งาน เรียงคงที่เพื่อให้ผลลัพธ์ทดสอบได้
func (m *Manager) Names() []string {
	out := make([]string, 0, len(m.providers))
	for _, name := range []string{ProviderGoogle, ProviderFacebook} {
		if _, ok := m.providers[name]; ok {
			out = append(out, name)
		}
	}
	return out
}

// RedirectURI คือ URL ที่ provider จะ redirect กลับมา ต้องเหมือนกันเป๊ะทั้งตอนขอ code
// และตอนแลก token ไม่งั้น provider ปฏิเสธ (ข้อกำหนดของ OAuth 2.0 ไม่ใช่ของเรา)
func (m *Manager) RedirectURI(name string) string {
	return m.callbackBase + "/api/v1/auth/oauth/" + name + "/callback"
}

// AuthCodeURL ประกอบ URL หน้ายินยอมของ provider
//
// state กัน CSRF (ผู้โจมตีหลอกให้เหยื่อกด callback ที่ผูกกับบัญชีของผู้โจมตีเอง)
// challenge คือ PKCE กันกรณี code รั่วระหว่างทางแล้วถูกเอาไปแลก token
func (m *Manager) AuthCodeURL(name, state, challenge string) (string, error) {
	p, ok := m.providers[name]
	if !ok {
		return "", ErrUnknownProvider
	}

	q := url.Values{}
	q.Set("client_id", p.clientID)
	q.Set("redirect_uri", m.RedirectURI(name))
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(p.scopes, " "))
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	if p.name == ProviderGoogle {
		// บังคับให้เลือกบัญชีทุกครั้ง ไม่งั้นเครื่องที่ค้าง login Google ไว้จะเด้งกลับทันที
		// โดยผู้ใช้ไม่ทันเห็นว่ากำลังเข้าด้วยบัญชีไหน
		q.Set("prompt", "select_account")
	}

	return p.authURL + "?" + q.Encode(), nil
}

// Exchange แลก code เป็นโปรไฟล์ผู้ใช้ — รวมสองขั้น (แลก token แล้วขอโปรไฟล์) ไว้ด้วยกัน
// เพราะ access token ของ provider ไม่มีประโยชน์อื่นกับระบบเรา และไม่ควรถูกเก็บไว้ที่ใดเลย
func (m *Manager) Exchange(ctx context.Context, name, code, verifier string) (Profile, error) {
	p, ok := m.providers[name]
	if !ok {
		return Profile{}, ErrUnknownProvider
	}

	token, err := m.exchangeCode(ctx, p, code, verifier)
	if err != nil {
		return Profile{}, err
	}
	return m.fetchProfile(ctx, p, token)
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (m *Manager) exchangeCode(ctx context.Context, p *provider, code, verifier string) (string, error) {
	form := url.Values{}
	form.Set("client_id", p.clientID)
	form.Set("client_secret", p.clientSecret)
	form.Set("code", code)
	form.Set("code_verifier", verifier)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", m.RedirectURI(p.name))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	body, err := m.do(req)
	if err != nil {
		return "", fmt.Errorf("oauth: แลก code กับ %s ไม่สำเร็จ: %w", p.name, err)
	}

	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", fmt.Errorf("oauth: อ่านคำตอบ token ของ %s ไม่ได้: %w", p.name, err)
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("oauth: %s ไม่ได้ส่ง access token กลับมา", p.name)
	}
	return tok.AccessToken, nil
}

func (m *Manager) fetchProfile(ctx context.Context, p *provider, token string) (Profile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.profileURL, nil)
	if err != nil {
		return Profile{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	body, err := m.do(req)
	if err != nil {
		return Profile{}, fmt.Errorf("oauth: ขอโปรไฟล์จาก %s ไม่สำเร็จ: %w", p.name, err)
	}

	profile, err := p.parse(body)
	if err != nil {
		return Profile{}, err
	}
	profile.Provider = p.name

	if profile.Subject == "" {
		return Profile{}, fmt.Errorf("oauth: %s ไม่ได้ส่งรหัสผู้ใช้กลับมา", p.name)
	}
	if profile.Email == "" {
		return Profile{}, ErrNoEmail
	}
	if !profile.EmailVerified {
		return Profile{}, ErrEmailNotVerified
	}
	return profile, nil
}

// do ยิง request แล้วอ่าน body ทั้งก้อน จำกัดขนาดกัน provider (หรือคนที่ปลอมเป็น provider)
// ตอบ body ยาวไม่รู้จบจนกินหน่วยความจำของเรา
func (m *Manager) do(req *http.Request) ([]byte, error) {
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// ตัดข้อความให้สั้นก่อนใส่ error — ข้อความนี้ไปจบที่ log ฝั่งเราเท่านั้น ไม่ถึงผู้ใช้
		return nil, fmt.Errorf("สถานะ %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	return body, nil
}

// ---------------------------------------------------------------- แปลงคำตอบรายเจ้า

// parseGoogle — Google ตอบตามมาตรฐาน OpenID Connect
func parseGoogle(body []byte) (Profile, error) {
	var raw struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return Profile{}, fmt.Errorf("oauth: อ่านโปรไฟล์ google ไม่ได้: %w", err)
	}

	first, last := raw.GivenName, raw.FamilyName
	if first == "" && last == "" {
		first, last = splitName(raw.Name)
	}
	return Profile{
		Subject:       raw.Sub,
		Email:         strings.ToLower(strings.TrimSpace(raw.Email)),
		EmailVerified: raw.EmailVerified,
		FirstName:     first,
		LastName:      last,
		AvatarURL:     raw.Picture,
	}, nil
}

// parseFacebook — Graph API มีรูปแบบของตัวเอง และส่ง email มาเฉพาะเมื่อผู้ใช้ยินยอม
func parseFacebook(body []byte) (Profile, error) {
	var raw struct {
		ID        string `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Picture   struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return Profile{}, fmt.Errorf("oauth: อ่านโปรไฟล์ facebook ไม่ได้: %w", err)
	}

	first, last := raw.FirstName, raw.LastName
	if first == "" && last == "" {
		first, last = splitName(raw.Name)
	}
	email := strings.ToLower(strings.TrimSpace(raw.Email))
	return Profile{
		Subject: raw.ID,
		Email:   email,
		// Facebook ไม่มีฟิลด์บอกสถานะยืนยันแยก แต่จะส่งอีเมลมาก็ต่อเมื่อยืนยันกับ Facebook แล้ว
		EmailVerified: email != "",
		FirstName:     first,
		LastName:      last,
		AvatarURL:     raw.Picture.Data.URL,
	}, nil
}

// splitName ตัดชื่อเต็มเป็นชื่อกับนามสกุลด้วยช่องว่างแรก ใช้เฉพาะตอน provider ไม่ได้
// แยกฟิลด์มาให้ ชื่อที่มีหลายคำจะถูกนับเป็นนามสกุลทั้งหมด
func splitName(full string) (first, last string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "", ""
	}
	if i := strings.Index(full, " "); i > 0 {
		return full[:i], strings.TrimSpace(full[i+1:])
	}
	return full, ""
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// ---------------------------------------------------------------- state + PKCE

// Handshake คือค่าสุ่มที่ต้องถือไว้ระหว่างพาผู้ใช้ไปหน้ายินยอมจนกลับมาที่ callback
type Handshake struct {
	State     string // ส่งไปกับ URL และเก็บใน cookie — ตอนกลับมาต้องตรงกัน
	Verifier  string // เก็บฝั่งเราเท่านั้น ส่งตอนแลก token
	Challenge string // ส่งไปกับ URL เป็น SHA-256 ของ Verifier
}

// NewHandshake สร้าง state และคู่ PKCE ชุดใหม่
func NewHandshake() (Handshake, error) {
	state, err := randomString()
	if err != nil {
		return Handshake{}, err
	}
	verifier, err := randomString()
	if err != nil {
		return Handshake{}, err
	}
	sum := sha256.Sum256([]byte(verifier))
	return Handshake{
		State:     state,
		Verifier:  verifier,
		Challenge: base64.RawURLEncoding.EncodeToString(sum[:]),
	}, nil
}

func randomString() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
