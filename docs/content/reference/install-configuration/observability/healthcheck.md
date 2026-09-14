---
title: "Traefik Health Check Documentation"
description: "In Traefik Proxy, CLI, Ping & Ready let you check the health and readiness of your Traefik instances. Read the technical documentation for configuration examples and options."
---

# CLI, Ping & Ready

Checking the Health and Readiness of your Traefik Instances
{: .subtitle }

## CLI

The CLI can be used to make a request to the `/ping` endpoint to check the health of Traefik. Its exit status is `0` if Traefik is healthy and `1` otherwise.

This can be used with [HEALTHCHECK](https://docs.docker.com/engine/reference/builder/#healthcheck) instruction or any other health check orchestration mechanism.

### Usage 

```sh
traefik healthcheck [command] [flags] [arguments]
```

Example:

```sh
$ traefik healthcheck
OK: http://:8082/ping
```

## Ping

The `/ping` health-check URL is enabled with the command-line `--ping` or config file option `[ping]`.

The entryPoint where the `/ping` is active can be customized with the `entryPoint` option,
whose default value is `traefik` (port `8080`).

| Path    | Method        | Description                                                                                         |
|---------|---------------|-----------------------------------------------------------------------------------------------------|
| <a id="opt-ping" href="#opt-ping" title="#opt-ping">`/ping`</a> | `GET`, `HEAD` | An endpoint to check for Traefik process liveness. Return a code `200` with the content: `OK` |

### Configuration Example

To enable the API handler:

```yaml tab="File (YAML)"
ping: {}
```

```toml tab="File (TOML)"
[ping]
```

```bash tab="CLI"
--ping=true
```

### Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| <a id="opt-ping-entryPoint" href="#opt-ping-entryPoint" title="#opt-ping-entryPoint">`ping.entryPoint`</a> | Enables `/ping` on a dedicated EntryPoint. | traefik  | No   |
| <a id="opt-ping-manualRouting" href="#opt-ping-manualRouting" title="#opt-ping-manualRouting">`ping.manualRouting`</a> | Disables the default internal router in order to allow one to create a custom router for the `ping@internal` service when set to `true`. | false | No   |
| <a id="opt-ping-terminatingStatusCode" href="#opt-ping-terminatingStatusCode" title="#opt-ping-terminatingStatusCode">`ping.terminatingStatusCode`</a> | Defines the status code for the ping handler during a graceful shut down. See more information [here](#terminatingstatuscode) | 503 | No   |

#### `terminatingStatusCode`

During the period in which Traefik is gracefully shutting down, the ping handler
returns a `503` status code by default.  
If Traefik is behind, for example a load-balancer
doing health checks (such as the Kubernetes LivenessProbe), another code might
be expected as the signal for graceful termination.  
In that case, the terminatingStatusCode can be used to set the code returned by the ping
handler during termination.

```yaml tab="File (YAML)"
ping:
  terminatingStatusCode: 204
```

```toml tab="File (TOML)"
[ping]
  terminatingStatusCode = 204
```

```bash tab="CLI"
--ping.terminatingStatusCode=204
```

## Ready

The `/ready` readiness-check URL is enabled with the command-line `--ready` or config file option `[ready]`.

Unlike `/ping`, which returns `200 OK` as soon as the Traefik process is alive and the entrypoint is bound,
`/ready` returns `200 OK` only after **all enabled providers have completed their initial configuration load**
and routes have been applied. Before that, it returns `503 Service Unavailable`.

This is useful in Kubernetes environments where pods should not receive traffic until all routes
from providers (Ingress, CRD, Gateway API, etc.) are compiled and ready to serve.

The entryPoint where `/ready` is active can be customized with the `entryPoint` option,
whose default value is `traefik` (port `8080`).

| Path    | Method        | Description                                                                                         |
|---------|---------------|-----------------------------------------------------------------------------------------------------|
| <a id="opt-ready" href="#opt-ready" title="#opt-ready">`/ready`</a> | `GET`, `HEAD` | An endpoint to check for Traefik readiness. Returns `200` only when all providers have loaded their initial configuration. |

### Configuration Example

To enable the ready handler:

```yaml tab="File (YAML)"
ready: {}
```

```toml tab="File (TOML)"
[ready]
```

```bash tab="CLI"
--ready=true
```

### Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| <a id="opt-ready-entryPoint" href="#opt-ready-entryPoint" title="#opt-ready-entryPoint">`ready.entryPoint`</a> | Enables `/ready` on a dedicated EntryPoint. | traefik  | No   |
| <a id="opt-ready-manualRouting" href="#opt-ready-manualRouting" title="#opt-ready-manualRouting">`ready.manualRouting`</a> | Disables the default internal router in order to allow one to create a custom router for the `ready@internal` service when set to `true`. | false | No   |
| <a id="opt-ready-terminatingStatusCode" href="#opt-ready-terminatingStatusCode" title="#opt-ready-terminatingStatusCode">`ready.terminatingStatusCode`</a> | Defines the status code for the ready handler during a graceful shut down. | 503 | No   |

### Kubernetes Usage Example

Use `/ready` as a `readinessProbe` and `/ping` as a `livenessProbe`:

```yaml
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 0
  periodSeconds: 5
livenessProbe:
  httpGet:
    path: /ping
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```
