package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/armon/go-proxyproto"
	"github.com/containous/traefik/config/static"
	"github.com/containous/traefik/h2c"
	"github.com/containous/traefik/ip"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/old/configuration"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/tls/generate"
	"github.com/containous/traefik/types"
	"github.com/sirupsen/logrus"

	"github.com/xenolf/lego/acme"
)

// EntryPoints map of EntryPoint
type EntryPoints map[string]*EntryPoint

// NewEntryPoint creates a new EntryPoint
func NewEntryPoint(ctx context.Context, configuration *static.EntryPoint) (*EntryPoint, error) {
	logger := log.FromContext(ctx)
	var err error

	router := middlewares.NewHandlerSwitcher(buildDefaultHTTPRouter())
	tracker := newHijackConnectionTracker()

	listener, err := buildListener(ctx, configuration)
	if err != nil {
		logger.Fatalf("Error preparing server: %v", err)
	}

	var tlsConfig *tls.Config
	var certificateStore *traefiktls.CertificateStore
	if configuration.TLS != nil {
		certificateStore, err = buildCertificateStore(*configuration.TLS)
		if err != nil {
			return nil, fmt.Errorf("error creating certificate store: %v", err)
		}

		tlsConfig, err = buildTLSConfig(*configuration.TLS)
		if err != nil {
			return nil, fmt.Errorf("error creating TLS config: %v", err)
		}
	}

	entryPoint := &EntryPoint{
		httpRouter:              router,
		transportConfiguration:  configuration.Transport,
		hijackConnectionTracker: tracker,
		listener:                listener,
		httpServer:              buildServer(ctx, configuration, tlsConfig, router, tracker),
		Certs:                   certificateStore,
	}

	if tlsConfig != nil {
		tlsConfig.GetCertificate = entryPoint.getCertificate
	}

	return entryPoint, nil
}

// EntryPoint holds everything about the entry point (httpServer, listener etc...)
type EntryPoint struct {
	RouteAppenderFactory    RouteAppenderFactory
	httpServer              *h2c.Server
	listener                net.Listener
	httpRouter              *middlewares.HandlerSwitcher
	Certs                   *traefiktls.CertificateStore
	OnDemandListener        func(string) (*tls.Certificate, error)
	TLSALPNGetter           func(string) (*tls.Certificate, error)
	hijackConnectionTracker *hijackConnectionTracker
	transportConfiguration  *static.EntryPointsTransport
}

// Start starts listening for traffic
func (s *EntryPoint) Start(ctx context.Context) {
	logger := log.FromContext(ctx)
	logger.Infof("Starting server on %s", s.httpServer.Addr)

	var err error
	if s.httpServer.TLSConfig != nil {
		err = s.httpServer.ServeTLS(s.listener, "", "")
	} else {
		err = s.httpServer.Serve(s.listener)
	}

	if err != http.ErrServerClosed {
		logger.Error("Cannot start server: %v", err)
	}
}

// Shutdown handles the entrypoint shutdown process
func (s EntryPoint) Shutdown(ctx context.Context) {
	logger := log.FromContext(ctx)

	reqAcceptGraceTimeOut := time.Duration(s.transportConfiguration.LifeCycle.RequestAcceptGraceTimeout)
	if reqAcceptGraceTimeOut > 0 {
		logger.Infof("Waiting %s for incoming requests to cease", reqAcceptGraceTimeOut)
		time.Sleep(reqAcceptGraceTimeOut)
	}

	graceTimeOut := time.Duration(s.transportConfiguration.LifeCycle.GraceTimeOut)
	ctx, cancel := context.WithTimeout(ctx, graceTimeOut)
	logger.Debugf("Waiting %s seconds before killing connections.", graceTimeOut)

	var wg sync.WaitGroup
	if s.httpServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.httpServer.Shutdown(ctx); err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					logger.Debugf("Wait server shutdown is overdue to: %s", err)
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
					logger.Debugf("Wait hijack connection is overdue to: %s", err)
					s.hijackConnectionTracker.Close()
				}
			}
		}()
	}

	wg.Wait()
	cancel()
}

