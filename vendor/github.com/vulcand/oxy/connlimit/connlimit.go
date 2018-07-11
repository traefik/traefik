// Package connlimit provides control over simultaneous connections coming from the same source
package connlimit

import (
	"fmt"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/utils"
)

// ConnLimiter tracks concurrent connection per token
// and is capable of rejecting connections if they are failed
type ConnLimiter struct {
	mutex            *sync.Mutex
	extract          utils.SourceExtractor
	connections      map[string]int64
	maxConnections   int64
	totalConnections int64
	next             http.Handler

	errHandler utils.ErrorHandler
	log        *log.Logger
}

// New creates a new ConnLimiter
func New(next http.Handler, extract utils.SourceExtractor, maxConnections int64, options ...ConnLimitOption) (*ConnLimiter, error) {
	if extract == nil {
		return nil, fmt.Errorf("Extract function can not be nil")
	}
	cl := &ConnLimiter{
		mutex:          &sync.Mutex{},
		extract:        extract,
		maxConnections: maxConnections,
		connections:    make(map[string]int64),
		next:           next,
		log:            log.StandardLogger(),
	}

	for _, o := range options {
		if err := o(cl); err != nil {
			return nil, err
		}
	}
	if cl.errHandler == nil {
		cl.errHandler = &ConnErrHandler{
			log: cl.log,
		}
	}
	return cl, nil
}

// Logger defines the logger the connection limiter will use.
//
// It defaults to logrus.StandardLogger(), the global logger used by logrus.
func Logger(l *log.Logger) ConnLimitOption {
	return func(cl *ConnLimiter) error {
		cl.log = l
		return nil
	}
}

// Wrap sets the next handler to be called by connexion limiter handler.
func (cl *ConnLimiter) Wrap(h http.Handler) {
	cl.next = h
}

func (cl *ConnLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, amount, err := cl.extract.Extract(r)
	if err != nil {
		cl.log.Errorf("failed to extract source of the connection: %v", err)
		cl.errHandler.ServeHTTP(w, r, err)
		return
	}
	if err := cl.acquire(token, amount); err != nil {
		cl.log.Debugf("limiting request source %s: %v", token, err)
		cl.errHandler.ServeHTTP(w, r, err)
		return
	}

	defer cl.release(token, amount)

	cl.next.ServeHTTP(w, r)
}

func (cl *ConnLimiter) acquire(token string, amount int64) error {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	connections := cl.connections[token]
	if connections >= cl.maxConnections {
		return &MaxConnError{max: cl.maxConnections}
	}

	cl.connections[token] += amount
	cl.totalConnections += amount
	return nil
}

func (cl *ConnLimiter) release(token string, amount int64) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	cl.connections[token] -= amount
	cl.totalConnections -= amount

	// Otherwise it would grow forever
	if cl.connections[token] == 0 {
		delete(cl.connections, token)
	}
}

// MaxConnError maximum connections reached error
type MaxConnError struct {
	max int64
}

func (m *MaxConnError) Error() string {
	return fmt.Sprintf("max connections reached: %d", m.max)
}

// ConnErrHandler connection limiter error handler
type ConnErrHandler struct {
	log *log.Logger
}

func (e *ConnErrHandler) ServeHTTP(w http.ResponseWriter, req *http.Request, err error) {
	if e.log.Level >= log.DebugLevel {
		logEntry := e.log.WithField("Request", utils.DumpHttpRequest(req))
		logEntry.Debug("vulcand/oxy/connlimit: begin ServeHttp on request")
		defer logEntry.Debug("vulcand/oxy/connlimit: completed ServeHttp on request")
	}

	if _, ok := err.(*MaxConnError); ok {
		w.WriteHeader(429)
		w.Write([]byte(err.Error()))
		return
	}
	utils.DefaultHandler.ServeHTTP(w, req, err)
}

// ConnLimitOption connection limit option type
type ConnLimitOption func(l *ConnLimiter) error

// ErrorHandler sets error handler of the server
func ErrorHandler(h utils.ErrorHandler) ConnLimitOption {
	return func(cl *ConnLimiter) error {
		cl.errHandler = h
		return nil
	}
}
