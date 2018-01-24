package etcdv3

import (
	"context"
	"crypto/tls"
	"strings"
	"sync"
	"time"

	"github.com/abronan/valkeyrie"
	"github.com/abronan/valkeyrie/store"
	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

const (
	defaultLockTTL     = 20 * time.Second
	etcdDefaultTimeout = 5 * time.Second
	lockSuffix         = "___lock"
)

// EtcdV3 is the receiver type for the
// Store interface
type EtcdV3 struct {
	client *etcd.Client
}

type etcdLock struct {
	lock  sync.Mutex
	store *EtcdV3

	mutex   *concurrency.Mutex
	session *concurrency.Session

	mutexKey string // mutexKey is the key to write appended with a "_lock" suffix
	writeKey string // writeKey is the actual key to update protected by the mutexKey
	value    string
	ttl      time.Duration
}

// Register registers etcd to valkeyrie
func Register() {
	valkeyrie.AddStore(store.ETCDV3, New)
}

// New creates a new Etcd client given a list
// of endpoints and an optional tls config
func New(addrs []string, options *store.Config) (store.Store, error) {
	s := &EtcdV3{}

	var (
		entries []string
		err     error
	)

	entries = store.CreateEndpoints(addrs, "http")
	cfg := &etcd.Config{
		Endpoints: entries,
	}

	// Set options
	if options != nil {
		if options.TLS != nil {
			setTLS(cfg, options.TLS, addrs)
		}
		if options.ConnectionTimeout != 0 {
			setTimeout(cfg, options.ConnectionTimeout)
		}
		if options.Username != "" {
			setCredentials(cfg, options.Username, options.Password)
		}
		if options.SyncPeriod != 0 {
			cfg.AutoSyncInterval = options.SyncPeriod
		}
	}

	s.client, err = etcd.New(*cfg)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// setTLS sets the tls configuration given a tls.Config scheme
func setTLS(cfg *etcd.Config, tls *tls.Config, addrs []string) {
	entries := store.CreateEndpoints(addrs, "https")
	cfg.Endpoints = entries
	cfg.TLS = tls
}

// setTimeout sets the timeout used for connecting to the store
func setTimeout(cfg *etcd.Config, time time.Duration) {
	cfg.DialTimeout = time
}

// setCredentials sets the username/password credentials for connecting to Etcd
func setCredentials(cfg *etcd.Config, username, password string) {
	cfg.Username = username
	cfg.Password = password
}

// Normalize the key for usage in Etcd
func (s *EtcdV3) normalize(key string) string {
	key = store.Normalize(key)
	return strings.TrimPrefix(key, "/")
}

// Get the value at "key", returns the last modified
// index to use in conjunction to Atomic calls
func (s *EtcdV3) Get(key string, opts *store.ReadOptions) (pair *store.KVPair, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), etcdDefaultTimeout)

	var result *etcd.GetResponse

	if opts != nil && !opts.Consistent {
		result, err = s.client.KV.Get(ctx, s.normalize(key), etcd.WithSerializable())
	} else {
		result, err = s.client.KV.Get(ctx, s.normalize(key))
	}

	cancel()

	if err != nil {
		return nil, err
	}

	if result.Count == 0 {
		return nil, store.ErrKeyNotFound
	}

	kvs := []*store.KVPair{}

	for _, pair := range result.Kvs {
		kvs = append(kvs, &store.KVPair{
			Key:       string(pair.Key),
			Value:     []byte(pair.Value),
			LastIndex: uint64(pair.ModRevision),
		})
	}

	return kvs[0], nil
}

// Put a value at "key"
func (s *EtcdV3) Put(key string, value []byte, opts *store.WriteOptions) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), etcdDefaultTimeout)
	pr := s.client.Txn(ctx)

	if opts != nil && opts.TTL > 0 {
		lease := etcd.NewLease(s.client)
		resp, err := lease.Grant(context.Background(), int64(opts.TTL/time.Second))
		if err != nil {
			cancel()
			return err
		}
		pr.Then(etcd.OpPut(key, string(value), etcd.WithLease(resp.ID)))
	} else {
		pr.Then(etcd.OpPut(key, string(value)))
	}

	_, err = pr.Commit()
	cancel()
	if err != nil {
		return err
	}

	return nil
}

// Delete a value at "key"
func (s *EtcdV3) Delete(key string) error {
	resp, err := s.client.KV.Delete(context.Background(), s.normalize(key))
	if resp.Deleted == 0 {
		return store.ErrKeyNotFound
	}
	return err
}

