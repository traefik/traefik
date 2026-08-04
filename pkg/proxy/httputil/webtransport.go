package httputil

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/quic-go/quic-go/http3"
	webtransport "github.com/quic-go/webtransport-go"
	"github.com/rs/zerolog/log"
)

type wtServerContextKey struct{}
type h3RWContextKey struct{}

// SetWebTransportServer stores the webtransport.Server in ctx so that proxy
// handlers downstream can call Upgrade without holding a direct reference.
func SetWebTransportServer(ctx context.Context, s *webtransport.Server) context.Context {
	return context.WithValue(ctx, wtServerContextKey{}, s)
}

// GetWebTransportServer retrieves the webtransport.Server previously stored by
// SetWebTransportServer. Returns nil when no server is in ctx.
func GetWebTransportServer(ctx context.Context) *webtransport.Server {
	s, _ := ctx.Value(wtServerContextKey{}).(*webtransport.Server)
	return s
}

// SetHTTP3ResponseWriter stores the raw http3.ResponseWriter in ctx before the
// middleware chain wraps it. Retrieved by the proxy handler to call Upgrade.
func SetHTTP3ResponseWriter(ctx context.Context, rw http.ResponseWriter) context.Context {
	return context.WithValue(ctx, h3RWContextKey{}, rw)
}

// GetHTTP3ResponseWriter retrieves the raw http3.ResponseWriter stored by
// SetHTTP3ResponseWriter. Returns nil when not in ctx.
func GetHTTP3ResponseWriter(ctx context.Context) http.ResponseWriter {
	rw, _ := ctx.Value(h3RWContextKey{}).(http.ResponseWriter)
	return rw
}

// IsWebTransportRequest reports whether r is an HTTP/3 extended-CONNECT request
// that initiates a WebTransport session (RFC 9220 + draft-ietf-webtrans-http3).
// quic-go maps the :protocol pseudo-header to r.Proto for extended CONNECT.
// Both the current ("webtransport-h3") and legacy ("webtransport") protocol
// strings are accepted to maximise client compatibility.
func IsWebTransportRequest(r *http.Request) bool {
	return r.Method == http.MethodConnect &&
		(r.Proto == "webtransport" || r.Proto == "webtransport-h3")
}

// webTransportProxyHandler intercepts WebTransport sessions before the regular
// HTTP reverse proxy can touch them, proxies them to the backend via a new
// WebTransport session, and bridges all streams and datagrams bidirectionally.
// Non-WebTransport requests are passed unchanged to next.
type webTransportProxyHandler struct {
	next      http.Handler
	targetURL *url.URL
	transport *webtransport.Transport
}

// newWebTransportProxyHandler wraps next with WebTransport session proxying for
// the given target. tlsConfig is used when dialing the backend; a nil
// tlsConfig falls back to sensible defaults (no client cert, system root CAs).
func newWebTransportProxyHandler(next http.Handler, targetURL *url.URL, tlsConfig *tls.Config) http.Handler {
	return &webTransportProxyHandler{
		next:      next,
		targetURL: targetURL,
		transport: &webtransport.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
}

// ServeHTTP implements http.Handler.
func (h *webTransportProxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !IsWebTransportRequest(req) {
		h.next.ServeHTTP(rw, req)
		return
	}

	wtServer := GetWebTransportServer(req.Context())
	if wtServer == nil {
		// HTTP/3 + WebTransport is not enabled on this entrypoint; fall through
		// so the request reaches the regular CONNECT handler and fails with a
		// meaningful error rather than silently doing nothing.
		h.next.ServeHTTP(rw, req)
		return
	}

	logger := log.Ctx(req.Context())

	// webtransport.Server.Upgrade() requires the response writer to implement
	// http3.Settingser and http3.HTTPStreamer. Traefik's middleware chain wraps
	// the response writer, hiding those interfaces. The raw http3.ResponseWriter
	// is stored in the request context by the HTTP/3 entrypoint handler before
	// any middleware can wrap it; use that writer for Upgrade.
	upgradeRW := GetHTTP3ResponseWriter(req.Context())
	if upgradeRW == nil {
		// Fallback: walk the Unwrap chain (works when all wrappers implement Unwrap).
		upgradeRW = unwrapToHTTP3RW(rw)
	}
	if upgradeRW == nil {
		ErrorHandlerWithContext(req.Context(), rw, fmt.Errorf("WebTransport not available: HTTP/3 response writer not accessible"))
		return
	}

	clientSession, err := wtServer.Upgrade(upgradeRW, req)
	if err != nil {
		ErrorHandlerWithContext(req.Context(), rw, fmt.Errorf("accepting WebTransport session: %w", err))
		return
	}

	// Build the backend URL from the configured target, preserving the
	// incoming path and query so the backend can route by URL.
	backendURL := *h.targetURL
	backendURL.Path = req.URL.Path
	backendURL.RawQuery = req.URL.RawQuery

	_, backendSession, err := h.transport.Dial(req.Context(), backendURL.String(), nil)
	if err != nil {
		_ = clientSession.CloseWithError(0, "backend unavailable")
		logger.Error().Err(err).Str("backend", backendURL.String()).Msg("WebTransport backend dial failed")
		return
	}

	logger.Debug().Str("backend", backendURL.String()).Msg("WebTransport session established")

	proxyWebTransportSession(req.Context(), clientSession, backendSession)
}

// unwrapToHTTP3RW walks the Unwrap chain of rw looking for an
// http.ResponseWriter that implements http3.Settingser and http3.HTTPStreamer.
// Returns nil if none is found (e.g. when wrappers don't implement Unwrap).
func unwrapToHTTP3RW(rw http.ResponseWriter) http.ResponseWriter {
	for {
		if isHTTP3ResponseWriter(rw) {
			return rw
		}
		u, ok := rw.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			return nil
		}
		rw = u.Unwrap()
	}
}

