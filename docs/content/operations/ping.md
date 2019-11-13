# Ping

Checking the Health of Your Traefik Instances
{: .subtitle }

## Configuration Examples

To enable the API handler:

```toml tab="File (TOML)"
[ping]
```

```yaml tab="File (YAML)"
ping: {}
```

```bash tab="CLI"
--ping=true
```

## Configuration Options

The `/ping` health-check URL is enabled with the command-line `--ping` or config file option `[ping]`.

You can customize the `entryPoint` where the `/ping` is active with the `entryPoint` option (default value: `traefik`)

| Path    | Method        | Description                                                                                         |
|---------|---------------|-----------------------------------------------------------------------------------------------------|
| `/ping` | `GET`, `HEAD` | A simple endpoint to check for Traefik process liveness. Return a code `200` with the content: `OK` |

!!! note
    The `cli` comes with a [`healthcheck`](./cli.md#healthcheck) command which can be used for calling this endpoint.

### `entryPoint`

Enabling /ping on a dedicated EntryPoint.

```toml tab="File (TOML)"
[entryPoints]
  [entryPoints.ping]
    address = ":8082"

[ping]
  entryPoint = "ping"
```

```yaml tab="File (YAML)"
entryPoints:
  ping:
    address: ":8082"

ping:
  entryPoint: "ping"
```

```bash tab="CLI"
--entryPoints.ping.address=:8082
--ping.entryPoint=ping
```
