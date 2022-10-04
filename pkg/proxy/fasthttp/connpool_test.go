package fasthttp

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnPool_ConnReuse(t *testing.T) {
	testCases := []struct {
		desc     string
		poolFn   func(pool *connPool)
		expected int
	}{
		{
			desc: "Simple case",
			poolFn: func(pool *connPool) {
				c1, _ := pool.AcquireConn()
				pool.ReleaseConn(c1)
			},
			expected: 1,
		},
		{
			desc: "Simple with reuse",
			poolFn: func(pool *connPool) {
				c1, _ := pool.AcquireConn()
				pool.ReleaseConn(c1)

				c2, _ := pool.AcquireConn()
				pool.ReleaseConn(c2)
			},
			expected: 1,
		},
		{
			desc: "Two connection at the same time",
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
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			var connAlloc int
			dialer := func() (net.Conn, error) {
				connAlloc++
				return &net.TCPConn{}, nil
			}
			pool := NewConnPool(2, 0, dialer)
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
			desc: "Simple case",
			poolFn: func(pool *connPool) {
				c1, _ := pool.AcquireConn()
				pool.ReleaseConn(c1)
			},
			maxIdleConn: 1,
			expected:    1,
		},
		{
			desc: "Multiple conn with release",
			poolFn: func(pool *connPool) {
				for i := 0; i < 7; i++ {
					c, _ := pool.AcquireConn()
					defer pool.ReleaseConn(c)
				}
			},
			maxIdleConn: 5,
			expected:    5,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var keepOpenedConn int
			dialer := func() (net.Conn, error) {
				keepOpenedConn++
				return &mockConn{closeFn: func() error {
					keepOpenedConn--
					return nil
				}}, nil
			}
			pool := NewConnPool(test.maxIdleConn, 0, dialer)
			test.poolFn(pool)

			assert.Equal(t, test.expected, keepOpenedConn)
		})
	}
}

type mockConn struct {
	closeFn func() error
}

func (m *mockConn) Read(_ []byte) (n int, err error) {
	panic("implement me")
}

func (m *mockConn) Write(_ []byte) (n int, err error) {
	panic("implement me")
}

func (m *mockConn) Close() error {
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
