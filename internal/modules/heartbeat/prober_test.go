package heartbeat

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestProber_ProbeHTTP_Success(t *testing.T) {
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("OK: system operational")),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	prober := NewProberWithClient(mockClient)
	ctx := context.Background()

	// 1. Success with keyword match
	res := prober.ProbeHTTP(ctx, "http://example.com/health", "GET", 200, "operational", 2*time.Second)
	assert.True(t, res.IsUp)
	assert.Equal(t, 200, res.StatusCode)
	assert.Empty(t, res.ErrorMessage)

	// 2. Failure on status code mismatch
	res = prober.ProbeHTTP(ctx, "http://example.com/health", "GET", 201, "", 2*time.Second)
	assert.False(t, res.IsUp)
	assert.Equal(t, 200, res.StatusCode)
	assert.Contains(t, res.ErrorMessage, "status code mismatch")

	// 3. Failure on keyword not found
	res = prober.ProbeHTTP(ctx, "http://example.com/health", "GET", 200, "nonexistent-keyword", 2*time.Second)
	assert.False(t, res.IsUp)
	assert.Contains(t, res.ErrorMessage, "not found in response body")
}

func TestProber_ProbeTCP_SSRFBlock(t *testing.T) {
	prober := NewProber(2 * time.Second)
	ctx := context.Background()

	// Probing loopback should fail safely due to SSRF protection
	res := prober.ProbeTCP(ctx, "127.0.0.1:80", 1*time.Second)
	assert.False(t, res.IsUp)
	assert.Contains(t, res.ErrorMessage, "connection blocked")
}
