# Documentation

etcd is a distributed key-value store designed to reliably and quickly preserve and provide access to critical data. It enables reliable distributed coordination through distributed locking, leader elections, and write barriers. An etcd cluster is intended for high availability and permanent data storage and retrieval.

## Getting started

New etcd users and developers should get started by [downloading and building][download_build] etcd. After getting etcd, follow this [quick demo][demo] to see the basics of creating and working with an etcd cluster.

## Developing with etcd

The easiest way to get started using etcd as a distributed key-value store is to [set up a local cluster][local_cluster].

 - [Setting up local clusters][local_cluster]
 - [Interacting with etcd][interacting]
 - gRPC [etcd core][api_ref] and [etcd concurrency][api_concurrency_ref] API references
 - [HTTP JSON API through the gRPC gateway][api_grpc_gateway]
 - [gRPC naming and discovery][grpc_naming]
 - [Client][namespace_client] and [proxy][namespace_proxy] namespacing
 - [Embedding etcd][embed_etcd]
 - [Experimental features and APIs][experimental]
 - [System limits][system-limit]

## Operating etcd clusters

Administrators who need to create reliable and scalable key-value stores for the developers they support should begin with a [cluster on multiple machines][clustering].

 - [Setting up etcd clusters][clustering]
 - [Setting up etcd gateways][gateway]
 - [Setting up etcd gRPC proxy][grpc_proxy]
 - [Hardware recommendations][hardware]
 - [Configuration][conf]
 - [Security][security]
 - [Authentication][authentication]
 - [Monitoring][monitoring]
 - [Maintenance][maintenance]
 - [Understand failures][failures]
 - [Disaster recovery][recovery]
 - [Performance][performance]
 - [Versioning][versioning]

### Platform guides

 - [Supported systems][supported_platforms]
 - [Docker container][container_docker]
 - [Container Linux, systemd][container_linux_platform]
 - [rkt container][container_rkt]
 - [Amazon Web Services][aws_platform]
 - [FreeBSD][freebsd_platform]

### Upgrading and compatibility

 - [Migrate applications from using API v2 to API v3][v2_migration]
 - [Upgrading a v2.3 cluster to v3.0][v3_upgrade]
 - [Upgrading a v3.0 cluster to v3.1][v31_upgrade]
 - [Upgrading a v3.1 cluster to v3.2][v32_upgrade]

## Learning

To learn more about the concepts and internals behind etcd, read the following pages:

 - [Why etcd?][why]
 - [Understand data model][data_model]
 - [Understand APIs][understand_apis]
 - [Glossary][glossary]
 - Internals
   - [Auth subsystem][auth_design]

## Frequently Asked Questions (FAQ)

Answers to [common questions] about etcd.

[api_ref]: dev-guide/api_reference_v3.md
[api_concurrency_ref]: dev-guide/api_concurrency_reference_v3.md
[api_grpc_gateway]: dev-guide/api_grpc_gateway.md
[clustering]: op-guide/clustering.md
[conf]: op-guide/configuration.md
[system-limit]: dev-guide/limit.md
[common questions]: faq.md
[why]: learning/why.md
[data_model]: learning/data_model.md
[demo]: demo.md
[download_build]: dl_build.md
[embed_etcd]: https://godoc.org/github.com/coreos/etcd/embed
[grpc_naming]: dev-guide/grpc_naming.md
[failures]: op-guide/failures.md
[gateway]: op-guide/gateway.md
[glossary]: learning/glossary.md
[namespace_client]: https://godoc.org/github.com/coreos/etcd/clientv3/namespace
[namespace_proxy]: op-guide/grpc_proxy.md#namespacing
[grpc_proxy]: op-guide/grpc_proxy.md
[hardware]: op-guide/hardware.md
[interacting]: dev-guide/interacting_v3.md
[local_cluster]: dev-guide/local_cluster.md
[performance]: op-guide/performance.md
[recovery]: op-guide/recovery.md
[maintenance]: op-guide/maintenance.md
[security]: op-guide/security.md
[monitoring]: op-guide/monitoring.md
[v2_migration]: op-guide/v2-migration.md
[container_rkt]: op-guide/container.md#rkt
[container_docker]: op-guide/container.md#docker
[understand_apis]: learning/api.md
[versioning]: op-guide/versioning.md
[supported_platforms]: op-guide/supported-platform.md
[container_linux_platform]: platforms/container-linux-systemd.md
[freebsd_platform]: platforms/freebsd.md
[aws_platform]: platforms/aws.md
[experimental]: dev-guide/experimental_apis.md
[v3_upgrade]: upgrades/upgrade_3_0.md
[v31_upgrade]: upgrades/upgrade_3_1.md
[v32_upgrade]: upgrades/upgrade_3_2.md
[authentication]: op-guide/authentication.md
[auth_design]: learning/auth_design.md
