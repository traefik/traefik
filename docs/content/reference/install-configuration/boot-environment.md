---
title: "Traefik Configuration Overview"
description: "Read the official Traefik documentation to get started with configuring the Traefik Proxy."
---

# Boot Environment

Traefik Proxy’s configuration is divided into two main categories:

- **Static Configuration**: Defines parameters that require Traefik to restart when changed. This includes entry points, providers, API/dashboard settings, and logging levels.
- **Dynamic Configuration**: Involves elements that can be updated without restarting Traefik, such as routers, services, and middlewares.

This section focuses on setting up the static configuration, which is essential for Traefik’s initial boot.

## Configuration Methods

Traefik offers multiple methods to define static configuration. 

!!! warning "Note"
    It’s crucial to choose one method and stick to it, as mixing different configuration options is not supported and can lead to unexpected behavior.

Here are the methods available for configuring the Traefik proxy:

- [File](#file) 
- [CLI](#cli)
- [Environment Variables](#environment-variables)
- [Helm](#helm)

## File

You can define the static configuration in a file using formats like YAML or TOML.

### Configuration Example

```yaml tab="traefik.yml (YAML)"
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

```toml tab="traefik.toml (TOML)"
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

At startup, Traefik searches for static configuration in a file named `traefik.yml` (or `traefik.yaml` or `traefik.toml`) in the following directories:

- `/etc/traefik/`
- `$XDG_CONFIG_HOME/`
- `$HOME/.config/`
- `.` (the current working directory).

You can override this behavior using the `configFile` argument like this:

```bash
traefik --configFile=foo/bar/myconfigfile.yml
```

## CLI

Using the CLI, you can pass static configuration directly as command-line arguments when starting Traefik. 

### Configuration Example

```sh tab="CLI"
traefik \
  --entryPoints.web.address=":80" \
  --entryPoints.websecure.address=":443" \
  --providers.docker \
  --api.dashboard \
  --log.level=INFO
```

## Environment Variables

You can also set the static configuration using environment variables. Each option corresponds to an environment variable prefixed with `TRAEFIK_`.

### Configuration Example

```sh tab="ENV"
TRAEFIK_ENTRYPOINTS_WEB_ADDRESS=":80" TRAEFIK_ENTRYPOINTS_WEBSECURE_ADDRESS=":443" TRAEFIK_PROVIDERS_DOCKER=true TRAEFIK_API_DASHBOARD=true TRAEFIK_LOG_LEVEL="INFO" traefik
```

## Helm

When deploying Traefik Proxy using Helm in a Kubernetes cluster, the static configuration is defined in a `values.yaml` file. 

You can find the official Traefik Helm chart on [GitHub](https://github.com/traefik/traefik-helm-chart/blob/master/traefik/VALUES.md)

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
helm repo add traefik https://traefik.github.io/charts
helm repo update
helm install traefik traefik/traefik -f values.yaml
```
