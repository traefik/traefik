// Package respondingtimeout enforces the read and write timeouts of a router, counted from the router receiving the request.
package respondingtimeout

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/containous/alice"
	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"golang.org/x/net/http/httpguts"
)

const (
	name     = "traefik-internal-responding-timeout"
	typeName = "RespondingTimeout"
)

type handler struct {
	next http.Handler
	// readTimeout and writeTimeout are nil when unset: the entrypoint ones then apply.
	readTimeout  *time.Duration
	writeTimeout *time.Duration
}

// WrapHandler wraps a router handler to enforce the given responding timeouts, as an alice.Constructor.
func WrapHandler(config dynamic.RouterRespondingTimeouts) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		if config.ReadTimeout == nil && config.WriteTimeout == nil {
			return next, nil
		}

		h := &handler{next: next}

		if config.ReadTimeout != nil {
			h.readTimeout = new(time.Duration(*config.ReadTimeout))
		}

		if config.WriteTimeout != nil {
			h.writeTimeout = new(time.Duration(*config.WriteTimeout))
		}

		return h, nil
	}
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), name, typeName)

	// Both timeouts count from this instant: equal timeouts share a single deadline.
	start := time.Now()

	// These connection deadlines replace, for this request, the ones armed from the entrypoint respondingTimeouts.
	rc := http.NewResponseController(rw)

	if h.readTimeout != nil && req.ContentLength != 0 {
		switch {
		case *h.readTimeout <= 0:
			if err := rc.SetReadDeadline(time.Time{}); err != nil {
				logger.Error().Err(err).Msg("Unable to clear read deadline")
			}
		default:
			// It is not cleared on return, as it also bounds the server draining an unread body after the handler.
			if err := rc.SetReadDeadline(start.Add(*h.readTimeout)); err != nil {
				logger.Error().Err(err).Msg("Unable to set read deadline")
			}

			body := &deadlineBody{ReadCloser: req.Body, responseController: rc, logger: logger}
			defer body.release()

			req.Body = body
		}
	}

	switch {
	// Upgrade requests (WebSocket, ...) are left to the entrypoint writeTimeout:
	// the proxy closes the tunnel when the request context is done, which a router deadline would do.
	case h.writeTimeout == nil || isUpgradeRequest(req):
		h.next.ServeHTTP(rw, req)
	case *h.writeTimeout <= 0:
		// Over HTTP/2, this also stops the entrypoint writeTimeout stream timer.
		// net/http resets the connection deadlines for the next request, so it is not restored on return.
		if err := rc.SetWriteDeadline(time.Time{}); err != nil {
			logger.Debug().Err(err).Msg("Unable to clear write deadline")
		}

		h.next.ServeHTTP(rw, req)
	default:
		h.serveWithWriteTimeout(rw, req, rc, start.Add(*h.writeTimeout), logger)
	}
}

func (h *handler) serveWithWriteTimeout(rw http.ResponseWriter, req *http.Request, rc *http.ResponseController, deadline time.Time, logger *zerolog.Logger) {
	writeTimeout := entryPointWriteTimeout(req)

	// Over HTTP/2, the entrypoint writeTimeout is a timer started with the stream, which resets it even before
	// anything is written. Moving it past the deadline keeps it from cutting a slow backend short or preventing the 504.
	// The actual write deadline is set on the first write (see armWriteDeadline).
	if writeTimeout > 0 {
		if err := rc.SetWriteDeadline(deadline.Add(writeTimeout)); err != nil {
			logger.Debug().Err(err).Msg("Unable to extend write deadline")
		}
	}

	rewriter := &statusRewriter{ResponseWriter: rw, responseController: rc, deadline: deadline, entrypointWriteTimeout: writeTimeout, logger: logger}

	defer func() {
		// The response is flushed after the handler returns, possibly past the deadline:
		// restore the entrypoint writeTimeout, or no deadline, so that it still reaches the client.
		var writeDeadline time.Time
		if writeTimeout > 0 {
			writeDeadline = time.Now().Add(writeTimeout)
		}

		if err := rc.SetWriteDeadline(writeDeadline); err != nil {
			logger.Debug().Err(err).Msg("Unable to restore write deadline")
		}
	}()

	// Cancels the backend request at the deadline, which the proxy reports as a 504.
	ctx, cancel := context.WithDeadline(req.Context(), deadline)
	defer cancel()

	h.next.ServeHTTP(rewriter, req.WithContext(ctx))
}

