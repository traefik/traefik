package plugins

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Constructor creates a plugin handler.
type Constructor func(context.Context, http.Handler) (http.Handler, error)

type pluginMiddleware interface {
	NewHandler(context.Context, http.Handler) (http.Handler, error)
}

type middlewareBuilder interface {
	newMiddleware(config map[string]interface{}, middlewareName string) (pluginMiddleware, error)
}

// Builder is a plugin builder.
type Builder struct {
	providerBuilders   map[string]providerBuilder
	middlewareBuilders map[string]middlewareBuilder
}

// NewBuilder creates a new Builder.
func NewBuilder(client *Client, plugins map[string]Descriptor, localPlugins map[string]LocalDescriptor) (*Builder, error) {
	pb := &Builder{
		middlewareBuilders: map[string]middlewareBuilder{},
		providerBuilders:   map[string]providerBuilder{},
	}

	for pName, desc := range plugins {
		manifest, err := client.ReadManifest(desc.ModuleName)
		if err != nil {
			_ = client.ResetAll()
			return nil, fmt.Errorf("%s: failed to read manifest: %w", desc.ModuleName, err)
		}

		logger := log.With().Str("plugin", "plugin-"+pName).Str("module", desc.ModuleName).Logger()

		switch manifest.Type {
		case middleware:
			switch manifest.Runtime {
			case RuntimeWasm:
				pb.middlewareBuilders[pName] = newWasmMiddlewareBuilder(client.GoPath(), desc.ModuleName, manifest)
			case RuntimeYaegi, "":
				middleware, err := newYaegiMiddlewareBuilder(logger, client.GoPath(), manifest)
				if err != nil {
					return nil, err
				}
				pb.middlewareBuilders[pName] = middleware
			default:
				return nil, fmt.Errorf("unknow plugin runtime: %s", manifest.Runtime)
			}
		case provider:
			i, err := initInterp(logger, client.GoPath(), manifest.Import)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", desc.ModuleName, err)
			}

			pb.providerBuilders[pName] = providerBuilder{
				interpreter: i,
				Import:      manifest.Import,
				BasePkg:     manifest.BasePkg,
			}
		default:
			return nil, fmt.Errorf("unknow plugin type: %s", manifest.Type)
		}
	}

	for pName, desc := range localPlugins {
		manifest, err := ReadManifest(localGoPath, desc.ModuleName)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to read manifest: %w", desc.ModuleName, err)
		}

		logger := log.With().Str("plugin", "plugin-"+pName).Str("module", desc.ModuleName).Logger()

		switch manifest.Type {
		case middleware:
			switch manifest.Runtime {
			case RuntimeWasm:
				pb.middlewareBuilders[pName] = newWasmMiddlewareBuilder(localGoPath, desc.ModuleName, manifest)
			case RuntimeYaegi, "":
				middleware, err := newYaegiMiddlewareBuilder(logger, localGoPath, manifest)
				if err != nil {
					return nil, err
				}
				pb.middlewareBuilders[pName] = middleware
			default:
				return nil, fmt.Errorf("unknow plugin runtime: %s", manifest.Runtime)
			}
		case provider:
			i, err := initInterp(logger, localGoPath, manifest.Import)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", desc.ModuleName, err)
			}

			pb.providerBuilders[pName] = providerBuilder{
				interpreter: i,
				Import:      manifest.Import,
				BasePkg:     manifest.BasePkg,
			}
		default:
			return nil, fmt.Errorf("unknow plugin type: %s", manifest.Type)
		}
	}
	return pb, nil
}

// Build builds a middleware plugin.
func (b Builder) Build(pName string, config map[string]interface{}, middlewareName string) (Constructor, error) {
	if b.middlewareBuilders == nil {
		return nil, fmt.Errorf("no plugin definitions in the static configuration: %s", pName)
	}

	// plugin (pName) can be located in yaegi or wasm middleware builders.
	if descriptor, ok := b.middlewareBuilders[pName]; ok {
		m, err := descriptor.newMiddleware(config, middlewareName)
		if err != nil {
			return nil, err
		}
		return m.NewHandler, nil
	}

	return nil, fmt.Errorf("unknown plugin type: %s", pName)
}
