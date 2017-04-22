package accesslog

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/vulcand/oxy/utils"
)

// SaveBackend sends the backend name to the logger. These are always used with a corresponding
// SaveFrontend handler.
type SaveBackend struct {
	next        http.Handler
	backendName string
}

// NewSaveBackend creates a SaveBackend handler.
func NewSaveBackend(next http.Handler, backendName string) http.Handler {
	return &SaveBackend{next, backendName}
}

func (sb *SaveBackend) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	table := GetLogDataTable(r)
	table.Core[BackendName] = sb.backendName
	table.Core[BackendURL] = r.URL // note that this is *not* the original incoming URL
	table.Core[BackendAddr] = r.URL.Host

	crw := &captureResponseWriter{rw: rw}
	start := time.Now().UTC()

	sb.next.ServeHTTP(crw, r)

	// use UTC to handle switchover of daylight saving correctly
	table.Core[OriginDuration] = time.Now().UTC().Sub(start)
	table.Core[OriginStatus] = crw.Status()
	table.Core[OriginStatusLine] = fmt.Sprintf("%03d %s", crw.Status(), http.StatusText(crw.Status()))
	// make copy of headers so we can ensure there is no subsequent mutation during response processing
	table.OriginResponse = make(http.Header)
	utils.CopyHeaders(table.OriginResponse, crw.Header())
	table.Core[OriginContentSize] = crw.Size()
}

//-------------------------------------------------------------------------------------------------

// SaveFrontend sends the frontend name to the logger. These are sometimes used with a corresponding
// SaveBackend handler, but not always. For example, redirected requests don't reach a backend.
type SaveFrontend struct {
	next         http.Handler
	frontendName string
}

// NewSaveFrontend creates a SaveFrontend handler.
func NewSaveFrontend(next http.Handler, frontendName string) http.Handler {
	return &SaveFrontend{next, frontendName}
}

func (sb *SaveFrontend) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	table := GetLogDataTable(r)
	table.Core[FrontendName] = strings.TrimPrefix(sb.frontendName, "frontend-")

	sb.next.ServeHTTP(rw, r)
}
