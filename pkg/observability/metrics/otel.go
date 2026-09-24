package metrics

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/observability"
	otypes "github.com/traefik/traefik/v3/pkg/observability/types"
	"github.com/traefik/traefik/v3/pkg/types"
	"github.com/traefik/traefik/v3/pkg/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/semconv/v1.37.0/httpconv"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

var (
	openTelemetryMeterProvider  *sdkmetric.MeterProvider
	openTelemetryGaugeCollector *gaugeCollector
)

// SetMeterProvider sets the meter provider for the tests.
func SetMeterProvider(meterProvider *sdkmetric.MeterProvider) {
	openTelemetryMeterProvider = meterProvider
	otel.SetMeterProvider(meterProvider)
}

// SemConvMetricsRegistry holds stables semantic conventions metric instruments.
type SemConvMetricsRegistry struct {
	// server metrics
	httpServerRequestDuration httpconv.ServerRequestDuration
	// client metrics
	httpClientRequestDuration httpconv.ClientRequestDuration
}

// NewSemConvMetricRegistry registers all stables semantic conventions metrics.
func NewSemConvMetricRegistry(ctx context.Context, config *otypes.OTLP) (*SemConvMetricsRegistry, error) {
	if err := observability.EnsureUserEnvVar(); err != nil {
		return nil, err
	}

	if openTelemetryMeterProvider == nil {
		var err error
		if openTelemetryMeterProvider, err = newOpenTelemetryMeterProvider(ctx, config); err != nil {
			log.Ctx(ctx).Err(err).Msg("Unable to create OpenTelemetry meter provider")

			return nil, nil
		}
	}

	meter := otel.Meter("github.com/traefik/traefik",
		metric.WithInstrumentationVersion(version.Version))

	httpServerRequestDuration, err := httpconv.NewServerRequestDuration(meter,
		metric.WithExplicitBucketBoundaries(config.ExplicitBoundaries...))
	if err != nil {
		return nil, fmt.Errorf("can't build httpServerRequestDuration histogram: %w", err)
	}

	httpClientRequestDuration, err := httpconv.NewClientRequestDuration(meter,
		metric.WithExplicitBucketBoundaries(config.ExplicitBoundaries...))
	if err != nil {
		return nil, fmt.Errorf("can't build httpClientRequestDuration histogram: %w", err)
	}

	return &SemConvMetricsRegistry{
		httpServerRequestDuration: httpServerRequestDuration,
		httpClientRequestDuration: httpClientRequestDuration,
	}, nil
}

// HTTPServerRequestDuration returns the HTTP server request duration histogram.
func (s *SemConvMetricsRegistry) HTTPServerRequestDuration() httpconv.ServerRequestDuration {
	if s == nil {
		return httpconv.ServerRequestDuration{}
	}

	return s.httpServerRequestDuration
}

// HTTPClientRequestDuration returns the HTTP client request duration histogram.
func (s *SemConvMetricsRegistry) HTTPClientRequestDuration() httpconv.ClientRequestDuration {
	if s == nil {
		return httpconv.ClientRequestDuration{}
	}

	return s.httpClientRequestDuration
}

