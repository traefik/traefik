package tls

import (
	"bytes"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/types"
	"golang.org/x/crypto/ocsp"
)

const certWithOCSPServer = `-----BEGIN CERTIFICATE-----
MIIBgjCCASegAwIBAgICIAAwCgYIKoZIzj0EAwIwEjEQMA4GA1UEAxMHVGVzdCBD
QTAeFw0yMzAxMDExMjAwMDBaFw0yMzAyMDExMjAwMDBaMCAxHjAcBgNVBAMTFU9D
U1AgVGVzdCBDZXJ0aWZpY2F0ZTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABIoe
I/bjo34qony8LdRJD+Jhuk8/S8YHXRHl6rH9t5VFCFtX8lIPN/Ll1zCrQ2KB3Wlb
fxSgiQyLrCpZyrdhVPSjXzBdMAwGA1UdEwEB/wQCMAAwHwYDVR0jBBgwFoAU+Eo3
5sST4LRrwS4dueIdGBZ5d7IwLAYIKwYBBQUHAQEEIDAeMBwGCCsGAQUFBzABhhBv
Y3NwLmV4YW1wbGUuY29tMAoGCCqGSM49BAMCA0kAMEYCIQDg94xY/+/VepESdvTT
ykCwiWOS2aCpjyryrKpwMKkR0AIhAPc/+ZEz4W10OENxC1t+NUTvS8JbEGOwulkZ
z9yfaLuD
-----END CERTIFICATE-----`

const certWithoutOCSPServer = `-----BEGIN CERTIFICATE-----
MIIBUzCB+aADAgECAgIgADAKBggqhkjOPQQDAjASMRAwDgYDVQQDEwdUZXN0IENB
MB4XDTIzMDEwMTEyMDAwMFoXDTIzMDIwMTEyMDAwMFowIDEeMBwGA1UEAxMVT0NT
UCBUZXN0IENlcnRpZmljYXRlMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEih4j
9uOjfiqifLwt1EkP4mG6Tz9LxgddEeXqsf23lUUIW1fyUg838uXXMKtDYoHdaVt/
FKCJDIusKlnKt2FU9KMxMC8wDAYDVR0TAQH/BAIwADAfBgNVHSMEGDAWgBT4Sjfm
xJPgtGvBLh254h0YFnl3sjAKBggqhkjOPQQDAgNJADBGAiEA3rWetLGblfSuNZKf
5CpZxhj3A0BjEocEh+2P+nAgIdUCIQDIgptabR1qTLQaF2u0hJsEX2IKuIUvYWH3
6Lb92+zIHg==
-----END CERTIFICATE-----`

// certKey is the private key for both certWithOCSPServer and
// certWithoutOCSPServer.
const certKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEINnVcgrSNh4HlThWlZpegq14M8G/p9NVDtdVjZrseUGLoAoGCCqGSM49
AwEHoUQDQgAEih4j9uOjfiqifLwt1EkP4mG6Tz9LxgddEeXqsf23lUUIW1fyUg83
8uXXMKtDYoHdaVt/FKCJDIusKlnKt2FU9A==
-----END EC PRIVATE KEY-----`

// caCert is the issuing certificate for certWithOCSPServer and
// certWithoutOCSPServer.
const caCert = `-----BEGIN CERTIFICATE-----
MIIBazCCARGgAwIBAgICEAAwCgYIKoZIzj0EAwIwEjEQMA4GA1UEAxMHVGVzdCBD
QTAeFw0yMzAxMDExMjAwMDBaFw0yMzAyMDExMjAwMDBaMBIxEDAOBgNVBAMTB1Rl
c3QgQ0EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASdKexSor/aeazDM57UHhAX
rCkJxUeF2BWf0lZYCRxc3f0GdrEsVvjJW8+/E06eAzDCGSdM/08Nvun1nb6AmAlt
o1cwVTAOBgNVHQ8BAf8EBAMCAQYwEwYDVR0lBAwwCgYIKwYBBQUHAwkwDwYDVR0T
AQH/BAUwAwEB/zAdBgNVHQ4EFgQU+Eo35sST4LRrwS4dueIdGBZ5d7IwCgYIKoZI
zj0EAwIDSAAwRQIgGbA39+kETTB/YMLBFoC2fpZe1cDWfFB7TUdfINUqdH4CIQCR
ByUFC8A+hRNkK5YNH78bgjnKk/88zUQF5ONy4oPGdQ==
-----END CERTIFICATE-----`

const caKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIDJ59ptjq3MzILH4zn5IKoH1sYn+zrUeq2kD8+DD2x+OoAoGCCqGSM49
AwEHoUQDQgAEnSnsUqK/2nmswzOe1B4QF6wpCcVHhdgVn9JWWAkcXN39BnaxLFb4
yVvPvxNOngMwwhknTP9PDb7p9Z2+gJgJbQ==
-----END EC PRIVATE KEY-----`

