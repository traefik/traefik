---
layout: "intro"
page_title: "KV Data"
sidebar_current: "gettingstarted-kv"
description: |-
  In addition to providing service discovery and integrated health checking, Consul provides an easy to use KV store. This can be used to hold dynamic configuration, assist in service coordination, build leader election, and enable anything else a developer can think to build.
---

# KV Data

In addition to providing service discovery and integrated health checking,
Consul provides an easy to use KV store. This can be used to hold
dynamic configuration, assist in service coordination, build leader election,
and enable anything else a developer can think to build.

This step assumes you have at least one Consul agent already running.

## Simple Usage

To demonstrate how simple it is to get started, we will manipulate a few keys in
the K/V store. There are two ways to interact with the Consul KV store: via the
HTTP API and via the Consul KV CLI. The examples below show using the Consul KV
CLI because it is the easiest to get started. For more advanced integrations,
you may want to use the [Consul KV HTTP API][kv-api]

First let us explore the KV store. We can ask Consul for the value of the key at
the path named `redis/config/minconns`:

```sh
$ consul kv get redis/config/minconns
Error! No key exists at: redis/config/minconns
```

As you can see, we get no result, which makes sense because there is no data in
the KV store. Next we can insert or "put" values into the KV store.

```sh
$ consul kv put redis/config/minconns 1
Success! Data written to: redis/config/minconns

$ consul kv put redis/config/maxconns 25
Success! Data written to: redis/config/maxconns

$ consul kv put -flags=42 redis/config/users/admin abcd1234
Success! Data written to: redis/config/users/admin
```

Now that we have keys in the store, we can query for the value of individual
keys:

```sh
$ consul kv get redis/config/minconns
1
```

Consul retains additional metadata about the field, which is retrieved using the
`-detailed` flag:

```sh
$ consul kv get -detailed redis/config/minconns
CreateIndex      207
Flags            0
Key              redis/config/minconns
LockIndex        0
ModifyIndex      207
Session          -
Value            1
```

For the key "redis/config/users/admin", we set a `flag` value of 42. All keys
support setting a 64-bit integer flag value. This is not used internally by
Consul, but it can be used by clients to add meaningful metadata to any KV.

It is possible to list all the keys in the store using the `recurse` options.
Results will be returned in lexicographical order:

```sh
$ consul kv get -recurse
redis/config/maxconns:25
redis/config/minconns:1
redis/config/users/admin:abcd1234
```

To delete a key from the Consul KV store, issue a "delete" call:

```sh
$ consul kv delete redis/config/minconns
Success! Deleted key: redis/config/minconns
```

It is also possible to delete an entire prefix using the `recurse` option:

```sh
$ consul kv delete -recurse redis
Success! Deleted keys with prefix: redis
```

To update the value of an existing key, "put" a value at the same path:

```sh
$ consul kv put foo bar

$ consul kv get foo
bar

$ consul kv put foo zip

$ consul kv get foo
zip
```

Consul can provide atomic key updates using a Check-And-Set operation. To perform a CAS operation, specify the `-cas` flag:

```sh
$ consul kv put -cas -modify-index=123 foo bar
Success! Data written to: foo

$ consul kv put -cas -modify-index=123 foo bar
Error! Did not write to foo: CAS failed
```

In this case, the first CAS update succeeds because the index is 123. The second
operation fails because the index is no longer 123.

## Next Steps

These are only a few examples of what the API supports. For the complete
documentation, please see [Consul KV HTTP API][kv-api] or
[Consul KV CLI][kv-cli] documentation.

Next, we will look at the [web UI](ui.html) options supported by Consul.

[kv-api]: /api/kv.html
[kv-cli]: /docs/commands/kv.html
