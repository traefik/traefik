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
# Redis distributed rate limiting is configured with all available options.
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        average: 100
        period: 1s
        burst: 200
        redis:
          endpoints:
            - "redis-primary.example.com:6379"
            - "redis-replica.example.com:6379"
          username: "ratelimit-user"
          password: "secure-password"
          db: 2
          poolSize: 50
          minIdleConns: 10
          maxActiveConns: 200
          readTimeout: 3s
          writeTimeout: 3s
          dialTimeout: 5s
          tls:
            ca: "/etc/ssl/redis-ca.crt"
            cert: "/etc/ssl/redis-client.crt"
            key: "/etc/ssl/redis-client.key"
            insecureSkipVerify: false
```

```toml tab="Structured (TOML)"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 200 requests is allowed.
# Redis distributed rate limiting is configured with all available options.
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    average = 100
    period = "1s"
    burst = 200
    [http.middlewares.test-ratelimit.rateLimit.redis]
      endpoints = ["redis-primary.example.com:6379", "redis-replica.example.com:6379"]
      username = "ratelimit-user"
      password = "secure-password"
      db = 2
      poolSize = 50
      minIdleConns = 10
      maxActiveConns = 200
      readTimeout = "3s"
      writeTimeout = "3s"
      dialTimeout = "5s"
      [http.middlewares.test-ratelimit.rateLimit.redis.tls]
        ca = "/etc/ssl/redis-ca.crt"
        cert = "/etc/ssl/redis-client.crt"
        key = "/etc/ssl/redis-client.key"
        insecureSkipVerify = false
```

```yaml tab="Labels"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 200 requests is allowed.
# Redis distributed rate limiting is configured with all available options.
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.average=100"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.period=1s"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.burst=200"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.endpoints=redis-primary.example.com:6379,redis-replica.example.com:6379"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.username=ratelimit-user"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.password=secure-password"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.db=2"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.poolSize=50"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.minIdleConns=10"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.maxActiveConns=200"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.readTimeout=3s"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.writeTimeout=3s"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.dialTimeout=5s"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.ca=/etc/ssl/redis-ca.crt"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.cert=/etc/ssl/redis-client.crt"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.key=/etc/ssl/redis-client.key"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.insecureSkipVerify=false"
```

```json tab="Tags"
// Here, an average of 100 requests per second is allowed.
// In addition, a burst of 200 requests is allowed.
// Redis distributed rate limiting is configured with all available options.
{
  "Tags": [
    "traefik.http.middlewares.test-ratelimit.ratelimit.average=100",
    "traefik.http.middlewares.test-ratelimit.ratelimit.period=1s",
    "traefik.http.middlewares.test-ratelimit.ratelimit.burst=200",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.endpoints=redis-primary.example.com:6379,redis-replica.example.com:6379",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.username=ratelimit-user",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.password=secure-password",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.db=2",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.poolSize=50",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.minIdleConns=10",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.maxActiveConns=200",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.readTimeout=3s",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.writeTimeout=3s",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.dialTimeout=5s",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.ca=/etc/ssl/redis-ca.crt",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.cert=/etc/ssl/redis-client.crt",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.key=/etc/ssl/redis-client.key",
    "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.insecureSkipVerify=false"
  ]
}
```

```yaml tab="Kubernetes"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 200 requests is allowed.
# Redis distributed rate limiting is configured with all available options.
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    average: 100
    period: 1s
    burst: 200
    redis:
      endpoints:
        - "redis-primary.example.com:6379"
        - "redis-replica.example.com:6379"
      secret: redis-credentials
      db: 2
      poolSize: 50
      minIdleConns: 10
      maxActiveConns: 200
      readTimeout: 3s
      writeTimeout: 3s
      dialTimeout: 5s
      tls:
        caSecret: redis-ca
        certSecret: redis-client-cert
        insecureSkipVerify: false

---
apiVersion: v1
kind: Secret
metadata:
  name: redis-credentials
  namespace: default
data:
  username: cmF0ZWxpbWl0LXVzZXI=  # base64 encoded "ratelimit-user"
  password: c2VjdXJlLXBhc3N3b3Jk  # base64 encoded "secure-password"

