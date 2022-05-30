package metrics

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/types"
	traefikversion "github.com/traefik/traefik/v2/pkg/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/unit"
	histogramAggregator "go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

var (
	openTelemetryController     *controller.Controller
	openTelemetryGaugeCollector *gaugeCollector
)

// RegisterOpenTelemetry registers all OpenTelemetry metrics.
func RegisterOpenTelemetry(ctx context.Context, config *types.OpenTelemetry) Registry {
	if openTelemetryController == nil {
		var err error
		if openTelemetryController, err = newOpenTelemetryController(ctx, config); err != nil {
			log.FromContext(ctx).Error(err)
			return nil
		}
	}

	if openTelemetryGaugeCollector == nil {
		openTelemetryGaugeCollector = newOpenTelemetryGaugeCollector()
	}

	// TODO add schema URL
	meter := global.Meter("github.com/traefik/traefik",
		metric.WithInstrumentationVersion(traefikversion.Version))

	reg := &standardRegistry{
		epEnabled:                      config.AddEntryPointsLabels,
		routerEnabled:                  config.AddRoutersLabels,
		svcEnabled:                     config.AddServicesLabels,
		configReloadsCounter:           newOTLPCounterFrom(meter, configReloadsTotalName, "Config reloads", unit.Dimensionless),
		configReloadsFailureCounter:    newOTLPCounterFrom(meter, configReloadsFailuresTotalName, "Config failure reloads", unit.Dimensionless),
		lastConfigReloadSuccessGauge:   newOTLPGaugeFrom(meter, configLastReloadSuccessName, "Last config reload success", unit.Milliseconds),
		lastConfigReloadFailureGauge:   newOTLPGaugeFrom(meter, configLastReloadFailureName, "Last config reload failure", unit.Milliseconds),
		tlsCertsNotAfterTimestampGauge: newOTLPGaugeFrom(meter, tlsCertsNotAfterTimestamp, "Certificate expiration timestamp", unit.Milliseconds),
	}

	if config.AddEntryPointsLabels {
		reg.entryPointReqsCounter = newOTLPCounterFrom(meter, entryPointReqsTotalName,
			"How many HTTP requests processed on an entrypoint, partitioned by status code, protocol, and method.",
			unit.Dimensionless)
		reg.entryPointReqsTLSCounter = newOTLPCounterFrom(meter, entryPointReqsTLSTotalName,
			"How many HTTP requests with TLS processed on an entrypoint, partitioned by TLS Version and TLS cipher Used.",
			unit.Dimensionless)
		reg.entryPointReqDurationHistogram, _ = NewHistogramWithScale(newOTLPHistogramFrom(meter, entryPointReqDurationName,
			"How long it took to process the request on an entrypoint, partitioned by status code, protocol, and method.",
			unit.Milliseconds), time.Second)
		reg.entryPointOpenConnsGauge = newOTLPGaugeFrom(meter, entryPointOpenConnsName,
			"How many open connections exist on an entrypoint, partitioned by method and protocol.",
			unit.Dimensionless)
	}

	if config.AddRoutersLabels {
		reg.routerReqsCounter = newOTLPCounterFrom(meter, routerReqsTotalName,
			"How many HTTP requests are processed on a router, partitioned by service, status code, protocol, and method.",
			unit.Dimensionless)
		reg.routerReqsTLSCounter = newOTLPCounterFrom(meter, routerReqsTLSTotalName,
			"How many HTTP requests with TLS are processed on a router, partitioned by service, TLS Version, and TLS cipher Used.",
			unit.Dimensionless)
		reg.routerReqDurationHistogram, _ = NewHistogramWithScale(newOTLPHistogramFrom(meter, routerReqDurationName,
			"How long it took to process the request on a router, partitioned by service, status code, protocol, and method.",
			unit.Milliseconds), time.Second)
		reg.routerOpenConnsGauge = newOTLPGaugeFrom(meter, routerOpenConnsName,
			"How many open connections exist on a router, partitioned by service, method, and protocol.",
			unit.Dimensionless)
	}

	if config.AddServicesLabels {
		reg.serviceReqsCounter = newOTLPCounterFrom(meter, serviceReqsTotalName,
			"How many HTTP requests processed on a service, partitioned by status code, protocol, and method.",
			unit.Dimensionless)
		reg.serviceReqsTLSCounter = newOTLPCounterFrom(meter, serviceReqsTLSTotalName,
			"How many HTTP requests with TLS processed on a service, partitioned by TLS version and TLS cipher.",
			unit.Dimensionless)
		reg.serviceReqDurationHistogram, _ = NewHistogramWithScale(newOTLPHistogramFrom(meter, serviceReqDurationName,
			"How long it took to process the request on a service, partitioned by status code, protocol, and method.",
			unit.Milliseconds), time.Second)
		reg.serviceOpenConnsGauge = newOTLPGaugeFrom(meter, serviceOpenConnsName,
			"How many open connections exist on a service, partitioned by method and protocol.",
			unit.Dimensionless)
		reg.serviceRetriesCounter = newOTLPCounterFrom(meter, serviceRetriesTotalName,
			"How many request retries happened on a service.",
			unit.Dimensionless)
		reg.serviceServerUpGauge = newOTLPGaugeFrom(meter, serviceServerUpName,
			"service server is up, described by gauge value of 0 or 1.",
			unit.Dimensionless)
	}

	return reg
}

