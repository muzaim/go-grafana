package worker

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeProbeMonitor  = "monitor:probe"
	TypeDispatchAlert = "alert:dispatch"
)

type ProbePayload struct {
	MonitorID          uuid.UUID `json:"monitor_id"`
	UserID             uuid.UUID `json:"user_id"`
	Name               string    `json:"name"`
	URL                string    `json:"url"`
	Protocol           string    `json:"protocol"`
	Method             string    `json:"method"`
	ExpectedStatusCode int       `json:"expected_status_code"`
	KeywordMatch       string    `json:"keyword_match,omitempty"`
	TimeoutSec         int       `json:"timeout_sec"`
}

func NewProbeTask(payload ProbePayload) (*asynq.Task, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal probe payload: %w", err)
	}
	return asynq.NewTask(TypeProbeMonitor, bytes), nil
}

type AlertPayload struct {
	IncidentID    uuid.UUID `json:"incident_id"`
	MonitorID     uuid.UUID `json:"monitor_id"`
	UserID        uuid.UUID `json:"user_id"`
	MonitorName   string    `json:"monitor_name"`
	URL           string    `json:"url"`
	EventType     string    `json:"event_type"` // "DOWN" or "RESOLVED"
	ErrorReason   string    `json:"error_reason,omitempty"`
	StartedAtUnix int64     `json:"started_at_unix"`
	DurationSec   int64     `json:"duration_sec,omitempty"`
}

func NewAlertTask(payload AlertPayload) (*asynq.Task, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal alert payload: %w", err)
	}
	return asynq.NewTask(TypeDispatchAlert, bytes), nil
}
