// Package respondingtimeout enforces the read and write timeouts of a router, counted from the router receiving the request.
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
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"golang.org/x/net/http/httpguts"
)

const (
	name     = "traefik-internal-responding-timeout"
	typeName = "RespondingTimeout"
)

// readDeadlineKey carries the read deadline set by a parent router: unlike the write one, it has no context deadline.
type readDeadlineKey struct{}

type handler struct {
	next http.Handler
	// readTimeout and writeTimeout are nil when unset: the entrypoint ones then apply.
	readTimeout  *time.Duration
	writeTimeout *time.Duration
}

// WrapHandler wraps a router handler to enforce the given responding timeouts, as an alice.Constructor.
func WrapHandler(config dynamic.RouterRespondingTimeouts) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
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

	if h.readTimeout != nil {
		req = h.armReadDeadline(req, rc, start, logger)
	}

	switch {
	// Upgrade requests (WebSocket, ...) are left to the entrypoint writeTimeout:
	// the proxy closes the tunnel when the request context is done, which a router deadline would do.
	case h.writeTimeout == nil || isUpgradeRequest(req):
		h.next.ServeHTTP(rw, req)
	case *h.writeTimeout <= 0:
		h.serveWithoutWriteTimeout(rw, req, rc, logger)
	default:
		h.serveWithWriteTimeout(rw, req, rc, start.Add(*h.writeTimeout), logger)
	}
}

// armReadDeadline sets the read deadline, or lifts the entrypoint one for a zero readTimeout,
// and records it in the returned request for nested routers.
func (h *handler) armReadDeadline(req *http.Request, rc *http.ResponseController, start time.Time, logger *zerolog.Logger) *http.Request {
	var deadline time.Time
	if *h.readTimeout > 0 {
		deadline = start.Add(*h.readTimeout)
	}

	// Nested routers: the most restrictive deadline wins; a child router cannot extend its parent's budget.
	if parent, ok := req.Context().Value(readDeadlineKey{}).(time.Time); ok && !parent.IsZero() &&
		(deadline.IsZero() || parent.Before(deadline)) {
		deadline = parent
	}

	// Only set when there is a body: without one, net/http reads in the background to detect a client
	// disconnection, and a deadline expiring there cancels the connection context, breaking the next keep-alive requests.
	// ContentLength is checked rather than req.Body, which upstream middlewares may wrap.
	// It is not cleared on return, as it also bounds the server draining an unread body after the handler.
	if req.ContentLength != 0 {
		if err := rc.SetReadDeadline(deadline); err != nil {
			logger.Debug().Err(err).Msg("Unable to set read deadline")
		}
	}

	return req.WithContext(context.WithValue(req.Context(), readDeadlineKey{}, deadline))
}

func (h *handler) serveWithWriteTimeout(rw http.ResponseWriter, req *http.Request, rc *http.ResponseController, deadline time.Time, logger *zerolog.Logger) {
	// Nested routers: the most restrictive deadline wins; a child router cannot extend its parent's budget.
	if d, ok := req.Context().Deadline(); ok && d.Before(deadline) {
		deadline = d
	}

	writeTimeout := entryPointWriteTimeout(req)

	// Over HTTP/2, the entrypoint writeTimeout is a timer started with the stream, which resets it even before
	// anything is written. Moving it past the deadline keeps it from cutting a slow backend short or preventing the 504.
	// The actual write deadline is set on the first write (see armWriteDeadline).
	if writeTimeout > 0 {
		if err := rc.SetWriteDeadline(deadline.Add(writeTimeout)); err != nil {
			logger.Debug().Err(err).Msg("Unable to extend write deadline")
		}
	}

	rewriter := &statusRewriter{ResponseWriter: rw, responseController: rc, deadline: deadline, writeTimeout: writeTimeout, logger: logger}

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

// serveWithoutWriteTimeout lifts the entrypoint write deadline for this request.
// net/http resets the connection deadlines for the next request, so it is not restored on return.
func (h *handler) serveWithoutWriteTimeout(rw http.ResponseWriter, req *http.Request, rc *http.ResponseController, logger *zerolog.Logger) {
	// Nested routers: a child router cannot extend its parent's budget.
	if _, ok := req.Context().Deadline(); ok {
		h.next.ServeHTTP(rw, req)
		return
	}

	// Over HTTP/2, this also stops the entrypoint writeTimeout stream timer.
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		logger.Debug().Err(err).Msg("Unable to clear write deadline")
	}

	h.next.ServeHTTP(rw, req)
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
