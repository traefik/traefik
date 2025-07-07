package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	th "github.com/traefik/traefik/v3/pkg/testhelpers"
	"github.com/traefik/traefik/v3/pkg/types"
)

func TestRegisterPromState(t *testing.T) {
	t.Cleanup(promState.reset)

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
			initPromState:        true,
			unregisterPromState:  false,
			expectedNbRegistries: 1,
		},
		{
			desc:                 "Register once with no promState init",
			prometheusSlice:      []*types.Prometheus{{}},
			initPromState:        false,
			unregisterPromState:  false,
			expectedNbRegistries: 1,
		},
		{
			desc:                 "Register twice",
			prometheusSlice:      []*types.Prometheus{{}, {}},
			initPromState:        true,
			unregisterPromState:  false,
			expectedNbRegistries: 2,
		},
		{
			desc:                 "Register twice with no promstate init",
			prometheusSlice:      []*types.Prometheus{{}, {}},
			initPromState:        false,
			unregisterPromState:  false,
			expectedNbRegistries: 2,
		},
		{
			desc:                 "Register twice with unregister",
			prometheusSlice:      []*types.Prometheus{{}, {}},
			initPromState:        true,
			unregisterPromState:  true,
			expectedNbRegistries: 2,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			actualNbRegistries := 0
			for _, prom := range test.prometheusSlice {
				if test.initPromState {
					initStandardRegistry(prom)
				}
				if registerPromState(t.Context()) {
					actualNbRegistries++
				}
				if test.unregisterPromState {
					promRegistry.Unregister(promState)
				}

				promState.reset()
			}

			promRegistry.Unregister(promState)
			assert.Equal(t, test.expectedNbRegistries, actualNbRegistries)
		})
	}
}

