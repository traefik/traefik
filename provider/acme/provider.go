package acme

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	fmtlog "log"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/flaeg"
	"github.com/containous/mux"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/rules"
	"github.com/containous/traefik/safe"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/pkg/errors"
	acme "github.com/xenolf/lego/acmev2"
	"github.com/xenolf/lego/providers/dns"
)

var (
	// OSCPMustStaple enables OSCP stapling as from https://github.com/xenolf/lego/issues/270
	OSCPMustStaple = false
)

// Configuration holds ACME configuration provided by users
type Configuration struct {
	Email         string         `description:"Email address used for registration"`
	ACMELogging   bool           `description:"Enable debug logging of ACME actions."`
	CAServer      string         `description:"CA server to use."`
	Storage       string         `description:"Storage to use."`
	EntryPoint    string         `description:"EntryPoint to use."`
	KeyType       string         `description:"KeyType used for generating certificate private key. Allow value 'EC256', 'EC384', 'RSA2048', 'RSA4096', 'RSA8192'. Default to 'RSA4096'"`
	OnHostRule    bool           `description:"Enable certificate generation on frontends Host rules."`
	OnDemand      bool           `description:"Enable on demand certificate generation. This will request a certificate from Let's Encrypt during the first TLS handshake for a hostname that does not yet have a certificate."` // Deprecated
	DNSChallenge  *DNSChallenge  `description:"Activate DNS-01 Challenge"`
	HTTPChallenge *HTTPChallenge `description:"Activate HTTP-01 Challenge"`
	Domains       []types.Domain `description:"CN and SANs (alternative domains) to each main domain using format: --acme.domains='main.com,san1.com,san2.com' --acme.domains='*.main.net'. No SANs for wildcards domain. Wildcard domains only accepted with DNSChallenge"`
}

// Provider holds configurations of the provider.
type Provider struct {
	*Configuration
	Store                  Store
	certificates           []*Certificate
	account                *Account
	client                 *acme.Client
	certsChan              chan *Certificate
	configurationChan      chan<- types.ConfigMessage
	certificateStore       *traefiktls.CertificateStore
	clientMutex            sync.Mutex
	configFromListenerChan chan types.Configuration
	pool                   *safe.Pool
}

