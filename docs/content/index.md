---
title: "Traefik Proxy Documentation"
description: "Traefik Proxy, an open source Edge Router, auto-discovers configurations and supports major orchestrators, like Kubernetes. Read the technical documentation."
---

# Welcome

![Architecture](assets/img/traefik-architecture.png)

Traefik is an [open-source](https://github.com/traefik/traefik) *Edge Router* that makes publishing your services a fun and easy experience. 
It receives requests on behalf of your system and finds out which components are responsible for handling them. 

What sets Traefik apart, besides its many features, is that it automatically discovers the right configuration for your services. 
The magic happens when Traefik inspects your infrastructure, where it finds relevant information and discovers which service serves which request. 

Traefik is natively compliant with every major cluster technology, such as Kubernetes, Docker, Docker Swarm, AWS, and [the list goes on](providers/overview.md); and can handle many at the same time. (It even works for legacy software running on bare metal.)
 
With Traefik, there is no need to maintain and synchronize a separate configuration file: everything happens automatically, in real time (no restarts, no connection interruptions).
With Traefik, you spend time developing and deploying new features to your system, not on configuring and maintaining its working state.   

Developing Traefik, our main goal is to make it simple to use, and we're sure you'll enjoy it.

-- The Traefik Maintainer Team 

!!! info

    Join our user friendly and active [Community Forum](https://community.traefik.io) to discuss, learn, and connect with the traefik community.
    
    Using Traefik for commercial applications?
    Consider the [Enterprise Edition](https://traefik.io/traefik-enterprise/) of Traefik as your [Kubernetes Ingress](https://traefik.io/solutions/kubernetes-ingress/), 
    your [Docker Swarm Load Balancer](https://traefik.io/solutions/docker-swarm-ingress/), 
    or your [API gateway](https://traefik.io/solutions/api-gateway/). 
    Get started with a [free 30-day trial](https://info.traefik.io/get-traefik-enterprise-free-for-30-days).
