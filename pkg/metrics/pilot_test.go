package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPilotCounter(t *testing.T) {
	pc := newPilotCounter("foo")

	pc.Add(1)

	counter, ok := pc.counters.Load("")
	assert.True(t, ok)
	assert.Equal(t, 1.0, counter.(*pilotCounter).c.Value())

	fooBar := pc.With("foo", "bar", "foo", "buz")
	fooBar.Add(1)

	counter, ok = pc.counters.Load("foo,bar,foo,buz")
	assert.True(t, ok)
	assert.Equal(t, 1.0, counter.(*pilotCounter).c.Value())

	counter, ok = pc.counters.Load("")
	assert.True(t, ok)
	assert.Equal(t, 1.0, counter.(*pilotCounter).c.Value())

	fooBar2 := pc.With("foo", "bar").With("foo", "buz")
	fooBar2.Add(1)

	counter, ok = pc.counters.Load("foo,bar,foo,buz")
	assert.True(t, ok)
	assert.Equal(t, 2.0, counter.(*pilotCounter).c.Value())
}

func TestPilotGauge(t *testing.T) {
	pg := newPilotGauge("foo")

	pg.Add(1)

	gauge, ok := pg.gauges.Load("")
	assert.True(t, ok)
	assert.Equal(t, 1.0, gauge.(*pilotGauge).g.Value())

	pg.Set(5.0)

	gauge, ok = pg.gauges.Load("")
	assert.True(t, ok)
	assert.Equal(t, 5.0, gauge.(*pilotGauge).g.Value())

	fooBar := pg.With("foo", "bar", "foo", "buz")
	fooBar.Add(1)

	gauge, ok = pg.gauges.Load("")
	assert.True(t, ok)
	assert.Equal(t, 5.0, gauge.(*pilotGauge).g.Value())

	gauge, ok = pg.gauges.Load("foo,bar,foo,buz")
	assert.True(t, ok)
	assert.Equal(t, 1.0, gauge.(*pilotGauge).g.Value())

	fooBar2 := pg.With("foo", "bar").With("foo", "buz")
	fooBar2.Add(1)

	gauge, ok = pg.gauges.Load("foo,bar,foo,buz")
	assert.True(t, ok)
	assert.Equal(t, 2.0, gauge.(*pilotGauge).g.Value())
}

func TestPilotHistogram(t *testing.T) {
	ph := newPilotHistogram("foo")
	ph.Observe(1)

	histogram, ok := ph.histograms.Load("")
	assert.True(t, ok)
	assert.Equal(t, 1.0, histogram.(*pilotHistogram).count.Value())
	assert.Equal(t, 1.0, histogram.(*pilotHistogram).total.Value())

	ph.Observe(2)

	histogram, ok = ph.histograms.Load("")
	assert.True(t, ok)
	assert.Equal(t, 2.0, histogram.(*pilotHistogram).count.Value())
	assert.Equal(t, 3.0, histogram.(*pilotHistogram).total.Value())

	fooBar := ph.With("foo", "bar", "foo", "buz")
	fooBar.Observe(1)

	histogram, ok = ph.histograms.Load("")
	assert.True(t, ok)
	assert.Equal(t, 2.0, histogram.(*pilotHistogram).count.Value())
	assert.Equal(t, 3.0, histogram.(*pilotHistogram).total.Value())

	histogram, ok = ph.histograms.Load("foo,bar,foo,buz")
	assert.True(t, ok)
	assert.Equal(t, 1.0, histogram.(*pilotHistogram).count.Value())
	assert.Equal(t, 1.0, histogram.(*pilotHistogram).total.Value())

	fooBar2 := ph.With("foo", "bar").With("foo", "buz")
	fooBar2.Observe(1)

	histogram, ok = ph.histograms.Load("foo,bar,foo,buz")
	assert.True(t, ok)
	assert.Equal(t, 2.0, histogram.(*pilotHistogram).count.Value())
	assert.Equal(t, 2.0, histogram.(*pilotHistogram).total.Value())
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
