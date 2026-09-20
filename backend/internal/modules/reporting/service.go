package reporting

import (
	"context"

	"backend/internal/middleware"
	"backend/internal/modules/booking"
	"backend/internal/modules/room"
	"backend/internal/shared/access"
	"backend/internal/shared/audit"
	"backend/internal/shared/types"
)

type Service struct {
	repo     *Repository
	rooms    *room.Service
	bookings *booking.Service
}

func NewService(repo *Repository, rooms *room.Service, bookings *booking.Service) *Service {
	return &Service{repo: repo, rooms: rooms, bookings: bookings}
}

// dashboardActivities คือจำนวนกิจกรรมล่าสุดที่แสดงบนแดชบอร์ด
const dashboardActivities = 10

// Dashboard — admin เห็นเฉพาะสาขาที่ตนดูแล, superadmin เห็นทุกสาขา
func (s *Service) Dashboard(ctx context.Context, identity middleware.Identity) (*Dashboard, error) {
	branchID, err := access.Branch(identity, nil)
	if err != nil {
		return nil, err
	}

	stats, err := s.rooms.Stats(ctx, branchID)
	if err != nil {
		return nil, err
	}
	recent, _, err := s.repo.ListLogs(ctx, ListParams{BranchID: branchID, Limit: dashboardActivities})
	if err != nil {
		return nil, access.MapErr(err)
	}
	return &Dashboard{Branches: stats, RecentActivities: recent}, nil
}

type ActivityFilter struct {
	ActorID   *types.UserID
	ActorRole *types.Role
	BranchID  *types.BranchID
	Action    string
	Search    string
	Limit     int
	Offset    int
}

func (s *Service) ActivityLogs(ctx context.Context, identity middleware.Identity, f ActivityFilter) ([]audit.Log, int, error) {
	branchID, err := access.Branch(identity, f.BranchID)
	if err != nil {
		return nil, 0, err
	}

	logs, total, err := s.repo.ListLogs(ctx, ListParams{
		ActorID:   f.ActorID,
		ActorRole: f.ActorRole,
		BranchID:  branchID,
		Action:    f.Action,
		Search:    f.Search,
		Limit:     f.Limit,
		Offset:    f.Offset,
	})
	return logs, total, access.MapErr(err)
}
