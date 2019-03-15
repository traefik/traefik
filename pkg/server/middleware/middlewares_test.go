package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/server/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_buildConstructorCircuitBreaker(t *testing.T) {
	testConfig := map[string]*config.Middleware{
		"empty": {
			CircuitBreaker: &config.CircuitBreaker{
				Expression: "",
			},
		},
		"foo": {
			CircuitBreaker: &config.CircuitBreaker{
				Expression: "NetworkErrorRatio() > 0.5",
			},
		},
	}
	middlewaresBuilder := NewBuilder(testConfig, nil)

	emptyHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	testCases := []struct {
		desc          string
		middlewareID  string
		expectedError bool
	}{
		{
			desc:          "Should fail at creating a circuit breaker with an empty expression",
			expectedError: true,
			middlewareID:  "empty",
		}, {
			desc:          "Should create a circuit breaker with a valid expression",
			expectedError: false,
			middlewareID:  "foo",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			constructor, err := middlewaresBuilder.buildConstructor(context.Background(), test.middlewareID, *testConfig[test.middlewareID])
			require.NoError(t, err)

			middleware, err2 := constructor(emptyHandler)

			if test.expectedError {
				require.Error(t, err2)
			} else {
				require.NoError(t, err)
				require.NotNil(t, middleware)
			}
		})
	}
}

func TestBuilder_BuildChainNilConfig(t *testing.T) {
	testConfig := map[string]*config.Middleware{
		"empty": {},
	}
	middlewaresBuilder := NewBuilder(testConfig, nil)

	chain := middlewaresBuilder.BuildChain(context.Background(), []string{"empty"})
	_, err := chain.Then(nil)
	require.Error(t, err)
}

func TestBuilder_BuildChainNonExistentChain(t *testing.T) {
	testConfig := map[string]*config.Middleware{
		"foobar": {},
	}
	middlewaresBuilder := NewBuilder(testConfig, nil)

	chain := middlewaresBuilder.BuildChain(context.Background(), []string{"empty"})
	_, err := chain.Then(nil)
	require.Error(t, err)
}

func TestBuilder_buildConstructorAddPrefix(t *testing.T) {
	testConfig := map[string]*config.Middleware{
		"empty": {
			AddPrefix: &config.AddPrefix{
				Prefix: "",
			},
		},
		"foo": {
			AddPrefix: &config.AddPrefix{
				Prefix: "foo/",
			},
		},
	}

	middlewaresBuilder := NewBuilder(testConfig, nil)

	testCases := []struct {
		desc          string
		middlewareID  string
		expectedError bool
	}{
		{
			desc:          "Should not create an empty AddPrefix middleware when given an empty prefix",
			middlewareID:  "empty",
			expectedError: true,
		}, {
			desc:          "Should create an AddPrefix middleware when given a valid configuration",
			middlewareID:  "foo",
			expectedError: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			constructor, err := middlewaresBuilder.buildConstructor(context.Background(), test.middlewareID, *testConfig[test.middlewareID])
			require.NoError(t, err)

			middleware, err2 := constructor(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))

			if test.expectedError {
				require.Error(t, err2)
			} else {
				require.NoError(t, err)
				require.NotNil(t, middleware)
			}
		})
	}
}

