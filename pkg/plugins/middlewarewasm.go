package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/http-wasm/http-wasm-host-go/handler"
	wasm "github.com/http-wasm/http-wasm-host-go/handler/nethttp"
	"github.com/tetratelabs/wazero"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

type wasmMiddlewareBuilder struct {
	path     string
	cache    wazero.CompilationCache
	settings Settings
}

func newWasmMiddlewareBuilder(goPath, moduleName, wasmPath string, settings Settings) (*wasmMiddlewareBuilder, error) {
	ctx := context.Background()
	path := filepath.Join(goPath, "src", moduleName, wasmPath)
	cache := wazero.NewCompilationCache()

	code, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading Wasm binary: %w", err)
	}

	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCompilationCache(cache))
	if _, err = rt.CompileModule(ctx, code); err != nil {
		return nil, fmt.Errorf("compiling guest module: %w", err)
	}

	return &wasmMiddlewareBuilder{path: path, cache: cache, settings: settings}, nil
}

func (b wasmMiddlewareBuilder) newMiddleware(config map[string]interface{}, middlewareName string) (pluginMiddleware, error) {
	return &WasmMiddleware{
		middlewareName: middlewareName,
		config:         reflect.ValueOf(config),
		builder:        b,
	}, nil
}

func (b wasmMiddlewareBuilder) newHandler(ctx context.Context, next http.Handler, cfg reflect.Value, middlewareName string) (http.Handler, error) {
	h, applyCtx, err := b.buildMiddleware(ctx, next, cfg, middlewareName)
	if err != nil {
		return nil, fmt.Errorf("building Wasm middleware: %w", err)
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		h.ServeHTTP(rw, req.WithContext(applyCtx(req.Context())))
	}), nil
}

func (b *wasmMiddlewareBuilder) buildMiddleware(ctx context.Context, next http.Handler, cfg reflect.Value, middlewareName string) (http.Handler, func(ctx context.Context) context.Context, error) {
	code, err := os.ReadFile(b.path)
	if err != nil {
		return nil, nil, fmt.Errorf("loading binary: %w", err)
	}

	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCompilationCache(b.cache))

	guestModule, err := rt.CompileModule(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("compiling guest module: %w", err)
	}

	applyCtx, err := InstantiateHost(ctx, rt, guestModule, b.settings)
	if err != nil {
		return nil, nil, fmt.Errorf("instantiating host module: %w", err)
	}

	logger := middlewares.GetLogger(ctx, middlewareName, "wasm")

	config := wazero.NewModuleConfig().WithSysWalltime().WithStartFunctions("_start", "_initialize")
	for _, env := range b.settings.Envs {
		config = config.WithEnv(env, os.Getenv(env))
	}

	if len(b.settings.Mounts) > 0 {
		fsConfig := wazero.NewFSConfig()
		for _, mount := range b.settings.Mounts {
			withDir := fsConfig.WithDirMount
			prefix, readOnly := strings.CutSuffix(mount, ":ro")
			if readOnly {
				withDir = fsConfig.WithReadOnlyDirMount
			}
			parts := strings.Split(prefix, ":")
			switch {
			case len(parts) == 1:
				fsConfig = withDir(parts[0], parts[0])
			case len(parts) == 2:
				fsConfig = withDir(parts[0], parts[1])
			default:
				return nil, nil, fmt.Errorf("invalid directory %q", mount)
			}
		}
		config = config.WithFSConfig(fsConfig)
	}

	opts := []handler.Option{
		handler.ModuleConfig(config),
		handler.Logger(logs.NewWasmLogger(logger)),
	}

	i := cfg.Interface()
	if i != nil {
		config, ok := i.(map[string]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("could not type assert config: %T", i)
		}

		data, err := json.Marshal(config)
		if err != nil {
			return nil, nil, fmt.Errorf("marshaling config: %w", err)
		}

		opts = append(opts, handler.GuestConfig(data))
	}

	opts = append(opts, handler.Runtime(func(ctx context.Context) (wazero.Runtime, error) {
		return rt, nil
	}))

	mw, err := wasm.NewMiddleware(applyCtx(ctx), code, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("creating middleware: %w", err)
	}

	h := mw.NewHandler(ctx, next)

	// Traefik does not Close the middleware when creating a new instance on a configuration change.
	// When the middleware is marked to be GC, we need to close it so the wasm instance is properly closed.
	// Reference: https://github.com/traefik/traefik/issues/11119
	runtime.SetFinalizer(h, func(_ http.Handler) {
		if err := mw.Close(ctx); err != nil {
			logger.Err(err).Msg("[wasm] middleware Close failed")
		} else {
			logger.Debug().Msg("[wasm] middleware Close ok")
		}
	})
	return h, applyCtx, nil
}

// WasmMiddleware is an HTTP handler plugin wrapper.
type WasmMiddleware struct {
	middlewareName string
	config         reflect.Value
	builder        wasmMiddlewareBuilder
}

// NewHandler creates a new HTTP handler.
func (m WasmMiddleware) NewHandler(ctx context.Context, next http.Handler) (http.Handler, error) {
	return m.builder.newHandler(ctx, next, m.config, m.middlewareName)
}
