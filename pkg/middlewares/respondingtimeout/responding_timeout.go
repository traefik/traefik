// Package respondingtimeout enforces a whole-transaction deadline (client -> proxy -> backend -> proxy -> client) on a router.
package respondingtimeout

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/containous/alice"
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

	// Nested routers: the most restrictive deadline wins,
	// a child router cannot extend the budget set by its parent (or by any upstream deadline source).
	if d, ok := req.Context().Deadline(); ok && d.Before(deadline) {
		deadline = d
	}

	// Bound slow clients on the inbound connection; this replaces the deadlines armed from the entrypoint
	// respondingTimeouts for this request.
	rc := http.NewResponseController(rw)

	// The read deadline is armed only for a request carrying a body, the sole window where it bounds anything:
	// it makes a stalled upload fail fast instead of holding the connection.
	// A body-less request must be left alone: the server has already started a background read on the
	// connection to detect client disconnections, and it reads any failure of that read — a deadline expiry
	// included — as a dead connection, canceling the connection context (net/http connReader.backgroundRead).
	// That context is the parent of every subsequent request context on the connection, so arming the deadline
	// here would make the next request on a keep-alive connection start already canceled.
	// The presence of a body is read from ContentLength, not from req.Body: an upstream middleware may wrap
	// req.Body (the capture middleware does), which defeats a req.Body != http.NoBody sentinel and would arm
	// the deadline on a body-less request. ContentLength is zero exactly when there is no body, -1 when chunked.
	// It is never cleared on return: the server drains the unread request body before flushing the response
	// (net/http chunkWriter.writeHeader), once this handler and its defers have run, and only a live deadline
	// bounds that drain. The server clears it on its own before reading the next request.
	if req.ContentLength != 0 {
		if err := rc.SetReadDeadline(deadline); err != nil {
			logger.Debug().Err(err).Msg("Unable to set read deadline")
		}
	}

	if err := rc.SetWriteDeadline(deadline); err != nil {
		logger.Debug().Err(err).Msg("Unable to set write deadline")
	}

	rewriter := &statusRewriter{ResponseWriter: rw, deadline: deadline}

	defer func() {
		// A hijacked connection no longer belongs to the server, which cleared its deadlines when handing it
		// over (net/http (*conn).hijackLocked): re-arming one here would bound a protocol-switched tunnel.
		if rewriter.hijacked {
			return
		}

		// The response is flushed after this defer (net/http finishRequest), by which time the transaction
		// deadline may have expired: the flush is granted a fresh budget, or a 504 could never reach the
		// client. That budget is the entrypoint writeTimeout, the bound the flush had before this middleware
		// replaced it, and the zero time when the entrypoint has none — which is also what keeps the deadline
		// from leaking onto the next request on the connection, as net/http re-arms it per request only when
		// Server.WriteTimeout > 0.
		var writeDeadline time.Time
		if writeTimeout := entryPointWriteTimeout(req); writeTimeout > 0 {
			writeDeadline = time.Now().Add(writeTimeout)
		}

		if err := rc.SetWriteDeadline(writeDeadline); err != nil {
			logger.Debug().Err(err).Msg("Unable to restore write deadline")
		}
	}()

	if isUpgradeRequest(req) {
		// For upgrade requests the deadline bounds the handshake only and is disarmed at the protocol switch:
		// a deadline must never tear down a healthy tunnel.
		// A disarmable timer replaces context.WithDeadline because a context deadline cannot be un-armed,
		// and once expired it would kill the backend leg of an established tunnel.
		// The cancellation carries context.DeadlineExceeded as its cause because the transport surfaces
		// context.Cause as the round-trip error: a handshake the backend never answers is then reported as a
		// timeout everywhere, including in the tracing span and the metrics derived from that error.
		ctx, cancel := context.WithCancelCause(req.Context())
		defer cancel(nil)

		timer := time.AfterFunc(time.Until(deadline), func() { cancel(context.DeadlineExceeded) })
		defer timer.Stop()

		// The timer has to be disarmed at the protocol switch, not on return: the proxy keeps serving the
		// tunnel until it closes (net/http/httputil handleUpgradeResponse blocks).
		rewriter.onHijack = func() { timer.Stop() }

		h.next.ServeHTTP(rewriter, req.WithContext(ctx))

		return
	}

	// Bound the backend leg and enable an early, clean 504.
	ctx, cancel := context.WithDeadline(req.Context(), deadline)
	defer cancel()

	h.next.ServeHTTP(rewriter, req.WithContext(ctx))
}

// entryPointWriteTimeout returns the writeTimeout of the entrypoint serving req, zero when it has none.
// It cannot be injected into the middleware: the router handler is built once per router and shared by every
// entrypoint the router is attached to (pkg/server/router.Manager). The net/http server publishes itself on
// the request context for exactly this purpose.
func entryPointWriteTimeout(req *http.Request) time.Duration {
	srv, ok := req.Context().Value(http.ServerContextKey).(*http.Server)
	if !ok {
		return 0
	}

	return srv.WriteTimeout
}

// isUpgradeRequest reports whether the request attempts a protocol switch.
// Any Connection: Upgrade protocol counts (e.g. SPDY as used by kubectl exec), not just WebSocket,
// mirroring the stdlib reverse proxy upgrade detection.
func isUpgradeRequest(req *http.Request) bool {
	return httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") &&
		req.Header.Get("Upgrade") != ""
}

// statusRewriter normalizes the response status code to 504 Gateway Timeout when the deadline has expired.
type statusRewriter struct {
	http.ResponseWriter

	deadline time.Time
	// onHijack is set for upgrade requests: it disarms the handshake timer.
	onHijack func()
	// hijacked reports whether the connection has been handed over.
	hijacked bool
}

func (s *statusRewriter) WriteHeader(code int) {
	// Past the deadline, a 5xx can only mean the transaction ran out of budget.
	// 499 is included because the connection read deadline and the context deadline expire at the same
	// instant: when the former wins the race, net/http cancels the request context with context.Canceled,
	// which the proxy error handler reads as a client disconnection. Comparing against the deadline, rather
	// than inspecting context.Cause, keeps the outcome independent of which of the two fired first.
	if (code >= http.StatusInternalServerError || code == httputil.StatusClientClosedRequest) &&
		!time.Now().Before(s.deadline) {
		code = http.StatusGatewayTimeout
	}

	s.ResponseWriter.WriteHeader(code)
}

// Hijack is the protocol-switch commitment point: the deadline is disarmed before the connection is handed over.
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
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (s *statusRewriter) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}
