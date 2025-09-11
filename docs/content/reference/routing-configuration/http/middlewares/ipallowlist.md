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
| <a id="sourceRange" href="#sourceRange" title="#sourceRange">`sourceRange`</a> | List of allowed IPs (or ranges of allowed IPs by using CIDR notation). |      | Yes      |
| <a id="ipStrategy-depth" href="#ipStrategy-depth" title="#ipStrategy-depth">`ipStrategy.depth`</a> | Depth position of the IP to select in the `X-Forwarded-For` header (starting from the right).<br />0 means no depth.<br />If greater than the total number of IPs in `X-Forwarded-For`, then the client IP is empty<br /> If higher than 0, the `excludedIPs` options is not evaluated.<br /> More information about [`ipStrategy](#ipstrategy), and [`depth`](#example-of-depth--x-forwarded-for) below. | 0      | No      |
| <a id="ipStrategy-excludedIPs" href="#ipStrategy-excludedIPs" title="#ipStrategy-excludedIPs">`ipStrategy.excludedIPs`</a> | Allows Traefik to scan the `X-Forwarded-For` header and select the first IP not in the list.<br />If `depth` is specified, `excludedIPs` is ignored.<br /> More information about [`ipStrategy](#ipstrategy), and [`excludedIPs`](#example-of-excludedips--x-forwarded-for) below. |       | No      |
| <a id="ipStrategy-ipv6Subnet" href="#ipStrategy-ipv6Subnet" title="#ipStrategy-ipv6Subnet">`ipStrategy.ipv6Subnet`</a> |  If `ipv6Subnet` is provided and the selected IP is IPv6, the IP is transformed into the first IP of the subnet it belongs to. <br />More information about [`ipStrategy.ipv6Subnet`](#ipstrategyipv6subnet), and [`excludedIPs`](#example-of-excludedips--x-forwarded-for) below. |       | No      |

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
| <a id="abcd111122223333" href="#abcd111122223333" title="#abcd111122223333">`"::abcd:1111:2222:3333"`</a> | `64`         | `"::0:0:0:0"`         |
| <a id="abcd111122223333-2" href="#abcd111122223333-2" title="#abcd111122223333-2">`"::abcd:1111:2222:3333"`</a> | `80`         | `"::abcd:0:0:0:0"`    |
| <a id="abcd111122223333-3" href="#abcd111122223333-3" title="#abcd111122223333-3">`"::abcd:1111:2222:3333"`</a> | `96`         | `"::abcd:1111:0:0:0"` |

### Example of Depth & X-Forwarded-For

If `depth` is set to 2, and the request `X-Forwarded-For` header is `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` then the "real" client IP is `"10.0.0.1"` (at depth 4) but the IP used as the criterion is `"12.0.0.1"` (`depth=2`).

| X-Forwarded-For                       | depth | clientIP     |
|-----------------------------------------|---------|--------------|
| <a id="10-0-0-111-0-0-112-0-0-113-0-0-1" href="#10-0-0-111-0-0-112-0-0-113-0-0-1" title="#10-0-0-111-0-0-112-0-0-113-0-0-1">`"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"`</a> | `1`     | `"13.0.0.1"` |
| <a id="10-0-0-111-0-0-112-0-0-113-0-0-1-2" href="#10-0-0-111-0-0-112-0-0-113-0-0-1-2" title="#10-0-0-111-0-0-112-0-0-113-0-0-1-2">`"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"`</a> | `3`     | `"11.0.0.1"` |
| <a id="10-0-0-111-0-0-112-0-0-113-0-0-1-3" href="#10-0-0-111-0-0-112-0-0-113-0-0-1-3" title="#10-0-0-111-0-0-112-0-0-113-0-0-1-3">`"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"`</a> | `5`     | `""`         |

### Example of ExcludedIPs & X-Forwarded-For

| X-Forwarded-For                       | excludedIPs         | clientIP     |
|-----------------------------------------|-----------------------|--------------|
| <a id="10-0-0-111-0-0-112-0-0-113-0-0-1-4" href="#10-0-0-111-0-0-112-0-0-113-0-0-1-4" title="#10-0-0-111-0-0-112-0-0-113-0-0-1-4">`"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"`</a> | `"12.0.0.1,13.0.0.1"` | `"11.0.0.1"` |
| <a id="10-0-0-111-0-0-112-0-0-113-0-0-1-5" href="#10-0-0-111-0-0-112-0-0-113-0-0-1-5" title="#10-0-0-111-0-0-112-0-0-113-0-0-1-5">`"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"`</a> | `"15.0.0.1,13.0.0.1"` | `"12.0.0.1"` |
| <a id="10-0-0-111-0-0-112-0-0-113-0-0-1-6" href="#10-0-0-111-0-0-112-0-0-113-0-0-1-6" title="#10-0-0-111-0-0-112-0-0-113-0-0-1-6">`"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"`</a> | `"10.0.0.1,13.0.0.1"` | `"12.0.0.1"` |
| <a id="10-0-0-111-0-0-112-0-0-113-0-0-1-7" href="#10-0-0-111-0-0-112-0-0-113-0-0-1-7" title="#10-0-0-111-0-0-112-0-0-113-0-0-1-7">`"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"`</a> | `"15.0.0.1,16.0.0.1"` | `"13.0.0.1"` |
| <a id="10-0-0-111-0-0-1" href="#10-0-0-111-0-0-1" title="#10-0-0-111-0-0-1">`"10.0.0.1,11.0.0.1"`</a> | `"10.0.0.1,11.0.0.1"` | `""`         |
