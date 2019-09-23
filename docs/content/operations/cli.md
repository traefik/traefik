# CLI

The Traefik Command Line
{: .subtitle }

## General

```bash
traefik [command] [flags] [arguments]
```

Use `traefik [command] --help` for help on any command.

Commands:

- `healthcheck` Calls Traefik `/ping` to check the health of Traefik (the API must be enabled).
- `version` Shows the current Traefik version.

Flag's usage:

```bash
# set flag_argument to flag(s)
traefik [--flag=flag_argument] [-f [flag_argument]]

# set true/false to boolean flag(s)
traefik [--flag[=true|false| ]] [-f [true|false| ]]
```

All flags are documented in the [(static configuration) CLI reference](../reference/static-configuration/cli.md).

!!! info "Flags are case insensitive."

### `healthcheck`

Calls Traefik `/ping` to check the health of Traefik.
Its exit status is `0` if Traefik is healthy and `1` otherwise.

This can be used with Docker [HEALTHCHECK](https://docs.docker.com/engine/reference/builder/#healthcheck) instruction
or any other health check orchestration mechanism.

!!! info
    The [`ping` endpoint](../operations/ping.md) must be enabled to allow the `healthcheck` command to call `/ping`.

Usage:

```bash
traefik healthcheck [command] [flags] [arguments]
```

Example:

```bash
$ traefik healthcheck
OK: http://:8082/ping
```

### `version`

Shows the current Traefik version.

Usage:

```bash
traefik version
```
