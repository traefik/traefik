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
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/armon/go-proxyproto"
	"github.com/containous/mux"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/configuration/router"
	"github.com/containous/traefik/h2c"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/metrics"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/middlewares/accesslog"
	"github.com/containous/traefik/middlewares/tracing"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/containous/traefik/whitelist"
	"github.com/go-acme/lego/challenge/tlsalpn01"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

var httpServerLogger = stdlog.New(log.WriterLevel(logrus.DebugLevel), "", 0)

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
			log.Errorf("Error while closing Hijacked conn: %v", err)
		}
		delete(h.conns, conn)
	}
}

// Server is the reverse-proxy/load-balancer engine
type Server struct {
	serverEntryPoints             serverEntryPoints
	configurationChan             chan types.ConfigMessage
	configurationValidatedChan    chan types.ConfigMessage
	signals                       chan os.Signal
	stopChan                      chan bool
	currentConfigurations         safe.Safe
	providerConfigUpdateMap       map[string]chan types.ConfigMessage
	globalConfiguration           configuration.GlobalConfiguration
	accessLoggerMiddleware        *accesslog.LogHandler
	tracingMiddleware             *tracing.Tracing
	routinesPool                  *safe.Pool
	leadership                    *cluster.Leadership
	defaultForwardingRoundTripper http.RoundTripper
	metricsRegistry               metrics.Registry
	provider                      provider.Provider
	configurationListeners        []func(types.Configuration)
	entryPoints                   map[string]EntryPoint
	bufferPool                    httputil.BufferPool
}

