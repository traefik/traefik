---
title: QueryEnrichment
description: Enrich ExternalAction events with bounded, configurable metadata lookups.
---

# QueryEnrichment

QueryEnrichment performs metadata lookups for an ExternalAction event. It does
not change the business request or response. Queries run only after a matching
action has rendered its destination payload and selected its credentials.

Use a separate Middleware for each metadata source. Reference them in order
from `externalAction.enrichment`. A QueryEnrichment can also appear before
ExternalAction in a middleware chain. An explicit reference is necessary for
background pollers, which do not traverse the incoming HTTP chain. A reference
present in both places runs once per event.

```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: order-metadata
spec:
  queryEnrichment:
    runtimeValues:
      apiToken:
        secretKeyRef:
          name: metadata-reader
          key: token
    inputTemplate: '{{ dict "orderID" (index .Payload "orderID") | toJson }}'
    queries:
      - name: order
        allowedOrigins: [https://metadata.example.com]
        timeout: 2s
        maxConcurrent: 8
        cacheTTL: 30s
        cacheEntries: 256
        cacheBytes: 8388608
        http:
          url: 'https://metadata.example.com/orders/{{ index .Input "orderID" | urlquery }}'
          method: GET
          auth:
            type: bearer
            token: runtime://apiToken
          maxResponseBytes: 262144
        resultTemplate: '{{ dict "product" (required "product is required" (jsonPath .Result "$.product")) | toJson }}'
    outputTemplate: |
      {{- $_ := set .Payload "product" (index .Enrichment "order" "product") -}}
      {{ .Payload | toJson }}
```

The consuming ExternalAction adds `enrichment: [order-metadata]`. Its existing
rules, body templates, response capture, authentication, deduplication and
lifecycle bindings continue to apply. Kubernetes references resolve within the
same namespace. Missing or invalid references reject the middleware configuration.

## Templates and authentication

- `.Payload`: the rendered destination JSON body, preserving numeric precision.
- `.Input`: the JSON produced by `inputTemplate` for this middleware.
- `.Result`: the current query's decoded response, available in `resultTemplate`.
- `.Enrichment`: mapped results of completed named queries.
- `.SourceAuthParameters`: the final action's bound credential parameters, before
  query-specific authentication is selected.
- `.LookupTime`, `.OutboundBody`: the query timestamp and exact rendered body.
- Request, response and foreach-item fields from ExternalAction remain available.

Query URLs, headers, JSON bodies, conditions (`when`), results and the outgoing
payload are configurable. Existing authentication and credential-set options are
supported. `hmacSHA256Base64 key message` enables an API's documented HMAC header
scheme without embedding its protocol into the middleware.

Kubernetes Secret/ConfigMap references use the provider's existing watches.
Referenced changes rebuild the configuration without restarting Traefik.
File-backed credential sets retain their reload interval. Cache keys include the
authenticated request, so a rotated credential does not reuse another credential's
result. Keep credential-set identities stable for in-flight lifecycle events;
project changes need new identities and rules rather than repurposed keys.

## Safety, capacity and failures

Each query must declare literal allowed origins. Only GET and POST are accepted;
configure POST only for read-only query APIs. Redirects are not followed. Response
status and configured business-success checks must pass before mapping. Query
credentials and upstream response cookies are not forwarded to clients.

Timeouts, concurrent lookups, response sizes, retained results and caches are
bounded. A query defaults to a 2-second timeout and 8 concurrent requests; its
maximum is 30 seconds and 64 concurrent requests. Each middleware has at most
16 queries and an event at most 8 enrichment middlewares. JSON bodies and retained
mapped results are capped at 1 MiB. Cache is disabled when `cacheTTL` is zero.

Set the parent action timeout to cover its queries plus the final call. Async
actions use the existing bounded queue; synchronous actions wait within their
timeout. Lookup failure follows the action's existing failure policy and releases
an acquired deduplication lease. It does not invent missing metadata or count the
report as delivered. Retry delivery still requires an existing poller, repeated
event, or configured producer retry; enrichment does not create a durable outbox.

## Pagination

GET queries can configure `pagination` with `pageParameter`, `pageSizeParameter`,
`startPage`, `pageSize`, `maxPages`, `itemsJSONPath` and optionally `totalJSONPath`.
Mapping runs per page and the named result becomes an array of mapped pages.
The entire lookup shares one timeout. Missing totals, inconsistent totals, early
empty pages and exhausting `maxPages` fail rather than return partial metadata.
Source APIs should provide stable ordering or snapshot semantics while paging.
