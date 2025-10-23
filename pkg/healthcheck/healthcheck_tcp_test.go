package healthcheck

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	truntime "github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

var localhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIIDJzCCAg+gAwIBAgIUe3vnWg3cTbflL6kz2TyPUxmV8Y4wDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAwwLZXhhbXBsZS5jb20wIBcNMjUwMzA1MjAwOTM4WhgPMjA1
NTAyMjYyMDA5MzhaMBYxFDASBgNVBAMMC2V4YW1wbGUuY29tMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4Mm4Sp6xzJvFZJWAv/KVmI1krywiuef8Fhlf
JR2M0caKixjBcNt4U8KwrzIrqL+8nilbps1QuwpQ09+6ztlbUXUL6DqR8ZC+4oCp
gOZ3yyVX2vhMigkATbQyJrX/WVjWSHD5rIUBP2BrsaYLt1qETnFP9wwQ3YEi7V4l
c4+jDrZOtJvrv+tRClt9gQJVgkr7Y30X+dx+rsh+ROaA2+/VTDX0qtoqd/4fjhcJ
OY9VLm0eU66VUMyOTNeUm6ZAXRBp/EonIM1FXOlj82S0pZQbPrvyWWqWoAjtPvLU
qRzqp/BQJqx3EHz1dP6s+xUjP999B+7jhiHoFhZ/bfVVlx8XkwIDAQABo2swaTAd
BgNVHQ4EFgQUhJiJ37LW6RODCpBPAApG1zQxFtAwHwYDVR0jBBgwFoAUhJiJ37LW
6RODCpBPAApG1zQxFtAwDwYDVR0TAQH/BAUwAwEB/zAWBgNVHREEDzANggtleGFt
cGxlLmNvbTANBgkqhkiG9w0BAQsFAAOCAQEAfnDPHllA1TFlQ6zY46tqM20d68bR
kXeGMKLoaATFPbDea5H8/GM5CU6CPD7RUuEB9CvxvaM0aOInxkgstozG7BOr8hcs
WS9fMgM0oO5yGiSOv+Qa0Rc0BFb6A1fUJRta5MI5DTdTJLoyoRX/5aocSI34T67x
ULbkJvVXw6hnx/KZ65apNobfmVQSy7DR8Fo82eB4hSoaLpXyUUTLmctGgrRCoKof
GVUJfKsDJ4Ts8WIR1np74flSoxksWSHEOYk79AZOPANYgJwPMMiiZKsKm17GBoGu
DxI0om4eX8GaSSZAtG6TOt3O3v1oCjKNsAC+u585HN0x0MFA33TUzC15NA==
-----END CERTIFICATE-----`)

var localhostKey = []byte(`-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDgybhKnrHMm8Vk
lYC/8pWYjWSvLCK55/wWGV8lHYzRxoqLGMFw23hTwrCvMiuov7yeKVumzVC7ClDT
37rO2VtRdQvoOpHxkL7igKmA5nfLJVfa+EyKCQBNtDImtf9ZWNZIcPmshQE/YGux
pgu3WoROcU/3DBDdgSLtXiVzj6MOtk60m+u/61EKW32BAlWCSvtjfRf53H6uyH5E
5oDb79VMNfSq2ip3/h+OFwk5j1UubR5TrpVQzI5M15SbpkBdEGn8SicgzUVc6WPz
ZLSllBs+u/JZapagCO0+8tSpHOqn8FAmrHcQfPV0/qz7FSM/330H7uOGIegWFn9t
9VWXHxeTAgMBAAECggEALinfGhv7Iaz/3cdCOKlGBZ1MBxmGTC2TPKqbOpEWAWLH
wwcjetznmjQKewBPrQkrYEPYGapioPbeYJS61Y4XzeO+vUOCA10ZhoSrytgJ1ANo
RoTlmxd8I3kVL5QCy8ONxjTFYaOy/OP9We9iypXhRAbLSE4HDKZfmOXTxSbDctql
Kq7uV3LX1KCfr9C6M8d79a0Rdr4p8IXp8MOg3tXq6n75vZbepRFyAujhg7o/kkTp
lgv87h89lrK97K+AjqtvCIT3X3VXfA+LYp3AoQFdOluKgyJT221MyHkTeI/7gggt
Z57lVGD71UJH/LGUJWrraJqXd9uDxZWprD/s66BIAQKBgQD8CtHUJ/VuS7gP0ebN
688zrmRtENj6Gqi+URm/Pwgr9b7wKKlf9jjhg5F/ue+BgB7/nK6N7yJ4Xx3JJ5ox
LqsRGLFa4fDBxogF/FN27obD8naOxe2wS1uTjM6LSrvdJ+HjeNEwHYhjuDjTAHj5
VVEMagZWgkE4jBiFUYefiYLsAQKBgQDkUVdW8cXaYri5xxDW86JNUzI1tUPyd6I+
AkOHV/V0y2zpwTHVLcETRpdVGpc5TH3J5vWf+5OvSz6RDTGjv7blDb8vB/kVkFmn
uXTi0dB9P+SYTsm+X3V7hOAFsyVYZ1D9IFsKUyMgxMdF+qgERjdPKx5IdLV/Jf3q
P9pQ922TkwKBgCKllhyU9Z8Y14+NKi4qeUxAb9uyUjFnUsT+vwxULNpmKL44yLfB
UCZoAKtPMwZZR2mZ70Dhm5pycNTDFeYm5Ssvesnkf0UT9oTkH9EcjvgGr5eGy9rN
MSSCWa46MsL/BYVQiWkU1jfnDiCrUvXrbX3IYWCo/TA5yfEhuQQMUiwBAoGADyzo
5TqEsBNHu/FjSSZAb2tMNw2pSoBxJDX6TxClm/G5d4AD0+uKncFfZaSy0HgpFDZp
tQx/sHML4ZBC8GNZwLe9MV8SS0Cg9Oj6v+i6Ntj8VLNH7YNix6b5TOevX8TeOTTh
WDpWZ2Ms65XRfRc9reFrzd0UAzN/QQaleCQ6AEkCgYBe4Ucows7JGbv7fNkz3nb1
kyH+hk9ecnq/evDKX7UUxKO1wwTi74IYKgcRB2uPLpHKL35gPz+LAfCphCW5rwpR
lvDhS+Pi/1KCBJxLHMv+V/WrckDRgHFnAhDaBZ+2vI/s09rKDnpjcTzV7x22kL0b
XIJCEEE8JZ4AXIZ+IcB6LA==
-----END PRIVATE KEY-----`)

func TestNewServiceTCPHealthChecker(t *testing.T) {
	testCases := []struct {
		desc             string
		config           *dynamic.TCPServerHealthCheck
		expectedInterval time.Duration
		expectedTimeout  time.Duration
	}{
		{
			desc:             "default values",
			config:           &dynamic.TCPServerHealthCheck{},
			expectedInterval: time.Duration(dynamic.DefaultHealthCheckInterval),
			expectedTimeout:  time.Duration(dynamic.DefaultHealthCheckTimeout),
		},
		{
			desc: "out of range values",
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(-time.Second),
				Timeout:  ptypes.Duration(-time.Second),
			},
			expectedInterval: time.Duration(dynamic.DefaultHealthCheckInterval),
			expectedTimeout:  time.Duration(dynamic.DefaultHealthCheckTimeout),
		},
		{
			desc: "custom durations",
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Second * 10),
				Timeout:  ptypes.Duration(time.Second * 5),
			},
			expectedInterval: time.Second * 10,
			expectedTimeout:  time.Second * 5,
		},
		{
			desc: "interval shorter than timeout",
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Second),
				Timeout:  ptypes.Duration(time.Second * 5),
			},
			expectedInterval: time.Second,
			expectedTimeout:  time.Second * 5,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			healthChecker := NewServiceTCPHealthChecker(t.Context(), test.config, nil, nil, nil, "")
			assert.Equal(t, test.expectedInterval, healthChecker.interval)
			assert.Equal(t, test.expectedTimeout, healthChecker.timeout)
		})
	}
}

func TestServiceTCPHealthChecker_executeHealthCheck_connection(t *testing.T) {
	testCases := []struct {
		desc            string
		address         string
		config          *dynamic.TCPServerHealthCheck
		expectedAddress string
	}{
		{
			desc:            "no port override - uses original address",
			address:         "127.0.0.1:8080",
			config:          &dynamic.TCPServerHealthCheck{Port: 0},
			expectedAddress: "127.0.0.1:8080",
		},
		{
			desc:            "port override - uses overridden port",
			address:         "127.0.0.1:8080",
			config:          &dynamic.TCPServerHealthCheck{Port: 9090},
			expectedAddress: "127.0.0.1:9090",
		},
		{
			desc:            "IPv6 address with port override",
			address:         "[::1]:8080",
			config:          &dynamic.TCPServerHealthCheck{Port: 9090},
			expectedAddress: "[::1]:9090",
		},
		{
			desc:            "successful connection without port override",
			address:         "localhost:3306",
			config:          &dynamic.TCPServerHealthCheck{Port: 0},
			expectedAddress: "localhost:3306",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			// Create a mock dialer that records the address it was asked to dial.
			var gotAddress string
			mockDialer := &dialerMock{
				onDial: func(network, addr string) (net.Conn, error) {
					gotAddress = addr
					return &connMock{}, nil
				},
			}

			targets := []TCPHealthCheckTarget{{
				Address: test.address,
				Dialer:  mockDialer,
			}}
			healthChecker := NewServiceTCPHealthChecker(t.Context(), test.config, nil, nil, targets, "test")

			// Execute a health check to see what address it tries to connect to.
			err := healthChecker.executeHealthCheck(t.Context(), test.config, &targets[0])
			require.NoError(t, err)

			// Verify that the health check attempted to connect to the expected address.
			assert.Equal(t, test.expectedAddress, gotAddress)
		})
	}
}

func TestServiceTCPHealthChecker_executeHealthCheck_payloadHandling(t *testing.T) {
	testCases := []struct {
		desc             string
		config           *dynamic.TCPServerHealthCheck
		mockResponse     string
		expectedSentData string
		expectedSuccess  bool
	}{
		{
			desc: "successful send and expect",
			config: &dynamic.TCPServerHealthCheck{
				Send:   "PING",
				Expect: "PONG",
			},
			mockResponse:     "PONG",
			expectedSentData: "PING",
			expectedSuccess:  true,
		},
		{
			desc: "send without expect",
			config: &dynamic.TCPServerHealthCheck{
				Send:   "STATUS",
				Expect: "",
			},
			expectedSentData: "STATUS",
			expectedSuccess:  true,
		},
		{
			desc: "send without expect, ignores response",
			config: &dynamic.TCPServerHealthCheck{
				Send: "STATUS",
			},
			mockResponse:     strings.Repeat("A", maxPayloadSize+1),
			expectedSentData: "STATUS",
			expectedSuccess:  true,
		},
		{
			desc: "expect without send",
			config: &dynamic.TCPServerHealthCheck{
				Expect: "READY",
			},
			mockResponse:    "READY",
			expectedSuccess: true,
		},
		{
			desc: "wrong response received",
			config: &dynamic.TCPServerHealthCheck{
				Send:   "PING",
				Expect: "PONG",
			},
			mockResponse:     "WRONG",
			expectedSentData: "PING",
			expectedSuccess:  false,
		},
		{
			desc: "send payload too large - gets truncated",
			config: &dynamic.TCPServerHealthCheck{
				Send:   strings.Repeat("A", maxPayloadSize+1), // Will be truncated to empty
				Expect: "OK",
			},
			mockResponse:    "OK",
			expectedSuccess: true,
		},
		{
			desc: "expect payload too large - gets truncated",
			config: &dynamic.TCPServerHealthCheck{
				Send:   "PING",
				Expect: strings.Repeat("B", maxPayloadSize+1), // Will be truncated to empty
			},
			expectedSentData: "PING",
			expectedSuccess:  true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var sentData []byte
			mockConn := &connMock{
				writeFunc: func(data []byte) (int, error) {
					sentData = append([]byte{}, data...)
					return len(data), nil
				},
				readFunc: func(buf []byte) (int, error) {
					return copy(buf, test.mockResponse), nil
				},
			}

			mockDialer := &dialerMock{
				onDial: func(network, addr string) (net.Conn, error) {
					return mockConn, nil
				},
			}

			targets := []TCPHealthCheckTarget{{
				Address: "127.0.0.1:8080",
				TLS:     false,
				Dialer:  mockDialer,
			}}

			healthChecker := NewServiceTCPHealthChecker(t.Context(), test.config, nil, nil, targets, "test")

			err := healthChecker.executeHealthCheck(t.Context(), test.config, &targets[0])

			if test.expectedSuccess {
				assert.NoError(t, err, "Health check should succeed")
			} else {
				assert.Error(t, err, "Health check should fail")
			}

			assert.Equal(t, test.expectedSentData, string(sentData), "Should send the expected data")
		})
	}
}

func TestServiceTCPHealthChecker_Launch(t *testing.T) {
	testCases := []struct {
		desc                  string
		server                *sequencedTCPServer
		config                *dynamic.TCPServerHealthCheck
		expNumRemovedServers  int
		expNumUpsertedServers int
		targetStatus          string
	}{
		{
			desc: "connection-only healthy server staying healthy",
			server: newTCPServer(t,
				false,
				tcpMockSequence{accept: true},
				tcpMockSequence{accept: true},
				tcpMockSequence{accept: true},
			),
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Millisecond * 50),
				Timeout:  ptypes.Duration(time.Millisecond * 40),
			},
			expNumRemovedServers:  0,
			expNumUpsertedServers: 3, // 3 health check sequences
			targetStatus:          truntime.StatusUp,
		},
		{
			desc: "connection-only healthy server becoming unhealthy",
			server: newTCPServer(t,
				false,
				tcpMockSequence{accept: true},
				tcpMockSequence{accept: false},
			),
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Millisecond * 50),
				Timeout:  ptypes.Duration(time.Millisecond * 40),
			},
			expNumRemovedServers:  1,
			expNumUpsertedServers: 1,
			targetStatus:          truntime.StatusDown,
		},
		{
			desc: "connection-only server toggling unhealthy to healthy",
			server: newTCPServer(t,
				false,
				tcpMockSequence{accept: false},
				tcpMockSequence{accept: true},
			),
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Millisecond * 50),
				Timeout:  ptypes.Duration(time.Millisecond * 40),
			},
			expNumRemovedServers:  1, // 1 failure call
			expNumUpsertedServers: 1, // 1 success call
			targetStatus:          truntime.StatusUp,
		},
		{
			desc: "connection-only server toggling healthy to unhealthy to healthy",
			server: newTCPServer(t,
				false,
				tcpMockSequence{accept: true},
				tcpMockSequence{accept: false},
				tcpMockSequence{accept: true},
			),
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Millisecond * 50),
				Timeout:  ptypes.Duration(time.Millisecond * 40),
			},
			expNumRemovedServers:  1,
			expNumUpsertedServers: 2,
			targetStatus:          truntime.StatusUp,
		},
		{
			desc: "send/expect healthy server staying healthy",
			server: newTCPServer(t,
				false,
				tcpMockSequence{accept: true, payloadIn: "PING", payloadOut: "PONG"},
				tcpMockSequence{accept: true, payloadIn: "PING", payloadOut: "PONG"},
			),
			config: &dynamic.TCPServerHealthCheck{
				Send:     "PING",
				Expect:   "PONG",
				Interval: ptypes.Duration(time.Millisecond * 50),
				Timeout:  ptypes.Duration(time.Millisecond * 40),
			},
			expNumRemovedServers:  0,
			expNumUpsertedServers: 2, // 2 successful health checks
			targetStatus:          truntime.StatusUp,
		},
		{
			desc: "send/expect server with wrong response",
			server: newTCPServer(t,
				false,
				tcpMockSequence{accept: true, payloadIn: "PING", payloadOut: "PONG"},
				tcpMockSequence{accept: true, payloadIn: "PING", payloadOut: "WRONG"},
			),
			config: &dynamic.TCPServerHealthCheck{
				Send:     "PING",
				Expect:   "PONG",
				Interval: ptypes.Duration(time.Millisecond * 50),
				Timeout:  ptypes.Duration(time.Millisecond * 40),
			},
			expNumRemovedServers:  1,
			expNumUpsertedServers: 1,
			targetStatus:          truntime.StatusDown,
		},
		{
			desc: "TLS healthy server staying healthy",
			server: newTCPServer(t,
				true,
				tcpMockSequence{accept: true, payloadIn: "HELLO", payloadOut: "WORLD"},
			),
			config: &dynamic.TCPServerHealthCheck{
				Send:     "HELLO",
				Expect:   "WORLD",
				Interval: ptypes.Duration(time.Millisecond * 500),
				Timeout:  ptypes.Duration(time.Millisecond * 2000), // Even longer timeout for TLS handshake
			},
			expNumRemovedServers:  0,
			expNumUpsertedServers: 1, // 1 TLS health check sequence
			targetStatus:          truntime.StatusUp,
		},
		{
			desc: "send-only healthcheck (no expect)",
			server: newTCPServer(t,
				false,
				tcpMockSequence{accept: true, payloadIn: "STATUS"},
				tcpMockSequence{accept: true, payloadIn: "STATUS"},
			),
			config: &dynamic.TCPServerHealthCheck{
				Send:     "STATUS",
				Interval: ptypes.Duration(time.Millisecond * 50),
				Timeout:  ptypes.Duration(time.Millisecond * 40),
			},
			expNumRemovedServers:  0,
			expNumUpsertedServers: 2,
			targetStatus:          truntime.StatusUp,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(log.Logger.WithContext(t.Context()))
			defer cancel()

			test.server.Start(t)

			dialerManager := tcp.NewDialerManager(nil)
			dialerManager.Update(map[string]*dynamic.TCPServersTransport{"default@internal": {
				TLS: &dynamic.TLSClientConfig{
					InsecureSkipVerify: true,
					ServerName:         "example.com",
				},
			}})

			dialer, err := dialerManager.Build(&dynamic.TCPServersLoadBalancer{}, test.server.TLS)
			require.NoError(t, err)

			targets := []TCPHealthCheckTarget{
				{
					Address: test.server.Addr.String(),
					TLS:     test.server.TLS,
					Dialer:  dialer,
				},
			}

			lb := &testLoadBalancer{}
			serviceInfo := &truntime.TCPServiceInfo{}

			service := NewServiceTCPHealthChecker(ctx, test.config, lb, serviceInfo, targets, "serviceName")

			go service.Launch(ctx)

			// How much time to wait for the health check to actually complete.
			deadline := time.Now().Add(200 * time.Millisecond)
			// TLS handshake can take much longer.
			if test.server.TLS {
				deadline = time.Now().Add(1000 * time.Millisecond)
			}

			// Wait for all health checks to complete deterministically
			for range test.server.StatusSequence {
				test.server.Next()

				initialUpserted := lb.numUpsertedServers
				initialRemoved := lb.numRemovedServers

				for time.Now().Before(deadline) {
					time.Sleep(5 * time.Millisecond)
					if lb.numUpsertedServers > initialUpserted || lb.numRemovedServers > initialRemoved {
						break
					}
				}
			}

			assert.Equal(t, test.expNumRemovedServers, lb.numRemovedServers, "removed servers")
			assert.Equal(t, test.expNumUpsertedServers, lb.numUpsertedServers, "upserted servers")
			assert.Equal(t, map[string]string{test.server.Addr.String(): test.targetStatus}, serviceInfo.GetAllStatus())
		})
	}
}

func TestServiceTCPHealthChecker_differentIntervals(t *testing.T) {
	// Test that unhealthy servers are checked more frequently than healthy servers
	// when UnhealthyInterval is set to a lower value than Interval
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	// Create a healthy TCP server that always accepts connections
	healthyServer := newTCPServer(t, false,
		tcpMockSequence{accept: true}, tcpMockSequence{accept: true}, tcpMockSequence{accept: true},
		tcpMockSequence{accept: true}, tcpMockSequence{accept: true},
	)
	healthyServer.Start(t)

	// Create an unhealthy TCP server that always rejects connections
	unhealthyServer := newTCPServer(t, false,
		tcpMockSequence{accept: false}, tcpMockSequence{accept: false}, tcpMockSequence{accept: false},
		tcpMockSequence{accept: false}, tcpMockSequence{accept: false}, tcpMockSequence{accept: false},
		tcpMockSequence{accept: false}, tcpMockSequence{accept: false}, tcpMockSequence{accept: false},
		tcpMockSequence{accept: false},
	)
	unhealthyServer.Start(t)

	lb := &testLoadBalancer{RWMutex: &sync.RWMutex{}}

	// Set normal interval to 500ms but unhealthy interval to 50ms
	// This means unhealthy servers should be checked 10x more frequently
	config := &dynamic.TCPServerHealthCheck{
		Interval:          ptypes.Duration(500 * time.Millisecond),
		UnhealthyInterval: pointer(ptypes.Duration(50 * time.Millisecond)),
		Timeout:           ptypes.Duration(100 * time.Millisecond),
	}

	// Set up dialer manager
	dialerManager := tcp.NewDialerManager(nil)
	dialerManager.Update(map[string]*dynamic.TCPServersTransport{
		"default@internal": {
			DialTimeout:   ptypes.Duration(100 * time.Millisecond),
			DialKeepAlive: ptypes.Duration(100 * time.Millisecond),
		},
	})

	// Get dialer for targets
	dialer, err := dialerManager.Build(&dynamic.TCPServersLoadBalancer{}, false)
	require.NoError(t, err)

	targets := []TCPHealthCheckTarget{
		{Address: healthyServer.Addr.String(), TLS: false, Dialer: dialer},
		{Address: unhealthyServer.Addr.String(), TLS: false, Dialer: dialer},
	}

	serviceInfo := &truntime.TCPServiceInfo{}
	hc := NewServiceTCPHealthChecker(ctx, config, lb, serviceInfo, targets, "test-service")

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		hc.Launch(ctx)
		wg.Done()
	}()

	// Let it run for 2 seconds to see the different check frequencies
	select {
	case <-time.After(2 * time.Second):
		cancel()
	case <-ctx.Done():
	}

	wg.Wait()

	lb.Lock()
	defer lb.Unlock()

	// The unhealthy server should be checked more frequently (50ms interval)
	// compared to the healthy server (500ms interval), so we should see
	// significantly more "removed" events than "upserted" events
	assert.Greater(t, lb.numRemovedServers, lb.numUpsertedServers, "unhealthy servers checked more frequently")
}

type tcpMockSequence struct {
	accept     bool
	payloadIn  string
	payloadOut string
}

type sequencedTCPServer struct {
	Addr           *net.TCPAddr
	StatusSequence []tcpMockSequence
	TLS            bool
	release        chan struct{}
}

func newTCPServer(t *testing.T, tlsEnabled bool, statusSequence ...tcpMockSequence) *sequencedTCPServer {
	t.Helper()

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	listener, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)

	listener.Close()

	return &sequencedTCPServer{
		Addr:           tcpAddr,
		TLS:            tlsEnabled,
		StatusSequence: statusSequence,
		release:        make(chan struct{}),
	}
}

func (s *sequencedTCPServer) Next() {
	s.release <- struct{}{}
}

func (s *sequencedTCPServer) Start(t *testing.T) {
	t.Helper()

	go func() {
		var listener net.Listener

		for _, seq := range s.StatusSequence {
			<-s.release
			if listener != nil {
				listener.Close()
			}

			if !seq.accept {
				continue
			}

			lis, err := net.ListenTCP("tcp", s.Addr)
			require.NoError(t, err)

			listener = lis

			if s.TLS {
				cert, err := tls.X509KeyPair(localhostCert, localhostKey)
				require.NoError(t, err)

				x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
				require.NoError(t, err)

				certpool := x509.NewCertPool()
				certpool.AddCert(x509Cert)

				listener = tls.NewListener(
					listener,
					&tls.Config{
						RootCAs:            certpool,
						Certificates:       []tls.Certificate{cert},
						InsecureSkipVerify: true,
						ServerName:         "example.com",
						MinVersion:         tls.VersionTLS12,
						MaxVersion:         tls.VersionTLS12,
						ClientAuth:         tls.VerifyClientCertIfGiven,
						ClientCAs:          certpool,
					},
				)
			}

			conn, err := listener.Accept()
			require.NoError(t, err)
			t.Cleanup(func() {
				_ = conn.Close()
			})

			// For TLS connections, perform handshake first
			if s.TLS {
				if tlsConn, ok := conn.(*tls.Conn); ok {
					if err := tlsConn.Handshake(); err != nil {
						continue // Skip this sequence on handshake failure
					}
				}
			}

			if seq.payloadIn == "" {
				continue
			}

			buf := make([]byte, len(seq.payloadIn))
			n, err := conn.Read(buf)
			require.NoError(t, err)

			recv := strings.TrimSpace(string(buf[:n]))

			switch recv {
			case seq.payloadIn:
				if _, err := conn.Write([]byte(seq.payloadOut)); err != nil {
					t.Errorf("failed to write payload: %v", err)
				}
			default:
				if _, err := conn.Write([]byte("FAULT\n")); err != nil {
					t.Errorf("failed to write payload: %v", err)
				}
			}
		}

		defer close(s.release)
	}()
}

type dialerMock struct {
	onDial func(network, addr string) (net.Conn, error)
}

func (dm *dialerMock) Dial(network, addr string, _ tcp.ClientConn) (net.Conn, error) {
	return dm.onDial(network, addr)
}

func (dm *dialerMock) DialContext(_ context.Context, network, addr string, _ tcp.ClientConn) (net.Conn, error) {
	return dm.onDial(network, addr)
}

func (dm *dialerMock) TerminationDelay() time.Duration {
	return 0
}

type connMock struct {
	writeFunc func([]byte) (int, error)
	readFunc  func([]byte) (int, error)
}

func (cm *connMock) Read(b []byte) (n int, err error) {
	if cm.readFunc != nil {
		return cm.readFunc(b)
	}
	return 0, nil
}

func (cm *connMock) Write(b []byte) (n int, err error) {
	if cm.writeFunc != nil {
		return cm.writeFunc(b)
	}
	return len(b), nil
}

func (cm *connMock) Close() error { return nil }

func (cm *connMock) LocalAddr() net.Addr { return &net.TCPAddr{} }

func (cm *connMock) RemoteAddr() net.Addr { return &net.TCPAddr{} }

func (cm *connMock) SetDeadline(_ time.Time) error { return nil }

func (cm *connMock) SetReadDeadline(_ time.Time) error { return nil }

func (cm *connMock) SetWriteDeadline(_ time.Time) error { return nil }
