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
