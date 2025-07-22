package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"hash/fnv"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/tls/generate"
	"github.com/traefik/traefik/v3/pkg/types"
)

const (
	// DefaultTLSConfigName is the name of the default set of options for configuring TLS.
	DefaultTLSConfigName = "default"
	// DefaultTLSStoreName is the name of the default store of TLS certificates.
	// Note that it actually is the only usable one for now.
	DefaultTLSStoreName = "default"
)

// DefaultTLSOptions the default TLS options.
var DefaultTLSOptions = Options{
	// ensure http2 enabled
	ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
	MinVersion:    "VersionTLS12",
	CipherSuites:  getCipherSuites(),
}

func getCipherSuites() []string {
	gsc := tls.CipherSuites()
	ciphers := make([]string, len(gsc))
	for idx, cs := range gsc {
		ciphers[idx] = cs.Name
	}
	return ciphers
}

// OCSPConfig contains the OCSP configuration.
type OCSPConfig struct {
	ResponderOverrides map[string]string `description:"Defines a map of OCSP responders to replace for querying OCSP servers." json:"responderOverrides,omitempty" toml:"responderOverrides,omitempty" yaml:"responderOverrides,omitempty"`
}

// Manager is the TLS option/store/configuration factory.
type Manager struct {
	lock         sync.RWMutex
	storesConfig map[string]Store
	stores       map[string]*CertificateStore
	configs      map[string]Options
	certs        []*CertAndStores

	// As of today, the TLS manager contains and is responsible for creating/starting the OCSP ocspStapler.
	// It would likely have been a Configuration listener but this implies that certs are re-parsed.
	// But this would probably have impact on resource consumption.
	ocspStapler *ocspStapler
}

// NewManager creates a new Manager.
func NewManager(ocspConfig *OCSPConfig) *Manager {
	manager := &Manager{
		stores: map[string]*CertificateStore{},
		configs: map[string]Options{
			"default": DefaultTLSOptions,
		},
	}

	if ocspConfig != nil {
		manager.ocspStapler = newOCSPStapler(ocspConfig.ResponderOverrides)
	}

	return manager
}

func (m *Manager) Run(ctx context.Context) {
	if m.ocspStapler != nil {
		m.ocspStapler.Run(ctx)
	}
}

// UpdateConfigs updates the TLS* configuration options.
// It initializes the default TLS store, and the TLS store for the ACME challenges.
func (m *Manager) UpdateConfigs(ctx context.Context, stores map[string]Store, configs map[string]Options, certs []*CertAndStores) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.configs = configs
	for optionName, option := range m.configs {
		// Handle `PreferServerCipherSuites` depreciation
		if option.PreferServerCipherSuites != nil {
			log.Ctx(ctx).Warn().Msgf("TLSOption %q uses `PreferServerCipherSuites` option, but this option is deprecated and ineffective, please remove this option.", optionName)
		}
	}

	m.storesConfig = stores
	m.certs = certs

	if m.storesConfig == nil {
		m.storesConfig = make(map[string]Store)
	}

	if _, ok := m.storesConfig[DefaultTLSStoreName]; !ok {
		m.storesConfig[DefaultTLSStoreName] = Store{}
	}

	if _, ok := m.storesConfig[tlsalpn01.ACMETLS1Protocol]; !ok {
		m.storesConfig[tlsalpn01.ACMETLS1Protocol] = Store{}
	}

	storesCertificates := make(map[string]map[string]*CertificateData)

	// Define the TTL for all the cache entries with no TTL.
	// This will discard entries that are not used anymore.
	if m.ocspStapler != nil {
		m.ocspStapler.ResetTTL()
	}

	for _, conf := range certs {
		if len(conf.Stores) == 0 {
			log.Ctx(ctx).Debug().MsgFunc(func() string {
				return fmt.Sprintf("No store is defined to add the certificate %s, it will be added to the default store",
					conf.Certificate.GetTruncatedCertificateName())
			})
			conf.Stores = []string{DefaultTLSStoreName}
		}

		cert, SANs, err := parseCertificate(&conf.Certificate)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msgf("Unable to parse certificate %s", conf.Certificate.GetTruncatedCertificateName())
			continue
		}

		var certHash string
		if m.ocspStapler != nil && len(cert.Leaf.OCSPServer) > 0 {
			certHash = hashRawCert(cert.Leaf.Raw)

			issuer := cert.Leaf
			if len(cert.Certificate) > 1 {
				issuer, err = x509.ParseCertificate(cert.Certificate[1])
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msgf("Unable to parse issuer certificate %s", conf.Certificate.GetTruncatedCertificateName())
					continue
				}
			}

			if err := m.ocspStapler.Upsert(certHash, cert.Leaf, issuer); err != nil {
				log.Ctx(ctx).Error().Err(err).Msgf("Unable to upsert OCSP certificate %s", conf.Certificate.GetTruncatedCertificateName())
				continue
			}
		}

		certData := &CertificateData{
			Certificate: &cert,
			Hash:        certHash,
		}

		for _, store := range conf.Stores {
			if _, ok := m.storesConfig[store]; !ok {
				m.storesConfig[store] = Store{}
			}

			appendCertificate(storesCertificates, SANs, store, certData)
		}
	}

	m.stores = make(map[string]*CertificateStore)

	for storeName, storeConfig := range m.storesConfig {
		st := NewCertificateStore(m.ocspStapler)
		m.stores[storeName] = st

		if certs, ok := storesCertificates[storeName]; ok {
			st.DynamicCerts.Set(certs)
		}

		// a default cert for the ACME store does not make any sense, so generating one is a waste.
		if storeName == tlsalpn01.ACMETLS1Protocol {
			continue
		}

		logger := log.Ctx(ctx).With().Str(logs.TLSStoreName, storeName).Logger()
		ctxStore := logger.WithContext(ctx)

		certificate, err := m.getDefaultCertificate(ctxStore, storeConfig, st)
		if err != nil {
			logger.Error().Err(err).Msg("Error while creating certificate store")
		}

		st.DefaultCertificate = certificate
	}

	if m.ocspStapler != nil {
		m.ocspStapler.ForceStapleUpdates()
	}
}

