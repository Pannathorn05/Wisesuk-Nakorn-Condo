// เทสสิ่งอำนวยความสะดวกรายห้อง ยิงผ่าน router ตัวจริงเหมือนที่ผู้ใช้เรียก
//
// สัญญาที่ต้องยืนยันมีสองข้อ: amenities โผล่เฉพาะหน้ารายละเอียดห้อง (GET /rooms/{id})
// ไม่ติดไปกับผลค้นหา และแอดมินแก้ชุดสิ่งอำนวยความสะดวกได้ผ่าน POST/PUT /admin/rooms
package room_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/shared/types"
	"backend/internal/testsupport"
)

type env struct {
	app testsupport.App
}

func newEnv(t *testing.T) env {
	t.Helper()
	return env{app: testsupport.NewApp(t)}
}

func (e env) do(t *testing.T, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	e.app.Handler.ServeHTTP(rec, req)
	return rec
}

func (e env) adminToken(t *testing.T) string {
	t.Helper()
	return testsupport.AccessToken(t, e.app.Pool, e.app.Fixture.AdminAID)
}

// insertAmenity ใส่สิ่งอำนวยความสะดวกลง master list — fixture ตั้งต้นไม่มีมาให้
func insertAmenity(t *testing.T, e env, code, name string, sortOrder int) types.AmenityID {
	t.Helper()
	var id types.AmenityID
	err := e.app.Pool.QueryRow(context.Background(),
		`INSERT INTO amenities (code, name, icon, sort_order) VALUES ($1, $2, '', $3) RETURNING id`,
		code, name, sortOrder).Scan(&id)
	if err != nil {
		t.Fatalf("สร้างสิ่งอำนวยความสะดวก %s ไม่สำเร็จ: %v", code, err)
	}
	return id
}

func insertRoom(t *testing.T, e env, roomNumber string) types.RoomID {
	t.Helper()
	var id types.RoomID
	err := e.app.Pool.QueryRow(context.Background(),
		`INSERT INTO rooms (branch_id, room_number, floor, stay_type, price)
		 VALUES ($1, $2, 2, 'monthly', 3400) RETURNING id`,
		e.app.Fixture.BranchAID, roomNumber).Scan(&id)
	if err != nil {
		t.Fatalf("สร้างห้อง %s ไม่สำเร็จ: %v", roomNumber, err)
	}
	return id
}

func linkAmenities(t *testing.T, e env, roomID types.RoomID, ids ...types.AmenityID) {
	t.Helper()
	for _, id := range ids {
		if _, err := e.app.Pool.Exec(context.Background(),
			`INSERT INTO room_amenities (room_id, amenity_id) VALUES ($1, $2)`, roomID, id); err != nil {
			t.Fatalf("ผูกสิ่งอำนวยความสะดวก %v กับห้องไม่สำเร็จ: %v", id, err)
		}
	}
}

