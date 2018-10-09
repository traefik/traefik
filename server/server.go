package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/armon/go-proxyproto"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/config/static"
	"github.com/containous/traefik/h2c"
	"github.com/containous/traefik/ip"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/metrics"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/middlewares/accesslog"
	"github.com/containous/traefik/middlewares/requestdecorator"
	"github.com/containous/traefik/old/configuration"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/server/middleware"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/tracing"
	"github.com/containous/traefik/tracing/datadog"
	"github.com/containous/traefik/tracing/jaeger"
	"github.com/containous/traefik/tracing/zipkin"
	"github.com/containous/traefik/types"
	"github.com/sirupsen/logrus"
	"github.com/xenolf/lego/acme"
)

func newHijackConnectionTracker() *hijackConnectionTracker {
	return &hijackConnectionTracker{
		conns: make(map[net.Conn]struct{}),
	}
}

type hijackConnectionTracker struct {
	conns map[net.Conn]struct{}
	lock  sync.RWMutex
}

// AddHijackedConnection add a connection in the tracked connections list
func (h *hijackConnectionTracker) AddHijackedConnection(conn net.Conn) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.conns[conn] = struct{}{}
}

// RemoveHijackedConnection remove a connection from the tracked connections list
func (h *hijackConnectionTracker) RemoveHijackedConnection(conn net.Conn) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.conns, conn)
}

// Shutdown wait for the connection closing
func (h *hijackConnectionTracker) Shutdown(ctx context.Context) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		h.lock.RLock()
		if len(h.conns) == 0 {
			return nil
		}
		h.lock.RUnlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// Close close all the connections in the tracked connections list
func (h *hijackConnectionTracker) Close() {
	for conn := range h.conns {
		if err := conn.Close(); err != nil {
			log.WithoutContext().Errorf("Error while closing Hijacked connection: %v", err)
		}
		delete(h.conns, conn)
	}
}

// Server is the reverse-proxy/load-balancer engine
type Server struct {
	serverEntryPoints          serverEntryPoints
	configurationChan          chan config.Message
	configurationValidatedChan chan config.Message
	signals                    chan os.Signal
	stopChan                   chan bool
	currentConfigurations      safe.Safe
	providerConfigUpdateMap    map[string]chan config.Message
	globalConfiguration        configuration.GlobalConfiguration
	accessLoggerMiddleware     *accesslog.Handler
	tracer                     *tracing.Tracing
	routinesPool               *safe.Pool
	leadership                 *cluster.Leadership
	defaultRoundTripper        http.RoundTripper
	metricsRegistry            metrics.Registry
	provider                   provider.Provider
	configurationListeners     []func(config.Configuration)
	entryPoints                map[string]EntryPoint
	requestDecorator           *requestdecorator.RequestDecorator
}

// RouteAppenderFactory the route appender factory interface
type RouteAppenderFactory interface {
	NewAppender(ctx context.Context, middlewaresBuilder *middleware.Builder, currentConfigurations *safe.Safe) types.RouteAppender
}

// EntryPoint entryPoint information (configuration + internalRouter)
type EntryPoint struct {
	RouteAppenderFactory RouteAppenderFactory
	Configuration        *configuration.EntryPoint
	OnDemandListener     func(string) (*tls.Certificate, error)
	TLSALPNGetter        func(string) (*tls.Certificate, error)
	CertificateStore     *traefiktls.CertificateStore
}

type serverEntryPoints map[string]*serverEntryPoint

type serverEntryPoint struct {
	httpServer              *h2c.Server
	listener                net.Listener
	httpRouter              *middlewares.HandlerSwitcher
	certs                   *traefiktls.CertificateStore
	onDemandListener        func(string) (*tls.Certificate, error)
	tlsALPNGetter           func(string) (*tls.Certificate, error)
	hijackConnectionTracker *hijackConnectionTracker
}

