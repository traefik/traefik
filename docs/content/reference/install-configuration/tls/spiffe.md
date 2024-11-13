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

## Configuration Example

Enabling SPIFFE can be done by using a file (YAML or TOML) or CLI arguments.

```yaml tab="File (YAML)"
# Default Servers Transport
serversTransport:
  spiffe:
    ids:
    - spiffe://trust-domain/id1
    - spiffe://trust-domain/id2
    trustDomain: "spiffe://trust-domain"
spiffe:
    workloadAPIAddr: localhost
```

```yaml tab="Override the default Servers Transport"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: apps
spec:
  spiffe:
    ids:
    - spiffe://trust-domain/specific-id1
    trustDomain: "spiffe://trust-domain"
```

## Configuration Option

### Workload API

The `workloadAPIAddr` configuration defines the address of the SPIFFE [Workload API](https://spiffe.io/docs/latest/spiffe-about/spiffe-concepts/#spiffe-workload-api).

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

!!! info "Enabling SPIFFE in ServersTransports"

    Enabling SPIFFE does not imply that backend connections are going to use it automatically.
    Each [ServersTransport](../../../routing/services/index.md#serverstransport_1) or [TCPServersTransport](../../../routing/services/index.md#serverstransport_2), that is meant to be secured with SPIFFE, must explicitly enable it (see [SPIFFE with ServersTransport](../../../routing/services/index.md#spiffe) or [SPIFFE with TCPServersTransport](../../../routing/services/index.md#spiffe_1)).

!!! warning "SPIFFE can cause Traefik to stall"
    When using SPIFFE,
    Traefik will wait for the first SVID to be delivered before starting.
    If Traefik is hanging when waiting on SPIFFE SVID delivery,
    please double check that it is correctly registered as workload in your SPIFFE infrastructure.
