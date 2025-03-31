package main

import (
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
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

// LocalhostCert is a PEM-encoded TLS cert with SAN IPs
// "127.0.0.1" and "[::1]", expiring at Jan 29 16:00:00 2084 GMT.
// generated from src/crypto/tls:
// go run generate_cert.go  --rsa-bits 1024 --host 127.0.0.1,::1,example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var (
	localhostCert = types.FileOrContent(`-----BEGIN CERTIFICATE-----
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
	localhostKey = types.FileOrContent(`-----BEGIN RSA PRIVATE KEY-----
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

	// LocalhostCert is a PEM-encoded TLS cert with SAN IPs
	// "127.0.0.1" and "[::1]", expiring at Jan 29 16:00:00 2084 GMT.
	// generated from src/crypto/tls:
	// go run generate_cert.go  --rsa-bits 1024 --host 127.0.0.1,::1,example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h

	localhostCert2 = types.FileOrContent(`-----BEGIN CERTIFICATE-----
MIICMjCCAZugAwIBAgIQU2ZgpNbD+iY0bT3uguidWDANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMB4XDTAwMDEwMTAwMDAwMFoXDTExMDUyOTE2MDAw
MFowEjEQMA4GA1UEChMHQWNtZSBDbzCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkC
gYEAoHPlkXE2KbnrpX/hMBh6N28EOGAIEcyquKUUASU38gJNBH2L302FOfbSl89M
wW2IzzKeErt2tHmMyRFTY8RrBb9NtBVmCd3u/DOPOxzD2Ixo7bDGTny4lAweWTac
6n+xaK6j4lqW7InzFeUlKnW2iR/aycZDjCLpBSH86hHIMUsCAwEAAaOBiDCBhTAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAdBgNVHQ4EFgQU5N5PHSuljAXRQLhGXe7rCNcPOxUwLgYDVR0RBCcwJYIL
ZXhhbXBsZS5jb22HBH8AAAGHEAAAAAAAAAAAAAAAAAAAAAEwDQYJKoZIhvcNAQEL
BQADgYEAjRCj6e8evJhYx3WAMQ+QncxB45Ck7YpjFT4yAQyLb55c1tqgnQNsdfNh
M34jKWtGacADNMB5I2/ZPvRK+vrsy2t0WXk0qEdEGsXNQ3amvmkmqdJl6GdSSoBG
3mv6CmILj46ycaklJl/SGYVJMkAbbFZ+sDvK+oy1xThYpitmXe0=
-----END CERTIFICATE-----
`)

	// LocalhostKey is the private key for localhostCert.
	localhostKey2 = types.FileOrContent(`-----BEGIN PRIVATE KEY-----
MIICdQIBADANBgkqhkiG9w0BAQEFAASCAl8wggJbAgEAAoGBAKBz5ZFxNim566V/
4TAYejdvBDhgCBHMqrilFAElN/ICTQR9i99NhTn20pfPTMFtiM8ynhK7drR5jMkR
U2PEawW/TbQVZgnd7vwzjzscw9iMaO2wxk58uJQMHlk2nOp/sWiuo+JaluyJ8xXl
JSp1tokf2snGQ4wi6QUh/OoRyDFLAgMBAAECgYA0P6lI5EHL8qP+n5bXz5C0zmzk
Yrkd+rS5LeBGwzTllMQ5qxxKGfdBOdO35aRL9HwxZH0/AlaUTGSA8Shje4mRsHYd
G+ScExUCjnvtdnC+lPdRSF1KoVfDFjxHeD4f/BGFILCLioZCv8RDaj0VnOlMeLZa
kkl7+oCGOG3tyLzTUQJBANIiq/LBwJaRXkWFkTg++bdOLOZIIdRGDbNP4ynWN8o3
BIwj/Y+8LjtiZTRARg3eCKtT27zBPIMezoVjWVN4CrMCQQDDeTNqalLOmqSwgYyY
QovHG1NuG3AgSqVeJvT1QepCNJkOppCnX6PtlHRlBwJ9qrvPSd5NmIxAJGmIsn1p
DWsJAkA8+bSdf51r04jgcY6fHJ8Hktayh9HRL/a/xnmrZS7RLb/TDoqAT+G2d6nY
TKJHWdt4I6BKmGP/xEu3Jwn/j4DDAkB4y0YVpbykRfYtyPDMCpt8IAvPiA8jNV25
sBNCGEiePwiygAX2GGkh4NKIt+s3IzHKKBjDFNjermG1Aq/zIkKZAkAvxCCd2umt
dHxbkSwS7Dl0ihhX38ZYmVpJery7/8g6HyMX0yV2D0/aK+vnzetKWVwrvxlCPgCd
B0+7l+Zl0B1y
-----END PRIVATE KEY-----
`)
)

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

func TestTlsUpdater(t *testing.T) {
	tlsManager := traefiktls.NewManager()

	testCases := []struct {
		desc     string
		certs    []traefiktls.Certificate
		expected map[string]float64
	}{
		{
			desc:     "No certs",
			certs:    []traefiktls.Certificate{},
			expected: map[string]float64{},
		},
		{
			desc: "Add new certs",
			certs: []traefiktls.Certificate{
				{CertFile: localhostCert, KeyFile: localhostKey},
			},
			expected: map[string]float64{
				"cn,,serial,97129276724337570813249812937731361303,sans,example.com": 3.6e+09,
			},
		},
		{
			desc: "Reset one old cert",
			certs: []traefiktls.Certificate{
				{CertFile: localhostCert2, KeyFile: localhostKey2},
			},
			expected: map[string]float64{
				"cn,,serial,110857498100925886701805594271659105624,sans,example.com": 1.3066848e+09,
				"cn,,serial,97129276724337570813249812937731361303,sans,example.com":  0,
			},
		},
		{
			desc:  "No certs, reset to 0",
			certs: []traefiktls.Certificate{},
			expected: map[string]float64{
				"cn,,serial,110857498100925886701805594271659105624,sans,example.com": 0,
				"cn,,serial,97129276724337570813249812937731361303,sans,example.com":  0,
			},
		},
	}

	gauge := &gaugeMock{
		metrics: map[string]float64{},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {

			certAndStores := make([]*traefiktls.CertAndStores, 0)
			for _, cert := range test.certs {
				certAndStores = append(certAndStores,
					&traefiktls.CertAndStores{
						Certificate: cert,
						Stores:      []string{traefiktls.DefaultTLSStoreName},
					})
			}

			dynamicConfig := dynamic.Configuration{
				TLS: &dynamic.TLSConfiguration{
					Certificates: certAndStores,
				},
			}

			tlsUpdater(tlsManager, gauge)(dynamicConfig)

			assert.Equal(t, test.expected, gauge.metrics)
		})
	}
}

