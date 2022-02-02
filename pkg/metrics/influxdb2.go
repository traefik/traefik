package metrics

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/generic"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	iLog "github.com/influxdata/influxdb-client-go/v2/log"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/types"
)

var (
	influxDB2Client   influxdb2.Client
	influxDB2WriteAPI api.WriteAPI
)

// RegisterInfluxDB2 creates metrics exporter for InfluxDB2.
func RegisterInfluxDB2(ctx context.Context, config *types.InfluxDB2) Registry {
	if influxDB2Client == nil {
		if err := initInfluxDB2Client(config); err != nil {
			log.FromContext(ctx).Error(err)
		}

		if err := initInfluxDB2WriteAPI(ctx, config.Org, config.Bucket); err != nil {
			log.FromContext(ctx).Error(err)
		}
	}

	registry := &standardRegistry{
		configReloadsCounter:           newInfluxDB2Counter(influxDBConfigReloadsName),
		configReloadsFailureCounter:    newInfluxDB2Counter(influxDBConfigReloadsFailureName),
		lastConfigReloadSuccessGauge:   newInfluxDB2Gauge(influxDBLastConfigReloadSuccessName),
		lastConfigReloadFailureGauge:   newInfluxDB2Gauge(influxDBLastConfigReloadFailureName),
		tlsCertsNotAfterTimestampGauge: newInfluxDB2Gauge(influxDBTLSCertsNotAfterTimestampName),
	}

	if config.AddEntryPointsLabels {
		registry.epEnabled = config.AddEntryPointsLabels
		registry.entryPointReqsCounter = newInfluxDB2Counter(influxDBEntryPointReqsName)
		registry.entryPointReqsTLSCounter = newInfluxDB2Counter(influxDBEntryPointReqsTLSName)
		registry.entryPointReqDurationHistogram, _ = NewHistogramWithScale(newInfluxDB2Histogram(influxDBEntryPointReqDurationName), time.Second)
		registry.entryPointOpenConnsGauge = newInfluxDB2Gauge(influxDBEntryPointOpenConnsName)
	}

	if config.AddRoutersLabels {
		registry.routerEnabled = config.AddRoutersLabels
		registry.routerReqsCounter = newInfluxDB2Counter(influxDBRouterReqsName)
		registry.routerReqsTLSCounter = newInfluxDB2Counter(influxDBRouterReqsTLSName)
		registry.routerReqDurationHistogram, _ = NewHistogramWithScale(newInfluxDB2Histogram(influxDBRouterReqsDurationName), time.Second)
		registry.routerOpenConnsGauge = newInfluxDB2Gauge(influxDBORouterOpenConnsName)
	}

	if config.AddServicesLabels {
		registry.svcEnabled = config.AddServicesLabels
		registry.serviceReqsCounter = newInfluxDB2Counter(influxDBServiceReqsName)
		registry.serviceReqsTLSCounter = newInfluxDB2Counter(influxDBServiceReqsTLSName)
		registry.serviceReqDurationHistogram, _ = NewHistogramWithScale(newInfluxDB2Histogram(influxDBServiceReqsDurationName), time.Second)
		registry.serviceRetriesCounter = newInfluxDB2Counter(influxDBServiceRetriesTotalName)
		registry.serviceOpenConnsGauge = newInfluxDB2Gauge(influxDBServiceOpenConnsName)
		registry.serviceServerUpGauge = newInfluxDB2Gauge(influxDBServiceServerUpName)
	}

	return registry
}

// initInfluxDB2Client creates a influxDBClient.
func initInfluxDB2Client(config *types.InfluxDB2) error {
	if influxDB2Client != nil {
		return nil
	}

	if config.Token == "" || config.Org == "" || config.Bucket == "" {
		return errors.New("token, org or bucket properties are missing")
	}

	if config.BatchSize <= 0 {
		return errors.New("batch size must be strictly greater than zero")
	}

	flushMs := uint(time.Duration(config.PushInterval).Milliseconds())
	options := influxdb2.DefaultOptions()
	options.SetBatchSize(uint(config.BatchSize))
	options.SetFlushInterval(flushMs)
	influxDB2Client = influxdb2.NewClientWithOptions(config.Address, config.Token, options)

	return nil
}

