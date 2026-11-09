package incident

import (
	"context"
	"fmt"

	"go-up-monitoring/internal/modules/monitor"
	"go-up-monitoring/internal/worker"

	"github.com/hibiken/asynq"
)

type AsynqAlertNotifier struct {
	client *asynq.Client
}

func NewAsynqAlertNotifier(client *asynq.Client) *AsynqAlertNotifier {
	return &AsynqAlertNotifier{client: client}
}

func (a *AsynqAlertNotifier) NotifyDown(ctx context.Context, inc *Incident, m *monitor.Monitor) error {
	if a.client == nil {
		return nil
	}

	payload := worker.AlertPayload{
		IncidentID:    inc.ID,
		MonitorID:     m.ID,
		UserID:        m.UserID,
		MonitorName:   m.Name,
		URL:           m.URL,
		EventType:     "DOWN",
		ErrorReason:   inc.ErrorReason,
		StartedAtUnix: inc.StartedAt.Unix(),
	}

	task, err := worker.NewAlertTask(payload)
	if err != nil {
		return fmt.Errorf("failed to create alert task: %w", err)
	}

	_, err = a.client.EnqueueContext(ctx, task, asynq.Queue("alerts"), asynq.MaxRetry(3))
	return err
}

func (a *AsynqAlertNotifier) NotifyResolved(ctx context.Context, inc *Incident, m *monitor.Monitor) error {
	if a.client == nil {
		return nil
	}

	duration := inc.DurationSec
	if duration <= 0 && inc.ResolvedAt != nil {
		duration = int64(inc.ResolvedAt.Sub(inc.StartedAt).Seconds())
	}

	payload := worker.AlertPayload{
		IncidentID:    inc.ID,
		MonitorID:     m.ID,
		UserID:        m.UserID,
		MonitorName:   m.Name,
		URL:           m.URL,
		EventType:     "RESOLVED",
		StartedAtUnix: inc.StartedAt.Unix(),
		DurationSec:   duration,
	}

	task, err := worker.NewAlertTask(payload)
	if err != nil {
		return fmt.Errorf("failed to create alert task: %w", err)
	}

	_, err = a.client.EnqueueContext(ctx, task, asynq.Queue("alerts"), asynq.MaxRetry(3))
	return err
}
