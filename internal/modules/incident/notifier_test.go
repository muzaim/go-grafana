package incident

import (
	"context"
	"testing"
	"time"

	"go-up-monitoring/internal/modules/monitor"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAsynqAlertNotifier_NilClientGraceful(t *testing.T) {
	notifier := NewAsynqAlertNotifier(nil)
	ctx := context.Background()

	inc := &Incident{
		ID:        uuid.New(),
		StartedAt: time.Now(),
	}
	mon := &monitor.Monitor{
		ID:   uuid.New(),
		Name: "Test App",
		URL:  "https://example.com",
	}

	err := notifier.NotifyDown(ctx, inc, mon)
	assert.NoError(t, err)

	err = notifier.NotifyResolved(ctx, inc, mon)
	assert.NoError(t, err)
}
