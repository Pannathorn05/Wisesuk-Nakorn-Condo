package room

import (
	"time"

	"backend/internal/shared/types"
)

// RoomStatus คือสถานะห้องที่แอดมินอัปเดตแบบ real-time
type RoomStatus string

const (
	RoomAvailable   RoomStatus = "available"   // ว่าง
	RoomOccupied    RoomStatus = "occupied"    // มีผู้เช่า
	RoomMaintenance RoomStatus = "maintenance" // ปิดปรับปรุง
)

func (s RoomStatus) Valid() bool {
	switch s {
	case RoomAvailable, RoomOccupied, RoomMaintenance:
		return true
	}
	return false
}

// RoomType คือประเภทห้องของแต่ละสาขา เช่น ห้องแอร์ ห้องพัดลม ห้องเปล่า
type RoomType struct {
	ID          types.RoomTypeID `json:"id"`
	BranchID    types.BranchID   `json:"branch_id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	SizeSqm     *float64         `json:"size_sqm,omitempty"`
	ImageURL    string           `json:"image_url"`
	SortOrder   int              `json:"sort_order"`
}

// Amenity คือสิ่งอำนวยความสะดวกของห้อง เช่น แอร์ ห้องน้ำในตัว คีย์การ์ด
// นิยามซ้ำกับ module branch โดยตั้งใจ เพื่อให้ room ไม่ต้อง import branch
type Amenity struct {
	ID        types.AmenityID `json:"id"`
	Code      string          `json:"code"`
	Name      string          `json:"name"`
	Icon      string          `json:"icon"`
	SortOrder int             `json:"sort_order"`
}

type Room struct {
	ID           types.RoomID      `json:"id"`
	BranchID     types.BranchID    `json:"branch_id"`
	BranchName   string            `json:"branch_name,omitempty"`
	RoomTypeID   *types.RoomTypeID `json:"room_type_id,omitempty"`
	RoomTypeName string            `json:"room_type_name,omitempty"`
	RoomNumber   string            `json:"room_number"`
	Building     string            `json:"building"`
	Floor        int               `json:"floor"`
	StayType     types.StayType    `json:"stay_type"`
	Price        float64           `json:"price"`
	WaterRate    float64           `json:"water_rate"`
	ElectricRate float64           `json:"electric_rate"`
	SizeSqm      *float64          `json:"size_sqm,omitempty"`
	Description  string            `json:"description"`
	ImageURL     string            `json:"image_url"`
	Status       RoomStatus        `json:"status"`
	IsActive     bool              `json:"is_active"`
	Amenities    []Amenity         `json:"amenities,omitempty"` // เติมเฉพาะตอนดึงห้องรายตัว
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// BranchStats คือตัวเลขห้องว่างที่แสดงบนแดชบอร์ด
type BranchStats struct {
	BranchID          types.BranchID `json:"branch_id"`
	BranchName        string         `json:"branch_name"`
	DailyRoomsFree    int            `json:"daily_rooms_free"`
	DailyRoomsTotal   int            `json:"daily_rooms_total"`
	MonthlyRoomsFree  int            `json:"monthly_rooms_free"`
	MonthlyRoomsTotal int            `json:"monthly_rooms_total"`
	PendingReview     int            `json:"pending_review"`
	BookingsTotal     int            `json:"bookings_total"`
}
