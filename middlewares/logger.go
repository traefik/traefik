package middlewares

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/context"
)

// Logger is a middleware handler that logs the request as it goes in and the response as it goes out.
type Logger struct {
	file *os.File
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
		myCombinedLoggingHandler(l.file, next).ServeHTTP(rw, r)
	}
}

// Close closes the logger (i.e. the file).
func (l *Logger) Close() {
	l.file.Close()
}

func myCombinedLoggingHandler(out io.Writer, h http.Handler) http.Handler {
	return combinedLoggingHandler{out, h}
}

type combinedLoggingHandler struct {
	writer  io.Writer
	handler http.Handler
}

func (h combinedLoggingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer context.Clear(req)

	t_start := time.Now()
	context.Set(req, "frontend", "Unknown frontend")
	context.Set(req, "oxy_backend", "Unknown backend")
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

	ts := t_start.Format("02/Jan/2006:15:04:05 -0700")
	method := req.Method
	uri := url.RequestURI()
	proto := req.Proto
	status := logger.Status()
	len := logger.Size()
	referer := req.Referer()
	agent := req.UserAgent()
	frontend := context.Get(req, "frontend")
	backend := context.Get(req, "oxy_backend")
	elapsed := time.Now().UTC().Sub(t_start.UTC())

	fmt.Fprintf(h.writer, `%s - %s [%s] "%s %s %s" %d %d "%s" "%s" "%s" "%s" %s%s`,
		host, username, ts, method, uri, proto, status, len, referer, agent, frontend, backend, elapsed, "\n")

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
