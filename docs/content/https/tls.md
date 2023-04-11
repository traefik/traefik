---
title: "Traefik TLS Documentation"
description: "Learn how to configure the transport layer security (TLS) connection in Traefik Proxy. Read the technical documentation."
---

# TLS

Transport Layer Security
{: .subtitle }

## Certificates Definition

### Automated

See the [Let's Encrypt](./acme.md) page.

### User defined

To add / remove TLS certificates, even when Traefik is already running, their definition can be added to the [dynamic configuration](../getting-started/configuration-overview.md), in the `[[tls.certificates]]` section:

```yaml tab="File (YAML)"
# Dynamic configuration

tls:
  certificates:
    - certFile: /path/to/domain.cert
      keyFile: /path/to/domain.key
    - certFile: /path/to/other-domain.cert
      keyFile: /path/to/other-domain.key
```

```toml tab="File (TOML)"
# Dynamic configuration

[[tls.certificates]]
  certFile = "/path/to/domain.cert"
  keyFile = "/path/to/domain.key"

[[tls.certificates]]
  certFile = "/path/to/other-domain.cert"
  keyFile = "/path/to/other-domain.key"
```

!!! important "Restriction"

    In the above example, we've used the [file provider](../providers/file.md) to handle these definitions.
    It is the only available method to configure the certificates (as well as the options and the stores).
    However, in [Kubernetes](../providers/kubernetes-crd.md), the certificates can and must be provided by [secrets](https://kubernetes.io/docs/concepts/configuration/secret/).

## Certificates Stores

In Traefik, certificates are grouped together in certificates stores, which are defined as such:

```yaml tab="File (YAML)"
# Dynamic configuration

tls:
  stores:
    default: {}
```

```toml tab="File (TOML)"
# Dynamic configuration

[tls.stores]
  [tls.stores.default]
```

!!! important "Restriction"

    Any store definition other than the default one (named `default`) will be ignored,
    and there is therefore only one globally available TLS store.

In the `tls.certificates` section, a list of stores can then be specified to indicate where the certificates should be stored:

```yaml tab="File (YAML)"
# Dynamic configuration

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

```toml tab="File (TOML)"
# Dynamic configuration

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

```yaml tab="File (YAML)"
# Dynamic configuration

tls:
  stores:
    default:
      defaultCertificate:
        certFile: path/to/cert.crt
        keyFile: path/to/cert.key
```

```toml tab="File (TOML)"
# Dynamic configuration

[tls.stores]
  [tls.stores.default]
    [tls.stores.default.defaultCertificate]
      certFile = "path/to/cert.crt"
      keyFile  = "path/to/cert.key"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: TLSStore
metadata:
  name: default
  namespace: default

spec:
  defaultCertificate:
    secretName: default-certificate
    
---
apiVersion: v1
kind: Secret
metadata:
  name: default-certificate
  namespace: default
  
type: Opaque
data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0=
```

If no `defaultCertificate` is provided, Traefik will use the generated one.

### ACME Default Certificate

You can configure Traefik to use an ACME provider (like Let's Encrypt) to generate the default certificate.
The configuration to resolve the default certificate should be defined in a TLS store:

!!! important "Precedence with the `defaultGeneratedCert` option"

    The `defaultGeneratedCert` definition takes precedence over the ACME default certificate configuration.

```yaml tab="File (YAML)"
# Dynamic configuration

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

```toml tab="File (TOML)"
# Dynamic configuration

[tls.stores]
  [tls.stores.default.defaultGeneratedCert]
    resolver = "myresolver"
    [tls.stores.default.defaultGeneratedCert.domain]
      main = "example.org"
      sans = ["foo.example.org", "bar.example.org"]
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: TLSStore
metadata:
  name: default
  namespace: default

spec:
  defaultGeneratedCert:
    resolver: myresolver
    domain:
      main: example.org
      sans:
        - foo.example.org
        - bar.example.org
```

```yaml tab="Docker"
## Dynamic configuration
labels:
  - "traefik.tls.stores.default.defaultgeneratedcert.resolver=myresolver"
  - "traefik.tls.stores.default.defaultgeneratedcert.domain.main=example.org"
  - "traefik.tls.stores.default.defaultgeneratedcert.domain.sans=foo.example.org, bar.example.org"
```

```json tab="Marathon"
labels: {
  "traefik.tls.stores.default.defaultgeneratedcert.resolver": "myresolver",
  "traefik.tls.stores.default.defaultgeneratedcert.domain.main": "example.org",
  "traefik.tls.stores.default.defaultgeneratedcert.domain.sans": "foo.example.org, bar.example.org",
}
```

## TLS Options

The TLS options allow one to configure some parameters of the TLS connection.

!!! important "'default' TLS Option"

    The `default` option is special.
    When no tls options are specified in a tls router, the `default` option is used.  
    When specifying the `default` option explicitly, make sure not to specify provider namespace as the `default` option does not have one.  
    Conversely, for cross-provider references, for example, when referencing the file provider from a docker label,
    you must specify the provider namespace, for example:  
    `traefik.http.routers.myrouter.tls.options=myoptions@file`

!!! important "TLSOption in Kubernetes"

    When using the [TLSOption resource](../../routing/providers/kubernetes-crd#kind-tlsoption) in Kubernetes, one might setup a default set of options that,
    if not explicitly overwritten, should apply to all ingresses.  
    To achieve that, you'll have to create a TLSOption resource with the name `default`.
    There may exist only one TLSOption with the name `default` (across all namespaces) - otherwise they will be dropped.  
    To explicitly use a different TLSOption (and using the Kubernetes Ingress resources)
    you'll have to add an annotation to the Ingress in the following form:
    `traefik.ingress.kubernetes.io/router.tls.options: <resource-namespace>-<resource-name>@kubernetescrd`

### Minimum TLS Version

```yaml tab="File (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      minVersion: VersionTLS12

    mintls13:
      minVersion: VersionTLS13
```

```toml tab="File (TOML)"
# Dynamic configuration

[tls.options]

  [tls.options.default]
    minVersion = "VersionTLS12"

  [tls.options.mintls13]
    minVersion = "VersionTLS13"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: TLSOption
metadata:
  name: default
  namespace: default

spec:
  minVersion: VersionTLS12

---
apiVersion: traefik.io/v1alpha1
kind: TLSOption
metadata:
  name: mintls13
  namespace: default

spec:
  minVersion: VersionTLS13
```

### Maximum TLS Version

We discourage the use of this setting to disable TLS1.3.

The recommended approach is to update the clients to support TLS1.3.

```yaml tab="File (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      maxVersion: VersionTLS13

    maxtls12:
      maxVersion: VersionTLS12
```

```toml tab="File (TOML)"
# Dynamic configuration

[tls.options]

  [tls.options.default]
    maxVersion = "VersionTLS13"

  [tls.options.maxtls12]
    maxVersion = "VersionTLS12"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: TLSOption
metadata:
  name: default
  namespace: default

spec:
  maxVersion: VersionTLS13

---
apiVersion: traefik.io/v1alpha1
kind: TLSOption
metadata:
  name: maxtls12
  namespace: default

spec:
  maxVersion: VersionTLS12
```

### Cipher Suites

See [cipherSuites](https://godoc.org/crypto/tls#pkg-constants) for more information.

```yaml tab="File (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      cipherSuites:
        - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
```

```toml tab="File (TOML)"
# Dynamic configuration

[tls.options]
  [tls.options.default]
    cipherSuites = [
      "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
    ]
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: TLSOption
metadata:
  name: default
  namespace: default

spec:
  cipherSuites:
    - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
```

!!! important "TLS 1.3"

    Cipher suites defined for TLS 1.2 and below cannot be used in TLS 1.3, and vice versa. (<https://tools.ietf.org/html/rfc8446>)  
    With TLS 1.3, the cipher suites are not configurable (all supported cipher suites are safe in this case).
    <https://golang.org/doc/go1.12#tls_1_3>

### Curve Preferences

This option allows to set the preferred elliptic curves in a specific order.

The names of the curves defined by [`crypto`](https://godoc.org/crypto/tls#CurveID) (e.g. `CurveP521`) and the [RFC defined names](https://tools.ietf.org/html/rfc8446#section-4.2.7) (e. g. `secp521r1`) can be used.

See [CurveID](https://godoc.org/crypto/tls#CurveID) for more information.

```yaml tab="File (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      curvePreferences:
        - CurveP521
        - CurveP384
```

```toml tab="File (TOML)"
# Dynamic configuration

[tls.options]
  [tls.options.default]
    curvePreferences = ["CurveP521", "CurveP384"]
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: TLSOption
metadata:
  name: default
  namespace: default

spec:
  curvePreferences:
    - CurveP521
    - CurveP384
```

### Strict SNI Checking

With strict SNI checking enabled, Traefik won't allow connections from clients that do not specify a server_name extension
or don't match any of the configured certificates.
The default certificate is irrelevant on that matter.

```yaml tab="File (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      sniStrict: true
```

```toml tab="File (TOML)"
# Dynamic configuration

[tls.options]
  [tls.options.default]
    sniStrict = true
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: TLSOption
metadata:
  name: default
  namespace: default

spec:
  sniStrict: true
```

### ALPN Protocols

_Optional, Default="h2, http/1.1, acme-tls/1"_

This option allows to specify the list of supported application level protocols for the TLS handshake,
in order of preference.
If the client supports ALPN, the selected protocol will be one from this list, 
and the connection will fail if there is no mutually supported protocol.

```yaml tab="File (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      alpnProtocols:
        - http/1.1
        - h2
```

```toml tab="File (TOML)"
# Dynamic configuration

[tls.options]
  [tls.options.default]
    alpnProtocols = ["http/1.1", "h2"]
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: TLSOption
metadata:
  name: default
  namespace: default

spec:
  alpnProtocols:
    - http/1.1
    - h2
```

### Client Authentication (mTLS)

Traefik supports mutual authentication, through the `clientAuth` section.

For authentication policies that require verification of the client certificate, the certificate authority for the certificates should be set in `clientAuth.caFiles`.

In Kubernetes environment, CA certificate can be set in `clientAuth.secretNames`. See [TLSOption resource](../../routing/providers/kubernetes-crd#kind-tlsoption) for more details.

The `clientAuth.clientAuthType` option governs the behaviour as follows:

- `NoClientCert`: disregards any client certificate.
- `RequestClientCert`: asks for a certificate but proceeds anyway if none is provided.
- `RequireAnyClientCert`: requires a certificate but does not verify if it is signed by a CA listed in `clientAuth.caFiles` or in `clientAuth.secretNames`.
- `VerifyClientCertIfGiven`: if a certificate is provided, verifies if it is signed by a CA listed in `clientAuth.caFiles` or in `clientAuth.secretNames`. Otherwise proceeds without any certificate.
- `RequireAndVerifyClientCert`: requires a certificate, which must be signed by a CA listed in `clientAuth.caFiles` or in `clientAuth.secretNames`.

```yaml tab="File (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      clientAuth:
        # in PEM format. each file can contain multiple CAs.
        caFiles:
          - tests/clientca1.crt
          - tests/clientca2.crt
        clientAuthType: RequireAndVerifyClientCert
```

```toml tab="File (TOML)"
# Dynamic configuration

[tls.options]
  [tls.options.default]
    [tls.options.default.clientAuth]
      # in PEM format. each file can contain multiple CAs.
      caFiles = ["tests/clientca1.crt", "tests/clientca2.crt"]
      clientAuthType = "RequireAndVerifyClientCert"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: TLSOption
metadata:
  name: default
  namespace: default

spec:
  clientAuth:
    # the CA certificate is extracted from key `tls.ca` or `ca.crt` of the given secrets.
    secretNames:
      - secretCA
    clientAuthType: RequireAndVerifyClientCert
```

{!traefik-for-business-applications.md!}
