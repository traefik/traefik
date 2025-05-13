package accesslog

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/vulcand/oxy/v2/utils"
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
	start := time.Now().UTC()

	next.ServeHTTP(rw, req)

	// use UTC to handle switchover of daylight saving correctly
	data.Core[OriginDuration] = time.Now().UTC().Sub(start)
	// make copy of headers, so we can ensure there is no subsequent mutation
	// during response processing
	data.OriginResponse = make(http.Header)
	utils.CopyHeaders(data.OriginResponse, rw.Header())

	ctx := req.Context()
	capt, err := capture.FromContext(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str(logs.MiddlewareType, "AccessLogs").Msg("Could not get Capture")
		return
	}

	data.Core[OriginStatus] = capt.StatusCode()
	data.Core[OriginContentSize] = capt.ResponseSize()
}

// InitServiceFields init service fields.
func InitServiceFields(rw http.ResponseWriter, req *http.Request, next http.Handler, data *LogData) {
	// Because they are expected to be initialized when the logger is processing the data table,
	// the origin fields are initialized in case the response is returned by Traefik itself, and not a service.
	data.Core[OriginDuration] = time.Duration(0)
	data.Core[OriginStatus] = 0
	data.Core[OriginContentSize] = int64(0)

	next.ServeHTTP(rw, req)
}
