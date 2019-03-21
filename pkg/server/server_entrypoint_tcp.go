package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/armon/go-proxyproto"
	"github.com/containous/traefik/pkg/config/static"
	"github.com/containous/traefik/pkg/h2c"
	"github.com/containous/traefik/pkg/ip"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/middlewares"
	"github.com/containous/traefik/pkg/middlewares/forwardedheaders"
	"github.com/containous/traefik/pkg/safe"
	"github.com/containous/traefik/pkg/tcp"
)

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
func (h *httpForwarder) ServeTCP(conn net.Conn) {
	h.connChan <- conn
}

// Accept retrieves a served connection in ServeTCP
func (h *httpForwarder) Accept() (net.Conn, error) {
	conn := <-h.connChan
	return conn, nil
}

// TCPEntryPoints holds a map of TCPEntryPoint (the entrypoint names being the keys)
type TCPEntryPoints map[string]*TCPEntryPoint

// TCPEntryPoint is the TCP server
type TCPEntryPoint struct {
	listener               net.Listener
	switcher               *tcp.HandlerSwitcher
	RouteAppenderFactory   RouteAppenderFactory
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

	httpServer, err := createHTTPServer(listener, configuration, true)
	if err != nil {
		return nil, fmt.Errorf("error preparing httpServer: %v", err)
	}

	router.HTTPForwarder(httpServer.Forwarder)

	httpsServer, err := createHTTPServer(listener, configuration, false)
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

func (e *TCPEntryPoint) startTCP(ctx context.Context) {
	log.FromContext(ctx).Debugf("Start TCP Server")

	for {
		conn, err := e.listener.Accept()
		if err != nil {
			log.Error(err)
			return
		}

		safe.Go(func() {
			e.switcher.ServeTCP(newTrackedConnection(conn, e.tracker))
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

func (e *TCPEntryPoint) switchRouter(router *tcp.Router) {
	router.HTTPForwarder(e.httpServer.Forwarder)
	router.HTTPSForwarder(e.httpsServer.Forwarder)

	httpHandler := router.GetHTTPHandler()
	httpsHandler := router.GetHTTPSHandler()
	if httpHandler == nil {
		httpHandler = buildDefaultHTTPRouter()
	}
	if httpsHandler == nil {
		httpsHandler = buildDefaultHTTPRouter()
	}

	e.httpServer.Switcher.UpdateHandler(httpHandler)
	e.httpsServer.Switcher.UpdateHandler(httpsHandler)
	e.switcher.Switch(router)
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

func createHTTPServer(ln net.Listener, configuration *static.EntryPoint, withH2c bool) (*httpServer, error) {
	httpSwitcher := middlewares.NewHandlerSwitcher(buildDefaultHTTPRouter())
	handler, err := forwardedheaders.NewXForwarded(
		configuration.ForwardedHeaders.Insecure,
		configuration.ForwardedHeaders.TrustedIPs,
		httpSwitcher)
	if err != nil {
		return nil, err
	}

	var serverHTTP stoppableServer

	if withH2c {
		serverHTTP = &h2c.Server{
			Server: &http.Server{
				Handler: handler,
			},
		}
	} else {
		serverHTTP = &http.Server{
			Handler: handler,
		}
	}

	listener := newHTTPForwarder(ln)
	go func() {
		err := serverHTTP.Serve(listener)
		if err != nil {
			log.Errorf("Error while starting server: %v", err)
		}
	}()
	return &httpServer{
		Server:    serverHTTP,
		Forwarder: listener,
		Switcher:  httpSwitcher,
	}, nil
}

func newTrackedConnection(conn net.Conn, tracker *connectionTracker) *trackedConnection {
	tracker.AddConnection(conn)
	return &trackedConnection{
		Conn:    conn,
		tracker: tracker,
	}
}

type trackedConnection struct {
	tracker *connectionTracker
	net.Conn
}

func (t *trackedConnection) Close() error {
	t.tracker.RemoveConnection(t.Conn)
	return t.Conn.Close()
}
