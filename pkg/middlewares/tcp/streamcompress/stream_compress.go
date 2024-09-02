package tcpstreamcompress

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

const (
	typeName = "TCPStreamCompress"
)

// streamCompress is a middleware that provides compression on TCP streams
type streamCompress struct {
	next      tcp.Handler
	algorithm string
	name      string
	dict      []byte
	level     int
}

// New builds a new TCP StreamCompress
func New(ctx context.Context, next tcp.Handler, config dynamic.TCPStreamCompress, name string) (tcp.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msgf("Creating middleware")

	switch config.Algorithm {
	case "zstd":
		// success
	default:
		return nil, errors.New(fmt.Sprintf("unknown compression algorithm %s", config.Algorithm))
	}

	s := &streamCompress{
		algorithm: config.Algorithm,
		next:      next,
		name:      name,
		level:     config.Level,
	}
	if config.Dictionary != "" {
		var err error
		// Attempt to read the dictionary from the specified file
		s.dict, err = ioutil.ReadFile(config.Dictionary)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to read dictionary file %s: %v", config.Dictionary, err))
		}
	}
	logger.Debug().Msgf("Setting up TCP Stream compression with algorithm: %s", config.Algorithm)

	return s, nil
}

func (s *streamCompress) ServeTCP(conn tcp.WriteCloser) {
	/*ctx := middlewares.GetLoggerCtx(context.Background(), s.name, typeName)
	logger := log.FromContext(ctx)

	addr := conn.RemoteAddr().String()

	err := s.whiteLister.IsAuthorized(addr)
	if err != nil {
		logger.Errorf("Connection from %s rejected: %v", addr, err)
		conn.Close()
		return
	}

	logger.Debugf("Connection from %s accepted", addr)
	*/

	// Wapper the connection with a compression algorithm

	// For Testing, wrapper with compress + decompress to show that all aspects work correctly. IE it should be plain in and plain out
	conn = NewZStdCompressor(conn, s.level, s.dict)
	conn = NewZStdDecompressor(conn, s.level, s.dict)
	conn = NewZStdCompressor(conn, s.level, s.dict)
	conn = NewZStdDecompressor(conn, s.level, s.dict)

	s.next.ServeTCP(conn)
}
