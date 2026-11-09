package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"
)

type SSLCertInfo struct {
	Issuer    string    `json:"issuer"`
	Subject   string    `json:"subject"`
	ValidFrom time.Time `json:"valid_from"`
	ValidTo   time.Time `json:"valid_to"`
	DaysLeft  int       `json:"days_left"`
	IsExpired bool      `json:"is_expired"`
}

func ExtractCertInfo(cert *x509.Certificate, now time.Time) *SSLCertInfo {
	if cert == nil {
		return nil
	}

	daysLeft := int(cert.NotAfter.Sub(now).Hours() / 24)
	isExpired := now.After(cert.NotAfter)

	return &SSLCertInfo{
		Issuer:    cert.Issuer.CommonName,
		Subject:   cert.Subject.CommonName,
		ValidFrom: cert.NotBefore,
		ValidTo:   cert.NotAfter,
		DaysLeft:  daysLeft,
		IsExpired: isExpired,
	}
}

func InspectSSLCertificate(host string, port int, timeout time.Duration) (*SSLCertInfo, error) {
	if port <= 0 {
		port = 443
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	targetAddr := fmt.Sprintf("%s:%d", host, port)

	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", targetAddr, &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	})
	if err != nil {
		return nil, fmt.Errorf("TLS handshake failed for %s: %w", targetAddr, err)
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("no peer certificates presented by %s", targetAddr)
	}

	return ExtractCertInfo(certs[0], time.Now()), nil
}
