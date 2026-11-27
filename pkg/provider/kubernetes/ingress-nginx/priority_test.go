package ingressnginx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/middlewares/requestdecorator"
	httpmuxer "github.com/traefik/traefik/v3/pkg/muxer/http"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
	netv1 "k8s.io/api/networking/v1"
)

func TestTranslate_locationPrecedence(t *testing.T) {
	prefix, exact := netv1.PathTypePrefix, netv1.PathTypeExact

	testCases := []struct {
		desc      string
		locations []*location
		path      string
		expected  string
	}{
		{
			desc: "short Exact wins over the root Prefix",
			locations: []*location{
				{Path: "/", PathType: &prefix, BackendName: "root"},
				{Path: "/login", PathType: &exact, BackendName: "login"},
			},
			path:     "/login",
			expected: "login",
		},
		{
			desc: "longer Prefix wins over a shorter Exact",
			locations: []*location{
				{Path: "/api", PathType: &exact, BackendName: "exact"},
				{Path: "/api/v1", PathType: &prefix, BackendName: "prefix"},
			},
			path:     "/api/v1",
			expected: "prefix",
		},
		{
			desc: "Exact wins over Prefix of the same length",
			locations: []*location{
				{Path: "/api", PathType: &prefix, BackendName: "prefix"},
				{Path: "/api", PathType: &exact, BackendName: "exact"},
			},
			path:     "/api",
			expected: "exact",
		},
		{
			desc: "Prefix still serves the other paths",
			locations: []*location{
				{Path: "/", PathType: &prefix, BackendName: "root"},
				{Path: "/login", PathType: &exact, BackendName: "login"},
			},
			path:     "/login/other",
			expected: "root",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			mc := &model{
				Backends: map[string]*backend{},
				Servers:  map[string]*server{"example.com": {Hostname: "example.com"}},
			}
			for i, loc := range test.locations {
				loc.LocationIndex = i
				mc.Backends[loc.BackendName] = &backend{Name: loc.BackendName}
				mc.Servers["example.com"].Locations = append(mc.Servers["example.com"].Locations, loc)
			}

			p := &Provider{NonTLSEntryPoints: []string{"web"}}
			conf := p.translate(t.Context(), mc)

			parser, err := httpmuxer.NewSyntaxParser()
			require.NoError(t, err)
			muxer := httpmuxer.NewMuxer(parser, nil)

			var served string
			for _, rt := range conf.HTTP.Routers {
				if rt.TLS != nil {
					continue
				}

				service := rt.Service

				// The router manager applies the rule length when no priority is set.
				priority := rt.Priority
				if priority == 0 {
					priority = httpmuxer.GetRulePriority(rt.Rule)
				}

				require.NoError(t, muxer.AddRoute(rt.Rule, rt.RuleSyntax, priority, "", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
					served = service
				})))
			}

			req := testhelpers.MustNewRequest(http.MethodGet, "http://example.com"+test.path, http.NoBody)
			requestdecorator.New(nil).ServeHTTP(httptest.NewRecorder(), req, muxer.ServeHTTP)

			assert.Equal(t, test.expected, served)
		})
	}
}
