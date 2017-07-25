# etcd3 API

This document is meant to give an overview of the etcd3 API's central design. It is by no means all encompassing, but intended to focus on the basic ideas needed to understand etcd without the distraction of less common API calls. All etcd3 API's are defined in [gRPC services][grpc-service], which categorize remote procedure calls (RPCs) understood by the etcd server. A full listing of all etcd RPCs are documented in markdown in the [gRPC API listing][grpc-api].

## gRPC Services

Every API request sent to an etcd server is a gRPC remote procedure call. RPCs in etcd3 are categorized based on functionality into services.

Services important for dealing with etcd's key space include:
* KV - Creates, updates, fetches, and deletes key-value pairs.
* Watch - Monitors changes to keys.
* Lease - Primitives for consuming client keep-alive messages.

Services which manage the cluster itself include:
* Auth - Role based authentication mechanism for authenticating users.
* Cluster - Provides membership information and configuration facilities.
* Maintenance - Takes recovery snapshots, defragments the store, and returns per-member status information.

### Requests and Responses

All RPCs in etcd3 follow the same format. Each RPC has a function `Name` which takes `NameRequest` as an argument and returns `NameResponse` as a response. For example, here is the `Range` RPC description:

```protobuf
service KV {
  Range(RangeRequest) returns (RangeResponse)
  ...
}
```

### Response header

All Responses from etcd API have an attached response header which includes cluster metadata for the response:

```proto
message ResponseHeader {
  uint64 cluster_id = 1;
  uint64 member_id = 2;
  int64 revision = 3;
  uint64 raft_term = 4;
}
```

* Cluster_ID - the ID of the cluster generating the response.
* Member_ID - the ID of the member generating the response.
* Revision - the revision of the key-value store when generating the response.
* Raft_Term - the Raft term of the member when generating the response.

An application may read the Cluster_ID (Member_ID) field to ensure it is communicating with the intended cluster (member).

Applications can use the `Revision` to know the latest revision of the key-value store. This is especially useful when applications specify a historical revision to make time `travel query` and wishes to know the latest revision at the time of the request.

Applications can use `Raft_Term` to detect when the cluster completes a new leader election.

## Key-Value API

The Key-Value API manipulates key-value pairs stored inside etcd. The majority of requests made to etcd are usually key-value requests.

### System primitives

### Key-Value pair

A key-value pair is the smallest unit that the key-value API can manipulate. Each key-value pair has a number of fields, defined in [protobuf format][kv-proto]:

```protobuf
message KeyValue {
  bytes key = 1;
  int64 create_revision = 2;
  int64 mod_revision = 3;
  int64 version = 4;
  bytes value = 5;
  int64 lease = 6;
}
```

* Key - key in bytes. An empty key is not allowed.
* Value - value in bytes.
* Version - version is the version of the key. A deletion resets the version to zero and any modification of the key increases its version.
* Create_Revision - revision of the last creation on the key.
* Mod_Revision - revision of the last modification on the key.
* Lease - the ID of the lease attached to the key. If lease is 0, then no lease is attached to the key.


In addition to just the key and value, etcd attaches additional revision metadata as part of the key message. This revision information orders keys by time of creation and modification, which is useful for managing concurrency for distributed synchronization. The etcd client's [distributed shared locks][locks] use the creation revision to wait for lock ownership. Similarly, the modification revision is used for detecting [software transactional memory][STM] read set conflicts and waiting on [leader election][elections] updates.

#### Revisions

etcd maintains a 64-bit cluster-wide counter, the store revision, that is incremented each time the key space is modified. The revision serves as a global logical clock, sequentially ordering all updates to the store. The change represented by a new revisions is incremental; the data associated with a revision is the data that changed the store. Internally, a new revision means writing the changes to the backend's B+tree, keyed by the incremented revision.

