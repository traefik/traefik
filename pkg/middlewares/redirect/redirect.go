package redirect

import (
	"net/http"
	"net/url"
	"regexp"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/tracing"
	"github.com/vulcand/oxy/v2/utils"
)

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

var uriRegexp = regexp.MustCompile(`^(https?):\/\/(\[[\w:.]+\]|[\w\._-]+)?(:\d+)?(.*)$`)

type redirect struct {
	next        http.Handler
	regex       *regexp.Regexp
	replacement string
	permanent   bool
	errHandler  utils.ErrorHandler
	name        string
	rawURL      func(*http.Request) string
}

// New creates a Redirect middleware.
func newRedirect(next http.Handler, regex, replacement string, permanent bool, rawURL func(*http.Request) string, name string) (http.Handler, error) {
	re, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}

	return &redirect{
		regex:       re,
		replacement: replacement,
		permanent:   permanent,
		errHandler:  utils.DefaultHandler,
		next:        next,
		name:        name,
		rawURL:      rawURL,
	}, nil
}

func (r *redirect) GetTracingInformation() (string, ext.SpanKindEnum) {
	return r.name, tracing.SpanKindNoneEnum
}

func (r *redirect) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	oldURL := r.rawURL(req)

	// If the Regexp doesn't match, skip to the next handler.
	if !r.regex.MatchString(oldURL) {
		r.next.ServeHTTP(rw, req)
		return
	}

	// Apply a rewrite regexp to the URL.
	newURL := r.regex.ReplaceAllString(oldURL, r.replacement)

	// Parse the rewritten URL and replace request URL with it.
	parsedURL, err := url.Parse(newURL)
	if err != nil {
		r.errHandler.ServeHTTP(rw, req, err)
		return
	}

	if newURL != oldURL {
		handler := &moveHandler{location: parsedURL, permanent: r.permanent}
		handler.ServeHTTP(rw, req)
		return
	}

	req.URL = parsedURL

	// Make sure the request URI corresponds the rewritten URL.
	req.RequestURI = req.URL.RequestURI()
	r.next.ServeHTTP(rw, req)
}

type moveHandler struct {
	location  *url.URL
	permanent bool
}

func (m *moveHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Location", m.location.String())

	status := http.StatusFound
	if req.Method != http.MethodGet {
		status = http.StatusTemporaryRedirect
	}

	if m.permanent {
		status = http.StatusMovedPermanently
		if req.Method != http.MethodGet {
			status = http.StatusPermanentRedirect
		}
	}
	rw.WriteHeader(status)
	_, err := rw.Write([]byte(http.StatusText(status)))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}
