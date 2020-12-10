# Datadog

To enable the Datadog:

```toml tab="File (TOML)"
[tracing]
  [tracing.datadog]
```

```yaml tab="File (YAML)"
tracing:
  datadog: {}
```

```bash tab="CLI"
--tracing.datadog=true
```

#### `localAgentHostPort`

_Required, Default="127.0.0.1:8126"_

Local Agent Host Port instructs reporter to send spans to datadog-tracing-agent at this address. 
This configuration checks for  `DD_AGENT_HOST` and `DD_TRACE_AGENT_PORT` environment variables and sets the 
`localAgentHostPort` value based on that. These environment variables can be set independent of each other. 
_Setting this configuration via YAML, TOML, or CLI will overwrite the 
environment variable_
```toml tab="File (TOML)"
[tracing]
  [tracing.datadog]
    localAgentHostPort = "127.0.0.1:8126"
```

```yaml tab="File (YAML)"
tracing:
  datadog:
    localAgentHostPort: 127.0.0.1:8126
```

```bash tab="CLI"
--tracing.datadog.localAgentHostPort=127.0.0.1:8126
```

```yaml tab="Environment Variable"
env:
  - name: DD_AGENT_HOST
    valueFrom:
      fieldRef:
        fieldPath: status.hostIP
```

#### `debug`

_Optional, Default=false_

Enable Datadog debug.

```toml tab="File (TOML)"
[tracing]
  [tracing.datadog]
    debug = true
```

```yaml tab="File (YAML)"
tracing:
  datadog:
    debug: true
```

```bash tab="CLI"
--tracing.datadog.debug=true
```

#### `globalTag`

_Optional, Default=empty_

Apply shared tag in a form of Key:Value to all the traces.

```toml tab="File (TOML)"
[tracing]
  [tracing.datadog]
    globalTag = "sample"
```

```yaml tab="File (YAML)"
tracing:
  datadog:
    globalTag: sample
```

```bash tab="CLI"
--tracing.datadog.globalTag=sample
```

#### `prioritySampling`

_Optional, Default=false_

Enable priority sampling. When using distributed tracing,
this option must be enabled in order to get all the parts of a distributed trace sampled.

```toml tab="File (TOML)"
[tracing]
  [tracing.datadog]
    prioritySampling = true
```

```yaml tab="File (YAML)"
tracing:
  datadog:
    prioritySampling: true
```

```bash tab="CLI"
--tracing.datadog.prioritySampling=true
```
