package accesslog

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/containous/alice"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

type key string

const (
	// DataTableKey is the key within the request context used to store the Log Data Table.
	DataTableKey key = "LogDataTable"

	// CommonFormat is the common logging format (CLF).
	CommonFormat string = "common"

	// JSONFormat is the JSON logging format.
	JSONFormat string = "json"
)

type noopCloser struct {
	*os.File
}

func (n noopCloser) Write(p []byte) (int, error) {
	return n.File.Write(p)
}

func (n noopCloser) Close() error {
	// noop
	return nil
}

type handlerParams struct {
	logDataTable *LogData
}

// Handler will write each request and its response to the access log.
type Handler struct {
	config         *types.AccessLog
	logger         *logrus.Logger
	file           io.WriteCloser
	mu             sync.Mutex
	httpCodeRanges types.HTTPCodeRanges
	logHandlerChan chan handlerParams
	wg             sync.WaitGroup
}

// WrapHandler Wraps access log handler into an Alice Constructor.
func WrapHandler(handler *Handler) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			handler.ServeHTTP(rw, req, next)
		}), nil
	}
}

// NewHandler creates a new Handler.
func NewHandler(config *types.AccessLog) (*Handler, error) {
	var file io.WriteCloser = noopCloser{os.Stdout}
	if len(config.FilePath) > 0 {
		f, err := openAccessLogFile(config.FilePath)
		if err != nil {
			return nil, fmt.Errorf("error opening access log file: %w", err)
		}
		file = f
	}
	logHandlerChan := make(chan handlerParams, config.BufferingSize)

	var formatter logrus.Formatter

	switch config.Format {
	case CommonFormat:
		formatter = new(CommonLogFormatter)
	case JSONFormat:
		formatter = new(logrus.JSONFormatter)
	default:
		log.Error().Msgf("Unsupported access log format: %q, defaulting to common format instead.", config.Format)
		formatter = new(CommonLogFormatter)
	}

	logger := &logrus.Logger{
		Out:       file,
		Formatter: formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}

	// Transform headers names in config to a canonical form, to be used as is without further transformations.
	if config.Fields != nil && config.Fields.Headers != nil && len(config.Fields.Headers.Names) > 0 {
		fields := map[string]string{}

		for h, v := range config.Fields.Headers.Names {
			fields[textproto.CanonicalMIMEHeaderKey(h)] = v
		}

		config.Fields.Headers.Names = fields
	}

	logHandler := &Handler{
		config:         config,
		logger:         logger,
		file:           file,
		logHandlerChan: logHandlerChan,
	}

	if config.Filters != nil {
		if httpCodeRanges, err := types.NewHTTPCodeRanges(config.Filters.StatusCodes); err != nil {
			log.Error().Err(err).Msg("Failed to create new HTTP code ranges")
		} else {
			logHandler.httpCodeRanges = httpCodeRanges
		}
	}

	if config.BufferingSize > 0 {
		logHandler.wg.Add(1)
		go func() {
			defer logHandler.wg.Done()
			for handlerParams := range logHandler.logHandlerChan {
				logHandler.logTheRoundTrip(handlerParams.logDataTable)
			}
		}()
	}

	return logHandler, nil
}

func openAccessLogFile(filePath string) (*os.File, error) {
	dir := filepath.Dir(filePath)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create log path %s: %w", dir, err)
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o664)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", filePath, err)
	}

	return file, nil
}

// GetLogData gets the request context object that contains logging data.
// This creates data as the request passes through the middleware chain.
func GetLogData(req *http.Request) *LogData {
	if ld, ok := req.Context().Value(DataTableKey).(*LogData); ok {
		return ld
	}
	return nil
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	now := time.Now().UTC()

	core := CoreLogData{
		StartUTC:   now,
		StartLocal: now.Local(),
	}

	logDataTable := &LogData{
		Core: core,
		Request: request{
			headers: req.Header,
		},
	}

	reqWithDataTable := req.WithContext(context.WithValue(req.Context(), DataTableKey, logDataTable))

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

	core[RequestScheme] = "http"
	if req.TLS != nil {
		core[RequestScheme] = "https"
		core[TLSVersion] = traefiktls.GetVersion(req.TLS)
		core[TLSCipher] = traefiktls.GetCipherName(req.TLS)
		if len(req.TLS.PeerCertificates) > 0 && req.TLS.PeerCertificates[0] != nil {
			core[TLSClientSubject] = req.TLS.PeerCertificates[0].Subject.String()
		}
	}

	core[ClientAddr] = req.RemoteAddr
	core[ClientHost], core[ClientPort] = silentSplitHostPort(req.RemoteAddr)

	if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		core[ClientHost] = forwardedFor
	}

	ctx := req.Context()
	capt, err := capture.FromContext(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str(logs.MiddlewareType, "AccessLogs").Msg("Could not get Capture")
		return
	}

	defer func(capt capture.Capture, rw http.ResponseWriter) {
		logDataTable.DownstreamResponse = downstreamResponse{
			headers: rw.Header().Clone(),
		}

		logDataTable.DownstreamResponse.status = capt.StatusCode()
		logDataTable.DownstreamResponse.size = capt.ResponseSize()
		logDataTable.Request.size = capt.RequestSize()

		if _, ok := core[ClientUsername]; !ok {
			core[ClientUsername] = usernameIfPresent(reqWithDataTable.URL)
		}

		if h.config.BufferingSize > 0 {
			h.logHandlerChan <- handlerParams{
				logDataTable: logDataTable,
			}
			return
		}

		h.logTheRoundTrip(logDataTable)
	}(capt, rw)

	next.ServeHTTP(rw, reqWithDataTable)
}

