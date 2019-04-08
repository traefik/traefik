# RedirectScheme

Redirecting the Client to a Different Scheme/Port
{: .subtitle }

`TODO: add schema`

RegexRedirect redirect request from a scheme to another.

## Configuration Examples

```yaml tab="Docker"
# Redirect to https
labels:
- "traefik.http.middlewares.test-redirectscheme.redirectscheme.scheme=https"
```

```yaml tab="Kubernetes"
# Redirect to https
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-redirectscheme
spec:
  redirectScheme:
    scheme: https
```

```yaml tab="Rancher"
# Redirect to https
labels:
- "traefik.http.middlewares.test-redirectscheme.redirectscheme.scheme=https"
```

```toml tab="File"
# Redirect to https
[http.middlewares]
  [http.middlewares.test-redirectscheme.redirectscheme]
    scheme = "https"
```

## Configuration Options

### `permanent`

Set the `permanent` option to `true` to apply a permanent redirection.

### `scheme`

The `scheme` option defines the scheme of the new url.

### `port`

The `port` option defines the port of the new url.
