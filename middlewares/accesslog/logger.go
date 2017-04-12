package accesslog

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

type key int

const (
	dataTableKey key = iota
)

// LogHandler writes each request and its response to the access log.
// It gets some information from the logInfoResponseWriter set up by previous middleware.
type LogHandler struct {
	appender LogAppender
}

// NewLogHandler creates a new LogHandler using the appender provided.
func NewLogHandler(appender LogAppender) *LogHandler {
	return &LogHandler{appender}
}

// GetLogDataTable gets the request context object that contains logging data. This accretes
// data as the request passes through the middleware chain.
func GetLogDataTable(req *http.Request) *LogData {
	return req.Context().Value(dataTableKey).(*LogData)
}

func (l *LogHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	now := time.Now().UTC()
	core := make(CoreLogData)

	logDataTable := &LogData{core, req.Header, nil, nil}
	core[StartUTC] = now
	core[StartLocal] = now.Local()

	r2 := req.WithContext(context.WithValue(req.Context(), dataTableKey, logDataTable))

	var crr *captureRequestReader
	if req.Body != nil {
		crr = &captureRequestReader{req.Body, 0}
		r2.Body = crr
	}

	if l.appender.IsOpen() {
		core[RequestCount] = nextRequestCount()
		if req.Host != "" {
			core[RequestAddr] = req.Host
			core[RequestHost], core[RequestPort] = silentSplitHostPort(req.Host)
		}
		// copy the URL without the scheme, hostname etc
		u := &url.URL{
			Path:       req.URL.Path,
			RawPath:    req.URL.RawPath,
			RawQuery:   req.URL.RawQuery,
			ForceQuery: req.URL.ForceQuery,
			Fragment:   req.URL.Fragment,
		}
		us := u.String()
		core[RequestMethod] = req.Method
		core[RequestPath] = us
		core[RequestProtocol] = req.Proto
		core[RequestLine] = fmt.Sprintf("%s %s %s", req.Method, us, req.Proto)

		core[ClientAddr] = req.RemoteAddr
		core[ClientHost], core[ClientPort] = silentSplitHostPort(req.RemoteAddr)
		core[ClientUsername] = usernameIfPresent(req.URL)

		crw := &captureResponseWriter{rw: rw}

		next.ServeHTTP(crw, r2)

		logDataTable.DownstreamResponse = crw.Header()
		l.logTheRoundTrip(logDataTable, crr, crw)

	} else {
		next(rw, r2)
	}
}

// Close closes the Logger (i.e. the file etc).
func (l *LogHandler) Close() error {
	return l.appender.Close()
}

func silentSplitHostPort(value string) (string, string) {
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return value, "-"
	}
	return host, port
}

func usernameIfPresent(u *url.URL) string {
	username := "-"
	if u.User != nil {
		if name := u.User.Username(); name != "" {
			username = name
		}
	}
	return username
}

// Logging handler to log frontend name, backend name, and elapsed time
func (l *LogHandler) logTheRoundTrip(logDataTable *LogData, crr *captureRequestReader, crw *captureResponseWriter) {

	core := logDataTable.Core

	if crr != nil {
		core[RequestContentSize] = crr.count
	}

	core[DownstreamStatus] = crw.Status()
	core[DownstreamStatusLine] = fmt.Sprintf("%03d %s", crw.Status(), http.StatusText(crw.Status()))
	core[DownstreamContentSize] = crw.Size()
	if original, ok := core[OriginContentSize]; ok {
		o64 := original.(int64)
		if o64 != crw.Size() {
			// n.b divide-by-zero is tolerated here and dealt with elsewhere as appropriate
			core[GzipRatio] = float64(o64) / float64(crw.Size())
		}
	}

	// n.b. take care to perform time arithmetic using UTC to avoid errors at DST boundaries
	total := time.Now().UTC().Sub(core[StartUTC].(time.Time))
	core[Duration] = total
	if origin, ok := core[OriginDuration]; ok {
		core[Overhead] = total - origin.(time.Duration)
	} else {
		core[Overhead] = total
	}

	l.appender.Write(logDataTable)
}

//-------------------------------------------------------------------------------------------------

var requestCounter uint64 // Request ID

func nextRequestCount() uint64 {
	return atomic.AddUint64(&requestCounter, 1)
}
