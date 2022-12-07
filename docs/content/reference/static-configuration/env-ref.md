<!--
CODE GENERATED AUTOMATICALLY
THIS FILE MUST NOT BE EDITED BY HAND
-->

`TRAEFIK_ACCESSLOG`:  
Access log settings. (Default: ```false```)

`TRAEFIK_ACCESSLOG_BUFFERINGSIZE`:  
Number of access log lines to process in a buffered way. (Default: ```0```)

`TRAEFIK_ACCESSLOG_FIELDS_DEFAULTMODE`:  
Default mode for fields: keep | drop (Default: ```keep```)

`TRAEFIK_ACCESSLOG_FIELDS_HEADERS_DEFAULTMODE`:  
Default mode for fields: keep | drop | redact (Default: ```drop```)

`TRAEFIK_ACCESSLOG_FIELDS_HEADERS_NAMES_<NAME>`:  
Override mode for headers

`TRAEFIK_ACCESSLOG_FIELDS_NAMES_<NAME>`:  
Override mode for fields

`TRAEFIK_ACCESSLOG_FILEPATH`:  
Access log file path. Stdout is used when omitted or empty.

`TRAEFIK_ACCESSLOG_FILTERS_MINDURATION`:  
Keep access logs when request took longer than the specified duration. (Default: ```0```)

`TRAEFIK_ACCESSLOG_FILTERS_RETRYATTEMPTS`:  
Keep access logs when at least one retry happened. (Default: ```false```)

`TRAEFIK_ACCESSLOG_FILTERS_STATUSCODES`:  
Keep access logs with status codes in the specified range.

`TRAEFIK_ACCESSLOG_FORMAT`:  
Access log format: json | common (Default: ```common```)

`TRAEFIK_API`:  
Enable api/dashboard. (Default: ```false```)

`TRAEFIK_API_DASHBOARD`:  
Activate dashboard. (Default: ```true```)

`TRAEFIK_API_DEBUG`:  
Enable additional endpoints for debugging and profiling. (Default: ```false```)

