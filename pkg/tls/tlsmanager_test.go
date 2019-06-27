package tls

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
)

// LocalhostCert is a PEM-encoded TLS cert with SAN IPs
// "127.0.0.1" and "[::1]", expiring at Jan 29 16:00:00 2084 GMT.
// generated from src/crypto/tls:
// go run generate_cert.go  --rsa-bits 1024 --host 127.0.0.1,::1,example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var (
	localhostCert = FileOrContent(`-----BEGIN CERTIFICATE-----
MIICEzCCAXygAwIBAgIQMIMChMLGrR+QvmQvpwAU6zANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9SjY1bIw4
iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZBl2+XsDul
rKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQABo2gwZjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAuBgNVHREEJzAlggtleGFtcGxlLmNvbYcEfwAAAYcQAAAAAAAAAAAAAAAA
AAAAATANBgkqhkiG9w0BAQsFAAOBgQCEcetwO59EWk7WiJsG4x8SY+UIAA+flUI9
tyC4lNhbcF2Idq9greZwbYCqTTTr2XiRNSMLCOjKyI7ukPoPjo16ocHj+P3vZGfs
h1fIw3cSS2OolhloGw/XM6RWPWtPAlGykKLciQrBru5NAPvCMsb/I1DAceTiotQM
fblo6RBxUQ==
-----END CERTIFICATE-----`)

	// LocalhostKey is the private key for localhostCert.
	localhostKey = FileOrContent(`-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9
SjY1bIw4iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZB
l2+XsDulrKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQAB
AoGAGRzwwir7XvBOAy5tM/uV6e+Zf6anZzus1s1Y1ClbjbE6HXbnWWF/wbZGOpet
3Zm4vD6MXc7jpTLryzTQIvVdfQbRc6+MUVeLKwZatTXtdZrhu+Jk7hx0nTPy8Jcb
uJqFk541aEw+mMogY/xEcfbWd6IOkp+4xqjlFLBEDytgbIECQQDvH/E6nk+hgN4H
qzzVtxxr397vWrjrIgPbJpQvBsafG7b0dA4AFjwVbFLmQcj2PprIMmPcQrooz8vp
jy4SHEg1AkEA/v13/5M47K9vCxmb8QeD/asydfsgS5TeuNi8DoUBEmiSJwma7FXY
fFUtxuvL7XvjwjN5B30pNEbc6Iuyt7y4MQJBAIt21su4b3sjXNueLKH85Q+phy2U
fQtuUE9txblTu14q3N7gHRZB4ZMhFYyDy8CKrN2cPg/Fvyt0Xlp/DoCzjA0CQQDU
y2ptGsuSmgUtWj3NM9xuwYPm+Z/F84K6+ARYiZ6PYj013sovGKUFfYAqVXVlxtIX
qyUBnu3X9ps8ZfjLZO7BAkEAlT4R5Yl6cGhaJQYZHOde3JEMhNRcVFMO8dJDaFeo
f9Oeos0UUothgiDktdQHxdNEwLjQf7lJJBzV+5OtwswCWA==
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
	tlsManager.UpdateConfigs(nil, nil, dynamicConfigs)

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
	tlsManager.UpdateConfigs(map[string]Store{
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
		"foo": {MinVersion: "VersionTLS12"},
		"bar": {MinVersion: "VersionTLS11"},
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
			desc:           "Get an tls config from an invalid name",
			tlsOptionsName: "unknown",
			expectedError:  true,
		},
		{
			desc:           "Get an tls config from unexisting 'default' name",
			tlsOptionsName: "default",
			expectedError:  true,
		},
	}

	tlsManager := NewManager()
	tlsManager.UpdateConfigs(nil, tlsConfigs, dynamicConfigs)

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
			assert.Equal(t, config.MinVersion, test.expectedMinVersion)
		})
	}
}
