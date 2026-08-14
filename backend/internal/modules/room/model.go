package room

import (
	"time"

	"github.com/google/uuid"

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
	ID          uuid.UUID `json:"id"`
	BranchID    uuid.UUID `json:"branch_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SizeSqm     *float64  `json:"size_sqm,omitempty"`
	ImageURL    string    `json:"image_url"`
	SortOrder   int       `json:"sort_order"`
}

type Room struct {
	ID           uuid.UUID      `json:"id"`
	BranchID     uuid.UUID      `json:"branch_id"`
	BranchName   string         `json:"branch_name,omitempty"`
	RoomTypeID   *uuid.UUID     `json:"room_type_id,omitempty"`
	RoomTypeName string         `json:"room_type_name,omitempty"`
	RoomNumber   string         `json:"room_number"`
	Building     string         `json:"building"`
	Floor        int            `json:"floor"`
	StayType     types.StayType `json:"stay_type"`
	Price        float64        `json:"price"`
	WaterRate    float64        `json:"water_rate"`
	ElectricRate float64        `json:"electric_rate"`
	SizeSqm      *float64       `json:"size_sqm,omitempty"`
	Description  string         `json:"description"`
	ImageURL     string         `json:"image_url"`
	Status       RoomStatus     `json:"status"`
	IsActive     bool           `json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// BranchStats คือตัวเลขห้องว่างที่แสดงบนแดชบอร์ด
type BranchStats struct {
	BranchID          uuid.UUID `json:"branch_id"`
	BranchName        string    `json:"branch_name"`
	DailyRoomsFree    int       `json:"daily_rooms_free"`
	DailyRoomsTotal   int       `json:"daily_rooms_total"`
	MonthlyRoomsFree  int       `json:"monthly_rooms_free"`
	MonthlyRoomsTotal int       `json:"monthly_rooms_total"`
	PendingReview     int       `json:"pending_review"`
	BookingsTotal     int       `json:"bookings_total"`
}
