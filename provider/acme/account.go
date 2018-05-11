package acme

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"github.com/containous/traefik/log"
	acme "github.com/xenolf/lego/acmev2"
)

// Account is used to store lets encrypt registration info
type Account struct {
	Email        string
	Registration *acme.RegistrationResource
	PrivateKey   []byte
	KeyType      acme.KeyType
}

const (
	// RegistrationURLPathV1Regexp is a regexp which match ACME registration URL in the V1 format
	RegistrationURLPathV1Regexp = `^.*/acme/reg/\d+$`
)

// NewAccount creates an account
func NewAccount(email string, keyTypeValue string) (*Account, error) {
	keyType := getKeyType(keyTypeValue)

	// Create a user. New accounts need an email and private key to start
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	return &Account{
		Email:      email,
		PrivateKey: x509.MarshalPKCS1PrivateKey(privateKey),
		KeyType:    keyType,
	}, nil
}

// GetEmail returns email
func (a *Account) GetEmail() string {
	return a.Email
}

// GetRegistration returns lets encrypt registration resource
func (a *Account) GetRegistration() *acme.RegistrationResource {
	return a.Registration
}

// GetPrivateKey returns private key
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	if privateKey, err := x509.ParsePKCS1PrivateKey(a.PrivateKey); err == nil {
		return privateKey
	}

	log.Errorf("Cannot unmarshal private key %+v", a.PrivateKey)
	return nil
}

func getKeyType(value string) acme.KeyType {
	keyType := acme.RSA4096
	switch value {
	case "EC256":
		keyType = acme.EC256
	case "EC384":
		keyType = acme.EC384
	case "RSA2048":
		keyType = acme.RSA2048
	case "RSA4096":
		keyType = acme.RSA4096
	case "RSA8192":
		keyType = acme.RSA8192
	}

	log.Debugf("Using %s as key type", keyType)
	return keyType
}
