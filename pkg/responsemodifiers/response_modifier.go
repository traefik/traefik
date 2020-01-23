package responsemodifiers

import (
	"context"
	"net/http"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/containous/traefik/v2/pkg/log"
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
func (f *Builder) Build(ctx context.Context, names []string) func(*http.Response) error {
	var modifiers []func(*http.Response) error

	for _, middleName := range names {
		conf, ok := f.configs[middleName]
		if !ok {
			getLogger(ctx, middleName, "undefined").Warn("Middleware name not found in config (ResponseModifier)")
			continue
		}
		if conf == nil || conf.Middleware == nil {
			getLogger(ctx, middleName, "undefined").Error("Invalid Middleware configuration (ResponseModifier)")
			continue
		}

		if conf.Headers != nil {
			getLogger(ctx, middleName, "Headers").Debug("Creating Middleware (ResponseModifier)")

			modifiers = append(modifiers, buildHeaders(conf.Headers))
		} else if conf.Chain != nil {
			chainCtx := addProviderInContext(ctx, middleName)
			getLogger(chainCtx, middleName, "Chain").Debug("Creating Middleware (ResponseModifier)")
			var qualifiedNames []string
			for _, name := range conf.Chain.Middlewares {
				qualifiedNames = append(qualifiedNames, getQualifiedName(chainCtx, name))
			}
			modifiers = append(modifiers, f.Build(ctx, qualifiedNames))
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

type contextKey int

const (
	providerKey contextKey = iota
)

// addProviderInContext adds the provider name in the context
func addProviderInContext(ctx context.Context, elementName string) context.Context {
	parts := strings.Split(elementName, "@")
	if len(parts) == 1 {
		log.FromContext(ctx).Debugf("Could not find a provider for %s.", elementName)
		return ctx
	}

	if name, ok := ctx.Value(providerKey).(string); ok && name == parts[1] {
		return ctx
	}

	return context.WithValue(ctx, providerKey, parts[1])
}

// getQualifiedName gets the fully qualified name.
func getQualifiedName(ctx context.Context, elementName string) string {
	parts := strings.Split(elementName, "@")
	if len(parts) == 1 {
		if providerName, ok := ctx.Value(providerKey).(string); ok {
			return makeQualifiedName(providerName, parts[0])
		}
	}
	return elementName
}

// makeQualifiedName creates a qualified name for an element
func makeQualifiedName(providerName string, elementName string) string {
	return elementName + "@" + providerName
}
