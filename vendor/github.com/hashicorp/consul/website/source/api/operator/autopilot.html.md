---
layout: api
page_title: Autopilot - Operator - HTTP API
sidebar_current: api-operator-autopilot
description: |-
  The /operator/autopilot endpoints allow for automatic operator-friendly
  management of Consul servers including cleanup of dead servers, monitoring
  the state of the Raft cluster, and stable server introduction.
---

# Autopilot Operator HTTP API

The `/operator/autopilot` endpoints allow for automatic operator-friendly
management of Consul servers including cleanup of dead servers, monitoring
the state of the Raft cluster, and stable server introduction.

Please see the [Autopilot Guide](/docs/guides/autopilot.html) for more details.

## Read Configuration

This endpoint retrieves its latest Autopilot configuration.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/operator/autopilot/configuration` | `application/json` |

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
    https://consul.rocks/operator/autopilot/configuration
```

### Sample Response

```json
{
  "CleanupDeadServers": true,
  "LastContactThreshold": "200ms",
  "MaxTrailingLogs": 250,
  "ServerStabilizationTime": "10s",
  "RedundancyZoneTag": "",
  "DisableUpgradeMigration": false,
  "CreateIndex": 4,
  "ModifyIndex": 4
}
```

For more information about the Autopilot configuration options, see the
[agent configuration section](/docs/agent/options.html#autopilot).

## Update Configuration

This endpoint updates the Autopilot configuration of the cluster.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/operator/autopilot/configuration` | `application/json` |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required     |
| ---------------- | ----------------- | ---------------- |
| `NO`             | `none`            | `opreator:write` |

### Parameters

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query string.

- `cas` `(int: 0)` - Specifies to use a Check-And-Set operation. The update will
  only happen if the given index matches the `ModifyIndex` of the configuration
  at the time of writing.

- `CleanupDeadServers` `(bool: true)` - Specifies automatic removal of dead
  server nodes periodically and whenever a new server is added to the cluster.

- `LastContactThreshold` `(string: "200ms")` - Specifies the maximum amount of
  time a server can go without contact from the leader before being considered
  unhealthy. Must be a duration value such as `10s`.

- `MaxTrailingLogs` `(int: 250)` specifies the maximum number of log entries
  that a server can trail the leader by before being considered unhealthy.

- `ServerStabilizationTime` `(string: "10s")` - Specifies the minimum amount of
  time a server must be stable in the 'healthy' state before being added to the
  cluster. Only takes effect if all servers are running Raft protocol version 3
  or higher. Must be a duration value such as `30s`.

- `RedundancyZoneTag` `(string: "")` controls the node-meta key to use when
  Autopilot is separating servers into zones for redundancy. Only one server in
  each zone can be a voting member at one time. If left blank, this feature will
  be disabled.

- `DisableUpgradeMigration` `(bool: false)` - Disables Autopilot's upgrade
  migration strategy in Consul Enterprise of waiting until enough
  newer-versioned servers have been added to the cluster before promoting any of
  them to voters.

### Sample Payload

```json
{
  "CleanupDeadServers": true,
  "LastContactThreshold": "200ms",
  "MaxTrailingLogs": 250,
  "ServerStabilizationTime": "10s",
  "RedundancyZoneTag": "",
  "DisableUpgradeMigration": false,
  "CreateIndex": 4,
  "ModifyIndex": 4
}
```

## Read Health

This endpoint queries the health of the autopilot status.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/operator/autopilot/health` | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required    |
| ---------------- | ----------------- | --------------- |
| `NO`             | `none`            | `opreator:read` |

### Parameters

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query string.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/operator/autopilot/health
```

### Sample response

```json
{
  "Healthy": true,
  "FailureTolerance": 0,
  "Servers": [
    {
      "ID": "e349749b-3303-3ddf-959c-b5885a0e1f6e",
      "Name": "node1",
      "Address": "127.0.0.1:8300",
      "SerfStatus": "alive",
      "Version": "0.7.4",
      "Leader": true,
      "LastContact": "0s",
      "LastTerm": 2,
      "LastIndex": 46,
      "Healthy": true,
      "Voter": true,
      "StableSince": "2017-03-06T22:07:51Z"
    },
    {
      "ID": "e36ee410-cc3c-0a0c-c724-63817ab30303",
      "Name": "node2",
      "Address": "127.0.0.1:8205",
      "SerfStatus": "alive",
      "Version": "0.7.4",
      "Leader": false,
      "LastContact": "27.291304ms",
      "LastTerm": 2,
      "LastIndex": 46,
      "Healthy": true,
      "Voter": false,
      "StableSince": "2017-03-06T22:18:26Z"
    }
  ]
}
```

- `Healthy` is whether all the servers are currently healthy.

- `FailureTolerance` is the number of redundant healthy servers that could be
  fail without causing an outage (this would be 2 in a healthy cluster of 5
  servers).

- `Servers` holds detailed health information on each server:

  - `ID` is the Raft ID of the server.

  - `Name` is the node name of the server.

  - `Address` is the address of the server.

  - `SerfStatus` is the SerfHealth check status for the server.

  - `Version` is the Consul version of the server.

  - `Leader` is whether this server is currently the leader.

  - `LastContact` is the time elapsed since this server's last contact with the leader.

  - `LastTerm` is the server's last known Raft leader term.

  - `LastIndex` is the index of the server's last committed Raft log entry.

  - `Healthy` is whether the server is healthy according to the current Autopilot configuration.

  - `Voter` is whether the server is a voting member of the Raft cluster.

  - `StableSince` is the time this server has been in its current `Healthy` state.
