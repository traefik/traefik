package fast

import (
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog/log"
)

// conn is an enriched net.Conn.
type conn struct {
	net.Conn

	idleAt      time.Time // the last time it was marked as idle.
	idleTimeout time.Duration
}

func (c *conn) isExpired() bool {
	expTime := c.idleAt.Add(c.idleTimeout)
	return c.idleTimeout > 0 && time.Now().After(expTime)
}

// connPool is a net.Conn pool implementation using channels.
type connPool struct {
	dialer          func() (net.Conn, error)
	idleConns       chan *conn
	idleConnTimeout time.Duration
	ticker          *time.Ticker
	doneCh          chan struct{}
}

// newConnPool creates a new connPool.
func newConnPool(maxIdleConn int, idleConnTimeout time.Duration, dialer func() (net.Conn, error)) *connPool {
	c := &connPool{
		dialer:          dialer,
		idleConns:       make(chan *conn, maxIdleConn),
		idleConnTimeout: idleConnTimeout,
		doneCh:          make(chan struct{}),
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

		if !co.isExpired() {
			return co, nil
		}

		// As the acquired conn is expired we can close it
		// without putting it again into the pool.
		if err := co.Close(); err != nil {
			log.Debug().
				Err(err).
				Msg("Unexpected error while releasing the connection")
		}
	}
}

// ReleaseConn releases the given net.Conn to the pool.
func (c *connPool) ReleaseConn(co *conn) {
	co.idleAt = time.Now()
	c.releaseConn(co)
}

// cleanIdleConns is a routine cleaning the expired connections at a regular basis.
func (c *connPool) cleanIdleConns() {
	for {
		select {
		case co := <-c.idleConns:
			if !co.isExpired() {
				c.releaseConn(co)
				return
			}

			if err := co.Close(); err != nil {
				log.Debug().
					Err(err).
					Msg("Unexpected error while releasing the connection")
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

	c.releaseConn(&conn{
		Conn:        co,
		idleAt:      time.Now(),
		idleTimeout: c.idleConnTimeout,
	})
}
