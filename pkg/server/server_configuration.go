package server

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/containous/alice"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/metrics"
	"github.com/containous/traefik/v2/pkg/middlewares/accesslog"
	metricsmiddleware "github.com/containous/traefik/v2/pkg/middlewares/metrics"
	"github.com/containous/traefik/v2/pkg/middlewares/requestdecorator"
	"github.com/containous/traefik/v2/pkg/middlewares/tracing"
	"github.com/containous/traefik/v2/pkg/responsemodifiers"
	"github.com/containous/traefik/v2/pkg/server/middleware"
	"github.com/containous/traefik/v2/pkg/server/router"
	routertcp "github.com/containous/traefik/v2/pkg/server/router/tcp"
	"github.com/containous/traefik/v2/pkg/server/service"
	"github.com/containous/traefik/v2/pkg/server/service/tcp"
	tcpCore "github.com/containous/traefik/v2/pkg/tcp"
	"github.com/eapache/channels"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// loadConfiguration manages dynamically routers, middlewares, servers and TLS configurations
func (s *Server) loadConfiguration(configMsg dynamic.Message) {
	currentConfigurations := s.currentConfigurations.Get().(dynamic.Configurations)

	// Copy configurations to new map so we don't change current if LoadConfig fails
	newConfigurations := currentConfigurations.DeepCopy()
	newConfigurations[configMsg.ProviderName] = configMsg.Configuration

	s.metricsRegistry.ConfigReloadsCounter().Add(1)

	handlersTCP := s.loadConfigurationTCP(newConfigurations)
	for entryPointName, router := range handlersTCP {
		s.entryPointsTCP[entryPointName].switchRouter(router)
	}

	s.metricsRegistry.LastConfigReloadSuccessGauge().Set(float64(time.Now().Unix()))

	s.currentConfigurations.Set(newConfigurations)

	for _, listener := range s.configurationListeners {
		listener(*configMsg.Configuration)
	}

	if s.metricsRegistry.IsEpEnabled() || s.metricsRegistry.IsSvcEnabled() {
		var entrypoints []string
		for key := range s.entryPointsTCP {
			entrypoints = append(entrypoints, key)
		}
		metrics.OnConfigurationUpdate(newConfigurations, entrypoints)
	}
}

// loadConfigurationTCP returns a new gorilla.mux Route from the specified global configuration and the dynamic
// provider configurations.
func (s *Server) loadConfigurationTCP(configurations dynamic.Configurations) map[string]*tcpCore.Router {
	ctx := context.TODO()

	var entryPoints []string
	for entryPointName := range s.entryPointsTCP {
		entryPoints = append(entryPoints, entryPointName)
	}

	conf := mergeConfiguration(configurations)

	s.tlsManager.UpdateConfigs(conf.TLS.Stores, conf.TLS.Options, conf.TLS.Certificates)

	rtConf := runtime.NewConfig(conf)
	handlersNonTLS, handlersTLS := s.createHTTPHandlers(ctx, rtConf, entryPoints)
	routersTCP := s.createTCPRouters(ctx, rtConf, entryPoints, handlersNonTLS, handlersTLS)
	rtConf.PopulateUsedBy()

	return routersTCP
}

// the given configuration must not be nil. its fields will get mutated.
func (s *Server) createTCPRouters(ctx context.Context, configuration *runtime.Configuration, entryPoints []string, handlers map[string]http.Handler, handlersTLS map[string]http.Handler) map[string]*tcpCore.Router {
	if configuration == nil {
		return make(map[string]*tcpCore.Router)
	}

	serviceManager := tcp.NewManager(configuration)

	routerManager := routertcp.NewManager(configuration, serviceManager, handlers, handlersTLS, s.tlsManager)

	return routerManager.BuildHandlers(ctx, entryPoints)
}

