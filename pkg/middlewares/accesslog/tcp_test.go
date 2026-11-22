package accesslog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/tcp/ipallowlist"
	otypes "github.com/traefik/traefik/v3/pkg/observability/types"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

func TestTCPMiddlewareMetadataIsLogged(t *testing.T) {
	handler, buf := newTCPAccessLog(t, nil)

	service := backend(func(_ context.Context, conn tcp.WriteCloser) {
		buf := make([]byte, 4)
		_, err := conn.Read(buf)
		require.NoError(t, err)
		_, err = conn.Write([]byte("pong"))
		require.NoError(t, err)
	})

	middleware := tcp.HandlerFunc(func(ctx context.Context, conn tcp.WriteCloser) {
		metadata := tcp.Metadata(ctx)
		require.NotNil(t, metadata)
		metadata["X-Client-IP"] = "10.0.0.8"
		metadata["X-Combined-Plugin"] = "combined"
		service.ServeTCP(ctx, conn)
	})

	handler.WrapTCP("web", "sftp@file", "sftp", middleware).ServeTCP(t.Context(), newStubConn("10.0.0.8:4242", "ping"))

	got := decodeAccessLog(t, buf)
	assert.Equal(t, "TCP", got[RequestProtocol])
	assert.Equal(t, "success", got[ConnectionStatus])
	assert.Equal(t, "10.0.0.8", got["X-Client-IP"])
	assert.Equal(t, "combined", got["X-Combined-Plugin"])
	assert.Equal(t, "web", got["entryPointName"])
	assert.Equal(t, "sftp@file", got[RouterName])
	assert.Equal(t, "sftp", got[ServiceName])
	assert.Equal(t, "10.0.0.8:4242", got[ClientAddr])
	assert.Equal(t, "10.0.0.8", got[ClientHost])
	assert.Equal(t, "4242", got[ClientPort])
	assert.Equal(t, "-", got[RequestMethod])
	assert.Equal(t, "-", got[RequestPath])
	assert.InDelta(t, 4, got[RequestContentSize], 0)
	assert.InDelta(t, 4, got[DownstreamContentSize], 0)
	_, hasError := got[Error]
	assert.False(t, hasError)
}

func TestTCPMiddlewareRejectionKeepsMetadata(t *testing.T) {
	handler, buf := newTCPAccessLog(t, nil)

	middleware := tcp.HandlerFunc(func(ctx context.Context, conn tcp.WriteCloser) {
		metadata := tcp.Metadata(ctx)
		require.NotNil(t, metadata)
		metadata["X-Client-IP"] = "192.168.0.9"
		metadata["X-Combined-Plugin"] = "combined"
		tcp.SetRejection(conn, assert.AnError)
		require.NoError(t, conn.Close())
	})

	handler.WrapTCP("web", "db@file", "db", middleware).ServeTCP(t.Context(), newStubConn("192.168.0.9:9", ""))

	got := decodeAccessLog(t, buf)
	assert.Equal(t, "rejected", got[ConnectionStatus])
	assert.Equal(t, assert.AnError.Error(), got[Error])
	assert.Equal(t, "192.168.0.9", got["X-Client-IP"])
	assert.Equal(t, "combined", got["X-Combined-Plugin"])
	assert.Equal(t, "TCP", got[RequestProtocol])
}

func TestTCPMiddlewareCloseWithoutReasonIsRejected(t *testing.T) {
	handler, buf := newTCPAccessLog(t, nil)

	middleware := tcp.HandlerFunc(func(ctx context.Context, conn tcp.WriteCloser) {
		tcp.Metadata(ctx)["X-Client-IP"] = "192.168.0.9"
		require.NoError(t, conn.Close())
	})

	handler.WrapTCP("web", "db@file", "db", middleware).ServeTCP(t.Context(), newStubConn("192.168.0.9:9", ""))

	got := decodeAccessLog(t, buf)
	assert.Equal(t, "rejected", got[ConnectionStatus])
	assert.Equal(t, tcpClosedBeforeService, got[Error])
	assert.Equal(t, "192.168.0.9", got["X-Client-IP"])
}

