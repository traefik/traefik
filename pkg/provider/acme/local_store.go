package acme

import (
	"context"
	"encoding/json"
	"io"
	"maps"
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/safe"
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

	// we cannot pass s.storedData directly, map is reference type and as result
	// we can face with race condition, so we need to work with objects copy
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
			logger := log.With().Str(logs.ProviderName, "acme").Logger()

			f, err := os.Open(s.filename)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			file, err := io.ReadAll(f)
			if err != nil {
				return nil, err
			}

			if len(file) > 0 {
				if err := json.Unmarshal(file, &s.storedData); err != nil {
					return nil, err
				}
			}

			// Delete all certificates with no value
			var certificates []*CertAndStore
			for _, storedData := range s.storedData {
				for _, certificate := range storedData.Certificates {
					if len(certificate.Certificate.Certificate) == 0 || len(certificate.Key) == 0 {
						logger.Debug().Msgf("Deleting empty certificate %v for %v", certificate, certificate.Domain.ToStrArray())
						continue
					}
					certificates = append(certificates, certificate)
				}
				if len(certificates) < len(storedData.Certificates) {
					storedData.Certificates = certificates

					// we cannot pass s.storedData directly, map is reference type and as result
					// we can face with race condition, so we need to work with objects copy
					s.saveDataChan <- s.unSafeCopyOfStoredData()
				}
			}
		}
	}

	if s.storedData[resolverName] == nil {
		s.storedData[resolverName] = &StoredData{}
	}
	return s.storedData[resolverName], nil
}

// listenSaveAction listens to a chan to store ACME data in json format into `LocalStore.filename`.
func (s *LocalStore) listenSaveAction(routinesPool *safe.Pool) {
	routinesPool.GoCtx(func(ctx context.Context) {
		logger := log.With().Str(logs.ProviderName, "acme").Logger()
		for {
			select {
			case <-ctx.Done():
				return

			case object := <-s.saveDataChan:
				select {
				case <-ctx.Done():
					// Stop handling events because Traefik is shutting down.
					return
				default:
				}

				data, err := json.MarshalIndent(object, "", "  ")
				if err != nil {
					logger.Error().Err(err).Send()
					continue
				}

				if err := writeFile(s.filename, data); err != nil {
					logger.Error().Err(err).Send()
				}
			}
		}
	})
}

// writeFile replaces filename through a temporary file and a rename,
// so that a failed or interrupted write never leaves a truncated file behind.
func writeFile(filename string, data []byte) error {
	// Replace the target of a symlink, not the symlink itself.
	if target, err := filepath.EvalSymlinks(filename); err == nil {
		filename = target
	}

	// CreateTemp creates the file with 0600 permissions.
	tmp, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".*.tmp")
	if err != nil {
		return os.WriteFile(filename, data, 0o600)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmp.Name(), filename); err != nil {
		// A file bind-mounted into a container cannot be replaced, write it in place.
		return os.WriteFile(filename, data, 0o600)
	}

	return nil
}

// unSafeCopyOfStoredData creates maps copy of storedData. Is not thread safe, you should use `s.lock`.
func (s *LocalStore) unSafeCopyOfStoredData() map[string]*StoredData {
	return maps.Clone(s.storedData)
}
