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

!!! info "Difference with the Mirroring service"

    The [Mirroring](../../../../reference/routing-configuration/http/load-balancing/service.md) service replays
    requests to another service, and the replayed requests are handled as real traffic.
    The `tap` middleware sends a description of the exchange, response included, and never replays anything.

## Configuration Examples

```yaml tab="Structured (YAML)"
http:
  middlewares:
    test-tap:
      tap:
        request:
          service: audit-requests
          path: /records/requests
        response:
          service: audit-responses
          path: /records/responses
```

```toml tab="Structured (TOML)"
[http.middlewares]
  [http.middlewares.test-tap.tap]
    [http.middlewares.test-tap.tap.request]
      service = "audit-requests"
      path = "/records/requests"
    [http.middlewares.test-tap.tap.response]
      service = "audit-responses"
      path = "/records/responses"
```

```yaml tab="Labels"
labels:
  - "traefik.http.middlewares.test-tap.tap.request.service=audit-requests"
  - "traefik.http.middlewares.test-tap.tap.request.path=/records/requests"
  - "traefik.http.middlewares.test-tap.tap.response.service=audit-responses"
  - "traefik.http.middlewares.test-tap.tap.response.path=/records/responses"
```

```json tab="Tags"
{
  //...
  "Tags" : [
    "traefik.http.middlewares.test-tap.tap.request.service=audit-requests",
    "traefik.http.middlewares.test-tap.tap.request.path=/records/requests",
    "traefik.http.middlewares.test-tap.tap.response.service=audit-responses",
    "traefik.http.middlewares.test-tap.tap.response.path=/records/responses"
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
    response:
      service:
        name: audit-responses
        port: 80
      path: /records/responses
```

## Configuration Options

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| <a id="opt-request" href="#opt-request" title="#opt-request">`request`</a> | Where and how the requests are sent. When omitted, requests are not sent. More information [here](#request-and-response). | | No |
| <a id="opt-response" href="#opt-response" title="#opt-response">`response`</a> | Where and how the responses are sent. When omitted, responses are not sent. More information [here](#request-and-response). | | No |
| <a id="opt-timeout" href="#opt-timeout" title="#opt-timeout">`timeout`</a> | Maximum duration allowed to send a record. | 10s | No |

At least one of `request` and `response` must be set.

### request and response

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| <a id="opt-service" href="#opt-service" title="#opt-service">`service`</a> | Name of the service the records are sent to. | | Yes |
| <a id="opt-path" href="#opt-path" title="#opt-path">`path`</a> | Path of the requests sending the records to the service. | / | No |
| <a id="opt-body" href="#opt-body" title="#opt-body">`body`</a> | Whether the body is part of the records. | true | No |
| <a id="opt-maxBodySize" href="#opt-maxBodySize" title="#opt-maxBodySize">`maxBodySize`</a> | Maximum body size in bytes kept in a record. A larger body is truncated, and the record is flagged as truncated. A negative value means no limit. More information [here](#maxbodysize). | -1 | No |
| <a id="opt-requestHeaders" href="#opt-requestHeaders" title="#opt-requestHeaders">`requestHeaders`</a> | Request headers described in the records, along with the request line. Only allowed on the `response` records. More information [here](#requestheaders). | | No |
| <a id="opt-failClosed" href="#opt-failClosed" title="#opt-failClosed">`failClosed`</a> | Rejects the request when the record cannot be sent. More information [here](#failclosed). | false | No |

### failClosed

By default the middleware is best-effort: a record that cannot be sent is logged, and the request is served anyway.

With `failClosed` set to `true` on a direction, its record is sent *before* the data it describes is handed over: the
request record before the request reaches the backend, and the response record before the response reaches the client.
When the record cannot be sent, the request is rejected with a `500 Internal Server Error`, and the backend response,
if any, is discarded. This guarantees that nothing is served without being recorded.

The two directions are independent. Failing closed on the requests alone guarantees that no request reaches the
backend unrecorded, while the responses keep being streamed and their records are sent on a best-effort basis.

The guarantee comes with two consequences:

- The client latency and availability are coupled to the tap service. Deploying it close to Traefik, as a sidecar,
  keeps that coupling small.
- A response failing closed is withheld until its record is accepted, which means it is buffered whole in memory and
  cannot be flushed as it goes. Do not set `failClosed` on the `response` records of routes serving large payloads or
  streamed responses (Server-Sent Events, long polling). Hijacked connections, such as WebSocket upgrades, are served
  as usual, and the record is then sent on a best-effort basis.

### requestHeaders

A response record describes the response alone, which is enough to store it, but not to tell which exchange it
belongs to without fetching the matching request record. `requestHeaders` adds a `request` object to the response
records, holding the request line and the listed headers, and nothing else: the body of a request is never part of a
response record.

It is typically used to carry a correlation identifier set by a middleware standing before the tap in the chain.

Headers absent from the request are absent from the record. The option is rejected on the `request` records, which
already describe the whole request.

### maxBodySize

Bodies are kept in memory while a record is built. `maxBodySize` bounds that memory, at the cost of truncated
records: the record carries the first `maxBodySize` bytes and sets `bodyTruncated` to `true`.

Truncation applies to the record only. The backend always receives the whole request, and the client always
receives the whole response.

Reading the request body is what makes Traefik answer a `100 Continue` to a client that sent
`Expect: 100-continue`. The expectation is therefore already honored when the middleware has read the body, and the
`Expect` header is not forwarded to the backend, so that the client sees a single informational response.

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

Bodies are base64-encoded, as they are not necessarily valid UTF-8, and carry a `bodyTruncated` flag when they
exceed `maxBodySize`.

The same information is also carried by two headers, so that a service handling both kinds of records can dispatch
them without parsing the payload:

| Header | Description |
|:-------|:------------|
| <a id="opt-X-Tap-Id" href="#opt-X-Tap-Id" title="#opt-X-Tap-Id">`X-Tap-Id`</a> | The record `id`. |
| <a id="opt-X-Tap-Record" href="#opt-X-Tap-Record" title="#opt-X-Tap-Record">`X-Tap-Record`</a> | The record `kind`. |

A tap service is expected to answer with a status below `400`. Anything else is treated as a failure to record.

!!! warning "Sensitive data"

    Records carry the headers and bodies as they go through Traefik, `Authorization` and `Cookie` included.
    Restrict what is captured with `body` and `maxBodySize`, and treat the tap services and their storage as
    holding production secrets.
