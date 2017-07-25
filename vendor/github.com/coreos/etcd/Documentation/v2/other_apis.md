**This is the documentation for etcd2 releases. Read [etcd3 doc][v3-docs] for etcd3 releases.**

[v3-docs]: ../docs.md#documentation


# Miscellaneous APIs

* [Getting the etcd version](#getting-the-etcd-version)
* [Checking health of an etcd member node](#checking-health-of-an-etcd-member-node)

## Getting the etcd version

The etcd version of a specific instance can be obtained from the `/version` endpoint.

```sh
curl -L http://127.0.0.1:2379/version
```

```
etcd 2.0.12
```

## Checking health of an etcd member node

etcd provides a `/health` endpoint to verify the health of a particular member.

```sh
curl http://10.0.0.10:2379/health
```

```json
{"health": "true"}
```
