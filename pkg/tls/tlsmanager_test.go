package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LocalhostCert is a PEM-encoded TLS cert with SAN IPs
// "127.0.0.1" and "[::1]", expiring at Jan 29 16:00:00 2084 GMT.
// generated from src/crypto/tls:
// go run generate_cert.go  --rsa-bits 1024 --host 127.0.0.1,::1,example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var (
	localhostCert = FileOrContent(`-----BEGIN CERTIFICATE-----
MIIDOTCCAiGgAwIBAgIQSRJrEpBGFc7tNb1fb5pKFzANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEA6Gba5tHV1dAKouAaXO3/ebDUU4rvwCUg/CNaJ2PT5xLD4N1Vcb8r
bFSW2HXKq+MPfVdwIKR/1DczEoAGf/JWQTW7EgzlXrCd3rlajEX2D73faWJekD0U
aUgz5vtrTXZ90BQL7WvRICd7FlEZ6FPOcPlumiyNmzUqtwGhO+9ad1W5BqJaRI6P
YfouNkwR6Na4TzSj5BrqUfP0FwDizKSJ0XXmh8g8G9mtwxOSN3Ru1QFc61Xyeluk
POGKBV/q6RBNklTNe0gI8usUMlYyoC7ytppNMW7X2vodAelSu25jgx2anj9fDVZu
h7AXF5+4nJS4AAt0n1lNY7nGSsdZas8PbQIDAQABo4GIMIGFMA4GA1UdDwEB/wQE
AwICpDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MB0GA1Ud
DgQWBBStsdjh3/JCXXYlQryOrL4Sh7BW5TAuBgNVHREEJzAlggtleGFtcGxlLmNv
bYcEfwAAAYcQAAAAAAAAAAAAAAAAAAAAATANBgkqhkiG9w0BAQsFAAOCAQEAxWGI
5NhpF3nwwy/4yB4i/CwwSpLrWUa70NyhvprUBC50PxiXav1TeDzwzLx/o5HyNwsv
cxv3HdkLW59i/0SlJSrNnWdfZ19oTcS+6PtLoVyISgtyN6DpkKpdG1cOkW3Cy2P2
+tK/tKHRP1Y/Ra0RiDpOAmqn0gCOFGz8+lqDIor/T7MTpibL3IxqWfPrvfVRHL3B
grw/ZQTTIVjjh4JBSW3WyWgNo/ikC1lrVxzl4iPUGptxT36Cr7Zk2Bsg0XqwbOvK
5d+NTDREkSnUbie4GeutujmX3Dsx88UiV6UY/4lHJa6I5leHUNOHahRbpbWeOfs/
WkBKOclmOV2xlTVuPw==
-----END CERTIFICATE-----`)

	// LocalhostKey is the private key for localhostCert.
	localhostKey = FileOrContent(`-----BEGIN RSA PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDoZtrm0dXV0Aqi
4Bpc7f95sNRTiu/AJSD8I1onY9PnEsPg3VVxvytsVJbYdcqr4w99V3AgpH/UNzMS
gAZ/8lZBNbsSDOVesJ3euVqMRfYPvd9pYl6QPRRpSDPm+2tNdn3QFAvta9EgJ3sW
URnoU85w+W6aLI2bNSq3AaE771p3VbkGolpEjo9h+i42TBHo1rhPNKPkGupR8/QX
AOLMpInRdeaHyDwb2a3DE5I3dG7VAVzrVfJ6W6Q84YoFX+rpEE2SVM17SAjy6xQy
VjKgLvK2mk0xbtfa+h0B6VK7bmODHZqeP18NVm6HsBcXn7iclLgAC3SfWU1jucZK
x1lqzw9tAgMBAAECggEABWzxS1Y2wckblnXY57Z+sl6YdmLV+gxj2r8Qib7g4ZIk
lIlWR1OJNfw7kU4eryib4fc6nOh6O4AWZyYqAK6tqNQSS/eVG0LQTLTTEldHyVJL
dvBe+MsUQOj4nTndZW+QvFzbcm2D8lY5n2nBSxU5ypVoKZ1EqQzytFcLZpTN7d89
EPj0qDyrV4NZlWAwL1AygCwnlwhMQjXEalVF1ylXwU3QzyZ/6MgvF6d3SSUlh+sq
XefuyigXw484cQQgbzopv6niMOmGP3of+yV4JQqUSb3IDmmT68XjGd2Dkxl4iPki
6ZwXf3CCi+c+i/zVEcufgZ3SLf8D99kUGE7v7fZ6AQKBgQD1ZX3RAla9hIhxCf+O
3D+I1j2LMrdjAh0ZKKqwMR4JnHX3mjQI6LwqIctPWTU8wYFECSh9klEclSdCa64s
uI/GNpcqPXejd0cAAdqHEEeG5sHMDt0oFSurL4lyud0GtZvwlzLuwEweuDtvT9cJ
Wfvl86uyO36IW8JdvUprYDctrQKBgQDycZ697qutBieZlGkHpnYWUAeImVA878sJ
w44NuXHvMxBPz+lbJGAg8Cn8fcxNAPqHIraK+kx3po8cZGQywKHUWsxi23ozHoxo
+bGqeQb9U661TnfdDspIXia+xilZt3mm5BPzOUuRqlh4Y9SOBpSWRmEhyw76w4ZP
OPxjWYAgwQKBgA/FehSYxeJgRjSdo+MWnK66tjHgDJE8bYpUZsP0JC4R9DL5oiaA
brd2fI6Y+SbyeNBallObt8LSgzdtnEAbjIH8uDJqyOmknNePRvAvR6mP4xyuR+Bv
m+Lgp0DMWTw5J9CKpydZDItc49T/mJ5tPhdFVd+am0NAQnmr1MCZ6nHxAoGABS3Y
LkaC9FdFUUqSU8+Chkd/YbOkuyiENdkvl6t2e52jo5DVc1T7mLiIrRQi4SI8N9bN
/3oJWCT+uaSLX2ouCtNFunblzWHBrhxnZzTeqVq4SLc8aESAnbslKL4i8/+vYZlN
s8xtiNcSvL+lMsOBORSXzpj/4Ot8WwTkn1qyGgECgYBKNTypzAHeLE6yVadFp3nQ
Ckq9yzvP/ib05rvgbvrne00YeOxqJ9gtTrzgh7koqJyX1L4NwdkEza4ilDWpucn0
xiUZS4SoaJq6ZvcBYS62Yr1t8n09iG47YL8ibgtmH3L+svaotvpVxVK+d7BLevA/
ZboOWVe3icTy64BT3OQhmg==
-----END RSA PRIVATE KEY-----`)
)

