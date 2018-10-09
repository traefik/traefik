package responsemodifiers

import (
	"context"
	"net/http"

	"github.com/containous/traefik/config"
)

// NewBuilder creates a builder.
func NewBuilder(configs map[string]*config.Middleware) *Builder {
	return &Builder{configs: configs}
}

// Builder holds builder configuration.
type Builder struct {
	configs map[string]*config.Middleware
}

// Build Builds the response modifier.
func (f *Builder) Build(ctx context.Context, names []string) func(*http.Response) error {
	var modifiers []func(*http.Response) error

	for _, middleName := range names {
		if conf, ok := f.configs[middleName]; ok {
			if conf.Headers != nil {
				getLogger(ctx, middleName, "Headers").Debug("Creating Middleware (ResponseModifier)")

				modifiers = append(modifiers, buildHeaders(conf.Headers))
			} else if conf.Chain != nil {
				getLogger(ctx, middleName, "Chain").Debug("Creating Middleware (ResponseModifier)")

				modifiers = append(modifiers, f.Build(ctx, conf.Chain.Middlewares))
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

	return func(response *http.Response) error { return nil }
}
