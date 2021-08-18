# Lightstep

To enable the Lightstep:

```yaml tab="File (YAML)"
tracing:
  lightstep: {}
```

```toml tab="File (TOML)"
[tracing]
  [tracing.lightstep]
```

```bash tab="CLI"
--tracing.lightstep=true
```

#### `ServerHost`

_Optional - by default no Satellite is used, Default=""_

serverHost is the hostname of the Lightstep Satellite server.

```yaml tab="File (YAML)"
tracing:
  lightstep:
    serverHost: "myHost"
```

```toml tab="File (TOML)"
[tracing]
  [tracing.lightstep]
    serverHost = "myHost"
```

```bash tab="CLI"
--tracing.lightstep.serverhost="myHost"
```

#### `ServerPort`

_Optional - by default no Satellite is used, Default=""_

serverPort is the listener port of the Lightstep Satellite server.

```yaml tab="File (YAML)"
tracing:
  lightstep:
    serverPort: 8383
```

```toml tab="File (TOML)"
[tracing]
  [tracing.lightstep]
    serverPort = 8383
```

```bash tab="CLI"
--tracing.lightstep.serverport=8383
```

#### `Plaintext`

_Optional - by default no Satellite is used, Default=""_

Plaintext is a switch to set plaintext or encrypted communication with the Lightstep Satellite server.

```yaml tab="File (YAML)"
tracing:
  lightstep:
    plaintext: true
```

```toml tab="File (TOML)"
[tracing]
  [tracing.lightstep]
    plaintext = true
```

```bash tab="CLI"
--tracing.lightstep.plaintext=true
```

#### `AccessToken`

_Optional, Default=""_

Lightstep Access Token is the token used to connect to your Lightstep project.

```yaml tab="File (YAML)"
tracing:
  lightstep:
    accessToken: "mytoken"
```

```toml tab="File (TOML)"
[tracing]
  [tracing.lightstep]
    accessToken = "mytoken"
```

```bash tab="CLI"
--tracing.lightstep.accesstoken="mytoken"
```
