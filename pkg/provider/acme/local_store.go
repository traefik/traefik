package acme

import (
	"context"
	"encoding/json"
	"io"
	"maps"
	"os"
	"sync"

	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/safe"
)

var _ Store = (*LocalStore)(nil)

// LocalStore Stores implementation for local file.
type LocalStore struct {
	saveDataChan chan map[string]*StoredData
	filename     string

	lock       sync.RWMutex
	storedData map[string]*StoredData
}

// NewLocalStore initializes a new LocalStore with a file name.
func NewLocalStore(filename string, routinesPool *safe.Pool) *LocalStore {
	store := &LocalStore{filename: filename, saveDataChan: make(chan map[string]*StoredData)}
	store.listenSaveAction(routinesPool)
	return store
}

// GetAccount returns ACME Account.
func (s *LocalStore) GetAccount(resolverName string) (*Account, error) {
	storedData, err := s.get(resolverName)
	if err != nil {
		return nil, err
	}

	return storedData.Account, nil
}

// SaveAccount stores ACME Account.
func (s *LocalStore) SaveAccount(resolverName string, account *Account) error {
	storedData, err := s.get(resolverName)
	if err != nil {
		return err
	}

	storedData.Account = account
	s.save(resolverName, storedData)

	return nil
}

// GetCertificates returns ACME Certificates list.
func (s *LocalStore) GetCertificates(resolverName string) ([]*CertAndStore, error) {
	storedData, err := s.get(resolverName)
	if err != nil {
		return nil, err
	}

	return storedData.Certificates, nil
}

// SaveCertificates stores ACME Certificates list.
func (s *LocalStore) SaveCertificates(resolverName string, certificates []*CertAndStore) error {
	storedData, err := s.get(resolverName)
	if err != nil {
		return err
	}

	storedData.Certificates = certificates
	s.save(resolverName, storedData)

	return nil
}

func (s *LocalStore) save(resolverName string, storedData *StoredData) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.storedData[resolverName] = storedData
	s.saveDataChan <- s.unSafeCopyOfStoredData()
}

func (s *LocalStore) get(resolverName string) (*StoredData, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.storedData == nil {
		s.storedData = map[string]*StoredData{}

		hasData, err := CheckFile(s.filename)
		if err != nil {
			return nil, err
		}

		if hasData {
			logger := log.WithoutContext().WithField(log.ProviderName, "acme")

			f, err := os.Open(s.filename)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			file, err := io.ReadAll(f)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(file, &s.storedData); err != nil {
				return nil, err
			}

			// Delete all certificates with no value
			for _, storedData := range s.storedData {
				for key, cert := range storedData.Certificates {
					if cert == nil {
						logger.Debugf("Deleting empty certificate %q for resolver %q", key, resolverName)
						storedData.Certificates = append(storedData.Certificates[:key], storedData.Certificates[key+1:]...)
					}
				}
			}
		}
	}

	if s.storedData[resolverName] == nil {
		s.storedData[resolverName] = &StoredData{}
	}
	return s.storedData[resolverName], nil
}

func (s *LocalStore) listenSaveAction(routinesPool *safe.Pool) {
	routinesPool.GoCtx(func(ctx context.Context) {
		logger := log.FromContext(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case object := <-s.saveDataChan:
				data, err := json.MarshalIndent(object, "", "  ")
				if err != nil {
					logger.Error(err)
				}

				if err := os.WriteFile(s.filename, data, 0o600); err != nil {
					logger.WithField(log.ProviderName, "acme").Error(err)
				}
			}
		}
	})
}

func (s *LocalStore) unSafeCopyOfStoredData() map[string]*StoredData {
	return maps.Clone(s.storedData)
}
