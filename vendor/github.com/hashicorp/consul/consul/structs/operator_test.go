package structs

import (
	"testing"
	"time"

	"github.com/hashicorp/serf/serf"
)

func TestServerHealth_IsHealthy(t *testing.T) {
	cases := []struct {
		health    ServerHealth
		lastTerm  uint64
		lastIndex uint64
		conf      AutopilotConfig
		expected  bool
	}{
		// Healthy server, all values within allowed limits
		{
			health:    ServerHealth{SerfStatus: serf.StatusAlive, LastTerm: 1, LastIndex: 0},
			lastTerm:  1,
			lastIndex: 10,
			conf:      AutopilotConfig{MaxTrailingLogs: 20},
			expected:  true,
		},
		// Serf status failed
		{
			health:   ServerHealth{SerfStatus: serf.StatusFailed},
			expected: false,
		},
		// Old value for lastTerm
		{
			health:   ServerHealth{SerfStatus: serf.StatusAlive, LastTerm: 0},
			lastTerm: 1,
			expected: false,
		},
		// Too far behind on logs
		{
			health:    ServerHealth{SerfStatus: serf.StatusAlive, LastIndex: 0},
			lastIndex: 10,
			conf:      AutopilotConfig{MaxTrailingLogs: 5},
			expected:  false,
		},
	}

	for index, tc := range cases {
		actual := tc.health.IsHealthy(tc.lastTerm, tc.lastIndex, &tc.conf)
		if actual != tc.expected {
			t.Fatalf("bad value for case %d: %v", index, actual)
		}
	}
}

func TestServerHealth_IsStable(t *testing.T) {
	start := time.Now()
	cases := []struct {
		health   *ServerHealth
		now      time.Time
		conf     AutopilotConfig
		expected bool
	}{
		// Healthy server, all values within allowed limits
		{
			health:   &ServerHealth{Healthy: true, StableSince: start},
			now:      start.Add(15 * time.Second),
			conf:     AutopilotConfig{ServerStabilizationTime: 10 * time.Second},
			expected: true,
		},
		// Unhealthy server
		{
			health:   &ServerHealth{Healthy: false},
			expected: false,
		},
		// Healthy server, hasn't reached stabilization time
		{
			health:   &ServerHealth{Healthy: true, StableSince: start},
			now:      start.Add(5 * time.Second),
			conf:     AutopilotConfig{ServerStabilizationTime: 10 * time.Second},
			expected: false,
		},
		// Nil struct
		{
			health:   nil,
			expected: false,
		},
	}

	for index, tc := range cases {
		actual := tc.health.IsStable(tc.now, &tc.conf)
		if actual != tc.expected {
			t.Fatalf("bad value for case %d: %v", index, actual)
		}
	}
}
