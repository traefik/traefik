package metrics

import (
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/containous/mux"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/go-kit/kit/metrics"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	metricNamePrefix = "traefik_"

	// server meta information
	configReloadsTotalName         = metricNamePrefix + "config_reloads_total"
	configReloadsFailuresTotalName = metricNamePrefix + "config_reloads_failure_total"
	configLastReloadSuccessName    = metricNamePrefix + "config_last_reload_success"
	configLastReloadFailureName    = metricNamePrefix + "config_last_reload_failure"

	// entrypoint
	entrypointReqsTotalName   = metricNamePrefix + "entrypoint_requests_total"
	entrypointReqDurationName = metricNamePrefix + "entrypoint_request_duration_seconds"
	entrypointOpenConnsName   = metricNamePrefix + "entrypoint_open_connections"

	// backend level
	backendReqsTotalName    = metricNamePrefix + "backend_requests_total"
	backendReqDurationName  = metricNamePrefix + "backend_request_duration_seconds"
	backendOpenConnsName    = metricNamePrefix + "backend_open_connections"
	backendRetriesTotalName = metricNamePrefix + "backend_retries_total"
	backendServerUpName     = metricNamePrefix + "backend_server_up"
)

const (
	// generationAgeForever indicates that a metric never gets outdated.
	generationAgeForever = 0
	// generationAgeDefault is the default age of three generations.
	generationAgeDefault = 3
)

// promState holds all metric state internally and acts as the only Collector we register for Prometheus.
//
// This enables control to remove metrics that belong to outdated configuration.
// As an example why this is required, consider Traefik learns about a new service.
// It populates the 'traefik_server_backend_up' metric for it with a value of 1 (alive).
// When the backend is undeployed now the metric is still there in the client library
// and will be until Traefik would be restarted.
//
// To solve this problem promState keeps track of configuration generations.
// Every time a new configuration is loaded, the generation is increased by one.
// Metrics that "belong" to a dynamic configuration part of Traefik (e.g. backend, entrypoint)
// are removed, given they were tracked more than 3 generations ago.
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
	stdprometheus.MustRegister(promState)

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

// OnConfigurationUpdate increases the current generation of the prometheus state.
func OnConfigurationUpdate() {
	promState.IncGeneration()
}

func newPrometheusState() *prometheusState {
	collectors := make(chan *collector)
	state := make(map[string]*collector)

	return &prometheusState{
		collectors: collectors,
		state:      state,
	}
}

type prometheusState struct {
	currentGeneration int
	collectors        chan *collector
	describers        []func(ch chan<- *stdprometheus.Desc)

	mtx   sync.Mutex
	state map[string]*collector
}

func (ps *prometheusState) IncGeneration() {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	ps.currentGeneration++
}

func (ps *prometheusState) ListenValueUpdates() {
	for collector := range ps.collectors {
		ps.mtx.Lock()
		collector.lastTrackedGeneration = ps.currentGeneration
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
// It's also responsible to remove metrics that were tracked
// at least three generations ago. Those metrics are cleaned up
// after the Collect of them were called.
func (ps *prometheusState) Collect(ch chan<- stdprometheus.Metric) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	outdatedKeys := []string{}
	for key, cs := range ps.state {
		cs.collector.Collect(ch)

		if cs.maxAge == generationAgeForever {
			continue
		}
		if ps.currentGeneration-cs.lastTrackedGeneration >= cs.maxAge {
			outdatedKeys = append(outdatedKeys, key)
		}
	}

	for _, key := range outdatedKeys {
		delete(ps.state, key)
	}
}

func newCollector(metricName string, lnvs labelNamesValues, c stdprometheus.Collector) *collector {
	maxAge := generationAgeDefault

	// metrics without labels should never become outdated
	if len(lnvs) == 0 {
		maxAge = generationAgeForever
	}

	return &collector{
		id:        buildMetricID(metricName, lnvs),
		maxAge:    maxAge,
		collector: c,
	}
}

// collector wraps a Collector object from the Prometheus client library.
// It adds information on how many generations this metric should be present
// in the /metrics output, relatived to the time it was last tracked.
type collector struct {
	id                    string
	collector             stdprometheus.Collector
	lastTrackedGeneration int
	maxAge                int
}

func buildMetricID(metricName string, lnvs labelNamesValues) string {
	newLnvs := append([]string{}, lnvs...)
	sort.Strings(newLnvs)
	return metricName + ":" + strings.Join(newLnvs, "|")
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
	collector := c.cv.With(c.labelNamesValues.ToLabels())
	collector.Add(delta)
	c.collectors <- newCollector(c.name, c.labelNamesValues, collector)
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
	collector := g.gv.With(g.labelNamesValues.ToLabels())
	collector.Add(delta)
	g.collectors <- newCollector(g.name, g.labelNamesValues, collector)
}

func (g *gauge) Set(value float64) {
	collector := g.gv.With(g.labelNamesValues.ToLabels())
	collector.Set(value)
	g.collectors <- newCollector(g.name, g.labelNamesValues, collector)
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
	collector := h.hv.With(h.labelNamesValues.ToLabels())
	collector.Observe(value)
	h.collectors <- newCollector(h.name, h.labelNamesValues, collector)
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
