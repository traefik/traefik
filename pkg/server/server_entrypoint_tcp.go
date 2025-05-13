package server

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/containous/alice"
	gokitmetrics "github.com/go-kit/kit/metrics"
	"github.com/pires/go-proxyproto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/ip"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/metrics"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/contenttype"
	"github.com/traefik/traefik/v3/pkg/middlewares/forwardedheaders"
	"github.com/traefik/traefik/v3/pkg/middlewares/requestdecorator"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/server/router"
	tcprouter "github.com/traefik/traefik/v3/pkg/server/router/tcp"
	"github.com/traefik/traefik/v3/pkg/server/service"
	"github.com/traefik/traefik/v3/pkg/tcp"
	"github.com/traefik/traefik/v3/pkg/types"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type key string

const (
	connStateKey       key    = "connState"
	debugConnectionEnv string = "DEBUG_CONNECTION"
)

var (
	clientConnectionStates   = map[string]*connState{}
	clientConnectionStatesMu = sync.RWMutex{}
)

type connState struct {
	State            string
	KeepAliveState   string
	Start            time.Time
	HTTPRequestCount int
}

type httpForwarder struct {
	net.Listener
	connChan chan net.Conn
	errChan  chan error
}

func newHTTPForwarder(ln net.Listener) *httpForwarder {
	return &httpForwarder{
		Listener: ln,
		connChan: make(chan net.Conn),
		errChan:  make(chan error),
	}
}

// ServeTCP uses the connection to serve it later in "Accept".
func (h *httpForwarder) ServeTCP(conn tcp.WriteCloser) {
	h.connChan <- conn
}

// Accept retrieves a served connection in ServeTCP.
func (h *httpForwarder) Accept() (net.Conn, error) {
	select {
	case conn := <-h.connChan:
		return conn, nil
	case err := <-h.errChan:
		return nil, err
	}
}

// TCPEntryPoints holds a map of TCPEntryPoint (the entrypoint names being the keys).
type TCPEntryPoints map[string]*TCPEntryPoint

// NewTCPEntryPoints creates a new TCPEntryPoints.
func NewTCPEntryPoints(entryPointsConfig static.EntryPoints, hostResolverConfig *types.HostResolverConfig, metricsRegistry metrics.Registry) (TCPEntryPoints, error) {
	if os.Getenv(debugConnectionEnv) != "" {
		expvar.Publish("clientConnectionStates", expvar.Func(func() any {
			return clientConnectionStates
		}))
	}

	serverEntryPointsTCP := make(TCPEntryPoints)
	for entryPointName, config := range entryPointsConfig {
		protocol, err := config.GetProtocol()
		if err != nil {
			return nil, fmt.Errorf("error while building entryPoint %s: %w", entryPointName, err)
		}

		if protocol != "tcp" {
			continue
		}

		ctx := log.With().Str(logs.EntryPointName, entryPointName).Logger().WithContext(context.Background())

		openConnectionsGauge := metricsRegistry.
			OpenConnectionsGauge().
			With("entrypoint", entryPointName, "protocol", "TCP")

		serverEntryPointsTCP[entryPointName], err = NewTCPEntryPoint(ctx, entryPointName, config, hostResolverConfig, openConnectionsGauge)
		if err != nil {
			return nil, fmt.Errorf("error while building entryPoint %s: %w", entryPointName, err)
		}
	}
	return serverEntryPointsTCP, nil
}

// Start the server entry points.
func (eps TCPEntryPoints) Start() {
	for entryPointName, serverEntryPoint := range eps {
		ctx := log.With().Str(logs.EntryPointName, entryPointName).Logger().WithContext(context.Background())
		go serverEntryPoint.Start(ctx)
	}
}

// Stop the server entry points.
func (eps TCPEntryPoints) Stop() {
	var wg sync.WaitGroup

	for epn, ep := range eps {
		wg.Add(1)

		go func(entryPointName string, entryPoint *TCPEntryPoint) {
			defer wg.Done()

			logger := log.With().Str(logs.EntryPointName, entryPointName).Logger()
			entryPoint.Shutdown(logger.WithContext(context.Background()))

			logger.Debug().Msg("Entrypoint closed")
		}(epn, ep)
	}

	wg.Wait()
}

