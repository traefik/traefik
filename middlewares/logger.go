package middlewares

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	loggerReqidHeader = "X-Traefik-Reqid"
	loggerFrontend    = 0
	loggerBackend     = 1
)

var (
	url2Backend  map[string]string       // Mapping of URLs to backend name
	reqidCounter uint64                  // Request ID
	reqid2Names  = map[string][]string{} // Map of reqid to frontend and backend names
)

// Logger is a middleware handler that logs the request as it goes in and the response as it goes out.
type Logger struct {
	file *os.File
}

// Logging handler to log frontend name, backend name, and elapsed time
type frontendBackendLoggingHandler struct {
	reqid   string
	writer  io.Writer
	handler http.Handler
}

// NewLogger returns a new Logger instance.
func NewLogger(file string) *Logger {
	if len(file) > 0 {
		fi, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Error opening file", err)
		}
		return &Logger{fi}
	}
	return &Logger{nil}
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if l.file == nil {
		next(rw, r)
	} else {
		reqidCounter++
		reqid := strconv.FormatUint(reqidCounter, 10)
		log.Debugf("Starting request %s", reqid)
		reqid2Names[reqid] = []string{"Unknown frontend", "Unknown backend"}
		defer deleteRequest(reqid)
		r.Header[loggerReqidHeader] = []string{reqid}
		frontendBackendLoggingHandler{reqid, l.file, next}.ServeHTTP(rw, r)
	}
}

// Delete a request from the map
func deleteRequest(reqid string) {
	log.Debugf("Ending request %s", reqid)
	delete(reqid2Names, reqid)
}

// Save a frontend or backend name for the logger
func saveNameForLogger(r *http.Request, nameType int, name string) {
	if reqidHdr := r.Header[loggerReqidHeader]; len(reqidHdr) == 1 {
		reqid := reqidHdr[0]
		if len(reqid2Names[reqid]) > 0 {
			reqid2Names[reqid][nameType] = name
		}
	}
}

// Close closes the logger (i.e. the file).
func (l *Logger) Close() {
	l.file.Close()
}

// Logging handler to log frontend name, backend name, and elapsed time
func (h frontendBackendLoggingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now()
	logger := &responseLogger{w: w}
	h.handler.ServeHTTP(logger, req)

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

	ts := startTime.Format("02/Jan/2006:15:04:05 -0700")
	method := req.Method
	uri := url.RequestURI()
	proto := req.Proto
	status := logger.Status()
	len := logger.Size()
	referer := req.Referer()
	agent := req.UserAgent()
	frontend := reqid2Names[h.reqid][loggerFrontend]
	backend := url2Backend[reqid2Names[h.reqid][loggerBackend]]
	elapsed := time.Now().UTC().Sub(startTime.UTC())

	fmt.Fprintf(h.writer, `%s - %s [%s] "%s %s %s" %d %d "%s" "%s" %s "%s" "%s" %s%s`,
		host, username, ts, method, uri, proto, status, len, referer, agent, h.reqid, frontend, backend, elapsed, "\n")

}

// responseLogger is wrapper of http.ResponseWriter that keeps track of its HTTP status
// code and body size
type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

func (l *responseLogger) Write(b []byte) (int, error) {
	if l.status == 0 {
		l.status = http.StatusOK
	}
	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func (l *responseLogger) Status() int {
	return l.status
}

func (l *responseLogger) Size() int {
	return l.size
}

func (l *responseLogger) Flush() {
	f, ok := l.w.(http.Flusher)
	if ok {
		f.Flush()
	}
}

// SetURLmap sets or updates the URL-to-backend name map
func SetURLmap(urlmap map[string]string) {
	url2Backend = urlmap
}
