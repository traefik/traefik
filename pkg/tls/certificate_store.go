package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/containous/traefik/v2/pkg/tls/certificate"
	"github.com/patrickmn/go-cache"
)

const (
	certTypeDelimiter = "/"
)

// certificateKey contains information by which certificates are tracked in certificate cache
type certificateKey struct {
	hostname string
	certType certificate.CertificateType
}

func (c certificateKey) String() string {
	return fmt.Sprintf("%s (%s)", c.hostname, c.certType)
}

// CertificateStore store for dynamic and static certificates
type CertificateStore struct {
	DynamicCerts        *safe.Safe
	DefaultCertificates []*tls.Certificate
	CertCache           *cache.Cache
}

// NewCertificateStore create a store for dynamic and static certificates
func NewCertificateStore() *CertificateStore {
	return &CertificateStore{
		DynamicCerts: &safe.Safe{},
		CertCache:    cache.New(1*time.Hour, 10*time.Minute),
	}
}

func (c CertificateStore) getDefaultCertificateDomains() []string {
	var allCerts []string

	if len(c.DefaultCertificates) < 1 {
		return allCerts
	}

	allCertsMap := map[string]bool{}

	for _, cert := range c.DefaultCertificates {
		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			log.WithoutContext().Errorf("Could not parse default certicate: %v", err)
			break
		}

		if len(x509Cert.Subject.CommonName) > 0 {
			allCertsMap[x509Cert.Subject.CommonName] = true
		}

		for _, name := range x509Cert.DNSNames {
			allCertsMap[name] = true
		}

		for _, ipSan := range x509Cert.IPAddresses {
			allCertsMap[ipSan.String()] = true
		}
	}

	for name := range allCertsMap {
		allCerts = append(allCerts, name)
	}

	return allCerts
}

// GetAllDomains return a slice with all the certificate domain
func (c CertificateStore) GetAllDomains() []string {
	allCerts := c.getDefaultCertificateDomains()

	// Get dynamic certificates
	if c.DynamicCerts != nil && c.DynamicCerts.Get() != nil {
		for domains := range c.DynamicCerts.Get().(map[string]*tls.Certificate) {
			allCerts = append(allCerts, domains)
		}
	}
	return allCerts
}

func getCertTypeForClientHello(hello *tls.ClientHelloInfo) (retval certificate.CertificateType) {
	retval = certificate.RSA

	// The "signature_algorithms" extension, if present, limits the key exchange
	// algorithms allowed by the cipher suites. See RFC 5246, section 7.4.1.4.1.
	if hello.SignatureSchemes != nil {
	schemeLoop:
		for _, scheme := range hello.SignatureSchemes {
			switch scheme {
			case tls.ECDSAWithSHA1,
				tls.ECDSAWithP256AndSHA256,
				tls.ECDSAWithP384AndSHA384,
				tls.ECDSAWithP521AndSHA512,
				tls.Ed25519:
				retval = certificate.EC
				break schemeLoop
			}
		}
		if retval != certificate.EC {
			return
		}
	}
	if hello.SupportedCurves != nil {
		retval = certificate.RSA
	curveLoop:
		for _, curve := range hello.SupportedCurves {
			switch curve {
			case tls.CurveP256,
				tls.CurveP384,
				tls.CurveP521,
				tls.X25519:
				retval = certificate.EC
				break curveLoop
			}
		}
		if retval != certificate.EC {
			return
		}
	}
	for _, suite := range hello.CipherSuites {
		retval = certificate.RSA
		switch suite {
		case tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:
			retval = certificate.EC
			return retval
		}
	}
	return
}

// GetBestCertificate returns the best match certificate, and caches the response
func (c CertificateStore) GetBestCertificate(clientHello *tls.ClientHelloInfo) *tls.Certificate {
	domainToCheck := strings.ToLower(strings.TrimSpace(clientHello.ServerName))
	if len(domainToCheck) == 0 {
		// If no ServerName is provided, Check for local IP address matches
		host, _, err := net.SplitHostPort(clientHello.Conn.LocalAddr().String())
		if err != nil {
			log.Debugf("Could not split host/port: %v", err)
		}
		domainToCheck = strings.TrimSpace(host)
	}

	// Append compatibility with EC to key before checking cache
	preferredCertType := getCertTypeForClientHello(clientHello)
	keyToCheck := domainToCheck
	if preferredCertType != certificate.RSA {
		keyToCheck += certTypeDelimiter + preferredCertType.String()
	}
	if cert, ok := c.CertCache.Get(keyToCheck); ok {
		return cert.(*tls.Certificate)
	}

	// Build list of certificate types allowed for this client in order of preference
	// (EC would come before RSA for clients compatible with it)
	certTypePreferences := []certificate.CertificateType{certificate.RSA}
	matchedCerts := map[certificate.CertificateType]map[string]*tls.Certificate{certificate.RSA: {}}
	if preferredCertType != certificate.RSA {
		matchedCerts[preferredCertType] = map[string]*tls.Certificate{}
		certTypePreferences = append([]certificate.CertificateType{preferredCertType}, certTypePreferences...)
	}

	if c.DynamicCerts != nil && c.DynamicCerts.Get() != nil {
		for key, cert := range c.DynamicCerts.Get().(map[certificateKey]*tls.Certificate) {
			domains := key.hostname

			// Compatible certificate?
			if _, ok := matchedCerts[key.certType]; !ok {
				continue
			}

			// Requested domain found in certificate?
			for _, certDomain := range strings.Split(domains, ",") {
				if MatchDomain(domainToCheck, certDomain) {
					matchedCerts[key.certType][certDomain] = cert
				}
			}
		}
	}

	for _, currentCertType := range certTypePreferences {
		matchedCertsForCurrentCertType := matchedCerts[currentCertType]
		if len(matchedCertsForCurrentCertType) > 0 {
			// sort map by keys
			keys := []string{}
			for key := range matchedCertsForCurrentCertType {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			// cache best match
			c.CertCache.SetDefault(keyToCheck, matchedCertsForCurrentCertType[keys[len(keys)-1]])
			return matchedCertsForCurrentCertType[keys[len(keys)-1]]
		}
	}

	return nil
}

// ResetCache clears the cache in the store
func (c CertificateStore) ResetCache() {
	if c.CertCache != nil {
		c.CertCache.Flush()
	}
}

// MatchDomain return true if a domain match the cert domain
func MatchDomain(domain string, certDomain string) bool {
	if domain == certDomain {
		return true
	}

	for len(certDomain) > 0 && certDomain[len(certDomain)-1] == '.' {
		certDomain = certDomain[:len(certDomain)-1]
	}

	labels := strings.Split(domain, ".")
	for i := range labels {
		labels[i] = "*"
		candidate := strings.Join(labels, ".")
		if certDomain == candidate {
			return true
		}
	}
	return false
}
