package kv

import (
	"github.com/abronan/valkeyrie/store"
	"github.com/traefik/traefik/v2/pkg/log"
)

type storeWrapper struct {
	store.Store
}

func (s *storeWrapper) Put(key string, value []byte, options *store.WriteOptions) error {
	log.WithoutContext().Debugf("Put: %s", key, string(value))

	if s.Store == nil {
		return nil
	}
	return s.Store.Put(key, value, options)
}

func (s *storeWrapper) Get(key string, options *store.ReadOptions) (*store.KVPair, error) {
	log.WithoutContext().Debugf("Get: %s", key)

	if s.Store == nil {
		return nil, nil
	}
	return s.Store.Get(key, options)
}

func (s *storeWrapper) Delete(key string) error {
	log.WithoutContext().Debugf("Delete: %s", key)

	if s.Store == nil {
		return nil
	}
	return s.Store.Delete(key)
}

func (s *storeWrapper) Exists(key string, options *store.ReadOptions) (bool, error) {
	log.WithoutContext().Debugf("Exists: %s", key)

	if s.Store == nil {
		return true, nil
	}
	return s.Store.Exists(key, options)
}

func (s *storeWrapper) Watch(key string, stopCh <-chan struct{}, options *store.ReadOptions) (<-chan *store.KVPair, error) {
	log.WithoutContext().Debugf("Watch: %s", key)

	if s.Store == nil {
		return nil, nil
	}
	return s.Store.Watch(key, stopCh, options)
}

func (s *storeWrapper) WatchTree(directory string, stopCh <-chan struct{}, options *store.ReadOptions) (<-chan []*store.KVPair, error) {
	log.WithoutContext().Debugf("WatchTree: %s", directory)

	if s.Store == nil {
		return nil, nil
	}
	return s.Store.WatchTree(directory, stopCh, options)
}

func (s *storeWrapper) NewLock(key string, options *store.LockOptions) (store.Locker, error) {
	log.WithoutContext().Debugf("NewLock: %s", key)

	if s.Store == nil {
		return nil, nil
	}
	return s.Store.NewLock(key, options)
}

func (s *storeWrapper) List(directory string, options *store.ReadOptions) ([]*store.KVPair, error) {
	log.WithoutContext().Debugf("List: %s", directory)

	if s.Store == nil {
		return nil, nil
	}
	return s.Store.List(directory, options)
}

func (s *storeWrapper) DeleteTree(directory string) error {
	log.WithoutContext().Debugf("DeleteTree: %s", directory)

	if s.Store == nil {
		return nil
	}
	return s.Store.DeleteTree(directory)
}

func (s *storeWrapper) AtomicPut(key string, value []byte, previous *store.KVPair, options *store.WriteOptions) (bool, *store.KVPair, error) {
	log.WithoutContext().Debugf("AtomicPut: %s", key, string(value), previous)

	if s.Store == nil {
		return true, nil, nil
	}
	return s.Store.AtomicPut(key, value, previous, options)
}

func (s *storeWrapper) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	log.WithoutContext().Debugf("AtomicDelete: %s", key, previous)

	if s.Store == nil {
		return true, nil
	}
	return s.Store.AtomicDelete(key, previous)
}

func (s *storeWrapper) Close() {
	log.WithoutContext().Debugf("Close")

	if s.Store == nil {
		return
	}
	s.Store.Close()
}
