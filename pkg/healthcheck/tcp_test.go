package healthcheck

import (
	"context"
	"crypto/tls"
	"net"
	"net/netip"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	truntime "github.com/traefik/traefik/v3/pkg/config/runtime"
	ttcp "github.com/traefik/traefik/v3/pkg/tcp"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

var LocalhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIICDDCCAXWgAwIBAgIQH20JmcOlcRWHNuf62SYwszANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQC0qINy3F4oq6viDnlpDDE5J08iSRGggg6EylJKBKZfphEG2ufgK78Dufl3
+7b0LlEY2AeZHwviHODqC9a6ihj1ZYQk0/djAh+OeOhFEWu+9T/VP8gVFarFqT8D
Opy+hrG7YJivUIzwb4fmJQRI7FajzsnGyM6LiXLU+0qzb7ZO/QIDAQABo2EwXzAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAnBgNVHREEIDAeggtleGFtcGxlLmNvbYIPd3d3LmV4YW1wbGUuY29tMA0G
CSqGSIb3DQEBCwUAA4GBAB+eluoQYzyyMfeEEAOtlldevx5MtDENT05NB0WI+91R
we7mX8lv763u0XuCWPxbHszhclI6FFjoQef0Z1NYLRm8ZRq58QqWDFZ3E6wdDK+B
+OWvkW+hRavo6R9LzIZPfbv8yBo4M9PK/DXw8hLqH7VkkI+Gh793iH7Ugd4A7wvT
-----END CERTIFICATE-----`)

// LocalhostKey is the private key for localhostCert.
var LocalhostKey = []byte(`-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBALSog3LcXiirq+IO
eWkMMTknTyJJEaCCDoTKUkoEpl+mEQba5+ArvwO5+Xf7tvQuURjYB5kfC+Ic4OoL
1rqKGPVlhCTT92MCH4546EURa771P9U/yBUVqsWpPwM6nL6GsbtgmK9QjPBvh+Yl
BEjsVqPOycbIzouJctT7SrNvtk79AgMBAAECgYB1wMT1MBgbkFIXpXGTfAP1id61
rUTVBxCpkypx3ngHLjo46qRq5Hi72BN4FlTY8fugIudI8giP2FztkMvkiLDc4m0p
Gn+QMJzjlBjjTuNLvLy4aSmNRLIC3mtbx9PdU71DQswEpJHFj/vmsxbuSrG1I1YE
r1reuSo2ow6fOAjXLQJBANpz+RkOiPSPuvl+gi1sp2pLuynUJVDVqWZi386YRpfg
DiKCLpqwqYDkOozm/fwFALvwXKGmsyyL43HO8eI+2NsCQQDTtY32V+02GPecdsyq
msK06EPVTSaYwj9Mm+q709KsmYFHLXDqXjcKV4UgKYKRPz7my1fXodMmGmfuh1a3
/HMHAkEAmOQKN0tA90mRJwUvvvMIyRBv0fq0kzq28P3KfiF9ZtZdjjFmxMVYHOmf
QPZ6VGR7+w1jB5BQXqEZcpHQIPSzeQJBAIy9tZJ/AYNlNbcegxEnsSjy/6VdlLsY
51vWi0Yym2uC4R6gZuBnoc+OP0ISVmqY0Qg9RjhjrCs4gr9f2ZaWjSECQCxqZMq1
3viJ8BGCC0m/5jv1EHur3YgwphYCkf4Li6DKwIdMLk1WXkTcPIY3V2Jqj8rPEB5V
rqPRSAtd/h6oZbs=
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
	t.Parallel()

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
				tcpMockSequence{accept: true},
				tcpMockSequence{accept: true},
			),
			config: &dynamic.TCPServerHealthCheck{
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
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
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
	listener, err := net.ListenTCP("tcp", net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:0")))
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
		for _, seq := range s.StatusSequence {
			<-s.release

			if !seq.accept {
				continue
			}

			var listener net.Listener

			listener, err := net.ListenTCP("tcp", s.Addr)
			require.NoError(t, err)

			if s.TLS {
				cert, err := tls.X509KeyPair(LocalhostCert, LocalhostKey)
				require.NoError(t, err)

				listener = tls.NewListener(
					listener,
					&tls.Config{
						Certificates: []tls.Certificate{cert},
					},
				)
			}

			conn, err := listener.Accept()
			require.NoError(t, err)

			listener.Close()

			if seq.payloadIn == "" {
				conn.Close()

				continue
			}

			conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			buf := make([]byte, 1024)
			n, _ := conn.Read(buf)

			recv := strings.TrimSpace(string(buf[:n]))

			switch recv {
			case seq.payloadIn:
				_, _ = conn.Write([]byte(seq.payloadOut))
			default:
				_, _ = conn.Write([]byte("FAULT\n"))
			}

			defer conn.Close()
		}

		close(s.release)
	}()
}
