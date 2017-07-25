---
layout: "docs"
page_title: "Semaphore"
sidebar_current: "docs-guides-semaphore"
description: |-
  This guide demonstrates how to implement a distributed semaphore using the Consul KV store.
---

# Semaphore

This guide demonstrates how to implement a distributed semaphore using the Consul
KV store. This is useful when you want to coordinate many services while
restricting access to certain resources.

~>  If you only need mutual exclusion or leader election,
[this guide](/docs/guides/leader-election.html)
provides a simpler algorithm that can be used instead.

There are a number of ways that a semaphore can be built, so our goal is not to
cover all the possible methods. Instead, we will focus on using Consul's support for
[sessions](/docs/internals/sessions.html). Sessions allow us to build a system that
can gracefully handle failures.

-> **Note:** JSON output in this guide has been pretty-printed for easier reading. Actual values returned from the API will not be formatted.

## Contending Nodes

Let's imagine we have a set of nodes who are attempting to acquire a slot in the
semaphore. All nodes that are participating should agree on three decisions: the
prefix in the KV store used to coordinate, a single key to use as a lock,
and a limit on the number of slot holders.

For the prefix we will be using for coordination, a good pattern is simply:

```text
service/<service name>/lock/
```

We'll abbreviate this pattern as simply `<prefix>` for the rest of this guide.

The first step is to create a session. This is done using the
[Session HTTP API](/api/session.html#session_create):

```text
curl  -X PUT -d '{"Name": "dbservice"}' \
  http://localhost:8500/v1/session/create
 ```

This will return a JSON object contain the session ID:

```text
{
  "ID": "4ca8e74b-6350-7587-addf-a18084928f3c"
}
```

Next, we create a contender entry. Each contender creates an entry that is tied
to a session. This is done so that if a contender is holding a slot and fails,
it can be detected by the other contenders.

Create the contender key by doing an `acquire` on `<prefix>/<session>` via `PUT`.
This is something like:

```text
curl -X PUT -d <body> http://localhost:8500/v1/kv/<prefix>/<session>?acquire=<session>
 ```

The `<session>` value is the ID returned by the call to
[`/v1/session/create`](/api/session.html#session_create).

`body` can be used to associate a meaningful value with the contender. This is opaque
to Consul but can be useful for human operators.

The call will either return `true` or `false`. If `true`, the contender entry has been
created. If `false`, the contender node was not created; it's likely that this indicates
a session invalidation.

The next step is to use a single key to coordinate which holders are currently
reserving a slot. A good choice for this lock key is simply `<prefix>/.lock`. We will
refer to this special coordinating key as `<lock>`.

The current state of the semaphore is read by doing a `GET` on the entire `<prefix>`:

```text
curl http://localhost:8500/v1/kv/<prefix>?recurse
 ```

Within the list of the entries, we should find the `<lock>`. That entry should hold
both the slot limit and the current holders. A simple JSON body like the following works:

```text
{
    "Limit": 3,
    "Holders": {
        "4ca8e74b-6350-7587-addf-a18084928f3c": true,
        "adf4238a-882b-9ddc-4a9d-5b6758e4159e": true
    }
}
```

When the `<lock>` is read, we can verify the remote `Limit` agrees with the local value. This
is used to detect a potential conflict. The next step is to determine which of the current
slot holders are still alive. As part of the results of the `GET`, we have all the contender
entries. By scanning those entries, we create a set of all the `Session` values. Any of the
`Holders` that are not in that set are pruned. In effect, we are creating a set of live contenders
based on the list results and doing a set difference with the `Holders` to detect and prune
any potentially failed holders.

If the number of holders (after pruning) is less than the limit, a contender attempts acquisition
by adding its own session to the `Holders` and doing a Check-And-Set update of the `<lock>`. This
performs an optimistic update.

This is done by:

```text
curl -X PUT -d <Updated Lock> http://localhost:8500/v1/kv/<lock>?cas=<lock-modify-index>
 ```

If this succeeds with `true`, the contender now holds a slot in the semaphore. If this fails
with `false`, then likely there was a race with another contender to acquire the slot.
Both code paths now go into an idle waiting state. In this state, we watch for changes
on `<prefix>`. This is because a slot may be released, a node may fail, etc.
Slot holders must also watch for changes since the slot may be released by an operator
or automatically released due to a false positive in the failure detector.

Note that the session by default makes use of only the gossip failure detector. That
is, the session is considered held by a node as long as the default Serf health check
has not declared the node unhealthy. Additional checks can be specified if desired.

Watching for changes is done via a blocking query against `<prefix>`. If a contender
holds a slot, then on any change the `<lock>` should be re-checked to ensure the slot is
still held. If no slot is held, then the same acquisition logic is triggered to check
and potentially re-attempt acquisition. This allows a contender to steal the slot from
a failed contender or one that has voluntarily released its slot.

If a slot holder ever wishes to release voluntarily, this should be done by doing a
Check-And-Set operation against `<lock>` to remove its session from the `Holders` object.
Once that is done, the contender entry at `<prefix>/<session>` should be deleted. Finally,
the session should be destroyed.
