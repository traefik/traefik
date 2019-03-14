package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/safe"
	"github.com/containous/traefik/pkg/types"
	vaultapi "github.com/hashicorp/vault/api"
)

var _ Store = (*VaultStore)(nil)

// VaultStore Store implementation for Vault
type VaultStore struct {
	pathPrefix   string
	storedData   *StoredData
	SaveDataChan chan *StoredData `json:"-"`
	lock         sync.RWMutex
	client       *vaultapi.Client
}

// NewVaultStore initializes a new LocalStore with a file name
func NewVaultStore(pathPrefix string) *VaultStore {
	store := &VaultStore{pathPrefix: pathPrefix, SaveDataChan: make(chan *StoredData)}
	store.listenSaveAction()
	return store
}

func (s *VaultStore) get() (*StoredData, error) {
	if s.storedData == nil {
		logger := log.WithoutContext().WithField(log.ProviderName, "vault")

		client, err := s.getClient()
		if err != nil {
			return nil, err
		}

		// get a list of all certificate secrets under our prefix
		secrets, err := client.Logical().List(fmt.Sprintf("/secret/metadata/%s", s.pathPrefix))
		if err != nil {
			return nil, err
		}
		var certificates []*Certificate
		// if we have a secret, we iterate over all and add them to the certificates list
		if secrets != nil {
			for _, key := range secrets.Data["keys"].([]string) {
				// fetch the certificate from vault
				secret, err := client.Logical().Read(fmt.Sprintf("/secret/data/%s/%s/certificate", s.pathPrefix, key))
				if err != nil {
					logger.Errorf("Failed to get certificate for %v: %v", key, err)
					continue
				}

				var sans []string
				err = json.Unmarshal(secret.Data["SANS"].([]byte), &sans)
				if err != nil {
					logger.Warnf("Failed to parse SANS for %v: %v", key, secret.Data["SANS"])
				}

				certificate := &Certificate{
					Domain: types.Domain{
						Main: key,
						SANs: sans,
					},
					Certificate: secret.Data["certificate"].(string),
					Key:         secret.Data["key"].(string),
				}

				if len(certificate.Certificate) == 0 || len(certificate.Key) == 0 {
					logger.Debugf("Ignoring empty certificate %v for %v", certificate, certificate.Domain.ToStrArray())
					continue
				}
				certificates = append(certificates, certificate)
			}
		}

		s.storedData = &StoredData{
			Certificates: certificates,
		}
	}

	return s.storedData, nil
}

// GetCertificates returns Vault Certificates list
func (s *VaultStore) GetCertificates() ([]*Certificate, error) {
	storedData, err := s.get()
	if err != nil {
		return nil, err
	}

	return storedData.Certificates, nil
}

// SaveCertificates stores Vault Certificates list
func (s *VaultStore) SaveCertificates(certificates []*Certificate) error {
	storedData := &StoredData{
		Certificates: certificates,
	}
	s.SaveDataChan <- storedData

	return nil
}

// listenSaveAction listens to a chan to store Certificate data in vault
func (s *VaultStore) listenSaveAction() {
	safe.Go(func() {
		logger := log.WithoutContext().WithField(log.ProviderName, "vault")
		for object := range s.SaveDataChan {
			// get the client
			client, err := s.getClient()
			if err != nil {
				logger.Error(err)
			}
			// iterate over all certificates and save them
			for _, certificate := range object.Certificates {
				// create our secret
				sans, err := json.Marshal(certificate.Domain.SANs)
				if err != nil {
					logger.Error(err)
					continue
				}
				secret := map[string]interface{}{
					"SANS":        sans,
					"certificate": certificate.Certificate,
					"key":         certificate.Key,
				}

				_, err = client.Logical().Write(fmt.Sprintf("/secret/data/%s/%s/certificate", s.pathPrefix, certificate.Domain.Main), secret)
				if err != nil {
					logger.Error(err)
				}
			}
		}
	})
}

// get the vault client
func (s *VaultStore) getClient() (*vaultapi.Client, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	ctx := log.With(context.Background(), log.Str(log.ProviderName, "vault"))
	logger := log.FromContext(ctx)

	if s.client != nil {
		return s.client, nil
	}

	logger.Debug("Building Vault client...")

	vaultConfig := vaultapi.DefaultConfig()
	// TODO: allow the user to configure all paremters within the config file
	// instead of the environment

	client, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}

	s.client = client
	return s.client, nil
}
