package main

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/go-acme/lego/v4/challenge"
	gokitmetrics "github.com/go-kit/kit/metrics"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/traefik/paerser/cli"
	"github.com/traefik/traefik/v3/cmd"
	"github.com/traefik/traefik/v3/cmd/healthcheck"
	cmdVersion "github.com/traefik/traefik/v3/cmd/version"
	tcli "github.com/traefik/traefik/v3/pkg/cli"
	"github.com/traefik/traefik/v3/pkg/collector"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/metrics"
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslog"
	"github.com/traefik/traefik/v3/pkg/provider/acme"
	"github.com/traefik/traefik/v3/pkg/provider/aggregator"
	"github.com/traefik/traefik/v3/pkg/provider/hub"
	"github.com/traefik/traefik/v3/pkg/provider/tailscale"
	"github.com/traefik/traefik/v3/pkg/provider/traefik"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/server"
	"github.com/traefik/traefik/v3/pkg/server/middleware"
	"github.com/traefik/traefik/v3/pkg/server/service"
	"github.com/traefik/traefik/v3/pkg/tcp"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"github.com/traefik/traefik/v3/pkg/tracing/jaeger"
	"github.com/traefik/traefik/v3/pkg/types"
	"github.com/traefik/traefik/v3/pkg/version"
)

func main() {
	// traefik config inits
	tConfig := cmd.NewTraefikConfiguration()

	loaders := []cli.ResourceLoader{&tcli.FileLoader{}, &tcli.FlagLoader{}, &tcli.EnvLoader{}}

	cmdTraefik := &cli.Command{
		Name: "traefik",
		Description: `Traefik is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease.
Complete documentation is available at https://traefik.io`,
		Configuration: tConfig,
		Resources:     loaders,
		Run: func(_ []string) error {
			return runCmd(&tConfig.Configuration)
		},
	}

	err := cmdTraefik.AddCommand(healthcheck.NewCmd(&tConfig.Configuration, loaders))
	if err != nil {
		stdlog.Println(err)
		os.Exit(1)
	}

	err = cmdTraefik.AddCommand(cmdVersion.NewCmd())
	if err != nil {
		stdlog.Println(err)
		os.Exit(1)
	}

	err = cli.Execute(cmdTraefik)
	if err != nil {
		log.Error().Err(err).Msg("Command error")
		logrus.Exit(1)
	}

	logrus.Exit(0)
}

func runCmd(staticConfiguration *static.Configuration) error {
	setupLogger(staticConfiguration)

	http.DefaultTransport.(*http.Transport).Proxy = http.ProxyFromEnvironment

	staticConfiguration.SetEffectiveConfiguration()
	if err := staticConfiguration.ValidateConfiguration(); err != nil {
		return err
	}

	log.Info().Str("version", version.Version).
		Msgf("Traefik version %s built on %s", version.Version, version.BuildDate)

	jsonConf, err := json.Marshal(staticConfiguration)
	if err != nil {
		log.Error().Err(err).Msg("Could not marshal static configuration")
		log.Debug().Interface("staticConfiguration", staticConfiguration).Msg("Static configuration loaded [struct]")
	} else {
		log.Debug().RawJSON("staticConfiguration", jsonConf).Msg("Static configuration loaded [json]")
	}

	if staticConfiguration.Global.CheckNewVersion {
		checkNewVersion()
	}

	stats(staticConfiguration)

	svr, err := setupServer(staticConfiguration)
	if err != nil {
		return err
	}

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if staticConfiguration.Ping != nil {
		staticConfiguration.Ping.WithContext(ctx)
	}

	svr.Start(ctx)
	defer svr.Close()

	sent, err := daemon.SdNotify(false, "READY=1")
	if !sent && err != nil {
		log.Error().Err(err).Msg("Failed to notify")
	}

	t, err := daemon.SdWatchdogEnabled(false)
	if err != nil {
		log.Error().Err(err).Msg("Could not enable Watchdog")
	} else if t != 0 {
		// Send a ping each half time given
		t /= 2
		log.Info().Msgf("Watchdog activated with timer duration %s", t)
		safe.Go(func() {
			tick := time.Tick(t)
			for range tick {
				resp, errHealthCheck := healthcheck.Do(*staticConfiguration)
				if resp != nil {
					_ = resp.Body.Close()
				}

				if staticConfiguration.Ping == nil || errHealthCheck == nil {
					if ok, _ := daemon.SdNotify(false, "WATCHDOG=1"); !ok {
						log.Error().Msg("Fail to tick watchdog")
					}
				} else {
					log.Error().Err(errHealthCheck).Send()
				}
			}
		})
	}

	svr.Wait()
	log.Info().Msg("Shutting down")
	return nil
}