Revisions become more valuable when taking considering etcd3's [multi-version concurrency control][mvcc] backend. The MVCC model means that the key-value store can be viewed from past revisions since historical key revisions are retained. The retention policy for this history can be configured by cluster administrators for fine-grained storage management; usually etcd3 discards old revisions of keys on a timer. A typical etcd3 cluster retains superseded key data for hours. This also buys reliable handling for long client disconnection, not just transient network disruptions: watchers simply resume from the last observed historical revision. Similarly, to read from the store at a particular point-in-time, read requests can be tagged with a revision to return keys from a view of the key space at the point in time that revision was committed.

#### Key ranges

The etcd3 data model indexes all keys over a flat binary key space. This differs from other key-value store systems that use a hierarchical system of organizing keys into directories. Instead of listing keys by directory, keys are listed by key intervals `[a, b)`.

These intervals are often referred to as "ranges" in etcd3. Operations over ranges are more powerful than operations on directories. Like a hierarchical store, intervals support single key lookups via `[a, a+1)` (e.g., ['a', 'a\x00') looks up 'a') and directory lookups by encoding keys by directory depth. In addition to those operations, intervals can also encode prefixes; for example  the interval `['a', 'b')` looks up all keys prefixed by the string 'a'.

By convention, ranges for a Request are denoted by the fields `key` and `range_end`. The `key` field is the first key of the range and should be non-empty. The `range_end` is the key following the last key of the range. If `range_end` is not given or empty, the range is defined to contain only the key argument. If `range_end` is `key` plus one (e.g., "aa"+1 == "ab", "a\xff"+1 == "b"), then the range represents all keys prefixed with key. If both `key` and `range_end` are '\0', then range represents all keys. If `range_end` is '\0', the range is all keys greater than or equal to the key argument.

### Range

Keys are fetched from the key-value store using the `Range` API call, which takes a `RangeRequest`:

```protobuf
message RangeRequest {
  enum SortOrder {
	NONE = 0; // default, no sorting
	ASCEND = 1; // lowest target value first
	DESCEND = 2; // highest target value first
  }
  enum SortTarget {
	KEY = 0;
	VERSION = 1;
	CREATE = 2;
	MOD = 3;
	VALUE = 4;
  }

  bytes key = 1;
  bytes range_end = 2;
  int64 limit = 3;
  int64 revision = 4;
  SortOrder sort_order = 5;
  SortTarget sort_target = 6;
  bool serializable = 7;
  bool keys_only = 8;
  bool count_only = 9;
  int64 min_mod_revision = 10;
  int64 max_mod_revision = 11;
  int64 min_create_revision = 12;
  int64 max_create_revision = 13;
}
```

* Key, Range_End - The key range to fetch.
* Limit - the maximum number of keys returned for the request. When limit is set to 0, it is treated as no limit.
* Revision - the point-in-time of the key-value store to use for the range. If revision is less or equal to zero, the range is over the latest key-value store If the revision is compacted, ErrCompacted is returned as a response.
* Sort_Order - the ordering for sorted requests.
* Sort_Target - the key-value field to sort.
* Serializable - sets the range request to use serializable member-local reads. By default, Range is linearizable; it reflects the current consensus of the cluster. For better performance and availability, in exchange for possible stale reads, a serializable range request is served locally without needing to reach consensus with other nodes in the cluster.
* Keys_Only - return only the keys and not the values.
* Count_Only - return only the count of the keys in the range.
* Min_Mod_Revision - the lower bound for key mod revisions; filters out lesser mod revisions.
* Max_Mod_Revision - the upper bound for key mod revisions; filters out greater mod revisions.
* Min_Create_Revision - the lower bound for key create revisions; filters out lesser create revisions.
* Max_Create_Revision - the upper bound for key create revisions; filters out greater create revisions.

The client receives a `RangeResponse` message from the `Range` call:

```protobuf
message RangeResponse {
  ResponseHeader header = 1;
  repeated mvccpb.KeyValue kvs = 2;
  bool more = 3;
  int64 count = 4;
}
```

* Kvs - the list of key-value pairs matched by the range request. When `Count_Only` is set, `Kvs` is empty.
* More - indicates if there are more keys to return in the requested range if `limit` is set.
* Count - the total number of keys satisfying the range request.

### Put

Keys are saved into the key-value store by issuing a `Put` call, which takes a `PutRequest`:

