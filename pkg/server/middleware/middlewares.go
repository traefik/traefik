package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/containous/alice"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/middlewares/addprefix"
	"github.com/traefik/traefik/v3/pkg/middlewares/auth"
	"github.com/traefik/traefik/v3/pkg/middlewares/buffering"
	"github.com/traefik/traefik/v3/pkg/middlewares/chain"
	"github.com/traefik/traefik/v3/pkg/middlewares/circuitbreaker"
	"github.com/traefik/traefik/v3/pkg/middlewares/compress"
	"github.com/traefik/traefik/v3/pkg/middlewares/contenttype"
	"github.com/traefik/traefik/v3/pkg/middlewares/customerrors"
	"github.com/traefik/traefik/v3/pkg/middlewares/gatewayapi/headermodifier"
	gapiredirect "github.com/traefik/traefik/v3/pkg/middlewares/gatewayapi/redirect"
	"github.com/traefik/traefik/v3/pkg/middlewares/gatewayapi/urlrewrite"
	"github.com/traefik/traefik/v3/pkg/middlewares/grpcweb"
	"github.com/traefik/traefik/v3/pkg/middlewares/headers"
	"github.com/traefik/traefik/v3/pkg/middlewares/inflightreq"
	"github.com/traefik/traefik/v3/pkg/middlewares/ipallowlist"
	"github.com/traefik/traefik/v3/pkg/middlewares/ipwhitelist"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/middlewares/passtlsclientcert"
	"github.com/traefik/traefik/v3/pkg/middlewares/ratelimiter"
	"github.com/traefik/traefik/v3/pkg/middlewares/redirect"
	"github.com/traefik/traefik/v3/pkg/middlewares/replacepath"
	"github.com/traefik/traefik/v3/pkg/middlewares/replacepathregex"
	"github.com/traefik/traefik/v3/pkg/middlewares/retry"
	"github.com/traefik/traefik/v3/pkg/middlewares/stripprefix"
	"github.com/traefik/traefik/v3/pkg/middlewares/stripprefixregex"
	"github.com/traefik/traefik/v3/pkg/server/provider"
)

type middlewareStackType int

const (
	middlewareStackKey middlewareStackType = iota
)

// Builder the middleware builder.
type Builder struct {
	configs        map[string]*runtime.MiddlewareInfo
	pluginBuilder  PluginsBuilder
	serviceBuilder serviceBuilder
}

type serviceBuilder interface {
	BuildHTTP(ctx context.Context, serviceName string) (http.Handler, error)
}

// NewBuilder creates a new Builder.
func NewBuilder(configs map[string]*runtime.MiddlewareInfo, serviceBuilder serviceBuilder, pluginBuilder PluginsBuilder) *Builder {
	return &Builder{configs: configs, serviceBuilder: serviceBuilder, pluginBuilder: pluginBuilder}
}

