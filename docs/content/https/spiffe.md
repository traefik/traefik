---
title: "Traefik SPIFFE Documentation"
description: "Learn how to configure Traefik to use SPIFFE. Read the technical documentation."
---

# SPIFFE

Secure the backend connection with SPIFFE.
{: .subtitle }

[SPIFFE](https://spiffe.io/docs/latest/spiffe-about/overview/) (Secure Production Identity Framework For Everyone), 
provides a secure identity in the form of a specially crafted X.509 certificate, 
to every workload in an environment.

Traefik is able to connect to the Workload API to obtain an x509-SVID used to secure the connection with SPIFFE enabled backends.

## Workload API

To connect to the SPIFFE [Workload API](https://spiffe.io/docs/latest/spiffe-about/spiffe-concepts/#spiffe-workload-api),
its address have to be configured in the static configuration.

!!! info "Enabling SPIFFE in ServersTransports"

    Enabling SPIFFE does not imply that backend connections are going to use it automatically.
    Each [ServersTransport](../routing/services/index.md#serverstransport_1) that is meant to be secured with SPIFFE must [explicitly](../routing/services/index.md#spiffe) enable it.

```yaml tab="File (YAML)"
## Static configuration
spiffe:
    workloadAPIAddr: localhost
```

```toml tab="File (TOML)"
## Static configuration
[spiffe]
    workloadAPIAddr: localhost
```

```bash tab="CLI"
## Static configuration
--spiffe.workloadAPIAddr=localhost
```
