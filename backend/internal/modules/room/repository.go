package room

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
	r.id, r.branch_id, r.room_type_id, r.room_number, r.building, r.floor,
	r.stay_type, r.price, r.water_rate, r.electric_rate, r.size_sqm,
	r.description, r.image_url, r.status, r.is_active, r.created_at, r.updated_at,
	COALESCE(b.name, ''), COALESCE(rt.name, '')`

const joins = `
	FROM rooms r
	LEFT JOIN branches b    ON b.id  = r.branch_id
	LEFT JOIN room_types rt ON rt.id = r.room_type_id`

func scan(row interface{ Scan(...any) error }) (*Room, error) {
	var rm Room
	err := row.Scan(
		&rm.ID, &rm.BranchID, &rm.RoomTypeID, &rm.RoomNumber, &rm.Building, &rm.Floor,
		&rm.StayType, &rm.Price, &rm.WaterRate, &rm.ElectricRate, &rm.SizeSqm,
		&rm.Description, &rm.ImageURL, &rm.Status, &rm.IsActive, &rm.CreatedAt, &rm.UpdatedAt,
		&rm.BranchName, &rm.RoomTypeName,
	)
	if err != nil {
		return nil, database.NormalizeErr(err)
	}
	return &rm, nil
}

// SearchParams รองรับตัวกรองในหน้า "ค้นหาห้องพัก" ทั้งหมด
type SearchParams struct {
	BranchID     *uuid.UUID
	RoomTypeID   *uuid.UUID
	StayType     *types.StayType
	CheckIn      *time.Time
	CheckOut     *time.Time
	MoveInDate   *time.Time
	MinPrice     *float64
	MaxPrice     *float64
	OnlyBookable bool // true = คืนเฉพาะห้องที่จองได้จริง (ใช้ฝั่งผู้ใช้)
	Limit        int
	Offset       int
}

func (r *Repository) Search(ctx context.Context, p SearchParams) ([]Room, int, error) {
	where := []string{"r.is_active"}
	b := &database.Binder{}

	if p.BranchID != nil {
		where = append(where, "r.branch_id = "+b.Bind(*p.BranchID))
	}
	if p.RoomTypeID != nil {
		where = append(where, "r.room_type_id = "+b.Bind(*p.RoomTypeID))
	}
	if p.StayType != nil {
		where = append(where, "r.stay_type = "+b.Bind(*p.StayType))
	}
	if p.MinPrice != nil {
		where = append(where, "r.price >= "+b.Bind(*p.MinPrice))
	}
	if p.MaxPrice != nil {
		where = append(where, "r.price <= "+b.Bind(*p.MaxPrice))
	}
	if p.OnlyBookable {
		where = append(where, "r.status = 'available'")
		where = append(where, availabilityClause(b.Bind(p.CheckIn), b.Bind(p.CheckOut), b.Bind(p.MoveInDate)))
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
		` ORDER BY r.branch_id, r.room_number LIMIT ` + b.Bind(limit) + ` OFFSET ` + b.Bind(p.Offset)

	rows, err := exec.Query(ctx, q, b.Args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []Room{}
	for rows.Next() {
		rm, err := scan(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *rm)
	}
	return out, total, rows.Err()
}

// availabilityClause ตัดห้องที่มีการจองทับช่วงวันที่ที่ขอออกไป
// นับเฉพาะการจองที่ยัง "มีชีวิต" คือ รอชำระ / รอตรวจสอบ / อนุมัติแล้ว
//   - รายวัน  : ไม่ว่างเมื่อช่วง [check_in, check_out) ซ้อนทับกัน
//   - รายเดือน: ห้องถูกจองยาว ถือว่าไม่ว่างตั้งแต่วันเข้าอยู่เป็นต้นไป
func availabilityClause(checkIn, checkOut, moveIn string) string {
	return `NOT EXISTS (
		SELECT 1 FROM bookings bk
		WHERE bk.room_id = r.id
		  AND bk.status IN ('pending_payment', 'awaiting_review', 'approved')
		  AND (
		    (bk.stay_type = 'daily'
		      AND ` + checkIn + `::date IS NOT NULL AND ` + checkOut + `::date IS NOT NULL
		      AND bk.check_in_date < ` + checkOut + `::date
		      AND bk.check_out_date > ` + checkIn + `::date)
		    OR (bk.stay_type = 'monthly'
		      AND (bk.move_in_date IS NULL
		           OR bk.move_in_date <= COALESCE(` + moveIn + `::date, CURRENT_DATE)))
		  )
	)`
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Room, error) {
	q := `SELECT ` + columns + joins + ` WHERE r.id = $1`
	return scan(r.db.Executor(ctx).QueryRow(ctx, q, id))
}

// Lock ล็อกแถวห้อง (SELECT ... FOR UPDATE) ต้องเรียกภายใน transaction
// คำขอจองที่เข้ามาพร้อมกันจะรอที่บรรทัดนี้ ทำให้ตรวจความว่างได้ถูกต้อง
func (r *Repository) Lock(ctx context.Context, roomID uuid.UUID) (*Room, error) {
	var id uuid.UUID
	if err := r.db.Executor(ctx).QueryRow(ctx,
		`SELECT id FROM rooms WHERE id = $1 FOR UPDATE`, roomID).Scan(&id); err != nil {
		return nil, database.NormalizeErr(err)
	}
	return r.GetByID(ctx, id)
}

// IsAvailable ตรวจซ้ำอีกครั้งตอนกดจอง หลังจากล็อกแถวห้องไว้แล้ว
func (r *Repository) IsAvailable(ctx context.Context, roomID uuid.UUID, checkIn, checkOut *time.Time) (bool, error) {
	const q = `
		SELECT r.status = 'available' AND r.is_active AND NOT EXISTS (
			SELECT 1 FROM bookings bk
			WHERE bk.room_id = r.id
			  AND bk.status IN ('pending_payment', 'awaiting_review', 'approved')
			  AND (
			    (bk.stay_type = 'daily' AND $2::date IS NOT NULL AND $3::date IS NOT NULL
			      AND bk.check_in_date < $3::date AND bk.check_out_date > $2::date)
			    OR (bk.stay_type = 'monthly')
			  )
		)
		FROM rooms r WHERE r.id = $1`
	var ok bool
	if err := r.db.Executor(ctx).QueryRow(ctx, q, roomID, checkIn, checkOut).Scan(&ok); err != nil {
		return false, database.NormalizeErr(err)
	}
	return ok, nil
}

type SaveParams struct {
	BranchID     uuid.UUID
	RoomTypeID   *uuid.UUID
	RoomNumber   string
	Building     string
	Floor        int
	StayType     types.StayType
	Price        float64
	WaterRate    float64
	ElectricRate float64
	SizeSqm      *float64
	Description  string
	ImageURL     *string
	Status       RoomStatus
}

func (r *Repository) Create(ctx context.Context, p SaveParams) (*Room, error) {
	const q = `
		INSERT INTO rooms (branch_id, room_type_id, room_number, building, floor, stay_type,
		                   price, water_rate, electric_rate, size_sqm, description, image_url, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, COALESCE($12, ''), $13)
		RETURNING id`
	var id uuid.UUID
	err := r.db.Executor(ctx).QueryRow(ctx, q,
		p.BranchID, p.RoomTypeID, p.RoomNumber, p.Building, p.Floor, p.StayType,
		p.Price, p.WaterRate, p.ElectricRate, p.SizeSqm, p.Description, p.ImageURL, p.Status,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

// Update จำกัดด้วย branch_id เสมอ แอดมินจึงแก้ได้เฉพาะห้องในสาขาตนเอง
func (r *Repository) Update(ctx context.Context, id, branchID uuid.UUID, p SaveParams) (*Room, error) {
	const q = `
		UPDATE rooms SET
			room_type_id = $3, room_number = $4, building = $5, floor = $6, stay_type = $7,
			price = $8, water_rate = $9, electric_rate = $10, size_sqm = $11,
			description = $12, image_url = COALESCE($13, image_url), status = $14,
			updated_at = now()
		WHERE id = $1 AND branch_id = $2`
	tag, err := r.db.Executor(ctx).Exec(ctx, q, id, branchID,
		p.RoomTypeID, p.RoomNumber, p.Building, p.Floor, p.StayType,
		p.Price, p.WaterRate, p.ElectricRate, p.SizeSqm, p.Description, p.ImageURL, p.Status)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, database.ErrNotFound
	}
	return r.GetByID(ctx, id)
}

// UpdateStatus ใช้กับปุ่มอัปเดตสถานะห้องแบบ real-time (ว่าง / มีผู้เช่า / ปิดปรับปรุง)
func (r *Repository) UpdateStatus(ctx context.Context, id, branchID uuid.UUID, status RoomStatus) (*Room, error) {
	tag, err := r.db.Executor(ctx).Exec(ctx,
		`UPDATE rooms SET status = $3, updated_at = now() WHERE id = $1 AND branch_id = $2`,
		id, branchID, status)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, database.ErrNotFound
	}
	return r.GetByID(ctx, id)
}

// ReleaseIfIdle คืนห้องเป็น "ว่าง" เมื่อไม่มีการจองที่ยังถือห้องนี้อยู่แล้ว
//
// เงื่อนไข status = 'occupied' กันไม่ให้ไปทับสถานะ maintenance ที่แอดมินตั้งไว้เอง
// ส่วน NOT EXISTS กันการปล่อยห้องทั้งที่ยังมีการจองใบอื่นถืออยู่
//
// ต้องเรียกหลังเปลี่ยนสถานะการจองแล้ว และอยู่ใน transaction เดียวกัน
// ไม่งั้น NOT EXISTS จะยังเห็นใบที่กำลังถูกยกเลิกอยู่
//
// ไม่ตรวจ RowsAffected เพราะ "ไม่มีอะไรเปลี่ยน" คือผลลัพธ์ที่ถูกต้อง ไม่ใช่ error
func (r *Repository) ReleaseIfIdle(ctx context.Context, roomID, branchID uuid.UUID) error {
	_, err := r.db.Executor(ctx).Exec(ctx, `
		UPDATE rooms SET status = 'available', updated_at = now()
		WHERE id = $1 AND branch_id = $2 AND status = 'occupied'
		  AND NOT EXISTS (
		      SELECT 1 FROM bookings
		      WHERE room_id = rooms.id
		        AND stay_type = 'monthly'
		        AND status IN ('pending_payment', 'awaiting_review', 'approved'))`,
		roomID, branchID)
	return err
}

// Delete เป็น soft delete เพื่อรักษาประวัติการจองที่อ้างถึงห้องนี้
func (r *Repository) Delete(ctx context.Context, id, branchID uuid.UUID) error {
	tag, err := r.db.Executor(ctx).Exec(ctx,
		`UPDATE rooms SET is_active = FALSE, updated_at = now() WHERE id = $1 AND branch_id = $2`,
		id, branchID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

// ---------------------------------------------------------------- room types

func (r *Repository) ListTypes(ctx context.Context, branchID *uuid.UUID) ([]RoomType, error) {
	rows, err := r.db.Executor(ctx).Query(ctx,
		`SELECT id, branch_id, name, description, size_sqm, image_url, sort_order
		 FROM room_types WHERE ($1::uuid IS NULL OR branch_id = $1)
		 ORDER BY sort_order, name`, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []RoomType{}
	for rows.Next() {
		var t RoomType
		if err := rows.Scan(&t.ID, &t.BranchID, &t.Name, &t.Description, &t.SizeSqm, &t.ImageURL, &t.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------- stats

// StatsByBranch สรุปห้องว่างรายวัน/รายเดือน และจำนวนรายการที่รอตรวจสอบ (ใช้บนแดชบอร์ด)
func (r *Repository) StatsByBranch(ctx context.Context, branchID *uuid.UUID) ([]BranchStats, error) {
	const q = `
		SELECT b.id, b.name,
		  COUNT(*) FILTER (WHERE r.stay_type = 'daily'   AND r.status = 'available')  AS daily_free,
		  COUNT(*) FILTER (WHERE r.stay_type = 'daily')                               AS daily_total,
		  COUNT(*) FILTER (WHERE r.stay_type = 'monthly' AND r.status = 'available')  AS monthly_free,
		  COUNT(*) FILTER (WHERE r.stay_type = 'monthly')                             AS monthly_total,
		  (SELECT COUNT(*) FROM bookings bk
		     WHERE bk.branch_id = b.id AND bk.status = 'awaiting_review')             AS pending_review,
		  (SELECT COUNT(*) FROM bookings bk WHERE bk.branch_id = b.id)                AS bookings_total
		FROM branches b
		LEFT JOIN rooms r ON r.branch_id = b.id AND r.is_active
		WHERE ($1::uuid IS NULL OR b.id = $1)
		GROUP BY b.id, b.name, b.created_at
		ORDER BY b.created_at`
	rows, err := r.db.Executor(ctx).Query(ctx, q, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []BranchStats{}
	for rows.Next() {
		var s BranchStats
		if err := rows.Scan(&s.BranchID, &s.BranchName,
			&s.DailyRoomsFree, &s.DailyRoomsTotal,
			&s.MonthlyRoomsFree, &s.MonthlyRoomsTotal,
			&s.PendingReview, &s.BookingsTotal); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
