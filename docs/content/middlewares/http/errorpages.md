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

```yaml tab="Docker"
# Dynamic Custom Error Page for 5XX Status Code
labels:
  - "traefik.http.middlewares.test-errors.errors.status=500-599"
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
      - "500-599"
    query: /{status}.html
    service:
      name: whoami
      port: 80
```

```yaml tab="Consul Catalog"
# Dynamic Custom Error Page for 5XX Status Code
- "traefik.http.middlewares.test-errors.errors.status=500-599"
- "traefik.http.middlewares.test-errors.errors.service=serviceError"
- "traefik.http.middlewares.test-errors.errors.query=/{status}.html"
```

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.test-errors.errors.status": "500-599",
  "traefik.http.middlewares.test-errors.errors.service": "serviceError",
  "traefik.http.middlewares.test-errors.errors.query": "/{status}.html"
}
```

```yaml tab="Rancher"
# Dynamic Custom Error Page for 5XX Status Code
labels:
  - "traefik.http.middlewares.test-errors.errors.status=500-599"
  - "traefik.http.middlewares.test-errors.errors.service=serviceError"
  - "traefik.http.middlewares.test-errors.errors.query=/{status}.html"
```

```yaml tab="File (YAML)"
# Custom Error Page for 5XX
http:
  middlewares:
    test-errors:
      errors:
        status:
          - "500-599"
        service: serviceError
        query: "/{status}.html"

  services:
    # ... definition of error-handler-service and my-service
```

```toml tab="File (TOML)"
# Custom Error Page for 5XX
[http.middlewares]
  [http.middlewares.test-errors.errors]
    status = ["500-599"]
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

The status code ranges are inclusive (`500-599` will trigger with every code between `500` and `599`, `500` and `599` included).

!!! note ""

    You can define either a status code as a number (`500`),
    as multiple comma-separated numbers (`500,502`),
    as ranges by separating two codes with a dash (`500-599`),
    or a combination of the two (`404,418,500-599`).

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

| Variable   | Value                                                              |
|------------|--------------------------------------------------------------------|
| `{status}` | The response status code.                                          |
| `{url}`    | The [escaped](https://pkg.go.dev/net/url#QueryEscape) request URL. |
