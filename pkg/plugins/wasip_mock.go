//go:build !linux && !darwin

package plugins

import (
	"context"

	"github.com/samyfodil/wazy"
)

type ContextApplier func(ctx context.Context) context.Context

// InstantiateHost instantiates the Host module.
func InstantiateHost(ctx context.Context, runtime wazy.Runtime, mod wazy.CompiledModule, settings Settings) (ContextApplier, error) {
	return func(ctx context.Context) context.Context {
		return ctx
	}, nil
}
