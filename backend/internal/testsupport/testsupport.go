// Package testsupport เตรียมสภาพแวดล้อมให้เทสแต่ละตัวแยกขาดจากกันโดยสมบูรณ์
//
// หลักการ: 1 test = 1 database + 1 โฟลเดอร์อัปโหลดของตัวเอง
// เทสจึงรันขนานกันได้โดยไม่ต้องกังวลว่าจะแย่งข้อมูลกัน และไม่มีลำดับการรันที่ซ่อนอยู่
//
// ทุกอย่างถูกเก็บกวาดด้วย t.Cleanup เสมอ ไม่ต้องเรียกเองที่ปลายเทส
package testsupport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/routes"
	"backend/internal/server"
	"backend/internal/shared/types"
	"backend/internal/storage"
)

// EnvDSN คือ environment variable ที่ชี้ไปยัง PostgreSQL ที่ใช้รันเทส
// ต้องเป็นบัญชีที่สร้างและลบ database ได้ เพราะเทสแต่ละตัวสร้าง database ของตัวเอง
const EnvDSN = "TEST_DATABASE_URL"

// TestPassword คือรหัสผ่านของบัญชีทุกใบที่ Seed สร้าง
const TestPassword = "Wisetsuk!2026"

// JWTSecret ต้องถูกใช้ทั้งตอนออก token และตอนสร้าง config ของ router ในเทส
// ไม่งั้น middleware จะตรวจ token ที่เทสออกให้ไม่ผ่าน
const JWTSecret = "testsupport-jwt-secret-at-least-32-characters"

// AccessTTL และ RefreshTTL ตรงกับ AC-3
const (
	AccessTTL  = 15 * time.Minute
	RefreshTTL = 30 * 24 * time.Hour
)

// NewDatabase สร้าง database เปล่าเฉพาะของเทสตัวนี้ รัน migration ให้ครบ
// แล้วลบทิ้งอัตโนมัติเมื่อเทสจบ ไม่ว่าจะผ่านหรือไม่ผ่าน
func NewDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()

	adminDSN := requireDSN(t)
	dbName := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")

	execOnAdmin(t, adminDSN, fmt.Sprintf(`CREATE DATABASE %q`, dbName))

	ctx := context.Background()
	pool, err := database.Connect(ctx, replaceDatabase(t, adminDSN, dbName), 4, 0, 10*time.Second)
	if err != nil {
		// ลบ database ที่เพิ่งสร้างทิ้ง ไม่งั้นจะค้างอยู่เมื่อเชื่อมต่อไม่สำเร็จ
		execOnAdmin(t, adminDSN, fmt.Sprintf(`DROP DATABASE IF EXISTS %q WITH (FORCE)`, dbName))
		t.Fatalf("เชื่อมต่อ database %s ไม่ได้: %v", dbName, err)
	}

	t.Cleanup(func() {
		pool.Close()
		execOnAdmin(t, adminDSN, fmt.Sprintf(`DROP DATABASE IF EXISTS %q WITH (FORCE)`, dbName))
	})

	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("รัน migration บน %s ไม่สำเร็จ: %v", dbName, err)
	}
	return pool
}

// DatabaseName คืนชื่อ database ที่ pool กำลังต่ออยู่ ใช้ยืนยันว่าเทสแต่ละตัวแยกกันจริง
func DatabaseName(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	var name string
	if err := pool.QueryRow(context.Background(), `SELECT current_database()`).Scan(&name); err != nil {
		t.Fatalf("อ่านชื่อ database ไม่ได้: %v", err)
	}
	return name
}

// UploadDir คืนโฟลเดอร์อัปโหลดเฉพาะของเทสตัวนี้ ใช้เป็นค่า UPLOAD_DIR
// ไฟล์สลิปของแต่ละเทสจึงไม่ปนกัน และถูกลบอัตโนมัติเมื่อเทสจบ
func UploadDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// AuthManager คืน auth.Manager สำหรับเทส ใช้ bcrypt cost ต่ำสุดเพื่อให้เทสไม่ช้า
func AuthManager(t *testing.T) *auth.Manager {
	t.Helper()
	return auth.NewManager(JWTSecret, AccessTTL, RefreshTTL, bcrypt.MinCost)
}

// App คือระบบทั้งก้อนที่ประกอบเสร็จแล้วพร้อมยิงด้วย httptest
type App struct {
	Handler   http.Handler
	Pool      *pgxpool.Pool
	Fixture   Fixture
	UploadDir string
}