```protobuf
message PutRequest {
  bytes key = 1;
  bytes value = 2;
  int64 lease = 3;
  bool prev_kv = 4;
  bool ignore_value = 5;
  bool ignore_lease = 6;
}
```

* Key - the name of the key to put into the key-value store.
* Value - the value, in bytes, to associate with the key in the key-value store.
* Lease - the lease ID to associate with the key in the key-value store. A lease value of 0 indicates no lease.
* Prev_Kv - when set, responds with the key-value pair data before the update from this `Put` request.
* Ignore_Value - when set, update the key without changing its current value. Returns an error if the key does not exist.
* Ignore_Lease - when set, update the key without changing its current lease. Returns an error if the key does not exist.

The client receives a `PutResponse` message from the `Put` call:

```protobuf
message PutResponse {
  ResponseHeader header = 1;
  mvccpb.KeyValue prev_kv = 2;
}
```

* Prev_Kv - the key-value pair overwritten by the `Put`, if `Prev_Kv` was set in the `PutRequest`.

### Delete Range

Ranges of keys are deleted using the `DeleteRange` call, which takes a `DeleteRangeRequest`:

```protobuf
message DeleteRangeRequest {
  bytes key = 1;
  bytes range_end = 2;
  bool prev_kv = 3;
}
```

* Key, Range_End - The key range to delete.
* Prev_Kv - when set, return the contents of the deleted key-value pairs.

The client receives a `DeleteRangeResponse` message from the `DeleteRange` call:

```protobuf
message DeleteRangeResponse {
  ResponseHeader header = 1;
  int64 deleted = 2;
  repeated mvccpb.KeyValue prev_kvs = 3;
}
```

* Deleted - number of keys deleted.
* Prev_Kv - a list of all key-value pairs deleted by the DeleteRange operation.

### Transaction

A transaction is an atomic If/Then/Else construct over the key-value store. It provides a primitive for grouping requests together in atomic blocks (i.e., then/else) whose execution is guarded (i.e., if) based on the contents of the key-value store. Transactions can be used for protecting keys from unintended concurrent updates, building compare-and-swap operations, and developing higher-level concurrency control.

A transaction can atomically process multiple requests in a single request. For modifications to the key-value store, this means the store's revision is incremented only once for the transaction and all events generated by the transaction will have the same revision. However, modifications to the same key multiple times within a single transaction are forbidden.

All transactions are guarded by a conjunction of comparisons, similar to an "If" statement. Each comparison checks a single key in the store. It may check for the absence or presence of a value, compare with a given value, or check a key's revision or version. Two different comparisons may apply to the same or different keys. All comparisons are applied atomically; if all comparisons are true, the transaction is said to succeed and etcd applies the transaction's then / `success` request block, otherwise it is said to fail and applies the else / `failure` request block.

Each comparison is encoded as a `Compare` message:

```protobuf
message Compare {
  enum CompareResult {
    EQUAL = 0;
    GREATER = 1;
    LESS = 2;
    NOT_EQUAL = 3;
  }
  enum CompareTarget {
    VERSION = 0;
    CREATE = 1;
    MOD = 2;
    VALUE= 3;
  }
  CompareResult result = 1;
  // target is the key-value field to inspect for the comparison.
  CompareTarget target = 2;
  // key is the subject key for the comparison operation.
  bytes key = 3;
  oneof target_union {
    int64 version = 4;
    int64 create_revision = 5;
    int64 mod_revision = 6;
    bytes value = 7;
  }
}
```

* Result - the kind of logical comparison operation (e.g., equal, less than, etc).
* Target - the key-value field to be compared. Either the key's version, create revision, modification revision, or value.
* Key - the key for the comparison.
* Target_Union - the user-specified data for the comparison.

After processing the comparison block, the transaction applies a block of requests. A block is a list of `RequestOp` messages:

```protobuf
message RequestOp {
  // request is a union of request types accepted by a transaction.
  oneof request {
    RangeRequest request_range = 1;
    PutRequest request_put = 2;
    DeleteRangeRequest request_delete_range = 3;
  }
}
```

