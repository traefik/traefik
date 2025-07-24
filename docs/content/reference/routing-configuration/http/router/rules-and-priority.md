---
title: "Traefik HTTP Routers Rules & Priority Documentation"
description: "In Traefik Proxy, an HTTP router is in charge of connecting incoming requests to the Services that can handle them. Read the technical documentation."
---

An HTTP router is in charge of connecting incoming requests to the services that can handle them. Traefik allows you to define your matching rules and [prioritize](#priority-calculation) the routes.

## Rules

Rules are a set of matchers configured with values, that determine if a particular request matches a specific criteria. 
If the rule is verified, the router becomes active, calls middlewares, and then forwards the request to the service.

- The character `@` is not authorized in the router name.
- To set the value of a rule, use [backticks](https://en.wiktionary.org/wiki/backtick) ` or escaped double-quotes ``\"``.
- Single quotes ' are not accepted since the values are [Go's String Literals](https://golang.org/ref/spec#String_literals).
- Regular Expressions:
    - Matchers that accept a regexp as their value use a [Go](https://golang.org/pkg/regexp/) flavored syntax.
    - The usual `AND` (&&) and `OR` (||) logical operators can be used, with the expected precedence rules, as well as parentheses to express complex rules.
    - The `NOT` (!) operator allows you to invert the matcher.

The table below lists all the available matchers:

| Matcher                                                         | Description                                                                    |
|-----------------------------------------------------------------|:-------------------------------------------------------------------------------|
| [```Header(`key`, `value`)```](#header-and-headerregexp)        | Matches requests containing a header named `key` set to `value`.               |
| [```HeaderRegexp(`key`, `regexp`)```](#header-and-headerregexp) | Matches requests containing a header named `key` matching `regexp`.            |
| [```Host(`domain`)```](#host-and-hostregexp)                    | Matches requests host set to `domain`.                                         |
| [```HostRegexp(`regexp`)```](#host-and-hostregexp)              | Matches requests host matching `regexp`.                                       |
| [```Method(`method`)```](#method)                               | Matches requests method set to `method`.                                       |
| [```Path(`path`)```](#path-pathprefix-and-pathregexp)           | Matches requests path set to `path`.                                           |
| [```PathPrefix(`prefix`)```](#path-pathprefix-and-pathregexp)   | Matches requests path prefix set to `prefix`.                                  |
| [```PathRegexp(`regexp`)```](#path-pathprefix-and-pathregexp)   | Matches request path using `regexp`.                                           |
| [```Query(`key`, `value`)```](#query-and-queryregexp)           | Matches requests query parameters named `key` set to `value`.                  |
| [```QueryRegexp(`key`, `regexp`)```](#query-and-queryregexp)    | Matches requests query parameters named `key` matching `regexp`.               |
| [```ClientIP(`ip`)```](#clientip)                               | Matches requests client IP using `ip`. It accepts IPv4, IPv6 and CIDR formats. |

### Header and HeaderRegexp

The `Header` and `HeaderRegexp` matchers allow matching requests that contain specific header.

| Behavior                                                        | Rule                                                                    |
|-----------------------------------------------------------------|:------------------------------------------------------------------------|
| Match requests with a `Content-Type` header set to `application/yaml`.| ```Header(`Content-Type`, `application/yaml`)``` |
| Match requests with a `Content-Type` header set to either `application/json` or `application/yaml`. | ```HeaderRegexp(`Content-Type`, `^application/(json\|yaml)$`)``` |
| Match headers [case-insensitively](https://en.wikipedia.org/wiki/Case_sensitivity). | ```HeaderRegexp(`Content-Type`, `(?i)^application/(json\|yaml)$`)``` |

### Host and HostRegexp

The `Host` and `HostRegexp` matchers allow matching requests that are targeted to a given host.

These matchers do not support non-ASCII characters, use punycode encoded values ([rfc 3492](https://tools.ietf.org/html/rfc3492)) to match such domains.

If no `Host` is set in the request URL (for example, it's an IP address), these matchers will look at the `Host` header.

These matchers will match the request's host in lowercase.

| Behavior                                                        | Rule                                                                    |
|-----------------------------------------------------------------|:------------------------------------------------------------------------|
| Match requests with `Host` set to `example.com`. | ```Host(`example.com`)``` |
| Match requests sent to any subdomain of `example.com`. | ```HostRegexp(`^.+\.example\.com$`)``` |
| Match requests with `Host` set to either `example.com` or `example.org`. | ```HostRegexp(`^example\.(com\|org)$`)``` |
| Match `Host` [case-insensitively](https://en.wikipedia.org/wiki/Case_sensitivity). | ```HostRegexp(`(?i)^example\.(com\|org)$`)``` |

### Method

The `Method` matchers allows matching requests sent based on their HTTP method (also known as request verb).

| Behavior                                                        | Rule                                                                    |
|-----------------------------------------------------------------|:------------------------------------------------------------------------|
| Match `OPTIONS` requests. | ```Method(`OPTIONS`)``` |

### Path, PathPrefix, and PathRegexp

These matchers allow matching requests based on their URL path.

For exact matches, use `Path` and its prefixed alternative `PathPrefix`, for regexp matches, use `PathRegexp`.

Path are always starting with a `/`, except for `PathRegexp`.

| Behavior                                                        | Rule                                                                    |
|-----------------------------------------------------------------|:------------------------------------------------------------------------|
| Match `/products` but neither `/products/shoes` nor `/products/`. | ```Path(`/products`)``` |
| Match `/products` as well as everything under `/products`, such as `/products/shoes`, `/products/` but also `/products-for-sale`. | ```PathPrefix(`/products`)``` |
| Match both `/products/shoes` and `/products/socks` with and ID like `/products/shoes/31`. | ```PathRegexp(`^/products/(shoes\|socks)/[0-9]+$`)``` |
| Match requests with a path ending in either `.jpeg`, `.jpg` or `.png`. | ```PathRegexp(`\.(jpeg\|jpg\|png)$`)``` |
| Match `/products` as well as everything under `/products`, such as `/products/shoes`, `/products/` but also `/products-for-sale`, [case-insensitively](https://en.wikipedia.org/wiki/Case_sensitivity). | ```HostRegexp(`(?i)^/products`)``` |

### Query and QueryRegexp

The `Query` and `QueryRegexp` matchers allow matching requests based on query parameters.

| Behavior                                                        | Rule                                                                    |
|-----------------------------------------------------------------|:------------------------------------------------------------------------|
| Match requests with a `mobile` query parameter set to `true`, such as in `/search?mobile=true`. | ```Query(`mobile`, `true`)``` |
| Match requests with a query parameter `mobile` that has no value, such as in `/search?mobile`. | ```Query(`mobile`)``` |
| Match requests with a `mobile` query parameter set to either `true` or `yes`. | ```QueryRegexp(`mobile`, `^(true\|yes)$`)``` |
| Match requests with a `mobile` query parameter set to any value (including the empty value). | ```QueryRegexp(`mobile`, `^.*$`)``` |
| Match query parameters [case-insensitively](https://en.wikipedia.org/wiki/Case_sensitivity). | ```QueryRegexp(`mobile`, `(?i)^(true\|yes)$`)``` |

### ClientIP

The `ClientIP` matcher allows matching requests sent from the given client IP.

It only matches the request client IP and does not use the `X-Forwarded-For` header for matching.

| Behavior                                                        | Rule                                                                    |
|-----------------------------------------------------------------|:------------------------------------------------------------------------|
| Match requests coming from a given IP (IPv4). | ```ClientIP(`10.76.105.11`)``` |
| Match requests coming from a given IP (IPv6). | ```ClientIP(`::1`)``` |
| Match requests coming from a given subnet (IPv4). | ```ClientIP(`192.168.1.0/24`)``` |
| Match requests coming from a given subnet (IPv6). | ```ClientIP(`fe80::/10`)``` |

### RuleSyntax

!!! warning

    RuleSyntax option is deprecated and will be removed in the next major version.
    Please do not use this field and rewrite the router rules to use the v3 syntax.

In Traefik v3 a new rule syntax has been introduced ([migration guide](../../../../migrate/v3.md)). the `ruleSyntax` option allows to configure the rule syntax to be used for parsing the rule on a per-router basis. This allows to have heterogeneous router configurations and ease migration.

The default value of the `ruleSyntax` option is inherited from the `defaultRuleSyntax` option in the install configuration (formerly known as static configuration). By default, the `defaultRuleSyntax` static option is v3, meaning that the default rule syntax is also v3

#### Configuration Example

The configuration below uses the [File Provider (Structured)](../../../install-configuration/providers/others/file.md) to configure the `ruleSyntax` to allow `Router-v2` to use v2 syntax, while for `Router-v3` it is configured to use v3 syntax.

```yaml tab="Structured (YAML)"
## Dynamic configuration
http:
  routers:
    Router-v3:
      rule: HostRegexp(`[a-z]+\\.traefik\\.com`)
      ruleSyntax: v3
    Router-v2:
      rule: HostRegexp(`{subdomain:[a-z]+}.traefik.com`)
      ruleSyntax: v2
```

```toml tab="Structured (TOML)"
## Dynamic configuration
[http.routers]
  [http.routers.Router-v3]
    rule = "HostRegexp(`[a-z]+\\.traefik\\.com`)"
    ruleSyntax = v3
  [http.routers.Router-v2]
    rule = "HostRegexp(`{subdomain:[a-z]+}.traefik.com`)"
    ruleSyntax = v2
```

```yaml tab="Labels"
labels:
  - "traefik.http.routers.Router-v3.rule=HostRegexp(`[a-z]+\\.traefik\\.com`)"
  - "traefik.http.routers.Router-v3.ruleSyntax=v3"
  - "traefik.http.routers.Router-v2.rule=HostRegexp(`{subdomain:[a-z]+}.traefik.com`)"
  - "traefik.http.routers.Router-v2.ruleSyntax=v2"
```

```json tab="Tags"
{
  // ...
  "Tags": [
    "traefik.http.routers.Router-v3.rule=HostRegexp(`[a-z]+\\.traefik\\.com`)",
    "traefik.http.routers.Router-v3.ruleSyntax=v3"
    "traefik.http.routers.Router-v2.rule=HostRegexp(`{subdomain:[a-z]+}.traefik.com`)",
    "traefik.http.routers.Router-v2.ruleSyntax=v2"
  ]
},
```

## Priority Calculation

??? info "How default priorities are computed"

    ```yaml tab="Structured (YAML)"
    http:
      routers:
        Router-1:
          rule: "HostRegexp(`[a-z]+\.traefik\.com`)"
          # ...
        Router-2:
          rule: "Host(`foobar.traefik.com`)"
          # ...
    ```

    ```toml tab="Structured (TOML)"
    [http.routers]
      [http.routers.Router-1]
        rule = "HostRegexp(`[a-z]+\\.traefik\\.com`)"
        # ...
      [http.routers.Router-2]
        rule = "Host(`foobar.traefik.com`)"
        # ...
    ```

    ```yaml tab="Labels"
    labels:
      - "traefik.http.routers.Router-1.rule=HostRegexp(`[a-z]+\\.traefik\\.com`)"
      - "traefik.http.routers.Router-2.rule=Host(`foobar.traefik.com`)"
    ```

    ```json tab="Tags"
    {
        // ...
        "Tags": [
          "traefik.http.routers.Router-1.rule=HostRegexp(`[a-z]+\\.traefik\\.com`)",
          "traefik.http.routers.Router-2.rule=Host(`foobar.traefik.com`)"
        ]
      }
    ```

    In this case, all requests with host `foobar.traefik.com` will be routed through `Router-1` instead of `Router-2`.

    | Name     | Rule                                     | Priority |
    |----------|------------------------------------------|----------|
    | Router-1 | ```HostRegexp(`[a-z]+\.traefik\.com`)``` | 34       |
    | Router-2 | ```Host(`foobar.traefik.com`)```         | 26       |

    The previous table shows that `Router-1` has a higher priority than `Router-2`.

    To solve this issue, the priority must be set.

To avoid path overlap, routes are sorted, by default, in descending order using rules length.
The priority is directly equal to the length of the rule, and so the longest length has the highest priority.

A value of `0` for the priority is ignored: `priority: 0` means that the default rules length sorting is used.

Traefik reserves a range of priorities for its internal routers, the maximum user-defined router priority value is:

- `(MaxInt32 - 1000)` for 32-bit platforms,
- `(MaxInt64 - 1000)` for 64-bit platforms.

### Example

```yaml tab="Structured (YAML)"
## Dynamic configuration
http:
  routers:
    Router-1:
      rule: "HostRegexp(`[a-z]+\\.traefik\\.com`)"
      entryPoints:
      - "web"
      service: service-1
      priority: 1
    Router-2:
      rule: "Host(`foobar.traefik.com`)"
      entryPoints:
      - "web"
      priority: 2
      service: service-2
```

```toml tab="Structured (TOML)"
## Dynamic configuration
[http.routers]
  [http.routers.Router-1]
    rule = "HostRegexp(`[a-z]+\\.traefik\\.com`)"
    entryPoints = ["web"]
    service = "service-1"
    priority = 1
  [http.routers.Router-2]
    rule = "Host(`foobar.traefik.com`)"
    entryPoints = ["web"]
    priority = 2
    service = "service-2"
```

```yaml tab="Labels"
labels:
  - "traefik.http.routers.Router-1.rule=HostRegexp(`[a-z]+\\.traefik\\.com`)"
  - "traefik.http.routers.Router-1.entryPoints=web"
  - "traefik.http.routers.Router-1.service=service-1"
  - "traefik.http.routers.Router-1.priority=1"
  - "traefik.http.routers.Router-2.rule=Host(`foobar.traefik.com`)"
  - "traefik.http.routers.Router-2.entryPoints=web"
  - "traefik.http.routers.Router-2.service=service-2"
  - "traefik.http.routers.Router-2.priority=2"
```

```json tab="Tags"
  {
    // ...
    "Tags": [
      "traefik.http.routers.Router-1.rule=HostRegexp(`[a-z]+\\.traefik\\.com`)",
      "traefik.http.routers.Router-1.entryPoints=web",
      "traefik.http.routers.Router-1.service=service-1",
      "traefik.http.routers.Router-1.priority=1"
      "traefik.http.routers.Router-2.rule=Host(`foobar.traefik.com`)",
      "traefik.http.routers.Router-2.entryPoints=web",
      "traefik.http.routers.Router-2.service=service-2",
      "traefik.http.routers.Router-2.priority=2"
    ]
  }
```

In the example above, the priority is configured to allow `Router-2` to handle requests with the `foobar.traefik.com` host.
