package engine

import (
	"crypto/tls"
	"fmt"
)

// TLSSettings is a JSON and API friendly version of some of the tls.Config parameters
type TLSSettings struct {
	// PreferServerCipherSuites controls whether the server selects the CipherSuites
	PreferServerCipherSuites bool

	// SkipVerify skips certificate check, very insecure
	InsecureSkipVerify bool

	// MinVersion is minimal TLS version "VersionTLS10" is default
	MinVersion string

	// MaxVersion is max supported TLS version, "VersionTLS12" is default
	MaxVersion string

	// SessionTicketsDisabled disables session ticket resumption support
	SessionTicketsDisabled bool

	// SessionCache specifies TLS session cache, default it 'LRU' with capacity 1024
	SessionCache TLSSessionCache

	// Preferred CipherSuites, default is:
	//
	// TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	// TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	// TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA
	// TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA
	// TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA
	// TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA
	// TLS_RSA_WITH_AES_256_CBC_SHA
	// TLS_RSA_WITH_AES_128_CBC_SHA
	CipherSuites []string
}

// TLSSessionCache sets up parameters for TLS session cache
type TLSSessionCache struct {
	// Session cache implementation, default is 'LRU' with capacity 1024
	Type     string
	Settings *LRUSessionCacheSettings
}

type LRUSessionCacheSettings struct {
	// LRU Capacity default is 1024
	Capacity int
}

// NewTLSConfig validates the TLSSettings and returns the tls.Config with the converted parameters
func NewTLSConfig(s *TLSSettings) (*tls.Config, error) {
	// Parse min and max TLS versions
	var min, max uint16
	var err error

	if s.MinVersion == "" {
		min = tls.VersionTLS10
	} else if min, err = ParseTLSVersion(s.MinVersion); err != nil {
		return nil, err
	}

	if s.MaxVersion == "" {
		max = tls.VersionTLS12
	} else if max, err = ParseTLSVersion(s.MaxVersion); err != nil {
		return nil, err
	}

	var css []uint16
	if len(s.CipherSuites) == 0 {
		css = []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,

			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,

			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		}
	} else {
		css = make([]uint16, len(s.CipherSuites))
		for i, suite := range s.CipherSuites {
			cs, err := ParseCipherSuite(suite)
			if err != nil {
				return nil, err
			}
			css[i] = cs
		}
	}

	var cache tls.ClientSessionCache
	if !s.SessionTicketsDisabled {
		cache, err = NewTLSSessionCache(&s.SessionCache)
		if err != nil {
			return nil, err
		}
	}

	return &tls.Config{
		MinVersion: min,
		MaxVersion: max,

		SessionTicketsDisabled: s.SessionTicketsDisabled,
		ClientSessionCache:     cache,

		PreferServerCipherSuites: s.PreferServerCipherSuites,
		CipherSuites:             css,

		InsecureSkipVerify: s.InsecureSkipVerify,
	}, nil
}

// NewTLSSessionCache validates parameters and creates a new TLS session cache
func NewTLSSessionCache(s *TLSSessionCache) (tls.ClientSessionCache, error) {
	cacheType := s.Type
	if cacheType == "" {
		cacheType = LRUCacheType
	}
	if cacheType != LRUCacheType {
		return nil, fmt.Errorf("unsupported session cache type: %v", s.Type)
	}
	var capacity int
	if params := s.Settings; params != nil {
		if params.Capacity < 0 {
			return nil, fmt.Errorf("bad LRU capacity: %v", params.Capacity)
		} else if params.Capacity == 0 {
			capacity = DefaultLRUCapacity
		} else {
			capacity = params.Capacity
		}
	}
	return tls.NewLRUClientSessionCache(capacity), nil
}

func ParseCipherSuite(cs string) (uint16, error) {
	switch cs {
	case "TLS_RSA_WITH_RC4_128_SHA":
		return tls.TLS_RSA_WITH_RC4_128_SHA, nil
	case "TLS_RSA_WITH_3DES_EDE_CBC_SHA":
		return tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA, nil
	case "TLS_RSA_WITH_AES_128_CBC_SHA":
		return tls.TLS_RSA_WITH_AES_128_CBC_SHA, nil
	case "TLS_RSA_WITH_AES_256_CBC_SHA":
		return tls.TLS_RSA_WITH_AES_256_CBC_SHA, nil
	case "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA":
		return tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA, nil
	case "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":
		return tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA, nil
	case "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":
		return tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA, nil
	case "TLS_ECDHE_RSA_WITH_RC4_128_SHA":
		return tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA, nil
	case "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":
		return tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA, nil
	case "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":
		return tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA, nil
	case "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":
		return tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA, nil
	case "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":
		return tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, nil
	case "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":
		return tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, nil
	}
	return 0, fmt.Errorf("unsupported cipher suite: %v", cs)
}

func ParseTLSVersion(version string) (uint16, error) {
	switch version {
	case "VersionTLS10":
		return tls.VersionTLS10, nil
	case "VersionTLS11":
		return tls.VersionTLS11, nil
	case "VersionTLS12":
		return tls.VersionTLS12, nil
	}
	return 0, fmt.Errorf("unsupported TLS version: %v", version)
}

const DefaultLRUCapacity = 1024
const LRUCacheType = "LRU"

func (s *TLSSettings) Equals(other *TLSSettings) bool {
	scfg, err := NewTLSConfig(s)
	if err != nil {
		return false
	}

	ocfg, err := NewTLSConfig(other)
	if err != nil {
		return false
	}

	if scfg.PreferServerCipherSuites != ocfg.PreferServerCipherSuites ||
		scfg.InsecureSkipVerify != ocfg.InsecureSkipVerify ||
		scfg.MinVersion != ocfg.MinVersion ||
		scfg.MaxVersion != ocfg.MaxVersion ||
		scfg.SessionTicketsDisabled != ocfg.SessionTicketsDisabled {
		return false
	}

	if len(scfg.CipherSuites) != len(ocfg.CipherSuites) {
		return false
	}

	for i := range scfg.CipherSuites {
		if scfg.CipherSuites[i] != ocfg.CipherSuites[i] {
			return false
		}
	}
	if !(&s.SessionCache).Equals(&other.SessionCache) {
		return false
	}

	return true
}

func (c *TLSSessionCache) Equals(o *TLSSessionCache) bool {
	if c.Type != o.Type {
		return false
	}

	if c.Settings == nil && o.Settings == nil {
		return true
	}

	cs, os := c.Settings, o.Settings

	if (cs == nil && os != nil) || (cs != nil && os == nil) {
		return false
	}

	if cs == nil && os == nil {
		return true
	}

	return cs.Capacity == os.Capacity
}
