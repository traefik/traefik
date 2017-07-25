---
layout: api
page_title: HTTP API
sidebar_current: api-overview
description: |-
  Consul exposes a RESTful HTTP API to control almost every aspect of the
  Consul agent.
---

# HTTP API

The main interface to Consul is a RESTful HTTP API. The API can basic perform
CRUD operations on nodes, services, checks, configuration, and more.

## Version Prefix

All API routes are prefixed with `/v1/`.

This documentation is only for the v1 API.

~> **Backwards compatibility:** At the current version, Consul does not yet
promise backwards compatibility even with the v1 prefix. We'll remove this
warning when this policy changes. We expect to reach API stability by Consul
1.0.

## ACLs

Several endpoints in Consul use or require ACL tokens to operate. An agent
can be configured to use a default token in requests using the `acl_token`
configuration option. However, the token can also be specified per-request
by using the `X-Consul-Token` request header or the `token` query string
parameter. The request header takes precedence over the default token, and
the query string parameter takes precedence over everything.

## Authentication

When authentication is enabled, a Consul token should be provided to API
requests using the  `X-Consul-Token` header. This reduces the probability of the
token accidentally getting logged or  exposed. When using authentication,
clients should communicate via TLS.

Here is an example using `curl`:

```text
$ curl \
    --header "X-Consul-Token: abcd1234" \
    https://consul.rocks/v1/agent/members
```

Previously this was provided via a `?token=` query parameter. This functionality
exists on many endpoints for backwards compatibility, but its use is **highly
discouraged**, since it can show up in access logs as part of the URL.

## Blocking Queries

Many endpoints in Consul support a feature known as "blocking queries". A
blocking query is used to wait for a potential change using long polling. Not
all endpoints support blocking, but each endpoint uniquely documents its support
for blocking queries in the documentation.

Endpoints that support blocking queries return an HTTP header named
`X-Consul-Index`. This is a unique identifier representing the current state of
the requested resource.

On subsequent requests for this resource, the client can set the `index` query
string parameter to the value of `X-Consul-Index`, indicating that the client
wishes to wait for any changes subsequent to that index.

When this is provided, the HTTP request will "hang" until a change in the system
occurs, or the maximum timeout is reached. A critical note is that the return of
a blocking request is **no guarantee** of a change. It is possible that the
timeout was reached or that there was an idempotent write that does not affect
the result of the query.

In addition to `index`, endpoints that support blocking will also honor a `wait`
parameter specifying a maximum duration for the blocking request. This is
limited to 10 minutes. If not set, the wait time defaults to 5 minutes. This
value can be specified in the form of "10s" or "5m" (i.e., 10 seconds or 5
minutes, respectively). A small random amount of additional wait time is added
to the supplied maximum `wait` time to spread out the wake up time of any
concurrent requests. This adds up to `wait / 16` additional time to the maximum
duration.

## Consistency Modes

Most of the read query endpoints support multiple levels of consistency. Since
no policy will suit all clients' needs, these consistency modes allow the user
to have the ultimate say in how to balance the trade-offs inherent in a
distributed system.

The three read modes are:

- `default` - If not specified, the default is strongly consistent in almost all
  cases. However, there is a small window in which a new leader may be elected
  during which the old leader may service stale values. The trade-off is fast
  reads but potentially stale values. The condition resulting in stale reads is
  hard to trigger, and most clients should not need to worry about this case.
  Also, note that this race condition only applies to reads, not writes.

- `consistent` - This mode is strongly consistent without caveats. It requires
  that a leader verify with a quorum of peers that it is still leader. This
  introduces an additional round-trip to all server nodes. The trade-off is
  increased latency due to an extra round trip. Most clients should not use this
  unless they cannot tolerate a stale read.

- `stale` - This mode allows any server to service the read regardless of
  whether it is the leader. This means reads can be arbitrarily stale; however,
  results are generally consistent to within 50 milliseconds of the leader. The
  trade-off is very fast and scalable reads with a higher likelihood of stale
  values. Since this mode allows reads without a leader, a cluster that is
  unavailable will still be able to respond to queries.

To switch these modes, either the `stale` or `consistent` query parameters
should be provided on requests. It is an error to provide both.

To support bounding the acceptable staleness of data, responses provide the
`X-Consul-LastContact` header containing the time in milliseconds that a server
was last contacted by the leader node. The `X-Consul-KnownLeader` header also
indicates if there is a known leader. These can be used by clients to gauge the
staleness of a result and take appropriate action.

## Formatted JSON Output

By default, the output of all HTTP API requests is minimized JSON. If the client
passes `pretty` on the query string, formatted JSON will be returned.

## HTTP Methods

Consul's API aims to be RESTful, although there are some exceptions. The API
responds to the standard HTTP verbs GET, PUT, and DELETE. Each API method will
clearly document the verb(s) it responds to and the generated response. The same
path with different verbs may trigger different behavior. For example:

```text
PUT /v1/kv/foo
GET /v1/kv/foo
```

Even though these share a path, the `PUT` operation creates a new key whereas
the `GET` operation reads an existing key.

Here is the same example using `curl`:

```shell
$ curl \
    --request PUT \
    --data 'hello consul' \
    https://consul.rocks/v1/kv/foo
```

## Translated Addresses

Consul 0.7 added the ability to translate addresses in HTTP response based on
the configuration setting for
[`translate_wan_addrs`](/docs/agent/options.html#translate_wan_addrs). In order
to allow clients to know if address translation is in effect, the
`X-Consul-Translate-Addresses` header will be added if translation is enabled,
and will have a value of `true`. If translation is not enabled then this header
will not be present.
