---
title: "Traefik Product Features Comparison"
description: "Compare features across Traefik Proxy, Traefik Hub API Gateway, Traefik Hub API Management, and Traefik AI Gateway to choose the right solution for your needs."
---

# Traefik Product Features Comparison

The Traefik ecosystem offers multiple products designed to meet different requirements, from basic reverse proxy functionality to comprehensive API management and AI gateway capabilities. This comparison matrix helps you understand the features available in each product and choose the right solution for your use case.

## Product Overview

- **Traefik Proxy** is the open-source application proxy that serves as the foundation for all Traefik products. It provides essential reverse proxy, load balancing, and service discovery capabilities.

- **[Traefik Hub API Gateway](https://traefik.io/solutions/api-gateway/)** builds on Traefik Proxy with enterprise-grade security, distributed features, and advanced access control for cloud-native API gateway scenarios.

- **[Traefik Hub API Management](https://traefik.io/solutions/api-management/)** adds comprehensive API lifecycle management, developer portals, and organizational features for teams managing multiple APIs across environments.

- **[Traefik AI Gateway](https://traefik.io/solutions/ai-gateway/)** transforms any AI endpoint into a managed API with unified access to multiple LLMs, centralized credential management, semantic caching, and comprehensive AI governance features.

## Features Matrix

| Feature | Traefik Proxy | Traefik Hub API Gateway | Traefik Hub API Management | Traefik AI Gateway |
|---------|---------------|------------------------|---------------------------|-------------------|
| **Core Networking** |
| Services Auto-Discovery | ✓ | ✓ | ✓ | ✓ |
| Graceful Configuration Reload | ✓ | ✓ | ✓ | ✓ |
| Websockets, HTTP/2, HTTP/3, TCP, UDP, GRPC | ✓ | ✓ | ✓ | ✓ |
| Real-time Metrics & Distributed Tracing | ✓ | ✓ | ✓ | ✓ |
| Canary Deployments | ✓ | ✓ | ✓ | ✓ |
| Let's Encrypt | ✓ | ✓ | ✓ | ✓ |
| **Plugin Ecosystem** |
| [Plugin Support](https://plugins.traefik.io/plugins) ([Go](https://github.com/traefik/yaegi), [WASM](https://webassembly.org/)) | ✓ | ✓ | ✓ | ✓ |
| **Deployment & Operations** |
| Hybrid cloud, multi-cloud & on-prem compatible | ✓ | ✓ | ✗ | ✓ |
| Per-cluster dashboard | ✓ | ✓ | ✓ | ✓ |
| GitOps-native declarative configuration | ✓ | ✓ | ✓ | ✓ |
| **Authentication & Authorization** |
| JWT Authentication | ✗ | ✓ | ✓  | ✓ |
| OAuth 2.0 Token Introspection Authentication | ✗ | ✓ | ✓  | ✓ |
| OAuth 2.0 Client Credentials Authentication | ✗ | ✓ | ✓ | ✓ |
| OpenID Connect Authentication | ✗ | ✓ | ✓ | ✓ |
| Lightweight Directory Access Protocol (LDAP) | ✗ | ✓ | ✓ | ✓ |
| API Key Authentication | ✗ | ✓ | ✓ | ✓ |
| **Security & Policy** |
| Open Policy Agent | ✗ | ✓ | ✓ | ✓ |
| Native Coraza Web Application Firewall (WAF) | ✗ | ✓ | ✓ | ✓ |
| HashiCorp Vault Integration | ✗ | ✓ | ✓ | ✓ |
| **Distributed Features** |
| Distributed Let's Encrypt | ✗ | ✓ | ✓ | ✓ |
| Distributed Rate Limit | ✗ | ✓ | ✓ | ✓ |
| HTTP Caching | ✗ | ✓ | ✓ | ✓ |
| **Compliance** |
| FIPS 140-2 Compliance | ✗ | ✓ | ✓ | ✓ |
| **API Management** |
| Flexible API grouping and versioning | ✗ | ✗ | ✓ | ✗ |
| API Developer Portal | ✗ | ✗ | ✓ | ✗ |
| OpenAPI Specifications Support | ✗ | ✗ | ✓ | ✗ |
| Multi-cluster dashboard | ✗ | ✗ | ✓ | ✗ |
| Built-in identity provider (or use your own) | ✗ | ✗ | ✓ | ✗ |
| Configuration linter & change impact analysis | ✗ | ✗ | ✓ | ✗ |
| Pre-built Grafana dashboards | ✗ | ✗ | ✓ | ✗ |
| Event correlation for quick incident mitigation | ✗ | ✗ | ✓ | ✗ |
| Traffic debugger  | ✗ | ✓ | ✓ | ✓ |
| **AI Gateway Capabilities** |
| Unified Multi-LLM API Access | ✗ | ✗ | ✗ | ✓ |
| Centralized AI Credential Management | ✗ | ✗ | ✗ | ✓ |
| AI Provider Flexibility (OpenAI, Anthropic, Azure OpenAI, AWS Bedrock, etc.) | ✗ | ✗ | ✗ | ✓ |
| Semantic Caching for AI Responses | ✗ | ✗ | ✗ | ✓ |
| Content Guard & PII Protection | ✗ | ✗ | ✗ | ✓ |
| AI-specific Observability & OpenTelemetry Integration | ✗ | ✗ | ✗ | ✓ |
| Support for Local/Self-hosted LLMs (Ollama, Mistral) | ✗ | ✗ | ✗ | ✓ |
| **Support** |
| Built-In Commercial Support | ✗ | ✓ | ✓ | ✓ |

## Choosing the Right Product

### Start with Traefik Proxy if you need:

- Basic reverse proxy and load balancing
- Service discovery for containerized applications
- Simple TLS termination and Let's Encrypt integration
- Cost-effective solution with community support

### Upgrade to Traefik Hub API Gateway for:

- Enterprise security requirements (JWT, OIDC, LDAP)
- Distributed deployments across multiple clusters
- Advanced rate limiting and caching
- WAF and policy enforcement
- Commercial support

### Choose Traefik Hub API Management when you have:

- Multiple APIs requiring centralized management
- Developer teams needing self-service portals
- Complex API versioning and lifecycle requirements
- Multi-cluster environments requiring unified dashboards
- Compliance and governance needs

### Consider Traefik AI Gateway for:

- Multi-LLM applications requiring unified API access
- Organizations using multiple AI providers (OpenAI, Anthropic, Azure OpenAI, AWS Bedrock, etc.)
- Centralized AI credential and security management
- Cost optimization through semantic caching
- PII protection and content filtering for AI interactions
- Local/self-hosted LLM deployments (Ollama, Mistral)
- Comprehensive AI observability and compliance requirements

## Migration Path

The Traefik ecosystem is designed for seamless upgrades. You can start with Traefik Proxy and add capabilities as your requirements grow:

1. **Traefik Proxy** → **Hub API Gateway**: Add enterprise security and distributed features
2. **Hub API Gateway** → **Hub API Management**: Add API management and governance features  
3. **AI Gateway**: Add AI-specific capabilities for machine learning workloads in both Hub API Gateway and Hub API Management

All products share the same core configuration concepts, making migration straightforward while preserving your existing configurations and operational knowledge.
