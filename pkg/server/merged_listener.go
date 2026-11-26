package server

import (
	"context"
	"errors"
	"net"
	"net/url"
	"sync"

	"github.com/rs/zerolog/log"
)

// mergedListener accepts on several listeners as one, which is how an
// entryPoint takes connections from its own address and from the sources a
// network extension gives it alike.
//
// A listener that fails for good is dropped rather than taking the others
// down with it: a network going away must not stop the host listener. Accept
// fails for good only once every listener has.
type mergedListener struct {
	listeners []net.Listener
	addr      net.Addr

	conns chan net.Conn
	errs  chan error

	closeOnce sync.Once
	done      chan struct{}

	// exhausted is closed once every listener has stopped, with lastErr the
	// reason the last of them gave.
	exhausted chan struct{}
	lastErrMu sync.Mutex
	lastErr   error
}

// newMergedListener accepts on every listener, reporting the first one's
// address as its own: that is the entryPoint's own address.
func newMergedListener(ctx context.Context, listeners []net.Listener) *mergedListener {
	m := &mergedListener{
		listeners: listeners,
		addr:      listeners[0].Addr(),
		conns:     make(chan net.Conn),
		errs:      make(chan error),
		done:      make(chan struct{}),
		exhausted: make(chan struct{}),
	}

	var wg sync.WaitGroup
	for _, ln := range listeners {
		wg.Go(func() { m.serve(ctx, ln) })
	}
	go func() {
		wg.Wait()
		close(m.exhausted)
	}()

	return m
}

func (m *mergedListener) Accept() (net.Conn, error) {
	select {
	case conn := <-m.conns:
		// A connection and the close can both be ready, and select takes either:
		// without this, a closed listener could still hand a connection out.
		select {
		case <-m.done:
			_ = conn.Close()
			return nil, net.ErrClosed
		default:
			return conn, nil
		}
	case err := <-m.errs:
		return nil, err
	case <-m.done:
		return nil, net.ErrClosed
	case <-m.exhausted:
		m.lastErrMu.Lock()
		defer m.lastErrMu.Unlock()
		if m.lastErr == nil {
			return nil, net.ErrClosed
		}
		return nil, m.lastErr
	}
}

func (m *mergedListener) Close() error {
	var errs []error
	m.closeOnce.Do(func() {
		close(m.done)
		for _, ln := range m.listeners {
			if err := ln.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				errs = append(errs, err)
			}
		}
	})
	return errors.Join(errs...)
}

func (m *mergedListener) Addr() net.Addr { return m.addr }

// serve accepts on one listener until it fails for good or the merged
// listener is closed. A temporary error is handed to Accept, whose caller
// logs it and carries on, exactly as it would for a single listener.
func (m *mergedListener) serve(ctx context.Context, ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err == nil {
			select {
			case m.conns <- conn:
			case <-m.done:
				_ = conn.Close()
				return
			}
			continue
		}

		select {
		case <-m.done:
			return
		default:
		}

		if isTemporaryAcceptError(err) {
			select {
			case m.errs <- err:
			case <-m.done:
				return
			}
			continue
		}

		log.Ctx(ctx).Error().Err(err).Stringer("address", ln.Addr()).
			Msg("Listener stopped; the entryPoint keeps accepting on its other listeners")

		m.lastErrMu.Lock()
		m.lastErr = err
		m.lastErrMu.Unlock()
		return
	}
}

// isTemporaryAcceptError reports whether an Accept error is one the
// entryPoint's accept loop retries after, rather than stopping on.
func isTemporaryAcceptError(err error) bool {
	if opErr, ok := errors.AsType[*net.OpError](err); ok && opErr.Temporary() {
		return true
	}
	if urlErr, ok := errors.AsType[*url.Error](err); ok && urlErr.Temporary() {
		return true
	}
	return false
}
