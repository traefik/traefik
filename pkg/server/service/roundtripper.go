package service

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/traefik/traefik/v2/pkg/config/static"
	"github.com/traefik/traefik/v2/pkg/log"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
	"golang.org/x/net/http2"
)

type h2cTransportWrapper struct {
	*http2.Transport
}

func (t *h2cTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	return t.Transport.RoundTrip(req)
}

// createRoundtripper creates an http.Roundtripper configured with the Transport configuration settings.
// For the settings that can't be configured in Traefik it uses the default http.Transport settings.
// An exception to this is the MaxIdleConns setting as we only provide the option MaxIdleConnsPerHost
// in Traefik at this point in time. Setting this value to the default of 100 could lead to confusing
// behavior and backwards compatibility issues.
func createRoundtripper(transportConfiguration *static.ServersTransport) (http.RoundTripper, error) {
	if transportConfiguration == nil {
		return nil, errors.New("no transport configuration given")
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
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
		transport.IdleConnTimeout = time.Duration(transportConfiguration.ForwardingTimeouts.IdleConnTimeout)
	}

	if transportConfiguration.InsecureSkipVerify || len(transportConfiguration.RootCAs) > 0 {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: transportConfiguration.InsecureSkipVerify,
			RootCAs:            createRootCACertPool(transportConfiguration.RootCAs),
		}
	}

	smartTransport, err := newSmartRoundTripper(transport)
	if err != nil {
		return nil, err
	}

	return smartTransport, nil
}

func createRootCACertPool(rootCAs []traefiktls.FileOrContent) *x509.CertPool {
	if len(rootCAs) == 0 {
		return nil
	}

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

func setupDefaultRoundTripper(conf *static.ServersTransport) http.RoundTripper {
	transport, err := createRoundtripper(conf)
	if err != nil {
		log.WithoutContext().Errorf("Could not configure HTTP Transport, fallbacking on default transport: %v", err)
		return http.DefaultTransport
	}

	return transport
}