func TestTLSInStore(t *testing.T) {
	dynamicConfigs := []*CertAndStores{{
		Certificate: Certificate{
			CertFile: localhostCert,
			KeyFile:  localhostKey,
		},
	}}

	tlsManager := NewManager()
	tlsManager.UpdateConfigs(context.Background(), nil, nil, dynamicConfigs)

	certs := tlsManager.GetStore("default").DynamicCerts.Get().(map[string]*tls.Certificate)
	if len(certs) == 0 {
		t.Fatal("got error: default store must have TLS certificates.")
	}
}

func TestTLSInvalidStore(t *testing.T) {
	dynamicConfigs := []*CertAndStores{{
		Certificate: Certificate{
			CertFile: localhostCert,
			KeyFile:  localhostKey,
		},
	}}

	tlsManager := NewManager()
	tlsManager.UpdateConfigs(context.Background(),
		map[string]Store{
			"default": {
				DefaultCertificate: &Certificate{
					CertFile: "/wrong",
					KeyFile:  "/wrong",
				},
			},
		}, nil, dynamicConfigs)

	certs := tlsManager.GetStore("default").DynamicCerts.Get().(map[string]*tls.Certificate)
	if len(certs) == 0 {
		t.Fatal("got error: default store must have TLS certificates.")
	}
}

func TestManager_Get(t *testing.T) {
	dynamicConfigs := []*CertAndStores{{
		Certificate: Certificate{
			CertFile: localhostCert,
			KeyFile:  localhostKey,
		},
	}}

	tlsConfigs := map[string]Options{
		"foo":     {MinVersion: "VersionTLS12"},
		"bar":     {MinVersion: "VersionTLS11"},
		"invalid": {CurvePreferences: []string{"42"}},
	}

	testCases := []struct {
		desc               string
		tlsOptionsName     string
		expectedMinVersion uint16
		expectedError      bool
	}{
		{
			desc:               "Get a tls config from a valid name",
			tlsOptionsName:     "foo",
			expectedMinVersion: uint16(tls.VersionTLS12),
		},
		{
			desc:               "Get another tls config from a valid name",
			tlsOptionsName:     "bar",
			expectedMinVersion: uint16(tls.VersionTLS11),
		},
		{
			desc:           "Get a tls config from an invalid name",
			tlsOptionsName: "unknown",
			expectedError:  true,
		},
		{
			desc:           "Get a tls config from unexisting 'default' name",
			tlsOptionsName: "default",
			expectedError:  true,
		},
		{
			desc:           "Get an invalid tls config",
			tlsOptionsName: "invalid",
			expectedError:  true,
		},
	}

	tlsManager := NewManager()
	tlsManager.UpdateConfigs(context.Background(), nil, tlsConfigs, dynamicConfigs)

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			config, err := tlsManager.Get("default", test.tlsOptionsName)
			if test.expectedError {
				require.Nil(t, config)
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, config.MinVersion, test.expectedMinVersion)
		})
	}
}

