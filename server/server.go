package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	stdlog "log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/armon/go-proxyproto"
	"github.com/containous/mux"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/healthcheck"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/metrics"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/middlewares/accesslog"
	mauth "github.com/containous/traefik/middlewares/auth"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/server/cookie"
	traefikTls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/containous/traefik/whitelist"
	"github.com/eapache/channels"
	thoas_stats "github.com/thoas/stats"
	"github.com/urfave/negroni"
	"github.com/vulcand/oxy/connlimit"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/ratelimit"
	"github.com/vulcand/oxy/roundrobin"
	"github.com/vulcand/oxy/utils"
	"golang.org/x/net/http2"
)

const (
	defaultRedirectRegex = `^(?:https?:\/\/)?([\w\._-]+)(?::\d+)?(.*)$`
)

var (
	httpServerLogger = stdlog.New(log.WriterLevel(logrus.DebugLevel), "", 0)
)

// Server is the reverse-proxy/load-balancer engine
type Server struct {
	serverEntryPoints             serverEntryPoints
	configurationChan             chan types.ConfigMessage
	configurationValidatedChan    chan types.ConfigMessage
	signals                       chan os.Signal
	stopChan                      chan bool
	providers                     []provider.Provider
	currentConfigurations         safe.Safe
	globalConfiguration           configuration.GlobalConfiguration
	accessLoggerMiddleware        *accesslog.LogHandler
	routinesPool                  *safe.Pool
	leadership                    *cluster.Leadership
	defaultForwardingRoundTripper http.RoundTripper
	metricsRegistry               metrics.Registry
}

type serverEntryPoints map[string]*serverEntryPoint

type serverEntryPoint struct {
	httpServer *http.Server
	listener   net.Listener
	httpRouter *middlewares.HandlerSwitcher
	certs      safe.Safe
}

type serverRoute struct {
	route              *mux.Route
	stripPrefixes      []string
	stripPrefixesRegex []string
	addPrefix          string
	replacePath        string
	replacePathRegex   string
}

// NewServer returns an initialized Server.
func NewServer(globalConfiguration configuration.GlobalConfiguration) *Server {
	server := new(Server)

	server.serverEntryPoints = make(map[string]*serverEntryPoint)
	server.configurationChan = make(chan types.ConfigMessage, 100)
	server.configurationValidatedChan = make(chan types.ConfigMessage, 100)
	server.signals = make(chan os.Signal, 1)
	server.stopChan = make(chan bool, 1)
	server.providers = []provider.Provider{}
	server.configureSignals()
	currentConfigurations := make(types.Configurations)
	server.currentConfigurations.Set(currentConfigurations)
	server.globalConfiguration = globalConfiguration
	if server.globalConfiguration.API != nil {
		server.globalConfiguration.API.CurrentConfigurations = &server.currentConfigurations
	}

	server.routinesPool = safe.NewPool(context.Background())
	server.defaultForwardingRoundTripper = createHTTPTransport(globalConfiguration)

	server.metricsRegistry = metrics.NewVoidRegistry()
	if globalConfiguration.Metrics != nil {
		server.registerMetricClients(globalConfiguration.Metrics)
	}

	if globalConfiguration.Cluster != nil {
		// leadership creation if cluster mode
		server.leadership = cluster.NewLeadership(server.routinesPool.Ctx(), globalConfiguration.Cluster)
	}

	if globalConfiguration.AccessLogsFile != "" {
		globalConfiguration.AccessLog = &types.AccessLog{FilePath: globalConfiguration.AccessLogsFile, Format: accesslog.CommonFormat}
	}

	if globalConfiguration.AccessLog != nil {
		var err error
		server.accessLoggerMiddleware, err = accesslog.NewLogHandler(globalConfiguration.AccessLog)
		if err != nil {
			log.Warnf("Unable to create log handler: %s", err)
		}
	}
	return server
}

// createHTTPTransport creates an http.Transport configured with the GlobalConfiguration settings.
// For the settings that can't be configured in Traefik it uses the default http.Transport settings.
// An exception to this is the MaxIdleConns setting as we only provide the option MaxIdleConnsPerHost
// in Traefik at this point in time. Setting this value to the default of 100 could lead to confusing
// behaviour and backwards compatibility issues.
func createHTTPTransport(globalConfiguration configuration.GlobalConfiguration) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   configuration.DefaultDialTimeout,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	if globalConfiguration.ForwardingTimeouts != nil {
		dialer.Timeout = time.Duration(globalConfiguration.ForwardingTimeouts.DialTimeout)
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConnsPerHost:   globalConfiguration.MaxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if globalConfiguration.ForwardingTimeouts != nil {
		transport.ResponseHeaderTimeout = time.Duration(globalConfiguration.ForwardingTimeouts.ResponseHeaderTimeout)
	}
	if globalConfiguration.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if len(globalConfiguration.RootCAs) > 0 {
		transport.TLSClientConfig = &tls.Config{
			RootCAs: createRootCACertPool(globalConfiguration.RootCAs),
		}
	}
	http2.ConfigureTransport(transport)

	return transport
}

func createRootCACertPool(rootCAs traefikTls.RootCAs) *x509.CertPool {
	roots := x509.NewCertPool()

	for _, cert := range rootCAs {
		certContent, err := cert.Read()
		if err != nil {
			log.Error("Error while read RootCAs", err)
			continue
		}
		roots.AppendCertsFromPEM(certContent)
	}

	return roots
}

// Start starts the server.
func (s *Server) Start() {
	s.startHTTPServers()
	s.startLeadership()
	s.routinesPool.Go(func(stop chan bool) {
		s.listenProviders(stop)
	})
	s.routinesPool.Go(func(stop chan bool) {
		s.listenConfigurations(stop)
	})
	s.configureProviders()
	s.startProviders()
	go s.listenSignals()
}

// Wait blocks until server is shutted down.
func (s *Server) Wait() {
	<-s.stopChan
}

// Stop stops the server
func (s *Server) Stop() {
	defer log.Info("Server stopped")
	var wg sync.WaitGroup
	for sepn, sep := range s.serverEntryPoints {
		wg.Add(1)
		go func(serverEntryPointName string, serverEntryPoint *serverEntryPoint) {
			defer wg.Done()
			graceTimeOut := time.Duration(s.globalConfiguration.LifeCycle.GraceTimeOut)
			ctx, cancel := context.WithTimeout(context.Background(), graceTimeOut)
			log.Debugf("Waiting %s seconds before killing connections on entrypoint %s...", graceTimeOut, serverEntryPointName)
			if err := serverEntryPoint.httpServer.Shutdown(ctx); err != nil {
				log.Debugf("Wait is over due to: %s", err)
				serverEntryPoint.httpServer.Close()
			}
			cancel()
			log.Debugf("Entrypoint %s closed", serverEntryPointName)
		}(sepn, sep)
	}
	wg.Wait()
	s.stopChan <- true
}

