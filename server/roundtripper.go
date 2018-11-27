package server

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"time"

	"github.com/containous/traefik/config/static"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/old/configuration"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/pkg/errors"
	"golang.org/x/net/http2"
)

type h2cTransportWrapper struct {
	*http2.Transport
}

func (t *h2cTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	return t.Transport.RoundTrip(req)
}

// createHTTPTransport creates an http.Transport configured with the Transport configuration settings.
// For the settings that can't be configured in Traefik it uses the default http.Transport settings.
// An exception to this is the MaxIdleConns setting as we only provide the option MaxIdleConnsPerHost
// in Traefik at this point in time. Setting this value to the default of 100 could lead to confusing
// behavior and backwards compatibility issues.
func createHTTPTransport(transportConfiguration *static.ServersTransport) (*http.Transport, error) {
	if transportConfiguration == nil {
		return nil, errors.New("no transport configuration given")
	}

	dialer := &net.Dialer{
		Timeout:   configuration.DefaultDialTimeout,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	if transportConfiguration.ForwardingTimeouts != nil {
		dialer.Timeout = time.Duration(transportConfiguration.ForwardingTimeouts.DialTimeout)
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConnsPerHost:   transportConfiguration.MaxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	transport.RegisterProtocol("h2c", &h2cTransportWrapper{
		Transport: &http2.Transport{
			DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(netw, addr)
			},
			AllowHTTP: true,
		},
	})

	if transportConfiguration.ForwardingTimeouts != nil {
		transport.ResponseHeaderTimeout = time.Duration(transportConfiguration.ForwardingTimeouts.ResponseHeaderTimeout)
	}

	if transportConfiguration.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if len(transportConfiguration.RootCAs) > 0 {
		transport.TLSClientConfig = &tls.Config{
			RootCAs: createRootCACertPool(transportConfiguration.RootCAs),
		}
	}

	err := http2.ConfigureTransport(transport)
	if err != nil {
		return nil, err
	}

	return transport, nil
}

func createRootCACertPool(rootCAs traefiktls.FilesOrContents) *x509.CertPool {
	roots := x509.NewCertPool()

	for _, cert := range rootCAs {
		certContent, err := cert.Read()
		if err != nil {
			log.WithoutContext().Error("Error while read RootCAs", err)
			continue
		}
		roots.AppendCertsFromPEM(certContent)
	}

	return roots
}
