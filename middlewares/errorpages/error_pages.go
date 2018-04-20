package errorpages

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/types"
	"github.com/pkg/errors"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/utils"
)

// Compile time validation that the response recorder implements http interfaces correctly.
var _ middlewares.Stateful = &responseRecorderWithCloseNotify{}

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

	recorder := newResponseRecorder(w)
	next.ServeHTTP(recorder, req)

	w.WriteHeader(recorder.GetCode())

	// check the recorder code against the configured http status code ranges
	for _, block := range h.httpCodeRanges {
		if recorder.GetCode() >= block[0] && recorder.GetCode() <= block[1] {
			log.Errorf("Caught HTTP Status Code %d, returning error page", recorder.GetCode())

			var query string
			if len(h.backendQuery) > 0 {
				query = "/" + strings.TrimPrefix(h.backendQuery, "/")
				query = strings.Replace(query, "{status}", strconv.Itoa(recorder.GetCode()), -1)
			}

			if u, err := url.Parse(h.backendURL + query); err != nil {
				log.Errorf("error pages: error when parse URL: %v", err)
				w.Write([]byte(http.StatusText(recorder.GetCode())))
			} else if newReq, err := http.NewRequest(http.MethodGet, u.String(), nil); err != nil {
				log.Errorf("error pages: error when create query: %v", err)
				w.Write([]byte(http.StatusText(recorder.GetCode())))
			} else {
				utils.CopyHeaders(newReq.Header, req.Header)
				newReq.RequestURI = u.RequestURI()
				h.backendHandler.ServeHTTP(w, newReq)
			}
			return
		}
	}

	// did not catch a configured status code so proceed with the request
	utils.CopyHeaders(w.Header(), recorder.Header())
	w.Write(recorder.GetBody().Bytes())
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
func (rw *responseRecorderWithCloseNotify) CloseNotify() <-chan bool {
	return rw.responseWriter.(http.CloseNotifier).CloseNotify()
}

// Header returns the response headers.
func (rw *responseRecorderWithoutCloseNotify) Header() http.Header {
	if rw.HeaderMap == nil {
		rw.HeaderMap = make(http.Header)
	}
	return rw.HeaderMap
}

func (rw *responseRecorderWithoutCloseNotify) GetCode() int {
	return rw.Code
}

func (rw *responseRecorderWithoutCloseNotify) GetBody() *bytes.Buffer {
	return rw.Body
}

func (rw *responseRecorderWithoutCloseNotify) IsStreamingResponseStarted() bool {
	return rw.streamingResponseStarted
}

// Write always succeeds and writes to rw.Body, if not nil.
func (rw *responseRecorderWithoutCloseNotify) Write(buf []byte) (int, error) {
	if rw.err != nil {
		return 0, rw.err
	}
	return rw.Body.Write(buf)
}

// WriteHeader sets rw.Code.
func (rw *responseRecorderWithoutCloseNotify) WriteHeader(code int) {
	rw.Code = code
}

// Hijack hijacks the connection
func (rw *responseRecorderWithoutCloseNotify) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.responseWriter.(http.Hijacker).Hijack()
}

// Flush sends any buffered data to the client.
func (rw *responseRecorderWithoutCloseNotify) Flush() {
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

	if flusher, ok := rw.responseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
