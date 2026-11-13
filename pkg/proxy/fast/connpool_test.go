package fast

import (
	"net"
	"net/http"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnPool_ConnReuse(t *testing.T) {
	testCases := []struct {
		desc     string
		poolFn   func(pool *connPool)
		expected int
	}{
		{
			desc: "One connection",
			poolFn: func(pool *connPool) {
				c1, _ := pool.AcquireConn()
				pool.ReleaseConn(c1)
			},
			expected: 1,
		},
		{
			desc: "Two connections with release",
			poolFn: func(pool *connPool) {
				c1, _ := pool.AcquireConn()
				pool.ReleaseConn(c1)

				c2, _ := pool.AcquireConn()
				pool.ReleaseConn(c2)
			},
			expected: 1,
		},
		{
			desc: "Two concurrent connections",
			poolFn: func(pool *connPool) {
				c1, _ := pool.AcquireConn()
				c2, _ := pool.AcquireConn()

				pool.ReleaseConn(c1)
				pool.ReleaseConn(c2)
			},
			expected: 2,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var connAlloc int
			dialer := func() (net.Conn, error) {
				connAlloc++
				// The connection Read blocks so the readLoop stays parked in Peek, as it would on a real idle
				// connection, instead of erroring immediately and marking the connection broken before it is reused.
				return &mockConn{doneCh: make(chan struct{})}, nil
			}

			pool := newConnPool(2, 0, 0, dialer)
			test.poolFn(pool)

			assert.Equal(t, test.expected, connAlloc)
		})
	}
}

func TestConnPool_MaxIdleConn(t *testing.T) {
	testCases := []struct {
		desc        string
		poolFn      func(pool *connPool)
		maxIdleConn int
		expected    int
	}{
		{
			desc: "One connection",
			poolFn: func(pool *connPool) {
				c1, _ := pool.AcquireConn()
				pool.ReleaseConn(c1)
			},
			maxIdleConn: 1,
			expected:    1,
		},
		{
			desc: "Multiple connections with defered release",
			poolFn: func(pool *connPool) {
				for range 7 {
					c, _ := pool.AcquireConn()
					defer pool.ReleaseConn(c)
				}
			},
			maxIdleConn: 5,
			expected:    5,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var keepOpenedConn int
			dialer := func() (net.Conn, error) {
				keepOpenedConn++
				return &mockConn{
					doneCh: make(chan struct{}),
					closeFn: func() error {
						keepOpenedConn--
						return nil
					},
				}, nil
			}

			pool := newConnPool(test.maxIdleConn, 0, 0, dialer)
			test.poolFn(pool)

			assert.Equal(t, test.expected, keepOpenedConn)
		})
	}
}

func TestConnPool_MaxIdleConnFallbacks(t *testing.T) {
	testCases := []struct {
		desc           string
		maxIdleConn    int
		expectedDialed int
		expectedOpened int
	}{
		{
			desc:           "Zero uses the default",
			maxIdleConn:    0,
			expectedDialed: 3,
			expectedOpened: http.DefaultMaxIdleConnsPerHost,
		},
		{
			desc:           "Negative disables reuse",
			maxIdleConn:    -1,
			expectedDialed: 4,
			expectedOpened: 0,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var dialed, closed int
			dialer := func() (net.Conn, error) {
				dialed++
				return &mockConn{
					doneCh: make(chan struct{}),
					closeFn: func() error {
						closed++
						return nil
					},
				}, nil
			}

			pool := newConnPool(test.maxIdleConn, 0, 0, dialer)

			var conns []*conn
			for range 3 {
				c, err := pool.AcquireConn()
				require.NoError(t, err)
				conns = append(conns, c)
			}
			for _, c := range conns {
				pool.ReleaseConn(c)
			}

			c, err := pool.AcquireConn()
			require.NoError(t, err)
			pool.ReleaseConn(c)

			assert.Equal(t, test.expectedDialed, dialed)
			assert.Equal(t, test.expectedOpened, dialed-closed)
		})
	}
}

func TestConnPool_NoReuseConcurrentAcquire(t *testing.T) {
	dialer := func() (net.Conn, error) {
		return &mockConn{doneCh: make(chan struct{})}, nil
	}

	pool := newConnPool(-1, 0, 0, dialer)

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			for range 1000 {
				c, err := pool.AcquireConn()
				if !assert.NoError(t, err) {
					return
				}
				pool.ReleaseConn(c)
			}
		})
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("AcquireConn is stuck")
	}
}

func TestConnPool_CloseIdleConns(t *testing.T) {
	var dialed, closed int
	dialer := func() (net.Conn, error) {
		dialed++
		return &mockConn{
			doneCh: make(chan struct{}),
			closeFn: func() error {
				closed++
				return nil
			},
		}, nil
	}

	pool := newConnPool(2, 0, 0, dialer)

	c1, err := pool.AcquireConn()
	require.NoError(t, err)
	pool.ReleaseConn(c1)

	pool.closeIdleConns()
	assert.Equal(t, 1, closed)

	// The closed connection stays in the pool, but must never be handed out again.
	c2, err := pool.AcquireConn()
	require.NoError(t, err)
	assert.NotSame(t, c1, c2)
	assert.Equal(t, 2, dialed)
}

func TestGC(t *testing.T) {
	// TODO: make the test stable if possible.
	t.Skip("This test is flaky")

	var isDestroyed bool
	pools := map[string]*connPool{}
	dialer := func() (net.Conn, error) {
		c := &mockConn{closeFn: func() error {
			return nil
		}}
		return c, nil
	}

	pools["test"] = newConnPool(10, 1*time.Second, 0, dialer)
	runtime.SetFinalizer(pools["test"], func(p *connPool) {
		isDestroyed = true
	})
	c, err := pools["test"].AcquireConn()
	require.NoError(t, err)

	pools["test"].ReleaseConn(c)

	pools["test"].Close()

	delete(pools, "test")

	runtime.GC()

	require.True(t, isDestroyed)
}

type mockConn struct {
	closeFn func() error
	doneCh  chan struct{} // makes sure that the readLoop is blocking avoiding close.
}

func (m *mockConn) Read(_ []byte) (n int, err error) {
	<-m.doneCh
	return 0, nil
}

func (m *mockConn) Write(_ []byte) (n int, err error) {
	panic("implement me")
}

func (m *mockConn) Close() error {
	defer close(m.doneCh)
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	panic("implement me")
}

func (m *mockConn) RemoteAddr() net.Addr {
	panic("implement me")
}

func (m *mockConn) SetDeadline(_ time.Time) error {
	panic("implement me")
}

func (m *mockConn) SetReadDeadline(_ time.Time) error {
	panic("implement me")
}

func (m *mockConn) SetWriteDeadline(_ time.Time) error {
	panic("implement me")
}
