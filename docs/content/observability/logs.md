# Logs

Reading What's Happening
{: .subtitle }

By default, logs are written to stdout, in text format.

## Configuration Example

??? example "Writing Logs in a File"

    ```toml
    [log]
    filePath = "/path/to/traefik.log"
    ```

??? example "Writing Logs in a File, in JSON"

    ```toml
    [log]
    filePath = "/path/to/log-file.log"
    format   = "json"
    ```

## Configuration Options

### General

Traefik logs concern everything that happens to Traefik itself (startup, configuration, events, shutdown, and so on).

#### filePath

By default, the logs are written to the standard output.
You can configure a file path instead using the `filePath` option.

#### format

By default, the logs use a text format (`common`), but you can also ask for the `json` format in the `format` option.   

#### log level

By default, the `level` is set to `error`, but you can choose amongst `debug`, `panic`, `fatal`, `error`, `warn`, and `info`. 

## Log Rotation

Traefik will close and reopen its log files, assuming they're configured, on receipt of a USR1 signal.
This allows the logs to be rotated and processed by an external program, such as `logrotate`.

!!! note
    This does not work on Windows due to the lack of USR signals.
