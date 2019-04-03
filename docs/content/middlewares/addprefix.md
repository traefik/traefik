# Add Prefix

Prefixing the Path 
{: .subtitle }

![AddPrefix](../assets/img/middleware/addprefix.png) 

The AddPrefix middleware updates the URL Path of the request before forwarding it.

## Configuration Examples

```yaml tab="Docker"
# Prefixing with /foo
labels:
- "traefik.http.middlewares.add-bar.addprefix.prefix=/foo"
```

```yaml tab="Kubernetes"
# Prefixing with /foo
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: addprefix
spec:
  addPrefix:
    prefix: /foo
```

```toml tab="File"
# Prefixing with /foo
[http.middlewares]
  [http.middlewares.add-foo.AddPrefix]
     prefix = "/foo"
```

## Configuration Options

### `prefix`

`prefix` is the string to add before the current path in the requested URL. It should include the leading slash (`/`).
