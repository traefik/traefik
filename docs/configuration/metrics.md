# Metrics Definition
```toml
# Metrics definition
#
# Default:
# [metrics]
#
[metrics]
```

## Prometheus

```toml
# To enable Traefik to export internal metrics to Prometheus
[metrics.prometheus]

# Buckets for latency metrics
#
# Optional
# Default: [0.1, 0.3, 1.2, 5]
buckets=[0.1,0.3,1.2,5.0]
    
# ...
```

### DataDog

```toml
# DataDog metrics exporter type
[metrics.datadog]

# DataDog's address.
#
# Required
# Default: "localhost:8125"
#
address = "localhost:8125"

# DataDog push interval
#
# Optional
# Default: "10s"
#
pushinterval = "10s"

# ...
```

### StatsD

```toml
# StatsD metrics exporter type
[metrics.statsd]

# StatD's address.
#
# Required
# Default: "localhost:8125"
#
address = "localhost:8125"

# StatD push interval
#
# Optional
# Default: "10s"
#
pushinterval = "10s"

# ...
```


## Statistics

```toml
[metrics]
# ...

# Enable more detailed statistics.
[metrics.statistics]

# Number of recent errors logged.
#
# Default: 10
#
recentErrors = 10

# ...
```
