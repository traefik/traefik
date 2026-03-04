package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ECHRequestConfig is a configuration struct for making ECH-enabled requests.
// This is only used for testing purposes.
type ECHRequestConfig[T []byte | string] struct {
	URL      string
	Host     string
	ECH      T
	Insecure bool
}

// RequestWithECH sends a GET request to a server using the provided ECH configuration.
// This is only used for testing purposes.
func RequestWithECH[T []byte | string](c ECHRequestConfig[T]) (body []byte, err error) {
	// Decode the ECH configuration from base64 if it's a string, otherwise use it directly.
	var ech []byte
	if s, ok := any(c.ECH).(string); ok {
		ech, err = base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, err
		}
	} else {
		ech = []byte(c.ECH)
	}

	requestURL, err := url.Parse(c.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if c.Host == "" {
		c.Host = requestURL.Hostname()
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName:                     c.Host,
				EncryptedClientHelloConfigList: ech,
				MinVersion:                     tls.VersionTLS13,
				InsecureSkipVerify:             c.Insecure,
			},
		},
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    requestURL,
		Host:   c.Host,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func TestNewECHKey(t *testing.T) {
	testCases := []struct {
		desc        string
		publicName  string
		expectError bool
	}{
		{
			desc:       "valid short public name",
			publicName: "server.local",
		},
		{
			desc:       "valid public name at max length",
			publicName: "abcdefghijklmnopqrstuvwxyz012345", // 32 chars
		},
		{
			desc:        "public name exceeds max length",
			publicName:  "abcdefghijklmnopqrstuvwxyz0123456", // 33 chars
			expectError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			echKey, err := NewECHKey(test.publicName)

			if test.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, echKey.Config)
			assert.NotEmpty(t, echKey.PrivateKey)
			assert.True(t, echKey.SendAsRetry)
		})
	}
}

func TestMarshalUnmarshalECHKey(t *testing.T) {
	testCases := []struct {
		desc       string
		publicName string
	}{
		{
			desc:       "standard domain",
			publicName: "server.local",
		},
		{
			desc:       "subdomain",
			publicName: "api.example.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			echKey, err := NewECHKey(test.publicName)
			require.NoError(t, err)

			echKeyBytes, err := MarshalECHKey(echKey)
			require.NoError(t, err)
			assert.NotEmpty(t, echKeyBytes)

			newKey, err := UnmarshalECHKey(echKeyBytes)
			require.NoError(t, err)

			assert.Equal(t, echKey.Config, newKey.Config)
			assert.Equal(t, echKey.PrivateKey, newKey.PrivateKey)
			assert.True(t, newKey.SendAsRetry)
		})
	}
}

func TestMarshalECHKey_Errors(t *testing.T) {
	testCases := []struct {
		desc string
		key  *tls.EncryptedClientHelloKey
	}{
		{
			desc: "missing config",
			key: &tls.EncryptedClientHelloKey{
				PrivateKey: []byte("some-key"),
			},
		},
		{
			desc: "missing private key",
			key: &tls.EncryptedClientHelloKey{
				Config: []byte("some-config"),
			},
		},
		{
			desc: "both missing",
			key:  &tls.EncryptedClientHelloKey{},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := MarshalECHKey(test.key)
			require.Error(t, err)
		})
	}
}

func TestUnmarshalECHKey_Errors(t *testing.T) {
	testCases := []struct {
		desc string
		data []byte
	}{
		{
			desc: "empty data",
			data: []byte{},
		},
		{
			desc: "invalid PEM",
			data: []byte("not a valid PEM"),
		},
		{
			desc: "unknown PEM block type",
			data: pem.EncodeToMemory(&pem.Block{Type: "UNKNOWN", Bytes: []byte("data")}),
		},
		{
			desc: "missing private key",
			data: func() []byte {
				// Create ECHCONFIG block with length prefix
				configBytes := append([]byte{0, 4}, []byte("test")...)
				return pem.EncodeToMemory(&pem.Block{Type: "ECHCONFIG", Bytes: configBytes})
			}(),
		},
		{
			desc: "missing config",
			data: pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: make([]byte, 32)}),
		},
		{
			desc: "private key too short",
			data: func() []byte {
				var pemData []byte
				pemData = append(pemData, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: make([]byte, 16)})...)
				configBytes := append([]byte{0, 4}, []byte("test")...)
				pemData = append(pemData, pem.EncodeToMemory(&pem.Block{Type: "ECHCONFIG", Bytes: configBytes})...)
				return pemData
			}(),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := UnmarshalECHKey(test.data)
			require.Error(t, err)
		})
	}
}

