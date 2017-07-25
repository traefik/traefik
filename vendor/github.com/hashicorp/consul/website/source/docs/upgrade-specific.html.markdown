---
layout: "docs"
page_title: "Upgrading Specific Versions"
sidebar_current: "docs-upgrading-specific"
description: |-
  Specific versions of Consul may have additional information about the upgrade process beyond the standard flow.
---

# Upgrading Specific Versions

The [upgrading page](/docs/upgrading.html) covers the details of doing
a standard upgrade. However, specific versions of Consul may have more
details provided for their upgrades as a result of new features or changed
behavior. This page is used to document those details separately from the
standard upgrade flow.

## Consul 0.8.0

#### Command-Line Interface RPC Deprecation

The RPC client interface has been removed. All CLI commands that used RPC and the
`-rpc-addr` flag to communicate with Consul have been converted to use the HTTP API
and the appropriate flags for it, and the `rpc` field has been removed from the port
and address binding configs. You will need to remove these fields from your config files
and update any scripts that passed a custom `-rpc-addr` to the following commands:

* `force-leave`
* `info`
* `join`
* `keyring`
* `leave`
* `members`
* `monitor`
* `reload`

#### Version 8 ACLs Are Now Opt-Out

The [`acl_enforce_version_8`](/docs/agent/options.html#acl_enforce_version_8) configuration now defaults to `true` to enable [full version 8 ACL support](/docs/guides/acl.html#version_8_acls) by default. If you are upgrading an existing cluster with ACLs enabled, you will need to set this to `false` during the upgrade on **both Consul agents and Consul servers**. Version 8 ACLs were also changed so that [`acl_datacenter`](/docs/agent/options.html#acl_datacenter) must be set on agents in order to enable the agent-side enforcement of ACLs. This makes for a smoother experience in clusters where ACLs aren't enabled at all, but where the agents would have to wait to contact a Consul server before learning that.

#### Remote Exec Is Now Opt-In

The default for [`disable_remote_exec`](/docs/agent/options.html#disable_remote_exec) was
changed to "true", so now operators need to opt-in to having agents support running
commands remotely via [`consul exec`](/docs/commands/exec.html).

#### Raft Protocol Version Compatibility

When upgrading to Consul 0.8.0 from a version lower than 0.7.0, users will need to
set the [`-raft-protocol`](/docs/agent/options.html#_raft_protocol) option to 1 in
order to maintain backwards compatibility with the old servers during the upgrade.
After the servers have been migrated to version 0.8.0, `-raft-protocol` can be moved
up to 2 and the servers restarted to match the default.

The Raft protocol must be stepped up in this way; only adjacent version numbers are
compatible (for example, version 1 cannot talk to version 3). Here is a table of the
Raft Protocol versions supported by each Consul version:

<table class="table table-bordered table-striped">
  <tr>
    <th>Version</th>
    <th>Supported Raft Protocols</th>
  </tr>
  <tr>
    <td>0.6 and earlier</td>
    <td>0</td>
  </tr>
  <tr>
    <td>0.7</td>
    <td>1</td>
  </tr>
  <tr>
    <td>0.8</td>
    <td>1, 2, 3</td>
  </tr>
</table>

In order to enable all [Autopilot](/docs/guides/autopilot.html) features, all servers
in a Consul cluster must be running with Raft protocol version 3 or later.

## Consul 0.7.1

#### Child Process Reaping

Child process reaping support has been removed, along with the `reap` configuration option. Reaping is also done via [dumb-init](https://github.com/Yelp/dumb-init) in the [Consul Docker image](https://github.com/hashicorp/docker-consul), so removing it from Consul itself simplifies the code and eases future maintenance for Consul. If you are running Consul as PID 1 in a container you will need to arrange for a wrapper process to reap child processes.

#### DNS Resiliency Defaults

The default for [`max_stale`](/docs/agent/options.html#max_stale) has been increased from 5 seconds to a near-indefinite threshold (10 years) to allow DNS queries to continue to be served in the event of a long outage with no leader. A new telemetry counter was added at `consul.dns.stale_queries` to track when agents serve DNS queries that are stale by more than 5 seconds.

## Consul 0.7

Consul version 0.7 is a very large release with many important changes. Changes
to be aware of during an upgrade are categorized below.

#### Performance Timing Defaults and Tuning

Consul 0.7 now defaults the DNS configuration to allow for stale queries by defaulting
[`allow_stale`](/docs/agent/options.html#allow_stale) to true for better utilization
of available servers. If you want to retain the previous behavior, set the following
configuration:

```javascript
{
  "dns_config": {
    "allow_stale": false
  }
}
```

Consul also 0.7 introduced support for tuning Raft performance using a new
[performance configuration block](/docs/agent/options.html#performance). Also,
the default Raft timing is set to a lower-performance mode suitable for
[minimal Consul servers](/docs/guides/performance.html#minumum).

To continue to use the high-performance settings that were the default prior to
Consul 0.7 (recommended for production servers), add the following configuration
to all Consul servers when upgrading:

```javascript
{
  "performance": {
    "raft_multiplier": 1
  }
}
```

See the [Server Performance](/docs/guides/performance.html) guide for more details.

#### Leave-Related Configuration Defaults

The default behavior of [`leave_on_terminate`](/docs/agent/options.html#leave_on_terminate)
and [`skip_leave_on_interrupt`](/docs/agent/options.html#skip_leave_on_interrupt)
are now dependent on whether or not the agent is acting as a server or client:

* For servers, `leave_on_terminate` defaults to "false" and `skip_leave_on_interrupt`
defaults to "true".

* For clients, `leave_on_terminate` defaults to "true" and `skip_leave_on_interrupt`
defaults to "false".

These defaults are designed to be safer for servers so that you must explicitly
configure them to leave the cluster. This also results in a better experience for
clients, especially in cloud environments where they may be created and destroyed
often and users prefer not to wait for the 72 hour reap time for cleanup.

#### Dropped Support for Protocol Version 1

Consul version 0.7 dropped support for protocol version 1, which means it
is no longer compatible with versions of Consul prior to 0.3. You will need
to upgrade all agents to a newer version of Consul before upgrading to Consul
0.7.

#### Prepared Query Changes

Consul version 0.7 adds a feature which allows prepared queries to store a
[`Near` parameter](/api/query.html#near) in the query definition
itself. This feature enables using the distance sorting features of prepared
queries without explicitly providing the node to sort near in requests, but
requires the agent servicing a request to send additional information about
itself to the Consul servers when executing the prepared query. Agents prior
to 0.7 do not send this information, which means they are unable to properly
execute prepared queries configured with a `Near` parameter. Similarly, any
server nodes prior to version 0.7 are unable to store the `Near` parameter,
making them unable to properly serve requests for prepared queries using the
feature. It is recommended that all agents be running version 0.7 prior to
using this feature.

#### WAN Address Translation in HTTP Endpoints

Consul version 0.7 added support for translating WAN addresses in certain
[HTTP endpoints](/docs/agent/options.html#translate_wan_addrs). The servers
and the agents need to be running version 0.7 or later in order to use this
feature.

These translated addresses could break HTTP endpoint consumers that are
expecting local addresses, so a new [`X-Consul-Translate-Addresses`](/api/index.html#translate_header)
header was added to allow clients to detect if translation is enabled for HTTP
responses. A "lan" tag was added to `TaggedAddresses` for clients that need
the local address regardless of translation.

#### Outage Recovery and `peers.json` Changes

The `peers.json` file is no longer present by default and is only used when
performing recovery. This file will be deleted after Consul starts and ingests
the file. Consul 0.7 also uses a new, automatically-created raft/peers.info file
to avoid ingesting the `peers.json` file on the first start after upgrading (the
`peers.json` file is simply deleted on the first start after upgrading).

Please be sure to review the [Outage Recovery Guide](/docs/guides/outage.html)
before upgrading for more details.

## Consul 0.6.4

Consul 0.6.4 made some substantial changes to how ACLs work with prepared
queries. Existing queries will execute with no changes, but there are important
differences to understand about how prepared queries are managed before you
upgrade. In particular, prepared queries with no `Name` defined will no longer
require any ACL to manage them, and prepared queries with a `Name` defined are
now governed by a new `query` ACL policy that will need to be configured
after the upgrade.

See the [ACL Guide](/docs/guides/acl.html#prepared_query_acls) for more details
about the new behavior and how it compares to previous versions of Consul.

## Consul 0.6

Consul version 0.6 is a very large release with many enhancements and
optimizations. Changes to be aware of during an upgrade are categorized below.

#### Data Store Changes

Consul changed the format used to store data on the server nodes in version 0.5
(see 0.5.1 notes below for details). Previously, Consul would automatically
detect data directories using the old LMDB format, and convert them to the newer
BoltDB format. This automatic upgrade has been removed for Consul 0.6, and
instead a safeguard has been put in place which will prevent Consul from booting
if the old directory format is detected.

It is still possible to migrate from a 0.5.x version of Consul to 0.6+ using the
[consul-migrate](https://github.com/hashicorp/consul-migrate) CLI utility. This
is the same tool that was previously embedded into Consul. See the
[releases](https://github.com/hashicorp/consul-migrate/releases) page for
downloadable versions of the tool.

Also, in this release Consul switched from LMDB to a fully in-memory database for
the state store. Because LMDB is a disk-based backing store, it was able to store
more data than could fit in RAM in some cases (though this is not a recommended
configuration for Consul). If you have an extremely large data set that won't fit
into RAM, you may encounter issues upgrading to Consul 0.6.0 and later. Consul
should be provisioned with physical memory approximately 2X the data set size to
allow for bursty allocations and subsequent garbage collection.

#### ACL Enhancements

Consul 0.6 introduces enhancements to the ACL system which may require special
handling:

* Service ACLs are enforced during service discovery (REST + DNS)

Previously, service discovery was wide open, and any client could query
information about any service without providing a token. Consul now requires
read-level access at a minimum when ACLs are enabled to return service
information over the REST or DNS interfaces. If clients depend on an open
service discovery system, then the following should be added to all ACL tokens
which require it:

    # Enable discovery of all services
    service "" {
        policy = "read"
    }

When the DNS interface is queried, the agent's
[`acl_token`](/docs/agent/options.html#acl_token) is used, so be sure
that token has sufficient privileges to return the DNS records you
expect to retrieve from it.

* Event and keyring ACLs

Similar to service discovery, the new event and keyring ACLs will block access
to these operations if the `acl_default_policy` is set to `deny`. If clients depend
on open access to these, then the following should be added to all ACL tokens which
require them:

    event "" {
      policy = "write"
    }

    keyring = "write"

Unfortunately, these are new ACLs for Consul 0.6, so they must be added after the
upgrade is complete.

#### Prepared Queries

Prepared queries introduce a new Raft log entry type that isn't supported on older
versions of Consul. It's important to not use the prepared query features of Consul
until all servers in a cluster have been upgraded to version 0.6.0.

#### Single Private IP Enforcement

Consul will refuse to start if there are multiple private IPs available, so
if this is the case you will need to configure Consul's advertise or bind addresses
before upgrading.

#### New Web UI File Layout

The release .zip file for Consul's web UI no longer contains a `dist` sub-folder;
everything has been moved up one level. If you have any automated scripts that
expect the old layout you may need to update them.

## Consul 0.5.1

Consul version 0.5.1 uses a different backend store for persisting the Raft
log. Because of this change, a data migration is necessary to move the log
entries out of LMDB and into the newer backend, BoltDB.

Consul version 0.5.1+ makes this transition seamless and easy. As a user, there
are no special steps you need to take. When Consul starts, it checks
for presence of the legacy LMDB data files, and migrates them automatically
if any are found. You will see a log emitted when Raft data is migrated, like
this:

```
==> Successfully migrated raft data in 5.839642ms
```

This automatic upgrade will only exist in Consul 0.5.1+ and it will
be removed starting with Consul 0.6.0+. It will still be possible to upgrade directly 
from pre-0.5.1 versions by using the consul-migrate utility, which is available on the
[Consul Tools page](/downloads_tools.html).

## Consul 0.5

Consul version 0.5 adds two features that complicate the upgrade process:

* ACL system includes service discovery and registration
* Internal use of tombstones to fix behavior of blocking queries
  in certain edge cases.

Users of the ACL system need to be aware that deploying Consul 0.5 will
cause service registration to be enforced. This means if an agent
attempts to register a service without proper privileges it will be denied.
If the `acl_default_policy` is "allow" then clients will continue to
work without an updated policy. If the policy is "deny", then all clients
will begin to have their registration rejected causing issues.

To avoid this situation, all the ACL policies should be updated to
add something like this:

    # Enable all services to be registered
    service "" {
        policy = "write"
    }

This will set the service policy to `write` level for all services.
The blank service name is the catch-all value. A more specific service
can also be specified:

    # Enable only the API service to be registered
    service "api" {
        policy = "write"
    }

The ACL policy can be updated while running 0.4, and enforcement will
being with the upgrade to 0.5. The policy updates will ensure the
availability of the cluster.

The second major change is the new internal command used for tombstones.
The details of the change are not important, however to function the leader
node will replicate a new command to its followers. Consul is designed
defensively, and when a command that is not recognized is received, the
server will panic. This is a purposeful design decision to avoid the possibility
of data loss, inconsistensies, or security issues caused by future incompatibility.

In practice, this means if a Consul 0.5 node is the leader, all of its
followers must also be running 0.5. There are a number of ways to do this
to ensure cluster availability:

* Add new 0.5 nodes, then remove the old servers. This will add the new
  nodes as followers, and once the old servers are removed, one of the
  0.5 nodes will become leader.

* Upgrade the followers first, then the leader last. Using `consul info`,
  you can determine which nodes are followers. Do an in-place upgrade
  on them first, and finally upgrade the leader last.

* Upgrade them in any order, but ensure all are done within 15 minutes.
  Even if the leader is upgraded to 0.5 first, as long as all of the followers
  are running 0.5 within 15 minutes there will be no issues.

Finally, even if any of the methods above are not possible or the process
fails for some reason, it is not fatal. The older version of the server
will simply panic and stop. At that point, you can upgrade to the new version
and restart the agent. There will be no data loss and the cluster will
resume operations.

