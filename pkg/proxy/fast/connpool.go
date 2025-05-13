package fast

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

// rwWithUpgrade contains a ResponseWriter and an upgradeHandler,
// used to upgrade the connection (e.g. Websockets).
type rwWithUpgrade struct {
	ReqMethod string
	RW        http.ResponseWriter
	Upgrade   upgradeHandler
}

// conn is an enriched net.Conn.
type conn struct {
	net.Conn

	RWCh  chan rwWithUpgrade
	ErrCh chan error

	br *bufio.Reader

	idleAt      time.Time // the last time it was marked as idle.
	idleTimeout time.Duration

	responseHeaderTimeout time.Duration

	expectedResponse atomic.Bool
	broken           atomic.Bool
	upgraded         atomic.Bool

	closeMu  sync.Mutex
	closed   bool
	closeErr error

	bufferPool        *pool[[]byte]
	limitedReaderPool *pool[*io.LimitedReader]
}

// Read reads data from the connection.
// Overrides conn Read to use the buffered reader.
func (c *conn) Read(b []byte) (n int, err error) {
	return c.br.Read(b)
}

// Close closes the connection.
// Ensures that connection is closed only once,
// to avoid duplicate close error.
func (c *conn) Close() error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if c.closed {
		return c.closeErr
	}

	c.closed = true
	c.closeErr = c.Conn.Close()

	return c.closeErr
}

// isStale returns whether the connection is in an invalid state (i.e. expired/broken).
func (c *conn) isStale() bool {
	expTime := c.idleAt.Add(c.idleTimeout)
	return c.idleTimeout > 0 && time.Now().After(expTime) || c.broken.Load()
}

// isUpgraded returns whether this connection has been upgraded (e.g. Websocket).
// An upgraded connection should not be reused and putted back in the connection pool.
func (c *conn) isUpgraded() bool {
	return c.upgraded.Load()
}

// readLoop handles the successive HTTP response read operations on the connection,
// and watches for unsolicited bytes or connection errors when idle.
func (c *conn) readLoop() {
	defer c.Close()

	for {
		_, err := c.br.Peek(1)
		if err != nil {
			select {
			// An error occurred while a response was expected to be handled.
			case <-c.RWCh:
				c.ErrCh <- err
			// An error occurred on an idle connection.
			default:
				c.broken.Store(true)
			}
			return
		}

		// Unsolicited response received on an idle connection.
		if !c.expectedResponse.Load() {
			c.broken.Store(true)
			return
		}

		r := <-c.RWCh
		if err = c.handleResponse(r); err != nil {
			c.ErrCh <- err
			return
		}

		c.expectedResponse.Store(false)
		c.ErrCh <- nil
	}
}

