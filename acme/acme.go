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
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/flaeg"
	"github.com/containous/mux"
	"github.com/containous/staert"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/log"
	acmeprovider "github.com/containous/traefik/provider/acme"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/containous/traefik/version"
	"github.com/eapache/channels"
	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/challenge"
	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/challenge/http01"
	"github.com/go-acme/lego/lego"
	legolog "github.com/go-acme/lego/log"
	"github.com/go-acme/lego/providers/dns"
	"github.com/go-acme/lego/registration"
	"github.com/sirupsen/logrus"
)

var (
	// OSCPMustStaple enables OSCP stapling as from https://github.com/go-acme/lego/issues/270
	OSCPMustStaple = false
)

// ACME allows to connect to lets encrypt and retrieve certs
// Deprecated Please use provider/acme/Provider
type ACME struct {
	Email                 string                      `description:"Email address used for registration"`
	Domains               []types.Domain              `description:"SANs (alternative domains) to each main domain using format: --acme.domains='main.com,san1.com,san2.com' --acme.domains='main.net,san1.net,san2.net'"`
	Storage               string                      `description:"File or key used for certificates storage."`
	StorageFile           string                      // Deprecated
	OnDemand              bool                        `description:"(Deprecated) Enable on demand certificate generation. This will request a certificate from Let's Encrypt during the first TLS handshake for a hostname that does not yet have a certificate."` // Deprecated
	OnHostRule            bool                        `description:"Enable certificate generation on frontends Host rules."`
	CAServer              string                      `description:"CA server to use."`
	EntryPoint            string                      `description:"Entrypoint to proxy acme challenge to."`
	KeyType               string                      `description:"KeyType used for generating certificate private key. Allow value 'EC256', 'EC384', 'RSA2048', 'RSA4096', 'RSA8192'. Default to 'RSA4096'"`
	DNSChallenge          *acmeprovider.DNSChallenge  `description:"Activate DNS-01 Challenge"`
	HTTPChallenge         *acmeprovider.HTTPChallenge `description:"Activate HTTP-01 Challenge"`
	TLSChallenge          *acmeprovider.TLSChallenge  `description:"Activate TLS-ALPN-01 Challenge"`
	DNSProvider           string                      `description:"(Deprecated) Activate DNS-01 Challenge"`                                                                    // Deprecated
	DelayDontCheckDNS     flaeg.Duration              `description:"(Deprecated) Assume DNS propagates after a delay in seconds rather than finding and querying nameservers."` // Deprecated
	ACMELogging           bool                        `description:"Enable debug logging of ACME actions."`
	OverrideCertificates  bool                        `description:"Enable to override certificates in key-value store when using storeconfig"`
	client                *lego.Client
	store                 cluster.Store
	challengeHTTPProvider *challengeHTTPProvider
	challengeTLSProvider  *challengeTLSProvider
	checkOnDemandDomain   func(domain string) bool
	jobs                  *channels.InfiniteChannel
	TLSConfig             *tls.Config `description:"TLS config in case wildcard certs are used"`
	dynamicCerts          *safe.Safe
	resolvingDomains      map[string]struct{}
	resolvingDomainsMutex sync.RWMutex
}

func (a *ACME) init() error {
	if a.ACMELogging {
		legolog.Logger = fmtlog.New(log.WriterLevel(logrus.InfoLevel), "legolog: ", 0)
	} else {
		legolog.Logger = fmtlog.New(ioutil.Discard, "", 0)
	}

	a.jobs = channels.NewInfiniteChannel()

	// Init the currently resolved domain map
	a.resolvingDomains = make(map[string]struct{})

	return nil
}

