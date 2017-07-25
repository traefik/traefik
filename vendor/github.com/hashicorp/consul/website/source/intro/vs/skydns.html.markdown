---
layout: "intro"
page_title: "Consul vs. SkyDNS"
sidebar_current: "vs-other-skydns"
description: |-
  SkyDNS is a tool designed to provide service discovery. It uses multiple central servers that are strongly-consistent and fault-tolerant. Nodes register services using an HTTP API, and queries can be made over HTTP or DNS to perform discovery.
---

# Consul vs. SkyDNS

SkyDNS is a tool designed to provide service discovery.
It uses multiple central servers that are strongly-consistent and
fault-tolerant. Nodes register services using an HTTP API, and
queries can be made over HTTP or DNS to perform discovery.

Consul is very similar but provides a superset of features. Consul
also relies on multiple central servers to provide strong consistency
and fault tolerance. Nodes can use an HTTP API or use an agent to
register services, and queries are made over HTTP or DNS.

However, the systems differ in many ways. Consul provides a much richer
health checking framework with support for arbitrary checks and
a highly scalable failure detection scheme. SkyDNS relies on naive
heartbeating and TTLs, an approach which has known scalability issues.
Additionally, the heartbeat only provides a limited liveness check
versus the rich health checks that Consul performs.

Multiple datacenters can be supported by using "regions" in SkyDNS;
however, the data is managed and queried from a single cluster. If servers
are split between datacenters, the replication protocol will suffer from
very long commit times. If all the SkyDNS servers are in a central datacenter,
then connectivity issues can cause entire datacenters to lose availability.
Additionally, even without a connectivity issue, query performance will
suffer as requests must always be performed in a remote datacenter.

Consul supports multiple datacenters out of the box, and it purposely
scopes the managed data to be per-datacenter. This means each datacenter
runs an independent cluster of servers. Requests are forwarded to remote
datacenters if necessary; requests for services within a datacenter
never go over the WAN, and connectivity issues between datacenters do not
affect availability within a datacenter. Additionally, the unavailability
of one datacenter does not affect the service discovery of services
in any other datacenter.
