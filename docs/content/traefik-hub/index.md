# Traefik Hub (Experimental)

Once the Traefik Hub Experimental feature is enabled in Traefik,
Traefik and its local agent communicate together.
This agent can:

* get the Traefik metrics to display them in the Traefik Hub UI 
* secure the Traefik routers 
* provide ACME certificates to Traefik 
* transfer requests from the SaaS Platform to Traefik (and then avoid the users to expose directly their infrastructure on the internet)

!!! important "Learn More About Traefik Hub"

    This section is intended only as a brief overview for Traefik users who are not familiar with Traefik Hub.
    To explore all that Traefik Hub has to offer, please consult the [Traefik Hub Documentation](https://doc.traefik.io/traefik-hub).

!!! Note "Prerequisites"

    * Traefik Hub is compatible with Traefik Proxy 2.7 or later.
    * The Traefik Hub Agent must be installed to connect to the Traefik Hub platform.
    * Activate this feature in the experimental section of the static configuration.

!!! example "Minimal Static Configuration to Activate Traefik Hub"

    ```yaml tab="File (YAML)"
    experimental:
      hub: true

    hub: {}
    ```

    ```toml tab="File (TOML)"
    [experimental]
        hub = true

    [hub]
    ```

    ```bash tab="CLI"
    --experimental.hub
    --hub=true
    ```

## Configuration

### `entrypoint`

_Optional, Default="traefik-hub"_

Defines the entrypoint that exposes data for Traefik Hub Agent.

!!! info

    * If no entryPoint are configured, a `traefik-hub` entrypoint is created.  
    * If the entryPoint named `traefik-hub` is not configured, it is automatically created on port `9900`.
    * In any other cases, the option value must match an existing entrypoint name.

```yaml tab="File (YAML)"
entrypoints:
    hub-ep: ":9000"

hub:
  entrypoint: "hub-ep"
```

```toml tab="File (TOML)"
[entrypoints.hub-ep]
    address = ":9000"

[hub]
  entrypoint = "hub-ep"
```

```bash tab="CLI"
--entrypoints.hub-ep.address=:9000
--hub.entrypoint=hub-ep
```
