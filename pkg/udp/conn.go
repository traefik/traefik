package udp

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/pires/go-proxyproto"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/ip"
)

// maxDatagramSize is the maximum size of a UDP datagram.
const maxDatagramSize = 65535

const closeRetryInterval = 500 * time.Millisecond

var errClosedListener = errors.New("udp: listener closed")

// ProxyProtocolConfig holds the Proxy Protocol configuration for UDP listeners.
type ProxyProtocolConfig struct {
	Insecure   bool
	TrustedIPs []string
}

// proxyProtocolConfig is the internal configuration with compiled IP checker.
type proxyProtocolConfig struct {
	insecure  bool
	ipChecker *ip.Checker
}

// Listener augments a session-oriented Listener over a UDP PacketConn.
type Listener struct {
	pConn *net.UDPConn

	connsMu sync.RWMutex
	conns   map[string]*Conn
	// connsProxyToClientAddr maps actual proxy source addresses to original client addresses.
	// e.g., "127.0.0.1:12345" (proxy) → "192.0.2.50:7777" (client)
	// Protected by connsMu.
	connsProxyToClientAddr map[string]string
	// accepting signifies whether the listener is still accepting new sessions.
	// It also serves as a sentinel for Shutdown to be idempotent.
	// Protected by connsMu.
	accepting bool

	acceptCh chan *Conn // no need for a Once, already indirectly guarded by accepting.

	// timeout defines how long to wait on an idle session,
	// before releasing its related resources.
	timeout time.Duration

	// readBufferPool is a pool of byte slices for UDP packet reading.
	readBufferPool sync.Pool

	// proxyProtocol holds Proxy Protocol configuration.
	proxyProtocol *proxyProtocolConfig
}

// ListenPacketConn creates a new listener from PacketConn.
func ListenPacketConn(packetConn net.PacketConn, timeout time.Duration, ppConfig *ProxyProtocolConfig) (*Listener, error) {
	if timeout <= 0 {
		return nil, errors.New("timeout should be greater than zero")
	}

	pConn, ok := packetConn.(*net.UDPConn)
	if !ok {
		return nil, errors.New("packet conn is not an UDPConn")
	}

	l := &Listener{
		pConn:     pConn,
		acceptCh:  make(chan *Conn),
		conns:     make(map[string]*Conn),
		accepting: true,
		timeout:   timeout,
		readBufferPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, maxDatagramSize)
			},
		},
		connsProxyToClientAddr: make(map[string]string),
	}

	if ppConfig != nil {
		l.proxyProtocol = &proxyProtocolConfig{
			insecure: ppConfig.Insecure,
		}
		if !ppConfig.Insecure && len(ppConfig.TrustedIPs) > 0 {
			checker, err := ip.NewChecker(ppConfig.TrustedIPs)
			if err != nil {
				return nil, fmt.Errorf("creating IP checker: %w", err)
			}
			l.proxyProtocol.ipChecker = checker
		}
	}

	go l.readLoop()

	return l, nil
}

// Listen creates a new listener.
func Listen(listenConfig net.ListenConfig, network, address string, timeout time.Duration, ppConfig *ProxyProtocolConfig) (*Listener, error) {
	if timeout <= 0 {
		return nil, errors.New("timeout should be greater than zero")
	}

	packetConn, err := listenConfig.ListenPacket(context.Background(), network, address)
	if err != nil {
		return nil, fmt.Errorf("listen packet: %w", err)
	}

	l, err := ListenPacketConn(packetConn, timeout, ppConfig)
	if err != nil {
		return nil, fmt.Errorf("listen packet conn: %w", err)
	}

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
	l.connsMu.Lock()
	defer l.connsMu.Unlock()
	err := l.pConn.Close()
	for k, v := range l.conns {
		v.close()
		delete(l.conns, k)
	}
	l.connsProxyToClientAddr = make(map[string]string)
	close(l.acceptCh)
	return err
}

