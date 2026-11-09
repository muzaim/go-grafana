package monitor

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMonitorService_Create_Validation(t *testing.T) {
	repo := newMockMonitorRepo()
	svc := NewService(repo)
	ctx := context.Background()
	userID := uuid.New()

	// 1. URL scheme mismatch (https URL with http protocol)
	_, err := svc.Create(ctx, userID, CreateMonitorRequest{
		Name:        "Test Monitor",
		URL:         "https://example.com",
		Protocol:    ProtocolHTTP,
		IntervalSec: 60,
		TimeoutSec:  10,
	})
	assert.ErrorIs(t, err, ErrProtocolMismatch)

	// 2. Timeout >= Interval
	_, err = svc.Create(ctx, userID, CreateMonitorRequest{
		Name:        "Test Monitor",
		URL:         "https://example.com",
		Protocol:    ProtocolHTTPS,
		IntervalSec: 10,
		TimeoutSec:  15,
	})
	assert.ErrorIs(t, err, ErrInvalidTimeout)

	// 3. Invalid URL format
	_, err = svc.Create(ctx, userID, CreateMonitorRequest{
		Name:        "Test Monitor",
		URL:         "not-a-valid-url",
		Protocol:    ProtocolHTTPS,
		IntervalSec: 60,
		TimeoutSec:  10,
	})
	assert.Error(t, err)

	// 4. Valid HTTPS monitor creation
	created, err := svc.Create(ctx, userID, CreateMonitorRequest{
		Name:               "Valid Monitor",
		URL:                "https://example.com/health",
		Protocol:           ProtocolHTTPS,
		Method:             "GET",
		ExpectedStatusCode: 200,
		IntervalSec:        30,
		TimeoutSec:         5,
		RetryThreshold:     3,
	})
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "Valid Monitor", created.Name)
	assert.Equal(t, StatusPending, created.Status)
	assert.True(t, created.IsActive)

	// 5. Toggle active state
	toggled, err := svc.Toggle(ctx, created.ID, userID, false)
	assert.NoError(t, err)
	assert.False(t, toggled.IsActive)
}
