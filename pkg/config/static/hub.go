package static

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider/hub"
)

func (c *Configuration) initHubProvider() error {
	if _, ok := c.EntryPoints[hub.TunnelEntrypoint]; !ok {
		var ep EntryPoint
		ep.SetDefaults()
		ep.Address = ":9901"
		c.EntryPoints[hub.TunnelEntrypoint] = &ep

		log.Info().Str(logs.EntryPointName, hub.TunnelEntrypoint).
			Msg("The entryPoint is created on port 9901 to allow exposition of services.")
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
		log.Warn().Msg("Hub is in `insecure` mode. Do not run in production with this setup.")
	}

	if _, ok := c.EntryPoints[hub.APIEntrypoint]; !ok {
		var ep EntryPoint
		ep.SetDefaults()
		ep.Address = ":9900"
		c.EntryPoints[hub.APIEntrypoint] = &ep

		log.Info().Str(logs.EntryPointName, hub.APIEntrypoint).
			Msg("The entryPoint is created on port 9900 to allow Traefik to communicate with the Hub Agent for Traefik.")
	}

	c.EntryPoints[hub.APIEntrypoint].HTTP.TLS = &TLSConfig{
		Options: "traefik-hub",
	}

	return nil
}
