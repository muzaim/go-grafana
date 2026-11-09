package heartbeat

import (
	"context"
	"time"

	"go-up-monitoring/internal/modules/monitor"

	"github.com/google/uuid"
)

type Service interface {
	GetHeartbeats(ctx context.Context, monitorID, userID uuid.UUID, limit, offset int) ([]Heartbeat, int64, error)
	GetStats(ctx context.Context, monitorID, userID uuid.UUID, period string) (*StatsResult, error)
}

type service struct {
	hbRepo      Repository
	monitorRepo monitor.Repository
}

func NewService(hbRepo Repository, monitorRepo monitor.Repository) Service {
	return &service{
		hbRepo:      hbRepo,
		monitorRepo: monitorRepo,
	}
}

func (s *service) GetHeartbeats(ctx context.Context, monitorID, userID uuid.UUID, limit, offset int) ([]Heartbeat, int64, error) {
	// Tenant check: ensure monitor belongs to user
	_, err := s.monitorRepo.FindByID(ctx, monitorID, userID)
	if err != nil {
		return nil, 0, err
	}

	return s.hbRepo.FindByMonitorID(ctx, monitorID, limit, offset)
}

func (s *service) GetStats(ctx context.Context, monitorID, userID uuid.UUID, period string) (*StatsResult, error) {
	// Tenant check: ensure monitor belongs to user
	_, err := s.monitorRepo.FindByID(ctx, monitorID, userID)
	if err != nil {
		return nil, err
	}

	duration := 24 * time.Hour
	switch period {
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	}

	since := time.Now().Add(-duration)
	return s.hbRepo.GetStats(ctx, monitorID, since)
}