// RegisterOpenTelemetry registers all OpenTelemetry metrics.
func RegisterOpenTelemetry(ctx context.Context, config *otypes.OTLP) Registry {
	if openTelemetryMeterProvider == nil {
		var err error
		if openTelemetryMeterProvider, err = newOpenTelemetryMeterProvider(ctx, config); err != nil {
			log.Ctx(ctx).Err(err).Msg("Unable to create OpenTelemetry meter provider")

			return nil
		}
	}
	if openTelemetryGaugeCollector == nil {
		openTelemetryGaugeCollector = newOpenTelemetryGaugeCollector()
	}

	meter := otel.Meter("github.com/traefik/traefik",
		metric.WithInstrumentationVersion(version.Version))

	reg := &standardRegistry{
		epEnabled:                      config.AddEntryPointsLabels,
		routerEnabled:                  config.AddRoutersLabels,
		svcEnabled:                     config.AddServicesLabels,
		configReloadsCounter:           newOTLPCounterFrom(meter, configReloadsTotalName, "Config reloads"),
		lastConfigReloadSuccessGauge:   newOTLPGaugeFrom(meter, configLastReloadSuccessName, "Last config reload success", "ms"),
		openConnectionsGauge:           newOTLPGaugeFrom(meter, openConnectionsName, "How many open connections exist, by entryPoint and protocol", "1"),
		tlsCertsNotAfterTimestampGauge: newOTLPGaugeFrom(meter, tlsCertsNotAfterTimestampName, "Certificate expiration timestamp", "s"),
	}

	if config.AddEntryPointsLabels {
		reg.entryPointReqsCounter = NewCounterWithNoopHeaders(newOTLPCounterFrom(meter, entryPointReqsTotalName,
			"How many HTTP requests processed on an entrypoint, partitioned by status code, protocol, and method."))
		reg.entryPointReqsTLSCounter = newOTLPCounterFrom(meter, entryPointReqsTLSTotalName,
			"How many HTTP requests with TLS processed on an entrypoint, partitioned by TLS Version and TLS cipher Used.")
		reg.entryPointReqDurationHistogram, _ = NewHistogramWithScale(newOTLPHistogramFrom(meter, entryPointReqDurationName,
			"How long it took to process the request on an entrypoint, partitioned by status code, protocol, and method.",
			"s"), time.Second)
		reg.entryPointReqsBytesCounter = newOTLPCounterFrom(meter, entryPointReqsBytesTotalName,
			"The total size of requests in bytes handled by an entrypoint, partitioned by status code, protocol, and method.")
		reg.entryPointRespsBytesCounter = newOTLPCounterFrom(meter, entryPointRespsBytesTotalName,
			"The total size of responses in bytes handled by an entrypoint, partitioned by status code, protocol, and method.")
	}

	if config.AddRoutersLabels {
		reg.routerReqsCounter = NewCounterWithNoopHeaders(newOTLPCounterFrom(meter, routerReqsTotalName,
			"How many HTTP requests are processed on a router, partitioned by service, status code, protocol, and method."))
		reg.routerReqsTLSCounter = newOTLPCounterFrom(meter, routerReqsTLSTotalName,
			"How many HTTP requests with TLS are processed on a router, partitioned by service, TLS Version, and TLS cipher Used.")
		reg.routerReqDurationHistogram, _ = NewHistogramWithScale(newOTLPHistogramFrom(meter, routerReqDurationName,
			"How long it took to process the request on a router, partitioned by service, status code, protocol, and method.",
			"s"), time.Second)
		reg.routerReqsBytesCounter = newOTLPCounterFrom(meter, routerReqsBytesTotalName,
			"The total size of requests in bytes handled by a router, partitioned by status code, protocol, and method.")
		reg.routerRespsBytesCounter = newOTLPCounterFrom(meter, routerRespsBytesTotalName,
			"The total size of responses in bytes handled by a router, partitioned by status code, protocol, and method.")
	}

	if config.AddServicesLabels {
		reg.serviceReqsCounter = NewCounterWithNoopHeaders(newOTLPCounterFrom(meter, serviceReqsTotalName,
			"How many HTTP requests processed on a service, partitioned by status code, protocol, and method."))
		reg.serviceReqsTLSCounter = newOTLPCounterFrom(meter, serviceReqsTLSTotalName,
			"How many HTTP requests with TLS processed on a service, partitioned by TLS version and TLS cipher.")
		reg.serviceReqDurationHistogram, _ = NewHistogramWithScale(newOTLPHistogramFrom(meter, serviceReqDurationName,
			"How long it took to process the request on a service, partitioned by status code, protocol, and method.",
			"s"), time.Second)
		reg.serviceRetriesCounter = newOTLPCounterFrom(meter, serviceRetriesTotalName,
			"How many request retries happened on a service.")
		reg.serviceServerUpGauge = newOTLPGaugeFrom(meter, serviceServerUpName,
			"service server is up, described by gauge value of 0 or 1.",
			"1")
		reg.serviceReqsBytesCounter = newOTLPCounterFrom(meter, serviceReqsBytesTotalName,
			"The total size of requests in bytes received by a service, partitioned by status code, protocol, and method.")
		reg.serviceRespsBytesCounter = newOTLPCounterFrom(meter, serviceRespsBytesTotalName,
			"The total size of responses in bytes returned by a service, partitioned by status code, protocol, and method.")
	}

	return reg
}

// StopOpenTelemetry stops and resets Open-Telemetry client.
func StopOpenTelemetry() {
	if openTelemetryMeterProvider == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := openTelemetryMeterProvider.Shutdown(ctx); err != nil {
		log.Err(err).Msg("Unable to shutdown OpenTelemetry meter provider")
	}

	openTelemetryMeterProvider = nil
}

