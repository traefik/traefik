package tcp

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/traefik/traefik/v2/pkg/log"
	tcpmuxer "github.com/traefik/traefik/v2/pkg/muxer/tcp"
	"github.com/traefik/traefik/v2/pkg/tcp"
)

const defaultBufSize = 4096

// Router is a TCP router.
type Router struct {
	// Contains TCP routes.
	muxerTCP tcpmuxer.Muxer
	// Contains TCP TLS routes.
	muxerTCPTLS tcpmuxer.Muxer
	// Contains HTTPS routes.
	muxerHTTPS tcpmuxer.Muxer

	// Forwarder handlers.
	// Handles all HTTP requests.
	httpForwarder tcp.Handler
	// Handles (indirectly through muxerHTTPS, or directly) all HTTPS requests.
	httpsForwarder tcp.Handler

	// Neither is used directly, but they are held here, and recreated on config
	// reload, so that they can be passed to the Switcher at the end of the config
	// reload phase.
	httpHandler  http.Handler
	httpsHandler http.Handler

	// TLS configs.
	httpsTLSConfig    *tls.Config            // default TLS config
	hostHTTPTLSConfig map[string]*tls.Config // TLS configs keyed by SNI
}

// NewRouter returns a new TCP router.
func NewRouter() (*Router, error) {
	muxTCP, err := tcpmuxer.NewMuxer()
	if err != nil {
		return nil, err
	}

	muxTCPTLS, err := tcpmuxer.NewMuxer()
	if err != nil {
		return nil, err
	}

	muxHTTPS, err := tcpmuxer.NewMuxer()
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
	// Handling Non-TLS TCP connection early if there is neither HTTP(S) nor TLS
	// routers on the entryPoint, and if there is at least one non-TLS TCP router.
	// In the case of a non-TLS TCP client (that does not "send" first), we would
	// block forever on clientHelloServerName, which is why we want to detect and
	// handle that case first and foremost.
	if r.muxerTCP.HasRoutes() && !r.muxerTCPTLS.HasRoutes() && !r.muxerHTTPS.HasRoutes() {
		connData, err := tcpmuxer.NewConnData("", conn)
		if err != nil {
			log.WithoutContext().Errorf("Error while reading TCP connection data: %v", err)
			conn.Close()
			return
		}

		handler := r.muxerTCP.Match(connData)
		// If there is a handler matching the connection metadata,
		// we let it handle the connection.
		if handler != nil {
			handler.ServeTCP(conn)
			return
		}
		// Otherwise, we keep going because:
		// 1) we could be in the case where we have HTTP routers.
		// 2) if it is an HTTPS request, even though we do not have any TLS routers,
		// we still need to reply with a 404.
	}

	// FIXME -- Check if ProxyProtocol changes the first bytes of the request
	br := bufio.NewReader(conn)
	serverName, tls, peeked, err := clientHelloServerName(br)
	if err != nil {
		conn.Close()
		return
	}

	// Remove read/write deadline and delegate this to underlying tcp server (for now only handled by HTTP Server)
	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		log.WithoutContext().Errorf("Error while setting read deadline: %v", err)
	}

	err = conn.SetWriteDeadline(time.Time{})
	if err != nil {
		log.WithoutContext().Errorf("Error while setting write deadline: %v", err)
	}

	connData, err := tcpmuxer.NewConnData(serverName, conn)
	if err != nil {
		log.WithoutContext().Errorf("Error while reading TCP connection data: %v", err)
		conn.Close()
		return
	}

	if !tls {
		handler := r.muxerTCP.Match(connData)
		switch {
		case handler != nil:
			handler.ServeTCP(r.GetConn(conn, peeked))
		case r.httpForwarder != nil:
			r.httpForwarder.ServeTCP(r.GetConn(conn, peeked))
		default:
			conn.Close()
		}
		return
	}

	handler := r.muxerTCPTLS.Match(connData)
	if handler != nil {
		handler.ServeTCP(r.GetConn(conn, peeked))
		return
	}

	// for real, the handler returned here is (almost) always the same:
	// it is the httpsForwarder that is used for all HTTPS connections that match
	// (which is also incidentally the same used in the last block below for 404s).
	// The added value from doing Match, is to find and use the specific TLS config
	// requested for the given HostSNI.
	handler = r.muxerHTTPS.Match(connData)
	if handler != nil {
		handler.ServeTCP(r.GetConn(conn, peeked))
		return
	}

	// needed to handle 404s for HTTPS, as well as all non-Host (e.g. PathPrefix) matches.
	if r.httpsForwarder != nil {
		r.httpsForwarder.ServeTCP(r.GetConn(conn, peeked))
		return
	}

	conn.Close()
}

// AddRoute defines a handler for the given rule.
func (r *Router) AddRoute(rule string, priority int, target tcp.Handler) error {
	return r.muxerTCP.AddRoute(rule, priority, target)
}

// AddRouteTLS defines a handler for a given rule and sets the matching tlsConfig.
func (r *Router) AddRouteTLS(rule string, priority int, target tcp.Handler, config *tls.Config) error {
	// TLS PassThrough
	if config == nil {
		return r.muxerTCPTLS.AddRoute(rule, priority, target)
	}

	return r.muxerTCPTLS.AddRoute(rule, priority, &tcp.TLSHandler{
		Next:   target,
		Config: config,
	})
}

