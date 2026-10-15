package udpmiddleware

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/middlewares/accessgovernance"
	"github.com/traefik/traefik/v3/pkg/server/provider"
	"github.com/traefik/traefik/v3/pkg/udp"
)

type middlewareStackType int

const middlewareStackKey middlewareStackType = iota

// Builder builds UDP middleware chains.
type Builder struct {
	configs map[string]*runtime.UDPMiddlewareInfo
}

// NewBuilder creates a UDP middleware builder.
func NewBuilder(configs map[string]*runtime.UDPMiddlewareInfo) *Builder {
	return &Builder{configs: configs}
}

// BuildChain creates a UDP middleware chain.
func (b *Builder) BuildChain(ctx context.Context, middlewares []string) *udp.Chain {
	chain := udp.NewChain()
	for _, name := range middlewares {
		middlewareName := provider.GetQualifiedName(ctx, name)
		chain = chain.Append(func(next udp.Handler) (udp.Handler, error) {
			constructorContext := provider.AddInContext(ctx, middlewareName)
			info, ok := b.configs[middlewareName]
			if !ok || info.UDPMiddleware == nil {
				return nil, fmt.Errorf("middleware %q does not exist", middlewareName)
			}
			var err error
			constructorContext, err = checkRecursion(constructorContext, middlewareName)
			if err != nil {
				info.AddError(err, true)
				return nil, err
			}
			constructor, err := b.buildConstructor(constructorContext, middlewareName)
			if err != nil {
				info.AddError(err, true)
				return nil, err
			}
			handler, err := constructor(next)
			if err != nil {
				info.AddError(err, true)
				return nil, err
			}
			return handler, nil
		})
	}
	return &chain
}

func (b *Builder) buildConstructor(ctx context.Context, middlewareName string) (udp.Constructor, error) {
	config := b.configs[middlewareName]
	if config == nil || config.UDPMiddleware == nil {
		return nil, fmt.Errorf("invalid middleware %q configuration", middlewareName)
	}
	if config.AccessGovernance == nil {
		return nil, fmt.Errorf("invalid middleware %q configuration: invalid middleware type or middleware does not exist", middlewareName)
	}
	return func(next udp.Handler) (udp.Handler, error) {
		return accessgovernance.NewUDP(ctx, next, *config.AccessGovernance, middlewareName)
	}, nil
}

func checkRecursion(ctx context.Context, middlewareName string) (context.Context, error) {
	current, _ := ctx.Value(middlewareStackKey).([]string)
	if slices.Contains(current, middlewareName) {
		return ctx, fmt.Errorf("could not instantiate middleware %s: recursion detected in %s", middlewareName, strings.Join(append(current, middlewareName), "->"))
	}
	return context.WithValue(ctx, middlewareStackKey, append(current, middlewareName)), nil
}