func TestBuild_BuildChainWithContext(t *testing.T) {
	testCases := []struct {
		desc            string
		buildChain      []string
		configuration   map[string]*config.Middleware
		expected        map[string]string
		contextProvider string
		expectedError   error
	}{
		{
			desc:       "Simple middleware",
			buildChain: []string{"middleware-1"},
			configuration: map[string]*config.Middleware{
				"middleware-1": {
					Headers: &config.Headers{
						CustomRequestHeaders: map[string]string{"middleware-1": "value-middleware-1"},
					},
				},
			},
			expected: map[string]string{"middleware-1": "value-middleware-1"},
		},
		{
			desc:       "Middleware that references a chain",
			buildChain: []string{"middleware-chain-1"},
			configuration: map[string]*config.Middleware{
				"middleware-1": {
					Headers: &config.Headers{
						CustomRequestHeaders: map[string]string{"middleware-1": "value-middleware-1"},
					},
				},
				"middleware-chain-1": {
					Chain: &config.Chain{
						Middlewares: []string{"middleware-1"},
					},
				},
			},
			expected: map[string]string{"middleware-1": "value-middleware-1"},
		},
		{
			desc:       "Should prefix the middlewareName with the provider in the context",
			buildChain: []string{"middleware-1"},
			configuration: map[string]*config.Middleware{
				"provider-1.middleware-1": {
					Headers: &config.Headers{
						CustomRequestHeaders: map[string]string{"provider-1.middleware-1": "value-middleware-1"},
					},
				},
			},
			expected:        map[string]string{"provider-1.middleware-1": "value-middleware-1"},
			contextProvider: "provider-1",
		},
		{
			desc:       "Should not prefix a qualified middlewareName with the provider in the context",
			buildChain: []string{"provider-1.middleware-1"},
			configuration: map[string]*config.Middleware{
				"provider-1.middleware-1": {
					Headers: &config.Headers{
						CustomRequestHeaders: map[string]string{"provider-1.middleware-1": "value-middleware-1"},
					},
				},
			},
			expected:        map[string]string{"provider-1.middleware-1": "value-middleware-1"},
			contextProvider: "provider-1",
		},
		{
			desc:       "Should be context aware if a chain references another middleware",
			buildChain: []string{"provider-1.middleware-chain-1"},
			configuration: map[string]*config.Middleware{
				"provider-1.middleware-1": {
					Headers: &config.Headers{
						CustomRequestHeaders: map[string]string{"middleware-1": "value-middleware-1"},
					},
				},
				"provider-1.middleware-chain-1": {
					Chain: &config.Chain{
						Middlewares: []string{"middleware-1"},
					},
				},
			},
			expected: map[string]string{"middleware-1": "value-middleware-1"},
		},
		{
			desc:       "Should handle nested chains with different context",
			buildChain: []string{"provider-1.middleware-chain-1", "middleware-chain-1"},
			configuration: map[string]*config.Middleware{
				"provider-1.middleware-1": {
					Headers: &config.Headers{
						CustomRequestHeaders: map[string]string{"middleware-1": "value-middleware-1"},
					},
				},
				"provider-1.middleware-2": {
					Headers: &config.Headers{
						CustomRequestHeaders: map[string]string{"middleware-2": "value-middleware-2"},
					},
				},
				"provider-1.middleware-chain-1": {
					Chain: &config.Chain{
						Middlewares: []string{"middleware-1"},
					},
				},
				"provider-1.middleware-chain-2": {
					Chain: &config.Chain{
						Middlewares: []string{"middleware-2"},
					},
				},
				"provider-2.middleware-chain-1": {
					Chain: &config.Chain{
						Middlewares: []string{"provider-1.middleware-2", "provider-1.middleware-chain-2"},
					},
				},
			},
			expected:        map[string]string{"middleware-1": "value-middleware-1", "middleware-2": "value-middleware-2"},
			contextProvider: "provider-2",
		},
		{
			desc:       "Detects recursion in Middleware chain",
			buildChain: []string{"m1"},
			configuration: map[string]*config.Middleware{
				"ok": {
					Retry: &config.Retry{},
				},
				"m1": {
					Chain: &config.Chain{
						Middlewares: []string{"m2"},
					},
				},
				"m2": {
					Chain: &config.Chain{
						Middlewares: []string{"ok", "m3"},
					},
				},
				"m3": {
					Chain: &config.Chain{
						Middlewares: []string{"m1"},
					},
				},
			},
			expectedError: errors.New("could not instantiate middleware m1: recursion detected in m1->m2->m3->m1"),
		},
		{
			desc:       "Detects recursion in Middleware chain",
			buildChain: []string{"provider.m1"},
			configuration: map[string]*config.Middleware{
				"provider2.ok": {
					Retry: &config.Retry{},
				},
				"provider.m1": {
					Chain: &config.Chain{
						Middlewares: []string{"provider2.m2"},
					},
				},
				"provider2.m2": {
					Chain: &config.Chain{
						Middlewares: []string{"ok", "provider.m3"},
					},
				},
				"provider.m3": {
					Chain: &config.Chain{
						Middlewares: []string{"m1"},
					},
				},
			},
			expectedError: errors.New("could not instantiate middleware provider.m1: recursion detected in provider.m1->provider2.m2->provider.m3->provider.m1"),
		},
		{
			buildChain: []string{"ok", "m0"},
			configuration: map[string]*config.Middleware{
				"ok": {
					Retry: &config.Retry{},
				},
				"m0": {
					Chain: &config.Chain{
						Middlewares: []string{"m0"},
					},
				},
			},
			expectedError: errors.New("could not instantiate middleware m0: recursion detected in m0->m0"),
		},
		{
			desc:       "Detects MiddlewareChain that references a Chain that references a Chain with a missing middleware",
			buildChain: []string{"m0"},
			configuration: map[string]*config.Middleware{
				"m0": {
					Chain: &config.Chain{
						Middlewares: []string{"m1"},
					},
				},
				"m1": {
					Chain: &config.Chain{
						Middlewares: []string{"m2"},
					},
				},
				"m2": {
					Chain: &config.Chain{
						Middlewares: []string{"m3"},
					},
				},
				"m3": {
					Chain: &config.Chain{
						Middlewares: []string{"m2"},
					},
				},
			},
			expectedError: errors.New("could not instantiate middleware m2: recursion detected in m0->m1->m2->m3->m2"),
		},
		{
			desc:       "--",
			buildChain: []string{"m0"},
			configuration: map[string]*config.Middleware{
				"m0": {
					Chain: &config.Chain{
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
				ctx = internal.AddProviderInContext(ctx, test.contextProvider+".foobar")
			}

			builder := NewBuilder(test.configuration, nil)

			result := builder.BuildChain(ctx, test.buildChain)

			handlers, err := result.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
			if test.expectedError != nil {
				require.NotNil(t, err)
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
