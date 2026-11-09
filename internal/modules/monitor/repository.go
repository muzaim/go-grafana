package monitor

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrMonitorNotFound = errors.New("monitor not found")
)

type Repository interface {
	Create(ctx context.Context, m *Monitor) error
	FindByID(ctx context.Context, id, userID uuid.UUID) (*Monitor, error)
	FindByIDInternal(ctx context.Context, id uuid.UUID) (*Monitor, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Monitor, int64, error)
	FindActiveMonitors(ctx context.Context) ([]Monitor, error)
	Update(ctx context.Context, m *Monitor) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status, consecutiveFails int, lastChecked time.Time) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, m *Monitor) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *repository) FindByID(ctx context.Context, id, userID uuid.UUID) (*Monitor, error) {
	var m Monitor
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMonitorNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *repository) FindByIDInternal(ctx context.Context, id uuid.UUID) (*Monitor, error) {
	var m Monitor
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMonitorNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *repository) FindAllByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Monitor, int64, error) {
	var monitors []Monitor
	var total int64

	query := r.db.WithContext(ctx).Model(&Monitor{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&monitors).Error
	if err != nil {
		return nil, 0, err
	}

	return monitors, total, nil
}

func (r *repository) FindActiveMonitors(ctx context.Context) ([]Monitor, error) {
	var monitors []Monitor
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&monitors).Error
	return monitors, err
}

func (r *repository) Update(ctx context.Context, m *Monitor) error {
	res := r.db.WithContext(ctx).
		Model(&Monitor{}).
		Where("id = ? AND user_id = ?", m.ID, m.UserID).
		Updates(m)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrMonitorNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&Monitor{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrMonitorNotFound
	}
	return nil
}

func (r *repository) UpdateStatus(ctx context.Context, id uuid.UUID, status Status, consecutiveFails int, lastChecked time.Time) error {
	return r.db.WithContext(ctx).
		Model(&Monitor{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":            status,
			"consecutive_fails": consecutiveFails,
			"last_checked_at":   lastChecked,
		}).Error
}
