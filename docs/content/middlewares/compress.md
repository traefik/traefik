# Compress

Compressing the Response before Sending it to the Client
{: .subtitle }

![Compress](../assets/img/middleware/compress.png)

The Compress middleware enables the gzip compression. 

## Configuration Examples

```yaml tab="Docker"
# Enable gzip compression
labels:
- "traefik.http.middlewares.test-compress.compress=true"
```

```yaml tab="Kubernetes"
# Enable gzip compression
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-compress
spec:
  compress: {}
```

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.test-compress.compress": "true"
}
```

```yaml tab="Rancher"
# Enable gzip compression
labels:
- "traefik.http.middlewares.test-compress.compress=true"
```

```toml tab="File (TOML)"
# Enable gzip compression
[http.middlewares]
  [http.middlewares.test-compress.compress]
```

```yaml tab="File (YAML)"
# Enable gzip compression
http:
  middlewares:
    test-compress:
      compress: {}
```

!!! info
    
    Responses are compressed when:
    
    * The response body is larger than `1400` bytes.
    * The `Accept-Encoding` request header contains `gzip`.
    * The response is not already compressed, i.e. the `Content-Encoding` response header is not already set.
