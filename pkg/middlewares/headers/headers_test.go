package headers

// Middleware tests based on https://github.com/unrolled/secure

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

func TestNew_withoutOptions(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	mid, err := New(context.Background(), next, dynamic.Headers{}, "testing")
	require.Errorf(t, err, "headers configuration not valid")

	assert.Nil(t, mid)
}

func TestNew_allowedHosts(t *testing.T) {
	testCases := []struct {
		desc     string
		fromHost string
		expected int
	}{
		{
			desc:     "Should accept the request when given a host that is in the list",
			fromHost: "foo.com",
			expected: http.StatusOK,
		},
		{
			desc:     "Should refuse the request when no host is given",
			fromHost: "",
			expected: http.StatusInternalServerError,
		},
		{
			desc:     "Should refuse the request when no matching host is given",
			fromHost: "boo.com",
			expected: http.StatusInternalServerError,
		},
	}

	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	cfg := dynamic.Headers{
		AllowedHosts: []string{"foo.com", "bar.com"},
	}

	mid, err := New(context.Background(), emptyHandler, cfg, "foo")
	require.NoError(t, err)

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/foo", nil)
			req.Host = test.fromHost

			rw := httptest.NewRecorder()

			mid.ServeHTTP(rw, req)

			assert.Equal(t, test.expected, rw.Code)
		})
	}
}

func TestNew_customHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	cfg := dynamic.Headers{
		CustomRequestHeaders: map[string]string{
			"X-Custom-Request-Header": "test_request",
		},
		CustomResponseHeaders: map[string]string{
			"X-Custom-Response-Header": "test_response",
		},
	}

	mid, err := New(context.Background(), next, cfg, "testing")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)

	rw := httptest.NewRecorder()

	mid.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "test_request", req.Header.Get("X-Custom-Request-Header"))
	assert.Equal(t, "test_response", rw.Header().Get("X-Custom-Response-Header"))
}

func Test_headers_getTracingInformation(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mid := &headers{
		handler: next,
		name:    "testing",
	}

	name, trace := mid.GetTracingInformation()

	assert.Equal(t, "testing", name)
	assert.Equal(t, tracing.SpanKindNoneEnum, trace)
}