func setupServer(staticConfiguration *static.Configuration) (*server.Server, error) {
	providerAggregator := aggregator.NewProviderAggregator(*staticConfiguration.Providers)

	ctx := context.Background()
	routinesPool := safe.NewPool(ctx)

	// adds internal provider
	err := providerAggregator.AddProvider(traefik.New(*staticConfiguration))
	if err != nil {
		return nil, err
	}

	// ACME

	tlsManager := traefiktls.NewManager()
	httpChallengeProvider := acme.NewChallengeHTTP()

	tlsChallengeProvider := acme.NewChallengeTLSALPN()
	err = providerAggregator.AddProvider(tlsChallengeProvider)
	if err != nil {
		return nil, err
	}

	acmeProviders := initACMEProvider(staticConfiguration, &providerAggregator, tlsManager, httpChallengeProvider, tlsChallengeProvider)

	// Tailscale

	tsProviders := initTailscaleProviders(staticConfiguration, &providerAggregator)

	// Metrics

	metricRegistries := registerMetricClients(staticConfiguration.Metrics)
	metricsRegistry := metrics.NewMultiRegistry(metricRegistries)

	// Entrypoints

	serverEntryPointsTCP, err := server.NewTCPEntryPoints(staticConfiguration.EntryPoints, staticConfiguration.HostResolver, metricsRegistry)
	if err != nil {
		return nil, err
	}

	serverEntryPointsUDP, err := server.NewUDPEntryPoints(staticConfiguration.EntryPoints)
	if err != nil {
		return nil, err
	}

	// Plugins

	pluginBuilder, err := createPluginBuilder(staticConfiguration)
	if err != nil {
		log.Error().Err(err).Msg("Plugins are disabled because an error has occurred.")
	}

	// Providers plugins

	for name, conf := range staticConfiguration.Providers.Plugin {
		if pluginBuilder == nil {
			break
		}

		p, err := pluginBuilder.BuildProvider(name, conf)
		if err != nil {
			return nil, fmt.Errorf("plugin: failed to build provider: %w", err)
		}

		err = providerAggregator.AddProvider(p)
		if err != nil {
			return nil, fmt.Errorf("plugin: failed to add provider: %w", err)
		}
	}

	// Traefik Hub

	if staticConfiguration.Hub != nil {
		if err = providerAggregator.AddProvider(staticConfiguration.Hub); err != nil {
			return nil, fmt.Errorf("adding Traefik Hub provider: %w", err)
		}

		// API is mandatory for Traefik Hub to access the dynamic configuration.
		if staticConfiguration.API == nil {
			staticConfiguration.API = &static.API{}
		}
	}

	// Service manager factory

	var spiffeX509Source *workloadapi.X509Source
	if staticConfiguration.Spiffe != nil && staticConfiguration.Spiffe.WorkloadAPIAddr != "" {
		log.Info().Str("workloadAPIAddr", staticConfiguration.Spiffe.WorkloadAPIAddr).
			Msg("Waiting on SPIFFE SVID delivery")

		spiffeX509Source, err = workloadapi.NewX509Source(
			ctx,
			workloadapi.WithClientOptions(
				workloadapi.WithAddr(
					staticConfiguration.Spiffe.WorkloadAPIAddr,
				),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create SPIFFE x509 source: %w", err)
		}
		log.Info().Msg("Successfully obtained SPIFFE SVID.")
	}

	roundTripperManager := service.NewRoundTripperManager(spiffeX509Source)
	dialerManager := tcp.NewDialerManager(spiffeX509Source)
	acmeHTTPHandler := getHTTPChallengeHandler(acmeProviders, httpChallengeProvider)
	managerFactory := service.NewManagerFactory(*staticConfiguration, routinesPool, metricsRegistry, roundTripperManager, acmeHTTPHandler)

	// Router factory

	accessLog := setupAccessLog(staticConfiguration.AccessLog)
	tracer := setupTracing(staticConfiguration.Tracing)

	chainBuilder := middleware.NewChainBuilder(metricsRegistry, accessLog, tracer)
	routerFactory := server.NewRouterFactory(*staticConfiguration, managerFactory, tlsManager, chainBuilder, pluginBuilder, metricsRegistry, dialerManager)

	// Watcher

	watcher := server.NewConfigurationWatcher(
		routinesPool,
		providerAggregator,
		getDefaultsEntrypoints(staticConfiguration),
		"internal",
	)

	// TLS
	watcher.AddListener(func(conf dynamic.Configuration) {
		ctx := context.Background()
		tlsManager.UpdateConfigs(ctx, conf.TLS.Stores, conf.TLS.Options, conf.TLS.Certificates)

		gauge := metricsRegistry.TLSCertsNotAfterTimestampGauge()
		for _, certificate := range tlsManager.GetServerCertificates() {
			appendCertMetric(gauge, certificate)
		}
	})

	// Metrics
	watcher.AddListener(func(_ dynamic.Configuration) {
		metricsRegistry.ConfigReloadsCounter().Add(1)
		metricsRegistry.LastConfigReloadSuccessGauge().Set(float64(time.Now().Unix()))
	})

	// Server Transports
	watcher.AddListener(func(conf dynamic.Configuration) {
		roundTripperManager.Update(conf.HTTP.ServersTransports)
		dialerManager.Update(conf.TCP.ServersTransports)
	})

	// Switch router
	watcher.AddListener(switchRouter(routerFactory, serverEntryPointsTCP, serverEntryPointsUDP))

	// Metrics
	if metricsRegistry.IsEpEnabled() || metricsRegistry.IsSvcEnabled() {
		var eps []string
		for key := range serverEntryPointsTCP {
			eps = append(eps, key)
		}
		watcher.AddListener(func(conf dynamic.Configuration) {
			metrics.OnConfigurationUpdate(conf, eps)
		})
	}

	// TLS challenge
	watcher.AddListener(tlsChallengeProvider.ListenConfiguration)

	// Certificate Resolvers

	resolverNames := map[string]struct{}{}

	// ACME
	for _, p := range acmeProviders {
		resolverNames[p.ResolverName] = struct{}{}
		watcher.AddListener(p.ListenConfiguration)
	}

	// Tailscale
	for _, p := range tsProviders {
		resolverNames[p.ResolverName] = struct{}{}
		watcher.AddListener(p.HandleConfigUpdate)
	}

	// Certificate resolver logs
	watcher.AddListener(func(config dynamic.Configuration) {
		for rtName, rt := range config.HTTP.Routers {
			if rt.TLS == nil || rt.TLS.CertResolver == "" {
				continue
			}

			if _, ok := resolverNames[rt.TLS.CertResolver]; !ok &&
				// "traefik-hub" is an allowed certificate resolver name in a Traefik Hub Experimental feature context.
				// It is used to activate its own certificate resolution, even though it is not a "classical" traefik certificate resolver.
				(staticConfiguration.Hub == nil || rt.TLS.CertResolver != "traefik-hub") {
				log.Error().Err(err).Str(logs.RouterName, rtName).Str("certificateResolver", rt.TLS.CertResolver).
					Msg("Router uses a non-existent certificate resolver")
			}
		}
	})

	return server.NewServer(routinesPool, serverEntryPointsTCP, serverEntryPointsUDP, watcher, chainBuilder, accessLog), nil
}

func getHTTPChallengeHandler(acmeProviders []*acme.Provider, httpChallengeProvider http.Handler) http.Handler {
	var acmeHTTPHandler http.Handler
	for _, p := range acmeProviders {
		if p != nil && p.HTTPChallenge != nil {
			acmeHTTPHandler = httpChallengeProvider
			break
		}
	}
	return acmeHTTPHandler
}

func getDefaultsEntrypoints(staticConfiguration *static.Configuration) []string {
	var defaultEntryPoints []string

	// Determines if at least one EntryPoint is configured to be used by default.
	var hasDefinedDefaults bool
	for _, ep := range staticConfiguration.EntryPoints {
		if ep.AsDefault {
			hasDefinedDefaults = true
			break
		}
	}

	for name, cfg := range staticConfiguration.EntryPoints {
		// By default all entrypoints are considered.
		// If at least one is flagged, then only flagged entrypoints are included.
		if hasDefinedDefaults && !cfg.AsDefault {
			continue
		}

		// Traefik Hub entryPoint should not be used as a default entryPoint.
		if hub.APIEntrypoint == name || hub.TunnelEntrypoint == name {
			continue
		}

		protocol, err := cfg.GetProtocol()
		if err != nil {
			// Should never happen because Traefik should not start if protocol is invalid.
			log.Error().Err(err).Msg("Invalid protocol")
		}

		if protocol != "udp" && name != static.DefaultInternalEntryPointName {
			defaultEntryPoints = append(defaultEntryPoints, name)
		}
	}

	sort.Strings(defaultEntryPoints)
	return defaultEntryPoints
}

func switchRouter(routerFactory *server.RouterFactory, serverEntryPointsTCP server.TCPEntryPoints, serverEntryPointsUDP server.UDPEntryPoints) func(conf dynamic.Configuration) {
	return func(conf dynamic.Configuration) {
		rtConf := runtime.NewConfig(conf)

		routers, udpRouters := routerFactory.CreateRouters(rtConf)

		serverEntryPointsTCP.Switch(routers)
		serverEntryPointsUDP.Switch(udpRouters)
	}
}

// initACMEProvider creates and registers acme.Provider instances corresponding to the configured ACME certificate resolvers.
func initACMEProvider(c *static.Configuration, providerAggregator *aggregator.ProviderAggregator, tlsManager *traefiktls.Manager, httpChallengeProvider, tlsChallengeProvider challenge.Provider) []*acme.Provider {
	localStores := map[string]*acme.LocalStore{}

	var resolvers []*acme.Provider
	for name, resolver := range c.CertificatesResolvers {
		if resolver.ACME == nil {
			continue
		}

		if localStores[resolver.ACME.Storage] == nil {
			localStores[resolver.ACME.Storage] = acme.NewLocalStore(resolver.ACME.Storage)
		}

		p := &acme.Provider{
			Configuration:         resolver.ACME,
			Store:                 localStores[resolver.ACME.Storage],
			ResolverName:          name,
			HTTPChallengeProvider: httpChallengeProvider,
			TLSChallengeProvider:  tlsChallengeProvider,
		}

		if err := providerAggregator.AddProvider(p); err != nil {
			log.Error().Err(err).Str("resolver", name).Msg("The ACME resolve is skipped from the resolvers list")
			continue
		}

		p.SetTLSManager(tlsManager)

		p.SetConfigListenerChan(make(chan dynamic.Configuration))

		resolvers = append(resolvers, p)
	}

	return resolvers
}

// initTailscaleProviders creates and registers tailscale.Provider instances corresponding to the configured Tailscale certificate resolvers.
func initTailscaleProviders(cfg *static.Configuration, providerAggregator *aggregator.ProviderAggregator) []*tailscale.Provider {
	var providers []*tailscale.Provider
	for name, resolver := range cfg.CertificatesResolvers {
		if resolver.Tailscale == nil {
			continue
		}

		tsProvider := &tailscale.Provider{ResolverName: name}

		if err := providerAggregator.AddProvider(tsProvider); err != nil {
			log.Error().Err(err).Str(logs.ProviderName, name).Msg("Unable to create Tailscale provider")
			continue
		}

		providers = append(providers, tsProvider)
	}

	return providers
}

func registerMetricClients(metricsConfig *types.Metrics) []metrics.Registry {
	if metricsConfig == nil {
		return nil
	}

	var registries []metrics.Registry

	if metricsConfig.Prometheus != nil {
		logger := log.With().Str(logs.MetricsProviderName, "prometheus").Logger()

		prometheusRegister := metrics.RegisterPrometheus(logger.WithContext(context.Background()), metricsConfig.Prometheus)
		if prometheusRegister != nil {
			registries = append(registries, prometheusRegister)
			logger.Debug().Msg("Configured Prometheus metrics")
		}
	}

	if metricsConfig.Datadog != nil {
		logger := log.With().Str(logs.MetricsProviderName, "datadog").Logger()

		registries = append(registries, metrics.RegisterDatadog(logger.WithContext(context.Background()), metricsConfig.Datadog))
		logger.Debug().
			Str("address", metricsConfig.Datadog.Address).
			Str("pushInterval", metricsConfig.Datadog.PushInterval.String()).
			Msgf("Configured Datadog metrics")
	}

	if metricsConfig.StatsD != nil {
		logger := log.With().Str(logs.MetricsProviderName, "statsd").Logger()

		registries = append(registries, metrics.RegisterStatsd(logger.WithContext(context.Background()), metricsConfig.StatsD))
		logger.Debug().
			Str("address", metricsConfig.StatsD.Address).
			Str("pushInterval", metricsConfig.StatsD.PushInterval.String()).
			Msg("Configured StatsD metrics")
	}

	if metricsConfig.InfluxDB2 != nil {
		logger := log.With().Str(logs.MetricsProviderName, "influxdb2").Logger()

		influxDB2Register := metrics.RegisterInfluxDB2(logger.WithContext(context.Background()), metricsConfig.InfluxDB2)
		if influxDB2Register != nil {
			registries = append(registries, influxDB2Register)
			logger.Debug().
				Str("address", metricsConfig.InfluxDB2.Address).
				Str("bucket", metricsConfig.InfluxDB2.Bucket).
				Str("organization", metricsConfig.InfluxDB2.Org).
				Str("pushInterval", metricsConfig.InfluxDB2.PushInterval.String()).
				Msg("Configured InfluxDB v2 metrics")
		}
	}

	if metricsConfig.OpenTelemetry != nil {
		logger := log.With().Str(logs.MetricsProviderName, "openTelemetry").Logger()

		openTelemetryRegistry := metrics.RegisterOpenTelemetry(logger.WithContext(context.Background()), metricsConfig.OpenTelemetry)
		if openTelemetryRegistry != nil {
			registries = append(registries, openTelemetryRegistry)
			logger.Debug().
				Str("address", metricsConfig.OpenTelemetry.Address).
				Str("pushInterval", metricsConfig.OpenTelemetry.PushInterval.String()).
				Msg("Configured OpenTelemetry metrics")
		}
	}

	return registries
}

func appendCertMetric(gauge gokitmetrics.Gauge, certificate *x509.Certificate) {
	sort.Strings(certificate.DNSNames)

	labels := []string{
		"cn", certificate.Subject.CommonName,
		"serial", certificate.SerialNumber.String(),
		"sans", strings.Join(certificate.DNSNames, ","),
	}

	notAfter := float64(certificate.NotAfter.Unix())

	gauge.With(labels...).Set(notAfter)
}

func setupAccessLog(conf *types.AccessLog) *accesslog.Handler {
	if conf == nil {
		return nil
	}

	accessLoggerMiddleware, err := accesslog.NewHandler(conf)
	if err != nil {
		log.Warn().Err(err).Msg("Unable to create access logger")
		return nil
	}

	return accessLoggerMiddleware
}

func setupTracing(conf *static.Tracing) *tracing.Tracing {
	if conf == nil {
		return nil
	}

	var backend tracing.Backend

	if conf.Jaeger != nil {
		backend = conf.Jaeger
	}

	if conf.Zipkin != nil {
		if backend != nil {
			log.Error().Msg("Multiple tracing backend are not supported: cannot create Zipkin backend.")
		} else {
			backend = conf.Zipkin
		}
	}

	if conf.Datadog != nil {
		if backend != nil {
			log.Error().Msg("Multiple tracing backend are not supported: cannot create Datadog backend.")
		} else {
			backend = conf.Datadog
		}
	}

	if conf.Instana != nil {
		if backend != nil {
			log.Error().Msg("Multiple tracing backend are not supported: cannot create Instana backend.")
		} else {
			backend = conf.Instana
		}
	}

	if conf.Haystack != nil {
		if backend != nil {
			log.Error().Msg("Multiple tracing backend are not supported: cannot create Haystack backend.")
		} else {
			backend = conf.Haystack
		}
	}

	if conf.Elastic != nil {
		if backend != nil {
			log.Error().Msg("Multiple tracing backend are not supported: cannot create Elastic backend.")
		} else {
			backend = conf.Elastic
		}
	}

	if conf.OpenTelemetry != nil {
		if backend != nil {
			log.Error().Msg("Tracing backends are all mutually exclusive: cannot create OpenTelemetry backend.")
		} else {
			backend = conf.OpenTelemetry
		}
	}

	if backend == nil {
		log.Debug().Msg("Could not initialize tracing, using Jaeger by default")
		defaultBackend := &jaeger.Config{}
		defaultBackend.SetDefaults()
		backend = defaultBackend
	}

	tracer, err := tracing.NewTracing(conf.ServiceName, conf.SpanNameLimit, backend)
	if err != nil {
		log.Warn().Err(err).Msg("Unable to create tracer")
		return nil
	}
	return tracer
}

func checkNewVersion() {
	ticker := time.Tick(24 * time.Hour)
	safe.Go(func() {
		for time.Sleep(10 * time.Minute); ; <-ticker {
			version.CheckNewVersion()
		}
	})
}

func stats(staticConfiguration *static.Configuration) {
	logger := log.Info()

	if staticConfiguration.Global.SendAnonymousUsage {
		logger.Msg(`Stats collection is enabled.`)
		logger.Msg(`Many thanks for contributing to Traefik's improvement by allowing us to receive anonymous information from your configuration.`)
		logger.Msg(`Help us improve Traefik by leaving this feature on :)`)
		logger.Msg(`More details on: https://doc.traefik.io/traefik/contributing/data-collection/`)
		collect(staticConfiguration)
	} else {
		logger.Msg(`
Stats collection is disabled.
Help us improve Traefik by turning this feature on :)
More details on: https://doc.traefik.io/traefik/contributing/data-collection/
`)
	}
}

func collect(staticConfiguration *static.Configuration) {
	ticker := time.Tick(24 * time.Hour)
	safe.Go(func() {
		for time.Sleep(10 * time.Minute); ; <-ticker {
			if err := collector.Collect(staticConfiguration); err != nil {
				log.Debug().Err(err).Send()
			}
		}
	})
}
