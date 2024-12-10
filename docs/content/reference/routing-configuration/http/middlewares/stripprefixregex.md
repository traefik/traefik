---
title: "Traefik StripPrefixRegex Documentation"
description: "In Traefik Proxy's HTTP middleware, StripPrefixRegex removes prefixes from paths before forwarding requests, using regex. Read the technical documentation."
---

The `stripPrefixRegex` middleware strips the matching path prefix and stores it in a `X-Forwarded-Prefix` header.

!!! tip

    Use a `stripPrefixRegex` middleware if your backend listens on the root path (`/`) but should be exposed on a specific prefix.

## Configuration Example

## Configuration Options

| Field                        | Description                                                                                                                                                                                                | Default | Required |
|:-----------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| `regex` | List of regular expressions to match the path prefix from the request URL.<br /> For instance, `/products` also matches `/products/shoes` and `/products/shirts`.<br />oOre information [here](#regex). | | No |

### regex

If your backend is serving assets (for example, images or JavaScript files), it can use the `X-Forwarded-Prefix` header to construct relative URLs.
Using the previous example, the backend should return `/products/shoes/image.png` (and not `/images.png`, which Traefik would likely not be able to associate with the same backend).

!!! tip

    Regular expressions and replacements can be tested using online tools such as [Go Playground](https://play.golang.org/p/mWU9p-wk2ru) or the [Regex101](https://regex101.com/r/58sIgx/2).

    When defining a regular expression within YAML, any escaped character needs to be escaped twice: `example\.com` needs to be written as `example\\.com`.
