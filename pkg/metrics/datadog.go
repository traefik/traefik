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

var datadogClient = dogstatsd.New("traefik.", kitlog.LoggerFunc(func(keyvals ...interface{}) error {
	log.WithoutContext().WithField(log.MetricsProviderName, "datadog").Info(keyvals)
	return nil
}))

var datadogTicker *time.Ticker

// Metric names consistent with https://github.com/DataDog/integrations-extras/pull/64
const (
	ddMetricsServiceReqsName      = "service.request.total"
	ddMetricsServiceLatencyName   = "service.request.duration"
	ddRetriesTotalName            = "service.retries.total"
	ddConfigReloadsName           = "config.reload.total"
	ddConfigReloadsFailureTagName = "failure"
	ddLastConfigReloadSuccessName = "config.reload.lastSuccessTimestamp"
	ddLastConfigReloadFailureName = "config.reload.lastFailureTimestamp"
	ddEntryPointReqsName          = "entrypoint.request.total"
	ddEntryPointReqDurationName   = "entrypoint.request.duration"
	ddEntryPointOpenConnsName     = "entrypoint.connections.open"
	ddOpenConnsName               = "service.connections.open"
	ddServerUpName                = "service.server.up"
)

// RegisterDatadog registers the metrics pusher if this didn't happen yet and creates a datadog Registry instance.
func RegisterDatadog(ctx context.Context, config *types.Datadog) Registry {
	if datadogTicker == nil {
		datadogTicker = initDatadogClient(ctx, config)
	}

	registry := &standardRegistry{
		configReloadsCounter:         datadogClient.NewCounter(ddConfigReloadsName, 1.0),
		configReloadsFailureCounter:  datadogClient.NewCounter(ddConfigReloadsName, 1.0).With(ddConfigReloadsFailureTagName, "true"),
		lastConfigReloadSuccessGauge: datadogClient.NewGauge(ddLastConfigReloadSuccessName),
		lastConfigReloadFailureGauge: datadogClient.NewGauge(ddLastConfigReloadFailureName),
	}

	if config.AddEntryPointsLabels {
		registry.epEnabled = config.AddEntryPointsLabels
		registry.entryPointReqsCounter = datadogClient.NewCounter(ddEntryPointReqsName, 1.0)
		registry.entryPointReqDurationHistogram, _ = NewHistogramWithScale(datadogClient.NewHistogram(ddEntryPointReqDurationName, 1.0), time.Second)
		registry.entryPointOpenConnsGauge = datadogClient.NewGauge(ddEntryPointOpenConnsName)
	}

	if config.AddServicesLabels {
		registry.svcEnabled = config.AddServicesLabels
		registry.serviceReqsCounter = datadogClient.NewCounter(ddMetricsServiceReqsName, 1.0)
		registry.serviceReqDurationHistogram, _ = NewHistogramWithScale(datadogClient.NewHistogram(ddMetricsServiceLatencyName, 1.0), time.Second)
		registry.serviceRetriesCounter = datadogClient.NewCounter(ddRetriesTotalName, 1.0)
		registry.serviceOpenConnsGauge = datadogClient.NewGauge(ddOpenConnsName)
		registry.serviceServerUpGauge = datadogClient.NewGauge(ddServerUpName)
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
		datadogClient.SendLoop(ctx, report.C, "udp", address)
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
