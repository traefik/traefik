---
layout: "docs"
page_title: "Sessions"
sidebar_current: "docs-internals-sessions"
description: |-
  Consul provides a session mechanism which can be used to build distributed locks. Sessions act as a binding layer between nodes, health checks, and key/value data. They are designed to provide granular locking and are heavily inspired by The Chubby Lock Service for Loosely-Coupled Distributed Systems.
---

# Sessions

Consul provides a session mechanism which can be used to build distributed locks.
Sessions act as a binding layer between nodes, health checks, and key/value data.
They are designed to provide granular locking and are heavily inspired by
[The Chubby Lock Service for Loosely-Coupled Distributed Systems](http://research.google.com/archive/chubby.html).

~> **Advanced Topic!** This page covers technical details of
the internals of Consul. You don't need to know these details to effectively
operate and use Consul. These details are documented here for those who wish
to learn about them without having to go spelunking through the source code.

## Session Design

A session in Consul represents a contract that has very specific semantics.
When a session is constructed, a node name, a list of health checks, a behavior,
a TTL, and a `lock-delay` may be provided. The newly constructed session is provided with
a named ID that can be used to identify it. This ID can be used with the KV
store to acquire locks: advisory mechanisms for mutual exclusion.

Below is a diagram showing the relationship between these components:

<div class="center">
![Consul Sessions](consul-sessions.png)
</div>

The contract that Consul provides is that under any of the following
situations, the session will be *invalidated*:

* Node is deregistered
* Any of the health checks are deregistered
* Any of the health checks go to the critical state
* Session is explicitly destroyed
* TTL expires, if applicable

When a session is invalidated, it is destroyed and can no longer
be used. What happens to the associated locks depends on the
behavior specified at creation time. Consul supports a `release`
and `delete` behavior. The `release` behavior is the default
if none is specified.

If the `release` behavior is being used, any of the locks held in
association with the session are released, and the `ModifyIndex` of
the key is incremented. Alternatively, if the `delete` behavior is
used, the key corresponding to any of the held locks is simply deleted.
This can be used to create ephemeral entries that are automatically
deleted by Consul.

While this is a simple design, it enables a multitude of usage
patterns. By default, the
[gossip based failure detector](/docs/internals/gossip.html)
is used as the associated health check. This failure detector allows
Consul to detect when a node that is holding a lock has failed and
to automatically release the lock. This ability provides **liveness** to
Consul locks; that is, under failure the system can continue to make
progress. However, because there is no perfect failure detector, it's possible
to have a false positive (failure detected) which causes the lock to
be released even though the lock owner is still alive. This means
we are sacrificing some **safety**.

Conversely, it is possible to create a session with no associated
health checks. This removes the possibility of a false positive
and trades liveness for safety. You can be absolutely certain Consul
will not release the lock even if the existing owner has failed.
Since Consul APIs allow a session to be force destroyed, this allows
systems to be built that require an operator to intervene in the
case of a failure while precluding the possibility of a split-brain.

A third health checking mechanism is session TTLs. When creating
a session, a TTL can be specified. If the TTL interval expires without
being renewed, the session has expired and an invalidation is triggered.
This type of failure detector is also known as a heartbeat failure detector.
It is less scalable than the gossip based failure detector as it places
an increased burden on the servers but may be applicable in some cases.
The contract of a TTL is that it represents a lower bound for invalidation;
that is, Consul will not expire the session before the TTL is reached, but it
is allowed to delay the expiration past the TTL. The TTL is renewed on
session creation, on session renew, and on leader failover. When a TTL
is being used, clients should be aware of clock skew issues: namely,
time may not progress at the same rate on the client as on the Consul servers.
It is best to set conservative TTL values and to renew in advance of the TTL
to account for network delay and time skew.

The final nuance is that sessions may provide a `lock-delay`. This
is a time duration, between 0 and 60 seconds. When a session invalidation
takes place, Consul prevents any of the previously held locks from
being re-acquired for the `lock-delay` interval; this is a safeguard
inspired by Google's Chubby. The purpose of this delay is to allow
the potentially still live leader to detect the invalidation and stop
processing requests that may lead to inconsistent state. While not a
bulletproof method, it does avoid the need to introduce sleep states
into application logic and can help mitigate many issues. While the
default is to use a 15 second delay, clients are able to disable this
mechanism by providing a zero delay value.

## K/V Integration

Integration between the KV store and sessions is the primary
place where sessions are used. A session must be created prior to use
and is then referred to by its ID.

The KV API is extended to support an `acquire` and `release` operation.
The `acquire` operation acts like a Check-And-Set operation except it
can only succeed if there is no existing lock holder (the current lock holder
can re-`acquire`, see below). On success, there is a normal key update, but
there is also an increment to the `LockIndex`, and the `Session` value is
updated to reflect the session holding the lock.

If the lock is already held by the given session during an `acquire`, then
the `LockIndex` is not incremented but the key contents are updated. This
lets the current lock holder update the key contents without having to give
up the lock and reacquire it.

Once held, the lock can be released using a corresponding `release` operation,
providing the same session. Again, this acts like a Check-And-Set operations
since the request will fail if given an invalid session. A critical note is
that the lock can be released without being the creator of the session.
This is by design as it allows operators to intervene and force terminate
a session if necessary. As mentioned above, a session invalidation will also
cause all held locks to be released or deleted. When a lock is released, the `LockIndex`
does not change; however, the `Session` is cleared and the `ModifyIndex` increments.

These semantics (heavily borrowed from Chubby), allow the tuple of (Key, LockIndex, Session)
to act as a unique "sequencer". This `sequencer` can be passed around and used
to verify if the request belongs to the current lock holder. Because the `LockIndex`
is incremented on each `acquire`, even if the same session re-acquires a lock,
the `sequencer` will be able to detect a stale request. Similarly, if a session is
invalided, the Session corresponding to the given `LockIndex` will be blank.

To be clear, this locking system is purely *advisory*. There is no enforcement
that clients must acquire a lock to perform any operation. Any client can
read, write, and delete a key without owning the corresponding lock. It is not
the goal of Consul to protect against misbehaving clients.

## Leader Election

The primitives provided by sessions and the locking mechanisms of the KV
store can be used to build client-side leader election algorithms.
These are covered in more detail in the [Leader Election guide](/docs/guides/leader-election.html).

## Prepared Query Integration

Prepared queries may be attached to a session in order to automatically delete
the prepared query when the session is invalidated.