// NewApp ประกอบ router ตัวจริงของระบบบน database และโฟลเดอร์ไฟล์เฉพาะของเทสนี้
// ใช้ routes.New เดียวกับที่ cmd/api ใช้ เทสจึงเจอ middleware ครบทุกชั้นเหมือนผู้ใช้จริง
func NewApp(t *testing.T) App {
	t.Helper()

	pool := NewDatabase(t)
	fixture := Seed(t, pool)
	dir := UploadDir(t)

	files, err := storage.NewLocalStore(dir, "http://localhost:8080", 5<<20)
	if err != nil {
		t.Fatalf("สร้าง storage ไม่สำเร็จ: %v", err)
	}

	cfg := &config.Config{
		Env:             "test",
		Port:            "8080",
		JWTSecret:       JWTSecret,
		AccessTokenTTL:  AccessTTL,
		RefreshTokenTTL: RefreshTTL,
		UploadDir:       dir,
		PublicBaseURL:   "http://localhost:8080",
		MaxUploadBytes:  5 << 20,
		AllowedOrigins:  []string{"http://localhost:3000"},
		BcryptCost:      bcrypt.MinCost,
	}

	return App{
		Handler:   routes.New(server.New(cfg, pool, files)),
		Pool:      pool,
		Fixture:   fixture,
		UploadDir: dir,
	}
}

// Fixture คือข้อมูลตั้งต้นขั้นต่ำที่ Seed ใส่ให้
//
// มี 2 สาขาเพราะกฎเหล็กเรื่อง admin ผูกสาขาเดียวต้องมีสาขาที่สองไว้ทดสอบการข้ามสาขาเสมอ
type Fixture struct {
	BranchAID    uuid.UUID
	BranchBID    uuid.UUID
	SuperAdminID uuid.UUID
	AdminAID     uuid.UUID
	AdminBID     uuid.UUID
	MemberID     uuid.UUID
}

// Seed ใส่ข้อมูลตั้งต้นขั้นต่ำ: 2 สาขา, super admin, admin ประจำแต่ละสาขา และสมาชิก 1 คน
// ทุกบัญชีใช้รหัสผ่าน TestPassword
func Seed(t *testing.T, pool *pgxpool.Pool) Fixture {
	t.Helper()

	hash, err := AuthManager(t).HashPassword(TestPassword)
	if err != nil {
		t.Fatalf("hash รหัสผ่านไม่สำเร็จ: %v", err)
	}

	f := Fixture{}
	f.BranchAID = insertBranch(t, pool, "branch-a", "สาขาทดสอบ A")
	f.BranchBID = insertBranch(t, pool, "branch-b", "สาขาทดสอบ B")
	f.SuperAdminID = insertUser(t, pool, "super@wisetsuk.test", hash, types.RoleSuperAdmin, nil)
	f.AdminAID = insertUser(t, pool, "admin.a@wisetsuk.test", hash, types.RoleAdmin, &f.BranchAID)
	f.AdminBID = insertUser(t, pool, "admin.b@wisetsuk.test", hash, types.RoleAdmin, &f.BranchBID)
	f.MemberID = insertUser(t, pool, "member@wisetsuk.test", hash, types.RoleMember, nil)

	return f
}

// AccessToken ออก access token ให้ผู้ใช้ที่ระบุ โดยอ่าน role และสาขาจากฐานข้อมูลจริง
// เทสจึงไม่ต้องรู้ว่า claim หน้าตาเป็นอย่างไร
func AccessToken(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) string {
	t.Helper()

	var (
		role     types.Role
		branchID *uuid.UUID
		name     string
	)
	const q = `SELECT role, branch_id, first_name || ' ' || last_name FROM users WHERE id = $1`
	if err := pool.QueryRow(context.Background(), q, userID).Scan(&role, &branchID, &name); err != nil {
		t.Fatalf("ไม่พบผู้ใช้ %s: %v", userID, err)
	}

	token, _, err := AuthManager(t).IssueAccessToken(auth.Subject{
		UserID:   userID,
		Role:     role,
		Name:     name,
		BranchID: branchID,
	})
	if err != nil {
		t.Fatalf("ออก access token ไม่สำเร็จ: %v", err)
	}
	return token
}

// ---------------------------------------------------------------- ภายใน

// dsnOnce ทำให้การหา DSN เกิดครั้งเดียวต่อการรันหนึ่งครั้ง
// ไม่ใช่ทุกเทส เพราะเทสรันขนานกันและต้องไม่แย่งกันอ่านไฟล์
var (
	dsnOnce sync.Once
	dsn     string
)

