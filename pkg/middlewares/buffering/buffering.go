package buffering

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	oxybuffer "github.com/vulcand/oxy/v2/buffer"
)

const (
	typeName = "Buffer"
)

type buffer struct {
	name                 string
	buffer               *oxybuffer.Buffer
	maxBodyBytes         int64
	disableRequestBuffer bool
}

// New creates a buffering middleware.
func New(ctx context.Context, next http.Handler, config dynamic.Buffering, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")
	logger.Debug().Msgf("Setting up buffering: request limits: %d (mem), %d (max), response limits: %d (mem), %d (max) with retry: '%s'",
		config.MemRequestBodyBytes, config.MaxRequestBodyBytes, config.MemResponseBodyBytes, config.MaxResponseBodyBytes, config.RetryExpression)

	options := []oxybuffer.Option{
		oxybuffer.MemRequestBodyBytes(config.MemRequestBodyBytes),
		oxybuffer.MaxRequestBodyBytes(config.MaxRequestBodyBytes),
		oxybuffer.MemResponseBodyBytes(config.MemResponseBodyBytes),
		oxybuffer.MaxResponseBodyBytes(config.MaxResponseBodyBytes),
		oxybuffer.Logger(logs.NewOxyWrapper(*logger)),
		oxybuffer.Verbose(logger.GetLevel() == zerolog.TraceLevel),
		oxybuffer.Cond(len(config.RetryExpression) > 0, oxybuffer.Retry(config.RetryExpression)),
	}

	if config.DisableRequestBuffer {
		options = append(options, oxybuffer.DisableRequestBuffer())
	}

	if config.DisableResponseBuffer {
		options = append(options, oxybuffer.DisableResponseBuffer())
	}

	oxyBuffer, err := oxybuffer.New(
		next,
		options...,
	)
	if err != nil {
		return nil, err
	}

	return &buffer{
		name:                 name,
		buffer:               oxyBuffer,
		maxBodyBytes:         config.MaxRequestBodyBytes,
		disableRequestBuffer: config.DisableRequestBuffer,
	}, nil
}

func (b *buffer) GetTracingInformation() (string, string) {
	return b.name, typeName
}

func (b *buffer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if b.disableRequestBuffer && b.maxBodyBytes > 0 && req.Body != nil && req.Body != http.NoBody {
		// Reject immediately when Content-Length is known and exceeds the limit.
		if req.ContentLength > b.maxBodyBytes {
			http.Error(rw, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
			return
		}
		// For streaming requests (chunked or unknown length), wrap the body so the
		// limit is enforced as bytes are read, matching NGINX's client_max_body_size
		// behavior without enabling request buffering.
		req.Body = http.MaxBytesReader(rw, req.Body, b.maxBodyBytes)
	}
	b.buffer.ServeHTTP(rw, req)
}
