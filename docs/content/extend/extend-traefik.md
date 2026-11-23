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

#### Yaegi v2

!!! info "Opt-in Runtime"

    Yaegi v2 is available since Traefik v3.8.
    Yaegi plugins still run with Yaegi v1 by default, and you choose which plugins run with Yaegi v2.

Yaegi v2 is a new Go interpreter written from scratch for Traefik plugins.
It relies on the official Go tools to parse and type-check the plugin code,
so a plugin that the Go compiler accepts gives the same result in Traefik as when it is compiled.

Yaegi v2 brings:

- The full Go language, including generics, iterators, and the latest built-in functions.
- Plugins that behave like compiled Go, including with libraries that use `reflect`, such as `encoding/json`, `fmt`, or `text/template`.
- Support for large libraries such as gRPC, protobuf, the AWS SDK, OpenTelemetry, or the Kubernetes client
  (most of them require `useUnsafe`).
- Faster request processing for most plugins.
- Guardrails that limit the impact of a plugin bug on Traefik: execution budget, goroutine limit, memory quota,
  and isolation of `os.Exit`, `log.Fatal`, and panics in plugin goroutines.

!!! note "Not a sandbox"

    A Yaegi v2 plugin still runs inside the Traefik process and shares its memory.
    The guardrails protect Traefik from plugin bugs, not from a malicious plugin.

To run a plugin with Yaegi v2, either:

- Set `runtimeVersion: v2` in the plugin manifest (`.traefik.yml`), so that this plugin always runs with Yaegi v2.
- Set the `experimental.defaultRuntimeVersion` option to `v2` in the install configuration, so that all the Yaegi plugins run with Yaegi v2.

```yaml tab="File (YAML)"
experimental:
  defaultRuntimeVersion: v2
```

```toml tab="File (TOML)"
[experimental]
  defaultRuntimeVersion = "v2"
```

```bash tab="CLI"
--experimental.defaultruntimeversion=v2
```

Yaegi v2 only loads plugins that the Go compiler accepts.
If you maintain a plugin, check that `go build` and `go vet` pass before switching it to Yaegi v2.

For more information about the runtime selection and the Yaegi v2 specific settings (`mounts`, `fileMapping`),
see the [plugins reference documentation](../reference/install-configuration/experimental/plugins.md#runtime-version).

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
