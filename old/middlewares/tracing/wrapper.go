package tracing

import (
	"net/http"

	"github.com/urfave/negroni"
)

// NewNegroniHandlerWrapper return a negroni.Handler struct
func (t *Tracing) NewNegroniHandlerWrapper(name string, handler negroni.Handler, clientSpanKind bool) negroni.Handler {
	if t.IsEnabled() && handler != nil {
		return &NegroniHandlerWrapper{
			name:           name,
			next:           handler,
			clientSpanKind: clientSpanKind,
		}
	}
	return handler
}

// NewHTTPHandlerWrapper return a http.Handler struct
func (t *Tracing) NewHTTPHandlerWrapper(name string, handler http.Handler, clientSpanKind bool) http.Handler {
	if t.IsEnabled() && handler != nil {
		return &HTTPHandlerWrapper{
			name:           name,
			handler:        handler,
			clientSpanKind: clientSpanKind,
		}
	}
	return handler
}

// NegroniHandlerWrapper is used to wrap negroni handler middleware
type NegroniHandlerWrapper struct {
	name           string
	next           negroni.Handler
	clientSpanKind bool
}

func (t *NegroniHandlerWrapper) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	var finish func()
	_, r, finish = StartSpan(r, t.name, t.clientSpanKind)
	defer finish()

	if t.next != nil {
		t.next.ServeHTTP(rw, r, next)
	}
}

// HTTPHandlerWrapper is used to wrap http handler middleware
type HTTPHandlerWrapper struct {
	name           string
	handler        http.Handler
	clientSpanKind bool
}

func (t *HTTPHandlerWrapper) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var finish func()
	_, r, finish = StartSpan(r, t.name, t.clientSpanKind)
	defer finish()

	if t.handler != nil {
		t.handler.ServeHTTP(rw, r)
	}

}
