//go:build !linux && !darwin

package plugins

import (
	"context"

	"github.com/tetratelabs/wazero"
)

// ContextApplier binds host module state to ctx for a single unit of work
// (e.g. one incoming HTTP request), returning the bound context, a cleanup
// function that must be called once that unit of work is done, and any error
// encountered while creating the state.
type ContextApplier func(ctx context.Context) (context.Context, func(), error)

// InstantiateHost instantiates the Host module.
func InstantiateHost(ctx context.Context, runtime wazero.Runtime, mod wazero.CompiledModule, settings Settings) (ContextApplier, error) {
	return func(ctx context.Context) (context.Context, func(), error) {
		return ctx, func() {}, nil
	}, nil
}
