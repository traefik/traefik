---
title: "Traefik Compress Documentation"
description: "Traefik Proxy's HTTP middleware lets you compress responses before sending them to the client. Read the technical documentation."
---

# Compress

Compress Responses before Sending them to the Client
{: .subtitle }

![Compress](../../assets/img/middleware/compress.png)

The Compress middleware uses gzip compression.

## Configuration Examples

```yaml tab="Docker"
# Enable gzip compression
labels:
  - "traefik.http.middlewares.test-compress.compress=true"
```

```yaml tab="Kubernetes"
# Enable gzip compression
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-compress
spec:
  compress: {}
```

```yaml tab="Consul Catalog"
# Enable gzip compression
- "traefik.http.middlewares.test-compress.compress=true"
```

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.test-compress.compress": "true"
}
```

```yaml tab="Rancher"
# Enable gzip compression
labels:
  - "traefik.http.middlewares.test-compress.compress=true"
```

```yaml tab="File (YAML)"
# Enable gzip compression
http:
  middlewares:
    test-compress:
      compress: {}
```

```toml tab="File (TOML)"
# Enable gzip compression
[http.middlewares]
  [http.middlewares.test-compress.compress]
```

!!! info

    Responses are compressed when the following criteria are all met:

    * The response body is larger than the configured minimum amount of bytes (default is `1024`).
    * The `Accept-Encoding` request header contains `gzip`.
    * The response is not already compressed, i.e. the `Content-Encoding` response header is not already set.

    If the `Content-Type` header is not defined, or empty, the compress middleware will automatically [detect](https://mimesniff.spec.whatwg.org/) a content type.
    It will also set the `Content-Type` header according to the detected MIME type.

## Configuration Options

### `excludedContentTypes`

`excludedContentTypes` specifies a list of content types to compare the `Content-Type` header of the incoming requests and responses before compressing.

The responses with content types defined in `excludedContentTypes` are not compressed.

Content types are compared in a case-insensitive, whitespace-ignored manner.

```yaml tab="Docker"
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

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.test-compress.compress.excludedcontenttypes": "text/event-stream"
}
```

```yaml tab="Rancher"
labels:
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

`minResponseBodyBytes` specifies the minimum amount of bytes a response body must have to be compressed.

The default value is `1024`, which should be a reasonable value for most cases.

Responses smaller than the specified values will not be compressed.

```yaml tab="Docker"
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

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.test-compress.compress.minresponsebodybytes": 1200
}
```

```yaml tab="Rancher"
labels:
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