// Close destroys the server
func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.globalConfiguration.LifeCycle.GraceTimeOut))
	go func(ctx context.Context) {
		<-ctx.Done()
		if ctx.Err() == context.Canceled {
			return
		} else if ctx.Err() == context.DeadlineExceeded {
			log.Warn("Timeout while stopping traefik, killing instance ✝")
			os.Exit(1)
		}
	}(ctx)
	stopMetricsClients()
	s.stopLeadership()
	s.routinesPool.Cleanup()
	close(s.configurationChan)
	close(s.configurationValidatedChan)
	signal.Stop(s.signals)
	close(s.signals)
	close(s.stopChan)
	if s.accessLoggerMiddleware != nil {
		if err := s.accessLoggerMiddleware.Close(); err != nil {
			log.Errorf("Error closing access log file: %s", err)
		}
	}
	cancel()
}

func (s *Server) startLeadership() {
	if s.leadership != nil {
		s.leadership.Participate(s.routinesPool)
	}
}

func (s *Server) stopLeadership() {
	if s.leadership != nil {
		s.leadership.Stop()
	}
}

func (s *Server) startHTTPServers() {
	s.serverEntryPoints = s.buildEntryPoints(s.globalConfiguration)

	for newServerEntryPointName, newServerEntryPoint := range s.serverEntryPoints {
		serverEntryPoint := s.setupServerEntryPoint(newServerEntryPointName, newServerEntryPoint)
		go s.startServer(serverEntryPoint, s.globalConfiguration)
	}
}

func (s *Server) setupServerEntryPoint(newServerEntryPointName string, newServerEntryPoint *serverEntryPoint) *serverEntryPoint {
	serverMiddlewares := []negroni.Handler{middlewares.NegroniRecoverHandler()}
	serverInternalMiddlewares := []negroni.Handler{middlewares.NegroniRecoverHandler()}
	if s.accessLoggerMiddleware != nil {
		serverMiddlewares = append(serverMiddlewares, s.accessLoggerMiddleware)
	}
	if s.metricsRegistry.IsEnabled() {
		serverMiddlewares = append(serverMiddlewares, middlewares.NewMetricsWrapper(s.metricsRegistry, newServerEntryPointName))
	}
	if s.globalConfiguration.API != nil {
		if s.globalConfiguration.API.Stats == nil {
			s.globalConfiguration.API.Stats = thoas_stats.New()
		}
		serverMiddlewares = append(serverMiddlewares, s.globalConfiguration.API.Stats)
		if s.globalConfiguration.API.Statistics != nil {
			if s.globalConfiguration.API.StatsRecorder == nil {
				s.globalConfiguration.API.StatsRecorder = middlewares.NewStatsRecorder(s.globalConfiguration.API.Statistics.RecentErrors)
			}
			serverMiddlewares = append(serverMiddlewares, s.globalConfiguration.API.StatsRecorder)
		}

	}
	if s.globalConfiguration.EntryPoints[newServerEntryPointName].Auth != nil {
		authMiddleware, err := mauth.NewAuthenticator(s.globalConfiguration.EntryPoints[newServerEntryPointName].Auth)
		if err != nil {
			log.Fatal("Error starting server: ", err)
		}
		serverMiddlewares = append(serverMiddlewares, authMiddleware)
		serverInternalMiddlewares = append(serverInternalMiddlewares, authMiddleware)
	}
	if s.globalConfiguration.EntryPoints[newServerEntryPointName].Compress {
		serverMiddlewares = append(serverMiddlewares, &middlewares.Compress{})
	}
	if len(s.globalConfiguration.EntryPoints[newServerEntryPointName].WhitelistSourceRange) > 0 {
		ipWhitelistMiddleware, err := middlewares.NewIPWhitelister(s.globalConfiguration.EntryPoints[newServerEntryPointName].WhitelistSourceRange)
		if err != nil {
			log.Fatal("Error starting server: ", err)
		}
		serverMiddlewares = append(serverMiddlewares, ipWhitelistMiddleware)
		serverInternalMiddlewares = append(serverInternalMiddlewares, ipWhitelistMiddleware)
	}
	newSrv, listener, err := s.prepareServer(newServerEntryPointName, s.globalConfiguration.EntryPoints[newServerEntryPointName], newServerEntryPoint.httpRouter, serverMiddlewares, serverInternalMiddlewares)
	if err != nil {
		log.Fatal("Error preparing server: ", err)
	}
	serverEntryPoint := s.serverEntryPoints[newServerEntryPointName]
	serverEntryPoint.httpServer = newSrv
	serverEntryPoint.listener = listener

	return serverEntryPoint
}

func (s *Server) listenProviders(stop chan bool) {
	for {
		select {
		case <-stop:
			return
		case configMsg, ok := <-s.configurationChan:
			if !ok || configMsg.Configuration == nil {
				return
			}
			s.preLoadConfiguration(configMsg)
		}
	}
}

func (s *Server) preLoadConfiguration(configMsg types.ConfigMessage) {
	providerConfigUpdateMap := map[string]chan types.ConfigMessage{}
	providersThrottleDuration := time.Duration(s.globalConfiguration.ProvidersThrottleDuration)
	s.defaultConfigurationValues(configMsg.Configuration)
	currentConfigurations := s.currentConfigurations.Get().(types.Configurations)
	jsonConf, _ := json.Marshal(configMsg.Configuration)
	log.Debugf("Configuration received from provider %s: %s", configMsg.ProviderName, string(jsonConf))
	if configMsg.Configuration == nil || configMsg.Configuration.Backends == nil && configMsg.Configuration.Frontends == nil && configMsg.Configuration.TLSConfiguration == nil {
		log.Infof("Skipping empty Configuration for provider %s", configMsg.ProviderName)
	} else if reflect.DeepEqual(currentConfigurations[configMsg.ProviderName], configMsg.Configuration) {
		log.Infof("Skipping same configuration for provider %s", configMsg.ProviderName)
	} else {
		if _, ok := providerConfigUpdateMap[configMsg.ProviderName]; !ok {
			providerConfigUpdate := make(chan types.ConfigMessage)
			providerConfigUpdateMap[configMsg.ProviderName] = providerConfigUpdate
			s.routinesPool.Go(func(stop chan bool) {
				throttleProviderConfigReload(providersThrottleDuration, s.configurationValidatedChan, providerConfigUpdate, stop)
			})
		}
		providerConfigUpdateMap[configMsg.ProviderName] <- configMsg
	}
}