// isHTTP3ResponseWriter reports whether rw implements both http3-specific
// interfaces that webtransport.Server.Upgrade requires.
func isHTTP3ResponseWriter(rw http.ResponseWriter) bool {
	_, okS := rw.(http3.Settingser)
	_, okH := rw.(http3.HTTPStreamer)
	return okS && okH
}

// proxyWebTransportSession bridges client and backend WebTransport sessions
// until either side closes or ctx is cancelled. All bidirectional streams,
// unidirectional streams, and datagrams are forwarded in both directions.
func proxyWebTransportSession(ctx context.Context, client, backend *webtransport.Session) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer func() { _ = client.CloseWithError(0, "") }()
	defer func() { _ = backend.CloseWithError(0, "") }()

	// Datagrams: forward in both directions concurrently.
	go proxyDatagrams(ctx, client, backend)
	go proxyDatagrams(ctx, backend, client)

	// Unidirectional streams opened by each side.
	go proxyUniStreams(ctx, client, backend)
	go proxyUniStreams(ctx, backend, client)

	// Bidirectional streams: accept from each side and open a matching stream
	// on the other. The two goroutines race; we wait for ctx cancellation (i.e.
	// either session closes) rather than for a particular goroutine to finish.
	done := make(chan struct{}, 2)
	go func() { proxyBidiStreams(ctx, client, backend); done <- struct{}{} }()
	go func() { proxyBidiStreams(ctx, backend, client); done <- struct{}{} }()

	// Block until one side is done or the context is cancelled.
	select {
	case <-done:
	case <-ctx.Done():
	}
}

// proxyBidiStreams accepts every bidirectional stream opened by src, opens a
// matching stream on dst, and copies data in both directions.
func proxyBidiStreams(ctx context.Context, src, dst *webtransport.Session) {
	for {
		srcStream, err := src.AcceptStream(ctx)
		if err != nil {
			return
		}

		go func() {
			dstStream, err := dst.OpenStreamSync(ctx)
			if err != nil {
				srcStream.CancelRead(0)
				srcStream.CancelWrite(0)
				return
			}
			proxyBidiStream(ctx, srcStream, dstStream)
		}()
	}
}

// proxyBidiStream proxies a bidirectional stream between a (proxy↔client side)
// and b (proxy↔backend side), forwarding FIN signals so that half-close and
// graceful echo patterns work correctly.
//
// Each direction is copied independently:
//   - client→backend: when a's read half closes (client FIN), b's write half is
//     closed to forward the FIN to the backend.
//   - backend→client: when b's read half closes (backend FIN / echo done),
//     a's write half is closed to deliver the FIN to the client.
//
// If ctx is cancelled before both copies finish, both streams are hard-reset.
func proxyBidiStream(ctx context.Context, a, b *webtransport.Stream) {
	done := make(chan struct{})

	go func() {
		defer close(done)
		var wg sync.WaitGroup
		wg.Add(2)

		// client → backend: forward FIN when client closes its write half.
		go func() {
			defer wg.Done()
			_, _ = io.Copy(b, a)
			_ = b.Close()
		}()

		// backend → client: forward FIN when backend closes its write half.
		go func() {
			defer wg.Done()
			_, _ = io.Copy(a, b)
			_ = a.Close()
		}()

		wg.Wait()
	}()

	select {
	case <-done:
		// Both halves completed cleanly — nothing to cancel.
	case <-ctx.Done():
		// Context cancelled: hard-reset both streams and wait for goroutines.
		a.CancelRead(0)
		a.CancelWrite(0)
		b.CancelRead(0)
		b.CancelWrite(0)
		<-done
	}
}

// proxyUniStreams accepts every unidirectional (receive) stream opened by src
// and pipes it to a new unidirectional (send) stream opened on dst.
func proxyUniStreams(ctx context.Context, src, dst *webtransport.Session) {
	for {
		recvStream, err := src.AcceptUniStream(ctx)
		if err != nil {
			return
		}

		go func() {
			sendStream, err := dst.OpenUniStreamSync(ctx)
			if err != nil {
				recvStream.CancelRead(0)
				return
			}
			_, _ = io.Copy(sendStream, recvStream)
			_ = sendStream.Close()
		}()
	}
}

// proxyDatagrams forwards datagrams from src to dst until ctx is cancelled or
// either session closes.
func proxyDatagrams(ctx context.Context, src, dst *webtransport.Session) {
	for {
		data, err := src.ReceiveDatagram(ctx)
		if err != nil {
			return
		}

		if err := dst.SendDatagram(data); err != nil {
			return
		}
	}
}
