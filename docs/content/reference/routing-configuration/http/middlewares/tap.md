---
title: "Traefik Tap Documentation"
description: "In Traefik Proxy's HTTP middleware, Tap sends a copy of the requests and responses going through it to Traefik services. Read the technical documentation."
---

The `tap` middleware sends a copy of the requests and responses going through it to Traefik services, as JSON records.

Because the destinations are ordinary Traefik services, they benefit from load balancing, health checks, retries and
`serversTransport` options like any other service, and they can be anything reachable over HTTP: a message queue
gateway, an audit store, or a sidecar of your own.

Requests and responses are sent as two separate records, each to its own service, so that both streams can be
consumed and scaled independently. Set only one of the two to capture a single side.

!!! warning "Sensitive data"

    Records carry the headers as they go through Traefik, `Authorization` and `Cookie` included,
    and the bodies too once `recordBody` is enabled.
    Restrict what is captured with `recordBody` and `maxRecordBodySize`,
    and treat the tap services and their storage as holding production secrets.

!!! info "Difference with the Mirroring service"

    The [Mirroring](../../../../reference/routing-configuration/http/load-balancing/service.md) service replays
    requests to another service, and the replayed requests are handled as real traffic.
    The `tap` middleware sends a description of the exchange, response included, and never replays anything.

!!! tip "Position in the middleware chain"

    The records describe the request and the response as they are at the position of the `tap` middleware in the chain.

    - Placed first, it records the request as received from the client, and the response as sent to the client.
    - Placed last, it records the request as sent to the backend, and the response as received from the backend.

    To get both, use two `tap` middlewares, one first and one last in the chain.
    Their records can be matched with the `traceId`, or with a correlation header listed in `requestHeaders`.

## Configuration Examples

```yaml tab="Structured (YAML)"
http:
  middlewares:
    test-tap:
      tap:
        request:
          service: audit-requests
          path: /records/requests
          recordBody: true
          maxRecordBodySize: 4096
          rejectOnRecordError: true
          timeout: 5s
        response:
          service: audit-responses
          path: /records/responses
          recordBody: true
          maxRecordBodySize: 4096
          requestHeaders:
            - X-Request-Id
          rejectOnRecordError: false
          timeout: 5s
```

```toml tab="Structured (TOML)"
[http.middlewares]
  [http.middlewares.test-tap.tap]
    [http.middlewares.test-tap.tap.request]
      service = "audit-requests"
      path = "/records/requests"
      recordBody = true
      maxRecordBodySize = 4096
      rejectOnRecordError = true
      timeout = "5s"
    [http.middlewares.test-tap.tap.response]
      service = "audit-responses"
      path = "/records/responses"
      recordBody = true
      maxRecordBodySize = 4096
      requestHeaders = ["X-Request-Id"]
      rejectOnRecordError = false
      timeout = "5s"
```

```yaml tab="Labels"
labels:
  - "traefik.http.middlewares.test-tap.tap.request.service=audit-requests"
  - "traefik.http.middlewares.test-tap.tap.request.path=/records/requests"
  - "traefik.http.middlewares.test-tap.tap.request.recordBody=true"
  - "traefik.http.middlewares.test-tap.tap.request.maxRecordBodySize=4096"
  - "traefik.http.middlewares.test-tap.tap.request.rejectOnRecordError=true"
  - "traefik.http.middlewares.test-tap.tap.request.timeout=5s"
  - "traefik.http.middlewares.test-tap.tap.response.service=audit-responses"
  - "traefik.http.middlewares.test-tap.tap.response.path=/records/responses"
  - "traefik.http.middlewares.test-tap.tap.response.recordBody=true"
  - "traefik.http.middlewares.test-tap.tap.response.maxRecordBodySize=4096"
  - "traefik.http.middlewares.test-tap.tap.response.requestHeaders=X-Request-Id"
  - "traefik.http.middlewares.test-tap.tap.response.rejectOnRecordError=false"
  - "traefik.http.middlewares.test-tap.tap.response.timeout=5s"
```

