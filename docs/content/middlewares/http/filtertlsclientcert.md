---
title: "Traefik FilterTLSClientCert Documentation"
description: "In Traefik Proxy's HTTP middleware, the FilterTLSClientCert rejects requests whose client certificate Subject or Issuer DN does not match a configured regex. Read the technical documentation."
---

# FilterTLSClientCert

Filtering Requests Based on Client Certificate DN
{: .subtitle }

FilterTLSClientCert rejects requests whose client certificate Subject or Issuer Distinguished Name (DN) does not match the configured regular expressions.

## Configuration Examples

```yaml tab="Docker & Swarm"
# Reject requests whose client certificate Subject does not match the given regex.
labels:
  - "traefik.http.middlewares.test-filtertlsclientcert.filterTLSClientCert.subject=CN=.*\\.example\\.com"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-filtertlsclientcert
spec:
  filterTLSClientCert:
    subject: "CN=.*\\.example\\.com"
    issuer: "CN=My CA"
```

```yaml tab="Consul Catalog"
# Reject requests whose client certificate Subject does not match the given regex.
- "traefik.http.middlewares.test-filtertlsclientcert.filterTLSClientCert.subject=CN=.*\\.example\\.com"
```

```yaml tab="File (YAML)"
# Reject requests whose client certificate Subject does not match the given regex.
http:
  middlewares:
    test-filtertlsclientcert:
      filterTLSClientCert:
        subject: "CN=.*\\.example\\.com"
        issuer: "CN=My CA"
```

```toml tab="File (TOML)"
# Reject requests whose client certificate Subject does not match the given regex.
[http.middlewares]
  [http.middlewares.test-filtertlsclientcert.filterTLSClientCert]
    subject = "CN=.*\\.example\\.com"
    issuer = "CN=My CA"
```

## Configuration Options

### `subject`

_Optional, Default=""_

A regular expression matched against the full Subject DN of the client certificate.

If empty, the Subject is not checked.

```yaml tab="File (YAML)"
http:
  middlewares:
    test-filtertlsclientcert:
      filterTLSClientCert:
        subject: "CN=.*\\.example\\.com"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-filtertlsclientcert.filterTLSClientCert]
    subject = "CN=.*\\.example\\.com"
```

### `issuer`

_Optional, Default=""_

A regular expression matched against the full Issuer DN of the client certificate.

If empty, the Issuer is not checked.

```yaml tab="File (YAML)"
http:
  middlewares:
    test-filtertlsclientcert:
      filterTLSClientCert:
        issuer: "CN=My CA,O=My Org"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-filtertlsclientcert.filterTLSClientCert]
    issuer = "CN=My CA,O=My Org"
```