// Close closes the Logger (i.e. the file, drain logHandlerChan, etc).
func (h *Handler) Close() error {
	close(h.logHandlerChan)
	h.wg.Wait()
	return h.file.Close()
}

// Rotate closes and reopens the log file to allow for rotation by an external source.
func (h *Handler) Rotate() error {
	if h.config.FilePath == "" {
		return nil
	}

	if h.file != nil {
		defer func(f io.Closer) { _ = f.Close() }(h.file)
	}

	var err error
	h.file, err = os.OpenFile(h.config.FilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o664)
	if err != nil {
		return err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logger.Out = h.file
	return nil
}

func silentSplitHostPort(value string) (host, port string) {
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return value, "-"
	}
	return host, port
}

func usernameIfPresent(theURL *url.URL) string {
	if theURL.User != nil {
		if name := theURL.User.Username(); name != "" {
			return name
		}
	}
	return "-"
}

// Logging handler to log frontend name, backend name, and elapsed time.
func (h *Handler) logTheRoundTrip(logDataTable *LogData) {
	core := logDataTable.Core

	retryAttempts, ok := core[RetryAttempts].(int)
	if !ok {
		retryAttempts = 0
	}
	core[RetryAttempts] = retryAttempts
	core[RequestContentSize] = logDataTable.Request.size

	status := logDataTable.DownstreamResponse.status
	core[DownstreamStatus] = status

	// n.b. take care to perform time arithmetic using UTC to avoid errors at DST boundaries.
	totalDuration := time.Now().UTC().Sub(core[StartUTC].(time.Time))
	core[Duration] = totalDuration

	if h.keepAccessLog(status, retryAttempts, totalDuration) {
		size := logDataTable.DownstreamResponse.size
		core[DownstreamContentSize] = size
		if original, ok := core[OriginContentSize]; ok {
			o64 := original.(int64)
			if size != o64 && size != 0 {
				core[GzipRatio] = float64(o64) / float64(size)
			}
		}

		core[Overhead] = totalDuration
		if origin, ok := core[OriginDuration]; ok {
			core[Overhead] = totalDuration - origin.(time.Duration)
		}

		fields := logrus.Fields{}

		for k, v := range logDataTable.Core {
			if h.config.Fields.Keep(k) {
				fields[k] = v
			}
		}

		h.redactHeaders(logDataTable.Request.headers, fields, "request_")
		h.redactHeaders(logDataTable.OriginResponse, fields, "origin_")
		h.redactHeaders(logDataTable.DownstreamResponse.headers, fields, "downstream_")

		h.mu.Lock()
		defer h.mu.Unlock()
		h.logger.WithFields(fields).Println()
	}
}

func (h *Handler) redactHeaders(headers http.Header, fields logrus.Fields, prefix string) {
	for k := range headers {
		v := h.config.Fields.KeepHeader(k)
		if v == types.AccessLogKeep {
			fields[prefix+k] = strings.Join(headers.Values(k), ",")
		} else if v == types.AccessLogRedact {
			fields[prefix+k] = "REDACTED"
		}
	}
}

func (h *Handler) keepAccessLog(statusCode, retryAttempts int, duration time.Duration) bool {
	if h.config.Filters == nil {
		// no filters were specified
		return true
	}

	if len(h.httpCodeRanges) == 0 && !h.config.Filters.RetryAttempts && h.config.Filters.MinDuration == 0 {
		// empty filters were specified, e.g. by passing --accessLog.filters only (without other filter options)
		return true
	}

	if h.httpCodeRanges.Contains(statusCode) {
		return true
	}

	if h.config.Filters.RetryAttempts && retryAttempts > 0 {
		return true
	}

	if h.config.Filters.MinDuration > 0 && (ptypes.Duration(duration) > h.config.Filters.MinDuration) {
		return true
	}

	return false
}

var requestCounter uint64 // Request ID

func nextRequestCount() uint64 {
	return atomic.AddUint64(&requestCounter, 1)
}
