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
		desc                string
		condition           string
		requestModifier     func(req *http.Request)
		wantResponseHeaders map[string]string
	}{
		{
			desc:      "match alt chain",
			condition: "Header[`Foo`].0 == `bar`",
			requestModifier: func(req *http.Request) {
				req.Header.Add("foo", "bar")
			},
			wantResponseHeaders: map[string]string{"chain": "foo"},
		},
		{
			desc:      "no match, default chain",
			condition: "Header[`Foo`].0 == `bar`",
			requestModifier: func(req *http.Request) {
				req.Header.Add("foo", "notbar")
			},
			wantResponseHeaders: map[string]string{"chain": "default"},
		},
		{
			desc:                "no match, field not found ",
			condition:           "Header[`Bar`].0 == `foo`",
			requestModifier:     nil,
			wantResponseHeaders: map[string]string{"chain": "default"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			backend := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("chain", "default")
				rw.WriteHeader(http.StatusOK)
			})

			availableMiddleware := map[string]*runtime.MiddlewareInfo{
				"foo-middleware": {
					Middleware: &dynamic.Middleware{
						Headers: &dynamic.Headers{
							CustomResponseHeaders: map[string]string{"chain": "foo"},
						},
					},
				},
				"bar-middleware": {},
			}
			middlewareBuilder := middleware.NewBuilder(availableMiddleware, nil, nil)

			cfg := dynamic.Branching{}
			cfg.Condition = test.condition
			cfg.Chain = &dynamic.Chain{
				Middlewares: []string{
					"foo-middleware",
				},
			}

			ctx := context.Background()
			plugin, err := branching.New(ctx, backend, cfg, middlewareBuilder, "test-branching")
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
			require.NoError(t, err)

			if test.requestModifier != nil {
				test.requestModifier(req)
			}

			plugin.ServeHTTP(recorder, req)

			response := recorder.Result()
			assert.Equal(t, response.StatusCode, http.StatusOK)

			for hk, hv := range test.wantResponseHeaders {
				assert.Equal(t, response.Header.Get(hk), hv)
			}
		})
	}
}
