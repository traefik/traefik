//go:build linux || darwin

package plugins

import (
	"context"
	"fmt"
	"os"

	"github.com/stealthrocket/wasi-go/imports"
	"github.com/tetratelabs/wazero"
	wazero_wasip1 "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// ContextApplier binds host module state to ctx for a single unit of work
// (e.g. one incoming HTTP request), returning the bound context, a cleanup
// function that must be called once that unit of work is done, and any error
// encountered while creating the state.
type ContextApplier func(ctx context.Context) (context.Context, func(), error)

// InstantiateHost instantiates the Host module according to the guest requirements (for now only SocketExtensions).
func InstantiateHost(ctx context.Context, runtime wazero.Runtime, mod wazero.CompiledModule, settings Settings) (ContextApplier, error) {
	if extension := imports.DetectSocketsExtension(mod); extension != nil {
		envs := []string{}
		for _, env := range settings.Envs {
			envs = append(envs, fmt.Sprintf("%s=%s", env, os.Getenv(env)))
		}

		builder := imports.NewBuilder().WithSocketsExtension("auto", mod)
		if len(envs) > 0 {
			builder.WithEnv(envs...)
		}

		if len(settings.Mounts) > 0 {
			builder.WithDirs(settings.Mounts...)
		}

		// wasi-go's System owns real OS file descriptors (including sockets) and is
		// not safe to share across concurrently executing guest instances, which
		// http-wasm-host-go pools and runs concurrently. Builder.Instantiate is
		// designed to be called more than once against the same runtime: it
		// registers the wasm-level host module on the first call and cheaply binds
		// a fresh System to it on every subsequent call, so we instantiate a new
		// one per unit of work instead of sharing a single one for the middleware's
		// lifetime. See https://github.com/traefik/traefik/issues/11629.
		return func(ctx context.Context) (context.Context, func(), error) {
			hostCtx, sys, err := builder.Instantiate(ctx, runtime)
			if err != nil {
				return nil, nil, err
			}

			return hostCtx, func() { _ = sys.Close(context.Background()) }, nil
		}, nil
	}

	_, err := wazero_wasip1.Instantiate(ctx, runtime)
	if err != nil {
		return nil, fmt.Errorf("wazero instantiation: %w", err)
	}

	return func(ctx context.Context) (context.Context, func(), error) {
		return ctx, func() {}, nil
	}, nil
}
