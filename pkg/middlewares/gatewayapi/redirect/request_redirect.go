package redirect

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/vulcand/oxy/v2/utils"
	"go.opentelemetry.io/otel/trace"
)

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
	typeName    = "RequestRedirect"
)

var uriRegexp = regexp.MustCompile(`^(https?):\/\/(\[[\w:.]+\]|[\w\._-]+)?(:\d+)?(.*)$`)

// NewRequestRedirect creates a redirect middleware.
func NewRequestRedirect(ctx context.Context, next http.Handler, conf dynamic.RequestRedirect, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")
	logger.Debug().Msgf("Setting up redirection from %s to %s", conf.Regex, conf.Replacement)

	re, err := regexp.Compile(conf.Regex)
	if err != nil {
		return nil, err
	}

	return &redirect{
		regex:       re,
		replacement: conf.Replacement,
		permanent:   conf.Permanent,
		errHandler:  utils.DefaultHandler,
		next:        next,
		name:        name,
		rawURL:      rawURL,
	}, nil
}

type redirect struct {
	next        http.Handler
	regex       *regexp.Regexp
	replacement string
	permanent   bool
	errHandler  utils.ErrorHandler
	name        string
	rawURL      func(*http.Request) string
}

func rawURL(req *http.Request) string {
	scheme := schemeHTTP
	host := req.Host
	port := ""
	uri := req.RequestURI

	if match := uriRegexp.FindStringSubmatch(req.RequestURI); len(match) > 0 {
		scheme = match[1]

		if len(match[2]) > 0 {
			host = match[2]
		}

		if len(match[3]) > 0 {
			port = match[3]
		}

		uri = match[4]
	}

	if req.TLS != nil {
		scheme = schemeHTTPS
	}

	return strings.Join([]string{scheme, "://", host, port, uri}, "")
}

func (r *redirect) GetTracingInformation() (string, string, trace.SpanKind) {
	return r.name, typeName, trace.SpanKindInternal
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

	handler := &moveHandler{location: parsedURL, permanent: r.permanent}
	handler.ServeHTTP(rw, req)
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
