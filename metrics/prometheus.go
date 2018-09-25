package metrics

import (
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/containous/mux"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/go-kit/kit/metrics"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// MetricNamePrefix prefix of all metric names
	MetricNamePrefix = "traefik_"

	// server meta information
	metricConfigPrefix             = MetricNamePrefix + "config_"
	configReloadsTotalName         = metricConfigPrefix + "reloads_total"
	configReloadsFailuresTotalName = metricConfigPrefix + "reloads_failure_total"
	configLastReloadSuccessName    = metricConfigPrefix + "last_reload_success"
	configLastReloadFailureName    = metricConfigPrefix + "last_reload_failure"

	// entrypoint
	metricEntryPointPrefix    = MetricNamePrefix + "entrypoint_"
	entrypointReqsTotalName   = metricEntryPointPrefix + "requests_total"
	entrypointReqDurationName = metricEntryPointPrefix + "request_duration_seconds"
	entrypointOpenConnsName   = metricEntryPointPrefix + "open_connections"

	// backend level.

	// MetricBackendPrefix prefix of all backend metric names
	MetricBackendPrefix     = MetricNamePrefix + "backend_"
	backendReqsTotalName    = MetricBackendPrefix + "requests_total"
	backendReqDurationName  = MetricBackendPrefix + "request_duration_seconds"
	backendOpenConnsName    = MetricBackendPrefix + "open_connections"
	backendRetriesTotalName = MetricBackendPrefix + "retries_total"
	backendServerUpName     = MetricBackendPrefix + "server_up"
)

// promState holds all metric state internally and acts as the only Collector we register for Prometheus.
//
// This enables control to remove metrics that belong to outdated configuration.
// As an example why this is required, consider Traefik learns about a new service.
// It populates the 'traefik_server_backend_up' metric for it with a value of 1 (alive).
// When the backend is undeployed now the metric is still there in the client library
// and will be returned on the metrics endpoint until Traefik would be restarted.
//
// To solve this problem promState keeps track of Traefik's dynamic configuration.
// Metrics that "belong" to a dynamic configuration part like backends or entrypoints
// are removed after they were scraped at least once when the corresponding object
// doesn't exist anymore.
var promState = newPrometheusState()

// PrometheusHandler exposes Prometheus routes.
type PrometheusHandler struct{}

// AddRoutes adds Prometheus routes on a router.
func (h PrometheusHandler) AddRoutes(router *mux.Router) {
	router.Methods(http.MethodGet).Path("/metrics").Handler(promhttp.Handler())
}

// RegisterPrometheus registers all Prometheus metrics.
// It must be called only once and failing to register the metrics will lead to a panic.
func RegisterPrometheus(config *types.Prometheus) Registry {
	standardRegistry := initStandardRegistry(config)

	if !registerPromState() {
		return nil
	}

	return standardRegistry
}

