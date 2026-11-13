package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/darkweak/wy/tree"
	"github.com/darkweak/wy/tree/native"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	// Registers the syscall symbols, which wy does not provide.
	_ "github.com/traefik/traefik/v3/pkg/plugins/wysyscall"
)

func init() {
	// Lets the plugins run by wy implement PP: wy hands the interpreted
	// value to Traefik through the _wyPP wrapper.
	native.Register(&native.Package{
		Path: "github.com/traefik/traefik/v3/pkg/plugins",
		Name: "plugins",
		Types: map[string]reflect.Type{
			"PP": reflect.TypeFor[PP](),
		},
		Wrappers: map[string]reflect.Type{
			"PP": reflect.TypeFor[_wyPP](),
		},
	})
}

// newWyInterpreter creates a wy interpreter and imports the plugin in it.
func newWyInterpreter(ctx context.Context, goPath string, manifest *Manifest, settings Settings) (*tree.Interp, error) {
	if manifest.UseUnsafe && !settings.UseUnsafe {
		return nil, errors.New("this plugin uses restricted imports. If you want to use it, you need to allow useUnsafe in the settings")
	}

	var forbid []string
	if !settings.UseUnsafe || !manifest.UseUnsafe {
		forbid = []string{"unsafe", "syscall"}
	}

	logger := log.Ctx(ctx)
	logger.Debug().Msg("Loading plugin with wy")

	i := tree.New(tree.Options{
		GoPath:     goPath,
		Env:        os.Environ(),
		Forbid:     forbid,
		Restricted: true,
		Stdout:     logs.NoLevel(*logger, zerolog.DebugLevel),
		Stderr:     logs.NoLevel(*logger, zerolog.ErrorLevel),
	})

	_, err := i.Eval(ctx, fmt.Sprintf(`import "%s"`, manifest.Import))
	if err != nil {
		// go/types reports import failures as text only.
		if len(forbid) > 0 && strings.Contains(err.Error(), "is not allowed") {
			return nil, fmt.Errorf("this plugin uses restricted imports. If you want to use it, you need to allow useUnsafe in the settings and in the plugin manifest: %w", err)
		}

		return nil, fmt.Errorf("failed to import plugin code %q: %w", manifest.Import, err)
	}

	return i, nil
}

// evalWyValue evaluates an expression that yields a single value.
func evalWyValue(ctx context.Context, i *tree.Interp, expr string) (reflect.Value, error) {
	values, err := i.Eval(ctx, expr)
	if err != nil {
		return reflect.Value{}, err
	}

	if len(values) != 1 {
		return reflect.Value{}, fmt.Errorf("expected 1 value, got %d", len(values))
	}

	return reflect.ValueOf(values[0]), nil
}

func wyBasePkg(basePkg, imp string) string {
	if basePkg != "" {
		return basePkg
	}

	return strings.ReplaceAll(path.Base(imp), "-", "_")
}

