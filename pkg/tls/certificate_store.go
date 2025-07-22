package tls

import (
	"crypto/tls"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/safe"
)

// CertificateData holds runtime data for runtime TLS certificate handling.
type CertificateData struct {
	Hash        string
	Certificate *tls.Certificate
}

// CertificateStore store for dynamic certificates.
type CertificateStore struct {
	DynamicCerts       *safe.Safe
	DefaultCertificate *CertificateData
	CertCache          *cache.Cache

	ocspStapler *ocspStapler
}

// NewCertificateStore create a store for dynamic certificates.
func NewCertificateStore(ocspStapler *ocspStapler) *CertificateStore {
	var dynamicCerts safe.Safe
	dynamicCerts.Set(make(map[string]*CertificateData))

	return &CertificateStore{
		DynamicCerts: &dynamicCerts,
		CertCache:    cache.New(1*time.Hour, 10*time.Minute),
		ocspStapler:  ocspStapler,
	}
}

// GetAllDomains return a slice with all the certificate domain.
func (c *CertificateStore) GetAllDomains() []string {
	allDomains := c.getDefaultCertificateDomains()

	// Get dynamic certificates
	if c.DynamicCerts != nil && c.DynamicCerts.Get() != nil {
		for domain := range c.DynamicCerts.Get().(map[string]*CertificateData) {
			allDomains = append(allDomains, domain)
		}
	}

	return allDomains
}

// GetDefaultCertificate returns the default certificate.
func (c *CertificateStore) GetDefaultCertificate() *tls.Certificate {
	if c == nil {
		return nil
	}

	if c.ocspStapler != nil && c.DefaultCertificate.Hash != "" {
		if staple, ok := c.ocspStapler.GetStaple(c.DefaultCertificate.Hash); ok {
			// We are updating the OCSPStaple of the certificate without any synchronization
			// as this should not cause any issue.
			c.DefaultCertificate.Certificate.OCSPStaple = staple
		}
	}

	return c.DefaultCertificate.Certificate
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

	if cert, ok := c.CertCache.Get(serverName); ok {
		certificateData := cert.(*CertificateData)
		if c.ocspStapler != nil && certificateData.Hash != "" {
			if staple, ok := c.ocspStapler.GetStaple(certificateData.Hash); ok {
				// We are updating the OCSPStaple of the certificate without any synchronization
				// as this should not cause any issue.
				certificateData.Certificate.OCSPStaple = staple
			}
		}

		return certificateData.Certificate
	}

	matchedCerts := map[string]*CertificateData{}
	if c.DynamicCerts != nil && c.DynamicCerts.Get() != nil {
		for domains, cert := range c.DynamicCerts.Get().(map[string]*CertificateData) {
			for _, certDomain := range strings.Split(domains, ",") {
				if matchDomain(serverName, certDomain) {
					matchedCerts[certDomain] = cert
				}
			}
		}
	}

	if len(matchedCerts) > 0 {
		// sort map by keys
		keys := make([]string, 0, len(matchedCerts))
		for k := range matchedCerts {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// cache best match
		certificateData := matchedCerts[keys[len(keys)-1]]
		c.CertCache.SetDefault(serverName, certificateData)

		if c.ocspStapler != nil && certificateData.Hash != "" {
			if staple, ok := c.ocspStapler.GetStaple(certificateData.Hash); ok {
				// We are updating the OCSPStaple of the certificate without any synchronization
				// as this should not cause any issue.
				certificateData.Certificate.OCSPStaple = staple
			}
		}

		return certificateData.Certificate
	}

	return nil
}

// GetCertificate returns the first certificate matching all the given domains.
func (c *CertificateStore) GetCertificate(domains []string) *CertificateData {
	if c == nil {
		return nil
	}

	sort.Strings(domains)
	domainsKey := strings.Join(domains, ",")

	if cert, ok := c.CertCache.Get(domainsKey); ok {
		return cert.(*CertificateData)
	}

	if c.DynamicCerts != nil && c.DynamicCerts.Get() != nil {
		for certDomains, cert := range c.DynamicCerts.Get().(map[string]*CertificateData) {
			if domainsKey == certDomains {
				c.CertCache.SetDefault(domainsKey, cert)
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
				c.CertCache.SetDefault(domainsKey, cert)
				return cert
			}
		}
	}

	return nil
}

// ResetCache clears the cache in the store.
func (c *CertificateStore) ResetCache() {
	if c.CertCache != nil {
		c.CertCache.Flush()
	}
}

func (c *CertificateStore) getDefaultCertificateDomains() []string {
	if c.DefaultCertificate == nil {
		return nil
	}

	defaultCert := c.DefaultCertificate.Certificate.Leaf

	var allCerts []string
	if len(defaultCert.Subject.CommonName) > 0 {
		allCerts = append(allCerts, defaultCert.Subject.CommonName)
	}

	allCerts = append(allCerts, defaultCert.DNSNames...)

	for _, ipSan := range defaultCert.IPAddresses {
		allCerts = append(allCerts, ipSan.String())
	}

	return allCerts
}

// appendCertificate appends a Certificate to a certificates map keyed by store name.
func appendCertificate(certs map[string]map[string]*CertificateData, subjectAltNames []string, storeName string, cert *CertificateData) {
	// Guarantees the order to produce a unique cert key.
	sort.Strings(subjectAltNames)
	certKey := strings.Join(subjectAltNames, ",")

	certExists := false
	if certs[storeName] == nil {
		certs[storeName] = make(map[string]*CertificateData)
	} else {
		for domains := range certs[storeName] {
			if domains == certKey {
				certExists = true
				break
			}
		}
	}
	if certExists {
		log.Debug().Msgf("Skipping addition of certificate for domain(s) %q, to TLS Store %s, as it already exists for this store.", certKey, storeName)
	} else {
		log.Debug().Msgf("Adding certificate for domain(s) %s", certKey)

		certs[storeName][certKey] = cert
	}
}

func parseCertificate(cert *Certificate) (tls.Certificate, []string, error) {
	certContent, err := cert.CertFile.Read()
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("unable to read CertFile: %w", err)
	}

	keyContent, err := cert.KeyFile.Read()
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("unable to read KeyFile: %w", err)
	}

	tlsCert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("unable to generate TLS certificate: %w", err)
	}

	var SANs []string
	if tlsCert.Leaf.Subject.CommonName != "" {
		SANs = append(SANs, strings.ToLower(tlsCert.Leaf.Subject.CommonName))
	}
	if tlsCert.Leaf.DNSNames != nil {
		for _, dnsName := range tlsCert.Leaf.DNSNames {
			if dnsName != tlsCert.Leaf.Subject.CommonName {
				SANs = append(SANs, strings.ToLower(dnsName))
			}
		}
	}
	if tlsCert.Leaf.IPAddresses != nil {
		for _, ip := range tlsCert.Leaf.IPAddresses {
			if ip.String() != tlsCert.Leaf.Subject.CommonName {
				SANs = append(SANs, strings.ToLower(ip.String()))
			}
		}
	}

	return tlsCert, SANs, err
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
