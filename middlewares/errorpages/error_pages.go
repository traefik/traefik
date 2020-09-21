package errorpages

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/traefik/traefik/log"
	"github.com/traefik/traefik/middlewares"
	"github.com/traefik/traefik/types"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/utils"
)

// Compile time validation that the response recorder implements http interfaces correctly.
var (
	_ middlewares.Stateful = &responseRecorderWithCloseNotify{}
	_ middlewares.Stateful = &codeCatcherWithCloseNotify{}
)

// Handler is a middleware that provides the custom error pages
type Handler struct {
	BackendName    string
	backendHandler http.Handler
	httpCodeRanges types.HTTPCodeRanges
	backendURL     string
	backendQuery   string
	FallbackURL    string // Deprecated
}

// NewHandler initializes the utils.ErrorHandler for the custom error pages
func NewHandler(errorPage *types.ErrorPage, backendName string) (*Handler, error) {
	if len(backendName) == 0 {
		return nil, errors.New("error pages: backend name is mandatory ")
	}

	httpCodeRanges, err := types.NewHTTPCodeRanges(errorPage.Status)
	if err != nil {
		return nil, err
	}

	return &Handler{
		BackendName:    backendName,
		httpCodeRanges: httpCodeRanges,
		backendQuery:   errorPage.Query,
		backendURL:     "http://0.0.0.0",
	}, nil
}

// PostLoad adds backend handler if available
func (h *Handler) PostLoad(backendHandler http.Handler) error {
	if backendHandler == nil {
		fwd, err := forward.New()
		if err != nil {
			return err
		}

		h.backendHandler = fwd
		h.backendURL = h.FallbackURL
	} else {
		h.backendHandler = backendHandler
	}

	return nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	if h.backendHandler == nil {
		log.Error("Error pages: no backend handler.")
		next.ServeHTTP(w, req)
		return
	}

	catcher := newCodeCatcher(w, h.httpCodeRanges)
	next.ServeHTTP(catcher, req)
	if !catcher.isFilteredCode() {
		return
	}

	// check the recorder code against the configured http status code ranges
	code := catcher.getCode()
	for _, block := range h.httpCodeRanges {
		if code >= block[0] && code <= block[1] {
			log.Errorf("Caught HTTP Status Code %d, returning error page", code)

			var query string
			if len(h.backendQuery) > 0 {
				query = "/" + strings.TrimPrefix(h.backendQuery, "/")
				query = strings.Replace(query, "{status}", strconv.Itoa(code), -1)
			}

			pageReq, err := newRequest(h.backendURL + query)
			if err != nil {
				log.Error(err)
				w.WriteHeader(code)
				fmt.Fprint(w, http.StatusText(code))
				return
			}

			recorderErrorPage := newResponseRecorder(w)
			utils.CopyHeaders(pageReq.Header, req.Header)

			h.backendHandler.ServeHTTP(recorderErrorPage, pageReq.WithContext(req.Context()))

			utils.CopyHeaders(w.Header(), recorderErrorPage.Header())
			w.WriteHeader(code)
			w.Write(recorderErrorPage.GetBody().Bytes())
			return
		}
	}
}

