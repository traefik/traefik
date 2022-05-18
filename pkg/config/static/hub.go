package static

import (
	"errors"

	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider/hub"
)

func (c *Configuration) initHubProvider() error {
	// Hub provider is an experimental feature. It requires the experimental flag to be enabled before continuing.
	if c.Experimental == nil || !c.Experimental.Hub {
		return errors.New("the experimental flag for Hub is not set")
	}

	if _, ok := c.EntryPoints[hub.TunnelEntrypoint]; !ok {
		var ep EntryPoint
		ep.SetDefaults()
		ep.Address = ":9901"
		c.EntryPoints[hub.TunnelEntrypoint] = &ep
		log.WithoutContext().Infof("The entryPoint %q is created on port 9901 to allow exposition of services.", hub.TunnelEntrypoint)
	}

	if c.Hub.TLS == nil {
		return nil
	}

	if c.Hub.TLS.Insecure && (c.Hub.TLS.CA != "" || c.Hub.TLS.Cert != "" || c.Hub.TLS.Key != "") {
		return errors.New("mTLS configuration for Hub and insecure TLS for Hub are mutually exclusive")
	}

	if !c.Hub.TLS.Insecure && (c.Hub.TLS.CA == "" || c.Hub.TLS.Cert == "" || c.Hub.TLS.Key == "") {
		return errors.New("incomplete mTLS configuration for Hub")
	}

	if c.Hub.TLS.Insecure {
		log.WithoutContext().Warn("Hub is in `insecure` mode. Do not run in production with this setup.")
	}

	if _, ok := c.EntryPoints[hub.APIEntrypoint]; !ok {
		var ep EntryPoint
		ep.SetDefaults()
		ep.Address = ":9900"
		c.EntryPoints[hub.APIEntrypoint] = &ep
		log.WithoutContext().Infof("The entryPoint %q is created on port 9900 to allow Traefik to communicate with the Hub Agent for Traefik.", hub.APIEntrypoint)
	}

	c.EntryPoints[hub.APIEntrypoint].HTTP.TLS = &TLSConfig{
		Options: "traefik-hub",
	}

	return nil
}