func TestAppendCertificate(t *testing.T) {
	storesCertificates := make(map[string]map[string]*CertificateData)
	data := CertificateData{
		config: &Certificate{
			CertFile: "../../integration/fixtures/https/snitest.com.cert",
			KeyFile:  "../../integration/fixtures/https/snitest.com.key",
		},
	}

	require.NoError(t, data.AppendCertificate(storesCertificates, DefaultTLSStoreName))
	assert.Equal(t, []string{"snitest.com"}, storesCertificates[DefaultTLSStoreName]["snitest.com"].config.SANs)
	assert.False(t, storesCertificates[DefaultTLSStoreName]["snitest.com"].MustStaple)

	data = CertificateData{
		config: &Certificate{
			CertFile: "../../integration/fixtures/https/uppercase_wildcard.www.snitest.com.cert",
			KeyFile:  "../../integration/fixtures/https/uppercase_wildcard.www.snitest.com.key",
		},
	}

	require.NoError(t, data.AppendCertificate(storesCertificates, DefaultTLSStoreName))
	assert.Equal(t, []string{"*.www.snitest.com", "foo.www.snitest.com"}, storesCertificates[DefaultTLSStoreName]["*.www.snitest.com,foo.www.snitest.com"].config.SANs)
	assert.False(t, storesCertificates[DefaultTLSStoreName]["*.www.snitest.com,foo.www.snitest.com"].MustStaple)
}

func TestStapleOCSP(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		config := types.OCSPConfig{DisableStapling: true}
		cert := mustMakeCertificate(t, config, certWithOCSPServer, certKey)
		err := cert.StapleOCSP()
		if err != nil {
			t.Error("unexpected error:", err)
		} else if cert.Certificate.OCSPStaple != nil {
			t.Error("unexpected OCSP staple")
		}
	})
	t.Run("no OCSP server", func(t *testing.T) {
		cert := mustMakeCertificate(t, types.OCSPConfig{}, certWithoutOCSPServer, certKey)
		err := cert.StapleOCSP()

		if err == nil {
			t.Error("unexpected success")
		}
		if err.Error() != "no OCSP stapling for []: no OCSP server specified in certificate" {
			t.Error("expected \"no OCSP stapling\" in error", err)
		}
	})

	// Start an OCSP responder test server.
	responses := make(map[string][]byte)
	responder := startOCSPResponder(t, responses)
	t.Cleanup(responder.Close)

	// The certWithOCSPServer certificate has a bogus ocsp.example.com endpoint.
	// Use the ResponderOverrides option to point to the test server instead.
	config := types.OCSPConfig{
		ResponderOverrides: map[string]string{
			"ocsp.example.com": responder.URL,
		},
	}

	ca := mustMakeCertificate(t, config, caCert, caKey)

	t.Run("ok", func(t *testing.T) {
		cert := mustMakeCertificate(t, config, certWithOCSPServer+"\n"+caCert, certKey)
		tpl := ocsp.Response{
			Status:       ocsp.Good,
			SerialNumber: cert.Certificate.Leaf.SerialNumber,
		}
		r, err := ocsp.CreateResponse(
			ca.Certificate.Leaf, ca.Certificate.Leaf, tpl, ca.Certificate.PrivateKey.(crypto.Signer))
		if err != nil {
			t.Fatal("couldn't create OCSP response", err)
		}
		responses[cert.Certificate.Leaf.SerialNumber.String()] = r

		err = cert.StapleOCSP()
		if err != nil {
			t.Error("unexpected error:", err)
		} else if !bytes.Equal(cert.Certificate.OCSPStaple, r) {
			t.Error("expected OCSP response to be stapled to certificate")
		}
	})
	t.Run("revoked", func(t *testing.T) {
		cert := mustMakeCertificate(t, config, certWithOCSPServer+"\n"+caCert, certKey)
		tpl := ocsp.Response{
			Status:       ocsp.Revoked,
			SerialNumber: cert.Certificate.Leaf.SerialNumber,
		}
		r, err := ocsp.CreateResponse(
			ca.Certificate.Leaf, ca.Certificate.Leaf, tpl, ca.Certificate.PrivateKey.(crypto.Signer))
		if err != nil {
			t.Fatal("couldn't create OCSP response", err)
		}
		responses[cert.Certificate.Leaf.SerialNumber.String()] = r

		err = cert.StapleOCSP()
		if err != nil {
			t.Error("unexpected error:", err)
		} else if cert.Certificate.OCSPStaple != nil {
			t.Error("revoked OCSP response should not be stapled")
		}
	})
	t.Run("no issuing cert", func(t *testing.T) {
		cert := mustMakeCertificate(t, config, certWithOCSPServer, certKey)
		cert.OCSPServer = nil

		err := cert.StapleOCSP()
		expected := "no OCSP stapling for []: no OCSP server specified in certificate"
		if err == nil || err.Error() != expected {
			t.Errorf("expected error %q but got %q", expected, err)
		}
	})
}

func mustMakeCertificate(t *testing.T, config types.OCSPConfig, cert, key string) CertificateData {
	t.Helper()
	tlsCert, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		t.Fatal("couldn't make certificate:", err)
	}

	parsedCert, _ := x509.ParseCertificate(tlsCert.Certificate[0])
	tlsCert.Leaf = parsedCert

	return CertificateData{
		config: &Certificate{
			CertFile: "",
			KeyFile:  "",
			OCSP:     config,
			SANs:     nil,
		},
		ocspLock:     nil,
		Certificate:  &tlsCert,
		MustStaple:   false,
		OCSPServer:   parsedCert.OCSPServer,
		OCSPResponse: nil,
	}
}

func startOCSPResponder(
	t *testing.T, responses map[string][]byte,
) *httptest.Server {
	t.Helper()

	h := func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/ocsp-request" {
			t.Errorf("unexpected request Content-Type %q", ct)
		}
		b, _ := io.ReadAll(r.Body)
		request, err := ocsp.ParseRequest(b)
		if err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/ocsp-response")
		_, _ = w.Write(responses[request.SerialNumber.String()])
	}
	return httptest.NewServer(http.HandlerFunc(h))
}
