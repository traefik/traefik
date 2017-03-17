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

	if l.appender.IsOpen() {
		core[RequestCount] = nextRequestCount()
		if req.Host != "" {
			core[HTTPAddr] = req.Host
			core[HTTPHost], core[HTTPPort] = silentSplitHostPort(req.Host)
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
		core[HTTPMethod] = req.Method
		core[HTTPRequestPath] = us
		core[HTTPProtocol] = req.Proto
		core[HTTPRequestLine] = fmt.Sprintf("%s %s %s", req.Method, us, req.Proto)

		core[ClientRemoteAddr] = req.RemoteAddr
		core[ClientHost], core[ClientPort] = silentSplitHostPort(req.RemoteAddr)
		core[ClientUsername] = usernameIfPresent(req.URL)

		crw := &captureResponseWriter{rw: rw}

		next.ServeHTTP(crw, r2)

		logDataTable.DownstreamResponse = crw.Header()
		l.logTheRoundTrip(logDataTable, crw)

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
func (l *LogHandler) logTheRoundTrip(logDataTable *LogData, crw *captureResponseWriter) {

	core := logDataTable.Core
	core[DownstreamStatus] = crw.Status()
	core[DownstreamContentSize] = crw.Size()
	if original, ok := core[OriginContentSize]; ok {
		core[GzipRatio] = float64(original.(int64)) / float64(crw.Size())
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
