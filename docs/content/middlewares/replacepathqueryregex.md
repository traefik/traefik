# ReplacePathQueryRegex

Updating the Path and Query Before Forwarding the Request (Using a Regex)
{: .subtitle }

`TODO: add schema`

The ReplaceRegex replace a path or query from a url to another with regex matching and replacement.

## Configuration Examples

```yaml tab="Docker"
# Replace path and query with regex
labels:
- "traefik.http.middlewares.test-replacepathqueryregex.replacepathqueryregex.regex=^/foo/(.*)"
- "traefik.http.middlewares.test-replacepathqueryregex.replacepathqueryregex.replacement=/bar?$1"
```

```yaml tab="Kubernetes"
# Replace path and query with regex
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-replacepathqueryregex
spec:
  replacePathQueryRegex:
    regex: ^/foo/(.*)
    replacement: /bar?$1
```

```json tab="Marathon"
# Replace path and query with regex
"labels": {
  "traefik.http.middlewares.test-replacepathqueryregex.replacepathqueryregex.regex": "^/foo/(.*)",
  "traefik.http.middlewares.test-replacepathqueryregex.replacepathqueryregex.replacement": "/bar?$1"
}
```

```yaml tab="Rancher"
# Replace path and query with regex
labels:
- "traefik.http.middlewares.test-replacepathqueryregex.replacepathqueryregex.regex=^/foo/(.*)"
- "traefik.http.middlewares.test-replacepathqueryregex.replacepathqueryregex.replacement=/bar?$1"
```

```toml tab="File"
# Redirect with domain replacement
[http.middlewares]
  [http.middlewares.test-replacepathqueryregex.replacePathQueryRegex]
    regex = "^/foo/(.*)"
    replacement = "/bar?$1"
```

## Configuration Options

### General

The ReplacePathQueryRegex middleware will:

- replace the matching path and query by the specified one.

### `regex`

The `Regex` option is the regular expression to match and capture the path and query from the request URL.

!!! warning

    Care should be taken when defining replacement expand variables: `$1x` is equivalent to `${1x}`, not `${1}x` (see [Regexp.Expand](https://golang.org/pkg/regexp/#Regexp.Expand)), so use `${1}` syntax.

!!! tip

    Regular expressions and replacements can be tested using online tools such as [Go Playground](https://play.golang.org/p/mWU9p-wk2ru) or the [Regex101](https://regex101.com/r/58sIgx/2).
    
### `replacement`

The `replacement` option defines how to modify the path and query to have the new target path and query.
