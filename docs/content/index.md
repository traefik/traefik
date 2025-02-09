---
title: "Traefik Proxy Documentation"
description: "Traefik Proxy, an open source Edge Router, auto-discovers configurations and supports major orchestrators, like Kubernetes. Read the technical documentation."
---

# Welcome

![Architecture](assets/img/traefik-architecture.png)

Traefik is an [open-source](https://github.com/traefik/traefik) *Application Proxy* that makes publishing your services a fun and easy experience. 
It receives requests on behalf of your system and identifies which components are responsible for handling them, and routes them securely. 

What sets Traefik apart, besides its many features, is that it automatically discovers the right configuration for your services. 
The magic happens when Traefik inspects your infrastructure, where it finds relevant information and discovers which service serves which request. 

Traefik is natively compliant with every major cluster technology, such as Kubernetes, Docker Swarm, AWS, and [the list goes on](providers/overview.md); and can handle many at the same time. (It even works for legacy software running on bare metal.)
 
With Traefik, there is no need to maintain and synchronize a separate configuration file: everything happens automatically, in real time (no restarts, no connection interruptions).
With Traefik, you spend time developing and deploying new features to your system, not on configuring and maintaining its working state.

And if your needs change, you can add API gateway and API management capabilities seamlessly to your existing Traefik deployments. It takes less than a minute, there’s no rip-and-replace, and all your configurations are preserved. See this in action in [our API gateway demo video](https://info.traefik.io/watch-traefik-api-gw-demo?cta=docs).

Developing Traefik, our main goal is to make it effortless to use, and we're sure you'll enjoy it.

-- The Traefik Maintainer Team 

!!! info

    Have a question? Join our [Community Forum](https://community.traefik.io "Link to Traefik Community Forum") to discuss, learn, and connect with the Traefik community.

    Using Traefik OSS in Production? Consider our enterprise-grade [API Gateway](https://info.traefik.io/watch-traefik-api-gw-demo?cta=doc) or our [24/7/365 OSS Support](https://info.traefik.io/request-commercial-support?cta=doc).

    Explore our API Gateway upgrade via [this short demo video](https://info.traefik.io/watch-traefik-api-gw-demo?cta=doc).
