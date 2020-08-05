package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPilotCounter(t *testing.T) {
	rootCounter := newPilotCounter("rootCounter")

	// Checks that a counter without labels can be incremented.
	rootCounter.Add(1)
	assertPilotCounterValue(t, 1.0, "", rootCounter)

	// Checks that a counter with labels can be incremented.
	counterWithLabels := rootCounter.With("foo", "bar", "foo", "buz")

	counterWithLabels.Add(1)
	assertPilotCounterValue(t, 1.0, "foo,bar,foo,buz", counterWithLabels)

	// Checks that the derived counter value has not changed.
	assertPilotCounterValue(t, 1.0, "", rootCounter)

	// Checks that an existing counter (with the same labels) can be incremented.
	existingCounterWithLabels := rootCounter.With("foo", "bar").With("foo", "buz")

	existingCounterWithLabels.Add(1)
	assertPilotCounterValue(t, 2.0, "foo,bar,foo,buz", existingCounterWithLabels)
}

func assertPilotCounterValue(t *testing.T, expValue float64, labels string, c metrics.Counter) {
	t.Helper()
	counter, ok := c.(*pilotCounter).counters.Load(labels)

	require.True(t, ok)
	assert.Equal(t, expValue, counter.(*pilotCounter).c.Value())
}

func TestPilotGauge(t *testing.T) {
	rootGauge := newPilotGauge("rootGauge")

	// Checks that a gauge without labels can be incremented.
	rootGauge.Add(1)

	assertPilotGaugeValue(t, 1.0, "", rootGauge)

	// Checks that a gauge (without labels) value can be set.
	rootGauge.Set(5.0)

	assertPilotGaugeValue(t, 5.0, "", rootGauge)

	// Checks that a gauge with labels can be incremented.
	gaugeWithLabels := rootGauge.With("foo", "bar", "foo", "buz")
	gaugeWithLabels.Add(1)

	assertPilotGaugeValue(t, 1.0, "foo,bar,foo,buz", gaugeWithLabels)

	// Checks that the derived gauge value has not changed.
	assertPilotGaugeValue(t, 5.0, "", rootGauge)

	// Checks that an existing gauge (with the same labels) can be incremented.
	existingGaugeWithLabels := rootGauge.With("foo", "bar").With("foo", "buz")
	existingGaugeWithLabels.Add(1)

	assertPilotGaugeValue(t, 2.0, "foo,bar,foo,buz", existingGaugeWithLabels)
}

func assertPilotGaugeValue(t *testing.T, expValue float64, labels string, g metrics.Gauge) {
	t.Helper()
	gauge, ok := g.(*pilotGauge).gauges.Load(labels)

	require.True(t, ok)
	assert.Equal(t, expValue, gauge.(*pilotGauge).g.Value())
}

func TestPilotHistogram(t *testing.T) {
	rootHistogram := newPilotHistogram("rootHistogram")

	// Checks that an histogram without labels can be updated.
	rootHistogram.Observe(1)

	assertPilotHistogramValues(t, 1.0, 1.0, "", rootHistogram)

	rootHistogram.Observe(2)

	assertPilotHistogramValues(t, 2.0, 3.0, "", rootHistogram)

	// Checks that an histogram with labels can be updated.
	histogramWithLabels := rootHistogram.With("foo", "bar", "foo", "buz")
	histogramWithLabels.Observe(1)

	assertPilotHistogramValues(t, 1.0, 1.0, "foo,bar,foo,buz", histogramWithLabels)

	// Checks that the derived histogram has not changed.
	assertPilotHistogramValues(t, 2.0, 3.0, "", rootHistogram)

	// Checks that an existing histogram (with the same labels) can be updated.
	existingHistogramWithLabels := rootHistogram.With("foo", "bar").With("foo", "buz")
	existingHistogramWithLabels.Observe(1)

	assertPilotHistogramValues(t, 2.0, 2.0, "foo,bar,foo,buz", existingHistogramWithLabels)
}

