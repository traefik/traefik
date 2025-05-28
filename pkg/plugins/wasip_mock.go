//go:build !linux && !darwin

package plugins

import (
	"context"

	"github.com/tetratelabs/wazero"
)

type ContextApplier func(ctx context.Context) context.Context

// InstantiateHost instantiates the Host module.
func InstantiateHost(ctx context.Context, runtime wazero.Runtime, mod wazero.CompiledModule, settings Settings) (ContextApplier, error) {
	return func(ctx context.Context) context.Context {
		return ctx
	}, nil
}
