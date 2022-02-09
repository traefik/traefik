package main

import (
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FooCert is a PEM-encoded TLS cert.
// generated from src/crypto/tls:
// go run generate_cert.go  --rsa-bits 1024 --host foo.org,foo.com  --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
const fooCert = `-----BEGIN CERTIFICATE-----
MIICHzCCAYigAwIBAgIQXQFLeYRwc5X21t457t2xADANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDCjn67GSs/khuGC4GNN+tVo1S+/eSHwr/hWzhfMqO7nYiXkFzmxi+u14CU
Pda6WOeps7T2/oQEFMxKKg7zYOqkLSbjbE0ZfosopaTvEsZm/AZHAAvoOrAsIJOn
SEiwy8h0tLA4z1SNR6rmIVQWyqBZEPAhBTQM1z7tFp48FakCFwIDAQABo3QwcjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAdBgNVHQ4EFgQUDHG3ASzeUezElup9zbPpBn/vjogwGwYDVR0RBBQwEoIH
Zm9vLm9yZ4IHZm9vLmNvbTANBgkqhkiG9w0BAQsFAAOBgQBT+VLMbB9u27tBX8Aw
ZrGY3rbNdBGhXVTksrjiF+6ZtDpD3iI56GH9zLxnqvXkgn3u0+Ard5TqF/xmdwVw
NY0V/aWYfcL2G2auBCQrPvM03ozRnVUwVfP23eUzX2ORNHCYhd2ObQx4krrhs7cJ
SWxtKwFlstoXY3K2g9oRD9UxdQ==
-----END CERTIFICATE-----`

// BarCert is a PEM-encoded TLS cert.
// generated from src/crypto/tls:
// go run generate_cert.go  --rsa-bits 1024 --host bar.org,bar.com  --ca --start-date "Jan 1 00:00:00 1970" --duration=10000h
const barCert = `-----BEGIN CERTIFICATE-----
MIICHTCCAYagAwIBAgIQcuIcNEXzBHPoxna5S6wG4jANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMB4XDTcwMDEwMTAwMDAwMFoXDTcxMDIyMTE2MDAw
MFowEjEQMA4GA1UEChMHQWNtZSBDbzCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkC
gYEAqtcrP+KA7D6NjyztGNIPMup9KiBMJ8QL+preog/YHR7SQLO3kGFhpS3WKMab
SzMypC3ZX1PZjBP5ZzwaV3PFbuwlCkPlyxR2lOWmullgI7mjY0TBeYLDIclIzGRp
mpSDDSpkW1ay2iJDSpXjlhmwZr84hrCU7BRTQJo91fdsRTsCAwEAAaN0MHIwDgYD
VR0PAQH/BAQDAgKkMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1UdEwEB/wQFMAMB
Af8wHQYDVR0OBBYEFK8jnzFQvBAgWtfzOyXY4VSkwrTXMBsGA1UdEQQUMBKCB2Jh
ci5vcmeCB2Jhci5jb20wDQYJKoZIhvcNAQELBQADgYEAJz0ifAExisC/ZSRhWuHz
7qs1i6Nd4+YgEVR8dR71MChP+AMxucY1/ajVjb9xlLys3GPE90TWSdVppabEVjZY
Oq11nPKc50ItTt8dMku6t0JHBmzoGdkN0V4zJCBqdQJxhop8JpYJ0S9CW0eT93h3
ipYQSsmIINGtMXJ8VkP/MlM=
-----END CERTIFICATE-----`

type gaugeMock struct {
	metrics map[string]float64
	labels  string
}

func (g gaugeMock) With(labelValues ...string) metrics.Gauge {
	g.labels = strings.Join(labelValues, ",")
	return g
}

func (g gaugeMock) Set(value float64) {
	g.metrics[g.labels] = value
}

func (g gaugeMock) Add(delta float64) {
	panic("implement me")
}

func TestAppendCertMetric(t *testing.T) {
	testCases := []struct {
		desc     string
		certs    []string
		expected map[string]float64
	}{
		{
			desc:     "No certs",
			certs:    []string{},
			expected: map[string]float64{},
		},
		{
			desc:  "One cert",
			certs: []string{fooCert},
			expected: map[string]float64{
				"cn,,serial,123624926713171615935660664614975025408,sans,foo.com,foo.org": 3.6e+09,
			},
		},
		{
			desc:  "Two certs",
			certs: []string{fooCert, barCert},
			expected: map[string]float64{
				"cn,,serial,123624926713171615935660664614975025408,sans,foo.com,foo.org": 3.6e+09,
				"cn,,serial,152706022658490889223053211416725817058,sans,bar.com,bar.org": 3.6e+07,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			gauge := &gaugeMock{
				metrics: map[string]float64{},
			}

			for _, cert := range test.certs {
				block, _ := pem.Decode([]byte(cert))
				parsedCert, err := x509.ParseCertificate(block.Bytes)
				require.NoError(t, err)

				appendCertMetric(gauge, parsedCert)
			}

			assert.Equal(t, test.expected, gauge.metrics)
		})
	}
}
