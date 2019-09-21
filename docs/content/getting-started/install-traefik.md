# Install Traefik

You can install Traefik with the following flavors:

* [Use the official Docker image](./#use-the-official-docker-image)
* [Use the binary distribution](./#use-the-binary-distribution)
* [Compile your binary from the sources](./#compile-your-binary-from-the-sources)

## Use the Official Docker Image

Choose one of the [official Docker images](https://hub.docker.com/_/traefik) and run it with the [sample configuration file](https://raw.githubusercontent.com/containous/traefik/v2.0/traefik.sample.toml):

```bash
docker run -d -p 8080:8080 -p 80:80 \
    -v $PWD/traefik.toml:/etc/traefik/traefik.toml traefik:v2.0
```

For more details, go to the [Docker provider documentation](../providers/docker.md)

!!! tip

    * Prefer a fixed version than the latest that could be an unexpected version.
    ex: `traefik:v2.0.0`
    * Docker images comes in 2 flavors: scratch based or alpine based.
    * All the orchestrator using docker images could fetch the official Traefik docker image.

## Use the Binary Distribution

Grab the latest binary from the [releases](https://github.com/containous/traefik/releases) page.

??? info "Check the integrity of the downloaded file"

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

??? info "Extract the downloaded archive"

    ```bash tab="Linux"
    tar -zxvf traefik_${traefik_version}_linux_${arch}.tar.gz
    ```

    ```bash tab="macOS"
    tar -zxvf ./traefik_${traefik_version}_darwin_amd64.tar.gz
    ```

    ```powershell tab="Windows PowerShell"
    Expand-Archive traefik_${traefik_version}_windows_${arch}.zip
    ```

And run it:

```bash
./traefik --help
```

## Compile your Binary from the Sources

All the details are available in the [Contributing Guide](../contributing/building-testing.md)
