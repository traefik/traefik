---
title: "Traefik Kubernetes Gateway"
description: "The Kubernetes Gateway API can be used as a provider for routing and load balancing in Traefik Proxy. View examples in the technical documentation."
---

# Traefik & Kubernetes with Gateway API

When using the Kubernetes Gateway API provider, Traefik leverages the Gateway API Custom Resource Definitions (CRDs) to obtain its routing configuration. 
For detailed information on the Gateway API concepts and resources, refer to the official [documentation](https://gateway-api.sigs.k8s.io/).

The Kubernetes Gateway API provider supports version [v1.3.0](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v1.3.0) of the specification.

It fully supports all `HTTPRoute` core and some extended features, like `GRPCRoute`, as well as the `TCPRoute` and `TLSRoute` resources from the [Experimental channel](https://gateway-api.sigs.k8s.io/concepts/versioning/?h=#release-channels). 

For more details, check out the conformance [report](https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports/v1.3.0/traefik-traefik).

## Deploying a Gateway

A `Gateway` is a core resource in the Gateway API specification that defines the entry point for traffic into a Kubernetes cluster. 
It is linked to a `GatewayClass`, which specifies the controller responsible for managing and handling the traffic, ensuring that it is directed to the appropriate Kubernetes backend services.

The `GatewayClass` is a cluster-scoped resource typically defined by the infrastructure provider.
The following `GatewayClass` defines that gateways attached to it must be managed by the Traefik controller.

```yaml tab="GatewayClass"
---
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: traefik
spec:
  controllerName: traefik.io/gateway-controller
```

Next, the following `Gateway` manifest configures the running Traefik controller to handle the incoming traffic.

!!! info "Listener ports"

    Please note that `Gateway` listener ports must match the configured [EntryPoint ports](../entrypoints.md) of the Traefik deployment. 
    In case they do not match, an `ERROR` message is logged, and the resource status is updated accordingly.

```yaml tab="Gateway"
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: traefik
  namespace: default
spec:
  gatewayClassName: traefik
  
  # Only Routes from the same namespace are allowed.
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: Same 

    - name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
          - name: secret-tls
            namespace: default

      allowedRoutes:
        namespaces:
          from: Same

    - name: tcp
      protocol: TCP
      port: 3000
      allowedRoutes:
        namespaces:
          from: Same

    - name: tls
      protocol: TLS
      port: 3443
      tls:
        mode: Terminate
        certificateRefs:
          - name: secret-tls
            namespace: default
            
      allowedRoutes:
        namespaces:
          from: Same
```

```yaml tab="Secret"
---
apiVersion: v1
kind: Secret
metadata:
  name: secret-tls
  namespace: default
type: kubernetes.io/tls
data:
  # Self-signed certificate for the whoami.localhost domain.
  tls.crt: |
    LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZWakNDQXo2Z0F3SUJBZ0lVZUUrZG94aTUrMTBMVi9DaUdTMkt2Q1dJR1dZd0RRWUpLb1pJaHZjTkFRRUwKQlFBd1JERUxNQWtHQTFVRUJoTUNSbEl4RFRBTEJnTlZCQWNNQkV4NWIyNHhGVEFUQmdOVkJBb01ERlJ5WVdWbQphV3NnVEdGaWN6RVBNQTBHQTFVRUF3d0dWMmh2WVcxcE1DQVhEVEkwTURjeE1ERTFNRGt3TjFvWUR6SXhNalF3Ck5qRTJNVFV3T1RBM1dqQkVNUXN3Q1FZRFZRUUdFd0pHVWpFTk1Bc0dBMVVFQnd3RVRIbHZiakVWTUJNR0ExVUUKQ2d3TVZISmhaV1pwYXlCTVlXSnpNUTh3RFFZRFZRUUREQVpYYUc5aGJXa3dnZ0lpTUEwR0NTcUdTSWIzRFFFQgpBUVVBQTRJQ0R3QXdnZ0lLQW9JQ0FRQ1pNNm1WNUJkV2V0QzZtdnp0YVBobFNjZ0ljbnd6Z3NsOEFxendEMk05ClJWVkZwRUxBbTh2OTNlRWtMZEY2ZnNkY0FhUXQxWlFDSFdYby9mTHBRNVVrUHh1djZNUCt2NG1KMHY4ZEtGWjcKUjcwaTVud1lCMkVlVkw2RUNZaWlxNmZ6VEtsa3F6U0QvNW93elN3L3pqa0dUYTBJdy92SDlhc0g3NEhqM1d0QQo3RythenZjaVlhQTZxK1dWYlZxNlBIanF6em9obEFuMkh1ano2aERqYWllc3ZMbHdBL0IvcmhEc0FLaCtpMHI4CkFKUTFqM0JiTGJuYkJyWmZqYnBJUjNhYVh5amkwK3RQWENnVTkzQmU5dm1LZTZTY0dSNy82T25tYmtTc0xjZFcKaFpVNzcrL2c4WllsZGhwS01nczc4ekJYUlh4bHlzSThHRUtqU1hud3k5WmZNT2RJNEpBTWNtT0lOaVlBY21abgpJTUJIa0xacFl3aUl0eFdseXVJcWxoZFpkSHNrTFdNSjBSTHUxRmhTMnBFV0NOZTM3VDBiOWhtNFg0OXJ1QWJ6Ckl1M01xSmczcXFIdGRMNGRkZ1JZRHNjMXd0cDJrR2dBZGxDaXJIclF6K1l4MEJNT1ZsZEczaG1SUUh5ZHEySHIKWW0xeEFDNWpMZ3FvaVZhY09wd0xKY21PcGsrZWVNQkNZNVo0ekNYN1hXeXdhVmNtMnN2aGlPMThCZFIraDloWQpiMkRNZDFCendDbE95endQcUlvQy9uNGRURG96Ry9GT3NySzgvNEZ4dzY2Q1ZmM3E4MzBNUHdSd2xDSzFDQjdGCjNQK3lKWkpPelRuK05QZ2dGQW9NaGZUYXBQWTFhUGlWajBzVG9vQjBaOGNFV1RkTnJxQU5tUGs5aDNoQjJwbjgKSndJREFRQUJvejR3UERBYkJnTlZIUkVFRkRBU2doQjNhRzloYldrdWJHOWpZV3hvYjNOME1CMEdBMVVkRGdRVwpCQlNGSjF4N01xdG9zQ3UwQmFWbEs1U054K2djampBTkJna3Foa2lHOXcwQkFRc0ZBQU9DQWdFQVdCOVc1eWkzClRpTWpmSThhSCtMZW1wZjc4clhyeWJ6UXJvSXdEazhqQXhnc3Nrc2V2ZEtIaXJIZGJMZ0RoS2krbkJLeEQ5S2QKNWM4RS9VL1VHWUhxaUowTVUzYkpoeTVNM3oyaklKd1hFa3FuVVhRd0dBNzVyU0QxWVBkOTlWeVpuNEJVRlEwdwpCT3loOU5DS3Z3ZTgycUVlOWZmeU5iem5JUEMrNS9pekhaYlNQMEpwRzdtNFQ5TXljdHV1OTlsaVhmSVlCMU1PCkRFRUdpamxhZ3JvdTliVlpsNmovR2xCaVZpU0JVQXhaRlNqdFErV2RFODJaZlFRUFVWdXQrUEY0SEl0N1dmYlgKaUpZbjRsMytJSVczNStvbUZ5QjR5WUJNdU9SVWRsZ3V5N1ZieEU5OTdPdHYzTnpDOGJYcGtaQVM0TkVzQVVFdwpJZ3lOcTFCdExsb3dZdjZXY05HbkJ5RE1NRUMzdUYzNEcxQkJCTzFDRHUrYXBVdW5NbVhUWmU5WlkrbXh4U2Z2CnBZclhHTHBoT2t4ZitQalpMbEpqQVFlcTNxblMvWWtLQmtYQi9zb282ZVVLTTlybyt5RTVMbnFrV20wZXpQWmwKc2Z5NGpqZ0lJUHlUMHhhZ0YyWExzUSs0M3N4aDRYTEhmc3Z3Zis2QnJVK2trTnoydmc2M3duLzJDQUNVVms3bgphSDdwZzZyZGt4T2pOTDJjUGd6ZzhWaExXbkVYYjhhUVJlVjY1WnlRc0xta21oOXBlSFRpYXBUb2xWa0d6TDIwCm9pdExZc3ZUcnhUR2NRd3Jpd3FaT1I3WjEvVEJLVnVoYnp0emxlRjFHRk9LdE52UmNSREVBeWVlbnJDRzRRMmgKMnFNNFh1RFFKcjJrR095OEV0dnlYTitENkNZUkg0ck5vZUk9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K

  tls.key: |
    LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUpRUUlCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQ1Nzd2dna25BZ0VBQW9JQ0FRQ1pNNm1WNUJkV2V0QzYKbXZ6dGFQaGxTY2dJY253emdzbDhBcXp3RDJNOVJWVkZwRUxBbTh2OTNlRWtMZEY2ZnNkY0FhUXQxWlFDSFdYbwovZkxwUTVVa1B4dXY2TVArdjRtSjB2OGRLRlo3UjcwaTVud1lCMkVlVkw2RUNZaWlxNmZ6VEtsa3F6U0QvNW93CnpTdy96amtHVGEwSXcvdkg5YXNINzRIajNXdEE3RythenZjaVlhQTZxK1dWYlZxNlBIanF6em9obEFuMkh1anoKNmhEamFpZXN2TGx3QS9CL3JoRHNBS2graTByOEFKUTFqM0JiTGJuYkJyWmZqYnBJUjNhYVh5amkwK3RQWENnVQo5M0JlOXZtS2U2U2NHUjcvNk9ubWJrU3NMY2RXaFpVNzcrL2c4WllsZGhwS01nczc4ekJYUlh4bHlzSThHRUtqClNYbnd5OVpmTU9kSTRKQU1jbU9JTmlZQWNtWm5JTUJIa0xacFl3aUl0eFdseXVJcWxoZFpkSHNrTFdNSjBSTHUKMUZoUzJwRVdDTmUzN1QwYjlobTRYNDlydUFiekl1M01xSmczcXFIdGRMNGRkZ1JZRHNjMXd0cDJrR2dBZGxDaQpySHJReitZeDBCTU9WbGRHM2htUlFIeWRxMkhyWW0xeEFDNWpMZ3FvaVZhY09wd0xKY21PcGsrZWVNQkNZNVo0CnpDWDdYV3l3YVZjbTJzdmhpTzE4QmRSK2g5aFliMkRNZDFCendDbE95endQcUlvQy9uNGRURG96Ry9GT3NySzgKLzRGeHc2NkNWZjNxODMwTVB3UndsQ0sxQ0I3RjNQK3lKWkpPelRuK05QZ2dGQW9NaGZUYXBQWTFhUGlWajBzVApvb0IwWjhjRVdUZE5ycUFObVBrOWgzaEIycG44SndJREFRQUJBb0lDQUVCa2dKRXA3ODAvamVBQktQSTR2cjhFCkJmblc5UEZKdFpwVUhaQkJSM3NIVzFJTU9xcHVVWTJBNXhLbjEzWmZOemdxMEhFYlpqeUZVc0pkaXU0VW8razYKUlU3b3pRaVVSU0VTK0h1dTZycWlhcEx5d1pIditCZ2hrbm80NzU4Lyt6VytNU3pJOFNmU0ZXTVJ1ZG1QdWxRMQo3ZGJUV1U2d3FaU0tUTlFUeXZMYzdnUHBuZUpybWtkTzNRNnppZ0RoVGdtVDFHRXNzZ3NxN3NzbXhMWnhkZithCnkyNlRtVkJ4UDFlUzV6OVpHTWxYRFBSK044RjdOTFVrMng3S21WT3NCZVBZdjN5bmlpNHZGQUhNQndWRFZadXAKWUlUajRpMjZIaVhtanlLM2t5T0F2anNWSElRMXh1QTBCZFROdC84WXRtYllJL005QitydVg0UDJiRFNUMktRKwo4TlN2Uk9wbVppcnBHZkY3bExMSGpJUjlTMFhCWDd6VDRoWnBRWnpqK3NEWnhDM2Y3TGIwRFlKYkp0TmlDYTQxCmNpTjhNUlNldzNneHZ0RVk0RzdnN3hjbkJNdjdNT3RwQTE3d2gvMHdLd0h0amYzSWh2TmIzdkZwT0k5d1FqSzYKSlRQMng4bENJV0tyalpObVN0UksreHJTN3hTOUZVdnBhSVlyclRLQkZWSmcyMURCYWI1L3hqRlBlQWxXejczSwpvVkhsa0hLdXNMSjZLczZzcXo0ZG9mbzg3dkFsUFJzTXRkZ1ZnZFIzNXhLTGtEWXNIbGxML3Z5dE9oSkNieXB5CkJqQm1TR0RMdzBDdWplaHVtU2czYjdSUGVTNk5rbHNqUEIrMVpzRjhpVGdCcjMyM1hvTmNha2dhWWVYQlg4NFAKaE1WZHUxWk1rbXJZMWhXTzEydnhBb0lCQVFEWU5Vb2xCMkhsdWlDcVFqdmU4UFNhV0YxRmhwNUlOMGJIZEppeApIdkhqMkplVHJ6V1pUZFlIdFNJR2RzalhTOTBESXo2bXJhMW9YZXFRYlgrODVlOUFQQkZnRTJmNU5uTzBCSVVJCk5hMXRiVGpIOUhjRGRzQmJKUkZwYnk1ODZUb3lhdFY3bS9zcjVpQUZlZFMyenFOTm1XUmZOdllZS2xselZoSEUKdUd4ZjZxMHJTWktVQUhja2s1bU5Yais3WFhZaTgyemErVEE3ZjBDVm5OamR3OXFpd3B2aTJKTFB2SnA0bWt6KwpyMEN1RW9yV2NhMUdTL2hTVWdXemw3NzhQdlRpZFI2RW4zMDB2ZlIyTE84aG1xRjhVL3Bpb2UrTDVjSllRNnNKCk1YMngrYThzWFFpZ3dwdG02aUNxQ2FuS3ExN3NUZS9RTmQ1czdwb3ZOaHVKOHd3dEFvSUJBUUMxWmM5ZktPUFgKVzZSN1VoaHRRcmhIc3htQnRIQ1BuNWRLbjN5MElNNWRBaEFSdFZDV1U0M0hOTHpCNW1LbjU2dnZrSVNFaXdBbAoxTGhuY2I3YXQ2cHpTSlZtMm5oTDhjeUZrdGQrNzVyL1FHNFlpNnZQbHNBODV5ZXZpZDhZcWswaHdaZXExY05xCmxETUN3NWsrV0drckM2VW5jZXNIc2FWbDJTUGdZV1c1L3I1NnVxUnBsaVFka1EyZmlEYWRyblVueU8ycHg3bFMKVG1HemZaNmtzTWh2MlZFR1NPRm05aUo0dWlPb0xyZVhoU1RQRmxTdjdZUTZSWWtSaGgwT0tqdXM5bXZacEIxWApjcytYN0UyVTNlM0RZelpCR0NFdmxxaFNWTFRScjdIN21pMWxUMEozR0RzbDdiUk9xOE50WVdQa3hhSlNCUnQ5Ck9TcTlkTm9CcGRvakFvSUJBQ2lQdnN3NW1WVW0yUS80QXhGdE5RWnJ3M3ZTcUlrMXpaS0h2a21rVzQ3NlNGMk4KaGttdmY1TE1tWWlLNmx6eHY1SGlIOVBYUzJ3RUNvaHo4bjMyeVM3TTFobW5LbDlucHNkRC9jMHZmTXpGcTl4ZgpjYUIxdTlxZGxxbW9FUm1nQzZuL3Z2TkVyUmRzUWQrbEhwSDVMRXZYbGl3Q3ZLS0Y5MmdhNHBSOFlPQ1J2MUVhCnFXUVl2a0ZmYTNSSkZUM0taK3BncnJCYUJZRnoreUxXWFIwbHJEUFN2TG9QRldQaHB6MHUvWGplV2cwT0wzdlIKc2NjNVkybldOM21jNDFpaFd3SE5KUitPYUVmbnh5QVFpQUJPNlRMUThtMWtvZk1sOUpMb2h3TGZoUXhKb21KNQpSYkFiTWxwWlhDMXFTSzliL1IvcDh5NmxuSWZsTDRuaDVjSzRsVFVDZ2dFQUpSSHVSQU1tTksrTXVJcjVaUEs2Cm1DUjR0UEg4QXMzWmJDMlZuWFlLMWlVQ3hhdXBFVjkzM05yaExEcjV0Rmg2NFpWR0Q1UWNicDYvSkp5eEpSOWQKblB1YlZJNlhBT1lrSnJQd2lBZE5SSmFWS1R6NTJvMXpNYjhIZEM4WHdZR2tDNTcxY0xzSW1YSTV6bm5NaWxvaworK0FBVzBSRGhLb0FKQVV3K0x6T3ZpamFJbGljR3R2TSs2SFdCK0VkVURJRHpTS1p0eFdTd01nMTNTbHh6elExCmNlNFdTZE9CQkxxT0p0L2JRNVp3ZkcyQUxUWGlEcVhhWE5JekJickRtMDUwTFkrYVVMcmlLQ25WVkxXODBReGQKZDQyQjIrR2pmb2NxVk5Ec3R1RlIzUm9QNXVGQXN2Zm50b09TVW5WMWxaZk9nMFVFUEFEQk1tRUpZL2hLU1FYcwp3d0tDQVFBNWQya2hFQ1c1V3QrMzRYWnl1b3NFTjZ4UDExbC96VDRBZjhGSWtQLzlkb0JXRnBhc28zbG1NcXZHCmhPeFErbnZBSjFhNzhZRjA4N3p1UC9DZkQ0UElOUTV4YzZHMDNQdG5JOVNVT0dpMDB4Zlg5MU5NMHBHYWJqb0QKZ0RqVzJxSkJDaVB5N0RIR1RlZkU5eUNUbkhrY1NBbWllVGc3aGFyeEZPOUREZTJKbzhKQXV2SHI1aGVxazVIcgpLYlgzTy9vNUMwcWVnYW1rWVRLcHZzV2VFdXhkY2l5LzFQd3NnV3BuV1JPWllQNENrSkJweEx1bDNVamVSY3dkCnRhcjBJYU52WlV2NFd4U0JZdWVHMDFyYUd2SDZtTTcyTEExR3MrMytwTnZwUVo3bGo2S09tcFlhQUlhemVxY2MKTjJjT2R5U1RqZmQ5OFlNVFAxbmIyK3N1Yy91VAotLS0tLUVORCBQUklWQVRFIEtFWS0tLS0tCg==
```

## Exposing a Route

Once a `Gateway` is deployed (see [Deploying a Gateway](#deploying-a-gateway)) `HTTPRoute`, `TCPRoute`, 
and/or `TLSRoute` resources must be deployed to forward some traffic to Kubernetes backend [services](https://kubernetes.io/docs/concepts/services-networking/service/).

!!! info "Attaching to Gateways"

    As demonstrated in the following examples, a Route resource must be configured with `ParentRefs` that reference the parent `Gateway` it should be associated with.

### HTTP/HTTPS

The `HTTPRoute` is a core resource in the Gateway API specification, designed to define how HTTP traffic should be routed within a Kubernetes cluster. 
It allows the specification of routing rules that direct HTTP requests to the appropriate Kubernetes backend services. 

For more details on the resource and concepts, check out the Kubernetes Gateway API [documentation](https://gateway-api.sigs.k8s.io/api-types/httproute/).

For example, the following manifests configure a whoami backend and its corresponding `HTTPRoute`, 
reachable through the [deployed `Gateway`](#deploying-a-gateway) at the `http://whoami.localhost` address.

```yaml tab="HTTPRoute"
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: whoami-http
  namespace: default
spec:
  parentRefs:
    - name: traefik
      sectionName: http
      kind: Gateway

  hostnames:
    - whoami.localhost

  rules:
     - matches:
        - path:
            type: PathPrefix
            value: /

       backendRefs:
        - name: whoami
          namespace: default
          port: 80
```

```yaml tab="Whoami deployment"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whoami
  namespace: default
spec:
  selector:
    matchLabels:
      app: whoami

  template:
    metadata:
      labels:
        app: whoami
    spec:
      containers:
        - name: whoami
          image: traefik/whoami

---
apiVersion: v1
kind: Service
metadata:
  name: whoami
  namespace: default
spec:
  selector:
    app: whoami

  ports:
    - port: 80
```

To secure the connection with HTTPS and redirect non-secure request to the secure endpoint,
we will update the above `HTTPRoute` manifest to add a `RequestRedirect` filter,
and add a new `HTTPRoute` which binds to the https `Listener` and forward the traffic to the whoami backend.

```yaml tab="HTTRoute (HTTP)"
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: whoami-http
  namespace: default
spec:
  parentRefs:
    - name: traefik
      sectionName: http
      kind: Gateway

  hostnames:
    - whoami.localhost

  rules:
    - filters:
        - type: RequestRedirect
          requestRedirect:
            scheme: https
```

```yaml tab="HTTRoute (HTTPS)"
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: whoami-https
  namespace: default
spec:
  parentRefs:
    - name: traefik
      sectionName: https
      kind: Gateway

  hostnames:
    - whoami.localhost

  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /

      backendRefs:
        - name: whoami
          namespace: default
          port: 80
```

Once everything is deployed, sending a `GET` request to the HTTP and HTTPS endpoints should return the following responses:

```shell
$ curl -I http://whoami.localhost

HTTP/1.1 302 Found
Location: https://whoami.localhost/
Date: Thu, 11 Jul 2024 15:11:31 GMT
Content-Length: 5

$ curl -k https://whoami.localhost
 
Hostname: whoami-697f8c6cbc-2krl7
IP: 127.0.0.1
IP: ::1
IP: 10.42.1.5
IP: fe80::60ed:22ff:fe10:3ced
RemoteAddr: 10.42.2.4:44682
GET / HTTP/1.1
Host: whoami.localhost
User-Agent: curl/7.87.1-DEV
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 10.42.1.0
X-Forwarded-Host: whoami.localhost
X-Forwarded-Port: 443
X-Forwarded-Proto: https
X-Forwarded-Server: traefik-6b66d45748-ns8mt
X-Real-Ip: 10.42.1.0
```

### GRPC

The `GRPCRoute` is an extended resource in the Gateway API specification, designed to define how GRPC traffic should be routed within a Kubernetes cluster. 
It allows the specification of routing rules that direct GRPC requests to the appropriate Kubernetes backend services. 

For more details on the resource and concepts, check out the Kubernetes Gateway API [documentation](https://gateway-api.sigs.k8s.io/api-types/grpcroute/).

For example, the following manifests configure an echo backend and its corresponding `GRPCRoute`, 
reachable through the [deployed `Gateway`](#deploying-a-gateway) at the `echo.localhost:80` address.

```yaml tab="GRPCRoute"
---
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: echo
  namespace: default
spec:
  parentRefs:
    - name: traefik
      sectionName: http
      kind: Gateway

  hostnames:
    - echo.localhost

  rules:
    - matches:
        - method:
            type: Exact
            service: grpc.reflection.v1alpha.ServerReflection

        - method:
            type: Exact
            service: gateway_api_conformance.echo_basic.grpcecho.GrpcEcho
            method: Echo

      backendRefs:
        - name: echo
          namespace: default
          port: 3000
```

```yaml tab="Echo deployment"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echo
  namespace: default
spec:
  selector:
    matchLabels:
      app: echo

  template:
    metadata:
      labels:
        app: echo
    spec:
      containers:
        - name: echo-basic
          image: gcr.io/k8s-staging-gateway-api/echo-basic
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: GRPC_ECHO_SERVER
              value: "1"

---
apiVersion: v1
kind: Service
metadata:
  name: echo
  namespace: default
spec:
  selector:
    app: echo

  ports:
    - port: 3000
```

Once everything is deployed, sending a GRPC request to the HTTP endpoint should return the following responses:

```shell
$ grpcurl -plaintext echo.localhost:80 gateway_api_conformance.echo_basic.grpcecho.GrpcEcho/Echo

{
  "assertions": {
    "fullyQualifiedMethod": "/gateway_api_conformance.echo_basic.grpcecho.GrpcEcho/Echo",
    "headers": [
      {
        "key": "x-real-ip",
        "value": "10.42.2.0"
      },
      {
        "key": "x-forwarded-server",
        "value": "traefik-74b4cf85d8-nkqqf"
      },
      {
        "key": "x-forwarded-port",
        "value": "80"
      },
      {
        "key": "x-forwarded-for",
        "value": "10.42.2.0"
      },
      {
        "key": "grpc-accept-encoding",
        "value": "gzip"
      },
      {
        "key": "user-agent",
        "value": "grpcurl/1.9.1 grpc-go/1.61.0"
      },
      {
        "key": "content-type",
        "value": "application/grpc"
      },
      {
        "key": "x-forwarded-host",
        "value": "echo.localhost:80"
      },
      {
        "key": ":authority",
        "value": "echo.localhost:80"
      },
      {
        "key": "accept-encoding",
        "value": "gzip"
      },
      {
        "key": "x-forwarded-proto",
        "value": "http"
      }
    ],
    "authority": "echo.localhost:80",
    "context": {
      "namespace": "default",
      "pod": "echo-78f76675cf-9k7rf"
    }
  }
}
```

### TCP

!!! info "Experimental Channel"

    The `TCPRoute` resource described below is currently available only in the Experimental channel of the Gateway API specification. 
    To use this resource, the [experimentalChannel](../../providers/kubernetes-gateway.md#experimentalchannel) option must be enabled in the Traefik deployment.

The `TCPRoute` is a resource in the Gateway API specification designed to define how TCP traffic should be routed within a Kubernetes cluster. 

For more details on the resource and concepts, check out the Kubernetes Gateway API [documentation](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TCPRoute).

For example, the following manifests configure a whoami backend and its corresponding `TCPRoute`, 
reachable through the [deployed `Gateway`](#deploying-a-gateway) at the `localhost:3000` address.

```yaml tab="TCPRoute"
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TCPRoute
metadata:
  name: whoami-tcp
  namespace: default
spec:
  parentRefs:
    - name: traefik
      sectionName: tcp
      kind: Gateway

  rules:
     - backendRefs:
        - name: whoamitcp
          namespace: default
          port: 3000
```

```yaml tab="Whoami deployment"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whoamitcp
  namespace: default
spec:
  selector:
    matchLabels:
      app: whoamitcp

  template:
    metadata:
      labels:
        app: whoamitcp
    spec:
      containers:
        - name: whoami
          image: traefik/whoamitcp
          args:
            - --port=:3000

---
apiVersion: v1
kind: Service
metadata:
  name: whoamitcp
  namespace: default
spec:
  selector:
    app: whoamitcp
  ports:
    - port: 3000
```

Once everything is deployed, sending the WHO command should return the following response:

```shell
$ nc localhost 3000

WHO
Hostname: whoamitcp-85d644bfc-ktzv4
IP: 127.0.0.1
IP: ::1
IP: 10.42.1.4
IP: fe80::b89e:85ff:fec2:7d21
```

### TLS

!!! info "Experimental Channel"

    The `TLSRoute` resource described below is currently available only in the Experimental channel of the Gateway API. 
    Therefore, to use this resource, the [experimentalChannel](../../providers/kubernetes-gateway.md#experimentalchannel) option must be enabled.

The `TLSRoute` is a resource in the Gateway API specification designed to define how TLS (Transport Layer Security) traffic should be routed within a Kubernetes cluster. 
It specifies routing rules for TLS connections, directing them to appropriate backend services based on the SNI (Server Name Indication) of the incoming connection.

For more details on the resource and concepts, check out the Kubernetes Gateway API [documentation](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TLSRoute).

For example, the following manifests configure a whoami backend and its corresponding `TLSRoute`, 
reachable through the [deployed `Gateway`](#deploying-a-gateway) at the `localhost:3443` address via a secure connection with the `whoami.localhost` SNI.

```yaml tab="TLSRoute"
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TLSRoute
metadata:
  name: whoami-tls
  namespace: default
spec:
  parentRefs:
    - name: traefik
      sectionName: tls
      kind: Gateway

  hostnames:
    - whoami.localhost

  rules:
    - backendRefs:
        - name: whoamitcp
          namespace: default
          port: 3000

```

```yaml tab="Whoami deployment"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whoamitcp
  namespace: default
spec:
  selector:
    matchLabels:
      app: whoamitcp

  template:
    metadata:
      labels:
        app: whoamitcp
    spec:
      containers:
        - name: whoami
          image: traefik/whoamitcp
          args:
            - --port=:3000

---
apiVersion: v1
kind: Service
metadata:
  name: whoamitcp
  namespace: default
spec:
  selector:
    app: whoamitcp
  ports:
    - port: 3000
```

Once everything is deployed, sending the WHO command should return the following response:

```shell
$ openssl s_client -quiet -connect localhost:3443 -servername whoami.localhost
Connecting to ::1
depth=0 C=FR, L=Lyon, O=Traefik Labs, CN=Whoami
verify error:num=18:self-signed certificate
verify return:1
depth=0 C=FR, L=Lyon, O=Traefik Labs, CN=Whoami
verify return:1

WHO
Hostname: whoamitcp-85d644bfc-hnmdz
IP: 127.0.0.1
IP: ::1
IP: 10.42.2.4
IP: fe80::d873:20ff:fef5:be86
```

## Using Traefik middleware as HTTPRoute filter

An HTTP [filter](https://gateway-api.sigs.k8s.io/api-types/httproute/#filters-optional) is an `HTTPRoute` component which enables the modification of HTTP requests and responses as they traverse the routing infrastructure.

There are three types of filters:

- **Core:** Mandatory filters for every Gateway controller, such as `RequestHeaderModifier` and `RequestRedirect`.
- **Extended:** Optional filters for Gateway controllers, such as `ResponseHeaderModifier` and `RequestMirror`.
- **ExtensionRef:** Additional filters provided by the Gateway controller. In Traefik, these are the [HTTP middlewares](https://doc.traefik.io/traefik/middlewares/http/overview/) supported through the [Middleware CRD](../providers/kubernetes-crd.md#kind-middleware).

!!! info "ExtensionRef Filters"

    To use Traefik middlewares as `ExtensionRef` filters, the Kubernetes IngressRoute provider must be enabled in the static configuration, as detailed in the [documentation](../../providers/kubernetes-crd.md). 

For example, the following manifests configure an `HTTPRoute` using the Traefik `AddPrefix` middleware, 
reachable through the [deployed `Gateway`](#deploying-a-gateway) at the `http://whoami.localhost` address:

```yaml tab="HTTRoute"
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: whoami-http
  namespace: default
spec:
  parentRefs:
    - name: traefik
      sectionName: http
      kind: Gateway

  hostnames:
    - whoami.localhost

  rules:
    - backendRefs:
        - name: whoami
          namespace: default
          port: 80

      filters:
        - type: ExtensionRef
          extensionRef:
            group: traefik.io
            kind: Middleware
            name: add-prefix
```

```yaml tab="AddPrefix middleware"
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: add-prefix
  namespace: default
spec:
  addPrefix:
    prefix: /prefix
```

```yaml tab="Whoami deployment"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whoami
  namespace: default
spec:
  selector:
    matchLabels:
      app: whoami

  template:
    metadata:
      labels:
        app: whoami
    spec:
      containers:
        - name: whoami
          image: traefik/whoami

---
apiVersion: v1
kind: Service
metadata:
  name: whoami
  namespace: default
spec:
  selector:
    app: whoami
  ports:
    - port: 80
```

Once everything is deployed, sending a `GET` request should return the following response:

```shell
$ curl http://whoami.localhost
                                                                                                    
Hostname: whoami-697f8c6cbc-kw954
IP: 127.0.0.1
IP: ::1
IP: 10.42.2.6
IP: fe80::a460:ecff:feb6:3a56
RemoteAddr: 10.42.2.4:54758
GET /prefix/ HTTP/1.1
Host: whoami.localhost
User-Agent: curl/7.87.1-DEV
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 10.42.2.1
X-Forwarded-Host: whoami.localhost
X-Forwarded-Port: 80
X-Forwarded-Proto: http
X-Forwarded-Server: traefik-6b66d45748-ns8mt
X-Real-Ip: 10.42.2.1
```

## Native Load Balancing

By default, Traefik sends the traffic directly to the pod IPs and reuses the established connections to the backends for performance purposes.

It is possible to override this behavior and configure Traefik to send the traffic to the service IP.
The Kubernetes service itself does the load balancing to the pods.
It can be done with the annotation `traefik.io/service.nativelb` on the backend `Service`.

By default, NativeLB is `false`.

!!! info "Default value"

    Note that it is possible to override the default value by using the option [`nativeLBByDefault`](../../providers/kubernetes-gateway.md#nativelbbydefault) at the provider level. 

```yaml
---
apiVersion: v1
kind: Service
metadata:
  name: myservice
  namespace: default
  annotations:
    traefik.io/service.nativelb: "true"
spec:
[...]
```

{!traefik-for-business-applications.md!}
