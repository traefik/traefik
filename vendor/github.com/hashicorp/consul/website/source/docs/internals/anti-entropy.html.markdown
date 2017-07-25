---
layout: "docs"
page_title: "Anti-Entropy"
sidebar_current: "docs-internals-anti-entropy"
description: >
  This section details the process and use of anti-entropy in Consul.
---

# Anti-Entropy

Consul uses an advanced method of maintaining service and health information.
This page details how services and checks are registered, how the catalog is
populated, and how health status information is updated as it changes.

~> **Advanced Topic!** This page covers technical details of
the internals of Consul. You don't need to know these details to effectively
operate and use Consul. These details are documented here for those who wish
to learn about them without having to go spelunking through the source code.

### Components

It is important to first understand the moving pieces involved in services and
health checks: the [agent](#agent) and the [catalog](#catalog). These are
described conceptually below to make anti-entropy easier to understand.

<a name="agent"></a>
#### Agent

Each Consul agent maintains its own set of service and check registrations as
well as health information. The agents are responsible for executing their own
health checks and updating their local state.

Services and checks within the context of an agent have a rich set of
configuration options available. This is because the agent is responsible for
generating information about its services and their health through the use of
[health checks](/docs/agent/checks.html).

<a name="catalog"></a>
#### Catalog

Consul's service discovery is backed by a service catalog. This catalog is
formed by aggregating information submitted by the agents. The catalog maintains
the high-level view of the cluster, including which services are available,
which nodes run those services, health information, and more. The catalog is
used to expose this information via the various interfaces Consul provides,
including DNS and HTTP.

Services and checks within the context of the catalog have a much more limited
set of fields when compared with the agent. This is because the catalog is only
responsible for recording and returning information *about* services, nodes, and
health.

The catalog is maintained only by server nodes. This is because the catalog is
replicated via the [Raft log](/docs/internals/consensus.html) to provide a
consolidated and consistent view of the cluster.

<a name="anti-entropy"></a>
### Anti-Entropy

Entropy is the tendency of systems to become increasingly disordered. Consul's
anti-entropy mechanisms are designed to counter this tendency, to keep the
state of the cluster ordered even through failures of its components.

Consul has a clear separation between the global service catalog and the agent
local state as discussed above. The anti-entropy mechanism reconciles these two
views of the world: anti-entropy is a synchronization of the local agent state and
the catalog. For example, when a user registers a new service or check with the
agent, the agent in turn notifies the catalog that this new check exists.
Similarly, when a check is deleted from the agent, it is consequently removed from
the catalog as well.

Anti-entropy is also used to update availability information. As agents run
their health checks, their status may change in which case their new status
is synced to the catalog. Using this information, the catalog can respond
intelligently to queries about its nodes and services based on their
availability.

During this synchronization, the catalog is also checked for correctness. If
any services or checks exist in the catalog that the agent is not aware of, they
will be automatically removed to make the catalog reflect the proper set of
services and health information for that agent. Consul treats the state of the
agent as authoritative; if there are any differences between the agent
and catalog view, the agent local view will always be used.

### Periodic Synchronization

In addition to running when changes to the agent occur, anti-entropy is also a
long-running process which periodically wakes up to sync service and check
status to the catalog. This ensures that the catalog closely matches the agent's
true state. This also allows Consul to re-populate the service catalog even in
the case of complete data loss.

To avoid saturation, the amount of time between periodic anti-entropy runs will
vary based on cluster size. The table below defines the relationship between
cluster size and sync interval:

<table class="table table-bordered table-striped">
  <tr>
    <th>Cluster Size</th>
    <th>Periodic Sync Interval</th>
  </tr>
  <tr>
    <td>1 - 128</td>
    <td>1 minute</td>
  </tr>
  <tr>
    <td>129 - 256</td>
    <td>2 minutes</td>
  </tr>
  <tr>
    <td>257 - 512</td>
    <td>3 minutes</td>
  </tr>
  <tr>
    <td>513 - 1024</td>
    <td>4 minutes</td>
  </tr>
  <tr>
    <td>...</td>
    <td>...</td>
  </tr>
</table>

The intervals above are approximate. Each Consul agent will choose a randomly
staggered start time within the interval window to avoid a thundering herd.

### Best-effort sync

Anti-entropy can fail in a number of cases, including misconfiguration of the
agent or its operating environment, I/O problems (full disk, filesystem
permission, etc.), networking problems (agent cannot communicate with server),
among others. Because of this, the agent attempts to sync in best-effort
fashion.

If an error is encountered during an anti-entropy run, the error is logged and
the agent continues to run. The anti-entropy mechanism is run periodically to
automatically recover from these types of transient failures.

### EnableTagOverride

Synchronization of service registration can be partially modified to
allow external agents to change the tags for a service. This can be
useful in situations where an external monitoring service needs to be
the source of truth for tag information. For example, the Redis
database and its monitoring service Redis Sentinel have this kind of
relationship. Redis instances are responsible for much of their
configuration, but Sentinels determine whether the Redis instance is a
primary or a secondary. Using the Consul service configuration item
[EnableTagOverride](/docs/agent/services.html) you can instruct the
Consul agent on which the Redis database is running to NOT update the
tags during anti-entropy synchronization. For more information see
[Services](/docs/agent/services.html) page.
