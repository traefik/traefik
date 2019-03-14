package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/containous/alice"
	"github.com/containous/mux"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/middlewares/accesslog"
	"github.com/containous/traefik/pkg/middlewares/requestdecorator"
	"github.com/containous/traefik/pkg/middlewares/tracing"
	"github.com/containous/traefik/pkg/responsemodifiers"
	"github.com/containous/traefik/pkg/server/middleware"
	"github.com/containous/traefik/pkg/server/router"
	routertcp "github.com/containous/traefik/pkg/server/router/tcp"
	"github.com/containous/traefik/pkg/server/service"
	"github.com/containous/traefik/pkg/server/service/tcp"
	tcpCore "github.com/containous/traefik/pkg/tcp"
	"github.com/eapache/channels"
	"github.com/sirupsen/logrus"
)

// loadConfiguration manages dynamically routers, middlewares, servers and TLS configurations
func (s *Server) loadConfiguration(configMsg config.Message) {
	currentConfigurations := s.currentConfigurations.Get().(config.Configurations)

	// Copy configurations to new map so we don't change current if LoadConfig fails
	newConfigurations := make(config.Configurations)
	for k, v := range currentConfigurations {
		newConfigurations[k] = v
	}
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

	s.postLoadConfiguration()
}

// loadConfigurationTCP returns a new gorilla.mux Route from the specified global configuration and the dynamic
// provider configurations.
func (s *Server) loadConfigurationTCP(configurations config.Configurations) map[string]*tcpCore.Router {
	ctx := context.TODO()

	var entryPoints []string
	for entryPointName := range s.entryPointsTCP {
		entryPoints = append(entryPoints, entryPointName)
	}

	conf := mergeConfiguration(configurations)

	s.tlsManager.UpdateConfigs(conf.TLSStores, conf.TLSOptions, conf.TLS)

	handlersNonTLS, handlersTLS := s.createHTTPHandlers(ctx, *conf.HTTP, entryPoints)

	routersTCP := s.createTCPRouters(ctx, conf.TCP, entryPoints, handlersNonTLS, handlersTLS, s.tlsManager.Get("default", "default"))

	return routersTCP
}

func (s *Server) createTCPRouters(ctx context.Context, configuration *config.TCPConfiguration, entryPoints []string, handlers map[string]http.Handler, handlersTLS map[string]http.Handler, tlsConfig *tls.Config) map[string]*tcpCore.Router {
	if configuration == nil {
		return make(map[string]*tcpCore.Router)
	}

	serviceManager := tcp.NewManager(configuration.Services)
	routerManager := routertcp.NewManager(configuration.Routers, serviceManager, handlers, handlersTLS, tlsConfig)

	return routerManager.BuildHandlers(ctx, entryPoints)

}

func (s *Server) createHTTPHandlers(ctx context.Context, configuration config.HTTPConfiguration, entryPoints []string) (map[string]http.Handler, map[string]http.Handler) {
	serviceManager := service.NewManager(configuration.Services, s.defaultRoundTripper)
	middlewaresBuilder := middleware.NewBuilder(configuration.Middlewares, serviceManager)
	responseModifierFactory := responsemodifiers.NewBuilder(configuration.Middlewares)

	routerManager := router.NewManager(configuration.Routers, serviceManager, middlewaresBuilder, responseModifierFactory)

	handlersNonTLS := routerManager.BuildHandlers(ctx, entryPoints, false)
	handlersTLS := routerManager.BuildHandlers(ctx, entryPoints, true)

	routerHandlers := make(map[string]http.Handler)

	for _, entryPointName := range entryPoints {
		internalMuxRouter := mux.NewRouter().
			SkipClean(true)

		ctx = log.With(ctx, log.Str(log.EntryPointName, entryPointName))

		factory := s.entryPointsTCP[entryPointName].RouteAppenderFactory
		if factory != nil {
			// FIXME remove currentConfigurations
			appender := factory.NewAppender(ctx, middlewaresBuilder, &s.currentConfigurations)
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

func isEmptyConfiguration(conf *config.Configuration) bool {
	if conf == nil {
		return true
	}
	if conf.TCP == nil {
		conf.TCP = &config.TCPConfiguration{}
	}
	if conf.HTTP == nil {
		conf.HTTP = &config.HTTPConfiguration{}
	}

	return conf.HTTP.Routers == nil &&
		conf.HTTP.Services == nil &&
		conf.HTTP.Middlewares == nil &&
		conf.TLS == nil &&
		conf.TCP.Routers == nil &&
		conf.TCP.Services == nil
}

func (s *Server) preLoadConfiguration(configMsg config.Message) {
	s.defaultConfigurationValues(configMsg.Configuration.HTTP)
	currentConfigurations := s.currentConfigurations.Get().(config.Configurations)

	logger := log.WithoutContext().WithField(log.ProviderName, configMsg.ProviderName)
	if log.GetLevel() == logrus.DebugLevel {
		jsonConf, _ := json.Marshal(configMsg.Configuration)
		logger.Debugf("Configuration received from provider %s: %s", configMsg.ProviderName, string(jsonConf))
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
		providerConfigUpdateCh = make(chan config.Message)
		s.providerConfigUpdateMap[configMsg.ProviderName] = providerConfigUpdateCh
		s.routinesPool.Go(func(stop chan bool) {
			s.throttleProviderConfigReload(s.providersThrottleDuration, s.configurationValidatedChan, providerConfigUpdateCh, stop)
		})
	}

	providerConfigUpdateCh <- configMsg
}

func (s *Server) defaultConfigurationValues(configuration *config.HTTPConfiguration) {
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

}

func buildDefaultHTTPRouter() *mux.Router {
	rt := mux.NewRouter()
	rt.NotFoundHandler = http.HandlerFunc(http.NotFound)
	rt.SkipClean(true)
	return rt
}