// Exists checks if the key exists inside the store
func (s *EtcdV3) Exists(key string, opts *store.ReadOptions) (bool, error) {
	_, err := s.Get(key, opts)
	if err != nil {
		if err == store.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Watch for changes on a "key"
// It returns a channel that will receive changes or pass
// on errors. Upon creation, the current value will first
// be sent to the channel. Providing a non-nil stopCh can
// be used to stop watching.
func (s *EtcdV3) Watch(key string, stopCh <-chan struct{}, opts *store.ReadOptions) (<-chan *store.KVPair, error) {
	wc := etcd.NewWatcher(s.client)

	// respCh is sending back events to the caller
	respCh := make(chan *store.KVPair)

	// Get the current value
	pair, err := s.Get(key, opts)
	if err != nil {
		return nil, err
	}

	go func() {
		defer wc.Close()
		defer close(respCh)

		// Push the current value through the channel.
		respCh <- pair

		watchCh := wc.Watch(context.Background(), s.normalize(key))

		for resp := range watchCh {
			// Check if the watch was stopped by the caller
			select {
			case <-stopCh:
				return
			default:
			}

			for _, ev := range resp.Events {
				respCh <- &store.KVPair{
					Key:       key,
					Value:     []byte(ev.Kv.Value),
					LastIndex: uint64(ev.Kv.ModRevision),
				}
			}
		}
	}()

	return respCh, nil
}

// WatchTree watches for changes on a "directory"
// It returns a channel that will receive changes or pass
// on errors. Upon creating a watch, the current childs values
// will be sent to the channel. Providing a non-nil stopCh can
// be used to stop watching.
func (s *EtcdV3) WatchTree(directory string, stopCh <-chan struct{}, opts *store.ReadOptions) (<-chan []*store.KVPair, error) {
	wc := etcd.NewWatcher(s.client)

	// respCh is sending back events to the caller
	respCh := make(chan []*store.KVPair)

	// Get the current value
	rev, pairs, err := s.list(directory, opts)
	if err != nil {
		return nil, err
	}

	go func() {
		defer wc.Close()
		defer close(respCh)

		// Push the current value through the channel.
		respCh <- pairs

		rev++
		watchCh := wc.Watch(context.Background(), s.normalize(directory), etcd.WithPrefix(), etcd.WithRev(rev))

		for resp := range watchCh {
			// Check if the watch was stopped by the caller
			select {
			case <-stopCh:
				return
			default:
			}

			list := make([]*store.KVPair, len(resp.Events))

			for i, ev := range resp.Events {
				list[i] = &store.KVPair{
					Key:       string(ev.Kv.Key),
					Value:     []byte(ev.Kv.Value),
					LastIndex: uint64(ev.Kv.ModRevision),
				}
			}

			respCh <- list
		}
	}()

	return respCh, nil
}

// AtomicPut puts a value at "key" if the key has not been
// modified in the meantime, throws an error if this is the case
func (s *EtcdV3) AtomicPut(key string, value []byte, previous *store.KVPair, opts *store.WriteOptions) (bool, *store.KVPair, error) {
	var cmp etcd.Cmp
	var testIndex bool

	if previous != nil {
		// We compare on the last modified index
		testIndex = true
		cmp = etcd.Compare(etcd.ModRevision(key), "=", int64(previous.LastIndex))
	} else {
		// Previous key is not given, thus we want the key not to exist
		testIndex = false
		cmp = etcd.Compare(etcd.CreateRevision(key), "=", 0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), etcdDefaultTimeout)
	pr := s.client.Txn(ctx).If(cmp)

	// We set the TTL if given
	if opts != nil && opts.TTL > 0 {
		lease := etcd.NewLease(s.client)
		resp, err := lease.Grant(context.Background(), int64(opts.TTL/time.Second))
		if err != nil {
			cancel()
			return false, nil, err
		}
		pr.Then(etcd.OpPut(key, string(value), etcd.WithLease(resp.ID)))
	} else {
		pr.Then(etcd.OpPut(key, string(value)))
	}

	txn, err := pr.Commit()
	cancel()
	if err != nil {
		return false, nil, err
	}

	if !txn.Succeeded {
		if testIndex {
			return false, nil, store.ErrKeyModified
		}
		return false, nil, store.ErrKeyExists
	}

	updated := &store.KVPair{
		Key:       key,
		Value:     value,
		LastIndex: uint64(txn.Header.Revision),
	}

	return true, updated, nil
}

// AtomicDelete deletes a value at "key" if the key
// has not been modified in the meantime, throws an
// error if this is the case
func (s *EtcdV3) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	if previous == nil {
		return false, store.ErrPreviousNotSpecified
	}

	// We compare on the last modified index
	cmp := etcd.Compare(etcd.ModRevision(key), "=", int64(previous.LastIndex))

	ctx, cancel := context.WithTimeout(context.Background(), etcdDefaultTimeout)
	txn, err := s.client.Txn(ctx).
		If(cmp).
		Then(etcd.OpDelete(key)).
		Commit()
	cancel()

	if err != nil {
		return false, err
	}

	if len(txn.Responses) == 0 {
		return false, store.ErrKeyNotFound
	}

	if !txn.Succeeded {
		return false, store.ErrKeyModified
	}

	return true, nil
}

// List child nodes of a given directory
func (s *EtcdV3) List(directory string, opts *store.ReadOptions) ([]*store.KVPair, error) {
	_, kv, err := s.list(directory, opts)
	return kv, err
}

// DeleteTree deletes a range of keys under a given directory
func (s *EtcdV3) DeleteTree(directory string) error {
	ctx, cancel := context.WithTimeout(context.Background(), etcdDefaultTimeout)
	resp, err := s.client.KV.Delete(ctx, s.normalize(directory), etcd.WithPrefix())
	cancel()
	if resp.Deleted == 0 {
		return store.ErrKeyNotFound
	}
	return err
}

// NewLock returns a handle to a lock struct which can
// be used to provide mutual exclusion on a key
func (s *EtcdV3) NewLock(key string, options *store.LockOptions) (lock store.Locker, err error) {
	var value string
	ttl := defaultLockTTL
	renewCh := make(chan struct{})

	// Apply options on Lock
	if options != nil {
		if options.Value != nil {
			value = string(options.Value)
		}
		if options.TTL != 0 {
			ttl = options.TTL
		}
		if options.RenewLock != nil {
			renewCh = options.RenewLock
		}
	}

	// Create Session for Mutex
	session, err := concurrency.NewSession(s.client, concurrency.WithTTL(int(ttl/time.Second)))
	if err != nil {
		return nil, err
	}

	go func() {
		<-renewCh
		session.Close()
		return
	}()

	// A Mutex is a simple key that can only be held by a single process.
	// An etcd mutex behaves like a Zookeeper lock: a side key is created with
	// a suffix (such as "_lock") and represents the mutex. Thus we have a pair
	// composed of the key to protect with a lock: "/key", and a side key that
	// acts as the lock: "/key_lock"
	mutexKey := s.normalize(key + lockSuffix)
	writeKey := s.normalize(key)

	// Create lock object
	lock = &etcdLock{
		store:    s,
		mutex:    concurrency.NewMutex(session, mutexKey),
		session:  session,
		mutexKey: mutexKey,
		writeKey: writeKey,
		value:    value,
		ttl:      ttl,
	}

	return lock, nil
}

// Lock attempts to acquire the lock and blocks while
// doing so. It returns a channel that is closed if our
// lock is lost or if an error occurs
func (l *etcdLock) Lock(stopChan chan struct{}) (<-chan struct{}, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-stopChan
		cancel()
	}()
	err := l.mutex.Lock(ctx)
	if err != nil {
		if err == context.Canceled {
			return nil, nil
		}
		return nil, err
	}

	err = l.store.Put(l.writeKey, []byte(l.value), nil)
	if err != nil {
		return nil, err
	}

	return l.session.Done(), nil
}

