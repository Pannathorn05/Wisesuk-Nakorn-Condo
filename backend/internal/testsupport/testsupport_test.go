package testsupport_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/uuid"

	"backend/internal/testsupport"
)

// เก็บชื่อ database และโฟลเดอร์อัปโหลดของทุกเทสไว้ตรวจว่าไม่ซ้ำกัน
// เทสทั้งสามตัวข้างล่างรันขนานกัน จึงต้องมี mutex คุม
var (
	mu       sync.Mutex
	dbNames  = map[string]string{}
	uploads  = map[string]string{}
	sharedID = "ค่าซ้ำที่ทุกเทสใช้เหมือนกัน"
)

// record จดชื่อที่เทสตัวนี้ได้รับ แล้วยืนยันว่ายังไม่มีเทสอื่นได้ชื่อเดียวกัน
func record(t *testing.T, store map[string]string, value, label string) {
	t.Helper()
	mu.Lock()
	defer mu.Unlock()
	if owner, taken := store[value]; taken {
		t.Fatalf("%s ซ้ำกับเทส %s: %s", label, owner, value)
	}
	store[value] = t.Name()
}

// isolate คือเนื้อเทสที่ใช้ร่วมกันทั้งสามตัว
//
// จุดสำคัญคือทุกตัวเขียนข้อมูลด้วย "ค่าเดียวกัน" ลงคอลัมน์ที่มี unique constraint
// ถ้า database ถูกใช้ร่วมกันแม้แต่คู่เดียว เทสจะพังทันทีที่ constraint ทำงาน
func isolate(t *testing.T) {
	t.Parallel()

	pool := testsupport.NewDatabase(t)
	fixture := testsupport.Seed(t, pool)
	dir := testsupport.UploadDir(t)

	record(t, dbNames, testsupport.DatabaseName(t, pool), "ชื่อ database")
	record(t, uploads, dir, "โฟลเดอร์อัปโหลด")

	ctx := context.Background()

	// ข้อมูลตั้งต้นต้องมีครบ 2 สาขาและผู้ใช้ 4 คนในทุก database
	var branches, users int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM branches`).Scan(&branches); err != nil {
		t.Fatalf("นับสาขาไม่ได้: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&users); err != nil {
		t.Fatalf("นับผู้ใช้ไม่ได้: %v", err)
	}
	if branches != 2 || users != 4 {
		t.Fatalf("ข้อมูลตั้งต้นไม่ตรง: สาขา=%d ผู้ใช้=%d ต้องการ 2 และ 4", branches, users)
	}

	// slug มี unique constraint — ทุกเทสใช้ค่าเดียวกัน ถ้าแชร์ database จะชนกันแน่นอน
	if _, err := pool.Exec(ctx,
		`INSERT INTO branches (slug, name) VALUES ($1, $2)`, "slug-ซ้ำได้เพราะคนละ database", sharedID,
	); err != nil {
		t.Fatalf("เขียนข้อมูลที่ใช้ค่าซ้ำกับเทสอื่นไม่สำเร็จ แปลว่า database ไม่ได้แยกกัน: %v", err)
	}

	// โฟลเดอร์อัปโหลดต้องว่างตอนเริ่ม แล้วเขียนไฟล์ชื่อเดียวกันได้ทุกเทส
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("อ่านโฟลเดอร์อัปโหลดไม่ได้: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("โฟลเดอร์อัปโหลดต้องว่างตอนเริ่มเทส แต่มี %d ไฟล์", len(entries))
	}
	if err := os.WriteFile(filepath.Join(dir, "slip.jpg"), []byte("ไฟล์ทดสอบ"), 0o600); err != nil {
		t.Fatalf("เขียนไฟล์ในโฟลเดอร์อัปโหลดไม่ได้: %v", err)
	}

	// ออก token ได้ครบทุกบทบาทจากข้อมูลตั้งต้น และ token ของคนละบทบาทต้องไม่เหมือนกัน
	seen := map[string]bool{}
	for label, userID := range map[string]uuid.UUID{
		"super admin":  fixture.SuperAdminID,
		"admin สาขา A": fixture.AdminAID,
		"admin สาขา B": fixture.AdminBID,
		"member":       fixture.MemberID,
	} {
		token := testsupport.AccessToken(t, pool, userID)
		if token == "" {
			t.Fatalf("ออก token ของ %s ไม่ได้", label)
		}
		if seen[token] {
			t.Fatalf("token ของ %s ซ้ำกับบทบาทอื่น", label)
		}
		seen[token] = true
	}
}

func TestIsolation_A(t *testing.T) { isolate(t) }
func TestIsolation_B(t *testing.T) { isolate(t) }
func TestIsolation_C(t *testing.T) { isolate(t) }
