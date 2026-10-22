//go:build !windows

package tcp

import (
	"io"
	"net"
	"os"
	"syscall"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsReadConnResetError(t *testing.T) {
	testCases := []struct {
		desc     string
		err      error
		expected bool
	}{
		{
			desc:     "read reset",
			err:      &net.OpError{Op: "read", Net: "tcp", Err: os.NewSyscallError("read", syscall.ECONNRESET)},
			expected: true,
		},
		{
			desc:     "write reset",
			err:      &net.OpError{Op: "write", Net: "tcp", Err: os.NewSyscallError("write", syscall.ECONNRESET)},
			expected: false,
		},
		{
			desc:     "read timeout",
			err:      &net.OpError{Op: "read", Net: "tcp", Err: os.ErrDeadlineExceeded},
			expected: false,
		},
		{
			desc:     "no error",
			err:      nil,
			expected: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			assert.Equal(t, test.expected, isReadConnResetError(test.err))
		})
	}
}

func TestIsReadConnResetErrorReadFrom(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	peer, err := listener.Accept()
	require.NoError(t, err)
	t.Cleanup(func() { _ = peer.Close() })

	// Copying into a *net.TCPConn goes through ReadFrom, which reports a read failure as a
	// "readfrom" operation wrapping the original "read" one, as seen with TLS passthrough.
	src := iotest.ErrReader(&net.OpError{Op: "read", Net: "tcp", Err: os.NewSyscallError("read", syscall.ECONNRESET)})

	_, err = io.Copy(conn, src)
	require.Error(t, err)

	assert.True(t, isReadConnResetError(err))
}
