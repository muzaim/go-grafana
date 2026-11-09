package monitor

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Protocol string

const (
	ProtocolHTTP  Protocol = "http"
	ProtocolHTTPS Protocol = "https"
	ProtocolTCP   Protocol = "tcp"
)

type Status string

const (
	StatusPending Status = "PENDING"
	StatusUp      Status = "UP"
	StatusDown    Status = "DOWN"
)

type Monitor struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID             uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Name               string     `gorm:"type:varchar(255);not null" json:"name"`
	URL                string     `gorm:"type:text;not null" json:"url"`
	Protocol           Protocol   `gorm:"type:varchar(10);not null;default:'https'" json:"protocol"`
	Method             string     `gorm:"type:varchar(10);not null;default:'GET'" json:"method"`
	ExpectedStatusCode int        `gorm:"not null;default:200" json:"expected_status_code"`
	KeywordMatch       string     `gorm:"type:text" json:"keyword_match,omitempty"`
	IntervalSec        int        `gorm:"not null;default:60" json:"interval_sec"`
	TimeoutSec         int        `gorm:"not null;default:10" json:"timeout_sec"`
	RetryThreshold     int        `gorm:"not null;default:3" json:"retry_threshold"`
	ConsecutiveFails   int        `gorm:"not null;default:0" json:"consecutive_fails"`
	Status             Status     `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	IsActive           bool       `gorm:"not null;default:true" json:"is_active"`
	LastCheckedAt      *time.Time `json:"last_checked_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (m *Monitor) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.Status == "" {
		m.Status = StatusPending
	}
	return nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Monitor{})
}
