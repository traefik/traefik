package middlewares

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/require"
)

const (
	rootCrt = `-----BEGIN CERTIFICATE-----
MIIDhjCCAm6gAwIBAgIJAIKZlW9a3VrYMA0GCSqGSIb3DQEBCwUAMFgxCzAJBgNV
BAYTAkZSMRMwEQYDVQQIDApTb21lLVN0YXRlMREwDwYDVQQHDAhUb3Vsb3VzZTEh
MB8GA1UECgwYSW50ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMB4XDTE4MDcxNzIwMzQz
OFoXDTE4MDgxNjIwMzQzOFowWDELMAkGA1UEBhMCRlIxEzARBgNVBAgMClNvbWUt
U3RhdGUxETAPBgNVBAcMCFRvdWxvdXNlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRn
aXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC1P8GJ
H9LkIxIIqK9MyUpushnjmjwccpSMB3OecISKYLy62QDIcAw6NzGcSe8hMwciMJr+
CdCjJlohybnaRI9hrJ3GPnI++UT/MMthf2IIcjmJxmD4k9L1fgs1V6zSTlo0+o0x
0gkAGlWvRkgA+3nt555ee84XQZuneKKeRRIlSA1ygycewFobZ/pGYijIEko+gYkV
sF3LnRGxNl673w+EQsvI7+z29T1nzjmM/xE7WlvnsrVd1/N61jAohLota0YTufwd
ioJZNryzuPejHBCiQRGMbJ7uEEZLiSCN6QiZEfqhS3AulykjgFXQQHn4zoVljSBR
UyLV0prIn5Scbks/AgMBAAGjUzBRMB0GA1UdDgQWBBTroRRnSgtkV+8dumtcftb/
lwIkATAfBgNVHSMEGDAWgBTroRRnSgtkV+8dumtcftb/lwIkATAPBgNVHRMBAf8E
BTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQAJ67U5cLa0ZFa/7zQQT4ldkY6YOEgR
0LNoTu51hc+ozaXSvF8YIBzkEpEnbGS3x4xodrwEBZjK2LFhNu/33gkCAuhmedgk
KwZrQM6lqRFGHGVOlkVz+QrJ2EsKYaO4SCUIwVjijXRLA7A30G5C/CIh66PsMgBY
6QHXVPEWm/v1d1Q/DfFfFzSOa1n1rIUw03qVJsxqSwfwYcegOF8YvS/eH4HUr2gF
cEujh6CCnylf35ExHa45atr3+xxbOVdNjobISkYADtbhAAn4KjLS4v8W6445vxxj
G5EIZLjOHyWg1sGaHaaAPkVpZQg8EKm21c4hrEEMfel60AMSSzad/a/V
-----END CERTIFICATE-----`

	minimalCert = `-----BEGIN CERTIFICATE-----
MIIDGTCCAgECCQCqLd75YLi2kDANBgkqhkiG9w0BAQsFADBYMQswCQYDVQQGEwJG
UjETMBEGA1UECAwKU29tZS1TdGF0ZTERMA8GA1UEBwwIVG91bG91c2UxITAfBgNV
BAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0xODA3MTgwODI4MTZaFw0x
ODA4MTcwODI4MTZaMEUxCzAJBgNVBAYTAkZSMRMwEQYDVQQIDApTb21lLVN0YXRl
MSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQC/+frDMMTLQyXG34F68BPhQq0kzK4LIq9Y0/gl
FjySZNn1C0QDWA1ubVCAcA6yY204I9cxcQDPNrhC7JlS5QA8Y5rhIBrqQlzZizAi
Rj3NTrRjtGUtOScnHuJaWjLy03DWD+aMwb7q718xt5SEABmmUvLwQK+EjW2MeDwj
y8/UEIpvrRDmdhGaqv7IFpIDkcIF7FowJ/hwDvx3PMc+z/JWK0ovzpvgbx69AVbw
ZxCimeha65rOqVi+lEetD26le+WnOdYsdJ2IkmpPNTXGdfb15xuAc+gFXfMCh7Iw
3Ynl6dZtZM/Ok2kiA7/OsmVnRKkWrtBfGYkI9HcNGb3zrk6nAgMBAAEwDQYJKoZI
hvcNAQELBQADggEBAC/R+Yvhh1VUhcbK49olWsk/JKqfS3VIDQYZg1Eo+JCPbwgS
I1BSYVfMcGzuJTX6ua3m/AHzGF3Tap4GhF4tX12jeIx4R4utnjj7/YKkTvuEM2f4
xT56YqI7zalGScIB0iMeyNz1QcimRl+M/49au8ow9hNX8C2tcA2cwd/9OIj/6T8q
SBRHc6ojvbqZSJCO0jziGDT1L3D+EDgTjED4nd77v/NRdP+egb0q3P0s4dnQ/5AV
aQlQADUn61j3ScbGJ4NSeZFFvsl38jeRi/MEzp0bGgNBcPj6JHi7qbbauZcZfQ05
jECvgAY7Nfd9mZ1KtyNaW31is+kag7NsvjxU/kM=
-----END CERTIFICATE-----`

	completeCert = `Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number: 3 (0x3)
    Signature Algorithm: sha1WithRSAEncryption
        Issuer: C=FR, ST=Some-State, L=Toulouse, O=Internet Widgits Pty Ltd
        Validity
            Not Before: Jul 18 08:00:16 2018 GMT
            Not After : Jul 18 08:00:16 2019 GMT
        Subject: C=FR, ST=SomeState, L=Toulouse, O=Cheese, CN=*.cheese.org
        Subject Public Key Info:
            Public Key Algorithm: rsaEncryption
                Public-Key: (2048 bit)
                Modulus:
                    00:a6:1f:96:7c:c1:cc:b8:1c:b5:91:5d:b8:bf:70:
                    bc:f7:b8:04:4f:2a:42:de:ea:c5:c3:19:0b:03:04:
                    ec:ef:a1:24:25:de:ad:05:e7:26:ea:89:6c:59:60:
                    10:18:0c:73:f1:bf:d3:cc:7b:ed:6b:9c:ea:1d:88:
                    e2:ee:14:81:d7:07:ee:87:95:3d:36:df:9c:38:b7:
                    7b:1e:2b:51:9c:4a:1f:d0:cc:5b:af:5d:6c:5c:35:
                    49:32:e4:01:5b:f9:8c:71:cf:62:48:5a:ea:b7:31:
                    58:e2:c6:d0:5b:1c:50:b5:5c:6d:5a:6f:da:41:5e:
                    d5:4c:6e:1a:21:f3:40:f9:9e:52:76:50:25:3e:03:
                    9b:87:19:48:5b:47:87:d3:67:c6:25:69:77:29:8e:
                    56:97:45:d9:6f:64:a8:4e:ad:35:75:2e:fc:6a:2e:
                    47:87:76:fc:4e:3e:44:e9:16:b2:c7:f0:23:98:13:
                    a2:df:15:23:cb:0c:3d:fd:48:5e:c7:2c:86:70:63:
                    8b:c6:c8:89:17:52:d5:a7:8e:cb:4e:11:9d:69:8e:
                    8e:59:cc:7e:a3:bd:a1:11:88:d7:cf:7b:8c:19:46:
                    9c:1b:7a:c9:39:81:4c:58:08:1f:c7:ce:b0:0e:79:
                    64:d3:11:72:65:e6:dd:bd:00:7f:22:30:46:9b:66:
                    9c:b9
                Exponent: 65537 (0x10001)
        X509v3 extensions:
            X509v3 Basic Constraints: 
                CA:FALSE
            X509v3 Subject Alternative Name: 
                DNS:*.cheese.org, DNS:*.cheese.net, DNS:cheese.in, IP Address:10.0.1.0, IP Address:10.0.1.2, email:test@cheese.org, email:test@cheese.net
            X509v3 Subject Key Identifier: 
                AB:6B:89:25:11:FC:5E:7B:D4:B0:F7:D4:B6:D9:EB:D0:30:93:E5:58
    Signature Algorithm: sha1WithRSAEncryption
         ad:87:84:a0:88:a3:4c:d9:0a:c0:14:e4:2d:9a:1d:bb:57:b7:
         12:ef:3a:fb:8b:b2:ce:32:b8:04:e6:59:c8:4f:14:6a:b5:12:
         46:e9:c9:0a:11:64:ea:a1:86:20:96:0e:a7:40:e3:aa:e5:98:
         91:36:89:77:b6:b9:73:7e:1a:58:19:ae:d1:14:83:1e:c1:5f:
         a5:a0:32:bb:52:68:b4:8d:a3:1d:b3:08:d7:45:6e:3b:87:64:
         7e:ef:46:e6:6f:d5:79:d7:1d:57:68:67:d8:18:39:61:5b:8b:
         1a:7f:88:da:0a:51:9b:3d:6c:5d:b1:cf:b7:e9:1e:06:65:8e:
         96:d3:61:96:f8:a2:61:f9:40:5e:fa:bc:76:b9:64:0e:6f:90:
         37:de:ac:6d:7f:36:84:35:19:88:8c:26:af:3e:c3:6a:1a:03:
         ed:d7:90:89:ed:18:4c:9e:94:1f:d8:ae:6c:61:36:17:72:f9:
         bb:de:0a:56:9a:79:b4:7d:4a:9d:cb:4a:7d:71:9f:38:e7:8d:
         f0:87:24:21:0a:24:1f:82:9a:6b:67:ce:7d:af:cb:91:6b:8a:
         de:e6:d8:6f:a1:37:b9:2d:d0:cb:e8:4e:f4:43:af:ad:90:13:
         7d:61:7a:ce:86:48:fc:00:8c:37:fb:e0:31:6b:e2:18:ad:fd:
         1e:df:08:db
-----BEGIN CERTIFICATE-----
MIIDvTCCAqWgAwIBAgIBAzANBgkqhkiG9w0BAQUFADBYMQswCQYDVQQGEwJGUjET
MBEGA1UECAwKU29tZS1TdGF0ZTERMA8GA1UEBwwIVG91bG91c2UxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0xODA3MTgwODAwMTZaFw0xOTA3
MTgwODAwMTZaMFwxCzAJBgNVBAYTAkZSMRIwEAYDVQQIDAlTb21lU3RhdGUxETAP
BgNVBAcMCFRvdWxvdXNlMQ8wDQYDVQQKDAZDaGVlc2UxFTATBgNVBAMMDCouY2hl
ZXNlLm9yZzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKYflnzBzLgc
tZFduL9wvPe4BE8qQt7qxcMZCwME7O+hJCXerQXnJuqJbFlgEBgMc/G/08x77Wuc
6h2I4u4UgdcH7oeVPTbfnDi3ex4rUZxKH9DMW69dbFw1STLkAVv5jHHPYkha6rcx
WOLG0FscULVcbVpv2kFe1UxuGiHzQPmeUnZQJT4Dm4cZSFtHh9NnxiVpdymOVpdF
2W9kqE6tNXUu/GouR4d2/E4+ROkWssfwI5gTot8VI8sMPf1IXscshnBji8bIiRdS
1aeOy04RnWmOjlnMfqO9oRGI1897jBlGnBt6yTmBTFgIH8fOsA55ZNMRcmXm3b0A
fyIwRptmnLkCAwEAAaOBjTCBijAJBgNVHRMEAjAAMF4GA1UdEQRXMFWCDCouY2hl
ZXNlLm9yZ4IMKi5jaGVlc2UubmV0ggljaGVlc2UuaW6HBAoAAQCHBAoAAQKBD3Rl
c3RAY2hlZXNlLm9yZ4EPdGVzdEBjaGVlc2UubmV0MB0GA1UdDgQWBBSra4klEfxe
e9Sw99S22evQMJPlWDANBgkqhkiG9w0BAQUFAAOCAQEArYeEoIijTNkKwBTkLZod
u1e3Eu86+4uyzjK4BOZZyE8UarUSRunJChFk6qGGIJYOp0DjquWYkTaJd7a5c34a
WBmu0RSDHsFfpaAyu1JotI2jHbMI10VuO4dkfu9G5m/VedcdV2hn2Bg5YVuLGn+I
2gpRmz1sXbHPt+keBmWOltNhlviiYflAXvq8drlkDm+QN96sbX82hDUZiIwmrz7D
ahoD7deQie0YTJ6UH9iubGE2F3L5u94KVpp5tH1KnctKfXGfOOeN8IckIQokH4Ka
a2fOfa/LkWuK3ubYb6E3uS3Qy+hO9EOvrZATfWF6zoZI/ACMN/vgMWviGK39Ht8I
2w==
-----END CERTIFICATE-----
`
)

