---
layout: api
page_title: Check - Agent - HTTP API
sidebar_current: api-agent-check
description: |-
  The /agent/check endpoints interact with checks on the local agent in Consul.
---

# Check - Agent HTTP API

The `/agent/check` endpoints interact with checks on the local agent in Consul.
These should not be confused with checks in the catalog.

## List Checks

This endpoint returns all checks that are registered with the local agent. These
checks were either provided through configuration files or added dynamically
using the HTTP API.

It is important to note that the checks known by the agent may be different from
those reported by the catalog. This is usually due to changes being made while
there is no leader elected. The agent performs active
[anti-entropy](/docs/internals/anti-entropy.html), so in most situations
everything will be in sync within a few seconds.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/agent/checks`              | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required             |
| ---------------- | ----------------- | ------------------------ |
| `NO`             | `none`            | `node:read,service:read` |

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/agent/checks
```

### Sample Response

```json
{
  "service:redis": {
    "Node": "foobar",
    "CheckID": "service:redis",
    "Name": "Service 'redis' check",
    "Status": "passing",
    "Notes": "",
    "Output": "",
    "ServiceID": "redis",
    "ServiceName": "redis"
  }
}
```

## Register Check

This endpoint adds a new check to the local agent. Checks may be of script,
HTTP, TCP, or TTL type. The agent is responsible for managing the status of the
check and keeping the Catalog in sync.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/agent/check/register`      | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required               |
| ---------------- | ----------------- | -------------------------- |
| `NO`             | `none`            | `node:write,service:write` |

### Parameters

- `Name` `(string: <required>)` - Specifies the name of the check.

- `ID` `(string: "")` - Specifies a unique ID for this check in the cluster.
  This defaults to the `"Name"` parameter, but it may be necessary to provide an
  ID for uniqueness.

- `Interval` `(string: "")` - Specifies the frequency at which to run this
  check. This is required for HTTP and TCP checks.

- `Notes` `(string: "")` - Specifies arbitrary information for humans. This is
  not used by Consul internally.

- `DeregisterCriticalServiceAfter` `(string: "")` - Specifies that checks
  associated with a service should deregister after this time. This is specified
  as a time duration with suffix like "10m". If a check is in the critical state
  for more than this configured value, then its associated service (and all of
  its associated checks) will automatically be deregistered. The minimum timeout
  is 1 minute, and the process that reaps critical services runs every 30
  seconds, so it may take slightly longer than the configured timeout to trigger
  the deregistration. This should generally be configured with a timeout that's
  much, much longer than any expected recoverable outage for the given service.

- `Script` `(string: "")` - Specifies a script or path to a script to run on
  `Interval` to update the status of the check. If specifying a path, this path
  must exist on disk and be readable by the Consul agent.

- `DockerContainerID` `(string: "")` - Specifies that the check is a Docker
  check, and Consul will evaluate the script every `Interval` in the given
  container using the specified `Shell`. Note that `Shell` is currently only
  supported for Docker checks.

- `HTTP` `(string: "")` - Specifies an `HTTP` check to perform a `GET` request
  against the value of `HTTP` (expected to be a URL) every `Interval`. If the
  response is any `2xx` code, the check is `passing`. If the response is `429
  Too Many Requests`, the check is `warning`. Otherwise, the check is
  `critical`. HTTP checks also support SSL. By default, a valid SSL certificate
  is expected. Certificate verification can be controlled using the
  `TLSSkipVerify`.

- `TLSSkipVerify` `(bool: false)` - Specifies if the certificate for an HTTPS
  check should not be verified.

- `TCP` `(string: "")` - Specifies a `TCP` to connect against the value of `TCP`
  (expected to be an IP or hostname plus port combination) every `Interval`. If
  the connection attempt is successful, the check is `passing`. If the
  connection attempt is unsuccessful, the check is `critical`. In the case of a
  hostname that resolves to both IPv4 and IPv6 addresses, an attempt will be
  made to both addresses, and the first successful connection attempt will
  result in a successful check.

- `TTL` `(string: "")` - Specifies this is a TTL check, and the TTL endpoint
  must be used periodically to update the state of the check.

- `ServiceID` `(string: "")` - Specifies the ID of a service to associate the
  registered check with an existing service provided by the agent.

- `Status` `(string: "")` - Specifies the initial status of the health check.

### Sample Payload

```json
{
  "ID": "mem",
  "Name": "Memory utilization",
  "Notes": "Ensure we don't oversubscribe memory",
  "DeregisterCriticalServiceAfter": "90m",
  "Script": "/usr/local/bin/check_mem.py",
  "DockerContainerID": "f972c95ebf0e",
  "Shell": "/bin/bash",
  "HTTP": "http://example.com",
  "TCP": "example.com:22",
  "Interval": "10s",
  "TTL": "15s",
  "TLSSkipVerify": true
}
```

### Sample Request

```text
$ curl \
   --request PUT \
   --data @payload.json \
   https://consul.rocks/v1/agent/check/register
