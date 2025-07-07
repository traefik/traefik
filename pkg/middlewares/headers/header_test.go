package headers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestNewHeader_customRequestHeader(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      dynamic.Headers
		expected http.Header
	}{
		{
			desc: "adds a header",
			cfg: dynamic.Headers{
				CustomRequestHeaders: map[string]string{
					"X-Custom-Request-Header": "test_request",
				},
			},
			expected: http.Header{"Foo": []string{"bar"}, "X-Custom-Request-Header": []string{"test_request"}},
		},
		{
			desc: "delete a header",
			cfg: dynamic.Headers{
				CustomRequestHeaders: map[string]string{
					"X-Forwarded-For":         "",
					"X-Custom-Request-Header": "",
					"Foo":                     "",
				},
			},
			expected: http.Header{
				"X-Forwarded-For": nil,
			},
		},
		{
			desc: "override a header",
			cfg: dynamic.Headers{
				CustomRequestHeaders: map[string]string{
					"Foo": "test",
				},
			},
			expected: http.Header{"Foo": []string{"test"}},
		},
	}

	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			mid, err := NewHeader(emptyHandler, test.cfg)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/foo", nil)
			req.Header.Set("Foo", "bar")

			rw := httptest.NewRecorder()

			mid.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusOK, rw.Code)
			assert.Equal(t, test.expected, req.Header)
		})
	}
}

func TestNewHeader_customRequestHeader_Host(t *testing.T) {
	testCases := []struct {
		desc            string
		customHeaders   map[string]string
		expectedHost    string
		expectedURLHost string
	}{
		{
			desc:            "standard Host header",
			customHeaders:   map[string]string{},
			expectedHost:    "example.org",
			expectedURLHost: "example.org",
		},
		{
			desc: "custom Host header",
			customHeaders: map[string]string{
				"Host": "example.com",
			},
			expectedHost:    "example.com",
			expectedURLHost: "example.org",
		},
	}

	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			mid, err := NewHeader(emptyHandler, dynamic.Headers{CustomRequestHeaders: test.customHeaders})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "http://example.org/foo", nil)

			rw := httptest.NewRecorder()

			mid.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusOK, rw.Code)
			assert.Equal(t, test.expectedHost, req.Host)
			assert.Equal(t, test.expectedURLHost, req.URL.Host)
		})
	}
}

