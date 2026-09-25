// Package tap provides a middleware that sends a copy of the requests and responses
// going through it to Traefik services.
package tap

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslog"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/utils/ptr"
)

const typeName = "Tap"

// Kinds of records sent to the tap services.
const (
	kindRequest  = "request"
	kindResponse = "response"
)

// Headers set on the requests sent to the tap services,
// so that a service serving several kinds of records can dispatch them without parsing the payload.
const (
	headerRecordID   = "X-Tap-Id"
	headerRecordKind = "X-Tap-Record"
)

type serviceBuilder interface {
	BuildHTTP(ctx context.Context, serviceName string) (http.Handler, error)
}

// record is the payload sent to the tap services.
type record struct {
	ID   string    `json:"id"`
	Kind string    `json:"kind"`
	Time time.Time `json:"time"`
	// TraceID is the trace the recorded request belongs to, when tracing is enabled.
	TraceID  string          `json:"traceId,omitempty"`
	Request  *requestRecord  `json:"request,omitempty"`
	Response *responseRecord `json:"response,omitempty"`
}

type requestRecord struct {
	Method     string      `json:"method"`
	URL        string      `json:"url"`
	Host       string      `json:"host"`
	Proto      string      `json:"proto"`
	RemoteAddr string      `json:"remoteAddr,omitempty"`
	Headers    http.Header `json:"headers,omitempty"`
	// Body is base64-encoded, as it is not necessarily valid UTF-8.
	Body          []byte `json:"body,omitempty"`
	BodyTruncated bool   `json:"bodyTruncated,omitempty"`
}

type responseRecord struct {
	Status  int         `json:"status"`
	Headers http.Header `json:"headers,omitempty"`
	// Body is base64-encoded, as it is not necessarily valid UTF-8.
	Body          []byte `json:"body,omitempty"`
	BodyTruncated bool   `json:"bodyTruncated,omitempty"`
	// Duration is the time elapsed between the reception of the request and the end of the response, in nanoseconds.
	Duration time.Duration `json:"duration"`
}

type destination struct {
	handler        http.Handler
	path           string
	body           bool
	maxBodySize    int64
	requestHeaders []string
	failClosed     bool
}

type tap struct {
	name string
	next http.Handler

	request  *destination
	response *destination

	timeout time.Duration
}

// New creates a new tap middleware.
func New(ctx context.Context, next http.Handler, config dynamic.Tap, serviceBuilder serviceBuilder, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	if config.Request == nil && config.Response == nil {
		return nil, errors.New("at least one of request or response must be defined")
	}

	if config.Request != nil && len(config.Request.RequestHeaders) > 0 {
		return nil, errors.New("requestHeaders is only allowed on the response records")
	}

	t := &tap{
		name:    name,
		next:    next,
		timeout: time.Duration(config.Timeout),
	}

	var err error
	if t.request, err = newDestination(ctx, config.Request, serviceBuilder); err != nil {
		return nil, fmt.Errorf("building request destination: %w", err)
	}

	if t.response, err = newDestination(ctx, config.Response, serviceBuilder); err != nil {
		return nil, fmt.Errorf("building response destination: %w", err)
	}

	return t, nil
}

func newDestination(ctx context.Context, config *dynamic.TapRecord, serviceBuilder serviceBuilder) (*destination, error) {
	if config == nil {
		return nil, nil
	}

	if config.Service == "" {
		return nil, errors.New("service must be defined")
	}

	handler, err := serviceBuilder.BuildHTTP(ctx, config.Service)
	if err != nil {
		return nil, err
	}

	path := config.Path
	if path == "" {
		path = dynamic.TapDefaultPath
	}

	// The path ends up in the request line sent to the service, where it must be absolute.
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	maxBodySize := dynamic.TapDefaultMaxBodySize
	if config.MaxBodySize != nil {
		maxBodySize = *config.MaxBodySize
	}

	return &destination{
		handler:        handler,
		path:           path,
		body:           ptr.Deref(config.Body, true),
		maxBodySize:    maxBodySize,
		requestHeaders: config.RequestHeaders,
		failClosed:     config.FailClosed,
	}, nil
}

func (t *tap) GetTracingInformation() (string, string) {
	return t.name, typeName
}

func (t *tap) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), t.name, typeName)
	ctx := logger.WithContext(req.Context())

	start := time.Now()

	id, err := newRecordID()
	if err != nil {
		logger.Error().Err(err).Msg("Unable to generate a record ID")
		t.reject(ctx, rw, kindRequest)
		return
	}

	var traceID string
	if spanContext := trace.SpanContextFromContext(ctx); spanContext.HasTraceID() {
		traceID = spanContext.TraceID().String()
	}

	// The request is described on arrival, before the next handlers alter it.
	var described *requestRecord
	if t.response != nil && len(t.response.requestHeaders) > 0 {
		described = describeRequest(req, t.response.requestHeaders)
	}

	if t.request != nil {
		reqRecord, err := newRequestRecord(req, t.request)
		if err != nil {
			logger.Error().Err(err).Msg("Unable to record the request")
			t.reject(ctx, rw, kindRequest)
			return
		}

		rec := &record{ID: id, Kind: kindRequest, Time: start, TraceID: traceID, Request: reqRecord}
		if err := t.send(ctx, t.request, req.Host, rec); err != nil {
			logger.Error().Err(err).Msg("Unable to send the request record")
			if t.request.failClosed {
				t.reject(ctx, rw, kindRequest)
				return
			}
		}
	}

	if t.response == nil {
		t.next.ServeHTTP(rw, req)
		return
	}

	capturer := newResponseCapturer(rw, t.response, t.response.failClosed)
	t.next.ServeHTTP(capturer, req)

	rec := &record{ID: id, Kind: kindResponse, Time: start, TraceID: traceID, Request: described, Response: capturer.record(time.Since(start))}

	if err := t.send(ctx, t.response, req.Host, rec); err != nil {
		logger.Error().Err(err).Msg("Unable to send the response record")
		if t.response.failClosed && !capturer.served() {
			t.reject(ctx, rw, kindResponse)
			return
		}
	}

	// When failing closed the response is withheld until the record is sent.
	if err := capturer.serve(); err != nil {
		logger.Debug().Err(err).Msg("Unable to serve the response")
	}
}