// AddRoutes add routes on internal router
func (a *ACME) AddRoutes(router *mux.Router) {
	router.Methods(http.MethodGet).
		Path(http01.ChallengePath("{token}")).
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
		return errors.New("empty Store, please provide a key for certs storage")
	}

	a.checkOnDemandDomain = checkOnDemandDomain
	a.dynamicCerts = certs

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
		// Reset Account values if caServer changed, thus registration URI can be updated
		if account != nil && account.Registration != nil && !isAccountMatchingCaServer(account.Registration.URI, a.CAServer) {
			log.Info("Account URI does not match the current CAServer. The account will be reset")
			account.reset()
		}

		var needRegister bool
		if account == nil || len(account.Email) == 0 {
			domainsCerts := DomainsCertificates{Certs: []*DomainsCertificate{}}
			if account != nil {
				domainsCerts = account.DomainsCertificate
			}

			account, err = NewAccount(a.Email, domainsCerts.Certs, a.KeyType)
			if err != nil {
				return err
			}

			needRegister = true
		} else if len(account.KeyType) == 0 {
			// Set the KeyType if not already defined in the account
			account.KeyType = acmeprovider.GetKeyType(a.KeyType)
		}

		a.client, err = a.buildACMEClient(account)
		if err != nil {
			return err
		}
		if needRegister {
			// New users will need to register; be sure to save it
			log.Debug("Register...")

			reg, err := a.client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
			if err != nil {
				return err
			}

			account.Registration = reg
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

func isAccountMatchingCaServer(accountURI string, serverURI string) bool {
	aru, err := url.Parse(accountURI)
	if err != nil {
		log.Infof("Unable to parse account.Registration URL : %v", err)
		return false
	}
	cau, err := url.Parse(serverURI)
	if err != nil {
		log.Infof("Unable to parse CAServer URL : %v", err)
		return false
	}
	return cau.Hostname() == aru.Hostname()
}

func (a *ACME) getCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	domain := types.CanonicalDomain(clientHello.ServerName)
	account := a.store.Get().(*Account)

	if challengeCert, ok := a.challengeTLSProvider.getCertificate(domain); ok {
		log.Debugf("ACME got challenge %s", domain)
		return challengeCert, nil
	}

	if providedCertificate := a.getProvidedCertificate(domain); providedCertificate != nil {
		return providedCertificate, nil
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

		a.deleteUnnecessaryDomains()

		for i := 0; i < len(a.Domains); i++ {
			domain := a.Domains[i]

			// check if cert isn't already loaded
			account := a.store.Get().(*Account)
			if _, exists := account.DomainsCertificate.exists(domain); !exists {
				var domains []string
				domains = append(domains, domain.Main)
				domains = append(domains, domain.SANs...)
				domains, err := a.getValidDomains(domains, true)
				if err != nil {
					log.Errorf("Error validating ACME certificate for domain %q: %s", domains, err)
					continue
				}

				certificateResource, err := a.getDomainsCertificates(domains)
				if err != nil {
					log.Errorf("Error getting ACME certificate for domain %q: %s", domains, err)
					continue
				}

				transaction, object, err := a.store.Begin()
				if err != nil {
					log.Errorf("Error creating ACME store transaction from domain %q: %s", domain, err)
					continue
				}

				account = object.(*Account)
				_, err = account.DomainsCertificate.addCertificateForDomains(certificateResource, domain)
				if err != nil {
					log.Errorf("Error adding ACME certificate for domain %q: %s", domains, err)
					continue
				}

				if err = transaction.Commit(account); err != nil {
					log.Errorf("Error Saving ACME account %+v: %s", account, err)
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
					return a.storeRenewedCertificate(certificateResource, renewedACMECert)
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
	renewedCert, err := a.client.Certificate.Renew(certificate.Resource{
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

func (a *ACME) storeRenewedCertificate(certificateResource *DomainsCertificate, renewedACMECert *Certificate) error {
	transaction, object, err := a.store.Begin()
	if err != nil {
		return fmt.Errorf("error during transaction initialization for renewing certificate: %v", err)
	}

	log.Infof("Renewing certificate in data store : %+v ", certificateResource.Domains)
	account := object.(*Account)
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

func (a *ACME) buildACMEClient(account *Account) (*lego.Client, error) {
	log.Debug("Building ACME client...")
	caServer := "https://acme-v02.api.letsencrypt.org/directory"
	if len(a.CAServer) > 0 {
		caServer = a.CAServer
	}

	config := lego.NewConfig(account)
	config.CADirURL = caServer
	config.Certificate.KeyType = account.KeyType
	config.UserAgent = fmt.Sprintf("containous-traefik/%s", version.Version)

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	// DNS challenge
	if a.DNSChallenge != nil && len(a.DNSChallenge.Provider) > 0 {
		log.Debugf("Using DNS Challenge provider: %s", a.DNSChallenge.Provider)

		var provider challenge.Provider
		provider, err = dns.NewDNSChallengeProviderByName(a.DNSChallenge.Provider)
		if err != nil {
			return nil, err
		}

		err = client.Challenge.SetDNS01Provider(provider,
			dns01.CondOption(len(a.DNSChallenge.Resolvers) > 0, dns01.AddRecursiveNameservers(a.DNSChallenge.Resolvers)),
			dns01.CondOption(a.DNSChallenge.DisablePropagationCheck || a.DNSChallenge.DelayBeforeCheck > 0,
				dns01.AddPreCheck(func(_, _ string) (bool, error) {
					if a.DNSChallenge.DelayBeforeCheck > 0 {
						log.Debugf("Delaying %d rather than validating DNS propagation now.", a.DNSChallenge.DelayBeforeCheck)
						time.Sleep(time.Duration(a.DNSChallenge.DelayBeforeCheck))
					}
					return true, nil
				})),
		)
		return client, err
	}

	// HTTP challenge
	if a.HTTPChallenge != nil && len(a.HTTPChallenge.EntryPoint) > 0 {
		log.Debug("Using HTTP Challenge provider.")

		a.challengeHTTPProvider = &challengeHTTPProvider{store: a.store}
		err = client.Challenge.SetHTTP01Provider(a.challengeHTTPProvider)
		return client, err
	}

	// TLS Challenge
	if a.TLSChallenge != nil {
		log.Debug("Using TLS Challenge provider.")

		err = client.Challenge.SetTLSALPN01Provider(a.challengeTLSProvider)
		return client, err
	}

	return nil, errors.New("ACME challenge not specified, please select TLS or HTTP or DNS Challenge")
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
	cert, err := account.DomainsCertificate.addCertificateForDomains(certificate, types.Domain{Main: domain})
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

		domains, err := a.getValidDomains(domains, false)
		if err != nil {
			log.Errorf("Error getting valid domain: %v", err)
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
		err = backoff.RetryNotify(safe.OperationWithRecover(operation), ebo, notify)
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

		a.addResolvingDomains(uncheckedDomains)
		defer a.removeResolvingDomains(uncheckedDomains)

		cert, err := a.getDomainsCertificates(uncheckedDomains)
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
		var domain types.Domain
		if len(uncheckedDomains) > 1 {
			domain = types.Domain{Main: uncheckedDomains[0], SANs: uncheckedDomains[1:]}
		} else {
			domain = types.Domain{Main: uncheckedDomains[0]}
		}
		account = object.(*Account)
		_, err = account.DomainsCertificate.addCertificateForDomains(cert, domain)
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

func (a *ACME) addResolvingDomains(resolvingDomains []string) {
	a.resolvingDomainsMutex.Lock()
	defer a.resolvingDomainsMutex.Unlock()

	for _, domain := range resolvingDomains {
		a.resolvingDomains[domain] = struct{}{}
	}
}

func (a *ACME) removeResolvingDomains(resolvingDomains []string) {
	a.resolvingDomainsMutex.Lock()
	defer a.resolvingDomainsMutex.Unlock()

	for _, domain := range resolvingDomains {
		delete(a.resolvingDomains, domain)
	}
}

// Get provided certificate which check a domains list (Main and SANs)
// from static and dynamic provided certificates
func (a *ACME) getProvidedCertificate(domains string) *tls.Certificate {
	log.Debugf("Looking for provided certificate to validate %s...", domains)
	cert := searchProvidedCertificateForDomains(domains, a.TLSConfig.NameToCertificate)
	if cert == nil && a.dynamicCerts != nil && a.dynamicCerts.Get() != nil {
		cert = searchProvidedCertificateForDomains(domains, a.dynamicCerts.Get().(map[string]*tls.Certificate))
	}
	if cert == nil {
		log.Debugf("No provided certificate found for domains %s, get ACME certificate.", domains)
	}
	return cert
}

func searchProvidedCertificateForDomains(domain string, certs map[string]*tls.Certificate) *tls.Certificate {
	// Use regex to test for provided certs that might have been added into TLSConfig
	for certDomains := range certs {
		domainChecked := false
		for _, certDomain := range strings.Split(certDomains, ",") {
			domainChecked = types.MatchDomain(domain, certDomain)
			if domainChecked {
				break
			}
		}
		if domainChecked {
			log.Debugf("Domain %q checked by provided certificate %q", domain, certDomains)
			return certs[certDomains]
		}
	}
	return nil
}

// Get provided certificate which check a domains list (Main and SANs)
// from static and dynamic provided certificates
func (a *ACME) getUncheckedDomains(domains []string, account *Account) []string {
	a.resolvingDomainsMutex.RLock()
	defer a.resolvingDomainsMutex.RUnlock()

	log.Debugf("Looking for provided certificate to validate %s...", domains)
	allCerts := make(map[string]*tls.Certificate)

	// Get static certificates
	for domains, certificate := range a.TLSConfig.NameToCertificate {
		allCerts[domains] = certificate
	}

	// Get dynamic certificates
	if a.dynamicCerts != nil && a.dynamicCerts.Get() != nil {
		for domains, certificate := range a.dynamicCerts.Get().(map[string]*tls.Certificate) {
			allCerts[domains] = certificate
		}
	}

	// Get ACME certificates
	if account != nil {
		for domains, certificate := range account.DomainsCertificate.toDomainsMap() {
			allCerts[domains] = certificate
		}
	}

	// Get currently resolved domains
	for domain := range a.resolvingDomains {
		if _, ok := allCerts[domain]; !ok {
			allCerts[domain] = &tls.Certificate{}
		}
	}

	// Get Configuration Domains
	for i := 0; i < len(a.Domains); i++ {
		allCerts[a.Domains[i].Main] = &tls.Certificate{}
		for _, san := range a.Domains[i].SANs {
			allCerts[san] = &tls.Certificate{}
		}
	}

	return searchUncheckedDomains(domains, allCerts)
}

func searchUncheckedDomains(domains []string, certs map[string]*tls.Certificate) []string {
	var uncheckedDomains []string
	for _, domainToCheck := range domains {
		if !isDomainAlreadyChecked(domainToCheck, certs) {
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
	var cleanDomains []string
	for _, domain := range domains {
		canonicalDomain := types.CanonicalDomain(domain)
		cleanDomain := dns01.UnFqdn(canonicalDomain)
		if canonicalDomain != cleanDomain {
			log.Warnf("FQDN detected, please remove the trailing dot: %s", canonicalDomain)
		}
		cleanDomains = append(cleanDomains, cleanDomain)
	}

	log.Debugf("Loading ACME certificates %s...", cleanDomains)
	bundle := true

	request := certificate.ObtainRequest{
		Domains:    cleanDomains,
		Bundle:     bundle,
		MustStaple: OSCPMustStaple,
	}

	cert, err := a.client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("cannot obtain certificates: %+v", err)
	}

	log.Debugf("Loaded ACME certificates %s", cleanDomains)
	return &Certificate{
		Domain:        cert.Domain,
		CertURL:       cert.CertURL,
		CertStableURL: cert.CertStableURL,
		PrivateKey:    cert.PrivateKey,
		Certificate:   cert.Certificate,
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

// getValidDomains checks if given domain is allowed to generate a ACME certificate and return it
func (a *ACME) getValidDomains(domains []string, wildcardAllowed bool) ([]string, error) {
	// Check if the domains array is empty or contains only one empty value
	if len(domains) == 0 || (len(domains) == 1 && len(domains[0]) == 0) {
		return nil, errors.New("unable to generate a certificate when no domain is given")
	}

	if strings.HasPrefix(domains[0], "*") {
		if !wildcardAllowed {
			return nil, fmt.Errorf("unable to generate a wildcard certificate for domain %q from a 'Host' rule", strings.Join(domains, ","))
		}

		if a.DNSChallenge == nil && len(a.DNSProvider) == 0 {
			return nil, fmt.Errorf("unable to generate a wildcard certificate for domain %q : ACME needs a DNSChallenge", strings.Join(domains, ","))
		}
		if strings.HasPrefix(domains[0], "*.*") {
			return nil, fmt.Errorf("unable to generate a wildcard certificate for domain %q : ACME does not allow '*.*' wildcard domain", strings.Join(domains, ","))
		}
	}

	domains = fun.Map(types.CanonicalDomain, domains).([]string)
	return domains, nil
}

func isDomainAlreadyChecked(domainToCheck string, existentDomains map[string]*tls.Certificate) bool {
	for certDomains := range existentDomains {
		for _, certDomain := range strings.Split(certDomains, ",") {
			if types.MatchDomain(domainToCheck, certDomain) {
				return true
			}
		}
	}
	return false
}

// deleteUnnecessaryDomains deletes from the configuration :
// - Duplicated domains
// - Domains which are checked by wildcard domain
func (a *ACME) deleteUnnecessaryDomains() {
	var newDomains []types.Domain

	for idxDomainToCheck, domainToCheck := range a.Domains {
		keepDomain := true

		for idxDomain, domain := range a.Domains {
			if idxDomainToCheck == idxDomain {
				continue
			}

			if reflect.DeepEqual(domain, domainToCheck) {
				if idxDomainToCheck > idxDomain {
					log.Warnf("The domain %v is duplicated in the configuration but will be process by ACME only once.", domainToCheck)
					keepDomain = false
				}
				break
			}

			var newDomainsToCheck []string

			// Check if domains can be validated by the wildcard domain
			domainsMap := make(map[string]*tls.Certificate)
			domainsMap[domain.Main] = &tls.Certificate{}
			if len(domain.SANs) > 0 {
				domainsMap[strings.Join(domain.SANs, ",")] = &tls.Certificate{}
			}

			for _, domainProcessed := range domainToCheck.ToStrArray() {
				if idxDomain < idxDomainToCheck && isDomainAlreadyChecked(domainProcessed, domainsMap) {
					// The domain is duplicated in a CN
					log.Warnf("Domain %q is duplicated in the configuration or validated by the domain %v. It will be processed once.", domainProcessed, domain)
					continue
				} else if domain.Main != domainProcessed && strings.HasPrefix(domain.Main, "*") && types.MatchDomain(domainProcessed, domain.Main) {
					// Check if a wildcard can validate the domain
					log.Warnf("Domain %q will not be processed by ACME provider because it is validated by the wildcard %q", domainProcessed, domain.Main)
					continue
				}
				newDomainsToCheck = append(newDomainsToCheck, domainProcessed)
			}

			// Delete the domain if both Main and SANs can be validated by the wildcard domain
			// otherwise keep the unchecked values
			if newDomainsToCheck == nil {
				keepDomain = false
				break
			}
			domainToCheck.Set(newDomainsToCheck)
		}

		if keepDomain {
			newDomains = append(newDomains, domainToCheck)
		}
	}

	a.Domains = newDomains
}
