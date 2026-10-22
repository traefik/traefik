package ingressnginx

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/requestdecorator"
	httpmuxer "github.com/traefik/traefik/v3/pkg/muxer/http"
	netv1 "k8s.io/api/networking/v1"
)

func TestExactPathRouting(t *testing.T) {
	testCases := []struct {
		desc         string
		path         string
		prefix       string
		aliases      []string
		prefixCanary *canaryConfig
		exactCanary  *canaryConfig
		cookie       string
		expected     string
	}{
		{
			desc:     "enclosing prefix",
			path:     "/admin/report",
			prefix:   "/admin",
			expected: "exact",
		},
		{
			desc:     "same path",
			path:     "/admin",
			prefix:   "/admin",
			expected: "exact",
		},
		{
			desc:     "root",
			path:     "/",
			prefix:   "/",
			expected: "exact",
		},
		{
			desc:     "trailing slash",
			path:     "/admin/",
			prefix:   "/admin/",
			expected: "exact",
		},
		{
			desc:     "prefix with aliases",
			path:     "/admin/report",
			prefix:   "/admin",
			aliases:  []string{"alias.example.com", "another.example.com"},
			expected: "exact",
		},
		{
			desc:         "prefix canary",
			path:         "/admin/report",
			prefix:       "/admin",
			prefixCanary: &canaryConfig{BackendName: "canary", Weight: 20, WeightTotal: 100, Cookie: "route"},
			cookie:       "route=always",
			expected:     "exact",
		},
		{
			desc:         "prefix non-canary",
			path:         "/admin/report",
			prefix:       "/admin",
			prefixCanary: &canaryConfig{BackendName: "canary", Weight: 20, WeightTotal: 100, Cookie: "route"},
			cookie:       "route=never",
			expected:     "exact",
		},
		{
			desc:         "exact canary",
			path:         "/admin/report",
			prefix:       "/admin",
			prefixCanary: &canaryConfig{BackendName: "canary", Weight: 20, WeightTotal: 100, Cookie: "a-very-long-routing-cookie"},
			exactCanary:  &canaryConfig{BackendName: "canary", Weight: 20, WeightTotal: 100, Cookie: "route"},
			cookie:       "route=always; a-very-long-routing-cookie=always",
			expected:     "exact-canary",
		},
		{
			desc:         "exact non-canary",
			path:         "/admin/report",
			prefix:       "/admin",
			prefixCanary: &canaryConfig{BackendName: "canary", Weight: 20, WeightTotal: 100, Cookie: "a-very-long-routing-cookie"},
			exactCanary:  &canaryConfig{BackendName: "canary", Weight: 20, WeightTotal: 100, Cookie: "route"},
			cookie:       "route=never; a-very-long-routing-cookie=always",
			expected:     "exact",
		},
		{
			desc:     "already higher priority",
			path:     "/admin/" + strings.Repeat("a", 100),
			prefix:   "/admin",
			expected: "exact",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{NonTLSEntryPoints: []string{"web"}, TLSEntryPoints: []string{"websecure"}}
			conf := p.translate(t.Context(), &model{
				Backends: map[string]*backend{"prefix": {}, "exact": {}, "canary": {}},
				Servers: map[string]*server{
					"example.com": {
						Hostname: "example.com",
						Locations: []*location{
							{
								Path: test.prefix, PathType: new(netv1.PathTypePrefix),
								Namespace: "default", IngressName: "prefix", BackendName: "prefix",
								Aliases: test.aliases, Canary: test.prefixCanary,
							},
							{
								Path: test.path, PathType: new(netv1.PathTypeExact),
								Namespace: "default", IngressName: "exact", BackendName: "exact",
								Aliases: test.aliases, Canary: test.exactCanary,
								BasicAuth: &dynamic.BasicAuth{
									Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"},
								},
							},
						},
					},
				},
			})

			for _, entryPoint := range []string{"web", "websecure"} {
				t.Run(entryPoint, func(t *testing.T) {
					parser, err := httpmuxer.NewSyntaxParser()
					require.NoError(t, err)
					mux := httpmuxer.NewMuxer(parser, nil)
					var selected *dynamic.Router
					for _, name := range slices.Sorted(maps.Keys(conf.HTTP.Routers)) {
						router := conf.HTTP.Routers[name]
						if !slices.Contains(router.EntryPoints, entryPoint) {
							continue
						}
						handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
							selected = router
						})
						priority := router.Priority
						if priority == 0 {
							priority = httpmuxer.GetRulePriority(router.Rule)
						}
						require.NoError(t, mux.AddRoute(router.Rule, router.RuleSyntax, priority, "kubernetesingressnginx", handler))
					}

					handler, err := requestdecorator.WrapHandler(requestdecorator.New(nil))(mux)
					require.NoError(t, err)
					req := httptest.NewRequest(http.MethodGet, "http://example.com"+test.path, nil)
					req.Header.Set("Cookie", test.cookie)
					recorder := httptest.NewRecorder()
					handler.ServeHTTP(recorder, req)
					require.NotNil(t, selected)
					assert.Equal(t, test.expected, selected.Service)
					require.Len(t, selected.Middlewares, 1)
					assert.NotNil(t, conf.HTTP.Middlewares[selected.Middlewares[0]].BasicAuth)

					req = httptest.NewRequest(http.MethodGet, "http://example.com"+strings.TrimSuffix(test.prefix, "/")+"/other", nil)
					selected = nil
					recorder = httptest.NewRecorder()
					handler.ServeHTTP(recorder, req)
					assert.Equal(t, http.StatusOK, recorder.Code)
					require.NotNil(t, selected)
					expectedPrefixService := "prefix"
					if test.prefixCanary != nil {
						expectedPrefixService = "prefix-wrr"
					}
					assert.Equal(t, expectedPrefixService, selected.Service)
					assert.Empty(t, selected.Middlewares)

					for _, alias := range test.aliases {
						req = httptest.NewRequest(http.MethodGet, "http://"+alias+test.path, nil)
						selected = nil
						handler.ServeHTTP(httptest.NewRecorder(), req)
						require.NotNil(t, selected)
						assert.Equal(t, test.expected, selected.Service)
					}
				})
			}
		})
	}
}

