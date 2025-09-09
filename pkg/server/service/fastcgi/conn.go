package fastcgi

import (
	"net"
	"time"
)

type poolConn struct {
	inner net.Conn
	// as long as we don't have multiplexing, we can use reqID per poolConn
	reqID int
}

func openConn(network, addr string, reqID int, dialer *net.Dialer) (poolConn, error) {
	conn, err := dialer.Dial(network, addr)
	if err != nil {
		return poolConn{}, err
	}
	return poolConn{
		reqID: reqID,
		inner: conn,
	}, nil
}

func (pc *poolConn) resetTimeouts() error {
	if err := pc.inner.SetWriteDeadline(time.Time{}); err != nil {
		return err
	}
	return pc.inner.SetReadDeadline(time.Time{})
}

func (pc *poolConn) setTimeouts(writeTimeout, readTimeout time.Duration) error {
	var (
		writeDeadline = time.Now().Add(writeTimeout)
		readDeadline  = time.Now().Add(readTimeout)
	)
	if err := pc.inner.SetWriteDeadline(writeDeadline); err != nil {
		return err
	}
	return pc.inner.SetReadDeadline(readDeadline)
}
