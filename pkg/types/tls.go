package types

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/traefik/traefik/v2/pkg/log"
)

// +k8s:deepcopy-gen=true

// ClientTLS holds TLS specific configurations as client
// CA, Cert and Key can be either path or file contents.
type ClientTLS struct {
	CA                 string `description:"TLS CA" json:"ca,omitempty" toml:"ca,omitempty" yaml:"ca,omitempty"`
	CAOptional         bool   `description:"TLS CA.Optional" json:"caOptional,omitempty" toml:"caOptional,omitempty" yaml:"caOptional,omitempty" export:"true"`
	Cert               string `description:"TLS cert" json:"cert,omitempty" toml:"cert,omitempty" yaml:"cert,omitempty"`
	Key                string `description:"TLS key" json:"key,omitempty" toml:"key,omitempty" yaml:"key,omitempty"`
	InsecureSkipVerify bool   `description:"TLS insecure skip verify" json:"insecureSkipVerify,omitempty" toml:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty" export:"true"`
}

// CreateTLSConfig creates a TLS config from ClientTLS structures.
func (clientTLS *ClientTLS) CreateTLSConfig(ctx context.Context) (*tls.Config, error) {
	if clientTLS == nil {
		log.FromContext(ctx).Warnf("clientTLS is nil")
		return nil, nil
	}

	caPool := x509.NewCertPool()
	clientAuth := tls.NoClientCert
	if clientTLS.CA != "" {
		var ca []byte
		if _, errCA := os.Stat(clientTLS.CA); errCA == nil {
			var err error
			ca, err = os.ReadFile(clientTLS.CA)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA. %w", err)
			}
		} else {
			ca = []byte(clientTLS.CA)
		}

		if !caPool.AppendCertsFromPEM(ca) {
			return nil, fmt.Errorf("failed to parse CA")
		}

		if clientTLS.CAOptional {
			clientAuth = tls.VerifyClientCertIfGiven
		} else {
			clientAuth = tls.RequireAndVerifyClientCert
		}
	}

	hasCert := len(clientTLS.Cert) > 0
	hasKey := len(clientTLS.Key) > 0

	if hasCert != hasKey {
		return nil, fmt.Errorf("TLS cert and key must be defined together")
	}

	var certificates []tls.Certificate

	if hasCert && hasKey {
		_, errKeyIsFile := os.Stat(clientTLS.Key)
		_, errCertIsFile := os.Stat(clientTLS.Cert)

		if errCertIsFile == nil {
			if errKeyIsFile != nil {
				return nil, fmt.Errorf("TLS cert is a file, but tls key is not")
			}

			cert, err := tls.LoadX509KeyPair(clientTLS.Cert, clientTLS.Key)
			if err != nil {
				return nil, fmt.Errorf("load TLS keypair from file: %w", err)
			}

			certificates = append(certificates, cert)
		} else {
			if errKeyIsFile == nil {
				return nil, fmt.Errorf("TLS key is a file, but tls cert is not")
			}

			cert, err := tls.X509KeyPair([]byte(clientTLS.Cert), []byte(clientTLS.Key))
			if err != nil {
				return nil, fmt.Errorf("load TLS keypair from bytes: %w", err)
			}

			certificates = append(certificates, cert)
		}
	}

	return &tls.Config{
		Certificates:       certificates,
		RootCAs:            caPool,
		InsecureSkipVerify: clientTLS.InsecureSkipVerify,
		ClientAuth:         clientAuth,
	}, nil
}
