package reporting

import (
	"context"

	"github.com/google/uuid"

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

// Dashboard — admin เห็นเฉพาะสาขาตัวเอง, superadmin เห็นทุกสาขา
func (s *Service) Dashboard(ctx context.Context, identity middleware.Identity) (*Dashboard, error) {
	branchID, err := access.Branch(identity, nil)
	if err != nil {
		return nil, err
	}

	stats, err := s.rooms.Stats(ctx, branchID)
	if err != nil {
		return nil, err
	}
	recent, err := s.bookings.Recent(ctx, branchID, 10)
	if err != nil {
		return nil, err
	}
	return &Dashboard{Branches: stats, Recent: recent}, nil
}

type ActivityFilter struct {
	ActorID   *uuid.UUID
	ActorRole *types.Role
	BranchID  *uuid.UUID
	Action    string
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
		Limit:     f.Limit,
		Offset:    f.Offset,
	})
	return logs, total, access.MapErr(err)
}
