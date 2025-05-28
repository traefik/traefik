package types

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

// +k8s:deepcopy-gen=true

// ClientTLS holds TLS specific configurations as client
// CA, Cert and Key can be either path or file contents.
type ClientTLS struct {
	CA                 string `description:"TLS CA" json:"ca,omitempty" toml:"ca,omitempty" yaml:"ca,omitempty"`
	Cert               string `description:"TLS cert" json:"cert,omitempty" toml:"cert,omitempty" yaml:"cert,omitempty"`
	Key                string `description:"TLS key" json:"key,omitempty" toml:"key,omitempty" yaml:"key,omitempty" loggable:"false"`
	InsecureSkipVerify bool   `description:"TLS insecure skip verify" json:"insecureSkipVerify,omitempty" toml:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty" export:"true"`
}

// CreateTLSConfig creates a TLS config from ClientTLS structures.
func (c *ClientTLS) CreateTLSConfig(ctx context.Context) (*tls.Config, error) {
	if c == nil {
		log.Ctx(ctx).Warn().Msg("clientTLS is nil")
		return nil, nil
	}

	// Not initialized, to rely on system bundle.
	var caPool *x509.CertPool

	if c.CA != "" {
		var ca []byte
		if _, errCA := os.Stat(c.CA); errCA == nil {
			var err error
			ca, err = os.ReadFile(c.CA)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA. %w", err)
			}
		} else {
			ca = []byte(c.CA)
		}

		caPool = x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(ca) {
			return nil, errors.New("failed to parse CA")
		}
	}

	hasCert := len(c.Cert) > 0
	hasKey := len(c.Key) > 0

	if hasCert != hasKey {
		return nil, errors.New("both TLS cert and key must be defined")
	}

	if !hasCert || !hasKey {
		return &tls.Config{
			RootCAs:            caPool,
			InsecureSkipVerify: c.InsecureSkipVerify,
		}, nil
	}

	cert, err := loadKeyPair(c.Cert, c.Key)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caPool,
		InsecureSkipVerify: c.InsecureSkipVerify,
	}, nil
}

func loadKeyPair(cert, key string) (tls.Certificate, error) {
	keyPair, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err == nil {
		return keyPair, nil
	}

	_, err = os.Stat(cert)
	if err != nil {
		return tls.Certificate{}, errors.New("cert file does not exist")
	}

	_, err = os.Stat(key)
	if err != nil {
		return tls.Certificate{}, errors.New("key file does not exist")
	}

	keyPair, err = tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	return keyPair, nil
}