// entryPointWriteTimeout returns the writeTimeout of the entrypoint serving req, or zero.
// It is read from the request because a router handler is shared by all the entrypoints of the router.
func entryPointWriteTimeout(req *http.Request) time.Duration {
	srv, ok := req.Context().Value(http.ServerContextKey).(*http.Server)
	if !ok {
		return 0
	}

	return srv.WriteTimeout
}

// deadlineBody clears the read deadline once the request body has been read to the end.
// net/http then reads the connection in the background to detect a client disconnection, and a read deadline
// expiring there cancels the request context. net/http clears it when starting that read, but the body may have been
// read before the router, by a parent router middleware buffering it, in which case the deadline was set afterwards.
type deadlineBody struct {
	io.ReadCloser

	responseController *http.ResponseController
	logger             *zerolog.Logger

	mu sync.Mutex
	// released is set once the handler has returned: the httputil proxy transport may still read the body from
	// another goroutine, while the ResponseWriter must not be used anymore (it panics over HTTP/2,
	// and would clear the next request read deadline over HTTP/1).
	released bool
	cleared  bool
}

func (b *deadlineBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)

	// We want to clear the deadline as the net/http server would do,
	// if the body we received was not from the Go server.
	if errors.Is(err, io.EOF) {
		b.clearReadDeadline()
	}

	return n, err
}

func (b *deadlineBody) clearReadDeadline() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.released || b.cleared {
		return
	}

	b.cleared = true

	if err := b.responseController.SetReadDeadline(time.Time{}); err != nil {
		b.logger.Error().Err(err).Msg("Unable to clear read deadline")
	}
}

func (b *deadlineBody) release() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.released = true
}

// isUpgradeRequest reports whether the request asks for a protocol switch (WebSocket, SPDY, ...),
// as detected by the stdlib reverse proxy.
func isUpgradeRequest(req *http.Request) bool {
	return httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") &&
		req.Header.Get("Upgrade") != ""
}

// statusRewriter turns the error status into a 504 once the deadline has passed, and sets the write deadline on the first write.
type statusRewriter struct {
	http.ResponseWriter

	responseController     *http.ResponseController
	deadline               time.Time
	entrypointWriteTimeout time.Duration
	logger                 *zerolog.Logger

	// armOnce guards against the reverse proxy flushing from another goroutine.
	armOnce sync.Once
}

func (s *statusRewriter) WriteHeader(code int) {
	s.armWriteDeadline()

	// Past the deadline, a 5xx means the request ran out of time. 499 as well: when the read deadline expires along with
	// the context one, and wins, net/http cancels the request context, which the proxy reports as a client disconnection.
	if (code >= http.StatusInternalServerError || code == httputil.StatusClientClosedRequest) &&
		!time.Now().Before(s.deadline) {
		code = http.StatusGatewayTimeout
	}

	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRewriter) Write(b []byte) (int, error) {
	s.armWriteDeadline()

	return s.ResponseWriter.Write(b)
}

// Hijack is exposed for the handlers asserting http.Hijacker directly, which do not go through Unwrap.
func (s *statusRewriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := s.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("not a hijacker: %T", s.ResponseWriter)
	}

	return hijacker.Hijack()
}

func (s *statusRewriter) Flush() {
	s.armWriteDeadline()

	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (s *statusRewriter) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}

// armWriteDeadline sets the write deadline once, on the first write.
// Before the deadline, the response must be sent by the deadline.
// Past it (the 504 case), a deadline in the past would fail the write, and reset the stream at once over HTTP/2:
// the response gets the entrypoint writeTimeout instead, or no deadline.
func (s *statusRewriter) armWriteDeadline() {
	s.armOnce.Do(func() {
		writeDeadline := s.deadline
		if time.Now().After(s.deadline) {
			writeDeadline = time.Time{}
			if s.entrypointWriteTimeout > 0 {
				writeDeadline = time.Now().Add(s.entrypointWriteTimeout)
			}
		}

		if err := s.responseController.SetWriteDeadline(writeDeadline); err != nil {
			s.logger.Error().Err(err).Msg("Unable to arm write deadline")
		}
	})
}
