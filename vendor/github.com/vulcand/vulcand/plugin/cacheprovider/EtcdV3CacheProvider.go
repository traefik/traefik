package cacheprovider

import (
	"context"

	etcd "github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
)

func NewEtcdV3CacheProvider(client *etcd.Client, vulcanPrefix string) T {
	return &etcdv3CacheProvider{
		client:       client,
		vulcanPrefix: vulcanPrefix,
	}
}

type etcdv3CacheProvider struct {
	client        *etcd.Client
	vulcanPrefix  string
	autoCertCache autocert.Cache
}

func (p *etcdv3CacheProvider) GetAutoCertCache() autocert.Cache {
	if p.autoCertCache == nil {
		p.autoCertCache = &etcdv3AutoCertCache{
			client: p.client,
			prefix: p.vulcanPrefix + "/autocert_cache/",
		}
	}
	return p.autoCertCache
}

type etcdv3AutoCertCache struct {
	client *etcd.Client
	prefix string
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (ng *etcdv3AutoCertCache) Get(ctx context.Context, rawKey string) ([]byte, error) {
	key := ng.normalize(rawKey)
	r, err := ng.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if r.Count == 0 {
		return nil, autocert.ErrCacheMiss
	}
	if r.Count > 1 {
		log.Errorf("Against all odds, multiple results returned from Etcd when looking for single key: %s. "+
			"Returning the first one.", key)
	}
	return r.Kvs[0].Value, nil
}

// Put stores the data in the cache under the specified key.
// Inderlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (ng *etcdv3AutoCertCache) Put(ctx context.Context, rawKey string, data []byte) error {
	key := ng.normalize(rawKey)
	_, err := ng.client.Put(ctx, key, string(data))
	return err
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (ng *etcdv3AutoCertCache) Delete(ctx context.Context, rawKey string) error {
	key := ng.normalize(rawKey)
	_, err := ng.client.Delete(ctx, key)
	return err
}

func (ng *etcdv3AutoCertCache) normalize(rawKey string) string {
	return ng.prefix + rawKey
}
