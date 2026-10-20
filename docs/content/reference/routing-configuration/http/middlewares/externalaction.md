---
title: "Traefik ExternalAction Middleware"
description: "Map source requests and responses to authenticated HTTP or executable actions."
---

# ExternalAction

The `externalAction` middleware triggers a generic external action for matching
HTTP requests. A rule can invoke an HTTP endpoint or a packaged executable
without a shell. An SDK integration should be compiled into a small adapter
binary and configured as an `exec` target.

## Configuration Example

The File provider renders Go templates when it loads routing configuration.
The example below escapes the ExternalAction expressions so they are evaluated
for each request instead of when the file is loaded.

```yaml tab="Structured (YAML)"
http:
  middlewares:
    action:
      externalAction:
        runtimeValues:
          destinationToken:
            file: /run/secrets/action-token
          destinationStaff:
            file: /run/secrets/action-caller
          destinationName:
            file: /run/secrets/action-project-name
          destinationID:
            file: /run/secrets/action-project-id
          actionDatabase:
            file: /run/secrets/action-state-dsn
        selectionMode: first
        credentialSets:
          deploymentDestination:
            reloadInterval: 5s
            immutableParameters: [projectid, projectname]
            auth:
              type: headers
              parameters:
                token: runtime://destinationToken
                staffname: runtime://destinationStaff
                projectname: runtime://destinationName
                projectid: runtime://destinationID
        queue:
          capacity: 2048
          workers: 8
          maxQueuedBytes: 67108864
          enqueueWait: 0s
        rules:
          - name: report-change
            match:
              methods: [POST]
              pathRegex: ^/deployments$
              headers:
                X-Action: ^report$
              bodyJSONPaths:
                $.operation: ^deploy$
              responseStatuses: ["200-299"]
              responseBodyJSONPaths:
                $.success: ^true$
            phase: after
            mode: async
            failurePolicy: ignore
            timeout: 5s
            body:
              readMode: eager
              maxBytes: 2097152
              overflow: skip
              format: json
              contentType: application/json
              template: |
                {
                  "id": {{`{{ jsonPath .Body "$.deployment.id" | toJson }}`}},
                  "resultId": {{`{{ jsonPath .Response.Body "$.data.id" | toJson }}`}},
                  "operator": {{`{{ coalesce (header . "X-Authenticated-User") (cookie . "account") | toJson }}`}},
                  "sourceStatus": {{`{{ .Response.Status }}`}}
                }
            responseBody:
              readMode: capture
              maxBytes: 1048576
              overflow: skip
              format: json
            deduplication:
              keyTemplate: 'deployment:{{`{{ jsonPath .Response.Body "$.data.id" }}`}}'
              ttl: 720h
              lease: 1m
              mysql:
                dsn: runtime://actionDatabase
                table: traefik_external_action_dedup
                autoCreate: true
            binding:
              mode: create
              keyTemplate: 'deployment:{{`{{ jsonPath .Response.Body "$.data.id" }}`}}'
              credentialSet: deploymentDestination
              ttl: 4320h
              mysql:
                dsn: runtime://actionDatabase
                table: traefik_external_action_binding
                autoCreate: true
            http:
              url: https://actions.example.com/v1/report
              method: POST
              headers:
                X-Source-Host: '{{`{{ .Request.Host }}`}}'
              tls:
                caFile: /run/secrets/action-ca.pem
              maxResponseBytes: 65536
              responseValidation:
                statuses: ["200-299"]
                successJSONPath: $.code
                successValues: ["0"]
                messageJSONPath: $.msg
```

## Match and Dispatch

The remaining fragments show request-time templates directly, as used in a
Kubernetes Middleware resource. When copying them into a File provider
configuration, escape each expression as in the full example above. Helm
templates introduce their own rendering step and also need to preserve the
request-time expressions.

Rules can match methods, hosts, a path regular expression, header regular
expressions, query regular expressions, and JSON Body fields. `bodyJSONPaths`
maps each JSONPath to a regular expression; all configured expressions must
match. An `after` rule can also restrict backend response status ranges and
response JSON fields through `responseBodyJSONPaths`.

