//go:build linux || darwin

package plugins

import (
	"context"
	"errors"
	"fmt"

	"github.com/samyfodil/wazy"
	wazy_wasip1 "github.com/samyfodil/wazy/imports/wasi_snapshot_preview1"
)

type ContextApplier func(ctx context.Context) context.Context

// hasSocketsExtension mirrors wasi-go's DetectSocketsExtension check.
func hasSocketsExtension(mod wazy.CompiledModule) bool {
	for _, f := range mod.ImportedFunctions() {
		moduleName, name, ok := f.Import()
		if !ok || moduleName != wazy_wasip1.ModuleName {
			continue
		}
		switch name {
		case "sock_open", "sock_bind", "sock_connect", "sock_listen", "sock_accept",
			"sock_send_to", "sock_recv_from", "sock_getsockopt", "sock_setsockopt",
			"sock_getlocaladdr", "sock_getpeeraddr", "sock_getaddrinfo":
			return true
		}
	}
	return false
}

// InstantiateHost instantiates the Host module according to the guest requirements (for now only SocketExtensions).
func InstantiateHost(ctx context.Context, runtime wazy.Runtime, mod wazy.CompiledModule, settings Settings) (ContextApplier, error) {
	if hasSocketsExtension(mod) {
		return nil, errors.New("wasm: the sockets extension is not supported on the wazy runtime")
	}

	_, err := wazy_wasip1.Instantiate(ctx, runtime)
	if err != nil {
		return nil, fmt.Errorf("wazy instantiation: %w", err)
	}

	return func(ctx context.Context) context.Context {
		return ctx
	}, nil
}
