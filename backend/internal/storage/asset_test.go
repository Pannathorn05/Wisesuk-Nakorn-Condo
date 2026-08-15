// เทสของการเก็บรูปไว้ในฐานข้อมูล: อัปโหลดผ่าน API ของแอดมิน แล้วอ่านกลับที่ GET /files/:assetID
package storage_test

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"backend/internal/testsupport"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// pngBytes สร้างไฟล์ PNG จริง เพราะ storage ตรวจชนิดไฟล์จากเนื้อไฟล์ ไม่ใช่จากนามสกุลหรือ header
func pngBytes(t *testing.T, c color.RGBA) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for x := range 2 {
		for y := range 2 {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("สร้าง PNG ไม่สำเร็จ: %v", err)
	}
	return buf.Bytes()
}

// uploadForm ประกอบ multipart body ที่มีไฟล์ในช่อง image พร้อมฟิลด์ข้อความเพิ่มเติม
func uploadForm(t *testing.T, content []byte, filename string, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("image", filename)
	if err != nil {
		t.Fatalf("สร้าง form file ไม่สำเร็จ: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("เขียนไฟล์ลง form ไม่สำเร็จ: %v", err)
	}
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("เขียนฟิลด์ %s ไม่สำเร็จ: %v", k, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("ปิด multipart writer ไม่สำเร็จ: %v", err)
	}
	return &body, w.FormDataContentType()
}

func upload(t *testing.T, app testsupport.App, path, token string, content []byte, filename string, fields map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	body, contentType := uploadForm(t, content, filename, fields)
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, req)
	return rec
}

// imageURLOf ดึง URL ของรูปออกจาก envelope ที่ API ตอบกลับ
func imageURLOf(t *testing.T, rec *httptest.ResponseRecorder, field string) string {
	t.Helper()

	var out struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("อ่าน response ไม่ได้: %v (body=%s)", err, rec.Body.String())
	}
	url, ok := out.Data[field].(string)
	if !ok || url == "" {
		t.Fatalf("ไม่พบ %s ใน response: %s", field, rec.Body.String())
	}
	return url
}

func get(t *testing.T, app testsupport.App, url string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, url, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, req)
	return rec
}

func TestUploadBranchImageStoredInDatabase(t *testing.T) {
	t.Parallel()

	app := testsupport.NewApp(t)
	token := testsupport.AccessToken(t, app.Pool, app.Fixture.AdminAID)
	content := pngBytes(t, color.RGBA{R: 200, G: 30, B: 30, A: 255})

	rec := upload(t, app, "/api/v1/admin/branch/images/upload", token, content, "cover.png",
		map[string]string{"caption": "อาคาร A", "sort_order": "2"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("อัปโหลดรูปสาขาได้ status %d ต้องการ 201 (body=%s)", rec.Code, rec.Body.String())
	}
	url := imageURLOf(t, rec, "image_url")

	// ไฟล์ต้องอยู่ในฐานข้อมูล ไม่ใช่บนดิสก์
	var count int
	if err := app.Pool.QueryRow(t.Context(), `SELECT count(*) FROM assets`).Scan(&count); err != nil {
		t.Fatalf("นับแถวใน assets ไม่ได้: %v", err)
	}
	if count != 1 {
		t.Fatalf("มี %d แถวในตาราง assets ต้องการ 1", count)
	}

	// อ่านกลับได้เนื้อไฟล์เดิมเป๊ะ ๆ พร้อมชนิดไฟล์ที่ถูกต้อง
	got := get(t, app, url, nil)
	if got.Code != http.StatusOK {
		t.Fatalf("ดึงรูปได้ status %d ต้องการ 200", got.Code)
	}
	if ct := got.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type = %q ต้องการ image/png", ct)
	}
	if !bytes.Equal(got.Body.Bytes(), content) {
		t.Errorf("เนื้อไฟล์ที่อ่านกลับไม่ตรงกับที่อัปโหลด (%d ไบต์ ต้องการ %d)", got.Body.Len(), len(content))
	}

	// ETag ทำให้เบราว์เซอร์ไม่ต้องโหลดซ้ำ
	etag := got.Header().Get("ETag")
	if etag == "" {
		t.Fatal("ไม่มี ETag ใน response")
	}
	cached := get(t, app, url, map[string]string{"If-None-Match": etag})
	if cached.Code != http.StatusNotModified {
		t.Errorf("ขอซ้ำพร้อม If-None-Match ได้ status %d ต้องการ 304", cached.Code)
	}
}

// อัปโหลดไฟล์เดิมซ้ำต้องใช้แถวเดิม ไม่เก็บ blob ซ้ำในฐานข้อมูล
func TestUploadSameFileTwiceReusesAsset(t *testing.T) {
	t.Parallel()

	app := testsupport.NewApp(t)
	token := testsupport.AccessToken(t, app.Pool, app.Fixture.AdminAID)
	content := pngBytes(t, color.RGBA{R: 10, G: 200, B: 90, A: 255})

	first := upload(t, app, "/api/v1/admin/branch/images/upload", token, content, "a.png", nil)
	second := upload(t, app, "/api/v1/admin/branch/images/upload", token, content, "b.png", nil)
	if first.Code != http.StatusCreated || second.Code != http.StatusCreated {
		t.Fatalf("อัปโหลดได้ status %d และ %d ต้องการ 201 ทั้งคู่", first.Code, second.Code)
	}
	if a, b := imageURLOf(t, first, "image_url"), imageURLOf(t, second, "image_url"); a != b {
		t.Errorf("ไฟล์เดียวกันได้ URL ต่างกัน: %s กับ %s", a, b)
	}

	var count int
	if err := app.Pool.QueryRow(t.Context(), `SELECT count(*) FROM assets`).Scan(&count); err != nil {
		t.Fatalf("นับแถวใน assets ไม่ได้: %v", err)
	}
	if count != 1 {
		t.Errorf("มี %d แถวในตาราง assets ต้องการ 1 เพราะเนื้อไฟล์เหมือนกัน", count)
	}
}

