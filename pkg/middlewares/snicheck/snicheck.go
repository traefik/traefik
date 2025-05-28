package snicheck

import (
	"net"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/middlewares/requestdecorator"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
)

// SNICheck is an HTTP handler that checks whether the TLS configuration for the server name is the same as for the host header.
type SNICheck struct {
	next              http.Handler
	tlsOptionsForHost map[string]string
}

// New creates a new SNICheck.
func New(tlsOptionsForHost map[string]string, next http.Handler) *SNICheck {
	return &SNICheck{next: next, tlsOptionsForHost: tlsOptionsForHost}
}

func (s SNICheck) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.TLS == nil {
		s.next.ServeHTTP(rw, req)
		return
	}

	host := getHost(req)
	serverName := strings.TrimSpace(req.TLS.ServerName)

	// Domain Fronting
	if !strings.EqualFold(host, serverName) {
		tlsOptionHeader := findTLSOptionName(s.tlsOptionsForHost, host, true)
		tlsOptionSNI := findTLSOptionName(s.tlsOptionsForHost, serverName, false)

		if tlsOptionHeader != tlsOptionSNI {
			log.Debug().
				Str("host", host).
				Str("req.Host", req.Host).
				Str("req.TLS.ServerName", req.TLS.ServerName).
				Msgf("TLS options difference: SNI:%s, Header:%s", tlsOptionSNI, tlsOptionHeader)
			http.Error(rw, http.StatusText(http.StatusMisdirectedRequest), http.StatusMisdirectedRequest)
			return
		}
	}

	s.next.ServeHTTP(rw, req)
}

func getHost(req *http.Request) string {
	h := requestdecorator.GetCNAMEFlatten(req.Context())
	if h != "" {
		return h
	}

	h = requestdecorator.GetCanonizedHost(req.Context())
	if h != "" {
		return h
	}

	host, _, err := net.SplitHostPort(req.Host)
	if err != nil {
		host = req.Host
	}

	return strings.TrimSpace(host)
}

func findTLSOptionName(tlsOptionsForHost map[string]string, host string, fqdn bool) string {
	name := findTLSOptName(tlsOptionsForHost, host, fqdn)
	if name != "" {
		return name
	}

	name = findTLSOptName(tlsOptionsForHost, strings.ToLower(host), fqdn)
	if name != "" {
		return name
	}

	return traefiktls.DefaultTLSConfigName
}

func findTLSOptName(tlsOptionsForHost map[string]string, host string, fqdn bool) string {
	if tlsOptions, ok := tlsOptionsForHost[host]; ok {
		return tlsOptions
	}

	if !fqdn {
		return ""
	}

	if last := len(host) - 1; last >= 0 && host[last] == '.' {
		if tlsOptions, ok := tlsOptionsForHost[host[:last]]; ok {
			return tlsOptions
		}

		return ""
	}

	if tlsOptions, ok := tlsOptionsForHost[host+"."]; ok {
		return tlsOptions
	}

	return ""
}
