package worker

import (
	"context"
	"time"

	"go-up-monitoring/internal/modules/monitor"
	"go-up-monitoring/pkg/logger"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type Scheduler struct {
	monitorRepo monitor.Repository
	asynqClient *asynq.Client
	interval    time.Duration
	stopCh      chan struct{}
}

func NewScheduler(monitorRepo monitor.Repository, asynqClient *asynq.Client, checkInterval time.Duration) *Scheduler {
	if checkInterval <= 0 {
		checkInterval = 10 * time.Second
	}
	return &Scheduler{
		monitorRepo: monitorRepo,
		asynqClient: asynqClient,
		interval:    checkInterval,
		stopCh:      make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	logger.Info("starting monitor probe scheduler", zap.Duration("interval", s.interval))

	for {
		select {
		case <-ctx.Done():
			logger.Info("scheduler stopping due to context cancellation")
			return
		case <-s.stopCh:
			logger.Info("scheduler stopped")
			return
		case <-ticker.C:
			s.dispatchActiveProbes(ctx)
		}
	}
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) dispatchActiveProbes(ctx context.Context) {
	monitors, err := s.monitorRepo.FindActiveMonitors(ctx)
	if err != nil {
		logger.Error("failed to retrieve active monitors for scheduling", zap.Error(err))
		return
	}

	now := time.Now()
	for _, mon := range monitors {
		// If never checked or interval elapsed, enqueue task
		shouldCheck := mon.LastCheckedAt == nil || now.Sub(*mon.LastCheckedAt) >= time.Duration(mon.IntervalSec)*time.Second
		if !shouldCheck {
			continue
		}

		payload := ProbePayload{
			MonitorID:          mon.ID,
			UserID:             mon.UserID,
			Name:               mon.Name,
			URL:                mon.URL,
			Protocol:           string(mon.Protocol),
			Method:             mon.Method,
			ExpectedStatusCode: mon.ExpectedStatusCode,
			KeywordMatch:       mon.KeywordMatch,
			TimeoutSec:         mon.TimeoutSec,
		}

		task, err := NewProbeTask(payload)
		if err != nil {
			logger.Error("failed to create probe task", zap.Error(err), zap.String("monitor_id", mon.ID.String()))
			continue
		}

		_, err = s.asynqClient.EnqueueContext(ctx, task, asynq.Queue("probes"), asynq.MaxRetry(2))
		if err != nil {
			logger.Error("failed to enqueue probe task", zap.Error(err), zap.String("monitor_id", mon.ID.String()))
		}
	}
}
