package metrics

import (
	"errors"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/multi"
)

const defaultMetricsPrefix = "traefik"

// Registry has to implemented by any system that wants to monitor and expose metrics.
type Registry interface {
	// IsEpEnabled shows whether metrics instrumentation is enabled on entry points.
	IsEpEnabled() bool
	// IsRouterEnabled shows whether metrics instrumentation is enabled on routers.
	IsRouterEnabled() bool
	// IsSvcEnabled shows whether metrics instrumentation is enabled on services.
	IsSvcEnabled() bool

	// server metrics

	ConfigReloadsCounter() metrics.Counter
	ConfigReloadsFailureCounter() metrics.Counter
	LastConfigReloadSuccessGauge() metrics.Gauge
	LastConfigReloadFailureGauge() metrics.Gauge

	// TLS

	TLSCertsNotAfterTimestampGauge() metrics.Gauge

	// entry point metrics

	EntryPointReqsCounter() CounterWithHeaders
	EntryPointReqsTLSCounter() metrics.Counter
	EntryPointReqDurationHistogram() ScalableHistogram
	EntryPointOpenConnsGauge() metrics.Gauge
	EntryPointReqsBytesCounter() metrics.Counter
	EntryPointRespsBytesCounter() metrics.Counter

	// router metrics

	RouterReqsCounter() CounterWithHeaders
	RouterReqsTLSCounter() metrics.Counter
	RouterReqDurationHistogram() ScalableHistogram
	RouterOpenConnsGauge() metrics.Gauge
	RouterReqsBytesCounter() metrics.Counter
	RouterRespsBytesCounter() metrics.Counter

	// service metrics

	ServiceReqsCounter() CounterWithHeaders
	ServiceReqsTLSCounter() metrics.Counter
	ServiceReqDurationHistogram() ScalableHistogram
	ServiceOpenConnsGauge() metrics.Gauge
	ServiceRetriesCounter() metrics.Counter
	ServiceServerUpGauge() metrics.Gauge
	ServiceReqsBytesCounter() metrics.Counter
	ServiceRespsBytesCounter() metrics.Counter
}

// NewVoidRegistry is a noop implementation of metrics.Registry.
// It is used to avoid nil checking in components that do metric collections.
func NewVoidRegistry() Registry {
	return NewMultiRegistry([]Registry{})
}

