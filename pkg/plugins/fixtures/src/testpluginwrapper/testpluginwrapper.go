package testpluginwrapper

import (
	"bufio"
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
)

type Config struct{}

func CreateConfig() *Config {
	return &Config{}
}

type plugin struct {
	next http.Handler
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &plugin{next: next}, nil
}

func (p *plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	w := &statusWriter{ResponseWriter: rw}
	p.next.ServeHTTP(w, req)
	rw.Header().Set("X-Status", strconv.Itoa(w.status))
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("not a hijacker")
	}
	return h.Hijack()
}