func getCleanCertContents(certContents []string) string {
	var re = regexp.MustCompile("-----BEGIN CERTIFICATE-----(?s)(.*)")

	var cleanedCertContent []string
	for _, certContent := range certContents {
		cert := re.FindString(string(certContent))
		cleanedCertContent = append(cleanedCertContent, sanitize([]byte(cert)))
	}

	return strings.Join(cleanedCertContent, ",")
}

func getCertificate(certContent string) *x509.Certificate {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootCrt))
	if !ok {
		panic("failed to parse root certificate")
	}

	block, _ := pem.Decode([]byte(certContent))
	if block == nil {
		panic("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}

	return cert
}

func buildTLSWith(certContents []string) *tls.ConnectionState {
	var peerCertificates []*x509.Certificate

	for _, certContent := range certContents {
		peerCertificates = append(peerCertificates, getCertificate(certContent))
	}

	return &tls.ConnectionState{PeerCertificates: peerCertificates}
}

var myPassTLSClientCustomHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("bar"))
})

func getExpectedSanitized(s string) string {
	return url.QueryEscape(strings.Replace(s, "\n", "", -1))
}

func TestSanitize(t *testing.T) {
	testCases := []struct {
		desc       string
		toSanitize []byte
		expected   string
	}{
		{
			desc: "Empty",
		},
		{
			desc:       "With a minimal cert",
			toSanitize: []byte(minimalCert),
			expected: getExpectedSanitized(`MIIDGTCCAgECCQCqLd75YLi2kDANBgkqhkiG9w0BAQsFADBYMQswCQYDVQQGEwJG
UjETMBEGA1UECAwKU29tZS1TdGF0ZTERMA8GA1UEBwwIVG91bG91c2UxITAfBgNV
BAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0xODA3MTgwODI4MTZaFw0x
ODA4MTcwODI4MTZaMEUxCzAJBgNVBAYTAkZSMRMwEQYDVQQIDApTb21lLVN0YXRl
MSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQC/+frDMMTLQyXG34F68BPhQq0kzK4LIq9Y0/gl
FjySZNn1C0QDWA1ubVCAcA6yY204I9cxcQDPNrhC7JlS5QA8Y5rhIBrqQlzZizAi
Rj3NTrRjtGUtOScnHuJaWjLy03DWD+aMwb7q718xt5SEABmmUvLwQK+EjW2MeDwj
y8/UEIpvrRDmdhGaqv7IFpIDkcIF7FowJ/hwDvx3PMc+z/JWK0ovzpvgbx69AVbw
ZxCimeha65rOqVi+lEetD26le+WnOdYsdJ2IkmpPNTXGdfb15xuAc+gFXfMCh7Iw
3Ynl6dZtZM/Ok2kiA7/OsmVnRKkWrtBfGYkI9HcNGb3zrk6nAgMBAAEwDQYJKoZI
hvcNAQELBQADggEBAC/R+Yvhh1VUhcbK49olWsk/JKqfS3VIDQYZg1Eo+JCPbwgS
I1BSYVfMcGzuJTX6ua3m/AHzGF3Tap4GhF4tX12jeIx4R4utnjj7/YKkTvuEM2f4
xT56YqI7zalGScIB0iMeyNz1QcimRl+M/49au8ow9hNX8C2tcA2cwd/9OIj/6T8q
SBRHc6ojvbqZSJCO0jziGDT1L3D+EDgTjED4nd77v/NRdP+egb0q3P0s4dnQ/5AV
aQlQADUn61j3ScbGJ4NSeZFFvsl38jeRi/MEzp0bGgNBcPj6JHi7qbbauZcZfQ05
jECvgAY7Nfd9mZ1KtyNaW31is+kag7NsvjxU/kM=`),
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.expected, sanitize(test.toSanitize), "The sanitized certificates should be equal")
		})
	}

}

