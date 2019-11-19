# Instana

To enable the Instana:

```toml tab="File (TOML)"
[tracing]
  [tracing.instana]
```

```yaml tab="File (YAML)"
tracing:
  instana: {}
```

```bash tab="CLI"
--tracing.instana=true
```

#### `localAgentHost`

_Require, Default="127.0.0.1"_

Local Agent Host instructs reporter to send spans to instana-agent at this address.

```toml tab="File (TOML)"
[tracing]
  [tracing.instana]
    localAgentHost = "127.0.0.1"
```

```yaml tab="File (YAML)"
tracing:
  instana:
    localAgentHost: 127.0.0.1
```

```bash tab="CLI"
--tracing.instana.localAgentHost=127.0.0.1
```

#### `localAgentPort`

_Require, Default=42699_

Local Agent port instructs reporter to send spans to the instana-agent at this port.

```toml tab="File (TOML)"
[tracing]
  [tracing.instana]
    localAgentPort = 42699
```

```yaml tab="File (YAML)"
tracing:
  instana:
    localAgentPort: 42699
```

```bash tab="CLI"
--tracing.instana.localAgentPort=42699
```

#### `logLevel`

_Require, Default="info"_

Set Instana tracer log level.

Valid values for logLevel field are:

- `error`
- `warn`
- `debug`
- `info`

```toml tab="File (TOML)"
[tracing]
  [tracing.instana]
    logLevel = "info"
```

```yaml tab="File (YAML)"
tracing:
  instana:
    logLevel: info
```

```bash tab="CLI"
--tracing.instana.logLevel=info
```