// Switch the TCP routers.
func (eps TCPEntryPoints) Switch(routersTCP map[string]*tcprouter.Router) {
	for entryPointName, rt := range routersTCP {
		eps[entryPointName].SwitchRouter(rt)
	}
}

// TCPEntryPoint is the TCP server.
type TCPEntryPoint struct {
	listener               net.Listener
	switcher               *tcp.HandlerSwitcher
	transportConfiguration *static.EntryPointsTransport
	tracker                *connectionTracker
	httpServer             *httpServer
	httpsServer            *httpServer

	http3Server *http3server
}

// NewTCPEntryPoint creates a new TCPEntryPoint.
func NewTCPEntryPoint(ctx context.Context, name string, config *static.EntryPoint, hostResolverConfig *types.HostResolverConfig, openConnectionsGauge gokitmetrics.Gauge) (*TCPEntryPoint, error) {
	tracker := newConnectionTracker(openConnectionsGauge)

	listener, err := buildListener(ctx, name, config)
	if err != nil {
		return nil, fmt.Errorf("error preparing server: %w", err)
	}

	rt, err := tcprouter.NewRouter()
	if err != nil {
		return nil, fmt.Errorf("error preparing tcp router: %w", err)
	}

	reqDecorator := requestdecorator.New(hostResolverConfig)

	httpServer, err := createHTTPServer(ctx, listener, config, true, reqDecorator)
	if err != nil {
		return nil, fmt.Errorf("error preparing http server: %w", err)
	}

	rt.SetHTTPForwarder(httpServer.Forwarder)

	httpsServer, err := createHTTPServer(ctx, listener, config, false, reqDecorator)
	if err != nil {
		return nil, fmt.Errorf("error preparing https server: %w", err)
	}

	h3Server, err := newHTTP3Server(ctx, name, config, httpsServer)
	if err != nil {
		return nil, fmt.Errorf("error preparing http3 server: %w", err)
	}

	rt.SetHTTPSForwarder(httpsServer.Forwarder)

	tcpSwitcher := &tcp.HandlerSwitcher{}
	tcpSwitcher.Switch(rt)

	return &TCPEntryPoint{
		listener:               listener,
		switcher:               tcpSwitcher,
		transportConfiguration: config.Transport,
		tracker:                tracker,
		httpServer:             httpServer,
		httpsServer:            httpsServer,
		http3Server:            h3Server,
	}, nil
}

// Start starts the TCP server.
func (e *TCPEntryPoint) Start(ctx context.Context) {
	logger := log.Ctx(ctx)
	logger.Debug().Msg("Starting TCP Server")

	if e.http3Server != nil {
		go func() { _ = e.http3Server.Start() }()
	}

	for {
		conn, err := e.listener.Accept()
		if err != nil {
			logger.Error().Err(err).Send()

			var opErr *net.OpError
			if errors.As(err, &opErr) && opErr.Temporary() {
				continue
			}

			var urlErr *url.Error
			if errors.As(err, &urlErr) && urlErr.Temporary() {
				continue
			}

			e.httpServer.Forwarder.errChan <- err
			e.httpsServer.Forwarder.errChan <- err

			return
		}

		writeCloser, err := writeCloser(conn)
		if err != nil {
			panic(err)
		}

		safe.Go(func() {
			// Enforce read/write deadlines at the connection level,
			// because when we're peeking the first byte to determine whether we are doing TLS,
			// the deadlines at the server level are not taken into account.
			if e.transportConfiguration.RespondingTimeouts.ReadTimeout > 0 {
				err := writeCloser.SetReadDeadline(time.Now().Add(time.Duration(e.transportConfiguration.RespondingTimeouts.ReadTimeout)))
				if err != nil {
					logger.Error().Err(err).Msg("Error while setting read deadline")
				}
			}

			if e.transportConfiguration.RespondingTimeouts.WriteTimeout > 0 {
				err = writeCloser.SetWriteDeadline(time.Now().Add(time.Duration(e.transportConfiguration.RespondingTimeouts.WriteTimeout)))
				if err != nil {
					logger.Error().Err(err).Msg("Error while setting write deadline")
				}
			}

			e.switcher.ServeTCP(newTrackedConnection(writeCloser, e.tracker))
		})
	}
}

