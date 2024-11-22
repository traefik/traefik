---
title: "Traefik Errors Documentation"
description: "In Traefik Proxy, the Errors middleware returns custom pages according to configured ranges of HTTP Status codes. Read the technical documentation."
---

![Errors](../../../../assets/img/middleware/errorpages.png)

The `errors` middleware returns a custom page in lieu of the default, according to configured ranges of HTTP Status codes.

## Configuration Examples

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

```yaml tab="Docker & Swarm"
# Dynamic Custom Error Page for 5XX Status Code
labels:
  - "traefik.http.middlewares.test-errors.errors.status=500,501,503,505-599"
  - "traefik.http.middlewares.test-errors.errors.service=serviceError"
  - "traefik.http.middlewares.test-errors.errors.query=/{status}.html"
```

```yaml tab="Consul Catalog"
# Dynamic Custom Error Page for 5XX Status Code excluding 502 and 504
- "traefik.http.middlewares.test-errors.errors.status=500,501,503,505-599"
- "traefik.http.middlewares.test-errors.errors.service=serviceError"
- "traefik.http.middlewares.test-errors.errors.query=/{status}.html"
```

## Configuration Options

| Field      | Description                                                                                                                                                                                 | Default | Required |
|:-----------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| `status` | Defines which status or range of statuses should result in an error page.< br/> The status code ranges are inclusive (`505-599` will trigger with every code between `505` and `599`, `505` and `599` included).<br /> You can define either a status code as a number (`500`), as multiple comma-separated numbers (`500,502`), as ranges by separating two codes with a dash (`505-599`), or a combination of the two (`404,418,505-599`).  | ""      | No      | 
| `service` | The Kubernetes Service that will serve the new requested error page.<br /> More information [here](#service-and-hostheader). | ""      | No      |
| `query` | The URL for the error page (hosted by `service`).<br /> More information [here](#query) | ""      | No      |

### service and HostHeader

By default, the client `Host` header value is forwarded to the configured error service.
To forward the `Host` value corresponding to the configured error service URL, 
the `passHostHeader` option must be set to `false`.

### query

There are multiple variables that can be placed in the `query` option to insert values in the URL.

The table below lists all the available variables and their associated values.

| Variable   | Value                                                            |
|------------|------------------------------------------------------------------|
| `{status}` | The response status code.                                        |
| `{url}`    | The [escaped](https://pkg.go.dev/net/url#QueryEscape) request URL.|
