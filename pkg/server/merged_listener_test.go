package server

import (
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubListener is a listener whose Accept is scripted:
// it returns each queued result in turn, then blocks until closed.
type stubListener struct {
	results chan acceptResult
	closed  chan struct{}
	once    sync.Once
}

type acceptResult struct {
	conn net.Conn
	err  error
}

func newStubListener(results ...acceptResult) *stubListener {
	l := &stubListener{results: make(chan acceptResult, len(results)), closed: make(chan struct{})}
	for _, r := range results {
		l.results <- r
	}
	return l
}

func (l *stubListener) Accept() (net.Conn, error) {
	select {
	case r := <-l.results:
		return r.conn, r.err
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

func (l *stubListener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

func (l *stubListener) Addr() net.Addr { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)} }

func pipeConn(t *testing.T) net.Conn {
	t.Helper()

	client, server := net.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})

	return server
}

// One listener failing for good must not stop the others:
// a network going away leaves the host listener serving.
func TestMergedListenerOutlivesAFailedListener(t *testing.T) {
	failed := newStubListener(acceptResult{err: errors.New("network gone")})
	conn := pipeConn(t)
	healthy := newStubListener()

	merged := newMergedListener(t.Context(), []net.Listener{healthy, failed})
	t.Cleanup(func() { _ = merged.Close() })

	// Give the failed listener's error time to be taken, and dropped.
	time.Sleep(50 * time.Millisecond)
	healthy.results <- acceptResult{conn: conn}

	got, err := merged.Accept()
	require.NoError(t, err)
	assert.Equal(t, conn, got)
}

// A temporary error reaches the accept loop, which logs it and carries on,
// and the listener that raised it keeps accepting.
func TestMergedListenerPassesTemporaryErrors(t *testing.T) {
	temporary := &net.OpError{Op: "accept", Err: &temporaryError{}}
	conn := pipeConn(t)
	listener := newStubListener(acceptResult{err: temporary}, acceptResult{conn: conn})

	merged := newMergedListener(t.Context(), []net.Listener{listener})
	t.Cleanup(func() { _ = merged.Close() })

	_, err := merged.Accept()
	require.ErrorIs(t, err, temporary)

	got, err := merged.Accept()
	require.NoError(t, err)
	assert.Equal(t, conn, got)
}

// Once every listener has failed, Accept fails for good, with the reason.
func TestMergedListenerFailsWhenEveryListenerHas(t *testing.T) {
	cause := errors.New("network gone")
	merged := newMergedListener(t.Context(), []net.Listener{
		newStubListener(acceptResult{err: cause}),
		newStubListener(acceptResult{err: cause}),
	})
	t.Cleanup(func() { _ = merged.Close() })

	_, err := merged.Accept()
	require.ErrorIs(t, err, cause)
}

func TestMergedListenerClosesEveryListener(t *testing.T) {
	first, second := newStubListener(), newStubListener()

	merged := newMergedListener(t.Context(), []net.Listener{first, second})
	require.NoError(t, merged.Close())

	_, err := merged.Accept()
	require.ErrorIs(t, err, net.ErrClosed)

	for _, listener := range []*stubListener{first, second} {
		select {
		case <-listener.closed:
		default:
			t.Fatal("a listener was left open")
		}
	}
}

type temporaryError struct{}

func (temporaryError) Error() string   { return "temporary" }
func (temporaryError) Timeout() bool   { return false }
func (temporaryError) Temporary() bool { return true }
