package metrics

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics/statsd"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/types"
)

var (
	statsdClient *statsd.Statsd
	statsdTicker *time.Ticker
)

const (
	statsdConfigReloadsName           = "config.reload.total"
	statsdLastConfigReloadSuccessName = "config.reload.lastSuccessTimestamp"
	statsdOpenConnectionsName         = "open.connections"

	statsdTLSCertsNotAfterTimestampName = "tls.certs.notAfterTimestamp"

	statsdEntryPointReqsName        = "entrypoint.request.total"
	statsdEntryPointReqsTLSName     = "entrypoint.request.tls.total"
	statsdEntryPointReqDurationName = "entrypoint.request.duration"
	statsdEntryPointReqsBytesName   = "entrypoint.requests.bytes.total"
	statsdEntryPointRespsBytesName  = "entrypoint.responses.bytes.total"

	statsdRouterReqsName         = "router.request.total"
	statsdRouterReqsTLSName      = "router.request.tls.total"
	statsdRouterReqsDurationName = "router.request.duration"
	statsdRouterReqsBytesName    = "router.requests.bytes.total"
	statsdRouterRespsBytesName   = "router.responses.bytes.total"

	statsdServiceReqsName         = "service.request.total"
	statsdServiceReqsTLSName      = "service.request.tls.total"
	statsdServiceReqsDurationName = "service.request.duration"
	statsdServiceRetriesTotalName = "service.retries.total"
	statsdServiceServerUpName     = "service.server.up"
	statsdServiceReqsBytesName    = "service.requests.bytes.total"
	statsdServiceRespsBytesName   = "service.responses.bytes.total"
)

// RegisterStatsd registers the metrics pusher if this didn't happen yet and creates a statsd Registry instance.
func RegisterStatsd(ctx context.Context, config *types.Statsd) Registry {
	// just to be sure there is a prefix defined
	if config.Prefix == "" {
		config.Prefix = defaultMetricsPrefix
	}

	statsdClient = statsd.New(config.Prefix+".", logs.NewGoKitWrapper(*log.Ctx(ctx)))

	if statsdTicker == nil {
		statsdTicker = initStatsdTicker(ctx, config)
	}

	registry := &standardRegistry{
		configReloadsCounter:           statsdClient.NewCounter(statsdConfigReloadsName, 1.0),
		lastConfigReloadSuccessGauge:   statsdClient.NewGauge(statsdLastConfigReloadSuccessName),
		tlsCertsNotAfterTimestampGauge: statsdClient.NewGauge(statsdTLSCertsNotAfterTimestampName),
		openConnectionsGauge:           statsdClient.NewGauge(statsdOpenConnectionsName),
	}

	if config.AddEntryPointsLabels {
		registry.epEnabled = config.AddEntryPointsLabels
		registry.entryPointReqsCounter = NewCounterWithNoopHeaders(statsdClient.NewCounter(statsdEntryPointReqsName, 1.0))
		registry.entryPointReqsTLSCounter = statsdClient.NewCounter(statsdEntryPointReqsTLSName, 1.0)
		registry.entryPointReqDurationHistogram, _ = NewHistogramWithScale(statsdClient.NewTiming(statsdEntryPointReqDurationName, 1.0), time.Millisecond)
		registry.entryPointReqsBytesCounter = statsdClient.NewCounter(statsdEntryPointReqsBytesName, 1.0)
		registry.entryPointRespsBytesCounter = statsdClient.NewCounter(statsdEntryPointRespsBytesName, 1.0)
	}

	if config.AddRoutersLabels {
		registry.routerEnabled = config.AddRoutersLabels
		registry.routerReqsCounter = NewCounterWithNoopHeaders(statsdClient.NewCounter(statsdRouterReqsName, 1.0))
		registry.routerReqsTLSCounter = statsdClient.NewCounter(statsdRouterReqsTLSName, 1.0)
		registry.routerReqDurationHistogram, _ = NewHistogramWithScale(statsdClient.NewTiming(statsdRouterReqsDurationName, 1.0), time.Millisecond)
		registry.routerReqsBytesCounter = statsdClient.NewCounter(statsdRouterReqsBytesName, 1.0)
		registry.routerRespsBytesCounter = statsdClient.NewCounter(statsdRouterRespsBytesName, 1.0)
	}

	if config.AddServicesLabels {
		registry.svcEnabled = config.AddServicesLabels
		registry.serviceReqsCounter = NewCounterWithNoopHeaders(statsdClient.NewCounter(statsdServiceReqsName, 1.0))
		registry.serviceReqsTLSCounter = statsdClient.NewCounter(statsdServiceReqsTLSName, 1.0)
		registry.serviceReqDurationHistogram, _ = NewHistogramWithScale(statsdClient.NewTiming(statsdServiceReqsDurationName, 1.0), time.Millisecond)
		registry.serviceRetriesCounter = statsdClient.NewCounter(statsdServiceRetriesTotalName, 1.0)
		registry.serviceServerUpGauge = statsdClient.NewGauge(statsdServiceServerUpName)
		registry.serviceReqsBytesCounter = statsdClient.NewCounter(statsdServiceReqsBytesName, 1.0)
		registry.serviceRespsBytesCounter = statsdClient.NewCounter(statsdServiceRespsBytesName, 1.0)
	}

	return registry
}

// initStatsdTicker initializes metrics pusher and creates a statsdClient if not created already.
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

// StopStatsd stops internal statsdTicker which controls the pushing of metrics to StatsD Agent and resets it to `nil`.
func StopStatsd() {
	if statsdTicker != nil {
		statsdTicker.Stop()
	}
	statsdTicker = nil
}
