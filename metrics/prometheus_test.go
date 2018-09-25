package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	th "github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestRegisterPromState(t *testing.T) {
	// Reset state of global promState.
	defer promState.reset()

	testCases := []struct {
		desc                 string
		prometheusSlice      []*types.Prometheus
		initPromState        bool
		unregisterPromState  bool
		expectedNbRegistries int
	}{
		{
			desc:                 "Register once",
			prometheusSlice:      []*types.Prometheus{{}},
			expectedNbRegistries: 1,
			initPromState:        true,
		},
		{
			desc:                 "Register once with no promState init",
			prometheusSlice:      []*types.Prometheus{{}},
			expectedNbRegistries: 0,
		},
		{
			desc:                 "Register twice",
			prometheusSlice:      []*types.Prometheus{{}, {}},
			expectedNbRegistries: 2,
			initPromState:        true,
		},
		{
			desc:                 "Register twice with no promstate init",
			prometheusSlice:      []*types.Prometheus{{}, {}},
			expectedNbRegistries: 0,
		},
		{
			desc:                 "Register twice with unregister",
			prometheusSlice:      []*types.Prometheus{{}, {}},
			unregisterPromState:  true,
			expectedNbRegistries: 2,
			initPromState:        true,
		},
		{
			desc:                 "Register twice with unregister but no promstate init",
			prometheusSlice:      []*types.Prometheus{{}, {}},
			unregisterPromState:  true,
			expectedNbRegistries: 0,
		},
	}

	for _, test := range testCases {
		actualNbRegistries := 0
		for _, prom := range test.prometheusSlice {
			if test.initPromState {
				initStandardRegistry(prom)
			}

			promReg := registerPromState()
			if promReg != false {
				actualNbRegistries++
			}

			if test.unregisterPromState {
				prometheus.Unregister(promState)
			}

			promState.reset()
		}

		prometheus.Unregister(promState)

		assert.Equal(t, test.expectedNbRegistries, actualNbRegistries)
	}
}

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
	configurations["providerName"] = th.BuildConfiguration(
		th.WithFrontends(
			th.WithFrontend("backend1", th.WithEntryPoints("entrypoint1")),
		),
		th.WithBackends(
			th.WithBackendNew("backend1", th.WithServersNew(th.WithServerNew("http://localhost:9000"))),
		),
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

func TestPrometheusRemovedMetricsReset(t *testing.T) {
	// Reset state of global promState.
	defer promState.reset()

	prometheusRegistry := RegisterPrometheus(&types.Prometheus{})
	defer prometheus.Unregister(promState)

	labelNamesValues := []string{
		"backend", "backend",
		"code", strconv.Itoa(http.StatusOK),
		"method", http.MethodGet,
		"protocol", "http",
	}
	prometheusRegistry.
		BackendReqsCounter().
		With(labelNamesValues...).
		Add(3)

	delayForTrackingCompletion()

	metricsFamilies := mustScrape()
	assertCounterValue(t, 3, findMetricFamily(backendReqsTotalName, metricsFamilies), labelNamesValues...)

	// There is no dynamic configuration and so this metric will be deleted
	// after the first scrape.
	assertMetricsAbsent(t, mustScrape(), backendReqsTotalName)

	prometheusRegistry.
		BackendReqsCounter().
		With(labelNamesValues...).
		Add(1)

	delayForTrackingCompletion()

	metricsFamilies = mustScrape()
	assertCounterValue(t, 1, findMetricFamily(backendReqsTotalName, metricsFamilies), labelNamesValues...)
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

func findMetricByLabelNamesValues(family *dto.MetricFamily, labelNamesValues ...string) *dto.Metric {
	if family == nil {
		return nil
	}

	for _, metric := range family.Metric {
		if hasMetricAllLabelPairs(metric, labelNamesValues...) {
			return metric
		}
	}

	return nil
}

func hasMetricAllLabelPairs(metric *dto.Metric, labelNamesValues ...string) bool {
	for i := 0; i < len(labelNamesValues); i += 2 {
		name, val := labelNamesValues[i], labelNamesValues[i+1]
		if !hasMetricLabelPair(metric, name, val) {
			return false
		}
	}
	return true
}

func hasMetricLabelPair(metric *dto.Metric, labelName, labelValue string) bool {
	for _, labelPair := range metric.Label {
		if labelPair.GetName() == labelName && labelPair.GetValue() == labelValue {
			return true
		}
	}
	return false
}

func assertCounterValue(t *testing.T, want float64, family *dto.MetricFamily, labelNamesValues ...string) {
	t.Helper()

	metric := findMetricByLabelNamesValues(family, labelNamesValues...)

	if metric == nil {
		t.Error("metric must not be nil")
		return
	}
	if metric.Counter == nil {
		t.Errorf("metric %s must be a counter", family.GetName())
		return
	}

	if cv := metric.Counter.GetValue(); cv != want {
		t.Errorf("metric %s has value %v, want %v", family.GetName(), cv, want)
	}
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