// Certificate is a struct which contains all data needed from an ACME certificate
type Certificate struct {
	Domain      types.Domain
	Certificate []byte
	Key         []byte
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

// SetConfigListenerChan initializes the configFromListenerChan
func (p *Provider) SetConfigListenerChan(configFromListenerChan chan types.Configuration) {
	p.configFromListenerChan = configFromListenerChan
}

func (p *Provider) init() error {
	if p.ACMELogging {
		acme.Logger = fmtlog.New(os.Stderr, "legolog: ", fmtlog.LstdFlags)
	} else {
		acme.Logger = fmtlog.New(ioutil.Discard, "", 0)
	}

	var err error
	if p.Store == nil {
		err = errors.New("no store found for the ACME provider")
		return err
	}

	p.account, err = p.Store.GetAccount()
	if err != nil {
		return fmt.Errorf("unable to get ACME account : %v", err)
	}

	p.certificates, err = p.Store.GetCertificates()
	if err != nil {
		return fmt.Errorf("unable to get ACME certificates : %v", err)
	}

	p.watchCertificate()
	p.watchNewDomains()

	return nil
}

func (p *Provider) initAccount() (*Account, error) {
	if p.account == nil || len(p.account.Email) == 0 {
		var err error
		p.account, err = NewAccount(p.Email, p.KeyType)
		if err != nil {
			return nil, err
		}
	}
	return p.account, nil
}

// ListenConfiguration sets a new Configuration into the configFromListenerChan
func (p *Provider) ListenConfiguration(config types.Configuration) {
	p.configFromListenerChan <- config
}

// ListenRequest resolves new certificates for a domain from an incoming request and return a valid Certificate to serve (onDemand option)
func (p *Provider) ListenRequest(domain string) (*tls.Certificate, error) {
	acmeCert, err := p.resolveCertificate(types.Domain{Main: domain}, false)
	if acmeCert == nil || err != nil {
		return nil, err
	}

	certificate, err := tls.X509KeyPair(acmeCert.Certificate, acmeCert.PrivateKey)

	return &certificate, err
}

func (p *Provider) watchNewDomains() {
	p.pool.Go(func(stop chan bool) {
		for {
			select {
			case config := <-p.configFromListenerChan:
				for _, frontend := range config.Frontends {
					for _, route := range frontend.Routes {
						domainRules := rules.Rules{}
						domains, err := domainRules.ParseDomains(route.Rule)
						if err != nil {
							log.Errorf("Error parsing domains in provider ACME: %v", err)
							continue
						}

						if len(domains) == 0 {
							log.Debugf("No domain parsed in rule %q", route.Rule)
							continue
						}

						log.Debugf("Try to challenge certificate for domain %v founded in Host rule", domains)

						var domain types.Domain
						if len(domains) > 0 {
							domain = types.Domain{Main: domains[0]}
							if len(domains) > 1 {
								domain.SANs = domains[1:]
							}

							safe.Go(func() {
								if _, err := p.resolveCertificate(domain, false); err != nil {
									log.Errorf("Unable to obtain ACME certificate for domains %q detected thanks to rule %q : %v", strings.Join(domains, ","), route.Rule, err)
								}
							})
						}
					}
				}
			case <-stop:
				return
			}
		}
	})
}

// SetCertificateStore allow to initialize certificate store
func (p *Provider) SetCertificateStore(certificateStore *traefiktls.CertificateStore) {
	p.certificateStore = certificateStore
}

func (p *Provider) resolveCertificate(domain types.Domain, domainFromConfigurationFile bool) (*acme.CertificateResource, error) {
	domains, err := p.getValidDomains(domain, domainFromConfigurationFile)
	if err != nil {
		return nil, err
	}

	// Check provided certificates
	uncheckedDomains := p.getUncheckedDomains(domains, !domainFromConfigurationFile)
	if len(uncheckedDomains) == 0 {
		return nil, nil
	}

	log.Debugf("Loading ACME certificates %+v...", uncheckedDomains)
	client, err := p.getClient()
	if err != nil {
		return nil, fmt.Errorf("cannot get ACME client %v", err)
	}

	bundle := true

	certificate, failures := client.ObtainCertificate(uncheckedDomains, bundle, nil, OSCPMustStaple)
	if len(failures) > 0 {
		return nil, fmt.Errorf("cannot obtain certificates %+v", failures)
	}

	if len(certificate.Certificate) == 0 || len(certificate.PrivateKey) == 0 {
		return nil, fmt.Errorf("domains %v generate certificate with no value: %v", uncheckedDomains, certificate)
	}
	log.Debugf("Certificates obtained for domains %+v", uncheckedDomains)

	if len(uncheckedDomains) > 1 {
		domain = types.Domain{Main: uncheckedDomains[0], SANs: uncheckedDomains[1:]}
	} else {
		domain = types.Domain{Main: uncheckedDomains[0]}
	}
	p.addCertificateForDomain(domain, certificate.Certificate, certificate.PrivateKey)

	return &certificate, nil
}

func (p *Provider) getClient() (*acme.Client, error) {
	p.clientMutex.Lock()
	defer p.clientMutex.Unlock()
	var account *Account
	if p.client == nil {
		var err error
		account, err = p.initAccount()
		if err != nil {
			return nil, err
		}

		log.Debug("Building ACME client...")
		caServer := "https://acme-v02.api.letsencrypt.org/directory"
		if len(p.CAServer) > 0 {
			caServer = p.CAServer
		}
		log.Debugf(caServer)
		client, err := acme.NewClient(caServer, account, account.KeyType)
		if err != nil {
			return nil, err
		}
		if account.GetRegistration() == nil {
			// New users will need to register; be sure to save it
			log.Info("Register...")
			reg, err := client.Register(true)
			if err != nil {
				return nil, err
			}
			account.Registration = reg
		}

		// Save the account once before all the certificates generation/storing
		// No certificate can be generated if account is not initialized
		err = p.Store.SaveAccount(account)
		if err != nil {
			return nil, err
		}

		if p.DNSChallenge != nil && len(p.DNSChallenge.Provider) > 0 {
			log.Debugf("Using DNS Challenge provider: %s", p.DNSChallenge.Provider)

			err = dnsOverrideDelay(p.DNSChallenge.DelayBeforeCheck)
			if err != nil {
				return nil, err
			}

			var provider acme.ChallengeProvider
			provider, err = dns.NewDNSChallengeProviderByName(p.DNSChallenge.Provider)
			if err != nil {
				return nil, err
			}

			client.ExcludeChallenges([]acme.Challenge{acme.HTTP01})
			err = client.SetChallengeProvider(acme.DNS01, provider)
			if err != nil {
				return nil, err
			}
		} else if p.HTTPChallenge != nil && len(p.HTTPChallenge.EntryPoint) > 0 {
			log.Debug("Using HTTP Challenge provider.")
			client.ExcludeChallenges([]acme.Challenge{acme.DNS01})
			err = client.SetChallengeProvider(acme.HTTP01, p)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("ACME challenge not specified, please select HTTP or DNS Challenge")
		}
		p.client = client
	}

	return p.client, nil
}

// Present presents a challenge to obtain new ACME certificate
func (p *Provider) Present(domain, token, keyAuth string) error {
	return presentHTTPChallenge(domain, token, keyAuth, p.Store)
}

// CleanUp cleans the challenges when certificate is obtained
func (p *Provider) CleanUp(domain, token, keyAuth string) error {
	return cleanUpHTTPChallenge(domain, token, p.Store)
}

// Provide allows the file provider to provide configurations to traefik
// using the given Configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	p.pool = pool
	err := p.init()
	if err != nil {
		return err
	}

	p.configurationChan = configurationChan
	p.refreshCertificates()

	p.deleteUnnecessaryDomains()
	for i := 0; i < len(p.Domains); i++ {
		domain := p.Domains[i]
		safe.Go(func() {
			if _, err := p.resolveCertificate(domain, true); err != nil {
				log.Errorf("Unable to obtain ACME certificate for domains %q : %v", strings.Join(domain.ToStrArray(), ","), err)
			}
		})
	}

	p.renewCertificates()

	ticker := time.NewTicker(24 * time.Hour)
	pool.Go(func(stop chan bool) {
		for {
			select {
			case <-ticker.C:
				p.renewCertificates()
			case <-stop:
				ticker.Stop()
				return
			}
		}
	})

	return nil
}

