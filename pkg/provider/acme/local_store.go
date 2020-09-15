package acme

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
func NewLocalStore(filename string) *LocalStore {
	store := &LocalStore{filename: filename, saveDataChan: make(chan map[string]*StoredData)}
	store.listenSaveAction()
	return store
}

func (s *LocalStore) save(resolverName string, storedData *StoredData) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.storedData[resolverName] = storedData
	s.saveDataChan <- s.storedData
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

			file, err := ioutil.ReadAll(f)
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
						logger.Debugf("Deleting empty certificate %v for %v", certificate, certificate.Domain.ToStrArray())
						continue
					}
					certificates = append(certificates, certificate)
				}
				if len(certificates) < len(storedData.Certificates) {
					storedData.Certificates = certificates
					s.saveDataChan <- s.storedData
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
func (s *LocalStore) listenSaveAction() {
	safe.Go(func() {
		logger := log.WithoutContext().WithField(log.ProviderName, "acme")
		for object := range s.saveDataChan {
			data, err := json.MarshalIndent(object, "", "  ")
			if err != nil {
				logger.Error(err)
			}

			err = ioutil.WriteFile(s.filename, data, 0o600)
			if err != nil {
				logger.Error(err)
			}
		}
	})
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

// LocalChallengeStore is an implementation of the ChallengeStore in memory.
type LocalChallengeStore struct {
	storedData *StoredChallengeData
	lock       sync.RWMutex
}

// NewLocalChallengeStore initializes a new LocalChallengeStore.
func NewLocalChallengeStore() *LocalChallengeStore {
	return &LocalChallengeStore{
		storedData: &StoredChallengeData{
			HTTPChallenges: make(map[string]map[string][]byte),
			TLSChallenges:  make(map[string]*Certificate),
		},
	}
}

// GetHTTPChallengeToken Get the http challenge token from the store.
func (s *LocalChallengeStore) GetHTTPChallengeToken(token, domain string) ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.storedData.HTTPChallenges == nil {
		s.storedData.HTTPChallenges = map[string]map[string][]byte{}
	}

	if _, ok := s.storedData.HTTPChallenges[token]; !ok {
		return nil, fmt.Errorf("cannot find challenge for token %v", token)
	}

	result, ok := s.storedData.HTTPChallenges[token][domain]
	if !ok {
		return nil, fmt.Errorf("cannot find challenge for token %v", token)
	}
	return result, nil
}

// SetHTTPChallengeToken Set the http challenge token in the store.
func (s *LocalChallengeStore) SetHTTPChallengeToken(token, domain string, keyAuth []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.storedData.HTTPChallenges == nil {
		s.storedData.HTTPChallenges = map[string]map[string][]byte{}
	}

	if _, ok := s.storedData.HTTPChallenges[token]; !ok {
		s.storedData.HTTPChallenges[token] = map[string][]byte{}
	}

	s.storedData.HTTPChallenges[token][domain] = keyAuth
	return nil
}

// RemoveHTTPChallengeToken Remove the http challenge token in the store.
func (s *LocalChallengeStore) RemoveHTTPChallengeToken(token, domain string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.storedData.HTTPChallenges == nil {
		return nil
	}

	if _, ok := s.storedData.HTTPChallenges[token]; ok {
		delete(s.storedData.HTTPChallenges[token], domain)
		if len(s.storedData.HTTPChallenges[token]) == 0 {
			delete(s.storedData.HTTPChallenges, token)
		}
	}
	return nil
}

// AddTLSChallenge Add a certificate to the ACME TLS-ALPN-01 certificates storage.
func (s *LocalChallengeStore) AddTLSChallenge(domain string, cert *Certificate) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.storedData.TLSChallenges == nil {
		s.storedData.TLSChallenges = make(map[string]*Certificate)
	}

	s.storedData.TLSChallenges[domain] = cert
	return nil
}

// GetTLSChallenge Get a certificate from the ACME TLS-ALPN-01 certificates storage.
func (s *LocalChallengeStore) GetTLSChallenge(domain string) (*Certificate, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.storedData.TLSChallenges == nil {
		s.storedData.TLSChallenges = make(map[string]*Certificate)
	}

	return s.storedData.TLSChallenges[domain], nil
}

// RemoveTLSChallenge Remove a certificate from the ACME TLS-ALPN-01 certificates storage.
func (s *LocalChallengeStore) RemoveTLSChallenge(domain string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.storedData.TLSChallenges == nil {
		return nil
	}

	delete(s.storedData.TLSChallenges, domain)
	return nil
}