`TRAEFIK_API_INSECURE`:  
Activate API directly on the entryPoint named traefik. (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>`:  
Certificates resolvers configuration. (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_CASERVER`:  
CA server to use. (Default: ```https://acme-v02.api.letsencrypt.org/directory```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_CERTIFICATESDURATION`:  
Certificates' duration in hours. (Default: ```2160```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_DNSCHALLENGE`:  
Activate DNS-01 Challenge. (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_DNSCHALLENGE_DELAYBEFORECHECK`:  
Assume DNS propagates after a delay in seconds rather than finding and querying nameservers. (Default: ```0```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_DNSCHALLENGE_DISABLEPROPAGATIONCHECK`:  
Disable the DNS propagation checks before notifying ACME that the DNS challenge is ready. [not recommended] (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_DNSCHALLENGE_PROVIDER`:  
Use a DNS-01 based challenge provider rather than HTTPS.

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_DNSCHALLENGE_RESOLVERS`:  
Use following DNS servers to resolve the FQDN authority.

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_EAB_HMACENCODED`:  
Base64 encoded HMAC key from External CA.

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_EAB_KID`:  
Key identifier from External CA.

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_EMAIL`:  
Email address used for registration.

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_HTTPCHALLENGE`:  
Activate HTTP-01 Challenge. (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_HTTPCHALLENGE_ENTRYPOINT`:  
HTTP challenge EntryPoint

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_KEYTYPE`:  
KeyType used for generating certificate private key. Allow value 'EC256', 'EC384', 'RSA2048', 'RSA4096', 'RSA8192'. (Default: ```RSA4096```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_PREFERREDCHAIN`:  
Preferred chain to use.

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_STORAGE`:  
Storage to use. (Default: ```acme.json```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_TLSCHALLENGE`:  
Activate TLS-ALPN-01 Challenge. (Default: ```true```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_TAILSCALE`:  
Enables Tailscale certificate resolution. (Default: ```true```)

`TRAEFIK_ENTRYPOINTS_<NAME>`:  
Entry points definition. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_ADDRESS`:  
Entry point address.

`TRAEFIK_ENTRYPOINTS_<NAME>_ASDEFAULT`:  
Adds this EntryPoint to the list of default EntryPoints to be used on routers that don't have any Entrypoint defined. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_FORWARDEDHEADERS_INSECURE`:  
Trust all forwarded headers. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_FORWARDEDHEADERS_TRUSTEDIPS`:  
Trust only forwarded headers from selected IPs.

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP`:  
HTTP configuration.

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP2_MAXCONCURRENTSTREAMS`:  
Specifies the number of concurrent streams per connection that each client is allowed to initiate. (Default: ```250```)

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP3`:  
HTTP/3 configuration. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP3_ADVERTISEDPORT`:  
UDP port to advertise, on which HTTP/3 is available. (Default: ```0```)

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_MIDDLEWARES`:  
Default middlewares for the routers linked to the entry point.

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_REDIRECTIONS_ENTRYPOINT_PERMANENT`:  
Applies a permanent redirection. (Default: ```true```)

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_REDIRECTIONS_ENTRYPOINT_PRIORITY`:  
Priority of the generated router. (Default: ```2147483646```)

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_REDIRECTIONS_ENTRYPOINT_SCHEME`:  
Scheme used for the redirection. (Default: ```https```)

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_REDIRECTIONS_ENTRYPOINT_TO`:  
Targeted entry point of the redirection.

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_TLS`:  
Default TLS configuration for the routers linked to the entry point. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_TLS_CERTRESOLVER`:  
Default certificate resolver for the routers linked to the entry point.

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_TLS_DOMAINS`:  
Default TLS domains for the routers linked to the entry point.

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_TLS_DOMAINS_n_MAIN`:  
Default subject name.

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_TLS_DOMAINS_n_SANS`:  
Subject alternative names.

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_TLS_OPTIONS`:  
Default TLS options for the routers linked to the entry point.

`TRAEFIK_ENTRYPOINTS_<NAME>_PROXYPROTOCOL`:  
Proxy-Protocol configuration. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_PROXYPROTOCOL_INSECURE`:  
Trust all. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_PROXYPROTOCOL_TRUSTEDIPS`:  
Trust only selected IPs.

`TRAEFIK_ENTRYPOINTS_<NAME>_TRANSPORT_LIFECYCLE_GRACETIMEOUT`:  
Duration to give active requests a chance to finish before Traefik stops. (Default: ```10```)

`TRAEFIK_ENTRYPOINTS_<NAME>_TRANSPORT_LIFECYCLE_REQUESTACCEPTGRACETIMEOUT`:  
Duration to keep accepting requests before Traefik initiates the graceful shutdown procedure. (Default: ```0```)

`TRAEFIK_ENTRYPOINTS_<NAME>_TRANSPORT_RESPONDINGTIMEOUTS_IDLETIMEOUT`:  
IdleTimeout is the maximum amount duration an idle (keep-alive) connection will remain idle before closing itself. If zero, no timeout is set. (Default: ```180```)

`TRAEFIK_ENTRYPOINTS_<NAME>_TRANSPORT_RESPONDINGTIMEOUTS_READTIMEOUT`:  
ReadTimeout is the maximum duration for reading the entire request, including the body. If zero, no timeout is set. (Default: ```0```)

`TRAEFIK_ENTRYPOINTS_<NAME>_TRANSPORT_RESPONDINGTIMEOUTS_WRITETIMEOUT`:  
WriteTimeout is the maximum duration before timing out writes of the response. If zero, no timeout is set. (Default: ```0```)

`TRAEFIK_ENTRYPOINTS_<NAME>_UDP_TIMEOUT`:  
Timeout defines how long to wait on an idle session before releasing the related resources. (Default: ```3```)

`TRAEFIK_EXPERIMENTAL_HUB`:  
Enable the Traefik Hub provider. (Default: ```false```)

`TRAEFIK_EXPERIMENTAL_KUBERNETESGATEWAY`:  
Allow the Kubernetes gateway api provider usage. (Default: ```false```)

`TRAEFIK_EXPERIMENTAL_LOCALPLUGINS_<NAME>`:  
Local plugins configuration. (Default: ```false```)

`TRAEFIK_EXPERIMENTAL_LOCALPLUGINS_<NAME>_MODULENAME`:  
plugin's module name.

`TRAEFIK_EXPERIMENTAL_PLUGINS_<NAME>_MODULENAME`:  
plugin's module name.

`TRAEFIK_EXPERIMENTAL_PLUGINS_<NAME>_VERSION`:  
plugin's version.

`TRAEFIK_GLOBAL_CHECKNEWVERSION`:  
Periodically check if a new version has been released. (Default: ```true```)

`TRAEFIK_GLOBAL_SENDANONYMOUSUSAGE`:  
Periodically send anonymous usage statistics. If the option is not specified, it will be enabled by default. (Default: ```false```)

`TRAEFIK_HOSTRESOLVER`:  
Enable CNAME Flattening. (Default: ```false```)

`TRAEFIK_HOSTRESOLVER_CNAMEFLATTENING`:  
A flag to enable/disable CNAME flattening (Default: ```false```)

`TRAEFIK_HOSTRESOLVER_RESOLVCONFIG`:  
resolv.conf used for DNS resolving (Default: ```/etc/resolv.conf```)

`TRAEFIK_HOSTRESOLVER_RESOLVDEPTH`:  
The maximal depth of DNS recursive resolving (Default: ```5```)

`TRAEFIK_HUB`:  
Traefik Hub configuration. (Default: ```false```)

`TRAEFIK_HUB_TLS_CA`:  
The certificate authority authenticates the Traefik Hub Agent certificate.

`TRAEFIK_HUB_TLS_CERT`:  
The TLS certificate for Traefik Proxy as a TLS client.

`TRAEFIK_HUB_TLS_INSECURE`:  
Enables an insecure TLS connection that uses default credentials, and which has no peer authentication between Traefik Proxy and the Traefik Hub Agent. (Default: ```false```)

`TRAEFIK_HUB_TLS_KEY`:  
The TLS key for Traefik Proxy as a TLS client.

`TRAEFIK_LOG`:  
Traefik log settings. (Default: ```false```)

`TRAEFIK_LOG_COMPRESS`:  
Determines if the rotated log files should be compressed using gzip. (Default: ```false```)

`TRAEFIK_LOG_FILEPATH`:  
Traefik log file path. Stdout is used when omitted or empty.

`TRAEFIK_LOG_FORMAT`:  
Traefik log format: json | common (Default: ```common```)

`TRAEFIK_LOG_LEVEL`:  
Log level set to traefik logs. (Default: ```ERROR```)

`TRAEFIK_LOG_MAXAGE`:  
Maximum number of days to retain old log files based on the timestamp encoded in their filename. (Default: ```0```)

`TRAEFIK_LOG_MAXBACKUPS`:  
Maximum number of old log files to retain. (Default: ```0```)

`TRAEFIK_LOG_MAXSIZE`:  
Maximum size in megabytes of the log file before it gets rotated. (Default: ```0```)

`TRAEFIK_LOG_NOCOLOR`:  
When using the 'common' format, disables the colorized output. (Default: ```false```)

`TRAEFIK_METRICS_DATADOG`:  
Datadog metrics exporter type. (Default: ```false```)

`TRAEFIK_METRICS_DATADOG_ADDENTRYPOINTSLABELS`:  
Enable metrics on entry points. (Default: ```true```)

`TRAEFIK_METRICS_DATADOG_ADDRESS`:  
Datadog's address. (Default: ```localhost:8125```)

`TRAEFIK_METRICS_DATADOG_ADDROUTERSLABELS`:  
Enable metrics on routers. (Default: ```false```)

`TRAEFIK_METRICS_DATADOG_ADDSERVICESLABELS`:  
Enable metrics on services. (Default: ```true```)

`TRAEFIK_METRICS_DATADOG_PREFIX`:  
Prefix to use for metrics collection. (Default: ```traefik```)

`TRAEFIK_METRICS_DATADOG_PUSHINTERVAL`:  
Datadog push interval. (Default: ```10```)

`TRAEFIK_METRICS_INFLUXDB`:  
InfluxDB metrics exporter type. (Default: ```false```)

`TRAEFIK_METRICS_INFLUXDB2`:  
InfluxDB v2 metrics exporter type. (Default: ```false```)

`TRAEFIK_METRICS_INFLUXDB2_ADDENTRYPOINTSLABELS`:  
Enable metrics on entry points. (Default: ```true```)

`TRAEFIK_METRICS_INFLUXDB2_ADDITIONALLABELS_<NAME>`:  
Additional labels (influxdb tags) on all metrics

`TRAEFIK_METRICS_INFLUXDB2_ADDRESS`:  
InfluxDB v2 address. (Default: ```http://localhost:8086```)

`TRAEFIK_METRICS_INFLUXDB2_ADDROUTERSLABELS`:  
Enable metrics on routers. (Default: ```false```)

`TRAEFIK_METRICS_INFLUXDB2_ADDSERVICESLABELS`:  
Enable metrics on services. (Default: ```true```)

`TRAEFIK_METRICS_INFLUXDB2_BUCKET`:  
InfluxDB v2 bucket ID.

`TRAEFIK_METRICS_INFLUXDB2_ORG`:  
InfluxDB v2 org ID.

`TRAEFIK_METRICS_INFLUXDB2_PUSHINTERVAL`:  
InfluxDB v2 push interval. (Default: ```10```)

`TRAEFIK_METRICS_INFLUXDB2_TOKEN`:  
InfluxDB v2 access token.

`TRAEFIK_METRICS_INFLUXDB_ADDENTRYPOINTSLABELS`:  
Enable metrics on entry points. (Default: ```true```)

`TRAEFIK_METRICS_INFLUXDB_ADDITIONALLABELS_<NAME>`:  
Additional labels (influxdb tags) on all metrics

`TRAEFIK_METRICS_INFLUXDB_ADDRESS`:  
InfluxDB address. (Default: ```localhost:8089```)

`TRAEFIK_METRICS_INFLUXDB_ADDROUTERSLABELS`:  
Enable metrics on routers. (Default: ```false```)

`TRAEFIK_METRICS_INFLUXDB_ADDSERVICESLABELS`:  
Enable metrics on services. (Default: ```true```)

`TRAEFIK_METRICS_INFLUXDB_DATABASE`:  
InfluxDB database used when protocol is http.

`TRAEFIK_METRICS_INFLUXDB_PASSWORD`:  
InfluxDB password (only with http).

`TRAEFIK_METRICS_INFLUXDB_PROTOCOL`:  
InfluxDB address protocol (udp or http). (Default: ```udp```)

`TRAEFIK_METRICS_INFLUXDB_PUSHINTERVAL`:  
InfluxDB push interval. (Default: ```10```)

`TRAEFIK_METRICS_INFLUXDB_RETENTIONPOLICY`:  
InfluxDB retention policy used when protocol is http.

`TRAEFIK_METRICS_INFLUXDB_USERNAME`:  
InfluxDB username (only with http).

`TRAEFIK_METRICS_OPENTELEMETRY`:  
OpenTelemetry metrics exporter type. (Default: ```false```)

`TRAEFIK_METRICS_OPENTELEMETRY_ADDENTRYPOINTSLABELS`:  
Enable metrics on entry points. (Default: ```true```)

`TRAEFIK_METRICS_OPENTELEMETRY_ADDRESS`:  
Address (host:port) of the collector endpoint. (Default: ```localhost:4318```)

`TRAEFIK_METRICS_OPENTELEMETRY_ADDROUTERSLABELS`:  
Enable metrics on routers. (Default: ```false```)

`TRAEFIK_METRICS_OPENTELEMETRY_ADDSERVICESLABELS`:  
Enable metrics on services. (Default: ```true```)

`TRAEFIK_METRICS_OPENTELEMETRY_EXPLICITBOUNDARIES`:  
Boundaries for latency metrics. (Default: ```0.005000, 0.010000, 0.025000, 0.050000, 0.100000, 0.250000, 0.500000, 1.000000, 2.500000, 5.000000, 10.000000```)

`TRAEFIK_METRICS_OPENTELEMETRY_GRPC`:  
gRPC specific configuration for the OpenTelemetry collector. (Default: ```true```)

`TRAEFIK_METRICS_OPENTELEMETRY_HEADERS_<NAME>`:  
Headers sent with payload.

`TRAEFIK_METRICS_OPENTELEMETRY_INSECURE`:  
Disables client transport security for the exporter. (Default: ```false```)

`TRAEFIK_METRICS_OPENTELEMETRY_PATH`:  
Set the URL path of the collector endpoint.

`TRAEFIK_METRICS_OPENTELEMETRY_PUSHINTERVAL`:  
Period between calls to collect a checkpoint. (Default: ```10```)

`TRAEFIK_METRICS_OPENTELEMETRY_TLS_CA`:  
TLS CA

`TRAEFIK_METRICS_OPENTELEMETRY_TLS_CERT`:  
TLS cert

`TRAEFIK_METRICS_OPENTELEMETRY_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_METRICS_OPENTELEMETRY_TLS_KEY`:  
TLS key

`TRAEFIK_METRICS_PROMETHEUS`:  
Prometheus metrics exporter type. (Default: ```false```)

`TRAEFIK_METRICS_PROMETHEUS_ADDENTRYPOINTSLABELS`:  
Enable metrics on entry points. (Default: ```true```)

`TRAEFIK_METRICS_PROMETHEUS_ADDROUTERSLABELS`:  
Enable metrics on routers. (Default: ```false```)

`TRAEFIK_METRICS_PROMETHEUS_ADDSERVICESLABELS`:  
Enable metrics on services. (Default: ```true```)

`TRAEFIK_METRICS_PROMETHEUS_BUCKETS`:  
Buckets for latency metrics. (Default: ```0.100000, 0.300000, 1.200000, 5.000000```)

`TRAEFIK_METRICS_PROMETHEUS_ENTRYPOINT`:  
EntryPoint (Default: ```traefik```)

`TRAEFIK_METRICS_PROMETHEUS_MANUALROUTING`:  
Manual routing (Default: ```false```)

`TRAEFIK_METRICS_STATSD`:  
StatsD metrics exporter type. (Default: ```false```)

`TRAEFIK_METRICS_STATSD_ADDENTRYPOINTSLABELS`:  
Enable metrics on entry points. (Default: ```true```)

`TRAEFIK_METRICS_STATSD_ADDRESS`:  
StatsD address. (Default: ```localhost:8125```)

`TRAEFIK_METRICS_STATSD_ADDROUTERSLABELS`:  
Enable metrics on routers. (Default: ```false```)

`TRAEFIK_METRICS_STATSD_ADDSERVICESLABELS`:  
Enable metrics on services. (Default: ```true```)

`TRAEFIK_METRICS_STATSD_PREFIX`:  
Prefix to use for metrics collection. (Default: ```traefik```)

`TRAEFIK_METRICS_STATSD_PUSHINTERVAL`:  
StatsD push interval. (Default: ```10```)

`TRAEFIK_PING`:  
Enable ping. (Default: ```false```)

`TRAEFIK_PING_ENTRYPOINT`:  
EntryPoint (Default: ```traefik```)

`TRAEFIK_PING_MANUALROUTING`:  
Manual routing (Default: ```false```)

`TRAEFIK_PING_TERMINATINGSTATUSCODE`:  
Terminating status code (Default: ```503```)

`TRAEFIK_PROVIDERS_CONSUL`:  
Enable Consul backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_CONSULCATALOG`:  
Enable ConsulCatalog backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_CONSULCATALOG_CACHE`:  
Use local agent caching for catalog reads. (Default: ```false```)

`TRAEFIK_PROVIDERS_CONSULCATALOG_CONNECTAWARE`:  
Enable Consul Connect support. (Default: ```false```)

`TRAEFIK_PROVIDERS_CONSULCATALOG_CONNECTBYDEFAULT`:  
Consider every service as Connect capable by default. (Default: ```false```)

`TRAEFIK_PROVIDERS_CONSULCATALOG_CONSTRAINTS`:  
Constraints is an expression that Traefik matches against the container's labels to determine whether to create any route for that container.

`TRAEFIK_PROVIDERS_CONSULCATALOG_DEFAULTRULE`:  
Default rule. (Default: ```Host(`{{ normalize .Name }}`)```)

`TRAEFIK_PROVIDERS_CONSULCATALOG_ENDPOINT_ADDRESS`:  
The address of the Consul server

`TRAEFIK_PROVIDERS_CONSULCATALOG_ENDPOINT_DATACENTER`:  
Data center to use. If not provided, the default agent data center is used

`TRAEFIK_PROVIDERS_CONSULCATALOG_ENDPOINT_ENDPOINTWAITTIME`:  
WaitTime limits how long a Watch will block. If not provided, the agent default values will be used (Default: ```0```)

`TRAEFIK_PROVIDERS_CONSULCATALOG_ENDPOINT_HTTPAUTH_PASSWORD`:  
Basic Auth password

`TRAEFIK_PROVIDERS_CONSULCATALOG_ENDPOINT_HTTPAUTH_USERNAME`:  
Basic Auth username

`TRAEFIK_PROVIDERS_CONSULCATALOG_ENDPOINT_SCHEME`:  
The URI scheme for the Consul server

`TRAEFIK_PROVIDERS_CONSULCATALOG_ENDPOINT_TLS_CA`:  
TLS CA

`TRAEFIK_PROVIDERS_CONSULCATALOG_ENDPOINT_TLS_CERT`:  
TLS cert

`TRAEFIK_PROVIDERS_CONSULCATALOG_ENDPOINT_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_PROVIDERS_CONSULCATALOG_ENDPOINT_TLS_KEY`:  
TLS key

`TRAEFIK_PROVIDERS_CONSULCATALOG_ENDPOINT_TOKEN`:  
Token is used to provide a per-request ACL token which overrides the agent's default token

`TRAEFIK_PROVIDERS_CONSULCATALOG_EXPOSEDBYDEFAULT`:  
Expose containers by default. (Default: ```true```)

`TRAEFIK_PROVIDERS_CONSULCATALOG_NAMESPACES`:  
Sets the namespaces used to discover services (Consul Enterprise only).

`TRAEFIK_PROVIDERS_CONSULCATALOG_PREFIX`:  
Prefix for consul service tags. (Default: ```traefik```)

`TRAEFIK_PROVIDERS_CONSULCATALOG_REFRESHINTERVAL`:  
Interval for check Consul API. (Default: ```15```)

`TRAEFIK_PROVIDERS_CONSULCATALOG_REQUIRECONSISTENT`:  
Forces the read to be fully consistent. (Default: ```false```)

`TRAEFIK_PROVIDERS_CONSULCATALOG_SERVICENAME`:  
Name of the Traefik service in Consul Catalog (needs to be registered via the orchestrator or manually). (Default: ```traefik```)

`TRAEFIK_PROVIDERS_CONSULCATALOG_STALE`:  
Use stale consistency for catalog reads. (Default: ```false```)

`TRAEFIK_PROVIDERS_CONSULCATALOG_WATCH`:  
Watch Consul API events. (Default: ```false```)

`TRAEFIK_PROVIDERS_CONSUL_ENDPOINTS`:  
KV store endpoints. (Default: ```127.0.0.1:8500```)

`TRAEFIK_PROVIDERS_CONSUL_NAMESPACES`:  
Sets the namespaces used to discover the configuration (Consul Enterprise only).

`TRAEFIK_PROVIDERS_CONSUL_ROOTKEY`:  
Root key used for KV store. (Default: ```traefik```)

`TRAEFIK_PROVIDERS_CONSUL_TLS_CA`:  
TLS CA

`TRAEFIK_PROVIDERS_CONSUL_TLS_CERT`:  
TLS cert

`TRAEFIK_PROVIDERS_CONSUL_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_PROVIDERS_CONSUL_TLS_KEY`:  
TLS key

`TRAEFIK_PROVIDERS_CONSUL_TOKEN`:  
Per-request ACL token.

`TRAEFIK_PROVIDERS_DOCKER`:  
Enable Docker backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_DOCKER_ALLOWEMPTYSERVICES`:  
Disregards the Docker containers health checks with respect to the creation or removal of the corresponding services. (Default: ```false```)

`TRAEFIK_PROVIDERS_DOCKER_CONSTRAINTS`:  
Constraints is an expression that Traefik matches against the container's labels to determine whether to create any route for that container.

`TRAEFIK_PROVIDERS_DOCKER_DEFAULTRULE`:  
Default rule. (Default: ```Host(`{{ normalize .Name }}`)```)

`TRAEFIK_PROVIDERS_DOCKER_ENDPOINT`:  
Docker server endpoint. Can be a tcp or a unix socket endpoint. (Default: ```unix:///var/run/docker.sock```)

`TRAEFIK_PROVIDERS_DOCKER_EXPOSEDBYDEFAULT`:  
Expose containers by default. (Default: ```true```)

`TRAEFIK_PROVIDERS_DOCKER_HTTPCLIENTTIMEOUT`:  
Client timeout for HTTP connections. (Default: ```0```)

`TRAEFIK_PROVIDERS_DOCKER_NETWORK`:  
Default Docker network used.

`TRAEFIK_PROVIDERS_DOCKER_SWARMMODE`:  
Use Docker on Swarm Mode. (Default: ```false```)

`TRAEFIK_PROVIDERS_DOCKER_SWARMMODEREFRESHSECONDS`:  
Polling interval for swarm mode. (Default: ```15```)

`TRAEFIK_PROVIDERS_DOCKER_TLS_CA`:  
TLS CA

`TRAEFIK_PROVIDERS_DOCKER_TLS_CERT`:  
TLS cert

`TRAEFIK_PROVIDERS_DOCKER_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_PROVIDERS_DOCKER_TLS_KEY`:  
TLS key

`TRAEFIK_PROVIDERS_DOCKER_USEBINDPORTIP`:  
Use the ip address from the bound port, rather than from the inner network. (Default: ```false```)

`TRAEFIK_PROVIDERS_DOCKER_WATCH`:  
Watch Docker events. (Default: ```true```)

`TRAEFIK_PROVIDERS_ECS`:  
Enable AWS ECS backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_ECS_ACCESSKEYID`:  
AWS credentials access key ID to use for making requests.

`TRAEFIK_PROVIDERS_ECS_AUTODISCOVERCLUSTERS`:  
Auto discover cluster. (Default: ```false```)

`TRAEFIK_PROVIDERS_ECS_CLUSTERS`:  
ECS Cluster names. (Default: ```default```)

`TRAEFIK_PROVIDERS_ECS_CONSTRAINTS`:  
Constraints is an expression that Traefik matches against the container's labels to determine whether to create any route for that container.

`TRAEFIK_PROVIDERS_ECS_DEFAULTRULE`:  
Default rule. (Default: ```Host(`{{ normalize .Name }}`)```)

`TRAEFIK_PROVIDERS_ECS_ECSANYWHERE`:  
Enable ECS Anywhere support. (Default: ```false```)

`TRAEFIK_PROVIDERS_ECS_EXPOSEDBYDEFAULT`:  
Expose services by default. (Default: ```true```)

`TRAEFIK_PROVIDERS_ECS_HEALTHYTASKSONLY`:  
Determines whether to discover only healthy tasks. (Default: ```false```)

`TRAEFIK_PROVIDERS_ECS_REFRESHSECONDS`:  
Polling interval (in seconds). (Default: ```15```)

`TRAEFIK_PROVIDERS_ECS_REGION`:  
AWS region to use for requests.

`TRAEFIK_PROVIDERS_ECS_SECRETACCESSKEY`:  
AWS credentials access key to use for making requests.

`TRAEFIK_PROVIDERS_ETCD`:  
Enable Etcd backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_ETCD_ENDPOINTS`:  
KV store endpoints. (Default: ```127.0.0.1:2379```)

`TRAEFIK_PROVIDERS_ETCD_PASSWORD`:  
Password for authentication.

`TRAEFIK_PROVIDERS_ETCD_ROOTKEY`:  
Root key used for KV store. (Default: ```traefik```)

`TRAEFIK_PROVIDERS_ETCD_TLS_CA`:  
TLS CA

`TRAEFIK_PROVIDERS_ETCD_TLS_CERT`:  
TLS cert

`TRAEFIK_PROVIDERS_ETCD_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_PROVIDERS_ETCD_TLS_KEY`:  
TLS key

`TRAEFIK_PROVIDERS_ETCD_USERNAME`:  
Username for authentication.

`TRAEFIK_PROVIDERS_FILE_DEBUGLOGGENERATEDTEMPLATE`:  
Enable debug logging of generated configuration template. (Default: ```false```)

`TRAEFIK_PROVIDERS_FILE_DIRECTORY`:  
Load dynamic configuration from one or more .yml or .toml files in a directory.

`TRAEFIK_PROVIDERS_FILE_FILENAME`:  
Load dynamic configuration from a file.

`TRAEFIK_PROVIDERS_FILE_WATCH`:  
Watch provider. (Default: ```true```)

`TRAEFIK_PROVIDERS_HTTP`:  
Enable HTTP backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_HTTP_ENDPOINT`:  
Load configuration from this endpoint.

`TRAEFIK_PROVIDERS_HTTP_HEADERS_<NAME>`:  
Define custom headers to be sent to the endpoint.

`TRAEFIK_PROVIDERS_HTTP_POLLINTERVAL`:  
Polling interval for endpoint. (Default: ```5```)

`TRAEFIK_PROVIDERS_HTTP_POLLTIMEOUT`:  
Polling timeout for endpoint. (Default: ```5```)

`TRAEFIK_PROVIDERS_HTTP_TLS_CA`:  
TLS CA

`TRAEFIK_PROVIDERS_HTTP_TLS_CERT`:  
TLS cert

`TRAEFIK_PROVIDERS_HTTP_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_PROVIDERS_HTTP_TLS_KEY`:  
TLS key

`TRAEFIK_PROVIDERS_KUBERNETESCRD`:  
Enable Kubernetes backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESCRD_ALLOWCROSSNAMESPACE`:  
Allow cross namespace resource reference. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESCRD_ALLOWEMPTYSERVICES`:  
Allow the creation of services without endpoints. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESCRD_ALLOWEXTERNALNAMESERVICES`:  
Allow ExternalName services. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESCRD_CERTAUTHFILEPATH`:  
Kubernetes certificate authority file path (not needed for in-cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESCRD_ENDPOINT`:  
Kubernetes server endpoint (required for external cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESCRD_INGRESSCLASS`:  
Value of kubernetes.io/ingress.class annotation to watch for.

`TRAEFIK_PROVIDERS_KUBERNETESCRD_LABELSELECTOR`:  
Kubernetes label selector to use.

`TRAEFIK_PROVIDERS_KUBERNETESCRD_NAMESPACES`:  
Kubernetes namespaces.

`TRAEFIK_PROVIDERS_KUBERNETESCRD_THROTTLEDURATION`:  
Ingress refresh throttle duration (Default: ```0```)

`TRAEFIK_PROVIDERS_KUBERNETESCRD_TOKEN`:  
Kubernetes bearer token (not needed for in-cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY`:  
Enable Kubernetes gateway api provider with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_CERTAUTHFILEPATH`:  
Kubernetes certificate authority file path (not needed for in-cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_ENDPOINT`:  
Kubernetes server endpoint (required for external cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_LABELSELECTOR`:  
Kubernetes label selector to select specific GatewayClasses.

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_NAMESPACES`:  
Kubernetes namespaces.

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_THROTTLEDURATION`:  
Kubernetes refresh throttle duration (Default: ```0```)

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_TOKEN`:  
Kubernetes bearer token (not needed for in-cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS`:  
Enable Kubernetes backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_ALLOWEMPTYSERVICES`:  
Allow creation of services without endpoints. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_ALLOWEXTERNALNAMESERVICES`:  
Allow ExternalName services. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_CERTAUTHFILEPATH`:  
Kubernetes certificate authority file path (not needed for in-cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_ENDPOINT`:  
Kubernetes server endpoint (required for external cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_INGRESSCLASS`:  
Value of kubernetes.io/ingress.class annotation or IngressClass name to watch for.

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_INGRESSENDPOINT_HOSTNAME`:  
Hostname used for Kubernetes Ingress endpoints.

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_INGRESSENDPOINT_IP`:  
IP used for Kubernetes Ingress endpoints.

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_INGRESSENDPOINT_PUBLISHEDSERVICE`:  
Published Kubernetes Service to copy status from.

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_LABELSELECTOR`:  
Kubernetes Ingress label selector to use.

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_NAMESPACES`:  
Kubernetes namespaces.

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_THROTTLEDURATION`:  
Ingress refresh throttle duration (Default: ```0```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_TOKEN`:  
Kubernetes bearer token (not needed for in-cluster client).

`TRAEFIK_PROVIDERS_MARATHON`:  
Enable Marathon backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_MARATHON_BASIC_HTTPBASICAUTHUSER`:  
Basic authentication User.

`TRAEFIK_PROVIDERS_MARATHON_BASIC_HTTPBASICPASSWORD`:  
Basic authentication Password.

`TRAEFIK_PROVIDERS_MARATHON_CONSTRAINTS`:  
Constraints is an expression that Traefik matches against the application's labels to determine whether to create any route for that application.

`TRAEFIK_PROVIDERS_MARATHON_DCOSTOKEN`:  
DCOSToken for DCOS environment, This will override the Authorization header.

`TRAEFIK_PROVIDERS_MARATHON_DEFAULTRULE`:  
Default rule. (Default: ```Host(`{{ normalize .Name }}`)```)

`TRAEFIK_PROVIDERS_MARATHON_DIALERTIMEOUT`:  
Set a dialer timeout for Marathon. (Default: ```5```)

`TRAEFIK_PROVIDERS_MARATHON_ENDPOINT`:  
Marathon server endpoint. You can also specify multiple endpoint for Marathon. (Default: ```http://127.0.0.1:8080```)

`TRAEFIK_PROVIDERS_MARATHON_EXPOSEDBYDEFAULT`:  
Expose Marathon apps by default. (Default: ```true```)

`TRAEFIK_PROVIDERS_MARATHON_FORCETASKHOSTNAME`:  
Force to use the task's hostname. (Default: ```false```)

`TRAEFIK_PROVIDERS_MARATHON_KEEPALIVE`:  
Set a TCP Keep Alive time. (Default: ```10```)

`TRAEFIK_PROVIDERS_MARATHON_RESPECTREADINESSCHECKS`:  
Filter out tasks with non-successful readiness checks during deployments. (Default: ```false```)

`TRAEFIK_PROVIDERS_MARATHON_RESPONSEHEADERTIMEOUT`:  
Set a response header timeout for Marathon. (Default: ```60```)

`TRAEFIK_PROVIDERS_MARATHON_TLSHANDSHAKETIMEOUT`:  
Set a TLS handshake timeout for Marathon. (Default: ```5```)

`TRAEFIK_PROVIDERS_MARATHON_TLS_CA`:  
TLS CA

`TRAEFIK_PROVIDERS_MARATHON_TLS_CERT`:  
TLS cert

`TRAEFIK_PROVIDERS_MARATHON_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_PROVIDERS_MARATHON_TLS_KEY`:  
TLS key

`TRAEFIK_PROVIDERS_MARATHON_TRACE`:  
Display additional provider logs. (Default: ```false```)

`TRAEFIK_PROVIDERS_MARATHON_WATCH`:  
Watch provider. (Default: ```true```)

`TRAEFIK_PROVIDERS_NOMAD`:  
Enable Nomad backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_NOMAD_CONSTRAINTS`:  
Constraints is an expression that Traefik matches against the Nomad service's tags to determine whether to create route(s) for that service.

`TRAEFIK_PROVIDERS_NOMAD_DEFAULTRULE`:  
Default rule. (Default: ```Host(`{{ normalize .Name }}`)```)

`TRAEFIK_PROVIDERS_NOMAD_ENDPOINT_ADDRESS`:  
The address of the Nomad server, including scheme and port.

`TRAEFIK_PROVIDERS_NOMAD_ENDPOINT_ENDPOINTWAITTIME`:  
WaitTime limits how long a Watch will block. If not provided, the agent default values will be used (Default: ```0```)

`TRAEFIK_PROVIDERS_NOMAD_ENDPOINT_REGION`:  
Nomad region to use. If not provided, the local agent region is used.

`TRAEFIK_PROVIDERS_NOMAD_ENDPOINT_TLS_CA`:  
TLS CA

`TRAEFIK_PROVIDERS_NOMAD_ENDPOINT_TLS_CERT`:  
TLS cert

`TRAEFIK_PROVIDERS_NOMAD_ENDPOINT_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_PROVIDERS_NOMAD_ENDPOINT_TLS_KEY`:  
TLS key

`TRAEFIK_PROVIDERS_NOMAD_ENDPOINT_TOKEN`:  
Token is used to provide a per-request ACL token.

`TRAEFIK_PROVIDERS_NOMAD_EXPOSEDBYDEFAULT`:  
Expose Nomad services by default. (Default: ```true```)

`TRAEFIK_PROVIDERS_NOMAD_NAMESPACES`:  
Sets the Nomad namespaces used to discover services.

`TRAEFIK_PROVIDERS_NOMAD_PREFIX`:  
Prefix for nomad service tags. (Default: ```traefik```)

`TRAEFIK_PROVIDERS_NOMAD_REFRESHINTERVAL`:  
Interval for polling Nomad API. (Default: ```15```)

`TRAEFIK_PROVIDERS_NOMAD_STALE`:  
Use stale consistency for catalog reads. (Default: ```false```)

`TRAEFIK_PROVIDERS_PLUGIN_<NAME>`:  
Plugins configuration.

`TRAEFIK_PROVIDERS_PROVIDERSTHROTTLEDURATION`:  
Backends throttle duration: minimum duration between 2 events from providers before applying a new configuration. It avoids unnecessary reloads if multiples events are sent in a short amount of time. (Default: ```2```)

`TRAEFIK_PROVIDERS_RANCHER`:  
Enable Rancher backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_RANCHER_CONSTRAINTS`:  
Constraints is an expression that Traefik matches against the container's labels to determine whether to create any route for that container.

`TRAEFIK_PROVIDERS_RANCHER_DEFAULTRULE`:  
Default rule. (Default: ```Host(`{{ normalize .Name }}`)```)

`TRAEFIK_PROVIDERS_RANCHER_ENABLESERVICEHEALTHFILTER`:  
Filter services with unhealthy states and inactive states. (Default: ```true```)

`TRAEFIK_PROVIDERS_RANCHER_EXPOSEDBYDEFAULT`:  
Expose containers by default. (Default: ```true```)

`TRAEFIK_PROVIDERS_RANCHER_INTERVALPOLL`:  
Poll the Rancher metadata service every 'rancher.refreshseconds' (less accurate). (Default: ```false```)

`TRAEFIK_PROVIDERS_RANCHER_PREFIX`:  
Prefix used for accessing the Rancher metadata service. (Default: ```latest```)

`TRAEFIK_PROVIDERS_RANCHER_REFRESHSECONDS`:  
Defines the polling interval in seconds. (Default: ```15```)

`TRAEFIK_PROVIDERS_RANCHER_WATCH`:  
Watch provider. (Default: ```true```)

`TRAEFIK_PROVIDERS_REDIS`:  
Enable Redis backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_REDIS_DB`:  
Database to be selected after connecting to the server. (Default: ```0```)

`TRAEFIK_PROVIDERS_REDIS_ENDPOINTS`:  
KV store endpoints. (Default: ```127.0.0.1:6379```)

`TRAEFIK_PROVIDERS_REDIS_PASSWORD`:  
Password for authentication.

`TRAEFIK_PROVIDERS_REDIS_ROOTKEY`:  
Root key used for KV store. (Default: ```traefik```)

`TRAEFIK_PROVIDERS_REDIS_TLS_CA`:  
TLS CA

`TRAEFIK_PROVIDERS_REDIS_TLS_CERT`:  
TLS cert

`TRAEFIK_PROVIDERS_REDIS_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_PROVIDERS_REDIS_TLS_KEY`:  
TLS key

`TRAEFIK_PROVIDERS_REDIS_USERNAME`:  
Username for authentication.

`TRAEFIK_PROVIDERS_REST`:  
Enable Rest backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_REST_INSECURE`:  
Activate REST Provider directly on the entryPoint named traefik. (Default: ```false```)

`TRAEFIK_PROVIDERS_ZOOKEEPER`:  
Enable ZooKeeper backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_ZOOKEEPER_ENDPOINTS`:  
KV store endpoints. (Default: ```127.0.0.1:2181```)

`TRAEFIK_PROVIDERS_ZOOKEEPER_PASSWORD`:  
Password for authentication.

`TRAEFIK_PROVIDERS_ZOOKEEPER_ROOTKEY`:  
Root key used for KV store. (Default: ```traefik```)

`TRAEFIK_PROVIDERS_ZOOKEEPER_USERNAME`:  
Username for authentication.

`TRAEFIK_SERVERSTRANSPORT_FORWARDINGTIMEOUTS_DIALTIMEOUT`:  
The amount of time to wait until a connection to a backend server can be established. If zero, no timeout exists. (Default: ```30```)

`TRAEFIK_SERVERSTRANSPORT_FORWARDINGTIMEOUTS_IDLECONNTIMEOUT`:  
The maximum period for which an idle HTTP keep-alive connection will remain open before closing itself (Default: ```90```)

`TRAEFIK_SERVERSTRANSPORT_FORWARDINGTIMEOUTS_RESPONSEHEADERTIMEOUT`:  
The amount of time to wait for a server's response headers after fully writing the request (including its body, if any). If zero, no timeout exists. (Default: ```0```)

`TRAEFIK_SERVERSTRANSPORT_INSECURESKIPVERIFY`:  
Disable SSL certificate verification. (Default: ```false```)

`TRAEFIK_SERVERSTRANSPORT_MAXIDLECONNSPERHOST`:  
If non-zero, controls the maximum idle (keep-alive) to keep per-host. If zero, DefaultMaxIdleConnsPerHost is used (Default: ```200```)

`TRAEFIK_SERVERSTRANSPORT_ROOTCAS`:  
Add cert file for self-signed certificate.

`TRAEFIK_SERVERSTRANSPORT_SPIFFE`:  
Defines the SPIFFE configuration. (Default: ```false```)

`TRAEFIK_SERVERSTRANSPORT_SPIFFE_IDS`:  
Defines the allowed SPIFFE IDs (takes precedence over the SPIFFE TrustDomain).

`TRAEFIK_SERVERSTRANSPORT_SPIFFE_TRUSTDOMAIN`:  
Defines the allowed SPIFFE trust domain.

`TRAEFIK_SPIFFE_WORKLOADAPIADDR`:  
Defines the workload API address.

`TRAEFIK_TRACING`:  
OpenTracing configuration. (Default: ```false```)

`TRAEFIK_TRACING_DATADOG`:  
Settings for Datadog. (Default: ```false```)

`TRAEFIK_TRACING_DATADOG_BAGAGEPREFIXHEADERNAME`:  
Sets the header name prefix used to store baggage items in a map.

`TRAEFIK_TRACING_DATADOG_DEBUG`:  
Enables Datadog debug. (Default: ```false```)

`TRAEFIK_TRACING_DATADOG_GLOBALTAGS_<NAME>`:  
Sets a list of key:value tags on all spans.

`TRAEFIK_TRACING_DATADOG_LOCALAGENTHOSTPORT`:  
Sets the Datadog Agent host:port. (Default: ```localhost:8126```)

`TRAEFIK_TRACING_DATADOG_PARENTIDHEADERNAME`:  
Sets the header name used to store the parent ID.

`TRAEFIK_TRACING_DATADOG_PRIORITYSAMPLING`:  
Enables priority sampling. When using distributed tracing, this option must be enabled in order to get all the parts of a distributed trace sampled. (Default: ```false```)

`TRAEFIK_TRACING_DATADOG_SAMPLINGPRIORITYHEADERNAME`:  
Sets the header name used to store the sampling priority.

`TRAEFIK_TRACING_DATADOG_TRACEIDHEADERNAME`:  
Sets the header name used to store the trace ID.

`TRAEFIK_TRACING_ELASTIC`:  
Settings for Elastic. (Default: ```false```)

`TRAEFIK_TRACING_ELASTIC_SECRETTOKEN`:  
Sets the token used to connect to Elastic APM Server.

`TRAEFIK_TRACING_ELASTIC_SERVERURL`:  
Sets the URL of the Elastic APM server.

`TRAEFIK_TRACING_ELASTIC_SERVICEENVIRONMENT`:  
Sets the name of the environment Traefik is deployed in, e.g. 'production' or 'staging'.

`TRAEFIK_TRACING_HAYSTACK`:  
Settings for Haystack. (Default: ```false```)

`TRAEFIK_TRACING_HAYSTACK_BAGGAGEPREFIXHEADERNAME`:  
Sets the header name prefix used to store baggage items in a map.

`TRAEFIK_TRACING_HAYSTACK_GLOBALTAG`:  
Sets a key:value tag on all spans.

`TRAEFIK_TRACING_HAYSTACK_LOCALAGENTHOST`:  
Sets the Haystack Agent host. (Default: ```127.0.0.1```)

`TRAEFIK_TRACING_HAYSTACK_LOCALAGENTPORT`:  
Sets the Haystack Agent port. (Default: ```35000```)

`TRAEFIK_TRACING_HAYSTACK_PARENTIDHEADERNAME`:  
Sets the header name used to store the parent ID.

`TRAEFIK_TRACING_HAYSTACK_SPANIDHEADERNAME`:  
Sets the header name used to store the span ID.

`TRAEFIK_TRACING_HAYSTACK_TRACEIDHEADERNAME`:  
Sets the header name used to store the trace ID.

`TRAEFIK_TRACING_INSTANA`:  
Settings for Instana. (Default: ```false```)

`TRAEFIK_TRACING_INSTANA_ENABLEAUTOPROFILE`:  
Enables automatic profiling for the Traefik process. (Default: ```false```)

`TRAEFIK_TRACING_INSTANA_LOCALAGENTHOST`:  
Sets the Instana Agent host.

`TRAEFIK_TRACING_INSTANA_LOCALAGENTPORT`:  
Sets the Instana Agent port. (Default: ```42699```)

`TRAEFIK_TRACING_INSTANA_LOGLEVEL`:  
Sets the log level for the Instana tracer. ('error','warn','info','debug') (Default: ```info```)

`TRAEFIK_TRACING_JAEGER`:  
Settings for Jaeger. (Default: ```false```)

`TRAEFIK_TRACING_JAEGER_COLLECTOR_ENDPOINT`:  
Instructs reporter to send spans to jaeger-collector at this URL.

`TRAEFIK_TRACING_JAEGER_COLLECTOR_PASSWORD`:  
Password for basic http authentication when sending spans to jaeger-collector.

`TRAEFIK_TRACING_JAEGER_COLLECTOR_USER`:  
User for basic http authentication when sending spans to jaeger-collector.

`TRAEFIK_TRACING_JAEGER_DISABLEATTEMPTRECONNECTING`:  
Disables the periodic re-resolution of the agent's hostname and reconnection if there was a change. (Default: ```true```)

`TRAEFIK_TRACING_JAEGER_GEN128BIT`:  
Generates 128 bits span IDs. (Default: ```false```)

`TRAEFIK_TRACING_JAEGER_LOCALAGENTHOSTPORT`:  
Sets the Jaeger Agent host:port. (Default: ```127.0.0.1:6831```)

`TRAEFIK_TRACING_JAEGER_PROPAGATION`:  
Sets the propagation format (jaeger/b3). (Default: ```jaeger```)

`TRAEFIK_TRACING_JAEGER_SAMPLINGPARAM`:  
Sets the sampling parameter. (Default: ```1.000000```)

`TRAEFIK_TRACING_JAEGER_SAMPLINGSERVERURL`:  
Sets the sampling server URL. (Default: ```http://localhost:5778/sampling```)

`TRAEFIK_TRACING_JAEGER_SAMPLINGTYPE`:  
Sets the sampling type. (Default: ```const```)

`TRAEFIK_TRACING_JAEGER_TRACECONTEXTHEADERNAME`:  
Sets the header name used to store the trace ID. (Default: ```uber-trace-id```)

`TRAEFIK_TRACING_OPENTELEMETRY`:  
Settings for OpenTelemetry. (Default: ```false```)

`TRAEFIK_TRACING_OPENTELEMETRY_ADDRESS`:  
Sets the address (host:port) of the collector endpoint. (Default: ```localhost:4318```)

`TRAEFIK_TRACING_OPENTELEMETRY_GRPC`:  
gRPC specific configuration for the OpenTelemetry collector. (Default: ```true```)

`TRAEFIK_TRACING_OPENTELEMETRY_HEADERS_<NAME>`:  
Defines additional headers to be sent with the payloads.

`TRAEFIK_TRACING_OPENTELEMETRY_INSECURE`:  
Disables client transport security for the exporter. (Default: ```false```)

`TRAEFIK_TRACING_OPENTELEMETRY_PATH`:  
Sets the URL path of the collector endpoint.

`TRAEFIK_TRACING_OPENTELEMETRY_TLS_CA`:  
TLS CA

`TRAEFIK_TRACING_OPENTELEMETRY_TLS_CERT`:  
TLS cert

`TRAEFIK_TRACING_OPENTELEMETRY_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_TRACING_OPENTELEMETRY_TLS_KEY`:  
TLS key

`TRAEFIK_TRACING_SERVICENAME`:  
Set the name for this service. (Default: ```traefik```)

`TRAEFIK_TRACING_SPANNAMELIMIT`:  
Set the maximum character limit for Span names (default 0 = no limit). (Default: ```0```)

`TRAEFIK_TRACING_ZIPKIN`:  
Settings for Zipkin. (Default: ```false```)

`TRAEFIK_TRACING_ZIPKIN_HTTPENDPOINT`:  
Sets the HTTP Endpoint to report traces to. (Default: ```http://localhost:9411/api/v2/spans```)

`TRAEFIK_TRACING_ZIPKIN_ID128BIT`:  
Uses 128 bits root span IDs. (Default: ```true```)

`TRAEFIK_TRACING_ZIPKIN_SAMESPAN`:  
Uses SameSpan RPC style traces. (Default: ```false```)

`TRAEFIK_TRACING_ZIPKIN_SAMPLERATE`:  
Sets the rate between 0.0 and 1.0 of requests to trace. (Default: ```1.000000```)
