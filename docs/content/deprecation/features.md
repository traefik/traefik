# Feature Deprecation Notices

This page is maintained and updated periodically to reflect our roadmap and any decisions around feature deprecation.

| Feature                                                       | Deprecated | End of Support | Removal |
|---------------------------------------------------------------|------------|----------------|---------|
| [Pilot Dashboard (Metrics)](#pilot-dashboard-metrics)         | 2.7        | 2.8            | 2.9     |
| [Pilot Plugins](#pilot-plugins)                               | 2.7        | 2.8            | 2.9     |
| [Consul Enterprise Namespaces](#consul-enterprise-namespaces) | 2.8        | TBD            | TBD     |

## Impact

### Pilot Dashboard (Metrics)

Metrics will continue to function normally up to 2.8, when they will be disabled.  
In 2.9, the Pilot platform and all Traefik integration code will be permanently removed.

### Pilot Plugins 

Starting on 2.7 the pilot token will not be a requirement anymore.  
At 2.9, a new plugin catalog home should be available, decoupled from pilot.

### Consul Enterprise Namespaces

Starting on 2.8 the `namespace` option of Consul and Consul Catalog providers is deprecated, 
please use the `namespaces` options instead.  
