package tcp

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	tcpmiddleware "github.com/traefik/traefik/v3/pkg/server/middleware/tcp"
	"github.com/traefik/traefik/v3/pkg/server/service/tcp"
	tcp2 "github.com/traefik/traefik/v3/pkg/tcp"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/tls/generate"
	"github.com/traefik/traefik/v3/pkg/types"
)

type applyRouter func(conf *runtime.Configuration)

type checkRouter func(addr string, timeout time.Duration) error

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

// Close closes the Listener.
func (h *httpForwarder) Close() error {
	h.errChan <- http.ErrServerClosed

	return nil
}

// ServeTCP uses the connection to serve it later in "Accept".
func (h *httpForwarder) ServeTCP(conn tcp2.WriteCloser) {
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

// Test_Routing aims to settle the behavior between routers of different types on the same TCP entryPoint.
// It has been introduced as a regression test following a fix on the v2.7 TCP Muxer.
//
// The routing precedence is roughly as follows:
// - TCP over HTTP
// - HTTPS over TCP-TLS
//
// Discrepancies for server sending first bytes support:
// - On v2.6, it is possible as long as you have one and only one TCP Non-TLS HostSNI(`*`) router (so called CatchAllNoTLS) defined.
// - On v2.7, it is possible as long as you have zero TLS/HTTPS router defined.
//
// Discrepancies in routing precedence between TCP and HTTP routers:
// - TCP HostSNI(`*`) and HTTP Host(`foobar`)
//   - On v2.6 and v2.7, the TCP one takes precedence.
//
// - TCP ClientIP(`[::]`) and HTTP Host(`foobar`)
//   - On v2.6, ClientIP matcher doesn't exist.
//   - On v2.7, the TCP one takes precedence.
//
// Routing precedence between TCP-TLS and HTTPS routers (considering a request/connection with the servername "foobar"):
// - TCP-TLS HostSNI(`*`) and HTTPS Host(`foobar`)
//   - On v2.6 and v2.7, the HTTPS one takes precedence.
//
// - TCP-TLS HostSNI(`foobar`) and HTTPS Host(`foobar`)
//   - On v2.6 and v2.7, the HTTPS one takes precedence (overriding the TCP-TLS one in v2.6).
//
// - TCP-TLS HostSNI(`*`) and HTTPS PathPrefix(`/`)
//   - On v2.6 and v2.7, the HTTPS one takes precedence (overriding the TCP-TLS one in v2.6).
//
// - TCP-TLS HostSNI(`foobar`) and HTTPS PathPrefix(`/`)
//   - On v2.6 and v2.7, the TCP-TLS one takes precedence.
func Test_Routing(t *testing.T) {
	// This listener simulates the backend service.
	// It is capable of switching into server first communication mode,
	// if the client takes to long to send the first bytes.
	tcpBackendListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	// This allows the closing of the TCP backend listener to happen last.
	t.Cleanup(func() {
		tcpBackendListener.Close()
	})

	go func() {
		for {
			conn, err := tcpBackendListener.Accept()
			if err != nil {
				var opErr *net.OpError
				if errors.As(err, &opErr) && opErr.Temporary() {
					continue
				}

				var urlErr *url.Error
				if errors.As(err, &urlErr) && urlErr.Temporary() {
					continue
				}

				return
			}

			err = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
			if err != nil {
				return
			}

			buf := make([]byte, 100)
			_, err = conn.Read(buf)

			var opErr *net.OpError
			if err == nil {
				_, err = fmt.Fprint(conn, "TCP-CLIENT-FIRST")
				require.NoError(t, err)
			} else if errors.As(err, &opErr) && opErr.Timeout() {
				_, err = fmt.Fprint(conn, "TCP-SERVER-FIRST")
				require.NoError(t, err)
			}

			err = conn.Close()
			require.NoError(t, err)
		}
	}()

	// Configuration defining the TCP backend service, used by TCP routers later.
	conf := &runtime.Configuration{
		TCPServices: map[string]*runtime.TCPServiceInfo{
			"tcp": {
				TCPService: &dynamic.TCPService{
					LoadBalancer: &dynamic.TCPServersLoadBalancer{
						Servers: []dynamic.TCPServer{
							{
								Address: tcpBackendListener.Addr().String(),
							},
						},
					},
				},
			},
		},
	}

	dialerManager := tcp2.NewDialerManager(nil)
	dialerManager.Update(map[string]*dynamic.TCPServersTransport{"default@internal": {}})
	serviceManager := tcp.NewManager(conf, dialerManager)

	certPEM, keyPEM, err := generate.KeyPair("foo.bar", time.Time{})
	require.NoError(t, err)

	// Creates the tlsManager and defines the TLS 1.0 and 1.2 TLSOptions.
	tlsManager := traefiktls.NewManager(nil)
	tlsManager.UpdateConfigs(
		t.Context(),
		map[string]traefiktls.Store{
			tlsalpn01.ACMETLS1Protocol: {},
		},
		map[string]traefiktls.Options{
			"default": {
				MinVersion: "VersionTLS10",
				MaxVersion: "VersionTLS10",
			},
			"tls10": {
				MinVersion: "VersionTLS10",
				MaxVersion: "VersionTLS10",
			},
			"tls12": {
				MinVersion: "VersionTLS12",
				MaxVersion: "VersionTLS12",
			},
		},
		[]*traefiktls.CertAndStores{{
			Certificate: traefiktls.Certificate{CertFile: types.FileOrContent(certPEM), KeyFile: types.FileOrContent(keyPEM)},
			Stores:      []string{tlsalpn01.ACMETLS1Protocol},
		}})

	middlewaresBuilder := tcpmiddleware.NewBuilder(conf.TCPMiddlewares)

	manager := NewManager(conf, serviceManager, middlewaresBuilder,
		nil, nil, tlsManager)

	type checkCase struct {
		checkRouter

		desc          string
		expectedError string
		timeout       time.Duration
	}

	testCases := []struct {
		desc                    string
		routers                 []applyRouter
		checks                  []checkCase
		allowACMETLSPassthrough bool
	}{
		{
			desc:    "No routers",
			routers: []applyRouter{},
			checks: []checkCase{
				{
					desc:        "ACME TLS Challenge",
					checkRouter: checkACMETLS,
				},
				{
					desc:          "TCP with client sending first bytes should fail",
					checkRouter:   checkTCPClientFirst,
					expectedError: "i/o timeout",
				},
				{
					desc:          "TCP with server sending first bytes should fail",
					checkRouter:   checkTCPServerFirst,
					expectedError: "i/o timeout",
				},
				{
					desc:        "HTTP request should be handled by HTTP service (404)",
					checkRouter: checkHTTP,
				},
				{
					desc:          "TCP TLS 1.0 connection should fail",
					checkRouter:   checkTCPTLS10,
					expectedError: "i/o timeout",
				},
				{
					desc:        "TCP TLS 1.2 connection should fail",
					checkRouter: checkTCPTLS12,
					// The HTTPS forwarder catches the connection with the TLS 1.0 config,
					// because no matching routes are defined with the custom TLS Config.
					expectedError: "wrong TLS version",
				},
				{
					desc:        "HTTPS TLS 1.0 request should be handled by HTTPS (HTTPS forwarder with tls 1.0 config) (404)",
					checkRouter: checkHTTPSTLS10,
				},
				{
					desc:          "HTTPS TLS 1.2 request should fail",
					checkRouter:   checkHTTPSTLS12,
					expectedError: "wrong TLS version",
				},
			},
		},
		{
			desc:    "TCP TLS passthrough does not catch ACME TLS",
			routers: []applyRouter{routerTCPTLSCatchAllPassthrough},
			checks: []checkCase{
				{
					desc:        "ACME TLS Challenge",
					checkRouter: checkACMETLS,
				},
			},
		},
		{
			desc:                    "TCP TLS passthrough catches ACME TLS",
			allowACMETLSPassthrough: true,
			routers:                 []applyRouter{routerTCPTLSCatchAllPassthrough},
			checks: []checkCase{
				{
					desc:          "ACME TLS Challenge",
					checkRouter:   checkACMETLS,
					expectedError: "tls: first record does not look like a TLS handshake",
				},
			},
		},
		{
			desc:    "Single TCP CatchAll router",
			routers: []applyRouter{routerTCPCatchAll},
			checks: []checkCase{
				{
					desc:        "TCP with client sending first bytes should be handled by TCP service",
					checkRouter: checkTCPClientFirst,
				},
				{
					desc:        "TCP with server sending first bytes should be handled by TCP service",
					checkRouter: checkTCPServerFirst,
				},
			},
		},
		{
			desc:    "Single HTTP router",
			routers: []applyRouter{routerHTTP},
			checks: []checkCase{
				{
					desc:        "HTTP request should be handled by HTTP service",
					checkRouter: checkHTTP,
				},
			},
		},
		{
			desc:    "Single TCP TLS router",
			routers: []applyRouter{routerTCPTLS},
			checks: []checkCase{
				{
					desc:          "TCP TLS 1.0 connection should fail",
					checkRouter:   checkTCPTLS10,
					expectedError: "wrong TLS version",
				},
				{
					desc:        "TCP TLS 1.2 connection should be handled by TCP service",
					checkRouter: checkTCPTLS12,
				},
			},
		},
		{
			desc:    "Single TCP TLS CatchAll router",
			routers: []applyRouter{routerTCPTLSCatchAll},
			checks: []checkCase{
				{
					desc:        "TCP TLS 1.0 connection should be handled by TCP service",
					checkRouter: checkTCPTLS10,
				},
				{
					desc:          "TCP TLS 1.2 connection should fail",
					checkRouter:   checkTCPTLS12,
					expectedError: "wrong TLS version",
				},
			},
		},
		{
			desc:    "Single HTTPS router",
			routers: []applyRouter{routerHTTPS},
			checks: []checkCase{
				{
					desc:          "HTTPS TLS 1.0 request should fail",
					checkRouter:   checkHTTPSTLS10,
					expectedError: "wrong TLS version",
				},
				{
					desc:        "HTTPS TLS 1.2 request should be handled by HTTPS service",
					checkRouter: checkHTTPSTLS12,
				},
			},
		},
		{
			desc:    "Single HTTPS PathPrefix router",
			routers: []applyRouter{routerHTTPSPathPrefix},
			checks: []checkCase{
				{
					desc:        "HTTPS TLS 1.0 request should be handled by HTTPS service",
					checkRouter: checkHTTPSTLS10,
				},
				{
					desc:          "HTTPS TLS 1.2 request should fail",
					checkRouter:   checkHTTPSTLS12,
					expectedError: "wrong TLS version",
				},
			},
		},
		{
			desc:    "TCP CatchAll router && HTTP router",
			routers: []applyRouter{routerTCPCatchAll, routerHTTP},
			checks: []checkCase{
				{
					desc:        "TCP client sending first bytes should be handled by TCP service",
					checkRouter: checkTCPClientFirst,
				},
				{
					desc:        "TCP server sending first bytes should be handled by TCP service",
					checkRouter: checkTCPServerFirst,
				},
				{
					desc:          "HTTP request should fail, because handled by TCP service",
					checkRouter:   checkHTTP,
					expectedError: "malformed HTTP response",
				},
			},
		},
		{
			desc:    "TCP TLS CatchAll router && HTTP router",
			routers: []applyRouter{routerTCPTLS, routerHTTP},
			checks: []checkCase{
				{
					desc:          "TCP TLS 1.0 connection should fail",
					checkRouter:   checkTCPTLS10,
					expectedError: "wrong TLS version",
				},
				{
					desc:        "TCP TLS 1.2 connection should be handled by TCP service",
					checkRouter: checkTCPTLS12,
				},
				{
					desc:        "HTTP request should be handled by HTTP service",
					checkRouter: checkHTTP,
				},
			},
		},
		{
			desc:    "TCP CatchAll router && HTTPS router",
			routers: []applyRouter{routerTCPCatchAll, routerHTTPS},
			checks: []checkCase{
				{
					desc:        "TCP client sending first bytes should be handled by TCP service",
					checkRouter: checkTCPClientFirst,
				},
				{
					desc:          "TCP server sending first bytes should timeout on clientHello",
					checkRouter:   checkTCPServerFirst,
					expectedError: "i/o timeout",
				},
				{
					desc:          "HTTP request should fail, because handled by TCP service",
					checkRouter:   checkHTTP,
					expectedError: "malformed HTTP response",
				},
				{
					desc:          "HTTPS TLS 1.0 request should be handled by HTTPS service",
					checkRouter:   checkHTTPSTLS10,
					expectedError: "wrong TLS version",
				},
				{
					desc:        "HTTPS TLS 1.2 request should be handled by HTTPS service",
					checkRouter: checkHTTPSTLS12,
				},
			},
		},
		{
			// We show that a not CatchAll HTTPS router takes priority over a TCP-TLS router.
			desc:    "TCP TLS router && HTTPS router",
			routers: []applyRouter{routerTCPTLS, routerHTTPS},
			checks: []checkCase{
				{
					desc:          "TCP TLS 1.0 connection should fail",
					checkRouter:   checkTCPTLS10,
					expectedError: "wrong TLS version",
				},
				{
					desc:        "TCP TLS 1.2 connection should fail",
					checkRouter: checkTCPTLS12,
					// The connection is handled by the HTTPS router,
					// which has the correct TLS config,
					// but the HTTP server is detecting a malformed request which ends with a timeout.
					expectedError: "i/o timeout",
				},
				{
					desc:          "HTTPS TLS 1.0 request should fail",
					checkRouter:   checkHTTPSTLS10,
					expectedError: "wrong TLS version",
				},
				{
					desc:        "HTTPS TLS 1.2 request should be handled by HTTPS service",
					checkRouter: checkHTTPSTLS12,
				},
			},
		},
		{
			// We show that a not CatchAll HTTPS router takes priority over a CatchAll TCP-TLS router.
			desc:    "TCP TLS CatchAll router && HTTPS router",
			routers: []applyRouter{routerTCPCatchAll, routerHTTPS},
			checks: []checkCase{
				{
					desc:          "TCP TLS 1.0 connection should fail",
					checkRouter:   checkTCPTLS10,
					expectedError: "wrong TLS version",
				},
				{
					desc:        "TCP TLS 1.2 connection should fail",
					checkRouter: checkTCPTLS12,
					// The connection is handled by the HTTPS router,
					// which has the correct TLS config,
					// but the HTTP server is detecting a malformed request which ends with a timeout.
					expectedError: "i/o timeout",
				},
				{
					desc:          "HTTPS TLS 1.0 request should fail",
					checkRouter:   checkHTTPSTLS10,
					expectedError: "wrong TLS version",
				},
				{
					desc:        "HTTPS TLS 1.2 request should be handled by HTTPS service",
					checkRouter: checkHTTPSTLS12,
				},
			},
		},
		{
			// We show that TCP-TLS router (not CatchAll) takes priority over non-Host rule HTTPS router (CatchAll).
			desc:    "TCP TLS router && HTTPS Path prefix router",
			routers: []applyRouter{routerTCPTLS, routerHTTPSPathPrefix},
			checks: []checkCase{
				{
					desc:          "TCP TLS 1.0 connection should fail",
					checkRouter:   checkTCPTLS10,
					expectedError: "wrong TLS version",
				},
				{
					desc:        "TCP TLS 1.2 connection should be handled by TCP service",
					checkRouter: checkTCPTLS12,
				},
				{
					desc:          "HTTPS TLS 1.0 request should fail",
					checkRouter:   checkHTTPSTLS10,
					expectedError: "malformed HTTP response",
				},
				{
					desc:          "HTTPS TLS 1.2 should fail",
					checkRouter:   checkHTTPSTLS12,
					expectedError: "malformed HTTP response",
				},
			},
		},
		{
			desc:    "TCP TLS router && TCP TLS CatchAll router",
			routers: []applyRouter{routerTCPTLS, routerTCPTLSCatchAll},
			checks: []checkCase{
				{
					desc:          "TCP TLS 1.0 connection should fail",
					checkRouter:   checkTCPTLS10,
					expectedError: "wrong TLS version",
				},
				{
					desc:        "TCP TLS 1.2 connection should be handled by TCP service",
					checkRouter: checkTCPTLS12,
				},
			},
		},
		{
			desc:    "HTTPS router && HTTPS CatchAll router",
			routers: []applyRouter{routerHTTPS, routerHTTPSPathPrefix},
			checks: []checkCase{
				{
					desc:          "HTTPS TLS 1.0 request should fail",
					checkRouter:   checkHTTPSTLS10,
					expectedError: "wrong TLS version",
				},
				{
					desc:        "HTTPS TLS 1.2 request should be handled by HTTPS service",
					checkRouter: checkHTTPSTLS12,
				},
			},
		},
		{
			desc:    "All routers, all checks",
			routers: []applyRouter{routerTCPCatchAll, routerHTTP, routerHTTPS, routerTCPTLS, routerTCPTLSCatchAll},
			checks: []checkCase{
				{
					desc:        "TCP client sending first bytes should be handled by TCP service",
					checkRouter: checkTCPClientFirst,
				},
				{
					desc:          "TCP server sending first bytes should timeout on clientHello",
					checkRouter:   checkTCPServerFirst,
					expectedError: "i/o timeout",
				},
				{
					desc:          "HTTP request should fail, because handled by TCP service",
					checkRouter:   checkHTTP,
					expectedError: "malformed HTTP response",
				},
				{
					desc:          "HTTPS TLS 1.0 request should fail",
					checkRouter:   checkHTTPSTLS10,
					expectedError: "wrong TLS version",
				},
				{
					desc:        "HTTPS TLS 1.2 request should be handled by HTTPS service",
					checkRouter: checkHTTPSTLS12,
				},
				{
					desc:          "TCP TLS 1.0 connection should fail",
					checkRouter:   checkTCPTLS10,
					expectedError: "wrong TLS version",
				},
				{
					desc:        "TCP TLS 1.2 connection should fail",
					checkRouter: checkTCPTLS12,
					// The connection is handled by the HTTPS router,
					// witch have the correct TLS config,
					// but the HTTP server is detecting a malformed request which ends with a timeout.
					expectedError: "i/o timeout",
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dynConf := &runtime.Configuration{
				Routers:    map[string]*runtime.RouterInfo{},
				TCPRouters: map[string]*runtime.TCPRouterInfo{},
			}

			for _, router := range test.routers {
				router(dynConf)
			}

			router, err := manager.buildEntryPointHandler(t.Context(), dynConf.TCPRouters, dynConf.Routers, nil, nil)
			require.NoError(t, err)

			if test.allowACMETLSPassthrough {
				router.EnableACMETLSPassthrough()
			}

			epListener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)

			// serverHTTP handler returns only the "HTTP" value as body for further checks.
			serverHTTP := &http.Server{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, err = fmt.Fprint(w, "HTTP")
					require.NoError(t, err)
				}),
			}

			stoppedHTTP := make(chan struct{})
			forwarder := newHTTPForwarder(epListener)
			go func() {
				defer close(stoppedHTTP)
				_ = serverHTTP.Serve(forwarder)
			}()

			router.SetHTTPForwarder(forwarder)

			// serverHTTPS handler returns only the "HTTPS" value as body for further checks.
			serverHTTPS := &http.Server{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, err = fmt.Fprint(w, "HTTPS")
					require.NoError(t, err)
				}),
			}

			stoppedHTTPS := make(chan struct{})
			httpsForwarder := newHTTPForwarder(epListener)
			go func() {
				defer close(stoppedHTTPS)
				_ = serverHTTPS.Serve(httpsForwarder)
			}()

			router.SetHTTPSForwarder(httpsForwarder)

			stoppedTCP := make(chan struct{})
			go func() {
				defer close(stoppedTCP)
				for {
					conn, err := epListener.Accept()
					if err != nil {
						return
					}

					tcpConn, ok := conn.(*net.TCPConn)
					if !ok {
						t.Error("not a write closer")
					}

					router.ServeTCP(tcpConn)
				}
			}()

			for _, check := range test.checks {
				timeout := 2 * time.Second
				if check.timeout > 0 {
					timeout = check.timeout
				}

				err := check.checkRouter(epListener.Addr().String(), timeout)

				if check.expectedError != "" {
					require.Error(t, err, check.desc)
					assert.Contains(t, err.Error(), check.expectedError, check.desc)
					continue
				}

				assert.NoError(t, err, check.desc)
			}

			epListener.Close()

			<-stoppedTCP

			serverHTTP.Close()
			serverHTTPS.Close()

			<-stoppedHTTP
			<-stoppedHTTPS
		})
	}
}

