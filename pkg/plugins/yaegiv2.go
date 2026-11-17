package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/yaegi/v2/host"
	"github.com/traefik/yaegi/v2/vm"
)

func loadYaegiV2Plugin(ctx context.Context, goPath string, manifest *Manifest, moduleName string, settings Settings, local bool) (*vm.Plugin, error) {
	if manifest.UseUnsafe && !settings.UseUnsafe {
		return nil, errors.New("this plugin uses restricted imports. If you want to use it, you need to allow useUnsafe in the settings")
	}

	logger := log.Ctx(ctx)
	logger.Debug().Msg("Loading plugin with Yaegi v2")

	opts := host.Options()
	opts.Root = filepath.Join(goPath, "src")
	opts.Dir = filepath.Join(opts.Root, filepath.FromSlash(moduleName))
	opts.UseUnsafe = settings.UseUnsafe && manifest.UseUnsafe
	// Same environment as with Yaegi v1, host.Options gives none.
	opts.Env = os.Environ()
	opts.Stdout = logs.NoLevel(*logger, zerolog.DebugLevel)
	opts.Stderr = logs.NoLevel(*logger, zerolog.ErrorLevel)
	opts.GoPanic = func(r any) {
		logger.Error().Interface("panic", r).Msg("Panic in a goroutine of the plugin")
	}

	// Only local plugins can come without a go.mod file.
	if local {
		opts.GoPath = goPath
	}

	plugin, err := host.LoadContext(ctx, opts, manifest.Import)
	if err != nil {
		return nil, fmt.Errorf("loading context: %w", err)
	}

	for _, note := range plugin.Notes {
		logger.Warn().Msg(note)
	}

	return plugin, nil
}

func createYaegiV2Config(plugin *vm.Plugin, config map[string]any) (*vm.Object, error) {
	results, err := plugin.Call(context.Background(), "CreateConfig")
	if err != nil {
		return nil, fmt.Errorf("calling CreateConfig: %w", err)
	}

	if len(results) != 1 {
		return nil, fmt.Errorf("invalid number of return for the CreateConfig function: %d", len(results))
	}

	vConfig := results[0]
	if len(config) == 0 {
		return vConfig, nil
	}

	err = vConfig.Decode(func(ptr any) error {
		cfg := &mapstructure.DecoderConfig{
			DecodeHook:       mapstructure.StringToSliceHookFunc(","),
			WeaklyTypedInput: true,
			Result:           ptr,
		}

		decoder, err := mapstructure.NewDecoder(cfg)
		if err != nil {
			return fmt.Errorf("creating configuration decoder: %w", err)
		}

		return decoder.Decode(config)
	})
	if err != nil {
		return nil, fmt.Errorf("decoding configuration: %w", err)
	}

	return vConfig, nil
}

// yaegiV2New calls the New function of the plugin and returns its first result.
func yaegiV2New(ctx context.Context, plugin *vm.Plugin, args ...any) (*vm.Object, error) {
	results, err := plugin.Call(ctx, "New", args...)
	if err != nil {
		return nil, fmt.Errorf("calling New: %w", err)
	}

	if len(results) == 0 {
		return nil, errors.New("invalid number of return for the New function: 0")
	}

	if len(results) > 1 {
		if err := yaegiV2Error(results[1]); err != nil {
			return nil, err
		}
	}

	return results[0], nil
}

// yaegiV2Error returns the error held by a result of a call into the plugin.
func yaegiV2Error(result *vm.Object) error {
	v, err := result.Interface()
	if err != nil {
		return err
	}

	if v == nil {
		return nil
	}

	resultErr, ok := v.(error)
	if !ok {
		return fmt.Errorf("invalid error type: %T", v)
	}

	return resultErr
}

type yaegiV2MiddlewareBuilder struct {
	plugin *vm.Plugin
}

func (b yaegiV2MiddlewareBuilder) newMiddleware(config map[string]any, middlewareName string) (pluginMiddleware, error) {
	vConfig, err := createYaegiV2Config(b.plugin, config)
	if err != nil {
		return nil, err
	}

	return &YaegiV2Middleware{
		middlewareName: middlewareName,
		config:         vConfig,
		plugin:         b.plugin,
	}, nil
}

// YaegiV2Middleware is an HTTP handler plugin wrapper, for the plugins run by Yaegi v2.
type YaegiV2Middleware struct {
	middlewareName string
	config         *vm.Object
	plugin         *vm.Plugin
}

// NewHandler creates a new HTTP handler.
func (m *YaegiV2Middleware) NewHandler(ctx context.Context, next http.Handler) (http.Handler, error) {
	return wrapExternal(next, func(next http.Handler) (http.Handler, error) {
		result, err := yaegiV2New(ctx, m.plugin, ctx, next, m.config, m.middlewareName)
		if err != nil {
			return nil, err
		}

		v, err := result.Interface()
		if err != nil {
			return nil, err
		}

		handler, ok := v.(http.Handler)
		if !ok {
			return nil, fmt.Errorf("invalid handler type: %T", v)
		}

		return handler, nil
	})
}

func newYaegiV2Provider(plugin *vm.Plugin, config map[string]any, providerName string) (*Provider, error) {
	vConfig, err := createYaegiV2Config(plugin, config)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	result, err := yaegiV2New(ctx, plugin, ctx, vConfig, providerName)
	if err != nil {
		return nil, err
	}

	return &Provider{name: providerName, pp: yaegiV2PP{object: result}}, nil
}

// yaegiV2PP is a provider run by Yaegi v2.
type yaegiV2PP struct {
	object *vm.Object
}

func (p yaegiV2PP) Init() error {
	return p.call("Init")
}

func (p yaegiV2PP) Provide(cfgChan chan<- json.Marshaler) error {
	return p.call("Provide", cfgChan)
}

func (p yaegiV2PP) Stop() error {
	return p.call("Stop")
}

func (p yaegiV2PP) call(name string, args ...any) error {
	results, err := p.object.Call(context.Background(), name, args...)
	if err != nil {
		return fmt.Errorf("calling %s: %w", name, err)
	}

	if len(results) != 1 {
		return fmt.Errorf("invalid number of return for the %s method: %d", name, len(results))
	}

	return yaegiV2Error(results[0])
}