// newOpenTelemetryMeterProvider creates a new controller.Controller.
func newOpenTelemetryMeterProvider(ctx context.Context, config *otypes.OTLP) (*sdkmetric.MeterProvider, error) {
	var (
		exporter sdkmetric.Exporter
		err      error
	)
	if config.GRPC != nil {
		exporter, err = newGRPCExporter(ctx, config.GRPC)
	} else {
		exporter, err = newHTTPExporter(ctx, config.HTTP)
	}
	if err != nil {
		return nil, fmt.Errorf("creating exporter: %w", err)
	}

	var resAttrs []attribute.KeyValue
	for k, v := range config.ResourceAttributes {
		resAttrs = append(resAttrs, attribute.String(k, v))
	}

	res, err := resource.New(ctx,
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithDetectors(types.K8sAttributesDetector{}),
		// The following order allows the user to override the service name and version,
		// as well as any other attributes set by the above detectors.
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(version.Version),
		),
		resource.WithAttributes(resAttrs...),
		// Use the environment variables to allow overriding above resource attributes.
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("building resource: %w", err)
	}

	opts := []sdkmetric.PeriodicReaderOption{
		sdkmetric.WithInterval(time.Duration(config.PushInterval)),
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, opts...)),
		// View to customize histogram buckets and rename a single histogram instrument.
		sdkmetric.WithView(sdkmetric.NewView(
			sdkmetric.Instrument{Name: "traefik_*_request_duration_seconds"},
			sdkmetric.Stream{Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: config.ExplicitBoundaries,
			}},
		)),
	)

	otel.SetMeterProvider(meterProvider)

	return meterProvider, nil
}

func newHTTPExporter(ctx context.Context, config *otypes.OTelHTTP) (sdkmetric.Exporter, error) {
	endpoint, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid collector endpoint %q: %w", config.Endpoint, err)
	}

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(endpoint.Host),
		otlpmetrichttp.WithHeaders(config.Headers),
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
	}

	if endpoint.Scheme == "http" {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	if endpoint.Path != "" {
		opts = append(opts, otlpmetrichttp.WithURLPath(endpoint.Path))
	}

	if config.TLS != nil {
		tlsConfig, err := config.TLS.CreateTLSConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating TLS client config: %w", err)
		}

		opts = append(opts, otlpmetrichttp.WithTLSClientConfig(tlsConfig))
	}

	return otlpmetrichttp.New(ctx, opts...)
}

func newGRPCExporter(ctx context.Context, config *otypes.OTelGRPC) (sdkmetric.Exporter, error) {
	host, port, err := net.SplitHostPort(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid collector endpoint %q: %w", config.Endpoint, err)
	}

	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(fmt.Sprintf("%s:%s", host, port)),
		otlpmetricgrpc.WithHeaders(config.Headers),
		otlpmetricgrpc.WithCompressor(gzip.Name),
	}

	if config.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	if config.TLS != nil {
		tlsConfig, err := config.TLS.CreateTLSConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating TLS client config: %w", err)
		}

		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
	}

	return otlpmetricgrpc.New(ctx, opts...)
}

func newOTLPCounterFrom(meter metric.Meter, name, desc string) *otelCounter {
	c, _ := meter.Float64Counter(name,
		metric.WithDescription(desc),
		metric.WithUnit("1"),
	)

	return &otelCounter{
		ip: c,
	}
}

type otelCounter struct {
	labelNamesValues otelLabelNamesValues
	ip               metric.Float64Counter
}

func (c *otelCounter) With(labelValues ...string) metrics.Counter {
	return &otelCounter{
		labelNamesValues: c.labelNamesValues.With(labelValues...),
		ip:               c.ip,
	}
}

func (c *otelCounter) Add(delta float64) {
	c.ip.Add(context.Background(), delta, metric.WithAttributes(c.labelNamesValues.ToLabels()...))
}

type gaugeValue struct {
	attributes otelLabelNamesValues
	value      float64
}