func assertPilotHistogramValues(t *testing.T, expCount, expTotal float64, labels string, h metrics.Histogram) {
	t.Helper()
	histogram, ok := h.(*pilotHistogram).histograms.Load(labels)

	require.True(t, ok)
	assert.Equal(t, expCount, histogram.(*pilotHistogram).count.Value())
	assert.Equal(t, expTotal, histogram.(*pilotHistogram).total.Value())
}

func TestPilotMetrics(t *testing.T) {
	pilotRegistry := RegisterPilot()

	if !pilotRegistry.IsEpEnabled() || !pilotRegistry.IsSvcEnabled() {
		t.Errorf("PilotRegistry should return true for IsEnabled() and IsSvcEnabled()")
	}

	pilotRegistry.ConfigReloadsCounter().Add(1)
	pilotRegistry.ConfigReloadsFailureCounter().Add(1)
	pilotRegistry.LastConfigReloadSuccessGauge().Set(float64(time.Now().Unix()))
	pilotRegistry.LastConfigReloadFailureGauge().Set(float64(time.Now().Unix()))

	pilotRegistry.
		EntryPointReqsCounter().
		With("code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http", "entrypoint", "http").
		Add(1)
	pilotRegistry.
		EntryPointReqDurationHistogram().
		With("code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http", "entrypoint", "http").
		Observe(1)
	pilotRegistry.
		EntryPointOpenConnsGauge().
		With("method", http.MethodGet, "protocol", "http", "entrypoint", "http").
		Set(1)

	pilotRegistry.
		ServiceReqsCounter().
		With("service", "service1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	pilotRegistry.
		ServiceReqDurationHistogram().
		With("service", "service1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Observe(10000)
	pilotRegistry.
		ServiceOpenConnsGauge().
		With("service", "service1", "method", http.MethodGet, "protocol", "http").
		Set(1)
	pilotRegistry.
		ServiceRetriesCounter().
		With("service", "service1").
		Add(1)
	pilotRegistry.
		ServiceServerUpGauge().
		With("service", "service1", "url", "http://127.0.0.10:80").
		Set(1)

	data := pilotRegistry.Data()

	testCases := []struct {
		name   string
		labels map[string]string
		assert func(*PilotMetric)
	}{
		{
			name:   pilotConfigReloadsTotalName,
			assert: buildPilotCounterAssert(t, pilotConfigReloadsTotalName, 1),
		},
		{
			name:   pilotConfigReloadsFailuresTotalName,
			assert: buildPilotCounterAssert(t, pilotConfigReloadsFailuresTotalName, 1),
		},
		{
			name:   pilotConfigLastReloadSuccessName,
			assert: buildPilotTimestampAssert(t, pilotConfigLastReloadSuccessName),
		},
		{
			name:   pilotConfigLastReloadFailureName,
			assert: buildPilotTimestampAssert(t, pilotConfigLastReloadFailureName),
		},
		{
			name: pilotEntryPointReqsTotalName,
			labels: map[string]string{
				"code":       "200",
				"method":     http.MethodGet,
				"protocol":   "http",
				"entrypoint": "http",
			},
			assert: buildPilotCounterAssert(t, pilotEntryPointReqsTotalName, 1),
		},
		{
			name: pilotEntryPointReqDurationName,
			labels: map[string]string{
				"code":       "200",
				"method":     http.MethodGet,
				"protocol":   "http",
				"entrypoint": "http",
			},
			assert: buildPilotHistogramAssert(t, pilotEntryPointReqDurationName, 1),
		},
		{
			name: pilotEntryPointOpenConnsName,
			labels: map[string]string{
				"method":     http.MethodGet,
				"protocol":   "http",
				"entrypoint": "http",
			},
			assert: buildPilotGaugeAssert(t, pilotEntryPointOpenConnsName, 1),
		},
		{
			name: pilotServiceReqsTotalName,
			labels: map[string]string{
				"code":     "200",
				"method":   http.MethodGet,
				"protocol": "http",
				"service":  "service1",
			},
			assert: buildPilotCounterAssert(t, pilotServiceReqsTotalName, 1),
		},
		{
			name: pilotServiceReqDurationName,
			labels: map[string]string{
				"code":     "200",
				"method":   http.MethodGet,
				"protocol": "http",
				"service":  "service1",
			},
			assert: buildPilotHistogramAssert(t, pilotServiceReqDurationName, 1),
		},
		{
			name: pilotServiceOpenConnsName,
			labels: map[string]string{
				"method":   http.MethodGet,
				"protocol": "http",
				"service":  "service1",
			},
			assert: buildPilotGaugeAssert(t, pilotServiceOpenConnsName, 1),
		},
		{
			name: pilotServiceRetriesTotalName,
			labels: map[string]string{
				"service": "service1",
			},
			assert: buildPilotGreaterThanCounterAssert(t, pilotServiceRetriesTotalName, 1),
		},
		{
			name: pilotServiceServerUpName,
			labels: map[string]string{
				"service": "service1",
				"url":     "http://127.0.0.10:80",
			},
			assert: buildPilotGaugeAssert(t, pilotServiceServerUpName, 1),
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			metric := findPilotMetric(test.name, data)
			if metric == nil {
				t.Errorf("metrics do not contain %q", test.name)
				return
			}

			for labels := range metric.Observations {
				if len(labels)%2 == 0 {
					splitLabels := strings.Split(labels, ",")
					for i := 0; i < len(splitLabels); i += 2 {
						label := splitLabels[i]
						value := splitLabels[i+1]
						val, ok := test.labels[label]
						if !ok {
							t.Errorf("%q metric contains unexpected label %q", test.name, label)
						} else if val != value {
							t.Errorf("label %q in metric %q has wrong value %q, expected %q", label, test.name, value, val)
						}
					}
				}
			}
			test.assert(metric)
		})
	}
}

func findPilotMetric(name string, metrics []PilotMetric) *PilotMetric {
	for _, metric := range metrics {
		if metric.Name == name {
			return &metric
		}
	}
	return nil
}

func buildPilotCounterAssert(t *testing.T, metricName string, expectedValue float64) func(metric *PilotMetric) {
	return func(metric *PilotMetric) {
		for _, value := range metric.Observations {
			if cv := value.(float64); cv != expectedValue {
				t.Errorf("metric %s has value %f, want %f", metricName, cv, expectedValue)
			}
			break
		}
	}
}

func buildPilotGreaterThanCounterAssert(t *testing.T, metricName string, expectedMinValue float64) func(metric *PilotMetric) {
	return func(metric *PilotMetric) {
		for _, value := range metric.Observations {
			if cv := value.(float64); cv < expectedMinValue {
				t.Errorf("metric %s has value %f, want at least %f", metricName, cv, expectedMinValue)
			}
			break
		}
	}
}

func buildPilotHistogramAssert(t *testing.T, metricName string, expectedSampleCount float64) func(metric *PilotMetric) {
	return func(metric *PilotMetric) {
		for _, value := range metric.Observations {
			if pho := value.(*pilotHistogramObservation); pho.Count != expectedSampleCount {
				t.Errorf("metric %s has sample count value %f, want %f", metricName, pho, expectedSampleCount)
			}
			break
		}
	}
}

func buildPilotGaugeAssert(t *testing.T, metricName string, expectedValue float64) func(metric *PilotMetric) {
	return func(metric *PilotMetric) {
		for _, value := range metric.Observations {
			if gv := value.(float64); gv != expectedValue {
				t.Errorf("metric %s has value %f, want %f", metricName, gv, expectedValue)
			}
			break
		}
	}
}

func buildPilotTimestampAssert(t *testing.T, metricName string) func(metric *PilotMetric) {
	return func(metric *PilotMetric) {
		for _, value := range metric.Observations {
			if ts := time.Unix(int64(value.(float64)), 0); time.Since(ts) > time.Minute {
				t.Errorf("metric %s has wrong timestamp %v", metricName, ts)
			}
			break
		}
	}
}