Every rule owns its target URL, authentication, Body transformation, and
response validation, so one middleware can route different change tasks to
different external systems or change-management projects. `selectionMode` is `all` by
default and dispatches every matching rule. Set it to `first` for exclusive
ordered routing: put task/project-specific rules first and an optional fallback
rule last.

`phase` is `before` or `after`. `mode` is `sync` or `async`. A synchronous
`before` action with `failurePolicy: fail` can reject the source request with
502. An `after` action cannot replace a response that has already been sent.

Asynchronous rules use a bounded shared queue. A full queue drops the action
without blocking when `enqueueWait` is zero, and drop logging is rate-limited.
`maxQueuedBytes` bounds the estimated request/response Body and header memory
retained by queued and executing actions; its default is 64 MiB.

## Generic Mapping Model

Templates use Go `text/template` and Sprig functions, except environment
expansion functions. The template data is:

- `.Request`: method, scheme, host, path, raw query, headers, query values, and cookies.
- `.Body`: parsed JSON/form data, or a string for other formats.
- `.RawBody`: the captured source body.
- `.Response`: status, headers, response bytes, parsed/raw bounded Body,
  overflow state, start/end time, and duration.
- `.Item` and `.ItemIndex`: the current array element selected by `forEach`.

Convenience functions are `jsonPath`, `header`, `query`, and `cookie`.
Templates are supported in the target URL, target headers, body, executable
arguments, environment, and standard input. Use `toJson` whenever a source
value is inserted into JSON.

`maxBytes` bounds memory per in-flight matched request. `overflow` is `skip` or
`reject`; `reject` is effective only before the backend response begins. The
source body is restored exactly after inspection.

## Response Body and Per-Item Actions

Set `responseBody.readMode` to `capture` on an `after` rule to inspect a backend
response. `maxBytes` is capped at 16 MiB; `overflow: skip` forwards the complete
backend response to the client but does not dispatch the action. `format` can
force JSON parsing when the backend content type is missing or inaccurate.

`jsonPath` supports object keys, array indexes, and quoted keys containing dots
or slashes, for example:

```text
$.items[0].labels["example.com/change-id"]
```

For a list/polling API, `forEach` selects an array and optionally filters each
element with JSONPath regular expressions:

```yaml
forEach:
  source: responseBody
  jsonPath: $.items
  matchJSONPaths:
    $.status: ^(succeeded|failed)$
```

Each matching element creates an independent action with `.Item` and
`.ItemIndex`. `source` can also be `requestBody`.

## Polling Deduplication

`deduplication.keyTemplate` renders a stable key immediately before the target
call. A claimed key is completed only after target response validation
succeeds. A failed call releases its claim so a later poll can retry.

The `memory` backend is bounded but process-local and is intended for a single
replica or tests. The `mysql` backend hashes keys with SHA-256 and uses a shared
lease/TTL table, suppressing duplicates across workers, rules, polls, and
Traefik replicas. `dsn` or `dsnFile` is required; a Kubernetes
`runtime://alias` DSN keeps material private. `autoCreate` creates the
configured table and expiration index. The database account needs `SELECT`,
`INSERT`, `UPDATE`, `DELETE`, and, with `autoCreate`, `CREATE` permissions.

## Credential Sets and Lifecycle Binding

`credentialSets` separates destination credentials from business rules. A
credential set can be selected directly with
`http.auth.credentialSet`, or indirectly by a lifecycle `binding`.
`runtimeValues` is an arbitrary alias map with `secretKeyRef`,
`configMapKeyRef`, `file`, or literal `value` sources. Exact references use
`runtime://alias`; `${runtime.alias}` interpolates into URLs, Headers, Body
templates, authentication, TLS paths, and Exec fields. The code assigns no
meaning to alias or Secret key names. Legacy `parameterFiles`, `tokenFile`, and
`passwordFile` remain supported and are re-read at `reloadInterval`.

`immutableParameters` defines the destination identity. Values such as a
project ID and project name belong here. If any immutable value changes under
an existing credential-set name, the reload is rejected and the last valid
snapshot remains active. Mutable values such as a token or service caller can
rotate without restarting Traefik. Kubernetes source changes trigger a dynamic
middleware rebuild while the process-level credential cache retains the last
valid immutable identity.

