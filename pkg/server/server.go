package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/containous/traefik/v2/pkg/config/static"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/metrics"
	"github.com/containous/traefik/v2/pkg/middlewares/accesslog"
	"github.com/containous/traefik/v2/pkg/middlewares/requestdecorator"
	"github.com/containous/traefik/v2/pkg/provider"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/containous/traefik/v2/pkg/server/middleware"
	"github.com/containous/traefik/v2/pkg/tls"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/containous/traefik/v2/pkg/tracing/jaeger"
	"github.com/containous/traefik/v2/pkg/types"
)

// Server is the reverse-proxy/load-balancer engine
type Server struct {
	entryPointsTCP             TCPEntryPoints
	configurationChan          chan dynamic.Message
	configurationValidatedChan chan dynamic.Message
	signals                    chan os.Signal
	stopChan                   chan bool
	currentConfigurations      safe.Safe
	providerConfigUpdateMap    map[string]chan dynamic.Message
	accessLoggerMiddleware     *accesslog.Handler
	tracer                     *tracing.Tracing
	routinesPool               *safe.Pool
	defaultRoundTripper        http.RoundTripper
	metricsRegistry            metrics.Registry
	provider                   provider.Provider
	configurationListeners     []func(dynamic.Configuration)
	requestDecorator           *requestdecorator.RequestDecorator
	providersThrottleDuration  time.Duration
	tlsManager                 *tls.Manager
}

// RouteAppenderFactory the route appender factory interface
type RouteAppenderFactory interface {
	NewAppender(ctx context.Context, middlewaresBuilder *middleware.Builder, runtimeConfiguration *runtime.Configuration) types.RouteAppender
}

func setupTracing(conf *static.Tracing) tracing.Backend {
	var backend tracing.Backend

	if conf.Jaeger != nil {
		backend = conf.Jaeger
	}

	if conf.Zipkin != nil {
		if backend != nil {
			log.WithoutContext().Error("Multiple tracing backend are not supported: cannot create Zipkin backend.")
		} else {
			backend = conf.Zipkin
		}
	}

	if conf.DataDog != nil {
		if backend != nil {
			log.WithoutContext().Error("Multiple tracing backend are not supported: cannot create DataDog backend.")
		} else {
			backend = conf.DataDog
		}
	}

	if conf.Instana != nil {
		if backend != nil {
			log.WithoutContext().Error("Multiple tracing backend are not supported: cannot create Instana backend.")
		} else {
			backend = conf.Instana
		}
	}

	if conf.Haystack != nil {
		if backend != nil {
			log.WithoutContext().Error("Multiple tracing backend are not supported: cannot create Haystack backend.")
		} else {
			backend = conf.Haystack
		}
	}

	if backend == nil {
		log.WithoutContext().Debug("Could not initialize tracing, use Jaeger by default")
		backend := &jaeger.Config{}
		backend.SetDefaults()
	}

	return backend
}

// NewServer returns an initialized Server.
func NewServer(staticConfiguration static.Configuration, provider provider.Provider, entryPoints TCPEntryPoints, tlsManager *tls.Manager) *Server {
	server := &Server{}

	server.provider = provider
	server.entryPointsTCP = entryPoints
	server.configurationChan = make(chan dynamic.Message, 100)
	server.configurationValidatedChan = make(chan dynamic.Message, 100)
	server.signals = make(chan os.Signal, 1)
	server.stopChan = make(chan bool, 1)
	server.configureSignals()
	currentConfigurations := make(dynamic.Configurations)
	server.currentConfigurations.Set(currentConfigurations)
	server.providerConfigUpdateMap = make(map[string]chan dynamic.Message)
	server.tlsManager = tlsManager

	if staticConfiguration.Providers != nil {
		server.providersThrottleDuration = time.Duration(staticConfiguration.Providers.ProvidersThrottleDuration)
	}

	transport, err := createHTTPTransport(staticConfiguration.ServersTransport)
	if err != nil {
		log.WithoutContext().Errorf("Could not configure HTTP Transport, fallbacking on default transport: %v", err)
		server.defaultRoundTripper = http.DefaultTransport
	} else {
		server.defaultRoundTripper = transport
	}

	server.routinesPool = safe.NewPool(context.Background())

	if staticConfiguration.Tracing != nil {
		tracingBackend := setupTracing(staticConfiguration.Tracing)
		if tracingBackend != nil {
			server.tracer, err = tracing.NewTracing(staticConfiguration.Tracing.ServiceName, staticConfiguration.Tracing.SpanNameLimit, tracingBackend)
			if err != nil {
				log.WithoutContext().Warnf("Unable to create tracer: %v", err)
			}
		}
	}

	server.requestDecorator = requestdecorator.New(staticConfiguration.HostResolver)

	server.metricsRegistry = registerMetricClients(staticConfiguration.Metrics)

	if staticConfiguration.AccessLog != nil {
		var err error
		server.accessLoggerMiddleware, err = accesslog.NewHandler(staticConfiguration.AccessLog)
		if err != nil {
			log.WithoutContext().Warnf("Unable to create access logger : %v", err)
		}
	}
	return server
}

// Start starts the server and Stop/Close it when context is Done
func (s *Server) Start(ctx context.Context) {
	go func() {
		defer s.Close()
		<-ctx.Done()
		logger := log.FromContext(ctx)
		logger.Info("I have to go...")
		logger.Info("Stopping server gracefully")
		s.Stop()
	}()

	s.startTCPServers()
	s.routinesPool.Go(func(stop chan bool) {
		s.listenProviders(stop)
	})
	s.routinesPool.Go(func(stop chan bool) {
		s.listenConfigurations(stop)
	})
	s.startProvider()
	s.routinesPool.Go(func(stop chan bool) {
		s.listenSignals(stop)
	})
}

