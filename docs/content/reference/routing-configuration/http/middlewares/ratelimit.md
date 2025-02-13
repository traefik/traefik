---
title: "Traefik RateLimit Documentation"
description: "Traefik Proxy's HTTP RateLimit middleware ensures Services receive fair amounts of requests. Read the technical documentation."
---

The `rateLimit` middleware ensures that services will receive a *fair* amount of requests, and allows you to define what fair is.

It is based on a [token bucket](https://en.wikipedia.org/wiki/Token_bucket) implementation.
In this analogy, the `average` and `period` parameters define the **rate** at which the bucket refills, and the `burst` is the size (volume) of the bucket

## Rate and Burst

The rate is defined by dividing `average` by `period`.
For a rate below 1 req/s, define a `period` larger than a second

## Configuration Example

```yaml tab="Structured (YAML)"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 200 requests is allowed.
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        average: 100
        burst: 200
```

```toml tab="Structured (TOML)"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 200 requests is allowed.
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    average = 100
    burst = 200
```

```yaml tab="Labels"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 200 requests is allowed.
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.average=100"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.burst=200"
```

```json tab="Tags"
// Here, an average of 100 requests per second is allowed.
// In addition, a burst of 200 requests is allowed.
{
  "Tags": [
    "traefik.http.middlewares.test-ratelimit.ratelimit.average=100",
    "traefik.http.middlewares.test-ratelimit.ratelimit.burst=50"
  ]
}
```

```yaml tab="Kubernetes"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 200 requests is allowed.
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    average: 100
    burst: 200
```

## Configuration Options

| Field      | Description       | Default | Required |
|:-----------|:-------------------------------------------------------|:--------|:---------|
| `average` | Number of requests used to define the rate using the `period`.<br /> 0 means **no rate limiting**.<br />More information [here](#rate-and-burst). | 0      | No      |
| `period` | Period of time used to define the rate.<br />More information [here](#rate-and-burst). | 1s | No |
| `burst` | Maximum number of requests allowed to go through at the very same moment.<br />More information [here](#rate-and-burst).| 1 | No |
| `sourceCriterion.requestHost` | Whether to consider the request host as the source.<br />More information about `sourceCriterion`[here](#sourcecriterion). | false      | No      |
| `sourceCriterion.requestHeaderName` | Name of the header used to group incoming requests.<br />More information about `sourceCriterion`[here](#sourcecriterion). | ""      | No      |
| `sourceCriterion.ipStrategy.depth` | Depth position of the IP to select in the `X-Forwarded-For` header (starting from the right).<br />0 means no depth.<br />If greater than the total number of IPs in `X-Forwarded-For`, then the client IP is empty<br />If higher than 0, the `excludedIPs` options is not evaluated.<br />More information about [`sourceCriterion`](#sourcecriterion), [`ipStrategy`](#ipstrategy), and [`depth`](#sourcecriterionipstrategydepth) below. | 0      | No      |
| `sourceCriterion.ipStrategy.excludedIPs` | Allows scanning the `X-Forwarded-For` header and select the first IP not in the list.<br />If `depth` is specified, `excludedIPs` is ignored.<br />More information about [`sourceCriterion`](#sourcecriterion), [`ipStrategy`](#ipstrategy), and [`excludedIPs`](#sourcecriterionipstrategyexcludedips) below. |       | No      |
| `sourceCriterion.ipStrategy.ipv6Subnet` |  If `ipv6Subnet` is provided and the selected IP is IPv6, the IP is transformed into the first IP of the subnet it belongs to. <br />More information about [`sourceCriterion`](#sourcecriterion), [`ipStrategy.ipv6Subnet`](#sourcecriterionipstrategyipv6subnet) below. |       | No      |

### sourceCriterion

The `sourceCriterion` option defines what criterion is used to group requests as originating from a common source.
If several strategies are defined at the same time, an error will be raised.
If none are set, the default is to use the request's remote address field (as an `ipStrategy`).

### ipStrategy

The `ipStrategy` option defines three parameters that configures how Traefik determines the client IP: `depth`, `excludedIPs` and `ipv6Subnet`.

As a middleware, rate-limiting happens before the actual proxying to the backend takes place.
In addition, the previous network hop only gets appended to `X-Forwarded-For` during the last stages of proxying, that is after it has already passed through rate-limiting.
Therefore, during rate-limiting, as the previous network hop is not yet present in `X-Forwarded-For`, it cannot be found and/or relied upon.

### `sourceCriterion.ipStrategy.ipv6Subnet`

This strategy applies to `Depth` and `RemoteAddr` strategy only.
If `ipv6Subnet` is provided and the selected IP is IPv6, the IP is transformed into the first IP of the subnet it belongs to.

This is useful for grouping IPv6 addresses into subnets to prevent bypassing this middleware by obtaining a new IPv6.

- `ipv6Subnet` is ignored if its value is outside of 0-128 interval

#### Example of ipv6Subnet

If `ipv6Subnet` is provided, the IP is transformed in the following way.

| `IP`                      | `ipv6Subnet` | clientIP              |
|---------------------------|--------------|-----------------------|
| `"::abcd:1111:2222:3333"` | `64`         | `"::0:0:0:0"`         |
| `"::abcd:1111:2222:3333"` | `80`         | `"::abcd:0:0:0:0"`    |
| `"::abcd:1111:2222:3333"` | `96`         | `"::abcd:1111:0:0:0"` |

### sourceCriterion.ipStrategy.depth

If `depth` is set to 2, and the request `X-Forwarded-For` header is `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` then the "real" client IP is `"10.0.0.1"` (at depth 4) but the IP used as the criterion is `"12.0.0.1"` (`depth=2`).

| `X-Forwarded-For`                       | `depth` | clientIP     |
|-----------------------------------------|---------|--------------|
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `1`     | `"13.0.0.1"` |
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `3`     | `"11.0.0.1"` |
| `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `5`     | `""`         |

### sourceCriterion.ipStrategy.excludedIPs

Contrary to what the name might suggest, this option is *not* about excluding an IP from the rate limiter, and therefore cannot be used to deactivate rate limiting for some IPs.

`excludedIPs` is meant to address two classes of somewhat distinct use-cases:

1. Distinguish IPs which are behind the same (set of) reverse-proxies so that each of them contributes, independently to the others, to its own rate-limit "bucket" (cf the [token bucket](https://en.wikipedia.org/wiki/Token_bucket)).
In this case, `excludedIPs` should be set to match the list of `X-Forwarded-For IPs` that are to be excluded, in order to find the actual clientIP.

Example to use each IP as a distinct source:

| X-Forwarded-For                | excludedIPs           | clientIP     |
|--------------------------------|-----------------------|--------------|
| `"10.0.0.1,11.0.0.1,12.0.0.1"` | `"11.0.0.1,12.0.0.1"` | `"10.0.0.1"` |
| `"10.0.0.2,11.0.0.1,12.0.0.1"` | `"11.0.0.1,12.0.0.1"` | `"10.0.0.2"` |

2. Group together a set of IPs (also behind a common set of reverse-proxies) so that they are considered the same source, and all contribute to the same rate-limit bucket.

Example to group IPs together as same source:

|  X-Forwarded-For               |  excludedIPs | clientIP     |
|--------------------------------|--------------|--------------|
| `"10.0.0.1,11.0.0.1,12.0.0.1"` | `"12.0.0.1"` | `"11.0.0.1"` |
| `"10.0.0.2,11.0.0.1,12.0.0.1"` | `"12.0.0.1"` | `"11.0.0.1"` |
| `"10.0.0.3,11.0.0.1,12.0.0.1"` | `"12.0.0.1"` | `"11.0.0.1"` |
