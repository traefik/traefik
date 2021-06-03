# FAQ

## Why is Traefik Answering XXX HTTP Response Status Code ?

Traefik is a dynamic reverse proxy,
and while the documentation often demonstrate configuration options through file examples,
the core feature of Traefik is its ability to be provided with configuration dynamically.

The file provider allows to provide [dynamic configuration](../configuration-overview/#the-dynamic-configuration) through a file or directory.
It has to be distinguished from the [static configuration](../configuration-overview/#the-static-configuration),
that can also be provided by a file.

For HTTP configurations, the EntryPoint concept matches the Transport Layer (TCP),
while the Router concept matches the Presentation (TLS) and Application layers (HTTP).

An EntryPoint is not tied to the protocol HTTP or HTTPS,
as one can attach a TLS router or a non-TLS router to the same EntryPoint.

***Therefore, in this dynamic context, when an EntryPoint is configured, 
Traefik is handling traffic on the EP port without the guaranty to match any configured route.***

### 404 Not found

Traefik returns 404 response code in the following situations:

- A request reaching an EntryPoint that has no Routers
- An HTTP request reaching an EntryPoint that has no HTTP Router
- An HTTPS request reaching an EntryPoint that has no HTTPS Router
- A request reaching an EntryPoint that has HTTP/HTTPS Routers that cannot be matched

From Traefik point of view, 
every times a request cannot be matched with a router the response code is a 404 Not found.

A 503 Service Unavailable response code would not be adapted in this situation 
because Traefik is not able to state that the lack of a matching router for a request is only temporary.
Traefik routing configuration is dynamic and aggregated from different providers,
hence it's not possible to assume at any moment that a specific route should be handled or not.

This behavior is consistent with the [rfc7231](https://datatracker.ietf.org/doc/html/rfc7231#section-6.6.4):
> The server is currently unable to handle the request due to a
> temporary overloading or maintenance of the server. The implication
> is that this is a temporary condition which will be alleviated after
> some delay. If known, the length of the delay MAY be indicated in a
> Retry-After header. If no Retry-After is given, the client SHOULD
> handle the response as it would for a 500 response.
>
>      Note: The existence of the 503 status code does not imply that a
>      server must use it when becoming overloaded. Some servers may wish
>      to simply refuse the connection.

### 502 Bad Gateway

Traefik returns 502 response code when an error happens while contacting the upstream service.

### 503 Service Unavailable

Traefik returns 503 response code when a Router has been matched 
but then there is no server ready to handle the request.

This situation is expected when a service has been explicitly configured without servers,
or when a service has healthcheck enabled and all servers are unhealthy.

### XXX Instead Of 404

Sometimes, the 404 response code doesn't play well with other parties.

In that situation, one could want that Traefik always replies with a 503 response code,
in place of a 404 response code.

To achieve this behavior, a simple catchall router, 
with the lowest possible priority and routing to a service without servers,
can handle all the requests when no other router has been matched.

The example below is only a file provider version (yaml) of the configuration.

**Static configuration (traefik.yaml):**

```yaml
entrypoints:
  web:
    address: :80

providers:
  file:
    filename: dynamic.yaml
```

**Dynamic configuration (dynamic.yaml):**

```yaml
http:
  routers:
    catchall:
      # attached only to web entryPoint
      entryPoints:
        - "web"
      # catchall rule
      rule: "PathPrefix(`/`)"
      service: unavailable
      # lowest possible priority
      # evaluated when no other router is matched
      priority: 1

  services:
    # Service that will always answer a 503 Service Unavailable response
    unavailable:
      loadBalancer:
        servers: {}
```

***Dedicated service***

To gain control over the response code and message,
a custom service would be associated with the catchall router.