// Shutdown stops the TCP connections.
func (e *TCPEntryPoint) Shutdown(ctx context.Context) {
	logger := log.Ctx(ctx)

	reqAcceptGraceTimeOut := time.Duration(e.transportConfiguration.LifeCycle.RequestAcceptGraceTimeout)
	if reqAcceptGraceTimeOut > 0 {
		logger.Info().Msgf("Waiting %s for incoming requests to cease", reqAcceptGraceTimeOut)
		time.Sleep(reqAcceptGraceTimeOut)
	}

	graceTimeOut := time.Duration(e.transportConfiguration.LifeCycle.GraceTimeOut)
	ctx, cancel := context.WithTimeout(ctx, graceTimeOut)
	logger.Debug().Msgf("Waiting %s seconds before killing connections", graceTimeOut)

	var wg sync.WaitGroup

	shutdownServer := func(server stoppable) {
		defer wg.Done()
		err := server.Shutdown(ctx)
		if err == nil {
			return
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logger.Debug().Err(err).Msg("Server failed to shutdown within deadline")
			if err = server.Close(); err != nil {
				logger.Error().Err(err).Send()
			}
			return
		}

		logger.Error().Err(err).Send()

		// We expect Close to fail again because Shutdown most likely failed when trying to close a listener.
		// We still call it however, to make sure that all connections get closed as well.
		server.Close()
	}

	if e.httpServer.Server != nil {
		wg.Add(1)
		go shutdownServer(e.httpServer.Server)
	}

	if e.httpsServer.Server != nil {
		wg.Add(1)
		go shutdownServer(e.httpsServer.Server)

		if e.http3Server != nil {
			wg.Add(1)
			go shutdownServer(e.http3Server)
		}
	}

	if e.tracker != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := e.tracker.Shutdown(ctx)
			if err == nil {
				return
			}
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				logger.Debug().Err(err).Msg("Server failed to shutdown before deadline")
			}
			e.tracker.Close()
		}()
	}

	wg.Wait()
	cancel()
}

// SwitchRouter switches the TCP router handler.
func (e *TCPEntryPoint) SwitchRouter(rt *tcprouter.Router) {
	rt.SetHTTPForwarder(e.httpServer.Forwarder)

	httpHandler := rt.GetHTTPHandler()
	if httpHandler == nil {
		httpHandler = router.BuildDefaultHTTPRouter()
	}

	e.httpServer.Switcher.UpdateHandler(httpHandler)

	rt.SetHTTPSForwarder(e.httpsServer.Forwarder)

	httpsHandler := rt.GetHTTPSHandler()
	if httpsHandler == nil {
		httpsHandler = router.BuildDefaultHTTPRouter()
	}

	e.httpsServer.Switcher.UpdateHandler(httpsHandler)

	e.switcher.Switch(rt)

	if e.http3Server != nil {
		e.http3Server.Switch(rt)
	}
}

// writeCloserWrapper wraps together a connection, and the concrete underlying
// connection type that was found to satisfy WriteCloser.
type writeCloserWrapper struct {
	net.Conn
	writeCloser tcp.WriteCloser
}

func (c *writeCloserWrapper) CloseWrite() error {
	return c.writeCloser.CloseWrite()
}