func (c *conn) handleResponse(r rwWithUpgrade) error {
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	res.Header.SetNoDefaultContentType(true)

	for {
		var (
			timer      *time.Timer
			errTimeout atomic.Pointer[timeoutError]
		)
		if c.responseHeaderTimeout > 0 {
			timer = time.AfterFunc(c.responseHeaderTimeout, func() {
				errTimeout.Store(&timeoutError{errors.New("timeout awaiting response headers")})
				c.Close() // This close call is needed to interrupt the read operation below when the timeout is over.
			})
		}

		res.Header.SetNoDefaultContentType(true)
		if err := res.Header.Read(c.br); err != nil {
			if c.responseHeaderTimeout > 0 {
				if errT := errTimeout.Load(); errT != nil {
					return errT
				}
			}
			return err
		}

		if timer != nil {
			timer.Stop()
		}

		fixPragmaCacheControl(&res.Header)

		resCode := res.StatusCode()
		is1xx := 100 <= resCode && resCode <= 199
		// treat 101 as a terminal status, see issue 26161
		is1xxNonTerminal := is1xx && resCode != http.StatusSwitchingProtocols
		if is1xxNonTerminal {
			removeConnectionHeaders(&res.Header)
			h := r.RW.Header()

			for _, header := range hopHeaders {
				res.Header.Del(header)
			}

			res.Header.VisitAll(func(key, value []byte) {
				r.RW.Header().Add(string(key), string(value))
			})

			r.RW.WriteHeader(res.StatusCode())
			// Clear headers, it's not automatically done by ResponseWriter.WriteHeader() for 1xx responses
			for k := range h {
				delete(h, k)
			}

			res.Reset()
			res.Header.SetNoDefaultContentType(true)

			continue
		}
		break
	}

	announcedTrailers := res.Header.Peek("Trailer")

	// Deal with 101 Switching Protocols responses: (WebSocket, h2c, etc)
	if res.StatusCode() == http.StatusSwitchingProtocols {
		r.Upgrade(r.RW, res, c)
		c.upgraded.Store(true) // As the connection has been upgraded, it cannot be added back to the pool.
		return nil
	}

	removeConnectionHeaders(&res.Header)

	for _, header := range hopHeaders {
		res.Header.Del(header)
	}

	if len(announcedTrailers) > 0 {
		res.Header.Add("Trailer", string(announcedTrailers))
	}

	res.Header.VisitAll(func(key, value []byte) {
		r.RW.Header().Add(string(key), string(value))
	})

	r.RW.WriteHeader(res.StatusCode())

	if noResponseBodyExpected(r.ReqMethod) {
		return nil
	}

	// When a body is not allowed for a given status code the body is ignored.
	// The connection will be marked as broken by the next Peek in the readloop.
	if !isBodyAllowedForStatus(res.StatusCode()) {
		return nil
	}

	contentLength := res.Header.ContentLength()

	if contentLength == 0 {
		return nil
	}

	// Chunked response, Content-Length is set to -1 by FastProxy when "Transfer-Encoding: chunked" header is received.
	if contentLength == -1 {
		cbr := httputil.NewChunkedReader(c.br)

		b := c.bufferPool.Get()
		if b == nil {
			b = make([]byte, bufferSize)
		}
		defer c.bufferPool.Put(b)

		if _, err := io.CopyBuffer(&writeFlusher{r.RW}, cbr, b); err != nil {
			return err
		}

		res.Header.Reset()
		res.Header.SetNoDefaultContentType(true)
		if err := res.Header.ReadTrailer(c.br); err != nil {
			return err
		}

		if res.Header.Len() > 0 {
			var announcedTrailersKey []string
			if len(announcedTrailers) > 0 {
				announcedTrailersKey = strings.Split(string(announcedTrailers), ",")
			}

			res.Header.VisitAll(func(key, value []byte) {
				for _, s := range announcedTrailersKey {
					if strings.EqualFold(s, strings.TrimSpace(string(key))) {
						r.RW.Header().Add(string(key), string(value))
						return
					}
				}

				r.RW.Header().Add(http.TrailerPrefix+string(key), string(value))
			})
		}

		return nil
	}

	// Response without Content-Length header.
	// The message body length is determined by the number of bytes received prior to the server closing the connection.
	if contentLength == -2 {
		b := c.bufferPool.Get()
		if b == nil {
			b = make([]byte, bufferSize)
		}
		defer c.bufferPool.Put(b)

		if _, err := io.CopyBuffer(r.RW, c.br, b); err != nil {
			return err
		}

		return nil
	}

	// Response with a valid Content-Length header.
	brl := c.limitedReaderPool.Get()
	if brl == nil {
		brl = &io.LimitedReader{}
	}
	defer c.limitedReaderPool.Put(brl)

	brl.R = c.br
	brl.N = int64(res.Header.ContentLength())

	b := c.bufferPool.Get()
	if b == nil {
		b = make([]byte, bufferSize)
	}
	defer c.bufferPool.Put(b)

	if _, err := io.CopyBuffer(r.RW, brl, b); err != nil {
		return err
	}

	return nil
}

// connPool is a net.Conn pool implementation using channels.
type connPool struct {
	dialer                func() (net.Conn, error)
	idleConns             chan *conn
	idleConnTimeout       time.Duration
	responseHeaderTimeout time.Duration
	ticker                *time.Ticker
	bufferPool            pool[[]byte]
	limitedReaderPool     pool[*io.LimitedReader]
	doneCh                chan struct{}
}

