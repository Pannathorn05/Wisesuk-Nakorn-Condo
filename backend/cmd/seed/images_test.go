// เทสการตั้งรูปเดโมจากโฟลเดอร์ uploads/ ตอนรัน seed
//
// สัญญาที่ต้องยืนยัน: image_url ชี้ไฟล์ใน uploads/ ที่ route /uploads/* เสิร์ฟได้จริง, รัน seed ซ้ำ
// ได้ผลเท่าเดิม และ seed ต้องไม่ไปลบหรือทับรูปที่แอดมินอัปโหลดเองผ่าน API
package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/internal/shared/types"
	"backend/internal/testsupport"
)

const testBaseURL = "http://localhost:8080"

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// writeFile สร้างไฟล์พร้อมโฟลเดอร์แม่ที่ยังไม่มี
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("สร้างโฟลเดอร์ไม่สำเร็จ: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("เขียนไฟล์ไม่สำเร็จ: %v", err)
	}
}

// ---------------------------------------------------------------- อ่านโฟลเดอร์

func TestScanBranchImages(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "demo", "b.PNG"), "x")
	writeFile(t, filepath.Join(dir, "demo", "notes.txt"), "ไม่ใช่รูป")
	writeFile(t, filepath.Join(dir, "demo", "days", "a.jpg"), "x")
	writeFile(t, filepath.Join(dir, "demo", "days", "caption.txt"), "ห้องพักรายวัน\n")
	writeFile(t, filepath.Join(dir, "demo", "month", "c (1).webp"), "x")

	got, err := scanBranchImages(dir, "demo")
	if err != nil {
		t.Fatalf("scanBranchImages: %v", err)
	}

	// เรียงตาม path เสมอ ลำดับนี้กลายเป็น sort_order ในแกลเลอรี
	want := []demoImage{
		{relPath: "b.PNG", caption: ""},
		{relPath: "days/a.jpg", caption: "ห้องพักรายวัน"},
		{relPath: "month/c (1).webp", caption: ""},
	}
	if len(got) != len(want) {
		t.Fatalf("ได้ %d รูป ต้องการ %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("รูปที่ %d = %+v ต้องการ %+v", i, got[i], want[i])
		}
	}
}

// สาขาที่ยังไม่มีโฟลเดอร์รูปไม่ใช่ error — seed ต้องทำงานต่อได้
func TestScanBranchImagesWithoutFolder(t *testing.T) {
	got, err := scanBranchImages(t.TempDir(), "no-such-branch")
	if err != nil {
		t.Fatalf("scanBranchImages: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ได้ %d รูป ต้องการ 0", len(got))
	}
}

func TestUploadURLEscapesEachSegment(t *testing.T) {
	got := uploadURL(testBaseURL, "prachauthit-45", "month/m-45-3 (1).jpg")
	want := testBaseURL + "/uploads/prachauthit-45/month/m-45-3%20%281%29.jpg"
	if got != want {
		t.Errorf("uploadURL = %q ต้องการ %q", got, want)
	}
}

// URL ที่ seed เขียนลงฐานข้อมูลต้องเปิดได้จริงผ่าน route /uploads/* ตัวจริงของระบบ
// ไม่งั้นหน้าเว็บจะได้ลิงก์รูปเสียทั้งที่ข้อมูลใน DB ดูถูกต้อง
func TestUploadURLIsServedByUploadsRoute(t *testing.T) {
	app := testsupport.NewApp(t)
	writeFile(t, filepath.Join(app.UploadDir, "demo", "month", "m-45-3 (1).jpg"), "jpeg-bytes")

	full := uploadURL(testBaseURL, "demo", "month/m-45-3 (1).jpg")
	req := httptest.NewRequest(http.MethodGet, strings.TrimPrefix(full, testBaseURL), nil)
	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d ต้องการ 200", full, rec.Code)
	}
	if rec.Body.String() != "jpeg-bytes" {
		t.Errorf("เนื้อไฟล์ = %q", rec.Body.String())
	}
}

// ---------------------------------------------------------------- แกลเลอรีสาขา

type galleryRow struct {
	url       string
	caption   string
	sortOrder int
}

