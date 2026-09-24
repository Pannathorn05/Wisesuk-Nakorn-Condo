// เทสแกลเลอรีรูปห้อง (room_images) ยิงผ่าน router ตัวจริงเหมือนที่ผู้ใช้เรียก
//
// สัญญาที่ต้องยืนยัน: แกลเลอรีเป็นคนละช่องกับรูปปก (Room.image_url) การเพิ่ม/ลบรูปแกลเลอรี
// ต้องไม่ไปแตะรูปปก, รูปเรียงตาม sort_order, ผลค้นหาไม่แบกแกลเลอรีไปด้วย และแอดมินสาขาหนึ่ง
// ต้องลบรูปของอีกสาขาไม่ได้แม้จะรู้ id
package room_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"backend/internal/shared/types"
	"backend/internal/testsupport"
)

func insertRoomIn(t *testing.T, e env, branchID types.BranchID, roomNumber string) types.RoomID {
	t.Helper()
	var id types.RoomID
	err := e.app.Pool.QueryRow(context.Background(),
		`INSERT INTO rooms (branch_id, room_number, floor, stay_type, price)
		 VALUES ($1, $2, 2, 'monthly', 3400) RETURNING id`,
		branchID, roomNumber).Scan(&id)
	if err != nil {
		t.Fatalf("สร้างห้อง %s ไม่สำเร็จ: %v", roomNumber, err)
	}
	return id
}

// addImage เรียก endpoint เพิ่มรูปแบบ JSON แล้วคืน id ของรูปที่เพิ่งสร้าง
func addImage(t *testing.T, e env, roomID types.RoomID, url string, sortOrder int, token string) string {
	t.Helper()

	body, err := json.Marshal(map[string]any{"image_url": url, "sort_order": sortOrder})
	if err != nil {
		t.Fatalf("สร้าง payload ไม่สำเร็จ: %v", err)
	}

	rec := e.do(t, http.MethodPost, "/api/v1/admin/rooms/"+roomID.String()+"/images", string(body), token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d ต้องเป็น 201 (body=%s)", rec.Code, rec.Body.String())
	}

	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("อ่าน response ไม่ได้: %v", err)
	}
	return created.Data.ID
}

// roomGallery แกะแกลเลอรีออกจาก GET /rooms/{id} โดยคงลำดับที่ API ส่งมา
func roomGallery(t *testing.T, e env, roomID types.RoomID) []string {
	t.Helper()

	rec := e.do(t, http.MethodGet, "/api/v1/rooms/"+roomID.String(), "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d ต้องเป็น 200 (body=%s)", rec.Code, rec.Body.String())
	}

	var body struct {
		Data struct {
			Images []struct {
				ImageURL string `json:"image_url"`
			} `json:"images"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("อ่าน response ไม่ได้: %v", err)
	}

	urls := make([]string, 0, len(body.Data.Images))
	for _, img := range body.Data.Images {
		urls = append(urls, img.ImageURL)
	}
	return urls
}

// coverOf อ่านรูปปกของห้องจาก API เพื่อพิสูจน์ว่าแกลเลอรีไม่ไปแตะช่องนี้
func coverOf(t *testing.T, e env, roomID types.RoomID) string {
	t.Helper()

	rec := e.do(t, http.MethodGet, "/api/v1/rooms/"+roomID.String(), "", "")
	var body struct {
		Data struct {
			ImageURL string `json:"image_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("อ่าน response ไม่ได้: %v", err)
	}
	return body.Data.ImageURL
}

// ใส่ sort_order สลับกับลำดับที่เพิ่ม เพื่อพิสูจน์ว่า API เรียงให้จริง ไม่ใช่บังเอิญตรง
func TestGalleryReturnsImagesInSortOrder(t *testing.T) {
	e := newEnv(t)
	token := e.adminToken(t)
	roomID := insertRoom(t, e, "501")

	addImage(t, e, roomID, "http://localhost:8080/files/ast-003", 5, token)
	addImage(t, e, roomID, "http://localhost:8080/files/ast-001", 0, token)
	addImage(t, e, roomID, "http://localhost:8080/files/ast-002", 2, token)

	want := []string{
		"http://localhost:8080/files/ast-001",
		"http://localhost:8080/files/ast-002",
		"http://localhost:8080/files/ast-003",
	}
	if got := roomGallery(t, e, roomID); !equalStrings(got, want) {
		t.Errorf("แกลเลอรี = %v ต้องเป็น %v", got, want)
	}
}