func TestECHConfigToConfigList(t *testing.T) {
	testCases := []struct {
		desc   string
		config []byte
	}{
		{
			desc:   "empty config",
			config: []byte{},
		},
		{
			desc:   "simple config",
			config: []byte{0x01, 0x02, 0x03, 0x04},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			configList, err := ECHConfigToConfigList(test.config)
			require.NoError(t, err)

			// Config list should have 2-byte length prefix followed by the config
			expectedLen := 2 + len(test.config)
			assert.Len(t, configList, expectedLen)
		})
	}
}

func TestRequestWithECH(t *testing.T) {
	const commonName = "server.local"

	echKey, err := NewECHKey(commonName)
	require.NoError(t, err)

	testCert, err := generateTestCert(commonName)
	require.NoError(t, err)

	echConfigList, err := ECHConfigToConfigList(echKey.Config)
	require.NoError(t, err)

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, ECH-enabled TLS server!")
	}))

	server.TLS = &tls.Config{
		Certificates:             []tls.Certificate{testCert},
		MinVersion:               tls.VersionTLS13,
		EncryptedClientHelloKeys: []tls.EncryptedClientHelloKey{*echKey},
	}

	server.StartTLS()
	t.Cleanup(server.Close)

	testCases := []struct {
		desc         string
		config       ECHRequestConfig[[]byte]
		expectedBody string
		expectError  bool
	}{
		{
			desc: "successful ECH request with bytes",
			config: ECHRequestConfig[[]byte]{
				URL:      server.URL + "/",
				Host:     commonName,
				ECH:      echConfigList,
				Insecure: true,
			},
			expectedBody: "Hello, ECH-enabled TLS server!",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			response, err := RequestWithECH(test.config)

			if test.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectedBody, string(response))
		})
	}
}

func TestRequestWithECH_StringConfig(t *testing.T) {
	const commonName = "server.local"

	echKey, err := NewECHKey(commonName)
	require.NoError(t, err)

	testCert, err := generateTestCert(commonName)
	require.NoError(t, err)

	echConfigList, err := ECHConfigToConfigList(echKey.Config)
	require.NoError(t, err)

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello from string config!")
	}))

	server.TLS = &tls.Config{
		Certificates:             []tls.Certificate{testCert},
		MinVersion:               tls.VersionTLS13,
		EncryptedClientHelloKeys: []tls.EncryptedClientHelloKey{*echKey},
	}

	server.StartTLS()
	t.Cleanup(server.Close)

	// Test with base64-encoded string config
	echConfigBase64 := base64.StdEncoding.EncodeToString(echConfigList)

	response, err := RequestWithECH(ECHRequestConfig[string]{
		URL:      server.URL + "/",
		Host:     commonName,
		ECH:      echConfigBase64,
		Insecure: true,
	})

	require.NoError(t, err)
	assert.Equal(t, "Hello from string config!", string(response))
}

func TestRequestWithECH_Errors(t *testing.T) {
	testCases := []struct {
		desc   string
		config ECHRequestConfig[string]
	}{
		{
			desc: "invalid base64 ECH config",
			config: ECHRequestConfig[string]{
				URL: "https://localhost:12345/",
				ECH: "not-valid-base64!!!",
			},
		},
		{
			desc: "invalid URL",
			config: ECHRequestConfig[string]{
				URL: "://invalid-url",
				ECH: "dGVzdA==",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := RequestWithECH(test.config)
			require.Error(t, err)
		})
	}
}

func generateTestCert(commonName string) (tls.Certificate, error) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(rsaKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})

	notBefore := time.Now()
	notAfter := notBefore.Add(24 * time.Hour)

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: commonName},
		DNSNames:              []string{commonName, "localhost"},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &rsaKey.PublicKey, rsaKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	return tls.X509KeyPair(certPEM, keyPEM)
}
