package fast

import (
	"net"
	"runtime"
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
				return &net.TCPConn{}, nil
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
