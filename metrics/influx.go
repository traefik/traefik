package metrics

import (
	"bytes"
	"time"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/influx"
	influxdb "github.com/influxdata/influxdb/client/v2"
)

var influxClient = influx.New(map[string]string{}, influxdb.BatchPointsConfig{}, kitlog.LoggerFunc(func(keyvals ...interface{}) error {
	log.Info(keyvals)
	return nil
}))

type influxWriter struct {
	buf    bytes.Buffer
	config *types.Influx
}

var influxTicker *time.Ticker

const (
	influxMetricsReqsName    = "traefik_requests_total"
	influxMetricsLatencyName = "traefik_request_duration"
	influxRetriesTotalName   = "traefik_backend_retries_total"
)

// RegisterInflux registers the metrics pusher if this didn't happen yet and creates a Influx Registry instance.
func RegisterInflux(config *types.Influx) Registry {
	if influxTicker == nil {
		influxTicker = initInfluxTicker(config)
	}

	return &standardRegistry{
		enabled:              true,
		reqsCounter:          influxClient.NewCounter(influxMetricsReqsName),
		reqDurationHistogram: influxClient.NewHistogram(influxMetricsLatencyName),
		retriesCounter:       influxClient.NewCounter(influxRetriesTotalName),
	}
}

// initInfluxTicker initializes metrics pusher and creates a influxClient if not created already
func initInfluxTicker(config *types.Influx) *time.Ticker {
	address := config.Address
	if len(address) == 0 {
		address = "localhost:8089"
	}

	pushInterval, err := time.ParseDuration(config.PushInterval)
	if err != nil {
		log.Warnf("Unable to parse %s into pushInterval, using 10s as default value", config.PushInterval)
		pushInterval = 10 * time.Second
	}

	report := time.NewTicker(pushInterval)

	safe.Go(func() {
		var buf bytes.Buffer
		influxClient.WriteLoop(report.C, &influxWriter{buf: buf, config: config})
	})

	return report
}

// StopInflux stops internal influxTicker which controls the pushing of metrics to Influx Agent and resets it to `nil`
func StopInflux() {
	if influxTicker != nil {
		influxTicker.Stop()
	}
	influxTicker = nil
}

func (w *influxWriter) Write(bp influxdb.BatchPoints) error {
	c, err := influxdb.NewUDPClient(influxdb.UDPConfig{
		Addr: w.config.Address,
	})

	if err != nil {
		return err
	}

	defer c.Close()

	return c.Write(bp)
}
