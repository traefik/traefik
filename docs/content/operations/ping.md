# Ping

Checking the Health of Your Traefik Instances
{: .subtitle }

## Configuration Examples

??? example "Enabling /ping"

```toml tab="File (TOML)"
[ping]
```

```yaml tab="File (YAML)"
ping: {}
```

```bash tab="CLI"
--ping=true
```

| Path    | Method        | Description                                                                                         |
|---------|---------------|-----------------------------------------------------------------------------------------------------|
| `/ping` | `GET`, `HEAD` | A simple endpoint to check for Traefik process liveness. Return a code `200` with the content: `OK` |

## Configuration Options

The `/ping` health-check URL is enabled with the command-line `--ping` or config file option `[ping]`.