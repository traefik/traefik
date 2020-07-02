package udp

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsecutiveWrites(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", ":0")
	require.NoError(t, err)

	ln, err := Listen("udp", addr)
	require.NoError(t, err)
	defer func() {
		err := ln.Close()
		require.NoError(t, err)
	}()

	go func() {
		for {
			conn, err := ln.Accept()
			if err == errClosedListener {
				return
			}
			require.NoError(t, err)

			go func() {
				b := make([]byte, 2048)
				b2 := make([]byte, 2048)
				var n int
				var n2 int

				n, err = conn.Read(b)
				require.NoError(t, err)
				// Wait to make sure that the second packet is received
				time.Sleep(10 * time.Millisecond)
				n2, err = conn.Read(b2)
				require.NoError(t, err)

				_, err = conn.Write(b[:n])
				require.NoError(t, err)
				_, err = conn.Write(b2[:n2])
				require.NoError(t, err)
			}()
		}
	}()

	udpConn, err := net.Dial("udp", ln.Addr().String())
	require.NoError(t, err)

	// Send multiple packets of different content and length consecutively
	// Read back packets afterwards and make sure that content matches
	// This checks if any buffers are overwritten while the receiver is enqueuing multiple packets
	b := make([]byte, 2048)
	var n int
	_, err = udpConn.Write([]byte("TESTLONG0"))
	require.NoError(t, err)
	_, err = udpConn.Write([]byte("1TEST"))
	require.NoError(t, err)

	n, err = udpConn.Read(b)
	require.NoError(t, err)
	require.Equal(t, "TESTLONG0", string(b[:n]))
	n, err = udpConn.Read(b)
	require.NoError(t, err)
	require.Equal(t, "1TEST", string(b[:n]))
}

func TestListenNotBlocking(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", ":0")

	require.NoError(t, err)

	ln, err := Listen("udp", addr)
	require.NoError(t, err)
	defer func() {
		err := ln.Close()
		require.NoError(t, err)
	}()

	go func() {
		for {
			conn, err := ln.Accept()
			if err == errClosedListener {
				return
			}
			require.NoError(t, err)

			go func() {
				b := make([]byte, 2048)
				n, err := conn.Read(b)
				require.NoError(t, err)
				_, err = conn.Write(b[:n])
				require.NoError(t, err)

				n, err = conn.Read(b)
				require.NoError(t, err)
				_, err = conn.Write(b[:n])
				require.NoError(t, err)

				// This should not block second call
				time.Sleep(10 * time.Second)
			}()
		}
	}()

	udpConn, err := net.Dial("udp", ln.Addr().String())
	require.NoError(t, err)

	_, err = udpConn.Write([]byte("TEST"))
	require.NoError(t, err)

	b := make([]byte, 2048)
	n, err := udpConn.Read(b)
	require.NoError(t, err)
	require.Equal(t, "TEST", string(b[:n]))

	_, err = udpConn.Write([]byte("TEST2"))
	require.NoError(t, err)

	n, err = udpConn.Read(b)
	require.NoError(t, err)
	require.Equal(t, "TEST2", string(b[:n]))

	_, err = udpConn.Write([]byte("TEST"))
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		udpConn2, err := net.Dial("udp", ln.Addr().String())
		require.NoError(t, err)

		_, err = udpConn2.Write([]byte("TEST"))
		require.NoError(t, err)

		n, err = udpConn2.Read(b)
		require.NoError(t, err)

		assert.Equal(t, "TEST", string(b[:n]))

		_, err = udpConn2.Write([]byte("TEST2"))
		require.NoError(t, err)

		n, err = udpConn2.Read(b)
		require.NoError(t, err)

		assert.Equal(t, "TEST2", string(b[:n]))

		close(done)
	}()

	select {
	case <-time.Tick(time.Second):
		t.Error("Timeout")
	case <-done:
	}
}

func TestTimeoutWithRead(t *testing.T) {
	testTimeout(t, true)
}

func TestTimeoutWithoutRead(t *testing.T) {
	testTimeout(t, false)
}

