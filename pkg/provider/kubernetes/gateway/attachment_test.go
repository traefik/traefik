package gateway

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

func Test_attachedRoutes_Claim(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Second)
	t2 := t0.Add(2 * time.Second)

	testCases := []struct {
		desc      string
		initial   []routeClaim
		kind      string
		hostname  gatev1.Hostname
		timestamp time.Time
		expected  []routeClaim
	}{
		{
			desc:      "no existing claims",
			kind:      kindHTTPRoute,
			hostname:  "foo.com",
			timestamp: t0,
			expected: []routeClaim{
				{kind: kindHTTPRoute, hostname: "foo.com", timestamp: t0},
			},
		},
		{
			desc: "non-intersecting claim",
			initial: []routeClaim{
				{kind: kindHTTPRoute, hostname: "foo.com", timestamp: t0},
			},
			kind:      kindHTTPRoute,
			hostname:  "bar.com",
			timestamp: t1,
			expected: []routeClaim{
				{kind: kindHTTPRoute, hostname: "foo.com", timestamp: t0},
				{kind: kindHTTPRoute, hostname: "bar.com", timestamp: t1},
			},
		},
		{
			desc: "new claim is older than the intersecting claim",
			initial: []routeClaim{
				{kind: kindHTTPRoute, hostname: "foo.com", timestamp: t1},
			},
			kind:      kindGRPCRoute,
			hostname:  "foo.com",
			timestamp: t0,
			expected: []routeClaim{
				{kind: kindGRPCRoute, hostname: "foo.com", timestamp: t0},
			},
		},
		{
			desc: "new claim is newer than the intersecting claim",
			initial: []routeClaim{
				{kind: kindHTTPRoute, hostname: "foo.com", timestamp: t0},
			},
			kind:      kindGRPCRoute,
			hostname:  "foo.com",
			timestamp: t1,
			expected: []routeClaim{
				{kind: kindHTTPRoute, hostname: "foo.com", timestamp: t0},
			},
		},
		{
			desc: "wildcard older than multiple exact claims",
			initial: []routeClaim{
				{kind: kindHTTPRoute, hostname: "a.example.com", timestamp: t1},
				{kind: kindHTTPRoute, hostname: "b.example.com", timestamp: t2},
			},
			kind:      kindGRPCRoute,
			hostname:  "*.example.com",
			timestamp: t0,
			expected: []routeClaim{
				{kind: kindGRPCRoute, hostname: "*.example.com", timestamp: t0},
			},
		},
		{
			desc: "wildcard blocked by one older exact claim",
			initial: []routeClaim{
				{kind: kindHTTPRoute, hostname: "a.example.com", timestamp: t0},
				{kind: kindHTTPRoute, hostname: "b.example.com", timestamp: t2},
			},
			kind:      kindGRPCRoute,
			hostname:  "*.example.com",
			timestamp: t1,
			expected: []routeClaim{
				{kind: kindHTTPRoute, hostname: "a.example.com", timestamp: t0},
				{kind: kindHTTPRoute, hostname: "b.example.com", timestamp: t2},
			},
		},
		{
			desc: "exact older than wildcard",
			initial: []routeClaim{
				{kind: kindHTTPRoute, hostname: "*.example.com", timestamp: t1},
			},
			kind:      kindGRPCRoute,
			hostname:  "a.example.com",
			timestamp: t0,
			expected: []routeClaim{
				{kind: kindGRPCRoute, hostname: "a.example.com", timestamp: t0},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			a := make(attachedRoutes)
			a["key"] = test.initial

			a.Claim("key", test.kind, test.hostname, test.timestamp)

			assert.Equal(t, test.expected, a["key"])
		})
	}
}

func Test_attachedRoutes_HasConflict(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		desc      string
		claims    []routeClaim
		kind      string
		hostnames []gatev1.Hostname
		expected  bool
	}{
		{
			desc:      "no claims",
			kind:      kindHTTPRoute,
			hostnames: []gatev1.Hostname{"foo.com"},
			expected:  false,
		},
		{
			desc: "same kind",
			claims: []routeClaim{
				{kind: kindHTTPRoute, hostname: "foo.com", timestamp: t0},
			},
			kind:      kindHTTPRoute,
			hostnames: []gatev1.Hostname{"foo.com"},
			expected:  false,
		},
		{
			desc: "different kind, non-intersecting hostname",
			claims: []routeClaim{
				{kind: kindHTTPRoute, hostname: "foo.com", timestamp: t0},
			},
			kind:      kindGRPCRoute,
			hostnames: []gatev1.Hostname{"bar.com"},
			expected:  false,
		},
		{
			desc: "different kind, exact match",
			claims: []routeClaim{
				{kind: kindHTTPRoute, hostname: "foo.com", timestamp: t0},
			},
			kind:      kindGRPCRoute,
			hostnames: []gatev1.Hostname{"foo.com"},
			expected:  true,
		},
		{
			desc: "different kind, wildcard claim covers queried exact hostname",
			claims: []routeClaim{
				{kind: kindHTTPRoute, hostname: "*.example.com", timestamp: t0},
			},
			kind:      kindGRPCRoute,
			hostnames: []gatev1.Hostname{"a.example.com"},
			expected:  true,
		},
		{
			desc: "different kind, exact claim, wildcard hostname queried",
			claims: []routeClaim{
				{kind: kindHTTPRoute, hostname: "a.example.com", timestamp: t0},
			},
			kind:      kindGRPCRoute,
			hostnames: []gatev1.Hostname{"*.example.com"},
			expected:  true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			a := make(attachedRoutes)
			a["key"] = test.claims

			assert.Equal(t, test.expected, a.HasConflict("key", test.kind, test.hostnames))
		})
	}
}
