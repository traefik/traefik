package rest

import (
	"log"
	"net/http"
)

// DoerFunc satisfies Interface. DoerFuncs are useful for adding
// logging/instrumentation to the http.Client that is used
// within the rest.APIClient.
type DoerFunc func(*http.Request) (*http.Response, error)

// Do is implementation of rest.Doer interface. Calls itself on the
// given http.Request.
func (f DoerFunc) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}

// A Decorator wraps a Doer with extra behavior, and doesnt
// affect the behavior of other instances of the same type.
type Decorator func(Doer) Doer

// Decorate decorates a Doer c with all the given Decorators, in order.
// Core object(Doer instance) that we want to apply layers(Decorator slice) to.
func Decorate(d Doer, ds ...Decorator) Doer {
	decorated := d
	for _, decorate := range ds {
		decorated = decorate(decorated)
	}
	return decorated
}

// Logging returns a Decorator that logs a Doer's requests.
// Dependency injection for the logger instance(inside the closures environment).
func Logging(l *log.Logger) Decorator {
	return func(d Doer) Doer {
		return DoerFunc(func(r *http.Request) (*http.Response, error) {
			l.Printf("%s: %s %s", r.UserAgent(), r.Method, r.URL)
			return d.Do(r)
		})
	}
}
