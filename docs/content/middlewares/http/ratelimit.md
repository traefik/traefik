---
title: "Traefik RateLimit Documentation"
description: "Traefik Proxy's HTTP RateLimit middleware ensures Services receive fair amounts of requests. Read the technical documentation."
---

# RateLimit

To Control the Number of Requests Going to a Service
{: .subtitle }

The RateLimit middleware ensures that services will receive a _fair_ amount of requests, and allows one to define what fair is.

It is based on a [token bucket](https://en.wikipedia.org/wiki/Token_bucket) implementation. In this analogy, the [average](#average) parameter (defined below) is the rate at which the bucket refills, and the [burst](#burst) is the size (volume) of the bucket.

## Configuration Example

```yaml tab="Docker & Swarm"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 200 requests is allowed.
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.average=100"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.burst=200"
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

```yaml tab="Consul Catalog"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 200 requests is allowed.
- "traefik.http.middlewares.test-ratelimit.ratelimit.average=100"
- "traefik.http.middlewares.test-ratelimit.ratelimit.burst=50"
```

```yaml tab="File (YAML)"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 200 requests is allowed.
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        average: 100
        burst: 200
```

```toml tab="File (TOML)"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 200 requests is allowed.
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    average = 100
    burst = 200
```

## Configuration Options

### `average`

`average` is the maximum rate, by default in requests per second, allowed from a given source.

It defaults to `0`, which means no rate limiting.

The rate is actually defined by dividing `average` by `period`.
So for a rate below 1 req/s, one needs to define a `period` larger than a second.

```yaml tab="Docker & Swarm"
# 100 reqs/s
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.average=100"
```

```yaml tab="Kubernetes"
# 100 reqs/s
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    average: 100
```

```yaml tab="Consul Catalog"
# 100 reqs/s
- "traefik.http.middlewares.test-ratelimit.ratelimit.average=100"
```

```yaml tab="File (YAML)"
# 100 reqs/s
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        average: 100
```

```toml tab="File (TOML)"
# 100 reqs/s
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    average = 100
```

### `period`

`period`, in combination with `average`, defines the actual maximum rate, such as:

```go
r = average / period
```

It defaults to `1` second.

```yaml tab="Docker & Swarm"
# 6 reqs/minute
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.average=6"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.period=1m"
```

```yaml tab="Kubernetes"
# 6 reqs/minute
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    period: 1m
    average: 6
```

```yaml tab="Consul Catalog"
# 6 reqs/minute
- "traefik.http.middlewares.test-ratelimit.ratelimit.average=6"
- "traefik.http.middlewares.test-ratelimit.ratelimit.period=1m"
```

```yaml tab="File (YAML)"
# 6 reqs/minute
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        average: 6
        period: 1m
```

```toml tab="File (TOML)"
# 6 reqs/minute
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    average = 6
    period = "1m"
```

### `burst`

`burst` is the maximum number of requests allowed to go through in the same arbitrarily small period of time.

It defaults to `1`.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.burst=100"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    burst: 100
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.burst=100"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        burst: 100
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    burst = 100
```

### `sourceCriterion`

The `sourceCriterion` option defines what criterion is used to group requests as originating from a common source.
If several strategies are defined at the same time, an error will be raised.
If none are set, the default is to use the request's remote address field (as an `ipStrategy`).

#### `sourceCriterion.ipStrategy`

The `ipStrategy` option defines three parameters that configures how Traefik determines the client IP: `depth`, `excludedIPs` and `ipv6Subnet`.

!!! important "As a middleware, rate-limiting happens before the actual proxying to the backend takes place. In addition, the previous network hop only gets appended to `X-Forwarded-For` during the last stages of proxying, i.e. after it has already passed through rate-limiting. Therefore, during rate-limiting, as the previous network hop is not yet present in `X-Forwarded-For`, it cannot be found and/or relied upon."

##### `ipStrategy.depth`

The `depth` option tells Traefik to use the `X-Forwarded-For` header and select the IP located at the `depth` position (starting from the right).

- If `depth` is greater than the total number of IPs in `X-Forwarded-For`, then the client IP is empty.
- `depth` is ignored if its value is less than or equal to 0.

If `ipStrategy.ipv6Subnet` is provided and the selected IP is IPv6, the IP is transformed into the first IP of the subnet it belongs to.  
See [ipStrategy.ipv6Subnet](#ipstrategyipv6subnet) for more details.

!!! example "Example of Depth & X-Forwarded-For"

    If `depth` is set to 2, and the request `X-Forwarded-For` header is `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` then the "real" client IP is `"10.0.0.1"` (at depth 4) but the IP used as the criterion is `"12.0.0.1"` (`depth=2`).

    | `X-Forwarded-For`                       | `depth` | clientIP     |
    |-----------------------------------------|---------|--------------|
    | `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `1`     | `"13.0.0.1"` |
    | `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `3`     | `"11.0.0.1"` |
    | `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` | `5`     | `""`         |

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.sourcecriterion.ipstrategy.depth=2"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    sourceCriterion:
      ipStrategy:
        depth: 2
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.sourcecriterion.ipstrategy.depth=2"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        sourceCriterion:
          ipStrategy:
            depth: 2
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.sourceCriterion.ipStrategy]
      depth = 2
```

##### `ipStrategy.excludedIPs`

!!! important "Contrary to what the name might suggest, this option is _not_ about excluding an IP from the rate limiter, and therefore cannot be used to deactivate rate limiting for some IPs."

!!! important "If `depth` is specified, `excludedIPs` is ignored."

`excludedIPs` is meant to address two classes of somewhat distinct use-cases:

1. Distinguish IPs which are behind the same (set of) reverse-proxies so that each of them contributes, independently to the others,
   to its own rate-limit "bucket" (cf the [leaky bucket analogy](https://wikipedia.org/wiki/Leaky_bucket)).
   In this case, `excludedIPs` should be set to match the list of `X-Forwarded-For IPs` that are to be excluded,
   in order to find the actual clientIP.

    !!! example "Each IP as a distinct source"

        | X-Forwarded-For                | excludedIPs           | clientIP     |
        |--------------------------------|-----------------------|--------------|
        | `"10.0.0.1,11.0.0.1,12.0.0.1"` | `"11.0.0.1,12.0.0.1"` | `"10.0.0.1"` |
        | `"10.0.0.2,11.0.0.1,12.0.0.1"` | `"11.0.0.1,12.0.0.1"` | `"10.0.0.2"` |

2. Group together a set of IPs (also behind a common set of reverse-proxies) so that they are considered the same source,
   and all contribute to the same rate-limit bucket.

    !!! example "Group IPs together as same source"

        |  X-Forwarded-For               |  excludedIPs | clientIP     |
        |--------------------------------|--------------|--------------|
        | `"10.0.0.1,11.0.0.1,12.0.0.1"` | `"12.0.0.1"` | `"11.0.0.1"` |
        | `"10.0.0.2,11.0.0.1,12.0.0.1"` | `"12.0.0.1"` | `"11.0.0.1"` |
        | `"10.0.0.3,11.0.0.1,12.0.0.1"` | `"12.0.0.1"` | `"11.0.0.1"` |

For completeness, below are additional examples to illustrate how the matching works.
For a given request the list of `X-Forwarded-For` IPs is checked from most recent to most distant against the `excludedIPs` pool,
and the first IP that is _not_ in the pool (if any) is returned.

!!! example "Matching for clientIP"

    |  X-Forwarded-For               |  excludedIPs          | clientIP     |
    |--------------------------------|-----------------------|--------------|
    | `"10.0.0.1,11.0.0.1,13.0.0.1"` | `"11.0.0.1"`          | `"13.0.0.1"` |
    | `"10.0.0.1,11.0.0.1,13.0.0.1"` | `"15.0.0.1,16.0.0.1"` | `"13.0.0.1"` |
    | `"10.0.0.1,11.0.0.1"`          | `"10.0.0.1,11.0.0.1"` | `""`         |

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.sourcecriterion.ipstrategy.excludedips=127.0.0.1/32, 192.168.1.7"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    sourceCriterion:
      ipStrategy:
        excludedIPs:
        - 127.0.0.1/32
        - 192.168.1.7
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.sourcecriterion.ipstrategy.excludedips=127.0.0.1/32, 192.168.1.7"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        sourceCriterion:
          ipStrategy:
            excludedIPs:
              - "127.0.0.1/32"
              - "192.168.1.7"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.sourceCriterion.ipStrategy]
      excludedIPs = ["127.0.0.1/32", "192.168.1.7"]
```

##### `ipStrategy.ipv6Subnet`

This strategy applies to `Depth` and `RemoteAddr` strategy only.
If `ipv6Subnet` is provided and the selected IP is IPv6, the IP is transformed into the first IP of the subnet it belongs to.

This is useful for grouping IPv6 addresses into subnets to prevent bypassing this middleware by obtaining a new IPv6.

- `ipv6Subnet` is ignored if its value is outside of 0-128 interval

!!! example "Example of ipv6Subnet"

    If `ipv6Subnet` is provided, the IP is transformed in the following way.

    | `IP`                      | `ipv6Subnet` | clientIP              |
    |---------------------------|--------------|-----------------------|
    | `"::abcd:1111:2222:3333"` | `64`         | `"::0:0:0:0"`         |
    | `"::abcd:1111:2222:3333"` | `80`         | `"::abcd:0:0:0:0"`    |
    | `"::abcd:1111:2222:3333"` | `96`         | `"::abcd:1111:0:0:0"` |

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.sourcecriterion.ipstrategy.ipv6Subnet=64"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  ratelimit:
    sourceCriterion:
      ipStrategy:
        ipv6Subnet: 64
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.sourcecriterion.ipstrategy.ipv6Subnet=64"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      ratelimit:
        sourceCriterion:
          ipStrategy:
            ipv6Subnet: 64
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.ratelimit]
    [http.middlewares.test-ratelimit.ratelimit.sourceCriterion.ipStrategy]
      ipv6Subnet = 64
