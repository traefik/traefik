# TODO -- Glossary

Where Every Technical Word finds its Definition`
{: .subtitle}

- [ ] Provider
    - [ ] Types of providers (KV, annotation based, label based, configuration based)
- [ ] Entrypoint
- [ ] Routers
- [ ] Middleware
- [ ] Service
- [ ] Static Configuration
- [ ] Dynamic Configuration
- [ ] ACME
- [ ] TraefikEE
- [ ] Tracing
- [ ] Metrics
- [ ] Orchestrator
- [ ] Key Value Store
- [ ] Logs
- [ ] Traefiker
- [ ] Traefik (How to pronounce)


# WIP

### Traefik Configuration

Traefik's configuration has two parts:

* The [static Traefik configuration](./#static-traefik-configuration).
* The [dynamic Traefik configuration](./#dynamic-traefik-configuration).


### Static Traefik Configuration

The static Traefik configuration is the global configuration which is setting up connections to configuration providers and entrypoints.  
The static Traefik configuration is loaded only at the beginning.

### Dynamic Traefik Configuration

The dynamic configuration concerns :

* HTTP
    * Routers
    * Middlewares
    * Services
* TCP
    * Routers
    * Services
* TLS (options and store)

The dynamic Traefik configuration can be hot-reloaded (no need to restart the process).
