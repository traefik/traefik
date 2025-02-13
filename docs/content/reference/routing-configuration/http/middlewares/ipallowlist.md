---
title: "Traefik HTTP Middlewares IPAllowList"
description: "Learn how to use IPAllowList in HTTP middleware for limiting clients to specific IPs in Traefik Proxy. Read the technical documentation."
---

`ipAllowList` accepts / refuses requests based on the client IP.

## Configuration Example

```yaml tab="Structured (YAML)"
# Accepts request from defined IP
http:
  middlewares:
    test-ipallowlist:
      ipAllowList:
        sourceRange:
          - "127.0.0.1/32"
          - "192.168.1.7"
```

```toml tab="Structured (TOML)"
# Accepts request from defined IP
[http.middlewares]
  [http.middlewares.test-ipallowlist.ipAllowList]
    sourceRange = ["127.0.0.1/32", "192.168.1.7"]
```

```yaml tab="Labels"
# Accepts request from defined IP
labels:
  - "traefik.http.middlewares.test-ipallowlist.ipallowlist.sourcerange=127.0.0.1/32, 192.168.1.7"
```

```json tab="Tags"
// Accepts request from defined IP
{
  "Tags" : [
    "traefik.http.middlewares.test-ipallowlist.ipallowlist.sourcerange=127.0.0.1/32, 192.168.1.7"
  ]
}
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ipallowlist
spec:
  ipAllowList:
    sourceRange:
      - 127.0.0.1/32
      - 192.168.1.7
```

## Configuration Options

| Field      | Description     | Default | Required |
|:-----------|:------------------------------|:--------|:---------|
| `sourceRange` | List of allowed IPs (or ranges of allowed IPs by using CIDR notation). |      | Yes      |
| `ipStrategy.depth` | Depth position of the IP to select in the `X-Forwarded-For` header (starting from the right).<br />0 means no depth.<br />If greater than the total number of IPs in `X-Forwarded-For`, then the client IP is empty<br /> If higher than 0, the `excludedIPs` options is not evaluated.<br /> More information about [`ipStrategy](#ipstrategy), and [`depth`](#example-of-depth--x-forwarded-for) below. | 0      | No      |
| `ipStrategy.excludedIPs` | Allows Traefik to scan the `X-Forwarded-For` header and select the first IP not in the list.<br />If `depth` is specified, `excludedIPs` is ignored.<br /> More information about [`ipStrategy](#ipstrategy), and [`excludedIPs`](#example-of-excludedips--x-forwarded-for) below. |       | No      |
| `ipStrategy.ipv6Subnet` |  If `ipv6Subnet` is provided and the selected IP is IPv6, the IP is transformed into the first IP of the subnet it belongs to. <br />More information about [`ipStrategy.ipv6Subnet`](#ipstrategyipv6subnet), and [`excludedIPs`](#example-of-excludedips--x-forwarded-for) below. |       | No      |

### ipStrategy

The `ipStrategy` option defines two parameters that configures how Traefik determines the client IP: `depth`, and `excludedIPs`.

If no strategy is set, the default behavior is to match `sourceRange` against the Remote address found in the request.

As a middleware, passlisting happens before the actual proxying to the backend takes place.
In addition, the previous network hop only gets appended to `X-Forwarded-For` during the last stages of proxying, that is after it has already passed through passlisting.
Therefore, during passlisting, as the previous network hop is not yet present in `X-Forwarded-For`, it cannot be matched against `sourceRange`.

#### `ipStrategy.depth`

The `depth` option tells Traefik to use the `X-Forwarded-For` header and take the IP located at the `depth` position (starting from the right).

- If `depth` is greater than the total number of IPs in `X-Forwarded-For`, then the client IP will be empty.
- `depth` is ignored if its value is less than or equal to 0.

If `ipStrategy.ipv6Subnet` is provided and the selected IP is IPv6, the IP is transformed into the first IP of the subnet it belongs to.  

### `ipStrategy.ipv6Subnet`

This strategy applies to `Depth` and `RemoteAddr` strategy only.
If `ipv6Subnet` is provided and the selected IP is IPv6, the IP is transformed into the first IP of the subnet it belongs to.

This is useful for grouping IPv6 addresses into subnets to prevent bypassing this middleware by obtaining a new IPv6.

- `ipv6Subnet` is ignored if its value is outside 0-128 interval

#### Example of ipv6Subnet

If `ipv6Subnet` is provided, the IP is transformed in the following way.

| IP                      | ipv6Subnet | clientIP              |
|---------------------------|--------------|-----------------------|
| `"::abcd:1111:2222:3333"` | `64`         | `"::0:0:0:0"`         |
| `"::abcd:1111:2222:3333"` | `80`         | `"::abcd:0:0:0:0"`    |
| `"::abcd:1111:2222:3333"` | `96`         | `"::abcd:1111:0:0:0"` |

### Example of Depth & X-Forwarded-For

If `depth` is set to 2, and the request `X-Forwarded-For` header is `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` then the "real" client IP is `"10.0.0.1"` (at depth 4) but the IP used as the criterion is `"12.0.0.1"` (`depth=2`).

| X-Forwarded-For                       | depth | clientIP     |
|-----------------------------------------|---------|--------------|
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `1`     | `"13.0.0.1"` |
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `3`     | `"11.0.0.1"` |
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `5`     | `""`         |

### Example of ExcludedIPs & X-Forwarded-For

| X-Forwarded-For                       | excludedIPs         | clientIP     |
|-----------------------------------------|-----------------------|--------------|
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `"12.0.0.1,13.0.0.1"` | `"11.0.0.1"` |
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `"15.0.0.1,13.0.0.1"` | `"12.0.0.1"` |
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `"10.0.0.1,13.0.0.1"` | `"12.0.0.1"` |
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `"15.0.0.1,16.0.0.1"` | `"13.0.0.1"` |
| `"10.0.0.1,11.0.0.1"`                   | `"10.0.0.1,11.0.0.1"` | `""`         |
