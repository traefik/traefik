package acme

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/containous/traefik/log"
	"github.com/xenolf/lego/acme"
)

// Account is used to store lets encrypt registration info
type Account struct {
	Email              string
	Registration       *acme.RegistrationResource
	PrivateKey         []byte
	DomainsCertificate DomainsCertificates
	ChallengeCerts     map[string]*ChallengeCert
}

// ChallengeCert stores a challenge certificate
type ChallengeCert struct {
	Certificate []byte
	PrivateKey  []byte
	certificate *tls.Certificate
}

// Init inits account struct
func (a *Account) Init() error {
	err := a.DomainsCertificate.Init()
	if err != nil {
		return err
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
func NewAccount(email string) (*Account, error) {
	// Create a user. New accounts need an email and private key to start
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}
	domainsCerts := DomainsCertificates{Certs: []*DomainsCertificate{}}
	domainsCerts.Init()
	return &Account{
		Email:              email,
		PrivateKey:         x509.MarshalPKCS1PrivateKey(privateKey),
		DomainsCertificate: DomainsCertificates{Certs: domainsCerts.Certs},
		ChallengeCerts:     map[string]*ChallengeCert{}}, nil
}

// GetEmail returns email
func (a *Account) GetEmail() string {
	return a.Email
}

// GetRegistration returns lets encrypt registration resource
func (a *Account) GetRegistration() *acme.RegistrationResource {
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

// Init inits DomainsCertificates
func (dc *DomainsCertificates) Init() error {
	dc.lock.Lock()
	defer dc.lock.Unlock()
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

func (dc *DomainsCertificates) renewCertificates(acmeCert *Certificate, domain Domain) error {
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
	return fmt.Errorf("Certificate to renew not found for domain %s", domain.Main)
}

func (dc *DomainsCertificates) addCertificateForDomains(acmeCert *Certificate, domain Domain) (*DomainsCertificate, error) {
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
		domains := []string{}
		domains = append(domains, domainsCertificate.Domains.Main)
		domains = append(domains, domainsCertificate.Domains.SANs...)
		for _, domain := range domains {
			if domain == domainToFind {
				return domainsCertificate, true
			}
		}
	}
	return nil, false
}

func (dc *DomainsCertificates) exists(domainToFind Domain) (*DomainsCertificate, bool) {
	dc.lock.RLock()
	defer dc.lock.RUnlock()
	for _, domainsCertificate := range dc.Certs {
		if reflect.DeepEqual(domainToFind, domainsCertificate.Domains) {
			return domainsCertificate, true
		}
	}
	return nil, false
}

// DomainsCertificate contains a certificate for multiple domains
type DomainsCertificate struct {
	Domains     Domain
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
		if crt.NotAfter.Before(time.Now().Add(time.Duration(24 * 30 * time.Hour))) {
			return true
		}
	}

	return false
}
