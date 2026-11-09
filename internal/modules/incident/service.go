package incident

import (
	"context"
	"time"

	"go-up-monitoring/internal/modules/monitor"

	"github.com/google/uuid"
)

type AlertNotifier interface {
	NotifyDown(ctx context.Context, inc *Incident, m *monitor.Monitor) error
	NotifyResolved(ctx context.Context, inc *Incident, m *monitor.Monitor) error
}

type Service interface {
	HandleProbeResult(ctx context.Context, m *monitor.Monitor, isUp bool, errMsg string) error
	GetIncidentsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Incident, int64, error)
	AcknowledgeIncident(ctx context.Context, id, userID uuid.UUID) error
}

type service struct {
	repo     Repository
	notifier AlertNotifier
}

func NewService(repo Repository, notifier AlertNotifier) Service {
	return &service{
		repo:     repo,
		notifier: notifier,
	}
}

func (s *service) HandleProbeResult(ctx context.Context, m *monitor.Monitor, isUp bool, errMsg string) error {
	activeInc, err := s.repo.FindActiveByMonitorID(ctx, m.ID)
	if err != nil {
		return err
	}

	now := time.Now()

	if !isUp {
		// Target is failing
		// Check if threshold is reached AND no active incident exists yet
		if m.ConsecutiveFails >= m.RetryThreshold && activeInc == nil {
			newInc := &Incident{
				MonitorID:   m.ID,
				UserID:      m.UserID,
				Status:      IncidentStatusTriggered,
				ErrorReason: errMsg,
				StartedAt:   now,
			}

			if err := s.repo.Create(ctx, newInc); err != nil {
				return err
			}

			if s.notifier != nil {
				_ = s.notifier.NotifyDown(ctx, newInc, m)
			}
		}
	} else {
		// Target is healthy
		// If there is an active open incident, resolve it
		if activeInc != nil {
			durationSec := int64(now.Sub(activeInc.StartedAt).Seconds())
			if durationSec < 0 {
				durationSec = 0
			}

			if err := s.repo.Resolve(ctx, activeInc.ID, now, durationSec); err != nil {
				return err
			}

			activeInc.ResolvedAt = &now
			activeInc.DurationSec = durationSec
			activeInc.Status = IncidentStatusResolved

			if s.notifier != nil {
				_ = s.notifier.NotifyResolved(ctx, activeInc, m)
			}
		}
	}

	return nil
}

func (s *service) GetIncidentsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Incident, int64, error) {
	return s.repo.FindByUserID(ctx, userID, limit, offset)
}

func (s *service) AcknowledgeIncident(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.Acknowledge(ctx, id, userID, time.Now())
}
