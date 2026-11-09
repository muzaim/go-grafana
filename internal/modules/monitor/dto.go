package monitor

import (
	"time"

	"github.com/google/uuid"
)

type CreateMonitorRequest struct {
	Name               string   `json:"name" binding:"required,min=2,max=255"`
	URL                string   `json:"url" binding:"required"`
	Protocol           Protocol `json:"protocol" binding:"required,oneof=http https tcp"`
	Method             string   `json:"method" binding:"omitempty,oneof=GET POST HEAD PUT DELETE"`
	ExpectedStatusCode int      `json:"expected_status_code" binding:"omitempty,min=100,max=599"`
	KeywordMatch       string   `json:"keyword_match,omitempty"`
	IntervalSec        int      `json:"interval_sec" binding:"omitempty,min=10,max=86400"`
	TimeoutSec         int      `json:"timeout_sec" binding:"omitempty,min=1,max=60"`
	RetryThreshold     int      `json:"retry_threshold" binding:"omitempty,min=1,max=10"`
}

type UpdateMonitorRequest struct {
	Name               string   `json:"name" binding:"required,min=2,max=255"`
	URL                string   `json:"url" binding:"required"`
	Protocol           Protocol `json:"protocol" binding:"required,oneof=http https tcp"`
	Method             string   `json:"method" binding:"omitempty,oneof=GET POST HEAD PUT DELETE"`
	ExpectedStatusCode int      `json:"expected_status_code" binding:"omitempty,min=100,max=599"`
	KeywordMatch       string   `json:"keyword_match,omitempty"`
	IntervalSec        int      `json:"interval_sec" binding:"omitempty,min=10,max=86400"`
	TimeoutSec         int      `json:"timeout_sec" binding:"omitempty,min=1,max=60"`
	RetryThreshold     int      `json:"retry_threshold" binding:"omitempty,min=1,max=10"`
}

type MonitorResponse struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             uuid.UUID  `json:"user_id"`
	Name               string     `json:"name"`
	URL                string     `json:"url"`
	Protocol           Protocol   `json:"protocol"`
	Method             string     `json:"method"`
	ExpectedStatusCode int        `json:"expected_status_code"`
	KeywordMatch       string     `json:"keyword_match,omitempty"`
	IntervalSec        int        `json:"interval_sec"`
	TimeoutSec         int        `json:"timeout_sec"`
	RetryThreshold     int        `json:"retry_threshold"`
	ConsecutiveFails   int        `json:"consecutive_fails"`
	Status             Status     `json:"status"`
	IsActive           bool       `json:"is_active"`
	LastCheckedAt      *time.Time `json:"last_checked_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func ToMonitorResponse(m *Monitor) MonitorResponse {
	return MonitorResponse{
		ID:                 m.ID,
		UserID:             m.UserID,
		Name:               m.Name,
		URL:                m.URL,
		Protocol:           m.Protocol,
		Method:             m.Method,
		ExpectedStatusCode: m.ExpectedStatusCode,
		KeywordMatch:       m.KeywordMatch,
		IntervalSec:        m.IntervalSec,
		TimeoutSec:         m.TimeoutSec,
		RetryThreshold:     m.RetryThreshold,
		ConsecutiveFails:   m.ConsecutiveFails,
		Status:             m.Status,
		IsActive:           m.IsActive,
		LastCheckedAt:      m.LastCheckedAt,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

type MonitorListResponse struct {
	Monitors []MonitorResponse `json:"monitors"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}
