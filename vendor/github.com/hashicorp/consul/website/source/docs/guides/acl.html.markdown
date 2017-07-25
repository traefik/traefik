---
layout: "docs"
page_title: "ACL System"
sidebar_current: "docs-guides-acl"
description: |-
  Consul provides an optional Access Control List (ACL) system which can be used to control access to data and APIs. The ACL system is a Capability-based system that relies on tokens which can have fine grained rules applied to them. It is very similar to AWS IAM in many ways.
---

# ACL System

Consul provides an optional Access Control List (ACL) system which can be used to control
access to data and APIs. The ACL is
[Capability-based](https://en.wikipedia.org/wiki/Capability-based_security), relying
on tokens to which fine grained rules can be applied. It is very similar to
[AWS IAM](http://aws.amazon.com/iam/) in many ways.

## ACL System Overview

The ACL system is designed to be easy to use, fast to enforce, and flexible to new policies,
all while providing administrative insight.

#### ACL Tokens

The ACL system is based on tokens, which are managed by Consul operators via Consul's
[ACL API](/api/acl.html), or systems like
[HashiCorp's Vault](https://www.vaultproject.io/docs/secrets/consul/index.html).

Every token has an ID, name, type, and rule set. The ID is a randomly generated
UUID, making it infeasible to guess. The name is opaque to Consul and human readable.
The type is either "client" (meaning the token cannot modify ACL rules) or "management"
(meaning the token is allowed to perform all actions).

The token ID is passed along with each RPC request to the servers. Consul's
[HTTP endpoints](http://localhost:4567/api/index.html) can accept tokens via the `token`
query string parameter, or the `X-Consul-Token` request header. Consul's
[CLI commands](http://localhost:4567/docs/commands/index.html) can accept tokens via the
`token` argument, or the `CONSUL_HTTP_TOKEN` environment variable.

If no token is provided, the rules associated with a special, configurable anonymous
token are automatically applied. The anonymous token is managed using the
[ACL API](/api/acl.html) like any other ACL token, but using `anonymous` for the ID.

#### ACL Rules and Scope

Tokens are bound to a set of rules that control which Consul resources the token
has access to. Policies can be defined in either a whitelist or blacklist mode
depending on the configuration of
[`acl_default_policy`](/docs/agent/options.html#acl_default_policy). If the default
policy is to "deny all" actions, then token rules can be set to whitelist specific
actions. In the inverse, the "allow all" default behavior is a blacklist where rules
are used to prohibit actions. By default, Consul will allow all actions.

The following table summarizes the ACL policies that are available for constructing
rules:

| Policy                   | Scope |
| ------------------------ | ----- |
| [`agent`](#agent-rules)          | Utility operations in the [Agent API](/api/agent.html), other than service and check registration |
| [`event`](#event-rules)          | Listing and firing events in the [Event API](/api/event.html) |
| [`key`](#key-value-rules)        | Key/value store operations in the [KV Store API](/api/kv.html) |
| [`keyring`](#keyring-rules)      | Keyring operations in the [Keyring API](/api/operator/keyring.html) |
| [`node`](#node-rules)            | Node-level catalog operations in the [Catalog API](/api/catalog.html), [Health API](/api/health.html), [Network Coordinate API](/api/coordinate.html), and [Agent API](/api/agent.html) |
| [`operator`](#operator-rules)    | Cluster-level operations in the [Operator API](/api/operator.html), other than the [Keyring API](/api/operator/keyring.html) |
| [`query`](#prepared-query-rules) | Prepared query operations in the [Prepared Query API](/api/query.html)
| [`service`](#service-rules)      | Service-level catalog operations in the [Catalog API](/api/catalog.html), [Health API](/api/health.html), and [Agent API](/api/agent.html) |
| [`session`](#session-rules)      | Session operations in the [Session API](/api/session.html) |

Since Consul snapshots actually contain ACL tokens, the
[Snapshot API](/api/snapshot.html) requires a management token for snapshot operations
and does not use a special policy.

The following resources are not covered by ACL policies:

1. The [Status API](/api/status.html) is used by servers when bootstrapping and exposes
basic IP and port information about the servers, and does not allow modification
of any state.

2. The datacenter listing operation of the
[Catalog API](/api/catalog.html#list-datacenters) similarly exposes the names of known
Consul datacenters, and does not allow modification of any state.

Constructing rules from these policies is covered in detail in the
[Rule Specification](#rule-specification) section below.

#### ACL Datacenter

All nodes (clients and servers) must be configured with an
[`acl_datacenter`](/docs/agent/options.html#acl_datacenter) which enables ACL
enforcement but also specifies the authoritative datacenter. Consul relies on
[RPC forwarding](/docs/internals/architecture.html) to support multi-datacenter
configurations. However, because requests can be made across datacenter boundaries,
ACL tokens must be valid globally. To avoid consistency issues, a single datacenter
is considered authoritative and stores the canonical set of tokens.

When a request is made to an agent in a non-authoritative datacenter, it must be
resolved into the appropriate policy. This is done by reading the token from the
authoritative server and caching the result for a configurable
[`acl_ttl`](/docs/agent/options.html#acl_ttl). The implication of caching is that
the cache TTL is an upper bound on the staleness of policy that is enforced. It is
possible to set a zero TTL, but this has adverse performance impacts, as every
request requires refreshing the policy via an RPC call.

#### Enabling ACLs

Enabling ACLs is done by setting up the following configuration options. These are
marked as to whether they are set on servers, clients, or both.

| Configuration Option | Servers | Clients | Purpose |
| -------------------- | ------- | ------- | ------- |
| [`acl_datacenter`](/docs/agent/options.html#acl_datacenter) | `REQUIRED` | `REQUIRED` | Master control that enables ACLs by defining the authoritative Consul datacenter for ACLs |
| [`acl_default_policy`](/docs/agent/options.html#acl_default_policy) | `OPTIONAL` | `N/A` | Determines whitelist or blacklist mode |
| [`acl_down_policy`](/docs/agent/options.html#acl_down_policy) | `OPTIONAL` | `OPTIONAL` | Determines what to when the ACL datacenter is offline |
| [`acl_ttl`](/docs/agent/options.html#acl_ttl) | `OPTIONAL` | `OPTIONAL` | Determines time-to-live for cached ACLs |

There are some additional configuration items related to [ACL replication](#replication) and
[Version 8 ACL support](#version_8_acls). These are discussed in those respective sections
below.

A number of special tokens can also be configured which allow for bootstrapping the ACL
system, or accessing Consul in special situations:

| Special Token | Servers | Clients | Purpose |
| ------------- | ------- | ------- | ------- |
| [`acl_agent_master_token`](/docs/agent/options.html#acl_agent_master_token) | `OPTIONAL` | `OPTIONAL` | Special token that can be used to access [Agent API](/api/agent.html) when the ACL datacenter isn't available, or servers are offline (for clients); used for setting up the cluster such as doing initial join operations |
| [`acl_agent_token`](/docs/agent/options.html#acl_agent_token) | `OPTIONAL` | `OPTIONAL` | Special token that is used for an agent's internal operations with the [Catalog API](/api/catalog.html); this needs to have at least `node` policy access so the agent can self update its registration information |
| [`acl_master_token`](/docs/agent/options.html#acl_master_token) | `REQUIRED` | `N/A` | Special token used to bootstrap the ACL system, see details below. |
| [`acl_token`](/docs/agent/options.html#acl_token) | `OPTIONAL` | `OPTIONAL` | Default token to use for client requests where no token is supplied; this is often configured with read-only access to services to enable DNS service discovery on agents |

Bootstrapping the ACL system is done by providing an initial
[`acl_master_token`](/docs/agent/options.html#acl_master_token) which will be created
as a "management" type token if it does not exist. The
[`acl_master_token`](/docs/agent/options.html#acl_master_token) is only installed when
a server acquires cluster leadership. If you would like to install or change the
[`acl_master_token`](/docs/agent/options.html#acl_master_token), set the new value for
[`acl_master_token`](/docs/agent/options.html#acl_master_token) in the configuration
for all servers. Once this is done, restart the current leader to force a leader election.

Once the ACL system is bootstrapped, ACL tokens can be managed through the
[ACL API](/api/acl.html).

## Rule Specification

A core part of the ACL system is the rule language which is used to describe the policy
that must be enforced. Most of the ACL rules are prefix-based, allowing operators to
define different namespaces within Consul's resource areas like the catalog and key/value
store, in order to delegate responsibility for these namespaces. Policies can have several
dispositions:

* `read`: allow the resource to be read but not modified
* `write`: allow the resource to be read and modified
* `deny`: do not allow the resource to be read or modified

With prefix-based rules, the most specific prefix match determines the action. This
allows for flexible rules like an empty prefix to allow read-only access to all
resources, along with some specific prefixes that allow write access or that are
denied all access.

We make use of the
[HashiCorp Configuration Language (HCL)](https://github.com/hashicorp/hcl/) to specify
rules. This language is human readable and interoperable with JSON making it easy to
machine-generate. Rules can make use of one or more policies.

Specification in the HCL format looks like:

```text
# These control access to the key/value store.
key "" {
  policy = "read"
}
key "foo/" {
  policy = "write"
}
key "foo/private/" {
  policy = "deny"
}

# This controls access to cluster-wide Consul operator information.
operator = "read"
```

This is equivalent to the following JSON input:

```javascript
{
  "key": {
    "": {
      "policy": "read"
    },
    "foo/": {
      "policy": "write"
    },
    "foo/private": {
      "policy": "deny"
    }
  },
  "operator": "read"
}
```

The [ACL API](/api/acl.html) allows either HCL or JSON to be used to define the content
of the rules section.

Here's a sample request using the HCL form:

```text
$ curl \
    --request PUT \
    --data \
'{
  "Name": "my-app-token",
  "Type": "client",
  "Rules": "key \"\" { policy = \"read\" } key \"foo/\" { policy = \"write\" } key \"foo/private/\" { policy = \"deny\" } operator = \"read\""
}' https://consul.rocks/v1/acl/create?token=<management token>
```

Here's an equivalent request using the JSON form:

```text
$ curl \
    --request PUT \
    --data \
'{
  "Name": "my-app-token",
  "Type": "client",
  "Rules": "{\"key\":{\"\":{\"policy\":\"read\"},\"foo/\":{\"policy\":\"write\"},\"foo/private\":{\"policy\":\"deny\"}},\"operator\":\"read\"}"
}' https://consul.rocks/v1/acl/create?token=<management token>
```

On success, the token ID is returned:

```json
{
  "ID": "adf4238a-882b-9ddc-4a9d-5b6758e4159e"
}
```

This token ID can then be passed into Consul's HTTP APIs via the `token`
query string parameter, or the `X-Consul-Token` request header, or Consul's
CLI commands via the `token` argument, or the `CONSUL_HTTP_TOKEN` environment
variable.

#### Agent Rules

The `agent` policy controls access to the utility operations in the [Agent API](/api/agent.html),
such as join and leave. All of the catalog-related operations are covered by the [`node`](#node-rules)
and [`service`](#service-rules) policies instead.

Agent rules look like this:

```text
agent "" {
  policy = "read"
}
agent "foo" {
  policy = "write"
}
agent "bar" {
  policy = "deny"
}
```

Agent rules are keyed by the node name prefix they apply to, using the longest prefix match rule. In
the example above, the rules allow read-only access to any node name with the empty prefix, allow
read-write access to any node name that starts with "foo", and deny all access to any node name that
starts with "bar".

Since [Agent API](/api/agent.html) utility operations may be required before an agent is joined to
a cluster, or during an outage of the Consul servers or ACL datacenter, a special token may be
configured with [`acl_agent_master_token`](/docs/agent/options.html#acl_agent_master_token) to allow
write access to these operations even if no ACL resolution capability is available.

#### Event Rules

The `event` policy controls access to event operations in the [Event API](/api/event.html), such as
firing events and listing events.

Event rules look like this:

```text
event "" {
  policy = "read"
}
event "deploy" {
  policy = "write"
}
```

Event rules are keyed by the event name prefix they apply to, using the longest prefix match rule.
In the example above, the rules allow read-only access to any event, and firing of any event that
starts with "deploy".

The [`consul exec`](/docs/commands/exec.html) command uses events with the "_rexec" prefix during
operation, so to enable this feature in a Consul environment with ACLs enabled, you will need to
give agents a token with access to this event prefix, in addition to configuring
[`disable_remote_exec`](/docs/agent/options.html#disable_remote_exec) to `false`.

#### Key/Value Rules

The `key` policy controls access to key/value store operations in the [KV API](/api/kv.html]. Key
rules look like this:

```text
key "" {
  policy = "read"
}
key "foo" {
  policy = "write"
}
key "bar" {
  policy = "deny"
}
```

Key rules are keyed by the key name prefix they apply to, using the longest prefix match rule. In
the example above, the rules allow read-only access to any key name with the empty prefix, allow
read-write access to any key name that starts with "foo", and deny all access to any key name that
starts with "bar".

#### Keyring Rules

The `keyring` policy controls access to keyring operations in the
[Keyring API](/api/operator/keyring.html).

Keyring rules look like this:

```text
keyring = "write"
```

There's only one keyring policy allowed per rule set, and its value is set to one of the policy
dispositions. In the example above, the keyring may be read and updated.

#### Node Rules

The `node` policy controls node-level registration and read access to the [Catalog API](/api/catalog.html),
service discovery with the [Health API](/api/health.html), and filters results in [Agent API](/api/agent.html)
operations like fetching the list of cluster members.

Node rules look like this:

```text
node "" {
  policy = "read"
}
node "app" {
  policy = "write"
}
node "admin" {
  policy = "deny"
}
```

Node rules are keyed by the node name prefix they apply to, using the longest prefix match rule. In
the example above, the rules allow read-only access to any node name with the empty prefix, allow
read-write access to any node name that starts with "app", and deny all access to any node name that
starts with "admin".

Agents need to be configured with an [`acl_agent_token`](/docs/agent/options.html#acl_agent_token)
with at least "write" privileges to their own node name in order to register their information with
the catalog, such as node metadata and tagged addresses. If this is configured incorrectly, the agent
will print an error to the console when it tries to sync its state with the catalog.

Consul's DNS interface is also affected by restrictions on node rules. If the
[`acl_token`](/docs/agent/options.html#acl_token) used by the agent does not have "read" access to a
given node, then the DNS interface will return no records when queried for it.

When reading from the catalog or retrieving information from the health endpoints, node rules are
used to filter the results of the query. This allows for configurations where a token has access
to a given service name, but only on an allowed subset of node names.

Node rules come into play when using the [Agent API](/api/agent.html) to register node-level
checks. The agent will check tokens locally as a check is registered, and Consul also performs
periodic [anti-entropy](/docs/internals/anti-entropy.html) syncs, which may require an
ACL token to complete. To accommodate this, Consul provides two methods of configuring ACL tokens
to use for registration events:

1. Using the [acl_token](/docs/agent/options.html#acl_token) configuration
   directive. This allows a single token to be configured globally and used
   during all check registration operations.
2. Providing an ACL token with service and check definitions at
   registration time. This allows for greater flexibility and enables the use
   of multiple tokens on the same agent. Examples of what this looks like are
   available for both [services](/docs/agent/services.html) and
   [checks](/docs/agent/checks.html). Tokens may also be passed to the
   [HTTP API](/api/index.html) for operations that require them.

#### Operator Rules

The `operator` policy controls access to cluster-level operations in the
[Operator API](/api/operator.html), other than the [Keyring API](/api/operator/keyring.html).

Operator rules look like this:

```text
operator = "read"
```

There's only one operator policy allowed per rule set, and its value is set to one of the policy
dispositions. In the example above, the token could be used to query the operator endpoints for
diagnostic purposes but not make any changes.

#### Prepared Query Rules

The `query` policy controls access to create, update, and delete prepared queries in the
[Prepared Query API](/api/query.html). Executing queries is subject to `node` and `service`
policies, as will be explained below.

Query rules look like this:

```text
query "" {
  policy = "read"
}
query "foo" {
  policy = "write"
}
```

Query rules are keyed by the query name prefix they apply to, using the longest prefix match rule. In
the example above, the rules allow read-only access to any query name with the empty prefix, and allow
read-write access to any query name that starts with "foo". This allows control of the query namespace
to be delegated based on ACLs.

There are a few variations when using ACLs with prepared queries, each of which uses ACLs in one of two
ways: open, protected by unguessable IDs or closed, managed by ACL policies. These variations are covered
here, with examples:

* Static queries with no `Name` defined are not controlled by any ACL policies.
  These types of queries are meant to be ephemeral and not shared to untrusted
  clients, and they are only reachable if the prepared query ID is known. Since
  these IDs are generated using the same random ID scheme as ACL Tokens, it is
  infeasible to guess them. When listing all prepared queries, only a management
  token will be able to see these types, though clients can read instances for
  which they have an ID. An example use for this type is a query built by a
  startup script, tied to a session, and written to a configuration file for a
  process to use via DNS.

* Static queries with a `Name` defined are controlled by the `query` ACL policy.
  Clients are required to have an ACL token with a prefix sufficient to cover
  the name they are trying to manage, with a longest prefix match providing a
  way to define more specific policies. Clients can list or read queries for
  which they have "read" access based on their prefix, and similar they can
  update any queries for which they have "write" access. An example use for
  this type is a query with a well-known name (eg. `prod-master-customer-db`)
  that is used and known by many clients to provide geo-failover behavior for
  a database.

* [Template queries](/api/query.html#templates)
  queries work like static queries with a `Name` defined, except that a catch-all
  template with an empty `Name` requires an ACL token that can write to any query
  prefix.

When prepared queries are executed via DNS lookups or HTTP requests, the ACL
checks are run against the service being queried, similar to how ACLs work with
other service lookups. There are several ways the ACL token is selected for this
check:

* If an ACL Token was captured when the prepared query was defined, it will be
  used to perform the service lookup. This allows queries to be executed by
  clients with lesser or even no ACL Token, so this should be used with care.

* If no ACL Token was captured, then the client's ACL Token will be used to
  perform the service lookup.

* If no ACL Token was captured and the client has no ACL Token, then the
  anonymous token will be used to perform the service lookup.

In the common case, the ACL Token of the invoker is used
to test the ability to look up a service. If a `Token` was specified when the
prepared query was created, the behavior changes and now the captured
ACL Token set by the definer of the query is used when looking up a service.

Capturing ACL Tokens is analogous to
[PostgreSQL’s](http://www.postgresql.org/docs/current/static/sql-createfunction.html)
`SECURITY DEFINER` attribute which can be set on functions, and using the client's ACL
Token is similar to the complementary `SECURITY INVOKER` attribute.

Prepared queries were originally introduced in Consul 0.6.0, and ACL behavior remained
unchanged through version 0.6.3, but was then changed to allow better management of the
prepared query namespace.

These differences are outlined in the table below:

<table class="table table-bordered table-striped">
  <tr>
    <th>Operation</th>
    <th>Version <= 0.6.3 </th>
    <th>Version > 0.6.3 </th>
  </tr>
  <tr>
    <td>Create static query without `Name`</td>
    <td>The ACL Token used to create the prepared query is checked to make sure it can access the service being queried. This token is captured as the `Token` to use when executing the prepared query.</td>
    <td>No ACL policies are used as long as no `Name` is defined. No `Token` is captured by default unless specifically supplied by the client when creating the query.</td>
  </tr>
  <tr>
    <td>Create static query with `Name`</td>
    <td>The ACL Token used to create the prepared query is checked to make sure it can access the service being queried. This token is captured as the `Token` to use when executing the prepared query.</td>
    <td>The client token's `query` ACL policy is used to determine if the client is allowed to register a query for the given `Name`. No `Token` is captured by default unless specifically supplied by the client when creating the query.</td>
  </tr>
  <tr>
    <td>Manage static query without `Name`</td>
    <td>The ACL Token used to create the query, or a management token must be supplied in order to perform these operations.</td>
    <td>Any client with the ID of the query can perform these operations.</td>
  </tr>
  <tr>
    <td>Manage static query with a `Name`</td>
    <td>The ACL token used to create the query, or a management token must be supplied in order to perform these operations.</td>
    <td>Similar to create, the client token's `query` ACL policy is used to determine if these operations are allowed.</td>
  </tr>
  <tr>
    <td>List queries</td>
    <td>A management token is required to list any queries.</td>
    <td>The client token's `query` ACL policy is used to determine which queries they can see. Only management tokens can see prepared queries without `Name`.</td>
  </tr>
  <tr>
    <td>Execute query</td>
    <td>Since a `Token` is always captured when a query is created, that is used to check access to the service being queried. Any token supplied by the client is ignored.</td>
    <td>The captured token, client's token, or anonymous token is used to filter the results, as described above.</td>
  </tr>
</table>

#### Service Rules

The `service` policy controls service-level registration and read access to the [Catalog API](/api/catalog.html)
and service discovery with the [Health API](/api/health.html).

Service rules look like this:

```text
service "" {
  policy = "read"
}
service "app" {
  policy = "write"
}
service "admin" {
  policy = "deny"
}
```

Service rules are keyed by the service name prefix they apply to, using the longest prefix match rule. In
the example above, the rules allow read-only access to any service name with the empty prefix, allow
read-write access to any service name that starts with "app", and deny all access to any service name that
starts with "admin".

Consul's DNS interface is affected by restrictions on service rules. If the
[`acl_token`](/docs/agent/options.html#acl_token) used by the agent does not have "read" access to a
given service, then the DNS interface will return no records when queried for it.

When reading from the catalog or retrieving information from the health endpoints, service rules are
used to filter the results of the query.

Service rules come into play when using the [Agent API](/api/agent.html) to register services or
checks. The agent will check tokens locally as a service or check is registered, and Consul also
performs periodic [anti-entropy](/docs/internals/anti-entropy.html) syncs, which may require an
ACL token to complete. To accommodate this, Consul provides two methods of configuring ACL tokens
to use for registration events:

1. Using the [acl_token](/docs/agent/options.html#acl_token) configuration
   directive. This allows a single token to be configured globally and used
   during all service and check registration operations.
2. Providing an ACL token with service and check definitions at
   registration time. This allows for greater flexibility and enables the use
   of multiple tokens on the same agent. Examples of what this looks like are
   available for both [services](/docs/agent/services.html) and
   [checks](/docs/agent/checks.html). Tokens may also be passed to the
   [HTTP API](/api/index.html) for operations that require them.

#### Session Rules

The `session` policy controls access to [Session API](/api/session.html)] operations.

Session rules look like this:

```text
session "" {
  policy = "read"
}
session "app" {
  policy = "write"
}
session "admin" {
  policy = "deny"
}
```

Session rules are keyed by the node name prefix they apply to, using the longest prefix match rule. In
the example above, the rules allow read-only access to sessions on node name with the empty prefix, allow
creating sessions on any node name that starts with "app", and deny all access to any sessions on a node
name that starts with "admin".

## Advanced Topics

<a name="replication"></a>
#### Outages and ACL Replication

The Consul ACL system is designed with flexible rules to accommodate for an outage
of the [`acl_datacenter`](/docs/agent/options.html#acl_datacenter) or networking
issues preventing access to it. In this case, it may be impossible for
agents in non-authoritative datacenters to resolve tokens. Consul provides
a number of configurable [`acl_down_policy`](/docs/agent/options.html#acl_down_policy)
choices to tune behavior. It is possible to deny or permit all actions or to ignore
cache TTLs and enter a fail-safe mode. The default is to ignore cache TTLs
for any previously resolved tokens and to deny any uncached tokens.

Consul 0.7 added an ACL Replication capability that can allow non-authoritative
datacenter agents to resolve even uncached tokens. This is enabled by setting an
[`acl_replication_token`](/docs/agent/options.html#acl_replication_token) in the
configuration on the servers in the non-authoritative datacenters. With replication
enabled, the servers will maintain a replica of the authoritative datacenter's full
set of ACLs on the non-authoritative servers. The ACL replication token needs to be
a valid ACL token with management privileges, it can also be the same as the master
ACL token.

Replication occurs with a background process that looks for new ACLs approximately
every 30 seconds. Replicated changes are written at a rate that's throttled to
100 updates/second, so it may take several minutes to perform the initial sync of
a large set of ACLs.

If there's a partition or other outage affecting the authoritative datacenter,
and the [`acl_down_policy`](/docs/agent/options.html#acl_down_policy)
is set to "extend-cache", tokens will be resolved during the outage using the
replicated set of ACLs. An [ACL replication status](/api/acl.html#acl_replication_status)
endpoint is available to monitor the health of the replication process.

Locally-resolved ACLs will be cached using the [`acl_ttl`](/docs/agent/options.html#acl_ttl)
setting of the non-authoritative datacenter, so these entries may persist in the
cache for up to the TTL, even after the authoritative datacenter comes back online.

ACL replication can also be used to migrate ACLs from one datacenter to another
using a process like this:

1. Enable ACL replication in all datacenters to allow continuation of service
during the migration, and to populate the target datacenter. Verify replication
is healthy and caught up to the current ACL index in the target datacenter
using the [ACL replication status](/api/acl.html#acl_replication_status)
endpoint.
2. Turn down the old authoritative datacenter servers.
3. Rolling restart the agents in the target datacenter and change the
`acl_datacenter` servers to itself. This will automatically turn off
replication and will enable the datacenter to start acting as the authoritative
datacenter, using its replicated ACLs from before.
3. Rolling restart the agents in other datacenters and change their `acl_datacenter`
configuration to the target datacenter.

<a name="version_8_acls"></a>
#### Complete ACL Coverage in Consul 0.8

Consul 0.8 added many more ACL policy types and brought ACL enforcement to Consul
agents for the first time. To ease the transition to Consul 0.8 for existing ACL
users, there's a configuration option to disable these new features. To disable
support for these new ACLs, set the
[`acl_enforce_version_8`](/docs/agent/options.html#acl_enforce_version_8) configuration
option to `false` on Consul clients and servers.

Here's a summary of the new features:

* Agents now check [`node`](#node-rules) and [`service`](#service-rules) ACL policies
  for catalog-related operations in `/v1/agent` endpoints, such as service and check
  registration and health check updates.
* Agents enforce a new [`agent`](#agent-rules) ACL policy for utility operations in
  `/v1/agent` endpoints, such as joins and leaves.
* A new [`node`](#node-rules) ACL policy is enforced throughout Consul, providing a
  mechanism to restrict registration and discovery of nodes by name. This also applies
  to service discovery, so provides an additional dimension for controlling access to
  services.
* A new [`session`](#session-rules) ACL policy controls the ability to create session
  objects by node name.
* Anonymous prepared queries (non-templates without a `Name`) now require a valid
  session, which ties their creation to the new [`session`](#session-rules) ACL policy.
* The existing [`event`](#event-rules) ACL policy has been applied to the
  `/v1/event/list` endpoint.

Two new configuration options are used once version 8 ACLs are enabled:

* [`acl_agent_master_token`](/docs/agent/options.html#acl_agent_master_token) is used as
  a special access token that has `agent` ACL policy `write` privileges on each agent where
  it is configured. This token should only be used by operators during outages when Consul
  servers aren't available to resolve ACL tokens. Applications should use regular ACL
  tokens during normal operation.
* [`acl_agent_token`](/docs/agent/options.html#acl_agent_token) is used internally by
  Consul agents to perform operations to the service catalog when registering themselves
  or sending network coordinates to the servers. This token must at least have `node` ACL
  policy `write` access to the node name it will register as in order to register any
  node-level information like metadata or tagged addresses.

Since clients now resolve ACLs locally, the [`acl_down_policy`](/docs/agent/options.html#acl_down_policy)
now applies to Consul clients as well as Consul servers. This will determine what the
client will do in the event that the servers are down.

Consul clients must have [`acl_datacenter`](/docs/agent/options.html#acl_datacenter) configured
in order to enable agent-level ACL features. If this is set, the agents will contact the Consul
servers to determine if ACLs are enabled at the cluster level. If they detect that ACLs are not
enabled, they will check at most every 2 minutes to see if they have become enabled, and will
start enforcing ACLs automatically. If an agent has an `acl_datacenter` defined, operators will
need to use the [`acl_agent_master_token`](/docs/agent/options.html#acl_agent_master_token) to
perform agent-level operations if the Consul servers aren't present (such as for a manual join
to the cluster), unless the [`acl_down_policy`](/docs/agent/options.html#acl_down_policy) on the
agent is set to "allow".

Non-server agents do not need to have the
[`acl_master_token`](/docs/agent/options.html#acl_agent_master_token) configured; it is not
used by agents in any way.