func TestTlsClientheadersWithPEM(t *testing.T) {
	testCases := []struct {
		desc                 string
		certContents         []string // set the request TLS attribute if defined
		tlsClientCertHeaders *types.TLSClientHeaders
		expectedHeader       string
	}{
		{
			desc: "No TLS, no option",
		},
		{
			desc:         "TLS, no option",
			certContents: []string{minimalCert},
		},
		{
			desc:                 "No TLS, with pem option true",
			tlsClientCertHeaders: &types.TLSClientHeaders{PEM: true},
		},
		{
			desc:                 "TLS with simple certificate, with pem option true",
			certContents:         []string{minimalCert},
			tlsClientCertHeaders: &types.TLSClientHeaders{PEM: true},
			expectedHeader:       getCleanCertContents([]string{minimalCert}),
		},
		{
			desc:                 "TLS with complete certificate, with pem option true",
			certContents:         []string{completeCert},
			tlsClientCertHeaders: &types.TLSClientHeaders{PEM: true},
			expectedHeader:       getCleanCertContents([]string{completeCert}),
		},
		{
			desc:                 "TLS with two certificate, with pem option true",
			certContents:         []string{minimalCert, completeCert},
			tlsClientCertHeaders: &types.TLSClientHeaders{PEM: true},
			expectedHeader:       getCleanCertContents([]string{minimalCert, completeCert}),
		},
	}

	for _, test := range testCases {
		tlsClientHeaders := NewTLSClientHeaders(&types.Frontend{PassTLSClientCert: test.tlsClientCertHeaders})

		res := httptest.NewRecorder()
		req := testhelpers.MustNewRequest(http.MethodGet, "http://example.com/foo", nil)

		if test.certContents != nil && len(test.certContents) > 0 {
			req.TLS = buildTLSWith(test.certContents)
		}

		tlsClientHeaders.ServeHTTP(res, req, myPassTLSClientCustomHandler)

		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, http.StatusOK, res.Code, "Http Status should be OK")
			require.Equal(t, "bar", res.Body.String(), "Should be the expected body")

			if test.expectedHeader != "" {
				require.Equal(t, getCleanCertContents(test.certContents), req.Header.Get(xForwardedTLSClientCert), "The request header should contain the cleaned certificate")
			} else {
				require.Empty(t, req.Header.Get(xForwardedTLSClientCert))
			}
			require.Empty(t, res.Header().Get(xForwardedTLSClientCert), "The response header should be always empty")
		})
	}

}

