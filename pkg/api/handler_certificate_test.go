package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/static"
	tlspkg "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

// generateTestCertificate creates a test certificate with the given parameters.
// The certificate will be valid from notBefore to notAfter.
func generateTestCertificate(commonName string, sans []string, notBefore, notAfter time.Time) (types.FileOrContent, types.FileOrContent, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
			CommonName:   commonName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Add SANs, distinguishing IP addresses from DNS names.
	for _, san := range sans {
		if ip := net.ParseIP(san); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, san)
		}
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return types.FileOrContent(certPEM), types.FileOrContent(keyPEM), nil
}

func TestHandler_Certificates(t *testing.T) {
	type expected struct {
		statusCode       int
		validateResponse func(t *testing.T, body []byte)
	}

	type certSetup struct {
		loadCerts         bool
		loadMultipleCerts bool
	}

	// Generate test certificates dynamically with valid expiration dates
	now := time.Now()

	// Certificate valid for 50+ years (status: "enabled")
	localhostCert, localhostKey, err := generateTestCertificate(
		"",
		[]string{"127.0.0.1", "::1", "example.com"},
		now.Add(-24*time.Hour),
		now.Add(50*365*24*time.Hour),
	)
	require.NoError(t, err)

	// Certificate with warning status (expires in 15 days)
	warnCert, warnKey, err := generateTestCertificate(
		"warning.com",
		[]string{"warning.com", "www.warning.com"},
		now.Add(-24*time.Hour),
		now.Add(15*24*time.Hour),
	)
	require.NoError(t, err)

	// Certificate with expired status (already expired)
	expiredCert, expiredKey, err := generateTestCertificate(
		"expired.com",
		[]string{"expired.com"},
		now.Add(-365*24*time.Hour),
		now.Add(-24*time.Hour),
	)
	require.NoError(t, err)

	// Certificate for search testing (different common name / SANs)
	acmeCert, acmeKey, err := generateTestCertificate(
		"acme.example.org",
		[]string{"acme.example.org", "api.acme.example.org"},
		now.Add(-24*time.Hour),
		now.Add(50*365*24*time.Hour),
	)
	require.NoError(t, err)

	// Compute fingerprint from the generated localhost cert PEM
	block, _ := pem.Decode([]byte(localhostCert))
	parsed, _ := x509.ParseCertificate(block.Bytes)
	hash := sha256.Sum256(parsed.Raw)
	localhostFingerprint := hex.EncodeToString(hash[:])

	testCases := []struct {
		desc     string
		path     string
		setup    certSetup
		expected expected
	}{
		{
			desc:  "all certificates, but no certificates loaded",
			path:  "/api/certificates",
			setup: certSetup{loadCerts: false},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					assert.Empty(t, certs)
				},
			},
		},
		{
			desc:  "all certificates, with one certificate loaded",
			path:  "/api/certificates",
			setup: certSetup{loadCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 1)

					cert := certs[0]
					assert.Regexp(t, `^[0-9a-f]{64}$`, cert["name"])
					assert.Equal(t, "Acme Co", cert["issuerOrg"])
					assert.Equal(t, "enabled", cert["status"])
					assert.ElementsMatch(t, []any{"127.0.0.1", "::1", "example.com"}, cert["sans"])
				},
			},
		},
		{
			desc:  "certificates filtered by search text - example",
			path:  "/api/certificates?search=example",
			setup: certSetup{loadCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 1)
					sans := certs[0]["sans"].([]any)
					assert.Contains(t, sans, "example.com")
				},
			},
		},
		{
			desc:  "certificates filtered by search text - no match",
			path:  "/api/certificates?search=nonexistent",
			setup: certSetup{loadCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					assert.Empty(t, certs)
				},
			},
		},
		{
			desc:  "certificates sorted by status",
			path:  "/api/certificates?sortBy=status",
			setup: certSetup{loadCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 1)
					assert.Equal(t, "enabled", certs[0]["status"])
				},
			},
		},
		{
			desc:  "certificates sorted by validUntil descending",
			path:  "/api/certificates?sortBy=validUntil&direction=desc",
			setup: certSetup{loadCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 1)
				},
			},
		},
		{
			desc:  "certificates filtered by status - enabled",
			path:  "/api/certificates?status=enabled",
			setup: certSetup{loadCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 1)
					assert.Equal(t, "enabled", certs[0]["status"])
				},
			},
		},
		{
			desc:  "certificates filtered by status - expired",
			path:  "/api/certificates?status=expired",
			setup: certSetup{loadCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					assert.Empty(t, certs)
				},
			},
		},
		{
			desc:  "one certificate by fingerprint",
			path:  "/api/certificates/" + localhostFingerprint,
			setup: certSetup{loadCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var cert map[string]any
					require.NoError(t, json.Unmarshal(body, &cert))
					assert.Regexp(t, `^[0-9a-f]{64}$`, cert["name"])
					assert.Equal(t, "enabled", cert["status"])
					assert.ElementsMatch(t, []any{"127.0.0.1", "::1", "example.com"}, cert["sans"])
				},
			},
		},
		{
			desc:  "certificate does not exist",
			path:  "/api/certificates/non-existent-certificate",
			setup: certSetup{loadCerts: false},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc:  "multiple certificates with different statuses",
			path:  "/api/certificates",
			setup: certSetup{loadMultipleCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 4)

					// Verify all statuses are present
					statuses := make(map[string]int)
					for _, cert := range certs {
						status := cert["status"].(string)
						statuses[status]++
					}
					assert.Equal(t, 2, statuses["enabled"])
					assert.Equal(t, 1, statuses["warning"])
					assert.Equal(t, 1, statuses["expired"])
				},
			},
		},
		{
			desc:  "certificates sorted by name ascending",
			path:  "/api/certificates?sortBy=name&direction=asc",
			setup: certSetup{loadMultipleCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 4)

					// Verify names are in ascending order
					prevName := ""
					for _, cert := range certs {
						commonName := cert["commonName"].(string)
						if prevName != "" {
							assert.LessOrEqual(t, prevName, commonName)
						}
						prevName = commonName
					}
				},
			},
		},
		{
			desc:  "certificates sorted by name descending",
			path:  "/api/certificates?sortBy=name&direction=desc",
			setup: certSetup{loadMultipleCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 4)

					// Verify names are in descending order
					prevName := "zzzzzzz"
					for _, cert := range certs {
						commonName := cert["commonName"].(string)
						assert.GreaterOrEqual(t, prevName, commonName)
						prevName = commonName
					}
				},
			},
		},
		{
			desc:  "certificates sorted by status ascending",
			path:  "/api/certificates?sortBy=status&direction=asc",
			setup: certSetup{loadMultipleCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 4)

					// Verify statuses are in ascending order (enabled < expired < warning)
					prevStatus := ""
					for _, cert := range certs {
						status := cert["status"].(string)
						if prevStatus != "" {
							assert.LessOrEqual(t, prevStatus, status)
						}
						prevStatus = status
					}
				},
			},
		},
		{
			desc:  "certificates sorted by issuer",
			path:  "/api/certificates?sortBy=issuer",
			setup: certSetup{loadMultipleCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 4)

					// All certificates have same issuer "Acme Co"
					for _, cert := range certs {
						assert.Equal(t, "Acme Co", cert["issuerOrg"])
					}
				},
			},
		},
		{
			desc:  "certificates sorted by validUntil ascending",
			path:  "/api/certificates?sortBy=validUntil&direction=asc",
			setup: certSetup{loadMultipleCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 4)

					// Verify notAfter dates are in ascending order
					var prevTime time.Time
					for _, cert := range certs {
						notAfter := cert["notAfter"].(string)
						certTime, err := time.Parse(time.RFC3339, notAfter)
						require.NoError(t, err)
						if !prevTime.IsZero() {
							assert.False(t, certTime.Before(prevTime))
						}
						prevTime = certTime
					}
				},
			},
		},
		{
			desc:  "certificates filtered by status - warning",
			path:  "/api/certificates?status=warning",
			setup: certSetup{loadMultipleCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 1)
					assert.Equal(t, "warning", certs[0]["status"])
					assert.Contains(t, certs[0]["sans"], "warning.com")
				},
			},
		},
		{
			desc:  "certificates filtered by status - expired",
			path:  "/api/certificates?status=expired",
			setup: certSetup{loadMultipleCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 1)
					assert.Equal(t, "expired", certs[0]["status"])
					assert.Contains(t, certs[0]["sans"], "expired.com")
				},
			},
		},
		{
			desc:  "certificates filtered by search - commonName",
			path:  "/api/certificates?search=acme.example.org",
			setup: certSetup{loadMultipleCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 1)
					assert.Equal(t, "acme.example.org", certs[0]["commonName"])
				},
			},
		},
		{
			desc:  "certificates filtered by search - issuerOrg",
			path:  "/api/certificates?search=Acme",
			setup: certSetup{loadMultipleCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					// All certificates have "Acme Co" as issuer
					require.Len(t, certs, 4)
				},
			},
		},
		{
			desc:  "certificates with comprehensive field validation",
			path:  "/api/certificates",
			setup: certSetup{loadMultipleCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					t.Helper()
					var certs []map[string]any
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 4)

					// Check the certificate with commonName set (warning.com)
					var certWithCN map[string]any
					for _, c := range certs {
						if c["commonName"] == "warning.com" {
							certWithCN = c
							break
						}
					}
					require.NotNil(t, certWithCN, "Should find certificate with commonName")

					// Validate all expected fields are present
					assert.NotEmpty(t, certWithCN["name"])
					assert.NotEmpty(t, certWithCN["sans"])
					assert.NotEmpty(t, certWithCN["notAfter"])
					assert.NotEmpty(t, certWithCN["notBefore"])
					assert.NotEmpty(t, certWithCN["serialNumber"])
					assert.Equal(t, "warning.com", certWithCN["commonName"])
					assert.NotEmpty(t, certWithCN["issuerOrg"])
					assert.NotEmpty(t, certWithCN["version"])
					assert.Equal(t, "RSA", certWithCN["keyType"])
					assert.InDelta(t, float64(2048), certWithCN["keySize"], 0)
					assert.NotEmpty(t, certWithCN["signatureAlgorithm"])
					assert.NotEmpty(t, certWithCN["certFingerprint"])
					assert.NotEmpty(t, certWithCN["publicKeyFingerprint"])
					assert.Equal(t, "warning", certWithCN["status"])
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			tlsManager := tlspkg.NewManager(nil)

			if test.setup.loadCerts {
				dynamicConfigs := []*tlspkg.CertAndStores{{
					Certificate: tlspkg.Certificate{
						CertFile: localhostCert,
						KeyFile:  localhostKey,
					},
				}}

				tlsManager.UpdateConfigs(t.Context(), nil, nil, dynamicConfigs)
			}

			if test.setup.loadMultipleCerts {
				dynamicConfigs := []*tlspkg.CertAndStores{
					{
						Certificate: tlspkg.Certificate{
							CertFile: localhostCert,
							KeyFile:  localhostKey,
						},
					},
					{
						Certificate: tlspkg.Certificate{
							CertFile: warnCert,
							KeyFile:  warnKey,
						},
					},
					{
						Certificate: tlspkg.Certificate{
							CertFile: expiredCert,
							KeyFile:  expiredKey,
						},
					},
					{
						Certificate: tlspkg.Certificate{
							CertFile: acmeCert,
							KeyFile:  acmeKey,
						},
					},
				}

				tlsManager.UpdateConfigs(t.Context(), nil, nil, dynamicConfigs)
			}

			handler := New(static.Configuration{API: &static.API{}, Global: &static.Global{}}, nil).WithTLSManager(tlsManager)
			server := httptest.NewServer(handler.createRouter())

			resp, err := http.DefaultClient.Get(server.URL + test.path)
			require.NoError(t, err)

			require.Equal(t, test.expected.statusCode, resp.StatusCode)

			contents, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			err = resp.Body.Close()
			require.NoError(t, err)

			// Only validate content type and body for success responses
			if resp.StatusCode == http.StatusOK {
				assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
				if test.expected.validateResponse != nil {
					test.expected.validateResponse(t, contents)
				}
			}
		})
	}
}
