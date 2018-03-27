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
	"reflect"
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
	acmeprovider "github.com/containous/traefik/provider/acme"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/tls/generate"
	"github.com/containous/traefik/types"
	"github.com/eapache/channels"
	acme "github.com/xenolf/lego/acmev2"
	"github.com/xenolf/lego/providers/dns"
)

var (
	// OSCPMustStaple enables OSCP stapling as from https://github.com/xenolf/lego/issues/270
	OSCPMustStaple = false
)

// ACME allows to connect to lets encrypt and retrieve certs
// Deprecated Please use provider/acme/Provider
type ACME struct {
	Email                 string                      `description:"Email address used for registration"`
	Domains               []types.Domain              `description:"SANs (alternative domains) to each main domain using format: --acme.domains='main.com,san1.com,san2.com' --acme.domains='main.net,san1.net,san2.net'"`
	Storage               string                      `description:"File or key used for certificates storage."`
	StorageFile           string                      // deprecated
	OnDemand              bool                        `description:"Enable on demand certificate generation. This will request a certificate from Let's Encrypt during the first TLS handshake for a hostname that does not yet have a certificate."` //deprecated
	OnHostRule            bool                        `description:"Enable certificate generation on frontends Host rules."`
	CAServer              string                      `description:"CA server to use."`
	EntryPoint            string                      `description:"Entrypoint to proxy acme challenge to."`
	DNSChallenge          *acmeprovider.DNSChallenge  `description:"Activate DNS-01 Challenge"`
	HTTPChallenge         *acmeprovider.HTTPChallenge `description:"Activate HTTP-01 Challenge"`
	DNSProvider           string                      `description:"Activate DNS-01 Challenge (Deprecated)"`                                                       // deprecated
	DelayDontCheckDNS     flaeg.Duration              `description:"Assume DNS propagates after a delay in seconds rather than finding and querying nameservers."` // deprecated
	ACMELogging           bool                        `description:"Enable debug logging of ACME actions."`
	client                *acme.Client
	defaultCertificate    *tls.Certificate
	store                 cluster.Store
	challengeHTTPProvider *challengeHTTPProvider
	checkOnDemandDomain   func(domain string) bool
	jobs                  *channels.InfiniteChannel
	TLSConfig             *tls.Config `description:"TLS config in case wildcard certs are used"`
	dynamicCerts          *safe.Safe
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
			domainsCerts := DomainsCertificates{Certs: []*DomainsCertificate{}}
			if account != nil {
				domainsCerts = account.DomainsCertificate
			}

			account, err = NewAccount(a.Email, domainsCerts.Certs)
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

			reg, err := a.client.Register(true)
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

func (a *ACME) getCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	domain := types.CanonicalDomain(clientHello.ServerName)
	account := a.store.Get().(*Account)

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

func dnsOverrideDelay(delay flaeg.Duration) error {
	var err error
	if delay > 0 {
		log.Debugf("Delaying %d rather than validating DNS propagation", delay)
		acme.PreCheckDNS = func(_, _ string) (bool, error) {
			time.Sleep(time.Duration(delay))
			return true, nil
		}
	} else if delay < 0 {
		err = fmt.Errorf("invalid negative DelayBeforeCheck: %d", delay)
	}
	return err
}

func (a *ACME) buildACMEClient(account *Account) (*acme.Client, error) {
	log.Debug("Building ACME client...")
	caServer := "https://acme-v02.api.letsencrypt.org/directory"
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

		client.ExcludeChallenges([]acme.Challenge{acme.HTTP01})
		err = client.SetChallengeProvider(acme.DNS01, provider)
	} else if a.HTTPChallenge != nil && len(a.HTTPChallenge.EntryPoint) > 0 {
		log.Debug("Using HTTP Challenge provider.")
		client.ExcludeChallenges([]acme.Challenge{acme.DNS01})
		a.challengeHTTPProvider = &challengeHTTPProvider{store: a.store}
		err = client.SetChallengeProvider(acme.HTTP01, a.challengeHTTPProvider)
	} else {
		return nil, errors.New("ACME challenge not specified, please select HTTP or DNS Challenge")
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
		var domain types.Domain
		if len(uncheckedDomains) > 1 {
			domain = types.Domain{Main: uncheckedDomains[0], SANs: uncheckedDomains[1:]}
		} else {
			domain = types.Domain{Main: uncheckedDomains[0]}
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
			domainChecked = searchProvidedCertificateForDomain(domain, certDomain)
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

func searchProvidedCertificateForDomain(domain string, certDomain string) bool {
	// Use regex to test for provided certs that might have been added into TLSConfig
	selector := "^" + strings.Replace(certDomain, "*.", "[^\\.]*\\.", -1) + "$"
	domainChecked, err := regexp.MatchString(selector, domain)
	if err != nil {
		log.Errorf("Unable to compare %q and %q : %s", domain, certDomain, err.Error())
	}
	return domainChecked
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

// getValidDomains checks if given domain is allowed to generate a ACME certificate and return it
func (a *ACME) getValidDomains(domains []string, wildcardAllowed bool) ([]string, error) {
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

		if len(domains) > 1 {
			return nil, fmt.Errorf("unable to generate a wildcard certificate for domain %q : SANs are not allowed", strings.Join(domains, ","))
		}
	} else {
		for _, san := range domains[1:] {
			if strings.HasPrefix(san, "*") {
				return nil, fmt.Errorf("unable to generate a certificate in ACME provider for domains %q: SANs can not be a wildcard domain", strings.Join(domains, ","))
			}
		}
	}

	domains = fun.Map(types.CanonicalDomain, domains).([]string)
	return domains, nil
}

func isDomainAlreadyChecked(domainToCheck string, existentDomains map[string]*tls.Certificate) bool {
	for certDomains := range existentDomains {
		for _, certDomain := range strings.Split(certDomains, ",") {
			if searchProvidedCertificateForDomain(domainToCheck, certDomain) {
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
			} else if strings.HasPrefix(domain.Main, "*") && domain.SANs == nil {
				// Check if domains can be validated by the wildcard domain

				var newDomainsToCheck []string

				// Check if domains can be validated by the wildcard domain
				domainsMap := make(map[string]*tls.Certificate)
				domainsMap[domain.Main] = &tls.Certificate{}

				for _, domainProcessed := range domainToCheck.ToStrArray() {
					if isDomainAlreadyChecked(domainProcessed, domainsMap) {
						log.Warnf("Domain %q will not be processed by ACME because it is validated by the wildcard %q", domainProcessed, domain.Main)
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
		}

		if keepDomain {
			newDomains = append(newDomains, domainToCheck)
		}
	}

	a.Domains = newDomains
}
