package healthcheck

import (
	"container/list"
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

func Test_ServiceTCPHealthChecker_Launch(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc                  string
		mode                  string
		status                int
		server                *TCPServer
		expNumRemovedServers  int
		expNumUpsertedServers int
		expGaugeValue         float64
		targetStatus          string
	}{
		{
			desc:                  "healthy server staying healthy",
			server:                newTCPServer(true),
			expNumRemovedServers:  0,
			expNumUpsertedServers: 1,
			expGaugeValue:         1,
			targetStatus:          truntime.StatusUp,
		},
		{
			desc:                  "healthy server becoming unhealthy",
			server:                newTCPServer(true, false),
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

			targetURL := test.server.Start(t, cancel)

			targets := make(map[string]*net.TCPAddr)
			targets["target1"] = targetURL

			lb := &testLoadBalancer{RWMutex: &sync.RWMutex{}}
			gauge := &testhelpers.CollectingGauge{}
			serviceInfo := &truntime.TCPServiceInfo{}
			config := &dynamic.TCPServerHealthCheck{
				Interval: ptypes.Duration(time.Millisecond * 100),
				Timeout:  ptypes.Duration(time.Millisecond * 99),
				TLS:      false,
			}

			service := NewServiceTCPHealthChecker(&MetricsMock{gauge}, config, lb, serviceInfo, targets, "serviceName")

			go service.Launch(ctx)

			runtime.Gosched()

			select {
			case <-time.After(100 * time.Millisecond):
				t.Fatal("test did not complete in time")
			default:
				<-ctx.Done()
			}

			lb.RLock()
			defer lb.RUnlock()

			assert.Equal(t, test.expNumRemovedServers, lb.numRemovedServers, "removed servers")
			assert.Equal(t, test.expNumUpsertedServers, lb.numUpsertedServers, "upserted servers")
			assert.InDelta(t, test.expGaugeValue, gauge.GaugeValue, delta, "ServerUp Gauge")
			assert.Equal(t, []string{"service", "serviceName", "url", targetURL.String()}, gauge.LastLabelValues)
			assert.Equal(t, map[string]string{targetURL.String(): test.targetStatus}, serviceInfo.GetAllStatus())
		})
	}
}

type TCPServer struct {
	status *list.List
}

func newTCPServer(statusSequence ...bool) *TCPServer {
	handler := &TCPServer{
		status: list.New(),
	}

	for _, status := range statusSequence {
		handler.status.PushBack(status)
	}

	return handler
}

func (s *TCPServer) Start(t *testing.T, done context.CancelFunc) *net.TCPAddr {
	t.Helper()

	listener, err := net.ListenTCP("tcp", net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:0")))
	require.NoError(t, err)

	t.Cleanup(func() { _ = listener.Close() })

	statusSeq := s.status.Front()
	status := statusSeq.Value.(bool)
	go func() {
		for {
			if status {
				conn, err := listener.Accept()
				assert.NoError(t, err)
				t.Cleanup(func() { _ = conn.Close() })

				go func() {
					for {
						buf := make([]byte, 1024)
						n, err := conn.Read(buf)
						if err != nil {
							return
						}

						recv := strings.TrimSpace(string(buf[:n]))

						switch recv {
						case "health":
							_, _ = conn.Write([]byte("OK\n"))
						case "quit":
							done()
							_, _ = conn.Write([]byte("Bye\n"))
							return
						}
					}
				}()
			}

			statusSeq = statusSeq.Next()
			if statusSeq == nil {
				done()
				return
			}
			status = statusSeq.Value.(bool)
		}
	}()

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)

	return tcpAddr
}
