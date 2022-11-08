package metrics

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/types"
	"github.com/traefik/traefik/v2/pkg/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/unit"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/view"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

var (
	openTelemetryMeterProvider  *sdkmetric.MeterProvider
	openTelemetryGaugeCollector *gaugeCollector
)

// RegisterOpenTelemetry registers all OpenTelemetry metrics.
func RegisterOpenTelemetry(ctx context.Context, config *types.OpenTelemetry) Registry {
	if openTelemetryMeterProvider == nil {
		var err error
		if openTelemetryMeterProvider, err = newOpenTelemetryMeterProvider(ctx, config); err != nil {
			log.FromContext(ctx).Error(err)
			return nil
		}
	}

	if openTelemetryGaugeCollector == nil {
		openTelemetryGaugeCollector = newOpenTelemetryGaugeCollector()
	}

	meter := global.Meter("github.com/traefik/traefik",
		metric.WithInstrumentationVersion(version.Version))

	reg := &standardRegistry{
		epEnabled:                      config.AddEntryPointsLabels,
		routerEnabled:                  config.AddRoutersLabels,
		svcEnabled:                     config.AddServicesLabels,
		configReloadsCounter:           newOTLPCounterFrom(meter, configReloadsTotalName, "Config reloads"),
		configReloadsFailureCounter:    newOTLPCounterFrom(meter, configReloadsFailuresTotalName, "Config failure reloads"),
		lastConfigReloadSuccessGauge:   newOTLPGaugeFrom(meter, configLastReloadSuccessName, "Last config reload success", unit.Milliseconds),
		lastConfigReloadFailureGauge:   newOTLPGaugeFrom(meter, configLastReloadFailureName, "Last config reload failure", unit.Milliseconds),
		tlsCertsNotAfterTimestampGauge: newOTLPGaugeFrom(meter, tlsCertsNotAfterTimestamp, "Certificate expiration timestamp", unit.Milliseconds),
	}

	if config.AddEntryPointsLabels {
		reg.entryPointReqsCounter = newOTLPCounterFrom(meter, entryPointReqsTotalName,
			"How many HTTP requests processed on an entrypoint, partitioned by status code, protocol, and method.")
		reg.entryPointReqsTLSCounter = newOTLPCounterFrom(meter, entryPointReqsTLSTotalName,
			"How many HTTP requests with TLS processed on an entrypoint, partitioned by TLS Version and TLS cipher Used.")
		reg.entryPointReqDurationHistogram, _ = NewHistogramWithScale(newOTLPHistogramFrom(meter, entryPointReqDurationName,
			"How long it took to process the request on an entrypoint, partitioned by status code, protocol, and method.",
			unit.Milliseconds), time.Second)
		reg.entryPointOpenConnsGauge = newOTLPGaugeFrom(meter, entryPointOpenConnsName,
			"How many open connections exist on an entrypoint, partitioned by method and protocol.",
			unit.Dimensionless)
	}

	if config.AddRoutersLabels {
		reg.routerReqsCounter = newOTLPCounterFrom(meter, routerReqsTotalName,
			"How many HTTP requests are processed on a router, partitioned by service, status code, protocol, and method.")
		reg.routerReqsTLSCounter = newOTLPCounterFrom(meter, routerReqsTLSTotalName,
			"How many HTTP requests with TLS are processed on a router, partitioned by service, TLS Version, and TLS cipher Used.")
		reg.routerReqDurationHistogram, _ = NewHistogramWithScale(newOTLPHistogramFrom(meter, routerReqDurationName,
			"How long it took to process the request on a router, partitioned by service, status code, protocol, and method.",
			unit.Milliseconds), time.Second)
		reg.routerOpenConnsGauge = newOTLPGaugeFrom(meter, routerOpenConnsName,
			"How many open connections exist on a router, partitioned by service, method, and protocol.",
			unit.Dimensionless)
	}

	if config.AddServicesLabels {
		reg.serviceReqsCounter = newOTLPCounterFrom(meter, serviceReqsTotalName,
			"How many HTTP requests processed on a service, partitioned by status code, protocol, and method.")
		reg.serviceReqsTLSCounter = newOTLPCounterFrom(meter, serviceReqsTLSTotalName,
			"How many HTTP requests with TLS processed on a service, partitioned by TLS version and TLS cipher.")
		reg.serviceReqDurationHistogram, _ = NewHistogramWithScale(newOTLPHistogramFrom(meter, serviceReqDurationName,
			"How long it took to process the request on a service, partitioned by status code, protocol, and method.",
			unit.Milliseconds), time.Second)
		reg.serviceOpenConnsGauge = newOTLPGaugeFrom(meter, serviceOpenConnsName,
			"How many open connections exist on a service, partitioned by method and protocol.",
			unit.Dimensionless)
		reg.serviceRetriesCounter = newOTLPCounterFrom(meter, serviceRetriesTotalName,
			"How many request retries happened on a service.")
		reg.serviceServerUpGauge = newOTLPGaugeFrom(meter, serviceServerUpName,
			"service server is up, described by gauge value of 0 or 1.",
			unit.Dimensionless)
	}

	return reg
}

