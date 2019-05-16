package metrics

import (
	"context"
	"time"

	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/safe"
	"github.com/containous/traefik/pkg/types"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/dogstatsd"
)

var datadogClient = dogstatsd.New("traefik.", kitlog.LoggerFunc(func(keyvals ...interface{}) error {
	log.WithoutContext().WithField(log.MetricsProviderName, "datadog").Info(keyvals)
	return nil
}))

var datadogTicker *time.Ticker

// Metric names consistent with https://github.com/DataDog/integrations-extras/pull/64
const (
	ddMetricsBackendReqsName      = "backend.request.total"
	ddMetricsBackendLatencyName   = "backend.request.duration"
	ddRetriesTotalName            = "backend.retries.total"
	ddConfigReloadsName           = "config.reload.total"
	ddConfigReloadsFailureTagName = "failure"
	ddLastConfigReloadSuccessName = "config.reload.lastSuccessTimestamp"
	ddLastConfigReloadFailureName = "config.reload.lastFailureTimestamp"
	ddEntrypointReqsName          = "entrypoint.request.total"
	ddEntrypointReqDurationName   = "entrypoint.request.duration"
	ddEntrypointOpenConnsName     = "entrypoint.connections.open"
	ddOpenConnsName               = "backend.connections.open"
	ddServerUpName                = "backend.server.up"
)

// RegisterDatadog registers the metrics pusher if this didn't happen yet and creates a datadog Registry instance.
func RegisterDatadog(ctx context.Context, config *types.Datadog) Registry {
	if datadogTicker == nil {
		datadogTicker = initDatadogClient(ctx, config)
	}

	registry := &standardRegistry{
		enabled:                        true,
		configReloadsCounter:           datadogClient.NewCounter(ddConfigReloadsName, 1.0),
		configReloadsFailureCounter:    datadogClient.NewCounter(ddConfigReloadsName, 1.0).With(ddConfigReloadsFailureTagName, "true"),
		lastConfigReloadSuccessGauge:   datadogClient.NewGauge(ddLastConfigReloadSuccessName),
		lastConfigReloadFailureGauge:   datadogClient.NewGauge(ddLastConfigReloadFailureName),
		entrypointReqsCounter:          datadogClient.NewCounter(ddEntrypointReqsName, 1.0),
		entrypointReqDurationHistogram: datadogClient.NewHistogram(ddEntrypointReqDurationName, 1.0),
		entrypointOpenConnsGauge:       datadogClient.NewGauge(ddEntrypointOpenConnsName),
		backendReqsCounter:             datadogClient.NewCounter(ddMetricsBackendReqsName, 1.0),
		backendReqDurationHistogram:    datadogClient.NewHistogram(ddMetricsBackendLatencyName, 1.0),
		backendRetriesCounter:          datadogClient.NewCounter(ddRetriesTotalName, 1.0),
		backendOpenConnsGauge:          datadogClient.NewGauge(ddOpenConnsName),
		backendServerUpGauge:           datadogClient.NewGauge(ddServerUpName),
	}

	return registry
}

func initDatadogClient(ctx context.Context, config *types.Datadog) *time.Ticker {
	address := config.Address
	if len(address) == 0 {
		address = "localhost:8125"
	}

	report := time.NewTicker(time.Duration(config.PushInterval))

	safe.Go(func() {
		datadogClient.SendLoop(report.C, "udp", address)
	})

	return report
}

// StopDatadog stops internal datadogTicker which controls the pushing of metrics to DD Agent and resets it to `nil`.
func StopDatadog() {
	if datadogTicker != nil {
		datadogTicker.Stop()
	}
	datadogTicker = nil
}
