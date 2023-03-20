package metrics

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/go-kit/kit/metrics"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/types"
)

const (
	// MetricNamePrefix prefix of all metric names.
	MetricNamePrefix = "traefik_"

	// server meta information.
	metricConfigPrefix             = MetricNamePrefix + "config_"
	configReloadsTotalName         = metricConfigPrefix + "reloads_total"
	configReloadsFailuresTotalName = metricConfigPrefix + "reloads_failure_total"
	configLastReloadSuccessName    = metricConfigPrefix + "last_reload_success"
	configLastReloadFailureName    = metricConfigPrefix + "last_reload_failure"

	// TLS.
	metricsTLSPrefix          = MetricNamePrefix + "tls_"
	tlsCertsNotAfterTimestamp = metricsTLSPrefix + "certs_not_after"

	// entry point.
	metricEntryPointPrefix        = MetricNamePrefix + "entrypoint_"
	entryPointReqsTotalName       = metricEntryPointPrefix + "requests_total"
	entryPointReqsTLSTotalName    = metricEntryPointPrefix + "requests_tls_total"
	entryPointReqDurationName     = metricEntryPointPrefix + "request_duration_seconds"
	entryPointOpenConnsName       = metricEntryPointPrefix + "open_connections"
	entryPointReqsBytesTotalName  = metricEntryPointPrefix + "requests_bytes_total"
	entryPointRespsBytesTotalName = metricEntryPointPrefix + "responses_bytes_total"

	// router level.
	metricRouterPrefix        = MetricNamePrefix + "router_"
	routerReqsTotalName       = metricRouterPrefix + "requests_total"
	routerReqsTLSTotalName    = metricRouterPrefix + "requests_tls_total"
	routerReqDurationName     = metricRouterPrefix + "request_duration_seconds"
	routerOpenConnsName       = metricRouterPrefix + "open_connections"
	routerReqsBytesTotalName  = metricRouterPrefix + "requests_bytes_total"
	routerRespsBytesTotalName = metricRouterPrefix + "responses_bytes_total"

	// service level.
	metricServicePrefix        = MetricNamePrefix + "service_"
	serviceReqsTotalName       = metricServicePrefix + "requests_total"
	serviceReqsTLSTotalName    = metricServicePrefix + "requests_tls_total"
	serviceReqDurationName     = metricServicePrefix + "request_duration_seconds"
	serviceOpenConnsName       = metricServicePrefix + "open_connections"
	serviceRetriesTotalName    = metricServicePrefix + "retries_total"
	serviceServerUpName        = metricServicePrefix + "server_up"
	serviceReqsBytesTotalName  = metricServicePrefix + "requests_bytes_total"
	serviceRespsBytesTotalName = metricServicePrefix + "responses_bytes_total"
)

// promState holds all metric state internally and acts as the only Collector we register for Prometheus.
//
// This enables control to remove metrics that belong to outdated configuration.
// As an example why this is required, consider Traefik learns about a new service.
// It populates the 'traefik_server_service_up' metric for it with a value of 1 (alive).
// When the service is undeployed now the metric is still there in the client library
// and will be returned on the metrics endpoint until Traefik would be restarted.
//
// To solve this problem promState keeps track of Traefik's dynamic configuration.
// Metrics that "belong" to a dynamic configuration part like services or entryPoints
// are removed after they were scraped at least once when the corresponding object
// doesn't exist anymore.
var promState = newPrometheusState()

var promRegistry = stdprometheus.NewRegistry()

// PrometheusHandler exposes Prometheus routes.
func PrometheusHandler() http.Handler {
	return promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{})
}

// RegisterPrometheus registers all Prometheus metrics.
// It must be called only once and failing to register the metrics will lead to a panic.
func RegisterPrometheus(ctx context.Context, config *types.Prometheus) Registry {
	standardRegistry := initStandardRegistry(config)

	if err := promRegistry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		var arErr stdprometheus.AlreadyRegisteredError
		if !errors.As(err, &arErr) {
			log.FromContext(ctx).Warn("ProcessCollector is already registered")
		}
	}

	if err := promRegistry.Register(collectors.NewGoCollector()); err != nil {
		var arErr stdprometheus.AlreadyRegisteredError
		if !errors.As(err, &arErr) {
			log.FromContext(ctx).Warn("GoCollector is already registered")
		}
	}

	if !registerPromState(ctx) {
		return nil
	}

	return standardRegistry
}