func TestClientAuth(t *testing.T) {
	tlsConfigs := map[string]Options{
		"eca": {
			ClientAuth: ClientAuth{},
		},
		"ecat": {
			ClientAuth: ClientAuth{ClientAuthType: ""},
		},
		"ncc": {
			ClientAuth: ClientAuth{ClientAuthType: "NoClientCert"},
		},
		"rcc": {
			ClientAuth: ClientAuth{ClientAuthType: "RequestClientCert"},
		},
		"racc": {
			ClientAuth: ClientAuth{ClientAuthType: "RequireAnyClientCert"},
		},
		"vccig": {
			ClientAuth: ClientAuth{
				CAFiles:        []FileOrContent{localhostCert},
				ClientAuthType: "VerifyClientCertIfGiven",
			},
		},
		"vccigwca": {
			ClientAuth: ClientAuth{ClientAuthType: "VerifyClientCertIfGiven"},
		},
		"ravcc": {
			ClientAuth: ClientAuth{ClientAuthType: "RequireAndVerifyClientCert"},
		},
		"ravccwca": {
			ClientAuth: ClientAuth{
				CAFiles:        []FileOrContent{localhostCert},
				ClientAuthType: "RequireAndVerifyClientCert",
			},
		},
		"ravccwbca": {
			ClientAuth: ClientAuth{
				CAFiles:        []FileOrContent{"Bad content"},
				ClientAuthType: "RequireAndVerifyClientCert",
			},
		},
		"ucat": {
			ClientAuth: ClientAuth{ClientAuthType: "Unknown"},
		},
	}

	block, _ := pem.Decode([]byte(localhostCert))
	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	testCases := []struct {
		desc               string
		tlsOptionsName     string
		expectedClientAuth tls.ClientAuthType
		expectedRawSubject []byte
		expectedError      bool
	}{
		{
			desc:               "Empty ClientAuth option should get a tls.NoClientCert (default value)",
			tlsOptionsName:     "eca",
			expectedClientAuth: tls.NoClientCert,
		},
		{
			desc:               "Empty ClientAuthType option should get a tls.NoClientCert (default value)",
			tlsOptionsName:     "ecat",
			expectedClientAuth: tls.NoClientCert,
		},
		{
			desc:               "NoClientCert option should get a tls.NoClientCert as ClientAuthType",
			tlsOptionsName:     "ncc",
			expectedClientAuth: tls.NoClientCert,
		},
		{
			desc:               "RequestClientCert option should get a tls.RequestClientCert as ClientAuthType",
			tlsOptionsName:     "rcc",
			expectedClientAuth: tls.RequestClientCert,
		},
		{
			desc:               "RequireAnyClientCert option should get a tls.RequireAnyClientCert as ClientAuthType",
			tlsOptionsName:     "racc",
			expectedClientAuth: tls.RequireAnyClientCert,
		},
		{
			desc:               "VerifyClientCertIfGiven option should get a tls.VerifyClientCertIfGiven as ClientAuthType",
			tlsOptionsName:     "vccig",
			expectedClientAuth: tls.VerifyClientCertIfGiven,
		},
		{
			desc:               "VerifyClientCertIfGiven option without CAFiles yields a default ClientAuthType (NoClientCert)",
			tlsOptionsName:     "vccigwca",
			expectedClientAuth: tls.NoClientCert,
			expectedError:      true,
		},
		{
			desc:               "RequireAndVerifyClientCert option without CAFiles yields a default ClientAuthType (NoClientCert)",
			tlsOptionsName:     "ravcc",
			expectedClientAuth: tls.NoClientCert,
			expectedError:      true,
		},
		{
			desc:               "RequireAndVerifyClientCert option should get a tls.RequireAndVerifyClientCert as ClientAuthType with CA files",
			tlsOptionsName:     "ravccwca",
			expectedClientAuth: tls.RequireAndVerifyClientCert,
			expectedRawSubject: cert.RawSubject,
		},
		{
			desc:               "Unknown option yields a default ClientAuthType (NoClientCert)",
			tlsOptionsName:     "ucat",
			expectedClientAuth: tls.NoClientCert,
			expectedError:      true,
		},
		{
			desc:               "Bad CA certificate content yields a default ClientAuthType (NoClientCert)",
			tlsOptionsName:     "ravccwbca",
			expectedClientAuth: tls.NoClientCert,
			expectedError:      true,
		},
	}

	tlsManager := NewManager()
	tlsManager.UpdateConfigs(context.Background(), nil, tlsConfigs, nil)

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			config, err := tlsManager.Get("default", test.tlsOptionsName)

			if test.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if test.expectedRawSubject != nil {
				subjects := config.ClientCAs.Subjects()
				assert.Len(t, subjects, 1)
				assert.Equal(t, subjects[0], test.expectedRawSubject)
			}

			assert.Equal(t, config.ClientAuth, test.expectedClientAuth)
		})
	}
}

func TestManager_Get_DefaultValues(t *testing.T) {
	tlsManager := NewManager()

	// Ensures we won't break things for Traefik users when updating Go
	config, _ := tlsManager.Get("default", "default")
	assert.Equal(t, config.MinVersion, uint16(tls.VersionTLS12))
	assert.Equal(t, config.NextProtos, []string{"h2", "http/1.1", "acme-tls/1"})
	assert.Equal(t, config.CipherSuites, []uint16{
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	})
}
