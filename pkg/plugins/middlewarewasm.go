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
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

type wasmMiddlewareBuilder struct {
	path string
}

func newWasmMiddlewareBuilder(goPath string, moduleName, wasmPath string) *wasmMiddlewareBuilder {
	return &wasmMiddlewareBuilder{path: filepath.Join(goPath, "src", moduleName, wasmPath)}
}

func (b wasmMiddlewareBuilder) newMiddleware(config map[string]interface{}, middlewareName string) (pluginMiddleware, error) {
	return &WasmMiddleware{
		middlewareName: middlewareName,
		config:         reflect.ValueOf(config),
		builder:        b,
	}, nil
}

func toWasmConfig(i any) ([]handler.Option, bool, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return nil, false, fmt.Errorf("marshaling config: %w", err)
	}

	var wc wasmConfig
	if err := json.Unmarshal(b, &wc); err != nil {
		return nil, false, fmt.Errorf("unmarshaling config: %w", err)
	}

	if wc.Runtime == nil && wc.Config == nil {
		return nil, false, nil
	}

	return []handler.Option{
		handler.GuestConfig(wc.Config),
		handler.ModuleConfig(
			wazero.NewModuleConfig().
				WithSysWalltime().
				WithFSConfig(wazero.NewFSConfig().WithReadOnlyDirMount(wc.Runtime.RootFS, "/")),
		),
	}, true, nil
}

func toConfig(i any) ([]handler.Option, error) {
	im, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("could not type assert config: %T", i)
	}

	data, err := json.Marshal(im)
	if err != nil {
		return nil, fmt.Errorf("marshaling config: %w", err)
	}

	return []handler.Option{
		handler.GuestConfig(data),
		handler.ModuleConfig(wazero.NewModuleConfig().WithSysWalltime()),
	}, nil
}

type wasmConfig struct {
	Runtime *struct {
		RootFS string `json:"rootFS"`
	} `json:"runtime"`
	Config json.RawMessage `json:"config"`
}

func (b wasmMiddlewareBuilder) newHandler(ctx context.Context, next http.Handler, cfg reflect.Value, middlewareName string) (http.Handler, error) {
	code, err := os.ReadFile(b.path)
	if err != nil {
		return nil, fmt.Errorf("loading Wasm binary: %w", err)
	}

	logger := middlewares.GetLogger(ctx, middlewareName, "wasm")

	opts := []handler.Option{
		handler.Logger(logs.NewWasmLogger(logger)),
	}

	i := cfg.Interface()
	if i != nil {
		var (
			cOpts []handler.Option
			ok    bool
			err   error
		)

		cOpts, ok, err = toWasmConfig(i)
		if !ok {
			cOpts, err = toConfig(i)
		}

		if err != nil {
			return nil, err
		}

		opts = append(opts, cOpts...)
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
func (m WasmMiddleware) NewHandler(ctx context.Context, next http.Handler) (http.Handler, error) {
	return m.builder.newHandler(ctx, next, m.config, m.middlewareName)
}
