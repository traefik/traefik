/*
Package manners provides a wrapper for a standard net/http server that
ensures all active HTTP client have completed their current request
before the server shuts down.

It can be used a drop-in replacement for the standard http package,
or can wrap a pre-configured Server.

eg.

	http.Handle("/hello", func(w http.ResponseWriter, r *http.Request) {
	  w.Write([]byte("Hello\n"))
	})

	log.Fatal(manners.ListenAndServe(":8080", nil))

or for a customized server:

	s := manners.NewWithServer(&http.Server{
		Addr:           ":8080",
		Handler:        myHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	})
	log.Fatal(s.ListenAndServe())

The server will shut down cleanly when the Close() method is called:

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt, os.Kill)
		<-sigchan
		log.Info("Shutting down...")
		manners.Close()
	}()

	http.Handle("/hello", myHandler)
	log.Fatal(manners.ListenAndServe(":8080", nil))
*/
package manners

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"unsafe"
)

// StateHandler can be called by the server if the state of the connection changes.
// Notice that it passed previous state and the new state as parameters.
type StateHandler func(net.Conn, http.ConnState, http.ConnState)

// Options used by NewWithOptions to provide parameters for a GracefulServer
// instance to be created.
type Options struct {
	Server       *http.Server
	StateHandler StateHandler
	Listener     net.Listener
}

// A GracefulServer maintains a WaitGroup that counts how many in-flight
// requests the server is handling. When it receives a shutdown signal,
// it stops accepting new requests but does not actually shut down until
// all in-flight requests terminate.
//
// GracefulServer embeds the underlying net/http.Server making its non-override
// methods and properties avaiable.
//
// It must be initialized by calling NewServer or NewWithServer
type GracefulServer struct {
	*http.Server
	shutdown         chan bool
	shutdownFinished chan bool
	wg               waitGroup
	routinesCount    int32
	listener         *GracefulListener
	stateHandler     StateHandler
	connToProps      map[net.Conn]connProperties
	connToPropsMtx   sync.RWMutex

	up chan net.Listener // Only used by test code.
}

// NewServer creates a new GracefulServer.
func NewServer() *GracefulServer {
	return NewWithOptions(Options{})
}

// NewWithServer wraps an existing http.Server object and returns a
// GracefulServer that supports all of the original Server operations.
func NewWithServer(s *http.Server) *GracefulServer {
	return NewWithOptions(Options{Server: s})
}

// NewWithOptions creates a GracefulServer instance with the specified options.
func NewWithOptions(o Options) *GracefulServer {
	var listener *GracefulListener
	if o.Listener != nil {
		g, ok := o.Listener.(*GracefulListener)
		if !ok {
			listener = NewListener(o.Listener)
		} else {
			listener = g
		}
	}
	if o.Server == nil {
		o.Server = new(http.Server)
	}
	return &GracefulServer{
		Server:           o.Server,
		listener:         listener,
		stateHandler:     o.StateHandler,
		shutdown:         make(chan bool),
		shutdownFinished: make(chan bool, 1),
		wg:               new(sync.WaitGroup),
		connToProps:      make(map[net.Conn]connProperties),
	}
}

// Close stops the server from accepting new requets and begins shutting down.
// It returns true if it's the first time Close is called.
func (gs *GracefulServer) Close() bool {
	return <-gs.shutdown
}

// BlockingClose is similar to Close, except that it blocks until the last
// connection has been closed.
func (gs *GracefulServer) BlockingClose() bool {
	result := gs.Close()
	<-gs.shutdownFinished
	return result
}

// ListenAndServe provides a graceful equivalent of net/http.Serve.ListenAndServe.
func (gs *GracefulServer) ListenAndServe() error {
	if gs.listener == nil {
		oldListener, err := net.Listen("tcp", gs.Addr)
		if err != nil {
			return err
		}
		gs.listener = NewListener(oldListener.(*net.TCPListener))
	}
	return gs.Serve(gs.listener)
}

// ListenAndServeTLS provides a graceful equivalent of net/http.Serve.ListenAndServeTLS.
func (gs *GracefulServer) ListenAndServeTLS(certFile, keyFile string) error {
	// direct lift from net/http/server.go
	addr := gs.Addr
	if addr == "" {
		addr = ":https"
	}
	config := &tls.Config{}
	if gs.TLSConfig != nil {
		*config = *gs.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	return gs.ListenAndServeTLSWithConfig(config)
}

// ListenAndServeTLS provides a graceful equivalent of net/http.Serve.ListenAndServeTLS.
func (gs *GracefulServer) ListenAndServeTLSWithConfig(config *tls.Config) error {
	addr := gs.Addr
	if addr == "" {
		addr = ":https"
	}

	if gs.listener == nil {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}

		tlsListener := NewTLSListener(TCPKeepAliveListener{ln.(*net.TCPListener)}, config)
		gs.listener = NewListener(tlsListener)
	}
	return gs.Serve(gs.listener)
}

func (gs *GracefulServer) GetFile() (*os.File, error) {
	return gs.listener.GetFile()
}

func (gs *GracefulServer) HijackListener(s *http.Server, config *tls.Config) (*GracefulServer, error) {
	listenerUnsafePtr := unsafe.Pointer(gs.listener)
	listener, err := (*GracefulListener)(atomic.LoadPointer(&listenerUnsafePtr)).Clone()
	if err != nil {
		return nil, err
	}

	if config != nil {
		listener = NewTLSListener(TCPKeepAliveListener{listener.(*net.TCPListener)}, config)
	}

	other := NewWithServer(s)
	other.listener = NewListener(listener)
	return other, nil
}

