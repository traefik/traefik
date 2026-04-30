package tcp

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/rs/zerolog/log"
	tcpmuxer "github.com/traefik/traefik/v3/pkg/muxer/tcp"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

// errClientHelloRead is used as a sentinel error to break the TLS handshake once we have read the ClientHello.
var errClientHelloRead = errors.New("client hello successfully read")

// Router is a TCP router.
type Router struct {
	acmeTLSPassthrough bool

	// Contains TCP routes.
	muxerTCP tcpmuxer.Muxer
	// Contains TCP TLS routes.
	muxerTCPTLS tcpmuxer.Muxer
	// Contains HTTPS routes.
	muxerHTTPS tcpmuxer.Muxer

	// Forwarder handlers.
	// httpForwarder handles all HTTP requests.
	httpForwarder tcp.Handler
	// httpsForwarder handles (indirectly through muxerHTTPS, or directly) all HTTPS requests.
	httpsForwarder tcp.Handler

	// Neither is used directly, but they are held here, and recreated on config reload,
	// so that they can be passed to the Switcher at the end of the config reload phase.
	httpHandler  http.Handler
	httpsHandler http.Handler

	// TLS configs.
	httpsTLSConfig *tls.Config // default TLS config
	// hostHTTPTLSConfig contains TLS configs keyed by SNI.
	// A nil config is the hint to set up a brokenTLSRouter.
	hostHTTPTLSConfig map[string]*tls.Config // TLS configs keyed by SNI
}

// NewRouter returns a new TCP router.
func NewRouter(providersPrecedence []string) (*Router, error) {
	muxTCP, err := tcpmuxer.NewMuxer(providersPrecedence)
	if err != nil {
		return nil, err
	}

	muxTCPTLS, err := tcpmuxer.NewMuxer(providersPrecedence)
	if err != nil {
		return nil, err
	}

	muxHTTPS, err := tcpmuxer.NewMuxer(providersPrecedence)
	if err != nil {
		return nil, err
	}

	return &Router{
		muxerTCP:    *muxTCP,
		muxerTCPTLS: *muxTCPTLS,
		muxerHTTPS:  *muxHTTPS,
	}, nil
}

// GetTLSGetClientInfo is called after a ClientHello is received from a client.
func (r *Router) GetTLSGetClientInfo() func(info *tls.ClientHelloInfo) (*tls.Config, error) {
	return func(info *tls.ClientHelloInfo) (*tls.Config, error) {
		if tlsConfig, ok := r.hostHTTPTLSConfig[info.ServerName]; ok {
			return tlsConfig, nil
		}

		return r.httpsTLSConfig, nil
	}
}

