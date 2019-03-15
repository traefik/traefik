package responsemodifiers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/middlewares/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stubResponse(_ map[string]*config.Middleware) *http.Response {
	return &http.Response{Header: make(http.Header)}
}

func TestBuilderBuild(t *testing.T) {
	testCases := []struct {
		desc        string
		middlewares []string
		// buildResponse is needed because secure use a private context key
		buildResponse  func(map[string]*config.Middleware) *http.Response
		conf           map[string]*config.Middleware
		assertResponse func(*testing.T, *http.Response)
	}{
		{
			desc:           "no configuration",
			middlewares:    []string{"foo", "bar"},
			buildResponse:  stubResponse,
			conf:           map[string]*config.Middleware{},
			assertResponse: func(t *testing.T, resp *http.Response) {},
		},
		{
			desc:          "one modifier",
			middlewares:   []string{"foo", "bar"},
			buildResponse: stubResponse,
			conf: map[string]*config.Middleware{
				"foo": {
					Headers: &config.Headers{
						CustomResponseHeaders: map[string]string{"X-Foo": "foo"},
					},
				},
			},
			assertResponse: func(t *testing.T, resp *http.Response) {
				t.Helper()

				assert.Equal(t, resp.Header.Get("X-Foo"), "foo")
			},
		},
		{
			desc:        "secure: one modifier",
			middlewares: []string{"foo", "bar"},
			buildResponse: func(middlewares map[string]*config.Middleware) *http.Response {
				ctx := context.Background()

				var request *http.Request
				next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					request = req
				})

				headerM := *middlewares["foo"].Headers
				handler, err := headers.New(ctx, next, headerM, "secure")
				require.NoError(t, err)

				handler.ServeHTTP(httptest.NewRecorder(),
					httptest.NewRequest(http.MethodGet, "http://foo.com", nil))

				return &http.Response{Header: make(http.Header), Request: request}
			},
			conf: map[string]*config.Middleware{
				"foo": {
					Headers: &config.Headers{
						ReferrerPolicy: "no-referrer",
					},
				},
				"bar": {
					Headers: &config.Headers{
						CustomResponseHeaders: map[string]string{"X-Bar": "bar"},
					},
				},
			},
			assertResponse: func(t *testing.T, resp *http.Response) {
				t.Helper()

				assert.Equal(t, resp.Header.Get("Referrer-Policy"), "no-referrer")
			},
		},
		{
			desc:          "two modifiers",
			middlewares:   []string{"foo", "bar"},
			buildResponse: stubResponse,
			conf: map[string]*config.Middleware{
				"foo": {
					Headers: &config.Headers{
						CustomResponseHeaders: map[string]string{"X-Foo": "foo"},
					},
				},
				"bar": {
					Headers: &config.Headers{
						CustomResponseHeaders: map[string]string{"X-Bar": "bar"},
					},
				},
			},
			assertResponse: func(t *testing.T, resp *http.Response) {
				t.Helper()

				assert.Equal(t, resp.Header.Get("X-Foo"), "foo")
				assert.Equal(t, resp.Header.Get("X-Bar"), "bar")
			},
		},
		{
			desc:          "modifier order",
			middlewares:   []string{"foo", "bar"},
			buildResponse: stubResponse,
			conf: map[string]*config.Middleware{
				"foo": {
					Headers: &config.Headers{
						CustomResponseHeaders: map[string]string{"X-Foo": "foo"},
					},
				},
				"bar": {
					Headers: &config.Headers{
						CustomResponseHeaders: map[string]string{"X-Foo": "bar"},
					},
				},
			},
			assertResponse: func(t *testing.T, resp *http.Response) {
				t.Helper()

				assert.Equal(t, resp.Header.Get("X-Foo"), "foo")
			},
		},
		{
			desc:          "chain",
			middlewares:   []string{"chain"},
			buildResponse: stubResponse,
			conf: map[string]*config.Middleware{
				"foo": {
					Headers: &config.Headers{
						CustomResponseHeaders: map[string]string{"X-Foo": "foo"},
					},
				},
				"bar": {
					Headers: &config.Headers{
						CustomResponseHeaders: map[string]string{"X-Foo": "bar"},
					},
				},
				"chain": {
					Chain: &config.Chain{
						Middlewares: []string{"foo", "bar"},
					},
				},
			},
			assertResponse: func(t *testing.T, resp *http.Response) {
				t.Helper()

				assert.Equal(t, resp.Header.Get("X-Foo"), "foo")
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			builder := NewBuilder(test.conf)

			rm := builder.Build(context.Background(), test.middlewares)

			resp := test.buildResponse(test.conf)

			err := rm(resp)
			require.NoError(t, err)

			test.assertResponse(t, resp)
		})
	}
}
