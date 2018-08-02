package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/mux"
	"github.com/containous/traefik/acme"
	"github.com/containous/traefik/api"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/ping"
	acmeprovider "github.com/containous/traefik/provider/acme"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/negroni"
)

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

type MockInternalRouterFunc func(systemRouter *mux.Router)

func (m MockInternalRouterFunc) AddRoutes(systemRouter *mux.Router) {
	m(systemRouter)
}

func TestWithMiddleware(t *testing.T) {
	router := WithMiddleware{
		router: MockInternalRouterFunc(func(systemRouter *mux.Router) {
			systemRouter.Handle("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				if _, err := w.Write([]byte("router")); err != nil {
					log.Error(err)
				}
			}))
		}),
		routerMiddlewares: []negroni.Handler{
			negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
				if _, err := rw.Write([]byte("before middleware1|")); err != nil {
					log.Error(err)
				}

				next.ServeHTTP(rw, r)

				if _, err := rw.Write([]byte("|after middleware1")); err != nil {
					log.Error(err)
				}

			}),
			negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
				if _, err := rw.Write([]byte("before middleware2|")); err != nil {
					log.Error(err)
				}

				next.ServeHTTP(rw, r)

				if _, err := rw.Write([]byte("|after middleware2")); err != nil {
					log.Error(err)
				}
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
