package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	webtransport "github.com/quic-go/webtransport-go"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/static"
	tcpmuxer "github.com/traefik/traefik/v3/pkg/muxer/tcp"
	proxyhttputil "github.com/traefik/traefik/v3/pkg/proxy/httputil"
	tcprouter "github.com/traefik/traefik/v3/pkg/server/router/tcp"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

type http3server struct {
	*http3.Server

	// wtServer wraps the http3.Server with WebTransport session management.
	// Handlers call wtServer.Upgrade to accept incoming WebTransport sessions.
	wtServer *webtransport.Server

	http3conn net.PacketConn

	lock   sync.RWMutex
	getter func(data tcpmuxer.ConnData) (*tls.Config, string, error)
}

func newHTTP3Server(ctx context.Context, name string, config *static.EntryPoint, httpsServer *httpServer) (*http3server, error) {
	var conn net.PacketConn
	var err error

	if config.HTTP3 == nil {
		return nil, nil
	}

	if config.HTTP3.AdvertisedPort < 0 {
		return nil, errors.New("advertised port must be greater than or equal to zero")
	}

	// if we have predefined connections from socket activation
	if socketActivation.isEnabled() {
		conn, err = socketActivation.getConn(name)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("name", name).Msg("Unable to use socket activation for entrypoint")
		}
	}

	if conn == nil {
		listenConfig := newListenConfig(config)
		conn, err = listenConfig.ListenPacket(ctx, "udp", config.GetAddress())
		if err != nil {
			return nil, fmt.Errorf("starting listener: %w", err)
		}
	}

	h3 := &http3server{
		http3conn: conn,
		getter: func(data tcpmuxer.ConnData) (*tls.Config, string, error) {
			return nil, "", errors.New("no TLS config")
		},
	}

	h3.Server = &http3.Server{
		Addr:    config.GetAddress(),
		Port:    config.HTTP3.AdvertisedPort,
		Handler: httpsServer.Server.(*http.Server).Handler,
		// TLSConfig uses GetConfigForClient for dynamic per-SNI TLS configuration.
		// It is wrapped by http3.ConfigureTLSConfig at Start() time so that QUIC
		// connections negotiate the "h3" ALPN and any config returned by
		// GetConfigForClient also carries that ALPN.
		TLSConfig: &tls.Config{GetConfigForClient: h3.getTLSConfigForClient},
		QUICConfig: &quic.Config{
			Allow0RTT: false,
		},
		ConnContext: func(ctx context.Context, c *quic.Conn) context.Context {
			tlsOptionsName, err := h3.getTLSOptionsName(c)
			if err != nil {
				log.Error().Msgf("Error getting TLS options name for client: %v", err)
				return ctx
			}
			return tcp.AddTLSOptionsNameInContext(ctx, tlsOptionsName)
		},
	}

	// webtransport.Server wraps the http3.Server and manages WebTransport sessions.
	// CheckOrigin is set to allow all origins because Traefik's routing rules
	// already govern access; CORS-style origin checking is redundant here.
	// ConfigureHTTP3Server (called lazily by wtServer.Serve) sets EnableDatagrams
	// and injects the QUIC connection into each request's context — both
	// prerequisites for Upgrade() to work.
	h3.wtServer = &webtransport.Server{
		H3:          h3.Server,
		CheckOrigin: func(_ *http.Request) bool { return true },
	}

	previousHandler := httpsServer.Server.(*http.Server).Handler

	wrappedHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := h3.Server.SetQUICHeaders(rw.Header()); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Failed to set HTTP3 headers")
		}

		// Make the WebTransport server available to proxy handlers so they can
		// call Upgrade without importing this package (avoiding an import cycle).
		req = req.WithContext(proxyhttputil.SetWebTransportServer(req.Context(), h3.wtServer))

		// webtransport.Server.Upgrade() requires the raw http3.ResponseWriter
		// (which implements http3.Settingser and http3.HTTPStreamer). The middleware
		// chain will wrap this writer, hiding those interfaces. Save the unwrapped
		// writer in context NOW, before any middleware can mask it.
		if proxyhttputil.IsWebTransportRequest(req) {
			req = req.WithContext(proxyhttputil.SetHTTP3ResponseWriter(req.Context(), rw))
		}

		previousHandler.ServeHTTP(rw, req)
	})

	// Both the TCP (HTTPS) and UDP (HTTP/3) paths must go through wrappedHandler
	// so that SetQUICHeaders and SetWebTransportServer are applied regardless of
	// which transport the client uses.
	httpsServer.Server.(*http.Server).Handler = wrappedHandler
	h3.Server.Handler = wrappedHandler

	return h3, nil
}

