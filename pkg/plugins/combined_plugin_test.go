package plugins

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

func newTestYaegiBuilder(t *testing.T, pluginImport string) *yaegiMiddlewareBuilder {
	t.Helper()

	interpreter, err := newInterpreter(t.Context(), "fixtures", &Manifest{Import: pluginImport}, Settings{})
	require.NoError(t, err)

	builder, err := newYaegiMiddlewareBuilder(interpreter, "", pluginImport)
	require.NoError(t, err)

	return builder
}

func TestCombinedPlugin_HTTP(t *testing.T) {
	builder := newTestYaegiBuilder(t, "testplugincombined")

	testCases := []struct {
		desc           string
		prefix         string
		remoteAddr     string
		expectedStatus int
		expectedHeader string
		expectedIP     string
	}{
		{
			desc:           "allowed IP",
			prefix:         "127",
			remoteAddr:     "127.0.0.1:1234",
			expectedStatus: http.StatusOK,
			expectedHeader: "combined",
			expectedIP:     "127.0.0.1",
		},
		{
			desc:           "blocked IP",
			prefix:         "127",
			remoteAddr:     "192.168.1.1:1234",
			expectedStatus: http.StatusForbidden,
		},
		{
			desc:           "configured prefix",
			prefix:         "10",
			remoteAddr:     "10.0.0.1:1234",
			expectedStatus: http.StatusOK,
			expectedHeader: "combined",
			expectedIP:     "10.0.0.1",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			middleware, err := builder.newMiddleware(map[string]any{
				"allowedIPPrefix": test.prefix,
				"httpHeaderValue": "combined",
			}, "test")
			require.NoError(t, err)

			handler, err := middleware.NewHandler(t.Context(), http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.WriteHeader(http.StatusOK)
			}))
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			req.RemoteAddr = test.remoteAddr
			rw := httptest.NewRecorder()

			handler.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatus, rw.Code)
			assert.Equal(t, test.expectedHeader, rw.Header().Get("X-Combined-Plugin"))
			assert.Equal(t, test.expectedIP, rw.Header().Get("X-Client-IP"))
		})
	}
}

func TestCombinedPlugin_TCP(t *testing.T) {
	builder := newTestYaegiBuilder(t, "testplugincombined")

	testCases := []struct {
		desc             string
		prefix           string
		remoteAddr       string
		expectedNext     bool
		expectedMetadata map[string]string
	}{
		{
			desc:         "allowed IP",
			prefix:       "127",
			remoteAddr:   "127.0.0.1:1234",
			expectedNext: true,
			expectedMetadata: map[string]string{
				"X-Combined-Plugin": "combined",
				"X-Client-IP":       "127.0.0.1",
			},
		},
		{
			desc:             "blocked IP",
			prefix:           "127",
			remoteAddr:       "192.168.1.1:1234",
			expectedMetadata: map[string]string{},
		},
		{
			desc:         "configured prefix",
			prefix:       "10",
			remoteAddr:   "10.0.0.1:1234",
			expectedNext: true,
			expectedMetadata: map[string]string{
				"X-Combined-Plugin": "combined",
				"X-Client-IP":       "10.0.0.1",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			// Going through Builder.BuildTCP covers the plugin type assertion that the runtime relies on.
			pluginBuilder := Builder{middlewareBuilders: map[string]middlewareBuilder{"combined": builder}}

			constructor, err := pluginBuilder.BuildTCP("combined", map[string]any{
				"allowedIPPrefix": test.prefix,
				"httpHeaderValue": "combined",
			}, "test")
			require.NoError(t, err)

			var nextCalled bool
			var nextMetadata map[string]string
			handler, err := constructor(tcp.HandlerFunc(func(ctx context.Context, _ tcp.WriteCloser) {
				nextCalled = true
				nextMetadata = tcp.Metadata(ctx)
			}))
			require.NoError(t, err)

			ctx := tcp.WithMetadata(t.Context())
			conn := &testTCPConn{remoteAddr: test.remoteAddr}

			handler.ServeTCP(ctx, conn)

			assert.Equal(t, test.expectedNext, nextCalled)
			assert.Equal(t, !test.expectedNext, conn.closed)
			assert.Equal(t, test.expectedMetadata, tcp.Metadata(ctx))
			if test.expectedNext {
				assert.Equal(t, test.expectedMetadata, nextMetadata)
			}
		})
	}
}

func TestBuildTCP_pluginWithoutTCP(t *testing.T) {
	pluginBuilder := Builder{middlewareBuilders: map[string]middlewareBuilder{"safe": newTestYaegiBuilder(t, "testpluginsafe")}}

	_, err := pluginBuilder.BuildTCP("safe", nil, "test")
	require.ErrorContains(t, err, "does not support TCP")
}

func TestBuildTCP_invalidNewTCPSignature(t *testing.T) {
	pluginBuilder := Builder{middlewareBuilders: map[string]middlewareBuilder{"invalid": newTestYaegiBuilder(t, "testplugintcpinvalid")}}

	_, err := pluginBuilder.BuildTCP("invalid", nil, "test")
	require.ErrorContains(t, err, "invalid NewTCP signature")
}

type testTCPConn struct {
	remoteAddr string
	closed     bool
}

func (t *testTCPConn) Read(_ []byte) (int, error)         { return 0, nil }
func (t *testTCPConn) Write(b []byte) (int, error)        { return len(b), nil }
func (t *testTCPConn) Close() error                       { t.closed = true; return nil }
func (t *testTCPConn) CloseWrite() error                  { return nil }
func (t *testTCPConn) LocalAddr() net.Addr                { return &testAddr{addr: "0.0.0.0:0"} }
func (t *testTCPConn) RemoteAddr() net.Addr               { return &testAddr{addr: t.remoteAddr} }
func (t *testTCPConn) SetDeadline(_ time.Time) error      { return nil }
func (t *testTCPConn) SetReadDeadline(_ time.Time) error  { return nil }
func (t *testTCPConn) SetWriteDeadline(_ time.Time) error { return nil }

type testAddr struct {
	addr string
}

func (t *testAddr) Network() string { return "tcp" }
func (t *testAddr) String() string  { return t.addr }
