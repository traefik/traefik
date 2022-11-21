# Feature Deprecation Notices

This page is maintained and updated periodically to reflect our roadmap and any decisions around feature deprecation.

| Feature                                                     | Deprecated | End of Support | Removal |
|-------------------------------------------------------------|------------|----------------|---------|
| [Consul Enterprise Namespace](#consul-enterprise-namespace) | 2.8        | N/A            | 3.0     |
| [TLS 1.0 and 1.1 Support](#tls-10-and-11)                   | N/A        | 2.8            | N/A     |
| [Nomad Namespace](#nomad-namespace)                         | 2.10       | N/A            | 3.0     |

## Impact

### Consul Enterprise Namespace

Starting on 2.8 the `namespace` option of Consul and Consul Catalog providers is deprecated, 
please use the `namespaces` options instead.  

### TLS 1.0 and 1.1

Starting on 2.8 the default TLS options will use the minimum version of TLS 1.2. Of course, it can still be overridden with custom configuration.  

### Nomad Namespace

Starting on 2.10 the `namespace` option of the Nomad provider is deprecated,
please use the `namespaces` options instead.  
