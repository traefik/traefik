package tcp

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeConnWithAddr struct {
	fakeConn

	remoteAddr net.Addr
}

func (f *fakeConnWithAddr) RemoteAddr() net.Addr {
	return f.remoteAddr
}

func newFakeConnWithAddr(addr string) *fakeConnWithAddr {
	return &fakeConnWithAddr{
		fakeConn:   fakeConn{writeCall: make(map[string]int)},
		remoteAddr: &net.TCPAddr{IP: net.ParseIP(addr), Port: 12345},
	}
}

func handledBy(conn *fakeConnWithAddr) string {
	for name, count := range conn.writeCall {
		if count > 0 {
			return name
		}
	}
	return ""
}

func TestIPHashLoadBalancer_SameIPAlwaysSameServer(t *testing.T) {
	balancer := NewIPHashLoadBalancer(false)

	for _, name := range []string{"h1", "h2", "h3"} {
		n := name
		balancer.Add(n, HandlerFunc(func(conn WriteCloser) {
			_, err := conn.Write([]byte(n))
			require.NoError(t, err)
		}), pointer(1))
	}

	var expected string
	for i := range 10 {
		conn := newFakeConnWithAddr("10.0.0.1")
		balancer.ServeTCP(conn)
		got := handledBy(conn)
		if i == 0 {
			expected = got
		}
		assert.Equal(t, expected, got, "same IP should always route to the same server")
	}
}

func TestIPHashLoadBalancer_DifferentIPsDifferentServers(t *testing.T) {
	balancer := NewIPHashLoadBalancer(false)

	for _, name := range []string{"h1", "h2", "h3"} {
		n := name
		balancer.Add(n, HandlerFunc(func(conn WriteCloser) {
			_, err := conn.Write([]byte(n))
			require.NoError(t, err)
		}), pointer(1))
	}

	hits := make(map[string]int)
	for i := range 30 {
		conn := newFakeConnWithAddr(fmt.Sprintf("10.0.0.%d", i+1))
		balancer.ServeTCP(conn)
		hits[handledBy(conn)]++
	}

	assert.Positive(t, hits["h1"], "h1 should receive some traffic")
	assert.Positive(t, hits["h2"], "h2 should receive some traffic")
	assert.Positive(t, hits["h3"], "h3 should receive some traffic")
}

func TestIPHashLoadBalancer_NoServers(t *testing.T) {
	balancer := NewIPHashLoadBalancer(false)

	conn := newFakeConnWithAddr("10.0.0.1")
	balancer.ServeTCP(conn)

	assert.Empty(t, conn.writeCall)
	assert.Equal(t, 1, conn.closeCall)
}

func TestIPHashLoadBalancer_AllServersDown(t *testing.T) {
	balancer := NewIPHashLoadBalancer(false)

	balancer.Add("h1", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("h1"))
		require.NoError(t, err)
	}), pointer(1))

	balancer.Add("h2", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("h2"))
		require.NoError(t, err)
	}), pointer(1))

	balancer.SetStatus(t.Context(), "h1", false)
	balancer.SetStatus(t.Context(), "h2", false)

	conn := newFakeConnWithAddr("10.0.0.1")
	balancer.ServeTCP(conn)

	assert.Empty(t, conn.writeCall)
	assert.Equal(t, 1, conn.closeCall)
}

func TestIPHashLoadBalancer_ServerDownThenUp(t *testing.T) {
	balancer := NewIPHashLoadBalancer(false)

	balancer.Add("h1", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("h1"))
		require.NoError(t, err)
	}), pointer(1))

	balancer.Add("h2", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("h2"))
		require.NoError(t, err)
	}), pointer(1))

	// Find which server "10.0.0.1" maps to.
	conn := newFakeConnWithAddr("10.0.0.1")
	balancer.ServeTCP(conn)
	naturalServer := handledBy(conn)

	// Mark that server down — traffic must go to the other one.
	balancer.SetStatus(t.Context(), naturalServer, false)

	conn = newFakeConnWithAddr("10.0.0.1")
	balancer.ServeTCP(conn)
	assert.NotEqual(t, naturalServer, handledBy(conn), "traffic should move away from the downed server")

	// Bring it back up — traffic should return to the natural server.
	balancer.SetStatus(t.Context(), naturalServer, true)

	conn = newFakeConnWithAddr("10.0.0.1")
	balancer.ServeTCP(conn)
	assert.Equal(t, naturalServer, handledBy(conn), "traffic should return to the recovered server")
}

// TestIPHashLoadBalancer_StabilityOnServerDown verifies the HRW property:
// when one server goes down, only IPs that were mapped to that server are
// redistributed — other IPs continue routing to their original server.
func TestIPHashLoadBalancer_StabilityOnServerDown(t *testing.T) {
	balancer := NewIPHashLoadBalancer(false)

	for _, name := range []string{"h1", "h2", "h3"} {
		n := name
		balancer.Add(n, HandlerFunc(func(conn WriteCloser) {
			_, err := conn.Write([]byte(n))
			require.NoError(t, err)
		}), pointer(1))
	}

	// Record which server each IP maps to with all servers healthy.
	ips := []string{
		"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4", "10.0.0.5",
		"10.0.0.6", "10.0.0.7", "10.0.0.8", "10.0.0.9", "10.0.0.10",
	}

	originalMapping := make(map[string]string)
	for _, ip := range ips {
		conn := newFakeConnWithAddr(ip)
		balancer.ServeTCP(conn)
		originalMapping[ip] = handledBy(conn)
	}

	// Take h3 down.
	balancer.SetStatus(t.Context(), "h3", false)

	// IPs that were NOT on h3 must still go to their original server.
	for _, ip := range ips {
		if originalMapping[ip] == "h3" {
			continue // these will be redistributed — that's expected
		}

		conn := newFakeConnWithAddr(ip)
		balancer.ServeTCP(conn)
		assert.Equal(t, originalMapping[ip], handledBy(conn),
			"IP %s should still route to %s after h3 went down", ip, originalMapping[ip])
	}
}

// TestIPHashLoadBalancer_Propagate verifies that status changes propagate
// to a parent balancer via RegisterStatusUpdater.
func TestIPHashLoadBalancer_Propagate(t *testing.T) {
	child := NewIPHashLoadBalancer(true)

	child.Add("h1", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("h1"))
		require.NoError(t, err)
	}), pointer(1))

	child.Add("h2", HandlerFunc(func(conn WriteCloser) {
		_, err := conn.Write([]byte("h2"))
		require.NoError(t, err)
	}), pointer(1))

	var upEvents []bool
	err := child.RegisterStatusUpdater(func(up bool) {
		upEvents = append(upEvents, up)
	})
	require.NoError(t, err)

	// Bringing one server down should NOT propagate — child is still up.
	child.SetStatus(t.Context(), "h1", false)
	assert.Empty(t, upEvents, "one server down should not propagate while another is healthy")

	// Bringing the last server down SHOULD propagate with up=false.
	child.SetStatus(t.Context(), "h2", false)
	assert.Equal(t, []bool{false}, upEvents, "all servers down should propagate false")

	// Bringing one server back up SHOULD propagate with up=true.
	child.SetStatus(t.Context(), "h1", true)
	assert.Equal(t, []bool{false, true}, upEvents, "server recovery should propagate true")
}

func TestIPHashLoadBalancer_RegisterStatusUpdater_NoHealthCheck(t *testing.T) {
	balancer := NewIPHashLoadBalancer(false)

	err := balancer.RegisterStatusUpdater(func(up bool) {})
	assert.Error(t, err)
}
