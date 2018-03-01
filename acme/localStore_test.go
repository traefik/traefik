package acme

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGet(t *testing.T) {
	acmeFile := "./acme_example.json"

	folder, prefix := filepath.Split(acmeFile)
	tmpFile, err := ioutil.TempFile(folder, prefix)
	defer os.Remove(tmpFile.Name())

	if err != nil {
		t.Error(err)
	}

	fileContent, err := ioutil.ReadFile(acmeFile)
	if err != nil {
		t.Error(err)
	}

	tmpFile.Write(fileContent)

	localStore := NewLocalStore(tmpFile.Name())
	account, err := localStore.Get()
	if err != nil {
		t.Error(err)
	}

	if len(account.DomainsCertificate.Certs) != 1 {
		t.Errorf("Must found %d and found %d certificates in Account", 3, len(account.DomainsCertificate.Certs))
	}
}
