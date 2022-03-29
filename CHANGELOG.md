## [v2.6.3](https://github.com/traefik/traefik/tree/v2.6.3) (2022-03-28)
[All Commits](https://github.com/traefik/traefik/compare/v2.6.2...v2.6.3)

**Bug fixes:**
- **[plugins]** Fix slice parsing for plugins ([#8886](https://github.com/traefik/traefik/pull/8886) by [ldez](https://github.com/ldez))
- **[tls]** Return TLS unrecognized_name error when no certificate is available ([#8893](https://github.com/traefik/traefik/pull/8893) by [rtribotte](https://github.com/rtribotte))

## [v2.6.2](https://github.com/traefik/traefik/tree/v2.6.2) (2022-03-24)
[All Commits](https://github.com/traefik/traefik/compare/v2.6.1...v2.6.2)

**Bug fixes:**
- **[file]** Bump paerser to v0.1.5 ([#8850](https://github.com/traefik/traefik/pull/8850) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme]** Fix certificates resolver typo ([#8859](https://github.com/traefik/traefik/pull/8859) by [NReilingh](https://github.com/NReilingh))
- **[docker]** doc: fix, docker uses Label(), not Tag() ([#8823](https://github.com/traefik/traefik/pull/8823) by [mpl](https://github.com/mpl))
- **[http3]** Fix CLI syntax in HTTP/3 documentation ([#8864](https://github.com/traefik/traefik/pull/8864) by [nstankov-bg](https://github.com/nstankov-bg))
- **[kv]** Fix small typo in Redis provider documentation ([#8858](https://github.com/traefik/traefik/pull/8858) by [lczw](https://github.com/lczw))
- **[marathon]** Fix brand typo ([#8788](https://github.com/traefik/traefik/pull/8788) by [0xflotus](https://github.com/0xflotus))
- **[middleware]** Fix fenced code block typo in Buffering middleware page ([#8855](https://github.com/traefik/traefik/pull/8855) by [Wingysam](https://github.com/Wingysam))
- **[rules]** Adjust rule length in routers documentation ([#8819](https://github.com/traefik/traefik/pull/8819) by [rtribotte](https://github.com/rtribotte))
- **[rules]** Fix HostRegexp examples ([#8817](https://github.com/traefik/traefik/pull/8817) by [kevinpollet](https://github.com/kevinpollet))
- **[tls,k8s/crd,k8s]** Add default certificate definition example for Kubernetes ([#8863](https://github.com/traefik/traefik/pull/8863) by [jwausle](https://github.com/jwausle))
- **[tls,k8s]** Clarify TLS Option documentation ([#8756](https://github.com/traefik/traefik/pull/8756) by [mloiseleur](https://github.com/mloiseleur))
- Clarify concepts documentation page ([#8836](https://github.com/traefik/traefik/pull/8836) by [NReilingh](https://github.com/NReilingh))
- Spelling ([#8791](https://github.com/traefik/traefik/pull/8791) by [jsoref](https://github.com/jsoref))
- Fix routing overview examples ([#8840](https://github.com/traefik/traefik/pull/8840) by [NReilingh](https://github.com/NReilingh))
- Add a deprecation notices section ([#8829](https://github.com/traefik/traefik/pull/8829) by [ddtmachado](https://github.com/ddtmachado))

## [v2.6.1](https://github.com/traefik/traefik/tree/v2.6.1) (2022-02-14)
[All Commits](https://github.com/traefik/traefik/compare/v2.6.0...v2.6.1)

**Bug fixes:**
- **[acme]** Add domain to HTTP challenge errors ([#8740](https://github.com/traefik/traefik/pull/8740) by [ldez](https://github.com/ldez))
- **[metrics]** Fix metrics bucket key high cardinality ([#8761](https://github.com/traefik/traefik/pull/8761) by [tomMoulard](https://github.com/tomMoulard))
- **[middleware,tls]** Use CNAME for SNI check on host header ([#8773](https://github.com/traefik/traefik/pull/8773) by [ldez](https://github.com/ldez))
- **[middleware,tracing]** Rename Datadog span tags ([#8323](https://github.com/traefik/traefik/pull/8323) by [luckielordie](https://github.com/luckielordie))
- **[tls]** Apply the same approach as the rules system on the TLS configuration choice ([#8764](https://github.com/traefik/traefik/pull/8764) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme]** Add Hurricane Electric to acme documentation ([#8746](https://github.com/traefik/traefik/pull/8746) by [vladshub](https://github.com/vladshub))
- **[acme]** Clarify that ACME challenge is mandatory ([#8739](https://github.com/traefik/traefik/pull/8739) by [mpl](https://github.com/mpl))
- **[http3]** Explain a bit more around enabling HTTP3 ([#8731](https://github.com/traefik/traefik/pull/8731) by [SantoDE](https://github.com/SantoDE))
- **[metrics]** Fix mixups in metrics documentation ([#8752](https://github.com/traefik/traefik/pull/8752) by [tomMoulard](https://github.com/tomMoulard))
- **[middleware,k8s/crd]** Fix Kubernetes TCP examples ([#8759](https://github.com/traefik/traefik/pull/8759) by [sylr](https://github.com/sylr))

## [v2.6.0](https://github.com/traefik/traefik/tree/v2.6.0) (2022-01-24)
[All Commits](https://github.com/traefik/traefik/compare/v2.5.0-rc1...v2.6.0)

**Enhancements:**
- **[acme]** Allow configuration of ACME certificates duration ([#8046](https://github.com/traefik/traefik/pull/8046) by [pmontepagano](https://github.com/pmontepagano))
- **[consul,consulcatalog]** Support consul enterprise namespaces in consul catalog provider ([#8592](https://github.com/traefik/traefik/pull/8592) by [kevinpollet](https://github.com/kevinpollet))
- **[k8s,k8s/gatewayapi]** Update gateway api provider to v1alpha2 ([#8535](https://github.com/traefik/traefik/pull/8535) by [kevinpollet](https://github.com/kevinpollet))
- **[k8s,k8s/gatewayapi]** Support gateway api RouteNamespaces ([#8299](https://github.com/traefik/traefik/pull/8299) by [tomMoulard](https://github.com/tomMoulard))
- **[k8s/crd]** Support Kubernetes basic-auth secrets ([#8189](https://github.com/traefik/traefik/pull/8189) by [dtomcej](https://github.com/dtomcej))
- **[metrics]** Add configurable tags to influxdb metrics ([#8308](https://github.com/traefik/traefik/pull/8308) by [Tetha](https://github.com/Tetha))
- **[metrics]** Add prefix to datadog metrics ([#8234](https://github.com/traefik/traefik/pull/8234) by [fredwangwang](https://github.com/fredwangwang))
- **[middleware,tcp]** Add in flight connection middleware ([#8429](https://github.com/traefik/traefik/pull/8429) by [tomMoulard](https://github.com/tomMoulard))
- **[middleware]** Add Organizational Unit to passtlscert middleware ([#7958](https://github.com/traefik/traefik/pull/7958) by [FernFerret](https://github.com/FernFerret))
- **[middleware]** Allow configuration of minimum body size for compress middleware ([#8239](https://github.com/traefik/traefik/pull/8239) by [lus](https://github.com/lus))
- **[middleware]** Ceil Retry-After value in the rate-limit middleware ([#8581](https://github.com/traefik/traefik/pull/8581) by [pyaillet](https://github.com/pyaillet))
- **[middleware]** Refactor Exponential Backoff ([#7519](https://github.com/traefik/traefik/pull/7519) by [danieladams456](https://github.com/danieladams456))
- **[server,k8s/crd,k8s]** Allow configuration of HTTP/2 readIdleTimeout and pingTimeout ([#8539](https://github.com/traefik/traefik/pull/8539) by [tomMoulard](https://github.com/tomMoulard))
- **[server]** Allow configuration of advertised port for HTTP/3 ([#8131](https://github.com/traefik/traefik/pull/8131) by [valerauko](https://github.com/valerauko))
- **[tracing]** Upgrade Instana tracer and make process profiling configurable ([#8334](https://github.com/traefik/traefik/pull/8334) by [andriikushch](https://github.com/andriikushch))

**Bug fixes:**
- **[consul,kv]** Support Consul KV Enterprise namespaces ([#8692](https://github.com/traefik/traefik/pull/8692) by [kevinpollet](https://github.com/kevinpollet))
- **[consul]** Support token authentication for Consul KV  ([#8712](https://github.com/traefik/traefik/pull/8712) by [kevinpollet](https://github.com/kevinpollet))
- **[consulcatalog]** Configure Consul Catalog namespace at client level ([#8725](https://github.com/traefik/traefik/pull/8725) by [kevinpollet](https://github.com/kevinpollet))
- **[tracing]** Upgrade Instana tracer dependency ([#8687](https://github.com/traefik/traefik/pull/8687) by [andriikushch](https://github.com/andriikushch))
- **[logs]** Redact credentials before logging ([#8699](https://github.com/traefik/traefik/pull/8699) by [ibrahimalihc](https://github.com/ibrahimalihc))

**Misc:**
- Merge current v2.5 into v2.6 ([#8720](https://github.com/traefik/traefik/pull/8720) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into v2.6 ([#8717](https://github.com/traefik/traefik/pull/8717) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into v2.6 ([#8714](https://github.com/traefik/traefik/pull/8714) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.5 into v2.6 ([#8688](https://github.com/traefik/traefik/pull/8688) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into v2.6 ([#8664](https://github.com/traefik/traefik/pull/8664) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into v2.6 ([#8651](https://github.com/traefik/traefik/pull/8651) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into master ([#8645](https://github.com/traefik/traefik/pull/8645) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into master ([#8609](https://github.com/traefik/traefik/pull/8609) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into master ([#8563](https://github.com/traefik/traefik/pull/8563) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.5 into master ([#8498](https://github.com/traefik/traefik/pull/8498) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.5 into master ([#8461](https://github.com/traefik/traefik/pull/8461) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into master ([#8435](https://github.com/traefik/traefik/pull/8435) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.5 into master ([#8419](https://github.com/traefik/traefik/pull/8419) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.5 into master ([#8411](https://github.com/traefik/traefik/pull/8411) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.5 into master ([#8316](https://github.com/traefik/traefik/pull/8316) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.5 into master ([#8298](https://github.com/traefik/traefik/pull/8298) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into master ([#8289](https://github.com/traefik/traefik/pull/8289) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.5 into master ([#8241](https://github.com/traefik/traefik/pull/8241) by [rtribotte](https://github.com/rtribotte))

## [v2.6.0-rc3](https://github.com/traefik/traefik/tree/v2.6.0-rc3) (2022-01-20)
[All Commits](https://github.com/traefik/traefik/compare/v2.6.0-rc2...v2.6.0-rc3)

**Bug fixes:**
- **[consul]** Support token authentication for Consul KV  ([#8712](https://github.com/traefik/traefik/pull/8712) by [kevinpollet](https://github.com/kevinpollet))

**Misc:**
- Merge current v2.5 into v2.6 ([#8717](https://github.com/traefik/traefik/pull/8717) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into v2.6 ([#8714](https://github.com/traefik/traefik/pull/8714) by [rtribotte](https://github.com/rtribotte))

## [v2.5.7](https://github.com/traefik/traefik/tree/v2.5.7) (2022-01-20)
[All Commits](https://github.com/traefik/traefik/compare/v2.5.6...v2.5.7)

**Bug fixes:**
- **[acme]** Update go-acme/lego to v4.6.0 ([#8716](https://github.com/traefik/traefik/pull/8716) by [ldez](https://github.com/ldez))
- **[logs]** Adjust log level from info to debug ([#8718](https://github.com/traefik/traefik/pull/8718) by [tomMoulard](https://github.com/tomMoulard))
- **[plugins]** Fix middleware plugins memory leak ([#8702](https://github.com/traefik/traefik/pull/8702) by [ldez](https://github.com/ldez))
- **[server]** Mitigate memory leak ([#8706](https://github.com/traefik/traefik/pull/8706) by [mpl](https://github.com/mpl))
- **[webui,middleware]** Fix middleware regexp&#39;s display ([#8697](https://github.com/traefik/traefik/pull/8697) by [tomMoulard](https://github.com/tomMoulard))

**Documentation:**
- **[http]** Fix HTTP provider endpoint config example  ([#8715](https://github.com/traefik/traefik/pull/8715) by [kevinpollet](https://github.com/kevinpollet))
- **[k8s]** Remove typo in Kubernetes providers labelSelector examples ([#8676](https://github.com/traefik/traefik/pull/8676) by [colinwilson](https://github.com/colinwilson))
- **[rules]** Improve regexp matcher documentation ([#8686](https://github.com/traefik/traefik/pull/8686) by [Hades32](https://github.com/Hades32))
- **[tracing]** Fix broken jaeger documentation link ([#8665](https://github.com/traefik/traefik/pull/8665) by [tomMoulard](https://github.com/tomMoulard))
- Update copyright for 2022 ([#8679](https://github.com/traefik/traefik/pull/8679) by [kevinpollet](https://github.com/kevinpollet))

## [v2.6.0-rc2](https://github.com/traefik/traefik/tree/v2.6.0-rc2) (2022-01-12)
[All Commits](https://github.com/traefik/traefik/compare/v2.6.0-rc1...v2.6.0-rc2)

**Bug fixes:**
- **[consul,kv]** Support Consul KV Enterprise namespaces ([#8692](https://github.com/traefik/traefik/pull/8692) by [kevinpollet](https://github.com/kevinpollet))
- **[tracing]** Upgrade Instana tracer dependency ([#8687](https://github.com/traefik/traefik/pull/8687) by [andriikushch](https://github.com/andriikushch))

**Misc:**
- Merge current v2.5 into v2.6 ([#8688](https://github.com/traefik/traefik/pull/8688) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into v2.6 ([#8664](https://github.com/traefik/traefik/pull/8664) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into v2.6 ([#8651](https://github.com/traefik/traefik/pull/8651) by [tomMoulard](https://github.com/tomMoulard))

## [v2.5.6](https://github.com/traefik/traefik/tree/v2.5.6) (2021-12-22)
[All Commits](https://github.com/traefik/traefik/compare/v2.5.5...v2.5.6)

**Bug fixes:**
- **[middleware]** Process all X-Forwarded-For headers in the request ([#8596](https://github.com/traefik/traefik/pull/8596) by [kevinpollet](https://github.com/kevinpollet))
- **[plugins]** Update Yaegi to v0.11.2 ([#8650](https://github.com/traefik/traefik/pull/8650) by [ldez](https://github.com/ldez))
- **[server]** Update golang.org/x/net dependency version ([#8635](https://github.com/traefik/traefik/pull/8635) by [kevinpollet](https://github.com/kevinpollet))

**Documentation:**
- **[api]** Add missing API endpoints documentation ([#8649](https://github.com/traefik/traefik/pull/8649) by [ichxxx](https://github.com/ichxxx))
- **[middleware]** Fix passTLSClientCert CRD example name ([#8637](https://github.com/traefik/traefik/pull/8637) by [ddtmachado](https://github.com/ddtmachado))
- **[middleware]** Correct documentation in middleware overview ([#8636](https://github.com/traefik/traefik/pull/8636) by [Alestrix](https://github.com/Alestrix))

## [v2.6.0-rc1](https://github.com/traefik/traefik/tree/v2.6.0-rc1) (2021-12-20)
[All Commits](https://github.com/traefik/traefik/compare/v2.5.0-rc1...v2.6.0-rc1)

**Enhancements:**
- **[acme]** Allow configuration of ACME certificates duration ([#8046](https://github.com/traefik/traefik/pull/8046) by [pmontepagano](https://github.com/pmontepagano))
- **[consul,consulcatalog]** Support consul enterprise namespaces in consul catalog provider ([#8592](https://github.com/traefik/traefik/pull/8592) by [kevinpollet](https://github.com/kevinpollet))
- **[k8s,k8s/gatewayapi]** Update gateway api provider to v1alpha2 ([#8535](https://github.com/traefik/traefik/pull/8535) by [kevinpollet](https://github.com/kevinpollet))
- **[k8s,k8s/gatewayapi]** Support gateway api RouteNamespaces ([#8299](https://github.com/traefik/traefik/pull/8299) by [tomMoulard](https://github.com/tomMoulard))
- **[k8s/crd]** Support Kubernetes basic-auth secrets ([#8189](https://github.com/traefik/traefik/pull/8189) by [dtomcej](https://github.com/dtomcej))
- **[metrics]** Add configurable tags to influxdb metrics ([#8308](https://github.com/traefik/traefik/pull/8308) by [Tetha](https://github.com/Tetha))
- **[metrics]** Add prefix to datadog metrics ([#8234](https://github.com/traefik/traefik/pull/8234) by [fredwangwang](https://github.com/fredwangwang))
- **[middleware,tcp]** Add in flight connection middleware ([#8429](https://github.com/traefik/traefik/pull/8429) by [tomMoulard](https://github.com/tomMoulard))
- **[middleware]** Add Organizational Unit to passtlscert middleware ([#7958](https://github.com/traefik/traefik/pull/7958) by [FernFerret](https://github.com/FernFerret))
- **[middleware]** Allow configuration of minimum body size for compress middleware ([#8239](https://github.com/traefik/traefik/pull/8239) by [lus](https://github.com/lus))
- **[middleware]** Ceil Retry-After value in the rate-limit middleware ([#8581](https://github.com/traefik/traefik/pull/8581) by [pyaillet](https://github.com/pyaillet))
- **[middleware]** Refactor Exponential Backoff ([#7519](https://github.com/traefik/traefik/pull/7519) by [danieladams456](https://github.com/danieladams456))
- **[server,k8s/crd,k8s]** Allow configuration of HTTP/2 readIdleTimeout and pingTimeout ([#8539](https://github.com/traefik/traefik/pull/8539) by [tomMoulard](https://github.com/tomMoulard))
- **[server]** Allow configuration of advertised port for HTTP/3 ([#8131](https://github.com/traefik/traefik/pull/8131) by [valerauko](https://github.com/valerauko))
- **[tracing]** Upgrade Instana tracer and make process profiling configurable ([#8334](https://github.com/traefik/traefik/pull/8334) by [andriikushch](https://github.com/andriikushch))

**Misc:**
- Merge current v2.5 into master ([#8609](https://github.com/traefik/traefik/pull/8609) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into master ([#8563](https://github.com/traefik/traefik/pull/8563) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.5 into master ([#8498](https://github.com/traefik/traefik/pull/8498) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.5 into master ([#8461](https://github.com/traefik/traefik/pull/8461) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into master ([#8435](https://github.com/traefik/traefik/pull/8435) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.5 into master ([#8419](https://github.com/traefik/traefik/pull/8419) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.5 into master ([#8411](https://github.com/traefik/traefik/pull/8411) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.5 into master ([#8316](https://github.com/traefik/traefik/pull/8316) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.5 into master ([#8298](https://github.com/traefik/traefik/pull/8298) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.5 into master ([#8289](https://github.com/traefik/traefik/pull/8289) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.5 into master ([#8241](https://github.com/traefik/traefik/pull/8241) by [rtribotte](https://github.com/rtribotte))

## [v2.5.5](https://github.com/traefik/traefik/tree/v2.5.5) (2021-12-09)
[All Commits](https://github.com/traefik/traefik/compare/v2.5.4...v2.5.5)

**Bug fixes:**
- **[acme]** Update go-acme/lego to v4.5.3 ([#8607](https://github.com/traefik/traefik/pull/8607) by [lippertmarkus](https://github.com/lippertmarkus))
- **[k8s/crd,k8s]** fix: propagate source criterion config to RateLimit middleware in Kubernetes CRD ([#8591](https://github.com/traefik/traefik/pull/8591) by [rbailly-talend](https://github.com/rbailly-talend))
- **[plugins]** plugins: start the go routine before calling Provide ([#8620](https://github.com/traefik/traefik/pull/8620) by [ldez](https://github.com/ldez))
- **[plugins]** Update yaegi to v0.11.1 ([#8600](https://github.com/traefik/traefik/pull/8600) by [tomMoulard](https://github.com/tomMoulard))
- **[plugins]** Update yaegi v0.11.0 ([#8564](https://github.com/traefik/traefik/pull/8564) by [ldez](https://github.com/ldez))
- **[udp]** fix: increase UDP read buffer length to max datagram size ([#8560](https://github.com/traefik/traefik/pull/8560) by [kevinpollet](https://github.com/kevinpollet))

**Documentation:**
- **[consul]** docs: removing typo in consul-catalog provider doc ([#8603](https://github.com/traefik/traefik/pull/8603) by [tomMoulard](https://github.com/tomMoulard))
- **[metrics]** docs: remove misleading metrics overview configuration ([#8579](https://github.com/traefik/traefik/pull/8579) by [gsilvapt](https://github.com/gsilvapt))
- **[middleware]** docs: align docker configuration example notes in basicauth HTTP middleware ([#8615](https://github.com/traefik/traefik/pull/8615) by [tomMoulard](https://github.com/tomMoulard))
- **[service]** docs: health check use readiness probe in k8s ([#8575](https://github.com/traefik/traefik/pull/8575) by [Vampouille](https://github.com/Vampouille))
- **[tls]** docs: uniformize client TLS config documentation ([#8602](https://github.com/traefik/traefik/pull/8602) by [kevinpollet](https://github.com/kevinpollet))
- Update CODE_OF_CONDUCT.md ([#8619](https://github.com/traefik/traefik/pull/8619) by [tfny](https://github.com/tfny))
- fixed minor spelling error in Regexp Syntax section ([#8565](https://github.com/traefik/traefik/pull/8565) by [kerrsmith](https://github.com/kerrsmith))

## [v2.5.4](https://github.com/traefik/traefik/tree/v2.5.4) (2021-11-08)
[All Commits](https://github.com/traefik/traefik/compare/v2.5.3...v2.5.4)

**Bug fixes:**
- **[acme]** Update go-acme/lego to v4.5.0 ([#8481](https://github.com/traefik/traefik/pull/8481) by [ldez](https://github.com/ldez))
- **[k8s/crd,k8s]** fix: add missing RequireAnyClientCert value to TLSOption CRD ([#8464](https://github.com/traefik/traefik/pull/8464) by [kevinpollet](https://github.com/kevinpollet))
- **[k8s/crd,k8s]** fix: normalize middleware names in ingress route config ([#8484](https://github.com/traefik/traefik/pull/8484) by [aaronraff](https://github.com/aaronraff))
- **[middleware,provider,tls]** fix: do not require a TLS client cert when InsecureSkipVerify is false ([#8525](https://github.com/traefik/traefik/pull/8525) by [kevinpollet](https://github.com/kevinpollet))
- **[middleware,tls]** fix: use host&#39;s root CA set if ClientTLS ca is not defined ([#8545](https://github.com/traefik/traefik/pull/8545) by [kevinpollet](https://github.com/kevinpollet))
- **[middleware]** fix: forward request Host to errors middleware service ([#8460](https://github.com/traefik/traefik/pull/8460) by [kevinpollet](https://github.com/kevinpollet))
- **[middleware]** fix: use EscapedPath as header value when RawPath is empty ([#8251](https://github.com/traefik/traefik/pull/8251) by [dtomcej](https://github.com/dtomcej))
- **[tcp,udp]** fix: TCP/UDP wrr when all servers have a weight set to 0 ([#8553](https://github.com/traefik/traefik/pull/8553) by [tomMoulard](https://github.com/tomMoulard))
- **[webui]** fix: bug parsing weighted service provider name ([#8522](https://github.com/traefik/traefik/pull/8522) by [cocoanton](https://github.com/cocoanton))

**Documentation:**
- **[acme]** docs: remove quotes in certificatesresolvers CLI examples ([#8544](https://github.com/traefik/traefik/pull/8544) by [rdxmb](https://github.com/rdxmb))
- **[k8s/ingress,k8s]** docs: clarify usage for cross provider references in Kubernetes ingress annotations ([#8536](https://github.com/traefik/traefik/pull/8536) by [rtribotte](https://github.com/rtribotte))
- **[k8s/ingress]** docs: networking.k8s.io/v1beta1 to networking.k8s.io/v1 ([#8523](https://github.com/traefik/traefik/pull/8523) by [pmareke](https://github.com/pmareke))
- **[k8s]** docs: replace links to French translation of k8s docs with English ones ([#8457](https://github.com/traefik/traefik/pull/8457) by [FoseFx](https://github.com/FoseFx))
- **[k8s]** docs: remove non-working kind config in IngressRouteTCP/UDP examples ([#8538](https://github.com/traefik/traefik/pull/8538) by [kevinpollet](https://github.com/kevinpollet))
- **[kv]** docs: fix typo in KV providers documentation ([#8477](https://github.com/traefik/traefik/pull/8477) by [rondoe](https://github.com/rondoe))
- **[metrics]** docs: fix typo in addRoutersLabels option title ([#8561](https://github.com/traefik/traefik/pull/8561) by [kevinpollet](https://github.com/kevinpollet))
- **[middleware]** fix: sourceCriterion documentation for InFlightReq and RateLimit middlewares ([#8524](https://github.com/traefik/traefik/pull/8524) by [pmareke](https://github.com/pmareke))
- **[middleware]** Mention escaping escape characters in YAML for regex usage ([#8496](https://github.com/traefik/traefik/pull/8496) by [JackMorganNZ](https://github.com/JackMorganNZ))
- **[rules]** docs: add named groups details to Regexp Syntax section ([#8559](https://github.com/traefik/traefik/pull/8559) by [kerrsmith](https://github.com/kerrsmith))
- **[tracing]** docs: reword tracing config descriptions to be consistent ([#8473](https://github.com/traefik/traefik/pull/8473) by [kevinpollet](https://github.com/kevinpollet))
- docs: remove link to microbadger.com ([#8555](https://github.com/traefik/traefik/pull/8555) by [CrispyBaguette](https://github.com/CrispyBaguette))
- docs: remove http scheme urls in documentation ([#8507](https://github.com/traefik/traefik/pull/8507) by [tomMoulard](https://github.com/tomMoulard))
- docs: update traefik image version ([#8533](https://github.com/traefik/traefik/pull/8533) by [tomMoulard](https://github.com/tomMoulard))

## [v2.5.3](https://github.com/traefik/traefik/tree/v2.5.3) (2021-09-20)
[All Commits](https://github.com/traefik/traefik/compare/v2.5.2...v2.5.3)

**Bug fixes:**
- **[consulcatalog]** Fix certChan defaulting on consul catalog provider ([#8439](https://github.com/traefik/traefik/pull/8439) by [tomMoulard](https://github.com/tomMoulard))
- **[k8s/crd,k8s]** Fix peerCertURI config for k8s crd provider  ([#8454](https://github.com/traefik/traefik/pull/8454) by [kevinpollet](https://github.com/kevinpollet))
- **[k8s/crd,k8s]** Ensure disableHTTP2 works with k8s crd ([#8448](https://github.com/traefik/traefik/pull/8448) by [ssboisen](https://github.com/ssboisen))
- **[k8s/crd,k8s]** Fix ServersTransport reference from IngressRoute service definition ([#8431](https://github.com/traefik/traefik/pull/8431) by [rtribotte](https://github.com/rtribotte))
- **[k8s/crd,k8s]** Add cross namespace verification in Kubernetes CRD ([#8422](https://github.com/traefik/traefik/pull/8422) by [tomMoulard](https://github.com/tomMoulard))
- **[metrics]** Fix Prometheus router&#39;s metrics ([#8425](https://github.com/traefik/traefik/pull/8425) by [tomMoulard](https://github.com/tomMoulard))
- **[plugins]** Update yaegi to v0.10.0 ([#8452](https://github.com/traefik/traefik/pull/8452) by [ldez](https://github.com/ldez))

**Documentation:**
- **[middleware,file]** Fix TCP middleware whitelist example ([#8421](https://github.com/traefik/traefik/pull/8421) by [tribal2](https://github.com/tribal2))
- **[middleware]** Add default proxy headers list ([#8418](https://github.com/traefik/traefik/pull/8418) by [aaronraff](https://github.com/aaronraff))
- Add Tom Moulard in maintainers team ([#8442](https://github.com/traefik/traefik/pull/8442) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Fix golang doc URLs ([#8434](https://github.com/traefik/traefik/pull/8434) by [jbdoumenjou](https://github.com/jbdoumenjou))

## [v2.5.2](https://github.com/traefik/traefik/tree/v2.5.2) (2021-09-02)
[All Commits](https://github.com/traefik/traefik/compare/v2.5.1...v2.5.2)

**Bug fixes:**
- **[http3]** Upgrade github.com/lucas-clemente/quic-go to v0.23.0 ([#8413](https://github.com/traefik/traefik/pull/8413) by [sylr](https://github.com/sylr))
- **[middleware]** Fix empty body error for mirroring middleware ([#8381](https://github.com/traefik/traefik/pull/8381) by [antgubarev](https://github.com/antgubarev))
- **[tracing]** Bump go.elastic.co/apm version to v1.13.1 ([#8399](https://github.com/traefik/traefik/pull/8399) by [rtribotte](https://github.com/rtribotte))
- Update x/sys to support go 1.17 ([#8368](https://github.com/traefik/traefik/pull/8368) by [roopakv](https://github.com/roopakv))
- Bump Alpine docker image version from 3.11 to 3.14 for official Traefik images

**Documentation:**
- **[k8s/ingress,k8s]** Adds pathType for v1 ingresses examples ([#8392](https://github.com/traefik/traefik/pull/8392) by [rtribotte](https://github.com/rtribotte))
- Fix http scheme urls in documentation ([#8395](https://github.com/traefik/traefik/pull/8395) by [rtribotte](https://github.com/rtribotte))

## [v2.5.1](https://github.com/traefik/traefik/tree/v2.5.1) (2021-08-20)
[All Commits](https://github.com/traefik/traefik/compare/v2.5.0...v2.5.1)

**Bug fixes:**
- **[middleware,http3]** Conditional CloseNotify in header middleware ([#8374](https://github.com/traefik/traefik/pull/8374) by [juliens](https://github.com/juliens))
- **[tls,tcp,k8s/crd,k8s]** Makes ALPN protocols configurable ([#8383](https://github.com/traefik/traefik/pull/8383) by [rtribotte](https://github.com/rtribotte))

**Documentation:**
- **[k8s]** Adds MiddlewareTCP CRD documentation ([#8369](https://github.com/traefik/traefik/pull/8369) by [perosb](https://github.com/perosb))
- **[middleware]** Adds ContentType to middleware&#39;s overview table ([#8350](https://github.com/traefik/traefik/pull/8350) by [euidong](https://github.com/euidong))

## [v2.5.0](https://github.com/traefik/traefik/tree/v2.5.0) (2021-08-17)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.0-rc1...v2.5.0)

**Enhancements:**
- **[consulcatalog]** Add Support for Consul Connect ([#7407](https://github.com/traefik/traefik/pull/7407) by [Gufran](https://github.com/Gufran))
- Update Go version ([#8355](https://github.com/traefik/traefik/pull/8355) by [mpl](https://github.com/mpl))
- **[file]** Update sprig to v3.2.0 ([#7746](https://github.com/traefik/traefik/pull/7746) by [sirlatrom](https://github.com/sirlatrom))
- **[healthcheck]** Healthcheck: add support at the load-balancers of services level ([#8057](https://github.com/traefik/traefik/pull/8057) by [mpl](https://github.com/mpl))
- **[http3]** Upgrade github.com/lucas-clemente/quic-go ([#8076](https://github.com/traefik/traefik/pull/8076) by [sylr](https://github.com/sylr))
- **[http3]** Add HTTP3 support (experimental) ([#7724](https://github.com/traefik/traefik/pull/7724) by [juliens](https://github.com/juliens))
- **[k8s,k8s/gatewayapi]** Add wildcard hostname rule to kubernetes gateway ([#7963](https://github.com/traefik/traefik/pull/7963) by [jberger](https://github.com/jberger))
- **[k8s,k8s/gatewayapi]** Add support for TCPRoute and TLSRoute ([#8054](https://github.com/traefik/traefik/pull/8054) by [tomMoulard](https://github.com/tomMoulard))
- **[k8s,k8s/gatewayapi]** Allow crossprovider service reference ([#7774](https://github.com/traefik/traefik/pull/7774) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s/crd,k8s]** Add named port support to Kubernetes IngressRoute CRDs ([#7668](https://github.com/traefik/traefik/pull/7668) by [Cirrith](https://github.com/Cirrith))
- **[k8s/crd,k8s]** Improve kubernetes external name service support for UDP ([#7773](https://github.com/traefik/traefik/pull/7773) by [kevinpollet](https://github.com/kevinpollet))
- **[k8s/crd,k8s]** Upgrade the CRD version from apiextensions.k8s.io/v1beta1 to apiextensions.k8s.io/v1 ([#7815](https://github.com/traefik/traefik/pull/7815) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s/ingress,k8s/crd,k8s]** Ignore empty endpoint changes ([#7646](https://github.com/traefik/traefik/pull/7646) by [hensur](https://github.com/hensur))
- **[k8s/ingress,k8s]** Upgrade Ingress Handling to work with networkingv1/Ingress ([#7549](https://github.com/traefik/traefik/pull/7549) by [SantoDE](https://github.com/SantoDE))
- **[k8s/ingress,k8s]** Filter ingress class resources by name ([#7915](https://github.com/traefik/traefik/pull/7915) by [tomMoulard](https://github.com/tomMoulard))
- **[k8s/ingress,k8s]** Add k8s provider option to create services without endpoints ([#7593](https://github.com/traefik/traefik/pull/7593) by [Lucaber](https://github.com/Lucaber))
- **[k8s/ingress,k8s]** Upgrade IngressClass to use v1 over v1Beta on Kube 1.19+ ([#8089](https://github.com/traefik/traefik/pull/8089) by [SantoDE](https://github.com/SantoDE))
- **[k8s/ingress,k8s]** Add ServersTransport annotation to k8s ingress provider ([#8084](https://github.com/traefik/traefik/pull/8084) by [wdullaer](https://github.com/wdullaer))
- **[logs,middleware]** Add TLS version and cipher to the accessLog ([#7478](https://github.com/traefik/traefik/pull/7478) by [na4ma4](https://github.com/na4ma4))
- **[metrics]** Add TLS certs expiration metric ([#6924](https://github.com/traefik/traefik/pull/6924) by [sylr](https://github.com/sylr))
- **[metrics]** Allow to define datadogs metrics endpoint with env vars ([#7968](https://github.com/traefik/traefik/pull/7968) by [sylr](https://github.com/sylr))
- **[middleware,metrics]** Add router metrics ([#7510](https://github.com/traefik/traefik/pull/7510) by [jorge07](https://github.com/jorge07))
- **[middleware,tcp]** Add TCP Middlewares support ([#7813](https://github.com/traefik/traefik/pull/7813) by [rtribotte](https://github.com/rtribotte))
- **[middleware]** Removes  headers middleware options  ([#8161](https://github.com/traefik/traefik/pull/8161) by [tomMoulard](https://github.com/tomMoulard))
- **[middleware]** Headers: add `permissionsPolicy` and deprecate `featurePolicy` ([#8200](https://github.com/traefik/traefik/pull/8200) by [WLun001](https://github.com/WLun001))
- **[middleware]** Deprecates ssl redirect headers middleware options ([#8160](https://github.com/traefik/traefik/pull/8160) by [tomMoulard](https://github.com/tomMoulard))
- **[plugins]** Local private plugins. ([#8224](https://github.com/traefik/traefik/pull/8224) by [ldez](https://github.com/ldez))
- **[provider,plugins]** Add plugin&#39;s support for provider ([#7794](https://github.com/traefik/traefik/pull/7794) by [ldez](https://github.com/ldez))
- **[rules]** Support not in rules definition ([#8164](https://github.com/traefik/traefik/pull/8164) by [juliens](https://github.com/juliens))
- **[rules]** Add routing IP rule matcher ([#8169](https://github.com/traefik/traefik/pull/8169) by [tomMoulard](https://github.com/tomMoulard))
- **[server]** Improve host name resolution for TCP proxy ([#7971](https://github.com/traefik/traefik/pull/7971) by [H-M-H](https://github.com/H-M-H))
- **[server]** Add ability to disable HTTP/2 in dynamic config ([#7645](https://github.com/traefik/traefik/pull/7645) by [jcuzzi](https://github.com/jcuzzi))
- **[sticky-session]** Add a mechanism to format the sticky cookie value ([#8103](https://github.com/traefik/traefik/pull/8103) by [tomMoulard](https://github.com/tomMoulard))
- **[tls]** Mutualize TLS version and cipher code ([#7779](https://github.com/traefik/traefik/pull/7779) by [rtribotte](https://github.com/rtribotte))
- **[tls,k8s/crd,k8s]** Improve CA certificate loading from kubernetes secret ([#7789](https://github.com/traefik/traefik/pull/7789) by [rio](https://github.com/rio))
- **[tls]** Do not build a default certificate for ACME challenges store ([#7833](https://github.com/traefik/traefik/pull/7833) by [rkojedzinszky](https://github.com/rkojedzinszky))
- **[tracing]** Use Datadog tracer environment variables to setup default config ([#7721](https://github.com/traefik/traefik/pull/7721) by [GianOrtiz](https://github.com/GianOrtiz))
- **[tracing]** Update Elastic APM from 1.7.0 to 1.11.0 ([#8187](https://github.com/traefik/traefik/pull/8187) by [afitzek](https://github.com/afitzek))
- **[tracing]** Override jaeger configuration with env variables ([#8198](https://github.com/traefik/traefik/pull/8198) by [mmatur](https://github.com/mmatur))
- **[udp]** Add udp timeout configuration ([#6982](https://github.com/traefik/traefik/pull/6982) by [Lindenk](https://github.com/Lindenk))

**Bug fixes:**
- **[k8s,k8s/gatewayapi]** Update Gateway API version to v0.3.0 ([#8253](https://github.com/traefik/traefik/pull/8253) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s]** Kubernetes: detect changes for resources other than endpoints ([#8313](https://github.com/traefik/traefik/pull/8313) by [rtribotte](https://github.com/rtribotte))
- **[middleware]** Library change for compress middleware to increase performance ([#8245](https://github.com/traefik/traefik/pull/8245) by [tomMoulard](https://github.com/tomMoulard))
- **[plugins]** Update yaegi to v0.9.21 ([#8285](https://github.com/traefik/traefik/pull/8285) by [ldez](https://github.com/ldez))
- **[plugins]** Downgrade yaegi to v0.9.19 ([#8282](https://github.com/traefik/traefik/pull/8282) by [ldez](https://github.com/ldez))
- **[webui]** Fix dashboard to display middleware details ([#8284](https://github.com/traefik/traefik/pull/8284) by [tomMoulard](https://github.com/tomMoulard))
- **[webui]** Fix dashboard title for TCP middlewares ([#8339](https://github.com/traefik/traefik/pull/8339) by [mschneider82](https://github.com/mschneider82))
- **[k8s]** Remove logging of changed object with cast ([#8128](https://github.com/traefik/traefik/pull/8128) by [hensur](https://github.com/hensur))

**Documentation:**
- Fix KV reference documentation ([#8280](https://github.com/traefik/traefik/pull/8280) by [rtribotte](https://github.com/rtribotte))
- Fix migration guide ([#8269](https://github.com/traefik/traefik/pull/8269) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Update generated and reference doc for plugins ([#8236](https://github.com/traefik/traefik/pull/8236) by [tomMoulard](https://github.com/tomMoulard))
- **[k8s/crd]** Fix: regenerate crd ([#8114](https://github.com/traefik/traefik/pull/8114) by [tomMoulard](https://github.com/tomMoulard))
- **[k8s]** Clarify doc for ingressclass name in k8s 1.18+ ([#7944](https://github.com/traefik/traefik/pull/7944) by [tomMoulard](https://github.com/tomMoulard))
- Update documentation references ([#8202](https://github.com/traefik/traefik/pull/8202) by [rtribotte](https://github.com/rtribotte))

**Misc:**
- Merge current v2.4 into v2.5 ([#8333](https://github.com/traefik/traefik/pull/8333) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.4 into v2.5 ([#8325](https://github.com/traefik/traefik/pull/8325) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.4 into v2.5 ([#8314](https://github.com/traefik/traefik/pull/8314) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into v2.5 ([#8296](https://github.com/traefik/traefik/pull/8296) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.4 into v2.5 ([#8287](https://github.com/traefik/traefik/pull/8287) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into v2.5 ([#8281](https://github.com/traefik/traefik/pull/8281) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into v2.5 ([#8263](https://github.com/traefik/traefik/pull/8263) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into master ([#8232](https://github.com/traefik/traefik/pull/8232) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into master ([#8210](https://github.com/traefik/traefik/pull/8210) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into master ([#8105](https://github.com/traefik/traefik/pull/8105) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.4 into master ([#8087](https://github.com/traefik/traefik/pull/8087) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.4 into master ([#8068](https://github.com/traefik/traefik/pull/8068) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.4 into master ([#8058](https://github.com/traefik/traefik/pull/8058) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.4 into master ([#8024](https://github.com/traefik/traefik/pull/8024) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.4 into master ([#7969](https://github.com/traefik/traefik/pull/7969) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.4 into master ([#7921](https://github.com/traefik/traefik/pull/7921) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into master ([#7901](https://github.com/traefik/traefik/pull/7901) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into master ([#7859](https://github.com/traefik/traefik/pull/7859) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.4 into master ([#7795](https://github.com/traefik/traefik/pull/7795) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.4 into master ([#8221](https://github.com/traefik/traefik/pull/8221) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into master ([#7781](https://github.com/traefik/traefik/pull/7781) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.4 into master ([#7766](https://github.com/traefik/traefik/pull/7766) by [ldez](https://github.com/ldez))
- Merge current v2.4 into master ([#7761](https://github.com/traefik/traefik/pull/7761) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into master ([#7748](https://github.com/traefik/traefik/pull/7748) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.4 into master ([#7728](https://github.com/traefik/traefik/pull/7728) by [mmatur](https://github.com/mmatur))


## [v2.4.14](https://github.com/traefik/traefik/tree/v2.4.14) (2021-08-16)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.13...v2.4.14)

**Bug fixes:**
- **[k8s/crd,k8s]** Avoid unauthorized middleware cross namespace reference  ([#8322](https://github.com/traefik/traefik/pull/8322) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[kv]** Remove unwanted trailing slash in key ([#8335](https://github.com/traefik/traefik/pull/8335) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware]** Redirect: fix comparison when explicit port request and implicit redirect port ([#8348](https://github.com/traefik/traefik/pull/8348) by [tcolgate](https://github.com/tcolgate))

**Documentation:**
- **[kv]** Fix a router&#39;s entryPoint definition example for KV provider ([#8357](https://github.com/traefik/traefik/pull/8357) by [avtion](https://github.com/avtion))

## [v2.5.0-rc6](https://github.com/traefik/traefik/tree/v2.5.0-rc6) (2021-08-13)
[All Commits](https://github.com/traefik/traefik/compare/v2.5.0-rc5...v2.5.0-rc6)

**Enhancements:**
- Update Go version ([#8355](https://github.com/traefik/traefik/pull/8355) by [mpl](https://github.com/mpl))

**Misc:**
- Merge current v2.4 into v2.5 ([#8333](https://github.com/traefik/traefik/pull/8333) by [jbdoumenjou](https://github.com/jbdoumenjou))

## [v2.5.0-rc5](https://github.com/traefik/traefik/tree/v2.5.0-rc5) (2021-08-03)
[All Commits](https://github.com/traefik/traefik/compare/v2.5.0-rc3...v2.5.0-rc5)

**Bug fixes:**
- **[k8s]** Kubernetes: detect changes for resources other than endpoints ([#8313](https://github.com/traefik/traefik/pull/8313) by [rtribotte](https://github.com/rtribotte))

**Misc:**
- Merge current v2.4 into v2.5 ([#8325](https://github.com/traefik/traefik/pull/8325) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.4 into v2.5 ([#8314](https://github.com/traefik/traefik/pull/8314) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into v2.5 ([#8296](https://github.com/traefik/traefik/pull/8296) by [tomMoulard](https://github.com/tomMoulard))

## [v2.5.0-rc4](https://github.com/traefik/traefik/tree/v2.5.0-rc4) (2021-08-03)

Release canceled.

## [v2.4.13](https://github.com/traefik/traefik/tree/v2.4.13) (2021-07-30)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.12...v2.4.13)

**Bug fixes:**
- **[authentication,middleware]** Remove hop-by-hop headers defined in connection header before some middleware ([#8319](https://github.com/traefik/traefik/pull/8319) by [ldez](https://github.com/ldez))

## [v2.4.12](https://github.com/traefik/traefik/tree/v2.4.12) (2021-07-26)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.11...v2.4.12)

**Bug fixes:**
- **[k8s,k8s/ingress]** Get Kubernetes server version early ([#8286](https://github.com/traefik/traefik/pull/8286) by [rtribotte](https://github.com/rtribotte))
- **[k8s,k8s/ingress]** Don&#39;t remove ingress config on API call failure ([#8185](https://github.com/traefik/traefik/pull/8185) by [dtomcej](https://github.com/dtomcej))
- **[middleware]** Ratelimiter: use correct ttlSeconds value, and always call Set ([#8254](https://github.com/traefik/traefik/pull/8254) by [mpl](https://github.com/mpl))
- **[tls]** Check if defaultcertificate is defined in store ([#8274](https://github.com/traefik/traefik/pull/8274) by [dtomcej](https://github.com/dtomcej))

## [v2.5.0-rc3](https://github.com/traefik/traefik/tree/v2.5.0-rc3) (2021-07-20)
[All Commits](https://github.com/traefik/traefik/compare/v2.5.0-rc2...v2.5.0-rc3)

**Enhancements:**
- **[consulcatalog]** Add Support for Consul Connect ([#7407](https://github.com/traefik/traefik/pull/7407) by [Gufran](https://github.com/Gufran))

**Bug fixes:**
- **[k8s,k8s/gatewayapi]** Update Gateway API version to v0.3.0 ([#8253](https://github.com/traefik/traefik/pull/8253) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware]** Library change for compress middleware to increase performance ([#8245](https://github.com/traefik/traefik/pull/8245) by [tomMoulard](https://github.com/tomMoulard))
- **[plugins]** Update yaegi to v0.9.21 ([#8285](https://github.com/traefik/traefik/pull/8285) by [ldez](https://github.com/ldez))
- **[plugins]** Downgrade yaegi to v0.9.19 ([#8282](https://github.com/traefik/traefik/pull/8282) by [ldez](https://github.com/ldez))
- **[webui]** Fix dashboard to display middleware details ([#8284](https://github.com/traefik/traefik/pull/8284) by [tomMoulard](https://github.com/tomMoulard))

**Documentation:**
- Fix KV reference documentation ([#8280](https://github.com/traefik/traefik/pull/8280) by [rtribotte](https://github.com/rtribotte))
- Fix migration guide ([#8269](https://github.com/traefik/traefik/pull/8269) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Update generated and reference doc for plugins ([#8236](https://github.com/traefik/traefik/pull/8236) by [tomMoulard](https://github.com/tomMoulard))

**Misc:**
- Merge current v2.4 into v2.5 ([#8263](https://github.com/traefik/traefik/pull/8263) by [rtribotte](https://github.com/rtribotte))

## [v2.4.11](https://github.com/traefik/traefik/tree/v2.4.11) (2021-07-15)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.9...v2.4.11)

**Bug fixes:**
- **[k8s,k8s/crd,k8s/ingress]** Disable ExternalName Services by default on Kubernetes providers ([#8261](https://github.com/traefik/traefik/pull/8261) by [dtomcej](https://github.com/dtomcej))
- **[k8s,k8s/crd,k8s/ingress]** Fix: malformed Kubernetes resource names and references in tests ([#8226](https://github.com/traefik/traefik/pull/8226) by [rtribotte](https://github.com/rtribotte))
- **[k8s,k8s/crd]** Disable Cross-Namespace by default for IngressRoute provider ([#8260](https://github.com/traefik/traefik/pull/8260) by [dtomcej](https://github.com/dtomcej))
- **[logs,middleware]** Accesslog: support multiple values for a given header ([#8258](https://github.com/traefik/traefik/pull/8258) by [ldez](https://github.com/ldez))
- **[logs]** Ignore http 1.0 request host missing errors ([#8252](https://github.com/traefik/traefik/pull/8252) by [dtomcej](https://github.com/dtomcej))
- **[middleware]** Headers Middleware: support http.CloseNotifier interface ([#8238](https://github.com/traefik/traefik/pull/8238) by [dtomcej](https://github.com/dtomcej))
- **[tls]** Detect certificates content modifications ([#8243](https://github.com/traefik/traefik/pull/8243) by [jbdoumenjou](https://github.com/jbdoumenjou))

**Documentation:**
- **[middleware,k8s]** Fix invalid subdomain ([#8212](https://github.com/traefik/traefik/pull/8212) by [WLun001](https://github.com/WLun001))
- Add the list of available provider names ([#8225](https://github.com/traefik/traefik/pull/8225) by [WLun001](https://github.com/WLun001))
- Fix maintainers-guidelines page title ([#8216](https://github.com/traefik/traefik/pull/8216) by [kubopanda](https://github.com/kubopanda))
- Typos in contributing section ([#8215](https://github.com/traefik/traefik/pull/8215) by [kubopanda](https://github.com/kubopanda))

## [v2.4.10](https://github.com/traefik/traefik/tree/v2.4.10) (2021-07-13)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.9...v2.4.10)

Release canceled.

## [v2.5.0-rc2](https://github.com/traefik/traefik/tree/v2.5.0-rc2) (2021-06-28)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.0-rc1...v2.5.0-rc2)

**Enhancements:**
- **[file]** Update sprig to v3.2.0 ([#7746](https://github.com/traefik/traefik/pull/7746) by [sirlatrom](https://github.com/sirlatrom))
- **[healthcheck]** Healthcheck: add support at the load-balancers of services level ([#8057](https://github.com/traefik/traefik/pull/8057) by [mpl](https://github.com/mpl))
- **[http3]** Upgrade github.com/lucas-clemente/quic-go ([#8076](https://github.com/traefik/traefik/pull/8076) by [sylr](https://github.com/sylr))
- **[http3]** Add HTTP3 support (experimental) ([#7724](https://github.com/traefik/traefik/pull/7724) by [juliens](https://github.com/juliens))
- **[k8s,k8s/crd,k8s/ingress]** Ignore empty endpoint changes ([#7646](https://github.com/traefik/traefik/pull/7646) by [hensur](https://github.com/hensur))
- **[k8s,k8s/crd]** Improve kubernetes external name service support for UDP ([#7773](https://github.com/traefik/traefik/pull/7773) by [kevinpollet](https://github.com/kevinpollet))
- **[k8s,k8s/crd]** Upgrade the CRD version from apiextensions.k8s.io/v1beta1 to apiextensions.k8s.io/v1 ([#7815](https://github.com/traefik/traefik/pull/7815) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/crd]** Add named port support to Kubernetes IngressRoute CRDs ([#7668](https://github.com/traefik/traefik/pull/7668) by [Cirrith](https://github.com/Cirrith))
- **[k8s,k8s/gatewayapi]** Add wildcard hostname rule to kubernetes gateway ([#7963](https://github.com/traefik/traefik/pull/7963) by [jberger](https://github.com/jberger))
- **[k8s,k8s/gatewayapi]** Allow crossprovider service reference ([#7774](https://github.com/traefik/traefik/pull/7774) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/gatewayapi]** Add support for TCPRoute and TLSRoute ([#8054](https://github.com/traefik/traefik/pull/8054) by [tomMoulard](https://github.com/tomMoulard))
- **[k8s,k8s/ingress]** Filter ingress class resources by name ([#7915](https://github.com/traefik/traefik/pull/7915) by [tomMoulard](https://github.com/tomMoulard))
- **[k8s,k8s/ingress]** Upgrade Ingress Handling to work with networkingv1/Ingress ([#7549](https://github.com/traefik/traefik/pull/7549) by [SantoDE](https://github.com/SantoDE))
- **[k8s,k8s/ingress]** Upgrade IngressClass to use v1 over v1Beta on Kube 1.19+ ([#8089](https://github.com/traefik/traefik/pull/8089) by [SantoDE](https://github.com/SantoDE))
- **[k8s,k8s/ingress]** Add k8s provider option to create services without endpoints ([#7593](https://github.com/traefik/traefik/pull/7593) by [Lucaber](https://github.com/Lucaber))
- **[k8s,k8s/ingress]** Add ServersTransport annotation to k8s ingress provider ([#8084](https://github.com/traefik/traefik/pull/8084) by [wdullaer](https://github.com/wdullaer))
- **[logs,middleware]** Add TLS version and cipher to the accessLog ([#7478](https://github.com/traefik/traefik/pull/7478) by [na4ma4](https://github.com/na4ma4))
- **[metrics]** Allow to define datadogs metrics endpoint with env vars ([#7968](https://github.com/traefik/traefik/pull/7968) by [sylr](https://github.com/sylr))
- **[metrics]** Add TLS certs expiration metric ([#6924](https://github.com/traefik/traefik/pull/6924) by [sylr](https://github.com/sylr))
- **[middleware,metrics]** Add router metrics ([#7510](https://github.com/traefik/traefik/pull/7510) by [jorge07](https://github.com/jorge07))
- **[middleware,tcp]** Add TCP Middlewares support ([#7813](https://github.com/traefik/traefik/pull/7813) by [rtribotte](https://github.com/rtribotte))
- **[middleware]** Deprecates ssl redirect headers middleware options ([#8160](https://github.com/traefik/traefik/pull/8160) by [tomMoulard](https://github.com/tomMoulard))
- **[middleware]** Headers: add `permissionsPolicy` and deprecate `featurePolicy` ([#8200](https://github.com/traefik/traefik/pull/8200) by [WLun001](https://github.com/WLun001))
- **[middleware]** Removes  headers middleware options  ([#8161](https://github.com/traefik/traefik/pull/8161) by [tomMoulard](https://github.com/tomMoulard))
- **[plugins,provider]** Add plugin&#39;s support for provider ([#7794](https://github.com/traefik/traefik/pull/7794) by [ldez](https://github.com/ldez))
- **[plugins]** Local private plugins. ([#8224](https://github.com/traefik/traefik/pull/8224) by [ldez](https://github.com/ldez))
- **[rules]** Add routing IP rule matcher ([#8169](https://github.com/traefik/traefik/pull/8169) by [tomMoulard](https://github.com/tomMoulard))
- **[rules]** Support not in rules definition ([#8164](https://github.com/traefik/traefik/pull/8164) by [juliens](https://github.com/juliens))
- **[server]** Improve host name resolution for TCP proxy ([#7971](https://github.com/traefik/traefik/pull/7971) by [H-M-H](https://github.com/H-M-H))
- **[server]** Add ability to disable HTTP/2 in dynamic config ([#7645](https://github.com/traefik/traefik/pull/7645) by [jcuzzi](https://github.com/jcuzzi))
- **[sticky-session]** Add a mechanism to format the sticky cookie value ([#8103](https://github.com/traefik/traefik/pull/8103) by [tomMoulard](https://github.com/tomMoulard))
- **[tls]** Mutualize TLS version and cipher code ([#7779](https://github.com/traefik/traefik/pull/7779) by [rtribotte](https://github.com/rtribotte))
- **[tls]** Do not build a default certificate for ACME challenges store ([#7833](https://github.com/traefik/traefik/pull/7833) by [rkojedzinszky](https://github.com/rkojedzinszky))
- **[tracing]** Use Datadog tracer environment variables to setup default config ([#7721](https://github.com/traefik/traefik/pull/7721) by [GianOrtiz](https://github.com/GianOrtiz))
- **[tracing]** Update Elastic APM from 1.7.0 to 1.11.0 ([#8187](https://github.com/traefik/traefik/pull/8187) by [afitzek](https://github.com/afitzek))
- **[tracing]** Override jaeger configuration with env variables ([#8198](https://github.com/traefik/traefik/pull/8198) by [mmatur](https://github.com/mmatur))
- **[udp]** Add udp timeout configuration ([#6982](https://github.com/traefik/traefik/pull/6982) by [Lindenk](https://github.com/Lindenk))

**Bug fixes:**
- **[k8s]** Remove logging of changed object with cast ([#8128](https://github.com/traefik/traefik/pull/8128) by [hensur](https://github.com/hensur))

**Documentation:**
- **[k8s/crd]** Fix: regenerate crd ([#8114](https://github.com/traefik/traefik/pull/8114) by [tomMoulard](https://github.com/tomMoulard))
- **[k8s]** Clarify doc for ingressclass name in k8s 1.18+ ([#7944](https://github.com/traefik/traefik/pull/7944) by [tomMoulard](https://github.com/tomMoulard))
- Update documentation references ([#8202](https://github.com/traefik/traefik/pull/8202) by [rtribotte](https://github.com/rtribotte))

**Misc:**
- **[k8s,k8s/crd,tls]** Improve CA certificate loading from kubernetes secret ([#7789](https://github.com/traefik/traefik/pull/7789) by [rio](https://github.com/rio))
- Merge current v2.4 into master ([#8221](https://github.com/traefik/traefik/pull/8221) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into master ([#8105](https://github.com/traefik/traefik/pull/8105) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.4 into master ([#8087](https://github.com/traefik/traefik/pull/8087) by [tomMoulard](https://github.com/tomMoulard))
- Merge current v2.4 into master ([#8068](https://github.com/traefik/traefik/pull/8068) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.4 into master ([#8058](https://github.com/traefik/traefik/pull/8058) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.4 into master ([#8024](https://github.com/traefik/traefik/pull/8024) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.4 into master ([#7969](https://github.com/traefik/traefik/pull/7969) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.4 into master ([#7921](https://github.com/traefik/traefik/pull/7921) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into master ([#7901](https://github.com/traefik/traefik/pull/7901) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into master ([#7859](https://github.com/traefik/traefik/pull/7859) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.4 into master ([#7795](https://github.com/traefik/traefik/pull/7795) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.4 into master ([#8210](https://github.com/traefik/traefik/pull/8210) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into master ([#7781](https://github.com/traefik/traefik/pull/7781) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.4 into master ([#7766](https://github.com/traefik/traefik/pull/7766) by [ldez](https://github.com/ldez))
- Merge current v2.4 into master ([#7761](https://github.com/traefik/traefik/pull/7761) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.4 into master ([#7748](https://github.com/traefik/traefik/pull/7748) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.4 into master ([#7728](https://github.com/traefik/traefik/pull/7728) by [mmatur](https://github.com/mmatur))

## [v2.5.0-rc1](https://github.com/traefik/traefik/tree/v2.5.0-rc1) (2021-06-28)

Release canceled.

## [v2.4.9](https://github.com/traefik/traefik/tree/v2.4.9) (2021-06-21)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.8...v2.4.9)

**Bug fixes:**
- **[acme]** Update go-acme/lego to v4.4.0 ([#8179](https://github.com/traefik/traefik/pull/8179) by [ldez](https://github.com/ldez))
- **[acme]** Fix: ACME preferred chain. ([#8146](https://github.com/traefik/traefik/pull/8146) by [ldez](https://github.com/ldez))
- **[k8s,k8s/gatewayapi]** Remove error when HTTProutes is empty ([#8023](https://github.com/traefik/traefik/pull/8023) by [tomMoulard](https://github.com/tomMoulard))
- **[k8s,k8s/ingress]** Fix incorrect behaviour with multi-port endpoint subsets ([#8156](https://github.com/traefik/traefik/pull/8156) by [coufalja](https://github.com/coufalja))
- **[k8s,k8s/ingress]** Kubernetes ingress provider to search via all endpoints ([#7997](https://github.com/traefik/traefik/pull/7997) by [martinvizvary](https://github.com/martinvizvary))
- **[plugins,windows]** Fix plugin unzip call on windows ([#8136](https://github.com/traefik/traefik/pull/8136) by [ddtmachado](https://github.com/ddtmachado))
- **[plugins]** Update Yaegi to v0.9.17 ([#8100](https://github.com/traefik/traefik/pull/8100) by [ldez](https://github.com/ldez))
- **[provider]** Bump paerser to v0.1.4 ([#8116](https://github.com/traefik/traefik/pull/8116) by [ldez](https://github.com/ldez))
- **[server]** Create buffered signals channel ([#8190](https://github.com/traefik/traefik/pull/8190) by [dtomcej](https://github.com/dtomcej))
- **[server]** Fix: use defaultEntryPoints when no entryPoint is defined in a TCPRouter ([#8111](https://github.com/traefik/traefik/pull/8111) by [LandryBe](https://github.com/LandryBe))
- **[tls]** Use a dynamic buffer to handle client Hello SNI detection ([#8194](https://github.com/traefik/traefik/pull/8194) by [ldez](https://github.com/ldez))
- **[tracing]** Error span on 5xx only ([#8033](https://github.com/traefik/traefik/pull/8033) by [kevtainer](https://github.com/kevtainer))

**Documentation:**
- **[k8s,k8s/crd]** Fix ingressRouteTCP external name service examples in documentation ([#8120](https://github.com/traefik/traefik/pull/8120) by [rtribotte](https://github.com/rtribotte))
- **[k8s,k8s/gatewayapi]** Fix Kubernetes Gateway API documentation links ([#8063](https://github.com/traefik/traefik/pull/8063) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/gatewayapi]** Fix: k8s gateway api link ([#8085](https://github.com/traefik/traefik/pull/8085) by [tomMoulard](https://github.com/tomMoulard))
- **[k8s,k8s/gatewayapi]** Fix the &#34;values&#34; field in the example of httproute ([#8192](https://github.com/traefik/traefik/pull/8192) by [maelvls](https://github.com/maelvls))
- **[k8s/crd]** Fix ServersTransport documentation ([#8019](https://github.com/traefik/traefik/pull/8019) by [tomMoulard](https://github.com/tomMoulard))
- **[k8s]** Correct annotation option ([#8031](https://github.com/traefik/traefik/pull/8031) by [cbergmann](https://github.com/cbergmann))
- **[metrics]** Add metrics documentation ([#8007](https://github.com/traefik/traefik/pull/8007) by [rtribotte](https://github.com/rtribotte))
- **[middleware]** Docs: add examples for removing headers ([#8030](https://github.com/traefik/traefik/pull/8030) by [SuperSandro2000](https://github.com/SuperSandro2000))
- **[middleware]** Doc: clarify usage for ratelimit&#39;s excludedIPs ([#8072](https://github.com/traefik/traefik/pull/8072) by [mpl](https://github.com/mpl))
- **[middleware]** Elaborate on possible use of status codes with the errors middleware ([#8176](https://github.com/traefik/traefik/pull/8176) by [Midnighter](https://github.com/Midnighter))
- **[middleware]** Doc: fix a syntax error in ratelimit TOML configuration sample ([#8101](https://github.com/traefik/traefik/pull/8101) by [mvertes](https://github.com/mvertes))
- **[pilot]** Docs: add pilot dashboard flag to static configuration file reference ([#8152](https://github.com/traefik/traefik/pull/8152) by [danshilm](https://github.com/danshilm))
- Adding Maintainers Guidelines ([#8168](https://github.com/traefik/traefik/pull/8168) by [jakubhajek](https://github.com/jakubhajek))
- Explains Traefik HTTP response status codes ([#8170](https://github.com/traefik/traefik/pull/8170) by [rtribotte](https://github.com/rtribotte))
- Doc: typo fix ([#8026](https://github.com/traefik/traefik/pull/8026) by [mpl](https://github.com/mpl))
- Adding formatting to the document. ([#8180](https://github.com/traefik/traefik/pull/8180) by [jakubhajek](https://github.com/jakubhajek))
- Changing default file format for the snippets from TOML to YAML ([#8193](https://github.com/traefik/traefik/pull/8193) by [tomMoulard](https://github.com/tomMoulard))

## [v2.4.8](https://github.com/traefik/traefik/tree/v2.4.8) (2021-03-22)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.7...v2.4.8)

**Bug fixes:**
- **[acme]** Update go-acme/lego to v4.3.1 ([#7980](https://github.com/traefik/traefik/pull/7980) by [ldez](https://github.com/ldez))
- **[acme]** Update go-acme/lego to v4.3.0 ([#7975](https://github.com/traefik/traefik/pull/7975) by [ldez](https://github.com/ldez))
- **[k8s,k8s/gatewayapi]** Update to gateway-api v0.2.0 ([#7943](https://github.com/traefik/traefik/pull/7943) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[pilot,webui]** Adding an option to (de)activate Pilot integration into the Traefik dashboard ([#7994](https://github.com/traefik/traefik/pull/7994) by [tomMoulard](https://github.com/tomMoulard))
- **[rules]** Raise errors for non-ASCII domain names in a router&#39;s rules ([#7986](https://github.com/traefik/traefik/pull/7986) by [rtribotte](https://github.com/rtribotte))
- **[server]** Update pires/go-proxyproto to v0.5.0 ([#7948](https://github.com/traefik/traefik/pull/7948) by [mschneider82](https://github.com/mschneider82))

**Documentation:**
- **[middleware]** Improve basic auth middleware httpasswd example ([#7992](https://github.com/traefik/traefik/pull/7992) by [d3473r](https://github.com/d3473r))
- **[middleware]** Add missing `traefik.` prefix across sample config ([#7990](https://github.com/traefik/traefik/pull/7990) by [deepyaman](https://github.com/deepyaman))
- **[middleware]** Remove a no longer needed note ([#7979](https://github.com/traefik/traefik/pull/7979) by [cmcga1125](https://github.com/cmcga1125))

## [v2.4.7](https://github.com/traefik/traefik/tree/v2.4.7) (2021-03-08)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.6...v2.4.7)

**Bug fixes:**
- **[acme]** Fix: double close chan on TLS challenge ([#7956](https://github.com/traefik/traefik/pull/7956) by [ldez](https://github.com/ldez))
- **[provider]** Bump paerser to v0.1.2 ([#7945](https://github.com/traefik/traefik/pull/7945) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[server]** Feature: tune transport buffer size to increase performance ([#7957](https://github.com/traefik/traefik/pull/7957) by [mvertes](https://github.com/mvertes))

**Documentation:**
- **[service]** Fix ServersTransport documentation ([#7942](https://github.com/traefik/traefik/pull/7942) by [rtribotte](https://github.com/rtribotte))

## [v2.4.6](https://github.com/traefik/traefik/tree/v2.4.6) (2021-03-01)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.5...v2.4.6)

**Bug fixes:**
- **[plugins]** Update Yaegi to v0.9.13 ([#7928](https://github.com/traefik/traefik/pull/7928) by [ldez](https://github.com/ldez))
- **[provider]** Fix: wait for file and internal before applying configurations ([#7925](https://github.com/traefik/traefik/pull/7925) by [juliens](https://github.com/juliens))

**Documentation:**
- **[file]** Fix reflink typo in file provider documentation ([#7913](https://github.com/traefik/traefik/pull/7913) by [vgerak](https://github.com/vgerak))
- **[k8s/serviceapi]** Fix Kubernetes Gateway API documentation links ([#7914](https://github.com/traefik/traefik/pull/7914) by [kevinpollet](https://github.com/kevinpollet))
- **[service]** Fix typo in routing/services/index.md ([#7922](https://github.com/traefik/traefik/pull/7922) by [snikch](https://github.com/snikch))
- Fixing doc for default value of checknewversion ([#7933](https://github.com/traefik/traefik/pull/7933) by [tomMoulard](https://github.com/tomMoulard))

## [v2.4.5](https://github.com/traefik/traefik/tree/v2.4.5) (2021-02-18)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.3...v2.4.5)

**Bug fixes:**
- **[webui]** Only allow iframes to be loaded from our domain ([#7904](https://github.com/traefik/traefik/pull/7904) by [SantoDE](https://github.com/SantoDE))

## [v2.4.4](https://github.com/traefik/traefik/tree/v2.4.4) (2021-02-18)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.3...v2.4.4)

Release canceled.

## [v2.4.3](https://github.com/traefik/traefik/tree/v2.4.3) (2021-02-15)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.2...v2.4.3)

**Bug fixes:**
- **[acme]** Fix TLS challenge timeout and validation error ([#7879](https://github.com/traefik/traefik/pull/7879) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Fixed typo in consul catalog tests ([#7865](https://github.com/traefik/traefik/pull/7865) by [apollo13](https://github.com/apollo13))
- **[middleware]** Apply content type exclusion on response ([#7888](https://github.com/traefik/traefik/pull/7888) by [jbdoumenjou](https://github.com/jbdoumenjou))

**Documentation:**
- **[middleware]** Add HEAD as available option for Method ([#7858](https://github.com/traefik/traefik/pull/7858) by [mlandauer](https://github.com/mlandauer))
- **[middleware]** Middleware documentation fixes ([#7808](https://github.com/traefik/traefik/pull/7808) by [Ullaakut](https://github.com/Ullaakut))
- **[provider]** Add missing doc about servers transport ([#7894](https://github.com/traefik/traefik/pull/7894) by [ldez](https://github.com/ldez))
- **[provider]** Provider documentation fixes ([#7823](https://github.com/traefik/traefik/pull/7823) by [Ullaakut](https://github.com/Ullaakut))
- Fix the static reference documentation for the internal redirection router ([#7860](https://github.com/traefik/traefik/pull/7860) by [jbdoumenjou](https://github.com/jbdoumenjou))

## [v2.4.2](https://github.com/traefik/traefik/tree/v2.4.2) (2021-02-02)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.1...v2.4.2)

**Bug fixes:**
- **[acme]** Fix the redirect entrypoint default priority ([#7851](https://github.com/traefik/traefik/pull/7851) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware]** Fix the infinite loop in forwarded header middleware. ([#7847](https://github.com/traefik/traefik/pull/7847) by [ldez](https://github.com/ldez))

**Documentation:**
- Fix the static configuration generation for environment variables ([#7849](https://github.com/traefik/traefik/pull/7849) by [jbdoumenjou](https://github.com/jbdoumenjou))

## [v2.4.1](https://github.com/traefik/traefik/tree/v2.4.1) (2021-02-01)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.0...v2.4.1)

**Bug fixes:**
- **[acme,provider]** Fix HTTP challenge router unexpected delayed creation ([#7805](https://github.com/traefik/traefik/pull/7805) by [jspdown](https://github.com/jspdown))
- **[acme]** Update go-acme/lego to v4.2.0 ([#7793](https://github.com/traefik/traefik/pull/7793) by [ldez](https://github.com/ldez))
- **[api,plugins]** Fix plugin type on middleware endpoint response ([#7782](https://github.com/traefik/traefik/pull/7782) by [jspdown](https://github.com/jspdown))
- **[authentication,middleware]** Forward Proxy-Authorization header to authentication server ([#7433](https://github.com/traefik/traefik/pull/7433) by [Scapal](https://github.com/Scapal))
- **[k8s,k8s/ingress]** Add support for multiple ingress classes ([#7799](https://github.com/traefik/traefik/pull/7799) by [LandryBe](https://github.com/LandryBe))
- **[middleware]** Improve forwarded header and recovery middlewares performances ([#7783](https://github.com/traefik/traefik/pull/7783) by [juliens](https://github.com/juliens))
- **[pilot]** Reduce pressure of pilot services when errors occurs ([#7824](https://github.com/traefik/traefik/pull/7824) by [darkweaver87](https://github.com/darkweaver87))
- **[provider]** Fix aggregator test comment ([#7840](https://github.com/traefik/traefik/pull/7840) by [rtribotte](https://github.com/rtribotte))
- **[provider]** Fix servers transport not found ([#7839](https://github.com/traefik/traefik/pull/7839) by [jspdown](https://github.com/jspdown))

**Documentation:**
- **[consulcatalog]** Fix refresh interval option description in consulcatalog provider ([#7810](https://github.com/traefik/traefik/pull/7810) by [GabeL7r](https://github.com/GabeL7r))
- **[docker]** Fix missing serverstransport documentation ([#7822](https://github.com/traefik/traefik/pull/7822) by [kevinpollet](https://github.com/kevinpollet))
- **[k8s]** Fix YAML syntax in providers docs ([#7787](https://github.com/traefik/traefik/pull/7787) by [4ops](https://github.com/4ops))
- **[service]** Fix typo in server transports documentation ([#7797](https://github.com/traefik/traefik/pull/7797) by [obezuk](https://github.com/obezuk))

## [v2.4.0](https://github.com/traefik/traefik/tree/v2.4.0) (2021-01-19)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.0-rc1...v2.4.0)

**Enhancements:**
- **[acme]** New HTTP and TLS challenges implementations ([#7458](https://github.com/traefik/traefik/pull/7458) by [ldez](https://github.com/ldez))
- **[acme]** Add external account binding support ([#7599](https://github.com/traefik/traefik/pull/7599) by [ldez](https://github.com/ldez))
- **[authentication,middleware]** Middlewares: add forwardAuth.authResponseHeadersRegex ([#7449](https://github.com/traefik/traefik/pull/7449) by [iamolegga](https://github.com/iamolegga))
- **[authentication,middleware]** Filter ForwardAuth request headers ([#7226](https://github.com/traefik/traefik/pull/7226) by [nkonev](https://github.com/nkonev))
- **[k8s,k8s/ingress]** Update more than one LoadBalancer IP ([#6951](https://github.com/traefik/traefik/pull/6951) by [iameli](https://github.com/iameli))
- **[k8s,k8s/ingress]** Set kubernetes client User-Agent to something meaningful ([#7392](https://github.com/traefik/traefik/pull/7392) by [sylr](https://github.com/sylr))
- **[k8s]** Add Kubernetes Gateway Provider ([#7416](https://github.com/traefik/traefik/pull/7416) by [rtribotte](https://github.com/rtribotte))
- **[k8s]** Bump k8s client to v0.19.2 ([#7402](https://github.com/traefik/traefik/pull/7402) by [rtribotte](https://github.com/rtribotte))
- **[kv]** Allows multi-level KV prefixes ([#6664](https://github.com/traefik/traefik/pull/6664) by [niki-timofe](https://github.com/niki-timofe))
- **[logs,middleware,docker]** Support configuring a HTTP client timeout in the Docker provider ([#7094](https://github.com/traefik/traefik/pull/7094) by [sirlatrom](https://github.com/sirlatrom))
- **[marathon]** Extend marathon port discovery to allow port names as identifier ([#7359](https://github.com/traefik/traefik/pull/7359) by [basert](https://github.com/basert))
- **[metrics]** Re-add server up metrics ([#6461](https://github.com/traefik/traefik/pull/6461) by [coder-hugo](https://github.com/coder-hugo))
- **[middleware]** Feature: Exponential Backoff in Retry Middleware ([#7460](https://github.com/traefik/traefik/pull/7460) by [danieladams456](https://github.com/danieladams456))
- **[middleware]** Allow to use regular expressions for `AccessControlAllowOriginList` ([#6881](https://github.com/traefik/traefik/pull/6881) by [jodosha](https://github.com/jodosha))
- **[pilot]** Enable stats collection when pilot is enabled ([#7483](https://github.com/traefik/traefik/pull/7483) by [mmatur](https://github.com/mmatur))
- **[pilot]** Send anonymized dynamic configuration to Pilot ([#7615](https://github.com/traefik/traefik/pull/7615) by [jspdown](https://github.com/jspdown))
- **[server]** Added support for tcp proxyProtocol v1&amp;v2 to backend ([#7320](https://github.com/traefik/traefik/pull/7320) by [mschneider82](https://github.com/mschneider82))
- **[service,tls]** Add ServersTransport on services ([#7203](https://github.com/traefik/traefik/pull/7203) by [juliens](https://github.com/juliens))
- **[webui]** Display Proxy Protocol version for backend services in web dashboard ([#7602](https://github.com/traefik/traefik/pull/7602) by [95ulisse](https://github.com/95ulisse))
- Improve setup readability ([#7604](https://github.com/traefik/traefik/pull/7604) by [juliens](https://github.com/juliens))

**Bug fixes:**
- **[docker]** Fix default value of docker client timeout ([#7345](https://github.com/traefik/traefik/pull/7345) by [kevinpollet](https://github.com/kevinpollet))
- **[middleware,k8s/crd]** Add AccessControlAllowOriginListRegex field to deepcopy ([#7512](https://github.com/traefik/traefik/pull/7512) by [kevinpollet](https://github.com/kevinpollet))

**Documentation:**
- **[middleware]** Rephrase forwardauth.authRequestHeaders documentation ([#7701](https://github.com/traefik/traefik/pull/7701) by [Beanow](https://github.com/Beanow))
- Update copyright year for 2021 ([#7754](https://github.com/traefik/traefik/pull/7754) by [kevinpollet](https://github.com/kevinpollet))
- Prepare release v2.4.0-rc2 ([#7747](https://github.com/traefik/traefik/pull/7747) by [kevinpollet](https://github.com/kevinpollet))
- **[kv]** KV doc reference ([#7415](https://github.com/traefik/traefik/pull/7415) by [rtribotte](https://github.com/rtribotte))
- Add jspdown to maintainers ([#7671](https://github.com/traefik/traefik/pull/7671) by [emilevauge](https://github.com/emilevauge))
- Add kevinpollet to maintainers ([#7464](https://github.com/traefik/traefik/pull/7464) by [SantoDE](https://github.com/SantoDE))
- Add security policies ([#7110](https://github.com/traefik/traefik/pull/7110) by [ldez](https://github.com/ldez))

**Misc:**
- Merge current v2.3 branch into v2.4 ([#7765](https://github.com/traefik/traefik/pull/7765) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into v2.4 ([#7760](https://github.com/traefik/traefik/pull/7760) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.3 branch into v2.4 ([#7744](https://github.com/traefik/traefik/pull/7744) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.3 branch into v2.4 ([#7742](https://github.com/traefik/traefik/pull/7742) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.3 branch into v2.4 ([#7727](https://github.com/traefik/traefik/pull/7727) by [mmatur](https://github.com/mmatur))
- Merge current v2.3 branch into v2.4 ([#7703](https://github.com/traefik/traefik/pull/7703) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.3 branch into v2.4 ([#7689](https://github.com/traefik/traefik/pull/7689) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.3 branch into master ([#7677](https://github.com/traefik/traefik/pull/7677) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.3 branch into master ([#7670](https://github.com/traefik/traefik/pull/7670) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.3 branch into master ([#7653](https://github.com/traefik/traefik/pull/7653) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.3 branch into master ([#7574](https://github.com/traefik/traefik/pull/7574) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.3 branch into master ([#7529](https://github.com/traefik/traefik/pull/7529) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7472](https://github.com/traefik/traefik/pull/7472) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.3 branch into master ([#7453](https://github.com/traefik/traefik/pull/7453) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.3 branch into master ([#7405](https://github.com/traefik/traefik/pull/7405) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7401](https://github.com/traefik/traefik/pull/7401) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.3 branch into master ([#7346](https://github.com/traefik/traefik/pull/7346) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7335](https://github.com/traefik/traefik/pull/7335) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7299](https://github.com/traefik/traefik/pull/7299) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7263](https://github.com/traefik/traefik/pull/7263) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7215](https://github.com/traefik/traefik/pull/7215) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7122](https://github.com/traefik/traefik/pull/7122) by [ldez](https://github.com/ldez))

## [v2.4.0-rc2](https://github.com/traefik/traefik/tree/v2.4.0-rc2) (2021-01-12)
[All Commits](https://github.com/traefik/traefik/compare/v2.4.0-rc1...v2.4.0-rc2)

**Documentation:**
- **[middleware]** Rephrase forwardauth.authRequestHeaders documentation ([#7701](https://github.com/traefik/traefik/pull/7701) by [Beanow](https://github.com/Beanow))

**Misc:**
- Merge current v2.3 branch into v2.4 ([#7744](https://github.com/traefik/traefik/pull/7744) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.3 branch into v2.4 ([#7742](https://github.com/traefik/traefik/pull/7742) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.3 branch into v2.4 ([#7727](https://github.com/traefik/traefik/pull/7727) by [mmatur](https://github.com/mmatur))
- Merge current v2.3 branch into v2.4 ([#7703](https://github.com/traefik/traefik/pull/7703) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.3 branch into v2.4 ([#7689](https://github.com/traefik/traefik/pull/7689) by [kevinpollet](https://github.com/kevinpollet))

## [v2.3.7](https://github.com/traefik/traefik/tree/v2.3.7) (2021-01-11)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.6...v2.3.7)

**Bug fixes:**
- **[k8s,k8s/ingress]** Fix wildcard hostname issue ([#7711](https://github.com/traefik/traefik/pull/7711) by [avdhoot](https://github.com/avdhoot))
- **[k8s,k8s/ingress]** Compile kubernetes ingress annotation regex only once ([#7647](https://github.com/traefik/traefik/pull/7647) by [hensur](https://github.com/hensur))
- **[middleware,webui]** webui: fix missing custom request and response header names ([#7706](https://github.com/traefik/traefik/pull/7706) by [kevinpollet](https://github.com/kevinpollet))
- **[middleware]** Fix log level on error pages middleware ([#7737](https://github.com/traefik/traefik/pull/7737) by [Nowheresly](https://github.com/Nowheresly))

**Documentation:**
- **[docker]** docs: fix broken links to docker-compose documentation ([#7702](https://github.com/traefik/traefik/pull/7702) by [kevinpollet](https://github.com/kevinpollet))
- **[ecs]** Add ECS to supported providers list ([#7714](https://github.com/traefik/traefik/pull/7714) by [anilmaurya](https://github.com/anilmaurya))
- Update copyright year for 2021 ([#7734](https://github.com/traefik/traefik/pull/7734) by [kevinpollet](https://github.com/kevinpollet))

## [v2.3.6](https://github.com/traefik/traefik/tree/v2.3.6) (2020-12-17)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.5...v2.3.6)

**Bug fixes:**
- **[logs]** Update Logrus to v1.7.0 ([#7663](https://github.com/traefik/traefik/pull/7663) by [jspdown](https://github.com/jspdown))
- **[plugins]** Update Yaegi to v0.9.8 ([#7659](https://github.com/traefik/traefik/pull/7659) by [ldez](https://github.com/ldez))
- **[rules]** Disable router when a rule has an error ([#7680](https://github.com/traefik/traefik/pull/7680) by [ldez](https://github.com/ldez))

**Documentation:**
- **[logs]** Add configuration example for access log filePath ([#7655](https://github.com/traefik/traefik/pull/7655) by [wernerfred](https://github.com/wernerfred))
- **[middleware]** Add missing quotes in errorpages k8s example yaml ([#7675](https://github.com/traefik/traefik/pull/7675) by [icelynjennings](https://github.com/icelynjennings))

## [v2.4.0-rc1](https://github.com/traefik/traefik/tree/v2.4.0-rc1) (2020-12-16)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.0-rc1...v2.4.0-rc1)

**Enhancements:**
- **[acme]** New HTTP and TLS challenges implementations ([#7458](https://github.com/traefik/traefik/pull/7458) by [ldez](https://github.com/ldez))
- **[acme]** Add external account binding support ([#7599](https://github.com/traefik/traefik/pull/7599) by [ldez](https://github.com/ldez))
- **[authentication,middleware]** Middlewares: add forwardAuth.authResponseHeadersRegex ([#7449](https://github.com/traefik/traefik/pull/7449) by [iamolegga](https://github.com/iamolegga))
- **[authentication,middleware]** Filter ForwardAuth request headers ([#7226](https://github.com/traefik/traefik/pull/7226) by [nkonev](https://github.com/nkonev))
- **[k8s,k8s/ingress]** Update more than one LoadBalancer IP ([#6951](https://github.com/traefik/traefik/pull/6951) by [iameli](https://github.com/iameli))
- **[k8s,k8s/ingress]** Set kubernetes client User-Agent to something meaningful ([#7392](https://github.com/traefik/traefik/pull/7392) by [sylr](https://github.com/sylr))
- **[k8s]** Add Kubernetes Gateway Provider ([#7416](https://github.com/traefik/traefik/pull/7416) by [rtribotte](https://github.com/rtribotte))
- **[k8s]** Bump k8s client to v0.19.2 ([#7402](https://github.com/traefik/traefik/pull/7402) by [rtribotte](https://github.com/rtribotte))
- **[kv]** Allows multi-level KV prefixes ([#6664](https://github.com/traefik/traefik/pull/6664) by [niki-timofe](https://github.com/niki-timofe))
- **[logs,middleware,docker]** Support configuring a HTTP client timeout in the Docker provider ([#7094](https://github.com/traefik/traefik/pull/7094) by [sirlatrom](https://github.com/sirlatrom))
- **[marathon]** Extend marathon port discovery to allow port names as identifier ([#7359](https://github.com/traefik/traefik/pull/7359) by [basert](https://github.com/basert))
- **[metrics]** Re-add server up metrics ([#6461](https://github.com/traefik/traefik/pull/6461) by [coder-hugo](https://github.com/coder-hugo))
- **[middleware]** Feature: Exponential Backoff in Retry Middleware ([#7460](https://github.com/traefik/traefik/pull/7460) by [danieladams456](https://github.com/danieladams456))
- **[middleware]** Allow to use regular expressions for `AccessControlAllowOriginList` ([#6881](https://github.com/traefik/traefik/pull/6881) by [jodosha](https://github.com/jodosha))
- **[pilot]** Enable stats collection when pilot is enabled ([#7483](https://github.com/traefik/traefik/pull/7483) by [mmatur](https://github.com/mmatur))
- **[pilot]** Send anonymized dynamic configuration to Pilot ([#7615](https://github.com/traefik/traefik/pull/7615) by [jspdown](https://github.com/jspdown))
- **[server]** Added support for tcp proxyProtocol v1&amp;v2 to backend ([#7320](https://github.com/traefik/traefik/pull/7320) by [mschneider82](https://github.com/mschneider82))
- **[service,tls]** Add ServersTransport on services ([#7203](https://github.com/traefik/traefik/pull/7203) by [juliens](https://github.com/juliens))
- **[webui]** Display Proxy Protocol version for backend services in web dashboard ([#7602](https://github.com/traefik/traefik/pull/7602) by [95ulisse](https://github.com/95ulisse))
- Improve setup readability ([#7604](https://github.com/traefik/traefik/pull/7604) by [juliens](https://github.com/juliens))

**Bug fixes:**
- **[docker]** Fix default value of docker client timeout ([#7345](https://github.com/traefik/traefik/pull/7345) by [kevinpollet](https://github.com/kevinpollet))
- **[middleware,k8s/crd]** Add AccessControlAllowOriginListRegex field to deepcopy ([#7512](https://github.com/traefik/traefik/pull/7512) by [kevinpollet](https://github.com/kevinpollet))

**Documentation:**
- **[kv]** KV doc reference ([#7415](https://github.com/traefik/traefik/pull/7415) by [rtribotte](https://github.com/rtribotte))
- Add jspdown to maintainers ([#7671](https://github.com/traefik/traefik/pull/7671) by [emilevauge](https://github.com/emilevauge))
- Add kevinpollet to maintainers ([#7464](https://github.com/traefik/traefik/pull/7464) by [SantoDE](https://github.com/SantoDE))
- Add security policies ([#7110](https://github.com/traefik/traefik/pull/7110) by [ldez](https://github.com/ldez))

**Misc:**
- Merge current v2.3 branch into master ([#7677](https://github.com/traefik/traefik/pull/7677) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.3 branch into master ([#7670](https://github.com/traefik/traefik/pull/7670) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.3 branch into master ([#7653](https://github.com/traefik/traefik/pull/7653) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.3 branch into master ([#7574](https://github.com/traefik/traefik/pull/7574) by [kevinpollet](https://github.com/kevinpollet))
- Merge current v2.3 branch into master ([#7529](https://github.com/traefik/traefik/pull/7529) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7472](https://github.com/traefik/traefik/pull/7472) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.3 branch into master ([#7453](https://github.com/traefik/traefik/pull/7453) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.3 branch into master ([#7405](https://github.com/traefik/traefik/pull/7405) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7401](https://github.com/traefik/traefik/pull/7401) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.3 branch into master ([#7346](https://github.com/traefik/traefik/pull/7346) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7335](https://github.com/traefik/traefik/pull/7335) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7299](https://github.com/traefik/traefik/pull/7299) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7263](https://github.com/traefik/traefik/pull/7263) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7215](https://github.com/traefik/traefik/pull/7215) by [ldez](https://github.com/ldez))
- Merge current v2.3 branch into master ([#7122](https://github.com/traefik/traefik/pull/7122) by [ldez](https://github.com/ldez))

## [v2.3.5](https://github.com/traefik/traefik/tree/v2.3.5) (2020-12-10)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.4...v2.3.5)

**Bug fixes:**
- **[acme]** Update go-acme/lego to v4.1.3 ([#7625](https://github.com/traefik/traefik/pull/7625) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd]** IngressRoute: add an option to disable cross-namespace routing ([#7595](https://github.com/traefik/traefik/pull/7595) by [rtribotte](https://github.com/rtribotte))
- **[k8s/crd,k8s/ingress]** Fix concatenation of IPv6 addresses and ports ([#7620](https://github.com/traefik/traefik/pull/7620) by [jspdown](https://github.com/jspdown))
- **[tcp,tls]** Fix TLS options fallback when domain and options are the same  ([#7609](https://github.com/traefik/traefik/pull/7609) by [jspdown](https://github.com/jspdown))
- **[webui]** Fix UI bug on long service name ([#7535](https://github.com/traefik/traefik/pull/7535) by [ipinak](https://github.com/ipinak))

**Documentation:**
- **[docker]** Add example for multiple service per container ([#7610](https://github.com/traefik/traefik/pull/7610) by [notsureifkevin](https://github.com/notsureifkevin))
- Documentation: Add spacing to sidebars so the last item is always visible ([#7616](https://github.com/traefik/traefik/pull/7616) by [paulocfjunior](https://github.com/paulocfjunior))
- Fix typos in migration guide  ([#7596](https://github.com/traefik/traefik/pull/7596) by [marsavela](https://github.com/marsavela))

## [v2.3.4](https://github.com/traefik/traefik/tree/v2.3.4) (2020-11-24)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.3...v2.3.4)

**Bug fixes:**
- **[acme]** Update go-acme/lego to v4.1.2 ([#7577](https://github.com/traefik/traefik/pull/7577) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd,k8s/ingress]** Apply labelSelector as a TweakListOptions for Kubernetes informers ([#7521](https://github.com/traefik/traefik/pull/7521) by [rtribotte](https://github.com/rtribotte))
- **[middleware]** Do not evaluate templated URL in redirectRegex middleware ([#7573](https://github.com/traefik/traefik/pull/7573) by [jspdown](https://github.com/jspdown))
- **[provider]** fix: invalid slice parsing. ([#7583](https://github.com/traefik/traefik/pull/7583) by [ldez](https://github.com/ldez))

**Documentation:**
- **[ecs]** Fix clusters option in ECS provider documentation ([#7586](https://github.com/traefik/traefik/pull/7586) by [skapin](https://github.com/skapin))

## [v2.3.3](https://github.com/traefik/traefik/tree/v2.3.3) (2020-11-19)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.2...v2.3.3)

**Bug fixes:**
- **[acme]** Update go-acme/lego to v4.1.0 ([#7526](https://github.com/traefik/traefik/pull/7526) by [ldez](https://github.com/ldez))
- **[consulcatalog,ecs]** Fix missing allow-empty tag on ECS and Consul Catalog providers ([#7561](https://github.com/traefik/traefik/pull/7561) by [jspdown](https://github.com/jspdown))
- **[consulcatalog]** consulcatalog to update before the first interval ([#7514](https://github.com/traefik/traefik/pull/7514) by [greut](https://github.com/greut))
- **[consulcatalog]** Fix consul catalog panic when health and services are not in sync ([#7558](https://github.com/traefik/traefik/pull/7558) by [jspdown](https://github.com/jspdown))
- **[ecs]** Translate configured server port into correct mapped host port ([#7480](https://github.com/traefik/traefik/pull/7480) by [alekitto](https://github.com/alekitto))
- **[k8s,k8s/crd,k8s/ingress]** Filter out Helm secrets from informer caches ([#7562](https://github.com/traefik/traefik/pull/7562) by [jspdown](https://github.com/jspdown))
- **[plugins]** Update Yaegi to v0.9.5 ([#7527](https://github.com/traefik/traefik/pull/7527) by [ldez](https://github.com/ldez))
- **[plugins]** Update Yaegi to v0.9.7 ([#7569](https://github.com/traefik/traefik/pull/7569) by [kevinpollet](https://github.com/kevinpollet))
- **[plugins]** Update Yaegi to v0.9.4 ([#7451](https://github.com/traefik/traefik/pull/7451) by [ldez](https://github.com/ldez))
- **[tcp]** Ignore errors when setting keepalive period is not supported by the system ([#7410](https://github.com/traefik/traefik/pull/7410) by [tristan-weil](https://github.com/tristan-weil))
- **[tcp]** Improve service name lookup on TCP routers ([#7370](https://github.com/traefik/traefik/pull/7370) by [ddtmachado](https://github.com/ddtmachado))
- Improve anonymize configuration ([#7482](https://github.com/traefik/traefik/pull/7482) by [mmatur](https://github.com/mmatur))

**Documentation:**
- **[ecs]** Add ECS menu to dynamic config reference ([#7501](https://github.com/traefik/traefik/pull/7501) by [kevinpollet](https://github.com/kevinpollet))
- **[k8s,k8s/ingress]** Fix ingress documentation ([#7424](https://github.com/traefik/traefik/pull/7424) by [rtribotte](https://github.com/rtribotte))
- **[k8s]** fix documentation ([#7469](https://github.com/traefik/traefik/pull/7469) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s]** Fix grammar in kubernetes ingress controller documentation ([#7565](https://github.com/traefik/traefik/pull/7565) by [ivorscott](https://github.com/ivorscott))
- **[logs]** Clarify time-based field units ([#7447](https://github.com/traefik/traefik/pull/7447) by [tomtastic](https://github.com/tomtastic))
- **[middleware]** Forwardauth headers ([#7506](https://github.com/traefik/traefik/pull/7506) by [w4tsn](https://github.com/w4tsn))
- **[provider]** fix typo in providers overview documentation ([#7441](https://github.com/traefik/traefik/pull/7441) by [pirey](https://github.com/pirey))
- **[tls]** Fix docs for TLS ([#7541](https://github.com/traefik/traefik/pull/7541) by [james426759](https://github.com/james426759))
- fix: exclude protected link from doc verify ([#7477](https://github.com/traefik/traefik/pull/7477) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Add missed tls config for yaml example ([#7450](https://github.com/traefik/traefik/pull/7450) by [andrew-demb](https://github.com/andrew-demb))
- Resolve broken URLs causing make docs to fail ([#7444](https://github.com/traefik/traefik/pull/7444) by [tomtastic](https://github.com/tomtastic))
- Fix Traefik Proxy product nav in docs ([#7523](https://github.com/traefik/traefik/pull/7523) by [PCM2](https://github.com/PCM2))
- add links to contributors guide ([#7435](https://github.com/traefik/traefik/pull/7435) by [notsureifkevin](https://github.com/notsureifkevin))

## [v2.3.2](https://github.com/traefik/traefik/tree/v2.3.2) (2020-10-19)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.1...v2.3.2)

**Bug fixes:**
- **[acme]** fix: restrict protocol for TLS Challenge. ([#7400](https://github.com/traefik/traefik/pull/7400) by [ldez](https://github.com/ldez))
- **[acme]** fix: use provider keytype instead of account keytype. ([#7387](https://github.com/traefik/traefik/pull/7387) by [mmatur](https://github.com/mmatur))
- **[acme]** acme: Fix race condition in LocalStore during saving. ([#7355](https://github.com/traefik/traefik/pull/7355) by [walkline](https://github.com/walkline))
- **[plugins]** fix: update Yaegi to v0.9.4 ([#7426](https://github.com/traefik/traefik/pull/7426) by [ldez](https://github.com/ldez))
- **[udp]** fix: udp json struct tag ([#7375](https://github.com/traefik/traefik/pull/7375) by [mschneider82](https://github.com/mschneider82))

**Documentation:**
- **[consulcatalog]** fix: Consul Catalog address documentation. ([#7429](https://github.com/traefik/traefik/pull/7429) by [ldez](https://github.com/ldez))
- **[middleware]** Moving Provider Namespace documentation topic to Configuration Discovery section ([#7423](https://github.com/traefik/traefik/pull/7423) by [AndrewSav](https://github.com/AndrewSav))
- **[pilot]** fix: pilot static configuration documentation ([#7399](https://github.com/traefik/traefik/pull/7399) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[plugins]** Revise Traefik Pilot documentation section ([#7427](https://github.com/traefik/traefik/pull/7427) by [PCM2](https://github.com/PCM2))
- **[tls]** Adding details about the default TLS options to the documentation ([#7422](https://github.com/traefik/traefik/pull/7422) by [AndrewSav](https://github.com/AndrewSav))
- doc: add YAML sample. ([#7397](https://github.com/traefik/traefik/pull/7397) by [ldez](https://github.com/ldez))
- Fix containous links in readme ([#7394](https://github.com/traefik/traefik/pull/7394) by [kevinpollet](https://github.com/kevinpollet))
- Fix broken logo ([#7390](https://github.com/traefik/traefik/pull/7390) by [Bencey](https://github.com/Bencey))

## [v2.3.1](https://github.com/traefik/traefik/tree/v2.3.1) (2020-09-28)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.0...v2.3.1)

**Bug fixes:**
- **[webui]** Fix blank webui on some browsers ([#7364](https://github.com/traefik/traefik/pull/7364) by [matthieuh](https://github.com/matthieuh))

**Documentation:**
- **[k8s/helm]** Update of the helm repo localisation ([#7352](https://github.com/traefik/traefik/pull/7352) by [dgoujard](https://github.com/dgoujard))
- restore traefik logo ([#7344](https://github.com/traefik/traefik/pull/7344) by [notsureifkevin](https://github.com/notsureifkevin))
- Removes invalid items in the changelog. ([#7339](https://github.com/traefik/traefik/pull/7339) by [ldez](https://github.com/ldez))

## [v2.3.0](https://github.com/traefik/traefik/tree/v2.3.0) (2020-09-23)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.0-rc1...v2.3.0)

**Enhancements:**
- **[api]** Add custom ping http code when Traefik is terminating ([#6696](https://github.com/traefik/traefik/pull/6696) by [L3o-pold](https://github.com/L3o-pold))
- **[ecs]** Add AWS ECS provider ([#6749](https://github.com/traefik/traefik/pull/6749) by [alekitto](https://github.com/alekitto))
- **[file]** feat: use parser to load dynamic config from file. ([#6875](https://github.com/traefik/traefik/pull/6875) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd,k8s/ingress]** Upgrade Client-go to 0.18.2 ([#6779](https://github.com/traefik/traefik/pull/6779) by [dtomcej](https://github.com/dtomcej))
- **[k8s,k8s/ingress]** Add new ingressClass support to ingress provider ([#6831](https://github.com/traefik/traefik/pull/6831) by [dtomcej](https://github.com/dtomcej))
- **[k8s,k8s/ingress]** Add example for the IngressClass usage ([#7219](https://github.com/traefik/traefik/pull/7219) by [SantoDE](https://github.com/SantoDE))
- **[metrics,pilot]** Pilot metrics provider ([#7139](https://github.com/traefik/traefik/pull/7139) by [rtribotte](https://github.com/rtribotte))
- **[pilot]** Moves pilot outside the experimental section. ([#7287](https://github.com/traefik/traefik/pull/7287) by [ldez](https://github.com/ldez))
- **[pilot,plugins]** Traefik Pilot: plugins support and alert system (EXPERIMENTAL FEATURES) ([#7041](https://github.com/traefik/traefik/pull/7041) by [ldez](https://github.com/ldez))
- **[plugins]** Improve plugins builder. ([#7255](https://github.com/traefik/traefik/pull/7255) by [ldez](https://github.com/ldez))
- **[provider]** Add HTTP Provider ([#6976](https://github.com/traefik/traefik/pull/6976) by [kevinpollet](https://github.com/kevinpollet))
- **[webui]** Add iOS specific icons ([#6946](https://github.com/traefik/traefik/pull/6946) by [Heisenberg74](https://github.com/Heisenberg74))

**Bug fixes:**
- **[acme]** fix: precheck function. ([#7333](https://github.com/traefik/traefik/pull/7333) by [ldez](https://github.com/ldez))
- **[ecs]** Improve region resolution for ECS provider ([#7145](https://github.com/traefik/traefik/pull/7145) by [kevinpollet](https://github.com/kevinpollet))
- **[k8s,k8s/ingress]** Delete an unnecessary warning log ([#6568](https://github.com/traefik/traefik/pull/6568) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/ingress]** Support Kubernetes Ingress pathType ([#7087](https://github.com/traefik/traefik/pull/7087) by [rtribotte](https://github.com/rtribotte))
- **[k8s,k8s/ingress]** Use semantic versioning to enable ingress class support ([#7065](https://github.com/traefik/traefik/pull/7065) by [kevinpollet](https://github.com/kevinpollet))
- **[metrics]** fix: uint64 alignment in go-kit. ([#7289](https://github.com/traefik/traefik/pull/7289) by [ldez](https://github.com/ldez))
- **[middleware]** Allow multiple secure middlewares to operate independently  ([#6604](https://github.com/traefik/traefik/pull/6604) by [dtomcej](https://github.com/dtomcej))
- **[pilot,webui]** Avoid Traefik Pilot iframe code in Traefik webui regarding notifications ([#7272](https://github.com/traefik/traefik/pull/7272) by [matthieuh](https://github.com/matthieuh))
- **[pilot,webui]** Add ability to dismiss pilot notification ([#7200](https://github.com/traefik/traefik/pull/7200) by [matthieuh](https://github.com/matthieuh))
- **[pilot]** fix: pilot metrics unit for req duration. ([#7309](https://github.com/traefik/traefik/pull/7309) by [ldez](https://github.com/ldez))
- **[pilot]** fix: start of Traefik Pilot ([#7304](https://github.com/traefik/traefik/pull/7304) by [ldez](https://github.com/ldez))
- **[provider]** file parser: skip nil value. ([#7058](https://github.com/traefik/traefik/pull/7058) by [ldez](https://github.com/ldez))
- **[tracing]** Update jaeger-client-go dependency to v2.25.0 ([#7198](https://github.com/traefik/traefik/pull/7198) by [kevinpollet](https://github.com/kevinpollet))

**Documentation:**
- **[consul]** Fix consul catalog router tag example ([#7332](https://github.com/traefik/traefik/pull/7332) by [rtribotte](https://github.com/rtribotte))
- **[ecs]** Fix documentation for ECS ([#7107](https://github.com/traefik/traefik/pull/7107) by [mmatur](https://github.com/mmatur))
- **[k8s]** docs: add missing apigroup to Kubernetes RBAC ([#7199](https://github.com/traefik/traefik/pull/7199) by [kevinpollet](https://github.com/kevinpollet))
- **[k8s]** Add the ingressclass resource in the ingress RBAC documentation ([#7290](https://github.com/traefik/traefik/pull/7290) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s]** Add migration documentation for IngressClass ([#7083](https://github.com/traefik/traefik/pull/7083) by [kevinpollet](https://github.com/kevinpollet))
- **[middleware]** Fixes config samples regarding forceSlash option ([#6811](https://github.com/traefik/traefik/pull/6811) by [volkerw00](https://github.com/volkerw00))
- **[plugins]** Update availability info ([#7060](https://github.com/traefik/traefik/pull/7060) by [PCM2](https://github.com/PCM2))
- Fix yaml documentation ([#7331](https://github.com/traefik/traefik/pull/7331) by [rtribotte](https://github.com/rtribotte))

**Misc:**
- Merge current v2.2 branch into v2.3 ([#7288](https://github.com/traefik/traefik/pull/7288) by [rtribotte](https://github.com/rtribotte))
- Merge current v2.2 branch into v2.3 ([#7257](https://github.com/traefik/traefik/pull/7257) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into v2.3 ([#7249](https://github.com/traefik/traefik/pull/7249) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into v2.3 ([#7218](https://github.com/traefik/traefik/pull/7218) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into v2.3 ([#7175](https://github.com/traefik/traefik/pull/7175) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into v2.3 ([#7160](https://github.com/traefik/traefik/pull/7160) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into v2.3 ([#7116](https://github.com/traefik/traefik/pull/7116) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into v2.3 ([#7086](https://github.com/traefik/traefik/pull/7086) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.2 branch into master ([#7052](https://github.com/traefik/traefik/pull/7052) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into master ([#7022](https://github.com/traefik/traefik/pull/7022) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.2 branch into master ([#6921](https://github.com/traefik/traefik/pull/6921) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.2 branch into master ([#6822](https://github.com/traefik/traefik/pull/6822) by [mmatur](https://github.com/mmatur))
- Merge current v2.2 branch into master ([#6754](https://github.com/traefik/traefik/pull/6754) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into master ([#6533](https://github.com/traefik/traefik/pull/6533) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into master ([#6468](https://github.com/traefik/traefik/pull/6468) by [ldez](https://github.com/ldez))

## [v2.3.0-rc7](https://github.com/traefik/traefik/tree/v2.3.0-rc7) (2020-09-18)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.0-rc6...v2.3.0-rc7)

**Bug fixes:**
- **[pilot]** fix: pilot metrics unit for req duration. ([#7309](https://github.com/traefik/traefik/pull/7309) by [ldez](https://github.com/ldez))
- **[pilot]** fix: start of Traefik Pilot ([#7304](https://github.com/traefik/traefik/pull/7304) by [ldez](https://github.com/ldez))

## [v2.3.0-rc6](https://github.com/traefik/traefik/tree/v2.3.0-rc6) (2020-09-16)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.0-rc5...v2.3.0-rc6)

**Enhancements:**
- **[pilot]** Moves pilot outside the experimental section. ([#7287](https://github.com/traefik/traefik/pull/7287) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[metrics]** fix: uint64 alignment in go-kit. ([#7289](https://github.com/traefik/traefik/pull/7289) by [ldez](https://github.com/ldez))
- **[pilot,webui]** Avoid Traefik Pilot iframe code in Traefik webui regarding notifications ([#7272](https://github.com/traefik/traefik/pull/7272) by [matthieuh](https://github.com/matthieuh))

**Documentation:**
- **[k8s]** Add the ingressclass resource in the ingress RBAC documentation ([#7290](https://github.com/traefik/traefik/pull/7290) by [jbdoumenjou](https://github.com/jbdoumenjou))

**Misc:**
- **[middleware]** Merge current v2.2 branch into v2.3 ([#7288](https://github.com/traefik/traefik/pull/7288) by [rtribotte](https://github.com/rtribotte))

## [v2.3.0-rc5](https://github.com/traefik/traefik/tree/v2.3.0-rc5) (2020-09-07)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.0-rc4...v2.3.0-rc5)

**Enhancements:**
- **[k8s,k8s/ingress]** Add example for the IngressClass usage ([#7219](https://github.com/traefik/traefik/pull/7219) by [SantoDE](https://github.com/SantoDE))
- **[plugins]** Improve plugins builder. ([#7255](https://github.com/traefik/traefik/pull/7255) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[pilot,webui]** Add ability to dismiss pilot notification ([#7200](https://github.com/traefik/traefik/pull/7200) by [matthieuh](https://github.com/matthieuh))

**Misc:**
- Merge current v2.2 branch into v2.3 ([#7249](https://github.com/traefik/traefik/pull/7249) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into v2.3 ([#7218](https://github.com/traefik/traefik/pull/7218) by [ldez](https://github.com/ldez))

## [v2.2.11](https://github.com/traefik/traefik/tree/v2.2.11) (2020-09-07)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.10...v2.2.11)

**Bug fixes:**
- **[middleware]** fix: header middleware response writer. ([#7252](https://github.com/traefik/traefik/pull/7252) by [ldez](https://github.com/ldez))

**Documentation:**
- **[healthcheck]** Clarified hostname documentation for load balancer healthcheck ([#7254](https://github.com/traefik/traefik/pull/7254) by [AndrewSav](https://github.com/AndrewSav))

## [v2.2.10](https://github.com/traefik/traefik/tree/v2.2.10) (2020-09-04)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.7...v2.2.10)

**Bug fixes:**
- **[acme]** Update go-acme/lego to v4.0.1 ([#7238](https://github.com/traefik/traefik/pull/7238) by [ldez](https://github.com/ldez))
- **[middleware]** Add missing IPStrategy struct tag for YAML ([#7233](https://github.com/traefik/traefik/pull/7233) by [kevinpollet](https://github.com/kevinpollet))
- **[middleware]** Headers response modifier is directly applied by headers middleware ([#7230](https://github.com/traefik/traefik/pull/7230) by [juliens](https://github.com/juliens))
- **[webui]** chore(webui): upgrade nodejs to Node current LTS ([#7125](https://github.com/traefik/traefik/pull/7125) by [Slashgear](https://github.com/Slashgear))

**Documentation:**
- **[docker]** doc: fix dead link. ([#7172](https://github.com/traefik/traefik/pull/7172) by [ldez](https://github.com/ldez))
- **[k8s]** kubernetes-crd: fix whitespace in configuration examples ([#7134](https://github.com/traefik/traefik/pull/7134) by [NT-florianernst](https://github.com/NT-florianernst))
- **[k8s]** doc: replace underscore by hyphen for k8s metadata names. ([#7131](https://github.com/traefik/traefik/pull/7131) by [ldez](https://github.com/ldez))
- **[logs]** doc: added tz section to access log ([#7178](https://github.com/traefik/traefik/pull/7178) by [notsureifkevin](https://github.com/notsureifkevin))
- **[tls]** doc: Minor language improvement in TLS documentation ([#7206](https://github.com/traefik/traefik/pull/7206) by [sharmarajdaksh](https://github.com/sharmarajdaksh))
- doc: fix typo in migration guide ([#7181](https://github.com/traefik/traefik/pull/7181) by [ScuttleSE](https://github.com/ScuttleSE))
- doc: specify HostSNI rule removal only for HTTP routers ([#7237](https://github.com/traefik/traefik/pull/7237) by [rtribotte](https://github.com/rtribotte))
- Reorder migrations for v2 minor upgrades ([#7214](https://github.com/traefik/traefik/pull/7214) by [peschmae](https://github.com/peschmae))
- Harmonize docs ([#7124](https://github.com/traefik/traefik/pull/7124) by [matthieuh](https://github.com/matthieuh))

## [v2.2.9](https://github.com/traefik/traefik/tree/v2.2.9) (2020-09-04)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.8...v2.2.9)

Release canceled due to a bad tag.

## [v2.3.0-rc4](https://github.com/traefik/traefik/tree/v2.3.0-rc4) (2020-08-19)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.0-rc3...v2.3.0-rc4)

**Enhancements:**
- **[metrics,pilot]** Pilot metrics provider ([#7139](https://github.com/traefik/traefik/pull/7139) by [rtribotte](https://github.com/rtribotte))

**Bug fixes:**
- **[ecs]** Improve region resolution for ECS provider ([#7145](https://github.com/traefik/traefik/pull/7145) by [kevinpollet](https://github.com/kevinpollet))
- **[tracing]** Update jaeger-client-go dependency to v2.25.0 ([#7198](https://github.com/traefik/traefik/pull/7198) by [kevinpollet](https://github.com/kevinpollet))

**Documentation:**
- **[k8s]** docs: add missing apigroup to Kubernetes RBAC ([#7199](https://github.com/traefik/traefik/pull/7199) by [kevinpollet](https://github.com/kevinpollet))

**Misc:**
- Merge current v2.2 branch into v2.3 ([#7175](https://github.com/traefik/traefik/pull/7175) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into v2.3 ([#7160](https://github.com/traefik/traefik/pull/7160) by [ldez](https://github.com/ldez))

## [v2.3.0-rc3](https://github.com/traefik/traefik/tree/v2.3.0-rc3) (2020-07-28)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.0-rc2...v2.3.0-rc3)

**Bug fixes:**
- **[k8s,k8s/ingress]** Support Kubernetes Ingress pathType ([#7087](https://github.com/traefik/traefik/pull/7087) by [rtribotte](https://github.com/rtribotte))
- **[k8s,k8s/ingress]** Use semantic versioning to enable ingress class support ([#7065](https://github.com/traefik/traefik/pull/7065) by [kevinpollet](https://github.com/kevinpollet))
- **[provider]** file parser: skip nil value. ([#7058](https://github.com/traefik/traefik/pull/7058) by [ldez](https://github.com/ldez))

**Documentation:**
- **[ecs]** Fix documentation for ECS ([#7107](https://github.com/traefik/traefik/pull/7107) by [mmatur](https://github.com/mmatur))
- **[k8s]** Add migration documentation for IngressClass ([#7083](https://github.com/traefik/traefik/pull/7083) by [kevinpollet](https://github.com/kevinpollet))
- **[plugins]** Update availability info ([#7060](https://github.com/traefik/traefik/pull/7060) by [PCM2](https://github.com/PCM2))

**Misc:**
- Merge current v2.2 branch into v2.3 ([#7116](https://github.com/traefik/traefik/pull/7116) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into v2.3 ([#7086](https://github.com/traefik/traefik/pull/7086) by [jbdoumenjou](https://github.com/jbdoumenjou))

## [v2.2.8](https://github.com/traefik/traefik/tree/v2.2.8) (2020-07-28)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.7...v2.2.8)

**Bug fixes:**
- **[webui]** fix: clean X-Forwarded-Prefix header for the dashboard. ([#7109](https://github.com/traefik/traefik/pull/7109) by [ldez](https://github.com/ldez))

**Documentation:**
- **[docker]** spelling(docs/content/routing/providers/docker.md) ([#7101](https://github.com/traefik/traefik/pull/7101) by [szczot3k](https://github.com/szczot3k))
- **[k8s]** doc: add name of used key for kubernetes client auth ([#7068](https://github.com/traefik/traefik/pull/7068) by [smueller18](https://github.com/smueller18))

## [v2.2.7](https://github.com/traefik/traefik/tree/v2.2.7) (2020-07-20)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.6...v2.2.7)

**Bug fixes:**
- **[server,tls]** fix: drop host port to compare with SNI. ([#7071](https://github.com/traefik/traefik/pull/7071) by [ldez](https://github.com/ldez))

## [v2.2.6](https://github.com/traefik/traefik/tree/v2.2.6) (2020-07-17)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.5...v2.2.6)

**Bug fixes:**
- **[logs]** fix: access logs header names filtering is case insensitive ([#6900](https://github.com/traefik/traefik/pull/6900) by [mjeanroy](https://github.com/mjeanroy))
- **[provider]** Get Entrypoints Port Address without protocol for redirect ([#7047](https://github.com/traefik/traefik/pull/7047) by [SantoDE](https://github.com/SantoDE))
- **[tls]** Fix domain fronting ([#7064](https://github.com/traefik/traefik/pull/7064) by [juliens](https://github.com/juliens))

**Documentation:**
- fix: documentation references. ([#7049](https://github.com/traefik/traefik/pull/7049) by [ldez](https://github.com/ldez))
- Add example for entrypoint on one ip address ([#6483](https://github.com/traefik/traefik/pull/6483) by [SimonHeimberg](https://github.com/SimonHeimberg))

## [v2.3.0-rc2](https://github.com/traefik/traefik/tree/v2.3.0-rc2) (2020-07-15)
[All Commits](https://github.com/traefik/traefik/compare/v2.3.0-rc1...v2.3.0-rc2)

**Misc:**
- fix: goreleaser build commands.

## [v2.3.0-rc1](https://github.com/traefik/traefik/tree/v2.3.0-rc1) (2020-07-15)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.0-rc1...v2.3.0-rc1)

**Enhancements:**
- **[api]** Add custom ping http code when Traefik is terminating ([#6696](https://github.com/traefik/traefik/pull/6696) by [L3o-pold](https://github.com/L3o-pold))
- **[ecs]** Add AWS ECS provider ([#6749](https://github.com/traefik/traefik/pull/6749) by [alekitto](https://github.com/alekitto))
- **[file]** feat: use parser to load dynamic config from file. ([#6875](https://github.com/traefik/traefik/pull/6875) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd,k8s/ingress]** Upgrade Client-go to 0.18.2 ([#6779](https://github.com/traefik/traefik/pull/6779) by [dtomcej](https://github.com/dtomcej))
- **[k8s,k8s/ingress]** Add new ingressClass support to ingress provider ([#6831](https://github.com/traefik/traefik/pull/6831) by [dtomcej](https://github.com/dtomcej))
- **[plugins]** Traefik Pilot: plugins support and alert system (EXPERIMENTAL FEATURES) ([#7041](https://github.com/traefik/traefik/pull/7041) by [ldez](https://github.com/ldez))
- **[provider]** Add HTTP Provider ([#6976](https://github.com/traefik/traefik/pull/6976) by [kevinpollet](https://github.com/kevinpollet))
- **[webui]** Add iOS specific icons ([#6946](https://github.com/traefik/traefik/pull/6946) by [Heisenberg74](https://github.com/Heisenberg74))

**Bug fixes:**
- **[k8s,k8s/ingress]** Delete an unnecessary warning log ([#6568](https://github.com/traefik/traefik/pull/6568) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware]** Allow multiple secure middlewares to operate independently  ([#6604](https://github.com/traefik/traefik/pull/6604) by [dtomcej](https://github.com/dtomcej))

**Documentation:**
- **[middleware]** Fixes config samples regarding forceSlash option ([#6811](https://github.com/traefik/traefik/pull/6811) by [volkerw00](https://github.com/volkerw00))

**Misc:**
- Merge current v2.2 branch into master ([#7052](https://github.com/traefik/traefik/pull/7052) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into master ([#7022](https://github.com/traefik/traefik/pull/7022) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.2 branch into master ([#6921](https://github.com/traefik/traefik/pull/6921) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge current v2.2 branch into master ([#6822](https://github.com/traefik/traefik/pull/6822) by [mmatur](https://github.com/mmatur))
- Merge current v2.2 branch into master ([#6754](https://github.com/traefik/traefik/pull/6754) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into master ([#6533](https://github.com/traefik/traefik/pull/6533) by [ldez](https://github.com/ldez))
- Merge current v2.2 branch into master ([#6468](https://github.com/traefik/traefik/pull/6468) by [ldez](https://github.com/ldez))

## [v2.2.5](https://github.com/traefik/traefik/tree/v2.2.5) (2020-07-13)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.4...v2.2.5)

**Bug fixes:**
- **[k8s,k8s/crd]** fix k8s crd to read contentType middleware into dynamic config ([#7034](https://github.com/traefik/traefik/pull/7034) by [johnpekcan](https://github.com/johnpekcan))
- **[rules,server,tls]** Revert domain fronting fix ([#7039](https://github.com/traefik/traefik/pull/7039) by [rtribotte](https://github.com/rtribotte))
- **[tls]** Fix default value for InsecureSNI when global is not set ([#7037](https://github.com/traefik/traefik/pull/7037) by [juliens](https://github.com/juliens))

## [v2.2.4](https://github.com/traefik/traefik/tree/v2.2.4) (2020-07-10)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.3...v2.2.4)

**Bug fixes:**
- **[tls]** Change the default value of insecureSNI ([#7027](https://github.com/traefik/traefik/pull/7027) by [jbdoumenjou](https://github.com/jbdoumenjou))

## [v2.2.3](https://github.com/traefik/traefik/tree/v2.2.3) (2020-07-09)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.2...v2.2.3)

**Bug fixes:**
- **[middleware]** Fix panic when using chain middleware. ([#7016](https://github.com/traefik/traefik/pull/7016) by [juliens](https://github.com/juliens))

## [v2.2.2](https://github.com/traefik/traefik/tree/v2.2.2) (2020-07-08)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.1...v2.2.2)

**Bug fixes:**
- **[acme]** Update go-acme/lego to v3.8.0 ([#6988](https://github.com/traefik/traefik/pull/6988) by [ldez](https://github.com/ldez))
- **[acme]** Fix triggering multiple concurrent requests to ACME ([#6939](https://github.com/traefik/traefik/pull/6939) by [ddtmachado](https://github.com/ddtmachado))
- **[acme]** Update go-acme/lego to v3.7.0 ([#6792](https://github.com/traefik/traefik/pull/6792) by [ldez](https://github.com/ldez))
- **[acme]** added required quotes to domains config ([#6867](https://github.com/traefik/traefik/pull/6867) by [tompson](https://github.com/tompson))
- **[authentication,logs,middleware]** Provide username in log data on auth failure ([#6827](https://github.com/traefik/traefik/pull/6827) by [rtribotte](https://github.com/rtribotte))
- **[docker]** Use specified network for &#34;container&#34; network mode ([#6763](https://github.com/traefik/traefik/pull/6763) by [bjeanes](https://github.com/bjeanes))
- **[k8s,k8s/crd]** Remove checkStringQuoteValidity in loadIngressRouteConf ([#6775](https://github.com/traefik/traefik/pull/6775) by [fefe982](https://github.com/fefe982))
- **[middleware,websocket]** Fix wss in x-forwarded-proto ([#6752](https://github.com/traefik/traefik/pull/6752) by [juliens](https://github.com/juliens))
- **[middleware]** internal handlers: support for response modifiers ([#6750](https://github.com/traefik/traefik/pull/6750) by [mpl](https://github.com/mpl))
- **[middleware]** Fix ipv6 handling in redirect middleware ([#6902](https://github.com/traefik/traefik/pull/6902) by [rtribotte](https://github.com/rtribotte))
- **[middleware]** refactor X-Forwarded-Proto ([#6863](https://github.com/traefik/traefik/pull/6863) by [jcgruenhage](https://github.com/jcgruenhage))
- **[provider]** Fix race condition issues with provided dynamic configuration ([#6979](https://github.com/traefik/traefik/pull/6979) by [kevinpollet](https://github.com/kevinpollet))
- **[rules,server,tls]** Disable domain fronting ([#7008](https://github.com/traefik/traefik/pull/7008) by [rtribotte](https://github.com/rtribotte))
- **[udp]** Fix mem leak on UDP connections ([#6815](https://github.com/traefik/traefik/pull/6815) by [ddtmachado](https://github.com/ddtmachado))
- **[udp]** Avoid overwriting already received UDP messages ([#6797](https://github.com/traefik/traefik/pull/6797) by [cbachert](https://github.com/cbachert))
- **[webui]** Add missing accessControlAllowOrigin list to middleware view ([#6747](https://github.com/traefik/traefik/pull/6747) by [barthez](https://github.com/barthez))

**Documentation:**
- **[acme]** Fix doc url for Aurora DNS provider ([#6899](https://github.com/traefik/traefik/pull/6899) by [rtribotte](https://github.com/rtribotte))
- **[acme]** Fix acme.md typo ([#6817](https://github.com/traefik/traefik/pull/6817) by [juliocc](https://github.com/juliocc))
- **[acme]** fix certResolver typo ([#6983](https://github.com/traefik/traefik/pull/6983) by [DavidBadura](https://github.com/DavidBadura))
- **[acme]** Fix statement about lego _FILE env var ([#6964](https://github.com/traefik/traefik/pull/6964) by [solvaholic](https://github.com/solvaholic))
- **[acme]** Improve acme CLI options in Let&#39;s Encrypt documentation  ([#6762](https://github.com/traefik/traefik/pull/6762) by [netoax](https://github.com/netoax))
- **[docker]** fix a broken link on Docker plugins documentation ([#6908](https://github.com/traefik/traefik/pull/6908) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[docker]** Fix healthcheck.interval in docs ([#6847](https://github.com/traefik/traefik/pull/6847) by [OndrejIT](https://github.com/OndrejIT))
- **[k8s,k8s/ingress]** Remove redundant paragraph in Kubernetes ingress documentation ([#6806](https://github.com/traefik/traefik/pull/6806) by [lpfann](https://github.com/lpfann))
- **[k8s,k8s/ingress]** Fix sticky cookie ingress annotation doc ([#6938](https://github.com/traefik/traefik/pull/6938) by [rtribotte](https://github.com/rtribotte))
- **[k8s]** fixing typo in Provider KubernetesIngress at Routing documentation ([#6845](https://github.com/traefik/traefik/pull/6845) by [sw360cab](https://github.com/sw360cab))
- **[k8s]** Update kubernetes-crd.md ([#6878](https://github.com/traefik/traefik/pull/6878) by [rherrick](https://github.com/rherrick))
- **[logs]** Fixed incorrect logging parameter in documentation ([#6819](https://github.com/traefik/traefik/pull/6819) by [cplewnia](https://github.com/cplewnia))
- **[logs]** Use &#34;headers&#34; instead of &#34;header&#34; in access log docs ([#6836](https://github.com/traefik/traefik/pull/6836) by [bradjones1](https://github.com/bradjones1))
- **[middleware,k8s/crd]** Fix Headers middleware documentation, usage of proper bool ([#6928](https://github.com/traefik/traefik/pull/6928) by [rtribotte](https://github.com/rtribotte))
- **[middleware]** Improve redirectScheme documentation ([#6769](https://github.com/traefik/traefik/pull/6769) by [dtomcej](https://github.com/dtomcej))
- **[middleware]** Update basicauth.md ([#6967](https://github.com/traefik/traefik/pull/6967) by [vitalets](https://github.com/vitalets))
- Update Dashboard examples and move it after &#39;Router Rule&#39; section ([#6874](https://github.com/traefik/traefik/pull/6874) by [ddtmachado](https://github.com/ddtmachado))
- Fix log field names in documentation ([#6952](https://github.com/traefik/traefik/pull/6952) by [gysel](https://github.com/gysel))
- Minor fix to Go templating documentation ([#6977](https://github.com/traefik/traefik/pull/6977) by [PCM2](https://github.com/PCM2))
- Add rtribotte to maintainers ([#6936](https://github.com/traefik/traefik/pull/6936) by [emilevauge](https://github.com/emilevauge))
- Update Copyright ([#6795](https://github.com/traefik/traefik/pull/6795) by [mmatur](https://github.com/mmatur))
- fix: dead link. ([#6876](https://github.com/traefik/traefik/pull/6876) by [ldez](https://github.com/ldez))
- Fix v1-&gt; v2 migration: unify domain name in documentation example ([#6904](https://github.com/traefik/traefik/pull/6904) by [sinacek](https://github.com/sinacek))

## [v2.2.1](https://github.com/traefik/traefik/tree/v2.2.1) (2020-04-29)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.0...v2.2.1)

**Bug fixes:**
- **[acme]** Update go-acme/lego to v3.6.0 ([#6727](https://github.com/traefik/traefik/pull/6727) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Normalize default names for ConsulCatalog. ([#6593](https://github.com/traefik/traefik/pull/6593) by [ldez](https://github.com/ldez))
- **[internal]** Change the default priority on the router created by the redirect. ([#6588](https://github.com/traefik/traefik/pull/6588) by [ldez](https://github.com/ldez))
- **[k8s,k8s/ingress]** Delete an unnecessary warning log ([#6624](https://github.com/traefik/traefik/pull/6624) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware]** ratelimit: do not default to ipstrategy too early ([#6713](https://github.com/traefik/traefik/pull/6713) by [mpl](https://github.com/mpl))
- **[rancher,webui]** It&#39;s just the one TLS, actually. ([#6606](https://github.com/traefik/traefik/pull/6606) by [RealOrangeOne](https://github.com/RealOrangeOne))
- **[server]** Fix case-sensitive header Sec-Websocket-Version ([#6698](https://github.com/traefik/traefik/pull/6698) by [tbrandstetter](https://github.com/tbrandstetter))
- **[udp]** fix: consider UDP when checking for empty config ([#6683](https://github.com/traefik/traefik/pull/6683) by [nrwiersma](https://github.com/nrwiersma))
- **[websocket]** FIx wS heAder ([#6660](https://github.com/traefik/traefik/pull/6660) by [mmatur](https://github.com/mmatur))
- **[websocket]** Manage case for all Websocket headers ([#6705](https://github.com/traefik/traefik/pull/6705) by [mmatur](https://github.com/mmatur))
- **[webui]** Disable distribution of the WebUI as PWA ([#6717](https://github.com/traefik/traefik/pull/6717) by [SantoDE](https://github.com/SantoDE))
- **[webui]** Add polling for getOverview in toolbar ([#6611](https://github.com/traefik/traefik/pull/6611) by [lukashass](https://github.com/lukashass))

**Documentation:**
- **[api]** Fix documentation about api.insecure defaults ([#6671](https://github.com/traefik/traefik/pull/6671) by [thisismydesign](https://github.com/thisismydesign))
- **[docker,k8s,k8s/ingress,marathon,rancher,sticky-session]** fix: cookie documentation. ([#6745](https://github.com/traefik/traefik/pull/6745) by [ldez](https://github.com/ldez))
- **[file]** Edit code indentation for correct alignment ([#6691](https://github.com/traefik/traefik/pull/6691) by [fbruetting](https://github.com/fbruetting))
- **[healthcheck,k8s,k8s/crd]** Add note about health check in kubernetes ([#6647](https://github.com/traefik/traefik/pull/6647) by [mmatur](https://github.com/mmatur))
- **[k8s,k8s/crd]** docs: Update kubernetes-crd-resource.yml ([#6741](https://github.com/traefik/traefik/pull/6741) by [rdxmb](https://github.com/rdxmb))
- **[k8s,k8s/crd]** doc: improve CRD documentation. ([#6681](https://github.com/traefik/traefik/pull/6681) by [ldez](https://github.com/ldez))
- **[k8s/crd]** doc: add apiVersion for &#34;kind: Middleware&#34; ([#6734](https://github.com/traefik/traefik/pull/6734) by [yuyicai](https://github.com/yuyicai))
- **[k8s/helm]** Update the documentation for helm chart ([#6744](https://github.com/traefik/traefik/pull/6744) by [mmatur](https://github.com/mmatur))
- **[k8s]** Add sentence about the resource namespace and middleware ([#6719](https://github.com/traefik/traefik/pull/6719) by [SantoDE](https://github.com/SantoDE))
- **[kv]** fix KV service docs for http:url and tcp:address ([#6720](https://github.com/traefik/traefik/pull/6720) by [bryfry](https://github.com/bryfry))
- **[logs]** Add Access log chapter for migration v1-&gt;v2 ([#6689](https://github.com/traefik/traefik/pull/6689) by [MartinKoerner](https://github.com/MartinKoerner))
- **[middleware]** Update headers.md ([#6675](https://github.com/traefik/traefik/pull/6675) by [jamct](https://github.com/jamct))
- **[middleware]** Doc middleware compress content type ([#6738](https://github.com/traefik/traefik/pull/6738) by [rtribotte](https://github.com/rtribotte))
- **[tracing]** Add link to tracing with elastic ([#6673](https://github.com/traefik/traefik/pull/6673) by [collinmutembei](https://github.com/collinmutembei))
- Added missing text `a yaml file` in Configuration ([#6663](https://github.com/traefik/traefik/pull/6663) by [fsoedjede](https://github.com/fsoedjede))
- Fix typos in the documentation ([#6650](https://github.com/traefik/traefik/pull/6650) by [SuperSandro2000](https://github.com/SuperSandro2000))
- Fix documentation ([#6648](https://github.com/traefik/traefik/pull/6648) by [mmatur](https://github.com/mmatur))
- Fix bad address syntax in Global HTTP to HTTPS redirection v2 TOML ([#6619](https://github.com/traefik/traefik/pull/6619) by [Beetix](https://github.com/Beetix))
- Doc Fix for 2.2 Redirects ([#6595](https://github.com/traefik/traefik/pull/6595) by [ajschmidt8](https://github.com/ajschmidt8))

## [v2.2.0](https://github.com/traefik/traefik/tree/v2.2.0) (2020-03-25)
[All Commits](https://github.com/traefik/traefik/compare/v2.1.0-rc1...v2.2.0)

**Enhancements:**
- **[acme,middleware,tls]** Entry point redirection and default routers configuration ([#6417](https://github.com/traefik/traefik/pull/6417) by [ldez](https://github.com/ldez))
- **[consul,etcd,kv,redis,zk]** Add KV store providers (dynamic configuration only) ([#5899](https://github.com/traefik/traefik/pull/5899) by [ldez](https://github.com/ldez))
- **[consulcatalog,docker,marathon,rancher,udp]** Add UDP in providers with labels ([#6327](https://github.com/traefik/traefik/pull/6327) by [juliens](https://github.com/juliens))
- **[docker]** Fix traefik behavior when network_mode is host ([#5698](https://github.com/traefik/traefik/pull/5698) by [FuNK3Y](https://github.com/FuNK3Y))
- **[docker]** Support SSH connection to Docker ([#5969](https://github.com/traefik/traefik/pull/5969) by [sh7dm](https://github.com/sh7dm))
- **[healthcheck]** Do not follow redirects for the health check URLs ([#5147](https://github.com/traefik/traefik/pull/5147) by [coder-hugo](https://github.com/coder-hugo))
- **[k8s,k8s/crd,udp]** Add UDP support in kubernetesCRD provider ([#6348](https://github.com/traefik/traefik/pull/6348) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/crd]** Add TLSStores to Kubernetes CRD ([#6270](https://github.com/traefik/traefik/pull/6270) by [dtomcej](https://github.com/dtomcej))
- **[k8s,k8s/crd]** Add namespace attribute on IngressRouteTCP service ([#6085](https://github.com/traefik/traefik/pull/6085) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/ingress]** Support &#39;networking.k8s.io/v1beta1&#39; ingress apiVersion ([#6171](https://github.com/traefik/traefik/pull/6171) by [ldez](https://github.com/ldez))
- **[k8s,k8s/ingress]** Update deprecated function call in k8s providers  ([#5241](https://github.com/traefik/traefik/pull/5241) by [Wagum](https://github.com/Wagum))
- **[k8s,k8s/ingress]** Add Ingress annotations support ([#6160](https://github.com/traefik/traefik/pull/6160) by [ldez](https://github.com/ldez))
- **[k8s,k8s/ingress]** systematically call updateIngressStatus ([#6148](https://github.com/traefik/traefik/pull/6148) by [mpl](https://github.com/mpl))
- **[logs,middleware]** Rename the non-exposed field &#34;count&#34; to &#34;size&#34; ([#6048](https://github.com/traefik/traefik/pull/6048) by [sylr](https://github.com/sylr))
- **[logs,middleware]** Add http request scheme to logger ([#6226](https://github.com/traefik/traefik/pull/6226) by [valtlfelipe](https://github.com/valtlfelipe))
- **[logs]** Decrease log level for client related error ([#6204](https://github.com/traefik/traefik/pull/6204) by [sylr](https://github.com/sylr))
- **[metrics]** Add metrics about TLS ([#6255](https://github.com/traefik/traefik/pull/6255) by [sylr](https://github.com/sylr))
- **[middleware]** Add period for rate limiter middleware ([#6055](https://github.com/traefik/traefik/pull/6055) by [mpl](https://github.com/mpl))
- **[middleware]** Let metrics libs handle the atomicity ([#5738](https://github.com/traefik/traefik/pull/5738) by [sylr](https://github.com/sylr))
- **[middleware]** Rework access control origin configuration ([#5996](https://github.com/traefik/traefik/pull/5996) by [dtomcej](https://github.com/dtomcej))
- **[middleware]** Add serial number certificate to forward headers ([#5915](https://github.com/traefik/traefik/pull/5915) by [dkijkuit](https://github.com/dkijkuit))
- **[rancher]** Duration order consistency when multiplying number by time unit ([#5885](https://github.com/traefik/traefik/pull/5885) by [maxifom](https://github.com/maxifom))
- **[server,udp]** UDP support ([#6172](https://github.com/traefik/traefik/pull/6172) by [mpl](https://github.com/mpl))
- **[service]** Use EDF schedule algorithm for WeightedRoundRobin ([#6206](https://github.com/traefik/traefik/pull/6206) by [pkumza](https://github.com/pkumza))
- **[service]** Support mirroring request body ([#6080](https://github.com/traefik/traefik/pull/6080) by [dmitriyminer](https://github.com/dmitriyminer))
- **[tls]** Allow PreferServerCipherSuites as a TLS Option ([#6248](https://github.com/traefik/traefik/pull/6248) by [dtomcej](https://github.com/dtomcej))
- **[tracing]** Update APM client. ([#6152](https://github.com/traefik/traefik/pull/6152) by [ldez](https://github.com/ldez))
- **[tracing]** Elastic APM tracer implementation ([#5870](https://github.com/traefik/traefik/pull/5870) by [amine7536](https://github.com/amine7536))
- **[udp,webui]** WebUI: add udp pages ([#6313](https://github.com/traefik/traefik/pull/6313) by [matthieuh](https://github.com/matthieuh))
- **[webui]** Web UI: Polling on tables ([#5909](https://github.com/traefik/traefik/pull/5909) by [matthieuh](https://github.com/matthieuh))
- **[webui]** Proxy API to Traefik in dev mode ([#5980](https://github.com/traefik/traefik/pull/5980) by [sh7dm](https://github.com/sh7dm))
- **[webui]** Web UI: Table infinite scroll ([#5875](https://github.com/traefik/traefik/pull/5875) by [matthieuh](https://github.com/matthieuh))
- **[webui]** Web UI: Take off logic from generic table component ([#5910](https://github.com/traefik/traefik/pull/5910) by [matthieuh](https://github.com/matthieuh))
- **[webui]** Add dark theme for Web UI ([#6036](https://github.com/traefik/traefik/pull/6036) by [sh7dm](https://github.com/sh7dm))
- Update dependencies ([#6359](https://github.com/traefik/traefik/pull/6359) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[acme]** Update go-acme/lego to v3.5.0 ([#6491](https://github.com/traefik/traefik/pull/6491) by [ldez](https://github.com/ldez))
- **[authentication,middleware]** digest auth: use RequireAuthStale when appropriate ([#6569](https://github.com/traefik/traefik/pull/6569) by [mpl](https://github.com/mpl))
- **[file]** Revert &#34;Allow fsnotify to reload config files on k8s (or symlinks)&#34; ([#6416](https://github.com/traefik/traefik/pull/6416) by [juliens](https://github.com/juliens))
- **[internal]** Fix entry point redirect behavior ([#6512](https://github.com/traefik/traefik/pull/6512) by [ldez](https://github.com/ldez))
- **[internal]** Router entry points on reload. ([#6444](https://github.com/traefik/traefik/pull/6444) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd]** Improve kubernetes external name service support ([#6428](https://github.com/traefik/traefik/pull/6428) by [rtribotte](https://github.com/rtribotte))
- **[k8s,k8s/ingress]** fix: Ingress TLS support ([#6504](https://github.com/traefik/traefik/pull/6504) by [ldez](https://github.com/ldez))
- **[k8s,k8s/ingress]** Improvement of the unique name of the router for Ingress. ([#6325](https://github.com/traefik/traefik/pull/6325) by [ldez](https://github.com/ldez))
- **[kv,redis]** Update valkeyrie to fix the support of Redis. ([#6291](https://github.com/traefik/traefik/pull/6291) by [ldez](https://github.com/ldez))
- **[kv]** fix: KV flaky tests. ([#6300](https://github.com/traefik/traefik/pull/6300) by [ldez](https://github.com/ldez))
- **[etcd,kv]** fix: etcd provider name. ([#6212](https://github.com/traefik/traefik/pull/6212) by [ldez](https://github.com/ldez))
- **[middleware]** fix: period field name. ([#6549](https://github.com/traefik/traefik/pull/6549) by [ldez](https://github.com/ldez))
- **[middleware]** fix: custom Host header. ([#6502](https://github.com/traefik/traefik/pull/6502) by [ldez](https://github.com/ldez))
- **[server,udp]** udp: replace concurrently reset timer with ticker ([#6498](https://github.com/traefik/traefik/pull/6498) by [mpl](https://github.com/mpl))
- **[server]** Drop traefik from default entry points. ([#6477](https://github.com/traefik/traefik/pull/6477) by [ldez](https://github.com/ldez))
- **[server]** fix: use MaxInt32. ([#5845](https://github.com/traefik/traefik/pull/5845) by [ldez](https://github.com/ldez))
- **[tracing]** Disable default APM tracer. ([#6410](https://github.com/traefik/traefik/pull/6410) by [ldez](https://github.com/ldez))
- **[udp]** Add missing generated element for UDP. ([#6309](https://github.com/traefik/traefik/pull/6309) by [ldez](https://github.com/ldez))
- **[udp]** Build all UDP services on an entrypoint ([#6329](https://github.com/traefik/traefik/pull/6329) by [juliens](https://github.com/juliens))

**Documentation:**
- **[authentication,middleware]** docs: terminology, replace &#39;encoded&#39; by &#39;hashed&#39; ([#6478](https://github.com/traefik/traefik/pull/6478) by [debovema](https://github.com/debovema))
- **[acme]** Doc: fix wrong name of config format ([#6519](https://github.com/traefik/traefik/pull/6519) by [Nek-](https://github.com/Nek-))
- **[docker]** Fix example values for swarmModeRefreshSeconds ([#6460](https://github.com/traefik/traefik/pull/6460) by [skjnldsv](https://github.com/skjnldsv))
- **[k8s,k8s/crd,sticky-session]** docs: clarify multi-levels stickiness ([#6475](https://github.com/traefik/traefik/pull/6475) by [mpl](https://github.com/mpl))
- **[k8s,k8s/crd]** doc: fix terminationDelay word case. ([#6532](https://github.com/traefik/traefik/pull/6532) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd]** Update the k8s CRD documentation ([#6426](https://github.com/traefik/traefik/pull/6426) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/ingress]** Improve documentation for kubernetes ingress configuration ([#6440](https://github.com/traefik/traefik/pull/6440) by [rtribotte](https://github.com/rtribotte))
- **[k8s/helm]** Update traefik install documentation ([#6466](https://github.com/traefik/traefik/pull/6466) by [mmatur](https://github.com/mmatur))
- **[provider]** Update supported providers list. ([#6190](https://github.com/traefik/traefik/pull/6190) by [ldez](https://github.com/ldez))
- **[tcp,tls]** Specify passthrough for TCP/TLS in its own section ([#6459](https://github.com/traefik/traefik/pull/6459) by [mpl](https://github.com/mpl))
- doc: Use neutral domains. ([#6471](https://github.com/traefik/traefik/pull/6471) by [ldez](https://github.com/ldez))
- doc: fix typo. ([#6472](https://github.com/traefik/traefik/pull/6472) by [ldez](https://github.com/ldez))
- Improve ping documentation. ([#6476](https://github.com/traefik/traefik/pull/6476) by [ldez](https://github.com/ldez))
- Remove  @dduportal from the maintainers team ([#6464](https://github.com/traefik/traefik/pull/6464) by [emilevauge](https://github.com/emilevauge))
- Fix wrong copy/pasted with service name warning ([#6510](https://github.com/traefik/traefik/pull/6510) by [Nek-](https://github.com/Nek-))
- Update migration documentation ([#6447](https://github.com/traefik/traefik/pull/6447) by [ldez](https://github.com/ldez))
- Update version references. ([#6434](https://github.com/traefik/traefik/pull/6434) by [ldez](https://github.com/ldez))
- Fix broken documentation link ([#6430](https://github.com/traefik/traefik/pull/6430) by [pbek](https://github.com/pbek))

**Misc:**
- **[rancher]** Stop using fork of go-rancher-metadata ([#6469](https://github.com/traefik/traefik/pull/6469) by [ibuildthecloud](https://github.com/ibuildthecloud))
- Merge current v2.1 branch into v2.2 ([#6564](https://github.com/traefik/traefik/pull/6564) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into v2.2 ([#6525](https://github.com/traefik/traefik/pull/6525) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into v2.2 ([#6516](https://github.com/traefik/traefik/pull/6516) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into master ([#6429](https://github.com/traefik/traefik/pull/6429) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into master ([#6409](https://github.com/traefik/traefik/pull/6409) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into master ([#6302](https://github.com/traefik/traefik/pull/6302) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into master ([#6216](https://github.com/traefik/traefik/pull/6216) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into master ([#6138](https://github.com/traefik/traefik/pull/6138) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into master ([#6004](https://github.com/traefik/traefik/pull/6004) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into master ([#5933](https://github.com/traefik/traefik/pull/5933) by [ldez](https://github.com/ldez))

## [v2.1.9](https://github.com/traefik/traefik/tree/v2.1.9) (2020-03-23)
[All Commits](https://github.com/traefik/traefik/compare/v2.1.8...v2.1.9)

**Bug fixes:**
- **[provider,sticky-session]** Fix sameSite ([#6538](https://github.com/traefik/traefik/pull/6538) by [ldez](https://github.com/ldez))
- **[server]** Force http/1.1 for upgrade ([#6554](https://github.com/traefik/traefik/pull/6554) by [juliens](https://github.com/juliens))

**Documentation:**
- Fix tab name ([#6543](https://github.com/traefik/traefik/pull/6543) by [mavimo](https://github.com/mavimo))

## [v2.2.0-rc4](https://github.com/traefik/traefik/tree/v2.2.0-rc4) (2020-03-19)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.0-rc3...v2.2.0-rc4)

**Documentation:**
- **[acme]** Doc: fix wrong name of config format ([#6519](https://github.com/traefik/traefik/pull/6519) by [Nek-](https://github.com/Nek-))

**Misc:**
- **[middleware]** Merge current v2.1 branch into v2.2 ([#6525](https://github.com/traefik/traefik/pull/6525) by [ldez](https://github.com/ldez))

## [v2.1.8](https://github.com/traefik/traefik/tree/v2.1.8) (2020-03-19)
[All Commits](https://github.com/traefik/traefik/compare/v2.1.7...v2.1.8)

**Bug fixes:**
- **[middleware,metrics]** Fix memory leak in metrics ([#6522](https://github.com/traefik/traefik/pull/6522) by [juliens](https://github.com/juliens))

## [v2.2.0-rc3](https://github.com/traefik/traefik/tree/v2.2.0-rc3) (2020-03-18)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.0-rc2...v2.2.0-rc3)

**Enhancements:**
- **[authentication,middleware]** docs: terminology, replace &#39;encoded&#39; by &#39;hashed&#39; ([#6478](https://github.com/traefik/traefik/pull/6478) by [debovema](https://github.com/debovema))

**Bug fixes:**
- **[acme]** Update go-acme/lego to v3.5.0 ([#6491](https://github.com/traefik/traefik/pull/6491) by [ldez](https://github.com/ldez))
- **[internal]** Fix entry point redirect behavior ([#6512](https://github.com/traefik/traefik/pull/6512) by [ldez](https://github.com/ldez))
- **[k8s,k8s/ingress]** fix: Ingress TLS support ([#6504](https://github.com/traefik/traefik/pull/6504) by [ldez](https://github.com/ldez))
- **[middleware]** fix: custom Host header. ([#6502](https://github.com/traefik/traefik/pull/6502) by [ldez](https://github.com/ldez))
- **[server,udp]** udp: replace concurrently reset timer with ticker ([#6498](https://github.com/traefik/traefik/pull/6498) by [mpl](https://github.com/mpl))
- **[server]** Drop traefik from default entry points. ([#6477](https://github.com/traefik/traefik/pull/6477) by [ldez](https://github.com/ldez))

**Documentation:**
- **[k8s,k8s/crd,sticky-session]** docs: clarify multi-levels stickiness ([#6475](https://github.com/traefik/traefik/pull/6475) by [mpl](https://github.com/mpl))
- **[k8s/helm]** Update traefik install documentation ([#6466](https://github.com/traefik/traefik/pull/6466) by [mmatur](https://github.com/mmatur))
- Fix wrong copy/pasted with service name warning ([#6510](https://github.com/traefik/traefik/pull/6510) by [Nek-](https://github.com/Nek-))
- Improve ping documentation. ([#6476](https://github.com/traefik/traefik/pull/6476) by [ldez](https://github.com/ldez))
- doc: fix typo. ([#6472](https://github.com/traefik/traefik/pull/6472) by [ldez](https://github.com/ldez))
- doc: Use neutral domains. ([#6471](https://github.com/traefik/traefik/pull/6471) by [ldez](https://github.com/ldez))

**Misc:**
- **[rancher]** Stop using fork of go-rancher-metadata ([#6469](https://github.com/traefik/traefik/pull/6469) by [ibuildthecloud](https://github.com/ibuildthecloud))

## [v2.1.7](https://github.com/traefik/traefik/tree/v2.1.7) (2020-03-18)
[All Commits](https://github.com/traefik/traefik/compare/v2.1.6...v2.1.7)

**Bug fixes:**
- **[logs,middleware]** Access log field quotes. ([#6484](https://github.com/traefik/traefik/pull/6484) by [ldez](https://github.com/ldez))
- **[metrics]** fix statsd scale for duration based metrics ([#6054](https://github.com/traefik/traefik/pull/6054) by [ddtmachado](https://github.com/ddtmachado))
- **[middleware]** Added support for replacement containing escaped characters ([#6413](https://github.com/traefik/traefik/pull/6413) by [rtribotte](https://github.com/rtribotte))

**Documentation:**
- **[acme,docker]** Add some missing doc. ([#6422](https://github.com/traefik/traefik/pull/6422) by [ldez](https://github.com/ldez))
- **[acme]** Added wildcard ACME example ([#6423](https://github.com/traefik/traefik/pull/6423) by [Basster](https://github.com/Basster))
- **[acme]** fix typo ([#6408](https://github.com/traefik/traefik/pull/6408) by [hamiltont](https://github.com/hamiltont))

## [v2.2.0-rc2](https://github.com/traefik/traefik/tree/v2.2.0-rc2) (2020-03-11)
[All Commits](https://github.com/traefik/traefik/compare/v2.2.0-rc1...v2.2.0-rc2)

**Bug fixes:**
- **[internal]** Router entry points on reload. ([#6444](https://github.com/traefik/traefik/pull/6444) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd]** Improve kubernetes external name service support ([#6428](https://github.com/traefik/traefik/pull/6428) by [rtribotte](https://github.com/rtribotte))

**Documentation:**
- **[docker]** Fix example values for swarmModeRefreshSeconds ([#6460](https://github.com/traefik/traefik/pull/6460) by [skjnldsv](https://github.com/skjnldsv))
- **[k8s,k8s/ingress]** Improve documentation for kubernetes ingress configuration ([#6440](https://github.com/traefik/traefik/pull/6440) by [rtribotte](https://github.com/rtribotte))
- **[tcp,tls]** Specify passthrough for TCP/TLS in its own section ([#6459](https://github.com/traefik/traefik/pull/6459) by [mpl](https://github.com/mpl))
- Remove  @dduportal from the maintainers team ([#6464](https://github.com/traefik/traefik/pull/6464) by [emilevauge](https://github.com/emilevauge))
- Update migration documentation ([#6447](https://github.com/traefik/traefik/pull/6447) by [ldez](https://github.com/ldez))
- Update version references. ([#6434](https://github.com/traefik/traefik/pull/6434) by [ldez](https://github.com/ldez))
- Fix broken documentation link ([#6430](https://github.com/traefik/traefik/pull/6430) by [pbek](https://github.com/pbek))

## [v2.2.0-rc1](https://github.com/traefik/traefik/tree/v2.2.0-rc1) (2020-03-05)
[All Commits](https://github.com/traefik/traefik/compare/v2.1.0-rc1...v2.2.0-rc1)

**Enhancements:**
- **[acme,middleware,tls]** Entry point redirection and default routers configuration ([#6417](https://github.com/traefik/traefik/pull/6417) by [ldez](https://github.com/ldez))
- **[consul,etcd,kv,redis,zk]** Add KV store providers (dynamic configuration only) ([#5899](https://github.com/traefik/traefik/pull/5899) by [ldez](https://github.com/ldez))
- **[consulcatalog,docker,marathon,rancher,udp]** Add UDP in providers with labels ([#6327](https://github.com/traefik/traefik/pull/6327) by [juliens](https://github.com/juliens))
- **[docker]** Fix traefik behavior when network_mode is host ([#5698](https://github.com/traefik/traefik/pull/5698) by [FuNK3Y](https://github.com/FuNK3Y))
- **[docker]** Support SSH connection to Docker ([#5969](https://github.com/traefik/traefik/pull/5969) by [sh7dm](https://github.com/sh7dm))
- **[healthcheck]** Do not follow redirects for the health check URLs ([#5147](https://github.com/traefik/traefik/pull/5147) by [coder-hugo](https://github.com/coder-hugo))
- **[k8s,k8s/crd,udp]** Add UDP support in kubernetesCRD provider ([#6348](https://github.com/traefik/traefik/pull/6348) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/crd]** Add TLSStores to Kubernetes CRD ([#6270](https://github.com/traefik/traefik/pull/6270) by [dtomcej](https://github.com/dtomcej))
- **[k8s,k8s/crd]** Add namespace attribute on IngressRouteTCP service ([#6085](https://github.com/traefik/traefik/pull/6085) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/ingress]** Support &#39;networking.k8s.io/v1beta1&#39; ingress apiVersion ([#6171](https://github.com/traefik/traefik/pull/6171) by [ldez](https://github.com/ldez))
- **[k8s,k8s/ingress]** Update deprecated function call in k8s providers  ([#5241](https://github.com/traefik/traefik/pull/5241) by [Wagum](https://github.com/Wagum))
- **[k8s,k8s/ingress]** Add Ingress annotations support ([#6160](https://github.com/traefik/traefik/pull/6160) by [ldez](https://github.com/ldez))
- **[k8s,k8s/ingress]** systematically call updateIngressStatus ([#6148](https://github.com/traefik/traefik/pull/6148) by [mpl](https://github.com/mpl))
- **[logs,middleware]** Rename the non-exposed field &#34;count&#34; to &#34;size&#34; ([#6048](https://github.com/traefik/traefik/pull/6048) by [sylr](https://github.com/sylr))
- **[logs,middleware]** Add http request scheme to logger ([#6226](https://github.com/traefik/traefik/pull/6226) by [valtlfelipe](https://github.com/valtlfelipe))
- **[logs]** Decrease log level for client related error ([#6204](https://github.com/traefik/traefik/pull/6204) by [sylr](https://github.com/sylr))
- **[metrics]** Add metrics about TLS ([#6255](https://github.com/traefik/traefik/pull/6255) by [sylr](https://github.com/sylr))
- **[middleware]** Add period for rate limiter middleware ([#6055](https://github.com/traefik/traefik/pull/6055) by [mpl](https://github.com/mpl))
- **[middleware]** Let metrics libs handle the atomicity ([#5738](https://github.com/traefik/traefik/pull/5738) by [sylr](https://github.com/sylr))
- **[middleware]** Rework access control origin configuration ([#5996](https://github.com/traefik/traefik/pull/5996) by [dtomcej](https://github.com/dtomcej))
- **[middleware]** Add serial number certificate to forward headers ([#5915](https://github.com/traefik/traefik/pull/5915) by [dkijkuit](https://github.com/dkijkuit))
- **[rancher]** Duration order consistency when multiplying number by time unit ([#5885](https://github.com/traefik/traefik/pull/5885) by [maxifom](https://github.com/maxifom))
- **[server,udp]** UDP support ([#6172](https://github.com/traefik/traefik/pull/6172) by [mpl](https://github.com/mpl))
- **[service]** Use EDF schedule algorithm for WeightedRoundRobin ([#6206](https://github.com/traefik/traefik/pull/6206) by [pkumza](https://github.com/pkumza))
- **[service]** Support mirroring request body ([#6080](https://github.com/traefik/traefik/pull/6080) by [dmitriyminer](https://github.com/dmitriyminer))
- **[tls]** Allow PreferServerCipherSuites as a TLS Option ([#6248](https://github.com/traefik/traefik/pull/6248) by [dtomcej](https://github.com/dtomcej))
- **[tracing]** Update APM client. ([#6152](https://github.com/traefik/traefik/pull/6152) by [ldez](https://github.com/ldez))
- **[tracing]** Elastic APM tracer implementation ([#5870](https://github.com/traefik/traefik/pull/5870) by [amine7536](https://github.com/amine7536))
- **[udp,webui]** WebUI: add udp pages ([#6313](https://github.com/traefik/traefik/pull/6313) by [matthieuh](https://github.com/matthieuh))
- **[webui]** Web UI: Polling on tables ([#5909](https://github.com/traefik/traefik/pull/5909) by [matthieuh](https://github.com/matthieuh))
- **[webui]** Proxy API to Traefik in dev mode ([#5980](https://github.com/traefik/traefik/pull/5980) by [sh7dm](https://github.com/sh7dm))
- **[webui]** Web UI: Table infinite scroll ([#5875](https://github.com/traefik/traefik/pull/5875) by [matthieuh](https://github.com/matthieuh))
- **[webui]** Web UI: Take off logic from generic table component ([#5910](https://github.com/traefik/traefik/pull/5910) by [matthieuh](https://github.com/matthieuh))
- **[webui]** Add dark theme for Web UI ([#6036](https://github.com/traefik/traefik/pull/6036) by [sh7dm](https://github.com/sh7dm))
- Update dependencies ([#6359](https://github.com/traefik/traefik/pull/6359) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[etcd,kv]** fix: etcd provider name. ([#6212](https://github.com/traefik/traefik/pull/6212) by [ldez](https://github.com/ldez))
- **[file]** Revert &#34;Allow fsnotify to reload config files on k8s (or symlinks)&#34; ([#6416](https://github.com/traefik/traefik/pull/6416) by [juliens](https://github.com/juliens))
- **[k8s,k8s/ingress]** Improvement of the unique name of the router for Ingress. ([#6325](https://github.com/traefik/traefik/pull/6325) by [ldez](https://github.com/ldez))
- **[kv,redis]** Update valkeyrie to fix the support of Redis. ([#6291](https://github.com/traefik/traefik/pull/6291) by [ldez](https://github.com/ldez))
- **[kv]** fix: KV flaky tests. ([#6300](https://github.com/traefik/traefik/pull/6300) by [ldez](https://github.com/ldez))
- **[server]** fix: use MaxInt32. ([#5845](https://github.com/traefik/traefik/pull/5845) by [ldez](https://github.com/ldez))
- **[tracing]** Disable default APM tracer. ([#6410](https://github.com/traefik/traefik/pull/6410) by [ldez](https://github.com/ldez))
- **[udp]** Add missing generated element for UDP. ([#6309](https://github.com/traefik/traefik/pull/6309) by [ldez](https://github.com/ldez))
- **[udp]** Build all UDP services on an entrypoint ([#6329](https://github.com/traefik/traefik/pull/6329) by [juliens](https://github.com/juliens))

**Documentation:**
- **[k8s,k8s/crd]** Update the k8s CRD documentation ([#6426](https://github.com/traefik/traefik/pull/6426) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[provider]** Update supported providers list. ([#6190](https://github.com/traefik/traefik/pull/6190) by [ldez](https://github.com/ldez))

**Misc:**
- Merge current v2.1 branch into master ([#6429](https://github.com/traefik/traefik/pull/6429) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into master ([#6409](https://github.com/traefik/traefik/pull/6409) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into master ([#6302](https://github.com/traefik/traefik/pull/6302) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into master ([#6216](https://github.com/traefik/traefik/pull/6216) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into master ([#6138](https://github.com/traefik/traefik/pull/6138) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into master ([#6004](https://github.com/traefik/traefik/pull/6004) by [ldez](https://github.com/ldez))
- Merge current v2.1 branch into master ([#5933](https://github.com/traefik/traefik/pull/5933) by [ldez](https://github.com/ldez))

## [v2.1.6](https://github.com/traefik/traefik/tree/v2.1.6) (2020-02-28)
[All Commits](https://github.com/traefik/traefik/compare/v2.1.4...v2.1.6)

**Bug fixes:**
- **[acme]** Update go-acme/lego to v3.4.0 ([#6376](https://github.com/traefik/traefik/pull/6376) by [ldez](https://github.com/ldez))
- **[api]** Return an error when ping is not enabled. ([#6304](https://github.com/traefik/traefik/pull/6304) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Early filter of the catalog services. ([#6307](https://github.com/traefik/traefik/pull/6307) by [ldez](https://github.com/ldez))
- **[consulcatalog]** fix: consul-catalog uses port from label instead of item port. ([#6345](https://github.com/traefik/traefik/pull/6345) by [ldez](https://github.com/ldez))
- **[file]** fix: YML example of template for the file provider. ([#6402](https://github.com/traefik/traefik/pull/6402) by [ldez](https://github.com/ldez))
- **[file]** Allow fsnotify to reload config files on k8s (or symlinks) ([#5037](https://github.com/traefik/traefik/pull/5037) by [dtomcej](https://github.com/dtomcej))
- **[healthcheck]** Launch healthcheck only one time instead of two ([#6372](https://github.com/traefik/traefik/pull/6372) by [juliens](https://github.com/juliens))
- **[k8s,k8s/crd,k8s/ingress]** Fix secret informer load ([#6364](https://github.com/traefik/traefik/pull/6364) by [mmatur](https://github.com/mmatur))
- **[k8s,k8s/crd]** Use consistent protocol determination ([#6365](https://github.com/traefik/traefik/pull/6365) by [dtomcej](https://github.com/dtomcej))
- **[k8s,k8s/crd]** fix: use the right error in the log ([#6311](https://github.com/traefik/traefik/pull/6311) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[provider]** Don&#39;t throw away valid configuration updates ([#5952](https://github.com/traefik/traefik/pull/5952) by [zaphod42](https://github.com/zaphod42))
- **[tls]** Consider SSLv2 as TLS in order to close the handshake correctly ([#6371](https://github.com/traefik/traefik/pull/6371) by [juliens](https://github.com/juliens))
- **[tracing]** Fix docs and code to match in haystack tracing. ([#6352](https://github.com/traefik/traefik/pull/6352) by [evanlurvey](https://github.com/evanlurvey))

**Documentation:**
- **[acme]** Improve documentation. ([#6324](https://github.com/traefik/traefik/pull/6324) by [ldez](https://github.com/ldez))
- **[file]** Add information about filename and directory options. ([#6333](https://github.com/traefik/traefik/pull/6333) by [ldez](https://github.com/ldez))
- **[k8s,k8s/ingress]** Docs: Clarifying format of ingress endpoint service name ([#6306](https://github.com/traefik/traefik/pull/6306) by [BretFisher](https://github.com/BretFisher))
- **[k8s/crd]** fix: dashboard example with k8s CRD. ([#6330](https://github.com/traefik/traefik/pull/6330) by [ldez](https://github.com/ldez))
- **[middleware,k8s]** Fix formatting in &#34;Kubernetes Namespace&#34; block ([#6305](https://github.com/traefik/traefik/pull/6305) by [berekuk](https://github.com/berekuk))
- **[tls]** Remove TLS cipher suites for TLS minVersion 1.3 ([#6328](https://github.com/traefik/traefik/pull/6328) by [rYR79435](https://github.com/rYR79435))
- **[tls]** Fix typo in the godoc of TLS option MaxVersion ([#6347](https://github.com/traefik/traefik/pull/6347) by [pschaub](https://github.com/pschaub))
- Use explicitly the word Kubernetes in the migration guide. ([#6380](https://github.com/traefik/traefik/pull/6380) by [ldez](https://github.com/ldez))
- Minor readme improvements ([#6293](https://github.com/traefik/traefik/pull/6293) by [Rowayda-Khayri](https://github.com/Rowayda-Khayri))
- Added link to community forum ([#6283](https://github.com/traefik/traefik/pull/6283) by [isaacnewtonfx](https://github.com/isaacnewtonfx))

## [v2.1.5](https://github.com/traefik/traefik/tree/v2.1.5) (2020-02-28)

Skipped.

## [v2.1.4](https://github.com/traefik/traefik/tree/v2.1.4) (2020-02-06)
[All Commits](https://github.com/traefik/traefik/compare/v2.1.3...v2.1.4)

**Bug fixes:**
- **[acme,logs]** Improvement of the certificates resolvers logs ([#6225](https://github.com/traefik/traefik/pull/6225) by [ldez](https://github.com/ldez))
- **[acme]** Fix kubernetes providers shutdown and clean safe.Pool ([#6244](https://github.com/traefik/traefik/pull/6244) by [juliens](https://github.com/juliens))
- **[authentication,middleware]** don&#39;t create http client for each request in forwardAuth middleware ([#6267](https://github.com/traefik/traefik/pull/6267) by [juliens](https://github.com/juliens))
- **[k8s,k8s/ingress]** Allow wildcard hosts in ingress provider ([#6251](https://github.com/traefik/traefik/pull/6251) by [dtomcej](https://github.com/dtomcej))
- **[logs,tls]** Properly purge default certificate from stores before logging ([#6281](https://github.com/traefik/traefik/pull/6281) by [dtomcej](https://github.com/dtomcej))
- **[middleware]** use provider-qualified name when recursing for chain ([#6233](https://github.com/traefik/traefik/pull/6233) by [mpl](https://github.com/mpl))

**Documentation:**
- **[acme,cli]** Documentation fix for acme.md CLI ([#6262](https://github.com/traefik/traefik/pull/6262) by [altano](https://github.com/altano))
- **[acme,k8s/crd]** Add missing certResolver in IngressRoute examples. ([#6265](https://github.com/traefik/traefik/pull/6265) by [ldez](https://github.com/ldez))
- **[k8s]** fix a typo ([#6279](https://github.com/traefik/traefik/pull/6279) by [silenceshell](https://github.com/silenceshell))
- **[middleware]** Minor documentation tweaks. ([#6218](https://github.com/traefik/traefik/pull/6218) by [stevegroom](https://github.com/stevegroom))
- Correct a trivial spelling mistake in the documentation. ([#6269](https://github.com/traefik/traefik/pull/6269) by [nepella](https://github.com/nepella))
- Update install-traefik.md ([#6260](https://github.com/traefik/traefik/pull/6260) by [bitfactory-sander-lissenburg](https://github.com/bitfactory-sander-lissenburg))
- doc: use the same entry point name everywhere ([#6219](https://github.com/traefik/traefik/pull/6219) by [ldez](https://github.com/ldez))
- readme: update links to use HTTPS ([#6274](https://github.com/traefik/traefik/pull/6274) by [imba-tjd](https://github.com/imba-tjd))

## [v2.1.3](https://github.com/traefik/traefik/tree/v2.1.3) (2020-01-21)
[All Commits](https://github.com/traefik/traefik/compare/v2.1.2...v2.1.3)

**Bug fixes:**
- **[acme]** Update go-acme/lego to v3.3.0 ([#6192](https://github.com/traefik/traefik/pull/6192) by [shilch](https://github.com/shilch))
- **[docker]** Use the calculated port when useBindPortIP is enabled ([#6199](https://github.com/traefik/traefik/pull/6199) by [juliens](https://github.com/juliens))
- **[docker]** fix: invalid service definition. ([#6198](https://github.com/traefik/traefik/pull/6198) by [ldez](https://github.com/ldez))
- **[server]** Remove Content-Type auto-detection ([#6097](https://github.com/traefik/traefik/pull/6097) by [juliens](https://github.com/juliens))
- **[service]** fix memleak in safe.Pool ([#6140](https://github.com/traefik/traefik/pull/6140) by [mpl](https://github.com/mpl))

**Documentation:**
- **[docker]** Fix typo in docker routing documentation ([#6147](https://github.com/traefik/traefik/pull/6147) by [tvrg](https://github.com/tvrg))
- **[k8s]** Fixed typo in k8s doc ([#6163](https://github.com/traefik/traefik/pull/6163) by [MyIgel](https://github.com/MyIgel))
- **[marathon]** Fix typo in Marathon doc. ([#6150](https://github.com/traefik/traefik/pull/6150) by [thatshubham](https://github.com/thatshubham))
- **[middleware]** Adding an explanation how to use `htpasswd` for k8s secret ([#6194](https://github.com/traefik/traefik/pull/6194) by [jamct](https://github.com/jamct))
- doc: adds an explanation of the global redirection pattern. ([#6195](https://github.com/traefik/traefik/pull/6195) by [ldez](https://github.com/ldez))
- Fix small typo in user-guides documentation ([#6154](https://github.com/traefik/traefik/pull/6154) by [evert-arias](https://github.com/evert-arias))

## [v2.1.2](https://github.com/traefik/traefik/tree/v2.1.2) (2020-01-07)
[All Commits](https://github.com/traefik/traefik/compare/v2.1.1...v2.1.2)

**Bug fixes:**
- **[authentication,middleware,tracing]** fix(tracing): makes sure tracing headers are being propagated when using forwardAuth ([#6072](https://github.com/traefik/traefik/pull/6072) by [jcchavezs](https://github.com/jcchavezs))
- **[cli]** fix: invalid label/flag parsing. ([#6028](https://github.com/traefik/traefik/pull/6028) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Query consul catalog for service health separately ([#6046](https://github.com/traefik/traefik/pull/6046) by [SantoDE](https://github.com/SantoDE))
- **[k8s,k8s/crd]** Restore ExternalName https support for Kubernetes CRD ([#6037](https://github.com/traefik/traefik/pull/6037) by [kpeiruza](https://github.com/kpeiruza))
- **[k8s,k8s/crd]** Log the ignored namespace only when needed ([#6087](https://github.com/traefik/traefik/pull/6087) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/ingress]** k8s Ingress: fix crash on rules with nil http ([#6121](https://github.com/traefik/traefik/pull/6121) by [grimmy](https://github.com/grimmy))
- **[logs]** Improves error message when a configuration file is empty. ([#6135](https://github.com/traefik/traefik/pull/6135) by [ldez](https://github.com/ldez))
- **[server]** Handle respondingTimeout and better shutdown tests. ([#6115](https://github.com/traefik/traefik/pull/6115) by [juliens](https://github.com/juliens))
- **[server]** Don&#39;t set user-agent to Go-http-client/1.1 ([#6030](https://github.com/traefik/traefik/pull/6030) by [sh7dm](https://github.com/sh7dm))
- **[tracing]** fix: Malformed x-b3-traceid Header ([#6079](https://github.com/traefik/traefik/pull/6079) by [ldez](https://github.com/ldez))
- **[webui]** fix: dashboard redirect loop ([#6078](https://github.com/traefik/traefik/pull/6078) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme]** Use consistent name in ACME documentation ([#6019](https://github.com/traefik/traefik/pull/6019) by [ldez](https://github.com/ldez))
- **[api,k8s/crd]** Add a documentation example for dashboard and api for kubernetes CRD ([#6022](https://github.com/traefik/traefik/pull/6022) by [dduportal](https://github.com/dduportal))
- **[cli]** Fix examples for the use of websecure via CLI ([#6116](https://github.com/traefik/traefik/pull/6116) by [tiagoboeing](https://github.com/tiagoboeing))
- **[k8s,k8s/crd]** Improve documentation about Kubernetes IngressRoute ([#6058](https://github.com/traefik/traefik/pull/6058) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware]** Improve sourceRange explanation for ipWhiteList ([#6070](https://github.com/traefik/traefik/pull/6070) by [der-domi](https://github.com/der-domi))

## [v2.1.1](https://github.com/traefik/traefik/tree/v2.1.1) (2019-12-12)
[All Commits](https://github.com/traefik/traefik/compare/v2.1.0...v2.1.1)

**Bug fixes:**
- **[logs,middleware,metrics]** CloseNotifier: return pointer instead of value ([#6010](https://github.com/traefik/traefik/pull/6010) by [mpl](https://github.com/mpl))

**Documentation:**
- Add Migration Guide for Traefik v2.1 ([#6017](https://github.com/traefik/traefik/pull/6017) by [SantoDE](https://github.com/SantoDE))

## [v2.1.0](https://github.com/traefik/traefik/tree/v2.1.0) (2019-12-10)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-rc1...v2.1.0)

**Enhancements:**
- **[consulcatalog]** Add consul catalog options: requireConsistent, stale, cache ([#5752](https://github.com/traefik/traefik/pull/5752) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Add Consul Catalog provider ([#5395](https://github.com/traefik/traefik/pull/5395) by [negasus](https://github.com/negasus))
- **[k8s,k8s/crd,service]** Support for all services kinds (and sticky) in CRD ([#5711](https://github.com/traefik/traefik/pull/5711) by [mpl](https://github.com/mpl))
- **[metrics]** Added configurable prefix for statsd metrics collection ([#5336](https://github.com/traefik/traefik/pull/5336) by [schulterklopfer](https://github.com/schulterklopfer))
- **[middleware]** Conditional compression based on request Content-Type ([#5721](https://github.com/traefik/traefik/pull/5721) by [ldez](https://github.com/ldez))
- **[server]** Add internal provider ([#5815](https://github.com/traefik/traefik/pull/5815) by [ldez](https://github.com/ldez))
- **[tls]** Add support for MaxVersion in tls.Options ([#5650](https://github.com/traefik/traefik/pull/5650) by [kmeekva](https://github.com/kmeekva))
- **[tls]** Add tls option for Elliptic Curve Preferences ([#5466](https://github.com/traefik/traefik/pull/5466) by [ksarink](https://github.com/ksarink))
- **[tracing]** Update jaeger dependencies ([#5637](https://github.com/traefik/traefik/pull/5637) by [mmatur](https://github.com/mmatur))

**Bug fixes:**
- **[api]** fix: debug endpoint when insecure API. ([#5937](https://github.com/traefik/traefik/pull/5937) by [ldez](https://github.com/ldez))
- **[cli]** fix: sub command help ([#5887](https://github.com/traefik/traefik/pull/5887) by [ldez](https://github.com/ldez))
- **[consulcatalog]** fix: consul catalog constraints. ([#5913](https://github.com/traefik/traefik/pull/5913) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Service registered with same id on Consul Catalog ([#5900](https://github.com/traefik/traefik/pull/5900) by [mmatur](https://github.com/mmatur))
- **[consulcatalog]** Fix empty address for registering service without IP ([#5826](https://github.com/traefik/traefik/pull/5826) by [mmatur](https://github.com/mmatur))
- **[logs,middleware,metrics]** detect CloseNotify capability in accesslog and metrics ([#5985](https://github.com/traefik/traefik/pull/5985) by [mpl](https://github.com/mpl))
- **[server]** fix: remove double call to server Close. ([#5960](https://github.com/traefik/traefik/pull/5960) by [ldez](https://github.com/ldez))
- **[webui]** Fix weighted service provider icon ([#5983](https://github.com/traefik/traefik/pull/5983) by [sh7dm](https://github.com/sh7dm))
- **[webui]** Fix http/tcp resources pagination ([#5986](https://github.com/traefik/traefik/pull/5986) by [matthieuh](https://github.com/matthieuh))
- **[webui]** Use valid condition in the service details panel UI ([#5984](https://github.com/traefik/traefik/pull/5984) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[webui]** Web UI: Avoid polling on /api/entrypoints ([#5863](https://github.com/traefik/traefik/pull/5863) by [matthieuh](https://github.com/matthieuh))
- **[webui]** Web UI: Sync toolbar table state with url query params ([#5861](https://github.com/traefik/traefik/pull/5861) by [matthieuh](https://github.com/matthieuh))

**Documentation:**
- **[consulcatalog]** fix: Consul Catalog documentation. ([#5725](https://github.com/traefik/traefik/pull/5725) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Fix consul catalog documentation ([#5661](https://github.com/traefik/traefik/pull/5661) by [mmatur](https://github.com/mmatur))
- Prepare release v2.1.0-rc2 ([#5846](https://github.com/traefik/traefik/pull/5846) by [ldez](https://github.com/ldez))
- Prepare release v2.1.0-rc1 ([#5844](https://github.com/traefik/traefik/pull/5844) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Several documentation fixes ([#5987](https://github.com/traefik/traefik/pull/5987) by [ldez](https://github.com/ldez))
- Prepare release v2.1.0-rc3 ([#5929](https://github.com/traefik/traefik/pull/5929) by [ldez](https://github.com/ldez))

**Misc:**
- **[cli]** Add custom help function to command ([#5923](https://github.com/traefik/traefik/pull/5923) by [Ullaakut](https://github.com/Ullaakut))
- **[server]** fix: use MaxInt32. ([#5845](https://github.com/traefik/traefik/pull/5845) by [ldez](https://github.com/ldez))
- Merge current v2.0 branch into master ([#5841](https://github.com/traefik/traefik/pull/5841) by [ldez](https://github.com/ldez))
- Merge current v2.0 branch into master  ([#5749](https://github.com/traefik/traefik/pull/5749) by [ldez](https://github.com/ldez))
- Merge current v2.0 branch into master  ([#5619](https://github.com/traefik/traefik/pull/5619) by [ldez](https://github.com/ldez))
- Merge current v2.0 branch into master  ([#5464](https://github.com/traefik/traefik/pull/5464) by [ldez](https://github.com/ldez))
- Merge v2.0.0 into master ([#5402](https://github.com/traefik/traefik/pull/5402) by [ldez](https://github.com/ldez))
- Merge v2.0.0-rc3 into master ([#5354](https://github.com/traefik/traefik/pull/5354) by [ldez](https://github.com/ldez))
- Merge v2.0.0-rc1 into master  ([#5253](https://github.com/traefik/traefik/pull/5253) by [ldez](https://github.com/ldez))
- Merge current v2.0 branch into v2.1 ([#5977](https://github.com/traefik/traefik/pull/5977) by [ldez](https://github.com/ldez))
- Merge current v2.0 branch into v2.1 ([#5931](https://github.com/traefik/traefik/pull/5931) by [ldez](https://github.com/ldez))
- Merge current v2.0 branch into v2.1 ([#5928](https://github.com/traefik/traefik/pull/5928) by [ldez](https://github.com/ldez))

## [v2.0.7](https://github.com/traefik/traefik/tree/v2.0.7) (2019-12-09)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.6...v2.0.7)

**Bug fixes:**
- **[logs,middleware]** Remove mirroring impact in accesslog ([#5967](https://github.com/traefik/traefik/pull/5967) by [juliens](https://github.com/juliens))
- **[middleware]** fix: PassClientTLSCert middleware separators and formatting ([#5921](https://github.com/traefik/traefik/pull/5921) by [ldez](https://github.com/ldez))
- **[server]** Do not stop to listen on tcp listeners on temporary errors  ([#5935](https://github.com/traefik/traefik/pull/5935) by [skwair](https://github.com/skwair))

**Documentation:**
- **[acme,k8s/crd,k8s/ingress]** Document LE caveats with Kubernetes on v2 ([#5902](https://github.com/traefik/traefik/pull/5902) by [dtomcej](https://github.com/dtomcej))
- **[acme]** The Cloudflare hint for the GLOBAL API KEY for CF MAIL/API_KEY ([#5964](https://github.com/traefik/traefik/pull/5964) by [EugenMayer](https://github.com/EugenMayer))
- **[acme]** Improve documentation for ACME/Let&#39;s Encrypt ([#5819](https://github.com/traefik/traefik/pull/5819) by [dduportal](https://github.com/dduportal))
- **[file]** Improve documentation on file provider limitations with file system notifications ([#5939](https://github.com/traefik/traefik/pull/5939) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Make trailing slash more prominent for the &#34;secure dashboard setup&#34; too ([#5963](https://github.com/traefik/traefik/pull/5963) by [EugenMayer](https://github.com/EugenMayer))
- Fix Docker example in &#34;Strip and Rewrite Path Prefixes&#34; in migration guide ([#5949](https://github.com/traefik/traefik/pull/5949) by [q210](https://github.com/q210))
- readme: Fix link to file backend/provider documentation ([#5945](https://github.com/traefik/traefik/pull/5945) by [hartwork](https://github.com/hartwork))

## [v2.1.0-rc3](https://github.com/traefik/traefik/tree/v2.1.0-rc3) (2019-12-02)
[All Commits](https://github.com/traefik/traefik/compare/v2.1.0-rc2...v2.1.0-rc3)

**Bug fixes:**
- **[cli]** fix: sub command help ([#5887](https://github.com/traefik/traefik/pull/5887) by [ldez](https://github.com/ldez))
- **[consulcatalog]** fix: consul catalog constraints. ([#5913](https://github.com/traefik/traefik/pull/5913) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Service registered with same id on Consul Catalog ([#5900](https://github.com/traefik/traefik/pull/5900) by [mmatur](https://github.com/mmatur))
- **[webui]** Web UI: Avoid polling on /api/entrypoints ([#5863](https://github.com/traefik/traefik/pull/5863) by [matthieuh](https://github.com/matthieuh))
- **[webui]** Web UI: Sync toolbar table state with url query params ([#5861](https://github.com/traefik/traefik/pull/5861) by [matthieuh](https://github.com/matthieuh))

**Misc:**
- **[cli]** Add custom help function to command ([#5923](https://github.com/traefik/traefik/pull/5923) by [Ullaakut](https://github.com/Ullaakut))

## [v2.0.6](https://github.com/traefik/traefik/tree/v2.0.6) (2019-12-02)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.5...v2.0.6)

**Bug fixes:**
- **[acme]** Update go-acme/lego to 3.2.0 ([#5839](https://github.com/traefik/traefik/pull/5839) by [kolaente](https://github.com/kolaente))
- **[cli,healthcheck]** Uses, if it exists, the ping entry point provided in the static configuration ([#5867](https://github.com/traefik/traefik/pull/5867) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[healthcheck]** Healthcheck managed for all related services ([#5860](https://github.com/traefik/traefik/pull/5860) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[logs,middleware]** Do not give responsewriter or its headers to asynchronous logging goroutine ([#5840](https://github.com/traefik/traefik/pull/5840) by [mpl](https://github.com/mpl))
- **[middleware]** X-Forwarded-Proto must not skip the redirection. ([#5836](https://github.com/traefik/traefik/pull/5836) by [ldez](https://github.com/ldez))
- **[middleware]** fix: location header rewrite. ([#5835](https://github.com/traefik/traefik/pull/5835) by [ldez](https://github.com/ldez))
- **[middleware]** Remove Request Headers CORS Preflight Requirement ([#5903](https://github.com/traefik/traefik/pull/5903) by [dtomcej](https://github.com/dtomcej))
- **[rancher]** Change service name in rancher provider to make webui service details view work ([#5895](https://github.com/traefik/traefik/pull/5895) by [SantoDE](https://github.com/SantoDE))
- **[tracing]** Fix extraction for zipkin tracing ([#5920](https://github.com/traefik/traefik/pull/5920) by [jcchavezs](https://github.com/jcchavezs))
- **[webui]** Web UI: Avoid unnecessary duplicated api calls ([#5884](https://github.com/traefik/traefik/pull/5884) by [matthieuh](https://github.com/matthieuh))
- **[webui]** Web UI: Avoid some router properties to overflow their container ([#5872](https://github.com/traefik/traefik/pull/5872) by [matthieuh](https://github.com/matthieuh))
- **[webui]** Web UI: Fix displayed tcp service details ([#5868](https://github.com/traefik/traefik/pull/5868) by [matthieuh](https://github.com/matthieuh))

**Documentation:**
- **[acme]** doc: fix wrong acme information ([#5837](https://github.com/traefik/traefik/pull/5837) by [ldez](https://github.com/ldez))
- **[docker,docker/swarm]** Add Swarm section to the Docker Provider Documentation ([#5874](https://github.com/traefik/traefik/pull/5874) by [dduportal](https://github.com/dduportal))
- **[docker]** Update router entrypoint example ([#5766](https://github.com/traefik/traefik/pull/5766) by [woto](https://github.com/woto))
- **[k8s/helm]** Mention the experimental Helm Chart in the installation section of documentation ([#5879](https://github.com/traefik/traefik/pull/5879) by [dduportal](https://github.com/dduportal))
- doc: remove double quotes on CLI flags. ([#5862](https://github.com/traefik/traefik/pull/5862) by [ldez](https://github.com/ldez))
- Fixed spelling error ([#5834](https://github.com/traefik/traefik/pull/5834) by [blakebuthod](https://github.com/blakebuthod))
- Add back the security section from v1 ([#5832](https://github.com/traefik/traefik/pull/5832) by [pascalandy](https://github.com/pascalandy))

## [v2.1.0-rc2](https://github.com/traefik/traefik/tree/v2.0.4) (2019-11-15)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-rc1...v2.1.0-rc2)

Fixes int overflow.
Same changelog as v2.1.0-rc1

## [v2.1.0-rc1](https://github.com/traefik/traefik/tree/v2.1.0-rc1) (2019-11-15)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-rc1...v2.1.0-rc1)

**Enhancements:**
- **[consulcatalog]** Add consul catalog options: requireConsistent, stale, cache ([#5752](https://github.com/traefik/traefik/pull/5752) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Add Consul Catalog provider ([#5395](https://github.com/traefik/traefik/pull/5395) by [negasus](https://github.com/negasus))
- **[k8s,k8s/crd,service]** Support for all services kinds (and sticky) in CRD ([#5711](https://github.com/traefik/traefik/pull/5711) by [mpl](https://github.com/mpl))
- **[metrics]** Added configurable prefix for statsd metrics collection ([#5336](https://github.com/traefik/traefik/pull/5336) by [schulterklopfer](https://github.com/schulterklopfer))
- **[middleware]** Conditional compression based on request Content-Type ([#5721](https://github.com/traefik/traefik/pull/5721) by [ldez](https://github.com/ldez))
- **[server]** Add internal provider ([#5815](https://github.com/traefik/traefik/pull/5815) by [ldez](https://github.com/ldez))
- **[tls]** Add support for MaxVersion in tls.Options ([#5650](https://github.com/traefik/traefik/pull/5650) by [kmeekva](https://github.com/kmeekva))
- **[tls]** Add tls option for Elliptic Curve Preferences ([#5466](https://github.com/traefik/traefik/pull/5466) by [ksarink](https://github.com/ksarink))
- **[tracing]** Update jaeger dependencies ([#5637](https://github.com/traefik/traefik/pull/5637) by [mmatur](https://github.com/mmatur))

**Bug fixes:**
- **[consulcatalog]** Fix empty address for registering service without IP ([#5826](https://github.com/traefik/traefik/pull/5826) by [mmatur](https://github.com/mmatur))

**Documentation:**
- **[consulcatalog]** fix: Consul Catalog documentation. ([#5725](https://github.com/traefik/traefik/pull/5725) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Fix consul catalog documentation ([#5661](https://github.com/traefik/traefik/pull/5661) by [mmatur](https://github.com/mmatur))

**Misc:**
- Merge current v2.0 branch into master  ([#5749](https://github.com/traefik/traefik/pull/5749) by [ldez](https://github.com/ldez))
- Merge current v2.0 branch into master  ([#5619](https://github.com/traefik/traefik/pull/5619) by [ldez](https://github.com/ldez))
- Merge current v2.0 branch into master  ([#5464](https://github.com/traefik/traefik/pull/5464) by [ldez](https://github.com/ldez))
- Merge v2.0.0 into master ([#5402](https://github.com/traefik/traefik/pull/5402) by [ldez](https://github.com/ldez))
- Merge v2.0.0-rc3 into master ([#5354](https://github.com/traefik/traefik/pull/5354) by [ldez](https://github.com/ldez))
- Merge v2.0.0-rc1 into master  ([#5253](https://github.com/traefik/traefik/pull/5253) by [ldez](https://github.com/ldez))

## [v2.0.5](https://github.com/traefik/traefik/tree/v2.0.5) (2019-11-14)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.4...v2.0.5)

**Bug fixes:**
- **[metrics]** fix: metric with services LB. ([#5759](https://github.com/traefik/traefik/pull/5759) by [ldez](https://github.com/ldez))
- **[middleware]** fix: stripPrefix middleware with empty resulting path. ([#5806](https://github.com/traefik/traefik/pull/5806) by [ldez](https://github.com/ldez))
- **[middleware]** Fix rate limiting and SSE ([#5737](https://github.com/traefik/traefik/pull/5737) by [sylr](https://github.com/sylr))
- **[tracing]** Upgrades zipkin library to avoid errors when using textMap. ([#5754](https://github.com/traefik/traefik/pull/5754) by [jcchavezs](https://github.com/jcchavezs))

**Documentation:**
- **[acme,cluster]** Update ACME storage docs to remove reference to KV store in CE ([#5433](https://github.com/traefik/traefik/pull/5433) by [bradjones1](https://github.com/bradjones1))
- **[api]** docs: remove field api.entryPoint ([#5776](https://github.com/traefik/traefik/pull/5776) by [waitingsong](https://github.com/waitingsong))
- **[api]** Adds missed quotes in api.md ([#5787](https://github.com/traefik/traefik/pull/5787) by [woto](https://github.com/woto))
- **[docker/swarm]** Dashboard example with swarm ([#5795](https://github.com/traefik/traefik/pull/5795) by [dduportal](https://github.com/dduportal))
- **[docker]** Fix error in link description for priority ([#5746](https://github.com/traefik/traefik/pull/5746) by [ASDFGamer](https://github.com/ASDFGamer))
- **[k8s]** Wrong endpoint on the TLS secret example ([#5817](https://github.com/traefik/traefik/pull/5817) by [yacinelazaar](https://github.com/yacinelazaar))
- **[middleware,docker]** Double dollar on docker-compose config ([#5775](https://github.com/traefik/traefik/pull/5775) by [clery](https://github.com/clery))
- Fix quickstart link in README ([#5794](https://github.com/traefik/traefik/pull/5794) by [mcky](https://github.com/mcky))
- fix typo in v1 to v2 migration guide ([#5820](https://github.com/traefik/traefik/pull/5820) by [fschl](https://github.com/fschl))
- slashes ended up in bad place. ([#5798](https://github.com/traefik/traefik/pull/5798) by [icepic](https://github.com/icepic))

## [v2.0.4](https://github.com/traefik/traefik/tree/v2.0.4) (2019-10-28)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.3...v2.0.4)

Fixes releases system.
Same changelog as v2.0.3.

## [v2.0.3](https://github.com/traefik/traefik/tree/v2.0.3) (2019-10-28)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.2...v2.0.3)

**Bug fixes:**
- **[acme,logs]** Use debug for log about skipping addition of cert ([#5641](https://github.com/traefik/traefik/pull/5641) by [sylr](https://github.com/sylr))
- **[file]** fix: add filename in the file provider logs. ([#5636](https://github.com/traefik/traefik/pull/5636) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd,k8s/ingress]** Remove unnecessary reload of the configuration. ([#5707](https://github.com/traefik/traefik/pull/5707) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd,k8s/ingress]** Fixing support for HTTPs backends with Kubernetes ExternalName services ([#5660](https://github.com/traefik/traefik/pull/5660) by [kpeiruza](https://github.com/kpeiruza))
- **[k8s,k8s/ingress]** Normalize service and router names for ingress. ([#5623](https://github.com/traefik/traefik/pull/5623) by [ldez](https://github.com/ldez))
- **[logs]** Set proxy protocol logger to DEBUG level ([#5712](https://github.com/traefik/traefik/pull/5712) by [mmatur](https://github.com/mmatur))
- **[middleware]** fix: add stacktrace when recover. ([#5654](https://github.com/traefik/traefik/pull/5654) by [ldez](https://github.com/ldez))
- **[tracing]** Let instana/go-sensor handle default agent host ([#5658](https://github.com/traefik/traefik/pull/5658) by [sylr](https://github.com/sylr))
- **[tracing]** fix: default tracing backend. ([#5717](https://github.com/traefik/traefik/pull/5717) by [ldez](https://github.com/ldez))
- fix: deep copy of passHostHeader on ServersLoadBalancer. ([#5720](https://github.com/traefik/traefik/pull/5720) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme]** Fix acme storage file docker mounting example ([#5633](https://github.com/traefik/traefik/pull/5633) by [jansauer](https://github.com/jansauer))
- **[acme]** fix incorrect DNS reference ([#5666](https://github.com/traefik/traefik/pull/5666) by [oskapt](https://github.com/oskapt))
- **[logs]** Clarify unit of duration field in access log ([#5664](https://github.com/traefik/traefik/pull/5664) by [Sarke](https://github.com/Sarke))
- **[middleware]** Fix Security Headers Doc ([#5706](https://github.com/traefik/traefik/pull/5706) by [FlorianPerrot](https://github.com/FlorianPerrot))
- **[middleware]** Migration guide: pathprefixstrip migration ([#5600](https://github.com/traefik/traefik/pull/5600) by [dduportal](https://github.com/dduportal))
- **[middleware]** fix ForwardAuth tls.skipverify examples ([#5683](https://github.com/traefik/traefik/pull/5683) by [remche](https://github.com/remche))
- **[rules]** Add documentation about backtick for rule definition. ([#5714](https://github.com/traefik/traefik/pull/5714) by [ldez](https://github.com/ldez))
- **[webui]** Improve documentation of the router rules for API and dashboard ([#5625](https://github.com/traefik/traefik/pull/5625) by [dduportal](https://github.com/dduportal))
- doc: @ is not authorized in names definition. ([#5734](https://github.com/traefik/traefik/pull/5734) by [ldez](https://github.com/ldez))
- Remove obsolete v2 remark from README ([#5669](https://github.com/traefik/traefik/pull/5669) by [dragetd](https://github.com/dragetd))
- Fix spelling mistake: &#34;founded&#34; -&gt; &#34;found&#34; ([#5674](https://github.com/traefik/traefik/pull/5674) by [ocanty](https://github.com/ocanty))
- fix typo for stripPrefix in tab File (YAML) ([#5694](https://github.com/traefik/traefik/pull/5694) by [nalakawula](https://github.com/nalakawula))
- Add example for changing the port used by traefik to connect to a service ([#5224](https://github.com/traefik/traefik/pull/5224) by [robertbaker](https://github.com/robertbaker))

**Misc:**
- **[logs,middleware]** Cherry pick v1.7 into v2.0 ([#5735](https://github.com/traefik/traefik/pull/5735) by [jbdoumenjou](https://github.com/jbdoumenjou))

## [v2.0.2](https://github.com/traefik/traefik/tree/v2.0.2) (2019-10-09)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.1...v2.0.2)

**Bug fixes:**
- **[acme]** fix: ovh client int overflow. ([#5607](https://github.com/traefik/traefik/pull/5607) by [ldez](https://github.com/ldez))
- **[api,k8s,k8s/ingress]** fix: default router name for k8s ingress. ([#5612](https://github.com/traefik/traefik/pull/5612) by [ldez](https://github.com/ldez))
- **[file]** fix: default passHostHeader for file provider. ([#5516](https://github.com/traefik/traefik/pull/5516) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd]** Fix typo in log ([#5590](https://github.com/traefik/traefik/pull/5590) by [XciD](https://github.com/XciD))
- **[middleware,metrics]** fix: panic with metrics recorder. ([#5536](https://github.com/traefik/traefik/pull/5536) by [ldez](https://github.com/ldez))
- **[webui]** Add a service sticky details vue component  ([#5579](https://github.com/traefik/traefik/pull/5579) by [jbdoumenjou](https://github.com/jbdoumenjou))
- fix: return an error instead of panic. ([#5549](https://github.com/traefik/traefik/pull/5549) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme,file]** Fix yaml domains example ([#5569](https://github.com/traefik/traefik/pull/5569) by [SuperSandro2000](https://github.com/SuperSandro2000))
- **[api,webui]** Clarifies how to configure and access the dashboard in the api &amp; dashboard documentations ([#5523](https://github.com/traefik/traefik/pull/5523) by [dduportal](https://github.com/dduportal))
- **[api]** Add overview to API documentation ([#5539](https://github.com/traefik/traefik/pull/5539) by [lnxbil](https://github.com/lnxbil))
- **[cli]** typo in cli command ([#5586](https://github.com/traefik/traefik/pull/5586) by [basraven](https://github.com/basraven))
- **[cli]** Replace ambiguous cli help message wording ([#5233](https://github.com/traefik/traefik/pull/5233) by [jansauer](https://github.com/jansauer))
- **[docker]** Fixed typo in routing/providers/docker documentation ([#5520](https://github.com/traefik/traefik/pull/5520) by [lyrixx](https://github.com/lyrixx))
- **[docker]** $ needs escaping in docker-compose.yml ([#5528](https://github.com/traefik/traefik/pull/5528) by [lnxbil](https://github.com/lnxbil))
- **[file]** State clearly, that they are mutual exclusive ([#5527](https://github.com/traefik/traefik/pull/5527) by [lnxbil](https://github.com/lnxbil))
- **[healthcheck]** fix: typo in healthCheck examples ([#5575](https://github.com/traefik/traefik/pull/5575) by [serpi90](https://github.com/serpi90))
- **[k8s/crd]** Update 04-ingressroutes.yml ([#5585](https://github.com/traefik/traefik/pull/5585) by [basraven](https://github.com/basraven))
- **[k8s/crd]** Update apiVersion in documentation descriptor ([#5605](https://github.com/traefik/traefik/pull/5605) by [pyaillet](https://github.com/pyaillet))
- **[metrics]** doc: fix influxDB and statsD case in configuration page. ([#5531](https://github.com/traefik/traefik/pull/5531) by [ldez](https://github.com/ldez))
- **[middleware]** Update scope of services and middlewares ([#5584](https://github.com/traefik/traefik/pull/5584) by [Thoorium](https://github.com/Thoorium))
- **[middleware]** Typo in documentation ([#5558](https://github.com/traefik/traefik/pull/5558) by [Constans](https://github.com/Constans))
- **[middleware]** Fix misleading text ([#5540](https://github.com/traefik/traefik/pull/5540) by [joassouza](https://github.com/joassouza))
- **[tls]** document serversTransport ([#5529](https://github.com/traefik/traefik/pull/5529) by [mpl](https://github.com/mpl))
- **[tls]** TLS_RSA_WITH_AES_256_GCM_SHA384 is considered weak ([#5578](https://github.com/traefik/traefik/pull/5578) by [Constans](https://github.com/Constans))
- **[tls]** Improve ciphersuite examples ([#5594](https://github.com/traefik/traefik/pull/5594) by [Constans](https://github.com/Constans))
- Remove deprecated videos ([#5570](https://github.com/traefik/traefik/pull/5570) by [emilevauge](https://github.com/emilevauge))
- fix: remove extra backtick from routers docs ([#5572](https://github.com/traefik/traefik/pull/5572) by [serpi90](https://github.com/serpi90))
- document providersThrottleDuration ([#5519](https://github.com/traefik/traefik/pull/5519) by [mpl](https://github.com/mpl))
- Add a response forwarding section to the service documentation  ([#5517](https://github.com/traefik/traefik/pull/5517) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Change instances of &#34;dynamic&#34; to &#34;dynamic&#34; ([#5504](https://github.com/traefik/traefik/pull/5504) by [dat-gitto-kid](https://github.com/dat-gitto-kid))
- Add the pass host header section to the services documentation ([#5500](https://github.com/traefik/traefik/pull/5500) by [jbdoumenjou](https://github.com/jbdoumenjou))
- fix misspelling on documentation landing page ([#5613](https://github.com/traefik/traefik/pull/5613) by [cthompson527](https://github.com/cthompson527))

## [v2.0.1](https://github.com/traefik/traefik/tree/v2.0.1) (2019-09-26)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0...v2.0.1)

**Bug fixes:**
- **[go,security]** This version is compiled with [Go 1.13.1](https://groups.google.com/d/msg/golang-announce/cszieYyuL9Q/g4Z7pKaqAgAJ), which fixes a vulnerability in previous versions. See the [CVE](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-16276) about it for more details.
- **[api,healthcheck]** Return an actual server status updater ([#5407](https://github.com/traefik/traefik/pull/5407) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[cli]** Flag names don&#39;t need a consistent case. ([#5438](https://github.com/traefik/traefik/pull/5438) by [ldez](https://github.com/ldez))
- **[docker]** fix: docker service name. ([#5491](https://github.com/traefik/traefik/pull/5491) by [ldez](https://github.com/ldez))
- **[logs,middleware]** fix: improve log for invalid middleware. ([#5486](https://github.com/traefik/traefik/pull/5486) by [ldez](https://github.com/ldez))
- **[middleware]** Update Casing on STS Header Directive ([#5492](https://github.com/traefik/traefik/pull/5492) by [dtomcej](https://github.com/dtomcej))
- **[server]** Do not initialize list of middlewares if not needed ([#5485](https://github.com/traefik/traefik/pull/5485) by [mpl](https://github.com/mpl))
- **[websocket]** Fix case-sensitive header in websocket ([#5397](https://github.com/traefik/traefik/pull/5397) by [juliens](https://github.com/juliens))

**Documentation:**
- **[acme,tls]** Improve TLS documentation. ([#5448](https://github.com/traefik/traefik/pull/5448) by [ldez](https://github.com/ldez))
- **[acme]** fix typo for kubectl version ([#5409](https://github.com/traefik/traefik/pull/5409) by [mpl](https://github.com/mpl))
- **[acme]** Wrong acme example. ([#5439](https://github.com/traefik/traefik/pull/5439) by [ldez](https://github.com/ldez))
- **[cli,docker]** doc: Flags and labels are case insensitive. ([#5428](https://github.com/traefik/traefik/pull/5428) by [ldez](https://github.com/ldez))
- **[docker,marathon,rancher]** clarify automatic service creation/assignment with labels ([#5493](https://github.com/traefik/traefik/pull/5493) by [mpl](https://github.com/mpl))
- **[file]** fix doc about file.filename ([#5494](https://github.com/traefik/traefik/pull/5494) by [ldez](https://github.com/ldez))
- **[k8s]** add indent to fix notes ([#5467](https://github.com/traefik/traefik/pull/5467) by [mpl](https://github.com/mpl))
- **[middleware,docker,marathon,tls]** Improve documentation for the TLS  section of the provider connection. ([#5437](https://github.com/traefik/traefik/pull/5437) by [ldez](https://github.com/ldez))
- **[yaml]** YAML I love you ([#5461](https://github.com/traefik/traefik/pull/5461) by [mmatur](https://github.com/mmatur))
- Improve routing documentation ([#5450](https://github.com/traefik/traefik/pull/5450) by [ldez](https://github.com/ldez))
- fix: typo in TOML for HTTP to HTTPS redirection ([#5452](https://github.com/traefik/traefik/pull/5452) by [krerkkiat](https://github.com/krerkkiat))
- document that /dashboard should be preferred over / ([#5431](https://github.com/traefik/traefik/pull/5431) by [mpl](https://github.com/mpl))
- Improve the migration guide ([#5430](https://github.com/traefik/traefik/pull/5430) by [jbdoumenjou](https://github.com/jbdoumenjou))
- fixed doc typoes ([#5425](https://github.com/traefik/traefik/pull/5425) by [mpl](https://github.com/mpl))
- fix indentation for tab on migration guide ([#5423](https://github.com/traefik/traefik/pull/5423) by [ViceIce](https://github.com/ViceIce))
- Update links in readme. ([#5411](https://github.com/traefik/traefik/pull/5411) by [ldez](https://github.com/ldez))
- Add the router priority documentation ([#5481](https://github.com/traefik/traefik/pull/5481) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Improve the Migration Guide ([#5391](https://github.com/traefik/traefik/pull/5391) by [jbdoumenjou](https://github.com/jbdoumenjou))

## [v1.7.18](https://github.com/traefik/traefik/tree/v1.7.18) (2019-09-23)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.17...v1.7.18)

**Bug fixes:**
- **[go,security]** This version is compiled with [Go 1.12.10](https://groups.google.com/d/msg/golang-announce/cszieYyuL9Q/g4Z7pKaqAgAJ), which fixes a vulnerability in previous versions. See the [CVE](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-16276) about it for more details.

## [v1.7.17](https://github.com/traefik/traefik/tree/v1.7.17) (2019-09-23)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.16...v1.7.17)

**Bug fixes:**
- **[logs,middleware]** Avoid closing stdout when the accesslog handler is closed ([#5459](https://github.com/traefik/traefik/pull/5459) by [nrwiersma](https://github.com/nrwiersma))
- **[middleware]** Actually send header and code during WriteHeader, if needed ([#5404](https://github.com/traefik/traefik/pull/5404) by [mpl](https://github.com/mpl))

**Documentation:**
- **[k8s]** Add note clarifying client certificate header ([#5362](https://github.com/traefik/traefik/pull/5362) by [bradjones1](https://github.com/bradjones1))
- **[webui]** Update docs links. ([#5412](https://github.com/traefik/traefik/pull/5412) by [ldez](https://github.com/ldez))
- Update Traefik image version. ([#5399](https://github.com/traefik/traefik/pull/5399) by [ldez](https://github.com/ldez))

## [v2.0.0](https://github.com/traefik/traefik/tree/v2.0.0) (2019-09-16)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-alpha1...v2.0.0)

**Enhancements:**
- **[acme,api,tracing]** New API security ([#5311](https://github.com/traefik/traefik/pull/5311) by [juliens](https://github.com/juliens))
- **[acme,k8s,k8s/crd]** Document the TLS with ACME case ([#4654](https://github.com/traefik/traefik/pull/4654) by [mpl](https://github.com/mpl))
- **[acme,kv]** Remove Deprecated StorageFile ([#4252](https://github.com/traefik/traefik/pull/4252) by [juliens](https://github.com/juliens))
- **[acme]** Remove timeout/interval from the ACME Provider ([#4842](https://github.com/traefik/traefik/pull/4842) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[acme]** Certificate resolvers. ([#5116](https://github.com/traefik/traefik/pull/5116) by [ldez](https://github.com/ldez))
- **[acme]** Improve acme logs. ([#5139](https://github.com/traefik/traefik/pull/5139) by [ldez](https://github.com/ldez))
- **[acme]** Migrate to go-acme/lego. ([#4589](https://github.com/traefik/traefik/pull/4589) by [ldez](https://github.com/ldez))
- **[api,provider]** Enhance REST provider ([#5072](https://github.com/traefik/traefik/pull/5072) by [dtomcej](https://github.com/dtomcej))
- **[api]** Adding content-header to api endpoints ([#5019](https://github.com/traefik/traefik/pull/5019) by [dalanmiller](https://github.com/dalanmiller))
- **[api]** Deal with multiple errors and their criticality ([#5070](https://github.com/traefik/traefik/pull/5070) by [mpl](https://github.com/mpl))
- **[api]** API: remove configuration of Entrypoint and Middlewares ([#5119](https://github.com/traefik/traefik/pull/5119) by [mpl](https://github.com/mpl))
- **[api]** Improve API endpoints ([#5080](https://github.com/traefik/traefik/pull/5080) by [ldez](https://github.com/ldez))
- **[api]** API: new contract ([#4964](https://github.com/traefik/traefik/pull/4964) by [mpl](https://github.com/mpl))
- **[api]** Improve API for the web UI ([#5267](https://github.com/traefik/traefik/pull/5267) by [ldez](https://github.com/ldez))
- **[api]** Manage status for TCP element in the endpoint overview. ([#5108](https://github.com/traefik/traefik/pull/5108) by [ldez](https://github.com/ldez))
- **[api]** API: expose runtime representation ([#4841](https://github.com/traefik/traefik/pull/4841) by [mpl](https://github.com/mpl))
- **[authentication,middleware,k8s,k8s/crd]** Auth middlewares in kubernetes CRD use secrets ([#5299](https://github.com/traefik/traefik/pull/5299) by [juliens](https://github.com/juliens))
- **[authentication,logs,etcd]** Remove deprecated elements ([#3715](https://github.com/traefik/traefik/pull/3715) by [geraldcroes](https://github.com/geraldcroes))
- **[authentication,middleware]** Basic Auth custom realm ([#3917](https://github.com/traefik/traefik/pull/3917) by [tcoupin](https://github.com/tcoupin))
- **[cli]** New static configuration loading system. ([#4935](https://github.com/traefik/traefik/pull/4935) by [ldez](https://github.com/ldez))
- **[docker,k8s,k8s/crd,k8s/ingress]** chore: update docker and k8s ([#5174](https://github.com/traefik/traefik/pull/5174) by [ldez](https://github.com/ldez))
- **[docker,k8s,k8s/crd,marathon,rancher,tcp]** Add weighted round robin load balancer on TCP ([#5380](https://github.com/traefik/traefik/pull/5380) by [juliens](https://github.com/juliens))
- **[docker,tcp]** Add support for TCP labels in Docker provider ([#4621](https://github.com/traefik/traefik/pull/4621) by [juliens](https://github.com/juliens))
- **[docker]** Adds default rule system on Docker provider. ([#4413](https://github.com/traefik/traefik/pull/4413) by [ldez](https://github.com/ldez))
- **[docker]** Adds Docker provider support ([#4399](https://github.com/traefik/traefik/pull/4399) by [ldez](https://github.com/ldez))
- **[docker]** Update to Go1.12. Support of TLS1.3 ([#4540](https://github.com/traefik/traefik/pull/4540) by [ldez](https://github.com/ldez))
- **[etcd]** Remove etcd v2 ([#3739](https://github.com/traefik/traefik/pull/3739) by [geraldcroes](https://github.com/geraldcroes))
- **[file]** Restrict traefik.toml to static configuration. ([#5090](https://github.com/traefik/traefik/pull/5090) by [ldez](https://github.com/ldez))
- **[file]** Support YAML for the dynamic configuration. ([#5024](https://github.com/traefik/traefik/pull/5024) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd,k8s/ingress]** Correct Kubernetes Ingress and IngressRoute port heuristic for choosing HTTPS ([#5167](https://github.com/traefik/traefik/pull/5167) by [seh](https://github.com/seh))
- **[k8s,k8s/crd,k8s/ingress]** Fix kubernetes id name ([#5383](https://github.com/traefik/traefik/pull/5383) by [mmatur](https://github.com/mmatur))
- **[k8s,k8s/crd,tcp]** Add support for TCP (in kubernetes CRD) ([#4885](https://github.com/traefik/traefik/pull/4885) by [mpl](https://github.com/mpl))
- **[k8s,k8s/crd,tls]** Define TLS options on the Router configuration for Kubernetes ([#4973](https://github.com/traefik/traefik/pull/4973) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/crd]** Add passHostHeader and responseForwarding in IngressRoute ([#5368](https://github.com/traefik/traefik/pull/5368) by [juliens](https://github.com/juliens))
- **[k8s,k8s/crd]** Add scheme to IngressRoute. ([#5062](https://github.com/traefik/traefik/pull/5062) by [ldez](https://github.com/ldez))
- **[k8s,k8s/ingress]** Renamed `kubernetes` provider in `kubernetesIngress` provider ([#5068](https://github.com/traefik/traefik/pull/5068) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/ingress]** Add TLS-enabled Router ([#5162](https://github.com/traefik/traefik/pull/5162) by [dtomcej](https://github.com/dtomcej))
- **[k8s/ingress]** Adds Kubernetes provider support ([#4476](https://github.com/traefik/traefik/pull/4476) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s/ingress]** Adds update ingress status ([#4603](https://github.com/traefik/traefik/pull/4603) by [juliens](https://github.com/juliens))
- **[k8s/ingress]** k8s integration tests ([#4569](https://github.com/traefik/traefik/pull/4569) by [juliens](https://github.com/juliens))
- **[k8s/ingress]** Custom resource definition ([#4591](https://github.com/traefik/traefik/pull/4591) by [ldez](https://github.com/ldez))
- **[logs]** Improve error on router without service. ([#5126](https://github.com/traefik/traefik/pull/5126) by [ldez](https://github.com/ldez))
- **[logs]** log.loglevel becomes log.level in configuration ([#4775](https://github.com/traefik/traefik/pull/4775) by [juliens](https://github.com/juliens))
- **[logs]** Drop headers by default in access logs. ([#5034](https://github.com/traefik/traefik/pull/5034) by [ldez](https://github.com/ldez))
- **[logs]** Default to CLF when accesslog format is unsupported ([#5314](https://github.com/traefik/traefik/pull/5314) by [mpl](https://github.com/mpl))
- **[marathon,tcp]** Handle TCP in the marathon provider ([#4728](https://github.com/traefik/traefik/pull/4728) by [juliens](https://github.com/juliens))
- **[marathon]** Adds Marathon support. ([#4415](https://github.com/traefik/traefik/pull/4415) by [ldez](https://github.com/ldez))
- **[metrics]** Add Metrics ([#5111](https://github.com/traefik/traefik/pull/5111) by [mmatur](https://github.com/mmatur))
- **[metrics]** Add HTTP authentication to influxdb metric backend ([#3600](https://github.com/traefik/traefik/pull/3600) by [halfa](https://github.com/halfa))
- **[middleware,k8s,k8s/crd]** k8s ErrorPage middleware now uses k8s service ([#5339](https://github.com/traefik/traefik/pull/5339) by [juliens](https://github.com/juliens))
- **[middleware,k8s/crd]** Handle cross-provider middleware in kubernetes CRD ([#5009](https://github.com/traefik/traefik/pull/5009) by [mpl](https://github.com/mpl))
- **[middleware,provider]** Change the provider separator from . to @ ([#4982](https://github.com/traefik/traefik/pull/4982) by [ldez](https://github.com/ldez))
- **[middleware,provider]** Add Feature-Policy header support ([#5156](https://github.com/traefik/traefik/pull/5156) by [dtomcej](https://github.com/dtomcej))
- **[middleware,tracing]** Re enable ratelimit integration tests ([#5288](https://github.com/traefik/traefik/pull/5288) by [mmatur](https://github.com/mmatur))
- **[middleware,provider]** IPStrategy for selecting IP in whitelist ([#3778](https://github.com/traefik/traefik/pull/3778) by [juliens](https://github.com/juliens))
- **[middleware,provider]** Enables the use of elements declared in other providers ([#4372](https://github.com/traefik/traefik/pull/4372) by [geraldcroes](https://github.com/geraldcroes))
- **[middleware]** Migrates the pass client tls cert middleware ([#4373](https://github.com/traefik/traefik/pull/4373) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware]** Migrates Compress from bool to struct ([#3714](https://github.com/traefik/traefik/pull/3714) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware]** Updates for jaeger tracing client. ([#3688](https://github.com/traefik/traefik/pull/3688) by [tcolgate](https://github.com/tcolgate))
- **[middleware]** Add forwarded headers on entry point configuration ([#4364](https://github.com/traefik/traefik/pull/4364) by [juliens](https://github.com/juliens))
- **[middleware]** SchemeRedirect Middleware ([#4400](https://github.com/traefik/traefik/pull/4400) by [geraldcroes](https://github.com/geraldcroes))
- **[middleware]** Add rate limiter, rename maxConn into inFlightReq ([#5246](https://github.com/traefik/traefik/pull/5246) by [mpl](https://github.com/mpl))
- **[middleware]** Disable RateLimit temporarily ([#5123](https://github.com/traefik/traefik/pull/5123) by [juliens](https://github.com/juliens))
- **[middleware]** Enable CORS configuration ([#3809](https://github.com/traefik/traefik/pull/3809) by [dtomcej](https://github.com/dtomcej))
- **[provider]** New constraints management. ([#4965](https://github.com/traefik/traefik/pull/4965) by [ldez](https://github.com/ldez))
- **[provider]** Remove BaseProvider ([#4661](https://github.com/traefik/traefik/pull/4661) by [ldez](https://github.com/ldez))
- **[provider]** Use name@provider instead of provider@name. ([#4990](https://github.com/traefik/traefik/pull/4990) by [ldez](https://github.com/ldez))
- **[provider]** Add health check timeout parameter ([#3813](https://github.com/traefik/traefik/pull/3813) by [jbiel](https://github.com/jbiel))
- **[provider]** Removes deprecated templates ([#3649](https://github.com/traefik/traefik/pull/3649) by [geraldcroes](https://github.com/geraldcroes))
- **[provider]** Remove everything templates related ([#4595](https://github.com/traefik/traefik/pull/4595) by [mpl](https://github.com/mpl))
- **[provider]** Small code enhancements on providers ([#3707](https://github.com/traefik/traefik/pull/3707) by [vdemeester](https://github.com/vdemeester))
- **[provider]** Migrate rest provider ([#4253](https://github.com/traefik/traefik/pull/4253) by [juliens](https://github.com/juliens))
- **[provider]** Labels parser. ([#4236](https://github.com/traefik/traefik/pull/4236) by [ldez](https://github.com/ldez))
- **[rancher]** Add Rancher provider ([#4647](https://github.com/traefik/traefik/pull/4647) by [SantoDE](https://github.com/SantoDE))
- **[rules]** New rule syntax ([#4437](https://github.com/traefik/traefik/pull/4437) by [juliens](https://github.com/juliens))
- **[server]** Adds mirroring service ([#5251](https://github.com/traefik/traefik/pull/5251) by [juliens](https://github.com/juliens))
- **[server]** Add support proxyprotocol v2 ([#4755](https://github.com/traefik/traefik/pull/4755) by [c0va23](https://github.com/c0va23))
- **[server]** WeightedRoundRobin load balancer ([#5237](https://github.com/traefik/traefik/pull/5237) by [juliens](https://github.com/juliens))
- **[server]** Make HTTP Keep-Alive timeout configurable for backend connections ([#4983](https://github.com/traefik/traefik/pull/4983) by [mszabo-wikia](https://github.com/mszabo-wikia))
- **[server]** Rework loadbalancer support ([#4933](https://github.com/traefik/traefik/pull/4933) by [juliens](https://github.com/juliens))
- **[server]** Use h2c from x/net to handle h2c requests ([#5045](https://github.com/traefik/traefik/pull/5045) by [juliens](https://github.com/juliens))
- **[server]** Dynamic Configuration Refactoring ([#4168](https://github.com/traefik/traefik/pull/4168) by [ldez](https://github.com/ldez))
- **[server]** Remove old global config and use new static config ([#4222](https://github.com/traefik/traefik/pull/4222) by [juliens](https://github.com/juliens))
- **[sticky-session]** HttpOnly and Secure flags on the affinity cookie ([#4947](https://github.com/traefik/traefik/pull/4947) by [gheibia](https://github.com/gheibia))
- **[tcp]** Adds TCP support ([#4587](https://github.com/traefik/traefik/pull/4587) by [juliens](https://github.com/juliens))
- **[tls]** Define a TLS section to group TLS, TLSOptions, and TLSStores. ([#5031](https://github.com/traefik/traefik/pull/5031) by [ldez](https://github.com/ldez))
- **[tls]** TLSOptions: handle conflict: same host name, different TLS options ([#5056](https://github.com/traefik/traefik/pull/5056) by [mpl](https://github.com/mpl))
- **[tls]** Define TLS options on the Router configuration ([#4931](https://github.com/traefik/traefik/pull/4931) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[tls]** Expand Client Auth Type configuration ([#5078](https://github.com/traefik/traefik/pull/5078) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[tracing]** Improve tracing ([#5010](https://github.com/traefik/traefik/pull/5010) by [mmatur](https://github.com/mmatur))
- **[tracing]** Add Jaeger collector endpoint ([#5082](https://github.com/traefik/traefik/pull/5082) by [rmfitzpatrick](https://github.com/rmfitzpatrick))
- **[tracing]** Update tracing dependencies ([#4721](https://github.com/traefik/traefik/pull/4721) by [ldez](https://github.com/ldez))
- **[tracing]** Added support for Haystack tracing ([#4555](https://github.com/traefik/traefik/pull/4555) by [aantono](https://github.com/aantono))
- **[tracing]** Update Zipkin OpenTracing driver to latest 0.4.3 release ([#5283](https://github.com/traefik/traefik/pull/5283) by [basvanbeek](https://github.com/basvanbeek))
- **[tracing]** Instana tracer implementation ([#4453](https://github.com/traefik/traefik/pull/4453) by [notsureifkevin](https://github.com/notsureifkevin))
- **[tracing]** Make Zipkin trace rate configurable ([#3968](https://github.com/traefik/traefik/pull/3968) by [negz](https://github.com/negz))
- **[webui]** refactor(webui): use @vue/cli to bootstrap new ui ([#5091](https://github.com/traefik/traefik/pull/5091) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Add a new dashboard page ([#5249](https://github.com/traefik/traefik/pull/5249) by [Basgrani](https://github.com/Basgrani))
- **[webui]** Add doc and version in navbar ([#5137](https://github.com/traefik/traefik/pull/5137) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Use components to split Home concerns ([#5136](https://github.com/traefik/traefik/pull/5136) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Add more pages in the WebUI ([#5278](https://github.com/traefik/traefik/pull/5278) by [Basgrani](https://github.com/Basgrani))
- **[webui]** feat(webui/dashboard): init new dashboard ([#5105](https://github.com/traefik/traefik/pull/5105) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Upgrade angular cli version ([#4450](https://github.com/traefik/traefik/pull/4450) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Update docker node version ([#4448](https://github.com/traefik/traefik/pull/4448) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Ignore target/dependencies in docker copy ([#4449](https://github.com/traefik/traefik/pull/4449) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Format code with prettier ([#4463](https://github.com/traefik/traefik/pull/4463) by [Slashgear](https://github.com/Slashgear))
- **[webui]** No need for npm progress=false ([#3702](https://github.com/traefik/traefik/pull/3702) by [vdemeester](https://github.com/vdemeester))
- **[webui]** Migrate to a work in progress webui ([#4568](https://github.com/traefik/traefik/pull/4568) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Include lint in build process ([#4462](https://github.com/traefik/traefik/pull/4462) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Dropping rxjs-compat in favor of pipe ([#4520](https://github.com/traefik/traefik/pull/4520) by [imcotton](https://github.com/imcotton))
- Move dynamic config into a dedicated package. ([#5075](https://github.com/traefik/traefik/pull/5075) by [ldez](https://github.com/ldez))
- Disable collect data by default. ([#5393](https://github.com/traefik/traefik/pull/5393) by [ldez](https://github.com/ldez))
- Bump x/sys to support Risc-V architecture ([#5245](https://github.com/traefik/traefik/pull/5245) by [carlosedp](https://github.com/carlosedp))
- New packaging system. ([#4593](https://github.com/traefik/traefik/pull/4593) by [ldez](https://github.com/ldez))
- Updates Backoff ([#4457](https://github.com/traefik/traefik/pull/4457) by [ldez](https://github.com/ldez))
- Remove the bug command ([#4556](https://github.com/traefik/traefik/pull/4556) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Small code enhancements ([#3712](https://github.com/traefik/traefik/pull/3712) by [mmatur](https://github.com/mmatur))
- Remove deprecated elements ([#3666](https://github.com/traefik/traefik/pull/3666) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Clean old ([#4612](https://github.com/traefik/traefik/pull/4612) by [ldez](https://github.com/ldez))
- Update anonymize/collect ([#4590](https://github.com/traefik/traefik/pull/4590) by [jbdoumenjou](https://github.com/jbdoumenjou))

**Bug fixes:**
- **[api,webui]** Improve documentation about API and Dashboard. ([#5364](https://github.com/traefik/traefik/pull/5364) by [ldez](https://github.com/ldez))
- **[api]** Add errors about unknown entryPoint in runtime api ([#5265](https://github.com/traefik/traefik/pull/5265) by [juliens](https://github.com/juliens))
- **[api]** Add provider in middleware chain ([#5334](https://github.com/traefik/traefik/pull/5334) by [juliens](https://github.com/juliens))
- **[cli]** fix: boolean flag parsing with map. ([#5372](https://github.com/traefik/traefik/pull/5372) by [ldez](https://github.com/ldez))
- **[cli]** Return an error when help is called on a non existing command. ([#4977](https://github.com/traefik/traefik/pull/4977) by [ldez](https://github.com/ldez))
- **[cli]** Filter env vars configuration ([#4985](https://github.com/traefik/traefik/pull/4985) by [ldez](https://github.com/ldez))
- **[cli]** Fix some CLI bugs ([#4989](https://github.com/traefik/traefik/pull/4989) by [ldez](https://github.com/ldez))
- **[cli]** Change the loading resource order ([#5007](https://github.com/traefik/traefik/pull/5007) by [ldez](https://github.com/ldez))
- **[cli]** Apply the case of the CLI flags for the configuration ([#5153](https://github.com/traefik/traefik/pull/5153) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[cli]** Don&#39;t allow non flag arguments by default. ([#4970](https://github.com/traefik/traefik/pull/4970) by [ldez](https://github.com/ldez))
- **[docker]** Insensitive case for allow-empty value. ([#4745](https://github.com/traefik/traefik/pull/4745) by [ldez](https://github.com/ldez))
- **[file]** fix: TLS configuration from directory. ([#5118](https://github.com/traefik/traefik/pull/5118) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd]** Fix log messages about label selector ([#4629](https://github.com/traefik/traefik/pull/4629) by [mpl](https://github.com/mpl))
- **[k8s,k8s/crd]** fix: TLS domains with IngressRoute. ([#5327](https://github.com/traefik/traefik/pull/5327) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd]** Remove IngressEndpoint in CRD provider ([#4616](https://github.com/traefik/traefik/pull/4616) by [juliens](https://github.com/juliens))
- **[logs]** fix: logger and context. ([#5370](https://github.com/traefik/traefik/pull/5370) by [ldez](https://github.com/ldez))
- **[logs]** fix: error log message. ([#5020](https://github.com/traefik/traefik/pull/5020) by [ldez](https://github.com/ldez))
- **[logs]** Fix typos in data collection message ([#4891](https://github.com/traefik/traefik/pull/4891) by [mpl](https://github.com/mpl))
- **[logs]** Allow user to configure traefik log ([#4604](https://github.com/traefik/traefik/pull/4604) by [mmatur](https://github.com/mmatur))
- **[metrics,tracing]** fix: Datadog case. ([#5272](https://github.com/traefik/traefik/pull/5272) by [ldez](https://github.com/ldez))
- **[metrics]** Fix prometheus metrics ([#5152](https://github.com/traefik/traefik/pull/5152) by [mmatur](https://github.com/mmatur))
- **[middleware,k8s,k8s/crd]** The chain middleware in k8s use middlewareRef ([#5290](https://github.com/traefik/traefik/pull/5290) by [juliens](https://github.com/juliens))
- **[middleware]** Set X-Forwarded-* headers ([#4707](https://github.com/traefik/traefik/pull/4707) by [mpl](https://github.com/mpl))
- **[middleware]** Fix `url.Parse` due to go1.12.8 changes. ([#5207](https://github.com/traefik/traefik/pull/5207) by [ldez](https://github.com/ldez))
- **[middleware]** fix: stripPrefix and stripPrefixRegex. ([#5291](https://github.com/traefik/traefik/pull/5291) by [ldez](https://github.com/ldez))
- **[middleware]** Improve rate limiter tests ([#5310](https://github.com/traefik/traefik/pull/5310) by [mpl](https://github.com/mpl))
- **[middleware]** Fix response modifier initial building ([#4719](https://github.com/traefik/traefik/pull/4719) by [mpl](https://github.com/mpl))
- **[middleware]** Remove X-Forwarded-(Uri, Method, Tls-Client-Cert and Tls-Client-Cert-Info) from untrusted IP ([#5012](https://github.com/traefik/traefik/pull/5012) by [stffabi](https://github.com/stffabi))
- **[middleware]** fix buffering middleware ([#5281](https://github.com/traefik/traefik/pull/5281) by [ldez](https://github.com/ldez))
- **[middleware]** Don&#39;t panic with undefined middleware ([#5289](https://github.com/traefik/traefik/pull/5289) by [ldez](https://github.com/ldez))
- **[middleware]** Properly add response headers for CORS ([#4857](https://github.com/traefik/traefik/pull/4857) by [dtomcej](https://github.com/dtomcej))
- **[rules]** Allow matching with FQDN hosts with trailing periods ([#4763](https://github.com/traefik/traefik/pull/4763) by [dtomcej](https://github.com/dtomcej))
- **[server]** Fix panic while server shutdown ([#4644](https://github.com/traefik/traefik/pull/4644) by [juliens](https://github.com/juliens))
- **[server]** Write HTTP server logs into the global logger. ([#5329](https://github.com/traefik/traefik/pull/5329) by [ldez](https://github.com/ldez))
- **[server]** Fix problem in aggregator provider ([#4625](https://github.com/traefik/traefik/pull/4625) by [juliens](https://github.com/juliens))
- **[server]** Fix lock problem in server ([#4600](https://github.com/traefik/traefik/pull/4600) by [juliens](https://github.com/juliens))
- **[service,websocket]** Fix recovered panic when websocket is mirrored ([#5255](https://github.com/traefik/traefik/pull/5255) by [juliens](https://github.com/juliens))
- **[tcp]** Fix EOF error ([#4733](https://github.com/traefik/traefik/pull/4733) by [juliens](https://github.com/juliens))
- **[tcp]** Don&#39;t add TCP proxy when error occurs during creation. ([#4858](https://github.com/traefik/traefik/pull/4858) by [ldez](https://github.com/ldez))
- **[tcp]** Remove first byte wait when tcp catches all ([#4938](https://github.com/traefik/traefik/pull/4938) by [juliens](https://github.com/juliens))
- **[tcp]** On client CloseWrite, do CloseWrite instead of Close for backend ([#5366](https://github.com/traefik/traefik/pull/5366) by [juliens](https://github.com/juliens))
- **[tls]** Fix panic in TLS stores handling ([#4997](https://github.com/traefik/traefik/pull/4997) by [juliens](https://github.com/juliens))
- **[webui]** Rest provider icon in the webui ([#5261](https://github.com/traefik/traefik/pull/5261) by [mmatur](https://github.com/mmatur))
- **[webui]** Web UI graph names. ([#5389](https://github.com/traefik/traefik/pull/5389) by [ldez](https://github.com/ldez))
- **[webui]** fix: passHostHeader in the webUI. ([#5369](https://github.com/traefik/traefik/pull/5369) by [ldez](https://github.com/ldez))
- Fix trailing slash with check new version ([#5266](https://github.com/traefik/traefik/pull/5266) by [mmatur](https://github.com/mmatur))
- Ensure WaitGroup.Done() is always called ([#5026](https://github.com/traefik/traefik/pull/5026) by [bsdelf](https://github.com/bsdelf))
- Clean files during tests. ([#4607](https://github.com/traefik/traefik/pull/4607) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme,docker]** Removed extra colon before the 8080 docker port ([#5209](https://github.com/traefik/traefik/pull/5209) by [fairwood136](https://github.com/fairwood136))
- **[acme,docker]** Add a docker-compose &amp; let&#39;s encrypt user-guide ([#5121](https://github.com/traefik/traefik/pull/5121) by [pbenefice](https://github.com/pbenefice))
- **[acme,docker]** Synchronize documentation ([#4571](https://github.com/traefik/traefik/pull/4571) by [juliens](https://github.com/juliens))
- **[acme,k8s,k8s/crd]** Full ACME+CRD example ([#4652](https://github.com/traefik/traefik/pull/4652) by [mpl](https://github.com/mpl))
- **[acme,k8s/crd]** Fix: CRD user guide ([#5244](https://github.com/traefik/traefik/pull/5244) by [ldez](https://github.com/ldez))
- **[acme,tls]** docs: rewrite of the HTTPS and TLS section ([#4980](https://github.com/traefik/traefik/pull/4980) by [mpl](https://github.com/mpl))
- **[acme]** Lets encrypt documentation typo ([#5127](https://github.com/traefik/traefik/pull/5127) by [juliens](https://github.com/juliens))
- **[acme]** Use the same case every where for entryPoints. ([#4764](https://github.com/traefik/traefik/pull/4764) by [ldez](https://github.com/ldez))
- **[acme]** doc/crd-acme: specify required kubectl version ([#5015](https://github.com/traefik/traefik/pull/5015) by [mpl](https://github.com/mpl))
- **[acme]** Enhance manual dnsChallenge documentation ([#4636](https://github.com/traefik/traefik/pull/4636) by [ntaranov](https://github.com/ntaranov))
- **[acme]** Fix error in the documentation for CLI configuration example ([#5392](https://github.com/traefik/traefik/pull/5392) by [MycTl](https://github.com/MycTl))
- **[acme]** Add note about ACME renewal ([#4860](https://github.com/traefik/traefik/pull/4860) by [dtomcej](https://github.com/dtomcej))
- **[acme]** Fix acme example ([#5130](https://github.com/traefik/traefik/pull/5130) by [jamct](https://github.com/jamct))
- **[acme]** Rename Docker_Acme.md to Readme.md ([#4025](https://github.com/traefik/traefik/pull/4025) by [vineetvermait](https://github.com/vineetvermait))
- **[acme]** Enhance acme page. ([#4611](https://github.com/traefik/traefik/pull/4611) by [ldez](https://github.com/ldez))
- **[acme]** fix: some DNS provider link. ([#3637](https://github.com/traefik/traefik/pull/3637) by [ldez](https://github.com/ldez))
- **[docker,marathon]** Update Dynamic Configuration Reference for both Docker and Marathon ([#5100](https://github.com/traefik/traefik/pull/5100) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[docker]** Remove traefik.port from documentation ([#4886](https://github.com/traefik/traefik/pull/4886) by [ldez](https://github.com/ldez))
- **[docker]** Fix two minor nits in Traefik 2.0 docs ([#4692](https://github.com/traefik/traefik/pull/4692) by [cfra](https://github.com/cfra))
- **[docker]** Fix Getting started ([#4646](https://github.com/traefik/traefik/pull/4646) by [mmatur](https://github.com/mmatur))
- **[docker]** docker-compose examples ([#4642](https://github.com/traefik/traefik/pull/4642) by [karnthis](https://github.com/karnthis))
- **[docker]** Clarify docs with labels in Swarm Mode ([#4847](https://github.com/traefik/traefik/pull/4847) by [mikesir87](https://github.com/mikesir87))
- **[file]** Update the file provider documentation ([#4588](https://github.com/traefik/traefik/pull/4588) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/crd]** k8s static configuration explanation ([#4767](https://github.com/traefik/traefik/pull/4767) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd]** doc: kubernetes CRD provider ([#4620](https://github.com/traefik/traefik/pull/4620) by [mpl](https://github.com/mpl))
- **[k8s,k8s/ingress]** Add documentation about Kubernetes Ingress provider ([#5112](https://github.com/traefik/traefik/pull/5112) by [mpl](https://github.com/mpl))
- **[k8s/crd]** user guide: fix a mistake in the deployment definition ([#5096](https://github.com/traefik/traefik/pull/5096) by [ldez](https://github.com/ldez))
- **[k8s]** Fix typo in the CRD documentation ([#4902](https://github.com/traefik/traefik/pull/4902) by [llussy](https://github.com/llussy))
- **[marathon]** Enhance Marathon documentation ([#4776](https://github.com/traefik/traefik/pull/4776) by [ldez](https://github.com/ldez))
- **[middleware,k8s,k8s/crd]** Fix typo: middleware -&gt; middlewares. ([#4781](https://github.com/traefik/traefik/pull/4781) by [ldez](https://github.com/ldez))
- **[middleware,k8s/crd]** doc: fix middleware names for CRD. ([#4966](https://github.com/traefik/traefik/pull/4966) by [ldez](https://github.com/ldez))
- **[middleware,provider]** fix the documentation about middleware labels. ([#4888](https://github.com/traefik/traefik/pull/4888) by [ldez](https://github.com/ldez))
- **[middleware]** Fix Kubernetes Docs for Middlewares ([#4943](https://github.com/traefik/traefik/pull/4943) by [HurricanKai](https://github.com/HurricanKai))
- **[middleware]** Adds a reference to the middleware overview. ([#4824](https://github.com/traefik/traefik/pull/4824) by [ldez](https://github.com/ldez))
- **[middleware]** docker-compose labels require $&#39;s to be escaped ([#5225](https://github.com/traefik/traefik/pull/5225) by [Makeshift](https://github.com/Makeshift))
- **[middleware]** Fix doc about removing headers ([#4708](https://github.com/traefik/traefik/pull/4708) by [mpl](https://github.com/mpl))
- **[middleware]** Remove invalid commas. ([#4706](https://github.com/traefik/traefik/pull/4706) by [ldez](https://github.com/ldez))
- **[middleware]** Adds middlewares examples for k8s. ([#4713](https://github.com/traefik/traefik/pull/4713) by [ldez](https://github.com/ldez))
- **[middleware]** Update the middleware documentation ([#4729](https://github.com/traefik/traefik/pull/4729) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware]** fix: stripPrefixRegex documentation. ([#5273](https://github.com/traefik/traefik/pull/5273) by [ldez](https://github.com/ldez))
- **[middleware]** Correct typo in documentation on rate limiting ([#4939](https://github.com/traefik/traefik/pull/4939) by [ableuler](https://github.com/ableuler))
- **[middleware]** Improve middleware documentation. ([#5003](https://github.com/traefik/traefik/pull/5003) by [ldez](https://github.com/ldez))
- **[middleware]** Enhance middleware examples. ([#4680](https://github.com/traefik/traefik/pull/4680) by [ldez](https://github.com/ldez))
- **[middleware]** docker-compose basic auth needs double dollar signs ([#4831](https://github.com/traefik/traefik/pull/4831) by [muhlemmer](https://github.com/muhlemmer))
- **[middleware]** Fixed a typo in label. ([#5128](https://github.com/traefik/traefik/pull/5128) by [jamct](https://github.com/jamct))
- **[middleware]** Review documentation ([#4798](https://github.com/traefik/traefik/pull/4798) by [ldez](https://github.com/ldez))
- **[middleware]** Kubernetes CRD documentation fixes ([#4971](https://github.com/traefik/traefik/pull/4971) by [orhanhenrik](https://github.com/orhanhenrik))
- **[middleware]** compress link fixed ([#4817](https://github.com/traefik/traefik/pull/4817) by [gato](https://github.com/gato))
- **[middleware]** Fix typo in forwardAuth middleware documentation ([#4638](https://github.com/traefik/traefik/pull/4638) by [AkeemMcLennon](https://github.com/AkeemMcLennon))
- **[middleware]** change doc references to scheme[Rr]edirect -&gt; redirect[Ss]cheme ([#4959](https://github.com/traefik/traefik/pull/4959) by [topiaruss](https://github.com/topiaruss))
- **[middleware]** Update headers middleware docs for kubernetes crd ([#4955](https://github.com/traefik/traefik/pull/4955) by [orhanhenrik](https://github.com/orhanhenrik))
- **[middleware]** Fix strip prefix documentation ([#4829](https://github.com/traefik/traefik/pull/4829) by [mmatur](https://github.com/mmatur))
- **[provider]** Improve providers documentation. ([#5050](https://github.com/traefik/traefik/pull/5050) by [ldez](https://github.com/ldez))
- **[rancher]** fix: Rancher documentation. ([#4818](https://github.com/traefik/traefik/pull/4818) by [ldez](https://github.com/ldez))
- **[rancher]** Specify that Rancher provider is for 1.x only ([#4923](https://github.com/traefik/traefik/pull/4923) by [bradjones1](https://github.com/bradjones1))
- **[server]** Add gRPC user guide ([#5042](https://github.com/traefik/traefik/pull/5042) by [ldez](https://github.com/ldez))
- **[tcp]** Use rule HostSNI in documentation ([#4592](https://github.com/traefik/traefik/pull/4592) by [bbinet](https://github.com/bbinet))
- **[tls]** fix: typo in routing example. ([#4849](https://github.com/traefik/traefik/pull/4849) by [ldez](https://github.com/ldez))
- **[tracing]** Improve tracing documentation ([#5102](https://github.com/traefik/traefik/pull/5102) by [mmatur](https://github.com/mmatur))
- **[tracing]** Fix typo in tracing docs ([#4737](https://github.com/traefik/traefik/pull/4737) by [timoschwarzer](https://github.com/timoschwarzer))
- **[webui]** change docs and adjust dashboard for v2 alpha ([#4632](https://github.com/traefik/traefik/pull/4632) by [SantoDE](https://github.com/SantoDE))
- doc: improve examples. ([#5132](https://github.com/traefik/traefik/pull/5132) by [ldez](https://github.com/ldez))
- Fixed readme misspelling ([#4882](https://github.com/traefik/traefik/pull/4882) by [antondalgren](https://github.com/antondalgren))
- Prepare release v2.0.0-rc2 ([#5293](https://github.com/traefik/traefik/pull/5293) by [ldez](https://github.com/ldez))
- Fix typos in documentation ([#4884](https://github.com/traefik/traefik/pull/4884) by [michael-k](https://github.com/michael-k))
- Fixed spelling typo ([#4848](https://github.com/traefik/traefik/pull/4848) by [mikesir87](https://github.com/mikesir87))
- Enhance the Retry Middleware Documentation ([#5298](https://github.com/traefik/traefik/pull/5298) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Clarification of the correct pronunciation of the word &#34;Traefik&#34; ([#4834](https://github.com/traefik/traefik/pull/4834) by [ylamlum-g4m](https://github.com/ylamlum-g4m))
- Improve the &#34;reading path&#34; for new contributors ([#4908](https://github.com/traefik/traefik/pull/4908) by [dduportal](https://github.com/dduportal))
- Fix some documentation issues ([#5286](https://github.com/traefik/traefik/pull/5286) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Entry points CLI description. ([#4896](https://github.com/traefik/traefik/pull/4896) by [ldez](https://github.com/ldez))
- Add Mathieu Lonjaret to maintainers ([#4950](https://github.com/traefik/traefik/pull/4950) by [emilevauge](https://github.com/emilevauge))
- Prepare release v2.0.0-alpha5 ([#4967](https://github.com/traefik/traefik/pull/4967) by [ldez](https://github.com/ldez))
- Minor fix in documentation ([#4811](https://github.com/traefik/traefik/pull/4811) by [mmatur](https://github.com/mmatur))
- Prepare release v2.0.0-alpha6. ([#4975](https://github.com/traefik/traefik/pull/4975) by [ldez](https://github.com/ldez))
- Fix a typo in documentation ([#4794](https://github.com/traefik/traefik/pull/4794) by [groovytron](https://github.com/groovytron))
- Prepare release v2.0.0-alpha4. ([#4788](https://github.com/traefik/traefik/pull/4788) by [ldez](https://github.com/ldez))
- Remove dumpcerts.sh ([#4783](https://github.com/traefik/traefik/pull/4783) by [ldez](https://github.com/ldez))
- Base of the migration guide ([#5263](https://github.com/traefik/traefik/pull/5263) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Prepare release v2.0.0-alpha7 ([#5001](https://github.com/traefik/traefik/pull/5001) by [ldez](https://github.com/ldez))
- misc documentation fixes ([#5302](https://github.com/traefik/traefik/pull/5302) by [mpl](https://github.com/mpl))
- Fix some minors errors on the documentation ([#4664](https://github.com/traefik/traefik/pull/4664) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Adds a note in traefik.sample.toml ([#4757](https://github.com/traefik/traefik/pull/4757) by [ldez](https://github.com/ldez))
- Prepare release v2.0.0-rc1 ([#5252](https://github.com/traefik/traefik/pull/5252) by [ldez](https://github.com/ldez))
- Use the same case everywhere ([#5043](https://github.com/traefik/traefik/pull/5043) by [ldez](https://github.com/ldez))
- Improve the Documentation with a Reference Section ([#4714](https://github.com/traefik/traefik/pull/4714) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Prepare release v2.0.0-alpha8 ([#5049](https://github.com/traefik/traefik/pull/5049) by [ldez](https://github.com/ldez))
- Add a basic Traefik install guide ([#5117](https://github.com/traefik/traefik/pull/5117) by [jbdoumenjou](https://github.com/jbdoumenjou))
- AML indent for domains under TLS documentation section ([#5173](https://github.com/traefik/traefik/pull/5173) by [edvincent](https://github.com/edvincent))
- Update to v2.0 readme links ([#4700](https://github.com/traefik/traefik/pull/4700) by [karnthis](https://github.com/karnthis))
- Prepare release v2.0.0-alpha3. ([#4693](https://github.com/traefik/traefik/pull/4693) by [ldez](https://github.com/ldez))
- Misc documentation fixes ([#5307](https://github.com/traefik/traefik/pull/5307) by [ldez](https://github.com/ldez))
- Update restrictions in the documentation. ([#5270](https://github.com/traefik/traefik/pull/5270) by [ldez](https://github.com/ldez))
- Prepare release v2.0.0-rc3 ([#5343](https://github.com/traefik/traefik/pull/5343) by [ldez](https://github.com/ldez))
- Fix typos in docs ([#4662](https://github.com/traefik/traefik/pull/4662) by [SeMeKh](https://github.com/SeMeKh))
- Update traefik.sample.toml ([#4657](https://github.com/traefik/traefik/pull/4657) by [ldez](https://github.com/ldez))
- fix: services configuration documentation. ([#5359](https://github.com/traefik/traefik/pull/5359) by [ldez](https://github.com/ldez))
- Remove old links in readme ([#4651](https://github.com/traefik/traefik/pull/4651) by [ldez](https://github.com/ldez))
- fix a service with one server .yaml example ([#5373](https://github.com/traefik/traefik/pull/5373) by [zaverden](https://github.com/zaverden))
- Prepare release v2.0.0-rc4 ([#5384](https://github.com/traefik/traefik/pull/5384) by [ldez](https://github.com/ldez))
- Fix dead maintainers link on the README.md ([#4639](https://github.com/traefik/traefik/pull/4639) by [benjaminch](https://github.com/benjaminch))
- Prepare release v2.0.0-beta1 ([#5129](https://github.com/traefik/traefik/pull/5129) by [ldez](https://github.com/ldez))
- Fix typo in documentation ([#5386](https://github.com/traefik/traefik/pull/5386) by [adrienbrignon](https://github.com/adrienbrignon))
- Prepare release v2.0.0-alpha2 ([#4635](https://github.com/traefik/traefik/pull/4635) by [ldez](https://github.com/ldez))
- Fix malformed rule ([#5133](https://github.com/traefik/traefik/pull/5133) by [dtomcej](https://github.com/dtomcej))
- Improve various parts of the documentation. ([#4996](https://github.com/traefik/traefik/pull/4996) by [ldez](https://github.com/ldez))
- Documentation Revamp ([#4475](https://github.com/traefik/traefik/pull/4475) by [geraldcroes](https://github.com/geraldcroes))
- Adds a maintainer&#39;s page into the documentation. ([#4614](https://github.com/traefik/traefik/pull/4614) by [ldez](https://github.com/ldez))
- Add Gerald, Jean-Baptiste and Damien to maintainers ([#3982](https://github.com/traefik/traefik/pull/3982) by [emilevauge](https://github.com/emilevauge))
- fix broken links in readme.md ([#3967](https://github.com/traefik/traefik/pull/3967) by [AndrewSav](https://github.com/AndrewSav))
- Add master overhaul notice ([#3961](https://github.com/traefik/traefik/pull/3961) by [emilevauge](https://github.com/emilevauge))
- Complete maintainers processes ([#3696](https://github.com/traefik/traefik/pull/3696) by [mmatur](https://github.com/mmatur))
- Complete maintainers processes ([#3681](https://github.com/traefik/traefik/pull/3681) by [emilevauge](https://github.com/emilevauge))
- Prepare release v2.0.0-alpha1 ([#4617](https://github.com/traefik/traefik/pull/4617) by [ldez](https://github.com/ldez))

**Misc:**
- Cherry pick v1.7 into v2.0 ([#5341](https://github.com/traefik/traefik/pull/5341) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Cherry pick v1.7 into v2.0 ([#5192](https://github.com/traefik/traefik/pull/5192) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into v2.0 ([#5115](https://github.com/traefik/traefik/pull/5115) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Cherry pick v1.7 into v2.0 ([#4948](https://github.com/traefik/traefik/pull/4948) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into v2.0 ([#4823](https://github.com/traefik/traefik/pull/4823) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into v2.0 ([#4787](https://github.com/traefik/traefik/pull/4787) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into v2.0 ([#4695](https://github.com/traefik/traefik/pull/4695) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Merge v2.0.0-rc1 into master  ([#5253](https://github.com/traefik/traefik/pull/5253) by [ldez](https://github.com/ldez))
- Merge branch v2.0 into master  ([#5180](https://github.com/traefik/traefik/pull/5180) by [ldez](https://github.com/ldez))
- Merge v2.0.0-alpha8 into master ([#5055](https://github.com/traefik/traefik/pull/5055) by [ldez](https://github.com/ldez))
- Merge current v2.0.0-alpha into master  ([#5022](https://github.com/traefik/traefik/pull/5022) by [ldez](https://github.com/ldez))
- Merge v2.0.0-alpha6 into master ([#4984](https://github.com/traefik/traefik/pull/4984) by [ldez](https://github.com/ldez))
- Merge v2.0.0-alpha4 into master ([#4789](https://github.com/traefik/traefik/pull/4789) by [ldez](https://github.com/ldez))
- Merge v2.0.0-alpha3 into master ([#4694](https://github.com/traefik/traefik/pull/4694) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4565](https://github.com/traefik/traefik/pull/4565) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Cherry pick v1.7 into master ([#4511](https://github.com/traefik/traefik/pull/4511) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4492](https://github.com/traefik/traefik/pull/4492) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4440](https://github.com/traefik/traefik/pull/4440) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4365](https://github.com/traefik/traefik/pull/4365) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4303](https://github.com/traefik/traefik/pull/4303) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4271](https://github.com/traefik/traefik/pull/4271) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4268](https://github.com/traefik/traefik/pull/4268) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4229](https://github.com/traefik/traefik/pull/4229) by [juliens](https://github.com/juliens))
- Cherry pick v1.7 into master ([#4206](https://github.com/traefik/traefik/pull/4206) by [ldez](https://github.com/ldez))
- Merge v1.7.4 into master ([#4137](https://github.com/traefik/traefik/pull/4137) by [ldez](https://github.com/ldez))
- Merge v1.7.3 into master ([#4046](https://github.com/traefik/traefik/pull/4046) by [ldez](https://github.com/ldez))
- Merge current v1.7 into master ([#3992](https://github.com/traefik/traefik/pull/3992) by [ldez](https://github.com/ldez))
- Merge v1.7.2 into master ([#3983](https://github.com/traefik/traefik/pull/3983) by [ldez](https://github.com/ldez))
- Merge v1.7.0 into master ([#3925](https://github.com/traefik/traefik/pull/3925) by [ldez](https://github.com/ldez))
- Merge v1.7.0-rc5 into master ([#3903](https://github.com/traefik/traefik/pull/3903) by [ldez](https://github.com/ldez))
- Merge v1.7.0-rc4 into master ([#3867](https://github.com/traefik/traefik/pull/3867) by [ldez](https://github.com/ldez))
- Merge v1.7.0-rc2 into master ([#3634](https://github.com/traefik/traefik/pull/3634) by [ldez](https://github.com/ldez))

## [v2.0.0-rc4](https://github.com/traefik/traefik/tree/v2.0.0-rc4) (2019-09-13)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-rc3...v2.0.0-rc4)

**Enhancements:**
- **[docker,k8s,k8s/crd,marathon,rancher,tcp]** Add weighted round robin load balancer on TCP ([#5380](https://github.com/traefik/traefik/pull/5380) by [juliens](https://github.com/juliens))
- **[k8s,k8s/crd,k8s/ingress]** Fix kubernetes id name ([#5383](https://github.com/traefik/traefik/pull/5383) by [mmatur](https://github.com/mmatur))
- **[k8s,k8s/crd]** Add passHostHeader and responseForwarding in IngressRoute ([#5368](https://github.com/traefik/traefik/pull/5368) by [juliens](https://github.com/juliens))

**Bug fixes:**
- **[api,webui]** Improve documentation about API and Dashboard. ([#5364](https://github.com/traefik/traefik/pull/5364) by [ldez](https://github.com/ldez))
- **[cli]** fix: boolean flag parsing with map. ([#5372](https://github.com/traefik/traefik/pull/5372) by [ldez](https://github.com/ldez))
- **[logs]** fix: logger and context. ([#5370](https://github.com/traefik/traefik/pull/5370) by [ldez](https://github.com/ldez))
- **[tcp]** On client CloseWrite, do CloseWrite instead of Close for backend ([#5366](https://github.com/traefik/traefik/pull/5366) by [juliens](https://github.com/juliens))
- **[webui]** fix: passHostHeader in the webUI. ([#5369](https://github.com/traefik/traefik/pull/5369) by [ldez](https://github.com/ldez))

**Documentation:**
- fix a service with one server .yaml example ([#5373](https://github.com/traefik/traefik/pull/5373) by [zaverden](https://github.com/zaverden))
- fix: services configuration documentation. ([#5359](https://github.com/traefik/traefik/pull/5359) by [ldez](https://github.com/ldez))

## [v1.7.16](https://github.com/traefik/traefik/tree/v1.7.16) (2019-09-13)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.15...v1.7.16)

**Bug fixes:**
- **[middleware,websocket]** implement Flusher and Hijacker for codeCatcher ([#5376](https://github.com/traefik/traefik/pull/5376) by [mpl](https://github.com/mpl))

## [v1.7.15](https://github.com/traefik/traefik/tree/v1.7.15) (2019-09-12)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.14...v1.7.15)

**Bug fixes:**
- **[authentication,k8s/ingress]** Kubernetes support for Auth.HeaderField ([#5235](https://github.com/traefik/traefik/pull/5235) by [ErikWegner](https://github.com/ErikWegner))
- **[k8s,k8s/ingress]** Finish kubernetes throttling refactoring ([#5269](https://github.com/traefik/traefik/pull/5269) by [mpl](https://github.com/mpl))
- **[k8s]** Throttle Kubernetes config refresh ([#4716](https://github.com/traefik/traefik/pull/4716) by [benweissmann](https://github.com/benweissmann))
- **[k8s]** Fix wrong handling of insecure tls auth forward ingress annotation ([#5319](https://github.com/traefik/traefik/pull/5319) by [majkrzak](https://github.com/majkrzak))
- **[middleware]** error pages: do not buffer response when it&#39;s not an error ([#5285](https://github.com/traefik/traefik/pull/5285) by [mpl](https://github.com/mpl))
- **[tls]** Consider default cert domain in certificate store ([#5353](https://github.com/traefik/traefik/pull/5353) by [nrwiersma](https://github.com/nrwiersma))
- **[tls]** Add TLS minversion constraint ([#5356](https://github.com/traefik/traefik/pull/5356) by [dtomcej](https://github.com/dtomcej))

**Documentation:**
- **[acme]** Update Acme doc - Vultr Wildcard &amp; Root ([#5320](https://github.com/traefik/traefik/pull/5320) by [ddymko](https://github.com/ddymko))
- **[consulcatalog]** Typo in basic auth usersFile label consul-catalog ([#5230](https://github.com/traefik/traefik/pull/5230) by [pitan](https://github.com/pitan))
- **[logs]** Improve Access Logs Documentation page ([#5238](https://github.com/traefik/traefik/pull/5238) by [dduportal](https://github.com/dduportal))

## [v2.0.0-rc3](https://github.com/traefik/traefik/tree/v2.0.0-rc3) (2019-09-10)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-rc2...v2.0.0-rc3)

**Enhancements:**
- **[acme,api,tracing]** New API security ([#5311](https://github.com/traefik/traefik/pull/5311) by [juliens](https://github.com/juliens))
- **[authentication,middleware,k8s,k8s/crd]** Auth middlewares in kubernetes CRD use secrets ([#5299](https://github.com/traefik/traefik/pull/5299) by [juliens](https://github.com/juliens))
- **[logs]** Default to CLF when accesslog format is unsupported ([#5314](https://github.com/traefik/traefik/pull/5314) by [mpl](https://github.com/mpl))
- **[middleware,k8s,k8s/crd]** k8s ErrorPage middleware now uses k8s service ([#5339](https://github.com/traefik/traefik/pull/5339) by [juliens](https://github.com/juliens))
- **[webui]** Add more pages in the WebUI ([#5278](https://github.com/traefik/traefik/pull/5278) by [Basgrani](https://github.com/Basgrani))

**Bug fixes:**
- **[api]** Add provider in middleware chain ([#5334](https://github.com/traefik/traefik/pull/5334) by [juliens](https://github.com/juliens))
- **[k8s,k8s/crd]** fix: TLS domains with IngressRoute. ([#5327](https://github.com/traefik/traefik/pull/5327) by [ldez](https://github.com/ldez))
- **[middleware]** Improve rate limiter tests ([#5310](https://github.com/traefik/traefik/pull/5310) by [mpl](https://github.com/mpl))
- **[server]** Write HTTP server logs into the global logger. ([#5329](https://github.com/traefik/traefik/pull/5329) by [ldez](https://github.com/ldez))

**Documentation:**
- Misc documentation fixes ([#5307](https://github.com/traefik/traefik/pull/5307) by [ldez](https://github.com/ldez))
- misc documentation fixes ([#5302](https://github.com/traefik/traefik/pull/5302) by [mpl](https://github.com/mpl))
- Enhance the Retry Middleware Documentation ([#5298](https://github.com/traefik/traefik/pull/5298) by [jbdoumenjou](https://github.com/jbdoumenjou))

**Misc:**
- Cherry pick v1.7 into v2.0 ([#5341](https://github.com/traefik/traefik/pull/5341) by [jbdoumenjou](https://github.com/jbdoumenjou))

## [v2.0.0-rc2](https://github.com/traefik/traefik/tree/v2.0.0-rc2) (2019-09-03)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-rc1...v2.0.0-rc2)

**Enhancements:**
- **[api]** Improve API for the web UI ([#5267](https://github.com/traefik/traefik/pull/5267) by [ldez](https://github.com/ldez))
- **[middleware,tracing]** Re enable ratelimit integration tests ([#5288](https://github.com/traefik/traefik/pull/5288) by [mmatur](https://github.com/mmatur))
- **[tracing]** Update Zipkin OpenTracing driver to latest 0.4.3 release ([#5283](https://github.com/traefik/traefik/pull/5283) by [basvanbeek](https://github.com/basvanbeek))

**Bug fixes:**
- **[api]** Add errors about unknown entryPoint in runtime api ([#5265](https://github.com/traefik/traefik/pull/5265) by [juliens](https://github.com/juliens))
- **[metrics,tracing]** fix: Datadog case. ([#5272](https://github.com/traefik/traefik/pull/5272) by [ldez](https://github.com/ldez))
- **[middleware,k8s,k8s/crd]** The chain middleware in k8s use middlewareRef ([#5290](https://github.com/traefik/traefik/pull/5290) by [juliens](https://github.com/juliens))
- **[middleware]** Don&#39;t panic with undefined middleware ([#5289](https://github.com/traefik/traefik/pull/5289) by [ldez](https://github.com/ldez))
- **[middleware]** fix buffering middleware ([#5281](https://github.com/traefik/traefik/pull/5281) by [ldez](https://github.com/ldez))
- **[middleware]** fix: stripPrefix and stripPrefixRegex. ([#5291](https://github.com/traefik/traefik/pull/5291) by [ldez](https://github.com/ldez))
- **[service,websocket]** Fix recovered panic when websocket is mirrored ([#5255](https://github.com/traefik/traefik/pull/5255) by [juliens](https://github.com/juliens))
- **[webui]** Rest provider icon in the webui ([#5261](https://github.com/traefik/traefik/pull/5261) by [mmatur](https://github.com/mmatur))
- Fix trailing slash with check new version ([#5266](https://github.com/traefik/traefik/pull/5266) by [mmatur](https://github.com/mmatur))

**Documentation:**
- **[middleware]** fix: stripPrefixRegex documentation. ([#5273](https://github.com/traefik/traefik/pull/5273) by [ldez](https://github.com/ldez))
- Fix some documentation issues ([#5286](https://github.com/traefik/traefik/pull/5286) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Update restrictions in the documentation. ([#5270](https://github.com/traefik/traefik/pull/5270) by [ldez](https://github.com/ldez))
- Base of the migration guide ([#5263](https://github.com/traefik/traefik/pull/5263) by [jbdoumenjou](https://github.com/jbdoumenjou))

## [v2.0.0-rc1](https://github.com/traefik/traefik/tree/v2.0.0-rc1) (2019-08-26)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-beta1...v2.0.0-rc1)

**Enhancements:**
- **[acme]** Improve acme logs. ([#5139](https://github.com/traefik/traefik/pull/5139) by [ldez](https://github.com/ldez))
- **[docker,k8s,k8s/crd,k8s/ingress]** chore: update docker and k8s ([#5174](https://github.com/traefik/traefik/pull/5174) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd,k8s/ingress]** Correct Kubernetes Ingress and IngressRoute port heuristic for choosing HTTPS ([#5167](https://github.com/traefik/traefik/pull/5167) by [seh](https://github.com/seh))
- **[k8s,k8s/ingress]** Add TLS-enabled Router ([#5162](https://github.com/traefik/traefik/pull/5162) by [dtomcej](https://github.com/dtomcej))
- **[middleware,provider]** Add Feature-Policy header support ([#5156](https://github.com/traefik/traefik/pull/5156) by [dtomcej](https://github.com/dtomcej))
- **[middleware]** Add rate limiter, rename maxConn into inFlightReq ([#5246](https://github.com/traefik/traefik/pull/5246) by [mpl](https://github.com/mpl))
- **[server]** WeightedRoundRobin load balancer ([#5237](https://github.com/traefik/traefik/pull/5237) by [juliens](https://github.com/juliens))
- **[server]** Adds mirroring service ([#5251](https://github.com/traefik/traefik/pull/5251) by [juliens](https://github.com/juliens))
- **[server]** Add support proxyprotocol v2 ([#4755](https://github.com/traefik/traefik/pull/4755) by [c0va23](https://github.com/c0va23))
- **[webui]** Add a new dashboard page ([#5249](https://github.com/traefik/traefik/pull/5249) by [Basgrani](https://github.com/Basgrani))
- **[webui]** Add doc and version in navbar ([#5137](https://github.com/traefik/traefik/pull/5137) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Use components to split Home concerns ([#5136](https://github.com/traefik/traefik/pull/5136) by [Slashgear](https://github.com/Slashgear))
- Bump x/sys to support Risc-V architecture ([#5245](https://github.com/traefik/traefik/pull/5245) by [carlosedp](https://github.com/carlosedp))

**Bug fixes:**
- **[cli]** Apply the case of the CLI flags for the configuration ([#5153](https://github.com/traefik/traefik/pull/5153) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[metrics]** Fix prometheus metrics ([#5152](https://github.com/traefik/traefik/pull/5152) by [mmatur](https://github.com/mmatur))
- **[middleware]** Fix `url.Parse` due to go1.12.8 changes. ([#5207](https://github.com/traefik/traefik/pull/5207) by [ldez](https://github.com/ldez))
- Ensure WaitGroup.Done() is always called ([#5026](https://github.com/traefik/traefik/pull/5026) by [bsdelf](https://github.com/bsdelf))

**Documentation:**
- **[acme,docker]** Add a docker-compose &amp; let&#39;s encrypt user-guide ([#5121](https://github.com/traefik/traefik/pull/5121) by [pbenefice](https://github.com/pbenefice))
- **[acme,docker]** Removed extra colon before the 8080 docker port ([#5209](https://github.com/traefik/traefik/pull/5209) by [fairwood136](https://github.com/fairwood136))
- **[acme,k8s/crd]** Fix: CRD user guide ([#5244](https://github.com/traefik/traefik/pull/5244) by [ldez](https://github.com/ldez))
- **[acme]** Fix acme example ([#5130](https://github.com/traefik/traefik/pull/5130) by [jamct](https://github.com/jamct))
- **[middleware]** docker-compose labels require $&#39;s to be escaped ([#5225](https://github.com/traefik/traefik/pull/5225) by [Makeshift](https://github.com/Makeshift))
- AML indent for domains under TLS documentation section ([#5173](https://github.com/traefik/traefik/pull/5173) by [edvincent](https://github.com/edvincent))
- Fix malformed rule ([#5133](https://github.com/traefik/traefik/pull/5133) by [dtomcej](https://github.com/dtomcej))
- doc: improve examples. ([#5132](https://github.com/traefik/traefik/pull/5132) by [ldez](https://github.com/ldez))

**Misc:**
- Cherry pick v1.7 into v2.0 ([#5192](https://github.com/traefik/traefik/pull/5192) by [ldez](https://github.com/ldez))

## [v1.7.14](https://github.com/traefik/traefik/tree/v1.7.14) (2019-08-14)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.13...v1.7.14)

**Bug fixes:**
- Update to go1.12.8 ([#5201](https://github.com/traefik/traefik/pull/5201) by [ldez](https://github.com/ldez)). HTTP/2 Denial of Service [CVE-2019-9512](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-9512) and [CVE-2019-9514](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-9514)
- **[server]** Make hijackConnectionTracker.Close thread safe ([#5194](https://github.com/traefik/traefik/pull/5194) by [jlevesy](https://github.com/jlevesy))

## [v1.7.13](https://github.com/traefik/traefik/tree/v1.7.13) (2019-08-07)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.12...v1.7.13)

**Bug fixes:**
- **[acme]** Update lego ([#5166](https://github.com/traefik/traefik/pull/5166) by [dabeck](https://github.com/dabeck))
- **[consulcatalog]** warning should not be a fail status ([#4537](https://github.com/traefik/traefik/pull/4537) by [saez0pub](https://github.com/saez0pub))
- **[docker]** Update docker api version ([#4909](https://github.com/traefik/traefik/pull/4909) by [dtomcej](https://github.com/dtomcej))
- **[dynamodb]** Use dynamodbav tags to override json tags. ([#5002](https://github.com/traefik/traefik/pull/5002) by [ldez](https://github.com/ldez))
- **[healthcheck]** Wrr loadbalancer honors old weight on recovered servers ([#5051](https://github.com/traefik/traefik/pull/5051) by [DougWagner](https://github.com/DougWagner))
- **[k8s]** Check for multiport services on Global Backend Ingress ([#5021](https://github.com/traefik/traefik/pull/5021) by [dtomcej](https://github.com/dtomcej))
- **[logs]** Allows logs to use local time zone instead of UTC ([#4954](https://github.com/traefik/traefik/pull/4954) by [dduportal](https://github.com/dduportal))
- **[middleware]** Clear TLS client headers if TLSMutualAuth is optional ([#4963](https://github.com/traefik/traefik/pull/4963) by [stffabi](https://github.com/stffabi))
- **[tls]** Add missing KeyUsages for default generated certificate ([#5150](https://github.com/traefik/traefik/pull/5150) by [dtomcej](https://github.com/dtomcej))

**Documentation:**
- **[acme]** Fixed doc link for AlibabaCloud ([#5109](https://github.com/traefik/traefik/pull/5109) by [ddymko](https://github.com/ddymko))
- **[docker]** Add example for CLI ([#5131](https://github.com/traefik/traefik/pull/5131) by [alvarezbruned](https://github.com/alvarezbruned))
- **[docker]** Use the latest stable version of traefik in the docs ([#4927](https://github.com/traefik/traefik/pull/4927) by [kolaente](https://github.com/kolaente))
- **[logs]** Update documentation to clarify the default format for logs ([#4953](https://github.com/traefik/traefik/pull/4953) by [dduportal](https://github.com/dduportal))
- **[rancher]** Add remarks about Rancher 2 ([#4999](https://github.com/traefik/traefik/pull/4999) by [ldez](https://github.com/ldez))
- **[tls]** Fixes the TLS Mutual Authentication documentation ([#5085](https://github.com/traefik/traefik/pull/5085) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Format YAML example on user guide ([#5067](https://github.com/traefik/traefik/pull/5067) by [gurayyildirim](https://github.com/gurayyildirim))
- Update Slack support channel references to Discourse community forum ([#5014](https://github.com/traefik/traefik/pull/5014) by [dduportal](https://github.com/dduportal))
- Updating Service Fabric documentation ([#5160](https://github.com/traefik/traefik/pull/5160) by [gheibia](https://github.com/gheibia))
- Improve API / Dashboard wording in documentation ([#4929](https://github.com/traefik/traefik/pull/4929) by [dduportal](https://github.com/dduportal))

## [v2.0.0-beta1](https://github.com/traefik/traefik/tree/v2.0.0-beta1) (2019-07-19)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-alpha8...v2.0.0-beta1)

**Enhancements:**
- **[acme]** Certificate resolvers. ([#5116](https://github.com/traefik/traefik/pull/5116) by [ldez](https://github.com/ldez))
- **[api,provider]** Enhance REST provider ([#5072](https://github.com/traefik/traefik/pull/5072) by [dtomcej](https://github.com/dtomcej))
- **[api]** Deal with multiple errors and their criticality ([#5070](https://github.com/traefik/traefik/pull/5070) by [mpl](https://github.com/mpl))
- **[api]** API: remove configuration of Entrypoint and Middlewares ([#5119](https://github.com/traefik/traefik/pull/5119) by [mpl](https://github.com/mpl))
- **[api]** Improve API endpoints ([#5080](https://github.com/traefik/traefik/pull/5080) by [ldez](https://github.com/ldez))
- **[api]** Manage status for TCP element in the endpoint overview. ([#5108](https://github.com/traefik/traefik/pull/5108) by [ldez](https://github.com/ldez))
- **[file]** Restrict traefik.toml to static configuration. ([#5090](https://github.com/traefik/traefik/pull/5090) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd]** Add scheme to IngressRoute. ([#5062](https://github.com/traefik/traefik/pull/5062) by [ldez](https://github.com/ldez))
- **[k8s,k8s/ingress]** Renamed `kubernetes` provider in `kubernetesIngress` provider ([#5068](https://github.com/traefik/traefik/pull/5068) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[logs]** Improve error on router without service. ([#5126](https://github.com/traefik/traefik/pull/5126) by [ldez](https://github.com/ldez))
- **[metrics]** Add Metrics ([#5111](https://github.com/traefik/traefik/pull/5111) by [mmatur](https://github.com/mmatur))
- **[middleware]** Disable RateLimit temporarily ([#5123](https://github.com/traefik/traefik/pull/5123) by [juliens](https://github.com/juliens))
- **[tls]** TLSOptions: handle conflict: same host name, different TLS options ([#5056](https://github.com/traefik/traefik/pull/5056) by [mpl](https://github.com/mpl))
- **[tls]** Expand Client Auth Type configuration ([#5078](https://github.com/traefik/traefik/pull/5078) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[tracing]** Add Jaeger collector endpoint ([#5082](https://github.com/traefik/traefik/pull/5082) by [rmfitzpatrick](https://github.com/rmfitzpatrick))
- **[webui]** refactor(webui): use @vue/cli to bootstrap new ui ([#5091](https://github.com/traefik/traefik/pull/5091) by [Slashgear](https://github.com/Slashgear))
- **[webui]** feat(webui/dashboard): init new dashboard ([#5105](https://github.com/traefik/traefik/pull/5105) by [Slashgear](https://github.com/Slashgear))
- Move dynamic config into a dedicated package. ([#5075](https://github.com/traefik/traefik/pull/5075) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[file]** fix: TLS configuration from directory. ([#5118](https://github.com/traefik/traefik/pull/5118) by [ldez](https://github.com/ldez))
- **[middleware]** Remove X-Forwarded-(Uri, Method, Tls-Client-Cert and Tls-Client-Cert-Info) from untrusted IP ([#5012](https://github.com/traefik/traefik/pull/5012) by [stffabi](https://github.com/stffabi))
- **[middleware]** Properly add response headers for CORS ([#4857](https://github.com/traefik/traefik/pull/4857) by [dtomcej](https://github.com/dtomcej))

**Documentation:**
- **[acme]** Lets encrypt documentation typo ([#5127](https://github.com/traefik/traefik/pull/5127) by [juliens](https://github.com/juliens))
- **[docker,marathon]** Update Dynamic Configuration Reference for both Docker and Marathon ([#5100](https://github.com/traefik/traefik/pull/5100) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s,k8s/ingress]** Add documentation about Kubernetes Ingress provider ([#5112](https://github.com/traefik/traefik/pull/5112) by [mpl](https://github.com/mpl))
- **[k8s/crd]** user guide: fix a mistake in the deployment definition ([#5096](https://github.com/traefik/traefik/pull/5096) by [ldez](https://github.com/ldez))
- **[middleware]** Fixed a typo in label. ([#5128](https://github.com/traefik/traefik/pull/5128) by [jamct](https://github.com/jamct))
- **[provider]** Improve providers documentation. ([#5050](https://github.com/traefik/traefik/pull/5050) by [ldez](https://github.com/ldez))
- **[tracing]** Improve tracing documentation ([#5102](https://github.com/traefik/traefik/pull/5102) by [mmatur](https://github.com/mmatur))
- Add a basic Traefik install guide ([#5117](https://github.com/traefik/traefik/pull/5117) by [jbdoumenjou](https://github.com/jbdoumenjou))

**Misc:**
- Cherry pick v1.7 into v2.0 ([#5115](https://github.com/traefik/traefik/pull/5115) by [jbdoumenjou](https://github.com/jbdoumenjou))

## [v2.0.0-alpha8](https://github.com/traefik/traefik/tree/v2.0.0-alpha8) (2019-07-01)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-alpha7...v2.0.0-alpha8)

**Enhancements:**
- **[api]** Adding content-header to api endpoints ([#5019](https://github.com/traefik/traefik/pull/5019) by [dalanmiller](https://github.com/dalanmiller))
- **[file]** Support YAML for the dynamic configuration. ([#5024](https://github.com/traefik/traefik/pull/5024) by [ldez](https://github.com/ldez))
- **[logs]** Drop headers by default in access logs. ([#5034](https://github.com/traefik/traefik/pull/5034) by [ldez](https://github.com/ldez))
- **[middleware,k8s/crd]** Handle cross-provider middleware in kubernetes CRD ([#5009](https://github.com/traefik/traefik/pull/5009) by [mpl](https://github.com/mpl))
- **[server]** Use h2c from x/net to handle h2c requests ([#5045](https://github.com/traefik/traefik/pull/5045) by [juliens](https://github.com/juliens))
- **[server]** Make HTTP Keep-Alive timeout configurable for backend connections ([#4983](https://github.com/traefik/traefik/pull/4983) by [mszabo-wikia](https://github.com/mszabo-wikia))
- **[tls]** Define a TLS section to group TLS, TLSOptions, and TLSStores. ([#5031](https://github.com/traefik/traefik/pull/5031) by [ldez](https://github.com/ldez))
- **[tracing]** Improve tracing ([#5010](https://github.com/traefik/traefik/pull/5010) by [mmatur](https://github.com/mmatur))

**Bug fixes:**
- **[cli]** Change the loading resource order ([#5007](https://github.com/traefik/traefik/pull/5007) by [ldez](https://github.com/ldez))
- **[logs]** fix: error log message. ([#5020](https://github.com/traefik/traefik/pull/5020) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme]** doc/crd-acme: specify required kubectl version ([#5015](https://github.com/traefik/traefik/pull/5015) by [mpl](https://github.com/mpl))
- **[middleware]** Improve middleware documentation. ([#5003](https://github.com/traefik/traefik/pull/5003) by [ldez](https://github.com/ldez))
- **[server]** Add gRPC user guide ([#5042](https://github.com/traefik/traefik/pull/5042) by [ldez](https://github.com/ldez))
- Use the same case everywhere ([#5043](https://github.com/traefik/traefik/pull/5043) by [ldez](https://github.com/ldez))

## [v2.0.0-alpha7](https://github.com/traefik/traefik/tree/v2.0.0-alpha7) (2019-06-21)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-alpha6...v2.0.0-alpha7)

**Enhancements:**
- **[api]** API: new contract ([#4964](https://github.com/traefik/traefik/pull/4964) by [mpl](https://github.com/mpl))
- **[k8s,k8s/crd,tls]** Define TLS options on the Router configuration for Kubernetes ([#4973](https://github.com/traefik/traefik/pull/4973) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware,provider]** Change the provider separator from . to @ ([#4982](https://github.com/traefik/traefik/pull/4982) by [ldez](https://github.com/ldez))
- **[provider]** Use name@provider instead of provider@name. ([#4990](https://github.com/traefik/traefik/pull/4990) by [ldez](https://github.com/ldez))
- **[provider]** New constraints management. ([#4965](https://github.com/traefik/traefik/pull/4965) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[cli]** Fix some CLI bugs ([#4989](https://github.com/traefik/traefik/pull/4989) by [ldez](https://github.com/ldez))
- **[cli]** Filter env vars configuration ([#4985](https://github.com/traefik/traefik/pull/4985) by [ldez](https://github.com/ldez))
- **[cli]** Return an error when help is called on a non existing command. ([#4977](https://github.com/traefik/traefik/pull/4977) by [ldez](https://github.com/ldez))
- **[tls]** Fix panic in TLS stores handling ([#4997](https://github.com/traefik/traefik/pull/4997) by [juliens](https://github.com/juliens))

**Documentation:**
- **[acme,tls]** docs: rewrite of the HTTPS and TLS section ([#4980](https://github.com/traefik/traefik/pull/4980) by [mpl](https://github.com/mpl))
- Improve various parts of the documentation. ([#4996](https://github.com/traefik/traefik/pull/4996) by [ldez](https://github.com/ldez))

## [v2.0.0-alpha6](https://github.com/traefik/traefik/tree/v2.0.0-alpha6) (2019-06-18)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-alpha5...v2.0.0-alpha6)

**Bug fixes:**
- **[cli]** Don&#39;t allow non flag arguments by default. ([#4970](https://github.com/traefik/traefik/pull/4970) by [ldez](https://github.com/ldez))

**Documentation:**
- **[middleware,k8s/crd]** doc: fix middleware names for CRD. ([#4966](https://github.com/traefik/traefik/pull/4966) by [ldez](https://github.com/ldez))
- **[middleware]** Kubernetes CRD documentation fixes ([#4971](https://github.com/traefik/traefik/pull/4971) by [orhanhenrik](https://github.com/orhanhenrik))

## [v2.0.0-alpha5](https://github.com/traefik/traefik/tree/v2.0.0-alpha5) (2019-06-17)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-alpha4...v2.0.0-alpha5)

**Enhancements:**
- **[acme]** Remove timeout/interval from the ACME Provider ([#4842](https://github.com/traefik/traefik/pull/4842) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[api]** API: expose runtime representation ([#4841](https://github.com/traefik/traefik/pull/4841) by [mpl](https://github.com/mpl))
- **[cli]** New static configuration loading system. ([#4935](https://github.com/traefik/traefik/pull/4935) by [ldez](https://github.com/ldez))
- **[k8s,k8s/crd,tcp]** Add support for TCP (in kubernetes CRD) ([#4885](https://github.com/traefik/traefik/pull/4885) by [mpl](https://github.com/mpl))
- **[server]** Rework loadbalancer support ([#4933](https://github.com/traefik/traefik/pull/4933) by [juliens](https://github.com/juliens))
- **[sticky-session]** HttpOnly and Secure flags on the affinity cookie ([#4947](https://github.com/traefik/traefik/pull/4947) by [gheibia](https://github.com/gheibia))
- **[tls]** Define TLS options on the Router configuration ([#4931](https://github.com/traefik/traefik/pull/4931) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[tracing]** Added support for Haystack tracing ([#4555](https://github.com/traefik/traefik/pull/4555) by [aantono](https://github.com/aantono))

**Bug fixes:**
- **[logs]** Fix typos in data collection message ([#4891](https://github.com/traefik/traefik/pull/4891) by [mpl](https://github.com/mpl))
- **[middleware]** change doc references to scheme[Rr]edirect -&gt; redirect[Ss]cheme ([#4959](https://github.com/traefik/traefik/pull/4959) by [topiaruss](https://github.com/topiaruss))
- **[rules]** Allow matching with FQDN hosts with trailing periods ([#4763](https://github.com/traefik/traefik/pull/4763) by [dtomcej](https://github.com/dtomcej))
- **[tcp]** Remove first byte wait when tcp catches all ([#4938](https://github.com/traefik/traefik/pull/4938) by [juliens](https://github.com/juliens))
- **[tcp]** Don&#39;t add TCP proxy when error occurs during creation. ([#4858](https://github.com/traefik/traefik/pull/4858) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme]** Add note about ACME renewal ([#4860](https://github.com/traefik/traefik/pull/4860) by [dtomcej](https://github.com/dtomcej))
- **[docker]** Remove traefik.port from documentation ([#4886](https://github.com/traefik/traefik/pull/4886) by [ldez](https://github.com/ldez))
- **[docker]** Clarify docs with labels in Swarm Mode ([#4847](https://github.com/traefik/traefik/pull/4847) by [mikesir87](https://github.com/mikesir87))
- **[k8s]** Fix typo in the CRD documentation ([#4902](https://github.com/traefik/traefik/pull/4902) by [llussy](https://github.com/llussy))
- **[middleware,provider]** fix the documentation about middleware labels. ([#4888](https://github.com/traefik/traefik/pull/4888) by [ldez](https://github.com/ldez))
- **[middleware]** Review documentation ([#4798](https://github.com/traefik/traefik/pull/4798) by [ldez](https://github.com/ldez))
- **[middleware]** compress link fixed ([#4817](https://github.com/traefik/traefik/pull/4817) by [gato](https://github.com/gato))
- **[middleware]** Fix strip prefix documentation ([#4829](https://github.com/traefik/traefik/pull/4829) by [mmatur](https://github.com/mmatur))
- **[middleware]** Fix Kubernetes Docs for Middlewares ([#4943](https://github.com/traefik/traefik/pull/4943) by [HurricanKai](https://github.com/HurricanKai))
- **[middleware]** Correct typo in documentation on rate limiting ([#4939](https://github.com/traefik/traefik/pull/4939) by [ableuler](https://github.com/ableuler))
- **[middleware]** docker-compose basic auth needs double dollar signs ([#4831](https://github.com/traefik/traefik/pull/4831) by [muhlemmer](https://github.com/muhlemmer))
- **[middleware]** Adds a reference to the middleware overview. ([#4824](https://github.com/traefik/traefik/pull/4824) by [ldez](https://github.com/ldez))
- **[middleware]** Update headers middleware docs for kubernetes crd ([#4955](https://github.com/traefik/traefik/pull/4955) by [orhanhenrik](https://github.com/orhanhenrik))
- **[rancher]** fix: Rancher documentation. ([#4818](https://github.com/traefik/traefik/pull/4818) by [ldez](https://github.com/ldez))
- **[rancher]** Specify that Rancher provider is for 1.x only ([#4923](https://github.com/traefik/traefik/pull/4923) by [bradjones1](https://github.com/bradjones1))
- **[tls]** fix: typo in routing example. ([#4849](https://github.com/traefik/traefik/pull/4849) by [ldez](https://github.com/ldez))
- Clarification of the correct pronunciation of the word &#34;Traefik&#34; ([#4834](https://github.com/traefik/traefik/pull/4834) by [ylamlum-g4m](https://github.com/ylamlum-g4m))
- Fix typos in documentation ([#4884](https://github.com/traefik/traefik/pull/4884) by [michael-k](https://github.com/michael-k))
- Entry points CLI description. ([#4896](https://github.com/traefik/traefik/pull/4896) by [ldez](https://github.com/ldez))
- Improve the &#34;reading path&#34; for new contributors ([#4908](https://github.com/traefik/traefik/pull/4908) by [dduportal](https://github.com/dduportal))
- Fixed spelling typo ([#4848](https://github.com/traefik/traefik/pull/4848) by [mikesir87](https://github.com/mikesir87))
- Fixed readme misspelling ([#4882](https://github.com/traefik/traefik/pull/4882) by [antondalgren](https://github.com/antondalgren))
- Minor fix in documentation ([#4811](https://github.com/traefik/traefik/pull/4811) by [mmatur](https://github.com/mmatur))
- Add Mathieu Lonjaret to maintainers ([#4950](https://github.com/traefik/traefik/pull/4950) by [emilevauge](https://github.com/emilevauge))
- Fix a typo in documentation ([#4794](https://github.com/traefik/traefik/pull/4794) by [groovytron](https://github.com/groovytron))

**Misc:**
- Cherry pick v1.7 into v2.0 ([#4948](https://github.com/traefik/traefik/pull/4948) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into v2.0 ([#4823](https://github.com/traefik/traefik/pull/4823) by [ldez](https://github.com/ldez))

## [v1.7.12](https://github.com/traefik/traefik/tree/v1.7.12) (2019-05-29)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.11...v1.7.12)

**Bug fixes:**
- **[acme]** Allow SANs for wildcards domain. ([#4821](https://github.com/traefik/traefik/pull/4821) by [vizv](https://github.com/vizv))
- **[acme]** fix: update lego. ([#4910](https://github.com/traefik/traefik/pull/4910) by [ldez](https://github.com/ldez))
- **[api,authentication]** Remove authentication hashes from API ([#4918](https://github.com/traefik/traefik/pull/4918) by [ldez](https://github.com/ldez))
- **[consul]** Enhance KV logs. ([#4877](https://github.com/traefik/traefik/pull/4877) by [ldez](https://github.com/ldez))
- **[k8s]** Fix kubernetes template for backend responseforwarding flushinterval setting ([#4901](https://github.com/traefik/traefik/pull/4901) by [ravilr](https://github.com/ravilr))
- **[metrics]** Upgraded Datadog tracing library to 1.13.0 ([#4878](https://github.com/traefik/traefik/pull/4878) by [aantono](https://github.com/aantono))
- **[server]** Add missing callback on close of hijacked connections ([#4900](https://github.com/traefik/traefik/pull/4900) by [ravilr](https://github.com/ravilr))

**Documentation:**
- **[docker]** Docs: Troubleshooting help for Docker Swarm labels ([#4751](https://github.com/traefik/traefik/pull/4751) by [gregberns](https://github.com/gregberns))
- **[logs]** Adds a log fields documentation. ([#4890](https://github.com/traefik/traefik/pull/4890) by [ldez](https://github.com/ldez))

## [v1.7.11](https://github.com/traefik/traefik/tree/v1.7.11) (2019-04-26)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.10...v1.7.11)

**Enhancements:**
- **[k8s,k8s/ingress]** Enhance k8s tests maintainability ([#4696](https://github.com/traefik/traefik/pull/4696) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[acme]** fix: update lego. ([#4800](https://github.com/traefik/traefik/pull/4800) by [ldez](https://github.com/ldez))
- **[authentication,middleware]** Forward all header values from forward auth response ([#4515](https://github.com/traefik/traefik/pull/4515) by [ctas582](https://github.com/ctas582))
- **[cluster]** Remove usage of github.com/satori/go.uuid ([#4722](https://github.com/traefik/traefik/pull/4722) by [aaslamin](https://github.com/aaslamin))
- **[kv]** Enhance KV client error management ([#4819](https://github.com/traefik/traefik/pull/4819) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[tls]** Improve log message about redundant TLS certificate ([#4765](https://github.com/traefik/traefik/pull/4765) by [mpl](https://github.com/mpl))
- **[tracing]** Update zipkin-go-opentracing. ([#4720](https://github.com/traefik/traefik/pull/4720) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme]** Documentation Update: Hosting.de wildcard support tested ([#4747](https://github.com/traefik/traefik/pull/4747) by [martinhoefling](https://github.com/martinhoefling))
- **[acme]** Update Wildcard Domain documentation ([#4682](https://github.com/traefik/traefik/pull/4682) by [DWSR](https://github.com/DWSR))
- **[middleware]** Keep consistent order ([#4690](https://github.com/traefik/traefik/pull/4690) by [maxifom](https://github.com/maxifom))

## [v2.0.0-alpha4](https://github.com/traefik/traefik/tree/v2.0.0-alpha4) (2019-04-17)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-alpha3...v2.0.0-alpha4)

**Enhancements:**
- **[logs]** log.loglevel becomes log.level in configuration ([#4775](https://github.com/traefik/traefik/pull/4775) by [juliens](https://github.com/juliens))
- **[marathon,tcp]** Handle TCP in the marathon provider ([#4728](https://github.com/traefik/traefik/pull/4728) by [juliens](https://github.com/juliens))
- **[middleware]** Enable CORS configuration ([#3809](https://github.com/traefik/traefik/pull/3809) by [dtomcej](https://github.com/dtomcej))
- **[rancher]** Add Rancher provider ([#4647](https://github.com/traefik/traefik/pull/4647) by [SantoDE](https://github.com/SantoDE))
- **[tracing]** Update tracing dependencies ([#4721](https://github.com/traefik/traefik/pull/4721) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[docker]** Insensitive case for allow-empty value. ([#4745](https://github.com/traefik/traefik/pull/4745) by [ldez](https://github.com/ldez))
- **[middleware]** Fix response modifier initial building ([#4719](https://github.com/traefik/traefik/pull/4719) by [mpl](https://github.com/mpl))
- **[middleware]** Set X-Forwarded-* headers ([#4707](https://github.com/traefik/traefik/pull/4707) by [mpl](https://github.com/mpl))
- **[tcp]** Fix EOF error ([#4733](https://github.com/traefik/traefik/pull/4733) by [juliens](https://github.com/juliens))

**Documentation:**
- **[acme]** Use the same case every where for entryPoints. ([#4764](https://github.com/traefik/traefik/pull/4764) by [ldez](https://github.com/ldez))
- **[docker]** Fix two minor nits in Traefik 2.0 docs ([#4692](https://github.com/traefik/traefik/pull/4692) by [cfra](https://github.com/cfra))
- **[k8s,k8s/crd]** k8s static configuration explanation ([#4767](https://github.com/traefik/traefik/pull/4767) by [ldez](https://github.com/ldez))
- **[marathon]** Enhance Marathon documentation ([#4776](https://github.com/traefik/traefik/pull/4776) by [ldez](https://github.com/ldez))
- **[middleware,k8s,k8s/crd]** Fix typo: middleware -&gt; middlewares. ([#4781](https://github.com/traefik/traefik/pull/4781) by [ldez](https://github.com/ldez))
- **[middleware]** Adds middlewares examples for k8s. ([#4713](https://github.com/traefik/traefik/pull/4713) by [ldez](https://github.com/ldez))
- **[middleware]** Remove invalid commas. ([#4706](https://github.com/traefik/traefik/pull/4706) by [ldez](https://github.com/ldez))
- **[middleware]** Fix doc about removing headers ([#4708](https://github.com/traefik/traefik/pull/4708) by [mpl](https://github.com/mpl))
- **[middleware]** Update the middleware documentation ([#4729](https://github.com/traefik/traefik/pull/4729) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[tracing]** Fix typo in tracing docs ([#4737](https://github.com/traefik/traefik/pull/4737) by [timoschwarzer](https://github.com/timoschwarzer))
- Improve the Documentation with a Reference Section ([#4714](https://github.com/traefik/traefik/pull/4714) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Adds a note in traefik.sample.toml ([#4757](https://github.com/traefik/traefik/pull/4757) by [ldez](https://github.com/ldez))
- Update to v2.0 readme links ([#4700](https://github.com/traefik/traefik/pull/4700) by [karnthis](https://github.com/karnthis))
- Remove dumpcerts.sh ([#4783](https://github.com/traefik/traefik/pull/4783) by [ldez](https://github.com/ldez))

**Misc:**
- Cherry pick v1.7 into v2.0 ([#4787](https://github.com/traefik/traefik/pull/4787) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into v2.0 ([#4695](https://github.com/traefik/traefik/pull/4695) by [jbdoumenjou](https://github.com/jbdoumenjou))

## [v2.0.0-alpha3](https://github.com/traefik/traefik/tree/v2.0.0-alpha3) (2019-03-29)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-alpha2...v2.0.0-alpha3)

**Enhancements:**
- **[acme,k8s,k8s/crd]** Document the TLS with ACME case ([#4654](https://github.com/traefik/traefik/pull/4654) by [mpl](https://github.com/mpl))
- **[docker,tcp]** Add support for TCP labels in Docker provider ([#4621](https://github.com/traefik/traefik/pull/4621) by [juliens](https://github.com/juliens))
- **[provider]** Remove BaseProvider ([#4661](https://github.com/traefik/traefik/pull/4661) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[server]** Fix panic while server shutdown ([#4644](https://github.com/traefik/traefik/pull/4644) by [juliens](https://github.com/juliens))

**Documentation:**
- **[acme,k8s,k8s/crd]** Full ACME+CRD example ([#4652](https://github.com/traefik/traefik/pull/4652) by [mpl](https://github.com/mpl))
- **[acme]** Enhance manual dnsChallenge documentation ([#4636](https://github.com/traefik/traefik/pull/4636) by [ntaranov](https://github.com/ntaranov))
- **[docker]** Fix Getting started ([#4646](https://github.com/traefik/traefik/pull/4646) by [mmatur](https://github.com/mmatur))
- **[docker]** docker-compose examples ([#4642](https://github.com/traefik/traefik/pull/4642) by [karnthis](https://github.com/karnthis))
- **[middleware]** Fix typo in forwardAuth middleware documentation ([#4638](https://github.com/traefik/traefik/pull/4638) by [AkeemMcLennon](https://github.com/AkeemMcLennon))
- **[middleware]** Enhance middleware examples. ([#4680](https://github.com/traefik/traefik/pull/4680) by [ldez](https://github.com/ldez))
- Fix typos in docs ([#4662](https://github.com/traefik/traefik/pull/4662) by [SeMeKh](https://github.com/SeMeKh))
- Remove old links in readme ([#4651](https://github.com/traefik/traefik/pull/4651) by [ldez](https://github.com/ldez))
- Fix some minors errors on the documentation ([#4664](https://github.com/traefik/traefik/pull/4664) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Fix dead maintainers link on the README.md ([#4639](https://github.com/traefik/traefik/pull/4639) by [benjaminch](https://github.com/benjaminch))
- Update traefik.sample.toml ([#4657](https://github.com/traefik/traefik/pull/4657) by [ldez](https://github.com/ldez))

## [v2.0.0-alpha2](https://github.com/traefik/traefik/tree/v2.0.0-alpha2) (2019-03-19)
[All Commits](https://github.com/traefik/traefik/compare/v2.0.0-alpha1...v2.0.0-alpha2)

**Bug fixes:**
- **[k8s,k8s/crd]** Fix log messages about label selector ([#4629](https://github.com/traefik/traefik/pull/4629) by [mpl](https://github.com/mpl))
- **[server]** Fix problem in aggregator provider ([#4625](https://github.com/traefik/traefik/pull/4625) by [juliens](https://github.com/juliens))

**Documentation:**
- **[k8s,k8s/crd]** doc: kubernetes CRD provider ([#4620](https://github.com/traefik/traefik/pull/4620) by [mpl](https://github.com/mpl))
- **[webui]** change docs and adjust dashboard for v2 alpha ([#4632](https://github.com/traefik/traefik/pull/4632) by [SantoDE](https://github.com/SantoDE))

## [v2.0.0-alpha1](https://github.com/traefik/traefik/tree/v2.0.0-alpha1) (2019-03-18)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.0-rc1...v2.0.0-alpha1)

**Enhancements:**
- **[acme,kv]** Remove Deprecated StorageFile ([#4252](https://github.com/traefik/traefik/pull/4252) by [juliens](https://github.com/juliens))
- **[acme]** Migrate to go-acme/lego. ([#4589](https://github.com/traefik/traefik/pull/4589) by [ldez](https://github.com/ldez))
- **[authentication,logs,etcd]** Remove deprecated elements ([#3715](https://github.com/traefik/traefik/pull/3715) by [geraldcroes](https://github.com/geraldcroes))
- **[authentication,middleware]** Basic Auth custom realm ([#3917](https://github.com/traefik/traefik/pull/3917) by [tcoupin](https://github.com/tcoupin))
- **[docker]** Adds default rule system on Docker provider. ([#4413](https://github.com/traefik/traefik/pull/4413) by [ldez](https://github.com/ldez))
- **[docker]** Adds Docker provider support ([#4399](https://github.com/traefik/traefik/pull/4399) by [ldez](https://github.com/ldez))
- **[docker]** Update to Go1.12. Support of TLS1.3 ([#4540](https://github.com/traefik/traefik/pull/4540) by [ldez](https://github.com/ldez))
- **[etcd]** Remove etcd v2 ([#3739](https://github.com/traefik/traefik/pull/3739) by [geraldcroes](https://github.com/geraldcroes))
- **[k8s/ingress]** Adds Kubernetes provider support ([#4476](https://github.com/traefik/traefik/pull/4476) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[k8s/ingress]** Adds update ingress status ([#4603](https://github.com/traefik/traefik/pull/4603) by [juliens](https://github.com/juliens))
- **[k8s/ingress]** k8s integration tests ([#4569](https://github.com/traefik/traefik/pull/4569) by [juliens](https://github.com/juliens))
- **[k8s/ingress]** Custom resource definition ([#4591](https://github.com/traefik/traefik/pull/4591) by [ldez](https://github.com/ldez))
- **[marathon]** Adds Marathon support. ([#4415](https://github.com/traefik/traefik/pull/4415) by [ldez](https://github.com/ldez))
- **[metrics]** Add HTTP authentication to influxdb metric backend ([#3600](https://github.com/traefik/traefik/pull/3600) by [halfa](https://github.com/halfa))
- **[middleware,provider]** IPStrategy for selecting IP in whitelist ([#3778](https://github.com/traefik/traefik/pull/3778) by [juliens](https://github.com/juliens))
- **[middleware,provider]** Enables the use of elements declared in other providers ([#4372](https://github.com/traefik/traefik/pull/4372) by [geraldcroes](https://github.com/geraldcroes))
- **[middleware]** Migrates the pass client tls cert middleware ([#4373](https://github.com/traefik/traefik/pull/4373) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware]** Migrates Compress from bool to struct ([#3714](https://github.com/traefik/traefik/pull/3714) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware]** Updates for jaeger tracing client. ([#3688](https://github.com/traefik/traefik/pull/3688) by [tcolgate](https://github.com/tcolgate))
- **[middleware]** Add forwarded headers on entry point configuration ([#4364](https://github.com/traefik/traefik/pull/4364) by [juliens](https://github.com/juliens))
- **[middleware]** SchemeRedirect Middleware ([#4400](https://github.com/traefik/traefik/pull/4400) by [geraldcroes](https://github.com/geraldcroes))
- **[provider]** Add health check timeout parameter ([#3813](https://github.com/traefik/traefik/pull/3813) by [jbiel](https://github.com/jbiel))
- **[provider]** Removes deprecated templates ([#3649](https://github.com/traefik/traefik/pull/3649) by [geraldcroes](https://github.com/geraldcroes))
- **[provider]** Remove everything templates related ([#4595](https://github.com/traefik/traefik/pull/4595) by [mpl](https://github.com/mpl))
- **[provider]** Small code enhancements on providers ([#3707](https://github.com/traefik/traefik/pull/3707) by [vdemeester](https://github.com/vdemeester))
- **[provider]** Migrate rest provider ([#4253](https://github.com/traefik/traefik/pull/4253) by [juliens](https://github.com/juliens))
- **[provider]** Labels parser. ([#4236](https://github.com/traefik/traefik/pull/4236) by [ldez](https://github.com/ldez))
- **[rules]** New rule syntax ([#4437](https://github.com/traefik/traefik/pull/4437) by [juliens](https://github.com/juliens))
- **[server]** Dynamic Configuration Refactoring ([#4168](https://github.com/traefik/traefik/pull/4168) by [ldez](https://github.com/ldez))
- **[server]** Remove old global config and use new static config ([#4222](https://github.com/traefik/traefik/pull/4222) by [juliens](https://github.com/juliens))
- **[tcp]** Adds TCP support ([#4587](https://github.com/traefik/traefik/pull/4587) by [juliens](https://github.com/juliens))
- **[tracing]** Instana tracer implementation ([#4453](https://github.com/traefik/traefik/pull/4453) by [notsureifkevin](https://github.com/notsureifkevin))
- **[tracing]** Make Zipkin trace rate configurable ([#3968](https://github.com/traefik/traefik/pull/3968) by [negz](https://github.com/negz))
- **[webui]** Upgrade angular cli version ([#4450](https://github.com/traefik/traefik/pull/4450) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Update docker node version ([#4448](https://github.com/traefik/traefik/pull/4448) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Ignore target/dependencies in docker copy ([#4449](https://github.com/traefik/traefik/pull/4449) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Format code with prettier ([#4463](https://github.com/traefik/traefik/pull/4463) by [Slashgear](https://github.com/Slashgear))
- **[webui]** No need for npm progress=false ([#3702](https://github.com/traefik/traefik/pull/3702) by [vdemeester](https://github.com/vdemeester))
- **[webui]** Migrate to a work in progress webui ([#4568](https://github.com/traefik/traefik/pull/4568) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Include lint in build process ([#4462](https://github.com/traefik/traefik/pull/4462) by [Slashgear](https://github.com/Slashgear))
- **[webui]** Dropping rxjs-compat in favor of pipe ([#4520](https://github.com/traefik/traefik/pull/4520) by [imcotton](https://github.com/imcotton))
- New packaging system. ([#4593](https://github.com/traefik/traefik/pull/4593) by [ldez](https://github.com/ldez))
- Updates Backoff ([#4457](https://github.com/traefik/traefik/pull/4457) by [ldez](https://github.com/ldez))
- Remove the bug command ([#4556](https://github.com/traefik/traefik/pull/4556) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Small code enhancements ([#3712](https://github.com/traefik/traefik/pull/3712) by [mmatur](https://github.com/mmatur))
- Remove deprecated elements ([#3666](https://github.com/traefik/traefik/pull/3666) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Clean old ([#4612](https://github.com/traefik/traefik/pull/4612) by [ldez](https://github.com/ldez))
- Update anonymize/collect ([#4590](https://github.com/traefik/traefik/pull/4590) by [jbdoumenjou](https://github.com/jbdoumenjou))

**Bug fixes:**
- **[k8s,k8s/crd]** Remove IngressEndpoint in CRD provider ([#4616](https://github.com/traefik/traefik/pull/4616) by [juliens](https://github.com/juliens))
- **[logs]** Allow user to configure traefik log ([#4604](https://github.com/traefik/traefik/pull/4604) by [mmatur](https://github.com/mmatur))
- **[server]** Fix lock problem in server ([#4600](https://github.com/traefik/traefik/pull/4600) by [juliens](https://github.com/juliens))
- Clean files during tests. ([#4607](https://github.com/traefik/traefik/pull/4607) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme,docker]** Synchronize documentation ([#4571](https://github.com/traefik/traefik/pull/4571) by [juliens](https://github.com/juliens))
- **[acme]** Enhance acme page. ([#4611](https://github.com/traefik/traefik/pull/4611) by [ldez](https://github.com/ldez))
- **[acme]** Rename Docker_Acme.md to Readme.md ([#4025](https://github.com/traefik/traefik/pull/4025) by [vineetvermait](https://github.com/vineetvermait))
- **[acme]** fix: some DNS provider link. ([#3637](https://github.com/traefik/traefik/pull/3637) by [ldez](https://github.com/ldez))
- **[file]** Update the file provider documentation ([#4588](https://github.com/traefik/traefik/pull/4588) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[tcp]** Use rule HostSNI in documentation ([#4592](https://github.com/traefik/traefik/pull/4592) by [bbinet](https://github.com/bbinet))
- Documentation Revamp ([#4475](https://github.com/traefik/traefik/pull/4475) by [geraldcroes](https://github.com/geraldcroes))
- Add Gerald, Jean-Baptiste and Damien to maintainers ([#3982](https://github.com/traefik/traefik/pull/3982) by [emilevauge](https://github.com/emilevauge))
- fix broken links in readme.md ([#3967](https://github.com/traefik/traefik/pull/3967) by [AndrewSav](https://github.com/AndrewSav))
- Add master overhaul notice ([#3961](https://github.com/traefik/traefik/pull/3961) by [emilevauge](https://github.com/emilevauge))
- Complete maintainers processes ([#3696](https://github.com/traefik/traefik/pull/3696) by [mmatur](https://github.com/mmatur))
- Complete maintainers processes ([#3681](https://github.com/traefik/traefik/pull/3681) by [emilevauge](https://github.com/emilevauge))
- Adds a maintainer&#39;s page into the documentation. ([#4614](https://github.com/traefik/traefik/pull/4614) by [ldez](https://github.com/ldez))

**Misc:**
- Cherry pick v1.7 into master ([#4565](https://github.com/traefik/traefik/pull/4565) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Cherry pick v1.7 into master ([#4511](https://github.com/traefik/traefik/pull/4511) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4492](https://github.com/traefik/traefik/pull/4492) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4440](https://github.com/traefik/traefik/pull/4440) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4365](https://github.com/traefik/traefik/pull/4365) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4303](https://github.com/traefik/traefik/pull/4303) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4271](https://github.com/traefik/traefik/pull/4271) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4268](https://github.com/traefik/traefik/pull/4268) by [ldez](https://github.com/ldez))
- Cherry pick v1.7 into master ([#4229](https://github.com/traefik/traefik/pull/4229) by [juliens](https://github.com/juliens))
- Cherry pick v1.7 into master ([#4206](https://github.com/traefik/traefik/pull/4206) by [ldez](https://github.com/ldez))
- Merge v1.7.4 into master ([#4137](https://github.com/traefik/traefik/pull/4137) by [ldez](https://github.com/ldez))
- Merge v1.7.3 into master ([#4046](https://github.com/traefik/traefik/pull/4046) by [ldez](https://github.com/ldez))
- Merge current v1.7 into master ([#3992](https://github.com/traefik/traefik/pull/3992) by [ldez](https://github.com/ldez))
- Merge v1.7.2 into master ([#3983](https://github.com/traefik/traefik/pull/3983) by [ldez](https://github.com/ldez))
- Merge v1.7.0 into master ([#3925](https://github.com/traefik/traefik/pull/3925) by [ldez](https://github.com/ldez))
- Merge v1.7.0-rc5 into master ([#3903](https://github.com/traefik/traefik/pull/3903) by [ldez](https://github.com/ldez))
- Merge v1.7.0-rc4 into master ([#3867](https://github.com/traefik/traefik/pull/3867) by [ldez](https://github.com/ldez))
- Merge v1.7.0-rc2 into master ([#3634](https://github.com/traefik/traefik/pull/3634) by [ldez](https://github.com/ldez))

## [v1.7.10](https://github.com/traefik/traefik/tree/v1.7.10) (2019-03-28)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.9...v1.7.10)

**Bug fixes:**
- **[acme]** fix: update lego. ([#4670](https://github.com/traefik/traefik/pull/4670) by [ldez](https://github.com/ldez))
- **[acme]** Migrate to go-acme/lego. ([#4577](https://github.com/traefik/traefik/pull/4577) by [ldez](https://github.com/ldez))
- **[authentication,middleware]** Reorder Auth and TLSClientHeaders middleware ([#4557](https://github.com/traefik/traefik/pull/4557) by [tomberek](https://github.com/tomberek))
- **[k8s/ingress]** Support external name service on global default backend ([#4564](https://github.com/traefik/traefik/pull/4564) by [kippandrew](https://github.com/kippandrew))
- **[k8s/ingress]** Loop through service ports for global backend ([#4486](https://github.com/traefik/traefik/pull/4486) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Add entrypoints prefix in kubernetes frontend/backend id  ([#4679](https://github.com/traefik/traefik/pull/4679) by [juliens](https://github.com/juliens))
- **[websocket]** Exclude websocket connections from Average Response Time ([#4313](https://github.com/traefik/traefik/pull/4313) by [siyu6974](https://github.com/siyu6974))
- **[middleware]** Added support for configuring trace headers for Datadog tracing ([#4516](https://github.com/traefik/traefik/pull/4516) by [aantono](https://github.com/aantono))

**Documentation:**
- **[acme]** Add _FILE Environment Variable Documentation ([#4643](https://github.com/traefik/traefik/pull/4643) by [dargmuesli](https://github.com/dargmuesli))
- **[docker]** Add TraefikEE as security workaround ([#4606](https://github.com/traefik/traefik/pull/4606) by [emilevauge](https://github.com/emilevauge))

## [v1.7.9](https://github.com/traefik/traefik/tree/v1.7.9) (2019-02-11)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.8...v1.7.9)

**Bug fixes:**
- **[acme]** Updates of Lego. ([#4480](https://github.com/traefik/traefik/pull/4480) by [ldez](https://github.com/ldez))
- **[k8s]** app-root on non-explicit path include &#34;/&#34; in the redirect ([#4458](https://github.com/traefik/traefik/pull/4458) by [doctori](https://github.com/doctori))
- **[middleware]** Missing trailers with retry ([#4442](https://github.com/traefik/traefik/pull/4442) by [juliens](https://github.com/juliens))
- **[rancher]** Handle errors when working with rancher ([#4378](https://github.com/traefik/traefik/pull/4378) by [apsifly](https://github.com/apsifly))
- **[servicefabric]** Add support for specifying the name of the endpoint. ([#4479](https://github.com/traefik/traefik/pull/4479) by [ldez](https://github.com/ldez))
- **[tls]** insecureSkipVerify for the passTLSCert transport ([#4438](https://github.com/traefik/traefik/pull/4438) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[tracing]** Add Tracing Header Context Name option for Jaeger ([#4459](https://github.com/traefik/traefik/pull/4459) by [gadoor](https://github.com/gadoor))

**Documentation:**
- **[metrics]** Update default value of buckets for Prometheus ([#4468](https://github.com/traefik/traefik/pull/4468) by [adam-golab](https://github.com/adam-golab))
- **[rules]** Fixes the display of the associativity rules. ([#4478](https://github.com/traefik/traefik/pull/4478) by [ldez](https://github.com/ldez))
- Fixed curl example ([#4471](https://github.com/traefik/traefik/pull/4471) by [rgarrigue](https://github.com/rgarrigue))

## [v1.7.8](https://github.com/traefik/traefik/tree/v1.7.8) (2019-01-29)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.7...v1.7.8)

**Bug fixes:**
- **[acme]** Updates lego. ([#4428](https://github.com/traefik/traefik/pull/4428) by [ldez](https://github.com/ldez))
- **[acme]** Updates lego. ([#4376](https://github.com/traefik/traefik/pull/4376) by [ldez](https://github.com/ldez))
- **[docker]** Fixes docker swarm mode refresh second for KV. ([#4420](https://github.com/traefik/traefik/pull/4420) by [ldez](https://github.com/ldez))
- **[ecs]** Generic awsvpc support, not just Fargate ([#4360](https://github.com/traefik/traefik/pull/4360) by [maartenvanderhoef](https://github.com/maartenvanderhoef))
- **[ecs]** Cache existing task definitions to avoid rate limiting ([#4177](https://github.com/traefik/traefik/pull/4177) by [hwhelan-CB](https://github.com/hwhelan-CB))
- **[tls]** Check for dynamic tls updates on configuration preload ([#4022](https://github.com/traefik/traefik/pull/4022) by [ffilippopoulos](https://github.com/ffilippopoulos))
- **[tracing]** Support Datadog tracer priority sampling ([#4359](https://github.com/traefik/traefik/pull/4359) by [jcassee](https://github.com/jcassee))
- Update to Go 1.11.5 [CVE-2019-6486](https://nvd.nist.gov/vuln/detail/CVE-2019-6486)

**Documentation:**
- **[acme]** More detailed info about Google Cloud DNS. ([#4395](https://github.com/traefik/traefik/pull/4395) by [ldez](https://github.com/ldez))
- **[acme]** Tested wildcard ACME challenge with DNSimple ([#4384](https://github.com/traefik/traefik/pull/4384) by [tstackhouse](https://github.com/tstackhouse))
- **[docker]** Note about quotes for entrypoint definition with docker-compose ([#4390](https://github.com/traefik/traefik/pull/4390) by [Dragnucs](https://github.com/Dragnucs))
- **[k8s]** Allow Træfik to update Ingress status ([#4397](https://github.com/traefik/traefik/pull/4397) by [rbq](https://github.com/rbq))
- **[k8s]** Minor formatting fixes ([#4394](https://github.com/traefik/traefik/pull/4394) by [dbirks](https://github.com/dbirks))
- **[metrics]** Missing information about statistics parameter ([#4393](https://github.com/traefik/traefik/pull/4393) by [decima](https://github.com/decima))
- **[rules]** Route priorities: document minimum priority value ([#4374](https://github.com/traefik/traefik/pull/4374) by [tw-360vier](https://github.com/tw-360vier))
- Removed repeated entryPoints.http from grpc.md ([#4370](https://github.com/traefik/traefik/pull/4370) by [ishaanbahal](https://github.com/ishaanbahal))
- Happy 2019 ([#4367](https://github.com/traefik/traefik/pull/4367) by [emilevauge](https://github.com/emilevauge))

**Misc:**
- Assert that test timeout service is ready. ([#4398](https://github.com/traefik/traefik/pull/4398) by [timoreimann](https://github.com/timoreimann))

## [v1.7.7](https://github.com/traefik/traefik/tree/v1.7.7) (2019-01-08)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.6...v1.7.7)

**Bug fixes:**
- **[acme]** Update Lego ([#4277](https://github.com/traefik/traefik/pull/4277) by [ldez](https://github.com/ldez))
- **[k8s]** Check for watched namespace before getting kubernetes objects ([#4327](https://github.com/traefik/traefik/pull/4327) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Allow empty path with App-root annotation ([#4326](https://github.com/traefik/traefik/pull/4326) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** kubernetes: sort and uniq TLS secrets ([#4307](https://github.com/traefik/traefik/pull/4307) by [zarqman](https://github.com/zarqman))
- **[k8s]** Skip TLS section with no secret in Kubernetes ingress ([#4340](https://github.com/traefik/traefik/pull/4340) by [dtomcej](https://github.com/dtomcej))
- **[middleware,consul,consulcatalog,docker,ecs,k8s,marathon,mesos,rancher]** Add Pass TLS Cert Issuer and Domain Component ([#4298](https://github.com/traefik/traefik/pull/4298) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware]** Retry middleware : store headers per attempts and propagate them when responding. ([#4299](https://github.com/traefik/traefik/pull/4299) by [jlevesy](https://github.com/jlevesy))
- **[middleware]** Redirection status codes for methods different than GET ([#4116](https://github.com/traefik/traefik/pull/4116) by [r--w](https://github.com/r--w))
- Test and exit for jq error before domain loop ([#4347](https://github.com/traefik/traefik/pull/4347) by [muhlemmer](https://github.com/muhlemmer))

**Documentation:**
- **[acme]** Letsencrypt - Add info on httpreq format ([#4355](https://github.com/traefik/traefik/pull/4355) by [goetas](https://github.com/goetas))
- **[docker]** Update broken link for Docker service constraints ([#4289](https://github.com/traefik/traefik/pull/4289) by [clrech](https://github.com/clrech))
- **[middleware]** Add extractorfunc values ([#4351](https://github.com/traefik/traefik/pull/4351) by [hsmade](https://github.com/hsmade))
- **[provider]** Rephrase the `traefik.backend` definition in documentation ([#4317](https://github.com/traefik/traefik/pull/4317) by [dduportal](https://github.com/dduportal))
- Harden Traefik systemd service ([#4302](https://github.com/traefik/traefik/pull/4302) by [jacksgt](https://github.com/jacksgt))

## [v1.7.6](https://github.com/traefik/traefik/tree/v1.7.6) (2018-12-07)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.5...v1.7.6)

**Bug fixes:**
- **[consulcatalog]** Fix label segmentation when using custom prefix ([#4272](https://github.com/traefik/traefik/pull/4272) by [hsmade](https://github.com/hsmade))
- Update to Go 1.11.3 [CVE-2018-16875](https://nvd.nist.gov/vuln/detail/CVE-2018-16875)

## [v1.7.5](https://github.com/traefik/traefik/tree/v1.7.5) (2018-12-03)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.4...v1.7.5)

**Enhancements:**
- **[docker]** [docker backend] - Add config flag to set refreshSeconds for swarmmode ticker ([#4105](https://github.com/traefik/traefik/pull/4105) by [WTFKr0](https://github.com/WTFKr0))
- **[k8s]** Support canary weight for external name service ([#4135](https://github.com/traefik/traefik/pull/4135) by [yue9944882](https://github.com/yue9944882))

**Bug fixes:**
- **[acme]** Fix ACME spec and Cloudflare. ([#4201](https://github.com/traefik/traefik/pull/4201) by [ldez](https://github.com/ldez))
- **[authentication,middleware]** Remove X-Forwarded-Uri and X-Forwarded-Method from untrusted IP ([#4036](https://github.com/traefik/traefik/pull/4036) by [stffabi](https://github.com/stffabi))
- **[authentication,middleware]** Allow usersFile comments ([#4159](https://github.com/traefik/traefik/pull/4159) by [thde](https://github.com/thde))
- **[authentication]** Fix partial declaration of authentication. ([#4212](https://github.com/traefik/traefik/pull/4212) by [ldez](https://github.com/ldez))
- **[docker]** Verify ctx when we send configuration message in docker provider ([#4185](https://github.com/traefik/traefik/pull/4185) by [juliens](https://github.com/juliens))
- **[ecs]** Filter ECS tasks by LastStatus before adding to list of service tasks ([#4255](https://github.com/traefik/traefik/pull/4255) by [hwhelan-CB](https://github.com/hwhelan-CB))
- **[healthcheck]** Query params in health check ([#4188](https://github.com/traefik/traefik/pull/4188) by [mmatur](https://github.com/mmatur))
- **[metrics]** Upgraded DD APM library ([#4189](https://github.com/traefik/traefik/pull/4189) by [aantono](https://github.com/aantono))
- **[middleware]** Fix ssl force host secure middleware ([#4138](https://github.com/traefik/traefik/pull/4138) by [mmatur](https://github.com/mmatur))
- **[oxy]** Fix unannounced trailers problem when body is empty ([#4258](https://github.com/traefik/traefik/pull/4258) by [juliens](https://github.com/juliens))
- **[provider,server]** Log configuration errors from providers and keeps listening ([#4230](https://github.com/traefik/traefik/pull/4230) by [geraldcroes](https://github.com/geraldcroes))
- **[tls]** Implement Case-insensitive SNI matching ([#4132](https://github.com/traefik/traefik/pull/4132) by [dtomcej](https://github.com/dtomcej))
- Use ParseInt instead of Atoi for parsing durations ([#4263](https://github.com/traefik/traefik/pull/4263) by [mmatur](https://github.com/mmatur))

**Documentation:**
- **[acme]** ACME DNS provider is called `acme-dns` ([#4166](https://github.com/traefik/traefik/pull/4166) by [robsdedude](https://github.com/robsdedude))
- **[docker]** Add a &#34;Security Consideration&#34; section in the Docker&#39;s backend section of the documentation ([#4225](https://github.com/traefik/traefik/pull/4225) by [dduportal](https://github.com/dduportal))
- **[docker]** Clarify swarm loadbalancer documentation ([#4194](https://github.com/traefik/traefik/pull/4194) by [jlevesy](https://github.com/jlevesy))
- **[docker]** Fix spelling in comment ([#4169](https://github.com/traefik/traefik/pull/4169) by [giocomai](https://github.com/giocomai))
- **[docker]** Update swarm mode endpoint ([#4208](https://github.com/traefik/traefik/pull/4208) by [siyu6974](https://github.com/siyu6974))
- **[k8s]** Include an explicit list of kubernetes protocol annotations in docs. ([#4170](https://github.com/traefik/traefik/pull/4170) by [shanna](https://github.com/shanna))
- **[k8s]** Improve kubernetes TLS user guide ([#4175](https://github.com/traefik/traefik/pull/4175) by [mterring](https://github.com/mterring))
- **[k8s]** frame-deny should be set to true to enable the header ([#4171](https://github.com/traefik/traefik/pull/4171) by [swestcott](https://github.com/swestcott))
- **[rules]** Matcher associativity rule. ([#4244](https://github.com/traefik/traefik/pull/4244) by [ldez](https://github.com/ldez))
- Documentation: Rename &#34;admin panel&#34; to &#34;dashboard ([#4156](https://github.com/traefik/traefik/pull/4156) by [thernstig](https://github.com/thernstig))

## [v1.7.4](https://github.com/traefik/traefik/tree/v1.7.4) (2018-10-30)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.3...v1.7.4)

**Bug fixes:**
- **[acme]** Support custom DNS resolvers for Let&#39;s Encrypt. ([#4101](https://github.com/traefik/traefik/pull/4101) by [ldez](https://github.com/ldez))
- **[acme]** fix: netcup and DuckDNS. ([#4094](https://github.com/traefik/traefik/pull/4094) by [ldez](https://github.com/ldez))
- **[authentication,logs,middleware]** Fix display of client username field ([#4093](https://github.com/traefik/traefik/pull/4093) by [Ullaakut](https://github.com/Ullaakut))
- **[authentication,middleware]** Nil request body with retry ([#4075](https://github.com/traefik/traefik/pull/4075) by [ldez](https://github.com/ldez))
- **[consul,consulcatalog,docker,ecs,k8s,marathon,mesos,rancher]** Add flush interval option on backend ([#4112](https://github.com/traefik/traefik/pull/4112) by [juliens](https://github.com/juliens))
- **[consulcatalog,docker,ecs,marathon,mesos,rancher]** Remove the trailing dot if the domain is not defined. ([#4095](https://github.com/traefik/traefik/pull/4095) by [ldez](https://github.com/ldez))
- **[docker]** Provider docker shutdown problem ([#4122](https://github.com/traefik/traefik/pull/4122) by [juliens](https://github.com/juliens))
- **[k8s]** Add default path if nothing present ([#4097](https://github.com/traefik/traefik/pull/4097) by [SantoDE](https://github.com/SantoDE))
- **[k8s]** Add the missing pass-client-tls annotation to the kubernetes provider ([#4118](https://github.com/traefik/traefik/pull/4118) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[logs]** Fix access log field parsing ([#4113](https://github.com/traefik/traefik/pull/4113) by [Ullaakut](https://github.com/Ullaakut))
- **[middleware]** Add static redirect ([#4090](https://github.com/traefik/traefik/pull/4090) by [SantoDE](https://github.com/SantoDE))
- **[rules]** Add keepTrailingSlash option ([#4062](https://github.com/traefik/traefik/pull/4062) by [juliens](https://github.com/juliens))
- **[rules]** Case insensitive host rule  ([#3931](https://github.com/traefik/traefik/pull/3931) by [bgandon](https://github.com/bgandon))
- **[tls]** Fix certificate insertion loop to keep valid certificate and ignore the bad one ([#4050](https://github.com/traefik/traefik/pull/4050) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[webui]** Typo in the UI. ([#4096](https://github.com/traefik/traefik/pull/4096) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme]** Adds the note: acme.domains is a startup configuration ([#4065](https://github.com/traefik/traefik/pull/4065) by [geraldcroes](https://github.com/geraldcroes))
- **[acme]** exoscale move from .ch to .com ([#4130](https://github.com/traefik/traefik/pull/4130) by [greut](https://github.com/greut))
- **[acme]** Fixing a typo. ([#4124](https://github.com/traefik/traefik/pull/4124) by [konovalov-nk](https://github.com/konovalov-nk))
- **[acme]** Add a note about TLS-ALPN challenge. ([#4106](https://github.com/traefik/traefik/pull/4106) by [ldez](https://github.com/ldez))
- **[acme]** Clarify DuckDNS does not support multiple TXT records ([#4061](https://github.com/traefik/traefik/pull/4061) by [KnicKnic](https://github.com/KnicKnic))
- **[docker]** Domain is also optional for &#34;normal&#34; mode ([#4086](https://github.com/traefik/traefik/pull/4086) by [herver](https://github.com/herver))
- **[provider]** Fix mistake in the documentation of several backends ([#4133](https://github.com/traefik/traefik/pull/4133) by [whalehub](https://github.com/whalehub))
- Replaces emilevauge/whoami by containous/whoami in the documentation ([#4111](https://github.com/traefik/traefik/pull/4111) by [geraldcroes](https://github.com/geraldcroes))
- Uses ASCII characters to spell Traefik ([#4063](https://github.com/traefik/traefik/pull/4063) by [geraldcroes](https://github.com/geraldcroes))

**Misc:**
- **[tls]** Add double wildcard test ([#4091](https://github.com/traefik/traefik/pull/4091) by [dtomcej](https://github.com/dtomcej))
- **[webui]** Removed unused imports ([#4123](https://github.com/traefik/traefik/pull/4123) by [mwvdev](https://github.com/mwvdev))

## [v1.7.3](https://github.com/traefik/traefik/tree/v1.7.3) (2018-10-15)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.2...v1.7.3)

**Enhancements:**
- Improve the CLI help ([#3996](https://github.com/traefik/traefik/pull/3996) by [dduportal](https://github.com/dduportal))

**Bug fixes:**
- **[acme]** DNS challenge Cloudflare auth zone ([#4042](https://github.com/traefik/traefik/pull/4042) by [ldez](https://github.com/ldez))
- **[acme]** ACME DNS challenges ([#3998](https://github.com/traefik/traefik/pull/3998) by [ldez](https://github.com/ldez))
- **[acme]** Don&#39;t initialize ACME provider if storage is empty ([#3988](https://github.com/traefik/traefik/pull/3988) by [nmengin](https://github.com/nmengin))
- **[acme]** Fix: acme DNS providers ([#4021](https://github.com/traefik/traefik/pull/4021) by [ldez](https://github.com/ldez))
- **[acme]** Prevent some malformed errors in LE. ([#4015](https://github.com/traefik/traefik/pull/4015) by [ldez](https://github.com/ldez))
- **[authentication,consulcatalog,docker,ecs,etcd,kv,marathon,mesos,rancher]** Add the AuthResponseHeaders to the labels ([#3973](https://github.com/traefik/traefik/pull/3973) by [Crypto89](https://github.com/Crypto89))
- **[docker]** usebindportip can fall back on the container ip / port ([#4018](https://github.com/traefik/traefik/pull/4018) by [geraldcroes](https://github.com/geraldcroes))
- **[k8s]** Avoid flapping of multiple Ingress definitions ([#3862](https://github.com/traefik/traefik/pull/3862) by [rtreffer](https://github.com/rtreffer))
- **[middleware,server]** Log stack on panic ([#4033](https://github.com/traefik/traefik/pull/4033) by [ldez](https://github.com/ldez))
- **[middleware,server]** Fix recover from panic handler ([#4031](https://github.com/traefik/traefik/pull/4031) by [mmatur](https://github.com/mmatur))
- **[server,websocket]** Fix update oxy ([#4009](https://github.com/traefik/traefik/pull/4009) by [mmatur](https://github.com/mmatur))

**Documentation:**
- **[docker]** Add tags label to Docker provider documentation ([#3896](https://github.com/traefik/traefik/pull/3896) by [artheus](https://github.com/artheus))
- **[docker]** Added two examples with labels in docker-compose.yml ([#3891](https://github.com/traefik/traefik/pull/3891) by [pascalandy](https://github.com/pascalandy))
- **[k8s]** Move buffering annotation documentation to service ([#3991](https://github.com/traefik/traefik/pull/3991) by [ldez](https://github.com/ldez))
- Fix a typo ([#3995](https://github.com/traefik/traefik/pull/3995) by [arnydo](https://github.com/arnydo))

## [v1.7.2](https://github.com/traefik/traefik/tree/v1.7.2) (2018-10-04)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.1...v1.7.2)

**Bug fixes:**
- **[acme,cluster,kv]** TLS, ACME, cluster and several entrypoints. ([#3962](https://github.com/traefik/traefik/pull/3962) by [ldez](https://github.com/ldez))
- **[cluster,kv]** Correctly initialize kv store if storage key missing ([#3958](https://github.com/traefik/traefik/pull/3958) by [jfrabaute](https://github.com/jfrabaute))
- **[cluster,kv]** Return an error if kv store CA cert is invalid ([#3956](https://github.com/traefik/traefik/pull/3956) by [jfrabaute](https://github.com/jfrabaute))
- **[file]** Do not Errorf during file watcher verification test loop. ([#3938](https://github.com/traefik/traefik/pull/3938) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Add Template-ability check to Kubernetes API Fields ([#3964](https://github.com/traefik/traefik/pull/3964) by [dtomcej](https://github.com/dtomcej))
- **[logs]** Colored logs on windows. ([#3966](https://github.com/traefik/traefik/pull/3966) by [ldez](https://github.com/ldez))
- **[middleware]** Whitelist log for deprecated configuration. ([#3963](https://github.com/traefik/traefik/pull/3963) by [ldez](https://github.com/ldez))
- **[middleware]** Trimming whitespace in XFF for IP whitelisting ([#3971](https://github.com/traefik/traefik/pull/3971) by [olmoser](https://github.com/olmoser))
- **[rules]** Rule parsing error. ([#3976](https://github.com/traefik/traefik/pull/3976) by [ldez](https://github.com/ldez))
- Global configuration log at start ([#3954](https://github.com/traefik/traefik/pull/3954) by [ldez](https://github.com/ldez))

**Documentation:**
- **[logs]** Document the default accessLog format ([#3942](https://github.com/traefik/traefik/pull/3942) by [dfredell](https://github.com/dfredell))

## [v1.7.1](https://github.com/traefik/traefik/tree/v1.7.1) (2018-09-28)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.0...v1.7.1)

**Bug fixes:**
- **[acme,cluster]** Don&#39;t remove static certs from config when cluster mode ([#3946](https://github.com/traefik/traefik/pull/3946) by [Juliens](https://github.com/Juliens))
- **[acme]** Fix TLS ALPN cluster mode. ([#3934](https://github.com/traefik/traefik/pull/3934) by [ldez](https://github.com/ldez))
- **[acme]** Don&#39;t challenge ACME when host rule on another entry point ([#3923](https://github.com/traefik/traefik/pull/3923) by [Juliens](https://github.com/Juliens))
- **[tls]** Use the first static certificate as a fallback when no default is given ([#3948](https://github.com/traefik/traefik/pull/3948) by [Juliens](https://github.com/Juliens))

## [v1.7.0](https://github.com/traefik/traefik/tree/v1.7.0) (2018-09-24)
[Commits](https://github.com/traefik/traefik/compare/v1.7.0-rc1...v1.7.0)
[Commits pre RC](https://github.com/traefik/traefik/compare/v1.6.0-rc1...v1.7.0-rc1)

**Enhancements:**
- **[acme]** Simplify get acme client ([#3499](https://github.com/traefik/traefik/pull/3499) by [ldez](https://github.com/ldez))
- **[acme]** Simplify acme e2e tests. ([#3534](https://github.com/traefik/traefik/pull/3534) by [ldez](https://github.com/ldez))
- **[acme]** Add option to select algorithm to generate ACME certificates ([#3319](https://github.com/traefik/traefik/pull/3319) by [mmatur](https://github.com/mmatur))
- **[acme]** Enable to override certificates in key-value store when using storeconfig ([#3202](https://github.com/traefik/traefik/pull/3202) by [thomasjpfan](https://github.com/thomasjpfan))
- **[acme]** ACME TLS ALPN ([#3553](https://github.com/traefik/traefik/pull/3553) by [ldez](https://github.com/ldez))
- **[acme]** Remove acme provider dependency in server ([#3225](https://github.com/traefik/traefik/pull/3225) by [Juliens](https://github.com/Juliens))
- **[acme]** Use official Pebble Image. ([#3708](https://github.com/traefik/traefik/pull/3708) by [ldez](https://github.com/ldez))
- **[api,cluster]** Improved cluster api to include the current leader node ([#3100](https://github.com/traefik/traefik/pull/3100) by [aantono](https://github.com/aantono))
- **[authentication,consul,consulcatalog,docker,ecs,kv,marathon,mesos,rancher]** Auth support in frontends ([#3559](https://github.com/traefik/traefik/pull/3559) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[authentication,k8s]** Auth support in frontends for k8s and file ([#3460](https://github.com/traefik/traefik/pull/3460) by [Zatte](https://github.com/Zatte))
- **[authentication,middleware]** Add xforwarded method ([#3424](https://github.com/traefik/traefik/pull/3424) by [erik-sjoestedt](https://github.com/erik-sjoestedt))
- **[authentication,middleware]** Forward auth headers ([#3521](https://github.com/traefik/traefik/pull/3521) by [hwhelan-CB](https://github.com/hwhelan-CB))
- **[consul,etcd,tls]** Improve TLS integration tests ([#3679](https://github.com/traefik/traefik/pull/3679) by [mmatur](https://github.com/mmatur))
- **[consulcatalog,docker,ecs,file,k8s,kv,marathon,mesos,rancher]** Add SSLForceHost support. ([#3246](https://github.com/traefik/traefik/pull/3246) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Multiple frontends for consulcatalog ([#3796](https://github.com/traefik/traefik/pull/3796) by [hsmade](https://github.com/hsmade))
- **[consulcatalog]** Add support for stale reads from Consul catalog ([#3523](https://github.com/traefik/traefik/pull/3523) by [marenzo](https://github.com/marenzo))
- **[docker]** Add a default value for the docker.network configuration ([#3471](https://github.com/traefik/traefik/pull/3471) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[ecs]** Support for AWS ECS Fargate ([#3379](https://github.com/traefik/traefik/pull/3379) by [mmatur](https://github.com/mmatur))
- **[ecs]** Add support for ECS constraints ([#3537](https://github.com/traefik/traefik/pull/3537) by [andrewstucki](https://github.com/andrewstucki))
- **[ecs]** Add segment support for ECS  ([#3817](https://github.com/traefik/traefik/pull/3817) by [mmatur](https://github.com/mmatur))
- **[ecs]** Support `traefik.backend` for ECS ([#3510](https://github.com/traefik/traefik/pull/3510) by [hwhelan-CB](https://github.com/hwhelan-CB))
- **[ecs]** Allow binding ECS container port ([#3533](https://github.com/traefik/traefik/pull/3533) by [andrewstucki](https://github.com/andrewstucki))
- **[healthcheck,consul,consulcatalog,docker,ecs,kv,marathon,mesos,rancher]** Override health check scheme ([#3315](https://github.com/traefik/traefik/pull/3315) by [ldez](https://github.com/ldez))
- **[healthcheck]** Support 3xx HTTP status codes for health check ([#3364](https://github.com/traefik/traefik/pull/3364) by [SniperCZE](https://github.com/SniperCZE))
- **[healthcheck]** Support all 2xx HTTP status code for health check. ([#3362](https://github.com/traefik/traefik/pull/3362) by [ldez](https://github.com/ldez))
- **[healthcheck]** Add HTTP headers to healthcheck. ([#3047](https://github.com/traefik/traefik/pull/3047) by [zetaab](https://github.com/zetaab))
- **[k8s]** Add more k8s tests ([#3491](https://github.com/traefik/traefik/pull/3491) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Substitute hardcoded &#34;&lt;namespace&gt;/&lt;name&gt;&#34; with k8s ListerGetter ([#3470](https://github.com/traefik/traefik/pull/3470) by [yue9944882](https://github.com/yue9944882))
- **[k8s]** Custom frontend name for test helper ([#3444](https://github.com/traefik/traefik/pull/3444) by [ldez](https://github.com/ldez))
- **[k8s]** Add annotation to allow modifiers to be used properly in kubernetes ([#3481](https://github.com/traefik/traefik/pull/3481) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Create Global Backend Ingress ([#3404](https://github.com/traefik/traefik/pull/3404) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Specify backend servers&#39; weight via annotation for kubernetes ([#3112](https://github.com/traefik/traefik/pull/3112) by [yue9944882](https://github.com/yue9944882))
- **[k8s]** Support multi-port services. ([#3121](https://github.com/traefik/traefik/pull/3121) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Mapping ExternalNames to custom ports ([#3231](https://github.com/traefik/traefik/pull/3231) by [gildas](https://github.com/gildas))
- **[k8s]** Allow any kubernetes ingressClass value ([#3516](https://github.com/traefik/traefik/pull/3516) by [rtreffer](https://github.com/rtreffer))
- **[k8s]** Enable Ingress Status updates ([#3324](https://github.com/traefik/traefik/pull/3324) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Add possibility to set a protocol ([#3648](https://github.com/traefik/traefik/pull/3648) by [SantoDE](https://github.com/SantoDE))
- **[k8s]** Remove unnecessary loop ([#3799](https://github.com/traefik/traefik/pull/3799) by [ZloyDyadka](https://github.com/ZloyDyadka))
- **[kv]** Use index-based syntax in KV tests. ([#3352](https://github.com/traefik/traefik/pull/3352) by [ldez](https://github.com/ldez))
- **[logs,middleware]** Make accesslogs.logTheRoundTrip async to get lost performance ([#3152](https://github.com/traefik/traefik/pull/3152) by [ryarnyah](https://github.com/ryarnyah))
- **[logs,middleware]** Added duration filter for logs ([#3463](https://github.com/traefik/traefik/pull/3463) by [rodrigodiez](https://github.com/rodrigodiez))
- **[marathon]** Sane default and configurable Marathon request timeouts ([#3286](https://github.com/traefik/traefik/pull/3286) by [marco-jantke](https://github.com/marco-jantke))
- **[marathon]** Adding compatibility for marathon 1.5 ([#3505](https://github.com/traefik/traefik/pull/3505) by [TrevinTeacutter](https://github.com/TrevinTeacutter))
- **[mesos]** Segments Labels: Mesos ([#3383](https://github.com/traefik/traefik/pull/3383) by [drewkerrigan](https://github.com/drewkerrigan))
- **[metrics]** Metrics: Add support for InfluxDB Database / RetentionPolicy and HTTP client ([#3391](https://github.com/traefik/traefik/pull/3391) by [drewkerrigan](https://github.com/drewkerrigan))
- **[middleware,consulcatalog,docker,ecs,kv,marathon,mesos,rancher]** Pass the TLS Cert infos in headers ([#3826](https://github.com/traefik/traefik/pull/3826) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[middleware,server]** Extreme Makeover: server refactoring ([#3461](https://github.com/traefik/traefik/pull/3461) by [ldez](https://github.com/ldez))
- **[middleware,tracing]** Added integration support for Datadog APM Tracing ([#3517](https://github.com/traefik/traefik/pull/3517) by [aantono](https://github.com/aantono))
- **[middleware,tracing]** Create a custom logger for jaeger ([#3541](https://github.com/traefik/traefik/pull/3541) by [mmatur](https://github.com/mmatur))
- **[middleware]** Performance enhancements for the rules matchers. ([#3563](https://github.com/traefik/traefik/pull/3563) by [ShaneSaww](https://github.com/ShaneSaww))
- **[middleware]** Extract internal router creation from server ([#3204](https://github.com/traefik/traefik/pull/3204) by [Juliens](https://github.com/Juliens))
- **[rules]** CNAME flattening ([#3403](https://github.com/traefik/traefik/pull/3403) by [gamalan](https://github.com/gamalan))
- **[servicefabric]** Add HTTP headers to healthcheck. ([#3205](https://github.com/traefik/traefik/pull/3205) by [ldez](https://github.com/ldez))
- **[tls]** Support TLS MinVersion and CipherSuite as CLI option. ([#3107](https://github.com/traefik/traefik/pull/3107) by [ldez](https://github.com/ldez))
- **[tls]** Improve TLS Handshake ([#3512](https://github.com/traefik/traefik/pull/3512) by [dtomcej](https://github.com/dtomcej))
- **[webui]** Add some missing elements in the WebUI ([#3327](https://github.com/traefik/traefik/pull/3327) by [ldez](https://github.com/ldez))
- Call functions to enable block/mutex pprof profiles. ([#3564](https://github.com/traefik/traefik/pull/3564) by [timoreimann](https://github.com/timoreimann))
- Minor changes ([#3554](https://github.com/traefik/traefik/pull/3554) by [ldez](https://github.com/ldez))
- Generated assets file are only mandatory in main ([#3386](https://github.com/traefik/traefik/pull/3386) by [Juliens](https://github.com/Juliens))
- h2c server ([#3387](https://github.com/traefik/traefik/pull/3387) by [Juliens](https://github.com/Juliens))
- Fix backend reuse ([#3312](https://github.com/traefik/traefik/pull/3312) by [arnested](https://github.com/arnested))
- Upgrade GRPC dependencies ([#3342](https://github.com/traefik/traefik/pull/3342) by [gottwald](https://github.com/gottwald))
- Implement h2c with backend ([#3371](https://github.com/traefik/traefik/pull/3371) by [Juliens](https://github.com/Juliens))

**Bug fixes:**
- **[acme,cluster]** StoreConfig always initializes the account if it is missing ([#3844](https://github.com/traefik/traefik/pull/3844) by [geraldcroes](https://github.com/geraldcroes))
- **[acme,provider]** Create init method on provider interface ([#3580](https://github.com/traefik/traefik/pull/3580) by [Juliens](https://github.com/Juliens))
- **[acme]** Does not generate ACME certificate if domain is checked by dynamic certificate ([#3238](https://github.com/traefik/traefik/pull/3238) by [Juliens](https://github.com/Juliens))
- **[acme]** Ensure only certificates from ACME enabled entrypoint are used ([#3880](https://github.com/traefik/traefik/pull/3880) by [dtomcej](https://github.com/dtomcej))
- **[acme]** Fix acme account deletion without provider change ([#3664](https://github.com/traefik/traefik/pull/3664) by [zyclonite](https://github.com/zyclonite))
- **[acme]** Fix some DNS providers issues ([#3915](https://github.com/traefik/traefik/pull/3915) by [ldez](https://github.com/ldez))
- **[acme]** Fix LEGO update ([#3895](https://github.com/traefik/traefik/pull/3895) by [ldez](https://github.com/ldez))
- **[acme]** Set a keyType to ACME if the account is stored with no KeyType ([#3733](https://github.com/traefik/traefik/pull/3733) by [nmengin](https://github.com/nmengin))
- **[acme]** Fix ACME certificate for wildcard and root domains ([#3675](https://github.com/traefik/traefik/pull/3675) by [nmengin](https://github.com/nmengin))
- **[acme]** Update lego ([#3659](https://github.com/traefik/traefik/pull/3659) by [mmatur](https://github.com/mmatur))
- **[acme]** Bump LEGO version ([#3888](https://github.com/traefik/traefik/pull/3888) by [ldez](https://github.com/ldez))
- **[acme]** Serve TLS-Challenge certificate in first ([#3605](https://github.com/traefik/traefik/pull/3605) by [nmengin](https://github.com/nmengin))
- **[api,authentication,webui]** Auth section in web UI.  ([#3628](https://github.com/traefik/traefik/pull/3628) by [ldez](https://github.com/ldez))
- **[api]** Remove TLS in API ([#3665](https://github.com/traefik/traefik/pull/3665) by [mmatur](https://github.com/mmatur))
- **[authentication,consulcatalog,docker,ecs,k8s,kv,marathon,mesos,rancher]** Auth Forward with certificates in templates. ([#3804](https://github.com/traefik/traefik/pull/3804) by [ldez](https://github.com/ldez))
- **[authentication,middleware,provider]** Don&#39;t pass the Authorization header to the backends ([#3606](https://github.com/traefik/traefik/pull/3606) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[authentication,middleware]** Do not copy hop-by-hop headers to forward auth request ([#3907](https://github.com/traefik/traefik/pull/3907) by [stffabi](https://github.com/stffabi))
- **[authentication,middleware]** Remove hop-by-hop headers from forward auth response ([#3900](https://github.com/traefik/traefik/pull/3900) by [stffabi](https://github.com/stffabi))
- **[docker]** Uses both binded HostIP and HostPort when useBindPortIP=true ([#3638](https://github.com/traefik/traefik/pull/3638) by [geraldcroes](https://github.com/geraldcroes))
- **[ecs]** Fix 400 bad request on  AWS ECS API ([#3629](https://github.com/traefik/traefik/pull/3629) by [mmatur](https://github.com/mmatur))
- **[k8s]** Fix Rewrite-target regex ([#3699](https://github.com/traefik/traefik/pull/3699) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Don&#39;t merge kubernetes ingresses when priority is set ([#3743](https://github.com/traefik/traefik/pull/3743) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Prevent unparsable strings from being rendered in the Kubernetes template ([#3753](https://github.com/traefik/traefik/pull/3753) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Correct App-Root kubernetes behavior ([#3592](https://github.com/traefik/traefik/pull/3592) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Add more K8s Unit Tests ([#3583](https://github.com/traefik/traefik/pull/3583) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Fix rewrite-target Annotation behavior ([#3582](https://github.com/traefik/traefik/pull/3582) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Fix panic setting ingress status ([#3492](https://github.com/traefik/traefik/pull/3492) by [dtomcej](https://github.com/dtomcej))
- **[kv]** KV and authentication ([#3615](https://github.com/traefik/traefik/pull/3615) by [ldez](https://github.com/ldez))
- **[kv]** Add missing quotes around backendName in kv template ([#3885](https://github.com/traefik/traefik/pull/3885) by [NatMarchand](https://github.com/NatMarchand))
- **[kv]** Include missing key in error message for KV store ([#3779](https://github.com/traefik/traefik/pull/3779) by [camelpunch](https://github.com/camelpunch))
- **[logs]** Add logs when error is generated in error handler ([#3571](https://github.com/traefik/traefik/pull/3571) by [Juliens](https://github.com/Juliens))
- **[logs]** Add interface to Træfik logger ([#3889](https://github.com/traefik/traefik/pull/3889) by [nmengin](https://github.com/nmengin))
- **[metrics]** Avoid a panic during Prometheus registering ([#3717](https://github.com/traefik/traefik/pull/3717) by [nmengin](https://github.com/nmengin))
- **[middleware,tracing]** Fix tracing duplicated headers ([#3878](https://github.com/traefik/traefik/pull/3878) by [mmatur](https://github.com/mmatur))
- **[middleware,websocket]** Enable retry on websocket ([#3825](https://github.com/traefik/traefik/pull/3825) by [Juliens](https://github.com/Juliens))
- **[middleware]** Avoid retries when any data was written to the backend ([#3285](https://github.com/traefik/traefik/pull/3285) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** Extend https redirection tests, and fix incorrect behavior ([#3742](https://github.com/traefik/traefik/pull/3742) by [dtomcej](https://github.com/dtomcej))
- **[middleware]** Send &#39;Retry-After&#39; to comply with RFC6585. ([#3593](https://github.com/traefik/traefik/pull/3593) by [ldez](https://github.com/ldez))
- **[middleware]** Correct Entrypoint Redirect with Stripped or Added Path ([#3631](https://github.com/traefik/traefik/pull/3631) by [dtomcej](https://github.com/dtomcej))
- **[middleware]** Fix error pages ([#3894](https://github.com/traefik/traefik/pull/3894) by [Juliens](https://github.com/Juliens))
- **[oxy]** Handle Te header when http2 ([#3824](https://github.com/traefik/traefik/pull/3824) by [Juliens](https://github.com/Juliens))
- **[server]** Avoid goroutine leak in server ([#3851](https://github.com/traefik/traefik/pull/3851) by [nmengin](https://github.com/nmengin))
- **[server]** Avoid panic during stop ([#3898](https://github.com/traefik/traefik/pull/3898) by [nmengin](https://github.com/nmengin))
- **[tracing]** Added default configuration for Datadog APM Tracer ([#3655](https://github.com/traefik/traefik/pull/3655) by [aantono](https://github.com/aantono))
- **[tracing]** Added support for Trace name truncation for traces ([#3689](https://github.com/traefik/traefik/pull/3689) by [aantono](https://github.com/aantono))
- **[websocket]** Handle shutdown of Hijacked connections ([#3636](https://github.com/traefik/traefik/pull/3636) by [Juliens](https://github.com/Juliens))
- **[webui]** Added Dashboard table item for Rate Limits ([#3893](https://github.com/traefik/traefik/pull/3893) by [codecyclist](https://github.com/codecyclist))
- Fix logger in Oxy ([#3913](https://github.com/traefik/traefik/pull/3913) by [ldez](https://github.com/ldez))
- H2C: Remove buggy line in init to make verbose switch working ([#3701](https://github.com/traefik/traefik/pull/3701) by [dduportal](https://github.com/dduportal))
- Updating oxy dependency ([#3700](https://github.com/traefik/traefik/pull/3700) by [crholm](https://github.com/crholm))

**Documentation:**
- **[acme]** Update ACME documentation about TLS-ALPN challenge ([#3756](https://github.com/traefik/traefik/pull/3756) by [ldez](https://github.com/ldez))
- **[acme]** Fix some DNS provider link ([#3639](https://github.com/traefik/traefik/pull/3639) by [ldez](https://github.com/ldez))
- **[acme]** Fix documentation for route53 acme provider ([#3811](https://github.com/traefik/traefik/pull/3811) by [A-Shleifman](https://github.com/A-Shleifman))
- **[acme]** Update Namecheap status ([#3604](https://github.com/traefik/traefik/pull/3604) by [stoinov](https://github.com/stoinov))
- **[docker]** Fix style in examples/quickstart ([#3705](https://github.com/traefik/traefik/pull/3705) by [korigod](https://github.com/korigod))
- **[docker]** Change syntax in quick start guide ([#3726](https://github.com/traefik/traefik/pull/3726) by [trotro](https://github.com/trotro))
- **[docker]** Typo in docker-and-lets-encrypt.md ([#3724](https://github.com/traefik/traefik/pull/3724) by [A-Shleifman](https://github.com/A-Shleifman))
- **[docker]** Improve the wording in the documentation for Docker and fix title for Docker User Guide ([#3797](https://github.com/traefik/traefik/pull/3797) by [dduportal](https://github.com/dduportal))
- **[k8s]** Add a k8s guide section on traffic splitting via service weights. ([#3556](https://github.com/traefik/traefik/pull/3556) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Change code block of traefik-web-ui to match file ([#3542](https://github.com/traefik/traefik/pull/3542) by [drewgwallace](https://github.com/drewgwallace))
- **[k8s]** Fix typo which breaks k8s example manifest ([#3441](https://github.com/traefik/traefik/pull/3441) by [GeertJohan](https://github.com/GeertJohan))
- **[k8s]** Correct Modifier in Kubernetes Documentation ([#3610](https://github.com/traefik/traefik/pull/3610) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Improve Connection Limit Kubernetes Documentation ([#3711](https://github.com/traefik/traefik/pull/3711) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Add traefik prefix to k8s annotations ([#3682](https://github.com/traefik/traefik/pull/3682) by [zifeo](https://github.com/zifeo))
- **[k8s]** Update kubernetes docs to reflect https options ([#3807](https://github.com/traefik/traefik/pull/3807) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Update kubernetes.md ([#3719](https://github.com/traefik/traefik/pull/3719) by [kmaris](https://github.com/kmaris))
- **[metrics]** Adding grafana dashboards based on prometheus metrics ([#3393](https://github.com/traefik/traefik/pull/3393) by [deimosfr](https://github.com/deimosfr))
- **[middleware,tracing]** Fix missing tracing backend in documentation ([#3706](https://github.com/traefik/traefik/pull/3706) by [mmatur](https://github.com/mmatur))
- **[provider]** Typo in auth labels. ([#3730](https://github.com/traefik/traefik/pull/3730) by [ldez](https://github.com/ldez))
- **[servicefabric]** Fix Service Fabric docs to use v1.6 labels ([#3209](https://github.com/traefik/traefik/pull/3209) by [jjcollinge](https://github.com/jjcollinge))
- **[tracing]** Simple documentation grammar update in tracing ([#3720](https://github.com/traefik/traefik/pull/3720) by [loadstar81](https://github.com/loadstar81))
- Replace unrendered emoji ([#3690](https://github.com/traefik/traefik/pull/3690) by [korigod](https://github.com/korigod))
- Make the &#34;base domain&#34; on all providers ([#3835](https://github.com/traefik/traefik/pull/3835) by [dduportal](https://github.com/dduportal))
- Prepare release v1.7.0-rc5 ([#3902](https://github.com/traefik/traefik/pull/3902) by [dduportal](https://github.com/dduportal))
- Prepare release v1.7.0-rc3 ([#3709](https://github.com/traefik/traefik/pull/3709) by [mmatur](https://github.com/mmatur))
- Prepare release v1.7.0-rc4 ([#3864](https://github.com/traefik/traefik/pull/3864) by [Juliens](https://github.com/Juliens))
- Prepare release v1.7.0-rc2 ([#3632](https://github.com/traefik/traefik/pull/3632) by [nmengin](https://github.com/nmengin))
- Prepare release v1.7.0-rc1 ([#3578](https://github.com/traefik/traefik/pull/3578) by [mmatur](https://github.com/mmatur))

**Misc:**
- **[webui]** Removed non-applicable default tests and fixed custom tests ([#3908](https://github.com/traefik/traefik/pull/3908) by [codecyclist](https://github.com/codecyclist))
- Merge v1.6.6 into v1.7 ([#3802](https://github.com/traefik/traefik/pull/3802) by [ldez](https://github.com/ldez))
- Merge v1.6.5 into v1.7 ([#3595](https://github.com/traefik/traefik/pull/3595) by [ldez](https://github.com/ldez))
- Merge v1.6.4 into master ([#3502](https://github.com/traefik/traefik/pull/3502) by [ldez](https://github.com/ldez))
- Merge v1.6.3 into master ([#3439](https://github.com/traefik/traefik/pull/3439) by [ldez](https://github.com/ldez))
- Merge v1.6.2 into master ([#3367](https://github.com/traefik/traefik/pull/3367) by [ldez](https://github.com/ldez))
- Merge v1.6.1 into master ([#3326](https://github.com/traefik/traefik/pull/3326) by [ldez](https://github.com/ldez))
- Merge v1.6.0 into master ([#3253](https://github.com/traefik/traefik/pull/3253) by [ldez](https://github.com/ldez))
- Merge v1.6.0-rc6 into master ([#3203](https://github.com/traefik/traefik/pull/3203) by [ldez](https://github.com/ldez))
- Merge v1.6.0-rc5 into master  ([#3180](https://github.com/traefik/traefik/pull/3180) by [ldez](https://github.com/ldez))
- Merge v1.6.0-rc4 into master  ([#3129](https://github.com/traefik/traefik/pull/3129) by [ldez](https://github.com/ldez))

## [v1.7.0-rc5](https://github.com/traefik/traefik/tree/v1.7.0-rc5) (2018-09-18)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.0-rc4...v1.7.0-rc5)

**Bug fixes:**
- **[acme]** Ensure only certificates from ACME enabled entrypoint are used ([#3880](https://github.com/traefik/traefik/pull/3880) by [dtomcej](https://github.com/dtomcej))
- **[acme]** Fix LEGO update ([#3895](https://github.com/traefik/traefik/pull/3895) by [ldez](https://github.com/ldez))
- **[acme]** Bump LEGO version ([#3888](https://github.com/traefik/traefik/pull/3888) by [ldez](https://github.com/ldez))
- **[authentication,middleware]** Remove hop-by-hop headers from forward auth response ([#3900](https://github.com/traefik/traefik/pull/3900) by [stffabi](https://github.com/stffabi))
- **[kv]** Add missing quotes around backendName in kv template ([#3885](https://github.com/traefik/traefik/pull/3885) by [NatMarchand](https://github.com/NatMarchand))
- **[logs]** Add interface to Træfik logger ([#3889](https://github.com/traefik/traefik/pull/3889) by [nmengin](https://github.com/nmengin))
- **[middleware,tracing]** Fix tracing duplicated headers ([#3878](https://github.com/traefik/traefik/pull/3878) by [mmatur](https://github.com/mmatur))
- **[middleware]** Fix error pages ([#3894](https://github.com/traefik/traefik/pull/3894) by [Juliens](https://github.com/Juliens))
- **[server]** Avoid panic during stop ([#3898](https://github.com/traefik/traefik/pull/3898) by [nmengin](https://github.com/nmengin))

## [v1.7.0-rc4](https://github.com/traefik/traefik/tree/v1.7.0-rc4) (2018-09-07)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.0-rc3...v1.7.0-rc4)

**Enhancements:**
- **[acme]** Use official Pebble Image. ([#3708](https://github.com/traefik/traefik/pull/3708) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Multiple frontends for consulcatalog ([#3796](https://github.com/traefik/traefik/pull/3796) by [hsmade](https://github.com/hsmade))
- **[ecs]** Add segment support for ECS  ([#3817](https://github.com/traefik/traefik/pull/3817) by [mmatur](https://github.com/mmatur))
- **[k8s]** Remove unnecessary loop ([#3799](https://github.com/traefik/traefik/pull/3799) by [ZloyDyadka](https://github.com/ZloyDyadka))
- **[middleware,consulcatalog,docker,ecs,kv,marathon,mesos,rancher]** Pass the TLS Cert infos in headers ([#3826](https://github.com/traefik/traefik/pull/3826) by [jbdoumenjou](https://github.com/jbdoumenjou))

**Bug fixes:**
- **[acme,cluster]** StoreConfig always initializes the account if it is missing ([#3844](https://github.com/traefik/traefik/pull/3844) by [geraldcroes](https://github.com/geraldcroes))
- **[acme]** Set a keyType to ACME if the account is stored with no KeyType ([#3733](https://github.com/traefik/traefik/pull/3733) by [nmengin](https://github.com/nmengin))
- **[authentication,consulcatalog,docker,ecs,k8s,kv,marathon,mesos,rancher]** Auth Forward with certificates in templates. ([#3804](https://github.com/traefik/traefik/pull/3804) by [ldez](https://github.com/ldez))
- **[k8s]** Prevent unparsable strings from being rendered in the Kubernetes template ([#3753](https://github.com/traefik/traefik/pull/3753) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Don&#39;t merge kubernetes ingresses when priority is set ([#3743](https://github.com/traefik/traefik/pull/3743) by [dtomcej](https://github.com/dtomcej))
- **[kv]** Include missing key in error message for KV store ([#3779](https://github.com/traefik/traefik/pull/3779) by [camelpunch](https://github.com/camelpunch))
- **[metrics]** Avoid a panic during Prometheus registering ([#3717](https://github.com/traefik/traefik/pull/3717) by [nmengin](https://github.com/nmengin))
- **[middleware,websocket]** Enable retry on websocket ([#3825](https://github.com/traefik/traefik/pull/3825) by [Juliens](https://github.com/Juliens))
- **[middleware]** Extend https redirection tests, and fix incorrect behavior ([#3742](https://github.com/traefik/traefik/pull/3742) by [dtomcej](https://github.com/dtomcej))
- **[oxy]** Handle Te header when http2 ([#3824](https://github.com/traefik/traefik/pull/3824) by [Juliens](https://github.com/Juliens))
- **[server]** Avoid goroutine leak in server ([#3851](https://github.com/traefik/traefik/pull/3851) by [nmengin](https://github.com/nmengin))

**Documentation:**
- **[acme]** Fix documentation for route53 acme provider ([#3811](https://github.com/traefik/traefik/pull/3811) by [A-Shleifman](https://github.com/A-Shleifman))
- **[acme]** Update ACME documentation about TLS-ALPN challenge ([#3756](https://github.com/traefik/traefik/pull/3756) by [ldez](https://github.com/ldez))
- **[docker]** Change syntax in quick start guide ([#3726](https://github.com/traefik/traefik/pull/3726) by [trotro](https://github.com/trotro))
- **[docker]** Improve the wording in the documentation for Docker and fix title for Docker User Guide ([#3797](https://github.com/traefik/traefik/pull/3797) by [dduportal](https://github.com/dduportal))
- **[docker]** Typo in docker-and-lets-encrypt.md ([#3724](https://github.com/traefik/traefik/pull/3724) by [A-Shleifman](https://github.com/A-Shleifman))
- **[k8s]** Update kubernetes docs to reflect https options ([#3807](https://github.com/traefik/traefik/pull/3807) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Update kubernetes.md ([#3719](https://github.com/traefik/traefik/pull/3719) by [kmaris](https://github.com/kmaris))
- **[k8s]** Improve Connection Limit Kubernetes Documentation ([#3711](https://github.com/traefik/traefik/pull/3711) by [dtomcej](https://github.com/dtomcej))
- **[provider]** Typo in auth labels. ([#3730](https://github.com/traefik/traefik/pull/3730) by [ldez](https://github.com/ldez))
- **[tracing]** Simple documentation grammar update in tracing ([#3720](https://github.com/traefik/traefik/pull/3720) by [loadstar81](https://github.com/loadstar81))
- Make the &#34;base domain&#34; on all providers ([#3835](https://github.com/traefik/traefik/pull/3835) by [dduportal](https://github.com/dduportal))

**Misc:**
- Merge v1.6.6 into v1.7 ([#3802](https://github.com/traefik/traefik/pull/3802) by [ldez](https://github.com/ldez))

## [v1.6.6](https://github.com/traefik/traefik/tree/v1.6.6) (2018-08-20)
[All Commits](https://github.com/traefik/traefik/compare/v1.6.5...v1.6.6)

**Bug fixes:**
- **[acme]** Avoid duplicated ACME resolution ([#3751](https://github.com/traefik/traefik/pull/3751) by [nmengin](https://github.com/nmengin))
- **[api]** Remove TLS in API ([#3788](https://github.com/traefik/traefik/pull/3788) by [Juliens](https://github.com/Juliens))
- **[cluster]** Remove unusable `--cluster` flag ([#3616](https://github.com/traefik/traefik/pull/3616) by [dtomcej](https://github.com/dtomcej))
- **[ecs]** Fix bad condition in ECS provider ([#3609](https://github.com/traefik/traefik/pull/3609) by [mmatur](https://github.com/mmatur))
- Set keepalive on TCP socket so idleTimeout works ([#3740](https://github.com/traefik/traefik/pull/3740) by [ajardan](https://github.com/ajardan))

**Documentation:**
- A tiny rewording on the documentation API&#39;s page ([#3794](https://github.com/traefik/traefik/pull/3794) by [dduportal](https://github.com/dduportal))
- Adding warnings and solution about the configuration exposure ([#3790](https://github.com/traefik/traefik/pull/3790) by [dduportal](https://github.com/dduportal))
- Fix path to the debug pprof API ([#3608](https://github.com/traefik/traefik/pull/3608) by [multani](https://github.com/multani))

**Misc:**
- **[oxy,websocket]** Update oxy dependency ([#3777](https://github.com/traefik/traefik/pull/3777) by [Juliens](https://github.com/Juliens))

## [v1.7.0-rc3](https://github.com/traefik/traefik/tree/v1.7.0-rc3) (2018-08-01)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.0-rc2...v1.7.0-rc3)

**Enhancements:**
- **[consul,etcd,tls]** Improve TLS integration tests ([#3679](https://github.com/traefik/traefik/pull/3679) by [mmatur](https://github.com/mmatur))
- **[k8s]** Add possibility to set a protocol ([#3648](https://github.com/traefik/traefik/pull/3648) by [SantoDE](https://github.com/SantoDE))

**Bug fixes:**
- **[acme]** Fix acme account deletion without provider change ([#3664](https://github.com/traefik/traefik/pull/3664) by [zyclonite](https://github.com/zyclonite))
- **[acme]** Update lego ([#3659](https://github.com/traefik/traefik/pull/3659) by [mmatur](https://github.com/mmatur))
- **[acme]** Fix ACME certificate for wildcard and root domains ([#3675](https://github.com/traefik/traefik/pull/3675) by [nmengin](https://github.com/nmengin))
- **[api]** Remove TLS in API ([#3665](https://github.com/traefik/traefik/pull/3665) by [mmatur](https://github.com/mmatur))
- **[docker]** Uses both binded HostIP and HostPort when useBindPortIP=true ([#3638](https://github.com/traefik/traefik/pull/3638) by [geraldcroes](https://github.com/geraldcroes))
- **[k8s]** Fix Rewrite-target regex ([#3699](https://github.com/traefik/traefik/pull/3699) by [dtomcej](https://github.com/dtomcej))
- **[middleware]** Correct Entrypoint Redirect with Stripped or Added Path ([#3631](https://github.com/traefik/traefik/pull/3631) by [dtomcej](https://github.com/dtomcej))
- **[tracing]** Added default configuration for Datadog APM Tracer ([#3655](https://github.com/traefik/traefik/pull/3655) by [aantono](https://github.com/aantono))
- **[tracing]** Added support for Trace name truncation for traces ([#3689](https://github.com/traefik/traefik/pull/3689) by [aantono](https://github.com/aantono))
- **[websocket]** Handle shutdown of Hijacked connections ([#3636](https://github.com/traefik/traefik/pull/3636) by [Juliens](https://github.com/Juliens))
- H2C: Remove buggy line in init to make verbose switch working ([#3701](https://github.com/traefik/traefik/pull/3701) by [dduportal](https://github.com/dduportal))
- Updating oxy dependency ([#3700](https://github.com/traefik/traefik/pull/3700) by [crholm](https://github.com/crholm))

**Documentation:**
- **[acme]** Update Namecheap status ([#3604](https://github.com/traefik/traefik/pull/3604) by [stoinov](https://github.com/stoinov))
- **[acme]** Fix some DNS provider link ([#3639](https://github.com/traefik/traefik/pull/3639) by [ldez](https://github.com/ldez))
- **[docker]** Fix style in examples/quickstart ([#3705](https://github.com/traefik/traefik/pull/3705) by [korigod](https://github.com/korigod))
- **[k8s]** Add traefik prefix to k8s annotations ([#3682](https://github.com/traefik/traefik/pull/3682) by [zifeo](https://github.com/zifeo))
- **[middleware,tracing]** Fix missing tracing backend in documentation ([#3706](https://github.com/traefik/traefik/pull/3706) by [mmatur](https://github.com/mmatur))
- Replace unrendered emoji ([#3690](https://github.com/traefik/traefik/pull/3690) by [korigod](https://github.com/korigod))

## [v1.7.0-rc2](https://github.com/traefik/traefik/tree/v1.7.0-rc2) (2018-07-17)
[All Commits](https://github.com/traefik/traefik/compare/v1.7.0-rc1...v1.7.0-rc2)


**Bug fixes:**
- **[acme,provider]** Create init method on provider interface ([#3580](https://github.com/traefik/traefik/pull/3580) by [Juliens](https://github.com/Juliens))
- **[acme]** Serve TLS-Challenge certificate in first ([#3605](https://github.com/traefik/traefik/pull/3605) by [nmengin](https://github.com/nmengin))
- **[api,authentication,webui]** Auth section in web UI.  ([#3628](https://github.com/traefik/traefik/pull/3628) by [ldez](https://github.com/ldez))
- **[authentication,middleware,provider]** Don&#39;t pass the Authorization header to the backends ([#3606](https://github.com/traefik/traefik/pull/3606) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[ecs]** Fix 400 bad request on  AWS ECS API ([#3629](https://github.com/traefik/traefik/pull/3629) by [mmatur](https://github.com/mmatur))
- **[k8s]** Fix rewrite-target Annotation behavior ([#3582](https://github.com/traefik/traefik/pull/3582) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Correct App-Root kubernetes behavior ([#3592](https://github.com/traefik/traefik/pull/3592) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Add more K8s Unit Tests ([#3583](https://github.com/traefik/traefik/pull/3583) by [dtomcej](https://github.com/dtomcej))
- **[kv]** KV and authentication ([#3615](https://github.com/traefik/traefik/pull/3615) by [ldez](https://github.com/ldez))
- **[middleware]** Send &#39;Retry-After&#39; to comply with RFC6585. ([#3593](https://github.com/traefik/traefik/pull/3593) by [ldez](https://github.com/ldez))

**Documentation:**
- **[k8s]** Correct Modifier in Kubernetes Documentation ([#3610](https://github.com/traefik/traefik/pull/3610) by [dtomcej](https://github.com/dtomcej))

**Misc:**
- Merge v1.6.5 into v1.7 ([#3595](https://github.com/traefik/traefik/pull/3595) by [ldez](https://github.com/ldez))

## [v1.6.5](https://github.com/traefik/traefik/tree/v1.6.5) (2018-07-09)
[All Commits](https://github.com/traefik/traefik/compare/v1.6.4...v1.6.5)

**Bug fixes:**
- **[acme]** Add a mutex on local store for HTTPChallenges ([#3579](https://github.com/traefik/traefik/pull/3579) by [Juliens](https://github.com/Juliens))
- **[consulcatalog]** Split the error handling from Consul Catalog (deadlock) ([#3560](https://github.com/traefik/traefik/pull/3560) by [ortz](https://github.com/ortz))
- **[docker]** segment labels: multiple frontends for one backend. ([#3511](https://github.com/traefik/traefik/pull/3511) by [ldez](https://github.com/ldez))
- **[kv]** Better support on same prefix at the same level in the KV ([#3532](https://github.com/traefik/traefik/pull/3532) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[logs]** Add logs when error is generated in error handler ([#3567](https://github.com/traefik/traefik/pull/3567) by [Juliens](https://github.com/Juliens))
- **[middleware]** Create middleware to be able to handle HTTP pipelining correctly ([#3513](https://github.com/traefik/traefik/pull/3513) by [mmatur](https://github.com/mmatur))

**Documentation:**
- **[acme]** The gandiv5 provider works with wildcard ([#3506](https://github.com/traefik/traefik/pull/3506) by [manu5801](https://github.com/manu5801))
- **[kv]** Update keyFile first/last line comment in kv-config.md ([#3558](https://github.com/traefik/traefik/pull/3558) by [madnight](https://github.com/madnight))
- Minor formatting issue in user-guide ([#3546](https://github.com/traefik/traefik/pull/3546) by [Vanuan](https://github.com/Vanuan))

## [v1.7.0-rc1](https://github.com/traefik/traefik/tree/v1.7.0-rc1) (2018-07-09)
[All Commits](https://github.com/traefik/traefik/compare/v1.6.0-rc1...v1.7.0-rc1)

**Enhancements:**
- **[acme]** Simplify get acme client ([#3499](https://github.com/traefik/traefik/pull/3499) by [ldez](https://github.com/ldez))
- **[acme]** Simplify acme e2e tests. ([#3534](https://github.com/traefik/traefik/pull/3534) by [ldez](https://github.com/ldez))
- **[acme]** Add option to select algorithm to generate ACME certificates ([#3319](https://github.com/traefik/traefik/pull/3319) by [mmatur](https://github.com/mmatur))
- **[acme]** Enable to override certificates in key-value store when using storeconfig ([#3202](https://github.com/traefik/traefik/pull/3202) by [thomasjpfan](https://github.com/thomasjpfan))
- **[acme]** ACME TLS ALPN ([#3553](https://github.com/traefik/traefik/pull/3553) by [ldez](https://github.com/ldez))
- **[acme]** Remove acme provider dependency in server ([#3225](https://github.com/traefik/traefik/pull/3225) by [Juliens](https://github.com/Juliens))
- **[api,cluster]** Improved cluster api to include the current leader node ([#3100](https://github.com/traefik/traefik/pull/3100) by [aantono](https://github.com/aantono))
- **[authentication,k8s]** Auth support in frontends for k8s and file ([#3460](https://github.com/traefik/traefik/pull/3460) by [Zatte](https://github.com/Zatte))
- **[authentication,middleware]** Add xforwarded method ([#3424](https://github.com/traefik/traefik/pull/3424) by [erik-sjoestedt](https://github.com/erik-sjoestedt))
- **[authentication,middleware]** Forward auth headers ([#3521](https://github.com/traefik/traefik/pull/3521) by [hwhelan-CB](https://github.com/hwhelan-CB))
- **[consul,consulcatalog,docker,ecs,kv,marathon,mesos,rancher]** Auth support in frontends ([#3559](https://github.com/traefik/traefik/pull/3559) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[consulcatalog,docker,ecs,file,k8s,kv,marathon,mesos,rancher]** Add SSLForceHost support. ([#3246](https://github.com/traefik/traefik/pull/3246) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Add support for stale reads from Consul catalog ([#3523](https://github.com/traefik/traefik/pull/3523) by [marenzo](https://github.com/marenzo))
- **[docker]** Add a default value for the docker.network configuration ([#3471](https://github.com/traefik/traefik/pull/3471) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[ecs]** Support for AWS ECS Fargate ([#3379](https://github.com/traefik/traefik/pull/3379) by [mmatur](https://github.com/mmatur))
- **[ecs]** Add support for ECS constraints ([#3537](https://github.com/traefik/traefik/pull/3537) by [andrewstucki](https://github.com/andrewstucki))
- **[ecs]** Support `traefik.backend` for ECS ([#3510](https://github.com/traefik/traefik/pull/3510) by [hwhelan-CB](https://github.com/hwhelan-CB))
- **[ecs]** Allow binding ECS container port ([#3533](https://github.com/traefik/traefik/pull/3533) by [andrewstucki](https://github.com/andrewstucki))
- **[healthcheck,consul,consulcatalog,docker,ecs,kv,marathon,mesos,rancher]** Override health check scheme ([#3315](https://github.com/traefik/traefik/pull/3315) by [ldez](https://github.com/ldez))
- **[healthcheck]** Support 3xx HTTP status codes for health check ([#3364](https://github.com/traefik/traefik/pull/3364) by [SniperCZE](https://github.com/SniperCZE))
- **[healthcheck]** Support all 2xx HTTP status code for health check. ([#3362](https://github.com/traefik/traefik/pull/3362) by [ldez](https://github.com/ldez))
- **[healthcheck]** Add HTTP headers to healthcheck. ([#3047](https://github.com/traefik/traefik/pull/3047) by [zetaab](https://github.com/zetaab))
- **[k8s]** Add more k8s tests ([#3491](https://github.com/traefik/traefik/pull/3491) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Substitute hardcoded &#34;&lt;namespace&gt;/&lt;name&gt;&#34; with k8s ListerGetter ([#3470](https://github.com/traefik/traefik/pull/3470) by [yue9944882](https://github.com/yue9944882))
- **[k8s]** Custom frontend name for test helper ([#3444](https://github.com/traefik/traefik/pull/3444) by [ldez](https://github.com/ldez))
- **[k8s]** Add annotation to allow modifiers to be used properly in kubernetes ([#3481](https://github.com/traefik/traefik/pull/3481) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Create Global Backend Ingress ([#3404](https://github.com/traefik/traefik/pull/3404) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Specify backend servers&#39; weight via annotation for kubernetes ([#3112](https://github.com/traefik/traefik/pull/3112) by [yue9944882](https://github.com/yue9944882))
- **[k8s]** Support multi-port services. ([#3121](https://github.com/traefik/traefik/pull/3121) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Mapping ExternalNames to custom ports ([#3231](https://github.com/traefik/traefik/pull/3231) by [gildas](https://github.com/gildas))
- **[k8s]** Allow any kubernetes ingressClass value ([#3516](https://github.com/traefik/traefik/pull/3516) by [rtreffer](https://github.com/rtreffer))
- **[k8s]** Enable Ingress Status updates ([#3324](https://github.com/traefik/traefik/pull/3324) by [dtomcej](https://github.com/dtomcej))
- **[kv]** Use index-based syntax in KV tests. ([#3352](https://github.com/traefik/traefik/pull/3352) by [ldez](https://github.com/ldez))
- **[logs,middleware]** Make accesslogs.logTheRoundTrip async to get lost performance ([#3152](https://github.com/traefik/traefik/pull/3152) by [ryarnyah](https://github.com/ryarnyah))
- **[logs,middleware]** Added duration filter for logs ([#3463](https://github.com/traefik/traefik/pull/3463) by [rodrigodiez](https://github.com/rodrigodiez))
- **[marathon]** Adding compatibility for marathon 1.5 ([#3505](https://github.com/traefik/traefik/pull/3505) by [TrevinTeacutter](https://github.com/TrevinTeacutter))
- **[marathon]** Sane default and configurable Marathon request timeouts ([#3286](https://github.com/traefik/traefik/pull/3286) by [marco-jantke](https://github.com/marco-jantke))
- **[mesos]** Segments Labels: Mesos ([#3383](https://github.com/traefik/traefik/pull/3383) by [drewkerrigan](https://github.com/drewkerrigan))
- **[metrics]** Metrics: Add support for InfluxDB Database / RetentionPolicy and HTTP client ([#3391](https://github.com/traefik/traefik/pull/3391) by [drewkerrigan](https://github.com/drewkerrigan))
- **[middleware,server]** Extreme Makeover: server refactoring ([#3461](https://github.com/traefik/traefik/pull/3461) by [ldez](https://github.com/ldez))
- **[middleware,tracing]** Added integration support for Datadog APM Tracing ([#3517](https://github.com/traefik/traefik/pull/3517) by [aantono](https://github.com/aantono))
- **[middleware,tracing]** Create a custom logger for jaeger ([#3541](https://github.com/traefik/traefik/pull/3541) by [mmatur](https://github.com/mmatur))
- **[middleware]** Performance enhancements for the rules matchers. ([#3563](https://github.com/traefik/traefik/pull/3563) by [ShaneSaww](https://github.com/ShaneSaww))
- **[middleware]** Extract internal router creation from server ([#3204](https://github.com/traefik/traefik/pull/3204) by [Juliens](https://github.com/Juliens))
- **[rules]** CNAME flattening ([#3403](https://github.com/traefik/traefik/pull/3403) by [gamalan](https://github.com/gamalan))
- **[servicefabric]** Add white list for Service Fabric ([#3079](https://github.com/traefik/traefik/pull/3079) by [ldez](https://github.com/ldez))
- **[servicefabric]** Add HTTP headers to healthcheck. ([#3205](https://github.com/traefik/traefik/pull/3205) by [ldez](https://github.com/ldez))
- **[tls]** Improve TLS Handshake ([#3512](https://github.com/traefik/traefik/pull/3512) by [dtomcej](https://github.com/dtomcej))
- **[tls]** Support TLS MinVersion and CipherSuite as CLI option. ([#3107](https://github.com/traefik/traefik/pull/3107) by [ldez](https://github.com/ldez))
- **[webui]** Add some missing elements in the WebUI ([#3327](https://github.com/traefik/traefik/pull/3327) by [ldez](https://github.com/ldez))
- Minor changes ([#3554](https://github.com/traefik/traefik/pull/3554) by [ldez](https://github.com/ldez))
- h2c server ([#3387](https://github.com/traefik/traefik/pull/3387) by [Juliens](https://github.com/Juliens))
- Fix backend reuse ([#3312](https://github.com/traefik/traefik/pull/3312) by [arnested](https://github.com/arnested))
- Call functions to enable block/mutex pprof profiles. ([#3564](https://github.com/traefik/traefik/pull/3564) by [timoreimann](https://github.com/timoreimann))
- Implement h2c with backend ([#3371](https://github.com/traefik/traefik/pull/3371) by [Juliens](https://github.com/Juliens))
- Upgrade GRPC dependencies ([#3342](https://github.com/traefik/traefik/pull/3342) by [gottwald](https://github.com/gottwald))
- Generated assets file are only mandatory in main ([#3386](https://github.com/traefik/traefik/pull/3386) by [Juliens](https://github.com/Juliens))

**Bug fixes:**
- **[acme]** Does not generate ACME certificate if domain is checked by dynamic certificate ([#3238](https://github.com/traefik/traefik/pull/3238) by [Juliens](https://github.com/Juliens))
- **[k8s]** Fix panic setting ingress status ([#3492](https://github.com/traefik/traefik/pull/3492) by [dtomcej](https://github.com/dtomcej))
- **[logs]** Add logs when error is generated in error handler ([#3571](https://github.com/traefik/traefik/pull/3571) by [Juliens](https://github.com/Juliens))
- **[middleware]** Avoid retries when any data was written to the backend ([#3285](https://github.com/traefik/traefik/pull/3285) by [marco-jantke](https://github.com/marco-jantke))

**Documentation:**
- **[k8s]** Add a k8s guide section on traffic splitting via service weights. ([#3556](https://github.com/traefik/traefik/pull/3556) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Change code block of traefik-web-ui to match file ([#3542](https://github.com/traefik/traefik/pull/3542) by [drewgwallace](https://github.com/drewgwallace))
- **[k8s]** Fix typo which breaks k8s example manifest ([#3441](https://github.com/traefik/traefik/pull/3441) by [GeertJohan](https://github.com/GeertJohan))
- **[metrics]** Adding grafana dashboards based on prometheus metrics ([#3393](https://github.com/traefik/traefik/pull/3393) by [deimosfr](https://github.com/deimosfr))
- **[servicefabric]** Fix Service Fabric docs to use v1.6 labels ([#3209](https://github.com/traefik/traefik/pull/3209) by [jjcollinge](https://github.com/jjcollinge))

**Misc:**
- Merge v1.6.4 into master ([#3502](https://github.com/traefik/traefik/pull/3502) by [ldez](https://github.com/ldez))
- Merge v1.6.3 into master ([#3439](https://github.com/traefik/traefik/pull/3439) by [ldez](https://github.com/ldez))
- Merge v1.6.2 into master ([#3367](https://github.com/traefik/traefik/pull/3367) by [ldez](https://github.com/ldez))
- Merge v1.6.1 into master ([#3326](https://github.com/traefik/traefik/pull/3326) by [ldez](https://github.com/ldez))
- Merge v1.6.0 into master ([#3253](https://github.com/traefik/traefik/pull/3253) by [ldez](https://github.com/ldez))
- Merge v1.6.0-rc6 into master ([#3203](https://github.com/traefik/traefik/pull/3203) by [ldez](https://github.com/ldez))
- Merge v1.6.0-rc5 into master  ([#3180](https://github.com/traefik/traefik/pull/3180) by [ldez](https://github.com/ldez))
- Merge v1.6.0-rc4 into master  ([#3129](https://github.com/traefik/traefik/pull/3129) by [ldez](https://github.com/ldez))

## [v1.6.4](https://github.com/traefik/traefik/tree/v1.6.4) (2018-06-15)
[All Commits](https://github.com/traefik/traefik/compare/v1.6.3...v1.6.4)

**Bug fixes:**
- **[acme]** Use logrus writer instead of os.Stderr ([#3498](https://github.com/traefik/traefik/pull/3498) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Enclose IPv6 addresses in &#34;[]&#34; ([#3477](https://github.com/traefik/traefik/pull/3477) by [herver](https://github.com/herver))
- **[docker,ecs,marathon,mesos,rancher]** Use net.JoinHostPort for servers URL ([#3484](https://github.com/traefik/traefik/pull/3484) by [ldez](https://github.com/ldez))
- **[docker]** Backend name with docker-compose and segments. ([#3485](https://github.com/traefik/traefik/pull/3485) by [ldez](https://github.com/ldez))
- **[oxy]** Handle buffer pool for oxy ([#3450](https://github.com/traefik/traefik/pull/3450) by [Juliens](https://github.com/Juliens))

**Documentation:**
- **[acme]** The exoscale provider works with wildcard ([#3479](https://github.com/traefik/traefik/pull/3479) by [greut](https://github.com/greut))
- **[consul,docker]** Edit wording ([#3438](https://github.com/traefik/traefik/pull/3438) by [mayank23](https://github.com/mayank23))
- **[k8s]** Add missing annotation documentation. ([#3454](https://github.com/traefik/traefik/pull/3454) by [ldez](https://github.com/ldez))
- **[kv]** Fix typo in kv user guide ([#3474](https://github.com/traefik/traefik/pull/3474) by [shambarick](https://github.com/shambarick))
- Clean metrics documentation. ([#3488](https://github.com/traefik/traefik/pull/3488) by [ldez](https://github.com/ldez))

## [v1.6.3](https://github.com/traefik/traefik/tree/v1.6.3) (2018-06-05)
[All Commits](https://github.com/traefik/traefik/compare/v1.6.2...v1.6.3)

**Enhancements:**
- **[acme]** Add user agent for ACME ([#3431](https://github.com/traefik/traefik/pull/3431) by [ldez](https://github.com/ldez))
- **[acme]** Use to the stable version of Lego ([#3418](https://github.com/traefik/traefik/pull/3418) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[acme,cluster]** Improve ACME account registration URI management ([#3398](https://github.com/traefik/traefik/pull/3398) by [nmengin](https://github.com/nmengin))
- **[acme,cluster]** Remove ACME empty certificates from KV store ([#3389](https://github.com/traefik/traefik/pull/3389) by [nmengin](https://github.com/nmengin))
- **[consulcatalog]** Reflect changes in catalog healthy nodes in healthCheck watch ([#3390](https://github.com/traefik/traefik/pull/3390) by [thebinary](https://github.com/thebinary))
- **[consulcatalog]** Detect change when service or node are in maintenance mode ([#3434](https://github.com/traefik/traefik/pull/3434) by [mmatur](https://github.com/mmatur))
- **[k8s]** Update Kubernetes provider to support IPv6 Backends ([#3432](https://github.com/traefik/traefik/pull/3432) by [dtomcej](https://github.com/dtomcej))
- **[logs,middleware]** Add URL and Host for some access logs. ([#3430](https://github.com/traefik/traefik/pull/3430) by [ldez](https://github.com/ldez))
- **[metrics]** Improve Prometheus metrics removal ([#3287](https://github.com/traefik/traefik/pull/3287) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** Whitelist and XFF. ([#3411](https://github.com/traefik/traefik/pull/3411) by [ldez](https://github.com/ldez))
- **[middleware]** Error pages and header merge ([#3394](https://github.com/traefik/traefik/pull/3394) by [ldez](https://github.com/ldez))
- **[websocket]** Includes the headers in the HTTP response of a websocket request ([#3425](https://github.com/traefik/traefik/pull/3425) by [geraldcroes](https://github.com/geraldcroes))
- **[webui]** Webui Whitelist overflow. ([#3412](https://github.com/traefik/traefik/pull/3412) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme]** Docs: ACME Overhaul ([#3421](https://github.com/traefik/traefik/pull/3421) by [Dargmuesli](https://github.com/Dargmuesli))
- **[acme]** Minor documentation changes ([#3405](https://github.com/traefik/traefik/pull/3405) by [amincheloh](https://github.com/amincheloh))
- **[k8s]** Helm installation using values ([#3392](https://github.com/traefik/traefik/pull/3392) by [erikaulin](https://github.com/erikaulin))
- **[k8s]** Update Kubernetes Port Documentation ([#3368](https://github.com/traefik/traefik/pull/3368) by [dtomcej](https://github.com/dtomcej))

## [v1.6.2](https://github.com/traefik/traefik/tree/v1.6.2) (2018-05-22)
[All Commits](https://github.com/traefik/traefik/compare/v1.6.1...v1.6.2)

**Bug fixes:**
- **[acme]** fix: acme errors management. ([#3329](https://github.com/traefik/traefik/pull/3329) by [ldez](https://github.com/ldez))
- **[acme]** Force to use ACME v02 endpoint. ([#3358](https://github.com/traefik/traefik/pull/3358) by [ldez](https://github.com/ldez))
- **[file]** No template parsing on traefik configuration file ([#3347](https://github.com/traefik/traefik/pull/3347) by [Juliens](https://github.com/Juliens))
- **[k8s]** Add redirect-permanent to kubernetes template ([#3332](https://github.com/traefik/traefik/pull/3332) by [dtomcej](https://github.com/dtomcej))
- **[logs]** Enhance Load-balancing method validation log. ([#3361](https://github.com/traefik/traefik/pull/3361) by [ldez](https://github.com/ldez))
- **[middleware]** Fix error pages content.  ([#3337](https://github.com/traefik/traefik/pull/3337) by [ldez](https://github.com/ldez))
- **[webui]** Route rules overlaps in UI ([#3333](https://github.com/traefik/traefik/pull/3333) by [ldez](https://github.com/ldez))
- **[webui]** WebUI typo into the buffering section. ([#3363](https://github.com/traefik/traefik/pull/3363) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme]** Update caServer to letsencrypt one in examples ([#3339](https://github.com/traefik/traefik/pull/3339) by [woernfl](https://github.com/woernfl))
- **[docker]** Add command for basic auth with Docker Compose ([#3346](https://github.com/traefik/traefik/pull/3346) by [DeamonMV](https://github.com/DeamonMV))
- **[docker]** Removes ambiguity with the word &#39;default&#39; ([#3344](https://github.com/traefik/traefik/pull/3344) by [ldez](https://github.com/ldez))
- **[kv]** Add basicAuth example for KV ([#3274](https://github.com/traefik/traefik/pull/3274) by [MichaelErmer](https://github.com/MichaelErmer))
- **[provider]** Update docs to reflect Provider wording ([#3331](https://github.com/traefik/traefik/pull/3331) by [dtomcej](https://github.com/dtomcej))
- **[servicefabric]** Update docs to match SF provider labels ([#3335](https://github.com/traefik/traefik/pull/3335) by [jjcollinge](https://github.com/jjcollinge))

## [v1.6.1](https://github.com/traefik/traefik/tree/v1.6.1) (2018-05-14)
[All Commits](https://github.com/traefik/traefik/compare/v1.6.0...v1.6.1)

**Bug fixes:**
- **[acme]** Add missing deprecation info in CLI help. ([#3291](https://github.com/traefik/traefik/pull/3291) by [ldez](https://github.com/ldez))
- **[docker,marathon,rancher]** Fix segment backend name ([#3317](https://github.com/traefik/traefik/pull/3317) by [ldez](https://github.com/ldez))
- **[logs,middleware]** Error when accesslog and error pages ([#3314](https://github.com/traefik/traefik/pull/3314) by [ldez](https://github.com/ldez))
- **[middleware,tracing]** Fix wrong tag in forward span in tracing middleware ([#3279](https://github.com/traefik/traefik/pull/3279) by [mmatur](https://github.com/mmatur))
- **[webui]** Fix webui ([#3299](https://github.com/traefik/traefik/pull/3299) by [ldez](https://github.com/ldez))

**Documentation:**
- **[k8s]** Add Documentation update for Kubernetes Ingress ([#3294](https://github.com/traefik/traefik/pull/3294) by [dtomcej](https://github.com/dtomcej))
- **[tls]** Enhance entry point TLS CLI reference. ([#3290](https://github.com/traefik/traefik/pull/3290) by [ldez](https://github.com/ldez))
- Typo in documentation ([#3261](https://github.com/traefik/traefik/pull/3261) by [blakethepatton](https://github.com/blakethepatton))

## [v1.6.0](https://github.com/traefik/traefik/tree/v1.6.0) (2018-04-30)
[Commits](https://github.com/traefik/traefik/compare/v1.5.0-rc1...v1.6.0)
[Commits pre RC](https://github.com/traefik/traefik/compare/v1.5.0-rc1...v1.6.0-rc1)

**Enhancements:**
- **[acme]** Create ACME Provider ([#2889](https://github.com/traefik/traefik/pull/2889) by [nmengin](https://github.com/nmengin))
- **[acme]** Update Lego (Gandi API v5, cloudxns, ...) ([#2844](https://github.com/traefik/traefik/pull/2844) by [ldez](https://github.com/ldez))
- **[acme]** Simplify storing renewed acme certificate ([#2614](https://github.com/traefik/traefik/pull/2614) by [ferhatelmas](https://github.com/ferhatelmas))
- **[acme]** ACME V2 Integration ([#3063](https://github.com/traefik/traefik/pull/3063) by [nmengin](https://github.com/nmengin))
- **[acme]** Bump Lego Version for GoDaddy DNS Provider ([#2482](https://github.com/traefik/traefik/pull/2482) by [sjawhar](https://github.com/sjawhar))
- **[acme]** Delete TLS-SNI-01 challenge from ACME ([#2971](https://github.com/traefik/traefik/pull/2971) by [nmengin](https://github.com/nmengin))
- **[acme]** Create backup file during migration from ACME V1 to ACME V2 ([#3191](https://github.com/traefik/traefik/pull/3191) by [nmengin](https://github.com/nmengin))
- **[acme]** Generate wildcard certificate with SANs in ACME ([#3167](https://github.com/traefik/traefik/pull/3167) by [nmengin](https://github.com/nmengin))
- **[api,cluster]** Added cluster/leader endpoint ([#3009](https://github.com/traefik/traefik/pull/3009) by [aantono](https://github.com/aantono))
- **[authentication]** Forward Authentication: add X-Forwarded-Uri ([#2398](https://github.com/traefik/traefik/pull/2398) by [sebastianbauer](https://github.com/sebastianbauer))
- **[boltdb,consul,etcd,kv,zk]** Add all available configuration to KV Backend ([#2652](https://github.com/traefik/traefik/pull/2652) by [ldez](https://github.com/ldez))
- **[boltdb,consul,etcd,kv,zk]** homogenization of templates: KV ([#2661](https://github.com/traefik/traefik/pull/2661) by [ldez](https://github.com/ldez))
- **[boltdb,consul,etcd,kv,zk]** Homogenization of the providers (part 1):  KV ([#2616](https://github.com/traefik/traefik/pull/2616) by [ldez](https://github.com/ldez))
- **[consul,consulcatalog]** Homogenization of templates: Consul Catalog ([#2668](https://github.com/traefik/traefik/pull/2668) by [ldez](https://github.com/ldez))
- **[consul,consulcatalog]** Split consul and consul catalog. ([#2655](https://github.com/traefik/traefik/pull/2655) by [ldez](https://github.com/ldez))
- **[consulcatalog,ecs,mesos]** Factorize labels managements. ([#3099](https://github.com/traefik/traefik/pull/3099) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Check for endpoints while detecting Consul service changes ([#2882](https://github.com/traefik/traefik/pull/2882) by [caseycs](https://github.com/caseycs))
- **[consulcatalog]** TLS Support for ConsulCatalog ([#2900](https://github.com/traefik/traefik/pull/2900) by [mmatur](https://github.com/mmatur))
- **[consulcatalog]** Add all available tags to Consul Catalog Backend ([#2646](https://github.com/traefik/traefik/pull/2646) by [ldez](https://github.com/ldez))
- **[docker,docker/swarm]** Fix support for macvlan driver in docker provider ([#2827](https://github.com/traefik/traefik/pull/2827) by [mmatur](https://github.com/mmatur))
- **[docker,marathon,rancher]** Segments Labels: Rancher &amp; Marathon ([#3073](https://github.com/traefik/traefik/pull/3073) by [ldez](https://github.com/ldez))
- **[docker]** Add all available labels to Docker Backend ([#2584](https://github.com/traefik/traefik/pull/2584) by [ldez](https://github.com/ldez))
- **[docker]** Homogenization of templates: Docker ([#2659](https://github.com/traefik/traefik/pull/2659) by [ldez](https://github.com/ldez))
- **[docker]** Custom headers by service labels for docker backends ([#2514](https://github.com/traefik/traefik/pull/2514) by [Tiscs](https://github.com/Tiscs))
- **[docker]** Segment labels: Docker ([#3055](https://github.com/traefik/traefik/pull/3055) by [ldez](https://github.com/ldez))
- **[dynamodb,ecs]** Upgrade AWS SKD to version v1.13.1 ([#2908](https://github.com/traefik/traefik/pull/2908) by [mmatur](https://github.com/mmatur))
- **[ecs]** Add all available labels to ECS Backend ([#2605](https://github.com/traefik/traefik/pull/2605) by [ldez](https://github.com/ldez))
- **[ecs]** Homogenization of templates: ECS ([#2663](https://github.com/traefik/traefik/pull/2663) by [ldez](https://github.com/ldez))
- **[ecs]** Factorize labels managements. ([#3159](https://github.com/traefik/traefik/pull/3159) by [ldez](https://github.com/ldez))
- **[eureka]** Homogenization of templates: Eureka  ([#2846](https://github.com/traefik/traefik/pull/2846) by [ldez](https://github.com/ldez))
- **[eureka]** Replace Delay by RefreshSecond in Eureka ([#2972](https://github.com/traefik/traefik/pull/2972) by [ldez](https://github.com/ldez))
- **[file]** Added support for templates to file provider ([#2991](https://github.com/traefik/traefik/pull/2991) by [aantono](https://github.com/aantono))
- **[healthcheck]** Toggle /ping to artificially return unhealthy response on SIGTERM during requestAcceptGraceTimeout interval ([#3062](https://github.com/traefik/traefik/pull/3062) by [ravilr](https://github.com/ravilr))
- **[healthcheck]** Improve logging output for failing healthchecks ([#2443](https://github.com/traefik/traefik/pull/2443) by [marco-jantke](https://github.com/marco-jantke))
- **[k8s,tls]** Add support for fetching k8s Ingress TLS data from secrets ([#2439](https://github.com/traefik/traefik/pull/2439) by [gopenguin](https://github.com/gopenguin))
- **[k8s]** Introduce k8s informer factory ([#2867](https://github.com/traefik/traefik/pull/2867) by [yue9944882](https://github.com/yue9944882))
- **[k8s]** Add all available annotations to k8s Backend ([#2612](https://github.com/traefik/traefik/pull/2612) by [ldez](https://github.com/ldez))
- **[k8s]** Bump kubernetes/client-go ([#2848](https://github.com/traefik/traefik/pull/2848) by [yue9944882](https://github.com/yue9944882))
- **[k8s]** Add app-root annotation support for kubernetes ingress ([#2522](https://github.com/traefik/traefik/pull/2522) by [yue9944882](https://github.com/yue9944882))
- **[k8s]** Builders in k8s tests ([#2513](https://github.com/traefik/traefik/pull/2513) by [ldez](https://github.com/ldez))
- **[k8s]** Allow custom value for kubernetes.io/ingress.class annotation ([#2222](https://github.com/traefik/traefik/pull/2222) by [yuvipanda](https://github.com/yuvipanda))
- **[logs,middleware]** Add access log filter for retry attempts ([#3042](https://github.com/traefik/traefik/pull/3042) by [marco-jantke](https://github.com/marco-jantke))
- **[logs,middleware]** Add username in accesslog ([#2111](https://github.com/traefik/traefik/pull/2111) by [bastiaanb](https://github.com/bastiaanb))
- **[logs,middleware]** Ultimate Access log filter ([#2988](https://github.com/traefik/traefik/pull/2988) by [mmatur](https://github.com/mmatur))
- **[logs]** Allow overriding the log level in debug mode. ([#3050](https://github.com/traefik/traefik/pull/3050) by [timoreimann](https://github.com/timoreimann))
- **[logs]** Display file log when test fails. ([#2801](https://github.com/traefik/traefik/pull/2801) by [ldez](https://github.com/ldez))
- **[marathon]** Remove health check filter from Marathon tasks. ([#2817](https://github.com/traefik/traefik/pull/2817) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Add all available labels to Marathon Backend ([#2602](https://github.com/traefik/traefik/pull/2602) by [ldez](https://github.com/ldez))
- **[marathon]** homogenization of templates: Marathon ([#2665](https://github.com/traefik/traefik/pull/2665) by [ldez](https://github.com/ldez))
- **[mesos]** Add all available labels to Mesos Backend  ([#2687](https://github.com/traefik/traefik/pull/2687) by [ldez](https://github.com/ldez))
- **[metrics]** Added entrypoint metrics to influxdb ([#2992](https://github.com/traefik/traefik/pull/2992) by [adityacs](https://github.com/adityacs))
- **[metrics]** Remove unnecessary conversion ([#2850](https://github.com/traefik/traefik/pull/2850) by [ferhatelmas](https://github.com/ferhatelmas))
- **[metrics]** Extend metrics and rebuild prometheus exporting logic ([#2567](https://github.com/traefik/traefik/pull/2567) by [marco-jantke](https://github.com/marco-jantke))
- **[metrics]** Added missing metrics to registry for Datadog and StatsD ([#2890](https://github.com/traefik/traefik/pull/2890) by [aantono](https://github.com/aantono))
- **[middleware,consul,consulcatalog,docker,ecs,k8s,marathon,mesos,rancher]** New option in secure middleware ([#2958](https://github.com/traefik/traefik/pull/2958) by [mmatur](https://github.com/mmatur))
- **[middleware,consulcatalog,docker,ecs,k8s,kv,marathon,mesos,rancher]** Ability to use &#34;X-Forwarded-For&#34; as a source of IP for white list. ([#3070](https://github.com/traefik/traefik/pull/3070) by [ldez](https://github.com/ldez))
- **[middleware,docker]** Use pointer of error pages ([#2607](https://github.com/traefik/traefik/pull/2607) by [ldez](https://github.com/ldez))
- **[middleware,provider]** Redirection: permanent move option. ([#2774](https://github.com/traefik/traefik/pull/2774) by [ldez](https://github.com/ldez))
- **[middleware]** Add tests on IPWhiteLister. ([#3106](https://github.com/traefik/traefik/pull/3106) by [ldez](https://github.com/ldez))
- **[middleware]** Change port of traefik for error pages integration test ([#2907](https://github.com/traefik/traefik/pull/2907) by [mmatur](https://github.com/mmatur))
- **[middleware]** Remove unnecessary returns in tracing setup ([#2880](https://github.com/traefik/traefik/pull/2880) by [ferhatelmas](https://github.com/ferhatelmas))
- **[middleware]** Request buffering middleware ([#2217](https://github.com/traefik/traefik/pull/2217) by [harnash](https://github.com/harnash))
- **[middleware]** Add new options to the CLI entrypoint definition.  ([#2799](https://github.com/traefik/traefik/pull/2799) by [ldez](https://github.com/ldez))
- **[provider]** No error pages must return nil. ([#2610](https://github.com/traefik/traefik/pull/2610) by [ldez](https://github.com/ldez))
- **[provider]** Homogenization of the providers (part 1) ([#2518](https://github.com/traefik/traefik/pull/2518) by [ldez](https://github.com/ldez))
- **[rancher]** Add all available labels to Rancher Backend ([#2601](https://github.com/traefik/traefik/pull/2601) by [ldez](https://github.com/ldez))
- **[rancher]** Homogenization of templates: Rancher ([#2662](https://github.com/traefik/traefik/pull/2662) by [ldez](https://github.com/ldez))
- **[rules]** Externalize Træfik rules in a dedicated package ([#2933](https://github.com/traefik/traefik/pull/2933) by [nmengin](https://github.com/nmengin))
- **[servicefabric]** Use shared label system ([#3197](https://github.com/traefik/traefik/pull/3197) by [ldez](https://github.com/ldez))
- **[servicefabric]** Update Service Fabric backend. ([#3064](https://github.com/traefik/traefik/pull/3064) by [ldez](https://github.com/ldez))
- **[servicefabric]** Add white list for Service Fabric ([#3079](https://github.com/traefik/traefik/pull/3079) by [ldez](https://github.com/ldez))
- **[tls]** Use default entryPoints when certificates are added with no entryPoints. ([#2534](https://github.com/traefik/traefik/pull/2534) by [nmengin](https://github.com/nmengin))
- **[tracing]** Handle zipkin collector creation ([#2860](https://github.com/traefik/traefik/pull/2860) by [ferhatelmas](https://github.com/ferhatelmas))
- **[tracing]** Opentracing support ([#2587](https://github.com/traefik/traefik/pull/2587) by [tcolgate](https://github.com/tcolgate) and [mmatur](https://github.com/mmatur))
- **[webui]** New web ui ([#2226](https://github.com/traefik/traefik/pull/2226) by [jkuri](https://github.com/jkuri))
- **[webui]** Add status code text to webui bar chart tooltip ([#2639](https://github.com/traefik/traefik/pull/2639) by [wader](https://github.com/wader))
- Logger and Leaks ([#2847](https://github.com/traefik/traefik/pull/2847) by [ldez](https://github.com/ldez))
- Separate command from the main package ([#2951](https://github.com/traefik/traefik/pull/2951) by [Juliens](https://github.com/Juliens))
- Use context in Server ([#3007](https://github.com/traefik/traefik/pull/3007) by [Juliens](https://github.com/Juliens))

**Bug fixes:**
- **[acme]** Check all the C/N and SANs of provided certificates before generating ACME certificates in ACME provider ([#2970](https://github.com/traefik/traefik/pull/2970) by [nmengin](https://github.com/nmengin))
- **[acme]** Update lego. ([#3158](https://github.com/traefik/traefik/pull/3158) by [ldez](https://github.com/ldez))
- **[acme]** Fix panic with wrong ACME configuration ([#3084](https://github.com/traefik/traefik/pull/3084) by [nmengin](https://github.com/nmengin))
- **[acme]** Minor updates to dumpcerts.sh ([#3116](https://github.com/traefik/traefik/pull/3116) by [mathuin](https://github.com/mathuin))
- **[acme]** Add ACME certificates only on ACME EntryPoint ([#3136](https://github.com/traefik/traefik/pull/3136) by [nmengin](https://github.com/nmengin))
- **[acme]** Add TTL and custom Timeout in DigitalOcean DNS provider  ([#3143](https://github.com/traefik/traefik/pull/3143) by [ldez](https://github.com/ldez))
- **[acme]** Fix acme.json file automatic creation ([#3156](https://github.com/traefik/traefik/pull/3156) by [nmengin](https://github.com/nmengin))
- **[acme]** Fix wildcard match to ACME domains in cluster mode ([#3080](https://github.com/traefik/traefik/pull/3080) by [oldmantaiter](https://github.com/oldmantaiter))
- **[api,cluster]** Moved /api/cluster/leadership handler under public routes (requires no authentication) ([#3101](https://github.com/traefik/traefik/pull/3101) by [aantono](https://github.com/aantono))
- **[authentication,middleware]** Forward auth: copy response headers when auth failed. ([#3207](https://github.com/traefik/traefik/pull/3207) by [ldez](https://github.com/ldez))
- **[consul,docker,ecs,eureka,k8s,kv,marathon,mesos,rancher]** Server weight zero ([#3130](https://github.com/traefik/traefik/pull/3130) by [ldez](https://github.com/ldez))
- **[docker,k8s,marathon]** Fix custom headers template ([#2622](https://github.com/traefik/traefik/pull/2622) by [ldez](https://github.com/ldez))
- **[docker,marathon,mesos,rancher]** Fix:  label &#39;traefik.domain&#39; ([#3201](https://github.com/traefik/traefik/pull/3201) by [ldez](https://github.com/ldez))
- **[docker,rancher]** Frontend rule and segment labels. ([#3091](https://github.com/traefik/traefik/pull/3091) by [ldez](https://github.com/ldez))
- **[docker,rancher]** Ignore server for container with empty IP address. ([#3213](https://github.com/traefik/traefik/pull/3213) by [ldez](https://github.com/ldez))
- **[docker]** Fix multiple frontends with docker-compose --scale ([#3190](https://github.com/traefik/traefik/pull/3190) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[healthcheck]** Remove unnecessary mutex usage in health checks ([#2726](https://github.com/traefik/traefik/pull/2726) by [marco-jantke](https://github.com/marco-jantke))
- **[k8s]** Missing annotation prefix support. ([#2915](https://github.com/traefik/traefik/pull/2915) by [ldez](https://github.com/ldez))
- **[k8s]** Remove hardcoded frontend prefix in Kubernetes template ([#2914](https://github.com/traefik/traefik/pull/2914) by [psalaberria002](https://github.com/psalaberria002))
- **[k8s]** Limit label selector to Ingress factory. ([#3137](https://github.com/traefik/traefik/pull/3137) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Fixes prefixed annotations support. ([#3110](https://github.com/traefik/traefik/pull/3110) by [ldez](https://github.com/ldez))
- **[logs,middleware]** Fix bad access log ([#2682](https://github.com/traefik/traefik/pull/2682) by [mmatur](https://github.com/mmatur))
- **[logs]** Add missing argument in log. ([#3188](https://github.com/traefik/traefik/pull/3188) by [chemidy](https://github.com/chemidy))
- **[marathon]** Several apps with same backend name in Marathon. ([#3109](https://github.com/traefik/traefik/pull/3109) by [ldez](https://github.com/ldez))
- **[mesos]** fix: overflow on 32 bits arch. ([#3127](https://github.com/traefik/traefik/pull/3127) by [ldez](https://github.com/ldez))
- **[metrics]** Fix duplicated tags in InfluxDB ([#3189](https://github.com/traefik/traefik/pull/3189) by [mmatur](https://github.com/mmatur))
- **[middleware,consul,consulcatalog,docker,ecs,kv,marathon,mesos,rancher]** Fix: error pages ([#3138](https://github.com/traefik/traefik/pull/3138) by [ldez](https://github.com/ldez))
- **[middleware,tracing]** Fix &lt;nil&gt; tracer value in KV ([#2911](https://github.com/traefik/traefik/pull/2911) by [mmatur](https://github.com/mmatur))
- **[middleware,tracing]** Fix nil value when tracing is enabled ([#3192](https://github.com/traefik/traefik/pull/3192) by [mmatur](https://github.com/mmatur))
- **[middleware]** Use responseModifier to override secure headers ([#2946](https://github.com/traefik/traefik/pull/2946) by [mmatur](https://github.com/mmatur))
- **[middleware]** Correct conditional setting of buffering retry expression. ([#2865](https://github.com/traefik/traefik/pull/2865) by [ldez](https://github.com/ldez))
- **[middleware]** Fix high memory usage in retry middleware ([#2740](https://github.com/traefik/traefik/pull/2740) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** Fix whitelist and XFF. ([#3211](https://github.com/traefik/traefik/pull/3211) by [ldez](https://github.com/ldez))
- **[middleware]** Fix panic in atomic on ARM and x86-32 platforms ([#3195](https://github.com/traefik/traefik/pull/3195) by [mmatur](https://github.com/mmatur))
- **[middleware]** Redirect to HTTPS first before basic auth if header redirect (secure) is set ([#3187](https://github.com/traefik/traefik/pull/3187) by [SantoDE](https://github.com/SantoDE))
- **[middleware]** Fix error pages redirect and headers. ([#3217](https://github.com/traefik/traefik/pull/3217) by [ldez](https://github.com/ldez))
- **[provider]** Add some missing quotes in templates ([#2973](https://github.com/traefik/traefik/pull/2973) by [ldez](https://github.com/ldez))
- **[servicefabric]** Fix backend name for stateful service and more. ([#3183](https://github.com/traefik/traefik/pull/3183) by [ldez](https://github.com/ldez))
- **[tracing]** Fix missing configuration for jaeger reporter ([#2720](https://github.com/traefik/traefik/pull/2720) by [mmatur](https://github.com/mmatur))
- **[tracing]** Tracing statusCodeTracker need to implement CloseNotify ([#2733](https://github.com/traefik/traefik/pull/2733) by [mmatur](https://github.com/mmatur))
- **[tracing]** Fix integration tests in tracing ([#2759](https://github.com/traefik/traefik/pull/2759) by [mmatur](https://github.com/mmatur))
- **[webui]** Remove useless ACME tab from UI. ([#3154](https://github.com/traefik/traefik/pull/3154) by [ldez](https://github.com/ldez))
- **[webui]** Add redirect section. ([#3243](https://github.com/traefik/traefik/pull/3243) by [ldez](https://github.com/ldez))

**Documentation:**
- **[docker]** Add default values for some Docker labels ([#2604](https://github.com/traefik/traefik/pull/2604) by [ldez](https://github.com/ldez))
- **[file]** Add documentation about Templating in backend file ([#3223](https://github.com/traefik/traefik/pull/3223) by [nmengin](https://github.com/nmengin))
- **[k8s]** Update traefik-ds.yaml with --api command line parameter ([#2803](https://github.com/traefik/traefik/pull/2803) by [maniankara](https://github.com/maniankara))
- **[k8s]** Remove web provider in example ([#2807](https://github.com/traefik/traefik/pull/2807) by [pigletfly](https://github.com/pigletfly))
- **[k8s]** Drop capabilities in Kubernetes DaemonSet example ([#3028](https://github.com/traefik/traefik/pull/3028) by [nogoegst](https://github.com/nogoegst))
- **[k8s]** Docs: Fix typos in k8s user-guide ([#2898](https://github.com/traefik/traefik/pull/2898) by [cez81](https://github.com/cez81))
- **[k8s]** Change boolean annotation values to string ([#2839](https://github.com/traefik/traefik/pull/2839) by [hobti01](https://github.com/hobti01))
- **[k8s]** Update kubernetes.md ([#3093](https://github.com/traefik/traefik/pull/3093) by [rdrgporto](https://github.com/rdrgporto))
- **[k8s]** Document custom k8s ingress class usage in guide. ([#3242](https://github.com/traefik/traefik/pull/3242) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Update kubernetes.md ([#3171](https://github.com/traefik/traefik/pull/3171) by [andreyfedoseev](https://github.com/andreyfedoseev))
- **[provider]** Split security labels and custom labels documentation. ([#2872](https://github.com/traefik/traefik/pull/2872) by [ldez](https://github.com/ldez))
- **[provider]** Remove non-supported label. ([#3065](https://github.com/traefik/traefik/pull/3065) by [ldez](https://github.com/ldez))
- **[provider]** Remove obsolete paragraph about error pages. ([#2608](https://github.com/traefik/traefik/pull/2608) by [ldez](https://github.com/ldez))
- **[provider]** Cleaning labels/annotations documentation. ([#3245](https://github.com/traefik/traefik/pull/3245) by [ldez](https://github.com/ldez))
- **[provider]** Fix template version documentation. ([#3184](https://github.com/traefik/traefik/pull/3184) by [ldez](https://github.com/ldez))
- **[servicefabric]** Add SF to supported backends in docs ([#3033](https://github.com/traefik/traefik/pull/3033) by [lawrencegripper](https://github.com/lawrencegripper))
- **[servicefabric]** Update SF white list documentation section. ([#3082](https://github.com/traefik/traefik/pull/3082) by [ldez](https://github.com/ldez))
- **[tracing]** Fix typo in doc for rate limit label ([#2790](https://github.com/traefik/traefik/pull/2790) by [mmatur](https://github.com/mmatur))
- **[tracing]** Add Tracing entry in the documentation. ([#2713](https://github.com/traefik/traefik/pull/2713) by [ldez](https://github.com/ldez))
- **[tracing]** Fix documentation for tracing with Jaeger ([#3227](https://github.com/traefik/traefik/pull/3227) by [mmatur](https://github.com/mmatur))
- **[webui]** doc: update Traefik images. ([#3241](https://github.com/traefik/traefik/pull/3241) by [ldez](https://github.com/ldez))
- Fix typo in documentation ([#3215](https://github.com/traefik/traefik/pull/3215) by [arnaslu](https://github.com/arnaslu))
- Minor improvements to documentation ([#3221](https://github.com/traefik/traefik/pull/3221) by [colincoller](https://github.com/colincoller))
- Update some examples ([#3150](https://github.com/traefik/traefik/pull/3150) by [zaporylie](https://github.com/zaporylie))
- Normalize parameter names in configs ([#3132](https://github.com/traefik/traefik/pull/3132) by [kachkaev](https://github.com/kachkaev))
- Fixed documentation urls on README.md ([#3102](https://github.com/traefik/traefik/pull/3102) by [emir](https://github.com/emir))
- Fix typo and tweak formatting in quickstart ([#3250](https://github.com/traefik/traefik/pull/3250) by [alexymik](https://github.com/alexymik))
- Fix basic documentation ([#3086](https://github.com/traefik/traefik/pull/3086) by [mmatur](https://github.com/mmatur))
- Prepare release v1.6.0-rc6 ([#3199](https://github.com/traefik/traefik/pull/3199) by [mmatur](https://github.com/mmatur))
- Prepare release v1.6.0-rc5 ([#3179](https://github.com/traefik/traefik/pull/3179) by [Juliens](https://github.com/Juliens))
- Prepare release v1.6.0-rc4 ([#3126](https://github.com/traefik/traefik/pull/3126) by [ldez](https://github.com/ldez))
- Prepare release v1.6.0-rc3 ([#3096](https://github.com/traefik/traefik/pull/3096) by [ldez](https://github.com/ldez))
- Prepare release v1.6.0-rc2 ([#3087](https://github.com/traefik/traefik/pull/3087) by [nmengin](https://github.com/nmengin))
- Prepare release v1.6.0-rc1 ([#3078](https://github.com/traefik/traefik/pull/3078) by [Juliens](https://github.com/Juliens))
- Prepare release v1.6.0 ([#3251](https://github.com/traefik/traefik/pull/3251) by [Juliens](https://github.com/Juliens))

**Misc:**
- **[oxy]** Disable closeNotify when method GET for http pipelining ([#3108](https://github.com/traefik/traefik/pull/3108) by [Juliens](https://github.com/Juliens))
- **[boltdb,consul,etcd,kv,zk]** Migrate from libkv to valkeyrie library ([#2743](https://github.com/traefik/traefik/pull/2743) by [nmengin](https://github.com/nmengin))
- Drop unnecessary type conversions ([#2583](https://github.com/traefik/traefik/pull/2583) by [ferhatelmas](https://github.com/ferhatelmas))
- Code simplification ([#2516](https://github.com/traefik/traefik/pull/2516) by [ferhatelmas](https://github.com/ferhatelmas))
- Merge v1.5.4 into master  ([#3024](https://github.com/traefik/traefik/pull/3024) by [ldez](https://github.com/ldez))
- Merge v1.5.3 into master ([#2943](https://github.com/traefik/traefik/pull/2943) by [ldez](https://github.com/ldez))
- Merge v1.5.2 into master  ([#2843](https://github.com/traefik/traefik/pull/2843) by [ldez](https://github.com/ldez))
- Merge v1.5.1 into master ([#2781](https://github.com/traefik/traefik/pull/2781) by [ldez](https://github.com/ldez))
- Merge v1.5.0-rc5 into master ([#2708](https://github.com/traefik/traefik/pull/2708) by [ldez](https://github.com/ldez))
- Merge v1.5.0-rc3 into master ([#2600](https://github.com/traefik/traefik/pull/2600) by [ldez](https://github.com/ldez))
- Merge v1.5.0-rc2 into master ([#2536](https://github.com/traefik/traefik/pull/2536) by [ldez](https://github.com/ldez))

## [v1.6.0-rc6](https://github.com/traefik/traefik/tree/v1.6.0-rc6) (2018-04-17)
[All Commits](https://github.com/traefik/traefik/compare/v1.6.0-rc5...v1.6.0-rc6)

**Enhancements:**
- **[acme]** Create backup file during migration from ACME V1 to ACME V2 ([#3191](https://github.com/traefik/traefik/pull/3191) by [nmengin](https://github.com/nmengin))
- **[servicefabric]** Use shared label system ([#3197](https://github.com/traefik/traefik/pull/3197) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[docker]** Fix multiple frontends with docker-compose --scale ([#3190](https://github.com/traefik/traefik/pull/3190) by [jbdoumenjou](https://github.com/jbdoumenjou))
- **[metrics]** Fix duplicated tags in InfluxDB ([#3189](https://github.com/traefik/traefik/pull/3189) by [mmatur](https://github.com/mmatur))
- **[middleware,tracing]** Fix nil value when tracing is enabled ([#3192](https://github.com/traefik/traefik/pull/3192) by [mmatur](https://github.com/mmatur))
- **[middleware]** Fix panic in atomic on ARM and x86-32 platforms ([#3195](https://github.com/traefik/traefik/pull/3195) by [mmatur](https://github.com/mmatur))
- **[middleware]** Redirect to HTTPS first before basic auth if header redirect (secure) is set ([#3187](https://github.com/traefik/traefik/pull/3187) by [SantoDE](https://github.com/SantoDE))
- **[servicefabric]** Fix backend name for stateful service and more. ([#3183](https://github.com/traefik/traefik/pull/3183) by [ldez](https://github.com/ldez))
- Add missing argument in log. ([#3188](https://github.com/traefik/traefik/pull/3188) by [chemidy](https://github.com/chemidy))

**Documentation:**
- **[provider]** Fix template version documentation. ([#3184](https://github.com/traefik/traefik/pull/3184) by [ldez](https://github.com/ldez))

## [v1.6.0-rc5](https://github.com/traefik/traefik/tree/v1.6.0-rc5) (2018-04-12)
[All Commits](https://github.com/traefik/traefik/compare/v1.6.0-rc4...v1.6.0-rc5)

**Enhancements:**
- **[acme]** Generate wildcard certificate with SANs in ACME ([#3167](https://github.com/traefik/traefik/pull/3167) by [nmengin](https://github.com/nmengin))
- **[ecs]** Factorize labels managements. ([#3159](https://github.com/traefik/traefik/pull/3159) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[acme]** Update lego. ([#3158](https://github.com/traefik/traefik/pull/3158) by [ldez](https://github.com/ldez))
- **[acme]** Fix acme.json file automatic creation ([#3156](https://github.com/traefik/traefik/pull/3156) by [nmengin](https://github.com/nmengin))
- **[acme]** Minor updates to dumpcerts.sh ([#3116](https://github.com/traefik/traefik/pull/3116) by [mathuin](https://github.com/mathuin))
- **[acme]** Add TTL and custom Timeout in DigitalOcean DNS provider  ([#3143](https://github.com/traefik/traefik/pull/3143) by [ldez](https://github.com/ldez))
- **[acme]** Add ACME certificates only on ACME EntryPoint ([#3136](https://github.com/traefik/traefik/pull/3136) by [nmengin](https://github.com/nmengin))
- **[consul,docker,ecs,eureka,k8s,kv,marathon,mesos,rancher]** Server weight zero ([#3130](https://github.com/traefik/traefik/pull/3130) by [ldez](https://github.com/ldez))
- **[k8s]** Limit label selector to Ingress factory. ([#3137](https://github.com/traefik/traefik/pull/3137) by [timoreimann](https://github.com/timoreimann))
- **[middleware,consul,consulcatalog,docker,ecs,kv,marathon,mesos,rancher]** Fix: error pages ([#3138](https://github.com/traefik/traefik/pull/3138) by [ldez](https://github.com/ldez))
- **[webui]** Remove useless ACME tab from UI. ([#3154](https://github.com/traefik/traefik/pull/3154) by [ldez](https://github.com/ldez))

**Documentation:**
- **[k8s]** Update kubernetes.md ([#3171](https://github.com/traefik/traefik/pull/3171) by [andreyfedoseev](https://github.com/andreyfedoseev))
- Update some examples ([#3150](https://github.com/traefik/traefik/pull/3150) by [zaporylie](https://github.com/zaporylie))
- Normalize parameter names in configs ([#3132](https://github.com/traefik/traefik/pull/3132) by [kachkaev](https://github.com/kachkaev))

**Misc:**
- **[oxy]** Disable closeNotify when method GET for http pipelining ([#3108](https://github.com/traefik/traefik/pull/3108) by [Juliens](https://github.com/Juliens))

## [v1.6.0-rc4](https://github.com/traefik/traefik/tree/v1.6.0-rc4) (2018-04-04)
[All Commits](https://github.com/traefik/traefik/compare/v1.6.0-rc3...v1.6.0-rc4)

**Enhancements:**
- **[consulcatalog,ecs,mesos]** Factorize labels managements. ([#3099](https://github.com/traefik/traefik/pull/3099) by [ldez](https://github.com/ldez))
- **[middleware]** Add tests on IPWhiteLister. ([#3106](https://github.com/traefik/traefik/pull/3106) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[api,cluster]** Moved /api/cluster/leadership handler under public routes (requires no authentication) ([#3101](https://github.com/traefik/traefik/pull/3101) by [aantono](https://github.com/aantono))
- **[k8s]** Fixes prefixed annotations support. ([#3110](https://github.com/traefik/traefik/pull/3110) by [ldez](https://github.com/ldez))
- **[marathon]** Several apps with same backend name in Marathon. ([#3109](https://github.com/traefik/traefik/pull/3109) by [ldez](https://github.com/ldez))

**Documentation:**
- **[k8s]** Update kubernetes.md ([#3093](https://github.com/traefik/traefik/pull/3093) by [rdrgporto](https://github.com/rdrgporto))
- Fixed documentation urls on README.md ([#3102](https://github.com/traefik/traefik/pull/3102) by [emir](https://github.com/emir))

## [v1.6.0-rc3](https://github.com/traefik/traefik/tree/v1.6.0-rc3) (2018-03-28)
[All Commits](https://github.com/traefik/traefik/compare/v1.6.0-rc2...v1.6.0-rc3)

**Bug fixes:**
- **[docker,rancher]** Frontend rule and segment labels. ([#3091](https://github.com/traefik/traefik/pull/3091) by [ldez](https://github.com/ldez))

## [v1.6.0-rc2](https://github.com/traefik/traefik/tree/v1.6.0-rc2) (2018-03-27)
[All Commits](https://github.com/traefik/traefik/compare/v1.6.0-rc1...v1.6.0-rc2)

**Bug fixes:**
- **[acme]** Fix panic with wrong ACME configuration ([#3084](https://github.com/traefik/traefik/pull/3084) by [nmengin](https://github.com/nmengin))
- **[acme]** Fix wildcard match to ACME domains in cluster mode ([#3080](https://github.com/traefik/traefik/pull/3080) by [oldmantaiter](https://github.com/oldmantaiter))

**Documentation:**
- **[servicefabric]** Update SF white list documentation section. ([#3082](https://github.com/traefik/traefik/pull/3082) by [ldez](https://github.com/ldez))
- Fix basic documentation ([#3086](https://github.com/traefik/traefik/pull/3086) by [mmatur](https://github.com/mmatur))

## [v1.6.0-rc1](https://github.com/traefik/traefik/tree/v1.6.0-rc1) (2018-03-26)
[All Commits](https://github.com/traefik/traefik/compare/v1.5.0-rc1...v1.6.0-rc1)

**Enhancements:**
- **[acme]** Bump Lego Version for GoDaddy DNS Provider ([#2482](https://github.com/traefik/traefik/pull/2482) by [sjawhar](https://github.com/sjawhar))
- **[acme]** Simplify storing renewed acme certificate ([#2614](https://github.com/traefik/traefik/pull/2614) by [ferhatelmas](https://github.com/ferhatelmas))
- **[acme]** Delete TLS-SNI-01 challenge from ACME ([#2971](https://github.com/traefik/traefik/pull/2971) by [nmengin](https://github.com/nmengin))
- **[acme]** ACME V2 Integration ([#3063](https://github.com/traefik/traefik/pull/3063) by [nmengin](https://github.com/nmengin))
- **[acme]** Update Lego (Gandi API v5, cloudxns, ...) ([#2844](https://github.com/traefik/traefik/pull/2844) by [ldez](https://github.com/ldez))
- **[acme]** Create ACME Provider ([#2889](https://github.com/traefik/traefik/pull/2889) by [nmengin](https://github.com/nmengin))
- **[api,cluster]** Added cluster/leader endpoint ([#3009](https://github.com/traefik/traefik/pull/3009) by [aantono](https://github.com/aantono))
- **[authentication]** Forward Authentication: add X-Forwarded-Uri ([#2398](https://github.com/traefik/traefik/pull/2398) by [sebastianbauer](https://github.com/sebastianbauer))
- **[boltdb,consul,etcd,kv,zk]** homogenization of templates: KV ([#2661](https://github.com/traefik/traefik/pull/2661) by [ldez](https://github.com/ldez))
- **[boltdb,consul,etcd,kv,zk]** Add all available configuration to KV Backend ([#2652](https://github.com/traefik/traefik/pull/2652) by [ldez](https://github.com/ldez))
- **[boltdb,consul,etcd,kv,zk]** Homogenization of the providers (part 1):  KV ([#2616](https://github.com/traefik/traefik/pull/2616) by [ldez](https://github.com/ldez))
- **[consul,consulcatalog]** Homogenization of templates: Consul Catalog ([#2668](https://github.com/traefik/traefik/pull/2668) by [ldez](https://github.com/ldez))
- **[consul,consulcatalog]** Split consul and consul catalog. ([#2655](https://github.com/traefik/traefik/pull/2655) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Add all available tags to Consul Catalog Backend ([#2646](https://github.com/traefik/traefik/pull/2646) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Check for endpoints while detecting Consul service changes ([#2882](https://github.com/traefik/traefik/pull/2882) by [caseycs](https://github.com/caseycs))
- **[consulcatalog]** TLS Support for ConsulCatalog ([#2900](https://github.com/traefik/traefik/pull/2900) by [mmatur](https://github.com/mmatur))
- **[docker,docker/swarm]** Fix support for macvlan driver in docker provider ([#2827](https://github.com/traefik/traefik/pull/2827) by [mmatur](https://github.com/mmatur))
- **[docker,marathon,rancher]** Segments Labels: Rancher &amp; Marathon ([#3073](https://github.com/traefik/traefik/pull/3073) by [ldez](https://github.com/ldez))
- **[docker]** Custom headers by service labels for docker backends ([#2514](https://github.com/traefik/traefik/pull/2514) by [Tiscs](https://github.com/Tiscs))
- **[docker]** Homogenization of templates: Docker ([#2659](https://github.com/traefik/traefik/pull/2659) by [ldez](https://github.com/ldez))
- **[docker]** Segment labels: Docker ([#3055](https://github.com/traefik/traefik/pull/3055) by [ldez](https://github.com/ldez))
- **[docker]** Add all available labels to Docker Backend ([#2584](https://github.com/traefik/traefik/pull/2584) by [ldez](https://github.com/ldez))
- **[dynamodb,ecs]** Upgrade AWS SKD to version v1.13.1 ([#2908](https://github.com/traefik/traefik/pull/2908) by [mmatur](https://github.com/mmatur))
- **[ecs]** Add all available labels to ECS Backend ([#2605](https://github.com/traefik/traefik/pull/2605) by [ldez](https://github.com/ldez))
- **[ecs]** Homogenization of templates: ECS ([#2663](https://github.com/traefik/traefik/pull/2663) by [ldez](https://github.com/ldez))
- **[eureka]** Replace Delay by RefreshSecond in Eureka ([#2972](https://github.com/traefik/traefik/pull/2972) by [ldez](https://github.com/ldez))
- **[eureka]** Homogenization of templates: Eureka  ([#2846](https://github.com/traefik/traefik/pull/2846) by [ldez](https://github.com/ldez))
- **[file]** Added support for templates to file provider ([#2991](https://github.com/traefik/traefik/pull/2991) by [aantono](https://github.com/aantono))
- **[healthcheck]** Toggle /ping to artificially return unhealthy response on SIGTERM during requestAcceptGraceTimeout interval ([#3062](https://github.com/traefik/traefik/pull/3062) by [ravilr](https://github.com/ravilr))
- **[healthcheck]** Improve logging output for failing healthchecks ([#2443](https://github.com/traefik/traefik/pull/2443) by [marco-jantke](https://github.com/marco-jantke))
- **[k8s,tls]** Add support for fetching k8s Ingress TLS data from secrets ([#2439](https://github.com/traefik/traefik/pull/2439) by [gopenguin](https://github.com/gopenguin))
- **[k8s]** Bump kubernetes/client-go ([#2848](https://github.com/traefik/traefik/pull/2848) by [yue9944882](https://github.com/yue9944882))
- **[k8s]** Allow custom value for kubernetes.io/ingress.class annotation ([#2222](https://github.com/traefik/traefik/pull/2222) by [yuvipanda](https://github.com/yuvipanda))
- **[k8s]** Add app-root annotation support for kubernetes ingress ([#2522](https://github.com/traefik/traefik/pull/2522) by [yue9944882](https://github.com/yue9944882))
- **[k8s]** Builders in k8s tests ([#2513](https://github.com/traefik/traefik/pull/2513) by [ldez](https://github.com/ldez))
- **[k8s]** Add all available annotations to k8s Backend ([#2612](https://github.com/traefik/traefik/pull/2612) by [ldez](https://github.com/ldez))
- **[k8s]** Introduce k8s informer factory ([#2867](https://github.com/traefik/traefik/pull/2867) by [yue9944882](https://github.com/yue9944882))
- **[logs,middleware]** Add access log filter for retry attempts ([#3042](https://github.com/traefik/traefik/pull/3042) by [marco-jantke](https://github.com/marco-jantke))
- **[logs,middleware]** Ultimate Access log filter ([#2988](https://github.com/traefik/traefik/pull/2988) by [mmatur](https://github.com/mmatur))
- **[logs,middleware]** Add username in accesslog ([#2111](https://github.com/traefik/traefik/pull/2111) by [bastiaanb](https://github.com/bastiaanb))
- **[logs]** Allow overriding the log level in debug mode. ([#3050](https://github.com/traefik/traefik/pull/3050) by [timoreimann](https://github.com/timoreimann))
- **[logs]** Display file log when test fails. ([#2801](https://github.com/traefik/traefik/pull/2801) by [ldez](https://github.com/ldez))
- **[marathon]** Remove health check filter from Marathon tasks. ([#2817](https://github.com/traefik/traefik/pull/2817) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Add all available labels to Marathon Backend ([#2602](https://github.com/traefik/traefik/pull/2602) by [ldez](https://github.com/ldez))
- **[marathon]** homogenization of templates: Marathon ([#2665](https://github.com/traefik/traefik/pull/2665) by [ldez](https://github.com/ldez))
- **[mesos]** Add all available labels to Mesos Backend  ([#2687](https://github.com/traefik/traefik/pull/2687) by [ldez](https://github.com/ldez))
- **[metrics]** Added entrypoint metrics to influxdb ([#2992](https://github.com/traefik/traefik/pull/2992) by [adityacs](https://github.com/adityacs))
- **[metrics]** Extend metrics and rebuild prometheus exporting logic ([#2567](https://github.com/traefik/traefik/pull/2567) by [marco-jantke](https://github.com/marco-jantke))
- **[metrics]** Added missing metrics to registry for Datadog and StatsD ([#2890](https://github.com/traefik/traefik/pull/2890) by [aantono](https://github.com/aantono))
- **[metrics]** Remove unnecessary conversion ([#2850](https://github.com/traefik/traefik/pull/2850) by [ferhatelmas](https://github.com/ferhatelmas))
- **[middleware,consul,consulcatalog,docker,ecs,k8s,marathon,mesos,rancher]** New option in secure middleware ([#2958](https://github.com/traefik/traefik/pull/2958) by [mmatur](https://github.com/mmatur))
- **[middleware,consulcatalog,docker,ecs,k8s,kv,marathon,mesos,rancher]** Ability to use &#34;X-Forwarded-For&#34; as a source of IP for white list. ([#3070](https://github.com/traefik/traefik/pull/3070) by [ldez](https://github.com/ldez))
- **[middleware,docker]** Use pointer of error pages ([#2607](https://github.com/traefik/traefik/pull/2607) by [ldez](https://github.com/ldez))
- **[middleware,provider]** Redirection: permanent move option. ([#2774](https://github.com/traefik/traefik/pull/2774) by [ldez](https://github.com/ldez))
- **[middleware]** Add new options to the CLI entrypoint definition.  ([#2799](https://github.com/traefik/traefik/pull/2799) by [ldez](https://github.com/ldez))
- **[middleware]** Change port of traefik for error pages integration test ([#2907](https://github.com/traefik/traefik/pull/2907) by [mmatur](https://github.com/mmatur))
- **[middleware]** Request buffering middleware ([#2217](https://github.com/traefik/traefik/pull/2217) by [harnash](https://github.com/harnash))
- **[middleware]** Remove unnecessary returns in tracing setup ([#2880](https://github.com/traefik/traefik/pull/2880) by [ferhatelmas](https://github.com/ferhatelmas))
- **[provider]** Homogenization of the providers (part 1) ([#2518](https://github.com/traefik/traefik/pull/2518) by [ldez](https://github.com/ldez))
- **[provider]** No error pages must return nil. ([#2610](https://github.com/traefik/traefik/pull/2610) by [ldez](https://github.com/ldez))
- **[rancher]** Homogenization of templates: Rancher ([#2662](https://github.com/traefik/traefik/pull/2662) by [ldez](https://github.com/ldez))
- **[rancher]** Add all available labels to Rancher Backend ([#2601](https://github.com/traefik/traefik/pull/2601) by [ldez](https://github.com/ldez))
- **[rules]** Externalize Træfik rules in a dedicated package ([#2933](https://github.com/traefik/traefik/pull/2933) by [nmengin](https://github.com/nmengin))
- **[servicefabric]** Update Service Fabric backend. ([#3064](https://github.com/traefik/traefik/pull/3064) by [ldez](https://github.com/ldez))
- **[servicefabric]** Add white list for Service Fabric. ([#3079](https://github.com/traefik/traefik/pull/3079) by [ldez](https://github.com/ldez))
- **[tls]** Use default entryPoints when certificates are added with no entryPoints. ([#2534](https://github.com/traefik/traefik/pull/2534) by [nmengin](https://github.com/nmengin))
- **[tracing]** Handle zipkin collector creation ([#2860](https://github.com/traefik/traefik/pull/2860) by [ferhatelmas](https://github.com/ferhatelmas))
- **[tracing]** Opentracing support ([#2587](https://github.com/traefik/traefik/pull/2587) by [mmatur](https://github.com/mmatur))
- **[webui]** Add status code text to webui bar chart tooltip ([#2639](https://github.com/traefik/traefik/pull/2639) by [wader](https://github.com/wader))
- Separate command from the main package ([#2951](https://github.com/traefik/traefik/pull/2951) by [Juliens](https://github.com/Juliens))
- Use context in Server ([#3007](https://github.com/traefik/traefik/pull/3007) by [Juliens](https://github.com/Juliens))
- Logger and Leaks ([#2847](https://github.com/traefik/traefik/pull/2847) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[acme]** Check all the C/N and SANs of provided certificates before generating ACME certificates in ACME provider ([#2970](https://github.com/traefik/traefik/pull/2970) by [nmengin](https://github.com/nmengin))
- **[docker,k8s,marathon]** Fix custom headers template ([#2622](https://github.com/traefik/traefik/pull/2622) by [ldez](https://github.com/ldez))
- **[k8s]** Missing annotation prefix support. ([#2915](https://github.com/traefik/traefik/pull/2915) by [ldez](https://github.com/ldez))
- **[k8s]** Remove hardcoded frontend prefix in Kubernetes template ([#2914](https://github.com/traefik/traefik/pull/2914) by [psalaberria002](https://github.com/psalaberria002))
- **[logs,middleware]** Fix bad access log ([#2682](https://github.com/traefik/traefik/pull/2682) by [mmatur](https://github.com/mmatur))
- **[middleware,tracing]** Fix &lt;nil&gt; tracer value in KV ([#2911](https://github.com/traefik/traefik/pull/2911) by [mmatur](https://github.com/mmatur))
- **[middleware]** Use responseModifier to override secure headers ([#2946](https://github.com/traefik/traefik/pull/2946) by [mmatur](https://github.com/mmatur))
- **[middleware]** Correct conditional setting of buffering retry expression. ([#2865](https://github.com/traefik/traefik/pull/2865) by [ldez](https://github.com/ldez))
- **[middleware]** Fix high memory usage in retry middleware ([#2740](https://github.com/traefik/traefik/pull/2740) by [marco-jantke](https://github.com/marco-jantke))
- **[provider]** Add some missing quotes in templates ([#2973](https://github.com/traefik/traefik/pull/2973) by [ldez](https://github.com/ldez))
- **[tracing]** Fix missing configuration for jaeger reporter ([#2720](https://github.com/traefik/traefik/pull/2720) by [mmatur](https://github.com/mmatur))
- **[tracing]** Tracing statusCodeTracker need to implement CloseNotify ([#2733](https://github.com/traefik/traefik/pull/2733) by [mmatur](https://github.com/mmatur))
- **[tracing]** Fix integration tests in tracing ([#2759](https://github.com/traefik/traefik/pull/2759) by [mmatur](https://github.com/mmatur))
- Remove unnecessary mutex usage in health checks ([#2726](https://github.com/traefik/traefik/pull/2726) by [marco-jantke](https://github.com/marco-jantke))

**Documentation:**
- **[docker]** Add default values for some Docker labels ([#2604](https://github.com/traefik/traefik/pull/2604) by [ldez](https://github.com/ldez))
- **[k8s]** Remove web provider in example ([#2807](https://github.com/traefik/traefik/pull/2807) by [pigletfly](https://github.com/pigletfly))
- **[k8s]** Update traefik-ds.yaml with --api command line parameter ([#2803](https://github.com/traefik/traefik/pull/2803) by [maniankara](https://github.com/maniankara))
- **[k8s]** Drop capabilities in Kubernetes DaemonSet example ([#3028](https://github.com/traefik/traefik/pull/3028) by [nogoegst](https://github.com/nogoegst))
- **[k8s]** Docs: Fix typos in k8s user-guide ([#2898](https://github.com/traefik/traefik/pull/2898) by [cez81](https://github.com/cez81))
- **[k8s]** Change boolean annotation values to string ([#2839](https://github.com/traefik/traefik/pull/2839) by [hobti01](https://github.com/hobti01))
- **[provider]** Split security labels and custom labels documentation. ([#2872](https://github.com/traefik/traefik/pull/2872) by [ldez](https://github.com/ldez))
- **[provider]** Remove non-supported label. ([#3065](https://github.com/traefik/traefik/pull/3065) by [ldez](https://github.com/ldez))
- **[provider]** Remove obsolete paragraph about error pages. ([#2608](https://github.com/traefik/traefik/pull/2608) by [ldez](https://github.com/ldez))
- **[servicefabric]** Add SF to supported backends in docs ([#3033](https://github.com/traefik/traefik/pull/3033) by [lawrencegripper](https://github.com/lawrencegripper))
- Fix typo in doc for rate limit label ([#2790](https://github.com/traefik/traefik/pull/2790) by [mmatur](https://github.com/mmatur))
- Add Tracing entry in the documentation. ([#2713](https://github.com/traefik/traefik/pull/2713) by [ldez](https://github.com/ldez))

**Misc:**
- **[boltdb,consul,etcd,kv,zk]** Migrate from libkv to valkeyrie library ([#2743](https://github.com/traefik/traefik/pull/2743) by [nmengin](https://github.com/nmengin))
- Merge v1.5.4 into master  ([#3024](https://github.com/traefik/traefik/pull/3024) by [ldez](https://github.com/ldez))
- Merge v1.5.3 into master ([#2943](https://github.com/traefik/traefik/pull/2943) by [ldez](https://github.com/ldez))
- Merge v1.5.2 into master  ([#2843](https://github.com/traefik/traefik/pull/2843) by [ldez](https://github.com/ldez))
- Merge v1.5.1 into master ([#2781](https://github.com/traefik/traefik/pull/2781) by [ldez](https://github.com/ldez))
- Merge v1.5.0-rc5 into master ([#2708](https://github.com/traefik/traefik/pull/2708) by [ldez](https://github.com/ldez))
- Merge 1.5.0-rc3 into master ([#2600](https://github.com/traefik/traefik/pull/2600) by [ldez](https://github.com/ldez))
- Drop unnecessary type conversions ([#2583](https://github.com/traefik/traefik/pull/2583) by [ferhatelmas](https://github.com/ferhatelmas))
- Merge 1.5.0-rc2 into master ([#2536](https://github.com/traefik/traefik/pull/2536) by [ldez](https://github.com/ldez))
- Code simplification ([#2516](https://github.com/traefik/traefik/pull/2516) by [ferhatelmas](https://github.com/ferhatelmas))

## [v1.5.4](https://github.com/traefik/traefik/tree/v1.5.4) (2018-03-15)
[All Commits](https://github.com/traefik/traefik/compare/v1.5.3...v1.5.4)

**Bug fixes:**
- **[acme]** Fix panic when parsing resolv.conf ([#2955](https://github.com/traefik/traefik/pull/2955) by [ldez](https://github.com/ldez))
- **[acme]** Don&#39;t failed traefik start if register and subscribe failed on acme ([#2977](https://github.com/traefik/traefik/pull/2977) by [Juliens](https://github.com/Juliens))
- **[ecs]** Safe access to ECS API pointer values. ([#2983](https://github.com/traefik/traefik/pull/2983) by [ldez](https://github.com/ldez))
- **[kv]** Add lower-case passHostHeader key support. ([#3015](https://github.com/traefik/traefik/pull/3015) by [ldez](https://github.com/ldez))
- **[middleware]** Propagate insecure in white list. ([#2981](https://github.com/traefik/traefik/pull/2981) by [ldez](https://github.com/ldez))
- **[rancher]** Fix Rancher Healthcheck when upgrading a service ([#2962](https://github.com/traefik/traefik/pull/2962) by [jmirc](https://github.com/jmirc))
- **[websocket]** Capitalize Sec-WebSocket-Protocol Header ([#2975](https://github.com/traefik/traefik/pull/2975) by [Juliens](https://github.com/Juliens))
- Use goroutine pool in throttleProvider ([#3013](https://github.com/traefik/traefik/pull/3013) by [Juliens](https://github.com/Juliens))
- Handle quoted strings in UnmarshalJSON ([#3004](https://github.com/traefik/traefik/pull/3004) by [Juliens](https://github.com/Juliens))

**Documentation:**
- **[acme]** Clarify some deprecations. ([#2959](https://github.com/traefik/traefik/pull/2959) by [ldez](https://github.com/ldez))
- **[acme]** Second defaultEntryPoint should be https, not http. ([#2948](https://github.com/traefik/traefik/pull/2948) by [GerbenWelter](https://github.com/GerbenWelter))
- **[api]** Enhance API, REST, ping documentation. ([#2950](https://github.com/traefik/traefik/pull/2950) by [ldez](https://github.com/ldez))
- **[k8s]** Add TLS Docs ([#3012](https://github.com/traefik/traefik/pull/3012) by [dtomcej](https://github.com/dtomcej))
- Enhance Traefik TOML sample. ([#2996](https://github.com/traefik/traefik/pull/2996) by [ldez](https://github.com/ldez))
- Fix typo in docs ([#2990](https://github.com/traefik/traefik/pull/2990) by [mo](https://github.com/mo))
- Clarify how setting a frontend priority works ([#2984](https://github.com/traefik/traefik/pull/2984) by [jbdoumenjou](https://github.com/jbdoumenjou))
- Add [file] in syntax reference ([#3016](https://github.com/traefik/traefik/pull/3016) by [ldez](https://github.com/ldez))
- Updated the test-it example according to the latest docker version ([#3000](https://github.com/traefik/traefik/pull/3000) by [geraldcroes](https://github.com/geraldcroes))

## [v1.5.3](https://github.com/traefik/traefik/tree/v1.5.3) (2018-02-27)
[All Commits](https://github.com/traefik/traefik/compare/v1.5.2...v1.5.3)

**Bug fixes:**
- **[acme]** Check all the C/N and SANs of provided certificates before generating ACME certificates ([#2913](https://github.com/traefik/traefik/pull/2913) by [nmengin](https://github.com/nmengin))
- **[docker/swarm]** Empty IP address when use endpoint mode dnsrr ([#2887](https://github.com/traefik/traefik/pull/2887) by [mmatur](https://github.com/mmatur))
- **[middleware]** Infinite entry point redirection. ([#2929](https://github.com/traefik/traefik/pull/2929) by [ldez](https://github.com/ldez))
- **[provider]** Isolate backend with same name on different provider ([#2862](https://github.com/traefik/traefik/pull/2862) by [Juliens](https://github.com/Juliens))
- **[tls]** Starting Træfik even if TLS certificates are in error ([#2909](https://github.com/traefik/traefik/pull/2909) by [nmengin](https://github.com/nmengin))
- **[tls]**  Add DEBUG log when no provided certificate can check a domain ([#2938](https://github.com/traefik/traefik/pull/2938) by [nmengin](https://github.com/nmengin))
- **[webui]** Smooth dashboard refresh. ([#2871](https://github.com/traefik/traefik/pull/2871) by [ldez](https://github.com/ldez))
- Fix Duration JSON unmarshal ([#2935](https://github.com/traefik/traefik/pull/2935) by [ldez](https://github.com/ldez))
- Default value for lifecycle ([#2934](https://github.com/traefik/traefik/pull/2934) by [Juliens](https://github.com/Juliens))
- Check ping configuration. ([#2852](https://github.com/traefik/traefik/pull/2852) by [ldez](https://github.com/ldez))

**Documentation:**
- **[docker]** it&#39;s -&gt; its ([#2901](https://github.com/traefik/traefik/pull/2901) by [piec](https://github.com/piec))
- **[tls]** Fix doc cipher suites ([#2894](https://github.com/traefik/traefik/pull/2894) by [emilevauge](https://github.com/emilevauge))
- Add a CLI help command for Docker. ([#2921](https://github.com/traefik/traefik/pull/2921) by [ldez](https://github.com/ldez))
- Fix traffic pronounce dead link ([#2870](https://github.com/traefik/traefik/pull/2870) by [emilevauge](https://github.com/emilevauge))
- Update documentation on onHostRule, ping examples, and web deprecation ([#2863](https://github.com/traefik/traefik/pull/2863) by [Juliens](https://github.com/Juliens))

## [v1.5.2](https://github.com/traefik/traefik/tree/v1.5.2) (2018-02-12)
[All Commits](https://github.com/traefik/traefik/compare/v1.5.1...v1.5.2)

**Bug fixes:**
- **[acme,cluster,kv]** Compress ACME certificates in KV stores. ([#2814](https://github.com/traefik/traefik/pull/2814) by [nmengin](https://github.com/nmengin))
- **[acme]** Traefik still start when Let&#39;s encrypt is down ([#2794](https://github.com/traefik/traefik/pull/2794) by [Juliens](https://github.com/Juliens))
- **[docker]** Fix dnsrr endpoint mode excluded when not using swarm LB ([#2795](https://github.com/traefik/traefik/pull/2795) by [mmatur](https://github.com/mmatur))
- **[eureka]** Continue refresh the configuration after a failure. ([#2838](https://github.com/traefik/traefik/pull/2838) by [ldez](https://github.com/ldez))
- **[logs]** Reduce oxy round trip logs to debug. ([#2821](https://github.com/traefik/traefik/pull/2821) by [timoreimann](https://github.com/timoreimann))
- **[websocket]** Fix goroutine leaks in websocket ([#2825](https://github.com/traefik/traefik/pull/2825) by [Juliens](https://github.com/Juliens))
- Hide the pflag error when displaying help. ([#2800](https://github.com/traefik/traefik/pull/2800) by [ldez](https://github.com/ldez))

**Documentation:**
- **[docker]** Explain how to write entrypoints definition in a compose file ([#2834](https://github.com/traefik/traefik/pull/2834) by [mmatur](https://github.com/mmatur))
- **[docker]** Fix typo ([#2813](https://github.com/traefik/traefik/pull/2813) by [uschtwill](https://github.com/uschtwill))
- **[k8s]** typo in &#34;i&#34;ngress annotations. ([#2780](https://github.com/traefik/traefik/pull/2780) by [RRAlex](https://github.com/RRAlex))
- Clarify how setting a frontend priority works ([#2818](https://github.com/traefik/traefik/pull/2818) by [sirlatrom](https://github.com/sirlatrom))
- Fixed typo. ([#2811](https://github.com/traefik/traefik/pull/2811) by [sonus21](https://github.com/sonus21))
- Docs: regex+replacement hints for URL rewriting ([#2802](https://github.com/traefik/traefik/pull/2802) by [djeeg](https://github.com/djeeg))
- Add documentation about entry points definition with CLI. ([#2798](https://github.com/traefik/traefik/pull/2798) by [ldez](https://github.com/ldez))

## [v1.5.1](https://github.com/traefik/traefik/tree/v1.5.1) (2018-01-29)
[All Commits](https://github.com/traefik/traefik/compare/v1.5.0...v1.5.1)

**Bug fixes:**
- **[acme]** Handle undefined entrypoint on ACME config and frontend config ([#2756](https://github.com/traefik/traefik/pull/2756) by [Juliens](https://github.com/Juliens))
- **[k8s]** Fix the k8s redirection template. ([#2748](https://github.com/traefik/traefik/pull/2748) by [ldez](https://github.com/ldez))
- **[middleware]** Change gzipwriter receiver to implement CloseNotifier ([#2766](https://github.com/traefik/traefik/pull/2766) by [Juliens](https://github.com/Juliens))
- **[tls]** Fix domain names in dynamic TLS configuration ([#2768](https://github.com/traefik/traefik/pull/2768) by [nmengin](https://github.com/nmengin))

**Documentation:**
- **[acme]** Add note on redirect for ACME http challenge ([#2767](https://github.com/traefik/traefik/pull/2767) by [Juliens](https://github.com/Juliens))
- **[file]** Enhance file provider documentation. ([#2777](https://github.com/traefik/traefik/pull/2777) by [ldez](https://github.com/ldez))

## [v1.5.0](https://github.com/traefik/traefik/tree/v1.5.0) (2018-01-23)
[All Commits](https://github.com/traefik/traefik/compare/v1.4.0-rc1...v1.5.0)

**Enhancements:**
- **[acme,tls]** Rename TLSConfigurations to TLS. ([#2744](https://github.com/traefik/traefik/pull/2744) by [ldez](https://github.com/ldez))
- **[acme,provider,docker,tls]** Make the TLS certificates management dynamic. ([#2233](https://github.com/traefik/traefik/pull/2233) by [nmengin](https://github.com/nmengin))
- **[acme]** Add Let&#39;s Encrypt HTTP Challenge ([#2701](https://github.com/traefik/traefik/pull/2701) by [Juliens](https://github.com/Juliens))
- **[acme]** Update github.com/xenolf/lego to 0.4.1 ([#2304](https://github.com/traefik/traefik/pull/2304) by [oldmantaiter](https://github.com/oldmantaiter))
- **[api,healthcheck,metrics,provider,webui]** Split Web into API/Dashboard, ping, metric and Rest Provider ([#2335](https://github.com/traefik/traefik/pull/2335) by [Juliens](https://github.com/Juliens))
- **[authentication]** Pass through certain forward auth negative response headers ([#2127](https://github.com/traefik/traefik/pull/2127) by [wheresmysocks](https://github.com/wheresmysocks))
- **[cluster,consul,file]** Add file to storeconfig ([#2419](https://github.com/traefik/traefik/pull/2419) by [emilevauge](https://github.com/emilevauge))
- **[cluster,provider]** Support Etcd v3, enhance KV support ([#2407](https://github.com/traefik/traefik/pull/2407) by [nmengin](https://github.com/nmengin))
- **[docker,k8s,rancher,webui]** Redirect to another entryPoint per frontend ([#2133](https://github.com/traefik/traefik/pull/2133) by [SantoDE](https://github.com/SantoDE))
- **[docker,k8s,rancher]** Support regex redirect by frontend ([#2570](https://github.com/traefik/traefik/pull/2570) by [ldez](https://github.com/ldez))
- **[docker]** Add Custom header parsing to Docker Provider ([#2030](https://github.com/traefik/traefik/pull/2030) by [dtomcej](https://github.com/dtomcej))
- **[docker]** Docker labels ([#2473](https://github.com/traefik/traefik/pull/2473) by [ldez](https://github.com/ldez))
- **[docker]** Add docker security headers via labels ([#2334](https://github.com/traefik/traefik/pull/2334) by [dtomcej](https://github.com/dtomcej))
- **[docker]** Use Node IP in Swarm Standalone with &#34;host&#34; NetworkMode ([#2274](https://github.com/traefik/traefik/pull/2274) by [BlakeMesdag](https://github.com/BlakeMesdag))
- **[ecs]** ECS provider refactoring ([#2050](https://github.com/traefik/traefik/pull/2050) by [mmatur](https://github.com/mmatur))
- **[ecs]** Add health check label to ECS ([#2421](https://github.com/traefik/traefik/pull/2421) by [oldmantaiter](https://github.com/oldmantaiter))
- **[ecs]** Support Host NetworkMode  for ECS provider  ([#2320](https://github.com/traefik/traefik/pull/2320) by [FriggaHel](https://github.com/FriggaHel))
- **[etcd]** Manage certificates dynamically in kv store ([#2411](https://github.com/traefik/traefik/pull/2411) by [dahefanteng](https://github.com/dahefanteng))
- **[healthcheck]** Use health check for systemd watchdog ([#2283](https://github.com/traefik/traefik/pull/2283) by [guilhem](https://github.com/guilhem))
- **[k8s]** Kubernetes security header annotations ([#2460](https://github.com/traefik/traefik/pull/2460) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Add labels for `traefik.frontend.entryPoints` &amp; `PassTLSCert` to Kubernetes ([#2324](https://github.com/traefik/traefik/pull/2324) by [ryarnyah](https://github.com/ryarnyah))
- **[k8s]** Only listen to configured k8s namespaces. ([#1895](https://github.com/traefik/traefik/pull/1895) by [timoreimann](https://github.com/timoreimann))
- **[logs,middleware,consul,docker]** Use constants from http package. ([#2425](https://github.com/traefik/traefik/pull/2425) by [ldez](https://github.com/ldez))
- **[logs]** Add json format support for Traefik logs ([#2056](https://github.com/traefik/traefik/pull/2056) by [marco-jantke](https://github.com/marco-jantke))
- **[marathon]** Marathon constraints filtering ([#2388](https://github.com/traefik/traefik/pull/2388) by [aantono](https://github.com/aantono))
- **[marathon]** Remove unused lightMarathonClient. ([#2383](https://github.com/traefik/traefik/pull/2383) by [timoreimann](https://github.com/timoreimann))
- **[metrics]** Add InfluxDB support for traefik metrics ([#2289](https://github.com/traefik/traefik/pull/2289) by [adityacs](https://github.com/adityacs))
- **[middleware]** Added ReplacePathRegex middleware ([#2033](https://github.com/traefik/traefik/pull/2033) by [Tiscs](https://github.com/Tiscs))
- **[middleware]** Fix custom headers replacement ([#2455](https://github.com/traefik/traefik/pull/2455) by [mmatur](https://github.com/mmatur))
- **[oxy]** Resync oxy with original repository ([#2451](https://github.com/traefik/traefik/pull/2451) by [Juliens](https://github.com/Juliens))
- **[provider]** Support template as raw string. ([#2413](https://github.com/traefik/traefik/pull/2413) by [ldez](https://github.com/ldez))
- **[rancher]** Run Rancher tests cases in parallel. ([#2424](https://github.com/traefik/traefik/pull/2424) by [ldez](https://github.com/ldez))
- **[rancher]** Update Rancher API integration to go-rancher client v2. ([#2291](https://github.com/traefik/traefik/pull/2291) by [rawmind0](https://github.com/rawmind0))
- **[servicefabric]** Add Service Fabric Provider ([#2117](https://github.com/traefik/traefik/pull/2117) by [lawrencegripper](https://github.com/lawrencegripper))
- **[tls]** Allow adding optional Client CA files ([#2306](https://github.com/traefik/traefik/pull/2306) by [nmengin](https://github.com/nmengin))
- **[websocket]** Add tests for websocket headers ([#2379](https://github.com/traefik/traefik/pull/2379) by [Juliens](https://github.com/Juliens))
- Upgrade libkermit/compose version ([#2071](https://github.com/traefik/traefik/pull/2071) by [nmengin](https://github.com/nmengin))
- Add proxy protocol tests ([#2325](https://github.com/traefik/traefik/pull/2325) by [emilevauge](https://github.com/emilevauge))
- Register pprof handlers. ([#2428](https://github.com/traefik/traefik/pull/2428) by [timoreimann](https://github.com/timoreimann))
- Rate limiting for frontends ([#2034](https://github.com/traefik/traefik/pull/2034) by [bparli](https://github.com/bparli))
- Stats collection. ([#2447](https://github.com/traefik/traefik/pull/2447) by [ldez](https://github.com/ldez))
- Add request accepting grace period delaying graceful shutdown. ([#1971](https://github.com/traefik/traefik/pull/1971) by [timoreimann](https://github.com/timoreimann))
- Put subcommand in dedicated files. ([#2265](https://github.com/traefik/traefik/pull/2265) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[acme,docker]** Modify ACME configuration migration into KV store ([#2598](https://github.com/traefik/traefik/pull/2598) by [nmengin](https://github.com/nmengin))
- **[acme,logs]** Modify DEBUG messages to get ACME certificates ([#2685](https://github.com/traefik/traefik/pull/2685) by [nmengin](https://github.com/nmengin))
- **[acme]** Modify the ACME renewing logs level ([#2520](https://github.com/traefik/traefik/pull/2520) by [nmengin](https://github.com/nmengin))
- **[acme]** ACME and corporate proxy. ([#2738](https://github.com/traefik/traefik/pull/2738) by [ldez](https://github.com/ldez))
- **[acme]** Challenge HTTP must ignore deprecated web.path option ([#2719](https://github.com/traefik/traefik/pull/2719) by [Juliens](https://github.com/Juliens))
- **[api]** Fix pprof route order. ([#2523](https://github.com/traefik/traefik/pull/2523) by [timoreimann](https://github.com/timoreimann))
- **[authentication,middleware]** Fix concurrent map writes on digest auth ([#2695](https://github.com/traefik/traefik/pull/2695) by [mmatur](https://github.com/mmatur))
- **[consulcatalog]** Use prefix for sticky and stickiness tags. ([#2624](https://github.com/traefik/traefik/pull/2624) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Fix bad Træfik update on Consul Catalog ([#2573](https://github.com/traefik/traefik/pull/2573) by [mmatur](https://github.com/mmatur))
- **[consulcatalog]** Reload configuration when port change for one service ([#2574](https://github.com/traefik/traefik/pull/2574) by [mmatur](https://github.com/mmatur))
- **[docker,k8s]** Fix Labels/annotation logs and values. ([#2488](https://github.com/traefik/traefik/pull/2488) by [ldez](https://github.com/ldez))
- **[docker,k8s]** Change custom headers separator ([#2509](https://github.com/traefik/traefik/pull/2509) by [ldez](https://github.com/ldez))
- **[docker]** Fix empty IP for backend when dnsrr in Docker swarm mode ([#2490](https://github.com/traefik/traefik/pull/2490) by [mmatur](https://github.com/mmatur))
- **[docker]** Quote template strings ([#2496](https://github.com/traefik/traefik/pull/2496) by [dtomcej](https://github.com/dtomcej))
- **[docker]** Return errors from Docker client.Events ([#2689](https://github.com/traefik/traefik/pull/2689) by [BlakeMesdag](https://github.com/BlakeMesdag))
- **[docker]** Typo in Docker template. ([#2692](https://github.com/traefik/traefik/pull/2692) by [ldez](https://github.com/ldez))
- **[ecs]** Add missing functions for ECS template ([#2312](https://github.com/traefik/traefik/pull/2312) by [oldmantaiter](https://github.com/oldmantaiter))
- **[file,tls]** Send empty configuration from file provider ([#2609](https://github.com/traefik/traefik/pull/2609) by [nmengin](https://github.com/nmengin))
- **[healthcheck]** Fix health check when web is not specified ([#2529](https://github.com/traefik/traefik/pull/2529) by [Juliens](https://github.com/Juliens))
- **[k8s]** Reduce logs with new Kubernetes security annotations ([#2506](https://github.com/traefik/traefik/pull/2506) by [ldez](https://github.com/ldez))
- **[k8s]** Add missing entry points template. ([#2594](https://github.com/traefik/traefik/pull/2594) by [ldez](https://github.com/ldez))
- **[kv]** Fix stickiness bug due to template syntax error ([#2591](https://github.com/traefik/traefik/pull/2591) by [dahefanteng](https://github.com/dahefanteng))
- **[kv]** List entries parsing. ([#2669](https://github.com/traefik/traefik/pull/2669) by [ldez](https://github.com/ldez))
- **[logs]** Fix traefik logs to behave like configured ([#2176](https://github.com/traefik/traefik/pull/2176) by [marco-jantke](https://github.com/marco-jantke))
- **[marathon]** Update go-marathon ([#2585](https://github.com/traefik/traefik/pull/2585) by [timoreimann](https://github.com/timoreimann))
- **[mesos]** Mesos: Use slave.PID.Host as task SlaveIP. ([#2590](https://github.com/traefik/traefik/pull/2590) by [nemosupremo](https://github.com/nemosupremo))
- **[metrics]** Fix breaking change in web metrics ([#2725](https://github.com/traefik/traefik/pull/2725) by [Juliens](https://github.com/Juliens))
- **[metrics]** Do not ignore web params when web.metrics.prometheus is set ([#2499](https://github.com/traefik/traefik/pull/2499) by [Juliens](https://github.com/Juliens))
- **[metrics]** Fix metrics problem on multiple entrypoints ([#2492](https://github.com/traefik/traefik/pull/2492) by [Juliens](https://github.com/Juliens))
- **[metrics]** Fix data races. ([#2287](https://github.com/traefik/traefik/pull/2287) by [tcolgate](https://github.com/tcolgate))
- **[metrics]** Flaky test Influxdb. ([#2386](https://github.com/traefik/traefik/pull/2386) by [ldez](https://github.com/ldez))
- **[middleware,docker,k8s]** Fix custom headers template ([#2621](https://github.com/traefik/traefik/pull/2621) by [ldez](https://github.com/ldez))
- **[middleware]** Don&#39;t panic if ResponseWriter does not implement CloseNotify ([#2651](https://github.com/traefik/traefik/pull/2651) by [Juliens](https://github.com/Juliens))
- **[middleware]** GzipResponse must implement CloseNotifier if ResponseWriter implement it ([#2657](https://github.com/traefik/traefik/pull/2657) by [Juliens](https://github.com/Juliens))
- **[middleware]** Fix RawPath handling in addPrefix ([#2560](https://github.com/traefik/traefik/pull/2560) by [risdenk](https://github.com/risdenk))
- **[middleware]** We need to flush the end of the body when retry is streamed ([#2644](https://github.com/traefik/traefik/pull/2644) by [Juliens](https://github.com/Juliens))
- **[provider]** Fix typo in frontend.headers.customresponseheaders label ([#2356](https://github.com/traefik/traefik/pull/2356) by [nmandery](https://github.com/nmandery))
- **[provider]** Fix concurrent provider config reloads ([#2276](https://github.com/traefik/traefik/pull/2276) by [marco-jantke](https://github.com/marco-jantke))
- **[rancher]** Don&#39;t reload configuration when rancher server is down ([#2706](https://github.com/traefik/traefik/pull/2706) by [wacken89](https://github.com/wacken89))
- **[rules]** Add non regex pathPrefix ([#2592](https://github.com/traefik/traefik/pull/2592) by [emilevauge](https://github.com/emilevauge))
- **[servicefabric]** Fix backend name for Stateful services. (Service Fabric) ([#2559](https://github.com/traefik/traefik/pull/2559) by [ldez](https://github.com/ldez))
- **[servicefabric]** Fix isHealthy logic. ([#2577](https://github.com/traefik/traefik/pull/2577) by [ldez](https://github.com/ldez))
- **[servicefabric]** Service Fabric &#39;expose&#39; as boolean. ([#2476](https://github.com/traefik/traefik/pull/2476) by [ldez](https://github.com/ldez))
- **[tls]** Allow deleting dynamically all TLS certificates from an entryPoint ([#2603](https://github.com/traefik/traefik/pull/2603) by [nmengin](https://github.com/nmengin))
- **[websocket]** Disable websocket compression ([#2727](https://github.com/traefik/traefik/pull/2727) by [Juliens](https://github.com/Juliens))
- **[websocket]** Add compression and better error handling ([#2702](https://github.com/traefik/traefik/pull/2702) by [Juliens](https://github.com/Juliens))
- **[websocket]** Use gorilla readMessage and writeMessage instead of just an io.Copy ([#2650](https://github.com/traefik/traefik/pull/2650) by [Juliens](https://github.com/Juliens))
- **[websocket]** RawPath and Transfer TLSConfig in websocket ([#2077](https://github.com/traefik/traefik/pull/2077) by [Juliens](https://github.com/Juliens))
- **[zk]** Change Zookeeper default prefix. ([#2580](https://github.com/traefik/traefik/pull/2580) by [ldez](https://github.com/ldez))
- Fix wrong default entry point and non-existing entry point issue ([#2501](https://github.com/traefik/traefik/pull/2501) by [Juliens](https://github.com/Juliens))
- Fix goroutine leak in throttler logic. ([#2739](https://github.com/traefik/traefik/pull/2739) by [timoreimann](https://github.com/timoreimann))
- Fix timeout integration test ([#2679](https://github.com/traefik/traefik/pull/2679) by [ldez](https://github.com/ldez))
- Fix frontend redirect ([#2544](https://github.com/traefik/traefik/pull/2544) by [ldez](https://github.com/ldez))
- Close ring buffer used in throttling function. ([#2532](https://github.com/traefik/traefik/pull/2532) by [timoreimann](https://github.com/timoreimann))

**Documentation:**
- **[acme]** Improve documentation for Cloudflare API key ([#2558](https://github.com/traefik/traefik/pull/2558) by [mmatur](https://github.com/mmatur))
- **[acme]** Update Let&#39;s Encrypt provider list ([#2347](https://github.com/traefik/traefik/pull/2347) by [mmatur](https://github.com/mmatur))
- **[cluster]** Add a clustering example with Docker Swarm ([#2589](https://github.com/traefik/traefik/pull/2589) by [jmaitrehenry](https://github.com/jmaitrehenry))
- **[consul,consulcatalog]** Split Consul and Consul Catalog documentation ([#2654](https://github.com/traefik/traefik/pull/2654) by [ldez](https://github.com/ldez))
- **[consul]** Improve Consul documentation ([#2485](https://github.com/traefik/traefik/pull/2485) by [mmatur](https://github.com/mmatur))
- **[docker/swarm]** Typo in docker.endpoint TCP port. ([#2626](https://github.com/traefik/traefik/pull/2626) by [redhandpl](https://github.com/redhandpl))
- **[docker]** Fix Docker labels documentation render. ([#2505](https://github.com/traefik/traefik/pull/2505) by [ldez](https://github.com/ldez))
- **[docker]** Add a note on how to add label to a docker compose file ([#2611](https://github.com/traefik/traefik/pull/2611) by [jmaitrehenry](https://github.com/jmaitrehenry))
- **[etcd]** Fix typo in examples ([#2446](https://github.com/traefik/traefik/pull/2446) by [dahefanteng](https://github.com/dahefanteng))
- **[k8s]** Add note to Kubernetes RBAC docs about RoleBindings and namespaces ([#2498](https://github.com/traefik/traefik/pull/2498) by [jmara](https://github.com/jmara))
- **[k8s]** k8s guide: Leave note about assumed DaemonSet usage. ([#2634](https://github.com/traefik/traefik/pull/2634) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Apply various contextual and stylish improvements to the k8s docs. ([#2677](https://github.com/traefik/traefik/pull/2677) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Document rewrite-target annotation. ([#2676](https://github.com/traefik/traefik/pull/2676) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Remove obsolete links in k8s docs ([#2465](https://github.com/traefik/traefik/pull/2465) by [marco-jantke](https://github.com/marco-jantke))
- **[k8s]** Document filename parameter for Kubernetes. ([#2464](https://github.com/traefik/traefik/pull/2464) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Improve Marathon service label documentation. ([#2635](https://github.com/traefik/traefik/pull/2635) by [timoreimann](https://github.com/timoreimann))
- **[metrics]** Add entrypoint in Prometheus doc and remove web on Influxdb doc ([#2452](https://github.com/traefik/traefik/pull/2452) by [Juliens](https://github.com/Juliens))
- **[provider,webui]** Fix redirect problem on dashboard + docs/tests on [web] ([#2686](https://github.com/traefik/traefik/pull/2686) by [Juliens](https://github.com/Juliens))
- **[servicefabric]** Describe &#39;refreshSecond&#39; configuration. ([#2471](https://github.com/traefik/traefik/pull/2471) by [ldez](https://github.com/ldez))
- **[tls]** Fix doc dynamic certificates ([#2737](https://github.com/traefik/traefik/pull/2737) by [emilevauge](https://github.com/emilevauge))
- **[tls]** Add link to crypto/tls godoc. ([#2470](https://github.com/traefik/traefik/pull/2470) by [ldez](https://github.com/ldez))
- Move rate limit documentation. ([#2588](https://github.com/traefik/traefik/pull/2588) by [ldez](https://github.com/ldez))
- Grammar ([#2562](https://github.com/traefik/traefik/pull/2562) by [geraldcroes](https://github.com/geraldcroes))
- Fix some doc links ([#2731](https://github.com/traefik/traefik/pull/2731) by [eldondev](https://github.com/eldondev))
- Fix broken links and improve ResponseCodeRatio() description ([#2538](https://github.com/traefik/traefik/pull/2538) by [mvasin](https://github.com/mvasin))
- Fix typo in anonymous usage log message. ([#2711](https://github.com/traefik/traefik/pull/2711) by [Yggdrasil](https://github.com/Yggdrasil))
- Fix typos in changelog ([#2387](https://github.com/traefik/traefik/pull/2387) by [ferhatelmas](https://github.com/ferhatelmas))
- Add mmatur to maintainers ([#2303](https://github.com/traefik/traefik/pull/2303) by [emilevauge](https://github.com/emilevauge))
- Add a note about redirection rule to precise how regex/replacement work. ([#2243](https://github.com/traefik/traefik/pull/2243) by [nmengin](https://github.com/nmengin))
- Add docker things for documentation ([#2020](https://github.com/traefik/traefik/pull/2020) by [tcoupin](https://github.com/tcoupin))
- Prepare release v1.5.0-rc5 ([#2707](https://github.com/traefik/traefik/pull/2707) by [mmatur](https://github.com/mmatur))
- Prepare release v1.5.0-rc4 ([#2656](https://github.com/traefik/traefik/pull/2656) by [Juliens](https://github.com/Juliens))
- Prepare release v1.5.0-rc3 ([#2599](https://github.com/traefik/traefik/pull/2599) by [ldez](https://github.com/ldez))
- Prepare release v1.5.0-rc2 ([#2533](https://github.com/traefik/traefik/pull/2533) by [ldez](https://github.com/ldez))
- Prepare release v1.5.0-rc1 ([#2480](https://github.com/traefik/traefik/pull/2480) by [ldez](https://github.com/ldez))

**Misc:**
- **[acme]** dumpcerts.sh: Fix call to &#34;base64&#34; for Alpine ([#2344](https://github.com/traefik/traefik/pull/2344) by [nknapp](https://github.com/nknapp))
- **[acme]** dumpcerts.sh: fixed sed, extracted domain keys ([#2161](https://github.com/traefik/traefik/pull/2161) by [sjawhar](https://github.com/sjawhar))
- **[etcd,kv,tls]** Add tests for TLS dynamic configuration in ETCD3 ([#2606](https://github.com/traefik/traefik/pull/2606) by [dahefanteng](https://github.com/dahefanteng))
- Upgrade libkermit/compose version ([#2074](https://github.com/traefik/traefik/pull/2074) by [nmengin](https://github.com/nmengin))
- Merge v1.4.6 into v1.5  ([#2642](https://github.com/traefik/traefik/pull/2642) by [ldez](https://github.com/ldez))
- Merge v1.4.5 into v1.5 ([#2530](https://github.com/traefik/traefik/pull/2530) by [mmatur](https://github.com/mmatur))
- Merge current v1.4 into master  ([#2479](https://github.com/traefik/traefik/pull/2479) by [ldez](https://github.com/ldez))
- Merge v1.4.3 into master ([#2415](https://github.com/traefik/traefik/pull/2415) by [ldez](https://github.com/ldez))
- Merge v1.4.4 into master ([#2457](https://github.com/traefik/traefik/pull/2457) by [ldez](https://github.com/ldez))
- Merge v1.4.3 into master ([#2406](https://github.com/traefik/traefik/pull/2406) by [ldez](https://github.com/ldez))
- Revert &#34;Merge v1.4.2 into master&#34; ([#2414](https://github.com/traefik/traefik/pull/2414) by [ldez](https://github.com/ldez))
- Merge v1.4.2 into master ([#2358](https://github.com/traefik/traefik/pull/2358) by [ldez](https://github.com/ldez))
- Merge v1.4.1 into master  ([#2318](https://github.com/traefik/traefik/pull/2318) by [ldez](https://github.com/ldez))
- Merge v1.4.0 ([#2271](https://github.com/traefik/traefik/pull/2271) by [ldez](https://github.com/ldez))
- Merge v1.4.0-rc5 into master  ([#2242](https://github.com/traefik/traefik/pull/2242) by [ldez](https://github.com/ldez))
- Merge v1.4.0-rc4 into master ([#2202](https://github.com/traefik/traefik/pull/2202) by [ldez](https://github.com/ldez))
- Merge current v1.4 into master  ([#2469](https://github.com/traefik/traefik/pull/2469) by [ldez](https://github.com/ldez))
- Merge current v1.4 ([#2154](https://github.com/traefik/traefik/pull/2154) by [ldez](https://github.com/ldez))
- Merge v1.4.0-rc3 into master ([#2140](https://github.com/traefik/traefik/pull/2140) by [ldez](https://github.com/ldez))
- Merge v1.4.0-rc2 into master ([#2092](https://github.com/traefik/traefik/pull/2092) by [ldez](https://github.com/ldez))
- Merge current 1.4 ([#2064](https://github.com/traefik/traefik/pull/2064) by [ldez](https://github.com/ldez))

## [v1.5.0-rc5](https://github.com/traefik/traefik/tree/v1.5.0-rc5) (2018-01-15)
[All Commits](https://github.com/traefik/traefik/compare/v1.5.0-rc4...v1.5.0-rc5)

**Enhancements:**
- **[acme]** Add Let&#39;s Encrypt HTTP Challenge ([#2701](https://github.com/traefik/traefik/pull/2701) by [Juliens](https://github.com/Juliens))

**Bug fixes:**
- **[acme,logs]** Modify DEBUG messages to get ACME certificates ([#2685](https://github.com/traefik/traefik/pull/2685) by [nmengin](https://github.com/nmengin))
- **[authentication,middleware]** Fix concurrent map writes on digest auth ([#2695](https://github.com/traefik/traefik/pull/2695) by [mmatur](https://github.com/mmatur))
- **[docker]** Typo in Docker template. ([#2692](https://github.com/traefik/traefik/pull/2692) by [ldez](https://github.com/ldez))
- **[docker]** Return errors from Docker client.Events ([#2689](https://github.com/traefik/traefik/pull/2689) by [BlakeMesdag](https://github.com/BlakeMesdag))
- **[kv]** List entries parsing. ([#2669](https://github.com/traefik/traefik/pull/2669) by [ldez](https://github.com/ldez))
- **[metrics]** Fix data races. ([#2287](https://github.com/traefik/traefik/pull/2287) by [tcolgate](https://github.com/tcolgate))
- **[middleware]** GzipResponse must implement CloseNotifier if ResponseWriter implement it ([#2657](https://github.com/traefik/traefik/pull/2657) by [Juliens](https://github.com/Juliens))
- **[websocket]** Add compression and better error handling ([#2702](https://github.com/traefik/traefik/pull/2702) by [Juliens](https://github.com/Juliens))
- Fix: timeout integration test ([#2679](https://github.com/traefik/traefik/pull/2679) by [ldez](https://github.com/ldez))

**Documentation:**
- **[cluster]** Add a clustering example with Docker Swarm ([#2589](https://github.com/traefik/traefik/pull/2589) by [jmaitrehenry](https://github.com/jmaitrehenry))
- **[k8s]** Apply various contextual and stylish improvements to the k8s docs. ([#2677](https://github.com/traefik/traefik/pull/2677) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Document rewrite-target annotation. ([#2676](https://github.com/traefik/traefik/pull/2676) by [timoreimann](https://github.com/timoreimann))
- **[provider,webui]** Fix redirect problem on dashboard + docs/tests on [web] ([#2686](https://github.com/traefik/traefik/pull/2686) by [Juliens](https://github.com/Juliens))

## [v1.5.0-rc4](https://github.com/traefik/traefik/tree/v1.5.0-rc4) (2018-01-04)
[All Commits](https://github.com/traefik/traefik/compare/v1.5.0-rc3...v1.5.0-rc4)

**Bug fixes:**
- **[consulcatalog]** Use prefix for sticky and stickiness tags. ([#2624](https://github.com/traefik/traefik/pull/2624) by [ldez](https://github.com/ldez))
- **[file,tls]** Send empty configuration from file provider ([#2609](https://github.com/traefik/traefik/pull/2609) by [nmengin](https://github.com/nmengin))
- **[middleware,docker,k8s]** Fix custom headers template ([#2621](https://github.com/traefik/traefik/pull/2621) by [ldez](https://github.com/ldez))
- **[middleware]** Don&#39;t panic if ResponseWriter does not implement CloseNotify ([#2651](https://github.com/traefik/traefik/pull/2651) by [Juliens](https://github.com/Juliens))
- **[middleware]** We need to flush the end of the body when retry is streamed ([#2644](https://github.com/traefik/traefik/pull/2644) by [Juliens](https://github.com/Juliens))
- **[tls]** Allow deleting dynamically all TLS certificates from an entryPoint ([#2603](https://github.com/traefik/traefik/pull/2603) by [nmengin](https://github.com/nmengin))
- **[websocket]** Use gorilla readMessage and writeMessage instead of just an io.Copy ([#2650](https://github.com/traefik/traefik/pull/2650) by [Juliens](https://github.com/Juliens))

**Documentation:**
- **[consul,consulcatalog]** Split Consul and Consul Catalog documentation ([#2654](https://github.com/traefik/traefik/pull/2654) by [ldez](https://github.com/ldez))
- **[docker/swarm]** Typo in docker.endpoint TCP port. ([#2626](https://github.com/traefik/traefik/pull/2626) by [redhandpl](https://github.com/redhandpl))
- **[docker]** Add a note on how to add label to a docker compose file ([#2611](https://github.com/traefik/traefik/pull/2611) by [jmaitrehenry](https://github.com/jmaitrehenry))
- **[k8s]** k8s guide: Leave note about assumed DaemonSet usage. ([#2634](https://github.com/traefik/traefik/pull/2634) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Improve Marathon service label documentation. ([#2635](https://github.com/traefik/traefik/pull/2635) by [timoreimann](https://github.com/timoreimann))

**Misc:**
- **[etcd,kv,tls]** Add tests for TLS dynamic configuration in ETCD3 ([#2606](https://github.com/traefik/traefik/pull/2606) by [dahefanteng](https://github.com/dahefanteng))
- Merge v1.4.6 into v1.5  ([#2642](https://github.com/traefik/traefik/pull/2642) by [ldez](https://github.com/ldez))

## [v1.4.6](https://github.com/traefik/traefik/tree/v1.4.6) (2018-01-02)
[All Commits](https://github.com/traefik/traefik/compare/v1.4.5...v1.4.6)

**Bug fixes:**
- **[docker]** Normalize serviceName added to the service backend names ([#2631](https://github.com/traefik/traefik/pull/2631) by [mmatur](https://github.com/mmatur))
- **[websocket]** Use gorilla readMessage and writeMessage instead of just an io.Copy ([#2640](https://github.com/traefik/traefik/pull/2640) by [Juliens](https://github.com/Juliens))
- Fix bug report command ([#2638](https://github.com/traefik/traefik/pull/2638) by [ldez](https://github.com/ldez))

## [v1.5.0-rc3](https://github.com/traefik/traefik/tree/v1.5.0-rc3) (2017-12-20)
[All Commits](https://github.com/traefik/traefik/compare/v1.5.0-rc2...v1.5.0-rc3)

**Enhancements:**
- **[docker,k8s,rancher]** Support regex redirect by frontend ([#2570](https://github.com/traefik/traefik/pull/2570) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[acme,docker]** Modify ACME configuration migration into KV store ([#2598](https://github.com/traefik/traefik/pull/2598) by [nmengin](https://github.com/nmengin))
- **[consulcatalog]** Reload configuration when port change for one service ([#2574](https://github.com/traefik/traefik/pull/2574) by [mmatur](https://github.com/mmatur))
- **[consulcatalog]** Fix bad Træfik update on Consul Catalog ([#2573](https://github.com/traefik/traefik/pull/2573) by [mmatur](https://github.com/mmatur))
- **[k8s]** Add missing entrypoints template. ([#2594](https://github.com/traefik/traefik/pull/2594) by [ldez](https://github.com/ldez))
- **[kv]** Fix stickiness bug due to template syntax error ([#2591](https://github.com/traefik/traefik/pull/2591) by [dahefanteng](https://github.com/dahefanteng))
- **[marathon]** Update go-marathon ([#2585](https://github.com/traefik/traefik/pull/2585) by [timoreimann](https://github.com/timoreimann))
- **[mesos]** Mesos: Use slave.PID.Host as task SlaveIP. ([#2590](https://github.com/traefik/traefik/pull/2590) by [nemosupremo](https://github.com/nemosupremo))
- **[middleware]** Fix RawPath handling in addPrefix ([#2560](https://github.com/traefik/traefik/pull/2560) by [risdenk](https://github.com/risdenk))
- **[rules]** Add non regex pathPrefix ([#2592](https://github.com/traefik/traefik/pull/2592) by [emilevauge](https://github.com/emilevauge))
- **[servicefabric]** Fix backend name for Stateful services. (Service Fabric) ([#2559](https://github.com/traefik/traefik/pull/2559) by [ldez](https://github.com/ldez))
- **[servicefabric]** Fix isHealthy logic. ([#2577](https://github.com/traefik/traefik/pull/2577) by [ldez](https://github.com/ldez))
- **[zk]** Change Zookeeper default prefix. ([#2580](https://github.com/traefik/traefik/pull/2580) by [ldez](https://github.com/ldez))
- Fix frontend redirect ([#2544](https://github.com/traefik/traefik/pull/2544) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme]** Improve documentation for Cloudflare API key ([#2558](https://github.com/traefik/traefik/pull/2558) by [mmatur](https://github.com/mmatur))
- Move rate limit documentation. ([#2588](https://github.com/traefik/traefik/pull/2588) by [ldez](https://github.com/ldez))
- Grammar ([#2562](https://github.com/traefik/traefik/pull/2562) by [geraldcroes](https://github.com/geraldcroes))
- Fix broken links and improve ResponseCodeRatio() description ([#2538](https://github.com/traefik/traefik/pull/2538) by [mvasin](https://github.com/mvasin))

## [v1.5.0-rc2](https://github.com/traefik/traefik/tree/v1.5.0-rc2) (2017-12-06)
[All Commits](https://github.com/traefik/traefik/compare/v1.5.0-rc1...v1.5.0-rc2)

**Bug fixes:**
- **[acme]** Modify the ACME renewing logs level ([#2520](https://github.com/traefik/traefik/pull/2520) by [nmengin](https://github.com/nmengin))
- **[api]** Fix pprof route order. ([#2523](https://github.com/traefik/traefik/pull/2523) by [timoreimann](https://github.com/timoreimann))
- **[docker,k8s]** Change custom headers separator ([#2509](https://github.com/traefik/traefik/pull/2509) by [ldez](https://github.com/ldez))
- **[docker,k8s]** Fix Labels/annotation logs and values. ([#2488](https://github.com/traefik/traefik/pull/2488) by [ldez](https://github.com/ldez))
- **[docker]** Quote template strings ([#2496](https://github.com/traefik/traefik/pull/2496) by [dtomcej](https://github.com/dtomcej))
- **[docker]** Fix empty IP for backend when dnsrr in Docker swarm mode ([#2490](https://github.com/traefik/traefik/pull/2490) by [mmatur](https://github.com/mmatur))
- **[healthcheck]** Fix healthcheck when web is not specified ([#2529](https://github.com/traefik/traefik/pull/2529) by [Juliens](https://github.com/Juliens))
- **[k8s]** Reduce logs with new Kubernetes security annotations ([#2506](https://github.com/traefik/traefik/pull/2506) by [ldez](https://github.com/ldez))
- **[metrics]** Do not ignore web params when web.metrics.prometheus is set ([#2499](https://github.com/traefik/traefik/pull/2499) by [Juliens](https://github.com/Juliens))
- **[metrics]** Fix metrics problem on multiple entrypoints ([#2492](https://github.com/traefik/traefik/pull/2492) by [Juliens](https://github.com/Juliens))
- Close ring buffer used in throttling function. ([#2532](https://github.com/traefik/traefik/pull/2532) by [timoreimann](https://github.com/timoreimann))
- Fix wrong default entrypoint and non-existing entrypoint issue ([#2501](https://github.com/traefik/traefik/pull/2501) by [Juliens](https://github.com/Juliens))

**Documentation:**
- **[consul]** Improve Consul documentation ([#2485](https://github.com/traefik/traefik/pull/2485) by [mmatur](https://github.com/mmatur))
- **[docker]** Fix Docker labels documentation render. ([#2505](https://github.com/traefik/traefik/pull/2505) by [ldez](https://github.com/ldez))
- **[k8s]** Add note to Kubernetes RBAC docs about RoleBindings and namespaces ([#2498](https://github.com/traefik/traefik/pull/2498) by [jmara](https://github.com/jmara))

**Misc:**
- Merge v1.4.5 into v1.5 ([#2530](https://github.com/traefik/traefik/pull/2530) by [mmatur](https://github.com/mmatur))

## [v1.4.5](https://github.com/traefik/traefik/tree/v1.4.5) (2017-12-05)
[All Commits](https://github.com/traefik/traefik/compare/v1.4.4...v1.4.5)

**Bug fixes:**
- **[docker]** Fix empty ip when container is stopped ([#2478](https://github.com/traefik/traefik/pull/2478) by [mmatur](https://github.com/mmatur))
- **[k8s]** Fix kubernetes path prefix rule with rewrite-target ([#2461](https://github.com/traefik/traefik/pull/2461) by [cheungpat](https://github.com/cheungpat))

**Documentation:**
- **[file]** Emphasize the necessity of enabling file backend ([#2483](https://github.com/traefik/traefik/pull/2483) by [mvasin](https://github.com/mvasin))
- Add link to future 1.5 documentation. ([#2477](https://github.com/traefik/traefik/pull/2477) by [ldez](https://github.com/ldez))

## [v1.5.0-rc1](https://github.com/traefik/traefik/tree/v1.5.0-rc1) (2017-11-28)
[All Commits](https://github.com/traefik/traefik/compare/v1.4.0-rc1...v1.5.0-rc1)

**Enhancements:**
- **[acme,provider,docker,tls]** Make the TLS certificates management dynamic. ([#2233](https://github.com/traefik/traefik/pull/2233) by [nmengin](https://github.com/nmengin))
- **[acme]** Update github.com/xenolf/lego to 0.4.1 ([#2304](https://github.com/traefik/traefik/pull/2304) by [oldmantaiter](https://github.com/oldmantaiter))
- **[api,healthcheck,metrics,provider,webui]** Split Web into API/Dashboard, ping, metric and Rest Provider ([#2335](https://github.com/traefik/traefik/pull/2335) by [Juliens](https://github.com/Juliens))
- **[authentication]** Pass through certain forward auth negative response headers ([#2127](https://github.com/traefik/traefik/pull/2127) by [wheresmysocks](https://github.com/wheresmysocks))
- **[cluster,consul,file]** Add file to storeconfig ([#2419](https://github.com/traefik/traefik/pull/2419) by [emilevauge](https://github.com/emilevauge))
- **[cluster,provider]** Support Etcd v3, enhance KV support ([#2407](https://github.com/traefik/traefik/pull/2407) by [nmengin](https://github.com/nmengin))
- **[docker,k8s,rancher,webui]** redirect to another entryPoint per frontend ([#2133](https://github.com/traefik/traefik/pull/2133) by [SantoDE](https://github.com/SantoDE))
- **[docker]** Add Custom header parsing to Docker Provider ([#2030](https://github.com/traefik/traefik/pull/2030) by [dtomcej](https://github.com/dtomcej))
- **[docker]** Docker labels ([#2473](https://github.com/traefik/traefik/pull/2473) by [ldez](https://github.com/ldez))
- **[docker]** Add docker security headers via labels ([#2334](https://github.com/traefik/traefik/pull/2334) by [dtomcej](https://github.com/dtomcej))
- **[docker]** Use Node IP in Swarm Standalone with &#34;host&#34; NetworkMode ([#2274](https://github.com/traefik/traefik/pull/2274) by [BlakeMesdag](https://github.com/BlakeMesdag))
- **[ecs]** ECS provider refactoring ([#2050](https://github.com/traefik/traefik/pull/2050) by [mmatur](https://github.com/mmatur))
- **[ecs]** Add health check label to ECS ([#2421](https://github.com/traefik/traefik/pull/2421) by [oldmantaiter](https://github.com/oldmantaiter))
- **[ecs]** Support Host NetworkMode  for ECS provider  ([#2320](https://github.com/traefik/traefik/pull/2320) by [FriggaHel](https://github.com/FriggaHel))
- **[etcd]** Manage certificates dynamically in kv store ([#2411](https://github.com/traefik/traefik/pull/2411) by [dahefanteng](https://github.com/dahefanteng))
- **[healthcheck]** Use healthcheck for systemd watchdog ([#2283](https://github.com/traefik/traefik/pull/2283) by [guilhem](https://github.com/guilhem))
- **[k8s]** Kubernetes security header annotations ([#2460](https://github.com/traefik/traefik/pull/2460) by [dtomcej](https://github.com/dtomcej))
- **[k8s]** Add labels for `traefik.frontend.entryPoints` &amp; `PassTLSCert` to Kubernetes ([#2324](https://github.com/traefik/traefik/pull/2324) by [ryarnyah](https://github.com/ryarnyah))
- **[k8s]** Only listen to configured k8s namespaces. ([#1895](https://github.com/traefik/traefik/pull/1895) by [timoreimann](https://github.com/timoreimann))
- **[logs,middleware,consul,docker]** Use constants from http package. ([#2425](https://github.com/traefik/traefik/pull/2425) by [ldez](https://github.com/ldez))
- **[logs]** Add json format support for Traefik logs ([#2056](https://github.com/traefik/traefik/pull/2056) by [marco-jantke](https://github.com/marco-jantke))
- **[marathon]** Marathon constraints filtering ([#2388](https://github.com/traefik/traefik/pull/2388) by [aantono](https://github.com/aantono))
- **[marathon]** Remove unused lightMarathonClient. ([#2383](https://github.com/traefik/traefik/pull/2383) by [timoreimann](https://github.com/timoreimann))
- **[metrics]** Add InfluxDB support for traefik metrics ([#2289](https://github.com/traefik/traefik/pull/2289) by [adityacs](https://github.com/adityacs))
- **[middleware]** Added ReplacePathRegex middleware ([#2033](https://github.com/traefik/traefik/pull/2033) by [Tiscs](https://github.com/Tiscs))
- **[middleware]** Fix custom headers replacement ([#2455](https://github.com/traefik/traefik/pull/2455) by [mmatur](https://github.com/mmatur))
- **[oxy]** Resync oxy with original repository ([#2451](https://github.com/traefik/traefik/pull/2451) by [Juliens](https://github.com/Juliens))
- **[provider]** Support template as raw string. ([#2413](https://github.com/traefik/traefik/pull/2413) by [ldez](https://github.com/ldez))
- **[rancher]** Run Rancher tests cases in parallel. ([#2424](https://github.com/traefik/traefik/pull/2424) by [ldez](https://github.com/ldez))
- **[rancher]** Update Rancher API integration to go-rancher client v2. ([#2291](https://github.com/traefik/traefik/pull/2291) by [rawmind0](https://github.com/rawmind0))
- **[servicefabric]** Add Service Fabric Provider ([#2117](https://github.com/traefik/traefik/pull/2117) by [lawrencegripper](https://github.com/lawrencegripper))
- **[tls]** Allow adding optional Client CA files ([#2306](https://github.com/traefik/traefik/pull/2306) by [nmengin](https://github.com/nmengin))
- **[websocket]** Add tests for websocket headers ([#2379](https://github.com/traefik/traefik/pull/2379) by [Juliens](https://github.com/Juliens))
- Upgrade libkermit/compose version ([#2071](https://github.com/traefik/traefik/pull/2071) by [nmengin](https://github.com/nmengin))
- Add proxy protocol tests ([#2325](https://github.com/traefik/traefik/pull/2325) by [emilevauge](https://github.com/emilevauge))
- Register pprof handlers. ([#2428](https://github.com/traefik/traefik/pull/2428) by [timoreimann](https://github.com/timoreimann))
- Rate limiting for frontends ([#2034](https://github.com/traefik/traefik/pull/2034) by [bparli](https://github.com/bparli))
- Stats collection. ([#2447](https://github.com/traefik/traefik/pull/2447) by [ldez](https://github.com/ldez))
- Add request accepting grace period delaying graceful shutdown. ([#1971](https://github.com/traefik/traefik/pull/1971) by [timoreimann](https://github.com/timoreimann))
- Put subcommand in dedicated files. ([#2265](https://github.com/traefik/traefik/pull/2265) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[ecs]** Add missing functions for ECS template ([#2312](https://github.com/traefik/traefik/pull/2312) by [oldmantaiter](https://github.com/oldmantaiter))
- **[logs]** Fix traefik logs to behave like configured ([#2176](https://github.com/traefik/traefik/pull/2176) by [marco-jantke](https://github.com/marco-jantke))
- **[metrics]** Flaky test Influxdb. ([#2386](https://github.com/traefik/traefik/pull/2386) by [ldez](https://github.com/ldez))
- **[provider]** Fix typo in frontend.headers.customresponseheaders label ([#2356](https://github.com/traefik/traefik/pull/2356) by [nmandery](https://github.com/nmandery))
- **[provider]** fix concurrent provider config reloads ([#2276](https://github.com/traefik/traefik/pull/2276) by [marco-jantke](https://github.com/marco-jantke))
- **[servicefabric]** Service Fabric &#39;expose&#39; as boolean. ([#2476](https://github.com/traefik/traefik/pull/2476) by [ldez](https://github.com/ldez))
- **[websocket]** RawPath and Transfer TLSConfig in websocket ([#2077](https://github.com/traefik/traefik/pull/2077) by [Juliens](https://github.com/Juliens))

**Documentation:**
- **[acme]** Update Let&#39;s Encrypt provider list ([#2347](https://github.com/traefik/traefik/pull/2347) by [mmatur](https://github.com/mmatur))
- **[etcd]** Fix typo in examples ([#2446](https://github.com/traefik/traefik/pull/2446) by [dahefanteng](https://github.com/dahefanteng))
- **[k8s]** Remove obsolete links in k8s docs ([#2465](https://github.com/traefik/traefik/pull/2465) by [marco-jantke](https://github.com/marco-jantke))
- **[k8s]** Document filename parameter for Kubernetes. ([#2464](https://github.com/traefik/traefik/pull/2464) by [timoreimann](https://github.com/timoreimann))
- **[metrics]** Add entrypoint in Prometheus doc and remove web on Influxdb doc ([#2452](https://github.com/traefik/traefik/pull/2452) by [Juliens](https://github.com/Juliens))
- **[servicefabric]** Describe &#39;refreshSecond&#39; configuration. ([#2471](https://github.com/traefik/traefik/pull/2471) by [ldez](https://github.com/ldez))
- **[tls]** Add link to crypto/tls godoc. ([#2470](https://github.com/traefik/traefik/pull/2470) by [ldez](https://github.com/ldez))
- Fix typos in changelog ([#2387](https://github.com/traefik/traefik/pull/2387) by [ferhatelmas](https://github.com/ferhatelmas))
- Add mmatur to maintainers ([#2303](https://github.com/traefik/traefik/pull/2303) by [emilevauge](https://github.com/emilevauge))
- Add a note about redirection rule to precise how regex/replacement work. ([#2243](https://github.com/traefik/traefik/pull/2243) by [nmengin](https://github.com/nmengin))
- Add docker things for documentation ([#2020](https://github.com/traefik/traefik/pull/2020) by [tcoupin](https://github.com/tcoupin))

**Misc:**
- **[acme]** dumpcerts.sh: Fix call to &#34;base64&#34; for Alpine ([#2344](https://github.com/traefik/traefik/pull/2344) by [nknapp](https://github.com/nknapp))
- **[acme]** Dumpcerts.sh: fixed sed, extracted domain keys ([#2161](https://github.com/traefik/traefik/pull/2161) by [sjawhar](https://github.com/sjawhar))
- Merge current v1.4 into master  ([#2469](https://github.com/traefik/traefik/pull/2469) by [ldez](https://github.com/ldez))
- Revert &#34;Merge v1.4.2 into master&#34; ([#2414](https://github.com/traefik/traefik/pull/2414) by [ldez](https://github.com/ldez))
-  Merge v1.4.3 into master ([#2406](https://github.com/traefik/traefik/pull/2406) by [ldez](https://github.com/ldez))
- Merge v1.4.2 into master ([#2358](https://github.com/traefik/traefik/pull/2358) by [ldez](https://github.com/ldez))
- Merge v1.4.3 into master ([#2415](https://github.com/traefik/traefik/pull/2415) by [ldez](https://github.com/ldez))
- Merge v1.4.1 into master  ([#2318](https://github.com/traefik/traefik/pull/2318) by [ldez](https://github.com/ldez))
- Merge v1.4.0 ([#2271](https://github.com/traefik/traefik/pull/2271) by [ldez](https://github.com/ldez))
- Merge v1.4.0-rc5 into master  ([#2242](https://github.com/traefik/traefik/pull/2242) by [ldez](https://github.com/ldez))
- Merge v1.4.0-rc4 into master ([#2202](https://github.com/traefik/traefik/pull/2202) by [ldez](https://github.com/ldez))
- Merge v1.4.4 into master ([#2457](https://github.com/traefik/traefik/pull/2457) by [ldez](https://github.com/ldez))
- Merge current v1.4 ([#2154](https://github.com/traefik/traefik/pull/2154) by [ldez](https://github.com/ldez))
- Merge v1.4.0-rc3 into master ([#2140](https://github.com/traefik/traefik/pull/2140) by [ldez](https://github.com/ldez))
- Merge v1.4.0-rc2 into master ([#2092](https://github.com/traefik/traefik/pull/2092) by [ldez](https://github.com/ldez))
- Upgrade libkermit/compose version ([#2074](https://github.com/traefik/traefik/pull/2074) by [nmengin](https://github.com/nmengin))
- Merge current 1.4 ([#2064](https://github.com/traefik/traefik/pull/2064) by [ldez](https://github.com/ldez))

## [v1.4.4](https://github.com/traefik/traefik/tree/v1.4.4) (2017-11-21)
[All Commits](https://github.com/traefik/traefik/compare/v1.4.3...v1.4.4)

**Enhancements:**
- **[middleware]** Remove GzipHandler Fork ([#2436](https://github.com/traefik/traefik/pull/2436) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[docker]** Fix problems about duplicated and missing Docker backends/frontends. ([#2434](https://github.com/traefik/traefik/pull/2434) by [nmengin](https://github.com/nmengin))
- **[middleware]** Fix raw path handling in strip prefix ([#2382](https://github.com/traefik/traefik/pull/2382) by [marco-jantke](https://github.com/marco-jantke))
- **[rancher]** Fix issue with label traefik.backend.loadbalancer.stickiness.cookieName ([#2423](https://github.com/traefik/traefik/pull/2423) by [rawmind0](https://github.com/rawmind0))
- http.Server log goes to Debug level. ([#2420](https://github.com/traefik/traefik/pull/2420) by [ldez](https://github.com/ldez))

**Documentation:**
- Documentation archive ([#2405](https://github.com/traefik/traefik/pull/2405) by [ldez](https://github.com/ldez))

## [v1.4.3](https://github.com/traefik/traefik/tree/v1.4.3) (2017-11-14)
[All Commits](https://github.com/traefik/traefik/compare/v1.4.2...v1.4.3)

**Bug fixes:**
- **[consulcatalog]** Fix Traefik reload if Consul Catalog tags change ([#2389](https://github.com/traefik/traefik/pull/2389) by [mmatur](https://github.com/mmatur))
- **[kv]** Add Traefik prefix to the KV key ([#2400](https://github.com/traefik/traefik/pull/2400) by [nmengin](https://github.com/nmengin))
- **[middleware]** Flush and Status code ([#2403](https://github.com/traefik/traefik/pull/2403) by [ldez](https://github.com/ldez))
- **[middleware]** Exclude GRPC from compress ([#2391](https://github.com/traefik/traefik/pull/2391) by [ldez](https://github.com/ldez))
- **[middleware]** Keep status when stream mode and compress ([#2380](https://github.com/traefik/traefik/pull/2380) by [Juliens](https://github.com/Juliens))

**Documentation:**
- **[acme]** Fix some typos ([#2363](https://github.com/traefik/traefik/pull/2363) by [tomsaleeba](https://github.com/tomsaleeba))
- **[docker]** Minor fix for docker volume vs created directory ([#2372](https://github.com/traefik/traefik/pull/2372) by [visibilityspots](https://github.com/visibilityspots))
- **[k8s]** Link corrected ([#2385](https://github.com/traefik/traefik/pull/2385) by [xlazex](https://github.com/xlazex))

**Misc:**
- **[k8s]** Add secret creation to docs for kubernetes backend ([#2374](https://github.com/traefik/traefik/pull/2374) by [shadycuz](https://github.com/shadycuz))

## [v1.4.2](https://github.com/traefik/traefik/tree/v1.4.2) (2017-11-02)
[All Commits](https://github.com/traefik/traefik/compare/v1.4.1...v1.4.2)

**Bug fixes:**
- **[cluster]** Fix datastore corruption on reload due to shrinking config size ([#2340](https://github.com/traefik/traefik/pull/2340) by [else](https://github.com/else))
- **[docker,docker/swarm]** Make frontend names differents for similar routes ([#2338](https://github.com/traefik/traefik/pull/2338) by [nmengin](https://github.com/nmengin))
- **[docker]** Fix IP address when Docker container network mode is container ([#2331](https://github.com/traefik/traefik/pull/2331) by [nmengin](https://github.com/nmengin))
- **[docker]** Make the traefik.port label optional when using service labels in Docker containers. ([#2330](https://github.com/traefik/traefik/pull/2330) by [nmengin](https://github.com/nmengin))
- **[docker]** Add unique ID to Docker services replicas ([#2314](https://github.com/traefik/traefik/pull/2314) by [nmengin](https://github.com/nmengin))
- **[marathon]** Missing Backend key in configuration when application has no tasks ([#2333](https://github.com/traefik/traefik/pull/2333) by [aantono](https://github.com/aantono))
- Remove hardcoded runtime.GOMAXPROCS. ([#2317](https://github.com/traefik/traefik/pull/2317) by [ldez](https://github.com/ldez))

**Documentation:**
- **[k8s]** fixed dead link in kubernetes backend config docs ([#2337](https://github.com/traefik/traefik/pull/2337) by [perplexa](https://github.com/perplexa))
- **[k8s]** Fix the k8s docs example deployment yaml ([#2308](https://github.com/traefik/traefik/pull/2308) by [gnur](https://github.com/gnur))
- Minor grammar change ([#2350](https://github.com/traefik/traefik/pull/2350) by [haxorjim](https://github.com/haxorjim))
- Minor typo ([#2343](https://github.com/traefik/traefik/pull/2343) by [burningTyger](https://github.com/burningTyger))

## [v1.4.1](https://github.com/traefik/traefik/tree/v1.4.1) (2017-10-24)
[All Commits](https://github.com/traefik/traefik/compare/v1.4.0...v1.4.1)

**Bug fixes:**
- **[docker]** Network filter ([#2301](https://github.com/traefik/traefik/pull/2301) by [ldez](https://github.com/ldez))
- **[healthcheck]** Fix healthcheck path ([#2295](https://github.com/traefik/traefik/pull/2295) by [emilevauge](https://github.com/emilevauge))
- **[rules]** Regex capturing group. ([#2296](https://github.com/traefik/traefik/pull/2296) by [ldez](https://github.com/ldez))
- **[websocket]** Force http/1.1 for websocket ([#2292](https://github.com/traefik/traefik/pull/2292) by [Juliens](https://github.com/Juliens))
- Stream mode when http2 ([#2309](https://github.com/traefik/traefik/pull/2309) by [Juliens](https://github.com/Juliens))
- Enhance Trust Forwarded Headers ([#2302](https://github.com/traefik/traefik/pull/2302) by [ldez](https://github.com/ldez))

## [v1.4.0](https://github.com/traefik/traefik/tree/v1.4.0) (2017-10-16)
[All Commits](https://github.com/traefik/traefik/compare/v1.3.0-rc1...v1.4.0)

**Enhancements:**
- **[acme]** Display Traefik logs in integration tests ([#2114](https://github.com/traefik/traefik/pull/2114) by [ldez](https://github.com/ldez))
- **[acme]** Make the ACME developments testing easier ([#1769](https://github.com/traefik/traefik/pull/1769) by [nmengin](https://github.com/nmengin))
- **[acme]** contrib: Dump keys/certs from acme.json to files ([#1484](https://github.com/traefik/traefik/pull/1484) by [brianredbeard](https://github.com/brianredbeard))
- **[api]** Add HTTP HEAD handling to /ping endpoint ([#1768](https://github.com/traefik/traefik/pull/1768) by [martinbaillie](https://github.com/martinbaillie))
- **[authentication,consulcatalog]** Add Basic auth for consul catalog ([#2027](https://github.com/traefik/traefik/pull/2027) by [mmatur](https://github.com/mmatur))
- **[authentication,marathon]** Add marathon label to configure basic auth ([#1799](https://github.com/traefik/traefik/pull/1799) by [nikore](https://github.com/nikore))
- **[authentication,ecs]** Add basic auth for ecs ([#2026](https://github.com/traefik/traefik/pull/2026) by [mmatur](https://github.com/mmatur))
- **[authentication,middleware]** Add forward authentication option ([#1972](https://github.com/traefik/traefik/pull/1972) by [drampelt](https://github.com/drampelt))
- **[authentication]** Manage Headers for the Authentication forwarding. ([#2132](https://github.com/traefik/traefik/pull/2132) by [ldez](https://github.com/ldez))
- **[consulcatalog,sticky-session]** Enable loadbalancer.sticky for Consul Catalog ([#1917](https://github.com/traefik/traefik/pull/1917) by [nbonneval](https://github.com/nbonneval))
- **[consulcatalog]** Exposed by default feature in Consul Catalog ([#2006](https://github.com/traefik/traefik/pull/2006) by [mmatur](https://github.com/mmatur))
- **[consulcatalog]** Speeding up consul catalog health change detection ([#1694](https://github.com/traefik/traefik/pull/1694) by [vholovko](https://github.com/vholovko))
- **[consulcatalog]** Enhanced flexibility in Consul Catalog configuration ([#1565](https://github.com/traefik/traefik/pull/1565) by [aantono](https://github.com/aantono))
- **[docker,k8s]** IP Whitelists for Frontend (with Docker- &amp; Kubernetes-Provider Support) ([#1332](https://github.com/traefik/traefik/pull/1332) by [MaZderMind](https://github.com/MaZderMind))
- **[ecs,sticky-session]** Enable loadbalancer.sticky for ECS ([#1925](https://github.com/traefik/traefik/pull/1925) by [mmatur](https://github.com/mmatur))
- **[ecs]** Add support for several ECS backends ([#1913](https://github.com/traefik/traefik/pull/1913) by [mmatur](https://github.com/mmatur))
- **[file]** Allow file provider to load service config from files in a directory. ([#1672](https://github.com/traefik/traefik/pull/1672) by [rjshep](https://github.com/rjshep))
- **[healthcheck]** Add healthcheck command ([#1982](https://github.com/traefik/traefik/pull/1982) by [emilevauge](https://github.com/emilevauge))
- **[healthcheck]** Allow overriding the port used for healthchecks ([#1567](https://github.com/traefik/traefik/pull/1567) by [bakins](https://github.com/bakins))
- **[k8s,rules]** kubernetes ingress rewrite-target implementation ([#1723](https://github.com/traefik/traefik/pull/1723) by [mlaccetti](https://github.com/mlaccetti))
- **[k8s]** Added ability to override frontend priority for k8s ingress router ([#1874](https://github.com/traefik/traefik/pull/1874) by [DiverOfDark](https://github.com/DiverOfDark))
- **[kv]** Adds definitions to backend kv template for health checking ([#1644](https://github.com/traefik/traefik/pull/1644) by [zachomedia](https://github.com/zachomedia))
- **[logs,dynamodb,ecs,marathon]** Link some providers logs to Traefik ([#1746](https://github.com/traefik/traefik/pull/1746) by [ldez](https://github.com/ldez))
- **[logs,marathon]** remove confusing go-marathon log message ([#1810](https://github.com/traefik/traefik/pull/1810) by [marco-jantke](https://github.com/marco-jantke))
- **[logs]** Send traefik logs to stdout instead stderr ([#2054](https://github.com/traefik/traefik/pull/2054) by [marco-jantke](https://github.com/marco-jantke))
- **[logs]** enable logging to stdout for access logs ([#1683](https://github.com/traefik/traefik/pull/1683) by [marco-jantke](https://github.com/marco-jantke))
- **[logs]** Logs &amp; errors review ([#1673](https://github.com/traefik/traefik/pull/1673) by [ldez](https://github.com/ldez))
- **[logs]** Switch access logging to logrus ([#1647](https://github.com/traefik/traefik/pull/1647) by [rjshep](https://github.com/rjshep))
- **[logs]** log X-Forwarded-For as ClientHost if present ([#1946](https://github.com/traefik/traefik/pull/1946) by [mildis](https://github.com/mildis))
- **[logs]** Restore: First stage of access logging middleware. ([#1571](https://github.com/traefik/traefik/pull/1571) by [ldez](https://github.com/ldez))
- **[logs]** Add log file close and reopen on receipt of SIGUSR1 ([#1761](https://github.com/traefik/traefik/pull/1761) by [rjshep](https://github.com/rjshep))
- **[logs]** add RetryAttempts to AccessLog in JSON format ([#1793](https://github.com/traefik/traefik/pull/1793) by [marco-jantke](https://github.com/marco-jantke))
- **[logs]** Add JSON as access logging format ([#1669](https://github.com/traefik/traefik/pull/1669) by [rjshep](https://github.com/rjshep))
- **[marathon]** Support multi-port service routing for containers running on Marathon ([#1742](https://github.com/traefik/traefik/pull/1742) by [aantono](https://github.com/aantono))
- **[marathon]** Improve Marathon integration tests. ([#1406](https://github.com/traefik/traefik/pull/1406) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Exported getSubDomain function from Marathon provider ([#1693](https://github.com/traefik/traefik/pull/1693) by [aantono](https://github.com/aantono))
- **[marathon]** Use test builder. ([#1871](https://github.com/traefik/traefik/pull/1871) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Add support for readiness checks. ([#1883](https://github.com/traefik/traefik/pull/1883) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Move marathon mock ([#1732](https://github.com/traefik/traefik/pull/1732) by [ldez](https://github.com/ldez))
- **[marathon]** Use single API call to fetch Marathon resources. ([#1815](https://github.com/traefik/traefik/pull/1815) by [timoreimann](https://github.com/timoreimann))
- **[metrics]** Added RetryMetrics to Datadog and StatsD providers ([#1884](https://github.com/traefik/traefik/pull/1884) by [aantono](https://github.com/aantono))
- **[metrics]** Extract metrics to own package and refactor implementations ([#1968](https://github.com/traefik/traefik/pull/1968) by [marco-jantke](https://github.com/marco-jantke))
- **[metrics]** Add metrics for backend_retries_total ([#1504](https://github.com/traefik/traefik/pull/1504) by [marco-jantke](https://github.com/marco-jantke))
- **[metrics]** Add status code to request duration metric ([#1755](https://github.com/traefik/traefik/pull/1755) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** Add trusted whitelist proxy protocol ([#2234](https://github.com/traefik/traefik/pull/2234) by [emilevauge](https://github.com/emilevauge)))
- **[metrics]** Datadog and StatsD Metrics Support ([#1701](https://github.com/traefik/traefik/pull/1701) by [aantono](https://github.com/aantono))
- **[middleware]** Create Header Middleware ([#1236](https://github.com/traefik/traefik/pull/1236) by [dtomcej](https://github.com/dtomcej))
- **[middleware]** Add configurable timeouts and curate default timeout settings ([#1873](https://github.com/traefik/traefik/pull/1873) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** Fix command bug content. ([#2002](https://github.com/traefik/traefik/pull/2002) by [ldez](https://github.com/ldez))
- **[middleware]** Retry only on real network errors ([#1549](https://github.com/traefik/traefik/pull/1549) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** Return 503 on empty backend ([#1748](https://github.com/traefik/traefik/pull/1748) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** Custom Error Pages ([#1675](https://github.com/traefik/traefik/pull/1675) by [bparli](https://github.com/bparli))
- **[oxy]** Support X-Forwarded-Port. ([#1960](https://github.com/traefik/traefik/pull/1960) by [ldez](https://github.com/ldez))
- **[provider,tls]** Added a check to ensure clientTLS configuration contains either a cert or a key ([#1932](https://github.com/traefik/traefik/pull/1932) by [aantono](https://github.com/aantono))
- **[provider]** Deflake integration tests ([#1599](https://github.com/traefik/traefik/pull/1599) by [ldez](https://github.com/ldez))
- **[provider]** Factorize labels ([#1843](https://github.com/traefik/traefik/pull/1843) by [ldez](https://github.com/ldez))
- **[provider]** Replace go routine by Safe.Go ([#1879](https://github.com/traefik/traefik/pull/1879) by [ldez](https://github.com/ldez))
- **[rancher]** Refactor into dual Rancher API/Metadata providers ([#1563](https://github.com/traefik/traefik/pull/1563) by [martinbaillie](https://github.com/martinbaillie))
- **[rules]** Add support for Query String filtering ([#1934](https://github.com/traefik/traefik/pull/1934) by [driverpt](https://github.com/driverpt))
- **[rules]** Simplify stripPrefix and stripPrefixRegex tests ([#1699](https://github.com/traefik/traefik/pull/1699) by [ldez](https://github.com/ldez))
- **[rules]** Enhance rules tests. ([#1679](https://github.com/traefik/traefik/pull/1679) by [ldez](https://github.com/ldez))
- **[sticky-session]** make the cookie name unique to the backend being served ([#1716](https://github.com/traefik/traefik/pull/1716) by [richardjq](https://github.com/richardjq))
- **[tls]** Handle RootCAs certificate ([#1789](https://github.com/traefik/traefik/pull/1789) by [Juliens](https://github.com/Juliens))
- **[tls]** enable TLS client forwarding ([#1446](https://github.com/traefik/traefik/pull/1446) by [drewwells](https://github.com/drewwells))
- **[websocket]** Add tests for urlencoded part in url ([#2199](https://github.com/traefik/traefik/pull/2199) by [Juliens](https://github.com/Juliens))
- **[websocket]** Add test for SSL TERMINATION in Websocket IT ([#2063](https://github.com/traefik/traefik/pull/2063) by [Juliens](https://github.com/Juliens)
- **[webui]** Proxy in dev mode ([#1544](https://github.com/traefik/traefik/pull/1544) by [maxwo](https://github.com/maxwo))
- **[webui]** Minor Health UI fixes ([#1651](https://github.com/traefik/traefik/pull/1651) by [mihaitodor](https://github.com/mihaitodor))
- Fail fast in IT and fix some flaky tests ([#2126](https://github.com/traefik/traefik/pull/2126) by [ldez](https://github.com/ldez))
- extract lb configuration steps into method ([#1841](https://github.com/traefik/traefik/pull/1841) by [marco-jantke](https://github.com/marco-jantke))
- Add whitelist configuration option for entrypoints ([#1702](https://github.com/traefik/traefik/pull/1702) by [christopherobin](https://github.com/christopherobin))
- Enhance integration tests ([#1842](https://github.com/traefik/traefik/pull/1842) by [ldez](https://github.com/ldez))
- Add helloworld tests with gRPC ([#1845](https://github.com/traefik/traefik/pull/1845) by [Juliens](https://github.com/Juliens))
- Add the sprig functions in the template engine ([#1891](https://github.com/traefik/traefik/pull/1891) by [thomasbach76](https://github.com/thomasbach76))
- Refactor globalConfiguration / WebProvider ([#1938](https://github.com/traefik/traefik/pull/1938) by [Juliens](https://github.com/Juliens))
- Code cleaning. ([#1956](https://github.com/traefik/traefik/pull/1956) by [ldez](https://github.com/ldez))
- Add proxy protocol ([#2004](https://github.com/traefik/traefik/pull/2004) by [emilevauge](https://github.com/emilevauge))
- Bump gorilla/mux version. ([#1954](https://github.com/traefik/traefik/pull/1954) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[cluster,kv]** Be certain to clear our marshalled representation before reloading it ([#2165](https://github.com/traefik/traefik/pull/2165) by [gozer](https://github.com/gozer))
- **[consulcatalog,docker,ecs,k8s,kv,marathon,rancher,sticky-session]** Backward compatibility for sticky ([#2266](https://github.com/traefik/traefik/pull/2266) by [ldez](https://github.com/ldez))
- **[consulcatalog,docker,ecs,k8s,marathon,rancher,sticky-session]** Stickiness cookie name ([#2232](https://github.com/traefik/traefik/pull/2232) by [ldez](https://github.com/ldez))
- **[consulcatalog,docker,ecs,k8s,marathon,rancher,sticky-session]** Stickiness cookie name. ([#2251](https://github.com/traefik/traefik/pull/2251) by [ldez](https://github.com/ldez))
- **[consulcatalog]** Fix consul catalog retry ([#2263](https://github.com/traefik/traefik/pull/2263) by [mmatur](https://github.com/mmatur))
- **[consulcatalog]** Flaky tests and refresh problem in consul catalog ([#2148](https://github.com/traefik/traefik/pull/2148) by [Juliens](https://github.com/Juliens))
- **[consulcatalog]** Consul catalog failed to remove service  ([#2157](https://github.com/traefik/traefik/pull/2157) by [Juliens](https://github.com/Juliens))
- **[consulcatalog]** Fix Consul Catalog refresh ([#2089](https://github.com/traefik/traefik/pull/2089) by [Juliens](https://github.com/Juliens))
- **[docker]** Changed Docker network filter to allow any swarm network ([#2244](https://github.com/traefik/traefik/pull/2244) by [pistolero](https://github.com/pistolero))
- **[docker]** Error handling for docker swarm mode ([#1533](https://github.com/traefik/traefik/pull/1533) by [tanyadegurechaff](https://github.com/tanyadegurechaff))
- **[ecs]** Handle empty ECS Clusters properly ([#2170](https://github.com/traefik/traefik/pull/2170) by [jeffreykoetsier](https://github.com/jeffreykoetsier))
- **[healthcheck]** Fix healthcheck port ([#2131](https://github.com/traefik/traefik/pull/2131) by [fredix](https://github.com/fredix))
- **[healthcheck]** Bind healthcheck to backend by entryPointName ([#1868](https://github.com/traefik/traefik/pull/1868) by [chrigl](https://github.com/chrigl))
- **[k8s]** Continue processing on invalid auth-realm annotation. ([#2252](https://github.com/traefik/traefik/pull/2252) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Use default frontend priority of zero. ([#1906](https://github.com/traefik/traefik/pull/1906) by [timoreimann](https://github.com/timoreimann))
- **[kv]** add retry backoff to staert config loading ([#2268](https://github.com/traefik/traefik/pull/2268) by [emilevauge](https://github.com/emilevauge))
- **[logs,middleware]** Enable loss less rotation of log files ([#2062](https://github.com/traefik/traefik/pull/2062) by [marco-jantke](https://github.com/marco-jantke))
- **[logs,middleware]** Access log default values ([#2061](https://github.com/traefik/traefik/pull/2061) by [ldez](https://github.com/ldez))
- **[logs]** Fix flakiness in log rotation test ([#2213](https://github.com/traefik/traefik/pull/2213) by [marco-jantke](https://github.com/marco-jantke))
- **[marathon]** Assign filtered tasks to apps contained in slice. ([#1881](https://github.com/traefik/traefik/pull/1881) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Fix fallback to other nodes for Marathon ([#1740](https://github.com/traefik/traefik/pull/1740) by [marco-jantke](https://github.com/marco-jantke))
- **[metrics]** prometheus, HTTP method and utf8 ([#2081](https://github.com/traefik/traefik/pull/2081) by [ldez](https://github.com/ldez))
- **[middleware]** Enable prefix matching within slash boundaries ([#2214](https://github.com/traefik/traefik/pull/2214) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** Fix SSE subscriptions when retries are enabled ([#2145](https://github.com/traefik/traefik/pull/2145) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** compress: preserve status code ([#1948](https://github.com/traefik/traefik/pull/1948) by [ldez](https://github.com/ldez))
- **[rancher]** Add stack name to backend name generation to fix rancher metadata backend ([#2107](https://github.com/traefik/traefik/pull/2107) by [SantoDE](https://github.com/SantoDE))
- **[rancher]** Rancher host IP address ([#2101](https://github.com/traefik/traefik/pull/2101) by [matq007](https://github.com/matq007))
- **[rancher]** fix seconds to really be seconds ([#2259](https://github.com/traefik/traefik/pull/2259) by [SantoDE](https://github.com/SantoDE))
- **[rancher]** fix rancher api environment get ([#2053](https://github.com/traefik/traefik/pull/2053) by [SantoDE](https://github.com/SantoDE))
- **[sticky-session]** Sanitize cookie names. ([#2216](https://github.com/traefik/traefik/pull/2216) by [timoreimann](https://github.com/timoreimann))
- **[sticky-session]** Setting the Cookie Path explicitly to root ([#1950](https://github.com/traefik/traefik/pull/1950) by [marcopaga](https://github.com/marcopaga))
- **[websocket]** Forward upgrade error from backend ([#2187](https://github.com/traefik/traefik/pull/2187) by [Juliens](https://github.com/Juliens))
- **[websocket]** RawPath and Transfer TLSConfig in websocket ([#2088](https://github.com/traefik/traefik/pull/2088) by [Juliens](https://github.com/Juliens))
- Nil body retries ([#2258](https://github.com/traefik/traefik/pull/2258) by [Juliens](https://github.com/Juliens))
- Fix deprecated IdleTimeout config ([#2143](https://github.com/traefik/traefik/pull/2143) by [marco-jantke](https://github.com/marco-jantke))
- Fixes entry points configuration. ([#2120](https://github.com/traefik/traefik/pull/2120) by [ldez](https://github.com/ldez))
- Delay first version check ([#2215](https://github.com/traefik/traefik/pull/2215) by [emilevauge](https://github.com/emilevauge))
- Move http2 configure transport  ([#2231](https://github.com/traefik/traefik/pull/2231) by [Juliens](https://github.com/Juliens))
- Fix error in prepareServer ([#2076](https://github.com/traefik/traefik/pull/2076) by [emilevauge](https://github.com/emilevauge))
- New entry point parser. ([#2248](https://github.com/traefik/traefik/pull/2248) by [ldez](https://github.com/ldez))
- Add TrustForwardHeader options. ([#2262](https://github.com/traefik/traefik/pull/2262) by [ldez](https://github.com/ldez))
- `bug` command. ([#2178](https://github.com/traefik/traefik/pull/2178) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme,provider]** Enhance documentation readability. ([#2095](https://github.com/traefik/traefik/pull/2095) by [ldez](https://github.com/ldez))
- **[acme,provider]** Fix whitespaces ([#2075](https://github.com/traefik/traefik/pull/2075) by [chulkilee](https://github.com/chulkilee))
- **[acme,provider]** Re-organize documentation ([#2012](https://github.com/traefik/traefik/pull/2012) by [jmaitrehenry](https://github.com/jmaitrehenry))
- **[acme]** Fix grammar ([#2208](https://github.com/traefik/traefik/pull/2208) by [mvasin](https://github.com/mvasin))
- **[acme]** Add guide for Docker, Traefik &amp; Letsencrypt ([#1923](https://github.com/traefik/traefik/pull/1923) by [mvdstam](https://github.com/mvdstam))
- **[acme]** Improve Let&#39;s Encrypt documentation ([#1885](https://github.com/traefik/traefik/pull/1885) by [nmengin](https://github.com/nmengin))
- **[acme]** Update docs for dnsimple env vars. ([#1872](https://github.com/traefik/traefik/pull/1872) by [untalpierre](https://github.com/untalpierre))
- **[api]** Add examples of proxying ping ([#2102](https://github.com/traefik/traefik/pull/2102) by [deitch](https://github.com/deitch))
- **[authentication,k8s]** traefik controller access to secrets ([#1707](https://github.com/traefik/traefik/pull/1707) by [spinto](https://github.com/spinto))
- **[consul,tls]** doc change regarding consul SSL ([#1774](https://github.com/traefik/traefik/pull/1774) by [bitsofinfo](https://github.com/bitsofinfo))
- **[consulcatalog,docker,ecs,k8s,marathon,rancher,sticky-session]** Stickiness documentation ([#2238](https://github.com/traefik/traefik/pull/2238) by [ldez](https://github.com/ldez))
- **[consul]** added consul acl token note ([#1720](https://github.com/traefik/traefik/pull/1720) by [bitsofinfo](https://github.com/bitsofinfo))
- **[docker]** Updating Docker output and curl for sticky sessions ([#2150](https://github.com/traefik/traefik/pull/2150) by [jtyr](https://github.com/jtyr))
- **[docker]** Add more visibility to docker stack deploy label issue ([#1984](https://github.com/traefik/traefik/pull/1984) by [jmaitrehenry](https://github.com/jmaitrehenry))
- **[ecs]** Fix IAM policy sid. ([#2066](https://github.com/traefik/traefik/pull/2066) by [charlieoleary](https://github.com/charlieoleary))
- **[k8s,marathon]** Mark Marathon and Kubernetes as constraint-supporting. ([#1964](https://github.com/traefik/traefik/pull/1964) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Add guide section on production advice, esp. CPU. ([#2113](https://github.com/traefik/traefik/pull/2113) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Document ways to partition Ingresses in the k8s guide. ([#2223](https://github.com/traefik/traefik/pull/2223) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Remove pod from RBAC rules. ([#2229](https://github.com/traefik/traefik/pull/2229) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Quote priority values in annotation examples. ([#2230](https://github.com/traefik/traefik/pull/2230) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Fix invalid service yaml example ([#2059](https://github.com/traefik/traefik/pull/2059) by [kairen](https://github.com/kairen))
- **[k8s]** Update usage of `.local` with `.minikube` in k8s docs ([#1551](https://github.com/traefik/traefik/pull/1551) by [errm](https://github.com/errm))
- **[k8s]** Update the documentation to use DaemonSet or Deployment ([#1735](https://github.com/traefik/traefik/pull/1735) by [saschagrunert](https://github.com/saschagrunert))
- **[k8s]** Fix docs about default namespaces. ([#1961](https://github.com/traefik/traefik/pull/1961) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Moved namespace to correct place ([#1911](https://github.com/traefik/traefik/pull/1911) by [markround](https://github.com/markround))
- **[k8s]** examples/k8s: fix ui ingress port out of sync with deployment ([#1943](https://github.com/traefik/traefik/pull/1943) by [borancar](https://github.com/borancar))
- **[k8s]** Add secrets resource to in-line RBAC spec. ([#1890](https://github.com/traefik/traefik/pull/1890) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Improve documentation. ([#1831](https://github.com/traefik/traefik/pull/1831) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Fix documentation glitches. ([#1996](https://github.com/traefik/traefik/pull/1996) by [timoreimann](https://github.com/timoreimann))
- **[metrics]** Enhance web backend documentation ([#2122](https://github.com/traefik/traefik/pull/2122) by [ldez](https://github.com/ldez))
- **[mesos]** fix: documentation Mesos. ([#2029](https://github.com/traefik/traefik/pull/2029) by [ldez](https://github.com/ldez))
- **[middleware]** Improve compression documentation ([#2184](https://github.com/traefik/traefik/pull/2184) by [errm](https://github.com/errm))
- **[provider]** Clarify that provider-enabling argument parameters set all defaults. ([#1830](https://github.com/traefik/traefik/pull/1830) by [timoreimann](https://github.com/timoreimann))
- **[rancher]** Update Rancher documentation. ([#1776](https://github.com/traefik/traefik/pull/1776) by [ldez](https://github.com/ldez))
- **[webui]** Document yarnpkg. ([#1558](https://github.com/traefik/traefik/pull/1558) by [Stibbons](https://github.com/Stibbons))
- Add forward auth documentation. ([#2110](https://github.com/traefik/traefik/pull/2110) by [ldez](https://github.com/ldez))
- User guide gRPC ([#2108](https://github.com/traefik/traefik/pull/2108) by [Juliens](https://github.com/Juliens))
- Document custom error page restrictions. ([#2104](https://github.com/traefik/traefik/pull/2104) by [timoreimann](https://github.com/timoreimann))
- Prepare release v1.4.0-rc3 ([#2135](https://github.com/traefik/traefik/pull/2135) by [Juliens](https://github.com/Juliens))
- Update gRPC example ([#2191](https://github.com/traefik/traefik/pull/2191) by [jsenon](https://github.com/jsenon))
- Prepare release v1.4.0-rc2 ([#2091](https://github.com/traefik/traefik/pull/2091) by [ldez](https://github.com/ldez))
- Fix grammar mistake in the kv-config docs ([#2197](https://github.com/traefik/traefik/pull/2197) by [chr4](https://github.com/chr4))
- Update cluster.md ([#2073](https://github.com/traefik/traefik/pull/2073) by [kmbremner](https://github.com/kmbremner))
- Prepare release v1.4.0-rc4 ([#2201](https://github.com/traefik/traefik/pull/2201) by [nmengin](https://github.com/nmengin))
- Prepare release v1.4.0-rc5 ([#2241](https://github.com/traefik/traefik/pull/2241) by [ldez](https://github.com/ldez))
- Enhance documentation. ([#2048](https://github.com/traefik/traefik/pull/2048) by [ldez](https://github.com/ldez))
- doc: add notes on server urls with path ([#2045](https://github.com/traefik/traefik/pull/2045) by [chulkilee](https://github.com/chulkilee))
- Enhance security headers doc. ([#2042](https://github.com/traefik/traefik/pull/2042) by [ldez](https://github.com/ldez))
- HTTPS for images, video and links in docs. ([#2041](https://github.com/traefik/traefik/pull/2041) by [ldez](https://github.com/ldez))
- Fix error pages configuration. ([#2038](https://github.com/traefik/traefik/pull/2038) by [ldez](https://github.com/ldez))
- Fix Proxy Protocol documentation ([#2253](https://github.com/traefik/traefik/pull/2253) by [emilevauge](https://github.com/emilevauge))
- Update GraceTimeOut documentation ([#1875](https://github.com/traefik/traefik/pull/1875) by [marco-jantke](https://github.com/marco-jantke))
- Release cycle. ([#1812](https://github.com/traefik/traefik/pull/1812) by [ldez](https://github.com/ldez))
- Update contributing guide build steps ([#1801](https://github.com/traefik/traefik/pull/1801) by [jsturtevant](https://github.com/jsturtevant))
- Add Nicolas Mengin to maintainers ([#1792](https://github.com/traefik/traefik/pull/1792) by [emilevauge](https://github.com/emilevauge))
- Add Julien Salleyron to maintainers ([#1790](https://github.com/traefik/traefik/pull/1790) by [emilevauge](https://github.com/emilevauge))
- Change to a more flexible PR review process ([#1781](https://github.com/traefik/traefik/pull/1781) by [emilevauge](https://github.com/emilevauge))
- Traefik &#34;bug&#34; command documentation ([#1811](https://github.com/traefik/traefik/pull/1811) by [ldez](https://github.com/ldez))
- Change Traefik intro video ([#1893](https://github.com/traefik/traefik/pull/1893) by [emilevauge](https://github.com/emilevauge))
- Prepare release v1.4.0-rc1 ([#2021](https://github.com/traefik/traefik/pull/2021) by [ldez](https://github.com/ldez))
- Add play-with-docker example ([#1726](https://github.com/traefik/traefik/pull/1726) by [marcosnils](https://github.com/marcosnils))
- Add Marco Jantke to maintainers ([#1980](https://github.com/traefik/traefik/pull/1980) by [emilevauge](https://github.com/emilevauge))
- Remove Russel from maintainers ([#1614](https://github.com/traefik/traefik/pull/1614) by [emilevauge](https://github.com/emilevauge))
- Update CONTRIBUTING.md. ([#1667](https://github.com/traefik/traefik/pull/1667) by [timoreimann](https://github.com/timoreimann))
- drop &#34;slave&#34; wording for &#34;worker&#34; ([#1645](https://github.com/traefik/traefik/pull/1645) by [djalal](https://github.com/djalal))
- Use more inclusive language in README.md {guys =&gt; folks} ([#1640](https://github.com/traefik/traefik/pull/1640) by [igorwwwwwwwwwwwwwwwwwwww](https://github.com/igorwwwwwwwwwwwwwwwwwwww))
- Remove Thomas Recloux from maintainers ([#1616](https://github.com/traefik/traefik/pull/1616) by [emilevauge](https://github.com/emilevauge))
- Update documentation for 1.4 release ([#2011](https://github.com/traefik/traefik/pull/2011) by [emilevauge](https://github.com/emilevauge))
- Small toml documentation update ([#1603](https://github.com/traefik/traefik/pull/1603) by [antoine-aumjaud](https://github.com/antoine-aumjaud))
- Add @ldez to maintainers ([#1589](https://github.com/traefik/traefik/pull/1589) by [emilevauge](https://github.com/emilevauge))
- doc: add labels documentation. ([#1582](https://github.com/traefik/traefik/pull/1582) by [ldez](https://github.com/ldez))
- Update golang version in contributing guide ([#2018](https://github.com/traefik/traefik/pull/2018) by [ArikaChen](https://github.com/ArikaChen))
- toml page - replace li by table ([#1995](https://github.com/traefik/traefik/pull/1995) by [jmaitrehenry](https://github.com/jmaitrehenry))

**Misc:**
- Merge v1.3.7 ([#2013](https://github.com/traefik/traefik/pull/2013) by [ldez](https://github.com/ldez))
- Merge 1.3.6 ([#1992](https://github.com/traefik/traefik/pull/1992) by [ldez](https://github.com/ldez))
- Merge 1.3.5 ([#1909](https://github.com/traefik/traefik/pull/1909) by [ldez](https://github.com/ldez))
- Merge 1.3.3 ([#1836](https://github.com/traefik/traefik/pull/1836) by [ldez](https://github.com/ldez))
- Merge v1.3.2 to master  ([#1809](https://github.com/traefik/traefik/pull/1809) by [ldez](https://github.com/ldez))
- Merge current v1.3 ([#1797](https://github.com/traefik/traefik/pull/1797) by [ldez](https://github.com/ldez))
- Merge current v1.3 ([#1786](https://github.com/traefik/traefik/pull/1786) by [ldez](https://github.com/ldez))
- Merge v1.3.1 to master ([#1763](https://github.com/traefik/traefik/pull/1763) by [ldez](https://github.com/ldez))
- Merge current v1.3 ([#1753](https://github.com/traefik/traefik/pull/1753) by [ldez](https://github.com/ldez))
- Merge current v1.3 ([#1705](https://github.com/traefik/traefik/pull/1705) by [ldez](https://github.com/ldez))
- Merge current v1.3 to master ([#1697](https://github.com/traefik/traefik/pull/1697) by [ldez](https://github.com/ldez))
- Merge v1 3 0 ([#1692](https://github.com/traefik/traefik/pull/1692) by [ldez](https://github.com/ldez))
- Merge current v1.3 to master (rc3) ([#1666](https://github.com/traefik/traefik/pull/1666) by [ldez](https://github.com/ldez))
- Merge current v1.3 to master  ([#1643](https://github.com/traefik/traefik/pull/1643) by [ldez](https://github.com/ldez))
- Merge v1.3.0-rc2 master ([#1613](https://github.com/traefik/traefik/pull/1613) by [emilevauge](https://github.com/emilevauge))
- Merge v1.3 branch into master [2017-05-11] ([#1548](https://github.com/traefik/traefik/pull/1548) by [timoreimann](https://github.com/timoreimann))

## [v1.4.0-rc5](https://github.com/traefik/traefik/tree/v1.4.0-rc5) (2017-10-10)
[All Commits](https://github.com/traefik/traefik/compare/v1.4.0-rc4...v1.4.0-rc5)

**Enhancements:**
- **[middleware]** Add trusted whitelist proxy protocol ([#2234](https://github.com/traefik/traefik/pull/2234) by [emilevauge](https://github.com/emilevauge))

**Bug fixes:**
- **[consul,docker,ecs,k8s,marathon,rancher,sticky-session]** Stickiness cookie name ([#2232](https://github.com/traefik/traefik/pull/2232) by [ldez](https://github.com/ldez))
- **[logs]** Fix flakiness in log rotation test ([#2213](https://github.com/traefik/traefik/pull/2213) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** Enable prefix matching within slash boundaries ([#2214](https://github.com/traefik/traefik/pull/2214) by [marco-jantke](https://github.com/marco-jantke))
- **[sticky-session]** Sanitize cookie names. ([#2216](https://github.com/traefik/traefik/pull/2216) by [timoreimann](https://github.com/timoreimann))
- Move http2 configure transport  ([#2231](https://github.com/traefik/traefik/pull/2231) by [Juliens](https://github.com/Juliens))
- Delay first version check ([#2215](https://github.com/traefik/traefik/pull/2215) by [emilevauge](https://github.com/emilevauge))

**Documentation:**
- **[acme]** Fix grammar ([#2208](https://github.com/traefik/traefik/pull/2208) by [mvasin](https://github.com/mvasin))
- **[docker,ecs,k8s,marathon,rancher]** Stickiness documentation ([#2238](https://github.com/traefik/traefik/pull/2238) by [ldez](https://github.com/ldez))
- **[k8s]** Quote priority values in annotation examples. ([#2230](https://github.com/traefik/traefik/pull/2230) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Remove pod from RBAC rules. ([#2229](https://github.com/traefik/traefik/pull/2229) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Document ways to partition Ingresses in the k8s guide. ([#2223](https://github.com/traefik/traefik/pull/2223) by [timoreimann](https://github.com/timoreimann))

## [v1.4.0-rc4](https://github.com/traefik/traefik/tree/v1.4.0-rc4) (2017-10-02)
[All Commits](https://github.com/traefik/traefik/compare/v1.4.0-rc3...v1.4.0-rc4)

**Bug fixes:**
- **[cluster,kv]** Be certain to clear our marshalled representation before reloading it ([#2165](https://github.com/traefik/traefik/pull/2165) by [gozer](https://github.com/gozer))
- **[consulcatalog]** Consul catalog failed to remove service  ([#2157](https://github.com/traefik/traefik/pull/2157) by [Juliens](https://github.com/Juliens))
- **[consulcatalog]** Flaky tests and refresh problem in consul catalog ([#2148](https://github.com/traefik/traefik/pull/2148) by [Juliens](https://github.com/Juliens))
- **[ecs]** Handle empty ECS Clusters properly ([#2170](https://github.com/traefik/traefik/pull/2170) by [jeffreykoetsier](https://github.com/jeffreykoetsier))
- **[middleware]** Fix SSE subscriptions when retries are enabled ([#2145](https://github.com/traefik/traefik/pull/2145) by [marco-jantke](https://github.com/marco-jantke))
- **[websocket]** Forward upgrade error from backend ([#2187](https://github.com/traefik/traefik/pull/2187) by [Juliens](https://github.com/Juliens))
- `bug` command. ([#2178](https://github.com/traefik/traefik/pull/2178) by [ldez](https://github.com/ldez))
- Fix deprecated IdleTimeout config ([#2143](https://github.com/traefik/traefik/pull/2143) by [marco-jantke](https://github.com/marco-jantke))

**Documentation:**
- **[docker]** Updating Docker output and curl for sticky sessions ([#2150](https://github.com/traefik/traefik/pull/2150) by [jtyr](https://github.com/jtyr))
- **[middleware]** Improve compression documentation ([#2184](https://github.com/traefik/traefik/pull/2184) by [errm](https://github.com/errm))
- Fix grammar mistake in the kv-config docs ([#2197](https://github.com/traefik/traefik/pull/2197) by [chr4](https://github.com/chr4))
- Update gRPC example ([#2191](https://github.com/traefik/traefik/pull/2191) by [jsenon](https://github.com/jsenon))

**Misc:**
- **[websocket]** Add tests for urlencoded part in url ([#2199](https://github.com/traefik/traefik/pull/2199) by [Juliens](https://github.com/Juliens))

## [v1.4.0-rc3](https://github.com/traefik/traefik/tree/v1.4.0-rc3) (2017-09-18)
[All Commits](https://github.com/traefik/traefik/compare/v1.4.0-rc2...v1.4.0-rc3)

**Enhancements:**
- **[acme]** Display Traefik logs in integration tests ([#2114](https://github.com/traefik/traefik/pull/2114) by [ldez](https://github.com/ldez))
- **[authentication]** Manage Headers for the Authentication forwarding. ([#2132](https://github.com/traefik/traefik/pull/2132) by [ldez](https://github.com/ldez))
- Fail fast in IT and fix some flaky tests ([#2126](https://github.com/traefik/traefik/pull/2126) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[consul]** Fix Consul Catalog refresh ([#2089](https://github.com/traefik/traefik/pull/2089) by [Juliens](https://github.com/Juliens))
- **[healthcheck]** Fix healthcheck port ([#2131](https://github.com/traefik/traefik/pull/2131) by [fredix](https://github.com/fredix))
- **[logs,middleware]** Enable loss less rotation of log files ([#2062](https://github.com/traefik/traefik/pull/2062) by [marco-jantke](https://github.com/marco-jantke))
- **[rancher]** Add stack name to backend name generation to fix rancher metadata backend ([#2107](https://github.com/traefik/traefik/pull/2107) by [SantoDE](https://github.com/SantoDE))
- **[rancher]** Rancher host IP address ([#2101](https://github.com/traefik/traefik/pull/2101) by [matq007](https://github.com/matq007))
- Fixes entry points configuration. ([#2120](https://github.com/traefik/traefik/pull/2120) by [ldez](https://github.com/ldez))

**Documentation:**
- **[acme,provider]** Enhance documentation readability. ([#2095](https://github.com/traefik/traefik/pull/2095) by [ldez](https://github.com/ldez))
- **[api]** Add examples of proxying ping ([#2102](https://github.com/traefik/traefik/pull/2102) by [deitch](https://github.com/deitch))
- **[k8s]** Add guide section on production advice, esp. CPU. ([#2113](https://github.com/traefik/traefik/pull/2113) by [timoreimann](https://github.com/timoreimann))
- **[metrics]** Enhance web backend documentation ([#2122](https://github.com/traefik/traefik/pull/2122) by [ldez](https://github.com/ldez))
- Add forward auth documentation. ([#2110](https://github.com/traefik/traefik/pull/2110) by [ldez](https://github.com/ldez))
- User guide gRPC ([#2108](https://github.com/traefik/traefik/pull/2108) by [Juliens](https://github.com/Juliens))
- Document custom error page restrictions. ([#2104](https://github.com/traefik/traefik/pull/2104) by [timoreimann](https://github.com/timoreimann))

## [v1.4.0-rc2](https://github.com/traefik/traefik/tree/v1.4.0-rc2) (2017-09-08)
[All Commits](https://github.com/traefik/traefik/compare/v1.4.0-rc1...v1.4.0-rc2)

**Enhancements:**
- **[authentication,consul]** Add Basic auth for consul catalog ([#2027](https://github.com/traefik/traefik/pull/2027) by [mmatur](https://github.com/mmatur))
- **[authentication,ecs]** Add basic auth for ecs ([#2026](https://github.com/traefik/traefik/pull/2026) by [mmatur](https://github.com/mmatur))
- **[logs]** Send traefik logs to stdout instead stderr ([#2054](https://github.com/traefik/traefik/pull/2054) by [marco-jantke](https://github.com/marco-jantke))
- **[websocket]** Add test for SSL TERMINATION in Websocket IT ([#2063](https://github.com/traefik/traefik/pull/2063) by [Juliens](https://github.com/Juliens))

**Bug fixes:**
- **[consul]** Fix consul catalog refresh problems ([#2089](https://github.com/traefik/traefik/pull/2089) by [Juliens](https://github.com/Juliens))
- **[logs,middleware]** Access log default values ([#2061](https://github.com/traefik/traefik/pull/2061) by [ldez](https://github.com/ldez))
- **[metrics]** prometheus, HTTP method and utf8 ([#2081](https://github.com/traefik/traefik/pull/2081) by [ldez](https://github.com/ldez))
- **[rancher]** fix rancher api environment get ([#2053](https://github.com/traefik/traefik/pull/2053) by [SantoDE](https://github.com/SantoDE))
- **[websocket]** RawPath and Transfer TLSConfig in websocket ([#2088](https://github.com/traefik/traefik/pull/2088) by [Juliens](https://github.com/Juliens))
- Fix error in prepareServer ([#2076](https://github.com/traefik/traefik/pull/2076) by [emilevauge](https://github.com/emilevauge))

**Documentation:**
- **[acme,provider]** Fix whitespaces ([#2075](https://github.com/traefik/traefik/pull/2075) by [chulkilee](https://github.com/chulkilee))
- **[ecs]** Fix IAM policy sid. ([#2066](https://github.com/traefik/traefik/pull/2066) by [charlieoleary](https://github.com/charlieoleary))
- **[k8s]** Fix invalid service yaml example ([#2059](https://github.com/traefik/traefik/pull/2059) by [kairen](https://github.com/kairen))
- **[mesos]** fix: documentation Mesos. ([#2029](https://github.com/traefik/traefik/pull/2029) by [ldez](https://github.com/ldez))
- Update cluster.md ([#2073](https://github.com/traefik/traefik/pull/2073) by [kmbremner](https://github.com/kmbremner))
- Enhance documentation. ([#2048](https://github.com/traefik/traefik/pull/2048) by [ldez](https://github.com/ldez))
- doc: add notes on server urls with path ([#2045](https://github.com/traefik/traefik/pull/2045) by [chulkilee](https://github.com/chulkilee))
- Enhance security headers doc. ([#2042](https://github.com/traefik/traefik/pull/2042) by [ldez](https://github.com/ldez))
- HTTPS for images, video and links in docs. ([#2041](https://github.com/traefik/traefik/pull/2041) by [ldez](https://github.com/ldez))
- Fix error pages configuration. ([#2038](https://github.com/traefik/traefik/pull/2038) by [ldez](https://github.com/ldez))

## [v1.4.0-rc1](https://github.com/traefik/traefik/tree/v1.4.0-rc1) (2017-08-28)
[All Commits](https://github.com/traefik/traefik/compare/v1.3.0-rc1...v1.4.0-rc1)

**Enhancements:**
- **[acme]** Make the ACME developments testing easier ([#1769](https://github.com/traefik/traefik/pull/1769) by [nmengin](https://github.com/nmengin))
- **[acme]** contrib: Dump keys/certs from acme.json to files ([#1484](https://github.com/traefik/traefik/pull/1484) by [brianredbeard](https://github.com/brianredbeard))
- **[api]** Add HTTP HEAD handling to /ping endpoint ([#1768](https://github.com/traefik/traefik/pull/1768) by [martinbaillie](https://github.com/martinbaillie))
- **[authentication,marathon]** Add marathon label to configure basic auth ([#1799](https://github.com/traefik/traefik/pull/1799) by [nikore](https://github.com/nikore))
- **[authentication,middleware]** Add forward authentication option ([#1972](https://github.com/traefik/traefik/pull/1972) by [drampelt](https://github.com/drampelt))
- **[consul,sticky-session]** Enable loadbalancer.sticky for Consul Catalog ([#1917](https://github.com/traefik/traefik/pull/1917) by [nbonneval](https://github.com/nbonneval))
- **[consul]** Enhanced flexibility in Consul Catalog configuration ([#1565](https://github.com/traefik/traefik/pull/1565) by [aantono](https://github.com/aantono))
- **[consul]** Exposed by default feature in Consul Catalog ([#2006](https://github.com/traefik/traefik/pull/2006) by [mmatur](https://github.com/mmatur))
- **[consul]** Speeding up consul catalog health change detection ([#1694](https://github.com/traefik/traefik/pull/1694) by [vholovko](https://github.com/vholovko))
- **[docker,k8s]** IP Whitelists for Frontend (with Docker- &amp; Kubernetes-Provider Support) ([#1332](https://github.com/traefik/traefik/pull/1332) by [MaZderMind](https://github.com/MaZderMind))
- **[ecs,sticky-session]** Enable loadbalancer.sticky for ECS ([#1925](https://github.com/traefik/traefik/pull/1925) by [mmatur](https://github.com/mmatur))
- **[ecs]** Add support for several ECS backends ([#1913](https://github.com/traefik/traefik/pull/1913) by [mmatur](https://github.com/mmatur))
- **[healthcheck]** Add healthcheck command ([#1982](https://github.com/traefik/traefik/pull/1982) by [emilevauge](https://github.com/emilevauge))
- **[healthcheck]** Allow overriding the port used for healthchecks ([#1567](https://github.com/traefik/traefik/pull/1567) by [bakins](https://github.com/bakins))
- **[k8s,rules]** kubernetes ingress rewrite-target implementation ([#1723](https://github.com/traefik/traefik/pull/1723) by [mlaccetti](https://github.com/mlaccetti))
- **[k8s]** Added ability to override frontend priority for k8s ingress router ([#1874](https://github.com/traefik/traefik/pull/1874) by [DiverOfDark](https://github.com/DiverOfDark))
- **[kv]** Adds definitions to backend kv template for health checking ([#1644](https://github.com/traefik/traefik/pull/1644) by [zachomedia](https://github.com/zachomedia))
- **[logs,dynamodb,ecs,marathon]** Link some providers logs to Traefik ([#1746](https://github.com/traefik/traefik/pull/1746) by [ldez](https://github.com/ldez))
- **[logs,marathon]** remove confusing go-marathon log message ([#1810](https://github.com/traefik/traefik/pull/1810) by [marco-jantke](https://github.com/marco-jantke))
- **[logs]** enable logging to stdout for access logs ([#1683](https://github.com/traefik/traefik/pull/1683) by [marco-jantke](https://github.com/marco-jantke))
- **[logs]** Logs &amp; errors review ([#1673](https://github.com/traefik/traefik/pull/1673) by [ldez](https://github.com/ldez))
- **[logs]** log X-Forwarded-For as ClientHost if present ([#1946](https://github.com/traefik/traefik/pull/1946) by [mildis](https://github.com/mildis))
- **[logs]** Switch access logging to logrus ([#1647](https://github.com/traefik/traefik/pull/1647) by [rjshep](https://github.com/rjshep))
- **[logs]** add RetryAttempts to AccessLog in JSON format ([#1793](https://github.com/traefik/traefik/pull/1793) by [marco-jantke](https://github.com/marco-jantke))
- **[logs]** Restore: First stage of access logging middleware. ([#1571](https://github.com/traefik/traefik/pull/1571) by [ldez](https://github.com/ldez))
- **[logs]** Add log file close and reopen on receipt of SIGUSR1 ([#1761](https://github.com/traefik/traefik/pull/1761) by [rjshep](https://github.com/rjshep))
- **[logs]** Add JSON as access logging format ([#1669](https://github.com/traefik/traefik/pull/1669) by [rjshep](https://github.com/rjshep))
- **[marathon]** Add support for readiness checks. ([#1883](https://github.com/traefik/traefik/pull/1883) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Exported getSubDomain function from Marathon provider ([#1693](https://github.com/traefik/traefik/pull/1693) by [aantono](https://github.com/aantono))
- **[marathon]** Improve Marathon integration tests. ([#1406](https://github.com/traefik/traefik/pull/1406) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Use single API call to fetch Marathon resources. ([#1815](https://github.com/traefik/traefik/pull/1815) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Move marathon mock ([#1732](https://github.com/traefik/traefik/pull/1732) by [ldez](https://github.com/ldez))
- **[marathon]** Support multi-port service routing for containers running on Marathon ([#1742](https://github.com/traefik/traefik/pull/1742) by [aantono](https://github.com/aantono))
- **[marathon]** Use test builder. ([#1871](https://github.com/traefik/traefik/pull/1871) by [timoreimann](https://github.com/timoreimann))
- **[metrics]** Datadog and StatsD Metrics Support ([#1701](https://github.com/traefik/traefik/pull/1701) by [aantono](https://github.com/aantono))
- **[metrics]** Add status code to request duration metric ([#1755](https://github.com/traefik/traefik/pull/1755) by [marco-jantke](https://github.com/marco-jantke))
- **[metrics]** Add metrics for backend_retries_total ([#1504](https://github.com/traefik/traefik/pull/1504) by [marco-jantke](https://github.com/marco-jantke))
- **[metrics]** Extract metrics to own package and refactor implementations ([#1968](https://github.com/traefik/traefik/pull/1968) by [marco-jantke](https://github.com/marco-jantke))
- **[metrics]** Added RetryMetrics to Datadog and StatsD providers ([#1884](https://github.com/traefik/traefik/pull/1884) by [aantono](https://github.com/aantono))
- **[middleware]** Return 503 on empty backend ([#1748](https://github.com/traefik/traefik/pull/1748) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** Add configurable timeouts and curate default timeout settings ([#1873](https://github.com/traefik/traefik/pull/1873) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** Custom Error Pages ([#1675](https://github.com/traefik/traefik/pull/1675) by [bparli](https://github.com/bparli))
- **[middleware]** Retry only on real network errors ([#1549](https://github.com/traefik/traefik/pull/1549) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** Fix command bug content. ([#2002](https://github.com/traefik/traefik/pull/2002) by [ldez](https://github.com/ldez))
- **[middleware]** Create Header Middleware ([#1236](https://github.com/traefik/traefik/pull/1236) by [dtomcej](https://github.com/dtomcej))
- **[oxy]** Support X-Forwarded-Port. ([#1960](https://github.com/traefik/traefik/pull/1960) by [ldez](https://github.com/ldez))
- **[provider,tls]** Added a check to ensure clientTLS configuration contains either a cert or a key ([#1932](https://github.com/traefik/traefik/pull/1932) by [aantono](https://github.com/aantono))
- **[provider]** Factorize labels ([#1843](https://github.com/traefik/traefik/pull/1843) by [ldez](https://github.com/ldez))
- **[provider]** Replace go routine by Safe.Go ([#1879](https://github.com/traefik/traefik/pull/1879) by [ldez](https://github.com/ldez))
- **[provider]** Deflake integration tests ([#1599](https://github.com/traefik/traefik/pull/1599) by [ldez](https://github.com/ldez))
- **[rancher]** Refactor into dual Rancher API/Metadata providers ([#1563](https://github.com/traefik/traefik/pull/1563) by [martinbaillie](https://github.com/martinbaillie))
- **[rules]** Simplify stripPrefix and stripPrefixRegex tests ([#1699](https://github.com/traefik/traefik/pull/1699) by [ldez](https://github.com/ldez))
- **[rules]** Add support for Query String filtering ([#1934](https://github.com/traefik/traefik/pull/1934) by [driverpt](https://github.com/driverpt))
- **[rules]** Enhance rules tests. ([#1679](https://github.com/traefik/traefik/pull/1679) by [ldez](https://github.com/ldez))
- **[sticky-session]** make the cookie name unique to the backend being served ([#1716](https://github.com/traefik/traefik/pull/1716) by [richardjq](https://github.com/richardjq))
- **[tls]** Handle RootCAs certificate ([#1789](https://github.com/traefik/traefik/pull/1789) by [Juliens](https://github.com/Juliens))
- **[tls]** enable TLS client forwarding ([#1446](https://github.com/traefik/traefik/pull/1446) by [drewwells](https://github.com/drewwells))
- **[webui]** Minor Health UI fixes ([#1651](https://github.com/traefik/traefik/pull/1651) by [mihaitodor](https://github.com/mihaitodor))
- **[webui]** Proxy in dev mode ([#1544](https://github.com/traefik/traefik/pull/1544) by [maxwo](https://github.com/maxwo))
- extract lb configuration steps into method ([#1841](https://github.com/traefik/traefik/pull/1841) by [marco-jantke](https://github.com/marco-jantke))
- Allow file provider to load service config from files in a directory. ([#1672](https://github.com/traefik/traefik/pull/1672) by [rjshep](https://github.com/rjshep))
- Add whitelist configuration option for entrypoints ([#1702](https://github.com/traefik/traefik/pull/1702) by [christopherobin](https://github.com/christopherobin))
- Enhance integration tests ([#1842](https://github.com/traefik/traefik/pull/1842) by [ldez](https://github.com/ldez))
- Add helloworld tests with gRPC ([#1845](https://github.com/traefik/traefik/pull/1845) by [Juliens](https://github.com/Juliens))
- Add the sprig functions in the template engine ([#1891](https://github.com/traefik/traefik/pull/1891) by [thomasbach76](https://github.com/thomasbach76))
- Refactor globalConfiguration / WebProvider ([#1938](https://github.com/traefik/traefik/pull/1938) by [Juliens](https://github.com/Juliens))
- Code cleaning. ([#1956](https://github.com/traefik/traefik/pull/1956) by [ldez](https://github.com/ldez))
- Add proxy protocol ([#2004](https://github.com/traefik/traefik/pull/2004) by [emilevauge](https://github.com/emilevauge))
- Bump gorilla/mux version. ([#1954](https://github.com/traefik/traefik/pull/1954) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[docker]** Error handling for docker swarm mode ([#1533](https://github.com/traefik/traefik/pull/1533) by [tanyadegurechaff](https://github.com/tanyadegurechaff))
- **[healthcheck]** Bind healthcheck to backend by entryPointName ([#1868](https://github.com/traefik/traefik/pull/1868) by [chrigl](https://github.com/chrigl))
- **[k8s]** Use default frontend priority of zero. ([#1906](https://github.com/traefik/traefik/pull/1906) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Assign filtered tasks to apps contained in slice. ([#1881](https://github.com/traefik/traefik/pull/1881) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Fix fallback to other nodes for Marathon ([#1740](https://github.com/traefik/traefik/pull/1740) by [marco-jantke](https://github.com/marco-jantke))
- **[middleware]** compress: preserve status code ([#1948](https://github.com/traefik/traefik/pull/1948) by [ldez](https://github.com/ldez))
- **[sticky-session]** Setting the Cookie Path explicitly to root ([#1950](https://github.com/traefik/traefik/pull/1950) by [marcopaga](https://github.com/marcopaga))

**Documentation:**
- **[acme,provider]** Re-organize documentation ([#2012](https://github.com/traefik/traefik/pull/2012) by [jmaitrehenry](https://github.com/jmaitrehenry))
- **[acme]** Add guide for Docker, Traefik &amp; Letsencrypt ([#1923](https://github.com/traefik/traefik/pull/1923) by [mvdstam](https://github.com/mvdstam))
- **[acme]** Update docs for dnsimple env vars. ([#1872](https://github.com/traefik/traefik/pull/1872) by [klud1](https://github.com/klud1))
- **[acme]** Improve Let&#39;s Encrypt documentation ([#1885](https://github.com/traefik/traefik/pull/1885) by [nmengin](https://github.com/nmengin))
- **[authentication,k8s]** traefik controller access to secrets ([#1707](https://github.com/traefik/traefik/pull/1707) by [spinto](https://github.com/spinto))
- **[consul,tls]** doc change regarding consul SSL ([#1774](https://github.com/traefik/traefik/pull/1774) by [bitsofinfo](https://github.com/bitsofinfo))
- **[consul]** added consul acl token note ([#1720](https://github.com/traefik/traefik/pull/1720) by [bitsofinfo](https://github.com/bitsofinfo))
- **[docker]** Add more visibility to docker stack deploy label issue ([#1984](https://github.com/traefik/traefik/pull/1984) by [jmaitrehenry](https://github.com/jmaitrehenry))
- **[k8s,marathon]** Mark Marathon and Kubernetes as constraint-supporting. ([#1964](https://github.com/traefik/traefik/pull/1964) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** examples/k8s: fix ui ingress port out of sync with deployment ([#1943](https://github.com/traefik/traefik/pull/1943) by [borancar](https://github.com/borancar))
- **[k8s]** Update the documentation to use DaemonSet or Deployment ([#1735](https://github.com/traefik/traefik/pull/1735) by [saschagrunert](https://github.com/saschagrunert))
- **[k8s]** Moved namespace to correct place ([#1911](https://github.com/traefik/traefik/pull/1911) by [markround](https://github.com/markround))
- **[k8s]** Improve documentation. ([#1831](https://github.com/traefik/traefik/pull/1831) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Add secrets resource to in-line RBAC spec. ([#1890](https://github.com/traefik/traefik/pull/1890) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Fix docs about default namespaces. ([#1961](https://github.com/traefik/traefik/pull/1961) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Update usage of `.local` with `.minikube` in k8s docs ([#1551](https://github.com/traefik/traefik/pull/1551) by [errm](https://github.com/errm))
- **[marathon]** Fix documentation glitches. ([#1996](https://github.com/traefik/traefik/pull/1996) by [timoreimann](https://github.com/timoreimann))
- **[provider]** Clarify that provider-enabling argument parameters set all defaults. ([#1830](https://github.com/traefik/traefik/pull/1830) by [timoreimann](https://github.com/timoreimann))
- **[rancher]** Update Rancher documentation. ([#1776](https://github.com/traefik/traefik/pull/1776) by [ldez](https://github.com/ldez))
- **[webui]** Document yarnpkg. ([#1558](https://github.com/traefik/traefik/pull/1558) by [Stibbons](https://github.com/Stibbons))
- Add play-with-docker example ([#1726](https://github.com/traefik/traefik/pull/1726) by [marcosnils](https://github.com/marcosnils))
- Update contributing guide build steps ([#1801](https://github.com/traefik/traefik/pull/1801) by [jsturtevant](https://github.com/jsturtevant))
- Add Nicolas Mengin to maintainers ([#1792](https://github.com/traefik/traefik/pull/1792) by [emilevauge](https://github.com/emilevauge))
- Add Julien Salleyron to maintainers ([#1790](https://github.com/traefik/traefik/pull/1790) by [emilevauge](https://github.com/emilevauge))
- Change to a more flexible PR review process ([#1781](https://github.com/traefik/traefik/pull/1781) by [emilevauge](https://github.com/emilevauge))
- Traefik &#34;bug&#34; command documentation ([#1811](https://github.com/traefik/traefik/pull/1811) by [ldez](https://github.com/ldez))
- Add Marco Jantke to maintainers ([#1980](https://github.com/traefik/traefik/pull/1980) by [emilevauge](https://github.com/emilevauge))
- toml page - replace li by table ([#1995](https://github.com/traefik/traefik/pull/1995) by [jmaitrehenry](https://github.com/jmaitrehenry))
- Update golang version in contributing guide ([#2018](https://github.com/traefik/traefik/pull/2018) by [ArikaChen](https://github.com/ArikaChen))
- Release cycle. ([#1812](https://github.com/traefik/traefik/pull/1812) by [ldez](https://github.com/ldez))
- Remove Russel from maintainers ([#1614](https://github.com/traefik/traefik/pull/1614) by [emilevauge](https://github.com/emilevauge))
- Update CONTRIBUTING.md. ([#1667](https://github.com/traefik/traefik/pull/1667) by [timoreimann](https://github.com/timoreimann))
- drop &#34;slave&#34; wording for &#34;worker&#34; ([#1645](https://github.com/traefik/traefik/pull/1645) by [djalal](https://github.com/djalal))
- Use more inclusive language in README.md {guys =&gt; folks} ([#1640](https://github.com/traefik/traefik/pull/1640) by [igorwwwwwwwwwwwwwwwwwwww](https://github.com/igorwwwwwwwwwwwwwwwwwwww))
- Remove Thomas Recloux from maintainers ([#1616](https://github.com/traefik/traefik/pull/1616) by [emilevauge](https://github.com/emilevauge))
- Update documentation for 1.4 release ([#2011](https://github.com/traefik/traefik/pull/2011) by [emilevauge](https://github.com/emilevauge))
- Small toml documentation update ([#1603](https://github.com/traefik/traefik/pull/1603) by [antoine-aumjaud](https://github.com/antoine-aumjaud))
- Add @ldez to maintainers ([#1589](https://github.com/traefik/traefik/pull/1589) by [emilevauge](https://github.com/emilevauge))
- doc: add labels documentation. ([#1582](https://github.com/traefik/traefik/pull/1582) by [ldez](https://github.com/ldez))
- Change Traefik intro video ([#1893](https://github.com/traefik/traefik/pull/1893) by [emilevauge](https://github.com/emilevauge))
- Update GraceTimeOut documentation ([#1875](https://github.com/traefik/traefik/pull/1875) by [marco-jantke](https://github.com/marco-jantke))

**Misc:**
- Merge v1.3.7 ([#2013](https://github.com/traefik/traefik/pull/2013) by [ldez](https://github.com/ldez))
- Merge 1.3.6 ([#1992](https://github.com/traefik/traefik/pull/1992) by [ldez](https://github.com/ldez))
- Merge 1.3.5 ([#1909](https://github.com/traefik/traefik/pull/1909) by [ldez](https://github.com/ldez))
- Merge 1.3.3 ([#1836](https://github.com/traefik/traefik/pull/1836) by [ldez](https://github.com/ldez))
- Merge v1.3.2 to master  ([#1809](https://github.com/traefik/traefik/pull/1809) by [ldez](https://github.com/ldez))
- Merge current v1.3 ([#1797](https://github.com/traefik/traefik/pull/1797) by [ldez](https://github.com/ldez))
- Merge current v1.3 ([#1786](https://github.com/traefik/traefik/pull/1786) by [ldez](https://github.com/ldez))
- Merge v1.3.1 to master ([#1763](https://github.com/traefik/traefik/pull/1763) by [ldez](https://github.com/ldez))
- Merge current v1.3 ([#1753](https://github.com/traefik/traefik/pull/1753) by [ldez](https://github.com/ldez))
- Merge current v1.3 ([#1705](https://github.com/traefik/traefik/pull/1705) by [ldez](https://github.com/ldez))
- Merge current v1.3 to master ([#1697](https://github.com/traefik/traefik/pull/1697) by [ldez](https://github.com/ldez))
- Merge v1 3 0 ([#1692](https://github.com/traefik/traefik/pull/1692) by [ldez](https://github.com/ldez))
- Merge current v1.3 to master (rc3) ([#1666](https://github.com/traefik/traefik/pull/1666) by [ldez](https://github.com/ldez))
- Merge current v1.3 to master  ([#1643](https://github.com/traefik/traefik/pull/1643) by [ldez](https://github.com/ldez))
- Merge v1.3.0-rc2 master ([#1613](https://github.com/traefik/traefik/pull/1613) by [emilevauge](https://github.com/emilevauge))

## [v1.3.8](https://github.com/traefik/traefik/tree/v1.3.8) (2017-09-07)
[All Commits](https://github.com/traefik/traefik/compare/v1.3.7...v1.3.8)

**Bug fixes:**
- **[middleware]** Compress and Websocket ([#2079](https://github.com/traefik/traefik/pull/2079) by [ldez](https://github.com/ldez))

## [v1.3.7](https://github.com/traefik/traefik/tree/v1.3.7) (2017-08-25)
[All Commits](https://github.com/traefik/traefik/compare/v1.3.6...v1.3.7)

**Bug fixes:**
- **[oxy]** Only forward X-Forwarded-Port. ([#2007](https://github.com/traefik/traefik/pull/2007) by [ldez](https://github.com/ldez))

## [v1.3.6](https://github.com/traefik/traefik/tree/v1.3.6) (2017-08-20)
[All Commits](https://github.com/traefik/traefik/compare/v1.3.5...v1.3.6)

**Bug fixes:**
- **[oxy,websocket]** Websocket parameters and protocol. ([#1970](https://github.com/traefik/traefik/pull/1970) by [ldez](https://github.com/ldez))

## [v1.3.5](https://github.com/traefik/traefik/tree/v1.3.5) (2017-08-01)
[All Commits](https://github.com/traefik/traefik/compare/v1.3.4...v1.3.5)

**Bug fixes:**
- **[websocket]** Oxy with fixes on websocket + integration tests ([#1905](https://github.com/traefik/traefik/pull/1905) by [Juliens](https://github.com/Juliens))

## [v1.3.4](https://github.com/traefik/traefik/tree/v1.3.4) (2017-07-27)
[All Commits](https://github.com/traefik/traefik/compare/v1.3.3...v1.3.4)

**Bug fixes:**
- **[middleware]** Double compression. ([#1863](https://github.com/traefik/traefik/pull/1863) by [ldez](https://github.com/ldez))
- **[middleware]** Fix replace path rule ([#1859](https://github.com/traefik/traefik/pull/1859) by [dedalusj](https://github.com/dedalusj))
- **[websocket]** New oxy with gorilla for websocket with integration tests ([#1896](https://github.com/traefik/traefik/pull/1896) by [Juliens](https://github.com/Juliens))

## [v1.3.3](https://github.com/traefik/traefik/tree/v1.3.3) (2017-07-06)
[All Commits](https://github.com/traefik/traefik/compare/v1.3.2...v1.3.3)

**Bug fixes:**
- **[k8s]** Undo the Secrets controller sync wait. ([#1828](https://github.com/traefik/traefik/pull/1828) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Tell glog to log everything into STDERR. ([#1817](https://github.com/traefik/traefik/pull/1817) by [timoreimann](https://github.com/timoreimann))

## [v1.3.2](https://github.com/traefik/traefik/tree/v1.3.2) (2017-06-29)
[All Commits](https://github.com/traefik/traefik/compare/v1.3.1...v1.3.2)

**Bug fixes:**
- **[acme]** Add provided certificate checking before LE certificate generation with OnHostRule option ([#1772](https://github.com/traefik/traefik/pull/1772) by [nmengin](https://github.com/nmengin))
- **[k8s]** Fix race on closing event channel. ([#1798](https://github.com/traefik/traefik/pull/1798) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Upgrade go-marathon to dd6cbd4. ([#1800](https://github.com/traefik/traefik/pull/1800) by [timoreimann](https://github.com/timoreimann))
- **[oxy,websocket]** Problem with keepalive when switching protocol failed ([#1782](https://github.com/traefik/traefik/pull/1782) by [ldez](https://github.com/ldez))
- **[oxy]** Fix proxying of unannounced trailers ([#1805](https://github.com/traefik/traefik/pull/1805) by [ldez](https://github.com/ldez))

## [v1.3.1](https://github.com/traefik/traefik/tree/v1.3.1) (2017-06-16)
[All Commits](https://github.com/traefik/traefik/compare/v1.3.0...v1.3.1)

**Enhancements:**
- **[logs,eureka,marathon]** Minor logs changes ([#1749](https://github.com/traefik/traefik/pull/1749) by [ldez](https://github.com/ldez))

**Bug fixes:**
- **[k8s]** Use correct type when watching for k8s secrets ([#1700](https://github.com/traefik/traefik/pull/1700) by [kekoav](https://github.com/kekoav))
- **[middleware]** fix: Double compression. ([#1714](https://github.com/traefik/traefik/pull/1714) by [ldez](https://github.com/ldez))
- **[webui]** Don&#39;t fail when backend or frontend are empty. ([#1757](https://github.com/traefik/traefik/pull/1757) by [ldez](https://github.com/ldez))

**Documentation:**
- **[k8s]** Fix capitalization of PathPrefixStrip in kubernetes doc ([#1695](https://github.com/traefik/traefik/pull/1695) by [Miouge1](https://github.com/Miouge1))

## [v1.3.0](https://github.com/traefik/traefik/tree/v1.3.0) (2017-05-31)
[All Commits](https://github.com/traefik/traefik/compare/v1.2.0-rc1...v1.3.0)

**Enhancements:**
- **[acme]** Tighten regex match for wildcard certs [Addendum to #1018] ([#1227](https://github.com/traefik/traefik/pull/1227) by [dtomcej](https://github.com/dtomcej))
- **[api,webui]** Feature web root path ([#1233](https://github.com/traefik/traefik/pull/1233) by [tcoupin](https://github.com/tcoupin))
- **[authentication,docker,rancher]** Add Basic Auth per Frontend ([#1147](https://github.com/traefik/traefik/pull/1147) by [SantoDE](https://github.com/SantoDE))
- **[authentication]** Allow usersFile to be specified for basic or digest auth ([#1189](https://github.com/traefik/traefik/pull/1189) by [krancour](https://github.com/krancour))
- **[docker]** Allow multiple rules from docker labels containers with traefik.&lt;servicename&gt;.* properties ([#1257](https://github.com/traefik/traefik/pull/1257) by [benoitf](https://github.com/benoitf))
- **[docker]** Use docker-compose labels for frontend and backend names ([#1235](https://github.com/traefik/traefik/pull/1235) by [tcoupin](https://github.com/tcoupin))
- **[dynamodb]** add dynamodb backend ([#1158](https://github.com/traefik/traefik/pull/1158) by [tskinn](https://github.com/tskinn))
- **[healthcheck,consul]** using more sensible consul blocking query to detect health check changes ([#1241](https://github.com/traefik/traefik/pull/1241) by [vholovko](https://github.com/vholovko))
- **[healthcheck]** Add global health check interval parameter. ([#1338](https://github.com/traefik/traefik/pull/1338) by [timoreimann](https://github.com/timoreimann))
- **[healthcheck]** Start health checks early. ([#1319](https://github.com/traefik/traefik/pull/1319) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Upgrade k8s.io/client-go to version 2 ([#1178](https://github.com/traefik/traefik/pull/1178) by [errm](https://github.com/errm))
- **[k8s]** Support cluster-external Kubernetes client. ([#1159](https://github.com/traefik/traefik/pull/1159) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Add basic auth to kubernetes provider ([#1488](https://github.com/traefik/traefik/pull/1488) by [alpe](https://github.com/alpe))
- **[k8s]** Adding support for Traefik to respect the K8s ingress class annotation ([#1182](https://github.com/traefik/traefik/pull/1182) by [Regner](https://github.com/Regner))
- **[k8s]** Refactor k8s rule type annotation parsing/retrieval. ([#1151](https://github.com/traefik/traefik/pull/1151) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Kubernetes support externalname service ([#1149](https://github.com/traefik/traefik/pull/1149) by [Regner](https://github.com/Regner))
- **[kv]** Add libkv Username and Password ([#1357](https://github.com/traefik/traefik/pull/1357) by [tcolgate](https://github.com/tcolgate))
- **[kv]** kv: Ignore backend servers with no url ([#1196](https://github.com/traefik/traefik/pull/1196) by [klausenbusk](https://github.com/klausenbusk))
- **[logs]** New access logger ([#1408](https://github.com/traefik/traefik/pull/1408) by [rjshep](https://github.com/rjshep))
- **[logs]** Revert &#34;New access logger&#34; ([#1541](https://github.com/traefik/traefik/pull/1541) by [emilevauge](https://github.com/emilevauge))
- **[marathon]** Allow traefik.port to not be in the list of marathon ports ([#1394](https://github.com/traefik/traefik/pull/1394) by [emilevauge](https://github.com/emilevauge))
- **[marathon]** Add tests lost during PR 1320. ([#1540](https://github.com/traefik/traefik/pull/1540) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Make Traefik health checks label-configurable with Marathon. ([#1320](https://github.com/traefik/traefik/pull/1320) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Detect proper hostname automatically. ([#1345](https://github.com/traefik/traefik/pull/1345) by [diegooliveira](https://github.com/diegooliveira))
- **[rancher]** Added constraint management for Rancher provider ([#1527](https://github.com/traefik/traefik/pull/1527) by [yyekhlef](https://github.com/yyekhlef))
- **[rancher]** Improve rancher provider handling of service and container health states ([#1343](https://github.com/traefik/traefik/pull/1343) by [kelchm](https://github.com/kelchm))
- **[rancher]** Fix Rancher API pagination limits ([#1453](https://github.com/traefik/traefik/pull/1453) by [martinbaillie](https://github.com/martinbaillie))
- **[rancher]** Fix Rancher backend left in uncommented state ([#1455](https://github.com/traefik/traefik/pull/1455) by [martinbaillie](https://github.com/martinbaillie))
- **[rules]** Add Path Replacement Rule ([#1374](https://github.com/traefik/traefik/pull/1374) by [ssttevee](https://github.com/ssttevee))
- **[rules]** Add PathStripRegex rule ([#1339](https://github.com/traefik/traefik/pull/1339) by [seguins](https://github.com/seguins))
- **[webui]** Working UI ([#1542](https://github.com/traefik/traefik/pull/1542) by [maxwo](https://github.com/maxwo))
- **[webui]** Dashboard filter ([#1437](https://github.com/traefik/traefik/pull/1437) by [ldez](https://github.com/ldez))
- Upgrade dependencies. ([#1170](https://github.com/traefik/traefik/pull/1170) by [timoreimann](https://github.com/timoreimann))
- Bump go 1.8 ([#1259](https://github.com/traefik/traefik/pull/1259) by [emilevauge](https://github.com/emilevauge))
- Update TLS Ciphers for Go 1.8 ([#1276](https://github.com/traefik/traefik/pull/1276) by [kekoav](https://github.com/kekoav))
- Add IdleConnTimeout to Traefik&#39;s http.server settings ([#1340](https://github.com/traefik/traefik/pull/1340) by [bparli](https://github.com/bparli))
- Pass stripped prefix downstream as header ([#1442](https://github.com/traefik/traefik/pull/1442) by [martinbaillie](https://github.com/martinbaillie))
- Extract some code in packages ([#1449](https://github.com/traefik/traefik/pull/1449) by [vdemeester](https://github.com/vdemeester))
- Vendor generated file ([#1464](https://github.com/traefik/traefik/pull/1464) by [vdemeester](https://github.com/vdemeester))
- Add unit tests for package safe ([#1517](https://github.com/traefik/traefik/pull/1517) by [gottwald](https://github.com/gottwald))
- Use TOML-compatible duration type. ([#1350](https://github.com/traefik/traefik/pull/1350) by [timoreimann](https://github.com/timoreimann))
- Get testify/require dependency. ([#1658](https://github.com/traefik/traefik/pull/1658) by [timoreimann](https://github.com/timoreimann))

**Bug fixes:**
- **[consul]** fix consul sample endpoints ([#1303](https://github.com/traefik/traefik/pull/1303) by [ruslansennov](https://github.com/ruslansennov))
- **[consul]** Fix Consul catalog prefix flags ([#1486](https://github.com/traefik/traefik/pull/1486) by [emilevauge](https://github.com/emilevauge))
- **[docker]** Make port deterministic ([#1523](https://github.com/traefik/traefik/pull/1523) by [tanyadegurechaff](https://github.com/tanyadegurechaff))
- **[k8s]** Remove rule type path list. ([#1630](https://github.com/traefik/traefik/pull/1630) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Ignore Ingresses with empty Endpoint subsets. ([#1604](https://github.com/traefik/traefik/pull/1604) by [timoreimann](https://github.com/timoreimann))
- **[k8s]** Ignore missing pass host header annotation. ([#1581](https://github.com/traefik/traefik/pull/1581) by [timoreimann](https://github.com/timoreimann))
- **[logs]** Fix empty basic auth ([#1601](https://github.com/traefik/traefik/pull/1601) by [emilevauge](https://github.com/emilevauge))
- **[logs]** Create log folder if not present ([#1507](https://github.com/traefik/traefik/pull/1507) by [tanyadegurechaff](https://github.com/tanyadegurechaff))
- **[marathon]** Upgrade go-marathon to 15ea23e. ([#1635](https://github.com/traefik/traefik/pull/1635) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Fix default timeouts for Marathon provider. ([#1398](https://github.com/traefik/traefik/pull/1398) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Check for explicitly defined Marathon port first. ([#1474](https://github.com/traefik/traefik/pull/1474) by [timoreimann](https://github.com/timoreimann))
- **[marathon]** Bump go-marathon dep ([#1524](https://github.com/traefik/traefik/pull/1524) by [jangie](https://github.com/jangie))
- **[middleware,rules]** Fix behavior for PathPrefixStrip ([#1638](https://github.com/traefik/traefik/pull/1638) by [seryl](https://github.com/seryl))
- **[middleware,websocket]** Fix stats hijack ([#1598](https://github.com/traefik/traefik/pull/1598) by [emilevauge](https://github.com/emilevauge))
- **[provider]** Fix exported fields providers ([#1588](https://github.com/traefik/traefik/pull/1588) by [emilevauge](https://github.com/emilevauge))
- **[rancher]** fix: Empty Rancher Service Labels. ([#1654](https://github.com/traefik/traefik/pull/1654) by [ldez](https://github.com/ldez))
- **[sticky-session]** Maintain sticky flag on LB method validation failure. ([#1585](https://github.com/traefik/traefik/pull/1585) by [timoreimann](https://github.com/timoreimann))
- Revert &#34;Vendor generated file&#34; ([#1534](https://github.com/traefik/traefik/pull/1534) by [ldez](https://github.com/ldez))
- Update golang.org/x/sys to fix windows compilation ([#1448](https://github.com/traefik/traefik/pull/1448) by [vdemeester](https://github.com/vdemeester))
- Fix systemd watchdog feature ([#1525](https://github.com/traefik/traefik/pull/1525) by [guilhem](https://github.com/guilhem))
- Fixed ReplacePath rule executing out of order, when combined with PathPrefixStrip ([#1577](https://github.com/traefik/traefik/pull/1577) by [aantono](https://github.com/aantono))

**Documentation:**
- **[cluster]** doc: Traefik cluster in beta. ([#1610](https://github.com/traefik/traefik/pull/1610) by [ldez](https://github.com/ldez))
- **[docker]** Fix error in documentation for Docker labels ([#1179](https://github.com/traefik/traefik/pull/1179) by [bgandon](https://github.com/bgandon))
- **[k8s]** Re Organise k8s docs to make 1.6 usage easier ([#1602](https://github.com/traefik/traefik/pull/1602) by [errm](https://github.com/errm))
- **[k8s]** Add documentation for k8s RBAC configuration ([#1404](https://github.com/traefik/traefik/pull/1404) by [aolwas](https://github.com/aolwas))
- **[k8s]** Add documentation about k8s Helm Chart ([#1367](https://github.com/traefik/traefik/pull/1367) by [seguins](https://github.com/seguins))
- **[marathon]** Add Marathon guide. ([#1578](https://github.com/traefik/traefik/pull/1578) by [Stibbons](https://github.com/Stibbons))
- **[metrics]** Fix prometheus metrics example ([#1157](https://github.com/traefik/traefik/pull/1157) by [solidnerd](https://github.com/solidnerd))
- **[metrics]** Make toml Bucket array homogeneous ([#1369](https://github.com/traefik/traefik/pull/1369) by [Starefossen](https://github.com/Starefossen))
- **[rancher]** make docs more clear about how to work with the current api ([#1337](https://github.com/traefik/traefik/pull/1337) by [SantoDE](https://github.com/SantoDE))
- **[rules]** Motivate and explain regular expression rules. ([#1216](https://github.com/traefik/traefik/pull/1216) by [timoreimann](https://github.com/timoreimann))
- **[rules]** Improve documentation for frontend rules. ([#1469](https://github.com/traefik/traefik/pull/1469) by [timoreimann](https://github.com/timoreimann))
- License 2017, Træfɪk =&gt; Træfik ([#1368](https://github.com/traefik/traefik/pull/1368) by [emilevauge](https://github.com/emilevauge))
- update wording ([#1458](https://github.com/traefik/traefik/pull/1458) by [ben-st](https://github.com/ben-st))
- Fix typo in command line help. ([#1467](https://github.com/traefik/traefik/pull/1467) by [mattcollier](https://github.com/mattcollier))
- Mention Traefik pronunciation in docs too. ([#1468](https://github.com/traefik/traefik/pull/1468) by [timoreimann](https://github.com/timoreimann))
- Correct typo in code comment. ([#1473](https://github.com/traefik/traefik/pull/1473) by [mattcollier](https://github.com/mattcollier))
- Change a word in the documentation ([#1274](https://github.com/traefik/traefik/pull/1274) by [sroze](https://github.com/sroze))
- Add @trecloux to Maintainers ([#1226](https://github.com/traefik/traefik/pull/1226) by [emilevauge](https://github.com/emilevauge))
- doc: enhance GitHub template. ([#1482](https://github.com/traefik/traefik/pull/1482) by [ldez](https://github.com/ldez))
- Add @timoreimann to list of maintainers. ([#1215](https://github.com/traefik/traefik/pull/1215) by [timoreimann](https://github.com/timoreimann))
- Add Traefik TOML sample section on how to bind to specific IP addr. ([#1194](https://github.com/traefik/traefik/pull/1194) by [timoreimann](https://github.com/timoreimann))
- doc: enhance Github templates. ([#1515](https://github.com/traefik/traefik/pull/1515) by [ldez](https://github.com/ldez))
- doc: small documentation review ([#1516](https://github.com/traefik/traefik/pull/1516) by [ldez](https://github.com/ldez))

**Misc:**
- **[docker]** Few refactoring around the docker provider ([#1440](https://github.com/traefik/traefik/pull/1440) by [vdemeester](https://github.com/vdemeester))
- **[k8s]** Updating Kubernetes tests to properly test missing endpoints code path ([#1436](https://github.com/traefik/traefik/pull/1436) by [Regner](https://github.com/Regner))
- **[provider]** Extract providers to their own packages ([#1444](https://github.com/traefik/traefik/pull/1444) by [vdemeester](https://github.com/vdemeester))
- Fix typo in server.go ([#1386](https://github.com/traefik/traefik/pull/1386) by [mihaitodor](https://github.com/mihaitodor))
- Vendor dependencies ([#1144](https://github.com/traefik/traefik/pull/1144) by [timoreimann](https://github.com/timoreimann))
- Prepare release v1.3.0-rc3 ([#1661](https://github.com/traefik/traefik/pull/1661) by [ldez](https://github.com/ldez))
- Prepare release v1.3.0-rc2 ([#1606](https://github.com/traefik/traefik/pull/1606) by [emilevauge](https://github.com/emilevauge))
- Prepare release v1.3.0-rc1 ([#1553](https://github.com/traefik/traefik/pull/1553) by [emilevauge](https://github.com/emilevauge))
- Merge v1.2.3 master ([#1538](https://github.com/traefik/traefik/pull/1538) by [emilevauge](https://github.com/emilevauge))
- Merge v1.2.1 master ([#1383](https://github.com/traefik/traefik/pull/1383) by [emilevauge](https://github.com/emilevauge))
- Merge v1.2.0 rc2 master ([#1208](https://github.com/traefik/traefik/pull/1208) by [emilevauge](https://github.com/emilevauge))

## [v1.3.0-rc3](https://github.com/traefik/traefik/tree/v1.3.0-rc3) (2017-05-24)
[All Commits](https://github.com/traefik/traefik/compare/v1.3.0-rc2...v1.3.0-rc3)

**Enhancements:**
- [#1658](https://github.com/traefik/traefik/issues/1658) Get testify/require dependency. ([timoreimann](https://github.com/timoreimann))

**Bug fixes:**
- [#1507](https://github.com/traefik/traefik/issues/1507) Create log folder if not present ([tanyadegurechaff](https://github.com/tanyadegurechaff))
- [#1604](https://github.com/traefik/traefik/issues/1604) [k8s] Ignore Ingresses with empty Endpoint subsets. ([timoreimann](https://github.com/timoreimann))
- [#1630](https://github.com/traefik/traefik/issues/1630) [k8s] Remove rule type path list. ([timoreimann](https://github.com/timoreimann))
- [#1635](https://github.com/traefik/traefik/issues/1635) Upgrade go-marathon to 15ea23e. ([timoreimann](https://github.com/timoreimann))
- [#1638](https://github.com/traefik/traefik/issues/1638) Fix behavior for PathPrefixStrip ([seryl](https://github.com/seryl))
- [#1654](https://github.com/traefik/traefik/issues/1654) fix: Empty Rancher Service Labels. ([ldez](https://github.com/ldez))

**Documentation:**
- [#1578](https://github.com/traefik/traefik/issues/1578) Add Marathon guide. ([Stibbons](https://github.com/Stibbons))
- [#1602](https://github.com/traefik/traefik/issues/1602) Re Organise k8s docs to make 1.6 usage easier ([errm](https://github.com/errm))
- [#1642](https://github.com/traefik/traefik/issues/1642) Update changelog ([ldez](https://github.com/ldez))

## [v1.3.0-rc2](https://github.com/traefik/traefik/tree/v1.3.0-rc2) (2017-05-16)
[All Commits](https://github.com/traefik/traefik/compare/v1.3.0-rc1...v1.3.0-rc2)

**Enhancements:**
- Fixed ReplacePath rule executing out of order, when combined with PathPrefixStrip [#1577](https://github.com/traefik/traefik/issues/1577) ([aantono](https://github.com/aantono))

**Bug fixes:**
- [Kubernetes] Ignore missing pass host header annotation. [#1581](https://github.com/traefik/traefik/issues/1581) ([timoreimann](https://github.com/timoreimann))
- Maintain sticky flag on LB method validation failure. [#1585](https://github.com/traefik/traefik/issues/1585) ([timoreimann](https://github.com/timoreimann))
- Fix exported fields providers [#1588](https://github.com/traefik/traefik/issues/1588) ([emilevauge](https://github.com/emilevauge))
- Fix stats hijack [#1598](https://github.com/traefik/traefik/issues/1598) ([emilevauge](https://github.com/emilevauge))
- Fix empty basic auth [#1601](https://github.com/traefik/traefik/issues/1601) ([emilevauge](https://github.com/emilevauge))

**Documentation:**
- doc: Traefik cluster in beta. [#1610](https://github.com/traefik/traefik/issues/1610) ([ldez](https://github.com/ldez))

## [v1.3.0-rc1](https://github.com/traefik/traefik/tree/v1.3.0-rc1) (2017-05-05)
[All Commits](https://github.com/traefik/traefik/compare/v1.2.0-rc1...v1.3.0-rc1)

**Enhancements:**
- Add Basic Auth per Frontend [#1147](https://github.com/traefik/traefik/issues/1147) ([SantoDE](https://github.com/SantoDE))
- Kubernetes support externalname service [#1149](https://github.com/traefik/traefik/issues/1149) ([Regner](https://github.com/Regner))
- add dynamodb backend [#1158](https://github.com/traefik/traefik/issues/1158) ([tskinn](https://github.com/tskinn))
- Support cluster-external Kubernetes client. [#1159](https://github.com/traefik/traefik/issues/1159) ([timoreimann](https://github.com/timoreimann))
- Add Traefik TOML sample section on how to bind to specific IP addr. [#1194](https://github.com/traefik/traefik/issues/1194) ([timoreimann](https://github.com/timoreimann))
- kv: Ignore backend servers with no url [#1196](https://github.com/traefik/traefik/issues/1196) ([klausenbusk](https://github.com/klausenbusk))
- Tighten regex match for wildcard certs [Addendum to #1018] [#1227](https://github.com/traefik/traefik/issues/1227) ([dtomcej](https://github.com/dtomcej))
- Feature web root path [#1233](https://github.com/traefik/traefik/issues/1233) ([tcoupin](https://github.com/tcoupin))
- using more sensible consul blocking query to detect health check changes [#1241](https://github.com/traefik/traefik/issues/1241) ([vholovko](https://github.com/vholovko))
- Allow multiple rules from docker labels containers with traefik.&lt;servicename&gt;.* properties [#1257](https://github.com/traefik/traefik/issues/1257) ([benoitf](https://github.com/benoitf))
- Update TLS Ciphers for Go 1.8 [#1276](https://github.com/traefik/traefik/issues/1276) ([kekoav](https://github.com/kekoav))
- Start health checks early. [#1319](https://github.com/traefik/traefik/issues/1319) ([timoreimann](https://github.com/timoreimann))
- Make Traefik health checks label-configurable with Marathon. [#1320](https://github.com/traefik/traefik/issues/1320) ([timoreimann](https://github.com/timoreimann))
- Append template section asking for debug log output. [#1324](https://github.com/traefik/traefik/issues/1324) ([timoreimann](https://github.com/timoreimann))
- Add global health check interval parameter. [#1338](https://github.com/traefik/traefik/issues/1338) ([timoreimann](https://github.com/timoreimann))
- Fix regex with PathStrip [#1339](https://github.com/traefik/traefik/issues/1339) ([seguins](https://github.com/seguins))
- Add IdleConnTimeout to Traefik&#39;s http.server settings [#1340](https://github.com/traefik/traefik/issues/1340) ([bparli](https://github.com/bparli))
- Improve rancher provider handling of service and container health states [#1343](https://github.com/traefik/traefik/issues/1343) ([kelchm](https://github.com/kelchm))
- [Marathon] Detect proper hostname automatically. [#1345](https://github.com/traefik/traefik/issues/1345) ([diegooliveira](https://github.com/diegooliveira))
- Use TOML-compatible duration type. [#1350](https://github.com/traefik/traefik/issues/1350) ([timoreimann](https://github.com/timoreimann))
- Add libkv Username and Password [#1357](https://github.com/traefik/traefik/issues/1357) ([tcolgate](https://github.com/tcolgate))
- Make toml Bucket array homogeneous [#1369](https://github.com/traefik/traefik/issues/1369) ([Starefossen](https://github.com/Starefossen))
- Add Path Replacement Rule [#1374](https://github.com/traefik/traefik/issues/1374) ([ssttevee](https://github.com/ssttevee))
- New access logger [#1408](https://github.com/traefik/traefik/issues/1408) ([rjshep](https://github.com/rjshep))
- feat(webui): Dashboard filter [#1437](https://github.com/traefik/traefik/issues/1437) ([ldez](https://github.com/ldez))
- Pass stripped prefix downstream as header (#985) [#1442](https://github.com/traefik/traefik/issues/1442) ([martinbaillie](https://github.com/martinbaillie))
- Extract some code in packages [#1449](https://github.com/traefik/traefik/issues/1449) ([vdemeester](https://github.com/vdemeester))
- Fix Rancher API pagination limits [#1453](https://github.com/traefik/traefik/issues/1453) ([martinbaillie](https://github.com/martinbaillie))
- Fix Rancher backend left in uncommented state [#1455](https://github.com/traefik/traefik/issues/1455) ([martinbaillie](https://github.com/martinbaillie))
- Vendor generated file [#1464](https://github.com/traefik/traefik/issues/1464) ([vdemeester](https://github.com/vdemeester))
- Add basic auth to kubernetes provider [#1488](https://github.com/traefik/traefik/issues/1488) ([alpe](https://github.com/alpe))
- Add unit tests for package safe [#1517](https://github.com/traefik/traefik/issues/1517) ([gottwald](https://github.com/gottwald))
- feat(rancher): added constraint management for rancher provider [#1527](https://github.com/traefik/traefik/issues/1527) ([yyekhlef](https://github.com/yyekhlef))
- refactor: fix for PR with master branch. [#1537](https://github.com/traefik/traefik/issues/1537) ([ldez](https://github.com/ldez))
- Add tests lost during PR 1320. [#1540](https://github.com/traefik/traefik/issues/1540) ([timoreimann](https://github.com/timoreimann))
- Working UI [#1542](https://github.com/traefik/traefik/issues/1542) ([maxwo](https://github.com/maxwo))

**Bug fixes:**
- Fix default timeouts for Marathon provider. [#1398](https://github.com/traefik/traefik/issues/1398) ([timoreimann](https://github.com/timoreimann))
- Update golang.org/x/sys to fix windows compilation [#1448](https://github.com/traefik/traefik/issues/1448) ([vdemeester](https://github.com/vdemeester))
- Check for explicitly defined Marathon port first. [#1474](https://github.com/traefik/traefik/issues/1474) ([timoreimann](https://github.com/timoreimann))
- Fix Consul catalog prefix flags [#1486](https://github.com/traefik/traefik/issues/1486) ([emilevauge](https://github.com/emilevauge))
- Move Docker test provider instantiation into t.Run body. [#1489](https://github.com/traefik/traefik/issues/1489) ([timoreimann](https://github.com/timoreimann))
- Make port deterministic [#1523](https://github.com/traefik/traefik/issues/1523) ([tanyadegurechaff](https://github.com/tanyadegurechaff))
- [Marathon] Bump go-marathon dep [#1524](https://github.com/traefik/traefik/issues/1524) ([jangie](https://github.com/jangie))
- Fix systemd watchdog feature [#1525](https://github.com/traefik/traefik/issues/1525) ([guilhem](https://github.com/guilhem))
- Revert &#34;Vendor generated file&#34; [#1534](https://github.com/traefik/traefik/issues/1534) ([ldez](https://github.com/ldez))

**Documentation:**
- Fix prometheus metrics example [#1157](https://github.com/traefik/traefik/issues/1157) ([solidnerd](https://github.com/solidnerd))
- Fix error in documentation for Docker labels [#1179](https://github.com/traefik/traefik/issues/1179) ([bgandon](https://github.com/bgandon))
- Motivate and explain regular expression rules. [#1216](https://github.com/traefik/traefik/issues/1216) ([timoreimann](https://github.com/timoreimann))
- Add @trecloux to Maintainers [#1226](https://github.com/traefik/traefik/issues/1226) ([emilevauge](https://github.com/emilevauge))
- Change a word in the documentation [#1274](https://github.com/traefik/traefik/issues/1274) ([sroze](https://github.com/sroze))
- make docs more clear about how to work with the current api [#1337](https://github.com/traefik/traefik/issues/1337) ([SantoDE](https://github.com/SantoDE))
- Add documentation about k8s Helm Chart [#1367](https://github.com/traefik/traefik/issues/1367) ([seguins](https://github.com/seguins))
- License 2017, Træfɪk =&gt; Træfik [#1368](https://github.com/traefik/traefik/issues/1368) ([emilevauge](https://github.com/emilevauge))
- Add documentation for k8s RBAC configuration [#1404](https://github.com/traefik/traefik/issues/1404) ([aolwas](https://github.com/aolwas))
- update wording [#1458](https://github.com/traefik/traefik/issues/1458) ([ben-st](https://github.com/ben-st))
- Fix typo in command line help. [#1467](https://github.com/traefik/traefik/issues/1467) ([mattcollier](https://github.com/mattcollier))
- Mention Traefik pronunciation in docs too. [#1468](https://github.com/traefik/traefik/issues/1468) ([timoreimann](https://github.com/timoreimann))
- Improve documentation for frontend rules. [#1469](https://github.com/traefik/traefik/issues/1469) ([timoreimann](https://github.com/timoreimann))
- Correct typo in code comment. [#1473](https://github.com/traefik/traefik/issues/1473) ([mattcollier](https://github.com/mattcollier))
- doc: enhance GitHub template. [#1482](https://github.com/traefik/traefik/issues/1482) ([ldez](https://github.com/ldez))
- doc: enhance Github templates. [#1515](https://github.com/traefik/traefik/issues/1515) ([ldez](https://github.com/ldez))
- doc: small documentation review [#1516](https://github.com/traefik/traefik/issues/1516) ([ldez](https://github.com/ldez))

**Misc:**
- Vendor dependencies [#1144](https://github.com/traefik/traefik/issues/1144) ([timoreimann](https://github.com/timoreimann))
- Refactor k8s rule type annotation parsing/retrieval. [#1151](https://github.com/traefik/traefik/issues/1151) ([timoreimann](https://github.com/timoreimann))
- Upgrade dependencies. [#1170](https://github.com/traefik/traefik/issues/1170) ([timoreimann](https://github.com/timoreimann))
- Remove .gitattributes file. [#1172](https://github.com/traefik/traefik/issues/1172) ([timoreimann](https://github.com/timoreimann))
- Upgrade k8s.io/client-go to version 2 [#1178](https://github.com/traefik/traefik/issues/1178) ([errm](https://github.com/errm))
- Adding support for Traefik to respect the K8s ingress class annotation [#1182](https://github.com/traefik/traefik/issues/1182) ([Regner](https://github.com/Regner))
- Allow usersFile to be specified for basic or digest auth [#1189](https://github.com/traefik/traefik/issues/1189) ([krancour](https://github.com/krancour))
- Merge v1.2.0 rc2 master [#1208](https://github.com/traefik/traefik/issues/1208) ([emilevauge](https://github.com/emilevauge))
- Add @timoreimann to list of maintainers. [#1215](https://github.com/traefik/traefik/issues/1215) ([timoreimann](https://github.com/timoreimann))
- Use docker-compose labels for frontend and backend names [#1235](https://github.com/traefik/traefik/issues/1235) ([tcoupin](https://github.com/tcoupin))
- Bump go 1.8 [#1259](https://github.com/traefik/traefik/issues/1259) ([emilevauge](https://github.com/emilevauge))
- fix consul sample endpoints [#1303](https://github.com/traefik/traefik/issues/1303) ([ruslansennov](https://github.com/ruslansennov))
- Merge v1.2.1 master [#1383](https://github.com/traefik/traefik/issues/1383) ([emilevauge](https://github.com/emilevauge))
- Fix typo in server.go [#1386](https://github.com/traefik/traefik/issues/1386) ([mihaitodor](https://github.com/mihaitodor))
- Allow traefik.port to not be in the list of marathon ports [#1394](https://github.com/traefik/traefik/issues/1394) ([emilevauge](https://github.com/emilevauge))
- Updating Kubernetes tests to properly test missing endpoints code path [#1436](https://github.com/traefik/traefik/issues/1436) ([Regner](https://github.com/Regner))
- Few refactoring around the docker provider [#1440](https://github.com/traefik/traefik/issues/1440) ([vdemeester](https://github.com/vdemeester))
- Extract providers to their own packages [#1444](https://github.com/traefik/traefik/issues/1444) ([vdemeester](https://github.com/vdemeester))
- Merge v1.2.3 master [#1538](https://github.com/traefik/traefik/issues/1538) ([emilevauge](https://github.com/emilevauge))
- Revert &#34;First stage of access logging middleware.  Initially without … [#1541](https://github.com/traefik/traefik/issues/1541) ([emilevauge](https://github.com/emilevauge))
- Prepare release v1.3.0-rc1 [#1553](https://github.com/traefik/traefik/issues/1553) ([emilevauge](https://github.com/emilevauge))

## [v1.2.3](https://github.com/traefik/traefik/tree/v1.2.3) (2017-04-13)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.2.2...v1.2.3)

**Merged pull requests:**

- Fix too many redirect [\#1433](https://github.com/traefik/traefik/pull/1433) ([emilevauge](https://github.com/emilevauge))

## [v1.2.2](https://github.com/traefik/traefik/tree/v1.2.2) (2017-04-11)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.2.1...v1.2.2)

**Merged pull requests:**

- Carry PR 1271 [\#1417](https://github.com/traefik/traefik/pull/1417) ([emilevauge](https://github.com/emilevauge))
- Fix postloadconfig acme & Docker filter empty rule [\#1401](https://github.com/traefik/traefik/pull/1401) ([emilevauge](https://github.com/emilevauge))

## [v1.2.1](https://github.com/traefik/traefik/tree/v1.2.1) (2017-03-27)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.2.0...v1.2.1)

**Merged pull requests:**

- bump lego 0e2937900 [\#1347](https://github.com/traefik/traefik/pull/1347) ([emilevauge](https://github.com/emilevauge))
- k8s: Do not log service fields when GetService is failing. [\#1331](https://github.com/traefik/traefik/pull/1331) ([timoreimann](https://github.com/timoreimann))

## [v1.2.0](https://github.com/traefik/traefik/tree/v1.2.0) (2017-03-20)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.1.2...v1.2.0)

**Merged pull requests:**

- Docker: Added warning if network could not be found [\#1310](https://github.com/traefik/traefik/pull/1310) ([zweizeichen](https://github.com/zweizeichen))
- Add filter on task status in addition to desired status \(Docker Provider - swarm\) [\#1304](https://github.com/traefik/traefik/pull/1304) ([Yshayy](https://github.com/Yshayy))
- Abort Kubernetes Ingress update if Kubernetes API call fails [\#1295](https://github.com/traefik/traefik/pull/1295) ([Regner](https://github.com/Regner))
- Small fixes [\#1291](https://github.com/traefik/traefik/pull/1291) ([emilevauge](https://github.com/emilevauge))
- Rename health check URL parameter to path. [\#1285](https://github.com/traefik/traefik/pull/1285) ([timoreimann](https://github.com/timoreimann))
- Update Oxy, fix for \#1199 [\#1278](https://github.com/traefik/traefik/pull/1278) ([akanto](https://github.com/akanto))
- Fix metrics registering [\#1258](https://github.com/traefik/traefik/pull/1258) ([matevzmihalic](https://github.com/matevzmihalic))
- Update DefaultMaxIdleConnsPerHost default in docs. [\#1239](https://github.com/traefik/traefik/pull/1239) ([timoreimann](https://github.com/timoreimann))
- Update WSS/WS Proto \[Fixes \#670\] [\#1225](https://github.com/traefik/traefik/pull/1225) ([dtomcej](https://github.com/dtomcej))
- Bump go-rancher version [\#1219](https://github.com/traefik/traefik/pull/1219) ([SantoDE](https://github.com/SantoDE))
- Chunk taskArns into groups of 100 [\#1209](https://github.com/traefik/traefik/pull/1209) ([owen](https://github.com/owen))
- Prepare release v1.2.0 rc2 [\#1204](https://github.com/traefik/traefik/pull/1204) ([emilevauge](https://github.com/emilevauge))
- Revert "Ensure that we don't add balances with no health check runs … [\#1198](https://github.com/traefik/traefik/pull/1198) ([jangie](https://github.com/jangie))
- Small fixes and improvements [\#1173](https://github.com/traefik/traefik/pull/1173) ([SantoDE](https://github.com/SantoDE))
- Fix docker issues with global and dead tasks [\#1167](https://github.com/traefik/traefik/pull/1167) ([christopherobin](https://github.com/christopherobin))
- Better ECS error checking [\#1143](https://github.com/traefik/traefik/pull/1143) ([lpetre](https://github.com/lpetre))
- Fix stats race condition [\#1141](https://github.com/traefik/traefik/pull/1141) ([emilevauge](https://github.com/emilevauge))
- ECS: Docs - info about cred. resolution and required access policies [\#1137](https://github.com/traefik/traefik/pull/1137) ([rickard-von-essen](https://github.com/rickard-von-essen))
- Healthcheck tests and doc [\#1132](https://github.com/traefik/traefik/pull/1132) ([Juliens](https://github.com/Juliens))
- Fix travis deploy [\#1128](https://github.com/traefik/traefik/pull/1128) ([emilevauge](https://github.com/emilevauge))
- Prepare release v1.2.0 rc1 [\#1126](https://github.com/traefik/traefik/pull/1126) ([emilevauge](https://github.com/emilevauge))
- Fix checkout initial before calling rmpr [\#1124](https://github.com/traefik/traefik/pull/1124) ([emilevauge](https://github.com/emilevauge))
- Feature rancher integration [\#1120](https://github.com/traefik/traefik/pull/1120) ([SantoDE](https://github.com/SantoDE))
- Fix glide go units [\#1119](https://github.com/traefik/traefik/pull/1119) ([emilevauge](https://github.com/emilevauge))
- Carry \#818 —  Add systemd watchdog feature [\#1116](https://github.com/traefik/traefik/pull/1116) ([vdemeester](https://github.com/vdemeester))
- Skip file permission check on Windows [\#1115](https://github.com/traefik/traefik/pull/1115) ([StefanScherer](https://github.com/StefanScherer))
- Fix Docker API version for Windows [\#1113](https://github.com/traefik/traefik/pull/1113) ([StefanScherer](https://github.com/StefanScherer))
- Fix git rpr [\#1109](https://github.com/traefik/traefik/pull/1109) ([emilevauge](https://github.com/emilevauge))
- Fix docker version specifier [\#1108](https://github.com/traefik/traefik/pull/1108) ([timoreimann](https://github.com/timoreimann))
- Merge v1.1.2 master [\#1105](https://github.com/traefik/traefik/pull/1105) ([emilevauge](https://github.com/emilevauge))
- add sh before script in deploy... [\#1103](https://github.com/traefik/traefik/pull/1103) ([emilevauge](https://github.com/emilevauge))
- \[doc\] typo fixes for kubernetes user guide [\#1102](https://github.com/traefik/traefik/pull/1102) ([bamarni](https://github.com/bamarni))
- add skip\_cleanup in deploy [\#1101](https://github.com/traefik/traefik/pull/1101) ([emilevauge](https://github.com/emilevauge))
- Fix k8s example UI port. [\#1098](https://github.com/traefik/traefik/pull/1098) ([ddunkin](https://github.com/ddunkin))
- Fix marathon provider [\#1090](https://github.com/traefik/traefik/pull/1090) ([diegooliveira](https://github.com/diegooliveira))
- Add an ECS provider [\#1088](https://github.com/traefik/traefik/pull/1088) ([lpetre](https://github.com/lpetre))
- Update comment to reflect the code [\#1087](https://github.com/traefik/traefik/pull/1087) ([np](https://github.com/np))
- update NYTimes/gziphandler fixes \#1059 [\#1084](https://github.com/traefik/traefik/pull/1084) ([JamesKyburz](https://github.com/JamesKyburz))
- Ensure that we don't add balances with no health check runs if there is a health check defined on it [\#1080](https://github.com/traefik/traefik/pull/1080) ([jangie](https://github.com/jangie))
- Add FreeBSD & OpenBSD to crossbinary [\#1078](https://github.com/traefik/traefik/pull/1078) ([geoffgarside](https://github.com/geoffgarside))
- Fix metrics for multiple entry points [\#1071](https://github.com/traefik/traefik/pull/1071) ([matevzmihalic](https://github.com/matevzmihalic))
- Allow setting load balancer method and sticky using service annotations [\#1068](https://github.com/traefik/traefik/pull/1068) ([bakins](https://github.com/bakins))
- Fix travis script [\#1067](https://github.com/traefik/traefik/pull/1067) ([emilevauge](https://github.com/emilevauge))
- Add missing fmt verb specifier in k8s provider. [\#1066](https://github.com/traefik/traefik/pull/1066) ([timoreimann](https://github.com/timoreimann))
- Add git rpr command [\#1063](https://github.com/traefik/traefik/pull/1063) ([emilevauge](https://github.com/emilevauge))
- Fix k8s example [\#1062](https://github.com/traefik/traefik/pull/1062) ([emilevauge](https://github.com/emilevauge))
- Replace underscores to dash in autogenerated urls \(docker provider\) [\#1061](https://github.com/traefik/traefik/pull/1061) ([WTFKr0](https://github.com/WTFKr0))
- Don't run go test on .glide cache folder [\#1057](https://github.com/traefik/traefik/pull/1057) ([vdemeester](https://github.com/vdemeester))
- Allow setting circuitbreaker expression via Kubernetes annotation [\#1056](https://github.com/traefik/traefik/pull/1056) ([bakins](https://github.com/bakins))
- Improving instrumentation. [\#1042](https://github.com/traefik/traefik/pull/1042) ([enxebre](https://github.com/enxebre))
- Update user guide for upcoming `docker stack deploy`  [\#1041](https://github.com/traefik/traefik/pull/1041) ([twelvelabs](https://github.com/twelvelabs))
- Support sticky sessions under SWARM Mode. \#1024 [\#1033](https://github.com/traefik/traefik/pull/1033) ([foleymic](https://github.com/foleymic))
- Allow for wildcards in k8s ingress host, fixes \#792 [\#1029](https://github.com/traefik/traefik/pull/1029) ([sheerun](https://github.com/sheerun))
- Don't fetch ACME certificates for frontends using non-TLS entrypoints \(\#989\) [\#1023](https://github.com/traefik/traefik/pull/1023) ([syfonseq](https://github.com/syfonseq))
- Return Proper Non-ACME certificate - Fixes Issue 672 [\#1018](https://github.com/traefik/traefik/pull/1018) ([dtomcej](https://github.com/dtomcej))
- Fix docs build and add missing benchmarks page [\#1017](https://github.com/traefik/traefik/pull/1017) ([csabapalfi](https://github.com/csabapalfi))
- Set a NopCloser request body with retry middleware [\#1016](https://github.com/traefik/traefik/pull/1016) ([bamarni](https://github.com/bamarni))
- instruct to flatten dependencies with glide [\#1010](https://github.com/traefik/traefik/pull/1010) ([bamarni](https://github.com/bamarni))
- check permissions on acme.json during startup [\#1009](https://github.com/traefik/traefik/pull/1009) ([bamarni](https://github.com/bamarni))
- \[doc\] few tweaks on the basics page [\#1005](https://github.com/traefik/traefik/pull/1005) ([bamarni](https://github.com/bamarni))
- Import order as goimports does [\#1004](https://github.com/traefik/traefik/pull/1004) ([vdemeester](https://github.com/vdemeester))
- See the right go report badge [\#991](https://github.com/traefik/traefik/pull/991) ([guilhem](https://github.com/guilhem))
- Add multiple values for one rule to docs [\#978](https://github.com/traefik/traefik/pull/978) ([j0hnsmith](https://github.com/j0hnsmith))
- Add ACME/Let’s Encrypt integration tests [\#975](https://github.com/traefik/traefik/pull/975) ([trecloux](https://github.com/trecloux))
- deploy.sh: upload release source tarball [\#969](https://github.com/traefik/traefik/pull/969) ([Mic92](https://github.com/Mic92))
- toml zookeeper doc fix [\#948](https://github.com/traefik/traefik/pull/948) ([brdude](https://github.com/brdude))
- Add Rule AddPrefix [\#931](https://github.com/traefik/traefik/pull/931) ([Juliens](https://github.com/Juliens))
- Add bug command [\#921](https://github.com/traefik/traefik/pull/921) ([emilevauge](https://github.com/emilevauge))
- \(WIP\) feat: HealthCheck [\#918](https://github.com/traefik/traefik/pull/918) ([Juliens](https://github.com/Juliens))
- Add ability to set authenticated user in request header [\#889](https://github.com/traefik/traefik/pull/889) ([ViViDboarder](https://github.com/ViViDboarder))
- IP-per-task: [\#841](https://github.com/traefik/traefik/pull/841) ([diegooliveira](https://github.com/diegooliveira))

## [v1.2.0-rc2](https://github.com/traefik/traefik/tree/v1.2.0-rc2) (2017-03-01)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.2.0-rc1...v1.2.0-rc2)

**Implemented enhancements:**

- Are there plans to support the service type ExternalName in Kubernetes? [\#1142](https://github.com/traefik/traefik/issues/1142)
- Kubernetes Ingress and sticky support [\#911](https://github.com/traefik/traefik/issues/911)
- kubernetes client does not support InsecureSkipVerify [\#876](https://github.com/traefik/traefik/issues/876)
- Support active health checking like HAProxy [\#824](https://github.com/traefik/traefik/issues/824)
- Allow k8s ingress controller serviceAccountToken and serviceAccountCACert to be changed [\#611](https://github.com/traefik/traefik/issues/611)

**Fixed bugs:**

- \[rancher\] invalid memory address or nil pointer dereference [\#1134](https://github.com/traefik/traefik/issues/1134)
- Kubernetes default backend should work [\#1073](https://github.com/traefik/traefik/issues/1073)

**Closed issues:**

- Are release Download links broken? [\#1201](https://github.com/traefik/traefik/issues/1201)
- Bind to specific ip address [\#1193](https://github.com/traefik/traefik/issues/1193)
- DNS01 challenge use the wrong zone through route53 [\#1192](https://github.com/traefik/traefik/issues/1192)
- Reverse proxy https to http backends fails [\#1180](https://github.com/traefik/traefik/issues/1180)
- Swarm Mode + Letsencrypt + KV Store [\#1176](https://github.com/traefik/traefik/issues/1176)
- docker deploy -c example.yml    e [\#1169](https://github.com/traefik/traefik/issues/1169)
- Traefik not finding dynamically added services \(Docker Swarm Mode\) [\#1168](https://github.com/traefik/traefik/issues/1168)
- Traefik with Kubernetes backend - keep getting 401 on all GET requests to kube-apiserver [\#1166](https://github.com/traefik/traefik/issues/1166)
- Near line 15 \(last key parsed 'backends.backend-monitor-viz.servers'\): Key 'backends.backend-monitor-viz.servers.server-monitor\_viz-1' has already been defined. [\#1154](https://github.com/traefik/traefik/issues/1154)
- How to reuse SSL certificates automatically fetched from Let´s encrypt? [\#1152](https://github.com/traefik/traefik/issues/1152)
- Dynamically ban ip when backend repeatedly returns specified status code. \( 403 \) [\#1136](https://github.com/traefik/traefik/issues/1136)
- Always get 404 accessing my nginx backend service [\#1112](https://github.com/traefik/traefik/issues/1112)
- Incomplete Docu [\#1091](https://github.com/traefik/traefik/issues/1091)
- LoadCertificateForDomains: runtime error: invalid memory address [\#1069](https://github.com/traefik/traefik/issues/1069)
- Traefik creating backends & mappings for ingress annotated with ingress.class: nginx [\#1058](https://github.com/traefik/traefik/issues/1058)
- ACME file format description [\#1012](https://github.com/traefik/traefik/issues/1012)
- SwarmMode - Not routing on worker node [\#838](https://github.com/traefik/traefik/issues/838)
- Migrate k8s to kubernetes/client-go  [\#678](https://github.com/traefik/traefik/issues/678)
- Support for sticky session with kubernetes ingress as backend [\#674](https://github.com/traefik/traefik/issues/674)

**Merged pull requests:**

- Revert "Ensure that we don't add balances with no health check runs … [\#1198](https://github.com/traefik/traefik/pull/1198) ([jangie](https://github.com/jangie))
- Small fixes and improvements [\#1173](https://github.com/traefik/traefik/pull/1173) ([SantoDE](https://github.com/SantoDE))
- Fix docker issues with global and dead tasks [\#1167](https://github.com/traefik/traefik/pull/1167) ([christopherobin](https://github.com/christopherobin))
- Better ECS error checking [\#1143](https://github.com/traefik/traefik/pull/1143) ([lpetre](https://github.com/lpetre))
- Fix stats race condition [\#1141](https://github.com/traefik/traefik/pull/1141) ([emilevauge](https://github.com/emilevauge))
- ECS: Docs - info about cred. resolution and required access policies [\#1137](https://github.com/traefik/traefik/pull/1137) ([rickard-von-essen](https://github.com/rickard-von-essen))
- Healthcheck tests and doc [\#1132](https://github.com/traefik/traefik/pull/1132) ([Juliens](https://github.com/Juliens))

## [v1.2.0-rc1](https://github.com/traefik/traefik/tree/v1.2.0-rc1) (2017-02-06)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.1.2...v1.2.0-rc1)

**Implemented enhancements:**

- Add FreeBSD and OpenBSD to release builds [\#923](https://github.com/traefik/traefik/issues/923)
- Write authenticated user to header key [\#802](https://github.com/traefik/traefik/issues/802)
- Question: Wildcard Host for Kubernetes Ingress [\#792](https://github.com/traefik/traefik/issues/792)
- First commit prometheus middleware. [\#1022](https://github.com/traefik/traefik/pull/1022) ([enxebre](https://github.com/enxebre))
- Use deployment primitives from travis [\#843](https://github.com/traefik/traefik/pull/843) ([guilhem](https://github.com/guilhem))

**Fixed bugs:**

- Increase Docker API version to work with Windows Containers [\#1094](https://github.com/traefik/traefik/issues/1094)

**Closed issues:**

- How could I know whether forwarding path is correctly set? [\#1111](https://github.com/traefik/traefik/issues/1111)
- ACME + Docker-compose labels [\#1099](https://github.com/traefik/traefik/issues/1099)
- Loadbalance between 2 containers in Docker Swarm Mode [\#1095](https://github.com/traefik/traefik/issues/1095)
- Add DNS01 letsencrypt challenge support through AWS. [\#1093](https://github.com/traefik/traefik/issues/1093)
- New Release Cut [\#1092](https://github.com/traefik/traefik/issues/1092)
- Marathon integration changed default backend server port from task-level to application-level [\#1072](https://github.com/traefik/traefik/issues/1072)
- websockets not working when compress = true in toml config. [\#1059](https://github.com/traefik/traefik/issues/1059)
- Proxying 403 http status into the application [\#1044](https://github.com/traefik/traefik/issues/1044)
- Normalize auto generated frontend-rule \(docker\) [\#1043](https://github.com/traefik/traefik/issues/1043)
- Traefik with Consul catalog backend + Registrator [\#1039](https://github.com/traefik/traefik/issues/1039)
- \[Configuration help\] Can't connect to docker containers under a domain path [\#1032](https://github.com/traefik/traefik/issues/1032)
- Kubernetes and etcd backend : `storeconfig` fails. [\#1031](https://github.com/traefik/traefik/issues/1031)
- kubernetes: Undefined backend 'X/' for frontend X/" [\#1026](https://github.com/traefik/traefik/issues/1026)
- TLS handshake error [\#1025](https://github.com/traefik/traefik/issues/1025)
- Traefik failing on POST request [\#1008](https://github.com/traefik/traefik/issues/1008)
- how config traffic.toml http 80 without basic auth, traefik WebUI 8080 with basic auth [\#1001](https://github.com/traefik/traefik/issues/1001)
- Docs 404 [\#995](https://github.com/traefik/traefik/issues/995)
- Disable acme for non https endpoints [\#989](https://github.com/traefik/traefik/issues/989)
- Add parameter to configure TLS entrypoints with ca-bundle file [\#984](https://github.com/traefik/traefik/issues/984)
- docker multiple networks routing [\#970](https://github.com/traefik/traefik/issues/970)
- don't add Docker containers not on the same network as traefik [\#959](https://github.com/traefik/traefik/issues/959)
- Multiple frontend routes [\#957](https://github.com/traefik/traefik/issues/957)
- SNI based routing without TLS offloading [\#933](https://github.com/traefik/traefik/issues/933)
- NEO4J + traefik proxy Issues  [\#907](https://github.com/traefik/traefik/issues/907)
- ACME OnDemand ignores entrypoint certificate [\#672](https://github.com/traefik/traefik/issues/672)
- Ability to use self-signed certificates for local development [\#399](https://github.com/traefik/traefik/issues/399)

**Merged pull requests:**

- Fix checkout initial before calling rmpr [\#1124](https://github.com/traefik/traefik/pull/1124) ([emilevauge](https://github.com/emilevauge))
- Feature rancher integration [\#1120](https://github.com/traefik/traefik/pull/1120) ([SantoDE](https://github.com/SantoDE))
- Fix glide go units [\#1119](https://github.com/traefik/traefik/pull/1119) ([emilevauge](https://github.com/emilevauge))
- Carry \#818 —  Add systemd watchdog feature [\#1116](https://github.com/traefik/traefik/pull/1116) ([vdemeester](https://github.com/vdemeester))
- Skip file permission check on Windows [\#1115](https://github.com/traefik/traefik/pull/1115) ([StefanScherer](https://github.com/StefanScherer))
- Fix Docker API version for Windows [\#1113](https://github.com/traefik/traefik/pull/1113) ([StefanScherer](https://github.com/StefanScherer))
- Fix git rpr [\#1109](https://github.com/traefik/traefik/pull/1109) ([emilevauge](https://github.com/emilevauge))
- Fix docker version specifier [\#1108](https://github.com/traefik/traefik/pull/1108) ([timoreimann](https://github.com/timoreimann))
- Merge v1.1.2 master [\#1105](https://github.com/traefik/traefik/pull/1105) ([emilevauge](https://github.com/emilevauge))
- add sh before script in deploy... [\#1103](https://github.com/traefik/traefik/pull/1103) ([emilevauge](https://github.com/emilevauge))
- \[doc\] typo fixes for kubernetes user guide [\#1102](https://github.com/traefik/traefik/pull/1102) ([bamarni](https://github.com/bamarni))
- add skip\_cleanup in deploy [\#1101](https://github.com/traefik/traefik/pull/1101) ([emilevauge](https://github.com/emilevauge))
- Fix k8s example UI port. [\#1098](https://github.com/traefik/traefik/pull/1098) ([ddunkin](https://github.com/ddunkin))
- Fix marathon provider [\#1090](https://github.com/traefik/traefik/pull/1090) ([diegooliveira](https://github.com/diegooliveira))
- Add an ECS provider [\#1088](https://github.com/traefik/traefik/pull/1088) ([lpetre](https://github.com/lpetre))
- Update comment to reflect the code [\#1087](https://github.com/traefik/traefik/pull/1087) ([np](https://github.com/np))
- update NYTimes/gziphandler fixes \#1059 [\#1084](https://github.com/traefik/traefik/pull/1084) ([JamesKyburz](https://github.com/JamesKyburz))
- Ensure that we don't add balances with no health check runs if there is a health check defined on it [\#1080](https://github.com/traefik/traefik/pull/1080) ([jangie](https://github.com/jangie))
- Add FreeBSD & OpenBSD to crossbinary [\#1078](https://github.com/traefik/traefik/pull/1078) ([geoffgarside](https://github.com/geoffgarside))
- Fix metrics for multiple entry points [\#1071](https://github.com/traefik/traefik/pull/1071) ([matevzmihalic](https://github.com/matevzmihalic))
- Allow setting load balancer method and sticky using service annotations [\#1068](https://github.com/traefik/traefik/pull/1068) ([bakins](https://github.com/bakins))
- Fix travis script [\#1067](https://github.com/traefik/traefik/pull/1067) ([emilevauge](https://github.com/emilevauge))
- Add missing fmt verb specifier in k8s provider. [\#1066](https://github.com/traefik/traefik/pull/1066) ([timoreimann](https://github.com/timoreimann))
- Add git rpr command [\#1063](https://github.com/traefik/traefik/pull/1063) ([emilevauge](https://github.com/emilevauge))
- Fix k8s example [\#1062](https://github.com/traefik/traefik/pull/1062) ([emilevauge](https://github.com/emilevauge))
- Replace underscores to dash in autogenerated urls \(docker provider\) [\#1061](https://github.com/traefik/traefik/pull/1061) ([WTFKr0](https://github.com/WTFKr0))
- Don't run go test on .glide cache folder [\#1057](https://github.com/traefik/traefik/pull/1057) ([vdemeester](https://github.com/vdemeester))
- Allow setting circuitbreaker expression via Kubernetes annotation [\#1056](https://github.com/traefik/traefik/pull/1056) ([bakins](https://github.com/bakins))
- Improving instrumentation. [\#1042](https://github.com/traefik/traefik/pull/1042) ([enxebre](https://github.com/enxebre))
- Update user guide for upcoming `docker stack deploy`  [\#1041](https://github.com/traefik/traefik/pull/1041) ([twelvelabs](https://github.com/twelvelabs))
- Support sticky sessions under SWARM Mode. \#1024 [\#1033](https://github.com/traefik/traefik/pull/1033) ([foleymic](https://github.com/foleymic))
- Allow for wildcards in k8s ingress host, fixes \#792 [\#1029](https://github.com/traefik/traefik/pull/1029) ([sheerun](https://github.com/sheerun))
- Don't fetch ACME certificates for frontends using non-TLS entrypoints \(\#989\) [\#1023](https://github.com/traefik/traefik/pull/1023) ([syfonseq](https://github.com/syfonseq))
- Return Proper Non-ACME certificate - Fixes Issue 672 [\#1018](https://github.com/traefik/traefik/pull/1018) ([dtomcej](https://github.com/dtomcej))
- Fix docs build and add missing benchmarks page [\#1017](https://github.com/traefik/traefik/pull/1017) ([csabapalfi](https://github.com/csabapalfi))
- Set a NopCloser request body with retry middleware [\#1016](https://github.com/traefik/traefik/pull/1016) ([bamarni](https://github.com/bamarni))
- instruct to flatten dependencies with glide [\#1010](https://github.com/traefik/traefik/pull/1010) ([bamarni](https://github.com/bamarni))
- check permissions on acme.json during startup [\#1009](https://github.com/traefik/traefik/pull/1009) ([bamarni](https://github.com/bamarni))
- \[doc\] few tweaks on the basics page [\#1005](https://github.com/traefik/traefik/pull/1005) ([bamarni](https://github.com/bamarni))
- Import order as goimports does [\#1004](https://github.com/traefik/traefik/pull/1004) ([vdemeester](https://github.com/vdemeester))
- See the right go report badge [\#991](https://github.com/traefik/traefik/pull/991) ([guilhem](https://github.com/guilhem))
- Add multiple values for one rule to docs [\#978](https://github.com/traefik/traefik/pull/978) ([j0hnsmith](https://github.com/j0hnsmith))
- Add ACME/Let’s Encrypt integration tests [\#975](https://github.com/traefik/traefik/pull/975) ([trecloux](https://github.com/trecloux))
- deploy.sh: upload release source tarball [\#969](https://github.com/traefik/traefik/pull/969) ([Mic92](https://github.com/Mic92))
- toml zookeeper doc fix [\#948](https://github.com/traefik/traefik/pull/948) ([brdude](https://github.com/brdude))
- Add Rule AddPrefix [\#931](https://github.com/traefik/traefik/pull/931) ([Juliens](https://github.com/Juliens))
- Add bug command [\#921](https://github.com/traefik/traefik/pull/921) ([emilevauge](https://github.com/emilevauge))
- \(WIP\) feat: HealthCheck [\#918](https://github.com/traefik/traefik/pull/918) ([Juliens](https://github.com/Juliens))
- Add ability to set authenticated user in request header [\#889](https://github.com/traefik/traefik/pull/889) ([ViViDboarder](https://github.com/ViViDboarder))
- IP-per-task: [\#841](https://github.com/traefik/traefik/pull/841) ([diegooliveira](https://github.com/diegooliveira))

## [v1.1.2](https://github.com/traefik/traefik/tree/v1.1.2) (2016-12-15)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.1.1...v1.1.2)

**Fixed bugs:**

- Problem during HTTPS redirection [\#952](https://github.com/traefik/traefik/issues/952)
- nil pointer with kubernetes ingress [\#934](https://github.com/traefik/traefik/issues/934)
- ConsulCatalog and File not working [\#903](https://github.com/traefik/traefik/issues/903)
- Traefik can not start [\#902](https://github.com/traefik/traefik/issues/902)
- Cannot connect to Kubernetes server failed to decode watch event [\#532](https://github.com/traefik/traefik/issues/532)

**Closed issues:**

- Updating certificates with configuration file. [\#968](https://github.com/traefik/traefik/issues/968)
- Let's encrypt retrieving certificate from wrong IP [\#962](https://github.com/traefik/traefik/issues/962)
- let's encrypt and dashboard? [\#961](https://github.com/traefik/traefik/issues/961)
- Working HTTPS example for GKE? [\#960](https://github.com/traefik/traefik/issues/960)
- GKE design pattern [\#958](https://github.com/traefik/traefik/issues/958)
- Consul Catalog constraints does not seem to work [\#954](https://github.com/traefik/traefik/issues/954)
- Issue in building traefik from master [\#949](https://github.com/traefik/traefik/issues/949)
- Proxy http application to https doesn't seem to work correctly for all services [\#937](https://github.com/traefik/traefik/issues/937)
- Excessive requests to kubernetes apiserver [\#922](https://github.com/traefik/traefik/issues/922)
- I am getting a connection error while creating traefik with consul backend "dial tcp 127.0.0.1:8500: getsockopt: connection refused" [\#917](https://github.com/traefik/traefik/issues/917)
- SwarmMode - 1.13 RC2 - DNS RR - Individual IPs not retrieved [\#913](https://github.com/traefik/traefik/issues/913)
- Panic in kubernetes ingress \(traefik 1.1.0\) [\#910](https://github.com/traefik/traefik/issues/910)
- Kubernetes updating deployment image requires Ingress to be remade  [\#909](https://github.com/traefik/traefik/issues/909)
- \[ACME\] Too many currently pending authorizations [\#905](https://github.com/traefik/traefik/issues/905)
- WEB UI Authentication and Let's Encrypt : error 404 [\#754](https://github.com/traefik/traefik/issues/754)
- Traefik as ingress controller for SNI based routing in kubernetes [\#745](https://github.com/traefik/traefik/issues/745)
- Kubernetes Ingress backend: using self-signed certificates [\#486](https://github.com/traefik/traefik/issues/486)
- Kubernetes Ingress backend: can't find token and ca.crt [\#484](https://github.com/traefik/traefik/issues/484)

**Merged pull requests:**

- Fix duplicate acme certificates [\#972](https://github.com/traefik/traefik/pull/972) ([emilevauge](https://github.com/emilevauge))
- Fix leadership panic [\#956](https://github.com/traefik/traefik/pull/956) ([emilevauge](https://github.com/emilevauge))
- Fix redirect regex [\#947](https://github.com/traefik/traefik/pull/947) ([emilevauge](https://github.com/emilevauge))
- Add operation recover [\#944](https://github.com/traefik/traefik/pull/944) ([emilevauge](https://github.com/emilevauge))

## [v1.1.1](https://github.com/traefik/traefik/tree/v1.1.1) (2016-11-29)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.1.0...v1.1.1)

**Implemented enhancements:**

- Getting "Kubernetes connection error failed to decode watch event : unexpected EOF" every minute in Traefik log [\#732](https://github.com/traefik/traefik/issues/732)

**Fixed bugs:**

- 1.1.0 kubernetes panic: send on closed channel [\#877](https://github.com/traefik/traefik/issues/877)
- digest auth example is incorrect [\#869](https://github.com/traefik/traefik/issues/869)
- Marathon & Mesos providers' GroupsAsSubDomains option broken [\#867](https://github.com/traefik/traefik/issues/867)
- 404 responses when a new Marathon leader is elected [\#653](https://github.com/traefik/traefik/issues/653)

**Closed issues:**

- traefik:latest fails to auto-detect Docker containers [\#901](https://github.com/traefik/traefik/issues/901)
- Panic error on bare metal Kubernetes \(installed with Kubeadm\) [\#897](https://github.com/traefik/traefik/issues/897)
- api backend readOnly: what is the purpose of this setting [\#893](https://github.com/traefik/traefik/issues/893)
- file backend: using external file - doesn't work [\#892](https://github.com/traefik/traefik/issues/892)
- auth support for web backend [\#891](https://github.com/traefik/traefik/issues/891)
- Basic auth with docker labels [\#890](https://github.com/traefik/traefik/issues/890)
- file vs inline config [\#888](https://github.com/traefik/traefik/issues/888)
- combine Host and HostRegexp rules [\#882](https://github.com/traefik/traefik/issues/882)
- \[Question\] Traefik + Kubernetes + Let's Encrypt \(ssl not used\) [\#881](https://github.com/traefik/traefik/issues/881)
- Traefik security for dashboard [\#880](https://github.com/traefik/traefik/issues/880)
- Kubernetes Nginx Deployment Panic [\#879](https://github.com/traefik/traefik/issues/879)
- Kubernetes Example Address already in use [\#872](https://github.com/traefik/traefik/issues/872)
- ETCD Backend - frontend/backends missing [\#866](https://github.com/traefik/traefik/issues/866)
- \[Swarm mode\] Dashboard does not work on RC4 [\#864](https://github.com/traefik/traefik/issues/864)
- Docker v1.1.0 image does not exist [\#861](https://github.com/traefik/traefik/issues/861)
- ConsulService catalog do not support multiple rules [\#859](https://github.com/traefik/traefik/issues/859)
- Update official docker repo [\#858](https://github.com/traefik/traefik/issues/858)
- Still a memory leak with k8s - 1.1 RC4 [\#844](https://github.com/traefik/traefik/issues/844)

**Merged pull requests:**

- Fix Swarm panic [\#908](https://github.com/traefik/traefik/pull/908) ([emilevauge](https://github.com/emilevauge))
- Fix k8s panic [\#900](https://github.com/traefik/traefik/pull/900) ([emilevauge](https://github.com/emilevauge))
- Fix missing value for k8s watch request parameter [\#874](https://github.com/traefik/traefik/pull/874) ([codablock](https://github.com/codablock))
- Fix GroupsAsSubDomains option for Mesos and Marathon [\#868](https://github.com/traefik/traefik/pull/868) ([ryanleary](https://github.com/ryanleary))
- Normalize backend even if is user-defined [\#865](https://github.com/traefik/traefik/pull/865) ([WTFKr0](https://github.com/WTFKr0))
- consul/kv.tmpl: weight default value should be a int [\#826](https://github.com/traefik/traefik/pull/826) ([klausenbusk](https://github.com/klausenbusk))

## [v1.1.0](https://github.com/traefik/traefik/tree/v1.1.0) (2016-11-17)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.0.0...v1.1.0)

**Implemented enhancements:**

- Support healthcheck if present for docker [\#666](https://github.com/traefik/traefik/issues/666)
- Standard unit for traefik latency in access log [\#559](https://github.com/traefik/traefik/issues/559)
- \[CI\] wiredep marked as unmaintained [\#550](https://github.com/traefik/traefik/issues/550)
- Feature Request: Enable Health checks to containers. [\#540](https://github.com/traefik/traefik/issues/540)
- Feature Request: SSL Cipher Selection [\#535](https://github.com/traefik/traefik/issues/535)
- Error with -consulcatalog and missing load balance method on 1.0.0 [\#524](https://github.com/traefik/traefik/issues/524)
- Running Traefik with Docker 1.12 Swarm Mode [\#504](https://github.com/traefik/traefik/issues/504)
- Kubernetes provider: should allow the master url to be override [\#501](https://github.com/traefik/traefik/issues/501)
- \[FRONTEND\]\[LE\] Pre-generate SSL certificates for "Host:" rules [\#483](https://github.com/traefik/traefik/issues/483)
- Frontend Rule evolution [\#437](https://github.com/traefik/traefik/issues/437)
- Add a Changelog [\#388](https://github.com/traefik/traefik/issues/388)
- Add label matching for kubernetes ingests [\#363](https://github.com/traefik/traefik/issues/363)
- Acme in HA Traefik Scenario [\#348](https://github.com/traefik/traefik/issues/348)
- HTTP Basic Auth support [\#77](https://github.com/traefik/traefik/issues/77)
- Session affinity / stickiness / persistence [\#5](https://github.com/traefik/traefik/issues/5)

**Fixed bugs:**

- 1.1.0-rc4 dashboard UX not displaying [\#828](https://github.com/traefik/traefik/issues/828)
- Traefik stopped serving on upgrade to v1.1.0-rc3 [\#807](https://github.com/traefik/traefik/issues/807)
- cannot access webui/dashboard  [\#796](https://github.com/traefik/traefik/issues/796)
- Traefik cannot read constraints from KV [\#794](https://github.com/traefik/traefik/issues/794)
- HTTP2 - configuration [\#790](https://github.com/traefik/traefik/issues/790)
- Cannot provide multiple certificates using flag [\#757](https://github.com/traefik/traefik/issues/757)
- Allow multiple certificates on a single entrypoint when trying to use TLS? [\#747](https://github.com/traefik/traefik/issues/747)
- traefik \* Users: unsupported type: slice [\#743](https://github.com/traefik/traefik/issues/743)
- \[Docker swarm mode\] The traefik.docker.network seems to have no effect [\#719](https://github.com/traefik/traefik/issues/719)
- traefik hangs - stops handling requests [\#662](https://github.com/traefik/traefik/issues/662)
- Add long jobs in exponential backoff providers  [\#626](https://github.com/traefik/traefik/issues/626)
- Tip of tree crashes on invalid pointer on Marathon provider [\#624](https://github.com/traefik/traefik/issues/624)
- ACME: revoke certificate on agreement update [\#579](https://github.com/traefik/traefik/issues/579)
- WebUI: Providers tabs disappeared [\#577](https://github.com/traefik/traefik/issues/577)
- traefik version command contains incorrect information when building from master branch [\#569](https://github.com/traefik/traefik/issues/569)
- Case sensitive domain names breaks routing  [\#562](https://github.com/traefik/traefik/issues/562)
- Flag --etcd.endpoint default [\#508](https://github.com/traefik/traefik/issues/508)
- Conditional ACME on demand generation [\#505](https://github.com/traefik/traefik/issues/505)
- Important delay with streams \(Mozilla EventSource\) [\#503](https://github.com/traefik/traefik/issues/503)
- Traefik crashing [\#458](https://github.com/traefik/traefik/issues/458)
- traefik.toml constraints error: `Expected map but found 'string'.` [\#451](https://github.com/traefik/traefik/issues/451)
- Multiple path separators in the url path causing redirect [\#167](https://github.com/traefik/traefik/issues/167)

**Closed issues:**

- All path rules require paths to be lowercase [\#851](https://github.com/traefik/traefik/issues/851)
- The UI stops working after a time and have to restart the service. [\#840](https://github.com/traefik/traefik/issues/840)
- Incorrect Dashboard page returned [\#831](https://github.com/traefik/traefik/issues/831)
- LoadBalancing doesn't work in single node Swarm-mode [\#815](https://github.com/traefik/traefik/issues/815)
- cannot connect to docker daemon [\#813](https://github.com/traefik/traefik/issues/813)
- Let's encrypt configuration not working [\#805](https://github.com/traefik/traefik/issues/805)
- Multiple subdomains for Marathon backend. [\#785](https://github.com/traefik/traefik/issues/785)
- traefik-1.1.0-rc1: build error [\#781](https://github.com/traefik/traefik/issues/781)
- dependencies installation error [\#755](https://github.com/traefik/traefik/issues/755)
- k8s provider w/ acme? [\#752](https://github.com/traefik/traefik/issues/752)
- Swarm Docs - How to use a FQDN [\#744](https://github.com/traefik/traefik/issues/744)
- Documented ProvidersThrottleDuration value is invalid [\#741](https://github.com/traefik/traefik/issues/741)
- Sensible configuration for consulCatalog [\#737](https://github.com/traefik/traefik/issues/737)
- Traefik ignoring container listening in more than one TCP port [\#734](https://github.com/traefik/traefik/issues/734)
- Loadbalancing issues with traefik and Docker Swarm cluster [\#730](https://github.com/traefik/traefik/issues/730)
- issues with marathon app ids containing a dot [\#726](https://github.com/traefik/traefik/issues/726)
- Error when using HA acme in kubernetes with etcd [\#725](https://github.com/traefik/traefik/issues/725)
- \[Docker swarm mode\] No round robin when using service [\#718](https://github.com/traefik/traefik/issues/718)
- Dose it support docker swarm mode  [\#712](https://github.com/traefik/traefik/issues/712)
- Kubernetes - Undefined backend  [\#710](https://github.com/traefik/traefik/issues/710)
- How Routing traffic depending on path not domain in docker [\#706](https://github.com/traefik/traefik/issues/706)
- Constraints on Consul Catalogue not working as expected [\#703](https://github.com/traefik/traefik/issues/703)
- Global InsecureSkipVerify does not work [\#700](https://github.com/traefik/traefik/issues/700)
- Traefik crashes when using Consul catalog [\#699](https://github.com/traefik/traefik/issues/699)
- \[documentation/feature\] Consul/etcd support atomic multiple key changes now [\#698](https://github.com/traefik/traefik/issues/698)
- How to configure which network to use when starting traefik binary? [\#694](https://github.com/traefik/traefik/issues/694)
- How to get multiple host headers working for docker labels? [\#692](https://github.com/traefik/traefik/issues/692)
- Requests with URL-encoded characters are not forwarded correctly [\#684](https://github.com/traefik/traefik/issues/684)
- File Watcher for rules does not work  [\#683](https://github.com/traefik/traefik/issues/683)
- Issue with global InsecureSkipVerify = true and self signed certificates [\#667](https://github.com/traefik/traefik/issues/667)
- Docker exposedbydefault = false didn't work [\#663](https://github.com/traefik/traefik/issues/663)
- swarm documentation needs update [\#656](https://github.com/traefik/traefik/issues/656)
- \[ACME\] Auto SAN Detection [\#655](https://github.com/traefik/traefik/issues/655)
- Fronting a domain with DNS A-record round-robin & ACME [\#654](https://github.com/traefik/traefik/issues/654)
- Overriding toml configuration with environment variables [\#650](https://github.com/traefik/traefik/issues/650)
- marathon provider exposedByDefault = false [\#647](https://github.com/traefik/traefik/issues/647)
- Add status URL for service up checks [\#642](https://github.com/traefik/traefik/issues/642)
- acme's storage file, containing private key, is word readable [\#638](https://github.com/traefik/traefik/issues/638)
- wildcard domain with exclusions [\#633](https://github.com/traefik/traefik/issues/633)
- Enable evenly distribution among backend [\#631](https://github.com/traefik/traefik/issues/631)
- Traefik sporadically failing when proxying requests [\#615](https://github.com/traefik/traefik/issues/615)
- TCP Proxy [\#608](https://github.com/traefik/traefik/issues/608)
- How to use in Windows? [\#605](https://github.com/traefik/traefik/issues/605)
- `ClientCAFiles` ignored [\#604](https://github.com/traefik/traefik/issues/604)
- Let`s Encrypt enable in etcd [\#600](https://github.com/traefik/traefik/issues/600)
- Support HTTP Basic Auth [\#599](https://github.com/traefik/traefik/issues/599)
- Consul KV seem broken [\#587](https://github.com/traefik/traefik/issues/587)
- HTTPS entryPoint not working [\#574](https://github.com/traefik/traefik/issues/574)
- Traefik stuck when used as frontend for a streaming API [\#560](https://github.com/traefik/traefik/issues/560)
- Exclude some frontends in consul catalog [\#555](https://github.com/traefik/traefik/issues/555)
- Update docs with new Mesos provider [\#548](https://github.com/traefik/traefik/issues/548)
- Can I use Traefik without a domain name? [\#539](https://github.com/traefik/traefik/issues/539)
- docker run syntax in swarm example has changed [\#528](https://github.com/traefik/traefik/issues/528)
- Priorities in 1.0.0 not behaving [\#506](https://github.com/traefik/traefik/issues/506)
- Route by path [\#500](https://github.com/traefik/traefik/issues/500)
- Secure WebSockets [\#467](https://github.com/traefik/traefik/issues/467)
- Container IP Lost [\#375](https://github.com/traefik/traefik/issues/375)
- Multiple routes support with Docker or Marathon labels [\#118](https://github.com/traefik/traefik/issues/118)

**Merged pull requests:**

- Fix path case sensitive v1.1 [\#855](https://github.com/traefik/traefik/pull/855) ([emilevauge](https://github.com/emilevauge))
- Fix golint in v1.1 [\#849](https://github.com/traefik/traefik/pull/849) ([emilevauge](https://github.com/emilevauge))
- Fix Kubernetes watch leak [\#845](https://github.com/traefik/traefik/pull/845) ([emilevauge](https://github.com/emilevauge))
- Pass Version, Codename and Date to crosscompiled [\#842](https://github.com/traefik/traefik/pull/842) ([guilhem](https://github.com/guilhem))
- Add Nvd3 Dependency to fix UI / Dashboard [\#829](https://github.com/traefik/traefik/pull/829) ([SantoDE](https://github.com/SantoDE))
- Fix mkdoc theme [\#823](https://github.com/traefik/traefik/pull/823) ([emilevauge](https://github.com/emilevauge))
- Prepare release v1.1.0 rc4 [\#822](https://github.com/traefik/traefik/pull/822) ([emilevauge](https://github.com/emilevauge))
- Check that we serve HTTP/2 [\#820](https://github.com/traefik/traefik/pull/820) ([trecloux](https://github.com/trecloux))
- Fix multiple issues [\#814](https://github.com/traefik/traefik/pull/814) ([emilevauge](https://github.com/emilevauge))
- Fix ACME renew & add version check [\#783](https://github.com/traefik/traefik/pull/783) ([emilevauge](https://github.com/emilevauge))
- Use first port by default [\#782](https://github.com/traefik/traefik/pull/782) ([guilhem](https://github.com/guilhem))
- Prepare release v1.1.0-rc3 [\#779](https://github.com/traefik/traefik/pull/779) ([emilevauge](https://github.com/emilevauge))
- Fix ResponseRecorder Flush [\#776](https://github.com/traefik/traefik/pull/776) ([emilevauge](https://github.com/emilevauge))
- Use sdnotify for systemd [\#768](https://github.com/traefik/traefik/pull/768) ([guilhem](https://github.com/guilhem))
- Fix providers throttle duration doc [\#760](https://github.com/traefik/traefik/pull/760) ([emilevauge](https://github.com/emilevauge))
- Fix mapstructure issue with anonymous slice [\#759](https://github.com/traefik/traefik/pull/759) ([emilevauge](https://github.com/emilevauge))
- Fix multiple certificates using flag [\#758](https://github.com/traefik/traefik/pull/758) ([emilevauge](https://github.com/emilevauge))
- Really fix deploy ghr... [\#748](https://github.com/traefik/traefik/pull/748) ([emilevauge](https://github.com/emilevauge))
- Fixes deploy ghr [\#742](https://github.com/traefik/traefik/pull/742) ([emilevauge](https://github.com/emilevauge))
- prepare v1.1.0-rc2 [\#740](https://github.com/traefik/traefik/pull/740) ([emilevauge](https://github.com/emilevauge))
- Fix case sensitive host [\#733](https://github.com/traefik/traefik/pull/733) ([emilevauge](https://github.com/emilevauge))
- Update Kubernetes examples [\#731](https://github.com/traefik/traefik/pull/731) ([Starefossen](https://github.com/Starefossen))
- fIx marathon template with dots in ID [\#728](https://github.com/traefik/traefik/pull/728) ([emilevauge](https://github.com/emilevauge))
- Fix networkMap construction in ListServices [\#724](https://github.com/traefik/traefik/pull/724) ([vincentlepot](https://github.com/vincentlepot))
- Add basic compatibility with marathon-lb [\#720](https://github.com/traefik/traefik/pull/720) ([guilhem](https://github.com/guilhem))
- Add Ed's video at ContainerCamp [\#717](https://github.com/traefik/traefik/pull/717) ([emilevauge](https://github.com/emilevauge))
- Add documentation for Træfik on docker swarm mode [\#715](https://github.com/traefik/traefik/pull/715) ([vdemeester](https://github.com/vdemeester))
- Remove duplicated link to Kubernetes.io in README.md [\#713](https://github.com/traefik/traefik/pull/713) ([oscerd](https://github.com/oscerd))
- Show current version in web UI [\#709](https://github.com/traefik/traefik/pull/709) ([vhf](https://github.com/vhf))
- Add support for docker healthcheck 👼 [\#708](https://github.com/traefik/traefik/pull/708) ([vdemeester](https://github.com/vdemeester))
- Fix syntax in Swarm example. Resolves \#528 [\#707](https://github.com/traefik/traefik/pull/707) ([billglover](https://github.com/billglover))
- Add HTTP compression [\#702](https://github.com/traefik/traefik/pull/702) ([tuier](https://github.com/tuier))
- Carry PR 446 - Add sticky session support \(round two!\) [\#701](https://github.com/traefik/traefik/pull/701) ([emilevauge](https://github.com/emilevauge))
- Remove unused endpoint when using constraints with Marathon provider [\#697](https://github.com/traefik/traefik/pull/697) ([tuier](https://github.com/tuier))
- Replace imagelayers.io with microbadger [\#696](https://github.com/traefik/traefik/pull/696) ([solidnerd](https://github.com/solidnerd))
- Selectable TLS Versions [\#690](https://github.com/traefik/traefik/pull/690) ([dtomcej](https://github.com/dtomcej))
- Carry pr 439 [\#689](https://github.com/traefik/traefik/pull/689) ([emilevauge](https://github.com/emilevauge))
- Disable gorilla/mux URL cleaning to prevent sending redirect [\#688](https://github.com/traefik/traefik/pull/688) ([ydubreuil](https://github.com/ydubreuil))
- Some fixes [\#687](https://github.com/traefik/traefik/pull/687) ([emilevauge](https://github.com/emilevauge))
- feat\(constraints\): Supports constraints for Marathon provider [\#686](https://github.com/traefik/traefik/pull/686) ([tuier](https://github.com/tuier))
- Update docs to improve contribution setup [\#685](https://github.com/traefik/traefik/pull/685) ([dtomcej](https://github.com/dtomcej))
- Add basic auth support for web backend [\#677](https://github.com/traefik/traefik/pull/677) ([SantoDE](https://github.com/SantoDE))
- Document accepted values for logLevel. [\#676](https://github.com/traefik/traefik/pull/676) ([jimmycuadra](https://github.com/jimmycuadra))
- If Marathon doesn't have healthcheck, assume it's ok [\#665](https://github.com/traefik/traefik/pull/665) ([gomes](https://github.com/gomes))
- ACME: renew certificates 30 days before expiry [\#660](https://github.com/traefik/traefik/pull/660) ([JayH5](https://github.com/JayH5))
- Update broken link and add a comment to sample config file  [\#658](https://github.com/traefik/traefik/pull/658) ([Yggdrasil](https://github.com/Yggdrasil))
- Add possibility to use BindPort IPAddress 👼 [\#657](https://github.com/traefik/traefik/pull/657) ([vdemeester](https://github.com/vdemeester))
- Update marathon [\#648](https://github.com/traefik/traefik/pull/648) ([emilevauge](https://github.com/emilevauge))
- Add backend features to docker [\#646](https://github.com/traefik/traefik/pull/646) ([jangie](https://github.com/jangie))
- enable consul catalog to use maxconn [\#645](https://github.com/traefik/traefik/pull/645) ([jangie](https://github.com/jangie))
- Adopt the Code Of Conduct from http://contributor-covenant.org [\#641](https://github.com/traefik/traefik/pull/641) ([errm](https://github.com/errm))
- Use secure mode 600 instead of 644 for acme.json [\#639](https://github.com/traefik/traefik/pull/639) ([discordianfish](https://github.com/discordianfish))
- docker clarification, fix dead urls, misc typos [\#637](https://github.com/traefik/traefik/pull/637) ([djalal](https://github.com/djalal))
- add PING handler to dashboard API [\#630](https://github.com/traefik/traefik/pull/630) ([jangie](https://github.com/jangie))
- Migrate to JobBackOff [\#628](https://github.com/traefik/traefik/pull/628) ([emilevauge](https://github.com/emilevauge))
- Add long job exponential backoff [\#627](https://github.com/traefik/traefik/pull/627) ([emilevauge](https://github.com/emilevauge))
- HA acme support [\#625](https://github.com/traefik/traefik/pull/625) ([emilevauge](https://github.com/emilevauge))
- Bump go v1.7 [\#620](https://github.com/traefik/traefik/pull/620) ([emilevauge](https://github.com/emilevauge))
- Make duration logging consistent [\#619](https://github.com/traefik/traefik/pull/619) ([jangie](https://github.com/jangie))
- fix for nil clientTLS causing issue [\#617](https://github.com/traefik/traefik/pull/617) ([jangie](https://github.com/jangie))
- Add ability for marathon provider to set maxconn values, loadbalancer algorithm, and circuit breaker expression [\#616](https://github.com/traefik/traefik/pull/616) ([jangie](https://github.com/jangie))
- Make systemd unit installable [\#613](https://github.com/traefik/traefik/pull/613) ([keis](https://github.com/keis))
- Merge v1.0.2 master [\#610](https://github.com/traefik/traefik/pull/610) ([emilevauge](https://github.com/emilevauge))
- update staert and flaeg [\#609](https://github.com/traefik/traefik/pull/609) ([cocap10](https://github.com/cocap10))
- \#504 Initial support for Docker 1.12 Swarm Mode [\#602](https://github.com/traefik/traefik/pull/602) ([diegofernandes](https://github.com/diegofernandes))
- Add Host cert ACME generation [\#601](https://github.com/traefik/traefik/pull/601) ([emilevauge](https://github.com/emilevauge))
- Fixed binary script so traefik version command doesn't just print default values [\#598](https://github.com/traefik/traefik/pull/598) ([keiths-osc](https://github.com/keiths-osc))
- Name servers after their pods [\#596](https://github.com/traefik/traefik/pull/596) ([errm](https://github.com/errm))
- Fix Consul prefix [\#589](https://github.com/traefik/traefik/pull/589) ([jippi](https://github.com/jippi))
- Prioritize kubernetes routes by path length [\#588](https://github.com/traefik/traefik/pull/588) ([philk](https://github.com/philk))
- beautify help [\#580](https://github.com/traefik/traefik/pull/580) ([cocap10](https://github.com/cocap10))
- Upgrade directives name since we use angular-ui-bootstrap [\#578](https://github.com/traefik/traefik/pull/578) ([micaelmbagira](https://github.com/micaelmbagira))
- Fix basic docs for configuration of multiple rules [\#576](https://github.com/traefik/traefik/pull/576) ([ajaegle](https://github.com/ajaegle))
- Fix k8s watch [\#573](https://github.com/traefik/traefik/pull/573) ([errm](https://github.com/errm))
- Add requirements.txt for netlify [\#567](https://github.com/traefik/traefik/pull/567) ([emilevauge](https://github.com/emilevauge))
- Merge v1.0.1 master [\#565](https://github.com/traefik/traefik/pull/565) ([emilevauge](https://github.com/emilevauge))
-  Move webui to FountainJS with Webpack [\#558](https://github.com/traefik/traefik/pull/558) ([micaelmbagira](https://github.com/micaelmbagira))
- Add global InsecureSkipVerify option to disable certificate checking [\#557](https://github.com/traefik/traefik/pull/557) ([stuart-c](https://github.com/stuart-c))
- Move version.go in its own package… [\#553](https://github.com/traefik/traefik/pull/553) ([vdemeester](https://github.com/vdemeester))
- Upgrade libkermit and dependencies [\#552](https://github.com/traefik/traefik/pull/552) ([vdemeester](https://github.com/vdemeester))
- Add command storeconfig [\#551](https://github.com/traefik/traefik/pull/551) ([cocap10](https://github.com/cocap10))
- Add basic/digest auth [\#547](https://github.com/traefik/traefik/pull/547) ([emilevauge](https://github.com/emilevauge))
- Bump node to 6 for webui [\#546](https://github.com/traefik/traefik/pull/546) ([vdemeester](https://github.com/vdemeester))
- Bump golang to 1.6.3 [\#545](https://github.com/traefik/traefik/pull/545) ([vdemeester](https://github.com/vdemeester))
- Fix typos [\#538](https://github.com/traefik/traefik/pull/538) ([jimt](https://github.com/jimt))
- Kubernetes user-guide [\#519](https://github.com/traefik/traefik/pull/519) ([errm](https://github.com/errm))
- Implement Kubernetes Selectors, minor kube endpoint fix [\#516](https://github.com/traefik/traefik/pull/516) ([pnegahdar](https://github.com/pnegahdar))
- Carry \#358 : Option to disable expose of all docker containers [\#514](https://github.com/traefik/traefik/pull/514) ([vdemeester](https://github.com/vdemeester))
- Remove traefik.frontend.value support in docker… [\#510](https://github.com/traefik/traefik/pull/510) ([vdemeester](https://github.com/vdemeester))
- Use KvStores as global config sources [\#481](https://github.com/traefik/traefik/pull/481) ([cocap10](https://github.com/cocap10))
- Add endpoint option to authenticate by client tls cert. [\#461](https://github.com/traefik/traefik/pull/461) ([andersbetner](https://github.com/andersbetner))
- add mesos provider inspired by mesos-dns & marathon provider [\#353](https://github.com/traefik/traefik/pull/353) ([skydjol](https://github.com/skydjol))

## [v1.1.0-rc4](https://github.com/traefik/traefik/tree/v1.1.0-rc4) (2016-11-10)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.1.0-rc3...v1.1.0-rc4)

**Implemented enhancements:**

- Feature Request: Enable Health checks to containers. [\#540](https://github.com/traefik/traefik/issues/540)

**Fixed bugs:**

- Traefik stopped serving on upgrade to v1.1.0-rc3 [\#807](https://github.com/traefik/traefik/issues/807)
- Traefik cannot read constraints from KV [\#794](https://github.com/traefik/traefik/issues/794)
- HTTP2 - configuration [\#790](https://github.com/traefik/traefik/issues/790)
- Allow multiple certificates on a single entrypoint when trying to use TLS? [\#747](https://github.com/traefik/traefik/issues/747)

**Closed issues:**

- LoadBalancing doesn't work in single node Swarm-mode [\#815](https://github.com/traefik/traefik/issues/815)
- cannot connect to docker daemon [\#813](https://github.com/traefik/traefik/issues/813)
- Let's encrypt configuration not working [\#805](https://github.com/traefik/traefik/issues/805)
- Question: Wildcard Host for Kubernetes Ingress [\#792](https://github.com/traefik/traefik/issues/792)
- Multiple subdomains for Marathon backend. [\#785](https://github.com/traefik/traefik/issues/785)
- traefik-1.1.0-rc1: build error [\#781](https://github.com/traefik/traefik/issues/781)
- Multiple routes support with Docker or Marathon labels [\#118](https://github.com/traefik/traefik/issues/118)

**Merged pull requests:**

- Prepare release v1.1.0 rc4 [\#822](https://github.com/traefik/traefik/pull/822) ([emilevauge](https://github.com/emilevauge))
- Fix multiple issues [\#814](https://github.com/traefik/traefik/pull/814) ([emilevauge](https://github.com/emilevauge))
- Fix ACME renew & add version check [\#783](https://github.com/traefik/traefik/pull/783) ([emilevauge](https://github.com/emilevauge))
- Use first port by default [\#782](https://github.com/traefik/traefik/pull/782) ([guilhem](https://github.com/guilhem))

## [v1.1.0-rc3](https://github.com/traefik/traefik/tree/v1.1.0-rc3) (2016-10-26)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.1.0-rc2...v1.1.0-rc3)

**Fixed bugs:**

- Cannot provide multiple certificates using flag [\#757](https://github.com/traefik/traefik/issues/757)
- traefik \* Users: unsupported type: slice [\#743](https://github.com/traefik/traefik/issues/743)
- \[Docker swarm mode\] The traefik.docker.network seems to have no effect [\#719](https://github.com/traefik/traefik/issues/719)
- Case sensitive domain names breaks routing  [\#562](https://github.com/traefik/traefik/issues/562)

**Closed issues:**

- dependencies installation error [\#755](https://github.com/traefik/traefik/issues/755)
- k8s provider w/ acme? [\#752](https://github.com/traefik/traefik/issues/752)
- Documented ProvidersThrottleDuration value is invalid [\#741](https://github.com/traefik/traefik/issues/741)
- Loadbalancing issues with traefik and Docker Swarm cluster [\#730](https://github.com/traefik/traefik/issues/730)
- issues with marathon app ids containing a dot [\#726](https://github.com/traefik/traefik/issues/726)
- How Routing traffic depending on path not domain in docker [\#706](https://github.com/traefik/traefik/issues/706)
- Traefik crashes when using Consul catalog [\#699](https://github.com/traefik/traefik/issues/699)
- File Watcher for rules does not work  [\#683](https://github.com/traefik/traefik/issues/683)

**Merged pull requests:**

- Fix ResponseRecorder Flush [\#776](https://github.com/traefik/traefik/pull/776) ([emilevauge](https://github.com/emilevauge))
- Use sdnotify for systemd [\#768](https://github.com/traefik/traefik/pull/768) ([guilhem](https://github.com/guilhem))
- Fix providers throttle duration doc [\#760](https://github.com/traefik/traefik/pull/760) ([emilevauge](https://github.com/emilevauge))
- Fix mapstructure issue with anonymous slice [\#759](https://github.com/traefik/traefik/pull/759) ([emilevauge](https://github.com/emilevauge))
- Fix multiple certificates using flag [\#758](https://github.com/traefik/traefik/pull/758) ([emilevauge](https://github.com/emilevauge))
- Really fix deploy ghr... [\#748](https://github.com/traefik/traefik/pull/748) ([emilevauge](https://github.com/emilevauge))

## [v1.1.0-rc2](https://github.com/traefik/traefik/tree/v1.1.0-rc2) (2016-10-17)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.1.0-rc1...v1.1.0-rc2)

**Implemented enhancements:**

- Support healthcheck if present for docker [\#666](https://github.com/traefik/traefik/issues/666)

**Closed issues:**

- Sensible configuration for consulCatalog [\#737](https://github.com/traefik/traefik/issues/737)
- Traefik ignoring container listening in more than one TCP port [\#734](https://github.com/traefik/traefik/issues/734)
- Error when using HA acme in kubernetes with etcd [\#725](https://github.com/traefik/traefik/issues/725)
- \[Docker swarm mode\] No round robin when using service [\#718](https://github.com/traefik/traefik/issues/718)
- Dose it support docker swarm mode  [\#712](https://github.com/traefik/traefik/issues/712)
- Kubernetes - Undefined backend  [\#710](https://github.com/traefik/traefik/issues/710)
- Constraints on Consul Catalogue not working as expected [\#703](https://github.com/traefik/traefik/issues/703)
- docker run syntax in swarm example has changed [\#528](https://github.com/traefik/traefik/issues/528)
- Secure WebSockets [\#467](https://github.com/traefik/traefik/issues/467)

**Merged pull requests:**

- Fix case sensitive host [\#733](https://github.com/traefik/traefik/pull/733) ([emilevauge](https://github.com/emilevauge))
- Update Kubernetes examples [\#731](https://github.com/traefik/traefik/pull/731) ([Starefossen](https://github.com/Starefossen))
- fIx marathon template with dots in ID [\#728](https://github.com/traefik/traefik/pull/728) ([emilevauge](https://github.com/emilevauge))
- Fix networkMap construction in ListServices [\#724](https://github.com/traefik/traefik/pull/724) ([vincentlepot](https://github.com/vincentlepot))
- Add basic compatibility with marathon-lb [\#720](https://github.com/traefik/traefik/pull/720) ([guilhem](https://github.com/guilhem))
- Add Ed's video at ContainerCamp [\#717](https://github.com/traefik/traefik/pull/717) ([emilevauge](https://github.com/emilevauge))
- Add documentation for Træfik on docker swarm mode [\#715](https://github.com/traefik/traefik/pull/715) ([vdemeester](https://github.com/vdemeester))
- Remove duplicated link to Kubernetes.io in README.md [\#713](https://github.com/traefik/traefik/pull/713) ([oscerd](https://github.com/oscerd))
- Show current version in web UI [\#709](https://github.com/traefik/traefik/pull/709) ([vhf](https://github.com/vhf))
- Add support for docker healthcheck 👼 [\#708](https://github.com/traefik/traefik/pull/708) ([vdemeester](https://github.com/vdemeester))
- Fix syntax in Swarm example. Resolves \#528 [\#707](https://github.com/traefik/traefik/pull/707) ([billglover](https://github.com/billglover))

## [v1.1.0-rc1](https://github.com/traefik/traefik/tree/v1.1.0-rc1) (2016-09-30)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.0.0...v1.1.0-rc1)

**Implemented enhancements:**

- Feature Request: SSL Cipher Selection [\#535](https://github.com/traefik/traefik/issues/535)
- Error with -consulcatalog and missing load balance method on 1.0.0 [\#524](https://github.com/traefik/traefik/issues/524)
- Running Traefik with Docker 1.12 Swarm Mode [\#504](https://github.com/traefik/traefik/issues/504)
- Kubernetes provider: should allow the master url to be override [\#501](https://github.com/traefik/traefik/issues/501)
- \[FRONTEND\]\[LE\] Pre-generate SSL certificates for "Host:" rules [\#483](https://github.com/traefik/traefik/issues/483)
- Frontend Rule evolution [\#437](https://github.com/traefik/traefik/issues/437)
- Add a Changelog [\#388](https://github.com/traefik/traefik/issues/388)
- Add label matching for kubernetes ingests [\#363](https://github.com/traefik/traefik/issues/363)
- Acme in HA Traefik Scenario [\#348](https://github.com/traefik/traefik/issues/348)
- HTTP Basic Auth support [\#77](https://github.com/traefik/traefik/issues/77)
- Session affinity / stickiness / persistence [\#5](https://github.com/traefik/traefik/issues/5)
- Kubernetes provider: traefik.frontend.rule.type logging [\#668](https://github.com/traefik/traefik/pull/668) ([yvespp](https://github.com/yvespp))

**Fixed bugs:**

- traefik hangs - stops handling requests [\#662](https://github.com/traefik/traefik/issues/662)
- Add long jobs in exponential backoff providers  [\#626](https://github.com/traefik/traefik/issues/626)
- Tip of tree crashes on invalid pointer on Marathon provider [\#624](https://github.com/traefik/traefik/issues/624)
- ACME: revoke certificate on agreement update [\#579](https://github.com/traefik/traefik/issues/579)
- WebUI: Providers tabs disappeared [\#577](https://github.com/traefik/traefik/issues/577)
- traefik version command contains incorrect information when building from master branch [\#569](https://github.com/traefik/traefik/issues/569)
- Flag --etcd.endpoint default [\#508](https://github.com/traefik/traefik/issues/508)
- Conditional ACME on demand generation [\#505](https://github.com/traefik/traefik/issues/505)
- Important delay with streams \(Mozilla EventSource\) [\#503](https://github.com/traefik/traefik/issues/503)
- Traefik crashing [\#458](https://github.com/traefik/traefik/issues/458)
- traefik.toml constraints error: `Expected map but found 'string'.` [\#451](https://github.com/traefik/traefik/issues/451)
- Multiple path separators in the url path causing redirect [\#167](https://github.com/traefik/traefik/issues/167)

**Closed issues:**

- Global InsecureSkipVerify does not work [\#700](https://github.com/traefik/traefik/issues/700)
- \[documentation/feature\] Consul/etcd support atomic multiple key changes now [\#698](https://github.com/traefik/traefik/issues/698)
- How to configure which network to use when starting traefik binary? [\#694](https://github.com/traefik/traefik/issues/694)
- How to get multiple host headers working for docker labels? [\#692](https://github.com/traefik/traefik/issues/692)
- Requests with URL-encoded characters are not forwarded correctly [\#684](https://github.com/traefik/traefik/issues/684)
- Issue with global InsecureSkipVerify = true and self signed certificates [\#667](https://github.com/traefik/traefik/issues/667)
- Docker exposedbydefault = false didn't work [\#663](https://github.com/traefik/traefik/issues/663)
- \[ACME\] Auto SAN Detection [\#655](https://github.com/traefik/traefik/issues/655)
- Fronting a domain with DNS A-record round-robin & ACME [\#654](https://github.com/traefik/traefik/issues/654)
- Overriding toml configuration with environment variables [\#650](https://github.com/traefik/traefik/issues/650)
- marathon provider exposedByDefault = false [\#647](https://github.com/traefik/traefik/issues/647)
- Add status URL for service up checks [\#642](https://github.com/traefik/traefik/issues/642)
- acme's storage file, containing private key, is word readable [\#638](https://github.com/traefik/traefik/issues/638)
- wildcard domain with exclusions [\#633](https://github.com/traefik/traefik/issues/633)
- Enable evenly distribution among backend [\#631](https://github.com/traefik/traefik/issues/631)
- Traefik sporadically failing when proxying requests [\#615](https://github.com/traefik/traefik/issues/615)
- TCP Proxy [\#608](https://github.com/traefik/traefik/issues/608)
- How to use in Windows? [\#605](https://github.com/traefik/traefik/issues/605)
- `ClientCAFiles` ignored [\#604](https://github.com/traefik/traefik/issues/604)
- Let`s Encrypt enable in etcd [\#600](https://github.com/traefik/traefik/issues/600)
- Support HTTP Basic Auth [\#599](https://github.com/traefik/traefik/issues/599)
- Consul KV seem broken [\#587](https://github.com/traefik/traefik/issues/587)
- HTTPS entryPoint not working [\#574](https://github.com/traefik/traefik/issues/574)
- Traefik stuck when used as frontend for a streaming API [\#560](https://github.com/traefik/traefik/issues/560)
- Exclude some frontends in consul catalog [\#555](https://github.com/traefik/traefik/issues/555)
- Can I use Traefik without a domain name? [\#539](https://github.com/traefik/traefik/issues/539)
- Priorities in 1.0.0 not behaving [\#506](https://github.com/traefik/traefik/issues/506)
- Route by path [\#500](https://github.com/traefik/traefik/issues/500)
- Container IP Lost [\#375](https://github.com/traefik/traefik/issues/375)

**Merged pull requests:**

- Add HTTP compression [\#702](https://github.com/traefik/traefik/pull/702) ([tuier](https://github.com/tuier))
- Carry PR 446 - Add sticky session support \(round two!\) [\#701](https://github.com/traefik/traefik/pull/701) ([emilevauge](https://github.com/emilevauge))
- Remove unused endpoint when using constraints with Marathon provider [\#697](https://github.com/traefik/traefik/pull/697) ([tuier](https://github.com/tuier))
- Replace imagelayers.io with microbadger [\#696](https://github.com/traefik/traefik/pull/696) ([solidnerd](https://github.com/solidnerd))
- Selectable TLS Versions [\#690](https://github.com/traefik/traefik/pull/690) ([dtomcej](https://github.com/dtomcej))
- Carry pr 439 [\#689](https://github.com/traefik/traefik/pull/689) ([emilevauge](https://github.com/emilevauge))
- Disable gorilla/mux URL cleaning to prevent sending redirect [\#688](https://github.com/traefik/traefik/pull/688) ([ydubreuil](https://github.com/ydubreuil))
- Some fixes [\#687](https://github.com/traefik/traefik/pull/687) ([emilevauge](https://github.com/emilevauge))
- feat\(constraints\): Supports constraints for Marathon provider [\#686](https://github.com/traefik/traefik/pull/686) ([tuier](https://github.com/tuier))
- Update docs to improve contribution setup [\#685](https://github.com/traefik/traefik/pull/685) ([dtomcej](https://github.com/dtomcej))
- Add basic auth support for web backend [\#677](https://github.com/traefik/traefik/pull/677) ([SantoDE](https://github.com/SantoDE))
- Document accepted values for logLevel. [\#676](https://github.com/traefik/traefik/pull/676) ([jimmycuadra](https://github.com/jimmycuadra))
- If Marathon doesn't have healthcheck, assume it's ok [\#665](https://github.com/traefik/traefik/pull/665) ([gomes](https://github.com/gomes))
- ACME: renew certificates 30 days before expiry [\#660](https://github.com/traefik/traefik/pull/660) ([JayH5](https://github.com/JayH5))
- Update broken link and add a comment to sample config file  [\#658](https://github.com/traefik/traefik/pull/658) ([Yggdrasil](https://github.com/Yggdrasil))
- Add possibility to use BindPort IPAddress 👼 [\#657](https://github.com/traefik/traefik/pull/657) ([vdemeester](https://github.com/vdemeester))
- Update marathon [\#648](https://github.com/traefik/traefik/pull/648) ([emilevauge](https://github.com/emilevauge))
- Add backend features to docker [\#646](https://github.com/traefik/traefik/pull/646) ([jangie](https://github.com/jangie))
- enable consul catalog to use maxconn [\#645](https://github.com/traefik/traefik/pull/645) ([jangie](https://github.com/jangie))
- Adopt the Code Of Conduct from http://contributor-covenant.org [\#641](https://github.com/traefik/traefik/pull/641) ([errm](https://github.com/errm))
- Use secure mode 600 instead of 644 for acme.json [\#639](https://github.com/traefik/traefik/pull/639) ([discordianfish](https://github.com/discordianfish))
- docker clarification, fix dead urls, misc typos [\#637](https://github.com/traefik/traefik/pull/637) ([djalal](https://github.com/djalal))
- add PING handler to dashboard API [\#630](https://github.com/traefik/traefik/pull/630) ([jangie](https://github.com/jangie))
- Migrate to JobBackOff [\#628](https://github.com/traefik/traefik/pull/628) ([emilevauge](https://github.com/emilevauge))
- Add long job exponential backoff [\#627](https://github.com/traefik/traefik/pull/627) ([emilevauge](https://github.com/emilevauge))
- HA acme support [\#625](https://github.com/traefik/traefik/pull/625) ([emilevauge](https://github.com/emilevauge))
- Bump go v1.7 [\#620](https://github.com/traefik/traefik/pull/620) ([emilevauge](https://github.com/emilevauge))
- Make duration logging consistent [\#619](https://github.com/traefik/traefik/pull/619) ([jangie](https://github.com/jangie))
- fix for nil clientTLS causing issue [\#617](https://github.com/traefik/traefik/pull/617) ([jangie](https://github.com/jangie))
- Add ability for marathon provider to set maxconn values, loadbalancer algorithm, and circuit breaker expression [\#616](https://github.com/traefik/traefik/pull/616) ([jangie](https://github.com/jangie))
- Make systemd unit installable [\#613](https://github.com/traefik/traefik/pull/613) ([keis](https://github.com/keis))
- Merge v1.0.2 master [\#610](https://github.com/traefik/traefik/pull/610) ([emilevauge](https://github.com/emilevauge))
- update staert and flaeg [\#609](https://github.com/traefik/traefik/pull/609) ([cocap10](https://github.com/cocap10))
- \#504 Initial support for Docker 1.12 Swarm Mode [\#602](https://github.com/traefik/traefik/pull/602) ([diegofernandes](https://github.com/diegofernandes))
- Add Host cert ACME generation [\#601](https://github.com/traefik/traefik/pull/601) ([emilevauge](https://github.com/emilevauge))
- Fixed binary script so traefik version command doesn't just print default values [\#598](https://github.com/traefik/traefik/pull/598) ([keiths-osc](https://github.com/keiths-osc))
- Name servers after their pods [\#596](https://github.com/traefik/traefik/pull/596) ([errm](https://github.com/errm))
- Fix Consul prefix [\#589](https://github.com/traefik/traefik/pull/589) ([jippi](https://github.com/jippi))
- Prioritize kubernetes routes by path length [\#588](https://github.com/traefik/traefik/pull/588) ([philk](https://github.com/philk))
- beautify help [\#580](https://github.com/traefik/traefik/pull/580) ([cocap10](https://github.com/cocap10))
- Upgrade directives name since we use angular-ui-bootstrap [\#578](https://github.com/traefik/traefik/pull/578) ([micaelmbagira](https://github.com/micaelmbagira))
- Fix basic docs for configuration of multiple rules [\#576](https://github.com/traefik/traefik/pull/576) ([ajaegle](https://github.com/ajaegle))
- Fix k8s watch [\#573](https://github.com/traefik/traefik/pull/573) ([errm](https://github.com/errm))
- Add requirements.txt for netlify [\#567](https://github.com/traefik/traefik/pull/567) ([emilevauge](https://github.com/emilevauge))
- Merge v1.0.1 master [\#565](https://github.com/traefik/traefik/pull/565) ([emilevauge](https://github.com/emilevauge))
-  Move webui to FountainJS with Webpack [\#558](https://github.com/traefik/traefik/pull/558) ([micaelmbagira](https://github.com/micaelmbagira))
- Add global InsecureSkipVerify option to disable certificate checking [\#557](https://github.com/traefik/traefik/pull/557) ([stuart-c](https://github.com/stuart-c))
- Move version.go in its own package… [\#553](https://github.com/traefik/traefik/pull/553) ([vdemeester](https://github.com/vdemeester))
- Upgrade libkermit and dependencies [\#552](https://github.com/traefik/traefik/pull/552) ([vdemeester](https://github.com/vdemeester))
- Add command storeconfig [\#551](https://github.com/traefik/traefik/pull/551) ([cocap10](https://github.com/cocap10))
- Add basic/digest auth [\#547](https://github.com/traefik/traefik/pull/547) ([emilevauge](https://github.com/emilevauge))
- Bump node to 6 for webui [\#546](https://github.com/traefik/traefik/pull/546) ([vdemeester](https://github.com/vdemeester))
- Bump golang to 1.6.3 [\#545](https://github.com/traefik/traefik/pull/545) ([vdemeester](https://github.com/vdemeester))
- Fix typos [\#538](https://github.com/traefik/traefik/pull/538) ([jimt](https://github.com/jimt))
- Kubernetes user-guide [\#519](https://github.com/traefik/traefik/pull/519) ([errm](https://github.com/errm))
- Implement Kubernetes Selectors, minor kube endpoint fix [\#516](https://github.com/traefik/traefik/pull/516) ([pnegahdar](https://github.com/pnegahdar))
- Carry \#358 : Option to disable expose of all docker containers [\#514](https://github.com/traefik/traefik/pull/514) ([vdemeester](https://github.com/vdemeester))
- Remove traefik.frontend.value support in docker… [\#510](https://github.com/traefik/traefik/pull/510) ([vdemeester](https://github.com/vdemeester))
- Use KvStores as global config sources [\#481](https://github.com/traefik/traefik/pull/481) ([cocap10](https://github.com/cocap10))
- Add endpoint option to authenticate by client tls cert. [\#461](https://github.com/traefik/traefik/pull/461) ([andersbetner](https://github.com/andersbetner))
- add mesos provider inspired by mesos-dns & marathon provider [\#353](https://github.com/traefik/traefik/pull/353) ([skydjol](https://github.com/skydjol))

## [v1.0.2](https://github.com/traefik/traefik/tree/v1.0.2) (2016-08-02)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.0.1...v1.0.2)

**Fixed bugs:**

- ACME: revoke certificate on agreement update [\#579](https://github.com/traefik/traefik/issues/579)

**Closed issues:**

- Exclude some frontends in consul catalog [\#555](https://github.com/traefik/traefik/issues/555)

**Merged pull requests:**

- Bump oxy version, fix streaming [\#584](https://github.com/traefik/traefik/pull/584) ([emilevauge](https://github.com/emilevauge))
- Fix ACME TOS [\#582](https://github.com/traefik/traefik/pull/582) ([emilevauge](https://github.com/emilevauge))

## [v1.0.1](https://github.com/traefik/traefik/tree/v1.0.1) (2016-07-19)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.0.0...v1.0.1)

**Implemented enhancements:**

- Error with -consulcatalog and missing load balance method on 1.0.0 [\#524](https://github.com/traefik/traefik/issues/524)
- Kubernetes provider: should allow the master url to be override [\#501](https://github.com/traefik/traefik/issues/501)

**Fixed bugs:**

- Flag --etcd.endpoint default [\#508](https://github.com/traefik/traefik/issues/508)
- Conditional ACME on demand generation [\#505](https://github.com/traefik/traefik/issues/505)
- Important delay with streams \(Mozilla EventSource\) [\#503](https://github.com/traefik/traefik/issues/503)

**Closed issues:**

- Can I use Traefik without a domain name? [\#539](https://github.com/traefik/traefik/issues/539)
- Priorities in 1.0.0 not behaving [\#506](https://github.com/traefik/traefik/issues/506)
- Route by path [\#500](https://github.com/traefik/traefik/issues/500)

**Merged pull requests:**

- Update server.go [\#531](https://github.com/traefik/traefik/pull/531) ([Jsewill](https://github.com/Jsewill))
- Add sse support [\#527](https://github.com/traefik/traefik/pull/527) ([emilevauge](https://github.com/emilevauge))
- Fix acme checkOnDemandDomain [\#512](https://github.com/traefik/traefik/pull/512) ([emilevauge](https://github.com/emilevauge))
- Fix default etcd port [\#511](https://github.com/traefik/traefik/pull/511) ([errm](https://github.com/errm))

## [v1.0.0](https://github.com/traefik/traefik/tree/v1.0.0) (2016-07-05)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.0.0-rc3...v1.0.0)

**Fixed bugs:**

- Enable to define empty TLS option by flag for Let's Encrypt [\#488](https://github.com/traefik/traefik/issues/488)
- \[Docker\] No IP in backend in host networking mode [\#487](https://github.com/traefik/traefik/issues/487)
- Response is compressed when not requested [\#485](https://github.com/traefik/traefik/issues/485)
- loadConfig modifies configuration causing same config check to fail [\#480](https://github.com/traefik/traefik/issues/480)

**Closed issues:**

- svg logo [\#482](https://github.com/traefik/traefik/issues/482)
- etcd tries to connect with TLS even with --etcd.tls=false [\#456](https://github.com/traefik/traefik/issues/456)
- Zookeeper - KV connection error: Failed to test KV store connection [\#455](https://github.com/traefik/traefik/issues/455)
- "Not Found" api response needed instead of 404  [\#454](https://github.com/traefik/traefik/issues/454)
- domain label doesn't work on docker [\#447](https://github.com/traefik/traefik/issues/447)
- Any chance of a windows release? [\#425](https://github.com/traefik/traefik/issues/425)

**Merged pull requests:**

- Fix windows builds [\#495](https://github.com/traefik/traefik/pull/495) ([emilevauge](https://github.com/emilevauge))
- Fix host Docker network [\#494](https://github.com/traefik/traefik/pull/494) ([emilevauge](https://github.com/emilevauge))
- Fix empty tls flag [\#493](https://github.com/traefik/traefik/pull/493) ([emilevauge](https://github.com/emilevauge))
- Fix webui proxying [\#492](https://github.com/traefik/traefik/pull/492) ([emilevauge](https://github.com/emilevauge))
- Fix default weight in server.LoadConfig [\#491](https://github.com/traefik/traefik/pull/491) ([emilevauge](https://github.com/emilevauge))
- Fix retry headers, simplify ResponseRecorder [\#490](https://github.com/traefik/traefik/pull/490) ([emilevauge](https://github.com/emilevauge))

## [v1.0.0-rc3](https://github.com/traefik/traefik/tree/v1.0.0-rc3) (2016-06-23)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.0.0-rc2...v1.0.0-rc3)

**Implemented enhancements:**

- support more than one rule to Docker backend [\#419](https://github.com/traefik/traefik/issues/419)

**Fixed bugs:**

- consulCatalog issue when serviceName contains a dot [\#475](https://github.com/traefik/traefik/issues/475)
- Issue with empty responses [\#463](https://github.com/traefik/traefik/issues/463)
- Severe memory leak in beta.470 and beyond crashes Traefik server  [\#462](https://github.com/traefik/traefik/issues/462)
- Marathon that starts with a space causes parsing errors. [\#459](https://github.com/traefik/traefik/issues/459)
- A frontend route without a rule \(or empty rule\) causes a crash when traefik starts [\#453](https://github.com/traefik/traefik/issues/453)
- container dropped out when connecting to Docker Swarm [\#442](https://github.com/traefik/traefik/issues/442)
- Traefik setting Accept-Encoding: gzip on requests \(Traefik may also be broken with chunked responses\) [\#421](https://github.com/traefik/traefik/issues/421)

**Closed issues:**

- HTTP headers case gets modified [\#466](https://github.com/traefik/traefik/issues/466)
- File frontend \> Marathon Backend [\#465](https://github.com/traefik/traefik/issues/465)
- Websocket: Unable to hijack the connection [\#452](https://github.com/traefik/traefik/issues/452)
- kubernetes: Received event spamming? [\#449](https://github.com/traefik/traefik/issues/449)
- kubernetes: backends not updated when i scale replication controller? [\#448](https://github.com/traefik/traefik/issues/448)
- Add href link on frontend [\#436](https://github.com/traefik/traefik/issues/436)
- Multiple Domains Rule [\#430](https://github.com/traefik/traefik/issues/430)

**Merged pull requests:**

- Disable constraints in doc until 1.1 [\#479](https://github.com/traefik/traefik/pull/479) ([emilevauge](https://github.com/emilevauge))
- Sort nodes before creating consul catalog config [\#478](https://github.com/traefik/traefik/pull/478) ([keis](https://github.com/keis))
- Fix spamming events in listenProviders [\#477](https://github.com/traefik/traefik/pull/477) ([emilevauge](https://github.com/emilevauge))
- Fix empty responses [\#476](https://github.com/traefik/traefik/pull/476) ([emilevauge](https://github.com/emilevauge))
- Fix acme renew [\#472](https://github.com/traefik/traefik/pull/472) ([emilevauge](https://github.com/emilevauge))
- Fix typo in error message. [\#471](https://github.com/traefik/traefik/pull/471) ([KevinBusse](https://github.com/KevinBusse))
- Fix errors load config [\#470](https://github.com/traefik/traefik/pull/470) ([emilevauge](https://github.com/emilevauge))
- Typo: Replace French words by English ones [\#469](https://github.com/traefik/traefik/pull/469) ([kumy](https://github.com/kumy))
- Fix marathon TLS/basic auth [\#468](https://github.com/traefik/traefik/pull/468) ([emilevauge](https://github.com/emilevauge))
- Fix memory leak in listenProviders [\#464](https://github.com/traefik/traefik/pull/464) ([emilevauge](https://github.com/emilevauge))
- Fix websocket connection Hijack [\#460](https://github.com/traefik/traefik/pull/460) ([emilevauge](https://github.com/emilevauge))
- Fix default KV configuration [\#450](https://github.com/traefik/traefik/pull/450) ([emilevauge](https://github.com/emilevauge))
- Fix panic if listContainers fails… [\#443](https://github.com/traefik/traefik/pull/443) ([vdemeester](https://github.com/vdemeester))
- mount acme folder instead of file [\#441](https://github.com/traefik/traefik/pull/441) ([NicolasGeraud](https://github.com/NicolasGeraud))
- feat\(constraints\): Supports constraints for docker backend [\#438](https://github.com/traefik/traefik/pull/438) ([samber](https://github.com/samber))

## [v1.0.0-rc2](https://github.com/traefik/traefik/tree/v1.0.0-rc2) (2016-06-07)
[Full Changelog](https://github.com/traefik/traefik/compare/v1.0.0-rc1...v1.0.0-rc2)

**Implemented enhancements:**

- Add @samber to maintainers [\#440](https://github.com/traefik/traefik/pull/440) ([emilevauge](https://github.com/emilevauge))

**Fixed bugs:**

- Panic on help [\#429](https://github.com/traefik/traefik/issues/429)
- Bad default values in configuration [\#427](https://github.com/traefik/traefik/issues/427)

**Closed issues:**

- Traefik doesn't listen on IPv4 ports [\#434](https://github.com/traefik/traefik/issues/434)
- Not listening on port 80 [\#432](https://github.com/traefik/traefik/issues/432)
- docs need updating for new frontend rules format [\#423](https://github.com/traefik/traefik/issues/423)
- Does traefik supports for Mac? \(For development\)  [\#417](https://github.com/traefik/traefik/issues/417)

**Merged pull requests:**

- Allow multiple rules [\#435](https://github.com/traefik/traefik/pull/435) ([fclaeys](https://github.com/fclaeys))
- Add routes priorities [\#433](https://github.com/traefik/traefik/pull/433) ([emilevauge](https://github.com/emilevauge))
- Fix default configuration [\#428](https://github.com/traefik/traefik/pull/428) ([emilevauge](https://github.com/emilevauge))
- Fix marathon groups subdomain [\#426](https://github.com/traefik/traefik/pull/426) ([emilevauge](https://github.com/emilevauge))
- Fix travis tag check [\#422](https://github.com/traefik/traefik/pull/422) ([emilevauge](https://github.com/emilevauge))
- log info about TOML configuration file using [\#420](https://github.com/traefik/traefik/pull/420) ([cocap10](https://github.com/cocap10))
- Doc about skipping some integration tests with '-check.f ConsulCatalogSuite' [\#418](https://github.com/traefik/traefik/pull/418) ([samber](https://github.com/samber))



\* *This Change Log was automatically generated by [gcg](https://github.com/ldez/gcg)*
