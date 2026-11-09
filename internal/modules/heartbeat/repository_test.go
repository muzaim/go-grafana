package heartbeat

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockHeartbeatRepo struct {
	items []Heartbeat
}

func newMockHeartbeatRepo() *mockHeartbeatRepo {
	return &mockHeartbeatRepo{items: []Heartbeat{}}
}

func (m *mockHeartbeatRepo) Create(ctx context.Context, h *Heartbeat) error {
	h.ID = uint64(len(m.items) + 1)
	m.items = append(m.items, *h)
	return nil
}

func (m *mockHeartbeatRepo) FindByMonitorID(ctx context.Context, monitorID uuid.UUID, limit, offset int) ([]Heartbeat, int64, error) {
	var list []Heartbeat
	for _, it := range m.items {
		if it.MonitorID == monitorID {
			list = append(list, it)
		}
	}
	return list, int64(len(list)), nil
}

func (m *mockHeartbeatRepo) GetStats(ctx context.Context, monitorID uuid.UUID, since time.Time) (*StatsResult, error) {
	var total, up int64
	var totalLatency int64

	for _, it := range m.items {
		if it.MonitorID == monitorID && it.CheckedAt.After(since) {
			total++
			if it.IsUp {
				up++
			}
			totalLatency += it.LatencyMs
		}
	}

	uptime := 0.0
	avgLat := 0.0
	if total > 0 {
		uptime = (float64(up) / float64(total)) * 100.0
		avgLat = float64(totalLatency) / float64(total)
	}

	return &StatsResult{
		TotalChecks:   total,
		UpChecks:      up,
		UptimePercent: uptime,
		AvgLatencyMs:  avgLat,
	}, nil
}

func (m *mockHeartbeatRepo) PruneOlderThan(ctx context.Context, threshold time.Time) (int64, error) {
	var kept []Heartbeat
	var pruned int64
	for _, it := range m.items {
		if it.CheckedAt.Before(threshold) {
			pruned++
		} else {
			kept = append(kept, it)
		}
	}
	m.items = kept
	return pruned, nil
}

func TestHeartbeatRepository_StatsAndPruning(t *testing.T) {
	repo := newMockHeartbeatRepo()
	ctx := context.Background()
	monitorID := uuid.New()
	now := time.Now()

	// Add 3 UP and 1 DOWN heartbeats
	_ = repo.Create(ctx, &Heartbeat{MonitorID: monitorID, IsUp: true, LatencyMs: 100, CheckedAt: now.Add(-10 * time.Minute)})
	_ = repo.Create(ctx, &Heartbeat{MonitorID: monitorID, IsUp: true, LatencyMs: 200, CheckedAt: now.Add(-5 * time.Minute)})
	_ = repo.Create(ctx, &Heartbeat{MonitorID: monitorID, IsUp: true, LatencyMs: 300, CheckedAt: now.Add(-2 * time.Minute)})
	_ = repo.Create(ctx, &Heartbeat{MonitorID: monitorID, IsUp: false, LatencyMs: 0, CheckedAt: now.Add(-1 * time.Minute)})

	stats, err := repo.GetStats(ctx, monitorID, now.Add(-1*time.Hour))
	assert.NoError(t, err)
	assert.Equal(t, int64(4), stats.TotalChecks)
	assert.Equal(t, int64(3), stats.UpChecks)
	assert.Equal(t, 75.0, stats.UptimePercent)
	assert.Equal(t, 150.0, stats.AvgLatencyMs)

	// Prune older than 3 minutes
	pruned, err := repo.PruneOlderThan(ctx, now.Add(-3*time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, int64(2), pruned)

	list, count, err := repo.FindByMonitorID(ctx, monitorID, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
	assert.Len(t, list, 2)
}
