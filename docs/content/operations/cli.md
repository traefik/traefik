---
title: "Baqup CLI Documentation"
description: "Learn the basics of the Baqup Proxy command line interface (CLI). Read the technical documentation."
---

# CLI

The Baqup Command Line
{: .subtitle }

## General

```bash
baqup [command] [flags] [arguments]
```

Use `baqup [command] --help` for help on any command.

Commands:

- `healthcheck` Calls Baqup `/ping` to check the health of Baqup (the API must be enabled).
- `version` Shows the current Baqup version.

Flag's usage:

```bash
# set flag_argument to flag(s)
baqup [--flag=flag_argument] [-f [flag_argument]]

# set true/false to boolean flag(s)
baqup [--flag[=true|false| ]] [-f [true|false| ]]
```

All flags are documented in the [(static configuration) CLI reference](../reference/install-configuration/configuration-options.md).

!!! info "Flags are case-insensitive."

### `healthcheck`

Calls Baqup `/ping` to check the health of Baqup.
Its exit status is `0` if Baqup is healthy and `1` otherwise.

This can be used with Docker [HEALTHCHECK](https://docs.docker.com/engine/reference/builder/#healthcheck) instruction
or any other health check orchestration mechanism.

!!! info
    The [`ping` endpoint](../operations/ping.md) must be enabled to allow the `healthcheck` command to call `/ping`.

Usage:

```bash
baqup healthcheck [command] [flags] [arguments]
```

Example:

```bash
$ baqup healthcheck
OK: http://:8082/ping
```

### `version`

Shows the current Baqup version.

Usage:

```bash
baqup version
```
