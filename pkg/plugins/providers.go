package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/yaegi/interp"
)

// PP the interface of a plugin's provider.
type PP interface {
	Init() error
	Provide(cfgChan chan<- json.Marshaler) error
	Stop() error
}

type _PP struct {
	IValue   interface{}
	WInit    func() error
	WProvide func(cfgChan chan<- json.Marshaler) error
	WStop    func() error
}

func (p _PP) Init() error {
	return p.WInit()
}

func (p _PP) Provide(cfgChan chan<- json.Marshaler) error {
	return p.WProvide(cfgChan)
}

func (p _PP) Stop() error {
	return p.WStop()
}

func ppSymbols() map[string]map[string]reflect.Value {
	return map[string]map[string]reflect.Value{
		"github.com/traefik/traefik/v3/pkg/plugins/plugins": {
			"PP":  reflect.ValueOf((*PP)(nil)),
			"_PP": reflect.ValueOf((*_PP)(nil)),
		},
	}
}

// BuildProvider builds a plugin's provider.
func (b Builder) BuildProvider(pName string, config map[string]interface{}) (provider.Provider, error) {
	if b.providerBuilders == nil {
		return nil, fmt.Errorf("no plugin definition in the static configuration: %s", pName)
	}

	builder, ok := b.providerBuilders[pName]
	if !ok {
		return nil, fmt.Errorf("unknown plugin type: %s", pName)
	}

	return newProvider(builder, config, "plugin-"+pName)
}

type providerBuilder struct {
	// Import plugin's import/package
	Import string `json:"import,omitempty" toml:"import,omitempty" yaml:"import,omitempty"`

	// BasePkg plugin's base package name (optional)
	BasePkg string `json:"basePkg,omitempty" toml:"basePkg,omitempty" yaml:"basePkg,omitempty"`

	interpreter *interp.Interpreter
}

// Provider is a plugin's provider wrapper.
type Provider struct {
	name string
	pp   PP
}

func newProvider(builder providerBuilder, config map[string]interface{}, providerName string) (*Provider, error) {
	basePkg := builder.BasePkg
	if basePkg == "" {
		basePkg = strings.ReplaceAll(path.Base(builder.Import), "-", "_")
	}

	vConfig, err := builder.interpreter.Eval(basePkg + `.CreateConfig()`)
	if err != nil {
		return nil, fmt.Errorf("failed to eval CreateConfig: %w", err)
	}

	cfg := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToSliceHookFunc(","),
		WeaklyTypedInput: true,
		Result:           vConfig.Interface(),
	}

	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration decoder: %w", err)
	}

	err = decoder.Decode(config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode configuration: %w", err)
	}

	_, err = builder.interpreter.Eval(`package wrapper

import (
	"context"

	` + basePkg + ` "` + builder.Import + `"
	"github.com/traefik/traefik/v3/pkg/plugins"
)

func NewWrapper(ctx context.Context, config *` + basePkg + `.Config, name string) (plugins.PP, error) {
	p, err := ` + basePkg + `.New(ctx, config, name)
	var pv plugins.PP = p
	return pv, err
}
`)
	if err != nil {
		return nil, fmt.Errorf("failed to eval wrapper: %w", err)
	}

	fnNew, err := builder.interpreter.Eval("wrapper.NewWrapper")
	if err != nil {
		return nil, fmt.Errorf("failed to eval New: %w", err)
	}

	ctx := context.Background()

	args := []reflect.Value{reflect.ValueOf(ctx), vConfig, reflect.ValueOf(providerName)}
	results := fnNew.Call(args)

	if len(results) > 1 && results[1].Interface() != nil {
		return nil, results[1].Interface().(error)
	}

	prov, ok := results[0].Interface().(PP)
	if !ok {
		return nil, fmt.Errorf("invalid provider type: %T", results[0].Interface())
	}

	return &Provider{name: providerName, pp: prov}, nil
}

// Init wraps the Init method of a plugin.
func (p *Provider) Init() error {
	return p.pp.Init()
}

// Provide wraps the Provide method of a plugin.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	defer func() {
		if err := recover(); err != nil {
			log.Error().Str(logs.ProviderName, p.name).Msgf("Panic inside the plugin %v", err)
		}
	}()

	cfgChan := make(chan json.Marshaler)

	pool.GoCtx(func(ctx context.Context) {
		logger := log.Ctx(ctx).With().Str(logs.ProviderName, p.name).Logger()

		for {
			select {
			case <-ctx.Done():
				err := p.pp.Stop()
				if err != nil {
					logger.Error().Err(err).Msg("Failed to stop the provider")
				}

				return

			case cfgPg := <-cfgChan:
				marshalJSON, err := cfgPg.MarshalJSON()
				if err != nil {
					logger.Error().Err(err).Msg("Failed to marshal configuration")
					continue
				}

				cfg := &dynamic.Configuration{}
				err = json.Unmarshal(marshalJSON, cfg)
				if err != nil {
					logger.Error().Err(err).Msg("Failed to unmarshal configuration")
					continue
				}

				configurationChan <- dynamic.Message{
					ProviderName:  p.name,
					Configuration: cfg,
				}
			}
		}
	})

	err := p.pp.Provide(cfgChan)
	if err != nil {
		return fmt.Errorf("error from %s: %w", p.name, err)
	}

	return nil
}
