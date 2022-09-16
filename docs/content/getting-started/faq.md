---
title: "Traefik Getting Started FAQ"
description: "Check out our FAQ page for answers to commonly asked questions on getting started with Traefik Proxy. Read the technical documentation."
---

# FAQ

## Why is Traefik Answering `XXX` HTTP Response Status Code?

Traefik is a dynamic reverse proxy,
and while the documentation often demonstrates configuration options through file examples,
the core feature of Traefik is its dynamic configurability,
directly reacting to changes from providers over time.

Notably, a part of the configuration is [static](../configuration-overview/#the-static-configuration),
and can be provided by a file on startup, whereas various providers,
such as the file provider,
contribute dynamically all along the traefik instance lifetime to its [dynamic configuration](../configuration-overview/#the-dynamic-configuration) changes.

In addition, the configuration englobes concepts such as the EntryPoint which can be seen as a listener on the Transport Layer (TCP),
as apposed to the Router which is more about the Presentation (TLS) and Application layers (HTTP).
And there can be as many routers as one wishes for a given EntryPoint.

In other words, for a given Entrypoint,
at any given time the traffic seen is not bound to be just about one protocol.
It could be HTTP, or otherwise. Over TLS, or not.
Not to mention that dynamic configuration changes potentially make that kind of traffic vary over time.

Therefore, in this dynamic context,
the static configuration of an `entryPoint` does not give any hint whatsoever about how the traffic going through that `entryPoint` is going to be routed.
Or whether it's even going to be routed at all,
i.e. whether there is a Router matching the kind of traffic going through it.

### `404 Not found`

Traefik returns a `404` response code in the following situations:

- A request reaching an EntryPoint that has no Routers
- An HTTP request reaching an EntryPoint that has no HTTP Router
- An HTTPS request reaching an EntryPoint that has no HTTPS Router
- A request reaching an EntryPoint that has HTTP/HTTPS Routers that cannot be matched

From Traefik's point of view,
every time a request cannot be matched with a router the correct response code is a `404 Not found`.

In this situation, the response code is not a `503 Service Unavailable`
because Traefik is not able to confirm that the lack of a matching router for a request is only temporary.
Traefik's routing configuration is dynamic and aggregated from different providers,
hence it's not possible to assume at any moment that a specific route should be handled or not.

??? info "This behavior is consistent with rfc7231"

    ```txt
    The server is currently unable to handle the request due to a
    temporary overloading or maintenance of the server. The implication
    is that this is a temporary condition which will be alleviated after
    some delay. If known, the length of the delay MAY be indicated in a
    Retry-After header. If no Retry-After is given, the client SHOULD
    handle the response as it would for a 500 response.

        Note: The existence of the 503 status code does not imply that a
        server must use it when becoming overloaded. Some servers may wish
        to simply refuse the connection.
    ```

    Extract from [rfc7231#section-6.6.4](https://datatracker.ietf.org/doc/html/rfc7231#section-6.6.4).

### `502 Bad Gateway`

Traefik returns a `502` response code when an error happens while contacting the upstream service.

### `503 Service Unavailable`

Traefik returns a `503` response code when a Router has been matched
but there are no servers ready to handle the request.

This situation is encountered when a service has been explicitly configured without servers,
or when a service has healthcheck enabled and all servers are unhealthy.

### `XXX` Instead of `404`

Sometimes, the `404` response code doesn't play well with other parties or services (such as CDNs).

In these situations, you may want Traefik to always reply with a `503` response code,
instead of a `404` response code.

To achieve this behavior, a simple catchall router,
with the lowest possible priority and routing to a service without servers,
can handle all the requests when no other router has been matched.

The example below is a file provider only version (`yaml`) of what this configuration could look like:

```yaml tab="Static configuration"
# traefik.yml

entrypoints:
  web:
    address: :80

providers:
  file:
    filename: dynamic.yaml
```

```yaml tab="Dynamic configuration"
# dynamic.yaml

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

!!! info "Dedicated service"
    If there is a need for a response code other than a `503` and/or a custom message,
    the principle of the above example above (a catchall router) still stands,
    but the `unavailable` service should be adapted to fit such a need.

## Why Is My TLS Certificate Not Reloaded When Its Contents Change? 

With the file provider,
a configuration update is only triggered when one of the [watched](../providers/file.md#provider-configuration) configuration files is modified.

Which is why, when a certificate is defined by path,
and the actual contents of this certificate change,
a configuration update is _not_ triggered.

To take into account the new certificate contents, the update of the dynamic configuration must be forced.
One way to achieve that, is to trigger a file notification,
for example, by using the `touch` command on the configuration file.

## What Are the Forwarded Headers When Proxying HTTP Requests?

By default, the following headers are automatically added when proxying requests:

| Property                  | HTTP Header                |
|---------------------------|----------------------------|
| Client's IP               | X-Forwarded-For, X-Real-Ip |
| Host                      | X-Forwarded-Host           |
| Port                      | X-Forwarded-Port           |
| Protocol                  | X-Forwarded-Proto          |
| Proxy Server's Hostname   | X-Forwarded-Server         |

For more details,
please check out the [forwarded header](../routing/entrypoints.md#forwarded-headers) documentation.

## What does the "field not found" error mean?

```shell
error: field not found, node: -badField-
```

The "field not found" error occurs, when an unknown property is encountered in the dynamic or static configuration.

One easy way to check whether a configuration file is well-formed, is to validate it with:

- [JSON Schema of the static configuration](https://json.schemastore.org/traefik-v2.json)
- [JSON Schema of the dynamic configuration](https://json.schemastore.org/traefik-v2-file-provider.json)

## Why are some resources (routers, middlewares, services...) not created/applied?

As a common tip, if a resource is dropped/not created by Traefik after the dynamic configuration was evaluated,
one should look for an error in the logs.

If found, the error obviously confirms that something went wrong while creating the resource,
and the message should help in figuring out the mistake(s) in the configuration, and how to fix it.

When using the file provider,
one easy way to check if the dynamic configuration is well-formed is to validate it with the [JSON Schema of the dynamic configuration](https://json.schemastore.org/traefik-v2-file-provider.json).
