package service

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
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
	Decorate(serverName string, dialer proxy.ContextDialer) proxy.ContextDialer
}

// decorateDialer iterate all decorator to decorate dialer
func decorateDialer(t *dynamic.ServersTransport, d proxy.ContextDialer) func(ctx context.Context, network, address string) (net.Conn, error) {
	if nil == t || len(dialerDecorators) < 1 {
		return d.DialContext
	}
	for _, decorator := range dialerDecorators {
		if dialer := decorator.Decorate(t.ServerName, d); nil != dialer {
			return dialer.DialContext
		}
	}
	return d.DialContext
}

// proxyURL proxy first if proxy option present, proxy second if proxy environment present.
func proxyURL(t *dynamic.ServersTransport) func(req *http.Request) (*url.URL, error) {
	if nil == t || "" == t.Proxy {
		return http.ProxyFromEnvironment
	}
	uri, err := url.Parse(t.Proxy)
	if nil != err {
		log.Error().Msgf("Error while create transport proxy, %v", err)
		return http.ProxyFromEnvironment
	}
	return http.ProxyURL(uri)
}
