package plugins

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

const devPluginName = "dev"

// Constructor creates a plugin handler.
type Constructor func(context.Context, http.Handler) (http.Handler, error)

// pluginContext The static part of a plugin configuration.
type pluginContext struct {
	// GoPath plugin's GOPATH
	GoPath string `json:"goPath,omitempty" toml:"goPath,omitempty" yaml:"goPath,omitempty"`

	// Import plugin's import/package
	Import string `json:"import,omitempty" toml:"import,omitempty" yaml:"import,omitempty"`

	// BasePkg plugin's base package name (optional)
	BasePkg string `json:"basePkg,omitempty" toml:"basePkg,omitempty" yaml:"basePkg,omitempty"`

	interpreter *interp.Interpreter
}

// Builder is a plugin builder.
type Builder struct {
	descriptors map[string]pluginContext
}

// NewBuilder creates a new Builder.
func NewBuilder(client *Client, plugins map[string]Descriptor, devPlugin *DevPlugin) (*Builder, error) {
	pb := &Builder{
		descriptors: map[string]pluginContext{},
	}

	for pName, desc := range plugins {
		manifest, err := client.ReadManifest(desc.ModuleName)
		if err != nil {
			_ = client.ResetAll()
			return nil, fmt.Errorf("%s: failed to read manifest: %w", desc.ModuleName, err)
		}

		i := interp.New(interp.Options{GoPath: client.GoPath()})
		i.Use(stdlib.Symbols)

		_, err = i.Eval(fmt.Sprintf(`import "%s"`, manifest.Import))
		if err != nil {
			return nil, fmt.Errorf("%s: failed to import plugin code %q: %w", desc.ModuleName, manifest.Import, err)
		}

		pb.descriptors[pName] = pluginContext{
			interpreter: i,
			GoPath:      client.GoPath(),
			Import:      manifest.Import,
			BasePkg:     manifest.BasePkg,
		}
	}

	if devPlugin != nil {
		manifest, err := ReadManifest(devPlugin.GoPath, devPlugin.ModuleName)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to read manifest: %w", devPlugin.ModuleName, err)
		}

		i := interp.New(interp.Options{GoPath: devPlugin.GoPath})
		i.Use(stdlib.Symbols)

		_, err = i.Eval(fmt.Sprintf(`import "%s"`, manifest.Import))
		if err != nil {
			return nil, fmt.Errorf("%s: failed to import plugin code %q: %w", devPlugin.ModuleName, manifest.Import, err)
		}

		pb.descriptors[devPluginName] = pluginContext{
			interpreter: i,
			GoPath:      devPlugin.GoPath,
			Import:      manifest.Import,
			BasePkg:     manifest.BasePkg,
		}
	}

	return pb, nil
}

// Build builds a plugin.
func (b Builder) Build(pName string, config map[string]interface{}, middlewareName string) (Constructor, error) {
	if b.descriptors == nil {
		return nil, fmt.Errorf("plugin: no plugin definition in the static configuration: %s", pName)
	}

	descriptor, ok := b.descriptors[pName]
	if !ok {
		return nil, fmt.Errorf("plugin: unknown plugin type: %s", pName)
	}

	m, err := newMiddleware(descriptor, config, middlewareName)
	if err != nil {
		return nil, err
	}

	return m.NewHandler, err
}

// Middleware is a HTTP handler plugin wrapper.
type Middleware struct {
	middlewareName string
	fnNew          reflect.Value
	config         reflect.Value
}

func newMiddleware(descriptor pluginContext, config map[string]interface{}, middlewareName string) (*Middleware, error) {
	basePkg := descriptor.BasePkg
	if basePkg == "" {
		basePkg = strings.ReplaceAll(path.Base(descriptor.Import), "-", "_")
	}

	vConfig, err := descriptor.interpreter.Eval(basePkg + `.CreateConfig()`)
	if err != nil {
		return nil, fmt.Errorf("plugin: failed to eval CreateConfig: %w", err)
	}

	cfg := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToSliceHookFunc(","),
		WeaklyTypedInput: true,
		Result:           vConfig.Interface(),
	}

	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return nil, fmt.Errorf("plugin: failed to create configuration decoder: %w", err)
	}

	err = decoder.Decode(config)
	if err != nil {
		return nil, fmt.Errorf("plugin: failed to decode configuration: %w", err)
	}

	fnNew, err := descriptor.interpreter.Eval(basePkg + `.New`)
	if err != nil {
		return nil, fmt.Errorf("plugin: failed to eval New: %w", err)
	}

	return &Middleware{
		middlewareName: middlewareName,
		fnNew:          fnNew,
		config:         vConfig,
	}, nil
}

// NewHandler creates a new HTTP handler.
func (m *Middleware) NewHandler(ctx context.Context, next http.Handler) (http.Handler, error) {
	args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(next), m.config, reflect.ValueOf(m.middlewareName)}
	results := m.fnNew.Call(args)

	if len(results) > 1 && results[1].Interface() != nil {
		return nil, results[1].Interface().(error)
	}

	handler, ok := results[0].Interface().(http.Handler)
	if !ok {
		return nil, fmt.Errorf("plugin: invalid handler type: %T", results[0].Interface())
	}

	return handler, nil
}
