# Exposing Services with Traefik Proxy

This section guides you through exposing services securely with Traefik Proxy. You'll learn how to route HTTP and HTTPS traffic to your services, add security features, and implement advanced load balancing.

## What You'll Accomplish

Following these guides, you'll learn how to:

- Route HTTP traffic to your services with [Gateway API](../reference/routing-configuration/kubernetes/gateway-api.md) and [IngressRoute](../reference/routing-configuration/kubernetes/crd/http/ingressroute.md)
- Configure routing rules to direct requests
- Enable HTTPS with TLS
- Add security middlewares
- Generate certificates automatically with Let's Encrypt
- Implement sticky sessions for session persistence

## Platform-Specific Guides

For detailed steps tailored to your environment, follow the guide for your platform:

- [Kubernetes](./kubernetes.md)
- [Docker](./docker.md)
- [Docker Swarm](./swarm.md)
