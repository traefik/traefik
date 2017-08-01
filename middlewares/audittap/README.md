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
```

The properties are as follow:

* type (mandatory): the type of sink audit events will be published to
* proxyingFor (mandatory): determines the auditing style. Values can be API or RATE
* auditSource (mandatory for API): the auditSource value to be included in API audit events
* auditType (mandatory for API): the auditType value to be included in API audit events
* encryptSecret (optional): base64 encoded AES-256 key, if provided logged audit events will be encrypted