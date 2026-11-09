package heartbeat

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StatsResult struct {
	TotalChecks   int64   `json:"total_checks"`
	UpChecks      int64   `json:"up_checks"`
	UptimePercent float64 `json:"uptime_percent"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
}

type Repository interface {
	Create(ctx context.Context, h *Heartbeat) error
	FindByMonitorID(ctx context.Context, monitorID uuid.UUID, limit, offset int) ([]Heartbeat, int64, error)
	GetStats(ctx context.Context, monitorID uuid.UUID, since time.Time) (*StatsResult, error)
	PruneOlderThan(ctx context.Context, threshold time.Time) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, h *Heartbeat) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *repository) FindByMonitorID(ctx context.Context, monitorID uuid.UUID, limit, offset int) ([]Heartbeat, int64, error) {
	var list []Heartbeat
	var total int64

	query := r.db.WithContext(ctx).Model(&Heartbeat{}).Where("monitor_id = ?", monitorID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	err := query.Order("checked_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *repository) GetStats(ctx context.Context, monitorID uuid.UUID, since time.Time) (*StatsResult, error) {
	type aggResult struct {
		TotalChecks int64
		UpChecks    int64
		AvgLatency  float64
	}

	var res aggResult
	err := r.db.WithContext(ctx).
		Model(&Heartbeat{}).
		Select("COUNT(*) as total_checks, SUM(CASE WHEN is_up = true THEN 1 ELSE 0 END) as up_checks, COALESCE(AVG(latency_ms), 0) as avg_latency").
		Where("monitor_id = ? AND checked_at >= ?", monitorID, since).
		Scan(&res).Error
	if err != nil {
		return nil, err
	}

	uptimePercent := 0.0
	if res.TotalChecks > 0 {
		uptimePercent = (float64(res.UpChecks) / float64(res.TotalChecks)) * 100.0
	}

	return &StatsResult{
		TotalChecks:   res.TotalChecks,
		UpChecks:      res.UpChecks,
		UptimePercent: uptimePercent,
		AvgLatencyMs:  res.AvgLatency,
	}, nil
}

func (r *repository) PruneOlderThan(ctx context.Context, threshold time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("checked_at < ?", threshold).
		Delete(&Heartbeat{})
	return res.RowsAffected, res.Error
}
