package cbreaker

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	if f.r.ContentType != "" {
		w.Header().Set("Content-Type", f.r.ContentType)
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(f.r.Body)))
	w.WriteHeader(f.r.StatusCode)
	w.Write(f.r.Body)
}

type Redirect struct {
	URL string
}

type RedirectFallback struct {
	u *url.URL
}

func NewRedirectFallback(r Redirect) (*RedirectFallback, error) {
	u, err := url.ParseRequestURI(r.URL)
	if err != nil {
		return nil, err
	}
	return &RedirectFallback{u: u}, nil
}

func (f *RedirectFallback) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Location", f.u.String())
	w.WriteHeader(http.StatusFound)
	w.Write([]byte(http.StatusText(http.StatusFound)))
}