// Shutdown closes the listener.
// It immediately stops accepting new sessions,
// and it waits for all existing sessions to terminate,
// and a maximum of graceTimeout.
// Then it forces close any session left.
func (l *Listener) Shutdown(graceTimeout time.Duration) error {
	l.connsMu.Lock()
	if !l.accepting {
		l.connsMu.Unlock()
		return nil
	}
	l.accepting = false
	l.connsMu.Unlock()

	retryInterval := closeRetryInterval
	if retryInterval > graceTimeout {
		retryInterval = graceTimeout
	}
	start := time.Now()
	end := start.Add(graceTimeout)
	for !time.Now().After(end) {
		l.connsMu.RLock()
		if len(l.conns) == 0 {
			l.connsMu.RUnlock()
			break
		}
		l.connsMu.RUnlock()

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
		// before c.msgs is emptied via Read().
		// Reuses buffers via the readBufferPool sync.Pool.
		buf := l.readBufferPool.Get().([]byte)

		n, raddr, err := l.pConn.ReadFrom(buf)
		if err != nil {
			l.readBufferPool.Put(buf)
			return
		}

		packet := buf[:n]
		originalSource := raddr

		// Proxy Protocol handling.
		if l.shouldParseProxyProtocol(originalSource) {
			header, payload, err := l.parseProxyProtocol(packet)
			if err != nil {
				// Log and drop packet on parse error.
				log.Debug().Err(err).
					Str("source", raddr.String()).
					Msg("Failed to parse Proxy Protocol header, dropping packet")
				l.readBufferPool.Put(buf)
				continue
			}

			if header != nil {
				// Use header's source address for session keying.
				srcAddr, _, ok := header.UDPAddrs()
				if ok {
					// Record mapping of actual proxy source to client address for subsequent packets.
					l.connsMu.Lock()
					l.connsProxyToClientAddr[originalSource.String()] = srcAddr.String()
					l.connsMu.Unlock()

					raddr = srcAddr
					packet = payload // Use stripped payload.
				} else {
					// Header present but not UDP protocol, log and drop.
					log.Debug().
						Str("source", originalSource.String()).
						Msg("Proxy Protocol header not UDP type, dropping packet")
					l.readBufferPool.Put(buf)
					continue
				}
			} else {
				// No header: check if this source has an existing Proxy Protocol session.
				l.connsMu.RLock()
				clientAddr, exists := l.connsProxyToClientAddr[originalSource.String()]
				l.connsMu.RUnlock()

				if exists {
					// Use the client address instead of actual source.
					clientUDPAddr, err := net.ResolveUDPAddr("udp", clientAddr)
					if err == nil {
						raddr = clientUDPAddr
					}
				}
			}
		}

		conn, err := l.getConn(raddr)
		if err != nil {
			l.readBufferPool.Put(buf)
			continue
		}

		select {
		// Receiver must call releaseReadBuffer() when done reading the data.
		case conn.receiveCh <- packet:
		case <-conn.doneCh:
			l.readBufferPool.Put(buf)
			continue
		}
	}
}

// shouldParseProxyProtocol determines if we should attempt to parse
// Proxy Protocol header from this source address.
func (l *Listener) shouldParseProxyProtocol(originalSource net.Addr) bool {
	if l.proxyProtocol == nil {
		return false
	}

	if l.proxyProtocol.insecure {
		return true
	}

	udpAddr, ok := originalSource.(*net.UDPAddr)
	if !ok {
		return false
	}

	if l.proxyProtocol.ipChecker == nil {
		return false
	}

	return l.proxyProtocol.ipChecker.ContainsIP(udpAddr.IP)
}

