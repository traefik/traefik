package metrics

import (
	"context"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/dogstatsd"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/safe"
	"github.com/traefik/traefik/v2/pkg/types"
)

var (
	datadogClient         *dogstatsd.Dogstatsd
	datadogLoopCancelFunc context.CancelFunc
)

// Metric names consistent with https://github.com/DataDog/integrations-extras/pull/64
const (
	ddConfigReloadsName             = "config.reload.total"
	ddConfigReloadsFailureTagName   = "failure"
	ddLastConfigReloadSuccessName   = "config.reload.lastSuccessTimestamp"
	ddLastConfigReloadFailureName   = "config.reload.lastFailureTimestamp"
	ddTLSCertsNotAfterTimestampName = "tls.certs.notAfterTimestamp"

	ddEntryPointReqsName        = "entrypoint.request.total"
	ddEntryPointReqsTLSName     = "entrypoint.request.tls.total"
	ddEntryPointReqDurationName = "entrypoint.request.duration"
	ddEntryPointOpenConnsName   = "entrypoint.connections.open"
	ddEntryPointReqsBytesName   = "entrypoint.requests.bytes.total"
	ddEntryPointRespsBytesName  = "entrypoint.responses.bytes.total"

	ddRouterReqsName         = "router.request.total"
	ddRouterReqsTLSName      = "router.request.tls.total"
	ddRouterReqsDurationName = "router.request.duration"
	ddRouterOpenConnsName    = "router.connections.open"
	ddRouterReqsBytesName    = "router.requests.bytes.total"
	ddRouterRespsBytesName   = "router.responses.bytes.total"

	ddServiceReqsName         = "service.request.total"
	ddServiceReqsTLSName      = "service.request.tls.total"
	ddServiceReqsDurationName = "service.request.duration"
	ddServiceRetriesName      = "service.retries.total"
	ddServiceOpenConnsName    = "service.connections.open"
	ddServiceServerUpName     = "service.server.up"
	ddServiceReqsBytesName    = "service.requests.bytes.total"
	ddServiceRespsBytesName   = "service.responses.bytes.total"
)

// RegisterDatadog registers the metrics pusher if this didn't happen yet and creates a datadog Registry instance.
func RegisterDatadog(ctx context.Context, config *types.Datadog) Registry {
	// Ensures there is only one DataDog client sending metrics at any given time.
	StopDatadog()

	// just to be sure there is a prefix defined
	if config.Prefix == "" {
		config.Prefix = defaultMetricsPrefix
	}

	datadogClient = dogstatsd.New(config.Prefix+".", kitlog.LoggerFunc(func(keyvals ...interface{}) error {
		log.WithoutContext().WithField(log.MetricsProviderName, "datadog").Info(keyvals...)
		return nil
	}))

	initDatadogClient(ctx, config)

	registry := &standardRegistry{
		configReloadsCounter:           datadogClient.NewCounter(ddConfigReloadsName, 1.0),
		configReloadsFailureCounter:    datadogClient.NewCounter(ddConfigReloadsName, 1.0).With(ddConfigReloadsFailureTagName, "true"),
		lastConfigReloadSuccessGauge:   datadogClient.NewGauge(ddLastConfigReloadSuccessName),
		lastConfigReloadFailureGauge:   datadogClient.NewGauge(ddLastConfigReloadFailureName),
		tlsCertsNotAfterTimestampGauge: datadogClient.NewGauge(ddTLSCertsNotAfterTimestampName),
	}

	if config.AddEntryPointsLabels {
		registry.epEnabled = config.AddEntryPointsLabels
		registry.entryPointReqsCounter = NewCounterWithNoopHeaders(datadogClient.NewCounter(ddEntryPointReqsName, 1.0))
		registry.entryPointReqsTLSCounter = datadogClient.NewCounter(ddEntryPointReqsTLSName, 1.0)
		registry.entryPointReqDurationHistogram, _ = NewHistogramWithScale(datadogClient.NewHistogram(ddEntryPointReqDurationName, 1.0), time.Second)
		registry.entryPointOpenConnsGauge = datadogClient.NewGauge(ddEntryPointOpenConnsName)
		registry.entryPointReqsBytesCounter = datadogClient.NewCounter(ddEntryPointReqsBytesName, 1.0)
		registry.entryPointRespsBytesCounter = datadogClient.NewCounter(ddEntryPointRespsBytesName, 1.0)
	}

	if config.AddRoutersLabels {
		registry.routerEnabled = config.AddRoutersLabels
		registry.routerReqsCounter = NewCounterWithNoopHeaders(datadogClient.NewCounter(ddRouterReqsName, 1.0))
		registry.routerReqsTLSCounter = datadogClient.NewCounter(ddRouterReqsTLSName, 1.0)
		registry.routerReqDurationHistogram, _ = NewHistogramWithScale(datadogClient.NewHistogram(ddRouterReqsDurationName, 1.0), time.Second)
		registry.routerOpenConnsGauge = datadogClient.NewGauge(ddRouterOpenConnsName)
		registry.routerReqsBytesCounter = datadogClient.NewCounter(ddRouterReqsBytesName, 1.0)
		registry.routerRespsBytesCounter = datadogClient.NewCounter(ddRouterRespsBytesName, 1.0)
	}

	if config.AddServicesLabels {
		registry.svcEnabled = config.AddServicesLabels
		registry.serviceReqsCounter = NewCounterWithNoopHeaders(datadogClient.NewCounter(ddServiceReqsName, 1.0))
		registry.serviceReqsTLSCounter = datadogClient.NewCounter(ddServiceReqsTLSName, 1.0)
		registry.serviceReqDurationHistogram, _ = NewHistogramWithScale(datadogClient.NewHistogram(ddServiceReqsDurationName, 1.0), time.Second)
		registry.serviceRetriesCounter = datadogClient.NewCounter(ddServiceRetriesName, 1.0)
		registry.serviceOpenConnsGauge = datadogClient.NewGauge(ddServiceOpenConnsName)
		registry.serviceServerUpGauge = datadogClient.NewGauge(ddServiceServerUpName)
		registry.serviceReqsBytesCounter = datadogClient.NewCounter(ddServiceReqsBytesName, 1.0)
		registry.serviceRespsBytesCounter = datadogClient.NewCounter(ddServiceRespsBytesName, 1.0)
	}

	return registry
}

func initDatadogClient(ctx context.Context, config *types.Datadog) {
	address := config.Address
	if len(address) == 0 {
		address = "localhost:8125"
	}

	ctx, datadogLoopCancelFunc = context.WithCancel(ctx)

	safe.Go(func() {
		ticker := time.NewTicker(time.Duration(config.PushInterval))
		defer ticker.Stop()

		datadogClient.SendLoop(ctx, ticker.C, "udp", address)
	})
}

// StopDatadog stops the Datadog metrics pusher.
func StopDatadog() {
	if datadogLoopCancelFunc != nil {
		datadogLoopCancelFunc()
		datadogLoopCancelFunc = nil
	}
}
