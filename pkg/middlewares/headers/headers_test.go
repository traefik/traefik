package headers

// Middleware tests based on https://github.com/unrolled/secure

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/middlewares/addprefix"
	"github.com/containous/traefik/pkg/middlewares/replacepath"
	"github.com/containous/traefik/pkg/middlewares/stripprefix"
	"github.com/containous/traefik/pkg/testhelpers"
	"github.com/containous/traefik/pkg/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomRequestHeader(t *testing.T) {
	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	header := NewHeader(emptyHandler, config.Headers{
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

func TestCustomRequestHeaderEmptyValue(t *testing.T) {
	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	header := NewHeader(emptyHandler, config.Headers{
		CustomRequestHeaders: map[string]string{
			"X-Custom-Request-Header": "test_request",
		},
	})

	res := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "/foo", nil)

	header.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "test_request", req.Header.Get("X-Custom-Request-Header"))

	header = NewHeader(emptyHandler, config.Headers{
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
	header, err := New(context.Background(), emptyHandler, config.Headers{
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
		secureMiddleware *SecureHeader
		expected         int
	}{
		{
			desc: "http should return a 301",
			host: "http://powpow.example.com",
			secureMiddleware: NewSecure(next, config.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusMovedPermanently,
		},
		{
			desc: "http sub domain should return a 301",
			host: "http://www.powpow.example.com",
			secureMiddleware: NewSecure(next, config.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusMovedPermanently,
		},
		{
			desc: "https should return a 200",
			host: "https://powpow.example.com",
			secureMiddleware: NewSecure(next, config.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusOK,
		},
		{
			desc: "https sub domain should return a 301",
			host: "https://www.powpow.example.com",
			secureMiddleware: NewSecure(next, config.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusMovedPermanently,
		},
		{
			desc: "http without force host and sub domain should return a 301",
			host: "http://www.powpow.example.com",
			secureMiddleware: NewSecure(next, config.Headers{
				SSLRedirect:  true,
				SSLForceHost: false,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusMovedPermanently,
		},
		{
			desc: "https without force host and sub domain should return a 301",
			host: "https://www.powpow.example.com",
			secureMiddleware: NewSecure(next, config.Headers{
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
			header: NewHeader(emptyHandler, config.Headers{
				AccessControlAllowMethods: []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOrigin:  "origin-list-or-null",
				AccessControlMaxAge:       600,
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
			header: NewHeader(emptyHandler, config.Headers{
				AccessControlAllowMethods: []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOrigin:  "*",
				AccessControlMaxAge:       600,
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
			header: NewHeader(emptyHandler, config.Headers{
				AccessControlAllowMethods:     []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOrigin:      "*",
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
			header: NewHeader(emptyHandler, config.Headers{
				AccessControlAllowMethods: []string{"GET", "OPTIONS", "PUT"},
				AccessControlAllowOrigin:  "*",
				AccessControlAllowHeaders: []string{"origin", "X-Forwarded-For"},
				AccessControlMaxAge:       600,
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

	_, err := New(context.Background(), next, config.Headers{}, "testing")
	require.Errorf(t, err, "headers configuration not valid")
}

func TestCustomHeaderHandler(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	header, _ := New(context.Background(), next, config.Headers{
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

	testCases := []struct {
		desc           string
		header         *Header
		requestHeaders http.Header
		expected       http.Header
	}{
		{
			desc: "Test Simple Request",
			header: NewHeader(emptyHandler, config.Headers{
				AccessControlAllowOrigin: "origin-list-or-null",
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
			header: NewHeader(emptyHandler, config.Headers{
				AccessControlAllowOrigin: "*",
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
			header: NewHeader(emptyHandler, config.Headers{
				AccessControlAllowOrigin: "origin-list-or-null",
			}),
			requestHeaders: map[string][]string{},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"null"},
			},
		},
		{
			desc:           "Not Defined origin Request",
			header:         NewHeader(emptyHandler, config.Headers{}),
			requestHeaders: map[string][]string{},
			expected:       map[string][]string{},
		},
		{
			desc: "Allow Credentials Request",
			header: NewHeader(emptyHandler, config.Headers{
				AccessControlAllowOrigin:      "*",
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
			header: NewHeader(emptyHandler, config.Headers{
				AccessControlAllowOrigin:   "*",
				AccessControlExposeHeaders: []string{"origin", "X-Forwarded-For"},
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
			header: NewHeader(emptyHandler, config.Headers{
				AccessControlAllowOrigin: "origin-list-or-null",
				AddVaryHeader:            true,
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
			header: NewHeader(nonEmptyHandler, config.Headers{
				AccessControlAllowOrigin: "origin-list-or-null",
				AddVaryHeader:            true,
			}),
			requestHeaders: map[string][]string{
				"Origin": {"https://foo.bar.org"},
			},
			expected: map[string][]string{
				"Access-Control-Allow-Origin": {"https://foo.bar.org"},
				"Vary":                        {"Testing,Origin"},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			req := testhelpers.MustNewRequest(http.MethodGet, "/foo", nil)
			req.Header = test.requestHeaders

			rw := httptest.NewRecorder()
			test.header.ServeHTTP(rw, req)
			err := test.header.ModifyResponseHeaders(rw.Result())
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
			header: NewHeader(emptyHandler, config.Headers{
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
			header: NewHeader(emptyHandler, config.Headers{
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
			err := test.header.ModifyResponseHeaders(rw.Result())
			require.NoError(t, err)
			assert.Equal(t, test.expected, rw.Result().Header)
		})
	}
}

func TestSSLRedirectWithModifiedRequest(t *testing.T) {
	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	testCases := []struct {
		desc          string
		addPrefix     bool
		replacePrefix bool
		stripPrefix   bool
		url           string
		key           string
		expected      string
	}{
		{
			desc:        "StripPrefix",
			stripPrefix: true,
			url:         "http://powpow.example.com/foo",
			key:         "/bacon/foo",
			expected:    "/bacon/foo",
		},
		{
			desc:      "AddPrefix",
			addPrefix: true,
			url:       "http://powpow.example.com/bacon/foo",
			key:       "/bacon",
			expected:  "/foo",
		},
		{
			desc:          "ReplacePrefix",
			replacePrefix: true,
			url:           "http://powpow.example.com/foo",
			key:           "/bacon/foo",
			expected:      "/bacon/foo",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {

			headerMiddleware := NewSecure(emptyHandler, config.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			})

			req := testhelpers.MustNewRequest(http.MethodGet, test.url, nil)
			switch {
			case test.stripPrefix:
				req = req.WithContext(context.WithValue(req.Context(), stripprefix.StripPrefixKey, test.key))
			case test.addPrefix:
				req = req.WithContext(context.WithValue(req.Context(), addprefix.AddPrefixKey, test.key))
			case test.replacePrefix:
				req = req.WithContext(context.WithValue(req.Context(), replacepath.ReplacePathKey, test.key))
			}
			req.RequestURI = req.URL.RequestURI()
			rw := httptest.NewRecorder()
			headerMiddleware.ServeHTTP(rw, req)
			err := headerMiddleware.ModifyResponseHeaders(rw.Result())
			require.NoError(t, err)
			returnedLocation, err := rw.Result().Location()
			require.NoError(t, err)
			assert.Equal(t, test.expected, returnedLocation.Path)
		})
	}
}
