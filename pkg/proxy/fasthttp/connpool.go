package fasthttp

import (
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog/log"
)

// Conn is an enriched net.Conn.
type Conn struct {
	net.Conn

	idleAt      time.Time // the last time it was marked as idle.
	idleTimeout time.Duration
}

func (c *Conn) isExpired() bool {
	expTime := c.idleAt.Add(c.idleTimeout)
	return c.idleTimeout > 0 && time.Now().After(expTime)
}

// ConnPool is a net.Conn pool implementation using channels.
type ConnPool struct {
	dialer          func() (net.Conn, error)
	idleConns       chan *Conn
	idleConnTimeout time.Duration
}

// NewConnPool creates a new ConnPool.
func NewConnPool(maxIdleConn int, idleConnTimeout time.Duration, dialer func() (net.Conn, error)) *ConnPool {
	c := &ConnPool{
		dialer:          dialer,
		idleConns:       make(chan *Conn, maxIdleConn),
		idleConnTimeout: idleConnTimeout,
	}
	c.cleanIdleConns()

	return c
}

// AcquireConn returns an idle net.Conn from the pool.
func (c *ConnPool) AcquireConn() (*Conn, error) {
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
func (c *ConnPool) ReleaseConn(co *Conn) {
	co.idleAt = time.Now()
	c.releaseConn(co)
}

// cleanIdleConns is a routine cleaning the expired connections at a regular basis.
func (c *ConnPool) cleanIdleConns() {
	defer time.AfterFunc(c.idleConnTimeout/2, c.cleanIdleConns)

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

func (c *ConnPool) acquireConn() (*Conn, error) {
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

func (c *ConnPool) releaseConn(co *Conn) {
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

func (c *ConnPool) askForNewConn(errCh chan<- error) {
	co, err := c.dialer()
	if err != nil {
		errCh <- fmt.Errorf("create conn: %w", err)
		return
	}

	c.releaseConn(&Conn{
		Conn:        co,
		idleAt:      time.Now(),
		idleTimeout: c.idleConnTimeout,
	})
}
