### etcd concurrency API Reference


This is a generated documentation. Please read the proto files for more.


##### service `Lock` (etcdserver/api/v3lock/v3lockpb/v3lock.proto)

The lock service exposes client-side locking facilities as a gRPC interface.

| Method | Request Type | Response Type | Description |
| ------ | ------------ | ------------- | ----------- |
| Lock | LockRequest | LockResponse | Lock acquires a distributed shared lock on a given named lock. On success, it will return a unique key that exists so long as the lock is held by the caller. This key can be used in conjunction with transactions to safely ensure updates to etcd only occur while holding lock ownership. The lock is held until Unlock is called on the key or the lease associate with the owner expires. |
| Unlock | UnlockRequest | UnlockResponse | Unlock takes a key returned by Lock and releases the hold on lock. The next Lock caller waiting for the lock will then be woken up and given ownership of the lock. |



##### message `LockRequest` (etcdserver/api/v3lock/v3lockpb/v3lock.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| name | name is the identifier for the distributed shared lock to be acquired. | bytes |
| lease | lease is the ID of the lease that will be attached to ownership of the lock. If the lease expires or is revoked and currently holds the lock, the lock is automatically released. Calls to Lock with the same lease will be treated as a single acquistion; locking twice with the same lease is a no-op. | int64 |



##### message `LockResponse` (etcdserver/api/v3lock/v3lockpb/v3lock.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| header |  | etcdserverpb.ResponseHeader |
| key | key is a key that will exist on etcd for the duration that the Lock caller owns the lock. Users should not modify this key or the lock may exhibit undefined behavior. | bytes |



##### message `UnlockRequest` (etcdserver/api/v3lock/v3lockpb/v3lock.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| key | key is the lock ownership key granted by Lock. | bytes |



##### message `UnlockResponse` (etcdserver/api/v3lock/v3lockpb/v3lock.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| header |  | etcdserverpb.ResponseHeader |



##### service `Election` (etcdserver/api/v3election/v3electionpb/v3election.proto)

The election service exposes client-side election facilities as a gRPC interface.

| Method | Request Type | Response Type | Description |
| ------ | ------------ | ------------- | ----------- |
| Campaign | CampaignRequest | CampaignResponse | Campaign waits to acquire leadership in an election, returning a LeaderKey representing the leadership if successful. The LeaderKey can then be used to issue new values on the election, transactionally guard API requests on leadership still being held, and resign from the election. |
| Proclaim | ProclaimRequest | ProclaimResponse | Proclaim updates the leader's posted value with a new value. |
| Leader | LeaderRequest | LeaderResponse | Leader returns the current election proclamation, if any. |
| Observe | LeaderRequest | LeaderResponse | Observe streams election proclamations in-order as made by the election's elected leaders. |
| Resign | ResignRequest | ResignResponse | Resign releases election leadership so other campaigners may acquire leadership on the election. |



##### message `CampaignRequest` (etcdserver/api/v3election/v3electionpb/v3election.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| name | name is the election's identifier for the campaign. | bytes |
| lease | lease is the ID of the lease attached to leadership of the election. If the lease expires or is revoked before resigning leadership, then the leadership is transferred to the next campaigner, if any. | int64 |
| value | value is the initial proclaimed value set when the campaigner wins the election. | bytes |



##### message `CampaignResponse` (etcdserver/api/v3election/v3electionpb/v3election.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| header |  | etcdserverpb.ResponseHeader |
| leader | leader describes the resources used for holding leadereship of the election. | LeaderKey |



##### message `LeaderKey` (etcdserver/api/v3election/v3electionpb/v3election.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| name | name is the election identifier that correponds to the leadership key. | bytes |
| key | key is an opaque key representing the ownership of the election. If the key is deleted, then leadership is lost. | bytes |
| rev | rev is the creation revision of the key. It can be used to test for ownership of an election during transactions by testing the key's creation revision matches rev. | int64 |
| lease | lease is the lease ID of the election leader. | int64 |



##### message `LeaderRequest` (etcdserver/api/v3election/v3electionpb/v3election.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| name | name is the election identifier for the leadership information. | bytes |



##### message `LeaderResponse` (etcdserver/api/v3election/v3electionpb/v3election.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| header |  | etcdserverpb.ResponseHeader |
| kv | kv is the key-value pair representing the latest leader update. | mvccpb.KeyValue |



##### message `ProclaimRequest` (etcdserver/api/v3election/v3electionpb/v3election.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| leader | leader is the leadership hold on the election. | LeaderKey |
| value | value is an update meant to overwrite the leader's current value. | bytes |



##### message `ProclaimResponse` (etcdserver/api/v3election/v3electionpb/v3election.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| header |  | etcdserverpb.ResponseHeader |



##### message `ResignRequest` (etcdserver/api/v3election/v3electionpb/v3election.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| leader | leader is the leadership to relinquish by resignation. | LeaderKey |



##### message `ResignResponse` (etcdserver/api/v3election/v3electionpb/v3election.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| header |  | etcdserverpb.ResponseHeader |



##### message `Event` (mvcc/mvccpb/kv.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| type | type is the kind of event. If type is a PUT, it indicates new data has been stored to the key. If type is a DELETE, it indicates the key was deleted. | EventType |
| kv | kv holds the KeyValue for the event. A PUT event contains current kv pair. A PUT event with kv.Version=1 indicates the creation of a key. A DELETE/EXPIRE event contains the deleted key with its modification revision set to the revision of deletion. | KeyValue |
| prev_kv | prev_kv holds the key-value pair before the event happens. | KeyValue |



##### message `KeyValue` (mvcc/mvccpb/kv.proto)

| Field | Description | Type |
| ----- | ----------- | ---- |
| key | key is the key in bytes. An empty key is not allowed. | bytes |
| create_revision | create_revision is the revision of last creation on this key. | int64 |
| mod_revision | mod_revision is the revision of last modification on this key. | int64 |
| version | version is the version of the key. A deletion resets the version to zero and any modification of the key increases its version. | int64 |
| value | value is the value held by the key, in bytes. | bytes |
| lease | lease is the ID of the lease that attached to key. When the attached lease expires, the key will be deleted. If lease is 0, then no lease is attached to the key. | int64 |



