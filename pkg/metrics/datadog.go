package metrics

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/go-kit/kit/metrics/dogstatsd"
	"github.com/go-kit/kit/util/conn"
	gokitlog "github.com/go-kit/log"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/types"
)

const (
	unixAddressPrefix         = "unix://"
	unixAddressDatagramPrefix = "unixgram://"
	unixAddressStreamPrefix   = "unixstream://"
)

var (
	datadogClient         *dogstatsd.Dogstatsd
	datadogLoopCancelFunc context.CancelFunc
)

// Metric names consistent with https://github.com/DataDog/integrations-extras/pull/64
const (
	ddConfigReloadsName           = "config.reload.total"
	ddLastConfigReloadSuccessName = "config.reload.lastSuccessTimestamp"
	ddOpenConnsName               = "open.connections"

	ddTLSCertsNotAfterTimestampName = "tls.certs.notAfterTimestamp"

	ddEntryPointReqsName        = "entrypoint.request.total"
	ddEntryPointReqsTLSName     = "entrypoint.request.tls.total"
	ddEntryPointReqDurationName = "entrypoint.request.duration"
	ddEntryPointReqsBytesName   = "entrypoint.requests.bytes.total"
	ddEntryPointRespsBytesName  = "entrypoint.responses.bytes.total"

	ddRouterReqsName         = "router.request.total"
	ddRouterReqsTLSName      = "router.request.tls.total"
	ddRouterReqsDurationName = "router.request.duration"
	ddRouterReqsBytesName    = "router.requests.bytes.total"
	ddRouterRespsBytesName   = "router.responses.bytes.total"

	ddServiceReqsName         = "service.request.total"
	ddServiceReqsTLSName      = "service.request.tls.total"
	ddServiceReqsDurationName = "service.request.duration"
	ddServiceRetriesName      = "service.retries.total"
	ddServiceServerUpName     = "service.server.up"
	ddServiceReqsBytesName    = "service.requests.bytes.total"
	ddServiceRespsBytesName   = "service.responses.bytes.total"
)

// RegisterDatadog registers the metrics pusher if this didn't happen yet and creates a datadog Registry instance.
func RegisterDatadog(ctx context.Context, config *types.Datadog) Registry {
	// Ensures there is only one DataDog client sending metrics at any given time.
	StopDatadog()

	// just to be sure there is a prefix defined
	if config.Prefix == "" {
		config.Prefix = defaultMetricsPrefix
	}

	datadogLogger := logs.NewGoKitWrapper(log.Logger.With().Str(logs.MetricsProviderName, "datadog").Logger())
	datadogClient = dogstatsd.New(config.Prefix+".", datadogLogger)

	initDatadogClient(ctx, config, datadogLogger)

	registry := &standardRegistry{
		configReloadsCounter:           datadogClient.NewCounter(ddConfigReloadsName, 1.0),
		lastConfigReloadSuccessGauge:   datadogClient.NewGauge(ddLastConfigReloadSuccessName),
		openConnectionsGauge:           datadogClient.NewGauge(ddOpenConnsName),
		tlsCertsNotAfterTimestampGauge: datadogClient.NewGauge(ddTLSCertsNotAfterTimestampName),
	}

	if config.AddEntryPointsLabels {
		registry.epEnabled = config.AddEntryPointsLabels
		registry.entryPointReqsCounter = NewCounterWithNoopHeaders(datadogClient.NewCounter(ddEntryPointReqsName, 1.0))
		registry.entryPointReqsTLSCounter = datadogClient.NewCounter(ddEntryPointReqsTLSName, 1.0)
		registry.entryPointReqDurationHistogram, _ = NewHistogramWithScale(datadogClient.NewHistogram(ddEntryPointReqDurationName, 1.0), time.Second)
		registry.entryPointReqsBytesCounter = datadogClient.NewCounter(ddEntryPointReqsBytesName, 1.0)
		registry.entryPointRespsBytesCounter = datadogClient.NewCounter(ddEntryPointRespsBytesName, 1.0)
	}

	if config.AddRoutersLabels {
		registry.routerEnabled = config.AddRoutersLabels
		registry.routerReqsCounter = NewCounterWithNoopHeaders(datadogClient.NewCounter(ddRouterReqsName, 1.0))
		registry.routerReqsTLSCounter = datadogClient.NewCounter(ddRouterReqsTLSName, 1.0)
		registry.routerReqDurationHistogram, _ = NewHistogramWithScale(datadogClient.NewHistogram(ddRouterReqsDurationName, 1.0), time.Second)
		registry.routerReqsBytesCounter = datadogClient.NewCounter(ddRouterReqsBytesName, 1.0)
		registry.routerRespsBytesCounter = datadogClient.NewCounter(ddRouterRespsBytesName, 1.0)
	}

	if config.AddServicesLabels {
		registry.svcEnabled = config.AddServicesLabels
		registry.serviceReqsCounter = NewCounterWithNoopHeaders(datadogClient.NewCounter(ddServiceReqsName, 1.0))
		registry.serviceReqsTLSCounter = datadogClient.NewCounter(ddServiceReqsTLSName, 1.0)
		registry.serviceReqDurationHistogram, _ = NewHistogramWithScale(datadogClient.NewHistogram(ddServiceReqsDurationName, 1.0), time.Second)
		registry.serviceRetriesCounter = datadogClient.NewCounter(ddServiceRetriesName, 1.0)
		registry.serviceServerUpGauge = datadogClient.NewGauge(ddServiceServerUpName)
		registry.serviceReqsBytesCounter = datadogClient.NewCounter(ddServiceReqsBytesName, 1.0)
		registry.serviceRespsBytesCounter = datadogClient.NewCounter(ddServiceRespsBytesName, 1.0)
	}

	return registry
}

func initDatadogClient(ctx context.Context, config *types.Datadog, logger gokitlog.LoggerFunc) {
	network, address := parseDatadogAddress(config.Address)

	ctx, datadogLoopCancelFunc = context.WithCancel(ctx)

	safe.Go(func() {
		ticker := time.NewTicker(time.Duration(config.PushInterval))
		defer ticker.Stop()

		dialer := func(network, address string) (net.Conn, error) {
			switch network {
			case "unix":
				// To mimic the Datadog client when the network is unix we will try to guess the UDS type.
				newConn, err := net.Dial("unixgram", address)
				if err != nil && strings.Contains(err.Error(), "protocol wrong type for socket") {
					return net.Dial("unix", address)
				}
				return newConn, err

			case "unixgram":
				return net.Dial("unixgram", address)

			case "unixstream":
				return net.Dial("unix", address)

			default:
				return net.Dial(network, address)
			}
		}
		datadogClient.WriteLoop(ctx, ticker.C, conn.NewManager(dialer, network, address, time.After, logger))
	})
}

// StopDatadog stops the Datadog metrics pusher.
func StopDatadog() {
	if datadogLoopCancelFunc != nil {
		datadogLoopCancelFunc()
		datadogLoopCancelFunc = nil
	}
}

func parseDatadogAddress(address string) (string, string) {
	network := "udp"

	var addr string
	switch {
	case strings.HasPrefix(address, unixAddressPrefix):
		network = "unix"
		addr = address[len(unixAddressPrefix):]
	case strings.HasPrefix(address, unixAddressDatagramPrefix):
		network = "unixgram"
		addr = address[len(unixAddressDatagramPrefix):]
	case strings.HasPrefix(address, unixAddressStreamPrefix):
		network = "unixstream"
		addr = address[len(unixAddressStreamPrefix):]
	case address != "":
		addr = address
	default:
		addr = "localhost:8125"
	}

	return network, addr
}
