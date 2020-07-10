package accesslog

import (
	"net/http"
	"time"

	"github.com/vulcand/oxy/utils"
)

// FieldApply function hook to add data in accesslog.
type FieldApply func(rw http.ResponseWriter, r *http.Request, next http.Handler, data *LogData)

// FieldHandler sends a new field to the logger.
type FieldHandler struct {
	next    http.Handler
	name    string
	value   string
	applyFn FieldApply
}

// NewFieldHandler creates a Field handler.
func NewFieldHandler(next http.Handler, name, value string, applyFn FieldApply) http.Handler {
	return &FieldHandler{next: next, name: name, value: value, applyFn: applyFn}
}

func (f *FieldHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	table := GetLogData(req)
	if table == nil {
		f.next.ServeHTTP(rw, req)
		return
	}

	table.Core[f.name] = f.value

	if f.applyFn != nil {
		f.applyFn(rw, req, f.next, table)
	} else {
		f.next.ServeHTTP(rw, req)
	}
}

// AddServiceFields add service fields.
func AddServiceFields(rw http.ResponseWriter, req *http.Request, next http.Handler, data *LogData) {
	data.Core[ServiceURL] = req.URL // note that this is *not* the original incoming URL
	data.Core[ServiceAddr] = req.URL.Host

	next.ServeHTTP(rw, req)
}

// AddOriginFields add origin fields.
func AddOriginFields(rw http.ResponseWriter, req *http.Request, next http.Handler, data *LogData) {
	crw := newCaptureResponseWriter(rw)
	start := time.Now().UTC()

	next.ServeHTTP(crw, req)

	// use UTC to handle switchover of daylight saving correctly
	data.Core[OriginDuration] = time.Now().UTC().Sub(start)
	data.Core[OriginStatus] = crw.Status()
	// make copy of headers so we can ensure there is no subsequent mutation during response processing
	data.OriginResponse = make(http.Header)
	utils.CopyHeaders(data.OriginResponse, crw.Header())
	data.Core[OriginContentSize] = crw.Size()
}