func initStandardRegistry(config *types.Prometheus) Registry {
	buckets := []float64{0.1, 0.3, 1.2, 5.0}
	if config.Buckets != nil {
		buckets = config.Buckets
	}

	configReloads := newCounterFrom(stdprometheus.CounterOpts{
		Name: configReloadsTotalName,
		Help: "Config reloads",
	}, []string{})
	configReloadsFailures := newCounterFrom(stdprometheus.CounterOpts{
		Name: configReloadsFailuresTotalName,
		Help: "Config failure reloads",
	}, []string{})
	lastConfigReloadSuccess := newGaugeFrom(stdprometheus.GaugeOpts{
		Name: configLastReloadSuccessName,
		Help: "Last config reload success",
	}, []string{})
	lastConfigReloadFailure := newGaugeFrom(stdprometheus.GaugeOpts{
		Name: configLastReloadFailureName,
		Help: "Last config reload failure",
	}, []string{})
	tlsCertsNotAfterTimestamp := newGaugeFrom(stdprometheus.GaugeOpts{
		Name: tlsCertsNotAfterTimestamp,
		Help: "Certificate expiration timestamp",
	}, []string{"cn", "serial", "sans"})

	promState.vectors = []vector{
		configReloads.cv,
		configReloadsFailures.cv,
		lastConfigReloadSuccess.gv,
		lastConfigReloadFailure.gv,
		tlsCertsNotAfterTimestamp.gv,
	}

	reg := &standardRegistry{
		epEnabled:                      config.AddEntryPointsLabels,
		routerEnabled:                  config.AddRoutersLabels,
		svcEnabled:                     config.AddServicesLabels,
		configReloadsCounter:           configReloads,
		configReloadsFailureCounter:    configReloadsFailures,
		lastConfigReloadSuccessGauge:   lastConfigReloadSuccess,
		lastConfigReloadFailureGauge:   lastConfigReloadFailure,
		tlsCertsNotAfterTimestampGauge: tlsCertsNotAfterTimestamp,
	}

	if config.AddEntryPointsLabels {
		entryPointReqs := newCounterWithHeadersFrom(stdprometheus.CounterOpts{
			Name: entryPointReqsTotalName,
			Help: "How many HTTP requests processed on an entrypoint, partitioned by status code, protocol, and method.",
		}, config.HeaderLabels, []string{"code", "method", "protocol", "entrypoint"})
		entryPointReqsTLS := newCounterFrom(stdprometheus.CounterOpts{
			Name: entryPointReqsTLSTotalName,
			Help: "How many HTTP requests with TLS processed on an entrypoint, partitioned by TLS Version and TLS cipher Used.",
		}, []string{"tls_version", "tls_cipher", "entrypoint"})
		entryPointReqDurations := newHistogramFrom(stdprometheus.HistogramOpts{
			Name:    entryPointReqDurationName,
			Help:    "How long it took to process the request on an entrypoint, partitioned by status code, protocol, and method.",
			Buckets: buckets,
		}, []string{"code", "method", "protocol", "entrypoint"})
		entryPointOpenConns := newGaugeFrom(stdprometheus.GaugeOpts{
			Name: entryPointOpenConnsName,
			Help: "How many open connections exist on an entrypoint, partitioned by method and protocol.",
		}, []string{"method", "protocol", "entrypoint"})
		entryPointReqsBytesTotal := newCounterFrom(stdprometheus.CounterOpts{
			Name: entryPointReqsBytesTotalName,
			Help: "The total size of requests in bytes handled by an entrypoint, partitioned by status code, protocol, and method.",
		}, []string{"code", "method", "protocol", "entrypoint"})
		entryPointRespsBytesTotal := newCounterFrom(stdprometheus.CounterOpts{
			Name: entryPointRespsBytesTotalName,
			Help: "The total size of responses in bytes handled by an entrypoint, partitioned by status code, protocol, and method.",
		}, []string{"code", "method", "protocol", "entrypoint"})

		promState.vectors = append(promState.vectors,
			entryPointReqs.cv,
			entryPointReqsTLS.cv,
			entryPointReqDurations.hv,
			entryPointOpenConns.gv,
			entryPointReqsBytesTotal.cv,
			entryPointRespsBytesTotal.cv,
		)

		reg.entryPointReqsCounter = entryPointReqs
		reg.entryPointReqsTLSCounter = entryPointReqsTLS
		reg.entryPointReqDurationHistogram, _ = NewHistogramWithScale(entryPointReqDurations, time.Second)
		reg.entryPointOpenConnsGauge = entryPointOpenConns
		reg.entryPointReqsBytesCounter = entryPointReqsBytesTotal
		reg.entryPointRespsBytesCounter = entryPointRespsBytesTotal
	}

	if config.AddRoutersLabels {
		routerReqs := newCounterWithHeadersFrom(stdprometheus.CounterOpts{
			Name: routerReqsTotalName,
			Help: "How many HTTP requests are processed on a router, partitioned by service, status code, protocol, and method.",
		}, config.HeaderLabels, []string{"code", "method", "protocol", "router", "service"})
		routerReqsTLS := newCounterFrom(stdprometheus.CounterOpts{
			Name: routerReqsTLSTotalName,
			Help: "How many HTTP requests with TLS are processed on a router, partitioned by service, TLS Version, and TLS cipher Used.",
		}, []string{"tls_version", "tls_cipher", "router", "service"})
		routerReqDurations := newHistogramFrom(stdprometheus.HistogramOpts{
			Name:    routerReqDurationName,
			Help:    "How long it took to process the request on a router, partitioned by service, status code, protocol, and method.",
			Buckets: buckets,
		}, []string{"code", "method", "protocol", "router", "service"})
		routerOpenConns := newGaugeFrom(stdprometheus.GaugeOpts{
			Name: routerOpenConnsName,
			Help: "How many open connections exist on a router, partitioned by service, method, and protocol.",
		}, []string{"method", "protocol", "router", "service"})
		routerReqsBytesTotal := newCounterFrom(stdprometheus.CounterOpts{
			Name: routerReqsBytesTotalName,
			Help: "The total size of requests in bytes handled by a router, partitioned by service, status code, protocol, and method.",
		}, []string{"code", "method", "protocol", "router", "service"})
		routerRespsBytesTotal := newCounterFrom(stdprometheus.CounterOpts{
			Name: routerRespsBytesTotalName,
			Help: "The total size of responses in bytes handled by a router, partitioned by service, status code, protocol, and method.",
		}, []string{"code", "method", "protocol", "router", "service"})

		promState.vectors = append(promState.vectors,
			routerReqs.cv,
			routerReqsTLS.cv,
			routerReqDurations.hv,
			routerOpenConns.gv,
			routerReqsBytesTotal.cv,
			routerRespsBytesTotal.cv,
		)
		reg.routerReqsCounter = routerReqs
		reg.routerReqsTLSCounter = routerReqsTLS
		reg.routerReqDurationHistogram, _ = NewHistogramWithScale(routerReqDurations, time.Second)
		reg.routerOpenConnsGauge = routerOpenConns
		reg.routerReqsBytesCounter = routerReqsBytesTotal
		reg.routerRespsBytesCounter = routerRespsBytesTotal
	}

	if config.AddServicesLabels {
		serviceReqs := newCounterWithHeadersFrom(stdprometheus.CounterOpts{
			Name: serviceReqsTotalName,
			Help: "How many HTTP requests processed on a service, partitioned by status code, protocol, and method.",
		}, config.HeaderLabels, []string{"code", "method", "protocol", "service"})
		serviceReqsTLS := newCounterFrom(stdprometheus.CounterOpts{
			Name: serviceReqsTLSTotalName,
			Help: "How many HTTP requests with TLS processed on a service, partitioned by TLS version and TLS cipher.",
		}, []string{"tls_version", "tls_cipher", "service"})
		serviceReqDurations := newHistogramFrom(stdprometheus.HistogramOpts{
			Name:    serviceReqDurationName,
			Help:    "How long it took to process the request on a service, partitioned by status code, protocol, and method.",
			Buckets: buckets,
		}, []string{"code", "method", "protocol", "service"})
		serviceOpenConns := newGaugeFrom(stdprometheus.GaugeOpts{
			Name: serviceOpenConnsName,
			Help: "How many open connections exist on a service, partitioned by method and protocol.",
		}, []string{"method", "protocol", "service"})
		serviceRetries := newCounterFrom(stdprometheus.CounterOpts{
			Name: serviceRetriesTotalName,
			Help: "How many request retries happened on a service.",
		}, []string{"service"})
		serviceServerUp := newGaugeFrom(stdprometheus.GaugeOpts{
			Name: serviceServerUpName,
			Help: "service server is up, described by gauge value of 0 or 1.",
		}, []string{"service", "url"})
		serviceReqsBytesTotal := newCounterFrom(stdprometheus.CounterOpts{
			Name: serviceReqsBytesTotalName,
			Help: "The total size of requests in bytes received by a service, partitioned by status code, protocol, and method.",
		}, []string{"code", "method", "protocol", "service"})
		serviceRespsBytesTotal := newCounterFrom(stdprometheus.CounterOpts{
			Name: serviceRespsBytesTotalName,
			Help: "The total size of responses in bytes returned by a service, partitioned by status code, protocol, and method.",
		}, []string{"code", "method", "protocol", "service"})

		promState.vectors = append(promState.vectors,
			serviceReqs.cv,
			serviceReqsTLS.cv,
			serviceReqDurations.hv,
			serviceOpenConns.gv,
			serviceRetries.cv,
			serviceServerUp.gv,
			serviceReqsBytesTotal.cv,
			serviceRespsBytesTotal.cv,
		)

		reg.serviceReqsCounter = serviceReqs
		reg.serviceReqsTLSCounter = serviceReqsTLS
		reg.serviceReqDurationHistogram, _ = NewHistogramWithScale(serviceReqDurations, time.Second)
		reg.serviceOpenConnsGauge = serviceOpenConns
		reg.serviceRetriesCounter = serviceRetries
		reg.serviceServerUpGauge = serviceServerUp
		reg.serviceReqsBytesCounter = serviceReqsBytesTotal
		reg.serviceRespsBytesCounter = serviceRespsBytesTotal
	}

	return reg
}

