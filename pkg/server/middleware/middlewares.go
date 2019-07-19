package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/containous/alice"
	"github.com/containous/traefik/pkg/config/runtime"
	"github.com/containous/traefik/pkg/middlewares/addprefix"
	"github.com/containous/traefik/pkg/middlewares/auth"
	"github.com/containous/traefik/pkg/middlewares/buffering"
	"github.com/containous/traefik/pkg/middlewares/chain"
	"github.com/containous/traefik/pkg/middlewares/circuitbreaker"
	"github.com/containous/traefik/pkg/middlewares/compress"
	"github.com/containous/traefik/pkg/middlewares/customerrors"
	"github.com/containous/traefik/pkg/middlewares/headers"
	"github.com/containous/traefik/pkg/middlewares/ipwhitelist"
	"github.com/containous/traefik/pkg/middlewares/maxconnection"
	"github.com/containous/traefik/pkg/middlewares/passtlsclientcert"
	"github.com/containous/traefik/pkg/middlewares/redirect"
	"github.com/containous/traefik/pkg/middlewares/replacepath"
	"github.com/containous/traefik/pkg/middlewares/replacepathregex"
	"github.com/containous/traefik/pkg/middlewares/retry"
	"github.com/containous/traefik/pkg/middlewares/stripprefix"
	"github.com/containous/traefik/pkg/middlewares/stripprefixregex"
	"github.com/containous/traefik/pkg/middlewares/tracing"
	"github.com/containous/traefik/pkg/server/internal"
)

type middlewareStackType int

const (
	middlewareStackKey middlewareStackType = iota
)

// Builder the middleware builder
type Builder struct {
	configs        map[string]*runtime.MiddlewareInfo
	serviceBuilder serviceBuilder
}

type serviceBuilder interface {
	BuildHTTP(ctx context.Context, serviceName string, responseModifier func(*http.Response) error) (http.Handler, error)
}

// NewBuilder creates a new Builder
func NewBuilder(configs map[string]*runtime.MiddlewareInfo, serviceBuilder serviceBuilder) *Builder {
	return &Builder{configs: configs, serviceBuilder: serviceBuilder}
}

// BuildChain creates a middleware chain
func (b *Builder) BuildChain(ctx context.Context, middlewares []string) *alice.Chain {
	chain := alice.New()
	for _, name := range middlewares {
		middlewareName := internal.GetQualifiedName(ctx, name)

		chain = chain.Append(func(next http.Handler) (http.Handler, error) {
			constructorContext := internal.AddProviderInContext(ctx, middlewareName)
			if midInf, ok := b.configs[middlewareName]; !ok || midInf.Middleware == nil {
				return nil, fmt.Errorf("middleware %q does not exist", middlewareName)
			}

			var err error
			if constructorContext, err = checkRecursion(constructorContext, middlewareName); err != nil {
				b.configs[middlewareName].AddError(err)
				return nil, err
			}

			constructor, err := b.buildConstructor(constructorContext, middlewareName)
			if err != nil {
				b.configs[middlewareName].AddError(err)
				return nil, err
			}

			handler, err := constructor(next)
			if err != nil {
				b.configs[middlewareName].AddError(err)
				return nil, err
			}

			return handler, nil
		})
	}
	return &chain
}

func checkRecursion(ctx context.Context, middlewareName string) (context.Context, error) {
	currentStack, ok := ctx.Value(middlewareStackKey).([]string)
	if !ok {
		currentStack = []string{}
	}
	if inSlice(middlewareName, currentStack) {
		return ctx, fmt.Errorf("could not instantiate middleware %s: recursion detected in %s", middlewareName, strings.Join(append(currentStack, middlewareName), "->"))
	}
	return context.WithValue(ctx, middlewareStackKey, append(currentStack, middlewareName)), nil
}