func galleryOf(t *testing.T, pool *pgxpool.Pool, branchID types.BranchID) []galleryRow {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT image_url, caption, sort_order FROM branch_images
		 WHERE branch_id = $1 ORDER BY sort_order, id`, branchID)
	if err != nil {
		t.Fatalf("อ่านแกลเลอรีไม่ได้: %v", err)
	}
	defer rows.Close()

	var out []galleryRow
	for rows.Next() {
		var r galleryRow
		if err := rows.Scan(&r.url, &r.caption, &r.sortOrder); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, r)
	}
	return out
}

func TestSeedBranchGallery(t *testing.T) {
	pool := testsupport.NewDatabase(t)
	branchID := testsupport.Seed(t, pool).BranchAID
	ctx := context.Background()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "demo", "a.jpg"), "x")
	writeFile(t, filepath.Join(dir, "demo", "days", "b.jpg"), "x")
	writeFile(t, filepath.Join(dir, "demo", "days", "caption.txt"), "ห้องพักรายวัน")
	imgs := demoImages{dir: dir, baseURL: testBaseURL}

	// รูปที่แอดมินอัปโหลดเองผ่าน API — seed ห้ามแตะ
	const adminURL = testBaseURL + "/files/ast-001"
	if _, err := pool.Exec(ctx,
		`INSERT INTO branch_images (branch_id, image_url, sort_order) VALUES ($1, $2, 99)`,
		branchID, adminURL); err != nil {
		t.Fatalf("ใส่รูปของแอดมินไม่สำเร็จ: %v", err)
	}

	// รันสองรอบ ผลต้องเท่ากับรอบเดียว
	for range 2 {
		if err := imgs.seedBranch(ctx, pool, branchID, "demo"); err != nil {
			t.Fatalf("seedBranch: %v", err)
		}
	}

	want := []galleryRow{
		{url: testBaseURL + "/uploads/demo/a.jpg", caption: "", sortOrder: 0},
		{url: testBaseURL + "/uploads/demo/days/b.jpg", caption: "ห้องพักรายวัน", sortOrder: 1},
		{url: adminURL, caption: "", sortOrder: 99},
	}
	got := galleryOf(t, pool, branchID)
	if len(got) != len(want) {
		t.Fatalf("แกลเลอรีมี %d รูป ต้องการ %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("รูปที่ %d = %+v ต้องการ %+v", i, got[i], want[i])
		}
	}
}

// ไฟล์ที่ถูกลบออกจาก repo ต้องหายจากแกลเลอรีเมื่อรัน seed รอบถัดไป
func TestSeedBranchGalleryDropsRemovedFiles(t *testing.T) {
	pool := testsupport.NewDatabase(t)
	branchID := testsupport.Seed(t, pool).BranchAID
	ctx := context.Background()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "demo", "a.jpg"), "x")
	writeFile(t, filepath.Join(dir, "demo", "b.jpg"), "x")
	imgs := demoImages{dir: dir, baseURL: testBaseURL}

	if err := imgs.seedBranch(ctx, pool, branchID, "demo"); err != nil {
		t.Fatalf("seedBranch รอบแรก: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, "demo", "b.jpg")); err != nil {
		t.Fatal(err)
	}
	if err := imgs.seedBranch(ctx, pool, branchID, "demo"); err != nil {
		t.Fatalf("seedBranch รอบสอง: %v", err)
	}

	got := galleryOf(t, pool, branchID)
	if len(got) != 1 || got[0].url != testBaseURL+"/uploads/demo/a.jpg" {
		t.Errorf("แกลเลอรี = %+v ต้องการเหลือแค่ a.jpg", got)
	}
}

// ---------------------------------------------------------------- รูปปกห้อง

func insertRoom(t *testing.T, pool *pgxpool.Pool, branchID types.BranchID, number, imageURL string) types.RoomID {
	t.Helper()
	var id types.RoomID
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO rooms (branch_id, room_number, stay_type, price, image_url)
		 VALUES ($1, $2, 'monthly', 3000, $3) RETURNING id`,
		branchID, number, imageURL).Scan(&id); err != nil {
		t.Fatalf("สร้างห้องไม่สำเร็จ: %v", err)
	}
	return id
}

func coverOf(t *testing.T, pool *pgxpool.Pool, roomID types.RoomID) string {
	t.Helper()
	var url string
	if err := pool.QueryRow(context.Background(),
		`SELECT image_url FROM rooms WHERE id = $1`, roomID).Scan(&url); err != nil {
		t.Fatalf("อ่านรูปปกไม่ได้: %v", err)
	}
	return url
}

func TestSetRoomCover(t *testing.T) {
	pool := testsupport.NewDatabase(t)
	branchID := testsupport.Seed(t, pool).BranchAID
	ctx := context.Background()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "demo", "month", "m-1 (1).jpg"), "x")
	imgs := demoImages{dir: dir, baseURL: testBaseURL}
	seedURL := testBaseURL + "/uploads/demo/month/m-1%20%281%29.jpg"

	cases := []struct {
		name   string
		before string
		want   string
	}{
		{"ห้องที่ยังไม่มีรูปปก ได้รูปเดโม", "", seedURL},
		{"รูปเดโมชุดเก่า ถูกแทนด้วยชุดใหม่", testBaseURL + "/uploads/demo/old.jpg", seedURL},
		{"รูปที่แอดมินอัปเอง ไม่ถูกทับ", testBaseURL + "/files/ast-007", testBaseURL + "/files/ast-007"},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			roomID := insertRoom(t, pool, branchID, "R"+string(rune('1'+i)), tc.before)
			if err := imgs.setRoomCover(ctx, pool, roomID, "demo", "month/m-1 (1).jpg"); err != nil {
				t.Fatalf("setRoomCover: %v", err)
			}
			if got := coverOf(t, pool, roomID); got != tc.want {
				t.Errorf("image_url = %q ต้องการ %q", got, tc.want)
			}
		})
	}
}

// ไฟล์ที่ mapping อ้างถึงแต่ไม่มีอยู่จริงคือ repo ไม่ครบ ต้องบอกให้รู้ ไม่ใช่เขียน URL เสียลงไปเงียบ ๆ
func TestSetRoomCoverMissingFile(t *testing.T) {
	pool := testsupport.NewDatabase(t)
	branchID := testsupport.Seed(t, pool).BranchAID
	roomID := insertRoom(t, pool, branchID, "R1", "")

	imgs := demoImages{dir: t.TempDir(), baseURL: testBaseURL}
	err := imgs.setRoomCover(context.Background(), pool, roomID, "demo", "nope.jpg")
	if err == nil {
		t.Fatal("ต้องได้ error เมื่อไม่พบไฟล์รูปปก")
	}
	if got := coverOf(t, pool, roomID); got != "" {
		t.Errorf("image_url = %q ต้องยังว่างอยู่", got)
	}
}
