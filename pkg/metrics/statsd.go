package metrics

import (
	"context"
	"time"

	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/safe"
	"github.com/containous/traefik/pkg/types"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/statsd"
)

var statsdClient = statsd.New("traefik.", kitlog.LoggerFunc(func(keyvals ...interface{}) error {
	log.WithoutContext().WithField(log.MetricsProviderName, "statsd").Info(keyvals)
	return nil
}))

var statsdTicker *time.Ticker

const (
	statsdMetricsServiceReqsName      = "service.request.total"
	statsdMetricsServiceLatencyName   = "service.request.duration"
	statsdRetriesTotalName            = "service.retries.total"
	statsdConfigReloadsName           = "config.reload.total"
	statsdConfigReloadsFailureName    = statsdConfigReloadsName + ".failure"
	statsdLastConfigReloadSuccessName = "config.reload.lastSuccessTimestamp"
	statsdLastConfigReloadFailureName = "config.reload.lastFailureTimestamp"
	statsdEntryPointReqsName          = "entrypoint.request.total"
	statsdEntryPointReqDurationName   = "entrypoint.request.duration"
	statsdEntryPointOpenConnsName     = "entrypoint.connections.open"
	statsdOpenConnsName               = "service.connections.open"
	statsdServerUpName                = "service.server.up"
)

// RegisterStatsd registers the metrics pusher if this didn't happen yet and creates a statsd Registry instance.
func RegisterStatsd(ctx context.Context, config *types.Statsd) Registry {
	if statsdTicker == nil {
		statsdTicker = initStatsdTicker(ctx, config)
	}

	registry := &standardRegistry{
		configReloadsCounter:         statsdClient.NewCounter(statsdConfigReloadsName, 1.0),
		configReloadsFailureCounter:  statsdClient.NewCounter(statsdConfigReloadsFailureName, 1.0),
		lastConfigReloadSuccessGauge: statsdClient.NewGauge(statsdLastConfigReloadSuccessName),
		lastConfigReloadFailureGauge: statsdClient.NewGauge(statsdLastConfigReloadFailureName),
	}

	if config.OnEntryPoints {
		registry.epEnabled = config.OnEntryPoints
		registry.entryPointReqsCounter = statsdClient.NewCounter(statsdEntryPointReqsName, 1.0)
		registry.entryPointReqDurationHistogram = statsdClient.NewTiming(statsdEntryPointReqDurationName, 1.0)
		registry.entryPointOpenConnsGauge = statsdClient.NewGauge(statsdEntryPointOpenConnsName)
	}

	if config.OnServices {
		registry.svcEnabled = config.OnServices
		registry.serviceReqsCounter = statsdClient.NewCounter(statsdMetricsServiceReqsName, 1.0)
		registry.serviceReqDurationHistogram = statsdClient.NewTiming(statsdMetricsServiceLatencyName, 1.0)
		registry.serviceRetriesCounter = statsdClient.NewCounter(statsdRetriesTotalName, 1.0)
		registry.serviceOpenConnsGauge = statsdClient.NewGauge(statsdOpenConnsName)
		registry.serviceServerUpGauge = statsdClient.NewGauge(statsdServerUpName)
	}

	return registry
}

// initStatsdTicker initializes metrics pusher and creates a statsdClient if not created already
func initStatsdTicker(ctx context.Context, config *types.Statsd) *time.Ticker {
	address := config.Address
	if len(address) == 0 {
		address = "localhost:8125"
	}

	report := time.NewTicker(time.Duration(config.PushInterval))

	safe.Go(func() {
		statsdClient.SendLoop(ctx, report.C, "udp", address)
	})

	return report
}

// StopStatsd stops internal statsdTicker which controls the pushing of metrics to StatsD Agent and resets it to `nil`
func StopStatsd() {
	if statsdTicker != nil {
		statsdTicker.Stop()
	}
	statsdTicker = nil
}