A long-running action commonly has separate start and terminal source
requests. Its start rule uses:

```yaml
binding:
  mode: create
  keyTemplate: 'deployment:{{ jsonPath .Response.Body "$.data.id" }}'
  credentialSet: deployment-project-1001
  ttl: 4320h
  mysql:
    dsn: runtime://actionDatabase
    table: traefik_external_action_binding
    autoCreate: true
```

When one source API can create lifecycles for multiple configured
destinations, the start rule can replace `credentialSet` with
`allowedCredentialSets`. ExternalAction evaluates the configured identity
match after `forEach` expansion and selects exactly one credential set whose
`immutableParameters` equal the source values. No project ID needs to be
hard-coded in the rule:

```yaml
binding:
  mode: create
  keyTemplate: 'execution:{{ jsonPath .Item "$.execution_id" }}'
  allowedCredentialSets:
    - deployment-project-1001
    - deployment-project-2002
  match:
    templates:
      projectid: '{{ scalarString (jsonPath .Item "$.project_id") }}'
      projectname: '{{ jsonPath .Item "$.project_name" }}'
  mysql:
    dsn: runtime://actionDatabase
    table: traefik_external_action_binding
```

Zero matches skip that item; multiple matches are rejected as ambiguous.
This also supports list responses whose items belong to different projects.

The terminal rule renders the same stable key and resolves the original
credential set:

```yaml
binding:
  mode: resolve
  onMissing: skip
  keyTemplate: 'deployment:{{ jsonPath .Response.Body "$.data.id" }}'
  allowedCredentialSets:
    - deployment-project-1001
  ttl: 4320h
  mysql:
    dsn: runtime://actionDatabase
    table: traefik_external_action_binding
    autoCreate: true
```

The MySQL table stores both the lifecycle-key mapping and a permanent
credential-set identity hash. It therefore detects a project name being reused
after a Pod restart. `memory` is available for tests and single-process
development, but production replicas should use the shared MySQL backend.

Resolve defaults to `onMissing: fail`. Polling/list endpoints that can return
historical terminal objects created before governance was enabled should use
`onMissing: skip`: an absent or expired binding becomes a successful no-op,
while the deduplication lease is released rather than completed. A later poll
can therefore retry if it initially raced ahead of the start rule. Database or
target failures are not skipped and continue to follow `failurePolicy`.

If historical objects must be reported instead of skipped, use
`mode: resolveOrCreate`. It first resolves the stable lifecycle key. When the
binding is absent, it selects exactly one `allowedCredentialSets` entry using
the same immutable `match` sources as a create binding and persists the result.
The action template can inspect `.LifecycleBinding.Created`; this is `true`
only for the dispatch that recovered the missing binding. For example, a
terminal reconciliation action can emit both a recovered start log and the end
log in one idempotent request, while normal bound lifecycles emit only the end:

```yaml
binding:
  mode: resolveOrCreate
  keyTemplate: 'deployment:{{ jsonPath .Response.Body "$.data.id" }}'
  allowedCredentialSets: [deployment-project-1001, deployment-project-2002]
  match:
    templates:
      projectid: '{{ scalarString (jsonPath .Response.Body "$.data.project_id") }}'
  mysql:
    dsn: runtime://actionDatabase
    table: traefik_external_action_binding
```

`resolveOrCreate` is destination-neutral: the middleware does not contain a
project ID, credential name, workflow field, or vendor-specific branch. Ambiguous or
unmatched identities are rejected rather than routed to a guessed destination.

To migrate a destination project, create new runtime aliases and a new
credential-set name, add a mutually exclusive start rule, and add the new name
to terminal rules' `allowedCredentialSets`. Retain the old credential set until
queued actions and all active business lifecycles have produced their terminal
event. Remove the old rules before removing their Secret keys.

## Target Authentication

