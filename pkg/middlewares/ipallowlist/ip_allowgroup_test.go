package ipallowlist

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestNewIPAllowGrouper(t *testing.T) {
	testCases := []struct {
		desc          string
		config        dynamic.IPAllowGroup
		expectedError bool
	}{
		{
			desc:          "empty rules",
			config:        dynamic.IPAllowGroup{Rules: nil},
			expectedError: true,
		},
		{
			desc: "empty sourceRange in rule",
			config: dynamic.IPAllowGroup{
				Rules: []dynamic.IPAllowList{
					{SourceRange: []string{}},
				},
			},
			expectedError: true,
		},
		{
			desc: "invalid CIDR in rule",
			config: dynamic.IPAllowGroup{
				Rules: []dynamic.IPAllowList{
					{SourceRange: []string{"not-a-cidr"}},
				},
			},
			expectedError: true,
		},
		{
			desc: "invalid reject status code in rule",
			config: dynamic.IPAllowGroup{
				Rules: []dynamic.IPAllowList{
					{SourceRange: []string{"10.0.0.0/8"}, RejectStatusCode: 600},
				},
			},
			expectedError: true,
		},
		{
			desc: "single valid rule",
			config: dynamic.IPAllowGroup{
				Rules: []dynamic.IPAllowList{
					{SourceRange: []string{"10.0.0.0/8"}},
				},
			},
		},
		{
			desc: "multiple valid rules",
			config: dynamic.IPAllowGroup{
				Rules: []dynamic.IPAllowList{
					{SourceRange: []string{"10.0.0.0/8"}},
					{SourceRange: []string{"192.168.0.0/16"}},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			handler, err := NewGroup(t.Context(), next, test.config, "test")

			if test.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, handler)
			}
		})
	}
}

func TestIPAllowGrouper_ServeHTTP(t *testing.T) {
	testCases := []struct {
		desc       string
		config     dynamic.IPAllowGroup
		remoteAddr string
		expected   int
	}{
		{
			desc: "IP matches first rule",
			config: dynamic.IPAllowGroup{
				Rules: []dynamic.IPAllowList{
					{SourceRange: []string{"10.0.0.0/8"}},
					{SourceRange: []string{"192.168.0.0/16"}},
				},
			},
			remoteAddr: "10.0.0.5:1234",
			expected:   http.StatusOK,
		},
		{
			desc: "IP matches second rule only (OR logic)",
			config: dynamic.IPAllowGroup{
				Rules: []dynamic.IPAllowList{
					{SourceRange: []string{"10.0.0.0/8"}},
					{SourceRange: []string{"192.168.0.0/16"}},
				},
			},
			remoteAddr: "192.168.1.50:1234",
			expected:   http.StatusOK,
		},
		{
			desc: "IP matches neither rule",
			config: dynamic.IPAllowGroup{
				Rules: []dynamic.IPAllowList{
					{SourceRange: []string{"10.0.0.0/8"}},
					{SourceRange: []string{"192.168.0.0/16"}},
				},
			},
			remoteAddr: "8.8.8.8:1234",
			expected:   http.StatusForbidden,
		},
		{
			desc: "single rule, authorized",
			config: dynamic.IPAllowGroup{
				Rules: []dynamic.IPAllowList{
					{SourceRange: []string{"20.20.20.20"}},
				},
			},
			remoteAddr: "20.20.20.20:1234",
			expected:   http.StatusOK,
		},
		{
			desc: "single rule, not authorized",
			config: dynamic.IPAllowGroup{
				Rules: []dynamic.IPAllowList{
					{SourceRange: []string{"20.20.20.20"}},
				},
			},
			remoteAddr: "20.20.20.21:1234",
			expected:   http.StatusForbidden,
		},
		{
			desc: "custom rejectStatusCode used when no rule matches",
			config: dynamic.IPAllowGroup{
				Rules: []dynamic.IPAllowList{
					{SourceRange: []string{"10.0.0.0/8"}, RejectStatusCode: http.StatusUnauthorized},
				},
			},
			remoteAddr: "8.8.8.8:1234",
			expected:   http.StatusUnauthorized,
		},
		{
			desc: "three rules: IP matches third rule",
			config: dynamic.IPAllowGroup{
				Rules: []dynamic.IPAllowList{
					{SourceRange: []string{"10.0.0.0/8"}},
					{SourceRange: []string{"192.168.0.0/16"}},
					{SourceRange: []string{"172.16.0.0/12"}},
				},
			},
			remoteAddr: "172.20.10.5:1234",
			expected:   http.StatusOK,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			handler, err := NewGroup(t.Context(), next, test.config, "test")
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			req.RemoteAddr = test.remoteAddr

			handler.ServeHTTP(recorder, req)

			assert.Equal(t, test.expected, recorder.Code)
		})
	}
}