// send sends a record to a tap service, and reports whether it has been accepted.
func (t *tap) send(ctx context.Context, dest *destination, host string, rec *record) error {
	payload, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshaling record: %w", err)
	}

	// A record must be sent even when the client is gone, and must not outlive the configured timeout.
	sendCtx, cancel := context.WithoutCancel(ctx), context.CancelFunc(func() {})
	if t.timeout > 0 {
		sendCtx, cancel = context.WithTimeout(sendCtx, t.timeout)
	}
	defer cancel()

	// The access log data table of the incoming request must not be mutated by the tap service call.
	sendCtx = context.WithValue(sendCtx, accesslog.DataTableKey, nil)

	sendReq, err := http.NewRequestWithContext(sendCtx, http.MethodPost, "http://"+host+dest.path, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating record request: %w", err)
	}

	sendReq.RequestURI = dest.path
	sendReq.ContentLength = int64(len(payload))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set(headerRecordID, rec.ID)
	sendReq.Header.Set(headerRecordKind, rec.Kind)

	recorder := &statusRecorder{header: make(http.Header)}
	dest.handler.ServeHTTP(recorder, sendReq)

	if recorder.status >= http.StatusBadRequest {
		return fmt.Errorf("tap service returned %d", recorder.status)
	}

	return nil
}

func (t *tap) reject(ctx context.Context, rw http.ResponseWriter, kind string) {
	observability.SetStatusErrorf(ctx, "Unable to record the %s", kind)
	http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// describeRequest describes the request with the given headers only, for a record that is not the
// request one.
func describeRequest(req *http.Request, headerNames []string) *requestRecord {
	headers := make(http.Header, len(headerNames))
	for _, name := range headerNames {
		for _, value := range req.Header.Values(name) {
			headers.Add(name, value)
		}
	}

	return &requestRecord{
		Method:     req.Method,
		URL:        req.URL.RequestURI(),
		Host:       req.Host,
		Proto:      req.Proto,
		RemoteAddr: req.RemoteAddr,
		Headers:    headers,
	}
}

func newRequestRecord(req *http.Request, dest *destination) (*requestRecord, error) {
	rec := &requestRecord{
		Method:     req.Method,
		URL:        req.URL.RequestURI(),
		Host:       req.Host,
		Proto:      req.Proto,
		RemoteAddr: req.RemoteAddr,
		Headers:    req.Header.Clone(),
	}

	if !dest.body {
		return rec, nil
	}

	body, truncated, err := readBody(req, dest.maxBodySize)
	if err != nil {
		return nil, fmt.Errorf("reading request body: %w", err)
	}

	rec.Body = body
	rec.BodyTruncated = truncated

	return rec, nil
}

// readBody reads at most maxBodySize bytes from the request body, and reports whether it has been truncated.
// The request body is replaced by an equivalent one, so that the next handler still reads it whole.
func readBody(req *http.Request, maxBodySize int64) ([]byte, bool, error) {
	if req.Body == nil || req.Body == http.NoBody || maxBodySize == 0 {
		return nil, false, nil
	}

	// Reading the body is what makes the server answer an Expect: 100-continue, so the
	// expectation is satisfied here and must not be forwarded: the backend would answer a
	// second informational response, which the client has no reason to see.
	if strings.EqualFold(req.Header.Get("Expect"), "100-continue") {
		req.Header.Del("Expect")
	}

	if maxBodySize < 0 {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, false, err
		}

		req.Body = replayBody(bytes.NewReader(body), req.Body)

		return body, false, nil
	}

	// One byte more than the limit is read, to tell a body at the limit from a truncated one.
	buf := make([]byte, maxBodySize+1)
	n, err := io.ReadFull(req.Body, buf)
	switch {
	// The whole buffer has been filled, which means the body is larger than the limit.
	// The bytes already read are put back in front of the body, so that the next handler still reads it whole.
	case err == nil:
		req.Body = replayBody(io.MultiReader(bytes.NewReader(buf[:n]), req.Body), req.Body)
		return buf[:maxBodySize], true, nil

	// io.EOF happens with HTTP/3, where the end of the body is framed at the stream
	// level rather than declared via Content-Length, so a bodyless request arrives
	// with a non-nil Body and ContentLength == -1. The same shape is possible for a
	// chunked request that sends no chunks.
	case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		req.Body = replayBody(bytes.NewReader(buf[:n]), req.Body)
		return buf[:n], false, nil

	default:
		return nil, false, err
	}
}

// replayBody returns a body reading from reader, while still closing the original body.
func replayBody(reader io.Reader, original io.Closer) io.ReadCloser {
	return struct {
		io.Reader
		io.Closer
	}{Reader: reader, Closer: original}
}

func newRecordID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf[:]), nil
}
