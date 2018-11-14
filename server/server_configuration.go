package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/containous/alice"
	"github.com/containous/mux"
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/config/static"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/middlewares/accesslog"
	"github.com/containous/traefik/middlewares/requestdecorator"
	"github.com/containous/traefik/middlewares/tracing"
	"github.com/containous/traefik/old/configuration"
	"github.com/containous/traefik/responsemodifiers"
	"github.com/containous/traefik/server/middleware"
	"github.com/containous/traefik/server/router"
	"github.com/containous/traefik/server/service"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/tls/generate"
	"github.com/eapache/channels"
	"github.com/sirupsen/logrus"
)

// loadConfiguration manages dynamically frontends, backends and TLS configurations
func (s *Server) loadConfiguration(configMsg config.Message) {
	logger := log.FromContext(log.With(context.Background(), log.Str(log.ProviderName, configMsg.ProviderName)))

	currentConfigurations := s.currentConfigurations.Get().(config.Configurations)

	// Copy configurations to new map so we don't change current if LoadConfig fails
	newConfigurations := make(config.Configurations)
	for k, v := range currentConfigurations {
		newConfigurations[k] = v
	}
	newConfigurations[configMsg.ProviderName] = configMsg.Configuration

	s.metricsRegistry.ConfigReloadsCounter().Add(1)

	handlers, certificates := s.loadConfig(newConfigurations, s.globalConfiguration)

	s.metricsRegistry.LastConfigReloadSuccessGauge().Set(float64(time.Now().Unix()))

	for entryPointName, handler := range handlers {
		s.serverEntryPoints[entryPointName].httpRouter.UpdateHandler(handler)
	}

	for entryPointName, serverEntryPoint := range s.serverEntryPoints {
		eLogger := logger.WithField(log.EntryPointName, entryPointName)
		if s.entryPoints[entryPointName].Configuration.TLS == nil {
			if len(certificates[entryPointName]) > 0 {
				eLogger.Debugf("Cannot configure certificates for the non-TLS %s entryPoint.", entryPointName)
			}
		} else {
			serverEntryPoint.certs.DynamicCerts.Set(certificates[entryPointName])
			serverEntryPoint.certs.ResetCache()
		}
		eLogger.Infof("Server configuration reloaded on %s", s.serverEntryPoints[entryPointName].httpServer.Addr)
	}

	s.currentConfigurations.Set(newConfigurations)

	for _, listener := range s.configurationListeners {
		listener(*configMsg.Configuration)
	}

	s.postLoadConfiguration()
}

// loadConfig returns a new gorilla.mux Route from the specified global configuration and the dynamic
// provider configurations.
func (s *Server) loadConfig(configurations config.Configurations, globalConfiguration configuration.GlobalConfiguration) (map[string]http.Handler, map[string]map[string]*tls.Certificate) {

	ctx := context.TODO()

	// FIXME manage duplicates
	conf := config.Configuration{
		Routers:     make(map[string]*config.Router),
		Middlewares: make(map[string]*config.Middleware),
		Services:    make(map[string]*config.Service),
	}
	for _, config := range configurations {
		for key, value := range config.Middlewares {
			conf.Middlewares[key] = value
		}

		for key, value := range config.Services {
			conf.Services[key] = value
		}

		for key, value := range config.Routers {
			conf.Routers[key] = value
		}

		conf.TLS = append(conf.TLS, config.TLS...)
	}

	handlers := s.applyConfiguration(ctx, conf)

	// Get new certificates list sorted per entry points
	// Update certificates
	entryPointsCertificates := s.loadHTTPSConfiguration(configurations, globalConfiguration.DefaultEntryPoints)

	return handlers, entryPointsCertificates
}