func TestExactPathPriorityIsolation(t *testing.T) {
	p := Provider{NonTLSEntryPoints: []string{"web"}, TLSEntryPoints: []string{"websecure"}}
	conf := p.translate(t.Context(), &model{
		Backends: map[string]*backend{"backend": {}},
		Servers: map[string]*server{
			"mixed.example.com": {
				Hostname: "mixed.example.com",
				Locations: []*location{
					{
						Path: "/admin", PathType: new(netv1.PathTypePrefix), IngressName: "prefix", BackendName: "backend",
						FromToWwwRedirect: &middlewareFromToWwwRedirect{
							ExtraRouterRule: `Host("www.mixed.example.com")`,
							TargetHostname:  "mixed.example.com",
						},
						AppRoot:                &dynamic.AppRoot{Path: "/admin"},
						AppRootExtraRouterRule: `Host("mixed.example.com") && Path("/")`,
					},
					{Path: "/admin/x", PathType: new(netv1.PathTypeExact), IngressName: "short", BackendName: "backend"},
					{Path: "/admin/longer", PathType: new(netv1.PathTypeExact), IngressName: "long", BackendName: "backend"},
				},
			},
			"exact.example.com": {
				Hostname: "exact.example.com",
				Locations: []*location{
					{Path: "/", PathType: new(netv1.PathTypeExact), IngressName: "only-exact", BackendName: "backend"},
				},
			},
			"prefix.example.com": {
				Hostname: "prefix.example.com",
				Locations: []*location{
					{Path: "/", PathType: new(netv1.PathTypePrefix), IngressName: "only-prefix", BackendName: "backend"},
				},
			},
			"regex.example.com": {
				Hostname: "regex.example.com",
				Locations: []*location{
					{Path: "/", UseRegex: true, PathType: new(netv1.PathTypePrefix), IngressName: "regex-prefix", BackendName: "backend"},
					{Path: "/x", UseRegex: true, PathType: new(netv1.PathTypeExact), IngressName: "regex-exact", BackendName: "backend"},
				},
			},
		},
	})

	for _, suffix := range []string{"", "-tls"} {
		prefix := conf.HTTP.Routers["-prefix-rule-0-path-0"+suffix]
		require.NotNil(t, prefix)
		assert.Equal(t, httpmuxer.GetRulePriority(`Host("mixed.example.com") && (Path("/admin") || PathPrefix("/admin/"))`), prefix.Priority)
	}
	for name, router := range conf.HTTP.Routers {
		if name == "-prefix-rule-0-path-0" || name == "-prefix-rule-0-path-0-tls" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			assert.Zero(t, router.Priority)
		})
	}
}

func TestExactPathHostPrecedence(t *testing.T) {
	p := Provider{NonTLSEntryPoints: []string{"web"}}
	conf := p.translate(t.Context(), &model{
		Backends: map[string]*backend{"fallback": {}, "unused": {}, "named": {}},
		Servers: map[string]*server{
			"": {Locations: []*location{
				{Path: "/x", PathType: new(netv1.PathTypeExact), IngressName: "fallback", BackendName: "fallback"},
				{Path: "/unrelated/" + strings.Repeat("long", 30), PathType: new(netv1.PathTypePrefix), IngressName: "unused", BackendName: "unused"},
			}},
			"example.com": {Hostname: "example.com", Locations: []*location{
				{Path: "/x", PathType: new(netv1.PathTypePrefix), IngressName: "named", BackendName: "named", BasicAuth: &dynamic.BasicAuth{Users: []string{"test:hash"}}},
			}},
		},
	})
	parser, err := httpmuxer.NewSyntaxParser()
	require.NoError(t, err)
	mux := httpmuxer.NewMuxer(parser, nil)
	var selected *dynamic.Router
	for _, name := range slices.Sorted(maps.Keys(conf.HTTP.Routers)) {
		router := conf.HTTP.Routers[name]
		if !slices.Contains(router.EntryPoints, "web") {
			continue
		}
		priority := router.Priority
		if priority == 0 {
			priority = httpmuxer.GetRulePriority(router.Rule)
		}
		require.NoError(t, mux.AddRoute(router.Rule, router.RuleSyntax, priority, "kubernetesingressnginx", http.HandlerFunc(func(http.ResponseWriter, *http.Request) { selected = router })))
	}
	handler, err := requestdecorator.WrapHandler(requestdecorator.New(nil))(mux)
	require.NoError(t, err)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://example.com/x", nil))
	require.NotNil(t, selected)
	assert.Equal(t, "named", selected.Service)
	require.Len(t, selected.Middlewares, 1)
	assert.NotNil(t, conf.HTTP.Middlewares[selected.Middlewares[0]].BasicAuth)
}