func initStandardRegistry(config *types.Prometheus) Registry {
	buckets := []float64{0.1, 0.3, 1.2, 5.0}
	if config.Buckets != nil {
		buckets = config.Buckets
	}

	safe.Go(func() {
		promState.ListenValueUpdates()
	})

	configReloads := newCounterFrom(promState.collectors, stdprometheus.CounterOpts{
		Name: configReloadsTotalName,
		Help: "Config reloads",
	}, []string{})
	configReloadsFailures := newCounterFrom(promState.collectors, stdprometheus.CounterOpts{
		Name: configReloadsFailuresTotalName,
		Help: "Config failure reloads",
	}, []string{})
	lastConfigReloadSuccess := newGaugeFrom(promState.collectors, stdprometheus.GaugeOpts{
		Name: configLastReloadSuccessName,
		Help: "Last config reload success",
	}, []string{})
	lastConfigReloadFailure := newGaugeFrom(promState.collectors, stdprometheus.GaugeOpts{
		Name: configLastReloadFailureName,
		Help: "Last config reload failure",
	}, []string{})

	entrypointReqs := newCounterFrom(promState.collectors, stdprometheus.CounterOpts{
		Name: entrypointReqsTotalName,
		Help: "How many HTTP requests processed on an entrypoint, partitioned by status code, protocol, and method.",
	}, []string{"code", "method", "protocol", "entrypoint"})
	entrypointReqDurations := newHistogramFrom(promState.collectors, stdprometheus.HistogramOpts{
		Name:    entrypointReqDurationName,
		Help:    "How long it took to process the request on an entrypoint, partitioned by status code, protocol, and method.",
		Buckets: buckets,
	}, []string{"code", "method", "protocol", "entrypoint"})
	entrypointOpenConns := newGaugeFrom(promState.collectors, stdprometheus.GaugeOpts{
		Name: entrypointOpenConnsName,
		Help: "How many open connections exist on an entrypoint, partitioned by method and protocol.",
	}, []string{"method", "protocol", "entrypoint"})

	backendReqs := newCounterFrom(promState.collectors, stdprometheus.CounterOpts{
		Name: backendReqsTotalName,
		Help: "How many HTTP requests processed on a backend, partitioned by status code, protocol, and method.",
	}, []string{"code", "method", "protocol", "backend"})
	backendReqDurations := newHistogramFrom(promState.collectors, stdprometheus.HistogramOpts{
		Name:    backendReqDurationName,
		Help:    "How long it took to process the request on a backend, partitioned by status code, protocol, and method.",
		Buckets: buckets,
	}, []string{"code", "method", "protocol", "backend"})
	backendOpenConns := newGaugeFrom(promState.collectors, stdprometheus.GaugeOpts{
		Name: backendOpenConnsName,
		Help: "How many open connections exist on a backend, partitioned by method and protocol.",
	}, []string{"method", "protocol", "backend"})
	backendRetries := newCounterFrom(promState.collectors, stdprometheus.CounterOpts{
		Name: backendRetriesTotalName,
		Help: "How many request retries happened on a backend.",
	}, []string{"backend"})
	backendServerUp := newGaugeFrom(promState.collectors, stdprometheus.GaugeOpts{
		Name: backendServerUpName,
		Help: "Backend server is up, described by gauge value of 0 or 1.",
	}, []string{"backend", "url"})

	promState.describers = []func(chan<- *stdprometheus.Desc){
		configReloads.cv.Describe,
		configReloadsFailures.cv.Describe,
		lastConfigReloadSuccess.gv.Describe,
		lastConfigReloadFailure.gv.Describe,
		entrypointReqs.cv.Describe,
		entrypointReqDurations.hv.Describe,
		entrypointOpenConns.gv.Describe,
		backendReqs.cv.Describe,
		backendReqDurations.hv.Describe,
		backendOpenConns.gv.Describe,
		backendRetries.cv.Describe,
		backendServerUp.gv.Describe,
	}

	return &standardRegistry{
		enabled:                        true,
		configReloadsCounter:           configReloads,
		configReloadsFailureCounter:    configReloadsFailures,
		lastConfigReloadSuccessGauge:   lastConfigReloadSuccess,
		lastConfigReloadFailureGauge:   lastConfigReloadFailure,
		entrypointReqsCounter:          entrypointReqs,
		entrypointReqDurationHistogram: entrypointReqDurations,
		entrypointOpenConnsGauge:       entrypointOpenConns,
		backendReqsCounter:             backendReqs,
		backendReqDurationHistogram:    backendReqDurations,
		backendOpenConnsGauge:          backendOpenConns,
		backendRetriesCounter:          backendRetries,
		backendServerUpGauge:           backendServerUp,
	}
}

func registerPromState() bool {
	if err := stdprometheus.Register(promState); err != nil {
		if _, ok := err.(stdprometheus.AlreadyRegisteredError); !ok {
			log.Errorf("Unable to register Traefik to Prometheus: %v", err)
			return false
		}
		log.Debug("Prometheus collector already registered.")
	}
	return true
}

// OnConfigurationUpdate receives the current configuration from Traefik.
// It then converts the configuration to the optimized package internal format
// and sets it to the promState.
func OnConfigurationUpdate(configurations types.Configurations) {
	dynamicConfig := newDynamicConfig()

	for _, config := range configurations {
		for _, frontend := range config.Frontends {
			for _, entrypointName := range frontend.EntryPoints {
				dynamicConfig.entrypoints[entrypointName] = true
			}
		}

		for backendName, backend := range config.Backends {
			dynamicConfig.backends[backendName] = make(map[string]bool)
			for _, server := range backend.Servers {
				dynamicConfig.backends[backendName][server.URL] = true
			}
		}
	}

	promState.SetDynamicConfig(dynamicConfig)
}

func newPrometheusState() *prometheusState {
	return &prometheusState{
		collectors:    make(chan *collector),
		dynamicConfig: newDynamicConfig(),
		state:         make(map[string]*collector),
	}
}