// StopOpenTelemetry stops and resets Open-Telemetry client.
func StopOpenTelemetry() {
	if openTelemetryMeterProvider == nil {
		return
	}

	if err := openTelemetryMeterProvider.Shutdown(context.Background()); err != nil {
		log.WithoutContext().Error(err)
	}

	openTelemetryMeterProvider = nil
}

// newOpenTelemetryMeterProvider creates a new controller.Controller.
func newOpenTelemetryMeterProvider(ctx context.Context, config *types.OpenTelemetry) (*sdkmetric.MeterProvider, error) {
	if config.Address == "" {
		return nil, errors.New("address property is missing")
	}

	var (
		exporter sdkmetric.Exporter
		err      error
	)
	if config.GRPC != nil {
		exporter, err = newGRPCExporter(ctx, config)
	} else {
		exporter, err = newHTTPExporter(ctx, config)
	}
	if err != nil {
		return nil, fmt.Errorf("create exporter: %w", err)
	}

	opts := []sdkmetric.PeriodicReaderOption{
		sdkmetric.WithInterval(time.Duration(config.PushInterval)),
	}

	// View to customize histogram buckets and rename a single histogram instrument.
	customBucketsView, err := view.New(
		// Match* to match instruments
		view.MatchInstrumentName("traefik_*_request_duration_seconds"),

		view.WithSetAggregation(aggregation.ExplicitBucketHistogram{
			Boundaries: config.ExplicitBoundaries,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create histogram view: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(
		sdkmetric.NewPeriodicReader(exporter, opts...),
		customBucketsView,
	))

	global.SetMeterProvider(meterProvider)

	return meterProvider, nil
}

func newHTTPExporter(ctx context.Context, config *types.OpenTelemetry) (sdkmetric.Exporter, error) {
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(config.Address),
		otlpmetrichttp.WithHeaders(config.Headers),
	}

	if config.Compress {
		opts = append(opts, otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression))
	}

	if config.Insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	if config.Path != "" {
		opts = append(opts, otlpmetrichttp.WithURLPath(config.Path))
	}

	if config.TLS != nil {
		tlsConfig, err := config.TLS.CreateTLSConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("create TLS client config: %w", err)
		}

		opts = append(opts, otlpmetrichttp.WithTLSClientConfig(tlsConfig))
	}

	return otlpmetrichttp.New(ctx, opts...)
}

func newGRPCExporter(ctx context.Context, config *types.OpenTelemetry) (sdkmetric.Exporter, error) {
	// TODO: handle DialOption
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(config.Address),
		otlpmetricgrpc.WithHeaders(config.Headers),
	}

	if config.Compress {
		opts = append(opts, otlpmetricgrpc.WithCompressor(gzip.Name))
	}

	if config.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	if config.TLS != nil {
		tlsConfig, err := config.TLS.CreateTLSConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("create TLS client config: %w", err)
		}

		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
	}

	return otlpmetricgrpc.New(ctx, opts...)
}

func newOTLPCounterFrom(meter metric.Meter, name, desc string) *otelCounter {
	c, _ := meter.SyncFloat64().Counter(name,
		instrument.WithDescription(desc),
		instrument.WithUnit(unit.Dimensionless),
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
