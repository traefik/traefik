package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/v2/pkg/config/static"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestNewRouteAppenderAggregator(t *testing.T) {
	testCases := []struct {
		desc       string
		staticConf static.Configuration
		expected   map[string]int
	}{
		{
			desc: "Secure API",
			staticConf: static.Configuration{
				Global: &static.Global{},
				API: &static.API{
					Insecure: false,
				},
				EntryPoints: static.EntryPoints{
					"traefik": {},
				},
			},
			expected: map[string]int{
				"/api/providers": http.StatusBadGateway,
			},
		},
		{
			desc: "Insecure API",
			staticConf: static.Configuration{
				Global: &static.Global{},
				API: &static.API{
					Insecure: true,
				},
				EntryPoints: static.EntryPoints{
					"traefik": {},
				},
			},
			expected: map[string]int{
				"/api/rawdata": http.StatusOK,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			router := NewRouteAppenderAggregator(ctx, test.staticConf, "traefik", nil)

			internalMuxRouter := mux.NewRouter()
			router.Append(internalMuxRouter)

			internalMuxRouter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			})

			actual := make(map[string]int)
			for calledURL := range test.expected {
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodGet, calledURL, nil)
				internalMuxRouter.ServeHTTP(recorder, request)
				actual[calledURL] = recorder.Code
			}

			assert.Equal(t, test.expected, actual)
		})
	}
}
