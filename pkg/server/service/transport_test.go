package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

// LocalhostCert is a PEM-encoded TLS cert
// for host example.com, www.example.com
// expiring at Jan 29 16:00:00 2084 GMT.
// go run $GOROOT/src/crypto/tls/generate_cert.go  --rsa-bits 1024 --host example.com,www.example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var LocalhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIICDDCCAXWgAwIBAgIQH20JmcOlcRWHNuf62SYwszANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQC0qINy3F4oq6viDnlpDDE5J08iSRGggg6EylJKBKZfphEG2ufgK78Dufl3
+7b0LlEY2AeZHwviHODqC9a6ihj1ZYQk0/djAh+OeOhFEWu+9T/VP8gVFarFqT8D
Opy+hrG7YJivUIzwb4fmJQRI7FajzsnGyM6LiXLU+0qzb7ZO/QIDAQABo2EwXzAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAnBgNVHREEIDAeggtleGFtcGxlLmNvbYIPd3d3LmV4YW1wbGUuY29tMA0G
CSqGSIb3DQEBCwUAA4GBAB+eluoQYzyyMfeEEAOtlldevx5MtDENT05NB0WI+91R
we7mX8lv763u0XuCWPxbHszhclI6FFjoQef0Z1NYLRm8ZRq58QqWDFZ3E6wdDK+B
+OWvkW+hRavo6R9LzIZPfbv8yBo4M9PK/DXw8hLqH7VkkI+Gh793iH7Ugd4A7wvT
-----END CERTIFICATE-----`)

// LocalhostKey is the private key for localhostCert.
var LocalhostKey = []byte(`-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBALSog3LcXiirq+IO
eWkMMTknTyJJEaCCDoTKUkoEpl+mEQba5+ArvwO5+Xf7tvQuURjYB5kfC+Ic4OoL
1rqKGPVlhCTT92MCH4546EURa771P9U/yBUVqsWpPwM6nL6GsbtgmK9QjPBvh+Yl
BEjsVqPOycbIzouJctT7SrNvtk79AgMBAAECgYB1wMT1MBgbkFIXpXGTfAP1id61
rUTVBxCpkypx3ngHLjo46qRq5Hi72BN4FlTY8fugIudI8giP2FztkMvkiLDc4m0p
Gn+QMJzjlBjjTuNLvLy4aSmNRLIC3mtbx9PdU71DQswEpJHFj/vmsxbuSrG1I1YE
r1reuSo2ow6fOAjXLQJBANpz+RkOiPSPuvl+gi1sp2pLuynUJVDVqWZi386YRpfg
DiKCLpqwqYDkOozm/fwFALvwXKGmsyyL43HO8eI+2NsCQQDTtY32V+02GPecdsyq
msK06EPVTSaYwj9Mm+q709KsmYFHLXDqXjcKV4UgKYKRPz7my1fXodMmGmfuh1a3
/HMHAkEAmOQKN0tA90mRJwUvvvMIyRBv0fq0kzq28P3KfiF9ZtZdjjFmxMVYHOmf
QPZ6VGR7+w1jB5BQXqEZcpHQIPSzeQJBAIy9tZJ/AYNlNbcegxEnsSjy/6VdlLsY
51vWi0Yym2uC4R6gZuBnoc+OP0ISVmqY0Qg9RjhjrCs4gr9f2ZaWjSECQCxqZMq1
3viJ8BGCC0m/5jv1EHur3YgwphYCkf4Li6DKwIdMLk1WXkTcPIY3V2Jqj8rPEB5V
rqPRSAtd/h6oZbs=
-----END PRIVATE KEY-----`)

//	openssl req -newkey rsa:2048 \
//	   -new -nodes -x509 \
//	   -days 3650 \
//	   -out cert.pem \
//	   -keyout key.pem \
//	   -subj "/CN=example.com"
//	   -addext "subjectAltName = DNS:example.com"
var mTLSCert = []byte(`-----BEGIN CERTIFICATE-----
MIIDJTCCAg2gAwIBAgIUYKnGcLnmMosOSKqTn4ydAMURE4gwDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAwwLZXhhbXBsZS5jb20wHhcNMjAwODEzMDkyNzIwWhcNMzAw
ODExMDkyNzIwWjAWMRQwEgYDVQQDDAtleGFtcGxlLmNvbTCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAOAe+QM1c9lZ2TPRgoiuPAq2A3Pfu+i82lmqrTJ0
PR2Cx1fPbccCUTFJPlxSDiaMrwtvqw1yP9L2Pu/vJK5BY4YDVDtFGKjpRBau1otJ
iY50O5qMo3sfLqR4/1VsQGlLVZYLD3dyc4ZTmOp8+7tJ2SyGorojbIKfimZT7XD7
dzrVr4h4Gn+SzzOnoKyx29uaNRP+XuMYHmHyQcJE03pUGhkTOvMwBlF96QdQ9WG0
D+1CxRciEsZXE+imKBHoaTgrTkpnFHzsrIEw+OHQYf30zuT/k/lkgv1vqEwINHjz
W2VeTur5eqVvA7zZdoEXMRy7BUvh/nZk5AXkXAmZLn0eUg8CAwEAAaNrMGkwHQYD
VR0OBBYEFEDrbhPDt+hi3ZOzk6S/CFAVHwk0MB8GA1UdIwQYMBaAFEDrbhPDt+hi
3ZOzk6S/CFAVHwk0MA8GA1UdEwEB/wQFMAMBAf8wFgYDVR0RBA8wDYILZXhhbXBs
ZS5jb20wDQYJKoZIhvcNAQELBQADggEBAG/JRJWeUNx2mDJAk8W7Syq3gmQB7s9f
+yY/XVRJZGahOPilABqFpC6GVn2HWuvuOqy8/RGk9ja5abKVXqE6YKrljqo3XfzB
KQcOz4SFirpkHvNCiEcK3kggN3wJWqL2QyXAxWldBBBCO9yx7a3cux31C//sTUOG
xq4JZDg171U1UOpfN1t0BFMdt05XZFEM247N7Dcf7HoXwAa7eyLKgtKWqPDqGrFa
fvGDDKK9X/KVsU2x9V3pG+LsJg7ogUnSyD2r5G1F3Y8OVs2T/783PaN0M35fDL38
09VbsxA2GasOHZrghUzT4UvZWWZbWEmG975hFYvdj6DlK9K0s5TdKIs=
-----END CERTIFICATE-----`)

var mTLSKey = []byte(`-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDgHvkDNXPZWdkz
0YKIrjwKtgNz37vovNpZqq0ydD0dgsdXz23HAlExST5cUg4mjK8Lb6sNcj/S9j7v
7ySuQWOGA1Q7RRio6UQWrtaLSYmOdDuajKN7Hy6keP9VbEBpS1WWCw93cnOGU5jq
fPu7SdkshqK6I2yCn4pmU+1w+3c61a+IeBp/ks8zp6CssdvbmjUT/l7jGB5h8kHC
RNN6VBoZEzrzMAZRfekHUPVhtA/tQsUXIhLGVxPopigR6Gk4K05KZxR87KyBMPjh
0GH99M7k/5P5ZIL9b6hMCDR481tlXk7q+XqlbwO82XaBFzEcuwVL4f52ZOQF5FwJ
mS59HlIPAgMBAAECggEAAKLV3hZ2v7UrkqQTlMO50+X0WI3YAK8Yh4yedTgzPDQ0
0KD8FMaC6HrmvGhXNfDMRmIIwD8Ew1qDjzbEieIRoD2+LXTivwf6c34HidmplEfs
K2IezKin/zuArgNio2ndUlGxt4sRnN373x5/sGZjQWcYayLSmgRN5kByuhFco0Qa
oSrXcXNUlb+KgRQXPDU4+M35tPHvLdyg+tko/m/5uK9dc9MNvGZHOMBKg0VNURJb
V1l3dR+evwvpqHzBvWiqN/YOiUUvIxlFKA35hJkfCl7ivFs4CLqqFNCKDao95fWe
s0UR9iMakT48jXV76IfwZnyX10OhIWzKls5trjDL8QKBgQD3thQJ8e0FL9y1W+Ph
mCdEaoffSPkgSn64wIsQ9bMmv4y+KYBK5AhqaHgYm4LgW4x1+CURNFu+YFEyaNNA
kNCXFyRX3Em3vxcShP5jIqg+f07mtXPKntWP/zBeKQWgdHX371oFTfaAlNuKX/7S
n0jBYjr4Iof1bnquMQvUoHCYWwKBgQDnntFU9/AQGaQIvhfeU1XKFkQ/BfhIsd27
RlmiCi0ee9Ce74cMAhWr/9yg0XUxzrh+Ui1xnkMVTZ5P8tWIxROokznLUTGJA5rs
zB+ovCPFZcquTwNzn7SBnpHTR0OqJd8sd89P5ST2SqufeSF/gGi5sTs4EocOLCpZ
EPVIfm47XQKBgB4d5RHQeCDJUOw739jtxthqm1pqZN+oLwAHaOEG/mEXqOT15sM0
NlG5oeBcB+1/M/Sj1t3gn8blrvmSBR00fifgiGqmPdA5S3TU9pjW/d2bXNxv80QP
S6fWPusz0ZtQjYc3cppygCXh808/nJu/AfmBF+pTSHRumjvTery/RPFBAoGBAMi/
zCta4cTylEvHhqR5kiefePMu120aTFYeuV1KeKStJ7o5XNE5lVMIZk80e+D5jMpf
q2eIhhgWuBoPHKh4N3uqbzMbYlWgvEx09xOmTVKv0SWW8iTqzOZza2y1nZ4BSRcf
mJ1ku86EFZAYysHZp+saA3usA0ZzXRjpK87zVdM5AoGBAOSqI+t48PnPtaUDFdpd
taNNVDbcecJatm3w8VDWnarahfWe66FIqc9wUkqekqAgwZLa0AGdUalvXfGrHfNs
PtvuNc5EImfSkuPBYLBslNxtjbBvAYgacEdY+gRhn2TeIUApnND58lCWsKbNHLFZ
ajIPbTY+Fe9OTOFTN48ujXNn
-----END PRIVATE KEY-----`)

func TestKeepConnectionWhenSameConfiguration(t *testing.T) {
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	connCount := pointer[int32](0)
	srv.Config.ConnState = func(conn net.Conn, state http.ConnState) {
		if state == http.StateNew {
			atomic.AddInt32(connCount, 1)
		}
	}

	cert, err := tls.X509KeyPair(LocalhostCert, LocalhostKey)
	require.NoError(t, err)

	srv.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}
	srv.StartTLS()

	transportManager := NewTransportManager(nil)

	dynamicConf := map[string]*dynamic.ServersTransport{
		"test": {
			ServerName: "example.com",
			RootCAs:    []types.FileOrContent{types.FileOrContent(LocalhostCert)},
		},
	}

	for range 10 {
		transportManager.Update(dynamicConf)

		tr, err := transportManager.GetRoundTripper("test")
		require.NoError(t, err)

		client := http.Client{Transport: tr}

		resp, err := client.Get(srv.URL)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}

	count := atomic.LoadInt32(connCount)
	require.EqualValues(t, 1, count)

	dynamicConf = map[string]*dynamic.ServersTransport{
		"test": {
			ServerName: "www.example.com",
			RootCAs:    []types.FileOrContent{types.FileOrContent(LocalhostCert)},
		},
	}

	transportManager.Update(dynamicConf)

	tr, err := transportManager.GetRoundTripper("test")
	require.NoError(t, err)

	client := http.Client{Transport: tr}

	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	count = atomic.LoadInt32(connCount)
	assert.EqualValues(t, 2, count)
}

func TestMTLS(t *testing.T) {
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	cert, err := tls.X509KeyPair(LocalhostCert, LocalhostKey)
	require.NoError(t, err)

	clientPool := x509.NewCertPool()
	clientPool.AppendCertsFromPEM(mTLSCert)

	srv.TLS = &tls.Config{
		// For TLS
		Certificates: []tls.Certificate{cert},

		// For mTLS
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  clientPool,
	}
	srv.StartTLS()

	transportManager := NewTransportManager(nil)

	dynamicConf := map[string]*dynamic.ServersTransport{
		"test": {
			ServerName: "example.com",
			// For TLS
			RootCAs: []types.FileOrContent{types.FileOrContent(LocalhostCert)},

			// For mTLS
			Certificates: traefiktls.Certificates{
				traefiktls.Certificate{
					CertFile: types.FileOrContent(mTLSCert),
					KeyFile:  types.FileOrContent(mTLSKey),
				},
			},
		},
	}

	transportManager.Update(dynamicConf)

	tr, err := transportManager.GetRoundTripper("test")
	require.NoError(t, err)

	client := http.Client{Transport: tr}

	resp, err := client.Get(srv.URL)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSpiffeMTLS(t *testing.T) {
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	trustDomain := spiffeid.RequireTrustDomainFromString("spiffe://traefik.test")

	pki, err := newFakeSpiffePKI(trustDomain)
	require.NoError(t, err)

	clientSVID, err := pki.genSVID(spiffeid.RequireFromPath(trustDomain, "/client"))
	require.NoError(t, err)

	serverSVID, err := pki.genSVID(spiffeid.RequireFromPath(trustDomain, "/server"))
	require.NoError(t, err)

	serverSource := fakeSpiffeSource{
		svid:   serverSVID,
		bundle: pki.bundle,
	}

	// go-spiffe's `tlsconfig.MTLSServerConfig` (that should be used here) does not set a certificate on
	// the returned `tls.Config` and relies instead on `GetCertificate` being always called.
	// But it turns out that `StartTLS` from `httptest.Server`, enforces a default certificate
	// if no certificate is previously set on the configured TLS config.
	// It makes the test server always serve the httptest default certificate, and not the SPIFFE certificate,
	// as GetCertificate is in that case never called (there's a default cert, and SNI is not used).
	// To bypass this issue, we're manually extracting the server ceritificate from the server SVID
	// and use another initialization method that forces serving the server SPIFFE certificate.
	serverCert, err := tlsconfig.GetCertificate(&serverSource)(nil)
	require.NoError(t, err)
	srv.TLS = tlsconfig.MTLSWebServerConfig(
		serverCert,
		&serverSource,
		tlsconfig.AuthorizeAny(),
	)
	srv.StartTLS()
	defer srv.Close()

	clientSource := fakeSpiffeSource{
		svid:   clientSVID,
		bundle: pki.bundle,
	}

	testCases := []struct {
		desc           string
		config         dynamic.Spiffe
		clientSource   SpiffeX509Source
		wantStatusCode int
		wantError      bool
	}{
		{
			desc:           "supports SPIFFE mTLS",
			config:         dynamic.Spiffe{},
			clientSource:   &clientSource,
			wantStatusCode: http.StatusOK,
		},
		{
			desc: "allows expected server SPIFFE ID",
			config: dynamic.Spiffe{
				IDs: []string{"spiffe://traefik.test/server"},
			},
			clientSource:   &clientSource,
			wantStatusCode: http.StatusOK,
		},
		{
			desc: "blocks unexpected server SPIFFE ID",
			config: dynamic.Spiffe{
				IDs: []string{"spiffe://traefik.test/not-server"},
			},
			clientSource: &clientSource,
			wantError:    true,
		},
		{
			desc: "allows expected server trust domain",
			config: dynamic.Spiffe{
				TrustDomain: "spiffe://traefik.test",
			},
			clientSource:   &clientSource,
			wantStatusCode: http.StatusOK,
		},
		{
			desc: "denies unexpected server trust domain",
			config: dynamic.Spiffe{
				TrustDomain: "spiffe://not-traefik.test",
			},
			clientSource: &clientSource,
			wantError:    true,
		},
		{
			desc: "spiffe IDs allowlist takes precedence",
			config: dynamic.Spiffe{
				IDs:         []string{"spiffe://traefik.test/not-server"},
				TrustDomain: "spiffe://not-traefik.test",
			},
			clientSource: &clientSource,
			wantError:    true,
		},
		{
			desc:         "raises an error when spiffe is enabled on the transport but no workloadapi address is given",
			config:       dynamic.Spiffe{},
			clientSource: nil,
			wantError:    true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			transportManager := NewTransportManager(test.clientSource)

			dynamicConf := map[string]*dynamic.ServersTransport{
				"test": {
					Spiffe: &test.config,
				},
			}

			transportManager.Update(dynamicConf)

			tr, err := transportManager.GetRoundTripper("test")
			require.NoError(t, err)

			client := http.Client{Transport: tr}

			resp, err := client.Get(srv.URL)
			if test.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.wantStatusCode, resp.StatusCode)
		})
	}
}

func TestDisableHTTP2(t *testing.T) {
	testCases := []struct {
		desc          string
		disableHTTP2  bool
		serverHTTP2   bool
		expectedProto string
	}{
		{
			desc:          "HTTP1 capable client with HTTP1 server",
			disableHTTP2:  true,
			expectedProto: "HTTP/1.1",
		},
		{
			desc:          "HTTP1 capable client with HTTP2 server",
			disableHTTP2:  true,
			serverHTTP2:   true,
			expectedProto: "HTTP/1.1",
		},
		{
			desc:          "HTTP2 capable client with HTTP1 server",
			expectedProto: "HTTP/1.1",
		},
		{
			desc:          "HTTP2 capable client with HTTP2 server",
			serverHTTP2:   true,
			expectedProto: "HTTP/2.0",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			}))

			srv.EnableHTTP2 = test.serverHTTP2
			srv.StartTLS()

			transportManager := NewTransportManager(nil)

			dynamicConf := map[string]*dynamic.ServersTransport{
				"test": {
					DisableHTTP2:       test.disableHTTP2,
					InsecureSkipVerify: true,
				},
			}

			transportManager.Update(dynamicConf)

			tr, err := transportManager.GetRoundTripper("test")
			require.NoError(t, err)

			client := http.Client{Transport: tr}

			resp, err := client.Get(srv.URL)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, test.expectedProto, resp.Proto)
		})
	}
}

// fakeSpiffePKI simulates a SPIFFE aware PKI and allows generating multiple valid SVIDs.
type fakeSpiffePKI struct {
	caPrivateKey *rsa.PrivateKey

	bundle *x509bundle.Bundle
}

func newFakeSpiffePKI(trustDomain spiffeid.TrustDomain) (fakeSpiffePKI, error) {
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fakeSpiffePKI{}, err
	}

	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2000),
		Subject: pkix.Name{
			Organization: []string{"spiffe"},
		},
		URIs:         []*url.URL{spiffeid.RequireFromPath(trustDomain, "/ca").URL()},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		SubjectKeyId: []byte("ca"),
		KeyUsage: x509.KeyUsageCertSign |
			x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		PublicKey:             caPrivateKey.Public(),
	}

	caCertDER, err := x509.CreateCertificate(
		rand.Reader,
		&caTemplate,
		&caTemplate,
		caPrivateKey.Public(),
		caPrivateKey,
	)
	if err != nil {
		return fakeSpiffePKI{}, err
	}

	bundle, err := x509bundle.ParseRaw(
		trustDomain,
		caCertDER,
	)
	if err != nil {
		return fakeSpiffePKI{}, err
	}

	return fakeSpiffePKI{
		bundle:       bundle,
		caPrivateKey: caPrivateKey,
	}, nil
}

func (f *fakeSpiffePKI) genSVID(id spiffeid.ID) (*x509svid.SVID, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(200001),
		URIs:         []*url.URL{id.URL()},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		SubjectKeyId: []byte("svid"),
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageKeyAgreement |
			x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
		PublicKey:             privateKey.PublicKey,
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		f.bundle.X509Authorities()[0],
		privateKey.Public(),
		f.caPrivateKey,
	)
	if err != nil {
		return nil, err
	}

	keyPKCS8, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return x509svid.ParseRaw(certDER, keyPKCS8)
}

// fakeSpiffeSource allows retrieving statically an SVID and its associated bundle.
type fakeSpiffeSource struct {
	bundle *x509bundle.Bundle
	svid   *x509svid.SVID
}

func (s *fakeSpiffeSource) GetX509BundleForTrustDomain(trustDomain spiffeid.TrustDomain) (*x509bundle.Bundle, error) {
	return s.bundle, nil
}

func (s *fakeSpiffeSource) GetX509SVID() (*x509svid.SVID, error) {
	return s.svid, nil
}

type roundTripperFn func(req *http.Request) (*http.Response, error)

func (r roundTripperFn) RoundTrip(request *http.Request) (*http.Response, error) {
	return r(request)
}

func TestKerberosRoundTripper(t *testing.T) {
	testCases := []struct {
		desc string

		originalRoundTripperHeaders map[string][]string

		expectedStatusCode     []int
		expectedDedicatedCount int
		expectedOriginalCount  int
	}{
		{
			desc:                  "without special header",
			expectedStatusCode:    []int{http.StatusUnauthorized, http.StatusUnauthorized, http.StatusUnauthorized},
			expectedOriginalCount: 3,
		},
		{
			desc:                        "with Negotiate (Kerberos)",
			originalRoundTripperHeaders: map[string][]string{"Www-Authenticate": {"Negotiate"}},
			expectedStatusCode:          []int{http.StatusUnauthorized, http.StatusOK, http.StatusOK},
			expectedOriginalCount:       1,
			expectedDedicatedCount:      2,
		},
		{
			desc:                        "with NTLM",
			originalRoundTripperHeaders: map[string][]string{"Www-Authenticate": {"NTLM"}},
			expectedStatusCode:          []int{http.StatusUnauthorized, http.StatusOK, http.StatusOK},
			expectedOriginalCount:       1,
			expectedDedicatedCount:      2,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			origCount := 0
			dedicatedCount := 0
			rt := kerberosRoundTripper{
				new: func() http.RoundTripper {
					return roundTripperFn(func(req *http.Request) (*http.Response, error) {
						dedicatedCount++
						return &http.Response{
							StatusCode: http.StatusOK,
						}, nil
					})
				},
				OriginalRoundTripper: roundTripperFn(func(req *http.Request) (*http.Response, error) {
					origCount++
					return &http.Response{
						StatusCode: http.StatusUnauthorized,
						Header:     test.originalRoundTripperHeaders,
					}, nil
				}),
			}

			ctx := AddTransportOnContext(t.Context())
			for _, expected := range test.expectedStatusCode {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1", http.NoBody)
				require.NoError(t, err)
				resp, err := rt.RoundTrip(req)
				require.NoError(t, err)
				require.Equal(t, expected, resp.StatusCode)
			}

			require.Equal(t, test.expectedOriginalCount, origCount)
			require.Equal(t, test.expectedDedicatedCount, dedicatedCount)
		})
	}
}
