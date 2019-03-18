package acme

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/containous/traefik/log"
	acmeprovider "github.com/containous/traefik/provider/acme"
	"github.com/containous/traefik/types"
	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/registration"
)

// Account is used to store lets encrypt registration info
type Account struct {
	Email              string
	Registration       *registration.Resource
	PrivateKey         []byte
	KeyType            certcrypto.KeyType
	DomainsCertificate DomainsCertificates
	ChallengeCerts     map[string]*ChallengeCert
	HTTPChallenge      map[string]map[string][]byte
}

// ChallengeCert stores a challenge certificate
type ChallengeCert struct {
	Certificate []byte
	PrivateKey  []byte
	certificate *tls.Certificate
}

// Init account struct
func (a *Account) Init() error {
	err := a.DomainsCertificate.Init()
	if err != nil {
		return err
	}

	err = a.RemoveAccountV1Values()
	if err != nil {
		log.Errorf("Unable to remove ACME Account V1 values during account initialization: %v", err)
	}

	for _, cert := range a.ChallengeCerts {
		if cert.certificate == nil {
			certificate, err := tls.X509KeyPair(cert.Certificate, cert.PrivateKey)
			if err != nil {
				return err
			}
			cert.certificate = &certificate
		}

		if cert.certificate.Leaf == nil {
			leaf, err := x509.ParseCertificate(cert.certificate.Certificate[0])
			if err != nil {
				return err
			}
			cert.certificate.Leaf = leaf
		}
	}
	return nil
}

// NewAccount creates an account
func NewAccount(email string, certs []*DomainsCertificate, keyTypeValue string) (*Account, error) {
	keyType := acmeprovider.GetKeyType(keyTypeValue)

	// Create a user. New accounts need an email and private key to start
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	domainsCerts := DomainsCertificates{Certs: certs}
	err = domainsCerts.Init()
	if err != nil {
		return nil, err
	}

	return &Account{
		Email:              email,
		PrivateKey:         x509.MarshalPKCS1PrivateKey(privateKey),
		KeyType:            keyType,
		DomainsCertificate: DomainsCertificates{Certs: domainsCerts.Certs},
		ChallengeCerts:     map[string]*ChallengeCert{}}, nil
}

// GetEmail returns email
func (a *Account) GetEmail() string {
	return a.Email
}

// GetRegistration returns lets encrypt registration resource
func (a *Account) GetRegistration() *registration.Resource {
	return a.Registration
}

// GetPrivateKey returns private key
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	if privateKey, err := x509.ParsePKCS1PrivateKey(a.PrivateKey); err == nil {
		return privateKey
	}

	log.Errorf("Cannot unmarshall private key %+v", a.PrivateKey)
	return nil
}

// RemoveAccountV1Values removes ACME account V1 values
func (a *Account) RemoveAccountV1Values() error {
	// Check if ACME Account is in ACME V1 format
	if a.Registration != nil {
		isOldRegistration, err := regexp.MatchString(acmeprovider.RegistrationURLPathV1Regexp, a.Registration.URI)
		if err != nil {
			return err
		}

		if isOldRegistration {
			a.reset()
		}
	}
	return nil
}

func (a *Account) reset() {
	log.Debug("Reset ACME account object.")
	a.Email = ""
	a.Registration = nil
	a.PrivateKey = nil
}

// Certificate is used to store certificate info
type Certificate struct {
	Domain        string
	CertURL       string
	CertStableURL string
	PrivateKey    []byte
	Certificate   []byte
}

// DomainsCertificates stores a certificate for multiple domains
type DomainsCertificates struct {
	Certs []*DomainsCertificate
	lock  sync.RWMutex
}

func (dc *DomainsCertificates) Len() int {
	return len(dc.Certs)
}

func (dc *DomainsCertificates) Swap(i, j int) {
	dc.Certs[i], dc.Certs[j] = dc.Certs[j], dc.Certs[i]
}

func (dc *DomainsCertificates) Less(i, j int) bool {
	if reflect.DeepEqual(dc.Certs[i].Domains, dc.Certs[j].Domains) {
		return dc.Certs[i].tlsCert.Leaf.NotAfter.After(dc.Certs[j].tlsCert.Leaf.NotAfter)
	}

	if dc.Certs[i].Domains.Main == dc.Certs[j].Domains.Main {
		return strings.Join(dc.Certs[i].Domains.SANs, ",") < strings.Join(dc.Certs[j].Domains.SANs, ",")
	}

	return dc.Certs[i].Domains.Main < dc.Certs[j].Domains.Main
}

func (dc *DomainsCertificates) removeDuplicates() {
	sort.Sort(dc)
	for i := 0; i < len(dc.Certs); i++ {
		for i2 := i + 1; i2 < len(dc.Certs); i2++ {
			if reflect.DeepEqual(dc.Certs[i].Domains, dc.Certs[i2].Domains) {
				// delete
				log.Warnf("Remove duplicate cert: %+v, expiration :%s", dc.Certs[i2].Domains, dc.Certs[i2].tlsCert.Leaf.NotAfter.String())
				dc.Certs = append(dc.Certs[:i2], dc.Certs[i2+1:]...)
				i2--
			}
		}
	}
}

