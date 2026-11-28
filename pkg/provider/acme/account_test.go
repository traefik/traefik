package acme

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"testing"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccount(t *testing.T) {
	testCases := []struct {
		desc          string
		keyType       certcrypto.KeyType
		expectedType  any
		expectedBits  int
		expectedCurve elliptic.Curve
	}{
		{
			desc:         "default RSA 4096",
			keyType:      certcrypto.RSA4096,
			expectedType: &rsa.PrivateKey{},
			expectedBits: 4096,
		},
		{
			desc:         "RSA 2048",
			keyType:      certcrypto.RSA2048,
			expectedType: &rsa.PrivateKey{},
			expectedBits: 2048,
		},
		{
			desc:          "EC 256",
			keyType:       certcrypto.EC256,
			expectedType:  &ecdsa.PrivateKey{},
			expectedCurve: elliptic.P256(),
		},
		{
			desc:          "EC 384",
			keyType:       certcrypto.EC384,
			expectedType:  &ecdsa.PrivateKey{},
			expectedCurve: elliptic.P384(),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			account, err := NewAccount("foo@foo.net", test.keyType)
			require.NoError(t, err)
			require.NotNil(t, account)

			key := account.GetPrivateKey()
			require.IsType(t, test.expectedType, key)

			switch key := key.(type) {
			case *rsa.PrivateKey:
				assert.Equal(t, test.expectedBits, key.N.BitLen())

				// RSA keys keep the historical PKCS#1 storage format.
				_, err := x509.ParsePKCS1PrivateKey(account.PrivateKey)
				assert.NoError(t, err)

			case *ecdsa.PrivateKey:
				assert.Equal(t, test.expectedCurve.Params().Name, key.Curve.Params().Name)

				// EC keys are stored in the PKCS#8 format.
				storedKey, err := x509.ParsePKCS8PrivateKey(account.PrivateKey)
				require.NoError(t, err)
				assert.IsType(t, &ecdsa.PrivateKey{}, storedKey)
			}
		})
	}
}
