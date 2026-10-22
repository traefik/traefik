// Package respondingtimeout enforces the read and write timeouts of a router, counted from the router receiving the request.
package respondingtimeout

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/containous/alice"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"golang.org/x/net/http/httpguts"
)

const (
	name     = "traefik-internal-responding-timeout"
	typeName = "RespondingTimeout"
)

// timeoutResponseGrace bounds writing a response once the write deadline has passed, which is the 504 in practice.
const timeoutResponseGrace = 5 * time.Second

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

		// Both count from the same instant, and the request ends at the writeTimeout: a longer readTimeout has no effect.
		// A zero writeTimeout means no timeout, and an unset one is the entrypoint one, unknown here.
		if config.ReadTimeout != nil && config.WriteTimeout != nil && *config.WriteTimeout > 0 && *config.ReadTimeout > *config.WriteTimeout {
			return nil, fmt.Errorf("readTimeout (%s) must not be longer than writeTimeout (%s)", *config.ReadTimeout, *config.WriteTimeout)
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
	// Both timeouts count from this instant: equal timeouts share a single deadline.
	start := time.Now()

	// These connection deadlines replace, for this request, the ones armed from the entrypoint respondingTimeouts.
	rc := http.NewResponseController(rw)

	if h.readTimeout != nil && req.ContentLength != 0 {
		switch {
		case *h.readTimeout <= 0:
			_ = rc.SetReadDeadline(time.Time{})
		default:
			// It is not cleared on return, as it also bounds the server draining an unread body after the handler.
			_ = rc.SetReadDeadline(start.Add(*h.readTimeout))

			body := &deadlineBody{ReadCloser: req.Body, responseController: rc}
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
		_ = rc.SetWriteDeadline(time.Time{})
		h.next.ServeHTTP(rw, req)
	default:
		h.serveWithWriteTimeout(rw, req, rc, start.Add(*h.writeTimeout))
	}
}

func (h *handler) serveWithWriteTimeout(rw http.ResponseWriter, req *http.Request, rc *http.ResponseController, deadline time.Time) {
	// Until the first write, the context deadline bounds the wait for the backend, and the write deadline is cleared:
	// it is set on the first write (see armWriteDeadline). Over HTTP/2, this also stops the entrypoint writeTimeout
	// stream timer, which resets the stream when it fires, even before anything is written.
	_ = rc.SetWriteDeadline(time.Time{})

	defer func() {
		// net/http only writes a buffered response after the handler returns: past the deadline,
		// it is written here instead, so that the grace period bounds it.
		if time.Now().After(deadline) {
			_ = rc.SetWriteDeadline(time.Now().Add(timeoutResponseGrace))

			if err := rc.Flush(); err != nil {
				logger := middlewares.GetLogger(req.Context(), name, typeName)
				logger.Error().Err(err).Msg("Unable to flush response after the write timeout was exceeded")
			}
		}

		// net/http only resets the write deadline for the next keep-alive request when the entrypoint has a writeTimeout:
		// left set, it would cut the responses to the next requests.
		_ = rc.SetWriteDeadline(time.Time{})
	}()

	rewriter := &statusRewriter{ResponseWriter: rw, responseController: rc, deadline: deadline}

	// Cancels the backend request at the deadline, which the proxy reports as a 504.
	ctx, cancel := context.WithDeadline(req.Context(), deadline)
	defer cancel()

	h.next.ServeHTTP(rewriter, req.WithContext(ctx))
}

// deadlineBody clears the read deadline once the request body has been read to the end.
// net/http then reads the connection in the background to detect a client disconnection, and a read deadline
// expiring there cancels the request context. net/http clears it when starting that read, but the body may have been
// read before the router, by a parent router middleware buffering it, in which case the deadline was set afterwards.
type deadlineBody struct {
	io.ReadCloser

	responseController *http.ResponseController

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

	_ = b.responseController.SetReadDeadline(time.Time{})
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

	responseController *http.ResponseController
	deadline           time.Time

	// armOnce guards against the reverse proxy flushing from another goroutine.
	armOnce sync.Once
}

func (s *statusRewriter) WriteHeader(code int) {
	s.armWriteDeadline()

	// Past the deadline, a 5xx means the request ran out of time. 499 as well: when the read deadline expires along with
	// the context one, and wins, net/http cancels the request context, which the proxy reports as a client disconnection.
	if time.Now().After(s.deadline) && (code >= http.StatusInternalServerError || code == httputil.StatusClientClosedRequest) {
		code = http.StatusGatewayTimeout
	}

	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRewriter) Write(b []byte) (int, error) {
	s.armWriteDeadline()

	return s.ResponseWriter.Write(b)
}

// FlushError arms the write deadline before flushing, and is preferred over Flush by http.ResponseController.
func (s *statusRewriter) FlushError() error {
	s.armWriteDeadline()

	return s.responseController.Flush()
}

func (s *statusRewriter) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}

// armWriteDeadline sets the write deadline once, on the first write: the response must be sent by the deadline.
// Past it (the 504 case), the deadline cleared before serving is kept, and the grace period is applied on return:
// a deadline in the past would fail the write, and reset the stream at once over HTTP/2.
func (s *statusRewriter) armWriteDeadline() {
	s.armOnce.Do(func() {
		if time.Now().After(s.deadline) {
			return
		}

		_ = s.responseController.SetWriteDeadline(s.deadline)
	})
}
