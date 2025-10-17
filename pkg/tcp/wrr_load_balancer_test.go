package tcp

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWRRLoadBalancer_LoadBalancing(t *testing.T) {
	testCases := []struct {
		desc          string
		serversWeight map[string]int
		totalCall     int
		expectedWrite map[string]int
		expectedClose int
	}{
		{
			desc: "RoundRobin",
			serversWeight: map[string]int{
				"h1": 1,
				"h2": 1,
			},
			totalCall: 4,
			expectedWrite: map[string]int{
				"h1": 2,
				"h2": 2,
			},
		},
		{
			desc: "WeighedRoundRobin",
			serversWeight: map[string]int{
				"h1": 3,
				"h2": 1,
			},
			totalCall: 4,
			expectedWrite: map[string]int{
				"h1": 3,
				"h2": 1,
			},
		},
		{
			desc: "WeighedRoundRobin with more call",
			serversWeight: map[string]int{
				"h1": 3,
				"h2": 1,
			},
			totalCall: 16,
			expectedWrite: map[string]int{
				"h1": 12,
				"h2": 4,
			},
		},
		{
			desc: "WeighedRoundRobin with one 0 weight server",
			serversWeight: map[string]int{
				"h1": 3,
				"h2": 0,
			},
			totalCall: 16,
			expectedWrite: map[string]int{
				"h1": 16,
			},
		},
		{
			desc: "WeighedRoundRobin with all servers with 0 weight",
			serversWeight: map[string]int{
				"h1": 0,
				"h2": 0,
				"h3": 0,
			},
			totalCall:     10,
			expectedWrite: map[string]int{},
			expectedClose: 10,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			balancer := NewWRRLoadBalancer(false)
			for server, weight := range test.serversWeight {
				balancer.Add(server, HandlerFunc(func(conn WriteCloser) {
					_, err := conn.Write([]byte(server))
					require.NoError(t, err)
				}), &weight)
			}

			conn := &fakeConn{writeCall: make(map[string]int)}
			for range test.totalCall {
				balancer.ServeTCP(conn)
			}

			assert.Equal(t, test.expectedWrite, conn.writeCall)
			assert.Equal(t, test.expectedClose, conn.closeCall)
		})
	}
}

func TestWRRLoadBalancer_NoServiceUp(t *testing.T) {
	balancer := NewWRRLoadBalancer(false)

	balancer.Add("first", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("first"))
		require.NoError(t, err)
	}), pointer(1))

	balancer.Add("second", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("second"))
		require.NoError(t, err)
	}), pointer(1))

	balancer.SetStatus(t.Context(), "first", false)
	balancer.SetStatus(t.Context(), "second", false)

	conn := &fakeConn{writeCall: make(map[string]int)}
	balancer.ServeTCP(conn)

	assert.Empty(t, conn.writeCall)
	assert.Equal(t, 1, conn.closeCall)
}

func TestWRRLoadBalancer_OneServerDown(t *testing.T) {
	balancer := NewWRRLoadBalancer(false)

	balancer.Add("first", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("first"))
		require.NoError(t, err)
	}), pointer(1))

	balancer.Add("second", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("second"))
		require.NoError(t, err)
	}), pointer(1))

	balancer.SetStatus(t.Context(), "second", false)

	conn := &fakeConn{writeCall: make(map[string]int)}
	for range 3 {
		balancer.ServeTCP(conn)
	}
	assert.Equal(t, 3, conn.writeCall["first"])
}

func TestWRRLoadBalancer_DownThenUp(t *testing.T) {
	balancer := NewWRRLoadBalancer(false)

	balancer.Add("first", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("first"))
		require.NoError(t, err)
	}), pointer(1))

	balancer.Add("second", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("second"))
		require.NoError(t, err)
	}), pointer(1))

	balancer.SetStatus(t.Context(), "second", false)

	conn := &fakeConn{writeCall: make(map[string]int)}
	for range 3 {
		balancer.ServeTCP(conn)
	}
	assert.Equal(t, 3, conn.writeCall["first"])

	balancer.SetStatus(t.Context(), "second", true)

	conn = &fakeConn{writeCall: make(map[string]int)}
	for range 2 {
		balancer.ServeTCP(conn)
	}
	assert.Equal(t, 1, conn.writeCall["first"])
	assert.Equal(t, 1, conn.writeCall["second"])
}

