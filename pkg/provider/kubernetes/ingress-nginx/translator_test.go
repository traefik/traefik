package ingressnginx

import (
	"net/http"
	"net/http/httptest"
	"slices"
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

// The upstream vhost rewrites the request Host for the backend. The auth
// middleware resolves $host in its auth-signin redirect from the incoming
// request, so it has to run before the rewrite, otherwise the client is
// redirected to the internal upstream host.
func TestApplyMiddlewaresUpstreamVhostAfterAuth(t *testing.T) {
	conf := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:     make(map[string]*dynamic.Router),
			Middlewares: make(map[string]*dynamic.Middleware),
		},
	}

	loc := &location{
		UpstreamVhost: &dynamic.UpstreamVHost{VHost: "svc.ns.svc.cluster.local:8084"},
		SnippetAuth: &dynamic.Snippet{
			Auth: &dynamic.Auth{
				Address:       "http://oauth.ns.svc.cluster.local:4180/oauth2/auth",
				AuthSigninURL: "https://$host/oauth2/start?rd=$escaped_request_uri",
			},
		},
	}
	rt := &dynamic.Router{EntryPoints: []string{"web"}, Service: "backend"}

	p := &Provider{}
	p.applyMiddlewares(&model{}, loc, "router", rt, conf)

	require.Contains(t, rt.Middlewares, "router-snippet")
	require.Contains(t, rt.Middlewares, "router-vhost")

	assert.Less(t,
		slices.Index(rt.Middlewares, "router-snippet"),
		slices.Index(rt.Middlewares, "router-vhost"))
}
