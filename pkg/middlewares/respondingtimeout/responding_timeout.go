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

	// Nested routers: the most restrictive deadline wins; a child router cannot extend its parent's budget.
	if d, ok := req.Context().Deadline(); ok && d.Before(deadline) {
		deadline = d
	}

	// These connection deadlines replace, for this request, the ones armed from the entrypoint respondingTimeouts.
	rc := http.NewResponseController(rw)

	// The read deadline is armed only for a request carrying a body, to bound a stalled upload.
	// It must not be set on a body-less request: net/http has started a background read to detect client
	// disconnection, and treats a deadline expiry as a dead connection, canceling the connection context that
	// parents every later request on the keep-alive connection.
	// Body presence comes from ContentLength (0 = none, -1 = chunked), not req.Body, because an upstream
	// middleware may wrap req.Body (capture does) and defeat a req.Body != http.NoBody check.
	// It is not cleared on return: the server drains the unread body before flushing the response, after this
	// handler returns, and only a live deadline bounds that drain (the server clears it itself before the next request).
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
		// A hijacked connection no longer belongs to the server, which cleared its deadlines when handing it over:
		// re-arming one here would bound a protocol-switched tunnel.
		if rewriter.hijacked {
			return
		}

		// The response is flushed after this defer runs, possibly past an expired deadline; restoring a live budget
		// lets a pending 504 reach the client. That budget is the entrypoint writeTimeout — the bound the flush had
		// before this middleware replaced it — or no deadline when the entrypoint has none.
		var writeDeadline time.Time
		if writeTimeout := entryPointWriteTimeout(req); writeTimeout > 0 {
			writeDeadline = time.Now().Add(writeTimeout)
		}

		if err := rc.SetWriteDeadline(writeDeadline); err != nil {
			logger.Debug().Err(err).Msg("Unable to restore write deadline")
		}
	}()

	if isUpgradeRequest(req) {
		// For an upgrade the deadline bounds the handshake only: a disarmable timer replaces context.WithDeadline,
		// which cannot be un-armed and would tear down an established tunnel at expiry.
		// It cancels with context.DeadlineExceeded as its cause so that a handshake the backend never answers is
		// reported as a timeout everywhere the transport surfaces the cause: client status, tracing span, and metrics.
		ctx, cancel := context.WithCancelCause(req.Context())
		defer cancel(nil)

		timer := time.AfterFunc(time.Until(deadline), func() { cancel(context.DeadlineExceeded) })
		defer timer.Stop()

		// Disarmed at the protocol switch, not on return: the proxy keeps serving the tunnel until it closes.
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
	// Past the deadline, a 5xx means the transaction ran out of budget. 499 is included because the read and
	// context deadlines expire together: when the read deadline wins, net/http cancels the request context with
	// context.Canceled, which the proxy error handler reads as a client disconnection. Comparing against the
	// deadline, rather than the cancellation cause, keeps the outcome independent of which of the two fired first.
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
