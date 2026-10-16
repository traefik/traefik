package ingressnginx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/redirect"
)

func TestApplyFromToWwwRedirect(t *testing.T) {
	conf := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:     make(map[string]*dynamic.Router),
			Middlewares: make(map[string]*dynamic.Middleware),
		},
	}

	loc := &location{
		FromToWwwRedirect: &middlewareFromToWwwRedirect{
			ExtraRouterRule: `Host("www.example.com")`,
			TargetHostname:  "example.com",
		},
	}
	rt := &dynamic.Router{
		EntryPoints: []string{"web"},
		Service:     "backend",
		Middlewares: []string{"router-basic-auth"},
	}

	applyFromToWwwRedirect(loc, "router", rt, nil, conf)

	router := conf.HTTP.Routers["router-from-to-www-redirect"]
	require.NotNil(t, router)

	assert.Equal(t, unavailableServiceName, router.Service)
	assert.Equal(t, []string{"router-from-to-www-redirect"}, router.Middlewares)

	middleware := conf.HTTP.Middlewares["router-from-to-www-redirect"]
	require.NotNil(t, middleware.RedirectRegex)

	next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusTeapot)
	})

	handler, err := redirect.NewRedirectRegex(t.Context(), next, *middleware.RedirectRegex, "test")
	require.NoError(t, err)

	testCases := []struct {
		desc         string
		host         string
		target       string
		expectedCode int
		expectedLoc  string
	}{
		{
			desc:         "no port",
			host:         "www.example.com",
			target:       "/foo",
			expectedCode: http.StatusPermanentRedirect,
			expectedLoc:  "http://example.com/foo",
		},
		{
			desc:         "numeric port",
			host:         "www.example.com:8080",
			target:       "/foo",
			expectedCode: http.StatusPermanentRedirect,
			expectedLoc:  "http://example.com:8080/foo",
		},
		{
			desc:         "non-numeric port",
			host:         "www.example.com:x",
			target:       "/foo",
			expectedCode: http.StatusPermanentRedirect,
			expectedLoc:  "http://example.com/foo",
		},
		{
			desc:         "empty port",
			host:         "www.example.com:",
			target:       "/foo",
			expectedCode: http.StatusPermanentRedirect,
			expectedLoc:  "http://example.com/foo",
		},
		{
			desc:         "multiple ports",
			host:         "www.example.com:8080:90",
			target:       "/foo",
			expectedCode: http.StatusPermanentRedirect,
			expectedLoc:  "http://example.com:8080/foo",
		},
		{
			desc:         "port followed by a delimiter",
			host:         "www.example.com:8080;x",
			target:       "/foo",
			expectedCode: http.StatusPermanentRedirect,
			expectedLoc:  "http://example.com:8080/foo",
		},
		{
			desc:         "IPv6 literal",
			host:         "[::1]:8080",
			target:       "/foo",
			expectedCode: http.StatusPermanentRedirect,
			expectedLoc:  "http://example.com:8080/foo",
		},
		{
			desc:         "empty authority",
			host:         "",
			target:       "/foo",
			expectedCode: http.StatusTeapot,
		},
		{
			desc:         "root path",
			host:         "www.example.com",
			target:       "/",
			expectedCode: http.StatusPermanentRedirect,
			expectedLoc:  "http://example.com/",
		},
		{
			desc:         "trailing slash",
			host:         "www.example.com",
			target:       "/foo/",
			expectedCode: http.StatusPermanentRedirect,
			expectedLoc:  "http://example.com/foo",
		},
		{
			desc:         "trailing slash followed by a query string",
			host:         "www.example.com",
			target:       "/foo/?bar=baz",
			expectedCode: http.StatusPermanentRedirect,
			expectedLoc:  "http://example.com/foo/?bar=baz",
		},
		{
			desc:         "double trailing slash",
			host:         "www.example.com",
			target:       "/foo//",
			expectedCode: http.StatusPermanentRedirect,
			expectedLoc:  "http://example.com/foo/",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, test.target, nil)
			req.Host = test.host

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, test.expectedCode, recorder.Code)
			assert.Equal(t, test.expectedLoc, recorder.Header().Get("Location"))
		})
	}
}

