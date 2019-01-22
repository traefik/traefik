package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/containous/alice"
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/middlewares/addprefix"
	"github.com/containous/traefik/middlewares/auth"
	"github.com/containous/traefik/middlewares/buffering"
	"github.com/containous/traefik/middlewares/chain"
	"github.com/containous/traefik/middlewares/circuitbreaker"
	"github.com/containous/traefik/middlewares/compress"
	"github.com/containous/traefik/middlewares/customerrors"
	"github.com/containous/traefik/middlewares/headers"
	"github.com/containous/traefik/middlewares/ipwhitelist"
	"github.com/containous/traefik/middlewares/maxconnection"
	"github.com/containous/traefik/middlewares/passtlsclientcert"
	"github.com/containous/traefik/middlewares/ratelimiter"
	"github.com/containous/traefik/middlewares/redirect"
	"github.com/containous/traefik/middlewares/replacepath"
	"github.com/containous/traefik/middlewares/replacepathregex"
	"github.com/containous/traefik/middlewares/retry"
	"github.com/containous/traefik/middlewares/stripprefix"
	"github.com/containous/traefik/middlewares/stripprefixregex"
	"github.com/containous/traefik/middlewares/tracing"
	"github.com/containous/traefik/server/internal"
	"github.com/pkg/errors"
)

type middlewareStackType int

const (
	middlewareStackKey middlewareStackType = iota
)

// Builder the middleware builder
type Builder struct {
	configs        map[string]*config.Middleware
	serviceBuilder serviceBuilder
}

type serviceBuilder interface {
	Build(ctx context.Context, serviceName string, responseModifier func(*http.Response) error) (http.Handler, error)
}

// NewBuilder creates a new Builder
func NewBuilder(configs map[string]*config.Middleware, serviceBuilder serviceBuilder) *Builder {
	return &Builder{configs: configs, serviceBuilder: serviceBuilder}
}

// BuildChain creates a middleware chain
func (b *Builder) BuildChain(ctx context.Context, middlewares []string) *alice.Chain {
	chain := alice.New()
	for _, middlewareName := range middlewares {
		middlewareName := internal.GetQualifiedName(ctx, middlewareName)
		constructorContext := internal.AddProviderInContext(ctx, middlewareName)

		chain = chain.Append(func(next http.Handler) (http.Handler, error) {
			if _, ok := b.configs[middlewareName]; !ok {
				return nil, fmt.Errorf("middleware %q does not exist", middlewareName)
			}

			var err error
			if constructorContext, err = checkRecursivity(constructorContext, middlewareName); err != nil {
				return nil, err
			}

			constructor, err := b.buildConstructor(constructorContext, middlewareName, *b.configs[middlewareName])
			if err != nil {
				return nil, fmt.Errorf("error during instanciation of %s: %v", middlewareName, err)
			}
			return constructor(next)
		})
	}
	return &chain
}

func checkRecursivity(ctx context.Context, middlewareName string) (context.Context, error) {
	currentStack, ok := ctx.Value(middlewareStackKey).([]string)
	if !ok {
		currentStack = []string{}
	}
	if inSlice(middlewareName, currentStack) {
		return ctx, fmt.Errorf("could not instantiate middleware %s: recursion detected in %s", middlewareName, strings.Join(append(currentStack, middlewareName), "->"))
	}
	return context.WithValue(ctx, middlewareStackKey, append(currentStack, middlewareName)), nil
}

func (b *Builder) buildConstructor(ctx context.Context, middlewareName string, config config.Middleware) (alice.Constructor, error) {
	var middleware alice.Constructor
	badConf := errors.New("cannot create middleware %q: multi-types middleware not supported, consider declaring two different pieces of middleware instead")

	// AddPrefix
	if config.AddPrefix != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return addprefix.New(ctx, next, *config.AddPrefix, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// BasicAuth
	if config.BasicAuth != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return auth.NewBasic(ctx, next, *config.BasicAuth, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// Buffering
	if config.Buffering != nil && config.MaxConn.Amount != 0 {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return buffering.New(ctx, next, *config.Buffering, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// Chain
	if config.Chain != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return chain.New(ctx, next, *config.Chain, b, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// CircuitBreaker
	if config.CircuitBreaker != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return circuitbreaker.New(ctx, next, *config.CircuitBreaker, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// Compress
	if config.Compress != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return compress.New(ctx, next, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// CustomErrors
	if config.Errors != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return customerrors.New(ctx, next, *config.Errors, b.serviceBuilder, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// DigestAuth
	if config.DigestAuth != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return auth.NewDigest(ctx, next, *config.DigestAuth, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// ForwardAuth
	if config.ForwardAuth != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return auth.NewForward(ctx, next, *config.ForwardAuth, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// Headers
	if config.Headers != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return headers.New(ctx, next, *config.Headers, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// IPWhiteList
	if config.IPWhiteList != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return ipwhitelist.New(ctx, next, *config.IPWhiteList, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// MaxConn
	if config.MaxConn != nil && config.MaxConn.Amount != 0 {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return maxconnection.New(ctx, next, *config.MaxConn, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// PassTLSClientCert
	if config.PassTLSClientCert != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return passtlsclientcert.New(ctx, next, *config.PassTLSClientCert, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// RateLimit
	if config.RateLimit != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return ratelimiter.New(ctx, next, *config.RateLimit, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// RedirectRegex
	if config.RedirectRegex != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return redirect.NewRedirectRegex(ctx, next, *config.RedirectRegex, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// RedirectScheme
	if config.RedirectScheme != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return redirect.NewRedirectScheme(ctx, next, *config.RedirectScheme, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// ReplacePath
	if config.ReplacePath != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return replacepath.New(ctx, next, *config.ReplacePath, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// ReplacePathRegex
	if config.ReplacePathRegex != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return replacepathregex.New(ctx, next, *config.ReplacePathRegex, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// Retry
	if config.Retry != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				// FIXME missing metrics / accessLog
				return retry.New(ctx, next, *config.Retry, retry.Listeners{}, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// StripPrefix
	if config.StripPrefix != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return stripprefix.New(ctx, next, *config.StripPrefix, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	// StripPrefixRegex
	if config.StripPrefixRegex != nil {
		if middleware == nil {
			middleware = func(next http.Handler) (http.Handler, error) {
				return stripprefixregex.New(ctx, next, *config.StripPrefixRegex, middlewareName)
			}
		} else {
			return nil, badConf
		}
	}

	if middleware == nil {
		return nil, fmt.Errorf("middleware %q does not exist", middlewareName)
	}

	return tracing.Wrap(ctx, middleware), nil
}

func inSlice(element string, stack []string) bool {
	for _, value := range stack {
		if value == element {
			return true
		}
	}
	return false
}