// newConnPool creates a new connPool.
func newConnPool(maxIdleConn int, idleConnTimeout, responseHeaderTimeout time.Duration, dialer func() (net.Conn, error)) *connPool {
	c := &connPool{
		dialer:                dialer,
		idleConns:             make(chan *conn, maxIdleConn),
		idleConnTimeout:       idleConnTimeout,
		responseHeaderTimeout: responseHeaderTimeout,
		doneCh:                make(chan struct{}),
	}

	if idleConnTimeout > 0 {
		c.ticker = time.NewTicker(c.idleConnTimeout / 2)
		go func() {
			for {
				select {
				case <-c.ticker.C:
					c.cleanIdleConns()
				case <-c.doneCh:
					return
				}
			}
		}()
	}

	return c
}

// Close closes stop the cleanIdleConn goroutine.
func (c *connPool) Close() {
	if c.idleConnTimeout > 0 {
		close(c.doneCh)
		c.ticker.Stop()
	}
}

// AcquireConn returns an idle net.Conn from the pool.
func (c *connPool) AcquireConn() (*conn, error) {
	for {
		co, err := c.acquireConn()
		if err != nil {
			return nil, err
		}

		if !co.isStale() {
			return co, nil
		}

		// As the acquired conn is stale we can close it
		// without putting it again into the pool.
		if err := co.Close(); err != nil {
			log.Debug().
				Err(err).
				Msg("Unexpected error while closing the connection")
		}
	}
}

// ReleaseConn releases the given net.Conn to the pool.
func (c *connPool) ReleaseConn(co *conn) {
	// An upgraded connection cannot be safely reused for another roundTrip,
	// thus we are not putting it back to the pool.
	if co.isUpgraded() {
		return
	}

	co.idleAt = time.Now()
	c.releaseConn(co)
}

// cleanIdleConns is a routine cleaning the expired connections at a regular basis.
func (c *connPool) cleanIdleConns() {
	for {
		select {
		case co := <-c.idleConns:
			if !co.isStale() {
				c.releaseConn(co)
				return
			}

			if err := co.Close(); err != nil {
				log.Debug().
					Err(err).
					Msg("Unexpected error while closing the connection")
			}

		default:
			return
		}
	}
}

func (c *connPool) acquireConn() (*conn, error) {
	select {
	case co := <-c.idleConns:
		return co, nil

	default:
		errCh := make(chan error, 1)
		go c.askForNewConn(errCh)

		select {
		case co := <-c.idleConns:
			return co, nil

		case err := <-errCh:
			return nil, err
		}
	}
}

func (c *connPool) releaseConn(co *conn) {
	select {
	case c.idleConns <- co:

	// Hitting the default case means that we have reached the maximum number of idle
	// connections, so we can close it.
	default:
		if err := co.Close(); err != nil {
			log.Debug().
				Err(err).
				Msg("Unexpected error while releasing the connection")
		}
	}
}

func (c *connPool) askForNewConn(errCh chan<- error) {
	co, err := c.dialer()
	if err != nil {
		errCh <- fmt.Errorf("create conn: %w", err)
		return
	}

	newConn := &conn{
		Conn:                  co,
		br:                    bufio.NewReaderSize(co, bufioSize),
		idleAt:                time.Now(),
		idleTimeout:           c.idleConnTimeout,
		responseHeaderTimeout: c.responseHeaderTimeout,
		RWCh:                  make(chan rwWithUpgrade),
		ErrCh:                 make(chan error),
		bufferPool:            &c.bufferPool,
		limitedReaderPool:     &c.limitedReaderPool,
	}
	go newConn.readLoop()

	c.releaseConn(newConn)
}

// isBodyAllowedForStatus reports whether a given response status code permits a body.
// See RFC 7230, section 3.3.
// From https://github.com/golang/go/blame/master/src/net/http/transfer.go#L459
func isBodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}

// noResponseBodyExpected reports whether a given request method permits a body.
// From https://github.com/golang/go/blame/master/src/net/http/transfer.go#L250
func noResponseBodyExpected(requestMethod string) bool {
	return requestMethod == "HEAD"
}
