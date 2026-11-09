package notification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Debouncer interface {
	ShouldAlert(ctx context.Context, monitorID uuid.UUID, eventType string, cooldown time.Duration) bool
}

type MemoryDebouncer struct {
	mu      sync.Mutex
	history map[string]time.Time
}

func NewMemoryDebouncer() *MemoryDebouncer {
	return &MemoryDebouncer{
		history: make(map[string]time.Time),
	}
}

func (m *MemoryDebouncer) ShouldAlert(ctx context.Context, monitorID uuid.UUID, eventType string, cooldown time.Duration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", monitorID.String(), eventType)
	lastSent, exists := m.history[key]
	now := time.Now()

	if exists && now.Sub(lastSent) < cooldown {
		return false // Debounce/suppress
	}

	m.history[key] = now
	return true
}

type RedisDebouncer struct {
	rdb *redis.Client
}

func NewRedisDebouncer(rdb *redis.Client) *RedisDebouncer {
	return &RedisDebouncer{rdb: rdb}
}

func (r *RedisDebouncer) ShouldAlert(ctx context.Context, monitorID uuid.UUID, eventType string, cooldown time.Duration) bool {
	key := fmt.Sprintf("debounce:%s:%s", monitorID.String(), eventType)

	// Try setting with NX (only if key does not exist)
	set, err := r.rdb.SetNX(ctx, key, "1", cooldown).Result()
	if err != nil {
		// Fallback to allow alert on redis error
		return true
	}
	return set
}
