package buffering

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/logs"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
	oxybuffer "github.com/vulcand/oxy/v2/buffer"
)

const (
	typeName = "Buffer"
)

type buffer struct {
	name   string
	buffer *oxybuffer.Buffer
}

// New creates a buffering middleware.
func New(ctx context.Context, next http.Handler, config dynamic.Buffering, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")
	logger.Debug().Msgf("Setting up buffering: request limits: %d (mem), %d (max), response limits: %d (mem), %d (max) with retry: '%s'",
		config.MemRequestBodyBytes, config.MaxRequestBodyBytes, config.MemResponseBodyBytes, config.MaxResponseBodyBytes, config.RetryExpression)

	oxyBuffer, err := oxybuffer.New(
		next,
		oxybuffer.MemRequestBodyBytes(config.MemRequestBodyBytes),
		oxybuffer.MaxRequestBodyBytes(config.MaxRequestBodyBytes),
		oxybuffer.MemResponseBodyBytes(config.MemResponseBodyBytes),
		oxybuffer.MaxResponseBodyBytes(config.MaxResponseBodyBytes),
		oxybuffer.Logger(logs.NewOxyWrapper(*logger)),
		oxybuffer.Verbose(logger.GetLevel() == zerolog.TraceLevel),
		oxybuffer.Cond(len(config.RetryExpression) > 0, oxybuffer.Retry(config.RetryExpression)),
	)
	if err != nil {
		return nil, err
	}

	return &buffer{
		name:   name,
		buffer: oxyBuffer,
	}, nil
}

func (b *buffer) GetTracingInformation() (string, ext.SpanKindEnum) {
	return b.name, tracing.SpanKindNoneEnum
}

func (b *buffer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	b.buffer.ServeHTTP(rw, req)
}
