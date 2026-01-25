package plugins

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/tcp"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/traefik/yaegi/stdlib/syscall"
	"github.com/traefik/yaegi/stdlib/unsafe"
)

type yaegiMiddlewareBuilder struct {
	fnNew          reflect.Value
	fnCreateConfig reflect.Value
	fnNewTCP       reflect.Value
	interpreter    *interp.Interpreter // Store for Eval() usage
	basePkg        string              // Store package name for Eval()
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

	// Optionally check for NewTCP function (for TCP support).
	fnNewTCP, _ := i.Eval(basePkg + `.NewTCP`)
	// No error if not found - it's optional.

	return &yaegiMiddlewareBuilder{
		fnNew:          fnNew,
		fnCreateConfig: fnCreateConfig,
		fnNewTCP:       fnNewTCP,
		interpreter:    i,
		basePkg:        basePkg,
	}, nil
}

func (b yaegiMiddlewareBuilder) newMiddleware(config map[string]any, middlewareName string) (pluginMiddleware, error) {
	vConfig, err := b.createConfig(config)
	if err != nil {
		return nil, err
	}

	return &YaegiMiddleware{
		middlewareName: middlewareName,
		config:         vConfig,
		builder:        b,
	}, nil
}

func (b yaegiMiddlewareBuilder) newHandler(ctx context.Context, next http.Handler, cfg reflect.Value, middlewareName string) (http.Handler, error) {
	args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(next), cfg, reflect.ValueOf(middlewareName)}
	results := b.fnNew.Call(args)

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

func (b yaegiMiddlewareBuilder) newTCPHandler(ctx context.Context, next tcp.Handler, cfg reflect.Value, middlewareName string) (tcp.Handler, error) {
	if !b.fnNewTCP.IsValid() {
		return nil, errors.New("plugin does not support TCP")
	}

	// Create a callback function that bridges to the compiled handler
	nextCallback := func(ctx context.Context, conn net.Conn, closeWrite func() error) {
		// Reconstruct tcp.WriteCloser from stdlib types
		wc := &connWithCloseWrite{
			Conn:       conn,
			closeWrite: closeWrite,
		}
		next.ServeTCP(ctx, wc)
	}

	// Now use reflect.Call() with ONLY stdlib types!
	args := []reflect.Value{
		reflect.ValueOf(ctx),            // context.Context - stdlib
		reflect.ValueOf(nextCallback),   // func(...) - stdlib
		cfg,                             // config
		reflect.ValueOf(middlewareName), // string - stdlib
	}

	results := b.fnNewTCP.Call(args)

	if len(results) > 1 && results[1].Interface() != nil {
		err, ok := results[1].Interface().(error)
		if !ok {
			return nil, fmt.Errorf("invalid error type: %T", results[1].Interface())
		}
		return nil, err
	}

	// Extract the handler function - yaegi returns it as a pure stdlib function type
	pluginHandlerValue := results[0]
	pluginFunc, ok := pluginHandlerValue.Interface().(func(context.Context, net.Conn, func() error))
	if !ok {
		return nil, fmt.Errorf("expected function, got %T", pluginHandlerValue.Interface())
	}

	// Wrap the function as tcp.Handler
	return tcp.HandlerFunc(func(ctx context.Context, conn tcp.WriteCloser) {
		closeWriteFunc := func() error {
			return conn.CloseWrite()
		}
		pluginFunc(ctx, conn, closeWriteFunc)
	}), nil
}

// connWithCloseWrite wraps net.Conn and adds CloseWrite method.
type connWithCloseWrite struct {
	net.Conn
	closeWrite func() error
}

func (c *connWithCloseWrite) CloseWrite() error {
	if c.closeWrite != nil {
		return c.closeWrite()
	}
	return nil
}

func (b yaegiMiddlewareBuilder) createConfig(config map[string]any) (reflect.Value, error) {
	results := b.fnCreateConfig.Call(nil)
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
	builder        yaegiMiddlewareBuilder
}

// NewHandler creates a new HTTP handler.
func (m *YaegiMiddleware) NewHandler(ctx context.Context, next http.Handler) (http.Handler, error) {
	return m.builder.newHandler(ctx, next, m.config, m.middlewareName)
}

// NewTCPHandler creates a TCP constructor.
func (m *YaegiMiddleware) NewTCPHandler(ctx context.Context, next tcp.Handler) (func(tcp.Handler) (tcp.Handler, error), error) {
	// Check if the plugin supports TCP
	if !m.builder.fnNewTCP.IsValid() {
		return nil, errors.New("plugin does not support TCP")
	}

	return func(next tcp.Handler) (tcp.Handler, error) {
		return m.builder.newTCPHandler(ctx, next, m.config, m.middlewareName)
	}, nil
}

func newInterpreter(ctx context.Context, goPath string, manifest *Manifest, settings Settings) (*interp.Interpreter, error) {
	i := interp.New(interp.Options{
		GoPath: goPath,
		Env:    os.Environ(),
		Stdout: logs.NoLevel(*log.Ctx(ctx), zerolog.DebugLevel),
		Stderr: logs.NoLevel(*log.Ctx(ctx), zerolog.ErrorLevel),
	})

	err := i.Use(stdlib.Symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to load symbols: %w", err)
	}

	if manifest.UseUnsafe && !settings.UseUnsafe {
		return nil, errors.New("this plugin uses restricted imports. If you want to use it, you need to allow useUnsafe in the settings")
	}

	if settings.UseUnsafe && manifest.UseUnsafe {
		err := i.Use(unsafe.Symbols)
		if err != nil {
			return nil, fmt.Errorf("failed to load unsafe symbols: %w", err)
		}

		err = i.Use(syscall.Symbols)
		if err != nil {
			return nil, fmt.Errorf("failed to load syscall symbols: %w", err)
		}
	}

	err = i.Use(ppSymbols())
	if err != nil {
		return nil, fmt.Errorf("failed to load provider symbols: %w", err)
	}

	// Note: TCP plugins use stdlib types only (net.Conn, context.Context, functions)
	// so we don't need to load Traefik's tcp package symbols

	_, err = i.Eval(fmt.Sprintf(`import "%s"`, manifest.Import))
	if err != nil {
		return nil, fmt.Errorf("failed to import plugin code %q: %w", manifest.Import, err)
	}

	return i, nil
}