func TestTCPIPAllowListRejectionIsLogged(t *testing.T) {
	handler, buf := newTCPAccessLog(t, nil)

	allow, err := ipallowlist.New(t.Context(), tcp.HandlerFunc(func(context.Context, tcp.WriteCloser) {
		t.Fatal("service must not be reached")
	}), dynamic.TCPIPAllowList{SourceRange: []string{"10.0.0.0/8"}}, "allow@file")
	require.NoError(t, err)

	middleware := tcp.HandlerFunc(func(ctx context.Context, conn tcp.WriteCloser) {
		tcp.Metadata(ctx)["X-Client-IP"] = "192.168.1.1"
		allow.ServeTCP(ctx, conn)
	})

	handler.WrapTCP("web", "db@file", "db", middleware).ServeTCP(t.Context(), newStubConn("192.168.1.1:9", ""))

	got := decodeAccessLog(t, buf)
	assert.Equal(t, "rejected", got[ConnectionStatus])
	assert.Contains(t, got[Error], "matched none of the trusted IPs")
	assert.Equal(t, "192.168.1.1", got["X-Client-IP"])
}

func TestTCPAccessLogFilters(t *testing.T) {
	t.Run("status codes do not drop tcp lines", func(t *testing.T) {
		handler, buf := newTCPAccessLog(t, &otypes.AccessLogFilters{StatusCodes: []string{"500"}})
		handler.WrapTCP("web", "db@file", "db", backend(nil)).
			ServeTCP(t.Context(), newStubConn("10.0.0.1:1", ""))

		got := decodeAccessLog(t, buf)
		assert.Equal(t, "success", got[ConnectionStatus])
	})

	t.Run("min duration drops short tcp lines", func(t *testing.T) {
		handler, buf := newTCPAccessLog(t, &otypes.AccessLogFilters{MinDuration: ptypes.Duration(time.Hour)})
		handler.WrapTCP("web", "db@file", "db", backend(nil)).
			ServeTCP(t.Context(), newStubConn("10.0.0.1:1", ""))

		assert.Empty(t, strings.TrimSpace(buf.String()))
	})

	t.Run("dropped metadata field is omitted", func(t *testing.T) {
		handler, buf := newTCPAccessLog(t, nil)
		handler.config.Fields.Names = map[string]string{"x-client-ip": otypes.AccessLogDrop}

		middleware := tcp.HandlerFunc(func(ctx context.Context, conn tcp.WriteCloser) {
			tcp.Metadata(ctx)["X-Client-IP"] = "10.0.0.8"
			tcp.Metadata(ctx)["X-Combined-Plugin"] = "combined"
		})
		handler.WrapTCP("web", "db@file", "db", middleware).ServeTCP(t.Context(), newStubConn("10.0.0.8:1", ""))

		got := decodeAccessLog(t, buf)
		_, dropped := got["X-Client-IP"]
		assert.False(t, dropped)
		assert.Equal(t, "combined", got["X-Combined-Plugin"])
		assert.Equal(t, "rejected", got[ConnectionStatus])
	})
}

func TestTCPBackendFailureIsRejected(t *testing.T) {
	t.Run("dial error", func(t *testing.T) {
		handler, buf := newTCPAccessLog(t, nil)

		proxy, err := tcp.NewProxy("10.0.0.1:5432", failingDialer{})
		require.NoError(t, err)

		handler.WrapTCP("web", "db@file", "db@file", proxy).ServeTCP(t.Context(), newStubConn("10.0.0.8:1", ""))

		got := decodeAccessLog(t, buf)
		assert.Equal(t, "rejected", got[ConnectionStatus])
		assert.Equal(t, "dialing backend: connection refused", got[Error])
	})

	t.Run("empty pool", func(t *testing.T) {
		handler, buf := newTCPAccessLog(t, nil)

		handler.WrapTCP("web", "db@file", "db@file", tcp.NewWRRLoadBalancer(false)).ServeTCP(t.Context(), newStubConn("10.0.0.8:1", ""))

		got := decodeAccessLog(t, buf)
		assert.Equal(t, "rejected", got[ConnectionStatus])
		assert.Equal(t, "no servers in the pool", got[Error])
	})
}

