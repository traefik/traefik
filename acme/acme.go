package acme

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/staert"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/xenolf/lego/acme"
	"golang.org/x/net/context"
	"io/ioutil"
	fmtlog "log"
	"os"
	"strings"
	"time"
)

// ACME allows to connect to lets encrypt and retrieve certs
type ACME struct {
	Email               string   `description:"Email address used for registration"`
	Domains             []Domain `description:"SANs (alternative domains) to each main domain using format: --acme.domains='main.com,san1.com,san2.com' --acme.domains='main.net,san1.net,san2.net'"`
	Storage             string   `description:"File or key used for certificates storage."`
	StorageFile         string   // deprecated
	OnDemand            bool     `description:"Enable on demand certificate. This will request a certificate from Let's Encrypt during the first TLS handshake for a hostname that does not yet have a certificate."`
	OnHostRule          bool     `description:"Enable certificate generation on frontends Host rules."`
	CAServer            string   `description:"CA server to use."`
	EntryPoint          string   `description:"Entrypoint to proxy acme challenge to."`
	client              *acme.Client
	defaultCertificate  *tls.Certificate
	store               cluster.Store
	challengeProvider   *challengeProvider
	checkOnDemandDomain func(domain string) bool
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
	acme.Logger = fmtlog.New(ioutil.Discard, "", 0)
	// no certificates in TLS config, so we add a default one
	cert, err := generateDefaultCertificate()
	if err != nil {
		return err
	}
	a.defaultCertificate = cert
	// TODO: to remove in the futurs
	if len(a.StorageFile) > 0 && len(a.Storage) == 0 {
		log.Warnf("ACME.StorageFile is deprecated, use ACME.Storage instead")
		a.Storage = a.StorageFile
	}
	return nil
}

// CreateClusterConfig creates a tls.config using ACME configuration in cluster mode
func (a *ACME) CreateClusterConfig(leadership *cluster.Leadership, tlsConfig *tls.Config, checkOnDemandDomain func(domain string) bool) error {
	err := a.init()
	if err != nil {
		return err
	}
	if len(a.Storage) == 0 {
		return errors.New("Empty Store, please provide a key for certs storage")
	}
	a.checkOnDemandDomain = checkOnDemandDomain
	tlsConfig.Certificates = append(tlsConfig.Certificates, *a.defaultCertificate)
	tlsConfig.GetCertificate = a.getCertificate
	listener := func(object cluster.Object) error {
		account := object.(*Account)
		account.Init()
		if !leadership.IsLeader() {
			a.client, err = a.buildACMEClient(account)
			if err != nil {
				log.Errorf("Error building ACME client %+v: %s", object, err.Error())
			}
		}
		return nil
	}

	datastore, err := cluster.NewDataStore(
		staert.KvSource{
			Store:  leadership.Store,
			Prefix: a.Storage,
		},
		leadership.Pool.Ctx(), &Account{},
		listener)
	if err != nil {
		return err
	}

	a.store = datastore
	a.challengeProvider = &challengeProvider{store: a.store}

	ticker := time.NewTicker(24 * time.Hour)
	leadership.Pool.AddGoCtx(func(ctx context.Context) {
		log.Infof("Starting ACME renew job...")
		defer log.Infof("Stopped ACME renew job...")
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := a.renewCertificates(); err != nil {
					log.Errorf("Error renewing ACME certificate: %s", err.Error())
				}
			}
		}
	})

	leadership.AddListener(func(elected bool) error {
		if elected {
			object, err := a.store.Load()
			if err != nil {
				return err
			}
			transaction, object, err := a.store.Begin()
			if err != nil {
				return err
			}
			account := object.(*Account)
			account.Init()
			var needRegister bool
			if account == nil || len(account.Email) == 0 {
				account, err = NewAccount(a.Email)
				if err != nil {
					return err
				}
				needRegister = true
			}
			if err != nil {
				return err
			}
			a.client, err = a.buildACMEClient(account)
			if err != nil {
				return err
			}
			if needRegister {
				// New users will need to register; be sure to save it
				log.Debugf("Register...")
				reg, err := a.client.Register()
				if err != nil {
					return err
				}
				account.Registration = reg
			}
			// The client has a URL to the current Let's Encrypt Subscriber
			// Agreement. The user will need to agree to it.
			log.Debugf("AgreeToTOS...")
			err = a.client.AgreeToTOS()
			if err != nil {
				// Let's Encrypt Subscriber Agreement renew ?
				reg, err := a.client.QueryRegistration()
				if err != nil {
					return err
				}
				account.Registration = reg
				err = a.client.AgreeToTOS()
				if err != nil {
					log.Errorf("Error sending ACME agreement to TOS: %+v: %s", account, err.Error())
				}
			}
			err = transaction.Commit(account)
			if err != nil {
				return err
			}
			safe.Go(func() {
				a.retrieveCertificates()
				if err := a.renewCertificates(); err != nil {
					log.Errorf("Error renewing ACME certificate %+v: %s", account, err.Error())
				}
			})
		}
		return nil
	})
	return nil
}

