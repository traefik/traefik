# The Debug Mode

Getting More Information (Not For Production)
{: .subtitle }

The debug mode will make Traefik be _extremely_ verbose in its logs, and is NOT intended for production purposes.

## Configuration Example

??? example "TOML -- Enabling the Debug Mode"

    ```toml
    [Global]
       debug = true
    ```