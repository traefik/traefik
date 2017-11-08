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

var influxDBClient = influx.New(map[string]string{}, influxdb.BatchPointsConfig{}, kitlog.LoggerFunc(func(keyvals ...interface{}) error {
	log.Info(keyvals)
	return nil
}))

type influxDBWriter struct {
	buf    bytes.Buffer
	config *types.InfluxDB
}

var influxDBTicker *time.Ticker

const (
	influxDBMetricsReqsName    = "traefik.requests.total"
	influxDBMetricsLatencyName = "traefik.request.duration"
	influxDBRetriesTotalName   = "traefik.backend.retries.total"
)

// RegisterInfluxDB registers the metrics pusher if this didn't happen yet and creates a InfluxDB Registry instance.
func RegisterInfluxDB(config *types.InfluxDB) Registry {
	if influxDBTicker == nil {
		influxDBTicker = initInfluxDBTicker(config)
	}

	return &standardRegistry{
		enabled:              true,
		reqsCounter:          influxDBClient.NewCounter(influxDBMetricsReqsName),
		reqDurationHistogram: influxDBClient.NewHistogram(influxDBMetricsLatencyName),
		retriesCounter:       influxDBClient.NewCounter(influxDBRetriesTotalName),
	}
}

// initInfluxDBTicker initializes metrics pusher and creates a influxDBClient if not created already
func initInfluxDBTicker(config *types.InfluxDB) *time.Ticker {
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
		influxDBClient.WriteLoop(report.C, &influxDBWriter{buf: buf, config: config})
	})

	return report
}

// StopInfluxDB stops internal influxDBTicker which controls the pushing of metrics to InfluxDB Agent and resets it to `nil`
func StopInfluxDB() {
	if influxDBTicker != nil {
		influxDBTicker.Stop()
	}
	influxDBTicker = nil
}

func (w *influxDBWriter) Write(bp influxdb.BatchPoints) error {
	c, err := influxdb.NewUDPClient(influxdb.UDPConfig{
		Addr: w.config.Address,
	})

	if err != nil {
		return err
	}

	defer c.Close()

	return c.Write(bp)
}
