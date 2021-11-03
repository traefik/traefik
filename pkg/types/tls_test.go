package types

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// go run $GOROOT/src/crypto/tls/generate_cert.go --rsa-bits 1024 --host localhost --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var cert = `-----BEGIN CERTIFICATE-----
MIIB9jCCAV+gAwIBAgIQI3edJckNbicw4WIHs5Ws9TANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQCb8oWyME1QRBoMLFei3M8TVKwfZfW74cVjtcugCBMTTOTCouEIgjjmiMv6
FdMio2uBcgeD9R3dOtjjnA7N+xjwZ4vIPqDlJRE3YbfpV9igVX3sXU7ssHTSH0vs
R0TuYJwGReIFUnu5QIjGwVorodF+CQ8dTnyXVLeQVU9kvjohHwIDAQABo0swSTAO
BgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYDVR0TAQH/BAIw
ADAUBgNVHREEDTALgglsb2NhbGhvc3QwDQYJKoZIhvcNAQELBQADgYEADqylUQ/4
lrxh4h8UUQ2wKATQ2kG2YvMGlaIhr2vPZo2QDBlmL2xzai7YXX3+JZyM15TNCamn
WtFR7WQIOHzKA1GkR9WkaXKmFbJjhGMSZVCG6ghhTjzB+stBYZXhBsdjCJbkZWBu
OeI73oivo0MdI+4iCYCo7TnoY4PZGObwcgI=
-----END CERTIFICATE-----`

var key = `-----BEGIN PRIVATE KEY-----
MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAJvyhbIwTVBEGgws
V6LczxNUrB9l9bvhxWO1y6AIExNM5MKi4QiCOOaIy/oV0yKja4FyB4P1Hd062OOc
Ds37GPBni8g+oOUlETdht+lX2KBVfexdTuywdNIfS+xHRO5gnAZF4gVSe7lAiMbB
Wiuh0X4JDx1OfJdUt5BVT2S+OiEfAgMBAAECgYA9+PbghQl0aFvhko2RDybLi86K
+73X2DTVFx3AjvTlqp0OLCQ5eWabVqmYzKuHDGJgoqwR6Irhq80dRpsriCm0YNui
mMV35bbimOKz9FoCTKx0ZB6xsqrVoFhjVmX3DOD9Txe41H42ZxmccOKZndR/QaXz
VV+1W/Wbz2VawnkyYQJBAMvF6w2eOJRRoN8e7GM7b7uqkupJPp9axgFREoJZb16W
mqXUZnH4Cydzc5keG4yknQRHdgz6RrQxnvR7GyKHLfUCQQDD6qG9D5BX0+mNW6TG
PRwW/L2qWgnmg9lxtSSQat9ZOnBhw2OLPi0zTu4p70oSmU67/YJr50HEoJpRccZJ
mnJDAkBdBTtY2xpe8qhqUjZ80hweYi5wzwDMQ+bRoQ2+/U6usjdkbgJaEm4dE0H4
6tqOqHKZCnokUHfIOEKkvjHT4DulAkBAgiJNSTGi6aDOLa28pGR6YS/mRo1Z/HH9
kcJ/VuFB1Q8p8Zb2QzvI2CVtY2AFbbtSBPALrXKnVqZZSNgcZiFXAkEAvcLKaEXE
haGMGwq2BLADPHqAR3hdCJL3ikMJwWUsTkTjm973iEIEZfF5j57EzRI4bASm4Zq5
Zt3BcblLODQ//w==
-----END PRIVATE KEY-----`

func TestClientTLS_CreateTLSConfig(t *testing.T) {
	tests := []struct {
		desc        string
		clientTLS   ClientTLS
		wantCertLen int
		wantCALen   int
		wantErr     bool
	}{
		{
			desc:      "Configure CA",
			clientTLS: ClientTLS{CA: cert},
			wantCALen: 1,
			wantErr:   false,
		},
		{
			desc:        "Configure the client keyPair from strings",
			clientTLS:   ClientTLS{Cert: cert, Key: key},
			wantCertLen: 1,
			wantErr:     false,
		},
		{
			desc:        "Configure the client keyPair from files",
			clientTLS:   ClientTLS{Cert: "fixtures/cert.pem", Key: "fixtures/key.pem"},
			wantCertLen: 1,
			wantErr:     false,
		},
		{
			desc:      "Configure InsecureSkipVerify",
			clientTLS: ClientTLS{InsecureSkipVerify: true},
			wantErr:   false,
		},
		{
			desc:      "Return an error if only the client cert is provided",
			clientTLS: ClientTLS{Cert: cert},
			wantErr:   true,
		},
		{
			desc:      "Return an error if only the client key is provided",
			clientTLS: ClientTLS{Key: key},
			wantErr:   true,
		},
		{
			desc:      "Return an error if only the client cert is of type file",
			clientTLS: ClientTLS{Cert: "fixtures/cert.pem", Key: key},
			wantErr:   true,
		},
		{
			desc:      "Return an error if only the client key is of type file",
			clientTLS: ClientTLS{Cert: cert, Key: "fixtures/key.pem"},
			wantErr:   true,
		},
		{
			desc:      "Return an error if the client cert does not exist",
			clientTLS: ClientTLS{Cert: "fixtures/cert2.pem", Key: "fixtures/key.pem"},
			wantErr:   true,
		},
		{
			desc:      "Return an error if the client key does not exist",
			clientTLS: ClientTLS{Cert: "fixtures/cert.pem", Key: "fixtures/key2.pem"},
			wantErr:   true,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			tlsConfig, err := test.clientTLS.CreateTLSConfig(context.Background())
			if test.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Len(t, tlsConfig.Certificates, test.wantCertLen)
			assert.Equal(t, test.clientTLS.InsecureSkipVerify, tlsConfig.InsecureSkipVerify)

			if test.wantCALen > 0 {
				assert.Len(t, tlsConfig.RootCAs.Subjects(), test.wantCALen)
				return
			}

			assert.Nil(t, tlsConfig.RootCAs)
		})
	}
}
