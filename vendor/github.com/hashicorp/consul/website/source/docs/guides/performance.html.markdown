---
layout: "docs"
page_title: "Server Performance"
sidebar_current: "docs-guides-performance"
description: |-
  Consul requires different amounts of compute resources, depending on cluster size and expected workload. This guide provides guidance on choosing compute resources.
---

# Server Performance

Since Consul servers run a [consensus protocol](/docs/internals/consensus.html) to
process all write operations and are contacted on nearly all read operations, server
performance is critical for overall throughput and health of a Consul cluster. Servers
are generally I/O bound for writes because the underlying Raft log store performs a sync
to disk every time an entry is appended. Servers are generally CPU bound for reads since
reads work from a fully in-memory data store that is optimized for concurrent access.

<a name="minimum"></a>
## Minimum Server Requirements

In Consul 0.7, the default server [performance parameters](/docs/agent/options.html#performance)
were tuned to allow Consul to run reliably (but relatively slowly) on a server cluster of three
[AWS t2.micro](https://aws.amazon.com/ec2/instance-types/) instances. These thresholds
were determined empirically using a leader instance that was under sufficient read, write,
and network load to cause it to permanently be at zero CPU credits, forcing it to the baseline
performance mode for that instance type. Real-world workloads typically have more bursts of
activity, so this is a conservative and pessimistic tuning strategy.

This default was chosen based on feedback from users, many of whom wanted a low cost way
to run small production or development clusters with low cost compute resources, at the
expense of some performance in leader failure detection and leader election times.

The default performance configuration is equivalent to this:

```javascript
{
  "performance": {
    "raft_multiplier": 5
  }
}
```

<a name="production"></a>
## Production Server Requirements

When running Consul 0.7 and later in production, it is recommended to configure the server
[performance parameters](/docs/agent/options.html#performance) back to Consul's original
high-performance settings. This will let Consul servers detect a failed leader and complete
leader elections much more quickly than the default configuration which extends key Raft
timeouts by a factor of 5, so it can be quite slow during these events.

The high performance configuration is simple and looks like this:

```javascript
{
  "performance": {
    "raft_multiplier": 1
  }
}
```

It's best to benchmark with a realistic workload when choosing a production server for Consul.
Here are some general recommendations:

* Consul will make use of multiple cores, and at least 2 cores are recommended.

* For write-heavy workloads, disk speed on the servers is key for performance. Use SSDs or
another fast disk technology for the best write throughput.

* <a name="last-contact"></a>Spurious leader elections can be caused by networking issues between
the servers or insufficient CPU resources. Users in cloud environments often bump their servers
up to the next instance class with improved networking and CPU until leader elections stabilize,
and in Consul 0.7 or later the [performance parameters](/docs/agent/options.html#performance)
configuration now gives you tools to trade off performance instead of upsizing servers. You can
use the [`consul.raft.leader.lastContact` telemetry](/docs/agent/telemetry.html#last-contact)
to observe how the Raft timing is performing and guide the decision to de-tune Raft performance
or add more powerful servers.

* For DNS-heavy workloads, configuring all Consul agents in a cluster with the
[`allow_stale`](/docs/agent/options.html#allow_stale) configuration option will allow reads to
scale across all Consul servers, not just the leader. Consul 0.7 and later enables stale reads
for DNS by default. See [Stale Reads](/docs/guides/dns-cache.html#stale) in the
[DNS Caching](/docs/guides/dns-cache.html) guide for more details. It's also good to set
reasonable, non-zero [DNS TTL values](/docs/guides/dns-cache.html#ttl) if your clients will
respect them.

* In other applications that perform high volumes of reads against Consul, consider using the
[stale consistency mode](/api/index.html#consistency) available to allow reads to scale
across all the servers and not just be forwarded to the leader.

## Memory Requirements

Consul server agents operate on a working set of data comprised of key/value
entries, the service catalog, prepared queries, access control lists, and
sessions in memory. These data are persisted through Raft to disk in the form
of a snapshot and log of changes since the previous snapshot for durability.

When planning for memory requirements, you should typically allocate
enough RAM for your server agents to contain between 2 to 4 times the working
set size. You can determine the working set size by noting the value of
`consul.runtime.alloc_bytes` in the [Telemetry data](/docs/agent/telemetry.html).

> NOTE: Consul is not designed to serve as a general purpose database, and you
> should keep this in mind when choosing what data are populated to the
> key/value store.
