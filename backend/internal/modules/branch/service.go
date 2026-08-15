package branch

import (
	"context"
	"strconv"
	"unicode/utf8"

	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/shared/access"
	"backend/internal/shared/audit"
	"backend/internal/validate"
)

type Service struct {
	repo  *Repository
	audit *audit.Recorder
}

func NewService(repo *Repository, rec *audit.Recorder) *Service {
	return &Service{repo: repo, audit: rec}
}

// List ใช้ทั้งหน้า "สาขาในเครือทั้งหมด" และหน้าติดต่อเรา
func (s *Service) List(ctx context.Context, includeInactive bool) ([]Branch, error) {
	branches, err := s.repo.List(ctx, includeInactive)
	if err != nil {
		return nil, access.MapErr(err)
	}
	// เติมรูป/สิ่งอำนวยความสะดวก/สถานที่ใกล้เคียง เพื่อให้หน้ารายการแสดง tag ได้ครบ
	for i := range branches {
		if err := s.repo.LoadRelations(ctx, &branches[i]); err != nil {
			return nil, access.MapErr(err)
		}
	}
	return branches, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Branch, error) {
	b, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, access.MapErr(err)
	}
	if err := s.repo.LoadRelations(ctx, b); err != nil {
		return nil, access.MapErr(err)
	}
	return b, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*Branch, error) {
	b, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, access.MapErr(err)
	}
	if err := s.repo.LoadRelations(ctx, b); err != nil {
		return nil, access.MapErr(err)
	}
	return b, nil
}

// Exists ทำให้ Service ใช้เป็น account.BranchChecker ได้
func (s *Service) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return s.repo.Exists(ctx, id)
}

// ContractFee คือค่ายืนยันการทำสัญญาที่ module booking ใช้คิดยอดของการจองรายเดือน
func (s *Service) ContractFee(ctx context.Context, branchID uuid.UUID) (float64, error) {
	b, err := s.repo.GetByID(ctx, branchID)
	if err != nil {
		return 0, access.MapErr(err)
	}
	return b.ContractFee, nil
}

func (s *Service) ListAmenities(ctx context.Context) ([]Amenity, error) {
	items, err := s.repo.ListAllAmenities(ctx)
	return items, access.MapErr(err)
}

// ---------------------------------------------------------------- admin

type UpdateInput struct {
	Name            string   `json:"name"`
	Tagline         string   `json:"tagline"`
	Description     string   `json:"description"`
	Address         string   `json:"address"`
	Phones          []string `json:"phones"`
	LineID          string   `json:"line_id"`
	Email           string   `json:"email"`
	Latitude        *float64 `json:"latitude"`
	Longitude       *float64 `json:"longitude"`
	MapURL          string   `json:"map_url"`
	BuildingCount   int      `json:"building_count"`
	FloorCount      int      `json:"floor_count"`
	DailyPriceFrom  *float64 `json:"daily_price_from"`
	MonthlyPriceMin *float64 `json:"monthly_price_min"`
	MonthlyPriceMax *float64 `json:"monthly_price_max"`
	WaterRate       float64  `json:"water_rate"`
	ElectricRate    float64  `json:"electric_rate"`
	Deposit         float64  `json:"deposit"`
	AdvancePayment  float64  `json:"advance_payment"`
	ContractFee     float64  `json:"contract_fee"`
	CoverImageURL   *string  `json:"cover_image_url"`
}

func (s *Service) Update(ctx context.Context, identity middleware.Identity, branchID *uuid.UUID, in UpdateInput, ip string) (*Branch, error) {
	id, err := access.RequireBranch(identity, branchID)
	if err != nil {
		return nil, err
	}

	v := validate.New()
	name := v.Required("name", in.Name)
	v.MaxLen("name", name, 200)
	v.Check(in.BuildingCount > 0, "building_count", "จำนวนอาคารต้องมากกว่า 0")
	v.Check(in.FloorCount > 0, "floor_count", "จำนวนชั้นต้องมากกว่า 0")
	if in.CoverImageURL != nil {
		v.ImageURL("cover_image_url", *in.CoverImageURL, false)
	}
	v.Check(in.WaterRate >= 0, "water_rate", "ค่าน้ำต้องไม่ติดลบ")
	v.Check(in.ElectricRate >= 0, "electric_rate", "ค่าไฟต้องไม่ติดลบ")
	if in.MonthlyPriceMin != nil && in.MonthlyPriceMax != nil {
		v.Check(*in.MonthlyPriceMin <= *in.MonthlyPriceMax,
			"monthly_price_max", "ราคาสูงสุดต้องไม่น้อยกว่าราคาต่ำสุด")
	}
	if err := v.Err(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, id, UpdateParams{
		Name: name, Tagline: in.Tagline, Description: in.Description, Address: in.Address,
		Phones: in.Phones, LineID: in.LineID, Email: in.Email,
		Latitude: in.Latitude, Longitude: in.Longitude, MapURL: in.MapURL,
		BuildingCount: in.BuildingCount, FloorCount: in.FloorCount,
		DailyPriceFrom: in.DailyPriceFrom, MonthlyPriceMin: in.MonthlyPriceMin, MonthlyPriceMax: in.MonthlyPriceMax,
		WaterRate: in.WaterRate, ElectricRate: in.ElectricRate, Deposit: in.Deposit,
		AdvancePayment: in.AdvancePayment, ContractFee: in.ContractFee,
		CoverImageURL: in.CoverImageURL,
	}); err != nil {
		return nil, access.MapErr(err)
	}

	s.record(ctx, identity, "branch.update", "branch", id.String(), map[string]any{"name": name}, ip)
	return s.Get(ctx, id)
}

