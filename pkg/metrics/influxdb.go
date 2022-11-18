package metrics

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/go-kit/kit/metrics/influx"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v2/pkg/logs"
	"github.com/traefik/traefik/v2/pkg/safe"
	"github.com/traefik/traefik/v2/pkg/types"
)

var (
	influxDBClient *influx.Influx
	influxDBTicker *time.Ticker
)

const (
	influxDBConfigReloadsName           = "traefik.config.reload.total"
	influxDBConfigReloadsFailureName    = influxDBConfigReloadsName + ".failure"
	influxDBLastConfigReloadSuccessName = "traefik.config.reload.lastSuccessTimestamp"
	influxDBLastConfigReloadFailureName = "traefik.config.reload.lastFailureTimestamp"

	influxDBTLSCertsNotAfterTimestampName = "traefik.tls.certs.notAfterTimestamp"

	influxDBEntryPointReqsName        = "traefik.entrypoint.requests.total"
	influxDBEntryPointReqsTLSName     = "traefik.entrypoint.requests.tls.total"
	influxDBEntryPointReqDurationName = "traefik.entrypoint.request.duration"
	influxDBEntryPointOpenConnsName   = "traefik.entrypoint.connections.open"
	influxDBEntryPointReqsBytesName   = "traefik.entrypoint.requests.bytes.total"
	influxDBEntryPointRespsBytesName  = "traefik.entrypoint.responses.bytes.total"

	influxDBRouterReqsName         = "traefik.router.requests.total"
	influxDBRouterReqsTLSName      = "traefik.router.requests.tls.total"
	influxDBRouterReqsDurationName = "traefik.router.request.duration"
	influxDBORouterOpenConnsName   = "traefik.router.connections.open"
	influxDBRouterReqsBytesName    = "traefik.router.requests.bytes.total"
	influxDBRouterRespsBytesName   = "traefik.router.responses.bytes.total"

	influxDBServiceReqsName         = "traefik.service.requests.total"
	influxDBServiceReqsTLSName      = "traefik.service.requests.tls.total"
	influxDBServiceReqsDurationName = "traefik.service.request.duration"
	influxDBServiceRetriesTotalName = "traefik.service.retries.total"
	influxDBServiceOpenConnsName    = "traefik.service.connections.open"
	influxDBServiceServerUpName     = "traefik.service.server.up"
	influxDBServiceReqsBytesName    = "traefik.service.requests.bytes.total"
	influxDBServiceRespsBytesName   = "traefik.service.responses.bytes.total"
)

const (
	protocolHTTP = "http"
	protocolUDP  = "udp"
)

// RegisterInfluxDB registers the metrics pusher if this didn't happen yet and creates a InfluxDB Registry instance.
func RegisterInfluxDB(ctx context.Context, config *types.InfluxDB) Registry {
	if influxDBClient == nil {
		influxDBClient = initInfluxDBClient(ctx, config)
	}
	if influxDBTicker == nil {
		influxDBTicker = initInfluxDBTicker(ctx, config)
	}

	registry := &standardRegistry{
		configReloadsCounter:           influxDBClient.NewCounter(influxDBConfigReloadsName),
		configReloadsFailureCounter:    influxDBClient.NewCounter(influxDBConfigReloadsFailureName),
		lastConfigReloadSuccessGauge:   influxDBClient.NewGauge(influxDBLastConfigReloadSuccessName),
		lastConfigReloadFailureGauge:   influxDBClient.NewGauge(influxDBLastConfigReloadFailureName),
		tlsCertsNotAfterTimestampGauge: influxDBClient.NewGauge(influxDBTLSCertsNotAfterTimestampName),
	}

	if config.AddEntryPointsLabels {
		registry.epEnabled = config.AddEntryPointsLabels
		registry.entryPointReqsCounter = influxDBClient.NewCounter(influxDBEntryPointReqsName)
		registry.entryPointReqsTLSCounter = influxDBClient.NewCounter(influxDBEntryPointReqsTLSName)
		registry.entryPointReqDurationHistogram, _ = NewHistogramWithScale(influxDBClient.NewHistogram(influxDBEntryPointReqDurationName), time.Second)
		registry.entryPointOpenConnsGauge = influxDBClient.NewGauge(influxDBEntryPointOpenConnsName)
		registry.entryPointReqsBytesCounter = influxDBClient.NewCounter(influxDBEntryPointReqsBytesName)
		registry.entryPointRespsBytesCounter = influxDBClient.NewCounter(influxDBEntryPointRespsBytesName)
	}

	if config.AddRoutersLabels {
		registry.routerEnabled = config.AddRoutersLabels
		registry.routerReqsCounter = influxDBClient.NewCounter(influxDBRouterReqsName)
		registry.routerReqsTLSCounter = influxDBClient.NewCounter(influxDBRouterReqsTLSName)
		registry.routerReqDurationHistogram, _ = NewHistogramWithScale(influxDBClient.NewHistogram(influxDBRouterReqsDurationName), time.Second)
		registry.routerOpenConnsGauge = influxDBClient.NewGauge(influxDBORouterOpenConnsName)
		registry.routerReqsBytesCounter = influxDBClient.NewCounter(influxDBRouterReqsBytesName)
		registry.routerRespsBytesCounter = influxDBClient.NewCounter(influxDBRouterRespsBytesName)
	}

	if config.AddServicesLabels {
		registry.svcEnabled = config.AddServicesLabels
		registry.serviceReqsCounter = influxDBClient.NewCounter(influxDBServiceReqsName)
		registry.serviceReqsTLSCounter = influxDBClient.NewCounter(influxDBServiceReqsTLSName)
		registry.serviceReqDurationHistogram, _ = NewHistogramWithScale(influxDBClient.NewHistogram(influxDBServiceReqsDurationName), time.Second)
		registry.serviceRetriesCounter = influxDBClient.NewCounter(influxDBServiceRetriesTotalName)
		registry.serviceOpenConnsGauge = influxDBClient.NewGauge(influxDBServiceOpenConnsName)
		registry.serviceServerUpGauge = influxDBClient.NewGauge(influxDBServiceServerUpName)
		registry.serviceReqsBytesCounter = influxDBClient.NewCounter(influxDBServiceReqsBytesName)
		registry.serviceRespsBytesCounter = influxDBClient.NewCounter(influxDBServiceRespsBytesName)
	}

	return registry
}

