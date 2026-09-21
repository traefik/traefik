package ingressnginx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/requestdecorator"
	httpmuxer "github.com/traefik/traefik/v3/pkg/muxer/http"
)

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

func TestBuildRuleLookaheadRequests(t *testing.T) {
	t.Parallel()

	loc := &location{
		Path:     "/a/((?!internal).*)",
		UseRegex: true,
		Aliases:  []string{"alias.localhost"},
	}
	resolveNegativeLookahead(loc)
	rule, originalRule := buildRule("example.localhost", loc)

	parser, err := httpmuxer.NewSyntaxParser()
	require.NoError(t, err)
	mux := httpmuxer.NewMuxer(parser, nil)
	err = mux.AddRoute(rule, "v3", pinnedPriority(rule, originalRule), "kubernetes", http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNoContent)
	}))
	require.NoError(t, err)

	testCases := []struct {
		path         string
		expectedCode int
	}{
		{path: "/a/public", expectedCode: http.StatusNoContent},
		{path: "/A/PUBLIC", expectedCode: http.StatusNoContent},
		{path: "/a/", expectedCode: http.StatusNoContent},
		{path: "/a/internal", expectedCode: http.StatusNotFound},
		{path: "/a/internalMore", expectedCode: http.StatusNotFound},
		{path: "/A/INTERNAL", expectedCode: http.StatusNotFound},
		{path: "/a/x/internal", expectedCode: http.StatusNoContent},
		{path: "/b/public", expectedCode: http.StatusNotFound},
	}
	for _, host := range []string{"example.localhost", "alias.localhost", "other.localhost"} {
		for _, test := range testCases {
			t.Run(host+test.path, func(t *testing.T) {
				t.Parallel()

				expectedCode := test.expectedCode
				if host == "other.localhost" {
					expectedCode = http.StatusNotFound
				}

				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "http://"+host+test.path, nil)
				requestdecorator.New(nil).ServeHTTP(recorder, req, mux.ServeHTTP)

				assert.Equal(t, expectedCode, recorder.Code)
			})
		}
	}
}
