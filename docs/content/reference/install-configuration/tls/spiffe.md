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

Traefik is able to connect to the Workload API to obtain an X509-SVID used to secure the connection with SPIFFE enabled backends.

!!! warning "SPIFFE can cause Traefik to stall"
    When using SPIFFE,
    Traefik will wait for the first SVID to be delivered before starting.
    If Traefik is hanging when waiting on SPIFFE SVID delivery,
    please double check that it is correctly registered as workload in your SPIFFE infrastructure.

## Workload API

To enable SPIFFE globally, you need to set up the [static configuration](../../../getting-started/configuration-overview.md#the-static-configuration). The `workloadAPIAddr` option specifies the address of the SPIFFE [Workload API](https://spiffe.io/docs/latest/spiffe-about/spiffe-concepts/#spiffe-workload-api).

```yaml tab="File (YAML)"
## Static configuration.
spiffe:
    workloadAPIAddr: localhost
```

```toml tab="File (TOML)"
## Static configuration
[spiffe]
    workloadAPIAddr: localhost
```

```bash tab="CLI"
## Static configuration.
--spiffe.workloadAPIAddr=localhost
```

## ServersTransport

Enabling SPIFFE does not imply that backend connections are going to use it automatically.
Each [ServersTransport](../../../routing/services/index.md#serverstransport_1) or [TCPServersTransport](../../../routing/services/index.md#serverstransport_2), that is meant to be secured with SPIFFE, must explicitly enable it (see [SPIFFE with ServersTransport](../../../routing/services/index.md#spiffe) or [SPIFFE with TCPServersTransport](../../../routing/services/index.md#spiffe_1)).

### Configuration Example

```yaml tab="File (YAML)"
serversTransport:
  spiffe:
    ids:
    - spiffe://trust-domain/id1
    - spiffe://trust-domain/id2
    trustDomain: "spiffe://trust-domain" 
```

```toml tab="File (TOML)"
[serversTransport.spiffe]
ids = [ "spiffe://trust-domain/id1", "spiffe://trust-domain/id2" ]
trustDomain = "spiffe://trust-domain"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
    spiffe:
      ids:
        - spiffe://trust-domain/id1
        - spiffe://trust-domain/id2
```
