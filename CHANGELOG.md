# Change Log

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