package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/codegangsta/negroni"
	"github.com/containous/traefik/types"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestPrometheus(t *testing.T) {
	recorder := httptest.NewRecorder()

	n := negroni.New()
	metricsMiddlewareBackend := NewMetricsWrapper(NewPrometheus("test", &types.Prometheus{}))
	n.Use(metricsMiddlewareBackend)
	r := http.NewServeMux()
	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc(`/ok`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	n.UseHandler(r)

	req1, err := http.NewRequest("GET", "http://localhost:3000/ok", nil)
	if err != nil {
		t.Error(err)
	}
	req2, err := http.NewRequest("GET", "http://localhost:3000/metrics", nil)
	if err != nil {
		t.Error(err)
	}

	n.ServeHTTP(recorder, req1)
	n.ServeHTTP(recorder, req2)
	body := recorder.Body.String()
	if !strings.Contains(body, reqsName) {
		t.Errorf("body does not contain request total entry '%s'", reqsName)
	}
	if !strings.Contains(body, latencyName) {
		t.Errorf("body does not contain request duration entry '%s'", reqsName)
	}
}
