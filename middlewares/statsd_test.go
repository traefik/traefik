package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/stvp/go-udp-testing"
	"github.com/urfave/negroni"
)

func TestStatsD(t *testing.T) {
	udp.SetAddr(":18125")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second
	recorder := httptest.NewRecorder()
	InitStatsdClient(&types.Statsd{":18125", "1s"})

	n := negroni.New()
	c := NewStatsD("test")
	defer StopStatsdClient()
	metricsMiddlewareBackend := NewMetricsWrapper(c)

	n.Use(metricsMiddlewareBackend)
	r := http.NewServeMux()
	r.HandleFunc(`/ok`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	r.HandleFunc(`/not-found`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "not-found")
	})
	n.UseHandler(r)

	req1 := testhelpers.MustNewRequest(http.MethodGet, "http://localhost:3000/ok", nil)
	req2 := testhelpers.MustNewRequest(http.MethodGet, "http://localhost:3000/not-found", nil)

	expected := []string{
		// We are only validating counts, as it is nearly impossible to validate latency, since it varies every run
		"traefik.requests.total:2.000000|c\n",
	}

	udp.ShouldReceiveAll(t, expected, func() {
		n.ServeHTTP(recorder, req1)
		n.ServeHTTP(recorder, req2)
	})
}
