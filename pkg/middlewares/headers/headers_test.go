package headers

// Middleware tests based on https://github.com/unrolled/secure

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/testhelpers"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomRequestHeader(t *testing.T) {
	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	header := NewHeader(emptyHandler, dynamic.Headers{
		CustomRequestHeaders: map[string]string{
			"X-Custom-Request-Header": "test_request",
		},
	})

	res := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "/foo", nil)

	header.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "test_request", req.Header.Get("X-Custom-Request-Header"))
}

func TestCustomRequestHeader_Host(t *testing.T) {
	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

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

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			header := NewHeader(emptyHandler, dynamic.Headers{
				CustomRequestHeaders: test.customHeaders,
			})

			res := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "http://example.org/foo", nil)
			require.NoError(t, err)

			header.ServeHTTP(res, req)

			assert.Equal(t, http.StatusOK, res.Code)
			assert.Equal(t, test.expectedHost, req.Host)
			assert.Equal(t, test.expectedURLHost, req.URL.Host)
		})
	}
}

func TestCustomRequestHeaderEmptyValue(t *testing.T) {
	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	header := NewHeader(emptyHandler, dynamic.Headers{
		CustomRequestHeaders: map[string]string{
			"X-Custom-Request-Header": "test_request",
		},
	})

	res := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "/foo", nil)

	header.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "test_request", req.Header.Get("X-Custom-Request-Header"))

	header = NewHeader(emptyHandler, dynamic.Headers{
		CustomRequestHeaders: map[string]string{
			"X-Custom-Request-Header": "",
		},
	})

	header.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "", req.Header.Get("X-Custom-Request-Header"))
}

func TestSecureHeader(t *testing.T) {
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

	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	header, err := New(context.Background(), emptyHandler, dynamic.Headers{
		AllowedHosts: []string{"foo.com", "bar.com"},
	}, "foo")
	require.NoError(t, err)

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			res := httptest.NewRecorder()
			req := testhelpers.MustNewRequest(http.MethodGet, "/foo", nil)
			req.Host = test.fromHost
			header.ServeHTTP(res, req)
			assert.Equal(t, test.expected, res.Code)
		})
	}
}

