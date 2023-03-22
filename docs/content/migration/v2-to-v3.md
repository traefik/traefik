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

## Deprecated Options Removal

- The `pilot` option has been removed from the static configuration.
- The `tracing.datadog.globaltag` option has been removed.
- The `namespace` option of Consul, Consul Catalog and Nomad providers has been removed.
- The `tls.caOptional` option has been removed from the ForwardAuth middleware, as well as from the HTTP, Consul, Etcd, Redis, ZooKeeper, Consul Catalog, and Docker providers.
- `sslRedirect`, `sslTemporaryRedirect`, `sslHost`, `sslForceHost` and `featurePolicy` options of the Headers middleware have been removed.
- The `forceSlash` option of the StripPrefix middleware has been removed.
- The `preferServerCipherSuites` option has been removed.

## Matchers

In v3, the `Headers` and `HeadersRegexp` matchers have been renamed to `Header` and `HeaderRegexp` respectively.

`PathPrefix` no longer uses regular expressions to match path prefixes.

`QueryRegexp` has been introduced to match query values using a regular expression.

`HeaderRegexp`, `HostRegexp`, `PathRegexp`, `QueryRegexp`, and `HostSNIRegexp` matchers now uses the [Go regexp syntax](https://golang.org/pkg/regexp/syntax/).

All matchers now take a single value (except `Header`, `HeaderRegexp`, `Query`, and `QueryRegexp` which take two)
and should be explicitly combined using logical operators to mimic previous behavior.

`Query` can take a single value to match is the query value that has no value (e.g. `/search?mobile`).

`HostHeader` has been removed, use `Host` instead.

## Content-Type Auto-Detection

In v3, the `Content-Type` header is not auto-detected anymore when it is not set by the backend.
One should use the `ContentType` middleware to enable the `Content-Type` header value auto-detection.

## HTTP/3

In v3, HTTP/3 is no longer an experimental feature.
The `experimental.http3` option has been removed from the static configuration.

## TCP ServersTransport

In v3, the support of `TCPServersTransport` has been introduced.
When using the KubernetesCRD provider, it is therefore necessary to update [RBAC](../reference/dynamic-configuration/kubernetes-crd.md#rbac) and [CRD](../reference/dynamic-configuration/kubernetes-crd.md) manifests.

### TCP LoadBalancer `terminationDelay` option

The TCP LoadBalancer `terminationDelay` option has been removed.
This option can now be configured directly on the `TCPServersTransport` level, please take a look at this [documentation](../routing/services/index.md#terminationdelay)

## Rancher v1

In v3, the rancher v1 provider has been removed because Rancher v1 is [no longer actively maintaned](https://rancher.com/docs/os/v1.x/en/support/) and v2 is supported as a standard Kubernetes provider.

Rancher 2.x requires Kubernetes and does not have a metadata endpoint of its own for Traefik to query.
As such, Rancher 2.x users should utilize the [Kubernetes CRD provider](../providers/kubernetes-crd.md) directly.

## Marathon provider

In v3, the Marathon provider has been removed.

## InfluxDB v1

In v3, the InfluxDB v1 metrics provider has been removed because InfluxDB v1.x maintenance [ended in 2021](https://www.influxdata.com/blog/influxdb-oss-and-enterprise-roadmap-update-from-influxdays-emea/).

### Kubernetes CRDs API Group `traefik.containo.us`

In v3 the Kubernetes CRDs API Group `traefik.containo.us` has been removed. 
Please use the API Group `traefik.io` instead.
