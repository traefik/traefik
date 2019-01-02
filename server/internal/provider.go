package internal

import (
	"context"
	"strings"

	"github.com/containous/traefik/log"
)

type contextKey int

const (
	providerKey contextKey = iota
)

// CreateProviderContext creates a context with the provider name extracted from elementName if specified
//  if not, returns the given context and the fully qualified elementName (prefixed with the provider name)
//  in case there is no context in the provider and no context: logs a warning
func CreateProviderContext(ctx context.Context, elementName string) (contextWithProvider context.Context, fullyQualifiedElementName string) {
	if providerName := getProviderName(elementName); len(providerName) > 0 {
		return AddProviderInContext(ctx, providerName), elementName
	} else {
		if providerName, ok := ctx.Value(providerKey).(string); ok {
			return ctx, providerName + "." + elementName
		}
	}

	log.FromContext(ctx).Debugf("Could not find a provider for %s.", elementName)
	return ctx, elementName
}

func getProviderName(middleware string) string {
	parts := strings.Split(middleware, ".")
	if len(parts) == 1 {
		return ""
	}
	return parts[0]
}

func AddProviderInContext(ctx context.Context, providerName string) context.Context {
	return context.WithValue(ctx, providerKey, providerName)
}
