
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
    
    For compatibility reason, the default behaviour on a router (without this middleware) , is to auto-detect the `Content-Type`.

## Configuration Options

### `autoDetect`

`autoDetect` specifies if we want to auto-detect `Content-Type`.
