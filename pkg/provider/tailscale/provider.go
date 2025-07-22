package tailscale

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tailscale/tscert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/muxer/http"
	"github.com/traefik/traefik/v3/pkg/muxer/tcp"
	"github.com/traefik/traefik/v3/pkg/safe"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

// Provider is the Tailscale certificates provider implementation. It receives
// configuration updates (e.g. new router, with new domain) from Traefik core,
// fetches the corresponding TLS certificates from the Tailscale daemon, and
// sends back to Traefik core a configuration updated with the certificates.
type Provider struct {
	ResolverName string

	dynConfigs  chan dynamic.Configuration // updates from Traefik core
	dynMessages chan<- dynamic.Message     // update to Traefik core

	certByDomainMu sync.RWMutex
	certByDomain   map[string]traefiktls.Certificate
}

// ThrottleDuration implements the aggregator.throttled interface, in order to
// ensure that this provider is unthrottled.
func (p *Provider) ThrottleDuration() time.Duration {
	return 0
}

// Init implements the provider.Provider interface.
func (p *Provider) Init() error {
	p.dynConfigs = make(chan dynamic.Configuration)
	p.certByDomain = make(map[string]traefiktls.Certificate)

	return nil
}

// HandleConfigUpdate hands out a configuration update to the provider.
func (p *Provider) HandleConfigUpdate(cfg dynamic.Configuration) {
	p.dynConfigs <- cfg
}

// Provide starts the provider, which will henceforth send configuration
// updates on dynMessages.
func (p *Provider) Provide(dynMessages chan<- dynamic.Message, pool *safe.Pool) error {
	p.dynMessages = dynMessages

	logger := log.With().Str(logs.ProviderName, p.ResolverName+".tailscale").Logger()

	pool.GoCtx(func(ctx context.Context) {
		p.watchDomains(logger.WithContext(ctx))
	})

	pool.GoCtx(func(ctx context.Context) {
		p.renewCertificates(logger.WithContext(ctx))
	})

	return nil
}

// watchDomains watches for Tailscale domain certificates that should be fetched from the Tailscale daemon.
func (p *Provider) watchDomains(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case cfg := <-p.dynConfigs:
			domains := p.findDomains(ctx, cfg)
			newDomains := p.findNewDomains(domains)
			purged := p.purgeUnusedCerts(domains)

			if len(newDomains) == 0 && !purged {
				continue
			}

			// TODO: what should we do if the fetched certificate is going to expire before the next refresh tick?
			p.fetchCerts(ctx, newDomains)
			p.sendDynamicConfig()
		}
	}
}

// renewCertificates routinely renews previously resolved Tailscale
// certificates before they expire.
func (p *Provider) renewCertificates(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			p.certByDomainMu.RLock()
			var domainsToRenew []string
			for domain, cert := range p.certByDomain {
				tlsCert, err := cert.GetCertificateFromBytes()
				if err != nil {
					log.Ctx(ctx).
						Err(err).
						Msgf("Unable to get certificate for domain %s", domain)
					continue
				}

				// Tailscale tries to renew certificates 14 days before its expiration date.
				// See https://github.com/tailscale/tailscale/blob/d9efbd97cbf369151e31453749f6692df7413709/ipn/localapi/cert.go#L116
				if isValidCert(tlsCert, domain, time.Now().AddDate(0, 0, 14)) {
					continue
				}

				domainsToRenew = append(domainsToRenew, domain)
			}
			p.certByDomainMu.RUnlock()

			if len(domainsToRenew) == 0 {
				continue
			}

			p.fetchCerts(ctx, domainsToRenew)
			p.sendDynamicConfig()
		}
	}
}

// findDomains goes through the given dynamic.Configuration and returns all
// Tailscale-specific domains found.
func (p *Provider) findDomains(ctx context.Context, cfg dynamic.Configuration) []string {
	logger := log.Ctx(ctx)

	var domains []string

	if cfg.HTTP != nil {
		for _, router := range cfg.HTTP.Routers {
			if router.TLS == nil || router.TLS.CertResolver != p.ResolverName {
				continue
			}

			// As a domain list is explicitly defined we are only using the
			// configured domains. Only the Main domain is considered as
			// Tailscale domain certificate does not support multiple SANs.
			if len(router.TLS.Domains) > 0 {
				for _, domain := range router.TLS.Domains {
					domains = append(domains, domain.Main)
				}

				continue
			}

			parsedDomains, err := http.ParseDomains(router.Rule)
			if err != nil {
				logger.Error().Err(err).Msg("Unable to parse HTTP router domains")
				continue
			}

			domains = append(domains, parsedDomains...)
		}
	}

	if cfg.TCP != nil {
		for _, router := range cfg.TCP.Routers {
			if router.TLS == nil || router.TLS.CertResolver != p.ResolverName {
				continue
			}

			// As a domain list is explicitly defined we are only using the
			// configured domains. Only the Main domain is considered as
			// Tailscale domain certificate does not support multiple SANs.
			if len(router.TLS.Domains) > 0 {
				for _, domain := range router.TLS.Domains {
					domains = append(domains, domain.Main)
				}

				continue
			}

			parsedDomains, err := tcp.ParseHostSNI(router.Rule)
			if err != nil {
				logger.Error().Err(err).Msg("Unable to parse TCP router domains")
				continue
			}

			domains = append(domains, parsedDomains...)
		}
	}

	return sanitizeDomains(ctx, domains)
}

