---
title: "Traefik RedirectRegex Documentation"
description: "In Traefik Proxy's HTTP middleware, RedirectRegex redirecting clients to different locations. Read the technical documentation."
---

The `RedirectRegex` redirects a request using regex matching and replacement.

## Configuration Examples

```yaml tab="Structured (YAML)"
# Redirect with domain replacement
http:
  middlewares:
    test-redirectregex:
      redirectRegex:
        regex: "^http://localhost/(.*)"
        replacement: "http://mydomain/${1}"
```

```toml tab="Structured (TOML)"
# Redirect with domain replacement
[http.middlewares]
  [http.middlewares.test-redirectregex.redirectRegex]
    regex = "^http://localhost/(.*)"
    replacement = "http://mydomain/${1}"
```

```yaml tab="Labels"
# Redirect with domain replacement
# Note: all dollar signs need to be doubled for escaping.
labels:
  - "traefik.http.middlewares.test-redirectregex.redirectregex.regex=^http://localhost/(.*)"
  - "traefik.http.middlewares.test-redirectregex.redirectregex.replacement=http://mydomain/$${1}"
```

```json tab="Tags"
// Redirect with domain replacement
// Note: all dollar signs need to be doubled for escaping.
{
  // ...
  "Tags" : [
    "traefik.http.middlewares.test-redirectregex.redirectregex.regex=^http://localhost/(.*)"
    "traefik.http.middlewares.test-redirectregex.redirectregex.replacement=http://mydomain/$${1}"
  ]
}
```

```yaml tab="Kubernetes"
# Redirect with domain replacement
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-redirectregex
spec:
  redirectRegex:
    regex: ^http://localhost/(.*)
    replacement: http://mydomain/${1}
```

## Configuration Options

<!-- markdownlint-disable MD013 -->

| Field                        | Description                                                                                                                                                                                                | Default | Required |
|:-----------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| `regex` | The `regex` option is the regular expression to match and capture elements from the request URL.| "" | Yes |
| `permanent` | Enable a permanent redirection. | false | No |
| `replacement` | The `replacement` option defines how to modify the URL to have the new target URL..<br /> `$1x` is equivalent to `${1x}`, not `${1}x` (see [Regexp.Expand](https://golang.org/pkg/regexp/#Regexp.Expand)), so use `${1}` syntax. | "" | No |

### `regex`

The `regex` option is the regular expression to match and capture elements from the request URL.

!!! tip

    Regular expressions and replacements can be tested using online tools such as [Go Playground](https://play.golang.org/p/mWU9p-wk2ru) or the [Regex101](https://regex101.com/r/58sIgx/2).

    When defining a regular expression within YAML, any escaped character needs to be escaped twice: `example\.com` needs to be written as `example\\.com`.
    
### `replacement`

The `replacement` option defines how to modify the URL to have the new target URL.

!!! warning

    Care should be taken when defining replacement expand variables: `$1x` is equivalent to `${1x}`, not `${1}x` (see [Regexp.Expand](https://golang.org/pkg/regexp/#Regexp.Expand)), so use `${1}` syntax.

{!traefik-for-business-applications.md!}