func TestNewHeader_CORSPreflights(t *testing.T) {
	testCases := []struct {
		desc           string
		cfg            dynamic.Headers
		requestHeaders http.Header
		expected       http.Header
	}{
		{
			desc: "Test Simple Preflight",
			cfg: dynamic.Headers{
				AccessControlAllowMethods:    []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOriginList: []string{"https://foo.bar.org"},
				AccessControlMaxAge:          600,
			},
			requestHeaders: map[string][]string{
				"Access-Control-Request-Headers": {"origin"},
				"Access-Control-Request-Method":  {"GET", "OPTIONS"},
				"Origin":                         {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Content-Length":               {"0"},
				"Access-Control-Allow-Origin":  {"https://foo.bar.org"},
				"Access-Control-Max-Age":       {"600"},
				"Access-Control-Allow-Methods": {"GET,OPTIONS,PUT"},
			},
		},
		{
			desc: "Wildcard origin Preflight",
			cfg: dynamic.Headers{
				AccessControlAllowMethods:    []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOriginList: []string{"*"},
				AccessControlMaxAge:          600,
			},
			requestHeaders: map[string][]string{
				"Access-Control-Request-Headers": {"origin"},
				"Access-Control-Request-Method":  {"GET", "OPTIONS"},
				"Origin":                         {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Content-Length":               {"0"},
				"Access-Control-Allow-Origin":  {"*"},
				"Access-Control-Max-Age":       {"600"},
				"Access-Control-Allow-Methods": {"GET,OPTIONS,PUT"},
			},
		},
		{
			desc: "Allow Credentials Preflight",
			cfg: dynamic.Headers{
				AccessControlAllowMethods:     []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOriginList:  []string{"*"},
				AccessControlAllowCredentials: true,
				AccessControlMaxAge:           600,
			},
			requestHeaders: map[string][]string{
				"Access-Control-Request-Headers": {"origin"},
				"Access-Control-Request-Method":  {"GET", "OPTIONS"},
				"Origin":                         {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Content-Length":                   {"0"},
				"Access-Control-Allow-Origin":      {"*"},
				"Access-Control-Max-Age":           {"600"},
				"Access-Control-Allow-Methods":     {"GET,OPTIONS,PUT"},
				"Access-Control-Allow-Credentials": {"true"},
			},
		},
		{
			desc: "Allow Headers Preflight",
			cfg: dynamic.Headers{
				AccessControlAllowMethods:    []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOriginList: []string{"*"},
				AccessControlAllowHeaders:    []string{"origin", "X-Forwarded-For"},
				AccessControlMaxAge:          600,
			},
			requestHeaders: map[string][]string{
				"Access-Control-Request-Headers": {"origin"},
				"Access-Control-Request-Method":  {"GET", "OPTIONS"},
				"Origin":                         {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Content-Length":               {"0"},
				"Access-Control-Allow-Origin":  {"*"},
				"Access-Control-Max-Age":       {"600"},
				"Access-Control-Allow-Methods": {"GET,OPTIONS,PUT"},
				"Access-Control-Allow-Headers": {"origin,X-Forwarded-For"},
			},
		},
		{
			desc: "No Request Headers Preflight",
			cfg: dynamic.Headers{
				AccessControlAllowMethods:    []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOriginList: []string{"*"},
				AccessControlAllowHeaders:    []string{"origin", "X-Forwarded-For"},
				AccessControlMaxAge:          600,
			},
			requestHeaders: map[string][]string{
				"Access-Control-Request-Method": {"GET", "OPTIONS"},
				"Origin":                        {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Content-Length":               {"0"},
				"Access-Control-Allow-Origin":  {"*"},
				"Access-Control-Max-Age":       {"600"},
				"Access-Control-Allow-Methods": {"GET,OPTIONS,PUT"},
				"Access-Control-Allow-Headers": {"origin,X-Forwarded-For"},
			},
		},
	}

	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			mid, err := NewHeader(emptyHandler, test.cfg)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodOptions, "/foo", nil)
			req.Header = test.requestHeaders

			rw := httptest.NewRecorder()

			mid.ServeHTTP(rw, req)

			assert.Equal(t, test.expected, rw.Result().Header)
		})
	}
}

func TestNewHeader_CORSResponses(t *testing.T) {
	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	testCases := []struct {
		desc           string
		next           http.Handler
		cfg            dynamic.Headers
		requestHeaders http.Header
		expected       http.Header
		expectedError  bool
	}{
		{
			desc: "Test Simple Request",
			next: emptyHandler,
			cfg: dynamic.Headers{
				AccessControlAllowOriginList: []string{"https://foo.bar.org"},
			},
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"https://foo.bar.org"},
			},
		},
		{
			desc: "Wildcard origin Request",
			next: emptyHandler,
			cfg: dynamic.Headers{
				AccessControlAllowOriginList: []string{"*"},
			},
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"*"},
			},
		},
		{
			desc: "Regexp Origin Request",
			next: emptyHandler,
			cfg: dynamic.Headers{
				AccessControlAllowOriginListRegex: []string{"^https?://([a-z]+)\\.bar\\.org$"},
			},
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"https://foo.bar.org"},
			},
		},
		{
			desc: "Partial Regexp Origin Request",
			next: emptyHandler,
			cfg: dynamic.Headers{
				AccessControlAllowOriginListRegex: []string{"([a-z]+)\\.bar"},
			},
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"https://foo.bar.org"},
			},
		},
		{
			desc: "Regexp Malformed Origin Request",
			next: emptyHandler,
			cfg: dynamic.Headers{
				AccessControlAllowOriginListRegex: []string{"a(b"},
			},
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expectedError: true,
		},
		{
			desc: "Regexp Origin Request without matching",
			next: emptyHandler,
			cfg: dynamic.Headers{
				AccessControlAllowOriginListRegex: []string{"([a-z]+)\\.bar\\.org"},
			},
			requestHeaders: map[string][]string{
				"Origin": {"https://bar.org"},
			},
			expected: map[string][]string{},
		},
		{
			desc: "Empty origin Request",
			next: emptyHandler,
			cfg: dynamic.Headers{
				AccessControlAllowOriginList: []string{"https://foo.bar.org"},
			},
			requestHeaders: map[string][]string{},
			expected:       map[string][]string{},
		},
		{
			desc:           "Not Defined origin Request",
			next:           emptyHandler,
			requestHeaders: map[string][]string{},
			expected:       map[string][]string{},
		},
		{
			desc: "Allow Credentials Request",
			next: emptyHandler,
			cfg: dynamic.Headers{
				AccessControlAllowOriginList:  []string{"*"},
				AccessControlAllowCredentials: true,
			},
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin":      {"*"},
				"Access-Control-Allow-Credentials": {"true"},
			},
		},
		{
			desc: "Expose Headers Request",
			next: emptyHandler,
			cfg: dynamic.Headers{
				AccessControlAllowOriginList: []string{"*"},
				AccessControlExposeHeaders:   []string{"origin", "X-Forwarded-For"},
			},
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin":   {"*"},
				"Access-Control-Expose-Headers": {"origin,X-Forwarded-For"},
			},
		},
		{
			desc: "Test Simple Request with Vary Headers",
			next: emptyHandler,
			cfg: dynamic.Headers{
				AccessControlAllowOriginList: []string{"https://foo.bar.org"},
				AddVaryHeader:                true,
			},
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"https://foo.bar.org"},
				"Vary":                        {"Origin"},
			},
		},
		{
			desc: "Test Simple Request with Vary Headers and non-empty response",
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// nonEmptyHandler
				w.Header().Set("Vary", "Testing")
				w.WriteHeader(http.StatusOK)
			}),
			cfg: dynamic.Headers{
				AccessControlAllowOriginList: []string{"https://foo.bar.org"},
				AddVaryHeader:                true,
			},
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"https://foo.bar.org"},
				"Vary":                        {"Testing,Origin"},
			},
		},
		{
			desc: "Test Simple Request with Vary Headers and existing vary:origin response",
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// existingOriginHandler
				w.Header().Set("Vary", "Origin")
				w.WriteHeader(http.StatusOK)
			}),
			cfg: dynamic.Headers{
				AccessControlAllowOriginList: []string{"https://foo.bar.org"},
				AddVaryHeader:                true,
			},
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"https://foo.bar.org"},
				"Vary":                        {"Origin"},
			},
		},
		{
			desc: "Test Simple Request with non-empty response: set ACAO",
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// existingAccessControlAllowOriginHandlerSet
				w.Header().Set("Access-Control-Allow-Origin", "http://foo.bar.org")
				w.WriteHeader(http.StatusOK)
			}),
			cfg: dynamic.Headers{
				AccessControlAllowOriginList: []string{"*"},
			},
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"*"},
			},
		},
		{
			desc: "Test Simple Request with non-empty response: add ACAO",
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// existingAccessControlAllowOriginHandlerAdd
				w.Header().Add("Access-Control-Allow-Origin", "http://foo.bar.org")
				w.WriteHeader(http.StatusOK)
			}),
			cfg: dynamic.Headers{
				AccessControlAllowOriginList: []string{"*"},
			},
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"*"},
			},
		},
		{
			desc: "Test Simple CustomRequestHeaders Not Hijacked by CORS",
			next: emptyHandler,
			cfg: dynamic.Headers{
				CustomRequestHeaders: map[string]string{"foo": "bar"},
			},
			requestHeaders: map[string][]string{
				"Access-Control-Request-Headers": {"origin"},
				"Access-Control-Request-Method":  {"GET", "OPTIONS"},
				"Origin":                         {"https://foo.bar.org"},
			},
			expected: map[string][]string{},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			mid, err := NewHeader(test.next, test.cfg)
			if test.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/foo", nil)
			req.Header = test.requestHeaders

			rw := httptest.NewRecorder()

			mid.ServeHTTP(rw, req)

			assert.Equal(t, test.expected, rw.Result().Header)
		})
	}
}

