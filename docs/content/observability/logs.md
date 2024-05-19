---
title: "Traefik Logs Documentation"
description: "Logs are a key part of observability in Traefik Proxy. Read the technical documentation to learn their configurations, rotations, and time zones."
---

# Logs

Reading What's Happening
{: .subtitle }

By default, logs are written to stdout, in text format.

## Configuration

### General

Traefik logs concern everything that happens to Traefik itself (startup, configuration, events, shutdown, and so on).

#### `filePath`

By default, the logs are written to the standard output.
You can configure a file path instead using the `filePath` option.

```yaml tab="File (YAML)"
# Writing Logs to a File
log:
  filePath: "/path/to/traefik.log"
```

```toml tab="File (TOML)"
# Writing Logs to a File
[log]
  filePath = "/path/to/traefik.log"
```

```bash tab="CLI"
# Writing Logs to a File
--log.filePath=/path/to/traefik.log
```

#### `format`

By default, the logs use a text format (`common`), but you can also ask for the `json` format in the `format` option.

```yaml tab="File (YAML)"
# Writing Logs to a File, in JSON
log:
  filePath: "/path/to/log-file.log"
  format: json
```

```toml tab="File (TOML)"
# Writing Logs to a File, in JSON
[log]
  filePath = "/path/to/log-file.log"
  format = "json"
```

```bash tab="CLI"
# Writing Logs to a File, in JSON
--log.filePath=/path/to/traefik.log
--log.format=json
```

#### `level`

By default, the `level` is set to `ERROR`.

Alternative logging levels are `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, and `PANIC`.

```yaml tab="File (YAML)"
log:
  level: DEBUG
```

```toml tab="File (TOML)"
[log]
  level = "DEBUG"
```

```bash tab="CLI"
--log.level=DEBUG
```

#### `noColor`

When using the 'common' format, disables the colorized output.

```yaml tab="File (YAML)"
log:
  noColor: true
```

```toml tab="File (TOML)"
[log]
  noColor = true
```

```bash tab="CLI"
--log.nocolor=true
```

## Log Rotation

The rotation of the log files can be configured with the following options.

### `maxSize`

`maxSize` is the maximum size in megabytes of the log file before it gets rotated.
It defaults to 100 megabytes.

```yaml tab="File (YAML)"
log:
  maxSize: 1
```

```toml tab="File (TOML)"
[log]
  maxSize = 1
```

```bash tab="CLI"
--log.maxsize=1
```

### `maxBackups`

`maxBackups` is the maximum number of old log files to retain.
The default is to retain all old log files (though `maxAge` may still cause them to get deleted).

```yaml tab="File (YAML)"
log:
  maxBackups: 3
```

```toml tab="File (TOML)"
[log]
  maxBackups = 3
```

```bash tab="CLI"
--log.maxbackups=3
```

### `maxAge`

`maxAge` is the maximum number of days to retain old log files based on the timestamp encoded in their filename.
Note that a day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings, leap seconds, etc.
The default is not to remove old log files based on age.

```yaml tab="File (YAML)"
log:
  maxAge: 3
```

```toml tab="File (TOML)"
[log]
  maxAge = 3
```

```bash tab="CLI"
--log.maxage=3
```

### `compress`

`compress` determines if the rotated log files should be compressed using gzip.
The default is not to perform compression.

```yaml tab="File (YAML)"
log:
  compress: true
```

```toml tab="File (TOML)"
[log]
  compress = true
```

```bash tab="CLI"
--log.compress=true
```