// StopOpenTelemetry stops and resets Open-Telemetry client.
func StopOpenTelemetry() {
	if openTelemetryController == nil || !openTelemetryController.IsRunning() {
		return
	}

	if err := openTelemetryController.Stop(context.Background()); err != nil {
		log.WithoutContext().Error(err)
	}

	openTelemetryController = nil
}

// newOpenTelemetryController creates a new controller.Controller.
func newOpenTelemetryController(ctx context.Context, config *types.OpenTelemetry) (*controller.Controller, error) {
	if config.PushInterval <= 0 {
		return nil, errors.New("PushInterval must be greater than zero")
	}

	factory := processor.NewFactory(
		simple.NewWithHistogramDistribution(histogramAggregator.WithExplicitBoundaries(config.ExplicitBoundaries)),
		aggregation.CumulativeTemporalitySelector(),
		processor.WithMemory(config.WithMemory),
	)

	var (
		exporter export.Exporter
		err      error
	)
	if config.GRPC != nil {
		exporter, err = newGRPCExporter(ctx, config)
	} else {
		exporter, err = newHTTPExporter(ctx, config)
	}
	if err != nil {
		return nil, err
	}

	// TODO add schema URL
	optsResource := []resource.Option{
		resource.WithAttributes(),
		resource.WithContainer(),
		resource.WithContainerID(),
		resource.WithDetectors(),
		resource.WithFromEnv(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithOSDescription(),
		resource.WithOSType(),
		resource.WithProcess(),
		resource.WithProcessCommandArgs(),
		resource.WithProcessExecutableName(),
		resource.WithProcessExecutablePath(),
		resource.WithProcessOwner(),
		resource.WithProcessPID(),
		resource.WithProcessRuntimeDescription(),
		resource.WithProcessRuntimeName(),
		resource.WithProcessRuntimeVersion(),
		resource.WithTelemetrySDK(),
	}

	r, err := resource.New(ctx, optsResource...)
	if err != nil {
		return nil, err
	}

	optsController := []controller.Option{
		controller.WithCollectPeriod(time.Duration(config.PushInterval)),
		controller.WithCollectTimeout(time.Duration(config.PushTimeout)),
		controller.WithExporter(exporter),
		controller.WithResource(r),
	}

	c := controller.New(factory, optsController...)

	global.SetMeterProvider(c)

	if err := c.Start(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

func newHTTPExporter(ctx context.Context, config *types.OpenTelemetry) (export.Exporter, error) {
	u, err := url.Parse(config.Address)
	if err != nil {
		return nil, err
	}

	// https://github.com/open-telemetry/opentelemetry-go/blob/exporters/otlp/otlpmetric/v0.30.0/exporters/otlp/otlpmetric/internal/otlpconfig/options.go#L39
	path := "/v1/metrics"
	if u.Path != "" {
		path = u.Path
	}

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(u.Host),
		otlpmetrichttp.WithHeaders(config.Headers),
		otlpmetrichttp.WithTimeout(config.Timeout),
		otlpmetrichttp.WithURLPath(path),
	}

	if config.Compress {
		opts = append(opts, otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression))
	}

	if config.Retry != nil {
		opts = append(opts, otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{
			Enabled:         true,
			InitialInterval: config.Retry.InitialInterval,
			MaxElapsedTime:  config.Retry.MaxElapsedTime,
			MaxInterval:     config.Retry.MaxInterval,
		}))
	}

	if u.Scheme == "http" {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	if config.TLS != nil {
		tlsConfig, err := config.TLS.CreateTLSConfig(ctx)
		if err != nil {
			return nil, err
		}

		opts = append(opts, otlpmetrichttp.WithTLSClientConfig(tlsConfig))
	}

	return otlpmetrichttp.New(ctx, opts...)
}

func newGRPCExporter(ctx context.Context, config *types.OpenTelemetry) (export.Exporter, error) {
	u, err := url.Parse(config.Address)
	if err != nil {
		return nil, err
	}

	// TODO: handle DialOption
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(u.Host),
		otlpmetricgrpc.WithHeaders(config.Headers),
		otlpmetricgrpc.WithReconnectionPeriod(config.GRPC.ReconnectionPeriod),
		otlpmetricgrpc.WithServiceConfig(config.GRPC.ServiceConfig),
		otlpmetricgrpc.WithTimeout(config.Timeout),
	}

	if config.Compress {
		opts = append(opts, otlpmetricgrpc.WithCompressor(gzip.Name))
	}

	if config.GRPC.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	if config.Retry != nil {
		opts = append(opts, otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: config.Retry.InitialInterval,
			MaxElapsedTime:  config.Retry.MaxElapsedTime,
			MaxInterval:     config.Retry.MaxInterval,
		}))
	}

	if config.TLS != nil {
		tlsConfig, err := config.TLS.CreateTLSConfig(ctx)
		if err != nil {
			return nil, err
		}

		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
	}

	return otlpmetricgrpc.New(ctx, opts...)
}

