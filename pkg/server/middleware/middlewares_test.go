package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/server/provider"
)

func TestBuilder_BuildChainNilConfig(t *testing.T) {
	testConfig := map[string]*runtime.MiddlewareInfo{
		"empty": {},
	}
	middlewaresBuilder := NewBuilder(testConfig, nil, nil)

	chain := middlewaresBuilder.BuildChain(context.Background(), []string{"empty"})
	_, err := chain.Then(nil)
	require.Error(t, err)
}

func TestBuilder_BuildChainNonExistentChain(t *testing.T) {
	testConfig := map[string]*runtime.MiddlewareInfo{
		"foobar": {},
	}
	middlewaresBuilder := NewBuilder(testConfig, nil, nil)

	chain := middlewaresBuilder.BuildChain(context.Background(), []string{"empty"})
	_, err := chain.Then(nil)
	require.Error(t, err)
}

func TestBuilder_BuildChainWithContext(t *testing.T) {
	testCases := []struct {
		desc            string
		buildChain      []string
		configuration   map[string]*dynamic.Middleware
		expected        map[string]string
		contextProvider string
		expectedError   error
	}{
		{
			desc:       "Simple middleware",
			buildChain: []string{"middleware-1"},
			configuration: map[string]*dynamic.Middleware{
				"middleware-1": {
					Headers: &dynamic.Headers{
						CustomRequestHeaders: map[string]string{"middleware-1": "value-middleware-1"},
					},
				},
			},
			expected: map[string]string{"middleware-1": "value-middleware-1"},
		},
		{
			desc:       "Middleware that references a chain",
			buildChain: []string{"middleware-chain-1"},
			configuration: map[string]*dynamic.Middleware{
				"middleware-1": {
					Headers: &dynamic.Headers{
						CustomRequestHeaders: map[string]string{"middleware-1": "value-middleware-1"},
					},
				},
				"middleware-chain-1": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"middleware-1"},
					},
				},
			},
			expected: map[string]string{"middleware-1": "value-middleware-1"},
		},
		{
			desc:       "Should suffix the middlewareName with the provider in the context",
			buildChain: []string{"middleware-1"},
			configuration: map[string]*dynamic.Middleware{
				"middleware-1@provider-1": {
					Headers: &dynamic.Headers{
						CustomRequestHeaders: map[string]string{"middleware-1@provider-1": "value-middleware-1"},
					},
				},
			},
			expected:        map[string]string{"middleware-1@provider-1": "value-middleware-1"},
			contextProvider: "provider-1",
		},
		{
			desc:       "Should not suffix a qualified middlewareName with the provider in the context",
			buildChain: []string{"middleware-1@provider-1"},
			configuration: map[string]*dynamic.Middleware{
				"middleware-1@provider-1": {
					Headers: &dynamic.Headers{
						CustomRequestHeaders: map[string]string{"middleware-1@provider-1": "value-middleware-1"},
					},
				},
			},
			expected:        map[string]string{"middleware-1@provider-1": "value-middleware-1"},
			contextProvider: "provider-1",
		},
		{
			desc:       "Should be context aware if a chain references another middleware",
			buildChain: []string{"middleware-chain-1@provider-1"},
			configuration: map[string]*dynamic.Middleware{
				"middleware-1@provider-1": {
					Headers: &dynamic.Headers{
						CustomRequestHeaders: map[string]string{"middleware-1": "value-middleware-1"},
					},
				},
				"middleware-chain-1@provider-1": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"middleware-1"},
					},
				},
			},
			expected: map[string]string{"middleware-1": "value-middleware-1"},
		},
		{
			desc:       "Should handle nested chains with different context",
			buildChain: []string{"middleware-chain-1@provider-1", "middleware-chain-1"},
			configuration: map[string]*dynamic.Middleware{
				"middleware-1@provider-1": {
					Headers: &dynamic.Headers{
						CustomRequestHeaders: map[string]string{"middleware-1": "value-middleware-1"},
					},
				},
				"middleware-2@provider-1": {
					Headers: &dynamic.Headers{
						CustomRequestHeaders: map[string]string{"middleware-2": "value-middleware-2"},
					},
				},
				"middleware-chain-1@provider-1": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"middleware-1"},
					},
				},
				"middleware-chain-2@provider-1": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"middleware-2"},
					},
				},
				"middleware-chain-1@provider-2": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"middleware-2@provider-1", "middleware-chain-2@provider-1"},
					},
				},
			},
			expected:        map[string]string{"middleware-1": "value-middleware-1", "middleware-2": "value-middleware-2"},
			contextProvider: "provider-2",
		},
		{
			desc:       "Detects recursion in Middleware chain",
			buildChain: []string{"m1"},
			configuration: map[string]*dynamic.Middleware{
				"ok": {
					Retry: &dynamic.Retry{},
				},
				"m1": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"m2"},
					},
				},
				"m2": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"ok", "m3"},
					},
				},
				"m3": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"m1"},
					},
				},
			},
			expectedError: errors.New("could not instantiate middleware m1: recursion detected in m1->m2->m3->m1"),
		},
		{
			desc:       "Detects recursion in Middleware chain",
			buildChain: []string{"m1@provider"},
			configuration: map[string]*dynamic.Middleware{
				"ok@provider2": {
					Retry: &dynamic.Retry{},
				},
				"m1@provider": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"m2@provider2"},
					},
				},
				"m2@provider2": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"ok", "m3@provider"},
					},
				},
				"m3@provider": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"m1"},
					},
				},
			},
			expectedError: errors.New("could not instantiate middleware m1@provider: recursion detected in m1@provider->m2@provider2->m3@provider->m1@provider"),
		},
		{
			buildChain: []string{"ok", "m0"},
			configuration: map[string]*dynamic.Middleware{
				"ok": {
					Retry: &dynamic.Retry{},
				},
				"m0": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"m0"},
					},
				},
			},
			expectedError: errors.New("could not instantiate middleware m0: recursion detected in m0->m0"),
		},
		{
			desc:       "Detects MiddlewareChain that references a Chain that references a Chain with a missing middleware",
			buildChain: []string{"m0"},
			configuration: map[string]*dynamic.Middleware{
				"m0": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"m1"},
					},
				},
				"m1": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"m2"},
					},
				},
				"m2": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"m3"},
					},
				},
				"m3": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"m2"},
					},
				},
			},
			expectedError: errors.New("could not instantiate middleware m2: recursion detected in m0->m1->m2->m3->m2"),
		},
		{
			desc:       "--",
			buildChain: []string{"m0"},
			configuration: map[string]*dynamic.Middleware{
				"m0": {
					Chain: &dynamic.Chain{
						Middlewares: []string{"m0"},
					},
				},
			},
			expectedError: errors.New("could not instantiate middleware m0: recursion detected in m0->m0"),
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if len(test.contextProvider) > 0 {
				ctx = provider.AddInContext(ctx, "foobar@"+test.contextProvider)
			}

			rtConf := runtime.NewConfig(dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: test.configuration,
				},
			})
			builder := NewBuilder(rtConf.Middlewares, nil, nil)

			result := builder.BuildChain(ctx, test.buildChain)

			handlers, err := result.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
			if test.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				recorder := httptest.NewRecorder()
				request, _ := http.NewRequest(http.MethodGet, "http://foo/", nil)
				handlers.ServeHTTP(recorder, request)

				for key, value := range test.expected {
					assert.Equal(t, value, request.Header.Get(key))
				}
			}
		})
	}
}