func (s *Service) SetAmenities(ctx context.Context, identity middleware.Identity, branchID *uuid.UUID, amenityIDs []uuid.UUID, ip string) ([]Amenity, error) {
	id, err := access.RequireBranch(identity, branchID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SetAmenities(ctx, id, amenityIDs); err != nil {
		return nil, access.MapErr(err)
	}
	s.record(ctx, identity, "branch.set_amenities", "branch", id.String(),
		map[string]any{"count": len(amenityIDs)}, ip)

	items, err := s.repo.ListAmenities(ctx, id)
	return items, access.MapErr(err)
}

func (s *Service) ReplaceNearby(ctx context.Context, identity middleware.Identity, branchID *uuid.UUID, items []NearbyInput, ip string) ([]NearbyPlace, error) {
	id, err := access.RequireBranch(identity, branchID)
	if err != nil {
		return nil, err
	}

	v := validate.New()
	for i, it := range items {
		if it.Name == "" {
			v.Check(false, "items", "รายการที่ "+strconv.Itoa(i+1)+" ยังไม่ได้กรอกชื่อสถานที่")
			break
		}
	}
	if err := v.Err(); err != nil {
		return nil, err
	}

	if err := s.repo.ReplaceNearby(ctx, id, items); err != nil {
		return nil, access.MapErr(err)
	}
	s.record(ctx, identity, "branch.update_nearby", "branch", id.String(),
		map[string]any{"count": len(items)}, ip)

	places, err := s.repo.ListNearby(ctx, id)
	return places, access.MapErr(err)
}

// SetCoverImage ใช้กับปุ่มเปลี่ยนรูปปกสาขา ซึ่งอัปโหลดรูปอย่างเดียวไม่ได้แก้ฟิลด์อื่น
func (s *Service) SetCoverImage(ctx context.Context, identity middleware.Identity, branchID *uuid.UUID, url, ip string) (*Branch, error) {
	id, err := access.RequireBranch(identity, branchID)
	if err != nil {
		return nil, err
	}
	v := validate.New()
	url = v.ImageURL("cover_image_url", url, true)
	if err := v.Err(); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateCoverImage(ctx, id, url); err != nil {
		return nil, access.MapErr(err)
	}
	s.record(ctx, identity, "branch.update_cover", "branch", id.String(), nil, ip)
	return s.Get(ctx, id)
}

func (s *Service) AddImage(ctx context.Context, identity middleware.Identity, branchID *uuid.UUID, url, caption string, sortOrder int, ip string) (*BranchImage, error) {
	id, err := access.RequireBranch(identity, branchID)
	if err != nil {
		return nil, err
	}
	v := validate.New()
	url = v.ImageURL("image_url", url, true)
	v.MaxLen("caption", caption, 200)
	// ฟิลด์ข้อความใน multipart ไม่ได้ผ่าน decoder ของ JSON จึงยังเป็นไบต์ดิบ
	// ถ้าไม่ใช่ UTF-8 (เช่น client ส่งมาเป็น CP874) ต้องบอกให้รู้ ไม่ใช่ปล่อยให้ INSERT พังเป็น 500
	v.Check(utf8.ValidString(caption), "caption", "คำบรรยายต้องเป็นข้อความ UTF-8")
	if err := v.Err(); err != nil {
		return nil, err
	}

	img, err := s.repo.AddImage(ctx, id, url, caption, sortOrder)
	if err != nil {
		return nil, access.MapErr(err)
	}
	s.record(ctx, identity, "branch.add_image", "branch_image", img.ID.String(), nil, ip)
	return img, nil
}

func (s *Service) DeleteImage(ctx context.Context, identity middleware.Identity, branchID *uuid.UUID, imageID uuid.UUID, ip string) error {
	id, err := access.RequireBranch(identity, branchID)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteImage(ctx, id, imageID); err != nil {
		return access.MapErr(err)
	}
	s.record(ctx, identity, "branch.delete_image", "branch_image", imageID.String(), nil, ip)
	return nil
}

func (s *Service) record(ctx context.Context, identity middleware.Identity, action, entityType, entityID string, detail map[string]any, ip string) {
	actorID := identity.UserID
	s.audit.Record(ctx, audit.Entry{
		ActorID: &actorID, ActorRole: identity.Role, ActorName: identity.Name,
		BranchID: identity.BranchID, Action: action,
		EntityType: entityType, EntityID: entityID, Detail: detail, IPAddress: ip,
	})
}
