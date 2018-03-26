package middlewares

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/utils"
)

// Compile time validation that the response recorder implements http interfaces correctly.
var _ Stateful = &errorPagesResponseRecorderWithCloseNotify{}

//ErrorPagesHandler is a middleware that provides the custom error pages
type ErrorPagesHandler struct {
	HTTPCodeRanges     types.HTTPCodeRanges
	BackendURL         string
	errorPageForwarder *forward.Forwarder
}

//NewErrorPagesHandler initializes the utils.ErrorHandler for the custom error pages
func NewErrorPagesHandler(errorPage *types.ErrorPage, backendURL string) (*ErrorPagesHandler, error) {
	fwd, err := forward.New()
	if err != nil {
		return nil, err
	}

	httpCodeRanges, err := types.NewHTTPCodeRanges(errorPage.Status)
	if err != nil {
		return nil, err
	}

	return &ErrorPagesHandler{
			HTTPCodeRanges:     httpCodeRanges,
			BackendURL:         backendURL + errorPage.Query,
			errorPageForwarder: fwd},
		nil
}

func (ep *ErrorPagesHandler) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	recorder := newErrorPagesResponseRecorder(w)

	next.ServeHTTP(recorder, req)

	w.WriteHeader(recorder.GetCode())
	//check the recorder code against the configured http status code ranges
	for _, block := range ep.HTTPCodeRanges {
		if recorder.GetCode() >= block[0] && recorder.GetCode() <= block[1] {
			log.Errorf("Caught HTTP Status Code %d, returning error page", recorder.GetCode())
			finalURL := strings.Replace(ep.BackendURL, "{status}", strconv.Itoa(recorder.GetCode()), -1)
			if newReq, err := http.NewRequest(http.MethodGet, finalURL, nil); err != nil {
				w.Write([]byte(http.StatusText(recorder.GetCode())))
			} else {
				ep.errorPageForwarder.ServeHTTP(w, newReq)
			}
			return
		}
	}

	//did not catch a configured status code so proceed with the request
	utils.CopyHeaders(w.Header(), recorder.Header())
	w.Write(recorder.GetBody().Bytes())
}

type errorPagesResponseRecorder interface {
	http.ResponseWriter
	http.Flusher
	GetCode() int
	GetBody() *bytes.Buffer
	IsStreamingResponseStarted() bool
}

// newErrorPagesResponseRecorder returns an initialized responseRecorder.
func newErrorPagesResponseRecorder(rw http.ResponseWriter) errorPagesResponseRecorder {
	recorder := &errorPagesResponseRecorderWithoutCloseNotify{
		HeaderMap:      make(http.Header),
		Body:           new(bytes.Buffer),
		Code:           http.StatusOK,
		responseWriter: rw,
	}
	if _, ok := rw.(http.CloseNotifier); ok {
		return &errorPagesResponseRecorderWithCloseNotify{recorder}
	}
	return recorder
}

// errorPagesResponseRecorderWithoutCloseNotify is an implementation of http.ResponseWriter that
// records its mutations for later inspection.
type errorPagesResponseRecorderWithoutCloseNotify struct {
	Code      int           // the HTTP response code from WriteHeader
	HeaderMap http.Header   // the HTTP response headers
	Body      *bytes.Buffer // if non-nil, the bytes.Buffer to append written data to

	responseWriter           http.ResponseWriter
	err                      error
	streamingResponseStarted bool
}

type errorPagesResponseRecorderWithCloseNotify struct {
	*errorPagesResponseRecorderWithoutCloseNotify
}

// CloseNotify returns a channel that receives at most a
// single value (true) when the client connection has gone
// away.
func (rw *errorPagesResponseRecorderWithCloseNotify) CloseNotify() <-chan bool {
	return rw.responseWriter.(http.CloseNotifier).CloseNotify()
}

// Header returns the response headers.
func (rw *errorPagesResponseRecorderWithoutCloseNotify) Header() http.Header {
	m := rw.HeaderMap
	if m == nil {
		m = make(http.Header)
		rw.HeaderMap = m
	}
	return m
}

func (rw *errorPagesResponseRecorderWithoutCloseNotify) GetCode() int {
	return rw.Code
}

func (rw *errorPagesResponseRecorderWithoutCloseNotify) GetBody() *bytes.Buffer {
	return rw.Body
}

func (rw *errorPagesResponseRecorderWithoutCloseNotify) IsStreamingResponseStarted() bool {
	return rw.streamingResponseStarted
}

// Write always succeeds and writes to rw.Body, if not nil.
func (rw *errorPagesResponseRecorderWithoutCloseNotify) Write(buf []byte) (int, error) {
	if rw.err != nil {
		return 0, rw.err
	}
	return rw.Body.Write(buf)
}

// WriteHeader sets rw.Code.
func (rw *errorPagesResponseRecorderWithoutCloseNotify) WriteHeader(code int) {
	rw.Code = code
}

// Hijack hijacks the connection
func (rw *errorPagesResponseRecorderWithoutCloseNotify) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.responseWriter.(http.Hijacker).Hijack()
}

// Flush sends any buffered data to the client.
func (rw *errorPagesResponseRecorderWithoutCloseNotify) Flush() {
	if !rw.streamingResponseStarted {
		utils.CopyHeaders(rw.responseWriter.Header(), rw.Header())
		rw.responseWriter.WriteHeader(rw.Code)
		rw.streamingResponseStarted = true
	}

	_, err := rw.responseWriter.Write(rw.Body.Bytes())
	if err != nil {
		log.Errorf("Error writing response in responseRecorder: %s", err)
		rw.err = err
	}
	rw.Body.Reset()
	flusher, ok := rw.responseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}
