---
layout: "intro"
page_title: "Run the Agent"
sidebar_current: "gettingstarted-agent"
description: >
  The Consul agent can run in either server or client mode. Each datacenter
  must have at least one server, though a cluster of 3 or 5 servers is
  recommended. A single server deployment is highly discouraged in production
  as data loss is inevitable in a failure scenario.
---

# Run the Consul Agent

After Consul is installed, the agent must be run. The agent can run either
in server or client mode. Each datacenter must have at least one server,
though a cluster of 3 or 5 servers is recommended. A single server deployment
is _**highly**_ discouraged as data loss is inevitable in a failure scenario.

All other agents run in client mode. A client is a very lightweight
process that registers services, runs health checks, and forwards queries to
servers. The agent must be run on every node that is part of the cluster.

For more detail on bootstrapping a datacenter, see
[this guide](/docs/guides/bootstrapping.html).

## Starting the Agent

For simplicity, we'll start the Consul agent in development mode for now. This
mode is useful for bringing up a single-node Consul environment quickly and
easily. It is **not** intended to be used in production as it does not persist
any state.

```text
-$ consul agent -dev
==> Starting Consul agent...
==> Starting Consul agent RPC...
==> Consul agent running!
           Version: 'v0.7.0'
         Node name: 'Armons-MacBook-Air'
        Datacenter: 'dc1'
            Server: true (bootstrap: false)
       Client Addr: 127.0.0.1 (HTTP: 8500, HTTPS: -1, DNS: 8600, RPC: 8400)
      Cluster Addr: 127.0.0.1 (LAN: 8301, WAN: 8302)
    Gossip encrypt: false, RPC-TLS: false, TLS-Incoming: false
             Atlas: <disabled>

==> Log data will now stream in as it occurs:

    2016/09/15 10:21:10 [INFO] raft: Initial configuration (index=1): [{Suffrage:Voter ID:127.0.0.1:8300 Address:127.0.0.1:8300}]
    2016/09/15 10:21:10 [INFO] raft: Node at 127.0.0.1:8300 [Follower] entering Follower state (Leader: "")
    2016/09/15 10:21:10 [INFO] serf: EventMemberJoin: Armons-MacBook-Air 127.0.0.1
    2016/09/15 10:21:10 [INFO] serf: EventMemberJoin: Armons-MacBook-Air.dc1 127.0.0.1
    2016/09/15 10:21:10 [INFO] consul: Adding LAN server Armons-MacBook-Air (Addr: tcp/127.0.0.1:8300) (DC: dc1)
    2016/09/15 10:21:10 [INFO] consul: Adding WAN server Armons-MacBook-Air.dc1 (Addr: tcp/127.0.0.1:8300) (DC: dc1)
    2016/09/15 10:21:13 [DEBUG] http: Request GET /v1/agent/services (180.708µs) from=127.0.0.1:52369
    2016/09/15 10:21:13 [DEBUG] http: Request GET /v1/agent/services (15.548µs) from=127.0.0.1:52369
    2016/09/15 10:21:17 [WARN] raft: Heartbeat timeout from "" reached, starting election
    2016/09/15 10:21:17 [INFO] raft: Node at 127.0.0.1:8300 [Candidate] entering Candidate state in term 2
    2016/09/15 10:21:17 [DEBUG] raft: Votes needed: 1
    2016/09/15 10:21:17 [DEBUG] raft: Vote granted from 127.0.0.1:8300 in term 2. Tally: 1
    2016/09/15 10:21:17 [INFO] raft: Election won. Tally: 1
    2016/09/15 10:21:17 [INFO] raft: Node at 127.0.0.1:8300 [Leader] entering Leader state
    2016/09/15 10:21:17 [INFO] consul: cluster leadership acquired
    2016/09/15 10:21:17 [DEBUG] consul: reset tombstone GC to index 3
    2016/09/15 10:21:17 [INFO] consul: New leader elected: Armons-MacBook-Air
    2016/09/15 10:21:17 [INFO] consul: member 'Armons-MacBook-Air' joined, marking health alive
    2016/09/15 10:21:17 [INFO] agent: Synced service 'consul'
```

As you can see, the Consul agent has started and has output some log
data. From the log data, you can see that our agent is running in server mode
and has claimed leadership of the cluster. Additionally, the local member has
been marked as a healthy member of the cluster.

~> **Note for OS X Users:** Consul uses your hostname as the
default node name. If your hostname contains periods, DNS queries to
that node will not work with Consul. To avoid this, explicitly set
the name of your node with the `-node` flag.

## Cluster Members

If you run [`consul members`](/docs/commands/members.html) in another terminal, you
can see the members of the Consul cluster. We'll cover joining clusters in the next
section, but for now, you should only see one member (yourself):

```text
$ consul members
Node                Address            Status  Type    Build     Protocol  DC
Armons-MacBook-Air  172.20.20.11:8301  alive   server  0.6.1dev  2         dc1
```

The output shows our own node, the address it is running on, its
health state, its role in the cluster, and some version information.
Additional metadata can be viewed by providing the `-detailed` flag.

The output of the [`members`](/docs/commands/members.html) command is based on
the [gossip protocol](/docs/internals/gossip.html) and is eventually consistent.
That is, at any point in time, the view of the world as seen by your local
agent may not exactly match the state on the servers. For a strongly consistent
view of the world, use the [HTTP API](/api/index.html) as it forwards the
request to the Consul servers:

```text
$ curl localhost:8500/v1/catalog/nodes
[{"Node":"Armons-MacBook-Air","Address":"127.0.0.1","TaggedAddresses":{"lan":"127.0.0.1","wan":"127.0.0.1"},"CreateIndex":4,"ModifyIndex":110}]
```

In addition to the HTTP API, the [DNS interface](/docs/agent/dns.html) can
be used to query the node. Note that you have to make sure to point your DNS
lookups to the Consul agent's DNS server which runs on port 8600 by default.
The format of the DNS entries (such as "Armons-MacBook-Air.node.consul") will
be covered in more detail later.

```text
$ dig @127.0.0.1 -p 8600 Armons-MacBook-Air.node.consul
...

;; QUESTION SECTION:
;Armons-MacBook-Air.node.consul.	IN	A

;; ANSWER SECTION:
Armons-MacBook-Air.node.consul.	0 IN	A	127.0.0.1
```

## <a name="stopping"></a>Stopping the Agent

You can use `Ctrl-C` (the interrupt signal) to gracefully halt the agent.
After interrupting the agent, you should see it leave the cluster
and shut down.

By gracefully leaving, Consul notifies other cluster members that the
node _left_. If you had forcibly killed the agent process, other members
of the cluster would have detected that the node _failed_. When a member leaves,
its services and checks are removed from the catalog. When a member fails,
its health is simply marked as critical, but it is not removed from the catalog.
Consul will automatically try to reconnect to _failed_ nodes, allowing it
to recover from certain network conditions, while _left_ nodes are no longer contacted.

Additionally, if an agent is operating as a server, a graceful leave is important
to avoid causing a potential availability outage affecting the
[consensus protocol](/docs/internals/consensus.html). See the
[guides section](/docs/guides/index.html) for details on how to safely add
and remove servers.

## Next Steps

Your simple Consul cluster is up and running. Let's give it some
[services](services.html)!