```json tab="Tags"
{
  //...
  "Tags" : [
    "traefik.http.middlewares.test-tap.tap.request.service=audit-requests",
    "traefik.http.middlewares.test-tap.tap.request.path=/records/requests",
    "traefik.http.middlewares.test-tap.tap.request.recordBody=true",
    "traefik.http.middlewares.test-tap.tap.request.maxRecordBodySize=4096",
    "traefik.http.middlewares.test-tap.tap.request.rejectOnRecordError=true",
    "traefik.http.middlewares.test-tap.tap.request.timeout=5s",
    "traefik.http.middlewares.test-tap.tap.response.service=audit-responses",
    "traefik.http.middlewares.test-tap.tap.response.path=/records/responses",
    "traefik.http.middlewares.test-tap.tap.response.recordBody=true",
    "traefik.http.middlewares.test-tap.tap.response.maxRecordBodySize=4096",
    "traefik.http.middlewares.test-tap.tap.response.requestHeaders=X-Request-Id",
    "traefik.http.middlewares.test-tap.tap.response.rejectOnRecordError=false",
    "traefik.http.middlewares.test-tap.tap.response.timeout=5s"
  ]
}
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-tap
spec:
  tap:
    request:
      service:
        name: audit-requests
        port: 80
      path: /records/requests
      recordBody: true
      maxRecordBodySize: 4096
      rejectOnRecordError: true
      timeout: 5s
    response:
      service:
        name: audit-responses
        port: 80
      path: /records/responses
      recordBody: true
      maxRecordBodySize: 4096
      requestHeaders:
        - X-Request-Id
      rejectOnRecordError: false
      timeout: 5s
```

