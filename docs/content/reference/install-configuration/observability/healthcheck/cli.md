---
title: "Baqup Health Check CLI Command Documentation"
description: "In Baqup Proxy, the healthcheck CLI command lets you check the health of your Baqup instances. Read the technical documentation for configuration examples and options."
---

# Healthcheck Command

Checking the Health of your Baqup Instances.
{: .subtitle }

## Usage

The healthcheck command allows you to make a request to the `/ping` endpoint (defined in the install (static) configuration) to check the health of Baqup. Its exit status is `0` if Baqup is healthy and `1` otherwise.

This can be used with [HEALTHCHECK](https://docs.docker.com/engine/reference/builder/#healthcheck) instruction or any other health check orchestration mechanism.

```sh
baqup healthcheck [command] [flags] [arguments]
```

Example:

```sh
$ baqup healthcheck
OK: http://:8082/ping
```

The command uses the [ping](./ping.md) endpoint that is defined in the Baqup install (static) configuration.
