---
title: "Traefik Ping Option Documentation"
description: "In Traefik Proxy, the option Ping lets you check the health of your Traefik instances. Read the technical documentation for configuration examples and options."
---

# Ping 

Checking the Health of your Traefik Instances
{: .subtitle }

The `ping` options allows you to enable the ping endpoint to check Traefik liveness.

The ping endpoint is reachable using the path `/ping` and the methods `GET`and `HEAD`.

If the Traefik instance is alive, it returns the `200` HTTP code with the content: `OK`.

## Configuration Example

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

## Configuration Options

The `ping` option is defined in the install (static) configuration.
You can define it using the same [configuration methods](../../boot-environment.md#configuration-methods) as Traefik.

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| <a id="ping-entryPoint" href="#ping-entryPoint" title="#ping-entryPoint">`ping.entryPoint`</a> | Enables `/ping` on a dedicated EntryPoint. | traefik  | No   |
| <a id="ping-manualRouting" href="#ping-manualRouting" title="#ping-manualRouting">`ping.manualRouting`</a> | Disables the default internal router in order to allow one to create a custom router for the `ping@internal` service when set to `true`. | false | No   |
| <a id="ping-terminatingStatusCode" href="#ping-terminatingStatusCode" title="#ping-terminatingStatusCode">`ping.terminatingStatusCode`</a> | Defines the status code for the ping handler during a graceful shut down. See more information [here](#terminatingstatuscode) | 503 | No   |

### `terminatingStatusCode`

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
