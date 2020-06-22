package tls

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
	"time"
)

func TestCompareX509TimeBoundaries(t *testing.T) {
	const monthDuration = 30 * 24 * 60 * 60 * time.Second

	now := time.Now()
	testCases := []struct {
		desc            string
		originNotBefore time.Time
		originNotAfter  time.Time
		targetNotBefore time.Time
		targetNotAfter  time.Time
		expected        bool
	}{
		{
			desc:            "Both invalid with past boundaries, origin with greatest NotAfter boundary",
			originNotBefore: now.Add(-14 * monthDuration),
			originNotAfter:  now.Add(-1 * monthDuration),
			targetNotBefore: now.Add(-15 * monthDuration),
			targetNotAfter:  now.Add(-2 * monthDuration),
			expected:        false,
		},
		{
			desc:            "Both invalid with past boundaries, target with greatest NotAfter boundary",
			originNotBefore: now.Add(-15 * monthDuration),
			originNotAfter:  now.Add(-2 * monthDuration),
			targetNotBefore: now.Add(-14 * monthDuration),
			targetNotAfter:  now.Add(-1 * monthDuration),
			expected:        true,
		},
		{
			desc:            "Both invalid with future boundaries, origin with smallest NotBefore boundary",
			originNotBefore: now.Add(1 * monthDuration),
			originNotAfter:  now.Add(14 * monthDuration),
			targetNotBefore: now.Add(2 * monthDuration),
			targetNotAfter:  now.Add(15 * monthDuration),
			expected:        false,
		},
		{
			desc:            "Both invalid with future boundaries, target with smallest NotBefore boundary",
			originNotBefore: now.Add(2 * monthDuration),
			originNotAfter:  now.Add(15 * monthDuration),
			targetNotBefore: now.Add(1 * monthDuration),
			targetNotAfter:  now.Add(14 * monthDuration),
			expected:        true,
		},
		{
			desc:            "Both valid with, origin with greatest NotAfter boundary",
			originNotBefore: now.Add(-1 * monthDuration),
			originNotAfter:  now.Add(12 * monthDuration),
			targetNotBefore: now.Add(-2 * monthDuration),
			targetNotAfter:  now.Add(11 * monthDuration),
			expected:        false,
		},
		{
			desc:            "Both valid with, target with greater NotAfter boundary",
			originNotBefore: now.Add(-2 * monthDuration),
			originNotAfter:  now.Add(11 * monthDuration),
			targetNotBefore: now.Add(-1 * monthDuration),
			targetNotAfter:  now.Add(12 * monthDuration),
			expected:        true,
		},
		{
			desc:            "Only origin valid",
			originNotBefore: now.Add(-1 * monthDuration),
			originNotAfter:  now.Add(12 * monthDuration),
			targetNotBefore: now.Add(-15 * monthDuration),
			targetNotAfter:  now.Add(-2 * monthDuration),
			expected:        false,
		},
		{
			desc:            "Only target valid",
			originNotBefore: now.Add(-15 * monthDuration),
			originNotAfter:  now.Add(-2 * monthDuration),
			targetNotBefore: now.Add(-1 * monthDuration),
			targetNotAfter:  now.Add(12 * monthDuration),
			expected:        true,
		},
		{
			desc:            "Same boundaries",
			originNotBefore: now.Add(-1 * monthDuration),
			originNotAfter:  now.Add(12 * monthDuration),
			targetNotBefore: now.Add(-1 * monthDuration),
			targetNotAfter:  now.Add(12 * monthDuration),
			expected:        false,
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			origin := &x509.Certificate{
				NotBefore: test.originNotBefore,
				NotAfter:  test.originNotAfter,
				Subject:   pkix.Name{CommonName: "origin"},
			}
			target := &x509.Certificate{
				NotBefore: test.targetNotBefore,
				NotAfter:  test.targetNotAfter,
				Subject:   pkix.Name{CommonName: "target"},
			}

			if CompareX509TimeBoundaries(origin, target) != test.expected {
				if !test.expected {
					t.Errorf(
						"Expected origin got target:\noriginNotBefore=%s originNotAfter=%s\ntargetNotBefore=%s targetNotAfter=%s",
						test.originNotBefore.Round(24*time.Hour),
						test.originNotAfter.Round(24*time.Hour),
						test.targetNotBefore.Round(24*time.Hour),
						test.targetNotAfter.Round(24*time.Hour),
					)
				} else {
					t.Errorf(
						"Expected target got origin:\noriginNotBefore=%s originNotAfter=%s\ntargetNotBefore=%s targetNotAfter=%s",
						test.originNotBefore.Round(24*time.Hour),
						test.originNotAfter.Round(24*time.Hour),
						test.targetNotBefore.Round(24*time.Hour),
						test.targetNotAfter.Round(24*time.Hour),
					)
				}
			}
		})
	}
}
