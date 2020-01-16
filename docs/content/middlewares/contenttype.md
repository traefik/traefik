
# ContentType

Handling ContentType auto-detection
{: .subtitle }

The Content-Type middleware handles enabling/disabling auto-detection of `Content-Type`. 

## Configuration Examples

```yaml tab="Docker"
# Disable auto-detection
labels:
  - "traefik.http.middlewares.autodetect.contenttype.autodetect=false"
```

```yaml tab="Kubernetes"
# Disable auto-detection
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: autodetect
spec:
  contentType:
    autoDetect: false
```

```yaml tab="Consul Catalog"
# Disable auto-detection
- "traefik.http.middlewares.autodetect.contenttype.autodetect=false"
```

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.autodetect.contenttype.autodetect": "false"
}
```

```yaml tab="Rancher"
# Disable auto-detection
labels:
  - "traefik.http.middlewares.autodetect.contenttype.autodetect=false"
```

```toml tab="File (TOML)"
# Disable auto-detection
[http.middlewares]
  [http.middlewares.autodetect.contentType]
     autoDetect=false
```

```yaml tab="File (YAML)"
# Disable auto-detection
http:
  middlewares:
    autodetect:
      contentType:
        autoDetect: false
```

!!! info
    
    For compatibility reason, the default behaviour on a router (without this middleware) , is to automatically set the `Content-Type` header (if it was unset), with a value derived from the contents of the response. Therefore, given the default value of the `autoDetect` option, simply enabling this middleware for a router potentially changes its behaviour.

## Configuration Options

### `autoDetect`

`autoDetect` specifies whether to let the `Content-Type` header, if it hasn't
been set by the backend, be automatically set to a value derived from the
contents of the response. As a proxy, the default behaviour should be to leave
the header alone, regardless of what the backend did with it. Unfortunately, the
historic default was to always auto-detect and set the header if it was nil, and
it is going to be kept that way in order to support users currently relying on
it. This middleware exists to enable the correct behaviour until at least the
default one can be changed in a future version.
