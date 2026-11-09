package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-up-monitoring/internal/modules/heartbeat"
	"go-up-monitoring/internal/modules/monitor"
	"go-up-monitoring/internal/modules/notification"
	"go-up-monitoring/pkg/logger"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type IncidentHandler interface {
	HandleProbeResult(ctx context.Context, m *monitor.Monitor, isUp bool, errMsg string) error
}

type Processor struct {
	monitorRepo   monitor.Repository
	heartbeatRepo heartbeat.Repository
	prober        heartbeat.Prober
	incidentHdlr  IncidentHandler
	telegram      notification.TelegramSender
	debouncer     notification.Debouncer
	defaultChatID string
}

func NewProcessor(
	monitorRepo monitor.Repository,
	heartbeatRepo heartbeat.Repository,
	prober heartbeat.Prober,
	incidentHdlr IncidentHandler,
) *Processor {
	return &Processor{
		monitorRepo:   monitorRepo,
		heartbeatRepo: heartbeatRepo,
		prober:        prober,
		incidentHdlr:  incidentHdlr,
		debouncer:     notification.NewMemoryDebouncer(),
	}
}

func (p *Processor) SetNotification(telegram notification.TelegramSender, debouncer notification.Debouncer, defaultChatID string) {
	p.telegram = telegram
	if debouncer != nil {
		p.debouncer = debouncer
	}
	p.defaultChatID = defaultChatID
}

func (p *Processor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	switch t.Type() {
	case TypeProbeMonitor:
		return p.handleProbeTask(ctx, t)
	case TypeDispatchAlert:
		return p.handleAlertTask(ctx, t)
	default:
		return fmt.Errorf("unexpected task type: %s", t.Type())
	}
}

func (p *Processor) handleProbeTask(ctx context.Context, t *asynq.Task) error {
	var payload ProbePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to parse probe payload: %w", err)
	}

	mon, err := p.monitorRepo.FindByIDInternal(ctx, payload.MonitorID)
	if err != nil {
		return fmt.Errorf("monitor %s not found: %w", payload.MonitorID, err)
	}

	if !mon.IsActive {
		logger.Debug("skipping inactive monitor", zap.String("id", mon.ID.String()))
		return nil
	}

	timeout := time.Duration(payload.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	var result heartbeat.ProbeResult
	if payload.Protocol == "tcp" {
		result = p.prober.ProbeTCP(ctx, payload.URL, timeout)
	} else {
		result = p.prober.ProbeHTTP(ctx, payload.URL, payload.Method, payload.ExpectedStatusCode, payload.KeywordMatch, timeout)
	}

	now := time.Now()
	hb := &heartbeat.Heartbeat{
		MonitorID:    mon.ID,
		StatusCode:   result.StatusCode,
		LatencyMs:    result.LatencyMs,
		IsUp:         result.IsUp,
		ErrorMessage: result.ErrorMessage,
		CheckedAt:    now,
	}

	if err := p.heartbeatRepo.Create(ctx, hb); err != nil {
		logger.Error("failed to record heartbeat", zap.Error(err), zap.String("monitor_id", mon.ID.String()))
	}

	newConsecutiveFails := 0
	newStatus := monitor.StatusUp
	if !result.IsUp {
		newConsecutiveFails = mon.ConsecutiveFails + 1
		if newConsecutiveFails >= mon.RetryThreshold {
			newStatus = monitor.StatusDown
		} else {
			newStatus = mon.Status
		}
	}

	_ = p.monitorRepo.UpdateStatus(ctx, mon.ID, newStatus, newConsecutiveFails, now)

	// Update in-memory struct for incident handler evaluation
	mon.ConsecutiveFails = newConsecutiveFails
	mon.Status = newStatus

	if p.incidentHdlr != nil {
		_ = p.incidentHdlr.HandleProbeResult(ctx, mon, result.IsUp, result.ErrorMessage)
	}

	logger.Info("probe completed",
		zap.String("monitor", mon.Name),
		zap.Bool("is_up", result.IsUp),
		zap.Int64("latency_ms", result.LatencyMs),
		zap.Int("status_code", result.StatusCode),
	)

	return nil
}

func (p *Processor) handleAlertTask(ctx context.Context, t *asynq.Task) error {
	var payload AlertPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to parse alert payload: %w", err)
	}

	if p.telegram == nil || p.defaultChatID == "" {
		logger.Warn("telegram dispatcher not configured, skipping alert",
			zap.String("monitor", payload.MonitorName),
			zap.String("event", payload.EventType))
		return nil
	}

	// Debounce check: 5 minute cooldown
	if p.debouncer != nil && !p.debouncer.ShouldAlert(ctx, payload.MonitorID, payload.EventType, 5*time.Minute) {
		logger.Info("alert debounced/suppressed",
			zap.String("monitor", payload.MonitorName),
			zap.String("event", payload.EventType))
		return nil
	}

	var message string
	startedAt := time.Unix(payload.StartedAtUnix, 0)

	if payload.EventType == "DOWN" {
		message = notification.FormatDownMessage(payload.MonitorName, payload.URL, payload.ErrorReason, startedAt)
	} else {
		message = notification.FormatResolvedMessage(payload.MonitorName, payload.URL, payload.DurationSec, time.Now())
	}

	if err := p.telegram.SendMessage(ctx, p.defaultChatID, message); err != nil {
		logger.Error("failed sending alert to telegram", zap.Error(err), zap.String("monitor", payload.MonitorName))
		return err
	}

	logger.Info("alert sent to telegram successfully",
		zap.String("monitor", payload.MonitorName),
		zap.String("event", payload.EventType))

	return nil
}
