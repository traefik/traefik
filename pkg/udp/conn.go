package udp

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

const receiveMTU = 8192

const closeRetryInterval = 500 * time.Millisecond

// connTimeout determines how long to wait on an idle session,
// before releasing all resources related to that session.
const connTimeout = 3 * time.Second

var timeoutTicker = connTimeout / 10

var errClosedListener = errors.New("udp: listener closed")

// Listener augments a session-oriented Listener over a UDP PacketConn.
type Listener struct {
	pConn *net.UDPConn

	mu    sync.RWMutex
	conns map[string]*Conn
	// accepting signifies whether the listener is still accepting new sessions.
	// It also serves as a sentinel for Shutdown to be idempotent.
	accepting bool

	acceptCh chan *Conn // no need for a Once, already indirectly guarded by accepting.
}

// Listen creates a new listener.
func Listen(network string, laddr *net.UDPAddr) (*Listener, error) {
	conn, err := net.ListenUDP(network, laddr)
	if err != nil {
		return nil, err
	}

	l := &Listener{
		pConn:     conn,
		acceptCh:  make(chan *Conn),
		conns:     make(map[string]*Conn),
		accepting: true,
	}

	go l.readLoop()

	return l, nil
}

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (*Conn, error) {
	c := <-l.acceptCh
	if c == nil {
		// l.acceptCh got closed
		return nil, errClosedListener
	}
	return c, nil
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return l.pConn.LocalAddr()
}

// Close closes the listener.
// It is like Shutdown with a zero graceTimeout.
func (l *Listener) Close() error {
	return l.Shutdown(0)
}

// close should not be called more than once.
func (l *Listener) close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	err := l.pConn.Close()
	for k, v := range l.conns {
		v.close()
		delete(l.conns, k)
	}
	close(l.acceptCh)
	return err
}

// Shutdown closes the listener.
// It immediately stops accepting new sessions,
// and it waits for all existing sessions to terminate,
// and a maximum of graceTimeout.
// Then it forces close any session left.
func (l *Listener) Shutdown(graceTimeout time.Duration) error {
	l.mu.Lock()
	if !l.accepting {
		l.mu.Unlock()
		return nil
	}
	l.accepting = false
	l.mu.Unlock()

	retryInterval := closeRetryInterval
	if retryInterval > graceTimeout {
		retryInterval = graceTimeout
	}
	start := time.Now()
	end := start.Add(graceTimeout)
	for {
		if time.Now().After(end) {
			break
		}

		l.mu.RLock()
		if len(l.conns) == 0 {
			l.mu.RUnlock()
			break
		}
		l.mu.RUnlock()

		time.Sleep(retryInterval)
	}
	return l.close()
}

// readLoop receives all packets from all remotes.
// If a packet comes from a remote that is already known to us (i.e. a "session"),
// we find that session, and otherwise we create a new one.
// We then send the data the session's readLoop.
func (l *Listener) readLoop() {
	for {
		// Allocating a new buffer for every read avoids
		// overwriting data in c.msgs in case the next packet is received
		// before c.msgs is emptied via Read()
		buf := make([]byte, receiveMTU)
		n, raddr, err := l.pConn.ReadFrom(buf)
		if err != nil {
			return
		}
		conn, err := l.getConn(raddr)
		if err != nil {
			continue
		}
		select {
		case conn.receiveCh <- buf[:n]:
		case <-conn.doneCh:
			continue
		}
	}
}

// getConn returns the ongoing session with raddr if it exists, or creates a new
// one otherwise.
func (l *Listener) getConn(raddr net.Addr) (*Conn, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	conn, ok := l.conns[raddr.String()]
	if ok {
		return conn, nil
	}

	if !l.accepting {
		return nil, errClosedListener
	}
	conn = l.newConn(raddr)
	l.conns[raddr.String()] = conn
	l.acceptCh <- conn
	go conn.readLoop()

	return conn, nil
}

func (l *Listener) newConn(rAddr net.Addr) *Conn {
	return &Conn{
		listener:  l,
		rAddr:     rAddr,
		receiveCh: make(chan []byte),
		readCh:    make(chan []byte),
		sizeCh:    make(chan int),
		doneCh:    make(chan struct{}),
		timeout:   timeoutTicker,
	}
}

// Conn represents an on-going session with a client, over UDP packets.
type Conn struct {
	listener *Listener
	rAddr    net.Addr

	receiveCh chan []byte // to receive the data from the listener's readLoop
	readCh    chan []byte // to receive the buffer into which we should Read
	sizeCh    chan int    // to synchronize with the end of a Read
	msgs      [][]byte    // to store data from listener, to be consumed by Reads

	muActivity   sync.RWMutex
	lastActivity time.Time // the last time the session saw either read or write activity

	timeout  time.Duration // for timeouts
	doneOnce sync.Once
	doneCh   chan struct{}
}

// readLoop waits for data to come from the listener's readLoop.
// It then waits for a Read operation to be ready to consume said data,
// that is to say it waits on readCh to receive the slice of bytes that the Read operation wants to read onto.
// The Read operation receives the signal that the data has been written to the slice of bytes through the sizeCh.
func (c *Conn) readLoop() {
	ticker := time.NewTicker(c.timeout)
	defer ticker.Stop()

	for {
		if len(c.msgs) == 0 {
			select {
			case msg := <-c.receiveCh:
				c.msgs = append(c.msgs, msg)
			case <-ticker.C:
				c.muActivity.RLock()
				deadline := c.lastActivity.Add(connTimeout)
				c.muActivity.RUnlock()
				if time.Now().After(deadline) {
					c.Close()
					return
				}
				continue
			}
		}

		select {
		case cBuf := <-c.readCh:
			msg := c.msgs[0]
			c.msgs = c.msgs[1:]
			n := copy(cBuf, msg)
			c.sizeCh <- n
		case msg := <-c.receiveCh:
			c.msgs = append(c.msgs, msg)
		case <-ticker.C:
			c.muActivity.RLock()
			deadline := c.lastActivity.Add(connTimeout)
			c.muActivity.RUnlock()
			if time.Now().After(deadline) {
				c.Close()
				return
			}
		}
	}
}

// Read implements io.Reader for a Conn.
func (c *Conn) Read(p []byte) (int, error) {
	select {
	case c.readCh <- p:
		n := <-c.sizeCh
		c.muActivity.Lock()
		c.lastActivity = time.Now()
		c.muActivity.Unlock()
		return n, nil
	case <-c.doneCh:
		return 0, io.EOF
	}
}

// Write implements io.Writer for a Conn.
func (c *Conn) Write(p []byte) (n int, err error) {
	l := c.listener
	if l == nil {
		return 0, io.EOF
	}

	c.muActivity.Lock()
	c.lastActivity = time.Now()
	c.muActivity.Unlock()
	return l.pConn.WriteTo(p, c.rAddr)
}

func (c *Conn) close() {
	c.doneOnce.Do(func() {
		close(c.doneCh)
	})
}

// Close releases resources related to the Conn.
func (c *Conn) Close() error {
	c.close()

	c.listener.mu.Lock()
	defer c.listener.mu.Unlock()
	delete(c.listener.conns, c.rAddr.String())
	return nil
}
