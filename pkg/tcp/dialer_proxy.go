package tcp

import (
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"golang.org/x/net/proxy"
	"net/url"
)

// serversDialer is a map from servername to a function that creates a Dialer
// from a URL with such a scheme.
var dialerDecorators []DialerDecorator

// RegisterDialerDecorator register dialer extension for servername to add some customized feature like ratelimit etc.
func RegisterDialerDecorator(dialer DialerDecorator) {
	dialerDecorators = append(dialerDecorators, dialer)
}

// DialerDecorator will decorate dialer for servername, if return nil, will use the origin dialer.
type DialerDecorator interface {

	// Decorate a dialer
	Decorate(serverName string, dialer proxy.Dialer) proxy.Dialer
}

// decorateDialer iterate all decorator to decorate dialer
func decorateDialer(t *dynamic.TCPServersTransport, d proxy.Dialer) proxy.Dialer {
	if nil == t || nil == t.TLS || len(dialerDecorators) < 1 {
		return d
	}
	for _, decorator := range dialerDecorators {
		if dialer := decorator.Decorate(t.TLS.ServerName, d); nil != dialer {
			return dialer
		}
	}
	return d
}

// proxyDialer decorate first if decorator present, proxy second if proxy enable.
func proxyDialer(t *dynamic.TCPServersTransport, d proxy.Dialer) proxy.Dialer {
	dialer := decorateDialer(t, d)
	if "" == t.Proxy {
		return proxy.FromEnvironmentUsing(dialer)
	}
	u, err := url.Parse(t.Proxy)
	if nil != err {
		log.Error().Msgf("Error while create transport proxy, %v", err)
		return proxy.FromEnvironmentUsing(dialer)
	}
	proxied, err := proxy.FromURL(u, dialer)
	if nil != err {
		log.Error().Msgf("Error while create transport proxy, %v", err)
		return proxy.FromEnvironmentUsing(dialer)
	}
	return proxied
}
