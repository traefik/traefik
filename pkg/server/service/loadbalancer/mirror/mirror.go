package mirror

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/healthcheck"
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslog"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/safe"
)

// Mirroring is an http.Handler that can mirror requests.
type Mirroring struct {
	handler        http.Handler
	mirrorHandlers []*mirrorHandler
	rw             http.ResponseWriter
	routinePool    *safe.Pool

	mirrorBody       bool
	maxBodySize      int64
	wantsHealthCheck bool

	// streaming sends the mirrored requests along with the main one, and streams the body to them while the main handler reads it.
	streaming bool

	lock  sync.RWMutex
	total uint64
}

// New returns a new instance of *Mirroring.
// When streaming, the request body is streamed to the mirrors instead of being buffered before calling the main handler,
// and maxBodySize is ignored.
// A mirror reading the body slower than the main handler is then aborted, so that it never slows the main request down.
func New(handler http.Handler, pool *safe.Pool, mirrorBody bool, maxBodySize int64, streaming bool, hc *dynamic.HealthCheck) *Mirroring {
	return &Mirroring{
		routinePool:      pool,
		handler:          handler,
		rw:               blackHoleResponseWriter{},
		mirrorBody:       mirrorBody,
		maxBodySize:      maxBodySize,
		wantsHealthCheck: hc != nil,
		// Without a mirrored body, there is nothing to stream.
		streaming: streaming && mirrorBody,
	}
}

type mirrorHandler struct {
	http.Handler

	numerator   uint64
	denominator uint64

	lock  sync.RWMutex
	count uint64
}

type clonableRequest interface {
	Clone(ctx context.Context) *http.Request
}

func (m *Mirroring) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	mirrors := m.getActiveMirrors()
	if len(mirrors) == 0 {
		m.handler.ServeHTTP(rw, req)
		return
	}

	if m.streaming {
		m.serveStreaming(rw, req, mirrors)
		return
	}

	logger := log.Ctx(req.Context())

	var rr clonableRequest = req

	if m.mirrorBody {
		var err error
		var bytesRead []byte
		rr, bytesRead, err = NewReusableRequest(req, m.maxBodySize)
		if err != nil && !errors.Is(err, ErrBodyTooLarge) {
			logger.Debug().Err(err).Msg("Error while creating reusable request for mirroring")
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if errors.Is(err, ErrBodyTooLarge) {
			req.Body = io.NopCloser(io.MultiReader(bytes.NewReader(bytesRead), req.Body))
			m.handler.ServeHTTP(rw, req)
			logger.Debug().Msg("No mirroring, request body larger than allowed size")
			return
		}
	}

	m.handler.ServeHTTP(rw, rr.Clone(req.Context()))

	select {
	case <-req.Context().Done():
		// No mirroring if request has been canceled during main handler ServeHTTP
		logger.Warn().Msg("No mirroring, request has been canceled during main handler ServeHTTP")
		return
	default:
	}

	m.routinePool.GoCtx(func(_ context.Context) {
		for _, handler := range mirrors {
			// prepare request, update body from buffer
			r := rr.Clone(req.Context())

			handler.ServeHTTP(m.rw, r.WithContext(mirrorContext(r.Context())))
		}
	})
}

// mirrorContext returns the context of a mirrored request.
func mirrorContext(ctx context.Context) context.Context {
	// In ServeHTTP, we rely on the presence of the accessLog datatable found in the request's context
	// to know whether we should mutate said datatable (and contribute some fields to the log).
	// In this instance, we do not want the mirrors mutating (i.e. changing the service name in)
	// the logs related to the mirrored server.
	// Especially since it would result in unguarded concurrent reads/writes on the datatable.
	// Therefore, we reset any potential datatable key in the new context that we pass around.
	ctx = context.WithValue(ctx, accesslog.DataTableKey, nil)

	// For the same reason, the mirrored services must not publish their metadata
	// into the state container seeded for the mirrored request, as the access log
	// middleware still reads it to emit the log of the main handler.
	// Therefore, we seed a new, throwaway container for the mirrors.
	ctx = observability.WithServiceObservabilityState(ctx)

	// When a request served by m.handler is successful, req.Context will be canceled,
	// which would trigger a cancellation of the ongoing mirrored requests.
	// Therefore, we give a new, non-cancellable context  to each of the mirrored calls,
	// so they can terminate by themselves.
	return contextStopPropagation{ctx}
}

// AddMirror adds an httpHandler to mirror numerator out of denominator requests to.
func (m *Mirroring) AddMirror(handler http.Handler, numerator, denominator int64) error {
	if denominator <= 0 || numerator < 0 || numerator > denominator {
		return fmt.Errorf("invalid fraction %d/%d: the denominator must be positive, and the numerator between 0 and the denominator", numerator, denominator)
	}
	m.mirrorHandlers = append(m.mirrorHandlers, &mirrorHandler{Handler: handler, numerator: uint64(numerator), denominator: uint64(denominator)})
	return nil
}

