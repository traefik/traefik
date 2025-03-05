---
title: "Traefik TCP Routers Rules & Priority Documentation"
description: "In Traefik Proxy, a router is in charge of connecting incoming requests to the Services that can handle them. Read the technical documentation."
---


## General 

!!! note
    - The character @ is not authorized in the router name
    - If both HTTP routers and TCP routers listen to the same [EntryPoint](../../../install-configuration/entrypoints.md), the TCP routers will apply before the HTTP routers. If no matching route is found for the TCP routers, then the HTTP routers will take over.

## Rules

Rules are a set of matchers configured with values, that determine if a particular connection matches specific criteria. If the rule is verified, the router becomes active, calls middlewares, and then forwards the request to the service.

The table below lists all the available matchers:

| Rule                                                        | Description                                                                                      |
|-------------------------------------------------------------|:-------------------------------------------------------------------------------------------------|
| [```HostSNI(`domain`)```](#hostsni-and-hostsniregexp)       | Checks if the connection's Server Name Indication is equal to `domain`.<br /> More information [here](#hostsni-and-hostsniregexp).                          |
| [```HostSNIRegexp(`regexp`)```](#hostsni-and-hostsniregexp) | Checks if the connection's Server Name Indication matches `regexp`.<br />Use a [Go](https://golang.org/pkg/regexp/) flavored syntax.<br /> More information [here](#hostsni-and-hostsniregexp). |
| [```ClientIP(`ip`)```](#clientip)                           | Checks if the connection's client IP correspond to `ip`. It accepts IPv4, IPv6 and CIDR formats.<br /> More information [here](#clientip). |
| [```ALPN(`protocol`)```](#alpn)                             | Checks if the connection's ALPN protocol equals `protocol`.<br /> More information [here](#alpn).          |

!!! tip "Backticks or Quotes?"

    To set the value of a rule, use [backticks](https://en.wiktionary.org/wiki/backtick) ``` ` ``` or escaped double-quotes `\"`.

    Single quotes `'` are not accepted since the values are [Go's String Literals](https://golang.org/ref/spec#String_literals).

### Expressing Complex Rules Using Operators and Parenthesis

The usual AND (`&&`) and OR (`||`) logical operators can be used, with the expected precedence rules,
as well as parentheses.

One can invert a matcher by using the NOT (`!`) operator.

The following rule matches connections where:

- Either Server Name Indication is `example.com` OR,
- Server Name Indication is `example.org` AND ALPN protocol is NOT `h2`

```yaml
HostSNI(`example.com`) || (HostSNI(`example.org`) && !ALPN(`h2`))
```

### HostSNI and HostSNIRegexp

`HostSNI` and `HostSNIRegexp` matchers allow to match connections targeted to a given domain.

These matchers do not support non-ASCII characters, use punycode encoded values ([rfc 3492](https://tools.ietf.org/html/rfc3492)) to match such domains.

!!! note "HostSNI & TLS"

    It is important to note that the Server Name Indication is an extension of the TLS protocol.
    Hence, only TLS routers will be able to specify a domain name with that rule.
    However, there is one special use case for `HostSNI` with non-TLS routers:
    when one wants a non-TLS router that matches all (non-TLS) requests,
    one should use the specific ```HostSNI(`*`)``` syntax.

#### Examples

Match all connections:

```yaml tab="HostSNI"
HostSNI(`*`)
```

```yaml tab="HostSNIRegexp"
HostSNIRegexp(`^.*$`)
```

Match TCP connections sent to `example.com`:

```yaml
HostSNI(`example.com`)
```

Match TCP connections opened on any subdomain of `example.com`:

```yaml
HostSNIRegexp(`^.+\.example\.com$`)
```

### ClientIP

The `ClientIP` matcher allows matching connections opened by a client with the given IP.

#### Examples

Match connections opened by a given IP:

```yaml tab="IPv4"
ClientIP(`10.76.105.11`)
```

```yaml tab="IPv6"
ClientIP(`::1`)
```

Match connections coming from a given subnet:

```yaml tab="IPv4"
ClientIP(`192.168.1.0/24`)
```

```yaml tab="IPv6"
ClientIP(`fe80::/10`)
```

### ALPN

The `ALPN` matcher allows matching connections the given protocol.

It would be a security issue to let a user-defined router catch the response to
an ACME TLS challenge previously initiated by Traefik.
For this reason, the `ALPN` matcher is not allowed to match the `ACME-TLS/1`
protocol, and Traefik returns an error if this is attempted.

#### Example

Match connections using the ALPN protocol `h2`:

```yaml
ALPN(`h2`)
```

## Priority Calculation

???+ info "How default priorities are computed"

    ```yaml tab="Structured (YAML)"
      tcp:
        routers:
          Router-1:
            rule: "ClientIP(`192.168.0.12`)"
            entryPoints:
            - "web"
            service: service-1
            priority: 2
          Router-2:
            rule: "ClientIP(`192.168.0.0/24`)"
            entryPoints:
            - "web"
            priority: 1
            service: service-2
    ```

      ```toml tab="Structured (TOML)"
      [tcp.routers]
        [tcp.routers.Router-1]
          rule = "ClientIP(`192.168.0.12`)"
          entryPoints = ["web"]
          service = "service-1"
          priority = 2
        [tcp.routers.Router-2]
          rule = "ClientIP(`192.168.0.0/24`)"
          entryPoints = ["web"]
          priority = 1
          service = "service-2
      ```

    ```yaml tab="Labels"
       labels:
        - "traefik.tcp.routers.Router-1.rule="ClientIP(`192.168.0.12`)"
        - "traefik.tcp.routers.Router-1.entryPoints=web"
        - "traefik.tcp.routers.Router-1.service=service-1"
        - "traefik.tcp.routers.Router-1.priority=2"
        - "traefik.tcp.routers.Router-2.rule="ClientIP(`192.168.0.0/24`)"
        - "traefik.tcp.routers.Router-2.entryPoints=web"
        - "traefik.tcp.routers.Router-2.service=service-2"
        - "traefik.tcp.routers.Router-2.priority=1"
    ```

    ```json tab="Tags"
      {
        //...
        "Tags": [
          "traefik.tcp.routers.Router-1.rule=ClientIP(`192.168.0.12`)",
          "traefik.tcp.routers.Router-1.entryPoints=web",
          "traefik.tcp.routers.Router-1.service=service-1",
          "traefik.tcp.routers.Router-1.priority=2",
          "traefik.tcp.routers.Router-2.rule=ClientIP(`192.168.0.0/24`)",
          "traefik.tcp.routers.Router-2.entryPoints=web",
          "traefik.tcp.routers.Router-2.service=service-2",
          "traefik.tcp.routers.Router-2.priority=1"
        ]
      }
    ```

    In the example above, the priority is configured so that `Router-1` will handle requests from `192.168.0.12`.

To avoid path overlap, routes are sorted, by default, in descending order using rules length.
The priority is directly equal to the length of the rule, and so the longest length has the highest priority.
A value of `0` for the priority is ignored: `priority: 0` means that the default rules length sorting is used.

Traefik reserves a range of priorities for its internal routers, the maximum user-defined router priority value is:

- `(MaxInt32 - 1000)` for 32-bit platforms,
- `(MaxInt64 - 1000)` for 64-bit platforms.
