package types

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// ID คือกุญแจหลักของทุกตารางในระบบ
//
// ในฐานข้อมูลเป็น BIGSERIAL ธรรมดา (JOIN และ index จึงเร็วเท่าเดิม) แต่ทุกครั้งที่
// ออกไปนอกระบบ — JSON, URL, activity log — จะมี prefix ของตารางนำหน้าเสมอ เช่น
// สาขาที่ id = 1 ปรากฏเป็น "brn-001" ทำให้อ่านแล้วรู้ทันทีว่ารหัสนี้เป็นของอะไร
// ไม่ต้องไล่ดูว่ามาจาก endpoint ไหน
//
// พารามิเตอร์ E บอกว่า id นี้เป็นของตารางไหน สองตารางจึงเป็นคนละชนิดในสายตา
// compiler — เผลอส่ง RoomID ไปยังฟังก์ชันที่รับ BranchID จะ build ไม่ผ่านตั้งแต่แรก
// ซึ่งเป็นความผิดพลาดที่เกิดง่ายมากตอนที่ทุก id เป็นตัวเลขเปล่าเหมือนกันหมด
type ID[E Entity] int64

// Entity คือตารางหนึ่งตาราง หน้าที่เดียวคือบอก prefix ของตัวเอง
type Entity interface {
	IDPrefix() string
}

// ---------------------------------------------------------------- ตารางทั้งหมด

type (
	Branch       struct{}
	BranchImage  struct{}
	Amenity      struct{}
	NearbyPlace  struct{}
	User         struct{}
	RoomType     struct{}
	Room         struct{}
	Booking      struct{}
	Payment      struct{}
	Notification struct{}
	ActivityLog  struct{}
	Asset        struct{}
)

func (Branch) IDPrefix() string       { return "brn" }
func (BranchImage) IDPrefix() string  { return "bimg" }
func (Amenity) IDPrefix() string      { return "amt" }
func (NearbyPlace) IDPrefix() string  { return "np" }
func (User) IDPrefix() string         { return "usr" }
func (RoomType) IDPrefix() string     { return "rmt" }
func (Room) IDPrefix() string         { return "rm" }
func (Booking) IDPrefix() string      { return "bkg" }
func (Payment) IDPrefix() string      { return "pay" }
func (Notification) IDPrefix() string { return "ntf" }
func (ActivityLog) IDPrefix() string  { return "log" }
func (Asset) IDPrefix() string        { return "ast" }

// ชื่อที่อ่านง่ายสำหรับใช้ประกาศฟิลด์ — ID[Branch] กับ BranchID คือชนิดเดียวกัน
type (
	BranchID       = ID[Branch]
	BranchImageID  = ID[BranchImage]
	AmenityID      = ID[Amenity]
	NearbyPlaceID  = ID[NearbyPlace]
	UserID         = ID[User]
	RoomTypeID     = ID[RoomType]
	RoomID         = ID[Room]
	BookingID      = ID[Booking]
	PaymentID      = ID[Payment]
	NotificationID = ID[Notification]
	ActivityLogID  = ID[ActivityLog]
	AssetID        = ID[Asset]
)

// ---------------------------------------------------------------- แปลงค่า

var ErrInvalidID = errors.New("types: รูปแบบรหัสอ้างอิงไม่ถูกต้อง")

// minDigits คือจำนวนหลักขั้นต่ำหลัง prefix — เลขที่ยาวกว่านี้ไม่ถูกตัด
// (amt-001 ... amt-999 แล้วต่อด้วย amt-1000 เอง ไม่มีเพดาน)
const minDigits = 3

// Prefix คืนคำนำหน้าของตารางที่ id นี้สังกัด
func (id ID[E]) Prefix() string {
	var e E
	return e.IDPrefix()
}

// Valid — id ที่ยังไม่ถูกกำหนดค่าคือ 0 ซึ่งไม่มีทางเป็นค่าจริงจาก BIGSERIAL (เริ่มที่ 1)
func (id ID[E]) Valid() bool { return id > 0 }

func (id ID[E]) String() string {
	digits := strconv.FormatInt(int64(id), 10)
	if n := minDigits - len(digits); n > 0 {
		digits = strings.Repeat("0", n) + digits
	}
	return id.Prefix() + "-" + digits
}

func (id ID[E]) MarshalJSON() ([]byte, error) {
	return []byte(`"` + id.String() + `"`), nil
}

func (id *ID[E]) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return ErrInvalidID
	}
	parsed, err := ParseID[E](raw)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// ParseID อ่านรหัสที่มี prefix กลับเป็นตัวเลข
//
// บังคับว่า prefix ต้องตรงกับตารางที่คาดไว้ ไม่รับตัวเลขเปล่าและไม่รับ prefix ของ
// ตารางอื่น — ส่ง "rm-001" มาที่ช่อง branch_id ต้องเป็น 400 ตั้งแต่ขอบระบบ ไม่ใช่
// ไปโผล่เป็น "ไม่พบข้อมูล" ทีหลังซึ่งตามหาต้นเหตุยากกว่ามาก
//
// ตัวพิมพ์ใหญ่-เล็กของ prefix ไม่มีผล เพราะผู้ใช้พิมพ์รหัสจากหน้าจอเข้ามาเองได้
func ParseID[E Entity](raw string) (ID[E], error) {
	var e E
	prefix := e.IDPrefix()

	rest, found := strings.CutPrefix(strings.TrimSpace(raw), prefix+"-")
	if !found {
		if rest, found = strings.CutPrefix(strings.ToLower(strings.TrimSpace(raw)), prefix+"-"); !found {
			return 0, ErrInvalidID
		}
	}

	// ตัดศูนย์นำหน้าทิ้งได้เอง แต่ต้องเป็นตัวเลขล้วน — ParseInt ปฏิเสธ "+1" และ " 1" ให้แล้ว
	n, err := strconv.ParseInt(rest, 10, 64)
	if err != nil || n < 1 {
		return 0, ErrInvalidID
	}
	return ID[E](n), nil
}
