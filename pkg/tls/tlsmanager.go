package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"

	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/tls/generate"
	"github.com/containous/traefik/pkg/types"
	"github.com/go-acme/lego/challenge/tlsalpn01"
	"github.com/sirupsen/logrus"
)

// Manager is the TLS option/store/configuration factory
type Manager struct {
	storesConfig  map[string]Store
	stores        map[string]*CertificateStore
	configs       map[string]TLS
	certs         []*Configuration
	TLSAlpnGetter func(string) (*tls.Certificate, error)
	lock          sync.RWMutex
}

// NewManager creates a new Manager
func NewManager() *Manager {
	return &Manager{}
}

// UpdateConfigs updates the TLS* configuration options
func (m *Manager) UpdateConfigs(stores map[string]Store, configs map[string]TLS, certs []*Configuration) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.configs = configs
	m.storesConfig = stores
	m.certs = certs

	m.stores = make(map[string]*CertificateStore)
	for storeName, storeConfig := range m.storesConfig {
		store, err := buildCertificateStore(storeConfig)
		if err != nil {
			log.Errorf("Error while creating certificate store %s: %v", storeName, err)
			continue
		}
		m.stores[storeName] = store
	}

	storesCertificates := make(map[string]map[string]*tls.Certificate)
	for _, conf := range certs {
		if len(conf.Stores) == 0 {
			if log.GetLevel() >= logrus.DebugLevel {
				log.Debugf("No store is defined to add the certificate %s, it will be added to the default store.",
					conf.Certificate.GetTruncatedCertificateName())
			}
			conf.Stores = []string{"default"}
		}
		for _, store := range conf.Stores {
			if err := conf.Certificate.AppendCertificate(storesCertificates, store); err != nil {
				log.Errorf("Unable to append certificate %s to store %s: %v", conf.Certificate.GetTruncatedCertificateName(), store, err)
			}
		}
	}

	for storeName, certs := range storesCertificates {
		m.getStore(storeName).DynamicCerts.Set(certs)
	}
}

// Get gets the TLS configuration to use for a given store / configuration
func (m *Manager) Get(storeName string, configName string) (*tls.Config, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	config, ok := m.configs[configName]
	if !ok && configName != "default" {
		return nil, fmt.Errorf("unknown TLS options: %s", configName)
	}

	store := m.getStore(storeName)

	tlsConfig, err := buildTLSConfig(config)
	if err != nil {
		log.Error(err)
		tlsConfig = &tls.Config{}
	}

	tlsConfig.GetCertificate = func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		domainToCheck := types.CanonicalDomain(clientHello.ServerName)

		if m.TLSAlpnGetter != nil {
			cert, err := m.TLSAlpnGetter(domainToCheck)
			if err != nil {
				return nil, err
			}

			if cert != nil {
				return cert, nil
			}
		}

		bestCertificate := store.GetBestCertificate(clientHello)
		if bestCertificate != nil {
			return bestCertificate, nil
		}

		if m.configs[configName].SniStrict {
			return nil, fmt.Errorf("strict SNI enabled - No certificate found for domain: %q, closing connection", domainToCheck)
		}

		log.WithoutContext().Debugf("Serving default certificate for request: %q", domainToCheck)
		return store.DefaultCertificate, nil
	}
	return tlsConfig, nil
}

func (m *Manager) getStore(storeName string) *CertificateStore {
	_, ok := m.stores[storeName]
	if !ok {
		m.stores[storeName], _ = buildCertificateStore(Store{})
	}
	return m.stores[storeName]
}

// GetStore gets the certificate store of a given name
func (m *Manager) GetStore(storeName string) *CertificateStore {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.getStore(storeName)
}

func buildCertificateStore(tlsStore Store) (*CertificateStore, error) {
	certificateStore := NewCertificateStore()
	certificateStore.DynamicCerts.Set(make(map[string]*tls.Certificate))

	if tlsStore.DefaultCertificate != nil {
		cert, err := buildDefaultCertificate(tlsStore.DefaultCertificate)
		if err != nil {
			return certificateStore, err
		}
		certificateStore.DefaultCertificate = cert
	} else {
		log.Debug("No default certificate, generate one")
		cert, err := generate.DefaultCertificate()
		if err != nil {
			return certificateStore, err
		}
		certificateStore.DefaultCertificate = cert
	}
	return certificateStore, nil
}

// creates a TLS config that allows terminating HTTPS for multiple domains using SNI
func buildTLSConfig(tlsOption TLS) (*tls.Config, error) {
	conf := &tls.Config{}

	// ensure http2 enabled
	conf.NextProtos = []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol}

	if len(tlsOption.ClientCA.Files) > 0 {
		pool := x509.NewCertPool()
		for _, caFile := range tlsOption.ClientCA.Files {
			data, err := caFile.Read()
			if err != nil {
				return nil, err
			}
			ok := pool.AppendCertsFromPEM(data)
			if !ok {
				return nil, fmt.Errorf("invalid certificate(s) in %s", caFile)
			}
		}
		conf.ClientCAs = pool
		if tlsOption.ClientCA.Optional {
			conf.ClientAuth = tls.VerifyClientCertIfGiven
		} else {
			conf.ClientAuth = tls.RequireAndVerifyClientCert
		}
	}

	// Set the minimum TLS version if set in the config TOML
	if minConst, exists := MinVersion[tlsOption.MinVersion]; exists {
		conf.PreferServerCipherSuites = true
		conf.MinVersion = minConst
	}

	// Set the list of CipherSuites if set in the config TOML
	if tlsOption.CipherSuites != nil {
		// if our list of CipherSuites is defined in the entryPoint config, we can re-initialize the suites list as empty
		conf.CipherSuites = make([]uint16, 0)
		for _, cipher := range tlsOption.CipherSuites {
			if cipherConst, exists := CipherSuites[cipher]; exists {
				conf.CipherSuites = append(conf.CipherSuites, cipherConst)
			} else {
				// CipherSuite listed in the toml does not exist in our listed
				return nil, fmt.Errorf("invalid CipherSuite: %s", cipher)
			}
		}
	}

	return conf, nil
}

func buildDefaultCertificate(defaultCertificate *Certificate) (*tls.Certificate, error) {
	certFile, err := defaultCertificate.CertFile.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to get cert file content: %v", err)
	}

	keyFile, err := defaultCertificate.KeyFile.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to get key file content: %v", err)
	}

	cert, err := tls.X509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load X509 key pair: %v", err)
	}
	return &cert, nil
}