// sanitizeDomains sanitizes the domain definition Main and SANS,
// and returns them as a slice.
// This func apply the same sanitization as the ACME provider do before resolving certificates.
func sanitizeDomains(domain types.Domain) ([]string, error) {
	domains := domain.ToStrArray()
	if len(domains) == 0 {
		return nil, errors.New("no domain was given")
	}

	var cleanDomains []string
	for _, domain := range domains {
		canonicalDomain := types.CanonicalDomain(domain)
		cleanDomain := dns01.UnFqdn(canonicalDomain)
		cleanDomains = append(cleanDomains, cleanDomain)
	}

	return cleanDomains, nil
}

// Get gets the TLS configuration to use for a given store / configuration.
func (m *Manager) Get(storeName, configName string) (*tls.Config, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	sniStrict := false
	config, ok := m.configs[configName]
	if !ok {
		return nil, fmt.Errorf("unknown TLS options: %s", configName)
	}

	sniStrict = config.SniStrict
	tlsConfig, err := buildTLSConfig(config)
	if err != nil {
		return nil, fmt.Errorf("building TLS config: %w", err)
	}

	store := m.getStore(storeName)
	if store == nil {
		err = fmt.Errorf("TLS store %s not found", storeName)
	}
	acmeTLSStore := m.getStore(tlsalpn01.ACMETLS1Protocol)
	if acmeTLSStore == nil && err == nil {
		err = fmt.Errorf("ACME TLS store %s not found", tlsalpn01.ACMETLS1Protocol)
	}

	tlsConfig.GetCertificate = func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		domainToCheck := types.CanonicalDomain(clientHello.ServerName)

		if slices.Contains(clientHello.SupportedProtos, tlsalpn01.ACMETLS1Protocol) {
			certificate := acmeTLSStore.GetBestCertificate(clientHello)
			if certificate == nil {
				log.Debug().Msgf("TLS: no certificate for TLSALPN challenge: %s", domainToCheck)
				// We want the user to eventually get the (alertUnrecognizedName) "unrecognized name" error.
				// Unfortunately, if we returned an error here,
				// since we can't use the unexported error (errNoCertificates) that our caller (config.getCertificate in crypto/tls) uses as a sentinel,
				// it would report an (alertInternalError) "internal error" instead of an alertUnrecognizedName.
				// Which is why we return no error, and we let the caller detect that there's actually no certificate,
				// and fall back into the flow that will report the desired error.
				// https://cs.opensource.google/go/go/+/dev.boringcrypto.go1.17:src/crypto/tls/common.go;l=1058
				return nil, nil
			}

			return certificate, nil
		}

		bestCertificate := store.GetBestCertificate(clientHello)
		if bestCertificate != nil {
			return bestCertificate, nil
		}

		if sniStrict {
			log.Debug().Msgf("TLS: strict SNI enabled - No certificate found for domain: %q, closing connection", domainToCheck)
			// Same comment as above, as in the isACMETLS case.
			return nil, nil
		}

		if store == nil {
			log.Error().Msgf("TLS: No certificate store found with this name: %q, closing connection", storeName)

			// Same comment as above, as in the isACMETLS case.
			return nil, nil
		}

		log.Debug().Msgf("Serving default certificate for request: %q", domainToCheck)

		return store.GetDefaultCertificate(), nil
	}

	return tlsConfig, err
}

