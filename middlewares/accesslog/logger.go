package accesslog

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/containous/flaeg/parse"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
	"github.com/sirupsen/logrus"
)

type key string

const (
	// DataTableKey is the key within the request context used to
	// store the Log Data Table
	DataTableKey key = "LogDataTable"

	// CommonFormat is the common logging format (CLF)
	CommonFormat = "common"

	// JSONFormat is the JSON logging format
	JSONFormat = "json"
)

type logHandlerParams struct {
	logDataTable *LogData
	crr          *captureRequestReader
	crw          *captureResponseWriter
}

// LogHandler will write each request and its response to the access log.
type LogHandler struct {
	config         *types.AccessLog
	logger         *logrus.Logger
	file           *os.File
	mu             sync.Mutex
	httpCodeRanges types.HTTPCodeRanges
	logHandlerChan chan logHandlerParams
	wg             sync.WaitGroup
}

// NewLogHandler creates a new LogHandler
func NewLogHandler(config *types.AccessLog) (*LogHandler, error) {
	file := os.Stdout
	if len(config.FilePath) > 0 {
		f, err := openAccessLogFile(config.FilePath)
		if err != nil {
			return nil, fmt.Errorf("error opening access log file: %s", err)
		}
		file = f
	}
	logHandlerChan := make(chan logHandlerParams, config.BufferingSize)

	var formatter logrus.Formatter

	switch config.Format {
	case CommonFormat:
		formatter = new(CommonLogFormatter)
	case JSONFormat:
		formatter = new(logrus.JSONFormatter)
	default:
		return nil, fmt.Errorf("unsupported access log format: %s", config.Format)
	}

	logger := &logrus.Logger{
		Out:       file,
		Formatter: formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}

	logHandler := &LogHandler{
		config:         config,
		logger:         logger,
		file:           file,
		logHandlerChan: logHandlerChan,
	}

	if config.Filters != nil {
		if httpCodeRanges, err := types.NewHTTPCodeRanges(config.Filters.StatusCodes); err != nil {
			log.Errorf("Failed to create new HTTP code ranges: %s", err)
		} else {
			logHandler.httpCodeRanges = httpCodeRanges
		}
	}

	if config.BufferingSize > 0 {
		logHandler.wg.Add(1)
		go func() {
			defer logHandler.wg.Done()
			for handlerParams := range logHandler.logHandlerChan {
				logHandler.logTheRoundTrip(handlerParams.logDataTable, handlerParams.crr, handlerParams.crw)
			}
		}()
	}

	return logHandler, nil
}

func openAccessLogFile(filePath string) (*os.File, error) {
	dir := filepath.Dir(filePath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log path %s: %s", dir, err)
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %s", filePath, err)
	}

	return file, nil
}

// GetLogDataTable gets the request context object that contains logging data.
// This creates data as the request passes through the middleware chain.
func GetLogDataTable(req *http.Request) *LogData {
	if ld, ok := req.Context().Value(DataTableKey).(*LogData); ok {
		return ld
	}
	log.Errorf("%s is nil", DataTableKey)
	return &LogData{Core: make(CoreLogData)}
}

func (l *LogHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	now := time.Now().UTC()

	core := CoreLogData{
		StartUTC:   now,
		StartLocal: now.Local(),
	}

	logDataTable := &LogData{Core: core, Request: req.Header}

	reqWithDataTable := req.WithContext(context.WithValue(req.Context(), DataTableKey, logDataTable))

	var crr *captureRequestReader
	if req.Body != nil {
		crr = &captureRequestReader{source: req.Body, count: 0}
		reqWithDataTable.Body = crr
	}

	core[RequestCount] = nextRequestCount()
	if req.Host != "" {
		core[RequestAddr] = req.Host
		core[RequestHost], core[RequestPort] = silentSplitHostPort(req.Host)
	}
	// copy the URL without the scheme, hostname etc
	urlCopy := &url.URL{
		Path:       req.URL.Path,
		RawPath:    req.URL.RawPath,
		RawQuery:   req.URL.RawQuery,
		ForceQuery: req.URL.ForceQuery,
		Fragment:   req.URL.Fragment,
	}
	urlCopyString := urlCopy.String()
	core[RequestMethod] = req.Method
	core[RequestPath] = urlCopyString
	core[RequestProtocol] = req.Proto
	core[RequestLine] = fmt.Sprintf("%s %s %s", req.Method, urlCopyString, req.Proto)

	core[ClientAddr] = req.RemoteAddr
	core[ClientHost], core[ClientPort] = silentSplitHostPort(req.RemoteAddr)

	if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		core[ClientHost] = forwardedFor
	}

	crw := &captureResponseWriter{rw: rw}

	next.ServeHTTP(crw, reqWithDataTable)

	core[ClientUsername] = formatUsernameForLog(core[ClientUsername])

	logDataTable.DownstreamResponse = crw.Header()

	if l.config.BufferingSize > 0 {
		l.logHandlerChan <- logHandlerParams{
			logDataTable: logDataTable,
			crr:          crr,
			crw:          crw,
		}
	} else {
		l.logTheRoundTrip(logDataTable, crr, crw)
	}
}

