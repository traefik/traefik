---
layout: api
page_title: Agent - HTTP API
sidebar_current: api-agent
description: |-
  The /agent endpoints interact with the local Consul agent to register
  services, checks, list members, and more.
---

# Agent HTTP API

The `/agent` endpoints are used to interact with the local Consul agent.
Usually, services and checks are registered with an agent which then takes on
the burden of keeping that data synchronized with the cluster. For example, the
agent registers services and checks with the Catalog and performs
[anti-entropy](/docs/internals/anti-entropy.html) to recover from outages.

In addition to these endpoints, additional endpoints are grouped in the
navigation for `Checks` and `Services`.

## List Members

This endpoint returns the members the agent sees in the cluster gossip pool. Due
to the nature of gossip, this is eventually consistent: the results may differ
by agent. The strongly consistent view of nodes is instead provided by
`/v1/catalog/nodes`.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/agent/members`             | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `none`            | `node:read`  |

### Parameters

- `wan` `(bool: false)` - Specifies to list WAN members instead of the LAN
  members (which is the default). This is only eligible for agents running in
  **server mode**. This is specified as part of the URL as a query parameter.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/agent/members
```

### Sample Response

```json
[
  {
    "Name": "foobar",
    "Addr": "10.1.10.12",
    "Port": 8301,
    "Tags": {
      "bootstrap": "1",
      "dc": "dc1",
      "port": "8300",
      "role": "consul"
    },
    "Status": 1,
    "ProtocolMin": 1,
    "ProtocolMax": 2,
    "ProtocolCur": 2,
    "DelegateMin": 1,
    "DelegateMax": 3,
    "DelegateCur": 3
  }
]
```

## Read Configuration

This endpoint returns the configuration and member information of the local
agent.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/agent/self`                | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `none`            | `agent:read` |

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/agent/self
```

### Sample Response

```json
{
  "Config": {
    "Bootstrap": true,
    "Server": true,
    "Datacenter": "dc1",
    "DataDir": "/tmp/consul",
    "DNSRecursor": "",
    "DNSRecursors": [],
    "Domain": "consul.",
    "LogLevel": "INFO",
    "NodeID": "40e4a748-2192-161a-0510-9bf59fe950b5",
    "NodeName": "foobar",
    "ClientAddr": "127.0.0.1",
    "BindAddr": "0.0.0.0",
    "AdvertiseAddr": "10.1.10.12",
    "Ports": {
      "DNS": 8600,
      "HTTP": 8500,
      "RPC": 8400,
      "SerfLan": 8301,
      "SerfWan": 8302,
      "Server": 8300
    },
    "LeaveOnTerm": false,
    "SkipLeaveOnInt": false,
    "StatsiteAddr": "",
    "Protocol": 1,
    "EnableDebug": false,
    "VerifyIncoming": false,
    "VerifyOutgoing": false,
    "CAFile": "",
    "CertFile": "",
    "KeyFile": "",
    "StartJoin": [],
    "UiDir": "",
    "PidFile": "",
    "EnableSyslog": false,
    "RejoinAfterLeave": false
  },
  "Coord": {
    "Adjustment": 0,
    "Error": 1.5,
    "Vec": [0,0,0,0,0,0,0,0]
  },
  "Member": {
    "Name": "foobar",
    "Addr": "10.1.10.12",
    "Port": 8301,
    "Tags": {
      "bootstrap": "1",
      "dc": "dc1",
      "id": "40e4a748-2192-161a-0510-9bf59fe950b5",
      "port": "8300",
      "role": "consul",
      "vsn": "1",
      "vsn_max": "1",
      "vsn_min": "1"
    },
    "Status": 1,
    "ProtocolMin": 1,
    "ProtocolMax": 2,
    "ProtocolCur": 2,
    "DelegateMin": 2,
    "DelegateMax": 4,
    "DelegateCur": 4
  },
  "Meta": {
    "instance_type": "i2.xlarge",
    "os_version": "ubuntu_16.04",
  }
}
```

## Reload Agent

This endpoint instructs the agent to reload its configuration. Any errors
encountered during this process are returned.

