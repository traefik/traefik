package healthcheck

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	gokitmetrics "github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type StartTestServer interface {
	Start(t *testing.T, done func()) (*url.URL, time.Duration)
}

type Status interface {
	~int | ~int32
}

type HealthSequence[T Status] struct {
	sequenceMu sync.Mutex
	sequence   []T
}

func (s *HealthSequence[T]) Pop() T {
	s.sequenceMu.Lock()
	defer s.sequenceMu.Unlock()

	stat := s.sequence[0]

	s.sequence = s.sequence[1:]

	return stat
}

func (s *HealthSequence[T]) IsEmpty() bool {
	s.sequenceMu.Lock()
	defer s.sequenceMu.Unlock()

	return len(s.sequence) == 0
}

type GRPCServer struct {
	status HealthSequence[healthpb.HealthCheckResponse_ServingStatus]
	done   func()
}

func newGRPCServer(healthSequence ...healthpb.HealthCheckResponse_ServingStatus) *GRPCServer {
	gRPCService := &GRPCServer{
		status: HealthSequence[healthpb.HealthCheckResponse_ServingStatus]{
			sequence: healthSequence,
		},
	}

	return gRPCService
}

func (s *GRPCServer) List(_ context.Context, _ *healthpb.HealthListRequest) (*healthpb.HealthListResponse, error) {
	return nil, nil
}

func (s *GRPCServer) Check(_ context.Context, _ *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	if s.status.IsEmpty() {
		s.done()
		return &healthpb.HealthCheckResponse{
			Status: healthpb.HealthCheckResponse_SERVICE_UNKNOWN,
		}, nil
	}
	stat := s.status.Pop()

	return &healthpb.HealthCheckResponse{
		Status: stat,
	}, nil
}

func (s *GRPCServer) Watch(_ *healthpb.HealthCheckRequest, server healthpb.Health_WatchServer) error {
	if s.status.IsEmpty() {
		s.done()
		return server.Send(&healthpb.HealthCheckResponse{
			Status: healthpb.HealthCheckResponse_SERVICE_UNKNOWN,
		})
	}
	stat := s.status.Pop()

	return server.Send(&healthpb.HealthCheckResponse{
		Status: stat,
	})
}

func (s *GRPCServer) Start(t *testing.T, done func()) (*url.URL, time.Duration) {
	t.Helper()

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	assert.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	server := grpc.NewServer()
	t.Cleanup(server.Stop)

	s.done = done

	healthpb.RegisterHealthServer(server, s)

	go func() {
		err := server.Serve(listener)
		assert.NoError(t, err)
	}()

	// Make test timeout dependent on number of expected requests, health check interval, and a safety margin.
	return testhelpers.MustParseURL("http://" + listener.Addr().String()), time.Duration(len(s.status.sequence)*int(dynamic.DefaultHealthCheckInterval) + 500)
}

type HTTPServer struct {
	status HealthSequence[int]
	done   func()
}

func newHTTPServer(healthSequence ...int) *HTTPServer {
	handler := &HTTPServer{
		status: HealthSequence[int]{
			sequence: healthSequence,
		},
	}

	return handler
}

// ServeHTTP returns HTTP response codes following a status sequences.
// It calls the given 'done' function once all request health indicators have been depleted.
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	if s.status.IsEmpty() {
		s.done()
		// This ensures that the health-checker will handle the context cancellation error before receiving the HTTP response.
		time.Sleep(500 * time.Millisecond)
		return
	}

	stat := s.status.Pop()

	w.WriteHeader(stat)
}

func (s *HTTPServer) Start(t *testing.T, done func()) (*url.URL, time.Duration) {
	t.Helper()

	s.done = done

	ts := httptest.NewServer(s)
	t.Cleanup(ts.Close)

	// Make test timeout dependent on number of expected requests, health check interval, and a safety margin.
	return testhelpers.MustParseURL(ts.URL), time.Duration(len(s.status.sequence)*int(dynamic.DefaultHealthCheckInterval) + 500)
}

type testLoadBalancer struct {
	// RWMutex needed due to parallel test execution: Both the system-under-test
	// and the test assertions reference the counters.
	*sync.RWMutex
	numRemovedServers  int
	numUpsertedServers int
}

func (lb *testLoadBalancer) SetStatus(ctx context.Context, childName string, up bool) {
	if up {
		lb.numUpsertedServers++
	} else {
		lb.numRemovedServers++
	}
}

type MetricsMock struct {
	Gauge gokitmetrics.Gauge
}

func (m *MetricsMock) ServiceServerUpGauge() gokitmetrics.Gauge {
	return m.Gauge
}
