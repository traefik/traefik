package accesslog

import (
	"net/http"
	"strings"

	"github.com/urfave/negroni"
)

// SaveFrontend sends the frontend name to the logger.
// These are sometimes used with a corresponding SaveBackend handler, but not always.
// For example, redirected requests don't reach a backend.
type SaveFrontend struct {
	next         http.Handler
	frontendName string
}

// NewSaveFrontend creates a SaveFrontend handler.
func NewSaveFrontend(next http.Handler, frontendName string) http.Handler {
	return &SaveFrontend{next, frontendName}
}

func (sf *SaveFrontend) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	serveSaveFrontend(r, sf.frontendName, func() {
		sf.next.ServeHTTP(rw, r)
	})
}

// SaveNegroniFrontend sends the frontend name to the logger.
type SaveNegroniFrontend struct {
	next         negroni.Handler
	frontendName string
}

// NewSaveNegroniFrontend creates a SaveNegroniFrontend handler.
func NewSaveNegroniFrontend(next negroni.Handler, frontendName string) negroni.Handler {
	return &SaveNegroniFrontend{next, frontendName}
}

func (sf *SaveNegroniFrontend) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	serveSaveFrontend(r, sf.frontendName, func() {
		sf.next.ServeHTTP(rw, r, next)
	})
}

func serveSaveFrontend(r *http.Request, frontendName string, apply func()) {
	table := GetLogDataTable(r)
	table.Core[FrontendName] = strings.TrimPrefix(frontendName, "frontend-")

	apply()
}
