package booking

import (
	"time"

	"github.com/google/uuid"

	"backend/internal/shared/types"
)

// BookingStatus คือสถานะการจอง ไล่ตามลำดับที่เกิดขึ้นจริงในระบบ
type BookingStatus string

const (
	BookingPendingPayment BookingStatus = "pending_payment" // จองแล้ว รอแจ้งชำระเงิน
	BookingAwaitingReview BookingStatus = "awaiting_review" // ส่งสลิปแล้ว รอตรวจสอบ
	BookingApproved       BookingStatus = "approved"        // อนุมัติแล้ว
	BookingRejected       BookingStatus = "rejected"        // ปฏิเสธ
	BookingCancelled      BookingStatus = "cancelled"       // ยกเลิก
	BookingCompleted      BookingStatus = "completed"       // เข้าพักครบแล้ว
)

func (s BookingStatus) Valid() bool {
	switch s {
	case BookingPendingPayment, BookingAwaitingReview, BookingApproved,
		BookingRejected, BookingCancelled, BookingCompleted:
		return true
	}
	return false
}

type PaymentStatus string

const (
	PaymentSubmitted PaymentStatus = "submitted"
	PaymentApproved  PaymentStatus = "approved"
	PaymentRejected  PaymentStatus = "rejected"
)

type Booking struct {
	ID       uuid.UUID      `json:"id"`
	Code     string         `json:"code"` // รูปแบบ PT-001
	UserID   uuid.UUID      `json:"user_id"`
	BranchID uuid.UUID      `json:"branch_id"`
	RoomID   uuid.UUID      `json:"room_id"`
	StayType types.StayType `json:"stay_type"`

	// ผู้เข้าพัก (กรอกในแบบฟอร์มจอง อาจต่างจากเจ้าของบัญชี)
	GuestFirstName    string `json:"guest_first_name"`
	GuestLastName     string `json:"guest_last_name"`
	GuestPhone        string `json:"guest_phone"`
	EmergencyPhone    string `json:"emergency_phone"`
	EmergencyRelation string `json:"emergency_relation"`

	// รายวัน
	CheckInDate  *time.Time `json:"check_in_date,omitempty"`
	CheckOutDate *time.Time `json:"check_out_date,omitempty"`
	Nights       *int       `json:"nights,omitempty"`

	// รายเดือน
	MoveInDate      *time.Time `json:"move_in_date,omitempty"`
	ContractDate    *time.Time `json:"contract_date,omitempty"`
	AppointmentAt   *time.Time `json:"appointment_at,omitempty"` // วันนัดหมายที่แอดมินกำหนด
	AppointmentNote string     `json:"appointment_note"`

	TotalAmount  float64       `json:"total_amount"`
	Status       BookingStatus `json:"status"`
	RejectReason string        `json:"reject_reason"`
	ReviewedBy   *uuid.UUID    `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time    `json:"reviewed_at,omitempty"`
	CancelledAt  *time.Time    `json:"cancelled_at,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`

	// ฟิลด์จากการ join ไว้แสดงในตาราง โดยไม่ต้องยิง API ซ้ำ
	BranchName    string   `json:"branch_name,omitempty"`
	RoomNumber    string   `json:"room_number,omitempty"`
	MemberName    string   `json:"member_name,omitempty"`
	MemberEmail   string   `json:"member_email,omitempty"`
	LatestPayment *Payment `json:"latest_payment,omitempty"`
}

// Payment คือการแจ้งชำระเงินหนึ่งครั้ง พร้อมสลิปที่แนบมา
type Payment struct {
	ID            uuid.UUID     `json:"id"`
	BookingID     uuid.UUID     `json:"booking_id"`
	Amount        float64       `json:"amount"`
	TransferredAt time.Time     `json:"transferred_at"`
	SlipURL       string        `json:"slip_url"`
	Note          string        `json:"note"`
	Status        PaymentStatus `json:"status"`
	ReviewedBy    *uuid.UUID    `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time    `json:"reviewed_at,omitempty"`
	RejectReason  string        `json:"reject_reason"`
	CreatedAt     time.Time     `json:"created_at"`
}