// NewMultiRegistry is an implementation of metrics.Registry that wraps multiple registries.
// It handles the case when a registry hasn't registered some metric and returns nil.
// This allows for feature disparity between the different metric implementations.
func NewMultiRegistry(registries []Registry) Registry {
	var configReloadsCounter []metrics.Counter
	var configReloadsFailureCounter []metrics.Counter
	var lastConfigReloadSuccessGauge []metrics.Gauge
	var lastConfigReloadFailureGauge []metrics.Gauge
	var tlsCertsNotAfterTimestampGauge []metrics.Gauge
	var entryPointReqsCounter []CounterWithHeaders
	var entryPointReqsTLSCounter []metrics.Counter
	var entryPointReqDurationHistogram []ScalableHistogram
	var entryPointOpenConnsGauge []metrics.Gauge
	var entryPointReqsBytesCounter []metrics.Counter
	var entryPointRespsBytesCounter []metrics.Counter
	var routerReqsCounter []CounterWithHeaders
	var routerReqsTLSCounter []metrics.Counter
	var routerReqDurationHistogram []ScalableHistogram
	var routerOpenConnsGauge []metrics.Gauge
	var routerReqsBytesCounter []metrics.Counter
	var routerRespsBytesCounter []metrics.Counter
	var serviceReqsCounter []CounterWithHeaders
	var serviceReqsTLSCounter []metrics.Counter
	var serviceReqDurationHistogram []ScalableHistogram
	var serviceOpenConnsGauge []metrics.Gauge
	var serviceRetriesCounter []metrics.Counter
	var serviceServerUpGauge []metrics.Gauge
	var serviceReqsBytesCounter []metrics.Counter
	var serviceRespsBytesCounter []metrics.Counter

	for _, r := range registries {
		if r.ConfigReloadsCounter() != nil {
			configReloadsCounter = append(configReloadsCounter, r.ConfigReloadsCounter())
		}
		if r.ConfigReloadsFailureCounter() != nil {
			configReloadsFailureCounter = append(configReloadsFailureCounter, r.ConfigReloadsFailureCounter())
		}
		if r.LastConfigReloadSuccessGauge() != nil {
			lastConfigReloadSuccessGauge = append(lastConfigReloadSuccessGauge, r.LastConfigReloadSuccessGauge())
		}
		if r.LastConfigReloadFailureGauge() != nil {
			lastConfigReloadFailureGauge = append(lastConfigReloadFailureGauge, r.LastConfigReloadFailureGauge())
		}
		if r.TLSCertsNotAfterTimestampGauge() != nil {
			tlsCertsNotAfterTimestampGauge = append(tlsCertsNotAfterTimestampGauge, r.TLSCertsNotAfterTimestampGauge())
		}
		if r.EntryPointReqsCounter() != nil {
			entryPointReqsCounter = append(entryPointReqsCounter, r.EntryPointReqsCounter())
		}
		if r.EntryPointReqsTLSCounter() != nil {
			entryPointReqsTLSCounter = append(entryPointReqsTLSCounter, r.EntryPointReqsTLSCounter())
		}
		if r.EntryPointReqDurationHistogram() != nil {
			entryPointReqDurationHistogram = append(entryPointReqDurationHistogram, r.EntryPointReqDurationHistogram())
		}
		if r.EntryPointOpenConnsGauge() != nil {
			entryPointOpenConnsGauge = append(entryPointOpenConnsGauge, r.EntryPointOpenConnsGauge())
		}
		if r.EntryPointReqsBytesCounter() != nil {
			entryPointReqsBytesCounter = append(entryPointReqsBytesCounter, r.EntryPointReqsBytesCounter())
		}
		if r.EntryPointRespsBytesCounter() != nil {
			entryPointRespsBytesCounter = append(entryPointRespsBytesCounter, r.EntryPointRespsBytesCounter())
		}
		if r.RouterReqsCounter() != nil {
			routerReqsCounter = append(routerReqsCounter, r.RouterReqsCounter())
		}
		if r.RouterReqsTLSCounter() != nil {
			routerReqsTLSCounter = append(routerReqsTLSCounter, r.RouterReqsTLSCounter())
		}
		if r.RouterReqDurationHistogram() != nil {
			routerReqDurationHistogram = append(routerReqDurationHistogram, r.RouterReqDurationHistogram())
		}
		if r.RouterOpenConnsGauge() != nil {
			routerOpenConnsGauge = append(routerOpenConnsGauge, r.RouterOpenConnsGauge())
		}
		if r.RouterReqsBytesCounter() != nil {
			routerReqsBytesCounter = append(routerReqsBytesCounter, r.RouterReqsBytesCounter())
		}
		if r.RouterRespsBytesCounter() != nil {
			routerRespsBytesCounter = append(routerRespsBytesCounter, r.RouterRespsBytesCounter())
		}
		if r.ServiceReqsCounter() != nil {
			serviceReqsCounter = append(serviceReqsCounter, r.ServiceReqsCounter())
		}
		if r.ServiceReqsTLSCounter() != nil {
			serviceReqsTLSCounter = append(serviceReqsTLSCounter, r.ServiceReqsTLSCounter())
		}
		if r.ServiceReqDurationHistogram() != nil {
			serviceReqDurationHistogram = append(serviceReqDurationHistogram, r.ServiceReqDurationHistogram())
		}
		if r.ServiceOpenConnsGauge() != nil {
			serviceOpenConnsGauge = append(serviceOpenConnsGauge, r.ServiceOpenConnsGauge())
		}
		if r.ServiceRetriesCounter() != nil {
			serviceRetriesCounter = append(serviceRetriesCounter, r.ServiceRetriesCounter())
		}
		if r.ServiceServerUpGauge() != nil {
			serviceServerUpGauge = append(serviceServerUpGauge, r.ServiceServerUpGauge())
		}
		if r.ServiceReqsBytesCounter() != nil {
			serviceReqsBytesCounter = append(serviceReqsBytesCounter, r.ServiceReqsBytesCounter())
		}
		if r.ServiceRespsBytesCounter() != nil {
			serviceRespsBytesCounter = append(serviceRespsBytesCounter, r.ServiceRespsBytesCounter())
		}
	}

	return &standardRegistry{
		epEnabled:                      len(entryPointReqsCounter) > 0 || len(entryPointReqDurationHistogram) > 0 || len(entryPointOpenConnsGauge) > 0,
		svcEnabled:                     len(serviceReqsCounter) > 0 || len(serviceReqDurationHistogram) > 0 || len(serviceOpenConnsGauge) > 0 || len(serviceRetriesCounter) > 0 || len(serviceServerUpGauge) > 0,
		routerEnabled:                  len(routerReqsCounter) > 0 || len(routerReqDurationHistogram) > 0 || len(routerOpenConnsGauge) > 0,
		configReloadsCounter:           multi.NewCounter(configReloadsCounter...),
		configReloadsFailureCounter:    multi.NewCounter(configReloadsFailureCounter...),
		lastConfigReloadSuccessGauge:   multi.NewGauge(lastConfigReloadSuccessGauge...),
		lastConfigReloadFailureGauge:   multi.NewGauge(lastConfigReloadFailureGauge...),
		tlsCertsNotAfterTimestampGauge: multi.NewGauge(tlsCertsNotAfterTimestampGauge...),
		entryPointReqsCounter:          NewMultiCounterWithHeaders(entryPointReqsCounter...),
		entryPointReqsTLSCounter:       multi.NewCounter(entryPointReqsTLSCounter...),
		entryPointReqDurationHistogram: MultiHistogram(entryPointReqDurationHistogram),
		entryPointOpenConnsGauge:       multi.NewGauge(entryPointOpenConnsGauge...),
		entryPointReqsBytesCounter:     multi.NewCounter(entryPointReqsBytesCounter...),
		entryPointRespsBytesCounter:    multi.NewCounter(entryPointRespsBytesCounter...),
		routerReqsCounter:              NewMultiCounterWithHeaders(routerReqsCounter...),
		routerReqsTLSCounter:           multi.NewCounter(routerReqsTLSCounter...),
		routerReqDurationHistogram:     MultiHistogram(routerReqDurationHistogram),
		routerOpenConnsGauge:           multi.NewGauge(routerOpenConnsGauge...),
		routerReqsBytesCounter:         multi.NewCounter(routerReqsBytesCounter...),
		routerRespsBytesCounter:        multi.NewCounter(routerRespsBytesCounter...),
		serviceReqsCounter:             NewMultiCounterWithHeaders(serviceReqsCounter...),
		serviceReqsTLSCounter:          multi.NewCounter(serviceReqsTLSCounter...),
		serviceReqDurationHistogram:    MultiHistogram(serviceReqDurationHistogram),
		serviceOpenConnsGauge:          multi.NewGauge(serviceOpenConnsGauge...),
		serviceRetriesCounter:          multi.NewCounter(serviceRetriesCounter...),
		serviceServerUpGauge:           multi.NewGauge(serviceServerUpGauge...),
		serviceReqsBytesCounter:        multi.NewCounter(serviceReqsBytesCounter...),
		serviceRespsBytesCounter:       multi.NewCounter(serviceRespsBytesCounter...),
	}
}