// GetServerCertificates returns all certificates from the default store,
// as well as the user-defined default certificate (if it exists).
func (m *Manager) GetServerCertificates() []*x509.Certificate {
	var certificates []*x509.Certificate

	// The default store is the only relevant, because it is the only one configurable.
	defaultStore, ok := m.stores[DefaultTLSStoreName]
	if !ok || defaultStore == nil {
		return certificates
	}

	// We iterate over all the certificates.
	if defaultStore.DynamicCerts != nil && defaultStore.DynamicCerts.Get() != nil {
		for _, cert := range defaultStore.DynamicCerts.Get().(map[string]*CertificateData) {
			x509Cert, err := x509.ParseCertificate(cert.Certificate.Certificate[0])
			if err != nil {
				continue
			}

			certificates = append(certificates, x509Cert)
		}
	}

	if defaultStore.DefaultCertificate != nil {
		x509Cert, err := x509.ParseCertificate(defaultStore.DefaultCertificate.Certificate.Certificate[0])
		if err != nil {
			return certificates
		}

		// Excluding the generated Traefik default certificate.
		if x509Cert.Subject.CommonName == generate.DefaultDomain {
			return certificates
		}

		certificates = append(certificates, x509Cert)
	}

	return certificates
}

// getStore returns the store found for storeName, or nil otherwise.
func (m *Manager) getStore(storeName string) *CertificateStore {
	st, ok := m.stores[storeName]
	if !ok {
		return nil
	}
	return st
}

// GetStore gets the certificate store of a given name.
func (m *Manager) GetStore(storeName string) *CertificateStore {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.getStore(storeName)
}

func (m *Manager) getDefaultCertificate(ctx context.Context, tlsStore Store, st *CertificateStore) (*CertificateData, error) {
	if tlsStore.DefaultCertificate != nil {
		cert, err := m.buildDefaultCertificate(tlsStore.DefaultCertificate)
		if err != nil {
			return nil, err
		}

		return cert, nil
	}

	defaultCert, err := generate.DefaultCertificate()
	if err != nil {
		return nil, err
	}

	defaultCertificate := &CertificateData{
		Certificate: defaultCert,
	}

	if tlsStore.DefaultGeneratedCert != nil && tlsStore.DefaultGeneratedCert.Domain != nil && tlsStore.DefaultGeneratedCert.Resolver != "" {
		domains, err := sanitizeDomains(*tlsStore.DefaultGeneratedCert.Domain)
		if err != nil {
			return defaultCertificate, fmt.Errorf("falling back to the internal generated certificate because invalid domains: %w", err)
		}

		defaultACMECert := st.GetCertificate(domains)
		if defaultACMECert == nil {
			return defaultCertificate, fmt.Errorf("unable to find certificate for domains %q: falling back to the internal generated certificate", strings.Join(domains, ","))
		}

		return defaultACMECert, nil
	}

	log.Ctx(ctx).Debug().Msg("No default certificate, fallback to the internal generated certificate")
	return defaultCertificate, nil
}

func (m *Manager) buildDefaultCertificate(defaultCertificate *Certificate) (*CertificateData, error) {
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

	var certHash string
	if m.ocspStapler != nil && len(cert.Leaf.OCSPServer) > 0 {
		certHash = hashRawCert(cert.Leaf.Raw)

		issuer := cert.Leaf
		if len(cert.Certificate) > 1 {
			issuer, err = x509.ParseCertificate(cert.Certificate[1])
			if err != nil {
				return nil, fmt.Errorf("parsing issuer certificate %s: %w", defaultCertificate.GetTruncatedCertificateName(), err)
			}
		}

		if err := m.ocspStapler.Upsert(certHash, cert.Leaf, issuer); err != nil {
			return nil, fmt.Errorf("upserting OCSP certificate %s: %w", defaultCertificate.GetTruncatedCertificateName(), err)
		}
	}

	return &CertificateData{
		Certificate: &cert,
		Hash:        certHash,
	}, nil
}

// creates a TLS config that allows terminating HTTPS for multiple domains using SNI.
func buildTLSConfig(tlsOption Options) (*tls.Config, error) {
	conf := &tls.Config{
		NextProtos:             tlsOption.ALPNProtocols,
		SessionTicketsDisabled: tlsOption.DisableSessionTickets,
	}

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

	// Set the minimum TLS version if set in the config
	if minConst, exists := MinVersion[tlsOption.MinVersion]; exists {
		conf.MinVersion = minConst
	}

	// Set the maximum TLS version if set in the config TOML
	if maxConst, exists := MaxVersion[tlsOption.MaxVersion]; exists {
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

func hashRawCert(rawCert []byte) string {
	hasher := fnv.New64()

	// purposely ignoring the error, as no error can be returned from the implementation.
	_, _ = hasher.Write(rawCert)
	return strconv.FormatUint(hasher.Sum64(), 16)
}
