package fast

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnPoolCloseClosesIdleConns(t *testing.T) {
	var accepted, closedByPool atomic.Int64

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}

			accepted.Add(1)

			go func(c net.Conn) {
				defer func() { _ = c.Close() }()

				br := bufio.NewReader(c)
				for {
					if _, err := http.ReadRequest(br); err != nil {
						closedByPool.Add(1)
						return
					}

					_, _ = io.WriteString(c, "HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nhi")
				}
			}(c)
		}
	}()

	const conns = 4

	pool := newConnPool(conns, 30*time.Second, 0, func() (net.Conn, error) {
		return net.Dial("tcp", ln.Addr().String())
	})

	acquired := make([]*conn, 0, conns)
	for range conns {
		co, err := pool.AcquireConn()
		require.NoError(t, err)
		acquired = append(acquired, co)
	}
	for _, co := range acquired {
		pool.ReleaseConn(co)
	}

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.Equal(c, int64(conns), accepted.Load())
	}, time.Second, 10*time.Millisecond)

	pool.Close()

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.Equal(c, int64(conns), closedByPool.Load())
	}, time.Second, 10*time.Millisecond)
}

func TestConnPoolCloseWithoutIdleConnTimeout(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			t.Cleanup(func() { _ = c.Close() })
		}
	}()

	// idleConnTimeout == 0 disables the cleaner goroutine; Close must still
	// close the pooled connections.
	pool := newConnPool(1, 0, 0, func() (net.Conn, error) {
		return net.Dial("tcp", ln.Addr().String())
	})

	co, err := pool.AcquireConn()
	require.NoError(t, err)
	pool.ReleaseConn(co)

	pool.Close()

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		co.closeMu.Lock()
		defer co.closeMu.Unlock()
		assert.True(c, co.closed)
	}, time.Second, 10*time.Millisecond)
}
