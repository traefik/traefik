# StripPrefix

Removing Prefixes From the Path Before Forwarding the Request (Using a Regex)
{: .subtitle }


`TODO: add schema`

Remove the matching prefixes from the URL path.

# Configuration Examples

```yaml tab="Docker"
# Replace the path by /foo
labels:
- "traefik.http.middlewares.test-stripprefixregex.stripprefixregex.regex=^/foo/(.*)",
```

```yaml tab="Kubernetes"
# Replace the path by /foo
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-stripprefixregex
spec:
  StripPrefixRegex:
    regex: "^/foo/(.*)"
```

```toml tab="File"
# Replace the path by /foo
[http.middlewares]
  [http.middlewares.test-stripprefixregex.StripPrefixRegex]
     regex: "^/foo/(.*)"
```

## Configuration Options

### General

The StripPrefixRegex middleware will:

* strip the matching path prefix.
* store the matching path prefix in a `X-Forwarded-Prefix` header.

!!! tip
    
    Use a `StripPrefixRegex` middleware if your backend listens on the root path (`/`) but should be routeable on a specific prefix.

### `regex`

The `regex` option is the regular expression to match the path prefix from the request URL.


!!! tip

    Regular expressions can be tested using online tools such as [Go Playground](https://play.golang.org/p/mWU9p-wk2ru) or the [Regex101](https://regex101.com/r/58sIgx/2).

For instance, `/products` would match `/products` but also `/products/shoes` and `/products/shirts`.  
Since the path is stripped prior to forwarding, your backend is expected to listen on `/`.  
If your backend is serving assets (e.g., images or Javascript files), chances are it must return properly constructed relative URLs.  
Continuing on the example, the backend should return `/products/shoes/image.png` (and not `/images.png` which Traefik would likely not be able to associate with the same backend).  
The `X-Forwarded-Prefix` header can be queried to build such URLs dynamically.
