# CLI

The Traefik Command Line
{: .subtitle }

## General

```bash
traefik [command] [--flag=flag_argument]
```

Available commands: 

- `version` : Print version
- `storeconfig` : Store the static Traefik configuration into a Key-value stores. Please refer to the [Store Traefik configuration](/user-guide/kv-config/#store-configuration-in-key-value-store) section to get documentation on it.
- `bug`: The easiest way to submit a pre-filled issue.
- `healthcheck`: Calls Traefik `/ping` to check health.

Each command can have additional flags.

All those flags will be displayed with:

```bash
traefik [command] --help
```

Each command is described at the beginning of the help section:

```bash
traefik --help

# or

docker run traefik[:version] --help
# ex: docker run traefik:1.5 --help
```

## Command: bug

The easiest way to submit a pre-filled issue on [Traefik GitHub](https://github.com/containous/traefik)! Watch [this demo](https://www.youtube.com/watch?v=Lyz62L8m93I) for more information.

```bash
traefik bug
```

### Command: healthcheck

Checks the health of Traefik.
Its exit status is `0` if Traefik is healthy and `1` if it is unhealthy.

This can be used with Docker [HEALTHCHECK](https://docs.docker.com/engine/reference/builder/#healthcheck) instruction or any other health check orchestration mechanism.

!!! note
    The [`ping`](/features/ping/) endpoint must be enabled to allow the `healthcheck` command to call `/ping`.

```bash
traefik healthcheck
```

```bash
OK: http://:8082/ping
```
