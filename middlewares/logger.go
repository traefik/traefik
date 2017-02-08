package middlewares

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/containous/traefik/log"
	"github.com/streamrail/concurrent-map"
)

const (
	loggerReqidHeader = "X-Traefik-Reqid"
)

/*
Logger writes each request and its response to the access log.
It gets some information from the logInfoResponseWriter set up by previous middleware.
*/
type Logger struct {
	file   *os.File
	format string
}

// Logging handler to log frontend name, backend name, and elapsed time
type frontendBackendLoggingHandler struct {
	reqid       string
	writer      io.Writer
	format      string
	handlerFunc http.HandlerFunc
}

var (
	reqidCounter        uint64       // Request ID
	infoRwMap           = cmap.New() // Map of reqid to response writer
	backend2FrontendMap *map[string]string
)

// logInfoResponseWriter is a wrapper of type http.ResponseWriter
// that tracks frontend and backend names and request status and size
type logInfoResponseWriter struct {
	rw       http.ResponseWriter
	backend  string
	frontend string
	status   int
	size     int
}

// logEntry is a single log entry for use in encoding to json
type logEntry struct {
	RemoteAddr    string `json:"remoteAddr"`
	Username      string `json:"username"`
	Timestamp     string `json:"timestamp"`
	Method        string `json:"method"`
	URI           string `json:"uri"`
	Protocol      string `json:"protocol"`
	Status        int    `json:"status"`
	Size          int    `json:"size"`
	Referer       string `json:"referer"`
	UserAgent     string `json:"userAgent"`
	RequestID     string `json:"requestID"`
	Frontend      string `json:"frontend"`
	Backend       string `json:"backend"`
	ElapsedMillis int64  `json:"elapsedMillis"`
	Host          string `json:"host"`
}

// logEntryPool is used as we allocate a new logEntry on every request
var logEntryPool = sync.Pool{
	New: func() interface{} {
		return &logEntry{}
	},
}

func (fblh *frontendBackendLoggingHandler) writeJSON(e *logEntry) {
	data, err := json.Marshal(e)
	if err != nil {
		log.Error("unable to marshal json for log entry", err)
		return
	}
	data = append(data, newLineByte)
	// must do single write, rather than two (data then newline) to avoid interleaving lines
	fblh.writer.Write(data)
}

func (fblh *frontendBackendLoggingHandler) writeText(e *logEntry) {
	fmt.Fprintf(fblh.writer, `%s - %s [%s] "%s %s %s" %d %d "%s" "%s" %s "%s" "%s" %dms%s`,
		e.Host, e.Username, e.Timestamp, e.Method, e.URI, e.Protocol, e.Status, e.Size, e.Referer, e.UserAgent, e.RequestID, e.Frontend, e.Backend, e.ElapsedMillis, "\n")
}

// newLineByte is simple "\n" as a byte
var newLineByte = []byte("\n")[0]

// NewLogger returns a new Logger instance.
func NewLogger(file, format string) *Logger {
	if len(file) > 0 {
		fi, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Error("Error opening file", err)
		}
		return &Logger{file: fi, format: format}
	}
	return &Logger{file: nil, format: format}
}

// SetBackend2FrontendMap is called by server.go to set up frontend translation
func SetBackend2FrontendMap(newMap *map[string]string) {
	backend2FrontendMap = newMap
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if l.file == nil {
		next(rw, r)
	} else {
		reqid := strconv.FormatUint(atomic.AddUint64(&reqidCounter, 1), 10)
		r.Header[loggerReqidHeader] = []string{reqid}
		defer deleteReqid(r, reqid)

		(&frontendBackendLoggingHandler{reqid, l.file, l.format, next}).ServeHTTP(rw, r)
	}
}

// Delete a reqid from the map and the request's headers
func deleteReqid(r *http.Request, reqid string) {
	infoRwMap.Remove(reqid)
	delete(r.Header, loggerReqidHeader)
}

// Save the backend name for the Logger
func saveBackendNameForLogger(r *http.Request, backendName string) {
	if reqidHdr := r.Header[loggerReqidHeader]; len(reqidHdr) == 1 {
		reqid := reqidHdr[0]
		if infoRw, ok := infoRwMap.Get(reqid); ok {
			infoRw.(*logInfoResponseWriter).SetBackend(backendName)
			infoRw.(*logInfoResponseWriter).SetFrontend((*backend2FrontendMap)[backendName])
		}
	}
}

// Close closes the Logger (i.e. the file).
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

// Logging handler to log frontend name, backend name, and elapsed time
func (fblh *frontendBackendLoggingHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	startTime := time.Now()
	infoRw := &logInfoResponseWriter{rw: rw}
	infoRwMap.Set(fblh.reqid, infoRw)
	fblh.handlerFunc(infoRw, req)

	username := "-"
	url := *req.URL
	if url.User != nil {
		if name := url.User.Username(); name != "" {
			username = name
		}
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}

	uri := url.RequestURI()
	if qmIndex := strings.Index(uri, "?"); qmIndex > 0 {
		uri = uri[0:qmIndex]
	}

	e := logEntryPool.Get().(*logEntry)
	defer logEntryPool.Put(e)

	e.RemoteAddr = host
	e.Username = username
	e.Timestamp = startTime.Format("02/Jan/2006:15:04:05 -0700")
	e.Method = req.Method
	e.URI = uri
	e.Protocol = req.Proto
	e.Status = infoRw.GetStatus()
	e.Size = infoRw.GetSize()
	e.Referer = req.Referer()
	e.UserAgent = req.UserAgent()
	e.RequestID = fblh.reqid
	e.Frontend = strings.TrimPrefix(infoRw.GetFrontend(), "frontend-")
	e.Backend = infoRw.GetBackend()
	e.ElapsedMillis = time.Since(startTime).Nanoseconds() / 1000000
	e.Host = req.Host

	if fblh.format == "json" {
		fblh.writeJSON(e)
	} else {
		fblh.writeText(e)
	}
}

func (lirw *logInfoResponseWriter) Header() http.Header {
	return lirw.rw.Header()
}

func (lirw *logInfoResponseWriter) Write(b []byte) (int, error) {
	if lirw.status == 0 {
		lirw.status = http.StatusOK
	}
	size, err := lirw.rw.Write(b)
	lirw.size += size
	return size, err
}

func (lirw *logInfoResponseWriter) WriteHeader(s int) {
	lirw.rw.WriteHeader(s)
	lirw.status = s
}

func (lirw *logInfoResponseWriter) Flush() {
	f, ok := lirw.rw.(http.Flusher)
	if ok {
		f.Flush()
	}
}

func (lirw *logInfoResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return lirw.rw.(http.Hijacker).Hijack()
}

func (lirw *logInfoResponseWriter) GetStatus() int {
	return lirw.status
}

func (lirw *logInfoResponseWriter) GetSize() int {
	return lirw.size
}

func (lirw *logInfoResponseWriter) GetBackend() string {
	return lirw.backend
}

func (lirw *logInfoResponseWriter) GetFrontend() string {
	return lirw.frontend
}

func (lirw *logInfoResponseWriter) SetBackend(backend string) {
	lirw.backend = backend
}

func (lirw *logInfoResponseWriter) SetFrontend(frontend string) {
	lirw.frontend = frontend
}