```

#### `sourceCriterion.requestHeaderName`

Name of the header used to group incoming requests.

!!! important "If the header is not present, rate limiting will still be applied, but all requests without the specified header will be grouped together."

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.sourcecriterion.requestheadername=username"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
	sourceCriterion:
      requestHeaderName: username
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.sourcecriterion.requestheadername=username"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        sourceCriterion:
          requestHeaderName: username
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.sourceCriterion]
      requestHeaderName = "username"
```

#### `sourceCriterion.requestHost`

Whether to consider the request host as the source.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.sourcecriterion.requesthost=true"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    sourceCriterion:
      requestHost: true
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.sourcecriterion.requesthost=true"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        sourceCriterion:
          requestHost: true
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.sourceCriterion]
      requestHost = true
```

### `redis`

Enables distributed rate limit using `redis` to store the tokens.
If not set, Traefik's in-memory storage is used by default.

#### `redis.endpoints`

_Required, Default="127.0.0.1:6379"_

Defines how to connect to the Redis server.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.endpoints=127.0.0.1:6379"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    # ...
    redis:
      endpoints:
        - "127.0.0.1:6379"
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.endpoints=127.0.0.1:6379"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        # ...
        redis:
          endpoints:
            - "127.0.0.1:6379"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.redis]
      endpoints = ["127.0.0.1:6379"]
```