// ServeTCP forwards the connection to the right TCP/HTTP handler.
func (r *Router) ServeTCP(conn tcp.WriteCloser) {
	// Handling Non-TLS TCP connection early if there is neither HTTP(S) nor TLS routers on the entryPoint,
	// and if there is at least one non-TLS TCP router.
	// In the case of a non-TLS TCP client (that does not "send" first),
	// we would block forever on clientHelloInfo,
	// which is why we want to detect and handle that case first and foremost.
	if r.muxerTCP.HasRoutes() && !r.muxerTCPTLS.HasRoutes() && !r.muxerHTTPS.HasRoutes() {
		connData, err := tcpmuxer.NewConnData("", conn, nil)
		if err != nil {
			log.Error().Err(err).Msg("Error while reading TCP connection data")
			conn.Close()
			return
		}

		handler, _ := r.muxerTCP.Match(connData)
		// If there is a handler matching the connection metadata,
		// we let it handle the connection.
		if handler != nil {
			// Remove read/write deadline and delegate this to underlying TCP server.
			if err := conn.SetDeadline(time.Time{}); err != nil {
				log.Error().Err(err).Msg("Error while setting deadline")
			}

			handler.ServeTCP(conn)
			return
		}
		// Otherwise, we keep going because:
		// 1) we could be in the case where we have HTTP routers.
		// 2) if it is an HTTPS request, even though we do not have any TLS routers,
		// we still need to reply with a 404.
	}

	// TODO -- Check if ProxyProtocol changes the first bytes of the request
	pConn := newPeekConn(conn)

	postgres, err := isPostgres(pConn)
	if err != nil {
		var opErr *net.OpError
		if !errors.Is(err, io.EOF) && (!errors.As(err, &opErr) || !opErr.Timeout()) {
			log.Debug().Err(err).Msg("Error while peeking first bytes")
		}
		_ = pConn.Close()
		return
	}

	if postgres {
		if err := r.servePostgres(pConn); err != nil {
			var opErr *net.OpError
			if !errors.Is(err, io.EOF) && (!errors.As(err, &opErr) || !opErr.Timeout()) {
				log.Debug().Err(err).Msg("Error while serving Postgres connection")
			}
		}
		_ = pConn.Close()
		return
	}

	hello, err := clientHelloInfo(pConn)
	if err != nil {
		var opErr *net.OpError
		if !errors.Is(err, io.EOF) && (!errors.As(err, &opErr) || !opErr.Timeout()) {
			log.Debug().Err(err).Msg("Error while reading client hello")
		}
		_ = pConn.Close()
		return
	}

	// The deadline was set to avoid blocking on the initial read of the ClientHello,
	// but now that we have it, we can remove it,
	// and delegate this to underlying TCP server (for now only handled by HTTP Server).
	if err := pConn.SetDeadline(time.Time{}); err != nil {
		log.Error().Err(err).Msg("Error while setting deadline")
	}

	connData, err := tcpmuxer.NewConnData(hello.serverName, pConn, hello.protos)
	if err != nil {
		log.Error().Err(err).Msg("Error while reading TCP connection data")
		_ = pConn.Close()
		return
	}

	if !hello.isTLS {
		handler, _ := r.muxerTCP.Match(connData)
		switch {
		case handler != nil:
			handler.ServeTCP(pConn)
		case r.httpForwarder != nil:
			r.httpForwarder.ServeTCP(pConn)
		default:
			_ = pConn.Close()
		}
		return
	}

	// Handling ACME-TLS/1 challenges.
	if !r.acmeTLSPassthrough && slices.Contains(hello.protos, tlsalpn01.ACMETLS1Protocol) {
		r.acmeTLSALPNHandler().ServeTCP(pConn)
		return
	}

	// For real, the handler eventually used for HTTPS is (almost) always the same:
	// it is the httpsForwarder that is used for all HTTPS connections that match
	// (which is also incidentally the same used in the last block below for 404s).
	// The added value from doing Match is to find and use the specific TLS config
	// (wrapped inside the returned handler) requested for the given HostSNI.
	handlerHTTPS, catchAllHTTPS := r.muxerHTTPS.Match(connData)
	if handlerHTTPS != nil && !catchAllHTTPS {
		// In order not to depart from the behavior in 2.6,
		// we only allow an HTTPS router to take precedence over a TCP-TLS router if it is _not_ an HostSNI(*) router
		// (so basically any router that has a specific HostSNI based rule).
		handlerHTTPS.ServeTCP(pConn)
		return
	}

	// Contains also TCP TLS passthrough routes.
	handlerTCPTLS, catchAllTCPTLS := r.muxerTCPTLS.Match(connData)
	if handlerTCPTLS != nil && !catchAllTCPTLS {
		handlerTCPTLS.ServeTCP(pConn)
		return
	}

	// Fallback on HTTPS catchAll.
	// We end up here for e.g. an HTTPS router that only has a PathPrefix rule,
	// which under the scenes is counted as an HostSNI(*) rule.
	if handlerHTTPS != nil {
		handlerHTTPS.ServeTCP(pConn)
		return
	}

	// Fallback on TCP TLS catchAll.
	if handlerTCPTLS != nil {
		handlerTCPTLS.ServeTCP(pConn)
		return
	}

	// To handle 404s for HTTPS.
	if r.httpsForwarder != nil {
		r.httpsForwarder.ServeTCP(pConn)
		return
	}

	_ = pConn.Close()
}

