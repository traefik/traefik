---
layout: api
page_title: Keyring - Operator - HTTP API
sidebar_current: api-operator-keyring
description: |-
  The /operator/keyring endpoints allow for management of the gossip encryption
  keyring.
---

# Keyring Operator HTTP API

The `/operator/keyring` endpoints allow for management of the gossip encryption
keyring. Please see the [Gossip Protocol Guide](/docs/internals/gossip.html) for
more details on the gossip protocol and its use.

## List Gossip Encryption Keys

This endpoint lists the gossip encryption keys installed on both the WAN and LAN
rings of every known datacenter.

If ACLs are enabled, the client will need to supply an ACL Token with `keyring`
read privileges.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/operator/keyring`          | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required   |
| ---------------- | ----------------- | -------------- |
| `NO`             | `none`            | `keyring:read` |

### Parameters

- `relay-factor` `(int: 0)` - Specifies the relay factor. Setting this to a
  non-zero value will cause nodes to relay their responses through this many
  randomly-chosen other nodes in the cluster. The maximum allowed value is `5`.
  This is specified as part of the URL as a query parameter.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/operator/keyring
```

### Sample Response

```json
[
  {
    "WAN": true,
    "Datacenter": "dc1",
    "Keys": {
      "0eK8RjnsGC/+I1fJErQsBA==": 1,
      "G/3/L4yOw3e5T7NTvuRi9g==": 1,
      "z90lFx3sZZLtTOkutXcwYg==": 1
    },
    "NumNodes": 1
  },
  {
    "WAN": false,
    "Datacenter": "dc1",
    "Keys": {
      "0eK8RjnsGC/+I1fJErQsBA==": 1,
      "G/3/L4yOw3e5T7NTvuRi9g==": 1,
      "z90lFx3sZZLtTOkutXcwYg==": 1
    },
    "NumNodes": 1
  }
]
```

- `WAN` is true if the block refers to the WAN ring of that datacenter (rather
   than LAN).

- `Datacenter` is the datacenter the block refers to.

- `Keys` is a map of each gossip key to the number of nodes it's currently
  installed on.

- `NumNodes` is the total number of nodes in the datacenter.

## Add New Gossip Encryption Key

This endpoint installs a new gossip encryption key into the cluster.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `POST` | `/operator/keyring`          | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required    |
| ---------------- | ----------------- | --------------- |
| `NO`             | `none`            | `keyring:write` |

### Parameters

- `relay-factor` `(int: 0)` - Specifies the relay factor. Setting this to a
  non-zero value will cause nodes to relay their responses through this many
  randomly-chosen other nodes in the cluster. The maximum allowed value is `5`.
  This is specified as part of the URL as a query parameter.

- `Key` `(string: <required>)` - Specifies the encryption key to install into
  the cluster.

### Sample Payload

```json
{
  "Key": "3lg9DxVfKNzI8O+IQ5Ek+Q=="
}
```

### Sample Request

```text
$ curl \
    --request POST \
    --data @payload.json \
    https://consul.rocks/v1/operator/keyring
```

## Change Primary Gossip Encryption Key

This endpoint changes the primary gossip encryption key. The key must already be
installed before this operation can succeed.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/operator/keyring`          | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required    |
| ---------------- | ----------------- | --------------- |
| `NO`             | `none`            | `keyring:write` |

### Parameters

- `relay-factor` `(int: 0)` - Specifies the relay factor. Setting this to a
  non-zero value will cause nodes to relay their responses through this many
  randomly-chosen other nodes in the cluster. The maximum allowed value is `5`.
  This is specified as part of the URL as a query parameter.

- `Key` `(string: <required>)` - Specifies the encryption key to begin using as
  primary into the cluster.

### Sample Payload

```json
{
 "Key": "3lg9DxVfKNzI8O+IQ5Ek+Q=="
}
```

### Sample Request

```text
$ curl \
    --request PUT \
    --data @payload.json \
    https://consul.rocks/v1/operator/keyring
```

## Delete Gossip Encryption Key

This endpoint removes a gossip encryption key from the cluster. This operation
may only be performed on keys which are not currently the primary key.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/operator/keyring`          | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required    |
| ---------------- | ----------------- | --------------- |
| `NO`             | `none`            | `keyring:write` |

### Parameters

- `relay-factor` `(int: 0)` - Specifies the relay factor. Setting this to a
  non-zero value will cause nodes to relay their responses through this many
  randomly-chosen other nodes in the cluster. The maximum allowed value is `5`.
  This is specified as part of the URL as a query parameter.

- `Key` `(string: <required>)` - Specifies the encryption key to delete.

### Sample Payload

```json
{
 "Key": "3lg9DxVfKNzI8O+IQ5Ek+Q=="
}
```

### Sample Request

```text
$ curl \
    --request DELETE \
    --data @payload.json \
    https://consul.rocks/v1/operator/keyring
```
