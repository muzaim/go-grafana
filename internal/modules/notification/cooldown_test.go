package notification

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMemoryDebouncer(t *testing.T) {
	debouncer := NewMemoryDebouncer()
	ctx := context.Background()
	monitorID := uuid.New()
	cooldown := 100 * time.Millisecond

	// 1. First alert -> allowed
	allowed := debouncer.ShouldAlert(ctx, monitorID, "DOWN", cooldown)
	assert.True(t, allowed)

	// 2. Immediate second alert -> debounced/blocked
	allowed = debouncer.ShouldAlert(ctx, monitorID, "DOWN", cooldown)
	assert.False(t, allowed)

	// 3. Different event type (RESOLVED) -> allowed
	allowed = debouncer.ShouldAlert(ctx, monitorID, "RESOLVED", cooldown)
	assert.True(t, allowed)

	// 4. After cooldown passes -> allowed again
	time.Sleep(120 * time.Millisecond)
	allowed = debouncer.ShouldAlert(ctx, monitorID, "DOWN", cooldown)
	assert.True(t, allowed)
}
