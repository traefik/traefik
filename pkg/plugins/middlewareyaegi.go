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
	// errNewTCP is why the plugin cannot build TCP handlers, nil when NewTCP is valid.
	errNewTCP error
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

	// NewTCP is optional: a plugin without it only supports HTTP.
	errNewTCP := errors.New("plugin does not support TCP")
	fnNewTCP, err := i.Eval(basePkg + `.NewTCP`)
	if err == nil {
		errNewTCP = checkNewTCPSignature(fnNewTCP)
	}

	return &yaegiMiddlewareBuilder{
		fnNew:          fnNew,
		fnCreateConfig: fnCreateConfig,
		fnNewTCP:       fnNewTCP,
		errNewTCP:      errNewTCP,
	}, nil
}

// pluginTCPHandler is the TCP handler shape exchanged with Yaegi plugins.
// It only uses standard library types, because interpreted code cannot implement Traefik's tcp.Handler.
type pluginTCPHandler = func(ctx context.Context, conn net.Conn, closeWrite func() error)

// checkNewTCPSignature rejects a NewTCP function that reflect.Call would panic on.
func checkNewTCPSignature(fn reflect.Value) error {
	const want = "func(ctx context.Context, next func(ctx context.Context, conn net.Conn, closeWrite func() error), config *Config, name string) (func(ctx context.Context, conn net.Conn, closeWrite func() error), error)"

	if fn.Kind() != reflect.Func {
		return fmt.Errorf("NewTCP is not a function, want %s", want)
	}

	handlerType := reflect.TypeFor[pluginTCPHandler]()
	t := fn.Type()

	validIn := t.NumIn() == 4 &&
		reflect.TypeFor[context.Context]().AssignableTo(t.In(0)) &&
		handlerType.AssignableTo(t.In(1)) &&
		reflect.TypeFor[string]().AssignableTo(t.In(3))
	validOut := t.NumOut() == 2 &&
		t.Out(0).ConvertibleTo(handlerType) &&
		t.Out(1).Implements(reflect.TypeFor[error]())

	if !validIn || !validOut {
		return fmt.Errorf("invalid NewTCP signature %s, want %s", t, want)
	}

	return nil
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

func (b yaegiMiddlewareBuilder) newTCPHandler(ctx context.Context, next tcp.Handler, cfg reflect.Value, middlewareName string) (tcp.Handler, error) {
	if b.errNewTCP != nil {
		return nil, b.errNewTCP
	}

	nextCallback := func(ctx context.Context, conn net.Conn, closeWrite func() error) {
		next.ServeTCP(ctx, &connWithCloseWrite{Conn: conn, closeWrite: closeWrite})
	}

	args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(nextCallback), cfg, reflect.ValueOf(middlewareName)}
	results := b.fnNewTCP.Call(args)

	if results[1].Interface() != nil {
		err, ok := reflect.TypeAssert[error](results[1])
		if !ok {
			return nil, fmt.Errorf("invalid error type: %T", results[1].Interface())
		}

		return nil, err
	}

	if results[0].IsNil() {
		return nil, errors.New("NewTCP returned a nil handler")
	}

	pluginFunc, ok := reflect.TypeAssert[pluginTCPHandler](results[0].Convert(reflect.TypeFor[pluginTCPHandler]()))
	if !ok {
		return nil, fmt.Errorf("invalid TCP handler type: %T", results[0].Interface())
	}

	return tcp.HandlerFunc(func(ctx context.Context, conn tcp.WriteCloser) {
		pluginFunc(pluginTCPContext{Context: tcp.WithMetadata(ctx)}, conn, conn.CloseWrite)
	}), nil
}

// pluginTCPContext answers tcp.PluginMetadataKey with the connection metadata map.
// Traefik stores the map under a private key, which interpreted code cannot reference.
type pluginTCPContext struct {
	context.Context
}

func (c pluginTCPContext) Value(key any) any {
	if k, ok := key.(string); ok && k == tcp.PluginMetadataKey {
		metadata := tcp.Metadata(c.Context)
		if metadata == nil {
			// A nil map wrapped in an interface is not nil: the plugin's type assertion would succeed and its write would panic.
			return nil
		}

		return metadata
	}

	return c.Context.Value(key)
}

// connWithCloseWrite rebuilds a tcp.WriteCloser from the standard library types a plugin hands to its next handler.
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

// MarkBackend forwards the access-log signal through the stdlib connection wrapper.
func (c *connWithCloseWrite) MarkBackend() {
	if marker, ok := c.Conn.(interface{ MarkBackend() }); ok {
		marker.MarkBackend()
	}
}

// SetRejection forwards a middleware rejection through the stdlib connection wrapper.
func (c *connWithCloseWrite) SetRejection(err error) {
	if setter, ok := c.Conn.(interface{ SetRejection(err error) }); ok {
		setter.SetRejection(err)
	}
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
func (m *YaegiMiddleware) NewTCPHandler(ctx context.Context, _ tcp.Handler) (func(tcp.Handler) (tcp.Handler, error), error) {
	if m.builder.errNewTCP != nil {
		return nil, m.builder.errNewTCP
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

	_, err = i.Eval(fmt.Sprintf(`import "%s"`, manifest.Import))
	if err != nil {
		return nil, fmt.Errorf("failed to import plugin code %q: %w", manifest.Import, err)
	}

	return i, nil
}
