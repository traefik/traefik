---
layout: api
page_title: Prepared Queries - HTTP API
sidebar_current: api-query
description: |-
  The /query endpoints manage and execute prepared queries in Consul.
---

# Prepared Query HTTP Endpoint

The `/query` endpoints create, update, destroy, and execute prepared queries.

Prepared queries allow you to register a complex service query and then execute
it later via its ID or name to get a set of healthy nodes that provide a given
service. This is particularly useful in combination with Consul's
[DNS Interface](/docs/agent/dns.html) as it allows for much richer queries than
would be possible given the limited entry points exposed by DNS.

See the [ACL Guide](/docs/guides/acl.html#prepared_query_acls)
prepared query section for more details about how prepared query policies work.

### Prepared Query Templates

Consul 0.6.4 and later support prepared query templates. These are created
similar to static templates, except with some additional fields and features.
Here is an example prepared query template:

```json
{
  "Template": {
    "Type": "name_prefix_match",
    "Regexp": "^geo-db-(.*?)-([^\\-]+?)$"
  }
}
```

The `Template` structure configures a prepared query as a template instead of a
static query. It has two fields:

- `Type` is the query type, which must be `name_prefix_match`. This means that
  the template will apply to any query lookup with a name whose prefix matches
  the `Name` field of the template. In this example, any query for `geo-db` will
  match this query. Query templates are resolved using a longest prefix match,
  so it's possible to have high-level templates that are overridden for specific
  services. Static queries are always resolved first, so they can also override
  templates.

- `Regexp` is an optional regular expression which is used to extract fields
  from the entire name, once this template is selected. In this example, the
  regular expression takes the first item after the "-" as the database name and
  everything else after as a tag. See the
  [RE2](https://github.com/google/re2/wiki/Syntax) reference for syntax of this
  regular expression.

All other fields of the query have the same meanings as for a static query,
except that several interpolation variables are available to dynamically
populate the query before it is executed. All of the string fields inside the
`Service` structure are interpolated, with the following variables available:

- `${name.full}` has the entire name that was queried. For example, a DNS lookup
  for `geo-db-customer-primary.query.consul` in the example above would set this
  variable to `geo-db-customer-primary`.

- `${name.prefix}` has the prefix that matched. This would always be `geo-db`
  for the example above.

- `${name.suffix}` has the suffix after the prefix. For example, a DNS lookup
  for `geo-db-customer-primary.query.consul` in the example above would set this
  variable to `-customer-primary`.

- `${match(N)}` returns the regular expression match at the given index N. The 0
  index will have the entire match, and >0 will have the results of each match
  group. For example, a DNS lookup for `geo-db-customer-primary.query.consul` in
  the example above with a `Regexp` field set to `^geo-db-(.*?)-([^\-]+?)$`
  would return `geo-db-customer-primary` for `${match(0)}`, `customer` for
  `${match(1)}`, and `primary` for `${match(2)}`. If the regular expression
  doesn't match, or an invalid index is given, then `${match(N)}` will return an
  empty string.

Using templates, it is possible to apply prepared query behaviors to many
services with a single template. Here's an example template that matches any
query and applies a failover policy to it:

```json
{
  "Name": "",
  "Template": {
    "Type": "name_prefix_match"
  },
  "Service": {
    "Service": "${name.full}",
    "Failover": {
      "NearestN": 3
    }
  }
}
```

This will match any lookup for `*.query.consul` and will attempt to find the
service locally, and otherwise attempt to find that service in the next three
closest datacenters. If ACLs are enabled, a catch-all template like this with
an empty `Name` requires an ACL token that can write to any query prefix. Also,
only a single catch-all template can be registered at any time.

## Create Prepared Query

This endpoint creates a new prepared query and returns its ID if it is created
successfully.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/query`                     | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required  |
| ---------------- | ----------------- | ------------- |
| `NO`             | `none`            | `query:write` |

### Parameters

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

- `Name` `(string: "")` - Specifies an optional friendly name that can be used
  to execute a query instead of using its ID.

- `Session` `(string: "")` - Specifies the ID of an existing session. This
  provides a way to automatically remove a prepared query when the given session
  is invalidated. If not given the prepared query must be manually removed when
  no longer needed.

- `Token` `(string: "")` - Specifies the ACL token to use each time the query is
  executed. This allows queries to be executed by clients with lesser or even no
  ACL Token, so this should be used with care. The token itself can only be seen
  by clients with a management token. If the `Token` field is left blank or
  omitted, the client's ACL Token will be used to determine if they have access
  to the service being queried. If the client does not supply an ACL Token, the
  anonymous token will be used.

- `Near` `(string: "")` - Specifies a node to sort near based on distance
  sorting using [Network Coordinates](/docs/internals/coordinates.html). The
  nearest instance to the specified node will be returned first, and subsequent
  nodes in the response will be sorted in ascending order of estimated
  round-trip times. If the node given does not exist, the nodes in the response
  will be shuffled. Using `_agent` is supported, and will automatically return
  results nearest the agent servicing the request. If unspecified, the response
  will be shuffled by default.

- `Service` `(Service: <required>)` - Specifies the structure to define the query's behavior.

  - `Service` `(string: <required>)` - Specifies the name of the service to
    query.

  - `Failover` contains two fields, both of which are optional, and determine
    what happens if no healthy nodes are available in the local datacenter when
    the query is executed. It allows the use of nodes in other datacenters with
    very little configuration.

      - `NearestN` `(int: 0)` - Specifies that the query will be forwarded to up
        to `NearestN` other datacenters based on their estimated network round
        trip time using [Network Coordinates](/docs/internals/coordinates.html)
        from the WAN gossip pool. The median round trip time from the server
        handling the query to the servers in the remote datacenter is used to
        determine the priority.

      - `Datacenters` `(array<string>: nil)` - Specifies a fixed list of remote
        datacenters to forward the query to if there are no healthy nodes in the
        local datacenter. Datacenters are queried in the order given in the
        list. If this option is combined with `NearestN`, then the `NearestN`
        queries will be performed first, followed by the list given by
        `Datacenters`. A given datacenter will only be queried one time during a
        failover, even if it is selected by both `NearestN` and is listed in
        `Datacenters`.

  - `OnlyPassing` `(bool: false)` - Specifies the behavior of the query's health
    check filtering. If this is set to false, the results will include nodes
    with checks in the passing as well as the warning states. If this is set to
    true, only nodes with checks in the passing state will be returned.

  - `Tags` `(array<string>: nil)` - Specifies a list of service tags to filter
    the query results. For a service to pass the tag filter it must have *all*
    of the required tags, and *none* of the excluded tags (prefixed with `!`).

  - `NodeMeta` `(map<string|string>: nil)` - Specifies a list of user-defined
    key/value pairs that will be used for filtering the query results to nodes
    with the given metadata values present.

- `DNS` `(DNS: nil)` - Specifies DNS configuration

  - `TTL` `(string: "")` - Specifies the TTL duration when query results are
    served over DNS. If this is specified, it will take precedence over any
    Consul agent-specific configuration.

### Sample Payload

```json
{
  "Name": "my-query",
  "Session": "adf4238a-882b-9ddc-4a9d-5b6758e4159e",
  "Token": "",
  "Service": {
    "Service": "redis",
    "Failover": {
      "NearestN": 3,
      "Datacenters": ["dc1", "dc2"]
    },
    "Near": "node1",
    "OnlyPassing": false,
    "Tags": ["primary", "!experimental"],
    "NodeMeta": {"instance_type": "m3.large"}
  },
  "DNS": {
    "TTL": "10s"
  }
}
```

### Sample Request

```text
$ curl \
    --request PUT \
    --data @payload.json \
    https://consul.rocks/v1/query
```

### Sample Response

```json
{
  "ID": "8f246b77-f3e1-ff88-5b48-8ec93abf3e05"
}
```

## Read Prepared Query

This endpoint returns a list of all prepared queries.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/query`                     | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `none`            | `query:read` |

### Parameters

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/query
```

### Sample Response

```json
[
  {
    "ID": "8f246b77-f3e1-ff88-5b48-8ec93abf3e05",
    "Name": "my-query",
    "Session": "adf4238a-882b-9ddc-4a9d-5b6758e4159e",
    "Token": "<hidden>",
    "Service": {
      "Service": "redis",
      "Failover": {
        "NearestN": 3,
        "Datacenters": ["dc1", "dc2"]
      },
      "OnlyPassing": false,
      "Tags": ["primary", "!experimental"],
      "NodeMeta": {"instance_type": "m3.large"}
    },
    "DNS": {
      "TTL": "10s"
    },
    "RaftIndex": {
      "CreateIndex": 23,
      "ModifyIndex": 42
    }
  }
]
```

### Update Prepared Query

This endpoint updates an existing prepared query. If no query exists by the
given ID, an error is returned.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/query/:uuid`               | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required  |
| ---------------- | ----------------- | ------------- |
| `NO`             | `none`            | `query:write` |

### Parameters

- `uuid` `(string: <required>)` - Specifies the UUID of the query to update.
  This is required and is specified as part of the URL path.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

The body is the same as is used to create a prepared query. Please see above for
more information.

### Sample Response

```text
$ curl \
    --request PUT \
    --data @payload.json \
    https://consul.rocks/v1/query/8f246b77-f3e1-ff88-5b48-8ec93abf3e05
```

## Read Prepared Query

This endpoint reads an existing prepared query. If no query exists by the
given ID, an error is returned.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/query/:uuid`               | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `none`            | `query:read` |

### Parameters

- `uuid` `(string: <required>)` - Specifies the UUID of the query to read.
  This is required and is specified as part of the URL path.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

### Sample Request

```text
$ curl \
  https://consul.rocks/v1/query/8f246b77-f3e1-ff88-5b48-8ec93abf3e05
```

### Sample Response

The returned response is the same as the list of prepared queries above, only
with a single item present.

## Delete Prepared Query

This endpoint deletes an existing prepared query. If no query exists by the
given ID, an error is returned.

| Method    | Path                         | Produces                   |
| --------- | ---------------------------- | -------------------------- |
| `DELETE`  | `/query/:uuid`               | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required  |
| ---------------- | ----------------- | ------------- |
| `NO`             | `none`            | `query:write` |

### Parameters

- `uuid` `(string: <required>)` - Specifies the UUID of the query to delete.
  This is required and is specified as part of the URL path.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

### Sample Request

```text
$ curl \
    --request DELETE \
    https://consul.rocks/v1/query/8f246b77-f3e1-ff88-5b48-8ec93abf3e05
```

## Execute Prepared Query

This endpoint executes an existing prepared query. If no query exists by the
given ID, an error is returned.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/query/:uuid/execute`       | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `none`            | `depends`<sup>1</sup>    |

<sup>1</sup> If an ACL Token was bound to the query when it was defined then it
will be used when executing the request. Otherwise, the client's supplied ACL
Token will be used.

### Parameters

- `uuid` `(string: <required>)` - Specifies the UUID of the query to execute.
  This is required and is specified as part of the URL path. This can also be
  the name of an existing prepared query, or a name that matches a prefix name
  for a prepared query template.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

- `near` `(string: "")` - Specifies to sort the resulting list in ascending
  order based on the estimated round trip time from that node. Passing
  `?near=_agent` will use the agent's node for the sort. If this is not present,
  the default behavior will shuffle the nodes randomly each time the query is
  executed.

- `limit` `(int: 0)` - Limit the size of the list to the given number of nodes.
  This is applied after any sorting or shuffling.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/query/8f246b77-f3e1-ff88-5b48-8ec93abf3e05/execute?near=_agent
```

### Sample Response

```json
{
  "Service": "redis",
  "Nodes": [
    {
      "Node": {
        "ID": "40e4a748-2192-161a-0510-9bf59fe950b5",
        "Node": "foobar",
        "Address": "10.1.10.12",
        "TaggedAddresses": {
          "lan": "10.1.10.12",
          "wan": "10.1.10.12"
        },
        "NodeMeta": {"instance_type": "m3.large"}
      },
      "Service": {
        "ID": "redis",
        "Service": "redis",
        "Tags": null,
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
      ],
    "DNS": {
      "TTL": "10s"
    },
    "Datacenter": "dc3",
    "Failovers": 2
  }]
}
```

- `Nodes` contains the list of healthy nodes providing the given service, as
  specified by the constraints of the prepared query.

- `Service` has the service name that the query was selecting. This is useful
  for context in case an empty list of nodes is returned.

- `DNS` has information used when serving the results over DNS. This is just a
  copy of the structure given when the prepared query was created.

- `Datacenter` has the datacenter that ultimately provided the list of nodes and
  `Failovers` has the number of remote datacenters that were queried while
  executing the query. This provides some insight into where the data came from.
  This will be zero during non-failover operations where there were healthy
  nodes found in the local datacenter.

## Explain Prepared Query

This endpoint generates a fully-rendered query for a given name, post
interpolation.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/query/:uuid/explain`       | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | ACL Required |
| ---------------- | ----------------- | ------------ |
| `NO`             | `none`            | `query:read` |

### Parameters

- `uuid` `(string: <required>)` - Specifies the UUID of the query to explain.
  This is required and is specified as part of the URL path. This can also be
  the name of an existing prepared query, or a name that matches a prefix name
  for a prepared query template.

- `dc` `(string: "")` - Specifies the datacenter to query. This will default to
  the datacenter of the agent being queried. This is specified as part of the
  URL as a query parameter.

### Sample Request

```text
$ curl \
    https://consul.rocks/v1/query/8f246b77-f3e1-ff88-5b48-8ec93abf3e05/explain
```

### Sample Response

```json
{
  "Query": {
    "ID": "8f246b77-f3e1-ff88-5b48-8ec93abf3e05",
    "Name": "my-query",
    "Session": "adf4238a-882b-9ddc-4a9d-5b6758e4159e",
    "Token": "<hidden>",
    "Name": "geo-db",
    "Template": {
      "Type": "name_prefix_match",
      "Regexp": "^geo-db-(.*?)-([^\\-]+?)$"
    },
    "Service": {
      "Service": "mysql-customer",
      "Failover": {
        "NearestN": 3,
        "Datacenters": ["dc1", "dc2"]
      },
      "OnlyPassing": true,
      "Tags": ["primary"],
      "NodeMeta": {"instance_type": "m3.large"}
    }
  }
}
```
