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

If no default certificate is provided, Traefik generates and uses a self-signed certificate.

## TLS Options

The TLS options allow one to configure some parameters on TLS connections.

!!! important "TLSOptions in Kubernetes"

    On Kubernetes you have to use [TLSOption](../../routing/providers/kubernetes-crd#kind-tlsoption).
    The configuration file won't apply on your IngressRoute / IngressRouteTCP.

On other environments, one can provide TLS Options with

**1)** A configuration file

```yaml tab="traefik-tls.yaml"
tls:
  options:
    default:
      minVersion: VersionTLS13

    tls12:
      minVersion: VersionTLS12
    cipherSuites:
      - "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
      - "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
      - "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
      - "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
      - "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305"
      - "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"

```

```toml tab="traefik-tls.toml"
[tls.options]
  [tls.options.default]
    minVersion = "VersionTLS13"

  [tls.options.tls12]
    minVersion = "VersionTLS12"
    cipherSuites = [
      "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
      "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
      "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
      "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
      "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
      "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"
    ]
```

**2)** a configuration file provider

```yaml tab="File (YAML)"
providers:
  docker:
    endpoint: "tcp://dockersocketproxy:2375"
    exposedByDefault: false

  file:
    filename: "/traefik-tls.yaml"
```

```toml tab="File (TOML)"
[providers]
  [providers.docker]
    endpoint = "tcp://dockersocketproxy:2375"
    exposedByDefault = false

  [providers.file]
    filename = "/traefik-tls.toml"

```

**3)** select tls options with labels on your routers

```yaml tab="Default"
traefik.http.routers.frontend.tls.options: "default"
```

```yaml tab="TLS 1.2"
traefik.http.routers.frontend.tls.options: "tls12@file"
```

!!! important "'default' TLS Option"

    The `default` option is special.
    When no tls options are specified in a tls router, the `default` option is used.
    Conversely, for cross-provider references, for example, when referencing the file
    provider from a docker label, you must specify the provider, like in the example above.

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
apiVersion: traefik.containo.us/v1alpha1
kind: TLSOption
metadata:
  name: default
  namespace: default

spec:
  minVersion: VersionTLS12

---
apiVersion: traefik.containo.us/v1alpha1
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
apiVersion: traefik.containo.us/v1alpha1
kind: TLSOption
metadata:
  name: default
  namespace: default

spec:
  maxVersion: VersionTLS13

---
apiVersion: traefik.containo.us/v1alpha1
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
apiVersion: traefik.containo.us/v1alpha1
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
apiVersion: traefik.containo.us/v1alpha1
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

With strict SNI checking enabled, Traefik won't allow connections from clients
that do not specify a server_name extension or don't match any certificate configured on the tlsOption.

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
apiVersion: traefik.containo.us/v1alpha1
kind: TLSOption
metadata:
  name: default
  namespace: default

spec:
  sniStrict: true
```

### Prefer Server Cipher Suites

This option allows the server to choose its most preferred cipher suite instead of the client's.
Please note that this is enabled automatically when `minVersion` or `maxVersion` are set.

```yaml tab="File (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      preferServerCipherSuites: true
```

```toml tab="File (TOML)"
# Dynamic configuration

[tls.options]
  [tls.options.default]
    preferServerCipherSuites = true
```

```yaml tab="Kubernetes"
apiVersion: traefik.containo.us/v1alpha1
kind: TLSOption
metadata:
  name: default
  namespace: default

spec:
  preferServerCipherSuites: true
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
apiVersion: traefik.containo.us/v1alpha1
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

For authentication policies that require verification of the client certificate, the certificate authority for the certificate should be set in `clientAuth.caFiles`.

The `clientAuth.clientAuthType` option governs the behaviour as follows:

- `NoClientCert`: disregards any client certificate.
- `RequestClientCert`: asks for a certificate but proceeds anyway if none is provided.
- `RequireAnyClientCert`: requires a certificate but does not verify if it is signed by a CA listed in `clientAuth.caFiles`.
- `VerifyClientCertIfGiven`: if a certificate is provided, verifies if it is signed by a CA listed in `clientAuth.caFiles`. Otherwise proceeds without any certificate.
- `RequireAndVerifyClientCert`: requires a certificate, which must be signed by a CA listed in `clientAuth.caFiles`.

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
apiVersion: traefik.containo.us/v1alpha1
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
