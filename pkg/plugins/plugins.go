package plugins

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/mod/module"
)

const localGoPath = "./plugins-local/"

// SetupRemotePlugins setup remote plugins environment.
func SetupRemotePlugins(manager *Manager, plugins map[string]Descriptor) error {
	err := checkRemotePluginsConfiguration(plugins)
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	err = manager.CleanArchives(plugins)
	if err != nil {
		return fmt.Errorf("unable to clean archives: %w", err)
	}

	ctx := context.Background()

	for pAlias, desc := range plugins {
		log.Ctx(ctx).Debug().Msgf("Installing plugin: %s: %s@%s", pAlias, desc.ModuleName, desc.Version)

		if err = manager.InstallPlugin(ctx, desc); err != nil {
			_ = manager.ResetAll()
			return fmt.Errorf("unable to install plugin %s: %w", pAlias, err)
		}
	}

	err = manager.WriteState(plugins)
	if err != nil {
		_ = manager.ResetAll()
		return fmt.Errorf("unable to write plugins state: %w", err)
	}

	return nil
}

func checkRemotePluginsConfiguration(plugins map[string]Descriptor) error {
	if plugins == nil {
		return nil
	}

	uniq := make(map[string]struct{})

	var errs []string
	for pAlias, descriptor := range plugins {
		if err := module.CheckPath(descriptor.ModuleName); err != nil {
			errs = append(errs, fmt.Sprintf("%s: malformed plugin module name is missing: %s", pAlias, err))
		}

		if descriptor.Version == "" {
			errs = append(errs, fmt.Sprintf("%s: plugin version is missing", pAlias))
		}

		if _, ok := uniq[descriptor.ModuleName]; ok {
			errs = append(errs, fmt.Sprintf("only one version of a plugin is allowed, there is a duplicate of %s", descriptor.ModuleName))
			continue
		}
		uniq[descriptor.ModuleName] = struct{}{}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ": "))
	}

	return nil
}

// SetupLocalPlugins setup local plugins environment.
func SetupLocalPlugins(plugins map[string]LocalDescriptor) error {
	if plugins == nil {
		return nil
	}

	uniq := make(map[string]struct{})

	var errs []error
	for pAlias, descriptor := range plugins {
		if descriptor.ModuleName == "" {
			errs = append(errs, fmt.Errorf("%s: plugin name is missing", pAlias))
		}

		if strings.HasPrefix(descriptor.ModuleName, "/") || strings.HasSuffix(descriptor.ModuleName, "/") {
			errs = append(errs, fmt.Errorf("%s: plugin name should not start or end with a /", pAlias))
			continue
		}

		if _, ok := uniq[descriptor.ModuleName]; ok {
			errs = append(errs, fmt.Errorf("only one version of a plugin is allowed, there is a duplicate of %s", descriptor.ModuleName))
			continue
		}

		uniq[descriptor.ModuleName] = struct{}{}

		if err := checkLocalPluginManifest(descriptor); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func checkLocalPluginManifest(descriptor LocalDescriptor) error {
	m, err := ReadManifest(localGoPath, descriptor.ModuleName)
	if err != nil {
		return err
	}

	var errs []error

	switch m.Type {
	case typeMiddleware:
		if m.Runtime != runtimeYaegi && m.Runtime != runtimeWasm && m.Runtime != "" {
			errs = append(errs, fmt.Errorf("%s: unsupported runtime '%q'", descriptor.ModuleName, m.Runtime))
		}

	case typeProvider:
		if m.Runtime != runtimeYaegi && m.Runtime != "" {
			errs = append(errs, fmt.Errorf("%s: unsupported runtime '%q'", descriptor.ModuleName, m.Runtime))
		}

	default:
		errs = append(errs, fmt.Errorf("%s: unsupported type %q", descriptor.ModuleName, m.Type))
	}

	if m.IsYaegiPlugin() {
		if m.Import == "" {
			errs = append(errs, fmt.Errorf("%s: missing import", descriptor.ModuleName))
		}

		if !strings.HasPrefix(m.Import, descriptor.ModuleName) {
			errs = append(errs, fmt.Errorf("the import %q must be related to the module name %q", m.Import, descriptor.ModuleName))
		}
	}

	if m.DisplayName == "" {
		errs = append(errs, fmt.Errorf("%s: missing DisplayName", descriptor.ModuleName))
	}

	if m.Summary == "" {
		errs = append(errs, fmt.Errorf("%s: missing Summary", descriptor.ModuleName))
	}

	if m.TestData == nil {
		errs = append(errs, fmt.Errorf("%s: missing TestData", descriptor.ModuleName))
	}

	return errors.Join(errs...)
}
