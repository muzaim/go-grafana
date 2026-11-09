package heartbeat

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Heartbeat struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MonitorID    uuid.UUID `gorm:"type:uuid;index:idx_heartbeat_monitor_checked,priority:1;not null" json:"monitor_id"`
	StatusCode   int       `gorm:"not null;default:0" json:"status_code"`
	LatencyMs    int64     `gorm:"not null;default:0" json:"latency_ms"`
	IsUp         bool      `gorm:"not null" json:"is_up"`
	ErrorMessage string    `gorm:"type:text" json:"error_message,omitempty"`
	CheckedAt    time.Time `gorm:"index:idx_heartbeat_monitor_checked,priority:2;not null" json:"checked_at"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Heartbeat{})
}