// RegisterStatusUpdater adds fn to the list of hooks that are run when the
// status of handler of the Mirroring changes.
// Not thread safe.
func (m *Mirroring) RegisterStatusUpdater(fn func(up bool)) error {
	// Since the status propagation is completely transparent through the
	// mirroring (because of the recursion on the underlying service), we could maybe
	// skip that below, and even not add HealthCheck as a field of
	// dynamic.Mirroring. But I think it's easier to understand for the user
	// if the HealthCheck is required absolutely everywhere in the config.
	if !m.wantsHealthCheck {
		return errors.New("healthCheck not enabled in config for this mirroring service")
	}

	updater, ok := m.handler.(healthcheck.StatusUpdater)
	if !ok {
		return fmt.Errorf("service of mirroring %T not a healthcheck.StatusUpdater", m.handler)
	}

	if err := updater.RegisterStatusUpdater(fn); err != nil {
		return fmt.Errorf("cannot register service of mirroring as updater: %w", err)
	}

	return nil
}

// serveStreaming sends the mirrored requests before calling the main handler,
// and streams them the request body while the main handler reads it.
func (m *Mirroring) serveStreaming(rw http.ResponseWriter, req *http.Request, mirrors []http.Handler) {
	hasBody := req.Body != nil && req.Body != http.NoBody

	var bodies []*streamedBody
	for _, handler := range mirrors {
		// The mirrored requests are cloned before calling the main handler,
		// so that the handlers it chains to do not alter them.
		mirrorReq := req.Clone(mirrorContext(req.Context()))

		// The body is streamed while the main handler reads it,
		// so the mirrored request must not wait for an interim response before sending it.
		mirrorReq.Header.Del("Expect")

		if hasBody {
			body := newStreamedBody()
			mirrorReq.Body = body
			mirrorReq.GetBody = nil
			bodies = append(bodies, body)
		}

		m.routinePool.GoCtx(func(_ context.Context) {
			handler.ServeHTTP(m.rw, mirrorReq)
		})
	}

	if len(bodies) > 0 {
		req.Body = &teeBody{ReadCloser: req.Body, mirrors: bodies}

		// When the main handler stops reading the body before its end,
		// the mirrored requests are aborted instead of being sent with a truncated body.
		defer func() {
			for _, body := range bodies {
				body.closeWithError(errBodyNotFullyRead)
			}
		}()
	}

	m.handler.ServeHTTP(rw, req)
}

func (m *Mirroring) inc() uint64 {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.total++
	return m.total
}

func (m *Mirroring) getActiveMirrors() []http.Handler {
	total := m.inc()

	var mirrors []http.Handler
	for _, handler := range m.mirrorHandlers {
		handler.lock.Lock()
		if handler.count*handler.denominator < total*handler.numerator {
			handler.count++
			handler.lock.Unlock()
			mirrors = append(mirrors, handler)
		} else {
			handler.lock.Unlock()
		}
	}
	return mirrors
}

type blackHoleResponseWriter struct{}

func (b blackHoleResponseWriter) Flush() {}

func (b blackHoleResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("connection on blackHoleResponseWriter cannot be hijacked")
}

func (b blackHoleResponseWriter) Header() http.Header {
	return http.Header{}
}

func (b blackHoleResponseWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (b blackHoleResponseWriter) WriteHeader(_ int) {}

type contextStopPropagation struct {
	context.Context
}

func (c contextStopPropagation) Done() <-chan struct{} {
	return make(chan struct{})
}

// ReusableRequest keeps in memory the body of the given request,
// so that the request can be fully cloned by each mirror.
type ReusableRequest struct {
	req  *http.Request
	body []byte
}

var ErrBodyTooLarge = errors.New("request body too large")

// NewReusableRequest returns a new reusable request. If the returned error is ErrBodyTooLarge, NewReusableRequest also returns the
// bytes already consumed from the request's body.
func NewReusableRequest(req *http.Request, maxBodySize int64) (*ReusableRequest, []byte, error) {
	if req == nil {
		return nil, nil, errors.New("nil input request")
	}
	if req.Body == nil {
		return &ReusableRequest{req: req}, nil, nil
	}

	// unbounded body size
	if maxBodySize < 0 {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, nil, err
		}
		return &ReusableRequest{
			req:  req,
			body: body,
		}, nil, nil
	}

	// we purposefully try to read _more_ than maxBodySize to detect whether
	// the request body is larger than what we allow for the mirrors.
	body := make([]byte, maxBodySize+1)
	n, err := io.ReadFull(req.Body, body)
	// This happens with HTTP/3, where the end of the body is framed at the stream
	// level rather than declared via Content-Length, so a bodyless request arrives
	// with a non-nil Body and ContentLength == -1. The same shape is possible for a
	// chunked request that sends no chunks.
	if errors.Is(err, io.EOF) {
		return &ReusableRequest{
			req:  req,
			body: body[:n],
		}, nil, nil
	}
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, nil, err
	}

	// we got an ErrUnexpectedEOF, which means there was less than maxBodySize data to read,
	// which permits us sending also to all the mirrors later.
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return &ReusableRequest{
			req:  req,
			body: body[:n],
		}, nil, nil
	}

	// err == nil , which means data size > maxBodySize
	return nil, body[:n], ErrBodyTooLarge
}

// Clone clones the request.
func (rr ReusableRequest) Clone(ctx context.Context) *http.Request {
	req := rr.req.Clone(ctx)

	if rr.body != nil {
		req.Body = io.NopCloser(bytes.NewReader(rr.body))
	}

	return req
}