func (s serverEntryPoint) Shutdown(ctx context.Context) {
	var wg sync.WaitGroup
	if s.httpServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.httpServer.Shutdown(ctx); err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					logger := log.FromContext(ctx)
					logger.Debugf("Wait server shutdown is over due to: %s", err)
					err = s.httpServer.Close()
					if err != nil {
						logger.Error(err)
					}
				}
			}
		}()
	}

	if s.hijackConnectionTracker != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.hijackConnectionTracker.Shutdown(ctx); err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					logger := log.FromContext(ctx)
					logger.Debugf("Wait hijack connection is over due to: %s", err)
					s.hijackConnectionTracker.Close()
				}
			}
		}()
	}

	wg.Wait()
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}

	if err = tc.SetKeepAlive(true); err != nil {
		return nil, err
	}

	if err = tc.SetKeepAlivePeriod(3 * time.Minute); err != nil {
		return nil, err
	}

	return tc, nil
}

func setupTracing(conf *static.Tracing) tracing.TrackingBackend {
	switch conf.Backend {
	case jaeger.Name:
		return conf.Jaeger
	case zipkin.Name:
		return conf.Zipkin
	case datadog.Name:
		return conf.DataDog
	default:
		log.WithoutContext().Warnf("Could not initialize tracing: unknown tracer %q", conf.Backend)
		return nil
	}
}

// NewServer returns an initialized Server.
func NewServer(globalConfiguration configuration.GlobalConfiguration, provider provider.Provider, entrypoints map[string]EntryPoint) *Server {
	server := &Server{}

	server.entryPoints = entrypoints
	server.provider = provider
	server.globalConfiguration = globalConfiguration
	server.serverEntryPoints = make(map[string]*serverEntryPoint)
	server.configurationChan = make(chan config.Message, 100)
	server.configurationValidatedChan = make(chan config.Message, 100)
	server.signals = make(chan os.Signal, 1)
	server.stopChan = make(chan bool, 1)
	server.configureSignals()
	currentConfigurations := make(config.Configurations)
	server.currentConfigurations.Set(currentConfigurations)
	server.providerConfigUpdateMap = make(map[string]chan config.Message)

	transport, err := createHTTPTransport(globalConfiguration)
	if err != nil {
		log.WithoutContext().Error(err)
		server.defaultRoundTripper = http.DefaultTransport
	} else {
		server.defaultRoundTripper = transport
	}

	if server.globalConfiguration.API != nil {
		server.globalConfiguration.API.CurrentConfigurations = &server.currentConfigurations
	}

	server.routinesPool = safe.NewPool(context.Background())

	if globalConfiguration.Tracing != nil {
		trackingBackend := setupTracing(static.ConvertTracing(globalConfiguration.Tracing))
		var err error
		server.tracer, err = tracing.NewTracing(globalConfiguration.Tracing.ServiceName, globalConfiguration.Tracing.SpanNameLimit, trackingBackend)
		if err != nil {
			log.WithoutContext().Warnf("Unable to create tracer: %v", err)
		}
	}

	server.requestDecorator = requestdecorator.New(static.ConvertHostResolverConfig(globalConfiguration.HostResolver))

	server.metricsRegistry = registerMetricClients(static.ConvertMetrics(globalConfiguration.Metrics))

	if globalConfiguration.Cluster != nil {
		// leadership creation if cluster mode
		server.leadership = cluster.NewLeadership(server.routinesPool.Ctx(), globalConfiguration.Cluster)
	}

	if globalConfiguration.AccessLog != nil {
		var err error
		server.accessLoggerMiddleware, err = accesslog.NewHandler(static.ConvertAccessLog(globalConfiguration.AccessLog))
		if err != nil {
			log.WithoutContext().Warnf("Unable to create access logger : %v", err)
		}
	}
	return server
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
	s.startProvider()
	s.routinesPool.Go(func(stop chan bool) {
		s.listenSignals(stop)
	})
}

// StartWithContext starts the server and Stop/Close it when context is Done
func (s *Server) StartWithContext(ctx context.Context) {
	go func() {
		defer s.Close()
		<-ctx.Done()
		logger := log.FromContext(ctx)
		logger.Info("I have to go...")

		reqAcceptGraceTimeOut := time.Duration(s.globalConfiguration.LifeCycle.RequestAcceptGraceTimeout)
		if reqAcceptGraceTimeOut > 0 {
			logger.Infof("Waiting %s for incoming requests to cease", reqAcceptGraceTimeOut)
			time.Sleep(reqAcceptGraceTimeOut)
		}

		logger.Info("Stopping server gracefully")
		s.Stop()
	}()
	s.Start()
}

