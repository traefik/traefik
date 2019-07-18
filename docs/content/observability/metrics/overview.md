# Metrics
Metrics system
{: .subtitle }

Traefik supports 4 metrics backends:

- [DataDog](./datadog.md)
- [InfluxDB](./influxdb.md)
- [Prometheus](./prometheus.md)
- [StatsD](./statsd.md)

## Configuration

To enable metrics:

```toml tab="File (TOML)"
[metrics]
```

```yaml tab="File (TOML)"
metrics: {}
```

```bash tab="CLI"
--metrics
```
