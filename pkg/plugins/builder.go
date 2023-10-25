package plugins

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// Constructor creates a plugin handler.
type Constructor func(context.Context, http.Handler) (http.Handler, error)

// Builder is a plugin builder.
type Builder struct {
	middlewareBuilders map[string]*middlewareBuilder
	providerBuilders   map[string]providerBuilder
}

// NewBuilder creates a new Builder.
func NewBuilder(client *Client, plugins map[string]Descriptor, localPlugins map[string]LocalDescriptor) (*Builder, error) {
	pb := &Builder{
		middlewareBuilders: map[string]*middlewareBuilder{},
		providerBuilders:   map[string]providerBuilder{},
	}

	for pName, desc := range plugins {
		manifest, err := client.ReadManifest(desc.ModuleName)
		if err != nil {
			_ = client.ResetAll()
			return nil, fmt.Errorf("%s: failed to read manifest: %w", desc.ModuleName, err)
		}

		logger := log.With().Str("plugin", "plugin-"+pName).Str("module", desc.ModuleName).Logger()

		i, wasmPath, err := initMiddlewareBuilder(logger, client.GoPath(), desc.ModuleName, manifest)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", desc.ModuleName, err)
		}

		switch manifest.Type {
		case "middleware":
			var middleware *middlewareBuilder
			if manifest.Runtime == RuntimeWasm {
				middleware = newWasmMiddlewareBuilder(wasmPath)
			} else if manifest.IsYaegiPlugin() {
				middleware, err = newYaegiMiddlewareBuilder(i, manifest.BasePkg, manifest.Import)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("unknow plugin runtime: %s", manifest.Runtime)
			}

			pb.middlewareBuilders[pName] = middleware
		case "provider":
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

		i, wasmPath, err := initMiddlewareBuilder(logger, localGoPath, desc.ModuleName, manifest)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", desc.ModuleName, err)
		}

		switch manifest.Type {
		case "middleware":
			var middleware *middlewareBuilder
			if manifest.Runtime == RuntimeWasm {
				middleware = newWasmMiddlewareBuilder(wasmPath)
			} else if manifest.IsYaegiPlugin() {
				middleware, err = newYaegiMiddlewareBuilder(i, manifest.BasePkg, manifest.Import)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("unknow plugin runtime: %s", manifest.Runtime)
			}
			pb.middlewareBuilders[pName] = middleware
		case "provider":
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

func initMiddlewareBuilder(logger zerolog.Logger, goPath string, moduleName string, manifest *Manifest) (*interp.Interpreter, string, error) {
	if !manifest.IsYaegiPlugin() {
		return nil, filepath.Join(goPath, "src", moduleName, manifest.WasmPath), nil
	}
	i, err := initInterp(logger, goPath, manifest.Import)
	return i, "", err
}

func initInterp(logger zerolog.Logger, goPath string, manifestImport string) (*interp.Interpreter, error) {
	i := interp.New(interp.Options{
		GoPath: goPath,
		Env:    os.Environ(),
		Stdout: logs.NoLevel(logger, zerolog.DebugLevel),
		Stderr: logs.NoLevel(logger, zerolog.ErrorLevel),
	})

	err := i.Use(stdlib.Symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to load symbols: %w", err)
	}

	err = i.Use(ppSymbols())
	if err != nil {
		return nil, fmt.Errorf("failed to load provider symbols: %w", err)
	}

	_, err = i.Eval(fmt.Sprintf(`import "%s"`, manifestImport))
	if err != nil {
		return nil, fmt.Errorf("failed to import plugin code %q: %w", manifestImport, err)
	}
	return i, nil
}
