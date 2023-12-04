package tcp

import (
	"context"
	"encoding/json"
	"golang.org/x/sys/unix"
	"net"

	"github.com/traefik/traefik/v2/pkg/log"
)

// Class to track bytes over the given connection
type baseConnectionLogger struct {
	WriteCloser
	ctx          context.Context
	bytesRead    uint64
	bytesWritten uint64
}

func (l *baseConnectionLogger) Read(b []byte) (n int, err error) {
	n, err = l.WriteCloser.Read(b)
	l.bytesRead += uint64(n)
	return n, err
}

func (l *baseConnectionLogger) Write(b []byte) (n int, err error) {
	n, err = l.WriteCloser.Write(b)
	l.bytesWritten += uint64(n)
	return n, err
}

// CoreLogData holds the fields computed from the request/response. Same as accesslog.CoreLogData
type CoreLogData map[string]interface{}

type ConnectionLogData struct {
	Core CoreLogData
}

const (
	ConnectionLogKey = "ConnectionLogKey"
)

func GetConnectionLog(ctx context.Context) *ConnectionLogData {
	if data, ok := ctx.Value(ConnectionLogKey).(*ConnectionLogData); ok {
		return data
	}
	return nil
}

type tcpConnectionLogger struct {
	baseConnectionLogger
}

func NewTCPConnectionLogger(ctx context.Context, conn WriteCloser) *tcpConnectionLogger {
	return &tcpConnectionLogger{
		baseConnectionLogger: baseConnectionLogger{
			WriteCloser: conn,
			ctx:         ctx,
		},
	}
}

func (l *tcpConnectionLogger) Close() error {
	l.AddStats()
	err := l.baseConnectionLogger.WriteCloser.Close()
	l.DumpStats()
	return err
}

func (l *tcpConnectionLogger) CloseWrite() error {
	l.AddStats()
	err := l.baseConnectionLogger.WriteCloser.CloseWrite()
	l.DumpStats()
	return err
}

func (l *tcpConnectionLogger) AddStats() {
	connectionLogData := GetConnectionLog(l.ctx)
	if connectionLogData == nil {
		return
	}
	connectionLogData.Core["tcpBytesReceived"] = l.baseConnectionLogger.bytesRead
	connectionLogData.Core["tcpBytesSent"] = l.baseConnectionLogger.bytesWritten

	tcp_conn := l.WriteCloser.(*net.TCPConn)
	address := tcp_conn.RemoteAddr()
	connectionLogData.Core["clientIp"] = address
    /*
	if address.(*net.TCPAddr).IP.To4() != nil {
		connectionLogData.Core["protocol"] = "ipv4"
	} else {
		connectionLogData.Core["protocol"] = "ipv6"
	}
    */

    // TODO: Linux-only stats?
	rawConn, err := tcp_conn.SyscallConn()
	if err != nil {
		log.WithoutContext().Errorf("Error getting raw connection: %v", err)
	} else {
		err := rawConn.Control(func(fd uintptr) {
			info, err := unix.GetsockoptTCPInfo(int(fd), unix.IPPROTO_TCP, unix.TCP_INFO)
			if err != nil {
				log.WithoutContext().Errorf("Error getting TCP_INFO: %v", err)
			}
			connectionLogData.Core["tcpLost"] = info.Lost
			connectionLogData.Core["tcpRetrans"] = info.Retrans
			connectionLogData.Core["tcpSegsOut"] = info.Segs_out
			connectionLogData.Core["tcpSegsIn"] = info.Segs_in

            // Override our own bytecounts as these from the kernel should be more accurate
			connectionLogData.Core["tcpBytesSent"] = info.Bytes_sent
			connectionLogData.Core["tcpBytesReceived"] = info.Bytes_received
		})

		if err != nil {
			log.WithoutContext().Errorf("Error getting raw connection: %v", err)
		}
	}
}

func (l *tcpConnectionLogger) DumpStats() {
    // TODO: Do this from config, allow field specification and JSON etc per accesslog
	if connectionLogData := GetConnectionLog(l.ctx); connectionLogData != nil {
		b, err := json.Marshal(connectionLogData.Core)
		if err != nil {
			log.Fatal(err)
		}
		log.WithoutContext().Infof("%s", string(b))
	}
}

// Create a net.Conn implementation which wraps the original connection and tracks the bytes on the connection
func NewConnectionLog(ctx context.Context, conn WriteCloser) (context.Context, *tcpConnectionLogger) {
	connectionLogData := ConnectionLogData{
		Core: CoreLogData{},
	}
	ctx = context.WithValue(ctx, ConnectionLogKey, &connectionLogData)

	return ctx, NewTCPConnectionLogger(ctx, conn)
}
