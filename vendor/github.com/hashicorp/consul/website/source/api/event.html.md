---
layout: api
page_title: Events - HTTP API
sidebar_current: api-event
description: |-
  The /event endpoints fire new events and to query the available events in
  Consul.
---

# Event HTTP Endpoint

The `/event` endpoints fire new events and to query the available events in
Consul.

## Fire Event

This endpoint triggers a new user event.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/event/fire/:name`          | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required  |
| ---------------- | ----------------- | ------------- |
| `NO`             | `none`            | `event:write` |

### Parameters

- `name` `(string: <required>)` - Specifies the name of the event to fire. This
  is specified as part of the URL. This name must not start with an underscore,
  since those are reserved for Consul internally.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

- `node` `(string: "")` - Specifies a regular expression to filter by node name.
  This is specified as part of the URL as a query parameter.

- `service` `(string: "")` - Specifies a regular expression to filter by service
  name. This is specified as part of the URL as a query parameter.

- `tag` `(string: "")` - Specifies a regular expression to filter by tag. This
  is specified as part of the URL as a query parameter.

### Sample Payload

The body contents are opaque to Consul and become the "payload" that is passed
onto the receiver of the event.

```text
Lorem ipsum dolor sit amet, consectetur adipisicing elit...
```

### Sample Request

```text
$ curl \
    --request PUT \
    --data @payload \
    https://consul.rocks/v1/event/fire/my-event
```

### Sample Response

```json
{
  "ID": "b54fe110-7af5-cafc-d1fb-afc8ba432b1c",
  "Name": "deploy",
  "Payload": null,
  "NodeFilter": "",
  "ServiceFilter": "",
  "TagFilter": "",
  "Version": 1,
  "LTime": 0
}
```

- `ID` is a unique identifier the newly fired event

## List Events

This endpoint returns the most recent events known by the agent. As a
consequence of how the [event command](/docs/commands/event.html) works, each
agent may have a different view of the events. Events are broadcast using the
[gossip protocol](/docs/internals/gossip.html), so they have no global ordering
nor do they make a promise of delivery.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/event/list`                | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `YES`            | `none`            | `event:read` |

### Parameters

- `name` `(string: <required>)` - Specifies the name of the event to filter.
  This is specified as part of the URL as a query parameter.

- `node` `(string: "")` - Specifies a regular expression to filter by node name.
  This is specified as part of the URL as a query parameter.

- `service` `(string: "")` - Specifies a regular expression to filter by service
  name. This is specified as part of the URL as a query parameter.

- `tag` `(string: "")` - Specifies a regular expression to filter by tag. This
  is specified as part of the URL as a query parameter.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/event/list
```

### Sample Response

```json
[
  {
    "ID": "b54fe110-7af5-cafc-d1fb-afc8ba432b1c",
    "Name": "deploy",
    "Payload": "MTYwOTAzMA==",
    "NodeFilter": "",
    "ServiceFilter": "",
    "TagFilter": "",
    "Version": 1,
    "LTime": 19
  }
]
```

### Caveat

The semantics of this endpoint's blocking queries are slightly different. Most
blocking queries provide a monotonic index and block until a newer index is
available. This can be supported as a consequence of the total ordering of the
[consensus protocol](/docs/internals/consensus.html). With gossip, there is no
ordering, and instead `X-Consul-Index` maps to the newest event that matches the
query.

In practice, this means the index is only useful when used against a single
agent and has no meaning globally. Because Consul defines the index as being
opaque, clients should not be expecting a natural ordering either.
