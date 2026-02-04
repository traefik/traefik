package api

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/static"
	tlspkg "github.com/traefik/traefik/v3/pkg/tls"
)

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

	// Certificate with disabled status (already expired)
	expiredCert, expiredKey, err := generateTestCertificate(
		"expired.com",
		[]string{"expired.com"},
		now.Add(-365*24*time.Hour),
		now.Add(-24*time.Hour),
	)
	require.NoError(t, err)

	// Certificate for search testing (different issuer)
	acmeCert, acmeKey, err := generateTestCertificate(
		"acme.example.org",
		[]string{"acme.example.org", "api.acme.example.org"},
		now.Add(-24*time.Hour),
		now.Add(50*365*24*time.Hour),
	)
	require.NoError(t, err)

	// Create a valid base64-encoded certKey for a non-existent certificate
	// "nonexistent.com" -> base64("nonexistent.com")
	nonExistentCertKey := base64.URLEncoding.EncodeToString([]byte("nonexistent.com"))

	// Create certKey for the localhost certificate
	// The certificate has domains: 127.0.0.1,::1,example.com (sorted)
	// certKey is base64("127.0.0.1,::1,example.com")
	localhostCertKey := base64.URLEncoding.EncodeToString([]byte("127.0.0.1,::1,example.com"))

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
					var certs []map[string]interface{}
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
					var certs []map[string]interface{}
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 1)

					cert := certs[0]
					assert.Equal(t, "MTI3LjAuMC4xLDo6MSxleGFtcGxlLmNvbQ==", cert["name"])
					assert.Equal(t, "Acme Co", cert["issuerOrg"])
					assert.Equal(t, "enabled", cert["status"])
					assert.ElementsMatch(t, []interface{}{"127.0.0.1", "::1", "example.com"}, cert["sans"])
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
					var certs []map[string]interface{}
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 1)
					sans := certs[0]["sans"].([]interface{})
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
					var certs []map[string]interface{}
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
					var certs []map[string]interface{}
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
					var certs []map[string]interface{}
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
					var certs []map[string]interface{}
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 1)
					assert.Equal(t, "enabled", certs[0]["status"])
				},
			},
		},
		{
			desc:  "certificates filtered by status - disabled",
			path:  "/api/certificates?status=disabled",
			setup: certSetup{loadCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					var certs []map[string]interface{}
					require.NoError(t, json.Unmarshal(body, &certs))
					assert.Empty(t, certs)
				},
			},
		},
		{
			desc:  "one certificate by certKey",
			path:  "/api/certificates/" + localhostCertKey,
			setup: certSetup{loadCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					var cert map[string]interface{}
					require.NoError(t, json.Unmarshal(body, &cert))
					assert.Equal(t, "MTI3LjAuMC4xLDo6MSxleGFtcGxlLmNvbQ==", cert["name"])
					assert.Equal(t, "enabled", cert["status"])
					assert.ElementsMatch(t, []interface{}{"127.0.0.1", "::1", "example.com"}, cert["sans"])
				},
			},
		},
		{
			desc:  "invalid certKey - not valid base64",
			path:  "/api/certificates/not-valid-base64!@#$",
			setup: certSetup{loadCerts: false},
			expected: expected{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			desc:  "valid certKey but certificate does not exist",
			path:  "/api/certificates/" + nonExistentCertKey,
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
					var certs []map[string]interface{}
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
					assert.Equal(t, 1, statuses["disabled"])
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
					var certs []map[string]interface{}
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
					var certs []map[string]interface{}
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
					var certs []map[string]interface{}
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 4)

					// Verify statuses are in ascending order (disabled < enabled < warning)
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
					var certs []map[string]interface{}
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
					var certs []map[string]interface{}
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 4)

					// Verify notAfter dates are in ascending order
					var prevTime time.Time
					for _, cert := range certs {
						notAfter := cert["notAfter"].(string)
						certTime, err := time.Parse(time.RFC3339, notAfter)
						require.NoError(t, err)
						if !prevTime.IsZero() {
							assert.True(t, !certTime.Before(prevTime))
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
					var certs []map[string]interface{}
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 1)
					assert.Equal(t, "warning", certs[0]["status"])
					assert.Contains(t, certs[0]["sans"], "warning.com")
				},
			},
		},
		{
			desc:  "certificates filtered by status - disabled (expired)",
			path:  "/api/certificates?status=disabled",
			setup: certSetup{loadMultipleCerts: true},
			expected: expected{
				statusCode: http.StatusOK,
				validateResponse: func(t *testing.T, body []byte) {
					var certs []map[string]interface{}
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 1)
					assert.Equal(t, "disabled", certs[0]["status"])
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
					var certs []map[string]interface{}
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
					var certs []map[string]interface{}
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
					var certs []map[string]interface{}
					require.NoError(t, json.Unmarshal(body, &certs))
					require.Len(t, certs, 4)

					// Check the certificate with commonName set (warning.com)
					var certWithCN map[string]interface{}
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
					assert.Equal(t, float64(2048), certWithCN["keySize"])
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

			handler := New(static.Configuration{API: &static.API{}, Global: &static.Global{}}, nil, tlsManager)
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