// roomAmenityNames แกะชื่อสิ่งอำนวยความสะดวกออกจาก response ของห้อง โดยคงลำดับไว้ตามที่ API ส่งมา
func roomAmenityNames(t *testing.T, rec *httptest.ResponseRecorder) []string {
	t.Helper()

	var body struct {
		Data struct {
			Amenities []struct {
				ID   string `json:"id"`
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"amenities"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("อ่าน response ไม่ได้: %v (body=%s)", err, rec.Body.String())
	}

	names := make([]string, 0, len(body.Data.Amenities))
	for _, a := range body.Data.Amenities {
		names = append(names, a.Name)
	}
	return names
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestGetRoomReturnsAmenitiesInSortOrder(t *testing.T) {
	e := newEnv(t)

	// ใส่ sort_order สลับกับลำดับที่ผูก เพื่อพิสูจน์ว่า API เรียงให้จริง ไม่ใช่บังเอิญตรง
	aircon := insertAmenity(t, e, "aircon", "เครื่องปรับอากาศ", 0)
	keycard := insertAmenity(t, e, "keycard", "คีย์การ์ด", 5)
	bathroom := insertAmenity(t, e, "private-bathroom", "ห้องน้ำในตัว", 2)

	roomID := insertRoom(t, e, "302")
	linkAmenities(t, e, roomID, keycard, aircon, bathroom)

	rec := e.do(t, http.MethodGet, "/api/v1/rooms/"+roomID.String(), "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d ต้องเป็น 200 (body=%s)", rec.Code, rec.Body.String())
	}

	want := []string{"เครื่องปรับอากาศ", "ห้องน้ำในตัว", "คีย์การ์ด"}
	if got := roomAmenityNames(t, rec); !equalStrings(got, want) {
		t.Errorf("amenities = %v ต้องเป็น %v", got, want)
	}
}

func TestGetRoomWithoutAmenitiesReturnsNone(t *testing.T) {
	e := newEnv(t)
	roomID := insertRoom(t, e, "303")

	rec := e.do(t, http.MethodGet, "/api/v1/rooms/"+roomID.String(), "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d ต้องเป็น 200 (body=%s)", rec.Code, rec.Body.String())
	}
	if got := roomAmenityNames(t, rec); len(got) != 0 {
		t.Errorf("amenities = %v ต้องว่าง", got)
	}
}

// ผลค้นหาคืนทีละหลายห้อง จึงไม่แบก amenities ไปด้วยตามที่ตกลงไว้ใน openapi
func TestSearchDoesNotReturnAmenities(t *testing.T) {
	e := newEnv(t)

	aircon := insertAmenity(t, e, "aircon", "เครื่องปรับอากาศ", 0)
	roomID := insertRoom(t, e, "304")
	linkAmenities(t, e, roomID, aircon)

	rec := e.do(t, http.MethodGet, "/api/v1/rooms/search?branch_id="+e.app.Fixture.BranchAID.String(), "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d ต้องเป็น 200 (body=%s)", rec.Code, rec.Body.String())
	}

	var body struct {
		Data []struct {
			RoomNumber string `json:"room_number"`
			Amenities  []any  `json:"amenities"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("อ่าน response ไม่ได้: %v", err)
	}
	if len(body.Data) == 0 {
		t.Fatal("ผลค้นหาว่าง ต้องเจอห้องที่เพิ่งสร้าง")
	}
	for _, rm := range body.Data {
		if len(rm.Amenities) != 0 {
			t.Errorf("ห้อง %s ในผลค้นหามี amenities ติดมาด้วย ต้องไม่มี", rm.RoomNumber)
		}
	}
}

func TestAdminCreatesRoomWithAmenities(t *testing.T) {
	e := newEnv(t)

	aircon := insertAmenity(t, e, "aircon", "เครื่องปรับอากาศ", 0)
	keycard := insertAmenity(t, e, "keycard", "คีย์การ์ด", 5)

	body := `{"room_number":"401","floor":4,"stay_type":"monthly","price":3400,
	          "amenity_ids":["` + aircon.String() + `","` + keycard.String() + `"]}`
	rec := e.do(t, http.MethodPost, "/api/v1/admin/rooms", body, e.adminToken(t))
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

	got := e.do(t, http.MethodGet, "/api/v1/rooms/"+created.Data.ID, "", "")
	want := []string{"เครื่องปรับอากาศ", "คีย์การ์ด"}
	if names := roomAmenityNames(t, got); !equalStrings(names, want) {
		t.Errorf("amenities = %v ต้องเป็น %v", names, want)
	}
}

func TestAdminUpdateReplacesAmenities(t *testing.T) {
	e := newEnv(t)

	aircon := insertAmenity(t, e, "aircon", "เครื่องปรับอากาศ", 0)
	keycard := insertAmenity(t, e, "keycard", "คีย์การ์ด", 5)
	bathroom := insertAmenity(t, e, "private-bathroom", "ห้องน้ำในตัว", 2)

	roomID := insertRoom(t, e, "402")
	linkAmenities(t, e, roomID, aircon, keycard)

	body := `{"room_number":"402","floor":4,"stay_type":"monthly","price":3400,
	          "amenity_ids":["` + bathroom.String() + `"]}`
	rec := e.do(t, http.MethodPut, "/api/v1/admin/rooms/"+roomID.String(), body, e.adminToken(t))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d ต้องเป็น 200 (body=%s)", rec.Code, rec.Body.String())
	}

	got := e.do(t, http.MethodGet, "/api/v1/rooms/"+roomID.String(), "", "")
	want := []string{"ห้องน้ำในตัว"}
	if names := roomAmenityNames(t, got); !equalStrings(names, want) {
		t.Errorf("amenities = %v ต้องถูกแทนที่เป็น %v", names, want)
	}
}

// ไม่ส่ง amenity_ids มา = คำขอไม่ได้พูดถึงเรื่องนี้ ของเดิมต้องอยู่ครบ
// ไม่ใช่ถูกล้างทิ้งเพราะฟิลด์หายไปจาก payload
func TestAdminUpdateWithoutAmenityIDsKeepsThem(t *testing.T) {
	e := newEnv(t)

	aircon := insertAmenity(t, e, "aircon", "เครื่องปรับอากาศ", 0)
	roomID := insertRoom(t, e, "403")
	linkAmenities(t, e, roomID, aircon)

	body := `{"room_number":"403","floor":4,"stay_type":"monthly","price":3400}`
	rec := e.do(t, http.MethodPut, "/api/v1/admin/rooms/"+roomID.String(), body, e.adminToken(t))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d ต้องเป็น 200 (body=%s)", rec.Code, rec.Body.String())
	}

	got := e.do(t, http.MethodGet, "/api/v1/rooms/"+roomID.String(), "", "")
	want := []string{"เครื่องปรับอากาศ"}
	if names := roomAmenityNames(t, got); !equalStrings(names, want) {
		t.Errorf("amenities = %v ต้องคงเดิมเป็น %v", names, want)
	}
}

// ส่ง array ว่างมา = ตั้งใจล้าง ต่างจากไม่ส่งฟิลด์มาเลย
func TestAdminUpdateWithEmptyAmenityIDsClearsThem(t *testing.T) {
	e := newEnv(t)

	aircon := insertAmenity(t, e, "aircon", "เครื่องปรับอากาศ", 0)
	roomID := insertRoom(t, e, "404")
	linkAmenities(t, e, roomID, aircon)

	body := `{"room_number":"404","floor":4,"stay_type":"monthly","price":3400,"amenity_ids":[]}`
	rec := e.do(t, http.MethodPut, "/api/v1/admin/rooms/"+roomID.String(), body, e.adminToken(t))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d ต้องเป็น 200 (body=%s)", rec.Code, rec.Body.String())
	}

	got := e.do(t, http.MethodGet, "/api/v1/rooms/"+roomID.String(), "", "")
	if names := roomAmenityNames(t, got); len(names) != 0 {
		t.Errorf("amenities = %v ต้องถูกล้างจนว่าง", names)
	}
}

// amenity ที่ไม่มีอยู่จริงคือข้อมูลจากผู้ใช้ที่ผิด ต้องเป็น 422 พร้อมบอกฟิลด์ ไม่ใช่ 500
func TestAdminCreateRejectsUnknownAmenity(t *testing.T) {
	e := newEnv(t)

	body := `{"room_number":"405","floor":4,"stay_type":"monthly","price":3400,"amenity_ids":["amt-999"]}`
	rec := e.do(t, http.MethodPost, "/api/v1/admin/rooms", body, e.adminToken(t))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d ต้องเป็น 422 (body=%s)", rec.Code, rec.Body.String())
	}

	var body422 struct {
		Error struct {
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body422); err != nil {
		t.Fatalf("อ่าน response ไม่ได้: %v", err)
	}
	if body422.Error.Fields["amenity_ids"] == "" {
		t.Errorf("ต้องบอกว่าผิดที่ฟิลด์ amenity_ids แต่ได้ %v", body422.Error.Fields)
	}

	// ห้องต้องไม่ถูกสร้างค้างไว้ เพราะทั้งสองขั้นอยู่ใน transaction เดียวกัน
	var count int
	if err := e.app.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM rooms WHERE room_number = '405'`).Scan(&count); err != nil {
		t.Fatalf("นับห้องไม่สำเร็จ: %v", err)
	}
	if count != 0 {
		t.Errorf("มีห้องค้างในฐานข้อมูล %d แถว ต้องถูก rollback ไปแล้ว", count)
	}
}

// ส่ง amenity ซ้ำมาไม่ควรชน primary key ของ room_amenities แล้วถูกรายงานผิดเป็น "เลขห้องซ้ำ"
func TestAdminCreateAcceptsDuplicateAmenityIDs(t *testing.T) {
	e := newEnv(t)

	aircon := insertAmenity(t, e, "aircon", "เครื่องปรับอากาศ", 0)

	body := `{"room_number":"406","floor":4,"stay_type":"monthly","price":3400,
	          "amenity_ids":["` + aircon.String() + `","` + aircon.String() + `"]}`
	rec := e.do(t, http.MethodPost, "/api/v1/admin/rooms", body, e.adminToken(t))
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

	got := e.do(t, http.MethodGet, "/api/v1/rooms/"+created.Data.ID, "", "")
	want := []string{"เครื่องปรับอากาศ"}
	if names := roomAmenityNames(t, got); !equalStrings(names, want) {
		t.Errorf("amenities = %v ต้องเหลือตัวเดียวเป็น %v", names, want)
	}
}
