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

// FIXME: add a note to the documentation saying that the midlleware should be placed as close to the entry point as possible,
//  so that every modification to the response will be recorded. What about the request?

const typeName = "Tap"

// Kinds of records sent to the tap services.
const (
	kindRequest  = "request"
	kindResponse = "response"
)

// Headers letting a tap service dispatch the records without parsing them.
const (
	headerRecordID   = "X-Tap-Id"
	headerRecordKind = "X-Tap-Record"
)

// errBodyTooLarge is returned when a body to hold in memory is over maxBodySize.
var errBodyTooLarge = errors.New("body too large")

type serviceBuilder interface {
	BuildHTTP(ctx context.Context, serviceName string) (http.Handler, error)
}

// record is the payload sent to the tap services.
type record struct {
	ID   string    `json:"id"`
	Kind string    `json:"kind"`
	Time time.Time `json:"time"`
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
	// Body is base64-encoded, as it may not be valid UTF-8.
	Body          []byte `json:"body,omitempty"`
	BodyTruncated bool   `json:"bodyTruncated,omitempty"`
}

type responseRecord struct {
	Status  int         `json:"status"`
	Headers http.Header `json:"headers,omitempty"`
	// Body is base64-encoded, as it may not be valid UTF-8.
	Body          []byte `json:"body,omitempty"`
	BodyTruncated bool   `json:"bodyTruncated,omitempty"`
	// Duration spans from the reception of the request to the end of the response, in nanoseconds.
	Duration time.Duration `json:"duration"`
}

type destination struct {
	handler        http.Handler
	path           string
	requestHeaders []string
	failClosed     bool
	timeout        time.Duration
	maxBodySize    int64

	recordBody        bool
	maxRecordBodySize int64
}

func newDestination(ctx context.Context, config dynamic.TapRequest, serviceBuilder serviceBuilder) (*destination, error) {
	if config.Service == "" {
		return nil, errors.New("service must be defined")
	}

	handler, err := serviceBuilder.BuildHTTP(ctx, config.Service)
	if err != nil {
		return nil, fmt.Errorf("building tap service handler: %w", err)
	}

	path := config.Path
	if path == "" {
		path = dynamic.TapDefaultPath
	}

	// The path ends up in the request line, where it must be absolute.
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return &destination{
		handler:           handler,
		path:              path,
		failClosed:        config.FailClosed,
		timeout:           time.Duration(config.Timeout),
		maxBodySize:       ptr.Deref(config.MaxBodySize, dynamic.TapDefaultMaxBodySize),
		recordBody:        config.RecordBody,
		maxRecordBodySize: ptr.Deref(config.MaxRecordBodySize, dynamic.TapDefaultMaxRecordBodySize),
	}, nil
}

type tap struct {
	name string
	next http.Handler

	request  *destination
	response *destination
}

// New creates a new tap middleware.
func New(ctx context.Context, next http.Handler, config dynamic.Tap, serviceBuilder serviceBuilder, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	if config.Request == nil && config.Response == nil {
		return nil, errors.New("at least one of request or response must be defined")
	}

	t := &tap{
		name: name,
		next: next,
	}

	var err error
	if config.Request != nil {
		if t.request, err = newDestination(ctx, *config.Request, serviceBuilder); err != nil {
			return nil, fmt.Errorf("building request destination: %w", err)
		}
	}

	if config.Response != nil {
		if t.response, err = newDestination(ctx, config.Response.TapRequest, serviceBuilder); err != nil {
			return nil, fmt.Errorf("building response destination: %w", err)
		}

		t.response.requestHeaders = config.Response.RequestHeaders
	}

	return t, nil
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
		t.reject(ctx, rw, kindRequest, http.StatusInternalServerError)
		return
	}

	var traceID string
	if spanContext := trace.SpanContextFromContext(ctx); spanContext.HasTraceID() {
		traceID = spanContext.TraceID().String()
	}

	if t.request != nil {
		reqRecord, err := newRequestRecord(req, t.request)
		if errors.Is(err, errBodyTooLarge) {
			logger.Error().Err(err).Msg("Unable to record the request")
			t.reject(ctx, rw, kindRequest, http.StatusRequestEntityTooLarge)
			return
		}

		if err != nil {
			logger.Error().Err(err).Msg("Unable to record the request")
			t.reject(ctx, rw, kindRequest, http.StatusInternalServerError)
			return
		}

		rec := &record{ID: id, Kind: kindRequest, Time: start, TraceID: traceID, Request: reqRecord}
		if err := t.send(ctx, t.request, req.Host, rec); err != nil {
			logger.Error().Err(err).Msg("Unable to send the request record")
			if t.request.failClosed {
				t.reject(ctx, rw, kindRequest, http.StatusInternalServerError)
				return
			}
		}
	}

	if t.response == nil {
		t.next.ServeHTTP(rw, req)
		return
	}

	// The request is described on arrival, before the next handlers alter it.
	var responseRecordRequest *requestRecord
	if t.response != nil && len(t.response.requestHeaders) > 0 {
		responseRecordRequest = describeRequest(req, t.response.requestHeaders)
	}

	capturer := newResponseCapturer(rw, t.response, t.response.failClosed)
	t.next.ServeHTTP(capturer, req)

	if capturer.bodyTooLarge {
		logger.Error().Err(errBodyTooLarge).Msg("Unable to record the response")
		t.reject(ctx, rw, kindResponse, http.StatusInternalServerError)
		return
	}

	rec := &record{ID: id, Kind: kindResponse, Time: start, TraceID: traceID, Request: responseRecordRequest, Response: capturer.record(time.Since(start))}

	if err := t.send(ctx, t.response, req.Host, rec); err != nil {
		logger.Error().Err(err).Msg("Unable to send the response record")
		if t.response.failClosed {
			t.reject(ctx, rw, kindResponse, http.StatusInternalServerError)
			return
		}
	}

	// A withheld response is served once its record is sent.
	if err := capturer.serve(); err != nil {
		logger.Debug().Err(err).Msg("Unable to serve the response")
	}
}

