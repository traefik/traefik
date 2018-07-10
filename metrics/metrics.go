package metrics

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/multi"
)

// Registry has to implemented by any system that wants to monitor and expose metrics.
type Registry interface {
	// IsEnabled shows whether metrics instrumentation is enabled.
	IsEnabled() bool

	// server metrics
	ConfigReloadsCounter() metrics.Counter
	ConfigReloadsFailureCounter() metrics.Counter
	LastConfigReloadSuccessGauge() metrics.Gauge
	LastConfigReloadFailureGauge() metrics.Gauge

	// entry point metrics
	EntrypointReqsCounter() metrics.Counter
	EntrypointReqDurationHistogram() metrics.Histogram
	EntrypointOpenConnsGauge() metrics.Gauge

	// backend metrics
	BackendReqsCounter() metrics.Counter
	BackendReqDurationHistogram() metrics.Histogram
	BackendOpenConnsGauge() metrics.Gauge
	BackendRetriesCounter() metrics.Counter
	BackendServerUpGauge() metrics.Gauge
}

// NewVoidRegistry is a noop implementation of metrics.Registry.
// It is used to avoid nil checking in components that do metric collections.
func NewVoidRegistry() Registry {
	return NewMultiRegistry([]Registry{})
}

// NewMultiRegistry is an implementation of metrics.Registry that wraps multiple registries.
// It handles the case when a registry hasn't registered some metric and returns nil.
// This allows for feature imparity between the different metric implementations.
func NewMultiRegistry(registries []Registry) Registry {
	var configReloadsCounter []metrics.Counter
	var configReloadsFailureCounter []metrics.Counter
	var lastConfigReloadSuccessGauge []metrics.Gauge
	var lastConfigReloadFailureGauge []metrics.Gauge
	var entrypointReqsCounter []metrics.Counter
	var entrypointReqDurationHistogram []metrics.Histogram
	var entrypointOpenConnsGauge []metrics.Gauge
	var backendReqsCounter []metrics.Counter
	var backendReqDurationHistogram []metrics.Histogram
	var backendOpenConnsGauge []metrics.Gauge
	var backendRetriesCounter []metrics.Counter
	var backendServerUpGauge []metrics.Gauge

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
		if r.EntrypointReqsCounter() != nil {
			entrypointReqsCounter = append(entrypointReqsCounter, r.EntrypointReqsCounter())
		}
		if r.EntrypointReqDurationHistogram() != nil {
			entrypointReqDurationHistogram = append(entrypointReqDurationHistogram, r.EntrypointReqDurationHistogram())
		}
		if r.EntrypointOpenConnsGauge() != nil {
			entrypointOpenConnsGauge = append(entrypointOpenConnsGauge, r.EntrypointOpenConnsGauge())
		}
		if r.BackendReqsCounter() != nil {
			backendReqsCounter = append(backendReqsCounter, r.BackendReqsCounter())
		}
		if r.BackendReqDurationHistogram() != nil {
			backendReqDurationHistogram = append(backendReqDurationHistogram, r.BackendReqDurationHistogram())
		}
		if r.BackendOpenConnsGauge() != nil {
			backendOpenConnsGauge = append(backendOpenConnsGauge, r.BackendOpenConnsGauge())
		}
		if r.BackendRetriesCounter() != nil {
			backendRetriesCounter = append(backendRetriesCounter, r.BackendRetriesCounter())
		}
		if r.BackendServerUpGauge() != nil {
			backendServerUpGauge = append(backendServerUpGauge, r.BackendServerUpGauge())
		}
	}

	return &standardRegistry{
		enabled:                        len(registries) > 0,
		configReloadsCounter:           multi.NewCounter(configReloadsCounter...),
		configReloadsFailureCounter:    multi.NewCounter(configReloadsFailureCounter...),
		lastConfigReloadSuccessGauge:   multi.NewGauge(lastConfigReloadSuccessGauge...),
		lastConfigReloadFailureGauge:   multi.NewGauge(lastConfigReloadFailureGauge...),
		entrypointReqsCounter:          multi.NewCounter(entrypointReqsCounter...),
		entrypointReqDurationHistogram: multi.NewHistogram(entrypointReqDurationHistogram...),
		entrypointOpenConnsGauge:       multi.NewGauge(entrypointOpenConnsGauge...),
		backendReqsCounter:             multi.NewCounter(backendReqsCounter...),
		backendReqDurationHistogram:    multi.NewHistogram(backendReqDurationHistogram...),
		backendOpenConnsGauge:          multi.NewGauge(backendOpenConnsGauge...),
		backendRetriesCounter:          multi.NewCounter(backendRetriesCounter...),
		backendServerUpGauge:           multi.NewGauge(backendServerUpGauge...),
	}
}

type standardRegistry struct {
	enabled                        bool
	configReloadsCounter           metrics.Counter
	configReloadsFailureCounter    metrics.Counter
	lastConfigReloadSuccessGauge   metrics.Gauge
	lastConfigReloadFailureGauge   metrics.Gauge
	entrypointReqsCounter          metrics.Counter
	entrypointReqDurationHistogram metrics.Histogram
	entrypointOpenConnsGauge       metrics.Gauge
	backendReqsCounter             metrics.Counter
	backendReqDurationHistogram    metrics.Histogram
	backendOpenConnsGauge          metrics.Gauge
	backendRetriesCounter          metrics.Counter
	backendServerUpGauge           metrics.Gauge
}

func (r *standardRegistry) IsEnabled() bool {
	return r.enabled
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

func (r *standardRegistry) EntrypointReqsCounter() metrics.Counter {
	return r.entrypointReqsCounter
}

func (r *standardRegistry) EntrypointReqDurationHistogram() metrics.Histogram {
	return r.entrypointReqDurationHistogram
}

func (r *standardRegistry) EntrypointOpenConnsGauge() metrics.Gauge {
	return r.entrypointOpenConnsGauge
}

func (r *standardRegistry) BackendReqsCounter() metrics.Counter {
	return r.backendReqsCounter
}

func (r *standardRegistry) BackendReqDurationHistogram() metrics.Histogram {
	return r.backendReqDurationHistogram
}

func (r *standardRegistry) BackendOpenConnsGauge() metrics.Gauge {
	return r.backendOpenConnsGauge
}

func (r *standardRegistry) BackendRetriesCounter() metrics.Counter {
	return r.backendRetriesCounter
}

func (r *standardRegistry) BackendServerUpGauge() metrics.Gauge {
	return r.backendServerUpGauge
}
