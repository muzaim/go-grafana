package httpclient

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsPrivateIP(t *testing.T) {
	testCases := []struct {
		ip        string
		isPrivate bool
	}{
		{"127.0.0.1", true},
		{"127.0.0.254", true},
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.0.1", true},
		{"192.168.1.100", true},
		{"169.254.169.254", true}, // AWS / Cloud metadata
		{"::1", true},
		{"fe80::1", true},
		{"8.8.8.8", false}, // Google DNS
		{"1.1.1.1", false}, // Cloudflare DNS
	}

	for _, tc := range testCases {
		t.Run(tc.ip, func(t *testing.T) {
			parsed := net.ParseIP(tc.ip)
			assert.NotNil(t, parsed)
			assert.Equal(t, tc.isPrivate, IsPrivateIP(parsed))
		})
	}
}

func TestSafeDialContext_BlocksPrivateIPs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. Dialing loopback IP directly
	_, err := SafeDialContext(ctx, "tcp", "127.0.0.1:8080")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrBlockedAddress)

	// 2. Dialing private RFC1918
	_, err = SafeDialContext(ctx, "tcp", "192.168.1.1:80")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrBlockedAddress)

	// 3. Dialing cloud metadata
	_, err = SafeDialContext(ctx, "tcp", "169.254.169.254:80")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrBlockedAddress)

	// 4. Dialing localhost hostname
	_, err = SafeDialContext(ctx, "tcp", "localhost:8080")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrBlockedAddress)
}

func TestNewSafeHTTPClient_BlocksRequestToLoopback(t *testing.T) {
	client := NewSafeHTTPClient(2 * time.Second)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:6379", nil)
	assert.NoError(t, err)

	_, err = client.Do(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection blocked")
}