// routerTCPCatchAll configures a TCP CatchAll No TLS - HostSNI(`*`) router.
func routerTCPCatchAll(conf *runtime.Configuration) {
	conf.TCPRouters["tcp-catchall"] = &runtime.TCPRouterInfo{
		TCPRouter: &dynamic.TCPRouter{
			EntryPoints: []string{"web"},
			Service:     "tcp",
			Rule:        "HostSNI(`*`)",
		},
	}
}

// routerHTTP configures an HTTP - Host(`foo.bar`) router.
func routerHTTP(conf *runtime.Configuration) {
	conf.Routers["http"] = &runtime.RouterInfo{
		Router: &dynamic.Router{
			EntryPoints: []string{"web"},
			Service:     "http",
			Rule:        "Host(`foo.bar`)",
		},
	}
}

// routerTCPTLSCatchAll a TCP TLS CatchAll - HostSNI(`*`) router with TLS 1.0 config.
func routerTCPTLSCatchAll(conf *runtime.Configuration) {
	conf.TCPRouters["tcp-tls-catchall"] = &runtime.TCPRouterInfo{
		TCPRouter: &dynamic.TCPRouter{
			EntryPoints: []string{"web"},
			Service:     "tcp",
			Rule:        "HostSNI(`*`)",
			TLS: &dynamic.RouterTCPTLSConfig{
				Options: "tls10",
			},
		},
	}
}

