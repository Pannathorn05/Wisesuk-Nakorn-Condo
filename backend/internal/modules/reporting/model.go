package reporting

import (
	"backend/internal/modules/booking"
	"backend/internal/modules/room"
)

// Dashboard คือหน้าแรกของแอดมิน — ตัวเลขห้องว่างรายสาขา + การจองล่าสุด
//
// เป็น model ที่ประกอบจากของ module อื่น จึงต้อง import room และ booking
// (ทั้งสอง module ไม่ import reporting กลับ จึงไม่เกิด import cycle)
type Dashboard struct {
	Branches []room.BranchStats `json:"branches"`
	Recent   []booking.Booking  `json:"recent_bookings"`
}
