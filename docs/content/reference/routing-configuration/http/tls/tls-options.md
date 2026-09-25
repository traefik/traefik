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

```yaml tab="Structured (YAML)"
## Install configuration
core:
  strictTLSOptions: true
```

```toml tab="Structured (TOML)"
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
| <a id="opt-RequireAndVerifyClientCertWithExpiry" href="#opt-RequireAndVerifyClientCertWithExpiry" title="#opt-RequireAndVerifyClientCertWithExpiry">`RequireAndVerifyClientCertWithExpiry`</a> |  requires a certificate, which must be signed by a CA listed in `clientAuth.caFiles` or in `clientAuth.secretNames`.<br /> Provided certificate must be valid according to it's CRL spec. More about CRL handling [here](#expiry-check) |

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

#### Expiry Validation

Traefik supports CRL handling through the `clientAuth.expiry.crl` section.
There are multiple ways to load and handle CRLs.

!!! important "Performance impact"

    CRL validation is handled after client certificate validation and increases the computational load of each request.
    You should be very careful about which routers use this option, as enabling it everywhere may significantly increase latency and/or lead to higher memory usage.
    Regardless of how CRLs are loaded, they are kept in Traefik's memory; very large CRL files and/or a large number of distribution points may lead to a significant memory footprint increase.

This section of the configuration requires `clientAuth.clientAuthType = "RequireAndVerifyClientCertWithExpiry"` to be effective.

!!! info "Global, file based CRLs"

    In addition to the per-TLS-Options settings described below, Traefik supports a global, file based CRL store configured once at the install configuration level. See [TLS install configuration](../../../install-configuration/tls/crl.md) for more details. When a certificate's distribution point is covered by that global store, it always takes precedence over the mechanisms described in this section.

| Option    |  Operation  |
| --------- | ----------- |
| <a id="opt-clientAuth-expiry-crl-mode" href="#opt-clientAuth-expiry-crl-mode" title="#opt-clientAuth-expiry-crl-mode">`clientAuth.expiry.crl.mode`</a> | Defines how Traefik will handle client certificate CRL validation.<br /> More info [here](#crl-validation-behavior) |
| <a id="opt-clientAuth-expiry-crl-reloadInterval" href="#opt-clientAuth-expiry-crl-reloadInterval" title="#opt-clientAuth-expiry-crl-reloadInterval">`clientAuth.expiry.crl.reloadInterval`</a> | Time between periodic refreshes of HTTP sourced CRLs. Setting this option to `0` disables the periodic refresh; the CRL is then only reloaded once it expires (`nextUpdate` elapsed).<br /> More info [here](#crl-sources-and-loading) |
| <a id="opt-clientAuth-expiry-crl-http-expirationStrategy" href="#opt-clientAuth-expiry-crl-http-expirationStrategy" title="#opt-clientAuth-expiry-crl-http-expirationStrategy">`clientAuth.expiry.crl.http.expirationStrategy`</a> | Defines how Traefik handles expired, or not yet refreshed, CRLs that were loaded through HTTP requests.<br /> More info [here](#crl-expiration-strategies). |
| <a id="opt-clientAuth-expiry-crl-http-timeout" href="#opt-clientAuth-expiry-crl-http-timeout" title="#opt-clientAuth-expiry-crl-http-timeout">`clientAuth.expiry.crl.http.timeout`</a> | CRL download timeout, for each CRL distribution point fetched over HTTP. If unset, a Traefik-wide default timeout is used.<br /> More info [here](#crl-sources-and-loading). |
| <a id="opt-clientAuth-expiry-crl-http-errorBackoff" href="#opt-clientAuth-expiry-crl-http-errorBackoff" title="#opt-clientAuth-expiry-crl-http-errorBackoff">`clientAuth.expiry.crl.http.errorBackoff`</a> | Duration between CRL refresh attempt on unrespondsive distribution points. Defaults to `30s` |
| <a id="opt-clientAuth-expiry-crl-http-maxCRLBytes" href="#opt-clientAuth-expiry-crl-http-maxCRLBytes" title="#opt-clientAuth-expiry-crl-http-maxCRLBytes">`clientAuth.expiry.crl.http.maxCRLBytes`</a> | Max CRL file size. defaults to `1MB` |
| <a id="opt-clientAuth-expiry-crl-http-whitelist-enabled" href="#opt-clientAuth-expiry-crl-http-whitelist-enabled" title="#opt-clientAuth-expiry-crl-http-whitelist-enabled">`clientAuth.expiry.crl.http.whitelist.enabled`</a> | Enables the CRL distribution point allow-list.<br /> More info [here](#crl-distribution-point-allow-listing). |
| <a id="opt-clientAuth-expiry-crl-http-whitelist-distributionPoints" href="#opt-clientAuth-expiry-crl-http-whitelist-distributionPoints" title="#opt-clientAuth-expiry-crl-http-whitelist-distributionPoints">`clientAuth.expiry.crl.http.whitelist.distributionPoints`</a> | Allowed CRL distribution point URLs (must serve a DER or PEM encoded CRL over HTTP(S)).<br /> More info [here](#crl-distribution-point-allow-listing). |

##### How It Works

When `clientAuthType` is set to `RequireAndVerifyClientCertWithExpiry`, Traefik evaluates, for every certificate of the verified client chain (excluding the self-signed root CA), whether that certificate appears as revoked in the CRL published at its `CRL Distribution Points` (CDP) X.509 extension. The certificate is validated against the CRL issued by its direct issuer, which is itself validated by the standard TLS handshake chain validation.

If no verified chain is available at all (e.g. no client certificate was ultimately validated), the connection is rejected in `Strict` mode, and allowed in `Lax`/`NOOP` modes.

##### CRL Validation Behavior

Traefik's validation behavior can be controlled via `clientAuth.expiry.crl.mode`:

| Mode    |  Description                                                  |
| ------- | ------------------------------------------------------------- |
| <a id="opt-NOOP" href="#opt-NOOP" title="#opt-NOOP">`NOOP`</a> | No operation: certificates are not checked for CRL attributes, and no revocation check is performed. This is the default. |
| <a id="opt-Lax" href="#opt-Lax" title="#opt-Lax">`Lax`</a> | Checks certificates for a CDP attribute and enforces revocation checking only when that attribute is present. Certificates without a CDP are allowed. For each certificate in the verified chain that does carry a CDP, Traefik checks it against the corresponding CRL. |
| <a id="opt-Strict" href="#opt-Strict" title="#opt-Strict">`Strict`</a> | Same as `Lax`, but additionally rejects certificates that do not carry a CDP attribute at all. |

##### CRL Sources and Loading

For a given CRL distribution point, Traefik first checks whether it is covered by the [global, file based CRL store](../../../install-configuration/tls/crl.md) configured at the install level. If so, that CRL is always used. Otherwise, provided the distribution point is allowed (see [allow-listing](#crl-distribution-point-allow-listing)), Traefik downloads the CRL directly from that URL over HTTP(S), on first use, and caches it in memory for the lifetime of the corresponding TLS Options. If a Distribution points fails, Traefik will make another download attempt at least after `clientAuth.expiry.crl.http.errorBackoff`.

Every downloaded CRL goes through the following integrity checks before being trusted:

- **CRL size**: the CRL file size must not be over `clientAuth.expiry.crl.http.maxCRLBytes`.
- **Signature verification**: the CRL must be signed by the same certificate that issued the client certificate being checked, and the CRL's declared issuer must match that issuer's subject.
- **Freshness window**: the CRL's `thisUpdate` must not be in the future, and its `nextUpdate` (when set) must not already be in the past.
- **Anti-rollback protection**: the CRL's `Number` must never be lower than the previously cached number for the same distribution point, preventing an older (and potentially stale or malicious) CRL from silently replacing a newer one.

A CRL failing any of these checks is rejected, and the previous cached data (if any) is kept or discarded depending on the configured [expiration strategy](#crl-expiration-strategies).

!!! info "Cache lifetime across configuration reloads"

    The HTTP CRL cache for a given TLS Options is preserved across dynamic configuration reloads, as long as the TLS Options itself still exists. Changing the CRL `mode` or `whitelist` only rebuilds the validation logic, without discarding the cache. Changing `http.timeout` only swaps the underlying HTTP client. This avoids unnecessary network calls and latency spikes on every configuration reload.

##### CRL Expiration Strategies

`clientAuth.expiry.crl.http.expirationStrategy` controls how Traefik behaves when an HTTP fetched CRL needs to be refreshed (because it is stale, close to or past its `nextUpdate`, or not yet loaded):

| Strategy | Description |
| -------- | ----------- |
| <a id="opt-Open" href="#opt-Open" title="#opt-Open">`Open`</a> | The first request needing a refresh triggers it; while that refresh is in progress, concurrent requests keep using the last known (possibly stale) CRL data. This favors availability and low latency over strict freshness. This is the default. |
| <a id="opt-FailedClosed" href="#opt-FailedClosed" title="#opt-FailedClosed">`FailedClosed`</a> | All requests needing a refresh wait for it to complete (expect a latency burst on refresh). If the refresh fails, the request is rejected rather than falling back to stale CRL data. This favors strict freshness over availability. |

`clientAuth.expiry.crl.http.timeout` sets the HTTP client timeout used when downloading a CRL for this TLS Options. If left unset, a Traefik-wide default timeout applies.

##### CRL Distribution Point Allow-listing

By default (`whitelist.enabled: false`), Traefik fetches CRLs from whichever distribution point URL is found in the presented certificate. Since this value comes from client-supplied certificates, enabling `clientAuth.expiry.crl.http.whitelist.enabled` and restricting `clientAuth.expiry.crl.http.whitelist.distributionPoints` lets you limit the URLs Traefik is allowed to contact for CRL retrieval.

Distribution points resolved through the [global, file based store](../../../install-configuration/tls/crl.md) are never subject to the allow-list, since they are fully operator-controlled.

```yaml tab="Structured (YAML)"
# Dynamic configuration