HTTP authentication types are `basic`, `bearer`, `header`, and `headers`.
Secrets can be loaded with `passwordFile`, `tokenFile`, or `parameterFiles`.
They can also be supplied through arbitrary runtime aliases in the equivalent
inline fields. Set only `credentialSet` under `http.auth` when selecting a
named set; it cannot be combined with inline authentication fields. TLS
supports a custom CA and mTLS certificate/key pair.

MySQL runtime sources support connection rotation. A candidate DSN is pinged and initialized before atomic
activation; in-flight operations drain on the previous pool, and a failed
rotation keeps the previous pool active. No Traefik Pod rollout is required.

## Target Response Validation

HTTP status alone is not sufficient for APIs that return application errors in
a successful HTTP response. `responseValidation` can validate both layers:

- `statuses` defines accepted HTTP status ranges. It defaults to `200-299`.
- `successJSONPath` selects an application result from a JSON response.
- `successValues` lists the exact string representations accepted as success.
- `messageJSONPath` selects a safe diagnostic message when validation fails.
- `bodyRegex` can replace `successJSONPath`/`successValues` for non-JSON
  responses.

`successJSONPath` and `bodyRegex` are mutually exclusive. A failed synchronous
`before` action follows `failurePolicy`; the selected diagnostic message is
written to the Traefik error log without logging the complete response body.
Authentication values loaded from configuration or secret files are replaced
with `<redacted>` if the target reflects them in its error message. When
`messageJSONPath` cannot select a message, a sanitized response summary limited
to 1024 characters is used instead.

For example, an API that returns `{"code": 5000, "msg": "invalid input"}` with
HTTP 200 fails when `successJSONPath` is `$.code` and `successValues` is
`["0"]`.

An `exec` target requires an absolute command path and never invokes a shell:

```yaml
exec:
  command: /opt/actions/sdk-adapter
  args:
    - --flow-id
    - '{{ jsonPath .Body "$.flow.id" }}'
  stdin: '{{ .RawBody }}'
```

Only trusted administrators should be allowed to configure executable actions.

## Periodic Observations

Optional `pollers` fetch an independently authenticated HTTP source and feed its
response into the same rules. Configure the source request, interval, timeout,
response byte limit, and an `observation` whose method, host, and path select
the intended rules. Pagination limits bound each poll. Multiple replicas need
shared MySQL coordination and action deduplication to avoid duplicate work.

A failed synchronous action is recorded without preventing independent items
or subsequent pages from being observed. After finishing the bounded scan, the
poller reports the failure and releases its coordination claim so a later poll
can retry. Successfully completed action keys remain deduplicated. Source fetch,
source validation, cancellation and pagination-limit errors still stop the scan.
With `failurePolicy: ignore`, action failures do not fail poller coordination.
Asynchronous dispatch records queue acceptance; worker failures are retried only
when the source is observed again and the failed action claim is available.

An optional `observation.event` wraps the response as a structured observation
event for `before` rules. Its principal is available as `.Body.data.principal`
and in the synthetic request's `X-Governance-Principal` header. This is configured
metadata, not an authenticated end-user identity. Observations are processed
locally; they do not create a public HTTP route or authentication endpoint.

## Operational Boundaries

The asynchronous queue is in memory, not a durable message broker. A restart,
full queue, or exhausted target attempt can lose an action. Polling and persistent
deduplication can reconcile observable source state, but do not guarantee
exactly-once delivery. Use an idempotent destination and an external durable
queue when lossless delivery is required.

Restrict configuration write access and outbound connectivity. Templates can
copy request data to configured targets; avoid forwarding credentials or
sensitive payload fields unintentionally. Use trusted identity sources, keep
secrets in runtime references, and authenticate callback routes separately.

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

Binding and deduplication statements are prepared on the acquired connection
or transaction and closed before releasing its pool lease. In-flight operations
can finish on the previous pool while new operations use the rotated pool.
TLS certificate verification is enabled by default; `insecureSkipVerify` is an
explicit operator opt-out and should remain disabled in production. Executables
are selected by an administrator-supplied absolute path. Templates supply literal
arguments, environment values, and stdin without invoking a shell. Use stdin for
untrusted payloads, or an adapter's end-of-options separator when passing them
as arguments; the adapter must validate any resulting application commands.
