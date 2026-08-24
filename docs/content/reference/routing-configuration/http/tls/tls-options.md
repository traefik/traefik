---
title: "Traefik TLS Options Documentation"
description: "Learn how to configure the transport layer security (TLS) connection in Traefik Proxy. Read the technical documentation."
---

The TLS options allow one to configure some parameters of the TLS connection.

!!! important "'default' TLS Option"

    The `default` option is special.
    When no tls options are specified in a tls router, the `default` option is used.  
    When specifying the `default` option explicitly, make sure not to specify provider namespace as the `default` option does not have one.  
    Conversely, for cross-provider references, for example, when referencing the file provider from a docker label,
    you must specify the provider namespace, for example:  
    `traefik.http.routers.myrouter.tls.options=myoptions@file`

!!! important "Providers"

    TLS options are not supported by label or tag-based providers. However, you can define them when using a [KV provider](../../other-providers/kv.md).

!!! important "TLSOption in Kubernetes"

    With the [TLSOption resource](../../kubernetes/crd/tls/tlsoption.md), the option named `default` applies to every router
    that does not reference a TLSOption explicitly, whatever the namespace it is defined in.
    The [`defaultTLSResourcesNamespace`](../../../install-configuration/providers/kubernetes/kubernetes-crd.md#defaulttlsresourcesnamespace) provider option
    restricts the namespace this cluster-wide default can be defined in.

### Server Name Association

The TLS options are configured on a router, but they are applied during the TLS handshake,
that is to say before the routing occurs, when the server name (SNI) is the only information available.
A TLS options reference is therefore always mapped to the host names found in the `Host` part of the router rule,
and neither to the router nor to its rule.
There could also be several `Host` parts in a rule, in which case the TLS options reference is mapped to as many host names.

In the case of domain fronting, if the TLS options associated with the Host header and the SNI are different,
Traefik responds with a `421 Misdirected Request` status code.

### Conflicting TLS Options

Since a TLS options reference is mapped to a host name, a conflict occurs when a configuration introduces a situation
where the same host name, on the same entry point, is matched with two different TLS options references,
such as in the example below:

```yaml tab="Structured (YAML)"
# Dynamic configuration

http:
  routers:
    routerfoo:
      rule: "Host(`example.com`) && Path(`/foo`)"
      tls:
        options: foo

    routerbar:
      rule: "Host(`example.com`) && Path(`/bar`)"
      tls:
        options: bar
```

```toml tab="Structured (TOML)"
# Dynamic configuration

[http.routers]
  [http.routers.routerfoo]
    rule = "Host(`example.com`) && Path(`/foo`)"
    [http.routers.routerfoo.tls]
      options = "foo"

  [http.routers.routerbar]
    rule = "Host(`example.com`) && Path(`/bar`)"
    [http.routers.routerbar.tls]
      options = "bar"
```

If that happens, both mappings are discarded, and the host name (`example.com` in this example)
gets associated with the `default` TLS options instead.

The conflict detection is not limited to a single provider:
routers coming from different providers, for example a router defined with a container label
and another one defined with the file provider, conflict with each other as soon as they serve
the same host name on the same entry point.

!!! important "Default TLS Options"

    The `default` TLS options are the fallback of the conflict resolution,
    and should therefore not be less secure than the options they can replace.
    A router relying on a mutual TLS authentication (`clientAuth`), for example,
    no longer enforces it if a conflict on its host name falls back to `default`
    TLS options that do not require it.

    The surest way to avoid this is to have all the routers serving the same host name,
    on the same entry point, reference the same TLS options.

#### Strict TLS Options

The [`core.strictTLSOptions`](../../../install-configuration/configuration-options.md#opt-core-stricttlsoptions)
install configuration option disables the fallback to the `default` TLS options.
When it is enabled, the routers involved in the conflict are marked in error and are not built at all,
and the host name is no longer mapped to any TLS options.

!!! warning "Disabled routers"

    Enabling `strictTLSOptions` fails closed: a conflict disables all the routers serving the conflicting host name
    on the concerned entry point, until the conflict is resolved.

```yaml tab="File (YAML)"
## Install configuration
core:
  strictTLSOptions: true
```

```toml tab="File (TOML)"
## Install configuration
[core]
  strictTLSOptions = true
```

```bash tab="CLI"
## Install configuration
--core.strictTLSOptions=true
```

### Minimum TLS Version

```yaml tab="Structured (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      minVersion: VersionTLS12

    mintls13:
      minVersion: VersionTLS13
```

```toml tab="Structured (TOML)"
# Dynamic configuration

[tls.options]

  [tls.options.default]
    minVersion = "VersionTLS12"

  [tls.options.mintls13]
    minVersion = "VersionTLS13"
```

### Maximum TLS Version

We discourage the use of this setting to disable TLS1.3.

The recommended approach is to update the clients to support TLS1.3.

```yaml tab="Structured (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      maxVersion: VersionTLS13

    maxtls12:
      maxVersion: VersionTLS12
```

```toml tab="Structured (TOML)"
# Dynamic configuration

[tls.options]

  [tls.options.default]
    maxVersion = "VersionTLS13"

  [tls.options.maxtls12]
    maxVersion = "VersionTLS12"
```

### Cipher Suites

See [cipherSuites](https://godoc.org/crypto/tls#pkg-constants) for more information.

```yaml tab="Structured (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      cipherSuites:
        - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
```

```toml tab="Structured (TOML)"
# Dynamic configuration

[tls.options]
  [tls.options.default]
    cipherSuites = [
      "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
    ]
```

!!! important "TLS 1.3"

    Cipher suites defined for TLS 1.2 and below cannot be used in TLS 1.3, and vice versa. (<https://tools.ietf.org/html/rfc8446>)  
    With TLS 1.3, the cipher suites are not configurable (all supported cipher suites are safe in this case).
    <https://golang.org/doc/go1.12#tls_1_3>

### Curve Preferences

This option allows setting the preferred elliptic curves.

The names of the curves defined by [`crypto`](https://godoc.org/crypto/tls#CurveID) (e.g. `CurveP521`) and the [RFC defined names](https://tools.ietf.org/html/rfc8446#section-4.2.7) (e. g. `secp521r1`) can be used.

See [CurveID](https://godoc.org/crypto/tls#CurveID) for more information.

```yaml tab="Structured (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      curvePreferences:
        - CurveP521
        - CurveP384
```

```toml tab="Structured (TOML)"
# Dynamic configuration

[tls.options]
  [tls.options.default]
    curvePreferences = ["CurveP521", "CurveP384"]
```

### Strict SNI Checking

With strict SNI checking enabled, Traefik won't allow connections from clients that do not specify a server_name extension
or don't match any of the configured certificates.
The default certificate is irrelevant on that matter.

```yaml tab="Structured (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      sniStrict: true
```

```toml tab="Structured (TOML)"
# Dynamic configuration

[tls.options]
  [tls.options.default]
    sniStrict = true
```

### ALPN Protocols

_Optional, Default="h2, http/1.1, acme-tls/1"_

This option allows specifying the list of supported application level protocols for the TLS handshake,
in order of preference.
If the client supports ALPN, the selected protocol will be one from this list, 
and the connection will fail if there is no mutually supported protocol.

```yaml tab="Structured (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      alpnProtocols:
        - http/1.1
        - h2
```

```toml tab="Structured (TOML)"
# Dynamic configuration

[tls.options]
  [tls.options.default]
    alpnProtocols = ["http/1.1", "h2"]
```

### Client Authentication (mTLS)

Traefik supports mutual authentication, through the `clientAuth` section.

For authentication policies that require verification of the client certificate, the certificate authority for the certificates should be set in `clientAuth.caFiles`.

In Kubernetes environment, CA certificate can be set in `clientAuth.secretNames`. See [TLSOption resource](../../kubernetes/crd/tls/tlsoption.md) for more details.

The `clientAuth.clientAuthType` option governs the behaviour as follows:

| Option    |  Operation | 
| --------- | ----------- |
| <a id="opt-NoClientCert" href="#opt-NoClientCert" title="#opt-NoClientCert">`NoClientCert`</a> | Disregards any client certificate.| 
| <a id="opt-RequestClientCert" href="#opt-RequestClientCert" title="#opt-RequestClientCert">`RequestClientCert`</a> | Asks for a certificate but proceeds anyway if none is provided. |
| <a id="opt-RequireAnyClientCert" href="#opt-RequireAnyClientCert" title="#opt-RequireAnyClientCert">`RequireAnyClientCert`</a> | Requires a certificate but does not verify if it is signed by a CA listed in `clientAuth.caFiles` or in `clientAuth.secretNames`. |
| <a id="opt-VerifyClientCertIfGiven" href="#opt-VerifyClientCertIfGiven" title="#opt-VerifyClientCertIfGiven">`VerifyClientCertIfGiven`</a> | If a certificate is provided, verifies if it is signed by a CA listed in `clientAuth.caFiles` or in `clientAuth.secretNames`. Otherwise proceeds without any certificate. |
| <a id="opt-RequireAndVerifyClientCert" href="#opt-RequireAndVerifyClientCert" title="#opt-RequireAndVerifyClientCert">`RequireAndVerifyClientCert`</a> |  requires a certificate, which must be signed by a CA listed in `clientAuth.caFiles` or in `clientAuth.secretNames`. |

```yaml tab="Structured (YAML)"
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

```toml tab="Structured (TOML)"
# Dynamic configuration

[tls.options]
  [tls.options.default]
    [tls.options.default.clientAuth]
      # in PEM format. each file can contain multiple CAs.
      caFiles = ["tests/clientca1.crt", "tests/clientca2.crt"]
      clientAuthType = "RequireAndVerifyClientCert"
```

### Disable Session Tickets

_Optional, Default="false"_

When set to true, Traefik disables the use of session tickets, forcing every client to perform a full TLS handshake instead of resuming sessions.

```yaml tab="Structured (YAML)"
# routing configuration

tls:
  options:
    default:
      disableSessionTickets: true
```

```toml tab="Structured (TOML)"
# routing configuration

[tls.options]
  [tls.options.default]
    disableSessionTickets = true
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: TLSOption
metadata:
  name: default
  namespace: default

spec:
  disableSessionTickets: true
```

### Encrypted Client Hello Keys

_Optional_

The `echKeys` option enables server-side Encrypted Client Hello (ECH), on both TCP (HTTP/1.1, HTTP/2) and HTTP/3 connections.
Each value references a PEM file (or holds inline PEM content) containing the private key and the ECH configuration list for a DNS public name.

The file format follows [RFC 9934](https://www.rfc-editor.org/rfc/rfc9934.html): a PKCS#8 `PRIVATE KEY` block followed by an `ECHCONFIG` block containing an encoded `ECHConfigList`.
Traefik currently supports X25519 ECH private keys.
See [RFC 9849](https://www.rfc-editor.org/rfc/rfc9849.html) for the ECH protocol.

!!! info "Provider availability"

    The `echKeys` option can be set with the [File](../../other-providers/file.md) and [KV](../../other-providers/kv.md) providers.
    The Kubernetes `TLSOption` custom resource does not expose it yet.

From a Traefik source checkout, generate a key file for a public name:

```bash
go run ./internal/ech/cmd generate example.com > /etc/traefik/ech.pem
```

```text
-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VuBCIEICjd4yGRdsoP9gU7YT7My8DHx1Tjme8GYDXrOMCi8v1V
-----END PRIVATE KEY-----
-----BEGIN ECHCONFIG-----
AD7+DQA65wAgACA8wVN2BtscOl3vQheUzHeIkVmKIiydUhDCliA4iyQRCwAEAAEA
AQALZXhhbXBsZS5jb20AAA==
-----END ECHCONFIG-----
```

An ECH connection starts with the public name as server name, so Traefik terminates it with the TLS option matching the public name — the `default` option unless a router matches the public name.
The ECH keys and a `minVersion` set to `VersionTLS13` must be configured on that option: ECH requires TLS 1.3, and the Go TLS stack expects `VersionTLS13` as minimum version when ECH keys are set.
Clients that do not support ECH can still connect with regular TLS 1.3.

```yaml tab="Structured (YAML)"
# Routing configuration

tls:
  options:
    default:
      minVersion: VersionTLS13
      echKeys:
        - /etc/traefik/ech.pem
```

```toml tab="Structured (TOML)"
# Routing configuration

[tls.options]
  [tls.options.default]
    minVersion = "VersionTLS13"
    echKeys = ["/etc/traefik/ech.pem"]
```

!!! warning "Shared TLS option"

    On TCP connections, the protected domains inherit the public name's TLS option for the whole connection: a request to a router configured with a different TLS option is rejected with a `421 Misdirected Request`.
    All domains protected behind a public name must therefore share the public name's TLS option.
    On HTTP/3 connections, the TLS option is instead selected with the decrypted (protected) server name.

A certificate valid for the public name must also be configured: clients validate it whenever ECH is rejected and retried, for example while a key rotation propagates.
Enabling `sniStrict` without such a certificate breaks this retry mechanism.

#### DNS record

Clients only send ECH once they discover the ECH configuration in the `HTTPS` DNS record of the protected domain (`ech` parameter, see [RFC 9848](https://www.rfc-editor.org/rfc/rfc9848.html)).
The base64 payload of the `ECHCONFIG` block is the `ECHConfigList` value to publish:

```text
protected.example.com. 300 IN HTTPS 1 . ech=AD7+DQA65wAgACA8wVN2BtscOl3vQheUzHeIkVmKIiydUhDCliA4iyQRCwAEAAEAAQALZXhhbXBsZS5jb20AAA==
```

When rotating keys, list both the new and the previous key files in `echKeys` until DNS caches no longer serve the old configuration.

!!! warning "Private key material"

    ECH files contain private key material. Restrict file access to the Traefik process.

{% include-markdown "includes/traefik-for-business-applications.md" %}
