package tls

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/containous/traefik/v2/pkg/tls/generate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecryptCertificatePrivateKey(t *testing.T) {
	testCases := []struct {
		desc             string
		encryptedKey     bool
		passphrase       string
		expectedCertLoad bool
	}{
		{
			desc:             "Encrypted Key with Correct Passphrase",
			encryptedKey:     true,
			passphrase:       "test",
			expectedCertLoad: true,
		},
		{
			desc:             "Encrypted Key with Incorrect Passphrase",
			encryptedKey:     true,
			passphrase:       "bacon",
			expectedCertLoad: false,
		},
		{
			desc:             "Non Encrypted Key with Passphrase",
			encryptedKey:     false,
			passphrase:       "test",
			expectedCertLoad: true,
		},
		{
			desc:             "Non Encrypted Key",
			encryptedKey:     false,
			expectedCertLoad: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			// Create a keypair.
			cert, key, err := generate.KeyPair("foo.bar", time.Time{})
			require.NoError(t, err)

			if test.encryptedKey {
				// Encrypt the key with passphrase "test".
				key, err = generateEncryptedKey(key, "test")
				require.NoError(t, err)
			}

			certs := Certificates{
				{
					CertFile:   FileOrContent(cert),
					KeyFile:    FileOrContent(key),
					Passphrase: test.passphrase,
				},
			}

			config, err := certs.CreateTLSConfig("test")
			assert.NoError(t, err)
			if test.expectedCertLoad {
				assert.Equal(t, 1, len(config.Certificates))
			} else {
				assert.Equal(t, 0, len(config.Certificates))
			}
		})
	}
}

func generateEncryptedKey(data []byte, pwd string) ([]byte, error) {
	var block *pem.Block
	// Convert the data to pem block.
	block, _ = pem.Decode(data)

	// Encrypt the block.
	if pwd != "" {
		var err error
		block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(pwd), x509.PEMCipherAES256)
		if err != nil {
			return nil, err
		}
	}

	return pem.EncodeToMemory(block), nil
}
