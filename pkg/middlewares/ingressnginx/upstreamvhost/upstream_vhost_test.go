package upstreamvhost

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestUpstreamVhost(t *testing.T) {
	testCases := []struct {
		desc         string
		config       dynamic.UpstreamVHost
		reqHost      string
		expectedHost string
		expectsError bool
	}{
		{
			desc: "empty vhost",
			config: dynamic.UpstreamVHost{
				VHost: "",
			},
			expectsError: true,
		},
		{
			desc: "static vhost",
			config: dynamic.UpstreamVHost{
				VHost: "backend.internal",
			},
			reqHost:      "site.example.com",
			expectedHost: "backend.internal",
		},
		{
			desc: "provider-supplied $service_name and $namespace",
			config: dynamic.UpstreamVHost{
				VHost: "$service_name.$namespace",
				Vars: map[string]string{
					"$service_name": "my-app",
					"$namespace":    "foo",
				},
			},
			reqHost:      "site.example.com",
			expectedHost: "my-app.foo",
		},
		{
			desc: "request-time $host",
			config: dynamic.UpstreamVHost{
				VHost: "$host",
			},
			reqHost:      "Site.Example.com:8080",
			expectedHost: "site.example.com",
		},
		{
			desc: "mix of static and request-time variables",
			config: dynamic.UpstreamVHost{
				VHost: "$service_name.$namespace.svc.cluster.local",
				Vars: map[string]string{
					"$service_name": "my-app",
					"$namespace":    "foo",
				},
			},
			reqHost:      "site.example.com",
			expectedHost: "my-app.foo.svc.cluster.local",
		},
		{
			desc: "unknown variable left as-is",
			config: dynamic.UpstreamVHost{
				VHost: "$does_not_exist",
			},
			reqHost:      "site.example.com",
			expectedHost: "$does_not_exist",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var gotHost string
			next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				gotHost = r.Host
			})

			mw, err := New(t.Context(), next, test.config, "upstream-vhost")
			if test.expectsError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "http://"+test.reqHost+"/", nil)
			req.Host = test.reqHost

			mw.ServeHTTP(httptest.NewRecorder(), req)

			assert.Equal(t, test.expectedHost, gotHost)
		})
	}
}
