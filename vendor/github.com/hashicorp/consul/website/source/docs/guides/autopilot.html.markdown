---
layout: "docs"
page_title: "Autopilot"
sidebar_current: "docs-guides-autopilot"
description: |-
  This guide covers how to configure and use Autopilot features.
---

# Autopilot

Autopilot is a set of new features added in Consul 0.8 to allow for automatic
operator-friendly management of Consul servers. It includes cleanup of dead
servers, monitoring the state of the Raft cluster, and stable server introduction.

To enable Autopilot features (with the exception of dead server cleanup),
the [`raft_protocol`](/docs/agent/options.html#_raft_protocol) setting in
the Agent configuration must be set to 3 or higher on all servers. In Consul
0.8 this setting defaults to 2; in Consul 0.9 it will default to 3. For more
information, see the [Version Upgrade section](/docs/upgrade-specific.html#raft_protocol)
on Raft Protocol versions.

## Configuration

The configuration of Autopilot is loaded by the leader from the agent's
[`autopilot`](/docs/agent/options.html#autopilot) settings when initially
bootstrapping the cluster. After bootstrapping, the configuration can
be viewed or modified either via the [`operator autopilot`]
(/docs/commands/operator/autopilot.html) subcommand or the
[`/v1/operator/autopilot/configuration`](/api/operator.html#autopilot-configuration)
HTTP endpoint:

```
$ consul operator autopilot get-config
CleanupDeadServers = true
LastContactThreshold = 200ms
MaxTrailingLogs = 250
ServerStabilizationTime = 10s
RedundancyZoneTag = ""
DisableUpgradeMigration = false

$ consul operator autopilot set-config -cleanup-dead-servers=false
Configuration updated!

$ consul operator autopilot get-config
CleanupDeadServers = false
LastContactThreshold = 200ms
MaxTrailingLogs = 250
ServerStabilizationTime = 10s
RedundancyZoneTag = ""
DisableUpgradeMigration = false
```

## Dead Server Cleanup

Dead servers will periodically be cleaned up and removed from the Raft peer
set, to prevent them from interfering with the quorum size and leader elections.
This cleanup will also happen whenever a new server is successfully added to the
cluster.

Prior to Autopilot, it would take 72 hours for dead servers to be automatically reaped,
or operators had to script a `consul force-leave`. If another server failure occurred,
it could jeopardize the quorum, even if the failed Consul server had been automatically
replaced. Autopilot helps prevent these kinds of outages by quickly removing failed
servers as soon as a replacement Consul server comes online. When servers are removed 
by the cleanup process they will enter the "left" state.

This option can be disabled by running `consul operator autopilot set-config`
with the `-cleanup-dead-servers=false` option.

## Server Health Checking

An internal health check runs on the leader to track the stability of servers.
</br>A server is considered healthy if:

- It has a SerfHealth status of 'Alive'
- The time since its last contact with the current leader is below
`LastContactThreshold`
- Its latest Raft term matches the leader's term
- The number of Raft log entries it trails the leader by does not exceed
`MaxTrailingLogs`

The status of these health checks can be viewed through the [`/v1/operator/autopilot/health`]
(/api/operator.html#autopilot-health) HTTP endpoint, with a top level
`Healthy` field indicating the overall status of the cluster:

```
$ curl localhost:8500/v1/operator/autopilot/health
{
    "Healthy": true,
    "FailureTolerance": 0,
    "Servers": [
        {
            "ID": "e349749b-3303-3ddf-959c-b5885a0e1f6e",
            "Name": "node1",
            "Address": "127.0.0.1:8300",
            "SerfStatus": "alive",
            "Version": "0.8.0",
            "Leader": true,
            "LastContact": "0s",
            "LastTerm": 2,
            "LastIndex": 10,
            "Healthy": true,
            "Voter": true,
            "StableSince": "2017-03-28T18:28:52Z"
        },
        {
            "ID": "e35bde83-4e9c-434f-a6ef-453f44ee21ea",
            "Name": "node2",
            "Address": "127.0.0.1:8705",
            "SerfStatus": "alive",
            "Version": "0.8.0",
            "Leader": false,
            "LastContact": "35.371007ms",
            "LastTerm": 2,
            "LastIndex": 10,
            "Healthy": true,
            "Voter": false,
            "StableSince": "2017-03-28T18:29:10Z"
        }
    ]
}
```

## Stable Server Introduction

When a new server is added to the cluster, there is a waiting period where it
must be healthy and stable for a certain amount of time before being promoted
to a full, voting member. This can be configured via the `ServerStabilizationTime`
setting.

---

~> The following Autopilot features are available only in
   [Consul Enterprise](https://www.hashicorp.com/products/consul/) version 0.8.0 and later.

## Server Read Scaling

With the [`-non-voting-server`](/docs/agent/options.html#_non_voting_server) option, a
server can be explicitly marked as a non-voter and will never be promoted to a voting
member. This can be useful when more read scaling is needed; being a non-voter means
that the server will still have data replicated to it, but it will not be part of the
quorum that the leader must wait for before committing log entries.

## Redundancy Zones

Prior to Autopilot, it was difficult to deploy servers in a way that took advantage of
isolated failure domains such as AWS Availability Zones; users would be forced to either
have an overly-large quorum (2-3 nodes per AZ) or give up redundancy within an AZ by
deploying just one server in each.

If the `RedundancyZoneTag` setting is set, Consul will use its value to look for a
zone in each server's specified [`-node-meta`](/docs/agent/options.html#_node_meta)
tag. For example, if `RedundancyZoneTag` is set to `zone`, and `-node-meta zone:east1a`
is used when starting a server, that server's redundancy zone will be `east1a`.

Consul will then use these values to partition the servers by redundancy zone, and will
aim to keep one voting server per zone. Extra servers in each zone will stay as non-voters
on standby to be promoted if the active voter leaves or dies.

## Upgrade Migrations

Autopilot in Consul Enterprise supports upgrade migrations by default. To disable this
functionality, set `DisableUpgradeMigration` to true.

When a new server is added and Autopilot detects that its Consul version is newer than
that of the existing servers, Autopilot will avoid promoting the new server until enough
newer-versioned servers have been added to the cluster. When the count of new servers
equals or exceeds that of the old servers, Autopilot will begin promoting the new servers
to voters and demoting the old servers. After this is finished, the old servers can be
safely removed from the cluster.

To check the consul version of the servers, either the [autopilot health]
(/api/operator.html#autopilot-health) endpoint or the `consul members`
command can be used:

```
$ consul members
Node   Address         Status  Type    Build  Protocol  DC
node1  127.0.0.1:8301  alive   server  0.7.5  2         dc1
node2  127.0.0.1:8703  alive   server  0.7.5  2         dc1
node3  127.0.0.1:8803  alive   server  0.7.5  2         dc1
node4  127.0.0.1:8203  alive   server  0.8.0  2         dc1
```
