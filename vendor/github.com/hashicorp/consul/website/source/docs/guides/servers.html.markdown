---
layout: "docs"
page_title: "Adding/Removing Servers"
sidebar_current: "docs-guides-servers"
description: |-
  Consul is designed to require minimal operator involvement, however any changes to the set of Consul servers must be handled carefully. To better understand why, reading about the consensus protocol will be useful. In short, the Consul servers perform leader election and replication. For changes to be processed, a minimum quorum of servers (N/2)+1 must be available. That means if there are 3 server nodes, at least 2 must be available.
---

# Adding/Removing Servers

Consul is designed to require minimal operator involvement, however any changes
to the set of Consul servers must be handled carefully. To better understand
why, reading about the [consensus protocol](/docs/internals/consensus.html) will
be useful. In short, the Consul servers perform leader election and replication.
For changes to be processed, a minimum quorum of servers (N/2)+1 must be available.
That means if there are 3 server nodes, at least 2 must be available.

In general, if you are ever adding and removing nodes simultaneously, it is better
to first add the new nodes and then remove the old nodes.

## Adding New Servers

Adding new servers is generally straightforward. Simply start the new
agent with the `-server` flag. At this point, the server will not be a member of
any cluster, and should emit something like:

```text
[WARN] raft: EnableSingleNode disabled, and no known peers. Aborting election.
```

This means that it does not know about any peers and is not configured to elect itself.
This is expected, and we can now add this node to the existing cluster using `join`.
From the new server, we can join any member of the existing cluster:

```text
$ consul join <Node Address>
Successfully joined cluster by contacting 1 nodes.
```

It is important to note that any node, including a non-server may be specified for
join. The gossip protocol is used to properly discover all the nodes in the cluster.
Once the node has joined, the existing cluster leader should log something like:

```text
[INFO] raft: Added peer 127.0.0.2:8300, starting replication
```

This means that raft, the underlying consensus protocol, has added the peer and begun
replicating state. Since the existing cluster may be very far ahead, it can take some
time for the new node to catch up. To check on this, run `info` on the leader:

```text
$ consul info
...
raft:
	applied_index = 47244
	commit_index = 47244
	fsm_pending = 0
	last_log_index = 47244
	last_log_term = 21
	last_snapshot_index = 40966
	last_snapshot_term = 20
	num_peers = 2
	state = Leader
	term = 21
...
```

This will provide various information about the state of Raft. In particular
the `last_log_index` shows the last log that is on disk. The same `info` command
can be run on the new server to see how far behind it is. Eventually the server
will be caught up, and the values should match.

It is best to add servers one at a time, allowing them to catch up. This avoids
the possibility of data loss in case the existing servers fail while bringing
the new servers up-to-date.

## Removing Servers

Removing servers must be done carefully to avoid causing an availability outage.
For a cluster of N servers, at least (N/2)+1 must be available for the cluster
to function. See this [deployment table](/docs/internals/consensus.html#toc_4).
If you have 3 servers, and 1 of them is currently failed, removing any servers
will cause the cluster to become unavailable.

To avoid this, it may be necessary to first add new servers to the cluster,
increasing the failure tolerance of the cluster, and then to remove old servers.
Even if all 3 nodes are functioning, removing one leaves the cluster in a state
that cannot tolerate the failure of any node.

Once you have verified the existing servers are healthy, and that the cluster
can handle a node leaving, the actual process is simple. You simply issue a
`leave` command to the server.

The server leaving should contain logs like:

```text
...
[INFO] consul: server starting leave
...
[INFO] raft: Removed ourself, transitioning to follower
...
```

The leader should also emit various logs including:

```text
...
[INFO] consul: member 'node-10-0-1-8' left, deregistering
[INFO] raft: Removed peer 10.0.1.8:8300, stopping replication
...
```

At this point the node has been gracefully removed from the cluster, and
will shut down.

## Forced Removal

In some cases, it may not be possible to gracefully remove a server. For example,
if the server simply fails, then there is no ability to issue a leave. Instead,
the cluster will detect the failure and replication will continuously retry.

If the server can be recovered, it is best to bring it back online and then gracefully
leave the cluster. However, if this is not a possibility, then the `force-leave` command
can be used to force removal of a server.

This is done by invoking that command with the name of the failed node. At this point,
the cluster leader will mark the node as having left the cluster and it will stop attempting to replicate.
