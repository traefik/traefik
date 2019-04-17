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

```toml tab="File"
# Enable gzip compression
[http.middlewares]
  [http.middlewares.test-compress.Compress]
```

## Notes

Responses are compressed when:

* The response body is larger than `512` bytes.
* The `Accept-Encoding` request header contains `gzip`.
* The response is not already compressed, i.e. the `Content-Encoding` response header is not already set.