// createWyConfig calls the CreateConfig function of the plugin and decodes config into its result.
func createWyConfig(fnCreateConfig reflect.Value, config map[string]any) (reflect.Value, error) {
	results := fnCreateConfig.Call(nil)
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

// wyError returns the error held by the second result of a call into the plugin.
func wyError(results []reflect.Value) error {
	if len(results) < 2 || results[1].Interface() == nil {
		return nil
	}

	err, ok := reflect.TypeAssert[error](results[1])
	if !ok {
		return fmt.Errorf("invalid error type: %T", results[1].Interface())
	}

	return err
}

type wyMiddlewareBuilder struct {
	fnNew          reflect.Value
	fnCreateConfig reflect.Value
}

func newWyMiddlewareBuilder(ctx context.Context, i *tree.Interp, basePkg, imp string) (*wyMiddlewareBuilder, error) {
	basePkg = wyBasePkg(basePkg, imp)

	fnNew, err := evalWyValue(ctx, i, basePkg+`.New`)
	if err != nil {
		return nil, fmt.Errorf("failed to eval New: %w", err)
	}

	fnCreateConfig, err := evalWyValue(ctx, i, basePkg+`.CreateConfig`)
	if err != nil {
		return nil, fmt.Errorf("failed to eval CreateConfig: %w", err)
	}

	return &wyMiddlewareBuilder{
		fnNew:          fnNew,
		fnCreateConfig: fnCreateConfig,
	}, nil
}

func (b wyMiddlewareBuilder) newMiddleware(config map[string]any, middlewareName string) (pluginMiddleware, error) {
	vConfig, err := createWyConfig(b.fnCreateConfig, config)
	if err != nil {
		return nil, err
	}

	return &WyMiddleware{
		middlewareName: middlewareName,
		config:         vConfig,
		builder:        b,
	}, nil
}

// WyMiddleware is an HTTP handler plugin wrapper, for the plugins run by wy.
type WyMiddleware struct {
	middlewareName string
	config         reflect.Value
	builder        wyMiddlewareBuilder
}

// NewHandler creates a new HTTP handler.
func (m *WyMiddleware) NewHandler(ctx context.Context, next http.Handler) (http.Handler, error) {
	return wrapExternal(next, func(next http.Handler) (http.Handler, error) {
		args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(next), m.config, reflect.ValueOf(m.middlewareName)}

		results := m.builder.fnNew.Call(args)
		if err := wyError(results); err != nil {
			return nil, err
		}

		handler, ok := reflect.TypeAssert[http.Handler](results[0])
		if !ok {
			return nil, fmt.Errorf("invalid handler type: %T", results[0].Interface())
		}

		return handler, nil
	})
}

func newWyProvider(builder providerBuilder, config map[string]any, providerName string) (*Provider, error) {
	basePkg := wyBasePkg(builder.BasePkg, builder.Import)
	i := builder.wyInterpreter

	ctx := context.Background()

	fnCreateConfig, err := evalWyValue(ctx, i, basePkg+`.CreateConfig`)
	if err != nil {
		return nil, fmt.Errorf("failed to eval CreateConfig: %w", err)
	}

	vConfig, err := createWyConfig(fnCreateConfig, config)
	if err != nil {
		return nil, err
	}

	// The interface conversion happens in interpreted code, which makes wy
	// wrap the plugin's provider in _wyPP.
	_, err = i.Eval(ctx, `import (
	"context"

	`+basePkg+` "`+builder.Import+`"
	"github.com/traefik/traefik/v3/pkg/plugins"
)

func traefikNewProvider(ctx context.Context, config *`+basePkg+`.Config, name string) (plugins.PP, error) {
	p, err := `+basePkg+`.New(ctx, config, name)
	var pv plugins.PP = p
	return pv, err
}
`)
	if err != nil {
		return nil, fmt.Errorf("failed to eval wrapper: %w", err)
	}

	fnNew, err := evalWyValue(ctx, i, "traefikNewProvider")
	if err != nil {
		return nil, fmt.Errorf("failed to eval New: %w", err)
	}

	results := fnNew.Call([]reflect.Value{reflect.ValueOf(ctx), vConfig, reflect.ValueOf(providerName)})
	if err := wyError(results); err != nil {
		return nil, err
	}

	prov, ok := reflect.TypeAssert[PP](results[0])
	if !ok {
		return nil, fmt.Errorf("invalid provider type: %T", results[0].Interface())
	}

	return &Provider{name: providerName, pp: prov}, nil
}

// _wyPP exposes a PP implemented by code run by wy to Traefik.
type _wyPP struct{ XWyBox native.Caller }

func (w _wyPP) Init() error {
	r := w.XWyBox.Call("Init", nil)
	return native.As[error](r[0])
}

func (w _wyPP) Provide(cfgChan chan<- json.Marshaler) error {
	r := w.XWyBox.Call("Provide", []any{cfgChan})
	return native.As[error](r[0])
}

func (w _wyPP) Stop() error {
	r := w.XWyBox.Call("Stop", nil)
	return native.As[error](r[0])
}