func (p *Provider) addCertificateForDomain(domain types.Domain, certificate []byte, key []byte) {
	p.certsChan <- &Certificate{Certificate: certificate, Key: key, Domain: domain}
}

func (p *Provider) watchCertificate() {
	p.certsChan = make(chan *Certificate)
	p.pool.Go(func(stop chan bool) {
		for {
			select {
			case cert := <-p.certsChan:
				certUpdated := false
				for _, domainsCertificate := range p.certificates {
					if reflect.DeepEqual(cert.Domain, domainsCertificate.Domain) {
						domainsCertificate.Certificate = cert.Certificate
						domainsCertificate.Key = cert.Key
						certUpdated = true
						break
					}
				}
				if !certUpdated {
					p.certificates = append(p.certificates, cert)
				}
				p.saveCertificates()

			case <-stop:
				return
			}
		}
	})
}

func (p *Provider) saveCertificates() {
	err := p.Store.SaveCertificates(p.certificates)
	if err != nil {
		log.Error(err)
	}
	p.refreshCertificates()
}

func (p *Provider) refreshCertificates() {
	config := types.ConfigMessage{
		ProviderName: "ACME",
		Configuration: &types.Configuration{
			Backends:  map[string]*types.Backend{},
			Frontends: map[string]*types.Frontend{},
			TLS:       []*traefiktls.Configuration{},
		},
	}

	for _, cert := range p.certificates {
		certificate := &traefiktls.Certificate{CertFile: traefiktls.FileOrContent(cert.Certificate), KeyFile: traefiktls.FileOrContent(cert.Key)}
		config.Configuration.TLS = append(config.Configuration.TLS, &traefiktls.Configuration{Certificate: certificate, EntryPoints: []string{p.EntryPoint}})
	}
	p.configurationChan <- config
}

