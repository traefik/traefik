# Feature Deprecation Notices

This page is maintained and updated periodically to reflect our roadmap and any decisions around feature deprecation.

| Feature                                                     | Deprecated | End of Support | Removal |
|-------------------------------------------------------------|------------|----------------|---------|
| [Pilot](#pilot)                                             | 2.7        | 2.8            | 2.9     |
| [Consul Enterprise Namespace](#consul-enterprise-namespace) | 2.8        | N/A            | 3.0     |
| [TLS 1.0 and 1.1 Support](#tls-10-and-11)                   | N/A        | 2.8            | N/A     |

## Impact

### Pilot

Metrics will continue to function normally up to 2.8, when they will be disabled.  
In 2.9, the Pilot platform and all Traefik integration code will be permanently removed.

Starting on 2.7 the pilot token will not be a requirement anymore for plugins.  
Since 2.8, a [new plugin catalog](https://plugins.traefik.io) is available, decoupled from Pilot.

### Consul Enterprise Namespace

Starting on 2.8 the `namespace` option of Consul and Consul Catalog providers is deprecated, 
please use the `namespaces` options instead.  

### TLS 1.0 and 1.1

Starting on 2.8 the default TLS options will use the minimum version of TLS 1.2. Of course, it can still be overridden with custom configuration.  