type gaugeCollector struct {
	mu     sync.Mutex
	values map[string]map[string]gaugeValue

	// dynConfig holds the current dynamic configuration. Together with the
	// deleted* fields below, it is used to detect gauge values that no
	// longer belong to any entryPoint, router, service, or server, the same
	// way promState does for Prometheus. Without this, gauge values for
	// dynamically created and removed entryPoints/routers/services/servers
	// would accumulate forever, leading to unbounded memory growth.
	dynConfig       *dynamicConfig
	deletedEP       []string
	deletedRouters  []string
	deletedServices []string
	deletedURLs     map[string][]string

	// pendingCallbacks counts how many of the registered gauges' collect
	// callbacks still have to run before every value marked for deletion
	// below has had a chance to be observed once more. It resets to
	// len(values) on every setDynamicConfig call, and reaching zero means
	// it is safe to clear the deleted* fields, the same way promState clears
	// them at the end of its single, shared Collect call.
	pendingCallbacks int
}

func newOpenTelemetryGaugeCollector() *gaugeCollector {
	return &gaugeCollector{
		values:      make(map[string]map[string]gaugeValue),
		dynConfig:   newDynamicConfig(),
		deletedURLs: make(map[string][]string),
	}
}

// setDynamicConfig updates the dynamic configuration used to detect stale
// gauge values, recording which entryPoints/routers/services/server URLs
// were removed since the previous call, mirroring promState.SetDynamicConfig.
func (c *gaugeCollector) setDynamicConfig(dynConfig *dynamicConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for ep := range c.dynConfig.entryPoints {
		if !dynConfig.hasEntryPoint(ep) {
			c.deletedEP = append(c.deletedEP, ep)
		}
	}

	for router := range c.dynConfig.routers {
		if !dynConfig.hasRouter(router) {
			c.deletedRouters = append(c.deletedRouters, router)
		}
	}

	for service, urls := range c.dynConfig.services {
		if !dynConfig.hasService(service) {
			c.deletedServices = append(c.deletedServices, service)
		}

		for url := range urls {
			if !dynConfig.hasServerURL(service, url) {
				c.deletedURLs[service] = append(c.deletedURLs[service], url)
			}
		}
	}

	c.dynConfig = dynConfig
	c.pendingCallbacks = len(c.values)
}

// isStale reports whether the given attributes reference an entryPoint,
// router, service, or server URL that was removed from the dynamic
// configuration. Unlike a plain absence-from-config check, this only
// flags values that were previously known and have since been removed, so
// a value that has simply not been declared yet (for instance because it
// is observed before the first setDynamicConfig call) is never mistaken
// for stale. The caller must hold c.mu.
func (c *gaugeCollector) isStale(attributes otelLabelNamesValues) bool {
	if ep, ok := attributes.value("entrypoint"); ok && slices.Contains(c.deletedEP, ep) && !c.dynConfig.hasEntryPoint(ep) {
		return true
	}

	if router, ok := attributes.value("router"); ok && slices.Contains(c.deletedRouters, router) && !c.dynConfig.hasRouter(router) {
		return true
	}

	service, hasService := attributes.value("service")
	if hasService && slices.Contains(c.deletedServices, service) && !c.dynConfig.hasService(service) {
		return true
	}

	if url, ok := attributes.value("url"); ok && hasService && slices.Contains(c.deletedURLs[service], url) && !c.dynConfig.hasServerURL(service, url) {
		return true
	}

	return false
}

// endCallback must be called once by every gauge's collect callback after
// it is done observing its values. Once every gauge registered on this
// collector has called it since the last setDynamicConfig call, the
// deleted* fields are cleared, as they have then all had a chance to prune
// their stale values. The caller must hold c.mu.
func (c *gaugeCollector) endCallback() {
	if c.pendingCallbacks == 0 {
		return
	}

	c.pendingCallbacks--
	if c.pendingCallbacks > 0 {
		return
	}

	c.deletedEP = nil
	c.deletedRouters = nil
	c.deletedServices = nil
	c.deletedURLs = make(map[string][]string)
}

func (c *gaugeCollector) add(name string, delta float64, attributes otelLabelNamesValues) {
	c.mu.Lock()
	defer c.mu.Unlock()

	str := strings.Join(attributes, "")

	if _, exists := c.values[name]; !exists {
		c.values[name] = map[string]gaugeValue{
			str: {
				attributes: attributes,
				value:      delta,
			},
		}
		return
	}

	v, exists := c.values[name][str]
	if !exists {
		c.values[name][str] = gaugeValue{
			attributes: attributes,
			value:      delta,
		}
		return
	}

	c.values[name][str] = gaugeValue{
		attributes: attributes,
		value:      v.value + delta,
	}
}