// createHTTPHandlers returns, for the given configuration and entryPoints, the HTTP handlers for non-TLS connections, and for the TLS ones. the given configuration must not be nil. its fields will get mutated.
func (s *Server) createHTTPHandlers(ctx context.Context, configuration *runtime.Configuration, entryPoints []string) (map[string]http.Handler, map[string]http.Handler) {
	serviceManager := service.NewManager(configuration.Services, s.defaultRoundTripper, s.metricsRegistry, s.routinesPool)
	middlewaresBuilder := middleware.NewBuilder(configuration.Middlewares, serviceManager)
	responseModifierFactory := responsemodifiers.NewBuilder(configuration.Middlewares)
	routerManager := router.NewManager(configuration, serviceManager, middlewaresBuilder, responseModifierFactory)

	handlersNonTLS := routerManager.BuildHandlers(ctx, entryPoints, false)
	handlersTLS := routerManager.BuildHandlers(ctx, entryPoints, true)

	routerHandlers := make(map[string]http.Handler)
	for _, entryPointName := range entryPoints {
		internalMuxRouter := mux.NewRouter().SkipClean(true)

		ctx = log.With(ctx, log.Str(log.EntryPointName, entryPointName))

		factory := s.entryPointsTCP[entryPointName].RouteAppenderFactory
		if factory != nil {
			// FIXME remove currentConfigurations
			appender := factory.NewAppender(ctx, middlewaresBuilder, configuration)
			appender.Append(internalMuxRouter)
		}

		if h, ok := handlersNonTLS[entryPointName]; ok {
			internalMuxRouter.NotFoundHandler = h
		} else {
			internalMuxRouter.NotFoundHandler = buildDefaultHTTPRouter()
		}

		routerHandlers[entryPointName] = internalMuxRouter

		chain := alice.New()

		if s.accessLoggerMiddleware != nil {
			chain = chain.Append(accesslog.WrapHandler(s.accessLoggerMiddleware))
		}

		if s.tracer != nil {
			chain = chain.Append(tracing.WrapEntryPointHandler(ctx, s.tracer, entryPointName))
		}

		if s.metricsRegistry.IsEpEnabled() {
			chain = chain.Append(metricsmiddleware.WrapEntryPointHandler(ctx, s.metricsRegistry, entryPointName))
		}

		chain = chain.Append(requestdecorator.WrapHandler(s.requestDecorator))

		handler, err := chain.Then(internalMuxRouter.NotFoundHandler)
		if err != nil {
			log.FromContext(ctx).Error(err)
			continue
		}
		internalMuxRouter.NotFoundHandler = handler

		handlerTLS, ok := handlersTLS[entryPointName]
		if ok {
			handlerTLSWithMiddlewares, err := chain.Then(handlerTLS)
			if err != nil {
				log.FromContext(ctx).Error(err)
				continue
			}
			handlersTLS[entryPointName] = handlerTLSWithMiddlewares
		}
	}

	return routerHandlers, handlersTLS
}

func isEmptyConfiguration(conf *dynamic.Configuration) bool {
	if conf == nil {
		return true
	}
	if conf.TCP == nil {
		conf.TCP = &dynamic.TCPConfiguration{}
	}
	if conf.HTTP == nil {
		conf.HTTP = &dynamic.HTTPConfiguration{}
	}

	return conf.HTTP.Routers == nil &&
		conf.HTTP.Services == nil &&
		conf.HTTP.Middlewares == nil &&
		(conf.TLS == nil || conf.TLS.Certificates == nil && conf.TLS.Stores == nil && conf.TLS.Options == nil) &&
		conf.TCP.Routers == nil &&
		conf.TCP.Services == nil
}

func (s *Server) preLoadConfiguration(configMsg dynamic.Message) {
	s.defaultConfigurationValues(configMsg.Configuration.HTTP)
	currentConfigurations := s.currentConfigurations.Get().(dynamic.Configurations)

	logger := log.WithoutContext().WithField(log.ProviderName, configMsg.ProviderName)
	if log.GetLevel() == logrus.DebugLevel {
		copyConf := configMsg.Configuration.DeepCopy()
		if copyConf.TLS != nil {
			copyConf.TLS.Certificates = nil

			for _, v := range copyConf.TLS.Stores {
				v.DefaultCertificate = nil
			}
		}

		jsonConf, err := json.Marshal(copyConf)
		if err != nil {
			logger.Errorf("Could not marshal dynamic configuration: %v", err)
			logger.Debugf("Configuration received from provider %s: [struct] %#v", configMsg.ProviderName, copyConf)
		} else {
			logger.Debugf("Configuration received from provider %s: %s", configMsg.ProviderName, string(jsonConf))
		}
	}

	if isEmptyConfiguration(configMsg.Configuration) {
		logger.Infof("Skipping empty Configuration for provider %s", configMsg.ProviderName)
		return
	}

	if reflect.DeepEqual(currentConfigurations[configMsg.ProviderName], configMsg.Configuration) {
		logger.Infof("Skipping same configuration for provider %s", configMsg.ProviderName)
		return
	}

	providerConfigUpdateCh, ok := s.providerConfigUpdateMap[configMsg.ProviderName]
	if !ok {
		providerConfigUpdateCh = make(chan dynamic.Message)
		s.providerConfigUpdateMap[configMsg.ProviderName] = providerConfigUpdateCh
		s.routinesPool.Go(func(stop chan bool) {
			s.throttleProviderConfigReload(s.providersThrottleDuration, s.configurationValidatedChan, providerConfigUpdateCh, stop)
		})
	}

	providerConfigUpdateCh <- configMsg
}

func (s *Server) defaultConfigurationValues(configuration *dynamic.HTTPConfiguration) {
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
func (s *Server) throttleProviderConfigReload(throttle time.Duration, publish chan<- dynamic.Message, in <-chan dynamic.Message, stop chan bool) {
	ring := channels.NewRingChannel(1)
	defer ring.Close()

	s.routinesPool.Go(func(stop chan bool) {
		for {
			select {
			case <-stop:
				return
			case nextConfig := <-ring.Out():
				if config, ok := nextConfig.(dynamic.Message); ok {
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

func buildDefaultHTTPRouter() *mux.Router {
	rt := mux.NewRouter()
	rt.NotFoundHandler = http.HandlerFunc(http.NotFound)
	rt.SkipClean(true)
	return rt
}
