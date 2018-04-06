package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestPrometheus(t *testing.T) {
	// Reset state of global promState.
	defer promState.reset()

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

func TestPrometheusMetricRemoval(t *testing.T) {
	// Reset state of global promState.
	defer promState.reset()

	prometheusRegistry := RegisterPrometheus(&types.Prometheus{})
	defer prometheus.Unregister(promState)

	configurations := make(types.Configurations)
	configurations["providerName"] = testhelpers.BuildDynamicConfig(
		testhelpers.WithFrontend("frontend", testhelpers.BuildFrontend(
			testhelpers.WithEntrypoint("entrypoint1"),
		)),
		testhelpers.WithBackend("backend1", testhelpers.BuildBackend(
			testhelpers.WithServer("server1", "http://localhost:9000"),
		)),
	)
	OnConfigurationUpdate(configurations)

	// Register some metrics manually that are not part of the active configuration.
	// Those metrics should be part of the /metrics output on the first scrape but
	// should be removed after that scrape.
	prometheusRegistry.
		EntrypointReqsCounter().
		With("entrypoint", "entrypoint2", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	prometheusRegistry.
		BackendReqsCounter().
		With("backend", "backend2", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	prometheusRegistry.
		BackendServerUpGauge().
		With("backend", "backend1", "url", "http://localhost:9999").
		Set(1)

	delayForTrackingCompletion()

	assertMetricsExist(t, mustScrape(), entrypointReqsTotalName, backendReqsTotalName, backendServerUpName)
	assertMetricsAbsent(t, mustScrape(), entrypointReqsTotalName, backendReqsTotalName, backendServerUpName)

	// To verify that metrics belonging to active configurations are not removed
	// here the counter examples.
	prometheusRegistry.
		EntrypointReqsCounter().
		With("entrypoint", "entrypoint1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)

	delayForTrackingCompletion()

	assertMetricsExist(t, mustScrape(), entrypointReqsTotalName)
	assertMetricsExist(t, mustScrape(), entrypointReqsTotalName)
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

func assertMetricsExist(t *testing.T, families []*dto.MetricFamily, metricNames ...string) {
	t.Helper()

	for _, metricName := range metricNames {
		if findMetricFamily(metricName, families) == nil {
			t.Errorf("gathered metrics should contain %q", metricName)
		}
	}
}

func assertMetricsAbsent(t *testing.T, families []*dto.MetricFamily, metricNames ...string) {
	t.Helper()

	for _, metricName := range metricNames {
		if findMetricFamily(metricName, families) != nil {
			t.Errorf("gathered metrics should not contain %q", metricName)
		}
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
