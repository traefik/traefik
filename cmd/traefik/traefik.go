package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"
	stdlog "log"
	"maps"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/v22/daemon"
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
	"github.com/traefik/traefik/v3/pkg/provider/tailscale"
	"github.com/traefik/traefik/v3/pkg/provider/traefik"
	"github.com/traefik/traefik/v3/pkg/proxy"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"github.com/traefik/traefik/v3/pkg/redactor"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/server"
	"github.com/traefik/traefik/v3/pkg/server/middleware"
	"github.com/traefik/traefik/v3/pkg/server/service"
	"github.com/traefik/traefik/v3/pkg/tcp"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"github.com/traefik/traefik/v3/pkg/types"
	"github.com/traefik/traefik/v3/pkg/version"
)

func main() {
	// traefik config inits
	tConfig := cmd.NewTraefikConfiguration()

	loaders := []cli.ResourceLoader{&tcli.DeprecationLoader{}, &tcli.FileLoader{}, &tcli.FlagLoader{}, &tcli.EnvLoader{}}

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
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := setupLogger(ctx, staticConfiguration); err != nil {
		return fmt.Errorf("setting up logger: %w", err)
	}

	http.DefaultTransport.(*http.Transport).Proxy = http.ProxyFromEnvironment

	staticConfiguration.SetEffectiveConfiguration()
	if err := staticConfiguration.ValidateConfiguration(); err != nil {
		return err
	}

	log.Info().Str("version", version.Version).
		Msgf("Traefik version %s built on %s", version.Version, version.BuildDate)

	redactedStaticConfiguration, err := redactor.RemoveCredentials(staticConfiguration)
	if err != nil {
		log.Error().Err(err).Msg("Could not redact static configuration")
	} else {
		log.Debug().RawJSON("staticConfiguration", []byte(redactedStaticConfiguration)).Msg("Static configuration loaded [json]")
	}

	if staticConfiguration.Global.CheckNewVersion {
		checkNewVersion()
	}

	stats(staticConfiguration)

	svr, err := setupServer(staticConfiguration)
	if err != nil {
		return err
	}

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

	tlsManager := traefiktls.NewManager(staticConfiguration.OCSP)
	routinesPool.GoCtx(tlsManager.Run)

	httpChallengeProvider := acme.NewChallengeHTTP()

	tlsChallengeProvider := acme.NewChallengeTLSALPN()
	err = providerAggregator.AddProvider(tlsChallengeProvider)
	if err != nil {
		return nil, err
	}

	acmeProviders := initACMEProvider(staticConfiguration, providerAggregator, tlsManager, httpChallengeProvider, tlsChallengeProvider, routinesPool)

	// Tailscale

	tsProviders := initTailscaleProviders(staticConfiguration, providerAggregator)

	// Observability

	metricRegistries := registerMetricClients(staticConfiguration.Metrics)
	var semConvMetricRegistry *metrics.SemConvMetricsRegistry
	if staticConfiguration.Metrics != nil && staticConfiguration.Metrics.OTLP != nil {
		semConvMetricRegistry, err = metrics.NewSemConvMetricRegistry(ctx, staticConfiguration.Metrics.OTLP)
		if err != nil {
			return nil, fmt.Errorf("unable to create SemConv metric registry: %w", err)
		}
	}
	metricsRegistry := metrics.NewMultiRegistry(metricRegistries)
	accessLog := setupAccessLog(ctx, staticConfiguration.AccessLog)
	tracer, tracerCloser := setupTracing(ctx, staticConfiguration.Tracing)
	observabilityMgr := middleware.NewObservabilityMgr(*staticConfiguration, metricsRegistry, semConvMetricRegistry, accessLog, tracer, tracerCloser)

	// Entrypoints

	serverEntryPointsTCP, err := server.NewTCPEntryPoints(staticConfiguration.EntryPoints, staticConfiguration.HostResolver, metricsRegistry)
	if err != nil {
		return nil, err
	}

	serverEntryPointsUDP, err := server.NewUDPEntryPoints(staticConfiguration.EntryPoints)
	if err != nil {
		return nil, err
	}

	if staticConfiguration.API != nil {
		version.DisableDashboardAd = staticConfiguration.API.DisableDashboardAd
	}

	// Plugins
	pluginLogger := log.Ctx(ctx).With().Logger()
	hasPlugins := staticConfiguration.Experimental != nil && (staticConfiguration.Experimental.Plugins != nil || staticConfiguration.Experimental.LocalPlugins != nil)
	if hasPlugins {
		pluginsList := slices.Collect(maps.Keys(staticConfiguration.Experimental.Plugins))
		pluginsList = append(pluginsList, slices.Collect(maps.Keys(staticConfiguration.Experimental.LocalPlugins))...)

		pluginLogger = pluginLogger.With().Strs("plugins", pluginsList).Logger()
		pluginLogger.Info().Msg("Loading plugins...")
	}

	pluginBuilder, err := createPluginBuilder(staticConfiguration)
	if err != nil && staticConfiguration.Experimental != nil && staticConfiguration.Experimental.AbortOnPluginFailure {
		return nil, fmt.Errorf("plugin: failed to create plugin builder: %w", err)
	}
	if err != nil {
		pluginLogger.Err(err).Msg("Plugins are disabled because an error has occurred.")
	} else if hasPlugins {
		pluginLogger.Info().Msg("Plugins loaded.")
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

	transportManager := service.NewTransportManager(spiffeX509Source)

	var proxyBuilder service.ProxyBuilder = httputil.NewProxyBuilder(transportManager, semConvMetricRegistry)
	if staticConfiguration.Experimental != nil && staticConfiguration.Experimental.FastProxy != nil {
		proxyBuilder = proxy.NewSmartBuilder(transportManager, proxyBuilder, *staticConfiguration.Experimental.FastProxy)
	}

	dialerManager := tcp.NewDialerManager(spiffeX509Source)
	acmeHTTPHandler := getHTTPChallengeHandler(acmeProviders, httpChallengeProvider)
	managerFactory := service.NewManagerFactory(*staticConfiguration, routinesPool, observabilityMgr, transportManager, proxyBuilder, acmeHTTPHandler)

	// Router factory

	routerFactory, err := server.NewRouterFactory(*staticConfiguration, managerFactory, tlsManager, observabilityMgr, pluginBuilder, dialerManager)
	if err != nil {
		return nil, fmt.Errorf("creating router factory: %w", err)
	}

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
		transportManager.Update(conf.HTTP.ServersTransports)
		proxyBuilder.Update(conf.HTTP.ServersTransports)
		dialerManager.Update(conf.TCP.ServersTransports)
	})

	// Switch router
	watcher.AddListener(switchRouter(routerFactory, serverEntryPointsTCP, serverEntryPointsUDP))

	// Metrics
	if metricsRegistry.IsEpEnabled() || metricsRegistry.IsRouterEnabled() || metricsRegistry.IsSvcEnabled() {
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

			if _, ok := resolverNames[rt.TLS.CertResolver]; !ok {
				log.Error().Err(err).Str(logs.RouterName, rtName).Str("certificateResolver", rt.TLS.CertResolver).
					Msg("Router uses a nonexistent certificate resolver")
			}
		}
	})

	return server.NewServer(routinesPool, serverEntryPointsTCP, serverEntryPointsUDP, watcher, observabilityMgr), nil
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

		protocol, err := cfg.GetProtocol()
		if err != nil {
			// Should never happen because Traefik should not start if protocol is invalid.
			log.Error().Err(err).Msg("Invalid protocol")
		}

		if protocol != "udp" && name != static.DefaultInternalEntryPointName {
			defaultEntryPoints = append(defaultEntryPoints, name)
		}
	}

	slices.Sort(defaultEntryPoints)
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
func initACMEProvider(c *static.Configuration, providerAggregator *aggregator.ProviderAggregator, tlsManager *traefiktls.Manager, httpChallengeProvider, tlsChallengeProvider challenge.Provider, routinesPool *safe.Pool) []*acme.Provider {
	localStores := map[string]*acme.LocalStore{}

	var resolvers []*acme.Provider
	for name, resolver := range c.CertificatesResolvers {
		if resolver.ACME == nil {
			continue
		}

		if localStores[resolver.ACME.Storage] == nil {
			localStores[resolver.ACME.Storage] = acme.NewLocalStore(resolver.ACME.Storage, routinesPool)
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

	if metricsConfig.OTLP != nil {
		logger := log.With().Str(logs.MetricsProviderName, "openTelemetry").Logger()

		openTelemetryRegistry := metrics.RegisterOpenTelemetry(logger.WithContext(context.Background()), metricsConfig.OTLP)
		if openTelemetryRegistry != nil {
			registries = append(registries, openTelemetryRegistry)
			logger.Debug().
				Str("pushInterval", metricsConfig.OTLP.PushInterval.String()).
				Msg("Configured OpenTelemetry metrics")
		}
	}

	return registries
}

func appendCertMetric(gauge gokitmetrics.Gauge, certificate *x509.Certificate) {
	slices.Sort(certificate.DNSNames)

	labels := []string{
		"cn", certificate.Subject.CommonName,
		"serial", certificate.SerialNumber.String(),
		"sans", strings.Join(certificate.DNSNames, ","),
	}

	notAfter := float64(certificate.NotAfter.Unix())

	gauge.With(labels...).Set(notAfter)
}

func setupAccessLog(ctx context.Context, conf *types.AccessLog) *accesslog.Handler {
	if conf == nil {
		return nil
	}

	accessLoggerMiddleware, err := accesslog.NewHandler(ctx, conf)
	if err != nil {
		log.Warn().Err(err).Msg("Unable to create access logger")
		return nil
	}

	return accessLoggerMiddleware
}

func setupTracing(ctx context.Context, conf *static.Tracing) (*tracing.Tracer, io.Closer) {
	if conf == nil {
		return nil, nil
	}

	tracer, closer, err := tracing.NewTracing(ctx, conf)
	if err != nil {
		log.Warn().Err(err).Msg("Unable to create tracer")
		return nil, nil
	}

	return tracer, closer
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
	logger := log.With().Logger()

	if staticConfiguration.Global.SendAnonymousUsage {
		logger.Info().Msg(`Stats collection is enabled.`)
		logger.Info().Msg(`Many thanks for contributing to Traefik's improvement by allowing us to receive anonymous information from your configuration.`)
		logger.Info().Msg(`Help us improve Traefik by leaving this feature on :)`)
		logger.Info().Msg(`More details on: https://doc.traefik.io/traefik/contributing/data-collection/`)
		collect(staticConfiguration)
	} else {
		logger.Info().Msg(`
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
