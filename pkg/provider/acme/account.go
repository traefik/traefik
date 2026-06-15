package acme

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/traefik/traefik/v2/pkg/log"
)

// Account is used to store lets encrypt registration info.
type Account struct {
	Email        string
	Registration *Resource
	PrivateKey   []byte
}

type Resource struct {
	Body acme.Account `json:"body"`
	URI  string       `json:"uri,omitempty"`
}

// NewAccount creates an account.
func NewAccount(email string) (*Account, error) {
	// Create a user. New accounts need an email and private key to start
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	return &Account{
		Email:      email,
		PrivateKey: x509.MarshalPKCS1PrivateKey(privateKey),
	}, nil
}

// GetEmail returns email.
func (a *Account) GetEmail() string {
	return a.Email
}

// GetRegistration returns lets encrypt registration resource.
func (a *Account) GetRegistration() *acme.ExtendedAccount {
	if a.Registration == nil {
		return nil
	}

	return &acme.ExtendedAccount{
		Account:  a.Registration.Body,
		Location: a.Registration.URI,
	}
}

// GetPrivateKey returns private key.
func (a *Account) GetPrivateKey() crypto.Signer {
	privateKey, err := x509.ParsePKCS1PrivateKey(a.PrivateKey)
	if err != nil {
		log.WithoutContext().WithField(log.ProviderName, "acme").
			Errorf("Cannot unmarshal private key %+v: %v", a.PrivateKey, err)
		return nil
	}

	return privateKey
}

// GetKeyType used to determine which algo to used.
func GetKeyType(ctx context.Context, value string) certcrypto.KeyType {
	logger := log.FromContext(ctx)

	switch value {
	case "EC256":
		return certcrypto.EC256
	case "EC384":
		return certcrypto.EC384
	case "RSA2048":
		return certcrypto.RSA2048
	case "RSA4096":
		return certcrypto.RSA4096
	case "RSA8192":
		return certcrypto.RSA8192
	case "":
		logger.Infof("The key type is empty. Use default key type %v.", certcrypto.RSA4096)
		return certcrypto.RSA4096
	default:
		logger.Infof("Unable to determine the key type value %q: falling back on %v.", value, certcrypto.RSA4096)
		return certcrypto.RSA4096
	}
}
