package acme

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/containous/traefik/log"
)

var _ Store = (*LocalStore)(nil)

// LocalStore Store implementation for local file
type LocalStore struct {
	filename       string
	httpChallenges map[string]map[string][]byte
}

// NewLocalStore initializes a new LocalStore with a file name
func NewLocalStore(filename string) LocalStore {
	return LocalStore{filename: filename}
}

func (s *LocalStore) get() (StoredData, error) {
	storedData := StoredData{}
	f, err := os.Open(s.filename)
	if err != nil {
		return storedData, err
	}
	defer f.Close()
	file, err := ioutil.ReadAll(f)
	if err != nil {
		return storedData, err
	}
	if len(file) > 0 {
		if err := json.Unmarshal(file, &storedData); err != nil {
			return storedData, err
		}
	}
	return storedData, nil
}

// Save stores ACME data in json format into LocalStore.filename
func (s *LocalStore) Save(object StoredData) error {
	data, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		log.Error(err)
		return err
	}
	err = ioutil.WriteFile(s.filename, data, 0600)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// GetAccount returns ACME Account
func (s *LocalStore) GetAccount() (*Account, error) {
	storedData, err := s.get()
	if err != nil {
		return nil, err
	}
	return storedData.Account, nil
}

// SaveAccount stores ACME Account
func (s *LocalStore) SaveAccount(account *Account) error {
	storedData, err := s.get()
	if err != nil {
		return err
	}
	storedData.Account = account
	return s.Save(storedData)
}

// GetCertificates returns ACME Certificates list
func (s *LocalStore) GetCertificates() ([]*Certificate, error) {
	storedData, err := s.get()
	if err != nil {
		return nil, err
	}
	return storedData.Certificates, nil
}

// SaveCertificates stores ACME Certificates list
func (s *LocalStore) SaveCertificates(certificates []*Certificate) error {
	storedData, err := s.get()
	if err != nil {
		return err
	}
	storedData.Certificates = certificates
	return s.Save(storedData)
}

// GetHTTPChallenges returns ACME HTTP Challenges list
func (s *LocalStore) GetHTTPChallenges() (map[string]map[string][]byte, error) {
	return s.httpChallenges, nil
}

// SaveHTTPChallenges stores ACME HTTP Challenges list
func (s *LocalStore) SaveHTTPChallenges(httpChallenges map[string]map[string][]byte) error {
	s.httpChallenges = httpChallenges
	return nil
}
