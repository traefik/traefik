# TLS

Transport Layer Security
{: .subtitle }

## Certificates Definition

### Automated

See the [Let's Encrypt](./acme.md) page.

### User defined

To add / remove TLS certificates, even when Traefik is already running, their definition can be added to the [dynamic configuration](../getting-started/configuration-overview.md), in the `[[tls.certificates]]` section:

```toml tab="TOML"
[[tls.certificates]]
  certFile = "/path/to/domain.cert"
  keyFile = "/path/to/domain.key"

[[tls.certificates]]
  certFile = "/path/to/other-domain.cert"
  keyFile = "/path/to/other-domain.key"
```

```yaml tab="YAML"
tls:
  certificates:
  - certFile: /path/to/domain.cert
    keyFile: /path/to/domain.key
  - certFile: /path/to/other-domain.cert
    keyFile: /path/to/other-domain.key
```

!!! important "File Provider Only"

    In the above example, we've used the [file provider](../providers/file.md) to handle these definitions.
    In its current alpha version, it is the only available method to configure the certificates (as well as the options and the stores).

## Certificates Stores

In Traefik, certificates are grouped together in certificates stores, which are defined as such:

```toml tab="TOML"
[tls.stores]
  [tls.stores.default]
```

```yaml tab="YAML"
tls:
  stores:
    default: {}
```

!!! important "Alpha restriction"

    During the alpha version, any store definition other than the default one (named `default`) will be ignored,
    and there is thefore only one globally available TLS store.

In the `tls.certificates` section, a list of stores can then be specified to indicate where the certificates should be stored:

```toml tab="TOML"
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

```yaml tab="YAML"
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

!!! important "Alpha restriction"

    During the alpha version, the `stores` list will actually be ignored and automatically set to `["default"]`.

### Default Certificate

Traefik can use a default certificate for connections without a SNI, or without a matching domain.
This default certificate should be defined in a TLS store:

```toml tab="TOML"
[tls.stores]
  [tls.stores.default]
    [tls.stores.default.defaultCertificate]
      certFile = "path/to/cert.crt"
      keyFile  = "path/to/cert.key"
```

```yaml tab="YAML"
tls:
  stores:
    default:
      defaultCertificate:
        certFile: path/to/cert.crt
        keyFile: path/to/cert.key
```

If no default certificate is provided, Traefik generates and uses a self-signed certificate.

## TLS Options

The TLS options allow one to configure some parameters of the TLS connection.

### Minimum TLS Version

```toml tab="TOML"
[tls.options]

  [tls.options.default]
    minVersion = "VersionTLS12"

  [tls.options.mintls13]
    minVersion = "VersionTLS13"
```

```yaml tab="YAML"
tls:
  options:
    default:
      minVersion: VersionTLS12

    mintls13:
      minVersion: VersionTLS13
```

### Mutual Authentication

Traefik supports both optional and strict (which is the default) mutual authentication, though the `ClientCA.files` section.
If present, connections from clients without a certificate will be rejected.

For clients with a certificate, the `optional` option governs the behaviour as follows:

- When `optional = false`, Traefik accepts connections only from clients presenting a certificate signed by a CA listed in `ClientCA.files`.
- When `optional = true`, Traefik authorizes connections from clients presenting a certificate signed by an unknown CA.

```toml tab="TOML"
[tls.options]
  [tls.options.default]
    [tls.options.default.clientCA]
      # in PEM format. each file can contain multiple CAs.
      files = ["tests/clientca1.crt", "tests/clientca2.crt"]
      optional = false
```

```yaml tab="YAML"
tls:
  options:
    default:
      clientCA:
        # in PEM format. each file can contain multiple CAs.
        files:
        - tests/clientca1.crt
        - tests/clientca2.crt
        optional: false
```

### Cipher Suites

See [cipherSuites](https://godoc.org/crypto/tls#pkg-constants) for more information.

```toml tab="TOML"
[tls.options]
  [tls.options.default]
    cipherSuites = [
      "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
      "TLS_RSA_WITH_AES_256_GCM_SHA384"
    ]
```

```yaml tab="YAML"
tls:
  options:
    default:
      cipherSuites:
      - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
      - TLS_RSA_WITH_AES_256_GCM_SHA384
```

### Strict SNI Checking

With strict SNI checking, Traefik won't allow connections from clients connections
that do not specify a server_name extension.

```toml tab="TOML"
[tls.options]
  [tls.options.default]
    sniStrict = true
```

```yaml tab="YAML"
tls:
  options:
    default:
      sniStrict: true
```
