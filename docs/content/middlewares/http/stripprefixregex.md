---
title: "Traefik StripPrefixRegex Documentation"
description: "In Traefik Proxy's HTTP middleware, StripPrefixRegex removes prefixes from paths before forwarding requests, using regex. Read the technical documentation."
---

# StripPrefixRegex

Removing Prefixes From the Path Before Forwarding the Request (Using a Regex)
{: .subtitle }

Remove the matching prefixes from the URL path.

## Configuration Examples

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-stripprefixregex.stripprefixregex.regex=/foo/[a-z0-9]+/[0-9]+/"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-stripprefixregex
spec:
  stripPrefixRegex:
    regex:
      - "/foo/[a-z0-9]+/[0-9]+/"
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-stripprefixregex.stripprefixregex.regex=/foo/[a-z0-9]+/[0-9]+/"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-stripprefixregex:
      stripPrefixRegex:
        regex:
          - "/foo/[a-z0-9]+/[0-9]+/"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-stripprefixregex.stripPrefixRegex]
    regex = ["/foo/[a-z0-9]+/[0-9]+/"]
```

## Configuration Options

### General

The StripPrefixRegex middleware strips the matching path prefix and stores it in a `X-Forwarded-Prefix` header.

!!! tip

    Use a `stripPrefixRegex` middleware if your backend listens on the root path (`/`) but should be exposed on a specific prefix.

### `regex`

The `regex` option is the regular expression to match the path prefix from the request URL.

For instance, `/products` also matches `/products/shoes` and `/products/shirts`.

If your backend is serving assets (e.g., images or JavaScript files), it can use the `X-Forwarded-Prefix` header to properly construct relative URLs.
Using the previous example, the backend should return `/products/shoes/image.png` (and not `/images.png`, which Traefik would likely not be able to associate with the same backend).

!!! tip

    Regular expressions and replacements can be tested using online tools such as [Go Playground](https://play.golang.org/p/mWU9p-wk2ru) or the [Regex101](https://regex101.com/r/58sIgx/2).

    When defining a regular expression within YAML, any escaped character needs to be escaped twice: `example\.com` needs to be written as `example\\.com`.
