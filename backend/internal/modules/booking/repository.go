package booking

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/database"
	"backend/internal/shared/types"
)

type Repository struct{ db *database.TxManager }

func NewRepository(db *database.TxManager) *Repository { return &Repository{db: db} }

const columns = `
	bk.id, bk.code, bk.user_id, bk.branch_id, bk.room_id, bk.stay_type,
	bk.guest_first_name, bk.guest_last_name, bk.guest_phone,
	bk.emergency_phone, bk.emergency_relation,
	bk.check_in_date, bk.check_out_date, bk.nights,
	bk.move_in_date, bk.contract_date, bk.appointment_at, bk.appointment_note,
	bk.total_amount, bk.status, bk.reviewed_by, bk.reviewed_at,
	bk.cancelled_at, bk.created_at, bk.updated_at,
	COALESCE(b.name, ''), COALESCE(r.room_number, ''),
	COALESCE(u.first_name || ' ' || u.last_name, ''), COALESCE(u.email, '')`

const joins = `
	FROM bookings bk
	LEFT JOIN branches b ON b.id = bk.branch_id
	LEFT JOIN rooms    r ON r.id = bk.room_id
	LEFT JOIN users    u ON u.id = bk.user_id`

func scan(row interface{ Scan(...any) error }) (*Booking, error) {
	var b Booking
	err := row.Scan(
		&b.ID, &b.Code, &b.UserID, &b.BranchID, &b.RoomID, &b.StayType,
		&b.GuestFirstName, &b.GuestLastName, &b.GuestPhone,
		&b.EmergencyPhone, &b.EmergencyRelation,
		&b.CheckInDate, &b.CheckOutDate, &b.Nights,
		&b.MoveInDate, &b.ContractDate, &b.AppointmentAt, &b.AppointmentNote,
		&b.TotalAmount, &b.Status, &b.ReviewedBy, &b.ReviewedAt,
		&b.CancelledAt, &b.CreatedAt, &b.UpdatedAt,
		&b.BranchName, &b.RoomNumber, &b.MemberName, &b.MemberEmail,
	)
	if err != nil {
		return nil, database.NormalizeErr(err)
	}
	return &b, nil
}

type CreateParams struct {
	UserID   uuid.UUID
	BranchID uuid.UUID
	RoomID   uuid.UUID
	StayType types.StayType

	GuestFirstName    string
	GuestLastName     string
	GuestPhone        string
	EmergencyPhone    string
	EmergencyRelation string

	CheckInDate  *time.Time
	CheckOutDate *time.Time
	Nights       *int
	MoveInDate   *time.Time
	ContractDate *time.Time

	TotalAmount float64
}

