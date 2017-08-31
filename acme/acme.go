package acme

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	fmtlog "log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/staert"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/eapache/channels"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/providers/dns"
)

var (
	// OSCPMustStaple enables OSCP stapling as from https://github.com/xenolf/lego/issues/270
	OSCPMustStaple = false
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
	DNSProvider         string   `description:"Use a DNS based challenge provider rather than HTTPS."`
	DelayDontCheckDNS   int      `description:"Assume DNS propagates after a delay in seconds rather than finding and querying nameservers."`
	ACMELogging         bool     `description:"Enable debug logging of ACME actions."`
	client              *acme.Client
	defaultCertificate  *tls.Certificate
	store               cluster.Store
	challengeProvider   *challengeProvider
	checkOnDemandDomain func(domain string) bool
	jobs                *channels.InfiniteChannel
	TLSConfig           *tls.Config `description:"TLS config in case wildcard certs are used"`
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
	if a.ACMELogging {
		acme.Logger = fmtlog.New(os.Stderr, "legolog: ", fmtlog.LstdFlags)
	} else {
		acme.Logger = fmtlog.New(ioutil.Discard, "", 0)
	}
	// no certificates in TLS config, so we add a default one
	cert, err := generateDefaultCertificate()
	if err != nil {
		return err
	}
	a.defaultCertificate = cert
	// TODO: to remove in the futurs
	if len(a.StorageFile) > 0 && len(a.Storage) == 0 {
		log.Warn("ACME.StorageFile is deprecated, use ACME.Storage instead")
		a.Storage = a.StorageFile
	}
	a.jobs = channels.NewInfiniteChannel()
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
	a.TLSConfig = tlsConfig
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
		leadership.Pool.Ctx(),
		staert.KvSource{
			Store:  leadership.Store,
			Prefix: a.Storage,
		},
		&Account{},
		listener)
	if err != nil {
		return err
	}

	a.store = datastore
	a.challengeProvider = &challengeProvider{store: a.store}

	ticker := time.NewTicker(24 * time.Hour)
	leadership.Pool.AddGoCtx(func(ctx context.Context) {
		log.Info("Starting ACME renew job...")
		defer log.Info("Stopped ACME renew job...")
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.renewCertificates()
			}
		}
	})

	leadership.AddListener(func(elected bool) error {
		if elected {
			_, err := a.store.Load()
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
				log.Debug("Register...")
				reg, err := a.client.Register()
				if err != nil {
					return err
				}
				account.Registration = reg
			}
			// The client has a URL to the current Let's Encrypt Subscriber
			// Agreement. The user will need to agree to it.
			log.Debug("AgreeToTOS...")
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

			a.retrieveCertificates()
			a.renewCertificates()
			a.runJobs()
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
	a.TLSConfig = tlsConfig
	localStore := NewLocalStore(a.Storage)
	a.store = localStore
	a.challengeProvider = &challengeProvider{store: a.store}

	var needRegister bool
	var account *Account

	if fileInfo, fileErr := os.Stat(a.Storage); fileErr == nil && fileInfo.Size() != 0 {
		log.Info("Loading ACME Account...")
		// load account
		object, err := localStore.Load()
		if err != nil {
			return err
		}
		account = object.(*Account)
	} else {
		log.Info("Generating ACME Account...")
		account, err = NewAccount(a.Email)
		if err != nil {
			return err
		}
		needRegister = true
	}

	a.client, err = a.buildACMEClient(account)
	if err != nil {
		return err
	}

	if needRegister {
		// New users will need to register; be sure to save it
		log.Info("Register...")
		reg, err := a.client.Register()
		if err != nil {
			return err
		}
		account.Registration = reg
	}

	// The client has a URL to the current Let's Encrypt Subscriber
	// Agreement. The user will need to agree to it.
	log.Debug("AgreeToTOS...")
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

	a.retrieveCertificates()
	a.renewCertificates()
	a.runJobs()

	ticker := time.NewTicker(24 * time.Hour)
	safe.Go(func() {
		for range ticker.C {
			a.renewCertificates()
		}
	})
	return nil
}

func (a *ACME) getCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	domain := types.CanonicalDomain(clientHello.ServerName)
	account := a.store.Get().(*Account)

	if providedCertificate := a.getProvidedCertificate([]string{domain}); providedCertificate != nil {
		return providedCertificate, nil
	}

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
	a.jobs.In() <- func() {
		log.Info("Retrieving ACME certificates...")
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
		log.Info("Retrieved ACME certificates")
	}
}

