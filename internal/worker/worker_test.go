package worker

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewProbeTask(t *testing.T) {
	monitorID := uuid.New()
	userID := uuid.New()

	payload := ProbePayload{
		MonitorID:          monitorID,
		UserID:             userID,
		Name:               "Test Server",
		URL:                "https://example.com",
		Protocol:           "https",
		Method:             "GET",
		ExpectedStatusCode: 200,
		TimeoutSec:         5,
	}

	task, err := NewProbeTask(payload)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeProbeMonitor, task.Type())

	var parsed ProbePayload
	err = json.Unmarshal(task.Payload(), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, monitorID, parsed.MonitorID)
	assert.Equal(t, "https://example.com", parsed.URL)
}

func TestNewAlertTask(t *testing.T) {
	incidentID := uuid.New()
	payload := AlertPayload{
		IncidentID:  incidentID,
		MonitorName: "Production API",
		URL:         "https://api.example.com",
		EventType:   "DOWN",
		ErrorReason: "connection timeout",
	}

	task, err := NewAlertTask(payload)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeDispatchAlert, task.Type())
}
