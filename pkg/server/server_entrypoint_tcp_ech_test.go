package server

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/static"
	tcprouter "github.com/traefik/traefik/v3/pkg/server/router/tcp"
	"github.com/traefik/traefik/v3/pkg/tcp"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
)

func TestTCPECHInnerTLSOptions(t *testing.T) {
	certContent, err := localhostCert.Read()
	require.NoError(t, err)

	keyContent, err := localhostKey.Read()
	require.NoError(t, err)

	tlsCert, err := tls.X509KeyPair(certContent, keyContent)
	require.NoError(t, err)

	echKey, err := traefiktls.NewECHKey("public.example.org")
	require.NoError(t, err)

	epConfig := &static.EntryPointsTransport{}
	epConfig.SetDefaults()

	entryPoint, err := NewTCPEntryPoint(t.Context(), "foo", &static.EntryPoint{
		Address:          "127.0.0.1:0",
		Transport:        epConfig,
		ForwardedHeaders: &static.ForwardedHeaders{},
		HTTP2:            &static.HTTP2Config{},
	}, nil, nil)
	require.NoError(t, err)

	router, err := tcprouter.NewRouter(nil)
	require.NoError(t, err)

	// The outer ClientHello carries the public name, so the ECH keys live on its TLS config.
	// Disabled session tickets exercise the ticket-keys blending of the standard library,
	// which aborts the handshake when the re-selected config re-enables tickets without keys.
	router.AddHTTPTLSConfig("public.example.org", &tls.Config{
		Certificates:             []tls.Certificate{tlsCert},
		EncryptedClientHelloKeys: []tls.EncryptedClientHelloKey{*echKey},
		SessionTicketsDisabled:   true,
	}, traefiktls.DefaultTLSConfigName)
	// The protected domain uses a distinct TLS option requesting a client certificate.
	router.AddHTTPTLSConfig("example.com", &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientAuth:   tls.RequestClientCert,
	}, "hidden@file")
	router.SetHTTPSHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprint(rw, tcp.GetTLSOptionsName(req.Context()))
	}), nil)

	ctx := t.Context()
	go entryPoint.Start(ctx)
	entryPoint.SwitchRouter(router)
	t.Cleanup(func() { entryPoint.Shutdown(ctx) })

	configList, err := traefiktls.ECHConfigToConfigList(echKey.Config)
	require.NoError(t, err)

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certContent)

	var clientCertRequested atomic.Bool
	conn, err := tls.Dial("tcp", entryPoint.listener.Addr().String(), &tls.Config{
		RootCAs:                        certPool,
		ServerName:                     "example.com",
		MinVersion:                     tls.VersionTLS13,
		EncryptedClientHelloConfigList: configList,
		GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			clientCertRequested.Store(true)
			return &tls.Certificate{}, nil
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	require.True(t, conn.ConnectionState().ECHAccepted)
	// The certificate request proves the hidden domain's TLS config governed the handshake.
	assert.True(t, clientCertRequested.Load())

	request, err := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	require.NoError(t, err)
	require.NoError(t, request.Write(conn))

	response, err := http.ReadResponse(bufio.NewReader(conn), request)
	require.NoError(t, err)
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, "hidden@file", string(body))

	// Without a default TLS config, an inner name matching no route must fail closed.
	unknownConn, err := tls.Dial("tcp", entryPoint.listener.Addr().String(), &tls.Config{
		InsecureSkipVerify:             true,
		RootCAs:                        certPool,
		ServerName:                     "unknown.example.net",
		MinVersion:                     tls.VersionTLS13,
		EncryptedClientHelloConfigList: configList,
	})
	if err == nil {
		_ = unknownConn.Close()
	}
	require.Error(t, err)
}
