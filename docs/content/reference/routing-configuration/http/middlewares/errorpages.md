---
title: "Traefik Errors Documentation"
description: "In Traefik Proxy, the Errors middleware returns custom pages according to configured ranges of HTTP Status codes. Read the technical documentation."
---

The `errors` middleware returns a custom page in lieu of the default, according to configured ranges of HTTP Status codes.

## Configuration Examples

```yaml tab="Structured (YAML)"
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
        statusRewrites:
          "418": 404
          "502-504": 500
        service: error-handler-service
        query: "/{status}.html"

  services:
    # ... definition of the error-handler-service
```

```toml tab="Structured (TOML)"
# Dynamic Custom Error Page for 5XX Status Code excluding 502 and 504
[http.middlewares]
  [http.middlewares.test-errors.errors]
    status = ["500","501","503","505-599"]
    service = "error-handler-service"
    query = "/{status}.html"

    [http.middlewares.test-errors.errors.statusRewrites]
      "418" = 404
      "502-504" = 500

[http.services]
  # ... definition of the error-handler-service
```

```yaml tab="Labels"
# Dynamic Custom Error Page for 5XX Status Code
labels:
  - "traefik.http.middlewares.test-errors.errors.status=500,501,503,505-599"
  - "traefik.http.middlewares.test-errors.errors.statusRewrites.418=404"
  - "traefik.http.middlewares.test-errors.errors.statusRewrites.502-504=500"
  - "traefik.http.middlewares.test-errors.errors.service=error-handler-service"
  - "traefik.http.middlewares.test-errors.errors.query=/{status}.html"
```

```json tab="Tags"
// Dynamic Custom Error Page for 5XX Status Code excluding 502 and 504
{
  // ...
  "Tags": [
    "traefik.http.middlewares.test-errors.errors.status=500,501,503,505-599",
    "traefik.http.middlewares.test-errors.errors.statusRewrites.418=404",
    "traefik.http.middlewares.test-errors.errors.statusRewrites.502-504=500",
    "traefik.http.middlewares.test-errors.errors.service=error-handler-service",
    "traefik.http.middlewares.test-errors.errors.query=/{status}.html"
  ]

}

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
    statusRewrites:
      "418": 404
      "502-504": 500
    query: /{status}.html
    service:
      name: error-handler-service
      port: 80
```

## Configuration Options

