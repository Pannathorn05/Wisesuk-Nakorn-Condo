package branch

import (
	"time"

	"github.com/google/uuid"
)

// Branch คือหอพักหนึ่งสาขาในเครือ
type Branch struct {
	ID              uuid.UUID `json:"id"`
	Slug            string    `json:"slug"`
	Name            string    `json:"name"`
	Tagline         string    `json:"tagline"`
	Description     string    `json:"description"`
	Address         string    `json:"address"`
	Phones          []string  `json:"phones"`
	LineID          string    `json:"line_id"`
	Email           string    `json:"email"`
	Latitude        *float64  `json:"latitude,omitempty"`
	Longitude       *float64  `json:"longitude,omitempty"`
	MapURL          string    `json:"map_url"`
	BuildingCount   int       `json:"building_count"`
	FloorCount      int       `json:"floor_count"`
	DailyPriceFrom  *float64  `json:"daily_price_from,omitempty"`
	MonthlyPriceMin *float64  `json:"monthly_price_min,omitempty"`
	MonthlyPriceMax *float64  `json:"monthly_price_max,omitempty"`
	WaterRate       float64   `json:"water_rate"`    // บาท/หน่วย
	ElectricRate    float64   `json:"electric_rate"` // บาท/หน่วย
	Deposit         float64   `json:"deposit"`
	AdvancePayment  float64   `json:"advance_payment"`
	ContractFee     float64   `json:"contract_fee"` // ค่ายืนยันการทำสัญญา (รายเดือน)
	CoverImageURL   string    `json:"cover_image_url"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	Images    []BranchImage `json:"images,omitempty"`
	Amenities []Amenity     `json:"amenities,omitempty"`
	Nearby    []NearbyPlace `json:"nearby_places,omitempty"`
}

type BranchImage struct {
	ID        uuid.UUID `json:"id"`
	BranchID  uuid.UUID `json:"branch_id"`
	ImageURL  string    `json:"image_url"`
	Caption   string    `json:"caption"`
	SortOrder int       `json:"sort_order"`
}

// Amenity คือสิ่งอำนวยความสะดวก เช่น แอร์ ไวไฟ เครื่องทำน้ำอุ่น
type Amenity struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	SortOrder int       `json:"sort_order"`
}

// NearbyPlace คือสถานที่ใกล้เคียง (สถานศึกษา / ห้างสรรพสินค้า / โรงพยาบาล ...)
type NearbyPlace struct {
	ID        uuid.UUID `json:"id"`
	BranchID  uuid.UUID `json:"branch_id"`
	Category  string    `json:"category"`
	Name      string    `json:"name"`
	Distance  string    `json:"distance"`
	SortOrder int       `json:"sort_order"`
}