// Wait blocks until server is shutted down.
func (s *Server) Wait() {
	<-s.stopChan
}

// Stop stops the server
func (s *Server) Stop() {
	defer log.WithoutContext().Info("Server stopped")

	var wg sync.WaitGroup
	for sepn, sep := range s.serverEntryPoints {
		wg.Add(1)
		go func(serverEntryPointName string, serverEntryPoint *serverEntryPoint) {
			defer wg.Done()
			logger := log.WithoutContext().WithField(log.EntryPointName, serverEntryPointName)

			graceTimeOut := time.Duration(s.globalConfiguration.LifeCycle.GraceTimeOut)
			ctx, cancel := context.WithTimeout(context.Background(), graceTimeOut)
			logger.Debugf("Waiting %s seconds before killing connections on entrypoint %s...", graceTimeOut, serverEntryPointName)

			serverEntryPoint.Shutdown(ctx)
			cancel()

			logger.Debugf("Entry point %s closed", serverEntryPointName)
		}(sepn, sep)
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
	s.stopLeadership()
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
	s.serverEntryPoints = s.buildServerEntryPoints()

	for newServerEntryPointName, newServerEntryPoint := range s.serverEntryPoints {
		ctx := log.With(context.Background(), log.Str(log.EntryPointName, newServerEntryPointName))
		serverEntryPoint := s.setupServerEntryPoint(ctx, newServerEntryPointName, newServerEntryPoint)
		go s.startServer(ctx, serverEntryPoint)
	}
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

// AddListener adds a new listener function used when new configuration is provided
func (s *Server) AddListener(listener func(config.Configuration)) {
	if s.configurationListeners == nil {
		s.configurationListeners = make([]func(config.Configuration), 0)
	}
	s.configurationListeners = append(s.configurationListeners, listener)
}

// getCertificate allows to customize tlsConfig.GetCertificate behavior to get the certificates inserted dynamically
func (s *serverEntryPoint) getCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	domainToCheck := types.CanonicalDomain(clientHello.ServerName)

	if s.tlsALPNGetter != nil {
		cert, err := s.tlsALPNGetter(domainToCheck)
		if err != nil {
			return nil, err
		}

		if cert != nil {
			return cert, nil
		}
	}

	bestCertificate := s.certs.GetBestCertificate(clientHello)
	if bestCertificate != nil {
		return bestCertificate, nil
	}

	if s.onDemandListener != nil && len(domainToCheck) > 0 {
		// Only check for an onDemandCert if there is a domain name
		return s.onDemandListener(domainToCheck)
	}

	if s.certs.SniStrict {
		return nil, fmt.Errorf("strict SNI enabled - No certificate found for domain: %q, closing connection", domainToCheck)
	}

	log.WithoutContext().Debugf("Serving default certificate for request: %q", domainToCheck)
	return s.certs.DefaultCertificate, nil
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

// creates a TLS config that allows terminating HTTPS for multiple domains using SNI
func (s *Server) createTLSConfig(entryPointName string, tlsOption *traefiktls.TLS, router *middlewares.HandlerSwitcher) (*tls.Config, error) {
	if tlsOption == nil {
		return nil, nil
	}

	conf, err := tlsOption.Certificates.CreateTLSConfig(entryPointName)
	if err != nil {
		return nil, err
	}

	s.serverEntryPoints[entryPointName].certs.DynamicCerts.Set(make(map[string]*tls.Certificate))

	// ensure http2 enabled
	conf.NextProtos = []string{"h2", "http/1.1", acme.ACMETLS1Protocol}

	if len(tlsOption.ClientCA.Files) > 0 {
		pool := x509.NewCertPool()
		for _, caFile := range tlsOption.ClientCA.Files {
			data, err := caFile.Read()
			if err != nil {
				return nil, err
			}
			ok := pool.AppendCertsFromPEM(data)
			if !ok {
				return nil, fmt.Errorf("invalid certificate(s) in %s", caFile)
			}
		}
		conf.ClientCAs = pool
		if tlsOption.ClientCA.Optional {
			conf.ClientAuth = tls.VerifyClientCertIfGiven
		} else {
			conf.ClientAuth = tls.RequireAndVerifyClientCert
		}
	}

	// FIXME onDemand
	if s.globalConfiguration.ACME != nil {
		// if entryPointName == s.globalConfiguration.ACME.EntryPoint {
		// 	checkOnDemandDomain := func(domain string) bool {
		// 		routeMatch := &mux.RouteMatch{}
		// 		match := router.GetHandler().Match(&http.Request{URL: &url.URL{}, Host: domain}, routeMatch)
		// 		if match && routeMatch.Route != nil {
		// 			return true
		// 		}
		// 		return false
		// 	}
		//
		// 	err := s.globalConfiguration.ACME.CreateClusterConfig(s.leadership, config, s.serverEntryPoints[entryPointName].certs.DynamicCerts, checkOnDemandDomain)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }
	} else {
		conf.GetCertificate = s.serverEntryPoints[entryPointName].getCertificate
	}

	if len(conf.Certificates) != 0 {
		certMap := s.buildNameOrIPToCertificate(conf.Certificates)

		if s.entryPoints[entryPointName].CertificateStore != nil {
			s.entryPoints[entryPointName].CertificateStore.StaticCerts.Set(certMap)
		}
	}

	// Remove certs from the TLS config object
	conf.Certificates = []tls.Certificate{}

	// Set the minimum TLS version if set in the config TOML
	if minConst, exists := traefiktls.MinVersion[tlsOption.MinVersion]; exists {
		conf.PreferServerCipherSuites = true
		conf.MinVersion = minConst
	}

	// Set the list of CipherSuites if set in the config TOML
	if tlsOption.CipherSuites != nil {
		// if our list of CipherSuites is defined in the entryPoint config, we can re-initialize the suites list as empty
		conf.CipherSuites = make([]uint16, 0)
		for _, cipher := range tlsOption.CipherSuites {
			if cipherConst, exists := traefiktls.CipherSuites[cipher]; exists {
				conf.CipherSuites = append(conf.CipherSuites, cipherConst)
			} else {
				// CipherSuite listed in the toml does not exist in our listed
				return nil, fmt.Errorf("invalid CipherSuite: %s", cipher)
			}
		}
	}

	return conf, nil
}

func (s *Server) startServer(ctx context.Context, serverEntryPoint *serverEntryPoint) {
	logger := log.FromContext(ctx)
	logger.Infof("Starting server on %s", serverEntryPoint.httpServer.Addr)

	var err error
	if serverEntryPoint.httpServer.TLSConfig != nil {
		err = serverEntryPoint.httpServer.ServeTLS(serverEntryPoint.listener, "", "")
	} else {
		err = serverEntryPoint.httpServer.Serve(serverEntryPoint.listener)
	}

	if err != http.ErrServerClosed {
		logger.Error("Cannot create server: %v", err)
	}
}

func (s *Server) setupServerEntryPoint(ctx context.Context, newServerEntryPointName string, newServerEntryPoint *serverEntryPoint) *serverEntryPoint {
	newSrv, listener, err := s.prepareServer(ctx, newServerEntryPointName, s.entryPoints[newServerEntryPointName].Configuration, newServerEntryPoint.httpRouter)
	if err != nil {
		log.FromContext(ctx).Fatalf("Error preparing server: %v", err)
	}

	serverEntryPoint := s.serverEntryPoints[newServerEntryPointName]
	serverEntryPoint.httpServer = newSrv
	serverEntryPoint.listener = listener

	serverEntryPoint.hijackConnectionTracker = newHijackConnectionTracker()
	serverEntryPoint.httpServer.ConnState = func(conn net.Conn, state http.ConnState) {
		switch state {
		case http.StateHijacked:
			serverEntryPoint.hijackConnectionTracker.AddHijackedConnection(conn)
		case http.StateClosed:
			serverEntryPoint.hijackConnectionTracker.RemoveHijackedConnection(conn)
		}
	}

	return serverEntryPoint
}

func (s *Server) prepareServer(ctx context.Context, entryPointName string, entryPoint *configuration.EntryPoint, router *middlewares.HandlerSwitcher) (*h2c.Server, net.Listener, error) {
	logger := log.FromContext(ctx)

	readTimeout, writeTimeout, idleTimeout := buildServerTimeouts(s.globalConfiguration)
	logger.
		WithField("readTimeout", readTimeout).
		WithField("writeTimeout", writeTimeout).
		WithField("idleTimeout", idleTimeout).
		Infof("Preparing server %+v", entryPoint)

	tlsConfig, err := s.createTLSConfig(entryPointName, entryPoint.TLS, router)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating TLS config: %v", err)
	}

	listener, err := net.Listen("tcp", entryPoint.Address)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening listener: %v", err)
	}

	listener = tcpKeepAliveListener{listener.(*net.TCPListener)}

	if entryPoint.ProxyProtocol != nil {
		listener, err = buildProxyProtocolListener(ctx, entryPoint, listener)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating proxy protocol listener: %v", err)
		}
	}

	httpServerLogger := stdlog.New(logger.WriterLevel(logrus.DebugLevel), "", 0)

	return &h2c.Server{
			Server: &http.Server{
				Addr:         entryPoint.Address,
				Handler:      router,
				TLSConfig:    tlsConfig,
				ReadTimeout:  readTimeout,
				WriteTimeout: writeTimeout,
				IdleTimeout:  idleTimeout,
				ErrorLog:     httpServerLogger,
			},
		},
		listener,
		nil
}

