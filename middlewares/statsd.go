package middlewares

import (
	"time"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/statsd"
)

var _ Metrics = (*Statsd)(nil)
var _ RetryMetrics = (*Statsd)(nil)

var statsdClient = statsd.New("traefik.", kitlog.LoggerFunc(func(keyvals ...interface{}) error {
	log.Info(keyvals)
	return nil
}))
var statsdTicker *time.Ticker

// Statsd is an Implementation for Metrics that exposes statsd metrics for the latency
// and the number of requests partitioned by status code and method.
// - number of requests partitioned by status code and method
// - request durations
// - amount of retries happened
type Statsd struct {
	reqsCounter          metrics.Counter
	reqDurationHistogram metrics.Histogram
	retryCounter         metrics.Counter
}

func (s *Statsd) getReqsCounter() metrics.Counter {
	return s.reqsCounter
}

func (s *Statsd) getReqDurationHistogram() metrics.Histogram {
	return s.reqDurationHistogram
}

func (s *Statsd) getRetryCounter() metrics.Counter {
	return s.retryCounter
}

// NewStatsD creates new instance of StatsD
func NewStatsD(name string) *Statsd {
	var m Statsd

	m.reqsCounter = statsdClient.NewCounter(ddMetricsReqsName, 1.0).With("service", name)
	m.reqDurationHistogram = statsdClient.NewTiming(ddMetricsLatencyName, 1.0).With("service", name)
	m.retryCounter = statsdClient.NewCounter(ddRetriesTotalName, 1.0).With("service", name)

	return &m
}

// InitStatsdClient initializes metrics pusher and creates a statsdClient if not created already
func InitStatsdClient(config *types.Statsd) *time.Ticker {
	if statsdTicker == nil {
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
			statsdClient.SendLoop(report.C, "udp", address)
		})

		statsdTicker = report
	}
	return statsdTicker
}

// StopStatsdClient stops internal statsdTicker which controls the pushing of metrics to StatsD Agent and resets it to `nil`
func StopStatsdClient() {
	if statsdTicker != nil {
		statsdTicker.Stop()
	}
	statsdTicker = nil
}