func (e *http3server) Start() error {
	// webtransport.Server.serve() creates its own QUIC listener directly via
	// quic.ListenEarly, bypassing http3.Server.setupListenerForConn. This means:
	//
	//  1. The TLS config is not pre-wrapped with http3.ConfigureTLSConfig, so the
	//     "h3" ALPN would not be advertised → fix by wrapping before calling Serve.
	//
	//  2. http3.Server.addListener is never called, so SetQUICHeaders cannot
	//     generate an Alt-Svc header (it requires len(listeners) > 0) → fix by
	//     registering a dummy listener with the http3.Server.
	//
	// Pre-configure TLS so the QUIC listener advertises "h3" ALPN and any config
	// returned per-connection by GetConfigForClient also carries that ALPN.
	e.Server.TLSConfig = http3.ConfigureTLSConfig(e.Server.TLSConfig)

	// Register a dummy listener with the http3.Server so that SetQUICHeaders can
	// generate a valid Alt-Svc header.  The listener never accepts connections; it
	// just provides an Addr and blocks until the http3.Server is closed (which
	// happens inside wtServer.Close → h3.Server.Close → close all listeners).
	altSvcOnly := newAltSvcOnlyListener(e.http3conn.LocalAddr())
	go func() {
		if err := e.Server.ServeListener(altSvcOnly); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("HTTP/3 alt-svc listener error")
		}
	}()

	// Serve QUIC connections via the WebTransport server so it can intercept each
	// connection and register the per-connection session manager that Upgrade()
	// needs to look up later.
	return e.wtServer.Serve(e.http3conn)
}

func (e *http3server) Switch(rt *tcprouter.Router) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.getter = rt.HTTP3TLSConfigMatcherFunc()
}

func (e *http3server) Shutdown(_ context.Context) error {
	// TODO: use e.Server.CloseGracefully() when available.
	// Closing via the WebTransport server terminates in-flight WebTransport
	// sessions cleanly. wtServer.Close also calls h3.Server.Close which in turn
	// closes the registered altSvcOnlyListener, unblocking its goroutine.
	return e.wtServer.Close()
}

// altSvcOnlyListener is a dummy http3.QUICListener that never accepts connections.
// It is registered with the http3.Server via ServeListener so that
// SetQUICHeaders can generate a valid Alt-Svc header (which requires at least
// one listener to be registered).  Actual QUIC serving is done by the
// webtransport.Server.
type altSvcOnlyListener struct {
	addr net.Addr
	once sync.Once
	done chan struct{}
}

func newAltSvcOnlyListener(addr net.Addr) *altSvcOnlyListener {
	return &altSvcOnlyListener{
		addr: addr,
		done: make(chan struct{}),
	}
}

func (l *altSvcOnlyListener) Accept(ctx context.Context) (*quic.Conn, error) {
	select {
	case <-l.done:
		return nil, net.ErrClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (l *altSvcOnlyListener) Addr() net.Addr { return l.addr }

func (l *altSvcOnlyListener) Close() error {
	l.once.Do(func() { close(l.done) })
	return nil
}

func (e *http3server) getTLSConfigForClient(info *tls.ClientHelloInfo) (*tls.Config, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	connData, err := tcpmuxer.NewConnData(info.ServerName, info.Conn.RemoteAddr(), info.SupportedProtos)
	if err != nil {
		return nil, fmt.Errorf("creating ConnData from client hello: %w", err)
	}

	conf, _, err := e.getter(connData)
	return conf, err
}

func (e *http3server) getTLSOptionsName(c *quic.Conn) (string, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	connData, err := tcpmuxer.NewConnData(c.ConnectionState().TLS.ServerName, c.RemoteAddr(), []string{c.ConnectionState().TLS.NegotiatedProtocol})
	if err != nil {
		return "", fmt.Errorf("creating ConnData from quic Conn: %w", err)
	}

	_, name, err := e.getter(connData)
	return name, err
}
