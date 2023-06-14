package customerrors

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
	"github.com/traefik/traefik/v2/pkg/types"
	"github.com/vulcand/oxy/v2/utils"
)

// Compile time validation that the response recorder implements http interfaces correctly.
var (
	// TODO: maybe remove at least for codeModifierWithCloseNotify.
	_ middlewares.Stateful = &codeModifierWithCloseNotify{}
	_ middlewares.Stateful = &codeCatcherWithCloseNotify{}
)

const typeName = "customError"

type serviceBuilder interface {
	BuildHTTP(ctx context.Context, serviceName string) (http.Handler, error)
}

// customErrors is a middleware that provides the custom error pages.
type customErrors struct {
	name           string
	next           http.Handler
	backendHandler http.Handler
	httpCodeRanges types.HTTPCodeRanges
	backendQuery   string
}

// New creates a new custom error pages middleware.
func New(ctx context.Context, next http.Handler, config dynamic.ErrorPage, serviceBuilder serviceBuilder, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

	httpCodeRanges, err := types.NewHTTPCodeRanges(config.Status)
	if err != nil {
		return nil, err
	}

	backend, err := serviceBuilder.BuildHTTP(ctx, config.Service)
	if err != nil {
		return nil, err
	}

	return &customErrors{
		name:           name,
		next:           next,
		backendHandler: backend,
		httpCodeRanges: httpCodeRanges,
		backendQuery:   config.Query,
	}, nil
}

func (c *customErrors) GetTracingInformation() (string, ext.SpanKindEnum) {
	return c.name, tracing.SpanKindNoneEnum
}

func (c *customErrors) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := middlewares.GetLoggerCtx(req.Context(), c.name, typeName)
	logger := log.FromContext(ctx)

	if c.backendHandler == nil {
		logger.Error("Error pages: no backend handler.")
		tracing.SetErrorWithEvent(req, "Error pages: no backend handler.")
		c.next.ServeHTTP(rw, req)
		return
	}

	catcher := newCodeCatcher(rw, c.httpCodeRanges)
	c.next.ServeHTTP(catcher, req)
	if !catcher.isFilteredCode() {
		return
	}

	// check the recorder code against the configured http status code ranges
	code := catcher.getCode()
	logger.Debugf("Caught HTTP Status Code %d, returning error page", code)

	var query string
	if len(c.backendQuery) > 0 {
		query = "/" + strings.TrimPrefix(c.backendQuery, "/")
		query = strings.ReplaceAll(query, "{status}", strconv.Itoa(code))
		query = strings.ReplaceAll(query, "{url}", url.QueryEscape(req.URL.String()))
	}

	pageReq, err := newRequest("http://" + req.Host + query)
	if err != nil {
		logger.Error(err)
		http.Error(rw, http.StatusText(code), code)
		return
	}

	utils.CopyHeaders(pageReq.Header, req.Header)

	c.backendHandler.ServeHTTP(newCodeModifier(rw, code),
		pageReq.WithContext(req.Context()))
}

func newRequest(baseURL string) (*http.Request, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error pages: error when parse URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error pages: error when create query: %w", err)
	}

	req.RequestURI = u.RequestURI()
	return req, nil
}

type responseInterceptor interface {
	http.ResponseWriter
	http.Flusher
	getCode() int
	isFilteredCode() bool
}

// codeCatcher is a response writer that detects as soon as possible
// whether the response is a code within the ranges of codes it watches for.
// If it is, it simply drops the data from the response.
// Otherwise, it forwards it directly to the original client (its responseWriter) without any buffering.
type codeCatcher struct {
	headerMap          http.Header
	code               int
	httpCodeRanges     types.HTTPCodeRanges
	caughtFilteredCode bool
	responseWriter     http.ResponseWriter
	headersSent        bool
}

func newCodeCatcher(rw http.ResponseWriter, httpCodeRanges types.HTTPCodeRanges) responseInterceptor {
	catcher := &codeCatcher{
		headerMap:      make(http.Header),
		code:           http.StatusOK, // If backend does not call WriteHeader on us, we consider it's a 200.
		responseWriter: rw,
		httpCodeRanges: httpCodeRanges,
	}
	if _, ok := rw.(http.CloseNotifier); ok {
		return &codeCatcherWithCloseNotify{catcher}
	}
	return catcher
}

func (cc *codeCatcher) Header() http.Header {
	if cc.headersSent {
		return cc.responseWriter.Header()
	}

	if cc.headerMap == nil {
		cc.headerMap = make(http.Header)
	}

	return cc.headerMap
}

func (cc *codeCatcher) getCode() int {
	return cc.code
}

// isFilteredCode returns whether the codeCatcher received a response code among the ones it is watching,
// and for which the response should be deferred to the error handler.
func (cc *codeCatcher) isFilteredCode() bool {
	return cc.caughtFilteredCode
}

func (cc *codeCatcher) Write(buf []byte) (int, error) {
	// If WriteHeader was already called from the caller, this is a NOOP.
	// Otherwise, cc.code is actually a 200 here.
	cc.WriteHeader(cc.code)

	if cc.caughtFilteredCode {
		// We don't care about the contents of the response,
		// since we want to serve the ones from the error page,
		// so we just drop them.
		return len(buf), nil
	}
	return cc.responseWriter.Write(buf)
}

