---
title: "Traefik TLS Certificates Documentation"
description: "Learn how to configure the transport layer security (TLS) connection in Traefik Proxy. Read the technical documentation."
---

!!! info
    When a router has to handle HTTPS traffic, it should be specified with a `tls` field of the router definition.

## Certificates Definition

### Automated

See the [Let's Encrypt](../../../install-configuration/tls/certificate-resolvers/acme.md) page.

### User defined

To add / remove TLS certificates, even when Traefik is already running, their definition can be added to the [dynamic configuration](../../dynamic-configuration-methods.md#providing-dynamic-routing-configuration-to-traefik), in the `[[tls.certificates]]` section:

```yaml tab="Structured (YAML)"
tls:
  certificates:
    - certFile: /path/to/domain.cert
      keyFile: /path/to/domain.key
    - certFile: /path/to/other-domain.cert
      keyFile: /path/to/other-domain.key
```

```toml tab="Structured (TOML)"
[[tls.certificates]]
  certFile = "/path/to/domain.cert"
  keyFile = "/path/to/domain.key"

[[tls.certificates]]
  certFile = "/path/to/other-domain.cert"
  keyFile = "/path/to/other-domain.key"
```

!!! important "Restriction"

    In the above example, we've used the [file provider](../../../install-configuration/providers/others/file.md) to handle these definitions.
    It is the only available method to configure the certificates (as well as the options and the stores).
    However, in [Kubernetes](../../../install-configuration/providers/kubernetes/kubernetes-crd.md), the certificates can and must be provided by [secrets](https://kubernetes.io/docs/concepts/configuration/secret/).

## Certificates Stores

In Traefik, certificates are grouped together in certificates stores.

!!! important "Restriction"

    Any store definition other than the default one (named `default`) will be ignored,
    and there is therefore only one globally available TLS store.

In the `tls.certificates` section, a list of stores can then be specified to indicate where the certificates should be stored:

```yaml tab="Structured (YAML)"
tls:
  certificates:
    - certFile: /path/to/domain.cert
      keyFile: /path/to/domain.key
      stores:
        - default
    # Note that since no store is defined,
    # the certificate below will be stored in the `default` store.
    - certFile: /path/to/other-domain.cert
      keyFile: /path/to/other-domain.key
```

```toml tab="Structured (TOML)"
[[tls.certificates]]
  certFile = "/path/to/domain.cert"
  keyFile = "/path/to/domain.key"
  stores = ["default"]

[[tls.certificates]]
  # Note that since no store is defined,
  # the certificate below will be stored in the `default` store.
  certFile = "/path/to/other-domain.cert"
  keyFile = "/path/to/other-domain.key"
```

!!! important "Restriction"

    The `stores` list will actually be ignored and automatically set to `["default"]`.

### Default Certificate

Traefik can use a default certificate for connections without a SNI, or without a matching domain.
This default certificate should be defined in a TLS store:

```yaml tab="Structured (YAML)"
tls:
  stores:
    default:
      defaultCertificate:
        certFile: path/to/cert.crt
        keyFile: path/to/cert.key
```

```toml tab="Structured (TOML)"
[tls.stores]
  [tls.stores.default]
    [tls.stores.default.defaultCertificate]
      certFile = "path/to/cert.crt"
      keyFile  = "path/to/cert.key"
```

If no `defaultCertificate` is provided, Traefik will use the generated one.

### ACME Default Certificate

You can configure Traefik to use an ACME provider (like Let's Encrypt) to generate the default certificate.
The configuration to resolve the default certificate should be defined in a TLS store:

!!! important "Precedence with the `defaultGeneratedCert` option"

    The `defaultGeneratedCert` definition takes precedence over the ACME default certificate configuration.

```yaml tab="Structured (YAML)"
tls:
  stores:
    default:
      defaultGeneratedCert:
        resolver: myresolver
        domain:
          main: example.org
          sans:
            - foo.example.org
            - bar.example.org
```

```toml tab="Structured (TOML)"
[tls.stores]
  [tls.stores.default.defaultGeneratedCert]
    resolver = "myresolver"
    [tls.stores.default.defaultGeneratedCert.domain]
      main = "example.org"
      sans = ["foo.example.org", "bar.example.org"]
```

```yaml tab="Labels"
labels:
  - "traefik.tls.stores.default.defaultgeneratedcert.resolver=myresolver"
  - "traefik.tls.stores.default.defaultgeneratedcert.domain.main=example.org"
  - "traefik.tls.stores.default.defaultgeneratedcert.domain.sans=foo.example.org, bar.example.org"
```

```json tab="Tags"
{
  "Name": "default",
  "Tags": [
    "traefik.tls.stores.default.defaultgeneratedcert.resolver=myresolver",
    "traefik.tls.stores.default.defaultgeneratedcert.domain.main=example.org",
    "traefik.tls.stores.default.defaultgeneratedcert.domain.sans=foo.example.org, bar.example.org"
  ]
}
```

{!traefik-for-business-applications.md!}
