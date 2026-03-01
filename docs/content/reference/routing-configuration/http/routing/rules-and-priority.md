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
| <a id="opt-Headerkey-value" href="#opt-Headerkey-value" title="#opt-Headerkey-value">[```Header(`key`, `value`)```](#header-and-headerregexp)</a> | Matches requests containing a header named `key` set to `value`.               |
| <a id="opt-HeaderRegexpkey-regexp" href="#opt-HeaderRegexpkey-regexp" title="#opt-HeaderRegexpkey-regexp">[```HeaderRegexp(`key`, `regexp`)```](#header-and-headerregexp)</a> | Matches requests containing a header named `key` matching `regexp`.            |
| <a id="opt-Hostdomain" href="#opt-Hostdomain" title="#opt-Hostdomain">[```Host(`domain`)```](#host-and-hostregexp)</a> | Matches requests host set to `domain`.                                         |
| <a id="opt-HostRegexpregexp" href="#opt-HostRegexpregexp" title="#opt-HostRegexpregexp">[```HostRegexp(`regexp`)```](#host-and-hostregexp)</a> | Matches requests host matching `regexp`.                                       |
| <a id="opt-Methodmethod" href="#opt-Methodmethod" title="#opt-Methodmethod">[```Method(`method`)```](#method)</a> | Matches requests method set to `method`.                                       |
| <a id="opt-Pathpath" href="#opt-Pathpath" title="#opt-Pathpath">[```Path(`path`)```](#path-pathprefix-and-pathregexp)</a> | Matches requests path set to `path`.                                           |
| <a id="opt-PathPrefixprefix" href="#opt-PathPrefixprefix" title="#opt-PathPrefixprefix">[```PathPrefix(`prefix`)```](#path-pathprefix-and-pathregexp)</a> | Matches requests path prefix set to `prefix`.                                  |
| <a id="opt-PathRegexpregexp" href="#opt-PathRegexpregexp" title="#opt-PathRegexpregexp">[```PathRegexp(`regexp`)```](#path-pathprefix-and-pathregexp)</a> | Matches request path using `regexp`.                                           |
| <a id="opt-Querykey-value" href="#opt-Querykey-value" title="#opt-Querykey-value">[```Query(`key`, `value`)```](#query-and-queryregexp)</a> | Matches requests query parameters named `key` set to `value`.                  |
| <a id="opt-QueryRegexpkey-regexp" href="#opt-QueryRegexpkey-regexp" title="#opt-QueryRegexpkey-regexp">[```QueryRegexp(`key`, `regexp`)```](#query-and-queryregexp)</a> | Matches requests query parameters named `key` matching `regexp`.               |
| <a id="opt-ClientIPip" href="#opt-ClientIPip" title="#opt-ClientIPip">[```ClientIP(`ip`)```](#clientip)</a> | Matches requests client IP using `ip`. It accepts IPv4, IPv6 and CIDR formats. |

### Header and HeaderRegexp

The `Header` and `HeaderRegexp` matchers allow matching requests that contain specific header.

| Behavior                                                        | Rule                                                                    |
|-----------------------------------------------------------------|:------------------------------------------------------------------------|
| <a id="opt-Match-requests-with-a-Content-Type-header-set-to-applicationyaml" href="#opt-Match-requests-with-a-Content-Type-header-set-to-applicationyaml" title="#opt-Match-requests-with-a-Content-Type-header-set-to-applicationyaml">Match requests with a `Content-Type` header set to `application/yaml`.</a> | ```Header(`Content-Type`, `application/yaml`)``` |
| <a id="opt-Match-requests-with-a-Content-Type-header-set-to-either-applicationjson-or-applicationyaml" href="#opt-Match-requests-with-a-Content-Type-header-set-to-either-applicationjson-or-applicationyaml" title="#opt-Match-requests-with-a-Content-Type-header-set-to-either-applicationjson-or-applicationyaml">Match requests with a `Content-Type` header set to either `application/json` or `application/yaml`.</a> | ```HeaderRegexp(`Content-Type`, `^application/(json\|yaml)$`)``` |
| <a id="opt-Match-headers-case-insensitively" href="#opt-Match-headers-case-insensitively" title="#opt-Match-headers-case-insensitively">Match headers [case-insensitively](https://en.wikipedia.org/wiki/Case_sensitivity).</a> | ```HeaderRegexp(`Content-Type`, `(?i)^application/(json\|yaml)$`)``` |

### Host and HostRegexp

The `Host` and `HostRegexp` matchers allow matching requests that are targeted to a given host.

These matchers do not support non-ASCII characters, use punycode encoded values ([rfc 3492](https://tools.ietf.org/html/rfc3492)) to match such domains.

If no `Host` is set in the request URL (for example, it's an IP address), these matchers will look at the `Host` header.

These matchers will match the request's host in lowercase.

| Behavior                                                        | Rule                                                                    |
|-----------------------------------------------------------------|:------------------------------------------------------------------------|
| <a id="opt-Match-requests-with-Host-set-to-example-com" href="#opt-Match-requests-with-Host-set-to-example-com" title="#opt-Match-requests-with-Host-set-to-example-com">Match requests with `Host` set to `example.com`.</a> | ```Host(`example.com`)``` |
| <a id="opt-Match-requests-sent-to-any-subdomain-of-example-com" href="#opt-Match-requests-sent-to-any-subdomain-of-example-com" title="#opt-Match-requests-sent-to-any-subdomain-of-example-com">Match requests sent to any subdomain of `example.com`.</a> | ```HostRegexp(`^.+\.example\.com$`)``` |
| <a id="opt-Match-requests-with-Host-set-to-either-example-com-or-example-org" href="#opt-Match-requests-with-Host-set-to-either-example-com-or-example-org" title="#opt-Match-requests-with-Host-set-to-either-example-com-or-example-org">Match requests with `Host` set to either `example.com` or `example.org`.</a> | ```HostRegexp(`^example\.(com\|org)$`)``` |
| <a id="opt-Match-Host-case-insensitively" href="#opt-Match-Host-case-insensitively" title="#opt-Match-Host-case-insensitively">Match `Host` [case-insensitively](https://en.wikipedia.org/wiki/Case_sensitivity).</a> | ```HostRegexp(`(?i)^example\.(com\|org)$`)``` |

### Method

The `Method` matchers allows matching requests sent based on their HTTP method (also known as request verb).

| Behavior                                                        | Rule                                                                    |
|-----------------------------------------------------------------|:------------------------------------------------------------------------|
| <a id="opt-Match-OPTIONS-requests" href="#opt-Match-OPTIONS-requests" title="#opt-Match-OPTIONS-requests">Match `OPTIONS` requests.</a> | ```Method(`OPTIONS`)``` |

### Path, PathPrefix, and PathRegexp

These matchers allow matching requests based on their URL path.

For exact matches, use `Path` and its prefixed alternative `PathPrefix`, for regexp matches, use `PathRegexp`.

Path are always starting with a `/`, except for `PathRegexp`.

| Behavior                                                        | Rule                                                                    |
|-----------------------------------------------------------------|:------------------------------------------------------------------------|
| <a id="opt-Match-products-but-neither-productsshoes-nor-products" href="#opt-Match-products-but-neither-productsshoes-nor-products" title="#opt-Match-products-but-neither-productsshoes-nor-products">Match `/products` but neither `/products/shoes` nor `/products/`.</a> | ```Path(`/products`)``` |
| <a id="opt-Match-products-as-well-as-everything-under-products-such-as-productsshoes-products-but-also-products-for-sale" href="#opt-Match-products-as-well-as-everything-under-products-such-as-productsshoes-products-but-also-products-for-sale" title="#opt-Match-products-as-well-as-everything-under-products-such-as-productsshoes-products-but-also-products-for-sale">Match `/products` as well as everything under `/products`, such as `/products/shoes`, `/products/` but also `/products-for-sale`.</a> | ```PathPrefix(`/products`)``` |
| <a id="opt-Match-both-productsshoes-and-productssocks-with-and-ID-like-productsshoes31" href="#opt-Match-both-productsshoes-and-productssocks-with-and-ID-like-productsshoes31" title="#opt-Match-both-productsshoes-and-productssocks-with-and-ID-like-productsshoes31">Match both `/products/shoes` and `/products/socks` with and ID like `/products/shoes/31`.</a> | ```PathRegexp(`^/products/(shoes\|socks)/[0-9]+$`)``` |
| <a id="opt-Match-requests-with-a-path-ending-in-either-jpeg-jpg-or-png" href="#opt-Match-requests-with-a-path-ending-in-either-jpeg-jpg-or-png" title="#opt-Match-requests-with-a-path-ending-in-either-jpeg-jpg-or-png">Match requests with a path ending in either `.jpeg`, `.jpg` or `.png`.</a> | ```PathRegexp(`\.(jpeg\|jpg\|png)$`)``` |
| <a id="opt-Match-products-as-well-as-everything-under-products-such-as-productsshoes-products-but-also-products-for-sale-case-insensitively" href="#opt-Match-products-as-well-as-everything-under-products-such-as-productsshoes-products-but-also-products-for-sale-case-insensitively" title="#opt-Match-products-as-well-as-everything-under-products-such-as-productsshoes-products-but-also-products-for-sale-case-insensitively">Match `/products` as well as everything under `/products`, such as `/products/shoes`, `/products/` but also `/products-for-sale`, [case-insensitively](https://en.wikipedia.org/wiki/Case_sensitivity).</a> | ```PathRegexp(`(?i)^/products`)``` |

### Query and QueryRegexp

The `Query` and `QueryRegexp` matchers allow matching requests based on query parameters.

| Behavior                                                        | Rule                                                                    |
|-----------------------------------------------------------------|:------------------------------------------------------------------------|
| <a id="opt-Match-requests-with-a-mobile-query-parameter-set-to-true-such-as-in-searchmobiletrue" href="#opt-Match-requests-with-a-mobile-query-parameter-set-to-true-such-as-in-searchmobiletrue" title="#opt-Match-requests-with-a-mobile-query-parameter-set-to-true-such-as-in-searchmobiletrue">Match requests with a `mobile` query parameter set to `true`, such as in `/search?mobile=true`.</a> | ```Query(`mobile`, `true`)``` |
| <a id="opt-Match-requests-with-a-query-parameter-mobile-that-has-no-value-such-as-in-searchmobile" href="#opt-Match-requests-with-a-query-parameter-mobile-that-has-no-value-such-as-in-searchmobile" title="#opt-Match-requests-with-a-query-parameter-mobile-that-has-no-value-such-as-in-searchmobile">Match requests with a query parameter `mobile` that has no value, such as in `/search?mobile`.</a> | ```Query(`mobile`)``` |
| <a id="opt-Match-requests-with-a-mobile-query-parameter-set-to-either-true-or-yes" href="#opt-Match-requests-with-a-mobile-query-parameter-set-to-either-true-or-yes" title="#opt-Match-requests-with-a-mobile-query-parameter-set-to-either-true-or-yes">Match requests with a `mobile` query parameter set to either `true` or `yes`.</a> | ```QueryRegexp(`mobile`, `^(true\|yes)$`)``` |
| <a id="opt-Match-requests-with-a-mobile-query-parameter-set-to-any-value-including-the-empty-value" href="#opt-Match-requests-with-a-mobile-query-parameter-set-to-any-value-including-the-empty-value" title="#opt-Match-requests-with-a-mobile-query-parameter-set-to-any-value-including-the-empty-value">Match requests with a `mobile` query parameter set to any value (including the empty value).</a> | ```QueryRegexp(`mobile`, `^.*$`)``` |
| <a id="opt-Match-query-parameters-case-insensitively" href="#opt-Match-query-parameters-case-insensitively" title="#opt-Match-query-parameters-case-insensitively">Match query parameters [case-insensitively](https://en.wikipedia.org/wiki/Case_sensitivity).</a> | ```QueryRegexp(`mobile`, `(?i)^(true\|yes)$`)``` |

### ClientIP

The `ClientIP` matcher allows matching requests sent from the given client IP.

It only matches the request client IP and does not use the `X-Forwarded-For` header for matching.

| Behavior                                                        | Rule                                                                    |
|-----------------------------------------------------------------|:------------------------------------------------------------------------|
| <a id="opt-Match-requests-coming-from-a-given-IP-IPv4" href="#opt-Match-requests-coming-from-a-given-IP-IPv4" title="#opt-Match-requests-coming-from-a-given-IP-IPv4">Match requests coming from a given IP (IPv4).</a> | ```ClientIP(`10.76.105.11`)``` |
| <a id="opt-Match-requests-coming-from-a-given-IP-IPv6" href="#opt-Match-requests-coming-from-a-given-IP-IPv6" title="#opt-Match-requests-coming-from-a-given-IP-IPv6">Match requests coming from a given IP (IPv6).</a> | ```ClientIP(`::1`)``` |
| <a id="opt-Match-requests-coming-from-a-given-subnet-IPv4" href="#opt-Match-requests-coming-from-a-given-subnet-IPv4" title="#opt-Match-requests-coming-from-a-given-subnet-IPv4">Match requests coming from a given subnet (IPv4).</a> | ```ClientIP(`192.168.1.0/24`)``` |
| <a id="opt-Match-requests-coming-from-a-given-subnet-IPv6" href="#opt-Match-requests-coming-from-a-given-subnet-IPv6" title="#opt-Match-requests-coming-from-a-given-subnet-IPv6">Match requests coming from a given subnet (IPv6).</a> | ```ClientIP(`fe80::/10`)``` |

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
    | <a id="opt-Router-1" href="#opt-Router-1" title="#opt-Router-1">Router-1</a> | ```HostRegexp(`[a-z]+\.traefik\.com`)``` | 34       |
    | <a id="opt-Router-2" href="#opt-Router-2" title="#opt-Router-2">Router-2</a> | ```Host(`foobar.traefik.com`)```         | 26       |

    The previous table shows that `Router-1` has a higher priority than `Router-2`.

    To solve this issue, the priority must be set.

To avoid path overlap, routes are sorted, by default, in descending order using rules length.
The priority is directly equal to the length of the rule, and so the longest length has the highest priority.

A value of `0` for the priority is ignored: `priority: 0` means that the default rules length sorting is used.

Negative priority values are supported.

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
