package middlewares

import (
	"time"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/dogstatsd"
)

var _ Metrics = (Metrics)(nil)

var datadogClient = dogstatsd.New("traefik.", kitlog.LoggerFunc(func(keyvals ...interface{}) error {
	log.Info(keyvals)
	return nil
}))

var datadogTicker *time.Ticker

// Metric names consistent with https://github.com/DataDog/integrations-extras/pull/64
const (
	ddMetricsReqsName    = "requests.total"
	ddMetricsLatencyName = "request.duration"
)

// Datadog is an Implementation for Metrics that exposes datadog metrics for the latency
// and the number of requests partitioned by status code and method.
// - number of requests partitioned by status code and method
// - request durations
// - amount of retries happened
type Datadog struct {
	reqsCounter          metrics.Counter
	reqDurationHistogram metrics.Histogram
	retryCounter         metrics.Counter
}

func (dd *Datadog) getReqsCounter() metrics.Counter {
	return dd.reqsCounter
}

func (dd *Datadog) getReqDurationHistogram() metrics.Histogram {
	return dd.reqDurationHistogram
}

func (dd *Datadog) getRetryCounter() metrics.Counter {
	return dd.retryCounter
}

// NewDataDog creates new instance of Datadog
func NewDataDog(name string) *Datadog {
	var m Datadog

	m.reqsCounter = datadogClient.NewCounter(ddMetricsReqsName, 1.0).With("service", name)
	m.reqDurationHistogram = datadogClient.NewHistogram(ddMetricsLatencyName, 1.0).With("service", name)

	return &m
}

// InitDatadogClient initializes metrics pusher and creates a datadogClient if not created already
func InitDatadogClient(config *types.Datadog) *time.Ticker {
	if datadogTicker == nil {
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

		datadogTicker = report
	}
	return datadogTicker
}

// StopDatadogClient stops internal datadogTicker which controls the pushing of metrics to DD Agent and resets it to `nil`
func StopDatadogClient() {
	if datadogTicker != nil {
		datadogTicker.Stop()
	}
	datadogTicker = nil
}
