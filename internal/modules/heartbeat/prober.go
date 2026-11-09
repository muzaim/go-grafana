package heartbeat

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go-up-monitoring/pkg/httpclient"
)

type ProbeResult struct {
	IsUp         bool   `json:"is_up"`
	StatusCode   int    `json:"status_code"`
	LatencyMs    int64  `json:"latency_ms"`
	ErrorMessage string `json:"error_message,omitempty"`
	SSLDaysLeft  *int   `json:"ssl_days_left,omitempty"`
}

type Prober interface {
	ProbeHTTP(ctx context.Context, targetURL, method string, expectedStatus int, keyword string, timeout time.Duration) ProbeResult
	ProbeTCP(ctx context.Context, address string, timeout time.Duration) ProbeResult
}

type prober struct {
	client *http.Client
}

func NewProber(timeout time.Duration) Prober {
	return &prober{
		client: httpclient.NewSafeHTTPClient(timeout),
	}
}

func NewProberWithClient(client *http.Client) Prober {
	return &prober{
		client: client,
	}
}

func (p *prober) ProbeHTTP(ctx context.Context, targetURL, method string, expectedStatus int, keyword string, timeout time.Duration) ProbeResult {
	if method == "" {
		method = http.MethodGet
	}
	if expectedStatus <= 0 {
		expectedStatus = http.StatusOK
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, method, targetURL, nil)
	if err != nil {
		return ProbeResult{
			IsUp:         false,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("User-Agent", "GoUptimeMonitor/1.0")

	start := time.Now()
	resp, err := p.client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return ProbeResult{
			IsUp:         false,
			LatencyMs:    latency,
			ErrorMessage: fmt.Sprintf("HTTP probe failed: %v", err),
		}
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != expectedStatus {
		return ProbeResult{
			IsUp:         false,
			StatusCode:   resp.StatusCode,
			LatencyMs:    latency,
			ErrorMessage: fmt.Sprintf("status code mismatch: expected %d, got %d", expectedStatus, resp.StatusCode),
		}
	}

	// Keyword match if configured
	if keyword != "" {
		bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // limit to 1MB
		if err != nil {
			return ProbeResult{
				IsUp:         false,
				StatusCode:   resp.StatusCode,
				LatencyMs:    latency,
				ErrorMessage: fmt.Sprintf("failed reading response body: %v", err),
			}
		}

		if !strings.Contains(string(bodyBytes), keyword) {
			return ProbeResult{
				IsUp:         false,
				StatusCode:   resp.StatusCode,
				LatencyMs:    latency,
				ErrorMessage: fmt.Sprintf("keyword %q not found in response body", keyword),
			}
		}
	}

	return ProbeResult{
		IsUp:       true,
		StatusCode: resp.StatusCode,
		LatencyMs:  latency,
	}
}

func (p *prober) ProbeTCP(ctx context.Context, address string, timeout time.Duration) ProbeResult {
	start := time.Now()

	conn, err := httpclient.SafeDialContext(ctx, "tcp", address)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return ProbeResult{
			IsUp:         false,
			LatencyMs:    latency,
			ErrorMessage: fmt.Sprintf("TCP dial failed: %v", err),
		}
	}
	_ = conn.Close()

	return ProbeResult{
		IsUp:       true,
		LatencyMs:  latency,
		StatusCode: 0,
	}
}
