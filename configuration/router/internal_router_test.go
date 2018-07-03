package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/mux"
	"github.com/containous/traefik/acme"
	"github.com/containous/traefik/api"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/ping"
	acmeprovider "github.com/containous/traefik/provider/acme"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/negroni"
)

func TestNewInternalRouterAggregatorWithWebPath(t *testing.T) {
	currentConfiguration := &safe.Safe{}
	currentConfiguration.Set(types.Configurations{})

	globalConfiguration := configuration.GlobalConfiguration{
		Web: &configuration.WebCompatibility{
			Path: "/prefix",
		},
		API: &api.Handler{
			EntryPoint:            "traefik",
			CurrentConfigurations: currentConfiguration,
		},
		Ping: &ping.Handler{
			EntryPoint: "traefik",
		},
		ACME: &acme.ACME{
			HTTPChallenge: &acmeprovider.HTTPChallenge{
				EntryPoint: "traefik",
			},
		},
		EntryPoints: configuration.EntryPoints{
			"traefik": &configuration.EntryPoint{},
		},
	}

	testCases := []struct {
		desc               string
		testedURL          string
		expectedStatusCode int
	}{
		{
			desc:               "Ping without prefix",
			testedURL:          "/ping",
			expectedStatusCode: 502,
		},
		{
			desc:               "Ping with prefix",
			testedURL:          "/prefix/ping",
			expectedStatusCode: 200,
		},
		{
			desc:               "acme without prefix",
			testedURL:          "/.well-known/acme-challenge/token",
			expectedStatusCode: 404,
		},
		{
			desc:               "api without prefix",
			testedURL:          "/api",
			expectedStatusCode: 502,
		},
		{
			desc:               "api with prefix",
			testedURL:          "/prefix/api",
			expectedStatusCode: 200,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			router := NewInternalRouterAggregator(globalConfiguration, "traefik")

			internalMuxRouter := mux.NewRouter()
			router.AddRoutes(internalMuxRouter)
			internalMuxRouter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			})

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.testedURL, nil)
			internalMuxRouter.ServeHTTP(recorder, request)

			assert.Equal(t, test.expectedStatusCode, recorder.Code)
		})
	}
}

func TestNewInternalRouterAggregatorWithAuth(t *testing.T) {
	currentConfiguration := &safe.Safe{}
	currentConfiguration.Set(types.Configurations{})

	globalConfiguration := configuration.GlobalConfiguration{
		API: &api.Handler{
			EntryPoint:            "traefik",
			CurrentConfigurations: currentConfiguration,
		},
		Ping: &ping.Handler{
			EntryPoint: "traefik",
		},
		ACME: &acme.ACME{
			HTTPChallenge: &acmeprovider.HTTPChallenge{
				EntryPoint: "traefik",
			},
		},
		EntryPoints: configuration.EntryPoints{
			"traefik": &configuration.EntryPoint{
				Auth: &types.Auth{
					Basic: &types.Basic{
						Users: types.Users{"test:test"},
					},
				},
			},
		},
	}

	testCases := []struct {
		desc               string
		testedURL          string
		expectedStatusCode int
	}{
		{
			desc:               "Wrong url",
			testedURL:          "/wrong",
			expectedStatusCode: 502,
		},
		{
			desc:               "Ping without auth",
			testedURL:          "/ping",
			expectedStatusCode: 200,
		},
		{
			desc:               "acme without auth",
			testedURL:          "/.well-known/acme-challenge/token",
			expectedStatusCode: 404,
		},
		{
			desc:               "api with auth",
			testedURL:          "/api",
			expectedStatusCode: 401,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			router := NewInternalRouterAggregator(globalConfiguration, "traefik")

			internalMuxRouter := mux.NewRouter()
			router.AddRoutes(internalMuxRouter)
			internalMuxRouter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			})

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.testedURL, nil)
			internalMuxRouter.ServeHTTP(recorder, request)

			assert.Equal(t, test.expectedStatusCode, recorder.Code)
		})
	}
}