## Configuration Options

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| <a id="opt-request" href="#opt-request" title="#opt-request">`request`</a> | Where and how the request records are sent. When omitted, requests are not recorded. More information [here](#request). | | No |
| <a id="opt-response" href="#opt-response" title="#opt-response">`response`</a> | Where and how the response records are sent. When omitted, responses are not recorded. More information [here](#response). | | No |

At least one of `request` and `response` must be set.

### request

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| <a id="opt-service" href="#opt-service" title="#opt-service">`service`</a> | Name of the service the records are sent to. | | Yes |
| <a id="opt-path" href="#opt-path" title="#opt-path">`path`</a> | Path of the requests sending the records to the service. | / | No |
| <a id="opt-recordBody" href="#opt-recordBody" title="#opt-recordBody">`recordBody`</a> | Whether the body is part of the records. | false | No |
| <a id="opt-maxRecordBodySize" href="#opt-maxRecordBodySize" title="#opt-maxRecordBodySize">`maxRecordBodySize`</a> | Maximum number of body bytes in a record, and held in memory. More information [here](#maxrecordbodysize). | | With `recordBody` |
| <a id="opt-rejectOnRecordError" href="#opt-rejectOnRecordError" title="#opt-rejectOnRecordError">`rejectOnRecordError`</a> | Rejects the request with a `500` when the record cannot be sent. More information [here](#rejectonrecorderror). | false | No |
| <a id="opt-timeout" href="#opt-timeout" title="#opt-timeout">`timeout`</a> | Maximum duration allowed to send a record. A value of `0` means no limit. | 10s | No |

### response

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| <a id="opt-service-2" href="#opt-service-2" title="#opt-service-2">`service`</a> | Name of the service the records are sent to. | | Yes |
| <a id="opt-path-2" href="#opt-path-2" title="#opt-path-2">`path`</a> | Path of the requests sending the records to the service. | / | No |
| <a id="opt-recordBody-2" href="#opt-recordBody-2" title="#opt-recordBody-2">`recordBody`</a> | Whether the body is part of the records. | false | No |
| <a id="opt-maxRecordBodySize-2" href="#opt-maxRecordBodySize-2" title="#opt-maxRecordBodySize-2">`maxRecordBodySize`</a> | Maximum number of body bytes in a record, and held in memory. More information [here](#maxrecordbodysize). | | With `recordBody` |
| <a id="opt-requestHeaders" href="#opt-requestHeaders" title="#opt-requestHeaders">`requestHeaders`</a> | Request headers described in the records, along with the request line. More information [here](#requestheaders). | | No |
| <a id="opt-rejectOnRecordError-2" href="#opt-rejectOnRecordError-2" title="#opt-rejectOnRecordError-2">`rejectOnRecordError`</a> | Rejects the request with a `500` when the record cannot be sent. More information [here](#rejectonrecorderror). | false | No |
| <a id="opt-timeout-2" href="#opt-timeout-2" title="#opt-timeout-2">`timeout`</a> | Maximum duration allowed to send a record. A value of `0` means no limit. | 10s | No |

### rejectOnRecordError

When the record cannot be sent:

- With `false`, the error is logged, and the request is served.
- With `true`, the request is rejected with a `500 Internal Server Error`.

With `true`, the record is sent first:
the request record before the request reaches the backend,
and the response record before the response reaches the client.
Nothing is served unrecorded.
The `request` and `response` settings are independent.

On `response`, `true` also means:

- The response is held whole in memory until its record is accepted.
  WebSocket upgrades are served as usual, and their record is sent as with `false`.
- Informational responses, such as `103 Early Hints`, are dropped.

In both cases, records are sent inline, within `timeout`:
the time taken by the tap service adds to the latency of every request.
Keep `timeout` short, and the tap service close to Traefik.

### requestHeaders

A response record describes the response alone, which is enough to store it, but not to tell which exchange it
belongs to without fetching the matching request record. `requestHeaders` adds a `request` object to the response
records, holding the request line and the listed headers, and nothing else: the body of a request is never part of a
response record.

It is typically used to carry a correlation identifier set by a middleware standing before the tap in the chain.

Headers absent from the request are absent from the record.
The option does not exist on the `request` records, which already describe the whole request.

### maxRecordBodySize

`maxRecordBodySize` is the maximum number of body bytes in a record.
A larger body is cut in the record, which sets `bodyTruncated` to `true`.
The backend and the client still get the whole body.

It also bounds the memory used for the body:
only the recorded bytes are held, and the rest of the body is streamed,
except for a response with `rejectOnRecordError`, which is held whole.
It has no default, as no limit fits every route, and must be set with `recordBody`.
A value of `-1` means no limit, and holds the whole body in memory.

## Records

A record is sent as a `POST` request to the configured service and path, with a JSON body:

```json
{
  "id": "9f8b1c8a3d4e5f60718293a4b5c6d7e8",
  "kind": "response",
  "time": "2026-08-28T14:02:11.123456789Z",
  "traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
  "response": {
    "status": 201,
    "headers": {
      "Content-Type": ["application/json"]
    },
    "body": "eyJpZCI6MX0=",
    "duration": 14234000
  }
}
```

| Field | Description |
|:------|:------------|
| <a id="opt-id" href="#opt-id" title="#opt-id">`id`</a> | Identifier of the exchange. The request record and the response record of the same exchange share it. |
| <a id="opt-kind" href="#opt-kind" title="#opt-kind">`kind`</a> | Either `request` or `response`. |
| <a id="opt-time" href="#opt-time" title="#opt-time">`time`</a> | Time at which the request was received. |
| <a id="opt-traceId" href="#opt-traceId" title="#opt-traceId">`traceId`</a> | Trace the request belongs to, when tracing is enabled. |
| <a id="opt-request-2" href="#opt-request-2" title="#opt-request-2">`request`</a> | Method, URL, host, protocol, remote address, headers and body. Set on `request` records, and on `response` records when `requestHeaders` is set. |
| <a id="opt-response-2" href="#opt-response-2" title="#opt-response-2">`response`</a> | Status, headers, body, and duration in nanoseconds. Set on `response` records. |

Bodies are part of the records only when `recordBody` is enabled.
They are base64-encoded, as they are not necessarily valid UTF-8, and carry a `bodyTruncated` flag when they exceed `maxRecordBodySize`.

The same information is also carried by two headers, so that a service handling both kinds of records can dispatch
them without parsing the payload:

| Header | Description |
|:-------|:------------|
| <a id="opt-X-Tap-Id" href="#opt-X-Tap-Id" title="#opt-X-Tap-Id">`X-Tap-Id`</a> | The record `id`. |
| <a id="opt-X-Tap-Record" href="#opt-X-Tap-Record" title="#opt-X-Tap-Record">`X-Tap-Record`</a> | The record `kind`. |

A tap service is expected to answer with a status below `400`.
Anything else is treated as a failure to record.
