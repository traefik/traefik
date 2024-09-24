//go:build linux || darwin

package plugins

import (
	"context"
	"fmt"
	"os"

	"github.com/stealthrocket/wasi-go/imports"
	wazergo_wasip1 "github.com/stealthrocket/wasi-go/imports/wasi_snapshot_preview1"
	"github.com/stealthrocket/wazergo"
	"github.com/tetratelabs/wazero"
	wazero_wasip1 "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type ContextApplier func(ctx context.Context) context.Context

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

		ctx, sys, err := builder.Instantiate(ctx, runtime)
		if err != nil {
			return nil, err
		}

		inst, err := wazergo.Instantiate(ctx, runtime, wazergo_wasip1.NewHostModule(*extension), wazergo_wasip1.WithWASI(sys))
		if err != nil {
			return nil, fmt.Errorf("wazergo instantiation: %w", err)
		}

		return func(ctx context.Context) context.Context {
			return wazergo.WithModuleInstance(ctx, inst)
		}, nil
	}

	_, err := wazero_wasip1.Instantiate(ctx, runtime)
	if err != nil {
		return nil, fmt.Errorf("wazero instantiation: %w", err)
	}

	return func(ctx context.Context) context.Context {
		return ctx
	}, nil
}
