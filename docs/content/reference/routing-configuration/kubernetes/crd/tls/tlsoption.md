---
title: "TLSOption"
description: "TLS Options in Traefik Proxy"
---

The TLS options allow you to configure some parameters of the TLS connection in Traefik.

Before creating `TLSOption` objects or referencing TLS options in the [`IngressRoute`](../http/ingressroute.md) / [`IngressRouteTCP`](../tcp/ingressroutetcp.md) objects, you need to apply the [Traefik Kubernetes CRDs](https://doc.traefik.io/traefik/reference/dynamic-configuration/kubernetes-crd/#definitions) to your Kubernetes cluster.

!!! tip "References and namespaces"
    If the optional namespace attribute is not set, the configuration will be applied with the namespace of the `IngressRoute`/`IngressRouteTCP`.

    Additionally, when the definition of the TLS option is from another provider, the cross-provider [syntax](../../../../install-configuration/providers/overview.md#provider-namespace) (`middlewarename@provider`) should be used to refer to the TLS option. Specifying a namespace attribute in this case would not make any sense, and will be ignored.

!!! important "TLSOption in Kubernetes"

    When using the `TLSOption` resource in Kubernetes, one might setup a default set of options that,
    if not explicitly overwritten, should apply to all ingresses.  
    To achieve that, you'll have to create a `TLSOption` resource with the name `default`.
    There may exist only one `TLSOption` with the name `default` (across all namespaces) - otherwise they will be dropped.  
    To explicitly use a different `TLSOption` (and using the Kubernetes Ingress resources)
    you'll have to add an annotation to the Ingress in the following form:
    `traefik.ingress.kubernetes.io/router.tls.options: <resource-namespace>-<resource-name>@kubernetescrd`

    This `default` TLSOption carries the TLS enforcement policy (client authentication, TLS versions, cipher suites, SNI strictness)
    of every router that does not reference a TLSOption explicitly.
    The [`defaultTLSResourcesNamespace`](../../../../install-configuration/providers/kubernetes/kubernetes-crd.md#defaulttlsresourcesnamespace) provider option
    restricts the namespace this cluster-wide default can be defined in.

## Configuration Example

```yaml tab="TLSOption"
apiVersion: traefik.io/v1alpha1
kind: TLSOption
metadata:
  name: mytlsoption
  namespace: default

spec:
  minVersion: VersionTLS12
  sniStrict: true
  cipherSuites:
    - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
    - TLS_RSA_WITH_AES_256_GCM_SHA384
  clientAuth:
    secretNames:
      - secret-ca1
      - secret-ca2
    clientAuthType: VerifyClientCertIfGiven
```

## Configuration Options

| Field                       | Description                                                                                                                                                                                                                                                                                                                                                                                              | Default                    | Required |
|:----------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------|:---------|
| <a id="opt-minVersion" href="#opt-minVersion" title="#opt-minVersion">`minVersion`</a> | Minimum TLS version that is acceptable.                                                                                                                                                                                                                                                                                                                                                                  | "VersionTLS12"             | No       |
| <a id="opt-maxVersion" href="#opt-maxVersion" title="#opt-maxVersion">`maxVersion`</a> | Maximum TLS version that is acceptable.<br />We do not recommend setting this option to disable TLS 1.3.                                                                                                                                                                                                                                                                                                 |                            | No       |
| <a id="opt-cipherSuites" href="#opt-cipherSuites" title="#opt-cipherSuites">`cipherSuites`</a> | List of supported [cipher suites](https://godoc.org/crypto/tls#pkg-constants) for TLS versions up to TLS 1.2.<br />[Cipher suites defined for TLS 1.2 and below cannot be used in TLS 1.3, and vice versa.](https://tools.ietf.org/html/rfc8446)<br />With TLS 1.3, [the cipher suites are not configurable](https://golang.org/doc/go1.12#tls_1_3) (all supported cipher suites are safe in this case). |                            | No       |
| <a id="opt-curvePreferences" href="#opt-curvePreferences" title="#opt-curvePreferences">`curvePreferences`</a> | List of the elliptic curves references that will be used in an ECDHE handshake.<br />Use curves names from [`crypto`](https://godoc.org/crypto/tls#CurveID) or the [RFC](https://tools.ietf.org/html/rfc8446#section-4.2.7).<br />See [CurveID](https://godoc.org/crypto/tls#CurveID) for more information.                                                                                              |                            | No       |
| <a id="opt-clientAuth-secretNames" href="#opt-clientAuth-secretNames" title="#opt-clientAuth-secretNames">`clientAuth.secretNames`</a> | Client Authentication (mTLS) option.<br />List of names of the referenced Kubernetes [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/) (in TLSOption namespace).<br /> The secret must contain a certificate under either a `tls.ca` or a `ca.crt` key.                                                                                                                               |                            | No       |
| <a id="opt-clientAuth-clientAuthType" href="#opt-clientAuth-clientAuthType" title="#opt-clientAuth-clientAuthType">`clientAuth.clientAuthType`</a> | Client Authentication (mTLS) option.<br />Client authentication type to apply. Available values [here](#client-authentication-mtls).                                                                                                                                                                                                                                                                     |                            | No       |
| <a id="opt-sniStrict" href="#opt-sniStrict" title="#opt-sniStrict">`sniStrict`</a> | Allow rejecting connections from clients connections that do not specify a server_name extension.<br />The [default certificate](../../../http/tls/tls-certificates.md#default-certificate) is never served is the option is enabled.                                                                                                                                                                    | false                      | No       |
| <a id="opt-alpnProtocols" href="#opt-alpnProtocols" title="#opt-alpnProtocols">`alpnProtocols`</a> | List of supported application level protocols for the TLS handshake, in order of preference.<br />If the client supports ALPN, the selected protocol will be one from this list, and the connection will fail if there is no mutually supported protocol.                                                                                                                                                | "h2, http/1.1, acme-tls/1" | No       |
| <a id="opt-disableSessionTickets" href="#opt-disableSessionTickets" title="#opt-disableSessionTickets">`disableSessionTickets`</a> | Allow disabling the use of session tickets, forcing every client to perform a full TLS handshake instead of resuming sessions.                                                                                                                                                                                                                                                                           | false                      | No       |

### Client Authentication (mTLS)

The `clientAuth.clientAuthType` option governs the behavior as follows:

- `NoClientCert`: disregards any client certificate.
- `RequestClientCert`: asks for a certificate but proceeds anyway if none is provided.
- `RequireAnyClientCert`: requires a certificate but does not verify if it is signed by a CA listed in `clientAuth.caFiles` or in `clientAuth.secretNames`.
- `VerifyClientCertIfGiven`: if a certificate is provided, verifies if it is signed by a CA listed in `clientAuth.caFiles` or in `clientAuth.secretNames`. Otherwise proceeds without any certificate.
- `RequireAndVerifyClientCert`: requires a certificate, which must be signed by a CA listed in `clientAuth.caFiles` or in `clientAuth.secretNames`.
- `RequireAndVerifyClientCertWithExpiry`: same as `RequireAndVerifyClientCert`, and additionally enforces the client certificate expiry/revocation rules configured in `clientAuth.expiry.crl` (see below).

!!! note "CA Secret"
    The CA secret must contain a base64 encoded certificate under either a `tls.ca` or a `ca.crt` key.

### Default TLS Option

When no TLS options are specified in an `IngressRoute`/`IngressRouteTCP`, the `default` option is used.
The default behavior is summed up in the table below:

| Configuration             | Behavior                                                    |
|:--------------------------|:------------------------------------------------------------|
| <a id="opt-No-default-TLS-Option" href="#opt-No-default-TLS-Option" title="#opt-No-default-TLS-Option">No `default` TLS Option</a> | Default internal set of TLS Options by default.             |
| <a id="opt-One-default-TLS-Option" href="#opt-One-default-TLS-Option" title="#opt-One-default-TLS-Option">One `default` TLS Option</a> | Custom TLS Options applied by default.                      |
| <a id="opt-Many-default-TLS-Option" href="#opt-Many-default-TLS-Option" title="#opt-Many-default-TLS-Option">Many `default` TLS Option</a> | Error log + Default internal set of TLS Options by default. |

### Client Certificate Revocation (CRL)

When `clientAuth.clientAuthType` is set to `RequireAndVerifyClientCertWithExpiry`, Traefik checks every certificate of the verified client chain against the Certificate Revocation List (CRL) published at its `CRL Distribution Points` extension.

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| <a id="opt-clientAuth-expiry-crl-mode" href="#opt-clientAuth-expiry-crl-mode" title="#opt-clientAuth-expiry-crl-mode">`clientAuth.expiry.crl.mode`</a> | Revocation checking mode. One of `NOOP`, `Lax`, `Strict`. `Lax` only enforces checking when a certificate carries a CRL distribution point; `Strict` additionally rejects certificates that do not carry one. | `NOOP` | No |
| <a id="opt-clientAuth-expiry-crl-reloadInterval" href="#opt-clientAuth-expiry-crl-reloadInterval" title="#opt-clientAuth-expiry-crl-reloadInterval">`clientAuth.expiry.crl.reloadInterval`</a> | Interval between periodic refreshes of HTTP fetched CRLs. `0` disables periodic refresh (CRLs are then only reloaded once expired). | | No |
| <a id="opt-clientAuth-expiry-crl-http-expirationStrategy" href="#opt-clientAuth-expiry-crl-http-expirationStrategy" title="#opt-clientAuth-expiry-crl-http-expirationStrategy">`clientAuth.expiry.crl.http.expirationStrategy`</a> | How to handle an expired/not-yet-loaded CRL. `Open`: serve stale data while refreshing in the background. `FailedClosed`: block and fail the request if the refresh fails. | `Open` | No |
| <a id="opt-clientAuth-expiry-crl-http-timeout" href="#opt-clientAuth-expiry-crl-http-timeout" title="#opt-clientAuth-expiry-crl-http-timeout">`clientAuth.expiry.crl.http.timeout`</a> | HTTP timeout used when downloading a CRL from its distribution point. | | No |
| <a id="opt-clientAuth-expiry-crl-http-errorBackoff" href="#opt-clientAuth-expiry-crl-http-errorBackoff" title="#opt-clientAuth-expiry-crl-http-errorBackoff">`clientAuth.expiry.crl.http.errorBackoff`</a> | Duration between CRL refresh attempt on unrespondsive distribution points. | `30s` | No |
| <a id="opt-clientAuth-expiry-crl-http-maxCRLBytes" href="#opt-clientAuth-expiry-crl-http-maxCRLBytes" title="#opt-clientAuth-expiry-crl-http-maxCRLBytes">`clientAuth.expiry.crl.http.maxCRLBytes`</a> | Max CRL file size. | `1024*1024` (1MB) | No |
| <a id="opt-clientAuth-expiry-crl-http-whitelist-enabled" href="#opt-clientAuth-expiry-crl-http-whitelist-enabled" title="#opt-clientAuth-expiry-crl-http-whitelist-enabled">`clientAuth.expiry.crl.http.whitelist.enabled`</a> | Restricts the CRL distribution point URLs Traefik is allowed to fetch over HTTP. | `false` | No |
| <a id="opt-clientAuth-expiry-crl-http-whitelist-distributionPoints" href="#opt-clientAuth-expiry-crl-http-whitelist-distributionPoints" title="#opt-clientAuth-expiry-crl-http-whitelist-distributionPoints">`clientAuth.expiry.crl.http.whitelist.distributionPoints`</a> | Allowed CRL distribution point URLs, used when `whitelist.enabled` is `true`. | | No |

Every downloaded CRL is verified against the certificate's issuer (signature and issuer name), checked for its validity window (`thisUpdate`/`nextUpdate`), and protected against rollback (its `Number` must never decrease compared to the previously cached CRL for the same distribution point).

!!! important "Global, file based CRLs are not configurable via this CRD"

    CRLs loaded from local files are configured once, process-wide, at the Traefik static/install configuration level (`tls.crlFiles` / `tls.crlReloadInterval`), and apply to every TLS Options regardless of the Kubernetes provider. When a certificate's distribution point is covered by that global store, it always takes precedence over the HTTP based mechanism described above, and bypasses the allow-list.

```yaml tab="TLSOption"
apiVersion: traefik.io/v1alpha1
kind: TLSOption
metadata:
  name: mytlsoption
  namespace: default

spec:
  clientAuth:
    secretNames:
      - secret-ca1
    clientAuthType: RequireAndVerifyClientCertWithExpiry
    expiry:
      crl:
        mode: Lax
        reloadInterval: 1h
        http:
          expirationStrategy: Open
          timeout: 5s
          errorBackoff: 30s
          maxCRLBytes: 1024 # 1KB
          whitelist:
            enabled: true
            distributionPoints:
              - http://ca.example.com/intermediate.crl
```
