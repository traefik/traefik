package cacheprovider

import (
	"context"
	"sync"

	"golang.org/x/crypto/acme/autocert"
)

func NewMemCacheProvider() T {
	return &memCacheProvider{}
}

type memCacheProvider struct {
	autocertCache autocert.Cache
}

func (p *memCacheProvider) GetAutoCertCache() autocert.Cache {
	if p.autocertCache == nil {
		p.autocertCache = &memAutoCertCache{
			kv: make(map[string][]byte),
		}
	}
	return p.autocertCache
}

type memAutoCertCache struct {
	kv  map[string][]byte
	mtx sync.Mutex
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (ng *memAutoCertCache) Get(ctx context.Context, key string) ([]byte, error) {
	ng.mtx.Lock()
	defer ng.mtx.Unlock()
	val, ok := ng.kv[key]
	if ok {
		return val, nil
	}
	return nil, autocert.ErrCacheMiss
}

// Put stores the data in the cache under the specified key.
// Inderlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (ng *memAutoCertCache) Put(ctx context.Context, key string, data []byte) error {
	ng.mtx.Lock()
	defer ng.mtx.Unlock()
	ng.kv[key] = data
	return nil
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (ng *memAutoCertCache) Delete(ctx context.Context, key string) error {
	ng.mtx.Lock()
	defer ng.mtx.Unlock()
	delete(ng.kv, key)
	return nil
}
