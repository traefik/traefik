package responsemodifiers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/containous/traefik/v2/pkg/middlewares/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stubResponse(_ map[string]*dynamic.Middleware) *http.Response {
	return &http.Response{Header: make(http.Header)}
}

func TestBuilderBuild(t *testing.T) {
	testCases := []struct {
		desc        string
		middlewares []string
		// buildResponse is needed because secure use a private context key
		buildResponse  func(map[string]*dynamic.Middleware) *http.Response
		conf           map[string]*dynamic.Middleware
		assertResponse func(*testing.T, *http.Response)
	}{
		{
			desc:           "no configuration",
			middlewares:    []string{"foo", "bar"},
			buildResponse:  stubResponse,
			conf:           map[string]*dynamic.Middleware{},
			assertResponse: func(t *testing.T, resp *http.Response) {},
		},
		{
			desc:          "one modifier",
			middlewares:   []string{"foo", "bar"},
			buildResponse: stubResponse,
			conf: map[string]*dynamic.Middleware{
				"foo": {
					Headers: &dynamic.Headers{
						CustomResponseHeaders: map[string]string{"X-Foo": "foo"},
					},
				},
			},
			assertResponse: func(t *testing.T, resp *http.Response) {
				t.Helper()

				assert.Equal(t, "foo", resp.Header.Get("X-Foo"))
			},
		},
		{
			desc:        "secure: one modifier",
			middlewares: []string{"foo", "bar"},
			buildResponse: func(middlewares map[string]*dynamic.Middleware) *http.Response {
				ctx := context.Background()

				var request *http.Request
				next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					request = req
				})

				headerM := *middlewares["foo"].Headers
				handler, err := headers.New(ctx, next, headerM, "foo")
				require.NoError(t, err)

				handler.ServeHTTP(httptest.NewRecorder(),
					httptest.NewRequest(http.MethodGet, "http://foo.com", nil))

				return &http.Response{Header: make(http.Header), Request: request}
			},
			conf: map[string]*dynamic.Middleware{
				"foo": {
					Headers: &dynamic.Headers{
						ReferrerPolicy: "no-referrer",
					},
				},
				"bar": {
					Headers: &dynamic.Headers{
						CustomResponseHeaders: map[string]string{"X-Bar": "bar"},
					},
				},
			},
			assertResponse: func(t *testing.T, resp *http.Response) {
				t.Helper()

				assert.Equal(t, "no-referrer", resp.Header.Get("Referrer-Policy"))
			},
		},
		{
			desc:          "two modifiers",
			middlewares:   []string{"foo", "bar"},
			buildResponse: stubResponse,
			conf: map[string]*dynamic.Middleware{
				"foo": {
					Headers: &dynamic.Headers{
						CustomResponseHeaders: map[string]string{"X-Foo": "foo"},
					},
				},
				"bar": {
					Headers: &dynamic.Headers{
						CustomResponseHeaders: map[string]string{"X-Bar": "bar"},
					},
				},
			},
			assertResponse: func(t *testing.T, resp *http.Response) {
				t.Helper()

				assert.Equal(t, "foo", resp.Header.Get("X-Foo"))
				assert.Equal(t, "bar", resp.Header.Get("X-Bar"))
			},
		},
		{
			desc:          "modifier order",
			middlewares:   []string{"foo", "bar"},
			buildResponse: stubResponse,
			conf: map[string]*dynamic.Middleware{
				"foo": {
					Headers: &dynamic.Headers{
						CustomResponseHeaders: map[string]string{"X-Foo": "foo"},
					},
				},
				"bar": {
					Headers: &dynamic.Headers{
						CustomResponseHeaders: map[string]string{"X-Foo": "bar"},
					},
				},
			},
			assertResponse: func(t *testing.T, resp *http.Response) {
				t.Helper()

				assert.Equal(t, "foo", resp.Header.Get("X-Foo"))
			},
		},
		{
			desc:          "chain",
			middlewares:   []string{"chain"},
			buildResponse: stubResponse,
			conf: map[string]*dynamic.Middleware{
				"foo": {
					Headers: &dynamic.Headers{
						CustomResponseHeaders: map[string]string{"X-Foo": "foo"},
					},
				},
				"bar": {
					Headers: &dynamic.Headers{
						CustomResponseHeaders: map[string]string{"X-Foo": "bar"},
					},
				},
				"chain": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"foo", "bar"},
					},
				},
			},
			assertResponse: func(t *testing.T, resp *http.Response) {
				t.Helper()

				assert.Equal(t, "foo", resp.Header.Get("X-Foo"))
			},
		},
		{
			desc:          "nil middleware",
			middlewares:   []string{"foo"},
			buildResponse: stubResponse,
			conf: map[string]*dynamic.Middleware{
				"foo": nil,
			},
			assertResponse: func(t *testing.T, resp *http.Response) {},
		},

		{
			desc:          "chain without headers",
			middlewares:   []string{"chain"},
			buildResponse: stubResponse,
			conf: map[string]*dynamic.Middleware{
				"foo": {IPWhiteList: &dynamic.IPWhiteList{}},
				"chain": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"foo"},
					},
				},
			},
			assertResponse: func(t *testing.T, resp *http.Response) {},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rtConf := runtime.NewConfig(dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: test.conf,
				},
			})
			builder := NewBuilder(rtConf.Middlewares)

			rm := builder.Build(context.Background(), test.middlewares)
			if rm == nil {
				return
			}

			resp := test.buildResponse(test.conf)

			err := rm(resp)
			require.NoError(t, err)

			test.assertResponse(t, resp)
		})
	}
}