// Timeout calculates the maximum of time allowed to resolved an ACME challenge
func (p *Provider) Timeout() (timeout, interval time.Duration) {
	return 60 * time.Second, 5 * time.Second
}

func (p *Provider) renewCertificates() {
	log.Info("Testing certificate renew...")
	for _, certificate := range p.certificates {
		crt, err := getX509Certificate(certificate)
		// If there's an error, we assume the cert is broken, and needs update
		// <= 30 days left, renew certificate
		if err != nil || crt == nil || crt.NotAfter.Before(time.Now().Add(24*30*time.Hour)) {
			client, err := p.getClient()
			if err != nil {
				log.Infof("Error renewing certificate from LE : %+v, %v", certificate.Domain, err)
				continue
			}

			log.Infof("Renewing certificate from LE : %+v", certificate.Domain)

			renewedCert, err := client.RenewCertificate(acme.CertificateResource{
				Domain:      certificate.Domain.Main,
				PrivateKey:  certificate.Key,
				Certificate: certificate.Certificate,
			}, true, OSCPMustStaple)

			if err != nil {
				log.Errorf("Error renewing certificate from LE: %v, %v", certificate.Domain, err)
				continue
			}

			if len(renewedCert.Certificate) == 0 || len(renewedCert.PrivateKey) == 0 {
				log.Errorf("domains %v renew certificate with no value: %v", certificate.Domain.ToStrArray(), certificate)
				continue
			}

			p.addCertificateForDomain(certificate.Domain, renewedCert.Certificate, renewedCert.PrivateKey)
		}
	}
}

// AddRoutes add routes on internal router
func (p *Provider) AddRoutes(router *mux.Router) {
	router.Methods(http.MethodGet).
		Path(acme.HTTP01ChallengePath("{token}")).
		Handler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			vars := mux.Vars(req)
			if token, ok := vars["token"]; ok {
				domain, _, err := net.SplitHostPort(req.Host)
				if err != nil {
					log.Debugf("Unable to split host and port: %v. Fallback to request host.", err)
					domain = req.Host
				}

				tokenValue := getTokenValue(token, domain, p.Store)
				if len(tokenValue) > 0 {
					rw.WriteHeader(http.StatusOK)
					_, err = rw.Write(tokenValue)
					if err != nil {
						log.Errorf("Unable to write token : %v", err)
					}
					return
				}
			}
			rw.WriteHeader(http.StatusNotFound)
		}))
}

// Get provided certificate which check a domains list (Main and SANs)
// from static and dynamic provided certificates
func (p *Provider) getUncheckedDomains(domainsToCheck []string, checkConfigurationDomains bool) []string {
	log.Debugf("Looking for provided certificate(s) to validate %q...", domainsToCheck)
	var allDomains []string

	allDomains = p.certificateStore.GetAllDomains()

	// Get ACME certificates
	for _, certificate := range p.certificates {
		allDomains = append(allDomains, strings.Join(certificate.Domain.ToStrArray(), ","))
	}

	// Get Configuration Domains
	if checkConfigurationDomains {
		for i := 0; i < len(p.Domains); i++ {
			allDomains = append(allDomains, strings.Join(p.Domains[i].ToStrArray(), ","))
		}
	}

	return searchUncheckedDomains(domainsToCheck, allDomains)
}

func searchUncheckedDomains(domainsToCheck []string, existentDomains []string) []string {
	var uncheckedDomains []string
	for _, domainToCheck := range domainsToCheck {
		if !isDomainAlreadyChecked(domainToCheck, existentDomains) {
			uncheckedDomains = append(uncheckedDomains, domainToCheck)
		}
	}
	if len(uncheckedDomains) == 0 {
		log.Debugf("No ACME certificate to generate for domains %q.", domainsToCheck)
	} else {
		log.Debugf("Domains %q need ACME certificates generation for domains %q.", domainsToCheck, strings.Join(uncheckedDomains, ","))
	}
	return uncheckedDomains
}