func TestSSLForceHost(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write([]byte("OK"))
	})

	testCases := []struct {
		desc             string
		host             string
		secureMiddleware *secureHeader
		expected         int
	}{
		{
			desc: "http should return a 301",
			host: "http://powpow.example.com",
			secureMiddleware: newSecure(next, dynamic.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusMovedPermanently,
		},
		{
			desc: "http sub domain should return a 301",
			host: "http://www.powpow.example.com",
			secureMiddleware: newSecure(next, dynamic.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusMovedPermanently,
		},
		{
			desc: "https should return a 200",
			host: "https://powpow.example.com",
			secureMiddleware: newSecure(next, dynamic.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusOK,
		},
		{
			desc: "https sub domain should return a 301",
			host: "https://www.powpow.example.com",
			secureMiddleware: newSecure(next, dynamic.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusMovedPermanently,
		},
		{
			desc: "http without force host and sub domain should return a 301",
			host: "http://www.powpow.example.com",
			secureMiddleware: newSecure(next, dynamic.Headers{
				SSLRedirect:  true,
				SSLForceHost: false,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusMovedPermanently,
		},
		{
			desc: "https without force host and sub domain should return a 301",
			host: "https://www.powpow.example.com",
			secureMiddleware: newSecure(next, dynamic.Headers{
				SSLRedirect:  true,
				SSLForceHost: false,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusOK,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			req := testhelpers.MustNewRequest(http.MethodGet, test.host, nil)

			rw := httptest.NewRecorder()
			test.secureMiddleware.ServeHTTP(rw, req)

			assert.Equal(t, test.expected, rw.Result().StatusCode)
		})
	}
}

func TestCORSPreflights(t *testing.T) {
	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	testCases := []struct {
		desc           string
		header         *Header
		requestHeaders http.Header
		expected       http.Header
	}{
		{
			desc: "Test Simple Preflight",
			header: NewHeader(emptyHandler, dynamic.Headers{
				AccessControlAllowMethods:    []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOriginList: []string{"https://foo.bar.org"},
				AccessControlMaxAge:          600,
			}),
			requestHeaders: map[string][]string{
				"Access-Control-Request-Headers": {"origin"},
				"Access-Control-Request-Method":  {"GET", "OPTIONS"},
				"Origin":                         {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin":  {"https://foo.bar.org"},
				"Access-Control-Max-Age":       {"600"},
				"Access-Control-Allow-Methods": {"GET,OPTIONS,PUT"},
			},
		},
		{
			desc: "Wildcard origin Preflight",
			header: NewHeader(emptyHandler, dynamic.Headers{
				AccessControlAllowMethods:    []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOriginList: []string{"*"},
				AccessControlMaxAge:          600,
			}),
			requestHeaders: map[string][]string{
				"Access-Control-Request-Headers": {"origin"},
				"Access-Control-Request-Method":  {"GET", "OPTIONS"},
				"Origin":                         {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin":  {"*"},
				"Access-Control-Max-Age":       {"600"},
				"Access-Control-Allow-Methods": {"GET,OPTIONS,PUT"},
			},
		},
		{
			desc: "Allow Credentials Preflight",
			header: NewHeader(emptyHandler, dynamic.Headers{
				AccessControlAllowMethods:     []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOriginList:  []string{"*"},
				AccessControlAllowCredentials: true,
				AccessControlMaxAge:           600,
			}),
			requestHeaders: map[string][]string{
				"Access-Control-Request-Headers": {"origin"},
				"Access-Control-Request-Method":  {"GET", "OPTIONS"},
				"Origin":                         {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin":      {"*"},
				"Access-Control-Max-Age":           {"600"},
				"Access-Control-Allow-Methods":     {"GET,OPTIONS,PUT"},
				"Access-Control-Allow-Credentials": {"true"},
			},
		},
		{
			desc: "Allow Headers Preflight",
			header: NewHeader(emptyHandler, dynamic.Headers{
				AccessControlAllowMethods:    []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOriginList: []string{"*"},
				AccessControlAllowHeaders:    []string{"origin", "X-Forwarded-For"},
				AccessControlMaxAge:          600,
			}),
			requestHeaders: map[string][]string{
				"Access-Control-Request-Headers": {"origin"},
				"Access-Control-Request-Method":  {"GET", "OPTIONS"},
				"Origin":                         {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin":  {"*"},
				"Access-Control-Max-Age":       {"600"},
				"Access-Control-Allow-Methods": {"GET,OPTIONS,PUT"},
				"Access-Control-Allow-Headers": {"origin,X-Forwarded-For"},
			},
		},
		{
			desc: "No Request Headers Preflight",
			header: NewHeader(emptyHandler, dynamic.Headers{
				AccessControlAllowMethods:    []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOriginList: []string{"*"},
				AccessControlAllowHeaders:    []string{"origin", "X-Forwarded-For"},
				AccessControlMaxAge:          600,
			}),
			requestHeaders: map[string][]string{
				"Access-Control-Request-Method": {"GET", "OPTIONS"},
				"Origin":                        {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin":  {"*"},
				"Access-Control-Max-Age":       {"600"},
				"Access-Control-Allow-Methods": {"GET,OPTIONS,PUT"},
				"Access-Control-Allow-Headers": {"origin,X-Forwarded-For"},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			req := testhelpers.MustNewRequest(http.MethodOptions, "/foo", nil)
			req.Header = test.requestHeaders

			rw := httptest.NewRecorder()
			test.header.ServeHTTP(rw, req)

			assert.Equal(t, test.expected, rw.Result().Header)
		})
	}
}

func TestEmptyHeaderObject(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	_, err := New(context.Background(), next, dynamic.Headers{}, "testing")
	require.Errorf(t, err, "headers configuration not valid")
}

func TestCustomHeaderHandler(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	header, _ := New(context.Background(), next, dynamic.Headers{
		CustomRequestHeaders: map[string]string{
			"X-Custom-Request-Header": "test_request",
		},
	}, "testing")

	res := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "/foo", nil)

	header.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "test_request", req.Header.Get("X-Custom-Request-Header"))
}

func TestGetTracingInformation(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	header := &headers{
		handler: next,
		name:    "testing",
	}

	name, trace := header.GetTracingInformation()

	assert.Equal(t, "testing", name)
	assert.Equal(t, tracing.SpanKindNoneEnum, trace)
}

func TestCORSResponses(t *testing.T) {
	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	nonEmptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Header().Set("Vary", "Testing") })
	existingOriginHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Header().Set("Vary", "Origin") })
	existingAccessControlAllowOriginHandlerSet := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://foo.bar.org")
	})
	existingAccessControlAllowOriginHandlerAdd := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "http://foo.bar.org")
	})

	testCases := []struct {
		desc           string
		header         *Header
		requestHeaders http.Header
		expected       http.Header
	}{
		{
			desc: "Test Simple Request",
			header: NewHeader(emptyHandler, dynamic.Headers{
				AccessControlAllowOriginList: []string{"https://foo.bar.org"},
			}),
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"https://foo.bar.org"},
			},
		},
		{
			desc: "Wildcard origin Request",
			header: NewHeader(emptyHandler, dynamic.Headers{
				AccessControlAllowOriginList: []string{"*"},
			}),
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"*"},
			},
		},
		{
			desc: "Empty origin Request",
			header: NewHeader(emptyHandler, dynamic.Headers{
				AccessControlAllowOriginList: []string{"https://foo.bar.org"},
			}),
			requestHeaders: map[string][]string{},
			expected:       map[string][]string{},
		},
		{
			desc:           "Not Defined origin Request",
			header:         NewHeader(emptyHandler, dynamic.Headers{}),
			requestHeaders: map[string][]string{},
			expected:       map[string][]string{},
		},
		{
			desc: "Allow Credentials Request",
			header: NewHeader(emptyHandler, dynamic.Headers{
				AccessControlAllowOriginList:  []string{"*"},
				AccessControlAllowCredentials: true,
			}),
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
			header: NewHeader(emptyHandler, dynamic.Headers{
				AccessControlAllowOriginList: []string{"*"},
				AccessControlExposeHeaders:   []string{"origin", "X-Forwarded-For"},
			}),
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
			header: NewHeader(emptyHandler, dynamic.Headers{
				AccessControlAllowOriginList: []string{"https://foo.bar.org"},
				AddVaryHeader:                true,
			}),
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
			header: NewHeader(nonEmptyHandler, dynamic.Headers{
				AccessControlAllowOriginList: []string{"https://foo.bar.org"},
				AddVaryHeader:                true,
			}),
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
			header: NewHeader(existingOriginHandler, dynamic.Headers{
				AccessControlAllowOriginList: []string{"https://foo.bar.org"},
				AddVaryHeader:                true,
			}),
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
			header: NewHeader(existingAccessControlAllowOriginHandlerSet, dynamic.Headers{
				AccessControlAllowOriginList: []string{"*"},
			}),
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"*"},
			},
		},
		{
			desc: "Test Simple Request with non-empty response: add ACAO",
			header: NewHeader(existingAccessControlAllowOriginHandlerAdd, dynamic.Headers{
				AccessControlAllowOriginList: []string{"*"},
			}),
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"*"},
			},
		}, {
			desc: "Test Simple CustomRequestHeaders Not Hijacked by CORS",
			header: NewHeader(emptyHandler, dynamic.Headers{
				CustomRequestHeaders: map[string]string{"foo": "bar"},
			}),
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
			req := testhelpers.MustNewRequest(http.MethodGet, "/foo", nil)
			req.Header = test.requestHeaders
			rw := httptest.NewRecorder()
			test.header.ServeHTTP(rw, req)
			res := rw.Result()
			res.Request = req
			err := test.header.PostRequestModifyResponseHeaders(res)
			require.NoError(t, err)
			assert.Equal(t, test.expected, rw.Result().Header)
		})
	}
}

func TestCustomResponseHeaders(t *testing.T) {
	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	testCases := []struct {
		desc     string
		header   *Header
		expected http.Header
	}{
		{
			desc: "Test Simple Response",
			header: NewHeader(emptyHandler, dynamic.Headers{
				CustomResponseHeaders: map[string]string{
					"Testing":  "foo",
					"Testing2": "bar",
				},
			}),
			expected: map[string][]string{
				"Testing":  {"foo"},
				"Testing2": {"bar"},
			},
		},
		{
			desc: "Deleting Custom Header",
			header: NewHeader(emptyHandler, dynamic.Headers{
				CustomResponseHeaders: map[string]string{
					"Testing":  "foo",
					"Testing2": "",
				},
			}),
			expected: map[string][]string{
				"Testing": {"foo"},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			req := testhelpers.MustNewRequest(http.MethodGet, "/foo", nil)
			rw := httptest.NewRecorder()
			test.header.ServeHTTP(rw, req)
			err := test.header.PostRequestModifyResponseHeaders(rw.Result())
			require.NoError(t, err)
			assert.Equal(t, test.expected, rw.Result().Header)
		})
	}
}
