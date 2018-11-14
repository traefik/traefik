package middleware

import (
	"context"
	"net/http"
	"testing"

	"github.com/containous/traefik/config"
	"github.com/stretchr/testify/require"
)

func TestMiddlewaresRegistry_BuildMiddlewareCircuitBreaker(t *testing.T) {
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

func TestMiddlewaresRegistry_BuildChainNilConfig(t *testing.T) {
	testConfig := map[string]*config.Middleware{
		"empty": {},
	}
	middlewaresBuilder := NewBuilder(testConfig, nil)

	chain, err := middlewaresBuilder.BuildChain(context.Background(), []string{"empty"})
	require.NoError(t, err)

	_, err = chain.Then(nil)
	require.NoError(t, err)
}

func TestMiddlewaresRegistry_BuildMiddlewareAddPrefix(t *testing.T) {
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
			desc:          "Should not create an emty AddPrefix middleware when given an empty prefix",
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