func TestGetSans(t *testing.T) {
	urlFoo, err := url.Parse("my.foo.com")
	require.NoError(t, err)
	urlBar, err := url.Parse("my.bar.com")
	require.NoError(t, err)

	testCases := []struct {
		desc     string
		cert     *x509.Certificate // set the request TLS attribute if defined
		expected []string
	}{
		{
			desc: "With nil",
		},
		{
			desc: "Certificate without Sans",
			cert: &x509.Certificate{},
		},
		{
			desc: "Certificate with all Sans",
			cert: &x509.Certificate{
				DNSNames:       []string{"foo", "bar"},
				EmailAddresses: []string{"test@test.com", "test2@test.com"},
				IPAddresses:    []net.IP{net.IPv4(10, 0, 0, 1), net.IPv4(10, 0, 0, 2)},
				URIs:           []*url.URL{urlFoo, urlBar},
			},
			expected: []string{"foo", "bar", "test@test.com", "test2@test.com", "10.0.0.1", "10.0.0.2", urlFoo.String(), urlBar.String()},
		},
	}

	for _, test := range testCases {
		sans := getSANs(test.cert)
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if len(test.expected) > 0 {
				for i, expected := range test.expected {
					require.Equal(t, expected, sans[i])
				}
			} else {
				require.Empty(t, sans)
			}
		})
	}

}

