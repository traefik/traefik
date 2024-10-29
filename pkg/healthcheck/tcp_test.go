package healthcheck

import (
	"context"
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
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

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
				tcpMockSequence{accept: true},
				tcpMockSequence{accept: true},
			),
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Millisecond * 100),
				Timeout:  ptypes.Duration(time.Millisecond * 99),
				TLS:      false,
			},
			expNumRemovedServers:  0,
			expNumUpsertedServers: 2,
			expGaugeValue:         1,
			targetStatus:          truntime.StatusUp,
		},
		{
			desc: "healthy server becoming unhealthy",
			server: newTCPServer(t,
				tcpMockSequence{accept: true},
				tcpMockSequence{accept: false},
			),
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Millisecond * 100),
				Timeout:  ptypes.Duration(time.Millisecond * 99),
				TLS:      false,
			},
			expNumRemovedServers:  1,
			expNumUpsertedServers: 1,
			expGaugeValue:         0,
			targetStatus:          truntime.StatusDown,
		},
		{
			desc: "healthy server with request and response",
			server: newTCPServer(t,
				tcpMockSequence{accept: true, payloadIn: "request", payloadOut: "response"},
				tcpMockSequence{accept: true, payloadIn: "request", payloadOut: "response"},
			),
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Millisecond * 100),
				Timeout:  ptypes.Duration(time.Millisecond * 99),
				TLS:      false,
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
				tcpMockSequence{accept: true, payloadIn: "request", payloadOut: "response"},
				tcpMockSequence{accept: true, payloadIn: "request", payloadOut: "bad response"},
			),
			config: &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Millisecond * 100),
				Timeout:  ptypes.Duration(time.Millisecond * 99),
				TLS:      false,
				Payload:  "request",
				Expected: "response",
			},
			expNumRemovedServers:  1,
			expNumUpsertedServers: 1,
			expGaugeValue:         0,
			targetStatus:          truntime.StatusDown,
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

			service := NewServiceTCPHealthChecker(&MetricsMock{gauge}, test.config, lb, serviceInfo, targets, "serviceName")

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
	release        chan struct{}
}

func newTCPServer(t *testing.T, statusSequence ...tcpMockSequence) *sequencedTcpServer {
	listener, err := net.ListenTCP("tcp", net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:0")))
	require.NoError(t, err)

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)

	listener.Close()

	return &sequencedTcpServer{
		Addr:           tcpAddr,
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

			listener, err := net.ListenTCP("tcp", s.Addr)
			require.NoError(t, err)

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