func (r *Repository) Create(ctx context.Context, p CreateParams) (*Booking, error) {
	// รหัสจองรูปแบบ PT-001 ตามที่ prototype แสดง
	const q = `
		INSERT INTO bookings (
			code, user_id, branch_id, room_id, stay_type,
			guest_first_name, guest_last_name, guest_phone, emergency_phone, emergency_relation,
			check_in_date, check_out_date, nights, move_in_date, contract_date, total_amount)
		VALUES (
			'PT-' || LPAD(nextval('booking_code_seq')::text, 3, '0'),
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id`
	var id uuid.UUID
	err := r.db.Executor(ctx).QueryRow(ctx, q,
		p.UserID, p.BranchID, p.RoomID, p.StayType,
		p.GuestFirstName, p.GuestLastName, p.GuestPhone, p.EmergencyPhone, p.EmergencyRelation,
		p.CheckInDate, p.CheckOutDate, p.Nights, p.MoveInDate, p.ContractDate, p.TotalAmount,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Booking, error) {
	q := `SELECT ` + columns + joins + ` WHERE bk.id = $1`
	return scan(r.db.Executor(ctx).QueryRow(ctx, q, id))
}

// ListParams รองรับตัวกรองของทั้งหน้า "ประวัติการจอง" (สมาชิก)
// และหน้า "จัดการการจอง" (แอดมิน: ทั้งหมด / รอตรวจสอบ / อนุมัติแล้ว / ปฏิเสธ, รายวัน / รายเดือน)
type ListParams struct {
	UserID   *uuid.UUID
	BranchID *uuid.UUID
	Status   *BookingStatus
	StayType *types.StayType
	Search   string
	Limit    int
	Offset   int
}

func (r *Repository) List(ctx context.Context, p ListParams) ([]Booking, int, error) {
	where := []string{"TRUE"}
	b := &database.Binder{}

	if p.UserID != nil {
		where = append(where, "bk.user_id = "+b.Bind(*p.UserID))
	}
	if p.BranchID != nil {
		where = append(where, "bk.branch_id = "+b.Bind(*p.BranchID))
	}
	if p.Status != nil {
		where = append(where, "bk.status = "+b.Bind(*p.Status))
	}
	if p.StayType != nil {
		where = append(where, "bk.stay_type = "+b.Bind(*p.StayType))
	}
	if s := strings.TrimSpace(p.Search); s != "" {
		ph := b.Bind(s)
		where = append(where, `(bk.code ILIKE '%' || `+ph+` || '%'
			OR r.room_number ILIKE '%' || `+ph+` || '%'
			OR u.first_name  ILIKE '%' || `+ph+` || '%'
			OR u.last_name   ILIKE '%' || `+ph+` || '%')`)
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")
	exec := r.db.Executor(ctx)

	var total int
	if err := exec.QueryRow(ctx, `SELECT COUNT(*) `+joins+` `+whereSQL, b.Args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := p.Limit
	if limit <= 0 {
		limit = 20
	}
	q := `SELECT ` + columns + joins + ` ` + whereSQL +
		` ORDER BY bk.created_at DESC LIMIT ` + b.Bind(limit) + ` OFFSET ` + b.Bind(p.Offset)

	rows, err := exec.Query(ctx, q, b.Args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return collect(rows, total)
}

// UpdateStatus เปลี่ยนสถานะการจอง โดยยืนยัน branch_id เพื่อกันแอดมินข้ามสาขา
// branchID = nil หมายถึงข้ามการตรวจ (super admin หรือเจ้าของการจองเอง)
func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, branchID *uuid.UUID, status BookingStatus, reviewerID *uuid.UUID) (*Booking, error) {
	// $3 ต้อง cast เป็น booking_status ให้ชัดทุกจุด เพราะถูกใช้ทั้งเป็นค่าที่กำหนดให้คอลัมน์
	// และเป็นค่าที่นำไปเปรียบเทียบ ถ้าไม่ cast PostgreSQL จะสรุปชนิดไม่ตรงกัน
	// แล้วตอบ "inconsistent types deduced for parameter $3" (SQLSTATE 42P08)
	const q = `
		UPDATE bookings SET
			status       = $3::booking_status,
			reviewed_by  = COALESCE($4, reviewed_by),
			reviewed_at  = CASE WHEN $4::uuid IS NULL THEN reviewed_at ELSE now() END,
			cancelled_at = CASE WHEN $3::booking_status = 'cancelled' THEN now() ELSE cancelled_at END,
			updated_at   = now()
		WHERE id = $1
		  AND ($2::uuid IS NULL OR branch_id = $2)`
	tag, err := r.db.Executor(ctx).Exec(ctx, q, id, branchID, status, reviewerID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, database.ErrNotFound
	}
	return r.GetByID(ctx, id)
}

// SetAppointment กำหนด/แก้ไขวันนัดหมายทำสัญญา (เฉพาะการจองรายเดือน)
func (r *Repository) SetAppointment(ctx context.Context, id, branchID uuid.UUID, at time.Time, note string) (*Booking, error) {
	const q = `
		UPDATE bookings SET appointment_at = $3, appointment_note = $4, updated_at = now()
		WHERE id = $1 AND branch_id = $2 AND stay_type = 'monthly'`
	tag, err := r.db.Executor(ctx).Exec(ctx, q, id, branchID, at, note)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, database.ErrNotFound
	}
	return r.GetByID(ctx, id)
}

// RecentByBranch ใช้ในการ์ด "กิจกรรมล่าสุด" บนแดชบอร์ด
func (r *Repository) RecentByBranch(ctx context.Context, branchID *uuid.UUID, limit int) ([]Booking, error) {
	q := `SELECT ` + columns + joins + `
	      WHERE ($1::uuid IS NULL OR bk.branch_id = $1)
	      ORDER BY bk.created_at DESC LIMIT $2`
	rows, err := r.db.Executor(ctx).Query(ctx, q, branchID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out, _, err := collect(rows, 0)
	return out, err
}

// rowSet คืออินเทอร์เฟซร่วมของ pgx.Rows เท่าที่ collect ต้องใช้
type rowSet interface {
	Next() bool
	Scan(...any) error
	Err() error
}

func collect(rs rowSet, total int) ([]Booking, int, error) {
	out := []Booking{}
	for rs.Next() {
		b, err := scan(rs)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *b)
	}
	return out, total, rs.Err()
}

// ---------------------------------------------------------------- payments

const paymentColumns = `
	id, booking_id, amount, transferred_at, slip_url, note, status,
	reviewed_by, reviewed_at, reject_reason, created_at`

func scanPayment(row interface{ Scan(...any) error }) (*Payment, error) {
	var p Payment
	err := row.Scan(
		&p.ID, &p.BookingID, &p.Amount, &p.TransferredAt, &p.SlipURL, &p.Note, &p.Status,
		&p.ReviewedBy, &p.ReviewedAt, &p.RejectReason, &p.CreatedAt,
	)
	if err != nil {
		return nil, database.NormalizeErr(err)
	}
	return &p, nil
}

type CreatePaymentParams struct {
	BookingID     uuid.UUID
	Amount        float64
	TransferredAt time.Time
	SlipURL       string
	Note          string
}

func (r *Repository) CreatePayment(ctx context.Context, p CreatePaymentParams) (*Payment, error) {
	q := `INSERT INTO payments (booking_id, amount, transferred_at, slip_url, note)
	      VALUES ($1, $2, $3, $4, $5)
	      RETURNING ` + paymentColumns
	return scanPayment(r.db.Executor(ctx).QueryRow(ctx, q,
		p.BookingID, p.Amount, p.TransferredAt, p.SlipURL, p.Note))
}

// LatestPayment คืนสลิปล่าสุด (ปุ่ม "ดูสลิป" ในหน้าจัดการการจอง)
func (r *Repository) LatestPayment(ctx context.Context, bookingID uuid.UUID) (*Payment, error) {
	q := `SELECT ` + paymentColumns + ` FROM payments
	      WHERE booking_id = $1 ORDER BY created_at DESC LIMIT 1`
	return scanPayment(r.db.Executor(ctx).QueryRow(ctx, q, bookingID))
}

// ReviewPayment บันทึกผลตรวจสอบสลิปของแอดมิน
func (r *Repository) ReviewPayment(ctx context.Context, bookingID, reviewerID uuid.UUID, status PaymentStatus, reason string) error {
	const q = `
		UPDATE payments SET
			status = $3, reviewed_by = $2, reviewed_at = now(), reject_reason = $4
		WHERE booking_id = $1 AND status = 'submitted'`
	_, err := r.db.Executor(ctx).Exec(ctx, q, bookingID, reviewerID, status, reason)
	return err
}