// routerTCPTLSCatchAllPassthrough a TCP TLS CatchAll Passthrough - HostSNI(`*`) router with TLS 1.2 config.
func routerTCPTLSCatchAllPassthrough(conf *runtime.Configuration) {
	conf.TCPRouters["tcp-tls-catchall-passthrough"] = &runtime.TCPRouterInfo{
		TCPRouter: &dynamic.TCPRouter{
			EntryPoints: []string{"web"},
			Service:     "tcp",
			Rule:        "HostSNI(`*`)",
			TLS: &dynamic.RouterTCPTLSConfig{
				Options:     "tls12",
				Passthrough: true,
			},
		},
	}
}

// routerTCPTLS configures a TCP TLS - HostSNI(`foo.bar`) router with TLS 1.2 config.
func routerTCPTLS(conf *runtime.Configuration) {
	conf.TCPRouters["tcp-tls"] = &runtime.TCPRouterInfo{
		TCPRouter: &dynamic.TCPRouter{
			EntryPoints: []string{"web"},
			Service:     "tcp",
			Rule:        "HostSNI(`foo.bar`)",
			TLS: &dynamic.RouterTCPTLSConfig{
				Options: "tls12",
			},
		},
	}
}

// routerHTTPSPathPrefix configures an HTTPS - PathPrefix(`/`) router with TLS 1.0 config.
func routerHTTPSPathPrefix(conf *runtime.Configuration) {
	conf.Routers["https"] = &runtime.RouterInfo{
		Router: &dynamic.Router{
			EntryPoints: []string{"web"},
			Service:     "http",
			Rule:        "PathPrefix(`/`)",
			TLS: &dynamic.RouterTLSConfig{
				Options: "tls10",
			},
		},
	}
}

