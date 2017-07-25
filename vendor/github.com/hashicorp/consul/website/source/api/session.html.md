---
layout: api
page_title: Session - HTTP API
sidebar_current: api-session
description: |-
  The /session endpoints create, destroy, and query sessions in Consul.
---

# Session HTTP Endpoint

The `/session` endpoints create, destroy, and query sessions in Cosul.

## Create Session

This endpoint initializes a new session. Sessions must be associated with a
node and may be associated with any number of checks.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/session/create`            | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required    |
| ---------------- | ----------------- | --------------- |
| `NO`             | `none`            | `session:write` |

### Parameters

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter. Using this across datacenters is not recommended.

- `LockDelay` `(string: "15s")` - Specifies the duration for the lock delay.

- `Node` `(string: "<agent>")` - Specifies the name of the node. This must refer
  to a node that is already registered.

- `Name` `(string: "")` - Specifies a human-readable name for the session.

- `Checks` `(array<string>: nil)` - specifies a list of associated health
  checks. It is highly recommended that, if you override this list, you include
  the default `serfHealth`.

- `Behavior` `(string: "release")` - Controls the behavior to take when a
  session is invalidated. Valid values are:

  - `release` - causes any locks that are held to be released
  - `delete` - causes an locks that are held to be deleted

- `TTL` `(string: "")` - Specifies the number of seconds (between 10s and
  86400s). If provided, the session is invalidated if it is not renewed before
  the TTL expires. The lowest practical TTL should be used to keep the number of
  managed sessions low. When locks are forcibly expired, such as during a leader
  election, sessions may not be reaped for up to double this TTL, so long TTL
  values (> 1 hour) should be avoided.

### Sample Payload

```json
{
  "LockDelay": "15s",
  "Name": "my-service-lock",
  "Node": "foobar",
  "Checks": ["a", "b", "c"],
  "Behavior": "release",
  "TTL": "30s"
}
```

### Sample Request

```text
$ curl \
    --request PUT \
    --data @payload.json \
    https://consul.rocks/v1/session/create
```

### Sample Response

```javascript
{
  "ID": "adf4238a-882b-9ddc-4a9d-5b6758e4159e"
}
```

- `ID` - the ID of the created session

## Delete Session

This endpoint destroys the session with the given name. If the session UUID is
malformed, an error is returned. If the session UUID does not exist or already
expired, a 200 is still returned (the operation is idempotent).

| Method | Path                         | Produces                   |
| :----- | :--------------------------- | -------------------------- |
| `PUT`  | `/session/destroy/:uuid`     | `application/json`         |

Even though the Content-Type is `application/json`, the response is
either a literal `true` or `false`, indicating of whether the destroy was
successful.

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required    |
| ---------------- | ----------------- | --------------- |
| `NO`             | `none`            | `session:write` |

### Parameters

- `uuid` `(string: <required>)` - Specifies the UUID of the session to destroy.
  This is required and is specified as part of the URL path.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter. Using this across datacenters is not recommended.

### Sample Request

```text
$ curl \
    --request PUT
    https://consul.rocks/v1/session/destroy/adf4238a-882b-9ddc-4a9d-5b6758e4159e
```

### Sample Response

```json
true
```

## Read Session

This endpoint returns the requested session information.

| Method | Path                         | Produces                   |
| :----- | :--------------------------- | -------------------------- |
| `GET`  | `/session/info/:uuid`        | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required   |
| ---------------- | ----------------- | -------------- |
| `YES`            | `all`             | `session:read` |

### Parameters

- `uuid` `(string: <required>)` - Specifies the UUID of the session to read.
  This is required and is specified as part of the URL path.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter. Using this across datacenters is not recommended.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/session/info/adf4238a-882b-9ddc-4a9d-5b6758e4159e
```

### Sample Response

```json
[
  {
    "LockDelay": 1.5e+10,
    "Checks": [
      "serfHealth"
    ],
    "Node": "foobar",
    "ID": "adf4238a-882b-9ddc-4a9d-5b6758e4159e",
    "CreateIndex": 1086449
  }
]
```

If the session does not exist, `null` is returned instead of a JSON list.

## List Sessions for Node

This endpoint returns the active sessions for a given node.

| Method | Path                         | Produces                   |
| :----- | :--------------------------- | -------------------------- |
| `GET`  | `/session/node/:node`        | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required   |
| ---------------- | ----------------- | -------------- |
| `YES`            | `all`             | `session:read` |

### Parameters

- `node` `(string: <required>)` - Specifies the name or ID of the node to query.
  This is required and is specified as part of the URL path.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter. Using this across datacenters is not recommended.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/session/node/node-abcd1234
```

### Sample Response

```json
[
  {
    "LockDelay": 1.5e+10,
    "Checks": [
      "serfHealth"
    ],
    "Node": "foobar",
    "ID": "adf4238a-882b-9ddc-4a9d-5b6758e4159e",
    "CreateIndex": 1086449
  },
]
```

## List Sessions

This endpoint returns the list of active sessions.

| Method | Path                         | Produces                   |
| :----- | :--------------------------- | -------------------------- |
| `GET`  | `/session/list`              | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required   |
| ---------------- | ----------------- | -------------- |
| `YES`            | `all`             | `session:read` |

### Parameters

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter. Using this across datacenters is not recommended.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/session/list
```

### Sample Response

```json
[
  {
    "LockDelay": 1.5e+10,
    "Checks": [
      "serfHealth"
    ],
    "Node": "foobar",
    "ID": "adf4238a-882b-9ddc-4a9d-5b6758e4159e",
    "CreateIndex": 1086449
  },
]
```

## Renew Session

This endpoint renews the given session. This is used with sessions that have a
TTL, and it extends the expiration by the TTL.

| Method | Path                         | Produces                   |
| :----- | :--------------------------- | -------------------------- |
| `PUT`  | `/session/renew/:uuid`       | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required    |
| ---------------- | ----------------- | --------------- |
| `NO`             | `none`            | `session:write` |

### Parameters

- `uuid` `(string: <required>)` - Specifies the UUID of the session to renew.
  This is required and is specified as part of the URL path.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter. Using this across datacenters is not recommended.

### Sample Request

```text
$ curl \
    --request PUT \
    https://consul.rocks/v1/session/renew/adf4238a-882b-9ddc-4a9d-5b6758e4159e
```

### Sample Response

```json
[
  {
    "LockDelay": 1.5e+10,
    "Checks": [
      "serfHealth"
    ],
    "Node": "foobar",
    "ID": "adf4238a-882b-9ddc-4a9d-5b6758e4159e",
    "CreateIndex": 1086449,
    "Behavior": "release",
    "TTL": "15s"
  }
]
```

-> **Note:** Consul may return a TTL value higher than the one specified during session creation. This indicates the server is under high load and is requesting clients renew less often.
