package acme

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	acmeFile := "./acme_example.json"

	folder, prefix := filepath.Split(acmeFile)
	tmpFile, err := ioutil.TempFile(folder, prefix)
	defer os.Remove(tmpFile.Name())

	assert.NoError(t, err)

	fileContent, err := ioutil.ReadFile(acmeFile)
	assert.NoError(t, err)

	tmpFile.Write(fileContent)

	localStore := NewLocalStore(tmpFile.Name())
	account, err := localStore.Get()
	assert.NoError(t, err)

	assert.Len(t, account.DomainsCertificate.Certs, 1)
}