// routerHTTPS configures an HTTPS - Host(`foo.bar`) router with TLS 1.2 config.
func routerHTTPS(conf *runtime.Configuration) {
	conf.Routers["https-custom-tls"] = &runtime.RouterInfo{
		Router: &dynamic.Router{
			EntryPoints: []string{"web"},
			Service:     "http",
			Rule:        "Host(`foo.bar`)",
			TLS: &dynamic.RouterTLSConfig{
				Options: "tls12",
			},
		},
	}
}

// checkACMETLS simulates a ACME TLS Challenge client connection.
// It returns an error if TLS handshake fails.
func checkACMETLS(addr string, _ time.Duration) (err error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "foo.bar",
		MinVersion:         tls.VersionTLS10,
		NextProtos:         []string{tlsalpn01.ACMETLS1Protocol},
	}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if conn.ConnectionState().Version != tls.VersionTLS10 {
		return fmt.Errorf("wrong TLS version. wanted %X, got %X", tls.VersionTLS10, conn.ConnectionState().Version)
	}

	return nil
}

// checkTCPClientFirst simulates a TCP client sending first bytes first.
// It returns an error if it doesn't receive the expected response.
func checkTCPClientFirst(addr string, timeout time.Duration) (err error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	fmt.Fprint(conn, "HELLO")

	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, conn)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(buf.String(), "TCP-CLIENT-FIRST") {
		return fmt.Errorf("unexpected response: %s", buf.String())
	}

	return nil
}

