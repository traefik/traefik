package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"

	"github.com/http-wasm/http-wasm-host-go/handler"
	wasm "github.com/http-wasm/http-wasm-host-go/handler/nethttp"
	"github.com/tetratelabs/wazero"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

type wasmMiddlewareBuilder struct {
	wasmPath string
}

func newWasmMiddlewareBuilder(goPath string, moduleName string, manifest *Manifest) *wasmMiddlewareBuilder {
	return &wasmMiddlewareBuilder{
		wasmPath: filepath.Join(goPath, "src", moduleName, manifest.WasmPath),
	}
}

func (b wasmMiddlewareBuilder) newMiddleware(config map[string]interface{}, middlewareName string) (pluginMiddleware, error) {
	return &WasmMiddleware{
		middlewareName: middlewareName,
		config:         reflect.ValueOf(config),
		builder:        b,
	}, nil
}

func (b wasmMiddlewareBuilder) newWasmHandler(ctx context.Context, next http.Handler, cfg reflect.Value, middlewareName string) (http.Handler, error) {
	code, err := os.ReadFile(b.wasmPath)
	if err != nil {
		return nil, fmt.Errorf("loading Wasm binary: %w", err)
	}
	logger := middlewares.GetLogger(ctx, middlewareName, "wasm")
	opts := []handler.Option{
		handler.ModuleConfig(wazero.NewModuleConfig().WithSysWalltime()),
		handler.Logger(initWasmLogger(logger)),
	}
	i := cfg.Interface()
	if i != nil {
		config, ok := i.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("could not type assert config: %T", i)
		}
		b, err := json.Marshal(config)
		if err != nil {
			return nil, fmt.Errorf("marshaling config: %w", err)
		}
		opts = append(opts, handler.GuestConfig(b))
	}

	mw, err := wasm.NewMiddleware(context.Background(), code, opts...)
	if err != nil {
		return nil, err
	}
	return mw.NewHandler(ctx, next), nil
}

// WasmMiddleware is an HTTP handler plugin wrapper.
type WasmMiddleware struct {
	middlewareName string
	config         reflect.Value
	builder        wasmMiddlewareBuilder
}

// NewHandler creates a new HTTP handler.
func (y WasmMiddleware) NewHandler(ctx context.Context, next http.Handler) (http.Handler, error) {
	return y.builder.newWasmHandler(ctx, next, y.config, y.middlewareName)
}