func getX509Certificate(certificate *Certificate) (*x509.Certificate, error) {
	var crt *x509.Certificate
	tlsCert, err := tls.X509KeyPair(certificate.Certificate, certificate.Key)
	if err != nil {
		log.Errorf("Failed to load TLS keypair from ACME certificate for domain %q (SAN : %q), certificate will be renewed : %v", certificate.Domain.Main, strings.Join(certificate.Domain.SANs, ","), err)
		return nil, err
	}
	crt = tlsCert.Leaf
	if crt == nil {
		crt, err = x509.ParseCertificate(tlsCert.Certificate[0])
		if err != nil {
			log.Errorf("Failed to parse TLS keypair from ACME certificate for domain %q (SAN : %q), certificate will be renewed : %v", certificate.Domain.Main, strings.Join(certificate.Domain.SANs, ","), err)
		}
	}
	return crt, err
}

// getValidDomains checks if given domain is allowed to generate a ACME certificate and return it
func (p *Provider) getValidDomains(domain types.Domain, wildcardAllowed bool) ([]string, error) {
	domains := domain.ToStrArray()
	if len(domains) == 0 {
		return nil, errors.New("unable to generate a certificate in ACME provider when no domain is given")
	}
	if strings.HasPrefix(domain.Main, "*") {
		if !wildcardAllowed {
			return nil, fmt.Errorf("unable to generate a wildcard certificate in ACME provider for domain %q from a 'Host' rule", strings.Join(domains, ","))
		}
		if p.DNSChallenge == nil {
			return nil, fmt.Errorf("unable to generate a wildcard certificate in ACME provider for domain %q : ACME needs a DNSChallenge", strings.Join(domains, ","))
		}
		if strings.HasPrefix(domain.Main, "*.*") {
			return nil, fmt.Errorf("unable to generate a wildcard certificate in ACME provider for domain %q : ACME does not allow '*.*' wildcard domain", strings.Join(domains, ","))
		}
	}
	for _, san := range domain.SANs {
		if strings.HasPrefix(san, "*") {
			return nil, fmt.Errorf("unable to generate a certificate in ACME provider for domains %q: SAN %q can not be a wildcard domain", strings.Join(domains, ","), san)
		}
	}

	domains = fun.Map(types.CanonicalDomain, domains).([]string)
	return domains, nil
}

func isDomainAlreadyChecked(domainToCheck string, existentDomains []string) bool {
	for _, certDomains := range existentDomains {
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
func (p *Provider) deleteUnnecessaryDomains() {
	var newDomains []types.Domain

	for idxDomainToCheck, domainToCheck := range p.Domains {
		keepDomain := true

		for idxDomain, domain := range p.Domains {
			if idxDomainToCheck == idxDomain {
				continue
			}

			if reflect.DeepEqual(domain, domainToCheck) {
				if idxDomainToCheck > idxDomain {
					log.Warnf("The domain %v is duplicated in the configuration but will be process by ACME provider only once.", domainToCheck)
					keepDomain = false
				}
				break
			}

			// Check if CN or SANS to check already exists
			// or can not be checked by a wildcard
			var newDomainsToCheck []string
			for _, domainProcessed := range domainToCheck.ToStrArray() {
				if idxDomain < idxDomainToCheck && isDomainAlreadyChecked(domainProcessed, domain.ToStrArray()) {
					// The domain is duplicated in a CN
					log.Warnf("Domain %q is duplicated in the configuration or validated by the domain %v. It will be processed once.", domainProcessed, domain)
					continue
				} else if domain.Main != domainProcessed && strings.HasPrefix(domain.Main, "*") && isDomainAlreadyChecked(domainProcessed, []string{domain.Main}) {
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

	p.Domains = newDomains
}
