package accesslog

import (
	"context"
	"net/http"

	"github.com/urfave/negroni"
)

const (
	clientUsernameKey key = "ClientUsername"
)

// SaveUsername sends the Username name to the access logger.
type SaveUsername struct {
	next http.Handler
}

// NewSaveUsername creates a SaveUsername handler.
func NewSaveUsername(next http.Handler) http.Handler {
	return &SaveUsername{next}
}

func (sf *SaveUsername) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	serveSaveUsername(r, func() {
		sf.next.ServeHTTP(rw, r)
	})
}

// SaveNegroniUsername adds the Username to the access logger data table.
type SaveNegroniUsername struct {
	next negroni.Handler
}

// NewSaveNegroniUsername creates a SaveNegroniUsername handler.
func NewSaveNegroniUsername(next negroni.Handler) negroni.Handler {
	return &SaveNegroniUsername{next}
}

func (sf *SaveNegroniUsername) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	serveSaveUsername(r, func() {
		sf.next.ServeHTTP(rw, r, next)
	})
}

func serveSaveUsername(r *http.Request, apply func()) {
	table := GetLogDataTable(r)

	username, ok := r.Context().Value(clientUsernameKey).(string)
	if ok {
		table.Core[ClientUsername] = username
	}

	apply()
}

// WithUserName adds a username to a requests' context
func WithUserName(req *http.Request, username string) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), clientUsernameKey, username))
}
