---
title: "Traefik Plugins Experimental Configuration"
description: "This section of the Traefik Proxy documentation explains how to use the new Plugins install configuration option."
---

# Traefik Plugins Experimental Configuration

## Overview

This guide provides instructions on how to configure and use the new experimental `plugins` install configuration option in Traefik. The `plugins` option introduces a system to extend Traefik capabilities with custom middlewares and providers.

!!! warning "Experimental"
    
    The `plugins` option is currently experimental and subject to change in future releases. 
    Use with caution in production environments.

## Enabling Plugins

The plugins option is an install configuration parameter.
To enable a plugin, you need to define it in your Traefik install configuration

```yaml tab="File (YAML)"
experimental:
  plugins:
    plugin-name: # The name of the plugin in the routing configuration
      moduleName: "github.com/github-organization/github-repository" # The plugin module name
      version: "vX.XX.X" # The version to use
```

```toml tab="File (TOML)"
[experimental.plugins.plugin-name]
  moduleName = "github.com/github-organization/github-repository" # The plugin module name
  version = "vX.XX.X" # The version to use
```

```bash tab="CLI"
# The plugin module name
# With plugin-name the name of the plugin in the routing configuration
--experimental.plugins.plugin-name.modulename=github.com/github-organization/github-repository
--experimental.plugins.plugin-name.version=vX.XX.X # The version to use
```

To learn more about how to add a new plugin to a Traefik instance, please refer to the [developer documentation](https://plugins.traefik.io/install).

### Plugin Options

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| <a id="opt-moduleName" href="#opt-moduleName" title="#opt-moduleName">`moduleName`</a> | Plugin's module name. | string | Yes |
| <a id="opt-version" href="#opt-version" title="#opt-version">`version`</a> | Plugin's version. | string | Yes |
| <a id="opt-hash" href="#opt-hash" title="#opt-hash">`hash`</a> | Plugin's hash to validate. | string | No |
| <a id="opt-settings" href="#opt-settings" title="#opt-settings">`settings`</a> | Plugin's settings (works only for wasm plugins). | object | No |
| <a id="opt-settings-envs" href="#opt-settings-envs" title="#opt-settings-envs">`settings.envs`</a> | Environment variables to forward to the wasm guest. | []string | No |
| <a id="opt-settings-mounts" href="#opt-settings-mounts" title="#opt-settings-mounts">`settings.mounts`</a> | Directory to mount to the wasm guest. | []string | No |
| <a id="opt-settings-useUnsafe" href="#opt-settings-useUnsafe" title="#opt-settings-useUnsafe">`settings.useUnsafe`</a> | Allow the plugin to use unsafe and syscall packages. | bool | No |

## Local Plugins

Local plugins allow you to use plugins from a local directory, without publishing them to the Traefik plugin catalog.

```yaml tab="File (YAML)"
experimental:
  localPlugins:
    plugin-name: # The name of the plugin in the routing configuration
      moduleName: "github.com/github-organization/github-repository" # The plugin module name
```

```toml tab="File (TOML)"
[experimental.localPlugins.plugin-name]
  moduleName = "github.com/github-organization/github-repository" # The plugin module name
```

```bash tab="CLI"
# The plugin module name
# With plugin-name the name of the plugin in the routing configuration
--experimental.localplugins.plugin-name.modulename=github.com/github-organization/github-repository
```

### Local Plugin Options

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| <a id="opt-moduleName-2" href="#opt-moduleName-2" title="#opt-moduleName-2">`moduleName`</a> | Plugin's module name. | string | Yes |
| <a id="opt-settings-2" href="#opt-settings-2" title="#opt-settings-2">`settings`</a> | Plugin's settings (works only for wasm plugins). | object | No |
| <a id="opt-settings-envs-2" href="#opt-settings-envs-2" title="#opt-settings-envs-2">`settings.envs`</a> | Environment variables to forward to the wasm guest. | []string | No |
| <a id="opt-settings-mounts-2" href="#opt-settings-mounts-2" title="#opt-settings-mounts-2">`settings.mounts`</a> | Directory to mount to the wasm guest. | []string | No |
| <a id="opt-settings-useUnsafe-2" href="#opt-settings-useUnsafe-2" title="#opt-settings-useUnsafe-2">`settings.useUnsafe`</a> | Allow the plugin to use unsafe and syscall packages. | bool | No |
