//go:build unix

package runtimeconfig

import (
	"net"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestReadRuntimeFileRejectsFIFOAndSocket(t *testing.T) {
	directory := t.TempDir()
	fifo := filepath.Join(directory, "fifo")
	require.NoError(t, unix.Mkfifo(fifo, 0o600))
	_, err := ReadFile(fifo)
	require.ErrorContains(t, err, "regular file")

	socket := filepath.Join(directory, "socket")
	t.Chdir(directory)
	listener, err := net.Listen("unix", "socket")
	require.NoError(t, err)
	defer listener.Close()
	_, err = ReadFile(socket)
	require.ErrorContains(t, err, "regular file")
}
