//go:build linux || darwin

package plugins

import (
	"context"
	"fmt"

	"github.com/samyfodil/wazy"
	wazy_wasip1 "github.com/samyfodil/wazy/imports/wasi_snapshot_preview1"
	"github.com/samyfodil/wazy/imports/wasmedge"
)

type ContextApplier func(ctx context.Context) context.Context

// InstantiateHost instantiates the Host module according to the guest requirements (for now only SocketExtensions).
func InstantiateHost(ctx context.Context, runtime wazy.Runtime, mod wazy.CompiledModule, settings Settings) (ContextApplier, error) {
	if v := wasmedge.Detect(mod.ImportedFunctions()); v != wasmedge.None {
		if _, err := wasmedge.Instantiate(ctx, runtime, v); err != nil {
			return nil, fmt.Errorf("wazy instantiation: %w", err)
		}
		return func(ctx context.Context) context.Context {
			return ctx
		}, nil
	}

	_, err := wazy_wasip1.Instantiate(ctx, runtime)
	if err != nil {
		return nil, fmt.Errorf("wazy instantiation: %w", err)
	}

	return func(ctx context.Context) context.Context {
		return ctx
	}, nil
}
