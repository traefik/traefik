package cacheprovider

import "golang.org/x/crypto/acme/autocert"

type T interface {
	GetAutoCertCache() autocert.Cache
}
