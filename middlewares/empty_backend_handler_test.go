package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/vulcand/oxy/roundrobin"
)

func TestEmptyBackendHandler(t *testing.T) {
	tests := []struct {
		amountServer   int
		wantStatusCode int
	}{
		{
			amountServer:   0,
			wantStatusCode: http.StatusServiceUnavailable,
		},
		{
			amountServer:   1,
			wantStatusCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(fmt.Sprintf("amount servers %d", test.amountServer), func(t *testing.T) {
			t.Parallel()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			handler := NewEmptyBackendHandler(newHealthCheckLoadBalancer(test.amountServer), nextHandler)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)

			handler.ServeHTTP(recorder, req)

			if recorder.Result().StatusCode != test.wantStatusCode {
				t.Errorf("Received status code %d, wanted %d", recorder.Result().StatusCode, test.wantStatusCode)
			}
		})
	}
}

type healthCheckLoadBalancer struct {
	servers []*url.URL
}

func newHealthCheckLoadBalancer(amountServer int) *healthCheckLoadBalancer {
	servers := make([]*url.URL, amountServer)
	for i := 0; i < amountServer; i++ {
		servers[i] = testhelpers.MustParseURL("http://localhost" + strconv.Itoa(i))
	}
	return &healthCheckLoadBalancer{servers: servers}
}

func (lb *healthCheckLoadBalancer) RemoveServer(u *url.URL) error {
	return nil
}

func (lb *healthCheckLoadBalancer) UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error {
	return nil
}

func (lb *healthCheckLoadBalancer) Servers() []*url.URL {
	return lb.servers
}
