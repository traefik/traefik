---
title: "Traefik Cert Manager Integration"
description: "Learn how to configure Traefik Proxy to use Cert Manager for certificate on your routers. Read the technical documentation."
---

# Cert Manager

Provision TLS Certificate for Traefik Proxy with Cert Manager on Kubernetes
{: .subtitle }

You can configure Traefik Proxy to use Certificates of Cert Manager on Kubernetes.

## Checklist

To obtain a certificate from Cert Manager, you'll need to:

1. Deploy cert manager
2. Configure an Issuer (or a ClusterIssuer)
3. Create a Certificate with this Issuer
4. Deploy Traefik Proxy
5. Use the certificate in an Ingress / IngressRoute / HTTPRoute

## Configuration Example with ACME and HTTP challenge

!!! example "ACME issuer on http challenge"

    ```yaml
    apiVersion: cert-manager.io/v1
    kind: Issuer
    metadata:
      name: acme

    spec:
      acme:
        # Production server is on https://acme-v02.api.letsencrypt.org/directory
        # Use staging by default.
        server: https://acme-staging-v02.api.letsencrypt.org/directory
        privateKeySecretRef:
          name: acme
        solvers:
          - http01:
              ingress:
                ingressClassName: traefik
    ```

!!! example "Certificate with this Issuer"

    ```yaml
    apiVersion: cert-manager.io/v1
    kind: Certificate
    metadata:
      name: whoami
      namespace: traefik
    spec:
      secretName: domain-tls        # <===  Name of secret here
      dnsNames:
        - "domain.example.com"
      issuerRef:
        name: acme
        kind: Issuer
    ```

!!! example "Route with this Certificate"

    ```yaml tab="Ingress"
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: domain
      annotations:
        traefik.ingress.kubernetes.io/router.entrypoints: websecure

    spec:
      rules:
      - host: domain.example.com
        http:
          paths:
          - path: /
            pathType: Exact
            backend:
              service:
                name:  domain-service
                port:
                  number: 80
      tls:
      - secretName: domain-tls # <=== Use name defined in Certificate resource
    ```

    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: domain

    spec:
      entryPoints:
        - websecure

      routes:
      - match: Host(`domain.example.com`)
        kind: Rule
        services:
        - name: domain-service
          port: 80
      tls:
        secretName: domain-tls    # <=== Use name defined in Certificate resource
    ```

    ```yaml tab="HTTPRoute"
    ---
    apiVersion: gateway.networking.k8s.io/v1
    kind: Gateway
    metadata:
      name: domain-gateway
    spec:
      gatewayClassName: traefik
      listeners:
        - name: websecure
          port: 8443
          protocol: HTTPS
          hostname: domain.example.com
          tls:
            certificateRefs:
              - name: domain-tls  # <==== Use name defined in Certificate resource
    ---
    apiVersion: gateway.networking.k8s.io/v1
    kind: HTTPRoute
    metadata:
      name: domain
    spec:
      parentRefs:
        - name: domain-gateway
      hostnames:
        - domain.example.com
      rules:
        - matches:
            - path:
                type: Exact
                value: /

          backendRefs:
            - name: domain-service
              port: 80
              weight: 1
    ```

## Troubleshooting

There are multiples sources available to investigate when using Cert Manager:

1. `Certificate` and `CertificateRequest` kubernetes events
2. logs of Cert Manager
3. dashboard and/or logs of Traefik Proxy

{!traefik-for-business-applications.md!}