// initInfluxDBClient creates a influxDBClient.
func initInfluxDBClient(ctx context.Context, config *types.InfluxDB) *influx.Influx {
	logger := log.Ctx(ctx)

	// TODO deprecated: move this switch into configuration.SetEffectiveConfiguration when web provider will be removed.
	switch config.Protocol {
	case protocolUDP:
		if len(config.Database) > 0 || len(config.RetentionPolicy) > 0 {
			logger.Warn().Msg("Database and RetentionPolicy options have no effect with UDP.")
			config.Database = ""
			config.RetentionPolicy = ""
		}
	case protocolHTTP:
		if u, err := url.Parse(config.Address); err == nil {
			if u.Scheme != "http" && u.Scheme != "https" {
				logger.Warn().Msgf("InfluxDB address %s should specify a scheme (http or https): falling back on HTTP.", config.Address)
				config.Address = "http://" + config.Address
			}
		} else {
			logger.Error().Err(err).Msg("Unable to parse the InfluxDB address: falling back on UDP.")
			config.Protocol = protocolUDP
			config.Database = ""
			config.RetentionPolicy = ""
		}
	default:
		logger.Warn().Msgf("Unsupported protocol %s: falling back on UDP.", config.Protocol)
		config.Protocol = protocolUDP
		config.Database = ""
		config.RetentionPolicy = ""
	}

	return influx.New(
		config.AdditionalLabels,
		influxdb.BatchPointsConfig{
			Database:        config.Database,
			RetentionPolicy: config.RetentionPolicy,
		},
		logs.NewGoKitWrapper(*logger),
	)
}

// initInfluxDBTicker initializes metrics pusher.
func initInfluxDBTicker(ctx context.Context, config *types.InfluxDB) *time.Ticker {
	report := time.NewTicker(time.Duration(config.PushInterval))

	safe.Go(func() {
		var buf bytes.Buffer
		influxDBClient.WriteLoop(ctx, report.C, &influxDBWriter{buf: buf, config: config})
	})

	return report
}

// StopInfluxDB stops internal influxDBTicker which controls the pushing of metrics to InfluxDB Agent and resets it to `nil`.
func StopInfluxDB() {
	if influxDBTicker != nil {
		influxDBTicker.Stop()
	}
	influxDBTicker = nil
}

type influxDBWriter struct {
	buf    bytes.Buffer
	config *types.InfluxDB
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
		logger := log.With().Str(logs.MetricsProviderName, "influxdb").Logger()
		logger.Error().Err(writeErr).Msg("Error while writing to InfluxDB")

		if handleErr := w.handleWriteError(logger.WithContext(context.Background()), c, writeErr); handleErr != nil {
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
			Addr:     w.config.Address,
			Username: w.config.Username,
			Password: w.config.Password,
		})
	}

	return influxdb.NewUDPClient(influxdb.UDPConfig{
		Addr: w.config.Address,
	})
}

func (w *influxDBWriter) handleWriteError(ctx context.Context, c influxdb.Client, writeErr error) error {
	if w.config.Protocol != protocolHTTP {
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

	logger := log.Ctx(ctx)

	logger.Debug().Msgf("InfluxDB database not found: attempting to create one with %s", qStr)

	q := influxdb.NewQuery(qStr, "", "")
	response, queryErr := c.Query(q)
	if queryErr == nil && response.Error() != nil {
		queryErr = response.Error()
	}
	if queryErr != nil {
		logger.Error().Err(queryErr).Msg("Error while creating the InfluxDB database")
		return queryErr
	}

	logger.Debug().Msgf("Successfully created the InfluxDB database %s", w.config.Database)
	return nil
}