// getCertificate allows to customize tlsConfig.GetCertificate behavior to get the certificates inserted dynamically
func (s *EntryPoint) getCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	domainToCheck := types.CanonicalDomain(clientHello.ServerName)

	if s.TLSALPNGetter != nil {
		cert, err := s.TLSALPNGetter(domainToCheck)
		if err != nil {
			return nil, err
		}

		if cert != nil {
			return cert, nil
		}
	}

	bestCertificate := s.Certs.GetBestCertificate(clientHello)
	if bestCertificate != nil {
		return bestCertificate, nil
	}

	if s.OnDemandListener != nil && len(domainToCheck) > 0 {
		// Only check for an onDemandCert if there is a domain name
		return s.OnDemandListener(domainToCheck)
	}

	if s.Certs.SniStrict {
		return nil, fmt.Errorf("strict SNI enabled - No certificate found for domain: %q, closing connection", domainToCheck)
	}

	log.WithoutContext().Debugf("Serving default certificate for request: %q", domainToCheck)
	return s.Certs.DefaultCertificate, nil
}

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

func buildProxyProtocolListener(ctx context.Context, entryPoint *static.EntryPoint, listener net.Listener) (net.Listener, error) {
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

func buildServerTimeouts(entryPointsTransport static.EntryPointsTransport) (readTimeout, writeTimeout, idleTimeout time.Duration) {
	readTimeout = time.Duration(0)
	writeTimeout = time.Duration(0)
	if entryPointsTransport.RespondingTimeouts != nil {
		readTimeout = time.Duration(entryPointsTransport.RespondingTimeouts.ReadTimeout)
		writeTimeout = time.Duration(entryPointsTransport.RespondingTimeouts.WriteTimeout)
	}

	if entryPointsTransport.RespondingTimeouts != nil {
		idleTimeout = time.Duration(entryPointsTransport.RespondingTimeouts.IdleTimeout)
	} else {
		idleTimeout = configuration.DefaultIdleTimeout
	}

	return readTimeout, writeTimeout, idleTimeout
}

func buildListener(ctx context.Context, entryPoint *static.EntryPoint) (net.Listener, error) {
	listener, err := net.Listen("tcp", entryPoint.Address)
	if err != nil {
		return nil, fmt.Errorf("error opening listener: %v", err)
	}

	listener = tcpKeepAliveListener{listener.(*net.TCPListener)}

	if entryPoint.ProxyProtocol != nil {
		listener, err = buildProxyProtocolListener(ctx, entryPoint, listener)
		if err != nil {
			return nil, fmt.Errorf("error creating proxy protocol listener: %v", err)
		}
	}
	return listener, nil
}

func buildCertificateStore(tlsOption traefiktls.TLS) (*traefiktls.CertificateStore, error) {
	certificateStore := traefiktls.NewCertificateStore()
	certificateStore.DynamicCerts.Set(make(map[string]*tls.Certificate))

	certificateStore.SniStrict = tlsOption.SniStrict

	if tlsOption.DefaultCertificate != nil {
		cert, err := buildDefaultCertificate(tlsOption.DefaultCertificate)
		if err != nil {
			return nil, err
		}
		certificateStore.DefaultCertificate = cert
	} else {
		cert, err := generate.DefaultCertificate()
		if err != nil {
			return nil, err
		}
		certificateStore.DefaultCertificate = cert
	}
	return certificateStore, nil
}

func buildServer(ctx context.Context, configuration *static.EntryPoint, tlsConfig *tls.Config, router http.Handler, tracker *hijackConnectionTracker) *h2c.Server {
	logger := log.FromContext(ctx)

	readTimeout, writeTimeout, idleTimeout := buildServerTimeouts(*configuration.Transport)
	logger.
		WithField("readTimeout", readTimeout).
		WithField("writeTimeout", writeTimeout).
		WithField("idleTimeout", idleTimeout).
		Infof("Preparing server")

	return &h2c.Server{
		Server: &http.Server{
			Addr:         configuration.Address,
			Handler:      router,
			TLSConfig:    tlsConfig,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
			ErrorLog:     stdlog.New(logger.WriterLevel(logrus.DebugLevel), "", 0),
			ConnState: func(conn net.Conn, state http.ConnState) {
				switch state {
				case http.StateHijacked:
					tracker.AddHijackedConnection(conn)
				case http.StateClosed:
					tracker.RemoveHijackedConnection(conn)
				}
			},
		},
	}
}

// creates a TLS config that allows terminating HTTPS for multiple domains using SNI
func buildTLSConfig(tlsOption traefiktls.TLS) (*tls.Config, error) {
	conf := &tls.Config{}

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