#### `redis.username`

_Optional, Default=""_

Defines the username used to authenticate with the Redis server.

```yaml tab="Docker & Swarm"
labels:
    - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.username=user"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
   name: test-ratelimit
spec:
   rateLimit:
      # ...
      redis:
         secret: mysecret

---
apiVersion: v1
kind: Secret
metadata:
   name: mysecret
   namespace: default

data:
   username: dXNlcm5hbWU=
   password: cGFzc3dvcmQ=
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.username=user"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        # ...
        redis:
          username: user
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.redis]
      username = "user"
```

#### `redis.password`

_Optional, Default=""_

Defines the password to authenticate against the Redis server.

```yaml tab="Docker & Swarm"
labels:
    - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.password=password"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
   name: test-ratelimit
spec:
   rateLimit:
      # ...
      redis:
         secret: mysecret

---
apiVersion: v1
kind: Secret
metadata:
   name: mysecret
   namespace: default

data:
   username: dXNlcm5hbWU=
   password: cGFzc3dvcmQ=
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.password=password"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        # ...
        redis:
          password: password
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.redis]
      password = "password"
```

#### `redis.db`

_Optional, Default=0_

Defines the database to select after connecting to the Redis.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.db=0"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    # ...
    redis:
      db: 0
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.db=0"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        # ...
        redis:
          db: 0
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.redis]
      db = 0
```

#### `redis.tls`

Same as this [config](https://doc.traefik.io/traefik/providers/redis/#tls)

_Optional_

Defines the TLS configuration used for the secure connection to Redis.

##### `redis.tls.ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to Redis,
it defaults to the system bundle.

```yaml tab="Docker & Swarm"
labels:
    - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.ca=path/to/ca.crt"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    # ...
    redis:
      tls:
        caSecret: mycasercret

---
apiVersion: v1
kind: Secret
metadata:
  name: mycasercret
  namespace: default

data:
  # Must contain a certificate under either a `tls.ca` or a `ca.crt` key. 
  tls.ca: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.ca=path/to/ca.crt"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    rateLimit:
      # ... 
      redis:
        tls:
          ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[providers.redis.tls]
  ca = "path/to/ca.crt"
```

##### `redis.tls.cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to Redis.
When this option is set, the `key` option is required.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.cert=path/to/foo.cert"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.key=path/to/foo.key"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
   name: test-ratelimit
