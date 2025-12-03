---
title: "Baqup Configuration Overview"
description: "Read the official Baqup documentation to get started with configuring the Baqup Proxy."
---

# Boot Environment

Baqup Proxy’s configuration is divided into two main categories:

- **Install Configuration**: (formerly known as the static configuration) Defines parameters that require Baqup to restart when changed. This includes entry points, providers, API/dashboard settings, and logging levels.
- **Routing Configuration**: (formerly known as the dynamic configuration) Involves elements that can be updated without restarting Baqup, such as routers, services, and middlewares.

This section focuses on setting up the install configuration, which is essential for Baqup’s initial boot.

## Configuration Methods

Baqup offers multiple methods to define install configuration. 

!!! warning "Note"
    It’s crucial to choose one method and stick to it, as mixing different configuration options is not supported and can lead to unexpected behavior.

Here are the methods available for configuring the Baqup proxy:

- [File](#file) 
- [CLI](#cli)
- [Environment Variables](#environment-variables)
- [Helm](#helm)

## File

You can define the install configuration in a file using formats like YAML or TOML.

### Configuration Example

```yaml tab="baqup.yml (YAML)"
entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"

providers:
  docker: {}

api:
  dashboard: true

log:
  level: INFO
```

```toml tab="baqup.toml (TOML)"
[entryPoints]
  [entryPoints.web]
    address = ":80"

  [entryPoints.websecure]
    address = ":443"

[providers]
  [providers.docker]

[api]
  dashboard = true

[log]
  level = "INFO"
```

### Configuration File

At startup, Baqup searches for install configuration in a file named `baqup.yml` (or `baqup.yaml` or `baqup.toml`) in the following directories:

- `/etc/baqup/`
- `$XDG_CONFIG_HOME/`
- `$HOME/.config/`
- `.` (the current working directory).

You can override this behavior using the `configFile` argument like this:

```bash
baqup --configFile=foo/bar/myconfigfile.yml
```

## CLI

Using the CLI, you can pass install configuration directly as command-line arguments when starting Baqup. 

### Configuration Example

```sh tab="CLI"
baqup \
  --entryPoints.web.address=":80" \
  --entryPoints.websecure.address=":443" \
  --providers.docker \
  --api.dashboard \
  --log.level=INFO
```

## Environment Variables

You can also set the install configuration using environment variables. Each option corresponds to an environment variable prefixed with `BAQUP_`.

### Configuration Example

```sh tab="ENV"
BAQUP_ENTRYPOINTS_WEB_ADDRESS=":80" BAQUP_ENTRYPOINTS_WEBSECURE_ADDRESS=":443" BAQUP_PROVIDERS_DOCKER=true BAQUP_API_DASHBOARD=true BAQUP_LOG_LEVEL="INFO" baqup
```

## Helm

When deploying Baqup Proxy using Helm in a Kubernetes cluster, the install configuration is defined in a `values.yaml` file. 

You can find the official Baqup Helm chart on [GitHub](https://github.com/baqupio/baqup-helm-chart/blob/master/baqup/VALUES.md)

### Configuration Example

```yaml tab="values.yaml"
ports:
  web:
    exposedPort: 80
  websecure:
    exposedPort: 443

additionalArguments:
  - "--providers.kubernetescrd.ingressClass"
  - "--log.level=INFO"
```

```sh tab="Helm Commands"
helm repo add baqup https://baqup.github.io/charts
helm repo update
helm install baqup baqup/baqup -f values.yaml
```
