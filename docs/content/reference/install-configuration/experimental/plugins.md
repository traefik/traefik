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