// findNewDomains returns the domains that have not already been fetched from
// the Tailscale daemon.
func (p *Provider) findNewDomains(domains []string) []string {
	p.certByDomainMu.RLock()
	defer p.certByDomainMu.RUnlock()

	var newDomains []string
	for _, domain := range domains {
		if _, ok := p.certByDomain[domain]; ok {
			continue
		}

		newDomains = append(newDomains, domain)
	}

	return newDomains
}

// purgeUnusedCerts purges the certByDomain map by removing unused certificates
// and returns whether some certificates have been removed.
func (p *Provider) purgeUnusedCerts(domains []string) bool {
	p.certByDomainMu.Lock()
	defer p.certByDomainMu.Unlock()

	newCertByDomain := make(map[string]traefiktls.Certificate)
	for _, domain := range domains {
		if cert, ok := p.certByDomain[domain]; ok {
			newCertByDomain[domain] = cert
		}
	}

	purged := len(p.certByDomain) > len(newCertByDomain)

	p.certByDomain = newCertByDomain

	return purged
}

// fetchCerts fetches the certificates for the provided domains from the
// Tailscale daemon.
func (p *Provider) fetchCerts(ctx context.Context, domains []string) {
	logger := log.Ctx(ctx)

	for _, domain := range domains {
		cert, key, err := tscert.CertPair(ctx, domain)
		if err != nil {
			logger.Error().Err(err).Msgf("Unable to fetch certificate for domain %q", domain)
			continue
		}

		logger.Debug().Msgf("Fetched certificate for domain %q", domain)

		p.certByDomainMu.Lock()
		p.certByDomain[domain] = traefiktls.Certificate{
			CertFile: types.FileOrContent(cert),
			KeyFile:  types.FileOrContent(key),
		}
		p.certByDomainMu.Unlock()
	}
}

// sendDynamicConfig sends a dynamic.Message with the dynamic.Configuration
// containing the newly generated (or renewed) Tailscale certs.
func (p *Provider) sendDynamicConfig() {
	p.certByDomainMu.RLock()
	defer p.certByDomainMu.RUnlock()

	// TODO: we always send back to traefik core the set of certificates
	// sorted, to make sure that two identical sets, that would be sorted
	// differently, do not trigger another configuration update because of the
	// mismatch. But in reality we should not end up sending a certificates
	// update if there was no new certs to generate or renew in the first
	// place, so this scenario should never happen, and the sorting might
	// actually not be needed.
	var sortedDomains []string
	for domain := range p.certByDomain {
		sortedDomains = append(sortedDomains, domain)
	}
	sort.Strings(sortedDomains)

	var certs []*traefiktls.CertAndStores
	for _, domain := range sortedDomains {
		// Only the default store is supported.
		certs = append(certs, &traefiktls.CertAndStores{
			Stores:      []string{traefiktls.DefaultTLSStoreName},
			Certificate: p.certByDomain[domain],
		})
	}

	p.dynMessages <- dynamic.Message{
		ProviderName: p.ResolverName + ".tailscale",
		Configuration: &dynamic.Configuration{
			TLS: &dynamic.TLSConfiguration{Certificates: certs},
		},
	}
}

// sanitizeDomains removes duplicated and invalid Tailscale subdomains, from
// the provided list.
func sanitizeDomains(ctx context.Context, domains []string) []string {
	logger := log.Ctx(ctx)

	seen := map[string]struct{}{}

	var sanitizedDomains []string
	for _, domain := range domains {
		if _, ok := seen[domain]; ok {
			continue
		}

		if !isTailscaleDomain(domain) {
			logger.Error().Msgf("Domain %s is not a valid Tailscale domain", domain)
			continue
		}

		sanitizedDomains = append(sanitizedDomains, domain)
		seen[domain] = struct{}{}
	}
	return sanitizedDomains
}

// isTailscaleDomain returns whether the given domain is a valid Tailscale
// domain. A valid Tailscale domain has the following form:
// machine-name.domains-alias.ts.net.
func isTailscaleDomain(domain string) bool {
	// TODO: extra check, against the actual list of allowed domains names,
	// provided by the Tailscale daemon status?
	labels := strings.Split(domain, ".")

	return len(labels) == 4 && labels[2] == "ts" && labels[3] == "net"
}

// isValidCert returns whether the given tls.Certificate is valid for the given
// domain at the given time.
func isValidCert(cert tls.Certificate, domain string, now time.Time) bool {
	var leaf *x509.Certificate

	intermediates := x509.NewCertPool()
	for i, raw := range cert.Certificate {
		der, err := x509.ParseCertificate(raw)
		if err != nil {
			return false
		}

		if i == 0 {
			leaf = der
			continue
		}

		intermediates.AddCert(der)
	}

	if leaf == nil {
		return false
	}

	_, err := leaf.Verify(x509.VerifyOptions{
		DNSName:       domain,
		Intermediates: intermediates,
		CurrentTime:   now,
	})

	return err == nil
}
