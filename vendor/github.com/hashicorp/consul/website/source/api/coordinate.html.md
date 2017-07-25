---
layout: api
page_title: Coordinate - HTTP API
sidebar_current: api-coordinate
description: |-
  The /coordinate endpoints query for the network coordinates for nodes in the
  local datacenter as well as Consul servers in the local datacenter and remote
  datacenters.
---

# Coordinate HTTP Endpoint

The `/coordinate` endpoints query for the network coordinates for nodes in the
local datacenter as well as Consul servers in the local datacenter and remote
datacenters.

Please see the [Network Coordinates](/docs/internals/coordinates.html) internals
guide for more information on how these coordinates are computed, and for
details on how to perform calculations with them.

## Read WAN Coordinates

This endpoint returns the WAN network coordinates for all Consul servers,
organized by datacenters. It serves data out of the server's local Serf data, so
its results may vary as requests are handled by different servers in the
cluster.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/coordinate/datacenters`    | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `none`            | `none`       |

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/coordinate/datacenters
```

### Sample Response

```json
[
  {
    "Datacenter": "dc1",
    "AreaID": "WAN",
    "Coordinates": [
      {
        "Node": "agent-one",
        "Coord": {
          "Adjustment": 0,
          "Error": 1.5,
          "Height": 0,
          "Vec": [0, 0, 0, 0, 0, 0, 0, 0]
        }
      }
    ]
  }
]
```

In **Consul Enterprise**, this will include coordinates for user-added network
areas as well, as indicated by the `AreaID`. Coordinates are only compatible
within the same area.

## Read LAN Coordinates

This endpoint returns the LAN network coordinates for all nodes in a given
datacenter.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/coordinate/nodes`          | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `YES`            | `all`             | `node:read`  |

### Parameters

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/coordinate/nodes
```

### Sample Response

```json
[
  {
    "Node": "agent-one",
    "Coord": {
      "Adjustment": 0,
      "Error": 1.5,
      "Height": 0,
      "Vec": [0, 0, 0, 0, 0, 0, 0, 0]
    }
  }
]
```
