package metrics

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/generic"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/types"
)

var (
	influxDB2Client   influxdb2.Client
	influxDB2WriteAPI api.WriteAPI
)

const (
	influxDB2MetricsServiceReqsName        = "traefik_service_requests_total"
	influxDB2MetricsServiceReqsTLSName     = "traefik_service_requests_tls_total"
	influxDB2MetricsServiceLatencyName     = "traefik_service_request_duration"
	influxDB2RetriesTotalName              = "traefik_service_retries_total"
	influxDB2ConfigReloadsName             = "traefik_config_reload_total"
	influxDB2ConfigReloadsFailureName      = influxDBConfigReloadsName + "_failure"
	influxDB2LastConfigReloadSuccessName   = "traefik_config_reload_lastSuccessTimestamp"
	influxDB2LastConfigReloadFailureName   = "traefik_config_reload_lastFailureTimestamp"
	influxDB2EntryPointReqsName            = "traefik_entrypoint_requests_total"
	influxDB2EntryPointReqsTLSName         = "traefik_entrypoint_requests_tls_total"
	influxDB2EntryPointReqDurationName     = "traefik_entrypoint_request_duration"
	influxDB2EntryPointOpenConnsName       = "traefik_entrypoint_connections_open"
	influxDB2RouterReqsName                = "traefik_router_requests_total"
	influxDB2RouterReqsTLSName             = "traefik_router_requests_tls_total"
	influxDB2RouterReqsDurationName        = "traefik_router_request_duration"
	influxDB2RouterOpenConnsName           = "traefik_router_connections_open"
	influxDB2OpenConnsName                 = "traefik_service_connections_open"
	influxDB2ServerUpName                  = "traefik_service_server_up"
	influxDB2TLSCertsNotAfterTimestampName = "traefik_tls_certs_notAfterTimestamp"
)

// RegisterInfluxDB2 creates metrics exporter for InfluxDB2.
func RegisterInfluxDB2(ctx context.Context, config *types.InfluxDB2) Registry {
	if influxDB2Client == nil {
		flushMs := uint(time.Duration(config.PushInterval).Milliseconds())
		options := influxdb2.DefaultOptions()
		options = options.SetBatchSize(config.BatchSize)
		options = options.SetFlushInterval(flushMs)
		influxDB2Client = influxdb2.NewClientWithOptions(config.Address, config.Token, options)
		if influxDB2Client == nil {
			log.FromContext(ctx).Error("Failed to connect to InfluxDB v2")
			return nil
		}

		influxDB2WriteAPI = influxDB2Client.WriteAPI(config.Org, config.Bucket)
		if influxDB2WriteAPI == nil {
			log.FromContext(ctx).Error("Failed to open InfluxDB v2 bucket")
			influxDB2Client.Close()
			influxDB2Client = nil
			return nil
		}
	}

	registry := &standardRegistry{
		configReloadsCounter:           newInfluxDB2Counter(influxDB2ConfigReloadsName),
		configReloadsFailureCounter:    newInfluxDB2Counter(influxDB2ConfigReloadsFailureName),
		lastConfigReloadSuccessGauge:   newInfluxDB2Gauge(influxDB2LastConfigReloadSuccessName),
		lastConfigReloadFailureGauge:   newInfluxDB2Gauge(influxDB2LastConfigReloadFailureName),
		tlsCertsNotAfterTimestampGauge: newInfluxDB2Gauge(influxDB2TLSCertsNotAfterTimestampName),
	}

	if config.AddEntryPointsLabels {
		registry.epEnabled = config.AddEntryPointsLabels
		registry.entryPointReqsCounter = newInfluxDB2Counter(influxDB2EntryPointReqsName)
		registry.entryPointReqsTLSCounter = newInfluxDB2Counter(influxDB2EntryPointReqsTLSName)
		registry.entryPointReqDurationHistogram, _ = NewHistogramWithScale(newInfluxDB2Histogram(influxDB2EntryPointReqDurationName), time.Second)
		registry.entryPointOpenConnsGauge = newInfluxDB2Gauge(influxDB2EntryPointOpenConnsName)
	}

	if config.AddRoutersLabels {
		registry.routerEnabled = config.AddRoutersLabels
		registry.routerReqsCounter = newInfluxDB2Counter(influxDB2RouterReqsName)
		registry.routerReqsTLSCounter = newInfluxDB2Counter(influxDB2RouterReqsTLSName)
		registry.routerReqDurationHistogram, _ = NewHistogramWithScale(newInfluxDB2Histogram(influxDB2RouterReqsDurationName), time.Second)
		registry.routerOpenConnsGauge = newInfluxDB2Gauge(influxDB2RouterOpenConnsName)
	}

	if config.AddServicesLabels {
		registry.svcEnabled = config.AddServicesLabels
		registry.serviceReqsCounter = newInfluxDB2Counter(influxDB2MetricsServiceReqsName)
		registry.serviceReqsTLSCounter = newInfluxDB2Counter(influxDB2MetricsServiceReqsTLSName)
		registry.serviceReqDurationHistogram, _ = NewHistogramWithScale(newInfluxDB2Histogram(influxDB2MetricsServiceLatencyName), time.Second)
		registry.serviceRetriesCounter = newInfluxDB2Counter(influxDB2RetriesTotalName)
		registry.serviceOpenConnsGauge = newInfluxDB2Gauge(influxDB2OpenConnsName)
		registry.serviceServerUpGauge = newInfluxDB2Gauge(influxDB2ServerUpName)
	}

	return registry
}

