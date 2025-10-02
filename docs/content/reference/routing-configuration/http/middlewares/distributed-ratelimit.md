---
title: "Distributed RateLimit"
description: "Traefik Hub API Gateway - The Distributed RateLimit middleware ensures Services receive fair amounts of requests throughout your cluster and not only on an individual proxy."
---

!!! info "Traefik Hub Feature"
    This middleware is available exclusively in [Traefik Hub](https://traefik.io/traefik-hub/). Learn more about [Traefik Hub's advanced features](https://doc.traefik.io/traefik-hub/api-gateway/intro).

The Distributed RateLimit middleware ensures that requests are limited over time throughout your cluster and not only on an individual proxy.

It is based on a [token bucket](https://en.wikipedia.org/wiki/Token_bucket) implementation.

---

## Configuration Example

Below is an advanced configuration that enables the Distributed RateLimit middleware with Redis backend for cluster-wide rate limiting.

```yaml tab="Middleware Distributed Rate Limit"
# Here, a limit of 100 requests per second is allowed.
# In addition, a burst of 200 requests is allowed.
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-distributedratelimit
  namespace: traefik
spec:
  plugin:
    distributedRateLimit:
      burst: 200
      denyOnError: false
      limit: 100
      period: 1s
      responseHeaders: true
      sourceCriterion:
        ipStrategy:
          excludedIPs:
            - 172.20.176.201
      store:
        redis:
          endpoints:
            - my-release-redis-master.default.svc.cluster.local:6379
          # Use the field password of the Secret redis in the same namespace
          password: urn:k8s:secret:redis:password
          timeout: 500ms
```

```yaml tab="Kubernetes Secret"
apiVersion: v1
kind: Secret
metadata:
  name: redis
  namespace: traefik
stringData:
  password: mysecret12345678
```

## Rate and Burst

The rate is defined by dividing `limit` by `period`.
For a rate below 1 req/s, define a `period` larger than a second

The middleware is based on a [token bucket](https://en.wikipedia.org/wiki/Token_bucket) implementation.
In this analogy, the `limit` and `period` parameters define the **rate** at which the bucket refills, and the `burst` is the size (volume) of the bucket.

```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-ratelimit
spec:
  plugin:
    distributedRateLimit:
      burst: 100
      period: 1m
      limit: 6
```

In the example above, the middleware allows up to 100 connections in parallel (`burst`).
Each connection consume a token, once the 100 tokens are consumed, the other ones are blocked until at least one token is available in the bucket.

When the bucket is not full, on token is generated every 10 seconds (6 every 1 minutes (`period` / `limit`)).

## Configuration Options

| Field      | Description                                                                                                                                                                                 | Default | Required |
|:-----------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| <a id="limit" href="#limit" title="#limit">`limit`</a> | Number of requests used to define the rate using the `period`.<br /> 0 means **no rate limiting**.<br />More information [here](#rate-and-burst).| 0      | No      |
| <a id="period" href="#period" title="#period">`period`</a> | Period of time used to define the rate.<br />More information [here](#rate-and-burst).| 1s | No |
| <a id="burst" href="#burst" title="#burst">`burst`</a> | Maximum number of requests allowed to go through at the very same moment.<br />More information [here](#rate-and-burst). | 1 | No |
| <a id="denyOnError" href="#denyOnError" title="#denyOnError">`denyOnError`</a> | Forces to return a 429 error if the number of remaining requests accepted cannot be get.<br /> Set to `false`, this option allows the request to reach the backend. | true    | No       |
| <a id="responseHeaders" href="#responseHeaders" title="#responseHeaders">`responseHeaders`</a> | Injects the following rate limiting headers in the response:<br />- X-Rate-Limit-Remaining<br />- X-Rate-Limit-Limit<br />- X-Rate-Limit-Period<br />- X-Rate-Limit-Reset<br />The added headers indicate how many tokens are left in the bucket (in the token bucket analogy) after the reservation for the request was made. | false   | No       |
| <a id="store-redis-endpoints" href="#store-redis-endpoints" title="#store-redis-endpoints">`store.redis.endpoints`</a> | Endpoints of the Redis instances to connect to (example: `redis.traefik-hub.svc.cluster.local:6379`) | "" | Yes      |
| <a id="store-redis-username" href="#store-redis-username" title="#store-redis-username">`store.redis.username`</a> | The username Traefik Hub will use to connect to Redis                                                | "" | No       |
| <a id="store-redis-password" href="#store-redis-password" title="#store-redis-password">`store.redis.password`</a> | The password Traefik Hub will use to connect to Redis                                                | "" | No       |
| <a id="store-redis-database" href="#store-redis-database" title="#store-redis-database">`store.redis.database`</a> | The database Traefik Hub will use to sore information (default: `0`)                                 | "" | No       |
| <a id="store-redis-cluster" href="#store-redis-cluster" title="#store-redis-cluster">`store.redis.cluster`</a> | Enable Redis Cluster                                                                                 | "" | No       |
| <a id="store-redis-tls-caBundle" href="#store-redis-tls-caBundle" title="#store-redis-tls-caBundle">`store.redis.tls.caBundle`</a> | Custom CA bundle                                                                                     | "" | No       |
| <a id="store-redis-tls-cert" href="#store-redis-tls-cert" title="#store-redis-tls-cert">`store.redis.tls.cert`</a> | TLS certificate                                                                                      | "" | No       |
| <a id="store-redis-tls-key" href="#store-redis-tls-key" title="#store-redis-tls-key">`store.redis.tls.key`</a> | TLS key                                                                                              | "" | No       |
| <a id="store-redis-tls-insecureSkipVerify" href="#store-redis-tls-insecureSkipVerify" title="#store-redis-tls-insecureSkipVerify">`store.redis.tls.insecureSkipVerify`</a> | Allow skipping the TLS verification                                                                  | "" | No       |
| <a id="store-redis-sentinel-masterSet" href="#store-redis-sentinel-masterSet" title="#store-redis-sentinel-masterSet">`store.redis.sentinel.masterSet`</a> | Name of the set of main nodes to use for main selection. Required when using Sentinel.               | "" | No       |
| <a id="store-redis-sentinel-username" href="#store-redis-sentinel-username" title="#store-redis-sentinel-username">`store.redis.sentinel.username`</a> | Username to use for sentinel authentication (can be different from `username`)                       | "" | No       |
| <a id="store-redis-sentinel-password" href="#store-redis-sentinel-password" title="#store-redis-sentinel-password">`store.redis.sentinel.password`</a> | Password to use for sentinel authentication (can be different from `password`)                       | "" | No       |
| <a id="sourceCriterion-requestHost" href="#sourceCriterion-requestHost" title="#sourceCriterion-requestHost">`sourceCriterion.requestHost`</a> | Whether to consider the request host as the source.<br />More information about `sourceCriterion`[here](#sourcecriterion). | false      | No      |
| <a id="sourceCriterion-requestHeaderName" href="#sourceCriterion-requestHeaderName" title="#sourceCriterion-requestHeaderName">`sourceCriterion.requestHeaderName`</a> | Name of the header used to group incoming requests.<br />More information about `sourceCriterion`[here](#sourcecriterion). | ""      | No      |
| <a id="sourceCriterion-ipStrategy-depth" href="#sourceCriterion-ipStrategy-depth" title="#sourceCriterion-ipStrategy-depth">`sourceCriterion.ipStrategy.depth`</a> | Depth position of the IP to select in the `X-Forwarded-For` header (starting from the right).<br />0 means no depth.<br />If greater than the total number of IPs in `X-Forwarded-For`, then the client IP is empty<br />If higher than 0, the `excludedIPs` options is not evaluated.<br />More information about [`sourceCriterion`](#sourcecriterion), [`ipStrategy`](#ipstrategy), and [`depth`](#sourcecriterionipstrategydepth) below. | 0      | No      |
| <a id="sourceCriterion-ipStrategy-excludedIPs" href="#sourceCriterion-ipStrategy-excludedIPs" title="#sourceCriterion-ipStrategy-excludedIPs">`sourceCriterion.ipStrategy.excludedIPs`</a> | Allows Traefik to scan the `X-Forwarded-For` header and select the first IP not in the list.<br />If `depth` is specified, `excludedIPs` is ignored.<br />More information about [`sourceCriterion`](#sourcecriterion), [`ipStrategy`](#ipstrategy), and [`excludedIPs`](#sourcecriterionipstrategyexcludedips) below. |       | No      |

### sourceCriterion

The `sourceCriterion` option defines what criterion is used to group requests as originating from a common source.
If several strategies are defined at the same time, an error will be raised.
If none are set, the default is to use the request's remote address field (as an `ipStrategy`).

### ipStrategy

The `ipStrategy` option defines two parameters that configures how Traefik determines the client IP: `depth`, and `excludedIPs`.

As a middleware, rate-limiting happens before the actual proxying to the backend takes place.
In addition, the previous network hop only gets appended to `X-Forwarded-For` during the last stages of proxying, that is after it has already passed through rate-limiting.
Therefore, during rate-limiting, as the previous network hop is not yet present in `X-Forwarded-For`, it cannot be found and/or relied upon.

### sourceCriterion.ipStrategy.depth

If `depth` is set to 2, and the request `X-Forwarded-For` header is `"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"` then the "real" client IP is `"10.0.0.1"` (at depth 4) but the IP used as the criterion is `"12.0.0.1"` (`depth=2`).

| `X-Forwarded-For`                       | `depth` | clientIP     |
|-----------------------------------------|---------|--------------|
| <a id="10-0-0-111-0-0-112-0-0-113-0-0-1" href="#10-0-0-111-0-0-112-0-0-113-0-0-1" title="#10-0-0-111-0-0-112-0-0-113-0-0-1">`"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"`</a> | `1`     | `"13.0.0.1"` |
| <a id="10-0-0-111-0-0-112-0-0-113-0-0-1-2" href="#10-0-0-111-0-0-112-0-0-113-0-0-1-2" title="#10-0-0-111-0-0-112-0-0-113-0-0-1-2">`"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"`</a> | `3`     | `"11.0.0.1"` |
| <a id="10-0-0-111-0-0-112-0-0-113-0-0-1-3" href="#10-0-0-111-0-0-112-0-0-113-0-0-1-3" title="#10-0-0-111-0-0-112-0-0-113-0-0-1-3">`"10.0.0.1,11.0.0.1,12.0.0.1,13.0.0.1"`</a> | `5`     | `""`         |

### sourceCriterion.ipStrategy.excludedIPs

Contrary to what the name might suggest, this option is *not* about excluding an IP from the rate limiter, and therefore cannot be used to deactivate rate limiting for some IPs.

`excludedIPs` is meant to address two classes of somewhat distinct use-cases:

1. Distinguish IPs which are behind the same (set of) reverse-proxies so that each of them contributes, independently to the others, to its own rate-limit "bucket" (cf the [token bucket](https://en.wikipedia.org/wiki/Token_bucket)).
In this case, `excludedIPs` should be set to match the list of `X-Forwarded-For IPs` that are to be excluded, in order to find the actual clientIP.

Example to use each IP as a distinct source:

| X-Forwarded-For                | excludedIPs           | clientIP     |
|--------------------------------|-----------------------|--------------|
| <a id="10-0-0-111-0-0-112-0-0-1" href="#10-0-0-111-0-0-112-0-0-1" title="#10-0-0-111-0-0-112-0-0-1">`"10.0.0.1,11.0.0.1,12.0.0.1"`</a> | `"11.0.0.1,12.0.0.1"` | `"10.0.0.1"` |
| <a id="10-0-0-211-0-0-112-0-0-1" href="#10-0-0-211-0-0-112-0-0-1" title="#10-0-0-211-0-0-112-0-0-1">`"10.0.0.2,11.0.0.1,12.0.0.1"`</a> | `"11.0.0.1,12.0.0.1"` | `"10.0.0.2"` |

2. Group together a set of IPs (also behind a common set of reverse-proxies) so that they are considered the same source, and all contribute to the same rate-limit bucket.

Example to group IPs together as same source:

|  X-Forwarded-For               |  excludedIPs | clientIP     |
|--------------------------------|--------------|--------------|
| <a id="10-0-0-111-0-0-112-0-0-1-2" href="#10-0-0-111-0-0-112-0-0-1-2" title="#10-0-0-111-0-0-112-0-0-1-2">`"10.0.0.1,11.0.0.1,12.0.0.1"`</a> | `"12.0.0.1"` | `"11.0.0.1"` |
| <a id="10-0-0-211-0-0-112-0-0-1-2" href="#10-0-0-211-0-0-112-0-0-1-2" title="#10-0-0-211-0-0-112-0-0-1-2">`"10.0.0.2,11.0.0.1,12.0.0.1"`</a> | `"12.0.0.1"` | `"11.0.0.1"` |
| <a id="10-0-0-311-0-0-112-0-0-1" href="#10-0-0-311-0-0-112-0-0-1" title="#10-0-0-311-0-0-112-0-0-1">`"10.0.0.3,11.0.0.1,12.0.0.1"`</a> | `"12.0.0.1"` | `"11.0.0.1"` |

### store

A Distributed Rate Limit middleware uses a persistent KV storage to store data.

Refer to the [redis options](#configuration-options) to configure the Redis connection.

Connection parameters to your [Redis](https://redis.io/ "Link to website of Redis") server are attached to your Middleware deployment.

The following Redis modes are supported:

- Single instance mode
- [Redis Cluster](https://redis.io/docs/management/scaling "Link to official Redis documentation about Redis Cluster mode")
- [Redis Sentinel](https://redis.io/docs/management/sentinel "Link to official Redis documentation about Redis Sentinel mode")

For more information about Redis, we recommend the [official Redis documentation](https://redis.io/docs/ "Link to official Redis documentation").

!!! info

    If you use Redis in single instance mode or Redis Sentinel, you can configure the `database` field.
    This value won't be taken into account if you use Redis Cluster (only database `0` is available).

    In this case, a warning is displayed, and the value is ignored.
