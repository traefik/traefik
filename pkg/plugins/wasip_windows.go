//go:build windows

package plugins

import (
	"context"

	"github.com/tetratelabs/wazero"
)

type ContextApplier func(ctx context.Context) context.Context

func Instantiate(ctx context.Context, runtime wazero.Runtime, mod wazero.CompiledModule, settings Settings) (ContextApplier, error) {
	return func(ctx context.Context) context.Context {
		return ctx
	}, nil
}
