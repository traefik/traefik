package redirect

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"text/template"

	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/middlewares"
	"github.com/urfave/negroni"
	"github.com/vulcand/oxy/utils"
)

const (
	defaultRedirectRegex = `^(?:https?:\/\/)?([\w\._-]+)(?::\d+)?(.*)$`
)

// NewEntryPointHandler create a new redirection handler base on entry point
func NewEntryPointHandler(dstEntryPoint *configuration.EntryPoint, permanent bool) (negroni.Handler, error) {
	exp := regexp.MustCompile(`(:\d+)`)
	match := exp.FindStringSubmatch(dstEntryPoint.Address)
	if len(match) == 0 {
		return nil, fmt.Errorf("bad Address format %q", dstEntryPoint.Address)
	}

	protocol := "http"
	if dstEntryPoint.TLS != nil {
		protocol = "https"
	}

	replacement := protocol + "://${1}" + match[0] + "${2}"

	return NewRegexHandler(defaultRedirectRegex, replacement, permanent)
}

// NewRegexHandler create a new redirection handler base on regex
func NewRegexHandler(exp string, replacement string, permanent bool) (negroni.Handler, error) {
	re, err := regexp.Compile(exp)
	if err != nil {
		return nil, err
	}

	return &handler{
		regexp:      re,
		replacement: replacement,
		permanent:   permanent,
		errHandler:  utils.DefaultHandler,
	}, nil
}

type handler struct {
	regexp      *regexp.Regexp
	replacement string
	permanent   bool
	errHandler  utils.ErrorHandler
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	oldURL := rawURL(req)

	// only continue if the Regexp param matches the URL
	if !h.regexp.MatchString(oldURL) {
		next.ServeHTTP(rw, req)
		return
	}

	// apply a rewrite regexp to the URL
	newURL := h.regexp.ReplaceAllString(oldURL, h.replacement)

	// replace any variables that may be in there
	rewrittenURL := &bytes.Buffer{}
	if err := applyString(newURL, rewrittenURL, req); err != nil {
		h.errHandler.ServeHTTP(rw, req, err)
		return
	}

	// parse the rewritten URL and replace request URL with it
	parsedURL, err := url.Parse(rewrittenURL.String())
	if err != nil {
		h.errHandler.ServeHTTP(rw, req, err)
		return
	}

	if stripPrefix, stripPrefixOk := req.Context().Value(middlewares.StripPrefixKey).(string); stripPrefixOk {
		if len(stripPrefix) > 0 {
			parsedURL.Path = stripPrefix
		}
	}

	if addPrefix, addPrefixOk := req.Context().Value(middlewares.AddPrefixKey).(string); addPrefixOk {
		if len(addPrefix) > 0 {
			parsedURL.Path = strings.Replace(parsedURL.Path, addPrefix, "", 1)
		}
	}

	if replacePath, replacePathOk := req.Context().Value(middlewares.ReplacePathKey).(string); replacePathOk {
		if len(replacePath) > 0 {
			parsedURL.Path = replacePath
		}
	}

	if newURL != oldURL {
		handler := &moveHandler{location: parsedURL, permanent: h.permanent}
		handler.ServeHTTP(rw, req)
		return
	}

	req.URL = parsedURL

	// make sure the request URI corresponds the rewritten URL
	req.RequestURI = req.URL.RequestURI()
	next.ServeHTTP(rw, req)
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
	rw.Write([]byte(http.StatusText(status)))
}

func rawURL(request *http.Request) string {
	scheme := "http"
	if request.TLS != nil || isXForwardedHTTPS(request) {
		scheme = "https"
	}

	return strings.Join([]string{scheme, "://", request.Host, request.RequestURI}, "")
}

func isXForwardedHTTPS(request *http.Request) bool {
	xForwardedProto := request.Header.Get("X-Forwarded-Proto")

	return len(xForwardedProto) > 0 && xForwardedProto == "https"
}

func applyString(in string, out io.Writer, request *http.Request) error {
	t, err := template.New("t").Parse(in)
	if err != nil {
		return err
	}

	data := struct {
		Request *http.Request
	}{
		Request: request,
	}

	return t.Execute(out, data)
}