func TestTCPMetadataCannotOverrideLogFields(t *testing.T) {
	handler, buf := newTCPAccessLog(t, nil)

	service := backend(func(ctx context.Context, _ tcp.WriteCloser) {
		metadata := tcp.Metadata(ctx)
		metadata[DownstreamStatus] = "200"
		metadata[StartLocal] = "fake"
		metadata[Error] = "fake"
		metadata["request_User-Agent"] = "fake"
		metadata["X-Client-IP"] = "10.0.0.8"
	})

	handler.WrapTCP("web", "db@file", "db@file", service).ServeTCP(t.Context(), newStubConn("10.0.0.8:1", ""))

	got := decodeAccessLog(t, buf)
	assert.Equal(t, "success", got[ConnectionStatus])
	assert.NotContains(t, got, DownstreamStatus)
	assert.NotContains(t, got, StartLocal)
	assert.NotContains(t, got, Error)
	assert.NotContains(t, got, "request_User-Agent")
	assert.Equal(t, "10.0.0.8", got["X-Client-IP"])
}

func TestTCPMetadataDoesNotBreakCommonFormat(t *testing.T) {
	handler, buf := newTCPAccessLog(t, nil)
	handler.config.Format = CommonFormat
	handler.logger.SetFormatter(new(CommonLogFormatter))
	handler.config.Fields.Names = map[string]string{"startutc": otypes.AccessLogDrop}

	service := backend(func(ctx context.Context, _ tcp.WriteCloser) {
		tcp.Metadata(ctx)[StartLocal] = "fake"
	})

	assert.NotPanics(t, func() {
		handler.WrapTCP("web", "db@file", "db@file", service).ServeTCP(t.Context(), newStubConn("10.0.0.8:1", ""))
	})
	assert.NotEmpty(t, strings.TrimSpace(buf.String()))
}

// backend stands for a TCP service that reached its server, as tcp.Proxy reports it after dialing.
func backend(serve func(ctx context.Context, conn tcp.WriteCloser)) tcp.Handler {
	return tcp.HandlerFunc(func(ctx context.Context, conn tcp.WriteCloser) {
		tcp.MarkBackend(conn)
		if serve != nil {
			serve(ctx, conn)
		}
	})
}

type failingDialer struct{}

func (failingDialer) Dial(string, string, tcp.ClientConn) (net.Conn, error) {
	return nil, errors.New("connection refused")
}

func (failingDialer) DialContext(context.Context, string, string, tcp.ClientConn) (net.Conn, error) {
	return nil, errors.New("connection refused")
}

func (failingDialer) TerminationDelay() time.Duration { return 0 }

func newTCPAccessLog(t *testing.T, filters *otypes.AccessLogFilters) (*Handler, *bytes.Buffer) {
	t.Helper()

	fields := &otypes.AccessLogFields{}
	fields.SetDefaults()

	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	return &Handler{
		config: &otypes.AccessLog{
			Format:  JSONFormat,
			Fields:  fields,
			Filters: filters,
		},
		logger: logger,
	}, &buf
}

func decodeAccessLog(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	line := strings.TrimSpace(buf.String())
	require.NotEmpty(t, line)

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(line), &got))
	return got
}

type stubAddr string

func (a stubAddr) Network() string { return "tcp" }
func (a stubAddr) String() string  { return string(a) }

type stubConn struct {
	net.Conn

	incoming *bytes.Reader
}

func newStubConn(remote, payload string) *stubConn {
	return &stubConn{incoming: bytes.NewReader([]byte(payload)), Conn: stubNetConn{remote: stubAddr(remote)}}
}

func (c *stubConn) Read(p []byte) (int, error) {
	return c.incoming.Read(p)
}

func (c *stubConn) CloseWrite() error { return nil }

type stubNetConn struct {
	remote net.Addr
}

func (c stubNetConn) Read([]byte) (int, error)         { return 0, nil }
func (c stubNetConn) Write(p []byte) (int, error)      { return len(p), nil }
func (c stubNetConn) Close() error                     { return nil }
func (c stubNetConn) CloseWrite() error                { return nil }
func (c stubNetConn) LocalAddr() net.Addr              { return stubAddr("127.0.0.1:443") }
func (c stubNetConn) RemoteAddr() net.Addr             { return c.remote }
func (c stubNetConn) SetDeadline(time.Time) error      { return nil }
func (c stubNetConn) SetReadDeadline(time.Time) error  { return nil }
func (c stubNetConn) SetWriteDeadline(time.Time) error { return nil }
