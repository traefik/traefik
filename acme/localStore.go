package acme

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider/acme"
)

// LocalStore is a store using a file as storage
type LocalStore struct {
	file string
}

// NewLocalStore create a LocalStore
func NewLocalStore(file string) *LocalStore {
	return &LocalStore{
		file: file,
	}
}

// Get loads file into store and returns the Account
func (s *LocalStore) Get() (*Account, error) {
	account := &Account{}

	hasData, err := checkFile(s.file)
	if err != nil {
		return nil, err
	}

	if hasData {
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
	}
	return account, nil
}

// ConvertToNewFormat converts old acme.json format to the new one and store the result into the file (used for the backward compatibility)
func ConvertToNewFormat(fileName string) {
	localStore := acme.NewLocalStore(fileName)
	storeAccount, err := localStore.GetAccount()
	if err != nil {
		log.Warnf("Failed to read new account, ACME data conversion is not available : %v", err)
		return
	}

	if storeAccount == nil {
		localStore := NewLocalStore(fileName)

		account, err := localStore.Get()
		if err != nil {
			log.Warnf("Failed to read old account, ACME data conversion is not available : %v", err)
			return
		}

		if account != nil {
			newAccount := &acme.Account{
				PrivateKey:   account.PrivateKey,
				Registration: account.Registration,
				Email:        account.Email,
			}

			var newCertificates []*acme.Certificate
			for _, cert := range account.DomainsCertificate.Certs {
				newCertificates = append(newCertificates, &acme.Certificate{
					Certificate: cert.Certificate.Certificate,
					Key:         cert.Certificate.PrivateKey,
					Domain:      cert.Domains,
				})
			}
			newLocalStore := acme.NewLocalStore(fileName)
			newLocalStore.SaveDataChan <- &acme.StoredData{Account: newAccount, Certificates: newCertificates}
		}
	}
}

// FromNewToOldFormat converts new acme.json format to the old one (used for the backward compatibility)
func FromNewToOldFormat(fileName string) (*Account, error) {
	localStore := acme.NewLocalStore(fileName)

	storeAccount, err := localStore.GetAccount()
	if err != nil {
		return nil, err
	}

	storeCertificates, err := localStore.GetCertificates()
	if err != nil {
		return nil, err
	}

	if storeAccount != nil {
		account := &Account{}
		account.Email = storeAccount.Email
		account.PrivateKey = storeAccount.PrivateKey
		account.Registration = storeAccount.Registration
		account.DomainsCertificate = DomainsCertificates{}

		for _, cert := range storeCertificates {
			_, err = account.DomainsCertificate.addCertificateForDomains(&Certificate{
				Domain:      cert.Domain.Main,
				Certificate: cert.Certificate,
				PrivateKey:  cert.Key,
			}, cert.Domain)
			if err != nil {
				return nil, err
			}
		}
		return account, nil
	}
	return nil, nil
}
