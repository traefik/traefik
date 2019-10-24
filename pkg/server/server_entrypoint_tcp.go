package server

import (
	"context"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"sync"
	"time"

	proxyprotocol "github.com/c0va23/go-proxyprotocol"
	"github.com/containous/traefik/v2/pkg/config/static"
	"github.com/containous/traefik/v2/pkg/ip"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/middlewares/forwardedheaders"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/containous/traefik/v2/pkg/tcp"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var httpServerLogger = stdlog.New(log.WithoutContext().WriterLevel(logrus.DebugLevel), "", 0)

type httpForwarder struct {
	net.Listener
	connChan chan net.Conn
}

func newHTTPForwarder(ln net.Listener) *httpForwarder {
	return &httpForwarder{
		Listener: ln,
		connChan: make(chan net.Conn),
	}
}

// ServeTCP uses the connection to serve it later in "Accept"
func (h *httpForwarder) ServeTCP(conn tcp.WriteCloser) {
	h.connChan <- conn
}

// Accept retrieves a served connection in ServeTCP
func (h *httpForwarder) Accept() (net.Conn, error) {
	conn := <-h.connChan
	return conn, nil
}

// TCPEntryPoints holds a map of TCPEntryPoint (the entrypoint names being the keys)
type TCPEntryPoints map[string]*TCPEntryPoint

// NewTCPEntryPoints creates a new TCPEntryPoints.
func NewTCPEntryPoints(staticConfiguration static.Configuration) (TCPEntryPoints, error) {
	serverEntryPointsTCP := make(TCPEntryPoints)
	for entryPointName, config := range staticConfiguration.EntryPoints {
		ctx := log.With(context.Background(), log.Str(log.EntryPointName, entryPointName))

		var err error
		serverEntryPointsTCP[entryPointName], err = NewTCPEntryPoint(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("error while building entryPoint %s: %v", entryPointName, err)
		}
	}
	return serverEntryPointsTCP, nil
}

// Start the server entry points.
func (eps TCPEntryPoints) Start() {
	for entryPointName, serverEntryPoint := range eps {
		ctx := log.With(context.Background(), log.Str(log.EntryPointName, entryPointName))
		go serverEntryPoint.StartTCP(ctx)
	}
}

// Stop the server entry points.
func (eps TCPEntryPoints) Stop() {
	var wg sync.WaitGroup

	for epn, ep := range eps {
		wg.Add(1)

		go func(entryPointName string, entryPoint *TCPEntryPoint) {
			defer wg.Done()

			ctx := log.With(context.Background(), log.Str(log.EntryPointName, entryPointName))
			entryPoint.Shutdown(ctx)

			log.FromContext(ctx).Debugf("Entry point %s closed", entryPointName)
		}(epn, ep)
	}

	wg.Wait()
}

// Switch the TCP routers.
func (eps TCPEntryPoints) Switch(routersTCP map[string]*tcp.Router) {
	for entryPointName, rt := range routersTCP {
		eps[entryPointName].SwitchRouter(rt)
	}
}

// TCPEntryPoint is the TCP server
type TCPEntryPoint struct {
	listener               net.Listener
	switcher               *tcp.HandlerSwitcher
	transportConfiguration *static.EntryPointsTransport
	tracker                *connectionTracker
	httpServer             *httpServer
	httpsServer            *httpServer
}

// NewTCPEntryPoint creates a new TCPEntryPoint
func NewTCPEntryPoint(ctx context.Context, configuration *static.EntryPoint) (*TCPEntryPoint, error) {
	tracker := newConnectionTracker()

	listener, err := buildListener(ctx, configuration)
	if err != nil {
		return nil, fmt.Errorf("error preparing server: %v", err)
	}

	router := &tcp.Router{}

	httpServer, err := createHTTPServer(ctx, listener, configuration, true)
	if err != nil {
		return nil, fmt.Errorf("error preparing httpServer: %v", err)
	}

	router.HTTPForwarder(httpServer.Forwarder)

	httpsServer, err := createHTTPServer(ctx, listener, configuration, false)
	if err != nil {
		return nil, fmt.Errorf("error preparing httpsServer: %v", err)
	}

	router.HTTPSForwarder(httpsServer.Forwarder)

	tcpSwitcher := &tcp.HandlerSwitcher{}
	tcpSwitcher.Switch(router)

	return &TCPEntryPoint{
		listener:               listener,
		switcher:               tcpSwitcher,
		transportConfiguration: configuration.Transport,
		tracker:                tracker,
		httpServer:             httpServer,
		httpsServer:            httpsServer,
	}, nil
}

// StartTCP starts the TCP server.
func (e *TCPEntryPoint) StartTCP(ctx context.Context) {
	logger := log.FromContext(ctx)
	logger.Debugf("Start TCP Server")

	for {
		conn, err := e.listener.Accept()
		if err != nil {
			logger.Error(err)
			return
		}

		writeCloser, err := writeCloser(conn)
		if err != nil {
			panic(err)
		}

		safe.Go(func() {
			e.switcher.ServeTCP(newTrackedConnection(writeCloser, e.tracker))
		})
	}
}

// Shutdown stops the TCP connections
func (e *TCPEntryPoint) Shutdown(ctx context.Context) {
	logger := log.FromContext(ctx)

	reqAcceptGraceTimeOut := time.Duration(e.transportConfiguration.LifeCycle.RequestAcceptGraceTimeout)
	if reqAcceptGraceTimeOut > 0 {
		logger.Infof("Waiting %s for incoming requests to cease", reqAcceptGraceTimeOut)
		time.Sleep(reqAcceptGraceTimeOut)
	}

	graceTimeOut := time.Duration(e.transportConfiguration.LifeCycle.GraceTimeOut)
	ctx, cancel := context.WithTimeout(ctx, graceTimeOut)
	logger.Debugf("Waiting %s seconds before killing connections.", graceTimeOut)

	var wg sync.WaitGroup
	if e.httpServer.Server != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := e.httpServer.Server.Shutdown(ctx); err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					logger.Debugf("Wait server shutdown is overdue to: %s", err)
					err = e.httpServer.Server.Close()
					if err != nil {
						logger.Error(err)
					}
				}
			}
		}()
	}

	if e.httpsServer.Server != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := e.httpsServer.Server.Shutdown(ctx); err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					logger.Debugf("Wait server shutdown is overdue to: %s", err)
					err = e.httpsServer.Server.Close()
					if err != nil {
						logger.Error(err)
					}
				}
			}
		}()
	}

	if e.tracker != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := e.tracker.Shutdown(ctx); err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					logger.Debugf("Wait hijack connection is overdue to: %s", err)
					e.tracker.Close()
				}
			}
		}()
	}

	wg.Wait()
	cancel()
}

