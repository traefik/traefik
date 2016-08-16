package acme

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	fmtlog "log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/containous/traefik/safe"
	"github.com/xenolf/lego/acme"
)

// Account is used to store lets encrypt registration info
type Account struct {
	Email              string
	Registration       *acme.RegistrationResource
	PrivateKey         []byte
	DomainsCertificate DomainsCertificates
}

// GetEmail returns email
func (a Account) GetEmail() string {
	return a.Email
}

// GetRegistration returns lets encrypt registration resource
func (a Account) GetRegistration() *acme.RegistrationResource {
	return a.Registration
}

// GetPrivateKey returns private key
func (a Account) GetPrivateKey() crypto.PrivateKey {
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
	lock  *sync.RWMutex
}

func (dc *DomainsCertificates) init() error {
	if dc.lock == nil {
		dc.lock = &sync.RWMutex{}
	}
	dc.lock.Lock()
	defer dc.lock.Unlock()
	for _, domainsCertificate := range dc.Certs {
		tlsCert, err := tls.X509KeyPair(domainsCertificate.Certificate.Certificate, domainsCertificate.Certificate.PrivateKey)
		if err != nil {
			return err
		}
		domainsCertificate.tlsCert = &tlsCert
	}
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
	return errors.New("Certificate to renew not found for domain " + domain.Main)
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

// ACME allows to connect to lets encrypt and retrieve certs
type ACME struct {
	Email       string   `description:"Email address used for registration"`
	Domains     []Domain `description:"SANs (alternative domains) to each main domain using format: --acme.domains='main.com,san1.com,san2.com' --acme.domains='main.net,san1.net,san2.net'"`
	StorageFile string   `description:"File used for certificates storage."`
	OnDemand    bool     `description:"Enable on demand certificate. This will request a certificate from Let's Encrypt during the first TLS handshake for a hostname that does not yet have a certificate."`
	OnHostRule  bool     `description:"Enable certificate generation on frontends Host rules."`
	CAServer    string   `description:"CA server to use."`
	EntryPoint  string   `description:"Entrypoint to proxy acme challenge to."`
	storageLock sync.RWMutex
	client      *acme.Client
	account     *Account
	defaultCertificate *tls.Certificate
}

//Domains parse []Domain
type Domains []Domain

//Set []Domain
func (ds *Domains) Set(str string) error {
	fargs := func(c rune) bool {
		return c == ',' || c == ';'
	}
	// get function
	slice := strings.FieldsFunc(str, fargs)
	if len(slice) < 1 {
		return fmt.Errorf("Parse error ACME.Domain. Imposible to parse %s", str)
	}
	d := Domain{
		Main: slice[0],
		SANs: []string{},
	}
	if len(slice) > 1 {
		d.SANs = slice[1:]
	}
	*ds = append(*ds, d)
	return nil
}

//Get []Domain
func (ds *Domains) Get() interface{} { return []Domain(*ds) }

//String returns []Domain in string
func (ds *Domains) String() string { return fmt.Sprintf("%+v", *ds) }

//SetValue sets []Domain into the parser
func (ds *Domains) SetValue(val interface{}) {
	*ds = Domains(val.([]Domain))
}

// Domain holds a domain name with SANs
type Domain struct {
	Main string
	SANs []string
}

func (a *ACME) init() error {
	if len(a.Store) == 0 {
		a.Store = a.StorageFile
	}
	acme.Logger = fmtlog.New(ioutil.Discard, "", 0)
	log.Debugf("Generating default certificate...")
	// no certificates in TLS config, so we add a default one
	cert, err := generateDefaultCertificate()
	if err != nil {
		return err
	}
	a.defaultCertificate = cert
	return nil
}

// CreateClusterConfig creates a tls.config from using ACME configuration
func (a *ACME) CreateClusterConfig(tlsConfig *tls.Config, CheckOnDemandDomain func(domain string) bool) error {
	err := a.init()
	if err != nil {
		return err
	}
	if len(a.Store) == 0 {
		return errors.New("Empty Store, please provide a filename/key for certs storage")
	}
	tlsConfig.Certificates = append(tlsConfig.Certificates, *a.defaultCertificate)
	return nil
}

// CreateLocalConfig creates a tls.config from using ACME configuration
func (a *ACME) CreateLocalConfig(tlsConfig *tls.Config, CheckOnDemandDomain func(domain string) bool) error {
	err := a.init()
	if err != nil {
		return err
	}
	if len(a.Store) == 0 {
		return errors.New("Empty Store, please provide a filename/key for certs storage")
	}
	tlsConfig.Certificates = append(tlsConfig.Certificates, *a.defaultCertificate)

	var needRegister bool
	var err error

	// if certificates in storage, load them
	if fileInfo, fileErr := os.Stat(a.Store); fileErr == nil && fileInfo.Size() != 0 {
		log.Infof("Loading ACME certificates...")
		// load account
		a.account, err = a.loadAccount(a)
		if err != nil {
			return err
		}
	} else {
		log.Infof("Generating ACME Account...")
		// Create a user. New accounts need an email and private key to start
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return err
		}
		a.account = &Account{
			Email:      a.Email,
			PrivateKey: x509.MarshalPKCS1PrivateKey(privateKey),
		}
		a.account.DomainsCertificate = DomainsCertificates{Certs: []*DomainsCertificate{}, lock: &sync.RWMutex{}}
		needRegister = true
	}

	a.client, err = a.buildACMEClient()
	if err != nil {
		return err
	}
	a.client.ExcludeChallenges([]acme.Challenge{acme.HTTP01, acme.DNS01})
	wrapperChallengeProvider := newWrapperChallengeProvider()
	err = client.SetChallengeProvider(acme.TLSSNI01, wrapperChallengeProvider)
	if err != nil {
		return err
	}

	if needRegister {
		// New users will need to register; be sure to save it
		reg, err := a.client.Register()
		if err != nil {
			return err
		}
		a.account.Registration = reg
	}

	// The client has a URL to the current Let's Encrypt Subscriber
	// Agreement. The user will need to agree to it.
	err = a.client.AgreeToTOS()
	if err != nil {
		// Let's Encrypt Subscriber Agreement renew ?
		reg, err := a.client.QueryRegistration()
		if err != nil {
			return err
		}
		a.account.Registration = reg
		err = a.client.AgreeToTOS()
		if err != nil {
			log.Errorf("Error sending ACME agreement to TOS: %+v: %s", a.account, err.Error())
		}
	}
	// save account
	err = a.saveAccount()
	if err != nil {
		return err
	}

	safe.Go(func() {
		a.retrieveCertificates(a.client)
		if err := a.renewCertificates(a.client); err != nil {
			log.Errorf("Error renewing ACME certificate %+v: %s", a.account, err.Error())
		}
	})

	tlsConfig.GetCertificate = func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		if challengeCert, ok := wrapperChallengeProvider.getCertificate(clientHello.ServerName); ok {
			return challengeCert, nil
		}
		if domainCert, ok := a.account.DomainsCertificate.getCertificateForDomain(clientHello.ServerName); ok {
			return domainCert.tlsCert, nil
		}
		if a.OnDemand {
			if CheckOnDemandDomain != nil && !CheckOnDemandDomain(clientHello.ServerName) {
				return nil, nil
			}
			return a.loadCertificateOnDemand(clientHello)
		}
		return nil, nil
	}

	ticker := time.NewTicker(24 * time.Hour)
	safe.Go(func() {
		for range ticker.C {
			if err := a.renewCertificates(client, account); err != nil {
				log.Errorf("Error renewing ACME certificate %+v: %s", account, err.Error())
			}
		}

	})
	return nil
}

