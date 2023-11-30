package tcp

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/sys/unix"
	"net"

	"github.com/traefik/traefik/v2/pkg/log"
)

// TLSHandler handles TLS connections.
type TLSHandler struct {
	Next   Handler
	Config *tls.Config
}

// ServeTCP terminates the TLS connection.
func (t *TLSHandler) ServeTCP(conn WriteCloser) {
	conn = tls.Server(conn, t.Config)
	conn = NewAccountedConnection(conn, "decrypted")
	t.Next.ServeTCP(conn)
}

// Create a net.Conn implementation which wraps the original connection and tracks the bytes on the connection
type accountedConnection struct {
	WriteCloser
	name         string
	bytesRead    uint64
	bytesWritten uint64
}

func NewAccountedConnection(conn WriteCloser, name string) *accountedConnection {
	return &accountedConnection{
		WriteCloser: conn,
		name:        name,
	}
}

func (c *accountedConnection) Read(b []byte) (n int, err error) {
	n, err = c.WriteCloser.Read(b)
	c.bytesRead += uint64(n)
	return n, err
}

func (c *accountedConnection) Write(b []byte) (n int, err error) {
	n, err = c.WriteCloser.Write(b)
	c.bytesWritten += uint64(n)
	return n, err
}

func (c *accountedConnection) DumpStats() {
	extra := ""

	switch c.WriteCloser.(type) {
	case *tls.Conn:
		// get the SNI from the TLS connection
		sni := c.WriteCloser.(*tls.Conn).ConnectionState().ServerName
		extra = fmt.Sprintf(", SNI: %s", sni)

	case *net.TCPConn:
		tcp_conn := c.WriteCloser.(*net.TCPConn)
		address := tcp_conn.RemoteAddr()
		if address.(*net.TCPAddr).IP.To4() != nil {
			extra = ", ipv4"
		} else {
			extra = ", ipv6"
		}

		rawConn, err := tcp_conn.SyscallConn()
		if err != nil {
			log.WithoutContext().Errorf("Error getting raw connection: %v", err)
		} else {
			err := rawConn.Control(func(fd uintptr) {
				info, err := unix.GetsockoptTCPInfo(int(fd), unix.IPPROTO_TCP, unix.TCP_INFO)
				if err != nil {
					log.WithoutContext().Errorf("Error getting TCP_INFO: %v", err)
				}
				extra += fmt.Sprintf(", lost: %d, retrans: %d, tcpi_segs_out: %d, tcpi_bytes_sent: %d, tcpi_segs_in: %d, tcpi_bytes_received: %d", info.Lost, info.Retrans, info.Segs_out, info.Bytes_sent, info.Segs_in, info.Bytes_received)
			})

			if err != nil {
				log.WithoutContext().Errorf("Error getting raw connection: %v", err)
			}
		}

	default:
		log.WithoutContext().Errorf("Unknown connection type: %T", c.WriteCloser)
	}
	log.WithoutContext().Infof("Connection closed: %s, Remote Address: %s, Bytes Read: %d, Bytes Written: %d%s", c.name, c.RemoteAddr().String(), c.bytesRead, c.bytesWritten, extra)
}

func (c *accountedConnection) Close() error {
	c.DumpStats()
	return c.WriteCloser.Close()
}

func (c *accountedConnection) CloseWrite() error {
	c.DumpStats()
	return c.WriteCloser.CloseWrite()
}
