package main

import (
	"bytes"
	"github.com/containous/traefik/plugin"
	"io"
	"io/ioutil"
	"net/http"
)

var _ plugin.Middleware = (*Middleware)(nil)

// Middleware is a working Middleware plugin
type Middleware struct {
}

// Load loads the plugin instance
func Load() interface{} {
	return &Middleware{}
}

// ServeHTTP implements the Middleware interface
func (s *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	r.Header.Add("Traefik-Plugin-Middleware", "plugin.middleware.request.header")
	w.Header().Add("Traefik-Plugin-Middleware", "plugin.middleware.response.header")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	body = append(body, []byte("\nplugin.middleware.reponse.body")...)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	// Create a response wrapper:
	mrw := &myResponseWriter{
		ResponseWriter: w,
		buf:            &bytes.Buffer{},
	}

	next(w, r)

	if _, err := io.Copy(w, mrw.buf); err != nil {
		return
	}
}

type myResponseWriter struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (mrw *myResponseWriter) Write(p []byte) (int, error) {
	return mrw.buf.Write(p)
}
