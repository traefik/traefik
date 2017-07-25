---
layout: api
page_title: Health - HTTP API
sidebar_current: api-health
description: |-
  The /health endpoints query health-related information for services and checks
  in Consul.
---

# Health HTTP Endpoint

The `/health` endpoints query health-related information. They are provided
separately from the `/catalog` endpoints since users may prefer not to use the
optional health checking mechanisms. Additionally, some of the query results
from the health endpoints are filtered while the catalog endpoints provide the
raw entries.

## List Checks for Node

This endpoint returns the checks specific to the node provided on the path.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/health/node/:node`         | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required             |
| ---------------- | ----------------- | ------------------------ |
| `YES`            | `all`             | `node:read,service:read` |

### Parameters

- `node` `(string: <required>)` - Specifies the name or ID of the node to query.
  This is specified as part of the URL

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/health/node/my-node
```

### Sample Response

```json
[
  {
    "ID": "40e4a748-2192-161a-0510-9bf59fe950b5",
    "Node": "foobar",
    "CheckID": "serfHealth",
    "Name": "Serf Health Status",
    "Status": "passing",
    "Notes": "",
    "Output": "",
    "ServiceID": "",
    "ServiceName": ""
  },
  {
    "ID": "40e4a748-2192-161a-0510-9bf59fe950b5",
    "Node": "foobar",
    "CheckID": "service:redis",
    "Name": "Service 'redis' check",
    "Status": "passing",
    "Notes": "",
    "Output": "",
    "ServiceID": "redis",
    "ServiceName": "redis"
  }
]
```

## List Checks for Service

This endpoint returns the checks associated with the service provided on the
path.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/health/checks/:service`    | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required             |
| ---------------- | ----------------- | ------------------------ |
| `YES`            | `all`             | `node:read,service:read` |

### Parameters

- `service` `(string: <required>)` - Specifies the service to list checks for.
  This is provided as part of the URL.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

- `near` `(string: "")` - Specifies a node name to sort the node list in
  ascending order based on the estimated round trip time from that node. Passing
  `?near=_agent` will use the agent's node for the sort. This is specified as
  part of the URL as a query parameter.

- `node-meta` `(string: "")` - Specifies a desired node metadata key/value pair
  of the form `key:value`. This parameter can be specified multiple times, and
  will filter the results to nodes with the specified key/value pairs. This is
  specified as part of the URL as a query parameter.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/health/checks/my-service
```

### Sample Response

```json
[
  {
    "Node": "foobar",
    "CheckID": "service:redis",
    "Name": "Service 'redis' check",
    "Status": "passing",
    "Notes": "",
    "Output": "",
    "ServiceID": "redis",
    "ServiceName": "redis"
  }
]
```

## List Nodes for Service

This endpoint returns the nodes providing the service indicated on the path.
Users can also build in support for dynamic load balancing and other features by
incorporating the use of health checks.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/health/service/:service`   | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required             |
| ---------------- | ----------------- | ------------------------ |
| `YES`            | `all`             | `node:read,service:read` |

### Parameters

- `service` `(string: <required>)` - Specifies the service to list services for.
  This is provided as part of the URL.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

- `near` `(string: "")` - Specifies a node name to sort the node list in
  ascending order based on the estimated round trip time from that node. Passing
  `?near=_agent` will use the agent's node for the sort. This is specified as
  part of the URL as a query parameter.

- `tag` `(string: "")` - Specifies the list of tags to filter the list. This is
  specifies as part of the URL as a query parameter.

- `node-meta` `(string: "")` - Specifies a desired node metadata key/value pair
  of the form `key:value`. This parameter can be specified multiple times, and
  will filter the results to nodes with the specified key/value pairs. This is
  specified as part of the URL as a query parameter.

- `passing` `(bool: false)` - Specifies that the server should return only nodes
  with all checks in the `passing` state. This can be used to avoid additional
  filtering on the client side.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/health/service/my-service
```

### Sample Response

```json
[
  {
    "Node": {
      "ID": "40e4a748-2192-161a-0510-9bf59fe950b5",
      "Node": "foobar",
      "Address": "10.1.10.12",
      "TaggedAddresses": {
        "lan": "10.1.10.12",
        "wan": "10.1.10.12"
      },
      "Meta": {
        "instance_type": "t2.medium"
      }
    },
    "Service": {
      "ID": "redis",
      "Service": "redis",
      "Tags": null,
      "Address": "10.1.10.12",
      "Port": 8000
    },
    "Checks": [
      {
        "Node": "foobar",
        "CheckID": "service:redis",
        "Name": "Service 'redis' check",
        "Status": "passing",
        "Notes": "",
        "Output": "",
        "ServiceID": "redis",
        "ServiceName": "redis"
      },
      {
        "Node": "foobar",
        "CheckID": "serfHealth",
        "Name": "Serf Health Status",
        "Status": "passing",
        "Notes": "",
        "Output": "",
        "ServiceID": "",
        "ServiceName": ""
      }
    ]
  }
]
```

## List Checks in State

This endpoint returns the checks in the state provided on the path.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/health/state/:state`       | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required             |
| ---------------- | ----------------- | ------------------------ |
| `YES`            | `all`             | `node:read,service:read` |

### Parameters

- `state` `(string: <required>)` - Specifies the state to query. Spported states
  are `any`, `passing`, `warning`, or `critical`. The `any` state is a wildcard
  that can be used to return all checks.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

- `near` `(string: "")` - Specifies a node name to sort the node list in
  ascending order based on the estimated round trip time from that node. Passing
  `?near=_agent` will use the agent's node for the sort. This is specified as
  part of the URL as a query parameter.

- `node-meta` `(string: "")` - Specifies a desired node metadata key/value pair
  of the form `key:value`. This parameter can be specified multiple times, and
  will filter the results to nodes with the specified key/value pairs. This is
  specified as part of the URL as a query parameter.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/health/state/passing
```

### Sample Response

```json
[
  {
    "Node": "foobar",
    "CheckID": "serfHealth",
    "Name": "Serf Health Status",
    "Status": "passing",
    "Notes": "",
    "Output": "",
    "ServiceID": "",
    "ServiceName": ""
  },
  {
    "Node": "foobar",
    "CheckID": "service:redis",
    "Name": "Service 'redis' check",
    "Status": "passing",
    "Notes": "",
    "Output": "",
    "ServiceID": "redis",
    "ServiceName": "redis"
  }
]
```
