---
title : 'ServersTransportTCP'
description : 'Understand the service routing configuration for the Kubernetes ServerTransportTCP & Traefik CRD'
---

`ServersTransportTCP` is the CRD implementation of [ServersTransportTCP](../../../tcp/serverstransport.md).

Before creating `ServersTransportTCP` objects, you need to apply the [Traefik Kubernetes CRDs](https://doc.traefik.io/traefik/reference/dynamic-configuration/kubernetes-crd/#definitions) to your Kubernetes cluster.

This registers the `ServersTransportTCP` kind and other Traefik-specific resources.

!!! tip "Default serversTransportTCP"
    If no `serversTransportTCP` is specified, the `default@internal` will be used. The `default@internal` `serversTransportTCP` is created from the install configuration (formerly known as static configuration).

!!! note "ServersTransport reference"
    By default, the referenced `ServersTransportTCP` CRD must be defined in the same Kubernetes service namespace.

    To reference a `ServersTransportTCP` CRD from another namespace, the value must be of form `namespace-name@kubernetescrd`, and the `allowCrossNamespace` option must be enabled.

    If the `ServersTransportTCP` CRD is defined in another provider the cross-provider format `name@provider` should be used.

## Configuration Example

```yaml tab="ServersTransportTCP"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
  tls:
    serverName: example.org
    insecureSkipVerify: true
```

## Configuration Options

| Field                                |  Description                    | Default                                   | Required |
|-------------------------------------|-----------------------------|-------------------------------------------|-----------------------|
| `dialTimeout`                         | The amount of time to wait until a connection to a server can be established. If zero, no timeout exists. | 30s | No |
| `dialKeepAlive`                       | The interval between keep-alive probes for an active network connection.<br />If this option is set to zero, keep-alive probes are sent with a default value (currently 15 seconds),<br />if supported by the protocol and operating system. Network protocols or operating systems that do not support keep-alives ignore this field.<br />If negative, keep-alive probes are turned off.| 15s | No |
| `terminationDelay`     | Defines the delay to wait before fully terminating the connection, after one connected peer has closed its writing capability.| 100ms | No |
| `tls.serverName`                      | ServerName used to contact the server. | "" | No |
| `tls.insecureSkipVerify`              | Controls whether the server's certificate chain and host name is verified. | false | No |
| `tls.peerCertURI`                     | Defines the URI used to match against SAN URIs during the server's certificate verification. | "" | No |
| `tls.rootCAsSecrets`                  | Defines the set of root certificate authorities to use when verifying server certificates.<br />The CA secret must contain a base64 encoded certificate under either a `tls.ca` or a `ca.crt` key.| "" | No |
| `tls.certificatesSecrets`             | Certificates to present to the server for mTLS.| "" | No |
| `spiffe`                              | Configures [SPIFFE](../../../../install-configuration/tls/spiffe.md) options. | "" | No |
| `spiffe.ids`                          | Defines the allowed SPIFFE IDs. This takes precedence over the SPIFFE `trustDomain`. |""| No |
| `spiffe.trustDomain`                  | Defines the allowed SPIFFE trust domain. | "" | No |