func registerPromState(ctx context.Context) bool {
	err := promRegistry.Register(promState)
	if err == nil {
		return true
	}

	logger := log.FromContext(ctx)

	var arErr stdprometheus.AlreadyRegisteredError
	if errors.As(err, &arErr) {
		logger.Debug("Prometheus collector already registered.")
		return true
	}

	logger.Errorf("Unable to register Traefik to Prometheus: %v", err)
	return false
}

// OnConfigurationUpdate receives the current configuration from Traefik.
// It then converts the configuration to the optimized package internal format
// and sets it to the promState.
func OnConfigurationUpdate(conf dynamic.Configuration, entryPoints []string) {
	dynCfg := newDynamicConfig()

	for _, value := range entryPoints {
		dynCfg.entryPoints[value] = true
	}

	if conf.HTTP == nil {
		promState.SetDynamicConfig(dynCfg)
		return
	}

	for name := range conf.HTTP.Routers {
		dynCfg.routers[name] = true
	}

	for serviceName, service := range conf.HTTP.Services {
		dynCfg.services[serviceName] = make(map[string]bool)
		if service.LoadBalancer != nil {
			for _, server := range service.LoadBalancer.Servers {
				dynCfg.services[serviceName][server.URL] = true
			}
		}
	}

	promState.SetDynamicConfig(dynCfg)
}

