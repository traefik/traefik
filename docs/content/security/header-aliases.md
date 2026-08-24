---
title: "Headers with Aliasing Names"
description: "Learn how Traefik handles the request headers whose name aliases another header name, to prevent header spoofing against backends that normalize header names. Read the technical documentation."
---

# Headers with Aliasing Names

Preventing Header Spoofing Through Aliasing Header Names
{: .subtitle }

Go canonicalizes header names on dashes only, therefore it handles `X-Auth-User`, `X_Auth_User` and `X.Auth.User`
as three distinct headers.

Many backends derive their variable names from the header names (CGI, WSGI, PHP, NGINX, ...): they uppercase the header
name, and replace every character that is neither a letter nor a digit with an underscore.
For them, the three names above are the same `HTTP_X_AUTH_USER` variable.

Underscores and dots are not the only characters concerned. On top of the letters, the digits and the dash,
HTTP allows the following characters in a header name:

```text
!  #  $  %  &  '  *  +  .  ^  _  `  |  ~
```

Each of them is turned into an underscore by such backends, hence each of them builds an alias:
`X_Auth_User`, `X.Auth.User`, `X!Auth!User`, `X~Auth~User` and `X+Auth+User` are all read as `X-Auth-User`.
Only the letters, the digits and the dash are left untouched, so only the dash form is unambiguous.

As a consequence, a client can smuggle an alias of a header managed by Traefik past the middleware managing it,
and have such a backend read the spoofed value instead of the value asserted by Traefik.

## Configuration

The [`aliasHeadersStrategy`](../routing/entrypoints.md#aliasheadersstrategy) entry point option is the mitigation
against this class of spoofing. It controls how the request headers whose name aliases another header name are
handled before routing:

- `keep` (default): request headers with an aliasing name are forwarded as is.
- `delete`: any request header whose name contains a character which is neither a letter, a digit, nor a dash is silently removed from the request.
- `reject`: any request carrying a header whose name contains a character which is neither a letter, a digit, nor a dash is rejected with a `400 Bad Request` response.

!!! warning "Legitimate headers are dropped too"

    The `delete` and `reject` strategies apply to every request header, not only to the ones Traefik manages:
    they cannot tell a spoofing attempt from a legitimate header whose name happens to contain such a character.
    An application relying on a header named for example `X_Request_Id` stops receiving it with `delete`,
    and its requests are rejected with `reject`.

    In practice, header names are built with letters, digits and dashes only, and no header of the IANA registry
    contains any other character. A backend deliberately using such names (which a default NGINX configuration
    already drops for underscores) is the case to check before enabling this option.

!!! warning "Security Considerations"

    Setting `aliasHeadersStrategy` to `delete` or `reject` is strongly recommended when an entry point fronts a backend
    normalizing the header names. The middlewares managing request headers (ForwardAuth `authResponseHeaders`,
    BasicAuth and DigestAuth `headerField`, PassTLSClientCert, Headers `customRequestHeaders`, StripPrefix,
    ReplacePath, ...) only manage the canonical form of the headers they set, and rely on this option to not be
    spoofed with an aliasing name.

    Traefik logs a warning at startup for every entry point left without this option configured.

```yaml tab="File (YAML)"
entryPoints:
  websecure:
    address: ":443"
    http:
      aliasHeadersStrategy: delete  # Default: keep
```

```toml tab="File (TOML)"
[entryPoints.websecure]
  address = ":443"

  [entryPoints.websecure.http]
    aliasHeadersStrategy = "delete"
```

```bash tab="CLI"
--entryPoints.websecure.address=:443
--entryPoints.websecure.http.aliasHeadersStrategy=delete
```

!!! info "Deprecated option"

    The `underscoreHeadersStrategy` option is deprecated in favor of `aliasHeadersStrategy`, which handles every
    aliasing character instead of the underscore only.
    See the [feature deprecation notices](../deprecation/features.md#entry-point-underscoreheadersstrategy-option)
    for more details.