func (s *Server) applyConfiguration(ctx context.Context, configuration config.Configuration) map[string]http.Handler {
	staticConfiguration := static.ConvertStaticConf(s.globalConfiguration)

	var entryPoints []string
	for entryPointName := range s.entryPoints {
		entryPoints = append(entryPoints, entryPointName)
	}

	serviceManager := service.NewManager(configuration.Services, s.defaultRoundTripper)
	middlewaresBuilder := middleware.NewBuilder(configuration.Middlewares, serviceManager)
	responseModifierFactory := responsemodifiers.NewBuilder(configuration.Middlewares)

	routerManager := router.NewManager(configuration.Routers, serviceManager, middlewaresBuilder, responseModifierFactory)

	handlers := routerManager.BuildHandlers(ctx, entryPoints, staticConfiguration.EntryPoints.Defaults)

	routerHandlers := make(map[string]http.Handler)

	for _, entryPointName := range entryPoints {
		internalMuxRouter := mux.NewRouter().
			SkipClean(true)

		ctx = log.With(ctx, log.Str(log.EntryPointName, entryPointName))

		factory := s.entryPoints[entryPointName].RouteAppenderFactory
		if factory != nil {
			// FIXME remove currentConfigurations
			appender := factory.NewAppender(ctx, middlewaresBuilder, &s.currentConfigurations)
			appender.Append(internalMuxRouter)
		}

		if h, ok := handlers[entryPointName]; ok {
			internalMuxRouter.NotFoundHandler = h
		} else {
			internalMuxRouter.NotFoundHandler = s.buildDefaultHTTPRouter()
		}

		routerHandlers[entryPointName] = internalMuxRouter

		chain := alice.New()

		if s.accessLoggerMiddleware != nil {
			chain = chain.Append(accesslog.WrapHandler(s.accessLoggerMiddleware))
		}

		if s.tracer != nil {
			chain = chain.Append(tracing.WrapEntryPointHandler(ctx, s.tracer, entryPointName))
		}

		chain = chain.Append(requestdecorator.WrapHandler(s.requestDecorator))

		handler, err := chain.Then(internalMuxRouter.NotFoundHandler)
		if err != nil {
			log.FromContext(ctx).Error(err)
			continue
		}
		internalMuxRouter.NotFoundHandler = handler
	}

	return routerHandlers
}

func (s *Server) preLoadConfiguration(configMsg config.Message) {
	providersThrottleDuration := time.Duration(s.globalConfiguration.ProvidersThrottleDuration)
	s.defaultConfigurationValues(configMsg.Configuration)
	currentConfigurations := s.currentConfigurations.Get().(config.Configurations)

	logger := log.WithoutContext().WithField(log.ProviderName, configMsg.ProviderName)
	if log.GetLevel() == logrus.DebugLevel {
		jsonConf, _ := json.Marshal(configMsg.Configuration)
		logger.Debugf("Configuration received from provider %s: %s", configMsg.ProviderName, string(jsonConf))
	}

	if configMsg.Configuration == nil || configMsg.Configuration.Routers == nil && configMsg.Configuration.Services == nil && configMsg.Configuration.Middlewares == nil && configMsg.Configuration.TLS == nil {
		logger.Infof("Skipping empty Configuration for provider %s", configMsg.ProviderName)
		return
	}

	if reflect.DeepEqual(currentConfigurations[configMsg.ProviderName], configMsg.Configuration) {
		logger.Infof("Skipping same configuration for provider %s", configMsg.ProviderName)
		return
	}

	providerConfigUpdateCh, ok := s.providerConfigUpdateMap[configMsg.ProviderName]
	if !ok {
		providerConfigUpdateCh = make(chan config.Message)
		s.providerConfigUpdateMap[configMsg.ProviderName] = providerConfigUpdateCh
		s.routinesPool.Go(func(stop chan bool) {
			s.throttleProviderConfigReload(providersThrottleDuration, s.configurationValidatedChan, providerConfigUpdateCh, stop)
		})
	}

	providerConfigUpdateCh <- configMsg
}

func (s *Server) defaultConfigurationValues(configuration *config.Configuration) {
	// FIXME create a config hook
}

func (s *Server) listenConfigurations(stop chan bool) {
	for {
		select {
		case <-stop:
			return
		case configMsg, ok := <-s.configurationValidatedChan:
			if !ok || configMsg.Configuration == nil {
				return
			}
			s.loadConfiguration(configMsg)
		}
	}
}

// throttleProviderConfigReload throttles the configuration reload speed for a single provider.
// It will immediately publish a new configuration and then only publish the next configuration after the throttle duration.
// Note that in the case it receives N new configs in the timeframe of the throttle duration after publishing,
// it will publish the last of the newly received configurations.
func (s *Server) throttleProviderConfigReload(throttle time.Duration, publish chan<- config.Message, in <-chan config.Message, stop chan bool) {
	ring := channels.NewRingChannel(1)
	defer ring.Close()

	s.routinesPool.Go(func(stop chan bool) {
		for {
			select {
			case <-stop:
				return
			case nextConfig := <-ring.Out():
				if config, ok := nextConfig.(config.Message); ok {
					publish <- config
					time.Sleep(throttle)
				}
			}
		}
	})

	for {
		select {
		case <-stop:
			return
		case nextConfig := <-in:
			ring.In() <- nextConfig
		}
	}
}