func TestRoomWithoutGalleryReturnsNone(t *testing.T) {
	e := newEnv(t)
	roomID := insertRoom(t, e, "502")

	if got := roomGallery(t, e, roomID); len(got) != 0 {
		t.Errorf("แกลเลอรี = %v ต้องว่าง", got)
	}
}

// แกลเลอรีกับรูปปกเป็นคนละช่อง เพิ่มรูปเข้าแกลเลอรีต้องไม่ไปทับรูปปกที่ตั้งไว้แล้ว
func TestAddGalleryImageDoesNotTouchCover(t *testing.T) {
	e := newEnv(t)
	token := e.adminToken(t)
	roomID := insertRoom(t, e, "503")

	const cover = "http://localhost:8080/files/ast-009"
	if _, err := e.app.Pool.Exec(context.Background(),
		`UPDATE rooms SET image_url = $2 WHERE id = $1`, roomID, cover); err != nil {
		t.Fatalf("ตั้งรูปปกไม่สำเร็จ: %v", err)
	}

	addImage(t, e, roomID, "http://localhost:8080/files/ast-001", 0, token)

	if got := coverOf(t, e, roomID); got != cover {
		t.Errorf("รูปปก = %q ต้องยังเป็น %q", got, cover)
	}
}

func TestDeleteGalleryImage(t *testing.T) {
	e := newEnv(t)
	token := e.adminToken(t)
	roomID := insertRoom(t, e, "504")

	keep := "http://localhost:8080/files/ast-001"
	addImage(t, e, roomID, keep, 0, token)
	drop := addImage(t, e, roomID, "http://localhost:8080/files/ast-002", 1, token)

	rec := e.do(t, http.MethodDelete,
		"/api/v1/admin/rooms/"+roomID.String()+"/images/"+drop, "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d ต้องเป็น 204 (body=%s)", rec.Code, rec.Body.String())
	}

	if got := roomGallery(t, e, roomID); !equalStrings(got, []string{keep}) {
		t.Errorf("แกลเลอรี = %v ต้องเหลือแค่ %v", got, []string{keep})
	}
}

// กฎเหล็ก: แอดมินเห็นและแก้ได้เฉพาะสาขาตัวเอง แม้จะรู้ id ของรูปอีกสาขาก็ต้องลบไม่ได้
func TestAdminCannotDeleteImageOfAnotherBranch(t *testing.T) {
	e := newEnv(t)

	tokenA := e.adminToken(t)
	roomA := insertRoom(t, e, "505")
	imageA := addImage(t, e, roomA, "http://localhost:8080/files/ast-001", 0, tokenA)

	tokenB := testsupport.AccessToken(t, e.app.Pool, e.app.Fixture.AdminBID)
	roomB := insertRoomIn(t, e, e.app.Fixture.BranchBID, "506")

	// แอดมิน B อ้างห้องของตัวเองแต่ส่ง id รูปของสาขา A มา ต้องไม่โดนลบ
	rec := e.do(t, http.MethodDelete,
		"/api/v1/admin/rooms/"+roomB.String()+"/images/"+imageA, "", tokenB)
	if rec.Code == http.StatusNoContent {
		t.Fatalf("ลบสำเร็จ ทั้งที่เป็นรูปของอีกสาขา")
	}

	if got := roomGallery(t, e, roomA); len(got) != 1 {
		t.Errorf("รูปของสาขา A เหลือ %d ใบ ต้องยังอยู่ครบ 1 ใบ", len(got))
	}
}

// ผลค้นหาโชว์แค่รูปปก แกลเลอรีเป็นของหน้ารายละเอียด ไม่ต้องแบกไปทั้งหน้า
func TestSearchDoesNotReturnGallery(t *testing.T) {
	e := newEnv(t)
	token := e.adminToken(t)
	roomID := insertRoom(t, e, "507")
	addImage(t, e, roomID, "http://localhost:8080/files/ast-001", 0, token)

	rec := e.do(t, http.MethodGet,
		"/api/v1/rooms/search?branch_id="+e.app.Fixture.BranchAID.String(), "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d ต้องเป็น 200 (body=%s)", rec.Code, rec.Body.String())
	}

	var body struct {
		Data []struct {
			RoomNumber string `json:"room_number"`
			Images     []any  `json:"images"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("อ่าน response ไม่ได้: %v", err)
	}
	if len(body.Data) == 0 {
		t.Fatal("ผลค้นหาว่าง ต้องเจอห้องที่เพิ่งสร้าง")
	}
	for _, rm := range body.Data {
		if len(rm.Images) != 0 {
			t.Errorf("ห้อง %s ในผลค้นหามีแกลเลอรีติดมาด้วย ต้องไม่มี", rm.RoomNumber)
		}
	}
}
