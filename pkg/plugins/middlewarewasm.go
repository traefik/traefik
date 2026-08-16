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

	"github.com/samyfodil/wazy"
	"github.com/samyfodil/wazy/imports/http_handler"
	wasm "github.com/samyfodil/wazy/imports/http_handler/nethttp"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
)

type wasmMiddlewareBuilder struct {
	path     string
	cache    wazy.CompilationCache
	settings Settings
}

func newWasmMiddlewareBuilder(goPath, moduleName, wasmPath string, settings Settings) (*wasmMiddlewareBuilder, error) {
	ctx := context.Background()
	path := filepath.Join(goPath, "src", moduleName, wasmPath)
	cache := wazy.NewCompilationCache()

	code, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading Wasm binary: %w", err)
	}

	rt := wazy.NewRuntimeWithConfig(ctx, wazy.NewRuntimeConfig().WithCompilationCache(cache))
	if _, err = rt.CompileModule(ctx, code); err != nil {
		return nil, fmt.Errorf("compiling guest module: %w", err)
	}

	return &wasmMiddlewareBuilder{path: path, cache: cache, settings: settings}, nil
}

func (b wasmMiddlewareBuilder) newMiddleware(config map[string]any, middlewareName string) (pluginMiddleware, error) {
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

	rt := wazy.NewRuntimeWithConfig(ctx, wazy.NewRuntimeConfig().WithCompilationCache(b.cache))

	guestModule, err := rt.CompileModule(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("compiling guest module: %w", err)
	}

	applyCtx, err := InstantiateHost(ctx, rt, guestModule, b.settings)
	if err != nil {
		return nil, nil, fmt.Errorf("instantiating host module: %w", err)
	}

	logger := middlewares.GetLogger(ctx, middlewareName, "wasm")

	config := wazy.NewModuleConfig().WithSysWalltime().WithStartFunctions("_start", "_initialize")
	for _, env := range b.settings.Envs {
		config = config.WithEnv(env, os.Getenv(env))
	}

	if len(b.settings.Mounts) > 0 {
		fsConfig := wazy.NewFSConfig()
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

	opts := []http_handler.Option{
		http_handler.WithModuleConfig(config),
		http_handler.WithLogger(logs.NewWasmLogger(logger)),
	}

	i := cfg.Interface()
	if i != nil {
		config, ok := i.(map[string]any)
		if !ok {
			return nil, nil, fmt.Errorf("could not type assert config: %T", i)
		}

		data, err := json.Marshal(config)
		if err != nil {
			return nil, nil, fmt.Errorf("marshaling config: %w", err)
		}

		opts = append(opts, http_handler.WithGuestConfig(data))
	}

	opts = append(opts, http_handler.WithRuntime(func(ctx context.Context) (wazy.Runtime, error) {
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
