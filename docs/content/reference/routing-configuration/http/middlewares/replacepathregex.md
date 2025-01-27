---
title: "Traefik ReplacePathRegex Documentation"
description: "In Traefik Proxy's HTTP middleware, ReplacePathRegex updates paths before forwarding requests, using a regex. Read the technical documentation."
---

The `replacePathRegex` middleware will:

- Replace the matching path with the specified one.
- Store the original path in an `X-Replaced-Path` header

## Configuration Examples

```yaml tab="File (YAML)"
# Replace path with regex
http:
  middlewares:
    test-replacepathregex:
      replacePathRegex:
        regex: "^/foo/(.*)"
        replacement: "/bar/$1"
```

```toml tab="File (TOML)"
# Replace path with regex
[http.middlewares]
  [http.middlewares.test-replacepathregex.replacePathRegex]
    regex = "^/foo/(.*)"
    replacement = "/bar/$1"
```

```yaml tab="Kubernetes"
# Replace path with regex
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-replacepathregex
spec:
  replacePathRegex:
    regex: "^/foo/(.*)"
    replacement: "/bar/$1"
```

```yaml tab="Docker & Swarm"
# Replace path with regex
labels:
  - "traefik.http.middlewares.test-replacepathregex.replacepathregex.regex=^/foo/(.*)"
  - "traefik.http.middlewares.test-replacepathregex.replacepathregex.replacement=/bar/$$1"
```

```yaml tab="Consul Catalog"
# Replace path with regex
- "traefik.http.middlewares.test-replacepathregex.replacepathregex.regex=^/foo/(.*)"
- "traefik.http.middlewares.test-replacepathregex.replacepathregex.replacement=/bar/$1"
```

## Configuration Options

| Field                        | Description      | Default | Required |
|:-----------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| `regex` | Regular expression to match and capture the path from the request URL. | | Yes |
| `replacement` | Replacement path format, which can include captured variables.<br /> `$1x` is equivalent to `${1x}`, not `${1}x` (see [Regexp.Expand](https://golang.org/pkg/regexp/#Regexp.Expand)), so use `${1}` syntax. | | No 

!!! tip

    Regular expressions and replacements can be tested using online tools such as [Go Playground](https://play.golang.org/p/mWU9p-wk2ru) or the [Regex101](https://regex101.com/r/58sIgx/2).

    When defining a regular expression within YAML, any escaped character needs to be escaped twice: `example\.com` needs to be written as `example\\.com`.
