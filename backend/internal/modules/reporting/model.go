package reporting

import (
	"backend/internal/modules/room"
	"backend/internal/shared/audit"
)

// Dashboard คือหน้าแรกของผู้ดูแล — ตัวเลขห้องว่างรายสาขา + กิจกรรมล่าสุด
//
// กล่องล่างเป็น activity log ไม่ใช่รายการจอง เพราะสิ่งที่ผู้ดูแลต้องรู้เมื่อเปิดหน้ามา
// คือ "เกิดอะไรขึ้นบ้างตั้งแต่ครั้งก่อน" ซึ่งรวมการอนุมัติ การแก้ห้อง และการแก้สาขาด้วย
// ไม่ใช่แค่ใบจองที่เข้ามาใหม่
type Dashboard struct {
	Branches         []room.BranchStats `json:"branches"`
	RecentActivities []audit.Log        `json:"recent_activities"`
}
