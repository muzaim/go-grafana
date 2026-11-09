package incident

import (
	"context"
	"testing"
	"time"

	"go-up-monitoring/internal/modules/monitor"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockIncidentRepo struct {
	data map[uuid.UUID]*Incident
}

func newMockIncidentRepo() *mockIncidentRepo {
	return &mockIncidentRepo{data: make(map[uuid.UUID]*Incident)}
}

func (m *mockIncidentRepo) Create(ctx context.Context, inc *Incident) error {
	if inc.ID == uuid.Nil {
		inc.ID = uuid.New()
	}
	m.data[inc.ID] = inc
	return nil
}

func (m *mockIncidentRepo) FindActiveByMonitorID(ctx context.Context, monitorID uuid.UUID) (*Incident, error) {
	for _, inc := range m.data {
		if inc.MonitorID == monitorID && inc.Status != IncidentStatusResolved {
			return inc, nil
		}
	}
	return nil, nil
}

func (m *mockIncidentRepo) Resolve(ctx context.Context, id uuid.UUID, resolvedAt time.Time, durationSec int64) error {
	inc, exists := m.data[id]
	if !exists {
		return ErrIncidentNotFound
	}
	inc.Status = IncidentStatusResolved
	inc.ResolvedAt = &resolvedAt
	inc.DurationSec = durationSec
	return nil
}

func (m *mockIncidentRepo) Acknowledge(ctx context.Context, id, userID uuid.UUID, ackAt time.Time) error {
	inc, exists := m.data[id]
	if !exists || inc.UserID != userID {
		return ErrIncidentNotFound
	}
	inc.Status = IncidentStatusAcknowledged
	inc.AcknowledgedAt = &ackAt
	return nil
}

func (m *mockIncidentRepo) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Incident, int64, error) {
	var list []Incident
	for _, inc := range m.data {
		if inc.UserID == userID {
			list = append(list, *inc)
		}
	}
	return list, int64(len(list)), nil
}

type mockNotifier struct {
	downCalled     bool
	resolvedCalled bool
}

func (n *mockNotifier) NotifyDown(ctx context.Context, inc *Incident, m *monitor.Monitor) error {
	n.downCalled = true
	return nil
}

func (n *mockNotifier) NotifyResolved(ctx context.Context, inc *Incident, m *monitor.Monitor) error {
	n.resolvedCalled = true
	return nil
}

func TestIncidentService_StateTransitions(t *testing.T) {
	repo := newMockIncidentRepo()
	notifier := &mockNotifier{}
	svc := NewService(repo, notifier)
	ctx := context.Background()

	mon := &monitor.Monitor{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		Name:             "Production App",
		RetryThreshold:   3,
		ConsecutiveFails: 1, // Below threshold
	}

	// 1. Probe fails 1st time -> Below retry_threshold (no incident yet)
	err := svc.HandleProbeResult(ctx, mon, false, "timeout")
	assert.NoError(t, err)
	assert.False(t, notifier.downCalled)

	// 2. Probe fails reaching threshold (3 fails) -> Incident created & alert sent!
	mon.ConsecutiveFails = 3
	err = svc.HandleProbeResult(ctx, mon, false, "timeout 3x")
	assert.NoError(t, err)
	assert.True(t, notifier.downCalled)

	// Verify active incident exists
	active, err := repo.FindActiveByMonitorID(ctx, mon.ID)
	assert.NoError(t, err)
	assert.NotNil(t, active)
	assert.Equal(t, IncidentStatusTriggered, active.Status)

	// 3. Probe fails again -> Active incident already exists, no duplicate created
	notifier.downCalled = false
	mon.ConsecutiveFails = 4
	err = svc.HandleProbeResult(ctx, mon, false, "timeout 4x")
	assert.NoError(t, err)
	assert.False(t, notifier.downCalled, "Should not trigger another alert while active incident exists")

	// 4. Target recovers -> Resolve incident & notify
	err = svc.HandleProbeResult(ctx, mon, true, "")
	assert.NoError(t, err)
	assert.True(t, notifier.resolvedCalled)

	// Verify incident is now resolved
	activeAfter, err := repo.FindActiveByMonitorID(ctx, mon.ID)
	assert.NoError(t, err)
	assert.Nil(t, activeAfter)
}