type standardRegistry struct {
	epEnabled                      bool
	routerEnabled                  bool
	svcEnabled                     bool
	configReloadsCounter           metrics.Counter
	configReloadsFailureCounter    metrics.Counter
	lastConfigReloadSuccessGauge   metrics.Gauge
	lastConfigReloadFailureGauge   metrics.Gauge
	tlsCertsNotAfterTimestampGauge metrics.Gauge
	entryPointReqsCounter          CounterWithHeaders
	entryPointReqsTLSCounter       metrics.Counter
	entryPointReqDurationHistogram ScalableHistogram
	entryPointOpenConnsGauge       metrics.Gauge
	entryPointReqsBytesCounter     metrics.Counter
	entryPointRespsBytesCounter    metrics.Counter
	routerReqsCounter              CounterWithHeaders
	routerReqsTLSCounter           metrics.Counter
	routerReqDurationHistogram     ScalableHistogram
	routerOpenConnsGauge           metrics.Gauge
	routerReqsBytesCounter         metrics.Counter
	routerRespsBytesCounter        metrics.Counter
	serviceReqsCounter             CounterWithHeaders
	serviceReqsTLSCounter          metrics.Counter
	serviceReqDurationHistogram    ScalableHistogram
	serviceOpenConnsGauge          metrics.Gauge
	serviceRetriesCounter          metrics.Counter
	serviceServerUpGauge           metrics.Gauge
	serviceReqsBytesCounter        metrics.Counter
	serviceRespsBytesCounter       metrics.Counter
}

func (r *standardRegistry) IsEpEnabled() bool {
	return r.epEnabled
}

func (r *standardRegistry) IsRouterEnabled() bool {
	return r.routerEnabled
}

func (r *standardRegistry) IsSvcEnabled() bool {
	return r.svcEnabled
}

func (r *standardRegistry) ConfigReloadsCounter() metrics.Counter {
	return r.configReloadsCounter
}

func (r *standardRegistry) ConfigReloadsFailureCounter() metrics.Counter {
	return r.configReloadsFailureCounter
}

func (r *standardRegistry) LastConfigReloadSuccessGauge() metrics.Gauge {
	return r.lastConfigReloadSuccessGauge
}

func (r *standardRegistry) LastConfigReloadFailureGauge() metrics.Gauge {
	return r.lastConfigReloadFailureGauge
}

func (r *standardRegistry) TLSCertsNotAfterTimestampGauge() metrics.Gauge {
	return r.tlsCertsNotAfterTimestampGauge
}

func (r *standardRegistry) EntryPointReqsCounter() CounterWithHeaders {
	return r.entryPointReqsCounter
}