| Field      | Description                                                                                                                                                                                 | Default | Required |
|:-----------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| <a id="opt-status" href="#opt-status" title="#opt-status">`status`</a> | Defines which status or range of statuses should result in an error page.<br/> The status code ranges are inclusive (`505-599` will trigger with every code between `505` and `599`, `505` and `599` included).<br /> You can define either a status code as a number (`500`), as multiple comma-separated numbers (`500,502`), as ranges by separating two codes with a dash (`505-599`), or a combination of the two (`404,418,505-599`).  | []     | No      | 
| <a id="opt-statusRewrites" href="#opt-statusRewrites" title="#opt-statusRewrites">`statusRewrites`</a> | An optional mapping of status codes to be rewritten. More information [here](#statusrewrites).  | []     | No      |
| <a id="opt-service" href="#opt-service" title="#opt-service">`service`</a> | The service that will serve the new requested error page.<br /> More information [here](#service-and-hostheader). | ""      | Yes      |
| <a id="opt-query" href="#opt-query" title="#opt-query">`query`</a> | The URL for the error page (hosted by `service`).<br /> More information [here](#query) | ""      | No      |
| <a id="opt-errorRequestHeaders" href="#opt-errorRequestHeaders" title="#opt-errorRequestHeaders">`errorRequestHeaders`</a> | Defines the list of original request headers forwarded to the error page service.<br /> More information [here](#errorrequestheaders) | []      | No      |
| <a id="opt-errorResponseHeaders" href="#opt-errorResponseHeaders" title="#opt-errorResponseHeaders">`errorResponseHeaders`</a> | Defines the list of original backend response headers forwarded to the error page service request.<br /> More information [here](#errorresponseheaders) | [] | No |

### service and HostHeader

By default, the client `Host` header value is forwarded to the configured error service.
To forward the `Host` value corresponding to the configured error service URL, 
the [`passHostHeader`](../load-balancing/service.md#opt-passHostHeader) option must be set to `false`.

!!!info "Kubernetes"
    When specifying a service in Kubernetes (e.g., in an IngressRoute), you need to reference the `name`, `namespace`, and `port` of your Kubernetes Service resource. For example, `my-service.my-namespace@kubernetescrd` (or `my-service.my-namespace@kubernetescrd:80`) ensures that requests go to the correct service and port.

!!! info "ServersTransport (Kubernetes)"

    To customize how Traefik connects to the error page service (for example, to configure TLS to the backend), set a [`serversTransport`](../../kubernetes/crd/http/serverstransport.md) on the middleware's `service`.
    The `traefik.ingress.kubernetes.io/service.serverstransport` annotation on the Kubernetes Service is not applied here: it only affects a Service used as an Ingress backend, not one referenced by a middleware.

    ```yaml
    apiVersion: traefik.io/v1alpha1
    kind: Middleware
    metadata:
      name: test-errors
    spec:
      errors:
        status:
          - "500"
          - "501"
        service:
          name: error-handler-service
          port: 80
          serversTransport: mytransport
    ```

### statusRewrites

`statusRewrites` is an optional mapping of status codes to be rewritten.

For example, if a service returns a 418, you might want to rewrite it to a 404.
You can map individual status codes or even ranges to a different status code.

The syntax for ranges follows the same rules as the <a href="#opt-status">`status`</a> option.

### query

There are multiple variables that can be placed in the `query` option to insert values in the URL.

The table below lists all the available variables and their associated values.

| Variable   | Value                                                            |
|------------|------------------------------------------------------------------|
| <a id="opt-status-2" href="#opt-status-2" title="#opt-status-2">`{status}`</a> | The response status code.                                        |
| <a id="opt-originalStatus" href="#opt-originalStatus" title="#opt-originalStatus">`{originalStatus}`</a> | The original response status code, if it has been modified by the `statusRewrites` option. |
| <a id="opt-url" href="#opt-url" title="#opt-url">`{url}`</a> | The [escaped](https://pkg.go.dev/net/url#QueryEscape) request URL.|

### `errorRequestHeaders`

Defines the list of original request headers forwarded to the error page service.

By default (`errorRequestHeaders` not set), all request headers — including authentication material such as `Authorization` and `Cookie` — are forwarded.
If the error page service is in a separate trust domain, use this option to restrict which headers cross the service boundary.

Set to an explicit list to forward only those headers, or set to an empty list (`errorRequestHeaders: []`) to forward no headers.

### `errorResponseHeaders`

Defines the list of headers copied from the original backend error response to the request sent to the error page service.
By default, or with an empty list, no backend response headers are copied to this request.

The selected response headers replace any matching original request headers forwarded by `errorRequestHeaders`.
If a selected header is absent from the backend response, the matching request header is removed.
All values of a selected header are copied. Header names are case-insensitive, and hop-by-hop headers are ignored.

This option works independently of `errorRequestHeaders`, including when it is an empty list.
It does not add headers to the client response: to send a header such as `WWW-Authenticate` back to the client, the error page service must return it (for example, an error page service that echoes selected request headers into its response).

For example, the following configuration lets the error page service read the original authentication challenge and decide how to render the error page:

```yaml tab="Structured (YAML)"
http:
  middlewares:
    test-errors:
      errors:
        status:
          - "401"
        service: serviceError
        query: /{status}.html
        errorRequestHeaders: []
        errorResponseHeaders:
          - WWW-Authenticate
```

Use `errorRequestHeaders` for client request headers and `errorResponseHeaders` for backend response headers sent to the error page service.
