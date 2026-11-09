package incident

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIncident_BeforeCreate(t *testing.T) {
	inc := &Incident{
		MonitorID:   uuid.New(),
		UserID:      uuid.New(),
		ErrorReason: "connection timed out",
	}

	assert.Equal(t, uuid.Nil, inc.ID)
	assert.Empty(t, inc.Status)

	err := inc.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, inc.ID)
	assert.Equal(t, IncidentStatusTriggered, inc.Status)
	assert.False(t, inc.StartedAt.IsZero())
}

func TestIncidentDuration(t *testing.T) {
	started := time.Now().Add(-10 * time.Minute)
	resolved := time.Now()

	duration := int64(resolved.Sub(started).Seconds())
	assert.True(t, duration >= 590 && duration <= 610)
}
