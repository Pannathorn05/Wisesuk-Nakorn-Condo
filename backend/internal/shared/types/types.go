// Package types เก็บ enum ที่ใช้ร่วมกันข้าม module เท่านั้น
//
// ตัว model ของแต่ละเรื่องอยู่ใน model.go ของ module นั้น ๆ
// จะย้ายมาที่นี่ก็ต่อเมื่อมีมากกว่าหนึ่ง module ต้องใช้จริง ๆ
package types

// Role คือสิทธิ์ผู้ใช้ 3 ระดับ (Guest ไม่ต้องมีบัญชีจึงไม่นับเป็น role)
// ใช้ใน account (เจ้าของ), middleware (ตรวจสิทธิ์) และ audit (บันทึกว่าใครทำ)
type Role string

const (
	RoleMember     Role = "member"
	RoleAdmin      Role = "admin"
	RoleSuperAdmin Role = "superadmin"
)

func (r Role) Valid() bool {
	switch r {
	case RoleMember, RoleAdmin, RoleSuperAdmin:
		return true
	}
	return false
}

// StayType คือรูปแบบการเข้าพัก ใช้ทั้งใน room (ห้องเปิดให้จองแบบไหน)
// และ booking (ลูกค้าจองแบบไหน)
type StayType string

const (
	StayDaily   StayType = "daily"
	StayMonthly StayType = "monthly"
)

func (s StayType) Valid() bool { return s == StayDaily || s == StayMonthly }
