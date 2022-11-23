---
title: "Traefik V3 Migration Documentation"
description: "Migrate from Traefik Proxy v2 to v3 and update all the necessary configurations to take advantage of all the improvements. Read the technical documentation."
---

# Migration Guide: From v2 to v3

How to Migrate from Traefik v2 to Traefik v3.
{: .subtitle }

The version 3 of Traefik introduces a number of breaking changes,
which require one to update their configuration when they migrate from v2 to v3.
The goal of this page is to recapitulate all of these changes, and in particular to give examples,
feature by feature, of how the configuration looked like in v2, and how it now looks like in v3.

## IPWhiteList

In v3, we renamed the `IPWhiteList` middleware to `IPAllowList` without changing anything to the configuration. 

## gRPC Metrics

In v3, the reported status code for gRPC requests is now the value of the `Grpc-Status` header.    

## PassHostHeader

In v3, the `PassHostHeader` config option has been moved to the [ServersTransport](../routing/services/index.md#passhostheader) resource.

## ResponseForwarding.flushInterval

In v3, the `ResponseForwarding.flushInterval` config option has been removed as chunked responses are now handled automatically.

##  HTTP/2

In v3, HTTP/2 is disabled by default between Traefik and the backends unless it is enabled in the configured [ServersTransport](../routing/services/index.md#passhostheader) 
or the backend URL uses the `h2c` scheme.

## ServersTransport

In v3, the `default` [ServersTransport](./routing/services/index.md) should now be configured in the dynamic configuration.
The ServersTransport resource now contains the [tls](../routing/services/index.md#tls) and [http](../routing/services/index.md#http) top level config options.  