func TestNewInternalRouterAggregatorWithAuthAndPrefix(t *testing.T) {
	currentConfiguration := &safe.Safe{}
	currentConfiguration.Set(types.Configurations{})

	globalConfiguration := configuration.GlobalConfiguration{
		Web: &configuration.WebCompatibility{
			Path: "/prefix",
		},
		API: &api.Handler{
			EntryPoint:            "traefik",
			CurrentConfigurations: currentConfiguration,
		},
		Ping: &ping.Handler{
			EntryPoint: "traefik",
		},
		ACME: &acme.ACME{
			HTTPChallenge: &acmeprovider.HTTPChallenge{
				EntryPoint: "traefik",
			},
		},
		EntryPoints: configuration.EntryPoints{
			"traefik": &configuration.EntryPoint{
				Auth: &types.Auth{
					Basic: &types.Basic{
						Users: types.Users{"test:test"},
					},
				},
			},
		},
	}

	testCases := []struct {
		desc               string
		testedURL          string
		expectedStatusCode int
	}{
		{
			desc:               "Ping without prefix",
			testedURL:          "/ping",
			expectedStatusCode: 502,
		},
		{
			desc:               "Ping without auth and with prefix",
			testedURL:          "/prefix/ping",
			expectedStatusCode: 200,
		},
		{
			desc:               "acme without auth and without prefix",
			testedURL:          "/.well-known/acme-challenge/token",
			expectedStatusCode: 404,
		},
		{
			desc:               "api with auth and prefix",
			testedURL:          "/prefix/api",
			expectedStatusCode: 401,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			router := NewInternalRouterAggregator(globalConfiguration, "traefik")

			internalMuxRouter := mux.NewRouter()
			router.AddRoutes(internalMuxRouter)
			internalMuxRouter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			})

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.testedURL, nil)
			internalMuxRouter.ServeHTTP(recorder, request)

			assert.Equal(t, test.expectedStatusCode, recorder.Code)
		})
	}
}

type MockInternalRouterFunc func(systemRouter *mux.Router)

func (m MockInternalRouterFunc) AddRoutes(systemRouter *mux.Router) {
	m(systemRouter)
}

func TestWithMiddleware(t *testing.T) {
	router := WithMiddleware{
		router: MockInternalRouterFunc(func(systemRouter *mux.Router) {
			systemRouter.Handle("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("router"))
			}))
		}),
		routerMiddlewares: []negroni.Handler{
			negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
				rw.Write([]byte("before middleware1|"))
				next.ServeHTTP(rw, r)
				rw.Write([]byte("|after middleware1"))

			}),
			negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
				rw.Write([]byte("before middleware2|"))
				next.ServeHTTP(rw, r)
				rw.Write([]byte("|after middleware2"))
			}),
		},
	}

	internalMuxRouter := mux.NewRouter()
	router.AddRoutes(internalMuxRouter)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	internalMuxRouter.ServeHTTP(recorder, request)

	obtained := recorder.Body.String()

	assert.Equal(t, "before middleware1|before middleware2|router|after middleware2|after middleware1", obtained)
}

func TestWithPrefix(t *testing.T) {
	testCases := []struct {
		desc               string
		prefix             string
		testedURL          string
		expectedStatusCode int
	}{
		{
			desc:               "No prefix",
			testedURL:          "/test",
			expectedStatusCode: 200,
		},
		{
			desc:               "With prefix and wrong url",
			prefix:             "/prefix",
			testedURL:          "/test",
			expectedStatusCode: 404,
		},
		{
			desc:               "With prefix",
			prefix:             "/prefix",
			testedURL:          "/prefix/test",
			expectedStatusCode: 200,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			router := WithPrefix{
				Router: MockInternalRouterFunc(func(systemRouter *mux.Router) {
					systemRouter.Handle("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}))
				}),

				PathPrefix: test.prefix,
			}
			internalMuxRouter := mux.NewRouter()
			router.AddRoutes(internalMuxRouter)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.testedURL, nil)
			internalMuxRouter.ServeHTTP(recorder, request)

			assert.Equal(t, test.expectedStatusCode, recorder.Code)
		})
	}
}
