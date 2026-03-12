package filtertlsclientcert

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

// minimalCheeseCrt is a client certificate with Subject: C=FR, ST=Some-State, O=Cheese
// issued by Simple Signing CA 2 (DC=org,DC=cheese).
const minimalCheeseCrt = `-----BEGIN CERTIFICATE-----
MIIEQDCCAygCFFRY0OBk/L5Se0IZRj3CMljawL2UMA0GCSqGSIb3DQEBCwUAMIIB
hDETMBEGCgmSJomT8ixkARkWA29yZzEWMBQGCgmSJomT8ixkARkWBmNoZWVzZTEP
MA0GA1UECgwGQ2hlZXNlMREwDwYDVQQKDAhDaGVlc2UgMjEfMB0GA1UECwwWU2lt
cGxlIFNpZ25pbmcgU2VjdGlvbjEhMB8GA1UECwwYU2ltcGxlIFNpZ25pbmcgU2Vj
dGlvbiAyMRowGAYDVQQDDBFTaW1wbGUgU2lnbmluZyBDQTEcMBoGA1UEAwwTU2lt
cGxlIFNpZ25pbmcgQ0EgMjELMAkGA1UEBhMCRlIxCzAJBgNVBAYTAlVTMREwDwYD
VQQHDAhUT1VMT1VTRTENMAsGA1UEBwwETFlPTjEWMBQGA1UECAwNU2lnbmluZyBT
dGF0ZTEYMBYGA1UECAwPU2lnbmluZyBTdGF0ZSAyMSEwHwYJKoZIhvcNAQkBFhJz
aW1wbGVAc2lnbmluZy5jb20xIjAgBgkqhkiG9w0BCQEWE3NpbXBsZTJAc2lnbmlu
Zy5jb20wHhcNMTgxMjA2MTExMDM2WhcNMjEwOTI1MTExMDM2WjAzMQswCQYDVQQG
EwJGUjETMBEGA1UECAwKU29tZS1TdGF0ZTEPMA0GA1UECgwGQ2hlZXNlMIIBIjAN
BgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAskX/bUtwFo1gF2BTPNaNcTUMaRFu
FMZozK8IgLjccZ4kZ0R9oFO6Yp8Zl/IvPaf7tE26PI7XP7eHriUdhnQzX7iioDd0
RZa68waIhAGc+xPzRFrP3b3yj3S2a9Rve3c0K+SCV+EtKAwsxMqQDhoo9PcBfo5B
RHfht07uD5MncUcGirwN+/pxHV5xzAGPcc7On0/5L7bq/G+63nhu78zw9XyuLaHC
PM5VbOUvpyIESJHbMMzTdFGL8ob9VKO+Kr1kVGdEA9i8FLGl3xz/GBKuW/JD0xyW
DrU29mri5vYWHmkuv7ZWHGXnpXjTtPHwveE9/0/ArnmpMyR9JtqFr1oEvQIDAQAB
MA0GCSqGSIb3DQEBCwUAA4IBAQBHta+NWXI08UHeOkGzOTGRiWXsOH2dqdX6gTe9
xF1AIjyoQ0gvpoGVvlnChSzmlUj+vnx/nOYGIt1poE3hZA3ZHZD/awsvGyp3GwWD
IfXrEViSCIyF+8tNNKYyUcEO3xdAsAUGgfUwwF/mZ6MBV5+A/ZEEILlTq8zFt9dV
vdKzIt7fZYxYBBHFSarl1x8pDgWXlf3hAufevGJXip9xGYmznF0T5cq1RbWJ4be3
/9K7yuWhuBYC3sbTbCneHBa91M82za+PIISc1ygCYtWSBoZKSAqLk0rkZpHaekDP
WqeUSNGYV//RunTeuRDAf5OxehERb1srzBXhRZ3cZdzXbgR/
-----END CERTIFICATE-----`

var nextOK = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func parseCert(t *testing.T, certPEM string) *x509.Certificate {
	t.Helper()

	block, _ := pem.Decode([]byte(certPEM))
	require.NotNil(t, block, "failed to PEM-decode certificate")

	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	return cert
}

func newTLS(certs ...*x509.Certificate) *tls.ConnectionState {
	return &tls.ConnectionState{PeerCertificates: certs}
}

func TestNew_InvalidRegex(t *testing.T) {
	t.Parallel()

	_, err := New(context.Background(), nextOK, dynamic.FilterTLSClientCert{Subject: "["}, "test")
	require.Error(t, err)

	_, err = New(context.Background(), nextOK, dynamic.FilterTLSClientCert{Issuer: "["}, "test")
	require.Error(t, err)
}

