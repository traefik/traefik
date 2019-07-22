# DataDog

To enable the DataDog:

```toml tab="File (TOML)"
[tracing]
  [tracing.dataDog]
```

```yaml tab="File (YAML)"
tracing:
  dataDog: {}
```

```bash tab="CLI"
--tracing.datadog=true
```

#### `localAgentHostPort`

_Required, Default="127.0.0.1:8126"_

Local Agent Host Port instructs reporter to send spans to datadog-tracing-agent at this address.

```toml tab="File (TOML)"
[tracing]
  [tracing.dataDog]
    localAgentHostPort = "127.0.0.1:8126"
```

```yaml tab="File (YAML)"
tracing:
  dataDog:
    localAgentHostPort: 127.0.0.1:8126
```

```bash tab="CLI"
--tracing.datadog.localAgentHostPort="127.0.0.1:8126"
```

#### `debug`

_Optional, Default=false_

Enable DataDog debug.

```toml tab="File (TOML)"
[tracing]
  [tracing.dataDog]
    debug = true
```

```yaml tab="File (YAML)"
tracing:
  dataDog:
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
  [tracing.dataDog]
    globalTag = "sample"
```

```yaml tab="File (YAML)"
tracing:
  dataDog:
    globalTag: sample
```

```bash tab="CLI"
--tracing.datadog.globalTag="sample"
```

#### `prioritySampling`

_Optional, Default=false_

Enable priority sampling. When using distributed tracing,
this option must be enabled in order to get all the parts of a distributed trace sampled.

```toml tab="File (TOML)"
[tracing]
  [tracing.dataDog]
    prioritySampling = true
```

```yaml tab="File (YAML)"
tracing:
  dataDog:
    prioritySampling: true
```

```bash tab="CLI"
--tracing.datadog.prioritySampling=true
```
