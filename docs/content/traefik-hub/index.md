# Traefik Hub

## Overview

Once the Traefik Hub feature is enabled in Traefik,
Traefik and its local agent communicate together.

This agent can:

* get the Traefik metrics to display them in the Traefik Hub UI
* secure the Traefik routers
* provide ACME certificates to Traefik
* transfer requests from the SaaS Platform to Traefik (and then avoid the users to expose directly their infrastructure on the internet)

!!! warning "Traefik Hub Entrypoints"

    When the Traefik Hub feature is enabled, Traefik exposes some services meant for the Traefik Hub Agent on dedicated entrypoints (on ports `9900` and `9901` by default).
    Given their sensitive nature, those services should not be publicly exposed.
    Also those dedicated entrypoints, regardless of how they are created (default, or user-defined), should not be used by anything other than the Hub Agent.

!!! important "Learn More About Traefik Hub"

    This section is intended only as a brief overview for Traefik users who are not familiar with Traefik Hub.
    To explore all that Traefik Hub has to offer, please consult the [Traefik Hub Documentation](https://doc.traefik.io/traefik-hub).

!!! Note "Prerequisites"

    * Traefik Hub is compatible with Traefik Proxy 2.7 or later.
    * The Traefik Hub Agent must be installed to connect to the Traefik Hub platform.

!!! information "Configuration Discovery"

    According to installation options, the Traefik Hub Agent listens to the Docker or Kubernetes API to discover containers/services.

    It doesn't support the routers discovered by Traefik Proxy using other providers, e.g., using the File provider.

!!! example "Minimal Static Configuration to Activate Traefik Hub for Docker"

    ```yaml tab="File (YAML)"
    hub:
      tls:
        insecure: true

    metrics:
      prometheus:
        addRoutersLabels: true
    ```

    ```toml tab="File (TOML)"
    [hub]
      [hub.tls]
        insecure = true

    [metrics]
      [metrics.prometheus]
        addRoutersLabels = true
    ```

    ```bash tab="CLI"
    --hub.tls.insecure
    --metrics.prometheus.addrouterslabels
    ```

!!! example "Minimal Static Configuration to Activate Traefik Hub for Kubernetes"

    ```yaml tab="File (YAML)"
    hub: {}

    metrics:
      prometheus:
        addRoutersLabels: true
    ```

    ```toml tab="File (TOML)"
    [hub]

    [metrics]
      [metrics.prometheus]
        addRoutersLabels = true
    ```

    ```bash tab="CLI"
    --hub
    --metrics.prometheus.addrouterslabels
    ```

## Configuration

### Entrypoints

#### `traefikhub-api`

This entrypoint is used to communicate between the Hub agent and Traefik. 
It allows the Hub agent to create routing.

This dedicated Traefik Hub entryPoint should not be used by anything other than Traefik Hub.

The default port is `9900`.
To change the port, you have to define an entrypoint named `traefikhub-api`.

```yaml tab="File (YAML)"
entryPoints:
  traefikhub-api: ":8000"
```

```toml tab="File (TOML)"
[entryPoints.traefikhub-api]
  address = ":8000"
```

```bash tab="CLI"
--entrypoints.traefikhub-api.address=:8000
```

#### `traefikhub-tunl`

This entrypoint is used to communicate between Traefik Hub and Traefik.
It allows to create secured tunnels.

This dedicated Traefik Hub entryPoint should not be used by anything other than Traefik Hub.

The default port is `9901`.
To change the port, you have to define an entrypoint named `traefikhub-tunl`.

```yaml tab="File (YAML)"
entryPoints:
  traefikhub-tunl: ":8000"
```

```toml tab="File (TOML)"
[entryPoints.traefikhub-tunl]
  address = ":8000"
```

```bash tab="CLI"
--entrypoints.traefikhub-tunl.address=:8000
```

### `tls`

_Optional, Default=None_

This section is required when using the Hub agent for Docker.

This section allows configuring mutual TLS connection between Traefik Proxy and the Traefik Hub Agent.
The key and the certificate are the credentials for Traefik Proxy as a TLS client.
The certificate authority authenticates the Traefik Hub Agent certificate.

!!! note "Certificate Domain"

    The certificate must be valid for the `proxy.traefik` domain.

!!! note "Certificates Definition"

    Certificates can be defined either by their content or their path.

!!! note "Insecure Mode"

    The `insecure` option is mutually exclusive with any other option.

```yaml tab="File (YAML)"
hub:
  tls:
    ca: /path/to/ca.pem
    cert: /path/to/cert.pem
    key: /path/to/key.pem
```

```toml tab="File (TOML)"
[hub.tls]
  ca= "/path/to/ca.pem"
  cert= "/path/to/cert.pem"
  key= "/path/to/key.pem"
```

```bash tab="CLI"
--hub.tls.ca=/path/to/ca.pem
--hub.tls.cert=/path/to/cert.pem
--hub.tls.key=/path/to/key.pem
```

### `tls.ca`

The certificate authority authenticates the Traefik Hub Agent certificate.

```yaml tab="File (YAML)"
hub:
  tls:
    ca: |-
      -----BEGIN CERTIFICATE-----
      MIIBcjCCARegAwIBAgIQaewCzGdRz5iNnjAiEoO5AzAKBggqhkjOPQQDAjASMRAw
      DgYDVQQKEwdBY21lIENvMCAXDTIyMDMyMTE2MTY0NFoYDzIxMjIwMjI1MTYxNjQ0
      WjASMRAwDgYDVQQKEwdBY21lIENvMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
      ZaKYPj2G8Hnmju6jbHt+vODwKqNDVQMH5nxhtAgSUZS61mLWwZvvUhIYLNPwHz8a
      x8C7+cuihEC6Tzvn8DeGeKNNMEswDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoG
      CCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwFgYDVR0RBA8wDYILZXhhbXBsZS5jb20w
      CgYIKoZIzj0EAwIDSQAwRgIhAO8sucDGY+JOrNgQg1a9ZqqYvbxPFnYsSZr7F/vz
      aUX2AiEAilZ+M5eX4RiMFc3nlm9qVs1LZhV3dZW/u80/mPQ/oaY=
      -----END CERTIFICATE-----
```

```toml tab="File (TOML)"
[hub.tls]
  ca = """-----BEGIN CERTIFICATE-----
MIIBcjCCARegAwIBAgIQaewCzGdRz5iNnjAiEoO5AzAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMCAXDTIyMDMyMTE2MTY0NFoYDzIxMjIwMjI1MTYxNjQ0
WjASMRAwDgYDVQQKEwdBY21lIENvMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
ZaKYPj2G8Hnmju6jbHt+vODwKqNDVQMH5nxhtAgSUZS61mLWwZvvUhIYLNPwHz8a
x8C7+cuihEC6Tzvn8DeGeKNNMEswDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoG
CCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwFgYDVR0RBA8wDYILZXhhbXBsZS5jb20w
CgYIKoZIzj0EAwIDSQAwRgIhAO8sucDGY+JOrNgQg1a9ZqqYvbxPFnYsSZr7F/vz
aUX2AiEAilZ+M5eX4RiMFc3nlm9qVs1LZhV3dZW/u80/mPQ/oaY=
-----END CERTIFICATE-----"""
```

```bash tab="CLI"
--hub.tls.ca=-----BEGIN CERTIFICATE-----
MIIBcjCCARegAwIBAgIQaewCzGdRz5iNnjAiEoO5AzAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMCAXDTIyMDMyMTE2MTY0NFoYDzIxMjIwMjI1MTYxNjQ0
WjASMRAwDgYDVQQKEwdBY21lIENvMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
ZaKYPj2G8Hnmju6jbHt+vODwKqNDVQMH5nxhtAgSUZS61mLWwZvvUhIYLNPwHz8a
x8C7+cuihEC6Tzvn8DeGeKNNMEswDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoG
CCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwFgYDVR0RBA8wDYILZXhhbXBsZS5jb20w
CgYIKoZIzj0EAwIDSQAwRgIhAO8sucDGY+JOrNgQg1a9ZqqYvbxPFnYsSZr7F/vz
aUX2AiEAilZ+M5eX4RiMFc3nlm9qVs1LZhV3dZW/u80/mPQ/oaY=
-----END CERTIFICATE-----
```

### `tls.cert`

The TLS certificate for Traefik Proxy as a TLS client.

!!! note "Certificate Domain"

    The certificate must be valid for the `proxy.traefik` domain.

```yaml tab="File (YAML)"
hub:
  tls:
    cert: |-
      -----BEGIN CERTIFICATE-----
      MIIBcjCCARegAwIBAgIQaewCzGdRz5iNnjAiEoO5AzAKBggqhkjOPQQDAjASMRAw
      DgYDVQQKEwdBY21lIENvMCAXDTIyMDMyMTE2MTY0NFoYDzIxMjIwMjI1MTYxNjQ0
      WjASMRAwDgYDVQQKEwdBY21lIENvMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
      ZaKYPj2G8Hnmju6jbHt+vODwKqNDVQMH5nxhtAgSUZS61mLWwZvvUhIYLNPwHz8a
      x8C7+cuihEC6Tzvn8DeGeKNNMEswDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoG
      CCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwFgYDVR0RBA8wDYILZXhhbXBsZS5jb20w
      CgYIKoZIzj0EAwIDSQAwRgIhAO8sucDGY+JOrNgQg1a9ZqqYvbxPFnYsSZr7F/vz
      aUX2AiEAilZ+M5eX4RiMFc3nlm9qVs1LZhV3dZW/u80/mPQ/oaY=
      -----END CERTIFICATE-----
```

```toml tab="File (TOML)"
[hub.tls]
  cert = """-----BEGIN CERTIFICATE-----
MIIBcjCCARegAwIBAgIQaewCzGdRz5iNnjAiEoO5AzAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMCAXDTIyMDMyMTE2MTY0NFoYDzIxMjIwMjI1MTYxNjQ0
WjASMRAwDgYDVQQKEwdBY21lIENvMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
ZaKYPj2G8Hnmju6jbHt+vODwKqNDVQMH5nxhtAgSUZS61mLWwZvvUhIYLNPwHz8a
x8C7+cuihEC6Tzvn8DeGeKNNMEswDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoG
CCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwFgYDVR0RBA8wDYILZXhhbXBsZS5jb20w
CgYIKoZIzj0EAwIDSQAwRgIhAO8sucDGY+JOrNgQg1a9ZqqYvbxPFnYsSZr7F/vz
aUX2AiEAilZ+M5eX4RiMFc3nlm9qVs1LZhV3dZW/u80/mPQ/oaY=
-----END CERTIFICATE-----"""
```

```bash tab="CLI"
--hub.tls.cert=-----BEGIN CERTIFICATE-----
MIIBcjCCARegAwIBAgIQaewCzGdRz5iNnjAiEoO5AzAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMCAXDTIyMDMyMTE2MTY0NFoYDzIxMjIwMjI1MTYxNjQ0
WjASMRAwDgYDVQQKEwdBY21lIENvMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
ZaKYPj2G8Hnmju6jbHt+vODwKqNDVQMH5nxhtAgSUZS61mLWwZvvUhIYLNPwHz8a
x8C7+cuihEC6Tzvn8DeGeKNNMEswDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoG
CCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwFgYDVR0RBA8wDYILZXhhbXBsZS5jb20w
CgYIKoZIzj0EAwIDSQAwRgIhAO8sucDGY+JOrNgQg1a9ZqqYvbxPFnYsSZr7F/vz
aUX2AiEAilZ+M5eX4RiMFc3nlm9qVs1LZhV3dZW/u80/mPQ/oaY=
-----END CERTIFICATE-----
```

### `tls.key`

The TLS key for Traefik Proxy as a TLS client.

```yaml tab="File (YAML)"
hub:
  tls:
    key: |-
      -----BEGIN PRIVATE KEY-----
      MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgm+XJ3LVrTbbirJea
      O+Crj2ADVsVHjMuiyd72VE3lgxihRANCAARlopg+PYbweeaO7qNse3684PAqo0NV
      AwfmfGG0CBJRlLrWYtbBm+9SEhgs0/AfPxrHwLv5y6KEQLpPO+fwN4Z4
      -----END PRIVATE KEY-----
```

```toml tab="File (TOML)"
[hub.tls]
  key = """-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgm+XJ3LVrTbbirJea
O+Crj2ADVsVHjMuiyd72VE3lgxihRANCAARlopg+PYbweeaO7qNse3684PAqo0NV
AwfmfGG0CBJRlLrWYtbBm+9SEhgs0/AfPxrHwLv5y6KEQLpPO+fwN4Z4
-----END PRIVATE KEY-----"""
```

```bash tab="CLI"
--hub.tls.key=-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgm+XJ3LVrTbbirJea
O+Crj2ADVsVHjMuiyd72VE3lgxihRANCAARlopg+PYbweeaO7qNse3684PAqo0NV
AwfmfGG0CBJRlLrWYtbBm+9SEhgs0/AfPxrHwLv5y6KEQLpPO+fwN4Z4
-----END PRIVATE KEY-----
```

### `tls.insecure`

_Optional, Default=false_

Enables an insecure TLS connection that uses default credentials,
and which has no peer authentication between Traefik Proxy and the Traefik Hub Agent.
The `insecure` option is mutually exclusive with any other option.

!!! warning "Security Consideration"

    Do not use this setup in production.
    This option implies sensitive data can be exposed to potential malicious third-party programs.

```yaml tab="File (YAML)"
hub:
  tls:
    insecure: true
```

```toml tab="File (TOML)"
[hub.tls]
  insecure = true
```

```bash tab="CLI"
--hub.tls.insecure=true
```