func newRequest(baseURL string) (*http.Request, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error pages: error when parse URL: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error pages: error when create query: %v", err)
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

// codeCatcher is a response writer that detects as soon as possible whether the
// response is a code within the ranges of codes it watches for. If it is, it
// simply drops the data from the response. Otherwise, it forwards it directly to
// the original client (its responseWriter) without any buffering.
type codeCatcher struct {
	headerMap          http.Header
	code               int
	httpCodeRanges     types.HTTPCodeRanges
	firstWrite         bool
	caughtFilteredCode bool
	responseWriter     http.ResponseWriter
	headersSent        bool
	err                error
}

type codeCatcherWithCloseNotify struct {
	*codeCatcher
}

// CloseNotify returns a channel that receives at most a
// single value (true) when the client connection has gone away.
func (cc *codeCatcherWithCloseNotify) CloseNotify() <-chan bool {
	return cc.responseWriter.(http.CloseNotifier).CloseNotify()
}

func newCodeCatcher(rw http.ResponseWriter, httpCodeRanges types.HTTPCodeRanges) responseInterceptor {
	catcher := &codeCatcher{
		headerMap:      make(http.Header),
		code:           http.StatusOK, // If backend does not call WriteHeader on us, we consider it's a 200.
		responseWriter: rw,
		httpCodeRanges: httpCodeRanges,
		firstWrite:     true,
	}
	if _, ok := rw.(http.CloseNotifier); ok {
		return &codeCatcherWithCloseNotify{catcher}
	}
	return catcher
}

func (cc *codeCatcher) Header() http.Header {
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
	if !cc.firstWrite {
		if cc.caughtFilteredCode {
			// We don't care about the contents of the response,
			// since we want to serve the ones from the error page,
			// so we just drop them.
			return len(buf), nil
		}
		return cc.responseWriter.Write(buf)
	}
	cc.firstWrite = false

	// If WriteHeader was already called from the caller, this is a NOOP.
	// Otherwise, cc.code is actually a 200 here.
	cc.WriteHeader(cc.code)

	if cc.caughtFilteredCode {
		return len(buf), nil
	}
	return cc.responseWriter.Write(buf)
}

func (cc *codeCatcher) WriteHeader(code int) {
	if cc.headersSent || cc.caughtFilteredCode {
		return
	}

	cc.code = code
	for _, block := range cc.httpCodeRanges {
		if cc.code >= block[0] && cc.code <= block[1] {
			cc.caughtFilteredCode = true
			break
		}
	}
	// it will be up to the other response recorder to send the headers,
	// so it is out of our hands now.
	if cc.caughtFilteredCode {
		return
	}
	utils.CopyHeaders(cc.responseWriter.Header(), cc.Header())
	cc.responseWriter.WriteHeader(cc.code)
	cc.headersSent = true
}

// Hijack hijacks the connection
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

	if flusher, ok := cc.responseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

type responseRecorder interface {
	http.ResponseWriter
	http.Flusher
	GetCode() int
	GetBody() *bytes.Buffer
	IsStreamingResponseStarted() bool
}

// newResponseRecorder returns an initialized responseRecorder.
func newResponseRecorder(rw http.ResponseWriter) responseRecorder {
	recorder := &responseRecorderWithoutCloseNotify{
		HeaderMap:      make(http.Header),
		Body:           new(bytes.Buffer),
		Code:           http.StatusOK,
		responseWriter: rw,
	}
	if _, ok := rw.(http.CloseNotifier); ok {
		return &responseRecorderWithCloseNotify{recorder}
	}
	return recorder
}

// responseRecorderWithoutCloseNotify is an implementation of http.ResponseWriter that
// records its mutations for later inspection.
type responseRecorderWithoutCloseNotify struct {
	Code      int           // the HTTP response code from WriteHeader
	HeaderMap http.Header   // the HTTP response headers
	Body      *bytes.Buffer // if non-nil, the bytes.Buffer to append written data to

	responseWriter           http.ResponseWriter
	err                      error
	streamingResponseStarted bool
}

type responseRecorderWithCloseNotify struct {
	*responseRecorderWithoutCloseNotify
}

// CloseNotify returns a channel that receives at most a
// single value (true) when the client connection has gone away.
func (r *responseRecorderWithCloseNotify) CloseNotify() <-chan bool {
	return r.responseWriter.(http.CloseNotifier).CloseNotify()
}

// Header returns the response headers.
func (r *responseRecorderWithoutCloseNotify) Header() http.Header {
	if r.HeaderMap == nil {
		r.HeaderMap = make(http.Header)
	}

	return r.HeaderMap
}

func (r *responseRecorderWithoutCloseNotify) GetCode() int {
	return r.Code
}

func (r *responseRecorderWithoutCloseNotify) GetBody() *bytes.Buffer {
	return r.Body
}

func (r *responseRecorderWithoutCloseNotify) IsStreamingResponseStarted() bool {
	return r.streamingResponseStarted
}

// Write always succeeds and writes to rw.Body, if not nil.
func (r *responseRecorderWithoutCloseNotify) Write(buf []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.Body.Write(buf)
}

// WriteHeader sets rw.Code.
func (r *responseRecorderWithoutCloseNotify) WriteHeader(code int) {
	r.Code = code
}

// Hijack hijacks the connection
func (r *responseRecorderWithoutCloseNotify) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.responseWriter.(http.Hijacker).Hijack()
}

// Flush sends any buffered data to the client.
func (r *responseRecorderWithoutCloseNotify) Flush() {
	if !r.streamingResponseStarted {
		utils.CopyHeaders(r.responseWriter.Header(), r.Header())
		r.responseWriter.WriteHeader(r.Code)
		r.streamingResponseStarted = true
	}

	_, err := r.responseWriter.Write(r.Body.Bytes())
	if err != nil {
		log.Errorf("Error writing response in responseRecorder: %v", err)
		r.err = err
	}
	r.Body.Reset()

	if flusher, ok := r.responseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
