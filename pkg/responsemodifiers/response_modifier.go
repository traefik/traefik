package responsemodifiers

import (
	"context"
	"net/http"

	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/containous/traefik/v2/pkg/server/provider"
)

// NewBuilder creates a builder.
func NewBuilder(configs map[string]*runtime.MiddlewareInfo) *Builder {
	return &Builder{configs: configs}
}

// Builder holds builder configuration.
type Builder struct {
	configs map[string]*runtime.MiddlewareInfo
}

// Build Builds the response modifier.
// It returns nil if there is no modifier to apply.
func (f *Builder) Build(ctx context.Context, names []string) func(*http.Response) error {
	var modifiers []func(*http.Response) error

	for _, middleName := range names {
		conf, ok := f.configs[middleName]
		if !ok {
			getLogger(ctx, middleName, "undefined").Debug("Middleware name not found in config (ResponseModifier)")
			continue
		}
		if conf == nil || conf.Middleware == nil {
			getLogger(ctx, middleName, "undefined").Error("Invalid Middleware configuration (ResponseModifier)")
			continue
		}

		if conf.Headers != nil {
			getLogger(ctx, middleName, "Headers").Debug("Creating Middleware (ResponseModifier)")

			modifiers = append(modifiers, buildHeaders(conf.Headers, middleName))
		} else if conf.Chain != nil {
			chainCtx := provider.AddInContext(ctx, middleName)
			getLogger(chainCtx, middleName, "Chain").Debug("Creating Middleware (ResponseModifier)")
			var qualifiedNames []string
			for _, name := range conf.Chain.Middlewares {
				qualifiedNames = append(qualifiedNames, provider.GetQualifiedName(chainCtx, name))
			}

			if rm := f.Build(ctx, qualifiedNames); rm != nil {
				modifiers = append(modifiers, rm)
			}
		}
	}

	if len(modifiers) > 0 {
		return func(resp *http.Response) error {
			for i := len(modifiers); i > 0; i-- {
				err := modifiers[i-1](resp)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	return nil
}
