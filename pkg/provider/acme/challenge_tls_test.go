package acme

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/asn1"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/challenge/tlsalpn01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/tls/generate"
	"github.com/traefik/traefik/v3/pkg/types"
)

func TestChallengeTLSALPNCertificateLifecycle(t *testing.T) {
	testCases := []struct {
		desc       string
		domain     string
		serverName string
	}{
		{
			desc:       "IPv4",
			domain:     "192.0.2.1",
			serverName: "1.2.0.192.in-addr.arpa",
		},
		{
			desc:       "IPv6",
			domain:     "2001:db8::1",
			serverName: "1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa",
		},
		{
			desc:       "DNS",
			domain:     "example.com",
			serverName: "example.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			certPEM, keyPEM, err := generate.KeyPair(test.serverName, time.Time{})
			require.NoError(t, err)
			regularCert, err := tls.X509KeyPair(certPEM, keyPEM)
			require.NoError(t, err)
			regularConfig := &traefiktls.CertAndStores{
				Certificate: traefiktls.Certificate{CertFile: types.FileOrContent(certPEM), KeyFile: types.FileOrContent(keyPEM)},
				Stores:      []string{traefiktls.DefaultTLSStoreName},
			}
			stores := map[string]traefiktls.Store{
				traefiktls.DefaultTLSStoreName: {DefaultCertificate: &regularConfig.Certificate},
			}
			options := traefiktls.DefaultTLSOptions
			options.SniStrict = true
			configs := map[string]traefiktls.Options{traefiktls.DefaultTLSConfigName: options}
			manager := traefiktls.NewManager(nil)

			messages := make(chan dynamic.Message, 1)
			challenge := NewChallengeTLSALPN()
			require.NoError(t, challenge.Provide(messages, nil))
			t.Cleanup(func() {
				challenge.muChans.Lock()
				defer challenge.muChans.Unlock()

				for key := range challenge.chans {
					challenge.cleanChan(key)
				}
			})
			presentDone := make(chan error, 1)
			go func() {
				presentDone <- challenge.Present(t.Context(), test.domain, "token", "keyAuth")
			}()

			message := receiveTLSChallengeMessage(t, messages)
			require.NotNil(t, message.Configuration.TLS)
			require.Len(t, message.Configuration.TLS.Certificates, 1)
			require.Equal(t, []string{tlsalpn01.ACMETLS1Protocol}, message.Configuration.TLS.Certificates[0].Stores)
			manager.UpdateConfigs(t.Context(), stores, configs, append([]*traefiktls.CertAndStores{regularConfig}, message.Configuration.TLS.Certificates...))
			challenge.ListenConfiguration(*message.Configuration)
			select {
			case err := <-presentDone:
				require.NoError(t, err)
			case <-time.After(5 * time.Second):
				t.Fatal("challenge presentation did not complete after configuration was applied")
			}

			config, err := manager.Get(traefiktls.DefaultTLSStoreName, traefiktls.DefaultTLSConfigName)
			require.NoError(t, err)
			hello := &tls.ClientHelloInfo{ServerName: test.serverName, SupportedProtos: []string{tlsalpn01.ACMETLS1Protocol}}
			certificate, err := config.GetCertificate(hello)
			require.NoError(t, err)
			require.NotNil(t, certificate)
			require.NoError(t, certificate.Leaf.VerifyHostname(test.domain))
			assert.NotEqual(t, regularCert.Certificate, certificate.Certificate)
			digest := sha256.Sum256([]byte("keyAuth"))
			extensionValue, err := asn1.Marshal(digest[:])
			require.NoError(t, err)
			var found bool
			for _, extension := range certificate.Leaf.Extensions {
				if extension.Id.Equal(asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 31}) {
					found = true
					assert.True(t, extension.Critical)
					assert.Equal(t, extensionValue, extension.Value)
				}
			}
			require.True(t, found)

			require.NoError(t, challenge.CleanUp(t.Context(), test.domain, "token", "keyAuth"))
			message = receiveTLSChallengeMessage(t, messages)
			require.NotNil(t, message.Configuration.TLS)
			assert.Empty(t, message.Configuration.TLS.Certificates)
			manager.UpdateConfigs(t.Context(), stores, configs, append([]*traefiktls.CertAndStores{regularConfig}, message.Configuration.TLS.Certificates...))
			config, err = manager.Get(traefiktls.DefaultTLSStoreName, traefiktls.DefaultTLSConfigName)
			require.NoError(t, err)
			certificate, err = config.GetCertificate(hello)
			require.NoError(t, err)
			assert.Nil(t, certificate)

			hello.SupportedProtos = []string{"h2"}
			certificate, err = config.GetCertificate(hello)
			require.NoError(t, err)
			require.NotNil(t, certificate)
			assert.Equal(t, regularCert.Certificate, certificate.Certificate)
		})
	}
}

func receiveTLSChallengeMessage(t *testing.T, messages <-chan dynamic.Message) dynamic.Message {
	t.Helper()

	select {
	case message := <-messages:
		require.Equal(t, providerNameALPN, message.ProviderName)
		require.NotNil(t, message.Configuration)
		return message
	case <-time.After(5 * time.Second):
		t.Fatal("TLS challenge configuration was not received")
		return dynamic.Message{}
	}
}