// AddTCPRoute defines a handler for the given rule.
func (r *Router) AddTCPRoute(rule string, priority int, providerName string, target tcp.Handler) error {
	return r.muxerTCP.AddRoute(rule, "", priority, providerName, target)
}

// AddHTTPTLSConfig defines a handler for a given sniHost and sets the matching tlsConfig.
func (r *Router) AddHTTPTLSConfig(sniHost string, config *tls.Config) {
	if r.hostHTTPTLSConfig == nil {
		r.hostHTTPTLSConfig = map[string]*tls.Config{}
	}

	r.hostHTTPTLSConfig[sniHost] = config
}

// GetHTTPHandler gets the attached http handler.
func (r *Router) GetHTTPHandler() http.Handler {
	return r.httpHandler
}

// GetHTTPSHandler gets the attached https handler.
func (r *Router) GetHTTPSHandler() http.Handler {
	return r.httpsHandler
}

// SetHTTPForwarder sets the tcp handler that will forward the connections to an http handler.
func (r *Router) SetHTTPForwarder(handler tcp.Handler) {
	r.httpForwarder = handler
}

// SetHTTPSForwarder sets the tcp handler that will forward the TLS connections to an HTTP handler.
// It also sets up each TLS handler (with its TLS config) for each Host(SNI) rule we previously kept track of.
// It sets up a special handler that closes the connection if a TLS config is nil.
func (r *Router) SetHTTPSForwarder(handler tcp.Handler) {
	for sniHost, tlsConf := range r.hostHTTPTLSConfig {
		var tcpHandler tcp.Handler
		if tlsConf == nil {
			tcpHandler = &brokenTLSRouter{}
		} else {
			tcpHandler = &tcp.TLSHandler{
				Next:   handler,
				Config: tlsConf,
			}
		}

		rule := fmt.Sprintf(`HostSNI(%q)`, sniHost)
		// As the hostHTTPTLSConfig contains only one TLS config per SNI,
		// there is no conflict thus the provider name can be passed as empty as no tie-break is needed.
		if err := r.muxerHTTPS.AddRoute(rule, "", tcpmuxer.GetRulePriority(rule), "", tcpHandler); err != nil {
			log.Error().Err(err).Msg("Error while adding route for host")
		}
	}

	if r.httpsTLSConfig == nil {
		r.httpsForwarder = &brokenTLSRouter{}
		return
	}

	r.httpsForwarder = &tcp.TLSHandler{
		Next:   handler,
		Config: r.httpsTLSConfig,
	}
}

// SetHTTPHandler attaches http handlers on the router.
func (r *Router) SetHTTPHandler(handler http.Handler) {
	r.httpHandler = handler
}

// SetHTTPSHandler attaches https handlers on the router.
func (r *Router) SetHTTPSHandler(handler http.Handler, config *tls.Config) {
	r.httpsHandler = handler
	r.httpsTLSConfig = config
}

func (r *Router) EnableACMETLSPassthrough() {
	r.acmeTLSPassthrough = true
}

// acmeTLSALPNHandler returns a special handler to solve ACME-TLS/1 challenges.
func (r *Router) acmeTLSALPNHandler() tcp.Handler {
	if r.httpsTLSConfig == nil {
		return &brokenTLSRouter{}
	}

	return tcp.HandlerFunc(func(conn tcp.WriteCloser) {
		tlsConn := tls.Server(conn, r.httpsTLSConfig)
		defer tlsConn.Close()

		// This avoids stale connections when validating the ACME challenge,
		// as we expect a validation request to complete in a short period of time.
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := tlsConn.HandshakeContext(ctx); err != nil {
			log.Debug().Err(err).Msg("Error during ACME-TLS/1 handshake")
		}
	})
}

// brokenTLSRouter is associated to a Host(SNI) rule for which we know the TLS conf is broken.
// It is used to make sure any attempt to connect to that hostname is closed,
// since we cannot proceed with the intended TLS conf.
type brokenTLSRouter struct{}