* Request_Range - a `RangeRequest`.
* Request_Put - a `PutRequest`. The keys must be unique. It may not share keys with any other Puts or Deletes.
* Request_Delete_Range - a `DeleteRangeRequest`. It may not share keys with any Puts or Deletes requests.

All together, a transaction is issued with a `Txn` API call, which takes a `TxnRequest`:

```protobuf
message TxnRequest {
  repeated Compare compare = 1;
  repeated RequestOp success = 2;
  repeated RequestOp failure = 3;
}
```

* Compare - A list of predicates representing a conjunction of terms for guarding the transaction.
* Success - A list of requests to process if all compare tests evaluate to true.
* Failure - A list of requests to process if any compare test evaluates to false.

The client receives a `TxnResponse` message from the `Txn` call:

```protobuf
message TxnResponse {
  ResponseHeader header = 1;
  bool succeeded = 2;
  repeated ResponseOp responses = 3;
}
```

* Succeeded - Whether `Compare` evaluated to true or false.
* Responses - A list of responses corresponding to the results from applying the `Success` block if succeeded is true or the `Failure` if succeeded is false.

The `Responses` list corresponds to the results from the applied `RequestOp` list, with each response encoded as a `ResponseOp`:

```protobuf
message ResponseOp {
  oneof response {
    RangeResponse response_range = 1;
    PutResponse response_put = 2;
    DeleteRangeResponse response_delete_range = 3;
  }
}
```

## Watch API

The Watch API provides an event-based interface for asynchronously monitoring changes to keys. An etcd3 watch waits for changes to keys by continuously watching from a given revision, either current or historical, and streams key updates back to the client.

### Events

Every change to every key is represented with `Event` messages. An `Event` message provides both the update's data and the type of update:

```protobuf
message Event {
  enum EventType {
    PUT = 0;
    DELETE = 1;
  }
  EventType type = 1;
  KeyValue kv = 2;
  KeyValue prev_kv = 3;
}
```

* Type - The kind of event. A PUT type indicates new data has been stored to the key. A DELETE indicates the key was deleted.
* KV - The KeyValue associated with the event. A PUT event contains current kv pair. A PUT event with kv.Version=1 indicates the creation of a key. A DELETE event contains the deleted key with its modification revision set to the revision of deletion.
* Prev_KV - The key-value pair for the key from the revision immediately before the event. To save bandwidth, it is only filled out if the watch has explicitly enabled it.

### Watch streams

Watches are long-running requests and use gRPC streams to stream event data. A watch stream is bi-directional; the client writes to the stream to establish watches and reads to receive watch event. A single watch stream can multiplex many distinct watches by tagging events with per-watch identifiers. This multiplexing helps reducing the memory footprint and connection overhead on the core etcd cluster.

Watches make three guarantees about events:
* Ordered - events are ordered by revision; an event will never appear on a watch if it precedes an event in time that has already been posted.
* Reliable - a sequence of events will never drop any subsequence of events; if there are events ordered in time as a < b < c, then if the watch receives events a and c, it is guaranteed to receive b.
* Atomic - a list of events is guaranteed to encompass complete revisions; updates in the same revision over multiple keys will not be split over several lists of events.

A client creates a watch by sending a `WatchCreateRequest` over a stream returned by `Watch`:

```protobuf
message WatchCreateRequest {
  bytes key = 1;
  bytes range_end = 2;
  int64 start_revision = 3;
  bool progress_notify = 4;

  enum FilterType {
    NOPUT = 0;
    NODELETE = 1;
  }
  repeated FilterType filters = 5;
  bool prev_kv = 6;
}
```

* Key, Range_End - The key range to watch.
* Start_Revision - An optional revision for where to inclusively begin watching. If not given, it will stream events following the revision of the watch creation response header revision. The entire available event history can be watched starting from the last compaction revision.
* Progress_Notify - When set, the watch will periodically receive a WatchResponse with no events, if there are no recent events. It is useful when clients wish to recover a disconnected watcher starting from a recent known revision. The etcd server decides how often to send notifications based on current server load.
* Filters - A list of event types to filter away at server side.
* Prev_Kv - When set, the watch receives the key-value data from before the event happens. This is useful for knowing what data has been overwritten.

