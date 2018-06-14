# Audittap Configuration

Audit taps can be enabled by including appropriate configuration in the toml
configuration consumed by Traefik.

An example follows

```
[auditSink]
  type = "AMQP"
  endpoint = "amqp://localhost:5672/"
  destination = "audit"
  numProducers = 1
  channelLength = 1
  diskStorePath = "/tmp/goque"
  proxyingFor = "API"
  auditSource = "localSource"
  auditType = "localType"
  encryptSecret = "RDFXVxTgrrT9IseypJrwDLzk/nTVeTjbjaUR3RVyv94="
  maxAuditLength = "2M"
  maxPayloadContentsLength = "99K"
  [auditSink.exclusions]
    [auditSink.exclusions.exc1]
    headerName = "RequestHost"
    startsWith = ["captain", "docktor"]
    [auditSink.exclusions.exc2]
    headerName = "RequestPath"
    contains = ["/ping/ping"]
```

The properties are as follow:

* type (mandatory): the type of sink audit events will be published to. Can be AMQP|Blackhole
* proxyingFor (mandatory): determines the auditing style. Values can be API or RATE
* auditSource (mandatory for API): the auditSource value to be included in API audit events
* auditType (mandatory for API): the auditType value to be included in API audit events
* encryptSecret (optional): base64 encoded AES-256 key, if provided logged audit events will be encrypted
* maxAuditLength (optional): maximum byte length of audit defaulted to 100K. e.g 33K or 3M
* maxPayloadContentsLength (optional): maximum combined byte length of audit.requestPayload.contents and audit.responsePayload.contents. e.g 15K or 2M
* auditSink.exclusions.excname (optional): excludes a request from auditing based on the header name when the header
satisfies any of the specified values. Matching condition can be
    * contains
    * endsWith
    * startsWith
    * matches (a regex pattern)