func newOTLPCounterFrom(meter metric.Meter, name, desc string, u unit.Unit) *otelCounter {
	c, _ := meter.SyncFloat64().Counter(name,
		instrument.WithDescription(desc),
		instrument.WithUnit(u),
	)

	return &otelCounter{
		ip: c,
	}
}

type otelCounter struct {
	labelNamesValues otelLabelNamesValues
	ip               syncfloat64.Counter
}

func (c *otelCounter) With(labelValues ...string) metrics.Counter {
	return &otelCounter{
		labelNamesValues: c.labelNamesValues.With(labelValues...),
		ip:               c.ip,
	}
}

func (c *otelCounter) Add(delta float64) {
	c.ip.Add(context.Background(), delta, c.labelNamesValues.ToLabels()...)
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

func (c *gaugeCollector) add(delta float64, name string, attributes otelLabelNamesValues) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exist := c.values[name]; !exist {
		c.values[name] = make(map[string]gaugeValue)
	}

	str := attributes.ToString()
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

func (c *gaugeCollector) set(value float64, name string, attributes otelLabelNamesValues) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exist := c.values[name]; !exist {
		c.values[name] = make(map[string]gaugeValue)
	}

	c.values[name][attributes.ToString()] = gaugeValue{
		attributes: attributes,
		value:      value,
	}
}

func newOTLPGaugeFrom(meter metric.Meter, name, desc string, u unit.Unit) *otelGauge {
	openTelemetryGaugeCollector.values[name] = make(map[string]gaugeValue)

	c, _ := meter.AsyncFloat64().Gauge(name,
		instrument.WithDescription(desc),
		instrument.WithUnit(u),
	)

	err := meter.RegisterCallback([]instrument.Asynchronous{c}, func(ctx context.Context) {
		openTelemetryGaugeCollector.mu.Lock()
		defer openTelemetryGaugeCollector.mu.Unlock()

		values, exist := openTelemetryGaugeCollector.values[name]
		if !exist {
			return
		}

		for _, value := range values {
			c.Observe(ctx, value.value, value.attributes.ToLabels()...)
		}
	})
	if err != nil {
		log.WithoutContext().Error(err)
	}

	return &otelGauge{
		ip:   c,
		name: name,
	}
}

type otelGauge struct {
	labelNamesValues otelLabelNamesValues
	ip               asyncfloat64.Gauge
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
	openTelemetryGaugeCollector.add(delta, g.name, g.labelNamesValues)
}

func (g *otelGauge) Set(value float64) {
	openTelemetryGaugeCollector.set(value, g.name, g.labelNamesValues)
}

func newOTLPHistogramFrom(meter metric.Meter, name, desc string, u unit.Unit) *otelHistogram {
	c, _ := meter.SyncFloat64().Histogram(name,
		instrument.WithDescription(desc),
		instrument.WithUnit(u),
	)

	return &otelHistogram{
		ip: c,
	}
}

type otelHistogram struct {
	labelNamesValues otelLabelNamesValues
	ip               syncfloat64.Histogram
}

func (h *otelHistogram) With(labelValues ...string) metrics.Histogram {
	return &otelHistogram{
		labelNamesValues: h.labelNamesValues.With(labelValues...),
		ip:               h.ip,
	}
}

func (h *otelHistogram) Observe(incr float64) {
	h.ip.Record(context.Background(), incr, h.labelNamesValues.ToLabels()...)
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

// ToString convert the otelLabelNamesValues to String.
func (lvs otelLabelNamesValues) ToString() string {
	var res string
	for _, lv := range lvs {
		res += lv
	}
	return res
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
