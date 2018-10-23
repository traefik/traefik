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

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/types"
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

	// check the recorder code against the configured http status code ranges
	for _, block := range h.httpCodeRanges {
		if recorder.GetCode() >= block[0] && recorder.GetCode() <= block[1] {
			log.Errorf("Caught HTTP Status Code %d, returning error page", recorder.GetCode())

			var query string
			if len(h.backendQuery) > 0 {
				query = "/" + strings.TrimPrefix(h.backendQuery, "/")
				query = strings.Replace(query, "{status}", strconv.Itoa(recorder.GetCode()), -1)
			}

			pageReq, err := newRequest(h.backendURL + query)
			if err != nil {
				log.Error(err)
				w.WriteHeader(recorder.GetCode())
				fmt.Fprint(w, http.StatusText(recorder.GetCode()))
				return
			}

			recorderErrorPage := newResponseRecorder(w)
			utils.CopyHeaders(pageReq.Header, req.Header)

			h.backendHandler.ServeHTTP(recorderErrorPage, pageReq.WithContext(req.Context()))

			utils.CopyHeaders(w.Header(), recorderErrorPage.Header())
			w.WriteHeader(recorder.GetCode())
			w.Write(recorderErrorPage.GetBody().Bytes())
			return
		}
	}

	// did not catch a configured status code so proceed with the request
	utils.CopyHeaders(w.Header(), recorder.Header())
	w.WriteHeader(recorder.GetCode())
	w.Write(recorder.GetBody().Bytes())
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