func newPrometheusState() *prometheusState {
	return &prometheusState{
		dynamicConfig: newDynamicConfig(),
		deletedURLs:   make(map[string][]string),
	}
}

type vector interface {
	stdprometheus.Collector
	DeletePartialMatch(labels stdprometheus.Labels) int
}

type prometheusState struct {
	vectors []vector

	mtx             sync.Mutex
	dynamicConfig   *dynamicConfig
	deletedEP       []string
	deletedRouters  []string
	deletedServices []string
	deletedURLs     map[string][]string
}

func (ps *prometheusState) SetDynamicConfig(dynamicConfig *dynamicConfig) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	for ep := range ps.dynamicConfig.entryPoints {
		if _, ok := dynamicConfig.entryPoints[ep]; !ok {
			ps.deletedEP = append(ps.deletedEP, ep)
		}
	}

	for router := range ps.dynamicConfig.routers {
		if _, ok := dynamicConfig.routers[router]; !ok {
			ps.deletedRouters = append(ps.deletedRouters, router)
		}
	}

	for service, serV := range ps.dynamicConfig.services {
		actualService, ok := dynamicConfig.services[service]
		if !ok {
			ps.deletedServices = append(ps.deletedServices, service)
		}
		for url := range serV {
			if _, ok := actualService[url]; !ok {
				ps.deletedURLs[service] = append(ps.deletedURLs[service], url)
			}
		}
	}

	ps.dynamicConfig = dynamicConfig
}

