# ReplaceQueryRegex

Updating the Query Before Forwarding the Request (Using a Regex)
{: .subtitle }

<!--
TODO: add schema
-->

The ReplaceRegex replace a query from a url to another with regex matching and replacement.

## Configuration Examples

```yaml tab="Docker"
# Replace query with regex
labels:
- "traefik.http.middlewares.test-replacequeryregex.replacequeryregex.regex=(.*)"
- "traefik.http.middlewares.test-replacequeryregex.replacequeryregex.replacement=$${1}&bar=foo"
```

```yaml tab="Kubernetes"
# Replace query with regex
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-replacequeryregex
spec:
  replaceQueryRegex:
    regex: (.*)
    replacement: ${1}&bar=foo
```

```json tab="Marathon"
# Replace query with regex
"labels": {
  "traefik.http.middlewares.test-replacequeryregex.replacequeryregex.regex": "(.*)",
  "traefik.http.middlewares.test-replacequeryregex.replacequeryregex.replacement": "${1}&bar=foo"
}
```

```yaml tab="Rancher"
# Replace query with regex
labels:
- "traefik.http.middlewares.test-replacequeryregex.replacequeryregex.regex=(.*)"
- "traefik.http.middlewares.test-replacequeryregex.replacequeryregex.replacement=${1}&bar=foo"
```

```toml tab="File (TOML)"
# Replace query with regex
[http.middlewares]
  [http.middlewares.test-replacequeryregex.replaceQueryRegex]
    regex = "(.*)"
    replacement = "${1}&bar=foo"
```

```yaml tab="File (YAML)"
# Replace query with regex
http:
  middlewares:
    test-replacequeryregex:
      replaceQueryRegex:
        regex: (.*)
        replacement: '${1}&bar=foo'
```

## Configuration Options

### General

The ReplaceQueryRegex middleware will:

- replace the matching query by the specified one.

### `regex`

The `regex` option is the regular expression to match and capture the query from the request URL.

!!! warning

    Care should be taken when defining replacement expand variables: `$1x` is equivalent to `${1x}`, not `${1}x` (see [Regexp.Expand](https://golang.org/pkg/regexp/#Regexp.Expand)), so use `${1}` syntax.

!!! info

    Regular expressions and replacements can be tested using online tools such as [Go Playground](https://play.golang.org/p/mWU9p-wk2ru) or the [Regex101](https://regex101.com/r/58sIgx/2).
    
!!! info

    This middleware only matches and modifies the query parameters of the request.
    You cannot match a non-existent query.
    If you want to ensure or add a query, use ReplacePathQueryRegex instead, as it can add queries to requests that don't have them.
    
### `replacement`

The `replacement` option defines how to modify the path and query to have the new target query.