func buildProxyProtocolListener(ctx context.Context, entryPoint *configuration.EntryPoint, listener net.Listener) (net.Listener, error) {
	var sourceCheck func(addr net.Addr) (bool, error)
	if entryPoint.ProxyProtocol.Insecure {
		sourceCheck = func(_ net.Addr) (bool, error) {
			return true, nil
		}
	} else {
		checker, err := ip.NewChecker(entryPoint.ProxyProtocol.TrustedIPs)
		if err != nil {
			return nil, err
		}

		sourceCheck = func(addr net.Addr) (bool, error) {
			ipAddr, ok := addr.(*net.TCPAddr)
			if !ok {
				return false, fmt.Errorf("type error %v", addr)
			}

			return checker.ContainsIP(ipAddr.IP), nil
		}
	}

	log.FromContext(ctx).Infof("Enabling ProxyProtocol for trusted IPs %v", entryPoint.ProxyProtocol.TrustedIPs)

	return &proxyproto.Listener{
		Listener:    listener,
		SourceCheck: sourceCheck,
	}, nil
}

func buildServerTimeouts(globalConfig configuration.GlobalConfiguration) (readTimeout, writeTimeout, idleTimeout time.Duration) {
	readTimeout = time.Duration(0)
	writeTimeout = time.Duration(0)
	if globalConfig.RespondingTimeouts != nil {
		readTimeout = time.Duration(globalConfig.RespondingTimeouts.ReadTimeout)
		writeTimeout = time.Duration(globalConfig.RespondingTimeouts.WriteTimeout)
	}

	if globalConfig.RespondingTimeouts != nil {
		idleTimeout = time.Duration(globalConfig.RespondingTimeouts.IdleTimeout)
	} else {
		idleTimeout = configuration.DefaultIdleTimeout
	}

	return readTimeout, writeTimeout, idleTimeout
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

	if metricsConfig.Datadog != nil {
		ctx := log.With(context.Background(), log.Str(log.MetricsProviderName, "datadog"))
		registries = append(registries, metrics.RegisterDatadog(ctx, metricsConfig.Datadog))
		log.FromContext(ctx).Debugf("Configured DataDog metrics: pushing to %s once every %s",
			metricsConfig.Datadog.Address, metricsConfig.Datadog.PushInterval)
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

func (s *Server) buildNameOrIPToCertificate(certs []tls.Certificate) map[string]*tls.Certificate {
	certMap := make(map[string]*tls.Certificate)
	for i := range certs {
		cert := &certs[i]
		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			continue
		}
		if len(x509Cert.Subject.CommonName) > 0 {
			certMap[x509Cert.Subject.CommonName] = cert
		}
		for _, san := range x509Cert.DNSNames {
			certMap[san] = cert
		}
		for _, ipSan := range x509Cert.IPAddresses {
			certMap[ipSan.String()] = cert
		}
	}
	return certMap
}
