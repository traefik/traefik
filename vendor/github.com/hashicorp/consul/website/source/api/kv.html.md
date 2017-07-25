---
layout: api
page_title: KV Store - HTTP API
sidebar_current: api-kv-store
description: |-
  The /kv endpoints access Consul's simple key/value store, useful for storing
  service configuration or other metadata.
---

# KV Store Endpoints

The `/kv` endpoints access Consul's simple key/value store, useful for storing
service configuration or other metadata.

It is important to note that each datacenter has its own KV store, and there is
no built-in replication between datacenters. If you are interested in
replication between datacenters, please view the
[Consul Replicate](https://github.com/hashicorp/consul-replicate) project.

~> Values in the KV store cannot be larger than 512kb.

For multi-key updates, please consider using [transaction](/api/txn.html).

## Read Key

This endpoint returns the specified key. If no key exists at the given path, a
404 is returned instead of a 200 response.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/kv/:key`                   | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `YES`            | `all`             | `key:read`   |

### Parameters

- `key` `(string: "")` - Specifies the path of the key to read.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

- `recurse` `(bool: false)` - Specifies if the lookup should be recursive and
  `key` treated as a prefix instead of a literal match. This is specified as
  part of the URL as a query parameter.

- `raw` `(bool: false)` - Specifies the response is just the raw value of the
  key, without any encoding or metadata. This is specified as part of the URL as
  a query parameter.

- `keys` `(bool: false)` - Specifies to return only keys (no values or
  metadata). Specifying this implies `recurse`. This is specified as part of the
  URL as a query parameter.

- `separator` `(string: '/')` - Specifies the character to use as a separator
  for recursive lookups. This is specified as part of the URL as a query
  parameter.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/kv/my-key
```

### Sample Response

#### Metadata Response

```json
[
  {
    "CreateIndex": 100,
    "ModifyIndex": 200,
    "LockIndex": 200,
    "Key": "zip",
    "Flags": 0,
    "Value": "dGVzdA==",
    "Session": "adf4238a-882b-9ddc-4a9d-5b6758e4159e"
  }
]
```

- `CreateIndex` is the internal index value that represents when the entry was
  created.

- `ModifyIndex` is the last index that modified this key. This index corresponds
  to the `X-Consul-Index` header value that is returned in responses, and it can
  be used to establish blocking queries by setting the `?index` query parameter.
  You can even perform blocking queries against entire subtrees of the KV store:
  if `?recurse` is provided, the returned `X-Consul-Index` corresponds to the
  latest `ModifyIndex` within the prefix, and a blocking query using that
  `?index` will wait until any key within that prefix is updated.

- `LockIndex` is the number of times this key has successfully been acquired in
  a lock. If the lock is held, the `Session` key provides the session that owns
  the lock.

- `Key` is simply the full path of the entry.

- `Flags` is an opaque unsigned integer that can be attached to each entry.
  Clients can choose to use this however makes sense for their application.

- `Value` is a base64-encoded blob of data.

#### Keys Response

When using the `?keys` query parameter, the response structure changes to an
array of strings instead of an array of JSON objects. Listing `/web/` with a `/`
separator may return:

```json
[
  "/web/bar",
  "/web/foo",
  "/web/subdir/"
]
```

Using the key listing method may be suitable when you do not need the values or
flags or want to implement a key-space explorer.

#### Raw Response

When using the `?raw` endpoint, the response is not `application/json`, but
rather the content type of the uploaded content.

```
)k������z^�-�ɑj�q����#u�-R�r��T�D��٬�Y��l,�ιK��Fm��}�#e��
```

(Yes, that is intentionally a bunch of gibberish characters to showcase the
response)

## Create/Update Key

This endpoint

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/kv/:key`                   | `application/json`         |

Even though the return type is `application/json`, the value is either `true` or
`false`, indicating whether the create/update succeeded.

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `none`            | `key:write`  |

### Parameters

- `key` `(string: "")` - Specifies the path of the key to read.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

- `flags` `(int: 0)` - Specifies an unsigned value between `0` and `(2^64)-1`.
  Clients can choose to use this however makes sense for their application. This
  is specified as part of the URL as a query parameter.

- `cas` `(int: 0)` - Specifies to use a Check-And-Set operation. This is very
  useful as a building block for more complex synchronization primitives. If the
  index is 0, Consul will only put the key if it does not already exist. If the
  index is non-zero, the key is only set if the index matches the `ModifyIndex`
  of that key.

- `acquire` `(string: "")` - Specifies to use a lock acquisition operation. This
  is useful as it allows leader election to be built on top of Consul. If the
  lock is not held and the session is valid, this increments the `LockIndex` and
  sets the `Session` value of the key in addition to updating the key contents.
  A key does not need to exist to be acquired. If the lock is already held by
  the given session, then the `LockIndex` is not incremented but the key
  contents are updated. This lets the current lock holder update the key
  contents without having to give up the lock and reacquire it.

- `release` `(string: "")` - Specifies to use a lock release operation. This is
  useful when paired with `?acquire=` as it allows clients to yield a lock. This
  will leave the `LockIndex` unmodified but will clear the associated `Session`
  of the key. The key must be held by this session to be unlocked.

### Sample Payload

The payload is arbitrary, and is loaded directly into Consul as supplied.

### Sample Request

```text
$ curl \
    --request PUT \
    --data @contents \
    https://consul.rocks/v1/kv/my-key
```

### Sample Response

```json
true
```

## Delete Key

This endpoint deletes a single key or all keys sharing a prefix.

| Method   | Path                         | Produces                   |
| -------- | ---------------------------- | -------------------------- |
| `DELETE` | `/kv/:key`                   | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `none`            | `key:write`  |

### Parameters

- `recurse` `(bool: false)` - Specifies to delete all keys which have the
  specified prefix. Without this, only a key with an exact match will be
  deleted.

- `cas` `(int: 0)` - Specifies to use a Check-And-Set operation. This is very
  useful as a building block for more complex synchronization primitives. Unlike
  `PUT`, the index must be greater than 0 for Consul to take any action: a 0
  index will not delete the key. If the index is non-zero, the key is only
  deleted if the index matches the `ModifyIndex` of that key.

### Sample Request

```text
$ curl \
    --request DELETE \
    https://consul.rocks/v1/kv/my-key
```

### Sample Response

```json
true
```
