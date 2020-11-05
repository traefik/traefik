package registry

import (
	"time"

	"github.com/go-kit/kit/metrics"
)

// ScalableHistogram is a Histogram with a predefined time unit,
// used when producing observations without explicitly setting the observed value.
type ScalableHistogram interface {
	With(labelValues ...string) ScalableHistogram
	Observe(v float64)
	ObserveFromStart(start time.Time)
}

// Registry has to implemented by any system that wants to monitor and expose metrics.
type Registry interface {
	// IsEpEnabled shows whether metrics instrumentation is enabled on entry points.
	IsEpEnabled() bool
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
	EntryPointReqsCounter() metrics.Counter
	EntryPointReqsTLSCounter() metrics.Counter
	EntryPointReqDurationHistogram() ScalableHistogram
	EntryPointOpenConnsGauge() metrics.Gauge

	// service metrics
	ServiceReqsCounter() metrics.Counter
	ServiceReqsTLSCounter() metrics.Counter
	ServiceReqDurationHistogram() ScalableHistogram
	ServiceOpenConnsGauge() metrics.Gauge
	ServiceRetriesCounter() metrics.Counter
	ServiceServerUpGauge() metrics.Gauge
}
