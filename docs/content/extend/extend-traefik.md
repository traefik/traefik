---
title: Extend Traefik
description: Extend Traefik with custom plugins using Yaegi and WebAssembly.
---

# Extend Traefik

Plugins are a powerful feature for extending Traefik with custom features and behaviors. The [Plugin Catalog](https://plugins.traefik.io/) is a software-as-a-service (SaaS) platform that provides an exhaustive list of the existing plugins.

??? note "Plugin Catalog Access"
    You can reach the [Plugin Catalog](https://plugins.traefik.io/) from the Traefik Dashboard using the `Plugins` menu entry.

## Add a new plugin to a Traefik instance

To add a new plugin to a Traefik instance, you must change that instance's install (static) configuration. Each plugin's **Install** section provides an install (static) configuration example. Many plugins have their own section in the Traefik routing (dynamic) configuration.

!!! danger "Experimental Features"
    Plugins can change the behavior of Traefik in unforeseen ways. Exercise caution when adding new plugins to production Traefik instances.

To learn more about how to add a new plugin to a Traefik instance, please refer to the [developer documentation](https://plugins.traefik.io/install).

## Plugin Systems

Traefik supports two different plugin systems, each designed for different use cases and developer preferences.

### Yaegi Plugin System

Traefik [Yaegi](https://github.com/traefik/yaegi) plugins are developed using the Go language. It is essentially a Go package. Unlike pre-compiled plugins, Yaegi plugins are executed on the fly by Yaegi, a Go interpreter embedded in Traefik.

This approach eliminates the need for compilation and a complex toolchain, making plugin development as straightforward as creating web browser extensions. Yaegi plugins support both middleware and provider functionality.

#### Key characteristics

- Written in Go language
- No compilation required
- Executed by embedded interpreter
- Supports full Go feature set
- Hot-reloadable during development

### WebAssembly (WASM) Plugin System

Traefik WASM plugins can be developed using any language that compiles to WebAssembly (WASM). This method is based on [http-wasm](https://http-wasm.io/).

WASM plugins compile to portable binary modules that execute with near-native performance while maintaining security isolation.

#### Key characteristics

- Multi-language support (Go, Rust, C++, etc.)
- Compiled to WebAssembly binary
- Near-native performance
- Strong security isolation
- Currently supports middleware only

## Build Your Own Plugins

Traefik users can create their own plugins and share them with the community using the [Plugin Catalog](https://plugins.traefik.io/). To learn more about Traefik plugin creation, please refer to the [developer documentation](https://plugins.traefik.io/create).