func TestWRRLoadBalancer_Propagate(t *testing.T) {
	balancer1 := NewWRRLoadBalancer(true)

	balancer1.Add("first", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("first"))
		require.NoError(t, err)
	}), pointer(1))

	balancer1.Add("second", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("second"))
		require.NoError(t, err)
	}), pointer(1))

	balancer2 := NewWRRLoadBalancer(true)

	balancer2.Add("third", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("third"))
		require.NoError(t, err)
	}), pointer(1))

	balancer2.Add("fourth", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("fourth"))
		require.NoError(t, err)
	}), pointer(1))

	topBalancer := NewWRRLoadBalancer(true)

	topBalancer.Add("balancer1", balancer1, pointer(1))
	_ = balancer1.RegisterStatusUpdater(func(up bool) {
		topBalancer.SetStatus(t.Context(), "balancer1", up)
	})

	topBalancer.Add("balancer2", balancer2, pointer(1))
	_ = balancer2.RegisterStatusUpdater(func(up bool) {
		topBalancer.SetStatus(t.Context(), "balancer2", up)
	})

	conn := &fakeConn{writeCall: make(map[string]int)}
	for range 8 {
		topBalancer.ServeTCP(conn)
	}
	assert.Equal(t, 2, conn.writeCall["first"])
	assert.Equal(t, 2, conn.writeCall["second"])
	assert.Equal(t, 2, conn.writeCall["third"])
	assert.Equal(t, 2, conn.writeCall["fourth"])

	// fourth gets downed, but balancer2 still up since third is still up.
	balancer2.SetStatus(t.Context(), "fourth", false)

	conn = &fakeConn{writeCall: make(map[string]int)}
	for range 8 {
		topBalancer.ServeTCP(conn)
	}
	assert.Equal(t, 2, conn.writeCall["first"])
	assert.Equal(t, 2, conn.writeCall["second"])
	assert.Equal(t, 4, conn.writeCall["third"])
	assert.Equal(t, 0, conn.writeCall["fourth"])

	// third gets downed, and the propagation triggers balancer2 to be marked as
	// down as well for topBalancer.
	balancer2.SetStatus(t.Context(), "third", false)

	conn = &fakeConn{writeCall: make(map[string]int)}
	for range 8 {
		topBalancer.ServeTCP(conn)
	}
	assert.Equal(t, 4, conn.writeCall["first"])
	assert.Equal(t, 4, conn.writeCall["second"])
	assert.Equal(t, 0, conn.writeCall["third"])
	assert.Equal(t, 0, conn.writeCall["fourth"])
}

func pointer[T any](v T) *T { return &v }

type fakeConn struct {
	writeCall map[string]int
	closeCall int
}

func (f *fakeConn) Read(b []byte) (n int, err error) {
	panic("implement me")
}

func (f *fakeConn) Write(b []byte) (n int, err error) {
	f.writeCall[string(b)]++
	return len(b), nil
}

func (f *fakeConn) Close() error {
	f.closeCall++
	return nil
}

func (f *fakeConn) LocalAddr() net.Addr {
	panic("implement me")
}

func (f *fakeConn) RemoteAddr() net.Addr {
	panic("implement me")
}

func (f *fakeConn) SetDeadline(t time.Time) error {
	panic("implement me")
}

func (f *fakeConn) SetReadDeadline(t time.Time) error {
	panic("implement me")
}

func (f *fakeConn) SetWriteDeadline(t time.Time) error {
	panic("implement me")
}

func (f *fakeConn) CloseWrite() error {
	panic("implement me")
}