---
apiVersion: v1
kind: Secret
metadata:
  name: redis-ca
  namespace: default
data:
  tls.ca: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t...

---
apiVersion: v1
kind: Secret
metadata:
  name: redis-client-cert
  namespace: default
data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t...
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0t...
```

## Configuration Options

| Field      | Description       | Default | Required |
|:-----------|:-------------------------------------------------------|:--------|:---------|
| <a id="opt-average" href="#opt-average" title="#opt-average">`average`</a> | Number of requests used to define the rate using the `period`.<br /> 0 means **no rate limiting**.<br />More information [here](#rate-and-burst). | 0      | No      |
| <a id="opt-period" href="#opt-period" title="#opt-period">`period`</a> | Period of time used to define the rate.<br />More information [here](#rate-and-burst). | 1s | No |
| <a id="opt-burst" href="#opt-burst" title="#opt-burst">`burst`</a> | Maximum number of requests allowed to go through at the very same moment.<br />More information [here](#rate-and-burst).| 1 | No |
| <a id="opt-sourceCriterion-requestHost" href="#opt-sourceCriterion-requestHost" title="#opt-sourceCriterion-requestHost">`sourceCriterion.requestHost`</a> | Whether to consider the request host as the source.<br />More information about `sourceCriterion`[here](#sourcecriterion). | false      | No      |
| <a id="opt-sourceCriterion-requestHeaderName" href="#opt-sourceCriterion-requestHeaderName" title="#opt-sourceCriterion-requestHeaderName">`sourceCriterion.requestHeaderName`</a> | Name of the header used to group incoming requests.<br />More information about `sourceCriterion`[here](#sourcecriterion). | ""      | No      |
| <a id="opt-sourceCriterion-ipStrategy-depth" href="#opt-sourceCriterion-ipStrategy-depth" title="#opt-sourceCriterion-ipStrategy-depth">`sourceCriterion.ipStrategy.depth`</a> | Depth position of the IP to select in the `X-Forwarded-For` header (starting from the right).<br />0 means no depth.<br />If greater than the total number of IPs in `X-Forwarded-For`, then the client IP is empty<br />If higher than 0, the `excludedIPs` options is not evaluated.<br />More information about [`sourceCriterion`](#sourcecriterion), [`ipStrategy`](#ipstrategy), and [`depth`](#sourcecriterionipstrategydepth) below. | 0      | No      |
| <a id="opt-sourceCriterion-ipStrategy-excludedIPs" href="#opt-sourceCriterion-ipStrategy-excludedIPs" title="#opt-sourceCriterion-ipStrategy-excludedIPs">`sourceCriterion.ipStrategy.excludedIPs`</a> | Allows scanning the `X-Forwarded-For` header and select the first IP not in the list.<br />If `depth` is specified, `excludedIPs` is ignored.<br />More information about [`sourceCriterion`](#sourcecriterion), [`ipStrategy`](#ipstrategy), and [`excludedIPs`](#sourcecriterionipstrategyexcludedips) below. |       | No      |
| <a id="opt-sourceCriterion-ipStrategy-ipv6Subnet" href="#opt-sourceCriterion-ipStrategy-ipv6Subnet" title="#opt-sourceCriterion-ipStrategy-ipv6Subnet">`sourceCriterion.ipStrategy.ipv6Subnet`</a> |  If `ipv6Subnet` is provided and the selected IP is IPv6, the IP is transformed into the first IP of the subnet it belongs to. <br />More information about [`sourceCriterion`](#sourcecriterion), [`ipStrategy.ipv6Subnet`](#sourcecriterionipstrategyipv6subnet) below. |       | No      |
| <a id="opt-redis" href="#opt-redis" title="#opt-redis">`redis`</a> | The `redis` configuration enables distributed rate limiting by using Redis to store rate limit tokens across multiple Traefik instances. This allows you to enforce consistent rate limits across a cluster of Traefik proxies. <br />When Redis is not configured, Traefik uses in-memory storage for rate limiting, which works only for the individual Traefik instance.|       | No      |
| <a id="opt-redis-endpoints" href="#opt-redis-endpoints" title="#opt-redis-endpoints">`redis.endpoints`</a> | List of Redis server endpoints for distributed rate limiting. You can specify multiple endpoints for Redis cluster or high availability setups. | "127.0.0.1:6379" | No |
| <a id="opt-redis-username" href="#opt-redis-username" title="#opt-redis-username">`redis.username`</a> | Username for Redis authentication. | "" | No |
| <a id="opt-redis-password" href="#opt-redis-password" title="#opt-redis-password">`redis.password`</a> | Password for Redis authentication. In Kubernetes, these can be provided via secrets. | "" | No |
| <a id="opt-redis-db" href="#opt-redis-db" title="#opt-redis-db">`redis.db`</a> | Redis database number to select. | 0 | No |
| <a id="opt-redis-poolSize" href="#opt-redis-poolSize" title="#opt-redis-poolSize">`redis.poolSize`</a> | Defines the base number of socket connections in the pool. If set to 0, it defaults to 10 connections per CPU core as reported by `runtime.GOMAXPROCS`. <br />If there are not enough connections in the pool, new connections will be allocated beyond `poolSize`, up to `maxActiveConns`. | 0 | No |
| <a id="opt-redis-minIdleConns" href="#opt-redis-minIdleConns" title="#opt-redis-minIdleConns">`redis.minIdleConns`</a> | Minimum number of idle connections to maintain in the pool. This is useful when establishing new connections is slow. A value of 0 means idle connections are not automatically closed. | 0 | No |
| <a id="opt-redis-maxActiveConns" href="#opt-redis-maxActiveConns" title="#opt-redis-maxActiveConns">`redis.maxActiveConns`</a> | Maximum number of connections the pool can allocate at any given time. A value of 0 means no limit. | 0 | No |
| <a id="opt-redis-readTimeout" href="#opt-redis-readTimeout" title="#opt-redis-readTimeout">`redis.readTimeout`</a> | Timeout for socket reads. If reached, commands will fail with a timeout instead of blocking. Zero means no timeout. | 3s | No |
| <a id="opt-redis-writeTimeout" href="#opt-redis-writeTimeout" title="#opt-redis-writeTimeout">`redis.writeTimeout`</a> | Timeout for socket writes. If reached, commands will fail with a timeout instead of blocking. Zero means no timeout. | 3s | No |
| <a id="opt-redis-dialTimeout" href="#opt-redis-dialTimeout" title="#opt-redis-dialTimeout">`redis.dialTimeout`</a> | Timeout for establishing new connections. Zero means no timeout. | 5s | No |
| <a id="opt-redis-tls-ca" href="#opt-redis-tls-ca" title="#opt-redis-tls-ca">`redis.tls.ca`</a> | Path to the certificate authority used for the secure connection to Redis, it defaults to the system bundle. | "" | No |
| <a id="opt-redis-tls-cert" href="#opt-redis-tls-cert" title="#opt-redis-tls-cert">`redis.tls.cert`</a> | Path to the public certificate used for the secure connection to Redis. When this option is set, the `key` option is required. | "" | No |
| <a id="opt-redis-tls-key" href="#opt-redis-tls-key" title="#opt-redis-tls-key">`redis.tls.key`</a> | Path to the private key used for the secure connection to Redis. When this option is set, the `cert` option is required. | "" | No |
| <a id="opt-redis-tls-insecureSkipVerify" href="#opt-redis-tls-insecureSkipVerify" title="#opt-redis-tls-insecureSkipVerify">`redis.tls.insecureSkipVerify`</a> | If `insecureSkipVerify` is `true`, the TLS connection to Redis accepts any certificate presented by the server regardless of the hostnames it covers. | false | No |

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
| <a id="opt-abcd111122223333" href="#opt-abcd111122223333" title="#opt-abcd111122223333">`"::abcd:1111:2222:3333"`</a> | `64`         | `"::0:0:0:0"`         |
| <a id="opt-abcd111122223333-2" href="#opt-abcd111122223333-2" title="#opt-abcd111122223333-2">`"::abcd:1111:2222:3333"`</a> | `80`         | `"::abcd:0:0:0:0"`    |
| <a id="opt-abcd111122223333-3" href="#opt-abcd111122223333-3" title="#opt-abcd111122223333-3">`"::abcd:1111:2222:3333"`</a> | `96`         | `"::abcd:1111:0:0:0"` |

### sourceCriterion.ipStrategy.depth

If `depth` is set to 2, and the request `X-Forwarded-For` header is `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` then the "real" client IP is `"10.0.0.1"` (at depth 4) but the IP used as the criterion is `"12.0.0.1"` (`depth=2`).

| `X-Forwarded-For`                       | `depth` | clientIP     |
|-----------------------------------------|---------|--------------|
| <a id="opt-10-0-0-111-0-0-112-0-0-113-0-0-1" href="#opt-10-0-0-111-0-0-112-0-0-113-0-0-1" title="#opt-10-0-0-111-0-0-112-0-0-113-0-0-1">`"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"`</a> | `1`     | `"13.0.0.1"` |
| <a id="opt-10-0-0-111-0-0-112-0-0-113-0-0-1-2" href="#opt-10-0-0-111-0-0-112-0-0-113-0-0-1-2" title="#opt-10-0-0-111-0-0-112-0-0-113-0-0-1-2">`"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"`</a> | `3`     | `"11.0.0.1"` |
| <a id="opt-10-0-0-111-0-0-112-0-0-113-0-0-1-3" href="#opt-10-0-0-111-0-0-112-0-0-113-0-0-1-3" title="#opt-10-0-0-111-0-0-112-0-0-113-0-0-1-3">`"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"`</a> | `5`     | `""`         |

### sourceCriterion.ipStrategy.excludedIPs

Contrary to what the name might suggest, this option is *not* about excluding an IP from the rate limiter, and therefore cannot be used to deactivate rate limiting for some IPs.

`excludedIPs` is meant to address two classes of somewhat distinct use-cases:

1. Distinguish IPs which are behind the same (set of) reverse-proxies so that each of them contributes, independently to the others, to its own rate-limit "bucket" (cf the [token bucket](https://en.wikipedia.org/wiki/Token_bucket)).
In this case, `excludedIPs` should be set to match the list of `X-Forwarded-For IPs` that are to be excluded, in order to find the actual clientIP.

Example to use each IP as a distinct source:

| `X-Forwarded-For`              | excludedIPs           | clientIP     |
|--------------------------------|-----------------------|--------------|
| <a id="opt-10-0-0-111-0-0-112-0-0-1" href="#opt-10-0-0-111-0-0-112-0-0-1" title="#opt-10-0-0-111-0-0-112-0-0-1">`"10.0.0.1,11.0.0.1,12.0.0.1"`</a> | `"11.0.0.1,12.0.0.1"` | `"10.0.0.1"` |
| <a id="opt-10-0-0-211-0-0-112-0-0-1" href="#opt-10-0-0-211-0-0-112-0-0-1" title="#opt-10-0-0-211-0-0-112-0-0-1">`"10.0.0.2,11.0.0.1,12.0.0.1"`</a> | `"11.0.0.1,12.0.0.1"` | `"10.0.0.2"` |

2. Group together a set of IPs (also behind a common set of reverse-proxies) so that they are considered the same source, and all contribute to the same rate-limit bucket.

Example to group IPs together as same source:

| `X-Forwarded-For`              | excludedIPs  | clientIP     |
|--------------------------------|--------------|--------------|
| <a id="opt-10-0-0-111-0-0-112-0-0-1-2" href="#opt-10-0-0-111-0-0-112-0-0-1-2" title="#opt-10-0-0-111-0-0-112-0-0-1-2">`"10.0.0.1,11.0.0.1,12.0.0.1"`</a> | `"12.0.0.1"` | `"11.0.0.1"` |
| <a id="opt-10-0-0-211-0-0-112-0-0-1-2" href="#opt-10-0-0-211-0-0-112-0-0-1-2" title="#opt-10-0-0-211-0-0-112-0-0-1-2">`"10.0.0.2,11.0.0.1,12.0.0.1"`</a> | `"12.0.0.1"` | `"11.0.0.1"` |
| <a id="opt-10-0-0-311-0-0-112-0-0-1" href="#opt-10-0-0-311-0-0-112-0-0-1" title="#opt-10-0-0-311-0-0-112-0-0-1">`"10.0.0.3,11.0.0.1,12.0.0.1"`</a> | `"12.0.0.1"` | `"11.0.0.1"` |