// BuildChain creates a middleware chain.
func (b *Builder) BuildChain(ctx context.Context, middlewares []string) *alice.Chain {
	chain := alice.New()
	for _, name := range middlewares {
		middlewareName := provider.GetQualifiedName(ctx, name)

		chain = chain.Append(func(next http.Handler) (http.Handler, error) {
			constructorContext := provider.AddInContext(ctx, middlewareName)
			if midInf, ok := b.configs[middlewareName]; !ok || midInf.Middleware == nil {
				return nil, fmt.Errorf("middleware %q does not exist", middlewareName)
			}

			var err error
			if constructorContext, err = checkRecursion(constructorContext, middlewareName); err != nil {
				b.configs[middlewareName].AddError(err, true)
				return nil, err
			}

			constructor, err := b.buildConstructor(constructorContext, middlewareName)
			if err != nil {
				b.configs[middlewareName].AddError(err, true)
				return nil, err
			}

			handler, err := constructor(next)
			if err != nil {
				b.configs[middlewareName].AddError(err, true)
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
	if slices.Contains(currentStack, middlewareName) {
		return ctx, fmt.Errorf("could not instantiate middleware %s: recursion detected in %s", middlewareName, strings.Join(append(currentStack, middlewareName), "->"))
	}
	return context.WithValue(ctx, middlewareStackKey, append(currentStack, middlewareName)), nil
}

// it is the responsibility of the caller to make sure that b.configs[middlewareName].Middleware exists.
func (b *Builder) buildConstructor(ctx context.Context, middlewareName string) (alice.Constructor, error) {
	config := b.configs[middlewareName]
	if config == nil || config.Middleware == nil {
		return nil, fmt.Errorf("invalid middleware %q configuration", middlewareName)
	}

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
	if config.Buffering != nil {
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

		var qualifiedNames []string
		for _, name := range config.Chain.Middlewares {
			qualifiedNames = append(qualifiedNames, provider.GetQualifiedName(ctx, name))
		}
		config.Chain.Middlewares = qualifiedNames
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
			return compress.New(ctx, next, *config.Compress, middlewareName)
		}
	}

	// ContentType
	if config.ContentType != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return contenttype.New(ctx, next, *config.ContentType, middlewareName)
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

	// GrpcWeb
	if config.GrpcWeb != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return grpcweb.New(ctx, next, *config.GrpcWeb, middlewareName), nil
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
		qualifiedName := provider.GetQualifiedName(ctx, middlewareName)
		log.Warn().Msgf("Middleware %q of type IPWhiteList is deprecated, please use IPAllowList instead.", qualifiedName)

		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return ipwhitelist.New(ctx, next, *config.IPWhiteList, middlewareName)
		}
	}

	// IPAllowList
	if config.IPAllowList != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return ipallowlist.New(ctx, next, *config.IPAllowList, middlewareName)
		}
	}

	// InFlightReq
	if config.InFlightReq != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return inflightreq.New(ctx, next, *config.InFlightReq, middlewareName)
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

	// RateLimit
	if config.RateLimit != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return ratelimiter.New(ctx, next, *config.RateLimit, middlewareName)
		}
	}

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
			// TODO missing metrics / accessLog
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

	// Plugin
	if config.Plugin != nil && !reflect.ValueOf(b.pluginBuilder).IsNil() { // Using "reflect" because "b.pluginBuilder" is an interface.
		if middleware != nil {
			return nil, badConf
		}

		pluginType, rawPluginConfig, err := findPluginConfig(config.Plugin)
		if err != nil {
			return nil, fmt.Errorf("plugin: %w", err)
		}

		plug, err := b.pluginBuilder.Build(pluginType, rawPluginConfig, middlewareName)
		if err != nil {
			return nil, fmt.Errorf("plugin: %w", err)
		}

		middleware = func(next http.Handler) (http.Handler, error) {
			return newTraceablePlugin(ctx, middlewareName, plug, next)
		}
	}

	// Gateway API HTTPRoute filters middlewares.
	if config.RequestHeaderModifier != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return headermodifier.NewRequestHeaderModifier(ctx, next, *config.RequestHeaderModifier, middlewareName)
		}
	}

	if config.RequestRedirect != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return gapiredirect.NewRequestRedirect(ctx, next, *config.RequestRedirect, middlewareName)
		}
	}

	if config.URLRewrite != nil {
		if middleware != nil {
			return nil, badConf
		}
		middleware = func(next http.Handler) (http.Handler, error) {
			return urlrewrite.NewURLRewrite(ctx, next, *config.URLRewrite, middlewareName)
		}
	}

	if middleware == nil {
		return nil, fmt.Errorf("invalid middleware %q configuration: invalid middleware type or middleware does not exist", middlewareName)
	}

	// The tracing middleware is a NOOP if tracing is not setup on the middleware chain.
	// Hence, regarding internal resources' observability deactivation,
	// this would not enable tracing.
	return observability.WrapMiddleware(ctx, middleware), nil
}
