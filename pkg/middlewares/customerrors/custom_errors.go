package customerrors

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/middlewares"
	"github.com/containous/traefik/pkg/tracing"
	"github.com/containous/traefik/pkg/types"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/utils"
)

// Compile time validation that the response recorder implements http interfaces correctly.
var _ middlewares.Stateful = &responseRecorderWithCloseNotify{}

const (
	typeName   = "customError"
	backendURL = "http://0.0.0.0"
)

type serviceBuilder interface {
	BuildHTTP(ctx context.Context, serviceName string, responseModifier func(*http.Response) error) (http.Handler, error)
}

// customErrors is a middleware that provides the custom error pages..
type customErrors struct {
	name           string
	next           http.Handler
	backendHandler http.Handler
	httpCodeRanges types.HTTPCodeRanges
	backendQuery   string
}

// New creates a new custom error pages middleware.
func New(ctx context.Context, next http.Handler, config config.ErrorPage, serviceBuilder serviceBuilder, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug("Creating middleware")

	httpCodeRanges, err := types.NewHTTPCodeRanges(config.Status)
	if err != nil {
		return nil, err
	}

	backend, err := serviceBuilder.BuildHTTP(ctx, config.Service, nil)
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
	logger := middlewares.GetLogger(req.Context(), c.name, typeName)

	if c.backendHandler == nil {
		logger.Error("Error pages: no backend handler.")
		tracing.SetErrorWithEvent(req, "Error pages: no backend handler.")
		c.next.ServeHTTP(rw, req)
		return
	}

	recorder := newResponseRecorder(rw, middlewares.GetLogger(context.Background(), "test", typeName))
	c.next.ServeHTTP(recorder, req)

	// check the recorder code against the configured http status code ranges
	for _, block := range c.httpCodeRanges {
		if recorder.GetCode() >= block[0] && recorder.GetCode() <= block[1] {
			logger.Errorf("Caught HTTP Status Code %d, returning error page", recorder.GetCode())

			var query string
			if len(c.backendQuery) > 0 {
				query = "/" + strings.TrimPrefix(c.backendQuery, "/")
				query = strings.Replace(query, "{status}", strconv.Itoa(recorder.GetCode()), -1)
			}

			pageReq, err := newRequest(backendURL + query)
			if err != nil {
				logger.Error(err)
				rw.WriteHeader(recorder.GetCode())
				_, err = fmt.Fprint(rw, http.StatusText(recorder.GetCode()))
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			recorderErrorPage := newResponseRecorder(rw, middlewares.GetLogger(context.Background(), "test", typeName))
			utils.CopyHeaders(pageReq.Header, req.Header)

			c.backendHandler.ServeHTTP(recorderErrorPage, pageReq.WithContext(req.Context()))

			utils.CopyHeaders(rw.Header(), recorderErrorPage.Header())
			rw.WriteHeader(recorder.GetCode())

			if _, err = rw.Write(recorderErrorPage.GetBody().Bytes()); err != nil {
				logger.Error(err)
			}
			return
		}
	}

	// did not catch a configured status code so proceed with the request
	utils.CopyHeaders(rw.Header(), recorder.Header())
	rw.WriteHeader(recorder.GetCode())
	_, err := rw.Write(recorder.GetBody().Bytes())
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func newRequest(baseURL string) (*http.Request, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error pages: error when parse URL: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
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
func newResponseRecorder(rw http.ResponseWriter, logger logrus.FieldLogger) responseRecorder {
	recorder := &responseRecorderWithoutCloseNotify{
		HeaderMap:      make(http.Header),
		Body:           new(bytes.Buffer),
		Code:           http.StatusOK,
		responseWriter: rw,
		logger:         logger,
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
	logger                   logrus.FieldLogger
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
		r.logger.Errorf("Error writing response in responseRecorder: %v", err)
		r.err = err
	}
	r.Body.Reset()

	if flusher, ok := r.responseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
