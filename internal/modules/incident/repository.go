package incident

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrIncidentNotFound = errors.New("incident not found")
)

type Repository interface {
	Create(ctx context.Context, inc *Incident) error
	FindActiveByMonitorID(ctx context.Context, monitorID uuid.UUID) (*Incident, error)
	Resolve(ctx context.Context, id uuid.UUID, resolvedAt time.Time, durationSec int64) error
	Acknowledge(ctx context.Context, id, userID uuid.UUID, ackAt time.Time) error
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Incident, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, inc *Incident) error {
	return r.db.WithContext(ctx).Create(inc).Error
}

func (r *repository) FindActiveByMonitorID(ctx context.Context, monitorID uuid.UUID) (*Incident, error) {
	var inc Incident
	err := r.db.WithContext(ctx).
		Where("monitor_id = ? AND status != ?", monitorID, IncidentStatusResolved).
		Order("started_at DESC").
		First(&inc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No active incident
		}
		return nil, err
	}
	return &inc, nil
}

func (r *repository) Resolve(ctx context.Context, id uuid.UUID, resolvedAt time.Time, durationSec int64) error {
	return r.db.WithContext(ctx).
		Model(&Incident{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       IncidentStatusResolved,
			"resolved_at":  resolvedAt,
			"duration_sec": durationSec,
		}).Error
}

func (r *repository) Acknowledge(ctx context.Context, id, userID uuid.UUID, ackAt time.Time) error {
	res := r.db.WithContext(ctx).
		Model(&Incident{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"status":          IncidentStatusAcknowledged,
			"acknowledged_at": ackAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrIncidentNotFound
	}
	return nil
}

func (r *repository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Incident, int64, error) {
	var list []Incident
	var total int64

	query := r.db.WithContext(ctx).Model(&Incident{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	err := query.Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