tls:
  options:
    default:
      clientAuth:
        caFiles:
          - tests/clientca1.crt
        clientAuthType: RequireAndVerifyClientCertWithExpiry
        expiry:
          crl:
            mode: Strict
            reloadInterval: 1h
            http:
              expirationStrategy: FailedClosed
              timeout: 5s
              errorBackoff: 30s
              maxCRLBytes: 1024 # 1KB
              whitelist:
                enabled: true
                distributionPoints:
                  - http://ca.example.com/intermediate.crl

```

```yaml tab="Structured (TOML)"
# Dynamic configuration

[tls.options]
  [tls.options.default]
    [tls.options.default.clientAuth]
      caFiles = ["tests/clientca1.crt"]
      clientAuthType = "RequireAndVerifyClientCertWithExpiry"
      [tls.options.default.clientAuth.expiry.crl]
        mode = "Strict"
        reloadInterval = "1h"
        [tls.options.default.clientAuth.expiry.crl.http]
          expirationStrategy = "FailedClosed"
          timeout = "5s"
          errorBackoff = "30s"
          maxCRLBytes = 1024
          [tls.options.default.clientAuth.expiry.crl.http.whitelist]
            enabled = true
            distributionPoints = ["http://ca.example.com/intermediate.crl"]
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

{% include-markdown "includes/traefik-for-business-applications.md" %}
