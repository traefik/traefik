package accesslog

import (
	"fmt"
	"net/http"
	"time"

	"github.com/urfave/negroni"
	"github.com/vulcand/oxy/utils"
)

// SaveBackend sends the backend name to the logger.
// These are always used with a corresponding SaveFrontend handler.
type SaveBackend struct {
	next        http.Handler
	backendName string
}

// NewSaveBackend creates a SaveBackend handler.
func NewSaveBackend(next http.Handler, backendName string) http.Handler {
	return &SaveBackend{next, backendName}
}

func (sb *SaveBackend) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	serveSaveBackend(rw, r, sb.backendName, func(crw *captureResponseWriter) {
		sb.next.ServeHTTP(crw, r)
	})
}

// SaveNegroniBackend sends the backend name to the logger.
type SaveNegroniBackend struct {
	next        negroni.Handler
	backendName string
}

// NewSaveNegroniBackend creates a SaveBackend handler.
func NewSaveNegroniBackend(next negroni.Handler, backendName string) negroni.Handler {
	return &SaveNegroniBackend{next, backendName}
}

func (sb *SaveNegroniBackend) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	serveSaveBackend(rw, r, sb.backendName, func(crw *captureResponseWriter) {
		sb.next.ServeHTTP(crw, r, next)
	})
}

func serveSaveBackend(rw http.ResponseWriter, r *http.Request, backendName string, apply func(*captureResponseWriter)) {
	table := GetLogDataTable(r)
	table.Core[BackendName] = backendName
	table.Core[BackendURL] = r.URL // note that this is *not* the original incoming URL
	table.Core[BackendAddr] = r.URL.Host

	crw := &captureResponseWriter{rw: rw}
	start := time.Now().UTC()

	apply(crw)

	// use UTC to handle switchover of daylight saving correctly
	table.Core[OriginDuration] = time.Now().UTC().Sub(start)
	table.Core[OriginStatus] = crw.Status()
	table.Core[OriginStatusLine] = fmt.Sprintf("%03d %s", crw.Status(), http.StatusText(crw.Status()))
	// make copy of headers so we can ensure there is no subsequent mutation during response processing
	table.OriginResponse = make(http.Header)
	utils.CopyHeaders(table.OriginResponse, crw.Header())
	table.Core[OriginContentSize] = crw.Size()
}