func (r *standardRegistry) EntryPointReqsTLSCounter() metrics.Counter {
	return r.entryPointReqsTLSCounter
}

func (r *standardRegistry) EntryPointReqDurationHistogram() ScalableHistogram {
	return r.entryPointReqDurationHistogram
}

func (r *standardRegistry) EntryPointOpenConnsGauge() metrics.Gauge {
	return r.entryPointOpenConnsGauge
}

func (r *standardRegistry) EntryPointReqsBytesCounter() metrics.Counter {
	return r.entryPointReqsBytesCounter
}

func (r *standardRegistry) EntryPointRespsBytesCounter() metrics.Counter {
	return r.entryPointRespsBytesCounter
}

func (r *standardRegistry) RouterReqsCounter() CounterWithHeaders {
	return r.routerReqsCounter
}

func (r *standardRegistry) RouterReqsTLSCounter() metrics.Counter {
	return r.routerReqsTLSCounter
}

func (r *standardRegistry) RouterReqDurationHistogram() ScalableHistogram {
	return r.routerReqDurationHistogram
}

func (r *standardRegistry) RouterOpenConnsGauge() metrics.Gauge {
	return r.routerOpenConnsGauge
}

func (r *standardRegistry) RouterReqsBytesCounter() metrics.Counter {
	return r.routerReqsBytesCounter
}

func (r *standardRegistry) RouterRespsBytesCounter() metrics.Counter {
	return r.routerRespsBytesCounter
}

func (r *standardRegistry) ServiceReqsCounter() CounterWithHeaders {
	return r.serviceReqsCounter
}

func (r *standardRegistry) ServiceReqsTLSCounter() metrics.Counter {
	return r.serviceReqsTLSCounter
}

func (r *standardRegistry) ServiceReqDurationHistogram() ScalableHistogram {
	return r.serviceReqDurationHistogram
}

func (r *standardRegistry) ServiceOpenConnsGauge() metrics.Gauge {
	return r.serviceOpenConnsGauge
}

func (r *standardRegistry) ServiceRetriesCounter() metrics.Counter {
	return r.serviceRetriesCounter
}

func (r *standardRegistry) ServiceServerUpGauge() metrics.Gauge {
	return r.serviceServerUpGauge
}

func (r *standardRegistry) ServiceReqsBytesCounter() metrics.Counter {
	return r.serviceReqsBytesCounter
}

func (r *standardRegistry) ServiceRespsBytesCounter() metrics.Counter {
	return r.serviceRespsBytesCounter
}

// ScalableHistogram is a Histogram with a predefined time unit,
// used when producing observations without explicitly setting the observed value.
type ScalableHistogram interface {
	With(labelValues ...string) ScalableHistogram
	Observe(v float64)
	ObserveFromStart(start time.Time)
}

// HistogramWithScale is a histogram that will convert its observed value to the specified unit.
type HistogramWithScale struct {
	histogram metrics.Histogram
	unit      time.Duration
}

// With implements ScalableHistogram.
func (s *HistogramWithScale) With(labelValues ...string) ScalableHistogram {
	h, _ := NewHistogramWithScale(s.histogram.With(labelValues...), s.unit)
	return h
}

// ObserveFromStart implements ScalableHistogram.
func (s *HistogramWithScale) ObserveFromStart(start time.Time) {
	if s.unit <= 0 {
		return
	}

	d := float64(time.Since(start).Nanoseconds()) / float64(s.unit)
	if d < 0 {
		d = 0
	}
	s.histogram.Observe(d)
}

// Observe implements ScalableHistogram.
func (s *HistogramWithScale) Observe(v float64) {
	s.histogram.Observe(v)
}

// NewHistogramWithScale returns a ScalableHistogram. It returns an error if the given unit is <= 0.
func NewHistogramWithScale(histogram metrics.Histogram, unit time.Duration) (ScalableHistogram, error) {
	if unit <= 0 {
		return nil, errors.New("invalid time unit")
	}
	return &HistogramWithScale{
		histogram: histogram,
		unit:      unit,
	}, nil
}

// MultiHistogram collects multiple individual histograms and treats them as a unit.
type MultiHistogram []ScalableHistogram

// ObserveFromStart implements ScalableHistogram.
func (h MultiHistogram) ObserveFromStart(start time.Time) {
	for _, histogram := range h {
		histogram.ObserveFromStart(start)
	}
}

// Observe implements ScalableHistogram.
func (h MultiHistogram) Observe(v float64) {
	for _, histogram := range h {
		histogram.Observe(v)
	}
}

// With implements ScalableHistogram.
func (h MultiHistogram) With(labelValues ...string) ScalableHistogram {
	next := make(MultiHistogram, len(h))
	for i := range h {
		next[i] = h[i].With(labelValues...)
	}
	return next
}
