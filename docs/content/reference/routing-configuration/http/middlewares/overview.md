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
| <a id="AddPrefix" href="#AddPrefix" title="#AddPrefix">[AddPrefix](addprefix.md)</a> | Adds a Path Prefix                                | Path Modifier               |
| <a id="BasicAuth" href="#BasicAuth" title="#BasicAuth">[BasicAuth](basicauth.md)</a> | Adds Basic Authentication                         | Security, Authentication    |
| <a id="Buffering" href="#Buffering" title="#Buffering">[Buffering](buffering.md)</a> | Buffers the request/response                      | Request Lifecycle           |
| <a id="Chain" href="#Chain" title="#Chain">[Chain](chain.md)</a> | Combines multiple pieces of middleware            | Misc                        |
| <a id="CircuitBreaker" href="#CircuitBreaker" title="#CircuitBreaker">[CircuitBreaker](circuitbreaker.md)</a> | Prevents calling unhealthy services               | Request Lifecycle           |
| <a id="Compress" href="#Compress" title="#Compress">[Compress](compress.md)</a> | Compresses the response                           | Content Modifier            |
| <a id="ContentType" href="#ContentType" title="#ContentType">[ContentType](contenttype.md)</a> | Handles Content-Type auto-detection               | Misc                        |
| <a id="DigestAuth" href="#DigestAuth" title="#DigestAuth">[DigestAuth](digestauth.md)</a> | Adds Digest Authentication                        | Security, Authentication    |
| <a id="Errors" href="#Errors" title="#Errors">[Errors](errorpages.md)</a> | Defines custom error pages                        | Request Lifecycle           |
| <a id="ForwardAuth" href="#ForwardAuth" title="#ForwardAuth">[ForwardAuth](forwardauth.md)</a> | Delegates Authentication                          | Security, Authentication    |
| <a id="GrpcWeb" href="#GrpcWeb" title="#GrpcWeb">[GrpcWeb](grpcweb.md)</a> | Converts gRPC Web requests to HTTP/2 gRPC requests.                           | Request                   |
| <a id="Headers" href="#Headers" title="#Headers">[Headers](headers.md)</a> | Adds / Updates headers                            | Security                    |
| <a id="IPAllowList" href="#IPAllowList" title="#IPAllowList">[IPAllowList](ipallowlist.md)</a> | Limits the allowed client IPs                     | Security, Request lifecycle |
| <a id="InFlightReq" href="#InFlightReq" title="#InFlightReq">[InFlightReq](inflightreq.md)</a> | Limits the number of simultaneous connections     | Security, Request lifecycle |
| <a id="PassTLSClientCert" href="#PassTLSClientCert" title="#PassTLSClientCert">[PassTLSClientCert](passtlsclientcert.md)</a> | Adds Client Certificates in a Header              | Security                    |
| <a id="RateLimit" href="#RateLimit" title="#RateLimit">[RateLimit](ratelimit.md)</a> | Limits the call frequency                         | Security, Request lifecycle |
| <a id="RedirectScheme" href="#RedirectScheme" title="#RedirectScheme">[RedirectScheme](redirectscheme.md)</a> | Redirects based on scheme                         | Request lifecycle           |
| <a id="RedirectRegex" href="#RedirectRegex" title="#RedirectRegex">[RedirectRegex](redirectregex.md)</a> | Redirects based on regex                          | Request lifecycle           |
| <a id="ReplacePath" href="#ReplacePath" title="#ReplacePath">[ReplacePath](replacepath.md)</a> | Changes the path of the request                   | Path Modifier               |
| <a id="ReplacePathRegex" href="#ReplacePathRegex" title="#ReplacePathRegex">[ReplacePathRegex](replacepathregex.md)</a> | Changes the path of the request                   | Path Modifier               |
| <a id="Retry" href="#Retry" title="#Retry">[Retry](retry.md)</a> | Automatically retries in case of error            | Request lifecycle           |
| <a id="StripPrefix" href="#StripPrefix" title="#StripPrefix">[StripPrefix](stripprefix.md)</a> | Changes the path of the request                   | Path Modifier               |
| <a id="StripPrefixRegex" href="#StripPrefixRegex" title="#StripPrefixRegex">[StripPrefixRegex](stripprefixregex.md)</a> | Changes the path of the request                   | Path Modifier               |

## Community Middlewares

Please take a look at the community-contributed plugins in the [plugin catalog](https://plugins.traefik.io/plugins).

{!traefik-for-business-applications.md!}
