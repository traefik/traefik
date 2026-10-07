// Package respondingtimeout enforces a whole-transaction deadline (client -> proxy -> backend -> proxy -> client) on a router.
package respondingtimeout

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/containous/alice"
	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"golang.org/x/net/http/httpguts"
)

const (
	name     = "traefik-internal-responding-timeout"
	typeName = "RespondingTimeout"
)

type handler struct {
	next    http.Handler
	timeout time.Duration
}

// WrapHandler wraps a router handler to enforce the given whole-transaction timeout, as an alice.Constructor.
func WrapHandler(timeout time.Duration) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return &handler{next: next, timeout: timeout}, nil
	}
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), name, typeName)

	deadline := time.Now().Add(h.timeout)

	// Nested routers: the most restrictive deadline wins; a child router cannot extend its parent's budget.
	if d, ok := req.Context().Deadline(); ok && d.Before(deadline) {
		deadline = d
	}

	// These connection deadlines replace, for this request, the ones armed from the entrypoint respondingTimeouts.
	rc := http.NewResponseController(rw)

	// Only set when there is a body: without one, net/http reads in the background to detect a client
	// disconnection, and a deadline expiring there cancels the connection context, breaking the next keep-alive requests.
	// ContentLength is checked rather than req.Body, which upstream middlewares may wrap.
	// It is not cleared on return, as it also bounds the server draining an unread body after the handler.
	if req.ContentLength != 0 {
		if err := rc.SetReadDeadline(deadline); err != nil {
			logger.Debug().Err(err).Msg("Unable to set read deadline")
		}
	}

	// The write deadline is set on the first write (see armWriteDeadline), not here: over HTTP/2 it is a timer
	// that resets the stream when it fires, which at the deadline would prevent the 504.
	writeTimeout := entryPointWriteTimeout(req)

	rewriter := &statusRewriter{ResponseWriter: rw, responseController: rc, deadline: deadline, writeTimeout: writeTimeout, logger: logger}

	defer func() {
		// net/http clears the deadlines of a hijacked connection, which no longer belongs to the server.
		if rewriter.hijacked {
			return
		}

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

	if isUpgradeRequest(req) {
		// The deadline only bounds the handshake. A context deadline cannot be disarmed and would close the tunnel,
		// so a timer cancels the context instead, and is stopped when the connection is hijacked.
		// The DeadlineExceeded cause makes the proxy, tracing, and metrics report it as a timeout.
		ctx, cancel := context.WithCancelCause(req.Context())
		defer cancel(nil)

		timer := time.AfterFunc(time.Until(deadline), func() { cancel(context.DeadlineExceeded) })
		defer timer.Stop()

		// Disarmed at the protocol switch, not on return: the proxy keeps serving the tunnel until it closes.
		rewriter.onHijack = func() { timer.Stop() }

		h.next.ServeHTTP(rewriter, req.WithContext(ctx))

		return
	}

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

// isUpgradeRequest reports whether the request asks for a protocol switch (WebSocket, SPDY, ...),
// as detected by the stdlib reverse proxy.
func isUpgradeRequest(req *http.Request) bool {
	return httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") &&
		req.Header.Get("Upgrade") != ""
}

// statusRewriter turns the error status into a 504 once the deadline has passed, and sets the write deadline on the first write.
type statusRewriter struct {
	http.ResponseWriter

	responseController *http.ResponseController
	deadline           time.Time
	writeTimeout       time.Duration
	logger             *zerolog.Logger

	// armOnce guards against the reverse proxy flushing from another goroutine.
	armOnce sync.Once

	// onHijack is set for upgrade requests: it disarms the handshake timer.
	onHijack func()
	// hijacked reports whether the connection has been handed over.
	hijacked bool
}

func (s *statusRewriter) WriteHeader(code int) {
	s.armWriteDeadline()

	// Past the deadline, a 5xx means the request ran out of time. 499 as well: the read deadline expires along with
	// the context one, and when it wins, net/http cancels the request context, which the proxy reports as a client disconnection.
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

// Hijack stops the handshake timer before the connection is handed over.
func (s *statusRewriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if s.onHijack != nil {
		s.onHijack()
	}

	hijacker, ok := s.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("not a hijacker: %T", s.ResponseWriter)
	}

	conn, brw, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, err
	}

	s.hijacked = true

	return conn, brw, nil
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
		if s.hijacked {
			return
		}

		writeDeadline := s.deadline
		if !time.Now().Before(s.deadline) {
			writeDeadline = time.Time{}
			if s.writeTimeout > 0 {
				writeDeadline = time.Now().Add(s.writeTimeout)
			}
		}

		if err := s.responseController.SetWriteDeadline(writeDeadline); err != nil {
			s.logger.Debug().Err(err).Msg("Unable to arm write deadline")
		}
	})
}
