# Metrics
Metrics system
{: .subtitle }

Traefik supports 4 metrics backends:

- [Datadog](./datadog.md)
- [InfluxDB](./influxdb.md)
- [Prometheus](./prometheus.md)
- [StatsD](./statsd.md)

## Configuration

To enable metrics:

```toml tab="File (TOML)"
[metrics]
```

```yaml tab="File (YAML)"
metrics: {}
```

```bash tab="CLI"
--metrics=true
```
