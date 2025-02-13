---
title: "Traefik Proxy Middleware Overview"
description: "There are several available middleware in Traefik Proxy used to modify requests or headers, take charge of redirections, add authentication, and so on."
---

# HTTP Middleware Overview

Attached to the routers, pieces of middleware are a means of tweaking the requests before they are sent to your service (or before the answer from the services are sent to the clients).

There are several available middlewares in Traefik, some can modify the request, the headers, some are in charge of redirections, some add authentication, and so on.

Middlewares that use the same protocol can be combined into chains to fit every scenario.

!!! warning "Provider Namespace"

    Be aware of the concept of Providers Namespace described in the [Configuration Discovery](../../../install-configuration/providers/overview.md#provider-namespace) section.
    It also applies to Middlewares.

## Available HTTP Middlewares

| Middleware                                | Purpose                                           | Area                        |
|-------------------------------------------|---------------------------------------------------|-----------------------------|
| [AddPrefix](addprefix.md)                 | Adds a Path Prefix                                | Path Modifier               |
| [BasicAuth](basicauth.md)                 | Adds Basic Authentication                         | Security, Authentication    |
| [Buffering](buffering.md)                 | Buffers the request/response                      | Request Lifecycle           |
| [Chain](chain.md)                         | Combines multiple pieces of middleware            | Misc                        |
| [CircuitBreaker](circuitbreaker.md)       | Prevents calling unhealthy services               | Request Lifecycle           |
| [Compress](compress.md)                   | Compresses the response                           | Content Modifier            |
| [ContentType](contenttype.md)             | Handles Content-Type auto-detection               | Misc                        |
| [DigestAuth](digestauth.md)               | Adds Digest Authentication                        | Security, Authentication    |
| [Errors](errorpages.md)                   | Defines custom error pages                        | Request Lifecycle           |
| [ForwardAuth](forwardauth.md)             | Delegates Authentication                          | Security, Authentication    |
| [GrpcWeb](grpcweb.md)                     | Converts gRPC Web requests to HTTP/2 gRPC requests.                           | Request                   |
| [Headers](headers.md)                     | Adds / Updates headers                            | Security                    |
| [IPAllowList](ipallowlist.md)             | Limits the allowed client IPs                     | Security, Request lifecycle |
| [InFlightReq](inflightreq.md)             | Limits the number of simultaneous connections     | Security, Request lifecycle |
| [PassTLSClientCert](passtlsclientcert.md) | Adds Client Certificates in a Header              | Security                    |
| [RateLimit](ratelimit.md)                 | Limits the call frequency                         | Security, Request lifecycle |
| [RedirectScheme](redirectscheme.md)       | Redirects based on scheme                         | Request lifecycle           |
| [RedirectRegex](redirectregex.md)         | Redirects based on regex                          | Request lifecycle           |
| [ReplacePath](replacepath.md)             | Changes the path of the request                   | Path Modifier               |
| [ReplacePathRegex](replacepathregex.md)   | Changes the path of the request                   | Path Modifier               |
| [Retry](retry.md)                         | Automatically retries in case of error            | Request lifecycle           |
| [StripPrefix](stripprefix.md)             | Changes the path of the request                   | Path Modifier               |
| [StripPrefixRegex](stripprefixregex.md)   | Changes the path of the request                   | Path Modifier               |

## Community Middlewares

Please take a look at the community-contributed plugins in the [plugin catalog](https://plugins.traefik.io/plugins).

{!traefik-for-business-applications.md!}
