package booking

import (
	"context"
	"time"

	"github.com/google/uuid"

	"backend/internal/database"
	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/modules/room"
	"backend/internal/shared/access"
	"backend/internal/shared/audit"
	"backend/internal/shared/types"
	"backend/internal/validate"
)

// Rooms / Branches / Notifier คือสัญญาที่ module นี้ต้องการจาก module อื่น
// ประกาศไว้ฝั่งผู้ใช้งาน ทำให้ booking ไม่ต้องผูกกับ module เหล่านั้นตรง ๆ
// และทดสอบแยกได้ด้วยการใส่ของปลอมเข้ามาแทน
type Rooms interface {
	LockForBooking(ctx context.Context, roomID uuid.UUID) (*room.Room, error)
	IsAvailable(ctx context.Context, roomID uuid.UUID, checkIn, checkOut *time.Time) (bool, error)
	MarkOccupied(ctx context.Context, roomID, branchID uuid.UUID) error
	ReleaseIfIdle(ctx context.Context, roomID, branchID uuid.UUID) error
}

type Branches interface {
	ContractFee(ctx context.Context, branchID uuid.UUID) (float64, error)
}

type Notifier interface {
	Notify(ctx context.Context, userID uuid.UUID, title, body, link string)
}

type Service struct {
	repo     *Repository
	tx       *database.TxManager
	rooms    Rooms
	branches Branches
	notifier Notifier
	audit    *audit.Recorder
}

func NewService(repo *Repository, tx *database.TxManager, rooms Rooms, branches Branches, notifier Notifier, rec *audit.Recorder) *Service {
	return &Service{repo: repo, tx: tx, rooms: rooms, branches: branches, notifier: notifier, audit: rec}
}

var errRoomTaken = httpx.NewError(409, "room_unavailable",
	"ขออภัย ห้องนี้ถูกจองไปแล้วในช่วงวันที่ที่เลือก กรุณาเลือกห้องหรือวันที่อื่น")

type CreateInput struct {
	RoomID   uuid.UUID `json:"room_id"`
	StayType string    `json:"stay_type"`

	GuestFirstName    string `json:"guest_first_name"`
	GuestLastName     string `json:"guest_last_name"`
	GuestPhone        string `json:"guest_phone"`
	EmergencyPhone    string `json:"emergency_phone"`
	EmergencyRelation string `json:"emergency_relation"`

	// รายวัน
	CheckInDate  string `json:"check_in_date"`  // YYYY-MM-DD
	CheckOutDate string `json:"check_out_date"` // YYYY-MM-DD
	// รายเดือน
	MoveInDate   string `json:"move_in_date"`
	ContractDate string `json:"contract_date"`
}

