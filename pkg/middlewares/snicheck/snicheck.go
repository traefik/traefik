package snicheck

import (
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/middlewares/requestdecorator"
	"github.com/traefik/traefik/v3/pkg/muxer/tcp"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
)

// SNICheck is an HTTP handler that checks whether the TLS configuration for the server name is the same as for the host header.
type SNICheck struct {
	next                   http.Handler
	tlsOptionsForHost      map[string]string
	tlsOptionsForHostRegex map[*regexp.Regexp]string
}

// New creates a new SNICheck.
func New(tlsOptionsForHost map[string]string, tlsOptionsForHostRegexp map[string]string, next http.Handler) *SNICheck {
	tlsOptionsForHostRegex := make(map[*regexp.Regexp]string)
	for hostRegexp, tlsOptions := range tlsOptionsForHostRegexp {
		preparePattern, err := tcp.PreparePattern(hostRegexp)
		if err != nil {
			log.Error().Err(err).Str("hostRegexp", hostRegexp).Msg("Failed to prepare pattern")
			continue
		}
		re, err := regexp.Compile(preparePattern)
		if err != nil {
			log.Error().Err(err).Str("host", hostRegexp).Str("pattern", preparePattern).Msg("Failed to compile regex")
			continue
		}
		tlsOptionsForHostRegex[re] = tlsOptions
	}
	return &SNICheck{next: next, tlsOptionsForHost: tlsOptionsForHost, tlsOptionsForHostRegex: tlsOptionsForHostRegex}
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
		tlsOptionHeader := s.findTLSOptionName(host, true)
		tlsOptionSNI := s.findTLSOptionName(serverName, false)

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

func (s SNICheck) findTLSOptionName(host string, fqdn bool) string {
	name := s.findTLSOptName(host, fqdn)
	if name != "" {
		return name
	}

	name = s.findTLSOptName(strings.ToLower(host), fqdn)
	if name != "" {
		return name
	}

	return traefiktls.DefaultTLSConfigName
}

func (s SNICheck) findTLSOptName(host string, fqdn bool) string {
	if tlsOptions, ok := s.tlsOptionsForHost[host]; ok {
		return tlsOptions
	}

	for regexp, tlsOptions := range s.tlsOptionsForHostRegex {
		if regexp.MatchString(host) {
			return tlsOptions
		}
	}

	if !fqdn {
		return ""
	}

	if last := len(host) - 1; last >= 0 && host[last] == '.' {
		if tlsOptions, ok := s.tlsOptionsForHost[host[:last]]; ok {
			return tlsOptions
		}

		for regexp, tlsOptions := range s.tlsOptionsForHostRegex {
			if regexp.MatchString(host[:last]) {
				return tlsOptions
			}
		}

		return ""
	}

	hostFqdn := host + "."
	if tlsOptions, ok := s.tlsOptionsForHost[hostFqdn]; ok {
		return tlsOptions
	}

	for regexp, tlsOptions := range s.tlsOptionsForHostRegex {
		if regexp.MatchString(hostFqdn) {
			return tlsOptions
		}
	}

	return ""
}
