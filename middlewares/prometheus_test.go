package middlewares

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/negroni"
)

func TestPrometheus(t *testing.T) {
	defer resetPrometheusValues()

	metricsFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("could not gather metrics family: %s", err)
	}
	initialMetricsFamilyCount := len(metricsFamilies)

	recorder := httptest.NewRecorder()

	req1 := testhelpers.MustNewRequest(http.MethodGet, "http://localhost:3000/ok", ioutil.NopCloser(nil))
	req2 := testhelpers.MustNewRequest(http.MethodGet, "http://localhost:3000/metrics", ioutil.NopCloser(nil))

	httpHandler := setupTestHTTPHandler()
	httpHandler.ServeHTTP(recorder, req1)
	httpHandler.ServeHTTP(recorder, req2)

	body := recorder.Body.String()
	if !strings.Contains(body, reqsTotalName) {
		t.Errorf("body does not contain request total entry '%s'", reqsTotalName)
	}
	if !strings.Contains(body, reqDurationName) {
		t.Errorf("body does not contain request duration entry '%s'", reqDurationName)
	}
	if !strings.Contains(body, retriesTotalName) {
		t.Errorf("body does not contain total retries entry '%s'", retriesTotalName)
	}

	metricsFamilies, err = prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("could not gather metrics families: %s", err)
	}

	tests := []struct {
		name   string
		labels map[string]string
		assert func(*dto.MetricFamily)
	}{
		{
			name: reqsTotalName,
			labels: map[string]string{
				"code":    "200",
				"method":  http.MethodGet,
				"service": "test",
			},
			assert: func(family *dto.MetricFamily) {
				cv := uint(family.Metric[0].Counter.GetValue())
				expectedCv := uint(2)
				if cv != expectedCv {
					t.Errorf("gathered metrics do not contain correct value for total requests, got %d expected %d", cv, expectedCv)
				}
			},
		},
		{
			name: reqDurationName,
			labels: map[string]string{
				"service": "test",
				"code":    "200",
			},
			assert: func(family *dto.MetricFamily) {
				sc := family.Metric[0].Histogram.GetSampleCount()
				expectedSc := uint64(2)
				if sc != expectedSc {
					t.Errorf("gathered metrics do not contain correct sample count for request duration, got %d expected %d", sc, expectedSc)
				}
			},
		},
		{
			name: retriesTotalName,
			labels: map[string]string{
				"service": "test",
			},
			assert: func(family *dto.MetricFamily) {
				cv := uint(family.Metric[0].Counter.GetValue())
				expectedCv := uint(1)
				if cv != expectedCv {
					t.Errorf("gathered metrics do not contain correct value for total retries, got '%d' expected '%d'", cv, expectedCv)
				}
			},
		},
	}

	assert.Equal(t, len(tests), len(metricsFamilies)-initialMetricsFamilyCount, "gathered Traefik metrics count does not match tests count")

	for _, test := range tests {
		family := findMetricFamily(test.name, metricsFamilies)
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

func TestPrometheusRegisterMetricsMultipleTimes(t *testing.T) {
	defer resetPrometheusValues()

	recorder := httptest.NewRecorder()
	req1 := testhelpers.MustNewRequest(http.MethodGet, "http://localhost:3000/ok", ioutil.NopCloser(nil))

	httpHandler := setupTestHTTPHandler()
	httpHandler.ServeHTTP(recorder, req1)

	httpHandler = setupTestHTTPHandler()
	httpHandler.ServeHTTP(recorder, req1)

	metricsFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("could not gather metrics families: %s", err)
	}

	reqsTotalFamily := findMetricFamily(reqsTotalName, metricsFamilies)
	if reqsTotalFamily == nil {
		t.Fatalf("gathered metrics do not contain '%s'", reqsTotalName)
	}

	cv := uint(reqsTotalFamily.Metric[0].Counter.GetValue())
	expectedCv := uint(2)
	if cv != expectedCv {
		t.Errorf("wrong counter value when registering metrics multiple times, got '%d' expected '%d'", cv, expectedCv)
	}
}

func setupTestHTTPHandler() http.Handler {
	serveMux := http.NewServeMux()
	serveMux.Handle("/metrics", promhttp.Handler())
	serveMux.Handle("/ok", &networkFailingHTTPHandler{failAtCalls: []int{1}, netErrorRecorder: &DefaultNetErrorRecorder{}})

	metrics, _ := newPrometheusMetrics()

	n := negroni.New()
	n.Use(NewMetricsWrapper(metrics))
	n.UseHandler(NewRetry(2, serveMux, NewMetricsRetryListener(metrics)))

	return n
}

func resetPrometheusValues() {
	_, collectors := newPrometheusMetrics()

	for _, collector := range collectors {
		prometheus.Unregister(collector)
	}
}

func newPrometheusMetrics() (*Prometheus, []prometheus.Collector) {
	prom, collectors, err := NewPrometheus("test", &types.Prometheus{})
	if err != nil {
		panic(fmt.Sprintf("Error creating Prometheus Metrics: %s", err))
	}
	return prom, collectors
}

func findMetricFamily(name string, families []*dto.MetricFamily) *dto.MetricFamily {
	for _, family := range families {
		if family.GetName() == name {
			return family
		}
	}
	return nil
}
