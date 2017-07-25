---
layout: "docs"
page_title: "Outage Recovery"
sidebar_current: "docs-guides-outage"
description: |-
  Don't panic! This is a critical first step. Depending on your deployment configuration, it may take only a single server failure for cluster unavailability. Recovery requires an operator to intervene, but recovery is straightforward.
---

# Outage Recovery

Don't panic! This is a critical first step.

Depending on your
[deployment configuration](/docs/internals/consensus.html#deployment_table), it
may take only a single server failure for cluster unavailability. Recovery
requires an operator to intervene, but the process is straightforward.

~> This guide is for recovery from a Consul outage due to a majority
of server nodes in a datacenter being lost. If you are just looking to
add or remove a server, [see this guide](/docs/guides/servers.html).

## Failure of a Single Server Cluster

If you had only a single server and it has failed, simply restart it. A
single server configuration requires the
[`-bootstrap`](/docs/agent/options.html#_bootstrap) or
[`-bootstrap-expect=1`](/docs/agent/options.html#_bootstrap_expect)
flag. If the server cannot be recovered, you need to bring up a new
server. See the [bootstrapping guide](/docs/guides/bootstrapping.html)
for more detail.

In the case of an unrecoverable server failure in a single server cluster, data
loss is inevitable since data was not replicated to any other servers. This is
why a single server deploy is **never** recommended.

Any services registered with agents will be re-populated when the new server
comes online as agents perform [anti-entropy](/docs/internals/anti-entropy.html).

## Failure of a Server in a Multi-Server Cluster

If you think the failed server is recoverable, the easiest option is to bring
it back online and have it rejoin the cluster with the same IP address, returning
the cluster to a fully healthy state. Similarly, even if you need to rebuild a
new Consul server to replace the failed node, you may wish to do that immediately.
Keep in mind that the rebuilt server needs to have the same IP address as the failed
server. Again, once this server is online and has rejoined, the cluster will return
to a fully healthy state.

Both of these strategies involve a potentially lengthy time to reboot or rebuild
a failed server. If this is impractical or if building a new server with the same
IP isn't an option, you need to remove the failed server. Usually, you can issue
a [`consul force-leave`](/docs/commands/force-leave.html) command to remove the failed
server if it's still a member of the cluster.

If [`consul force-leave`](/docs/commands/force-leave.html) isn't able to remove the
server, you have two methods available to remove it, depending on your version of Consul:

* In Consul 0.7 and later, you can use the [`consul operator`](/docs/commands/operator.html#raft-remove-peer) command to remove the stale peer server on the fly with no downtime if the cluster has a leader.

* In versions of Consul prior to 0.7, you can manually remove the stale peer
server using the `raft/peers.json` recovery file on all remaining servers. See
the [section below](#peers.json) for details on this procedure. This process
requires a Consul downtime to complete.

In Consul 0.7 and later, you can use the [`consul operator`](/docs/commands/operator.html#raft-list-peers)
command to inspect the Raft configuration:

```
$ consul operator raft -list-peers
Node     ID              Address         State     Voter
alice    10.0.1.8:8300   10.0.1.8:8300   follower  true
bob      10.0.1.6:8300   10.0.1.6:8300   leader    true
carol    10.0.1.7:8300   10.0.1.7:8300   follower  true
```

## Failure of Multiple Servers in a Multi-Server Cluster

In the event that multiple servers are lost, causing a loss of quorum and a
complete outage, partial recovery is possible using data on the remaining
servers in the cluster. There may be data loss in this situation because multiple
servers were lost, so information about what's committed could be incomplete.
The recovery process implicitly commits all outstanding Raft log entries, so
it's also possible to commit data that was uncommitted before the failure.

See the [section below](#peers.json) for details of the recovery procedure. You
simply include just the remaining servers in the `raft/peers.json` recovery file.
The cluster should be able to elect a leader once the remaining servers are all
restarted with an identical `raft/peers.json` configuration.

Any new servers you introduce later can be fresh with totally clean data directories
and joined using Consul's `join` command.

In extreme cases, it should be possible to recover with just a single remaining
server by starting that single server with itself as the only peer in the
`raft/peers.json` recovery file.

Prior to Consul 0.7 it wasn't always possible to recover from certain
types of outages with `raft/peers.json` because this was ingested before any Raft
log entries were played back. In Consul 0.7 and later, the `raft/peers.json`
recovery file is final, and a snapshot is taken after it is ingested, so you are
guaranteed to start with your recovered configuration. This does implicitly commit
all Raft log entries, so should only be used to recover from an outage, but it
should allow recovery from any situation where there's some cluster data available.

<a name="peers.json"></a>
## Manual Recovery Using peers.json

To begin, stop all remaining servers. You can attempt a graceful leave,
but it will not work in most cases. Do not worry if the leave exits with an
error. The cluster is in an unhealthy state, so this is expected.

In Consul 0.7 and later, the `peers.json` file is no longer present
by default and is only used when performing recovery. This file will be deleted
after Consul starts and ingests this file. Consul 0.7 also uses a new, automatically-
created `raft/peers.info` file to avoid ingesting the `raft/peers.json` file on the
first start after upgrading. Be sure to leave `raft/peers.info` in place for proper
operation.

Using `raft/peers.json` for recovery can cause uncommitted Raft log entries to be
implicitly committed, so this should only be used after an outage where no
other option is available to recover a lost server. Make sure you don't have
any automated processes that will put the peers file in place on a
periodic basis.

The next step is to go to the [`-data-dir`](/docs/agent/options.html#_data_dir)
of each Consul server. Inside that directory, there will be a `raft/`
sub-directory. We need to create a `raft/peers.json` file. It should look
something like:

```javascript
[
"10.0.1.8:8300",
"10.0.1.6:8300",
"10.0.1.7:8300"
]
```

Simply create entries for all remaining servers. You must confirm
that servers you do not include here have indeed failed and will not later
rejoin the cluster. Ensure that this file is the same across all remaining
server nodes.

At this point, you can restart all the remaining servers. In Consul 0.7 and
later you will see them ingest recovery file:

```text
...
2016/08/16 14:39:20 [INFO] consul: found peers.json file, recovering Raft configuration...
2016/08/16 14:39:20 [INFO] consul.fsm: snapshot created in 12.484µs
2016/08/16 14:39:20 [INFO] snapshot: Creating new snapshot at /tmp/peers/raft/snapshots/2-5-1471383560779.tmp
2016/08/16 14:39:20 [INFO] consul: deleted peers.json file after successful recovery
2016/08/16 14:39:20 [INFO] raft: Restored from snapshot 2-5-1471383560779
2016/08/16 14:39:20 [INFO] raft: Initial configuration (index=1): [{Suffrage:Voter ID:10.212.15.121:8300 Address:10.212.15.121:8300}]
...
```

If any servers managed to perform a graceful leave, you may need to have them
rejoin the cluster using the [`join`](/docs/commands/join.html) command:

```text
$ consul join <Node Address>
Successfully joined cluster by contacting 1 nodes.
```

It should be noted that any existing member can be used to rejoin the cluster
as the gossip protocol will take care of discovering the server nodes.

At this point, the cluster should be in an operable state again. One of the
nodes should claim leadership and emit a log like:

```text
[INFO] consul: cluster leadership acquired
```

In Consul 0.7 and later, you can use the [`consul operator`](/docs/commands/operator.html#raft-list-peers)
command to inspect the Raft configuration:

```
$ consul operator raft -list-peers
Node     ID              Address         State     Voter
alice    10.0.1.8:8300   10.0.1.8:8300   follower  true
bob      10.0.1.6:8300   10.0.1.6:8300   leader    true
carol    10.0.1.7:8300   10.0.1.7:8300   follower  true
```
