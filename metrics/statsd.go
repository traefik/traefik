package metrics

import (
	"time"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/statsd"
)

var statsdClient = statsd.New("traefik.", kitlog.LoggerFunc(func(keyvals ...interface{}) error {
	log.Info(keyvals)
	return nil
}))

var statsdTicker *time.Ticker

// RegisterStatsd registers the metrics pusher if this didn't happen yet and creates a statsd Registry instance.
func RegisterStatsd(config *types.Statsd) Registry {
	if statsdTicker == nil {
		statsdTicker = initStatsdTicker(config)
	}

	return &standardRegistry{
		enabled:              true,
		reqsCounter:          statsdClient.NewCounter(ddMetricsReqsName, 1.0),
		reqDurationHistogram: statsdClient.NewTiming(ddMetricsLatencyName, 1.0),
		retriesCounter:       statsdClient.NewCounter(ddRetriesTotalName, 1.0),
	}
}

// initStatsdTicker initializes metrics pusher and creates a statsdClient if not created already
func initStatsdTicker(config *types.Statsd) *time.Ticker {
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

	return report
}

// StopStatsd stops internal statsdTicker which controls the pushing of metrics to StatsD Agent and resets it to `nil`
func StopStatsd() {
	if statsdTicker != nil {
		statsdTicker.Stop()
	}
	statsdTicker = nil
}