// writeCloser returns the given connection, augmented with the WriteCloser
// implementation, if any was found within the underlying conn.
func writeCloser(conn net.Conn) (tcp.WriteCloser, error) {
	switch typedConn := conn.(type) {
	case *proxyproto.Conn:
		underlying, ok := typedConn.TCPConn()
		if !ok {
			return nil, errors.New("underlying connection is not a tcp connection")
		}
		return &writeCloserWrapper{writeCloser: underlying, Conn: typedConn}, nil
	case *net.TCPConn:
		return typedConn, nil
	default:
		return nil, fmt.Errorf("unknown connection type %T", typedConn)
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

	if err := tc.SetKeepAlive(true); err != nil {
		return nil, err
	}

	if err := tc.SetKeepAlivePeriod(3 * time.Minute); err != nil {
		// Some systems, such as OpenBSD, have no user-settable per-socket TCP keepalive options.
		if !errors.Is(err, syscall.ENOPROTOOPT) {
			return nil, err
		}
	}

	return tc, nil
}

func buildProxyProtocolListener(ctx context.Context, entryPoint *static.EntryPoint, listener net.Listener) (net.Listener, error) {
	timeout := entryPoint.Transport.RespondingTimeouts.ReadTimeout
	// proxyproto use 200ms if ReadHeaderTimeout is set to 0 and not no timeout
	if timeout == 0 {
		timeout = -1
	}
	proxyListener := &proxyproto.Listener{Listener: listener, ReadHeaderTimeout: time.Duration(timeout)}

	if entryPoint.ProxyProtocol.Insecure {
		log.Ctx(ctx).Info().Msg("Enabling ProxyProtocol without trusted IPs: Insecure")
		return proxyListener, nil
	}

	checker, err := ip.NewChecker(entryPoint.ProxyProtocol.TrustedIPs)
	if err != nil {
		return nil, err
	}

	proxyListener.Policy = func(upstream net.Addr) (proxyproto.Policy, error) {
		ipAddr, ok := upstream.(*net.TCPAddr)
		if !ok {
			return proxyproto.REJECT, fmt.Errorf("type error %v", upstream)
		}

		if !checker.ContainsIP(ipAddr.IP) {
			log.Ctx(ctx).Debug().Msgf("IP %s is not in trusted IPs list, ignoring ProxyProtocol Headers and bypass connection", ipAddr.IP)
			return proxyproto.IGNORE, nil
		}
		return proxyproto.USE, nil
	}

	log.Ctx(ctx).Info().Msgf("Enabling ProxyProtocol for trusted IPs %v", entryPoint.ProxyProtocol.TrustedIPs)

	return proxyListener, nil
}

func buildListener(ctx context.Context, name string, config *static.EntryPoint) (net.Listener, error) {
	var listener net.Listener
	var err error

	// if we have predefined listener from socket activation
	if socketActivation.isEnabled() {
		listener, err = socketActivation.getListener(name)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("name", name).Msg("Unable to use socket activation for entrypoint")
		}
	}

	if listener == nil {
		listenConfig := newListenConfig(config)
		listener, err = listenConfig.Listen(ctx, "tcp", config.GetAddress())
		if err != nil {
			return nil, fmt.Errorf("error opening listener: %w", err)
		}
	}

	listener = tcpKeepAliveListener{listener.(*net.TCPListener)}

	if config.ProxyProtocol != nil {
		listener, err = buildProxyProtocolListener(ctx, config, listener)
		if err != nil {
			return nil, fmt.Errorf("error creating proxy protocol listener: %w", err)
		}
	}
	return listener, nil
}

func newConnectionTracker(openConnectionsGauge gokitmetrics.Gauge) *connectionTracker {
	return &connectionTracker{
		conns:                make(map[net.Conn]struct{}),
		openConnectionsGauge: openConnectionsGauge,
	}
}

type connectionTracker struct {
	connsMu sync.RWMutex
	conns   map[net.Conn]struct{}

	openConnectionsGauge gokitmetrics.Gauge
}

// AddConnection add a connection in the tracked connections list.
func (c *connectionTracker) AddConnection(conn net.Conn) {
	defer c.syncOpenConnectionGauge()

	c.connsMu.Lock()
	c.conns[conn] = struct{}{}
	c.connsMu.Unlock()
}

// RemoveConnection remove a connection from the tracked connections list.
func (c *connectionTracker) RemoveConnection(conn net.Conn) {
	defer c.syncOpenConnectionGauge()

	c.connsMu.Lock()
	delete(c.conns, conn)
	c.connsMu.Unlock()
}

// syncOpenConnectionGauge updates openConnectionsGauge value with the conns map length.
func (c *connectionTracker) syncOpenConnectionGauge() {
	if c.openConnectionsGauge == nil {
		return
	}

	c.connsMu.RLock()
	c.openConnectionsGauge.Set(float64(len(c.conns)))
	c.connsMu.RUnlock()
}

func (c *connectionTracker) isEmpty() bool {
	c.connsMu.RLock()
	defer c.connsMu.RUnlock()
	return len(c.conns) == 0
}

