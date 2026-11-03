package tls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// certKey is the private key for both certWithOCSPServer and certWithoutOCSPServer.
const certKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEINnVcgrSNh4HlThWlZpegq14M8G/p9NVDtdVjZrseUGLoAoGCCqGSM49
AwEHoUQDQgAEih4j9uOjfiqifLwt1EkP4mG6Tz9LxgddEeXqsf23lUUIW1fyUg83
8uXXMKtDYoHdaVt/FKCJDIusKlnKt2FU9A==
-----END EC PRIVATE KEY-----`

// caCert is the issuing certificate for certWithOCSPServer and certWithoutOCSPServer.
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

func TestOCSPStapler_Upsert(t *testing.T) {
	ocspStapler := newOCSPStapler(nil)

	issuerCert, err := tls.X509KeyPair([]byte(caCert), []byte(caKey))
	require.NoError(t, err)

	leafCert, err := tls.X509KeyPair([]byte(certWithOCSPServer), []byte(certKey))
	require.NoError(t, err)

	// Upsert a certificate without an OCSP server should raise an error.
	leafCertWithoutOCSPServer, err := tls.X509KeyPair([]byte(certWithoutOCSPServer), []byte(certKey))
	require.NoError(t, err)

	err = ocspStapler.Upsert("foo", leafCertWithoutOCSPServer.Leaf, issuerCert.Leaf)
	require.Error(t, err)

	// Upsert a certificate with an OCSP server.
	err = ocspStapler.Upsert("foo", leafCert.Leaf, issuerCert.Leaf)
	require.NoError(t, err)

	i, ok := ocspStapler.cache.Get("foo")
	require.True(t, ok)

	e, ok := i.(*ocspEntry)
	require.True(t, ok)

	assert.Equal(t, leafCert.Leaf, e.leaf)
	assert.Equal(t, issuerCert.Leaf, e.issuer)
	assert.Nil(t, e.staple)
	assert.Equal(t, []string{"ocsp.example.com"}, e.responders)
	assert.Equal(t, int64(0), ocspStapler.cache.Items()["foo"].Expiration)

	// Upsert an existing entry to make sure that the existing staple is preserved.
	e.staple = []byte("foo")
	e.nextUpdate = time.Now()
	e.responders = []string{"foo.com"}

	err = ocspStapler.Upsert("foo", leafCert.Leaf, issuerCert.Leaf)
	require.NoError(t, err)

	i, ok = ocspStapler.cache.Get("foo")
	require.True(t, ok)

	e, ok = i.(*ocspEntry)
	require.True(t, ok)

	assert.Equal(t, leafCert.Leaf, e.leaf)
	assert.Equal(t, issuerCert.Leaf, e.issuer)
	assert.Equal(t, []byte("foo"), e.staple)
	assert.NotZero(t, e.nextUpdate)
	assert.Equal(t, []string{"foo.com"}, e.responders)
	assert.Equal(t, int64(0), ocspStapler.cache.Items()["foo"].Expiration)
}

func TestOCSPStapler_Upsert_withResponderOverrides(t *testing.T) {
	ocspStapler := newOCSPStapler(map[string]string{
		"ocsp.example.com": "foo.com",
	})

	issuerCert, err := tls.X509KeyPair([]byte(caCert), []byte(caKey))
	require.NoError(t, err)

	leafCert, err := tls.X509KeyPair([]byte(certWithOCSPServer), []byte(certKey))
	require.NoError(t, err)

	err = ocspStapler.Upsert("foo", leafCert.Leaf, issuerCert.Leaf)
	require.NoError(t, err)

	i, ok := ocspStapler.cache.Get("foo")
	require.True(t, ok)

	e, ok := i.(*ocspEntry)
	require.True(t, ok)

	assert.Equal(t, leafCert.Leaf, e.leaf)
	assert.Equal(t, issuerCert.Leaf, e.issuer)
	assert.Nil(t, e.staple)
	assert.Equal(t, []string{"foo.com"}, e.responders)
}

func TestOCSPStapler_ResetTTL(t *testing.T) {
	ocspStapler := newOCSPStapler(nil)

	issuerCert, err := tls.X509KeyPair([]byte(caCert), []byte(caKey))
	require.NoError(t, err)

	leafCert, err := tls.X509KeyPair([]byte(certWithOCSPServer), []byte(certKey))
	require.NoError(t, err)

	ocspStapler.cache.Set("foo", &ocspEntry{
		leaf:       leafCert.Leaf,
		issuer:     issuerCert.Leaf,
		responders: []string{"foo.com"},
		nextUpdate: time.Now(),
		staple:     []byte("foo"),
	}, cache.NoExpiration)

	ocspStapler.cache.Set("bar", &ocspEntry{
		leaf:       leafCert.Leaf,
		issuer:     issuerCert.Leaf,
		responders: []string{"bar.com"},
		nextUpdate: time.Now(),
		staple:     []byte("bar"),
	}, time.Hour)

	wantBarExpiration := ocspStapler.cache.Items()["bar"].Expiration

	ocspStapler.ResetTTL()

	item, ok := ocspStapler.cache.Items()["foo"]
	require.True(t, ok)

	e, ok := item.Object.(*ocspEntry)
	require.True(t, ok)

	assert.Positive(t, item.Expiration)
	assert.Equal(t, leafCert.Leaf, e.leaf)
	assert.Equal(t, issuerCert.Leaf, e.issuer)
	assert.Equal(t, []byte("foo"), e.staple)
	assert.NotZero(t, e.nextUpdate)
	assert.Equal(t, []string{"foo.com"}, e.responders)

	item, ok = ocspStapler.cache.Items()["bar"]
	require.True(t, ok)

	e, ok = item.Object.(*ocspEntry)
	require.True(t, ok)

	assert.Equal(t, wantBarExpiration, item.Expiration)
	assert.Equal(t, leafCert.Leaf, e.leaf)
	assert.Equal(t, issuerCert.Leaf, e.issuer)
	assert.Equal(t, []byte("bar"), e.staple)
	assert.NotZero(t, e.nextUpdate)
	assert.Equal(t, []string{"bar.com"}, e.responders)
}

func TestOCSPStapler_GetStaple(t *testing.T) {
	ocspStapler := newOCSPStapler(nil)

	// Get an un-existing staple.
	staple, exists := ocspStapler.GetStaple("foo")

	assert.False(t, exists)
	assert.Nil(t, staple)

	// Get an existing staple.
	ocspStapler.cache.Set("foo", &ocspEntry{staple: []byte("foo")}, cache.NoExpiration)

	staple, exists = ocspStapler.GetStaple("foo")

	assert.True(t, exists)
	assert.Equal(t, []byte("foo"), staple)
}

func TestOCSPStapler_updateStaple(t *testing.T) {
	leafCert, err := tls.X509KeyPair([]byte(certWithOCSPServer), []byte(certKey))
	require.NoError(t, err)

	issuerCert, err := tls.X509KeyPair([]byte(caCert), []byte(caKey))
	require.NoError(t, err)

	thisUpdate, err := time.Parse("2006-01-02", "2025-01-01")
	require.NoError(t, err)
	nextUpdate, err := time.Parse("2006-01-02", "2025-01-02")
	require.NoError(t, err)
	stapleUpdate := thisUpdate.Add(nextUpdate.Sub(thisUpdate) / 2)

	ocspResponseTmpl := ocsp.Response{
		SerialNumber:    leafCert.Leaf.SerialNumber,
		TBSResponseData: []byte("foo"),
		ThisUpdate:      thisUpdate,
		NextUpdate:      nextUpdate,
	}

	ocspResponse, err := ocsp.CreateResponse(leafCert.Leaf, leafCert.Leaf, ocspResponseTmpl, issuerCert.PrivateKey.(crypto.Signer))
	require.NoError(t, err)

	handler := func(rw http.ResponseWriter, req *http.Request) {
		ct := req.Header.Get("Content-Type")
		assert.Equal(t, "application/ocsp-request", ct)

		reqBytes, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		_, err = ocsp.ParseRequest(reqBytes)
		require.NoError(t, err)

		rw.Header().Set("Content-Type", "application/ocsp-response")

		_, err = rw.Write(ocspResponse)
		require.NoError(t, err)
	}

	responder := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(responder.Close)

	responderStatusNotOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(responderStatusNotOK.Close)

	testCases := []struct {
		desc        string
		entry       *ocspEntry
		expectError bool
	}{
		{
			desc: "no responder",
			entry: &ocspEntry{
				leaf:   leafCert.Leaf,
				issuer: issuerCert.Leaf,
			},
			expectError: true,
		},
		{
			desc: "wrong responder",
			entry: &ocspEntry{
				leaf:       leafCert.Leaf,
				issuer:     issuerCert.Leaf,
				responders: []string{"http://foo.bar"},
			},
			expectError: true,
		},
		{
			desc: "not ok status responder",
			entry: &ocspEntry{
				leaf:       leafCert.Leaf,
				issuer:     issuerCert.Leaf,
				responders: []string{responderStatusNotOK.URL},
			},
			expectError: true,
		},
		{
			desc: "one wrong responder, one ok",
			entry: &ocspEntry{
				leaf:       leafCert.Leaf,
				issuer:     issuerCert.Leaf,
				responders: []string{"http://foo.bar", responder.URL},
			},
		},
		{
			desc: "ok responder",
			entry: &ocspEntry{
				leaf:       leafCert.Leaf,
				issuer:     issuerCert.Leaf,
				responders: []string{responder.URL},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ocspStapler := newOCSPStapler(nil)
			ocspStapler.client = &http.Client{Timeout: time.Second}

			err = ocspStapler.updateStaple(t.Context(), test.entry)
			if test.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, ocspResponse, test.entry.staple)
			assert.Equal(t, stapleUpdate.UTC(), test.entry.nextUpdate)
		})
	}
}

func TestOCSPStapler_updateStaple_withoutNextUpdate(t *testing.T) {
	leafCert, err := tls.X509KeyPair([]byte(certWithOCSPServer), []byte(certKey))
	require.NoError(t, err)

	issuerCert, err := tls.X509KeyPair([]byte(caCert), []byte(caKey))
	require.NoError(t, err)

	thisUpdate, err := time.Parse("2006-01-02", "2025-01-01")
	require.NoError(t, err)

	ocspResponseTmpl := ocsp.Response{
		SerialNumber:    leafCert.Leaf.SerialNumber,
		TBSResponseData: []byte("foo"),
		ThisUpdate:      thisUpdate,
	}

	ocspResponse, err := ocsp.CreateResponse(leafCert.Leaf, leafCert.Leaf, ocspResponseTmpl, issuerCert.PrivateKey.(crypto.Signer))
	require.NoError(t, err)

	handler := func(rw http.ResponseWriter, req *http.Request) {
		ct := req.Header.Get("Content-Type")
		assert.Equal(t, "application/ocsp-request", ct)

		reqBytes, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		_, err = ocsp.ParseRequest(reqBytes)
		require.NoError(t, err)

		rw.Header().Set("Content-Type", "application/ocsp-response")

		_, err = rw.Write(ocspResponse)
		require.NoError(t, err)
	}

	responder := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(responder.Close)

	responderStatusNotOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(responderStatusNotOK.Close)

	ocspStapler := newOCSPStapler(nil)
	ocspStapler.client = &http.Client{Timeout: time.Second}

	entry := &ocspEntry{
		leaf:       leafCert.Leaf,
		issuer:     issuerCert.Leaf,
		responders: []string{responder.URL},
	}
	err = ocspStapler.updateStaple(t.Context(), entry)
	require.NoError(t, err)

	assert.Equal(t, ocspResponse, entry.staple)
	assert.NotZero(t, entry.nextUpdate)
	assert.Greater(t, time.Now(), entry.nextUpdate)
}

func TestOCSPStapler_updateStaples(t *testing.T) {
	leafCert, err := tls.X509KeyPair([]byte(certWithOCSPServer), []byte(certKey))
	require.NoError(t, err)

	issuerCert, err := tls.X509KeyPair([]byte(caCert), []byte(caKey))
	require.NoError(t, err)

	thisUpdate, err := time.Parse("2006-01-02", "2025-01-01")
	require.NoError(t, err)
	nextUpdate, err := time.Parse("2006-01-02", "2025-01-02")
	require.NoError(t, err)
	stapleUpdate := thisUpdate.Add(nextUpdate.Sub(thisUpdate) / 2)

	ocspResponseTmpl := ocsp.Response{
		SerialNumber:    leafCert.Leaf.SerialNumber,
		TBSResponseData: []byte("foo"),
		ThisUpdate:      thisUpdate,
		NextUpdate:      nextUpdate,
	}

	ocspResponse, err := ocsp.CreateResponse(leafCert.Leaf, leafCert.Leaf, ocspResponseTmpl, issuerCert.PrivateKey.(crypto.Signer))
	require.NoError(t, err)

	handler := func(rw http.ResponseWriter, req *http.Request) {
		ct := req.Header.Get("Content-Type")
		assert.Equal(t, "application/ocsp-request", ct)

		reqBytes, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		_, err = ocsp.ParseRequest(reqBytes)
		require.NoError(t, err)

		rw.Header().Set("Content-Type", "application/ocsp-response")

		_, err = rw.Write(ocspResponse)
		require.NoError(t, err)
	}

	responder := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(responder.Close)

	ocspStapler := newOCSPStapler(nil)
	ocspStapler.client = &http.Client{Timeout: time.Second}

	// nil staple entry
	ocspStapler.cache.Set("nilStaple", &ocspEntry{
		leaf:       leafCert.Leaf,
		issuer:     issuerCert.Leaf,
		responders: []string{responder.URL},
		nextUpdate: time.Now().Add(-time.Hour),
	}, cache.NoExpiration)
	// staple entry with nextUpdate in the past
	ocspStapler.cache.Set("toUpdate", &ocspEntry{
		leaf:       leafCert.Leaf,
		issuer:     issuerCert.Leaf,
		responders: []string{responder.URL},
		staple:     []byte("foo"),
		nextUpdate: time.Now().Add(-time.Hour),
	}, cache.NoExpiration)
	// staple entry with nextUpdate in the future
	inOneHour := time.Now().Add(time.Hour)
	ocspStapler.cache.Set("noUpdate", &ocspEntry{
		leaf:       leafCert.Leaf,
		issuer:     issuerCert.Leaf,
		responders: []string{responder.URL},
		staple:     []byte("foo"),
		nextUpdate: inOneHour,
	}, cache.NoExpiration)

	ocspStapler.updateStaples(t.Context())

	nilStaple, ok := ocspStapler.cache.Get("nilStaple")
	require.True(t, ok)

	assert.Equal(t, ocspResponse, nilStaple.(*ocspEntry).staple)
	assert.Equal(t, stapleUpdate.UTC(), nilStaple.(*ocspEntry).nextUpdate)

	toUpdate, ok := ocspStapler.cache.Get("toUpdate")
	require.True(t, ok)

	assert.Equal(t, ocspResponse, toUpdate.(*ocspEntry).staple)
	assert.Equal(t, stapleUpdate.UTC(), nilStaple.(*ocspEntry).nextUpdate)

	noUpdate, ok := ocspStapler.cache.Get("noUpdate")
	require.True(t, ok)

	assert.Equal(t, []byte("foo"), noUpdate.(*ocspEntry).staple)
	assert.Equal(t, inOneHour, noUpdate.(*ocspEntry).nextUpdate)
}

type testChain struct {
	rootCert         *x509.Certificate
	rootKey          *ecdsa.PrivateKey
	intermediateCert *x509.Certificate
	intermediateKey  *ecdsa.PrivateKey
	responderCert    *x509.Certificate
	responderKey     *ecdsa.PrivateKey
	leafCert         *x509.Certificate
	leafKey          *ecdsa.PrivateKey
}

func generateTestChain(t *testing.T, ocspServerURL string) testChain {
	t.Helper()

	rootKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	rootTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Root CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	rootDER, err := x509.CreateCertificate(rand.Reader, rootTmpl, rootTmpl, &rootKey.PublicKey, rootKey)
	require.NoError(t, err)
	rootCert, err := x509.ParseCertificate(rootDER)
	require.NoError(t, err)

	interKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	interTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "Intermediate CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	interDER, err := x509.CreateCertificate(rand.Reader, interTmpl, rootCert, &interKey.PublicKey, rootKey)
	require.NoError(t, err)
	interCert, err := x509.ParseCertificate(interDER)
	require.NoError(t, err)

	responderKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	responderTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "Delegated OCSP Responder"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageOCSPSigning},
	}
	responderDER, err := x509.CreateCertificate(rand.Reader, responderTmpl, interCert, &responderKey.PublicKey, interKey)
	require.NoError(t, err)
	responderCert, err := x509.ParseCertificate(responderDER)
	require.NoError(t, err)

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	leafTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(4),
		Subject:      pkix.Name{CommonName: "Leaf Certificate"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		OCSPServer:   []string{ocspServerURL},
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, interCert, &leafKey.PublicKey, interKey)
	require.NoError(t, err)
	leafCert, err := x509.ParseCertificate(leafDER)
	require.NoError(t, err)

	return testChain{
		rootCert:         rootCert,
		rootKey:          rootKey,
		intermediateCert: interCert,
		intermediateKey:  interKey,
		responderCert:    responderCert,
		responderKey:     responderKey,
		leafCert:         leafCert,
		leafKey:          leafKey,
	}
}

func TestFindIssuer(t *testing.T) {
	chain := generateTestChain(t, "http://ocsp.example.com")

	testCases := []struct {
		desc       string
		leaf       *x509.Certificate
		chain      [][]byte
		wantIssuer *x509.Certificate
	}{
		{
			desc:       "leaf with intermediate and root (intermediate first)",
			leaf:       chain.leafCert,
			chain:      [][]byte{chain.leafCert.Raw, chain.intermediateCert.Raw, chain.rootCert.Raw},
			wantIssuer: chain.intermediateCert,
		},
		{
			desc:       "leaf with intermediate and root (root first)",
			leaf:       chain.leafCert,
			chain:      [][]byte{chain.leafCert.Raw, chain.rootCert.Raw, chain.intermediateCert.Raw},
			wantIssuer: chain.intermediateCert,
		},
		{
			desc:       "leaf with only intermediate",
			leaf:       chain.leafCert,
			chain:      [][]byte{chain.leafCert.Raw, chain.intermediateCert.Raw},
			wantIssuer: chain.intermediateCert,
		},
		{
			desc:       "intermediate with root",
			leaf:       chain.intermediateCert,
			chain:      [][]byte{chain.intermediateCert.Raw, chain.rootCert.Raw},
			wantIssuer: chain.rootCert,
		},
		{
			desc:       "leaf alone falls back to leaf",
			leaf:       chain.leafCert,
			chain:      [][]byte{chain.leafCert.Raw},
			wantIssuer: chain.leafCert,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			issuer, err := findIssuer(tc.leaf, tc.chain)
			require.NoError(t, err)
			assert.Equal(t, tc.wantIssuer.SerialNumber, issuer.SerialNumber)
			assert.Equal(t, tc.wantIssuer.Subject.CommonName, issuer.Subject.CommonName)
		})
	}
}

func TestOCSPStapler_DelegatedResponderMultiCertChain(t *testing.T) {
	var responderURL string
	var chain testChain

	handler := func(rw http.ResponseWriter, req *http.Request) {
		ct := req.Header.Get("Content-Type")
		assert.Equal(t, "application/ocsp-request", ct)

		reqBytes, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		_, err = ocsp.ParseRequest(reqBytes)
		require.NoError(t, err)

		thisUpdate := time.Now().Truncate(time.Minute)
		nextUpdate := thisUpdate.Add(time.Hour)

		ocspResponseTmpl := ocsp.Response{
			Status:       ocsp.Good,
			SerialNumber: chain.leafCert.SerialNumber,
			ThisUpdate:   thisUpdate,
			NextUpdate:   nextUpdate,
			Certificate:  chain.responderCert,
		}

		ocspResponse, err := ocsp.CreateResponse(chain.intermediateCert, chain.responderCert, ocspResponseTmpl, chain.responderKey)
		require.NoError(t, err)

		rw.Header().Set("Content-Type", "application/ocsp-response")
		_, err = rw.Write(ocspResponse)
		require.NoError(t, err)
	}

	responder := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(responder.Close)
	responderURL = responder.URL

	chain = generateTestChain(t, responderURL)

	// Simulate multi-certificate chain where root appears before intermediate.
	rawChain := [][]byte{chain.leafCert.Raw, chain.rootCert.Raw, chain.intermediateCert.Raw}

	// findIssuer must identify intermediateCert as issuer, not leafCert or rootCert.
	resolvedIssuer, err := findIssuer(chain.leafCert, rawChain)
	require.NoError(t, err)
	assert.Equal(t, chain.intermediateCert.Subject.CommonName, resolvedIssuer.Subject.CommonName)

	ocspStapler := newOCSPStapler(nil)
	ocspStapler.client = &http.Client{Timeout: time.Second}

	// 1. With properly resolved intermediate issuer, updateStaple must succeed for delegated responder.
	entry := &ocspEntry{
		leaf:       chain.leafCert,
		issuer:     resolvedIssuer,
		responders: []string{responderURL},
	}

	err = ocspStapler.updateStaple(t.Context(), entry)
	require.NoError(t, err)
	assert.NotEmpty(t, entry.staple)

	// 2. Reproduce failure if entry.issuer is incorrectly set to rootCert.
	badRootEntry := &ocspEntry{
		leaf:       chain.leafCert,
		issuer:     chain.rootCert,
		responders: []string{responderURL},
	}
	err = ocspStapler.updateStaple(t.Context(), badRootEntry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no OCSP staple obtained from any responders")

	// 3. Reproduce failure if entry.issuer is incorrectly set to leafCert itself.
	badLeafEntry := &ocspEntry{
		leaf:       chain.leafCert,
		issuer:     chain.leafCert,
		responders: []string{responderURL},
	}
	err = ocspStapler.updateStaple(t.Context(), badLeafEntry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no OCSP staple obtained from any responders")
}
