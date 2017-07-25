---
layout: "docs"
page_title: "Consensus Protocol"
sidebar_current: "docs-internals-consensus"
description: |-
  Consul uses a consensus protocol to provide Consistency as defined by CAP. The consensus protocol is based on Raft: In search of an Understandable Consensus Algorithm. For a visual explanation of Raft, see The Secret Lives of Data.
---

# Consensus Protocol

Consul uses a [consensus protocol](https://en.wikipedia.org/wiki/Consensus_(computer_science))
to provide [Consistency (as defined by CAP)](https://en.wikipedia.org/wiki/CAP_theorem).
The consensus protocol is based on
["Raft: In search of an Understandable Consensus Algorithm"](https://ramcloud.stanford.edu/wiki/download/attachments/11370504/raft.pdf).
For a visual explanation of Raft, see [The Secret Lives of Data](http://thesecretlivesofdata.com/raft).

~> **Advanced Topic!** This page covers technical details of
the internals of Consul. You don't need to know these details to effectively
operate and use Consul. These details are documented here for those who wish
to learn about them without having to go spelunking through the source code.

## Raft Protocol Overview

Raft is a consensus algorithm that is based on
[Paxos](https://en.wikipedia.org/wiki/Paxos_%28computer_science%29). Compared
to Paxos, Raft is designed to have fewer states and a simpler, more
understandable algorithm.

There are a few key terms to know when discussing Raft:

* Log - The primary unit of work in a Raft system is a log entry. The problem
of consistency can be decomposed into a *replicated log*. A log is an ordered
sequence of entries. We consider the log consistent if all members agree on
the entries and their order.

* FSM - [Finite State Machine](https://en.wikipedia.org/wiki/Finite-state_machine).
An FSM is a collection of finite states with transitions between them. As new logs
are applied, the FSM is allowed to transition between states. Application of the
same sequence of logs must result in the same state, meaning behavior must be deterministic.

* Peer set - The peer set is the set of all members participating in log replication.
For Consul's purposes, all server nodes are in the peer set of the local datacenter.

* Quorum - A quorum is a majority of members from a peer set: for a set of size `n`,
quorum requires at least `(n/2)+1` members.
For example, if there are 5 members in the peer set, we would need 3 nodes
to form a quorum. If a quorum of nodes is unavailable for any reason, the
cluster becomes *unavailable* and no new logs can be committed.

* Committed Entry - An entry is considered *committed* when it is durably stored
on a quorum of nodes. Once an entry is committed it can be applied.

* Leader - At any given time, the peer set elects a single node to be the leader.
The leader is responsible for ingesting new log entries, replicating to followers,
and managing when an entry is considered committed.

Raft is a complex protocol and will not be covered here in detail (for those who
desire a more comprehensive treatment, the full specification is available in this
[paper](https://ramcloud.stanford.edu/wiki/download/attachments/11370504/raft.pdf)).
We will, however, attempt to provide a high level description which may be useful
for building a mental model.

Raft nodes are always in one of three states: follower, candidate, or leader. All
nodes initially start out as a follower. In this state, nodes can accept log entries
from a leader and cast votes. If no entries are received for some time, nodes
self-promote to the candidate state. In the candidate state, nodes request votes from
their peers. If a candidate receives a quorum of votes, then it is promoted to a leader.
The leader must accept new log entries and replicate to all the other followers.
In addition, if stale reads are not acceptable, all queries must also be performed on
the leader.

Once a cluster has a leader, it is able to accept new log entries. A client can
request that a leader append a new log entry (from Raft's perspective, a log entry
is an opaque binary blob). The leader then writes the entry to durable storage and
attempts to replicate to a quorum of followers. Once the log entry is considered
*committed*, it can be *applied* to a finite state machine. The finite state machine
is application specific; in Consul's case, we use
[BoltDB](https://github.com/boltdb/bolt) to maintain cluster state.

Obviously, it would be undesirable to allow a replicated log to grow in an unbounded
fashion. Raft provides a mechanism by which the current state is snapshotted and the
log is compacted. Because of the FSM abstraction, restoring the state of the FSM must
result in the same state as a replay of old logs. This allows Raft to capture the FSM
state at a point in time and then remove all the logs that were used to reach that
state. This is performed automatically without user intervention and prevents unbounded
disk usage while also minimizing time spent replaying logs. One of the advantages of
using BoltDB is that it allows Consul to continue accepting new transactions even while
old state is being snapshotted, preventing any availability issues.

Consensus is fault-tolerant up to the point where quorum is available.
If a quorum of nodes is unavailable, it is impossible to process log entries or reason
about peer membership. For example, suppose there are only 2 peers: A and B. The quorum
size is also 2, meaning both nodes must agree to commit a log entry. If either A or B
fails, it is now impossible to reach quorum. This means the cluster is unable to add
or remove a node or to commit any additional log entries. This results in
*unavailability*. At this point, manual intervention would be required to remove
either A or B and to restart the remaining node in bootstrap mode.

A Raft cluster of 3 nodes can tolerate a single node failure while a cluster
of 5 can tolerate 2 node failures. The recommended configuration is to either
run 3 or 5 Consul servers per datacenter. This maximizes availability without
greatly sacrificing performance. The [deployment table](#deployment_table) below
summarizes the potential cluster size options and the fault tolerance of each.

In terms of performance, Raft is comparable to Paxos. Assuming stable leadership,
committing a log entry requires a single round trip to half of the cluster.
Thus, performance is bound by disk I/O and network latency. Although Consul is
not designed to be a high-throughput write system, it should handle on the order
of hundreds to thousands of transactions per second depending on network and
hardware configuration.

## Raft in Consul

Only Consul server nodes participate in Raft and are part of the peer set. All
client nodes forward requests to servers. Part of the reason for this design is
that, as more members are added to the peer set, the size of the quorum also increases.
This introduces performance problems as you may be waiting for hundreds of machines
to agree on an entry instead of a handful.

When getting started, a single Consul server is put into "bootstrap" mode. This mode
allows it to self-elect as a leader. Once a leader is elected, other servers can be
added to the peer set in a way that preserves consistency and safety. Eventually,
once the first few servers are added, bootstrap mode can be disabled. See [this
guide](/docs/guides/bootstrapping.html) for more details.

Since all servers participate as part of the peer set, they all know the current
leader. When an RPC request arrives at a non-leader server, the request is
forwarded to the leader. If the RPC is a *query* type, meaning it is read-only,
the leader generates the result based on the current state of the FSM. If
the RPC is a *transaction* type, meaning it modifies state, the leader
generates a new log entry and applies it using Raft. Once the log entry is committed
and applied to the FSM, the transaction is complete.

Because of the nature of Raft's replication, performance is sensitive to network
latency. For this reason, each datacenter elects an independent leader and maintains
a disjoint peer set. Data is partitioned by datacenter, so each leader is responsible
only for data in their datacenter. When a request is received for a remote datacenter,
the request is forwarded to the correct leader. This design allows for lower latency
transactions and higher availability without sacrificing consistency.

## Consistency Modes

Although all writes to the replicated log go through Raft, reads are more
flexible. To support various trade-offs that developers may want, Consul
supports 3 different consistency modes for reads.

The three read modes are:

* `default` - Raft makes use of leader leasing, providing a time window
  in which the leader assumes its role is stable. However, if a leader
  is partitioned from the remaining peers, a new leader may be elected
  while the old leader is holding the lease. This means there are 2 leader
  nodes. There is no risk of a split-brain since the old leader will be
  unable to commit new logs. However, if the old leader services any reads,
  the values are potentially stale. The default consistency mode relies only
  on leader leasing, exposing clients to potentially stale values. We make
  this trade-off because reads are fast, usually strongly consistent, and
  only stale in a hard-to-trigger situation. The time window of stale reads
  is also bounded since the leader will step down due to the partition.

* `consistent` - This mode is strongly consistent without caveats. It requires
  that a leader verify with a quorum of peers that it is still leader. This
  introduces an additional round-trip to all server nodes. The trade-off is
  always consistent reads but increased latency due to the extra round trip.

* `stale` - This mode allows any server to service the read regardless of if
  it is the leader. This means reads can be arbitrarily stale but are generally
  within 50 milliseconds of the leader. The trade-off is very fast and scalable
  reads but with stale values. This mode allows reads without a leader meaning
  a cluster that is unavailable will still be able to respond.

For more documentation about using these various modes, see the
[HTTP API](/api/index.html).

## <a name="deployment_table"></a>Deployment Table

Below is a table that shows quorum size and failure tolerance for various
cluster sizes. The recommended deployment is either 3 or 5 servers. A single
server deployment is _**highly**_ discouraged as data loss is inevitable in a
failure scenario.

<table class="table table-bordered table-striped">
  <tr>
    <th>Servers</th>
    <th>Quorum Size</th>
    <th>Failure Tolerance</th>
  </tr>
  <tr>
    <td>1</td>
    <td>1</td>
    <td>0</td>
  </tr>
  <tr>
    <td>2</td>
    <td>2</td>
    <td>0</td>
  </tr>
  <tr class="warning">
    <td>3</td>
    <td>2</td>
    <td>1</td>
  </tr>
  <tr>
    <td>4</td>
    <td>3</td>
    <td>1</td>
  </tr>
  <tr class="warning">
    <td>5</td>
    <td>3</td>
    <td>2</td>
  </tr>
  <tr>
    <td>6</td>
    <td>4</td>
    <td>2</td>
  </tr>
  <tr>
    <td>7</td>
    <td>4</td>
    <td>3</td>
  </tr>
</table>