// Create ทำรายการจองทั้งหมดใน transaction เดียว
//
// transaction ถูกส่งผ่าน context จึงครอบทั้งการล็อกห้อง (module room),
// การอ่านค่าธรรมเนียมของสาขา (module branch) และการเขียนตาราง bookings
// ถ้ามีขั้นตอนใดล้ม ทุกอย่างถูก rollback พร้อมกัน
func (s *Service) Create(ctx context.Context, identity middleware.Identity, in CreateInput, ip string) (*Booking, error) {
	v := validate.New()
	firstName := v.Required("guest_first_name", in.GuestFirstName)
	lastName := v.Required("guest_last_name", in.GuestLastName)
	phone := v.Phone("guest_phone", in.GuestPhone, true)
	emergencyPhone := v.Phone("emergency_phone", in.EmergencyPhone, false)

	stayType := types.StayType(in.StayType)
	v.Check(stayType.Valid(), "stay_type", "ประเภทการเข้าพักต้องเป็น daily หรือ monthly")
	v.Check(in.RoomID != uuid.Nil, "room_id", "กรุณาเลือกห้องพัก")

	var checkIn, checkOut, moveIn, contractDate *time.Time
	var nights *int

	switch stayType {
	case types.StayDaily:
		checkIn = parseDate(v, "check_in_date", in.CheckInDate, true)
		checkOut = parseDate(v, "check_out_date", in.CheckOutDate, true)
		if checkIn != nil && checkOut != nil {
			if !checkOut.After(*checkIn) {
				v.Check(false, "check_out_date", "วันเช็คเอาท์ต้องอยู่หลังวันเข้าพัก")
			} else {
				n := int(checkOut.Sub(*checkIn).Hours() / 24)
				nights = &n
			}
		}
		if checkIn != nil && checkIn.Before(startOfToday()) {
			v.Check(false, "check_in_date", "วันเข้าพักต้องไม่เป็นวันที่ผ่านมาแล้ว")
		}
	case types.StayMonthly:
		moveIn = parseDate(v, "move_in_date", in.MoveInDate, true)
		contractDate = parseDate(v, "contract_date", in.ContractDate, false)
		if moveIn != nil && moveIn.Before(startOfToday()) {
			v.Check(false, "move_in_date", "วันที่ต้องการเข้าอยู่ต้องไม่เป็นวันที่ผ่านมาแล้ว")
		}
	}

	if err := v.Err(); err != nil {
		return nil, err
	}

	var created *Booking
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		// ล็อกห้องไว้ตลอด transaction — คำขอที่มาพร้อมกันจะรอตรงนี้
		rm, err := s.rooms.LockForBooking(ctx, in.RoomID)
		if err != nil {
			return err
		}
		if !rm.IsActive {
			return httpx.ErrNotFound
		}
		if rm.StayType != stayType {
			return httpx.ValidationFailed(map[string]string{
				"stay_type": "ห้องนี้ไม่ได้เปิดให้จองแบบที่เลือก",
			})
		}

		available, err := s.rooms.IsAvailable(ctx, rm.ID, checkIn, checkOut)
		if err != nil {
			return err
		}
		if !available {
			return errRoomTaken
		}

		contractFee, err := s.branches.ContractFee(ctx, rm.BranchID)
		if err != nil {
			return err
		}

		created, err = s.repo.Create(ctx, CreateParams{
			UserID: identity.UserID, BranchID: rm.BranchID, RoomID: rm.ID, StayType: stayType,
			GuestFirstName: firstName, GuestLastName: lastName, GuestPhone: phone,
			EmergencyPhone: emergencyPhone, EmergencyRelation: in.EmergencyRelation,
			CheckInDate: checkIn, CheckOutDate: checkOut, Nights: nights,
			MoveInDate: moveIn, ContractDate: contractDate,
			TotalAmount: computeTotal(stayType, rm.Price, nights, contractFee),
		})
		return access.MapErr(err)
	})
	if err != nil {
		return nil, err
	}

	s.record(ctx, identity, "booking.create", created.ID.String(),
		map[string]any{"code": created.Code, "stay_type": string(stayType)}, ip)
	s.notifier.Notify(ctx, identity.UserID,
		"สร้างรายการจองสำเร็จ",
		"รายการจอง "+created.Code+" รอการแจ้งชำระเงิน",
		"/bookings/"+created.ID.String())

	return created, nil
}

// computeTotal:
//   - รายวัน  : ราคาห้อง × จำนวนคืน
//   - รายเดือน: เก็บเฉพาะค่ายืนยันการทำสัญญาตอนจอง ค่าเช่าเก็บวันทำสัญญา
func computeTotal(stayType types.StayType, roomPrice float64, nights *int, contractFee float64) float64 {
	if stayType == types.StayDaily && nights != nil {
		return roomPrice * float64(*nights)
	}
	return contractFee
}

// ---------------------------------------------------------------- member

func (s *Service) ListMine(ctx context.Context, identity middleware.Identity, status *BookingStatus, stayType *types.StayType, limit, offset int) ([]Booking, int, error) {
	userID := identity.UserID
	bookings, total, err := s.repo.List(ctx, ListParams{
		UserID: &userID, Status: status, StayType: stayType, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, 0, access.MapErr(err)
	}
	s.attachLatestPayments(ctx, bookings)
	return bookings, total, nil
}

// Get บังคับสิทธิ์: สมาชิกเห็นเฉพาะของตัวเอง, แอดมินเห็นเฉพาะสาขาตัวเอง
func (s *Service) Get(ctx context.Context, identity middleware.Identity, id uuid.UUID) (*Booking, error) {
	b, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, access.MapErr(err)
	}
	if err := canAccess(identity, b); err != nil {
		return nil, err
	}
	if p, err := s.repo.LatestPayment(ctx, b.ID); err == nil {
		b.LatestPayment = p
	}
	return b, nil
}

