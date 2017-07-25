# etcd gateway

## What is etcd gateway

etcd gateway is a simple TCP proxy that forwards network data to the etcd cluster. The gateway is stateless and transparent; it neither inspects client requests nor interferes with cluster responses.

The gateway supports multiple etcd server endpoints and works on a simple round-robin policy. It only routes to available enpoints and hides failures from its clients. Other retry policies, such as weighted round-robin, may be supported in the future.

## When to use etcd gateway

Every application that accesses etcd must first have the address of an etcd cluster client endpoint. If multiple applications on the same server access the same etcd cluster, every application still needs to know the advertised client endpoints of the etcd cluster. If the etcd cluster is reconfigured to have different endpoints, every application may also need to update its endpoint list. This wide-scale reconfiguration is both tedious and error prone.

etcd gateway solves this problem by serving as a stable local endpoint. A typical etcd gateway configuration has each machine running a gateway listening on a local address and every etcd application connecting to its local gateway. The upshot is only the gateway needs to update its endpoints instead of updating each and every application.

In summary, to automatically propagate cluster endpoint changes, the etcd gateway runs on every machine serving multiple applications accessing the same etcd cluster.

## When not to use etcd gateway

- Improving performance

The gateway is not designed for improving etcd cluster performance. It does not provide caching, watch coalescing or batching. The etcd team is developing a caching proxy designed for improving cluster scalability. 

- Running on a cluster management system

Advanced cluster management systems like Kubernetes natively support service discovery. Applications can access an etcd cluster with a DNS name or a virtual IP address managed by the system. For example, kube-proxy is equivalent to etcd gateway.

## Start etcd gateway

Consider an etcd cluster with the following static endpoints:

|Name|Address|Hostname|
|------|---------|------------------|
|infra0|10.0.1.10|infra0.example.com|
|infra1|10.0.1.11|infra1.example.com|
|infra2|10.0.1.12|infra2.example.com|

Start the etcd gateway to use these static endpoints with the command:

```bash
$ etcd gateway start --endpoints=infra0.example.com,infra1.example.com,infra2.example.com
2016-08-16 11:21:18.867350 I | tcpproxy: ready to proxy client requests to [...]
```

Alternatively, if using DNS for service discovery, consider the DNS SRV entries:

```bash
$ dig +noall +answer SRV _etcd-client._tcp.example.com
_etcd-client._tcp.example.com. 300 IN SRV 0 0 2379 infra0.example.com.
_etcd-client._tcp.example.com. 300 IN SRV 0 0 2379 infra1.example.com.
_etcd-client._tcp.example.com. 300 IN SRV 0 0 2379 infra2.example.com.
```

```bash
$ dig +noall +answer infra0.example.com infra1.example.com infra2.example.com
infra0.example.com.  300  IN  A  10.0.1.10
infra1.example.com.  300  IN  A  10.0.1.11
infra2.example.com.  300  IN  A  10.0.1.12
```

Start the etcd gateway to fetch the endpoints from the DNS SRV entries with the command:

```bash
$ etcd gateway --discovery-srv=example.com
2016-08-16 11:21:18.867350 I | tcpproxy: ready to proxy client requests to [...]
```

## Configuration flags

### etcd cluster

#### --endpoints

 * Comma-separated list of etcd server targets for forwarding client connections.
 * Default: `127.0.0.1:2379`
 * Invalid example: `https://127.0.0.1:2379` (gateway does not terminate TLS)

#### --discovery-srv

 * DNS domain used to bootstrap cluster endpoints through SRV recrods.
 * Default: (not set)

### Network

#### --listen-addr

 * Interface and port to bind for accepting client requests.
 * Default: `127.0.0.1:23790`

#### --retry-delay

 * Duration of delay before retrying to connect to failed endpoints.
 * Default: 1m0s
 * Invalid example: "123" (expects time unit in format)

### Security

#### --insecure-discovery

 * Accept SRV records that are insecure or susceptible to man-in-the-middle attacks.
 * Default: `false`

#### --trusted-ca-file

 * Path to the client TLS CA file for the etcd cluster. Used to authenticate endpoints.
 * Default: (not set)
