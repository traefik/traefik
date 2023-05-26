---
title: "Traefik Compress Documentation"
description: "Traefik Proxy's HTTP middleware lets you compress responses before sending them to the client. Read the technical documentation."
---

# Compress

Compress Allows Compressing Responses before Sending them to the Client
{: .subtitle }

![Compress](../../assets/img/middleware/compress.png)

The Compress middleware supports gzip and Brotli compression.
The activation of compression, and the compression method choice rely (among other things) on the request's `Accept-Encoding` header.

## Configuration Examples

```yaml tab="Docker & Swarm"
# Enable compression
labels:
  - "traefik.http.middlewares.test-compress.compress=true"
```

```yaml tab="Kubernetes"
# Enable compression
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-compress
spec:
  compress: {}
```

```yaml tab="Consul Catalog"
# Enable compression
- "traefik.http.middlewares.test-compress.compress=true"
```

```yaml tab="File (YAML)"
# Enable compression
http:
  middlewares:
    test-compress:
      compress: {}
```

```toml tab="File (TOML)"
# Enable compression
[http.middlewares]
  [http.middlewares.test-compress.compress]
```

!!! info

    Responses are compressed when the following criteria are all met:

    * The `Accept-Encoding` request header contains `gzip`, `*`, and/or `br` with or without [quality values](https://developer.mozilla.org/en-US/docs/Glossary/Quality_values).
    If the `Accept-Encoding` request header is absent, it is meant as br compression is requested.
    If it is present, but its value is the empty string, then compression is disabled.
    * The response is not already compressed, i.e. the `Content-Encoding` response header is not already set.
    * The response`Content-Type` header is not one among the [excludedContentTypes options](#excludedcontenttypes).
    * The response body is larger than the [configured minimum amount of bytes](#minresponsebodybytes) (default is `1024`).

## Configuration Options

### `excludedContentTypes`

_Optional, Default=""_ 

`excludedContentTypes` specifies a list of content types to compare the `Content-Type` header of the incoming requests and responses before compressing.

The responses with content types defined in `excludedContentTypes` are not compressed.

Content types are compared in a case-insensitive, whitespace-ignored manner.

!!! info "In the case of gzip"

    If the `Content-Type` header is not defined, or empty, the compress middleware will automatically [detect](https://mimesniff.spec.whatwg.org/) a content type.
    It will also set the `Content-Type` header according to the detected MIME type.

!!! info "gRPC"

    Note that `application/grpc` is never compressed.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-compress.compress.excludedcontenttypes=text/event-stream"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-compress
spec:
  compress:
    excludedContentTypes:
      - text/event-stream
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-compress.compress.excludedcontenttypes=text/event-stream"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-compress:
      compress:
        excludedContentTypes:
          - text/event-stream
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-compress.compress]
    excludedContentTypes = ["text/event-stream"]
```

### `minResponseBodyBytes`

_Optional, Default=1024_

`minResponseBodyBytes` specifies the minimum amount of bytes a response body must have to be compressed.

Responses smaller than the specified values will not be compressed.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-compress.compress.minresponsebodybytes=1200"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-compress
spec:
  compress:
    minResponseBodyBytes: 1200
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-compress.compress.minresponsebodybytes=1200"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-compress:
      compress:
        minResponseBodyBytes: 1200
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-compress.compress]
    minResponseBodyBytes = 1200
```