// SwitchRouter switches the TCP router handler.
func (e *TCPEntryPoint) SwitchRouter(router *tcp.Router) {
	router.HTTPForwarder(e.httpServer.Forwarder)

	httpHandler := router.GetHTTPHandler()
	if httpHandler == nil {
		httpHandler = buildDefaultHTTPRouter()
	}

	e.httpServer.Switcher.UpdateHandler(httpHandler)

	router.HTTPSForwarder(e.httpsServer.Forwarder)

	httpsHandler := router.GetHTTPSHandler()
	if httpsHandler == nil {
		httpsHandler = buildDefaultHTTPRouter()
	}

	e.httpsServer.Switcher.UpdateHandler(httpsHandler)

	e.switcher.Switch(router)
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
	case *proxyprotocol.Conn:
		underlying, err := writeCloser(typedConn.Conn)
		if err != nil {
			return nil, err
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

	if err = tc.SetKeepAlive(true); err != nil {
		return nil, err
	}

	if err = tc.SetKeepAlivePeriod(3 * time.Minute); err != nil {
		return nil, err
	}

	return tc, nil
}

type proxyProtocolLogger struct {
	log.Logger
}

// Printf force log level to debug.
func (p proxyProtocolLogger) Printf(format string, v ...interface{}) {
	p.Debugf(format, v...)
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

	return proxyprotocol.NewDefaultListener(listener).
		WithSourceChecker(sourceCheck).
		WithLogger(proxyProtocolLogger{Logger: log.FromContext(ctx)}), nil
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

func newConnectionTracker() *connectionTracker {
	return &connectionTracker{
		conns: make(map[net.Conn]struct{}),
	}
}

type connectionTracker struct {
	conns map[net.Conn]struct{}
	lock  sync.RWMutex
}

// AddConnection add a connection in the tracked connections list
func (c *connectionTracker) AddConnection(conn net.Conn) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.conns[conn] = struct{}{}
}

// RemoveConnection remove a connection from the tracked connections list
func (c *connectionTracker) RemoveConnection(conn net.Conn) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.conns, conn)
}

func (c *connectionTracker) isEmpty() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.conns) == 0
}

// Shutdown wait for the connection closing
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

// Close close all the connections in the tracked connections list
func (c *connectionTracker) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for conn := range c.conns {
		if err := conn.Close(); err != nil {
			log.WithoutContext().Errorf("Error while closing connection: %v", err)
		}
		delete(c.conns, conn)
	}
}

type stoppableServer interface {
	Shutdown(context.Context) error
	Close() error
	Serve(listener net.Listener) error
}

type httpServer struct {
	Server    stoppableServer
	Forwarder *httpForwarder
	Switcher  *middlewares.HTTPHandlerSwitcher
}

func createHTTPServer(ctx context.Context, ln net.Listener, configuration *static.EntryPoint, withH2c bool) (*httpServer, error) {
	httpSwitcher := middlewares.NewHandlerSwitcher(buildDefaultHTTPRouter())

	var handler http.Handler
	var err error
	handler, err = forwardedheaders.NewXForwarded(
		configuration.ForwardedHeaders.Insecure,
		configuration.ForwardedHeaders.TrustedIPs,
		httpSwitcher)
	if err != nil {
		return nil, err
	}

	if withH2c {
		handler = h2c.NewHandler(handler, &http2.Server{})
	}

	serverHTTP := &http.Server{
		Handler:  handler,
		ErrorLog: httpServerLogger,
	}

	listener := newHTTPForwarder(ln)
	go func() {
		err := serverHTTP.Serve(listener)
		if err != nil {
			log.FromContext(ctx).Errorf("Error while starting server: %v", err)
		}
	}()
	return &httpServer{
		Server:    serverHTTP,
		Forwarder: listener,
		Switcher:  httpSwitcher,
	}, nil
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