// Shutdown wait for the connection closing.
func (c *connectionTracker) Shutdown(ctx context.Context) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		if c.isEmpty() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// Close close all the connections in the tracked connections list.
func (c *connectionTracker) Close() {
	c.connsMu.Lock()
	defer c.connsMu.Unlock()
	for conn := range c.conns {
		if err := conn.Close(); err != nil {
			log.Error().Err(err).Msg("Error while closing connection")
		}
		delete(c.conns, conn)
	}
}

type stoppable interface {
	Shutdown(ctx context.Context) error
	Close() error
}

type stoppableServer interface {
	stoppable
	Serve(listener net.Listener) error
}

type httpServer struct {
	Server    stoppableServer
	Forwarder *httpForwarder
	Switcher  *middlewares.HTTPHandlerSwitcher
}

func createHTTPServer(ctx context.Context, ln net.Listener, configuration *static.EntryPoint, withH2c bool, reqDecorator *requestdecorator.RequestDecorator) (*httpServer, error) {
	if configuration.HTTP2.MaxConcurrentStreams < 0 {
		return nil, errors.New("max concurrent streams value must be greater than or equal to zero")
	}

	httpSwitcher := middlewares.NewHandlerSwitcher(router.BuildDefaultHTTPRouter())

	next, err := alice.New(requestdecorator.WrapHandler(reqDecorator)).Then(httpSwitcher)
	if err != nil {
		return nil, err
	}

	var handler http.Handler
	handler, err = forwardedheaders.NewXForwarded(
		configuration.ForwardedHeaders.Insecure,
		configuration.ForwardedHeaders.TrustedIPs,
		configuration.ForwardedHeaders.Connection,
		next)
	if err != nil {
		return nil, err
	}

	if configuration.HTTP.SanitizePath != nil && *configuration.HTTP.SanitizePath {
		// sanitizePath is used to clean the URL path by removing /../, /./ and duplicate slash sequences,
		// to make sure the path is interpreted by the backends as it is evaluated inside rule matchers.
		handler = sanitizePath(handler)
	}

	if configuration.HTTP.EncodeQuerySemicolons {
		handler = encodeQuerySemicolons(handler)
	} else {
		handler = http.AllowQuerySemicolons(handler)
	}

	handler = contenttype.DisableAutoDetection(handler)

	debugConnection := os.Getenv(debugConnectionEnv) != ""
	if debugConnection || (configuration.Transport != nil && (configuration.Transport.KeepAliveMaxTime > 0 || configuration.Transport.KeepAliveMaxRequests > 0)) {
		handler = newKeepAliveMiddleware(handler, configuration.Transport.KeepAliveMaxRequests, configuration.Transport.KeepAliveMaxTime)
	}

	if withH2c {
		handler = h2c.NewHandler(handler, &http2.Server{
			MaxConcurrentStreams: uint32(configuration.HTTP2.MaxConcurrentStreams),
		})
	}

	handler = denyFragment(handler)

	serverHTTP := &http.Server{
		Handler:        handler,
		ErrorLog:       stdlog.New(logs.NoLevel(log.Logger, zerolog.DebugLevel), "", 0),
		ReadTimeout:    time.Duration(configuration.Transport.RespondingTimeouts.ReadTimeout),
		WriteTimeout:   time.Duration(configuration.Transport.RespondingTimeouts.WriteTimeout),
		IdleTimeout:    time.Duration(configuration.Transport.RespondingTimeouts.IdleTimeout),
		MaxHeaderBytes: configuration.HTTP.MaxHeaderBytes,
	}
	if debugConnection || (configuration.Transport != nil && (configuration.Transport.KeepAliveMaxTime > 0 || configuration.Transport.KeepAliveMaxRequests > 0)) {
		serverHTTP.ConnContext = func(ctx context.Context, c net.Conn) context.Context {
			cState := &connState{Start: time.Now()}
			if debugConnection {
				clientConnectionStatesMu.Lock()
				clientConnectionStates[getConnKey(c)] = cState
				clientConnectionStatesMu.Unlock()
			}
			return context.WithValue(ctx, connStateKey, cState)
		}

		if debugConnection {
			serverHTTP.ConnState = func(c net.Conn, state http.ConnState) {
				clientConnectionStatesMu.Lock()
				if clientConnectionStates[getConnKey(c)] != nil {
					clientConnectionStates[getConnKey(c)].State = state.String()
				}
				clientConnectionStatesMu.Unlock()
			}
		}
	}

	prevConnContext := serverHTTP.ConnContext
	serverHTTP.ConnContext = func(ctx context.Context, c net.Conn) context.Context {
		// This adds an empty struct in order to store a RoundTripper in the ConnContext in case of Kerberos or NTLM.
		ctx = service.AddTransportOnContext(ctx)
		if prevConnContext != nil {
			return prevConnContext(ctx, c)
		}
		return ctx
	}

	// ConfigureServer configures HTTP/2 with the MaxConcurrentStreams option for the given server.
	// Also keeping behavior the same as
	// https://cs.opensource.google/go/go/+/refs/tags/go1.17.7:src/net/http/server.go;l=3262
	if !strings.Contains(os.Getenv("GODEBUG"), "http2server=0") {
		err = http2.ConfigureServer(serverHTTP, &http2.Server{
			MaxConcurrentStreams: uint32(configuration.HTTP2.MaxConcurrentStreams),
			NewWriteScheduler:    func() http2.WriteScheduler { return http2.NewPriorityWriteScheduler(nil) },
		})
		if err != nil {
			return nil, fmt.Errorf("configure HTTP/2 server: %w", err)
		}
	}

	listener := newHTTPForwarder(ln)
	go func() {
		err := serverHTTP.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Ctx(ctx).Error().Err(err).Msg("Error while starting server")
		}
	}()
	return &httpServer{
		Server:    serverHTTP,
		Forwarder: listener,
		Switcher:  httpSwitcher,
	}, nil
}

