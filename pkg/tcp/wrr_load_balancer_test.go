package tcp

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeConn struct {
	call map[string]int
}

func (f *fakeConn) Read(b []byte) (n int, err error) {
	panic("implement me")
}

func (f *fakeConn) Write(b []byte) (n int, err error) {
	f.call[string(b)]++
	return len(b), nil
}

func (f *fakeConn) Close() error {
	panic("implement me")
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

func TestLoadBalancing(t *testing.T) {
	testCases := []struct {
		desc          string
		serversWeight map[string]int
		totalCall     int
		expected      map[string]int
	}{
		{
			desc: "RoundRobin",
			serversWeight: map[string]int{
				"h1": 1,
				"h2": 1,
			},
			totalCall: 4,
			expected: map[string]int{
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
			expected: map[string]int{
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
			expected: map[string]int{
				"h1": 12,
				"h2": 4,
			},
		},
		{
			desc: "WeighedRoundRobin with 0 weight server",
			serversWeight: map[string]int{
				"h1": 3,
				"h2": 0,
			},
			totalCall: 16,
			expected: map[string]int{
				"h1": 16,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			balancer := NewWRRLoadBalancer()
			for server, weight := range test.serversWeight {
				server := server
				balancer.AddWeightServer(HandlerFunc(func(conn WriteCloser) {
					_, err := conn.Write([]byte(server))
					require.NoError(t, err)
				}), &weight)
			}

			conn := &fakeConn{call: make(map[string]int)}
			for i := 0; i < test.totalCall; i++ {
				balancer.ServeTCP(conn)
			}

			assert.Equal(t, test.expected, conn.call)
		})
	}
}
