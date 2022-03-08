package branching_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/middlewares/branching"
	"github.com/traefik/traefik/v2/pkg/server/middleware"
)

func TestBuilder(t *testing.T) {
	testCases := []struct {
		desc      string
		condition string
		wantErr   bool
	}{
		{
			desc:      "invalid condition",
			condition: "HumHum",
			wantErr:   true,
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			cfg := dynamic.Branching{}
			cfg.Condition = test.condition

			availableMiddleware := map[string]*runtime.MiddlewareInfo{
				"empty": {},
			}
			middlewareBuilder := middleware.NewBuilder(availableMiddleware, nil, nil)

			_, err := branching.New(
				context.Background(),
				http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}),
				cfg,
				middlewareBuilder,
				"test-plugin")

			if err != nil && !test.wantErr {
				t.Fatal("unexpected error", err)
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
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
			if err != nil {
				t.Fatal(err)
			}

			if test.requestModifier != nil {
				test.requestModifier(req)
			}

			plugin.ServeHTTP(recorder, req)

			response := recorder.Result()
			if response.StatusCode != http.StatusOK {
				t.Fatal("failed response")
			}

			for hk, hv := range test.wantResponseHeaders {
				if response.Header.Get(hk) != hv {
					t.Fatal("unexpected chain header")
				}
			}
		})
	}
}