func TestMakeTrailingGroupOptional(t *testing.T) {
	testCases := []struct {
		desc     string
		path     string
		expected string
	}{
		{
			desc:     "trailing capture group",
			path:     "/original/(.*)",
			expected: "/original(?:/(.*))?",
		},
		{
			desc:     "multiple groups wraps only the last one",
			path:     "/foo/(bar)/(.*)",
			expected: "/foo/(bar)(?:/(.*))?",
		},
		{
			desc:     "path starting with a capture group is left unchanged",
			path:     "/(.*)",
			expected: "/(.*)",
		},
		{
			desc:     "path starting with a capture group followed by another group is left unchanged",
			path:     "/(something)(/.+)",
			expected: "/(something)(/.+)",
		},
		{
			desc:     "documented ingress-nginx rewrite pattern is left unchanged",
			path:     "/something(/|$)(.*)",
			expected: "/something(/|$)(.*)",
		},
		{
			desc:     "literal after the group is left unchanged",
			path:     "/foo/(.*)/bar",
			expected: "/foo/(.*)/bar",
		},
		{
			desc:     "escaped literal after the group is left unchanged",
			path:     `/foo/(.*)\.json`,
			expected: `/foo/(.*)\.json`,
		},
		{
			desc:     "anchor after the group is left unchanged",
			path:     "/foo/(.*)$",
			expected: "/foo/(.*)$",
		},
		{
			desc:     "nested groups closing at the end",
			path:     "/foo/((a|b)/.*)",
			expected: "/foo(?:/((a|b)/.*))?",
		},
		{
			desc:     "unbalanced group is left unchanged",
			path:     "/foo/((.*)",
			expected: "/foo/((.*)",
		},
		{
			desc:     "path without capture group",
			path:     "/foo",
			expected: "/foo",
		},
		{
			desc:     "path with trailing slash",
			path:     "/foo/",
			expected: "/foo/",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			assert.Equal(t, test.expected, makeTrailingGroupOptional(test.path))
		})
	}
}

func TestTranslateLookaheadPriority(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc           string
		path           string
		translated     bool
		sslPassthrough bool
	}{
		{desc: "translated HTTP and TLS", path: "/a/((?!internal).*)", translated: true},
		{desc: "translated passthrough", path: "/a/((?!internal).*)", translated: true, sslPassthrough: true},
		{desc: "unchanged HTTP and TLS", path: "/a/(.*)"},
		{desc: "unchanged passthrough", path: "/a/(.*)", sslPassthrough: true},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			loc := &location{
				Path:           test.path,
				UseRegex:       true,
				BackendName:    "backend",
				Namespace:      "ns",
				IngressName:    "ing",
				SSLPassthrough: test.sslPassthrough,
				Canary: &canaryConfig{
					Header: "X-Canary", Cookie: "canary", Weight: 20, WeightTotal: 100, BackendName: "canary",
				},
			}
			_, originalRule := buildRule("example.localhost", loc)
			resolveNegativeLookahead(loc)

			p := &Provider{NonTLSEntryPoints: []string{"http"}, TLSEntryPoints: []string{"https"}}
			conf := p.translate(t.Context(), &model{
				Servers: map[string]*server{
					"example.localhost": {Hostname: "example.localhost", Locations: []*location{loc}},
				},
				Backends: map[string]*backend{
					"backend": {Name: "backend"},
					"canary":  {Name: "canary"},
				},
			})

			expectedCount := 6
			if test.sslPassthrough {
				expectedCount = 3
			}
			require.Len(t, conf.HTTP.Routers, expectedCount)

			originalRules := map[string]string{
				"":            originalRule,
				"-canary":     appendCanaryRule(originalRule, loc.Canary),
				"-non-canary": appendNonCanaryRule(originalRule, loc.Canary),
			}
			for suffix, original := range originalRules {
				key := "ns-ing-rule-0-path-0" + suffix
				router := conf.HTTP.Routers[key]
				require.NotNil(t, router)

				expectedPriority := 0
				if test.translated {
					expectedPriority = len(original)
				}
				assert.Equal(t, expectedPriority, router.Priority)

				tlsRouter := conf.HTTP.Routers[key+"-tls"]
				if test.sslPassthrough {
					assert.Nil(t, tlsRouter)
					continue
				}

				require.NotNil(t, tlsRouter)
				assert.Equal(t, router.Rule, tlsRouter.Rule)
				assert.Equal(t, expectedPriority, tlsRouter.Priority)
			}
		})
	}
}

func TestApplyFromToWwwRedirectPinnedPriority(t *testing.T) {
	t.Parallel()

	conf := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:     make(map[string]*dynamic.Router),
			Middlewares: make(map[string]*dynamic.Middleware),
		},
	}
	loc := &location{
		FromToWwwRedirect: &middlewareFromToWwwRedirect{
			ExtraRouterRule: `Host("www.example.com")`,
			TargetHostname:  "example.com",
		},
	}
	parent := &dynamic.Router{Priority: 91}

	applyFromToWwwRedirect(loc, "router", parent, nil, conf)

	router := conf.HTTP.Routers["router-from-to-www-redirect"]
	require.NotNil(t, router)
	assert.Zero(t, router.Priority)
	assert.Equal(t, loc.FromToWwwRedirect.ExtraRouterRule, router.Rule)
	assert.Equal(t, 91, parent.Priority)
}
