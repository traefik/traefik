package vault

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/rules"
	"github.com/containous/traefik/pkg/safe"
	traefiktls "github.com/containous/traefik/pkg/tls"
	"github.com/containous/traefik/pkg/types"
	vaultapi "github.com/hashicorp/vault/api"
)

// Configuration holds Vault configuration provided by users
type Configuration struct {
	VaultServer string `description:"Vault server to use."`
	Storage     string `description:"Storage to use."` // TODO: we should adopt our storage format to the ACME one
	EntryPoint  string `description:"EntryPoint to use."`
	// TODO: we should generate the private key ourselfs, but for now we use the vault role, and let vault
	// do the magic for us
	// KeyType      string `description:"KeyType used for generating certificate private key. Allow value 'EC256', 'EC384', 'RSA2048', 'RSA4096', 'RSA8192'. Default to 'RSA4096'"`
	OnHostRule bool   `description:"Enable certificate generation on frontends Host rules."`
	PkiPath    string `description:"Path for the vault pki endpoint"`
	Role       string `description:"Vault PKI role to request certifcate"`
	ttl        int    `description:"TTL for the certficate in hours"`
}

// Certificate is a struct which contains all data needed from an Vault certificate
type Certificate struct {
	Domain      types.Domain
	Certificate string
	Key         string
}

//

// Provider holds configurations of the provider.
type Provider struct {
	*Configuration
	Store                  Store
	certificates           []*Certificate
	client                 *vaultapi.Client
	certsChan              chan *Certificate
	configurationChan      chan<- config.Message
	tlsManager             *traefiktls.Manager
	clientMutex            sync.Mutex
	configFromListenerChan chan config.Configuration
	pool                   *safe.Pool
	resolvingDomains       map[string]struct{}
	resolvingDomainsMutex  sync.RWMutex
}

// SetTLSManager sets the tls manager to use
func (p *Provider) SetTLSManager(tlsManager *traefiktls.Manager) {
	p.tlsManager = tlsManager
}

// SetConfigListenerChan initializes the configFromListenerChan
func (p *Provider) SetConfigListenerChan(configFromListenerChan chan config.Configuration) {
	p.configFromListenerChan = configFromListenerChan
}

// ListenConfiguration sets a new Configuration into the configFromListenerChan
func (p *Provider) ListenConfiguration(config config.Configuration) {
	p.configFromListenerChan <- config
}

// Init the Vault provider
func (p *Provider) Init() error {
	if len(p.Configuration.Storage) == 0 {
		return errors.New("unable to initialize ACME provider with no storage location for the certificates")
	}
	p.Store = NewVaultStore(p.Configuration.Storage)

	var err error

	p.certificates, err = p.Store.GetCertificates()
	if err != nil {
		return fmt.Errorf("unable to get ACME certificates : %v", err)
	}

	// Init the currently resolved domain map
	p.resolvingDomains = make(map[string]struct{})

	return nil
}

// Provide allows the vault provider to provide configurations to traefik
// using the given Configuration channel.
func (p *Provider) Provide(configurationChan chan<- config.Message, pool *safe.Pool) error {
	ctx := log.With(context.Background(), log.Str(log.ProviderName, "Vault"))

	p.pool = pool

	p.watchCertificate(ctx)
	p.watchNewDomains(ctx)

	p.configurationChan = configurationChan
	p.refreshCertificates()

	p.renewCertificates(ctx)

	// we check ever hour for expired certficates
	// TODO: we should adjust this according to the TTL value
	ticker := time.NewTicker(time.Hour)
	pool.Go(func(stop chan bool) {
		for {
			select {
			case <-ticker.C:
				p.renewCertificates(ctx)
			case <-stop:
				ticker.Stop()
				return
			}
		}
	})

	return nil
}

// get the vault client
func (p *Provider) getClient() (*vaultapi.Client, error) {
	p.clientMutex.Lock()
	defer p.clientMutex.Unlock()

	ctx := log.With(context.Background(), log.Str(log.ProviderName, "vault"))
	logger := log.FromContext(ctx)

	if p.client != nil {
		return p.client, nil
	}

	logger.Debug("Building Vault client...")

	vaultConfig := vaultapi.DefaultConfig()
	// TODO: allow the user to configure all paremters within the config file
	// instead of the environment

	client, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}

	p.client = client
	return p.client, nil
}

func (p *Provider) resolveDomains(ctx context.Context, domains []string) {
	if len(domains) == 0 {
		log.FromContext(ctx).Debug("No domain parsed in provider Vault")
		return
	}

	log.FromContext(ctx).Debugf("Try to challenge certificate for domain %v founded in HostSNI rule", domains)

	var domain types.Domain
	if len(domains) > 0 {
		domain = types.Domain{Main: domains[0]}
		if len(domains) > 1 {
			domain.SANs = domains[1:]
		}

		safe.Go(func() {
			if _, err := p.resolveCertificate(ctx, domain, false); err != nil {
				log.FromContext(ctx).Errorf("Unable to obtain Vault certificate for domains %q: %v", strings.Join(domains, ","), err)
			}
		})
	}
}

