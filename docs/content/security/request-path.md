---
title: "Request Path"
description: "Learn how Traefik processes and secures request paths through sanitization and encoded character filtering to protect against path traversal and injection attacks."
---

# Request Path

Protecting Against Path-Based Attacks Through Sanitization and Filtering
{: .subtitle }

Traefik implements multiple layers of security when processing incoming request paths.
This includes path sanitization to normalize potentially dangerous sequences and encoded character filtering to prevent attack vectors that use URL encoding.
Understanding how Traefik handles request paths is crucial for maintaining a secure routing infrastructure.

## How Traefik Processes Request Paths

When Traefik receives an HTTP request, it processes the request path through several security-focused stages:

### 1. Encoded Character Filtering

Traefik inspects the path for potentially dangerous encoded characters and rejects requests containing them unless explicitly allowed.

Here is the list of the encoded characters that are allowed by default:

| Encoded Character | Character               |
|-------------------|-------------------------|
| `%2f` or `%2F`    | `/` (slash)             |
| `%5c` or `%5C`    | `\` (backslash)         |
| `%00`             | `NULL` (null character) |
| `%3b` or `%3B`    | `;` (semicolon)         |
| `%25`             | `%` (percent)           |
| `%3f` or `%3F`    | `?` (question mark)     |
| `%23`             | `#` (hash)              |

### 2. Path Normalization

Traefik normalizes the request path by decoding the unreserved percent-encoded characters,
as they are equivalent to their non-encoded form (according to [rfc3986#section-2.3](https://datatracker.ietf.org/doc/html/rfc3986#section-2.3)),
and capitalizing the percent-encoded characters (according to [rfc3986#section-6.2.2.1](https://datatracker.ietf.org/doc/html/rfc3986#section-6.2.2.1)).

### 3. Path Sanitization

Traefik sanitizes request paths to prevent common attack vectors,
by removing the `..`, `.` and duplicate slash segments from the URL (according to [rfc3986#section-6.2.2.3](https://datatracker.ietf.org/doc/html/rfc3986#section-6.2.2.3)).

## Path Security Configuration

Traefik provides two main mechanisms for path security that work together to protect your applications.

### Path Sanitization

Path sanitization is enabled by default and helps prevent directory traversal attacks by normalizing request paths.
Configure it in the [EntryPoints](../routing/entrypoints.md#sanitizepath) HTTP section:

```yaml tab="File (YAML)"
entryPoints:
  websecure:
    address: ":443"
    http:
      sanitizePath: true  # Default: true (recommended)
```

```toml tab="File (TOML)"
[entryPoints.websecure]
  address = ":443"

  [entryPoints.websecure.http]
    sanitizePath = true
```

```bash tab="CLI"
--entryPoints.websecure.address=:443
--entryPoints.websecure.http.sanitizePath=true
```

**Sanitization behavior:**

- `./foo/bar` → `/foo/bar` (removes relative current directory)
- `/foo/../bar` → `/bar` (resolves parent directory traversal)
- `/foo/bar//` → `/foo/bar/` (removes duplicate slashes)
- `/./foo/../bar//` → `/bar/` (combines all normalizations)

### Encoded Character Filtering

Encoded character filtering provides an additional security layer by rejecting potentially dangerous URL-encoded characters.
Configure it in the [EntryPoints](../routing/entrypoints.md#encoded-characters) HTTP section.

This filtering occurs before path sanitization and catches attack attempts that use encoding to bypass other security controls.

All encoded character filtering is disabled by default (`true` means encoded characters are allowed).

!!! info "Security Considerations"

    When your backend is not fully compliant with [RFC 3986](https://datatracker.ietf.org/doc/html/rfc3986) and notably decode encoded reserved characters in the requets path,
    it is recommended to set these options to `false` to avoid split-view situation and helps prevent path traversal attacks or other malicious attempts to bypass security controls.

```yaml tab="File (YAML)"
entryPoints:
  websecure:
    address: ":443"
    http:
      encodedCharacters:
        allowEncodedSlash: false          # %2F - Default: true
        allowEncodedBackSlash: false      # %5C - Default: true
        allowEncodedNullCharacter: false  # %00 - Default: true
        allowEncodedSemicolon: false      # %3B - Default: true
        allowEncodedPercent: false        # %25 - Default: true
        allowEncodedQuestionMark: false   # %3F - Default: true
        allowEncodedHash: false           # %23 - Default: true
```

```toml tab="File (TOML)"
[entryPoints.websecure]
  address = ":443"

  [entryPoints.websecure.http.encodedCharacters]
    allowEncodedSlash = false
    allowEncodedBackSlash = false
    allowEncodedNullCharacter = false
    allowEncodedSemicolon = false
    allowEncodedPercent = false
    allowEncodedQuestionMark = false
    allowEncodedHash = false
```

```bash tab="CLI"
--entryPoints.websecure.address=:443
--entryPoints.websecure.http.encodedCharacters.allowEncodedSlash=false
--entryPoints.websecure.http.encodedCharacters.allowEncodedBackSlash=false
--entryPoints.websecure.http.encodedCharacters.allowEncodedNullCharacter=false
--entryPoints.websecure.http.encodedCharacters.allowEncodedSemicolon=false
--entryPoints.websecure.http.encodedCharacters.allowEncodedPercent=false
--entryPoints.websecure.http.encodedCharacters.allowEncodedQuestionMark=false
--entryPoints.websecure.http.encodedCharacters.allowEncodedHash=false
```
