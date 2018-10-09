package types

import (
	"fmt"
	"testing"

	"github.com/containous/traefik/ip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders_ShouldReturnFalseWhenNotHasCustomHeadersDefined(t *testing.T) {
	headers := Headers{}

	assert.False(t, headers.HasCustomHeadersDefined())
}

func TestHeaders_ShouldReturnTrueWhenHasCustomHeadersDefined(t *testing.T) {
	headers := Headers{}

	headers.CustomRequestHeaders = map[string]string{
		"foo": "bar",
	}

	assert.True(t, headers.HasCustomHeadersDefined())
}

func TestHeaders_ShouldReturnFalseWhenNotHasSecureHeadersDefined(t *testing.T) {
	headers := Headers{}

	assert.False(t, headers.HasSecureHeadersDefined())
}

func TestHeaders_ShouldReturnTrueWhenHasSecureHeadersDefined(t *testing.T) {
	headers := Headers{}

	headers.SSLRedirect = true

	assert.True(t, headers.HasSecureHeadersDefined())
}

func TestNewHTTPCodeRanges(t *testing.T) {
	testCases := []struct {
		desc        string
		strBlocks   []string
		expected    HTTPCodeRanges
		errExpected bool
	}{
		{
			desc: "Should return 2 code range",
			strBlocks: []string{
				"200-500",
				"502",
			},
			expected:    HTTPCodeRanges{[2]int{200, 500}, [2]int{502, 502}},
			errExpected: false,
		},
		{
			desc: "Should return 2 code range",
			strBlocks: []string{
				"200-500",
				"205",
			},
			expected:    HTTPCodeRanges{[2]int{200, 500}, [2]int{205, 205}},
			errExpected: false,
		},
		{
			desc: "invalid code range",
			strBlocks: []string{
				"200-500",
				"aaa",
			},
			expected:    nil,
			errExpected: true,
		},
		{
			desc:        "invalid code range nil",
			strBlocks:   nil,
			expected:    nil,
			errExpected: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual, err := NewHTTPCodeRanges(test.strBlocks)
			assert.Equal(t, test.expected, actual)
			if test.errExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHTTPCodeRanges_Contains(t *testing.T) {
	testCases := []struct {
		strBlocks  []string
		statusCode int
		contains   bool
	}{
		{
			strBlocks:  []string{"200-299"},
			statusCode: 200,
			contains:   true,
		},
		{
			strBlocks:  []string{"200"},
			statusCode: 200,
			contains:   true,
		},
		{
			strBlocks:  []string{"201"},
			statusCode: 200,
			contains:   false,
		},
		{
			strBlocks:  []string{"200-299", "500-599"},
			statusCode: 400,
			contains:   false,
		},
	}

	for _, test := range testCases {
		test := test
		testName := fmt.Sprintf("%q contains %d", test.strBlocks, test.statusCode)
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			httpCodeRanges, err := NewHTTPCodeRanges(test.strBlocks)
			assert.NoError(t, err)

			assert.Equal(t, test.contains, httpCodeRanges.Contains(test.statusCode))
		})
	}
}

func TestIPStrategy_Get(t *testing.T) {
	testCases := []struct {
		desc       string
		ipStrategy *IPStrategy
		expected   ip.Strategy
	}{
		{
			desc:     "IPStrategy is nil",
			expected: &ip.RemoteAddrStrategy{},
		},
		{
			desc:       "IPStrategy is not nil but with no values",
			ipStrategy: &IPStrategy{},
			expected:   &ip.RemoteAddrStrategy{},
		},
		{
			desc:       "IPStrategy with Depth",
			ipStrategy: &IPStrategy{Depth: 3},
			expected:   &ip.DepthStrategy{},
		},
		{
			desc:       "IPStrategy with ExcludedIPs",
			ipStrategy: &IPStrategy{ExcludedIPs: []string{"10.0.0.1"}},
			expected:   &ip.CheckerStrategy{},
		},
		{
			desc:       "IPStrategy with ExcludedIPs and Depth",
			ipStrategy: &IPStrategy{Depth: 4, ExcludedIPs: []string{"10.0.0.1"}},
			expected:   &ip.DepthStrategy{},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			strategy, err := test.ipStrategy.Get()
			require.NoError(t, err)

			assert.IsType(t, test.expected, strategy)

		})
	}
}