func (s *Server) postLoadConfiguration() {
	// FIXME metrics
	// if s.metricsRegistry.IsEnabled() {
	// 	activeConfig := s.currentConfigurations.Get().(config.Configurations)
	// 	metrics.OnConfigurationUpdate(activeConfig)
	// }

	if s.globalConfiguration.ACME == nil || s.leadership == nil || !s.leadership.IsLeader() {
		return
	}

	// FIXME acme
	// if s.globalConfiguration.ACME.OnHostRule {
	// 	currentConfigurations := s.currentConfigurations.Get().(config.Configurations)
	// 	for _, config := range currentConfigurations {
	// 		for _, frontend := range config.Frontends {
	//
	// 			// check if one of the frontend entrypoints is configured with TLS
	// 			// and is configured with ACME
	// 			acmeEnabled := false
	// 			for _, entryPoint := range frontend.EntryPoints {
	// 				if s.globalConfiguration.ACME.EntryPoint == entryPoint && s.entryPoints[entryPoint].Configuration.TLS != nil {
	// 					acmeEnabled = true
	// 					break
	// 				}
	// 			}
	//
	// 			if acmeEnabled {
	// 				for _, route := range frontend.Routes {
	// 					rls := rules.Rules{}
	// 					domains, err := rls.ParseDomains(route.Rule)
	// 					if err != nil {
	// 						log.Errorf("Error parsing domains: %v", err)
	// 					} else if len(domains) == 0 {
	// 						log.Debugf("No domain parsed in rule %q", route.Rule)
	// 					} else {
	// 						s.globalConfiguration.ACME.LoadCertificateForDomains(domains)
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }
}

// loadHTTPSConfiguration add/delete HTTPS certificate managed dynamically
func (s *Server) loadHTTPSConfiguration(configurations config.Configurations, defaultEntryPoints configuration.DefaultEntryPoints) map[string]map[string]*tls.Certificate {
	newEPCertificates := make(map[string]map[string]*tls.Certificate)
	// Get all certificates
	for _, config := range configurations {
		if config.TLS != nil && len(config.TLS) > 0 {
			traefiktls.SortTLSPerEntryPoints(config.TLS, newEPCertificates, defaultEntryPoints)
		}
	}
	return newEPCertificates
}

func (s *Server) buildServerEntryPoints() map[string]*serverEntryPoint {
	serverEntryPoints := make(map[string]*serverEntryPoint)

	ctx := context.Background()

	handlers := s.applyConfiguration(ctx, config.Configuration{})

	for entryPointName, entryPoint := range s.entryPoints {
		serverEntryPoints[entryPointName] = &serverEntryPoint{
			httpRouter:       middlewares.NewHandlerSwitcher(handlers[entryPointName]),
			onDemandListener: entryPoint.OnDemandListener,
			tlsALPNGetter:    entryPoint.TLSALPNGetter,
		}

		if entryPoint.CertificateStore != nil {
			serverEntryPoints[entryPointName].certs = entryPoint.CertificateStore
		} else {
			serverEntryPoints[entryPointName].certs = traefiktls.NewCertificateStore()
		}

		if entryPoint.Configuration.TLS != nil {
			logger := log.FromContext(ctx).WithField(log.EntryPointName, entryPointName)

			serverEntryPoints[entryPointName].certs.SniStrict = entryPoint.Configuration.TLS.SniStrict

			if entryPoint.Configuration.TLS.DefaultCertificate != nil {
				cert, err := buildDefaultCertificate(entryPoint.Configuration.TLS.DefaultCertificate)
				if err != nil {
					logger.Error(err)
					continue
				}
				serverEntryPoints[entryPointName].certs.DefaultCertificate = cert
			} else {
				cert, err := generate.DefaultCertificate()
				if err != nil {
					logger.Error(err)
					continue
				}
				serverEntryPoints[entryPointName].certs.DefaultCertificate = cert
			}
			if len(entryPoint.Configuration.TLS.Certificates) > 0 {
				config, _ := entryPoint.Configuration.TLS.Certificates.CreateTLSConfig(entryPointName)
				certMap := s.buildNameOrIPToCertificate(config.Certificates)
				serverEntryPoints[entryPointName].certs.StaticCerts.Set(certMap)

			}
		}
	}
	return serverEntryPoints
}

func (s *Server) buildDefaultHTTPRouter() *mux.Router {
	rt := mux.NewRouter()
	rt.NotFoundHandler = http.HandlerFunc(http.NotFound)
	rt.SkipClean(true)
	return rt
}

func buildDefaultCertificate(defaultCertificate *traefiktls.Certificate) (*tls.Certificate, error) {
	certFile, err := defaultCertificate.CertFile.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to get cert file content: %v", err)
	}

	keyFile, err := defaultCertificate.KeyFile.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to get key file content: %v", err)
	}

	cert, err := tls.X509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load X509 key pair: %v", err)
	}
	return &cert, nil
}