type prometheusState struct {
	collectors chan *collector
	describers []func(ch chan<- *stdprometheus.Desc)

	mtx           sync.Mutex
	dynamicConfig *dynamicConfig
	state         map[string]*collector
}

// reset is a utility method for unit testing. It should be called after each
// test run that changes promState internally in order to avoid dependencies
// between unit tests.
func (ps *prometheusState) reset() {
	ps.collectors = make(chan *collector)
	ps.describers = []func(ch chan<- *stdprometheus.Desc){}
	ps.dynamicConfig = newDynamicConfig()
	ps.state = make(map[string]*collector)
}

func (ps *prometheusState) SetDynamicConfig(dynamicConfig *dynamicConfig) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	ps.dynamicConfig = dynamicConfig
}

func (ps *prometheusState) ListenValueUpdates() {
	for collector := range ps.collectors {
		ps.mtx.Lock()
		ps.state[collector.id] = collector
		ps.mtx.Unlock()
	}
}

// Describe implements prometheus.Collector and simply calls
// the registered describer functions.
func (ps *prometheusState) Describe(ch chan<- *stdprometheus.Desc) {
	for _, desc := range ps.describers {
		desc(ch)
	}
}

// Collect implements prometheus.Collector. It calls the Collect
// method of all metrics it received on the collectors channel.
// It's also responsible to remove metrics that belong to an outdated configuration.
// The removal happens only after their Collect method was called to ensure that
// also those metrics will be exported on the current scrape.
func (ps *prometheusState) Collect(ch chan<- stdprometheus.Metric) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	var outdatedKeys []string
	for key, cs := range ps.state {
		cs.collector.Collect(ch)

		if ps.isOutdated(cs) {
			outdatedKeys = append(outdatedKeys, key)
		}
	}

	for _, key := range outdatedKeys {
		ps.state[key].delete()
		delete(ps.state, key)
	}
}

// isOutdated checks whether the passed collector has labels that mark
// it as belonging to an outdated configuration of Traefik.
func (ps *prometheusState) isOutdated(collector *collector) bool {
	labels := collector.labels

	if entrypointName, ok := labels["entrypoint"]; ok && !ps.dynamicConfig.hasEntrypoint(entrypointName) {
		return true
	}

	if backendName, ok := labels["backend"]; ok {
		if !ps.dynamicConfig.hasBackend(backendName) {
			return true
		}
		if url, ok := labels["url"]; ok && !ps.dynamicConfig.hasServerURL(backendName, url) {
			return true
		}
	}

	return false
}

func newDynamicConfig() *dynamicConfig {
	return &dynamicConfig{
		entrypoints: make(map[string]bool),
		backends:    make(map[string]map[string]bool),
	}
}

// dynamicConfig holds the current configuration for entrypoints, backends,
// and server URLs in an optimized way to check for existence. This provides
// a performant way to check whether the collected metrics belong to the
// current configuration or to an outdated one.
type dynamicConfig struct {
	entrypoints map[string]bool
	backends    map[string]map[string]bool
}

func (d *dynamicConfig) hasEntrypoint(entrypointName string) bool {
	_, ok := d.entrypoints[entrypointName]
	return ok
}

func (d *dynamicConfig) hasBackend(backendName string) bool {
	_, ok := d.backends[backendName]
	return ok
}

func (d *dynamicConfig) hasServerURL(backendName, serverURL string) bool {
	if backend, hasBackend := d.backends[backendName]; hasBackend {
		_, ok := backend[serverURL]
		return ok
	}
	return false
}

func newCollector(metricName string, labels stdprometheus.Labels, c stdprometheus.Collector, delete func()) *collector {
	return &collector{
		id:        buildMetricID(metricName, labels),
		labels:    labels,
		collector: c,
		delete:    delete,
	}
}

// collector wraps a Collector object from the Prometheus client library.
// It adds information on how many generations this metric should be present
// in the /metrics output, relatived to the time it was last tracked.
type collector struct {
	id        string
	labels    stdprometheus.Labels
	collector stdprometheus.Collector
	delete    func()
}

func buildMetricID(metricName string, labels stdprometheus.Labels) string {
	var labelNamesValues []string
	for name, value := range labels {
		labelNamesValues = append(labelNamesValues, name, value)
	}
	sort.Strings(labelNamesValues)
	return metricName + ":" + strings.Join(labelNamesValues, "|")
}

