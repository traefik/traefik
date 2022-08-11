package kv

import (
	"context"

	"github.com/kvtools/valkeyrie/store"
	"github.com/traefik/traefik/v2/pkg/log"
)

type storeWrapper struct {
	store.Store
}

func (s *storeWrapper) Put(ctx context.Context, key string, value []byte, options *store.WriteOptions) error {
	log.WithoutContext().Debugf("Put: %s", key, string(value))

	if s.Store == nil {
		return nil
	}
	return s.Store.Put(ctx, key, value, options)
}

func (s *storeWrapper) Get(ctx context.Context, key string, options *store.ReadOptions) (*store.KVPair, error) {
	log.WithoutContext().Debugf("Get: %s", key)

	if s.Store == nil {
		return nil, nil
	}
	return s.Store.Get(ctx, key, options)
}

func (s *storeWrapper) Delete(ctx context.Context, key string) error {
	log.WithoutContext().Debugf("Delete: %s", key)

	if s.Store == nil {
		return nil
	}
	return s.Store.Delete(ctx, key)
}

func (s *storeWrapper) Exists(ctx context.Context, key string, options *store.ReadOptions) (bool, error) {
	log.WithoutContext().Debugf("Exists: %s", key)

	if s.Store == nil {
		return true, nil
	}
	return s.Store.Exists(ctx, key, options)
}

func (s *storeWrapper) Watch(ctx context.Context, key string, options *store.ReadOptions) (<-chan *store.KVPair, error) {
	log.WithoutContext().Debugf("Watch: %s", key)

	if s.Store == nil {
		return nil, nil
	}
	return s.Store.Watch(ctx, key, options)
}

func (s *storeWrapper) WatchTree(ctx context.Context, directory string, options *store.ReadOptions) (<-chan []*store.KVPair, error) {
	log.WithoutContext().Debugf("WatchTree: %s", directory)

	if s.Store == nil {
		return nil, nil
	}
	return s.Store.WatchTree(ctx, directory, options)
}

func (s *storeWrapper) NewLock(ctx context.Context, key string, options *store.LockOptions) (store.Locker, error) {
	log.WithoutContext().Debugf("NewLock: %s", key)

	if s.Store == nil {
		return nil, nil
	}
	return s.Store.NewLock(ctx, key, options)
}

func (s *storeWrapper) List(ctx context.Context, directory string, options *store.ReadOptions) ([]*store.KVPair, error) {
	log.WithoutContext().Debugf("List: %s", directory)

	if s.Store == nil {
		return nil, nil
	}
	return s.Store.List(ctx, directory, options)
}

func (s *storeWrapper) DeleteTree(ctx context.Context, directory string) error {
	log.WithoutContext().Debugf("DeleteTree: %s", directory)

	if s.Store == nil {
		return nil
	}
	return s.Store.DeleteTree(ctx, directory)
}

func (s *storeWrapper) AtomicPut(ctx context.Context, key string, value []byte, previous *store.KVPair, options *store.WriteOptions) (bool, *store.KVPair, error) {
	log.WithoutContext().Debugf("AtomicPut: %s", key, string(value), previous)

	if s.Store == nil {
		return true, nil, nil
	}
	return s.Store.AtomicPut(ctx, key, value, previous, options)
}

func (s *storeWrapper) AtomicDelete(ctx context.Context, key string, previous *store.KVPair) (bool, error) {
	log.WithoutContext().Debugf("AtomicDelete: %s", key, previous)

	if s.Store == nil {
		return true, nil
	}
	return s.Store.AtomicDelete(ctx, key, previous)
}

func (s *storeWrapper) Close() error {
	log.WithoutContext().Debugf("Close")

	if s.Store == nil {
		return nil
	}
	return s.Store.Close()
}
