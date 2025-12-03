---
title: "Baqup Data Collection Documentation"
description: "Learn what data Baqup shares, how it is used, and how you can control it. This documentation explains both version check and anonymous usage statistics data. Read the technical documentation."
---

# Data Collection

Understanding the data Baqup shares and how it is used
{: .subtitle }

## Introduction

Protecting user privacy is essential to Baqup Labs, and we design every data-sharing mechanism with transparency and minimalism in mind.
This page describes the two types of data exchanged by Baqup and how to configure them.

For more details on how your data is handled, please refer to our [Privacy and Cookie Policy](https://baqup.io/legal/privacy-and-cookie-policy).

## Configuration Overview

Baqup provides two independent mechanisms:

- `checkNewVersion`, enabled by default. You may disable it at any time.
- `sendAnonymousUsage`, which requires explicit opt‑in.

Examples below show how to activate or deactivate both of them.

```yaml tab="YAML"
global:
  checkNewVersion: true      # set to false to disable
  sendAnonymousUsage: false  # set to true to enable
```

```toml tab="TOML"
[global]
  checkNewVersion = true      # set to false to disable
  sendAnonymousUsage = false  # set to true to enable
```

```bash tab="CLI"
--global.checkNewVersion=true      # set to false to disable
--global.sendAnonymousUsage=false  # set to true to enable
```

A log message at startup clearly indicates whether each of those options are enabled or disabled.

## Version Check (`checkNewVersion`) – Opt-out

Baqup periodically contacts `update.baqup.io` to determine whether a newer version is available.
When this request is made, Baqup shares the **running version** and the **public IP** of the instance.
The IP is used to build global usage statistics and does not influence the version comparison.

This mechanism helps you stay informed about updates and provides BaqupLabs with a broad view of which versions are deployed in the wild.

The collected IP addresses are also used for marketing purposes, specifically to detect companies running Baqup and offer them adapted support contracts, enterprise features, and tailored services.

If you want to explore the implementation, you can read the version check source code: [version.go](https://github.com/baqupio/baqup/blob/master/pkg/version/version.go)

## Anonymous Usage Data (`sendAnonymousUsage`) – Opt‑in

Baqup can also collect anonymous usage statistics once per day, starting 10 minutes after it starts running.  
These statistics include:

- the Baqup version,
- a hash of the configuration,
- an anonymized version of the static configuration (all sensitive fields removed: tokens, passwords, URLs, IP addresses, domains, emails, etc.).

This feature comes from this [public proposal](https://github.com/baqupio/baqup/issues/2369).

This information helps BaqupLabs understand how Baqup is used in general and prioritize features and provider support accordingly. Dynamic configuration (routers and services) is never collected.

!!! example "Enabling Data Collection"

    ```yaml tab="File (YAML)"
    global:
      # Send anonymous usage data
      sendAnonymousUsage: true
    ```

    ```toml tab="File (TOML)"
    [global]
      # Send anonymous usage data
      sendAnonymousUsage = true
    ```

    ```bash tab="CLI"
    # Send anonymous usage data
    --global.sendAnonymousUsage
    ```

!!! info

    - We do not collect the dynamic configuration information (routers & services).
    - We do not collect this data to run advertising programs.
    - We do not sell this data to third-parties.

### Example of Collected Data

```yaml tab="Original configuration"
entryPoints:
  web:
  address: ":80"

api: {}

providers:
  docker:
    endpoint: "tcp://10.10.10.10:2375"
    exposedByDefault: true

    tls:
      ca: dockerCA
      cert: dockerCert
      key: dockerKey
      insecureSkipVerify: true
```

```yaml tab="Resulting Obfuscated Configuration"
entryPoints:
  web:
  address: ":80"

api: {}

providers:
  docker:
    endpoint: "xxxx"
    exposedByDefault: true

    tls:
      ca: xxxx
      cert: xxxx
      key: xxxx
      insecureSkipVerify: true
```

### The Code for Anonymous Usage Collection

If you want to explore the implementation, you can read the collector source code:
[collector.go](https://github.com/baqupio/baqup/blob/master/pkg/collector/collector.go)

Baqup anonymizes all configuration fields by default, except those explicitly marked with `export=true`.