func TestNewHeader_customResponseHeaders(t *testing.T) {
	testCases := []struct {
		desc     string
		config   map[string]string
		expected http.Header
	}{
		{
			desc: "Test Simple Response",
			config: map[string]string{
				"Testing":  "foo",
				"Testing2": "bar",
			},
			expected: map[string][]string{
				"Foo":      {"bar"},
				"Testing":  {"foo"},
				"Testing2": {"bar"},
			},
		},
		{
			desc: "empty Custom Header",
			config: map[string]string{
				"Testing":  "foo",
				"Testing2": "",
			},
			expected: map[string][]string{
				"Foo":     {"bar"},
				"Testing": {"foo"},
			},
		},
		{
			desc: "Deleting Custom Header",
			config: map[string]string{
				"Testing": "foo",
				"Foo":     "",
			},
			expected: map[string][]string{
				"Testing": {"foo"},
			},
		},
	}

	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Foo", "bar")
		w.WriteHeader(http.StatusOK)
	})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			mid, err := NewHeader(emptyHandler, dynamic.Headers{CustomResponseHeaders: test.config})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/foo", nil)

			rw := httptest.NewRecorder()

			mid.ServeHTTP(rw, req)

			assert.Equal(t, test.expected, rw.Result().Header)
		})
	}
}
