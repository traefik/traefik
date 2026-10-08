package fast

import (
	"bufio"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// concurrentReadDetector reports whether two goroutines were ever inside Read at the same time. The pool
// connection's bufio.Reader is the only reader of this conn, so an overlap means two goroutines were driving
// that single reader concurrently.
type concurrentReadDetector struct {
	net.Conn

	inRead     atomic.Int32
	overlapped atomic.Bool
}

func (c *concurrentReadDetector) Read(b []byte) (int, error) {
	if c.inRead.Add(1) > 1 {
		c.overlapped.Store(true)
	}
	defer c.inRead.Add(-1)

	return c.Conn.Read(b)
}

// After a 101, the connection is handed over to the tunnel copiers, which read it through the same
// bufio.Reader: the pool readLoop must not read from it anymore, or both goroutines drive one bufio.Reader
// and corrupt its internal offsets (see GHSA-ww7c-q3mv-xx22).
func TestUpgradedConnNotReadByReadLoop(t *testing.T) {
	upgraded := make(chan struct{})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()

		br := bufio.NewReader(c)
		for {
			line, err := br.ReadString('\n')
			if err != nil || line == "\r\n" {
				break
			}
		}

		if _, err := c.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nUpgrade: test\r\nConnection: Upgrade\r\n\r\n")); err != nil {
			return
		}

		// Stay silent so copyFromBackend parks inside c.br.Read instead of being woken up and dying on the
		// write to the aborted client, which is what keeps the second reader alive and the overlap observable.
		<-upgraded
	}()

	var detector atomic.Pointer[concurrentReadDetector]

	pool := newConnPool(200, 0, 0, func() (net.Conn, error) {
		c, err := net.Dial("tcp", ln.Addr().String())
		if err != nil {
			return nil, err
		}

		d := &concurrentReadDetector{Conn: c}
		detector.Store(d)

		return d, nil
	})
	t.Cleanup(pool.Close)
	t.Cleanup(func() { close(upgraded) })

	proxy := createProxyWithForwarder(t, "http://"+ln.Addr().String(), pool)

	cConn, err := net.Dial("tcp", proxy.Listener.Addr().String())
	require.NoError(t, err)

	_, err = cConn.Write([]byte("GET /ws HTTP/1.1\r\nHost: localhost\r\nConnection: Upgrade\r\nUpgrade: test\r\n\r\n"))
	require.NoError(t, err)

	buf := make([]byte, 1024)
	_, err = cConn.Read(buf)
	require.NoError(t, err)

	// Abort so copyToBackend returns first and the upgrade handler returns while copyFromBackend is still
	// parked in the shared reader.
	require.NoError(t, cConn.(*net.TCPConn).SetLinger(0))
	require.NoError(t, cConn.Close())

	// Let the readLoop come back around to Peek on the upgraded connection.
	time.Sleep(500 * time.Millisecond)

	d := detector.Load()
	require.NotNil(t, d)
	assert.False(t, d.overlapped.Load())
}
