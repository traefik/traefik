---
layout: api
page_title: Raft - Operator - HTTP API
sidebar_current: api-operator-raft
description: |-
  The /operator/raft endpoints provide tools for management of Raft the
  consensus subsystem and cluster quorum.
---

# Raft Operator HTTP API

The `/operator/raft` endpoints provide tools for management of Raft the
consensus subsystem and cluster quorum.

Please see the [Consensus Protocol Guide](/docs/internals/consensus.html) for
more information about Raft consensus protocol and its use.

## Read Configuration

This endpoint reads the current raft configuration.

| Method | Path                           | Produces                   |
| ------ | ------------------------------ | -------------------------- |
| `GET`  | `/operator/raft/configuration` | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required    |
| ---------------- | ----------------- | --------------- |
| `NO`             | `none`            | `operator:read` |

### Parameters

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query string.

- `stale` `(bool: false)` - If the cluster does not currently have a leader an
  error will be returned. You can use the `?stale` query parameter to read the
  Raft configuration from any of the Consul servers.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/operator/raft/configuration
```

### Sample Response

```json
{
  "Servers": [
    {
      "ID": "127.0.0.1:8300",
      "Node": "alice",
      "Address": "127.0.0.1:8300",
      "Leader": true,
      "Voter": true
    },
    {
      "ID": "127.0.0.2:8300",
      "Node": "bob",
      "Address": "127.0.0.2:8300",
      "Leader": false,
      "Voter": true
    },
    {
      "ID": "127.0.0.3:8300",
      "Node": "carol",
      "Address": "127.0.0.3:8300",
      "Leader": false,
      "Voter": true
    }
  ],
  "Index": 22
}
```

- `Servers` is has information about the servers in the Raft peer configuration:

  - `ID` is the ID of the server. This is the same as the `Address` in Consul
    0.7 but may be upgraded to a GUID in a future version of Consul.

  - `Node` is the node name of the server, as known to Consul, or "(unknown)" if
    the node is stale and not known.

  - `Address` is the IP:port for the server.

  - `Leader` is either "true" or "false" depending on the server's role in the
    Raft configuration.

  - `Voter` is "true" or "false", indicating if the server has a vote in the
    Raft configuration. Future versions of Consul may add support for non-voting
    servers.

- `Index` is the Raft corresponding to this configuration. The latest
  configuration may not yet be committed if changes are in flight.

## Delete Raft Peer

This endpoint removes the Consul server with given address from the Raft
configuration.

There are rare cases where a peer may be left behind in the Raft configuration
even though the server is no longer present and known to the cluster. This
endpoint can be used to remove the failed server so that it is no longer affects
the Raft quorum.

If ACLs are enabled, the client will need to supply an ACL Token with `operator`
write privileges.

| Method   | Path                         | Produces                   |
| -------- | ---------------------------- | -------------------------- |
| `DELETE` | `/operator/raft/peer`        | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required     |
| ---------------- | ----------------- | ---------------- |
| `NO`             | `none`            | `operator:write` |

### Parameters

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query string.

- `id|address` `(string: <required>)` - Specifies the ID or address (IP:port) of the raft peer to remove.

### Sample Request

```text
$ curl \
    --request DELETE \
    https://consul.rocks/v1/operator/raft/peer?address=1.2.3.4:5678
```
