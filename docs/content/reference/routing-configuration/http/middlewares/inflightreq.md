---
title: "Traefik InFlightReq Documentation"
description: "Traefik Proxy's HTTP middleware lets you limit the number of simultaneous in-flight requests. Read the technical documentation."
---

The `inFlightReq` middleware proactively prevents services from being overwhelmed with high load.

## Configuration Examples

```yaml tab="Structured (YAML)"
# Limiting to 10 simultaneous connections
http:
  middlewares:
    test-inflightreq:
      inFlightReq:
        amount: 10
```

```toml tab="Structured (TOML)"
# Limiting to 10 simultaneous connections
[http.middlewares]
  [http.middlewares.test-inflightreq.inFlightReq]
    amount = 10
```

```yaml tab="Labels"
labels:
  - "traefik.http.middlewares.test-inflightreq.inflightreq.amount=10"
```

```json tab="Consul Catalog"
// Limiting to 10 simultaneous connections
{
  "Tags" : [
    "traefik.http.middlewares.test-inflightreq.inflightreq.amount=10"
  ]
}

```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-inflightreq
spec:
  inFlightReq:
    amount: 10
```

## Configuration Options

<!-- markdownlint-disable MD013 -->

| Field      | Description                                                                                                                                                                                 | Default | Required |
|:-----------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| `amount` | The `amount` option defines the maximum amount of allowed simultaneous in-flight request. <br /> The middleware responds with `HTTP 429 Too Many Requests` if there are already `amount` requests in progress (based on the same `sourceCriterion` strategy). | 0      | No      |
| `sourceCriterion.requestHost` | Whether to consider the request host as the source.<br /> More information about `sourceCriterion`[here](#sourcecriterion). | false      | No      |
| `sourceCriterion.requestHeaderName` | Name of the header used to group incoming requests.<br /> More information about `sourceCriterion`[here](#sourcecriterion). | ""      | No      |
| `sourceCriterion.ipStrategy.depth` | Depth position of the IP to select in the `X-Forwarded-For` header (starting from the right).<br />0 means no depth.<br />If greater than the total number of IPs in `X-Forwarded-For`, then the client IP is empty<br />If higher than 0, the `excludedIPs` options is not evaluated.<br /> More information about [`sourceCriterion`](#sourcecriterion), [`ipStrategy](#ipstrategy), and [`depth`](#example-of-depth--x-forwarded-for) below. | 0      | No      |
| `sourceCriterion.ipStrategy.excludedIPs` | Allows Traefik to scan the `X-Forwarded-For` header and select the first IP not in the list.<br />If `depth` is specified, `excludedIPs` is ignored.<br /> More information about [`sourceCriterion`](#sourcecriterion), [`ipStrategy](#ipstrategy), and [`excludedIPs`](#example-of-excludedips--x-forwarded-for) below. | | No      |
| `sourceCriterion.ipStrategy.ipv6Subnet` |  If `ipv6Subnet` is provided and the selected IP is IPv6, the IP is transformed into the first IP of the subnet it belongs to. <br /> More information about [`sourceCriterion`](#sourcecriterion), [`ipStrategy.ipv6Subnet`](#ipstrategyipv6subnet), and [`excludedIPs`](#example-of-excludedips--x-forwarded-for) below. |  | No      |

### sourceCriterion

The `sourceCriterion` option defines what criterion is used to group requests as originating from a common source.
If several strategies are defined at the same time, an error will be raised.
If none are set, the default is to use the `requestHost`.

### ipStrategy

The `ipStrategy` option defines three parameters that configures how Traefik determines the client IP: `depth`, `excludedIPs` and `ipv6Subnet`.

As a middleware, `inFlightReq` happens before the actual proxying to the backend takes place.
In addition, the previous network hop only gets appended to `X-Forwarded-For` during the last stages of proxying, that is after it has already passed through the middleware.
Therefore, during InFlightReq, as the previous network hop is not yet present in `X-Forwarded-For`, it cannot be used and/or relied upon.

### `ipStrategy.ipv6Subnet`

This strategy applies to `Depth` and `RemoteAddr` strategy only.
If `ipv6Subnet` is provided and the selected IP is IPv6, the IP is transformed into the first IP of the subnet it belongs to.

This is useful for grouping IPv6 addresses into subnets to prevent bypassing this middleware by obtaining a new IPv6.

- `ipv6Subnet` is ignored if its value is outside 0-128 interval

#### Example of ipv6Subnet

If `ipv6Subnet` is provided, the IP is transformed in the following way.

| IP                     | ipv6Subnet | clientIP              |
|---------------------------|--------------|-----------------------|
| `"::abcd:1111:2222:3333"` | `64`         | `"::0:0:0:0"`         |
| `"::abcd:1111:2222:3333"` | `80`         | `"::abcd:0:0:0:0"`    |
| `"::abcd:1111:2222:3333"` | `96`         | `"::abcd:1111:0:0:0"` |

### Example of Depth & X-Forwarded-For

If `depth` is set to 2, and the request `X-Forwarded-For` header is `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` then the "real" client IP is `"10.0.0.1"` (at depth 4) but the IP used as the criterion is `"12.0.0.1"` (`depth=2`).

| X-Forwarded-For                      | depth | clientIP     |
|-----------------------------------------|---------|--------------|
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `1`     | `"13.0.0.1"` |
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `3`     | `"11.0.0.1"` |
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `5`     | `""`         |

### Example of ExcludedIPs & X-Forwarded-For

| X-Forwarded-For                       | excludedIPs        | clientIP     |
|-----------------------------------------|-----------------------|--------------|
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `"12.0.0.1,13.0.0.1"` | `"11.0.0.1"` |
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `"15.0.0.1,13.0.0.1"` | `"12.0.0.1"` |
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `"10.0.0.1,13.0.0.1"` | `"12.0.0.1"` |
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `"15.0.0.1,16.0.0.1"` | `"13.0.0.1"` |
| `"10.0.0.1,11.0.0.1"`                   | `"10.0.0.1,11.0.0.1"` | `""`         |
