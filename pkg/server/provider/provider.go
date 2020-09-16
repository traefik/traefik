package provider

import (
	"context"
	"strings"

	"github.com/traefik/traefik/v2/pkg/log"
)

type contextKey int

const (
	key contextKey = iota
)

// AddInContext Adds the provider name in the context.
func AddInContext(ctx context.Context, elementName string) context.Context {
	parts := strings.Split(elementName, "@")
	if len(parts) == 1 {
		log.FromContext(ctx).Debugf("Could not find a provider for %s.", elementName)
		return ctx
	}

	if name, ok := ctx.Value(key).(string); ok && name == parts[1] {
		return ctx
	}

	return context.WithValue(ctx, key, parts[1])
}

// GetQualifiedName Gets the fully qualified name.
func GetQualifiedName(ctx context.Context, elementName string) string {
	parts := strings.Split(elementName, "@")
	if len(parts) == 1 {
		if providerName, ok := ctx.Value(key).(string); ok {
			return MakeQualifiedName(providerName, parts[0])
		}
	}
	return elementName
}

// MakeQualifiedName Creates a qualified name for an element.
func MakeQualifiedName(providerName, elementName string) string {
	return elementName + "@" + providerName
}
