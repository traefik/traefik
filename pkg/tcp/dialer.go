package tcp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pires/go-proxyproto"
	"github.com/rs/zerolog/log"
	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

// ClientConn is the interface that provides information about the client connection.
type ClientConn interface {
	// LocalAddr returns the local network address, if known.
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address, if known.
	RemoteAddr() net.Addr
}

// Dialer is an interface to dial a network connection, with support for PROXY protocol and termination delay.
type Dialer interface {
	Dial(network, addr string, clientConn ClientConn) (c net.Conn, err error)
	TerminationDelay() time.Duration
}

type tcpDialer struct {
	dialer           *net.Dialer
	terminationDelay time.Duration
	proxyProtocol    *dynamic.ProxyProtocol
}

// TerminationDelay returns the termination delay duration.
func (d tcpDialer) TerminationDelay() time.Duration {
	return d.terminationDelay
}

// Dial dials a network connection and optionally sends a PROXY protocol header.
func (d tcpDialer) Dial(network, addr string, clientConn ClientConn) (net.Conn, error) {
	conn, err := d.dialer.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	if d.proxyProtocol != nil && clientConn != nil && d.proxyProtocol.Version > 0 && d.proxyProtocol.Version < 3 {
		header := proxyproto.HeaderProxyFromAddrs(byte(d.proxyProtocol.Version), clientConn.RemoteAddr(), clientConn.LocalAddr())
		if _, err := header.WriteTo(conn); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("writing PROXY Protocol header: %w", err)
		}
	}

	return conn, nil
}

type tcpTLSDialer struct {
	tcpDialer
	tlsConfig *tls.Config
}

// Dial dials a network connection with the wrapped tcpDialer and performs a TLS handshake.
func (d tcpTLSDialer) Dial(network, addr string, clientConn ClientConn) (net.Conn, error) {
	conn, err := d.tcpDialer.Dial(network, addr, clientConn)
	if err != nil {
		return nil, err
	}

	// Now perform TLS handshake on the connection
	tlsConn := tls.Client(conn, d.tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	return tlsConn, nil
}

// SpiffeX509Source allows retrieving a x509 SVID and bundle.
type SpiffeX509Source interface {
	x509svid.Source
	x509bundle.Source
}

// DialerManager handles dialer for the reverse proxy.
type DialerManager struct {
	serversTransportsMu sync.RWMutex
	serversTransports   map[string]*dynamic.TCPServersTransport
	spiffeX509Source    SpiffeX509Source
}

// NewDialerManager creates a new DialerManager.
func NewDialerManager(spiffeX509Source SpiffeX509Source) *DialerManager {
	return &DialerManager{
		serversTransports: make(map[string]*dynamic.TCPServersTransport),
		spiffeX509Source:  spiffeX509Source,
	}
}

// Update updates the TCP serversTransport configurations.
func (d *DialerManager) Update(configs map[string]*dynamic.TCPServersTransport) {
	d.serversTransportsMu.Lock()
	defer d.serversTransportsMu.Unlock()

	d.serversTransports = configs
}

// Build builds a dialer by name.
func (d *DialerManager) Build(config *dynamic.TCPServersLoadBalancer, isTLS bool) (Dialer, error) {
	name := "default@internal"
	if config.ServersTransport != "" {
		name = config.ServersTransport
	}

	var st *dynamic.TCPServersTransport
	d.serversTransportsMu.RLock()
	st, ok := d.serversTransports[name]
	d.serversTransportsMu.RUnlock()
	if !ok || st == nil {
		return nil, fmt.Errorf("no transport configuration found for %q", name)
	}

	// Handle TerminationDelay and ProxyProtocol deprecated options.
	var terminationDelay ptypes.Duration
	if config.TerminationDelay != nil {
		terminationDelay = ptypes.Duration(*config.TerminationDelay)
	}
	proxyProtocol := config.ProxyProtocol

	if config.ServersTransport != "" {
		terminationDelay = st.TerminationDelay
		proxyProtocol = st.ProxyProtocol
	}

	if proxyProtocol != nil && (proxyProtocol.Version < 1 || proxyProtocol.Version > 2) {
		return nil, fmt.Errorf("unknown proxyProtocol version: %d", proxyProtocol.Version)
	}

	var tlsConfig *tls.Config
	if st.TLS != nil {
		if st.TLS.Spiffe != nil {
			if d.spiffeX509Source == nil {
				return nil, errors.New("SPIFFE is enabled for this transport, but not configured")
			}

			authorizer, err := buildSpiffeAuthorizer(st.TLS.Spiffe)
			if err != nil {
				return nil, fmt.Errorf("unable to build SPIFFE authorizer: %w", err)
			}

			tlsConfig = tlsconfig.MTLSClientConfig(d.spiffeX509Source, d.spiffeX509Source, authorizer)
		}

		if st.TLS.InsecureSkipVerify || len(st.TLS.RootCAs) > 0 || len(st.TLS.ServerName) > 0 || len(st.TLS.Certificates) > 0 || st.TLS.PeerCertURI != "" {
			if tlsConfig != nil {
				return nil, errors.New("TLS and SPIFFE configuration cannot be defined at the same time")
			}

			tlsConfig = &tls.Config{
				ServerName:         st.TLS.ServerName,
				InsecureSkipVerify: st.TLS.InsecureSkipVerify,
				RootCAs:            createRootCACertPool(st.TLS.RootCAs),
				Certificates:       st.TLS.Certificates.GetCertificates(),
			}

			if st.TLS.PeerCertURI != "" {
				tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
					return traefiktls.VerifyPeerCertificate(st.TLS.PeerCertURI, tlsConfig, rawCerts)
				}
			}
		}
	}

	dialer := tcpDialer{
		dialer: &net.Dialer{
			Timeout:   time.Duration(st.DialTimeout),
			KeepAlive: time.Duration(st.DialKeepAlive),
		},
		terminationDelay: time.Duration(terminationDelay),
		proxyProtocol:    proxyProtocol,
	}

	if !isTLS {
		return dialer, nil
	}
	return tcpTLSDialer{dialer, tlsConfig}, nil
}

func createRootCACertPool(rootCAs []types.FileOrContent) *x509.CertPool {
	if len(rootCAs) == 0 {
		return nil
	}

	roots := x509.NewCertPool()

	for _, cert := range rootCAs {
		certContent, err := cert.Read()
		if err != nil {
			log.Err(err).Msg("Error while read RootCAs")
			continue
		}

		roots.AppendCertsFromPEM(certContent)
	}

	return roots
}

func buildSpiffeAuthorizer(cfg *dynamic.Spiffe) (tlsconfig.Authorizer, error) {
	switch {
	case len(cfg.IDs) > 0:
		spiffeIDs := make([]spiffeid.ID, 0, len(cfg.IDs))
		for _, rawID := range cfg.IDs {
			id, err := spiffeid.FromString(rawID)
			if err != nil {
				return nil, fmt.Errorf("invalid SPIFFE ID: %w", err)
			}

			spiffeIDs = append(spiffeIDs, id)
		}

		return tlsconfig.AuthorizeOneOf(spiffeIDs...), nil

	case cfg.TrustDomain != "":
		trustDomain, err := spiffeid.TrustDomainFromString(cfg.TrustDomain)
		if err != nil {
			return nil, fmt.Errorf("invalid SPIFFE trust domain: %w", err)
		}

		return tlsconfig.AuthorizeMemberOf(trustDomain), nil

	default:
		return tlsconfig.AuthorizeAny(), nil
	}
}