func TestPrometheus(t *testing.T) {
	promState = newPrometheusState()
	promRegistry = prometheus.NewRegistry()
	t.Cleanup(promState.reset)

	prometheusRegistry := RegisterPrometheus(t.Context(), &types.Prometheus{
		AddEntryPointsLabels: true,
		AddRoutersLabels:     true,
		AddServicesLabels:    true,
		HeaderLabels:         map[string]string{"useragent": "User-Agent"},
	})
	defer promRegistry.Unregister(promState)

	if !prometheusRegistry.IsEpEnabled() || !prometheusRegistry.IsRouterEnabled() || !prometheusRegistry.IsSvcEnabled() {
		t.Errorf("PrometheusRegistry should return true for IsEnabled(), IsRouterEnabled() and IsSvcEnabled()")
	}

	prometheusRegistry.ConfigReloadsCounter().Add(1)
	prometheusRegistry.LastConfigReloadSuccessGauge().Set(float64(time.Now().Unix()))
	prometheusRegistry.
		OpenConnectionsGauge().
		With("entrypoint", "test", "protocol", "TCP").
		Set(1)

	prometheusRegistry.
		TLSCertsNotAfterTimestampGauge().
		With("cn", "value", "serial", "value", "sans", "value").
		Set(float64(time.Now().Unix()))

	prometheusRegistry.
		EntryPointReqsCounter().
		With(map[string][]string{"User-Agent": {"foobar"}}, "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http", "entrypoint", "http").
		Add(1)
	prometheusRegistry.
		EntryPointReqDurationHistogram().
		With("code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http", "entrypoint", "http").
		Observe(1)
	prometheusRegistry.
		EntryPointRespsBytesCounter().
		With("code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http", "entrypoint", "http").
		Add(1)
	prometheusRegistry.
		EntryPointReqsBytesCounter().
		With("code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http", "entrypoint", "http").
		Add(1)

	prometheusRegistry.
		RouterReqsCounter().
		With(nil, "router", "demo", "service", "service1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	prometheusRegistry.
		RouterReqsTLSCounter().
		With("router", "demo", "service", "service1", "tls_version", "foo", "tls_cipher", "bar").
		Add(1)
	prometheusRegistry.
		RouterReqDurationHistogram().
		With("router", "demo", "service", "service1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Observe(10000)
	prometheusRegistry.
		RouterRespsBytesCounter().
		With("router", "demo", "service", "service1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	prometheusRegistry.
		RouterReqsBytesCounter().
		With("router", "demo", "service", "service1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)

	prometheusRegistry.
		ServiceReqsCounter().
		With(map[string][]string{"User-Agent": {"foobar"}}, "service", "service1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	prometheusRegistry.
		ServiceReqsTLSCounter().
		With("service", "service1", "tls_version", "foo", "tls_cipher", "bar").
		Add(1)
	prometheusRegistry.
		ServiceReqDurationHistogram().
		With("service", "service1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Observe(10000)
	prometheusRegistry.
		ServiceRetriesCounter().
		With("service", "service1").
		Add(1)
	prometheusRegistry.
		ServiceServerUpGauge().
		With("service", "service1", "url", "http://127.0.0.10:80").
		Set(1)
	prometheusRegistry.
		ServiceRespsBytesCounter().
		With("service", "service1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	prometheusRegistry.
		ServiceReqsBytesCounter().
		With("service", "service1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)

	delayForTrackingCompletion()

	metricsFamilies := mustScrape()

	testCases := []struct {
		name   string
		labels map[string]string
		assert func(*dto.MetricFamily)
	}{
		{
			name:   configReloadsTotalName,
			assert: buildCounterAssert(t, configReloadsTotalName, 1),
		},
		{
			name:   configLastReloadSuccessName,
			assert: buildTimestampAssert(t, configLastReloadSuccessName),
		},
		{
			name: openConnectionsName,
			labels: map[string]string{
				"protocol":   "TCP",
				"entrypoint": "test",
			},
			assert: buildGaugeAssert(t, openConnectionsName, 1),
		},
		{
			name: tlsCertsNotAfterTimestampName,
			labels: map[string]string{
				"cn":     "value",
				"serial": "value",
				"sans":   "value",
			},
			assert: buildTimestampAssert(t, tlsCertsNotAfterTimestampName),
		},
		{
			name: entryPointReqsTotalName,
			labels: map[string]string{
				"code":       "200",
				"method":     http.MethodGet,
				"protocol":   "http",
				"entrypoint": "http",
				"useragent":  "foobar",
			},
			assert: buildCounterAssert(t, entryPointReqsTotalName, 1),
		},
		{
			name: entryPointReqDurationName,
			labels: map[string]string{
				"code":       "200",
				"method":     http.MethodGet,
				"protocol":   "http",
				"entrypoint": "http",
			},
			assert: buildHistogramAssert(t, entryPointReqDurationName, 1),
		},
		{
			name: entryPointReqsBytesTotalName,
			labels: map[string]string{
				"code":       "200",
				"method":     http.MethodGet,
				"protocol":   "http",
				"entrypoint": "http",
			},
			assert: buildCounterAssert(t, entryPointReqsBytesTotalName, 1),
		},
		{
			name: entryPointRespsBytesTotalName,
			labels: map[string]string{
				"code":       "200",
				"method":     http.MethodGet,
				"protocol":   "http",
				"entrypoint": "http",
			},
			assert: buildCounterAssert(t, entryPointRespsBytesTotalName, 1),
		},
		{
			name: routerReqsTotalName,
			labels: map[string]string{
				"code":      "200",
				"method":    http.MethodGet,
				"protocol":  "http",
				"service":   "service1",
				"router":    "demo",
				"useragent": "",
			},
			assert: buildCounterAssert(t, routerReqsTotalName, 1),
		},
		{
			name: routerReqsTLSTotalName,
			labels: map[string]string{
				"service":     "service1",
				"router":      "demo",
				"tls_version": "foo",
				"tls_cipher":  "bar",
			},
			assert: buildCounterAssert(t, routerReqsTLSTotalName, 1),
		},
		{
			name: routerReqDurationName,
			labels: map[string]string{
				"code":     "200",
				"method":   http.MethodGet,
				"protocol": "http",
				"service":  "service1",
				"router":   "demo",
			},
			assert: buildHistogramAssert(t, routerReqDurationName, 1),
		},
		{
			name: routerReqsBytesTotalName,
			labels: map[string]string{
				"code":     "200",
				"method":   http.MethodGet,
				"protocol": "http",
				"service":  "service1",
				"router":   "demo",
			},
			assert: buildCounterAssert(t, routerReqsBytesTotalName, 1),
		},
		{
			name: routerRespsBytesTotalName,
			labels: map[string]string{
				"code":     "200",
				"method":   http.MethodGet,
				"protocol": "http",
				"service":  "service1",
				"router":   "demo",
			},
			assert: buildCounterAssert(t, routerRespsBytesTotalName, 1),
		},
		{
			name: serviceReqsTotalName,
			labels: map[string]string{
				"code":      "200",
				"method":    http.MethodGet,
				"protocol":  "http",
				"service":   "service1",
				"useragent": "foobar",
			},
			assert: buildCounterAssert(t, serviceReqsTotalName, 1),
		},
		{
			name: serviceReqsTLSTotalName,
			labels: map[string]string{
				"service":     "service1",
				"tls_version": "foo",
				"tls_cipher":  "bar",
			},
			assert: buildCounterAssert(t, serviceReqsTLSTotalName, 1),
		},
		{
			name: serviceReqDurationName,
			labels: map[string]string{
				"code":     "200",
				"method":   http.MethodGet,
				"protocol": "http",
				"service":  "service1",
			},
			assert: buildHistogramAssert(t, serviceReqDurationName, 1),
		},
		{
			name: serviceRetriesTotalName,
			labels: map[string]string{
				"service": "service1",
			},
			assert: buildGreaterThanCounterAssert(t, serviceRetriesTotalName, 1),
		},
		{
			name: serviceServerUpName,
			labels: map[string]string{
				"service": "service1",
				"url":     "http://127.0.0.10:80",
			},
			assert: buildGaugeAssert(t, serviceServerUpName, 1),
		},
		{
			name: serviceReqsBytesTotalName,
			labels: map[string]string{
				"code":     "200",
				"method":   http.MethodGet,
				"protocol": "http",
				"service":  "service1",
			},
			assert: buildCounterAssert(t, serviceReqsBytesTotalName, 1),
		},
		{
			name: serviceRespsBytesTotalName,
			labels: map[string]string{
				"code":     "200",
				"method":   http.MethodGet,
				"protocol": "http",
				"service":  "service1",
			},
			assert: buildCounterAssert(t, serviceRespsBytesTotalName, 1),
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			family := findMetricFamily(test.name, metricsFamilies)
			if family == nil {
				t.Errorf("gathered metrics do not contain %q", test.name)
				return
			}

			for _, label := range family.GetMetric()[0].GetLabel() {
				val, ok := test.labels[label.GetName()]
				if !ok {
					t.Errorf("%q metric contains unexpected label %q", test.name, label.GetName())
				} else if val != label.GetValue() {
					t.Errorf("label %q in metric %q has wrong value %q, expected %q", label.GetName(), test.name, label.GetValue(), val)
				}
			}
			test.assert(family)
		})
	}
}

func TestPrometheusMetricRemoval(t *testing.T) {
	promState = newPrometheusState()
	promRegistry = prometheus.NewRegistry()
	t.Cleanup(promState.reset)

	prometheusRegistry := RegisterPrometheus(t.Context(), &types.Prometheus{AddEntryPointsLabels: true, AddServicesLabels: true, AddRoutersLabels: true})
	defer promRegistry.Unregister(promState)

	conf1 := dynamic.Configuration{
		HTTP: th.BuildConfiguration(
			th.WithRouters(
				th.WithRouter("foo@providerName", th.WithServiceName("bar")),
				th.WithRouter("router2", th.WithServiceName("bar@providerName")),
			),
			th.WithLoadBalancerServices(
				th.WithService("bar@providerName", th.WithServers(
					th.WithServer("http://localhost:9000"),
					th.WithServer("http://localhost:9999"),
					th.WithServer("http://localhost:9998"),
				)),
				th.WithService("service1", th.WithServers(th.WithServer("http://localhost:9000"))),
			),
		),
	}

	conf2 := dynamic.Configuration{
		HTTP: th.BuildConfiguration(
			th.WithRouters(
				th.WithRouter("foo@providerName", th.WithServiceName("bar")),
			),
			th.WithLoadBalancerServices(
				th.WithService("bar@providerName", th.WithServers(th.WithServer("http://localhost:9000"))),
			),
		),
	}

	OnConfigurationUpdate(conf1, []string{"entrypoint1", "entrypoint2"})
	OnConfigurationUpdate(conf2, []string{"entrypoint1"})

	// Register some metrics manually that are not part of the active configuration.
	// Those metrics should be part of the /metrics output on the first scrape but
	// should be removed after that scrape.
	prometheusRegistry.
		EntryPointReqsCounter().
		With(nil, "entrypoint", "entrypoint2", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	prometheusRegistry.
		RouterReqsCounter().
		With(nil, "router", "router2", "service", "bar@providerName", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	prometheusRegistry.
		ServiceReqsCounter().
		With(nil, "service", "service1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	prometheusRegistry.
		ServiceServerUpGauge().
		With("service", "bar@providerName", "url", "http://localhost:9999").
		Set(1)
	prometheusRegistry.
		ServiceServerUpGauge().
		With("service", "bar@providerName", "url", "http://localhost:9998").
		Set(1)

	assertMetricsExist(t, mustScrape(), entryPointReqsTotalName, routerReqsTotalName, serviceReqsTotalName, serviceServerUpName)
	assertMetricsAbsent(t, mustScrape(), entryPointReqsTotalName, routerReqsTotalName, serviceReqsTotalName, serviceServerUpName)

	// To verify that metrics belonging to active configurations are not removed
	// here the counter examples.
	prometheusRegistry.
		EntryPointReqsCounter().
		With(nil, "entrypoint", "entrypoint1", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	prometheusRegistry.
		RouterReqsCounter().
		With(nil, "router", "foo@providerName", "service", "bar", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	prometheusRegistry.
		ServiceReqsCounter().
		With(nil, "service", "bar@providerName", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet, "protocol", "http").
		Add(1)
	prometheusRegistry.
		ServiceServerUpGauge().
		With("service", "bar@providerName", "url", "http://localhost:9000").
		Set(1)

	delayForTrackingCompletion()

	assertMetricsExist(t, mustScrape(), entryPointReqsTotalName, serviceReqsTotalName, serviceServerUpName, routerReqsTotalName)
	assertMetricsExist(t, mustScrape(), entryPointReqsTotalName, serviceReqsTotalName, serviceServerUpName, routerReqsTotalName)
}

func TestPrometheusMetricRemoveEndpointForRecoveredService(t *testing.T) {
	promState = newPrometheusState()
	promRegistry = prometheus.NewRegistry()
	t.Cleanup(promState.reset)

	prometheusRegistry := RegisterPrometheus(t.Context(), &types.Prometheus{AddServicesLabels: true})
	defer promRegistry.Unregister(promState)

	conf1 := dynamic.Configuration{
		HTTP: th.BuildConfiguration(
			th.WithLoadBalancerServices(
				th.WithService("service1", th.WithServers(th.WithServer("http://localhost:9000"))),
			),
		),
	}

	conf2 := dynamic.Configuration{
		HTTP: th.BuildConfiguration(),
	}

	conf3 := dynamic.Configuration{
		HTTP: th.BuildConfiguration(
			th.WithLoadBalancerServices(
				th.WithService("service1", th.WithServers(th.WithServer("http://localhost:9001"))),
			),
		),
	}

	OnConfigurationUpdate(conf1, []string{})
	OnConfigurationUpdate(conf2, []string{})
	OnConfigurationUpdate(conf3, []string{})

	prometheusRegistry.
		ServiceServerUpGauge().
		With("service", "service1", "url", "http://localhost:9000").
		Add(1)

	assertMetricsExist(t, mustScrape(), serviceServerUpName)
	assertMetricsAbsent(t, mustScrape(), serviceServerUpName)
}

func TestPrometheusRemovedMetricsReset(t *testing.T) {
	t.Cleanup(promState.reset)

	prometheusRegistry := RegisterPrometheus(t.Context(), &types.Prometheus{AddEntryPointsLabels: true, AddServicesLabels: true})
	defer promRegistry.Unregister(promState)

	conf1 := dynamic.Configuration{
		HTTP: th.BuildConfiguration(
			th.WithLoadBalancerServices(th.WithService("service",
				th.WithServers(th.WithServer("http://localhost:9000"))),
			),
		),
	}
	OnConfigurationUpdate(conf1, []string{"entrypoint1", "entrypoint2"})
	OnConfigurationUpdate(dynamic.Configuration{}, nil)

	labelNamesValues := []string{
		"service", "service",
		"code", strconv.Itoa(http.StatusOK),
		"method", http.MethodGet,
		"protocol", "http",
	}
	prometheusRegistry.
		ServiceReqsCounter().
		With(nil, labelNamesValues...).
		Add(3)

	delayForTrackingCompletion()

	metricsFamilies := mustScrape()
	assertCounterValue(t, 3, findMetricFamily(serviceReqsTotalName, metricsFamilies), labelNamesValues...)

	// There is no dynamic configuration and so this metric will be deleted
	// after the first scrape.
	assertMetricsAbsent(t, mustScrape(), serviceReqsTotalName)

	prometheusRegistry.
		ServiceReqsCounter().
		With(nil, labelNamesValues...).
		Add(1)

	delayForTrackingCompletion()

	metricsFamilies = mustScrape()
	assertCounterValue(t, 1, findMetricFamily(serviceReqsTotalName, metricsFamilies), labelNamesValues...)
}

// reset is a utility method for unit testing.
// It should be called after each test run that changes promState internally
// in order to avoid dependencies between unit tests.
func (ps *prometheusState) reset() {
	ps.dynamicConfig = newDynamicConfig()
	ps.vectors = nil
	ps.deletedEP = nil
	ps.deletedRouters = nil
	ps.deletedServices = nil
	ps.deletedURLs = make(map[string][]string)
}

// Tracking and gathering the metrics happens concurrently.
// In practice this is no problem, because in case a tracked metric would miss the current scrape,
// it would just be there in the next one.
// That we can test reliably the tracking of all metrics here,
// we sleep for a short amount of time,
// to make sure the metric will be present in the next scrape.
func delayForTrackingCompletion() {
	time.Sleep(250 * time.Millisecond)
}

func mustScrape() []*dto.MetricFamily {
	families, err := promRegistry.Gather()
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

	for _, metric := range family.GetMetric() {
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
	for _, labelPair := range metric.GetLabel() {
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
	if metric.GetCounter() == nil {
		t.Errorf("metric %s must be a counter", family.GetName())
		return
	}

	if cv := metric.GetCounter().GetValue(); cv != want {
		t.Errorf("metric %s has value %v, want %v", family.GetName(), cv, want)
	}
}

func buildCounterAssert(t *testing.T, metricName string, expectedValue int) func(family *dto.MetricFamily) {
	t.Helper()

	return func(family *dto.MetricFamily) {
		if cv := int(family.GetMetric()[0].GetCounter().GetValue()); cv != expectedValue {
			t.Errorf("metric %s has value %d, want %d", metricName, cv, expectedValue)
		}
	}
}

func buildGreaterThanCounterAssert(t *testing.T, metricName string, expectedMinValue int) func(family *dto.MetricFamily) {
	t.Helper()

	return func(family *dto.MetricFamily) {
		if cv := int(family.GetMetric()[0].GetCounter().GetValue()); cv < expectedMinValue {
			t.Errorf("metric %s has value %d, want at least %d", metricName, cv, expectedMinValue)
		}
	}
}

func buildHistogramAssert(t *testing.T, metricName string, expectedSampleCount int) func(family *dto.MetricFamily) {
	t.Helper()

	return func(family *dto.MetricFamily) {
		if sc := int(family.GetMetric()[0].GetHistogram().GetSampleCount()); sc != expectedSampleCount {
			t.Errorf("metric %s has sample count value %d, want %d", metricName, sc, expectedSampleCount)
		}
	}
}

func buildGaugeAssert(t *testing.T, metricName string, expectedValue int) func(family *dto.MetricFamily) {
	t.Helper()

	return func(family *dto.MetricFamily) {
		if gv := int(family.GetMetric()[0].GetGauge().GetValue()); gv != expectedValue {
			t.Errorf("metric %s has value %d, want %d", metricName, gv, expectedValue)
		}
	}
}

func buildTimestampAssert(t *testing.T, metricName string) func(family *dto.MetricFamily) {
	t.Helper()

	return func(family *dto.MetricFamily) {
		if ts := time.Unix(int64(family.GetMetric()[0].GetGauge().GetValue()), 0); time.Since(ts) > time.Minute {
			t.Errorf("metric %s has wrong timestamp %v", metricName, ts)
		}
	}
}