func TestFilterTLSClientCert_ServeHTTP(t *testing.T) {
	t.Parallel()

	cert := parseCert(t, minimalCheeseCrt)

	testCases := []struct {
		desc           string
		config         dynamic.FilterTLSClientCert
		tlsState       *tls.ConnectionState
		expectedStatus int
	}{
		{
			desc:           "no TLS state → 403",
			config:         dynamic.FilterTLSClientCert{},
			tlsState:       nil,
			expectedStatus: http.StatusForbidden,
		},
		{
			desc:           "TLS without peer cert → 403",
			config:         dynamic.FilterTLSClientCert{},
			tlsState:       &tls.ConnectionState{},
			expectedStatus: http.StatusForbidden,
		},
		{
			desc:           "no filter, cert present → 200",
			config:         dynamic.FilterTLSClientCert{},
			tlsState:       newTLS(cert),
			expectedStatus: http.StatusOK,
		},
		{
			desc:           "subject regex matches O=Cheese → 200",
			config:         dynamic.FilterTLSClientCert{Subject: `O=Cheese`},
			tlsState:       newTLS(cert),
			expectedStatus: http.StatusOK,
		},
		{
			desc:           "subject regex does not match → 403",
			config:         dynamic.FilterTLSClientCert{Subject: `O=NotMine`},
			tlsState:       newTLS(cert),
			expectedStatus: http.StatusForbidden,
		},
		{
			desc:           "issuer regex matches Simple Signing CA → 200",
			config:         dynamic.FilterTLSClientCert{Issuer: `CN=Simple Signing CA`},
			tlsState:       newTLS(cert),
			expectedStatus: http.StatusOK,
		},
		{
			desc:           "issuer regex does not match → 403",
			config:         dynamic.FilterTLSClientCert{Issuer: `CN=Unknown CA`},
			tlsState:       newTLS(cert),
			expectedStatus: http.StatusForbidden,
		},
		{
			desc:           "both subject and issuer match → 200",
			config:         dynamic.FilterTLSClientCert{Subject: `O=Cheese`, Issuer: `CN=Simple Signing CA`},
			tlsState:       newTLS(cert),
			expectedStatus: http.StatusOK,
		},
		{
			desc:           "subject matches but issuer does not → 403",
			config:         dynamic.FilterTLSClientCert{Subject: `O=Cheese`, Issuer: `CN=Unknown CA`},
			tlsState:       newTLS(cert),
			expectedStatus: http.StatusForbidden,
		},
		{
			desc:           "issuer matches but subject does not → 403",
			config:         dynamic.FilterTLSClientCert{Subject: `O=NotMine`, Issuer: `CN=Simple Signing CA`},
			tlsState:       newTLS(cert),
			expectedStatus: http.StatusForbidden,
		},
		{
			desc:           "subject anchored match (starts with C=FR) → 200",
			config:         dynamic.FilterTLSClientCert{Subject: `^C=FR,`},
			tlsState:       newTLS(cert),
			expectedStatus: http.StatusOK,
		},
		{
			desc:           "subject full anchored match → 200",
			config:         dynamic.FilterTLSClientCert{Subject: `^C=FR,ST=Some-State,O=Cheese$`},
			tlsState:       newTLS(cert),
			expectedStatus: http.StatusOK,
		},
		{
			// The issuer contains "CN=Simple Signing CA" so CN=.+ must match.
			desc:           "issuer CN=.+ matches non-empty CN → 200",
			config:         dynamic.FilterTLSClientCert{Issuer: `CN=.+`},
			tlsState:       newTLS(cert),
			expectedStatus: http.StatusOK,
		},
		{
			// The subject (C=FR,ST=Some-State,O=Cheese) has no CN field at all.
			desc:           "subject CN=.+ on cert without CN → 403",
			config:         dynamic.FilterTLSClientCert{Subject: `CN=.+`},
			tlsState:       newTLS(cert),
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler, err := New(context.Background(), nextOK, test.config, "test-filter")
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.tlsState != nil {
				req.TLS = test.tlsState
			}

			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatus, rw.Code)
		})
	}
}

func TestBuildDN(t *testing.T) {
	t.Parallel()

	cert := parseCert(t, minimalCheeseCrt)

	subject := buildDN(&cert.Subject)
	assert.Contains(t, subject, "C=FR")
	assert.Contains(t, subject, "ST=Some-State")
	assert.Contains(t, subject, "O=Cheese")

	issuer := buildDN(&cert.Issuer)
	assert.Contains(t, issuer, "CN=Simple Signing CA")
}
