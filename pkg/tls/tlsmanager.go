package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/go-kit/kit/metrics"
	"github.com/sirupsen/logrus"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/tls/generate"
	"github.com/traefik/traefik/v2/pkg/types"
)

// DefaultTLSOptions the default TLS options.
var DefaultTLSOptions = Options{}

// Manager is the TLS option/store/configuration factory.
type Manager struct {
	storesConfig map[string]Store
	stores       map[string]*CertificateStore
	configs      map[string]Options
	certs        []*CertAndStores
	lock         sync.RWMutex

	metricsRegistry Registry
}

// Registry interface which defines the subset of methods needed in this package
type Registry interface {
	TLSCertsNotAfterTimestampGauge() metrics.Gauge
}

// NewManager creates a new Manager.
func NewManager(metricsRegistry Registry) *Manager {
	return &Manager{
		stores: map[string]*CertificateStore{},
		configs: map[string]Options{
			"default": DefaultTLSOptions,
		},
		metricsRegistry: metricsRegistry,
	}
}

// UpdateConfigs updates the TLS* configuration options.
func (m *Manager) UpdateConfigs(ctx context.Context, stores map[string]Store, configs map[string]Options, certs []*CertAndStores) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.configs = configs
	m.storesConfig = stores
	m.certs = certs

	m.stores = make(map[string]*CertificateStore)
	for storeName, storeConfig := range m.storesConfig {
		ctxStore := log.With(ctx, log.Str(log.TLSStoreName, storeName))
		store, err := buildCertificateStore(ctxStore, storeConfig)
		if err != nil {
			log.FromContext(ctxStore).Errorf("Error while creating certificate store: %v", err)
			continue
		}
		m.stores[storeName] = store
	}

	storesCertificates := make(map[string]map[string]*tls.Certificate)
	for _, conf := range certs {
		if len(conf.Stores) == 0 {
			if log.GetLevel() >= logrus.DebugLevel {
				log.FromContext(ctx).Debugf("No store is defined to add the certificate %s, it will be added to the default store.",
					conf.Certificate.GetTruncatedCertificateName())
			}
			conf.Stores = []string{"default"}
		}
		for _, store := range conf.Stores {
			ctxStore := log.With(ctx, log.Str(log.TLSStoreName, store))
			if err := conf.Certificate.AppendCertificate(storesCertificates, store); err != nil {
				log.FromContext(ctxStore).Errorf("Unable to append certificate %s to store: %v", conf.Certificate.GetTruncatedCertificateName(), err)
			}
		}
	}

	for storeName, certs := range storesCertificates {
		m.getStore(storeName).DynamicCerts.Set(certs)
		m.updateMetrics(&certs)
	}
}

