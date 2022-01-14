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
)

// Build builds a middleware plugin.
func (b Builder) Build(pName string, config map[string]interface{}, middlewareName string) (Constructor, error) {
	if b.middlewareBuilders == nil {
		return nil, fmt.Errorf("no plugin definition in the static configuration: %s", pName)
	}

	descriptor, ok := b.middlewareBuilders[pName]
	if !ok {
		return nil, fmt.Errorf("unknown plugin type: %s", pName)
	}

	m, err := newMiddleware(descriptor, config, middlewareName)
	if err != nil {
		return nil, err
	}

	return m.NewHandler, err
}

type middlewareBuilder struct {
	fnNew          reflect.Value
	fnCreateConfig reflect.Value
}

func newMiddlewareBuilder(i *interp.Interpreter, basePkg, imp string) (*middlewareBuilder, error) {
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

	return &middlewareBuilder{
		fnNew:          fnNew,
		fnCreateConfig: fnCreateConfig,
	}, nil
}

func (p middlewareBuilder) newHandler(ctx context.Context, next http.Handler, cfg reflect.Value, middlewareName string) (http.Handler, error) {
	args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(next), cfg, reflect.ValueOf(middlewareName)}
	results := p.fnNew.Call(args)

	if len(results) > 1 && results[1].Interface() != nil {
		return nil, results[1].Interface().(error)
	}

	handler, ok := results[0].Interface().(http.Handler)
	if !ok {
		return nil, fmt.Errorf("invalid handler type: %T", results[0].Interface())
	}

	return handler, nil
}

func (p middlewareBuilder) createConfig(config map[string]interface{}) (reflect.Value, error) {
	results := p.fnCreateConfig.Call(nil)
	if len(results) != 1 {
		return reflect.Value{}, fmt.Errorf("invalid return of the CreateConfig function: %d", len(results))
	}

	vConfig := results[0]

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

// Middleware is an HTTP handler plugin wrapper.
type Middleware struct {
	middlewareName string
	config         reflect.Value
	pm             *middlewareBuilder
}

func newMiddleware(descriptor *middlewareBuilder, config map[string]interface{}, middlewareName string) (*Middleware, error) {
	vConfig, err := descriptor.createConfig(config)
	if err != nil {
		return nil, err
	}

	return &Middleware{
		middlewareName: middlewareName,
		config:         vConfig,
		pm:             descriptor,
	}, nil
}

// NewHandler creates a new HTTP handler.
func (m *Middleware) NewHandler(ctx context.Context, next http.Handler) (http.Handler, error) {
	return m.pm.newHandler(ctx, next, m.config, m.middlewareName)
}
