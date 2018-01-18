package cacheprovider

import (
	"context"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/crypto/acme/autocert"
)

func NewEtcdV2CacheProvider(kapi etcd.KeysAPI, vulcanPrefix string) T {
	return &etcdv2CacheProvider{
		kapi:         kapi,
		vulcanPrefix: vulcanPrefix,
	}
}

type etcdv2CacheProvider struct {
	kapi          etcd.KeysAPI
	vulcanPrefix  string
	autoCertCache autocert.Cache
}

func (p *etcdv2CacheProvider) GetAutoCertCache() autocert.Cache {
	if p.autoCertCache == nil {
		p.autoCertCache = &etcdv2AutoCertCache{
			kapi:   p.kapi,
			prefix: p.vulcanPrefix + "/autocert_cache/",
		}
	}
	return p.autoCertCache
}

type etcdv2AutoCertCache struct {
	kapi   etcd.KeysAPI
	prefix string
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (ng *etcdv2AutoCertCache) Get(ctx context.Context, rawKey string) ([]byte, error) {
	key := ng.normalized(rawKey)
	r, err := ng.kapi.Get(ctx, key, &etcd.GetOptions{})
	if err != nil {
		return nil, err
	}
	if r.Node == nil {
		return nil, autocert.ErrCacheMiss
	}
	return []byte(r.Node.Value), nil
}

// Put stores the data in the cache under the specified key.
// Inderlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (ng *etcdv2AutoCertCache) Put(ctx context.Context, rawKey string, data []byte) error {
	key := ng.normalized(rawKey)
	_, err := ng.kapi.Set(ctx, key, string(data), &etcd.SetOptions{})
	return err
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (ng *etcdv2AutoCertCache) Delete(ctx context.Context, rawKey string) error {
	key := ng.normalized(rawKey)
	_, err := ng.kapi.Delete(ctx, key, &etcd.DeleteOptions{})
	return err
}

func (ng *etcdv2AutoCertCache) normalized(key string) string {
	return ng.prefix + key
}