// Close closes the Logger (i.e. the file, drain logHandlerChan, etc).
func (l *LogHandler) Close() error {
	close(l.logHandlerChan)
	l.wg.Wait()
	return l.file.Close()
}

// Rotate closes and reopens the log file to allow for rotation
// by an external source.
func (l *LogHandler) Rotate() error {
	var err error

	if l.file != nil {
		defer func(f *os.File) {
			f.Close()
		}(l.file)
	}

	l.file, err = os.OpenFile(l.config.FilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Out = l.file
	return nil
}

func silentSplitHostPort(value string) (host string, port string) {
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return value, "-"
	}
	return host, port
}

func formatUsernameForLog(usernameField interface{}) string {
	username, ok := usernameField.(string)
	if ok && len(username) != 0 {
		return username
	}
	return "-"
}

// Logging handler to log frontend name, backend name, and elapsed time
func (l *LogHandler) logTheRoundTrip(logDataTable *LogData, crr *captureRequestReader, crw *captureResponseWriter) {
	core := logDataTable.Core

	retryAttempts, ok := core[RetryAttempts].(int)
	if !ok {
		retryAttempts = 0
	}
	core[RetryAttempts] = retryAttempts

	if crr != nil {
		core[RequestContentSize] = crr.count
	}

	core[DownstreamStatus] = crw.Status()

	// n.b. take care to perform time arithmetic using UTC to avoid errors at DST boundaries
	totalDuration := time.Now().UTC().Sub(core[StartUTC].(time.Time))
	core[Duration] = totalDuration

	if l.keepAccessLog(crw.Status(), retryAttempts, totalDuration) {
		core[DownstreamStatusLine] = fmt.Sprintf("%03d %s", crw.Status(), http.StatusText(crw.Status()))
		core[DownstreamContentSize] = crw.Size()
		if original, ok := core[OriginContentSize]; ok {
			o64 := original.(int64)
			if o64 != crw.Size() && 0 != crw.Size() {
				core[GzipRatio] = float64(o64) / float64(crw.Size())
			}
		}

		core[Overhead] = totalDuration
		if origin, ok := core[OriginDuration]; ok {
			core[Overhead] = totalDuration - origin.(time.Duration)
		}

		fields := logrus.Fields{}

		for k, v := range logDataTable.Core {
			if l.config.Fields.Keep(k) {
				fields[k] = v
			}
		}

		l.redactHeaders(logDataTable.Request, fields, "request_")
		l.redactHeaders(logDataTable.OriginResponse, fields, "origin_")
		l.redactHeaders(logDataTable.DownstreamResponse, fields, "downstream_")

		l.mu.Lock()
		defer l.mu.Unlock()
		l.logger.WithFields(fields).Println()
	}
}

func (l *LogHandler) redactHeaders(headers http.Header, fields logrus.Fields, prefix string) {
	for k := range headers {
		v := l.config.Fields.KeepHeader(k)
		if v == types.AccessLogKeep {
			fields[prefix+k] = headers.Get(k)
		} else if v == types.AccessLogRedact {
			fields[prefix+k] = "REDACTED"
		}
	}
}

func (l *LogHandler) keepAccessLog(statusCode, retryAttempts int, duration time.Duration) bool {
	if l.config.Filters == nil {
		// no filters were specified
		return true
	}

	if len(l.httpCodeRanges) == 0 && !l.config.Filters.RetryAttempts && l.config.Filters.MinDuration == 0 {
		// empty filters were specified, e.g. by passing --accessLog.filters only (without other filter options)
		return true
	}

	if l.httpCodeRanges.Contains(statusCode) {
		return true
	}

	if l.config.Filters.RetryAttempts && retryAttempts > 0 {
		return true
	}

	if l.config.Filters.MinDuration > 0 && (parse.Duration(duration) > l.config.Filters.MinDuration) {
		return true
	}

	return false
}

var requestCounter uint64 // Request ID

func nextRequestCount() uint64 {
	return atomic.AddUint64(&requestCounter, 1)
}
