// Package headers Middleware based on https://github.com/unrolled/secure.
package headers

import (
	"context"
	"errors"
	"net/http"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/middlewares/connectionheader"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

const (
	typeName = "Headers"
)

func handleDeprecation(ctx context.Context, cfg *dynamic.Headers) {
	logger := log.Ctx(ctx).Warn()

	if cfg.SSLRedirect {
		logger.Msg("SSLRedirect is deprecated, please use entrypoint redirection instead.")
	}
	if cfg.SSLTemporaryRedirect {
		logger.Msg("SSLTemporaryRedirect is deprecated, please use entrypoint redirection instead.")
	}
	if cfg.SSLHost != "" {
		logger.Msg("SSLHost is deprecated, please use RedirectRegex middleware instead.")
	}
	if cfg.SSLForceHost {
		logger.Msg("SSLForceHost is deprecated, please use RedirectScheme middleware instead.")
	}
	if cfg.FeaturePolicy != "" {
		logger.Msg("FeaturePolicy is deprecated, please use PermissionsPolicy header instead.")
	}
}

type headers struct {
	name    string
	handler http.Handler
}

// New creates a Headers middleware.
func New(ctx context.Context, next http.Handler, cfg dynamic.Headers, name string) (http.Handler, error) {
	// HeaderMiddleware -> SecureMiddleWare -> next
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	mCtx := logger.WithContext(ctx)

	handleDeprecation(mCtx, &cfg)

	hasSecureHeaders := cfg.HasSecureHeadersDefined()
	hasCustomHeaders := cfg.HasCustomHeadersDefined()
	hasCorsHeaders := cfg.HasCorsHeadersDefined()

	if !hasSecureHeaders && !hasCustomHeaders && !hasCorsHeaders {
		return nil, errors.New("headers configuration not valid")
	}

	var handler http.Handler
	nextHandler := next

	if hasSecureHeaders {
		logger.Debug().Msgf("Setting up secureHeaders from %v", cfg)
		handler = newSecure(next, cfg, name)
		nextHandler = handler
	}

	if hasCustomHeaders || hasCorsHeaders {
		logger.Debug().Msgf("Setting up customHeaders/Cors from %v", cfg)
		h, err := NewHeader(nextHandler, cfg)
		if err != nil {
			return nil, err
		}

		handler = connectionheader.Remover(h)
	}

	return &headers{
		handler: handler,
		name:    name,
	}, nil
}

func (h *headers) GetTracingInformation() (string, ext.SpanKindEnum) {
	return h.name, tracing.SpanKindNoneEnum
}

func (h *headers) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.handler.ServeHTTP(rw, req)
}
