package acme

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/safe"
)

func TestLocalStore_GetAccount(t *testing.T) {
	acmeFile := filepath.Join(t.TempDir(), "acme.json")

	email := "some42@email.com"
	filePayload := fmt.Sprintf(`{
  "test": {
    "Account": {
      "Email": "%s"
    }
  }
}`, email)

	err := os.WriteFile(acmeFile, []byte(filePayload), 0o600)
	require.NoError(t, err)

	testCases := []struct {
		desc     string
		filename string
		expected *Account
	}{
		{
			desc:     "empty file",
			filename: filepath.Join(t.TempDir(), "acme-empty.json"),
			expected: nil,
		},
		{
			desc:     "file with data",
			filename: acmeFile,
			expected: &Account{Email: "some42@email.com"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			s := NewLocalStore(test.filename, safe.NewPool(t.Context()))

			account, err := s.GetAccount("test")
			require.NoError(t, err)

			assert.Equal(t, test.expected, account)
		})
	}
}

func TestLocalStore_SaveAccount(t *testing.T) {
	acmeFile := filepath.Join(t.TempDir(), "acme.json")

	s := NewLocalStore(acmeFile, safe.NewPool(t.Context()))

	email := "some@email.com"

	err := s.SaveAccount("test", &Account{Email: email})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	file, err := os.ReadFile(acmeFile)
	require.NoError(t, err)

	expected := `{
  "test": {
    "Account": {
      "Email": "some@email.com",
      "Registration": null,
      "PrivateKey": null,
      "KeyType": ""
    },
    "Certificates": null
  }
}`

	assert.Equal(t, expected, string(file))
}
