package tls

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/tls/generate"
)

// CertificateStore store for dynamic certificates.
type CertificateStore struct {
	name          string
	certCache     *cache.Cache
	generatedCert *tls.Certificate

	lock               sync.RWMutex
	dynamicCerts       map[string]*tls.Certificate
	defaultCertificate *tls.Certificate
	config             Store
}

// NewCertificateStore create a store for dynamic certificates.
func NewCertificateStore(name string, config Store) (*CertificateStore, error) {
	generatedCert, err := generate.DefaultCertificate()
	if err != nil {
		return nil, err
	}

	return &CertificateStore{
		name:          name,
		certCache:     cache.New(1*time.Hour, 10*time.Minute),
		dynamicCerts:  make(map[string]*tls.Certificate),
		generatedCert: generatedCert,
		config:        config,
	}, nil
}

func (c *CertificateStore) getDefaultCertificateDomains() []string {
	var allCerts []string

	if c.defaultCertificate == nil {
		return allCerts
	}

	x509Cert, err := x509.ParseCertificate(c.defaultCertificate.Certificate[0])
	if err != nil {
		log.Error().Err(err).Msg("Could not parse default certificate")
		return allCerts
	}

	if len(x509Cert.Subject.CommonName) > 0 {
		allCerts = append(allCerts, x509Cert.Subject.CommonName)
	}

	allCerts = append(allCerts, x509Cert.DNSNames...)

	for _, ipSan := range x509Cert.IPAddresses {
		allCerts = append(allCerts, ipSan.String())
	}

	return allCerts
}

// GetAllDomains return a slice with all the certificate domain.
func (c *CertificateStore) GetAllDomains() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	allDomains := c.getDefaultCertificateDomains()

	// Get dynamic certificates
	for domain := range c.dynamicCerts {
		allDomains = append(allDomains, domain)
	}

	return allDomains
}

