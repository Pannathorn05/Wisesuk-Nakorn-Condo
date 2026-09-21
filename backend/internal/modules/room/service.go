package room

import (
	"context"
	"time"

	"backend/internal/database"
	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/shared/access"
	"backend/internal/shared/audit"
	"backend/internal/shared/types"
	"backend/internal/validate"
)

type Service struct {
	repo  *Repository
	tx    *database.TxManager
	audit *audit.Recorder
}

func NewService(repo *Repository, tx *database.TxManager, rec *audit.Recorder) *Service {
	return &Service{repo: repo, tx: tx, audit: rec}
}

type SearchInput struct {
	BranchID   *types.BranchID
	RoomTypeID *types.RoomTypeID
	StayType   *types.StayType
	CheckIn    *time.Time
	CheckOut   *time.Time
	MoveInDate *time.Time
	MinPrice   *float64
	MaxPrice   *float64
	Limit      int
	Offset     int
}

// Search คือหน้า "ค้นหาห้องพัก" — คืนเฉพาะห้องที่จองได้จริงในช่วงที่ขอ
func (s *Service) Search(ctx context.Context, in SearchInput) ([]Room, int, error) {
	if in.StayType != nil && *in.StayType == types.StayDaily && in.CheckIn != nil && in.CheckOut != nil {
		if !in.CheckOut.After(*in.CheckIn) {
			return nil, 0, httpx.ValidationFailed(map[string]string{
				"check_out": "วันเช็คเอาท์ต้องอยู่หลังวันเข้าพัก",
			})
		}
	}

	rooms, total, err := s.repo.Search(ctx, SearchParams{
		BranchID: in.BranchID, RoomTypeID: in.RoomTypeID, StayType: in.StayType,
		CheckIn: in.CheckIn, CheckOut: in.CheckOut, MoveInDate: in.MoveInDate,
		MinPrice: in.MinPrice, MaxPrice: in.MaxPrice,
		OnlyBookable: true,
		Limit:        in.Limit, Offset: in.Offset,
	})
	return rooms, total, access.MapErr(err)
}

// Get คือหน้ารายละเอียดห้อง — เป็นที่เดียวที่คืนสิ่งอำนวยความสะดวกของห้องมาด้วย
func (s *Service) Get(ctx context.Context, id types.RoomID) (*Room, error) {
	rm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, access.MapErr(err)
	}
	amenities, err := s.repo.ListAmenities(ctx, id)
	if err != nil {
		return nil, access.MapErr(err)
	}
	rm.Amenities = amenities
	return rm, nil
}

func (s *Service) ListTypes(ctx context.Context, branchID *types.BranchID) ([]RoomType, error) {
	types, err := s.repo.ListTypes(ctx, branchID)
	return types, access.MapErr(err)
}

// Stats คืนตัวเลขรายสาขาบนแดชบอร์ด — branchID nil = ทุกสาขา
func (s *Service) Stats(ctx context.Context, branchID *types.BranchID) ([]BranchStats, error) {
	stats, err := s.repo.StatsByBranch(ctx, branchID)
	return stats, access.MapErr(err)
}

// ---------------------------------------------------------------- ใช้โดย module booking

// LockForBooking ล็อกห้องและคืนข้อมูลห้อง ต้องเรียกใน transaction ของผู้เรียก
func (s *Service) LockForBooking(ctx context.Context, roomID types.RoomID) (*Room, error) {
	rm, err := s.repo.Lock(ctx, roomID)
	return rm, access.MapErr(err)
}

func (s *Service) IsAvailable(ctx context.Context, roomID types.RoomID, checkIn, checkOut *time.Time) (bool, error) {
	ok, err := s.repo.IsAvailable(ctx, roomID, checkIn, checkOut)
	return ok, access.MapErr(err)
}

// MarkOccupied ใช้เมื่ออนุมัติการจองรายเดือน
func (s *Service) MarkOccupied(ctx context.Context, roomID types.RoomID, branchID types.BranchID) error {
	_, err := s.repo.UpdateStatus(ctx, roomID, branchID, RoomOccupied)
	return access.MapErr(err)
}

// ReleaseIfIdle เป็นคู่ตรงข้ามของ MarkOccupied — ใช้เมื่อการจองที่ถือห้องอยู่จบลง
// (ยกเลิก / หมดสัญญา) ห้องจะกลับมาปล่อยเช่าได้เฉพาะเมื่อไม่มีใบอื่นถืออยู่แล้ว
func (s *Service) ReleaseIfIdle(ctx context.Context, roomID types.RoomID, branchID types.BranchID) error {
	return access.MapErr(s.repo.ReleaseIfIdle(ctx, roomID, branchID))
}

// ---------------------------------------------------------------- admin