func newCounterFrom(collectors chan<- *collector, opts stdprometheus.CounterOpts, labelNames []string) *counter {
	cv := stdprometheus.NewCounterVec(opts, labelNames)
	c := &counter{
		name:       opts.Name,
		cv:         cv,
		collectors: collectors,
	}
	if len(labelNames) == 0 {
		c.Add(0)
	}
	return c
}

type counter struct {
	name             string
	cv               *stdprometheus.CounterVec
	labelNamesValues labelNamesValues
	collectors       chan<- *collector
}

func (c *counter) With(labelValues ...string) metrics.Counter {
	return &counter{
		name:             c.name,
		cv:               c.cv,
		labelNamesValues: c.labelNamesValues.With(labelValues...),
		collectors:       c.collectors,
	}
}

func (c *counter) Add(delta float64) {
	labels := c.labelNamesValues.ToLabels()
	collector := c.cv.With(labels)
	collector.Add(delta)
	c.collectors <- newCollector(c.name, labels, collector, func() {
		c.cv.Delete(labels)
	})
}

func (c *counter) Describe(ch chan<- *stdprometheus.Desc) {
	c.cv.Describe(ch)
}

func newGaugeFrom(collectors chan<- *collector, opts stdprometheus.GaugeOpts, labelNames []string) *gauge {
	gv := stdprometheus.NewGaugeVec(opts, labelNames)
	g := &gauge{
		name:       opts.Name,
		gv:         gv,
		collectors: collectors,
	}
	if len(labelNames) == 0 {
		g.Set(0)
	}
	return g
}

type gauge struct {
	name             string
	gv               *stdprometheus.GaugeVec
	labelNamesValues labelNamesValues
	collectors       chan<- *collector
}

func (g *gauge) With(labelValues ...string) metrics.Gauge {
	return &gauge{
		name:             g.name,
		gv:               g.gv,
		labelNamesValues: g.labelNamesValues.With(labelValues...),
		collectors:       g.collectors,
	}
}

func (g *gauge) Add(delta float64) {
	labels := g.labelNamesValues.ToLabels()
	collector := g.gv.With(labels)
	collector.Add(delta)
	g.collectors <- newCollector(g.name, labels, collector, func() {
		g.gv.Delete(labels)
	})
}

func (g *gauge) Set(value float64) {
	labels := g.labelNamesValues.ToLabels()
	collector := g.gv.With(labels)
	collector.Set(value)
	g.collectors <- newCollector(g.name, labels, collector, func() {
		g.gv.Delete(labels)
	})
}

func (g *gauge) Describe(ch chan<- *stdprometheus.Desc) {
	g.gv.Describe(ch)
}

func newHistogramFrom(collectors chan<- *collector, opts stdprometheus.HistogramOpts, labelNames []string) *histogram {
	hv := stdprometheus.NewHistogramVec(opts, labelNames)
	return &histogram{
		name:       opts.Name,
		hv:         hv,
		collectors: collectors,
	}
}

type histogram struct {
	name             string
	hv               *stdprometheus.HistogramVec
	labelNamesValues labelNamesValues
	collectors       chan<- *collector
}

func (h *histogram) With(labelValues ...string) metrics.Histogram {
	return &histogram{
		name:             h.name,
		hv:               h.hv,
		labelNamesValues: h.labelNamesValues.With(labelValues...),
		collectors:       h.collectors,
	}
}

func (h *histogram) Observe(value float64) {
	labels := h.labelNamesValues.ToLabels()
	collector := h.hv.With(labels)
	collector.Observe(value)
	h.collectors <- newCollector(h.name, labels, collector, func() {
		h.hv.Delete(labels)
	})
}

func (h *histogram) Describe(ch chan<- *stdprometheus.Desc) {
	h.hv.Describe(ch)
}

// labelNamesValues is a type alias that provides validation on its With method.
// Metrics may include it as a member to help them satisfy With semantics and
// save some code duplication.
type labelNamesValues []string

// With validates the input, and returns a new aggregate labelNamesValues.
func (lvs labelNamesValues) With(labelValues ...string) labelNamesValues {
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, "unknown")
	}
	return append(lvs, labelValues...)
}

// ToLabels is a convenience method to convert a labelNamesValues
// to the native prometheus.Labels.
func (lvs labelNamesValues) ToLabels() stdprometheus.Labels {
	labels := stdprometheus.Labels{}
	for i := 0; i < len(lvs); i += 2 {
		labels[lvs[i]] = lvs[i+1]
	}
	return labels
}
