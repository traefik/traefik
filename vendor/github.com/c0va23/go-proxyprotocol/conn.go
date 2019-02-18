package proxyprotocol

import (
	"bufio"
	"net"
	"sync"
	"time"
)

/*
Conn is wrapper on net.Conn with overrided RemoteAddr().

On first call Read() or RemoteAddr() parse proxyprotocol header and store
local and remote addresses.
*/
type Conn struct {
	conn         net.Conn
	logger       Logger
	readBuf      *bufio.Reader
	header       *Header
	headerErr    error
	headerParser HeaderParser
	once         sync.Once
}

// NewConn create wrapper on net.Conn.
func NewConn(
	conn net.Conn,
	logger Logger,
	headerParser HeaderParser,
) net.Conn {
	readBuf := bufio.NewReaderSize(conn, bufferSize)

	return &Conn{
		conn:         conn,
		readBuf:      readBuf,
		logger:       logger,
		headerParser: headerParser,
	}
}

func (conn *Conn) parseHeader() {
	conn.header, conn.headerErr = conn.headerParser.Parse(conn.readBuf)
	if conn.headerErr != nil {
		conn.logger.Printf("Header parse error: %s", conn.headerErr)
	}
}

/*
Read on first call parse proxyprotocol header.

If header parser return error, then error stored and returned. Otherwise call
Read on source connection.

Following calls of Read function check parse header error.
If error not nil, then error returned. Otherwise called source "conn.Read".
*/
func (conn *Conn) Read(buf []byte) (int, error) {
	conn.once.Do(conn.parseHeader)

	if nil != conn.headerErr {
		return 0, conn.headerErr
	}

	return conn.readBuf.Read(buf)
}

// Write proxy to conn.Write
func (conn *Conn) Write(buf []byte) (int, error) {
	return conn.conn.Write(buf)
}

// Close proxy to conn.Close
func (conn *Conn) Close() error {
	return conn.conn.Close()
}

// LocalAddr proxy to conn.LocalAddr
func (conn *Conn) LocalAddr() net.Addr {
	conn.once.Do(conn.parseHeader)
	if nil != conn.header {
		return conn.header.DstAddr
	}
	return conn.conn.LocalAddr()
}

/*
RemoteAddr on first call parse proxyprotocol header.

If header parser return header, then return source address from header.
Otherwise return original source address.
*/
func (conn *Conn) RemoteAddr() net.Addr {
	conn.once.Do(conn.parseHeader)
	if nil != conn.header {
		return conn.header.SrcAddr
	}
	return conn.conn.RemoteAddr()
}

// SetDeadline proxy to conn.SetDeadline
func (conn *Conn) SetDeadline(t time.Time) error {
	return conn.conn.SetDeadline(t)
}

// SetReadDeadline proxy to conn.SetReadDeadline
func (conn *Conn) SetReadDeadline(t time.Time) error {
	return conn.conn.SetReadDeadline(t)
}

// SetWriteDeadline  proxy to conn.SetWriteDeadline
func (conn *Conn) SetWriteDeadline(t time.Time) error {
	return conn.conn.SetWriteDeadline(t)
}