func (dc *DomainsCertificates) removeEmpty() {
	var certs []*DomainsCertificate
	for _, cert := range dc.Certs {
		if cert.Certificate != nil && len(cert.Certificate.Certificate) > 0 && len(cert.Certificate.PrivateKey) > 0 {
			certs = append(certs, cert)
		}
	}
	dc.Certs = certs
}

// Init DomainsCertificates
func (dc *DomainsCertificates) Init() error {
	dc.lock.Lock()
	defer dc.lock.Unlock()

	dc.removeEmpty()

	for _, domainsCertificate := range dc.Certs {
		tlsCert, err := tls.X509KeyPair(domainsCertificate.Certificate.Certificate, domainsCertificate.Certificate.PrivateKey)
		if err != nil {
			return err
		}

		domainsCertificate.tlsCert = &tlsCert

		if domainsCertificate.tlsCert.Leaf == nil {
			leaf, err := x509.ParseCertificate(domainsCertificate.tlsCert.Certificate[0])
			if err != nil {
				return err
			}

			domainsCertificate.tlsCert.Leaf = leaf
		}
	}

	dc.removeDuplicates()
	return nil
}

func (dc *DomainsCertificates) renewCertificates(acmeCert *Certificate, domain types.Domain) error {
	dc.lock.Lock()
	defer dc.lock.Unlock()

	for _, domainsCertificate := range dc.Certs {
		if reflect.DeepEqual(domain, domainsCertificate.Domains) {
			tlsCert, err := tls.X509KeyPair(acmeCert.Certificate, acmeCert.PrivateKey)
			if err != nil {
				return err
			}

			domainsCertificate.Certificate = acmeCert
			domainsCertificate.tlsCert = &tlsCert
			return nil
		}
	}

	return fmt.Errorf("certificate to renew not found for domain %s", domain.Main)
}

func (dc *DomainsCertificates) addCertificateForDomains(acmeCert *Certificate, domain types.Domain) (*DomainsCertificate, error) {
	dc.lock.Lock()
	defer dc.lock.Unlock()

	tlsCert, err := tls.X509KeyPair(acmeCert.Certificate, acmeCert.PrivateKey)
	if err != nil {
		return nil, err
	}

	cert := DomainsCertificate{Domains: domain, Certificate: acmeCert, tlsCert: &tlsCert}
	dc.Certs = append(dc.Certs, &cert)
	return &cert, nil
}

func (dc *DomainsCertificates) getCertificateForDomain(domainToFind string) (*DomainsCertificate, bool) {
	dc.lock.RLock()
	defer dc.lock.RUnlock()

	for _, domainsCertificate := range dc.Certs {
		for _, domain := range domainsCertificate.Domains.ToStrArray() {
			if strings.HasPrefix(domain, "*.") && types.MatchDomain(domainToFind, domain) {
				return domainsCertificate, true
			}
			if domain == domainToFind {
				return domainsCertificate, true
			}
		}
	}
	return nil, false
}

func (dc *DomainsCertificates) exists(domainToFind types.Domain) (*DomainsCertificate, bool) {
	dc.lock.RLock()
	defer dc.lock.RUnlock()

	for _, domainsCertificate := range dc.Certs {
		if reflect.DeepEqual(domainToFind, domainsCertificate.Domains) {
			return domainsCertificate, true
		}
	}
	return nil, false
}

func (dc *DomainsCertificates) toDomainsMap() map[string]*tls.Certificate {
	domainsCertificatesMap := make(map[string]*tls.Certificate)

	for _, domainCertificate := range dc.Certs {
		certKey := domainCertificate.Domains.Main

		if domainCertificate.Domains.SANs != nil {
			sort.Strings(domainCertificate.Domains.SANs)

			for _, dnsName := range domainCertificate.Domains.SANs {
				if dnsName != domainCertificate.Domains.Main {
					certKey += fmt.Sprintf(",%s", dnsName)
				}
			}
		}
		domainsCertificatesMap[certKey] = domainCertificate.tlsCert
	}
	return domainsCertificatesMap
}

// DomainsCertificate contains a certificate for multiple domains
type DomainsCertificate struct {
	Domains     types.Domain
	Certificate *Certificate
	tlsCert     *tls.Certificate
}

func (dc *DomainsCertificate) needRenew() bool {
	for _, c := range dc.tlsCert.Certificate {
		crt, err := x509.ParseCertificate(c)
		if err != nil {
			// If there's an error, we assume the cert is broken, and needs update
			return true
		}

		// <= 30 days left, renew certificate
		if crt.NotAfter.Before(time.Now().Add(24 * 30 * time.Hour)) {
			return true
		}
	}

	return false
}
