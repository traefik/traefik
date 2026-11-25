package tcp

import (
	"bytes"
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloseWrite(t *testing.T) {
	backendListener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go fakeServer(t, backendListener)
	_, port, err := net.SplitHostPort(backendListener.Addr().String())
	require.NoError(t, err)

	dialer := tcpDialer{&net.Dialer{}, 10 * time.Millisecond, nil}

	proxy, err := NewProxy(":"+port, dialer)
	require.NoError(t, err)

	proxyListener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go func() {
		for {
			conn, err := proxyListener.Accept()
			require.NoError(t, err)
			proxy.ServeTCP(conn.(*net.TCPConn))
		}
	}()

	_, port, err = net.SplitHostPort(proxyListener.Addr().String())
	require.NoError(t, err)

	conn, err := net.Dial("tcp", ":"+port)
	require.NoError(t, err)

	_, err = conn.Write([]byte("ping\n"))
	require.NoError(t, err)

	err = conn.(*net.TCPConn).CloseWrite()
	require.NoError(t, err)

	var buf []byte
	buffer := bytes.NewBuffer(buf)
	n, err := io.Copy(buffer, conn)
	require.NoError(t, err)

	require.Equal(t, int64(4), n)
	require.Equal(t, "PONG", buffer.String())
}

func TestIsReadConnResetError(t *testing.T) {
	testCases := []struct {
		desc        string
		resetClient bool
	}{
		{
			desc:        "client resets the connection",
			resetClient: true,
		},
		{
			desc:        "backend resets the connection",
			resetClient: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client, clientConn := tcpConnPair(t)
			backendConn, backend := tcpConnPair(t)

			// Hides (*net.TCPConn).WriteTo, as the client connection is wrapped when handled by the TCP proxy.
			wrappedClientConn := struct{ WriteCloser }{clientConn}

			errCh := make(chan error, 2)
			go func() {
				_, err := io.Copy(backendConn, wrappedClientConn)
				errCh <- err
			}()
			go func() {
				_, err := io.Copy(wrappedClientConn, backendConn)
				errCh <- err
			}()

			resetConn := backend
			if test.resetClient {
				resetConn = client
			}

			// Closing the connection with a zero linger timeout sends an RST packet.
			require.NoError(t, resetConn.SetLinger(0))
			require.NoError(t, resetConn.Close())

			err := <-errCh
			require.Error(t, err)
			assert.True(t, isReadConnResetError(err), err)
		})
	}

	assert.False(t, isReadConnResetError(nil))
	assert.False(t, isReadConnResetError(io.EOF))
}

// tcpConnPair returns both ends of a TCP connection.
func tcpConnPair(t *testing.T) (*net.TCPConn, *net.TCPConn) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	dialed, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = dialed.Close() })

	accepted, err := listener.Accept()
	require.NoError(t, err)
	t.Cleanup(func() { _ = accepted.Close() })

	return dialed.(*net.TCPConn), accepted.(*net.TCPConn)
}

func fakeServer(t *testing.T, listener net.Listener) {
	t.Helper()

	for {
		conn, err := listener.Accept()
		require.NoError(t, err)

		for {
			withErr := false
			buf := make([]byte, 64)
			if _, err := conn.Read(buf); err != nil {
				withErr = true
			}

			if string(buf[:4]) == "ping" {
				time.Sleep(1 * time.Millisecond)
				if _, err := conn.Write([]byte("PONG")); err != nil {
					_ = conn.Close()
					return
				}
			}

			if withErr {
				_ = conn.Close()
				return
			}
		}
	}
}

// halfCloselessDialer opens connections that cannot close their write half, as net.Pipe ones cannot.
type halfCloselessDialer struct {
	backend net.Conn
}

func (d halfCloselessDialer) Dial(string, string, ClientConn) (net.Conn, error) {
	return d.backend, nil
}

func (d halfCloselessDialer) DialContext(context.Context, string, string, ClientConn) (net.Conn, error) {
	return d.backend, nil
}

func (d halfCloselessDialer) TerminationDelay() time.Duration { return 0 }

// A dialer is not always a host one, and the connections of another network are not always *net.TCPConn:
// one that cannot close its write half fails the connection it was opened for, and not the process.
func TestProxyRefusesABackendConnectionThatCannotCloseItsWriteHalf(t *testing.T) {
	backend, backendPeer := net.Pipe()
	t.Cleanup(func() { _ = backendPeer.Close() })

	proxy, err := NewProxy("backend.invalid:8080", halfCloselessDialer{backend: backend})
	require.NoError(t, err)

	clientListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientListener.Close() })

	served := make(chan struct{})
	go func() {
		defer close(served)

		conn, err := clientListener.Accept()
		if err != nil {
			return
		}

		proxy.ServeTCP(conn.(*net.TCPConn))
	}()

	client, err := net.Dial("tcp", clientListener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	select {
	case <-served:
	case <-time.After(5 * time.Second):
		t.Fatal("the connection was not refused")
	}

	// The client connection is closed, and so is the backend one that could not be used.
	require.NoError(t, client.SetReadDeadline(time.Now().Add(5*time.Second)))
	_, err = client.Read(make([]byte, 1))
	require.ErrorIs(t, err, io.EOF)

	_, err = backendPeer.Read(make([]byte, 1))
	require.ErrorIs(t, err, io.EOF)
}