// CreateLocalConfig creates a tls.config using local ACME configuration
func (a *ACME) CreateLocalConfig(tlsConfig *tls.Config, checkOnDemandDomain func(domain string) bool) error {
	err := a.init()
	if err != nil {
		return err
	}
	if len(a.Storage) == 0 {
		return errors.New("Empty Store, please provide a filename for certs storage")
	}
	a.checkOnDemandDomain = checkOnDemandDomain
	tlsConfig.Certificates = append(tlsConfig.Certificates, *a.defaultCertificate)
	tlsConfig.GetCertificate = a.getCertificate

	localStore := NewLocalStore(a.Storage)
	a.store = localStore
	a.challengeProvider = &challengeProvider{store: a.store}

	var needRegister bool
	var account *Account

	if fileInfo, fileErr := os.Stat(a.Storage); fileErr == nil && fileInfo.Size() != 0 {
		log.Infof("Loading ACME Account...")
		// load account
		object, err := localStore.Load()
		if err != nil {
			return err
		}
		account = object.(*Account)
	} else {
		log.Infof("Generating ACME Account...")
		account, err = NewAccount(a.Email)
		if err != nil {
			return err
		}
		needRegister = true
	}

	log.Infof("buildACMEClient...")
	a.client, err = a.buildACMEClient(account)
	if err != nil {
		return err
	}

	if needRegister {
		// New users will need to register; be sure to save it
		log.Infof("Register...")
		reg, err := a.client.Register()
		if err != nil {
			return err
		}
		account.Registration = reg
	}

	// The client has a URL to the current Let's Encrypt Subscriber
	// Agreement. The user will need to agree to it.
	log.Infof("AgreeToTOS...")
	err = a.client.AgreeToTOS()
	if err != nil {
		// Let's Encrypt Subscriber Agreement renew ?
		reg, err := a.client.QueryRegistration()
		if err != nil {
			return err
		}
		account.Registration = reg
		err = a.client.AgreeToTOS()
		if err != nil {
			log.Errorf("Error sending ACME agreement to TOS: %+v: %s", account, err.Error())
		}
	}
	// save account
	transaction, _, err := a.store.Begin()
	if err != nil {
		return err
	}
	err = transaction.Commit(account)
	if err != nil {
		return err
	}

	safe.Go(func() {
		a.retrieveCertificates()
		if err := a.renewCertificates(); err != nil {
			log.Errorf("Error renewing ACME certificate %+v: %s", account, err.Error())
		}
	})

	ticker := time.NewTicker(24 * time.Hour)
	safe.Go(func() {
		for range ticker.C {
			if err := a.renewCertificates(); err != nil {
				log.Errorf("Error renewing ACME certificate %+v: %s", account, err.Error())
			}
		}

	})
	return nil
}

func (a *ACME) getCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	domain := types.CanonicalDomain(clientHello.ServerName)
	account := a.store.Get().(*Account)
	if challengeCert, ok := a.challengeProvider.getCertificate(domain); ok {
		log.Debugf("ACME got challenge %s", domain)
		return challengeCert, nil
	}
	if domainCert, ok := account.DomainsCertificate.getCertificateForDomain(domain); ok {
		log.Debugf("ACME got domain cert %s", domain)
		return domainCert.tlsCert, nil
	}
	if a.OnDemand {
		if a.checkOnDemandDomain != nil && !a.checkOnDemandDomain(domain) {
			return nil, nil
		}
		return a.loadCertificateOnDemand(clientHello)
	}
	log.Debugf("ACME got nothing %s", domain)
	return nil, nil
}

func (a *ACME) retrieveCertificates() {
	log.Infof("Retrieving ACME certificates...")
	for _, domain := range a.Domains {
		// check if cert isn't already loaded
		account := a.store.Get().(*Account)
		if _, exists := account.DomainsCertificate.exists(domain); !exists {
			domains := []string{}
			domains = append(domains, domain.Main)
			domains = append(domains, domain.SANs...)
			certificateResource, err := a.getDomainsCertificates(domains)
			if err != nil {
				log.Errorf("Error getting ACME certificate for domain %s: %s", domains, err.Error())
				continue
			}
			transaction, object, err := a.store.Begin()
			if err != nil {
				log.Errorf("Error creating ACME store transaction from domain %s: %s", domain, err.Error())
				continue
			}
			account = object.(*Account)
			_, err = account.DomainsCertificate.addCertificateForDomains(certificateResource, domain)
			if err != nil {
				log.Errorf("Error adding ACME certificate for domain %s: %s", domains, err.Error())
				continue
			}

			if err = transaction.Commit(account); err != nil {
				log.Errorf("Error Saving ACME account %+v: %s", account, err.Error())
				continue
			}
		}
	}
	log.Infof("Retrieved ACME certificates")
}

