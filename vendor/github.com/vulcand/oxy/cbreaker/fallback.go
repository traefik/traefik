package cbreaker

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/utils"
)

// Response response model
type Response struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

// ResponseFallback fallback response handler
type ResponseFallback struct {
	r Response

	log *log.Logger
}

// NewResponseFallbackWithLogger creates a new ResponseFallback
func NewResponseFallbackWithLogger(r Response, l *log.Logger) (*ResponseFallback, error) {
	if r.StatusCode == 0 {
		return nil, fmt.Errorf("response code should not be 0")
	}
	return &ResponseFallback{r: r, log: l}, nil
}

// NewResponseFallback creates a new ResponseFallback
func NewResponseFallback(r Response) (*ResponseFallback, error) {
	return NewResponseFallbackWithLogger(r, log.StandardLogger())
}

func (f *ResponseFallback) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if f.log.Level >= log.DebugLevel {
		logEntry := f.log.WithField("Request", utils.DumpHttpRequest(req))
		logEntry.Debug("vulcand/oxy/fallback/response: begin ServeHttp on request")
		defer logEntry.Debug("vulcand/oxy/fallback/response: completed ServeHttp on request")
	}

	if f.r.ContentType != "" {
		w.Header().Set("Content-Type", f.r.ContentType)
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(f.r.Body)))
	w.WriteHeader(f.r.StatusCode)
	_, err := w.Write(f.r.Body)
	if err != nil {
		f.log.Errorf("vulcand/oxy/fallback/response: failed to write response, err: %v", err)
	}
}

// Redirect redirect model
type Redirect struct {
	URL          string
	PreservePath bool
}

// RedirectFallback fallback redirect handler
type RedirectFallback struct {
	r Redirect

	u *url.URL

	log *log.Logger
}

// NewRedirectFallbackWithLogger creates a new RedirectFallback
func NewRedirectFallbackWithLogger(r Redirect, l *log.Logger) (*RedirectFallback, error) {
	u, err := url.ParseRequestURI(r.URL)
	if err != nil {
		return nil, err
	}
	return &RedirectFallback{r: r, u: u, log: l}, nil
}

// NewRedirectFallback creates a new RedirectFallback
func NewRedirectFallback(r Redirect) (*RedirectFallback, error) {
	return NewRedirectFallbackWithLogger(r, log.StandardLogger())
}

func (f *RedirectFallback) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if f.log.Level >= log.DebugLevel {
		logEntry := f.log.WithField("Request", utils.DumpHttpRequest(req))
		logEntry.Debug("vulcand/oxy/fallback/redirect: begin ServeHttp on request")
		defer logEntry.Debug("vulcand/oxy/fallback/redirect: completed ServeHttp on request")
	}

	location := f.u.String()
	if f.r.PreservePath {
		location += req.URL.Path
	}

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusFound)
	_, err := w.Write([]byte(http.StatusText(http.StatusFound)))
	if err != nil {
		f.log.Errorf("vulcand/oxy/fallback/redirect: failed to write response, err: %v", err)
	}
}
