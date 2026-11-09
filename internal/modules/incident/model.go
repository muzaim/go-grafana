package incident

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IncidentStatus string

const (
	IncidentStatusTriggered    IncidentStatus = "TRIGGERED"
	IncidentStatusAcknowledged IncidentStatus = "ACKNOWLEDGED"
	IncidentStatusResolved     IncidentStatus = "RESOLVED"
)

type Incident struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	MonitorID      uuid.UUID      `gorm:"type:uuid;index;not null" json:"monitor_id"`
	UserID         uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	Status         IncidentStatus `gorm:"type:varchar(20);not null;default:'TRIGGERED'" json:"status"`
	ErrorReason    string         `gorm:"type:text" json:"error_reason"`
	StartedAt      time.Time      `gorm:"not null" json:"started_at"`
	ResolvedAt     *time.Time     `json:"resolved_at,omitempty"`
	DurationSec    int64          `gorm:"default:0" json:"duration_sec"`
	AcknowledgedAt *time.Time     `json:"acknowledged_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func (i *Incident) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	if i.Status == "" {
		i.Status = IncidentStatusTriggered
	}
	if i.StartedAt.IsZero() {
		i.StartedAt = time.Now()
	}
	return nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Incident{})
}