// WriteHeader is, in the specific case of 1xx status codes, a direct call to the wrapped ResponseWriter, without marking headers as sent,
// allowing so further calls.
func (cc *codeCatcher) WriteHeader(code int) {
	if cc.headersSent || cc.caughtFilteredCode {
		return
	}

	// Handling informational headers.
	if code >= 100 && code <= 199 {
		// Multiple informational status codes can be used,
		// so here the copy is not appending the values to not repeat them.
		for k, v := range cc.Header() {
			cc.responseWriter.Header()[k] = v
		}

		cc.responseWriter.WriteHeader(code)
		return
	}

	cc.code = code
	for _, block := range cc.httpCodeRanges {
		if cc.code >= block[0] && cc.code <= block[1] {
			cc.caughtFilteredCode = true
			// it will be up to the caller to send the headers,
			// so it is out of our hands now.
			return
		}
	}

	// The copy is not appending the values,
	// to not repeat them in case any informational status code has been written.
	for k, v := range cc.Header() {
		cc.responseWriter.Header()[k] = v
	}
	cc.responseWriter.WriteHeader(cc.code)
	cc.headersSent = true
}

// Hijack hijacks the connection.
func (cc *codeCatcher) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := cc.responseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("%T is not a http.Hijacker", cc.responseWriter)
}

// Flush sends any buffered data to the client.
func (cc *codeCatcher) Flush() {
	// If WriteHeader was already called from the caller, this is a NOOP.
	// Otherwise, cc.code is actually a 200 here.
	cc.WriteHeader(cc.code)

	// We don't care about the contents of the response,
	// since we want to serve the ones from the error page,
	// so we just don't flush.
	// (e.g., To prevent superfluous WriteHeader on request with a
	// `Transfert-Encoding: chunked` header).
	if cc.caughtFilteredCode {
		return
	}

	if flusher, ok := cc.responseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

type codeCatcherWithCloseNotify struct {
	*codeCatcher
}

// CloseNotify returns a channel that receives at most a single value (true)
// when the client connection has gone away.
func (cc *codeCatcherWithCloseNotify) CloseNotify() <-chan bool {
	return cc.responseWriter.(http.CloseNotifier).CloseNotify()
}

// codeModifier forwards a response back to the client,
// while enforcing a given response code.
type codeModifier interface {
	http.ResponseWriter
}

// newCodeModifier returns a codeModifier that enforces the given code.
func newCodeModifier(rw http.ResponseWriter, code int) codeModifier {
	codeMod := &codeModifierWithoutCloseNotify{
		headerMap:      make(http.Header),
		code:           code,
		responseWriter: rw,
	}
	if _, ok := rw.(http.CloseNotifier); ok {
		return &codeModifierWithCloseNotify{codeMod}
	}
	return codeMod
}

type codeModifierWithoutCloseNotify struct {
	code int // the code enforced in the response.

	// headerSent is whether the headers have already been sent,
	// either through Write or WriteHeader.
	headerSent bool
	headerMap  http.Header // the HTTP response headers from the backend.

	responseWriter http.ResponseWriter
}

// Header returns the response headers.
func (r *codeModifierWithoutCloseNotify) Header() http.Header {
	if r.headerSent {
		return r.responseWriter.Header()
	}

	if r.headerMap == nil {
		r.headerMap = make(http.Header)
	}

	return r.headerMap
}

// Write calls WriteHeader to send the enforced code,
// then writes the data directly to r.responseWriter.
func (r *codeModifierWithoutCloseNotify) Write(buf []byte) (int, error) {
	r.WriteHeader(r.code)
	return r.responseWriter.Write(buf)
}

// WriteHeader sends the headers, with the enforced code (the code in argument is always ignored),
// if it hasn't already been done.
// WriteHeader is, in the specific case of 1xx status codes, a direct call to the wrapped ResponseWriter, without marking headers as sent,
// allowing so further calls.
func (r *codeModifierWithoutCloseNotify) WriteHeader(code int) {
	if r.headerSent {
		return
	}

	// Handling informational headers.
	if code >= 100 && code <= 199 {
		// Multiple informational status codes can be used,
		// so here the copy is not appending the values to not repeat them.
		for k, v := range r.headerMap {
			r.responseWriter.Header()[k] = v
		}

		r.responseWriter.WriteHeader(code)
		return
	}

	for k, v := range r.headerMap {
		r.responseWriter.Header()[k] = v
	}
	r.responseWriter.WriteHeader(r.code)
	r.headerSent = true
}

// Hijack hijacks the connection.
func (r *codeModifierWithoutCloseNotify) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.responseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("%T is not a http.Hijacker", r.responseWriter)
	}
	return hijacker.Hijack()
}

// Flush sends any buffered data to the client.
func (r *codeModifierWithoutCloseNotify) Flush() {
	r.WriteHeader(r.code)

	if flusher, ok := r.responseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

type codeModifierWithCloseNotify struct {
	*codeModifierWithoutCloseNotify
}

// CloseNotify returns a channel that receives at most a single value (true)
// when the client connection has gone away.
func (r *codeModifierWithCloseNotify) CloseNotify() <-chan bool {
	return r.responseWriter.(http.CloseNotifier).CloseNotify()
}
