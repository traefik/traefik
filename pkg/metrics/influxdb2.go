package metrics

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/metrics/influx"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2api "github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	influxdb2log "github.com/influxdata/influxdb-client-go/v2/log"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/types"
)

var (
	influxDB2Ticker *time.Ticker
	influxDB2Store  *influx.Influx
	influxDB2Client influxdb2.Client
)

const (
	influxDBConfigReloadsName           = "traefik.config.reload.total"
	influxDBLastConfigReloadSuccessName = "traefik.config.reload.lastSuccessTimestamp"
	influxDBOpenConnsName               = "traefik.open.connections"

	influxDBTLSCertsNotAfterTimestampName = "traefik.tls.certs.notAfterTimestamp"

	influxDBEntryPointReqsName        = "traefik.entrypoint.requests.total"
	influxDBEntryPointReqsTLSName     = "traefik.entrypoint.requests.tls.total"
	influxDBEntryPointReqDurationName = "traefik.entrypoint.request.duration"
	influxDBEntryPointReqsBytesName   = "traefik.entrypoint.requests.bytes.total"
	influxDBEntryPointRespsBytesName  = "traefik.entrypoint.responses.bytes.total"

	influxDBRouterReqsName         = "traefik.router.requests.total"
	influxDBRouterReqsTLSName      = "traefik.router.requests.tls.total"
	influxDBRouterReqsDurationName = "traefik.router.request.duration"
	influxDBRouterReqsBytesName    = "traefik.router.requests.bytes.total"
	influxDBRouterRespsBytesName   = "traefik.router.responses.bytes.total"

	influxDBServiceReqsName         = "traefik.service.requests.total"
	influxDBServiceReqsTLSName      = "traefik.service.requests.tls.total"
	influxDBServiceReqsDurationName = "traefik.service.request.duration"
	influxDBServiceRetriesTotalName = "traefik.service.retries.total"
	influxDBServiceServerUpName     = "traefik.service.server.up"
	influxDBServiceReqsBytesName    = "traefik.service.requests.bytes.total"
	influxDBServiceRespsBytesName   = "traefik.service.responses.bytes.total"
)

