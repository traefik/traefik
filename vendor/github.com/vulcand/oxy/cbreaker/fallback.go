package cbreaker

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/utils"
)

type Response struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

type ResponseFallback struct {
	r Response
}

func NewResponseFallback(r Response) (*ResponseFallback, error) {
	if r.StatusCode == 0 {
		return nil, fmt.Errorf("response code should not be 0")
	}
	return &ResponseFallback{r: r}, nil
}

func (f *ResponseFallback) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if log.GetLevel() >= log.DebugLevel {
		logEntry := log.WithField("Request", utils.DumpHttpRequest(req))
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
		log.Errorf("vulcand/oxy/fallback/response: failed to write response, err: %v", err)
	}
}

type Redirect struct {
	URL          string
	PreservePath bool
}

type RedirectFallback struct {
	u *url.URL
	r Redirect
}

func NewRedirectFallback(r Redirect) (*RedirectFallback, error) {
	u, err := url.ParseRequestURI(r.URL)
	if err != nil {
		return nil, err
	}
	return &RedirectFallback{u: u, r: r}, nil
}

func (f *RedirectFallback) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if log.GetLevel() >= log.DebugLevel {
		logEntry := log.WithField("Request", utils.DumpHttpRequest(req))
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
		log.Errorf("vulcand/oxy/fallback/redirect: failed to write response, err: %v", err)
	}
}