func TestTlsClientheadersWithCertInfos(t *testing.T) {
	minimalCertAllInfos := `Subject="C=FR,ST=Some-State,O=Internet Widgits Pty Ltd",NB=1531902496,NA=1534494496,SAN=`
	completeCertAllInfos := `Subject="C=FR,ST=SomeState,L=Toulouse,O=Cheese,CN=*.cheese.org",NB=1531900816,NA=1563436816,SAN=*.cheese.org,*.cheese.net,cheese.in,test@cheese.org,test@cheese.net,10.0.1.0,10.0.1.2`

	testCases := []struct {
		desc                 string
		certContents         []string // set the request TLS attribute if defined
		tlsClientCertHeaders *types.TLSClientHeaders
		expectedHeader       string
	}{
		{
			desc: "No TLS, no option",
		},
		{
			desc:         "TLS, no option",
			certContents: []string{minimalCert},
		},
		{
			desc: "No TLS, with pem option true",
			tlsClientCertHeaders: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					Subject: &types.TLSCLientCertificateSubjectInfos{
						CommonName:   true,
						Organization: true,
						Locality:     true,
						Province:     true,
						Country:      true,
						SerialNumber: true,
					},
				},
			},
		},
		{
			desc: "No TLS, with pem option true with no flag",
			tlsClientCertHeaders: &types.TLSClientHeaders{
				PEM: false,
				Infos: &types.TLSClientCertificateInfos{
					Subject: &types.TLSCLientCertificateSubjectInfos{},
				},
			},
		},
		{
			desc:         "TLS with simple certificate, with all infos",
			certContents: []string{minimalCert},
			tlsClientCertHeaders: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					NotAfter:  true,
					NotBefore: true,
					Subject: &types.TLSCLientCertificateSubjectInfos{
						CommonName:   true,
						Organization: true,
						Locality:     true,
						Province:     true,
						Country:      true,
						SerialNumber: true,
					},
					Sans: true,
				},
			},
			expectedHeader: url.QueryEscape(minimalCertAllInfos),
		},
		{
			desc:         "TLS with simple certificate, with some infos",
			certContents: []string{minimalCert},
			tlsClientCertHeaders: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					NotAfter: true,
					Subject: &types.TLSCLientCertificateSubjectInfos{
						Organization: true,
					},
					Sans: true,
				},
			},
			expectedHeader: url.QueryEscape(`Subject="O=Internet Widgits Pty Ltd",NA=1534494496,SAN=`),
		},
		{
			desc:         "TLS with complete certificate, with all infos",
			certContents: []string{completeCert},
			tlsClientCertHeaders: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					NotAfter:  true,
					NotBefore: true,
					Subject: &types.TLSCLientCertificateSubjectInfos{
						CommonName:   true,
						Organization: true,
						Locality:     true,
						Province:     true,
						Country:      true,
						SerialNumber: true,
					},
					Sans: true,
				},
			},
			expectedHeader: url.QueryEscape(completeCertAllInfos),
		},
		{
			desc:         "TLS with 2 certificates, with all infos",
			certContents: []string{minimalCert, completeCert},
			tlsClientCertHeaders: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					NotAfter:  true,
					NotBefore: true,
					Subject: &types.TLSCLientCertificateSubjectInfos{
						CommonName:   true,
						Organization: true,
						Locality:     true,
						Province:     true,
						Country:      true,
						SerialNumber: true,
					},
					Sans: true,
				},
			},
			expectedHeader: url.QueryEscape(strings.Join([]string{minimalCertAllInfos, completeCertAllInfos}, ";")),
		},
	}

	for _, test := range testCases {
		tlsClientHeaders := NewTLSClientHeaders(&types.Frontend{PassTLSClientCert: test.tlsClientCertHeaders})

		res := httptest.NewRecorder()
		req := testhelpers.MustNewRequest(http.MethodGet, "http://example.com/foo", nil)

		if test.certContents != nil && len(test.certContents) > 0 {
			req.TLS = buildTLSWith(test.certContents)
		}

		tlsClientHeaders.ServeHTTP(res, req, myPassTLSClientCustomHandler)

		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, http.StatusOK, res.Code, "Http Status should be OK")
			require.Equal(t, "bar", res.Body.String(), "Should be the expected body")

			if test.expectedHeader != "" {
				require.Equal(t, test.expectedHeader, req.Header.Get(xForwardedTLSClientCertInfos), "The request header should contain the cleaned certificate")
			} else {
				require.Empty(t, req.Header.Get(xForwardedTLSClientCertInfos))
			}
			require.Empty(t, res.Header().Get(xForwardedTLSClientCertInfos), "The response header should be always empty")
		})
	}

}

