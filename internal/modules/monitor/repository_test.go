package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockMonitorRepository struct {
	data map[uuid.UUID]*Monitor
}

func newMockMonitorRepo() *mockMonitorRepository {
	return &mockMonitorRepository{data: make(map[uuid.UUID]*Monitor)}
}

func (m *mockMonitorRepository) Create(ctx context.Context, mon *Monitor) error {
	if mon.ID == uuid.Nil {
		mon.ID = uuid.New()
	}
	m.data[mon.ID] = mon
	return nil
}

func (m *mockMonitorRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*Monitor, error) {
	mon, exists := m.data[id]
	if !exists || mon.UserID != userID {
		return nil, ErrMonitorNotFound
	}
	return mon, nil
}

func (m *mockMonitorRepository) FindByIDInternal(ctx context.Context, id uuid.UUID) (*Monitor, error) {
	mon, exists := m.data[id]
	if !exists {
		return nil, ErrMonitorNotFound
	}
	return mon, nil
}

func (m *mockMonitorRepository) FindAllByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Monitor, int64, error) {
	var list []Monitor
	for _, mon := range m.data {
		if mon.UserID == userID {
			list = append(list, *mon)
		}
	}
	return list, int64(len(list)), nil
}

func (m *mockMonitorRepository) FindActiveMonitors(ctx context.Context) ([]Monitor, error) {
	var list []Monitor
	for _, mon := range m.data {
		if mon.IsActive {
			list = append(list, *mon)
		}
	}
	return list, nil
}

func (m *mockMonitorRepository) Update(ctx context.Context, mon *Monitor) error {
	existing, exists := m.data[mon.ID]
	if !exists || existing.UserID != mon.UserID {
		return ErrMonitorNotFound
	}
	m.data[mon.ID] = mon
	return nil
}

func (m *mockMonitorRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	existing, exists := m.data[id]
	if !exists || existing.UserID != userID {
		return ErrMonitorNotFound
	}
	delete(m.data, id)
	return nil
}

func (m *mockMonitorRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status Status, consecutiveFails int, lastChecked time.Time) error {
	existing, exists := m.data[id]
	if !exists {
		return ErrMonitorNotFound
	}
	existing.Status = status
	existing.ConsecutiveFails = consecutiveFails
	existing.LastCheckedAt = &lastChecked
	return nil
}

func TestRepository_TenantIsolation(t *testing.T) {
	repo := newMockMonitorRepo()
	ctx := context.Background()

	userA := uuid.New()
	userB := uuid.New()

	monA := &Monitor{
		UserID:   userA,
		Name:     "User A Website",
		URL:      "https://usera.com",
		Protocol: ProtocolHTTPS,
		IsActive: true,
	}
	err := repo.Create(ctx, monA)
	assert.NoError(t, err)

	// User A can find their monitor
	foundA, err := repo.FindByID(ctx, monA.ID, userA)
	assert.NoError(t, err)
	assert.Equal(t, "User A Website", foundA.Name)

	// User B CANNOT find User A's monitor (tenant isolation!)
	_, err = repo.FindByID(ctx, monA.ID, userB)
	assert.ErrorIs(t, err, ErrMonitorNotFound)

	// User B CANNOT delete User A's monitor
	err = repo.Delete(ctx, monA.ID, userB)
	assert.ErrorIs(t, err, ErrMonitorNotFound)

	// User A CAN delete their own monitor
	err = repo.Delete(ctx, monA.ID, userA)
	assert.NoError(t, err)

	// It is now deleted
	_, err = repo.FindByID(ctx, monA.ID, userA)
	assert.ErrorIs(t, err, ErrMonitorNotFound)
}
