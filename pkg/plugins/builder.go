package plugins

import (
	"context"
	"fmt"
	"net/http"
	"os"

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

		i := interp.New(interp.Options{
			GoPath: client.GoPath(),
			Env:    os.Environ(),
			Stdout: logs.NoLevel(logger, zerolog.DebugLevel),
			Stderr: logs.NoLevel(logger, zerolog.ErrorLevel),
		})

		err = i.Use(stdlib.Symbols)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to load symbols: %w", desc.ModuleName, err)
		}

		err = i.Use(ppSymbols())
		if err != nil {
			return nil, fmt.Errorf("%s: failed to load provider symbols: %w", desc.ModuleName, err)
		}

		_, err = i.Eval(fmt.Sprintf(`import "%s"`, manifest.Import))
		if err != nil {
			return nil, fmt.Errorf("%s: failed to import plugin code %q: %w", desc.ModuleName, manifest.Import, err)
		}

		switch manifest.Type {
		case "middleware":
			middleware, err := newMiddlewareBuilder(i, manifest.BasePkg, manifest.Import)
			if err != nil {
				return nil, err
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

		i := interp.New(interp.Options{
			GoPath: localGoPath,
			Env:    os.Environ(),
			Stdout: logs.NoLevel(logger, zerolog.DebugLevel),
			Stderr: logs.NoLevel(logger, zerolog.ErrorLevel),
		})

		err = i.Use(stdlib.Symbols)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to load symbols: %w", desc.ModuleName, err)
		}

		err = i.Use(ppSymbols())
		if err != nil {
			return nil, fmt.Errorf("%s: failed to load provider symbols: %w", desc.ModuleName, err)
		}

		_, err = i.Eval(fmt.Sprintf(`import "%s"`, manifest.Import))
		if err != nil {
			return nil, fmt.Errorf("%s: failed to import plugin code %q: %w", desc.ModuleName, manifest.Import, err)
		}

		switch manifest.Type {
		case "middleware":
			middleware, err := newMiddlewareBuilder(i, manifest.BasePkg, manifest.Import)
			if err != nil {
				return nil, err
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
