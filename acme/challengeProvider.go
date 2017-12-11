package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/tls/generate"
	"github.com/xenolf/lego/acme"
)

var _ acme.ChallengeProviderTimeout = (*challengeProvider)(nil)

type challengeProvider struct {
	store cluster.Store
	lock  sync.RWMutex
}

func (c *challengeProvider) getCertificate(domain string) (cert *tls.Certificate, exists bool) {
	log.Debugf("Challenge GetCertificate %s", domain)
	if !strings.HasSuffix(domain, ".acme.invalid") {
		return nil, false
	}
	c.lock.RLock()
	defer c.lock.RUnlock()
	account := c.store.Get().(*Account)
	if account.ChallengeCerts == nil {
		return nil, false
	}
	account.Init()
	var result *tls.Certificate
	operation := func() error {
		for _, cert := range account.ChallengeCerts {
			for _, dns := range cert.certificate.Leaf.DNSNames {
				if domain == dns {
					result = cert.certificate
					return nil
				}
			}
		}
		return fmt.Errorf("cannot find challenge cert for domain %s", domain)
	}
	notify := func(err error, time time.Duration) {
		log.Errorf("Error getting cert: %v, retrying in %s", err, time)
	}
	ebo := backoff.NewExponentialBackOff()
	ebo.MaxElapsedTime = 60 * time.Second
	err := backoff.RetryNotify(safe.OperationWithRecover(operation), ebo, notify)
	if err != nil {
		log.Errorf("Error getting cert: %v", err)
		return nil, false
	}
	return result, true
}

func (c *challengeProvider) Present(domain, token, keyAuth string) error {
	log.Debugf("Challenge Present %s", domain)
	cert, _, err := tlsSNI01ChallengeCert(keyAuth)
	if err != nil {
		return err
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	transaction, object, err := c.store.Begin()
	if err != nil {
		return err
	}
	account := object.(*Account)
	if account.ChallengeCerts == nil {
		account.ChallengeCerts = map[string]*ChallengeCert{}
	}
	account.ChallengeCerts[domain] = &cert
	return transaction.Commit(account)
}

func (c *challengeProvider) CleanUp(domain, token, keyAuth string) error {
	log.Debugf("Challenge CleanUp %s", domain)
	c.lock.Lock()
	defer c.lock.Unlock()
	transaction, object, err := c.store.Begin()
	if err != nil {
		return err
	}
	account := object.(*Account)
	delete(account.ChallengeCerts, domain)
	return transaction.Commit(account)
}

func (c *challengeProvider) Timeout() (timeout, interval time.Duration) {
	return 60 * time.Second, 5 * time.Second
}

// tlsSNI01ChallengeCert returns a certificate and target domain for the `tls-sni-01` challenge
func tlsSNI01ChallengeCert(keyAuth string) (ChallengeCert, string, error) {
	// generate a new RSA key for the certificates
	var tempPrivKey crypto.PrivateKey
	tempPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return ChallengeCert{}, "", err
	}
	rsaPrivKey := tempPrivKey.(*rsa.PrivateKey)
	rsaPrivPEM := pemEncode(rsaPrivKey)

	zBytes := sha256.Sum256([]byte(keyAuth))
	z := hex.EncodeToString(zBytes[:sha256.Size])
	domain := fmt.Sprintf("%s.%s.acme.invalid", z[:32], z[32:])
	tempCertPEM, err := generate.PemCert(rsaPrivKey, domain, time.Time{})
	if err != nil {
		return ChallengeCert{}, "", err
	}

	certificate, err := tls.X509KeyPair(tempCertPEM, rsaPrivPEM)
	if err != nil {
		return ChallengeCert{}, "", err
	}

	return ChallengeCert{Certificate: tempCertPEM, PrivateKey: rsaPrivPEM, certificate: &certificate}, domain, nil
}

func pemEncode(data interface{}) []byte {
	var pemBlock *pem.Block
	switch key := data.(type) {
	case *ecdsa.PrivateKey:
		keyBytes, _ := x509.MarshalECPrivateKey(key)
		pemBlock = &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}
	case *rsa.PrivateKey:
		pemBlock = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	case *x509.CertificateRequest:
		pemBlock = &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: key.Raw}
	case []byte:
		pemBlock = &pem.Block{Type: "CERTIFICATE", Bytes: data.([]byte)}
	}

	return pem.EncodeToMemory(pemBlock)
}
