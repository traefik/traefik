package plugins

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// Constructor creates a plugin handler.
type Constructor func(context.Context, http.Handler) (http.Handler, error)

type pluginMiddleware interface {
	NewHandler(ctx context.Context, next http.Handler) (http.Handler, error)
}

type middlewareBuilder interface {
	newMiddleware(config map[string]any, middlewareName string) (pluginMiddleware, error)
}

// Builder is a plugin builder.
type Builder struct {
	providerBuilders   map[string]providerBuilder
	middlewareBuilders map[string]middlewareBuilder
}

// NewBuilder creates a new Builder.
func NewBuilder(manager *Manager, plugins map[string]Descriptor, localPlugins map[string]LocalDescriptor, defaultRuntimeVersion string) (*Builder, error) {
	ctx := context.Background()

	pb := &Builder{
		middlewareBuilders: map[string]middlewareBuilder{},
		providerBuilders:   map[string]providerBuilder{},
	}

	for pName, desc := range plugins {
		manifest, err := manager.ReadManifest(desc.ModuleName)
		if err != nil {
			_ = manager.ResetAll()
			return nil, fmt.Errorf("%s: failed to read manifest: %w", desc.ModuleName, err)
		}

		logger := log.With().
			Str("plugin", "plugin-"+pName).
			Str("module", desc.ModuleName).
			Str("runtime", manifest.Runtime).
			Logger()
		logCtx := logger.WithContext(ctx)

		switch manifest.Type {
		case typeMiddleware:
			middleware, err := newMiddlewareBuilder(logCtx, manager.GoPath(), manifest, desc.ModuleName, desc.Settings, defaultRuntimeVersion, false)
			if err != nil {
				return nil, err
			}

			pb.middlewareBuilders[pName] = middleware

		case typeProvider:
			pBuilder, err := newProviderBuilder(logCtx, manifest, manager.GoPath(), desc.ModuleName, desc.Settings, defaultRuntimeVersion, false)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", desc.ModuleName, err)
			}

			pb.providerBuilders[pName] = pBuilder

		default:
			return nil, fmt.Errorf("unknown plugin type: %s", manifest.Type)
		}
	}

	for pName, desc := range localPlugins {
		manifest, err := ReadManifest(localGoPath, desc.ModuleName)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to read manifest: %w", desc.ModuleName, err)
		}

		logger := log.With().
			Str("plugin", "plugin-"+pName).
			Str("module", desc.ModuleName).
			Str("runtime", manifest.Runtime).
			Logger()
		logCtx := logger.WithContext(ctx)

		switch manifest.Type {
		case typeMiddleware:
			middleware, err := newMiddlewareBuilder(logCtx, localGoPath, manifest, desc.ModuleName, desc.Settings, defaultRuntimeVersion, true)
			if err != nil {
				return nil, err
			}

			pb.middlewareBuilders[pName] = middleware

		case typeProvider:
			builder, err := newProviderBuilder(logCtx, manifest, localGoPath, desc.ModuleName, desc.Settings, defaultRuntimeVersion, true)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", desc.ModuleName, err)
			}

			pb.providerBuilders[pName] = builder

		default:
			return nil, fmt.Errorf("unknown plugin type: %s", manifest.Type)
		}
	}
	return pb, nil
}

// Build builds a middleware plugin.
func (b Builder) Build(pName string, config map[string]any, middlewareName string) (Constructor, error) {
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

// resolveYaegiVersion returns the version of the Yaegi interpreter that runs a plugin.
// defaultVersion is the experimental defaultRuntimeVersion option.
func resolveYaegiVersion(defaultVersion string, manifest *Manifest) (string, error) {
	switch manifest.RuntimeVersion {
	case "":
	case YaegiV2:
		return YaegiV2, nil
	default:
		return "", fmt.Errorf("invalid runtimeVersion in the plugin manifest: %s (only %s is allowed)", manifest.RuntimeVersion, YaegiV2)
	}

	switch defaultVersion {
	case YaegiV1, YaegiV2:
		return defaultVersion, nil
	default:
		return "", fmt.Errorf("unknown Yaegi version: %q", defaultVersion)
	}
}

func newMiddlewareBuilder(ctx context.Context, goPath string, manifest *Manifest, moduleName string, settings Settings, defaultRuntimeVersion string, local bool) (middlewareBuilder, error) {
	switch manifest.Runtime {
	case runtimeWasm:
		wasmPath, err := getWasmPath(manifest)
		if err != nil {
			return nil, fmt.Errorf("wasm path: %w", err)
		}

		return newWasmMiddlewareBuilder(goPath, moduleName, wasmPath, settings)

	case runtimeYaegi, "":
		version, err := resolveYaegiVersion(defaultRuntimeVersion, manifest)
		if err != nil {
			return nil, fmt.Errorf("resolving Yaegi version: %w", err)
		}

		if version == YaegiV2 {
			plugin, err := loadYaegiV2Plugin(ctx, goPath, manifest, moduleName, settings, local)
			if err != nil {
				return nil, fmt.Errorf("loading plugin with Yaegi v2: %w", err)
			}

			return yaegiV2MiddlewareBuilder{plugin: plugin}, nil
		}

		i, err := newInterpreter(ctx, goPath, manifest, settings)
		if err != nil {
			return nil, fmt.Errorf("failed to create Yaegi interpreter: %w", err)
		}

		return newYaegiMiddlewareBuilder(i, manifest.BasePkg, manifest.Import)

	default:
		return nil, fmt.Errorf("unknown plugin runtime: %s", manifest.Runtime)
	}
}

func newProviderBuilder(ctx context.Context, manifest *Manifest, goPath, moduleName string, settings Settings, defaultRuntimeVersion string, local bool) (providerBuilder, error) {
	switch manifest.Runtime {
	case runtimeYaegi, "":
		version, err := resolveYaegiVersion(defaultRuntimeVersion, manifest)
		if err != nil {
			return providerBuilder{}, fmt.Errorf("resolving Yaegi version: %w", err)
		}

		if version == YaegiV2 {
			plugin, err := loadYaegiV2Plugin(ctx, goPath, manifest, moduleName, settings, local)
			if err != nil {
				return providerBuilder{}, fmt.Errorf("loading plugin with Yaegi v2: %w", err)
			}

			return providerBuilder{plugin: plugin}, nil
		}

		i, err := newInterpreter(ctx, goPath, manifest, settings)
		if err != nil {
			return providerBuilder{}, err
		}

		return providerBuilder{
			interpreter: i,
			Import:      manifest.Import,
			BasePkg:     manifest.BasePkg,
		}, nil

	default:
		return providerBuilder{}, fmt.Errorf("unknown plugin runtime: %s", manifest.Runtime)
	}
}

func getWasmPath(manifest *Manifest) (string, error) {
	wasmPath := manifest.WasmPath
	if wasmPath == "" {
		wasmPath = "plugin.wasm"
	}

	if !filepath.IsLocal(wasmPath) {
		return "", errors.New("wasmPath must be a local path")
	}

	return wasmPath, nil
}
