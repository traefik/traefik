package common

import "net/http"

// Middleware types should embed BasicMiddleware and implement ServeHTTP and Close as required.
type Middleware interface {
	http.Handler
	Next() http.Handler
	Close()
}

// BasicMiddleware should be embedded in Middleware types. These implement ServeHTTP, and must
// be careful to call the next middleware (unless explicitly require not to).
//
// Close should also be overridden if required.
type BasicMiddleware struct {
	next http.Handler
}

// NewMiddleware creates a new BasicMiddleware.
func NewMiddleware(next http.Handler) BasicMiddleware {
	return BasicMiddleware{next}
}

// Next returbs the next middleware in the chain.
func (m *BasicMiddleware) Next() http.Handler {
	return m.next
}

// Close closes the next middleware in the chain but does not do anything to this one.
// This method should be overridden if specialised behaviour is needed.
// Note that any errors should be logged; they are not propagated back.
func (m *BasicMiddleware) Close() {
	if m.next != nil {
		if c, ok := m.next.(Middleware); ok {
			c.Close()
		}
	}
	// no further action in this case
}
