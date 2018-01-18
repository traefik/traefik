package cacheprovider

import "golang.org/x/crypto/acme/autocert"

type noOpCacheProvider struct{}

func (*noOpCacheProvider) GetAutoCertCache() autocert.Cache { return nil }

func NoOp() T {
	return &noOpCacheProvider{}
}
