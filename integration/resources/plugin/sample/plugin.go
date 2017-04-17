package main

import (
	"bytes"
	"github.com/containous/traefik/plugin"
	"io"
	"io/ioutil"
	"net/http"
)

var _ plugin.Middleware = (*Sample)(nil)

type Sample struct {
}

func Load() interface{} {
	return &Sample{}
}

func (s *Sample) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	r.Header.Add("Traefik-Plugin-Sample", "sample.request.header")
	w.Header().Add("Traefik-Plugin-Sample", "sample.response.header")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	body = append(body, []byte("\nsample.reponse.body")...)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	// Create a response wrapper:
	mrw := &MyResponseWriter{
		ResponseWriter: w,
		buf:            &bytes.Buffer{},
	}

	next(w, r)

	if _, err := io.Copy(w, mrw.buf); err != nil {
		return
	}
}

type MyResponseWriter struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (mrw *MyResponseWriter) Write(p []byte) (int, error) {
	return mrw.buf.Write(p)
}
