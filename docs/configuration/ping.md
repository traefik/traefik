# Ping Definition

## Configuration

```toml
# Ping definition
[ping]
  # Name of the related entry point
  #
  # Optional
  # Default: "traefik"
  #
  entryPoint = "traefik"
```

| Path    | Method        | Description                                                                                        |
|---------|---------------|----------------------------------------------------------------------------------------------------|
| `/ping` | `GET`, `HEAD` | A simple endpoint to check for Traefik process liveness. Return a code `200` with the content: `OK` |


!!! warning
    Even if you have authentication configured on entry point, the `/ping` path of the api is excluded from authentication.

## Examples

The `/ping` health-check URL is enabled with the command-line `--ping` or config file option `[ping]`.
Thus, if you have a regular path for `/foo` and an entrypoint on `:80`, you would access them as follows:

* Regular path: `http://hostname:80/foo`
* Admin panel: `http://hostname:8080/`
* Ping URL: `http://hostname:8080/ping`

However, for security reasons, you may want to be able to expose the `/ping` health-check URL to outside health-checkers, e.g. an Internet service or cloud load-balancer, _without_ exposing your dashboard's port.
In many environments, the security staff may not _allow_ you to expose it.

You have two options:

* Enable `/ping` on a regular entry point
* Enable `/ping` on a dedicated port

### Ping health check on a regular entry point

To proxy `/ping` from a regular entry point to the administration one without exposing the dashboard, do the following:

```toml
defaultEntryPoints = ["http"]

[entryPoints]
  [entryPoints.http]
  address = ":80"

[ping]
entryPoint = "http"

```

The above link `ping` on the `http` entry point and then expose it on port `80`

### Enable ping health check on dedicated port

If you do not want to or cannot expose the health-check on a regular entry point - e.g. your security rules do not allow it, or you have a conflicting path - then you can enable health-check on its own entry point.
Use the following configuration:

```toml
defaultEntryPoints = ["http"]

[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.ping]
  address = ":8082"

[ping]
entryPoint = "ping"
```

The above is similar to the previous example, but instead of enabling `/ping` on the _default_ entry point, we enable it on a _dedicated_ entry point.

In the above example, you would access a regular path and health-check as follows:

* Regular path: `http://hostname:80/foo`
* Ping URL: `http://hostname:8082/ping`

Note the dedicated port `:8082` for `/ping`.

In the above example, it is _very_ important to create a named dedicated entry point, and do **not** include it in `defaultEntryPoints`.
Otherwise, you are likely to expose _all_ services via this entry point.

### Using ping for external Load-balancer rotation health check

If you are running traefik behind a external Load-balancer, and want to configure rotation health check on the Load-balancer to take a traefik instance out of rotation gracefully, you can configure [lifecycle.requestAcceptGraceTimeout](/configuration/commons.md#life-cycle) and the ping endpoint will return `503` response on traefik server termination, so that the Load-balancer can take the terminating traefik instance out of rotation, before it stops responding.
