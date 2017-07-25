---
layout: "docs"
page_title: "Consul Architecture"
sidebar_current: "docs-internals-architecture"
description: |-
  Consul is a complex system that has many different moving parts. To help users and developers of Consul form a mental model of how it works, this page documents the system architecture.
---

# Consul Architecture

Consul is a complex system that has many different moving parts. To help
users and developers of Consul form a mental model of how it works, this
page documents the system architecture.

~> **Advanced Topic!** This page covers technical details of
the internals of Consul. You don't need to know these details to effectively
operate and use Consul. These details are documented here for those who wish
to learn about them without having to go spelunking through the source code.

## Glossary

Before describing the architecture, we provide a glossary of terms to help
clarify what is being discussed:

* Agent - An agent is the long running daemon on every member of the Consul cluster.
It is started by running `consul agent`. The agent is able to run in either *client*
or *server* mode. Since all nodes must be running an agent, it is simpler to refer to
the node as being either a client or server, but there are other instances of the agent. All
agents can run the DNS or HTTP interfaces, and are responsible for running checks and
keeping services in sync.

* Client - A client is an agent that forwards all RPCs to a server. The client is relatively
stateless. The only background activity a client performs is taking part in the LAN gossip
pool. This has a minimal resource overhead and consumes only a small amount of network
bandwidth.

* Server - A server is an agent with an expanded set of responsibilities including
participating in the Raft quorum, maintaining cluster state, responding to RPC queries,
exchanging WAN gossip with other datacenters, and forwarding queries to leaders or
remote datacenters.

* Datacenter - While the definition of a datacenter seems obvious, there are subtle details
that must be considered. For example, in EC2, are multiple availability zones considered
to comprise a single datacenter? We define a datacenter to be a networking environment that is
private, low latency, and high bandwidth. This excludes communication that would traverse
the public internet, but for our purposes multiple availability zones within a single EC2
region would be considered part of a single datacenter.

* Consensus - When used in our documentation we use consensus to mean agreement upon
the elected leader as well as agreement on the ordering of transactions. Since these
transactions are applied to a
[finite-state machine](https://en.wikipedia.org/wiki/Finite-state_machine), our definition
of consensus implies the consistency of a replicated state machine. Consensus is described
in more detail on [Wikipedia](https://en.wikipedia.org/wiki/Consensus_(computer_science)),
and our implementation is described [here](/docs/internals/consensus.html).

* Gossip - Consul is built on top of [Serf](https://www.serf.io/) which provides a full
[gossip protocol](https://en.wikipedia.org/wiki/Gossip_protocol) that is used for multiple purposes.
Serf provides membership, failure detection, and event broadcast. Our use of these
is described more in the [gossip documentation](/docs/internals/gossip.html). It is enough to know
that gossip involves random node-to-node communication, primarily over UDP.

* LAN Gossip - Refers to the LAN gossip pool which contains nodes that are all
located on the same local area network or datacenter.

* WAN Gossip - Refers to the WAN gossip pool which contains only servers. These
servers are primarily located in different datacenters and typically communicate
over the internet or wide area network.

* RPC - Remote Procedure Call. This is a request / response mechanism allowing a
client to make a request of a server.

## 10,000 foot view

From a 10,000 foot altitude the architecture of Consul looks like this:

<div class="center">
[![Consul Architecture](/assets/images/consul-arch.png)](/assets/images/consul-arch.png)
</div>

Let's break down this image and describe each piece. First of all, we can see
that there are two datacenters, labeled "one" and "two". Consul has first
class support for [multiple datacenters](/docs/guides/datacenters.html) and
expects this to be the common case.

Within each datacenter, we have a mixture of clients and servers. It is expected
that there be between three to five servers. This strikes a balance between
availability in the case of failure and performance, as consensus gets progressively
slower as more machines are added. However, there is no limit to the number of clients,
and they can easily scale into the thousands or tens of thousands.

All the nodes that are in a datacenter participate in a [gossip protocol](/docs/internals/gossip.html).
This means there is a gossip pool that contains all the nodes for a given datacenter. This serves
a few purposes: first, there is no need to configure clients with the addresses of servers;
discovery is done automatically. Second, the work of detecting node failures
is not placed on the servers but is distributed. This makes failure detection much more
scalable than naive heartbeating schemes. Thirdly, it is used as a messaging layer to notify
when important events such as leader election take place.

The servers in each datacenter are all part of a single Raft peer set. This means that
they work together to elect a single leader, a selected server which has extra duties. The leader
is responsible for processing all queries and transactions. Transactions must also be replicated to
all peers as part of the [consensus protocol](/docs/internals/consensus.html). Because of this
requirement, when a non-leader server receives an RPC request, it forwards it to the cluster leader.

The server nodes also operate as part of a WAN gossip pool. This pool is different from the LAN pool
as it is optimized for the higher latency of the internet and is expected to contain only
other Consul server nodes. The purpose of this pool is to allow datacenters to discover each
other in a low-touch manner. Bringing a new datacenter online is as easy as joining the existing
WAN gossip. Because the servers are all operating in this pool, it also enables cross-datacenter
requests. When a server receives a request for a different datacenter, it forwards it to a random
server in the correct datacenter. That server may then forward to the local leader.

This results in a very low coupling between datacenters, but because of failure detection,
connection caching and multiplexing, cross-datacenter requests are relatively fast and reliable.

## Getting in depth

At this point we've covered the high level architecture of Consul, but there are many
more details for each of the subsystems. The [consensus protocol](/docs/internals/consensus.html) is
documented in detail as is the [gossip protocol](/docs/internals/gossip.html). The [documentation](/docs/internals/security.html)
for the security model and protocols used are also available.

For other details, either consult the code, ask in IRC, or reach out to the mailing list.
