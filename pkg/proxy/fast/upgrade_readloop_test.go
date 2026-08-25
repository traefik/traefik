package fast

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

// signalConn reports when the connection is closed.
type signalConn struct {
	net.Conn
	once   sync.Once
	closed chan struct{}
}

func (s *signalConn) Close() error {
	s.once.Do(func() { close(s.closed) })
	return s.Conn.Close()
}

// TestReadLoopReturnsAfterUpgrade pins the invariant behind the WebSocket race:
// once a connection has been upgraded, readLoop must stop watching it instead of
// looping back to Peek the bufio.Reader that a protocol copier now owns. It is a
// deterministic check that does not need the race detector - the upgrade handler
// here is a no-op, so the only thing that could read the connection again is a
// second Peek. With readLoop returning on upgrade its deferred Close runs; if it
// loops back to Peek instead, it blocks and the connection is never closed.
func TestReadLoopReturnsAfterUpgrade(t *testing.T) {
	proxyEnd, backendEnd := net.Pipe()
	t.Cleanup(func() { _ = backendEnd.Close() })

	sc := &signalConn{Conn: proxyEnd, closed: make(chan struct{})}
	c := &conn{
		Conn:  sc,
		RWCh:  make(chan rwWithUpgrade),
		ErrCh: make(chan error, 1),
		br:    bufio.NewReader(sc),
	}
	c.expectedResponse.Store(true)

	go c.readLoop()

	go func() {
		_, _ = backendEnd.Write([]byte("HTTP/1.1 101 Switching Protocols\r\n" +
			"Connection: Upgrade\r\nUpgrade: websocket\r\n\r\n"))
	}()

	// Hand off the response with an upgrade handler that does nothing, so no
	// copier competes for the reader.
	c.RWCh <- rwWithUpgrade{
		ReqMethod: http.MethodGet,
		RW:        httptest.NewRecorder(),
		Upgrade:   func(_ http.ResponseWriter, _ *fasthttp.Response, _ net.Conn) {},
	}

	require.NoError(t, <-c.ErrCh)
	require.True(t, c.isUpgraded())

	select {
	case <-sc.closed:
	case <-time.After(2 * time.Second):
		t.Fatal("readLoop kept the upgraded connection and peeked its reader again")
	}
}
