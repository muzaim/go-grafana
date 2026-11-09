package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-up-monitoring/pkg/logger"

	"go.uber.org/zap"
)

type TelegramSender interface {
	SendMessage(ctx context.Context, chatID, text string) error
}

type TelegramClient struct {
	botToken   string
	httpClient *http.Client
}

func NewTelegramClient(botToken string, httpClient *http.Client) *TelegramClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &TelegramClient{
		botToken:   botToken,
		httpClient: httpClient,
	}
}

type telegramSendMessagePayload struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

func (t *TelegramClient) SendMessage(ctx context.Context, chatID, text string) error {
	if t.botToken == "" || chatID == "" {
		logger.Warn("telegram bot_token or chat_id is empty, skipping message dispatch",
			zap.String("chat_id", chatID))
		return nil
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	payload := telegramSendMessagePayload{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed sending telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram API returned error status %d", resp.StatusCode)
	}

	return nil
}

func FormatDuration(seconds int64) string {
	d := time.Duration(seconds) * time.Second
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, mins, secs)
	}
	if mins > 0 {
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

func FormatDownMessage(monitorName, targetURL, errorReason string, startedAt time.Time) string {
	return fmt.Sprintf(
		"🚨 *MONITOR DOWN* 🚨\n\n"+
			"*Target:* `%s`\n"+
			"*URL:* %s\n"+
			"*Error:* %s\n"+
			"*Time:* %s",
		monitorName,
		targetURL,
		errorReason,
		startedAt.UTC().Format("2006-01-02 15:04:05 MST"),
	)
}

func FormatResolvedMessage(monitorName, targetURL string, durationSec int64, resolvedAt time.Time) string {
	return fmt.Sprintf(
		"✅ *MONITOR RESOLVED* ✅\n\n"+
			"*Target:* `%s`\n"+
			"*URL:* %s\n"+
			"*Downtime Duration:* %s\n"+
			"*Resolved At:* %s",
		monitorName,
		targetURL,
		FormatDuration(durationSec),
		resolvedAt.UTC().Format("2006-01-02 15:04:05 MST"),
	)
}
