package udp

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWRRLoadBalancerWeightedDrop(t *testing.T) {
	listener, err := Listen(net.ListenConfig{}, "udp", "127.0.0.1:0", time.Second)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, listener.Close()) })

	echo := HandlerFunc(func(conn *Conn) {
		defer conn.Close()

		buffer := make([]byte, 4)
		n, readErr := conn.Read(buffer)
		if readErr != nil {
			return
		}

		_, _ = conn.Write(buffer[:n])
	})

	empty := NewWRRLoadBalancer()
	wrr := NewWRRLoadBalancer()
	wrr.AddWeightedServer(echo, new(7))
	wrr.AddWeightedServer(empty, new(3))

	go func() {
		for {
			conn, acceptErr := listener.Accept()
			if acceptErr != nil {
				return
			}

			go wrr.ServeUDP(conn)
		}
	}()

	var responses int
	var drops int
	for range 10 {
		conn, dialErr := net.Dial("udp", listener.Addr().String())
		require.NoError(t, dialErr)

		require.NoError(t, conn.SetDeadline(time.Now().Add(500*time.Millisecond)))
		_, writeErr := conn.Write([]byte("PING"))
		require.NoError(t, writeErr)

		buffer := make([]byte, 4)
		_, readErr := conn.Read(buffer)
		require.NoError(t, conn.Close())

		if readErr == nil {
			responses++
			continue
		}

		var netErr net.Error
		require.ErrorAs(t, readErr, &netErr)
		assert.True(t, netErr.Timeout())
		drops++
	}

	assert.Equal(t, 7, responses)
	assert.Equal(t, 3, drops)
}
