package acme

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const localStorageFileName = "local_storage_test.json"

func TestLocalStoreGetAccountWithEmptyFile(t *testing.T) {
	defer os.Remove(localStorageFileName)

	s := NewLocalStore(localStorageFileName)
	r, err := s.GetAccount("traefik.wtf")

	assert.Nil(t, err)
	assert.Nil(t, r)
}

func TestLocalStoreSaveAndGetAccount(t *testing.T) {
	defer os.Remove(localStorageFileName)

	email := "some@email.com"
	s := NewLocalStore(localStorageFileName)
	err := s.SaveAccount("traefik.wtf", &Account{
		Email: email,
	})
	assert.Nil(t, err)

	r, err := s.GetAccount("traefik.wtf")
	assert.Nil(t, err)
	assert.Equal(t, r.Email, email)
}

func TestLocalStoreGetAccountReadFromFile(t *testing.T) {
	// give some time for `listenSaveAction` goroutine from prev tests
	time.Sleep(time.Millisecond * 100)
	os.Remove(localStorageFileName)

	defer os.Remove(localStorageFileName)

	email := "some42@email.com"
	filePayload := `{
  "traefik.wtf": {
    "Account": {
      "Email": "`+email+`"
    }
  }
}`
	err := ioutil.WriteFile(localStorageFileName, []byte(filePayload), 0o600)
	assert.Nil(t, err)

	s := NewLocalStore(localStorageFileName)

	r, err := s.GetAccount("traefik.wtf")
	if r == nil {
		t.Fatal("should not be nil")
	}
	assert.Nil(t, err)
	assert.Equal(t, email, r.Email)
}
