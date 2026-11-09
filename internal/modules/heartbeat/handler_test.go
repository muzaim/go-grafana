package heartbeat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-up-monitoring/internal/modules/monitor"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockMonitorRepoForHB struct {
	monitors map[uuid.UUID]*monitor.Monitor
}

func (m *mockMonitorRepoForHB) Create(ctx context.Context, mon *monitor.Monitor) error {
	m.monitors[mon.ID] = mon
	return nil
}
func (m *mockMonitorRepoForHB) FindByID(ctx context.Context, id, userID uuid.UUID) (*monitor.Monitor, error) {
	mon, exists := m.monitors[id]
	if !exists || mon.UserID != userID {
		return nil, monitor.ErrMonitorNotFound
	}
	return mon, nil
}
func (m *mockMonitorRepoForHB) FindByIDInternal(ctx context.Context, id uuid.UUID) (*monitor.Monitor, error) {
	mon, exists := m.monitors[id]
	if !exists {
		return nil, monitor.ErrMonitorNotFound
	}
	return mon, nil
}
func (m *mockMonitorRepoForHB) FindAllByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]monitor.Monitor, int64, error) {
	return nil, 0, nil
}
func (m *mockMonitorRepoForHB) FindActiveMonitors(ctx context.Context) ([]monitor.Monitor, error) {
	return nil, nil
}
func (m *mockMonitorRepoForHB) Update(ctx context.Context, mon *monitor.Monitor) error {
	return nil
}
func (m *mockMonitorRepoForHB) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return nil
}
func (m *mockMonitorRepoForHB) UpdateStatus(ctx context.Context, id uuid.UUID, status monitor.Status, consecutiveFails int, lastChecked time.Time) error {
	return nil
}

func TestHeartbeatHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	userID := uuid.New()
	monitorID := uuid.New()

	monRepo := &mockMonitorRepoForHB{
		monitors: map[uuid.UUID]*monitor.Monitor{
			monitorID: {
				ID:     monitorID,
				UserID: userID,
				Name:   "Test Monitor",
			},
		},
	}
	hbRepo := newMockHeartbeatRepo()
	_ = hbRepo.Create(context.Background(), &Heartbeat{
		MonitorID: monitorID,
		IsUp:      true,
		LatencyMs: 50,
		CheckedAt: time.Now(),
	})

	svc := NewService(hbRepo, monRepo)
	handler := NewHandler(svc)

	v1 := r.Group("/api/v1")
	fakeAuthMiddleware := func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
	handler.RegisterRoutes(v1, fakeAuthMiddleware)

	// 1. Get Heartbeats
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/monitors/"+monitorID.String()+"/heartbeats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. Get Stats
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/monitors/"+monitorID.String()+"/stats?period=24h", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Nonexistent monitor
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/monitors/"+uuid.New().String()+"/heartbeats", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
