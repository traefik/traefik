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
	fnNewTCP       reflect.Value
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

	// NewTCP is optional: only the plugins that can handle TCP connections export it.
	fnNewTCP, _ := i.Eval(basePkg + `.NewTCP`)

	return &yaegiMiddlewareBuilder{
		fnNew:          fnNew,
		fnNewTCP:       fnNewTCP,
		fnCreateConfig: fnCreateConfig,
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
	return wrapExternal(next, func(next http.Handler) (http.Handler, error) {
		return b.newPluginHandler(ctx, next, cfg, middlewareName)
	})
}

func (b yaegiMiddlewareBuilder) newPluginHandler(ctx context.Context, next http.Handler, cfg reflect.Value, middlewareName string) (http.Handler, error) {
	args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(next), cfg, reflect.ValueOf(middlewareName)}
	results := b.fnNew.Call(args)

	if len(results) > 1 && results[1].Interface() != nil {
		err, ok := reflect.TypeAssert[error](results[1])
		if !ok {
			return nil, fmt.Errorf("invalid error type: %T", results[1].Interface())
		}

		return nil, err
	}

	handler, ok := reflect.TypeAssert[http.Handler](results[0])
	if !ok {
		return nil, fmt.Errorf("invalid handler type: %T", results[0].Interface())
	}

	return handler, nil
}

// newTCPHandler only exchanges standard library types with the plugin,
// so that the interpreter does not have to load Traefik's own packages.
func (b yaegiMiddlewareBuilder) newTCPHandler(ctx context.Context, next tcp.Handler, cfg reflect.Value, middlewareName string) (tcp.Handler, error) {
	if !b.fnNewTCP.IsValid() {
		return nil, errors.New("the plugin does not export a NewTCP function")
	}

	nextFunc := func(_ context.Context, conn net.Conn, closeWrite func() error) {
		// A connection wrapped by the plugin loses CloseWrite when it crosses the interpreter boundary.
		if wc, ok := conn.(tcp.WriteCloser); ok {
			next.ServeTCP(wc)
			return
		}

		next.ServeTCP(&closeWriteConn{Conn: conn, closeWrite: closeWrite})
	}

	args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(nextFunc), cfg, reflect.ValueOf(middlewareName)}
	results := b.fnNewTCP.Call(args)

	if len(results) > 1 && results[1].Interface() != nil {
		err, ok := reflect.TypeAssert[error](results[1])
		if !ok {
			return nil, fmt.Errorf("invalid error type: %T", results[1].Interface())
		}

		return nil, err
	}

	handle, ok := reflect.TypeAssert[func(context.Context, net.Conn, func() error)](results[0])
	if !ok {
		return nil, fmt.Errorf("invalid TCP handler type: %T", results[0].Interface())
	}

	return tcp.HandlerFunc(func(conn tcp.WriteCloser) {
		connCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		handle(connCtx, conn, conn.CloseWrite)
	}), nil
}

type closeWriteConn struct {
	net.Conn

	closeWrite func() error
}

func (c *closeWriteConn) CloseWrite() error {
	if c.closeWrite == nil {
		return nil
	}

	return c.closeWrite()
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

// NewTCPHandler creates a new TCP handler.
func (m *YaegiMiddleware) NewTCPHandler(ctx context.Context, next tcp.Handler) (tcp.Handler, error) {
	return m.builder.newTCPHandler(ctx, next, m.config, m.middlewareName)
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

	_, err = i.Eval(fmt.Sprintf(`import "%s"`, manifest.Import))
	if err != nil {
		return nil, fmt.Errorf("failed to import plugin code %q: %w", manifest.Import, err)
	}

	return i, nil
}
