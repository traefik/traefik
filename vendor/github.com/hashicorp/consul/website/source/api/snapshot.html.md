---
layout: api
page_title: Snapshot - HTTP API
sidebar_current: api-snapshot
description: |-
  The /snapshot endpoints save and restore Consul's server state for disaster
  recovery.
---

# Snapshot HTTP Endpoint

The `/snapshot` endpoints save and restore the state of the Consul
servers for disaster recovery. Snapshots include all state managed by Consul's
Raft [consensus protocol](/docs/internals/consensus.html).

## Generate Snapshot

This endpoint generates and returns an atomic, point-in-time snapshot of the
Consul server state.

Snapshots are exposed as gzipped tar archives which internally contain the Raft
metadata required to restore, as well as a binary serialized version of the
Consul server state. The contents are covered internally by SHA-256 hashes.
These hashes are verified during snapshot restore operations. The structure of
the archive is internal to Consul and not intended to be used other than for
restore operations. The archives are not designed to be modified before a
restore.

| Method | Path                         | Produces                   |
| :----- | :--------------------------- | -------------------------- |
| `GET`  | `/snapshot`                  | `200 application/x-gzip`   |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `default,stale`   | `management` |

### Parameters

- `dc` `(string: "")` - Specifies the datacenter to query. This will default
  to the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

- `stale` `(bool: false)` - Specifies that any follower may reply. By default
  requests are forwarded to the leader. Followers may be faster to respond, but
  may have stale data. To support bounding the acceptable staleness of
  snapshots, responses provide the `X-Consul-LastContact` header containing the
  time in milliseconds that a server was last contacted by the leader node. The
  `X-Consul-KnownLeader` header also indicates if there is a known leader. These
  can be used by clients to gauge the staleness of a snapshot and take
  appropriate action. The stale mode is particularly useful for taking a
  snapshot of a cluster in a failed state with no current leader.

### Sample Request

With a custom datacenter:

```text
$ curl https://consul.rocks/v1/snapshot?dc=my-datacenter
```

### Sample Response

```text
<gzipped tarball ...>
```

In addition to the Consul standard stale-related headers, the `X-Consul-Index`
header will contain the index at which the snapshot took place.

## Restore Snapshot

This endpoint restores a point-in-time snapshot of the Consul server state.

Restores involve a potentially dangerous low-level Raft operation that is not
designed to handle server failures during a restore. This operation is primarily
intended to be used when recovering from a disaster, restoring into a fresh
cluster of Consul servers.

The body of the request should be a snapshot archive returned from a previous
call to the `GET` method.

| Method | Path                         | Produces                      |
| :----- | :--------------------------- | ----------------------------- |
| `PUT`  | `/snapshot`                  | `200 text/plain (empty body)` |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `none`            | `management` |
### Parameters

- `dc` `(string: "")` - Specifies the datacenter to query. This will default
  to the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

### Sample Request

```text
$ curl \
    --request PUT \
    --data-binary @snapshot \
    https://consul.rocks/v1/snapshot
```

~> Some tools default to www/encoded uploads. Consul expects the snapshot to be
in pure binary form.
