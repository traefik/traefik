// Package memcached implements a HTTP provider for solving the HTTP-01 challenge using memcached
// in combination with a webserver.
package memcached

import (
	"fmt"
	"path"

	"github.com/rainycape/memcache"
	"github.com/xenolf/lego/acme"
)

// HTTPProvider implements ChallengeProvider for `http-01` challenge
type MemcachedProvider struct {
	hosts []string
}

// NewHTTPProvider returns a HTTPProvider instance with a configured webroot path
func NewMemcachedProvider(hosts []string) (*MemcachedProvider, error) {
	if len(hosts) == 0 {
		return nil, fmt.Errorf("No memcached hosts provided")
	}

	c := &MemcachedProvider{
		hosts: hosts,
	}

	return c, nil
}

// Present makes the token available at `HTTP01ChallengePath(token)` by creating a file in the given webroot path
func (w *MemcachedProvider) Present(domain, token, keyAuth string) error {
	var errs []error

	challengePath := path.Join("/", acme.HTTP01ChallengePath(token))
	for _, host := range w.hosts {
		mc, err := memcache.New(host)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		mc.Add(&memcache.Item{
			Key:        challengePath,
			Value:      []byte(keyAuth),
			Expiration: 60,
		})
	}

	if len(errs) == len(w.hosts) {
		return fmt.Errorf("Unable to store key in any of the memcache hosts -> %v", errs)
	}

	return nil
}

// CleanUp removes the file created for the challenge
func (w *MemcachedProvider) CleanUp(domain, token, keyAuth string) error {
	// Memcached will clean up itself, that's what expiration is for.
	return nil
}
