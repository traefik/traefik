---
layout: api
page_title: Catalog - HTTP API
sidebar_current: api-catalog
description: |-
  The /catalog endpoints register and deregister nodes, services, and checks in
  Consul.
---

# Catalog HTTP API

The `/catalog` endpoints register and deregister nodes, services, and checks in
Consul. The catalog should not be confused with the agent, since some of the
API methods look similar.

## Register Entity

This endpoint is a low-level mechanism for registering or updating
entries in the catalog. It is usually preferable to instead use the
[agent endpoints](/api/agent.html) for registration as they are simpler and
perform [anti-entropy](/docs/internals/anti-entropy.html).

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/catalog/register`          | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required               |
| ---------------- | ----------------- | -------------------------- |
| `NO`             | `none`            | `node:write,service:write` |

### Parameters

- `ID` `(string: "")` - An optional UUID to assign to the service. If not
  provided, one is generated. This must be a 36-character UUID.

- `Node` `(string: <required>)` - Specifies the node ID to register.

- `Address` `(string: <required>)` - Specifies the address to register.

- `Datacenter` `(string: "")` - Specifies the datacenter, which defaults to the
  agent's datacenter if not provided.

- `TaggedAddresses` `(map<string|string>: nil)` - Specifies the tagged
  addresses.

- `Meta` `(map<string|string>: nil)` - Specifies arbitrary KV metadata
  pairs for filtering purposes.

- `Service` `(Service: nil)` - Specifies to register a service. If `ID` is not
  provided, it will be defaulted to the value of the `Service.Service` property.
  Only one service with a given `ID` may be present per node. The service
  `Tags`, `Address`, and `Port` fields are all optional.

- `Check` `(Check: nil)` - Specifies to register a check. The register API
  manipulates the health check entry in the Catalog, but it does not setup the
  script, TTL, or HTTP check to monitor the node's health. To truly enable a new
  health check, the check must either be provided in agent configuration or set
  via the [agent endpoint](agent.html).

    The `CheckID` can be omitted and will default to the value of `Name`. As
    with `Service.ID`, the `CheckID` must be unique on this node. `Notes` is an
    opaque field that is meant to hold human-readable text. If a `ServiceID` is
    provided that matches the `ID` of a service on that node, the check is
    treated as a service level health check, instead of a node level health
    check. The `Status` must be one of `passing`, `warning`, or `critical`.

    Multiple checks can be provided by replacing `Check` with `Checks` and
    sending an array of `Check` objects.

It is important to note that `Check` does not have to be provided with `Service`
and vice versa. A catalog entry can have either, neither, or both.

### Sample Payload

```json
{
  "Datacenter": "dc1",
  "ID": "40e4a748-2192-161a-0510-9bf59fe950b5",
  "Node": "foobar",
  "Address": "192.168.10.10",
  "TaggedAddresses": {
    "lan": "192.168.10.10",
    "wan": "10.0.10.10"
  },
  "NodeMeta": {
    "somekey": "somevalue"
  },
  "Service": {
    "ID": "redis1",
    "Service": "redis",
    "Tags": [
      "primary",
      "v1"
    ],
    "Address": "127.0.0.1",
    "Port": 8000
  },
  "Check": {
    "Node": "foobar",
    "CheckID": "service:redis1",
    "Name": "Redis health check",
    "Notes": "Script based health check",
    "Status": "passing",
    "ServiceID": "redis1"
  }
}
```

### Sample Request

```text
$ curl \
    --request PUT \
    --data @payload.json \
    https://consul.rocks/v1/catalog/register
```

## Deregister Entity

This endpoint is a low-level mechanism for directly removing
entries from the Catalog. It is usually preferable to instead use the
[agent endpoints](/api/agent.html) for deregistration as they are simpler and
perform [anti-entropy](/docs/internals/anti-entropy.html).

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/catalog/deregister`        | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required               |
| ---------------- | ----------------- | -------------------------- |
| `NO`             | `none`            | `node:write,service:write` |

### Parameters

The behavior of the endpoint depends on what keys are provided.

- `Node` `(string: <required>)` - Specifies the ID of the node. If no other
  values are provided, this node, all its services, and all its checks are
  removed.

- `Datacenter` `(string: "")` - Specifies the datacenter, which defaults to the
  agent's datacenter if not provided.

- `CheckID` `(string: "")` - Specifies the ID of the check to remove.

- `ServiceID` `(string: "")` - Specifies the ID of the service to remove. The
  service and all associated checks will be removed.

### Sample Payloads

```json
{
  "Datacenter": "dc1",
  "Node": "foobar"
}
```

```json
{
  "Datacenter": "dc1",
  "Node": "foobar",
  "CheckID": "service:redis1"
}
```

```json
{
  "Datacenter": "dc1",
  "Node": "foobar",
  "ServiceID": "redis1"
}
```

### Sample Request