// throttleProviderConfigReload throttles the configuration reload speed for a single provider.
// It will immediately publish a new configuration and then only publish the next configuration after the throttle duration.
// Note that in the case it receives N new configs in the timeframe of the throttle duration after publishing,
// it will publish the last of the newly received configurations.
func throttleProviderConfigReload(throttle time.Duration, publish chan<- types.ConfigMessage, in <-chan types.ConfigMessage, stop chan bool) {
	ring := channels.NewRingChannel(1)
	defer ring.Close()

	safe.Go(func() {
		for {
			select {
			case <-stop:
				return
			case nextConfig := <-ring.Out():
				publish <- nextConfig.(types.ConfigMessage)
				time.Sleep(throttle)
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

func (s *Server) defaultConfigurationValues(configuration *types.Configuration) {
	if configuration == nil || configuration.Frontends == nil {
		return
	}
	s.configureFrontends(configuration.Frontends)
	s.configureBackends(configuration.Backends)
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

// loadConfiguration manages dynamically frontends, backends and TLS configurations
func (s *Server) loadConfiguration(configMsg types.ConfigMessage) {
	currentConfigurations := s.currentConfigurations.Get().(types.Configurations)

	// Copy configurations to new map so we don't change current if LoadConfig fails
	newConfigurations := make(types.Configurations)
	for k, v := range currentConfigurations {
		newConfigurations[k] = v
	}
	newConfigurations[configMsg.ProviderName] = configMsg.Configuration

	newServerEntryPoints, err := s.loadConfig(newConfigurations, s.globalConfiguration)
	if err == nil {
		for newServerEntryPointName, newServerEntryPoint := range newServerEntryPoints {
			s.serverEntryPoints[newServerEntryPointName].httpRouter.UpdateHandler(newServerEntryPoint.httpRouter.GetHandler())
			if s.globalConfiguration.EntryPoints[newServerEntryPointName].TLS == nil {
				if newServerEntryPoint.certs.Get() != nil {
					log.Debugf("Certificates not added to non-TLS entryPoint %s.", newServerEntryPointName)
				}
			} else {
				s.serverEntryPoints[newServerEntryPointName].certs.Set(newServerEntryPoint.certs.Get())
			}
			log.Infof("Server configuration reloaded on %s", s.serverEntryPoints[newServerEntryPointName].httpServer.Addr)
		}
		s.currentConfigurations.Set(newConfigurations)
		s.postLoadConfiguration()
	} else {
		log.Error("Error loading new configuration, aborted ", err)
	}
}

// loadHTTPSConfiguration add/delete HTTPS certificate managed dynamically
func (s *Server) loadHTTPSConfiguration(configurations types.Configurations) (map[string]*traefikTls.DomainsCertificates, error) {
	newEPCertificates := make(map[string]*traefikTls.DomainsCertificates)
	// Get all certificates
	for _, configuration := range configurations {
		if configuration.TLSConfiguration != nil && len(configuration.TLSConfiguration) > 0 {
			if err := traefikTls.SortTLSConfigurationPerEntryPoints(configuration.TLSConfiguration, newEPCertificates); err != nil {
				return nil, err
			}
		}
	}
	return newEPCertificates, nil
}

// getCertificate allows to customize tlsConfig.Getcertificate behaviour to get the certificates inserted dynamically
func (s *serverEntryPoint) getCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if s.certs.Get() != nil {
		domainToCheck := types.CanonicalDomain(clientHello.ServerName)
		for domains, cert := range *s.certs.Get().(*traefikTls.DomainsCertificates) {
			for _, domain := range strings.Split(domains, ",") {
				selector := "^" + strings.Replace(domain, "*.", "[^\\.]*\\.?", -1) + "$"
				domainCheck, _ := regexp.MatchString(selector, domainToCheck)
				if domainCheck {
					return cert, nil
				}
			}
		}
	}
	return nil, nil
}

func (s *Server) postLoadConfiguration() {
	if s.globalConfiguration.ACME == nil {
		return
	}
	if s.leadership != nil && !s.leadership.IsLeader() {
		return
	}
	if s.globalConfiguration.ACME.OnHostRule {
		currentConfigurations := s.currentConfigurations.Get().(types.Configurations)
		for _, config := range currentConfigurations {
			for _, frontend := range config.Frontends {

				// check if one of the frontend entrypoints is configured with TLS
				// and is configured with ACME
				ACMEEnabled := false
				for _, entryPoint := range frontend.EntryPoints {
					if s.globalConfiguration.ACME.EntryPoint == entryPoint && s.globalConfiguration.EntryPoints[entryPoint].TLS != nil {
						ACMEEnabled = true
						break
					}
				}

				if ACMEEnabled {
					for _, route := range frontend.Routes {
						rules := Rules{}
						domains, err := rules.ParseDomains(route.Rule)
						if err != nil {
							log.Errorf("Error parsing domains: %v", err)
						} else {
							s.globalConfiguration.ACME.LoadCertificateForDomains(domains)
						}
					}
				}
			}
		}
	}
}

func (s *Server) configureProviders() {
	// configure providers
	if s.globalConfiguration.Docker != nil {
		s.providers = append(s.providers, s.globalConfiguration.Docker)
	}
	if s.globalConfiguration.Marathon != nil {
		s.providers = append(s.providers, s.globalConfiguration.Marathon)
	}
	if s.globalConfiguration.File != nil {
		s.providers = append(s.providers, s.globalConfiguration.File)
	}
	if s.globalConfiguration.Rest != nil {
		s.providers = append(s.providers, s.globalConfiguration.Rest)
		s.globalConfiguration.Rest.CurrentConfigurations = &s.currentConfigurations
	}
	if s.globalConfiguration.Consul != nil {
		s.providers = append(s.providers, s.globalConfiguration.Consul)
	}
	if s.globalConfiguration.ConsulCatalog != nil {
		s.providers = append(s.providers, s.globalConfiguration.ConsulCatalog)
	}
	if s.globalConfiguration.Etcd != nil {
		s.providers = append(s.providers, s.globalConfiguration.Etcd)
	}
	if s.globalConfiguration.Zookeeper != nil {
		s.providers = append(s.providers, s.globalConfiguration.Zookeeper)
	}
	if s.globalConfiguration.Boltdb != nil {
		s.providers = append(s.providers, s.globalConfiguration.Boltdb)
	}
	if s.globalConfiguration.Kubernetes != nil {
		s.providers = append(s.providers, s.globalConfiguration.Kubernetes)
	}
	if s.globalConfiguration.Mesos != nil {
		s.providers = append(s.providers, s.globalConfiguration.Mesos)
	}
	if s.globalConfiguration.Eureka != nil {
		s.providers = append(s.providers, s.globalConfiguration.Eureka)
	}
	if s.globalConfiguration.ECS != nil {
		s.providers = append(s.providers, s.globalConfiguration.ECS)
	}
	if s.globalConfiguration.Rancher != nil {
		s.providers = append(s.providers, s.globalConfiguration.Rancher)
	}
	if s.globalConfiguration.DynamoDB != nil {
		s.providers = append(s.providers, s.globalConfiguration.DynamoDB)
	}
	if s.globalConfiguration.ServiceFabric != nil {
		s.providers = append(s.providers, s.globalConfiguration.ServiceFabric)
	}
}

func (s *Server) startProviders() {
	// start providers
	for _, p := range s.providers {
		providerType := reflect.TypeOf(p)
		jsonConf, _ := json.Marshal(p)
		log.Infof("Starting provider %v %s", providerType, jsonConf)
		currentProvider := p
		safe.Go(func() {
			err := currentProvider.Provide(s.configurationChan, s.routinesPool, s.globalConfiguration.Constraints)
			if err != nil {
				log.Errorf("Error starting provider %v: %s", providerType, err)
			}
		})
	}
}

func createClientTLSConfig(entryPointName string, tlsOption *traefikTls.TLS) (*tls.Config, error) {
	if tlsOption == nil {
		return nil, errors.New("no TLS provided")
	}

	config, _, err := tlsOption.Certificates.CreateTLSConfig(entryPointName)
	if err != nil {
		return nil, err
	}

	if len(tlsOption.ClientCAFiles) > 0 {
		log.Warnf("Deprecated configuration found during client TLS configuration creation: %s. Please use %s (which allows to make the CA Files optional).", "tls.ClientCAFiles", "tls.ClientCA.files")
		tlsOption.ClientCA.Files = tlsOption.ClientCAFiles
		tlsOption.ClientCA.Optional = false
	}
	if len(tlsOption.ClientCA.Files) > 0 {
		pool := x509.NewCertPool()
		for _, caFile := range tlsOption.ClientCA.Files {
			data, err := ioutil.ReadFile(caFile)
			if err != nil {
				return nil, err
			}
			if !pool.AppendCertsFromPEM(data) {
				return nil, errors.New("invalid certificate(s) in " + caFile)
			}
		}
		config.RootCAs = pool
	}
	config.BuildNameToCertificate()
	return config, nil
}

// creates a TLS config that allows terminating HTTPS for multiple domains using SNI
func (s *Server) createTLSConfig(entryPointName string, tlsOption *traefikTls.TLS, router *middlewares.HandlerSwitcher) (*tls.Config, error) {
	if tlsOption == nil {
		return nil, nil
	}

	config, epDomainsCertificates, err := tlsOption.Certificates.CreateTLSConfig(entryPointName)
	if err != nil {
		return nil, err
	}
	epDomainsCertificatesTmp := new(traefikTls.DomainsCertificates)
	if epDomainsCertificates[entryPointName] != nil {
		epDomainsCertificatesTmp = epDomainsCertificates[entryPointName]
	} else {
		*epDomainsCertificatesTmp = make(map[string]*tls.Certificate)
	}
	s.serverEntryPoints[entryPointName].certs.Set(epDomainsCertificatesTmp)
	// ensure http2 enabled
	config.NextProtos = []string{"h2", "http/1.1"}

	if len(tlsOption.ClientCAFiles) > 0 {
		log.Warnf("Deprecated configuration found during TLS configuration creation: %s. Please use %s (which allows to make the CA Files optional).", "tls.ClientCAFiles", "tls.ClientCA.files")
		tlsOption.ClientCA.Files = tlsOption.ClientCAFiles
		tlsOption.ClientCA.Optional = false
	}
	if len(tlsOption.ClientCA.Files) > 0 {
		pool := x509.NewCertPool()
		for _, caFile := range tlsOption.ClientCA.Files {
			data, err := ioutil.ReadFile(caFile)
			if err != nil {
				return nil, err
			}
			ok := pool.AppendCertsFromPEM(data)
			if !ok {
				return nil, errors.New("invalid certificate(s) in " + caFile)
			}
		}
		config.ClientCAs = pool
		if tlsOption.ClientCA.Optional {
			config.ClientAuth = tls.VerifyClientCertIfGiven
		} else {
			config.ClientAuth = tls.RequireAndVerifyClientCert
		}
	}

	if s.globalConfiguration.ACME != nil {
		if _, ok := s.serverEntryPoints[s.globalConfiguration.ACME.EntryPoint]; ok {
			if entryPointName == s.globalConfiguration.ACME.EntryPoint {
				checkOnDemandDomain := func(domain string) bool {
					routeMatch := &mux.RouteMatch{}
					router := router.GetHandler()
					match := router.Match(&http.Request{URL: &url.URL{}, Host: domain}, routeMatch)
					if match && routeMatch.Route != nil {
						return true
					}
					return false
				}
				if s.leadership == nil {
					err := s.globalConfiguration.ACME.CreateLocalConfig(config, &s.serverEntryPoints[entryPointName].certs, checkOnDemandDomain)
					if err != nil {
						return nil, err
					}
				} else {
					err := s.globalConfiguration.ACME.CreateClusterConfig(s.leadership, config, &s.serverEntryPoints[entryPointName].certs, checkOnDemandDomain)
					if err != nil {
						return nil, err
					}
				}
			}
		} else {
			return nil, errors.New("Unknown entrypoint " + s.globalConfiguration.ACME.EntryPoint + " for ACME configuration")
		}
	} else {
		config.GetCertificate = s.serverEntryPoints[entryPointName].getCertificate
	}
	if len(config.Certificates) == 0 {
		return nil, errors.New("No certificates found for TLS entrypoint " + entryPointName)
	}
	// BuildNameToCertificate parses the CommonName and SubjectAlternateName fields
	// in each certificate and populates the config.NameToCertificate map.
	config.BuildNameToCertificate()
	//Set the minimum TLS version if set in the config TOML
	if minConst, exists := traefikTls.MinVersion[s.globalConfiguration.EntryPoints[entryPointName].TLS.MinVersion]; exists {
		config.PreferServerCipherSuites = true
		config.MinVersion = minConst
	}
	//Set the list of CipherSuites if set in the config TOML
	if s.globalConfiguration.EntryPoints[entryPointName].TLS.CipherSuites != nil {
		//if our list of CipherSuites is defined in the entrypoint config, we can re-initilize the suites list as empty
		config.CipherSuites = make([]uint16, 0)
		for _, cipher := range s.globalConfiguration.EntryPoints[entryPointName].TLS.CipherSuites {
			if cipherConst, exists := traefikTls.CipherSuites[cipher]; exists {
				config.CipherSuites = append(config.CipherSuites, cipherConst)
			} else {
				//CipherSuite listed in the toml does not exist in our listed
				return nil, errors.New("Invalid CipherSuite: " + cipher)
			}
		}
	}
	return config, nil
}

func (s *Server) startServer(serverEntryPoint *serverEntryPoint, globalConfiguration configuration.GlobalConfiguration) {
	log.Infof("Starting server on %s", serverEntryPoint.httpServer.Addr)
	var err error
	if serverEntryPoint.httpServer.TLSConfig != nil {
		err = serverEntryPoint.httpServer.ServeTLS(serverEntryPoint.listener, "", "")
	} else {
		err = serverEntryPoint.httpServer.Serve(serverEntryPoint.listener)
	}
	if err != http.ErrServerClosed {
		log.Error("Error creating server: ", err)
	}
}

func (s *Server) addInternalRoutes(entryPointName string, router *mux.Router) {
	if s.globalConfiguration.Metrics != nil && s.globalConfiguration.Metrics.Prometheus != nil && s.globalConfiguration.Metrics.Prometheus.EntryPoint == entryPointName {
		metrics.PrometheusHandler{}.AddRoutes(router)
	}

	if s.globalConfiguration.Rest != nil && s.globalConfiguration.Rest.EntryPoint == entryPointName {
		s.globalConfiguration.Rest.AddRoutes(router)
	}

	if s.globalConfiguration.API != nil && s.globalConfiguration.API.EntryPoint == entryPointName {
		s.globalConfiguration.API.AddRoutes(router)
	}
}

func (s *Server) addInternalPublicRoutes(entryPointName string, router *mux.Router) {
	if s.globalConfiguration.Ping != nil && s.globalConfiguration.Ping.EntryPoint != "" && s.globalConfiguration.Ping.EntryPoint == entryPointName {
		s.globalConfiguration.Ping.AddRoutes(router)
	}
}

func (s *Server) addACMERoutes(entryPointName string, router *mux.Router) {
	if s.globalConfiguration.ACME != nil && s.globalConfiguration.ACME.HTTPChallenge != nil && s.globalConfiguration.ACME.HTTPChallenge.EntryPoint == entryPointName {
		s.globalConfiguration.ACME.AddRoutes(router)
	}
}

func (s *Server) prepareServer(entryPointName string, entryPoint *configuration.EntryPoint, router *middlewares.HandlerSwitcher, middlewares []negroni.Handler, internalMiddlewares []negroni.Handler) (*http.Server, net.Listener, error) {
	readTimeout, writeTimeout, idleTimeout := buildServerTimeouts(s.globalConfiguration)
	log.Infof("Preparing server %s %+v with readTimeout=%s writeTimeout=%s idleTimeout=%s", entryPointName, entryPoint, readTimeout, writeTimeout, idleTimeout)

	// middlewares
	n := negroni.New()
	for _, middleware := range middlewares {
		n.Use(middleware)
	}
	n.UseHandler(router)

	path := "/"
	if s.globalConfiguration.Web != nil && s.globalConfiguration.Web.Path != "" {
		path = s.globalConfiguration.Web.Path
	}

	internalMuxRouter := s.buildInternalRouter(entryPointName, path, internalMiddlewares)
	internalMuxRouter.NotFoundHandler = n

	tlsConfig, err := s.createTLSConfig(entryPointName, entryPoint.TLS, router)
	if err != nil {
		log.Errorf("Error creating TLS config: %s", err)
		return nil, nil, err
	}

	listener, err := net.Listen("tcp", entryPoint.Address)
	if err != nil {
		log.Error("Error opening listener ", err)
		return nil, nil, err
	}

	if entryPoint.ProxyProtocol != nil {
		IPs, err := whitelist.NewIP(entryPoint.ProxyProtocol.TrustedIPs, entryPoint.ProxyProtocol.Insecure)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating whitelist: %s", err)
		}
		log.Infof("Enabling ProxyProtocol for trusted IPs %v", entryPoint.ProxyProtocol.TrustedIPs)
		listener = &proxyproto.Listener{
			Listener: listener,
			SourceCheck: func(addr net.Addr) (bool, error) {
				ip, ok := addr.(*net.TCPAddr)
				if !ok {
					return false, fmt.Errorf("type error %v", addr)
				}
				return IPs.ContainsIP(ip.IP)
			},
		}
	}

	return &http.Server{
			Addr:         entryPoint.Address,
			Handler:      internalMuxRouter,
			TLSConfig:    tlsConfig,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
			ErrorLog:     httpServerLogger,
		},
		listener,
		nil
}

func (s *Server) buildInternalRouter(entryPointName, path string, internalMiddlewares []negroni.Handler) *mux.Router {
	internalMuxRouter := mux.NewRouter()
	internalMuxRouter.StrictSlash(true)
	internalMuxRouter.SkipClean(true)

	internalMuxSubrouter := internalMuxRouter.PathPrefix(path).Subrouter()
	internalMuxSubrouter.StrictSlash(true)
	internalMuxSubrouter.SkipClean(true)

	s.addInternalRoutes(entryPointName, internalMuxSubrouter)
	internalMuxRouter.Walk(wrapRoute(internalMiddlewares))

	s.addInternalPublicRoutes(entryPointName, internalMuxSubrouter)

	s.addACMERoutes(entryPointName, internalMuxRouter)

	return internalMuxRouter
}

// wrapRoute with middlewares
func wrapRoute(middlewares []negroni.Handler) func(*mux.Route, *mux.Router, []*mux.Route) error {
	return func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		middles := append(middlewares, negroni.Wrap(route.GetHandler()))
		route.Handler(negroni.New(middles...))
		return nil
	}
}

func buildServerTimeouts(globalConfig configuration.GlobalConfiguration) (readTimeout, writeTimeout, idleTimeout time.Duration) {
	readTimeout = time.Duration(0)
	writeTimeout = time.Duration(0)
	if globalConfig.RespondingTimeouts != nil {
		readTimeout = time.Duration(globalConfig.RespondingTimeouts.ReadTimeout)
		writeTimeout = time.Duration(globalConfig.RespondingTimeouts.WriteTimeout)
	}

	// Prefer legacy idle timeout parameter for backwards compatibility reasons
	if globalConfig.IdleTimeout > 0 {
		idleTimeout = time.Duration(globalConfig.IdleTimeout)
		log.Warn("top-level idle timeout configuration has been deprecated -- please use responding timeouts")
	} else if globalConfig.RespondingTimeouts != nil {
		idleTimeout = time.Duration(globalConfig.RespondingTimeouts.IdleTimeout)
	} else {
		idleTimeout = time.Duration(configuration.DefaultIdleTimeout)
	}

	return readTimeout, writeTimeout, idleTimeout
}

func (s *Server) buildEntryPoints(globalConfiguration configuration.GlobalConfiguration) map[string]*serverEntryPoint {
	serverEntryPoints := make(map[string]*serverEntryPoint)
	for entryPointName := range globalConfiguration.EntryPoints {
		router := s.buildDefaultHTTPRouter()
		serverEntryPoints[entryPointName] = &serverEntryPoint{
			httpRouter: middlewares.NewHandlerSwitcher(router),
		}
	}
	return serverEntryPoints
}

// getRoundTripper will either use server.defaultForwardingRoundTripper or create a new one
// given a custom TLS configuration is passed and the passTLSCert option is set to true.
func (s *Server) getRoundTripper(entryPointName string, globalConfiguration configuration.GlobalConfiguration, passTLSCert bool, tls *traefikTls.TLS) (http.RoundTripper, error) {
	if passTLSCert {
		tlsConfig, err := createClientTLSConfig(entryPointName, tls)
		if err != nil {
			log.Errorf("Failed to create TLSClientConfig: %s", err)
			return nil, err
		}

		transport := createHTTPTransport(globalConfiguration)
		transport.TLSClientConfig = tlsConfig
		return transport, nil
	}

	return s.defaultForwardingRoundTripper, nil
}

// LoadConfig returns a new gorilla.mux Route from the specified global configuration and the dynamic
// provider configurations.
func (s *Server) loadConfig(configurations types.Configurations, globalConfiguration configuration.GlobalConfiguration) (map[string]*serverEntryPoint, error) {
	serverEntryPoints := s.buildEntryPoints(globalConfiguration)
	redirectHandlers := make(map[string]negroni.Handler)
	backends := map[string]http.Handler{}
	backendsHealthCheck := map[string]*healthcheck.BackendHealthCheck{}
	errorHandler := NewRecordingErrorHandler(middlewares.DefaultNetErrorRecorder{})

	for _, config := range configurations {
		frontendNames := sortedFrontendNamesForConfig(config)
	frontend:
		for _, frontendName := range frontendNames {
			frontend := config.Frontends[frontendName]

			log.Debugf("Creating frontend %s", frontendName)

			if len(frontend.EntryPoints) == 0 {
				log.Errorf("No entrypoint defined for frontend %s, defaultEntryPoints:%s", frontendName, globalConfiguration.DefaultEntryPoints)
				log.Errorf("Skipping frontend %s...", frontendName)
				continue frontend
			}
			var failedEntrypoints int
			for _, entryPointName := range frontend.EntryPoints {
				log.Debugf("Wiring frontend %s to entryPoint %s", frontendName, entryPointName)
				if _, ok := serverEntryPoints[entryPointName]; !ok {
					log.Errorf("Undefined entrypoint '%s' for frontend %s", entryPointName, frontendName)
					failedEntrypoints++
					if failedEntrypoints == len(frontend.EntryPoints) {
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}
					continue
				}

				newServerRoute := &serverRoute{route: serverEntryPoints[entryPointName].httpRouter.GetHandler().NewRoute().Name(frontendName)}
				for routeName, route := range frontend.Routes {
					err := getRoute(newServerRoute, &route)
					if err != nil {
						log.Errorf("Error creating route for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}
					log.Debugf("Creating route %s %s", routeName, route.Rule)
				}

				entryPoint := globalConfiguration.EntryPoints[entryPointName]
				n := negroni.New()
				if entryPoint.Redirect != nil {
					if redirectHandlers[entryPointName] != nil {
						n.Use(redirectHandlers[entryPointName])
					} else if handler, err := s.buildRedirectHandler(entryPointName, entryPoint.Redirect); err != nil {
						log.Errorf("Error loading entrypoint configuration for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					} else {
						if s.accessLoggerMiddleware != nil {
							saveFrontend := accesslog.NewSaveNegroniFrontend(handler, frontendName)
							n.Use(saveFrontend)
							redirectHandlers[entryPointName] = saveFrontend
						} else {
							n.Use(handler)
							redirectHandlers[entryPointName] = handler
						}
					}
				}
				if backends[entryPointName+frontend.Backend] == nil {
					log.Debugf("Creating backend %s", frontend.Backend)

					roundTripper, err := s.getRoundTripper(entryPointName, globalConfiguration, frontend.PassTLSCert, entryPoint.TLS)
					if err != nil {
						log.Errorf("Failed to create RoundTripper for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					rewriter, err := NewHeaderRewriter(entryPoint.ForwardedHeaders.TrustedIPs, entryPoint.ForwardedHeaders.Insecure)
					if err != nil {
						log.Errorf("Error creating rewriter for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					headerMiddleware := middlewares.NewHeaderFromStruct(frontend.Headers)
					var responseModifier func(res *http.Response) error
					if headerMiddleware != nil {
						responseModifier = headerMiddleware.ModifyResponseHeaders
					}

					fwd, err := forward.New(
						forward.Stream(true),
						forward.PassHostHeader(frontend.PassHostHeader),
						forward.RoundTripper(roundTripper),
						forward.ErrorHandler(errorHandler),
						forward.Rewriter(rewriter),
						forward.ResponseModifier(responseModifier),
					)

					if err != nil {
						log.Errorf("Error creating forwarder for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					var rr *roundrobin.RoundRobin
					var saveFrontend http.Handler
					if s.accessLoggerMiddleware != nil {
						saveBackend := accesslog.NewSaveBackend(fwd, frontend.Backend)
						saveFrontend = accesslog.NewSaveFrontend(saveBackend, frontendName)
						rr, _ = roundrobin.New(saveFrontend)
					} else {
						rr, _ = roundrobin.New(fwd)
					}

					if config.Backends[frontend.Backend] == nil {
						log.Errorf("Undefined backend '%s' for frontend %s", frontend.Backend, frontendName)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					lbMethod, err := types.NewLoadBalancerMethod(config.Backends[frontend.Backend].LoadBalancer)
					if err != nil {
						log.Errorf("Error loading load balancer method '%+v' for frontend %s: %v", config.Backends[frontend.Backend].LoadBalancer, frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					var sticky *roundrobin.StickySession
					var cookieName string
					if stickiness := config.Backends[frontend.Backend].LoadBalancer.Stickiness; stickiness != nil {
						cookieName = cookie.GetName(stickiness.CookieName, frontend.Backend)
						sticky = roundrobin.NewStickySession(cookieName)
					}

					var lb http.Handler
					switch lbMethod {
					case types.Drr:
						log.Debugf("Creating load-balancer drr")
						rebalancer, _ := roundrobin.NewRebalancer(rr)
						if sticky != nil {
							log.Debugf("Sticky session with cookie %v", cookieName)
							rebalancer, _ = roundrobin.NewRebalancer(rr, roundrobin.RebalancerStickySession(sticky))
						}
						lb = rebalancer
						if err := configureLBServers(rebalancer, config, frontend); err != nil {
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
						hcOpts := parseHealthCheckOptions(rebalancer, frontend.Backend, config.Backends[frontend.Backend].HealthCheck, globalConfiguration.HealthCheck)
						if hcOpts != nil {
							log.Debugf("Setting up backend health check %s", *hcOpts)
							hcOpts.Transport = s.defaultForwardingRoundTripper
							backendsHealthCheck[entryPointName+frontend.Backend] = healthcheck.NewBackendHealthCheck(*hcOpts)
						}
						lb = middlewares.NewEmptyBackendHandler(rebalancer, lb)
					case types.Wrr:
						log.Debugf("Creating load-balancer wrr")
						if sticky != nil {
							log.Debugf("Sticky session with cookie %v", cookieName)
							if s.accessLoggerMiddleware != nil {
								rr, _ = roundrobin.New(saveFrontend, roundrobin.EnableStickySession(sticky))
							} else {
								rr, _ = roundrobin.New(fwd, roundrobin.EnableStickySession(sticky))
							}
						}
						lb = rr
						if err := configureLBServers(rr, config, frontend); err != nil {
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
						hcOpts := parseHealthCheckOptions(rr, frontend.Backend, config.Backends[frontend.Backend].HealthCheck, globalConfiguration.HealthCheck)
						if hcOpts != nil {
							log.Debugf("Setting up backend health check %s", *hcOpts)
							hcOpts.Transport = s.defaultForwardingRoundTripper
							backendsHealthCheck[entryPointName+frontend.Backend] = healthcheck.NewBackendHealthCheck(*hcOpts)
						}
						lb = middlewares.NewEmptyBackendHandler(rr, lb)
					}

					if len(frontend.Errors) > 0 {
						for _, errorPage := range frontend.Errors {
							if config.Backends[errorPage.Backend] != nil && config.Backends[errorPage.Backend].Servers["error"].URL != "" {
								errorPageHandler, err := middlewares.NewErrorPagesHandler(errorPage, config.Backends[errorPage.Backend].Servers["error"].URL)
								if err != nil {
									log.Errorf("Error creating custom error page middleware, %v", err)
								} else {
									n.Use(errorPageHandler)
								}
							} else {
								log.Errorf("Error Page is configured for Frontend %s, but either Backend %s is not set or Backend URL is missing", frontendName, errorPage.Backend)
							}
						}
					}

					if frontend.RateLimit != nil && len(frontend.RateLimit.RateSet) > 0 {
						lb, err = s.buildRateLimiter(lb, frontend.RateLimit)
						if err != nil {
							log.Errorf("Error creating rate limiter: %v", err)
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
					}

					maxConns := config.Backends[frontend.Backend].MaxConn
					if maxConns != nil && maxConns.Amount != 0 {
						extractFunc, err := utils.NewExtractor(maxConns.ExtractorFunc)
						if err != nil {
							log.Errorf("Error creating connlimit: %v", err)
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
						log.Debugf("Creating load-balancer connlimit")
						lb, err = connlimit.New(lb, extractFunc, maxConns.Amount)
						if err != nil {
							log.Errorf("Error creating connlimit: %v", err)
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
					}

					if globalConfiguration.Retry != nil {
						countServers := len(config.Backends[frontend.Backend].Servers)
						lb = s.buildRetryMiddleware(lb, globalConfiguration, countServers, frontend.Backend)
					}

					if s.metricsRegistry.IsEnabled() {
						n.Use(middlewares.NewMetricsWrapper(s.metricsRegistry, frontend.Backend))
					}

					ipWhitelistMiddleware, err := configureIPWhitelistMiddleware(frontend.WhitelistSourceRange)
					if err != nil {
						log.Errorf("Error creating IP Whitelister: %s", err)
					} else if ipWhitelistMiddleware != nil {
						n.Use(ipWhitelistMiddleware)
						log.Infof("Configured IP Whitelists: %s", frontend.WhitelistSourceRange)
					}

					if frontend.Redirect != nil {
						rewrite, err := s.buildRedirectHandler(entryPointName, frontend.Redirect)
						if err != nil {
							log.Errorf("Error creating Frontend Redirect: %v", err)
						}
						n.Use(rewrite)
						log.Debugf("Frontend %s redirect created", frontendName)
					}

					if len(frontend.BasicAuth) > 0 {
						users := types.Users{}
						for _, user := range frontend.BasicAuth {
							users = append(users, user)
						}

						auth := &types.Auth{}
						auth.Basic = &types.Basic{
							Users: users,
						}
						authMiddleware, err := mauth.NewAuthenticator(auth)
						if err != nil {
							log.Errorf("Error creating Auth: %s", err)
						} else {
							n.Use(authMiddleware)
						}
					}

					if headerMiddleware != nil {
						log.Debugf("Adding header middleware for frontend %s", frontendName)
						n.Use(headerMiddleware)
					}

					secureMiddleware := middlewares.NewSecure(frontend.Headers)
					if secureMiddleware != nil {
						log.Debugf("Adding secure middleware for frontend %s", frontendName)
						n.UseFunc(secureMiddleware.HandlerFuncWithNext)
					}

					if config.Backends[frontend.Backend].CircuitBreaker != nil {
						log.Debugf("Creating circuit breaker %s", config.Backends[frontend.Backend].CircuitBreaker.Expression)
						circuitBreaker, err := middlewares.NewCircuitBreaker(lb, config.Backends[frontend.Backend].CircuitBreaker.Expression)
						if err != nil {
							log.Errorf("Error creating circuit breaker: %v", err)
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
						n.Use(circuitBreaker)
					} else {
						n.UseHandler(lb)
					}
					backends[entryPointName+frontend.Backend] = n
				} else {
					log.Debugf("Reusing backend %s", frontend.Backend)
				}
				if frontend.Priority > 0 {
					newServerRoute.route.Priority(frontend.Priority)
				}
				s.wireFrontendBackend(newServerRoute, backends[entryPointName+frontend.Backend])

				err := newServerRoute.route.GetError()
				if err != nil {
					log.Errorf("Error building route: %s", err)
				}
			}
		}
	}
	healthcheck.GetHealthCheck().SetBackendsConfiguration(s.routinesPool.Ctx(), backendsHealthCheck)
	// Get new certificates list sorted per entrypoints
	// Update certificates
	entryPointsCertificates, err := s.loadHTTPSConfiguration(configurations)
	//sort routes and update certificates
	for serverEntryPointName, serverEntryPoint := range serverEntryPoints {
		serverEntryPoint.httpRouter.GetHandler().SortRoutes()
		_, exists := entryPointsCertificates[serverEntryPointName]
		if exists {
			serverEntryPoint.certs.Set(entryPointsCertificates[serverEntryPointName])
		}
	}

	return serverEntryPoints, err
}

func configureLBServers(lb healthcheck.LoadBalancer, config *types.Configuration, frontend *types.Frontend) error {
	for serverName, server := range config.Backends[frontend.Backend].Servers {
		u, err := url.Parse(server.URL)
		if err != nil {
			log.Errorf("Error parsing server URL %s: %v", server.URL, err)
			return err
		}
		log.Debugf("Creating server %s at %s with weight %d", serverName, u, server.Weight)
		if err := lb.UpsertServer(u, roundrobin.Weight(server.Weight)); err != nil {
			log.Errorf("Error adding server %s to load balancer: %v", server.URL, err)
			return err
		}
	}
	return nil
}

func configureIPWhitelistMiddleware(whitelistSourceRanges []string) (negroni.Handler, error) {
	if len(whitelistSourceRanges) > 0 {
		ipSourceRanges := whitelistSourceRanges
		ipWhitelistMiddleware, err := middlewares.NewIPWhitelister(ipSourceRanges)

		if err != nil {
			return nil, err
		}

		return ipWhitelistMiddleware, nil
	}

	return nil, nil
}

func (s *Server) wireFrontendBackend(serverRoute *serverRoute, handler http.Handler) {
	// path replace - This needs to always be the very last on the handler chain (first in the order in this function)
	// -- Replacing Path should happen at the very end of the Modifier chain, after all the Matcher+Modifiers ran
	if len(serverRoute.replacePath) > 0 {
		handler = &middlewares.ReplacePath{
			Path:    serverRoute.replacePath,
			Handler: handler,
		}
	}

	if len(serverRoute.replacePathRegex) > 0 {
		sp := strings.Split(serverRoute.replacePathRegex, " ")
		if len(sp) == 2 {
			handler = middlewares.NewReplacePathRegexHandler(sp[0], sp[1], handler)
		} else {
			log.Warnf("Invalid syntax for ReplacePathRegex: %s. Separate the regular expression and the replacement by a space.", serverRoute.replacePathRegex)
		}
	}

	// add prefix - This needs to always be right before ReplacePath on the chain (second in order in this function)
	// -- Adding Path Prefix should happen after all *Strip Matcher+Modifiers ran, but before Replace (in case it's configured)
	if len(serverRoute.addPrefix) > 0 {
		handler = &middlewares.AddPrefix{
			Prefix:  serverRoute.addPrefix,
			Handler: handler,
		}
	}

	// strip prefix
	if len(serverRoute.stripPrefixes) > 0 {
		handler = &middlewares.StripPrefix{
			Prefixes: serverRoute.stripPrefixes,
			Handler:  handler,
		}
	}

	// strip prefix with regex
	if len(serverRoute.stripPrefixesRegex) > 0 {
		handler = middlewares.NewStripPrefixRegex(handler, serverRoute.stripPrefixesRegex)
	}

	serverRoute.route.Handler(handler)
}

func (s *Server) buildRedirectHandler(srcEntryPointName string, redirect *types.Redirect) (*middlewares.Rewrite, error) {
	// entry point redirect
	if len(redirect.EntryPoint) > 0 {
		return s.buildEntryPointRedirect(srcEntryPointName, redirect.EntryPoint)
	}

	// regex redirect
	rewrite, err := middlewares.NewRewrite(redirect.Regex, redirect.Replacement, true)
	if err != nil {
		return nil, err
	}
	log.Debugf("Creating entryPoint redirect %s -> %s -> %s", srcEntryPointName, redirect.Regex, redirect.Replacement)

	return rewrite, nil
}

func (s *Server) buildEntryPointRedirect(srcEntryPointName string, redirectEntryPoint string) (*middlewares.Rewrite, error) {
	regex, replacement, err := s.buildRedirect(redirectEntryPoint)
	if err != nil {
		return nil, err
	}

	rewrite, err := middlewares.NewRewrite(regex, replacement, true)
	if err != nil {
		// Impossible case because error is always nil
		return nil, err
	}
	log.Debugf("Creating entryPoint redirect %s -> %s : %s -> %s", srcEntryPointName, redirectEntryPoint, regex, replacement)

	return rewrite, nil
}

func (s *Server) buildRedirect(entryPointName string) (string, string, error) {
	entryPoint := s.globalConfiguration.EntryPoints[entryPointName]
	if entryPoint == nil {
		return "", "", fmt.Errorf("unknown target entrypoint %q", entryPointName)
	}

	exp := regexp.MustCompile(`(:\d+)`)
	match := exp.FindStringSubmatch(entryPoint.Address)
	if len(match) == 0 {
		return "", "", fmt.Errorf("bad Address format %q", entryPoint.Address)
	}

	var protocol = "http"
	if s.globalConfiguration.EntryPoints[entryPointName].TLS != nil {
		protocol = "https"
	}

	replacement := protocol + "://$1" + match[0] + "$2"
	return defaultRedirectRegex, replacement, nil
}

func (s *Server) buildDefaultHTTPRouter() *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	router.StrictSlash(true)
	router.SkipClean(true)
	return router
}

func parseHealthCheckOptions(lb healthcheck.LoadBalancer, backend string, hc *types.HealthCheck, hcConfig *configuration.HealthCheckConfig) *healthcheck.Options {
	if hc == nil || hc.Path == "" || hcConfig == nil {
		return nil
	}

	interval := time.Duration(hcConfig.Interval)
	if hc.Interval != "" {
		intervalOverride, err := time.ParseDuration(hc.Interval)
		switch {
		case err != nil:
			log.Errorf("Illegal healthcheck interval for backend '%s': %s", backend, err)
		case intervalOverride <= 0:
			log.Errorf("Healthcheck interval smaller than zero for backend '%s', backend", backend)
		default:
			interval = intervalOverride
		}
	}

	return &healthcheck.Options{
		Path:     hc.Path,
		Port:     hc.Port,
		Interval: interval,
		LB:       lb,
	}
}

func getRoute(serverRoute *serverRoute, route *types.Route) error {
	rules := Rules{route: serverRoute}
	newRoute, err := rules.Parse(route.Rule)
	if err != nil {
		return err
	}
	newRoute.Priority(serverRoute.route.GetPriority() + len(route.Rule))
	serverRoute.route = newRoute
	return nil
}

func sortedFrontendNamesForConfig(configuration *types.Configuration) []string {
	keys := []string{}
	for key := range configuration.Frontends {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (s *Server) configureFrontends(frontends map[string]*types.Frontend) {
	for _, frontend := range frontends {
		// default endpoints if not defined in frontends
		if len(frontend.EntryPoints) == 0 {
			frontend.EntryPoints = s.globalConfiguration.DefaultEntryPoints
		}
	}
}

func (*Server) configureBackends(backends map[string]*types.Backend) {
	for backendName := range backends {
		backend := backends[backendName]
		if backend.LoadBalancer != nil && backend.LoadBalancer.Sticky {
			log.Warnf("Deprecated configuration found: %s. Please use %s.", "backend.LoadBalancer.Sticky", "backend.LoadBalancer.Stickiness")
		}

		_, err := types.NewLoadBalancerMethod(backend.LoadBalancer)
		if err == nil {
			if backend.LoadBalancer != nil && backend.LoadBalancer.Stickiness == nil && backend.LoadBalancer.Sticky {
				backend.LoadBalancer.Stickiness = &types.Stickiness{
					CookieName: "_TRAEFIK_BACKEND",
				}
			}
		} else {
			log.Debugf("Validation of load balancer method for backend %s failed: %s. Using default method wrr.", backendName, err)

			var stickiness *types.Stickiness
			if backend.LoadBalancer != nil {
				if backend.LoadBalancer.Stickiness == nil {
					if backend.LoadBalancer.Sticky {
						stickiness = &types.Stickiness{
							CookieName: "_TRAEFIK_BACKEND",
						}
					}
				} else {
					stickiness = backend.LoadBalancer.Stickiness
				}
			}
			backend.LoadBalancer = &types.LoadBalancer{
				Method:     "wrr",
				Stickiness: stickiness,
			}
		}
	}
}

func (s *Server) registerMetricClients(metricsConfig *types.Metrics) {
	registries := []metrics.Registry{}

	if metricsConfig.Prometheus != nil {
		registries = append(registries, metrics.RegisterPrometheus(metricsConfig.Prometheus))
		log.Debug("Configured Prometheus metrics")
	}
	if metricsConfig.Datadog != nil {
		registries = append(registries, metrics.RegisterDatadog(metricsConfig.Datadog))
		log.Debugf("Configured DataDog metrics pushing to %s once every %s", metricsConfig.Datadog.Address, metricsConfig.Datadog.PushInterval)
	}
	if metricsConfig.StatsD != nil {
		registries = append(registries, metrics.RegisterStatsd(metricsConfig.StatsD))
		log.Debugf("Configured StatsD metrics pushing to %s once every %s", metricsConfig.StatsD.Address, metricsConfig.StatsD.PushInterval)
	}
	if metricsConfig.InfluxDB != nil {
		registries = append(registries, metrics.RegisterInfluxDB(metricsConfig.InfluxDB))
		log.Debugf("Configured InfluxDB metrics pushing to %s once every %s", metricsConfig.InfluxDB.Address, metricsConfig.InfluxDB.PushInterval)
	}

	if len(registries) > 0 {
		s.metricsRegistry = metrics.NewMultiRegistry(registries)
	}
}

func stopMetricsClients() {
	metrics.StopDatadog()
	metrics.StopStatsd()
	metrics.StopInfluxDB()
}

func (s *Server) buildRateLimiter(handler http.Handler, rlConfig *types.RateLimit) (http.Handler, error) {
	extractFunc, err := utils.NewExtractor(rlConfig.ExtractorFunc)
	if err != nil {
		return nil, err
	}
	log.Debugf("Creating load-balancer rate limiter")
	rateSet := ratelimit.NewRateSet()
	for _, rate := range rlConfig.RateSet {
		if err := rateSet.Add(time.Duration(rate.Period), rate.Average, rate.Burst); err != nil {
			return nil, err
		}
	}
	return ratelimit.New(handler, extractFunc, rateSet)
}

func (s *Server) buildRetryMiddleware(handler http.Handler, globalConfig configuration.GlobalConfiguration, countServers int, backendName string) http.Handler {
	retryListeners := middlewares.RetryListeners{}
	if s.metricsRegistry.IsEnabled() {
		retryListeners = append(retryListeners, middlewares.NewMetricsRetryListener(s.metricsRegistry, backendName))
	}
	if s.accessLoggerMiddleware != nil {
		retryListeners = append(retryListeners, &accesslog.SaveRetries{})
	}

	retryAttempts := countServers
	if globalConfig.Retry.Attempts > 0 {
		retryAttempts = globalConfig.Retry.Attempts
	}

	log.Debugf("Creating retries max attempts %d", retryAttempts)

	return middlewares.NewRetry(retryAttempts, handler, retryListeners)
}
