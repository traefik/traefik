package branching_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/middlewares/branching"
	"github.com/traefik/traefik/v2/pkg/server/middleware"
)

func TestBuilder(t *testing.T) {
	testCases := []struct {
		desc    string
		cfg     dynamic.Branching
		wantErr string
	}{
		{
			desc:    "invalid condition",
			cfg:     dynamic.Branching{Condition: "ni", Chain: &dynamic.Chain{Middlewares: []string{"foo", "bar"}}},
			wantErr: "failed to create evaluator for expression \"ni\"",
		},
		{
			desc:    "valid condition, empty chain",
			cfg:     dynamic.Branching{Condition: "ni"},
			wantErr: "empty branch chain",
		},
		{
			desc:    "valid condition, valid chain",
			cfg:     dynamic.Branching{Condition: "Host == `foo.bar`", Chain: &dynamic.Chain{Middlewares: []string{"some-middleware"}}},
			wantErr: "",
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			cfg := test.cfg

			availableMiddleware := map[string]*runtime.MiddlewareInfo{
				"some-middleware": {},
			}
			middlewareBuilder := middleware.NewBuilder(availableMiddleware, nil, nil)

			_, err := branching.New(
				context.Background(),
				http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}),
				cfg,
				middlewareBuilder,
				"test-plugin")

			if test.wantErr != "" {
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestBranching(t *testing.T) {
	testCases := []struct {
		desc            string
		condition       string
		requestModifier func(req *http.Request)
		wantedHeaders   []string
		unwantedHeaders []string
	}{
		{
			desc:      "match alt chain",
			condition: "Header[`X-Chain`].0 == `B`",
			requestModifier: func(req *http.Request) {
				req.Header.Add("X-Chain", "B")
			},
			wantedHeaders: []string{"X-Chain-A", "X-Chain-B", "X-Chain-C"},
		},
		{
			desc:      "no match, default chain",
			condition: "Header[`X-Chain`].0 == `Z`",
			requestModifier: func(req *http.Request) {
				req.Header.Add("X-Chain", "B")
			},
			wantedHeaders:   []string{"X-Chain-A", "X-Chain-C"},
			unwantedHeaders: []string{"X-Chain-B"},
		},
		{
			desc:            "no match, field not found ",
			condition:       "Header[`Z-Chain`].0 == `foo`",
			requestModifier: nil,
			wantedHeaders:   []string{"X-Chain-A", "X-Chain-C"},
			unwantedHeaders: []string{"X-Chain-B"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			backend := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			})

			availableMiddleware := map[string]*runtime.MiddlewareInfo{
				"A-middleware": {
					Middleware: &dynamic.Middleware{
						Headers: &dynamic.Headers{
							CustomResponseHeaders: map[string]string{"X-Chain-A": "A"},
						},
					},
				},
				"B-middleware": {
					Middleware: &dynamic.Middleware{
						Headers: &dynamic.Headers{
							CustomResponseHeaders: map[string]string{"X-Chain-B": "B"},
						},
					},
				},
				"C-middleware": {
					Middleware: &dynamic.Middleware{
						Headers: &dynamic.Headers{
							CustomResponseHeaders: map[string]string{"X-Chain-C": "C"},
						},
					},
				},
				"branch-middleware": {
					Middleware: &dynamic.Middleware{
						Branching: &dynamic.Branching{
							Condition: test.condition,
							Chain: &dynamic.Chain{
								Middlewares: []string{
									"B-middleware",
								},
							},
						},
					},
				},
				"bar-middleware": {},
			}
			middlewareBuilder := middleware.NewBuilder(availableMiddleware, nil, nil)

			ctx := context.Background()
			chain := middlewareBuilder.BuildChain(ctx, []string{"A-middleware", "branch-middleware", "C-middleware"})
			handler, err := chain.Then(backend)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
			require.NoError(t, err)

			if test.requestModifier != nil {
				test.requestModifier(req)
			}

			handler.ServeHTTP(recorder, req)

			response := recorder.Result()
			assert.Equal(t, response.StatusCode, http.StatusOK)

			for _, hk := range test.wantedHeaders {
				assert.NotEmpty(t, response.Header.Get(hk))
			}
			for _, hk := range test.unwantedHeaders {
				assert.Empty(t, response.Header.Get(hk))
			}
		})
	}
}
