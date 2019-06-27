package internal

import (
	"context"
	"strings"

	"github.com/containous/traefik/pkg/log"
)

type contextKey int

const (
	providerKey contextKey = iota
)

// AddProviderInContext Adds the provider name in the context
func AddProviderInContext(ctx context.Context, elementName string) context.Context {
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

// GetQualifiedName Gets the fully qualified name.
func GetQualifiedName(ctx context.Context, elementName string) string {
	parts := strings.Split(elementName, "@")
	if len(parts) == 1 {
		if providerName, ok := ctx.Value(providerKey).(string); ok {
			return MakeQualifiedName(providerName, parts[0])
		}
	}
	return elementName
}

// MakeQualifiedName Creates a qualified name for an element
func MakeQualifiedName(providerName string, elementName string) string {
	return elementName + "@" + providerName
}