func (a *ACME) renewCertificates() {
	a.jobs.In() <- func() {
		log.Debug("Testing certificate renew...")
		account := a.store.Get().(*Account)
		for _, certificateResource := range account.DomainsCertificate.Certs {
			if certificateResource.needRenew() {
				log.Debugf("Renewing certificate %+v", certificateResource.Domains)
				renewedCert, err := a.client.RenewCertificate(acme.CertificateResource{
					Domain:        certificateResource.Certificate.Domain,
					CertURL:       certificateResource.Certificate.CertURL,
					CertStableURL: certificateResource.Certificate.CertStableURL,
					PrivateKey:    certificateResource.Certificate.PrivateKey,
					Certificate:   certificateResource.Certificate.Certificate,
				}, true, OSCPMustStaple)
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
				transaction, object, err := a.store.Begin()
				if err != nil {
					log.Errorf("Error renewing certificate: %v", err)
					continue
				}
				account = object.(*Account)
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
	}
}

func dnsOverrideDelay(delay int) error {
	var err error
	if delay > 0 {
		log.Debugf("Delaying %d seconds rather than validating DNS propagation", delay)
		acme.PreCheckDNS = func(_, _ string) (bool, error) {
			time.Sleep(time.Duration(delay) * time.Second)
			return true, nil
		}
	} else if delay < 0 {
		err = fmt.Errorf("Invalid negative DelayDontCheckDNS: %d", delay)
	}
	return err
}

func (a *ACME) buildACMEClient(account *Account) (*acme.Client, error) {
	log.Debug("Building ACME client...")
	caServer := "https://acme-v01.api.letsencrypt.org/directory"
	if len(a.CAServer) > 0 {
		caServer = a.CAServer
	}
	client, err := acme.NewClient(caServer, account, acme.RSA4096)
	if err != nil {
		return nil, err
	}

	if len(a.DNSProvider) > 0 {
		log.Debugf("Using DNS Challenge provider: %s", a.DNSProvider)

		err = dnsOverrideDelay(a.DelayDontCheckDNS)
		if err != nil {
			return nil, err
		}

		var provider acme.ChallengeProvider
		provider, err = dns.NewDNSChallengeProviderByName(a.DNSProvider)
		if err != nil {
			return nil, err
		}

		client.ExcludeChallenges([]acme.Challenge{acme.HTTP01, acme.TLSSNI01})
		err = client.SetChallengeProvider(acme.DNS01, provider)
	} else {
		client.ExcludeChallenges([]acme.Challenge{acme.HTTP01, acme.DNS01})
		err = client.SetChallengeProvider(acme.TLSSNI01, a.challengeProvider)
	}

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
	a.jobs.In() <- func() {
		log.Debugf("LoadCertificateForDomains %v...", domains)

		if len(domains) == 0 {
			// no domain
			return
		}

		domains = fun.Map(types.CanonicalDomain, domains).([]string)

		// Check provided certificates
		if a.getProvidedCertificate(domains) != nil {
			return
		}

		operation := func() error {
			if a.client == nil {
				return errors.New("ACME client still not built")
			}
			return nil
		}
		notify := func(err error, time time.Duration) {
			log.Errorf("Error getting ACME client: %v, retrying in %s", err, time)
		}
		ebo := backoff.NewExponentialBackOff()
		ebo.MaxElapsedTime = 30 * time.Second
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), ebo, notify)
		if err != nil {
			log.Errorf("Error getting ACME client: %v", err)
			return
		}
		account := a.store.Get().(*Account)
		var domain Domain
		if len(domains) > 1 {
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
	}
}

// Get provided certificate which check a domains list (Main and SANs)
func (a *ACME) getProvidedCertificate(domains []string) *tls.Certificate {
	// Use regex to test for provided certs that might have been added into TLSConfig
	providedCertMatch := false
	log.Debugf("Look for provided certificate to validate %s...", domains)
	for k := range a.TLSConfig.NameToCertificate {
		selector := "^" + strings.Replace(k, "*.", "[^\\.]*\\.?", -1) + "$"
		for _, domainToCheck := range domains {
			providedCertMatch, _ = regexp.MatchString(selector, domainToCheck)
			if !providedCertMatch {
				break
			}
		}
		if providedCertMatch {
			log.Debugf("Got provided certificate for domains %s", domains)
			return a.TLSConfig.NameToCertificate[k]

		}
	}
	log.Debugf("No provided certificate found for domains %s, get ACME certificate.", domains)
	return nil
}

func (a *ACME) getDomainsCertificates(domains []string) (*Certificate, error) {
	domains = fun.Map(types.CanonicalDomain, domains).([]string)
	log.Debugf("Loading ACME certificates %s...", domains)
	bundle := true
	certificate, failures := a.client.ObtainCertificate(domains, bundle, nil, OSCPMustStaple)
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

func (a *ACME) runJobs() {
	safe.Go(func() {
		for job := range a.jobs.Out() {
			function := job.(func())
			function()
		}
	})
}