// parseProxyProtocol attempts to parse Proxy Protocol header from packet.
// Returns (header, payload, error) where payload has header bytes stripped.
// If no header present, returns (nil, originalPacket, nil).
func (l *Listener) parseProxyProtocol(packet []byte) (*proxyproto.Header, []byte, error) {
	reader := bufio.NewReader(bytes.NewReader(packet))

	header, err := proxyproto.Read(reader)
	if err != nil {
		if errors.Is(err, proxyproto.ErrNoProxyProtocol) {
			return nil, packet, nil
		}
		return nil, nil, fmt.Errorf("parsing proxy protocol: %w", err)
	}

	headerLen := len(packet) - reader.Buffered()
	if headerLen < 0 || headerLen > len(packet) {
		return nil, nil, fmt.Errorf("invalid header length: %d", headerLen)
	}

	payload := packet[headerLen:]
	return header, payload, nil
}

// getConn returns the ongoing session with raddr if it exists, or creates a new
// one otherwise.
func (l *Listener) getConn(raddr net.Addr) (*Conn, error) {
	l.connsMu.Lock()
	defer l.connsMu.Unlock()

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
		timeout:   l.timeout,
	}
}

// Conn represents an on-going session with a client, over UDP packets.
type Conn struct {
	listener *Listener
	rAddr    net.Addr

	receiveCh chan []byte // to receive the data from the listener's readLoop.
	readCh    chan []byte // to receive the buffer into which we should Read.
	sizeCh    chan int    // to synchronize with the end of a Read.
	msgs      [][]byte    // to store data from listener, to be consumed by Reads.

	muActivity   sync.RWMutex
	lastActivity time.Time // the last time the session saw either read or write activity.

	timeout  time.Duration // for timeouts.
	doneOnce sync.Once
	doneCh   chan struct{}
}

// readLoop waits for data to come from the listener's readLoop.
// It then waits for a Read operation to be ready to consume said data,
// that is to say it waits on readCh to receive the slice of bytes that the Read operation wants to read onto.
// The Read operation receives the signal that the data has been written to the slice of bytes through the sizeCh.
func (c *Conn) readLoop() {
	ticker := time.NewTicker(c.timeout / 10)
	defer ticker.Stop()

	for {
		if len(c.msgs) == 0 {
			select {
			case msg := <-c.receiveCh:
				c.msgs = append(c.msgs, msg)
			case <-ticker.C:
				c.muActivity.RLock()
				deadline := c.lastActivity.Add(c.timeout)
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
			// Return buffer to sync.Pool once done reading from it.
			c.listener.readBufferPool.Put(msg)
			c.sizeCh <- n
		case msg := <-c.receiveCh:
			c.msgs = append(c.msgs, msg)
		case <-ticker.C:
			c.muActivity.RLock()
			deadline := c.lastActivity.Add(c.timeout)
			c.muActivity.RUnlock()
			if time.Now().After(deadline) {
				c.Close()
				return
			}
		}
	}
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.rAddr
}

// Read reads up to len(p) bytes into p from the connection.
// Each call corresponds to at most one datagram.
// If p is smaller than the datagram, the extra bytes will be discarded.
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

// Write writes len(p) bytes from p to the underlying connection.
// Each call sends at most one datagram.
// It is an error to send a message larger than the system's max UDP datagram size.
func (c *Conn) Write(p []byte) (n int, err error) {
	c.muActivity.Lock()
	c.lastActivity = time.Now()
	c.muActivity.Unlock()

	return c.listener.pConn.WriteTo(p, c.rAddr)
}

func (c *Conn) close() {
	c.doneOnce.Do(func() {
		// Release any buffered data before closing.
		for _, msg := range c.msgs {
			c.listener.readBufferPool.Put(msg)
		}
		c.msgs = nil
		close(c.doneCh)
	})
}

// Close releases resources related to the Conn.
func (c *Conn) Close() error {
	c.close()

	c.listener.connsMu.Lock()
	defer c.listener.connsMu.Unlock()
	delete(c.listener.conns, c.rAddr.String())
	return nil
}
