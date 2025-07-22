package ip

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ipv6Basic            = "::abcd:ffff:c0a8:1"
	ipv6BracketsPort     = "[::abcd:ffff:c0a8:1]:80"
	ipv6BracketsZonePort = "[::abcd:ffff:c0a8:1%1]:80"
)

func TestRemoteAddrStrategy_GetIP(t *testing.T) {
	testCases := []struct {
		desc       string
		expected   string
		remoteAddr string
		ipv6Subnet *int
	}{
		// Valid IP format
		{
			desc:     "Use RemoteAddr, ipv4",
			expected: "192.0.2.1",
		},
		{
			desc:       "Use RemoteAddr, ipv6 brackets with port, no IPv6 subnet",
			remoteAddr: ipv6BracketsPort,
			expected:   "::abcd:ffff:c0a8:1",
		},
		{
			desc:       "Use RemoteAddr, ipv6 brackets with zone and port, no IPv6 subnet",
			remoteAddr: ipv6BracketsZonePort,
			expected:   "::abcd:ffff:c0a8:1%1",
		},

		// Invalid IPv6 format
		{
			desc:       "Use RemoteAddr, ipv6 basic, missing brackets, no IPv6 subnet",
			remoteAddr: ipv6Basic,
			expected:   ipv6Basic,
		},

		// Valid IP format with subnet
		{
			desc:       "Use RemoteAddr, ipv4, ignore subnet",
			expected:   "192.0.2.1",
			ipv6Subnet: intPtr(24),
		},
		{
			desc:       "Use RemoteAddr, ipv6 brackets with port, subnet",
			remoteAddr: ipv6BracketsPort,
			expected:   "::abcd:0:0:0",
			ipv6Subnet: intPtr(80),
		},
		{
			desc:       "Use RemoteAddr, ipv6 brackets with zone and port, subnet",
			remoteAddr: ipv6BracketsZonePort,
			expected:   "::abcd:0:0:0",
			ipv6Subnet: intPtr(80),
		},

		// Valid IP, invalid subnet
		{
			desc:       "Use RemoteAddr, ipv6 brackets with port, invalid subnet",
			remoteAddr: ipv6BracketsPort,
			expected:   "::abcd:ffff:c0a8:1",
			ipv6Subnet: intPtr(500),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			strategy := RemoteAddrStrategy{
				IPv6Subnet: test.ipv6Subnet,
			}
			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			if test.remoteAddr != "" {
				req.RemoteAddr = test.remoteAddr
			}
			actual := strategy.GetIP(req)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDepthStrategy_GetIP(t *testing.T) {
	testCases := []struct {
		desc          string
		depth         int
		xForwardedFor string
		expected      string
		ipv6Subnet    *int
	}{
		{
			desc:          "Use depth",
			depth:         3,
			xForwardedFor: "10.0.0.4,10.0.0.3,10.0.0.2,10.0.0.1",
			expected:      "10.0.0.3",
		},
		{
			desc:          "Use nonexistent depth in XForwardedFor",
			depth:         2,
			xForwardedFor: "",
			expected:      "",
		},
		{
			desc:          "Use depth that match the first IP in XForwardedFor",
			depth:         2,
			xForwardedFor: "10.0.0.2,10.0.0.1",
			expected:      "10.0.0.2",
		},
		{
			desc:          "Use depth with IPv4 subnet",
			depth:         2,
			xForwardedFor: "10.0.0.3,10.0.0.2,10.0.0.1",
			expected:      "10.0.0.2",
			ipv6Subnet:    intPtr(80),
		},
		{
			desc:          "Use depth with IPv6 subnet",
			depth:         2,
			xForwardedFor: "10.0.0.3," + ipv6Basic + ",10.0.0.1",
			expected:      "::abcd:0:0:0",
			ipv6Subnet:    intPtr(80),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			strategy := DepthStrategy{
				Depth:      test.depth,
				IPv6Subnet: test.ipv6Subnet,
			}
			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			req.Header.Set(xForwardedFor, test.xForwardedFor)
			actual := strategy.GetIP(req)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestTrustedIPsStrategy_GetIP(t *testing.T) {
	testCases := []struct {
		desc          string
		trustedIPs    []string
		xForwardedFor string
		expected      string
		useRemote     bool
	}{
		{
			desc:          "Trust all IPs",
			trustedIPs:    []string{"10.0.0.4", "10.0.0.3", "10.0.0.2", "10.0.0.1"},
			xForwardedFor: "10.0.0.4,10.0.0.3,10.0.0.2,10.0.0.1",
			expected:      "",
		},
		{
			desc:          "Do not trust all IPs",
			trustedIPs:    []string{"10.0.0.2", "10.0.0.1"},
			xForwardedFor: "10.0.0.4,10.0.0.3,10.0.0.2,10.0.0.1",
			expected:      "10.0.0.3",
		},
		{
			desc:          "Do not trust all IPs with CIDR",
			trustedIPs:    []string{"10.0.0.1/24"},
			xForwardedFor: "127.0.0.1,10.0.0.4,10.0.0.3,10.0.0.2,10.0.0.1",
			expected:      "127.0.0.1",
		},
		{
			desc:          "Trust all IPs with CIDR",
			trustedIPs:    []string{"10.0.0.1/24"},
			xForwardedFor: "10.0.0.4,10.0.0.3,10.0.0.2,10.0.0.1",
			expected:      "",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			checker, err := NewChecker(test.trustedIPs)
			require.NoError(t, err)

			strategy := PoolStrategy{Checker: checker}
			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			req.Header.Set(xForwardedFor, test.xForwardedFor)
			actual := strategy.GetIP(req)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func intPtr(value int) *int {
	return &value
}
