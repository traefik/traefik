package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/http-wasm/http-wasm-host-go/handler"
	wasm "github.com/http-wasm/http-wasm-host-go/handler/nethttp"
	"github.com/mitchellh/mapstructure"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/yaegi/interp"
	"github.com/tetratelabs/wazero"
)

// Build builds a middleware plugin.
func (b Builder) Build(pName string, config map[string]interface{}, middlewareName string) (Constructor, error) {
	if b.yaegiMiddlewareBuilders == nil || b.wasmMiddlewareBuilders == nil {
		return nil, fmt.Errorf("no plugin definitions in the static configuration: %s", pName)
	}

	// plugin (pName) can be located in yaegi middleware builders.
	if yaegiDescriptor, ok := b.yaegiMiddlewareBuilders[pName]; ok {
		m, err := newYaegiMiddleware(yaegiDescriptor, config, middlewareName)
		if err != nil {
			return nil, err
		}
		return m.NewHandler, nil
	}

	// Or in wasm middleware builders.
	if wasmDescriptor, ok := b.wasmMiddlewareBuilders[pName]; ok {
		return newWasmMiddleware(wasmDescriptor, config, middlewareName).NewHandler, nil
	}

	return nil, fmt.Errorf("unknown plugin type: %s", pName)
}

type yaegiMiddlewareBuilder struct {
	fnNew          reflect.Value
	fnCreateConfig reflect.Value
}

func newYaegiMiddlewareBuilder(i *interp.Interpreter, basePkg, imp string) (*yaegiMiddlewareBuilder, error) {
	if basePkg == "" {
		basePkg = strings.ReplaceAll(path.Base(imp), "-", "_")
	}

	fnNew, err := i.Eval(basePkg + `.New`)
	if err != nil {
		return nil, fmt.Errorf("failed to eval New: %w", err)
	}

	fnCreateConfig, err := i.Eval(basePkg + `.CreateConfig`)
	if err != nil {
		return nil, fmt.Errorf("failed to eval CreateConfig: %w", err)
	}

	return &yaegiMiddlewareBuilder{
		fnNew:          fnNew,
		fnCreateConfig: fnCreateConfig,
	}, nil
}

func (p yaegiMiddlewareBuilder) newHandler(ctx context.Context, next http.Handler, cfg reflect.Value, middlewareName string) (http.Handler, error) {
	args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(next), cfg, reflect.ValueOf(middlewareName)}
	results := p.fnNew.Call(args)

	if len(results) > 1 && results[1].Interface() != nil {
		err, ok := results[1].Interface().(error)
		if !ok {
			return nil, fmt.Errorf("invalid error type: %T", results[0].Interface())
		}
		return nil, err
	}

	handler, ok := results[0].Interface().(http.Handler)
	if !ok {
		return nil, fmt.Errorf("invalid handler type: %T", results[0].Interface())
	}
	return handler, nil
}

func (p yaegiMiddlewareBuilder) createConfig(config map[string]interface{}) (reflect.Value, error) {
	results := p.fnCreateConfig.Call(nil)
	if len(results) != 1 {
		return reflect.Value{}, fmt.Errorf("invalid number of return for the CreateConfig function: %d", len(results))
	}

	vConfig := results[0]
	if len(config) == 0 {
		return vConfig, nil
	}

	cfg := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToSliceHookFunc(","),
		WeaklyTypedInput: true,
		Result:           vConfig.Interface(),
	}

	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to create configuration decoder: %w", err)
	}

	err = decoder.Decode(config)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to decode configuration: %w", err)
	}

	return vConfig, nil
}

// YaegiMiddleware is an HTTP handler plugin wrapper.
type YaegiMiddleware struct {
	middlewareName string
	config         reflect.Value
	builder        *yaegiMiddlewareBuilder
}

func newYaegiMiddleware(builder *yaegiMiddlewareBuilder, config map[string]interface{}, middlewareName string) (*YaegiMiddleware, error) {
	vConfig, err := builder.createConfig(config)
	if err != nil {
		return nil, err
	}

	return &YaegiMiddleware{
		middlewareName: middlewareName,
		config:         vConfig,
		builder:        builder,
	}, nil
}

// NewHandler creates a new HTTP handler.
func (m *YaegiMiddleware) NewHandler(ctx context.Context, next http.Handler) (http.Handler, error) {
	return m.builder.newHandler(ctx, next, m.config, m.middlewareName)
}

type wasmMiddlewareBuilder struct {
	wasmPath string
}

func newWasmMiddlewareBuilder(wasmPath string) *wasmMiddlewareBuilder {
	return &wasmMiddlewareBuilder{
		wasmPath: wasmPath,
	}
}

func (p wasmMiddlewareBuilder) newWasmHandler(ctx context.Context, next http.Handler, cfg reflect.Value, middlewareName string) (http.Handler, error) {
	code, err := os.ReadFile(p.wasmPath)
	if err != nil {
		return nil, err
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
			return nil, err
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
	builder        *wasmMiddlewareBuilder
}

func newWasmMiddleware(builder *wasmMiddlewareBuilder, config map[string]interface{}, middlewareName string) *WasmMiddleware {
	return &WasmMiddleware{
		middlewareName: middlewareName,
		config:         reflect.ValueOf(config),
		builder:        builder,
	}
}

// NewHandler creates a new HTTP handler.
func (y *WasmMiddleware) NewHandler(ctx context.Context, next http.Handler) (http.Handler, error) {
	return y.builder.newWasmHandler(ctx, next, y.config, y.middlewareName)
}
