---
layout: "intro"
page_title: "Consul vs. Serf"
sidebar_current: "vs-other-serf"
description: |-
  Serf is a node discovery and orchestration tool and is the only tool discussed so far that is built on an eventually-consistent gossip model with no centralized servers. It provides a number of features, including group membership, failure detection, event broadcasts, and a query mechanism. However, Serf does not provide any high-level features such as service discovery, health checking or key/value storage. Consul is a complete system providing all of those features.
---

# Consul vs. Serf

[Serf](https://www.serf.io) is a node discovery and orchestration tool and is the only
tool discussed so far that is built on an eventually-consistent gossip model
with no centralized servers. It provides a number of features, including group
membership, failure detection, event broadcasts, and a query mechanism. However,
Serf does not provide any high-level features such as service discovery, health
checking or key/value storage. Consul is a complete system providing all of those
features. 

The internal [gossip protocol](/docs/internals/gossip.html) used within Consul is in
fact powered by the Serf library: Consul leverages the membership and failure detection
features and builds upon them to add service discovery. By contrast, the discovery
feature of Serf is at a node level, while Consul provides a service and node level
abstraction.

The health checking provided by Serf is very low level and only indicates if the
agent is alive. Consul extends this to provide a rich health checking system
that handles liveness in addition to arbitrary host and service-level checks.
Health checks are integrated with a central catalog that operators can easily
query to gain insight into the cluster.

The membership provided by Serf is at a node level, while Consul focuses
on the service level abstraction, mapping single nodes to multiple services.
This can be simulated in Serf using tags, but it is much more limited and does
not provide useful query interfaces. Consul also makes use of a strongly-consistent
catalog while Serf is only eventually-consistent.

In addition to the service level abstraction and improved health checking,
Consul provides a key/value store and support for multiple datacenters.
Serf can run across the WAN but with degraded performance. Consul makes use
of [multiple gossip pools](/docs/internals/architecture.html) so that
the performance of Serf over a LAN can be retained while still using it over
a WAN for linking together multiple datacenters.

Consul is opinionated in its usage while Serf is a more flexible and
general purpose tool. In [CAP](https://en.wikipedia.org/wiki/CAP_theorem) terms,
Consul uses a CP architecture, favoring consistency over availability. Serf is an
AP system and sacrifices consistency for availability. This means Consul cannot
operate if the central servers cannot form a quorum while Serf will continue to
function under almost all circumstances.