// Wait blocks until server is shutted down.
func (s *Server) Wait() {
	<-s.stopChan
}

// Stop stops the server
func (s *Server) Stop() {
	defer log.WithoutContext().Info("Server stopped")

	var wg sync.WaitGroup
	for epn, ep := range s.entryPointsTCP {
		wg.Add(1)
		go func(entryPointName string, entryPoint *TCPEntryPoint) {
			ctx := log.With(context.Background(), log.Str(log.EntryPointName, entryPointName))
			defer wg.Done()

			entryPoint.Shutdown(ctx)

			log.FromContext(ctx).Debugf("Entry point %s closed", entryPointName)
		}(epn, ep)
	}
	wg.Wait()
	s.stopChan <- true
}

// Close destroys the server
func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func(ctx context.Context) {
		<-ctx.Done()
		if ctx.Err() == context.Canceled {
			return
		} else if ctx.Err() == context.DeadlineExceeded {
			panic("Timeout while stopping traefik, killing instance ✝")
		}
	}(ctx)

	stopMetricsClients()
	s.routinesPool.Cleanup()
	close(s.configurationChan)
	close(s.configurationValidatedChan)
	signal.Stop(s.signals)
	close(s.signals)
	close(s.stopChan)

	if s.accessLoggerMiddleware != nil {
		if err := s.accessLoggerMiddleware.Close(); err != nil {
			log.WithoutContext().Errorf("Could not close the access log file: %s", err)
		}
	}

	if s.tracer != nil {
		s.tracer.Close()
	}

	cancel()
}

func (s *Server) startTCPServers() {
	// Use an empty configuration in order to initialize the default handlers with internal routes
	routers := s.loadConfigurationTCP(dynamic.Configurations{})
	for entryPointName, router := range routers {
		s.entryPointsTCP[entryPointName].switchRouter(router)
	}

	for entryPointName, serverEntryPoint := range s.entryPointsTCP {
		ctx := log.With(context.Background(), log.Str(log.EntryPointName, entryPointName))
		go serverEntryPoint.startTCP(ctx)
	}
}

func (s *Server) listenProviders(stop chan bool) {
	for {
		select {
		case <-stop:
			return
		case configMsg, ok := <-s.configurationChan:
			if !ok {
				return
			}
			if configMsg.Configuration != nil {
				s.preLoadConfiguration(configMsg)
			} else {
				log.Debugf("Received nil configuration from provider %q, skipping.", configMsg.ProviderName)
			}
		}
	}
}

// AddListener adds a new listener function used when new configuration is provided
func (s *Server) AddListener(listener func(dynamic.Configuration)) {
	if s.configurationListeners == nil {
		s.configurationListeners = make([]func(dynamic.Configuration), 0)
	}
	s.configurationListeners = append(s.configurationListeners, listener)
}

func (s *Server) startProvider() {
	jsonConf, err := json.Marshal(s.provider)
	if err != nil {
		log.WithoutContext().Debugf("Unable to marshal provider configuration %T: %v", s.provider, err)
	}

	log.WithoutContext().Infof("Starting provider %T %s", s.provider, jsonConf)
	currentProvider := s.provider

	safe.Go(func() {
		err := currentProvider.Provide(s.configurationChan, s.routinesPool)
		if err != nil {
			log.WithoutContext().Errorf("Error starting provider %T: %s", s.provider, err)
		}
	})
}

func registerMetricClients(metricsConfig *types.Metrics) metrics.Registry {
	if metricsConfig == nil {
		return metrics.NewVoidRegistry()
	}

	var registries []metrics.Registry

	if metricsConfig.Prometheus != nil {
		ctx := log.With(context.Background(), log.Str(log.MetricsProviderName, "prometheus"))
		prometheusRegister := metrics.RegisterPrometheus(ctx, metricsConfig.Prometheus)
		if prometheusRegister != nil {
			registries = append(registries, prometheusRegister)
			log.FromContext(ctx).Debug("Configured Prometheus metrics")
		}
	}

	if metricsConfig.DataDog != nil {
		ctx := log.With(context.Background(), log.Str(log.MetricsProviderName, "datadog"))
		registries = append(registries, metrics.RegisterDatadog(ctx, metricsConfig.DataDog))
		log.FromContext(ctx).Debugf("Configured DataDog metrics: pushing to %s once every %s",
			metricsConfig.DataDog.Address, metricsConfig.DataDog.PushInterval)
	}

	if metricsConfig.StatsD != nil {
		ctx := log.With(context.Background(), log.Str(log.MetricsProviderName, "statsd"))
		registries = append(registries, metrics.RegisterStatsd(ctx, metricsConfig.StatsD))
		log.FromContext(ctx).Debugf("Configured StatsD metrics: pushing to %s once every %s",
			metricsConfig.StatsD.Address, metricsConfig.StatsD.PushInterval)
	}

	if metricsConfig.InfluxDB != nil {
		ctx := log.With(context.Background(), log.Str(log.MetricsProviderName, "influxdb"))
		registries = append(registries, metrics.RegisterInfluxDB(ctx, metricsConfig.InfluxDB))
		log.FromContext(ctx).Debugf("Configured InfluxDB metrics: pushing to %s once every %s",
			metricsConfig.InfluxDB.Address, metricsConfig.InfluxDB.PushInterval)
	}

	return metrics.NewMultiRegistry(registries)
}

func stopMetricsClients() {
	metrics.StopDatadog()
	metrics.StopStatsd()
	metrics.StopInfluxDB()
}