// Serve provides a graceful equivalent net/http.Server.Serve.
//
// If listener is not an instance of *GracefulListener it will be wrapped
// to become one.
func (gs *GracefulServer) Serve(listener net.Listener) error {
	// Accept a net.Listener to preserve the interface compatibility with the
	// standard http.Server. If it is not a GracefulListener then wrap it into
	// one.
	gracefulListener, ok := listener.(*GracefulListener)
	if !ok {
		gracefulListener = NewListener(listener)
		listener = gracefulListener
	}
	listenerUnsafePtr := unsafe.Pointer(gs.listener)
	atomic.StorePointer(&listenerUnsafePtr, unsafe.Pointer(gracefulListener))

	// Wrap the server HTTP handler into graceful one, that will close kept
	// alive connections if a new request is received after shutdown.
	gracefulHandler := newGracefulHandler(gs.Server.Handler)
	gs.Server.Handler = gracefulHandler

	// Start a goroutine that waits for a shutdown signal and will stop the
	// listener when it receives the signal. That in turn will result in
	// unblocking of the http.Serve call.
	go func() {
		gs.shutdown <- true
		close(gs.shutdown)
		gracefulHandler.Close()
		gs.Server.SetKeepAlivesEnabled(false)
		gracefulListener.Close()
	}()

	originalConnState := gs.Server.ConnState

	// s.ConnState is invoked by the net/http.Server every time a connection
	// changes state. It keeps track of each connection's state over time,
	// enabling manners to handle persisted connections correctly.
	gs.ConnState = func(conn net.Conn, newState http.ConnState) {
		gs.connToPropsMtx.RLock()
		connProps := gs.connToProps[conn]
		gs.connToPropsMtx.RUnlock()

		switch newState {

		case http.StateNew:
			// New connection -> StateNew
			connProps.protected = true
			gs.StartRoutine()

		case http.StateActive:
			// (StateNew, StateIdle) -> StateActive
			if gracefulHandler.IsClosed() {
				conn.Close()
				break
			}

			if !connProps.protected {
				connProps.protected = true
				gs.StartRoutine()
			}

		default:
			// (StateNew, StateActive) -> (StateIdle, StateClosed, StateHiJacked)
			if connProps.protected {
				gs.FinishRoutine()
				connProps.protected = false
			}
		}

		if gs.stateHandler != nil {
			gs.stateHandler(conn, connProps.state, newState)
		}
		connProps.state = newState

		gs.connToPropsMtx.Lock()
		if newState == http.StateClosed || newState == http.StateHijacked {
			delete(gs.connToProps, conn)
		} else {
			gs.connToProps[conn] = connProps
		}
		gs.connToPropsMtx.Unlock()

		if originalConnState != nil {
			originalConnState(conn, newState)
		}
	}

	// A hook to allow the server to notify others when it is ready to receive
	// requests; only used by tests.
	if gs.up != nil {
		gs.up <- listener
	}

	err := gs.Server.Serve(listener)
	// An error returned on shutdown is not worth reporting.
	if err != nil && gracefulHandler.IsClosed() {
		err = nil
	}

	// Wait for pending requests to complete regardless the Serve result.
	gs.wg.Wait()
	gs.shutdownFinished <- true
	return err
}

// StartRoutine increments the server's WaitGroup. Use this if a web request
// starts more goroutines and these goroutines are not guaranteed to finish
// before the request.
func (gs *GracefulServer) StartRoutine() {
	gs.wg.Add(1)
	atomic.AddInt32(&gs.routinesCount, 1)
}

// FinishRoutine decrements the server's WaitGroup. Use this to complement
// StartRoutine().
func (gs *GracefulServer) FinishRoutine() {
	gs.wg.Done()
	atomic.AddInt32(&gs.routinesCount, -1)
}

// RoutinesCount returns the number of currently running routines
func (gs *GracefulServer) RoutinesCount() int {
	return int(atomic.LoadInt32(&gs.routinesCount))
}

// gracefulHandler is used by GracefulServer to prevent calling ServeHTTP on
// to be closed kept-alive connections during the server shutdown.
type gracefulHandler struct {
	closed  int32 // accessed atomically.
	wrapped http.Handler
}

func newGracefulHandler(wrapped http.Handler) *gracefulHandler {
	return &gracefulHandler{
		wrapped: wrapped,
	}
}

func (gh *gracefulHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&gh.closed) == 0 {
		gh.wrapped.ServeHTTP(w, r)
		return
	}
	r.Body.Close()
	// Server is shutting down at this moment, and the connection that this
	// handler is being called on is about to be closed. So we do not need to
	// actually execute the handler logic.
}

func (gh *gracefulHandler) Close() {
	atomic.StoreInt32(&gh.closed, 1)
}

func (gh *gracefulHandler) IsClosed() bool {
	return atomic.LoadInt32(&gh.closed) == 1
}

// connProperties is used to keep track of various properties of connections
// maintained by a gracefulServer.
type connProperties struct {
	// Last known connection state.
	state http.ConnState
	// Tells whether the connection is going to defer server shutdown until
	// the current HTTP request is completed.
	protected bool
}
