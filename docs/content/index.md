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

And if your needs change, you can add API gateway and API management capabilities seamlessly to your existing Traefik deployments. It takes less than a minute, there’s no rip-and-replace, and all your configurations are preserved. See how it works in this video:

<div style="text-align: center;">
    <iframe src="https://www.youtube.com/embed/zriUO5YPgFg?modestbranding=1&rel=0&controls=1" width="560" height="315" title="Upgrade Traefik Proxy to API Gateway and API Management in Seconds // Traefik Labs" frameborder="0" allowfullscreen></iframe>
</div>

Developing Traefik, our main goal is to make it effortless to use, and we're sure you'll enjoy it.

-- The Traefik Maintainer Team 

!!! info

    Join our user friendly and active [Community Forum](https://community.traefik.io "Link to Traefik Community Forum") to discuss, learn, and connect with the Traefik community.

    Using Traefik OSS in Production? Consider our enterprise-grade [API Gateway](https://traefik.io/traefik-hub-api-gateway/), [API Management](https://traefik.io/traefik-hub/), and [Commercial Support](https://info.traefik.io/request-commercial-support) solutions.
