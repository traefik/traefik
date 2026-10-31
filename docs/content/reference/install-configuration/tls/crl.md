---
title: "Traefik CRL Documentation"
description: "Learn how to configure Traefik to use CRL. Read the technical documentation."
---

# CRL

Traefik lets you configure a global, file based Certificate Revocation List (CRL) store, shared by every TLS Options defined in the dynamic configuration. This is useful for offline or air-gapped CRLs that should not be fetched over the network.

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| <a id="opt-tls-crlFiles" href="#opt-tls-crlFiles" title="#opt-tls-crlFiles">`tls.crlFiles`</a> | Maps a CRL distribution point URL (as it would appear in a certificate's `CRL Distribution Points` extension) to a local file path containing the corresponding CRL (DER or PEM encoded). | | No |
| <a id="opt-tls-crlReloadInterval" href="#opt-tls-crlReloadInterval" title="#opt-tls-crlReloadInterval">`tls.crlReloadInterval`</a> | Interval between periodic reloads of all the files referenced in `tls.crlFiles`. Setting it to `0` disables the periodic reload; files are still watched for changes on disk. | `1h` | No |

!!! info "File watching"

    In addition to the periodic reload, Traefik watches the configured CRL files on disk and reloads them as soon as a change is detected. Note that, depending on the underlying filesystem, some file change events may not be detected by the watcher.

!!! important "Precedence over HTTP loaded CRLs"

    When a certificate's CRL distribution point is covered by this global store, it always takes precedence over the per-TLS-Options, HTTP based CRL loading mechanism described in the [TLS Options documentation](../../routing/providers/tls-options.md#expiry-validation), and bypasses its allow-list.

!!! important "Expiration strategy"
    File Based CRLs provided by this way will always be used in CRL revocation checks. It is your responsibility to load/reload a valid CRL for Traefik to consume.

```yaml tab="File (YAML)"
## Install configuration
tls:
  crlReloadInterval: 1h
  crlFiles:
    "http://ca.example.com/root.crl": /etc/traefik/crl/root.crl
```

```toml tab="File (TOML)"
## Install configuration
[tls]
  crlReloadInterval = "1h"
  [tls.crlFiles]
    "http://ca.example.com/root.crl" = "/etc/traefik/crl/root.crl"
```
