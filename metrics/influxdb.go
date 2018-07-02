package metrics

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/influx"
	influxdb "github.com/influxdata/influxdb/client/v2"
)

var influxDBClient *influx.Influx

type influxDBWriter struct {
	buf    bytes.Buffer
	config *types.InfluxDB
}

var influxDBTicker *time.Ticker

const (
	influxDBMetricsBackendReqsName      = "traefik.backend.requests.total"
	influxDBMetricsBackendLatencyName   = "traefik.backend.request.duration"
	influxDBRetriesTotalName            = "traefik.backend.retries.total"
	influxDBConfigReloadsName           = "traefik.config.reload.total"
	influxDBConfigReloadsFailureName    = influxDBConfigReloadsName + ".failure"
	influxDBLastConfigReloadSuccessName = "traefik.config.reload.lastSuccessTimestamp"
	influxDBLastConfigReloadFailureName = "traefik.config.reload.lastFailureTimestamp"
	influxDBEntrypointReqsName          = "traefik.entrypoint.requests.total"
	influxDBEntrypointReqDurationName   = "traefik.entrypoint.request.duration"
	influxDBEntrypointOpenConnsName     = "traefik.entrypoint.connections.open"
	influxDBOpenConnsName               = "traefik.backend.connections.open"
	influxDBServerUpName                = "traefik.backend.server.up"
)

// RegisterInfluxDB registers the metrics pusher if this didn't happen yet and creates a InfluxDB Registry instance.
func RegisterInfluxDB(config *types.InfluxDB) Registry {
	if influxDBClient == nil {
		influxDBClient = initInfluxDBClient(config)
	}
	if influxDBTicker == nil {
		influxDBTicker = initInfluxDBTicker(config)
	}

	return &standardRegistry{
		enabled:                        true,
		configReloadsCounter:           influxDBClient.NewCounter(influxDBConfigReloadsName),
		configReloadsFailureCounter:    influxDBClient.NewCounter(influxDBConfigReloadsFailureName),
		lastConfigReloadSuccessGauge:   influxDBClient.NewGauge(influxDBLastConfigReloadSuccessName),
		lastConfigReloadFailureGauge:   influxDBClient.NewGauge(influxDBLastConfigReloadFailureName),
		entrypointReqsCounter:          influxDBClient.NewCounter(influxDBEntrypointReqsName),
		entrypointReqDurationHistogram: influxDBClient.NewHistogram(influxDBEntrypointReqDurationName),
		entrypointOpenConnsGauge:       influxDBClient.NewGauge(influxDBEntrypointOpenConnsName),
		backendReqsCounter:             influxDBClient.NewCounter(influxDBMetricsBackendReqsName),
		backendReqDurationHistogram:    influxDBClient.NewHistogram(influxDBMetricsBackendLatencyName),
		backendRetriesCounter:          influxDBClient.NewCounter(influxDBRetriesTotalName),
		backendOpenConnsGauge:          influxDBClient.NewGauge(influxDBOpenConnsName),
		backendServerUpGauge:           influxDBClient.NewGauge(influxDBServerUpName),
	}
}

// initInfluxDBTicker creates a influxDBClient
func initInfluxDBClient(config *types.InfluxDB) *influx.Influx {
	// TODO deprecated: move this switch into configuration.SetEffectiveConfiguration when web provider will be removed.
	switch config.Protocol {
	case "udp":
		if len(config.Database) > 0 || len(config.RetentionPolicy) > 0 {
			log.Warn("Database and RetentionPolicy are only used when protocol is http.")
			config.Database = ""
			config.RetentionPolicy = ""
		}
	case "http":
		if u, err := url.Parse(config.Address); err == nil {
			if u.Scheme != "http" && u.Scheme != "https" {
				log.Warnf("InfluxDB address %s should specify a scheme of http or https, defaulting to http.", config.Address)
				config.Address = "http://" + config.Address
			}
		} else {
			log.Errorf("Unable to parse influxdb address: %v, defaulting to udp.", err)
			config.Protocol = "udp"
			config.Database = ""
			config.RetentionPolicy = ""
		}
	default:
		log.Warnf("Unsupported protocol: %s, defaulting to udp.", config.Protocol)
		config.Protocol = "udp"
		config.Database = ""
		config.RetentionPolicy = ""
	}

	return influx.New(
		map[string]string{},
		influxdb.BatchPointsConfig{
			Database:        config.Database,
			RetentionPolicy: config.RetentionPolicy,
		},
		kitlog.LoggerFunc(func(keyvals ...interface{}) error {
			log.Info(keyvals)
			return nil
		}))
}

// initInfluxDBTicker initializes metrics pusher
func initInfluxDBTicker(config *types.InfluxDB) *time.Ticker {
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

// Write creates a http or udp client and attempts to write BatchPoints.
// If a "database not found" error is encountered, a CREATE DATABASE
// query is attempted when using protocol http.
func (w *influxDBWriter) Write(bp influxdb.BatchPoints) error {
	c, err := w.initWriteClient()
	if err != nil {
		return err
	}

	defer c.Close()

	if writeErr := c.Write(bp); writeErr != nil {
		log.Errorf("Error writing to influx: %s", writeErr.Error())
		if handleErr := w.handleWriteError(c, writeErr); handleErr != nil {
			return handleErr
		}
		// Retry write after successful handling of writeErr
		return c.Write(bp)
	}
	return nil
}

func (w *influxDBWriter) initWriteClient() (influxdb.Client, error) {
	if w.config.Protocol == "http" {
		return influxdb.NewHTTPClient(influxdb.HTTPConfig{
			Addr: w.config.Address,
		})
	}

	return influxdb.NewUDPClient(influxdb.UDPConfig{
		Addr: w.config.Address,
	})
}

func (w *influxDBWriter) handleWriteError(c influxdb.Client, writeErr error) error {
	if w.config.Protocol != "http" {
		return writeErr
	}

	match, matchErr := regexp.MatchString("database not found", writeErr.Error())

	if matchErr != nil || !match {
		return writeErr
	}

	qStr := fmt.Sprintf("CREATE DATABASE \"%s\"", w.config.Database)
	if w.config.RetentionPolicy != "" {
		qStr = fmt.Sprintf("%s WITH NAME \"%s\"", qStr, w.config.RetentionPolicy)
	}

	log.Debugf("Influx database does not exist, attempting to create with query: %s", qStr)

	q := influxdb.NewQuery(qStr, "", "")
	response, queryErr := c.Query(q)
	if queryErr == nil && response.Error() != nil {
		queryErr = response.Error()
	}
	if queryErr != nil {
		log.Errorf("Error creating InfluxDB database: %s", queryErr)
		return queryErr
	}

	log.Debugf("Successfully created influx database: %s", w.config.Database)
	return nil
}
