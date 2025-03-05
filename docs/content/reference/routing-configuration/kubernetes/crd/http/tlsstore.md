---
title: "TLSStore"
description: "TLS Store in Traefik Proxy"
---

In Traefik, certificates are grouped together in certificates stores. 

`TLSStore` is the CRD implementation of a [Traefik TLS Store](../../../http/tls/tls-certificates.md#certificates-stores).

Before creating `TLSStore` objects, you need to apply the [Traefik Kubernetes CRDs](https://doc.traefik.io/traefik/reference/dynamic-configuration/kubernetes-crd/#definitions) to your Kubernetes cluster.

!!! Tip "Default TLS Store"
    Traefik currently only uses the TLS Store named "default". This default `TLSStore` should be in a namespace discoverable by Traefik. Since it is used by default on `IngressRoute` and `IngressRouteTCP` objects, there never is a need to actually reference it. This means that you cannot have two stores that are named default in different Kubernetes namespaces. As a consequence, with respect to TLS stores, the only change that makes sense (and only if needed) is to configure the default `TLSStore`.

## Configuration Example

```yaml tab="TLSStore"
apiVersion: traefik.io/v1alpha1
kind: TLSStore
metadata:
  name: default
  
spec:
  defaultCertificate:
    secretName:  supersecret
```

## Configuration Options

| Field                                  | Description    | Required |
|:---------------------------------------|:-------------------------|:---------|
| `certificates[n].secretName`                         | List of Kubernetes [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/), each of them holding a key/certificate pair to add to the store. | No      |
| `defaultCertificate.secretName`        | Name of the Kubernetes [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) served for connections without a SNI, or without a matching domain. If no default certificate is provided, Traefik will use the generated one. Do not use if the option `defaultGeneratedCert` is set.  | No      |
| `defaultGeneratedCert.resolver`        | Name of the ACME resolver to use to generate the default certificate.<br /> Do not use if the option `defaultCertificate` is set.     | No      |
| `defaultGeneratedCert.domain.main`     | Main domain used to generate the default certificate.<br /> Do not use if the option `defaultCertificate` is set.      | No      |
| `defaultGeneratedCert.domain.sans`     | List of [Subject Alternative Name](https://en.wikipedia.org/wiki/Subject_Alternative_Name) used to generate the default certificate.<br /> Do not use if the option `defaultCertificate` is set.   | No      |

!!! note "DefaultCertificate vs DefaultGeneratedCert"
    If both `defaultCertificate` and `defaultGeneratedCert` are set, the TLS certificate contained in `defaultCertificate.secretName` is served. The ACME default certificate is not generated.
