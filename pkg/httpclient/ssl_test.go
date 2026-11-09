package httpclient

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExtractCertInfo(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// 1. Valid certificate expiring in 30 days
	validCert := &x509.Certificate{
		Issuer:    pkix.Name{CommonName: "Let's Encrypt Authority"},
		Subject:   pkix.Name{CommonName: "example.com"},
		NotBefore: now.Add(-10 * 24 * time.Hour),
		NotAfter:  now.Add(30 * 24 * time.Hour),
	}

	info := ExtractCertInfo(validCert, now)
	assert.NotNil(t, info)
	assert.Equal(t, "example.com", info.Subject)
	assert.Equal(t, "Let's Encrypt Authority", info.Issuer)
	assert.Equal(t, 30, info.DaysLeft)
	assert.False(t, info.IsExpired)

	// 2. Expired certificate
	expiredCert := &x509.Certificate{
		Issuer:    pkix.Name{CommonName: "DigiCert"},
		Subject:   pkix.Name{CommonName: "expired.com"},
		NotBefore: now.Add(-60 * 24 * time.Hour),
		NotAfter:  now.Add(-5 * 24 * time.Hour),
	}

	infoExpired := ExtractCertInfo(expiredCert, now)
	assert.NotNil(t, infoExpired)
	assert.True(t, infoExpired.IsExpired)
	assert.True(t, infoExpired.DaysLeft < 0)
}
