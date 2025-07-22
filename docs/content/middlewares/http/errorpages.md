---
title: "Traefik Errors Documentation"
description: "In Traefik Proxy, the Errors middleware returns custom pages according to configured ranges of HTTP Status codes. Read the technical documentation."
---

# Errors

It Has Never Been Easier to Say That Something Went Wrong
{: .subtitle }

![Errors](../../assets/img/middleware/errorpages.png)

The Errors middleware returns a custom page in lieu of the default, according to configured ranges of HTTP Status codes.

!!! important

    The error page itself is _not_ hosted by Traefik.

## Configuration Examples

```yaml tab="Docker & Swarm"
# Dynamic Custom Error Page for 5XX Status Code
labels:
  - "traefik.http.middlewares.test-errors.errors.status=500,501,503,505-599"
  - "traefik.http.middlewares.test-errors.errors.service=serviceError"
  - "traefik.http.middlewares.test-errors.errors.query=/{status}.html"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-errors
spec:
  errors:
    status:
      - "500"
      - "501"
      - "503"
      - "505-599"
    query: /{status}.html
    service:
      name: whoami
      port: 80
```

```yaml tab="Consul Catalog"
# Dynamic Custom Error Page for 5XX Status Code excluding 502 and 504
- "traefik.http.middlewares.test-errors.errors.status=500,501,503,505-599"
- "traefik.http.middlewares.test-errors.errors.service=serviceError"
- "traefik.http.middlewares.test-errors.errors.query=/{status}.html"
```

```yaml tab="File (YAML)"
# Dynamic Custom Error Page for 5XX Status Code excluding 502 and 504
http:
  middlewares:
    test-errors:
      errors:
        status:
          - "500"
          - "501"
          - "503"
          - "505-599"
        service: serviceError
        query: "/{status}.html"

  services:
    # ... definition of error-handler-service and my-service
```

```toml tab="File (TOML)"
# Dynamic Custom Error Page for 5XX Status Code excluding 502 and 504
[http.middlewares]
  [http.middlewares.test-errors.errors]
    status = ["500","501","503","505-599"]
    service = "serviceError"
    query = "/{status}.html"

[http.services]
  # ... definition of error-handler-service and my-service
```

!!! note ""

    In this example, the error page URL is based on the status code (`query=/{status}.html`).

## Configuration Options

### `status`

The `status` option defines which status or range of statuses should result in an error page.

The status code ranges are inclusive (`505-599` will trigger with every code between `505` and `599`, `505` and `599` included).

!!! note ""

    You can define either a status code as a number (`500`),
    as multiple comma-separated numbers (`500,502`),
    as ranges by separating two codes with a dash (`505-599`),
    or a combination of the two (`404,418,505-599`).
    The comma-separated syntax is only available for label-based providers.
    The examples above demonstrate which syntax is appropriate for each provider.

### `statusRewrites`

An optional mapping of status codes to be rewritten. For example, if a service returns a 418, you might want to rewrite it to a 404.
You can map individual status codes or even ranges to a different status code. The syntax for ranges follows the same rules as the `status` option.

Here is an example:

```yml
statusRewrites:
  "500-503": 500
  "418": 404
```

### `service`

The service that will serve the new requested error page.

!!! note ""

    In Kubernetes, you need to reference a Kubernetes Service instead of a Traefik service.

!!! info "Host Header"

    By default, the client `Host` header value is forwarded to the configured error [service](#service).
    To forward the `Host` value corresponding to the configured error service URL, the [passHostHeader](../../../routing/services/#pass-host-header) option must be set to `false`.

### `query`

The URL for the error page (hosted by [`service`](#service))).

There are multiple variables that can be placed in the `query` option to insert values in the URL.

The table below lists all the available variables and their associated values.

| Variable           | Value                                                                                      |
|--------------------|--------------------------------------------------------------------------------------------|
| `{status}`         | The response status code. It may be rewritten when using the `statusRewrites` option.      |
| `{originalStatus}` | The original response status code, if it has been modified by the `statusRewrites` option. |
| `{url}`            | The [escaped](https://pkg.go.dev/net/url#QueryEscape) request URL.                         |