// requireDSN หา DSN ของ PostgreSQL ที่ใช้รันเทส ตามลำดับนี้
//
//  1. environment variable TEST_DATABASE_URL (ใช้ใน CI และตอนรันในคอนเทนเนอร์)
//  2. ไฟล์ backend/.env ที่นักพัฒนามีอยู่แล้ว — อ่าน DATABASE_URL หรือ POSTGRES_*
//     แล้วเปลี่ยนปลายทางเป็น database ชื่อ postgres เพื่อใช้สร้าง/ลบ database ของแต่ละเทส
//
// ข้อ 2 มีไว้เพื่อให้พิมพ์ go test ./... เฉย ๆ แล้วรันได้ทันที
func requireDSN(t *testing.T) string {
	t.Helper()
	dsnOnce.Do(resolveDSN)
	if dsn == "" {
		t.Fatalf("หา PostgreSQL สำหรับรันเทสไม่เจอ — ตั้ง %s "+
			"เช่น postgres://user:pass@127.0.0.1:5432/postgres?sslmode=disable "+
			"หรือใส่ DATABASE_URL ไว้ใน backend/.env (ต้องเป็นบัญชีที่สร้าง/ลบ database ได้)", EnvDSN)
	}
	return dsn
}

func resolveDSN() {
	if v := strings.TrimSpace(os.Getenv(EnvDSN)); v != "" {
		dsn = v
		return
	}

	root, err := moduleRoot()
	if err != nil {
		return
	}
	values, err := godotenv.Read(filepath.Join(root, ".env"))
	if err != nil {
		return
	}

	if raw := strings.TrimSpace(values["DATABASE_URL"]); raw != "" {
		if u, err := url.Parse(raw); err == nil {
			u.Path = "/" + adminDatabase
			dsn = u.String()
			return
		}
	}

	user := strings.TrimSpace(values["POSTGRES_USER"])
	pass := strings.TrimSpace(values["POSTGRES_PASSWORD"])
	if user == "" || pass == "" {
		return
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, pass),
		Host:     "127.0.0.1:5432",
		Path:     "/" + adminDatabase,
		RawQuery: "sslmode=disable",
	}
	dsn = u.String()
}

// adminDatabase คือ database ที่ต่อเข้าไปเพื่อสั่ง CREATE/DROP DATABASE ของเทส
// ใช้ postgres เพราะมีอยู่ทุกเครื่องและไม่ใช่ฐานข้อมูลของงานจริง
const adminDatabase = "postgres"

// moduleRoot ไล่หาโฟลเดอร์ที่มี go.mod ขึ้นไปจาก working directory ของเทส
// เพราะ go test ตั้ง working directory เป็นโฟลเดอร์ของ package ไม่ใช่รากโมดูล
func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("testsupport: หา go.mod ไม่เจอ")
		}
		dir = parent
	}
}

func execOnAdmin(t *testing.T, adminDSN, sql string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, adminDSN)
	if err != nil {
		t.Fatalf("เชื่อมต่อ %s ไม่ได้: %v", EnvDSN, err)
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, sql); err != nil {
		t.Fatalf("รันคำสั่งจัดการ database ไม่สำเร็จ: %v", err)
	}
}

func replaceDatabase(t *testing.T, dsn, dbName string) string {
	t.Helper()
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("%s ไม่ใช่ URL ที่ถูกต้อง: %v", EnvDSN, err)
	}
	u.Path = "/" + dbName
	return u.String()
}

func insertBranch(t *testing.T, pool *pgxpool.Pool, slug, name string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	const q = `INSERT INTO branches (slug, name) VALUES ($1, $2) RETURNING id`
	if err := pool.QueryRow(context.Background(), q, slug, name).Scan(&id); err != nil {
		t.Fatalf("สร้างสาขา %s ไม่สำเร็จ: %v", slug, err)
	}
	return id
}

func insertUser(t *testing.T, pool *pgxpool.Pool, email, hash string, role types.Role, branchID *uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	const q = `
		INSERT INTO users (email, password_hash, first_name, last_name, phone, role, branch_id)
		VALUES ($1, $2, $3, $4, '0800000000', $5, $6)
		RETURNING id`
	if err := pool.QueryRow(context.Background(), q, email, hash, "ทดสอบ", string(role), role, branchID).Scan(&id); err != nil {
		t.Fatalf("สร้างผู้ใช้ %s ไม่สำเร็จ: %v", email, err)
	}
	return id
}
