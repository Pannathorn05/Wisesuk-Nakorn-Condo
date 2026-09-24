// เทส "ปิดแล้วปิดเลย" ฝั่งห้อง: ห้องที่ถูกลบ หรืออยู่ในสาขาที่ปิดใช้งาน ต้องหายจากฝั่งสาธารณะ
// ทั้งหน้ารายละเอียด (GET /rooms/{id}) และผลค้นหา (GET /rooms/search)
// แต่ฝั่ง admin ยังต้องเห็นและแก้ได้ตามปกติ
package room_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/shared/types"
)

func deactivateRoom(t *testing.T, e env, id types.RoomID) {
	t.Helper()
	if _, err := e.app.Pool.Exec(context.Background(),
		`UPDATE rooms SET is_active = FALSE WHERE id = $1`, id); err != nil {
		t.Fatalf("ปิดห้องไม่สำเร็จ: %v", err)
	}
}

func deactivateBranch(t *testing.T, e env, id types.BranchID) {
	t.Helper()
	if _, err := e.app.Pool.Exec(context.Background(),
		`UPDATE branches SET is_active = FALSE WHERE id = $1`, id); err != nil {
		t.Fatalf("ปิดสาขาไม่สำเร็จ: %v", err)
	}
}

func assertNotFound(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d ต้องเป็น 404 (body=%s)", rec.Code, rec.Body.String())
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("อ่าน response ไม่ได้: %v (body=%s)", err, rec.Body.String())
	}
	if body.Error.Code != "not_found" {
		t.Errorf("error.code = %q ต้องเป็น not_found", body.Error.Code)
	}
}

func searchRoomNumbers(t *testing.T, e env, branchID types.BranchID) map[string]bool {
	t.Helper()
	rec := e.do(t, http.MethodGet, "/api/v1/rooms/search?branch_id="+branchID.String(), "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d ต้องเป็น 200 (body=%s)", rec.Code, rec.Body.String())
	}
	var body struct {
		Data []struct {
			RoomNumber string `json:"room_number"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("อ่าน response ไม่ได้: %v", err)
	}
	got := map[string]bool{}
	for _, rm := range body.Data {
		got[rm.RoomNumber] = true
	}
	return got
}

func TestGetDeletedRoomReturns404(t *testing.T) {
	e := newEnv(t)
	roomID := insertRoom(t, e, "501")
	deactivateRoom(t, e, roomID)

	assertNotFound(t, e.do(t, http.MethodGet, "/api/v1/rooms/"+roomID.String(), "", ""))
}

// ตัวห้องยังเปิดอยู่ แต่สาขาปิดไปแล้ว = ปิดตามสาขา
func TestGetRoomInClosedBranchReturns404(t *testing.T) {
	e := newEnv(t)
	roomID := insertRoom(t, e, "502")
	deactivateBranch(t, e, e.app.Fixture.BranchAID)

	assertNotFound(t, e.do(t, http.MethodGet, "/api/v1/rooms/"+roomID.String(), "", ""))
}

func TestSearchExcludesRoomsInClosedBranch(t *testing.T) {
	e := newEnv(t)
	insertRoom(t, e, "503")

	// ยืนยันก่อนว่าห้องโผล่จริงตอนสาขาเปิด เพื่อไม่ให้เทสผ่านเพราะเหตุอื่น
	if got := searchRoomNumbers(t, e, e.app.Fixture.BranchAID); !got["503"] {
		t.Fatalf("ห้อง 503 ต้องโผล่ในผลค้นหาตอนสาขายังเปิด แต่ได้ %v", got)
	}

	deactivateBranch(t, e, e.app.Fixture.BranchAID)
	if got := searchRoomNumbers(t, e, e.app.Fixture.BranchAID); got["503"] {
		t.Errorf("ห้อง 503 อยู่ในสาขาที่ปิดแล้ว ต้องไม่โผล่ในผลค้นหา")
	}
}

// ฝั่ง admin ต้องไม่ถูกกรองไปด้วย — แอดมินสาขาที่ปิดยังต้องเห็นและแก้ห้องของตัวเองได้
func TestAdminStillManagesRoomsInClosedBranch(t *testing.T) {
	e := newEnv(t)
	roomID := insertRoom(t, e, "504")
	deactivateBranch(t, e, e.app.Fixture.BranchAID)
	token := e.adminToken(t)

	list := e.do(t, http.MethodGet, "/api/v1/admin/rooms", "", token)
	if list.Code != http.StatusOK {
		t.Fatalf("GET /admin/rooms status = %d ต้องเป็น 200 (body=%s)", list.Code, list.Body.String())
	}
	var body struct {
		Data []struct {
			RoomNumber string `json:"room_number"`
		} `json:"data"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &body); err != nil {
		t.Fatalf("อ่าน response ไม่ได้: %v", err)
	}
	found := false
	for _, rm := range body.Data {
		if rm.RoomNumber == "504" {
			found = true
		}
	}
	if !found {
		t.Errorf("แอดมินต้องยังเห็นห้อง 504 ใน /admin/rooms")
	}

	update := `{"room_number":"504","floor":5,"stay_type":"monthly","price":3600}`
	rec := e.do(t, http.MethodPut, "/api/v1/admin/rooms/"+roomID.String(), update, token)
	if rec.Code != http.StatusOK {
		t.Errorf("PUT /admin/rooms status = %d ต้องเป็น 200 (body=%s)", rec.Code, rec.Body.String())
	}
}