// initInfluxDB2WriteAPI creates a influxDBClient.
func initInfluxDB2WriteAPI(ctx context.Context, org, bucket string) error {
	if influxDB2Client == nil {
		return errors.New("cannot initialize write API without client")
	}

	influxDB2WriteAPI = influxDB2Client.WriteAPI(org, bucket)

	iLog.Log = nil // Disable influxDB2 internal logs in favor of internal logger
	go func() {
		for {
			if influxDB2WriteAPI == nil {
				return
			}

			select {
			case <-ctx.Done():
				return
			case err := <-influxDB2WriteAPI.Errors():
				if err != nil {
					log.FromContext(ctx).Error(err)
				}
			}
		}
	}()

	return nil
}

// StopInfluxDB2 flushes and removes InfluxDB2 client and WriteAPI.
func StopInfluxDB2() {
	if influxDB2WriteAPI != nil {
		influxDB2WriteAPI.Flush()
	}
	influxDB2WriteAPI = nil

	if influxDB2Client != nil {
		influxDB2Client.Close()
	}
	influxDB2Client = nil
}

func sendInfluxDB2(name string, labels []string, value interface{}) {
	if influxDB2WriteAPI == nil {
		return
	}

	point := influxdb2.NewPointWithMeasurement("traefik")

	for i := 0; i < len(labels); i += 2 { // sets pairs of labels as tags
		point.AddTag(labels[i], labels[i+1])
	}

	influxDB2WriteAPI.WritePoint(point.AddField(name, value).SetTime(time.Now()))
}

type influxDB2Counter struct {
	c        *generic.Counter
	counters *sync.Map
}

func newInfluxDB2Counter(name string) *influxDB2Counter {
	return &influxDB2Counter{
		c:        generic.NewCounter(name),
		counters: &sync.Map{},
	}
}

// With returns a new influxDB2Counter with the given labels.
func (c *influxDB2Counter) With(labels ...string) metrics.Counter {
	newCounter := c.c.With(labels...).(*generic.Counter)
	newCounter.ValueReset()

	return &influxDB2Counter{
		c:        newCounter,
		counters: c.counters,
	}
}

// Add adds the given delta to the counter.
func (c *influxDB2Counter) Add(delta float64) {
	labelsKey := strings.Join(c.c.LabelValues(), ",")
	v, _ := c.counters.LoadOrStore(labelsKey, c)
	counter := v.(*influxDB2Counter)
	counter.c.Add(delta)

	sendInfluxDB2(counter.c.Name, counter.c.LabelValues(), counter.c.Value())
}

type influxDB2Gauge struct {
	g      *generic.Gauge
	gauges *sync.Map
}

func newInfluxDB2Gauge(name string) *influxDB2Gauge {
	return &influxDB2Gauge{
		g:      generic.NewGauge(name),
		gauges: &sync.Map{},
	}
}

// With returns a new pilotGauge with the given labels.
func (g *influxDB2Gauge) With(labels ...string) metrics.Gauge {
	newGauge := g.g.With(labels...).(*generic.Gauge)
	newGauge.Set(0)

	return &influxDB2Gauge{
		g:      newGauge,
		gauges: g.gauges,
	}
}

// Set sets the given value to the gauge.
func (g *influxDB2Gauge) Set(value float64) {
	labelsKey := strings.Join(g.g.LabelValues(), ",")
	v, _ := g.gauges.LoadOrStore(labelsKey, g)
	gauge := v.(*influxDB2Gauge)
	gauge.g.Set(value)

	sendInfluxDB2(gauge.g.Name, gauge.g.LabelValues(), value)
}

// Add adds the given delta to the gauge.
func (g *influxDB2Gauge) Add(delta float64) {
	labelsKey := strings.Join(g.g.LabelValues(), ",")
	v, _ := g.gauges.LoadOrStore(labelsKey, g)
	gauge := v.(*influxDB2Gauge)
	gauge.g.Add(delta)

	sendInfluxDB2(gauge.g.Name, gauge.g.LabelValues(), gauge.g.Value())
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
