package notification

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockTelegramTransport struct {
	lastBody string
}

func (m *mockTelegramTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	bytes, _ := io.ReadAll(req.Body)
	m.lastBody = string(bytes)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
		Header:     make(http.Header),
	}, nil
}

func TestFormatMessages(t *testing.T) {
	started := time.Date(2026, 11, 9, 10, 0, 0, 0, time.UTC)
	resolved := time.Date(2026, 11, 9, 10, 5, 30, 0, time.UTC)
	duration := int64(resolved.Sub(started).Seconds()) // 330s -> 5m 30s

	downMsg := FormatDownMessage("Main API", "https://api.example.com", "HTTP 500", started)
	assert.Contains(t, downMsg, "🚨 *MONITOR DOWN* 🚨")
	assert.Contains(t, downMsg, "`Main API`")
	assert.Contains(t, downMsg, "HTTP 500")

	resolvedMsg := FormatResolvedMessage("Main API", "https://api.example.com", duration, resolved)
	assert.Contains(t, resolvedMsg, "✅ *MONITOR RESOLVED* ✅")
	assert.Contains(t, resolvedMsg, "5m 30s")
}

func TestTelegramClient_SendMessage(t *testing.T) {
	transport := &mockTelegramTransport{}
	client := NewTelegramClient("dummy_token", &http.Client{Transport: transport})

	err := client.SendMessage(context.Background(), "-10012345678", "Hello Telegram")
	assert.NoError(t, err)
	assert.Contains(t, transport.lastBody, "-10012345678")
	assert.Contains(t, transport.lastBody, "Hello Telegram")
	assert.Contains(t, transport.lastBody, "Markdown")
}