// checkTCPServerFirst simulates a TCP client waiting for the server first bytes.
// It returns an error if it doesn't receive the expected response.
func checkTCPServerFirst(addr string, timeout time.Duration) (err error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, conn)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(buf.String(), "TCP-SERVER-FIRST") {
		return fmt.Errorf("unexpected response: %s", buf.String())
	}

	return nil
}

// checkHTTP simulates an HTTP client.
// It returns an error if it doesn't receive the expected response.
func checkHTTP(addr string, timeout time.Duration) error {
	httpClient := &http.Client{Timeout: timeout}

	req, err := http.NewRequest(http.MethodGet, "http://"+addr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Host", "foo.bar")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if !strings.Contains(string(body), "HTTP") {
		return fmt.Errorf("unexpected response: %s", string(body))
	}

	return nil
}

// checkTCPTLS simulates a TCP client connection.
// It returns an error if it doesn't receive the expected response.
func checkTCPTLS(addr string, timeout time.Duration, tlsVersion uint16) (err error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "foo.bar",
		MinVersion:         tls.VersionTLS10,
		MaxVersion:         tls.VersionTLS12,
	}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if conn.ConnectionState().Version != tlsVersion {
		return fmt.Errorf("wrong TLS version. wanted %X, got %X", tlsVersion, conn.ConnectionState().Version)
	}

	fmt.Fprint(conn, "HELLO")

	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, conn)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(buf.String(), "TCP-CLIENT-FIRST") {
		return fmt.Errorf("unexpected response: %s", buf.String())
	}

	return nil
}

