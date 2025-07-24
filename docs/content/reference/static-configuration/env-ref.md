<!--
CODE GENERATED AUTOMATICALLY
THIS FILE MUST NOT BE EDITED BY HAND
-->

`TRAEFIK_ACCESSLOG`:  
Access log settings. (Default: ```false```)

`TRAEFIK_ACCESSLOG_ADDINTERNALS`:  
Enables access log for internal services (ping, dashboard, etc...). (Default: ```false```)

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

`TRAEFIK_ACCESSLOG_OTLP`:  
Settings for OpenTelemetry. (Default: ```false```)

`TRAEFIK_ACCESSLOG_OTLP_GRPC`:  
gRPC configuration for the OpenTelemetry collector. (Default: ```false```)

`TRAEFIK_ACCESSLOG_OTLP_GRPC_ENDPOINT`:  
Sets the gRPC endpoint (host:port) of the collector. (Default: ```localhost:4317```)

`TRAEFIK_ACCESSLOG_OTLP_GRPC_HEADERS_<NAME>`:  
Headers sent with payload.

`TRAEFIK_ACCESSLOG_OTLP_GRPC_INSECURE`:  
Disables client transport security for the exporter. (Default: ```false```)

`TRAEFIK_ACCESSLOG_OTLP_GRPC_TLS_CA`:  
TLS CA

`TRAEFIK_ACCESSLOG_OTLP_GRPC_TLS_CERT`:  
TLS cert

`TRAEFIK_ACCESSLOG_OTLP_GRPC_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_ACCESSLOG_OTLP_GRPC_TLS_KEY`:  
TLS key

`TRAEFIK_ACCESSLOG_OTLP_HTTP`:  
HTTP configuration for the OpenTelemetry collector. (Default: ```false```)

