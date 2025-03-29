package healthcheck

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"runtime"
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
	ttcp "github.com/traefik/traefik/v3/pkg/tcp"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

var LocalhostCert = []byte(`-----BEGIN CERTIFICATE-----
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

var LocalhostKey = []byte(`-----BEGIN PRIVATE KEY-----
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

//	openssl req -newkey rsa:2048 \
//	   -new -nodes -x509 \
//	   -days 3650 \
//	   -out cert.pem \
//	   -keyout key.pem \
//	   -subj "/CN=example.com"
//	   -addext "subjectAltName = DNS:example.com"
var mTLSCert = []byte(`-----BEGIN CERTIFICATE-----
MIIDJTCCAg2gAwIBAgIUYKnGcLnmMosOSKqTn4ydAMURE4gwDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAwwLZXhhbXBsZS5jb20wHhcNMjAwODEzMDkyNzIwWhcNMzAw
ODExMDkyNzIwWjAWMRQwEgYDVQQDDAtleGFtcGxlLmNvbTCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAOAe+QM1c9lZ2TPRgoiuPAq2A3Pfu+i82lmqrTJ0
PR2Cx1fPbccCUTFJPlxSDiaMrwtvqw1yP9L2Pu/vJK5BY4YDVDtFGKjpRBau1otJ
iY50O5qMo3sfLqR4/1VsQGlLVZYLD3dyc4ZTmOp8+7tJ2SyGorojbIKfimZT7XD7
dzrVr4h4Gn+SzzOnoKyx29uaNRP+XuMYHmHyQcJE03pUGhkTOvMwBlF96QdQ9WG0
D+1CxRciEsZXE+imKBHoaTgrTkpnFHzsrIEw+OHQYf30zuT/k/lkgv1vqEwINHjz
W2VeTur5eqVvA7zZdoEXMRy7BUvh/nZk5AXkXAmZLn0eUg8CAwEAAaNrMGkwHQYD
VR0OBBYEFEDrbhPDt+hi3ZOzk6S/CFAVHwk0MB8GA1UdIwQYMBaAFEDrbhPDt+hi
3ZOzk6S/CFAVHwk0MA8GA1UdEwEB/wQFMAMBAf8wFgYDVR0RBA8wDYILZXhhbXBs
ZS5jb20wDQYJKoZIhvcNAQELBQADggEBAG/JRJWeUNx2mDJAk8W7Syq3gmQB7s9f
+yY/XVRJZGahOPilABqFpC6GVn2HWuvuOqy8/RGk9ja5abKVXqE6YKrljqo3XfzB
KQcOz4SFirpkHvNCiEcK3kggN3wJWqL2QyXAxWldBBBCO9yx7a3cux31C//sTUOG
xq4JZDg171U1UOpfN1t0BFMdt05XZFEM247N7Dcf7HoXwAa7eyLKgtKWqPDqGrFa
fvGDDKK9X/KVsU2x9V3pG+LsJg7ogUnSyD2r5G1F3Y8OVs2T/783PaN0M35fDL38
09VbsxA2GasOHZrghUzT4UvZWWZbWEmG975hFYvdj6DlK9K0s5TdKIs=
-----END CERTIFICATE-----`)

var mTLSKey = []byte(`-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDgHvkDNXPZWdkz
0YKIrjwKtgNz37vovNpZqq0ydD0dgsdXz23HAlExST5cUg4mjK8Lb6sNcj/S9j7v
7ySuQWOGA1Q7RRio6UQWrtaLSYmOdDuajKN7Hy6keP9VbEBpS1WWCw93cnOGU5jq
fPu7SdkshqK6I2yCn4pmU+1w+3c61a+IeBp/ks8zp6CssdvbmjUT/l7jGB5h8kHC
RNN6VBoZEzrzMAZRfekHUPVhtA/tQsUXIhLGVxPopigR6Gk4K05KZxR87KyBMPjh
0GH99M7k/5P5ZIL9b6hMCDR481tlXk7q+XqlbwO82XaBFzEcuwVL4f52ZOQF5FwJ
mS59HlIPAgMBAAECggEAAKLV3hZ2v7UrkqQTlMO50+X0WI3YAK8Yh4yedTgzPDQ0
0KD8FMaC6HrmvGhXNfDMRmIIwD8Ew1qDjzbEieIRoD2+LXTivwf6c34HidmplEfs
K2IezKin/zuArgNio2ndUlGxt4sRnN373x5/sGZjQWcYayLSmgRN5kByuhFco0Qa
oSrXcXNUlb+KgRQXPDU4+M35tPHvLdyg+tko/m/5uK9dc9MNvGZHOMBKg0VNURJb
V1l3dR+evwvpqHzBvWiqN/YOiUUvIxlFKA35hJkfCl7ivFs4CLqqFNCKDao95fWe
s0UR9iMakT48jXV76IfwZnyX10OhIWzKls5trjDL8QKBgQD3thQJ8e0FL9y1W+Ph
mCdEaoffSPkgSn64wIsQ9bMmv4y+KYBK5AhqaHgYm4LgW4x1+CURNFu+YFEyaNNA
kNCXFyRX3Em3vxcShP5jIqg+f07mtXPKntWP/zBeKQWgdHX371oFTfaAlNuKX/7S
n0jBYjr4Iof1bnquMQvUoHCYWwKBgQDnntFU9/AQGaQIvhfeU1XKFkQ/BfhIsd27
RlmiCi0ee9Ce74cMAhWr/9yg0XUxzrh+Ui1xnkMVTZ5P8tWIxROokznLUTGJA5rs
zB+ovCPFZcquTwNzn7SBnpHTR0OqJd8sd89P5ST2SqufeSF/gGi5sTs4EocOLCpZ
EPVIfm47XQKBgB4d5RHQeCDJUOw739jtxthqm1pqZN+oLwAHaOEG/mEXqOT15sM0
NlG5oeBcB+1/M/Sj1t3gn8blrvmSBR00fifgiGqmPdA5S3TU9pjW/d2bXNxv80QP
S6fWPusz0ZtQjYc3cppygCXh808/nJu/AfmBF+pTSHRumjvTery/RPFBAoGBAMi/
zCta4cTylEvHhqR5kiefePMu120aTFYeuV1KeKStJ7o5XNE5lVMIZk80e+D5jMpf
q2eIhhgWuBoPHKh4N3uqbzMbYlWgvEx09xOmTVKv0SWW8iTqzOZza2y1nZ4BSRcf
mJ1ku86EFZAYysHZp+saA3usA0ZzXRjpK87zVdM5AoGBAOSqI+t48PnPtaUDFdpd
taNNVDbcecJatm3w8VDWnarahfWe66FIqc9wUkqekqAgwZLa0AGdUalvXfGrHfNs
PtvuNc5EImfSkuPBYLBslNxtjbBvAYgacEdY+gRhn2TeIUApnND58lCWsKbNHLFZ
ajIPbTY+Fe9OTOFTN48ujXNn
-----END PRIVATE KEY-----`)

func Test_ServiceTCPHealthChecker_Check(t *testing.T) {
	testCases := []struct {
		desc                  string
		server                *sequencedTcpServer
		config                *dynamic.TCPServerHealthCheck
		expNumRemovedServers  int
		expNumUpsertedServers int
		expGaugeValue         float64
		targetStatus          string
	}{
		{
			desc: "healthy server staying healthy",
			server: newTCPServer(t,
				false,
				tcpMockSequence{accept: true},
				tcpMockSequence{accept: true},
			),
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Millisecond * 100),
				Timeout:  ptypes.Duration(time.Millisecond * 99),
			},
			expNumRemovedServers:  0,
			expNumUpsertedServers: 2,
			expGaugeValue:         1,
			targetStatus:          truntime.StatusUp,
		},
		{
			desc: "healthy server becoming unhealthy",
			server: newTCPServer(t,
				false,
				tcpMockSequence{accept: true},
				tcpMockSequence{accept: false},
			),
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Millisecond * 100),
				Timeout:  ptypes.Duration(time.Millisecond * 99),
			},
			expNumRemovedServers:  1,
			expNumUpsertedServers: 1,
			expGaugeValue:         0,
			targetStatus:          truntime.StatusDown,
		},
		{
			desc: "healthy server with request and response",
			server: newTCPServer(t,
				false,
				tcpMockSequence{accept: true, payloadIn: "request", payloadOut: "response"},
				tcpMockSequence{accept: true, payloadIn: "request", payloadOut: "response"},
			),
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Millisecond * 100),
				Timeout:  ptypes.Duration(time.Millisecond * 99),
				Payload:  "request",
				Expected: "response",
			},
			expNumRemovedServers:  0,
			expNumUpsertedServers: 2,
			expGaugeValue:         1,
			targetStatus:          truntime.StatusUp,
		},
		{
			desc: "healthy server with request and response becoming unhealthy",
			server: newTCPServer(t,
				false,
				tcpMockSequence{accept: true, payloadIn: "request", payloadOut: "response"},
				tcpMockSequence{accept: true, payloadIn: "request", payloadOut: "bad response"},
			),
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Millisecond * 100),
				Timeout:  ptypes.Duration(time.Millisecond * 99),
				Payload:  "request",
				Expected: "response",
			},
			expNumRemovedServers:  1,
			expNumUpsertedServers: 1,
			expGaugeValue:         0,
			targetStatus:          truntime.StatusDown,
		},
		{
			desc: "healthy server with TLS certificate",
			server: newTCPServer(t,
				true,
				tcpMockSequence{accept: true, payloadIn: "request", payloadOut: "response"},
				tcpMockSequence{accept: true, payloadIn: "request", payloadOut: "response"},
			),
			config: &dynamic.TCPServerHealthCheck{
				Payload:  "request",
				Expected: "response",
				Interval: ptypes.Duration(time.Millisecond * 100),
				Timeout:  ptypes.Duration(time.Millisecond * 99),
				TLS:      true,
			},
			expNumRemovedServers:  0,
			expNumUpsertedServers: 2,
			expGaugeValue:         1,
			targetStatus:          truntime.StatusUp,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			c, cancel := context.WithCancel(context.Background())
			ctx := log.Logger.WithContext(c)
			defer t.Cleanup(cancel)

			test.server.Start(t)

			targets := make(map[string]*net.TCPAddr)
			targets["target1"] = test.server.Addr

			lb := &testLoadBalancer{RWMutex: &sync.RWMutex{}}
			gauge := &testhelpers.CollectingGauge{}
			serviceInfo := &truntime.TCPServiceInfo{}

			dialerManager := ttcp.NewDialerManager(nil)
			dialerManager.Update(map[string]*dynamic.TCPServersTransport{"default@internal": {
				TLS: &dynamic.TLSClientConfig{
					InsecureSkipVerify: true,
					ServerName:         "example.com",
				},
			}})
			service := NewServiceTCPHealthChecker(dialerManager, &MetricsMock{gauge}, test.config, lb, serviceInfo, targets, "serviceName")

			for range test.server.StatusSequence {
				test.server.Next()
				runtime.Gosched()
				service.Check(ctx)
			}

			lb.RLock()
			defer lb.RUnlock()

			assert.Equal(t, test.expNumRemovedServers, lb.numRemovedServers, "removed servers")
			assert.Equal(t, test.expNumUpsertedServers, lb.numUpsertedServers, "upserted servers")
			assert.InDelta(t, test.expGaugeValue, gauge.GaugeValue, delta, "ServerUp Gauge")
			assert.Equal(t, []string{"service", "serviceName", "url", test.server.Addr.String()}, gauge.LastLabelValues)
			assert.Equal(t, map[string]string{test.server.Addr.String(): test.targetStatus}, serviceInfo.GetAllStatus())
		})
	}
}

type tcpMockSequence struct {
	accept     bool
	payloadIn  string
	payloadOut string
}

type sequencedTcpServer struct {
	Addr           *net.TCPAddr
	StatusSequence []tcpMockSequence
	TLS            bool
	release        chan struct{}
}

func newTCPServer(t *testing.T, tlsEnabled bool, statusSequence ...tcpMockSequence) *sequencedTcpServer {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	listener, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)

	listener.Close()

	return &sequencedTcpServer{
		Addr:           tcpAddr,
		TLS:            tlsEnabled,
		StatusSequence: statusSequence,
		release:        make(chan struct{}),
	}
}

func (s *sequencedTcpServer) Next() {
	s.release <- struct{}{}
}

func (s *sequencedTcpServer) Start(t *testing.T) {
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
				cert, err := tls.X509KeyPair(LocalhostCert, LocalhostKey)
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