func getConnKey(conn net.Conn) string {
	return fmt.Sprintf("%s => %s", conn.RemoteAddr(), conn.LocalAddr())
}

func newTrackedConnection(conn tcp.WriteCloser, tracker *connectionTracker) *trackedConnection {
	tracker.AddConnection(conn)
	return &trackedConnection{
		WriteCloser: conn,
		tracker:     tracker,
	}
}

type trackedConnection struct {
	tracker *connectionTracker
	tcp.WriteCloser
}

func (t *trackedConnection) Close() error {
	t.tracker.RemoveConnection(t.WriteCloser)
	return t.WriteCloser.Close()
}

// This function is inspired by http.AllowQuerySemicolons.
func encodeQuerySemicolons(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.RawQuery, ";") {
			r2 := new(http.Request)
			*r2 = *req
			r2.URL = new(url.URL)
			*r2.URL = *req.URL

			r2.URL.RawQuery = strings.ReplaceAll(req.URL.RawQuery, ";", "%3B")
			// Because the reverse proxy director is building query params from requestURI it needs to be updated as well.
			r2.RequestURI = r2.URL.RequestURI()

			h.ServeHTTP(rw, r2)
		} else {
			h.ServeHTTP(rw, req)
		}
	})
}

// When go receives an HTTP request, it assumes the absence of fragment URL.
// However, it is still possible to send a fragment in the request.
// In this case, Traefik will encode the '#' character, altering the request's intended meaning.
// To avoid this behavior, the following function rejects requests that include a fragment in the URL.
func denyFragment(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.RawPath, "#") {
			log.Debug().Msgf("Rejecting request because it contains a fragment in the URL path: %s", req.URL.RawPath)
			rw.WriteHeader(http.StatusBadRequest)

			return
		}

		h.ServeHTTP(rw, req)
	})
}

// sanitizePath removes the "..", "." and duplicate slash segments from the URL.
// It cleans the request URL Path and RawPath, and updates the request URI.
func sanitizePath(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r2 := new(http.Request)
		*r2 = *req

		// Cleans the URL raw path and path.
		r2.URL = r2.URL.JoinPath()

		// Because the reverse proxy director is building query params from requestURI it needs to be updated as well.
		r2.RequestURI = r2.URL.RequestURI()

		h.ServeHTTP(rw, r2)
	})
}