// AddHTTPTLSConfig defines a handler for a given sniHost and sets the matching tlsConfig.
func (r *Router) AddHTTPTLSConfig(sniHost string, config *tls.Config) {
	if r.hostHTTPTLSConfig == nil {
		r.hostHTTPTLSConfig = map[string]*tls.Config{}
	}

	r.hostHTTPTLSConfig[sniHost] = config
}

// GetConn creates a connection proxy with a peeked string.
func (r *Router) GetConn(conn tcp.WriteCloser, peeked string) tcp.WriteCloser {
	// FIXME should it really be on Router ?
	conn = &Conn{
		Peeked:      []byte(peeked),
		WriteCloser: conn,
	}

	return conn
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

// SetHTTPSForwarder sets the tcp handler that will forward the TLS connections to an http handler.
func (r *Router) SetHTTPSForwarder(handler tcp.Handler) {
	for sniHost, tlsConf := range r.hostHTTPTLSConfig {
		// muxerHTTPS only contains single HostSNI rules (and no other kind of rules),
		// so there's no need for specifying a priority for them.
		err := r.muxerHTTPS.AddRoute("HostSNI(`"+sniHost+"`)", 0, &tcp.TLSHandler{
			Next:   handler,
			Config: tlsConf,
		})
		if err != nil {
			log.WithoutContext().Errorf("Error while adding route for host: %w", err)
		}
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

// Conn is a connection proxy that handles Peeked bytes.
type Conn struct {
	// Peeked are the bytes that have been read from Conn for the
	// purposes of route matching, but have not yet been consumed
	// by Read calls. It set to nil by Read when fully consumed.
	Peeked []byte

	// Conn is the underlying connection.
	// It can be type asserted against *net.TCPConn or other types
	// as needed. It should not be read from directly unless
	// Peeked is nil.
	tcp.WriteCloser
}

// Read reads bytes from the connection (using the buffer prior to actually reading).
func (c *Conn) Read(p []byte) (n int, err error) {
	if len(c.Peeked) > 0 {
		n = copy(p, c.Peeked)
		c.Peeked = c.Peeked[n:]
		if len(c.Peeked) == 0 {
			c.Peeked = nil
		}
		return n, nil
	}
	return c.WriteCloser.Read(p)
}

// clientHelloServerName returns the SNI server name inside the TLS ClientHello,
// without consuming any bytes from br.
// On any error, the empty string is returned.
func clientHelloServerName(br *bufio.Reader) (string, bool, string, error) {
	hdr, err := br.Peek(1)
	if err != nil {
		var opErr *net.OpError
		if !errors.Is(err, io.EOF) && (!errors.As(err, &opErr) || opErr.Timeout()) {
			log.WithoutContext().Errorf("Error while Peeking first byte: %s", err)
		}

		return "", false, "", err
	}

	// No valid TLS record has a type of 0x80, however SSLv2 handshakes
	// start with a uint16 length where the MSB is set and the first record
	// is always < 256 bytes long. Therefore typ == 0x80 strongly suggests
	// an SSLv2 client.
	const recordTypeSSLv2 = 0x80
	const recordTypeHandshake = 0x16
	if hdr[0] != recordTypeHandshake {
		if hdr[0] == recordTypeSSLv2 {
			// we consider SSLv2 as TLS and it will be refused by real TLS handshake.
			return "", true, getPeeked(br), nil
		}
		return "", false, getPeeked(br), nil // Not TLS.
	}

	const recordHeaderLen = 5
	hdr, err = br.Peek(recordHeaderLen)
	if err != nil {
		log.Errorf("Error while Peeking hello: %s", err)
		return "", false, getPeeked(br), nil
	}

	recLen := int(hdr[3])<<8 | int(hdr[4]) // ignoring version in hdr[1:3]

	if recordHeaderLen+recLen > defaultBufSize {
		br = bufio.NewReaderSize(br, recordHeaderLen+recLen)
	}

	helloBytes, err := br.Peek(recordHeaderLen + recLen)
	if err != nil {
		log.Errorf("Error while Hello: %s", err)
		return "", true, getPeeked(br), nil
	}

	sni := ""
	server := tls.Server(sniSniffConn{r: bytes.NewReader(helloBytes)}, &tls.Config{
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			sni = hello.ServerName
			return nil, nil
		},
	})
	_ = server.Handshake()

	return sni, true, getPeeked(br), nil
}

func getPeeked(br *bufio.Reader) string {
	peeked, err := br.Peek(br.Buffered())
	if err != nil {
		log.Errorf("Could not get anything: %s", err)
		return ""
	}
	return string(peeked)
}

// sniSniffConn is a net.Conn that reads from r, fails on Writes,
// and crashes otherwise.
type sniSniffConn struct {
	r        io.Reader
	net.Conn // nil; crash on any unexpected use
}

// Read reads from the underlying reader.
func (c sniSniffConn) Read(p []byte) (int, error) { return c.r.Read(p) }

// Write crashes all the time.
func (sniSniffConn) Write(p []byte) (int, error) { return 0, io.EOF }