// Unlock the "key". Calling unlock while
// not holding the lock will throw an error
func (l *etcdLock) Unlock() error {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.mutex.Unlock(context.Background())
}

// Close closes the client connection
func (s *EtcdV3) Close() {
	s.client.Close()
}

// list child nodes of a given directory and return revision number
func (s *EtcdV3) list(directory string, opts *store.ReadOptions) (int64, []*store.KVPair, error) {
	ctx, cancel := context.WithTimeout(context.Background(), etcdDefaultTimeout)

	var resp *etcd.GetResponse
	var err error

	if opts != nil && !opts.Consistent {
		resp, err = s.client.KV.Get(ctx, s.normalize(directory), etcd.WithSerializable(), etcd.WithPrefix(), etcd.WithSort(etcd.SortByKey, etcd.SortDescend))
	} else {
		resp, err = s.client.KV.Get(ctx, s.normalize(directory), etcd.WithPrefix(), etcd.WithSort(etcd.SortByKey, etcd.SortDescend))
	}

	cancel()

	if err != nil {
		return 0, nil, err
	}

	if resp.Count == 0 {
		return 0, nil, store.ErrKeyNotFound
	}

	kv := []*store.KVPair{}

	for _, n := range resp.Kvs {
		if string(n.Key) == directory {
			continue
		}

		// Filter out etcd mutex side keys with `___lock` suffix
		if strings.Contains(string(n.Key), lockSuffix) {
			continue
		}

		kv = append(kv, &store.KVPair{
			Key:       string(n.Key),
			Value:     []byte(n.Value),
			LastIndex: uint64(n.ModRevision),
		})
	}

	return resp.Header.Revision, kv, nil
}