// checkTCPTLS10 simulates a TCP client connection with TLS 1.0.
// It returns an error if it doesn't receive the expected response.
func checkTCPTLS10(addr string, timeout time.Duration) error {
	return checkTCPTLS(addr, timeout, tls.VersionTLS10)
}

// checkTCPTLS12 simulates a TCP client connection with TLS 1.2.
// It returns an error if it doesn't receive the expected response.
func checkTCPTLS12(addr string, timeout time.Duration) error {
	return checkTCPTLS(addr, timeout, tls.VersionTLS12)
}

// checkHTTPS makes an HTTPS request and checks the given TLS.
// It returns an error if it doesn't receive the expected response.
func checkHTTPS(addr string, timeout time.Duration, tlsVersion uint16) error {
	req, err := http.NewRequest(http.MethodGet, "https://"+addr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Host", "foo.bar")

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         "foo.bar",
				MinVersion:         tls.VersionTLS10,
				MaxVersion:         tls.VersionTLS12,
			},
		},
		Timeout: timeout,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.TLS.Version != tlsVersion {
		return fmt.Errorf("wrong TLS version. wanted %X, got %X", tlsVersion, resp.TLS.Version)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if !strings.Contains(string(body), "HTTPS") {
		return fmt.Errorf("unexpected response: %s", string(body))
	}

	return nil
}

