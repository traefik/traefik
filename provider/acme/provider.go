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
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/flaeg"
	"github.com/containous/mux"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/rules"
	"github.com/containous/traefik/safe"
	traefikTLS "github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/pkg/errors"
	acme "github.com/xenolf/lego/acmev2"
	"github.com/xenolf/lego/providers/dns"
)

var (
	// OSCPMustStaple enables OSCP stapling as from https://github.com/xenolf/lego/issues/270
	OSCPMustStaple = false
	provider       = &Provider{}
)

// Configuration holds ACME configuration provided by users
type Configuration struct {
	Email         string         `description:"Email address used for registration"`
	ACMELogging   bool           `description:"Enable debug logging of ACME actions."`
	CAServer      string         `description:"CA server to use."`
	Storage       string         `description:"Storage to use."`
	EntryPoint    string         `description:"EntryPoint to use."`
	OnHostRule    bool           `description:"Enable certificate generation on frontends Host rules."`
	OnDemand      bool           `description:"Enable on demand certificate generation. This will request a certificate from Let's Encrypt during the first TLS handshake for a hostname that does not yet have a certificate."` //deprecated
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
	dynamicCerts           *safe.Safe
	staticCerts            map[string]*tls.Certificate
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

// Get returns the provider instance
func Get() *Provider {
	return provider
}

// IsEnabled returns true if the provider instance and its configuration are not nil, otherwise false
func IsEnabled() bool {
	return provider != nil && provider.Configuration != nil
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
		return fmt.Errorf("unable to get ACME account : %v", err)
	}

	p.watchCertificate()
	p.watchNewDomains()

	return nil
}

func (p *Provider) initAccount() (*Account, error) {
	if p.account == nil || len(p.account.Email) == 0 {
		var err error
		p.account, err = NewAccount(p.Email)
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

// SetDynamicCertificates allow to initialize dynamicCerts map
func (p *Provider) SetDynamicCertificates(safe *safe.Safe) {
	p.dynamicCerts = safe
}

// SetStaticCertificates allow to initialize staticCerts map
func (p *Provider) SetStaticCertificates(staticCerts map[string]*tls.Certificate) {
	p.staticCerts = staticCerts
}

func (p *Provider) resolveCertificate(domain types.Domain, wildcardAllowed bool) (*acme.CertificateResource, error) {
	domains, err := p.getValidDomains(domain, wildcardAllowed)
	if err != nil {
		return nil, err
	}

	// Check provided certificates
	uncheckedDomains := p.getUncheckedDomains(domains)
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
	log.Debugf("Certificates obtained for domain %+v", uncheckedDomains)
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
		client, err := acme.NewClient(caServer, account, acme.RSA4096)
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

	for _, domain := range p.Domains {
		safe.Go(func() {
			if _, err := p.resolveCertificate(domain, true); err != nil {
				domains := []string{domain.Main}
				domains = append(domains, domain.SANs...)
				log.Errorf("Unable to obtain ACME certificate for domains %q : %v", domains, err)
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

func (p *Provider) deleteCertificateForDomain(domain types.Domain) {
	for k, cert := range p.certificates {
		if reflect.DeepEqual(cert.Domain, domain) {
			p.certificates = append(p.certificates[:k], p.certificates[k+1:]...)
		}
	}
	p.saveCertificates()
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
			TLS:       []*traefikTLS.Configuration{},
		},
	}

	for _, cert := range p.certificates {
		certificate := &traefikTLS.Certificate{CertFile: traefikTLS.FileOrContent(cert.Certificate), KeyFile: traefikTLS.FileOrContent(cert.Key)}
		config.Configuration.TLS = append(config.Configuration.TLS, &traefikTLS.Configuration{Certificate: certificate})
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
func (p *Provider) getUncheckedDomains(domainsToCheck []string) []string {
	log.Debugf("Looking for provided certificate(s) to validate %q...", domainsToCheck)
	var allCerts []string

	// Get static certificates
	for domains := range p.staticCerts {
		allCerts = append(allCerts, domains)
	}

	// Get dynamic certificates
	if p.dynamicCerts != nil && p.dynamicCerts.Get() != nil {
		for domains := range p.dynamicCerts.Get().(map[string]*tls.Certificate) {
			allCerts = append(allCerts, domains)
		}
	}

	// Get ACME certificates
	for _, certificate := range p.certificates {
		var domains []string
		if len(certificate.Domain.Main) > 0 {
			domains = []string{certificate.Domain.Main}
		}
		domains = append(domains, certificate.Domain.SANs...)
		allCerts = append(allCerts, strings.Join(domains, ","))
	}

	return searchUncheckedDomains(domainsToCheck, allCerts)
}

func searchUncheckedDomains(domains []string, certs []string) []string {
	uncheckedDomains := []string{}
	for _, domainToCheck := range domains {
		domainCheck := false
		for _, certDomains := range certs {
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
	var domains []string
	if len(domain.Main) > 0 {
		domains = []string{domain.Main}
	}
	domains = append(domains, domain.SANs...)
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
		if len(domain.SANs) > 0 {
			return nil, fmt.Errorf("unable to generate a wildcard certificate in ACME provider for domain %q : SANs are not allowed", strings.Join(domains, ","))
		}
	}
	domains = fun.Map(types.CanonicalDomain, domains).([]string)
	return domains, nil
}