func TestBuilder_buildConstructor(t *testing.T) {
	testConfig := map[string]*dynamic.Middleware{
		"cb-empty": {
			CircuitBreaker: &dynamic.CircuitBreaker{
				Expression: "",
			},
		},
		"cb-foo": {
			CircuitBreaker: &dynamic.CircuitBreaker{
				Expression: "NetworkErrorRatio() > 0.5",
			},
		},
		"ap-empty": {
			AddPrefix: &dynamic.AddPrefix{
				Prefix: "",
			},
		},
		"ap-foo": {
			AddPrefix: &dynamic.AddPrefix{
				Prefix: "foo/",
			},
		},
		"buff-foo": {
			Buffering: &dynamic.Buffering{
				MaxRequestBodyBytes:  1,
				MemRequestBodyBytes:  2,
				MaxResponseBodyBytes: 3,
				MemResponseBodyBytes: 5,
			},
		},
	}

	rtConf := runtime.NewConfig(dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Middlewares: testConfig,
		},
	})
	middlewaresBuilder := NewBuilder(rtConf.Middlewares, nil, nil)

	testCases := []struct {
		desc          string
		middlewareID  string
		expectedError bool
	}{
		{
			desc:          "Should fail at creating a circuit breaker with an empty expression",
			middlewareID:  "cb-empty",
			expectedError: true,
		},
		{
			desc:          "Should create a circuit breaker with a valid expression",
			middlewareID:  "cb-foo",
			expectedError: false,
		},
		{
			desc:          "Should create a buffering middleware",
			middlewareID:  "buff-foo",
			expectedError: false,
		},
		{
			desc:          "Should not create an empty AddPrefix middleware when given an empty prefix",
			middlewareID:  "ap-empty",
			expectedError: true,
		},
		{
			desc:          "Should create an AddPrefix middleware when given a valid configuration",
			middlewareID:  "ap-foo",
			expectedError: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			constructor, err := middlewaresBuilder.buildConstructor(context.Background(), test.middlewareID)
			require.NoError(t, err)

			middleware, err2 := constructor(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

			if test.expectedError {
				require.Error(t, err2)
			} else {
				require.NoError(t, err)
				require.NotNil(t, middleware)
			}
		})
	}
}
