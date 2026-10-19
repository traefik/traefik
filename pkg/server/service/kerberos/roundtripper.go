package kerberos

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"sync"
)

// stickyRoundTrippers is keyed by owner, as there is one roundTripper per ServersTransport and a dedicated
// round tripper carries the whole transport configuration (TLS client configuration, timeouts, connection pool) of
// the ServersTransport that created it, hence cannot be shared with another one.
// The map is lazily allocated, as most client connections never reach a Kerberos or NTLM backend.
type stickyRoundTrippers struct {
	mu            sync.Mutex
	roundTrippers map[*roundTripper]http.RoundTripper
}

func (s *stickyRoundTrippers) get(owner *roundTripper) http.RoundTripper {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.roundTrippers[owner]
}

// stick returns the dedicated round tripper for the given owner, creating and storing it on the first call
// and returning the existing one afterwards. It keeps the round tripper already stored, as concurrent requests
// on the same client connection (HTTP/2 streams) can race on the same owner, and replacing an in-use round
// tripper would orphan the backend connections it holds.
func (s *stickyRoundTrippers) stick(owner *roundTripper) http.RoundTripper {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rt, ok := s.roundTrippers[owner]; ok {
		return rt
	}

	rt := owner.new()

	if s.roundTrippers == nil {
		s.roundTrippers = make(map[*roundTripper]http.RoundTripper)
	}
	s.roundTrippers[owner] = rt

	return rt
}

type transportKeyType string

var transportKey transportKeyType = "transport"

// AddTransportOnContext returns a context holding the round trippers dedicated to a client connection.
func AddTransportOnContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, transportKey, &stickyRoundTrippers{})
}

type roundTripper struct {
	new      func() http.RoundTripper
	original http.RoundTripper
}

// NewRoundTripper creates a round tripper sharing original between client connections,
// and calling newRoundTripper to create the round tripper dedicated to a client connection.
func NewRoundTripper(original http.RoundTripper, newRoundTripper func() http.RoundTripper) http.RoundTripper {
	return &roundTripper{
		new:      newRoundTripper,
		original: original,
	}
}

func (k *roundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	connRoundTrippers, ok := request.Context().Value(transportKey).(*stickyRoundTrippers)
	if !ok {
		// This should never happen, as the context is set in the entrypoint handler,
		// but we fall back to the original round tripper to avoid panics.
		return k.original.RoundTrip(request)
	}

	// Once a dedicated round tripper has been stuck for this owner, every subsequent request on the same
	// client connection reuses it, whether it is a concurrent request (HTTP/2 streams) or a later
	// credential-less request relying on the connection-bound NTLM/Kerberos authentication.
	if rt := connRoundTrippers.get(k); rt != nil {
		return rt.RoundTrip(request)
	}

	// NTLM and Kerberos (Negotiate) are connection-bound authentication schemes: the request carrying
	// the credential authenticates the backend connection it is dispatched on, and the backend
	// keeps that connection bound to the authenticated identity for reuse.
	// Such a request must never be dispatched through the shared original round tripper, otherwise the
	// now authenticated connection would stay in the globally shared pool where an unrelated frontend
	// connection could later be handed the same socket and inherit the identity.
	if containsNTLMorNegotiate(request.Header.Values("Authorization")) {
		return connRoundTrippers.stick(k).RoundTrip(request)
	}

	return k.original.RoundTrip(request)
}

func containsNTLMorNegotiate(h []string) bool {
	return slices.ContainsFunc(h, func(s string) bool {
		// RFC 9110 section 11.1 defines the auth-scheme as case-insensitive,
		// hence a credential or challenge is matched on its scheme token whatever its case.
		scheme, _, _ := strings.Cut(s, " ")
		return strings.EqualFold(scheme, "NTLM") || strings.EqualFold(scheme, "Negotiate")
	})
}
