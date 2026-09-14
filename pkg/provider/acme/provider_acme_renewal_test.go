package acme

import (
	"crypto/x509"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetCertificateRenewDurations(t *testing.T) {
	tests := []struct {
		name           string
		certHours      int
		wantRenew      time.Duration
		wantCheck      time.Duration
	}{
		{
			name:      "one year",
			certHours: 365 * 24,
			wantRenew: 4 * 30 * 24 * time.Hour,
			wantCheck: 7 * 24 * time.Hour,
		},
		{
			name:      "ninety days",
			certHours: 90 * 24,
			wantRenew: 30 * 24 * time.Hour,
			wantCheck: 24 * time.Hour,
		},
		{
			name:      "thirty days",
			certHours: 30 * 24,
			wantRenew: 10 * 24 * time.Hour,
			wantCheck: 12 * time.Hour,
		},
		{
			name:      "one week",
			certHours: 7 * 24,
			wantRenew: 2 * 24 * time.Hour,
			wantCheck: 2 * time.Hour,
		},
		{
			name:      "one day",
			certHours: 24,
			wantRenew: 6 * time.Hour,
			wantCheck: 10 * time.Minute,
		},
		{
			name:      "short lived",
			certHours: 12,
			wantRenew: 20 * time.Minute,
			wantCheck: time.Minute,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			renew, check := getCertificateRenewDurations(tc.certHours)
			assert.Equal(t, tc.wantRenew, renew)
			assert.Equal(t, tc.wantCheck, check)
		})
	}
}

func TestShouldRenewBasedOnTime(t *testing.T) {
	now := time.Now().UTC()
	cert := &x509.Certificate{
		NotBefore: now.Add(-24 * time.Hour),
		NotAfter:  now.Add(30 * 24 * time.Hour),
	}

	assert.False(t, shouldRenewBasedOnTime(cert, 7*24*time.Hour))
	assert.True(t, shouldRenewBasedOnTime(cert, 60*24*time.Hour))
	assert.True(t, shouldRenewBasedOnTime(nil, time.Hour))
}

func TestCertLifetimeHours(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cert := &x509.Certificate{
		NotBefore: start,
		NotAfter:  start.Add(90 * 24 * time.Hour),
		SerialNumber: big.NewInt(1),
	}

	assert.Equal(t, 90*24, certLifetimeHours(cert))
}

func TestCertRenewalInfoIsARI(t *testing.T) {
	assert.False(t, (&certRenewalInfo{}).isARI())
	assert.True(t, (&certRenewalInfo{renewalID: "abc"}).isARI())
}

func TestGetNextCheckTime(t *testing.T) {
	fallback := 30 * time.Minute
	infos := []certRenewalInfo{
		{timeToWaitForNextCheck: 2 * time.Hour},
		{timeToWaitForNextCheck: 15 * time.Minute},
		{timeToWaitForNextCheck: 0},
	}

	assert.Equal(t, 15*time.Minute, getNextCheckTime(infos, fallback))
	assert.Equal(t, fallback, getNextCheckTime(nil, fallback))
	assert.Equal(t, fallback, getNextCheckTime([]certRenewalInfo{{}}, fallback))
}