// Describe implements prometheus.Collector and simply calls
// the registered describer functions.
func (ps *prometheusState) Describe(ch chan<- *stdprometheus.Desc) {
	for _, v := range ps.vectors {
		v.Describe(ch)
	}
}

// Collect implements prometheus.Collector. It calls the Collect
// method of all metrics it received on the collectors channel.
// It's also responsible to remove metrics that belong to an outdated configuration.
// The removal happens only after their Collect method was called to ensure that
// also those metrics will be exported on the current scrape.
func (ps *prometheusState) Collect(ch chan<- stdprometheus.Metric) {
	for _, v := range ps.vectors {
		v.Collect(ch)
	}

	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	for _, ep := range ps.deletedEP {
		if !ps.dynamicConfig.hasEntryPoint(ep) {
			ps.DeletePartialMatch(map[string]string{"entrypoint": ep})
		}
	}

	for _, router := range ps.deletedRouters {
		if !ps.dynamicConfig.hasRouter(router) {
			ps.DeletePartialMatch(map[string]string{"router": router})
		}
	}

	for _, service := range ps.deletedServices {
		if !ps.dynamicConfig.hasService(service) {
			ps.DeletePartialMatch(map[string]string{"service": service})
		}
	}

	for service, urls := range ps.deletedURLs {
		for _, url := range urls {
			if !ps.dynamicConfig.hasServerURL(service, url) {
				ps.DeletePartialMatch(map[string]string{"service": service, "url": url})
			}
		}
	}

	ps.deletedEP = nil
	ps.deletedRouters = nil
	ps.deletedServices = nil
	ps.deletedURLs = make(map[string][]string)
}

// DeletePartialMatch deletes all metrics where the variable labels contain all of those passed in as labels.
// The order of the labels does not matter.
// It returns the number of metrics deleted.
func (ps *prometheusState) DeletePartialMatch(labels stdprometheus.Labels) int {
	var count int
	for _, elem := range ps.vectors {
		count += elem.DeletePartialMatch(labels)
	}
	return count
}

func newDynamicConfig() *dynamicConfig {
	return &dynamicConfig{
		entryPoints: make(map[string]bool),
		routers:     make(map[string]bool),
		services:    make(map[string]map[string]bool),
	}
}

// dynamicConfig holds the current configuration for entryPoints, services,
// and server URLs in an optimized way to check for existence. This provides
// a performant way to check whether the collected metrics belong to the
// current configuration or to an outdated one.
type dynamicConfig struct {
	entryPoints map[string]bool
	routers     map[string]bool
	services    map[string]map[string]bool
}

func (d *dynamicConfig) hasEntryPoint(entrypointName string) bool {
	_, ok := d.entryPoints[entrypointName]
	return ok
}

func (d *dynamicConfig) hasService(serviceName string) bool {
	_, ok := d.services[serviceName]
	return ok
}

func (d *dynamicConfig) hasRouter(routerName string) bool {
	_, ok := d.routers[routerName]
	return ok
}

func (d *dynamicConfig) hasServerURL(serviceName, serverURL string) bool {
	if service, hasService := d.services[serviceName]; hasService {
		_, ok := service[serverURL]
		return ok
	}
	return false
}

func newCounterWithHeadersFrom(opts stdprometheus.CounterOpts, headers map[string]string, labelNames []string) *counterWithHeaders {
	var headerLabels []string
	for k := range headers {
		headerLabels = append(headerLabels, k)
	}

	cv := stdprometheus.NewCounterVec(opts, append(labelNames, headerLabels...))
	c := &counterWithHeaders{
		name:    opts.Name,
		headers: headers,
		cv:      cv,
	}
	if len(labelNames) == 0 && len(headerLabels) == 0 {
		c.collector = cv.WithLabelValues()
		c.Add(0)
	}
	return c
}

type counterWithHeaders struct {
	name             string
	cv               *stdprometheus.CounterVec
	labelNamesValues labelNamesValues
	headers          map[string]string
	collector        stdprometheus.Counter
}

