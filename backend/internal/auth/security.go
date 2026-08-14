package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"backend/internal/shared/types"
)

var ErrInvalidToken = errors.New("auth: token ไม่ถูกต้องหรือหมดอายุ")

// Claims คือ payload ของ access token
type Claims struct {
	jwt.RegisteredClaims
	Role     types.Role `json:"role"`
	BranchID string     `json:"branch_id,omitempty"`
	Name     string     `json:"name"`
}

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	bcryptCost int
	dummyHash  string
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration, bcryptCost int) *Manager {
	if bcryptCost < bcrypt.MinCost || bcryptCost > bcrypt.MaxCost {
		bcryptCost = 12
	}
	m := &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		bcryptCost: bcryptCost,
	}
	// hash จริงที่ไม่มีใครรู้รหัสผ่าน ใช้เผาเวลาให้เท่ากับการ login ปกติ (ดู DummyVerify)
	if h, err := bcrypt.GenerateFromPassword([]byte(uuid.NewString()), bcryptCost); err == nil {
		m.dummyHash = string(h)
	}
	return m
}

func (m *Manager) AccessTTL() time.Duration  { return m.accessTTL }
func (m *Manager) RefreshTTL() time.Duration { return m.refreshTTL }

// ---------------------------------------------------------------- password

func (m *Manager) HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), m.bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (m *Manager) VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// DummyVerify เผาเวลาเท่ากับการตรวจรหัสผ่านจริง เรียกเมื่อไม่พบอีเมลในระบบ
// เพื่อไม่ให้ผู้โจมตีจับเวลาตอบกลับแล้วเดาได้ว่าอีเมลใดมีบัญชีอยู่
func (m *Manager) DummyVerify(plain string) {
	if m.dummyHash != "" {
		_ = bcrypt.CompareHashAndPassword([]byte(m.dummyHash), []byte(plain))
	}
}

// ---------------------------------------------------------------- access token

// Subject คือข้อมูลผู้ใช้เท่าที่จำเป็นต่อการออก token
// รับเป็น struct เล็ก ๆ แทนที่จะรับ user ทั้งก้อน เพื่อไม่ให้ package auth
// ต้องผูกกับ module ใด module หนึ่ง
type Subject struct {
	UserID   uuid.UUID
	Role     types.Role
	Name     string
	BranchID *uuid.UUID
}

func (m *Manager) IssueAccessToken(s Subject) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.accessTTL)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   s.UserID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    "wisetsuk-api",
			ID:        uuid.NewString(),
		},
		Role: s.Role,
		Name: s.Name,
	}
	if s.BranchID != nil {
		claims.BranchID = s.BranchID.String()
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (m *Manager) ParseAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer("wisetsuk-api"))

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	if !claims.Role.Valid() {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ---------------------------------------------------------------- refresh token

// NewRefreshToken สร้าง token แบบสุ่ม พร้อม hash ที่จะเก็บลง DB
// (เก็บเฉพาะ hash เพื่อไม่ให้ token ใช้ได้ทันทีหากฐานข้อมูลรั่ว)
func NewRefreshToken() (plain, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	plain = base64.RawURLEncoding.EncodeToString(buf)
	return plain, HashRefreshToken(plain), nil
}

func HashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
