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

| Middleware                                                                                                                               | Purpose                                           | Area                        |
|------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------|-----------------------------|
| <a id="opt-AddPrefix" href="#opt-AddPrefix" title="#opt-AddPrefix">[AddPrefix](addprefix.md)</a> | Adds a Path Prefix                                | Path Modifier               |
| <a id="opt-BasicAuth" href="#opt-BasicAuth" title="#opt-BasicAuth">[BasicAuth](basicauth.md)</a> | Adds Basic Authentication                         | Security, Authentication    |
| <a id="opt-Buffering" href="#opt-Buffering" title="#opt-Buffering">[Buffering](buffering.md)</a> | Buffers the request/response                      | Request Lifecycle           |
| <a id="opt-Chain" href="#opt-Chain" title="#opt-Chain">[Chain](chain.md)</a> | Combines multiple pieces of middleware            | Misc                        |
| <a id="opt-CircuitBreaker" href="#opt-CircuitBreaker" title="#opt-CircuitBreaker">[CircuitBreaker](circuitbreaker.md)</a> | Prevents calling unhealthy services               | Request Lifecycle           |
| <a id="opt-Compress" href="#opt-Compress" title="#opt-Compress">[Compress](compress.md)</a> | Compresses the response                           | Content Modifier            |
| <a id="opt-ContentType" href="#opt-ContentType" title="#opt-ContentType">[ContentType](contenttype.md)</a> | Handles Content-Type auto-detection               | Misc                        |
| <a id="opt-DigestAuth" href="#opt-DigestAuth" title="#opt-DigestAuth">[DigestAuth](digestauth.md)</a> | Adds Digest Authentication                        | Security, Authentication    |
| <a id="opt-EncodedCharacters" href="#opt-EncodedCharacters" title="#opt-EncodedCharacters">[EncodedCharacters](encodedcharacters.md)</a> | Defines allowed reserved encoded characters in the request path | Security, Request Lifecycle           |
| <a id="opt-Errors" href="#opt-Errors" title="#opt-Errors">[Errors](errorpages.md)</a> | Defines custom error pages                        | Request Lifecycle           |
| <a id="opt-ForwardAuth" href="#opt-ForwardAuth" title="#opt-ForwardAuth">[ForwardAuth](forwardauth.md)</a> | Delegates Authentication                          | Security, Authentication    |
| <a id="opt-GrpcWeb" href="#opt-GrpcWeb" title="#opt-GrpcWeb">[GrpcWeb](grpcweb.md)</a> | Converts gRPC Web requests to HTTP/2 gRPC requests.                           | Request                   |
| <a id="opt-Headers" href="#opt-Headers" title="#opt-Headers">[Headers](headers.md)</a> | Adds / Updates headers                            | Security                    |
| <a id="opt-IPAllowList" href="#opt-IPAllowList" title="#opt-IPAllowList">[IPAllowList](ipallowlist.md)</a> | Limits the allowed client IPs                     | Security, Request lifecycle |
| <a id="opt-InFlightReq" href="#opt-InFlightReq" title="#opt-InFlightReq">[InFlightReq](inflightreq.md)</a> | Limits the number of simultaneous connections     | Security, Request lifecycle |
| <a id="opt-PassTLSClientCert" href="#opt-PassTLSClientCert" title="#opt-PassTLSClientCert">[PassTLSClientCert](passtlsclientcert.md)</a> | Adds Client Certificates in a Header              | Security                    |
| <a id="opt-RateLimit" href="#opt-RateLimit" title="#opt-RateLimit">[RateLimit](ratelimit.md)</a> | Limits the call frequency                         | Security, Request lifecycle |
| <a id="opt-RedirectScheme" href="#opt-RedirectScheme" title="#opt-RedirectScheme">[RedirectScheme](redirectscheme.md)</a> | Redirects based on scheme                         | Request lifecycle           |
| <a id="opt-RedirectRegex" href="#opt-RedirectRegex" title="#opt-RedirectRegex">[RedirectRegex](redirectregex.md)</a> | Redirects based on regex                          | Request lifecycle           |
| <a id="opt-ReplacePath" href="#opt-ReplacePath" title="#opt-ReplacePath">[ReplacePath](replacepath.md)</a> | Changes the path of the request                   | Path Modifier               |
| <a id="opt-ReplacePathRegex" href="#opt-ReplacePathRegex" title="#opt-ReplacePathRegex">[ReplacePathRegex](replacepathregex.md)</a> | Changes the path of the request                   | Path Modifier               |
| <a id="opt-Retry" href="#opt-Retry" title="#opt-Retry">[Retry](retry.md)</a> | Automatically retries in case of error            | Request lifecycle           |
| <a id="opt-StripPrefix" href="#opt-StripPrefix" title="#opt-StripPrefix">[StripPrefix](stripprefix.md)</a> | Changes the path of the request                   | Path Modifier               |
| <a id="opt-StripPrefixRegex" href="#opt-StripPrefixRegex" title="#opt-StripPrefixRegex">[StripPrefixRegex](stripprefixregex.md)</a> | Changes the path of the request                   | Path Modifier               |

## Community Middlewares

Please take a look at the community-contributed plugins in the [plugin catalog](https://plugins.traefik.io/plugins).

{% include-markdown "includes/traefik-for-business-applications.md" %}