func canAccess(identity middleware.Identity, b *Booking) error {
	switch identity.Role {
	case types.RoleSuperAdmin:
		return nil
	case types.RoleAdmin:
		if identity.BranchID != nil && *identity.BranchID == b.BranchID {
			return nil
		}
		return httpx.ErrForbidden
	default:
		if identity.UserID == b.UserID {
			return nil
		}
		// ตอบ 404 แทน 403 เพื่อไม่ให้รู้ว่ารหัสจองนี้มีอยู่จริง
		return httpx.ErrNotFound
	}
}

type SubmitPaymentInput struct {
	Amount        float64
	TransferredAt time.Time
	SlipURL       string
	Note          string
}

// SubmitPayment รับสลิปจากสมาชิก แล้วเปลี่ยนสถานะการจองเป็น "รอตรวจสอบ"
func (s *Service) SubmitPayment(ctx context.Context, identity middleware.Identity, bookingID uuid.UUID, in SubmitPaymentInput, ip string) (*Booking, error) {
	b, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, access.MapErr(err)
	}
	if b.UserID != identity.UserID {
		return nil, httpx.ErrNotFound
	}
	if b.Status != BookingPendingPayment && b.Status != BookingRejected {
		return nil, httpx.NewError(409, "invalid_state",
			"รายการจองนี้ไม่อยู่ในสถานะที่แจ้งชำระเงินได้")
	}

	v := validate.New()
	v.Positive("amount", in.Amount)
	v.Check(!in.TransferredAt.IsZero(), "transferred_at", "กรุณาระบุวันและเวลาที่โอน")
	v.Check(!in.TransferredAt.After(time.Now().Add(5*time.Minute)), "transferred_at", "วันเวลาที่โอนต้องไม่เป็นอนาคต")
	v.Check(in.SlipURL != "", "slip", "กรุณาแนบหลักฐานการโอนเงิน")
	if err := v.Err(); err != nil {
		return nil, err
	}

	var updated *Booking
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.repo.CreatePayment(ctx, CreatePaymentParams{
			BookingID:     b.ID,
			Amount:        in.Amount,
			TransferredAt: in.TransferredAt,
			SlipURL:       in.SlipURL,
			Note:          in.Note,
		}); err != nil {
			return access.MapErr(err)
		}
		var err error
		updated, err = s.repo.UpdateStatus(ctx, b.ID, nil, BookingAwaitingReview, nil, "")
		return access.MapErr(err)
	})
	if err != nil {
		return nil, err
	}

	// เติมสลิปที่เพิ่งบันทึกกลับไปด้วย ให้ response มีรูปร่างเดียวกับ GET /bookings/{id}
	// ฝั่งหน้าเว็บจะได้แสดงยืนยันได้ทันทีโดยไม่ต้องยิงซ้ำ
	if p, err := s.repo.LatestPayment(ctx, updated.ID); err == nil {
		updated.LatestPayment = p
	}

	s.record(ctx, identity, "booking.submit_payment", b.ID.String(),
		map[string]any{"code": b.Code, "amount": in.Amount}, ip)
	s.notifier.Notify(ctx, identity.UserID,
		"ส่งหลักฐานการชำระเงินแล้ว",
		"รายการจอง "+b.Code+" อยู่ระหว่างรอผู้ดูแลระบบตรวจสอบ",
		"/bookings/"+b.ID.String())

	return updated, nil
}

