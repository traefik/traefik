package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/codegangsta/negroni"
	"github.com/containous/traefik/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestPrometheus(t *testing.T) {
	metricsFamily, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("could not gather metrics family: %s", err)
	}
	initialMetricsFamilyCount := len(metricsFamily)

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
		t.Errorf("body does not contain request duration entry '%s'", latencyName)
	}

	// Register the same metrics again
	metricsMiddlewareBackend = NewMetricsWrapper(NewPrometheus("test", &types.Prometheus{}))
	n = negroni.New()
	n.Use(metricsMiddlewareBackend)
	n.UseHandler(r)

	n.ServeHTTP(recorder, req2)

	metricsFamily, err = prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("could not gather metrics family: %s", err)
	}

	tests := []struct {
		name   string
		labels map[string]string
		assert func(*dto.MetricFamily)
	}{
		{
			name: reqsName,
			labels: map[string]string{
				"code":    "200",
				"method":  "GET",
				"service": "test",
			},
			assert: func(family *dto.MetricFamily) {
				cv := uint(family.Metric[0].Counter.GetValue())
				if cv != 3 {
					t.Errorf("gathered metrics do not contain correct value for total requests, got %d", cv)
				}
			},
		},
		{
			name: latencyName,
			labels: map[string]string{
				"service": "test",
			},
			assert: func(family *dto.MetricFamily) {
				sc := family.Metric[0].Histogram.GetSampleCount()
				if sc != 3 {
					t.Errorf("gathered metrics do not contain correct sample count for request duration, got %d", sc)
				}
			},
		},
	}

	assert.Equal(t, len(tests), len(metricsFamily)-initialMetricsFamilyCount, "gathered traefic metrics count does not match tests count")

	for _, test := range tests {
		family := findMetricFamily(test.name, metricsFamily)
		if family == nil {
			t.Errorf("gathered metrics do not contain '%s'", test.name)
			continue
		}
		for _, label := range family.Metric[0].Label {
			val, ok := test.labels[*label.Name]
			if !ok {
				t.Errorf("'%s' metric contains unexpected label '%s'", test.name, label)
			} else if val != *label.Value {
				t.Errorf("label '%s' in metric '%s' has wrong value '%s'", label, test.name, *label.Value)
			}
		}
		test.assert(family)
	}
}

func findMetricFamily(name string, families []*dto.MetricFamily) *dto.MetricFamily {
	for _, family := range families {
		if family.GetName() == name {
			return family
		}
	}
	return nil
}
