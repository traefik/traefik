package route

import (
	"fmt"
	"net/http"
)

// Mux implements router compatible with http.Handler
type Mux struct {
	// NotFound sets handler for routes that are not found
	notFound http.Handler
	router   Router
}

// NewMux returns new Mux router
func NewMux() *Mux {
	return &Mux{
		router:   New(),
		notFound: &notFound{},
	}
}

// Handle adds http handler for route expression
func (m *Mux) Handle(expr string, handler http.Handler) error {
	return m.router.UpsertRoute(expr, handler)
}

// Handle adds http handler function for route expression
func (m *Mux) HandleFunc(expr string, handler func(http.ResponseWriter, *http.Request)) error {
	return m.Handle(expr, http.HandlerFunc(handler))
}

func (m *Mux) Remove(expr string) error {
	return m.router.RemoveRoute(expr)
}

// ServeHTTP routes the request and passes it to handler
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, err := m.router.Route(r)
	if err != nil || h == nil {
		m.notFound.ServeHTTP(w, r)
		return
	}
	h.(http.Handler).ServeHTTP(w, r)
}

func (m *Mux) SetNotFound(n http.Handler) error {
	if n == nil {
		return fmt.Errorf("Not Found handler cannot be nil. Operation rejected.")
	}
	m.notFound = n
	return nil
}

func (m *Mux) GetNotFound() http.Handler {
	return m.notFound
}

func (m *Mux) IsValid(expr string) bool {
	return IsValid(expr)
}

// NotFound is a generic http.Handler for request
type notFound struct {
}

// ServeHTTP returns a simple 404 Not found response
func (notFound) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "Not found")

}

