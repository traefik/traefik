---
title: "Traefik Compress Documentation"
description: "Traefik Proxy's HTTP middleware lets you compress responses before sending them to the client. Read the technical documentation."
---

The `compress` middleware compresses response. It supports Gzip, Brotli and Zstandard compression

## Configuration Examples

```yaml tab="Structured (YAML)"
# Enable compression
http:
  middlewares:
    test-compress:
      compress: {}
```

```toml tab="Structured (TOML)"
# Enable compression
[http.middlewares]
  [http.middlewares.test-compress.compress]
```

```yaml tab="Labels"
# Enable compression
labels:
  - "traefik.http.middlewares.test-compress.compress=true"
```

```json tab="Tags"
// Enable compression
{
  //...
  "Tags": [
    "traefik.http.middlewares.test-compress.compress=true"
  ]
}
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

## Configuration Options

| Field                        | Description                                                                                                                                                                                                | Default | Required |
|:-----------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
|`excludedContentTypes` | List of content types to compare the `Content-Type` header of the incoming requests and responses before compressing. <br /> The responses with content types defined in `excludedContentTypes` are not compressed. <br /> Content types are compared in a case-insensitive, whitespace-ignored manner. <br /> **The `excludedContentTypes` and `includedContentTypes` options are mutually exclusive.** | "" | No |
|`defaultEncoding` | specifies the default encoding if the `Accept-Encoding` header is not in the request or contains a wildcard (`*`). | "" | No |
|`encodings` | Specifies the list of supported compression encodings. At least one encoding value must be specified, and valid entries are `zstd` (Zstandard), `br` (Brotli), and `gzip` (Gzip). The order of the list also sets the priority, the top entry has the highest priority. | zstd, br, gzip | No |
| `includedContentTypes` | List of content types to compare the `Content-Type` header of the responses before compressing. <br /> The responses with content types defined in `includedContentTypes` are compressed. <br /> Content types are compared in a case-insensitive, whitespace-ignored manner.<br /> **The `excludedContentTypes` and `includedContentTypes` options are mutually exclusive.** | "" | No |
| `minResponseBodyBytes` | `Minimum amount of bytes a response body must have to be compressed. <br />Responses smaller than the specified values will **not** be compressed. | 1024 | No |

## Compression activation

The activation of compression, and the compression method choice rely (among other things) on the request's `Accept-Encoding` header.

Responses are compressed when the following criteria are all met:

- The `Accept-Encoding` request header contains `gzip`, `*`, and/or `br`, and/or `zstd` with or without [quality values](https://developer.mozilla.org/en-US/docs/Glossary/Quality_values).
If the `Accept-Encoding` request header is absent, the response won't be encoded.
If it is present, but its value is the empty string, then compression is turned off.
- The response is not already compressed, that is the `Content-Encoding` response header is not already set.
- The response`Content-Type` header is not one among the `excludedContentTypes` options, or is one among the `includedContentTypes` options.
- The response body is larger than the configured minimum amount of bytes(option `minResponseBodyBytes`) (default is `1024`).

## Empty Content-Type Header

If the `Content-Type` header is not defined, or empty, the compress middleware will automatically [detect](https://mimesniff.spec.whatwg.org/) a content type.
It will also set the `Content-Type` header according to the detected MIME type.

## GRPC application

Note that `application/grpc` is never compressed.
