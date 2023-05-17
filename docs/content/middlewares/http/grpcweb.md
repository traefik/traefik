---
title: "Traefik GrpcWeb Documentation"
description: "In Traefik Proxy's HTTP middleware, GrpcWeb converts a gRPC Web requests to HTTP/2 gRPC requests. Read the technical documentation."
---

# GrpcWeb

Converting gRPC Web requests to HTTP/2 gRPC requests.
{: .subtitle }

The GrpcWeb middleware converts gRPC Web requests to HTTP/2 gRPC requests before forwarding them to the backends.

!!! tip

    Please note, that Traefik needs to communicate using gRPC with the backends (h2c or HTTP/2 over TLS).
    Check out the [gRPC](../../user-guides/grpc.md) user guide for more details.

## Configuration Examples

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-grpcweb.grpcweb.allowOrigins=*"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-grpcweb
spec:
  grpcWeb:
    allowOrigins:
      - "*"
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-grpcweb.grpcWeb.allowOrigins=*"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-grpcweb:
      grpcWeb:
        allowOrigins:
          - "*"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-grpcweb.grpcWeb]
    allowOrigins = ["*"]
```

## Configuration Options

### `allowOrigins`

The `allowOrigins` contains the list of allowed origins.
A wildcard origin `*` can also be configured to match all requests.

More information including how to use the settings can be found at:

- [Mozilla.org](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin)
- [w3](https://fetch.spec.whatwg.org/#http-access-control-allow-origin)
- [IETF](https://tools.ietf.org/html/rfc6454#section-7.1)
