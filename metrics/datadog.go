package metrics

import (
	"time"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/dogstatsd"
)

var datadogClient = dogstatsd.New("traefik.", kitlog.LoggerFunc(func(keyvals ...interface{}) error {
	log.Info(keyvals)
	return nil
}))

var datadogTicker *time.Ticker

// Metric names consistent with https://github.com/DataDog/integrations-extras/pull/64
const (
	ddMetricsReqsName    = "requests.total"
	ddMetricsLatencyName = "request.duration"
	ddRetriesTotalName   = "backend.retries.total"
)

// RegisterDatadog registers the metrics pusher if this didn't happen yet and creates a datadog Registry instance.
func RegisterDatadog(config *types.Datadog) Registry {
	if datadogTicker == nil {
		datadogTicker = initDatadogClient(config)
	}

	registry := &datadogRegistry{
		reqsCounter:          datadogClient.NewCounter(ddMetricsReqsName, 1.0),
		reqDurationHistogram: datadogClient.NewHistogram(ddMetricsLatencyName, 1.0),
		retriesCounter:       datadogClient.NewCounter(ddRetriesTotalName, 1.0),
	}

	return registry
}

type datadogRegistry struct {
	reqsCounter          metrics.Counter
	reqDurationHistogram metrics.Histogram
	retriesCounter       metrics.Counter
}

func (dd *datadogRegistry) ReqsCounter() metrics.Counter {
	return dd.reqsCounter
}

func (dd *datadogRegistry) ReqDurationHistogram() metrics.Histogram {
	return dd.reqDurationHistogram
}

func (dd *datadogRegistry) RetriesCounter() metrics.Counter {
	return dd.retriesCounter
}

func initDatadogClient(config *types.Datadog) *time.Ticker {
	address := config.Address
	if len(address) == 0 {
		address = "localhost:8125"
	}
	pushInterval, err := time.ParseDuration(config.PushInterval)
	if err != nil {
		log.Warnf("Unable to parse %s into pushInterval, using 10s as default value", config.PushInterval)
		pushInterval = 10 * time.Second
	}

	report := time.NewTicker(pushInterval)

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
