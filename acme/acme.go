package acme

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	fmtlog "log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/flaeg"
	"github.com/containous/mux"
	"github.com/containous/staert"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	traefikTls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/tls/generate"
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
	Email                 string         `description:"Email address used for registration"`
	Domains               []Domain       `description:"SANs (alternative domains) to each main domain using format: --acme.domains='main.com,san1.com,san2.com' --acme.domains='main.net,san1.net,san2.net'"`
	Storage               string         `description:"File or key used for certificates storage."`
	StorageFile           string         // deprecated
	OnDemand              bool           `description:"Enable on demand certificate generation. This will request a certificate from Let's Encrypt during the first TLS handshake for a hostname that does not yet have a certificate."` //deprecated
	OnHostRule            bool           `description:"Enable certificate generation on frontends Host rules."`
	CAServer              string         `description:"CA server to use."`
	EntryPoint            string         `description:"Entrypoint to proxy acme challenge to."`
	DNSChallenge          *DNSChallenge  `description:"Activate DNS-01 Challenge"`
	HTTPChallenge         *HTTPChallenge `description:"Activate HTTP-01 Challenge"`
	DNSProvider           string         `description:"Use a DNS-01 acme challenge rather than TLS-SNI-01 challenge."`                                // deprecated
	DelayDontCheckDNS     flaeg.Duration `description:"Assume DNS propagates after a delay in seconds rather than finding and querying nameservers."` // deprecated
	ACMELogging           bool           `description:"Enable debug logging of ACME actions."`
	client                *acme.Client
	defaultCertificate    *tls.Certificate
	store                 cluster.Store
	challengeTLSProvider  *challengeTLSProvider
	challengeHTTPProvider *challengeHTTPProvider
	checkOnDemandDomain   func(domain string) bool
	jobs                  *channels.InfiniteChannel
	TLSConfig             *tls.Config `description:"TLS config in case wildcard certs are used"`
	dynamicCerts          *safe.Safe
}

// DNSChallenge contains DNS challenge Configuration
type DNSChallenge struct {
	Provider         string         `description:"Use a DNS-01 based challenge provider rather than HTTPS."`
	DelayBeforeCheck flaeg.Duration `description:"Assume DNS propagates after a delay in seconds rather than finding and querying nameservers."`
}

// HTTPChallenge contains HTTP challenge Configuration
type HTTPChallenge struct {
	EntryPoint string `description:"HTTP challenge EntryPoint"`
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
	// FIXME temporary fix, waiting for https://github.com/xenolf/lego/pull/478
	acme.HTTPClient = http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 15 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	if a.ACMELogging {
		acme.Logger = fmtlog.New(os.Stderr, "legolog: ", fmtlog.LstdFlags)
	} else {
		acme.Logger = fmtlog.New(ioutil.Discard, "", 0)
	}
	// no certificates in TLS config, so we add a default one
	cert, err := generate.DefaultCertificate()
	if err != nil {
		return err
	}
	a.defaultCertificate = cert

	a.jobs = channels.NewInfiniteChannel()
	return nil
}

// AddRoutes add routes on internal router
func (a *ACME) AddRoutes(router *mux.Router) {
	router.Methods(http.MethodGet).
		Path(acme.HTTP01ChallengePath("{token}")).
		Handler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if a.challengeHTTPProvider == nil {
				rw.WriteHeader(http.StatusNotFound)
				return
			}

			vars := mux.Vars(req)
			if token, ok := vars["token"]; ok {
				domain, _, err := net.SplitHostPort(req.Host)
				if err != nil {
					log.Debugf("Unable to split host and port: %v. Fallback to request host.", err)
					domain = req.Host
				}
				tokenValue := a.challengeHTTPProvider.getTokenValue(token, domain)
				if len(tokenValue) > 0 {
					rw.WriteHeader(http.StatusOK)
					rw.Write(tokenValue)
					return
				}
			}
			rw.WriteHeader(http.StatusNotFound)
		}))
}

