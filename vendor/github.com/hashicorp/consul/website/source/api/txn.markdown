---
layout: api
page_title: Transaction - HTTP API
sidebar_current: api-txn
description: |-
  The /txn endpoints manage updates or fetches of multiple keys inside a single,
  atomic transaction.
---

# Transactions HTTP API

The `/txn` endpoints manage updates or fetches of multiple keys inside a single,
atomic transaction. It is important to note that each datacenter has its own KV
store, and there is no built-in replication between datacenters.

## Create Transaction

This endpoint permits submitting a list of operations to apply to the KV store
inside of a transaction. If any operation fails, the transaction is rolled back
and none of the changes are applied.

If the transaction does not contain any write operations then it will be
fast-pathed internally to an endpoint that works like other reads, except that
blocking queries are not currently supported. In this mode, you may supply the
`?stale` or `?consistent` query parameters with the request to control
consistency. To support bounding the acceptable staleness of data, read-only
transaction responses provide the `X-Consul-LastContact` header containing the
time in milliseconds that a server was last contacted by the leader node. The
`X-Consul-KnownLeader` header also indicates if there is a known leader. These
won't be present if the transaction contains any write operations, and any
consistency query parameters will be ignored, since writes are always managed by
the leader via the Raft consensus protocol.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/txn`                       | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `all`<sup>1</sup> | `key:read,key:write`<sup>2</sup> |

<sup>1</sup> For read-only transactions
<br>
<sup>2</sup> The ACL required depends on the operations in the transaction.

### Parameters

- `dc` `(string: "")` - Specifies the datacenter to query. This will default
  to the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

- `KV` is the only available operation type, though other types may be added in the future.

  - `Verb` `(string: <required>)` - Specifies the type of operation to perform.
    Please see the table below for available verbs.

  - `Key` `(string: <required>)` - Specifies the full path of the entry.

  - `Value` `(string: "")` - Specifies a **base64-encoded** blob of data. Values
    cannot be larger than 512kB.

  - `Flags` `(int: 0)` - Specifies an opaque unsigned integer that can be
    attached to each entry. Clients can choose to use this however makes sense
    for their application.

  - `Index` `(int: 0)` - Specifies an index. See the table below for more
    information.

  - `Session` `(string: "")` - Specifies a session. See the table below for more
    information.

### Sample Payload

The body of the request should be a list of operations to perform inside the
atomic transaction. Up to 64 operations may be present in a single transaction.

```javascript
[
  {
    "KV": {
      "Verb": "<verb>",
      "Key": "<key>",
      "Value": "<Base64-encoded blob of data>",
      "Flags": <flags>,
      "Index": <index>,
      "Session": "<session id>"
    }
  }
]
```

### Sample Request

```text
$ curl \
    --request PUT \
    --data @payload.json \
    https://consul.rocks/v1/txn
```

### Sample Response

If the transaction can be processed, a status code of 200 will be returned if it
was successfully applied, or a status code of 409 will be returned if it was
rolled back. If either of these status codes are returned, the response will
look like this:

```javascript
{
  "Results": [
    {
      "KV": {
        "LockIndex": <lock index>,
        "Key": "<key>",
        "Flags": <flags>,
        "Value": "<Base64-encoded blob of data, or null>",
        "CreateIndex": <index>,
        "ModifyIndex": <index>
      }
    }
  ],
  "Errors": [
    {
      "OpIndex": <index of failed operation>,
      "What": "<error message for failed operation>"
    },
  ]
}
```

- `Results` has entries for some operations if the transaction was successful.
  To save space, the `Value` will be `null` for any `Verb` other than "get" or
  "get-tree". Like the `/v1/kv/<key>` endpoint, `Value` will be Base64-encoded
  if it is present. Also, no result entries  will be added for verbs that delete
  keys.

- `Errors` has entries describing which operations failed if the transaction was
  rolled back. The `OpIndex` gives the index of the failed operation in the
  transaction, and `What` is a string with an error message about why that
  operation failed.

### Table of Operations

The following table summarizes the available verbs and the fields that apply to
that operation ("X" means a field is required and "O" means it is optional):

| Verb            | Operation                                    | Key  | Value | Flags | Index | Session |
| --------------- | -------------------------------------------- | :--: | :---: | :---: | :---: | :-----: |
| `set`           | Sets the `Key` to the given `Value`          | `x`  | `x`   | `o`   |       |         |  
| `cas`           | Sets, but with CAS semantics                 | `x`  | `x`   | `o`   | `x`   |         |  
| `lock`          | Lock with the given `Session`                | `x`  | `x`   | `o`   |       | `x`     |  
| `unlock`        | Unlock with the given `Session`              | `x`  | `x`   | `o`   |       | `x`     |  
| `get`           | Get the key, fails if it does not exist      | `x`  |       |       |       |         |  
| `get-tree`      | Gets all keys with the prefix                | `x`  |       |       |       |         |  
| `check-index`   | Fail if modify index != index                | `x`  |       |       | `x`   |         |  
| `check-session` | Fail if not locked by session                | `x`  |       |       |       | `x`     |  
| `delete`        | Delete the key                               | `x`  |       |       |       |         |  
| `delete-tree`   | Delete all keys with a prefix                | `x`  |       |       |       |         |  
| `delete-cas`    | Delete, but with CAS semantics               | `x`  |       |       | `x`   |         |  