`TRAEFIK_ACCESSLOG_OTLP_HTTP_ENDPOINT`:  
Sets the HTTP endpoint (scheme://host:port/path) of the collector. (Default: ```https://localhost:4318```)

`TRAEFIK_ACCESSLOG_OTLP_HTTP_HEADERS_<NAME>`:  
Headers sent with payload.

`TRAEFIK_ACCESSLOG_OTLP_HTTP_TLS_CA`:  
TLS CA

`TRAEFIK_ACCESSLOG_OTLP_HTTP_TLS_CERT`:  
TLS cert

`TRAEFIK_ACCESSLOG_OTLP_HTTP_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_ACCESSLOG_OTLP_HTTP_TLS_KEY`:  
TLS key

`TRAEFIK_ACCESSLOG_OTLP_RESOURCEATTRIBUTES_<NAME>`:  
Defines additional resource attributes (key:value).

`TRAEFIK_ACCESSLOG_OTLP_SERVICENAME`:  
Defines the service name resource attribute. (Default: ```traefik```)

`TRAEFIK_API`:  
Enable api/dashboard. (Default: ```false```)

`TRAEFIK_API_BASEPATH`:  
Defines the base path where the API and Dashboard will be exposed. (Default: ```/```)

`TRAEFIK_API_DASHBOARD`:  
Activate dashboard. (Default: ```true```)

`TRAEFIK_API_DEBUG`:  
Enable additional endpoints for debugging and profiling. (Default: ```false```)

`TRAEFIK_API_DISABLEDASHBOARDAD`:  
Disable ad in the dashboard. (Default: ```false```)

`TRAEFIK_API_INSECURE`:  
Activate API directly on the entryPoint named traefik. (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>`:  
Certificates resolvers configuration. (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_CACERTIFICATES`:  
Specify the paths to PEM encoded CA Certificates that can be used to authenticate an ACME server with an HTTPS certificate not issued by a CA in the system-wide trusted root list.

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_CASERVER`:  
CA server to use. (Default: ```https://acme-v02.api.letsencrypt.org/directory```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_CASERVERNAME`:  
Specify the CA server name that can be used to authenticate an ACME server with an HTTPS certificate not issued by a CA in the system-wide trusted root list.

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_CASYSTEMCERTPOOL`:  
Define if the certificates pool must use a copy of the system cert pool. (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_CERTIFICATESDURATION`:  
Certificates' duration in hours. (Default: ```2160```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_CLIENTRESPONSEHEADERTIMEOUT`:  
Timeout for receiving the response headers when communicating with the ACME server. (Default: ```30```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_CLIENTTIMEOUT`:  
Timeout for a complete HTTP transaction with the ACME server. (Default: ```120```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_DNSCHALLENGE`:  
Activate DNS-01 Challenge. (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_DNSCHALLENGE_DELAYBEFORECHECK`:  
(Deprecated) Assume DNS propagates after a delay in seconds rather than finding and querying nameservers. (Default: ```0```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_DNSCHALLENGE_DISABLEPROPAGATIONCHECK`:  
(Deprecated) Disable the DNS propagation checks before notifying ACME that the DNS challenge is ready. [not recommended] (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_DNSCHALLENGE_PROPAGATION`:  
DNS propagation checks configuration (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_DNSCHALLENGE_PROPAGATION_DELAYBEFORECHECKS`:  
Defines the delay before checking the challenge TXT record propagation. (Default: ```0```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_DNSCHALLENGE_PROPAGATION_DISABLEANSCHECKS`:  
Disables the challenge TXT record propagation checks against authoritative nameservers. (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_DNSCHALLENGE_PROPAGATION_DISABLECHECKS`:  
Disables the challenge TXT record propagation checks (not recommended). (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_DNSCHALLENGE_PROPAGATION_REQUIREALLRNS`:  
Requires the challenge TXT record to be propagated to all recursive nameservers. (Default: ```false```)

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

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_EMAILADDRESSES`:  
CSR email addresses to use.

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_HTTPCHALLENGE`:  
Activate HTTP-01 Challenge. (Default: ```false```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_HTTPCHALLENGE_DELAY`:  
Delay between the creation of the challenge and the validation. (Default: ```0```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_HTTPCHALLENGE_ENTRYPOINT`:  
HTTP challenge EntryPoint

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_KEYTYPE`:  
KeyType used for generating certificate private key. Allow value 'EC256', 'EC384', 'RSA2048', 'RSA4096', 'RSA8192'. (Default: ```RSA4096```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_PREFERREDCHAIN`:  
Preferred chain to use.

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_PROFILE`:  
Certificate profile to use.

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_STORAGE`:  
Storage to use. (Default: ```acme.json```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_ACME_TLSCHALLENGE`:  
Activate TLS-ALPN-01 Challenge. (Default: ```true```)

`TRAEFIK_CERTIFICATESRESOLVERS_<NAME>_TAILSCALE`:  
Enables Tailscale certificate resolution. (Default: ```true```)

`TRAEFIK_CORE_DEFAULTRULESYNTAX`:  
Defines the rule parser default syntax (v2 or v3) (Default: ```v3```)

`TRAEFIK_ENTRYPOINTS_<NAME>`:  
Entry points definition. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_ADDRESS`:  
Entry point address.

`TRAEFIK_ENTRYPOINTS_<NAME>_ALLOWACMEBYPASS`:  
Enables handling of ACME TLS and HTTP challenges with custom routers. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_ASDEFAULT`:  
Adds this EntryPoint to the list of default EntryPoints to be used on routers that don't have any Entrypoint defined. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_FORWARDEDHEADERS_CONNECTION`:  
List of Connection headers that are allowed to pass through the middleware chain before being removed.

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

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_ENCODEQUERYSEMICOLONS`:  
Defines whether request query semicolons should be URLEncoded. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_MAXHEADERBYTES`:  
Maximum size of request headers in bytes. (Default: ```1048576```)

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_MIDDLEWARES`:  
Default middlewares for the routers linked to the entry point.

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_REDIRECTIONS_ENTRYPOINT_PERMANENT`:  
Applies a permanent redirection. (Default: ```true```)

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_REDIRECTIONS_ENTRYPOINT_PRIORITY`:  
Priority of the generated router. (Default: ```9223372036854775806```)

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_REDIRECTIONS_ENTRYPOINT_SCHEME`:  
Scheme used for the redirection. (Default: ```https```)

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_REDIRECTIONS_ENTRYPOINT_TO`:  
Targeted entry point of the redirection.

`TRAEFIK_ENTRYPOINTS_<NAME>_HTTP_SANITIZEPATH`:  
Defines whether to enable request path sanitization (removal of /./, /../ and multiple slash sequences). (Default: ```true```)

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

`TRAEFIK_ENTRYPOINTS_<NAME>_OBSERVABILITY_ACCESSLOGS`:  
Enables access-logs for this entryPoint. (Default: ```true```)

`TRAEFIK_ENTRYPOINTS_<NAME>_OBSERVABILITY_METRICS`:  
Enables metrics for this entryPoint. (Default: ```true```)

`TRAEFIK_ENTRYPOINTS_<NAME>_OBSERVABILITY_TRACEVERBOSITY`:  
Defines the tracing verbosity level for this entryPoint. (Default: ```minimal```)

`TRAEFIK_ENTRYPOINTS_<NAME>_OBSERVABILITY_TRACING`:  
Enables tracing for this entryPoint. (Default: ```true```)

`TRAEFIK_ENTRYPOINTS_<NAME>_PROXYPROTOCOL`:  
Proxy-Protocol configuration. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_PROXYPROTOCOL_INSECURE`:  
Trust all. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_PROXYPROTOCOL_TRUSTEDIPS`:  
Trust only selected IPs.

`TRAEFIK_ENTRYPOINTS_<NAME>_REUSEPORT`:  
Enables EntryPoints from the same or different processes listening on the same TCP/UDP port. (Default: ```false```)

`TRAEFIK_ENTRYPOINTS_<NAME>_TRANSPORT_KEEPALIVEMAXREQUESTS`:  
Maximum number of requests before closing a keep-alive connection. (Default: ```0```)

`TRAEFIK_ENTRYPOINTS_<NAME>_TRANSPORT_KEEPALIVEMAXTIME`:  
Maximum duration before closing a keep-alive connection. (Default: ```0```)

`TRAEFIK_ENTRYPOINTS_<NAME>_TRANSPORT_LIFECYCLE_GRACETIMEOUT`:  
Duration to give active requests a chance to finish before Traefik stops. (Default: ```10```)

`TRAEFIK_ENTRYPOINTS_<NAME>_TRANSPORT_LIFECYCLE_REQUESTACCEPTGRACETIMEOUT`:  
Duration to keep accepting requests before Traefik initiates the graceful shutdown procedure. (Default: ```0```)

`TRAEFIK_ENTRYPOINTS_<NAME>_TRANSPORT_RESPONDINGTIMEOUTS_IDLETIMEOUT`:  
IdleTimeout is the maximum amount duration an idle (keep-alive) connection will remain idle before closing itself. If zero, no timeout is set. (Default: ```180```)

`TRAEFIK_ENTRYPOINTS_<NAME>_TRANSPORT_RESPONDINGTIMEOUTS_READTIMEOUT`:  
ReadTimeout is the maximum duration for reading the entire request, including the body. If zero, no timeout is set. (Default: ```60```)

`TRAEFIK_ENTRYPOINTS_<NAME>_TRANSPORT_RESPONDINGTIMEOUTS_WRITETIMEOUT`:  
WriteTimeout is the maximum duration before timing out writes of the response. If zero, no timeout is set. (Default: ```0```)

`TRAEFIK_ENTRYPOINTS_<NAME>_UDP_TIMEOUT`:  
Timeout defines how long to wait on an idle session before releasing the related resources. (Default: ```3```)

`TRAEFIK_EXPERIMENTAL_ABORTONPLUGINFAILURE`:  
Defines whether all plugins must be loaded successfully for Traefik to start. (Default: ```false```)

`TRAEFIK_EXPERIMENTAL_FASTPROXY`:  
Enables the FastProxy implementation. (Default: ```false```)

`TRAEFIK_EXPERIMENTAL_FASTPROXY_DEBUG`:  
Enable debug mode for the FastProxy implementation. (Default: ```false```)

`TRAEFIK_EXPERIMENTAL_KUBERNETESGATEWAY`:  
(Deprecated) Allow the Kubernetes gateway api provider usage. (Default: ```false```)

`TRAEFIK_EXPERIMENTAL_KUBERNETESINGRESSNGINX`:  
Allow the Kubernetes Ingress NGINX provider usage. (Default: ```false```)

`TRAEFIK_EXPERIMENTAL_LOCALPLUGINS_<NAME>`:  
Local plugins configuration. (Default: ```false```)

`TRAEFIK_EXPERIMENTAL_LOCALPLUGINS_<NAME>_MODULENAME`:  
Plugin's module name.

`TRAEFIK_EXPERIMENTAL_LOCALPLUGINS_<NAME>_SETTINGS`:  
Plugin's settings (works only for wasm plugins).

`TRAEFIK_EXPERIMENTAL_LOCALPLUGINS_<NAME>_SETTINGS_ENVS`:  
Environment variables to forward to the wasm guest.

`TRAEFIK_EXPERIMENTAL_LOCALPLUGINS_<NAME>_SETTINGS_MOUNTS`:  
Directory to mount to the wasm guest.

`TRAEFIK_EXPERIMENTAL_LOCALPLUGINS_<NAME>_SETTINGS_USEUNSAFE`:  
Allow the plugin to use unsafe package. (Default: ```false```)

`TRAEFIK_EXPERIMENTAL_OTLPLOGS`:  
Enables the OpenTelemetry logs integration. (Default: ```false```)

`TRAEFIK_EXPERIMENTAL_PLUGINS_<NAME>_MODULENAME`:  
plugin's module name.

`TRAEFIK_EXPERIMENTAL_PLUGINS_<NAME>_SETTINGS`:  
Plugin's settings (works only for wasm plugins).

`TRAEFIK_EXPERIMENTAL_PLUGINS_<NAME>_SETTINGS_ENVS`:  
Environment variables to forward to the wasm guest.

`TRAEFIK_EXPERIMENTAL_PLUGINS_<NAME>_SETTINGS_MOUNTS`:  
Directory to mount to the wasm guest.

`TRAEFIK_EXPERIMENTAL_PLUGINS_<NAME>_SETTINGS_USEUNSAFE`:  
Allow the plugin to use unsafe package. (Default: ```false```)

`TRAEFIK_EXPERIMENTAL_PLUGINS_<NAME>_VERSION`:  
plugin's version.

`TRAEFIK_GLOBAL_CHECKNEWVERSION`:  
Periodically check if a new version has been released. (Default: ```true```)

`TRAEFIK_GLOBAL_SENDANONYMOUSUSAGE`:  
Periodically send anonymous usage statistics. If the option is not specified, it will be disabled by default. (Default: ```false```)

`TRAEFIK_HOSTRESOLVER`:  
Enable CNAME Flattening. (Default: ```false```)

`TRAEFIK_HOSTRESOLVER_CNAMEFLATTENING`:  
A flag to enable/disable CNAME flattening (Default: ```false```)

`TRAEFIK_HOSTRESOLVER_RESOLVCONFIG`:  
resolv.conf used for DNS resolving (Default: ```/etc/resolv.conf```)

`TRAEFIK_HOSTRESOLVER_RESOLVDEPTH`:  
The maximal depth of DNS recursive resolving (Default: ```5```)

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

`TRAEFIK_LOG_OTLP`:  
Settings for OpenTelemetry. (Default: ```false```)

`TRAEFIK_LOG_OTLP_GRPC`:  
gRPC configuration for the OpenTelemetry collector. (Default: ```false```)

`TRAEFIK_LOG_OTLP_GRPC_ENDPOINT`:  
Sets the gRPC endpoint (host:port) of the collector. (Default: ```localhost:4317```)

`TRAEFIK_LOG_OTLP_GRPC_HEADERS_<NAME>`:  
Headers sent with payload.

`TRAEFIK_LOG_OTLP_GRPC_INSECURE`:  
Disables client transport security for the exporter. (Default: ```false```)

`TRAEFIK_LOG_OTLP_GRPC_TLS_CA`:  
TLS CA

`TRAEFIK_LOG_OTLP_GRPC_TLS_CERT`:  
TLS cert

`TRAEFIK_LOG_OTLP_GRPC_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_LOG_OTLP_GRPC_TLS_KEY`:  
TLS key

`TRAEFIK_LOG_OTLP_HTTP`:  
HTTP configuration for the OpenTelemetry collector. (Default: ```false```)

`TRAEFIK_LOG_OTLP_HTTP_ENDPOINT`:  
Sets the HTTP endpoint (scheme://host:port/path) of the collector. (Default: ```https://localhost:4318```)

`TRAEFIK_LOG_OTLP_HTTP_HEADERS_<NAME>`:  
Headers sent with payload.

`TRAEFIK_LOG_OTLP_HTTP_TLS_CA`:  
TLS CA

`TRAEFIK_LOG_OTLP_HTTP_TLS_CERT`:  
TLS cert

`TRAEFIK_LOG_OTLP_HTTP_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_LOG_OTLP_HTTP_TLS_KEY`:  
TLS key

`TRAEFIK_LOG_OTLP_RESOURCEATTRIBUTES_<NAME>`:  
Defines additional resource attributes (key:value).

`TRAEFIK_LOG_OTLP_SERVICENAME`:  
Defines the service name resource attribute. (Default: ```traefik```)

`TRAEFIK_METRICS_ADDINTERNALS`:  
Enables metrics for internal services (ping, dashboard, etc...). (Default: ```false```)

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

`TRAEFIK_METRICS_OTLP`:  
OpenTelemetry metrics exporter type. (Default: ```false```)

`TRAEFIK_METRICS_OTLP_ADDENTRYPOINTSLABELS`:  
Enable metrics on entry points. (Default: ```true```)

`TRAEFIK_METRICS_OTLP_ADDROUTERSLABELS`:  
Enable metrics on routers. (Default: ```false```)

`TRAEFIK_METRICS_OTLP_ADDSERVICESLABELS`:  
Enable metrics on services. (Default: ```true```)

`TRAEFIK_METRICS_OTLP_EXPLICITBOUNDARIES`:  
Boundaries for latency metrics. (Default: ```0.005000, 0.010000, 0.025000, 0.050000, 0.075000, 0.100000, 0.250000, 0.500000, 0.750000, 1.000000, 2.500000, 5.000000, 7.500000, 10.000000```)

`TRAEFIK_METRICS_OTLP_GRPC`:  
gRPC configuration for the OpenTelemetry collector. (Default: ```false```)

`TRAEFIK_METRICS_OTLP_GRPC_ENDPOINT`:  
Sets the gRPC endpoint (host:port) of the collector. (Default: ```localhost:4317```)

`TRAEFIK_METRICS_OTLP_GRPC_HEADERS_<NAME>`:  
Headers sent with payload.

`TRAEFIK_METRICS_OTLP_GRPC_INSECURE`:  
Disables client transport security for the exporter. (Default: ```false```)

`TRAEFIK_METRICS_OTLP_GRPC_TLS_CA`:  
TLS CA

`TRAEFIK_METRICS_OTLP_GRPC_TLS_CERT`:  
TLS cert

`TRAEFIK_METRICS_OTLP_GRPC_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_METRICS_OTLP_GRPC_TLS_KEY`:  
TLS key

`TRAEFIK_METRICS_OTLP_HTTP`:  
HTTP configuration for the OpenTelemetry collector. (Default: ```false```)

`TRAEFIK_METRICS_OTLP_HTTP_ENDPOINT`:  
Sets the HTTP endpoint (scheme://host:port/path) of the collector. (Default: ```https://localhost:4318```)

`TRAEFIK_METRICS_OTLP_HTTP_HEADERS_<NAME>`:  
Headers sent with payload.

`TRAEFIK_METRICS_OTLP_HTTP_TLS_CA`:  
TLS CA

`TRAEFIK_METRICS_OTLP_HTTP_TLS_CERT`:  
TLS cert

`TRAEFIK_METRICS_OTLP_HTTP_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_METRICS_OTLP_HTTP_TLS_KEY`:  
TLS key

`TRAEFIK_METRICS_OTLP_PUSHINTERVAL`:  
Period between calls to collect a checkpoint. (Default: ```10```)

`TRAEFIK_METRICS_OTLP_RESOURCEATTRIBUTES_<NAME>`:  
Defines additional resource attributes (key:value).

`TRAEFIK_METRICS_OTLP_SERVICENAME`:  
Defines the service name resource attribute. (Default: ```traefik```)

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

`TRAEFIK_METRICS_PROMETHEUS_HEADERLABELS_<NAME>`:  
Defines the extra labels for the requests_total metrics, and for each of them, the request header containing the value for this label.

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

`TRAEFIK_OCSP`:  
OCSP configuration. (Default: ```false```)

`TRAEFIK_OCSP_RESPONDEROVERRIDES_<NAME>`:  
Defines a map of OCSP responders to replace for querying OCSP servers.

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

`TRAEFIK_PROVIDERS_CONSULCATALOG_STRICTCHECKS`:  
A list of service health statuses to allow taking traffic. (Default: ```passing, warning```)

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
Docker server endpoint. Can be a TCP or a Unix socket endpoint. (Default: ```unix:///var/run/docker.sock```)

`TRAEFIK_PROVIDERS_DOCKER_EXPOSEDBYDEFAULT`:  
Expose containers by default. (Default: ```true```)

`TRAEFIK_PROVIDERS_DOCKER_HTTPCLIENTTIMEOUT`:  
Client timeout for HTTP connections. (Default: ```0```)

`TRAEFIK_PROVIDERS_DOCKER_NETWORK`:  
Default Docker network used.

`TRAEFIK_PROVIDERS_DOCKER_PASSWORD`:  
Password for Basic HTTP authentication.

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

`TRAEFIK_PROVIDERS_DOCKER_USERNAME`:  
Username for Basic HTTP authentication.

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

`TRAEFIK_PROVIDERS_KUBERNETESCRD_DISABLECLUSTERSCOPERESOURCES`:  
Disables the lookup of cluster scope resources (incompatible with IngressClasses and NodePortLB enabled services). (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESCRD_ENDPOINT`:  
Kubernetes server endpoint (required for external cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESCRD_INGRESSCLASS`:  
Value of kubernetes.io/ingress.class annotation to watch for.

`TRAEFIK_PROVIDERS_KUBERNETESCRD_LABELSELECTOR`:  
Kubernetes label selector to use.

`TRAEFIK_PROVIDERS_KUBERNETESCRD_NAMESPACES`:  
Kubernetes namespaces.

`TRAEFIK_PROVIDERS_KUBERNETESCRD_NATIVELBBYDEFAULT`:  
Defines whether to use Native Kubernetes load-balancing mode by default. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESCRD_THROTTLEDURATION`:  
Ingress refresh throttle duration (Default: ```0```)

`TRAEFIK_PROVIDERS_KUBERNETESCRD_TOKEN`:  
Kubernetes bearer token (not needed for in-cluster client). It accepts either a token value or a file path to the token.

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY`:  
Enable Kubernetes gateway api provider with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_CERTAUTHFILEPATH`:  
Kubernetes certificate authority file path (not needed for in-cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_ENDPOINT`:  
Kubernetes server endpoint (required for external cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_EXPERIMENTALCHANNEL`:  
Toggles Experimental Channel resources support (TCPRoute, TLSRoute...). (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_LABELSELECTOR`:  
Kubernetes label selector to select specific GatewayClasses.

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_NAMESPACES`:  
Kubernetes namespaces.

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_NATIVELBBYDEFAULT`:  
Defines whether to use Native Kubernetes load-balancing by default. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_STATUSADDRESS_HOSTNAME`:  
Hostname used for Kubernetes Gateway status address.

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_STATUSADDRESS_IP`:  
IP used to set Kubernetes Gateway status address.

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_STATUSADDRESS_SERVICE`:  
Published Kubernetes Service to copy status addresses from.

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_STATUSADDRESS_SERVICE_NAME`:  
Name of the Kubernetes service.

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_STATUSADDRESS_SERVICE_NAMESPACE`:  
Namespace of the Kubernetes service.

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_THROTTLEDURATION`:  
Kubernetes refresh throttle duration (Default: ```0```)

`TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_TOKEN`:  
Kubernetes bearer token (not needed for in-cluster client). It accepts either a token value or a file path to the token.

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS`:  
Enable Kubernetes backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX`:  
Enable Kubernetes Ingress NGINX provider. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_CERTAUTHFILEPATH`:  
Kubernetes certificate authority file path (not needed for in-cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_CONTROLLERCLASS`:  
Ingress Class Controller value this controller satisfies. (Default: ```k8s.io/ingress-nginx```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_DEFAULTBACKENDSERVICE`:  
Service used to serve HTTP requests not matching any known server name (catch-all). Takes the form 'namespace/name'.

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_DISABLESVCEXTERNALNAME`:  
Disable support for Services of type ExternalName. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_ENDPOINT`:  
Kubernetes server endpoint (required for external cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_INGRESSCLASS`:  
Name of the ingress class this controller satisfies. (Default: ```nginx```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_INGRESSCLASSBYNAME`:  
Define if Ingress Controller should watch for Ingress Class by Name together with Controller Class. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_PUBLISHSERVICE`:  
Service fronting the Ingress controller. Takes the form 'namespace/name'.

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_PUBLISHSTATUSADDRESS`:  
Customized address (or addresses, separated by comma) to set as the load-balancer status of Ingress objects this controller satisfies.

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_THROTTLEDURATION`:  
Ingress refresh throttle duration. (Default: ```0```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_TOKEN`:  
Kubernetes bearer token (not needed for in-cluster client). It accepts either a token value or a file path to the token.

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_WATCHINGRESSWITHOUTCLASS`:  
Define if Ingress Controller should also watch for Ingresses without an IngressClass or the annotation specified. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_WATCHNAMESPACE`:  
Namespace the controller watches for updates to Kubernetes objects. All namespaces are watched if this parameter is left empty.

`TRAEFIK_PROVIDERS_KUBERNETESINGRESSNGINX_WATCHNAMESPACESELECTOR`:  
Selector selects namespaces the controller watches for updates to Kubernetes objects.

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_ALLOWEMPTYSERVICES`:  
Allow creation of services without endpoints. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_ALLOWEXTERNALNAMESERVICES`:  
Allow ExternalName services. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_CERTAUTHFILEPATH`:  
Kubernetes certificate authority file path (not needed for in-cluster client).

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_DISABLECLUSTERSCOPERESOURCES`:  
Disables the lookup of cluster scope resources (incompatible with IngressClasses and NodePortLB enabled services). (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_DISABLEINGRESSCLASSLOOKUP`:  
Disables the lookup of IngressClasses (Deprecated, please use DisableClusterScopeResources). (Default: ```false```)

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

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_NATIVELBBYDEFAULT`:  
Defines whether to use Native Kubernetes load-balancing mode by default. (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_STRICTPREFIXMATCHING`:  
Make prefix matching strictly comply with the Kubernetes Ingress specification (path-element-wise matching instead of character-by-character string matching). (Default: ```false```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_THROTTLEDURATION`:  
Ingress refresh throttle duration (Default: ```0```)

`TRAEFIK_PROVIDERS_KUBERNETESINGRESS_TOKEN`:  
Kubernetes bearer token (not needed for in-cluster client). It accepts either a token value or a file path to the token.

`TRAEFIK_PROVIDERS_NOMAD`:  
Enable Nomad backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_NOMAD_ALLOWEMPTYSERVICES`:  
Allow the creation of services without endpoints. (Default: ```false```)

`TRAEFIK_PROVIDERS_NOMAD_CONSTRAINTS`:  
Constraints is an expression that Traefik matches against the Nomad service's tags to determine whether to create route(s) for that service.

`TRAEFIK_PROVIDERS_NOMAD_DEFAULTRULE`:  
Default rule. (Default: ```Host(`{{ normalize .Name }}`)```)

`TRAEFIK_PROVIDERS_NOMAD_ENDPOINT_ADDRESS`:  
The address of the Nomad server, including scheme and port. (Default: ```http://127.0.0.1:4646```)

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

`TRAEFIK_PROVIDERS_NOMAD_THROTTLEDURATION`:  
Watch throttle duration. (Default: ```0```)

`TRAEFIK_PROVIDERS_NOMAD_WATCH`:  
Watch Nomad Service events. (Default: ```false```)

`TRAEFIK_PROVIDERS_PLUGIN_<NAME>`:  
Plugins configuration.

`TRAEFIK_PROVIDERS_PROVIDERSTHROTTLEDURATION`:  
Backends throttle duration: minimum duration between 2 events from providers before applying a new configuration. It avoids unnecessary reloads if multiples events are sent in a short amount of time. (Default: ```2```)

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

`TRAEFIK_PROVIDERS_REDIS_SENTINEL_LATENCYSTRATEGY`:  
Defines whether to route commands to the closest master or replica nodes (mutually exclusive with RandomStrategy and ReplicaStrategy). (Default: ```false```)

`TRAEFIK_PROVIDERS_REDIS_SENTINEL_MASTERNAME`:  
Name of the master.

`TRAEFIK_PROVIDERS_REDIS_SENTINEL_PASSWORD`:  
Password for Sentinel authentication.

`TRAEFIK_PROVIDERS_REDIS_SENTINEL_RANDOMSTRATEGY`:  
Defines whether to route commands randomly to master or replica nodes (mutually exclusive with LatencyStrategy and ReplicaStrategy). (Default: ```false```)

`TRAEFIK_PROVIDERS_REDIS_SENTINEL_REPLICASTRATEGY`:  
Defines whether to route all commands to replica nodes (mutually exclusive with LatencyStrategy and RandomStrategy). (Default: ```false```)

`TRAEFIK_PROVIDERS_REDIS_SENTINEL_USEDISCONNECTEDREPLICAS`:  
Use replicas disconnected with master when cannot get connected replicas. (Default: ```false```)

`TRAEFIK_PROVIDERS_REDIS_SENTINEL_USERNAME`:  
Username for Sentinel authentication.

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

`TRAEFIK_PROVIDERS_SWARM`:  
Enable Docker Swarm backend with default settings. (Default: ```false```)

`TRAEFIK_PROVIDERS_SWARM_ALLOWEMPTYSERVICES`:  
Disregards the Docker containers health checks with respect to the creation or removal of the corresponding services. (Default: ```false```)

`TRAEFIK_PROVIDERS_SWARM_CONSTRAINTS`:  
Constraints is an expression that Traefik matches against the container's labels to determine whether to create any route for that container.

`TRAEFIK_PROVIDERS_SWARM_DEFAULTRULE`:  
Default rule. (Default: ```Host(`{{ normalize .Name }}`)```)

`TRAEFIK_PROVIDERS_SWARM_ENDPOINT`:  
Docker server endpoint. Can be a TCP or a Unix socket endpoint. (Default: ```unix:///var/run/docker.sock```)

`TRAEFIK_PROVIDERS_SWARM_EXPOSEDBYDEFAULT`:  
Expose containers by default. (Default: ```true```)

`TRAEFIK_PROVIDERS_SWARM_HTTPCLIENTTIMEOUT`:  
Client timeout for HTTP connections. (Default: ```0```)

`TRAEFIK_PROVIDERS_SWARM_NETWORK`:  
Default Docker network used.

`TRAEFIK_PROVIDERS_SWARM_PASSWORD`:  
Password for Basic HTTP authentication.

`TRAEFIK_PROVIDERS_SWARM_REFRESHSECONDS`:  
Polling interval for swarm mode. (Default: ```15```)

`TRAEFIK_PROVIDERS_SWARM_TLS_CA`:  
TLS CA

`TRAEFIK_PROVIDERS_SWARM_TLS_CERT`:  
TLS cert

`TRAEFIK_PROVIDERS_SWARM_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_PROVIDERS_SWARM_TLS_KEY`:  
TLS key

`TRAEFIK_PROVIDERS_SWARM_USEBINDPORTIP`:  
Use the ip address from the bound port, rather than from the inner network. (Default: ```false```)

`TRAEFIK_PROVIDERS_SWARM_USERNAME`:  
Username for Basic HTTP authentication.

`TRAEFIK_PROVIDERS_SWARM_WATCH`:  
Watch Docker events. (Default: ```true```)

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

`TRAEFIK_TCPSERVERSTRANSPORT_DIALKEEPALIVE`:  
Defines the interval between keep-alive probes for an active network connection. If zero, keep-alive probes are sent with a default value (currently 15 seconds), if supported by the protocol and operating system. Network protocols or operating systems that do not support keep-alives ignore this field. If negative, keep-alive probes are disabled (Default: ```15```)

`TRAEFIK_TCPSERVERSTRANSPORT_DIALTIMEOUT`:  
Defines the amount of time to wait until a connection to a backend server can be established. If zero, no timeout exists. (Default: ```30```)

`TRAEFIK_TCPSERVERSTRANSPORT_TERMINATIONDELAY`:  
Defines the delay to wait before fully terminating the connection, after one connected peer has closed its writing capability. (Default: ```0```)

`TRAEFIK_TCPSERVERSTRANSPORT_TLS`:  
Defines the TLS configuration. (Default: ```false```)

`TRAEFIK_TCPSERVERSTRANSPORT_TLS_INSECURESKIPVERIFY`:  
Disables SSL certificate verification. (Default: ```false```)

`TRAEFIK_TCPSERVERSTRANSPORT_TLS_ROOTCAS`:  
Defines a list of CA secret used to validate self-signed certificate

`TRAEFIK_TCPSERVERSTRANSPORT_TLS_SPIFFE`:  
Defines the SPIFFE TLS configuration. (Default: ```false```)

`TRAEFIK_TCPSERVERSTRANSPORT_TLS_SPIFFE_IDS`:  
Defines the allowed SPIFFE IDs (takes precedence over the SPIFFE TrustDomain).

`TRAEFIK_TCPSERVERSTRANSPORT_TLS_SPIFFE_TRUSTDOMAIN`:  
Defines the allowed SPIFFE trust domain.

`TRAEFIK_TRACING`:  
Tracing configuration. (Default: ```false```)

`TRAEFIK_TRACING_ADDINTERNALS`:  
Enables tracing for internal services (ping, dashboard, etc...). (Default: ```false```)

`TRAEFIK_TRACING_CAPTUREDREQUESTHEADERS`:  
Request headers to add as attributes for server and client spans.

`TRAEFIK_TRACING_CAPTUREDRESPONSEHEADERS`:  
Response headers to add as attributes for server and client spans.

`TRAEFIK_TRACING_GLOBALATTRIBUTES_<NAME>`:  
(Deprecated) Defines additional resource attributes (key:value).

`TRAEFIK_TRACING_OTLP`:  
Settings for OpenTelemetry. (Default: ```false```)

`TRAEFIK_TRACING_OTLP_GRPC`:  
gRPC configuration for the OpenTelemetry collector. (Default: ```false```)

`TRAEFIK_TRACING_OTLP_GRPC_ENDPOINT`:  
Sets the gRPC endpoint (host:port) of the collector. (Default: ```localhost:4317```)

`TRAEFIK_TRACING_OTLP_GRPC_HEADERS_<NAME>`:  
Headers sent with payload.

`TRAEFIK_TRACING_OTLP_GRPC_INSECURE`:  
Disables client transport security for the exporter. (Default: ```false```)

`TRAEFIK_TRACING_OTLP_GRPC_TLS_CA`:  
TLS CA

`TRAEFIK_TRACING_OTLP_GRPC_TLS_CERT`:  
TLS cert

`TRAEFIK_TRACING_OTLP_GRPC_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_TRACING_OTLP_GRPC_TLS_KEY`:  
TLS key

`TRAEFIK_TRACING_OTLP_HTTP`:  
HTTP configuration for the OpenTelemetry collector. (Default: ```false```)

`TRAEFIK_TRACING_OTLP_HTTP_ENDPOINT`:  
Sets the HTTP endpoint (scheme://host:port/path) of the collector. (Default: ```https://localhost:4318```)

`TRAEFIK_TRACING_OTLP_HTTP_HEADERS_<NAME>`:  
Headers sent with payload.

`TRAEFIK_TRACING_OTLP_HTTP_TLS_CA`:  
TLS CA

`TRAEFIK_TRACING_OTLP_HTTP_TLS_CERT`:  
TLS cert

`TRAEFIK_TRACING_OTLP_HTTP_TLS_INSECURESKIPVERIFY`:  
TLS insecure skip verify (Default: ```false```)

`TRAEFIK_TRACING_OTLP_HTTP_TLS_KEY`:  
TLS key

`TRAEFIK_TRACING_RESOURCEATTRIBUTES_<NAME>`:  
Defines additional resource attributes (key:value).

`TRAEFIK_TRACING_SAFEQUERYPARAMS`:  
Query params to not redact.

`TRAEFIK_TRACING_SAMPLERATE`:  
Sets the rate between 0.0 and 1.0 of requests to trace. (Default: ```1.000000```)

`TRAEFIK_TRACING_SERVICENAME`:  
Defines the service name resource attribute. (Default: ```traefik```)
