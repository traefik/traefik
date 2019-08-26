package ip

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteAddrStrategy_GetIP(t *testing.T) {
	testCases := []struct {
		desc     string
		expected string
	}{
		{
			desc:     "Use RemoteAddr",
			expected: "192.0.2.1",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			strategy := RemoteAddrStrategy{}
			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
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
	}{
		{
			desc:          "Use depth",
			depth:         3,
			xForwardedFor: "10.0.0.4,10.0.0.3,10.0.0.2,10.0.0.1",
			expected:      "10.0.0.3",
		},
		{
			desc:          "Use non existing depth in XForwardedFor",
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
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			strategy := DepthStrategy{Depth: test.depth}
			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			req.Header.Set(xForwardedFor, test.xForwardedFor)
			actual := strategy.GetIP(req)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestExcludedIPsStrategy_GetIP(t *testing.T) {
	testCases := []struct {
		desc          string
		excludedIPs   []string
		xForwardedFor string
		expected      string
	}{
		{
			desc:          "Use excluded all IPs",
			excludedIPs:   []string{"10.0.0.4", "10.0.0.3", "10.0.0.2", "10.0.0.1"},
			xForwardedFor: "10.0.0.4,10.0.0.3,10.0.0.2,10.0.0.1",
			expected:      "",
		},
		{
			desc:          "Use excluded IPs",
			excludedIPs:   []string{"10.0.0.2", "10.0.0.1"},
			xForwardedFor: "10.0.0.4,10.0.0.3,10.0.0.2,10.0.0.1",
			expected:      "10.0.0.3",
		},
		{
			desc:          "Use excluded IPs CIDR",
			excludedIPs:   []string{"10.0.0.1/24"},
			xForwardedFor: "127.0.0.1,10.0.0.4,10.0.0.3,10.0.0.2,10.0.0.1",
			expected:      "127.0.0.1",
		},
		{
			desc:          "Use excluded all IPs CIDR",
			excludedIPs:   []string{"10.0.0.1/24"},
			xForwardedFor: "10.0.0.4,10.0.0.3,10.0.0.2,10.0.0.1",
			expected:      "",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			checker, err := NewChecker(test.excludedIPs)
			require.NoError(t, err)

			strategy := CheckerStrategy{Checker: checker}
			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			req.Header.Set(xForwardedFor, test.xForwardedFor)
			actual := strategy.GetIP(req)
			assert.Equal(t, test.expected, actual)
		})
	}
}
