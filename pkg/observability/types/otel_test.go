package types

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"google.golang.org/grpc"
)

type traceCollector struct {
	ptraceotlp.UnimplementedGRPCServer

	called atomic.Bool
}

func (c *traceCollector) Export(context.Context, ptraceotlp.ExportRequest) (ptraceotlp.ExportResponse, error) {
	c.called.Store(true)

	return ptraceotlp.NewExportResponse(), nil
}

type logCollector struct {
	plogotlp.UnimplementedGRPCServer

	called atomic.Bool
}

func (c *logCollector) Export(context.Context, plogotlp.ExportRequest) (plogotlp.ExportResponse, error) {
	c.called.Store(true)

	return plogotlp.NewExportResponse(), nil
}

func startIPv6GRPCServer(t *testing.T, register func(*grpc.Server)) string {
	t.Helper()

	listener, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Skipf("IPv6 loopback is not available: %v", err)
	}

	server := grpc.NewServer()
	register(server)

	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	return listener.Addr().String()
}

func TestOTelTracing_setupGRPCExporter_IPv6Endpoint(t *testing.T) {
	collector := &traceCollector{}
	endpoint := startIPv6GRPCServer(t, func(server *grpc.Server) {
		ptraceotlp.RegisterGRPCServer(server, collector)
	})

	config := &OTelTracing{GRPC: &OTelGRPC{Endpoint: endpoint, Insecure: true}}
	exporter, err := config.setupGRPCExporter()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	spans := tracetest.SpanStubs{{Name: "test"}}.Snapshots()
	require.NoError(t, exporter.ExportSpans(ctx, spans))
	assert.True(t, collector.called.Load())
}

func TestOTelLog_buildGRPCExporter_IPv6Endpoint(t *testing.T) {
	collector := &logCollector{}
	endpoint := startIPv6GRPCServer(t, func(server *grpc.Server) {
		plogotlp.RegisterGRPCServer(server, collector)
	})

	config := &OTelLog{GRPC: &OTelGRPC{Endpoint: endpoint, Insecure: true}}
	exporter, err := config.buildGRPCExporter()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	require.NoError(t, exporter.Export(ctx, []sdklog.Record{{}}))
	assert.True(t, collector.called.Load())
}