func (a *ACME) renewCertificates() error {
	log.Debugf("Testing certificate renew...")
	account := a.store.Get().(*Account)
	for _, certificateResource := range account.DomainsCertificate.Certs {
		if certificateResource.needRenew() {
			transaction, object, err := a.store.Begin()
			if err != nil {
				return err
			}
			account = object.(*Account)
			log.Debugf("Renewing certificate %+v", certificateResource.Domains)
			renewedCert, err := a.client.RenewCertificate(acme.CertificateResource{
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
			err = account.DomainsCertificate.renewCertificates(renewedACMECert, certificateResource.Domains)
			if err != nil {
				log.Errorf("Error renewing certificate: %v", err)
				continue
			}

			if err = transaction.Commit(account); err != nil {
				log.Errorf("Error Saving ACME account %+v: %s", account, err.Error())
				continue
			}
		}
	}
	return nil
}

func (a *ACME) buildACMEClient(account *Account) (*acme.Client, error) {
	log.Debugf("Building ACME client...")
	caServer := "https://acme-v01.api.letsencrypt.org/directory"
	if len(a.CAServer) > 0 {
		caServer = a.CAServer
	}
	client, err := acme.NewClient(caServer, account, acme.RSA4096)
	if err != nil {
		return nil, err
	}
	client.ExcludeChallenges([]acme.Challenge{acme.HTTP01, acme.DNS01})
	err = client.SetChallengeProvider(acme.TLSSNI01, a.challengeProvider)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (a *ACME) loadCertificateOnDemand(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	domain := types.CanonicalDomain(clientHello.ServerName)
	account := a.store.Get().(*Account)
	if certificateResource, ok := account.DomainsCertificate.getCertificateForDomain(domain); ok {
		return certificateResource.tlsCert, nil
	}
	certificate, err := a.getDomainsCertificates([]string{domain})
	if err != nil {
		return nil, err
	}
	log.Debugf("Got certificate on demand for domain %s", domain)

	transaction, object, err := a.store.Begin()
	if err != nil {
		return nil, err
	}
	account = object.(*Account)
	cert, err := account.DomainsCertificate.addCertificateForDomains(certificate, Domain{Main: domain})
	if err != nil {
		return nil, err
	}
	if err = transaction.Commit(account); err != nil {
		return nil, err
	}
	return cert.tlsCert, nil
}

// LoadCertificateForDomains loads certificates from ACME for given domains
func (a *ACME) LoadCertificateForDomains(domains []string) {
	domains = fun.Map(types.CanonicalDomain, domains).([]string)
	safe.Go(func() {
		operation := func() error {
			if a.client == nil {
				return fmt.Errorf("ACME client still not built")
			}
			return nil
		}
		notify := func(err error, time time.Duration) {
			log.Errorf("Error getting ACME client: %v, retrying in %s", err, time)
		}
		ebo := backoff.NewExponentialBackOff()
		ebo.MaxElapsedTime = 30 * time.Second
		err := backoff.RetryNotify(operation, ebo, notify)
		if err != nil {
			log.Errorf("Error getting ACME client: %v", err)
			return
		}
		account := a.store.Get().(*Account)
		var domain Domain
		if len(domains) == 0 {
			// no domain
			return

		} else if len(domains) > 1 {
			domain = Domain{Main: domains[0], SANs: domains[1:]}
		} else {
			domain = Domain{Main: domains[0]}
		}
		if _, exists := account.DomainsCertificate.exists(domain); exists {
			// domain already exists
			return
		}
		certificate, err := a.getDomainsCertificates(domains)
		if err != nil {
			log.Errorf("Error getting ACME certificates %+v : %v", domains, err)
			return
		}
		log.Debugf("Got certificate for domains %+v", domains)
		transaction, object, err := a.store.Begin()

		if err != nil {
			log.Errorf("Error creating transaction %+v : %v", domains, err)
			return
		}
		account = object.(*Account)
		_, err = account.DomainsCertificate.addCertificateForDomains(certificate, domain)
		if err != nil {
			log.Errorf("Error adding ACME certificates %+v : %v", domains, err)
			return
		}
		if err = transaction.Commit(account); err != nil {
			log.Errorf("Error Saving ACME account %+v: %v", account, err)
			return
		}
	})
}

func (a *ACME) getDomainsCertificates(domains []string) (*Certificate, error) {
	domains = fun.Map(types.CanonicalDomain, domains).([]string)
	log.Debugf("Loading ACME certificates %s...", domains)
	bundle := true
	certificate, failures := a.client.ObtainCertificate(domains, bundle, nil)
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
