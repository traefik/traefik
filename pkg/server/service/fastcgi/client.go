package fastcgi

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

var errMaxConnsExceeded = errors.New("max connections exceeded")

var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

type Client struct {
	network string
	addr    string

	writeTimeout   time.Duration
	readTimeout    time.Duration
	dialTimeout    time.Duration
	acquireTimeout time.Duration
	logStdErr      bool

	maxConns    int
	actualConns atomic.Int32
	conns       chan poolConn
}

func NewClient(
	network, addr string,
	maxConns int,
	writeTimout, readTimeout, dialTimeout, acquireTimeout time.Duration,
	logStdErr bool,
) (*Client, error) {
	// TODO make GetValuesResult request to get client MaxConns
	return newClient(
		network, addr,
		maxConns,
		writeTimout,
		readTimeout,
		dialTimeout,
		acquireTimeout,
		logStdErr,
	), nil
}

func newClient(
	network, addr string,
	maxConns int,
	writeTimout, readTimeout, dialTimeout, acquireTimeout time.Duration,
	logStdErr bool,
) *Client {
	return &Client{
		network: network,
		addr:    addr,

		writeTimeout: writeTimout,
		readTimeout:  readTimeout,
		dialTimeout:  dialTimeout,
		logStdErr:    logStdErr,

		acquireTimeout: acquireTimeout,
		maxConns:       maxConns,
		actualConns:    atomic.Int32{},
		conns:          make(chan poolConn, maxConns),
	}
}

func (c *Client) Do(req *Request) (*ConnReadCloser, error) {
	req.params["REQUEST_METHOD"] = req.httpMethod
	conn, err := c.getConn()
	if err != nil {
		return nil, err
	}
	if err = conn.setTimeouts(c.writeTimeout, c.readTimeout); err != nil {
		c.closeConn(conn)
		return nil, err
	}

	writer := fastcgiWriter{
		reqID:     uint16(conn.reqID),
		reqWriter: conn.inner,
		buff:      bufPool.Get().(*bytes.Buffer),
	}
	defer bufPool.Put(writer.buff)

	if err = writer.writeBeginReq(req.role, true); err != nil {
		c.closeConn(conn)
		return nil, err
	}

	if err = writer.writeParamsReq(req.params); err != nil {
		c.closeConn(conn)
		return nil, err
	}

	if req.body != nil {
		err = writer.writeStdinReq(req.body)
		if err != nil {
			c.closeConn(conn)
			return nil, err
		}
	}

	fastCgiReader := newFastCgiReader(conn.inner)
	return &ConnReadCloser{
		client: c,
		conn:   conn,
		reader: bufio.NewReader(fastCgiReader),
		error:  &fastCgiReader.error,
	}, nil
}

func (c *Client) getConn() (poolConn, error) {
	// try to get connection from pool
	select {
	case conn := <-c.conns:
		return conn, nil
	default:
	}

	// try to open new connection
	newConn, err := c.newPoolConn()
	if err == nil {
		return newConn, nil
	}
	if !errors.Is(err, errMaxConnsExceeded) {
		return poolConn{}, err
	}

	// waiting for connection to be released
	timer := time.NewTimer(c.acquireTimeout)
	defer timer.Stop()
	select {
	case conn := <-c.conns:
		return conn, nil
	case <-timer.C:
		return poolConn{}, errors.New("connection acquire timeout occurred")
	}
}

func (c *Client) newPoolConn() (poolConn, error) {
	// CAS loop to increment actualConns value
	for {
		n := c.actualConns.Load()
		if int(n) >= c.maxConns {
			return poolConn{}, errMaxConnsExceeded
		}
		if c.actualConns.CompareAndSwap(n, n+1) {
			break
		}
	}
	dialer := net.Dialer{Timeout: c.dialTimeout}
	newConn, err := openConn(c.network, c.addr, 1, &dialer)
	if err != nil {
		c.actualConns.Add(-1)
		return poolConn{}, err
	}

	return newConn, nil
}

func (c *Client) closeConn(conn poolConn) {
	if err := conn.inner.Close(); err != nil {
		log.Err(err).Msg("failed to close fastcgi connection")
	}
	c.actualConns.Add(-1)
}

func (c *Client) putConn(conn poolConn) {
	if err := conn.resetTimeouts(); err != nil {
		// suppress the error and close connection
		// because request is completed anyway
		c.closeConn(conn)
		log.Err(err).Msg("failed to reset fastcgi conn timeout")
		return
	}
	c.conns <- conn
}

type ConnReadCloser struct {
	reader  *bufio.Reader
	client  *Client
	error   *bytes.Buffer
	conn    poolConn
	drained bool
}

func (crc *ConnReadCloser) Read(p []byte) (int, error) {
	n, err := crc.reader.Read(p)
	// errFastCgiRequestEOF indicates that FCGI_END_REQUEST has been read
	// in such case connection can be reused
	if errors.Is(err, errFastCgiRequestEOF) {
		crc.drained = true
		err = io.EOF
	}

	return n, err
}

func (crc *ConnReadCloser) Close() error {
	if !crc.drained {
		crc.client.closeConn(crc.conn)
		return nil
	}
	crc.client.putConn(crc.conn)

	if !crc.client.logStdErr {
		return nil
	}

	// log stderr if exist
	stderr := crc.error.String()
	if len(stderr) > 0 {
		log.Error().Msgf("FastCGI: stderr received: '%s'", stderr)
	}

	return nil
}