func testTimeout(t *testing.T, withRead bool) {
	addr, err := net.ResolveUDPAddr("udp", ":0")
	require.NoError(t, err)

	ln, err := Listen("udp", addr)
	require.NoError(t, err)
	defer func() {
		err := ln.Close()
		require.NoError(t, err)
	}()

	go func() {
		for {
			conn, err := ln.Accept()
			if err == errClosedListener {
				return
			}
			require.NoError(t, err)

			if withRead {
				buf := make([]byte, 1024)
				_, err = conn.Read(buf)

				require.NoError(t, err)
			}
		}
	}()

	for i := 0; i < 10; i++ {
		udpConn2, err := net.Dial("udp", ln.Addr().String())
		require.NoError(t, err)

		_, err = udpConn2.Write([]byte("TEST"))
		require.NoError(t, err)
	}

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 10, len(ln.conns))

	time.Sleep(4 * time.Second)
	assert.Equal(t, 0, len(ln.conns))
}

func TestShutdown(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", ":0")
	require.NoError(t, err)

	l, err := Listen("udp", addr)
	require.NoError(t, err)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}

			go func() {
				conn := conn
				for {
					b := make([]byte, 1024*1024)
					n, err := conn.Read(b)
					require.NoError(t, err)
					// We control the termination,
					// otherwise we would block on the Read above,
					// until conn is closed by a timeout.
					// Which means we would get an error,
					// and even though we are in a goroutine and the current test might be over,
					// go test would still yell at us if this happens while other tests are still running.
					if string(b[:n]) == "CLOSE" {
						return
					}
					_, err = conn.Write(b[:n])
					require.NoError(t, err)
				}
			}()
		}
	}()

	conn, err := net.Dial("udp", l.Addr().String())
	require.NoError(t, err)

	// Start sending packets, to create a "session" with the server.
	requireEcho(t, "TEST", conn, time.Second)

	doneChan := make(chan struct{})
	go func() {
		err := l.Shutdown(5 * time.Second)
		require.NoError(t, err)
		close(doneChan)
	}()

	// Make sure that our session is still live even after the shutdown.
	requireEcho(t, "TEST2", conn, time.Second)

	// And make sure that on the other hand, opening new sessions is not possible anymore.
	conn2, err := net.Dial("udp", l.Addr().String())
	require.NoError(t, err)

	_, err = conn2.Write([]byte("TEST"))
	// Packet is accepted, but dropped
	require.NoError(t, err)

	// Make sure that our session is yet again still live.
	// This is specifically to make sure we don't create a regression in listener's readLoop,
	// i.e. that we only terminate the listener's readLoop goroutine by closing its pConn.
	requireEcho(t, "TEST3", conn, time.Second)

	done := make(chan bool)
	go func() {
		defer close(done)
		b := make([]byte, 1024*1024)
		n, err := conn2.Read(b)
		require.Error(t, err)
		assert.Equal(t, 0, n)
	}()

	conn2.Close()

	select {
	case <-done:
	case <-time.Tick(time.Second):
		t.Fatal("Timeout")
	}

	_, err = conn.Write([]byte("CLOSE"))
	require.NoError(t, err)

	select {
	case <-doneChan:
	case <-time.Tick(5 * time.Second):
		// In case we introduce a regression that would make the test wait forever.
		t.Fatal("Timeout during shutdown")
	}
}

// requireEcho tests that the conn session is live and functional,
// by writing data through it, and expecting the same data as a response when reading on it.
// It fatals if the read blocks longer than timeout,
// which is useful to detect regressions that would make a test wait forever.
func requireEcho(t *testing.T, data string, conn io.ReadWriter, timeout time.Duration) {
	_, err := conn.Write([]byte(data))
	require.NoError(t, err)

	doneChan := make(chan struct{})
	go func() {
		b := make([]byte, 1024*1024)
		n, err := conn.Read(b)
		require.NoError(t, err)
		assert.Equal(t, data, string(b[:n]))
		close(doneChan)
	}()

	select {
	case <-doneChan:
	case <-time.Tick(timeout):
		t.Fatalf("Timeout during echo for: %s", data)
	}
}
