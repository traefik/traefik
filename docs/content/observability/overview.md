---
title: "Traefik Observability Overview"
description: "Traefik provides Logs, Access Logs, Metrics and Tracing. Read the full documentation to get started."
---

# Overview

Traefik's Observability system
{: .subtitle }

## Logs

Traefik logs informs about everything that happens within Traefik (startup, configuration, events, shutdown, and so on).

Read the [Logs documentation](./logs.md) to learn how to configure it.

## Access Logs

Access logs are a key part of observability in Traefik.

They are providing valuable insights about incoming traffic, and allow to monitor it.
The access logs record detailed information about each request received by Traefik,
including the source IP address, requested URL, response status code, and more.

Read the [Access Logs documentation](./access-logs.md) to learn how to configure it.

## Metrics

Traefik offers a metrics feature that provides valuable insights about the performance and usage.
These metrics include the number of requests received, the requests duration, and more.

Traefik supports these metrics systems: Prometheus, Datadog, InfluxDB 2.X, and StatsD.

Read the [Metrics documentation](./metrics/overview.md) to learn how to configure it.

## Tracing

The Traefik tracing system allows developers to gain deep visibility into the flow of requests through their infrastructure.

Traefik supports these tracing with OpenTelemetry.

Read the [Tracing documentation](./tracing/overview.md) to learn how to configure it.
