package acme

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/log"
)

var _ cluster.Store = (*LocalStore)(nil)

// LocalStore is a store using a file as storage
type LocalStore struct {
	file        string
	storageLock sync.RWMutex
	account     *Account
}

// NewLocalStore create a LocalStore
func NewLocalStore(file string) *LocalStore {
	return &LocalStore{
		file: file,
	}
}

// Get atomically a struct from the file storage
func (s *LocalStore) Get() cluster.Object {
	s.storageLock.RLock()
	defer s.storageLock.RUnlock()
	return s.account
}

// Load loads file into store
func (s *LocalStore) Load() (cluster.Object, error) {
	s.storageLock.Lock()
	defer s.storageLock.Unlock()
	account := &Account{}

	err := checkPermissions(s.file)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(s.file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	file, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(file, &account); err != nil {
		return nil, err
	}
	account.Init()
	s.account = account
	log.Infof("Loaded ACME config from store %s", s.file)
	return account, nil
}

// Begin creates a transaction with the KV store.
func (s *LocalStore) Begin() (cluster.Transaction, cluster.Object, error) {
	s.storageLock.Lock()
	return &localTransaction{LocalStore: s}, s.account, nil
}

var _ cluster.Transaction = (*localTransaction)(nil)

type localTransaction struct {
	*LocalStore
	dirty bool
}

// Commit allows to set an object in the file storage
func (t *localTransaction) Commit(object cluster.Object) error {
	t.LocalStore.account = object.(*Account)
	defer t.storageLock.Unlock()
	if t.dirty {
		return fmt.Errorf("transaction already used, please begin a new one")
	}

	// write account to file
	data, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(t.file, data, 0600)
	if err != nil {
		return err
	}
	t.dirty = true
	return nil
}