// checkHTTPSTLS10 makes an HTTP request with TLS version 1.0.
// It returns an error if it doesn't receive the expected response.
func checkHTTPSTLS10(addr string, timeout time.Duration) error {
	return checkHTTPS(addr, timeout, tls.VersionTLS10)
}

// checkHTTPSTLS12 makes an HTTP request with TLS version 1.2.
// It returns an error if it doesn't receive the expected response.
func checkHTTPSTLS12(addr string, timeout time.Duration) error {
	return checkHTTPS(addr, timeout, tls.VersionTLS12)
}

func TestPostgres(t *testing.T) {
	router, err := NewRouter()
	require.NoError(t, err)

	// This test requires to have a TLS route, but does not actually check the
	// content of the handler. It would require to code a TLS handshake to
	// check the SNI and content of the handlerFunc.
	err = router.muxerTCPTLS.AddRoute("HostSNI(`test.localhost`)", "", 0, nil)
	require.NoError(t, err)

	err = router.muxerTCP.AddRoute("HostSNI(`*`)", "", 0, tcp2.HandlerFunc(func(conn tcp2.WriteCloser) {
		_, _ = conn.Write([]byte("OK"))
		_ = conn.Close()
	}))
	require.NoError(t, err)

	mockConn := NewMockConn()
	go router.ServeTCP(mockConn)

	mockConn.dataRead <- PostgresStartTLSMsg
	b := <-mockConn.dataWrite
	require.Equal(t, PostgresStartTLSReply, b)

	mockConn = NewMockConn()
	go router.ServeTCP(mockConn)

	mockConn.dataRead <- []byte("HTTP")
	b = <-mockConn.dataWrite
	require.Equal(t, []byte("OK"), b)
}

func NewMockConn() *MockConn {
	return &MockConn{
		dataRead:  make(chan []byte),
		dataWrite: make(chan []byte),
	}
}

type MockConn struct {
	dataRead  chan []byte
	dataWrite chan []byte
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	temp := <-m.dataRead
	copy(b, temp)
	return len(temp), nil
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	m.dataWrite <- b
	return len(b), nil
}

func (m *MockConn) Close() error {
	close(m.dataRead)
	close(m.dataWrite)
	return nil
}

func (m *MockConn) LocalAddr() net.Addr {
	return nil
}

func (m *MockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{}
}

func (m *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) CloseWrite() error {
	close(m.dataRead)
	close(m.dataWrite)
	return nil
}