// EntryPoint entryPoint information (configuration + internalRouter)
type EntryPoint struct {
	InternalRouter   types.InternalRouter
	Configuration    *configuration.EntryPoint
	OnDemandListener func(string) (*tls.Certificate, error)
	TLSALPNGetter    func(string) (*tls.Certificate, error)
	CertificateStore *traefiktls.CertificateStore
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
					log.Debugf("Wait server shutdown is over due to: %s", err)
					err = s.httpServer.Close()
					if err != nil {
						log.Error(err)
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
					log.Debugf("Wait hijack connection is over due to: %s", err)
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
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

// NewServer returns an initialized Server.
func NewServer(globalConfiguration configuration.GlobalConfiguration, provider provider.Provider, entrypoints map[string]EntryPoint) *Server {
	server := &Server{}

	server.entryPoints = entrypoints
	server.provider = provider
	server.globalConfiguration = globalConfiguration
	server.serverEntryPoints = make(map[string]*serverEntryPoint)
	server.configurationChan = make(chan types.ConfigMessage, 100)
	server.configurationValidatedChan = make(chan types.ConfigMessage, 100)
	server.signals = make(chan os.Signal, 1)
	server.stopChan = make(chan bool, 1)
	server.configureSignals()
	currentConfigurations := make(types.Configurations)
	server.currentConfigurations.Set(currentConfigurations)
	server.providerConfigUpdateMap = make(map[string]chan types.ConfigMessage)

	if server.globalConfiguration.API != nil {
		server.globalConfiguration.API.CurrentConfigurations = &server.currentConfigurations
	}

	server.bufferPool = newBufferPool()

	server.routinesPool = safe.NewPool(context.Background())

	transport, err := createHTTPTransport(globalConfiguration)
	if err != nil {
		log.Errorf("failed to create HTTP transport: %v", err)
	}

	server.defaultForwardingRoundTripper = transport

	server.tracingMiddleware = globalConfiguration.Tracing
	if server.tracingMiddleware != nil && server.tracingMiddleware.Backend != "" {
		server.tracingMiddleware.Setup()
	}

	server.metricsRegistry = registerMetricClients(globalConfiguration.Metrics)

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
		log.Info("I have to go...")
		reqAcceptGraceTimeOut := time.Duration(s.globalConfiguration.LifeCycle.RequestAcceptGraceTimeout)
		if reqAcceptGraceTimeOut > 0 {
			log.Infof("Waiting %s for incoming requests to cease", reqAcceptGraceTimeOut)
			time.Sleep(reqAcceptGraceTimeOut)
		}
		log.Info("Stopping server gracefully")
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
	defer log.Info("Server stopped")
	var wg sync.WaitGroup
	for sepn, sep := range s.serverEntryPoints {
		wg.Add(1)
		go func(serverEntryPointName string, serverEntryPoint *serverEntryPoint) {
			defer wg.Done()
			graceTimeOut := time.Duration(s.globalConfiguration.LifeCycle.GraceTimeOut)
			ctx, cancel := context.WithTimeout(context.Background(), graceTimeOut)
			log.Debugf("Waiting %s seconds before killing connections on entrypoint %s...", graceTimeOut, serverEntryPointName)
			serverEntryPoint.Shutdown(ctx)
			cancel()
			log.Debugf("Entrypoint %s closed", serverEntryPointName)
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
	s.serverEntryPoints = s.buildServerEntryPoints()

	for newServerEntryPointName, newServerEntryPoint := range s.serverEntryPoints {
		serverEntryPoint := s.setupServerEntryPoint(newServerEntryPointName, newServerEntryPoint)
		go s.startServer(serverEntryPoint)
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
func (s *Server) AddListener(listener func(types.Configuration)) {
	if s.configurationListeners == nil {
		s.configurationListeners = make([]func(types.Configuration), 0)
	}
	s.configurationListeners = append(s.configurationListeners, listener)
}

// getCertificate allows to customize tlsConfig.GetCertificate behaviour to get the certificates inserted dynamically
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

	log.Debugf("Serving default cert for request: %q", domainToCheck)
	return s.certs.DefaultCertificate, nil
}

func (s *Server) startProvider() {
	// start providers
	jsonConf, err := json.Marshal(s.provider)
	if err != nil {
		log.Debugf("Unable to marshal provider conf %T with error: %v", s.provider, err)
	}
	log.Infof("Starting provider %T %s", s.provider, jsonConf)
	currentProvider := s.provider
	safe.Go(func() {
		err := currentProvider.Provide(s.configurationChan, s.routinesPool)
		if err != nil {
			log.Errorf("Error starting provider %T: %s", s.provider, err)
		}
	})
}

// creates a TLS config that allows terminating HTTPS for multiple domains using SNI
func (s *Server) createTLSConfig(entryPointName string, tlsOption *traefiktls.TLS, router *middlewares.HandlerSwitcher) (*tls.Config, error) {
	if tlsOption == nil {
		return nil, nil
	}

	config, err := tlsOption.Certificates.CreateTLSConfig(entryPointName)
	if err != nil {
		return nil, err
	}

	s.serverEntryPoints[entryPointName].certs.DynamicCerts.Set(make(map[string]*tls.Certificate))

	// ensure http2 enabled
	config.NextProtos = []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol}

	if len(tlsOption.ClientCAFiles) > 0 {
		log.Warnf("Deprecated configuration found during TLS configuration creation: %s. Please use %s (which allows to make the CA Files optional).", "tls.ClientCAFiles", "tls.ClientCA.files")
		tlsOption.ClientCA.Files = tlsOption.ClientCAFiles
		tlsOption.ClientCA.Optional = false
	}

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
		config.ClientCAs = pool
		if tlsOption.ClientCA.Optional {
			config.ClientAuth = tls.VerifyClientCertIfGiven
		} else {
			config.ClientAuth = tls.RequireAndVerifyClientCert
		}
	}

	if s.globalConfiguration.ACME != nil && entryPointName == s.globalConfiguration.ACME.EntryPoint {
		checkOnDemandDomain := func(domain string) bool {
			routeMatch := &mux.RouteMatch{}
			match := router.GetHandler().Match(&http.Request{URL: &url.URL{}, Host: domain}, routeMatch)
			if match && routeMatch.Route != nil {
				return true
			}
			return false
		}

		err := s.globalConfiguration.ACME.CreateClusterConfig(s.leadership, config, s.serverEntryPoints[entryPointName].certs.DynamicCerts, checkOnDemandDomain)
		if err != nil {
			return nil, err
		}
	} else {
		config.GetCertificate = s.serverEntryPoints[entryPointName].getCertificate
		if len(config.Certificates) != 0 {
			certMap := s.buildNameOrIPToCertificate(config.Certificates)

			if s.entryPoints[entryPointName].CertificateStore != nil {
				s.entryPoints[entryPointName].CertificateStore.StaticCerts.Set(certMap)
			}
		}

		// Remove certs from the TLS config object
		config.Certificates = []tls.Certificate{}
	}

	// Set the minimum TLS version if set in the config TOML
	if minConst, exists := traefiktls.MinVersion[s.entryPoints[entryPointName].Configuration.TLS.MinVersion]; exists {
		config.PreferServerCipherSuites = true
		config.MinVersion = minConst
	}

	// Set the list of CipherSuites if set in the config TOML
	if s.entryPoints[entryPointName].Configuration.TLS.CipherSuites != nil {
		// if our list of CipherSuites is defined in the entrypoint config, we can re-initilize the suites list as empty
		config.CipherSuites = make([]uint16, 0)
		for _, cipher := range s.entryPoints[entryPointName].Configuration.TLS.CipherSuites {
			if cipherConst, exists := traefiktls.CipherSuites[cipher]; exists {
				config.CipherSuites = append(config.CipherSuites, cipherConst)
			} else {
				// CipherSuite listed in the toml does not exist in our listed
				return nil, fmt.Errorf("invalid CipherSuite: %s", cipher)
			}
		}
	}

	return config, nil
}

func (s *Server) startServer(serverEntryPoint *serverEntryPoint) {
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

func (s *Server) setupServerEntryPoint(newServerEntryPointName string, newServerEntryPoint *serverEntryPoint) *serverEntryPoint {
	serverMiddlewares, err := s.buildServerEntryPointMiddlewares(newServerEntryPointName, newServerEntryPoint)
	if err != nil {
		log.Fatal("Error preparing server: ", err)
	}

	newSrv, listener, err := s.prepareServer(newServerEntryPointName, s.entryPoints[newServerEntryPointName].Configuration, newServerEntryPoint.httpRouter, serverMiddlewares)
	if err != nil {
		log.Fatal("Error preparing server: ", err)
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

func (s *Server) prepareServer(entryPointName string, entryPoint *configuration.EntryPoint, router *middlewares.HandlerSwitcher, middlewares []negroni.Handler) (*h2c.Server, net.Listener, error) {
	readTimeout, writeTimeout, idleTimeout := buildServerTimeouts(s.globalConfiguration)
	log.Infof("Preparing server %s %+v with readTimeout=%s writeTimeout=%s idleTimeout=%s", entryPointName, entryPoint, readTimeout, writeTimeout, idleTimeout)

	// middlewares
	n := negroni.New()
	for _, middleware := range middlewares {
		n.Use(middleware)
	}
	n.UseHandler(router)

	internalMuxRouter := s.buildInternalRouter(entryPointName)
	internalMuxRouter.NotFoundHandler = n

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
		listener, err = buildProxyProtocolListener(entryPoint, listener)
		if err != nil {
			return nil, nil, err
		}
	}

	return &h2c.Server{
			Server: &http.Server{
				Addr:         entryPoint.Address,
				Handler:      internalMuxRouter,
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

func buildProxyProtocolListener(entryPoint *configuration.EntryPoint, listener net.Listener) (net.Listener, error) {
	IPs, err := whitelist.NewIP(entryPoint.ProxyProtocol.TrustedIPs, entryPoint.ProxyProtocol.Insecure, false)
	if err != nil {
		return nil, fmt.Errorf("error creating whitelist: %s", err)
	}

	log.Infof("Enabling ProxyProtocol for trusted IPs %v", entryPoint.ProxyProtocol.TrustedIPs)

	return &proxyproto.Listener{
		Listener: listener,
		SourceCheck: func(addr net.Addr) (bool, error) {
			ip, ok := addr.(*net.TCPAddr)
			if !ok {
				return false, fmt.Errorf("type error %v", addr)
			}

			return IPs.ContainsIP(ip.IP), nil
		},
	}, nil
}

func (s *Server) buildInternalRouter(entryPointName string) *mux.Router {
	internalMuxRouter := mux.NewRouter()
	internalMuxRouter.StrictSlash(!s.globalConfiguration.KeepTrailingSlash)
	internalMuxRouter.SkipClean(true)

	if entryPoint, ok := s.entryPoints[entryPointName]; ok && entryPoint.InternalRouter != nil {
		entryPoint.InternalRouter.AddRoutes(internalMuxRouter)

		if s.globalConfiguration.API != nil && s.globalConfiguration.API.EntryPoint == entryPointName && s.leadership != nil {
			if s.globalConfiguration.Web != nil && s.globalConfiguration.Web.Path != "" {
				rt := router.WithPrefix{Router: s.leadership, PathPrefix: s.globalConfiguration.Web.Path}
				rt.AddRoutes(internalMuxRouter)
			} else {
				s.leadership.AddRoutes(internalMuxRouter)
			}
		}
	}

	return internalMuxRouter
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
		prometheusRegister := metrics.RegisterPrometheus(metricsConfig.Prometheus)
		if prometheusRegister != nil {
			registries = append(registries, prometheusRegister)
			log.Debug("Configured Prometheus metrics")
		}
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
			certMap[strings.ToLower(x509Cert.Subject.CommonName)] = cert
		}
		for _, san := range x509Cert.DNSNames {
			certMap[strings.ToLower(san)] = cert
		}
		for _, ipSan := range x509Cert.IPAddresses {
			certMap[strings.ToLower(ipSan.String())] = cert
		}
	}
	return certMap
}