// CreateClusterConfig creates a tls.config using ACME configuration in cluster mode
func (a *ACME) CreateClusterConfig(leadership *cluster.Leadership, tlsConfig *tls.Config, certs *safe.Safe, checkOnDemandDomain func(domain string) bool) error {
	err := a.init()
	if err != nil {
		return err
	}
	if len(a.Storage) == 0 {
		return errors.New("Empty Store, please provide a key for certs storage")
	}
	a.checkOnDemandDomain = checkOnDemandDomain
	a.dynamicCerts = certs
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
	a.challengeTLSProvider = &challengeTLSProvider{store: a.store}

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

	leadership.AddListener(a.leadershipListener)
	return nil
}

func (a *ACME) leadershipListener(elected bool) error {
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
			log.Debug(err)
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
}

// CreateLocalConfig creates a tls.config using local ACME configuration
func (a *ACME) CreateLocalConfig(tlsConfig *tls.Config, certs *safe.Safe, checkOnDemandDomain func(domain string) bool) error {
	defer a.runJobs()
	err := a.init()
	if err != nil {
		return err
	}
	if len(a.Storage) == 0 {
		return errors.New("Empty Store, please provide a filename for certs storage")
	}
	a.checkOnDemandDomain = checkOnDemandDomain
	a.dynamicCerts = certs
	tlsConfig.Certificates = append(tlsConfig.Certificates, *a.defaultCertificate)
	tlsConfig.GetCertificate = a.getCertificate
	a.TLSConfig = tlsConfig
	localStore := NewLocalStore(a.Storage)
	a.store = localStore
	a.challengeTLSProvider = &challengeTLSProvider{store: a.store}

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
		log.Errorf(`Failed to build ACME client: %s
Let's Encrypt functionality will be limited until Traefik is restarted.`, err)
		return nil
	}

	if needRegister {
		// New users will need to register; be sure to save it
		log.Info("Register...")
		reg, err := a.client.Register()
		if err != nil {
			log.Errorf(`Failed to register user: %s
Let's Encrypt functionality will be limited until Traefik is restarted.`, err)
			return nil
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
			log.Errorf(`Failed to renew subscriber agreement: %s
Let's Encrypt functionality will be limited until Traefik is restarted.`, err)
			return nil
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

	if providedCertificate := a.getProvidedCertificate(domain); providedCertificate != nil {
		return providedCertificate, nil
	}

	if challengeCert, ok := a.challengeTLSProvider.getCertificate(domain); ok {
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
	log.Debugf("No certificate found or generated for %s", domain)
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
		log.Info("Testing certificate renew...")
		account := a.store.Get().(*Account)
		for _, certificateResource := range account.DomainsCertificate.Certs {
			if certificateResource.needRenew() {
				log.Infof("Renewing certificate from LE : %+v", certificateResource.Domains)
				renewedACMECert, err := a.renewACMECertificate(certificateResource)
				if err != nil {
					log.Errorf("Error renewing certificate from LE: %v", err)
					continue
				}
				operation := func() error {
					return a.storeRenewedCertificate(account, certificateResource, renewedACMECert)
				}
				notify := func(err error, time time.Duration) {
					log.Warnf("Renewed certificate storage error: %v, retrying in %s", err, time)
				}
				ebo := backoff.NewExponentialBackOff()
				ebo.MaxElapsedTime = 60 * time.Second
				err = backoff.RetryNotify(safe.OperationWithRecover(operation), ebo, notify)
				if err != nil {
					log.Errorf("Datastore cannot sync: %v", err)
					continue
				}
			}
		}
	}
}

func (a *ACME) renewACMECertificate(certificateResource *DomainsCertificate) (*Certificate, error) {
	renewedCert, err := a.client.RenewCertificate(acme.CertificateResource{
		Domain:        certificateResource.Certificate.Domain,
		CertURL:       certificateResource.Certificate.CertURL,
		CertStableURL: certificateResource.Certificate.CertStableURL,
		PrivateKey:    certificateResource.Certificate.PrivateKey,
		Certificate:   certificateResource.Certificate.Certificate,
	}, true, OSCPMustStaple)
	if err != nil {
		return nil, err
	}
	log.Infof("Renewed certificate from  LE: %+v", certificateResource.Domains)
	return &Certificate{
		Domain:        renewedCert.Domain,
		CertURL:       renewedCert.CertURL,
		CertStableURL: renewedCert.CertStableURL,
		PrivateKey:    renewedCert.PrivateKey,
		Certificate:   renewedCert.Certificate,
	}, nil
}

func (a *ACME) storeRenewedCertificate(account *Account, certificateResource *DomainsCertificate, renewedACMECert *Certificate) error {
	transaction, object, err := a.store.Begin()
	if err != nil {
		return fmt.Errorf("error during transaction initialization for renewing certificate: %v", err)
	}

	log.Infof("Renewing certificate in data store : %+v ", certificateResource.Domains)
	account = object.(*Account)
	err = account.DomainsCertificate.renewCertificates(renewedACMECert, certificateResource.Domains)
	if err != nil {
		return fmt.Errorf("error renewing certificate in datastore: %v ", err)
	}

	log.Infof("Commit certificate renewed in data store : %+v", certificateResource.Domains)
	if err = transaction.Commit(account); err != nil {
		return fmt.Errorf("error saving ACME account %+v: %v", account, err)
	}

	oldAccount := a.store.Get().(*Account)
	for _, oldCertificateResource := range oldAccount.DomainsCertificate.Certs {
		if oldCertificateResource.Domains.Main == certificateResource.Domains.Main && strings.Join(oldCertificateResource.Domains.SANs, ",") == strings.Join(certificateResource.Domains.SANs, ",") && certificateResource.Certificate != renewedACMECert {
			return fmt.Errorf("renewed certificate not stored: %+v", certificateResource.Domains)
		}
	}

	log.Infof("Certificate successfully renewed in data store: %+v", certificateResource.Domains)
	return nil
}

func dnsOverrideDelay(delay flaeg.Duration) error {
	var err error
	if delay > 0 {
		log.Debugf("Delaying %d rather than validating DNS propagation", delay)
		acme.PreCheckDNS = func(_, _ string) (bool, error) {
			time.Sleep(time.Duration(delay) * time.Second)
			return true, nil
		}
	} else if delay < 0 {
		err = fmt.Errorf("invalid negative DelayBeforeCheck: %d", delay)
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

	if a.DNSChallenge != nil && len(a.DNSChallenge.Provider) > 0 {
		log.Debugf("Using DNS Challenge provider: %s", a.DNSChallenge.Provider)

		err = dnsOverrideDelay(a.DNSChallenge.DelayBeforeCheck)
		if err != nil {
			return nil, err
		}

		var provider acme.ChallengeProvider
		provider, err = dns.NewDNSChallengeProviderByName(a.DNSChallenge.Provider)
		if err != nil {
			return nil, err
		}

		client.ExcludeChallenges([]acme.Challenge{acme.HTTP01, acme.TLSSNI01})
		err = client.SetChallengeProvider(acme.DNS01, provider)
	} else if a.HTTPChallenge != nil && len(a.HTTPChallenge.EntryPoint) > 0 {
		client.ExcludeChallenges([]acme.Challenge{acme.DNS01, acme.TLSSNI01})
		a.challengeHTTPProvider = &challengeHTTPProvider{store: a.store}
		err = client.SetChallengeProvider(acme.HTTP01, a.challengeHTTPProvider)
	} else {
		client.ExcludeChallenges([]acme.Challenge{acme.HTTP01, acme.DNS01})
		err = client.SetChallengeProvider(acme.TLSSNI01, a.challengeTLSProvider)
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

		// Check provided certificates
		uncheckedDomains := a.getUncheckedDomains(domains, account)
		if len(uncheckedDomains) == 0 {
			return
		}
		certificate, err := a.getDomainsCertificates(uncheckedDomains)
		if err != nil {
			log.Errorf("Error getting ACME certificates %+v : %v", uncheckedDomains, err)
			return
		}
		log.Debugf("Got certificate for domains %+v", uncheckedDomains)
		transaction, object, err := a.store.Begin()

		if err != nil {
			log.Errorf("Error creating transaction %+v : %v", uncheckedDomains, err)
			return
		}
		var domain Domain
		if len(uncheckedDomains) > 1 {
			domain = Domain{Main: uncheckedDomains[0], SANs: uncheckedDomains[1:]}
		} else {
			domain = Domain{Main: uncheckedDomains[0]}
		}
		account = object.(*Account)
		_, err = account.DomainsCertificate.addCertificateForDomains(certificate, domain)
		if err != nil {
			log.Errorf("Error adding ACME certificates %+v : %v", uncheckedDomains, err)
			return
		}
		if err = transaction.Commit(account); err != nil {
			log.Errorf("Error Saving ACME account %+v: %v", account, err)
			return
		}
	}
}

// Get provided certificate which check a domains list (Main and SANs)
// from static and dynamic provided certificates
func (a *ACME) getProvidedCertificate(domains string) *tls.Certificate {
	log.Debugf("Looking for provided certificate to validate %s...", domains)
	cert := searchProvidedCertificateForDomains(domains, a.TLSConfig.NameToCertificate)
	if cert == nil && a.dynamicCerts != nil && a.dynamicCerts.Get() != nil {
		cert = searchProvidedCertificateForDomains(domains, a.dynamicCerts.Get().(*traefikTls.DomainsCertificates).Get().(map[string]*tls.Certificate))
	}
	if cert == nil {
		log.Debugf("No provided certificate found for domains %s, get ACME certificate.", domains)
	}
	return cert
}

func searchProvidedCertificateForDomains(domain string, certs map[string]*tls.Certificate) *tls.Certificate {
	// Use regex to test for provided certs that might have been added into TLSConfig
	for certDomains := range certs {
		domainCheck := false
		for _, certDomain := range strings.Split(certDomains, ",") {
			selector := "^" + strings.Replace(certDomain, "*.", "[^\\.]*\\.?", -1) + "$"
			domainCheck, _ = regexp.MatchString(selector, domain)
			if domainCheck {
				break
			}
		}
		if domainCheck {
			log.Debugf("Domain %q checked by provided certificate %q", domain, certDomains)
			return certs[certDomains]
		}
	}
	return nil
}

// Get provided certificate which check a domains list (Main and SANs)
// from static and dynamic provided certificates
func (a *ACME) getUncheckedDomains(domains []string, account *Account) []string {
	log.Debugf("Looking for provided certificate to validate %s...", domains)
	allCerts := make(map[string]*tls.Certificate)

	// Get static certificates
	for domains, certificate := range a.TLSConfig.NameToCertificate {
		allCerts[domains] = certificate
	}

	// Get dynamic certificates
	if a.dynamicCerts != nil && a.dynamicCerts.Get() != nil {
		for domains, certificate := range a.dynamicCerts.Get().(*traefikTls.DomainsCertificates).Get().(map[string]*tls.Certificate) {
			allCerts[domains] = certificate
		}
	}

	// Get ACME certificates
	if account != nil {
		for domains, certificate := range account.DomainsCertificate.toDomainsMap() {
			allCerts[domains] = certificate
		}
	}

	return searchUncheckedDomains(domains, allCerts)
}

func searchUncheckedDomains(domains []string, certs map[string]*tls.Certificate) []string {
	uncheckedDomains := []string{}
	for _, domainToCheck := range domains {
		domainCheck := false
		for certDomains := range certs {
			domainCheck = false
			for _, certDomain := range strings.Split(certDomains, ",") {
				// Use regex to test for provided certs that might have been added into TLSConfig
				selector := "^" + strings.Replace(certDomain, "*.", "[^\\.]*\\.?", -1) + "$"
				domainCheck, _ = regexp.MatchString(selector, domainToCheck)
				if domainCheck {
					break
				}
			}
			if domainCheck {
				break
			}
		}
		if !domainCheck {
			uncheckedDomains = append(uncheckedDomains, domainToCheck)
		}
	}
	if len(uncheckedDomains) == 0 {
		log.Debugf("No ACME certificate to generate for domains %q.", domains)
	} else {
		log.Debugf("Domains %q need ACME certificates generation for domains %q.", domains, strings.Join(uncheckedDomains, ","))
	}
	return uncheckedDomains
}

func (a *ACME) getDomainsCertificates(domains []string) (*Certificate, error) {
	domains = fun.Map(types.CanonicalDomain, domains).([]string)
	log.Debugf("Loading ACME certificates %s...", domains)
	bundle := true
	certificate, failures := a.client.ObtainCertificate(domains, bundle, nil, OSCPMustStaple)
	if len(failures) > 0 {
		log.Error(failures)
		return nil, fmt.Errorf("cannot obtain certificates %+v", failures)
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
