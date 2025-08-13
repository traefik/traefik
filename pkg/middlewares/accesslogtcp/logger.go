// Package accesslogtcp provides a minimal TCP access logging middleware for Traefik.
// It is inspired by the HTTP access log middleware, but logs TCP connection events.
package accesslogtcp

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/traefik/traefik/v3/pkg/types"
)

// Handler is the main struct for TCP access logging.
// It holds the logger, config, and output file, and ensures thread-safe writes.
type Handler struct {
	config *types.AccessLog   // Reuse the HTTP access log config struct for consistency.
	logger *logrus.Logger     // The logger instance used to write log entries.
	file   io.WriteCloser     // The file or output stream for logs.
	mu     sync.Mutex         // Mutex to ensure logs are written atomically.
}

// NewHandler creates a new TCP access log handler.
// It sets up the log file, format, and logger, reusing config from HTTP.
func NewHandler(config *types.AccessLog) (*Handler, error) {
	var file io.WriteCloser = os.Stdout // Default to stdout if no file is specified.
	if len(config.FilePath) > 0 {
		f, err := os.OpenFile(config.FilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o664)
		if err != nil {
			return nil, fmt.Errorf("error opening access log file: %w", err)
		}
		file = f
	}

	// Choose log format: JSON or plain text.
	var formatter logrus.Formatter
	switch config.Format {
	case "json":
		formatter = new(logrus.JSONFormatter)
	default:
		formatter = new(logrus.TextFormatter)
	}

	// Set up the logger.
	logger := &logrus.Logger{
		Out:       file,
		Formatter: formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}

	return &Handler{
		config: config,
		logger: logger,
		file:   file,
	}, nil
}

// LogConnectionStart logs when a new TCP connection is accepted.
// - clientAddr: the remote address of the client
// - serverAddr: the local address of the server (Traefik's listening address)
func (h *Handler) LogConnectionStart(ctx context.Context, clientAddr, serverAddr string, tlsState *tls.ConnectionState) {
	fields := logrus.Fields{
		"event":       "connection_start",         // Mark this as a connection start event
		"client_addr": clientAddr,                 // Remote client address
		"server_addr": serverAddr,                 // Local server address
		"timestamp":   time.Now().Format(time.RFC3339Nano), // Log the precise time
	}

	if tlsState != nil && len(tlsState.PeerCertificates) > 0 {
		cert := tlsState.PeerCertificates[0]
		buf := new(bytes.Buffer)
		err := pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		if err == nil {
			fields["client_cert_pem"] = buf.String()
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.logger.WithFields(fields).Println()
}

// LogConnectionEnd logs when a TCP connection is closed.
// - clientAddr: the remote address of the client
// - serverAddr: the local address of the server
// - bytesIn:    bytes received from the client
// - bytesOut:   bytes sent to the client
// - duration:   how long the connection was open
// - err:        any error that occurred, or nil for success
func (h *Handler) LogConnectionEnd(ctx context.Context, clientAddr, serverAddr string, bytesIn, bytesOut int64, duration time.Duration, err error, tlsState *tls.ConnectionState) {
	fields := logrus.Fields{
		"event":       "connection_end",
		"client_addr": clientAddr,
		"server_addr": serverAddr,
		"bytes_in":    bytesIn,
		"bytes_out":   bytesOut,
		"duration_ms": duration.Milliseconds(),
		"timestamp":   time.Now().Format(time.RFC3339Nano),
	}
	if err != nil {
		fields["error"] = err.Error()
		fields["status"] = "error"
	} else {
		fields["status"] = "success"
	}

	if tlsState != nil && len(tlsState.PeerCertificates) > 0 {
		cert := tlsState.PeerCertificates[0]
		buf := new(bytes.Buffer)
		err := pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		if err == nil {
			fields["client_cert_pem"] = buf.String()
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.logger.WithFields(fields).Println()
}