// it is the responsibility of the caller to make sure that b.configs[middlewareName].Middleware exists
func (b *Builder) buildConstructor(ctx context.Context, middlewareName string) (alice.Constructor, error) {
	config := b.configs[middlewareName]
	var middleware alice.Constructor
	badConf := errors.New("cannot create middleware: multi-types middleware not supported, consider declaring two different pieces of middleware instead")

	// AddPrefix
	if config.AddPrefix != nil {
		middleware = func(next http.Handler) (http.Handler, error) {
			return addprefix.New(ctx, next, *config.AddPrefix, middlewareName)
		}
	}

	// BasicAuth
	if config.BasicAuth != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return auth.NewBasic(ctx, next, *config.BasicAuth, middlewareName)
		}
	}

	// Buffering
	if config.Buffering != nil && config.MaxConn.Amount != 0 {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return buffering.New(ctx, next, *config.Buffering, middlewareName)
		}
	}

	// Chain
	if config.Chain != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return chain.New(ctx, next, *config.Chain, b, middlewareName)
		}
	}

	// CircuitBreaker
	if config.CircuitBreaker != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return circuitbreaker.New(ctx, next, *config.CircuitBreaker, middlewareName)
		}
	}

	// Compress
	if config.Compress != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return compress.New(ctx, next, middlewareName)
		}
	}

	// CustomErrors
	if config.Errors != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return customerrors.New(ctx, next, *config.Errors, b.serviceBuilder, middlewareName)
		}
	}

	// DigestAuth
	if config.DigestAuth != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return auth.NewDigest(ctx, next, *config.DigestAuth, middlewareName)
		}
	}

	// ForwardAuth
	if config.ForwardAuth != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return auth.NewForward(ctx, next, *config.ForwardAuth, middlewareName)
		}
	}

	// Headers
	if config.Headers != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return headers.New(ctx, next, *config.Headers, middlewareName)
		}
	}

	// IPWhiteList
	if config.IPWhiteList != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return ipwhitelist.New(ctx, next, *config.IPWhiteList, middlewareName)
		}
	}

	// MaxConn
	if config.MaxConn != nil && config.MaxConn.Amount != 0 {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return maxconnection.New(ctx, next, *config.MaxConn, middlewareName)
		}
	}

	// PassTLSClientCert
	if config.PassTLSClientCert != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return passtlsclientcert.New(ctx, next, *config.PassTLSClientCert, middlewareName)
		}
	}

	// TODO: disable temporarily (rateLimit)
	// RateLimit
	// if config.RateLimit != nil {
	// 	if middleware != nil {
	// 		return nil, badConf
	// 	}
	// 	middleware = func(next http.Handler) (http.Handler, error) {
	// 		return ratelimiter.New(ctx, next, *config.RateLimit, middlewareName)
	// 	}
	// }

	// RedirectRegex
	if config.RedirectRegex != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return redirect.NewRedirectRegex(ctx, next, *config.RedirectRegex, middlewareName)
		}
	}

	// RedirectScheme
	if config.RedirectScheme != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return redirect.NewRedirectScheme(ctx, next, *config.RedirectScheme, middlewareName)
		}
	}

	// ReplacePath
	if config.ReplacePath != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return replacepath.New(ctx, next, *config.ReplacePath, middlewareName)
		}
	}

	// ReplacePathRegex
	if config.ReplacePathRegex != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return replacepathregex.New(ctx, next, *config.ReplacePathRegex, middlewareName)
		}
	}

	// Retry
	if config.Retry != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			// FIXME missing metrics / accessLog
			return retry.New(ctx, next, *config.Retry, retry.Listeners{}, middlewareName)
		}
	}

	// StripPrefix
	if config.StripPrefix != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return stripprefix.New(ctx, next, *config.StripPrefix, middlewareName)
		}
	}

	// StripPrefixRegex
	if config.StripPrefixRegex != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return stripprefixregex.New(ctx, next, *config.StripPrefixRegex, middlewareName)
		}
	}

	if middleware == nil {
		return nil, errors.New("middleware does not exist")
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
