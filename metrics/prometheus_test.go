package metrics

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/containous/traefik/types"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestPrometheus(t *testing.T) {
	prometheusRegistry := RegisterPrometheus(&types.Prometheus{})

	if !prometheusRegistry.IsEnabled() {
		t.Errorf("PrometheusRegistry should return true for IsEnabled()")
	}
	prometheusRegistry.ReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	prometheusRegistry.ReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	prometheusRegistry.ReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
	prometheusRegistry.ReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
	prometheusRegistry.RetriesCounter().With("service", "test").Add(1)

	metricsFamilies, err := prometheus.DefaultGatherer.Gather()
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
				cv := family.Metric[0].Counter.GetValue()
				expectedCv := float64(2)
				if cv != expectedCv {
					t.Errorf("gathered metrics do not contain correct value for total requests, got %f expected %f", cv, expectedCv)
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
				cv := family.Metric[0].Counter.GetValue()
				expectedCv := float64(1)
				if cv != expectedCv {
					t.Errorf("gathered metrics do not contain correct value for total retries, got %f expected %f", cv, expectedCv)
				}
			},
		},
	}

	for _, test := range tests {
		family := findMetricFamily(test.name, metricsFamilies)
		if family == nil {
			t.Errorf("gathered metrics do not contain %q", test.name)
			continue
		}
		for _, label := range family.Metric[0].Label {
			val, ok := test.labels[*label.Name]
			if !ok {
				t.Errorf("%q metric contains unexpected label %q", test.name, *label.Name)
			} else if val != *label.Value {
				t.Errorf("label %q in metric %q has wrong value %q, expected %q", *label.Name, test.name, *label.Value, val)
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
