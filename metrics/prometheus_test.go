package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/containous/traefik/types"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestPrometheus(t *testing.T) {
	prometheusRegistry := RegisterPrometheus(&types.Prometheus{})
	defer prometheus.Unregister(promState)

	if !prometheusRegistry.IsEnabled() {
		t.Errorf("PrometheusRegistry should return true for IsEnabled()")
	}

	prometheusRegistry.ConfigReloadsCounter().Add(1)
	prometheusRegistry.ConfigReloadsFailureCounter().Add(1)
	prometheusRegistry.LastConfigReloadSuccessGauge().Set(float64(time.Now().Unix()))
	prometheusRegistry.LastConfigReloadFailureGauge().Set(float64(time.Now().Unix()))

	prometheusRegistry.
		EntrypointReqsCounter().
		With("code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http", "entrypoint", "http").
		Add(1)
	prometheusRegistry.
		EntrypointReqDurationHistogram().
		With("code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http", "entrypoint", "http").
		Observe(1)
	prometheusRegistry.
		EntrypointOpenConnsGauge().
		With("method", http.MethodGet, "protocol", "http", "entrypoint", "http").
		Set(1)

	prometheusRegistry.
		BackendReqsCounter().
		With("backend", "backend1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	prometheusRegistry.
		BackendReqDurationHistogram().
		With("backend", "backend1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Observe(10000)
	prometheusRegistry.
		BackendOpenConnsGauge().
		With("backend", "backend1", "method", http.MethodGet, "protocol", "http").
		Set(1)
	prometheusRegistry.
		BackendRetriesCounter().
		With("backend", "backend1").
		Add(1)
	prometheusRegistry.
		BackendServerUpGauge().
		With("backend", "backend1", "url", "http://127.0.0.10:80").
		Set(1)

	delayForTrackingCompletion()

	metricsFamilies := mustScrape()

	tests := []struct {
		name   string
		labels map[string]string
		assert func(*dto.MetricFamily)
	}{
		{
			name:   configReloadsTotalName,
			assert: buildCounterAssert(t, configReloadsTotalName, 1),
		},
		{
			name:   configReloadsFailuresTotalName,
			assert: buildCounterAssert(t, configReloadsFailuresTotalName, 1),
		},
		{
			name:   configLastReloadSuccessName,
			assert: buildTimestampAssert(t, configLastReloadSuccessName),
		},
		{
			name:   configLastReloadFailureName,
			assert: buildTimestampAssert(t, configLastReloadFailureName),
		},
		{
			name: entrypointReqsTotalName,
			labels: map[string]string{
				"code":       "200",
				"method":     http.MethodGet,
				"protocol":   "http",
				"entrypoint": "http",
			},
			assert: buildCounterAssert(t, entrypointReqsTotalName, 1),
		},
		{
			name: entrypointReqDurationName,
			labels: map[string]string{
				"code":       "200",
				"method":     http.MethodGet,
				"protocol":   "http",
				"entrypoint": "http",
			},
			assert: buildHistogramAssert(t, entrypointReqDurationName, 1),
		},
		{
			name: entrypointOpenConnsName,
			labels: map[string]string{
				"method":     http.MethodGet,
				"protocol":   "http",
				"entrypoint": "http",
			},
			assert: buildGaugeAssert(t, entrypointOpenConnsName, 1),
		},
		{
			name: backendReqsTotalName,
			labels: map[string]string{
				"code":     "200",
				"method":   http.MethodGet,
				"protocol": "http",
				"backend":  "backend1",
			},
			assert: buildCounterAssert(t, backendReqsTotalName, 1),
		},
		{
			name: backendReqDurationName,
			labels: map[string]string{
				"code":     "200",
				"method":   http.MethodGet,
				"protocol": "http",
				"backend":  "backend1",
			},
			assert: buildHistogramAssert(t, backendReqDurationName, 1),
		},
		{
			name: backendOpenConnsName,
			labels: map[string]string{
				"method":   http.MethodGet,
				"protocol": "http",
				"backend":  "backend1",
			},
			assert: buildGaugeAssert(t, backendOpenConnsName, 1),
		},
		{
			name: backendRetriesTotalName,
			labels: map[string]string{
				"backend": "backend1",
			},
			assert: buildGreaterThanCounterAssert(t, backendRetriesTotalName, 1),
		},
		{
			name: backendServerUpName,
			labels: map[string]string{
				"backend": "backend1",
				"url":     "http://127.0.0.10:80",
			},
			assert: buildGaugeAssert(t, backendServerUpName, 1),
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

func TestPrometheusGenerationLogicForMetricWithLabel(t *testing.T) {
	prometheusRegistry := RegisterPrometheus(&types.Prometheus{})
	defer prometheus.Unregister(promState)

	// Metrics with labels belonging to a specific configuration in Traefik
	// should be removed when the generationMaxAge is exceeded. As example
	// we use the traefik_backend_requests_total metric.
	prometheusRegistry.
		BackendReqsCounter().
		With("backend", "backend1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)

	delayForTrackingCompletion()

	assertMetricExists(t, backendReqsTotalName, mustScrape())

	// Increase the config generation one more than the max age of a metric.
	for i := 0; i < generationAgeDefault+1; i++ {
		OnConfigurationUpdate()
	}

	// On the next scrape the metric still exists and will be removed
	// after the scrape completed.
	assertMetricExists(t, backendReqsTotalName, mustScrape())

	// Now the metric should be absent.
	assertMetricAbsent(t, backendReqsTotalName, mustScrape())
}

func TestPrometheusGenerationLogicForMetricWithoutLabel(t *testing.T) {
	prometheusRegistry := RegisterPrometheus(&types.Prometheus{})
	defer prometheus.Unregister(promState)

	// Metrics without labels like traefik_config_reloads_total should live forever
	// and never get removed.
	prometheusRegistry.ConfigReloadsCounter().Add(1)

	delayForTrackingCompletion()

	assertMetricExists(t, configReloadsTotalName, mustScrape())

	// Increase the config generation one more than the max age of a metric.
	for i := 0; i < generationAgeDefault+100; i++ {
		OnConfigurationUpdate()
	}

	// Scrape two times in order to verify, that it is not removed after the
	// first scrape completed.
	assertMetricExists(t, configReloadsTotalName, mustScrape())
	assertMetricExists(t, configReloadsTotalName, mustScrape())
}

// Tracking and gathering the metrics happens concurrently.
// In practice this is no problem, because in case a tracked metric would miss
// the current scrape, it would just be there in the next one.
// That we can test reliably the tracking of all metrics here, we sleep
// for a short amount of time, to make sure the metric will be present
// in the next scrape.
func delayForTrackingCompletion() {
	time.Sleep(250 * time.Millisecond)
}

func mustScrape() []*dto.MetricFamily {
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		panic(fmt.Sprintf("could not gather metrics families: %s", err))
	}
	return families
}

func assertMetricExists(t *testing.T, name string, families []*dto.MetricFamily) {
	t.Helper()
	if findMetricFamily(name, families) == nil {
		t.Errorf("gathered metrics do not contain %q", name)
	}
}

func assertMetricAbsent(t *testing.T, name string, families []*dto.MetricFamily) {
	t.Helper()
	if findMetricFamily(name, families) != nil {
		t.Errorf("gathered metrics contain %q, but should not", name)
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

func buildCounterAssert(t *testing.T, metricName string, expectedValue int) func(family *dto.MetricFamily) {
	return func(family *dto.MetricFamily) {
		if cv := int(family.Metric[0].Counter.GetValue()); cv != expectedValue {
			t.Errorf("metric %s has value %d, want %d", metricName, cv, expectedValue)
		}
	}
}

func buildGreaterThanCounterAssert(t *testing.T, metricName string, expectedMinValue int) func(family *dto.MetricFamily) {
	return func(family *dto.MetricFamily) {
		if cv := int(family.Metric[0].Counter.GetValue()); cv < expectedMinValue {
			t.Errorf("metric %s has value %d, want at least %d", metricName, cv, expectedMinValue)
		}
	}
}

func buildHistogramAssert(t *testing.T, metricName string, expectedSampleCount int) func(family *dto.MetricFamily) {
	return func(family *dto.MetricFamily) {
		if sc := int(family.Metric[0].Histogram.GetSampleCount()); sc != expectedSampleCount {
			t.Errorf("metric %s has sample count value %d, want %d", metricName, sc, expectedSampleCount)
		}
	}
}

func buildGaugeAssert(t *testing.T, metricName string, expectedValue int) func(family *dto.MetricFamily) {
	return func(family *dto.MetricFamily) {
		if gv := int(family.Metric[0].Gauge.GetValue()); gv != expectedValue {
			t.Errorf("metric %s has value %d, want %d", metricName, gv, expectedValue)
		}
	}
}

func buildTimestampAssert(t *testing.T, metricName string) func(family *dto.MetricFamily) {
	return func(family *dto.MetricFamily) {
		if ts := time.Unix(int64(family.Metric[0].Gauge.GetValue()), 0); time.Since(ts) > time.Minute {
			t.Errorf("metric %s has wrong timestamp %v", metricName, ts)
		}
	}
}