func (a *ACME) retrieveCertificates(client *acme.Client) {
	log.Infof("Retrieving ACME certificates...")
	for _, domain := range a.Domains {
		// check if cert isn't already loaded
		if _, exists := a.account.DomainsCertificate.exists(domain); !exists {
			domains := []string{}
			domains = append(domains, domain.Main)
			domains = append(domains, domain.SANs...)
			certificateResource, err := a.getDomainsCertificates(client, domains)
			if err != nil {
				log.Errorf("Error getting ACME certificate for domain %s: %s", domains, err.Error())
				continue
			}
			_, err = a.account.DomainsCertificate.addCertificateForDomains(certificateResource, domain)
			if err != nil {
				log.Errorf("Error adding ACME certificate for domain %s: %s", domains, err.Error())
				continue
			}
			if err = a.saveAccount(); err != nil {
				log.Errorf("Error Saving ACME account %+v: %s", a.account, err.Error())
				continue
			}
		}
	}
	log.Infof("Retrieved ACME certificates")
}

func (a *ACME) renewCertificates(client *acme.Client) error {
	log.Debugf("Testing certificate renew...")
	for _, certificateResource := range a.account.DomainsCertificate.Certs {
		if certificateResource.needRenew() {
			log.Debugf("Renewing certificate %+v", certificateResource.Domains)
			renewedCert, err := client.RenewCertificate(acme.CertificateResource{
				Domain:        certificateResource.Certificate.Domain,
				CertURL:       certificateResource.Certificate.CertURL,
				CertStableURL: certificateResource.Certificate.CertStableURL,
				PrivateKey:    certificateResource.Certificate.PrivateKey,
				Certificate:   certificateResource.Certificate.Certificate,
			}, true)
			if err != nil {
				log.Errorf("Error renewing certificate: %v", err)
				continue
			}
			log.Debugf("Renewed certificate %+v", certificateResource.Domains)
			renewedACMECert := &Certificate{
				Domain:        renewedCert.Domain,
				CertURL:       renewedCert.CertURL,
				CertStableURL: renewedCert.CertStableURL,
				PrivateKey:    renewedCert.PrivateKey,
				Certificate:   renewedCert.Certificate,
			}
			err = a.account.DomainsCertificate.renewCertificates(renewedACMECert, certificateResource.Domains)
			if err != nil {
				log.Errorf("Error renewing certificate: %v", err)
				continue
			}
			if err = a.saveAccount(); err != nil {
				log.Errorf("Error saving ACME account: %v", err)
				continue
			}
		}
	}
	return nil
}

