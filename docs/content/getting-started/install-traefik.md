# Install Traefik

The first step, before using all the awesome features of Traefik, is to install it.
There are 3 main ways to install Traefik, you can:

* [Use the official Docker Image](./#from-official-docker-image)
* [Use the Prebuild binary](./#from-prebuilt-binary)
* [Compile your binary from the sources](./#from-the-sources)

## From Official Docker Image

Choose one of the [official tiny Docker image](https://hub.docker.com/_/traefik) and run it with the [sample configuration file](https://raw.githubusercontent.com/containous/traefik/master/traefik.sample.toml):

```shell
docker run -d -p 8080:8080 -p 80:80 -v $PWD/traefik.toml:/etc/traefik/traefik.toml traefik:v2.0
```

For more details, go to the [Docker provider documentation](../providers/docker.md)

!!! tip
    Prefer a fixed version than the latest that could be an unexpected version.
    ex: `traefik:v2.0.0`

!!! tip "All the orchestrator using docker images could fetch the official Traefik docker image"

## From Prebuilt Binary

Grab the latest binary from the [releases](https://github.com/containous/traefik/releases) page and run it with the [sample configuration file](https://raw.githubusercontent.com/containous/traefik/master/traefik.sample.toml):

```bash
./traefik
```

??? tip "Check the integrity of the downloaded file"

    ```bash tab="Linux"
    # Compare this value to the one found in traefik-${traefik_version}_checksums.txt
    sha256sum ./traefik_${traefik_version}_linux_${arch}.tar.gz
    ```

    ```bash tab="macOS"
    # Compare this value to the one found in traefik-${traefik_version}_checksums.txt
    shasum -a256 ./traefik_${traefik_version}_darwin_amd64.tar.gz
    ```

    ```powershell tab="Windows PowerShell"
    # Compare this value to the one found in traefik-${traefik_version}_checksums.txt
    Get-FileHash ./traefik_${traefik_version}_windows_${arch}.zip -Algorithm SHA256
    ```

??? tip "Extract the downloaded archive"

    ```bash tab="Linux"
    tar -zxvf traefik_${traefik_version}_linux_${arch}.tar.gz
    ```

    ```bash tab="macOS"
    tar -zxvf -a256 ./traefik_${traefik_version}_darwin_amd64.tar.gz
    ```

    ```powershell tab="Windows PowerShell"
    Expand-Archive traefik_${traefik_version}_windows_${arch}.zip
    ```

## From the Sources

All the details are available in the [Contributing Guide](../contributing/building-testing.md)