// GetBestCertificate returns the best match certificate, and caches the response.
func (c *CertificateStore) GetBestCertificate(clientHello *tls.ClientHelloInfo) *tls.Certificate {
	if c == nil {
		return nil
	}
	serverName := strings.ToLower(strings.TrimSpace(clientHello.ServerName))
	if len(serverName) == 0 {
		// If no ServerName is provided, Check for local IP address matches
		host, _, err := net.SplitHostPort(clientHello.Conn.LocalAddr().String())
		if err != nil {
			log.Debug().Err(err).Msg("Could not split host/port")
		}
		serverName = strings.TrimSpace(host)
	}

	if cert, ok := c.certCache.Get(serverName); ok {
		return cert.(*tls.Certificate)
	}

	matchedCerts := map[string]*tls.Certificate{}
	c.lock.RLock()
	for domains, cert := range c.dynamicCerts {
		for _, certDomain := range strings.Split(domains, ",") {
			if matchDomain(serverName, certDomain) {
				matchedCerts[certDomain] = cert
			}
		}
	}
	defer c.lock.RUnlock()

	if len(matchedCerts) > 0 {
		// sort map by keys
		keys := make([]string, 0, len(matchedCerts))
		for k := range matchedCerts {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// cache best match
		c.certCache.SetDefault(serverName, matchedCerts[keys[len(keys)-1]])
		return matchedCerts[keys[len(keys)-1]]
	}

	return nil
}

// GetCertificate returns the first certificate matching all the given domains.
func (c *CertificateStore) GetCertificate(domains []string) *tls.Certificate {
	if c == nil {
		return nil
	}

	sort.Strings(domains)
	domainsKey := strings.Join(domains, ",")

	if cert, ok := c.certCache.Get(domainsKey); ok {
		return cert.(*tls.Certificate)
	}

	c.lock.RLock()
	defer c.lock.RUnlock()
	for certDomains, cert := range c.dynamicCerts {
		if domainsKey == certDomains {
			c.certCache.SetDefault(domainsKey, cert)
			return cert
		}

		var matchedDomains []string
		for _, certDomain := range strings.Split(certDomains, ",") {
			for _, checkDomain := range domains {
				if certDomain == checkDomain {
					matchedDomains = append(matchedDomains, certDomain)
				}
			}
		}

		if len(matchedDomains) == len(domains) {
			c.certCache.SetDefault(domainsKey, cert)
			return cert
		}
	}

	return nil
}

// ResetCache clears the cache in the store.
func (c *CertificateStore) ResetCache() {
	if c.certCache != nil {
		c.certCache.Flush()
	}
}

func (c *CertificateStore) setCertificates(certificates []*tls.Certificate) {
	certKeyMap := certKeyMap(certificates...)

	c.lock.Lock()
	defer c.lock.Unlock()

	var toDelete []string
	for certKey := range c.dynamicCerts {
		if _, exists := certKeyMap[certKey]; !exists {
			toDelete = append(toDelete, certKey)
		}
	}

	for _, certKey := range toDelete {
		log.Debug().Msgf("Removing certificate for domain(s) %s", certKey)
		delete(c.dynamicCerts, certKey)
	}

	for certKey, cert := range certKeyMap {
		if storeCert, exists := c.dynamicCerts[certKey]; exists {
			if storeCert.Leaf.Equal(cert.Leaf) {
				log.Debug().Msgf("Skipping addition of certificate for domain(s) %q, to TLS Store %s, as it already exists for this store.", certKey, c.name)
				continue
			}
			// TODO - An option to control the behavior on multiple certs for the same domain could be added at some point
			// For ex.: using the latest one, or the one with the longest validity period.
			log.Warn().Msgf("Replacing certificate for domain(s) %q, in TLS Store %s.", certKey, c.name)
		}

		log.Debug().Msgf("Adding certificate for domain(s) %s", certKey)
		c.dynamicCerts[certKey] = cert
	}
}

func (c *CertificateStore) Certificates() []*x509.Certificate {
	c.lock.RLock()
	defer c.lock.RUnlock()

	certs := make([]*x509.Certificate, 0, len(c.dynamicCerts))
	for _, cert := range c.dynamicCerts {
		err := parseCertificate(cert)
		if err != nil {
			continue
		}

		certs = append(certs, cert.Leaf)
	}

	return certs
}

// certKeyMap returns a map of certificates with keys composed by certificate DNS names and IP addresses.
// Each certificate must already be parsed and contains the Leaf field, otherwise it is skipped and will be missing from the result map.
func certKeyMap(certs ...*tls.Certificate) map[string]*tls.Certificate {
	certKeyMap := make(map[string]*tls.Certificate, len(certs))
	for _, cert := range certs {
		if cert.Leaf == nil {
			continue
		}
		certKeyMap[orderedDomains(cert.Leaf)] = cert
	}
	return certKeyMap
}

// orderedDomains returns a sorted comma separated string with the domain names and addresses of the certificate.
func orderedDomains(tlsCert *x509.Certificate) string {
	var SANs []string
	if tlsCert.Subject.CommonName != "" {
		SANs = append(SANs, strings.ToLower(tlsCert.Subject.CommonName))
	}
	if tlsCert.DNSNames != nil {
		for _, dnsName := range tlsCert.DNSNames {
			if dnsName != tlsCert.Subject.CommonName {
				SANs = append(SANs, strings.ToLower(dnsName))
			}
		}
	}
	if tlsCert.IPAddresses != nil {
		for _, ip := range tlsCert.IPAddresses {
			if ip.String() != tlsCert.Subject.CommonName {
				SANs = append(SANs, strings.ToLower(ip.String()))
			}
		}
	}

	// Guarantees the order to produce a unique cert key.
	sort.Strings(SANs)
	certKey := strings.Join(SANs, ",")
	return certKey
}

// matchDomain returns whether the server name matches the cert domain.
// The server name, from TLS SNI, must not have trailing dots (https://datatracker.ietf.org/doc/html/rfc6066#section-3).
// This is enforced by https://github.com/golang/go/blob/d3d7998756c33f69706488cade1cd2b9b10a4c7f/src/crypto/tls/handshake_messages.go#L423-L427.
func matchDomain(serverName, certDomain string) bool {
	// TODO: assert equality after removing the trailing dots?
	if serverName == certDomain {
		return true
	}

	for len(certDomain) > 0 && certDomain[len(certDomain)-1] == '.' {
		certDomain = certDomain[:len(certDomain)-1]
	}

	labels := strings.Split(serverName, ".")
	for i := range labels {
		labels[i] = "*"
		candidate := strings.Join(labels, ".")
		if certDomain == candidate {
			return true
		}
	}
	return false
}