func (a *ACME) buildACMEClient() (*acme.Client, error) {
	caServer := "https://acme-v01.api.letsencrypt.org/directory"
	if len(a.CAServer) > 0 {
		caServer = a.CAServer
	}
	client, err := acme.NewClient(caServer, a.account, acme.RSA4096)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (a *ACME) loadCertificateOnDemand(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if certificateResource, ok := a.account.DomainsCertificate.getCertificateForDomain(clientHello.ServerName); ok {
		return certificateResource.tlsCert, nil
	}
	certificate, err := a.getDomainsCertificates(a.client, []string{clientHello.ServerName})
	if err != nil {
		return nil, err
	}
	log.Debugf("Got certificate on demand for domain %s", clientHello.ServerName)
	cert, err := a.account.DomainsCertificate.addCertificateForDomains(certificate, Domain{Main: clientHello.ServerName})
	if err != nil {
		return nil, err
	}
	if err = a.saveAccount(); err != nil {
		return nil, err
	}
	return cert.tlsCert, nil
}

// LoadCertificateForDomains loads certificates from ACME for given domains
func (a *ACME) LoadCertificateForDomains(domains []string) {
	safe.Go(func() {
		var domain Domain
		if len(domains) == 0 {
			// no domain
			return

		} else if len(domains) > 1 {
			domain = Domain{Main: domains[0], SANs: domains[1:]}
		} else {
			domain = Domain{Main: domains[0]}
		}
		if _, exists := a.account.DomainsCertificate.exists(domain); exists {
			// domain already exists
			return
		}
		certificate, err := a.getDomainsCertificates(a.client, domains)
		if err != nil {
			log.Errorf("Error getting ACME certificates %+v : %v", domains, err)
			return
		}
		log.Debugf("Got certificate for domains %+v", domains)
		_, err = a.account.DomainsCertificate.addCertificateForDomains(certificate, domain)
		if err != nil {
			log.Errorf("Error adding ACME certificates %+v : %v", domains, err)
			return
		}
		if err = a.saveAccount(); err != nil {
			log.Errorf("Error Saving ACME account %+v: %v", a.account, err)
			return
		}
	})
}

func (a *ACME) loadAccount(acmeConfig *ACME) (*Account, error) {
	a.storageLock.RLock()
	defer a.storageLock.RUnlock()
	account := Account{
		DomainsCertificate: DomainsCertificates{},
	}
	file, err := ioutil.ReadFile(acmeConfig.Store)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(file, &account); err != nil {
		return nil, err
	}
	err = account.DomainsCertificate.init()
	if err != nil {
		return nil, err
	}
	log.Infof("Loaded ACME config from store %s", acmeConfig.Store)
	return &account, nil
}

func (a *ACME) saveAccount() error {
	a.storageLock.Lock()
	defer a.storageLock.Unlock()
	// write account to file
	data, err := json.MarshalIndent(a.account, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(a.Store, data, 0600)
}

func (a *ACME) getDomainsCertificates(client *acme.Client, domains []string) (*Certificate, error) {
	log.Debugf("Loading ACME certificates %s...", domains)
	bundle := true
	certificate, failures := client.ObtainCertificate(domains, bundle, nil)
	if len(failures) > 0 {
		log.Error(failures)
		return nil, fmt.Errorf("Cannot obtain certificates %s+v", failures)
	}
	log.Debugf("Loaded ACME certificates %s", domains)
	return &Certificate{
		Domain:        certificate.Domain,
		CertURL:       certificate.CertURL,
		CertStableURL: certificate.CertStableURL,
		PrivateKey:    certificate.PrivateKey,
		Certificate:   certificate.Certificate,
	}, nil
}
