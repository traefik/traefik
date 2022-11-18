package acme

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/registration"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v2/pkg/logs"
)

// Account is used to store lets encrypt registration info.
type Account struct {
	Email        string
	Registration *registration.Resource
	PrivateKey   []byte
	KeyType      certcrypto.KeyType
}

const (
	// RegistrationURLPathV1Regexp is a regexp which match ACME registration URL in the V1 format.
	RegistrationURLPathV1Regexp = `^.*/acme/reg/\d+$`
)

// NewAccount creates an account.
func NewAccount(ctx context.Context, email, keyTypeValue string) (*Account, error) {
	keyType := GetKeyType(ctx, keyTypeValue)

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

// GetEmail returns email.
func (a *Account) GetEmail() string {
	return a.Email
}

// GetRegistration returns lets encrypt registration resource.
func (a *Account) GetRegistration() *registration.Resource {
	return a.Registration
}

// GetPrivateKey returns private key.
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	privateKey, err := x509.ParsePKCS1PrivateKey(a.PrivateKey)
	if err != nil {
		log.Error().Str(logs.ProviderName, "acme").
			Err(err).Msgf("Cannot unmarshal private key %+v", a.PrivateKey)
		return nil
	}

	return privateKey
}

// GetKeyType used to determine which algo to used.
func GetKeyType(ctx context.Context, value string) certcrypto.KeyType {
	logger := log.Ctx(ctx)

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
		logger.Info().Msgf("The key type is empty. Use default key type %v.", certcrypto.RSA4096)
		return certcrypto.RSA4096
	default:
		logger.Info().Msgf("Unable to determine the key type value %q: falling back on %v.", value, certcrypto.RSA4096)
		return certcrypto.RSA4096
	}
}
