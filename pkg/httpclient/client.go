package httpclient

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

var (
	ErrBlockedAddress = errors.New("connection blocked: target resolves to a private or disallowed IP range (SSRF protection)")
)

var privateIPBlocks []*net.IPNet

func init() {
	cidrs := []string{
		"127.0.0.0/8",          // IPv4 loopback
		"10.0.0.0/8",           // RFC1918
		"172.16.0.0/12",        // RFC1918
		"192.168.0.0/16",       // RFC1918
		"169.254.0.0/16",       // IPv4 link-local (AWS metadata, etc.)
		"::1/128",              // IPv6 loopback
		"fc00::/7",             // IPv6 unique local
		"fe80::/10",            // IPv6 link-local
		"0.0.0.0/8",            // Current network
		"100.64.0.0/10",        // Carrier-grade NAT
		"198.18.0.0/15",        // Benchmarking
	}

	for _, cidr := range cidrs {
		_, block, err := net.ParseCIDR(cidr)
		if err == nil {
			privateIPBlocks = append(privateIPBlocks, block)
		}
	}
}

func IsPrivateIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() {
		return true
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// SafeDialContext performs DNS resolution and verifies all resolved IPs are public before dialing
func SafeDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address %q: %w", address, err)
	}

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("DNS resolution failed for %q: %w", host, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no IP addresses found for host %q", host)
	}

	for _, ip := range ips {
		if IsPrivateIP(ip) {
			return nil, fmt.Errorf("%w: %s resolves to %s", ErrBlockedAddress, host, ip.String())
		}
	}

	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// Dial the first verified IP
	targetAddr := net.JoinHostPort(ips[0].String(), port)
	return dialer.DialContext(ctx, network, targetAddr)
}

func NewSafeHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	transport := &http.Transport{
		DialContext: SafeDialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