```text
$ curl \
    --request PUT \
    --data @payload.json \
    https://consul.rocks/v1/catalog/deregister
```

## List Datacenters

This endpoint returns the list of all known datacenters. The datacenters will be
sorted in ascending order based on the estimated median round trip time from the
server to the servers in that datacenter.

This endpoint does not require a cluster leader and will succeed even during an
availability outage. Therefore, it can be used as a simple check to see if any
Consul servers are routable.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/catalog/datacenters`       | `application/json`         |

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
    https://consul.rocks/v1/catalog/datacenters
```

### Sample Respons

```json
["dc1", "dc2"]
```

## List Nodes

This endpoint and returns the nodes registered in a given datacenter.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/catalog/nodes`             | `application/json`         |

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
    https://consul.rocks/v1/catalog/nodes
```

### Sample Response

```json
[
  {
    "ID": "40e4a748-2192-161a-0510-9bf59fe950b5",
    "Node": "baz",
    "Address": "10.1.10.11",
    "TaggedAddresses": {
      "lan": "10.1.10.11",
      "wan": "10.1.10.11"
    },
    "Meta": {
      "instance_type": "t2.medium"
    }
  },
  {
    "ID": "8f246b77-f3e1-ff88-5b48-8ec93abf3e05",
    "Node": "foobar",
    "Address": "10.1.10.12",
    "TaggedAddresses": {
      "lan": "10.1.10.11",
      "wan": "10.1.10.12"
    },
    "Meta": {
      "instance_type": "t2.large"
    }
  }
]
```

## List Services

This endpoint returns the services registered in a given datacenter.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/catalog/services`          | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required   |
| ---------------- | ----------------- | -------------- |
| `TEST`           | `all`             | `service:read` |

### Parameters

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

- `node-meta` `(string: "")` - Specifies a desired node metadata key/value pair
  of the form `key:value`. This parameter can be specified multiple times, and
  will filter the results to nodes with the specified key/value pairs. This is
  specified as part of the URL as a query parameter.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/catalog/services
```

### Sample Response

```json
{
  "consul": [],
  "redis": [],
  "postgresql": [
    "primary",
    "secondary"
  ]
}
```

The keys are the service names, and the array values provide all known tags for
a given service.

## List Nodes for Service

This endpoint returns the nodes providing a service in a given datacenter.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/catalog/service/:service`  | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required             |
| ---------------- | ----------------- | ------------------------ |
| `YES`            | `all`             | `node:read,service:read` |

### Parameters

- `service` `(string: <required>)` - Specifies the name of the service for which
  to list nodes. This is specified as part of the URL.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

- `tag` `(string: "")` - Specifies tags to filter on.

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
    https://consul.rocks/v1/catalog/service/my-service
```

### Sample Response

```json
[
  {
    "ID": "40e4a748-2192-161a-0510-9bf59fe950b5",
    "Node": "foobar",
    "Address": "192.168.10.10",
    "TaggedAddresses": {
      "lan": "192.168.10.10",
      "wan": "10.0.10.10"
    },
    "Meta": {
      "instance_type": "t2.medium"
    },
    "CreateIndex": 51,
    "ModifyIndex": 51,
    "ServiceAddress": "172.17.0.3",
    "ServiceEnableTagOverride": false,
    "ServiceID": "32a2a47f7992:nodea:5000",
    "ServiceName": "foobar",
    "ServicePort": 5000,
    "ServiceTags": [
      "tacos"
    ]
  }
]
```

- `Address` is the IP address of the Consul node on which the service is
  registered.

- `TaggedAddresses` is the list of explicit LAN and WAN IP addresses for the
  agent

- `Meta` is a list of user-defined metadata key/value pairs for the node

- `CreateIndex` is an internal index value representing when the service was
  created

- `ModifyIndex` is the last index that modified the service

- `Node` is the name of the Consul node on which the service is registered

- `ServiceAddress` is the IP address of the service host — if empty, node
  address should be used

- `ServiceEnableTagOverride` indicates whether service tags can be overridden on
  this service

- `ServiceID` is a unique service instance identifier

- `ServiceName` is the name of the service

- `ServicePort` is the port number of the service

- `ServiceTags` is a list of tags for the service

## List Services for Node

This endpoint returns the node's registered services.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/catalog/node/:node`        | `application/json`         |

The table below shows this endpoint's support for blocking queries and
consistency modes.

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required             |
| ---------------- | ----------------- | ------------------------ |
| `YES`            | `all`             | `node:read,service:read` |

### Parameters

- `node` `(string: <required>)` - Specifies the name of the node for which
  to list services. This is specified as part of the URL.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/catalog/node/my-node
```

### Sample Response

```json
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
  "Services": {
    "consul": {
      "ID": "consul",
      "Service": "consul",
      "Tags": null,
      "Port": 8300
    },
    "redis": {
      "ID": "redis",
      "Service": "redis",
      "Tags": [
        "v1"
      ],
      "Port": 8000
    }
  }
}
```