// ServeTCP instantly closes the connection.
func (t *brokenTLSRouter) ServeTCP(conn tcp.WriteCloser) {
	_ = conn.Close()
}

// peekConn wraps a tcp.WriteCloser with a bufio.Reader for Peek operations
// and a peeked buffer that accumulates bytes consumed during protocol detection
// so they can be replayed on subsequent Read calls.
type peekConn struct {
	tcp.WriteCloser

	peeked []byte
	reader *bufio.Reader
}

func newPeekConn(conn tcp.WriteCloser) *peekConn {
	return &peekConn{
		WriteCloser: conn,
		reader:      bufio.NewReader(conn),
	}
}

// Peek allows peeking into the connection without consuming bytes, by using the bufio.Reader's Peek method.
func (c *peekConn) Peek(n int) ([]byte, error) {
	return c.reader.Peek(n)
}

// PeekRead reads from the connection and accumulates the read bytes into the peeked buffer, allowing them to be replayed later.
// Note that PeekRead advances the reader, thus the next call to Peek will not return the same result as the previous one.
func (c *peekConn) PeekRead(p []byte) (int, error) {
	n, err := c.reader.Read(p)
	if n > 0 {
		c.peeked = append(c.peeked, p[:n]...)
	}
	return n, err
}

// Read drains accumulated peeked bytes first, then reads from the bufio reader.
func (c *peekConn) Read(p []byte) (int, error) {
	if len(c.peeked) > 0 {
		n := copy(p, c.peeked)
		c.peeked = c.peeked[n:]
		if len(c.peeked) == 0 {
			c.peeked = nil
		}
		return n, nil
	}

	return c.reader.Read(p)
}

type clientHello struct {
	serverName string   // SNI server name
	protos     []string // ALPN protocols list
	isTLS      bool     // whether we are a TLS handshake
}

// clientHelloInfo returns various data from the clientHello handshake,
// without consuming any bytes from conn.
// It returns an error if it can't peek the first byte from the connection.
func clientHelloInfo(conn *peekConn) (*clientHello, error) {
	hdr, err := conn.Peek(1)
	if err != nil {
		return nil, fmt.Errorf("peeking first byte: %w", err)
	}

	// No valid TLS record has a type of 0x80, however SSLv2 handshakes start with an uint16 length
	// where the MSB is set and the first record is always < 256 bytes long.
	// Therefore, typ == 0x80 strongly suggests an SSLv2 client.
	const recordTypeSSLv2 = 0x80
	const recordTypeHandshake = 0x16
	if hdr[0] != recordTypeHandshake {
		if hdr[0] == recordTypeSSLv2 {
			// we consider SSLv2 as TLS, and it will be refused by real TLS handshake.
			return &clientHello{
				isTLS: true,
			}, nil
		}
		return &clientHello{}, nil // Not TLS.
	}

	var (
		sni    string
		protos []string
	)
	server := tls.Server(readOnlyConn{conn: conn}, &tls.Config{
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			sni = hello.ServerName
			protos = hello.SupportedProtos
			// This error prevents unnecessary additional steps in the TLS ClientHello message processing.
			return nil, errClientHelloRead
		},
	})

	if handshakeErr := server.Handshake(); !errors.Is(handshakeErr, errClientHelloRead) {
		return nil, fmt.Errorf("reading client hello: %w", handshakeErr)
	}

	return &clientHello{
		serverName: sni,
		isTLS:      true,
		protos:     protos,
	}, nil
}

// readOnlyConn is a net.Conn that reads from conn, fails on Writes and crashes otherwise.
type readOnlyConn struct {
	net.Conn // nil; crash on any unexpected use

	conn *peekConn
}

// Read reads from the conn using PeekRead to keep the read bytes.
func (c readOnlyConn) Read(p []byte) (int, error) {
	return c.conn.PeekRead(p)
}

// Write crashes all the time.
func (readOnlyConn) Write(_ []byte) (int, error) { return 0, io.EOF }
