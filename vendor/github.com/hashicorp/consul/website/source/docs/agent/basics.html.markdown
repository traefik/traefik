---
layout: "docs"
page_title: "Agent"
sidebar_current: "docs-agent-running"
description: |-
  The Consul agent is the core process of Consul. The agent maintains membership information, registers services, runs checks, responds to queries, and more. The agent must run on every node that is part of a Consul cluster.
---

# Consul Agent

The Consul agent is the core process of Consul. The agent maintains membership
information, registers services, runs checks, responds to queries,
and more. The agent must run on every node that is part of a Consul cluster.

Any agent may run in one of two modes: client or server. A server
node takes on the additional responsibility of being part of the [consensus quorum](/docs/internals/consensus.html).
These nodes take part in Raft and provide strong consistency and availability in
the case of failure. The higher burden on the server nodes means that usually they
should be run on dedicated instances -- they are more resource intensive than a client
node. Client nodes make up the majority of the cluster, and they are very lightweight
as they interface with the server nodes for most operations and maintain very little state
of their own.

## Running an Agent

The agent is started with the [`consul agent`](/docs/commands/agent.html) command. This
command blocks, running forever or until told to quit. The agent command takes a variety
of configuration options, but most have sane defaults.

When running [`consul agent`](/docs/commands/agent.html), you should see output similar to this:

```text
$ consul agent -data-dir=/tmp/consul
==> Starting Consul agent...
==> Consul agent running!
       Node name: 'Armons-MacBook-Air'
      Datacenter: 'dc1'
          Server: false (bootstrap: false)
     Client Addr: 127.0.0.1 (HTTP: 8500, DNS: 8600)
    Cluster Addr: 192.168.1.43 (LAN: 8301, WAN: 8302)
           Atlas: (Infrastructure: 'hashicorp/test' Join: true)

==> Log data will now stream in as it occurs:

    [INFO] serf: EventMemberJoin: Armons-MacBook-Air.local 192.168.1.43
...
```

There are several important messages that [`consul agent`](/docs/commands/agent.html) outputs:

* **Node name**: This is a unique name for the agent. By default, this
  is the hostname of the machine, but you may customize it using the
  [`-node`](/docs/agent/options.html#_node) flag.

* **Datacenter**: This is the datacenter in which the agent is configured to run.
 Consul has first-class support for multiple datacenters; however, to work efficiently,
 each node must be configured to report its datacenter. The [`-datacenter`](/docs/agent/options.html#_datacenter)
 flag can be used to set the datacenter. For single-DC configurations, the agent
 will default to "dc1".

* **Server**: This indicates whether the agent is running in server or client mode.
  Server nodes have the extra burden of participating in the consensus quorum,
  storing cluster state, and handling queries. Additionally, a server may be
  in ["bootstrap"](/docs/agent/options.html#_bootstrap_expect) mode. Multiple servers
  cannot be in bootstrap mode as that would put the cluster in an inconsistent state.

* **Client Addr**: This is the address used for client interfaces to the agent.
  This includes the ports for the HTTP and DNS interfaces. By default, this binds only
  to localhost. If you change this address or port, you'll have to specify a `-http-addr`
  whenever you run commands such as [`consul members`](/docs/commands/members.html) to
  indicate how to reach the agent. Other applications can also use the HTTP address and port
  [to control Consul](/api/index.html).

* **Cluster Addr**: This is the address and set of ports used for communication between
  Consul agents in a cluster. Not all Consul agents in a cluster have to
  use the same port, but this address **MUST** be reachable by all other nodes.

* **Atlas**: This shows the [Atlas infrastructure](https://atlas.hashicorp.com)
  with which the node is registered. It also indicates if auto-join is enabled.
  The Atlas infrastructure is set using [`-atlas`](/docs/agent/options.html#_atlas)
  and auto-join is enabled by setting [`-atlas-join`](/docs/agent/options.html#_atlas_join).

## Stopping an Agent

An agent can be stopped in two ways: gracefully or forcefully. To gracefully
halt an agent, send the process an interrupt signal (usually
`Ctrl-C` from a terminal or running `kill -INT consul_pid` ). When gracefully exiting, the agent first notifies
the cluster it intends to leave the cluster. This way, other cluster members
notify the cluster that the node has _left_.

Alternatively, you can force kill the agent by sending it a kill signal.
When force killed, the agent ends immediately. The rest of the cluster will
eventually (usually within seconds) detect that the node has died and
notify the cluster that the node has _failed_.

It is especially important that a server node be allowed to leave gracefully
so that there will be a minimal impact on availability as the server leaves
the consensus quorum.

For client agents, the difference between a node _failing_ and a node _leaving_
may not be important for your use case. For example, for a web server and load
balancer setup, both result in the same outcome: the web node is removed
from the load balancer pool.

## Lifecycle

Every agent in the Consul cluster goes through a lifecycle. Understanding
this lifecycle is useful for building a mental model of an agent's interactions
with a cluster and how the cluster treats a node.

When an agent is first started, it does not know about any other node in the cluster.
To discover its peers, it must _join_ the cluster. This is done with the
[`join`](/docs/commands/join.html)
command or by providing the proper configuration to auto-join on start. Once a node
joins, this information is gossiped to the entire cluster, meaning all nodes will
eventually be aware of each other. If the agent is a server, existing servers will
begin replicating to the new node.

In the case of a network failure, some nodes may be unreachable by other nodes.
In this case, unreachable nodes are marked as _failed_. It is impossible to distinguish
between a network failure and an agent crash, so both cases are handled the same.
Once a node is marked as failed, this information is updated in the service catalog.

-> **Note:** There is some nuance here since this update is only possible if the servers can still [form a quorum](/docs/internals/consensus.html). Once the network recovers or a crashed agent restarts the cluster will repair itself and unmark a node as failed. The health check in the catalog will also be updated to reflect this.

When a node _leaves_, it specifies its intent to do so, and the cluster
marks that node as having _left_. Unlike the _failed_ case, all of the
services provided by a node are immediately deregistered. If the agent was
a server, replication to it will stop.

To prevent an accumulation of dead nodes (nodes in either _failed_ or _left_
states), Consul will automatically remove dead nodes out of the catalog. This
process is called _reaping_. This is currently done on a configurable
interval of 72 hours (changing the reap interval is *not* recommended due to
its consequences during outage situations). Reaping is similar to leaving,
causing all associated services to be deregistered.