// RegisterInfluxDB2 creates metrics exporter for InfluxDB2.
func RegisterInfluxDB2(ctx context.Context, config *types.InfluxDB2) Registry {
	logger := log.Ctx(ctx)

	if influxDB2Client == nil {
		var err error
		if influxDB2Client, err = newInfluxDB2Client(config); err != nil {
			logger.Error().Err(err).Send()
			return nil
		}
	}

	if influxDB2Store == nil {
		influxDB2Store = influx.New(
			config.AdditionalLabels,
			influxdb.BatchPointsConfig{},
			logs.NewGoKitWrapper(*logger),
		)

		influxDB2Ticker = time.NewTicker(time.Duration(config.PushInterval))

		safe.Go(func() {
			wc := influxDB2Client.WriteAPIBlocking(config.Org, config.Bucket)
			influxDB2Store.WriteLoop(ctx, influxDB2Ticker.C, influxDB2Writer{wc: wc})
		})
	}

	registry := &standardRegistry{
		configReloadsCounter:           influxDB2Store.NewCounter(influxDBConfigReloadsName),
		lastConfigReloadSuccessGauge:   influxDB2Store.NewGauge(influxDBLastConfigReloadSuccessName),
		openConnectionsGauge:           influxDB2Store.NewGauge(influxDBOpenConnsName),
		tlsCertsNotAfterTimestampGauge: influxDB2Store.NewGauge(influxDBTLSCertsNotAfterTimestampName),
	}

	if config.AddEntryPointsLabels {
		registry.epEnabled = config.AddEntryPointsLabels
		registry.entryPointReqsCounter = NewCounterWithNoopHeaders(influxDB2Store.NewCounter(influxDBEntryPointReqsName))
		registry.entryPointReqsTLSCounter = influxDB2Store.NewCounter(influxDBEntryPointReqsTLSName)
		registry.entryPointReqDurationHistogram, _ = NewHistogramWithScale(influxDB2Store.NewHistogram(influxDBEntryPointReqDurationName), time.Second)
		registry.entryPointReqsBytesCounter = influxDB2Store.NewCounter(influxDBEntryPointReqsBytesName)
		registry.entryPointRespsBytesCounter = influxDB2Store.NewCounter(influxDBEntryPointRespsBytesName)
	}

	if config.AddRoutersLabels {
		registry.routerEnabled = config.AddRoutersLabels
		registry.routerReqsCounter = NewCounterWithNoopHeaders(influxDB2Store.NewCounter(influxDBRouterReqsName))
		registry.routerReqsTLSCounter = influxDB2Store.NewCounter(influxDBRouterReqsTLSName)
		registry.routerReqDurationHistogram, _ = NewHistogramWithScale(influxDB2Store.NewHistogram(influxDBRouterReqsDurationName), time.Second)
		registry.routerReqsBytesCounter = influxDB2Store.NewCounter(influxDBRouterReqsBytesName)
		registry.routerRespsBytesCounter = influxDB2Store.NewCounter(influxDBRouterRespsBytesName)
	}

	if config.AddServicesLabels {
		registry.svcEnabled = config.AddServicesLabels
		registry.serviceReqsCounter = NewCounterWithNoopHeaders(influxDB2Store.NewCounter(influxDBServiceReqsName))
		registry.serviceReqsTLSCounter = influxDB2Store.NewCounter(influxDBServiceReqsTLSName)
		registry.serviceReqDurationHistogram, _ = NewHistogramWithScale(influxDB2Store.NewHistogram(influxDBServiceReqsDurationName), time.Second)
		registry.serviceRetriesCounter = influxDB2Store.NewCounter(influxDBServiceRetriesTotalName)
		registry.serviceServerUpGauge = influxDB2Store.NewGauge(influxDBServiceServerUpName)
		registry.serviceReqsBytesCounter = influxDB2Store.NewCounter(influxDBServiceReqsBytesName)
		registry.serviceRespsBytesCounter = influxDB2Store.NewCounter(influxDBServiceRespsBytesName)
	}

	return registry
}

// StopInfluxDB2 stops and resets InfluxDB2 client, ticker and store.
func StopInfluxDB2() {
	if influxDB2Client != nil {
		influxDB2Client.Close()
	}
	influxDB2Client = nil

	if influxDB2Ticker != nil {
		influxDB2Ticker.Stop()
	}
	influxDB2Ticker = nil

	influxDB2Store = nil
}

// newInfluxDB2Client creates an influxdb2.Client.
func newInfluxDB2Client(config *types.InfluxDB2) (influxdb2.Client, error) {
	if config.Token == "" || config.Org == "" || config.Bucket == "" {
		return nil, errors.New("token, org or bucket property is missing")
	}

	// Disable InfluxDB2 logs.
	// See https://github.com/influxdata/influxdb-client-go/blob/v2.7.0/options.go#L128
	influxdb2log.Log = nil

	return influxdb2.NewClient(config.Address, config.Token), nil
}

type influxDB2Writer struct {
	wc influxdb2api.WriteAPIBlocking
}

func (w influxDB2Writer) Write(bp influxdb.BatchPoints) error {
	logger := log.With().Str(logs.MetricsProviderName, "influxdb2").Logger()

	wps := make([]*write.Point, 0, len(bp.Points()))
	for _, p := range bp.Points() {
		fields, err := p.Fields()
		if err != nil {
			logger.Error().Err(err).Msgf("Error while getting %s point fields", p.Name())
			continue
		}

		wps = append(wps, influxdb2.NewPoint(
			p.Name(),
			p.Tags(),
			fields,
			p.Time(),
		))
	}

	ctx := logger.WithContext(context.Background())

	return w.wc.WritePoint(ctx, wps...)
}