func (s *Service) Cancel(ctx context.Context, identity middleware.Identity, bookingID uuid.UUID, ip string) (*Booking, error) {
	b, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, access.MapErr(err)
	}
	if err := canAccess(identity, b); err != nil {
		return nil, err
	}
	// สมาชิกยกเลิกได้เฉพาะก่อนได้รับอนุมัติ หลังจากนั้นต้องให้แอดมินจัดการ
	if identity.Role == types.RoleMember && b.Status != BookingPendingPayment {
		return nil, httpx.NewError(409, "invalid_state",
			"รายการจองนี้ไม่สามารถยกเลิกเองได้ กรุณาติดต่อผู้ดูแลระบบของสาขา")
	}

	// ยกเลิกใบที่ถือห้องอยู่แล้วห้องต้องกลับมาปล่อยเช่าได้
	// ทั้งสองขั้นอยู่ใน transaction เดียวกัน ถ้าคืนห้องพลาดก็ไม่ยกเลิกการจองด้วย
	var updated *Booking
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		var err error
		updated, err = s.repo.UpdateStatus(ctx, b.ID, nil, BookingCancelled, nil, "")
		if err != nil {
			return access.MapErr(err)
		}
		if b.StayType == types.StayMonthly {
			return s.rooms.ReleaseIfIdle(ctx, b.RoomID, b.BranchID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.record(ctx, identity, "booking.cancel", b.ID.String(), map[string]any{"code": b.Code}, ip)
	s.notifier.Notify(ctx, b.UserID, "รายการจองถูกยกเลิก",
		"รายการจอง "+b.Code+" ถูกยกเลิกแล้ว", "/bookings/"+b.ID.String())

	return updated, nil
}

// ---------------------------------------------------------------- admin

func (s *Service) ListForAdmin(ctx context.Context, identity middleware.Identity, branchID *uuid.UUID, status *BookingStatus, stayType *types.StayType, search string, limit, offset int) ([]Booking, int, error) {
	scoped, err := access.Branch(identity, branchID)
	if err != nil {
		return nil, 0, err
	}
	bookings, total, err := s.repo.List(ctx, ListParams{
		BranchID: scoped, Status: status, StayType: stayType, Search: search,
		Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, 0, access.MapErr(err)
	}
	s.attachLatestPayments(ctx, bookings)
	return bookings, total, nil
}

// Review คือปุ่มอนุมัติ/ปฏิเสธหลังตรวจสลิป
// อนุมัติแล้วห้องรายเดือนจะถูกตั้งเป็น "มีผู้เช่า" อัตโนมัติภายใน transaction เดียวกัน
func (s *Service) Review(ctx context.Context, identity middleware.Identity, bookingID uuid.UUID, approve bool, reason string, ip string) (*Booking, error) {
	b, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, access.MapErr(err)
	}
	branchID, err := access.RequireBranch(identity, &b.BranchID)
	if err != nil {
		return nil, err
	}
	if b.Status != BookingAwaitingReview {
		return nil, httpx.NewError(409, "invalid_state", "รายการจองนี้ไม่ได้อยู่ในสถานะรอตรวจสอบ")
	}
	if !approve && reason == "" {
		return nil, httpx.ValidationFailed(map[string]string{
			"reason": "กรุณาระบุเหตุผลที่ปฏิเสธ เพื่อแจ้งกลับไปยังสมาชิก",
		})
	}

	newStatus, paymentStatus := BookingRejected, PaymentRejected
	if approve {
		newStatus, paymentStatus = BookingApproved, PaymentApproved
	}

	var updated *Booking
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.repo.ReviewPayment(ctx, b.ID, identity.UserID, paymentStatus, reason); err != nil {
			return access.MapErr(err)
		}

		reviewer := identity.UserID
		var err error
		updated, err = s.repo.UpdateStatus(ctx, b.ID, &branchID, newStatus, &reviewer, reason)
		if err != nil {
			return access.MapErr(err)
		}

		if approve && b.StayType == types.StayMonthly {
			return s.rooms.MarkOccupied(ctx, b.RoomID, branchID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	action, title, body := "booking.reject", "การจองถูกปฏิเสธ", "รายการจอง "+b.Code+" ถูกปฏิเสธ: "+reason
	if approve {
		action, title, body = "booking.approve", "การจองได้รับการอนุมัติ", "รายการจอง "+b.Code+" ได้รับการอนุมัติเรียบร้อยแล้ว"
	}
	s.record(ctx, identity, action, b.ID.String(), map[string]any{"code": b.Code}, ip)
	s.notifier.Notify(ctx, b.UserID, title, body, "/bookings/"+b.ID.String())

	return updated, nil
}

type AppointmentInput struct {
	AppointmentAt time.Time
	Note          string
}

// SetAppointment กำหนดวันนัดหมายทำสัญญาสำหรับการจองรายเดือน
func (s *Service) SetAppointment(ctx context.Context, identity middleware.Identity, bookingID uuid.UUID, in AppointmentInput, ip string) (*Booking, error) {
	b, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, access.MapErr(err)
	}
	branchID, err := access.RequireBranch(identity, &b.BranchID)
	if err != nil {
		return nil, err
	}
	if b.StayType != types.StayMonthly {
		return nil, httpx.BadRequest("วันนัดหมายทำสัญญาใช้ได้เฉพาะการจองรายเดือน")
	}

	v := validate.New()
	v.Check(!in.AppointmentAt.IsZero(), "appointment_at", "กรุณาระบุวันและเวลานัดหมาย")
	v.Check(in.AppointmentAt.After(time.Now()), "appointment_at", "วันนัดหมายต้องเป็นวันในอนาคต")
	if err := v.Err(); err != nil {
		return nil, err
	}

	updated, err := s.repo.SetAppointment(ctx, b.ID, branchID, in.AppointmentAt, in.Note)
	if err != nil {
		return nil, access.MapErr(err)
	}

	s.record(ctx, identity, "booking.set_appointment", b.ID.String(),
		map[string]any{"code": b.Code, "appointment_at": in.AppointmentAt}, ip)
	s.notifier.Notify(ctx, b.UserID, "นัดหมายทำสัญญา",
		"รายการจอง "+b.Code+" นัดทำสัญญาวันที่ "+in.AppointmentAt.Format("02-01-2006 15:04"),
		"/bookings/"+b.ID.String())

	return updated, nil
}

// ListByMember ให้แอดมินดูประวัติการจองของสมาชิกรายคน
func (s *Service) ListByMember(ctx context.Context, identity middleware.Identity, memberID uuid.UUID, limit, offset int) ([]Booking, int, error) {
	scoped, err := access.Branch(identity, nil)
	if err != nil {
		return nil, 0, err
	}
	bookings, total, err := s.repo.List(ctx, ListParams{
		UserID: &memberID, BranchID: scoped, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, 0, access.MapErr(err)
	}
	s.attachLatestPayments(ctx, bookings)
	return bookings, total, nil
}

// Recent ใช้บนแดชบอร์ด (เรียกโดย module reporting)
func (s *Service) Recent(ctx context.Context, branchID *uuid.UUID, limit int) ([]Booking, error) {
	items, err := s.repo.RecentByBranch(ctx, branchID, limit)
	return items, access.MapErr(err)
}

// ---------------------------------------------------------------- helpers

// attachLatestPayments เติมสลิปล่าสุดให้แต่ละรายการ เพื่อให้ตารางแสดงปุ่ม "ดูสลิป" ได้
func (s *Service) attachLatestPayments(ctx context.Context, bookings []Booking) {
	for i := range bookings {
		if bookings[i].Status == BookingPendingPayment {
			continue
		}
		if p, err := s.repo.LatestPayment(ctx, bookings[i].ID); err == nil {
			bookings[i].LatestPayment = p
		}
	}
}

func (s *Service) record(ctx context.Context, identity middleware.Identity, action, entityID string, detail map[string]any, ip string) {
	actorID := identity.UserID
	s.audit.Record(ctx, audit.Entry{
		ActorID: &actorID, ActorRole: identity.Role, ActorName: identity.Name,
		BranchID: identity.BranchID, Action: action,
		EntityType: "booking", EntityID: entityID, Detail: detail, IPAddress: ip,
	})
}

func parseDate(v *validate.Validator, field, raw string, required bool) *time.Time {
	if raw == "" {
		if required {
			v.Check(false, field, "กรุณาเลือกวันที่")
		}
		return nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		v.Check(false, field, "รูปแบบวันที่ต้องเป็น YYYY-MM-DD")
		return nil
	}
	return &t
}

func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
