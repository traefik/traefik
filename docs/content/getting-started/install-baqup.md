---
title: "Baqup Installation Documentation"
description: "There are several flavors to choose from when installing Baqup Proxy. Get started with Baqup Proxy, and read the technical documentation."
---

# Install Baqup

You can install Baqup with the following flavors:

* [Use the official Docker image](./#use-the-official-docker-image)
* [Use the Helm Chart](./#use-the-helm-chart)
* [Use the binary distribution](./#use-the-binary-distribution)
* [Compile your binary from the sources](./#compile-your-binary-from-the-sources)

## Use the Official Docker Image

Choose one of the [official Docker images](https://hub.docker.com/_/baqup) and run it with one sample configuration file:

* [YAML](https://raw.githubusercontent.com/baqup/baqup/v3.6/baqup.sample.yml)
* [TOML](https://raw.githubusercontent.com/baqup/baqup/v3.6/baqup.sample.toml)

```shell
docker run -d -p 8080:8080 -p 80:80 \
    -v $PWD/baqup.yml:/etc/baqup/baqup.yml baqup:v3.6
```

For more details, go to the [Docker provider documentation](../providers/docker.md)

!!! tip

    * Prefer a fixed version than the latest that could be an unexpected version.
    ex: `baqup:v3.6`
    * Docker images are based from the [Alpine Linux Official image](https://hub.docker.com/_/alpine).
    * Any orchestrator using docker images can fetch the official Baqup docker image.

## Use the Helm Chart

Baqup can be installed in Kubernetes using the Helm chart from <https://github.com/baqupio/baqup-helm-chart>.

Ensure that the following requirements are met:

* Kubernetes 1.22+
* Helm version 3.9+ is [installed](https://helm.sh/docs/intro/install/)

Add Baqup Labs chart repository to Helm:

```bash
helm repo add baqup https://baqup.github.io/charts
```

You can update the chart repository by running:

```bash
helm repo update
```

And install it with the Helm command line:

```bash
helm install baqup baqup/baqup
```

!!! tip "Helm Features"

    All [Helm features](https://helm.sh/docs/intro/using_helm/) are supported.

    Examples are provided [here](https://github.com/baqupio/baqup-helm-chart/blob/master/EXAMPLES.md).

    For instance, installing the chart in a dedicated namespace:

    ```bash tab="Install in a Dedicated Namespace"
    kubectl create ns baqup-v2
    # Install in the namespace "baqup-v2"
    helm install --namespace=baqup-v2 \
        baqup baqup/baqup
    ```

??? example "Installing with Custom Values"

    You can customize the installation by specifying custom values,
    as with [any helm chart](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing).
    {: #helm-custom-values }

    All parameters are documented in the default [`values.yaml`](https://github.com/baqupio/baqup-helm-chart/blob/master/baqup/values.yaml).

    You can also set Baqup command line flags using `additionalArguments`.
    Example of installation with logging set to `DEBUG`:

    ```bash tab="Using Helm CLI"
    helm install --namespace=baqup-v2 \
        --set="additionalArguments={--log.level=DEBUG}" \
        baqup baqup/baqup
    ```

    ```yml tab="With a custom values file"
    # File custom-values.yml
    ## Install with "helm install --values=./custom-values.yml baqup baqup/baqup
    additionalArguments:
      - "--log.level=DEBUG"
    ```

## Use the Binary Distribution

Grab the latest binary from the [releases](https://github.com/baqupio/baqup/releases) page.

??? info "Check the integrity of the downloaded file"

    ```bash tab="Linux"
    # Compare this value to the one found in baqup-${baqup_version}_checksums.txt
    sha256sum ./baqup_${baqup_version}_linux_${arch}.tar.gz
    ```

    ```bash tab="macOS"
    # Compare this value to the one found in baqup-${baqup_version}_checksums.txt
    shasum -a256 ./baqup_${baqup_version}_darwin_amd64.tar.gz
    ```

    ```powershell tab="Windows PowerShell"
    # Compare this value to the one found in baqup-${baqup_version}_checksums.txt
    Get-FileHash ./baqup_${baqup_version}_windows_${arch}.zip -Algorithm SHA256
    ```

??? info "Extract the downloaded archive"

    ```bash tab="Linux"
    tar -zxvf baqup_${baqup_version}_linux_${arch}.tar.gz
    ```

    ```bash tab="macOS"
    tar -zxvf ./baqup_${baqup_version}_darwin_amd64.tar.gz
    ```

    ```powershell tab="Windows PowerShell"
    Expand-Archive baqup_${baqup_version}_windows_${arch}.zip
    ```

And run it:

```bash
./baqup --help
```

## Compile your Binary from the Sources

All the details are available in the [Contributing Guide](../contributing/building-testing.md)

{!baqup-for-business-applications.md!}