spec:
   rateLimit:
      # ...
      redis:
         tls:
           certSecret: mytlscert

---
apiVersion: v1
kind: Secret
metadata:
   name: mytlscert
   namespace: default

data:
   tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
   tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0=
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.cert=path/to/foo.cert"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.key=path/to/foo.key"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        redis:
          tls:
            cert: path/to/foo.cert
            key: path/to/foo.key
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.redis]
      [http.middlewares.test-ratelimit.rateLimit.redis.tls]
        cert = "path/to/foo.cert"
        key = "path/to/foo.key"
```

##### `redis.tls.key`

_Optional_

`key` is the path to the private key used for the secure connection to Redis.
When this option is set, the `cert` option is required.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.cert=path/to/foo.cert"
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.key=path/to/foo.key"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
   name: test-ratelimit
spec:
   rateLimit:
      # ...
      redis:
         tls:
            certSecret: mytlscert

---
apiVersion: v1
kind: Secret
metadata:
   name: mytlscert
   namespace: default

data:
   tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
   tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0=
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.cert=path/to/foo.cert"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.key=path/to/foo.key"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        redis:
          tls:
            cert: path/to/foo.cert
            key: path/to/foo.key
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.redis]
      [http.middlewares.test-ratelimit.rateLimit.redis.tls]
        cert = "path/to/foo.cert"
        key = "path/to/foo.key"
```

##### `redis.tls.insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`, the TLS connection to Redis accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.insecureSkipVerify=true"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    # ...
    redis:
      tls:
        insecureSkipVerify: true
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.tls.insecureSkipVerify=true"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        # ...
        redis:
          tls:
            insecureSkipVerify: true
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.redis]
      [http.middlewares.test-ratelimit.rateLimit.redis.tls]
        insecureSkipVerify = true
```

#### `redis.poolSize`

_Optional, Default=0_

Defines the base number of socket connections.

If there are not enough connections in the pool, new connections will be allocated beyond `redis.poolSize`. 
You can limit this using `redis.maxActiveConns`.

Zero means 10 connections per every available CPU as reported by runtime.GOMAXPROCS.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.poolSize=42"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    # ...
    redis:
      poolSize: 42
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.poolSize=42"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        # ...
        redis:
          poolSize: 42
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.redis]
      poolSize = 42
```

#### `redis.minIdleConns`

_Optional, Default=0_

Defines the minimum number of idle connections, which is useful when establishing new connections is slow.
Zero means that idle connections are not closed.

```yaml tab="Docker & Swarm"
labels:
    - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.minIdleConns=42"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    # ...
    redis:
      minIdleConns: 42
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.minIdleConns=42"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        # ...
        redis:
          minIdleConns: 42
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.redis]
      minIdleConns = 42
```

#### `redis.maxActiveConns`

_Optional, Default=0_

Defines the maximum number of connections the pool can allocate at a given time.
Zero means no limit.

```yaml tab="Docker & Swarm"
labels:
    - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.maxActiveConns=42"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    # ...
    redis:
      maxActiveConns: 42
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.maxActiveConns=42"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        # ...
        redis:
          maxActiveConns: 42
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.redis]
      maxActiveConns = 42
```

#### `redis.readTimeout`

_Optional, Default=3s_

Defines the timeout for socket reads. 
If reached, commands will fail with a timeout instead of blocking.
Zero means no timeout.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.readTimeout=42s"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    # ...
    redis:
      readTimeout: 42s
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.readTimeout=42s"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        # ...
        redis:
          readTimeout: 42s
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.redis]
      readTimeout = "42s"
```

#### `redis.writeTimeout`

_Optional, Default=3s_

Defines the timeout for socket writes. 
If reached, commands will fail with a timeout instead of blocking. 
Zero means no timeout.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.writeTimeout=42s"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    # ...
    redis:
      writeTimeout: 42s
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.writeTimeout=42s"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        # ...
        redis:
          writeTimeout: 42s
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.redis]
      writeTimeout = "42s"
```

#### `redis.dialTimeout`

_Optional, Default=5s_

Defines the dial timeout for establishing new connections.
Zero means no timeout.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-ratelimit.ratelimit.redis.dialTimeout=42s"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    # ...
    redis:
      dialTimeout: 42s
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-ratelimit.ratelimit.redis.dialTimeout=42s"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-ratelimit:
      rateLimit:
        # ...
        redis:
          dialTimeout: 42s
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-ratelimit.rateLimit]
    [http.middlewares.test-ratelimit.rateLimit.redis]
      dialTimeout = "42s"
```