func TestGetDefaultsEntrypoints(t *testing.T) {
	testCases := []struct {
		desc        string
		entrypoints static.EntryPoints
		expected    []string
	}{
		{
			desc: "Skips special names",
			entrypoints: map[string]*static.EntryPoint{
				"web": {
					Address: ":80",
				},
				"traefik": {
					Address: ":8080",
				},
			},
			expected: []string{"web"},
		},
		{
			desc: "Two EntryPoints not attachable",
			entrypoints: map[string]*static.EntryPoint{
				"web": {
					Address: ":80",
				},
				"websecure": {
					Address: ":443",
				},
			},
			expected: []string{"web", "websecure"},
		},
		{
			desc: "Two EntryPoints only one attachable",
			entrypoints: map[string]*static.EntryPoint{
				"web": {
					Address: ":80",
				},
				"websecure": {
					Address:   ":443",
					AsDefault: true,
				},
			},
			expected: []string{"websecure"},
		},
		{
			desc: "Two attachable EntryPoints",
			entrypoints: map[string]*static.EntryPoint{
				"web": {
					Address:   ":80",
					AsDefault: true,
				},
				"websecure": {
					Address:   ":443",
					AsDefault: true,
				},
			},
			expected: []string{"web", "websecure"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			actual := getDefaultsEntrypoints(&static.Configuration{
				EntryPoints: test.entrypoints,
			})

			assert.ElementsMatch(t, test.expected, actual)
		})
	}
}