func (c *counterWithHeaders) With(headers http.Header, labelValues ...string) CounterWithHeaders {
	for headerLabel, headerKey := range c.headers {
		labelValues = append(labelValues, headerLabel, headers.Get(headerKey))
	}
	lnv := c.labelNamesValues.With(labelValues...)
	return &counterWithHeaders{
		name:             c.name,
		headers:          c.headers,
		cv:               c.cv,
		labelNamesValues: lnv,
		collector:        c.cv.With(lnv.ToLabels()),
	}
}

func (c *counterWithHeaders) Add(delta float64) {
	c.collector.Add(delta)
}

func (c *counterWithHeaders) Describe(ch chan<- *stdprometheus.Desc) {
	c.cv.Describe(ch)
}

func newCounterFrom(opts stdprometheus.CounterOpts, labelNames []string) *counter {
	cv := stdprometheus.NewCounterVec(opts, labelNames)
	c := &counter{
		name: opts.Name,
		cv:   cv,
	}
	if len(labelNames) == 0 {
		c.collector = cv.WithLabelValues()
		c.Add(0)
	}
	return c
}

type counter struct {
	name             string
	cv               *stdprometheus.CounterVec
	labelNamesValues labelNamesValues
	collector        stdprometheus.Counter
}

func (c *counter) With(labelValues ...string) metrics.Counter {
	lnv := c.labelNamesValues.With(labelValues...)
	return &counter{
		name:             c.name,
		cv:               c.cv,
		labelNamesValues: lnv,
		collector:        c.cv.With(lnv.ToLabels()),
	}
}

func (c *counter) Add(delta float64) {
	c.collector.Add(delta)
}

func (c *counter) Describe(ch chan<- *stdprometheus.Desc) {
	c.cv.Describe(ch)
}

func newGaugeFrom(opts stdprometheus.GaugeOpts, labelNames []string) *gauge {
	gv := stdprometheus.NewGaugeVec(opts, labelNames)
	g := &gauge{
		name: opts.Name,
		gv:   gv,
	}

	if len(labelNames) == 0 {
		g.collector = gv.WithLabelValues()
		g.Set(0)
	}
	return g
}

type gauge struct {
	name             string
	gv               *stdprometheus.GaugeVec
	labelNamesValues labelNamesValues
	collector        stdprometheus.Gauge
}

func (g *gauge) With(labelValues ...string) metrics.Gauge {
	lnv := g.labelNamesValues.With(labelValues...)
	return &gauge{
		name:             g.name,
		gv:               g.gv,
		labelNamesValues: lnv,
		collector:        g.gv.With(lnv.ToLabels()),
	}
}

func (g *gauge) Add(delta float64) {
	g.collector.Add(delta)
}

func (g *gauge) Set(value float64) {
	g.collector.Set(value)
}

func (g *gauge) Describe(ch chan<- *stdprometheus.Desc) {
	g.gv.Describe(ch)
}

func newHistogramFrom(opts stdprometheus.HistogramOpts, labelNames []string) *histogram {
	hv := stdprometheus.NewHistogramVec(opts, labelNames)
	return &histogram{
		name: opts.Name,
		hv:   hv,
	}
}

type histogram struct {
	name             string
	hv               *stdprometheus.HistogramVec
	labelNamesValues labelNamesValues
	collector        stdprometheus.Observer
}

func (h *histogram) With(labelValues ...string) metrics.Histogram {
	lnv := h.labelNamesValues.With(labelValues...)
	return &histogram{
		name:             h.name,
		hv:               h.hv,
		labelNamesValues: lnv,
		collector:        h.hv.With(lnv.ToLabels()),
	}
}

func (h *histogram) Observe(value float64) {
	h.collector.Observe(value)
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

	labels := make([]string, len(lvs)+len(labelValues))
	n := copy(labels, lvs)
	copy(labels[n:], labelValues)

	return labels
}

// ToLabels is a convenience method to convert a labelNamesValues
// to the native prometheus.Labels.
func (lvs labelNamesValues) ToLabels() stdprometheus.Labels {
	labels := make(map[string]string, len(lvs)/2)
	for i := 0; i < len(lvs); i += 2 {
		labels[lvs[i]] = lvs[i+1]
	}
	return labels
}