func (p *Provider) watchNewDomains(ctx context.Context) {
	p.pool.Go(func(stop chan bool) {
		for {
			select {
			case config := <-p.configFromListenerChan:
				if config.TCP != nil {
					for routerName, route := range config.TCP.Routers {
						ctxRouter := log.With(ctx, log.Str(log.RouterName, routerName), log.Str(log.Rule, route.Rule))

						domains, err := rules.ParseHostSNI(route.Rule)
						if err != nil {
							log.FromContext(ctxRouter).Errorf("Error parsing domains in provider Vault: %v", err)
							continue
						}
						p.resolveDomains(ctxRouter, domains)
					}
				}

				for routerName, route := range config.HTTP.Routers {
					ctxRouter := log.With(ctx, log.Str(log.RouterName, routerName), log.Str(log.Rule, route.Rule))

					domains, err := rules.ParseDomains(route.Rule)
					if err != nil {
						log.FromContext(ctxRouter).Errorf("Error parsing domains in provider Vault: %v", err)
						continue
					}
					p.resolveDomains(ctxRouter, domains)
				}
			case <-stop:
				return
			}
		}
	})
}

func (p *Provider) resolveCertificate(ctx context.Context, domain types.Domain, domainFromConfigurationFile bool) (*vaultapi.Secret, error) {
	domains, err := p.getValidDomains(ctx, domain, domainFromConfigurationFile)
	if err != nil {
		return nil, err
	}

	// Check provided certificates
	uncheckedDomains := p.getUncheckedDomains(ctx, domains, !domainFromConfigurationFile)
	if len(uncheckedDomains) == 0 {
		return nil, nil
	}

	p.addResolvingDomains(uncheckedDomains)
	defer p.removeResolvingDomains(uncheckedDomains)

	logger := log.FromContext(ctx)
	logger.Debugf("Loading Vault certificates %+v...", uncheckedDomains)

	client, err := p.getClient()
	if err != nil {
		return nil, fmt.Errorf("cannot get Vault client %v", err)
	}

	secret, err := obtainCertificate(ctx, uncheckedDomains, p.Role, p.PkiPath, p.ttl, client)

	cert := secret.Data["certificate"].(string)
	key := secret.Data["private_key"].(string)
	if len(cert) == 0 || len(key) == 0 {
		return nil, fmt.Errorf("domains %v generate certificate with no value: %v", uncheckedDomains, secret)
	}

	logger.Debugf("Certificates obtained for domains %+v", uncheckedDomains)

	if len(uncheckedDomains) > 1 {
		domain = types.Domain{Main: uncheckedDomains[0], SANs: uncheckedDomains[1:]}
	} else {
		domain = types.Domain{Main: uncheckedDomains[0]}
	}
	p.addCertificateForDomain(domain, cert, key)

	return secret, nil
}

func (p *Provider) removeResolvingDomains(resolvingDomains []string) {
	p.resolvingDomainsMutex.Lock()
	defer p.resolvingDomainsMutex.Unlock()

	for _, domain := range resolvingDomains {
		delete(p.resolvingDomains, domain)
	}
}

func (p *Provider) addResolvingDomains(resolvingDomains []string) {
	p.resolvingDomainsMutex.Lock()
	defer p.resolvingDomainsMutex.Unlock()

	for _, domain := range resolvingDomains {
		p.resolvingDomains[domain] = struct{}{}
	}
}

func (p *Provider) addCertificateForDomain(domain types.Domain, certificate string, key string) {
	//
	p.certsChan <- &Certificate{Certificate: certificate, Key: key, Domain: domain}
}

func (p *Provider) renewCertificates(ctx context.Context) {
	logger := log.FromContext(ctx)

	logger.Info("Testing certificate renew...")
	for _, cert := range p.certificates {
		crt, err := getX509Certificate(ctx, cert)
		// If there's an error, we assume the cert is broken, and needs update
		// <= 2 hours left, renew certificate
		if err != nil || crt == nil || crt.NotAfter.Before(time.Now().Add(2*time.Hour)) {
			client, err := p.getClient()
			if err != nil {
				logger.Infof("Error renewing certificate from Vault : %+v, %v", cert.Domain, err)
				continue
			}

			logger.Infof("Renewing certificate from Vault : %+v", cert.Domain)

			domains := []string{cert.Domain.Main}
			domains = append(domains, cert.Domain.SANs...)

			secret, err := obtainCertificate(ctx, domains, p.Role, p.PkiPath, p.ttl, client)

			if err != nil {
				logger.Errorf("Error renewing certificate from Vault: %v, %v", cert.Domain, err)
				continue
			}

			newCert := secret.Data["certificate"].(string)
			newKey := secret.Data["private_key"].(string)
			if len(newCert) == 0 || len(newKey) == 0 {
				logger.Errorf("domains %v generate certificate with no value: %v", domains, secret)
				continue
			}

			p.addCertificateForDomain(cert.Domain, newCert, newKey)
		}
	}
}

