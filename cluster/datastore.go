package cluster

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/cenkalti/backoff"
	"github.com/containous/staert"
	"github.com/docker/libkv/store"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"sync"
	"time"
)

// Object is the struct to store
type Object interface{}

// Metadata stores Object plus metadata
type Metadata struct {
	Lock string
}

// Datastore holds a struct synced in a KV store
type Datastore struct {
	kv        *staert.KvSource
	ctx       context.Context
	localLock *sync.RWMutex
	object    Object
	meta      *Metadata
	lockKey   string
}

// NewDataStore creates a Datastore
func NewDataStore(kvSource *staert.KvSource, ctx context.Context, object Object) (*Datastore, error) {
	datastore := Datastore{
		kv:        kvSource,
		ctx:       ctx,
		meta:      &Metadata{},
		object:    object,
		lockKey:   kvSource.Prefix + "/lock",
		localLock: &sync.RWMutex{},
	}
	err := datastore.watchChanges()
	if err != nil {
		return nil, err
	}
	return &datastore, nil
}

func (d *Datastore) watchChanges() error {
	stopCh := make(chan struct{})
	kvCh, err := d.kv.Watch(d.lockKey, stopCh)
	if err != nil {
		return err
	}
	go func() {
		ctx, cancel := context.WithCancel(d.ctx)
		operation := func() error {
			for {
				select {
				case <-ctx.Done():
					stopCh <- struct{}{}
					return nil
				case _, ok := <-kvCh:
					if !ok {
						cancel()
						return err
					}
					d.localLock.Lock()
					err := d.kv.LoadConfig(d.object)
					if err != nil {
						d.localLock.Unlock()
						return err
					}
					err = d.kv.LoadConfig(d.meta)
					if err != nil {
						d.localLock.Unlock()
						return err
					}
					d.localLock.Unlock()
				}
			}
		}
		notify := func(err error, time time.Duration) {
			log.Errorf("Error in watch datastore: %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
		if err != nil {
			log.Errorf("Error in watch datastore: %v", err)
		}
	}()
	return nil
}

// Begin creates a transaction with the KV store.
func (d *Datastore) Begin() (*Transaction, error) {
	id := uuid.NewV4().String()
	remoteLock, err := d.kv.NewLock(d.lockKey, &store.LockOptions{TTL: 20 * time.Second, Value: []byte(id)})
	if err != nil {
		return nil, err
	}
	stopCh := make(chan struct{})
	ctx, cancel := context.WithCancel(d.ctx)
	var errLock error
	go func() {
		_, errLock = remoteLock.Lock(stopCh)
		cancel()
	}()
	select {
	case <-ctx.Done():
		if errLock != nil {
			return nil, errLock
		}
	case <-d.ctx.Done():
		stopCh <- struct{}{}
		return nil, d.ctx.Err()
	}

	// we got the lock! Now make sure we are synced with KV store
	operation := func() error {
		meta := d.get()
		if meta.Lock != id {
			return fmt.Errorf("Object lock value: expected %s, got %s", id, meta.Lock)
		}
		return nil
	}
	notify := func(err error, time time.Duration) {
		log.Errorf("Datastore sync error: %v, retrying in %s", err, time)
	}
	ebo := backoff.NewExponentialBackOff()
	ebo.MaxElapsedTime = 60 * time.Second
	err = backoff.RetryNotify(operation, ebo, notify)
	if err != nil {
		return nil, fmt.Errorf("Datastore cannot sync: %v", err)
	}

	// we synced with KV store, we can now return Setter
	return &Transaction{
		Datastore:  d,
		remoteLock: remoteLock,
	}, nil
}

func (d *Datastore) get() *Metadata {
	d.localLock.RLock()
	defer d.localLock.RUnlock()
	return d.meta
}

// Get atomically a struct from the KV store
func (d *Datastore) Get() Object {
	d.localLock.RLock()
	defer d.localLock.RUnlock()
	return d.object
}

// Transaction allows to set a struct in the KV store
type Transaction struct {
	*Datastore
	remoteLock store.Locker
	dirty      bool
}

// Commit allows to set an object in the KV store
func (s *Transaction) Commit(object Object) error {
	s.localLock.Lock()
	defer s.localLock.Unlock()
	if s.dirty {
		return fmt.Errorf("Transaction already used. Please begin a new one.")
	}
	err := s.kv.StoreConfig(object)
	if err != nil {
		return err
	}

	err = s.remoteLock.Unlock()
	if err != nil {
		return err
	}

	s.Datastore.object = object
	s.dirty = true
	return nil
}
