package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/alice"
	"github.com/containous/traefik/v2/pkg/config/static"
	"github.com/containous/traefik/v2/pkg/ping"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type ChainBuilderMock struct {
	middles map[string]alice.Constructor
}

func (c *ChainBuilderMock) BuildChain(ctx context.Context, middles []string) *alice.Chain {
	chain := alice.New()

	for _, mName := range middles {
		if constructor, ok := c.middles[mName]; ok {
			chain = chain.Append(constructor)
		}
	}

	return &chain
}

func TestNewRouteAppenderAggregator(t *testing.T) {
	t.Skip("Waiting for new api handler implementation")
	testCases := []struct {
		desc       string
		staticConf static.Configuration
		middles    map[string]alice.Constructor
		expected   map[string]int
	}{
		{
			desc: "API with auth, ping without auth",
			staticConf: static.Configuration{
				Global: &static.Global{},
				API:    &static.API{
					// EntryPoint:  "traefik",
					// Middlewares: []string{"dumb"},
				},
				Ping: &ping.Handler{
					// EntryPoint: "traefik",
				},
				EntryPoints: static.EntryPoints{
					"traefik": {},
				},
			},
			middles: map[string]alice.Constructor{
				"dumb": func(_ http.Handler) (http.Handler, error) {
					return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnauthorized)
					}), nil
				},
			},
			expected: map[string]int{
				"/wrong": http.StatusBadGateway,
				"/ping":  http.StatusOK,
				// "/.well-known/acme-challenge/token": http.StatusNotFound, // FIXME
				"/api/rawdata": http.StatusUnauthorized,
			},
		},
		{
			desc: "Wrong entrypoint name",
			staticConf: static.Configuration{
				Global: &static.Global{},
				API:    &static.API{
					// EntryPoint: "no",
				},
				EntryPoints: static.EntryPoints{
					"traefik": {},
				},
			},
			expected: map[string]int{
				"/api/providers": http.StatusBadGateway,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			chainBuilder := &ChainBuilderMock{middles: test.middles}

			ctx := context.Background()

			router := NewRouteAppenderAggregator(ctx, chainBuilder, test.staticConf, "traefik", nil)

			internalMuxRouter := mux.NewRouter()
			router.Append(internalMuxRouter)

			internalMuxRouter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			})

			actual := make(map[string]int)
			for calledURL := range test.expected {
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodGet, calledURL, nil)
				internalMuxRouter.ServeHTTP(recorder, request)
				actual[calledURL] = recorder.Code
			}

			assert.Equal(t, test.expected, actual)
		})
	}
}
