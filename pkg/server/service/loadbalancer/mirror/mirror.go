package mirror

import (
	"bufio"
	"context"
	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/containous/traefik/v2/pkg/middlewares/accesslog"
	"github.com/containous/traefik/v2/pkg/safe"
)

// Mirroring is an http.Handler that can mirror requests.
type Mirroring struct {
	handler        http.Handler
	mirrorHandlers []*mirrorHandler
	rw             http.ResponseWriter
	routinePool    *safe.Pool

	lock  sync.RWMutex
	total uint64
}

// New returns a new instance of *Mirroring.
func New(handler http.Handler, pool *safe.Pool) *Mirroring {
	return &Mirroring{
		routinePool: pool,
		handler:     handler,
		rw:          blackholeResponseWriter{},
	}
}

func (m *Mirroring) inc() uint64 {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.total++
	return m.total
}

type mirrorHandler struct {
	http.Handler
	percent int

	lock  sync.RWMutex
	count uint64
}

func (m *Mirroring) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.handler.ServeHTTP(rw, req)

	select {
	case <-req.Context().Done():
		// No mirroring if request has been canceled during main handler ServeHTTP
		return
	default:
	}

	m.routinePool.GoCtx(func(_ context.Context) {
		total := m.inc()
		for _, handler := range m.mirrorHandlers {
			handler.lock.Lock()
			if handler.count*100 < total*uint64(handler.percent) {
				handler.count++
				handler.lock.Unlock()

				// In ServeHTTP, we rely on the presence of the accesslog datatable found in the
				// request's context to know whether we should mutate said datatable (and
				// contribute some fields to the log). In this instance, we do not want the mirrors
				// mutating (i.e. changing the service name in) the logs related to the mirrored
				// server. Especially since it would result in unguarded concurrent reads/writes on
				// the datatable. Therefore, we reset any potential datatable key in the new
				// context that we pass around.
				ctx := context.WithValue(req.Context(), accesslog.DataTableKey, nil)

				// When a request served by m.handler is successful, req.Context will be canceled,
				// which would trigger a cancellation of the ongoing mirrored requests.
				// Therefore, we give a new, non-cancellable context  to each of the mirrored calls,
				// so they can terminate by themselves.
				handler.ServeHTTP(m.rw, req.WithContext(contextStopPropagation{ctx}))
			} else {
				handler.lock.Unlock()
			}
		}
	})
}

// AddMirror adds an httpHandler to mirror to.
func (m *Mirroring) AddMirror(handler http.Handler, percent int) error {
	if percent < 0 || percent > 100 {
		return errors.New("percent must be between 0 and 100")
	}
	m.mirrorHandlers = append(m.mirrorHandlers, &mirrorHandler{Handler: handler, percent: percent})
	return nil
}

type blackholeResponseWriter struct{}

func (b blackholeResponseWriter) Flush() {}

func (b blackholeResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("connection on blackholeResponseWriter cannot be hijacked")
}

func (b blackholeResponseWriter) Header() http.Header {
	return http.Header{}
}

func (b blackholeResponseWriter) Write(bytes []byte) (int, error) {
	return len(bytes), nil
}

func (b blackholeResponseWriter) WriteHeader(statusCode int) {}

type contextStopPropagation struct {
	context.Context
}

func (c contextStopPropagation) Done() <-chan struct{} {
	return make(chan struct{})
}
