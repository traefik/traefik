---
title: "Traefik TCP Routers Rules & Priority Documentation"
description: "In Traefik Proxy, a router is in charge of connecting incoming requests to the Services that can handle them. Read the technical documentation."
---


## General 

!!! note
    - The character @ is not authorized in the router name
    - If HTTP routers and TCP routers listen to the same [entry point](../../../install-configuration/entrypoints.md), the order in which they apply depends on whether the connection uses TLS. See [TCP and HTTP Routers on the Same Entry Point](#tcp-and-http-routers-on-the-same-entry-point).

## Rules

Rules are a set of matchers configured with values, that determine if a particular connection matches specific criteria. If the rule is verified, the router becomes active, calls middlewares, and then forwards the request to the service.

The table below lists all the available matchers:

| Rule                                                        | Description                                                                                      |
|-------------------------------------------------------------|:-------------------------------------------------------------------------------------------------|
| <a id="opt-HostSNIdomain" href="#opt-HostSNIdomain" title="#opt-HostSNIdomain">[```HostSNI(`domain`)```](#hostsni-and-hostsniregexp)</a> | Checks if the connection's Server Name Indication is equal to `domain`. Supports wildcard subdomain matching (e.g. `*.example.com`).<br /> More information [here](#hostsni-and-hostsniregexp). |
| <a id="opt-HostSNIRegexpregexp" href="#opt-HostSNIRegexpregexp" title="#opt-HostSNIRegexpregexp">[```HostSNIRegexp(`regexp`)```](#hostsni-and-hostsniregexp)</a> | Checks if the connection's Server Name Indication matches `regexp`.<br />Use a [Go](https://golang.org/pkg/regexp/) flavored syntax.<br /> More information [here](#hostsni-and-hostsniregexp). |
| <a id="opt-ClientIPip" href="#opt-ClientIPip" title="#opt-ClientIPip">[```ClientIP(`ip`)```](#clientip)</a> | Checks if the connection's client IP correspond to `ip`. It accepts IPv4, IPv6 and CIDR formats.<br /> More information [here](#clientip). |
| <a id="opt-ALPNprotocol" href="#opt-ALPNprotocol" title="#opt-ALPNprotocol">[```ALPN(`protocol`)```](#alpn)</a> | Checks if the connection's ALPN protocol equals `protocol`.<br /> More information [here](#alpn).          |

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

`HostSNI` and `HostSNIRegexp` matchers allow matching connections targeted to a given domain.

These matchers do not support non-ASCII characters, use punycode encoded values ([rfc 3492](https://tools.ietf.org/html/rfc3492)) to match such domains.

!!! note "HostSNI & TLS"

    It is important to note that the Server Name Indication is an extension of the TLS protocol.
    Hence, only TLS routers will be able to specify a domain name with that rule.
    However, there is one special use case for `HostSNI` with non-TLS routers:
    when one wants a non-TLS router that matches all (non-TLS) requests,
    one should use the specific ```HostSNI(`*`)``` syntax.

!!! info "Wildcard subdomain matching"

    The `HostSNI` matcher supports a single-level wildcard prefix (`*.example.com`) to match any direct subdomain of `example.com`.
    It should be preferred over the `HostSNIRegexp` matcher as it allows attaching a TLS option and is more efficient.

    A wildcard matches exactly one subdomain label: `*.example.com` matches `foo.example.com` but not `foo.bar.example.com` or `example.com` itself.    

    This is only available with the **v3 rule syntax** (the default).

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

Match TCP connections opened on any direct subdomain of `example.com` (e.g. `foo.example.com`):

```yaml
HostSNI(`*.example.com`)
```

Match TCP connections opened on any subdomain of `example.com` (including nested subdomains), using a regular expression:

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
          service = "service-2"
      ```

    ```yaml tab="Labels"
       labels:
        - 'traefik.tcp.routers.Router-1.rule=ClientIP(`192.168.0.12`)'
        - "traefik.tcp.routers.Router-1.entryPoints=web"
        - "traefik.tcp.routers.Router-1.service=service-1"
        - "traefik.tcp.routers.Router-1.priority=2"
        - 'traefik.tcp.routers.Router-2.rule=ClientIP(`192.168.0.0/24`)'
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

Negative priority values are supported.

Traefik reserves a range of priorities for its internal routers, the maximum user-defined router priority value is:

- `(MaxInt32 - 1000)` for 32-bit platforms,
- `(MaxInt64 - 1000)` for 64-bit platforms.

!!! info "Providers Precedence"

    When two routes from **different providers** share the same numeric priority,
    Traefik uses the [`providers.precedence`](../../../install-configuration/providers/overview.md#providers-precedence) install configuration option to determine which route takes precedence.
    The provider listed first in `precedence` wins the tie.

## TCP and HTTP Routers on the Same Entry Point

When TCP routers and HTTP routers listen to the same entry point, Traefik sends each connection to either the TCP routers or the HTTP routers, before it reads an HTTP request.
The order depends on whether the connection starts with a TLS handshake.

For a non-TLS connection, the TCP routers without TLS apply first.
If none of them matches, the HTTP routers handle the connection.

TCP routers without TLS don't apply to TLS connections, unless the entry point has no HTTPS routers and no TCP routers with TLS.
In that case, the TCP routers without TLS also get TLS connections, and Traefik forwards them without decrypting them.

For a TLS connection, Traefik uses the first case that applies:

1. The SNI matches a `Host` matcher of an HTTPS router (an HTTP router with TLS), including a wildcard value such as ``Host(`*.example.com`)``. A negated `!Host` matcher doesn't count, and ``Host(`*`)`` belongs to case 3. The HTTPS routers handle the connection.
2. The highest-priority TCP router with TLS (including `tls.passthrough`) that matches the connection has a rule other than exactly ``HostSNI(`*`)``. The matching TCP router handles the connection. A rule with only a `ClientIP` matcher belongs to this case.
3. At least one HTTPS router on the entry point has no `Host` matcher (a negated `!Host` doesn't count) or has ``Host(`*`)``, for example a rule with only `PathPrefix` or `HostRegexp` matchers. The HTTPS routers handle the connection.
4. The highest-priority TCP router with TLS that matches the connection has the rule ``HostSNI(`*`)``. The ``HostSNI(`*`)`` router handles the connection.
5. In any other case, the HTTPS routers handle the connection.

A connection without SNI never matches case 1.
Traefik doesn't match `HostRegexp` against the SNI.
For example, an HTTPS router with the rule ``HostRegexp(`^.+\.example\.com$`) || Host(`example.com`)`` takes a connection at case 1 only when the SNI is `example.com`.

When the HTTPS routers handle a connection, they match each request on its `Host` header, path, and other HTTP matchers.
If no HTTPS router matches the request, Traefik returns a `404` and doesn't try the TCP routers.

The `priority` option orders TCP routers among themselves and HTTP routers among themselves, so a TCP router with `priority: 1000` still applies after an HTTPS router that matches in case 1.
A ``HostSNI(`*`)`` TCP router has the priority `-1` by default, so it comes after TCP routers with a default priority.
If you give a ``HostSNI(`*`)`` TCP router a higher `priority` than another TCP router with TLS, Traefik matches it first, and connections for the other router go to case 3 or case 4.
