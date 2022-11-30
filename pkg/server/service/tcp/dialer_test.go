package tcp

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
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

func fakeRedis(t *testing.T, listener net.Listener, wait time.Duration) {
	t.Helper()

	for {
		time.Sleep(wait)

		conn, err := listener.Accept()
		fmt.Println("Accept on server")
		require.NoError(t, err)

		for {
			withErr := false
			buf := make([]byte, 64)
			if _, err := conn.Read(buf); err != nil {
				withErr = true
			}

			if string(buf[:4]) == "ping" {
				if _, err := conn.Write([]byte("PONG")); err != nil {
					_ = conn.Close()
					return
				}
			}

			if withErr {
				_ = conn.Close()
				return
			}
		}
	}
}

func TestTLS(t *testing.T) {
	cert, err := tls.X509KeyPair(LocalhostCert, LocalhostKey)
	require.NoError(t, err)

	backendListener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	tlsListener := tls.NewListener(backendListener, &tls.Config{Certificates: []tls.Certificate{cert}})
	defer tlsListener.Close()

	go fakeRedis(t, tlsListener, 1*time.Millisecond)
	_, port, err := net.SplitHostPort(tlsListener.Addr().String())
	require.NoError(t, err)

	dialerManager := NewDialerManager()

	dynamicConf := map[string]*dynamic.TCPServersTransport{
		"test": {
			ServerName: "example.com",
			RootCAs:    []traefiktls.FileOrContent{traefiktls.FileOrContent(LocalhostCert)},
		},
	}

	dialerManager.Update(dynamicConf)

	dialer, err := dialerManager.Get("test")
	require.NoError(t, err)

	conn, err := dialer.Dial("tcp", ":"+port)
	require.NoError(t, err)

	_, err = conn.Write([]byte("ping\n"))
	require.NoError(t, err)

	err = conn.(*tls.Conn).CloseWrite()
	require.NoError(t, err)

	var buf []byte
	buffer := bytes.NewBuffer(buf)
	n, err := io.Copy(buffer, conn)
	fmt.Println(buffer.String())
	require.NoError(t, err)

	assert.Equal(t, int64(4), n)
	assert.Equal(t, "PONG", buffer.String())
}

func TestTLSWithInsecureSkipVerify(t *testing.T) {
	cert, err := tls.X509KeyPair(LocalhostCert, LocalhostKey)
	require.NoError(t, err)

	backendListener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	tlsListener := tls.NewListener(backendListener, &tls.Config{Certificates: []tls.Certificate{cert}})
	defer tlsListener.Close()

	go fakeRedis(t, tlsListener, 1*time.Millisecond)
	_, port, err := net.SplitHostPort(tlsListener.Addr().String())
	require.NoError(t, err)

	dialerManager := NewDialerManager()

	dynamicConf := map[string]*dynamic.TCPServersTransport{
		"test": {
			ServerName:         "bad-domain.com",
			RootCAs:            []traefiktls.FileOrContent{traefiktls.FileOrContent(LocalhostCert)},
			InsecureSkipVerify: true,
		},
	}

	dialerManager.Update(dynamicConf)

	dialer, err := dialerManager.Get("test")
	require.NoError(t, err)

	conn, err := dialer.Dial("tcp", ":"+port)
	require.NoError(t, err)

	_, err = conn.Write([]byte("ping\n"))
	require.NoError(t, err)

	err = conn.(*tls.Conn).CloseWrite()
	require.NoError(t, err)

	var buf []byte
	buffer := bytes.NewBuffer(buf)
	n, err := io.Copy(buffer, conn)
	fmt.Println(buffer.String())
	require.NoError(t, err)

	assert.Equal(t, int64(4), n)
	assert.Equal(t, "PONG", buffer.String())
}

func TestMTLS(t *testing.T) {
	cert, err := tls.X509KeyPair(LocalhostCert, LocalhostKey)
	require.NoError(t, err)

	clientPool := x509.NewCertPool()
	clientPool.AppendCertsFromPEM(mTLSCert)

	backendListener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	tlsListener := tls.NewListener(backendListener, &tls.Config{
		// For TLS
		Certificates: []tls.Certificate{cert},

		// For mTLS
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  clientPool,
	})
	defer tlsListener.Close()

	go fakeRedis(t, tlsListener, 1*time.Millisecond)
	_, port, err := net.SplitHostPort(tlsListener.Addr().String())
	require.NoError(t, err)

	dialerManager := NewDialerManager()

	dynamicConf := map[string]*dynamic.TCPServersTransport{
		"test": {
			ServerName: "example.com",
			// For TLS
			RootCAs: []traefiktls.FileOrContent{traefiktls.FileOrContent(LocalhostCert)},

			// For mTLS
			Certificates: traefiktls.Certificates{
				traefiktls.Certificate{
					CertFile: traefiktls.FileOrContent(mTLSCert),
					KeyFile:  traefiktls.FileOrContent(mTLSKey),
				},
			},
		},
	}

	dialerManager.Update(dynamicConf)

	dialer, err := dialerManager.Get("test")
	require.NoError(t, err)

	conn, err := dialer.Dial("tcp", ":"+port)
	require.NoError(t, err)

	_, err = conn.Write([]byte("ping\n"))
	require.NoError(t, err)

	err = conn.(*tls.Conn).CloseWrite()
	require.NoError(t, err)

	var buf []byte
	buffer := bytes.NewBuffer(buf)
	n, err := io.Copy(buffer, conn)
	fmt.Println(buffer.String())
	require.NoError(t, err)

	assert.Equal(t, int64(4), n)
	assert.Equal(t, "PONG", buffer.String())
}

func TestDialTimeout(t *testing.T) { // FIXME
	backendListener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	defer backendListener.Close()

	go fakeRedis(t, backendListener, 10*time.Second)
	_, port, err := net.SplitHostPort(backendListener.Addr().String())
	require.NoError(t, err)

	dialerManager := NewDialerManager()

	dynamicConf := map[string]*dynamic.TCPServersTransport{
		"test": {
			DialTimeout: ptypes.Duration(1 * time.Millisecond),
		},
	}

	dialerManager.Update(dynamicConf)

	dialer, err := dialerManager.Get("test")
	require.NoError(t, err)

	_, err = dialer.Dial("tcp", ":"+port)
	require.Error(t, err)
}
