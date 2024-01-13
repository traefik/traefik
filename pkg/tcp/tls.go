package tcp

import (
	"context"
	"crypto/tls"
)

// TLSHandler handles TLS connections.
type TLSHandler struct {
	Next   Handler
	Config *tls.Config
}

// ServeTCP terminates the TLS connection.
func (t *TLSHandler) ServeTCP(ctx context.Context, conn WriteCloser) {
	conn = tls.Server(conn, t.Config)
	conn = NewTLSConnectionLogger(ctx, conn)
	t.Next.ServeTCP(ctx, conn)
}

// Track the decrypted byte counters on TLS connections
type tlsConnectionLogger struct {
	baseConnectionLogger
}

func (l *tlsConnectionLogger) Close() error {
	l.AddStats()
	return l.baseConnectionLogger.WriteCloser.Close()
}

func (l *tlsConnectionLogger) CloseWrite() error {
	l.AddStats()
	return l.baseConnectionLogger.WriteCloser.CloseWrite()
}

func (l *tlsConnectionLogger) AddStats() {
	connectionLogData := GetConnectionLog(l.ctx)
	if connectionLogData != nil {
		connectionLogData.Core["tls"] = true

		connectionLogData.Core["plainBytesRead"] = l.baseConnectionLogger.bytesRead
		connectionLogData.Core["plainBytesWritten"] = l.baseConnectionLogger.bytesWritten

		// get the SNI from the TLS connection
		connectionLogData.Core["sni"] = l.WriteCloser.(*tls.Conn).ConnectionState().ServerName
	}
}

func NewTLSConnectionLogger(ctx context.Context, conn WriteCloser) *tlsConnectionLogger {
	return &tlsConnectionLogger{
		baseConnectionLogger: baseConnectionLogger{
			WriteCloser: conn,
			ctx:         ctx,
		},
	}
}
