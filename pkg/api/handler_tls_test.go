package api

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/config/static"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/tls/generate"
	"github.com/traefik/traefik/v3/pkg/types"
)

func TestHandler_TLS_Certificates(t *testing.T) {
	testCases := []struct {
		desc           string
		path           string
		tlsManager     *traefiktls.Manager
		expectedStatus int
		expectedCount  int
	}{
		{
			desc:           "no TLS manager",
			path:           "/api/tls/certificates",
			tlsManager:     nil,
			expectedStatus: http.StatusServiceUnavailable,
			expectedCount:  0,
		},
		{
			desc:           "empty TLS manager",
			path:           "/api/tls/certificates",
			tlsManager:     traefiktls.NewManager(nil),
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			desc:           "TLS manager with certificates",
			path:           "/api/tls/certificates",
			tlsManager:     createTLSManagerWithCerts(t),
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := New(static.Configuration{API: &static.API{}, Global: &static.Global{}}, &runtime.Configuration{}, test.tlsManager)
			server := httptest.NewServer(handler.createRouter())

			resp, err := http.DefaultClient.Get(server.URL + test.path)
			require.NoError(t, err)

			require.Equal(t, test.expectedStatus, resp.StatusCode)

			if test.expectedStatus == http.StatusOK {
				assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
				contents, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var certificates []CertificateInfo
				err = json.Unmarshal(contents, &certificates)
				require.NoError(t, err)

				assert.Len(t, certificates, test.expectedCount)

				if test.expectedCount > 0 {
					cert := certificates[0]
					assert.NotEmpty(t, cert.Name)
					assert.NotEmpty(t, cert.Domains)
					assert.False(t, cert.Expiration.IsZero())
				}
			}

			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}

func createTLSManagerWithCerts(t *testing.T) *traefiktls.Manager {
	t.Helper()

	// Create a test certificate
	certPEM, keyPEM, err := generate.KeyPair("example.com", time.Now().Add(365*24*time.Hour))
	require.NoError(t, err)

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)

	// Create TLS manager and add certificate
	tlsManager := traefiktls.NewManager(nil)

	// Create certificate data
	certData := &traefiktls.CertificateData{
		Certificate: &cert,
	}

	// Create a store with the certificate
	store := traefiktls.NewCertificateStore(nil)
	certs := map[string]*traefiktls.CertificateData{
		"example.com": certData,
	}
	store.DynamicCerts.Set(certs)

	// Update the TLS manager with the store
	stores := map[string]traefiktls.Store{
		traefiktls.DefaultTLSStoreName: {},
	}

	certificates := []*traefiktls.CertAndStores{
		{
			Certificate: traefiktls.Certificate{
				CertFile: types.FileOrContent(string(certPEM)),
				KeyFile:  types.FileOrContent(string(keyPEM)),
			},
			Stores: []string{traefiktls.DefaultTLSStoreName},
		},
	}

	tlsManager.UpdateConfigs(t.Context(), stores, nil, certificates)

	return tlsManager
}