// ListForAdmin คืนห้องทุกสถานะในสาขาที่ผู้เรียกดูแล (หน้า "จัดการห้องพัก")
func (s *Service) ListForAdmin(ctx context.Context, identity middleware.Identity, in SearchInput) ([]Room, int, error) {
	branchID, err := access.Branch(identity, in.BranchID)
	if err != nil {
		return nil, 0, err
	}
	rooms, total, err := s.repo.Search(ctx, SearchParams{
		BranchID: branchID, RoomTypeID: in.RoomTypeID, StayType: in.StayType,
		MinPrice: in.MinPrice, MaxPrice: in.MaxPrice,
		OnlyBookable: false,
		Limit:        in.Limit, Offset: in.Offset,
	})
	return rooms, total, access.MapErr(err)
}

type SaveInput struct {
	BranchID     *types.BranchID   `json:"branch_id"`
	RoomTypeID   *types.RoomTypeID `json:"room_type_id"`
	RoomNumber   string            `json:"room_number"`
	Building     string            `json:"building"`
	Floor        int               `json:"floor"`
	StayType     string            `json:"stay_type"`
	Price        float64           `json:"price"`
	WaterRate    float64           `json:"water_rate"`
	ElectricRate float64           `json:"electric_rate"`
	SizeSqm      *float64          `json:"size_sqm"`
	Description  string            `json:"description"`
	ImageURL     *string           `json:"image_url"`
	Status       string            `json:"status"`
	// nil = ไม่แตะสิ่งอำนวยความสะดวกเดิม, [] = ล้างทิ้งทั้งหมด
	AmenityIDs *[]types.AmenityID `json:"amenity_ids"`
}

func (in SaveInput) validated() (SaveParams, error) {
	v := validate.New()
	roomNumber := v.Required("room_number", in.RoomNumber)
	v.MaxLen("room_number", roomNumber, 20)
	v.Positive("price", in.Price)
	v.Check(in.Floor > 0, "floor", "ชั้นต้องมากกว่า 0")
	v.Check(in.WaterRate >= 0, "water_rate", "ค่าน้ำต้องไม่ติดลบ")
	v.Check(in.ElectricRate >= 0, "electric_rate", "ค่าไฟต้องไม่ติดลบ")
	if in.ImageURL != nil {
		v.ImageURL("image_url", *in.ImageURL, false)
	}

	stayType := types.StayType(in.StayType)
	v.Check(stayType.Valid(), "stay_type", "ประเภทการเข้าพักต้องเป็น daily หรือ monthly")

	status := RoomStatus(in.Status)
	if in.Status == "" {
		status = RoomAvailable
	}
	v.Check(status.Valid(), "status", "สถานะห้องต้องเป็น available, occupied หรือ maintenance")

	building := in.Building
	if building == "" {
		building = "1"
	}

	if err := v.Err(); err != nil {
		return SaveParams{}, err
	}
	return SaveParams{
		RoomTypeID: in.RoomTypeID, RoomNumber: roomNumber, Building: building, Floor: in.Floor,
		StayType: stayType, Price: in.Price, WaterRate: in.WaterRate, ElectricRate: in.ElectricRate,
		SizeSqm: in.SizeSqm, Description: in.Description, ImageURL: in.ImageURL, Status: status,
		AmenityIDs: in.AmenityIDs,
	}, nil
}

var errDuplicateRoom = httpx.ValidationFailed(map[string]string{
	"room_number": "เลขห้องนี้มีอยู่ในสาขาแล้ว",
})

func (s *Service) Create(ctx context.Context, identity middleware.Identity, in SaveInput, ip string) (*Room, error) {
	branchID, err := access.RequireBranch(identity, in.BranchID)
	if err != nil {
		return nil, err
	}
	params, err := in.validated()
	if err != nil {
		return nil, err
	}
	params.BranchID = branchID

	var rm *Room
	// ห้องกับสิ่งอำนวยความสะดวกต้องลงพร้อมกัน ถ้าผูก amenity พลาดก็ต้องไม่เหลือห้องค้างไว้
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		created, err := s.repo.Create(ctx, params)
		if err != nil {
			return err
		}
		rm = created
		return s.saveAmenities(ctx, created.ID, params.AmenityIDs)
	})
	if err != nil {
		if database.IsUniqueViolation(err) {
			return nil, errDuplicateRoom
		}
		return nil, access.MapErr(err)
	}

	s.record(ctx, identity, branchID, "room.create", rm.ID.String(),
		map[string]any{"room_number": rm.RoomNumber}, ip)
	return rm, nil
}