func (c *gaugeCollector) set(name string, value float64, attributes otelLabelNamesValues) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.values[name]; !exists {
		c.values[name] = make(map[string]gaugeValue)
	}

	c.values[name][strings.Join(attributes, "")] = gaugeValue{
		attributes: attributes,
		value:      value,
	}
}

func newOTLPGaugeFrom(meter metric.Meter, name, desc string, unit string) *otelGauge {
	openTelemetryGaugeCollector.values[name] = make(map[string]gaugeValue)

	c, _ := meter.Float64ObservableGauge(name,
		metric.WithDescription(desc),
		metric.WithUnit(unit),
	)

	_, err := meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		openTelemetryGaugeCollector.mu.Lock()
		defer openTelemetryGaugeCollector.mu.Unlock()

		values, exists := openTelemetryGaugeCollector.values[name]
		if !exists {
			return nil
		}

		for key, value := range values {
			observer.ObserveFloat64(c, value.value, metric.WithAttributes(value.attributes.ToLabels()...))

			// Remove stale values once they have been observed a last time,
			// so that gauges for entryPoints/routers/services/servers that
			// no longer exist in the dynamic configuration don't accumulate
			// forever and leak memory.
			if openTelemetryGaugeCollector.isStale(value.attributes) {
				delete(values, key)
			}
		}

		openTelemetryGaugeCollector.endCallback()

		return nil
	}, c)
	if err != nil {
		log.Err(err).Msg("Unable to register OpenTelemetry meter callback")
	}

	return &otelGauge{
		ip:   c,
		name: name,
	}
}

type otelGauge struct {
	labelNamesValues otelLabelNamesValues
	ip               metric.Float64ObservableGauge
	name             string
}

func (g *otelGauge) With(labelValues ...string) metrics.Gauge {
	return &otelGauge{
		labelNamesValues: g.labelNamesValues.With(labelValues...),
		ip:               g.ip,
		name:             g.name,
	}
}

func (g *otelGauge) Add(delta float64) {
	openTelemetryGaugeCollector.add(g.name, delta, g.labelNamesValues)
}

func (g *otelGauge) Set(value float64) {
	openTelemetryGaugeCollector.set(g.name, value, g.labelNamesValues)
}

func newOTLPHistogramFrom(meter metric.Meter, name, desc string, unit string) *otelHistogram {
	c, _ := meter.Float64Histogram(name,
		metric.WithDescription(desc),
		metric.WithUnit(unit),
	)

	return &otelHistogram{
		ip: c,
	}
}

type otelHistogram struct {
	labelNamesValues otelLabelNamesValues
	ip               metric.Float64Histogram
}

func (h *otelHistogram) With(labelValues ...string) metrics.Histogram {
	return &otelHistogram{
		labelNamesValues: h.labelNamesValues.With(labelValues...),
		ip:               h.ip,
	}
}

func (h *otelHistogram) Observe(incr float64) {
	h.ip.Record(context.Background(), incr, metric.WithAttributes(h.labelNamesValues.ToLabels()...))
}

// otelLabelNamesValues is the equivalent of prometheus' labelNamesValues
// but adapted to OpenTelemetry.
// otelLabelNamesValues is a type alias that provides validation on its With
// method.
// Metrics may include it as a member to help them satisfy With semantics and
// save some code duplication.
type otelLabelNamesValues []string

// With validates the input, and returns a new aggregate otelLabelNamesValues.
func (lvs otelLabelNamesValues) With(labelValues ...string) otelLabelNamesValues {
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, "unknown")
	}
	return append(lvs, labelValues...)
}

// value returns the value associated with the given label name, and whether
// it was found.
func (lvs otelLabelNamesValues) value(name string) (string, bool) {
	for i := 0; i+1 < len(lvs); i += 2 {
		if lvs[i] == name {
			return lvs[i+1], true
		}
	}
	return "", false
}

// ToLabels is a convenience method to convert a otelLabelNamesValues
// to the native attribute.KeyValue.
func (lvs otelLabelNamesValues) ToLabels() []attribute.KeyValue {
	labels := make([]attribute.KeyValue, len(lvs)/2)
	for i := range labels {
		labels[i] = attribute.String(lvs[2*i], lvs[2*i+1])
	}
	return labels
}
