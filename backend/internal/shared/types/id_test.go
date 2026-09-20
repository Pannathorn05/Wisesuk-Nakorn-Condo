package types_test

import (
	"encoding/json"
	"testing"

	"backend/internal/shared/types"
)

func TestIDString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"เติมศูนย์ครบ 3 หลัก", types.AmenityID(1).String(), "amt-001"},
		{"สองหลัก", types.AmenityID(42).String(), "amt-042"},
		{"ครบ 3 หลักพอดี", types.AmenityID(999).String(), "amt-999"},
		{"เกิน 3 หลักแล้วยาวเอง ไม่ถูกตัด", types.AmenityID(1000).String(), "amt-1000"},
		{"เลขใหญ่มาก", types.AmenityID(1234567).String(), "amt-1234567"},
		{"แต่ละตารางมี prefix ของตัวเอง", types.BranchID(1).String(), "brn-001"},
		{"ห้อง", types.RoomID(7).String(), "rm-007"},
		{"การจอง", types.BookingID(12).String(), "bkg-012"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s: ได้ %q ต้องการ %q", tc.name, tc.got, tc.want)
		}
	}
}

func TestParseIDAcceptsValidInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw  string
		want types.AmenityID
	}{
		{"amt-001", 1},
		{"amt-1", 1},        // ไม่เติมศูนย์ก็ต้องอ่านได้ คนพิมพ์เองไม่เติมแน่นอน
		{"amt-0000042", 42}, // ศูนย์เกินมาก็ยังเป็นเลขเดิม
		{"AMT-001", 1},      // ตัวพิมพ์ใหญ่ — ผู้ใช้พิมพ์รหัสจากหน้าจอเข้ามาเอง
		{"  amt-001  ", 1},  // ช่องว่างหัวท้ายจากการคัดลอกมาวาง
		{"amt-1000", 1000},
	}
	for _, tc := range cases {
		got, err := types.ParseID[types.Amenity](tc.raw)
		if err != nil {
			t.Errorf("ParseID(%q) คืน error: %v", tc.raw, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseID(%q) = %d ต้องการ %d", tc.raw, got, tc.want)
		}
	}
}

func TestParseIDRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	// prefix ของตารางอื่นต้องไม่ผ่าน ไม่งั้นส่ง room id ไปช่อง branch id แล้วเงียบ
	bad := []string{
		"1",                         // ตัวเลขเปล่า — ต้องมี prefix เสมอ
		"rm-001",                    // prefix ของตารางอื่น
		"brn-001",                   // prefix ของตารางอื่น
		"amt-",                      // ไม่มีเลข
		"amt-0",                     // 0 ไม่มีทางเป็น id จริง (BIGSERIAL เริ่มที่ 1)
		"amt--1",                    // ติดลบ
		"amt-1.5",                   // ไม่ใช่จำนวนเต็ม
		"amt-abc",                   // ไม่ใช่ตัวเลข
		"amt-1 OR 1=1",              // ความพยายามแทรก SQL
		"amt_001",                   // คั่นผิดตัว
		"",                          // ว่าง
		"amt-999999999999999999999", // ล้น int64
	}
	for _, raw := range bad {
		if got, err := types.ParseID[types.Amenity](raw); err == nil {
			t.Errorf("ParseID(%q) ควรเป็น error แต่ได้ %d", raw, got)
		}
	}
}

func TestIDJSONRoundTrip(t *testing.T) {
	t.Parallel()

	type payload struct {
		RoomID   types.RoomID   `json:"room_id"`
		BranchID types.BranchID `json:"branch_id"`
	}

	out, err := json.Marshal(payload{RoomID: 7, BranchID: 1})
	if err != nil {
		t.Fatalf("marshal ไม่สำเร็จ: %v", err)
	}
	if want := `{"room_id":"rm-007","branch_id":"brn-001"}`; string(out) != want {
		t.Errorf("marshal ได้ %s ต้องการ %s", out, want)
	}

	var back payload
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("unmarshal ไม่สำเร็จ: %v", err)
	}
	if back.RoomID != 7 || back.BranchID != 1 {
		t.Errorf("อ่านกลับได้ %v ต้องการ room=7 branch=1", back)
	}

	// สลับ prefix ระหว่างสองช่องต้องไม่ผ่าน
	if err := json.Unmarshal([]byte(`{"room_id":"brn-001","branch_id":"brn-001"}`), &back); err == nil {
		t.Error("ใส่ prefix ผิดตารางใน JSON ควรเป็น error")
	}

	// ตัวเลขดิบต้องไม่ผ่านเช่นกัน จะได้ไม่มีใครเผลอส่ง id ที่ไม่รู้ว่าของตารางไหน
	if err := json.Unmarshal([]byte(`{"room_id":7,"branch_id":1}`), &back); err == nil {
		t.Error("ใส่ตัวเลขดิบใน JSON ควรเป็น error")
	}
}

func TestIDValid(t *testing.T) {
	t.Parallel()

	if types.RoomID(0).Valid() {
		t.Error("id = 0 ต้องไม่ valid เพราะ BIGSERIAL เริ่มที่ 1")
	}
	if !types.RoomID(1).Valid() {
		t.Error("id = 1 ต้อง valid")
	}
}