Not all configuration options are reloadable. See the
[Reloadable Configuration](/docs/agent/options.html#reloadable-configuration)
section on the agent options page for details on which options are supported.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/agent/reload`              | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required  |
| ---------------- | ----------------- | ------------- |
| `NO`             | `none`            | `agent:write` |

### Sample Request

```text
$ curl \
    --request PUT \
    https://consul.rocks/v1/agent/reload
```

## Enable Maintenance Mode

This endpoint places the agent into "maintenance mode". During maintenance mode,
the node will be marked as unavailable and will not be present in DNS or API
queries. This API call is idempotent.

Maintenance mode is persistent and will be automatically restored on agent
restart.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/agent/maintenance`         | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `none`            | `node:write` |

### Parameters

- `enable` `(bool: <required>)` - Specifies whether to enable or disable
  maintenance mode. This is specified as part of the URL as a query string
  parameter.

- `reason` `(string: "")` - Specifies a text string explaining the reason for
  placing the node into maintenance mode. This is simply to aid human operators.
  If no reason is provided, a default value will be used instead. This is
  specified as part of the URL as a query string parameter, and, as such, must
  be URI-encoded.

### Sample Request

```text
$ curl \
    --request PUT \
    https://consul.rocks/v1/agent/maintenance?enable=true&reason=For+API+docs
```

## Stream Logs

This endpoint streams logs from the local agent until the connection is closed.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/agent/monitor`             | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `none`            | `agent:read` |

### Parameters

- `loglevel` `(string: "info")` - Specifies a text string containing a log level
  to filter on, such as `info`.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/agent/monitor
```

### Sample Response

```text
YYYY/MM/DD HH:MM:SS [INFO] raft: Initial configuration (index=1): [{Suffrage:Voter ID:127.0.0.1:8300 Address:127.0.0.1:8300}]
YYYY/MM/DD HH:MM:SS [INFO] raft: Node at 127.0.0.1:8300 [Follower] entering Follower state (Leader: "")
YYYY/MM/DD HH:MM:SS [INFO] serf: EventMemberJoin: machine-osx 127.0.0.1
YYYY/MM/DD HH:MM:SS [INFO] consul: Adding LAN server machine-osx (Addr: tcp/127.0.0.1:8300) (DC: dc1)
YYYY/MM/DD HH:MM:SS [INFO] serf: EventMemberJoin: machine-osx.dc1 127.0.0.1
YYYY/MM/DD HH:MM:SS [INFO] consul: Handled member-join event for server "machine-osx.dc1" in area "wan"
# ...
```

## Join Agent

This endpoint instructs the agent to attempt to connect to a given address.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/agent/join/:address`       | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required  |
| ---------------- | ----------------- | ------------- |
| `NO`             | `none`            | `agent:write` |

### Parameters

- `address` `(string: <required>)` - Specifies the address of the other agent to
  join. This is specified as part of the URL.

- `wan` `(bool: false)` - Specifies to try and join over the WAN pool. This is
  only optional for agents running in server mode. This is specified as part of
  the URL as a query parameter

### Sample Request

```text
$ curl \
    https://consul.rocks/agent/join/1.2.3.4
```

## Graceful Leave and Shutdown

This endpoint triggers a graceful leave and shutdown of the agent. It is used to
ensure other nodes see the agent as "left" instead of "failed". Nodes that leave
will not attempt to re-join the cluster on restarting with a snapshot.

For nodes in server mode, the node is removed from the Raft peer set in a
graceful manner. This is critical, as in certain situations a non-graceful leave
can affect cluster availability.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/agent/leave`               | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required  |
| ---------------- | ----------------- | ------------- |
| `NO`             | `none`            | `agent:write` |

### Sample Request

```text
$ curl \
    --request PUT \
    https://consul.rocks/v1/agent/leave
```

## Force Leave and Shutdown

This endpoint instructs the agent to force a node into the `left` state. If a
node fails unexpectedly, then it will be in a `failed` state. Once in the
`failed` state, Consul will attempt to reconnect, and the services and checks
belonging to that node will not be cleaned up. Forcing a node into the `left`
state allows its old entries to be removed.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/agent/force-leave`         | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required  |
| ---------------- | ----------------- | ------------- |
| `NO`             | `none`            | `agent:write` |

### Sample Request

```text
$ curl \
    --request PUT \
    https://consul.rocks/v1/agent/force-leave
```