```

## Deregister Check

This endpoint remove a check from the local agent. The agent will take care of
deregistering the check from the catalog. If the check with the provided ID does
not exist, no action is taken.

| Method | Path                                | Produces                   |
| ------ | ----------------------------------- | -------------------------- |
| `PUT`  | `/agent/check/deregister/:check_id` | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required               |
| ---------------- | ----------------- | -------------------------- |
| `NO`             | `none`            | `node:write,service:write` |

### Parameters

- `check_id` `(string: "")` - Specifies the unique ID of the check to
  deregister. This is specified as part of the URL.

### Sample Request

```text
$ curl \
    --request PUT \
    https://consul.rocks/v1/agent/check/deregister/my-check-id
```

## TTL Check Pass

This endpoint is used with a TTL type check to set the status of the check to
`passing` and to reset the TTL clock.

| Method | Path                          | Produces                   |
| ------ | ----------------------------- | -------------------------- |
| `GET`  | `/agent/check/pass/:check_id` | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required               |
| ---------------- | ----------------- | -------------------------- |
| `NO`             | `none`            | `node:write,service:write` |

### Parameters

- `check_id` `(string: "")` - Specifies the unique ID of the check to
  use. This is specified as part of the URL.

- `note` `(string: "")` - Specifies a human-readable message. This will be
  passed through to the check's `Output` field.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/agent/check/pass/my-check-id
```

## TTL Check Warn

This endpoint is used with a TTL type check to set the status of the check to
`warning` and to reset the TTL clock.

| Method | Path                          | Produces                   |
| ------ | ----------------------------- | -------------------------- |
| `GET`  | `/agent/check/warn/:check_id` | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required               |
| ---------------- | ----------------- | -------------------------- |
| `NO`             | `none`            | `node:write,service:write` |

### Parameters

- `check_id` `(string: "")` - Specifies the unique ID of the check to
  use. This is specified as part of the URL.

- `note` `(string: "")` - Specifies a human-readable message. This will be
  passed through to the check's `Output` field.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/agent/check/warn/my-check-id
```

## TTL Check Fail

This endpoint is used with a TTL type check to set the status of the check to
`critical` and to reset the TTL clock.

| Method | Path                          | Produces                   |
| ------ | ----------------------------- | -------------------------- |
| `GET`  | `/agent/check/fail/:check_id` | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required               |
| ---------------- | ----------------- | -------------------------- |
| `NO`             | `none`            | `node:write,service:write` |

### Parameters

- `check_id` `(string: "")` - Specifies the unique ID of the check to
  use. This is specified as part of the URL.

- `note` `(string: "")` - Specifies a human-readable message. This will be
  passed through to the check's `Output` field.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/agent/check/fail/my-check-id
```

## TTL Check Update

This endpoint is used with a TTL type check to set the status of the check and
to reset the TTL clock.

| Method | Path                            | Produces                   |
| ------ | ------------------------------- | -------------------------- |
| `PUT`  | `/agent/check/update/:check_id` | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required               |
| ---------------- | ----------------- | -------------------------- |
| `NO`             | `none`            | `node:write,service:write` |

### Parameters

- `check_id` `(string: "")` - Specifies the unique ID of the check to
  use. This is specified as part of the URL.

- `Status` `(string: "")` - Specifies the status of the check. Valid values are
  `"passing"`, `"warning"`, and `"critical"`.

- `Output` `(string: "")` - Specifies a human-readable message. This will be
  passed through to the check's `Output` field.

### Sample Payload

```json
{
  "Status": "passing",
  "Output": "curl reported a failure:\n\n..."
}
```

### Sample Request

```text
$ curl \
    --request PUT \
    --data @payload.json \
    https://consul.rocks/v1/agent/check/update/my-check-id
```
