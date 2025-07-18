package metrics

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/types"
	"github.com/traefik/traefik/v3/pkg/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
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
	httpServerRequestDuration metric.Float64Histogram
	// client metrics
	httpClientRequestDuration metric.Float64Histogram
}

// NewSemConvMetricRegistry registers all stables semantic conventions metrics.
func NewSemConvMetricRegistry(ctx context.Context, config *types.OTLP) (*SemConvMetricsRegistry, error) {
	if openTelemetryMeterProvider == nil {
		var err error
		if openTelemetryMeterProvider, err = newOpenTelemetryMeterProvider(ctx, config); err != nil {
			log.Ctx(ctx).Err(err).Msg("Unable to create OpenTelemetry meter provider")

			return nil, nil
		}
	}

	meter := otel.Meter("github.com/traefik/traefik",
		metric.WithInstrumentationVersion(version.Version))

	httpServerRequestDuration, err := meter.Float64Histogram(semconv.HTTPServerRequestDurationName,
		metric.WithDescription(semconv.HTTPServerRequestDurationDescription),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(config.ExplicitBoundaries...))
	if err != nil {
		return nil, fmt.Errorf("can't build httpServerRequestDuration histogram: %w", err)
	}

	httpClientRequestDuration, err := meter.Float64Histogram(semconv.HTTPClientRequestDurationName,
		metric.WithDescription(semconv.HTTPClientRequestDurationDescription),
		metric.WithUnit("s"),
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
func (s *SemConvMetricsRegistry) HTTPServerRequestDuration() metric.Float64Histogram {
	if s == nil {
		return nil
	}

	return s.httpServerRequestDuration
}

// HTTPClientRequestDuration returns the HTTP client request duration histogram.
func (s *SemConvMetricsRegistry) HTTPClientRequestDuration() metric.Float64Histogram {
	if s == nil {
		return nil
	}

	return s.httpClientRequestDuration
}

// RegisterOpenTelemetry registers all OpenTelemetry metrics.
func RegisterOpenTelemetry(ctx context.Context, config *types.OTLP) Registry {
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
		tlsCertsNotAfterTimestampGauge: newOTLPGaugeFrom(meter, tlsCertsNotAfterTimestampName, "Certificate expiration timestamp", "ms"),
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
func newOpenTelemetryMeterProvider(ctx context.Context, config *types.OTLP) (*sdkmetric.MeterProvider, error) {
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

func newHTTPExporter(ctx context.Context, config *types.OTelHTTP) (sdkmetric.Exporter, error) {
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

func newGRPCExporter(ctx context.Context, config *types.OTelGRPC) (sdkmetric.Exporter, error) {
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
}

func newOpenTelemetryGaugeCollector() *gaugeCollector {
	return &gaugeCollector{
		values: make(map[string]map[string]gaugeValue),
	}
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

		for _, value := range values {
			observer.ObserveFloat64(c, value.value, metric.WithAttributes(value.attributes.ToLabels()...))
		}

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

// ToLabels is a convenience method to convert a otelLabelNamesValues
// to the native attribute.KeyValue.
func (lvs otelLabelNamesValues) ToLabels() []attribute.KeyValue {
	labels := make([]attribute.KeyValue, len(lvs)/2)
	for i := 0; i < len(labels); i++ {
		labels[i] = attribute.String(lvs[2*i], lvs[2*i+1])
	}
	return labels
}
