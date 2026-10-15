---
title: "Traefik AccessGovernance Middleware"
description: "Persist bounded, asynchronous HTTP, TCP, and UDP access governance events to MySQL."
---

# AccessGovernance

The `accessGovernance` middleware records authentication, authorization, audit,
and metering dimensions without putting MySQL writes on the proxy request path.
It does not make authorization decisions itself; place it before the
authentication and authorization middleware whose outcome it must observe.

## HTTP Configuration Example

```yaml tab="Structured (YAML)"
http:
  middlewares:
    governance:
      accessGovernance:
        runtimeValues:
          primaryDSN:
            file: /run/secrets/audit-primary-dsn
          sensitiveDSN:
            file: /run/secrets/audit-sensitive-dsn
        defaultDatabase: default
        databases:
          default:
            dsn: runtime://primaryDSN
            httpTable: traefik_http_access_events
            networkTable: traefik_network_access_events
            maxOpenConnections: 8
            maxIdleConnections: 4
            connectionMaxAge: 30m
          sensitive:
            dsn: runtime://sensitiveDSN
            httpTable: traefik_sensitive_http_access_events
        routes:
          - name: sensitive-api
            database: sensitive
            protocols: [http]
            methods: [POST]
            pathRegex: ^/sensitive(?:/|$)
            bodyJSONPaths:
              $.operation: ^delete$
        identity:
          - kind: header
            name: X-Authenticated-User
          - kind: cookie
            name: account
            decode: url
          - kind: json
            jsonPath: $.operator.name
        capture:
          bodyReadMode: eager
          maxBodyBytes: 65536
          storeQuery: false
          storeUserAgent: true
        queue:
          capacity: 8192
          workers: 2
          batchSize: 200
          flushInterval: 1s
          enqueueWait: 0s
        dimensions:
          tenantID: tenant-a
          authnMethod: forward-auth
          policyID: policy-a
          billingAccount: account-a
          meter: http_request
          labels:
            environment: production
        schema:
          autoCreate: true
```

The first matching `routes` entry selects a database. An unmatched event uses
`defaultDatabase`; omitting the default intentionally discards unmatched
events. A route can match protocol, entry point, router, service, method, host,
path, headers, and JSON Body fields. `bodyJSONPaths` maps each JSONPath to a
regular expression; all configured expressions must match. HTTP protocol
matching includes HTTP/1, HTTP/2, and HTTP/3 when `protocols` contains `http`.

`databases` is a named map and has no fixed destination limit. Each destination
has an independent DSN, table names, connection pool, and schema bootstrap.
Put more specific routes first to send different requests, including requests
to the same path with different JSON Body values, to different databases.

## Identity Extraction

Identity sources are evaluated in order. Supported `kind` values are `header`,
`cookie`, `query`, `basic`, `clientCert`, `json`, and `form`. A forwarded
`Cookie` request header is parsed by the standard HTTP cookie parser, so
cookie-based gateway authentication remains compatible.

`regex` and `group` can extract part of a value. `decode` accepts `url`,
`base64`, or `base64url`.

Extraction does not authenticate a principal. Client-supplied headers, cookies,
query parameters, and body fields are untrusted unless verified by a trusted
authentication layer. Choose the identity source and middleware order accordingly;
do not treat a recorded principal as proof of authentication.

`bodyReadMode: eager` reads at most `maxBodyBytes + 1` before forwarding and
then restores the exact request stream. It guarantees JSON/form extraction
even when a backend does not consume the body. `lazy` avoids this initial read,
but body identity is available only for bytes consumed downstream. Body content
is never stored in the event table.

When a request satisfies the non-Body selectors of a database route that
defines `bodyJSONPaths`, the middleware automatically uses eager bounded
capture for deterministic routing, regardless of `bodyReadMode`. Other paths
do not pay that eager-read cost. Requests that exceed `maxBodyBytes` cannot
match a Body route and therefore use the next matching route or
`defaultDatabase`.

## Queue and Failure Behavior

Each distinct middleware configuration owns a bounded, shared worker pool.
Workers batch events by destination and use bounded MySQL connection pools.
`enqueueWait: 0s` makes a full queue drop the audit event instead of delaying
the proxied request. Drop warnings are rate-limited to avoid a secondary log
storm.

Database or schema errors never fail a proxied request. Failed batches are
reported and dropped; the database connection is retried on later batches.

## Schema

With `schema.autoCreate: true`, the middleware creates both tables and adds
missing columns and named indexes. It does not change existing column types.
Use a migration-capable account for bootstrap and a write-only account
afterward.

The HTTP table includes route and service identity, principal, authentication
and authorization results, policy, tenant, billing account, meter, billable
units, request/response sizes, timing, trace identifiers, labels, and extension
metadata. The network table stores the corresponding TCP/UDP session fields.

For Kubernetes CRDs, declare arbitrary `runtimeValues` and set `dsn` to an
exact `runtime://alias` reference. The provider resolves the selected
Secret/ConfigMap value into a non-serialized field; the dynamic API and debug
JSON retain only source metadata and `resourceVersion`.

For example, a Kubernetes `Middleware` can reference a Secret in its own namespace:

```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: audit
spec:
  accessGovernance:
    runtimeValues:
      primaryDSN:
        secretKeyRef:
          name: audit-database
          key: dsn
    defaultDatabase: primary
    databases:
      primary:
        dsn: runtime://primaryDSN
        httpTable: http_access_events
        networkTable: network_access_events
```

Secret/ConfigMap changes rebuild the dynamic middleware without restarting the
Traefik Pod. A process-level connection manager opens and pings a candidate,
applies schema bootstrap, atomically switches successful pools, and closes the
old pool only after in-flight operations drain. A failed candidate leaves the
old pool active and is retried. `dsnFile` remains supported; its contents are
checked by the resident connection manager.

## Apply to Every Router

An entry-point default applies the middleware to all HTTP routers on that entry
point:

```yaml
entryPoints:
  websecure:
    address: ":443"
    http:
      middlewares:
        - governance@file
```

## Runtime File and SQL Safety

Runtime values and credential files are operator-controlled configuration, never
paths selected from an incoming request. Each file must be a regular file no
larger than 1 MiB. Relative symlinks within the configured file's parent directory
are supported, including Kubernetes projected Secret keys. Absolute symlinks,
links outside that directory, directories, devices, sockets, and named pipes are
rejected. Mount the required files into a dedicated directory with restrictive
permissions; this boundary does not restrict which directory an operator can
configure or isolate files in the same mount.

Projected Secret reads pin the `..data` revision while reading related files.
Subsequent reads observe atomic Secret updates without restarting Traefik.
A rejected rotation keeps the last successfully activated MySQL pool or
credential set; initial activation without valid material fails.

SQL table names must match `^[A-Za-z][A-Za-z0-9_]{0,63}$`.
The SQL execution paths validate and quote identifiers separately from values.
Request-derived values use SQL parameters and never become table names.

MySQL connections enforce server-side parameter binding and disable multi-statement
execution and unrestricted local-file uploads. DSN options `interpolateParams`,
`multiStatements`, and `allowAllFiles` are forced to `false`, including after Secret
rotation. Other connection, TLS, and pool settings continue to apply.