// Get provided certificate which check a domains list (Main and SANs)
// from dynamic provided certificates
func (p *Provider) getUncheckedDomains(ctx context.Context, domainsToCheck []string, checkConfigurationDomains bool) []string {
	p.resolvingDomainsMutex.RLock()
	defer p.resolvingDomainsMutex.RUnlock()

	log.FromContext(ctx).Debugf("Looking for provided certificate(s) to validate %q...", domainsToCheck)

	allDomains := p.tlsManager.GetStore("default").GetAllDomains()

	// Get Vault certificates
	for _, cert := range p.certificates {
		allDomains = append(allDomains, strings.Join(cert.Domain.ToStrArray(), ","))
	}

	// Get currently resolved domains
	for domain := range p.resolvingDomains {
		allDomains = append(allDomains, domain)
	}

	return searchUncheckedDomains(ctx, domainsToCheck, allDomains)
}

func searchUncheckedDomains(ctx context.Context, domainsToCheck []string, existentDomains []string) []string {
	var uncheckedDomains []string
	for _, domainToCheck := range domainsToCheck {
		if !isDomainAlreadyChecked(domainToCheck, existentDomains) {
			uncheckedDomains = append(uncheckedDomains, domainToCheck)
		}
	}

	logger := log.FromContext(ctx)
	if len(uncheckedDomains) == 0 {
		logger.Debugf("No Vault certificate generation required for domains %q.", domainsToCheck)
	} else {
		logger.Debugf("Domains %q need Vault certificates generation for domains %q.", domainsToCheck, strings.Join(uncheckedDomains, ","))
	}
	return uncheckedDomains
}

func getX509Certificate(ctx context.Context, cert *Certificate) (*x509.Certificate, error) {
	logger := log.FromContext(ctx)

	tlsCert, err := tls.X509KeyPair([]byte(cert.Certificate), []byte(cert.Key))
	if err != nil {
		logger.Errorf("Failed to load TLS key pair from Vault certificate for domain %q (SAN : %q), certificate will be renewed : %v", cert.Domain.Main, strings.Join(cert.Domain.SANs, ","), err)
		return nil, err
	}

	crt := tlsCert.Leaf
	if crt == nil {
		crt, err = x509.ParseCertificate(tlsCert.Certificate[0])
		if err != nil {
			logger.Errorf("Failed to parse TLS key pair from Vault certificate for domain %q (SAN : %q), certificate will be renewed : %v", cert.Domain.Main, strings.Join(cert.Domain.SANs, ","), err)
		}
	}

	return crt, err
}

// getValidDomains checks if given domain is allowed to generate a Vault certificate and return it
func (p *Provider) getValidDomains(ctx context.Context, domain types.Domain, wildcardAllowed bool) ([]string, error) {
	domains := domain.ToStrArray()
	if len(domains) == 0 {
		return nil, errors.New("unable to generate a certificate in Vault provider when no domain is given")
	}

	if strings.HasPrefix(domain.Main, "*") {
		if !wildcardAllowed {
			return nil, fmt.Errorf("unable to generate a wildcard certificate in Vault provider for domain %q from a 'Host' rule", strings.Join(domains, ","))
		}
	}

	var canonicalDomains []string
	for _, domain := range domains {
		canonicalDomain := types.CanonicalDomain(domain)
		canonicalDomains = append(canonicalDomains, canonicalDomain)
	}

	return canonicalDomains, nil
}

func (p *Provider) watchCertificate(ctx context.Context) {
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

				err := p.saveCertificates()
				if err != nil {
					log.FromContext(ctx).Error(err)
				}
			case <-stop:
				return
			}
		}
	})
}

func (p *Provider) saveCertificates() error {
	err := p.Store.SaveCertificates(p.certificates)

	p.refreshCertificates()

	return err
}

func (p *Provider) refreshCertificates() {
	conf := config.Message{
		ProviderName: "Vault",
		Configuration: &config.Configuration{
			HTTP: &config.HTTPConfiguration{
				Routers:     map[string]*config.Router{},
				Middlewares: map[string]*config.Middleware{},
				Services:    map[string]*config.Service{},
			},
			TLS: []*traefiktls.Configuration{},
		},
	}

	for _, cert := range p.certificates {
		cert := &traefiktls.Certificate{CertFile: traefiktls.FileOrContent(cert.Certificate), KeyFile: traefiktls.FileOrContent(cert.Key)}
		conf.Configuration.TLS = append(conf.Configuration.TLS, &traefiktls.Configuration{Certificate: cert})
	}
	p.configurationChan <- conf
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

func obtainCertificate(ctx context.Context, domains []string, role string, pkiPath string, ttl int, client *vaultapi.Client) (*vaultapi.Secret, error) {
	logger := log.FromContext(ctx)

	var err error

	var secret *vaultapi.Secret
	// request the cert from vault
	request := map[string]interface{}{
		"name":        role,
		"common_name": domains[0],
		"format":      "pem_bundle",
		"ttl":         fmt.Sprintf("%dh", ttl),
	}
	if len(domains) > 1 {
		request["alt_names"] = domains[1:]
	}

	secret, err = client.Logical().Write(pkiPath, request)

	if err != nil {
		logger.Errorf("Error obtaining certificate: %v", err)
		return nil, err
	}

	return secret, nil
}

func (p *Provider) getCertificate() {

}
