# Change Log

## [v1.2.1](https://github.com/containous/traefik/tree/v1.2.1) (2017-03-27)
[Full Changelog](https://github.com/containous/traefik/compare/v1.2.0...v1.2.1)

**Merged pull requests:**

- bump lego 0e2937900 [\#1347](https://github.com/containous/traefik/pull/1347) ([emilevauge](https://github.com/emilevauge))
- k8s: Do not log service fields when GetService is failing. [\#1331](https://github.com/containous/traefik/pull/1331) ([timoreimann](https://github.com/timoreimann))

## [v1.2.0](https://github.com/containous/traefik/tree/v1.2.0) (2017-03-20)
[Full Changelog](https://github.com/containous/traefik/compare/v1.1.2...v1.2.0)

**Merged pull requests:**

- Docker: Added warning if network could not be found [\#1310](https://github.com/containous/traefik/pull/1310) ([zweizeichen](https://github.com/zweizeichen))
- Add filter on task status in addition to desired status \(Docker Provider - swarm\) [\#1304](https://github.com/containous/traefik/pull/1304) ([Yshayy](https://github.com/Yshayy))
- Abort Kubernetes Ingress update if Kubernetes API call fails [\#1295](https://github.com/containous/traefik/pull/1295) ([Regner](https://github.com/Regner))
- Small fixes [\#1291](https://github.com/containous/traefik/pull/1291) ([emilevauge](https://github.com/emilevauge))
- Rename health check URL parameter to path. [\#1285](https://github.com/containous/traefik/pull/1285) ([timoreimann](https://github.com/timoreimann))
- Update Oxy, fix for \#1199 [\#1278](https://github.com/containous/traefik/pull/1278) ([akanto](https://github.com/akanto))
- Fix metrics registering [\#1258](https://github.com/containous/traefik/pull/1258) ([matevzmihalic](https://github.com/matevzmihalic))
- Update DefaultMaxIdleConnsPerHost default in docs. [\#1239](https://github.com/containous/traefik/pull/1239) ([timoreimann](https://github.com/timoreimann))
- Update WSS/WS Proto \[Fixes \#670\] [\#1225](https://github.com/containous/traefik/pull/1225) ([dtomcej](https://github.com/dtomcej))
- Bump go-rancher version [\#1219](https://github.com/containous/traefik/pull/1219) ([SantoDE](https://github.com/SantoDE))
- Chunk taskArns into groups of 100 [\#1209](https://github.com/containous/traefik/pull/1209) ([owen](https://github.com/owen))
- Prepare release v1.2.0 rc2 [\#1204](https://github.com/containous/traefik/pull/1204) ([emilevauge](https://github.com/emilevauge))
- Revert "Ensure that we don't add balancees with no health check runs … [\#1198](https://github.com/containous/traefik/pull/1198) ([jangie](https://github.com/jangie))
- Small fixes and improvments [\#1173](https://github.com/containous/traefik/pull/1173) ([SantoDE](https://github.com/SantoDE))
- Fix docker issues with global and dead tasks [\#1167](https://github.com/containous/traefik/pull/1167) ([christopherobin](https://github.com/christopherobin))
- Better ECS error checking [\#1143](https://github.com/containous/traefik/pull/1143) ([lpetre](https://github.com/lpetre))
- Fix stats race condition [\#1141](https://github.com/containous/traefik/pull/1141) ([emilevauge](https://github.com/emilevauge))
- ECS: Docs - info about cred. resolution and required access policies [\#1137](https://github.com/containous/traefik/pull/1137) ([rickard-von-essen](https://github.com/rickard-von-essen))
- Healthcheck tests and doc [\#1132](https://github.com/containous/traefik/pull/1132) ([Juliens](https://github.com/Juliens))
- Fix travis deploy [\#1128](https://github.com/containous/traefik/pull/1128) ([emilevauge](https://github.com/emilevauge))
- Prepare release v1.2.0 rc1 [\#1126](https://github.com/containous/traefik/pull/1126) ([emilevauge](https://github.com/emilevauge))
- Fix checkout initial before calling rmpr [\#1124](https://github.com/containous/traefik/pull/1124) ([emilevauge](https://github.com/emilevauge))
- Feature rancher integration [\#1120](https://github.com/containous/traefik/pull/1120) ([SantoDE](https://github.com/SantoDE))
- Fix glide go units [\#1119](https://github.com/containous/traefik/pull/1119) ([emilevauge](https://github.com/emilevauge))
- Carry \#818 —  Add systemd watchdog feature [\#1116](https://github.com/containous/traefik/pull/1116) ([vdemeester](https://github.com/vdemeester))
- Skip file permission check on Windows [\#1115](https://github.com/containous/traefik/pull/1115) ([StefanScherer](https://github.com/StefanScherer))
- Fix Docker API version for Windows [\#1113](https://github.com/containous/traefik/pull/1113) ([StefanScherer](https://github.com/StefanScherer))
- Fix git rpr [\#1109](https://github.com/containous/traefik/pull/1109) ([emilevauge](https://github.com/emilevauge))
- Fix docker version specifier [\#1108](https://github.com/containous/traefik/pull/1108) ([timoreimann](https://github.com/timoreimann))
- Merge v1.1.2 master [\#1105](https://github.com/containous/traefik/pull/1105) ([emilevauge](https://github.com/emilevauge))
- add sh before script in deploy... [\#1103](https://github.com/containous/traefik/pull/1103) ([emilevauge](https://github.com/emilevauge))
- \[doc\] typo fixes for kubernetes user guide [\#1102](https://github.com/containous/traefik/pull/1102) ([bamarni](https://github.com/bamarni))
- add skip\_cleanup in deploy [\#1101](https://github.com/containous/traefik/pull/1101) ([emilevauge](https://github.com/emilevauge))
- Fix k8s example UI port. [\#1098](https://github.com/containous/traefik/pull/1098) ([ddunkin](https://github.com/ddunkin))
- Fix marathon provider [\#1090](https://github.com/containous/traefik/pull/1090) ([diegooliveira](https://github.com/diegooliveira))
- Add an ECS provider [\#1088](https://github.com/containous/traefik/pull/1088) ([lpetre](https://github.com/lpetre))
- Update comment to reflect the code [\#1087](https://github.com/containous/traefik/pull/1087) ([np](https://github.com/np))
- update NYTimes/gziphandler fixes \#1059 [\#1084](https://github.com/containous/traefik/pull/1084) ([JamesKyburz](https://github.com/JamesKyburz))
- Ensure that we don't add balancees with no health check runs if there is a health check defined on it [\#1080](https://github.com/containous/traefik/pull/1080) ([jangie](https://github.com/jangie))
- Add FreeBSD & OpenBSD to crossbinary [\#1078](https://github.com/containous/traefik/pull/1078) ([geoffgarside](https://github.com/geoffgarside))
- Fix metrics for multiple entry points [\#1071](https://github.com/containous/traefik/pull/1071) ([matevzmihalic](https://github.com/matevzmihalic))
- Allow setting load balancer method and sticky using service annotations [\#1068](https://github.com/containous/traefik/pull/1068) ([bakins](https://github.com/bakins))
- Fix travis script [\#1067](https://github.com/containous/traefik/pull/1067) ([emilevauge](https://github.com/emilevauge))
- Add missing fmt verb specifier in k8s provider. [\#1066](https://github.com/containous/traefik/pull/1066) ([timoreimann](https://github.com/timoreimann))
- Add git rpr command [\#1063](https://github.com/containous/traefik/pull/1063) ([emilevauge](https://github.com/emilevauge))
- Fix k8s example [\#1062](https://github.com/containous/traefik/pull/1062) ([emilevauge](https://github.com/emilevauge))
- Replace underscores to dash in autogenerated urls \(docker provider\) [\#1061](https://github.com/containous/traefik/pull/1061) ([WTFKr0](https://github.com/WTFKr0))
- Don't run go test on .glide cache folder [\#1057](https://github.com/containous/traefik/pull/1057) ([vdemeester](https://github.com/vdemeester))
- Allow setting circuitbreaker expression via Kubernetes annotation [\#1056](https://github.com/containous/traefik/pull/1056) ([bakins](https://github.com/bakins))
- Improving instrumentation. [\#1042](https://github.com/containous/traefik/pull/1042) ([enxebre](https://github.com/enxebre))
- Update user guide for upcoming `docker stack deploy`  [\#1041](https://github.com/containous/traefik/pull/1041) ([twelvelabs](https://github.com/twelvelabs))
- Support sticky sessions under SWARM Mode. \#1024 [\#1033](https://github.com/containous/traefik/pull/1033) ([foleymic](https://github.com/foleymic))
- Allow for wildcards in k8s ingress host, fixes \#792 [\#1029](https://github.com/containous/traefik/pull/1029) ([sheerun](https://github.com/sheerun))
- Don't fetch ACME certificates for frontends using non-TLS entrypoints \(\#989\) [\#1023](https://github.com/containous/traefik/pull/1023) ([syfonseq](https://github.com/syfonseq))
- Return Proper Non-ACME certificate - Fixes Issue 672 [\#1018](https://github.com/containous/traefik/pull/1018) ([dtomcej](https://github.com/dtomcej))
- Fix docs build and add missing benchmarks page [\#1017](https://github.com/containous/traefik/pull/1017) ([csabapalfi](https://github.com/csabapalfi))
- Set a NopCloser request body with retry middleware [\#1016](https://github.com/containous/traefik/pull/1016) ([bamarni](https://github.com/bamarni))
- instruct to flatten dependencies with glide [\#1010](https://github.com/containous/traefik/pull/1010) ([bamarni](https://github.com/bamarni))
- check permissions on acme.json during startup [\#1009](https://github.com/containous/traefik/pull/1009) ([bamarni](https://github.com/bamarni))
- \[doc\] few tweaks on the basics page [\#1005](https://github.com/containous/traefik/pull/1005) ([bamarni](https://github.com/bamarni))
- Import order as goimports does [\#1004](https://github.com/containous/traefik/pull/1004) ([vdemeester](https://github.com/vdemeester))
- See the right go report badge [\#991](https://github.com/containous/traefik/pull/991) ([guilhem](https://github.com/guilhem))
- Add multiple values for one rule to docs [\#978](https://github.com/containous/traefik/pull/978) ([j0hnsmith](https://github.com/j0hnsmith))
- Add ACME/Let’s Encrypt integration tests [\#975](https://github.com/containous/traefik/pull/975) ([trecloux](https://github.com/trecloux))
- deploy.sh: upload release source tarball [\#969](https://github.com/containous/traefik/pull/969) ([Mic92](https://github.com/Mic92))
- toml zookeeper doc fix [\#948](https://github.com/containous/traefik/pull/948) ([brdude](https://github.com/brdude))
- Add Rule AddPrefix [\#931](https://github.com/containous/traefik/pull/931) ([Juliens](https://github.com/Juliens))
- Add bug command [\#921](https://github.com/containous/traefik/pull/921) ([emilevauge](https://github.com/emilevauge))
- \(WIP\) feat: HealthCheck [\#918](https://github.com/containous/traefik/pull/918) ([Juliens](https://github.com/Juliens))
- Add ability to set authenticated user in request header [\#889](https://github.com/containous/traefik/pull/889) ([ViViDboarder](https://github.com/ViViDboarder))
- IP-per-task: [\#841](https://github.com/containous/traefik/pull/841) ([diegooliveira](https://github.com/diegooliveira))

## [v1.2.0-rc2](https://github.com/containous/traefik/tree/v1.2.0-rc2) (2017-03-01)
[Full Changelog](https://github.com/containous/traefik/compare/v1.2.0-rc1...v1.2.0-rc2)

**Implemented enhancements:**

- Are there plans to support the service type ExternalName in Kubernetes? [\#1142](https://github.com/containous/traefik/issues/1142)
- Kubernetes Ingress and sticky support [\#911](https://github.com/containous/traefik/issues/911)
- kubernetes client does not support InsecureSkipVerify [\#876](https://github.com/containous/traefik/issues/876)
- Support active health checking like HAProxy [\#824](https://github.com/containous/traefik/issues/824)
- Allow k8s ingress controller serviceAccountToken and serviceAccountCACert to be changed [\#611](https://github.com/containous/traefik/issues/611)

**Fixed bugs:**

- \[rancher\] invalid memory address or nil pointer dereference [\#1134](https://github.com/containous/traefik/issues/1134)
- Kubernetes default backend should work [\#1073](https://github.com/containous/traefik/issues/1073)

**Closed issues:**

- Are release Download links broken? [\#1201](https://github.com/containous/traefik/issues/1201)
- Bind to specific ip address [\#1193](https://github.com/containous/traefik/issues/1193)
- DNS01 challenge use the wrong zone through route53 [\#1192](https://github.com/containous/traefik/issues/1192)
- Reverse proxy https to http backends fails [\#1180](https://github.com/containous/traefik/issues/1180)
- Swarm Mode + Letsecrypt + KV Store [\#1176](https://github.com/containous/traefik/issues/1176)
- docker deploy -c example.yml    e [\#1169](https://github.com/containous/traefik/issues/1169)
- Traefik not finding dynamically added services \(Docker Swarm Mode\) [\#1168](https://github.com/containous/traefik/issues/1168)
- Traefik with Kubernetes backend - keep getting 401 on all GET requests to kube-apiserver [\#1166](https://github.com/containous/traefik/issues/1166)
- Near line 15 \(last key parsed 'backends.backend-monitor-viz.servers'\): Key 'backends.backend-monitor-viz.servers.server-monitor\_viz-1' has already been defined. [\#1154](https://github.com/containous/traefik/issues/1154)
- How to reuse SSL certificates automatically fetched from Let´s encrypt? [\#1152](https://github.com/containous/traefik/issues/1152)
- Dynamically ban ip when backend repeatedly returns specified status code. \( 403 \) [\#1136](https://github.com/containous/traefik/issues/1136)
- Always get 404 accessing my nginx backend service [\#1112](https://github.com/containous/traefik/issues/1112)
- Incomplete Docu [\#1091](https://github.com/containous/traefik/issues/1091)
- LoadCertificateForDomains: runtime error: invalid memory address [\#1069](https://github.com/containous/traefik/issues/1069)
- Traefik creating backends & mappings for ingress annotated with ingress.class: nginx [\#1058](https://github.com/containous/traefik/issues/1058)
- ACME file format description [\#1012](https://github.com/containous/traefik/issues/1012)
- SwarmMode - Not routing on worker node [\#838](https://github.com/containous/traefik/issues/838)
- Migrate k8s to kubernetes/client-go  [\#678](https://github.com/containous/traefik/issues/678)
- Support for sticky session with kubernetes ingress as backend [\#674](https://github.com/containous/traefik/issues/674)

**Merged pull requests:**

- Revert "Ensure that we don't add balancees with no health check runs … [\#1198](https://github.com/containous/traefik/pull/1198) ([jangie](https://github.com/jangie))
- Small fixes and improvments [\#1173](https://github.com/containous/traefik/pull/1173) ([SantoDE](https://github.com/SantoDE))
- Fix docker issues with global and dead tasks [\#1167](https://github.com/containous/traefik/pull/1167) ([christopherobin](https://github.com/christopherobin))
- Better ECS error checking [\#1143](https://github.com/containous/traefik/pull/1143) ([lpetre](https://github.com/lpetre))
- Fix stats race condition [\#1141](https://github.com/containous/traefik/pull/1141) ([emilevauge](https://github.com/emilevauge))
- ECS: Docs - info about cred. resolution and required access policies [\#1137](https://github.com/containous/traefik/pull/1137) ([rickard-von-essen](https://github.com/rickard-von-essen))
- Healthcheck tests and doc [\#1132](https://github.com/containous/traefik/pull/1132) ([Juliens](https://github.com/Juliens))

## [v1.2.0-rc1](https://github.com/containous/traefik/tree/v1.2.0-rc1) (2017-02-06)
[Full Changelog](https://github.com/containous/traefik/compare/v1.1.2...v1.2.0-rc1)

**Implemented enhancements:**

- Add FreeBSD and OpenBSD to release builds [\#923](https://github.com/containous/traefik/issues/923)
- Write authenticated user to header key [\#802](https://github.com/containous/traefik/issues/802)
- Question: Wildcard Host for Kubernetes Ingress [\#792](https://github.com/containous/traefik/issues/792)
- First commit prometheus middleware. [\#1022](https://github.com/containous/traefik/pull/1022) ([enxebre](https://github.com/enxebre))
- Use deployment primitives from travis [\#843](https://github.com/containous/traefik/pull/843) ([guilhem](https://github.com/guilhem))

**Fixed bugs:**

- Increase Docker API version to work with Windows Containers [\#1094](https://github.com/containous/traefik/issues/1094)

**Closed issues:**

- How could I know whether forwarding path is correctly set? [\#1111](https://github.com/containous/traefik/issues/1111)
- ACME + Docker-compose labels [\#1099](https://github.com/containous/traefik/issues/1099)
- Loadbalance between 2 containers in Docker Swarm Mode [\#1095](https://github.com/containous/traefik/issues/1095)
- Add DNS01 letsencrypt challenge support through AWS. [\#1093](https://github.com/containous/traefik/issues/1093)
- New Release Cut [\#1092](https://github.com/containous/traefik/issues/1092)
- Marathon integration changed default backend server port from task-level to application-level [\#1072](https://github.com/containous/traefik/issues/1072)
- websockets not working when compress = true in toml config. [\#1059](https://github.com/containous/traefik/issues/1059)
- Proxying 403 http status into the application [\#1044](https://github.com/containous/traefik/issues/1044)
- Normalize auto generated frontend-rule \(docker\) [\#1043](https://github.com/containous/traefik/issues/1043)
- Traefik with Consul catalog backend + Registrator [\#1039](https://github.com/containous/traefik/issues/1039)
- \[Configuration help\] Can't connect to docker containers under a domain path [\#1032](https://github.com/containous/traefik/issues/1032)
- Kubernetes and etcd backend : `storeconfig` fails. [\#1031](https://github.com/containous/traefik/issues/1031)
- kubernetes: Undefined backend 'X/' for frontend X/" [\#1026](https://github.com/containous/traefik/issues/1026)
- TLS handshake error [\#1025](https://github.com/containous/traefik/issues/1025)
- Traefik failing on POST request [\#1008](https://github.com/containous/traefik/issues/1008)
- how config traffic.toml http 80 without basic auth, traefik WebUI 8080 with basic auth [\#1001](https://github.com/containous/traefik/issues/1001)
- Docs 404 [\#995](https://github.com/containous/traefik/issues/995)
- Disable acme for non https endpoints [\#989](https://github.com/containous/traefik/issues/989)
- Add parameter to configure TLS entrypoints with ca-bundle file [\#984](https://github.com/containous/traefik/issues/984)
- docker multiple networks routing [\#970](https://github.com/containous/traefik/issues/970)
- don't add Docker containers not on the same network as traefik [\#959](https://github.com/containous/traefik/issues/959)
- Multiple frontend routes [\#957](https://github.com/containous/traefik/issues/957)
- SNI based routing without TLS offloading [\#933](https://github.com/containous/traefik/issues/933)
- NEO4J + traefik proxy Issues  [\#907](https://github.com/containous/traefik/issues/907)
- ACME OnDemand ignores entrypoint certificate [\#672](https://github.com/containous/traefik/issues/672)
- Ability to use self-signed certificates for local development [\#399](https://github.com/containous/traefik/issues/399)

**Merged pull requests:**

- Fix checkout initial before calling rmpr [\#1124](https://github.com/containous/traefik/pull/1124) ([emilevauge](https://github.com/emilevauge))
- Feature rancher integration [\#1120](https://github.com/containous/traefik/pull/1120) ([SantoDE](https://github.com/SantoDE))
- Fix glide go units [\#1119](https://github.com/containous/traefik/pull/1119) ([emilevauge](https://github.com/emilevauge))
- Carry \#818 —  Add systemd watchdog feature [\#1116](https://github.com/containous/traefik/pull/1116) ([vdemeester](https://github.com/vdemeester))
- Skip file permission check on Windows [\#1115](https://github.com/containous/traefik/pull/1115) ([StefanScherer](https://github.com/StefanScherer))
- Fix Docker API version for Windows [\#1113](https://github.com/containous/traefik/pull/1113) ([StefanScherer](https://github.com/StefanScherer))
- Fix git rpr [\#1109](https://github.com/containous/traefik/pull/1109) ([emilevauge](https://github.com/emilevauge))
- Fix docker version specifier [\#1108](https://github.com/containous/traefik/pull/1108) ([timoreimann](https://github.com/timoreimann))
- Merge v1.1.2 master [\#1105](https://github.com/containous/traefik/pull/1105) ([emilevauge](https://github.com/emilevauge))
- add sh before script in deploy... [\#1103](https://github.com/containous/traefik/pull/1103) ([emilevauge](https://github.com/emilevauge))
- \[doc\] typo fixes for kubernetes user guide [\#1102](https://github.com/containous/traefik/pull/1102) ([bamarni](https://github.com/bamarni))
- add skip\_cleanup in deploy [\#1101](https://github.com/containous/traefik/pull/1101) ([emilevauge](https://github.com/emilevauge))
- Fix k8s example UI port. [\#1098](https://github.com/containous/traefik/pull/1098) ([ddunkin](https://github.com/ddunkin))
- Fix marathon provider [\#1090](https://github.com/containous/traefik/pull/1090) ([diegooliveira](https://github.com/diegooliveira))
- Add an ECS provider [\#1088](https://github.com/containous/traefik/pull/1088) ([lpetre](https://github.com/lpetre))
- Update comment to reflect the code [\#1087](https://github.com/containous/traefik/pull/1087) ([np](https://github.com/np))
- update NYTimes/gziphandler fixes \#1059 [\#1084](https://github.com/containous/traefik/pull/1084) ([JamesKyburz](https://github.com/JamesKyburz))
- Ensure that we don't add balancees with no health check runs if there is a health check defined on it [\#1080](https://github.com/containous/traefik/pull/1080) ([jangie](https://github.com/jangie))
- Add FreeBSD & OpenBSD to crossbinary [\#1078](https://github.com/containous/traefik/pull/1078) ([geoffgarside](https://github.com/geoffgarside))
- Fix metrics for multiple entry points [\#1071](https://github.com/containous/traefik/pull/1071) ([matevzmihalic](https://github.com/matevzmihalic))
- Allow setting load balancer method and sticky using service annotations [\#1068](https://github.com/containous/traefik/pull/1068) ([bakins](https://github.com/bakins))
- Fix travis script [\#1067](https://github.com/containous/traefik/pull/1067) ([emilevauge](https://github.com/emilevauge))
- Add missing fmt verb specifier in k8s provider. [\#1066](https://github.com/containous/traefik/pull/1066) ([timoreimann](https://github.com/timoreimann))
- Add git rpr command [\#1063](https://github.com/containous/traefik/pull/1063) ([emilevauge](https://github.com/emilevauge))
- Fix k8s example [\#1062](https://github.com/containous/traefik/pull/1062) ([emilevauge](https://github.com/emilevauge))
- Replace underscores to dash in autogenerated urls \(docker provider\) [\#1061](https://github.com/containous/traefik/pull/1061) ([WTFKr0](https://github.com/WTFKr0))
- Don't run go test on .glide cache folder [\#1057](https://github.com/containous/traefik/pull/1057) ([vdemeester](https://github.com/vdemeester))
- Allow setting circuitbreaker expression via Kubernetes annotation [\#1056](https://github.com/containous/traefik/pull/1056) ([bakins](https://github.com/bakins))
- Improving instrumentation. [\#1042](https://github.com/containous/traefik/pull/1042) ([enxebre](https://github.com/enxebre))
- Update user guide for upcoming `docker stack deploy`  [\#1041](https://github.com/containous/traefik/pull/1041) ([twelvelabs](https://github.com/twelvelabs))
- Support sticky sessions under SWARM Mode. \#1024 [\#1033](https://github.com/containous/traefik/pull/1033) ([foleymic](https://github.com/foleymic))
- Allow for wildcards in k8s ingress host, fixes \#792 [\#1029](https://github.com/containous/traefik/pull/1029) ([sheerun](https://github.com/sheerun))
- Don't fetch ACME certificates for frontends using non-TLS entrypoints \(\#989\) [\#1023](https://github.com/containous/traefik/pull/1023) ([syfonseq](https://github.com/syfonseq))
- Return Proper Non-ACME certificate - Fixes Issue 672 [\#1018](https://github.com/containous/traefik/pull/1018) ([dtomcej](https://github.com/dtomcej))
- Fix docs build and add missing benchmarks page [\#1017](https://github.com/containous/traefik/pull/1017) ([csabapalfi](https://github.com/csabapalfi))
- Set a NopCloser request body with retry middleware [\#1016](https://github.com/containous/traefik/pull/1016) ([bamarni](https://github.com/bamarni))
- instruct to flatten dependencies with glide [\#1010](https://github.com/containous/traefik/pull/1010) ([bamarni](https://github.com/bamarni))
- check permissions on acme.json during startup [\#1009](https://github.com/containous/traefik/pull/1009) ([bamarni](https://github.com/bamarni))
- \[doc\] few tweaks on the basics page [\#1005](https://github.com/containous/traefik/pull/1005) ([bamarni](https://github.com/bamarni))
- Import order as goimports does [\#1004](https://github.com/containous/traefik/pull/1004) ([vdemeester](https://github.com/vdemeester))
- See the right go report badge [\#991](https://github.com/containous/traefik/pull/991) ([guilhem](https://github.com/guilhem))
- Add multiple values for one rule to docs [\#978](https://github.com/containous/traefik/pull/978) ([j0hnsmith](https://github.com/j0hnsmith))
- Add ACME/Let’s Encrypt integration tests [\#975](https://github.com/containous/traefik/pull/975) ([trecloux](https://github.com/trecloux))
- deploy.sh: upload release source tarball [\#969](https://github.com/containous/traefik/pull/969) ([Mic92](https://github.com/Mic92))
- toml zookeeper doc fix [\#948](https://github.com/containous/traefik/pull/948) ([brdude](https://github.com/brdude))
- Add Rule AddPrefix [\#931](https://github.com/containous/traefik/pull/931) ([Juliens](https://github.com/Juliens))
- Add bug command [\#921](https://github.com/containous/traefik/pull/921) ([emilevauge](https://github.com/emilevauge))
- \(WIP\) feat: HealthCheck [\#918](https://github.com/containous/traefik/pull/918) ([Juliens](https://github.com/Juliens))
- Add ability to set authenticated user in request header [\#889](https://github.com/containous/traefik/pull/889) ([ViViDboarder](https://github.com/ViViDboarder))
- IP-per-task: [\#841](https://github.com/containous/traefik/pull/841) ([diegooliveira](https://github.com/diegooliveira))

## [v1.1.2](https://github.com/containous/traefik/tree/v1.1.2) (2016-12-15)
[Full Changelog](https://github.com/containous/traefik/compare/v1.1.1...v1.1.2)

**Fixed bugs:**

- Problem during HTTPS redirection [\#952](https://github.com/containous/traefik/issues/952)
- nil pointer with kubernetes ingress [\#934](https://github.com/containous/traefik/issues/934)
- ConsulCatalog and File not working [\#903](https://github.com/containous/traefik/issues/903)
- Traefik can not start [\#902](https://github.com/containous/traefik/issues/902)
- Cannot connect to Kubernetes server failed to decode watch event [\#532](https://github.com/containous/traefik/issues/532)

**Closed issues:**

- Updating certificates with configuration file. [\#968](https://github.com/containous/traefik/issues/968)
- Let's encrypt retrieving certificate from wrong IP [\#962](https://github.com/containous/traefik/issues/962)
- let's encrypt and dashboard? [\#961](https://github.com/containous/traefik/issues/961)
- Working HTTPS example for GKE? [\#960](https://github.com/containous/traefik/issues/960)
- GKE design pattern [\#958](https://github.com/containous/traefik/issues/958)
- Consul Catalog constraints does not seem to work [\#954](https://github.com/containous/traefik/issues/954)
- Issue in building traefik from master [\#949](https://github.com/containous/traefik/issues/949)
- Proxy http application to https doesn't seem to work correctly for all services [\#937](https://github.com/containous/traefik/issues/937)
- Excessive requests to kubernetes apiserver [\#922](https://github.com/containous/traefik/issues/922)
- I am getting a connection error while creating traefik with consul backend "dial tcp 127.0.0.1:8500: getsockopt: connection refused" [\#917](https://github.com/containous/traefik/issues/917)
- SwarmMode - 1.13 RC2 - DNS RR - Individual IPs not retrieved [\#913](https://github.com/containous/traefik/issues/913)
- Panic in kubernetes ingress \(traefik 1.1.0\) [\#910](https://github.com/containous/traefik/issues/910)
- Kubernetes updating deployment image requires Ingress to be remade  [\#909](https://github.com/containous/traefik/issues/909)
- \[ACME\] Too many currently pending authorizations [\#905](https://github.com/containous/traefik/issues/905)
- WEB UI Authentication and Let's Encrypt : error 404 [\#754](https://github.com/containous/traefik/issues/754)
- Traefik as ingress controller for SNI based routing in kubernetes [\#745](https://github.com/containous/traefik/issues/745)
- Kubernetes Ingress backend: using self-signed certificates [\#486](https://github.com/containous/traefik/issues/486)
- Kubernetes Ingress backend: can't find token and ca.crt [\#484](https://github.com/containous/traefik/issues/484)

**Merged pull requests:**

- Fix duplicate acme certificates [\#972](https://github.com/containous/traefik/pull/972) ([emilevauge](https://github.com/emilevauge))
- Fix leadership panic [\#956](https://github.com/containous/traefik/pull/956) ([emilevauge](https://github.com/emilevauge))
- Fix redirect regex [\#947](https://github.com/containous/traefik/pull/947) ([emilevauge](https://github.com/emilevauge))
- Add operation recover [\#944](https://github.com/containous/traefik/pull/944) ([emilevauge](https://github.com/emilevauge))

## [v1.1.1](https://github.com/containous/traefik/tree/v1.1.1) (2016-11-29)
[Full Changelog](https://github.com/containous/traefik/compare/v1.1.0...v1.1.1)

**Implemented enhancements:**

- Getting "Kubernetes connection error failed to decode watch event : unexpected EOF" every minute in Traefik log [\#732](https://github.com/containous/traefik/issues/732)

**Fixed bugs:**

- 1.1.0 kubernetes panic: send on closed channel [\#877](https://github.com/containous/traefik/issues/877)
- digest auth example is incorrect [\#869](https://github.com/containous/traefik/issues/869)
- Marathon & Mesos providers' GroupsAsSubDomains option broken [\#867](https://github.com/containous/traefik/issues/867)
- 404 responses when a new Marathon leader is elected [\#653](https://github.com/containous/traefik/issues/653)

**Closed issues:**

- traefik:latest fails to auto-detect Docker containers [\#901](https://github.com/containous/traefik/issues/901)
- Panic error on bare metal Kubernetes \(installed with Kubeadm\) [\#897](https://github.com/containous/traefik/issues/897)
- api backend readOnly: what is the purpose of this setting [\#893](https://github.com/containous/traefik/issues/893)
- file backend: using external file - doesn't work [\#892](https://github.com/containous/traefik/issues/892)
- auth support for web backend [\#891](https://github.com/containous/traefik/issues/891)
- Basic auth with docker labels [\#890](https://github.com/containous/traefik/issues/890)
- file vs inline config [\#888](https://github.com/containous/traefik/issues/888)
- combine Host and HostRegexp rules [\#882](https://github.com/containous/traefik/issues/882)
- \[Question\] Traefik + Kubernetes + Let's Encrypt \(ssl not used\) [\#881](https://github.com/containous/traefik/issues/881)
- Traefik security for dashboard [\#880](https://github.com/containous/traefik/issues/880)
- Kubernetes Nginx Deployment Panic [\#879](https://github.com/containous/traefik/issues/879)
- Kubernetes Example Address already in use [\#872](https://github.com/containous/traefik/issues/872)
- ETCD Backend - frontend/backends missing [\#866](https://github.com/containous/traefik/issues/866)
- \[Swarm mode\] Dashboard does not work on RC4 [\#864](https://github.com/containous/traefik/issues/864)
- Docker v1.1.0 image does not exist [\#861](https://github.com/containous/traefik/issues/861)
- ConsulService catalog do not support multiple rules [\#859](https://github.com/containous/traefik/issues/859)
- Update official docker repo [\#858](https://github.com/containous/traefik/issues/858)
- Still a memory leak with k8s - 1.1 RC4 [\#844](https://github.com/containous/traefik/issues/844)

**Merged pull requests:**

- Fix Swarm panic [\#908](https://github.com/containous/traefik/pull/908) ([emilevauge](https://github.com/emilevauge))
- Fix k8s panic [\#900](https://github.com/containous/traefik/pull/900) ([emilevauge](https://github.com/emilevauge))
- Fix missing value for k8s watch request parameter [\#874](https://github.com/containous/traefik/pull/874) ([codablock](https://github.com/codablock))
- Fix GroupsAsSubDomains option for Mesos and Marathon [\#868](https://github.com/containous/traefik/pull/868) ([ryanleary](https://github.com/ryanleary))
- Normalize backend even if is user-defined [\#865](https://github.com/containous/traefik/pull/865) ([WTFKr0](https://github.com/WTFKr0))
- consul/kv.tmpl: weight default value should be a int [\#826](https://github.com/containous/traefik/pull/826) ([klausenbusk](https://github.com/klausenbusk))

## [v1.1.0](https://github.com/containous/traefik/tree/v1.1.0) (2016-11-17)
[Full Changelog](https://github.com/containous/traefik/compare/v1.0.0...v1.1.0)

**Implemented enhancements:**

- Support healthcheck if present for docker [\#666](https://github.com/containous/traefik/issues/666)
- Standard unit for traefik latency in access log [\#559](https://github.com/containous/traefik/issues/559)
- \[CI\] wiredep marked as unmaintained [\#550](https://github.com/containous/traefik/issues/550)
- Feature Request: Enable Health checks to containers. [\#540](https://github.com/containous/traefik/issues/540)
- Feature Request: SSL Cipher Selection [\#535](https://github.com/containous/traefik/issues/535)
- Error with -consulcatalog and missing load balance method on 1.0.0 [\#524](https://github.com/containous/traefik/issues/524)
- Running Traefik with Docker 1.12 Swarm Mode [\#504](https://github.com/containous/traefik/issues/504)
- Kubernetes provider: should allow the master url to be override [\#501](https://github.com/containous/traefik/issues/501)
- \[FRONTEND\]\[LE\] Pre-generate SSL certificates for "Host:" rules [\#483](https://github.com/containous/traefik/issues/483)
- Frontend Rule evolution [\#437](https://github.com/containous/traefik/issues/437)
- Add a Changelog [\#388](https://github.com/containous/traefik/issues/388)
- Add label matching for kubernetes ingests [\#363](https://github.com/containous/traefik/issues/363)
- Acme in HA Traefik Scenario [\#348](https://github.com/containous/traefik/issues/348)
- HTTP Basic Auth support [\#77](https://github.com/containous/traefik/issues/77)
- Session affinity / stickiness / persistence [\#5](https://github.com/containous/traefik/issues/5)

**Fixed bugs:**

- 1.1.0-rc4 dashboard UX not displaying [\#828](https://github.com/containous/traefik/issues/828)
- Traefik stopped serving on upgrade to v1.1.0-rc3 [\#807](https://github.com/containous/traefik/issues/807)
- cannot access webui/dashboard  [\#796](https://github.com/containous/traefik/issues/796)
- Traefik cannot read constraints from KV [\#794](https://github.com/containous/traefik/issues/794)
- HTTP2 - configuration [\#790](https://github.com/containous/traefik/issues/790)
- Cannot provide multiple certificates using flag [\#757](https://github.com/containous/traefik/issues/757)
- Allow multiple certificates on a single entrypoint when trying to use TLS? [\#747](https://github.com/containous/traefik/issues/747)
- traefik \* Users: unsupported type: slice [\#743](https://github.com/containous/traefik/issues/743)
- \[Docker swarm mode\] The traefik.docker.network seems to have no effect [\#719](https://github.com/containous/traefik/issues/719)
- traefik hangs - stops handling requests [\#662](https://github.com/containous/traefik/issues/662)
- Add long jobs in exponential backoff providers  [\#626](https://github.com/containous/traefik/issues/626)
- Tip of tree crashes on invalid pointer on Marathon provider [\#624](https://github.com/containous/traefik/issues/624)
- ACME: revoke certificate on agreement update [\#579](https://github.com/containous/traefik/issues/579)
- WebUI: Providers tabs disappeared [\#577](https://github.com/containous/traefik/issues/577)
- traefik version command contains incorrect information when building from master branch [\#569](https://github.com/containous/traefik/issues/569)
- Case sensitive domain names breaks routing  [\#562](https://github.com/containous/traefik/issues/562)
- Flag --etcd.endpoint default [\#508](https://github.com/containous/traefik/issues/508)
- Conditional ACME on demand generation [\#505](https://github.com/containous/traefik/issues/505)
- Important delay with streams \(Mozilla EventSource\) [\#503](https://github.com/containous/traefik/issues/503)
- Traefik crashing [\#458](https://github.com/containous/traefik/issues/458)
- traefik.toml constraints error: `Expected map but found 'string'.` [\#451](https://github.com/containous/traefik/issues/451)
- Multiple path separators in the url path causing redirect [\#167](https://github.com/containous/traefik/issues/167)

**Closed issues:**

- All path rules require paths to be lowercase [\#851](https://github.com/containous/traefik/issues/851)
- The UI stops working after a time and have to restart the service. [\#840](https://github.com/containous/traefik/issues/840)
- Incorrect Dashboard page returned [\#831](https://github.com/containous/traefik/issues/831)
- LoadBalancing doesn't work in single node Swarm-mode [\#815](https://github.com/containous/traefik/issues/815)
- cannot connect to docker daemon [\#813](https://github.com/containous/traefik/issues/813)
- Let's encrypt configuration not working [\#805](https://github.com/containous/traefik/issues/805)
- Multiple subdomains for Marathon backend. [\#785](https://github.com/containous/traefik/issues/785)
- traefik-1.1.0-rc1: build error [\#781](https://github.com/containous/traefik/issues/781)
- dependencies installation error [\#755](https://github.com/containous/traefik/issues/755)
- k8s provider w/ acme? [\#752](https://github.com/containous/traefik/issues/752)
- Swarm Docs - How to use a FQDN [\#744](https://github.com/containous/traefik/issues/744)
- Documented ProvidersThrottleDuration value is invalid [\#741](https://github.com/containous/traefik/issues/741)
- Sensible configuration for consulCatalog [\#737](https://github.com/containous/traefik/issues/737)
- Traefik ignoring container listening in more than one TCP port [\#734](https://github.com/containous/traefik/issues/734)
- Loadbalaning issues with traefik and Docker Swarm cluster [\#730](https://github.com/containous/traefik/issues/730)
- issues with marathon app ids containing a dot [\#726](https://github.com/containous/traefik/issues/726)
- Error when using HA acme in kubernetes with etcd [\#725](https://github.com/containous/traefik/issues/725)
- \[Docker swarm mode\] No round robin when using service [\#718](https://github.com/containous/traefik/issues/718)
- Dose it support docker swarm mode  [\#712](https://github.com/containous/traefik/issues/712)
- Kubernetes - Undefined backend  [\#710](https://github.com/containous/traefik/issues/710)
- How Routing traffic depending on path not domain in docker [\#706](https://github.com/containous/traefik/issues/706)
- Constraints on Consul Catalogue not working as expected [\#703](https://github.com/containous/traefik/issues/703)
- Global InsecureSkipVerify does not work [\#700](https://github.com/containous/traefik/issues/700)
- Traefik crashes when using Consul catalog [\#699](https://github.com/containous/traefik/issues/699)
- \[documentation/feature\] Consul/etcd support atomic multiple key changes now [\#698](https://github.com/containous/traefik/issues/698)
- How to configure which network to use when starting traefik binary? [\#694](https://github.com/containous/traefik/issues/694)
- How to get multiple host headers working for docker labels? [\#692](https://github.com/containous/traefik/issues/692)
- Requests with URL-encoded characters are not forwarded correctly [\#684](https://github.com/containous/traefik/issues/684)
- File Watcher for rules does not work  [\#683](https://github.com/containous/traefik/issues/683)
- Issue with global InsecureSkipVerify = true and self signed certificates [\#667](https://github.com/containous/traefik/issues/667)
- Docker exposedbydefault = false didn't work [\#663](https://github.com/containous/traefik/issues/663)
- swarm documentation needs update [\#656](https://github.com/containous/traefik/issues/656)
- \[ACME\] Auto SAN Detection [\#655](https://github.com/containous/traefik/issues/655)
- Fronting a domain with DNS A-record round-robin & ACME [\#654](https://github.com/containous/traefik/issues/654)
- Overriding toml configuration with environment variables [\#650](https://github.com/containous/traefik/issues/650)
- marathon provider exposedByDefault = false [\#647](https://github.com/containous/traefik/issues/647)
- Add status URL for service up checks [\#642](https://github.com/containous/traefik/issues/642)
- acme's storage file, containing private key, is word readable [\#638](https://github.com/containous/traefik/issues/638)
- wildcard domain with exclusions [\#633](https://github.com/containous/traefik/issues/633)
- Enable evenly distribution among backend [\#631](https://github.com/containous/traefik/issues/631)
- Traefik sporadically failing when proxying requests [\#615](https://github.com/containous/traefik/issues/615)
- TCP Proxy [\#608](https://github.com/containous/traefik/issues/608)
- How to use in Windows? [\#605](https://github.com/containous/traefik/issues/605)
- `ClientCAFiles` ignored [\#604](https://github.com/containous/traefik/issues/604)
- Let`s Encrypt enable in etcd [\#600](https://github.com/containous/traefik/issues/600)
- Support HTTP Basic Auth [\#599](https://github.com/containous/traefik/issues/599)
- Consul KV seem broken [\#587](https://github.com/containous/traefik/issues/587)
- HTTPS entryPoint not working [\#574](https://github.com/containous/traefik/issues/574)
- Traefik stuck when used as frontend for a streaming API [\#560](https://github.com/containous/traefik/issues/560)
- Exclude some frontends in consul catalog [\#555](https://github.com/containous/traefik/issues/555)
- Update docs with new Mesos provider [\#548](https://github.com/containous/traefik/issues/548)
- Can I use Traefik without a domain name? [\#539](https://github.com/containous/traefik/issues/539)
- docker run syntax in swarm example has changed [\#528](https://github.com/containous/traefik/issues/528)
- Priortities in 1.0.0 not behaving [\#506](https://github.com/containous/traefik/issues/506)
- Route by path [\#500](https://github.com/containous/traefik/issues/500)
- Secure WebSockets [\#467](https://github.com/containous/traefik/issues/467)
- Container IP Lost [\#375](https://github.com/containous/traefik/issues/375)
- Multiple routes support with Docker or Marathon labels [\#118](https://github.com/containous/traefik/issues/118)

**Merged pull requests:**

- Fix path case sensitive v1.1 [\#855](https://github.com/containous/traefik/pull/855) ([emilevauge](https://github.com/emilevauge))
- Fix golint in v1.1 [\#849](https://github.com/containous/traefik/pull/849) ([emilevauge](https://github.com/emilevauge))
- Fix Kubernetes watch leak [\#845](https://github.com/containous/traefik/pull/845) ([emilevauge](https://github.com/emilevauge))
- Pass Version, Codename and Date to crosscompiled [\#842](https://github.com/containous/traefik/pull/842) ([guilhem](https://github.com/guilhem))
- Add Nvd3 Dependency to fix UI / Dashboard [\#829](https://github.com/containous/traefik/pull/829) ([SantoDE](https://github.com/SantoDE))
- Fix mkdoc theme [\#823](https://github.com/containous/traefik/pull/823) ([emilevauge](https://github.com/emilevauge))
- Prepare release v1.1.0 rc4 [\#822](https://github.com/containous/traefik/pull/822) ([emilevauge](https://github.com/emilevauge))
- Check that we serve HTTP/2 [\#820](https://github.com/containous/traefik/pull/820) ([trecloux](https://github.com/trecloux))
- Fix multiple issues [\#814](https://github.com/containous/traefik/pull/814) ([emilevauge](https://github.com/emilevauge))
- Fix ACME renew & add version check [\#783](https://github.com/containous/traefik/pull/783) ([emilevauge](https://github.com/emilevauge))
- Use first port by default [\#782](https://github.com/containous/traefik/pull/782) ([guilhem](https://github.com/guilhem))
- Prepare release v1.1.0-rc3 [\#779](https://github.com/containous/traefik/pull/779) ([emilevauge](https://github.com/emilevauge))
- Fix ResponseRecorder Flush [\#776](https://github.com/containous/traefik/pull/776) ([emilevauge](https://github.com/emilevauge))
- Use sdnotify for systemd [\#768](https://github.com/containous/traefik/pull/768) ([guilhem](https://github.com/guilhem))
- Fix providers throttle duration doc [\#760](https://github.com/containous/traefik/pull/760) ([emilevauge](https://github.com/emilevauge))
- Fix mapstructure issue with anonymous slice [\#759](https://github.com/containous/traefik/pull/759) ([emilevauge](https://github.com/emilevauge))
- Fix multiple certificates using flag [\#758](https://github.com/containous/traefik/pull/758) ([emilevauge](https://github.com/emilevauge))
- Really fix deploy ghr... [\#748](https://github.com/containous/traefik/pull/748) ([emilevauge](https://github.com/emilevauge))
- Fixes deploy ghr [\#742](https://github.com/containous/traefik/pull/742) ([emilevauge](https://github.com/emilevauge))
- prepare v1.1.0-rc2 [\#740](https://github.com/containous/traefik/pull/740) ([emilevauge](https://github.com/emilevauge))
- Fix case sensitive host [\#733](https://github.com/containous/traefik/pull/733) ([emilevauge](https://github.com/emilevauge))
- Update Kubernetes examples [\#731](https://github.com/containous/traefik/pull/731) ([Starefossen](https://github.com/Starefossen))
- fIx marathon template with dots in ID [\#728](https://github.com/containous/traefik/pull/728) ([emilevauge](https://github.com/emilevauge))
- Fix networkMap construction in ListServices [\#724](https://github.com/containous/traefik/pull/724) ([vincentlepot](https://github.com/vincentlepot))
- Add basic compatibility with marathon-lb [\#720](https://github.com/containous/traefik/pull/720) ([guilhem](https://github.com/guilhem))
- Add Ed's video at ContainerCamp [\#717](https://github.com/containous/traefik/pull/717) ([emilevauge](https://github.com/emilevauge))
- Add documentation for Træfik on docker swarm mode [\#715](https://github.com/containous/traefik/pull/715) ([vdemeester](https://github.com/vdemeester))
- Remove duplicated link to Kubernetes.io in README.md [\#713](https://github.com/containous/traefik/pull/713) ([oscerd](https://github.com/oscerd))
- Show current version in web UI [\#709](https://github.com/containous/traefik/pull/709) ([vhf](https://github.com/vhf))
- Add support for docker healthcheck 👼 [\#708](https://github.com/containous/traefik/pull/708) ([vdemeester](https://github.com/vdemeester))
- Fix syntax in Swarm example. Resolves \#528 [\#707](https://github.com/containous/traefik/pull/707) ([billglover](https://github.com/billglover))
- Add HTTP compression [\#702](https://github.com/containous/traefik/pull/702) ([tuier](https://github.com/tuier))
- Carry PR 446 - Add sticky session support \(round two!\) [\#701](https://github.com/containous/traefik/pull/701) ([emilevauge](https://github.com/emilevauge))
- Remove unused endpoint when using constraints with Marathon provider [\#697](https://github.com/containous/traefik/pull/697) ([tuier](https://github.com/tuier))
- Replace imagelayers.io with microbadger [\#696](https://github.com/containous/traefik/pull/696) ([solidnerd](https://github.com/solidnerd))
- Selectable TLS Versions [\#690](https://github.com/containous/traefik/pull/690) ([dtomcej](https://github.com/dtomcej))
- Carry pr 439 [\#689](https://github.com/containous/traefik/pull/689) ([emilevauge](https://github.com/emilevauge))
- Disable gorilla/mux URL cleaning to prevent sending redirect [\#688](https://github.com/containous/traefik/pull/688) ([ydubreuil](https://github.com/ydubreuil))
- Some fixes [\#687](https://github.com/containous/traefik/pull/687) ([emilevauge](https://github.com/emilevauge))
- feat\(constraints\): Supports constraints for Marathon provider [\#686](https://github.com/containous/traefik/pull/686) ([tuier](https://github.com/tuier))
- Update docs to improve contribution setup [\#685](https://github.com/containous/traefik/pull/685) ([dtomcej](https://github.com/dtomcej))
- Add basic auth support for web backend [\#677](https://github.com/containous/traefik/pull/677) ([SantoDE](https://github.com/SantoDE))
- Document accepted values for logLevel. [\#676](https://github.com/containous/traefik/pull/676) ([jimmycuadra](https://github.com/jimmycuadra))
- If Marathon doesn't have healthcheck, assume it's ok [\#665](https://github.com/containous/traefik/pull/665) ([gomes](https://github.com/gomes))
- ACME: renew certificates 30 days before expiry [\#660](https://github.com/containous/traefik/pull/660) ([JayH5](https://github.com/JayH5))
- Update broken link and add a comment to sample config file  [\#658](https://github.com/containous/traefik/pull/658) ([Yggdrasil](https://github.com/Yggdrasil))
- Add possibility to use BindPort IPAddress 👼 [\#657](https://github.com/containous/traefik/pull/657) ([vdemeester](https://github.com/vdemeester))
- Update marathon [\#648](https://github.com/containous/traefik/pull/648) ([emilevauge](https://github.com/emilevauge))
- Add backend features to docker [\#646](https://github.com/containous/traefik/pull/646) ([jangie](https://github.com/jangie))
- enable consul catalog to use maxconn [\#645](https://github.com/containous/traefik/pull/645) ([jangie](https://github.com/jangie))
- Adopt the Code Of Coduct from http://contributor-covenant.org [\#641](https://github.com/containous/traefik/pull/641) ([errm](https://github.com/errm))
- Use secure mode 600 instead of 644 for acme.json [\#639](https://github.com/containous/traefik/pull/639) ([discordianfish](https://github.com/discordianfish))
- docker clarification, fix dead urls, misc typos [\#637](https://github.com/containous/traefik/pull/637) ([djalal](https://github.com/djalal))
- add PING handler to dashboard API [\#630](https://github.com/containous/traefik/pull/630) ([jangie](https://github.com/jangie))
- Migrate to JobBackOff [\#628](https://github.com/containous/traefik/pull/628) ([emilevauge](https://github.com/emilevauge))
- Add long job exponential backoff [\#627](https://github.com/containous/traefik/pull/627) ([emilevauge](https://github.com/emilevauge))
- HA acme support [\#625](https://github.com/containous/traefik/pull/625) ([emilevauge](https://github.com/emilevauge))
- Bump go v1.7 [\#620](https://github.com/containous/traefik/pull/620) ([emilevauge](https://github.com/emilevauge))
- Make duration logging consistent [\#619](https://github.com/containous/traefik/pull/619) ([jangie](https://github.com/jangie))
- fix for nil clientTLS causing issue [\#617](https://github.com/containous/traefik/pull/617) ([jangie](https://github.com/jangie))
- Add ability for marathon provider to set maxconn values, loadbalancer algorithm, and circuit breaker expression [\#616](https://github.com/containous/traefik/pull/616) ([jangie](https://github.com/jangie))
- Make systemd unit installable [\#613](https://github.com/containous/traefik/pull/613) ([keis](https://github.com/keis))
- Merge v1.0.2 master [\#610](https://github.com/containous/traefik/pull/610) ([emilevauge](https://github.com/emilevauge))
- update staert and flaeg [\#609](https://github.com/containous/traefik/pull/609) ([cocap10](https://github.com/cocap10))
- \#504 Initial support for Docker 1.12 Swarm Mode [\#602](https://github.com/containous/traefik/pull/602) ([diegofernandes](https://github.com/diegofernandes))
- Add Host cert ACME generation [\#601](https://github.com/containous/traefik/pull/601) ([emilevauge](https://github.com/emilevauge))
- Fixed binary script so traefik version command doesn't just print default values [\#598](https://github.com/containous/traefik/pull/598) ([keiths-osc](https://github.com/keiths-osc))
- Name servers after thier pods [\#596](https://github.com/containous/traefik/pull/596) ([errm](https://github.com/errm))
- Fix Consul prefix [\#589](https://github.com/containous/traefik/pull/589) ([jippi](https://github.com/jippi))
- Prioritize kubernetes routes by path length [\#588](https://github.com/containous/traefik/pull/588) ([philk](https://github.com/philk))
- beautify help [\#580](https://github.com/containous/traefik/pull/580) ([cocap10](https://github.com/cocap10))
- Upgrade directives name since we use angular-ui-bootstrap [\#578](https://github.com/containous/traefik/pull/578) ([micaelmbagira](https://github.com/micaelmbagira))
- Fix basic docs for configuration of multiple rules [\#576](https://github.com/containous/traefik/pull/576) ([ajaegle](https://github.com/ajaegle))
- Fix k8s watch [\#573](https://github.com/containous/traefik/pull/573) ([errm](https://github.com/errm))
- Add requirements.txt for netlify [\#567](https://github.com/containous/traefik/pull/567) ([emilevauge](https://github.com/emilevauge))
- Merge v1.0.1 master [\#565](https://github.com/containous/traefik/pull/565) ([emilevauge](https://github.com/emilevauge))
-  Move webui to FountainJS with Webpack [\#558](https://github.com/containous/traefik/pull/558) ([micaelmbagira](https://github.com/micaelmbagira))
- Add global InsecureSkipVerify option to disable certificate checking [\#557](https://github.com/containous/traefik/pull/557) ([stuart-c](https://github.com/stuart-c))
- Move version.go in its own package… [\#553](https://github.com/containous/traefik/pull/553) ([vdemeester](https://github.com/vdemeester))
- Upgrade libkermit and dependencies [\#552](https://github.com/containous/traefik/pull/552) ([vdemeester](https://github.com/vdemeester))
- Add command storeconfig [\#551](https://github.com/containous/traefik/pull/551) ([cocap10](https://github.com/cocap10))
- Add basic/digest auth [\#547](https://github.com/containous/traefik/pull/547) ([emilevauge](https://github.com/emilevauge))
- Bump node to 6 for webui [\#546](https://github.com/containous/traefik/pull/546) ([vdemeester](https://github.com/vdemeester))
- Bump golang to 1.6.3 [\#545](https://github.com/containous/traefik/pull/545) ([vdemeester](https://github.com/vdemeester))
- Fix typos [\#538](https://github.com/containous/traefik/pull/538) ([jimt](https://github.com/jimt))
- Kubernetes user-guide [\#519](https://github.com/containous/traefik/pull/519) ([errm](https://github.com/errm))
- Implement Kubernetes Selectors, minor kube endpoint fix [\#516](https://github.com/containous/traefik/pull/516) ([pnegahdar](https://github.com/pnegahdar))
- Carry \#358 : Option to disable expose of all docker containers [\#514](https://github.com/containous/traefik/pull/514) ([vdemeester](https://github.com/vdemeester))
- Remove traefik.frontend.value support in docker… [\#510](https://github.com/containous/traefik/pull/510) ([vdemeester](https://github.com/vdemeester))
- Use KvStores as global config sources [\#481](https://github.com/containous/traefik/pull/481) ([cocap10](https://github.com/cocap10))
- Add endpoint option to authenticate by client tls cert. [\#461](https://github.com/containous/traefik/pull/461) ([andersbetner](https://github.com/andersbetner))
- add mesos provider inspired by mesos-dns & marathon provider [\#353](https://github.com/containous/traefik/pull/353) ([skydjol](https://github.com/skydjol))

## [v1.1.0-rc4](https://github.com/containous/traefik/tree/v1.1.0-rc4) (2016-11-10)
[Full Changelog](https://github.com/containous/traefik/compare/v1.1.0-rc3...v1.1.0-rc4)

**Implemented enhancements:**

- Feature Request: Enable Health checks to containers. [\#540](https://github.com/containous/traefik/issues/540)

**Fixed bugs:**

- Traefik stopped serving on upgrade to v1.1.0-rc3 [\#807](https://github.com/containous/traefik/issues/807)
- Traefik cannot read constraints from KV [\#794](https://github.com/containous/traefik/issues/794)
- HTTP2 - configuration [\#790](https://github.com/containous/traefik/issues/790)
- Allow multiple certificates on a single entrypoint when trying to use TLS? [\#747](https://github.com/containous/traefik/issues/747)

**Closed issues:**

- LoadBalancing doesn't work in single node Swarm-mode [\#815](https://github.com/containous/traefik/issues/815)
- cannot connect to docker daemon [\#813](https://github.com/containous/traefik/issues/813)
- Let's encrypt configuration not working [\#805](https://github.com/containous/traefik/issues/805)
- Question: Wildcard Host for Kubernetes Ingress [\#792](https://github.com/containous/traefik/issues/792)
- Multiple subdomains for Marathon backend. [\#785](https://github.com/containous/traefik/issues/785)
- traefik-1.1.0-rc1: build error [\#781](https://github.com/containous/traefik/issues/781)
- Multiple routes support with Docker or Marathon labels [\#118](https://github.com/containous/traefik/issues/118)

**Merged pull requests:**

- Prepare release v1.1.0 rc4 [\#822](https://github.com/containous/traefik/pull/822) ([emilevauge](https://github.com/emilevauge))
- Fix multiple issues [\#814](https://github.com/containous/traefik/pull/814) ([emilevauge](https://github.com/emilevauge))
- Fix ACME renew & add version check [\#783](https://github.com/containous/traefik/pull/783) ([emilevauge](https://github.com/emilevauge))
- Use first port by default [\#782](https://github.com/containous/traefik/pull/782) ([guilhem](https://github.com/guilhem))

## [v1.1.0-rc3](https://github.com/containous/traefik/tree/v1.1.0-rc3) (2016-10-26)
[Full Changelog](https://github.com/containous/traefik/compare/v1.1.0-rc2...v1.1.0-rc3)

**Fixed bugs:**

- Cannot provide multiple certificates using flag [\#757](https://github.com/containous/traefik/issues/757)
- traefik \* Users: unsupported type: slice [\#743](https://github.com/containous/traefik/issues/743)
- \[Docker swarm mode\] The traefik.docker.network seems to have no effect [\#719](https://github.com/containous/traefik/issues/719)
- Case sensitive domain names breaks routing  [\#562](https://github.com/containous/traefik/issues/562)

**Closed issues:**

- dependencies installation error [\#755](https://github.com/containous/traefik/issues/755)
- k8s provider w/ acme? [\#752](https://github.com/containous/traefik/issues/752)
- Documented ProvidersThrottleDuration value is invalid [\#741](https://github.com/containous/traefik/issues/741)
- Loadbalaning issues with traefik and Docker Swarm cluster [\#730](https://github.com/containous/traefik/issues/730)
- issues with marathon app ids containing a dot [\#726](https://github.com/containous/traefik/issues/726)
- How Routing traffic depending on path not domain in docker [\#706](https://github.com/containous/traefik/issues/706)
- Traefik crashes when using Consul catalog [\#699](https://github.com/containous/traefik/issues/699)
- File Watcher for rules does not work  [\#683](https://github.com/containous/traefik/issues/683)

**Merged pull requests:**

- Fix ResponseRecorder Flush [\#776](https://github.com/containous/traefik/pull/776) ([emilevauge](https://github.com/emilevauge))
- Use sdnotify for systemd [\#768](https://github.com/containous/traefik/pull/768) ([guilhem](https://github.com/guilhem))
- Fix providers throttle duration doc [\#760](https://github.com/containous/traefik/pull/760) ([emilevauge](https://github.com/emilevauge))
- Fix mapstructure issue with anonymous slice [\#759](https://github.com/containous/traefik/pull/759) ([emilevauge](https://github.com/emilevauge))
- Fix multiple certificates using flag [\#758](https://github.com/containous/traefik/pull/758) ([emilevauge](https://github.com/emilevauge))
- Really fix deploy ghr... [\#748](https://github.com/containous/traefik/pull/748) ([emilevauge](https://github.com/emilevauge))

## [v1.1.0-rc2](https://github.com/containous/traefik/tree/v1.1.0-rc2) (2016-10-17)
[Full Changelog](https://github.com/containous/traefik/compare/v1.1.0-rc1...v1.1.0-rc2)

**Implemented enhancements:**

- Support healthcheck if present for docker [\#666](https://github.com/containous/traefik/issues/666)

**Closed issues:**

- Sensible configuration for consulCatalog [\#737](https://github.com/containous/traefik/issues/737)
- Traefik ignoring container listening in more than one TCP port [\#734](https://github.com/containous/traefik/issues/734)
- Error when using HA acme in kubernetes with etcd [\#725](https://github.com/containous/traefik/issues/725)
- \[Docker swarm mode\] No round robin when using service [\#718](https://github.com/containous/traefik/issues/718)
- Dose it support docker swarm mode  [\#712](https://github.com/containous/traefik/issues/712)
- Kubernetes - Undefined backend  [\#710](https://github.com/containous/traefik/issues/710)
- Constraints on Consul Catalogue not working as expected [\#703](https://github.com/containous/traefik/issues/703)
- docker run syntax in swarm example has changed [\#528](https://github.com/containous/traefik/issues/528)
- Secure WebSockets [\#467](https://github.com/containous/traefik/issues/467)

**Merged pull requests:**

- Fix case sensitive host [\#733](https://github.com/containous/traefik/pull/733) ([emilevauge](https://github.com/emilevauge))
- Update Kubernetes examples [\#731](https://github.com/containous/traefik/pull/731) ([Starefossen](https://github.com/Starefossen))
- fIx marathon template with dots in ID [\#728](https://github.com/containous/traefik/pull/728) ([emilevauge](https://github.com/emilevauge))
- Fix networkMap construction in ListServices [\#724](https://github.com/containous/traefik/pull/724) ([vincentlepot](https://github.com/vincentlepot))
- Add basic compatibility with marathon-lb [\#720](https://github.com/containous/traefik/pull/720) ([guilhem](https://github.com/guilhem))
- Add Ed's video at ContainerCamp [\#717](https://github.com/containous/traefik/pull/717) ([emilevauge](https://github.com/emilevauge))
- Add documentation for Træfik on docker swarm mode [\#715](https://github.com/containous/traefik/pull/715) ([vdemeester](https://github.com/vdemeester))
- Remove duplicated link to Kubernetes.io in README.md [\#713](https://github.com/containous/traefik/pull/713) ([oscerd](https://github.com/oscerd))
- Show current version in web UI [\#709](https://github.com/containous/traefik/pull/709) ([vhf](https://github.com/vhf))
- Add support for docker healthcheck 👼 [\#708](https://github.com/containous/traefik/pull/708) ([vdemeester](https://github.com/vdemeester))
- Fix syntax in Swarm example. Resolves \#528 [\#707](https://github.com/containous/traefik/pull/707) ([billglover](https://github.com/billglover))

## [v1.1.0-rc1](https://github.com/containous/traefik/tree/v1.1.0-rc1) (2016-09-30)
[Full Changelog](https://github.com/containous/traefik/compare/v1.0.0...v1.1.0-rc1)

**Implemented enhancements:**

- Feature Request: SSL Cipher Selection [\#535](https://github.com/containous/traefik/issues/535)
- Error with -consulcatalog and missing load balance method on 1.0.0 [\#524](https://github.com/containous/traefik/issues/524)
- Running Traefik with Docker 1.12 Swarm Mode [\#504](https://github.com/containous/traefik/issues/504)
- Kubernetes provider: should allow the master url to be override [\#501](https://github.com/containous/traefik/issues/501)
- \[FRONTEND\]\[LE\] Pre-generate SSL certificates for "Host:" rules [\#483](https://github.com/containous/traefik/issues/483)
- Frontend Rule evolution [\#437](https://github.com/containous/traefik/issues/437)
- Add a Changelog [\#388](https://github.com/containous/traefik/issues/388)
- Add label matching for kubernetes ingests [\#363](https://github.com/containous/traefik/issues/363)
- Acme in HA Traefik Scenario [\#348](https://github.com/containous/traefik/issues/348)
- HTTP Basic Auth support [\#77](https://github.com/containous/traefik/issues/77)
- Session affinity / stickiness / persistence [\#5](https://github.com/containous/traefik/issues/5)
- Kubernetes provider: traefik.frontend.rule.type logging [\#668](https://github.com/containous/traefik/pull/668) ([yvespp](https://github.com/yvespp))

**Fixed bugs:**

- traefik hangs - stops handling requests [\#662](https://github.com/containous/traefik/issues/662)
- Add long jobs in exponential backoff providers  [\#626](https://github.com/containous/traefik/issues/626)
- Tip of tree crashes on invalid pointer on Marathon provider [\#624](https://github.com/containous/traefik/issues/624)
- ACME: revoke certificate on agreement update [\#579](https://github.com/containous/traefik/issues/579)
- WebUI: Providers tabs disappeared [\#577](https://github.com/containous/traefik/issues/577)
- traefik version command contains incorrect information when building from master branch [\#569](https://github.com/containous/traefik/issues/569)
- Flag --etcd.endpoint default [\#508](https://github.com/containous/traefik/issues/508)
- Conditional ACME on demand generation [\#505](https://github.com/containous/traefik/issues/505)
- Important delay with streams \(Mozilla EventSource\) [\#503](https://github.com/containous/traefik/issues/503)
- Traefik crashing [\#458](https://github.com/containous/traefik/issues/458)
- traefik.toml constraints error: `Expected map but found 'string'.` [\#451](https://github.com/containous/traefik/issues/451)
- Multiple path separators in the url path causing redirect [\#167](https://github.com/containous/traefik/issues/167)

**Closed issues:**

- Global InsecureSkipVerify does not work [\#700](https://github.com/containous/traefik/issues/700)
- \[documentation/feature\] Consul/etcd support atomic multiple key changes now [\#698](https://github.com/containous/traefik/issues/698)
- How to configure which network to use when starting traefik binary? [\#694](https://github.com/containous/traefik/issues/694)
- How to get multiple host headers working for docker labels? [\#692](https://github.com/containous/traefik/issues/692)
- Requests with URL-encoded characters are not forwarded correctly [\#684](https://github.com/containous/traefik/issues/684)
- Issue with global InsecureSkipVerify = true and self signed certificates [\#667](https://github.com/containous/traefik/issues/667)
- Docker exposedbydefault = false didn't work [\#663](https://github.com/containous/traefik/issues/663)
- \[ACME\] Auto SAN Detection [\#655](https://github.com/containous/traefik/issues/655)
- Fronting a domain with DNS A-record round-robin & ACME [\#654](https://github.com/containous/traefik/issues/654)
- Overriding toml configuration with environment variables [\#650](https://github.com/containous/traefik/issues/650)
- marathon provider exposedByDefault = false [\#647](https://github.com/containous/traefik/issues/647)
- Add status URL for service up checks [\#642](https://github.com/containous/traefik/issues/642)
- acme's storage file, containing private key, is word readable [\#638](https://github.com/containous/traefik/issues/638)
- wildcard domain with exclusions [\#633](https://github.com/containous/traefik/issues/633)
- Enable evenly distribution among backend [\#631](https://github.com/containous/traefik/issues/631)
- Traefik sporadically failing when proxying requests [\#615](https://github.com/containous/traefik/issues/615)
- TCP Proxy [\#608](https://github.com/containous/traefik/issues/608)
- How to use in Windows? [\#605](https://github.com/containous/traefik/issues/605)
- `ClientCAFiles` ignored [\#604](https://github.com/containous/traefik/issues/604)
- Let`s Encrypt enable in etcd [\#600](https://github.com/containous/traefik/issues/600)
- Support HTTP Basic Auth [\#599](https://github.com/containous/traefik/issues/599)
- Consul KV seem broken [\#587](https://github.com/containous/traefik/issues/587)
- HTTPS entryPoint not working [\#574](https://github.com/containous/traefik/issues/574)
- Traefik stuck when used as frontend for a streaming API [\#560](https://github.com/containous/traefik/issues/560)
- Exclude some frontends in consul catalog [\#555](https://github.com/containous/traefik/issues/555)
- Can I use Traefik without a domain name? [\#539](https://github.com/containous/traefik/issues/539)
- Priortities in 1.0.0 not behaving [\#506](https://github.com/containous/traefik/issues/506)
- Route by path [\#500](https://github.com/containous/traefik/issues/500)
- Container IP Lost [\#375](https://github.com/containous/traefik/issues/375)

**Merged pull requests:**

- Add HTTP compression [\#702](https://github.com/containous/traefik/pull/702) ([tuier](https://github.com/tuier))
- Carry PR 446 - Add sticky session support \(round two!\) [\#701](https://github.com/containous/traefik/pull/701) ([emilevauge](https://github.com/emilevauge))
- Remove unused endpoint when using constraints with Marathon provider [\#697](https://github.com/containous/traefik/pull/697) ([tuier](https://github.com/tuier))
- Replace imagelayers.io with microbadger [\#696](https://github.com/containous/traefik/pull/696) ([solidnerd](https://github.com/solidnerd))
- Selectable TLS Versions [\#690](https://github.com/containous/traefik/pull/690) ([dtomcej](https://github.com/dtomcej))
- Carry pr 439 [\#689](https://github.com/containous/traefik/pull/689) ([emilevauge](https://github.com/emilevauge))
- Disable gorilla/mux URL cleaning to prevent sending redirect [\#688](https://github.com/containous/traefik/pull/688) ([ydubreuil](https://github.com/ydubreuil))
- Some fixes [\#687](https://github.com/containous/traefik/pull/687) ([emilevauge](https://github.com/emilevauge))
- feat\(constraints\): Supports constraints for Marathon provider [\#686](https://github.com/containous/traefik/pull/686) ([tuier](https://github.com/tuier))
- Update docs to improve contribution setup [\#685](https://github.com/containous/traefik/pull/685) ([dtomcej](https://github.com/dtomcej))
- Add basic auth support for web backend [\#677](https://github.com/containous/traefik/pull/677) ([SantoDE](https://github.com/SantoDE))
- Document accepted values for logLevel. [\#676](https://github.com/containous/traefik/pull/676) ([jimmycuadra](https://github.com/jimmycuadra))
- If Marathon doesn't have healthcheck, assume it's ok [\#665](https://github.com/containous/traefik/pull/665) ([gomes](https://github.com/gomes))
- ACME: renew certificates 30 days before expiry [\#660](https://github.com/containous/traefik/pull/660) ([JayH5](https://github.com/JayH5))
- Update broken link and add a comment to sample config file  [\#658](https://github.com/containous/traefik/pull/658) ([Yggdrasil](https://github.com/Yggdrasil))
- Add possibility to use BindPort IPAddress 👼 [\#657](https://github.com/containous/traefik/pull/657) ([vdemeester](https://github.com/vdemeester))
- Update marathon [\#648](https://github.com/containous/traefik/pull/648) ([emilevauge](https://github.com/emilevauge))
- Add backend features to docker [\#646](https://github.com/containous/traefik/pull/646) ([jangie](https://github.com/jangie))
- enable consul catalog to use maxconn [\#645](https://github.com/containous/traefik/pull/645) ([jangie](https://github.com/jangie))
- Adopt the Code Of Coduct from http://contributor-covenant.org [\#641](https://github.com/containous/traefik/pull/641) ([errm](https://github.com/errm))
- Use secure mode 600 instead of 644 for acme.json [\#639](https://github.com/containous/traefik/pull/639) ([discordianfish](https://github.com/discordianfish))
- docker clarification, fix dead urls, misc typos [\#637](https://github.com/containous/traefik/pull/637) ([djalal](https://github.com/djalal))
- add PING handler to dashboard API [\#630](https://github.com/containous/traefik/pull/630) ([jangie](https://github.com/jangie))
- Migrate to JobBackOff [\#628](https://github.com/containous/traefik/pull/628) ([emilevauge](https://github.com/emilevauge))
- Add long job exponential backoff [\#627](https://github.com/containous/traefik/pull/627) ([emilevauge](https://github.com/emilevauge))
- HA acme support [\#625](https://github.com/containous/traefik/pull/625) ([emilevauge](https://github.com/emilevauge))
- Bump go v1.7 [\#620](https://github.com/containous/traefik/pull/620) ([emilevauge](https://github.com/emilevauge))
- Make duration logging consistent [\#619](https://github.com/containous/traefik/pull/619) ([jangie](https://github.com/jangie))
- fix for nil clientTLS causing issue [\#617](https://github.com/containous/traefik/pull/617) ([jangie](https://github.com/jangie))
- Add ability for marathon provider to set maxconn values, loadbalancer algorithm, and circuit breaker expression [\#616](https://github.com/containous/traefik/pull/616) ([jangie](https://github.com/jangie))
- Make systemd unit installable [\#613](https://github.com/containous/traefik/pull/613) ([keis](https://github.com/keis))
- Merge v1.0.2 master [\#610](https://github.com/containous/traefik/pull/610) ([emilevauge](https://github.com/emilevauge))
- update staert and flaeg [\#609](https://github.com/containous/traefik/pull/609) ([cocap10](https://github.com/cocap10))
- \#504 Initial support for Docker 1.12 Swarm Mode [\#602](https://github.com/containous/traefik/pull/602) ([diegofernandes](https://github.com/diegofernandes))
- Add Host cert ACME generation [\#601](https://github.com/containous/traefik/pull/601) ([emilevauge](https://github.com/emilevauge))
- Fixed binary script so traefik version command doesn't just print default values [\#598](https://github.com/containous/traefik/pull/598) ([keiths-osc](https://github.com/keiths-osc))
- Name servers after thier pods [\#596](https://github.com/containous/traefik/pull/596) ([errm](https://github.com/errm))
- Fix Consul prefix [\#589](https://github.com/containous/traefik/pull/589) ([jippi](https://github.com/jippi))
- Prioritize kubernetes routes by path length [\#588](https://github.com/containous/traefik/pull/588) ([philk](https://github.com/philk))
- beautify help [\#580](https://github.com/containous/traefik/pull/580) ([cocap10](https://github.com/cocap10))
- Upgrade directives name since we use angular-ui-bootstrap [\#578](https://github.com/containous/traefik/pull/578) ([micaelmbagira](https://github.com/micaelmbagira))
- Fix basic docs for configuration of multiple rules [\#576](https://github.com/containous/traefik/pull/576) ([ajaegle](https://github.com/ajaegle))
- Fix k8s watch [\#573](https://github.com/containous/traefik/pull/573) ([errm](https://github.com/errm))
- Add requirements.txt for netlify [\#567](https://github.com/containous/traefik/pull/567) ([emilevauge](https://github.com/emilevauge))
- Merge v1.0.1 master [\#565](https://github.com/containous/traefik/pull/565) ([emilevauge](https://github.com/emilevauge))
-  Move webui to FountainJS with Webpack [\#558](https://github.com/containous/traefik/pull/558) ([micaelmbagira](https://github.com/micaelmbagira))
- Add global InsecureSkipVerify option to disable certificate checking [\#557](https://github.com/containous/traefik/pull/557) ([stuart-c](https://github.com/stuart-c))
- Move version.go in its own package… [\#553](https://github.com/containous/traefik/pull/553) ([vdemeester](https://github.com/vdemeester))
- Upgrade libkermit and dependencies [\#552](https://github.com/containous/traefik/pull/552) ([vdemeester](https://github.com/vdemeester))
- Add command storeconfig [\#551](https://github.com/containous/traefik/pull/551) ([cocap10](https://github.com/cocap10))
- Add basic/digest auth [\#547](https://github.com/containous/traefik/pull/547) ([emilevauge](https://github.com/emilevauge))
- Bump node to 6 for webui [\#546](https://github.com/containous/traefik/pull/546) ([vdemeester](https://github.com/vdemeester))
- Bump golang to 1.6.3 [\#545](https://github.com/containous/traefik/pull/545) ([vdemeester](https://github.com/vdemeester))
- Fix typos [\#538](https://github.com/containous/traefik/pull/538) ([jimt](https://github.com/jimt))
- Kubernetes user-guide [\#519](https://github.com/containous/traefik/pull/519) ([errm](https://github.com/errm))
- Implement Kubernetes Selectors, minor kube endpoint fix [\#516](https://github.com/containous/traefik/pull/516) ([pnegahdar](https://github.com/pnegahdar))
- Carry \#358 : Option to disable expose of all docker containers [\#514](https://github.com/containous/traefik/pull/514) ([vdemeester](https://github.com/vdemeester))
- Remove traefik.frontend.value support in docker… [\#510](https://github.com/containous/traefik/pull/510) ([vdemeester](https://github.com/vdemeester))
- Use KvStores as global config sources [\#481](https://github.com/containous/traefik/pull/481) ([cocap10](https://github.com/cocap10))
- Add endpoint option to authenticate by client tls cert. [\#461](https://github.com/containous/traefik/pull/461) ([andersbetner](https://github.com/andersbetner))
- add mesos provider inspired by mesos-dns & marathon provider [\#353](https://github.com/containous/traefik/pull/353) ([skydjol](https://github.com/skydjol))

## [v1.0.2](https://github.com/containous/traefik/tree/v1.0.2) (2016-08-02)
[Full Changelog](https://github.com/containous/traefik/compare/v1.0.1...v1.0.2)

**Fixed bugs:**

- ACME: revoke certificate on agreement update [\#579](https://github.com/containous/traefik/issues/579)

**Closed issues:**

- Exclude some frontends in consul catalog [\#555](https://github.com/containous/traefik/issues/555)

**Merged pull requests:**

- Bump oxy version, fix streaming [\#584](https://github.com/containous/traefik/pull/584) ([emilevauge](https://github.com/emilevauge))
- Fix ACME TOS [\#582](https://github.com/containous/traefik/pull/582) ([emilevauge](https://github.com/emilevauge))

## [v1.0.1](https://github.com/containous/traefik/tree/v1.0.1) (2016-07-19)
[Full Changelog](https://github.com/containous/traefik/compare/v1.0.0...v1.0.1)

**Implemented enhancements:**

- Error with -consulcatalog and missing load balance method on 1.0.0 [\#524](https://github.com/containous/traefik/issues/524)
- Kubernetes provider: should allow the master url to be override [\#501](https://github.com/containous/traefik/issues/501)

**Fixed bugs:**

- Flag --etcd.endpoint default [\#508](https://github.com/containous/traefik/issues/508)
- Conditional ACME on demand generation [\#505](https://github.com/containous/traefik/issues/505)
- Important delay with streams \(Mozilla EventSource\) [\#503](https://github.com/containous/traefik/issues/503)

**Closed issues:**

- Can I use Traefik without a domain name? [\#539](https://github.com/containous/traefik/issues/539)
- Priortities in 1.0.0 not behaving [\#506](https://github.com/containous/traefik/issues/506)
- Route by path [\#500](https://github.com/containous/traefik/issues/500)

**Merged pull requests:**

- Update server.go [\#531](https://github.com/containous/traefik/pull/531) ([Jsewill](https://github.com/Jsewill))
- Add sse support [\#527](https://github.com/containous/traefik/pull/527) ([emilevauge](https://github.com/emilevauge))
- Fix acme checkOnDemandDomain [\#512](https://github.com/containous/traefik/pull/512) ([emilevauge](https://github.com/emilevauge))
- Fix default etcd port [\#511](https://github.com/containous/traefik/pull/511) ([errm](https://github.com/errm))

## [v1.0.0](https://github.com/containous/traefik/tree/v1.0.0) (2016-07-05)
[Full Changelog](https://github.com/containous/traefik/compare/v1.0.0-rc3...v1.0.0)

**Fixed bugs:**

- Enable to define empty TLS option by flag for Let's Encrypt [\#488](https://github.com/containous/traefik/issues/488)
- \[Docker\] No IP in backend in host networking mode [\#487](https://github.com/containous/traefik/issues/487)
- Response is compressed when not requested [\#485](https://github.com/containous/traefik/issues/485)
- loadConfig modifies configuration causing same config check to fail [\#480](https://github.com/containous/traefik/issues/480)

**Closed issues:**

- svg logo [\#482](https://github.com/containous/traefik/issues/482)
- etcd tries to connect with TLS even with --etcd.tls=false [\#456](https://github.com/containous/traefik/issues/456)
- Zookeeper - KV connection error: Failed to test KV store connection [\#455](https://github.com/containous/traefik/issues/455)
- "Not Found" api response needed instead of 404  [\#454](https://github.com/containous/traefik/issues/454)
- domain label doesn't work on docker [\#447](https://github.com/containous/traefik/issues/447)
- Any chance of a windows release? [\#425](https://github.com/containous/traefik/issues/425)

**Merged pull requests:**

- Fix windows builds [\#495](https://github.com/containous/traefik/pull/495) ([emilevauge](https://github.com/emilevauge))
- Fix host Docker network [\#494](https://github.com/containous/traefik/pull/494) ([emilevauge](https://github.com/emilevauge))
- Fix empty tls flag [\#493](https://github.com/containous/traefik/pull/493) ([emilevauge](https://github.com/emilevauge))
- Fix webui proxying [\#492](https://github.com/containous/traefik/pull/492) ([emilevauge](https://github.com/emilevauge))
- Fix default weight in server.LoadConfig [\#491](https://github.com/containous/traefik/pull/491) ([emilevauge](https://github.com/emilevauge))
- Fix retry headers, simplify ResponseRecorder [\#490](https://github.com/containous/traefik/pull/490) ([emilevauge](https://github.com/emilevauge))

## [v1.0.0-rc3](https://github.com/containous/traefik/tree/v1.0.0-rc3) (2016-06-23)
[Full Changelog](https://github.com/containous/traefik/compare/v1.0.0-rc2...v1.0.0-rc3)

**Implemented enhancements:**

- support more than one rule to Docker backend [\#419](https://github.com/containous/traefik/issues/419)

**Fixed bugs:**

- consulCatalog issue when serviceName contains a dot [\#475](https://github.com/containous/traefik/issues/475)
- Issue with empty responses [\#463](https://github.com/containous/traefik/issues/463)
- Severe memory leak in beta.470 and beyond crashes Traefik server  [\#462](https://github.com/containous/traefik/issues/462)
- Marathon that starts with a space causes parsing errors. [\#459](https://github.com/containous/traefik/issues/459)
- A frontend route without a rule \(or empty rule\) causes a crash when traefik starts [\#453](https://github.com/containous/traefik/issues/453)
- container dropped out when connecting to Docker Swarm [\#442](https://github.com/containous/traefik/issues/442)
- Traefik setting Accept-Encoding: gzip on requests \(Traefik may also be broken with chunked responses\) [\#421](https://github.com/containous/traefik/issues/421)

**Closed issues:**

- HTTP headers case gets modified [\#466](https://github.com/containous/traefik/issues/466)
- File frontend \> Marathon Backend [\#465](https://github.com/containous/traefik/issues/465)
- Websocket: Unable to hijack the connection [\#452](https://github.com/containous/traefik/issues/452)
- kubernetes: Received event spamming? [\#449](https://github.com/containous/traefik/issues/449)
- kubernetes: backends not updated when i scale replication controller? [\#448](https://github.com/containous/traefik/issues/448)
- Add href link on frontend [\#436](https://github.com/containous/traefik/issues/436)
- Multiple Domains Rule [\#430](https://github.com/containous/traefik/issues/430)

**Merged pull requests:**

- Disable constraints in doc until 1.1 [\#479](https://github.com/containous/traefik/pull/479) ([emilevauge](https://github.com/emilevauge))
- Sort nodes before creating consul catalog config [\#478](https://github.com/containous/traefik/pull/478) ([keis](https://github.com/keis))
- Fix spamming events in listenProviders [\#477](https://github.com/containous/traefik/pull/477) ([emilevauge](https://github.com/emilevauge))
- Fix empty responses [\#476](https://github.com/containous/traefik/pull/476) ([emilevauge](https://github.com/emilevauge))
- Fix acme renew [\#472](https://github.com/containous/traefik/pull/472) ([emilevauge](https://github.com/emilevauge))
- Fix typo in error message. [\#471](https://github.com/containous/traefik/pull/471) ([KevinBusse](https://github.com/KevinBusse))
- Fix errors load config [\#470](https://github.com/containous/traefik/pull/470) ([emilevauge](https://github.com/emilevauge))
- Typo: Replace French words by English ones [\#469](https://github.com/containous/traefik/pull/469) ([kumy](https://github.com/kumy))
- Fix marathon TLS/basic auth [\#468](https://github.com/containous/traefik/pull/468) ([emilevauge](https://github.com/emilevauge))
- Fix memory leak in listenProviders [\#464](https://github.com/containous/traefik/pull/464) ([emilevauge](https://github.com/emilevauge))
- Fix websocket connection Hijack [\#460](https://github.com/containous/traefik/pull/460) ([emilevauge](https://github.com/emilevauge))
- Fix default KV configuration [\#450](https://github.com/containous/traefik/pull/450) ([emilevauge](https://github.com/emilevauge))
- Fix panic if listContainers fails… [\#443](https://github.com/containous/traefik/pull/443) ([vdemeester](https://github.com/vdemeester))
- mount acme folder instead of file [\#441](https://github.com/containous/traefik/pull/441) ([NicolasGeraud](https://github.com/NicolasGeraud))
- feat\(constraints\): Supports constraints for docker backend [\#438](https://github.com/containous/traefik/pull/438) ([samber](https://github.com/samber))

## [v1.0.0-rc2](https://github.com/containous/traefik/tree/v1.0.0-rc2) (2016-06-07)
[Full Changelog](https://github.com/containous/traefik/compare/v1.0.0-rc1...v1.0.0-rc2)

**Implemented enhancements:**

- Add @samber to maintainers [\#440](https://github.com/containous/traefik/pull/440) ([emilevauge](https://github.com/emilevauge))

**Fixed bugs:**

- Panic on help [\#429](https://github.com/containous/traefik/issues/429)
- Bad default values in configuration [\#427](https://github.com/containous/traefik/issues/427)

**Closed issues:**

- Traefik doesn't listen on IPv4 ports [\#434](https://github.com/containous/traefik/issues/434)
- Not listening on port 80 [\#432](https://github.com/containous/traefik/issues/432)
- docs need updating for new frontend rules format [\#423](https://github.com/containous/traefik/issues/423)
- Does traefik supports for Mac? \(For devlelopment\)  [\#417](https://github.com/containous/traefik/issues/417)

**Merged pull requests:**

- Allow multiple rules [\#435](https://github.com/containous/traefik/pull/435) ([fclaeys](https://github.com/fclaeys))
- Add routes priorities [\#433](https://github.com/containous/traefik/pull/433) ([emilevauge](https://github.com/emilevauge))
- Fix default configuration [\#428](https://github.com/containous/traefik/pull/428) ([emilevauge](https://github.com/emilevauge))
- Fix marathon groups subdomain [\#426](https://github.com/containous/traefik/pull/426) ([emilevauge](https://github.com/emilevauge))
- Fix travis tag check [\#422](https://github.com/containous/traefik/pull/422) ([emilevauge](https://github.com/emilevauge))
- log info about TOML configuration file using [\#420](https://github.com/containous/traefik/pull/420) ([cocap10](https://github.com/cocap10))
- Doc about skipping some integration tests with '-check.f ConsulCatalogSuite' [\#418](https://github.com/containous/traefik/pull/418) ([samber](https://github.com/samber))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*