func (s *Service) Update(ctx context.Context, identity middleware.Identity, roomID types.RoomID, in SaveInput, ip string) (*Room, error) {
	branchID, err := s.branchOf(ctx, identity, roomID)
	if err != nil {
		return nil, err
	}
	params, err := in.validated()
	if err != nil {
		return nil, err
	}
	params.BranchID = branchID

	var rm *Room
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		updated, err := s.repo.Update(ctx, roomID, branchID, params)
		if err != nil {
			return err
		}
		rm = updated
		return s.saveAmenities(ctx, roomID, params.AmenityIDs)
	})
	if err != nil {
		if database.IsUniqueViolation(err) {
			return nil, errDuplicateRoom
		}
		return nil, access.MapErr(err)
	}

	s.record(ctx, identity, branchID, "room.update", rm.ID.String(),
		map[string]any{"room_number": rm.RoomNumber}, ip)
	return rm, nil
}

// SetImage ใช้กับปุ่มอัปโหลดรูปห้อง ซึ่งเปลี่ยนรูปอย่างเดียวไม่ได้แก้ฟิลด์อื่น
func (s *Service) SetImage(ctx context.Context, identity middleware.Identity, roomID types.RoomID, url, ip string) (*Room, error) {
	branchID, err := s.branchOf(ctx, identity, roomID)
	if err != nil {
		return nil, err
	}

	v := validate.New()
	url = v.ImageURL("image_url", url, true)
	if err := v.Err(); err != nil {
		return nil, err
	}

	rm, err := s.repo.UpdateImage(ctx, roomID, branchID, url)
	if err != nil {
		return nil, access.MapErr(err)
	}

	s.record(ctx, identity, branchID, "room.update_image", rm.ID.String(),
		map[string]any{"room_number": rm.RoomNumber}, ip)
	return rm, nil
}

// UpdateStatus คือปุ่มอัปเดตสถานะห้องแบบ real-time ในหน้าจัดการห้องพัก
func (s *Service) UpdateStatus(ctx context.Context, identity middleware.Identity, roomID types.RoomID, status string, ip string) (*Room, error) {
	branchID, err := s.branchOf(ctx, identity, roomID)
	if err != nil {
		return nil, err
	}

	st := RoomStatus(status)
	if !st.Valid() {
		return nil, httpx.ValidationFailed(map[string]string{
			"status": "สถานะห้องต้องเป็น available, occupied หรือ maintenance",
		})
	}

	rm, err := s.repo.UpdateStatus(ctx, roomID, branchID, st)
	if err != nil {
		return nil, access.MapErr(err)
	}

	s.record(ctx, identity, branchID, "room.update_status", rm.ID.String(),
		map[string]any{"room_number": rm.RoomNumber, "status": string(st)}, ip)
	return rm, nil
}

func (s *Service) Delete(ctx context.Context, identity middleware.Identity, roomID types.RoomID, ip string) error {
	branchID, err := s.branchOf(ctx, identity, roomID)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, roomID, branchID); err != nil {
		return access.MapErr(err)
	}
	s.record(ctx, identity, branchID, "room.delete", roomID.String(), nil, ip)
	return nil
}

// saveAmenities แทนที่สิ่งอำนวยความสะดวกของห้องทั้งชุด
// ids เป็น nil แปลว่าคำขอไม่ได้พูดถึงเรื่องนี้ จึงไม่แตะของเดิม
func (s *Service) saveAmenities(ctx context.Context, roomID types.RoomID, ids *[]types.AmenityID) error {
	if ids == nil {
		return nil
	}
	err := s.repo.SetAmenities(ctx, roomID, *ids)
	if database.IsForeignKeyViolation(err) {
		return httpx.ValidationFailed(map[string]string{
			"amenity_ids": "มีสิ่งอำนวยความสะดวกที่ไม่มีอยู่ในระบบ",
		})
	}
	return err
}

// branchOf คืนสาขาของห้อง หลังยืนยันแล้วว่าผู้เรียกมีสิทธิ์กับสาขานั้น
func (s *Service) branchOf(ctx context.Context, identity middleware.Identity, roomID types.RoomID) (types.BranchID, error) {
	rm, err := s.repo.GetByID(ctx, roomID)
	if err != nil {
		return 0, access.MapErr(err)
	}
	return access.RequireBranch(identity, &rm.BranchID)
}

// record ผูก log กับสาขาของห้องที่ถูกกระทำ ไม่ใช่สาขาของผู้กระทำ
// (ผู้ดูแลหนึ่งคนดูแลได้หลายสาขา สาขาของผู้กระทำจึงไม่ได้บอกว่าลงมือที่ไหน)
func (s *Service) record(ctx context.Context, identity middleware.Identity, branchID types.BranchID,
	action, entityID string, detail map[string]any, ip string) {
	actorID := identity.UserID
	s.audit.Record(ctx, audit.Entry{
		ActorID: &actorID, ActorRole: identity.Role, ActorName: identity.Name,
		BranchID: &branchID, Action: action,
		EntityType: "room", EntityID: entityID, Detail: detail, IPAddress: ip,
	})
}