func TestUploadBranchCover(t *testing.T) {
	t.Parallel()

	app := testsupport.NewApp(t)
	token := testsupport.AccessToken(t, app.Pool, app.Fixture.AdminAID)

	rec := upload(t, app, "/api/v1/admin/branch/cover", token,
		pngBytes(t, color.RGBA{B: 255, A: 255}), "cover.png", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("อัปโหลดรูปปกได้ status %d ต้องการ 200 (body=%s)", rec.Code, rec.Body.String())
	}

	url := imageURLOf(t, rec, "cover_image_url")
	if got := get(t, app, url, nil); got.Code != http.StatusOK {
		t.Errorf("ดึงรูปปกได้ status %d ต้องการ 200", got.Code)
	}
}

func TestUploadRejectsNonImage(t *testing.T) {
	t.Parallel()

	app := testsupport.NewApp(t)
	token := testsupport.AccessToken(t, app.Pool, app.Fixture.AdminAID)

	// นามสกุลบอกว่าเป็นรูป แต่เนื้อไฟล์ไม่ใช่ — ต้องถูกปฏิเสธ
	rec := upload(t, app, "/api/v1/admin/branch/images/upload", token,
		[]byte("#!/bin/sh\necho hello\n"), "evil.png", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("อัปโหลดไฟล์ที่ไม่ใช่รูปได้ status %d ต้องการ 400 (body=%s)", rec.Code, rec.Body.String())
	}
}

// ฟิลด์ข้อความใน multipart ไม่ได้ผ่าน decoder ของ JSON จึงอาจเป็นไบต์ที่ไม่ใช่ UTF-8
// ต้องได้ 422 พร้อมชื่อฟิลด์ ไม่ใช่ 500 ตอน INSERT ลงฐานข้อมูล
func TestUploadRejectsNonUTF8Caption(t *testing.T) {
	t.Parallel()

	app := testsupport.NewApp(t)
	token := testsupport.AccessToken(t, app.Pool, app.Fixture.AdminAID)

	rec := upload(t, app, "/api/v1/admin/branch/images/upload", token,
		pngBytes(t, color.RGBA{R: 90, A: 255}), "a.png",
		// "ทดสอบ" ที่เข้ารหัสแบบ CP874 (ไบต์เดียวต่อตัวอักษร) ไม่ใช่ UTF-8
		map[string]string{"caption": "\xb7\xb4\xca\xcd\xba"})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ส่ง caption ที่ไม่ใช่ UTF-8 ได้ status %d ต้องการ 422 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestServeUnknownAssetReturns404(t *testing.T) {
	t.Parallel()

	app := testsupport.NewApp(t)
	rec := get(t, app, "/files/6f9619ff-8b86-d011-b42d-00c04fc964ff", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("ดึงรูปที่ไม่มีอยู่ได้ status %d ต้องการ 404", rec.Code)
	}
}

// แอดมินสาขาหนึ่งต้องอัปโหลดรูปให้ห้องของอีกสาขาไม่ได้
func TestUploadRoomImageAcrossBranchIsForbidden(t *testing.T) {
	t.Parallel()

	app := testsupport.NewApp(t)
	adminA := testsupport.AccessToken(t, app.Pool, app.Fixture.AdminAID)
	adminB := testsupport.AccessToken(t, app.Pool, app.Fixture.AdminBID)

	var roomID string
	const insertRoom = `
		INSERT INTO rooms (branch_id, room_number, floor, stay_type, price)
		VALUES ($1, '101', 1, 'monthly', 3000) RETURNING id`
	if err := app.Pool.QueryRow(t.Context(), insertRoom, app.Fixture.BranchAID).Scan(&roomID); err != nil {
		t.Fatalf("สร้างห้องทดสอบไม่สำเร็จ: %v", err)
	}

	content := pngBytes(t, color.RGBA{G: 180, A: 255})
	path := "/api/v1/admin/rooms/" + roomID + "/image"

	if rec := upload(t, app, path, adminB, content, "room.png", nil); rec.Code != http.StatusForbidden {
		t.Errorf("แอดมินสาขาอื่นอัปโหลดได้ status %d ต้องการ 403", rec.Code)
	}

	rec := upload(t, app, path, adminA, content, "room.png", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("แอดมินเจ้าของสาขาอัปโหลดได้ status %d ต้องการ 200 (body=%s)", rec.Code, rec.Body.String())
	}
	if got := get(t, app, imageURLOf(t, rec, "image_url"), nil); got.Code != http.StatusOK {
		t.Errorf("ดึงรูปห้องได้ status %d ต้องการ 200", got.Code)
	}
}