func TestNewTLSClientHeadersFromStruct(t *testing.T) {
	testCases := []struct {
		desc     string
		frontend *types.Frontend
		expected *TLSClientHeaders
	}{
		{
			desc: "Without frontend",
		},
		{
			desc:     "frontend without the option",
			frontend: &types.Frontend{},
			expected: &TLSClientHeaders{},
		},
		{
			desc: "frontend with the pem set false",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					PEM: false,
				},
			},
			expected: &TLSClientHeaders{PEM: false},
		},
		{
			desc: "frontend with the pem set true",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					PEM: true,
				},
			},
			expected: &TLSClientHeaders{PEM: true},
		},
		{
			desc: "frontend with the Infos with no flag",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					Infos: &types.TLSClientCertificateInfos{
						NotAfter:  false,
						NotBefore: false,
						Sans:      false,
					},
				},
			},
			expected: &TLSClientHeaders{
				PEM:   false,
				Infos: &TLSClientCertificateInfos{},
			},
		},
		{
			desc: "frontend with the Infos basic",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					Infos: &types.TLSClientCertificateInfos{
						NotAfter:  true,
						NotBefore: true,
						Sans:      true,
					},
				},
			},
			expected: &TLSClientHeaders{
				PEM: false,
				Infos: &TLSClientCertificateInfos{
					NotBefore: true,
					NotAfter:  true,
					Sans:      true,
				},
			},
		},
		{
			desc: "frontend with the Infos NotAfter",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					Infos: &types.TLSClientCertificateInfos{
						NotAfter: true,
					},
				},
			},
			expected: &TLSClientHeaders{
				PEM: false,
				Infos: &TLSClientCertificateInfos{
					NotAfter: true,
				},
			},
		},
		{
			desc: "frontend with the Infos NotBefore",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					Infos: &types.TLSClientCertificateInfos{
						NotBefore: true,
					},
				},
			},
			expected: &TLSClientHeaders{
				PEM: false,
				Infos: &TLSClientCertificateInfos{
					NotBefore: true,
				},
			},
		},
		{
			desc: "frontend with the Infos Sans",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					Infos: &types.TLSClientCertificateInfos{
						Sans: true,
					},
				},
			},
			expected: &TLSClientHeaders{
				PEM: false,
				Infos: &TLSClientCertificateInfos{
					Sans: true,
				},
			},
		},
		{
			desc: "frontend with the Infos Subject Organization",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					Infos: &types.TLSClientCertificateInfos{
						Subject: &types.TLSCLientCertificateSubjectInfos{
							Organization: true,
						},
					},
				},
			},
			expected: &TLSClientHeaders{
				PEM: false,
				Infos: &TLSClientCertificateInfos{
					Subject: &TLSCLientCertificateSubjectInfos{
						Organization: true,
					},
				},
			},
		},
		{
			desc: "frontend with the Infos Subject Country",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					Infos: &types.TLSClientCertificateInfos{
						Subject: &types.TLSCLientCertificateSubjectInfos{
							Country: true,
						},
					},
				},
			},
			expected: &TLSClientHeaders{
				PEM: false,
				Infos: &TLSClientCertificateInfos{
					Subject: &TLSCLientCertificateSubjectInfos{
						Country: true,
					},
				},
			},
		},
		{
			desc: "frontend with the Infos Subject SerialNumber",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					Infos: &types.TLSClientCertificateInfos{
						Subject: &types.TLSCLientCertificateSubjectInfos{
							SerialNumber: true,
						},
					},
				},
			},
			expected: &TLSClientHeaders{
				PEM: false,
				Infos: &TLSClientCertificateInfos{
					Subject: &TLSCLientCertificateSubjectInfos{
						SerialNumber: true,
					},
				},
			},
		},
		{
			desc: "frontend with the Infos Subject Province",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					Infos: &types.TLSClientCertificateInfos{
						Subject: &types.TLSCLientCertificateSubjectInfos{
							Province: true,
						},
					},
				},
			},
			expected: &TLSClientHeaders{
				PEM: false,
				Infos: &TLSClientCertificateInfos{
					Subject: &TLSCLientCertificateSubjectInfos{
						Province: true,
					},
				},
			},
		},
		{
			desc: "frontend with the Infos Subject Locality",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					Infos: &types.TLSClientCertificateInfos{
						Subject: &types.TLSCLientCertificateSubjectInfos{
							Locality: true,
						},
					},
				},
			},
			expected: &TLSClientHeaders{
				PEM: false,
				Infos: &TLSClientCertificateInfos{
					Subject: &TLSCLientCertificateSubjectInfos{
						Locality: true,
					},
				},
			},
		},
		{
			desc: "frontend with the Infos Subject CommonName",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					Infos: &types.TLSClientCertificateInfos{
						Subject: &types.TLSCLientCertificateSubjectInfos{
							CommonName: true,
						},
					},
				},
			},
			expected: &TLSClientHeaders{
				PEM: false,
				Infos: &TLSClientCertificateInfos{
					Subject: &TLSCLientCertificateSubjectInfos{
						CommonName: true,
					},
				},
			},
		},
		{
			desc: "frontend with the Infos NotBefore",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					Infos: &types.TLSClientCertificateInfos{
						Sans: true,
					},
				},
			},
			expected: &TLSClientHeaders{
				PEM: false,
				Infos: &TLSClientCertificateInfos{
					Sans: true,
				},
			},
		},
		{
			desc: "frontend with the Infos all",
			frontend: &types.Frontend{
				PassTLSClientCert: &types.TLSClientHeaders{
					Infos: &types.TLSClientCertificateInfos{
						NotAfter:  true,
						NotBefore: true,
						Subject: &types.TLSCLientCertificateSubjectInfos{
							CommonName:   true,
							Country:      true,
							Locality:     true,
							Organization: true,
							Province:     true,
							SerialNumber: true,
						},
						Sans: true,
					},
				},
			},
			expected: &TLSClientHeaders{
				PEM: false,
				Infos: &TLSClientCertificateInfos{
					NotBefore: true,
					NotAfter:  true,
					Sans:      true,
					Subject: &TLSCLientCertificateSubjectInfos{
						Province:     true,
						Organization: true,
						Locality:     true,
						Country:      true,
						CommonName:   true,
						SerialNumber: true,
					},
				}},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.expected, NewTLSClientHeaders(test.frontend))
		})
	}

}