In response to a `WatchCreateRequest` or if there is a new event for some established watch, the client receives a `WatchResponse`:

```protobuf
message WatchResponse {
  ResponseHeader header = 1;
  int64 watch_id = 2;
  bool created = 3;
  bool canceled = 4;
  int64 compact_revision = 5;

  repeated mvccpb.Event events = 11;
}
```

* Watch_ID - the ID of the watch that corresponds to the response.
* Created - set to true if the response is for a create watch request. The client should record ID and expect to receive events for the watch on the stream. All events sent to the created watcher will have the same watch_id.
* Canceled - set to true if the response is for a cancel watch request. No further events will be sent to the canceled watcher.
* Compact_Revision - set to the minimum historical revision available to etcd if a watcher tries watching at a compacted revision. This happens when creating a watcher at a compacted revision or the watcher cannot catch up with the progress of the key-value store. The watcher will be canceled; creating new watches with the same start_revision will fail.
* Events - a list of new events in sequence corresponding to the given watch ID.

If the client wishes to stop receiving events for a watch, it issues a `WatchCancelRequest`:

```protobuf
message WatchCancelRequest {
   int64 watch_id = 1;
}
```

* Watch_ID - the ID of the watch to cancel so that no more events are transmitted.

## Lease API

Leases are a mechanism for detecting client liveness. The cluster grants leases with a time-to-live. A lease expires if the etcd cluster does not receive a keepAlive within a given TTL period.

To tie leases into the key-value store, each key may be attached to at most one lease. When a lease expires or is revoked, all keys attached to that lease will be deleted. Each expired key generates a delete event in the event history.

### Obtaining leases

Leases are obtained through the `LeaseGrant` API call, which takes a `LeaseGrantRequest`:

```protobuf
message LeaseGrantRequest {
  int64 TTL = 1;
  int64 ID = 2;
}
```

* TTL - the advisory time-to-live, in seconds.
* ID - the requested ID for the lease. If ID is set to 0, etcd will choose an ID.

The client receives a `LeaseGrantResponse` from the `LeaseGrant` call:

```protobuf
message LeaseGrantResponse {
  ResponseHeader header = 1;
  int64 ID = 2;
  int64 TTL = 3;
}
```

* ID - the lease ID for the granted lease.
* TTL - is the server selected time-to-live, in seconds, for the lease.

```protobuf
message LeaseRevokeRequest {
  int64 ID = 1;
}
```

* ID - the lease ID to revoke. When the lease is revoked, all attached keys are deleted.

### Keep alives

Leases are refreshed using a bi-directional stream created with the `LeaseKeepAlive` API call. When the client wishes to refresh a lease, it sends a `LeaseKeepAliveRequest` over the stream:

```protobuf
message LeaseKeepAliveRequest {
  int64 ID = 1;
}
```

* ID - the lease ID for the lease to keep alive.

The keep alive stream responds with a `LeaseKeepAliveResponse`:

```protobuf
message LeaseKeepAliveResponse {
  ResponseHeader header = 1;
  int64 ID = 2;
  int64 TTL = 3;
}
```

* ID - the lease that was refreshed with a new TTL.
* TTL - the new time-to-live, in seconds, that the lease has remaining.

[elections]: https://github.com/coreos/etcd/blob/master/clientv3/concurrency/election.go
[kv-proto]: https://github.com/coreos/etcd/blob/master/mvcc/mvccpb/kv.proto
[grpc-api]: ../dev-guide/api_reference_v3.md
[grpc-service]: https://github.com/coreos/etcd/blob/master/etcdserver/etcdserverpb/rpc.proto
[locks]: https://github.com/coreos/etcd/blob/master/clientv3/concurrency/mutex.go
[mvcc]: https://en.wikipedia.org/wiki/Multiversion_concurrency_control
[stm]: https://github.com/coreos/etcd/blob/master/clientv3/concurrency/stm.go