// send sends a record to a tap service, and returns an error unless it is accepted.
func (t *tap) send(ctx context.Context, dest *destination, host string, rec *record) error {
	payload, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshaling record: %w", err)
	}

	// The record is sent even when the client is gone, within the timeout.
	sendCtx := context.WithoutCancel(ctx)
	if dest.timeout > 0 {
		var cancel context.CancelFunc
		sendCtx, cancel = context.WithTimeout(sendCtx, dest.timeout)

		defer cancel()
	}

	// The tap service call must not alter the access log data of the request.
	sendCtx = context.WithValue(sendCtx, accesslog.DataTableKey, nil)

	sendReq, err := http.NewRequestWithContext(sendCtx, http.MethodPost, "http://"+host+dest.path, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating record request: %w", err)
	}

	sendReq.RequestURI = dest.path
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

func (t *tap) reject(ctx context.Context, rw http.ResponseWriter, kind string, status int) {
	observability.SetStatusErrorf(ctx, "Unable to record the %s", kind)
	http.Error(rw, http.StatusText(status), status)
}

// describeRequest describes the request with the given headers only, for a response record.
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

	if !dest.recordBody {
		return rec, nil
	}

	body, truncated, err := readBody(req, dest.maxRecordBodySize, dest.maxBodySize)
	if err != nil {
		return nil, fmt.Errorf("reading request body: %w", err)
	}

	rec.Body = body
	rec.BodyTruncated = truncated

	return rec, nil
}

// readBody reads at most maxRecordBodySize bytes of the body, and reports whether it was truncated.
// The body is replaced, so that the next handler still reads it whole. A body read whole is held
// in memory, so one larger than maxBodySize is rejected with errBodyTooLarge.
func readBody(req *http.Request, maxRecordBodySize, maxBodySize int64) ([]byte, bool, error) {
	if req.Body == nil || req.Body == http.NoBody || maxRecordBodySize == 0 {
		return nil, false, nil
	}

	// A body announced too large is rejected unread, so that a client waiting for a 100 Continue gets the rejection.
	if maxRecordBodySize < 0 && maxBodySize >= 0 && req.ContentLength > maxBodySize {
		return nil, false, errBodyTooLarge
	}

	// Reading the body answers an Expect: 100-continue, so it is not forwarded,
	// which would make the backend answer a second one.
	if strings.EqualFold(req.Header.Get("Expect"), "100-continue") {
		req.Header.Del("Expect")
	}

	if maxRecordBodySize < 0 {
		reader := io.Reader(req.Body)
		if maxBodySize >= 0 {
			// One byte more than the limit tells a body at the limit from a larger one.
			reader = io.LimitReader(req.Body, maxBodySize+1)
		}

		body, err := io.ReadAll(reader)
		if err != nil {
			return nil, false, fmt.Errorf("reading request body: %w", err)
		}

		if maxBodySize >= 0 && int64(len(body)) > maxBodySize {
			return nil, false, errBodyTooLarge
		}

		req.Body = replayBody(bytes.NewReader(body), req.Body)

		return body, false, nil
	}

	// One byte more than the limit tells a body at the limit from a truncated one.
	buf := make([]byte, maxRecordBodySize+1)
	n, err := io.ReadFull(req.Body, buf)
	switch {
	// The buffer is full, so the body is over the limit: the bytes read are put back in front of it.
	case err == nil:
		req.Body = replayBody(io.MultiReader(bytes.NewReader(buf[:n]), req.Body), req.Body)
		return buf[:maxRecordBodySize], true, nil

	// io.EOF happens with HTTP/3, where a bodyless request has a non-nil Body and no
	// Content-Length, and with a chunked request sending no chunks.
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