// StopInfluxDB2 flushes and removes InfluxDB2 client and WriteAPI.
func StopInfluxDB2() {
	if influxDB2WriteAPI != nil {
		influxDB2WriteAPI.Flush()
	}
	if influxDB2Client != nil {
		influxDB2Client.Close()
	}
	influxDB2WriteAPI = nil
	influxDB2Client = nil
}

func sendInfluxDB2(name string, labels []string, value interface{}) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	for i := range labels {
		if i%2 != 0 {
			continue
		} else if i+1 >= len(labels) {
			break
		}
		tags[labels[i]] = labels[i+1]
	}
	fields[name] = value
	p := influxdb2.NewPoint("traefik", tags, fields, time.Now())
	influxDB2WriteAPI.WritePoint(p)
}

type influxDB2Counter struct {
	c *generic.Counter
}

func newInfluxDB2Counter(name string) *influxDB2Counter {
	return &influxDB2Counter{
		c: generic.NewCounter(name),
	}
}

// With returns a new influxDB2Counter with the given labels.
func (c *influxDB2Counter) With(labels ...string) metrics.Counter {
	newCounter := c.c.With(labels...).(*generic.Counter)
	newCounter.ValueReset()

	return &influxDB2Counter{
		c: newCounter,
	}
}

// Add adds the given delta to the counter.
func (c *influxDB2Counter) Add(delta float64) {
	c.c.Add(delta)
	sendInfluxDB2(c.c.Name, c.c.LabelValues(), c.c.Value())
}

type influxDB2Gauge struct {
	g *generic.Gauge
}

func newInfluxDB2Gauge(name string) *influxDB2Gauge {
	return &influxDB2Gauge{
		g: generic.NewGauge(name),
	}
}

// With returns a new pilotGauge with the given labels.
func (g *influxDB2Gauge) With(labels ...string) metrics.Gauge {
	newGauge := g.g.With(labels...).(*generic.Gauge)
	newGauge.Set(0)

	return &influxDB2Gauge{
		g: newGauge,
	}
}

// Set sets the given value to the gauge.
func (g *influxDB2Gauge) Set(value float64) {
	g.g.Set(value)
	sendInfluxDB2(g.g.Name, g.g.LabelValues(), g.g.Value())
}

// Add adds the given delta to the gauge.
func (g *influxDB2Gauge) Add(delta float64) {
	g.g.Add(delta)
	sendInfluxDB2(g.g.Name, g.g.LabelValues(), g.g.Value())
}

type influxDB2Histogram struct {
	g *generic.Gauge
}

func newInfluxDB2Histogram(name string) *influxDB2Histogram {
	return &influxDB2Histogram{
		g: generic.NewGauge(name),
	}
}

// With returns a new influxDB2Histogram with the given labels.
func (h *influxDB2Histogram) With(labels ...string) metrics.Histogram {
	newGauge := h.g.With(labels...).(*generic.Gauge)
	newGauge.Set(0)

	return &influxDB2Histogram{
		g: newGauge,
	}
}

// Observe records a new value into the histogram.
func (h *influxDB2Histogram) Observe(value float64) {
	h.g.Set(value)
	sendInfluxDB2(h.g.Name, h.g.LabelValues(), value)
}