// Get gets the TLS configuration to use for a given store / configuration.
func (m *Manager) Get(storeName, configName string) (*tls.Config, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var tlsConfig *tls.Config
	var err error

	config, ok := m.configs[configName]
	if !ok {
		err = fmt.Errorf("unknown TLS options: %s", configName)
		tlsConfig = &tls.Config{}
	}

	store := m.getStore(storeName)
	acmeTLSStore := m.getStore(tlsalpn01.ACMETLS1Protocol)

	if err == nil {
		tlsConfig, err = buildTLSConfig(config)
		if err != nil {
			tlsConfig = &tls.Config{}
		}
	}

	tlsConfig.GetCertificate = func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		domainToCheck := types.CanonicalDomain(clientHello.ServerName)

		if isACMETLS(clientHello) {
			certificate := acmeTLSStore.GetBestCertificate(clientHello)
			if certificate == nil {
				return nil, fmt.Errorf("no certificate for TLSALPN challenge: %s", domainToCheck)
			}

			return certificate, nil
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

	return tlsConfig, err
}

// updateMetrics produce metrics for a given *map[string]*tls.Certificate which must not be nil.
func (m *Manager) updateMetrics(certs *map[string]*tls.Certificate) {
	tlsNotAfter := m.metricsRegistry.TLSCertsNotAfterTimestampGauge()
	if tlsNotAfter != nil {
		for _, cert := range *certs {
			// We iterate over all the certificates, even the ones for the CA chain
			for i := range cert.Certificate {
				x509Cert, err := x509.ParseCertificate(cert.Certificate[i])

				if err != nil {
					continue
				}

				DNSNames := make([]string, 0, len(x509Cert.DNSNames))
				DNSNames = append(DNSNames, x509Cert.DNSNames...)
				sort.Strings(DNSNames)
				SANs := strings.Join(DNSNames, ",")

				labels := []string{
					"cn", x509Cert.Subject.CommonName,
					"serial", x509Cert.SerialNumber.String(),
					"sans", SANs,
				}

				tlsNotAfter.With(labels...).Set(float64(x509Cert.NotAfter.Unix()))
			}
		}
	}
}

func (m *Manager) getStore(storeName string) *CertificateStore {
	_, ok := m.stores[storeName]
	if !ok {
		m.stores[storeName], _ = buildCertificateStore(context.Background(), Store{})
	}
	return m.stores[storeName]
}

// GetStore gets the certificate store of a given name.
func (m *Manager) GetStore(storeName string) *CertificateStore {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.getStore(storeName)
}

func buildCertificateStore(ctx context.Context, tlsStore Store) (*CertificateStore, error) {
	certificateStore := NewCertificateStore()
	certificateStore.DynamicCerts.Set(make(map[string]*tls.Certificate))

	if tlsStore.DefaultCertificate != nil {
		cert, err := buildDefaultCertificate(tlsStore.DefaultCertificate)
		if err != nil {
			return certificateStore, err
		}
		certificateStore.DefaultCertificate = cert
	} else {
		log.FromContext(ctx).Debug("No default certificate, generating one")
		cert, err := generate.DefaultCertificate()
		if err != nil {
			return certificateStore, err
		}
		certificateStore.DefaultCertificate = cert
	}
	return certificateStore, nil
}

// creates a TLS config that allows terminating HTTPS for multiple domains using SNI.
func buildTLSConfig(tlsOption Options) (*tls.Config, error) {
	conf := &tls.Config{}

	// ensure http2 enabled
	conf.NextProtos = []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol}

	if len(tlsOption.ClientAuth.CAFiles) > 0 {
		pool := x509.NewCertPool()
		for _, caFile := range tlsOption.ClientAuth.CAFiles {
			data, err := caFile.Read()
			if err != nil {
				return nil, err
			}
			ok := pool.AppendCertsFromPEM(data)
			if !ok {
				if caFile.IsPath() {
					return nil, fmt.Errorf("invalid certificate(s) in %s", caFile)
				}
				return nil, errors.New("invalid certificate(s) content")
			}
		}
		conf.ClientCAs = pool
		conf.ClientAuth = tls.RequireAndVerifyClientCert
	}

	clientAuthType := tlsOption.ClientAuth.ClientAuthType
	if len(clientAuthType) > 0 {
		if conf.ClientCAs == nil && (clientAuthType == "VerifyClientCertIfGiven" ||
			clientAuthType == "RequireAndVerifyClientCert") {
			return nil, fmt.Errorf("invalid clientAuthType: %s, CAFiles is required", clientAuthType)
		}

		switch clientAuthType {
		case "NoClientCert":
			conf.ClientAuth = tls.NoClientCert
		case "RequestClientCert":
			conf.ClientAuth = tls.RequestClientCert
		case "RequireAnyClientCert":
			conf.ClientAuth = tls.RequireAnyClientCert
		case "VerifyClientCertIfGiven":
			conf.ClientAuth = tls.VerifyClientCertIfGiven
		case "RequireAndVerifyClientCert":
			conf.ClientAuth = tls.RequireAndVerifyClientCert
		default:
			return nil, fmt.Errorf("unknown client auth type %q", clientAuthType)
		}
	}

	// Set PreferServerCipherSuites.
	conf.PreferServerCipherSuites = tlsOption.PreferServerCipherSuites

	// Set the minimum TLS version if set in the config
	if minConst, exists := MinVersion[tlsOption.MinVersion]; exists {
		conf.PreferServerCipherSuites = true
		conf.MinVersion = minConst
	}

	// Set the maximum TLS version if set in the config TOML
	if maxConst, exists := MaxVersion[tlsOption.MaxVersion]; exists {
		conf.PreferServerCipherSuites = true
		conf.MaxVersion = maxConst
	}

	// Set the list of CipherSuites if set in the config
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

	// Set the list of CurvePreferences/CurveIDs if set in the config
	if tlsOption.CurvePreferences != nil {
		conf.CurvePreferences = make([]tls.CurveID, 0)
		// if our list of CurvePreferences/CurveIDs is defined in the config, we can re-initialize the list as empty
		for _, curve := range tlsOption.CurvePreferences {
			if curveID, exists := CurveIDs[curve]; exists {
				conf.CurvePreferences = append(conf.CurvePreferences, curveID)
			} else {
				// CurveID listed in the toml does not exist in our listed
				return nil, fmt.Errorf("invalid CurveID in curvePreferences: %s", curve)
			}
		}
	}

	return conf, nil
}

func buildDefaultCertificate(defaultCertificate *Certificate) (*tls.Certificate, error) {
	certFile, err := defaultCertificate.CertFile.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to get cert file content: %w", err)
	}

	keyFile, err := defaultCertificate.KeyFile.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to get key file content: %w", err)
	}

	cert, err := tls.X509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load X509 key pair: %w", err)
	}
	return &cert, nil
}

func isACMETLS(clientHello *tls.ClientHelloInfo) bool {
	for _, proto := range clientHello.SupportedProtos {
		if proto == tlsalpn01.ACMETLS1Protocol {
			return true
		}
	}

	return false
}
